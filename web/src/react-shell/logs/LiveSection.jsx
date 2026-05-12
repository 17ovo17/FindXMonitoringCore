import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { formatLogError, LOG_BLOCKERS, LOG_SOURCES, logsApi } from '../api/logs.js'
import { Blocked, Field } from './LogsShared.jsx'
import {
  LogsQueryBuilder,
  SeverityChips,
  LogChartView,
  buildHistogram,
  normalizeLevel,
  formatTime,
} from './LogsViewKit.jsx'

const LIVE_STATES = { stopped: 'STOPPED', playing: 'PLAYING', paused: 'PAUSED' }

const INTERVALS = [
  { value: 3, label: '3 秒' },
  { value: 5, label: '5 秒' },
  { value: 10, label: '10 秒' },
  { value: 30, label: '30 秒' },
]

export function LiveSection() {
  const live = useAuditLiveFollow()
  const histogram = useMemo(() => buildHistogram(live.rows, 24), [live.rows])
  return (
    <section className='fx-logs-work'>
      <LogsQueryBuilder
        query={live.query}
        onQueryChange={value => live.setControlValue({ query: value })}
        onSubmit={live.startFollow}
        loading={live.loading}
        hideTimeRange
        submitLabel='开始跟随'
        submittingLabel='连接中...'
        placeholder='输入关键词过滤实时日志（动作 / 资源 / Trace / 摘要）'
        extraRight={
          <>
            <Field label='来源'>
              <select value={live.source} onChange={event => live.changeSource(event.target.value)} className='fx-qb__select'>
                {LOG_SOURCES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
              </select>
            </Field>
            <IntervalSelect value={live.intervalSeconds} onChange={v => live.setControlValue({ intervalSeconds: v })} />
            <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={live.openAgent}>进入主机 Agent</button>
            <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={live.loadArrival} disabled={live.arrivalLoading}>
              {live.arrivalLoading ? '读取中' : '数据到达'}
            </button>
          </>
        }
      />

      <SeverityChips
        value={live.severityFilter}
        onChange={value => live.setControlValue({ severityFilter: value })}
        serviceFilter={live.serviceFilter}
        onServiceChange={value => live.setControlValue({ serviceFilter: value })}
        hostFilter={live.hostFilter}
        onHostChange={value => live.setControlValue({ hostFilter: value })}
        onClearAll={live.severityFilter || live.serviceFilter || live.hostFilter
          ? () => live.setControlValue({ severityFilter: '', serviceFilter: '', hostFilter: '' })
          : null}
      />

      <LiveToolbar live={live} />

      {live.source === 'findx_audit' && live.blocked && <Blocked>{live.blocked}</Blocked>}

      <LogChartView
        histogram={histogram}
        title={`实时日志速率（近 ${live.rows.length} 条 / 每桶数量）`}
        height={160}
      />
      <LiveStreamView live={live} />
    </section>
  )
}

function IntervalSelect({ value, onChange }) {
  return (
    <Field label='刷新间隔'>
      <select value={value} onChange={event => onChange(Number(event.target.value))} className='fx-qb__select'>
        {INTERVALS.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
      </select>
    </Field>
  )
}

function LiveToolbar({ live }) {
  const playing = live.state === LIVE_STATES.playing
  return (
    <div className='fx-qb' style={{ gap: 8 }}>
      {!playing ? (
        <button type='button' className='fx-qb__btn' onClick={live.startFollow} disabled={live.loading}>
          {live.loading ? '连接中...' : '开始跟随'}
        </button>
      ) : (
        <button type='button' className='fx-qb__btn fx-qb__btn--warn' onClick={live.pauseFollow}>
          暂停
        </button>
      )}
      <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={live.stopFollow} disabled={live.state === LIVE_STATES.stopped && !live.rows.length}>
        停止
      </button>
      <button type='button' className='fx-qb__btn fx-qb__btn--danger' onClick={live.clearLogs} disabled={!live.rows.length}>
        清空
      </button>
      <span className='fx-qb__divider' />
      <LiveIndicator state={live.state} />
      {live.updatedAt && <span style={{ fontSize: 12, color: 'var(--fx-text-weak,#66758d)' }}>最近刷新：{live.updatedAt}</span>}
      <span style={{ fontSize: 12, color: 'var(--fx-text-weak,#66758d)' }}>已接收 {live.rows.length} 条 · 轮询 {live.pollCount} 次</span>
    </div>
  )
}

function LiveIndicator({ state }) {
  const cls = state === LIVE_STATES.playing ? 'is-playing'
           : state === LIVE_STATES.paused  ? 'is-paused'
           : 'is-stopped'
  const text = state === LIVE_STATES.playing ? '实时跟随中'
            : state === LIVE_STATES.paused  ? '已暂停'
            : '未启动'
  return (
    <span className={`fx-live-indicator ${cls}`}>
      <span className='fx-live-indicator__dot' />
      {text}
    </span>
  )
}

function LiveStreamView({ live }) {
  const containerRef = useRef(null)
  const [autoScroll, setAutoScroll] = useState(true)

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [live.rows.length, autoScroll])

  const handleScroll = () => {
    if (!containerRef.current) return
    const { scrollTop, scrollHeight, clientHeight } = containerRef.current
    const atBottom = scrollHeight - scrollTop - clientHeight < 40
    setAutoScroll(atBottom)
  }

  if (live.source !== 'findx_audit') {
    return (
      <div className='fx-logs-panel'>
        <p style={{ margin: 0, color: 'var(--fx-text-weak,#66758d)' }}>通用 OTel 实时日志仍为 BLOCKED_BY_CONTRACT，未发起实时请求。</p>
      </div>
    )
  }

  return (
    <div className='fx-live'>
      {!autoScroll && live.state === LIVE_STATES.playing && (
        <button type='button' className='fx-live__notice' onClick={() => setAutoScroll(true)}>
          有新日志到达，点击回到底部
        </button>
      )}
      <div ref={containerRef} onScroll={handleScroll} className='fx-live__scroll'>
        {live.rows.length === 0 ? (
          <div className='fx-live__empty'>
            <big>{'>_'}</big>
            <p style={{ margin: 0 }}>
              {live.state === LIVE_STATES.playing
                ? '等待日志到达...'
                : `点击「开始跟随」启动实时日志流（轮询模式，每 ${live.intervalSeconds} 秒刷新）`}
            </p>
          </div>
        ) : (
          live.rows.map((row, idx) => <LiveLogLine key={row.id || idx} row={row} />)
        )}
      </div>
      <div className='fx-live__foot'>
        <span>模式：轮询（WebSocket 未接入） · 间隔：{live.intervalSeconds}s</span>
        <span>{live.rows.length} 条日志 · {autoScroll ? '自动滚动中' : '手动浏览'}</span>
      </div>
    </div>
  )
}

function LiveLogLine({ row }) {
  const level = normalizeLevel(row.severity_text || row.level)
  const svc = row.source_name || row.service_name || row.source || '-'
  return (
    <div className={`fx-live-line is-${level}`}>
      <span className='fx-live-line__ts'>{formatTime(row.timestamp)}</span>
      <span className={`fx-live-line__lvl is-${level}`}>{level}</span>
      <span className='fx-live-line__svc'>{svc}</span>
      <span className='fx-live-line__body'>{row.body || row.message || '-'}</span>
    </div>
  )
}

function useAuditLiveFollow() {
  const timerRef = useRef(null)
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

  const clearTimer = () => closeTimer(timerRef)
  const setControlValue = patch => setControl(current => ({ ...current, ...patch }))

  const fetchRows = useCallback(() => {
    fetchLiveRows({ control, setData, setBlocked, setLoading, clearTimer, setControlValue })
  }, [control])

  const loadArrival = async () => {
    setBlocked('')
    setArrivalLoading(true)
    try {
      const rows = await agentApi.dataArrival()
      const logsRow = findLogsArrival(rows)
      setArrival(logsRow)
      if (!logsRow) setBlocked('BLOCKED_BY_CONTRACT: FindX Agent 数据到达契约未返回日志通道。')
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

  const clearLogs = () => setData(emptyData())

  const stopFollow = () => {
    clearTimer()
    setControlValue({ state: LIVE_STATES.stopped })
  }

  const startFollow = () => {
    clearTimer()
    setBlocked('')
    if (control.source !== 'findx_audit') {
      setControlValue({ state: LIVE_STATES.stopped })
      setBlocked(LOG_BLOCKERS.live)
      return
    }
    setControlValue({ state: LIVE_STATES.playing })
    fetchRows()
    timerRef.current = setInterval(fetchRows, control.intervalSeconds * 1000)
  }

  const pauseFollow = () => {
    clearTimer()
    setControlValue({ state: data.rows.length ? LIVE_STATES.paused : LIVE_STATES.stopped })
  }

  const changeSource = source => {
    clearTimer()
    setControlValue({ source, state: LIVE_STATES.stopped })
    setData(emptyData())
    setBlocked('')
  }

  useEffect(() => () => clearTimer(), [])
  useEffect(() => {
    if (control.state === LIVE_STATES.playing) startFollow()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [control.intervalSeconds])

  return {
    ...control,
    ...data,
    arrival,
    blocked,
    loading,
    arrivalLoading,
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

async function fetchLiveRows({ control, setData, setBlocked, setLoading, clearTimer, setControlValue }) {
  if (control.source !== 'findx_audit') {
    setBlocked(LOG_BLOCKERS.live)
    return
  }
  setLoading(true)
  try {
    const resp = await logsApi.query({
      source: control.source,
      query: control.query,
      severity: control.severityFilter || undefined,
      service: control.serviceFilter || undefined,
      host: control.hostFilter || undefined,
      limit: 100,
    })
    const newItems = Array.isArray(resp?.items) ? resp.items : []
    setData(current => {
      const existingIds = new Set(current.rows.map(r => r.id))
      const fresh = newItems.filter(r => r.id && !existingIds.has(r.id))
      const merged = [...current.rows, ...fresh].slice(-500)
      return {
        rows: merged.length > 0 ? merged : newItems,
        meta: resp || null,
        updatedAt: new Date().toLocaleString(),
        pollCount: current.pollCount + 1,
      }
    })
    setBlocked('')
  } catch (err) {
    setBlocked(formatLogError(err))
    clearTimer()
    setControlValue({ state: LIVE_STATES.stopped })
  } finally {
    setLoading(false)
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

function closeTimer(timerRef) {
  if (!timerRef.current) return
  clearInterval(timerRef.current)
  timerRef.current = null
}
