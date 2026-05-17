package handler

import (
	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"sort"
	"strings"
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
		Status:               "accepted",
		Blocker:              "",
		Audit:                "findx_agent.config_rollout.created",
		CredentialRefPresent: strings.TrimSpace(req.CredentialRef) != "",
		Metadata:             metadata,
	}
}

func configRolloutBlocker(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	missing := missingConfigRolloutRefs(req, metadata)
	return configRolloutBlockerFromMissing(missing)
}

func configRolloutBlockerFromMissing(missing []string) string {
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
		values = append(values, "UNSAFE_PLUGIN_PENDING")
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
	if isUnsafeFindXPlugin(req.PluginID) {
		missingSet["unsafe_plugin_policy_ref"] = true
	}
	if isCMDBHostPluginRollout(req, metadata) && configRolloutOperationIdentityConflict(req, metadata) {
		missingSet[configRolloutPluginOperationConflict] = true
	}
	if requiresWindowsServiceRestartRefs(values) {
		missingSet["data_arrival_validator_ref"] = true
		missingSet["restart_strategy_ref"] = true
		missingSet["service_restart_receipt_ref"] = true
	}
	if isCMDBHostPluginRollout(req, metadata) {
		for _, key := range cmdbHostPluginRolloutRuntimeContracts(req, metadata) {
			missingSet[key] = true
		}
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
		isFindXPluginTemplate(req.TemplateID) ||
		strings.TrimSpace(req.PluginID) != ""
}

func isCMDBHostPluginRollout(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) bool {
	template := strings.ToLower(strings.TrimSpace(req.TemplateID))
	provider := strings.ToLower(strings.TrimSpace(req.ProviderMode))
	return strings.HasPrefix(template, "cmdb-host-plugin") ||
		provider == "cmdb_host_probe" ||
		(strings.EqualFold(strings.TrimSpace(metadata["scope"]), configRolloutScopeCMDBHost) && strings.Contains(template, "cmdb"))
}

func cmdbHostPluginRolloutRuntimeContracts(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) []string {
	return cmdbHostPluginOperationMissingContracts(req, configRolloutOperationMode(req, metadata))
}

func isFindXPluginTemplate(templateID string) bool {
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

func isUnsafeFindXPlugin(pluginID string) bool {
	clean := strings.ToLower(strings.TrimSpace(pluginID))
	clean = strings.TrimPrefix(clean, "inputs.")
	clean = strings.TrimPrefix(clean, "input.")
	switch clean {
	case "exec", "smart", "ipmi":
		return true
	default:
		return false
	}
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


const (
	cmdbAgentPluginAssignmentStoreContract = "cmdb_agent_plugin_assignment_store_contract"
	cmdbAgentPluginTargetBindingContract   = "cmdb_agent_plugin_target_binding_contract"
	cmdbAgentPluginAssignmentAuditContract = "cmdb_agent_plugin_assignment_audit_contract"
	cmdbAgentPluginAssignmentRefContract   = "cmdb_agent_plugin_assignment_ref_contract"
)

type cmdbPluginAssignmentContext struct {
	Assignment      model.FindXAgentPluginAssignment
	TargetBinding   model.FindXAgentPluginTargetBinding
	AssignmentReady bool
	BindingReady    bool
	AuditReady      bool
}

func cmdbResolvePluginAssignmentContext(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) cmdbPluginAssignmentContext {
	if !isCMDBHostPluginRollout(req, metadata) {
		return cmdbPluginAssignmentContext{}
	}
	if configRolloutOperationMode(req, metadata) != configRolloutPluginOperationDispatch {
		return cmdbPluginAssignmentContext{}
	}
	assignment, ok := cmdbFindPluginAssignment(req, metadata)
	if !ok {
		return cmdbPluginAssignmentContext{}
	}
	binding, bindingOK := cmdbFindPluginTargetBinding(assignment)
	return cmdbPluginAssignmentContext{
		Assignment:      assignment,
		TargetBinding:   binding,
		AssignmentReady: true,
		BindingReady:    bindingOK,
		AuditReady:      strings.TrimSpace(assignment.AuditRef) != "",
	}
}

func cmdbPersistPluginAssignment(c *gin.Context, req model.FindXAgentConfigRolloutRequest, metadata map[string]string, saved model.FindXAgentConfigRollout, missing []string) (cmdbPluginAssignmentContext, error) {
	if !isCMDBHostPluginRollout(req, metadata) || configRolloutOperationMode(req, metadata) != configRolloutPluginOperationAssign {
		return cmdbPluginAssignmentContext{}, nil
	}
	hostRef := cmdbPluginAssignmentHostRef(req, metadata)
	agentRef := cmdbPluginAssignmentAgentRef(req, metadata)
	pluginID := sanitizeRemoteMutationValue("plugin_id", req.PluginID)
	if hostRef == "" || agentRef == "" || pluginID == "" {
		return cmdbPluginAssignmentContext{}, nil
	}
	auditRef := "cmdb-agent-plugin-assignment-audit-" + store.NewID()
	assignment := model.FindXAgentPluginAssignment{
		SourceRolloutID:      saved.ID,
		HostRef:              hostRef,
		AgentRef:             agentRef,
		PluginID:             pluginID,
		PluginVersion:        sanitizeRemoteMutationValue("plugin_version", req.PluginVersion),
		ConfigSnippetRef:     sanitizeRemoteMutationValue("config_snippet_ref", req.ConfigSnippetRef),
		ConfigFormat:         sanitizeRemoteMutationValue("config_format", req.ConfigFormat),
		ProviderMode:         sanitizeRemoteMutationValue("provider_mode", req.ProviderMode),
		AuditRef:             auditRef,
		AssignmentContract:   "cmdb.agent.plugin.assignment.v1",
		Status:               "accepted",
		Blocker:              "",
		CredentialRefPresent: strings.TrimSpace(req.CredentialRef) != "",
		DashboardRefsJSON:    lifecycleJSONString(cmdbPluginAssignmentDashboardRefs(metadata)),
		DashboardRefsCount:   len(cmdbPluginAssignmentDashboardRefs(metadata)),
		MissingContractsJSON: lifecycleJSONString(cmdbFilterMissingContractsForAssignment(missing, true)),
		MetadataJSON:         lifecycleJSONString(metadata),
	}
	savedAssignment, err := store.SaveFindXAgentPluginAssignment(assignment)
	if err != nil {
		return cmdbPluginAssignmentContext{}, err
	}
	binding := model.FindXAgentPluginTargetBinding{
		AssignmentID:         savedAssignment.ID,
		SourceRolloutID:      saved.ID,
		HostRef:              hostRef,
		TargetID:             cmdbPluginAssignmentTargetID(req, metadata),
		AgentRef:             agentRef,
		PluginID:             pluginID,
		BindingType:          "cmdb_host_plugin",
		Status:               "accepted",
		Blocker:              "",
		CredentialRefPresent: strings.TrimSpace(req.CredentialRef) != "",
		DashboardRefsCount:   len(cmdbPluginAssignmentDashboardRefs(metadata)),
		ContractID:           "cmdb.agent.plugin.target_binding.v1",
		AuditRef:             auditRef,
	}
	savedBinding, err := store.SaveFindXAgentPluginTargetBinding(binding)
	if err != nil {
		return cmdbPluginAssignmentContext{}, err
	}
	savedAssignment.TargetBindingRef = savedBinding.ID
	savedAssignment, err = store.SaveFindXAgentPluginAssignment(savedAssignment)
	if err != nil {
		return cmdbPluginAssignmentContext{}, err
	}
	if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
		ID:           auditRef,
		Actor:        requestActor(c),
		Action:       "cmdb.agent.plugin.assignment.save",
		ResourceType: "cmdb_agent_plugin_assignment",
		ResourceID:   savedAssignment.ID,
		Scope:        "cmdb",
		Status:       "accepted",
		TraceID:      c.GetHeader("X-Test-Batch-Id"),
		ClientIP:     c.ClientIP(),
		Summary:      "CMDB plugin assignment saved",
		Details: map[string]any{
			"assignment_id":          savedAssignment.ID,
			"target_binding_id":      savedBinding.ID,
			"host_ref":               hostRef,
			"agent_ref":              agentRef,
			"plugin_id":              pluginID,
			"credential_ref_present": savedAssignment.CredentialRefPresent,
			"dashboard_refs_count":   savedAssignment.DashboardRefsCount,
			"source_rollout_id":      saved.ID,
		},
	}); err != nil {
		return cmdbPluginAssignmentContext{}, err
	}
	return cmdbPluginAssignmentContext{
		Assignment:      savedAssignment,
		TargetBinding:   savedBinding,
		AssignmentReady: true,
		BindingReady:    true,
		AuditReady:      true,
	}, nil
}

func cmdbFindPluginAssignment(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) (model.FindXAgentPluginAssignment, bool) {
	if ref := strings.TrimSpace(metadata["assignment_ref"]); ref != "" {
		assignment, ok, err := store.GetFindXAgentPluginAssignment(ref)
		if err != nil || !ok || !cmdbPluginAssignmentMatchesRequest(assignment, req, metadata) {
			return model.FindXAgentPluginAssignment{}, false
		}
		return assignment, true
	}
	assignment, ok, err := store.FindFindXAgentPluginAssignment(
		cmdbPluginAssignmentHostRef(req, metadata),
		cmdbPluginAssignmentAgentRef(req, metadata),
		sanitizeRemoteMutationValue("plugin_id", req.PluginID),
	)
	if err != nil || !ok {
		return model.FindXAgentPluginAssignment{}, false
	}
	return assignment, true
}

func cmdbFindPluginTargetBinding(assignment model.FindXAgentPluginAssignment) (model.FindXAgentPluginTargetBinding, bool) {
	if strings.TrimSpace(assignment.TargetBindingRef) != "" {
		binding, ok, err := store.GetFindXAgentPluginTargetBinding(assignment.TargetBindingRef)
		if err == nil && ok {
			return binding, true
		}
	}
	bindings, err := store.ListFindXAgentPluginTargetBindings(assignment.ID)
	if err != nil || len(bindings) == 0 {
		return model.FindXAgentPluginTargetBinding{}, false
	}
	return bindings[0], true
}

func cmdbPluginAssignmentMatchesRequest(assignment model.FindXAgentPluginAssignment, req model.FindXAgentConfigRolloutRequest, metadata map[string]string) bool {
	return assignment.HostRef == cmdbPluginAssignmentHostRef(req, metadata) &&
		assignment.AgentRef == cmdbPluginAssignmentAgentRef(req, metadata) &&
		assignment.PluginID == sanitizeRemoteMutationValue("plugin_id", req.PluginID)
}

func cmdbPluginAssignmentHostRef(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	for _, value := range []string{metadata["cmdb_host_ref"], metadata["cmdb_host_id"]} {
		if clean := sanitizeRemoteMutationValue("cmdb_host_ref", value); clean != "" {
			return clean
		}
	}
	targets := cleanAgentLifecycleValues(req.TargetIDs)
	if len(targets) > 0 {
		return sanitizeRemoteMutationValue("target_id", targets[0])
	}
	return ""
}

func cmdbPluginAssignmentAgentRef(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	if clean := sanitizeRemoteMutationValue("agent_ref", metadata["agent_ref"]); clean != "" {
		return clean
	}
	agents := cleanAgentLifecycleValues(req.AgentIDs)
	if len(agents) > 0 {
		return sanitizeRemoteMutationValue("agent_id", agents[0])
	}
	return ""
}

func cmdbPluginAssignmentTargetID(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	if clean := sanitizeRemoteMutationValue("target_id", metadata["target_id"]); clean != "" {
		return clean
	}
	return cmdbPluginAssignmentHostRef(req, metadata)
}

func cmdbPluginAssignmentDashboardRefs(metadata map[string]string) []string {
	raw := strings.TrimSpace(metadata["dashboard_refs"])
	if raw == "" {
		return nil
	}
	values := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '/'
	})
	out := make([]string, 0, len(values))
	for _, value := range values {
		if clean := sanitizeRemoteMutationValue("dashboard_ref", value); clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

func cmdbFilterMissingContractsForAssignment(missing []string, removeAssignmentContracts bool) []string {
	out := make([]string, 0, len(missing))
	for _, value := range missing {
		if removeAssignmentContracts && (value == cmdbAgentPluginAssignmentStoreContract ||
			value == cmdbAgentPluginTargetBindingContract ||
			value == cmdbAgentPluginAssignmentAuditContract) {
			continue
		}
		out = append(out, value)
	}
	return uniquePackageRepositoryBlockers(out)
}

func cmdbFilterMissingContractsForDispatch(missing []string, ctx cmdbPluginAssignmentContext) []string {
	out := make([]string, 0, len(missing))
	for _, value := range missing {
		if ctx.AssignmentReady && value == cmdbAgentPluginAssignmentRefContract {
			continue
		}
		out = append(out, value)
	}
	return uniquePackageRepositoryBlockers(out)
}

func cmdbPluginAssignmentResponse(ctx cmdbPluginAssignmentContext) gin.H {
	if !ctx.AssignmentReady {
		return nil
	}
	return gin.H{
		"assignment_ref":          ctx.Assignment.ID,
		"target_binding_ref":      ctx.TargetBinding.ID,
		"contract_id":             "cmdb.agent.plugin.assignment.store.read.v1",
		"status":                  "accepted",
		"credential_ref_present":  ctx.Assignment.CredentialRefPresent,
		"dashboard_refs_count":    ctx.Assignment.DashboardRefsCount,
		"audit_ref":               ctx.Assignment.AuditRef,
		"source_rollout_id":       ctx.Assignment.SourceRolloutID,
		"remote_mutation":         false,
		"dispatch_receipts_state": "accepted",
		"findx_audit_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_agent_plugin_assignment",
			"resource_id":   ctx.Assignment.ID,
			"action":        "cmdb.agent.plugin.assignment.save",
		},
	}
}

