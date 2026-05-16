package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func ListFindXAgentInstallPlans(c *gin.Context) {
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		item, ok, err := store.GetFindXAgentInstallPlan(id)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "install plan detail unavailable"})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "install plan not found"})
			return
		}
		c.JSON(http.StatusOK, item)
		return
	}
	items, err := store.ListFindXAgentInstallPlans()
	writeAgentLifecycleList(c, items, err, "install plan list unavailable")
}

func ListFindXAgentInstallExecutions(c *gin.Context) {
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		item, ok, err := store.GetFindXAgentInstallExecution(id)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "install execution detail unavailable"})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "install execution not found"})
			return
		}
		c.JSON(http.StatusOK, item)
		return
	}
	items, err := store.ListFindXAgentInstallExecutions()
	writeAgentLifecycleList(c, items, err, "install execution list unavailable")
}

func ListFindXAgentConfigRollouts(c *gin.Context) {
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		item, ok, err := store.GetFindXAgentConfigRollout(id)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config rollout detail unavailable"})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "config rollout not found"})
			return
		}
		if gate := configRolloutRuntimeReadGateForItem(item); gate.Blocked {
			writeConfigRolloutRuntimeReadBlocked(c, item, gate)
			return
		}
		c.JSON(http.StatusOK, safeConfigRolloutRuntimeReadDetail(item))
		return
	}
	items, err := store.ListFindXAgentConfigRollouts()
	writeAgentLifecycleList(c, items, err, "config rollout list unavailable")
}

func safeConfigRolloutRuntimeReadDetail(item model.FindXAgentConfigRollout) model.FindXAgentConfigRollout {
	metadata := make(map[string]string, len(item.Metadata))
	for key, value := range item.Metadata {
		cleanKey := strings.ToLower(strings.TrimSpace(key))
		if cleanKey == "writer_request_ref" || cleanKey == "cmdb_agent_rollout_writer_request_ref_contract" {
			continue
		}
		metadata[key] = value
	}
	if isCMDBHostPluginDispatchRolloutRecord(item) {
		missing := cmdbAgentRolloutRuntimeExecutorGapContractsForItem(item)
		metadata["runtime_read_status"] = "blocked"
		metadata["runtime_read_contract"] = cmdbAgentRolloutRuntimeReadContract
		metadata["runtime_read_missing_contracts"] = configRolloutRuntimeReadMissingJSON(missing)
	}
	item.Metadata = metadata
	return item
}

func ListFindXAgentTasks(c *gin.Context) {
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		item, ok, err := store.GetFindXAgentExecutionTask(id)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent task detail unavailable"})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent task not found"})
			return
		}
		c.JSON(http.StatusOK, item)
		return
	}
	items, err := store.ListFindXAgentExecutionTasks()
	writeAgentLifecycleList(c, items, err, "agent task list unavailable")
}

func ListFindXAgentDataArrivalEvidence(c *gin.Context) {
	if handled, err := listFindXAgentDataArrivalEvidenceRuntimeRead(c); handled {
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "data arrival evidence runtime read unavailable"})
		}
		return
	}
	items, err := store.ListFindXAgentDataArrivalEvidence()
	writeAgentLifecycleList(c, items, err, "data arrival evidence list unavailable")
}

func writeAgentLifecycleList(c *gin.Context, items any, err error, message string) {
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": message})
		return
	}
	c.JSON(http.StatusOK, items)
}

