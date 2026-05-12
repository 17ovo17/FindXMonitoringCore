import React from 'react'
import { AGENT_BLOCKERS } from '../api/agents.js'
import { agentOnline, capabilityPackages } from './agentModel.js'
import { Blocked, Status } from './AgentShared.jsx'

const normalizePackage = item => ({
  ...item,
  capabilityDomain: item.capabilityDomain || item.capability_domain || '未分组',
  sourceState: item.source_state || item.sourceState,
})

export function OverviewSection({ agents, packages, lifecycle, onNavigate }) {
  const online = agents.filter(agentOnline).length
  const packageRows = (packages?.length ? packages : capabilityPackages).map(normalizePackage)
  const domains = [...new Set(packageRows.map(item => item.capabilityDomain))]
  const blockers = [AGENT_BLOCKERS.packageLifecycle, AGENT_BLOCKERS.installLifecycle, AGENT_BLOCKERS.configLifecycle, AGENT_BLOCKERS.traceLinkage]
  const phases = Array.isArray(lifecycle?.phases) ? lifecycle.phases : []

  return (
    <section className='fx-agent-work'>
      <div className='fx-agent-toolbar'>
        <button type='button' onClick={() => onNavigate({ section: 'packages' })}>能力包矩阵</button>
        <button type='button' onClick={() => onNavigate({ section: 'templates' })}>统一配置下发</button>
        <button type='button' onClick={() => onNavigate({ section: 'install' })}>安装向导</button>
        <button type='button' onClick={() => onNavigate({ section: 'hosts' })}>主机 Agent</button>
      </div>
      <div className='fx-agent-grid'>
        <article className='fx-agent-card'><strong>{packageRows.length}</strong><span>能力包类型</span></article>
        <article className='fx-agent-card'><strong>{domains.length}</strong><span>统一能力域</span></article>
        <article className='fx-agent-card'><strong>{online}/{agents.length}</strong><span>在线 Agent</span></article>
        <article className='fx-agent-card'><strong>{packageRows.filter(item => item.signature === 'ready').length}</strong><span>签名证据</span></article>
      </div>
      <div className='fx-agent-domain-strip'>
        {domains.map(item => <span key={item}>{item}<strong>{packageRows.filter(row => row.capabilityDomain === item).length}</strong></span>)}
      </div>
      <div className='fx-agent-split'>
        <div className='fx-agent-panel'>
          <h3>能力状态</h3>
          <table><tbody>
            {phases.length ? phases.map(item => <tr key={item.key}><td>{item.name}</td><td><Status ok={item.status === 'ready'}>{item.status === 'ready' ? '已具备证据' : '缺契约'}</Status></td></tr>) : (
              <>
                <tr><td>能力包矩阵</td><td><Status>待补齐本地包证据</Status></td></tr>
                <tr><td>内置包仓库</td><td><Status>缺契约</Status></td></tr>
                <tr><td>接入向导</td><td><Status>占位模板，待契约</Status></td></tr>
                <tr><td>链路联动</td><td><Status>缺契约</Status></td></tr>
              </>
            )}
          </tbody></table>
        </div>
        <div className='fx-agent-panel'>
          <h3>阻断契约</h3>
          {blockers.map(item => <Blocked key={item}>{item}</Blocked>)}
        </div>
      </div>
    </section>
  )
}
