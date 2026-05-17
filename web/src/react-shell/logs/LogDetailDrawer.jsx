import React, { useEffect, useMemo, useState } from 'react'
import { formatLogError, logsApi } from '../api/logs.js'
import { SEVERITY_META, normalizeLevel, formatTime, formatDetailValue } from './LogsViewKit.jsx'

/**
 * DEGRADE-049: 日志详情抽屉 + 上下文查看
 * 在 LogDetailDrawer 中添加"查看上下文"按钮
 * 点击后查询该日志前后 N 条（默认前后各 10 条）
 * 显示为时间线列表，当前日志高亮
 */
export function LogDetailDrawer({ row, onClose, similarRows = [], onSelectSimilar, onAddToQuery }) {
  const entries = useMemo(() => flattenRow(row), [row])
  const levelKey = normalizeLevel(row.severity_text || row.level)
  const ts = formatTime(row.timestamp)
  const svc = row.source_name || row.service_name || row.source || '-'
  const host = row.host_name || row.host || '-'
  const trace = row.trace_id || ''
  const span = row.span_id || ''

  const [contextRows, setContextRows] = useState([])
  const [contextLoading, setContextLoading] = useState(false)
  const [contextError, setContextError] = useState('')
  const [contextBefore, setContextBefore] = useState(10)
  const [contextAfter, setContextAfter] = useState(10)

  useEffect(() => {
    setContextRows([])
    setContextError('')
  }, [row])

  const loadContext = async () => {
    setContextLoading(true)
    setContextError('')
    try {
      const resp = await logsApi.contextById(row.id, { before: contextBefore, after: contextAfter })
      const items = Array.isArray(resp?.items) ? resp.items : []
      setContextRows(items)
      if (!items.length) setContextError('未找到上下文日志。')
    } catch (err) {
      setContextRows([])
      if (err?.status === 404 || err?.status === 501) {
        setContextError('后端不支持 /api/v1/logs/context 接口。')
      } else {
        setContextError(formatLogError(err))
      }
    } finally {
      setContextLoading(false)
    }
  }

  const openTrace = () => {
    if (!trace) return
    window.location.href = `/tracing?traceId=${encodeURIComponent(trace)}${span ? `&spanId=${encodeURIComponent(span)}` : ''}`
  }

  const copyJson = () => {
    try {
      const text = JSON.stringify(row, null, 2)
      if (navigator?.clipboard) navigator.clipboard.writeText(text)
    } catch (_) { /* noop */ }
  }

  const handleMaskClick = (event) => {
    if (event.target === event.currentTarget) onClose()
  }

  const messageBody = row.body || row.message || ''
  const parsedJson = tryParseJson(messageBody)

  return (
    <>
      <div className='fx-drawer-mask' onClick={handleMaskClick} />
      <aside className='fx-drawer' role='dialog' aria-label='日志详情'>
        <div className='fx-drawer__head'>
          <h3 className='fx-drawer__title'>日志详情</h3>
          <button type='button' className='fx-drawer__close' onClick={onClose} aria-label='关闭详情'>×</button>
        </div>
        <div className='fx-drawer__meta'>
          <span><b>时间</b>{ts}</span>
          <span style={{ color: SEVERITY_META[levelKey]?.color, fontWeight: 700 }}><b>级别</b>{levelKey.toUpperCase()}</span>
          <span><b>服务</b>{svc}</span>
          <span><b>主机</b>{host}</span>
          {trace && <span><b>Trace</b>{trace.slice(0, 16)}...</span>}
        </div>
        <div className='fx-drawer__body'>
          <div className='fx-drawer__section'>
            <h4>消息体</h4>
            {parsedJson ? (
              <JsonTreeView data={parsedJson} onCopy={copyJson} onAddToQuery={onAddToQuery} />
            ) : (
              <pre className='fx-msg'>{messageBody || '-'}</pre>
            )}
          </div>
          <div className='fx-drawer__section'>
            <h4>属性（{entries.length} 项）</h4>
            <dl className='fx-kv'>
              {entries.map(([key, value]) => (
                <React.Fragment key={key}>
                  <dt>{key}</dt>
                  <dd>
                    {value}
                    {onAddToQuery && (
                      <button
                        type='button'
                        onClick={() => onAddToQuery(key, value)}
                        style={{ marginLeft: 6, fontSize: 10, padding: '1px 6px', border: '1px solid var(--fx-border)', borderRadius: 3, background: '#fff', cursor: 'pointer', color: 'var(--fx-primary, #1769ff)' }}
                        title='添加到查询条件'
                      >
                        +查询
                      </button>
                    )}
                  </dd>
                </React.Fragment>
              ))}
            </dl>
          </div>
          <div className='fx-drawer__section'>
            <h4>上下文日志</h4>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 8 }}>
              <label style={{ fontSize: 12 }}>前 <input type='number' min={1} max={50} value={contextBefore} onChange={e => setContextBefore(Number(e.target.value))} style={{ width: 48, textAlign: 'center' }} /> 条</label>
              <label style={{ fontSize: 12 }}>后 <input type='number' min={1} max={50} value={contextAfter} onChange={e => setContextAfter(Number(e.target.value))} style={{ width: 48, textAlign: 'center' }} /> 条</label>
              <button type='button' className='fx-qb__btn' style={{ minHeight: 28, fontSize: 12, padding: '0 12px' }} onClick={loadContext} disabled={contextLoading}>
                {contextLoading ? '加载中...' : '查看上下文'}
              </button>
            </div>
            {contextError && <div className='fx-logs-blocked' style={{ marginBottom: 8 }}>{contextError}</div>}
            {contextRows.length > 0 && <ContextTimeline rows={contextRows} currentId={row.id} />}
          </div>
          {similarRows.length > 0 && (
            <div className='fx-drawer__section'>
              <h4>相似日志（{similarRows.length}）</h4>
              <div className='fx-loglist' style={{ borderRadius: 6 }}>
                <div className='fx-loglist__body' style={{ maxHeight: 220 }}>
                  {similarRows.map((r, idx) => {
                    const lk = normalizeLevel(r.severity_text || r.level)
                    return (
                      <div key={r.id || idx} className={`fx-logrow is-${lk}`} onClick={() => onSelectSimilar && onSelectSimilar(r)}>
                        <span className='fx-logrow__ts'>{formatTime(r.timestamp)}</span>
                        <span className={`fx-logrow__lvl is-${lk}`}>{lk}</span>
                        <span className='fx-logrow__svc'>{r.source_name || r.service_name || '-'}</span>
                        <span className='fx-logrow__body'>{r.body || r.message || '-'}</span>
                        <span className='fx-logrow__trace'>{r.trace_id ? r.trace_id.slice(0,12)+'...' : '-'}</span>
                      </div>
                    )
                  })}
                </div>
              </div>
            </div>
          )}
        </div>
        <div className='fx-drawer__foot'>
          <button type='button' onClick={copyJson}>复制 JSON</button>
          <button type='button' disabled={!trace} onClick={openTrace} className={trace ? 'is-primary' : ''}>
            {trace ? '查看链路详情' : '无 Trace ID'}
          </button>
        </div>
      </aside>
    </>
  )
}