func saveBlockedAgentTask(c *gin.Context, req model.FindXAgentTaskRequest, action string) (model.FindXAgentExecutionTask, error) {
	metadata := safeAgentLifecycleMetadata(req.Metadata)
	credentialRefPresent := strings.TrimSpace(req.CredentialRef) != ""
	blocker := blockedAgentTaskReason(action, metadata, credentialRefPresent)
	task := model.FindXAgentExecutionTask{
		Action:               action,
		AgentIDs:             cleanAgentLifecycleValues(req.AgentIDs),
		TargetIDs:            cleanAgentLifecycleValues(req.TargetIDs),
		PackageID:            sanitizeRemoteMutationValue("package_id", req.PackageID),
		ConfigVersion:        sanitizeRemoteMutationValue("config_version", req.ConfigVersion),
		Status:               "blocked",
		Blocker:              blocker,
		Audit:                "findx_agent.task.requested",
		CredentialRefPresent: credentialRefPresent,
		Metadata:             metadata,
	}
	saved, err := store.SaveFindXAgentExecutionTask(task)
	if err != nil {
		return model.FindXAgentExecutionTask{}, err
	}
	auditEvent(c, "findx_agent.task.requested", saved.ID, "medium", "blocked", saved.Blocker, c.GetHeader("X-Test-Batch-Id"))
	return saved, nil
}

func blockedAgentTaskReason(action string, metadata map[string]string, credentialRefPresent bool) string {
	missing := missingAgentTaskRefs(action, metadata, credentialRefPresent)
	if len(missing) > 0 {
		return fmt.Sprintf("PENDING: missing %s", strings.Join(missing, ", "))
	}
	return "PENDING: executor not enabled / execution protocol not open"
}

func missingAgentTaskRefs(action string, metadata map[string]string, credentialRefPresent bool) []string {
	required := append(requiredAgentTaskRefs(action), requiredRemoteExecutionRefs()...)
	required = append(required, requiredAgentTaskExecutorReceiptRefs(metadata)...)
	if isKubernetesAgentTask(metadata) {
		required = append(required, requiredKubernetesAgentTaskRefs(action, metadata)...)
	}
	missingSet := map[string]bool{}
	for _, key := range required {
		if strings.TrimSpace(metadata[key]) == "" {
			missingSet[key] = true
		}
	}
	if !credentialRefPresent {
		missingSet["credential_ref"] = true
	}
	for _, key := range missingRemoteExecutionChoiceRefs(metadata) {
		missingSet[key] = true
	}
	if isKubernetesAgentTask(metadata) {
		for _, key := range missingKubernetesAgentTaskChoiceRefs(action, metadata) {
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

func agentTaskResponseBlockers(missing []string) []string {
	if len(missing) == 0 {
		return []string{agentBlocked, "EXECUTOR_DISABLED_BY_CONTRACT"}
	}
	values := append([]string{agentBlocked, "MISSING_CONTRACTS"}, missing...)
	return uniquePackageRepositoryBlockers(values)
}

func requiredRemoteExecutionRefs() []string {
	return []string{"idempotency_key", "target_os", "timeout_policy_ref"}
}

func requiredAgentTaskExecutorReceiptRefs(metadata map[string]string) []string {
	refs := []string{"data_arrival_validator_ref"}
	text := agentTaskMatrixText(metadata)
	if isLocalLinuxAgentTask(metadata, text) {
		refs = append(refs, "local_executor_ref", "linux_installer_ref", "service_receipt_ref")
	}
	if isLocalWindowsAgentTask(metadata, text) {
		refs = append(refs, "local_executor_ref", "windows_installer_ref", "service_receipt_ref")
	}
	addAgentTaskRefsIf(&refs, text, []string{"local-service"}, "service_manifest_ref", "service_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"ssh"}, "ssh_runner_ref", "remote_executor_ref")
	addAgentTaskRefsIf(&refs, text, []string{"winrm"}, "winrm_endpoint_ref", "winrm_transport_ref", "remote_executor_ref")
	addAgentTaskRefsIf(&refs, text, []string{"systemd"}, "systemd_unit_ref", "systemd_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"windows-service", "windows service"}, "windows_service_ref", "windows_service_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"iis"}, "iis_site_ref", "iis_app_pool_ref", "iis_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"docker"}, "container_ref", "image_ref", "docker_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"operator"}, "operator_ref", "crd_ref", "controller_receipt_ref")
	addAgentTaskRefsIf(&refs, text, []string{"daemonset"}, "daemonset_ref")
	addAgentTaskRefsIf(&refs, text, []string{"sidecar"}, "sidecar_injection_ref")
	addAgentTaskRefsIf(&refs, text, []string{"initcontainer", "init-container"}, "init_container_ref")
	return refs
}

func agentTaskMatrixText(metadata map[string]string) string {
	keys := []string{"target_os", "transport", "runner", "method", "orchestrator", "workload_kind", "service_type", "runtime"}
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, strings.ToLower(strings.TrimSpace(metadata[key])))
	}
	return strings.Join(values, " ")
}