func cmdbOperationContractWithAssignment(base gin.H, ctx cmdbPluginAssignmentContext) gin.H {
	if !ctx.AssignmentReady {
		return base
	}
	base["assignment_ref"] = ctx.Assignment.ID
	base["target_binding_ref"] = ctx.TargetBinding.ID
	base["assignment_contract"] = cmdbPluginAssignmentResponse(ctx)
	return base
}

func cmdbOperationContractWithAssignmentAndCredential(base gin.H, assignmentCtx cmdbPluginAssignmentContext, credentialCtx cmdbPluginCredentialContext) gin.H {
	base = cmdbOperationContractWithAssignment(base, assignmentCtx)
	if credentialContract := cmdbPluginCredentialContractResponse(credentialCtx); credentialContract != nil {
		base["credential_contract"] = credentialContract
		if credentialCtx.Resolved {
			base["credential_ref_resolved"] = true
		}
		if credentialCtx.ScopePolicyResolved && credentialCtx.ScopeAuthorized {
			base["scope_policy_resolved"] = true
			base["scope_authorized"] = true
			base["policy_approval_ref"] = credentialCtx.PolicyApprovalRef
		}
	}
	return base
}

func lifecycleJSONString(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}


const (
	configRolloutPluginOperationAssign   = "assign"
	configRolloutPluginOperationDispatch = "dispatch"
	configRolloutPluginOperationConflict = "cmdb_agent_plugin_operation_identity_contract"
)

