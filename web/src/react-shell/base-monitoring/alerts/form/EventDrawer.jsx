import React, { useState } from 'react'
import {
  displayDate,
  displayJson,
  formatDuration,
  mapToPairs,
  severityLabel,
  statusLabel,
} from '../alertModel.js'

/**
 * 事件详情抽屉组件
 */
export function EventDrawer({ event, onClose, onAction }) {
  const [assignee, setAssignee] = useState(event.assignee || '')
  const [reason, setReason] = useState('')
  return (
    <div className='fx-alert-drawer'>
      <aside>
        <header><h2>事件详情</h2><button type='button' onClick={onClose}>关闭</button></header>
        <dl>
          <dt>名称</dt><dd>{event.name}</dd>
          <dt>状态</dt><dd>{statusLabel(event.status)}</dd>
          <dt>级别</dt><dd>{severityLabel(event.severity)}</dd>
          <dt>目标</dt><dd>{event.target}</dd>
          <dt>业务组</dt><dd>{event.businessGroup}</dd>
          <dt>分类</dt><dd>{event.category}</dd>
          <dt>数据源</dt><dd>{event.datasourceId}</dd>
          <dt>首次触发</dt><dd>{displayDate(event.firstSeen)}</dd>
          <dt>最近触发</dt><dd>{displayDate(event.lastSeen)}</dd>
          <dt>持续时间</dt><dd>{formatDuration(event.firstSeen, event.lastSeen || Date.now())}</dd>
          <dt>指纹</dt><dd>{event.fingerprint || '-'}</dd>
        </dl>
        <section className='fx-alert-detail-grid'>
          <article><h3>标签</h3>{mapToPairs(event.labels).map((item) => <span className='fx-alert-tag' key={item}>{item}</span>)}</article>
          <article><h3>注解</h3>{mapToPairs(event.annotations).map((item) => <span className='fx-alert-tag' key={item}>{item}</span>)}</article>
          <article><h3>处置记录</h3><pre>{displayJson(event.actionLog)}</pre></article>
        </section>
        <div className='fx-alert-form is-compact'>
          <label><span>分派给</span><input value={assignee} onChange={(e) => setAssignee(e.target.value)} /></label>
          <label><span>原因</span><input value={reason} onChange={(e) => setReason(e.target.value)} /></label>
        </div>
        <div className='fx-alert-actions'>
          <button type='button' onClick={() => onAction('ack', event, { reason })}>确认</button>
          <button type='button' onClick={() => onAction('assign', event, { assignee, reason })}>分派</button>
          <button type='button' onClick={() => onAction('resolve', event, { reason })}>恢复</button>
          <button type='button' onClick={() => onAction('archive', event, { reason })}>归档</button>
          <button type='button' onClick={() => onAction('mute', event, {})}>屏蔽</button>
          <button type='button' onClick={() => onAction('delete', event, {})}>删除</button>
        </div>
      </aside>
    </div>
  )
}
