package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/security"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func findXAgentInstallPlanMode(req model.FindXAgentInstallPlanRequest) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		if req.Execute {
			return "execute", nil
		}
		return "plan", nil
	}
	if mode != "plan" && mode != "execute" {
		return "", errors.New("mode must be plan or execute")
	}
	return mode, nil
}

func newBlockedFindXAgentInstallPlan(req model.FindXAgentInstallPlanRequest, pkg model.FindXAgentPackage, targetIDs []string) model.FindXAgentInstallPlan {
	return model.FindXAgentInstallPlan{
		PackageID:            pkg.ID,
		OS:                   strings.TrimSpace(req.OS),
		Method:               strings.TrimSpace(req.Method),
		TargetIDs:            cleanAgentLifecycleValues(targetIDs),
		Status:               model.FindXAgentExecutionStatusBlocked,
		Blocker:              packageInstallBlocker(pkg),
		Audit:                "findx_agent.install_plan.requested",
		CredentialRefPresent: strings.TrimSpace(req.CredentialRef) != "",
		Metadata:             safeLinuxInstallerMetadata(req.Metadata),
	}
}

func createBlockedFindXAgentInstallExecution(c *gin.Context, req model.FindXAgentInstallPlanRequest, plan model.FindXAgentInstallPlan) {
	targetID := firstCleanAgentLifecycleValue(plan.TargetIDs)
	gate := security.EvaluateLinuxInstallerPrerequisites(linuxInstallerPrerequisitesFromRequest(req))
	now := time.Now()
	execution := model.FindXAgentInstallExecution{
		PlanID:       plan.ID,
		TargetID:     targetID,
		Runner:       gate.Runner,
		Status:       "accepted",
		Steps:        pendingInstallExecutionSteps(now),
		EvidenceRefs: []string{"install-plan:" + plan.ID},
		ErrorSummary: "",
		StartedAt:    &now,
	}
	saved, err := store.SaveFindXAgentInstallExecution(execution)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "install execution persistence unavailable"})
		return
	}
	auditEvent(c, "findx_agent.install_execution.created", saved.ID, "high", "accepted", "", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{
		"code":      http.StatusOK,
		"status":    "accepted",
		"data":      plan,
		"execution": saved,
	})
}

func pendingInstallExecutionSteps(now time.Time) []model.FindXAgentInstallExecutionStep {
	names := []string{
		"preflight",
		"download_package",
		"verify_package",
		"register_service",
		"enable_or_start_service",
		"verify_heartbeat",
		"verify_data_arrival",
		"capture_evidence",
	}
	steps := make([]model.FindXAgentInstallExecutionStep, 0, len(names))
	for _, name := range names {
		steps = append(steps, model.FindXAgentInstallExecutionStep{
			Name:      name,
			Status:    "pending",
			UpdatedAt: now,
		})
	}
	return steps
}

func linuxInstallExecutionReceiptScope(runner string) string {
	if strings.EqualFold(strings.TrimSpace(runner), "ssh") {
		return "ssh"
	}
	return "linux_local"
}

func installExecutionMissingContracts(scope string, reason string) []string {
	missing := []string{
		"executor_contract",
		"execution_receipt_contract",
		"service_receipt_contract",
		"heartbeat_receipt_contract",
		"data_arrival_receipt_contract",
		"evidence_chain_contract",
	}
	missing = append(missing, realPluginInstallContractInventory()...)
	for _, key := range []string{
		"package_repository_ref",
		"signature_ref",
		"checksum",
		"script_manifest_ref",
		"executor_ref",
		"safety_policy_path",
		"systemd_unit_ref",
		"curl_installer_ref",
		"env_template_ref",
		"service_receipt_ref",
		"heartbeat_validator_ref",
		"data_arrival_validator_ref",
		"audit_ref_or_evidence_chain_ref",
		"runner_whitelist_ref",
		"systemd_mode",
		"systemd_unit_name_ref_or_systemd_unit_path_ref",
		"curl_command_ref",
		"curl_manifest_ref",
		"reload_receipt_ref",
		"service_status_receipt_ref",
		"credential_ref",
		"remote_executor_ref",
		"ssh_runner_ref",
		"ssh_host_key_or_fingerprint",
		"winrm_endpoint_ref",
		"winrm_transport_ref",
		"timeout_policy_ref",
		"idempotency_key",
		"install_receipt_ref",
		"execution_receipt_ref",
	} {
		if strings.Contains(reason, key) {
			missing = append(missing, key)
		}
	}
	switch scope {
	case "windows_local", "windows_service":
		missing = append(missing, "windows_service_contract")
	case "kubernetes", "helm", "operator", "daemonset", "sidecar", "initcontainer":
		missing = append(missing, "cluster_rbac_contract", "rollout_receipt_contract", "workload_status_receipt_contract")
		if scope != "kubernetes" {
			missing = append(missing, scope+"_receipt_contract")
		}
	case "ssh":
		missing = append(missing, "ssh_runner_contract", "host_key_contract")
	case "winrm":
		missing = append(missing, "winrm_transport_contract")
	default:
		missing = append(missing, "systemd_unit_contract")
	}
	return uniquePackageRepositoryBlockers(missing)
}

