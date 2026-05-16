import React, { useEffect, useMemo, useState } from 'react'
import { AGENT_BLOCKERS, agentApi, formatAgentError } from '../api/agents.js'
import { agentOnline, capabilityPackages, configTemplates, displayText, fmtTime, installCommands, rowText } from './agentModel.js'
import { Blocked, CopyBlock, Empty, ErrorBox, Field, Status, Tags } from './AgentShared.jsx'
import { AgentBlockedRecords } from './AgentBlockedRecords.jsx'
import { HostProviderRefs, REMOTE_MUTATION_REF_KEYS } from './HostProviderRefs.jsx'
import { HostDataArrivalEvidence } from './HostDataArrivalEvidence.jsx'
import { InstallCommandPreview } from './InstallCommandPreview.jsx'
import { InstallEnvironmentSummary, normalizeInstallEnvironment } from './PackagesSection.jsx'
import { Pagination } from '../shared/ConfirmModal.jsx'
const T = {
  metrics: '\u6307\u6807',
  heartbeat: '\u5fc3\u8df3',
  host: '\u4e3b\u673a',
  process: '\u8fdb\u7a0b',
  container: '\u5bb9\u5668',
  logs: '\u65e5\u5fd7',
  tracing: '\u94fe\u8def',
  profiling: '\u6027\u80fd\u5206\u6790',
  inspection: '\u5de1\u68c0',
  topology: '\u62d3\u6251',
  rum: '\u524d\u7aef\u4f53\u9a8c',
  config: '\u914d\u7f6e',
  ungrouped: '\u672a\u5206\u7ec4',
  template: '\u914d\u7f6e\u6a21\u677f',
  allAgents: '\u5168\u90e8 Agent',
  business: '\u4e1a\u52a1\u7ec4',
  abilityPackage: '\u80fd\u529b\u5305',
  saveTemplate: '\u4fdd\u5b58\u6a21\u677f',
  canary: '\u7070\u5ea6\u4e0b\u53d1',
  rollout: '\u5168\u91cf\u4e0b\u53d1',
  rollback: '\u56de\u6eda',
  installGuide: '\u5b89\u88c5\u5411\u5bfc',
  heartbeatEvidence: '\u5fc3\u8df3\u8bc1\u636e',
  dataEvidence: '\u6570\u636e\u5230\u8fbe\u8bc1\u636e',
  configDelivery: '\u914d\u7f6e\u4e0b\u53d1',
  pluginDelivery: '\u63d2\u4ef6\u4e0b\u53d1',
  cmdbHosts: 'CMDB \u4e3b\u673a',
  installedAgents: '\u5df2\u5b89\u88c5 Agent',
  missingAgents: '\u672a\u5b89\u88c5 Agent',
  onlineHeartbeat: '\u5728\u7ebf\u5fc3\u8df3',
  refresh: '\u5237\u65b0',
  installUpgrade: '\u5b89\u88c5 / \u5347\u7ea7',
  configPlugin: '\u914d\u7f6e / \u63d2\u4ef6\u4e0b\u53d1',
  selected: '\u5df2\u9009\u62e9',
  currentFilter: '\uff08\u5f53\u524d\u7b5b\u9009\u8303\u56f4\uff09',
  machine: '\u4e3b\u673a',
  heartbeatColumn: '\u5fc3\u8df3',
  dataArrival: '\u6570\u636e\u5230\u8fbe',
  configVersion: '\u914d\u7f6e\u7248\u672c',
  actions: '\u64cd\u4f5c',
  registered: '\u5df2\u767b\u8bb0',
  heartbeatOnly: '\u4ec5\u5fc3\u8df3',
  installed: '\u5df2\u5b89\u88c5',
  notInstalled: '\u672a\u5b89\u88c5',
  online: '\u5728\u7ebf',
  offline: '\u79bb\u7ebf',
  notReported: '\u672a\u4e0a\u62a5',
  reinstall: '\u91cd\u88c5/\u5347\u7ea7',
  install: '\u5b89\u88c5',
  plugin: '\u914d\u7f6e/\u63d2\u4ef6',
  heartbeatDetail: '\u5fc3\u8df3\u8be6\u60c5',
  validateData: '\u6570\u636e\u9a8c\u8bc1',
  evidence: '\u8bc1\u636e',
  latest: '\u6700\u8fd1',
  target: '\u4e0b\u53d1\u76ee\u6807',
  os: '\u7cfb\u7edf',
  installMethod: '\u5b89\u88c5\u65b9\u5f0f',
  createPlan: '\u751f\u6210\u5b89\u88c5\u8ba1\u5212',
  executeInstall: '\u6267\u884c\u5b89\u88c5',
  deliveryMode: '\u4e0b\u53d1\u6a21\u5f0f',
  deliveryStrategy: '\u4e0b\u53d1\u7b56\u7565',
  fields: '\u914d\u7f6e\u5b57\u6bb5',
  pluginMutation: '\u91c7\u96c6\u63d2\u4ef6\u8fdc\u7a0b\u4fee\u6539',
  provider: '\u63d0\u4f9b\u65b9',
  format: '\u914d\u7f6e\u683c\u5f0f',
  mutationBlocked: '\u8fdc\u7a0b\u4fee\u6539\u5951\u7ea6\u5f85\u5f00\u653e',
  close: '\u5173\u95ed',
  noTargets: '\u5f53\u524d\u7b5b\u9009\u8303\u56f4\u5185\u6ca1\u6709\u53ef\u4e0b\u53d1\u76ee\u6807\u3002',
  chooseTarget: '\u8bf7\u9009\u62e9\u81f3\u5c11\u4e00\u4e2a CMDB \u4e3b\u673a\u6216 Agent \u76ee\u6807\u3002',
}

