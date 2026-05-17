import React, { useEffect, useState } from 'react'
import { AGENT_BLOCKERS, agentApi, formatAgentError } from '../api/agents.js'
import { configTemplates } from './agentModel.js'
import { Blocked, CopyBlock, Tags } from './AgentShared.jsx'

const localTemplateMap = new Map(configTemplates.map(item => [item.id, item]))
const normalizePluginConfig = value => value ? {
  pluginId: value.pluginId || value.plugin_id || '',
  pluginVersion: value.pluginVersion || value.plugin_version || '<PLUGIN_VERSION>',
  configFormat: value.configFormat || value.config_format || 'toml',
  configSnippetRef: value.configSnippetRef || value.config_snippet_ref || '<CONFIG_SNIPPET_REF>',
  providerModes: value.providerModes || value.provider_modes || ['local', 'http'],
  reloadStrategy: value.reloadStrategy || value.reload_strategy || '',
  restartStrategy: value.restartStrategy || value.restart_strategy || '',
  remoteMutation: value.remoteMutation ?? value.remote_mutation ?? false,
  remoteMutationStatus: value.remoteMutationStatus || value.remote_mutation_status || '',
  rolloutMetadata: value.rolloutMetadata || value.rollout_metadata || [],
  credentialRefRequired: value.credentialRefRequired ?? value.credential_ref_required ?? false,
  auditEvent: value.auditEvent || value.audit_event || '',
} : null

const normalizeTemplate = item => {
  const fallback = localTemplateMap.get(item.id) || {}
  return {
    ...fallback,
    ...item,
    configKind: item.configKind || item.config_kind || fallback.configKind || '配置模板',
    fields: item.fields || fallback.fields || [],
    targetModes: item.targetModes || item.target_scopes || fallback.targetModes || fallback.targetScopes || ['全部 Agent', '业务组', '主机'],
    rolloutScopes: item.rolloutScopes || item.rollout_scopes || fallback.rolloutScopes || item.target_scopes || fallback.targetScopes || [],
    rolloutStrategies: item.rolloutStrategies || item.rollout_strategies || fallback.rolloutStrategies || ['保存模板', '灰度下发', '全量下发', '回滚'],
    remoteDelivery: item.remoteDelivery ?? item.remote_distribution ?? fallback.remoteDelivery ?? fallback.remoteDistribution ?? true,
    rollbackPolicy: item.rollbackPolicy || item.rollback_policy || fallback.rollbackPolicy || '按配置版本回滚到上一套稳定版本',
    capabilityPackages: item.capabilityPackages || item.capability_packages || fallback.capabilityPackages || ['能力包'],
    blocker: item.blocker || fallback.blocker,
    pluginConfig: normalizePluginConfig(item.pluginConfig || item.plugin_config || fallback.pluginConfig),
  }
}

