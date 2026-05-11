package handler

import (
	"strings"

	"ai-workbench-api/internal/model"
)

var packageRepositoryCriticalTools = []string{"signature_verifier", "checksum_verifier", "archive_extractor", "service_manager", "plugin_config_writer", "plugin_reloader"}

type packageRepositoryFieldRef struct {
	field string
	name  string
}

type packageRepositoryInstallerFieldRef struct {
	field    string
	name     string
	platform string
}

var packageRepositoryCommandRefs = []packageRepositoryFieldRef{
	{field: "install_command_ref", name: "INSTALL_COMMAND_REF_MISSING"},
	{field: "uninstall_command_ref", name: "UNINSTALL_COMMAND_REF_MISSING"},
	{field: "config_command_ref", name: "CONFIG_COMMAND_REF_MISSING"},
	{field: "plugin_command_ref", name: "PLUGIN_COMMAND_REF_MISSING"},
	{field: "service_manifest_ref", name: "SERVICE_MANIFEST_REF_MISSING"},
	{field: "rollback_manifest_ref", name: "ROLLBACK_MANIFEST_REF_MISSING"},
	{field: "data_arrival_validator_ref", name: "DATA_ARRIVAL_VALIDATOR_REF_MISSING"},
}

var packageRepositoryPrereqFields = []packageRepositoryFieldRef{
	{field: "package_repository_ref", name: "PACKAGE_REPOSITORY_REF_MISSING"},
	{field: "signature_ref", name: "SIGNATURE_REF_MISSING"},
	{field: "checksum", name: "CHECKSUM_MISSING"},
	{field: "script_manifest_ref", name: "SCRIPT_MANIFEST_REF_MISSING"},
	{field: "executor_ref", name: "EXECUTOR_REF_MISSING"},
	{field: "safety_policy_path", name: "SAFETY_POLICY_PATH_MISSING"},
	{field: "runner", name: "RUNNER_MISSING"},
	{field: "ssh_host_key", name: "SSH_HOST_KEY_MISSING"},
	{field: "ssh_fingerprint", name: "SSH_FINGERPRINT_MISSING"},
	{field: "systemd_unit_ref", name: "SYSTEMD_UNIT_REF_MISSING"},
	{field: "curl_installer_ref", name: "CURL_INSTALLER_REF_MISSING"},
	{field: "env_template_ref", name: "ENV_TEMPLATE_REF_MISSING"},
	{field: "windows_installer_ref", name: "WINDOWS_INSTALLER_REF_MISSING"},
	{field: "install_root_policy_ref", name: "INSTALL_ROOT_POLICY_REF_MISSING"},
	{field: "rollback_manifest_ref", name: "ROLLBACK_MANIFEST_REF_MISSING"},
	{field: "uninstall_manifest_ref", name: "UNINSTALL_MANIFEST_REF_MISSING"},
	{field: "receiver_endpoint_ref", name: "RECEIVER_ENDPOINT_REF_MISSING"},
	{field: "cluster_ref", name: "CLUSTER_REF_MISSING"},
	{field: "namespace_ref", name: "NAMESPACE_REF_MISSING"},
	{field: "workload_selector_ref", name: "WORKLOAD_SELECTOR_REF_MISSING"},
	{field: "helm_chart_ref", name: "HELM_CHART_REF_MISSING"},
	{field: "manifest_bundle_ref", name: "MANIFEST_BUNDLE_REF_MISSING"},
	{field: "values_ref", name: "VALUES_REF_MISSING"},
	{field: "rbac_ref", name: "RBAC_REF_MISSING"},
	{field: "service_account_ref", name: "SERVICE_ACCOUNT_REF_MISSING"},
	{field: "image_ref", name: "IMAGE_REF_MISSING"},
	{field: "config_map_ref", name: "CONFIG_MAP_REF_MISSING"},
	{field: "secret_ref", name: "SECRET_REF_MISSING"},
	{field: "rollout_strategy_ref", name: "ROLLOUT_STRATEGY_REF_MISSING"},
	{field: "rollout_receipt_ref", name: "ROLLOUT_RECEIPT_REF_MISSING"},
	{field: "data_arrival_validator_ref", name: "DATA_ARRIVAL_VALIDATOR_REF_MISSING"},
	{field: "audit_ref", name: "AUDIT_REF_MISSING"},
	{field: "evidence_chain_ref", name: "EVIDENCE_CHAIN_REF_MISSING"},
	{field: "method", name: "METHOD_MISSING"},
	{field: "os", name: "OS_MISSING"},
}

