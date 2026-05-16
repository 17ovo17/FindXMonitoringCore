import React, { useEffect, useState } from 'react'
import { ASSET_BLOCKERS, assetsApi, formatAssetError } from '../api/assets.js'
import { cmdbApi, isCmdbContractAudit } from '../api/cmdb.js'
import { agentOnline, displayText, fmtTime, hostIp, hostKey, hostName, isHostOnline, sectionSet, sections } from './assetsModel.js'
import { AgentsSection } from './AgentsSection.jsx'
import { BusinessSection } from './BusinessSection.jsx'
import { DatabaseSection } from './DatabaseSection.jsx'
import { DeployTasksSection } from './DeployTasksSection.jsx'
import { HostsSection } from './HostsSection.jsx'
import { InstanceDetailContractSection, RelationGraphSection, TopologySection } from './RelationTopologySection.jsx'
import { ModelDetailSection } from './ModelDetailSection.jsx'
import { ModelTreeSection } from './ModelTreeSection.jsx'
import { GroupTree, ResourceGroupsSection } from './ResourceGroupsSection.jsx'
import { DatacenterView } from './DatacenterView.jsx'
import { Blocked, ErrorBox, Status } from './Shared.jsx'
import './assets.css'

const CMDB_BUILD = '20260511b'
const blockedRouteSections = [
  'recycle-bin',
  'room-view',
  'model-relations',
  'attribute-units',
  'auto-discovery',
  'auto-mapping',
  'discovery-records',
  'change-stats',
  'model-change',
  'instance-change-top',
  'custom-report',
  'cloud-bill',
  'compliance-check',
  'compliance-stats',
  'change-notice',
  'spare-management',
  'inspection-config',
  'relationship-query',
  'event-subscription',
  'notice-records',
  'change-records',
  'subscription-records',
]
const approvalRouteViews = {
  'approval-mine': 'mine',
  'approval-todo': 'todo',
  'approval-archive': 'archive',
}
const routeSections = new Set([...sectionSet, 'search', 'topology', 'instance-detail', 'resource-stats', 'datacenter-view', ...Object.keys(approvalRouteViews), ...blockedRouteSections])
const routeMeta = {
  topology: { label: '\u0043\u004d\u0044\u0042 \u5173\u7cfb\u62d3\u6251', desc: '\u6309\u5b9e\u4f8b\u3001\u4e1a\u52a1\u4e0a\u4e0b\u6587\u548c\u5ba1\u8ba1\u5951\u7ea6\u5c55\u793a\u5173\u7cfb\u3002' },
  'instance-detail': { label: '\u0043\u004d\u0044\u0042 \u5b9e\u4f8b\u8be6\u60c5', desc: '\u5c55\u793a\u5b9e\u4f8b\u8be6\u60c5\u3001\u5173\u7cfb\u53cd\u67e5\u548c\u76d1\u63a7\u7ed1\u5b9a\u5951\u7ea6\u3002' },
  'datacenter-view': { label: '机房视图', desc: '按机房、机柜、U位展示物理资源布局和容量。' },
}
const blockedRouteMeta = {
  search: { label: '\u5168\u6587\u68c0\u7d22', desc: '\u9700\u8981\u771f\u5b9e\u7d22\u5f15\u3001\u6743\u9650\u8fc7\u6ee4\u548c\u5ba1\u8ba1\u56de\u6267\u3002' },
  'room-view': { label: '\u673a\u623f\u89c6\u56fe', desc: '\u9700\u8981\u771f\u5b9e\u673a\u623f\u3001\u673a\u67dc\u3001\u673a\u4f4d\u548c\u5bb9\u91cf\u6570\u636e\u6e90\u3002' },
  'recycle-bin': { label: '\u56de\u6536\u7ad9', desc: '\u9700\u8981\u5220\u9664\u5ba1\u8ba1\u3001\u6062\u590d\u56de\u6267\u548c\u4fdd\u7559\u7b56\u7565\u5951\u7ea6\u3002' },
  'model-relations': { label: '\u5173\u8054\u7c7b\u578b', desc: '\u9700\u8981\u6a21\u578b\u5173\u7cfb\u7c7b\u578b\u3001\u65b9\u5411\u548c\u53d8\u66f4\u5ba1\u8ba1\u5951\u7ea6\u3002' },
  'attribute-units': { label: '\u5c5e\u6027\u5355\u4f4d', desc: '\u9700\u8981\u5c5e\u6027\u5355\u4f4d\u5b57\u5178\u548c\u53d8\u66f4\u56de\u6267\u5951\u7ea6\u3002' },
  'auto-discovery': { label: '\u81ea\u52a8\u53d1\u73b0', desc: '\u9700\u8981\u53d1\u73b0\u89c4\u5219\u3001\u4efb\u52a1\u6267\u884c\u5668\u548c\u8d44\u6e90\u5199\u5165\u5951\u7ea6\u3002' },
  'auto-mapping': { label: '\u81ea\u52a8\u5316\u6620\u5c04', desc: '\u9700\u8981\u6620\u5c04\u89c4\u5219\u3001\u51b2\u7a81\u7b56\u7565\u548c\u5ba1\u6279\u56de\u6267\u5951\u7ea6\u3002' },
  'discovery-records': { label: '\u6267\u884c\u8bb0\u5f55', desc: '\u9700\u8981\u53d1\u73b0\u4efb\u52a1\u3001\u6267\u884c\u65e5\u5fd7\u548c\u5ba1\u8ba1\u67e5\u8be2\u5951\u7ea6\u3002' },
  'approval-mine': { label: '\u6211\u7684\u7533\u8bf7', desc: '\u8bfb\u53d6\u540e\u7aef\u771f\u5b9e\u5ba1\u6279\u8bf7\u6c42\u3002' },
  'approval-todo': { label: '\u6211\u7684\u5f85\u529e', desc: '\u8bfb\u53d6\u540e\u7aef\u771f\u5b9e\u5f85\u529e\u5ba1\u6279\u3002' },
  'approval-archive': { label: '\u5df2\u5f52\u6863', desc: '\u8bfb\u53d6\u540e\u7aef\u771f\u5b9e\u5f52\u6863\u5ba1\u6279\u3002' },
  'resource-stats': { label: '\u8d44\u6e90\u7edf\u8ba1', desc: '\u57fa\u4e8e CMDB \u771f\u5b9e\u805a\u5408\u5c55\u793a\u8d44\u6e90\u603b\u89c8\u3002' },
  'change-stats': { label: '\u53d8\u66f4\u7edf\u8ba1', desc: '\u9700\u8981\u53d8\u66f4\u4e8b\u4ef6\u3001\u5f71\u54cd\u8303\u56f4\u548c\u5ba1\u8ba1\u805a\u5408\u5951\u7ea6\u3002' },
  'model-change': { label: '\u6a21\u578b\u53d8\u66f4', desc: '\u9700\u8981\u6a21\u578b\u7248\u672c\u3001\u5dee\u5f02\u5b57\u6bb5\u548c\u56de\u6eda\u5f71\u54cd\u5951\u7ea6\u3002' },
  'instance-change-top': { label: '\u5b9e\u4f8b\u53d8\u66f4TOP', desc: '\u9700\u8981\u771f\u5b9e\u53d8\u66f4\u6392\u884c\u548c\u5ba1\u8ba1\u6765\u6e90\u5951\u7ea6\u3002' },
  'custom-report': { label: '\u81ea\u5b9a\u4e49\u62a5\u8868', desc: '\u9700\u8981\u62a5\u8868\u5b9a\u4e49\u3001\u6743\u9650\u6821\u9a8c\u548c\u5bfc\u51fa\u56de\u6267\u5951\u7ea6\u3002' },
  'cloud-bill': { label: '\u4e91\u5e73\u53f0\u8d26\u5355', desc: '\u9700\u8981\u8d26\u5355\u5bfc\u5165\u3001\u8d44\u6e90\u6620\u5c04\u548c\u8131\u654f\u5951\u7ea6\u3002' },
  'compliance-check': { label: '\u5408\u89c4\u6027\u68c0\u67e5', desc: '\u9700\u8981\u89c4\u5219\u96c6\u3001\u68c0\u67e5\u6267\u884c\u5668\u548c\u7ed3\u679c\u56de\u6267\u5951\u7ea6\u3002' },
  'compliance-stats': { label: '\u5408\u89c4\u6027\u7edf\u8ba1', desc: '\u9700\u8981\u5408\u89c4\u7ed3\u679c\u805a\u5408\u548c\u8d8b\u52bf\u5951\u7ea6\u3002' },
  'change-notice': { label: '\u53d8\u66f4\u901a\u77e5', desc: '\u9700\u8981\u8ba2\u9605\u89c4\u5219\u3001\u901a\u77e5\u6e20\u9053\u548c\u53d1\u9001\u56de\u6267\u5951\u7ea6\u3002' },
  'spare-management': { label: '\u5907\u4ef6\u7ba1\u7406', desc: '\u9700\u8981\u5907\u4ef6\u5e93\u5b58\u3001\u9886\u7528\u5ba1\u6279\u548c\u5ba1\u8ba1\u5951\u7ea6\u3002' },
  'inspection-config': { label: '\u5de1\u68c0\u914d\u7f6e', desc: '\u9700\u8981\u5de1\u68c0\u6a21\u677f\u3001\u76ee\u6807\u9009\u62e9\u5668\u548c\u7ed3\u679c\u56de\u6267\u5951\u7ea6\u3002' },
  'relationship-query': { label: '\u5173\u7cfb\u67e5\u8be2', desc: '\u9700\u8981\u9012\u5f52\u5173\u7cfb\u8def\u5f84\u3001\u6df1\u5ea6\u9650\u5236\u548c\u6743\u9650\u8fc7\u6ee4\u5951\u7ea6\u3002' },
  'event-subscription': { label: '\u4e8b\u4ef6\u8ba2\u9605', desc: '\u9700\u8981\u4e8b\u4ef6\u6e90\u3001\u8ba2\u9605\u89c4\u5219\u548c\u53d1\u9001\u56de\u6267\u5951\u7ea6\u3002' },
  'notice-records': { label: '\u901a\u77e5\u8bb0\u5f55', desc: '\u9700\u8981\u901a\u77e5\u6d41\u6c34\u3001\u6e20\u9053\u7ed3\u679c\u548c\u5ba1\u8ba1\u67e5\u8be2\u5951\u7ea6\u3002' },
  'change-records': { label: '\u53d8\u66f4\u8bb0\u5f55', desc: '\u9700\u8981\u5b9e\u4f8b\u53d8\u66f4\u3001\u5b57\u6bb5\u5dee\u5f02\u548c\u5ba1\u8ba1\u67e5\u8be2\u5951\u7ea6\u3002' },
  'subscription-records': { label: '\u8ba2\u9605\u8bb0\u5f55', desc: '\u9700\u8981\u8ba2\u9605\u5bf9\u8c61\u3001\u4e8b\u4ef6\u7c7b\u578b\u548c\u5ba1\u8ba1\u67e5\u8be2\u5951\u7ea6\u3002' },
}

