import React from 'react'
import { AGENT_BLOCKERS } from '../api/agents.js'
import { agentOnline, displayText, fmtTime } from './agentModel.js'
import { Blocked, Empty, Status } from './AgentShared.jsx'

export function HeartbeatSection({ rows }) {
  const online = rows.filter(agentOnline).length
  return (
    <section className='fx-agent-work'>
      <div className='fx-agent-grid'>
        <article className='fx-agent-card'><strong>{rows.length}</strong><span>注册 Agent</span></article>
        <article className='fx-agent-card'><strong>{online}</strong><span>在线</span></article>
        <article className='fx-agent-card'><strong>{rows.length - online}</strong><span>离线</span></article>
        <article className='fx-agent-card'><strong>-</strong><span>服务注册详情</span></article>
      </div>
      <Blocked>{AGENT_BLOCKERS.heartbeat}</Blocked>
      <div className='fx-agent-table'>
        <table><thead><tr><th>主机</th><th>状态</th><th>进程</th><th>版本漂移</th><th>最后心跳</th></tr></thead>
          <tbody>{rows.map(row => <tr key={row.id || row.ip || row.ident}><td>{displayText(row.hostname || row.ip || row.ident)}</td><td><Status ok={agentOnline(row)}>{agentOnline(row) ? '在线' : '离线'}</Status></td><td>{displayText(row.process_status)}</td><td>{displayText(row.version_drift)}</td><td>{fmtTime(row.last_seen || row.last_seen_at || row.updated_at)}</td></tr>)}</tbody>
        </table>
        {!rows.length && <Empty>暂无心跳记录。</Empty>}
      </div>
    </section>
  )
}