const focusLabels = { install: T.installGuide, templates: T.template, heartbeat: T.heartbeatEvidence, 'data-arrival': T.dataEvidence, config: T.configDelivery, plugins: T.pluginDelivery }
const summaryId = item => String(item?.kind || item?.id || '').toLowerCase()
const summaryOk = item => item?.status === 'reported' && Number(item?.evidence_count || 0) > 0
const hostId = host => host?.host_id || host?.id || host?.target_id || ''
const hostIps = host => [...(Array.isArray(host?.ip_list) ? host.ip_list : []), host?.ip].filter(Boolean)
const hostName = row => displayText(row.host?.hostname || row.host?.name || row.host?.ident || row.agent?.hostname || row.agent?.ident)
const hostIp = row => displayText(hostIps(row.host)[0] || row.agent?.ip)
const agentIdentity = agent => agent?.id || agent?.ident || agent?.target_id || agent?.ip || agent?.hostname || ''
const rowKey = row => row?.id || hostId(row?.host) || agentIdentity(row?.agent) || `${hostName(row)}-${hostIp(row)}`
const rowTargetId = row => hostId(row?.host) || row?.agent?.target_id || ''
const rowAgentId = row => row?.agent?.id || ''
const selectedRowsFrom = (rows, keys) => rows.filter(row => keys.has(rowKey(row)))
const collectIds = rows => ({ targetIds: rows.map(rowTargetId).filter(Boolean), agentIds: rows.map(rowAgentId).filter(Boolean) })
const norm = value => String(value || '').trim().toLowerCase()
const rowEvidenceKeys = row => [rowTargetId(row), rowAgentId(row), row.host?.agent_id, row.host?.ident, row.agent?.ident, row.host?.hostname, row.agent?.hostname, row.host?.ip, row.agent?.ip, ...hostIps(row.host)].map(norm).filter(Boolean)
const cleanName = value => displayText(value).replace(/[\u4e00-\u9fa5]*[?][\u4e00-\u9fa5]*/g, '').trim() || displayText(value)
const hostText = row => rowText({ ...(row.host || {}), ...(row.agent || {}), installStatus: row.installed ? 'installed' : 'not-installed' })

