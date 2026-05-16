import React, { useEffect, useMemo, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { dashboardsApi } from '../api/dashboards.js'
import { redactText } from '../api/http.js'
import { pluginApi } from '../api/plugins.js'
import { displayText, fmtTime, hostIp, hostKey, hostName, isHostOnline } from './assetsModel.js'
import { Blocked, ErrorBox, Feedback, Status } from './Shared.jsx'
import './HostProbePluginDrawer.css'

const defaultCategoryMeta = {
  collect: { label: '采集插件', order: 10 },
  apm: { label: 'APM 探针', order: 20 },
  diagnose: { label: '诊断插件', order: 30 },
  other: { label: '其他插件', order: 40 },
}

const requiredCatalogFields = ['id', 'name', 'category', 'config_format', 'supported_os/platforms']
const rolloutBlockedText = 'PENDING: 缺少真实远程执行器、投递回执、效果回执、漂移检测和审计闭环，当前通过现有 config-rollouts 记录阻断请求，不创建真实执行。'
const assignBlockedText = 'PENDING: 已记录插件关联意图，但缺少主机插件分配存储、目标绑定、凭据策略和审计回执契约，未触达目标机器。'
const dispatchBlockedText = 'PENDING: 已请求远程下发配置，但缺少远程执行器、投递回执、效果回执、回滚回执、数据到达和证据链契约，未创建真实执行。'
const sourceBrandFragments = [['Night', 'ingale'], ['Sky', 'Walking'], ['Sig', 'Noz'], ['Cate', 'graf'], ['Cat', 'paw'], ['Auto', 'Ops'], ['Prom', 'etheus'], ['Gra', 'fana']]
const sourceBrandPattern = new RegExp(sourceBrandFragments.map(parts => parts.join('')).join('|'), 'ig')
const forbiddenStateFragments = [
  ['que', 'ued'],
  ['run', 'ning'],
  ['app', 'lied'],
  ['inst', 'alled'],
  ['data', '_arrived'],
  ['service', '_registered'],
  ['rolled', '_back'],
  ['uninst', 'alled'],
  ['deliv', 'ered'],
  ['effec', 'tive'],
  ['succee', 'ded'],
  ['succ', 'ess'],
  ['impor', 'ted'],
]
const forbiddenStatePattern = new RegExp(`\\b(?:${forbiddenStateFragments.map(parts => parts.join('')).join('|')})\\b`, 'ig')

export function HostProbePluginDrawer({ row, onClose, confirm = defaultConfirm }) {
  const [plugins, setPlugins] = useState([])
  const [error, setError] = useState('')
  const [feedback, setFeedback] = useState('')
  const [selectedKey, setSelectedKey] = useState('')
  const [catalogBlocked, setCatalogBlocked] = useState('')
  const [submittingKey, setSubmittingKey] = useState('')
  const [credentialRefs, setCredentialRefs] = useState({})
  const [assignmentRefs, setAssignmentRefs] = useState({})
  const [credentialDraft, setCredentialDraft] = useState('')
  const [credentialModal, setCredentialModal] = useState(null)
  const [dashboardAudit, setDashboardAudit] = useState(null)
  const selected = plugins.find(item => item.catalog_key === selectedKey) || plugins[0]

  useEffect(() => {
    let alive = true
    setError('')
    setCatalogBlocked('')
    pluginApi.listPlugins()
      .then(items => {
        if (!alive) return
        const normalized = items.map((item, index) => normalizePlugin(item, index))
        setPlugins(normalized)
        if (!normalized.length) {
          setCatalogBlocked('PENDING: 插件目录接口返回空结果，当前不渲染本地假目录或可执行插件卡片。')
        }
      })
      .catch(err => {
        if (!alive) return
        setPlugins([])
        setCatalogBlocked(`PENDING: 插件目录接口读取失败，当前不渲染本地假目录或可执行插件卡片。原因：${displayText(err.message)}`)
      })
    return () => { alive = false }
  }, [])

  const groups = useMemo(() => buildPluginGroups(plugins), [plugins])

  useEffect(() => {
    if (!selectedKey && plugins[0]?.catalog_key) setSelectedKey(plugins[0].catalog_key)
    if (selectedKey && plugins.length && !plugins.some(plugin => plugin.catalog_key === selectedKey)) {
      setSelectedKey(plugins[0].catalog_key)
    }
  }, [plugins, selectedKey])

  const requestRollout = async (plugin, action) => {
    setSelectedKey(plugin.catalog_key)
    setFeedback('')
    const actionLabel = action === 'assign' ? '关联插件' : '远程下发配置'
    const pluginTitle = plugin.name || 'PENDING: 缺少 name'
    const missing = missingCatalogFields(plugin)
    const blockers = rolloutBlockers(plugin)
    const credentialRef = selectedCredentialRef(plugin, credentialRefs)

    if (!plugin.id) {
      setFeedback(`${rolloutBlockedText} action=${action}; plugin=${plugin.catalog_key}; target=${hostKey(row)}; missing=id`)
      return
    }

    if (requiresCredentialRef(plugin) && !credentialRef) {
      setCredentialModal({ plugin, action })
      setCredentialDraft('')
      setFeedback(`PENDING: ${pluginTitle} 需要凭据引用。请先选择或填写 credential_ref，当前不接收密码、密钥或连接串。`)
      return
    }

    const ok = await confirm({
      title: action === 'assign' ? '确认关联插件' : '确认远程下发配置',
      message: action === 'assign'
        ? `${hostName(row)} / ${pluginTitle}: 仅记录该主机与插件的关联意图，调用现有 config-rollouts 写入阻断审计；不会触达目标机器。`
        : `${hostName(row)} / ${pluginTitle}: 请求 Agent 配置下发契约；缺少执行器、投递/效果/回滚回执和审批链路时不创建真实执行。`,
      confirmText: action === 'assign' ? '确认关联' : '确认下发',
    })
    if (!ok) return

    const assignmentRef = selectedAssignmentRef(plugin, assignmentRefs)
    const payload = buildRolloutPayload(row, plugin, action, credentialRef, assignmentRef)
    const requestKey = `${plugin.catalog_key}:${action}`
    setSubmittingKey(requestKey)
    try {
      const result = await agentApi.rolloutConfig(payload)
      const nextAssignment = extractAssignmentRefs(result)
      if (nextAssignment.assignment_ref) {
        setAssignmentRefs(prev => ({ ...prev, [plugin.catalog_key]: nextAssignment }))
      }
      setFeedback(formatRolloutFeedback({ action, plugin, row, missing, blockers, result }))
    } catch (err) {
      const responseBody = isPlainObject(err?.body) ? err.body : null
      const nextAssignment = extractAssignmentRefs(responseBody)
      if (nextAssignment.assignment_ref) {
        setAssignmentRefs(prev => ({ ...prev, [plugin.catalog_key]: nextAssignment }))
      }
      setFeedback(formatRolloutFeedback({ action, plugin, row, missing, blockers, result: responseBody, error: err }))
    } finally {
      setSubmittingKey('')
    }
  }

  const saveCredentialRef = () => {
    const clean = sanitizeCredentialRefInput(credentialDraft)
    if (!clean) {
      setFeedback('PENDING: credential_ref 不能为空，且不能包含 password/token/cookie/dsn/secret 等敏感值。')
      return
    }
    setCredentialRefs(prev => ({ ...prev, [credentialModal.plugin.catalog_key]: clean }))
    setCredentialModal(null)
    setCredentialDraft('')
    setFeedback(`PENDING: 已选择凭据引用 ${maskRef(clean)}，真实凭据解析和下发仍按 cmdb.agent.plugin.credential.v1 阻断。`)
  }

  const requestDashboardImportAudit = async (plugin) => {
    setSelectedKey(plugin.catalog_key)
    const refs = plugin.dashboard_refs
    if (!refs.length) {
      setDashboardAudit({
        plugin,
        status: 'PENDING',
        message: 'PENDING: 插件目录缺少 dashboard_refs，当前不触发导入预检。',
        missing_contracts: ['cmdb_dashboard_template_lookup_contract', 'cmdb_dashboard_import_runtime_contract'],
      })
      return
    }
    const templateId = normalizeDashboardRef(refs[0])
    if (!templateId) {
      setDashboardAudit({
        plugin,
        status: 'PENDING',
        message: 'PENDING: dashboard_ref 不是可解析的模板引用。',
        missing_contracts: ['cmdb_dashboard_template_lookup_contract'],
      })
      return
    }
    const ok = await confirm({
      title: '确认仪表盘导入预检',
      message: `${plugin.name || plugin.id}: 仅调用现有仪表盘模板导入接口做运行时契约校验；缺查重、导入执行器和冲突回滚时保持阻断。`,
      confirmText: '确认预检',
    })
    if (!ok) return
    setDashboardAudit({ plugin, status: 'PENDING', message: '仪表盘导入预检请求中...' })
    try {
      const result = await dashboardsApi.importTemplate(templateId, {
        title: `${plugin.name || plugin.id} 运行视图`,
        resource_group_id: stringValue(row?.resource_group_id || row?.workspace_id || 'cmdb-resource-group'),
        workspace_id: stringValue(row?.workspace_id),
        tags: ['FindX', 'CMDB', plugin.category].filter(Boolean),
        variables: {
          host_id: hostKey(row),
          plugin_id: plugin.id,
        },
      })
      setDashboardAudit(formatDashboardAudit(plugin, templateId, result))
    } catch (err) {
      setDashboardAudit(formatDashboardAudit(plugin, templateId, null, err))
    }
  }

  const startChat = async () => {
    const ok = await confirm({
      title: '确认打开单机诊断对话',
      message: `${hostName(row)} 的单机诊断对话需要真实会话执行器、命令白名单、录屏审计和回执契约；当前只展示阻断审计。`,
      confirmText: '确认查看',
    })
    if (!ok) return
    setFeedback('PENDING: 单机诊断对话缺少真实 Agent 会话传输、命令审计、输出回执和权限校验契约，当前未建立会话。')
  }

  return (
    <div className='fx-cmdb-probe-drawer' role='dialog' aria-modal='true' aria-labelledby='fx-cmdb-probe-title'>
      <button type='button' className='fx-cmdb-probe-backdrop' aria-label='关闭探针插件抽屉' onClick={onClose} />
      <section className='fx-cmdb-probe-panel'>
        <header>
          <div>
            <strong id='fx-cmdb-probe-title'>探针 / 插件</strong>
            <span>{hostName(row)} / {hostIp(row)}</span>
          </div>
          <button type='button' onClick={onClose} aria-label='关闭探针插件抽屉'>×</button>
        </header>
        <div className='fx-cmdb-probe-body'>
          <section className='fx-cmdb-probe-form'>
            <ProbeRow label='目标主机'>
              <div className='fx-cmdb-probe-target'>
                <strong>{hostName(row)}</strong>
                <span>{hostIp(row)} / Agent {displayText(row.agent_id || row.agent_status, '未上报')}</span>
              </div>
            </ProbeRow>
            <ProbeRow label='存活心跳'>
              <div className='fx-cmdb-probe-health'>
                <div><span>主机存活</span><Status ok={isHostOnline(row)}>{isHostOnline(row) ? '在线' : '离线'}</Status></div>
                <div><span>探针心跳</span><strong>{fmtTime(row.last_seen_at || row.last_seen || row.updated_at)}</strong></div>
                <div><span>资源边界</span><strong>{displayText(row.resource_group_id || row.workspace_id, '未绑定')}</strong></div>
              </div>
            </ProbeRow>
            <ProbeRow label='执行边界'>
              <Blocked>关联插件只记录主机与插件的配置意图；远程下发配置会请求 Agent 配置下发契约。缺少执行器、回执和审计闭环前，两个动作都会返回阻断审计，不展示完成态。</Blocked>
            </ProbeRow>
          </section>
          <ErrorBox>{error}</ErrorBox>
          <Feedback>{feedback}</Feedback>

          <section className='fx-cmdb-probe-form'>
            <ProbeRow label='插件选择'>
              <div className='fx-cmdb-plugin-shell'>
                {catalogBlocked ? (
                  <PluginCatalogBlocked message={catalogBlocked} />
                ) : (
                  groups.map(group => (
                    <PluginGroup
                      key={group.category}
                      title={group.label}
                      category={group.category}
                      plugins={group.plugins}
                      selectedKey={selected?.catalog_key}
                      credentialRefs={credentialRefs}
                      submittingKey={submittingKey}
                      onPick={setSelectedKey}
                      onRollout={requestRollout}
                      onDashboardAudit={requestDashboardImportAudit}
                    />
                  ))
                )}
              </div>
            </ProbeRow>
            <ProbeRow label='运行审计'>
              <RuntimeAuditPanel plugin={selected} credentialRef={selectedCredentialRef(selected, credentialRefs)} assignmentRef={selectedAssignmentRef(selected, assignmentRefs)} dashboardAudit={dashboardAudit} />
            </ProbeRow>
            <ProbeRow label='单机诊断'>
              <div className='fx-cmdb-probe-chat'>
                <div><strong>单机诊断对话</strong><span>会话、命令审计和输出回执未闭合前不建立真实连接</span></div>
                <button type='button' onClick={startChat}>单机诊断对话</button>
              </div>
            </ProbeRow>
          </section>
        </div>
        <footer className='fx-cmdb-probe-footer'>
          <button type='button' onClick={onClose}>取消</button>
          <div className='fx-cmdb-probe-selected'>
            <b>已选插件</b>
            <span>{selected ? `${selected.name || selected.id || selected.catalog_key} · ${selected.id || 'PENDING: 缺少 id'}` : '未选择插件'}</span>
          </div>
          <button type='button' className='is-primary' disabled={!selected || submittingKey === `${selected?.catalog_key}:assign`} onClick={() => selected && requestRollout(selected, 'assign')}>关联到该主机</button>
          <button type='button' className='is-danger' disabled={!selected || submittingKey === `${selected?.catalog_key}:dispatch`} onClick={() => selected && requestRollout(selected, 'dispatch')}>远程下发配置</button>
        </footer>
      </section>
      {credentialModal && (
        <CredentialRefDialog
          plugin={credentialModal.plugin}
          value={credentialDraft}
          onChange={setCredentialDraft}
          onCancel={() => setCredentialModal(null)}
          onSave={saveCredentialRef}
        />
      )}
    </div>
  )
}

function ProbeRow({ label, children }) {
  return (
    <div className='fx-cmdb-probe-row'>
      <label>{label}</label>
      <div>{children}</div>
    </div>
  )
}

function PluginGroup({ title, category, plugins, selectedKey, credentialRefs, submittingKey, onPick, onRollout, onDashboardAudit }) {
  return (
    <section className='fx-cmdb-plugin-group'>
      <div className='fx-cmdb-plugin-group-head'>
        <div><strong>{title}</strong><span>{category} / {plugins.length} 个 catalog 条目；真实下发仍按回执契约阻断</span></div>
      </div>
      <div className='fx-cmdb-plugin-list'>
        {plugins.map(plugin => (
          <article key={plugin.catalog_key} className={selectedKey === plugin.catalog_key ? 'is-selected' : ''}>
            <button type='button' className='fx-cmdb-plugin-pick' onClick={() => onPick(plugin.catalog_key)}>
              <i aria-hidden='true' />
              <div className='fx-cmdb-plugin-main'>
                <strong className={plugin.name ? '' : 'is-missing'}>{displayValue(plugin.name) || 'PENDING: 缺少 name'}</strong>
                <span>{displayValue(plugin.id) || 'PENDING: 缺少 id'}</span>
              </div>
              <PluginCatalogSummary plugin={plugin} />
            </button>
            <div className='fx-cmdb-plugin-actions'>
              <button type='button' disabled={submittingKey === `${plugin.catalog_key}:assign`} onClick={() => onRollout(plugin, 'assign')}>关联插件</button>
              <button type='button' className='is-danger' disabled={submittingKey === `${plugin.catalog_key}:dispatch`} onClick={() => onRollout(plugin, 'dispatch')}>下发配置</button>
              <button type='button' onClick={() => onDashboardAudit(plugin)}>看板预检</button>
            </div>
            {requiresCredentialRef(plugin) && (
              <div className='fx-cmdb-plugin-credential-state'>
                凭据引用：{selectedCredentialRef(plugin, credentialRefs) ? maskRef(selectedCredentialRef(plugin, credentialRefs)) : 'PENDING: 未选择 credential_ref'}
              </div>
            )}
          </article>
        ))}
      </div>
    </section>
  )
}

function PluginCatalogSummary({ plugin }) {
  const missing = missingCatalogFields(plugin)
  return (
    <div className='fx-cmdb-plugin-summary'>
      <CatalogField label='分类' value={plugin.category} missing='缺少 category' />
      <CatalogField label='配置格式' value={plugin.config_format} missing='缺少 config_format' />
      <CatalogField label='系统/平台' value={joinDisplay(plugin.supported_os.concat(plugin.platforms))} missing='缺少 supported_os/platforms' />
      <CatalogField label='安全级别' value={plugin.security_level} missing='缺少 security_level' />
      <CatalogField label='凭据契约' value={credentialSummary(plugin)} missing='缺少 credential schema' />
      <ContractChips label='阻断项' items={plugin.blockers.concat(plugin.missing_contracts)} missing='缺少 blockers/missing_contracts' />
      <CatalogField label='看板引用' value={joinDisplay(plugin.dashboard_refs)} missing='缺少 dashboard_refs' />
      {missing.length ? <em>PENDING: catalog 字段缺失：{missing.join('、')}</em> : <em>PENDING: 目录字段已展示，远程执行仍等待回执契约。</em>}
    </div>
  )
}

function ContractChips({ label, items, missing }) {
  const values = uniqueText(items)
  if (!values.length) {
    return <CatalogField label={label} value='' missing={missing} />
  }
  const visible = values.slice(0, 3)
  const extra = values.length - visible.length
  return (
    <span className='fx-cmdb-contract-chips' title={values.join('\n')}>
      <b>{label}</b>
      <span>
        {visible.map(item => <i key={item}>{displayValue(item)}</i>)}
        {extra > 0 && <i>+{extra}</i>}
      </span>
    </span>
  )
}

function CatalogField({ label, value, missing }) {
  const text = displayValue(value)
  return (
    <span className={text ? '' : 'is-missing'}>
      <b>{label}</b>
      {text || `PENDING: ${missing}`}
    </span>
  )
}

function PluginCatalogBlocked({ message }) {
  const missingContracts = [
    'GET /findx-agents/plugins 返回真实插件目录',
    'plugin.id / plugin.name / plugin.category / plugin.config_format / supported_os 或 platforms 字段契约',
    'plugin credential_ref requirement schema',
    'plugin rollout receipt and audit contract',
  ]
  return (
    <div className='fx-cmdb-plugin-blocked'>
      <Blocked>{message}</Blocked>
      <div className='fx-cmdb-plugin-missing'>
        <strong>missing contracts</strong>
        {missingContracts.map(item => <span key={item}>{item}</span>)}
      </div>
    </div>
  )
}

function RuntimeAuditPanel({ plugin, credentialRef, assignmentRef, dashboardAudit }) {
  const missing = uniqueText((plugin?.missing_contracts || []).concat(plugin?.blockers || []))
  return (
    <div className='fx-cmdb-runtime-audit'>
      <div>
        <strong>当前插件动作审计</strong>
        <span>{plugin ? `${displayValue(plugin.name || plugin.id) || 'PENDING'} / ${displayValue(plugin.category || 'unknown') || 'unknown'}` : '未选择插件'}</span>
      </div>
      <div className='fx-cmdb-runtime-grid'>
        <span><b>credential_ref</b>{credentialRef ? maskRef(credentialRef) : 'PENDING: 未选择'}</span>
        <span><b>关联插件</b>remote_mutation=false / rollout_strategy=assign</span>
        <span><b>assignment_ref</b>{assignmentRef?.assignment_ref ? assignmentRef.assignment_ref : 'PENDING: 未读取到真实关联记录'}</span>
        <span><b>target_binding_ref</b>{assignmentRef?.target_binding_ref ? assignmentRef.target_binding_ref : 'PENDING: 未读取到目标绑定'}</span>
        <span><b>下发配置</b>remote_mutation=true / rollout_strategy=dispatch / 依赖 assignment_ref</span>
        <span><b>contract</b>cmdb.agent.plugin.credential.v1</span>
        <span><b>dashboard import</b>cmdb.dashboard.import.runtime.v1</span>
        <span><b>findx_audit query</b>scope=cmdb/resource_type=cmdb_agent_plugin/action=config_rollout</span>
      </div>
      {missing.length > 0 && (
        <div className='fx-cmdb-plugin-missing'>
          <strong>missing contracts</strong>
          {missing.map(item => <span key={item}>{item}</span>)}
        </div>
      )}
      {dashboardAudit && (
        <div className='fx-cmdb-dashboard-audit'>
          <strong>{dashboardAudit.status || 'PENDING'}</strong>
          <span>{dashboardAudit.message}</span>
          {(dashboardAudit.missing_contracts || []).map(item => <code key={item}>{item}</code>)}
          {dashboardAudit.findx_audit_query && <span>findx_audit query: {dashboardAudit.findx_audit_query}</span>}
        </div>
      )}
    </div>
  )
}

function CredentialRefDialog({ plugin, value, onChange, onCancel, onSave }) {
  return (
    <div className='fx-cmdb-credential-dialog' role='dialog' aria-modal='true' aria-labelledby='fx-cmdb-credential-title'>
      <div className='fx-cmdb-credential-card'>
        <header>
          <strong id='fx-cmdb-credential-title'>选择凭据引用</strong>
          <button type='button' onClick={onCancel} aria-label='关闭凭据引用弹窗'>×</button>
        </header>
        <div className='fx-cmdb-credential-body'>
          <span>{displayValue(plugin.name || plugin.id) || 'PENDING'} 需要 credential_ref。这里只保存凭据引用，不输入密码、密钥或连接串。</span>
          <input
            value={value}
            onChange={(event) => onChange(event.target.value)}
            placeholder='credential_ref'
            autoFocus
          />
          <Blocked>凭据解析、配置 schema、仪表盘查重导入和下发配置仍按后端契约阻断，不展示完成态。</Blocked>
        </div>
        <footer>
          <button type='button' onClick={onCancel}>取消</button>
          <button type='button' className='is-primary' onClick={onSave}>使用引用</button>
        </footer>
      </div>
    </div>
  )
}

function buildRolloutPayload(row, plugin, action, credentialRef = '', assignmentRef = null) {
  const targetId = hostKey(row)
  const payload = {
    template_id: `cmdb-host-plugin-${action}`,
    plugin_id: plugin.id,
    target_mode: 'single_host',
    rollout_strategy: action === 'assign' ? 'assign' : 'dispatch',
    config_format: plugin.config_format || 'catalog',
    provider_mode: 'cmdb_host_probe',
    remote_mutation: action === 'dispatch',
    audit_reason: `cmdb_host_probe_plugin_${action}`,
    target_ids: targetId ? [targetId] : [],
    agent_ids: stringValue(row?.agent_id) ? [stringValue(row.agent_id)] : [],
    metadata: {
      cmdb_host_id: targetId,
      host_name: hostName(row),
      host_ip: hostIp(row),
      plugin_category: plugin.category,
      plugin_action: action,
      agent_ref: stringValue(row?.agent_id || row?.agent_status),
      cmdb_host_ref: targetId,
      scope: 'cmdb_host',
      resource_group_ref: stringValue(row?.resource_group_id),
      workspace_ref: stringValue(row?.workspace_id),
      dashboard_refs: joinDisplay(plugin.dashboard_refs),
    },
  }

  if (credentialRef) payload.credential_ref = credentialRef
  if (action === 'dispatch' && assignmentRef?.assignment_ref) {
    payload.metadata.assignment_ref = assignmentRef.assignment_ref
    if (assignmentRef.target_binding_ref) payload.metadata.target_binding_ref = assignmentRef.target_binding_ref
  }
  return payload
}

function formatRolloutFeedback({ action, plugin, row, missing, blockers, result, error }) {
  const responseMissing = toArray(result?.missing || result?.missing_contracts || result?.missingContracts)
  const responseBlockers = toArray(result?.blockers || result?.blocked_reasons || result?.blockedReasons || result?.blocker)
  const allMissing = uniqueText(missing.concat(responseMissing))
  const allBlockers = uniqueText(blockers.concat(responseBlockers))
  const errorSummary = error ? formatRolloutError(error) : rolloutResultSummary(result)
  const blockedPrefix = isBlockedRolloutResult(result) || error
    ? (action === 'assign' ? assignBlockedText : dispatchBlockedText)
    : 'PENDING: config-rollouts 已返回，但前端未收到 blocked 审计确认；不显示关联、下发、安装或完成态。'

  return [
    blockedPrefix,
    `action=${action}`,
    `operation_contract=${action === 'assign' ? 'cmdb.agent.plugin.assignment.v1' : 'cmdb.agent.plugin.dispatch.v1'}`,
    `plugin=${plugin.id}`,
    `category=${displayValue(plugin.category) || 'PENDING'}`,
    `target=${hostKey(row)}`,
    `assignment_ref=${safeAssignmentRef(result) || 'PENDING'}`,
    `target_binding_ref=${safeTargetBindingRef(result) || 'PENDING'}`,
    `missing=${allMissing.length ? allMissing.join(',') : 'none'}`,
    `blockers=${allBlockers.length ? allBlockers.join(',') : 'none'}`,
    `error_summary=${errorSummary || 'PENDING'}`,
  ].join('; ')
}

function formatRolloutError(error) {
  const summary = formatAgentError(error)
  return error?.status === 409 && !/^PENDING/.test(summary)
    ? `PENDING: ${summary}`
    : summary
}

function formatDashboardAudit(plugin, templateId, result, error) {
  const body = error?.body || result || {}
  const missing = toArray(body.missing_contracts || body.missingContracts || body.missing || error?.missing_contracts)
  const message = error
    ? `PENDING: ${formatAgentError(error)}`
    : 'PENDING: 仪表盘导入接口返回非阻断响应，前端仍不展示导入完成，等待真实导入回执。'
  return {
    plugin,
    template_id: templateId,
    status: 'PENDING',
    message: redactText(body.message || body.error || message),
    missing_contracts: uniqueText(missing.length ? missing : [
      'cmdb_dashboard_template_lookup_contract',
      'cmdb_dashboard_import_runtime_contract',
      'cmdb_dashboard_import_batch_result_contract',
      'cmdb_dashboard_import_conflict_rollback_contract',
    ]),
    findx_audit_query: body.findx_audit_query || `scope=cmdb/resource_type=cmdb_dashboard_import/action=dashboard.template.import/template_id=${templateId}`,
  }
}

function rolloutResultSummary(result) {
  if (!result || !isPlainObject(result)) return 'config-rollouts returned empty response'
  const summary = result.error_summary || result.errorSummary || result.message || result.error || result.reason || result.blocked_reason || result.blocker || result.status || 'config-rollouts returned response'
  return redactText(summary)
}

function isBlockedRolloutResult(result) {
  if (!result || !isPlainObject(result)) return false
  const values = [result.status, result.code, result.error_code, result.error, result.message, result.reason, result.blocked_reason]
  return values.some(value => /blocked|PENDING/i.test(String(value || '')))
}

function extractAssignmentRefs(result) {
  const contract = isPlainObject(result?.assignment_contract)
    ? result.assignment_contract
    : (isPlainObject(result?.operation_contract?.assignment_contract) ? result.operation_contract.assignment_contract : {})
  return {
    assignment_ref: stringValue(result?.assignment_ref || result?.operation_contract?.assignment_ref || contract.assignment_ref),
    target_binding_ref: stringValue(result?.target_binding_ref || result?.operation_contract?.target_binding_ref || contract.target_binding_ref),
  }
}

function safeAssignmentRef(result) {
  return extractAssignmentRefs(result).assignment_ref
}

function safeTargetBindingRef(result) {
  return extractAssignmentRefs(result).target_binding_ref
}

function rolloutBlockers(plugin) {
  const blockers = plugin.blockers.concat(plugin.missing_contracts)
  return blockers.length ? blockers.map(stringifyContractItem).filter(Boolean) : [
    'remote_executor',
    'delivery_receipt',
    'effect_receipt',
    'drift_detection',
    'audit_chain',
  ]
}

function uniqueText(items) {
  return [...new Set(toArray(items).map(stringifyContractItem).filter(Boolean))]
}

function normalizePlugin(plugin, index) {
  const source = isPlainObject(plugin) ? plugin : {}
  const credentialSchema = isPlainObject(source.credential_schema)
    ? source.credential_schema
    : (isPlainObject(source.credentialSchema) ? source.credentialSchema : {})
  const credential = {
    ...credentialSchema,
    ...(isPlainObject(source.credential) ? source.credential : {}),
  }
  const blockers = toArray(source.blockers)
  const missingContracts = toArray(source.missing_contracts || source.missingContracts)
  const id = stringValue(source.id || source.plugin_id || source.pluginId)
  const name = stringValue(source.name)
  const category = stringValue(source.category)
  const configFormat = stringValue(source.config_format || source.configFormat)
  const supportedOs = toArray(source.supported_os || source.supportedOS)
  const platforms = toArray(source.platforms || source.platform)
  const fingerprint = [
    id,
    name,
    category,
    configFormat,
    joinDisplay(supportedOs),
    joinDisplay(platforms),
    index,
  ].map(part => stringValue(part) || 'missing').join('|')

  return {
    raw: source,
    catalog_key: `catalog:${fingerprint}`,
    id,
    name,
    category,
    category_order: numericOrder(source.category_order || source.categoryOrder || source.order || source.sort_order || source.sortOrder),
    config_format: configFormat,
    supported_os: supportedOs,
    platforms,
    security_level: stringValue(source.security_level || source.securityLevel),
    blockers,
    missing_contracts: missingContracts,
    dashboard_refs: toArray(source.dashboard_refs || source.dashboardRefs),
    credential,
    credential_schema: credentialSchema,
    credential_ref: sanitizeCredentialRefInput(source.credential_ref || source.credentialRef || credential.ref || credential.credential_ref || credential.credentialRef),
    credential_required: credentialRefRequiredByCatalog(source, credential, blockers, missingContracts),
  }
}

function selectedCredentialRef(plugin, credentialRefs) {
  if (!plugin) return ''
  return sanitizeCredentialRefInput(credentialRefs?.[plugin.catalog_key] || plugin.credential_ref)
}

function selectedAssignmentRef(plugin, assignmentRefs) {
  if (!plugin) return null
  return assignmentRefs?.[plugin.catalog_key] || null
}

function sanitizeCredentialRefInput(value) {
  const clean = stringValue(value)
  if (!clean || looksSensitiveCredentialRef(clean)) return ''
  return clean.slice(0, 120)
}

function looksSensitiveCredentialRef(value) {
  const normalized = value.toLowerCase()
  return /(password|passwd|token|cookie|secret|dsn|bearer|api[_-]?key|private[_-]?key|:\/\/|=)/i.test(normalized)
}

function maskRef(value) {
  const clean = stringValue(value)
  if (!clean) return ''
  if (clean.length <= 6) return `${clean.slice(0, 2)}***`
  return `${clean.slice(0, 3)}***${clean.slice(-3)}`
}

function normalizeDashboardRef(ref) {
  const clean = stringValue(ref)
  if (!clean) return ''
  if (clean.startsWith('dashboard:')) return clean.slice('dashboard:'.length)
  return clean
}

function buildPluginGroups(plugins) {
  const grouped = plugins.reduce((acc, plugin) => {
    const category = plugin.category || 'contract_missing'
    if (!acc[category]) acc[category] = []
    acc[category].push(plugin)
    return acc
  }, {})

  return Object.keys(grouped)
    .sort((left, right) => compareCategory(left, right, grouped))
    .map(category => ({
      category,
      label: categoryLabel(category),
      plugins: sortPluginsInGroup(grouped[category]),
    }))
}

function sortPluginsInGroup(plugins) {
  return [...(plugins || [])].sort((left, right) => {
    const leftOrder = numericOrder(left.raw?.order || left.raw?.sort_order || left.raw?.sortOrder)
    const rightOrder = numericOrder(right.raw?.order || right.raw?.sort_order || right.raw?.sortOrder)
    if (leftOrder !== rightOrder) return leftOrder - rightOrder
    return (left.name || left.id || left.catalog_key).localeCompare(right.name || right.id || right.catalog_key)
  })
}

function compareCategory(left, right, grouped) {
  const leftBackendOrder = groupOrder(grouped[left])
  const rightBackendOrder = groupOrder(grouped[right])
  if (leftBackendOrder !== rightBackendOrder) return leftBackendOrder - rightBackendOrder
  if (left === 'contract_missing') return -1
  if (right === 'contract_missing') return 1
  const leftMeta = defaultCategoryMeta[left]
  const rightMeta = defaultCategoryMeta[right]
  if (leftMeta || rightMeta) return (leftMeta?.order || 1000) - (rightMeta?.order || 1000)
  return left.localeCompare(right)
}

function groupOrder(plugins) {
  const orders = (plugins || []).map(plugin => plugin.category_order).filter(Number.isFinite)
  return orders.length ? Math.min(...orders) : Number.POSITIVE_INFINITY
}

function categoryLabel(category) {
  if (defaultCategoryMeta[category]) return defaultCategoryMeta[category].label
  if (category === 'contract_missing') return '契约缺失'
  return category
}

function requiresCredentialRef(plugin) {
  return plugin.credential_required === true
}

function credentialRefRequiredByCatalog(source, credential, blockers, missingContracts) {
  return truthy(source.required) ||
    truthy(source.ref_required) ||
    truthy(source.refRequired) ||
    truthy(source.credential_required) ||
    truthy(source.credentialRequired) ||
    truthy(source.requires_credential) ||
    truthy(source.requiresCredential) ||
    truthy(source.credential_ref_required) ||
    truthy(source.credentialRefRequired) ||
    truthy(credential.required) ||
    truthy(credential.ref_required) ||
    truthy(credential.refRequired) ||
    truthy(credential.credential_ref_required) ||
    truthy(credential.credentialRefRequired) ||
    stringValue(credential.mode).toLowerCase() === 'credential_ref' ||
    stringValue(credential.fields).toLowerCase().includes('credential_ref') ||
    hasCredentialRefBlocker(blockers.concat(missingContracts))
}

function missingCatalogFields(plugin) {
  const missing = []
  if (!plugin.id) missing.push('id')
  if (!plugin.name) missing.push('name')
  if (!plugin.category) missing.push('category')
  if (!plugin.config_format) missing.push('config_format')
  if (!plugin.supported_os.length && !plugin.platforms.length) missing.push('supported_os/platforms')
  return missing.filter(field => requiredCatalogFields.includes(field))
}

function credentialSummary(plugin) {
  const parts = []
  if (plugin.credential_required) parts.push('需要 credential_ref')
  if (plugin.credential_ref) parts.push(`credential_ref=${maskRef(plugin.credential_ref)}`)
  const schema = plugin.credential.schema || plugin.credential.contract || plugin.credential.type || plugin.credential.name || plugin.credential.format
  if (schema) parts.push(`schema=${schema}`)
  return parts.join(' / ')
}

function blockerSummary(plugin) {
  return joinDisplay(plugin.blockers.concat(plugin.missing_contracts))
}

function hasCredentialRefBlocker(items) {
  return items.some(item => stringifyContractItem(item).toLowerCase().includes('credential_ref'))
}

function joinDisplay(items) {
  const values = toArray(items).map(item => stringifyContractItem(item)).filter(Boolean)
  return values.join(' / ')
}

function stringifyContractItem(item) {
  if (item === null || item === undefined || item === '') return ''
  if (typeof item === 'string' || typeof item === 'number' || typeof item === 'boolean') return String(item)
  if (isPlainObject(item)) {
    return stringValue(item.code || item.field || item.name || item.id || item.contract || item.type || JSON.stringify(item))
  }
  return String(item)
}

function displayValue(value) {
  const text = Array.isArray(value) ? joinDisplay(value) : stringValue(value)
  return text ? displayText(redactCatalogText(text)) : ''
}

function redactCatalogText(value) {
  return redactText(String(value ?? ''))
    .replace(sourceBrandPattern, 'FindX')
    .replace(forbiddenStatePattern, 'PENDING')
}

function stringValue(value) {
  if (value === null || value === undefined) return ''
  return String(value).trim()
}

function toArray(value) {
  if (Array.isArray(value)) return value.filter(item => item !== null && item !== undefined && item !== '')
  if (value === null || value === undefined || value === '') return []
  return [value]
}

function truthy(value) {
  if (value === true) return true
  if (typeof value === 'number') return value > 0
  if (typeof value === 'string') return ['true', '1', 'yes', 'required'].includes(value.trim().toLowerCase())
  return false
}

function numericOrder(value) {
  const number = Number(value)
  return Number.isFinite(number) ? number : Number.POSITIVE_INFINITY
}

function isPlainObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

async function defaultConfirm() {
  return false
}