async function settle(label, fn) {
  try {
    return { label, rows: await fn(), error: '' }
  } catch (error) {
    return { label, rows: [], error: formatAssetError(error) }
  }
}

export function AssetsPage({ query, onNavigate }) {
  const section = routeSections.has(query?.section) ? query.section : 'overview'
  const meta = blockedRouteMeta[section] || routeMeta[section] || sections.find(item => item.value === section) || sections[0]
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
        <div><p>{'\u8d44\u4ea7\u4e2d\u5fc3'}</p><h1>{meta.label}</h1><span>{meta.desc}</span></div>
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
      {section === 'topology' && <TopologySection query={query} />}
      {section === 'instance-detail' && <InstanceDetailContractSection query={query} onNavigate={onNavigate} />}
      {section === 'resource-stats' && <ResourceStatsSection />}
      {section === 'datacenter-view' && <DatacenterView />}
      {approvalRouteViews[section] && <ResourceApprovalSection section={section} meta={meta} />}
      {(section === 'search' || blockedRouteSections.includes(section)) && <CmdbCapabilityBlockedSection section={section} meta={meta} />}
    </main>
  )
}

function OverviewSection({ rows, errors, loading, onNavigate, onRefresh }) {
  const onlineAgents = rows.agents.filter(agentOnline).length
  const cards = [['\u4e1a\u52a1\u7ec4', rows.workspaces.length], ['\u8d44\u6e90\u7ec4', rows.groups.length], ['\u4e3b\u673a\u8d44\u4ea7', rows.hosts.length], ['\u5728\u7ebf FindX Agent', onlineAgents + '/' + rows.agents.length]]
  return (
    <section className='fx-assets-work'>
      <div className='fx-assets-toolbar'>
        <button type='button' onClick={onRefresh}>{loading ? '\u5237\u65b0\u4e2d...' : '\u5237\u65b0'}</button>
        <button type='button' onClick={() => onNavigate({ section: 'cmdb' })}>{'\u8fdb\u5165 CMDB'}</button>
        <button type='button' onClick={() => onNavigate({ section: 'databases' })}>{'\u6570\u636e\u5e93\u8d44\u4ea7'}</button>
        <button type='button' onClick={() => onNavigate({ section: 'deploy-tasks' })}>{'\u90e8\u7f72\u4efb\u52a1'}</button>
        <button type='button' onClick={() => onNavigate({ section: 'agents' })}>{'Agent \u7ba1\u7406\u4e2d\u5fc3'}</button>
        <button type='button' onClick={() => onNavigate({ section: 'datacenter-view' })}>{'\u673a\u623f\u89c6\u56fe'}</button>
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
      <table><thead><tr><th>{'\u4e3b\u673a'}</th><th>IP</th><th>{'\u72b6\u6001'}</th><th>FindX Agent</th><th>{'\u6700\u8fd1\u5fc3\u8df3'}</th></tr></thead>
        <tbody>{recent.map(row => <tr key={hostKey(row)}><td>{hostName(row)}</td><td>{hostIp(row)}</td><td><Status ok={isHostOnline(row)}>{isHostOnline(row) ? '\u5728\u7ebf' : '\u79bb\u7ebf'}</Status></td><td>{displayText(row.agent_id || row.agent_status)}</td><td>{fmtTime(row.last_seen_at || row.last_seen || row.updated_at)}</td></tr>)}
          {!recent.length && <tr><td colSpan='5'>{'\u6682\u65e0\u4e3b\u673a\u8d44\u4ea7\u3002FindX Agent \u5fc3\u8df3\u6216\u4e3b\u673a\u5951\u7ea6\u63a5\u5165\u540e\u4f1a\u81ea\u52a8\u51fa\u73b0\u3002'}</td></tr>}</tbody>
      </table>
      <p className='fx-assets-muted'>{'\u5f53\u524d Agent \u8bb0\u5f55\uff1a'}{agents.length}{' \u6761\u3002'}</p>
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
        <RelationGraphSection groupId={selected} />
        <HostsSection groups={groups} workspaces={workspaces} initialQuery={{ ...initialQuery, group: selected }} onRefreshAll={onRefreshAll} embedded />
      </div>
    </section>
  )
}