const normalizePackage = item => ({ ...item, name: item.name || item.id, capabilityDomain: item.capabilityDomain || item.capability_domain || T.ungrouped, os: item.supported_os || item.os || [], installEnvironment: normalizeInstallEnvironment(item.install_environment || item.installEnvironment) })
const normalizePluginConfig = value => value ? {
  pluginId: value.pluginId || value.plugin_id || '',
  pluginVersion: value.pluginVersion || value.plugin_version || '<PLUGIN_VERSION>',
  configFormat: value.configFormat || value.config_format || 'toml',
  configSnippetRef: value.configSnippetRef || value.config_snippet_ref || '<CONFIG_SNIPPET_REF>',
  providerModes: value.providerModes || value.provider_modes || ['local', 'http'],
  reloadStrategy: value.reloadStrategy || value.reload_strategy || '',
  restartStrategy: value.restartStrategy || value.restart_strategy || '',
  remoteMutation: value.remoteMutation ?? value.remote_mutation ?? false,
  remoteMutationStatus: value.remoteMutationStatus || value.remote_mutation_status || 'PENDING',
  rolloutMetadata: value.rolloutMetadata || value.rollout_metadata || [],
} : null
const templateFallbacks = new Map(configTemplates.map(item => [item.id, item]))
const safeMetadataRef = key => `<${String(key || '').replace(/[^a-z0-9]+/gi, '_').toUpperCase()}>`
const buildSafeMetadataRefs = rolloutMetadata => Object.fromEntries([...new Set([...REMOTE_MUTATION_REF_KEYS, ...(Array.isArray(rolloutMetadata) ? rolloutMetadata : [])].filter(Boolean))].map(key => [key, safeMetadataRef(key)]))
const normalizeTemplate = item => {
  const fallback = templateFallbacks.get(item.id) || {}
  return {
    ...fallback, ...item,
    name: cleanName(item.name || fallback.name || item.id),
    targetModes: item.targetModes || item.target_scopes || fallback.targetModes || [T.allAgents, T.business, T.machine, T.abilityPackage],
    rolloutStrategies: item.rolloutStrategies || item.rollout_strategies || fallback.rolloutStrategies || [T.saveTemplate, T.canary, T.rollout, T.rollback],
    capabilityPackages: item.capabilityPackages || item.capability_packages || fallback.capabilityPackages || [T.abilityPackage],
    fields: item.fields || fallback.fields || [],
    pluginConfig: normalizePluginConfig(item.pluginConfig || item.plugin_config || fallback.pluginConfig),
  }
}

const matchAgent = (host, agents, used) => {
  const id = hostId(host)
  const ips = hostIps(host)
  const hit = agents.find(agent => {
    const key = agentIdentity(agent)
    if (key && used.has(key)) return false
    return (id && agent.target_id === id) || (host?.agent_id && agent.id === host.agent_id) || (host?.ident && agent.ident === host.ident) || (host?.hostname && agent.hostname === host.hostname) || (ips.length && ips.includes(agent.ip))
  })
  const key = agentIdentity(hit)
  if (key) used.add(key)
  return hit
}

const mergeHostAgents = (hosts, agents) => {
  const used = new Set()
  const merged = hosts.map(host => {
    const agent = matchAgent(host, agents, used)
    return { id: hostId(host) || host.ident || host.hostname || host.ip, host, agent, installed: Boolean(agent?.id || host?.agent_id) }
  })
  agents.forEach(agent => {
    const key = agentIdentity(agent)
    if (!key || !used.has(key)) merged.push({ id: key || `${agent.ip}-${agent.hostname}`, host: null, agent, installed: true })
  })
  return merged
}