func configRolloutOperationContract(req model.FindXAgentConfigRolloutRequest, metadata map[string]string, missing []string) gin.H {
	if !isCMDBHostPluginRollout(req, metadata) {
		return gin.H{
			"mode":              configRolloutOperationMode(req, metadata),
			"contract":          "findx.agent.config_rollout.runtime.v1",
			"remote_mutation":   req.RemoteMutation,
			"required_receipts": []string{"writer_receipt", "reload_receipt", "restart_receipt", "drift_receipt", "rollback_receipt", "data_arrival_receipt", "evidence_chain"},
			"missing_contracts": configRolloutReceiptMissingContracts(missing),
			"status":            model.FindXAgentExecutionStateBlockedByContract,
		}
	}
	mode := configRolloutOperationMode(req, metadata)
	return gin.H{
		"mode":              mode,
		"contract":          configRolloutOperationContractID(mode),
		"remote_mutation":   mode == configRolloutPluginOperationDispatch,
		"required_receipts": configRolloutOperationRequiredReceipts(mode),
		"missing_contracts": configRolloutOperationMissingContracts(req, missing, mode),
		"status":            model.FindXAgentExecutionStateBlockedByContract,
	}
}

func configRolloutOperationMode(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	for _, value := range []string{
		req.TemplateID,
		req.RolloutStrategy,
	} {
		if mode, ok := configRolloutOperationFromValue(value); ok {
			return mode
		}
	}
	if req.RemoteMutation {
		return configRolloutPluginOperationDispatch
	}
	return configRolloutPluginOperationAssign
}

