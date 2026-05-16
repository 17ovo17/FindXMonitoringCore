import React, { useState } from 'react'
import { post } from '../../api/http.js'
import {
  displayDate,
  displayJson,
  formatDuration,
  makeError,
  mapToPairs,
  severityLabel,
  statusLabel,
} from './alertModel.js'

const ACTIONS = [
  { value: 'ack', label: '确认', desc: '确认已知晓此告警' },
  { value: 'mute', label: '屏蔽', desc: '临时屏蔽此告警' },
  { value: 'assign', label: '分配', desc: '分配给指定人员处理' },
  { value: 'resolve', label: '关闭', desc: '标记为已解决' },
]

export function AlertEventDetailDrawer({ event, onClose, onActionDone }) {
  const [actionType, setActionType] = useState('')
  const [reason, setReason] = useState('')
  const [assignee, setAssignee] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const handleAction = async (action) => {
    setSubmitting(true); setError('')
    try {
      const body = { action, reason }
      if (action === 'assign') body.assignee = assignee
      await post(`/monitor-alert-events/${encodeURIComponent(event.id)}/action`, body)
      setActionType('')
      setReason('')
      onActionDone?.()
    } catch (err) {
      setError(makeError(err, '操作失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className='fx-alert-drawer'>
      <aside>
        <header>
          <h2>事件详情</h2>
          <button type='button' onClick={onClose}>关闭</button>
        </header>

        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
          <strong style={{ fontSize: 16 }}>{event.name}</strong>
          <span className={`fx-alert-severity is-${event.severity}`}>{severityLabel(event.severity)}</span>
          <span className={`fx-alert-state ${event.status === 'firing' ? 'is-on' : ''}`}>{statusLabel(event.status)}</span>
          <span style={{ color: '#64748b', fontSize: 12, marginLeft: 'auto' }}>
            持续 {formatDuration(event.firstSeen, event.lastSeen || Date.now())}
          </span>
        </div>

        <section className='fx-alert-detail-grid'>
          <article>
            <h3>标签</h3>
            {mapToPairs(event.labels).length > 0
              ? mapToPairs(event.labels).map((item) => <span className='fx-alert-tag' key={item}>{item}</span>)
              : <span style={{ color: '#64748b', fontSize: 12 }}>无标签</span>
            }
          </article>
          <article>
            <h3>注解</h3>
            {mapToPairs(event.annotations).length > 0
              ? mapToPairs(event.annotations).map((item) => <span className='fx-alert-tag' key={item}>{item}</span>)
              : <span style={{ color: '#64748b', fontSize: 12 }}>无注解</span>
            }
          </article>
          <article>
            <h3>当前值</h3>
            <pre style={{ margin: 0 }}>{event.value || '-'}</pre>
          </article>
          <article>
            <h3>处置时间线</h3>
            {event.actionLog && event.actionLog.length > 0 ? (
              <div style={{ maxHeight: 200, overflow: 'auto' }}>
                {event.actionLog.map((log, idx) => (
                  <div key={idx} style={{ padding: '4px 0', borderBottom: '1px solid #f1f5f9', fontSize: 12 }}>
                    <span style={{ color: '#246bfe' }}>{log.action || log.type}</span>
                    {log.user && <span style={{ marginLeft: 8, color: '#475569' }}>{log.user}</span>}
                    {log.time && <span style={{ marginLeft: 8, color: '#94a3b8' }}>{displayDate(log.time)}</span>}
                    {log.reason && <span style={{ marginLeft: 8 }}>{log.reason}</span>}
                  </div>
                ))}
              </div>
            ) : <span style={{ color: '#64748b', fontSize: 12 }}>暂无处置记录</span>}
          </article>
        </section>

        <dl style={{ marginTop: 12 }}>
          <dt>规则 ID</dt><dd>{event.ruleId || '-'}</dd>
          <dt>目标</dt><dd>{event.target || '-'}</dd>
          <dt>数据源</dt><dd>{event.datasourceId || '-'}</dd>
          <dt>业务组</dt><dd>{event.businessGroup || '-'}</dd>
          <dt>首次触发</dt><dd>{displayDate(event.firstSeen)}</dd>
          <dt>最近触发</dt><dd>{displayDate(event.lastSeen)}</dd>
          <dt>指纹</dt><dd>{event.fingerprint || '-'}</dd>
          <dt>确认人</dt><dd>{event.ackBy || '-'}</dd>
          <dt>分配人</dt><dd>{event.assignee || '-'}</dd>
        </dl>

        {error && <div className='fx-alert-message is-error' style={{ marginTop: 12 }}>{error}</div>}

        <div style={{ marginTop: 16, borderTop: '1px solid #e2e8f0', paddingTop: 12 }}>
          <strong style={{ fontSize: 13, display: 'block', marginBottom: 8 }}>操作</strong>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 8 }}>
            {ACTIONS.map((a) => (
              <button key={a.value} type='button' className={actionType === a.value ? 'is-active' : ''} onClick={() => setActionType(a.value)} title={a.desc}>
                {a.label}
              </button>
            ))}
          </div>
          {actionType && (
            <div className='fx-alert-form is-compact' style={{ marginBottom: 8 }}>
              {actionType === 'assign' && (
                <label>
                  <span>分配给</span>
                  <input value={assignee} onChange={(e) => setAssignee(e.target.value)} placeholder='用户名' />
                </label>
              )}
              <label className='is-wide'>
                <span>原因</span>
                <input value={reason} onChange={(e) => setReason(e.target.value)} placeholder='操作原因（可选）' />
              </label>
            </div>
          )}
          {actionType && (
            <button type='button' className='is-primary' disabled={submitting} onClick={() => handleAction(actionType)}>
              {submitting ? '提交中...' : `执行${ACTIONS.find((a) => a.value === actionType)?.label || ''}`}
            </button>
          )}
        </div>
      </aside>
    </div>
  )
}