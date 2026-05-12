import React, { useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { Blocked, ErrorBox, Status, Tags } from './AgentShared.jsx'

const CONTRACT_BLOCKED = 'BLOCKED_BY_CONTRACT: 已写入 blocked task audit record，真实执行仍缺少远程执行器、回执和 Evidence Chain 契约。'
const NO_TARGETS = '请选择至少一个 CMDB 主机或 Agent 目标'
const BASE_METADATA = {
  credential_ref: '<CREDENTIAL_REF>',
  transport: '<REMOTE_TRANSPORT_REF>',
  runner: '<REMOTE_RUNNER_REF>',
  target_os: '<TARGET_OS>',
  idempotency_key: '<IDEMPOTENCY_KEY>',
  timeout_policy_ref: '<TIMEOUT_POLICY_REF>',
  execution_receipt_ref: '<EXECUTION_RECEIPT_REF>',
  audit_ref: '<AUDIT_REF>',
  evidence_chain_ref: '<EVIDENCE_CHAIN_REF>',
}
const ACTIONS = [
  {
    action: 'uninstall',
    label: '卸载 Agent',
    refs: { uninstall_manifest_ref: '<UNINSTALL_MANIFEST_REF>', executor_ref: '<UNINSTALL_EXECUTOR_REF>' },
  },
  {
    action: 'upgrade',
    label: '升级 Agent',
    packageId: '<PACKAGE_ID>',
    refs: {
      package_id: '<PACKAGE_ID>',
      package_repository_ref: '<PACKAGE_REPOSITORY_REF>',
      signature_ref: '<SIGNATURE_REF>',
      checksum: '<CHECKSUM_REF>',
      script_manifest_ref: '<SCRIPT_MANIFEST_REF>',
      executor_ref: '<UPGRADE_EXECUTOR_REF>',
    },
  },
  {
    action: 'rollback',
    label: '回滚 Agent',
    refs: { rollback_manifest_ref: '<ROLLBACK_MANIFEST_REF>', state_snapshot_ref: '<STATE_SNAPSHOT_REF>', executor_ref: '<ROLLBACK_EXECUTOR_REF>' },
  },
  {
    action: 'restart',
    label: '重启 Agent',
    refs: { service_ref: '<SERVICE_REF>', executor_ref: '<RESTART_EXECUTOR_REF>' },
  },
]

const unique = values => Array.from(new Set((values || []).filter(Boolean)))
const hasBlockedPrefix = value => String(value || '').startsWith('BLOCKED_BY_CONTRACT')
const blockedError = err => {
  const message = formatAgentError(err)
  return hasBlockedPrefix(message) ? message : `BLOCKED_BY_CONTRACT: ${message}`
}

function buildTaskBody(action, targetIds, agentIds) {
  const cleanTargetIds = unique(targetIds)
  const cleanAgentIds = unique(agentIds)
  return {
    action: action.action,
    package_id: action.packageId || '',
    credential_ref: '<CREDENTIAL_REF>',
    target_ids: cleanTargetIds,
    agent_ids: cleanAgentIds,
    metadata: {
      ...BASE_METADATA,
      ...action.refs,
      target_count: String(cleanTargetIds.length + cleanAgentIds.length),
      action_surface: 'host-agent-lifecycle-actions',
    },
  }
}

export function HostLifecycleTaskActions({ targetIds = [], agentIds = [], targetCount = 0 }) {
  const [feedback, setFeedback] = useState('')
  const [busyAction, setBusyAction] = useState('')
  const cleanTargetIds = unique(targetIds)
  const cleanAgentIds = unique(agentIds)
  const selectedCount = targetCount || cleanTargetIds.length + cleanAgentIds.length

  const requestTask = action => {
    if (!cleanTargetIds.length && !cleanAgentIds.length) {
      setFeedback(NO_TARGETS)
      return
    }
    setBusyAction(action.action)
    setFeedback('')
    agentApi.task(buildTaskBody(action, cleanTargetIds, cleanAgentIds))
      .then(() => setFeedback(CONTRACT_BLOCKED))
      .catch(err => setFeedback(blockedError(err)))
      .finally(() => setBusyAction(''))
  }

  return (
    <section className='fx-agent-panel'>
      <h3>Host Agent 生命周期任务</h3>
      <p>仅创建 blocked task audit records；卸载、升级、回滚和重启不会进入真实执行。</p>
      <div className='fx-agent-summary-row'>
        <Status ok={false}>blocked-only</Status>
        <span>目标 {selectedCount} 个</span>
        <Tags items={[`target_ids ${cleanTargetIds.length}`, `agent_ids ${cleanAgentIds.length}`]} />
      </div>
      <div className='fx-agent-actions'>
        {ACTIONS.map(action => <button key={action.action} type='button' disabled={busyAction === action.action} onClick={() => requestTask(action)}>{busyAction === action.action ? '创建 blocked 审计中...' : action.label}</button>)}
      </div>
      {hasBlockedPrefix(feedback) ? <Blocked>{feedback}</Blocked> : <ErrorBox>{feedback}</ErrorBox>}
    </section>
  )
}