func configRolloutOperationFromValue(value string) (string, bool) {
	clean := strings.ToLower(strings.TrimSpace(value))
	switch {
	case clean == configRolloutPluginOperationAssign || strings.Contains(clean, "plugin-assign"):
		return configRolloutPluginOperationAssign, true
	case clean == configRolloutPluginOperationDispatch || strings.Contains(clean, "plugin-dispatch"):
		return configRolloutPluginOperationDispatch, true
	default:
		return "", false
	}
}

func configRolloutOperationIdentityConflict(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) bool {
	if strings.EqualFold(strings.TrimSpace(metadata["plugin_action_conflict"]), "blocked") {
		return true
	}
	expected := ""
	for _, value := range []string{
		req.TemplateID,
		req.RolloutStrategy,
		metadata["plugin_action"],
	} {
		mode, ok := configRolloutOperationFromValue(value)
		if !ok {
			continue
		}
		if expected == "" {
			expected = mode
			continue
		}
		if mode != expected {
			return true
		}
	}
	if expected == "" {
		return false
	}
	return (expected == configRolloutPluginOperationAssign && req.RemoteMutation) ||
		(expected == configRolloutPluginOperationDispatch && !req.RemoteMutation)
}

func configRolloutOperationContractID(mode string) string {
	if mode == configRolloutPluginOperationAssign {
		return "cmdb.agent.plugin.assignment.v1"
	}
	return "cmdb.agent.plugin.dispatch.v1"
}

func configRolloutOperationRequiredReceipts(mode string) []string {
	if mode == configRolloutPluginOperationAssign {
		return []string{"assignment_record", "target_binding_ref", "credential_policy_ref", "audit_ref"}
	}
	return []string{"assignment_ref", "writer_receipt", "delivery_receipt", "effect_receipt", "rollback_receipt", "data_arrival_receipt", "evidence_chain"}
}

func configRolloutOperationMissingContracts(req model.FindXAgentConfigRolloutRequest, missing []string, mode string) []string {
	required := cmdbHostPluginOperationMissingContractsFromExisting(req, missing, mode)
	for _, value := range missing {
		if !containsConfigRolloutRef(required, value) && cmdbHostPluginOperationMissingContractAllowed(value, mode) {
			required = append(required, value)
		}
	}
	return uniquePackageRepositoryBlockers(required)
}

func cmdbHostPluginOperationMissingContracts(req model.FindXAgentConfigRolloutRequest, mode string) []string {
	return cmdbHostPluginOperationMissingContractsFromExisting(req, nil, mode)
}

func cmdbHostPluginOperationMissingContractsFromExisting(req model.FindXAgentConfigRolloutRequest, existing []string, mode string) []string {
	missing := cmdbHostPluginBaseMissingContracts(existing)
	if strings.TrimSpace(req.CredentialRef) == "" {
		missing = append(missing, "credential_ref")
	}
	if mode == configRolloutPluginOperationAssign && req.RemoteMutation {
		missing = append(missing, configRolloutPluginOperationConflict)
	}
	if mode == configRolloutPluginOperationAssign {
		for _, value := range []string{
			cmdbAgentPluginAssignmentStoreContract,
			cmdbAgentPluginTargetBindingContract,
			cmdbAgentPluginAssignmentAuditContract,
		} {
			if len(existing) == 0 || containsConfigRolloutRef(existing, value) {
				missing = append(missing, value)
			}
		}
		return missing
	}
	if len(existing) == 0 || containsConfigRolloutRef(existing, cmdbAgentPluginAssignmentRefContract) {
		missing = append(missing, cmdbAgentPluginAssignmentRefContract)
	}
	return append(missing,
		"cmdb_agent_rollout_writer_receipt_contract",
		"cmdb_agent_rollout_delivery_receipt_contract",
		"cmdb_agent_rollout_effect_receipt_contract",
		"cmdb_agent_rollout_rollback_receipt_contract",
		"cmdb_agent_rollout_data_arrival_contract",
		"cmdb_agent_rollout_evidence_chain_contract",
	)
}