function ResourceStatsSection() {
  const [state, setState] = useState({ loading: true, data: null, error: '' })
  const loadStats = async () => {
    setState(prev => ({ ...prev, loading: true, error: '' }))
    try {
      const data = await cmdbApi.resourceStats()
      setState({ loading: false, data, error: '' })
    } catch (error) {
      setState({ loading: false, data: null, error: formatAssetError(error) })
    }
  }

  useEffect(() => { loadStats() }, [])

  const audit = isCmdbContractAudit(state.data) ? state.data : null
  return (
    <section className='fx-assets-work fx-resource-stats'>
      <div className='fx-resource-stats-head'>
        <div>
          <h2>{'\u8d44\u6e90\u7edf\u8ba1'}</h2>
          <p>{'\u76f4\u63a5\u8bfb\u53d6 CMDB \u805a\u5408\u63a5\u53e3\uff0c\u7f3a\u5c11\u53ef\u6838\u9a8c\u6570\u636e\u65f6\u4fdd\u6301\u5951\u7ea6\u963b\u65ad\u3002'}</p>
        </div>
        <button type='button' className='fx-assets-button is-primary' onClick={loadStats} disabled={state.loading}>
          {state.loading ? '\u8bfb\u53d6\u4e2d...' : '\u91cd\u65b0\u8bfb\u53d6'}
        </button>
      </div>
      {state.error && <ErrorBox>{state.error}</ErrorBox>}
      {audit && <ResourceStatsContractAudit audit={audit} />}
      {!audit && state.data && <ResourceStatsDashboard stats={state.data} />}
      {!audit && !state.data && state.loading && <div className='fx-assets-empty'>{'\u6b63\u5728\u8bfb\u53d6 CMDB \u8d44\u6e90\u7edf\u8ba1\u5951\u7ea6\u3002'}</div>}
    </section>
  )
}