func realPluginInstallContractInventory() []string {
	return []string{
		"cmdb_agent_plugin_package_trust_verifier_contract",
		"cmdb_agent_plugin_remote_writer_contract",
		"cmdb_agent_plugin_delivery_executor_contract",
		"cmdb_agent_plugin_effect_executor_contract",
		"cmdb_agent_plugin_rollback_executor_contract",
		"cmdb_agent_plugin_service_registration_receipt_contract",
		"cmdb_agent_plugin_data_arrival_receiver_ingestion_contract",
		"cmdb_agent_plugin_evidence_chain_attestation_contract",
		"cmdb_agent_plugin_credential_policy_approval_contract",
	}
}

func linuxInstallerPrerequisitesFromRequest(req model.FindXAgentInstallPlanRequest) security.LinuxInstallerPrerequisites {
	metadata := req.Metadata
	return security.LinuxInstallerPrerequisites{
		PackageRepositoryRef:    installerMetadataValue(metadata, "package_repository_ref"),
		SignatureRef:            installerMetadataValue(metadata, "signature_ref"),
		Checksum:                installerMetadataValue(metadata, "checksum"),
		ScriptManifestRef:       installerMetadataValue(metadata, "script_manifest_ref"),
		ExecutorRef:             installerMetadataValue(metadata, "executor_ref"),
		SafetyPolicyPath:        installerMetadataValue(metadata, "safety_policy_path"),
		Runner:                  installerMetadataValue(metadata, "runner"),
		SSHHostKey:              installerMetadataValue(metadata, "ssh_host_key"),
		SSHFingerprint:          installerMetadataValue(metadata, "ssh_fingerprint"),
		SystemdUnitRef:          installerMetadataValue(metadata, "systemd_unit_ref"),
		SystemdUnitNameRef:      installerMetadataValue(metadata, "systemd_unit_name_ref"),
		SystemdUnitPathRef:      installerMetadataValue(metadata, "systemd_unit_path_ref"),
		SystemdMode:             installerMetadataValue(metadata, "systemd_mode"),
		CurlInstallerRef:        installerMetadataValue(metadata, "curl_installer_ref"),
		CurlCommandRef:          installerMetadataValue(metadata, "curl_command_ref"),
		CurlManifestRef:         installerMetadataValue(metadata, "curl_manifest_ref"),
		EnvTemplateRef:          installerMetadataValue(metadata, "env_template_ref"),
		ServiceReceiptRef:       installerMetadataValue(metadata, "service_receipt_ref"),
		HeartbeatValidatorRef:   installerMetadataValue(metadata, "heartbeat_validator_ref"),
		DataArrivalValidatorRef: installerMetadataValue(metadata, "data_arrival_validator_ref"),
		AuditRef:                installerMetadataValue(metadata, "audit_ref"),
		EvidenceChainRef:        installerMetadataValue(metadata, "evidence_chain_ref"),
		RunnerWhitelistRef:      installerMetadataValue(metadata, "runner_whitelist_ref"),
		ReloadReceiptRef:        installerMetadataValue(metadata, "reload_receipt_ref"),
		ServiceStatusReceiptRef: installerMetadataValue(metadata, "service_status_receipt_ref"),
	}
}

func safeLinuxInstallerMetadata(input map[string]string) map[string]string {
	out := safeAgentLifecycleMetadata(input)
	if runner, ok := out["runner"]; ok {
		switch normalizeSafeLinuxInstallerRunner(runner) {
		case "local", "systemd", "ssh":
			out["runner"] = normalizeSafeLinuxInstallerRunner(runner)
		default:
			delete(out, "runner")
		}
	}
	return out
}

func normalizeSafeLinuxInstallerRunner(value string) string {
	clean := strings.ToLower(strings.TrimSpace(value))
	switch clean {
	case "ssh":
		return "ssh"
	case "local", "controlled-local":
		return "local"
	case "systemd", "local-systemd", "linux-systemd":
		return "systemd"
	default:
		return ""
	}
}

func installerMetadataValue(metadata map[string]string, key string) string {
	if metadata == nil {
		return ""
	}
	value := strings.TrimSpace(metadata[key])
	if value == "" || looksSensitive(value) {
		return ""
	}
	return value
}

func blockedInstallExecutionSteps(reason string, now time.Time) []model.FindXAgentInstallExecutionStep {
	summary := sanitizeInstallExecutionSummary(reason)
	names := []string{
		"preflight",
		"download_package",
		"verify_package",
		"register_service",
		"enable_or_start_service",
		"verify_heartbeat",
		"verify_data_arrival",
		"capture_evidence",
	}
	steps := make([]model.FindXAgentInstallExecutionStep, 0, len(names))
	for _, name := range names {
		steps = append(steps, model.FindXAgentInstallExecutionStep{
			Name:         name,
			Status:       model.FindXAgentExecutionStatusBlocked,
			ErrorSummary: summary,
			UpdatedAt:    now,
		})
	}
	return steps
}

func firstCleanAgentLifecycleValue(values []string) string {
	for _, value := range values {
		if clean := strings.TrimSpace(value); clean != "" {
			return clean
		}
	}
	return ""
}

func sanitizeInstallExecutionSummary(value string) string {
	clean := strings.TrimSpace(removeControlRunes(value))
	if clean == "" || looksSensitive(clean) {
		return ""
	}
	const maxInstallExecutionSummaryLen = 500
	runes := []rune(clean)
	if len(runes) > maxInstallExecutionSummaryLen {
		clean = string(runes[:maxInstallExecutionSummaryLen])
	}
	return clean
}