export function TemplatesSection() {
  const initialTemplates = configTemplates.map(normalizeTemplate)
  const [selected, setSelected] = useState(initialTemplates[0])
  const [templates, setTemplates] = useState(initialTemplates)
  const [blocked, setBlocked] = useState('')
  const [targetMode, setTargetMode] = useState(initialTemplates[0].targetModes?.[0] || '全部 Agent')
  const [strategy, setStrategy] = useState(initialTemplates[0].rolloutStrategies?.[1] || '灰度下发')

  useEffect(() => {
    let alive = true
    agentApi.templates()
      .then(rows => {
        if (alive && rows.length) {
          const mapped = rows.map(normalizeTemplate)
          setTemplates(mapped)
          setSelected(mapped[0])
          setTargetMode(mapped[0].targetModes?.[0] || '全部 Agent')
          setStrategy(mapped[0].rolloutStrategies?.[1] || mapped[0].rolloutStrategies?.[0] || '灰度下发')
        }
      })
      .catch(err => { if (alive) setBlocked(formatAgentError(err)) })
    return () => { alive = false }
  }, [])

  const chooseTemplate = item => {
    setSelected(item)
    setTargetMode(item.targetModes?.[0] || '全部 Agent')
    setStrategy(item.rolloutStrategies?.[1] || item.rolloutStrategies?.[0] || '灰度下发')
  }

  const rollout = rolloutStrategy => {
    setBlocked(`统一配置模板页缺少真实 CMDB/主机 Agent 目标选择器，请从主机 Agent 的配置/插件下发抽屉选择真实目标后执行 ${rolloutStrategy}。`)
  }

  const buildRolloutPayload = rolloutStrategy => ({
    template_id: selected.id,
    credential_ref: '<CREDENTIAL_REF>',
    config_version: '<CONFIG_ID>',
    config_snippet_ref: selected.pluginConfig?.configSnippetRef || '<CONFIG_SNIPPET_REF>',
    config_format: selected.pluginConfig?.configFormat || 'toml',
    provider_mode: selected.pluginConfig?.providerModes?.[1] || selected.pluginConfig?.providerModes?.[0] || 'local',
    plugin_id: selected.pluginConfig?.pluginId || '<PLUGIN_ID>',
    plugin_version: selected.pluginConfig?.pluginVersion || '<PLUGIN_VERSION>',
    reload_strategy: selected.pluginConfig?.reloadStrategy || '<RELOAD_STRATEGY>',
    restart_strategy: selected.pluginConfig?.restartStrategy || '<RESTART_STRATEGY>',
    rollout_strategy: rolloutStrategy,
    rollback_ref: '<ROLLBACK_REF>',
    audit_reason: '<AUDIT_REASON>',
    change_ticket: '<CHANGE_TICKET>',
    remote_mutation: Boolean(selected.pluginConfig?.remoteMutation),
    canary_percent: rolloutStrategy === '灰度下发' ? 10 : 100,
    target_selector: '在主机 Agent 抽屉中选择真实目标',
  })

  const payload = {
    ...buildRolloutPayload(strategy),
    template_id: selected.id,
    config_kind: selected.configKind,
    config_ref: '<CONFIG_ID>',
    remote_distribution: selected.remoteDelivery,
    target_mode: targetMode,
    rollback_policy: selected.rollbackPolicy,
    plugin_config: selected.pluginConfig ? {
      plugin_id: selected.pluginConfig.pluginId,
      plugin_version: selected.pluginConfig.pluginVersion,
      config_format: selected.pluginConfig.configFormat,
      config_snippet_ref: selected.pluginConfig.configSnippetRef,
      provider_mode: selected.pluginConfig.providerModes?.[1] || selected.pluginConfig.providerModes?.[0] || 'local',
      reload_strategy: selected.pluginConfig.reloadStrategy,
      restart_strategy: selected.pluginConfig.restartStrategy,
      remote_mutation: selected.pluginConfig.remoteMutation,
      canary_percent: strategy === '灰度下发' ? 10 : 100,
      rollback_ref: '<ROLLBACK_REF>',
      audit_reason: '<AUDIT_REASON>',
      change_ticket: '<CHANGE_TICKET>',
      credential_ref: '<CREDENTIAL_REF>',
    } : undefined,
    fields: Object.fromEntries(selected.fields.map(field => [field, `<${field.toUpperCase()}>`])),
  }

  return (
    <section className='fx-agent-split'>
      <div className='fx-agent-list'>
        {templates.map(item => (
          <button key={item.id} type='button' className={selected.id === item.id ? 'is-active' : ''} onClick={() => chooseTemplate(item)}>
            <strong>{item.name}</strong>
            <span>{item.configKind} · {item.scope}</span>
          </button>
        ))}
      </div>
      <div className='fx-agent-panel'>
        <h3>{selected.name}</h3>
        <p>{selected.scope}</p>
        <div className='fx-agent-template-meta'>
          <span>{selected.configKind}</span>
          <span>{selected.remoteDelivery ? '远程下发契约待开放' : '仅本地模板'}</span>
        </div>
        <div className='fx-agent-template-plan'>
          <label>
            <span>下发目标</span>
            <select value={targetMode} onChange={event => setTargetMode(event.target.value)}>
              {(selected.targetModes || []).map(item => <option key={item} value={item}>{item}</option>)}
            </select>
          </label>
          <label>
            <span>下发策略</span>
            <select value={strategy} onChange={event => setStrategy(event.target.value)}>
              {(selected.rolloutStrategies || []).map(item => <option key={item} value={item}>{item}</option>)}
            </select>
          </label>
        </div>
        <p className='fx-agent-muted'>覆盖范围：{(selected.rolloutScopes || []).join(' / ') || '按目标选择器下发'}；回滚策略：{selected.rollbackPolicy}</p>
        <div className='fx-agent-subtitle'>适用能力包</div>
        <Tags items={selected.capabilityPackages || []} />
        <div className='fx-agent-subtitle'>配置字段</div>
        <Tags items={selected.fields} />
        {selected.pluginConfig && <>
          <div className='fx-agent-subtitle'>采集插件远程修改</div>
          <Tags items={[
            `格式 ${selected.pluginConfig.configFormat}`,
            `Provider ${(selected.pluginConfig.providerModes || []).join('/')}`,
            selected.pluginConfig.remoteMutationStatus === '' ? '远程修改契约待开放' : '远程修改',
          ]} />
        </>}
        <CopyBlock>{JSON.stringify(payload, null, 2)}</CopyBlock>
        <div className='fx-agent-actions'>
          <button type='button' onClick={() => rollout('保存模板')}>保存模板</button>
          <button type='button' onClick={() => rollout('灰度下发')}>灰度下发</button>
          <button type='button' onClick={() => rollout('全量下发')}>全量下发</button>
          <button type='button' onClick={() => rollout('回滚')}>回滚</button>
        </div>
        {blocked && <Blocked>{blocked}</Blocked>}
        <Blocked>{selected.blocker || AGENT_BLOCKERS.configLifecycle}</Blocked>
      </div>
    </section>
  )
}
