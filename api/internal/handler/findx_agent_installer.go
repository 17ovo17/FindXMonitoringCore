package handler

import (
	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/security"
	"ai-workbench-api/internal/store"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
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


func isWindowsInstallerInstallPlan(req model.FindXAgentInstallPlanRequest) bool {
	if !strings.EqualFold(strings.TrimSpace(req.OS), "windows") {
		return false
	}
	method := strings.ToLower(strings.TrimSpace(req.Method))
	return strings.Contains(method, "powershell") ||
		strings.Contains(method, "invoke-webrequest") ||
		strings.EqualFold(method, "cmd") ||
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
		Status:       "accepted",
		Steps:        pendingWindowsInstallExecutionSteps(now),
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
		"data":      safeWindowsInstallPlanResponse(plan),
		"execution": saved,
	})
}

func pendingWindowsInstallExecutionSteps(now time.Time) []model.FindXAgentInstallExecutionStep {
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

func safeWindowsInstallPlanResponse(plan model.FindXAgentInstallPlan) model.FindXAgentInstallPlan {
	out := plan
	if looksSuspiciousWindowsInstallerMethod(out.Method) {
		out.Method = ""
	}
	out.Metadata = safeWindowsInstallerMetadata(out.Metadata)
	return out
}

func safeWindowsInstallerMetadata(input map[string]string) map[string]string {
	out := safeLinuxInstallerMetadata(input)
	delete(out, "custom_runner")
	for _, key := range []string{"runner", "transport", "method"} {
		value := out[key]
		if value == "" {
			continue
		}
		if looksSuspiciousWindowsInstallerMethod(value) || !isAllowedWindowsInstallerMetadataValue(value) {
			delete(out, key)
		}
	}
	return out
}

func isAllowedWindowsInstallerMetadataValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(removeControlRunes(value))) {
	case "windows-powershell", "windows-cmd", "windows-installer", "local", "windows_local", "windows_service":
		return true
	default:
		return false
	}
}

func looksSuspiciousWindowsInstallerMethod(value string) bool {
	clean := strings.ToLower(strings.TrimSpace(removeControlRunes(value)))
	if clean == "" {
		return false
	}
	for _, marker := range []string{"&&", "||", ";", "`", "|", ">", "<", "$(", "%comspec%", "cmd /c", "powershell -", " start-process ", " invoke-expression "} {
		if strings.Contains(clean, marker) {
			return true
		}
	}
	return looksSensitive(clean)
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
	execution := model.FindXAgentInstallExecution{
		PlanID:       plan.ID,
		TargetID:     targetID,
		Runner:       gate.Runner,
		Status:       "accepted",
		Steps:        pendingKubernetesInstallExecutionSteps(now),
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

func pendingKubernetesInstallExecutionSteps(now time.Time) []model.FindXAgentInstallExecutionStep {
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
			Name:      name,
			Status:    "pending",
			UpdatedAt: now,
		})
	}
	return steps
}

func kubernetesInstallExecutionReceiptScope(runner string) string {
	switch strings.ToLower(strings.TrimSpace(runner)) {
	case "helm":
		return "helm"
	case "kubernetes-daemonset":
		return "daemonset"
	case "kubernetes-sidecar":
		return "sidecar"
	case "kubernetes-initcontainer":
		return "initcontainer"
	default:
		return "kubernetes"
	}
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
		Status:       "accepted",
		Steps:        pendingRemoteInstallExecutionSteps(gate.Scope, now),
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
		"data":      remoteInstallPlanResponse(plan, gate.Scope),
		"execution": saved,
	})
}

func pendingRemoteInstallExecutionSteps(scope string, now time.Time) []model.FindXAgentInstallExecutionStep {
	names := []string{
		"preflight",
		scope + "_transport",
		"remote_executor",
		"dispatch",
		"service_registration",
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
