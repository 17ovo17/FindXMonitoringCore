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

func isKubernetesInstallerInstallPlan(req model.FindXAgentInstallPlanRequest) bool {
	method := strings.ToLower(strings.TrimSpace(req.Method))
	if strings.EqualFold(strings.TrimSpace(req.OS), "kubernetes") {
		return true
	}
	for _, marker := range []string{"helm", "kubernetes", "k8s", "daemonset", "sidecar", "initcontainer", "init-container"} {
		if strings.Contains(method, marker) {
			return true
		}
	}
	return false
}

func createBlockedFindXAgentKubernetesInstallExecution(c *gin.Context, req model.FindXAgentInstallPlanRequest, plan model.FindXAgentInstallPlan) {
	targetID := firstCleanAgentLifecycleValue(plan.TargetIDs)
	gate := security.EvaluateKubernetesInstallerPrerequisites(kubernetesInstallerPrerequisitesFromRequest(req))
	now := time.Now()
	responseError := sanitizeKubernetesInstallExecutionSummary(gate.Reason)
	execution := model.FindXAgentInstallExecution{
		PlanID:       plan.ID,
		TargetID:     targetID,
		Runner:       gate.Runner,
		Status:       gate.Status,
		Steps:        blockedKubernetesInstallExecutionSteps(gate.Reason, now),
		EvidenceRefs: []string{"install-plan:" + plan.ID, "blocked-contract:kubernetes-lifecycle"},
		ErrorSummary: sanitizeKubernetesInstallExecutionSummary(gate.Reason),
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
		"error":     responseError,
		"data":      plan,
		"execution": saved,
	})
}

func kubernetesInstallerPrerequisitesFromRequest(req model.FindXAgentInstallPlanRequest) security.KubernetesInstallerPrerequisites {
	metadata := req.Metadata
	return security.KubernetesInstallerPrerequisites{
		CredentialRefPresent:    strings.TrimSpace(req.CredentialRef) != "",
		ClusterRef:              installerMetadataValue(metadata, "cluster_ref"),
		NamespaceRef:            installerMetadataValue(metadata, "namespace_ref"),
		WorkloadSelectorRef:     installerMetadataValue(metadata, "workload_selector_ref"),
		HelmChartRef:            installerMetadataValue(metadata, "helm_chart_ref"),
		ManifestBundleRef:       installerMetadataValue(metadata, "manifest_bundle_ref"),
		ValuesRef:               installerMetadataValue(metadata, "values_ref"),
		RBACRef:                 installerMetadataValue(metadata, "rbac_ref"),
		ServiceAccountRef:       installerMetadataValue(metadata, "service_account_ref"),
		ImageRef:                installerMetadataValue(metadata, "image_ref"),
		PackageRepositoryRef:    installerMetadataValue(metadata, "package_repository_ref"),
		SignatureRef:            installerMetadataValue(metadata, "signature_ref"),
		Checksum:                installerMetadataValue(metadata, "checksum"),
		ConfigMapRef:            installerMetadataValue(metadata, "config_map_ref"),
		SecretRef:               installerMetadataValue(metadata, "secret_ref"),
		RolloutStrategyRef:      installerMetadataValue(metadata, "rollout_strategy_ref"),
		RolloutReceiptRef:       installerMetadataValue(metadata, "rollout_receipt_ref"),
		DataArrivalValidatorRef: installerMetadataValue(metadata, "data_arrival_validator_ref"),
		ExecutorRef:             installerMetadataValue(metadata, "executor_ref"),
		AuditRef:                installerMetadataValue(metadata, "audit_ref"),
		EvidenceChainRef:        installerMetadataValue(metadata, "evidence_chain_ref"),
		Method:                  req.Method,
		OS:                      req.OS,
	}
}

func blockedKubernetesInstallExecutionSteps(reason string, now time.Time) []model.FindXAgentInstallExecutionStep {
	summary := sanitizeKubernetesInstallExecutionSummary(reason)
	names := []string{
		"resolve_cluster",
		"validate_namespace",
		"verify_package",
		"verify_rbac",
		"render_manifest",
		"prepare_rollout",
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

func sanitizeKubernetesInstallExecutionSummary(value string) string {
	clean := strings.TrimSpace(removeControlRunes(value))
	if clean == "" {
		return ""
	}
	const maxKubernetesInstallExecutionSummaryLen = 500
	runes := []rune(clean)
	if len(runes) > maxKubernetesInstallExecutionSummaryLen {
		clean = string(runes[:maxKubernetesInstallExecutionSummaryLen])
	}
	return clean
}
