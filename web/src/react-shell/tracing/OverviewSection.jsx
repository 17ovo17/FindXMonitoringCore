import React, { useState } from 'react'
import { TRACING_BLOCKERS } from '../api/tracing.js'
import { AgentLinkActions, Blocked, ErrorBox, Status } from './TracingShared.jsx'

export function OverviewSection({ onNavigate }) {
  const [summary, setSummary] = useState(null)
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const load = async () => {
    setSummary(null); setError(''); setBlocked(TRACING_BLOCKERS.overview)
  }
  const cards = [
    ['采集覆盖', summary?.coverage],
    ['服务健康', summary?.healthy_services],
    ['错误 Trace', summary?.error_traces],
    ['链路告警', summary?.alarms],
  ]
  return (
    <section className='fx-tracing-work'>
      <div className='fx-tracing-toolbar'>
        <button type='button' onClick={load}>刷新</button>
        <button type='button' onClick={() => onNavigate({ section: 'services' })}>服务目录</button>
        <button type='button' onClick={() => onNavigate({ section: 'traces' })}>Trace 检索</button>
        <AgentLinkActions onNavigate={onNavigate} />
      </div>
      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}
      <Blocked>{TRACING_BLOCKERS.agentLinkage}</Blocked>
      <div className='fx-tracing-grid'>
        {cards.map(([label, value]) => <article key={label} className='fx-tracing-card'><strong>{value ?? 'BLOCKED'}</strong><span>{label}</span></article>)}
      </div>
      <div className='fx-tracing-table'>
        <table><tbody>
          <tr><th>链路适配器</th><td><Status ok={!!summary}> {summary ? '已接入' : 'BLOCKED_BY_CONTRACT'} </Status></td></tr>
          <tr><th>链路查询服务</th><td>{summary?.query_status || '缺少代理契约'}</td></tr>
          <tr><th>Agent 联动</th><td>{summary?.agent_linkage || '缺少服务覆盖率和探针反查契约'}</td></tr>
        </tbody></table>
      </div>
    </section>
  )
}
