import React, { useMemo, useState } from 'react'
import { AGENT_BLOCKERS, agentApi, formatAgentError } from '../api/agents.js'
import { capabilityPackages } from './agentModel.js'
import { Blocked, ErrorBox, Field, Status, Tags } from './AgentShared.jsx'
import { PluginDeliveryMatrix, ProbeEnvironmentMatrix } from './ProbeEnvironmentMatrix.jsx'

const BLOCKED = 'BLOCKED_BY_CONTRACT'
const blockedEnvironment = ['install_environment 契约未返回，不能判定平台自带安装能力。']

const sourceStateText = state => ({
  LOCAL_SOURCE_MISSING: '源码缺失',
  LOCAL_SOURCE_PRESENT: '源码线索已归档',
  PACKAGE_MISSING: '包缺失',
  READY: '可见但需验证',
  BLOCKED: '接入阻断',
})[state] || state || '未知'

const toList = value => Array.isArray(value) ? value.filter(Boolean) : String(value || '').split(/[,，\n]/).map(item => item.trim()).filter(Boolean)
const toLowerState = value => String(value || 'blocked').toLowerCase()
const methodCatalog = {
  'linux-curl': ['Linux', 'curl -kfsSL', '命令预览可见；真实本机/SSH 执行和 systemd 回执仍缺契约。'],
  'windows-cmd': ['Windows CMD', 'certutil -urlcache -f', '命令预览可见；真实 WinRM、Windows Service 和卸载验证仍缺契约。'],
  'windows-powershell': ['PowerShell', 'Invoke-WebRequest', '命令预览可见；真实 PowerShell 执行和服务回执仍缺契约。'],
  ssh: ['Linux SSH', '远程 SSH 下发', '目标选择和凭据 ref 可表达；真实 SSH 执行器、回执和幂等验证缺契约。'],
  winrm: ['Windows WinRM', '远程 WinRM 下发', '目标选择和凭据 ref 可表达；真实 WinRM 执行器、回执和幂等验证缺契约。'],
  helm: ['Kubernetes', 'Helm / Operator / DaemonSet / Sidecar / InitContainer', '预览入口可见；真实集群 apply、RBAC、回滚和数据到达缺契约。'],
}

const normalizeInstallTool = item => ({
  name: item?.name || 'unknown-tool',
  os: item?.os || '',
  arch: item?.arch || '',
  required: item?.required !== false,
  bundled: Boolean(item?.bundled),
  evidenceRef: item?.evidence_ref || item?.evidenceRef || '',
  status: toLowerState(item?.status || (item?.bundled ? 'bundled' : 'blocked')),
  blocker: item?.blocker || '',
})

export const normalizeInstallEnvironment = value => {
  if (!value || typeof value !== 'object') return { status: 'blocked', platforms: [], tools: [], blockers: blockedEnvironment }
  const tools = toList(value.tools).map(normalizeInstallTool)
  const blockers = toList(value.blockers)
  if (value.test_only || value.testOnly) blockers.push('PACKAGE_REPOSITORY_TEST_ONLY')
  const hasBlockedTool = tools.some(tool => tool.status === 'blocked' || !tool.bundled)
  const status = toLowerState(value.status || value.state || (blockers.length || hasBlockedTool ? 'blocked' : 'preview_only'))
  return { status, platforms: toList(value.platforms), tools, blockers }
}

const normalizePluginConfig = spec => spec ? {
  pluginId: spec.plugin_id || spec.pluginId,
  configFormat: (spec.config_format || spec.configFormat || '').toUpperCase(),
  configSnippetRef: spec.config_snippet_ref || spec.configSnippetRef,
  providerModes: spec.provider_modes || spec.providerModes || [],
  remoteMutationStatus: spec.remote_mutation_status || spec.remoteMutationStatus || BLOCKED,
  remoteReloadStatus: spec.remote_reload_status || spec.remoteReloadStatus || BLOCKED,
  driftDetectionStatus: spec.drift_detection_status || spec.driftDetectionStatus || BLOCKED,
  rollbackStatus: spec.rollback_status || spec.rollbackStatus || BLOCKED,
  receiptStatus: spec.receipt_status || spec.receiptStatus || BLOCKED,
  deliveryScope: spec.delivery_scope || spec.deliveryScope || [],
} : null

