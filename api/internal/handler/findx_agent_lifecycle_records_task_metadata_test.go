package handler

func completeRemotePreflightMetadata() string {
	return `"transport":"ssh","target_os":"linux","idempotency_key":"idem-1","timeout_policy_ref":"timeout-1","execution_receipt_ref":"receipt-1","audit_ref":"audit-1","data_arrival_validator_ref":"validator-1","ssh_runner_ref":"ssh-runner-1","ssh_host_key":"host-key-1","ssh_fingerprint":"fingerprint-1","remote_executor_ref":"remote-executor-1"`
}

func completeKubernetesTaskMetadata() string {
	return `"cluster_ref":"cluster-1","namespace_ref":"namespace-1","workload_selector_ref":"workload-1","rbac_ref":"rbac-1","service_account_ref":"sa-1","rollout_strategy_ref":"rollout-strategy-1","rollout_receipt_ref":"rollout-receipt-1","data_arrival_validator_ref":"validator-1","daemonset_ref":"daemonset-1"`
}