function Feedback({ children }) {
  if (!children) return null
  return String(children).startsWith('PENDING') ? <Blocked>{children}</Blocked> : <ErrorBox>{children}</ErrorBox>
}

export function HostsSection({ rows, hosts = [], packages, dataArrival = [], dataArrivalEvidence = [], focus, error, q, onRefresh }) {
  const [drawer, setDrawer] = useState(null)
  const [blocked, setBlocked] = useState('')
  const [selectedKeys, setSelectedKeys] = useState(new Set())
  const [templates, setTemplates] = useState(configTemplates.map(normalizeTemplate))
  const [page, setPage] = useState(1)
  const PAGE_SIZE = 20
  const packageRows = useMemo(() => (packages?.length ? packages : capabilityPackages).map(normalizePackage), [packages])
  const merged = useMemo(() => mergeHostAgents(hosts || [], rows || []), [hosts, rows])
  const filtered = useMemo(() => {
    const keyword = String(q || '').trim().toLowerCase()
    setPage(1)
    return keyword ? merged.filter(row => hostText(row).includes(keyword)) : merged
  }, [merged, q])
  const paged = useMemo(() => filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE), [filtered, page])
  const selectedRows = useMemo(() => selectedRowsFrom(filtered, selectedKeys), [filtered, selectedKeys])
  const reportedArrival = useMemo(() => dataArrival.filter(summaryOk), [dataArrival])
  const allVisibleSelected = filtered.length > 0 && filtered.every(row => selectedKeys.has(rowKey(row)))

  useEffect(() => {
    let alive = true
    agentApi.templates().then(value => { if (alive && value.length) setTemplates(value.map(normalizeTemplate)) }).catch(() => {})
    return () => { alive = false }
  }, [])
  useEffect(() => {
    const valid = new Set(filtered.map(rowKey))
    setSelectedKeys(prev => new Set([...prev].filter(key => valid.has(key))))
  }, [filtered])

  const installed = merged.filter(row => row.installed).length
  const online = merged.filter(row => row.agent && agentOnline(row.agent)).length
  const toggleRow = (row, checked) => setSelectedKeys(prev => { const next = new Set(prev); checked ? next.add(rowKey(row)) : next.delete(rowKey(row)); return next })
  const openDrawer = type => setDrawer({ type, targets: selectedRows.length ? selectedRows : filtered })

  return <section className='fx-agent-work'>
    <div className='fx-agent-grid'><Card value={hosts.length} label={T.cmdbHosts} /><Card value={installed} label={T.installedAgents} /><Card value={merged.length - installed} label={T.missingAgents} /><Card value={`${online}/${installed}`} label={T.onlineHeartbeat} /></div>
    {focusLabels[focus] && <div className='fx-agent-focus'><strong>{focusLabels[focus]}</strong><span>{'\u8be5\u5165\u53e3\u5df2\u5f52\u5e76\u5230\u4e3b\u673a Agent\uff0c\u672c\u9875\u5c55\u793a\u540c\u4e00\u6279\u4e3b\u673a\u7684\u5b89\u88c5\u3001\u5fc3\u8df3\u548c\u6570\u636e\u5230\u8fbe\u8bc1\u636e\u3002'}</span></div>}
    <DataArrivalSummary rows={reportedArrival} />
    <div className='fx-agent-toolbar'><button type='button' onClick={onRefresh}>{T.refresh}</button><button type='button' onClick={() => openDrawer('install')}>{T.installUpgrade}</button><button type='button' onClick={() => openDrawer('config')}>{T.configPlugin}</button><span className='fx-agent-selected'>{T.selected} {selectedRows.length || filtered.length} {'\u53f0\u76ee\u6807'}{selectedRows.length ? '' : T.currentFilter}</span></div>
    <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}<Blocked>{AGENT_BLOCKERS.installLifecycle}</Blocked><Blocked>{AGENT_BLOCKERS.configLifecycle}</Blocked>
    <AgentBlockedRecords {...collectIds(selectedRows.length ? selectedRows : filtered)} targetCount={selectedRows.length || filtered.length} />
    <HostTable rows={paged} selectedKeys={selectedKeys} allSelected={allVisibleSelected} dataArrival={dataArrival} dataArrivalEvidence={dataArrivalEvidence} onToggle={toggleRow} onToggleAll={checked => setSelectedKeys(checked ? new Set(filtered.map(rowKey)) : new Set())} onDrawer={setDrawer} onBlocked={setBlocked} />
    <Pagination total={filtered.length} page={page} pageSize={PAGE_SIZE} onPageChange={setPage} />
    {drawer?.type === 'install' && <InstallDrawer targets={drawer.targets || []} packages={packageRows} onClose={() => setDrawer(null)} />}
    {drawer?.type === 'config' && <ConfigDrawer targets={drawer.targets || []} templates={templates} onClose={() => setDrawer(null)} />}
  </section>
}