const normalizePackage = item => ({
  ...item,
  capabilityDomain: item.capabilityDomain || item.capability_domain || '未分组',
  os: item.supported_os || item.os || [],
  sourceState: item.source_state || item.sourceState,
  packageState: item.package_state || item.packageState || (item.status === 'ready' ? 'READY' : 'PACKAGE_MISSING'),
  packageShape: item.package_shape || item.packageShape,
  telemetryKinds: item.telemetry_kinds || item.telemetryKinds || [],
  configKeys: item.config_keys || item.configKeys || [],
  configTemplateIds: item.config_template_ids || item.configTemplateIds || [],
  installMethods: item.install_methods || item.installMethods || [],
  environmentMatrix: item.environment_matrix || item.environmentMatrix || [],
  pluginConfig: normalizePluginConfig(item.plugin_config || item.pluginConfig),
  installEnvironment: normalizeInstallEnvironment(item.install_environment || item.installEnvironment),
})

const summarizeToolEvidence = environment => {
  const bundled = environment.tools.filter(tool => tool.bundled)
  if (!bundled.length) return '自带工具证据未返回'
  return bundled.map(tool => `${tool.name}${tool.os ? ` ${tool.os}${tool.arch ? `/${tool.arch}` : ''}` : ''}${tool.evidenceRef ? ` evidence ${tool.evidenceRef}` : ''}`).join(' / ')
}

const deriveEnvironmentMatrix = item => {
  if (item.environmentMatrix.length) return item.environmentMatrix
  const methods = item.installMethods.length ? item.installMethods : ['linux-curl', 'windows-cmd', 'windows-powershell', 'ssh', 'winrm', 'helm']
  const blocker = [item.blocker, ...(item.blockers || []), ...(item.installEnvironment.blockers || [])].filter(Boolean).join(' / ') || AGENT_BLOCKERS.installLifecycle
  const toolEvidence = summarizeToolEvidence(item.installEnvironment)
  return methods.map(method => {
    const [platform, label, note] = methodCatalog[method] || ['未知平台', method, '安装方式契约未细化。']
    return {
      platform,
      method: label,
      toolEvidence: `${toolEvidence}；${note}`,
      sourceState: item.sourceState || 'LOCAL_SOURCE_MISSING',
      packageState: item.packageState || 'PACKAGE_MISSING',
      executor: BLOCKED,
      serviceRegistration: BLOCKED,
      configDelivery: item.pluginConfig ? '插件配置可进入下发控制面；真实远程修改仍阻断' : BLOCKED,
      uninstall: BLOCKED,
      rollback: BLOCKED,
      dataArrival: BLOCKED,
      blocker,
    }
  })
}

export function PackagesSection({ packages }) {
  const [blocked, setBlocked] = useState('')
  const [domain, setDomain] = useState('')
  const [runtime, setRuntime] = useState('')
  const [os, setOs] = useState('')
  const sourceRows = useMemo(() => (
    packages?.length ? packages.map(normalizePackage) : capabilityPackages.map(normalizePackage)
  ), [packages])
  const rows = useMemo(() => sourceRows.filter(item =>
    (!domain || item.capabilityDomain === domain) &&
    (!runtime || item.runtime === runtime) &&
    (!os || item.os.includes(os))
  ), [sourceRows, domain, runtime, os])
  const domains = useMemo(() => [...new Set(sourceRows.map(item => item.capabilityDomain))], [sourceRows])
  const runtimes = useMemo(() => [...new Set(sourceRows.map(item => item.runtime))], [sourceRows])
  const systems = useMemo(() => [...new Set(sourceRows.flatMap(item => item.os))], [sourceRows])

  const runPackageAction = (action, packageId = 'all') => {
    setBlocked('请求已提交到 blocked 审计控制面，等待后端返回契约状态。')
    agentApi.packageAction(action, packageId)
      .then(() => setBlocked(AGENT_BLOCKERS.packageLifecycle))
      .catch(err => setBlocked(formatAgentError(err)))
  }

  return (
    <section className='fx-agent-work'>
      <PackageHeader onSync={() => runPackageAction('sync_package_repository')} />
      <PackageFilters domains={domains} runtimes={runtimes} systems={systems} values={{ domain, runtime, os }} setters={{ setDomain, setRuntime, setOs }} />
      {blocked && <Blocked>{blocked}</Blocked>}
      <Blocked>{AGENT_BLOCKERS.packageLifecycle}</Blocked>
      <div className='fx-agent-domain-strip'>
        {domains.map(item => <span key={item}>{item}<strong>{sourceRows.filter(row => row.capabilityDomain === item).length}</strong></span>)}
      </div>
      <div className='fx-agent-package-grid'>
        {rows.map(item => <PackageCard key={item.id} item={item} onPackageAction={runPackageAction} />)}
      </div>
    </section>
  )
}