function ResourceApprovalSection({ section, meta }) {
  const view = approvalRouteViews[section] || 'mine'
  const [state, setState] = useState({ loading: true, data: null, error: '' })
  const loadApprovals = async () => {
    setState(prev => ({ ...prev, loading: true, error: '' }))
    try {
      const data = await cmdbApi.resourceApprovals(view)
      setState({ loading: false, data, error: '' })
    } catch (error) {
      setState({ loading: false, data: null, error: formatAssetError(error) })
    }
  }

  useEffect(() => { loadApprovals() }, [view])

  const audit = isCmdbContractAudit(state.data) ? state.data : null
  const list = state.data?.available && state.data?.mode === 'list' ? state.data.approval_requests || [] : []
  return (
    <section className='fx-assets-work fx-resource-stats'>
      <div className='fx-resource-stats-head'>
        <div>
          <h2>{meta.label}</h2>
          <p>{meta.desc}</p>
        </div>
        <button type='button' className='fx-assets-button is-primary' onClick={loadApprovals} disabled={state.loading}>
          {state.loading ? '\u8bfb\u53d6\u4e2d...' : '\u91cd\u65b0\u8bfb\u53d6'}
        </button>
      </div>
      {state.error && <ErrorBox>{state.error}</ErrorBox>}
      {audit && <ResourceStatsContractAudit audit={audit} title={'\u8d44\u6e90\u5ba1\u6279 runtime \u5951\u7ea6\u5ba1\u8ba1'} />}
      {!audit && state.data?.available && <ResourceApprovalList data={state.data} rows={list} />}
      {!state.data && state.loading && <div className='fx-assets-empty'>{'\u6b63\u5728\u8bfb\u53d6 CMDB \u8d44\u6e90\u5ba1\u6279 runtime \u5951\u7ea6\u3002'}</div>}
    </section>
  )
}

