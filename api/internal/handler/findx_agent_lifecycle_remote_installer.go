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

func isRemoteInstallerInstallPlan(req model.FindXAgentInstallPlanRequest) bool {
	return remoteInstallerScopeFromRequest(req) != ""
}

func remoteInstallerScopeFromRequest(req model.FindXAgentInstallPlanRequest) string {
	for _, value := range []string{req.Method, req.Metadata["transport"], req.Metadata["method"]} {
		if scope := security.NormalizeRemoteInstallerScope(value); scope != "" {
			return scope
		}
	}
	if strings.Contains(strings.ToLower(strings.TrimSpace(req.Method)), "remote") {
		return security.NormalizeRemoteInstallerScope(req.Metadata["runner"])
	}
	return ""
}

func createBlockedFindXAgentRemoteInstallExecution(c *gin.Context, req model.FindXAgentInstallPlanRequest, plan model.FindXAgentInstallPlan) {
	targetID := firstCleanAgentLifecycleValue(plan.TargetIDs)
	gate := security.EvaluateRemoteInstallerPreflight(remoteInstallerPreflightFromRequest(req))
	now := time.Now()
	execution := model.FindXAgentInstallExecution{
		PlanID:       plan.ID,
		TargetID:     targetID,
		Runner:       gate.Runner,
		Status:       gate.Status,
		Steps:        blockedRemoteInstallExecutionSteps(gate.Scope, gate.Reason, now),
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
	missing := installExecutionMissingContracts(gate.Scope, gate.Reason)
	c.JSON(http.StatusConflict, gin.H{
		"code":              http.StatusConflict,
		"error":             saved.ErrorSummary,
		"status":            "blocked",
		"state_machine":     blockedExecutionStateMachine(saved.ErrorSummary),
		"receipt_contract":  installReceiptContract(gate.Scope, req, saved.Runner, missing),
		"receipt_matrix":    findXAgentReceiptContractMatrix(),
		"missing_contracts": missing,
		"safe_to_retry":     false,
		"data":              remoteInstallPlanResponse(plan, gate.Scope),
		"execution":         saved,
	})
}

func remoteInstallerPreflightFromRequest(req model.FindXAgentInstallPlanRequest) security.RemoteInstallerPreflightInput {
	metadata := req.Metadata
	return security.RemoteInstallerPreflightInput{
		Scope:                   remoteInstallerScopeFromRequest(req),
		CredentialRef:           strings.TrimSpace(req.CredentialRef),
		RemoteExecutorRef:       installerMetadataValue(metadata, "remote_executor_ref"),
		SSHRunnerRef:            installerMetadataValue(metadata, "ssh_runner_ref"),
		SSHHostKey:              installerMetadataValue(metadata, "ssh_host_key"),
		SSHFingerprint:          installerMetadataValue(metadata, "ssh_fingerprint"),
		WinRMEndpointRef:        installerMetadataValue(metadata, "winrm_endpoint_ref"),
		WinRMTransportRef:       installerMetadataValue(metadata, "winrm_transport_ref"),
		TimeoutPolicyRef:        installerMetadataValue(metadata, "timeout_policy_ref"),
		IdempotencyKey:          installerMetadataValue(metadata, "idempotency_key"),
		InstallReceiptRef:       installerMetadataValue(metadata, "install_receipt_ref"),
		ExecutionReceiptRef:     installerMetadataValue(metadata, "execution_receipt_ref"),
		ServiceReceiptRef:       installerMetadataValue(metadata, "service_receipt_ref"),
		HeartbeatValidatorRef:   installerMetadataValue(metadata, "heartbeat_validator_ref"),
		DataArrivalValidatorRef: installerMetadataValue(metadata, "data_arrival_validator_ref"),
		AuditRef:                installerMetadataValue(metadata, "audit_ref"),
		EvidenceChainRef:        installerMetadataValue(metadata, "evidence_chain_ref"),
		PackageRepositoryRef:    installerMetadataValue(metadata, "package_repository_ref"),
		SignatureRef:            installerMetadataValue(metadata, "signature_ref"),
		Checksum:                installerMetadataValue(metadata, "checksum"),
		ScriptManifestRef:       installerMetadataValue(metadata, "script_manifest_ref"),
	}
}

func remoteInstallPlanResponse(plan model.FindXAgentInstallPlan, scope string) model.FindXAgentInstallPlan {
	out := plan
	out.Metadata = safeRemoteInstallerMetadata(out.Metadata, scope)
	if strings.TrimSpace(out.Method) != "" {
		out.Method = scope
	}
	return out
}

func safeRemoteInstallerMetadata(input map[string]string, scope string) map[string]string {
	out := safeLinuxInstallerMetadata(input)
	for _, key := range []string{"transport", "runner", "method"} {
		if out[key] == "" {
			continue
		}
		if normalized := security.NormalizeRemoteInstallerScope(out[key]); normalized == scope {
			out[key] = normalized
			continue
		}
		delete(out, key)
	}
	return out
}

func blockedRemoteInstallExecutionSteps(scope string, reason string, now time.Time) []model.FindXAgentInstallExecutionStep {
	summary := sanitizeInstallExecutionSummary(reason)
	names := []string{
		"preflight",
		scope + "_transport_contract",
		"remote_executor_contract",
		"dispatch_receipt_contract",
		"service_receipt_contract",
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
