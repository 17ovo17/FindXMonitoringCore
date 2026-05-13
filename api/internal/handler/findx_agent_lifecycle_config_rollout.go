package handler

import (
	"fmt"
	"sort"
	"strings"

	"ai-workbench-api/internal/model"
)

const (
	configRolloutTemplateHostPlugin      = "host-plugin"
	configRolloutTemplateContainerPlugin = "container-plugin"
	configRolloutTemplateGatewayPlugin   = "gateway-plugin"
)

const (
	configRolloutScopeAgent         = "agent"
	configRolloutScopeCMDBHost      = "cmdb_host"
	configRolloutScopeBusinessGroup = "business_group"
	configRolloutScopeNamespace     = "namespace"
	configRolloutScopeWorkload      = "workload"
)

const (
	configRolloutStrategyCanary   = "canary"
	configRolloutStrategyFull     = "full"
	configRolloutStrategyRollback = "rollback"
	configRolloutStrategyDrift    = "drift"
	configRolloutStrategyReload   = "reload"
	configRolloutStrategyEvidence = "evidence"
)

const (
	configRolloutRefAgent            = "agent_ref"
	configRolloutRefBusinessGroup    = "business_group_ref"
	configRolloutRefCMDBHost         = "cmdb_host_ref"
	configRolloutRefDriftCheck       = "drift_check_ref"
	configRolloutRefEvidenceChain    = "evidence_chain_ref"
	configRolloutRefNamespace        = "namespace_ref"
	configRolloutRefReloadReceipt    = "reload_receipt_ref"
	configRolloutRefRolloutReceipt   = "rollout_receipt_ref"
	configRolloutRefRollbackReceipt  = "rollback_receipt_ref"
	configRolloutRefRolloutStrategy  = "rollout_strategy_ref"
	configRolloutRefWorkloadSelector = "workload_selector_ref"
)

func newBlockedFindXAgentConfigRollout(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) model.FindXAgentConfigRollout {
	return model.FindXAgentConfigRollout{
		TemplateID:           strings.TrimSpace(req.TemplateID),
		AgentIDs:             cleanAgentLifecycleValues(req.AgentIDs),
		TargetIDs:            cleanAgentLifecycleValues(req.TargetIDs),
		ConfigVersion:        sanitizeRemoteMutationValue("config_version", req.ConfigVersion),
		ConfigSnippetRef:     sanitizeRemoteMutationValue("config_snippet_ref", req.ConfigSnippetRef),
		ConfigFormat:         sanitizeRemoteMutationValue("config_format", req.ConfigFormat),
		ProviderMode:         sanitizeRemoteMutationValue("provider_mode", req.ProviderMode),
		PluginID:             sanitizeRemoteMutationValue("plugin_id", req.PluginID),
		PluginVersion:        sanitizeRemoteMutationValue("plugin_version", req.PluginVersion),
		ReloadStrategy:       sanitizeRemoteMutationValue("reload_strategy", req.ReloadStrategy),
		RestartStrategy:      sanitizeRemoteMutationValue("restart_strategy", req.RestartStrategy),
		RolloutStrategy:      sanitizeRemoteMutationValue("rollout_strategy", req.RolloutStrategy),
		RollbackRef:          sanitizeRemoteMutationValue("rollback_ref", req.RollbackRef),
		AuditReason:          sanitizeRemoteMutationValue("audit_reason", req.AuditReason),
		ChangeTicket:         sanitizeRemoteMutationValue("change_ticket", req.ChangeTicket),
		RemoteMutation:       req.RemoteMutation,
		CanaryPercent:        req.CanaryPercent,
		Status:               "blocked",
		Blocker:              configRolloutBlocker(req, metadata),
		Audit:                "findx_agent.config_rollout.requested",
		CredentialRefPresent: strings.TrimSpace(req.CredentialRef) != "",
		Metadata:             metadata,
	}
}

