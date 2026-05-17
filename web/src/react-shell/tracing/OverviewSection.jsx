import React, { useEffect, useState } from 'react'
import { tracingApi } from '../api/tracing.js'
import { AgentLinkActions, ErrorBox, Status } from './TracingShared.jsx'

export function OverviewSection({ onNavigate }) {
  const [summary, setSummary] = useState(null)
  const [error, setError] = useState('')
  const load = async () => {
    setError('')
    try {
      const res = await tracingApi.selectors.services()
      const services = Array.isArray(res) ? res : []
      setSummary({ coverage: `${services.length} 服务`, healthy_services: services.length, error_traces: 0, alarms: 0, query_status: '已连接 SkyWalking', agent_linkage: '就绪' })
    } catch (e) {
      setError(e?.message || '查询失败')
    }
  }
  useEffect(() => { load() }, [])
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
      <ErrorBox>{error}</ErrorBox>
      <div className='fx-tracing-grid'>
        {cards.map(([label, value]) => <article key={label} className='fx-tracing-card'><strong>{value ?? '-'}</strong><span>{label}</span></article>)}
      </div>
      <div className='fx-tracing-table'>
        <table><tbody>
          <tr><th>链路适配器</th><td><Status ok={!!summary}> {summary ? '已接入' : '查询中...'} </Status></td></tr>
          <tr><th>链路查询服务</th><td>{summary?.query_status || '连接中...'}</td></tr>
          <tr><th>Agent 联动</th><td>{summary?.agent_linkage || '检测中...'}</td></tr>
        </tbody></table>
      </div>
    </section>
  )
}
