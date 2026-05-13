import React, { useEffect, useState } from 'react'
import { ASSET_BLOCKERS, assetsApi, formatAssetError } from '../api/assets.js'
import { agentOnline, displayText, fmtTime, hostIp, hostKey, hostName, isHostOnline, sectionSet, sections } from './assetsModel.js'
import { AgentsSection } from './AgentsSection.jsx'
import { BusinessSection } from './BusinessSection.jsx'
import { DatabaseSection } from './DatabaseSection.jsx'
import { DeployTasksSection } from './DeployTasksSection.jsx'
import { HostsSection } from './HostsSection.jsx'
import { ModelDetailSection } from './ModelDetailSection.jsx'
import { ModelTreeSection } from './ModelTreeSection.jsx'
import { GroupTree, ResourceGroupsSection } from './ResourceGroupsSection.jsx'
import { Blocked, ErrorBox, Status } from './Shared.jsx'
import './assets.css'

const CMDB_BUILD = '20260511b'

async function settle(label, fn) {
  try {
    return { label, rows: await fn(), error: '' }
  } catch (error) {
    return { label, rows: [], error: formatAssetError(error) }
  }
}

export function AssetsPage({ query, onNavigate }) {
  const section = sectionSet.has(query?.section) ? query.section : 'overview'
  const meta = sections.find(item => item.value === section) || sections[0]
  const [workspaces, setWorkspaces] = useState([])
  const [groups, setGroups] = useState([])
  const [hosts, setHosts] = useState([])
  const [agents, setAgents] = useState([])
  const [errors, setErrors] = useState({})
  const [loading, setLoading] = useState(false)
  const [reloadToken, setReloadToken] = useState(0)

  const loadAll = async () => {
    setLoading(true)
    const results = await Promise.all([
      settle('business', () => assetsApi.workspaces.list()),
      settle('groups', () => assetsApi.resourceGroups.list()),
      settle('hosts', () => assetsApi.hosts.list()),
      settle('agents', () => assetsApi.agents.list()),
    ])
    const nextErrors = {}
    results.forEach(result => {
      if (result.label === 'business') setWorkspaces(result.rows)
      if (result.label === 'groups') setGroups(result.rows)
      if (result.label === 'hosts') setHosts(result.rows)
      if (result.label === 'agents') setAgents(result.rows)
      if (result.error) nextErrors[result.label] = result.error
    })
    setErrors(nextErrors)
    setLoading(false)
  }

  useEffect(() => { loadAll() }, [reloadToken])
  const refresh = () => setReloadToken(prev => prev + 1)

  return (
    <main className='fx-assets-page' data-build={CMDB_BUILD}>
      <header className='fx-assets-head'>
        <div><p>资产中心</p><h1>{meta.label}</h1><span>{meta.desc}</span></div>
      </header>
      {section === 'overview' && <OverviewSection rows={{ workspaces, groups, hosts, agents }} errors={errors} loading={loading} onNavigate={onNavigate} onRefresh={refresh} />}
      {section === 'business' && <BusinessSection rows={workspaces} error={errors.business} q={query.q || ''} onRefresh={refresh} />}
      {section === 'resource-groups' && <ResourceGroupsSection rows={groups} workspaces={workspaces} error={errors.groups} q={query.q || ''} onRefresh={refresh} />}
      {section === 'hosts' && <HostsSection groups={groups} workspaces={workspaces} initialQuery={query} onRefreshAll={refresh} />}
      {section === 'agents' && <AgentsSection rows={agents} hosts={hosts} error={errors.agents} q={query.q || ''} onRefresh={refresh} />}
      {section === 'databases' && <DatabaseSection onRefresh={refresh} />}
      {section === 'deploy-tasks' && <DeployTasksSection />}
      {section === 'models' && <ModelTreeSection onNavigate={onNavigate} />}
      {section === 'model-detail' && <ModelDetailSection query={query} onNavigate={onNavigate} />}
      {section === 'cmdb' && <CmdbSection groups={groups} workspaces={workspaces} initialQuery={query} onRefreshAll={refresh} />}
    </main>
  )
}

function OverviewSection({ rows, errors, loading, onNavigate, onRefresh }) {
  const onlineAgents = rows.agents.filter(agentOnline).length
  const cards = [['业务组', rows.workspaces.length], ['资源组', rows.groups.length], ['主机资产', rows.hosts.length], ['在线 FindX Agent', `${onlineAgents}/${rows.agents.length}`]]
  return (
    <section className='fx-assets-work'>
      <div className='fx-assets-toolbar'>
        <button type='button' onClick={onRefresh}>{loading ? '刷新中...' : '刷新'}</button>
        <button type='button' onClick={() => onNavigate({ section: 'cmdb' })}>进入 CMDB</button>
        <button type='button' onClick={() => onNavigate({ section: 'databases' })}>数据库资产</button>
        <button type='button' onClick={() => onNavigate({ section: 'deploy-tasks' })}>部署任务</button>
        <button type='button' onClick={() => onNavigate({ section: 'agents' })}>Agent 管理中心</button>
      </div>
      <div className='fx-assets-grid'>{cards.map(([label, value]) => <article className='fx-assets-card' key={label}><strong>{value}</strong><span>{label}</span></article>)}</div>
      {Object.values(errors).map(error => <ErrorBox key={error}>{error}</ErrorBox>)}
      <Blocked>{ASSET_BLOCKERS.terminal}</Blocked><Blocked>{ASSET_BLOCKERS.agentLifecycle}</Blocked>
      <HostSnapshot rows={rows.hosts} agents={rows.agents} />
    </section>
  )
}

function HostSnapshot({ rows, agents }) {
  const recent = rows.slice(0, 5)
  return (
    <div className='fx-assets-table' style={{ marginTop: 12 }}>
      <table><thead><tr><th>主机</th><th>IP</th><th>状态</th><th>FindX Agent</th><th>最近心跳</th></tr></thead>
        <tbody>{recent.map(row => <tr key={hostKey(row)}><td>{hostName(row)}</td><td>{hostIp(row)}</td><td><Status ok={isHostOnline(row)}>{isHostOnline(row) ? '在线' : '离线'}</Status></td><td>{displayText(row.agent_id || row.agent_status)}</td><td>{fmtTime(row.last_seen_at || row.last_seen || row.updated_at)}</td></tr>)}
          {!recent.length && <tr><td colSpan='5'>暂无主机资产。FindX Agent 心跳或主机契约接入后会自动出现。</td></tr>}</tbody>
      </table>
      <p className='fx-assets-muted'>当前 Agent 记录：{agents.length} 条。</p>
    </div>
  )
}

function CmdbSection({ groups, workspaces, initialQuery, onRefreshAll }) {
  const [selected, setSelected] = useState(initialQuery.group || '')
  return (
    <section className='fx-assets-split'>
      <GroupTree rows={groups} selected={selected} onSelect={setSelected} onCreateRoot={() => {}} onCreateChild={() => {}} onEdit={() => {}} onDelete={() => {}} readonly />
      <div className='fx-assets-detail'>
        <Blocked>{ASSET_BLOCKERS.terminal}</Blocked>
        <Blocked>{ASSET_BLOCKERS.monitor}</Blocked>
        <HostsSection groups={groups} workspaces={workspaces} initialQuery={{ ...initialQuery, group: selected }} onRefreshAll={onRefreshAll} embedded />
      </div>
    </section>
  )
}
