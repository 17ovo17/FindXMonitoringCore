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
		Metadata:             safeAgentLifecycleMetadata(req.Metadata),
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
		Status:       gate.Status,
		Steps:        blockedInstallExecutionSteps(gate.Reason, now),
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
	c.JSON(http.StatusConflict, gin.H{
		"code":      http.StatusConflict,
		"error":     saved.ErrorSummary,
		"data":      plan,
		"execution": saved,
	})
}

func linuxInstallerPrerequisitesFromRequest(req model.FindXAgentInstallPlanRequest) security.LinuxInstallerPrerequisites {
	metadata := req.Metadata
	return security.LinuxInstallerPrerequisites{
		PackageRepositoryRef: installerMetadataValue(metadata, "package_repository_ref"),
		SignatureRef:         installerMetadataValue(metadata, "signature_ref"),
		Checksum:             installerMetadataValue(metadata, "checksum"),
		ScriptManifestRef:    installerMetadataValue(metadata, "script_manifest_ref"),
		ExecutorRef:          installerMetadataValue(metadata, "executor_ref"),
		SafetyPolicyPath:     installerMetadataValue(metadata, "safety_policy_path"),
		Runner:               installerMetadataValue(metadata, "runner"),
		SSHHostKey:           installerMetadataValue(metadata, "ssh_host_key"),
		SSHFingerprint:       installerMetadataValue(metadata, "ssh_fingerprint"),
		SystemdUnitRef:       installerMetadataValue(metadata, "systemd_unit_ref"),
		CurlInstallerRef:     installerMetadataValue(metadata, "curl_installer_ref"),
		EnvTemplateRef:       installerMetadataValue(metadata, "env_template_ref"),
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
