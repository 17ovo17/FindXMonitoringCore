import React, { useEffect, useState, useCallback } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { formatLogError, LOG_BLOCKERS, LOG_SOURCES, logsApi } from '../api/logs.js'
import { Blocked, Empty, Field, JsonPreview, Status } from './LogsShared.jsx'

const TIME_RANGES = [
  { value: '15m', label: '最近 15 分钟' },
  { value: '1h', label: '最近 1 小时' },
  { value: '6h', label: '最近 6 小时' },
  { value: '24h', label: '最近 24 小时' },
  { value: '7d', label: '最近 7 天' },
  { value: '30d', label: '最近 30 天' },
]

const SEVERITY_LEVELS = [
  { value: '', label: '全部级别' },
  { value: 'debug', label: 'DEBUG', color: '#8c8c8c' },
  { value: 'info', label: 'INFO', color: '#1769ff' },
  { value: 'warn', label: 'WARN', color: '#d4a017' },
  { value: 'error', label: 'ERROR', color: '#cf1322' },
]

const PAGE_SIZE = 50

export function QuerySection() {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [arrivalLoading, setArrivalLoading] = useState(false)
  const [query, setQuery] = useState('')
  const [source, setSource] = useState('findx_audit')
  const [view, setView] = useState('table')
  const [timeRange, setTimeRange] = useState('24h')
  const [severity, setSeverity] = useState('')
  const [serviceFilter, setServiceFilter] = useState('')
  const [hostFilter, setHostFilter] = useState('')
  const [rows, setRows] = useState([])
  const [meta, setMeta] = useState(null)
  const [arrival, setArrival] = useState(null)
  const [expandedId, setExpandedId] = useState(null)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(false)

  const scope = query || firstRowScope(rows)

  const buildParams = useCallback((pageNum) => ({
    source,
    query,
    view,
    limit: PAGE_SIZE,
    offset: (pageNum - 1) * PAGE_SIZE,
    time_range: timeRange,
    severity: severity || undefined,
    service: serviceFilter || undefined,
    host: hostFilter || undefined,
  }), [source, query, view, timeRange, severity, serviceFilter, hostFilter])

  const runQuery = async () => {
    setLoading(true)
    setError('')
    setExpandedId(null)
    setPage(1)
    if (source !== 'findx_audit') {
      setRows([])
      setMeta(null)
      setError(LOG_BLOCKERS.query)
      setLoading(false)
      return
    }
    try {
      const resp = await logsApi.query(buildParams(1))
      const items = Array.isArray(resp?.items) ? resp.items : []
      setRows(items)
      setMeta(resp || null)
      setHasMore(items.length >= PAGE_SIZE)
    } catch (err) {
      setRows([])
      setMeta(null)
      setError(formatLogError(err))
      setHasMore(false)
    } finally {
      setLoading(false)
    }
  }

  const loadMore = async () => {
    if (loadingMore || !hasMore) return
    setLoadingMore(true)
    const nextPage = page + 1
    try {
      const resp = await logsApi.query(buildParams(nextPage))
      const items = Array.isArray(resp?.items) ? resp.items : []
      setRows(prev => [...prev, ...items])
      setMeta(resp || null)
      setPage(nextPage)
      setHasMore(items.length >= PAGE_SIZE)
    } catch (err) {
      setError(formatLogError(err))
    } finally {
      setLoadingMore(false)
    }
  }

  const openAgent = () => {
    window.location.href = `/agents?section=hosts&package=logs&q=${encodeURIComponent(scope || '')}`
  }

  const loadArrival = async () => {
    setError('')
    setArrivalLoading(true)
    try {
      const rowsValue = await agentApi.dataArrival()
      const logsRow = findLogsArrival(rowsValue)
      setArrival(logsRow)
      if (!logsRow) setError('BLOCKED_BY_CONTRACT: FindX Agent 数据到达契约未返回日志通道。')
      if (logsRow && logsRow.status !== 'reported') setError(logsRow.blocker || LOG_BLOCKERS.agentLinkage)
    } catch (err) {
      setArrival(null)
      setError(formatAgentError(err))
    } finally {
      setArrivalLoading(false)
    }
  }

  const handleKeyDown = (event) => {
    if (event.key === 'Enter') runQuery()
  }

  const toggleExpand = (id) => {
    setExpandedId(prev => prev === id ? null : id)
  }

  const clearFilters = () => {
    setSeverity('')
    setServiceFilter('')
    setHostFilter('')
  }

  useEffect(() => {
    runQuery()
  }, [])

  const selectedSource = LOG_SOURCES.find(item => item.value === source)
  const activeFilters = [severity, serviceFilter, hostFilter].filter(Boolean)

  return (
    <section className='fx-logs-work'>
      {/* 搜索栏 */}
      <div style={{ marginBottom: 12 }}>
        <div style={{ display: 'flex', gap: 8, alignItems: 'stretch' }}>
          <input
            value={query}
            onChange={event => setQuery(event.target.value)}
            onKeyDown={handleKeyDown}
            placeholder='输入关键词检索日志：动作、资源、Trace ID、摘要内容...'
            style={{
              flex: 1,
              minHeight: 42,
              border: '2px solid var(--fx-border, #d8e1ee)',
              borderRadius: 8,
              padding: '8px 14px',
              fontSize: 14,
              fontFamily: 'inherit',
              color: 'var(--fx-text, #25324a)',
              background: 'var(--fx-bg, #fff)',
            }}
          />
          <select
            value={timeRange}
            onChange={event => setTimeRange(event.target.value)}
            style={{
              minHeight: 42,
              border: '2px solid var(--fx-border, #d8e1ee)',
              borderRadius: 8,
              padding: '8px 12px',
              fontSize: 13,
              color: 'var(--fx-text, #25324a)',
              background: 'var(--fx-bg, #fff)',
              cursor: 'pointer',
            }}
          >
            {TIME_RANGES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
          </select>
          <button
            type='button'
            onClick={runQuery}
            disabled={loading}
            style={{
              minHeight: 42,
              padding: '0 20px',
              border: 'none',
              borderRadius: 8,
              background: 'var(--fx-primary, #1769ff)',
              color: '#fff',
              fontSize: 14,
              fontWeight: 700,
              cursor: loading ? 'not-allowed' : 'pointer',
              opacity: loading ? 0.7 : 1,
            }}
          >
            {loading ? '检索中...' : '检索'}
          </button>
        </div>
      </div>

      {/* 过滤条件 */}
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center', marginBottom: 12 }}>
        <span style={{ fontSize: 12, color: '#66758d', fontWeight: 700 }}>过滤：</span>
        {SEVERITY_LEVELS.map(item => (
          <FilterChip
            key={item.value}
            label={item.label}
            color={item.color}
            active={severity === item.value}
            onClick={() => setSeverity(severity === item.value ? '' : item.value)}
          />
        ))}
        <span style={{ width: 1, height: 20, background: '#e3e8f1', margin: '0 4px' }} />
        <input
          value={serviceFilter}
          onChange={event => setServiceFilter(event.target.value)}
          placeholder='服务名'
          style={{
            height: 28,
            border: '1px solid #d8e1ee',
            borderRadius: 14,
            padding: '0 10px',
            fontSize: 12,
            width: 120,
            color: 'var(--fx-text, #25324a)',
            background: serviceFilter ? '#eef5ff' : '#fff',
          }}
        />
        <input
          value={hostFilter}
          onChange={event => setHostFilter(event.target.value)}
          placeholder='主机名'
          style={{
            height: 28,
            border: '1px solid #d8e1ee',
            borderRadius: 14,
            padding: '0 10px',
            fontSize: 12,
            width: 120,
            color: 'var(--fx-text, #25324a)',
            background: hostFilter ? '#eef5ff' : '#fff',
          }}
        />
        {activeFilters.length > 0 && (
          <button
            type='button'
            onClick={clearFilters}
            style={{
              height: 28,
              border: '1px solid #d8e1ee',
              borderRadius: 14,
              padding: '0 10px',
              fontSize: 12,
              background: '#fff',
              color: '#9f3a38',
              cursor: 'pointer',
            }}
          >
            清除过滤
          </button>
        )}
        <span style={{ width: 1, height: 20, background: '#e3e8f1', margin: '0 4px' }} />
        <Field label='来源'>
          <select value={source} onChange={event => setSource(event.target.value)} style={{ height: 28, borderRadius: 14, fontSize: 12 }}>
            {LOG_SOURCES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
          </select>
        </Field>
        <button type='button' onClick={openAgent} style={{ height: 28, border: '1px solid #d8e1ee', borderRadius: 14, padding: '0 10px', fontSize: 12, background: '#fff', cursor: 'pointer' }}>进入主机 Agent</button>
        <button type='button' onClick={loadArrival} disabled={arrivalLoading} style={{ height: 28, border: '1px solid #d8e1ee', borderRadius: 14, padding: '0 10px', fontSize: 12, background: '#fff', cursor: 'pointer' }}>{arrivalLoading ? '读取中' : '数据到达'}</button>
      </div>

      {source !== 'findx_audit' && <Blocked>{LOG_BLOCKERS.query}</Blocked>}
      {error && <Blocked>{error}</Blocked>}

      {/* 结果区域 */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: 12 }}>
        <div className='fx-logs-panel'>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
            <h3 style={{ margin: 0, color: '#17233c', fontSize: 15 }}>
              检索结果 {meta?.total != null && <span style={{ fontSize: 12, color: '#66758d', fontWeight: 400 }}>（共 {meta.total} 条）</span>}
            </h3>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <select value={view} onChange={event => setView(event.target.value)} style={{ height: 28, borderRadius: 8, border: '1px solid #d8e1ee', fontSize: 12, padding: '0 8px' }}>
                <option value='table'>表格视图</option>
                <option value='raw'>原始视图</option>
              </select>
            </div>
          </div>
          {rows.length ? (
            <div>
              <div style={{ overflowX: 'auto' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: 760 }}>
                  <thead>
                    <tr>
                      <th style={thStyle}>时间</th>
                      <th style={thStyle}>级别</th>
                      <th style={thStyle}>服务</th>
                      <th style={{ ...thStyle, width: '40%' }}>内容</th>
                      <th style={thStyle}>Trace</th>
                    </tr>
                  </thead>
                  <tbody>
                    {rows.map((row, index) => {
                      const rowId = row.id || index
                      const isExpanded = expandedId === rowId
                      return (
                        <LogRow
                          key={rowId}
                          row={row}
                          isExpanded={isExpanded}
                          onToggle={() => toggleExpand(rowId)}
                        />
                      )
                    })}
                  </tbody>
                </table>
              </div>
              {hasMore && (
                <div style={{ textAlign: 'center', marginTop: 12 }}>
                  <button
                    type='button'
                    onClick={loadMore}
                    disabled={loadingMore}
                    style={{
                      minHeight: 36,
                      padding: '0 24px',
                      border: '1px solid var(--fx-primary, #1769ff)',
                      borderRadius: 8,
                      background: '#fff',
                      color: 'var(--fx-primary, #1769ff)',
                      fontSize: 13,
                      fontWeight: 600,
                      cursor: loadingMore ? 'not-allowed' : 'pointer',
                    }}
                  >
                    {loadingMore ? '加载中...' : '加载更多'}
                  </button>
                </div>
              )}
            </div>
          ) : (
            <NoDataGuide source={source} error={error} />
          )}
        </div>
      </div>
    </section>
  )
}

/* 过滤标签组件 */
function FilterChip({ label, color, active, onClick }) {
  return (
    <button
      type='button'
      onClick={onClick}
      style={{
        height: 28,
        border: active ? `1px solid ${color || '#1769ff'}` : '1px solid #d8e1ee',
        borderRadius: 14,
        padding: '0 10px',
        fontSize: 12,
        fontWeight: active ? 700 : 400,
        background: active ? (color ? `${color}18` : '#eef5ff') : '#fff',
        color: active ? (color || '#1769ff') : '#25324a',
        cursor: 'pointer',
        transition: 'all .15s',
      }}
    >
      {color && <span style={{ display: 'inline-block', width: 8, height: 8, borderRadius: '50%', background: color, marginRight: 4 }} />}
      {label}
    </button>
  )
}

/* 单条日志行 + 展开详情 */
function LogRow({ row, isExpanded, onToggle }) {
  const level = (row.severity_text || row.level || '').toLowerCase()
  return (
    <>
      <tr onClick={onToggle} style={{ cursor: 'pointer', background: isExpanded ? '#f8fbff' : 'transparent' }}>
        <td style={tdStyle}>{formatTime(row.timestamp)}</td>
        <td style={tdStyle}><LevelBadge level={level} /></td>
        <td style={tdStyle}>{row.source_name || row.service_name || row.source || '-'}</td>
        <td style={{ ...tdStyle, fontFamily: 'monospace', fontSize: 12, maxWidth: 400, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{row.body || row.message || '-'}</td>
        <td style={{ ...tdStyle, fontFamily: 'monospace', fontSize: 11 }}>{row.trace_id ? row.trace_id.slice(0, 12) + '...' : '-'}</td>
      </tr>
      {isExpanded && (
        <tr>
          <td colSpan={5} style={{ padding: 0, border: 'none' }}>
            <LogDetail row={row} />
          </td>
        </tr>
      )}
    </>
  )
}

/* 日志详情展开面板 */
function LogDetail({ row }) {
  const entries = Object.entries(row).filter(([key]) => key !== '__proto__')
  return (
    <div style={{
      background: '#17233c',
      color: '#f8fbff',
      padding: '12px 16px',
      margin: '0 0 1px',
      fontSize: 12,
      fontFamily: 'monospace',
      lineHeight: 1.8,
      maxHeight: 300,
      overflowY: 'auto',
      borderRadius: '0 0 8px 8px',
    }}>
      <div style={{ marginBottom: 8, color: '#8ca0c0', fontSize: 11 }}>点击行可收起详情</div>
      {entries.map(([key, value]) => (
        <div key={key} style={{ display: 'flex', gap: 8 }}>
          <span style={{ color: '#6eb4ff', minWidth: 140, flexShrink: 0 }}>{key}:</span>
          <span style={{ color: '#f8fbff', wordBreak: 'break-all' }}>{formatDetailValue(value)}</span>
        </div>
      ))}
    </div>
  )
}

/* 级别标签 */
function LevelBadge({ level }) {
  const colors = {
    error: { bg: '#fff1f0', color: '#cf1322', border: '#ffa39e' },
    warn: { bg: '#fffbe6', color: '#ad6800', border: '#ffe58f' },
    warning: { bg: '#fffbe6', color: '#ad6800', border: '#ffe58f' },
    info: { bg: '#e6f7ff', color: '#0050b3', border: '#91d5ff' },
    debug: { bg: '#f5f5f5', color: '#595959', border: '#d9d9d9' },
  }
  const style = colors[level] || colors.info
  return (
    <span style={{
      display: 'inline-block',
      padding: '2px 8px',
      borderRadius: 4,
      fontSize: 11,
      fontWeight: 700,
      fontFamily: 'monospace',
      background: style.bg,
      color: style.color,
      border: `1px solid ${style.border}`,
      textTransform: 'uppercase',
    }}>
      {level || '-'}
    </span>
  )
}

/* 无数据引导 */
function NoDataGuide({ source, error }) {
  if (source !== 'findx_audit' || error) {
    return <Empty>{'通用 OTel 日志数据源仍为 BLOCKED_BY_CONTRACT，未展示静态日志行。'}</Empty>
  }
  return (
    <div style={{
      textAlign: 'center',
      padding: '32px 16px',
      color: '#66758d',
      fontSize: 13,
      lineHeight: 2,
    }}>
      <div style={{ fontSize: 32, marginBottom: 8 }}>{'{ }'}</div>
      <p style={{ margin: 0, fontWeight: 700, color: '#17233c' }}>暂无匹配的日志数据</p>
      <p style={{ margin: '8px 0 0' }}>如需接入日志源，请按以下步骤操作：</p>
      <ol style={{ textAlign: 'left', display: 'inline-block', margin: '12px 0 0', padding: '0 0 0 20px' }}>
        <li>在「接入管道」页面配置日志解析规则</li>
        <li>在目标主机安装 FindX Agent 并启用日志采集能力包</li>
        <li>确认 Agent 心跳正常后，返回此页面检索</li>
      </ol>
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

function formatTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString()
}

function formatDetailValue(value) {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'object') {
    try { return JSON.stringify(value) } catch { return String(value) }
  }
  return String(value)
}

const thStyle = {
  borderBottom: '1px solid #e8edf5',
  padding: '10px',
  textAlign: 'left',
  fontSize: 12,
  color: '#66758d',
  background: '#f8fbff',
  fontWeight: 800,
}

const tdStyle = {
  borderBottom: '1px solid #e8edf5',
  padding: '10px',
  textAlign: 'left',
  fontSize: 13,
  verticalAlign: 'top',
  color: '#25324a',
}