var packageRepositoryInstallerPrereqFields = []packageRepositoryInstallerFieldRef{
	{field: "package_repository_ref", name: "PACKAGE_REPOSITORY_REF_MISSING"},
	{field: "signature_ref", name: "SIGNATURE_REF_MISSING"},
	{field: "checksum", name: "CHECKSUM_MISSING"},
	{field: "script_manifest_ref", name: "SCRIPT_MANIFEST_REF_MISSING"},
	{field: "executor_ref", name: "EXECUTOR_REF_MISSING"},
	{field: "service_manifest_ref", name: "SERVICE_MANIFEST_REF_MISSING"},
	{field: "install_root_policy_ref", name: "INSTALL_ROOT_POLICY_REF_MISSING"},
	{field: "rollback_manifest_ref", name: "ROLLBACK_MANIFEST_REF_MISSING"},
	{field: "uninstall_manifest_ref", name: "UNINSTALL_MANIFEST_REF_MISSING"},
	{field: "receiver_endpoint_ref", name: "RECEIVER_ENDPOINT_REF_MISSING"},
	{field: "data_arrival_validator_ref", name: "DATA_ARRIVAL_VALIDATOR_REF_MISSING"},
	{field: "audit_ref", name: "AUDIT_REF_MISSING"},
	{field: "evidence_chain_ref", name: "EVIDENCE_CHAIN_REF_MISSING"},
	{field: "method", name: "METHOD_MISSING"},
	{field: "os", name: "OS_MISSING"},
	{field: "safety_policy_path", name: "SAFETY_POLICY_PATH_MISSING", platform: "linux"},
	{field: "runner", name: "RUNNER_MISSING", platform: "linux"},
	{field: "ssh_host_key", name: "SSH_HOST_KEY_MISSING", platform: "linux"},
	{field: "ssh_fingerprint", name: "SSH_FINGERPRINT_MISSING", platform: "linux"},
	{field: "systemd_unit_ref", name: "SYSTEMD_UNIT_REF_MISSING", platform: "linux"},
	{field: "curl_installer_ref", name: "CURL_INSTALLER_REF_MISSING", platform: "linux"},
	{field: "env_template_ref", name: "ENV_TEMPLATE_REF_MISSING", platform: "linux"},
	{field: "windows_installer_ref", name: "WINDOWS_INSTALLER_REF_MISSING", platform: "windows"},
}

func agentPackageInstallEnvironment(packageID string) model.FindXAgentInstallEnvironment {
	diagnosticBlockers := []string{}
	for _, root := range agentPackageRepositoryRoots() {
		diagnostics := diagnosePackageRepositoryManifest(root)
		if !diagnostics.Ready {
			diagnosticBlockers = append(diagnosticBlockers, diagnostics.Blockers...)
			continue
		}
		manifest := diagnostics.Manifest
		if !manifestHasPackageArtifact(manifest, packageID) {
			continue
		}
		blockers := packageRepositoryManifestDiagnosticBlockers(root, packageID)
		blockers = append(blockers, packageRepositoryTrustChainDiagnosticBlockers(root, packageID)...)
		if len(blockers) > 0 {
			return blockedInstallEnvironmentWithRepositoryEvidence(root, manifest, blockers)
		}
		return packageRepositoryInstallEnvironment(root, manifest)
	}
	if containsString(diagnosticBlockers, "PACKAGE_REPOSITORY_MANIFEST_INVALID") {
		return blockedManifestDiagnosticInstallEnvironment(diagnosticBlockers)
	}
	return blockedInstallEnvironment(append(diagnosticBlockers, []string{
		"INSTALL_ENVIRONMENT_MANIFEST_MISSING",
		"REQUIRED_TOOLS_MISSING",
		"BUNDLED_TOOLS_MISSING",
	}...))
}