func cmdbHostPluginBaseMissingContracts(existing []string) []string {
	base := []string{
		"cmdb_agent_plugin_credential_contract",
		"cmdb_credential_ref_resolve_contract",
		"cmdb_plugin_config_schema_contract",
		"cmdb_dashboard_template_lookup_contract",
		"cmdb_dashboard_import_runtime_contract",
	}
	if existing == nil {
		return base
	}
	missing := make([]string, 0, len(base))
	for _, value := range base {
		if containsConfigRolloutRef(existing, value) {
			missing = append(missing, value)
		}
	}
	return missing
}

func cmdbHostPluginOperationMissingContractAllowed(value, mode string) bool {
	if mode == configRolloutPluginOperationAssign {
		return !strings.Contains(value, "rollout_delivery") &&
			!strings.Contains(value, "rollout_effect") &&
			!strings.Contains(value, "rollout_writer") &&
			!strings.Contains(value, "rollout_rollback") &&
			!strings.Contains(value, "rollout_data_arrival") &&
			!strings.Contains(value, "rollout_evidence_chain") &&
			!strings.Contains(value, "assignment_ref")
	}
	return !strings.Contains(value, "assignment_store") &&
		!strings.Contains(value, "target_binding") &&
		!strings.Contains(value, "assignment_audit")
}


const (
	cmdbAgentPluginCredentialContract        = "cmdb_agent_plugin_credential_contract"
	cmdbCredentialRefResolveContract         = "cmdb_credential_ref_resolve_contract"
	cmdbCredentialScopePolicyContract        = "cmdb_credential_scope_policy_contract"
	cmdbAgentPluginCredentialScopeContract   = "cmdb_agent_plugin_credential_scope_contract"
	cmdbAgentPluginCredentialPolicyContract  = "cmdb_agent_plugin_credential_policy_contract"
	cmdbCredentialScopeAuthorizationContract = "cmdb_credential_scope_authorization_contract"
	cmdbCredentialDispatchPolicyGateContract = "cmdb_credential_dispatch_policy_gate_contract"
)

type cmdbPluginCredentialContext struct {
	Required               bool
	Provided               bool
	Resolved               bool
	Mode                   string
	Protocol               string
	PortPresent            bool
	ScopePolicyRequired    bool
	ScopePolicyResolved    bool
	ScopeAuthorized        bool
	PolicyApprovalRef      string
	HostRefPresent         bool
	AgentRefPresent        bool
	BusinessContextPresent bool
	PluginRefPresent       bool
	ProviderModePresent    bool
	PolicyReasonCode       string
}

func cmdbResolvePluginCredentialContext(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) cmdbPluginCredentialContext {
	if !isCMDBHostPluginRollout(req, metadata) || !cmdbPluginCatalogRequiresCredential(req.PluginID) {
		return cmdbPluginCredentialContext{}
	}
	ref := strings.TrimSpace(req.CredentialRef)
	ctx := cmdbPluginCredentialContext{
		Required:               true,
		Provided:               ref != "",
		Mode:                   "credential_ref",
		ScopePolicyRequired:    true,
		HostRefPresent:         cmdbPluginAssignmentHostRef(req, metadata) != "",
		AgentRefPresent:        cmdbPluginAssignmentAgentRef(req, metadata) != "",
		BusinessContextPresent: cmdbCredentialBusinessContextPresent(metadata),
		PluginRefPresent:       sanitizeRemoteMutationValue("plugin_id", req.PluginID) != "",
		ProviderModePresent:    sanitizeRemoteMutationValue("provider_mode", req.ProviderMode) != "",
	}
	if ref == "" {
		ctx.PolicyReasonCode = "credential_ref_required"
		return ctx
	}
	credential, ok := store.GetCredential(ref)
	if !ok || credential == nil || strings.TrimSpace(credential.ID) == "" {
		ctx.PolicyReasonCode = "credential_ref_unresolved"
		return ctx
	}
	ctx.Resolved = true
	ctx.Protocol = sanitizeRemoteMutationValue("protocol", credential.Protocol)
	ctx.PortPresent = credential.Port > 0
	ctx.PolicyReasonCode = "scope_policy_contract_missing"
	cmdbApplyCredentialPolicyApproval(req, metadata, &ctx)
	return ctx
}

func cmdbPluginCatalogRequiresCredential(pluginID string) bool {
	clean := normalizeFindXPluginCredentialID(pluginID)
	if clean == "" {
		return false
	}
	for _, plugin := range pluginCatalog {
		if normalizeFindXPluginCredentialID(plugin.ID) == clean {
			return plugin.CredentialRequired
		}
	}
	return findXPluginCredentialRequired(clean)
}

func normalizeFindXPluginCredentialID(pluginID string) string {
	clean := strings.ToLower(strings.TrimSpace(pluginID))
	clean = strings.TrimPrefix(clean, "inputs.")
	clean = strings.TrimPrefix(clean, "input.")
	return clean
}