function ResourceApprovalList({ data, rows }) {
  return (
    <div className='fx-resource-approval-body'>
      <ResourceApprovalContractGaps data={data} />
      <div className='fx-resource-approval-list'>
        {rows.map(row => <ResourceApprovalCard key={row.id} row={row} />)}
      </div>
    </div>
  )
}

function ResourceApprovalCard({ row }) {
  const missingContracts = row.missing_contracts || []
  const fields = [
    ['resource', row.resource],
    ['action', row.action],
    ['risk', row.risk],
    ['status', formatApprovalStatus(row.status)],
    ['requester', row.requester],
    ['approver', row.approver],
    ['audit_ref', row.audit_ref],
    ['risk_record_id', row.risk_record_id],
    ['created_at', fmtTime(row.created_at) || row.created_at],
    ['updated_at', fmtTime(row.updated_at) || row.updated_at],
  ]
  return (
    <article className='fx-resource-approval-card'>
      <header>
        <div>
          <h3>{row.title}</h3>
          {row.summary && <p>{row.summary}</p>}
        </div>
        <Status ok={false}>{formatApprovalStatus(row.status)}</Status>
      </header>
      <dl>
        {fields.map(([label, value]) => (
          <React.Fragment key={label}>
            <dt>{label}</dt>
            <dd>{value || '-'}</dd>
          </React.Fragment>
        ))}
      </dl>
      {missingContracts.length > 0 && (
        <div className='fx-resource-approval-contracts'>
          <strong>missing contracts</strong>
          {missingContracts.map(item => <code key={item.id}>{item.status ? item.id + ' / ' + item.status : item.id}</code>)}
        </div>
      )}
    </article>
  )
}