function ContextTimeline({ rows, currentId }) {
  return (
    <div className='fx-loglist' style={{ borderRadius: 6 }}>
      <div className='fx-loglist__body' style={{ maxHeight: 300 }}>
        {rows.map((r, idx) => {
          const lk = normalizeLevel(r.severity_text || r.level)
          const isCurrent = r.id === currentId
          return (
            <div
              key={r.id || idx}
              className={`fx-logrow is-${lk}${isCurrent ? ' is-active' : ''}`}
              style={isCurrent ? { background: 'var(--fx-primary-weak, rgba(23,105,255,.12))', fontWeight: 700 } : undefined}
            >
              <span className='fx-logrow__ts'>{formatTime(r.timestamp)}</span>
              <span className={`fx-logrow__lvl is-${lk}`}>{lk}</span>
              <span className='fx-logrow__svc'>{r.source_name || r.service_name || '-'}</span>
              <span className='fx-logrow__body'>{r.body || r.message || '-'}</span>
              <span className='fx-logrow__trace'>{r.trace_id ? r.trace_id.slice(0,12)+'...' : '-'}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function flattenRow(row) {
  const out = []
  for (const [key, value] of Object.entries(row || {})) {
    if (key === '__proto__') continue
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      for (const [k2, v2] of Object.entries(value)) {
        out.push([`${key}.${k2}`, formatDetailValue(v2)])
      }
    } else {
      out.push([key, formatDetailValue(value)])
    }
  }
  return out
}

function tryParseJson(value) {
  if (!value || typeof value !== 'string') return null
  const trimmed = value.trim()
  if ((trimmed.startsWith('{') && trimmed.endsWith('}')) || (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
    try { return JSON.parse(trimmed) } catch (_) { return null }
  }
  return null
}

function JsonTreeView({ data, onCopy, onAddToQuery }) {
  const [collapsed, setCollapsed] = useState({})

  const toggleCollapse = (path) => {
    setCollapsed(prev => ({ ...prev, [path]: !prev[path] }))
  }

  const copyRawJson = () => {
    try {
      const text = JSON.stringify(data, null, 2)
      if (navigator?.clipboard) navigator.clipboard.writeText(text)
    } catch (_) { /* noop */ }
  }

  return (
    <div style={{ position: 'relative', border: '1px solid var(--fx-border, #e3e8f1)', borderRadius: 8, padding: '10px 12px', background: 'var(--fx-bg-subtle, #f8fbff)' }}>
      <div style={{ position: 'absolute', top: 8, right: 8, display: 'flex', gap: 4 }}>
        <button
          type='button'
          onClick={copyRawJson}
          style={{ fontSize: 10, padding: '2px 8px', border: '1px solid var(--fx-border)', borderRadius: 4, background: '#fff', cursor: 'pointer' }}
          title='复制原始 JSON'
        >
          复制
        </button>
      </div>
      <div style={{ fontFamily: 'ui-monospace, Consolas, monospace', fontSize: 12, lineHeight: 1.6 }}>
        <JsonNode data={data} path='' collapsed={collapsed} onToggle={toggleCollapse} onAddToQuery={onAddToQuery} />
      </div>
    </div>
  )
}

function JsonNode({ data, path, collapsed, onToggle, onAddToQuery, indent = 0 }) {
  if (data === null) return <span style={{ color: '#8b5cf6' }}>null</span>
  if (typeof data === 'boolean') return <span style={{ color: '#8b5cf6' }}>{String(data)}</span>
  if (typeof data === 'number') return <span style={{ color: '#17a86b' }}>{data}</span>
  if (typeof data === 'string') {
    return (
      <span style={{ color: '#c45656' }}>
        &quot;{data.length > 200 ? data.slice(0, 200) + '...' : data}&quot;
      </span>
    )
  }

  const isArray = Array.isArray(data)
  const entries = isArray ? data.map((v, i) => [i, v]) : Object.entries(data)
  const isCollapsed = collapsed[path]
  const bracket = isArray ? ['[', ']'] : ['{', '}']

  if (!entries.length) {
    return <span>{bracket[0]}{bracket[1]}</span>
  }

  return (
    <span>
      <span
        onClick={() => onToggle(path)}
        style={{ cursor: 'pointer', userSelect: 'none', color: 'var(--fx-text-weak)' }}
      >
        {isCollapsed ? '▶ ' : '▼ '}
      </span>
      {bracket[0]}
      {isCollapsed ? (
        <span style={{ color: 'var(--fx-text-weak)' }}> ...{entries.length} 项 </span>
      ) : (
        <div style={{ paddingLeft: 16 }}>
          {entries.map(([key, value], idx) => {
            const childPath = path ? `${path}.${key}` : String(key)
            return (
              <div key={childPath} style={{ display: 'flex', alignItems: 'flex-start', gap: 4 }}>
                {!isArray && (
                  <span style={{ color: '#1769ff', fontWeight: 600 }}>
                    &quot;{key}&quot;
                    {onAddToQuery && typeof value !== 'object' && (
                      <button
                        type='button'
                        onClick={(e) => { e.stopPropagation(); onAddToQuery(childPath, String(value)) }}
                        style={{ marginLeft: 4, fontSize: 9, padding: '0 4px', border: '1px solid var(--fx-border)', borderRadius: 3, background: '#fff', cursor: 'pointer', color: 'var(--fx-primary, #1769ff)', verticalAlign: 'middle' }}
                        title='添加到查询条件'
                      >
                        +
                      </button>
                    )}
                  </span>
                )}
                {!isArray && <span>: </span>}
                <JsonNode data={value} path={childPath} collapsed={collapsed} onToggle={onToggle} onAddToQuery={onAddToQuery} indent={indent + 1} />
                {idx < entries.length - 1 && <span>,</span>}
              </div>
            )
          })}
        </div>
      )}
      {bracket[1]}
    </span>
  )
}