func blockedInstallEnvironmentWithRepositoryEvidence(root string, manifest packageRepositoryManifest, blockers []string) model.FindXAgentInstallEnvironment {
	environment := packageRepositoryInstallEnvironment(root, manifest)
	environment.Status = "blocked"
	environment.Blockers = uniquePackageRepositoryBlockers(append(environment.Blockers, blockers...))
	return environment
}

func packageRepositoryInstallEnvironment(root string, manifest packageRepositoryManifest) model.FindXAgentInstallEnvironment {
	return packageRepositoryInstallEnvironmentForOS(root, manifest, "")
}

func packageRepositoryInstallEnvironmentForOS(root string, manifest packageRepositoryManifest, osName string) model.FindXAgentInstallEnvironment {
	tools, blockers := packageRepositoryToolStatus(root, manifest, osName)
	blockers = append(blockers, packageRepositoryEnvironmentContractBlockers(manifest)...)
	blockers = append(blockers, packageRepositoryPrereqBlockersForOS(manifest, osName)...)
	blockers = append(blockers, packageRepositoryCommandRefBlockers(root, manifest)...)
	if len(blockers) > 0 {
		return model.FindXAgentInstallEnvironment{
			Status:    "blocked",
			Platforms: packageRepositoryToolPlatforms(tools),
			Tools:     tools,
			Blockers:  uniquePackageRepositoryBlockers(blockers),
		}
	}
	return model.FindXAgentInstallEnvironment{
		Status:    "ready",
		Platforms: packageRepositoryToolPlatforms(tools),
		Tools:     tools,
	}
}

func packageRepositoryEnvironmentContractBlockers(manifest packageRepositoryManifest) []string {
	if len(manifest.RequiredTools) == 0 && len(manifest.BundledTools) == 0 {
		return []string{"INSTALL_ENVIRONMENT_MANIFEST_MISSING"}
	}
	return nil
}

func blockedInstallEnvironment(blockers []string) model.FindXAgentInstallEnvironment {
	tools := make([]model.FindXAgentInstallToolEvidence, 0, len(packageRepositoryCriticalTools))
	for _, name := range packageRepositoryCriticalTools {
		tools = append(tools, model.FindXAgentInstallToolEvidence{
			Name:     name,
			Required: true,
			Status:   "blocked",
			Blocker:  name + "_MISSING",
		})
	}
	return model.FindXAgentInstallEnvironment{
		Status:   "blocked",
		Tools:    tools,
		Blockers: uniquePackageRepositoryBlockers(blockers),
	}
}

func blockedManifestDiagnosticInstallEnvironment(blockers []string) model.FindXAgentInstallEnvironment {
	return model.FindXAgentInstallEnvironment{
		Status:   "blocked",
		Blockers: uniquePackageRepositoryBlockers(blockers),
	}
}

func packageRepositoryToolStatus(root string, manifest packageRepositoryManifest, osName string) ([]model.FindXAgentInstallToolEvidence, []string) {
	tools := make([]model.FindXAgentInstallToolEvidence, 0, len(packageRepositoryCriticalTools))
	blockers := []string{}
	for _, name := range packageRepositoryCriticalTools {
		requiredTools := packageRepositoryToolsByName(manifest.RequiredTools, name, osName)
		if len(requiredTools) == 0 {
			tools = append(tools, blockedToolEvidence(name, "", "", name+"_REQUIRED_TOOL_MISSING"))
			blockers = append(blockers, name+"_REQUIRED_TOOL_MISSING")
			continue
		}
		for _, required := range requiredTools {
			bundled, ok := packageRepositoryBundledToolForRequired(manifest.BundledTools, required)
			if !ok || !toolEvidenceExists(root, bundled) {
				tools = append(tools, blockedToolEvidence(name, required.OS, required.Arch, name+"_BUNDLED_TOOL_EVIDENCE_MISSING"))
				blockers = append(blockers, name+"_BUNDLED_TOOL_EVIDENCE_MISSING")
				continue
			}
			tools = append(tools, readyToolEvidence(required, bundled))
		}
	}
	return tools, blockers
}

