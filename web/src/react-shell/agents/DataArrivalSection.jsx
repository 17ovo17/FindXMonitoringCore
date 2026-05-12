import React, { useEffect, useState } from 'react'
import { AGENT_BLOCKERS, agentApi, formatAgentError } from '../api/agents.js'
import { Blocked, Status } from './AgentShared.jsx'
import { buildDataArrivalRows } from './HostDataArrivalEvidence.jsx'

export function DataArrivalSection() {
  const [rows, setRows] = useState([])
  const [blocked, setBlocked] = useState('')
  useEffect(() => {
    let alive = true
    agentApi.dataArrival()
      .then(value => { if (alive) { setRows(value); setBlocked('') } })
      .catch(err => { if (alive) setBlocked(formatAgentError(err)) })
    return () => { alive = false }
  }, [])
  const displayRows = buildDataArrivalRows(rows)
  return (
    <section className='fx-agent-work'>
      <Blocked>{blocked || AGENT_BLOCKERS.dataArrival}</Blocked>
      <div className='fx-agent-table'>
        <table><thead><tr><th>数据通道</th><th>验收点</th><th>当前状态</th><th>阻断原因</th></tr></thead>
          <tbody>{displayRows.map(row => <tr key={row.kind}><td>{row.label}</td><td>{row.acceptance}</td><td><Status ok={row.status === 'reported'}>{row.status === 'reported' ? `reported ${row.count}` : 'blocked'}</Status></td><td>{row.status === 'reported' ? `已通过 receiver evidence 验证，最近 ${row.lastSeen || '-'}` : row.blocker}</td></tr>)}</tbody>
        </table>
      </div>
      <Blocked>{AGENT_BLOCKERS.traceLinkage}</Blocked>
    </section>
  )
}