function PackageHeader({ onSync }) {
  return (
    <div className='fx-agent-banner'>
      <div>
        <h3>FindX Agent 能力包矩阵</h3>
        <p>平台展示自带探针安装环境、内置工具证据和环境缺口；命令预览、blocked 审计和下载入口不等于真实安装完成。</p>
      </div>
      <button type='button' onClick={onSync}>同步包仓库</button>
    </div>
  )
}

function PackageFilters({ domains, runtimes, systems, values, setters }) {
  return (
    <div className='fx-agent-filter fx-agent-filter-3'>
      <Field label='能力域'><SelectAll value={values.domain} items={domains} onChange={setters.setDomain} /></Field>
      <Field label='运行时'><SelectAll value={values.runtime} items={runtimes} onChange={setters.setRuntime} /></Field>
      <Field label='系统'><SelectAll value={values.os} items={systems} onChange={setters.setOs} /></Field>
    </div>
  )
}

function SelectAll({ value, items, onChange }) {
  return (
    <select value={value} onChange={event => onChange(event.target.value)}>
      <option value=''>全部</option>
      {items.map(item => <option key={item} value={item}>{item}</option>)}
    </select>
  )
}

function PackageCard({ item, onPackageAction }) {
  return (
    <article className='fx-agent-package'>
      <header>
        <div><span className='fx-agent-domain'>{item.capabilityDomain}</span><h3>{item.name}</h3></div>
        <Status ok={false}>{sourceStateText(item.sourceState)}</Status>
      </header>
      <p>{item.packageShape}</p>
      <Tags items={item.os} />
      <InstallEnvironmentSummary environment={item.installEnvironment} />
      <PackageState item={item} />
      <ProbeEnvironmentMatrix matrix={deriveEnvironmentMatrix(item)} />
      <PackageConfig item={item} />
      {item.pluginConfig && <PackagePluginConfig pluginConfig={item.pluginConfig} />}
      {item.pluginConfig && <PluginDeliveryMatrix pluginConfig={item.pluginConfig} />}
      {!!item.blockers?.length && <Blocked>{item.blockers.join(' / ')}</Blocked>}
      <div className='fx-agent-actions'>
        <button type='button' onClick={() => onPackageAction('publish_package', item.id)}>发布审计</button>
        <button type='button' onClick={() => onPackageAction('download_package', item.id)}>下载入口审计</button>
        <button type='button' onClick={() => onPackageAction('verify_package_signature', item.id)}>签名校验审计</button>
      </div>
    </article>
  )
}

export function InstallEnvironmentSummary({ environment }) {
  const missingTools = environment.tools.filter(tool => tool.status === 'blocked' || !tool.bundled)
  const blockers = [...environment.blockers, ...missingTools.map(tool => tool.blocker || `${tool.name}_BUNDLED_TOOL_EVIDENCE_MISSING`)].filter(Boolean)
  const bundledTools = environment.tools.filter(tool => tool.bundled).map(tool => `${tool.name}${tool.evidenceRef ? ` evidence ${tool.evidenceRef}` : ''}`)
  return (
    <div className='fx-agent-plugin-config'>
      <div className='fx-agent-subtitle'>安装环境摘要</div>
      <div className='fx-agent-plugin-config-row'>
        <span>状态 {environment.status}</span>
        <span>平台 {environment.platforms.length ? environment.platforms.join(' / ') : BLOCKED}</span>
      </div>
      <Tags items={bundledTools.length ? bundledTools : ['自带工具证据未返回']} />
      {!!blockers.length && <Blocked>{blockers.join(' / ')}</Blocked>}
    </div>
  )
}

function PackageState({ item }) {
  return (
    <>
      <div className='fx-agent-subtitle'>包状态</div>
      <Tags items={[`源码 ${sourceStateText(item.sourceState)}`, `包 ${sourceStateText(item.packageState)}`, `签名 ${item.signature || 'missing'}`]} />
    </>
  )
}

function PackageConfig({ item }) {
  return (
    <>
      <div className='fx-agent-subtitle'>数据能力</div>
      <Tags items={item.telemetryKinds} />
      <div className='fx-agent-subtitle'>可下发配置</div>
      <Tags items={item.configTemplateIds} />
      <div className='fx-agent-config-list'>{item.configKeys.map(key => <code key={key}>{key}</code>)}</div>
    </>
  )
}

function PackagePluginConfig({ pluginConfig }) {
  return (
    <div className='fx-agent-plugin-config'>
      <div className='fx-agent-subtitle'>采集插件配置</div>
      <div className='fx-agent-plugin-config-row'>
        <span>格式 {pluginConfig.configFormat}</span>
        <span>Provider {pluginConfig.providerModes.join('/')}</span>
        <span>config_snippet_ref {pluginConfig.configSnippetRef}</span>
        <span>远程修改 {pluginConfig.remoteMutationStatus}</span>
      </div>
    </div>
  )
}