function Card({ value, label }) {
  return <article className='fx-agent-card'><strong>{value}</strong><span>{label}</span></article>
}

function DataArrivalSummary({ rows }) {
  return <div className='fx-agent-summary-row'>{rows.length ? <><Tags items={rows.map(item => `${displayText(item.name || item.kind)}\u5df2\u4e0a\u62a5`)} />{rows.map(item => <span key={summaryId(item)}>{displayText(item.name || item.kind)} {T.evidence} {Number(item.evidence_count || 0)}{'\uff0c'}{T.latest} {fmtTime(item.last_seen)}</span>)}</> : <Blocked>{AGENT_BLOCKERS.dataArrival}</Blocked>}</div>
}

function HostTable({ rows, selectedKeys, allSelected, dataArrivalEvidence, onToggle, onToggleAll, onDrawer, onBlocked }) {
  return <div className='fx-agent-table'><table><thead><tr><th className='fx-agent-select-cell'><input type='checkbox' checked={allSelected} onChange={event => onToggleAll(event.target.checked)} aria-label='\u9009\u62e9\u5f53\u524d\u7b5b\u9009\u8303\u56f4\u5185\u5168\u90e8\u4e3b\u673a' /></th><th>{T.machine}</th><th>CMDB</th><th>FindX Agent</th><th>{T.heartbeatColumn}</th><th>{T.dataArrival}</th><th>{T.configVersion}</th><th>{T.actions}</th></tr></thead><tbody>{rows.map(row => <HostRow key={rowKey(row)} row={row} checked={selectedKeys.has(rowKey(row))} dataArrivalEvidence={dataArrivalEvidence} onToggle={onToggle} onDrawer={onDrawer} onBlocked={onBlocked} />)}</tbody></table>{!rows.length && <Empty>{'\u6682\u65e0 CMDB \u4e3b\u673a\u6216 FindX Agent \u5fc3\u8df3\u8bb0\u5f55\u3002CMDB \u65b0\u589e\u4e3b\u673a\u540e\u4f1a\u81ea\u52a8\u51fa\u73b0\u5728\u8fd9\u91cc\uff0c\u672a\u5b89\u88c5\u65f6\u663e\u793a\u672a\u5b89\u88c5\u3002'}</Empty>}</div>
}