func isLocalLinuxAgentTask(metadata map[string]string, text string) bool {
	targetOS := strings.ToLower(strings.TrimSpace(metadata["target_os"]))
	return strings.Contains(text, "local-linux") || (targetOS == "linux" && strings.TrimSpace(metadata["transport"]) == "" && strings.TrimSpace(metadata["runner"]) == "")
}

func isLocalWindowsAgentTask(metadata map[string]string, text string) bool {
	targetOS := strings.ToLower(strings.TrimSpace(metadata["target_os"]))
	return strings.Contains(text, "local-windows") || (targetOS == "windows" && strings.TrimSpace(metadata["transport"]) == "" && strings.TrimSpace(metadata["runner"]) == "")
}

func addAgentTaskRefsIf(refs *[]string, text string, markers []string, values ...string) {
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			*refs = append(*refs, values...)
			return
		}
	}
}

func missingRemoteExecutionChoiceRefs(metadata map[string]string) []string {
	missing := []string{}
	text := agentTaskMatrixText(metadata)
	if strings.TrimSpace(metadata["transport"]) == "" && strings.TrimSpace(metadata["runner"]) == "" && !isLocalLinuxAgentTask(metadata, text) && !isLocalWindowsAgentTask(metadata, text) {
		missing = append(missing, "transport_or_runner")
	}
	if strings.TrimSpace(metadata["execution_receipt_ref"]) == "" && strings.TrimSpace(metadata["receipt_ref"]) == "" {
		missing = append(missing, "execution_receipt_ref_or_receipt_ref")
	}
	if strings.TrimSpace(metadata["audit_ref"]) == "" && strings.TrimSpace(metadata["evidence_chain_ref"]) == "" {
		missing = append(missing, "audit_ref_or_evidence_chain_ref")
	}
	if strings.Contains(text, "ssh") && strings.TrimSpace(metadata["ssh_host_key"]) == "" && strings.TrimSpace(metadata["ssh_fingerprint"]) == "" {
		missing = append(missing, "ssh_host_key_or_fingerprint")
		missing = append(missing, "ssh_host_key", "ssh_fingerprint")
	}
	return missing
}

func requiredAgentTaskRefs(action string) []string {
	switch action {
	case "sync_package_repository":
		return []string{"package_repository_ref", "manifest_ref", "checksum", "signature_ref", "executor_ref"}
	case "publish_package":
		return []string{"package_id", "package_repository_ref", "release_manifest_ref", "checksum", "signature_ref", "executor_ref"}
	case "download_package":
		return []string{"package_id", "package_repository_ref", "artifact_ref", "checksum", "signature_ref", "public_key_ref", "executor_ref"}
	case "verify_package_signature":
		return []string{"package_id", "package_repository_ref", "checksum", "signature_ref", "public_key_ref", "verifier_ref"}
	case "uninstall":
		return []string{"uninstall_manifest_ref", "executor_ref"}
	case "rollback":
		return []string{"rollback_manifest_ref", "state_snapshot_ref", "executor_ref"}
	case "upgrade":
		return []string{"package_repository_ref", "signature_ref", "checksum", "script_manifest_ref", "executor_ref"}
	case "restart":
		return []string{"service_ref", "executor_ref"}
	default:
		return []string{}
	}
}

func isKubernetesAgentTask(metadata map[string]string) bool {
	for _, key := range kubernetesAgentTaskMarkerKeys() {
		value := strings.ToLower(strings.TrimSpace(metadata[key]))
		for _, marker := range kubernetesAgentTaskMarkers() {
			if strings.Contains(value, marker) {
				return true
			}
		}
	}
	return false
}

