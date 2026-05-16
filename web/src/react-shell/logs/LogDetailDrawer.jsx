import React, { useEffect, useMemo, useState } from 'react'
import { formatLogError, logsApi } from '../api/logs.js'
import { SEVERITY_META, normalizeLevel, formatTime, formatDetailValue } from './LogsViewKit.jsx'

/**
 * DEGRADE-049: 日志详情抽屉 + 上下文查看
 * 在 LogDetailDrawer 中添加"查看上下文"按钮
 * 点击后查询该日志前后 N 条（默认前后各 10 条）
 * 显示为时间线列表，当前日志高亮
 */
export function LogDetailDrawer({ row, onClose, similarRows = [], onSelectSimilar }) {
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
        setContextError('PENDING: 后端不支持 /api/v1/logs/context 接口。')
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
            <pre className='fx-msg'>{row.body || row.message || '-'}</pre>
          </div>
          <div className='fx-drawer__section'>
            <h4>属性（{entries.length} 项）</h4>
            <dl className='fx-kv'>
              {entries.map(([key, value]) => (
                <React.Fragment key={key}>
                  <dt>{key}</dt>
                  <dd>{value}</dd>
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
