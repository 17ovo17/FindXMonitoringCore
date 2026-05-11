import React, { useEffect, useRef, useState, useCallback } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { formatLogError, LOG_BLOCKERS, LOG_SOURCES, logsApi } from '../api/logs.js'
import { Blocked, Empty, Field, JsonPreview, Status } from './LogsShared.jsx'

const LIVE_STATES = {
  stopped: 'STOPPED',
  playing: 'PLAYING',
  paused: 'PAUSED',
}

const intervals = [
  { value: 3, label: '3 秒' },
  { value: 5, label: '5 秒' },
  { value: 10, label: '10 秒' },
  { value: 30, label: '30 秒' },
]

const SEVERITY_LEVELS = [
  { value: '', label: '全部' },
  { value: 'debug', label: 'DEBUG', color: '#8c8c8c' },
  { value: 'info', label: 'INFO', color: '#1769ff' },
  { value: 'warn', label: 'WARN', color: '#d4a017' },
  { value: 'error', label: 'ERROR', color: '#cf1322' },
]

const streamBlocker = 'BLOCKED_BY_CONTRACT: 生产 OTel 实时日志流、断线重连、背压和权限续期尚未接入；当前仅以轮询方式跟随 FindX 审计日志。'

export function LiveSection() {
  const live = useAuditLiveFollow()
  return (
    <section className='fx-logs-work'>
      <LiveFilters live={live} />
      <LiveToolbar live={live} />
      {live.source === 'findx_audit' && live.blocked && <Blocked>{live.blocked}</Blocked>}
      <LiveStreamView live={live} />
    </section>
  )
}

function useAuditLiveFollow() {
  const timerRef = useRef(null)
  const [control, setControl] = useState({ source: 'findx_audit', query: '', severityFilter: '', serviceFilter: '', intervalSeconds: 3, state: LIVE_STATES.stopped })
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

  const clearLogs = () => {
    setData(emptyData())
  }

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
  }, [control.intervalSeconds])

  return { ...control, ...data, arrival, blocked, loading, arrivalLoading, startFollow, pauseFollow, stopFollow, clearLogs, changeSource, setControlValue, loadArrival, openAgent }
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
      limit: 100,
    })
    const newItems = Array.isArray(resp?.items) ? resp.items : []
    setData(current => {
      // Append new items, deduplicate by id, keep last 500
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

function LiveFilters({ live }) {
  return (
    <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center', marginBottom: 12 }}>
      <Field label='来源'>
        <select value={live.source} onChange={event => live.changeSource(event.target.value)} style={{ height: 32, borderRadius: 8, fontSize: 12 }}>
          {LOG_SOURCES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
        </select>
      </Field>
      <Field label='过滤关键词'>
        <input value={live.query} onChange={event => live.setControlValue({ query: event.target.value })} placeholder='动作 / 资源 / Trace / 摘要' style={{ height: 32, borderRadius: 8, border: '1px solid #d8e1ee', padding: '0 10px', fontSize: 12, minWidth: 180 }} />
      </Field>
      <Field label='级别'>
        <select value={live.severityFilter} onChange={event => live.setControlValue({ severityFilter: event.target.value })} style={{ height: 32, borderRadius: 8, fontSize: 12 }}>
          {SEVERITY_LEVELS.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
        </select>
      </Field>
      <Field label='服务'>
        <input value={live.serviceFilter} onChange={event => live.setControlValue({ serviceFilter: event.target.value })} placeholder='服务名' style={{ height: 32, borderRadius: 8, border: '1px solid #d8e1ee', padding: '0 10px', fontSize: 12, width: 120 }} />
      </Field>
      <Field label='刷新间隔'>
        <select value={live.intervalSeconds} onChange={event => live.setControlValue({ intervalSeconds: Number(event.target.value) })} style={{ height: 32, borderRadius: 8, fontSize: 12 }}>
          {intervals.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
        </select>
      </Field>
    </div>
  )
}

function LiveToolbar({ live }) {
  return (
    <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center', marginBottom: 12 }}>
      {live.state !== LIVE_STATES.playing ? (
        <button type='button' onClick={live.startFollow} disabled={live.loading} style={toolbarBtnStyle('#1769ff', '#fff')}>
          {live.loading ? '连接中...' : '开始跟随'}
        </button>
      ) : (
        <button type='button' onClick={live.pauseFollow} style={toolbarBtnStyle('#d4a017', '#fff')}>
          暂停
        </button>
      )}
      <button type='button' onClick={live.stopFollow} disabled={live.state === LIVE_STATES.stopped && !live.rows.length} style={toolbarBtnStyle('#fff', '#25324a')}>
        停止
      </button>
      <button type='button' onClick={live.clearLogs} disabled={!live.rows.length} style={toolbarBtnStyle('#fff', '#9f3a38')}>
        清空
      </button>
      <span style={{ width: 1, height: 24, background: '#e3e8f1', margin: '0 4px' }} />
      <LiveIndicator state={live.state} />
      {live.updatedAt && <span style={{ fontSize: 12, color: '#66758d' }}>最近刷新：{live.updatedAt}</span>}
      <span style={{ fontSize: 12, color: '#66758d' }}>已接收 {live.rows.length} 条 | 轮询 {live.pollCount} 次</span>
      <span style={{ marginLeft: 'auto' }} />
      <button type='button' onClick={live.openAgent} style={toolbarBtnStyle('#fff', '#25324a')}>进入主机 Agent</button>
      <button type='button' onClick={live.loadArrival} disabled={live.arrivalLoading} style={toolbarBtnStyle('#fff', '#25324a')}>{live.arrivalLoading ? '读取中' : '数据到达'}</button>
    </div>
  )
}

function LiveIndicator({ state }) {
  const isPlaying = state === LIVE_STATES.playing
  const isPaused = state === LIVE_STATES.paused
  return (
    <span style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: 6,
      padding: '4px 10px',
      borderRadius: 14,
      fontSize: 12,
      fontWeight: 700,
      background: isPlaying ? '#e8f8ef' : isPaused ? '#fffbe6' : '#f5f5f5',
      color: isPlaying ? '#167346' : isPaused ? '#ad6800' : '#595959',
    }}>
      <span style={{
        width: 8,
        height: 8,
        borderRadius: '50%',
        background: isPlaying ? '#52c41a' : isPaused ? '#faad14' : '#bfbfbf',
        animation: isPlaying ? 'fx-pulse 1.5s infinite' : 'none',
      }} />
      {isPlaying ? '实时跟随中' : isPaused ? '已暂停' : '未启动'}
    </span>
  )
}

/* 实时日志流视图 - 自动滚动 */
function LiveStreamView({ live }) {
  const containerRef = useRef(null)
  const [autoScroll, setAutoScroll] = useState(true)

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [live.rows.length, autoScroll])

  // Detect manual scroll to disable auto-scroll
  const handleScroll = () => {
    if (!containerRef.current) return
    const { scrollTop, scrollHeight, clientHeight } = containerRef.current
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 40
    setAutoScroll(isAtBottom)
  }

  const filteredRows = live.rows

  if (live.source !== 'findx_audit') {
    return (
      <div className='fx-logs-panel'>
        <Empty>通用 OTel 实时日志仍为 BLOCKED_BY_CONTRACT，未发起实时请求。</Empty>
      </div>
    )
  }

  return (
    <div className='fx-logs-panel' style={{ padding: 0, overflow: 'hidden' }}>
      {/* 自动滚动提示 */}
      {!autoScroll && live.state === LIVE_STATES.playing && (
        <div style={{
          position: 'sticky',
          top: 0,
          zIndex: 2,
          textAlign: 'center',
          padding: '6px 12px',
          background: 'rgba(23, 105, 255, 0.9)',
          color: '#fff',
          fontSize: 12,
          cursor: 'pointer',
        }} onClick={() => setAutoScroll(true)}>
          有新日志到达，点击回到底部
        </div>
      )}

      {/* 日志流容器 */}
      <div
        ref={containerRef}
        onScroll={handleScroll}
        style={{
          maxHeight: 520,
          overflowY: 'auto',
          background: '#17233c',
          padding: '8px 0',
        }}
      >
        {filteredRows.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '48px 16px', color: '#8ca0c0', fontSize: 13 }}>
            <div style={{ fontSize: 24, marginBottom: 8 }}>{'>'}_</div>
            <p style={{ margin: 0 }}>
              {live.state === LIVE_STATES.playing
                ? '等待日志到达...'
                : '点击「开始跟随」启动实时日志流（轮询模式，每 ' + live.intervalSeconds + ' 秒刷新）'}
            </p>
          </div>
        ) : (
          filteredRows.map((row, index) => (
            <LiveLogLine key={row.id || index} row={row} />
          ))
        )}
      </div>

      {/* 底部状态栏 */}
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '6px 12px',
        background: '#f8fbff',
        borderTop: '1px solid #e3e8f1',
        fontSize: 12,
        color: '#66758d',
      }}>
        <span>模式：轮询（WebSocket 未接入）| 间隔：{live.intervalSeconds}s</span>
        <span>{filteredRows.length} 条日志 | {autoScroll ? '自动滚动' : '手动浏览'}</span>
      </div>
    </div>
  )
}

