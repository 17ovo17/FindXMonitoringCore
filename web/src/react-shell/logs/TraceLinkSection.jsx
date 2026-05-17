import React, { useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { LOG_BLOCKERS } from '../api/logs.js'
import { Blocked, Field, JsonPreview, Status } from './LogsShared.jsx'

export function TraceLinkSection({ query, onOpenTrace, onOpenAgent }) {
  const [traceId, setTraceId] = useState(query.traceId || '')
  const [spanId, setSpanId] = useState(query.spanId || '')
  const [service, setService] = useState(query.service || '')
  const [blocked, setBlocked] = useState('')
  const [arrival, setArrival] = useState(null)
  const [loadingArrival, setLoadingArrival] = useState(false)

  const scope = service || traceId || spanId

  const openTrace = () => {
    if (!traceId) {
      setBlocked(LOG_BLOCKERS.traceLink)
      return
    }
    onOpenTrace(traceId)
  }

  const openAgent = () => {
    onOpenAgent({ q: scope, packageName: 'logs' })
  }

  const loadArrival = async () => {
    setBlocked('')
    setLoadingArrival(true)
    try {
      const rows = await agentApi.dataArrival()
      const logsRow = findLogsArrival(rows)
      setArrival(logsRow)
      if (!logsRow) setBlocked('FindX Agent 数据到达契约未返回日志通道。')
      if (logsRow && logsRow.status !== 'reported') setBlocked(logsRow.blocker || LOG_BLOCKERS.agentLinkage)
    } catch (err) {
      setArrival(null)
      setBlocked(formatAgentError(err))
    } finally {
      setLoadingArrival(false)
    }
  }

  return (
    <section className='fx-logs-work'>
      <div className='fx-logs-filter is-logs-trace'>
        <Field label='Trace ID'><input value={traceId} onChange={event => setTraceId(event.target.value)} placeholder='<TRACE_ID>' /></Field>
        <Field label='Span ID'><input value={spanId} onChange={event => setSpanId(event.target.value)} placeholder='<SPAN_ID>' /></Field>
        <Field label='服务 / 主机'><input value={service} onChange={event => setService(event.target.value)} placeholder='service_name / host_name' /></Field>
        <button type='button' onClick={openTrace}>打开链路详情</button>
        <button type='button' onClick={openAgent}>进入主机 FindX Agent</button>
        <button type='button' onClick={loadArrival} disabled={loadingArrival}>{loadingArrival ? '读取中' : '查看日志采集覆盖'}</button>
      </div>
      {blocked && <Blocked>{blocked}</Blocked>}
      <Blocked>{LOG_BLOCKERS.traceLink}</Blocked>
      <Blocked>{arrival?.blocker || LOG_BLOCKERS.agentLinkage}</Blocked>
      <LogsArrivalEvidence row={arrival} />
      <JsonPreview value={{
        trace_id: traceId || '<TRACE_ID>',
        span_id: spanId || '<SPAN_ID>',
        agent_filter: scope || '<SERVICE_OR_TRACE_ID>',
        agent_hosts_url: `/agents?section=hosts&package=logs&q=${encodeURIComponent(scope || '')}`,
        permission: 'FindX permission check required',
      }} />
    </section>
  )
}

function LogsArrivalEvidence({ row }) {
  const status = row?.status || 'blocked'
  const reported = status === 'reported'
  return (
    <div className='fx-logs-table'>
      <table>
        <thead><tr><th>数据通道</th><th>Agent 数</th><th>状态</th><th>最近上报</th><th>证据数</th><th>阻断原因</th></tr></thead>
        <tbody>
          <tr>
            <td>日志</td>
            <td>{row?.agent_count ?? 0}</td>
            <td><Status ok={reported}>{reported ? 'reported' : status}</Status></td>
            <td>{formatTime(row?.last_seen)}</td>
            <td>{row?.evidence_count ?? 0}</td>
            <td>{reported ? '后端已返回日志数据到达 reported。' : row?.blocker || LOG_BLOCKERS.agentLinkage}</td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}

function findLogsArrival(rows) {
  return Array.isArray(rows) ? rows.find(row => row?.kind === 'logs') || null : null
}

function formatTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString()
}