func kubernetesAgentTaskMarkerKeys() []string {
	return []string{"target_os", "runner", "transport", "orchestrator", "method", "workload_kind"}
}

func kubernetesAgentTaskMarkers() []string {
	return []string{"kubernetes", "k8s", "helm", "daemonset", "sidecar", "initcontainer", "init-container"}
}

func requiredKubernetesAgentTaskRefs(action string, metadata map[string]string) []string {
	required := []string{
		"cluster_ref",
		"namespace_ref",
		"workload_selector_ref",
		"rbac_ref",
		"service_account_ref",
		"rollout_strategy_ref",
		"rollout_receipt_ref",
		"data_arrival_validator_ref",
	}
	if isHelmAgentTask(metadata) {
		required = append(required, "helm_release_ref")
	}
	switch action {
	case "upgrade":
		required = append(required, "values_ref", "image_ref", "config_map_ref")
	case "rollback":
		required = append(required, "rollback_revision_ref")
	case "uninstall":
		required = append(required, "teardown_plan_ref")
	case "restart":
		required = append(required, "restart_strategy_ref")
	}
	return required
}

func missingKubernetesAgentTaskChoiceRefs(action string, metadata map[string]string) []string {
	if action != "upgrade" {
		return nil
	}
	if strings.TrimSpace(metadata["helm_chart_ref"]) != "" || strings.TrimSpace(metadata["manifest_bundle_ref"]) != "" {
		return nil
	}
	return []string{"helm_chart_ref_or_manifest_bundle_ref"}
}

func isHelmAgentTask(metadata map[string]string) bool {
	for _, key := range kubernetesAgentTaskMarkerKeys() {
		if strings.Contains(strings.ToLower(strings.TrimSpace(metadata[key])), "helm") {
			return true
		}
	}
	return false
}

func hasAgentTaskTarget(req model.FindXAgentTaskRequest) bool {
	return len(cleanAgentLifecycleValues(req.AgentIDs)) > 0 || len(cleanAgentLifecycleValues(req.TargetIDs)) > 0
}

func cleanAgentLifecycleValues(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean != "" && !seen[clean] {
			seen[clean] = true
			out = append(out, clean)
		}
	}
	return out
}

func safeAgentLifecycleMetadata(input map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range input {
		cleanKey := strings.TrimSpace(key)
		allowedRef := allowedAgentLifecycleReferenceKey(cleanKey)
		if cleanKey == "" || (!allowedRef && looksSensitive(value)) {
			continue
		}
		if looksSensitive(cleanKey) && !allowedAgentLifecycleReferenceKey(cleanKey) {
			continue
		}
		cleanValue := sanitizeRemoteMutationValue(cleanKey, value)
		if allowedRef {
			cleanValue = sanitizeAgentLifecycleReferenceValue(value)
		}
		if cleanValue != "" {
			out[cleanKey] = cleanValue
		}
	}
	return out
}

func allowedAgentLifecycleReferenceKey(key string) bool {
	switch strings.TrimSpace(key) {
	case "provider_auth_ref":
		return true
	default:
		return false
	}
}

func sanitizeAgentLifecycleReferenceValue(value string) string {
	clean := strings.TrimSpace(removeControlRunes(value))
	if clean == "" || looksSensitiveReferenceValue(clean) {
		return ""
	}
	const maxAgentLifecycleReferenceLen = 120
	runes := []rune(clean)
	if len(runes) > maxAgentLifecycleReferenceLen {
		clean = string(runes[:maxAgentLifecycleReferenceLen])
	}
	return clean
}

func looksSensitiveReferenceValue(value string) bool {
	normalized := strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(strings.ToLower(value))
	for _, marker := range []string{"password", "passwd", "secret", "token", "cookie", "bearer", "api_key", "apikey", "access_key", "private_key", "session", "dsn"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}
