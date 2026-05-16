import { useCallback, useEffect, useRef, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { formatLogError, LOG_BLOCKERS, logsApi } from '../api/logs.js'

export const LIVE_STATES = { stopped: 'STOPPED', playing: 'PLAYING', paused: 'PAUSED' }
export const WS_STATES = { connected: 'WS', polling: 'POLL', connecting: 'CONNECTING' }

const MAX_RETRY = 3

/**
 * DEGRADE-050: WebSocket 实时推送 hook
 * 优先使用 WebSocket，连接失败（3 次指数退避重试后）自动降级为轮询
 */
export function useLiveFollow() {
  const timerRef = useRef(null)
  const wsRef = useRef(null)
  const retryCountRef = useRef(0)
  const [control, setControl] = useState({
    source: 'findx_audit',
    query: '',
    severityFilter: '',
    serviceFilter: '',
    hostFilter: '',
    intervalSeconds: 3,
    state: LIVE_STATES.stopped,
  })
  const [data, setData] = useState(emptyData())
  const [arrival, setArrival] = useState(null)
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [arrivalLoading, setArrivalLoading] = useState(false)
  const [transport, setTransport] = useState(WS_STATES.polling)
  const [wsNotice, setWsNotice] = useState('')

  const clearTimer = () => { if (timerRef.current) { clearInterval(timerRef.current); timerRef.current = null } }
  const closeWs = () => { if (wsRef.current) { wsRef.current.close(); wsRef.current = null } }
  const setControlValue = patch => setControl(current => ({ ...current, ...patch }))

  const appendRows = useCallback((newItems) => {
    setData(current => {
      const existingIds = new Set(current.rows.map(r => r.id))
      const fresh = newItems.filter(r => r.id && !existingIds.has(r.id))
      const merged = [...current.rows, ...fresh].slice(-500)
      return {
        rows: merged.length > 0 ? merged : (newItems.length ? newItems : current.rows),
        meta: null,
        updatedAt: new Date().toLocaleString(),
        pollCount: current.pollCount + 1,
      }
    })
  }, [])

  const fetchPoll = useCallback(async (ctrl) => {
    if (ctrl.source !== 'findx_audit') { setBlocked(LOG_BLOCKERS.live); return }
    setLoading(true)
    try {
      const resp = await logsApi.query({
        source: ctrl.source,
        query: ctrl.query,
        severity: ctrl.severityFilter || undefined,
        service: ctrl.serviceFilter || undefined,
        host: ctrl.hostFilter || undefined,
        limit: 100,
      })
      const items = Array.isArray(resp?.items) ? resp.items : []
      appendRows(items)
      setBlocked('')
    } catch (err) {
      setBlocked(formatLogError(err))
      clearTimer()
      setControlValue({ state: LIVE_STATES.stopped })
    } finally {
      setLoading(false)
    }
  }, [appendRows])

  const connectWebSocket = useCallback((ctrl) => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsHost = window.location.host
    const url = `${protocol}//${wsHost}/api/v1/logs/live`
    setTransport(WS_STATES.connecting)
    setWsNotice('')

    try {
      const ws = new WebSocket(url)
      wsRef.current = ws

      ws.onopen = () => {
        retryCountRef.current = 0
        setTransport(WS_STATES.connected)
        setWsNotice('')
        ws.send(JSON.stringify({
          source: ctrl.source,
          query: ctrl.query,
          severity: ctrl.severityFilter || undefined,
          service: ctrl.serviceFilter || undefined,
          host: ctrl.hostFilter || undefined,
        }))
      }

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          const items = Array.isArray(msg) ? msg : msg?.items ? msg.items : [msg]
          appendRows(items)
        } catch (_) { /* 忽略非 JSON 消息 */ }
      }

      ws.onerror = () => { /* onclose 会处理 */ }

      ws.onclose = () => {
        wsRef.current = null
        if (ctrl.state === LIVE_STATES.playing || control.state === LIVE_STATES.playing) {
          retryCountRef.current += 1
          if (retryCountRef.current <= MAX_RETRY) {
            const delay = Math.pow(2, retryCountRef.current) * 1000
            setTimeout(() => {
              if (control.state === LIVE_STATES.playing) connectWebSocket(ctrl)
            }, delay)
          } else {
            setTransport(WS_STATES.polling)
            setWsNotice('WebSocket 不可用，已降级为轮询模式')
            startPolling(ctrl)
          }
        }
      }
    } catch (_) {
      setTransport(WS_STATES.polling)
      setWsNotice('WebSocket 不可用，已降级为轮询模式')
      startPolling(ctrl)
    }
  }, [appendRows, control.state])

  const startPolling = useCallback((ctrl) => {
    clearTimer()
    setTransport(WS_STATES.polling)
    fetchPoll(ctrl)
    timerRef.current = setInterval(() => fetchPoll(ctrl), ctrl.intervalSeconds * 1000)
  }, [fetchPoll])

  const startFollow = () => {
    clearTimer()
    closeWs()
    setBlocked('')
    setWsNotice('')
    retryCountRef.current = 0
    if (control.source !== 'findx_audit') {
      setControlValue({ state: LIVE_STATES.stopped })
      setBlocked(LOG_BLOCKERS.live)
      return
    }
    setControlValue({ state: LIVE_STATES.playing })
    connectWebSocket(control)
  }

  const pauseFollow = () => {
    clearTimer()
    closeWs()
    setControlValue({ state: data.rows.length ? LIVE_STATES.paused : LIVE_STATES.stopped })
  }

  const stopFollow = () => {
    clearTimer()
    closeWs()
    setControlValue({ state: LIVE_STATES.stopped })
    setTransport(WS_STATES.polling)
    setWsNotice('')
  }

  const changeSource = source => {
    clearTimer()
    closeWs()
    setControlValue({ source, state: LIVE_STATES.stopped })
    setData(emptyData())
    setBlocked('')
    setWsNotice('')
  }

  const clearLogs = () => setData(emptyData())

  const loadArrival = async () => {
    setBlocked('')
    setArrivalLoading(true)
    try {
      const rows = await agentApi.dataArrival()
      const logsRow = findLogsArrival(rows)
      setArrival(logsRow)
      if (!logsRow) setBlocked('PENDING: FindX Agent 数据到达契约未返回日志通道。')
      if (logsRow && logsRow.status !== 'reported') setBlocked(logsRow.blocker || LOG_BLOCKERS.agentLinkage)
    } catch (err) {
      setArrival(null)
      setBlocked(formatAgentError(err))
    } finally {
      setArrivalLoading(false)
    }
  }

  const openAgent = () => {
    const scope = control.query || firstRowScope(data.rows)
    window.location.href = `/agents?section=hosts&package=logs&q=${encodeURIComponent(scope || '')}`
  }

  useEffect(() => () => { clearTimer(); closeWs() }, [])
  useEffect(() => {
    if (control.state === LIVE_STATES.playing && transport === WS_STATES.polling) {
      startPolling(control)
    }
  }, [control.intervalSeconds])

  return {
    ...control,
    ...data,
    arrival,
    blocked,
    loading,
    arrivalLoading,
    transport,
    wsNotice,
    startFollow,
    pauseFollow,
    stopFollow,
    clearLogs,
    changeSource,
    setControlValue,
    loadArrival,
    openAgent,
  }
}

function findLogsArrival(rows) {
  return Array.isArray(rows) ? rows.find(r => r?.kind === 'logs') || null : null
}

function firstRowScope(rows) {
  const row = Array.isArray(rows) ? rows[0] : null
  return row?.service_name || row?.source_name || row?.trace_id || row?.scope || ''
}

function emptyData() {
  return { rows: [], meta: null, updatedAt: '', pollCount: 0 }
}
