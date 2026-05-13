package handler

import (
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/security"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func isWindowsInstallerInstallPlan(req model.FindXAgentInstallPlanRequest) bool {
	if !strings.EqualFold(strings.TrimSpace(req.OS), "windows") {
		return false
	}
	method := strings.ToLower(strings.TrimSpace(req.Method))
	return strings.Contains(method, "powershell") ||
		strings.Contains(method, "windows-cmd") ||
		strings.Contains(method, "certutil")
}

func createBlockedFindXAgentWindowsInstallExecution(c *gin.Context, req model.FindXAgentInstallPlanRequest, plan model.FindXAgentInstallPlan) {
	targetID := firstCleanAgentLifecycleValue(plan.TargetIDs)
	gate := security.EvaluateWindowsInstallerPrerequisites(windowsInstallerPrerequisitesFromRequest(req))
	now := time.Now()
	execution := model.FindXAgentInstallExecution{
		PlanID:       plan.ID,
		TargetID:     targetID,
		Runner:       gate.Runner,
		Status:       gate.Status,
		Steps:        blockedWindowsInstallExecutionSteps(gate.Reason, now),
		EvidenceRefs: []string{"install-plan:" + plan.ID},
		ErrorSummary: sanitizeInstallExecutionSummary(gate.Reason),
		StartedAt:    &now,
		FinishedAt:   &now,
	}
	saved, err := store.SaveFindXAgentInstallExecution(execution)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "install execution persistence unavailable"})
		return
	}
	auditEvent(c, "findx_agent.install_execution.blocked", saved.ID, "high", saved.Status, saved.ErrorSummary, c.GetHeader("X-Test-Batch-Id"))
	scope := windowsInstallExecutionReceiptScope(req)
	missing := installExecutionMissingContracts(scope, gate.Reason)
	c.JSON(http.StatusConflict, gin.H{
		"code":              http.StatusConflict,
		"error":             saved.ErrorSummary,
		"status":            "blocked",
		"state_machine":     blockedExecutionStateMachine(saved.ErrorSummary),
		"receipt_contract":  installReceiptContract(scope, req, saved.Runner, missing),
		"receipt_matrix":    findXAgentReceiptContractMatrix(),
		"missing_contracts": missing,
		"safe_to_retry":     false,
		"data":              plan,
		"execution":         saved,
	})
}

func windowsInstallExecutionReceiptScope(req model.FindXAgentInstallPlanRequest) string {
	text := strings.ToLower(strings.TrimSpace(req.Method))
	for _, key := range []string{"service_manifest_ref", "windows_service_name_ref", "windows_service_policy_ref", "service_receipt_ref", "service_status_receipt_ref"} {
		if strings.TrimSpace(req.Metadata[key]) != "" {
			return "windows_service"
		}
	}
	if strings.Contains(text, "service") {
		return "windows_service"
	}
	return "windows_local"
}

func windowsInstallerPrerequisitesFromRequest(req model.FindXAgentInstallPlanRequest) security.WindowsInstallerPrerequisites {
	metadata := req.Metadata
	return security.WindowsInstallerPrerequisites{
		PackageRepositoryRef:    installerMetadataValue(metadata, "package_repository_ref"),
		SignatureRef:            installerMetadataValue(metadata, "signature_ref"),
		Checksum:                installerMetadataValue(metadata, "checksum"),
		ScriptManifestRef:       installerMetadataValue(metadata, "script_manifest_ref"),
		ExecutorRef:             installerMetadataValue(metadata, "executor_ref"),
		WindowsInstallerRef:     installerMetadataValue(metadata, "windows_installer_ref"),
		PowerShellInstallerRef:  installerMetadataValue(metadata, "powershell_installer_ref"),
		CertutilInstallerRef:    installerMetadataValue(metadata, "certutil_installer_ref"),
		WindowsCmdInstallerRef:  installerMetadataValue(metadata, "windows_cmd_installer_ref"),
		ServiceManifestRef:      installerMetadataValue(metadata, "service_manifest_ref"),
		InstallRootPolicyRef:    installerMetadataValue(metadata, "install_root_policy_ref"),
		WindowsServiceNameRef:   installerMetadataValue(metadata, "windows_service_name_ref"),
		WindowsServicePolicyRef: installerMetadataValue(metadata, "windows_service_policy_ref"),
		ServiceReceiptRef:       installerMetadataValue(metadata, "service_receipt_ref"),
		ServiceStatusReceiptRef: installerMetadataValue(metadata, "service_status_receipt_ref"),
		InstallReceiptRef:       installerMetadataValue(metadata, "install_receipt_ref"),
		UninstallManifestRef:    installerMetadataValue(metadata, "uninstall_manifest_ref"),
		UninstallReceiptRef:     installerMetadataValue(metadata, "uninstall_receipt_ref"),
		RollbackManifestRef:     installerMetadataValue(metadata, "rollback_manifest_ref"),
		RollbackReceiptRef:      installerMetadataValue(metadata, "rollback_receipt_ref"),
		HeartbeatValidatorRef:   installerMetadataValue(metadata, "heartbeat_validator_ref"),
		DataArrivalValidatorRef: installerMetadataValue(metadata, "data_arrival_validator_ref"),
		AuditRef:                installerMetadataValue(metadata, "audit_ref"),
		EvidenceChainRef:        installerMetadataValue(metadata, "evidence_chain_ref"),
		ReceiverEndpointRef:     installerMetadataValue(metadata, "receiver_endpoint_ref"),
		Method:                  req.Method,
	}
}

func blockedWindowsInstallExecutionSteps(reason string, now time.Time) []model.FindXAgentInstallExecutionStep {
	summary := sanitizeInstallExecutionSummary(reason)
	names := []string{
		"download_script",
		"download_package",
		"verify_signature",
		"install_files",
		"register_windows_service",
		"start_windows_service",
		"verify_service_status",
		"verify_heartbeat",
		"verify_data_arrival",
		"capture_evidence",
		"rollback_or_cleanup",
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