func cmdbApplyCredentialResolveGate(missing []string, ctx cmdbPluginCredentialContext) []string {
	if !ctx.Required {
		return missing
	}
	out := make([]string, 0, len(missing)+2)
	for _, value := range missing {
		if ctx.Resolved && value == cmdbCredentialRefResolveContract {
			continue
		}
		out = append(out, value)
	}
	if !ctx.Provided {
		out = appendMissingConfigRolloutContract(out, "credential_ref")
	}
	out = appendMissingConfigRolloutContract(out, cmdbAgentPluginCredentialContract)
	if !ctx.Resolved {
		out = appendMissingConfigRolloutContract(out, cmdbCredentialRefResolveContract)
	}
	return uniquePackageRepositoryBlockers(out)
}

func cmdbApplyCredentialScopePolicyGate(missing []string, ctx cmdbPluginCredentialContext, mode string) []string {
	if !ctx.Required || !ctx.Resolved {
		return uniquePackageRepositoryBlockers(missing)
	}
	if ctx.ScopePolicyResolved && ctx.ScopeAuthorized {
		return cmdbFilterCredentialScopePolicyMissingContracts(missing, mode)
	}
	out := append([]string{}, missing...)
	for _, contract := range []string{
		cmdbCredentialScopePolicyContract,
		cmdbAgentPluginCredentialScopeContract,
		cmdbAgentPluginCredentialPolicyContract,
		cmdbCredentialScopeAuthorizationContract,
	} {
		out = appendMissingConfigRolloutContract(out, contract)
	}
	if mode == configRolloutPluginOperationDispatch {
		out = appendMissingConfigRolloutContract(out, cmdbCredentialDispatchPolicyGateContract)
	}
	return uniquePackageRepositoryBlockers(out)
}

func cmdbApplyCredentialPolicyApproval(req model.FindXAgentConfigRolloutRequest, metadata map[string]string, ctx *cmdbPluginCredentialContext) {
	if ctx == nil || !ctx.Required || !ctx.Resolved {
		return
	}
	approvalRef := cmdbCredentialPolicyApprovalRef(metadata)
	if approvalRef == "" || !cmdbCredentialPolicyApprovalMatches(approvalRef, req, metadata, *ctx) {
		return
	}
	ctx.ScopePolicyResolved = true
	ctx.ScopeAuthorized = true
	ctx.PolicyApprovalRef = approvalRef
	ctx.PolicyReasonCode = "credential_policy_release_recorded"
}

func cmdbCredentialPolicyApprovalRef(metadata map[string]string) string {
	for _, key := range []string{"scope_policy_approval_ref", "credential_policy_approval_ref", "credential_scope_policy_approval_ref"} {
		if ref := sanitizeRemoteMutationValue("approval_ref", metadata[key]); ref != "" {
			return ref
		}
	}
	return ""
}

func cmdbCredentialPolicyApprovalMatches(ref string, req model.FindXAgentConfigRolloutRequest, metadata map[string]string, ctx cmdbPluginCredentialContext) bool {
	approval, ok := store.GetCmdbResourceApproval(ref)
	if !ok || approval == nil {
		return false
	}
	if approval.Status != "review_accept_recorded" || approval.WorkflowState != "review_accept_recorded" {
		return false
	}
	if approval.ResourceType != "cmdb_agent_plugin_credential_policy" || approval.Action != "cmdb.agent.plugin.credential_policy.release" {
		return false
	}
	if approval.ResourceID != cmdbCredentialPolicyResourceID(req, metadata) {
		return false
	}
	context := map[string]string{}
	if err := json.Unmarshal([]byte(approval.ContextJSON), &context); err != nil {
		return false
	}
	expected := map[string]string{
		"host_ref":            cmdbPluginAssignmentHostRef(req, metadata),
		"agent_ref":           cmdbPluginAssignmentAgentRef(req, metadata),
		"plugin_id":           sanitizeRemoteMutationValue("plugin_id", req.PluginID),
		"provider_mode":       sanitizeRemoteMutationValue("provider_mode", req.ProviderMode),
		"operation_mode":      configRolloutOperationMode(req, metadata),
		"credential_protocol": ctx.Protocol,
	}
	for key, want := range expected {
		if want == "" || cmdbCredentialApprovalContextValue(key, context[key]) != want {
			return false
		}
	}
	for _, key := range []string{"business_group_ref", "team_ref"} {
		if want := cmdbCredentialApprovalContextValue(key, metadata[key]); want != "" && cmdbCredentialApprovalContextValue(key, context[key]) != want {
			return false
		}
	}
	return true
}

func cmdbCredentialApprovalContextValue(key, value string) string {
	if key == "credential_protocol" {
		return sanitizeRemoteMutationValue("protocol", value)
	}
	return sanitizeRemoteMutationValue(key, value)
}

func cmdbCredentialPolicyResourceID(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	parts := []string{
		cmdbPluginAssignmentHostRef(req, metadata),
		cmdbPluginAssignmentAgentRef(req, metadata),
		sanitizeRemoteMutationValue("plugin_id", req.PluginID),
	}
	for _, part := range parts {
		if part == "" {
			return ""
		}
	}
	return strings.Join(parts, ":")
}