func packageRepositoryToolsByName(tools []packageRepositoryToolEvidence, name, osName string) []packageRepositoryToolEvidence {
	expectedOS := strings.ToLower(strings.TrimSpace(osName))
	matches := []packageRepositoryToolEvidence{}
	for _, tool := range tools {
		toolOS := strings.ToLower(strings.TrimSpace(tool.OS))
		if strings.EqualFold(strings.TrimSpace(tool.Name), name) && (expectedOS == "" || toolOS == expectedOS) {
			matches = append(matches, tool)
		}
	}
	return matches
}

func packageRepositoryBundledToolForRequired(tools []packageRepositoryToolEvidence, required packageRepositoryToolEvidence) (packageRepositoryToolEvidence, bool) {
	for _, tool := range tools {
		if strings.EqualFold(strings.TrimSpace(tool.Name), strings.TrimSpace(required.Name)) &&
			strings.EqualFold(strings.TrimSpace(tool.OS), strings.TrimSpace(required.OS)) &&
			strings.EqualFold(strings.TrimSpace(tool.Arch), strings.TrimSpace(required.Arch)) {
			return tool, true
		}
	}
	return packageRepositoryToolEvidence{}, false
}

func toolEvidenceExists(root string, tool packageRepositoryToolEvidence) bool {
	ref := strings.TrimSpace(tool.EvidenceRef)
	if ref == "" {
		ref = tool.Path
	}
	return safePackageRepositoryFileExists(root, ref)
}

func blockedToolEvidence(name, osName, arch, blocker string) model.FindXAgentInstallToolEvidence {
	return model.FindXAgentInstallToolEvidence{
		Name:     name,
		OS:       osName,
		Arch:     arch,
		Required: true,
		Status:   "blocked",
		Blocker:  blocker,
	}
}

func readyToolEvidence(required, bundled packageRepositoryToolEvidence) model.FindXAgentInstallToolEvidence {
	ref := strings.TrimSpace(bundled.EvidenceRef)
	if ref == "" {
		ref = bundled.Path
	}
	return model.FindXAgentInstallToolEvidence{
		Name:        strings.TrimSpace(required.Name),
		OS:          strings.TrimSpace(required.OS),
		Arch:        strings.TrimSpace(required.Arch),
		Required:    true,
		Bundled:     true,
		EvidenceRef: ref,
		Status:      "ready",
	}
}

func packageRepositoryCommandRefBlockers(root string, manifest packageRepositoryManifest) []string {
	refs := map[string]string{
		"install_command_ref":        manifest.InstallCommandRef,
		"uninstall_command_ref":      manifest.UninstallCommandRef,
		"config_command_ref":         manifest.ConfigCommandRef,
		"plugin_command_ref":         manifest.PluginCommandRef,
		"service_manifest_ref":       manifest.ServiceManifestRef,
		"rollback_manifest_ref":      manifest.RollbackManifestRef,
		"data_arrival_validator_ref": manifest.DataArrivalValidatorRef,
	}
	blockers := []string{}
	for _, item := range packageRepositoryCommandRefs {
		if !safePackageRepositoryFileExists(root, refs[item.field]) {
			blockers = append(blockers, item.name)
		}
	}
	return blockers
}

func packageRepositoryPrereqBlockers(manifest packageRepositoryManifest) []string {
	return packageRepositoryPrereqBlockersFromFields(manifest, packageRepositoryPrereqFields)
}