function HostRow({ row, checked, dataArrivalEvidence, onToggle, onDrawer, onBlocked }) {
  const agent = row.agent
  return <tr><td className='fx-agent-select-cell'><input type='checkbox' checked={checked} onChange={event => onToggle(row, event.target.checked)} aria-label={`select ${hostName(row)} ${hostIp(row)}`} /></td><td><strong>{hostName(row)}</strong><div className='fx-agent-muted'>{hostIp(row)} {displayText(row.host?.os || agent?.os, '')} {displayText(row.host?.arch || agent?.arch, '')}</div></td><td><Status ok={Boolean(row.host)}>{row.host ? T.registered : T.heartbeatOnly}</Status><div className='fx-agent-muted'>{displayText(row.host?.resource_group_id || row.host?.workspace_id, '')}</div></td><td><Status ok={row.installed}>{row.installed ? T.installed : T.notInstalled}</Status><div className='fx-agent-muted'>{displayText(agent?.version || row.host?.agent_version, '')}</div></td><td><Status ok={agentOnline(agent)}>{agentOnline(agent) ? T.online : row.installed ? T.offline : T.notReported}</Status><div className='fx-agent-muted'>{fmtTime(agent?.last_seen || row.host?.last_seen_at)}</div></td><td><HostDataArrivalEvidence evidenceRows={dataArrivalEvidence} identityKeys={rowEvidenceKeys(row)} installed={row.installed} capabilities={agent?.capabilities} /></td><td>{displayText(agent?.config_version)}</td><td className='fx-agent-actions'><button type='button' onClick={() => onDrawer({ type: 'install', targets: [row] })}>{row.installed ? T.reinstall : T.install}</button><button type='button' onClick={() => onDrawer({ type: 'config', targets: [row] })}>{T.plugin}</button><button type='button' onClick={() => onBlocked(AGENT_BLOCKERS.heartbeat)}>{T.heartbeatDetail}</button><button type='button' onClick={() => onBlocked(AGENT_BLOCKERS.dataArrival)}>{T.validateData}</button></td></tr>
}

function Drawer({ title, targets, onClose, children }) {
  return <div className='fx-agent-drawer-mask' role='presentation'><aside className='fx-agent-drawer' role='dialog' aria-label={title}><header><div><h3>{title}</h3><p>{targets.length ? `\u5df2\u52a0\u8f7d ${targets.length} \u4e2a CMDB / Agent \u76ee\u6807` : T.noTargets}</p></div><button type='button' onClick={onClose}>{T.close}</button></header>{children}</aside></div>
}

function TargetPicker({ targets, selectedKeys, onToggle }) {
  if (!targets.length) return <Empty>{T.noTargets}</Empty>
  return <div className='fx-agent-target-list'>{targets.map(row => <label key={rowKey(row)} className='fx-agent-target-item'><input type='checkbox' checked={selectedKeys.has(rowKey(row))} onChange={event => onToggle(row, event.target.checked)} /><span><strong>{hostName(row)}</strong><em>{hostIp(row)} / {row.host ? T.registered : T.heartbeatOnly} / {row.installed ? T.installed : T.notInstalled}</em></span></label>)}</div>
}

function useTargetSelection(targets) {
  const [selectedTargetKeys, setSelectedTargetKeys] = useState(() => new Set(targets.map(rowKey)))
  useEffect(() => { setSelectedTargetKeys(new Set(targets.map(rowKey))) }, [targets])
  const selectedTargets = useMemo(() => selectedRowsFrom(targets, selectedTargetKeys), [targets, selectedTargetKeys])
  const toggleTarget = (row, checked) => setSelectedTargetKeys(prev => { const next = new Set(prev); checked ? next.add(rowKey(row)) : next.delete(rowKey(row)); return next })
  return { selectedTargetKeys, selectedTargets, toggleTarget }
}

