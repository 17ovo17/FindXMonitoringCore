import React, { useState } from 'react'
import { displayText, fmtTime } from './tracingModel.js'
import { Empty, Status } from './TracingShared.jsx'

const TABS = [
  { key: 'tags', label: 'Tags' },
  { key: 'logs', label: 'Logs' },
  { key: 'refs', label: 'Refs' },
  { key: 'exceptions', label: 'Exceptions' },
]

function TagsTab({ span, onAddToFilter }) {
  const tags = span.raw?.tags || []
  if (!tags.length) return <Empty>暂无 Tags 数据</Empty>
  return (
    <table>
      <thead><tr><th>Key</th><th>Value</th><th>操作</th></tr></thead>
      <tbody>
        {tags.map((tag, idx) => (
          <tr key={idx}>
            <td style={{ fontWeight: 600 }}>{displayText(tag.key)}</td>
            <td style={{ wordBreak: 'break-all' }}>{displayText(tag.value)}</td>
            <td>
              <button
                type='button'
                onClick={() => onAddToFilter && onAddToFilter(tag.key, tag.value)}
                style={{ fontSize: 11, padding: '2px 8px', border: '1px solid var(--fx-border)', borderRadius: 4, background: '#fff', cursor: 'pointer' }}
              >
                添加到过滤
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

function LogsTab({ span }) {
  const logs = span.raw?.logs || []
  const [levelFilter, setLevelFilter] = useState('')
  const [expandedIdx, setExpandedIdx] = useState({})

  if (!logs.length) return <Empty>暂无 Span 日志</Empty>

  const LEVELS = ['error', 'warn', 'info']

  const getLogLevel = (log) => {
    const data = log.data || []
    const levelEntry = data.find(d => d.key === 'level' || d.key === 'severity' || d.key === 'log.level')
    if (levelEntry) return String(levelEntry.value || '').toLowerCase()
    const hasError = data.some(d => d.key === 'error.kind' || d.key === 'error' || d.key === 'exception')
    if (hasError) return 'error'
    return 'info'
  }

  const filtered = levelFilter
    ? logs.filter(log => getLogLevel(log) === levelFilter)
    : logs

  const toggleExpand = (idx) => {
    setExpandedIdx(prev => ({ ...prev, [idx]: !prev[idx] }))
  }

  const tryParseJson = (value) => {
    if (!value || typeof value !== 'string') return null
    const trimmed = value.trim()
    if ((trimmed.startsWith('{') && trimmed.endsWith('}')) || (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
      try { return JSON.parse(trimmed) } catch (_) { return null }
    }
    return null
  }

  return (
    <div>
      <div style={{ display: 'flex', gap: 6, marginBottom: 8 }}>
        <button
          type='button'
          style={{ fontSize: 11, padding: '2px 10px', border: '1px solid var(--fx-border)', borderRadius: 4, background: !levelFilter ? 'var(--fx-primary, #1769ff)' : '#fff', color: !levelFilter ? '#fff' : 'inherit', cursor: 'pointer' }}
          onClick={() => setLevelFilter('')}
        >
          全部
        </button>
        {LEVELS.map(lv => (
          <button
            key={lv}
            type='button'
            style={{ fontSize: 11, padding: '2px 10px', border: '1px solid var(--fx-border)', borderRadius: 4, background: levelFilter === lv ? 'var(--fx-primary, #1769ff)' : '#fff', color: levelFilter === lv ? '#fff' : 'inherit', cursor: 'pointer' }}
            onClick={() => setLevelFilter(lv === levelFilter ? '' : lv)}
          >
            {lv.toUpperCase()}
          </button>
        ))}
        <span style={{ fontSize: 11, color: 'var(--fx-muted)', marginLeft: 8, alignSelf: 'center' }}>
          {filtered.length}/{logs.length} 条
        </span>
      </div>
      <table>
        <thead><tr><th style={{ width: 140 }}>时间</th><th>级别</th><th>内容</th><th style={{ width: 40 }}></th></tr></thead>
        <tbody>
          {filtered.map((log, idx) => {
            const level = getLogLevel(log)
            const isExpanded = expandedIdx[idx]
            const logContent = (log.data || []).map(d => `${d.key}: ${d.value}`).join('\n') || log.message || JSON.stringify(log)
            const jsonParsed = tryParseJson(log.message || (log.data?.length === 1 ? log.data[0].value : null))

            return (
              <React.Fragment key={idx}>
                <tr style={{ cursor: 'pointer' }} onClick={() => toggleExpand(idx)}>
                  <td style={{ whiteSpace: 'nowrap', fontSize: 11 }}>{fmtTime(log.time || log.timestamp)}</td>
                  <td>
                    <span style={{ fontSize: 10, padding: '1px 6px', borderRadius: 3, background: level === 'error' ? '#fff1f0' : level === 'warn' ? '#fffbe6' : '#f0f5ff', color: level === 'error' ? '#cf1322' : level === 'warn' ? '#d4a017' : '#1769ff' }}>
                      {level.toUpperCase()}
                    </span>
                  </td>
                  <td style={{ wordBreak: 'break-all', fontSize: 12 }}>
                    {(log.data || []).slice(0, 2).map((d, i) => (
                      <span key={i}><strong>{displayText(d.key)}</strong>: {displayText(String(d.value || '').slice(0, 80))}{i < 1 && log.data.length > 1 ? ' | ' : ''}</span>
                    ))}
                    {!log.data && displayText((log.message || '').slice(0, 100))}
                  </td>
                  <td style={{ textAlign: 'center', fontSize: 10 }}>{isExpanded ? '▼' : '▶'}</td>
                </tr>
                {isExpanded && (
                  <tr>
                    <td colSpan={4} style={{ padding: '8px 12px', background: 'var(--fx-bg-subtle, #f8fbff)' }}>
                      {jsonParsed ? (
                        <pre style={{ margin: 0, fontSize: 11, whiteSpace: 'pre-wrap', fontFamily: 'ui-monospace, Consolas, monospace' }}>
                          {JSON.stringify(jsonParsed, null, 2)}
                        </pre>
                      ) : (
                        <pre style={{ margin: 0, fontSize: 11, whiteSpace: 'pre-wrap', fontFamily: 'ui-monospace, Consolas, monospace' }}>
                          {logContent}
                        </pre>
                      )}
                    </td>
                  </tr>
                )}
              </React.Fragment>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
function RefsTab({ span }) {
  const refs = span.raw?.refs || []
  if (!refs.length) return <Empty>暂无引用关系</Empty>
  return (
    <table>
      <thead><tr><th>类型</th><th>父 Segment ID</th><th>父 Span ID</th><th>Trace ID</th></tr></thead>
      <tbody>
        {refs.map((ref, idx) => (
          <tr key={idx}>
            <td>{displayText(ref.type || ref.refType || 'CrossProcess')}</td>
            <td style={{ fontFamily: 'monospace', fontSize: 11 }}>{displayText(ref.parentSegmentId)}</td>
            <td>{ref.parentSpanId ?? '-'}</td>
            <td style={{ fontFamily: 'monospace', fontSize: 11 }}>{displayText(ref.traceId)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

function ExceptionsTab({ span }) {
  if (!span.isError) return <Empty>该 Span 无错误信息</Empty>
  const logs = span.raw?.logs || []
  const errorLogs = logs.filter(l => (l.data || []).some(d => d.key === 'error.kind' || d.key === 'message' || d.key === 'stack'))
  const tags = span.raw?.tags || []
  const errorTags = tags.filter(t => t.key && (t.key.includes('error') || t.key.includes('exception') || t.key === 'status_code'))

  return (
    <div>
      {errorTags.length > 0 && (
        <table>
          <thead><tr><th>Key</th><th>Value</th></tr></thead>
          <tbody>
            {errorTags.map((t, i) => <tr key={i}><td>{displayText(t.key)}</td><td style={{ wordBreak: 'break-all', color: 'var(--fx-danger, #c45656)' }}>{displayText(t.value)}</td></tr>)}
          </tbody>
        </table>
      )}
      {errorLogs.length > 0 && errorLogs.map((log, idx) => (
        <div key={idx} style={{ marginTop: 8, padding: 8, background: '#fff7f7', borderRadius: 6, border: '1px solid #f3d2d2' }}>
          {(log.data || []).map((d, i) => (
            <div key={i} style={{ marginBottom: 4 }}>
              <strong style={{ color: '#9f3a38' }}>{displayText(d.key)}</strong>
              <pre style={{ margin: '2px 0 0', whiteSpace: 'pre-wrap', fontSize: 11 }}>{displayText(d.value)}</pre>
            </div>
          ))}
        </div>
      ))}
      {!errorTags.length && !errorLogs.length && (
        <div style={{ color: 'var(--fx-danger)', fontSize: 13 }}>Span 标记为错误，但未找到详细错误信息。</div>
      )}
    </div>
  )
}

export function SpanDetailPanel({ span, onClose, onAddTagToFilter }) {
  const [activeTab, setActiveTab] = useState('tags')
  if (!span) return null

  return (
    <div style={{ marginTop: 10, border: '1px solid var(--fx-border, #e3e8f1)', borderRadius: 8, background: 'var(--fx-panel, #fff)', padding: '10px 12px' }}>
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
        <h4 style={{ margin: 0, fontSize: 14 }}>
          <Status ok={!span.isError}>{span.isError ? '错误' : '正常'}</Status>
          {' '}{displayText(span.endpointName)} · {span.duration} ms
        </h4>
        <button type='button' onClick={onClose} style={{ border: '1px solid var(--fx-border)', borderRadius: 6, background: '#fff', padding: '4px 10px', cursor: 'pointer' }}>关闭</button>
      </header>
      <div style={{ fontSize: 12, color: 'var(--fx-muted)', marginBottom: 6 }}>
        {displayText(span.serviceCode)} · {displayText(span.serviceInstanceName)} · {span.type}
      </div>
      <div className='fx-span-detail-tabs'>
        {TABS.map(tab => (
          <button key={tab.key} type='button' className={activeTab === tab.key ? 'is-active' : ''} onClick={() => setActiveTab(tab.key)}>
            {tab.label}{tab.key === 'exceptions' && span.isError ? ' !' : ''}
          </button>
        ))}
      </div>
      <div className='fx-span-detail-content'>
        {activeTab === 'tags' && <TagsTab span={span} onAddToFilter={onAddTagToFilter} />}
        {activeTab === 'logs' && <LogsTab span={span} />}
        {activeTab === 'refs' && <RefsTab span={span} />}
        {activeTab === 'exceptions' && <ExceptionsTab span={span} />}
      </div>
    </div>
  )
}
