export const pluginConfigSpec = (pluginId, reloadStrategy) => ({
  plugin_id: pluginId,
  plugin_version: '<PLUGIN_VERSION>',
  config_format: 'toml',
  config_snippet_ref: '<CONFIG_SNIPPET_REF>',
  provider_modes: ['local', 'http'],
  reload_strategy: reloadStrategy,
  restart_strategy: 'restart-if-plugin-requires',
  remote_mutation: true,
  remote_mutation_status: 'PENDING',
  rollout_metadata: ['config_version', 'canary_percent', 'rollback_ref', 'change_ticket', 'audit_reason', 'plugin_config_writer_ref', 'reload_command_ref', 'reload_receipt_ref', 'drift_check_ref', 'evidence_chain_ref', 'rollback_receipt_ref', 'provider_endpoint_ref', 'provider_response_version_ref', 'provider_checksum_ref', 'provider_headers_ref', 'provider_auth_ref', 'provider_tls_ref', 'provider_scope_ref', 'provider_mutation_receipt_ref', 'provider_audit_ref', 'provider_record_detail_ref', 'reload_interval_ref', 'timeout_ref', 'config_serving_receipt_ref', 'checksum_registry_ref', 'cluster_ref', 'namespace_ref', 'workload_ref', 'workload_selector_ref', 'config_map_ref', 'rollout_strategy_ref', 'rollout_receipt_ref', 'data_arrival_ref', 'data_arrival_validator_ref', 'executor_ref', 'target_os', 'transport', 'runner', 'idempotency_key', 'timeout_policy_ref', 'execution_receipt_ref', 'audit_ref', 'helm_chart_ref', 'manifest_bundle_ref'],
  credential_ref_required: true,
  audit_event: 'findx_agent.plugin_config.remote_mutation.requested',
})