function InstallDrawer({ targets, packages, onClose }) {
  const { selectedTargetKeys, selectedTargets, toggleTarget } = useTargetSelection(targets)
  const [packageId, setPackageId] = useState(packages[0]?.id || 'agent-core')
  const [method, setMethod] = useState('linux-curl')
  const [os, setOs] = useState(targets[0]?.host?.os || targets[0]?.agent?.os || 'Linux')
  const [feedback, setFeedback] = useState('')
  const selectedPackage = packages.find(item => item.id === packageId) || packages[0]
  const commandLabel = { 'linux-curl': 'Linux curl', 'windows-cmd': 'Windows CMD certutil', 'windows-powershell': 'PowerShell Invoke-WebRequest', helm: 'Kubernetes Helm', 'kubernetes-daemonset': 'Kubernetes DaemonSet', 'kubernetes-operator': 'Kubernetes Operator', 'kubernetes-sidecar': 'Kubernetes Sidecar', 'kubernetes-initcontainer': 'Kubernetes InitContainer' }
  const requestPlan = () => {
    if (!selectedTargets.length) return setFeedback(T.chooseTarget)
    const { targetIds, agentIds } = collectIds(selectedTargets)
    agentApi.createInstallPlan({ package_id: packageId, method, os, target_id: targetIds[0] || '', target_ids: targetIds, agent_ids: agentIds, credential_ref: '<CREDENTIAL_REF>' }).then(() => setFeedback(AGENT_BLOCKERS.installLifecycle)).catch(err => setFeedback(formatAgentError(err)))
  }
  return <Drawer title={`${T.installUpgrade} FindX Agent`} targets={targets} onClose={onClose}><div className='fx-agent-drawer-body'><Field label={T.target}><TargetPicker targets={targets} selectedKeys={selectedTargetKeys} onToggle={toggleTarget} /></Field><Field label={T.abilityPackage}><select value={packageId} onChange={event => setPackageId(event.target.value)}>{packages.map(item => <option key={item.id} value={item.id}>{cleanName(item.name)}</option>)}</select></Field><Field label={T.os}><select value={os} onChange={event => setOs(event.target.value)}><option>Linux</option><option>Windows</option><option>Kubernetes</option></select></Field><Field label={T.installMethod}><select value={method} onChange={event => setMethod(event.target.value)}>{installCommands.map(item => <option key={item.id} value={item.id}>{commandLabel[item.id] || item.id}</option>)}</select></Field><InstallEnvironmentSummary environment={selectedPackage?.installEnvironment} /><InstallCommandPreview packageId={selectedPackage?.id || packageId} selectedMethod={method} onSelectMethod={setMethod} /><div className='fx-agent-actions'><button type='button' onClick={requestPlan}>{T.createPlan}</button><button type='button' onClick={() => setFeedback(AGENT_BLOCKERS.installLifecycle)}>{T.executeInstall}</button></div><Feedback>{feedback}</Feedback><Blocked>{AGENT_BLOCKERS.installLifecycle}</Blocked></div></Drawer>
}