function ResourceApprovalContractGaps({ data }) {
  const gaps = data.contract_gaps || {}
  const groups = [
    ['workflow', gaps.workflow || []],
    ['risk', gaps.risk || []],
    ['audit', gaps.audit || []],
  ]
  return (
    <article className='fx-resource-approval-audit'>
      <div className='fx-resource-approval-audit-head'>
        <div>
          <h3>{'\u5951\u7ea6\u7f3a\u53e3'}</h3>
          <p>{'\u5ba1\u6279\u5217\u8868\u6765\u81ea\u540e\u7aef approval_requests\uff1b\u6267\u884c\u6295\u9012\u548c\u653e\u884c\u56de\u6267\u4ecd\u6309\u5951\u7ea6\u7f3a\u53e3\u4fdd\u5b88\u5c55\u793a\u3002'}</p>
        </div>
        {data.raw_status && <Status ok={false}>{data.raw_status}</Status>}
      </div>
      <div className='fx-resource-approval-gap-grid'>
        {groups.map(([name, rows]) => (
          <div key={name}>
            <strong>{name}</strong>
            {rows.length ? rows.map(item => <code key={item.id}>{item.status ? item.id + ' / ' + item.status : item.id}</code>) : <span>{'\u672a\u8fd4\u56de\u7f3a\u53e3'}</span>}
          </div>
        ))}
      </div>
      {data.findx_audit_query && (
        <div className='fx-resource-stat-contract-list'>
          <strong>findx_audit query</strong>
          {Object.entries(data.findx_audit_query).map(([key, value]) => <code key={key}>{`${key}=${value}`}</code>)}
        </div>
      )}
    </article>
  )
}

function ResourceStatsDashboard({ stats }) {
  const summary = stats.summary || {}
  const dimensions = stats.dimensions || {}
  const cards = [
    ['CMDB instances', summary.cmdb_instances],
    ['resource models', summary.cmdb_models],
    ['host assets', summary.host_assets],
    ['FindX Agent', summary.findx_agents],
    ['monitor bindings', summary.monitor_binding_rows],
    ['bound instances', summary.monitor_bound_instances],
    ['relation edges', summary.relation_edges],
    ['resource groups', summary.resource_groups],
  ]
  return (
    <div className='fx-resource-stats-body'>
      <div className='fx-resource-stat-cards'>
        {cards.map(([label, value]) => (
          <article className='fx-resource-stat-card' key={label}>
            <strong>{formatStatNumber(value)}</strong>
            <span>{label}</span>
          </article>
        ))}
      </div>
      <div className='fx-resource-stat-panels'>
        <ResourceStatsList title='model distribution' rows={dimensions.by_model} />
        <ResourceStatsList title='resource group distribution' rows={dimensions.by_group} />
        <ResourceStatsList title='business distribution' rows={dimensions.by_workspace} />
        <ResourceStatsContractAudit audit={stats} compact />
      </div>
      {stats.updated_at && <p className='fx-assets-muted'>{'updated_at='}{fmtTime(stats.updated_at)}</p>}
    </div>
  )
}

function ResourceStatsList({ title, rows = [] }) {
  const max = Math.max(...rows.map(row => Number(row.value) || 0), 1)
  return (
    <article className='fx-resource-stat-panel'>
      <h3>{title}</h3>
      {rows.length ? rows.slice(0, 8).map(row => {
        const value = Number(row.value) || 0
        return (
          <div className='fx-resource-stat-row' key={row.id || row.label}>
            <div><strong>{row.label}</strong><span>{formatStatNumber(value)}</span></div>
            <i style={{ width: Math.max(4, Math.round((value / max) * 100)) + '%' }} />
          </div>
        )
      }) : <div className='fx-assets-empty'>no rows returned</div>}
    </article>
  )
}

