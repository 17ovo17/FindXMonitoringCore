import React, { useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { formatLogError, LOG_BLOCKERS, LOG_SOURCES, logsApi } from '../api/logs.js'
import { displayText } from './logsModel.js'
import { Blocked, Empty, Field, JsonPreview, Status } from './LogsShared.jsx'

export function ContextSection({ query, onOpenTrace, onOpenAgent }) {
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [arrivalLoading, setArrivalLoading] = useState(false)
  const [source, setSource] = useState(query.source || 'findx_audit')
  const [logId, setLogId] = useState(query.logId || '')
  const [traceId, setTraceId] = useState(query.traceId || '')
  const [service, setService] = useState(query.service || query.scope || '')
  const [context, setContext] = useState(null)
  const [arrival, setArrival] = useState(null)

  const agentScope = resolveAgentScope({ service, traceId, logId, context })

  const openTrace = () => {
    const id = traceId || context?.center?.trace_id
    if (!id) {
      setBlocked(LOG_BLOCKERS.traceLink)
      return
    }
    onOpenTrace(id)
  }

  const openAgent = () => {
    if (!agentScope) {
      setBlocked(LOG_BLOCKERS.agentLinkage)
      return
    }
    onOpenAgent({ q: agentScope, packageName: 'logs' })
  }

  const loadContext = async () => {
    setBlocked('')
    setContext(null)
    if (source !== 'findx_audit') {
      setBlocked(LOG_BLOCKERS.context)
      return
    }
    if (!logId && !traceId && !service) {
      setBlocked('请输入日志 ID、Trace ID 或 Scope/服务后再加载上下文。')
      return
    }
    setLoading(true)
    try {
      const resp = await logsApi.context({ source, log_id: logId, trace_id: traceId, scope: service, before: 5, after: 5 })
      setContext(resp)
    } catch (err) {
      setBlocked(formatLogError(err))
    } finally {
      setLoading(false)
    }
  }

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

  return (
    <section className='fx-logs-work'>
      <div className='fx-logs-filter is-logs-context'>
        <Field label='来源'>
          <select value={source} onChange={event => setSource(event.target.value)}>
            {LOG_SOURCES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
          </select>
        </Field>
        <Field label='日志 ID'><input value={logId} onChange={event => setLogId(event.target.value)} placeholder='<LOG_ID>' /></Field>
        <Field label='Trace ID'><input value={traceId} onChange={event => setTraceId(event.target.value)} placeholder='<TRACE_ID>' /></Field>
        <Field label='Scope / 服务'><input value={service} onChange={event => setService(event.target.value)} placeholder='scope / service_name' /></Field>
        <Field label='窗口'><select disabled><option>前后 50 条</option></select></Field>
        <button type='button' onClick={loadContext} disabled={loading}>{loading ? '加载中' : '加载上下文'}</button>
        <button type='button' onClick={openTrace}>打开链路详情</button>
        <button type='button' onClick={openAgent}>进入主机 FindX Agent</button>
        <button type='button' onClick={loadArrival} disabled={arrivalLoading}>{arrivalLoading ? '读取中' : '查看数据到达'}</button>
      </div>
      {blocked && <Blocked>{blocked}</Blocked>}
      {source !== 'findx_audit' && <Blocked>{LOG_BLOCKERS.context}</Blocked>}
      <Blocked>{arrival?.blocker || LOG_BLOCKERS.agentLinkage}</Blocked>
      <LogsArrivalEvidence row={arrival} />
      <div className='fx-logs-split'>
        <div className='fx-logs-panel'>
          <h3>上下文窗口</h3>
          {context?.items?.length ? (
            <div className='fx-logs-table'><table><thead><tr><th>时间</th><th>级别</th><th>动作</th><th>内容</th><th>Trace</th></tr></thead><tbody>{context.items.map(row => <tr key={row.id} className={context.center?.id === row.id ? 'is-current' : ''}><td>{formatTime(row.timestamp)}</td><td>{displayText(row.severity_text)}</td><td>{displayText(row.attributes?.action)}</td><td>{displayText(row.body)}</td><td>{displayText(row.trace_id)}</td></tr>)}</tbody></table></div>
          ) : (
            <Empty>{context ? '未找到匹配日志 ID、Trace ID 或 Scope/服务的 FindX 审计日志上下文。' : '输入日志 ID、Trace ID 或 Scope/服务后加载真实 FindX 审计上下文。'}</Empty>
          )}
        </div>
        <div className='fx-logs-panel'>
          <h3>上下文载荷</h3>
          <JsonPreview value={context || {
            source,
            log_id: logId || '<LOG_ID>',
            before: '<CURSOR_BEFORE>',
            after: '<CURSOR_AFTER>',
            trace_id: traceId || '<TRACE_ID>',
            agent_filter: agentScope || '<SERVICE_OR_TRACE_ID>',
            agent_hosts_url: `/agents?section=hosts&package=logs&q=${encodeURIComponent(agentScope || '')}`,
          }} />
        </div>
      </div>
    </section>
  )
}

function LogsArrivalEvidence({ row }) {
  const status = row?.status || 'blocked'
  const reported = status === 'reported'
  return (
    <div className='fx-logs-table'>
      <table><thead><tr><th>数据通道</th><th>Agent 数</th><th>状态</th><th>最近上报</th><th>证据数</th><th>阻断原因</th></tr></thead>
        <tbody><tr><td>日志</td><td>{row?.agent_count ?? 0}</td><td><Status ok={reported}>{reported ? 'reported' : status}</Status></td><td>{formatTime(row?.last_seen)}</td><td>{row?.evidence_count ?? 0}</td><td>{reported ? '后端已返回日志数据到达 reported。' : row?.blocker || LOG_BLOCKERS.agentLinkage}</td></tr></tbody>
      </table>
    </div>
  )
}

function findLogsArrival(rows) {
  return Array.isArray(rows) ? rows.find(row => row?.kind === 'logs') || null : null
}

function resolveAgentScope({ service, traceId, logId, context }) {
  const center = context?.center || null
  return service || traceId || center?.trace_id || center?.attributes?.scope || ''
}

function formatTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString()
}