/* 单行实时日志 */
function LiveLogLine({ row }) {
  const level = (row.severity_text || row.level || 'info').toLowerCase()
  const levelColors = {
    error: '#ff4d4f',
    warn: '#faad14',
    warning: '#faad14',
    info: '#1890ff',
    debug: '#8c8c8c',
  }
  const color = levelColors[level] || levelColors.info

  return (
    <div style={{
      display: 'flex',
      gap: 8,
      padding: '3px 12px',
      fontFamily: 'monospace',
      fontSize: 12,
      lineHeight: 1.7,
      color: '#d4dce8',
      borderLeft: `3px solid ${color}`,
      marginBottom: 1,
    }}>
      <span style={{ color: '#6eb4ff', flexShrink: 0, minWidth: 150 }}>{formatTime(row.timestamp)}</span>
      <span style={{
        color: color,
        fontWeight: 700,
        minWidth: 48,
        flexShrink: 0,
        textTransform: 'uppercase',
      }}>{level}</span>
      <span style={{ color: '#a0c4ff', flexShrink: 0, minWidth: 100, maxWidth: 140, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
        {row.source_name || row.service_name || row.source || '-'}
      </span>
      <span style={{ color: '#f0f4f8', flex: 1, wordBreak: 'break-all' }}>
        {row.body || row.message || '-'}
      </span>
    </div>
  )
}

function findLogsArrival(rows) {
  return Array.isArray(rows) ? rows.find(row => row?.kind === 'logs') || null : null
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

function formatTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString()
}

function toolbarBtnStyle(bg, color) {
  return {
    height: 32,
    padding: '0 14px',
    border: bg === '#fff' ? '1px solid #d8e1ee' : 'none',
    borderRadius: 8,
    background: bg,
    color: color,
    fontSize: 13,
    fontWeight: 600,
    cursor: 'pointer',
  }
}
