package handler

import (
	"strings"

	"ai-workbench-api/internal/model"
)

const (
	environmentToolMissing        = "TOOL_EVIDENCE_MISSING"
	environmentPackageMissing     = "PACKAGE_MISSING"
	environmentSourceMissing      = "LOCAL_SOURCE_MISSING"
	environmentTestOnlyRepository = "PACKAGE_REPOSITORY_TEST_ONLY"
	environmentTestOnlySignature  = "SIGNATURE_TEST_ONLY"
)

type packageEnvironmentMethod struct {
	platform string
	method   string
}

var packageEnvironmentMethods = []packageEnvironmentMethod{
	{platform: "linux", method: "curl -kfsSL"},
	{platform: "windows", method: "certutil -urlcache -f"},
	{platform: "windows", method: "Invoke-WebRequest"},
	{platform: "linux", method: "SSH"},
	{platform: "windows", method: "WinRM"},
	{platform: "kubernetes", method: "Helm"},
	{platform: "kubernetes", method: "Operator"},
	{platform: "kubernetes", method: "DaemonSet"},
	{platform: "kubernetes", method: "Sidecar"},
	{platform: "kubernetes", method: "InitContainer"},
}

func agentPackageEnvironmentMatrix(
	def agentPackageDef,
	environment model.FindXAgentInstallEnvironment,
	blockers []string,
	sourceState string,
) []model.FindXAgentPackageEnvironmentMatrixRow {
	matrix := make([]model.FindXAgentPackageEnvironmentMatrixRow, 0, len(packageEnvironmentMethods))
	for _, method := range packageEnvironmentMethods {
		row := packageEnvironmentRow(def, environment, blockers, sourceState, method)
		matrix = append(matrix, row)
	}
	return matrix
}

func packageEnvironmentRow(
	def agentPackageDef,
	environment model.FindXAgentInstallEnvironment,
	blockers []string,
	sourceState string,
	method packageEnvironmentMethod,
) model.FindXAgentPackageEnvironmentMatrixRow {
	state := packageEnvironmentSourceState(def.id, sourceState)
	return model.FindXAgentPackageEnvironmentMatrixRow{
		Platform:            method.platform,
		InstallMethod:       method.method,
		ToolEvidence:        packageEnvironmentToolEvidence(environment, method.platform),
		SourceState:         state,
		PackageState:        packageEnvironmentPackageState(def.id, blockers, state),
		Executor:            packageEnvironmentBlocked("EXECUTOR_CONTRACT_MISSING", blockers),
		ServiceRegistration: packageEnvironmentBlocked("SERVICE_REGISTRATION_CONTRACT_MISSING", blockers),
		ConfigDelivery:      packageEnvironmentConfigDelivery(def.pluginConfig != nil),
		Uninstall:           packageEnvironmentBlocked("UNINSTALL_CONTRACT_MISSING", blockers),
		Rollback:            packageEnvironmentBlocked("ROLLBACK_CONTRACT_MISSING", blockers),
		DataArrival:         packageEnvironmentBlocked("DATA_ARRIVAL_CONTRACT_MISSING", blockers),
		Blocker:             packageEnvironmentBlocker(def.id, blockers, state),
	}
}

func packageEnvironmentSourceState(packageID, sourceState string) string {
	if isProbePackageWithoutLocalSource(packageID) {
		return environmentSourceMissing
	}
	if strings.TrimSpace(sourceState) == "" {
		return environmentSourceMissing
	}
	return sourceState
}

func isProbePackageWithoutLocalSource(packageID string) bool {
	return strings.HasSuffix(packageID, "-app") ||
		packageID == "gateway-probe" ||
		packageID == "browser-client"
}

func packageEnvironmentToolEvidence(environment model.FindXAgentInstallEnvironment, platform string) string {
	if packageEnvironmentHasReadyTool(environment, platform) &&
		containsLifecycleString(environment.Blockers, environmentTestOnlyRepository) {
		return strings.Join([]string{
			environmentTestOnlyRepository,
			environmentTestOnlySignature,
			"TOOL_EVIDENCE_TEST_ONLY",
		}, ";")
	}
	return packageEnvironmentBlocked(environmentToolMissing, environment.Blockers)
}

func packageEnvironmentHasReadyTool(environment model.FindXAgentInstallEnvironment, platform string) bool {
	for _, tool := range environment.Tools {
		if strings.EqualFold(tool.Status, "ready") && strings.EqualFold(tool.OS, platform) {
			return true
		}
	}
	return false
}

func packageEnvironmentPackageState(packageID string, blockers []string, sourceState string) string {
	if sourceState == environmentSourceMissing || isProbePackageWithoutLocalSource(packageID) {
		return environmentPackageMissing
	}
	if containsLifecycleString(blockers, environmentTestOnlyRepository) {
		return strings.Join([]string{
			environmentTestOnlyRepository,
			environmentTestOnlySignature,
			"PRODUCTION_PACKAGE_REPOSITORY_MISSING",
			"PRODUCTION_SIGNATURE_MISSING",
		}, ";")
	}
	return packageEnvironmentBlocked(environmentPackageMissing, blockers)
}

func packageEnvironmentConfigDelivery(hasPluginConfig bool) string {
	if !hasPluginConfig {
		return packageEnvironmentBlocked("CONFIG_DELIVERY_CONTRACT_MISSING", nil)
	}
	return strings.Join([]string{
		"FINDX_AGENT_CONTROL_PLANE_ENTRY",
		"REMOTE_MUTATION_PENDING",
		"RELOAD_PENDING",
		"DRIFT_PENDING",
		"ROLLBACK_PENDING",
		"RECEIPT_PENDING",
	}, ";")
}

func packageEnvironmentBlocker(packageID string, blockers []string, sourceState string) string {
	if sourceState == environmentSourceMissing || isProbePackageWithoutLocalSource(packageID) {
		return agentBlocked + ": LOCAL_SOURCE_MISSING;PACKAGE_MISSING"
	}
	return packageEnvironmentBlocked("ENVIRONMENT_MATRIX_PENDING", blockers)
}

func packageEnvironmentBlocked(reason string, blockers []string) string {
	values := []string{agentBlocked}
	values = append(values, reason)
	values = append(values, blockers...)
	return strings.Join(uniquePackageRepositoryBlockers(values), ":")
}

func containsLifecycleString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