function ConfigDrawer({ targets, templates, onClose }) {
  const { selectedTargetKeys, selectedTargets, toggleTarget } = useTargetSelection(targets)
  const [templateId, setTemplateId] = useState(templates[0]?.id || 'agent-core')
  const selected = templates.find(item => item.id === templateId) || templates[0]
  const [targetMode, setTargetMode] = useState(selected?.targetModes?.[0] || T.machine)
  const [strategy, setStrategy] = useState(selected?.rolloutStrategies?.[1] || T.canary)
  const [feedback, setFeedback] = useState('')
  const chooseTemplate = value => { const next = templates.find(item => item.id === value) || templates[0]; setTemplateId(value); setTargetMode(next?.targetModes?.[0] || T.machine); setStrategy(next?.rolloutStrategies?.[1] || next?.rolloutStrategies?.[0] || T.canary) }
  const { targetIds, agentIds } = collectIds(selectedTargets)
  const rollout = rolloutStrategy => {
    if (!selectedTargets.length) return setFeedback(T.chooseTarget)
    agentApi.rolloutConfig(buildRolloutBody(selected, rolloutStrategy, targetIds, agentIds, targetMode)).then(() => setFeedback(AGENT_BLOCKERS.configLifecycle)).catch(err => setFeedback(formatAgentError(err)))
  }
  const payload = buildRolloutBody(selected, strategy, targetIds.length ? targetIds : ['<CMDB_TARGET_ID>'], agentIds.length ? agentIds : ['<AGENT_ID>'], targetMode)
  return <Drawer title={T.configPlugin} targets={targets} onClose={onClose}><div className='fx-agent-drawer-body'><Field label={T.target}><TargetPicker targets={targets} selectedKeys={selectedTargetKeys} onToggle={toggleTarget} /></Field><Field label={T.template}><select value={templateId} onChange={event => chooseTemplate(event.target.value)}>{templates.map(item => <option key={item.id} value={item.id}>{cleanName(item.name)}</option>)}</select></Field><Field label={T.deliveryMode}><select value={targetMode} onChange={event => setTargetMode(event.target.value)}>{(selected.targetModes || []).map(item => <option key={item} value={item}>{item}</option>)}</select></Field><Field label={T.deliveryStrategy}><select value={strategy} onChange={event => setStrategy(event.target.value)}>{(selected.rolloutStrategies || []).map(item => <option key={item} value={item}>{item}</option>)}</select></Field><div className='fx-agent-subtitle'>{T.abilityPackage}</div><Tags items={(selected.capabilityPackages || []).map(cleanName)} /><div className='fx-agent-subtitle'>{T.fields}</div><Tags items={selected.fields || []} />{selected.pluginConfig && <PluginConfigTags pluginConfig={selected.pluginConfig} />}<HostProviderRefs rolloutMetadata={selected.pluginConfig?.rolloutMetadata} /><CopyBlock>{JSON.stringify(payload, null, 2)}</CopyBlock><div className='fx-agent-actions'><button type='button' onClick={() => rollout(T.saveTemplate)}>{T.saveTemplate}</button><button type='button' onClick={() => rollout(T.canary)}>{T.canary}</button><button type='button' onClick={() => rollout(T.rollout)}>{T.rollout}</button><button type='button' onClick={() => rollout(T.rollback)}>{T.rollback}</button></div><Feedback>{feedback}</Feedback><Blocked>{AGENT_BLOCKERS.configLifecycle}</Blocked></div></Drawer>
}

function PluginConfigTags({ pluginConfig }) { return <><div className='fx-agent-subtitle'>{T.pluginMutation}</div><Tags items={[`${T.format} ${pluginConfig.configFormat}`, `${T.provider} ${(pluginConfig.providerModes || []).join('/')}`, pluginConfig.remoteMutationStatus === 'PENDING' ? T.mutationBlocked : T.pluginMutation]} /></> }

function buildRolloutBody(template, strategy, targetIds, agentIds, targetMode = undefined) {
  return { template_id: template.id, target_mode: targetMode, rollout_strategy: strategy, credential_ref: '<CREDENTIAL_REF>', config_version: '<CONFIG_ID>', config_snippet_ref: template.pluginConfig?.configSnippetRef || '<CONFIG_SNIPPET_REF>', config_format: template.pluginConfig?.configFormat || 'toml', provider_mode: template.pluginConfig?.providerModes?.[1] || template.pluginConfig?.providerModes?.[0] || 'local', plugin_id: template.pluginConfig?.pluginId || '<PLUGIN_ID>', plugin_version: template.pluginConfig?.pluginVersion || '<PLUGIN_VERSION>', reload_strategy: template.pluginConfig?.reloadStrategy || '<RELOAD_STRATEGY>', restart_strategy: template.pluginConfig?.restartStrategy || '<RESTART_STRATEGY>', remote_mutation: Boolean(template.pluginConfig?.remoteMutation), canary_percent: strategy === T.canary ? 10 : 100, rollback_ref: '<ROLLBACK_REF>', audit_reason: '<AUDIT_REASON>', change_ticket: '<CHANGE_TICKET>', cluster_ref: '<CLUSTER_REF>', namespace_ref: '<NAMESPACE_REF>', workload_ref: '<WORKLOAD_REF>', config_map_ref: '<CONFIG_MAP_REF>', executor_ref: '<EXECUTOR_REF>', metadata: buildSafeMetadataRefs(template.pluginConfig?.rolloutMetadata), agent_ids: agentIds, target_ids: targetIds }
}