func configRolloutBlocker(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	missing := missingConfigRolloutRefs(req, metadata)
	if len(missing) > 0 {
		return fmt.Sprintf("%s: missing %s", agentBlocked, strings.Join(missing, ", "))
	}
	return agentBlocked + ": executor not enabled / config rollout protocol not open"
}

func configRolloutResponseBlockers(missing []string) []string {
	if len(missing) == 0 {
		return []string{agentBlocked, "EXECUTOR_DISABLED_BY_CONTRACT"}
	}
	values := append([]string{agentBlocked, "MISSING_CONTRACTS"}, missing...)
	if containsConfigRolloutRef(missing, "unsafe_plugin_policy_ref") {
		values = append(values, "UNSAFE_PLUGIN_BLOCKED_BY_CONTRACT")
	}
	return uniquePackageRepositoryBlockers(values)
}

func missingConfigRolloutRefs(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) []string {
	required := append(requiredConfigRolloutRefs(req), requiredRemoteExecutionRefs()...)
	required = append(required, requiredConfigRolloutScopeRefs(metadata)...)
	required = append(required, requiredConfigRolloutStrategyRefs(req)...)
	if isKubernetesAgentTask(metadata) {
		required = append(required, requiredKubernetesConfigRolloutRefs(metadata)...)
	}
	missingSet := map[string]bool{}
	values := sanitizedConfigRolloutRefs(req, metadata)
	for _, key := range required {
		if strings.TrimSpace(values[key]) == "" {
			missingSet[key] = true
		}
	}
	if strings.TrimSpace(req.CredentialRef) == "" {
		missingSet["credential_ref"] = true
	}
	for _, key := range missingRemoteExecutionChoiceRefs(metadata) {
		missingSet[key] = true
	}
	if isKubernetesAgentTask(metadata) {
		for _, key := range missingKubernetesConfigRolloutChoiceRefs(values) {
			missingSet[key] = true
		}
	}
	if isUnsafeCategrafPlugin(req.PluginID) {
		missingSet["unsafe_plugin_policy_ref"] = true
	}
	if requiresWindowsServiceRestartRefs(values) {
		missingSet["data_arrival_validator_ref"] = true
		missingSet["restart_strategy_ref"] = true
		missingSet["service_restart_receipt_ref"] = true
	}
	missing := make([]string, 0, len(missingSet))
	for key := range missingSet {
		missing = append(missing, key)
	}
	sort.Strings(missing)
	return missing
}

func requiredConfigRolloutRefs(req model.FindXAgentConfigRolloutRequest) []string {
	required := []string{"config_snippet_ref", "config_version", "executor_ref", "rollback_ref"}
	if isPluginConfigRollout(req) {
		required = append(required,
			"config_format",
			"drift_check_ref",
			"evidence_chain_ref",
			"data_arrival_validator_ref",
			"plugin_config_writer_ref",
			"plugin_id",
			"provider_mode",
			"reload_command_ref",
			"reload_receipt_ref",
			"reload_strategy",
			"rollback_receipt_ref",
		)
		if isHTTPProviderConfigRollout(req) {
			required = append(required, requiredHTTPProviderConfigRolloutRefs()...)
		}
	}
	return required
}

func requiredConfigRolloutScopeRefs(metadata map[string]string) []string {
	switch strings.ToLower(strings.TrimSpace(metadata["scope"])) {
	case configRolloutScopeAgent:
		return []string{configRolloutRefAgent}
	case configRolloutScopeCMDBHost:
		return []string{configRolloutRefCMDBHost}
	case configRolloutScopeBusinessGroup:
		return []string{configRolloutRefBusinessGroup}
	case configRolloutScopeNamespace:
		return []string{configRolloutRefNamespace}
	case configRolloutScopeWorkload:
		return []string{
			configRolloutRefNamespace,
			configRolloutRefWorkloadSelector,
		}
	default:
		return nil
	}
}

