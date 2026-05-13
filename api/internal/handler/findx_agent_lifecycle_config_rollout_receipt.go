package handler

import (
	"strings"

	"ai-workbench-api/internal/model"
)

func configRolloutReceiptContract(req model.FindXAgentConfigRolloutRequest, metadata map[string]string, credentialProvided bool, missing []string) model.FindXAgentReceiptContract {
	return model.FindXAgentReceiptContract{
		ID:                 "categraf_plugin_config_rollout_receipt_contract",
		Scope:              configRolloutReceiptScope(req, metadata),
		Transport:          configRolloutReceiptTransport(metadata),
		Runner:             configRolloutReceiptRunner(metadata),
		RequiredReceipts:   []string{"writer_receipt", "reload_receipt", "restart_receipt", "drift_receipt", "rollback_receipt", "data_arrival_receipt", "evidence_chain"},
		MissingContracts:   configRolloutReceiptMissingContracts(missing),
		CredentialRequired: true,
		CredentialProvided: credentialProvided,
		Status:             model.FindXAgentExecutionStateBlockedByContract,
		Blocker:            agentBlocked + ": config rollout executor and receipt protocol are not open",
	}
}

func configRolloutReceiptMissingContracts(missing []string) []string {
	if len(missing) == 0 {
		return []string{"executor_disabled_contract"}
	}
	return uniquePackageRepositoryBlockers(missing)
}

func configRolloutReceiptScope(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	if isPluginConfigRollout(req) {
		if scope := configRolloutAllowedScope(metadata["scope"]); scope != "" {
			return "categraf_plugin_config_rollout_" + scope
		}
		return "categraf_plugin_config_rollout"
	}
	return "config_rollout"
}

func configRolloutAllowedScope(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case configRolloutScopeAgent:
		return configRolloutScopeAgent
	case configRolloutScopeCMDBHost:
		return configRolloutScopeCMDBHost
	case configRolloutScopeBusinessGroup:
		return configRolloutScopeBusinessGroup
	case configRolloutScopeNamespace:
		return configRolloutScopeNamespace
	case configRolloutScopeWorkload:
		return configRolloutScopeWorkload
	default:
		return ""
	}
}

func configRolloutReceiptTransport(metadata map[string]string) string {
	if transport := normalizeInstallPlanTransport(metadata["transport"]); transport != "" {
		return transport
	}
	if runner := normalizeInstallPlanTransport(metadata["runner"]); runner != "" {
		return runner
	}
	if isKubernetesAgentTask(metadata) {
		return "kubernetes"
	}
	return "local"
}

func configRolloutReceiptRunner(metadata map[string]string) string {
	if runner := normalizeInstallPlanTransport(metadata["runner"]); runner != "" {
		return runner
	}
	switch configRolloutReceiptTransport(metadata) {
	case "ssh":
		return "ssh"
	case "winrm":
		return "winrm"
	case "helm":
		return "helm"
	case "operator":
		return "operator"
	default:
		return ""
	}
}

func configRolloutReceiptContractMatrix() []model.FindXAgentReceiptContractMatrixRow {
	missing := []string{
		"config_writer_receipt_contract",
		"reload_receipt_contract",
		"restart_receipt_contract",
		"drift_receipt_contract",
		"rollback_receipt_contract",
		"data_arrival_receipt_contract",
		"evidence_chain_contract",
	}
	return []model.FindXAgentReceiptContractMatrixRow{
		configRolloutReceiptMatrixRow("writer", "all", "Categraf plugin config writer", missing),
		configRolloutReceiptMatrixRow("reload", "linux/windows/kubernetes", "plugin reload or rollout reload", missing),
		configRolloutReceiptMatrixRow("restart", "linux/windows/kubernetes", "service or workload restart", missing),
		configRolloutReceiptMatrixRow("drift", "all", "post-rollout drift detection", missing),
		configRolloutReceiptMatrixRow("rollback", "all", "rollback receipt and restore evidence", missing),
		configRolloutReceiptMatrixRow("data_arrival", "metrics/logs/traces/profiling/inspection", "signal data-arrival validation", missing),
		configRolloutReceiptMatrixRow("evidence_chain", "all", "audit and evidence chain linkage", missing),
	}
}

func configRolloutReceiptMatrixRow(scope, platform, surface string, missing []string) model.FindXAgentReceiptContractMatrixRow {
	return model.FindXAgentReceiptContractMatrixRow{
		Scope:             scope,
		Platform:          platform,
		ExecutionSurface:  surface,
		RequiredContracts: []string{"writer", "reload", "restart", "drift", "rollback", "data_arrival", "evidence_chain"},
		MissingContracts:  uniquePackageRepositoryBlockers(missing),
		Status:            model.FindXAgentExecutionStateBlockedByContract,
		Blocker:           agentBlocked + ": config rollout " + scope + " receipt contract is not open",
	}
}