func cmdbFilterCredentialScopePolicyMissingContracts(missing []string, mode string) []string {
	out := make([]string, 0, len(missing))
	for _, value := range missing {
		switch value {
		case cmdbCredentialScopePolicyContract,
			cmdbAgentPluginCredentialScopeContract,
			cmdbAgentPluginCredentialPolicyContract,
			cmdbCredentialScopeAuthorizationContract:
			continue
		case cmdbCredentialDispatchPolicyGateContract:
			if mode == configRolloutPluginOperationDispatch {
				continue
			}
		}
		out = append(out, value)
	}
	return uniquePackageRepositoryBlockers(out)
}

func cmdbCredentialBusinessContextPresent(metadata map[string]string) bool {
	for _, key := range []string{"business_group_ref", "business_group_id", "team_ref", "team_id", "tenant_id"} {
		if sanitizeRemoteMutationValue(key, metadata[key]) != "" {
			return true
		}
	}
	return false
}

func appendMissingConfigRolloutContract(values []string, contract string) []string {
	if contract == "" || containsConfigRolloutRef(values, contract) {
		return values
	}
	return append(values, contract)
}

func cmdbPluginCredentialContractResponse(ctx cmdbPluginCredentialContext) gin.H {
	if !ctx.Required {
		return nil
	}
	out := gin.H{
		"contract_id":              "cmdb.agent.plugin.credential_ref.read.v1",
		"mode":                     ctx.Mode,
		"credential_ref_required":  true,
		"credential_ref_present":   ctx.Provided,
		"credential_ref_resolved":  ctx.Resolved,
		"scope_policy_contract_id": "cmdb.credential.scope_policy.read.v1",
		"scope_policy_required":    ctx.ScopePolicyRequired,
		"scope_policy_resolved":    ctx.ScopePolicyResolved,
		"scope_authorized":         ctx.ScopeAuthorized,
		"host_ref_present":         ctx.HostRefPresent,
		"agent_ref_present":        ctx.AgentRefPresent,
		"business_context_present": ctx.BusinessContextPresent,
		"plugin_ref_present":       ctx.PluginRefPresent,
		"provider_mode_present":    ctx.ProviderModePresent,
		"policy_reason_code":       ctx.PolicyReasonCode,
		"status":                   model.FindXAgentExecutionStateBlockedByContract,
	}
	if ctx.Resolved {
		out["protocol"] = ctx.Protocol
		out["port_present"] = ctx.PortPresent
	}
	if ctx.ScopePolicyResolved && ctx.ScopeAuthorized && ctx.PolicyApprovalRef != "" {
		out["policy_approval_ref"] = ctx.PolicyApprovalRef
		out["policy_release_contract_id"] = "cmdb.agent.plugin.credential_policy.release.v1"
	}
	return out
}


func cmdbAttachConfigRolloutReceiptRequestRefs(saved model.FindXAgentConfigRollout, req model.FindXAgentConfigRolloutRequest, metadata map[string]string, actor string) (model.FindXAgentConfigRollout, error) {
	if !isCMDBHostPluginRollout(req, metadata) || configRolloutOperationMode(req, metadata) != configRolloutPluginOperationDispatch {
		return saved, nil
	}
	if saved.Metadata == nil {
		saved.Metadata = map[string]string{}
	}
	for _, receipt := range []string{"delivery", "effect", "rollback"} {
		refKey := receipt + "_request_ref"
		if strings.TrimSpace(saved.Metadata[refKey]) != "" {
			continue
		}
		task, err := cmdbCreateConfigRolloutReceiptRequest(saved, req, metadata, receipt, actor)
		if err != nil {
			return saved, err
		}
		saved.Metadata[refKey] = task.ID
	}
	return store.SaveFindXAgentConfigRollout(saved)
}

func cmdbCreateConfigRolloutReceiptRequest(saved model.FindXAgentConfigRollout, req model.FindXAgentConfigRolloutRequest, metadata map[string]string, receipt, actor string) (model.FindXAgentExecutionTask, error) {
	taskMetadata := cmdbConfigRolloutReceiptRequestMetadata(saved, req, metadata, receipt, actor)
	task := model.FindXAgentExecutionTask{
		Action:               "config_rollout_receipt",
		AgentIDs:             cleanAgentLifecycleValues(saved.AgentIDs),
		TargetIDs:            cleanAgentLifecycleValues(saved.TargetIDs),
		ConfigVersion:        sanitizeRemoteMutationValue("config_version", saved.ConfigVersion),
		Status:               "accepted",
		Blocker:              "",
		Audit:                "findx_agent.task.created",
		CredentialRefPresent: strings.TrimSpace(req.CredentialRef) != "",
		Metadata:             taskMetadata,
	}
	return store.SaveFindXAgentExecutionTask(task)
}