func requiredConfigRolloutStrategyRefs(req model.FindXAgentConfigRolloutRequest) []string {
	switch strings.ToLower(strings.TrimSpace(req.RolloutStrategy)) {
	case configRolloutStrategyCanary:
		return []string{configRolloutRefRolloutReceipt}
	case configRolloutStrategyFull:
		return []string{configRolloutRefRolloutStrategy, configRolloutRefRolloutReceipt}
	case configRolloutStrategyRollback:
		return []string{configRolloutRefRollbackReceipt}
	case configRolloutStrategyDrift:
		return []string{configRolloutRefDriftCheck}
	case configRolloutStrategyReload:
		return []string{configRolloutRefReloadReceipt}
	case configRolloutStrategyEvidence:
		return []string{configRolloutRefEvidenceChain}
	default:
		return nil
	}
}

func requiredHTTPProviderConfigRolloutRefs() []string {
	return []string{
		"provider_endpoint_ref",
		"provider_response_version_ref",
		"provider_checksum_ref",
		"checksum_registry_ref",
		"provider_headers_ref",
		"provider_auth_ref",
		"provider_tls_ref",
		"reload_interval_ref",
		"timeout_ref",
		"config_serving_receipt_ref",
	}
}

func requiredKubernetesConfigRolloutRefs(metadata map[string]string) []string {
	required := []string{
		"cluster_ref",
		"namespace_ref",
		"workload_selector_ref",
		"config_map_ref",
		"rollout_strategy_ref",
		"rollout_receipt_ref",
		"reload_receipt_ref",
		"drift_check_ref",
		"data_arrival_validator_ref",
		"executor_ref",
	}
	if isHelmAgentTask(metadata) {
		required = append(required, "helm_release_ref")
	}
	return required
}

func missingKubernetesConfigRolloutChoiceRefs(values map[string]string) []string {
	if strings.TrimSpace(values["helm_chart_ref"]) != "" || strings.TrimSpace(values["manifest_bundle_ref"]) != "" {
		return nil
	}
	return []string{"helm_chart_ref_or_manifest_bundle_ref"}
}