func packageRepositoryPrereqBlockersForOS(manifest packageRepositoryManifest, osName string) []string {
	osName = strings.ToLower(strings.TrimSpace(osName))
	if osName == "" {
		return packageRepositoryPrereqBlockers(manifest)
	}
	fields := make([]packageRepositoryFieldRef, 0, len(packageRepositoryInstallerPrereqFields))
	for _, item := range packageRepositoryInstallerPrereqFields {
		if item.platform == "" || item.platform == osName {
			fields = append(fields, packageRepositoryFieldRef{field: item.field, name: item.name})
		}
	}
	return packageRepositoryPrereqBlockersFromFields(manifest, fields)
}

func packageRepositoryPrereqBlockersFromFields(manifest packageRepositoryManifest, fields []packageRepositoryFieldRef) []string {
	values := packageRepositoryPrereqFieldValues(manifest)
	blockers := []string{}
	for _, item := range fields {
		if strings.TrimSpace(values[item.field]) == "" {
			blockers = append(blockers, item.name)
		}
	}
	return blockers
}

func packageRepositoryPrereqFieldValues(manifest packageRepositoryManifest) map[string]string {
	values := map[string]string{
		"package_repository_ref":     manifest.PackageRepositoryRef,
		"signature_ref":              manifest.SignatureRef,
		"checksum":                   manifest.Checksum,
		"script_manifest_ref":        manifest.ScriptManifestRef,
		"executor_ref":               manifest.ExecutorRef,
		"safety_policy_path":         manifest.SafetyPolicyPath,
		"runner":                     manifest.Runner,
		"ssh_host_key":               manifest.SSHHostKey,
		"ssh_fingerprint":            manifest.SSHFingerprint,
		"systemd_unit_ref":           manifest.SystemdUnitRef,
		"curl_installer_ref":         manifest.CurlInstallerRef,
		"env_template_ref":           manifest.EnvTemplateRef,
		"windows_installer_ref":      manifest.WindowsInstallerRef,
		"service_manifest_ref":       manifest.ServiceManifestRef,
		"install_root_policy_ref":    manifest.InstallRootPolicyRef,
		"rollback_manifest_ref":      manifest.RollbackManifestRef,
		"uninstall_manifest_ref":     manifest.UninstallManifestRef,
		"receiver_endpoint_ref":      manifest.ReceiverEndpointRef,
		"cluster_ref":                manifest.ClusterRef,
		"namespace_ref":              manifest.NamespaceRef,
		"workload_selector_ref":      manifest.WorkloadSelectorRef,
		"helm_chart_ref":             manifest.HelmChartRef,
		"manifest_bundle_ref":        manifest.ManifestBundleRef,
		"values_ref":                 manifest.ValuesRef,
		"rbac_ref":                   manifest.RBACRef,
		"service_account_ref":        manifest.ServiceAccountRef,
		"image_ref":                  manifest.ImageRef,
		"config_map_ref":             manifest.ConfigMapRef,
		"secret_ref":                 manifest.SecretRef,
		"rollout_strategy_ref":       manifest.RolloutStrategyRef,
		"rollout_receipt_ref":        manifest.RolloutReceiptRef,
		"data_arrival_validator_ref": manifest.DataArrivalValidatorRef,
		"audit_ref":                  manifest.AuditRef,
		"evidence_chain_ref":         manifest.EvidenceChainRef,
		"method":                     manifest.Method,
		"os":                         manifest.OS,
	}
	return values
}

func packageRepositoryToolPlatforms(tools []model.FindXAgentInstallToolEvidence) []string {
	platforms := []string{}
	for _, tool := range tools {
		if tool.Status == "ready" && strings.TrimSpace(tool.OS) != "" {
			platforms = appendPackageRepositoryPlatform(platforms, tool.OS, tool.Arch)
		}
	}
	return platforms
}

func appendPackageRepositoryPlatform(platforms []string, osName, arch string) []string {
	platform := strings.TrimSpace(osName)
	if strings.TrimSpace(arch) != "" {
		platform += "/" + strings.TrimSpace(arch)
	}
	for _, existing := range platforms {
		if strings.EqualFold(existing, platform) {
			return platforms
		}
	}
	return append(platforms, platform)
}

func uniquePackageRepositoryBlockers(values []string) []string {
	seen := map[string]bool{}
	unique := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}