func cmdbConfigRolloutReceiptRequestMetadata(saved model.FindXAgentConfigRollout, req model.FindXAgentConfigRolloutRequest, metadata map[string]string, receipt, actor string) map[string]string {
	out := map[string]string{
		"scope":              configRolloutScopeCMDBHost,
		"source_rollout_id":  saved.ID,
		"config_rollout_id":  saved.ID,
		"rollout_ref":        saved.ID,
		"receipt_kind":       receipt,
		"receipt_type":       receipt,
		"phase":              receipt,
		"plugin_id":          sanitizeRemoteMutationValue("plugin_id", saved.PluginID),
		"provider_mode":      sanitizeRemoteMutationValue("provider_mode", saved.ProviderMode),
		"config_version":     sanitizeRemoteMutationValue("config_version", saved.ConfigVersion),
		"config_snippet_ref": sanitizeRemoteMutationValue("config_snippet_ref", saved.ConfigSnippetRef),
		"rollback_ref":       sanitizeRemoteMutationValue("rollback_ref", saved.RollbackRef),
		"idempotency_key":    sanitizeRemoteMutationValue("idempotency_key", "cmdb-agent-rollout-"+receipt+"-"+saved.ID),
		"requested_by":       sanitizeRemoteMutationValue("requested_by", actor),
	}
	for dst, src := range map[string]string{
		"cmdb_host_ref":      firstNonEmpty(metadata["cmdb_host_ref"], cmdbPluginAssignmentHostRef(req, metadata)),
		"agent_ref":          firstNonEmpty(metadata["agent_ref"], cmdbPluginAssignmentAgentRef(req, metadata)),
		"assignment_ref":     metadata["assignment_ref"],
		"target_binding_ref": metadata["target_binding_ref"],
	} {
		if clean := sanitizeRemoteMutationValue(dst, src); clean != "" {
			out[dst] = clean
		}
	}
	if receipt == "delivery" {
		cmdbAttachDeliveryExecutorBoundaryMetadata(out)
	}
	if receipt == "effect" {
		cmdbAttachEffectExecutorBoundaryMetadata(out)
	}
	if receipt == "rollback" {
		cmdbAttachRollbackExecutorBoundaryMetadata(out)
	}
	return out
}

func cmdbAttachDeliveryExecutorBoundaryMetadata(metadata map[string]string) {
	metadata["delivery_executor_contract_status"] = "blocked"
	metadata["delivery_executor_registration_contract"] = "cmdb_agent_rollout_delivery_executor_registration_contract"
	metadata["delivery_runner_identity_contract"] = "cmdb_agent_rollout_delivery_runner_identity_contract"
	metadata["delivery_attested_receipt_contract"] = "cmdb_agent_rollout_delivery_attested_receipt_contract"
	metadata["delivery_target_binding_contract"] = "cmdb_agent_rollout_delivery_target_binding_contract"
	metadata["delivery_request_ref_match_contract"] = "cmdb_agent_rollout_delivery_request_ref_match_contract"
	metadata["delivery_executor_missing_contracts"] = configRolloutReceiptIngestionMissingJSON([]string{
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_delivery_executor_registration_contract",
		"cmdb_agent_rollout_delivery_runner_identity_contract",
		"cmdb_agent_rollout_delivery_attested_receipt_contract",
		"cmdb_agent_rollout_delivery_target_binding_contract",
		"cmdb_agent_rollout_delivery_request_ref_match_contract",
	})
}

func cmdbAttachEffectExecutorBoundaryMetadata(metadata map[string]string) {
	metadata["effect_executor_contract_status"] = "blocked"
	metadata["effect_executor_registration_contract"] = "cmdb_agent_rollout_effect_executor_registration_contract"
	metadata["effect_runner_identity_contract"] = "cmdb_agent_rollout_effect_runner_identity_contract"
	metadata["effect_delivery_evidence_contract"] = "cmdb_agent_rollout_effect_delivery_evidence_match_contract"
	metadata["effect_attested_receipt_contract"] = "cmdb_agent_rollout_effect_attested_receipt_contract"
	metadata["effect_request_ref_match_contract"] = "cmdb_agent_rollout_effect_request_ref_match_contract"
	metadata["effect_executor_missing_contracts"] = configRolloutReceiptIngestionMissingJSON([]string{
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_effect_executor_registration_contract",
		"cmdb_agent_rollout_effect_runner_identity_contract",
		"cmdb_agent_rollout_effect_delivery_evidence_match_contract",
		"cmdb_agent_rollout_effect_attested_receipt_contract",
		"cmdb_agent_rollout_effect_request_ref_match_contract",
	})
}

func cmdbAttachRollbackExecutorBoundaryMetadata(metadata map[string]string) {
	metadata["rollback_executor_contract_status"] = "blocked"
	metadata["rollback_executor_registration_contract"] = "cmdb_agent_rollout_rollback_executor_registration_contract"
	metadata["rollback_runner_identity_contract"] = "cmdb_agent_rollout_rollback_runner_identity_contract"
	metadata["rollback_operation_context_contract"] = "cmdb_agent_rollout_rollback_operation_context_contract"
	metadata["rollback_attested_receipt_contract"] = "cmdb_agent_rollout_rollback_attested_receipt_contract"
	metadata["rollback_request_ref_match_contract"] = "cmdb_agent_rollout_rollback_request_ref_match_contract"
	metadata["rollback_executor_missing_contracts"] = configRolloutReceiptIngestionMissingJSON([]string{
		"cmdb_agent_rollout_rollback_executor_contract",
		"cmdb_agent_rollout_rollback_executor_registration_contract",
		"cmdb_agent_rollout_rollback_runner_identity_contract",
		"cmdb_agent_rollout_rollback_operation_context_contract",
		"cmdb_agent_rollout_rollback_attested_receipt_contract",
		"cmdb_agent_rollout_rollback_request_ref_match_contract",
	})
}