// --- P5: Package Registration Form ---

export function RegisterPackageForm({ onRegistered }) {
  const [form, setForm] = useState({ name: '', version: '', platform: 'linux', arch: 'amd64', size: '', checksum: '', url: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const update = (key, value) => setForm(prev => ({ ...prev, [key]: value }))

  const handleSubmit = (e) => {
    e.preventDefault()
    if (!form.name || !form.version) { setError('名称和版本必填'); return }
    setLoading(true)
    setError('')
    agentApi.registerAgentPackage({ ...form, size: Number(form.size) || 0 })
      .then(() => { setForm({ name: '', version: '', platform: 'linux', arch: 'amd64', size: '', checksum: '', url: '' }); onRegistered?.() })
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setLoading(false))
  }

  return (
    <form className='fx-agent-panel' onSubmit={handleSubmit}>
      <h4>注册新能力包</h4>
      <ErrorBox>{error}</ErrorBox>
      <div className='fx-agent-filter fx-agent-filter-3'>
        <Field label='名称'><input value={form.name} onChange={e => update('name', e.target.value)} required /></Field>
        <Field label='版本'><input value={form.version} onChange={e => update('version', e.target.value)} required /></Field>
        <Field label='平台'>
          <select value={form.platform} onChange={e => update('platform', e.target.value)}>
            <option value='linux'>Linux</option>
            <option value='windows'>Windows</option>
            <option value='darwin'>Darwin</option>
          </select>
        </Field>
      </div>
      <div className='fx-agent-filter fx-agent-filter-3'>
        <Field label='架构'>
          <select value={form.arch} onChange={e => update('arch', e.target.value)}>
            <option value='amd64'>amd64</option>
            <option value='arm64'>arm64</option>
          </select>
        </Field>
        <Field label='大小 (bytes)'><input type='number' value={form.size} onChange={e => update('size', e.target.value)} /></Field>
        <Field label='校验和'><input value={form.checksum} onChange={e => update('checksum', e.target.value)} placeholder='sha256' /></Field>
      </div>
      <Field label='下载 URL'><input value={form.url} onChange={e => update('url', e.target.value)} placeholder='https://...' style={{ width: '100%' }} /></Field>
      <div className='fx-agent-actions'>
        <button type='submit' disabled={loading}>{loading ? '提交中...' : '注册包'}</button>
      </div>
    </form>
  )
}

// --- P5: Package List Table ---

export function PackageListTable({ onRefresh }) {
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [deleting, setDeleting] = useState('')

  const load = () => {
    setLoading(true)
    setError('')
    agentApi.listAgentPackagesV2()
      .then(setRows)
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setLoading(false))
  }

  React.useEffect(() => { load() }, [])

  const handleDelete = (id) => {
    setDeleting(id)
    agentApi.deleteAgentPackage(id)
      .then(() => { load(); onRefresh?.() })
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setDeleting(''))
  }

  const fmtSize = (bytes) => {
    if (!bytes) return '-'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / 1048576).toFixed(1)} MB`
  }

  return (
    <div className='fx-agent-panel'>
      <div className='fx-agent-toolbar'>
        <h4>已注册包列表</h4>
        <button type='button' onClick={load} disabled={loading}>{loading ? '刷新中...' : '刷新'}</button>
      </div>
      <ErrorBox>{error}</ErrorBox>
      {!rows.length ? <p className='fx-agent-muted'>暂无已注册包</p> : (
        <div className='fx-agent-table'>
          <table>
            <thead><tr><th>名称</th><th>版本</th><th>平台</th><th>架构</th><th>大小</th><th>校验和</th><th>操作</th></tr></thead>
            <tbody>{rows.map(row => (
              <tr key={row.id}>
                <td>{row.name}</td>
                <td>{row.version}</td>
                <td>{row.platform}</td>
                <td>{row.arch}</td>
                <td>{fmtSize(row.size)}</td>
                <td><span className='fx-agent-muted'>{(row.checksum || '-').slice(0, 16)}{row.checksum?.length > 16 ? '...' : ''}</span></td>
                <td className='fx-agent-actions'>
                  <button type='button' disabled={deleting === row.id} onClick={() => handleDelete(row.id)}>
                    {deleting === row.id ? '删除中...' : '删除'}
                  </button>
                </td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}
    </div>
  )
}