function ResourceStatsContractAudit({ audit, compact = false, title }) {
  const blockedContracts = (audit.blocked_contracts || [])
    .map(item => typeof item === 'string' ? { id: item } : item)
    .filter(item => item && item.id)
    .filter((item, index, rows) => rows.findIndex(row => row.id === item.id) === index)
  const missingContracts = (audit.missing_contracts || [])
    .map(item => typeof item === 'string' ? { id: item } : item)
    .filter(item => item && item.id)
    .filter((item, index, rows) => rows.findIndex(row => row.id === item.id) === index)
  const evidence = audit.source_evidence || []
  return (
    <article className={`fx-resource-stat-contract ${compact ? 'is-compact' : ''}`}>
      <div className='fx-cmdb-blocked-head'>
        <div>
          <h2>{title || (compact ? 'contract audit' : 'resource statistics contract blocked')}</h2>
          <p>{audit.message || 'missing resource statistics contract; empty statistics are not rendered'}</p>
        </div>
        {isCmdbContractAudit(audit) && <Status ok={false}>PENDING</Status>}
      </div>
      {audit.contract_id && (
        <div className='fx-resource-stat-contract-list'>
          <strong>contract audit</strong>
          <code>{`contract_id=${audit.contract_id}`}</code>
          {audit.view && <code>{`view=${audit.view}`}</code>}
          {audit.status && <code>{`status=${audit.status}`}</code>}
        </div>
      )}
      {blockedContracts.length > 0 && (
        <div className='fx-resource-stat-contract-list'>
          <strong>blocked contracts</strong>
          {blockedContracts.map(item => <code key={item.id}>{item.status ? item.id + ' / ' + item.status : item.id}</code>)}
        </div>
      )}
      {missingContracts.length > 0 && (
        <div className='fx-resource-stat-contract-list'>
          <strong>missing contracts</strong>
          {missingContracts.map(item => <code key={item.id}>{item.status ? item.id + ' / ' + item.status : item.id}</code>)}
        </div>
      )}
      {audit.findx_audit_query && (
        <div className='fx-resource-stat-contract-list'>
          <strong>findx_audit query</strong>
          {Object.entries(audit.findx_audit_query).map(([key, value]) => <code key={key}>{`${key}=${value}`}</code>)}
        </div>
      )}
      {evidence.length > 0 && (
        <div className='fx-resource-stat-contract-list'>
          <strong>evidence</strong>
          {evidence.slice(0, 4).map(item => <span key={item}>{item}</span>)}
        </div>
      )}
    </article>
  )
}

function formatStatNumber(value) {
  return value === null || value === undefined ? '-' : Number(value).toLocaleString('zh-CN')
}

function formatApprovalStatus(status) {
  const value = String(status || '').trim()
  if (value === 'pending_review') return '\u5f85\u590d\u6838'
  if (value === 'review_accept_recorded') return '\u590d\u6838\u610f\u89c1\u5df2\u8bb0\u5f55'
  return value || '-'
}

function CmdbCapabilityBlockedSection({ section, meta }) {
  const contractBase = `cmdb_${section.replace(/-/g, '_')}`
  const missing = [
    `${contractBase}_store_contract`,
    `${contractBase}_handler_contract`,
    `${contractBase}_audit_receipt_contract`,
  ]
  return (
    <section className='fx-assets-work fx-cmdb-capability-blocked'>
      <div className='fx-cmdb-blocked-head'>
        <div>
          <h2>{meta.label}</h2>
          <p>{meta.desc}</p>
        </div>
        <Status ok={false}>PENDING</Status>
      </div>
      <Blocked>PENDING: backend contract is missing; only contract audit is rendered.</Blocked>
      <div className='fx-cmdb-blocked-grid'>
        <article>
          <strong>missing contracts</strong>
          {missing.map(item => <code key={item}>{item}</code>)}
        </article>
        <article>
          <strong>findx_audit query</strong>
          <code>{`source=findx_audit scope=cmdb resource_type=${section} action=cmdb.${section}.request`}</code>
        </article>
      </div>
    </section>
  )
}
