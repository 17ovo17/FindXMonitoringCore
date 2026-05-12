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
  if (!logs.length) return <Empty>暂无 Span 日志</Empty>
  return (
    <table>
      <thead><tr><th>时间</th><th>内容</th></tr></thead>
      <tbody>
        {logs.map((log, idx) => (
          <tr key={idx}>
            <td style={{ whiteSpace: 'nowrap' }}>{fmtTime(log.time || log.timestamp)}</td>
            <td style={{ wordBreak: 'break-all' }}>
              {(log.data || []).map((d, i) => (
                <div key={i}><strong>{displayText(d.key)}</strong>: {displayText(d.value)}</div>
              ))}
              {!log.data && displayText(log.message || JSON.stringify(log))}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
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
