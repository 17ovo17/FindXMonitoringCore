import React from 'react'

const CONTRACT_STATUS = ''
export const REMOTE_MUTATION_REF_KEYS = [
  'plugin_config_writer_ref',
  'reload_command_ref',
  'reload_receipt_ref',
  'drift_check_ref',
  'evidence_chain_ref',
  'rollback_ref',
  'rollback_receipt_ref',
  'provider_endpoint_ref',
  'provider_response_version_ref',
  'provider_checksum_ref',
  'provider_headers_ref',
  'provider_auth_ref',
  'provider_tls_ref',
  'provider_scope_ref',
  'provider_mutation_receipt_ref',
  'provider_audit_ref',
  'provider_record_detail_ref',
  'reload_interval_ref',
  'timeout_ref',
  'config_serving_receipt_ref',
  'checksum_registry_ref',
  'cluster_ref',
  'namespace_ref',
  'workload_ref',
  'config_map_ref',
  'rollout_strategy_ref',
  'rollout_receipt_ref',
  'data_arrival_ref',
  'data_arrival_validator_ref',
  'executor_ref',
  'target_os',
  'transport',
  'runner',
  'idempotency_key',
  'timeout_policy_ref',
  'execution_receipt_ref',
  'audit_ref',
  'workload_selector_ref',
  'helm_chart_ref',
  'manifest_bundle_ref',
]
const ROLLOUT_REF_KEYS = new Set(REMOTE_MUTATION_REF_KEYS)
const REMOTE_MUTATION_SCOPES = ['FindX Agent', 'CMDB 主机', '业务组', 'namespace', 'workload']

const refNameFrom = item => {
  if (typeof item === 'string') return item
  if (!item || typeof item !== 'object') return ''
  return item.name || item.key || item.id || ''
}

const isVisibleRef = name => {
  const value = String(name || '').trim()
  return (value.startsWith('provider_') && value.endsWith('_ref')) || ROLLOUT_REF_KEYS.has(value)
}

export function HostProviderRefs({ rolloutMetadata }) {
  const refs = Array.from(new Set([...REMOTE_MUTATION_REF_KEYS, ...(rolloutMetadata || []).map(refNameFrom)].filter(isVisibleRef)))
  if (!refs.length) return null

  return (
    <div className='fx-agent-provider-refs' aria-label='配置来源与下发回执引用'>
      <div className='fx-agent-subtitle'>远程修改作用域</div>
      <ul>{REMOTE_MUTATION_SCOPES.map(name => <li key={name}><code>{name}</code><span>{CONTRACT_STATUS}</span></li>)}</ul>
      <div className='fx-agent-subtitle'>writer / reload / drift / rollback / Evidence refs</div>
      <ul>
        {refs.map(name => (
          <li key={name}>
            <code>{name}</code>
            <span>{CONTRACT_STATUS}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}