func sanitizedConfigRolloutRefs(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) map[string]string {
	return map[string]string{
		"cluster_ref":                   strings.TrimSpace(metadata["cluster_ref"]),
		"config_format":                 sanitizeRemoteMutationValue("config_format", req.ConfigFormat),
		"config_map_ref":                strings.TrimSpace(metadata["config_map_ref"]),
		"config_snippet_ref":            sanitizeRemoteMutationValue("config_snippet_ref", req.ConfigSnippetRef),
		"config_version":                sanitizeRemoteMutationValue("config_version", req.ConfigVersion),
		"agent_ref":                     strings.TrimSpace(metadata["agent_ref"]),
		"business_group_ref":            strings.TrimSpace(metadata["business_group_ref"]),
		"cmdb_host_ref":                 strings.TrimSpace(metadata["cmdb_host_ref"]),
		"data_arrival_validator_ref":    strings.TrimSpace(metadata["data_arrival_validator_ref"]),
		"drift_check_ref":               strings.TrimSpace(metadata["drift_check_ref"]),
		"evidence_chain_ref":            strings.TrimSpace(metadata["evidence_chain_ref"]),
		"executor_ref":                  strings.TrimSpace(metadata["executor_ref"]),
		"helm_chart_ref":                strings.TrimSpace(metadata["helm_chart_ref"]),
		"helm_release_ref":              strings.TrimSpace(metadata["helm_release_ref"]),
		"idempotency_key":               strings.TrimSpace(metadata["idempotency_key"]),
		"manifest_bundle_ref":           strings.TrimSpace(metadata["manifest_bundle_ref"]),
		"namespace_ref":                 strings.TrimSpace(metadata["namespace_ref"]),
		"plugin_config_writer_ref":      strings.TrimSpace(metadata["plugin_config_writer_ref"]),
		"plugin_id":                     sanitizeRemoteMutationValue("plugin_id", req.PluginID),
		"provider_auth_ref":             strings.TrimSpace(metadata["provider_auth_ref"]),
		"provider_checksum_ref":         strings.TrimSpace(metadata["provider_checksum_ref"]),
		"checksum_registry_ref":         strings.TrimSpace(metadata["checksum_registry_ref"]),
		"provider_endpoint_ref":         strings.TrimSpace(metadata["provider_endpoint_ref"]),
		"provider_headers_ref":          strings.TrimSpace(metadata["provider_headers_ref"]),
		"provider_mode":                 sanitizeRemoteMutationValue("provider_mode", req.ProviderMode),
		"provider_response_version_ref": strings.TrimSpace(metadata["provider_response_version_ref"]),
		"provider_tls_ref":              strings.TrimSpace(metadata["provider_tls_ref"]),
		"config_serving_receipt_ref":    strings.TrimSpace(metadata["config_serving_receipt_ref"]),
		"reload_command_ref":            strings.TrimSpace(metadata["reload_command_ref"]),
		"reload_interval_ref":           strings.TrimSpace(metadata["reload_interval_ref"]),
		"reload_receipt_ref":            strings.TrimSpace(metadata["reload_receipt_ref"]),
		"reload_strategy":               sanitizeRemoteMutationValue("reload_strategy", req.ReloadStrategy),
		"restart_strategy_ref":          strings.TrimSpace(metadata["restart_strategy_ref"]),
		"rollback_ref":                  sanitizeRemoteMutationValue("rollback_ref", req.RollbackRef),
		"rollback_receipt_ref":          strings.TrimSpace(metadata["rollback_receipt_ref"]),
		"rollout_receipt_ref":           strings.TrimSpace(metadata["rollout_receipt_ref"]),
		"rollout_strategy_ref":          strings.TrimSpace(metadata["rollout_strategy_ref"]),
		"service_restart_receipt_ref":   strings.TrimSpace(metadata["service_restart_receipt_ref"]),
		"target_os":                     strings.TrimSpace(metadata["target_os"]),
		"timeout_ref":                   strings.TrimSpace(metadata["timeout_ref"]),
		"timeout_policy_ref":            strings.TrimSpace(metadata["timeout_policy_ref"]),
		"transport":                     normalizeInstallPlanTransport(metadata["transport"]),
		"workload_selector_ref":         strings.TrimSpace(metadata["workload_selector_ref"]),
	}
}

func hasConfigRolloutTarget(req model.FindXAgentConfigRolloutRequest) bool {
	return len(cleanAgentLifecycleValues(req.AgentIDs)) > 0 || len(cleanAgentLifecycleValues(req.TargetIDs)) > 0
}

func isPluginConfigRollout(req model.FindXAgentConfigRolloutRequest) bool {
	return req.RemoteMutation ||
		isCategrafPluginTemplate(req.TemplateID) ||
		strings.TrimSpace(req.PluginID) != ""
}

func isCategrafPluginTemplate(templateID string) bool {
	switch strings.TrimSpace(templateID) {
	case configRolloutTemplateHostPlugin,
		configRolloutTemplateContainerPlugin,
		configRolloutTemplateGatewayPlugin:
		return true
	default:
		return false
	}
}

func isHTTPProviderConfigRollout(req model.FindXAgentConfigRolloutRequest) bool {
	return strings.EqualFold(strings.TrimSpace(req.ProviderMode), "http")
}

func containsConfigRolloutRef(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func isUnsafeCategrafPlugin(pluginID string) bool {
	clean := strings.ToLower(strings.TrimSpace(pluginID))
	clean = strings.TrimPrefix(clean, "inputs.")
	clean = strings.TrimPrefix(clean, "input.")
	return clean == "exec"
}

func requiresWindowsServiceRestartRefs(values map[string]string) bool {
	targetOS := strings.ToLower(strings.TrimSpace(values["target_os"]))
	transport := strings.ToLower(strings.TrimSpace(values["transport"]))
	reloadStrategy := strings.ToLower(strings.TrimSpace(values["reload_strategy"]))
	if targetOS != "windows" && !strings.Contains(transport, "windows_service") {
		return false
	}
	return reloadStrategy == "hup" || strings.Contains(transport, "windows_service")
}
