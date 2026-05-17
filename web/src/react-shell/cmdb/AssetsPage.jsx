import React, { useEffect, useState } from 'react'
import { assetsApi, formatAssetError } from '../api/assets.js'
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
import { ErrorBox, Status } from './Shared.jsx'
import './assets.css'

const CMDB_BUILD = '20260511b'
const approvalRouteViews = {
  'approval-mine': 'mine',
  'approval-todo': 'todo',
  'approval-archive': 'archive',
}
const routeSections = new Set([...sectionSet, 'search', 'topology', 'instance-detail', 'resource-stats', 'datacenter-view', ...Object.keys(approvalRouteViews)])
const routeMeta = {
  topology: { label: '\u0043\u004d\u0044\u0042 \u5173\u7cfb\u62d3\u6251', desc: '\u6309\u5b9e\u4f8b\u3001\u4e1a\u52a1\u4e0a\u4e0b\u6587\u548c\u5ba1\u8ba1\u5951\u7ea6\u5c55\u793a\u5173\u7cfb\u3002' },
  'instance-detail': { label: '\u0043\u004d\u0044\u0042 \u5b9e\u4f8b\u8be6\u60c5', desc: '\u5c55\u793a\u5b9e\u4f8b\u8be6\u60c5\u3001\u5173\u7cfb\u53cd\u67e5\u548c\u76d1\u63a7\u7ed1\u5b9a\u5951\u7ea6\u3002' },
  'datacenter-view': { label: '机房视图', desc: '按机房、机柜、U位展示物理资源布局和容量。' },
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
  const meta = routeMeta[section] || sections.find(item => item.value === section) || sections[0]
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

