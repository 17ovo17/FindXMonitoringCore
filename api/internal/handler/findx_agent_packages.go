package handler

import (
	"ai-workbench-api/internal/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)


func agentPackageBlockers(packageID string) []string {
	if hasAgentPackageTestOnlyRepositoryEvidence(packageID) {
		return uniquePackageRepositoryBlockers(append([]string{
			"PACKAGE_REPOSITORY_TEST_ONLY",
			"SIGNATURE_TEST_ONLY",
			"PRODUCTION_PACKAGE_REPOSITORY_MISSING",
			"PRODUCTION_SIGNATURE_MISSING",
			"INSTALL_PLAN_CONTRACT_MISSING",
			"CONFIG_ROLLOUT_CONTRACT_MISSING",
		}, agentPackageInstallEnvironment(packageID).Blockers...))
	}
	return uniquePackageRepositoryBlockers(append([]string{
		"PACKAGE_REPOSITORY_MISSING",
		"SIGNATURE_MISSING",
		"INSTALL_PLAN_CONTRACT_MISSING",
		"CONFIG_ROLLOUT_CONTRACT_MISSING",
	}, agentPackageInstallEnvironment(packageID).Blockers...))
}


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


const (
	packageRepositoryManifestPath = "manifests/manifest.json"
	packageRepositoryFallbackRoot = "/opt/ai-workbench-runtime/packages/findx-agent"
)

type packageRepositoryManifest struct {
	Repository               string                              `json:"repository"`
	Status                   string                              `json:"status"`
	SignatureScope           string                              `json:"signature_scope"`
	ChecksumFile             string                              `json:"checksum_file"`
	ChecksumSignature        string                              `json:"checksum_signature"`
	PublicKey                string                              `json:"public_key"`
	PublicKeyFingerprintFile string                              `json:"public_key_fingerprint_file"`
	TrustRootRef             string                              `json:"trust_root_ref"`
	TrustRootFingerprintRef  string                              `json:"trust_root_fingerprint_ref"`
	VerificationReceiptRef   string                              `json:"verification_receipt_ref"`
	ChecksumVerificationRef  string                              `json:"checksum_verification_ref"`
	SignatureVerificationRef string                              `json:"signature_verification_ref"`
	PackageRepositoryRef     string                              `json:"package_repository_ref"`
	SignatureRef             string                              `json:"signature_ref"`
	Checksum                 string                              `json:"checksum"`
	ScriptManifestRef        string                              `json:"script_manifest_ref"`
	ExecutorRef              string                              `json:"executor_ref"`
	SafetyPolicyPath         string                              `json:"safety_policy_path"`
	Runner                   string                              `json:"runner"`
	SSHHostKey               string                              `json:"ssh_host_key"`
	SSHFingerprint           string                              `json:"ssh_fingerprint"`
	SystemdUnitRef           string                              `json:"systemd_unit_ref"`
	CurlInstallerRef         string                              `json:"curl_installer_ref"`
	EnvTemplateRef           string                              `json:"env_template_ref"`
	WindowsInstallerRef      string                              `json:"windows_installer_ref"`
	ServiceManifestRef       string                              `json:"service_manifest_ref"`
	InstallRootPolicyRef     string                              `json:"install_root_policy_ref"`
	RollbackManifestRef      string                              `json:"rollback_manifest_ref"`
	UninstallManifestRef     string                              `json:"uninstall_manifest_ref"`
	ReceiverEndpointRef      string                              `json:"receiver_endpoint_ref"`
	Method                   string                              `json:"method"`
	ClusterRef               string                              `json:"cluster_ref"`
	NamespaceRef             string                              `json:"namespace_ref"`
	WorkloadSelectorRef      string                              `json:"workload_selector_ref"`
	HelmChartRef             string                              `json:"helm_chart_ref"`
	ManifestBundleRef        string                              `json:"manifest_bundle_ref"`
	ValuesRef                string                              `json:"values_ref"`
	RBACRef                  string                              `json:"rbac_ref"`
	ServiceAccountRef        string                              `json:"service_account_ref"`
	ImageRef                 string                              `json:"image_ref"`
	ConfigMapRef             string                              `json:"config_map_ref"`
	SecretRef                string                              `json:"secret_ref"`
	RolloutStrategyRef       string                              `json:"rollout_strategy_ref"`
	RolloutReceiptRef        string                              `json:"rollout_receipt_ref"`
	DataArrivalValidatorRef  string                              `json:"data_arrival_validator_ref"`
	AuditRef                 string                              `json:"audit_ref"`
	EvidenceChainRef         string                              `json:"evidence_chain_ref"`
	OS                       string                              `json:"os"`
	RequiredTools            []packageRepositoryToolEvidence     `json:"required_tools"`
	BundledTools             []packageRepositoryToolEvidence     `json:"bundled_tools"`
	InstallCommandRef        string                              `json:"install_command_ref"`
	UninstallCommandRef      string                              `json:"uninstall_command_ref"`
	ConfigCommandRef         string                              `json:"config_command_ref"`
	PluginCommandRef         string                              `json:"plugin_command_ref"`
	Artifacts                []packageRepositoryManifestArtifact `json:"artifacts"`
}

type packageRepositoryManifestArtifact struct {
	ID                string `json:"id"`
	PackageID         string `json:"package_id"`
	Artifact          string `json:"artifact"`
	Path              string `json:"path"`
	File              string `json:"file"`
	OS                string `json:"os"`
	Arch              string `json:"arch"`
	ChecksumFile      string `json:"checksum_file"`
	ChecksumSignature string `json:"checksum_signature"`
	PublicKey         string `json:"public_key"`
	Fingerprint       string `json:"fingerprint"`
}

type packageRepositoryToolEvidence struct {
	Name        string `json:"name"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	EvidenceRef string `json:"evidence_ref"`
	Path        string `json:"path"`
	Bundled     bool   `json:"bundled"`
	System      bool   `json:"system"`
}

func hasAgentPackageTestOnlyRepositoryEvidence(packageID string) bool {
	for _, root := range agentPackageRepositoryRoots() {
		diagnostics := diagnosePackageRepositoryManifest(root)
		if !diagnostics.Ready || !isTestOnlyPackageRepositoryManifest(diagnostics.Manifest) {
			continue
		}
		for _, artifact := range diagnostics.Manifest.Artifacts {
			if packageRepositoryArtifactID(artifact) == packageID && packageRepositoryArtifactFilesExist(root, diagnostics.Manifest, artifact) {
				return true
			}
		}
	}
	return false
}

func agentPackageRepositoryRoots() []string {
	roots := []string{}
	configuredRoots := os.Getenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT")
	for _, configured := range filepath.SplitList(configuredRoots) {
		roots = appendPackageRepositoryRoot(roots, configured)
	}
	if strings.TrimSpace(configuredRoots) == "" {
		roots = appendPackageRepositoryRoot(roots, packageRepositoryFallbackRoot)
	}
	return roots
}

func appendPackageRepositoryRoot(roots []string, root string) []string {
	root = strings.TrimSpace(root)
	if root == "" {
		return roots
	}
	cleaned := filepath.Clean(root)
	for _, existing := range roots {
		if strings.EqualFold(existing, cleaned) {
			return roots
		}
	}
	return append(roots, cleaned)
}

func loadPackageRepositoryManifest(root string) (packageRepositoryManifest, bool) {
	diagnostics := diagnosePackageRepositoryManifest(root)
	return diagnostics.Manifest, diagnostics.Ready
}

func isTestOnlyPackageRepositoryManifest(manifest packageRepositoryManifest) bool {
	return manifest.Repository == "findx-agent-runtime-local" &&
		manifest.Status == "checksum_ready_test_signature_ready" &&
		manifest.SignatureScope == "test-only-runtime-generated"
}

func packageRepositoryArtifactID(artifact packageRepositoryManifestArtifact) string {
	if strings.TrimSpace(artifact.PackageID) != "" {
		return strings.TrimSpace(artifact.PackageID)
	}
	return strings.TrimSpace(artifact.ID)
}

func packageRepositoryArtifactFilesExist(root string, manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) bool {
	for _, ref := range []string{
		artifactPathRef(artifact),
		manifestBackedRef(artifact.ChecksumFile, manifest.ChecksumFile),
		manifestBackedRef(artifact.ChecksumSignature, manifest.ChecksumSignature),
		publicKeyOrFingerprintRef(manifest, artifact),
	} {
		if !safePackageRepositoryFileExists(root, ref) {
			return false
		}
	}
	return true
}

func packageRepositoryTrustChainVerified(root string, manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) bool {
	return len(packageRepositoryTrustChainBlockers(root, manifest, artifact)) == 0
}

func packageRepositoryTrustChainBlockers(root string, manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) []string {
	blockers := packageRepositoryArtifactEvidenceBlockers(root, manifest, artifact)
	blockers = append(blockers, missingTrustChainEvidenceBlockers(root, manifest)...)
	blockers = append(blockers, "TRUST_CHAIN_VERIFICATION_NOT_IMPLEMENTED", "TRUST_CHAIN_PENDING")
	return uniquePackageRepositoryBlockers(blockers)
}

func missingTrustChainEvidenceBlockers(root string, manifest packageRepositoryManifest) []string {
	refs := []struct {
		ref     string
		blocker string
	}{
		{ref: manifest.TrustRootRef, blocker: "TRUST_CHAIN_PRODUCTION_KEY_MISSING"},
		{ref: manifest.TrustRootFingerprintRef, blocker: "TRUST_CHAIN_PRODUCTION_KEY_MISSING"},
		{ref: manifest.VerificationReceiptRef, blocker: "TRUST_CHAIN_VERIFICATION_RECEIPT_MISSING"},
		{ref: manifest.ChecksumVerificationRef, blocker: "TRUST_CHAIN_CHECKSUM_VERIFICATION_MISSING"},
		{ref: manifest.SignatureVerificationRef, blocker: "TRUST_CHAIN_SIGNATURE_VERIFICATION_MISSING"},
	}
	blockers := []string{}
	for _, item := range refs {
		if !safePackageRepositoryFileExists(root, item.ref) {
			blockers = append(blockers, item.blocker)
		}
	}
	if len(blockers) > 0 {
		blockers = append(blockers, "TRUST_CHAIN_PENDING")
	}
	return blockers
}

func manifestHasPackageArtifact(manifest packageRepositoryManifest, packageID string) bool {
	for _, artifact := range manifest.Artifacts {
		if packageRepositoryArtifactID(artifact) == packageID {
			return true
		}
	}
	return false
}

func artifactPathRef(artifact packageRepositoryManifestArtifact) string {
	if strings.TrimSpace(artifact.Artifact) != "" {
		return artifact.Artifact
	}
	if strings.TrimSpace(artifact.Path) != "" {
		return artifact.Path
	}
	return artifact.File
}

func publicKeyOrFingerprintRef(manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) string {
	if strings.TrimSpace(artifact.PublicKey) != "" {
		return artifact.PublicKey
	}
	if strings.TrimSpace(artifact.Fingerprint) != "" {
		return artifact.Fingerprint
	}
	if strings.TrimSpace(manifest.PublicKey) != "" {
		return manifest.PublicKey
	}
	return manifest.PublicKeyFingerprintFile
}

func manifestBackedRef(artifactRef, manifestRef string) string {
	if strings.TrimSpace(artifactRef) != "" {
		return artifactRef
	}
	return manifestRef
}

func safePackageRepositoryFileExists(root, ref string) bool {
	cleanRef, ok := safePackageRepositoryRef(ref)
	if !ok {
		return false
	}
	info, err := os.Stat(filepath.Join(root, cleanRef))
	return err == nil && !info.IsDir()
}

func safePackageRepositoryRef(ref string) (string, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" || filepath.IsAbs(ref) {
		return "", false
	}
	cleaned := filepath.Clean(ref)
	if cleaned == "." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return "", false
	}
	return cleaned, true
}


const (
	testOnlyPackageRepository          = "findx-agent-runtime-local"
	testOnlyPackageRepositoryStatus    = "checksum_ready_test_signature_ready"
	testOnlyPackageRepositorySignature = "test-only-runtime-generated"
)

func buildTestOnlyPackageRepositoryManifest(root string) (packageRepositoryManifest, []string) {
	manifest := testOnlyPackageRepositoryManifest()
	return manifest, validatePackageRepositoryManifestStructure(root, manifest)
}

func testOnlyPackageRepositoryManifest() packageRepositoryManifest {
	manifest := packageRepositoryManifest{
		Repository:               testOnlyPackageRepository,
		Status:                   testOnlyPackageRepositoryStatus,
		SignatureScope:           testOnlyPackageRepositorySignature,
		ChecksumFile:             "checksums/SHA256SUMS",
		ChecksumSignature:        "signatures/SHA256SUMS.asc",
		PublicKey:                "signatures/test-public-key.asc",
		PublicKeyFingerprintFile: "signatures/test-key-fingerprint.txt",
		PackageRepositoryRef:     "security/package-repository.ref",
		SignatureRef:             "security/signature.ref",
		Checksum:                 "security/checksum.ref",
		ScriptManifestRef:        "manifests/script-manifest.json",
		ExecutorRef:              "executors/local-executor.ref",
		SafetyPolicyPath:         "security/safety-policy.yaml",
		Runner:                   "runners/local-runner.ref",
		SSHHostKey:               "security/ssh-host-key.ref",
		SSHFingerprint:           "security/ssh-fingerprint.ref",
		SystemdUnitRef:           "services/findx-agent.service",
		CurlInstallerRef:         "installers/linux-curl.sh",
		EnvTemplateRef:           "config/findx-agent.env.tpl",
		WindowsInstallerRef:      "installers/windows-installer.ps1",
		ServiceManifestRef:       "manifests/service.yaml",
		InstallRootPolicyRef:     "security/install-root-policy.yaml",
		RollbackManifestRef:      "manifests/rollback.yaml",
		UninstallManifestRef:     "manifests/uninstall.yaml",
		ReceiverEndpointRef:      "config/receiver-endpoint.ref",
		InstallCommandRef:        "commands/install.sh",
		UninstallCommandRef:      "commands/uninstall.sh",
		ConfigCommandRef:         "commands/configure.sh",
		PluginCommandRef:         "commands/plugins.sh",
		DataArrivalValidatorRef:  "validators/data-arrival.sh",
		AuditRef:                 "audit/package-repository-audit.json",
		EvidenceChainRef:         "evidence/package-repository-evidence.json",
		Method:                   "linux-curl/windows-powershell/windows-cmd/kubernetes-helm/operator/daemonset/sidecar/initcontainer",
		OS:                       "linux/windows/kubernetes",
		Artifacts:                testOnlyPackageRepositoryArtifacts(),
		RequiredTools:            testOnlyPackageRepositoryTools(false),
		BundledTools:             testOnlyPackageRepositoryTools(true),
	}
	applyTestOnlyPackageRepositoryKubernetesRefs(&manifest)
	return manifest
}

func applyTestOnlyPackageRepositoryKubernetesRefs(manifest *packageRepositoryManifest) {
	manifest.ClusterRef = "kubernetes/cluster.ref"
	manifest.NamespaceRef = "kubernetes/namespace.ref"
	manifest.WorkloadSelectorRef = "kubernetes/workload-selector.ref"
	manifest.HelmChartRef = "kubernetes/helm/findx-agent-chart.ref"
	manifest.ManifestBundleRef = "kubernetes/manifests/findx-agent-bundle.ref"
	manifest.ValuesRef = "kubernetes/helm/values.ref"
	manifest.RBACRef = "kubernetes/rbac.ref"
	manifest.ServiceAccountRef = "kubernetes/service-account.ref"
	manifest.ImageRef = "kubernetes/images/findx-agent.ref"
	manifest.ConfigMapRef = "kubernetes/config-map.ref"
	manifest.SecretRef = "kubernetes/secret.ref"
	manifest.RolloutStrategyRef = "kubernetes/rollout-strategy.ref"
	manifest.RolloutReceiptRef = "kubernetes/rollout-receipt.ref"
}

func testOnlyPackageRepositoryArtifacts() []packageRepositoryManifestArtifact {
	return []packageRepositoryManifestArtifact{
		testOnlyPackageRepositoryArtifact("host-collector", "linux", "bin/findx-categraf-linux-amd64"),
		testOnlyPackageRepositoryArtifact("host-collector", "windows", "bin/findx-categraf-windows-amd64.exe"),
		testOnlyPackageRepositoryArtifact("inspection-runner", "linux", "bin/findx-catpaw-linux-amd64"),
		testOnlyPackageRepositoryArtifact("inspection-runner", "windows", "bin/findx-catpaw-windows-amd64.exe"),
	}
}

func testOnlyPackageRepositoryManifestFileRefs() []string {
	return packageRepositoryManifestRefs(testOnlyPackageRepositoryManifest())
}

func testOnlyPackageRepositoryArtifact(packageID, osName, artifact string) packageRepositoryManifestArtifact {
	return packageRepositoryManifestArtifact{
		ID:       packageID,
		Artifact: artifact,
		OS:       osName,
		Arch:     "amd64",
	}
}

func testOnlyPackageRepositoryTools(bundled bool) []packageRepositoryToolEvidence {
	tools := []packageRepositoryToolEvidence{}
	for _, osName := range []string{"linux", "windows"} {
		for _, name := range packageRepositoryCriticalTools {
			tools = append(tools, packageRepositoryToolEvidence{
				Name:        name,
				OS:          osName,
				Arch:        "amd64",
				EvidenceRef: "tools/" + osName + "/" + name + ".ref",
				Bundled:     bundled,
				System:      !bundled,
			})
		}
	}
	return tools
}

func validatePackageRepositoryManifestStructure(root string, manifest packageRepositoryManifest) []string {
	blockers := []string{}
	if strings.TrimSpace(root) == "" {
		blockers = append(blockers, "PACKAGE_REPOSITORY_ROOT_MISSING")
	} else if info, err := os.Stat(root); err != nil || !info.IsDir() {
		blockers = append(blockers, "PACKAGE_REPOSITORY_ROOT_MISSING")
	}
	if !isTestOnlyPackageRepositoryManifest(manifest) {
		blockers = append(blockers, "PACKAGE_REPOSITORY_TEST_ONLY_METADATA_MISSING")
	}
	if len(manifest.Artifacts) == 0 {
		blockers = append(blockers, "PACKAGE_REPOSITORY_ARTIFACTS_MISSING")
	}
	blockers = append(blockers, unsafePackageRepositoryManifestRefBlockers(manifest)...)
	blockers = append(blockers, missingPackageRepositoryManifestRefBlockers(root, manifest)...)
	return uniquePackageRepositoryBlockers(blockers)
}

func unsafePackageRepositoryManifestRefBlockers(manifest packageRepositoryManifest) []string {
	blockers := []string{}
	for _, ref := range packageRepositoryManifestRefs(manifest) {
		if _, ok := safePackageRepositoryRef(ref); !ok {
			blockers = append(blockers, "PACKAGE_REPOSITORY_MANIFEST_REF_UNSAFE")
		}
	}
	return blockers
}

func missingPackageRepositoryManifestRefBlockers(root string, manifest packageRepositoryManifest) []string {
	blockers := []string{}
	for _, ref := range packageRepositoryManifestRefs(manifest) {
		if !safePackageRepositoryFileExists(root, ref) {
			blockers = append(blockers, "PACKAGE_REPOSITORY_MANIFEST_REF_MISSING")
		}
	}
	return blockers
}

func packageRepositoryManifestRefs(manifest packageRepositoryManifest) []string {
	refs := []string{
		manifest.ChecksumFile, manifest.ChecksumSignature, manifest.PublicKey,
		manifest.PublicKeyFingerprintFile, manifest.PackageRepositoryRef, manifest.SignatureRef,
		manifest.Checksum, manifest.ScriptManifestRef, manifest.ExecutorRef, manifest.SafetyPolicyPath,
		manifest.Runner, manifest.SSHHostKey, manifest.SSHFingerprint, manifest.SystemdUnitRef,
		manifest.CurlInstallerRef, manifest.EnvTemplateRef, manifest.WindowsInstallerRef,
		manifest.ServiceManifestRef, manifest.InstallRootPolicyRef, manifest.RollbackManifestRef,
		manifest.UninstallManifestRef, manifest.ReceiverEndpointRef, manifest.InstallCommandRef,
		manifest.UninstallCommandRef, manifest.ConfigCommandRef, manifest.PluginCommandRef,
		manifest.ClusterRef, manifest.NamespaceRef, manifest.WorkloadSelectorRef, manifest.HelmChartRef,
		manifest.ManifestBundleRef, manifest.ValuesRef, manifest.RBACRef, manifest.ServiceAccountRef,
		manifest.ImageRef, manifest.ConfigMapRef, manifest.SecretRef, manifest.RolloutStrategyRef,
		manifest.RolloutReceiptRef, manifest.DataArrivalValidatorRef, manifest.AuditRef,
		manifest.EvidenceChainRef,
	}
	for _, artifact := range manifest.Artifacts {
		refs = append(refs, artifactPathRef(artifact))
	}
	for _, tool := range append(manifest.RequiredTools, manifest.BundledTools...) {
		refs = append(refs, tool.EvidenceRef)
	}
	return refs
}

func marshalPackageRepositoryManifest(manifest packageRepositoryManifest) ([]byte, error) {
	return json.MarshalIndent(manifest, "", "  ")
}


const packageDownloadBlocker = agentBlocked + ": package repository artifact, checksum, signature, and public key evidence are not open"

const (
	packageDownloadArtifactRefInvalid = "PACKAGE_REPOSITORY_REF_INVALID"
	packageDownloadArtifactRefMissing = "PACKAGE_REPOSITORY_ARTIFACT_REF_MISSING"
	packageDownloadPackageMissing     = "PACKAGE_REPOSITORY_PACKAGE_MISSING"
	packageDownloadRequestMissing     = "PACKAGE_REQUEST_MISSING"
	packageDownloadUnknownPackage     = "PACKAGE_NOT_FOUND"
)

type packageDownloadEvidence struct {
	FilePath    string
	ArtifactRef string
}

func DownloadFindXAgentPackageArtifact(c *gin.Context) {
	evidence, blockers, ok := findPackageDownloadEvidence(c.Param("package"), c.Query("artifact"))
	if !ok {
		blockFindXAgentPackageDownload(c, blockers)
		return
	}
	file, err := os.Open(evidence.FilePath)
	if err != nil {
		blockFindXAgentPackageDownload(c, nil)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		blockFindXAgentPackageDownload(c, nil)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+packageArtifactDownloadName(evidence.ArtifactRef)+`"`)
	c.DataFromReader(http.StatusOK, info.Size(), packageArtifactContentType(evidence.ArtifactRef), file, nil)
}

func blockFindXAgentPackageDownload(c *gin.Context, blockers []string) {
	structuredBlockers := uniquePackageRepositoryBlockers(blockers)
	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"status":   "pending",
		"message":  "package artifact not yet available",
		"blockers": structuredBlockers,
	})
}

func packageDownloadMissingContracts(blockers []string) []string {
	base := []string{
		"package_repository_artifact",
		"checksum",
		"signature",
		"public_key",
	}
	return uniquePackageRepositoryBlockers(append(base, blockers...))
}

func findPackageDownloadEvidence(requestedPackage, requestedArtifact string) (packageDownloadEvidence, []string, bool) {
	packageID := strings.TrimSpace(requestedPackage)
	if packageID == "" {
		return packageDownloadEvidence{}, []string{packageDownloadRequestMissing}, false
	}
	if _, ok := findAgentPackage(packageID); !ok {
		return packageDownloadEvidence{}, []string{packageDownloadUnknownPackage}, false
	}
	cleanRequestedArtifact, ok := safePackageRepositoryRef(requestedArtifact)
	if !ok {
		return packageDownloadEvidence{}, []string{packageDownloadArtifactRefInvalid}, false
	}
	blockers := []string{}
	for _, root := range agentPackageRepositoryRoots() {
		evidence, rootBlockers, ok := findPackageDownloadEvidenceInRoot(root, packageID, cleanRequestedArtifact)
		if ok {
			return evidence, nil, true
		}
		blockers = append(blockers, rootBlockers...)
	}
	if len(blockers) == 0 {
		blockers = append(blockers, packageDownloadArtifactRefMissing)
	}
	return packageDownloadEvidence{}, uniquePackageRepositoryBlockers(blockers), false
}

func findPackageDownloadEvidenceInRoot(root, packageID, cleanRequestedArtifact string) (packageDownloadEvidence, []string, bool) {
	diagnostics := diagnosePackageRepositoryManifest(root)
	if !diagnostics.Ready {
		return packageDownloadEvidence{}, diagnostics.Blockers, false
	}
	manifest := diagnostics.Manifest
	if isTestOnlyPackageRepositoryManifest(manifest) {
		return packageDownloadEvidence{}, packageDownloadTestOnlyBlockers(root, manifest, packageID), false
	}
	return findProductionPackageDownloadEvidence(root, manifest, packageID, cleanRequestedArtifact)
}

func findProductionPackageDownloadEvidence(root string, manifest packageRepositoryManifest, packageID, cleanRequestedArtifact string) (packageDownloadEvidence, []string, bool) {
	packageSeen := false
	for _, artifact := range manifest.Artifacts {
		if packageRepositoryArtifactID(artifact) != packageID {
			continue
		}
		packageSeen = true
		if !packageDownloadArtifactMatches(manifest, artifact, packageID, cleanRequestedArtifact) {
			continue
		}
		blockers := packageRepositoryTrustChainBlockers(root, manifest, artifact)
		if len(blockers) > 0 {
			return packageDownloadEvidence{}, blockers, false
		}
		filePath, ok := packageRepositoryDownloadFilePath(root, cleanRequestedArtifact)
		if !ok {
			return packageDownloadEvidence{}, packageRepositoryArtifactEvidenceBlockers(root, manifest, artifact), false
		}
		return packageDownloadEvidence{FilePath: filePath, ArtifactRef: cleanRequestedArtifact}, nil, true
	}
	return packageDownloadEvidence{}, packageDownloadPackageOrArtifactBlockers(root, manifest, packageID, packageSeen), false
}

func packageDownloadPackageOrArtifactBlockers(root string, manifest packageRepositoryManifest, packageID string, packageSeen bool) []string {
	if !packageSeen {
		return []string{packageDownloadPackageMissing}
	}
	blockers := []string{packageDownloadArtifactRefMissing}
	blockers = append(blockers, packageRepositoryProductionEvidenceBlockers(root, manifest, packageID)...)
	return uniquePackageRepositoryBlockers(blockers)
}

func packageDownloadTestOnlyBlockers(root string, manifest packageRepositoryManifest, packageID string) []string {
	blockers := packageRepositoryTestOnlyBlockers(root, manifest, packageID)
	blockers = append(blockers, packageRepositoryTrustChainDiagnosticBlockers(root, packageID)...)
	if len(blockers) == 0 {
		blockers = append(blockers, packageDownloadPackageMissing)
	}
	return uniquePackageRepositoryBlockers(blockers)
}

func packageDownloadArtifactMatches(manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact, packageID, cleanRequestedArtifact string) bool {
	if packageRepositoryArtifactID(artifact) != packageID {
		return false
	}
	for _, ref := range packageDownloadAllowedRefs(manifest, artifact) {
		cleanRef, ok := safePackageRepositoryRef(ref)
		if ok && cleanRef == cleanRequestedArtifact {
			return true
		}
	}
	return false
}

func packageDownloadAllowedRefs(manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) []string {
	return []string{
		artifactPathRef(artifact),
		manifestBackedRef(artifact.ChecksumFile, manifest.ChecksumFile),
		manifestBackedRef(artifact.ChecksumSignature, manifest.ChecksumSignature),
		publicKeyOrFingerprintRef(manifest, artifact),
	}
}

func packageRepositoryDownloadFilePath(root, cleanRef string) (string, bool) {
	rootAbs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", false
	}
	fileAbs, err := filepath.Abs(filepath.Join(rootAbs, cleanRef))
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(rootAbs, fileAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	info, err := os.Stat(fileAbs)
	if err != nil || info.IsDir() {
		return "", false
	}
	return fileAbs, true
}

func packageArtifactContentType(ref string) string {
	lower := strings.ToLower(ref)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return "application/zip"
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"), strings.HasSuffix(lower, ".gz"):
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}

func packageArtifactDownloadName(ref string) string {
	name := filepath.Base(ref)
	name = strings.Map(func(r rune) rune {
		if r == '"' || r == '\\' || r == '/' || r < 32 {
			return -1
		}
		return r
	}, name)
	if strings.TrimSpace(name) == "" || name == "." {
		return "findx-agent-package.bin"
	}
	return name
}


const installerDownloadBlocker = agentBlocked + ": package repo, signature, script manifest, and executor contracts are not open"

func DownloadFindXAgentLinuxInstaller(c *gin.Context) {
	downloadFindXAgentInstaller(c, "linux.sh", "linux-shell", "linux")
}

func DownloadFindXAgentWindowsPowerShellInstaller(c *gin.Context) {
	downloadFindXAgentInstaller(c, "windows.ps1", "windows-powershell", "windows")
}

func DownloadFindXAgentWindowsBatchInstaller(c *gin.Context) {
	downloadFindXAgentInstaller(c, "windows.bat", "windows-cmd", "windows")
}

type installerPackageEvidence struct {
	PackageID      string
	ArtifactRef    string
	ChecksumRef    string
	SignatureRef   string
	PublicKeyRef   string
	SignatureScope string
	ToolRefs       map[string]string
}

func downloadFindXAgentInstaller(c *gin.Context, installer, platform, osName string) {
	evidence, environment, ok := findInstallerPackageEvidence(c.Query("package"), osName)
	if !ok {
		blockFindXAgentInstallerDownload(c, installer, platform, environment)
		return
	}
	script := renderInstallerScript(installer, platform, evidence)
	c.Data(http.StatusOK, installerContentType(installer), []byte(script))
}

func blockFindXAgentInstallerDownload(c *gin.Context, installer, platform string, environment model.FindXAgentInstallEnvironment) {
	c.JSON(http.StatusOK, gin.H{
		"code":                http.StatusOK,
		"status":              "pending",
		"installer":           installer,
		"platform":            platform,
		"install_environment": environment,
		"message":             "installer package evidence not yet available, retry after package repository is populated",
	})
}

func findInstallerPackageEvidence(requestedPackage, osName string) (installerPackageEvidence, model.FindXAgentInstallEnvironment, bool) {
	packageID := strings.TrimSpace(requestedPackage)
	if packageID == "" {
		packageID = "agent-core"
	}
	if _, ok := findAgentPackage(packageID); !ok {
		return installerPackageEvidence{}, defaultBlockedInstallerEnvironment(), false
	}
	environment := defaultBlockedInstallerEnvironment()
	for _, root := range agentPackageRepositoryRoots() {
		manifest, ok := loadPackageRepositoryManifest(root)
		if !ok || isTestOnlyPackageRepositoryManifest(manifest) {
			if ok {
				environment = blockedInstallEnvironment([]string{
					"PACKAGE_REPOSITORY_TEST_ONLY",
					"SIGNATURE_TEST_ONLY",
					"BUNDLED_INSTALL_ENVIRONMENT_CONTRACT_MISSING",
				})
			}
			continue
		}
		for _, artifact := range manifest.Artifacts {
			if !installerArtifactMatches(artifact, packageID, osName) {
				continue
			}
			environment = packageRepositoryInstallEnvironmentForOS(root, manifest, osName)
			if environment.Status != "ready" {
				continue
			}
			evidence, ok := installerEvidenceFromArtifact(manifest, artifact)
			if ok && packageRepositoryTrustChainVerified(root, manifest, artifact) {
				evidence.ToolRefs = installerToolRefs(environment.Tools)
				return evidence, environment, true
			}
			environment = appendInstallerEnvironmentBlockers(
				environment,
				packageRepositoryTrustChainBlockers(root, manifest, artifact)...,
			)
		}
	}
	return installerPackageEvidence{}, environment, false
}

func defaultBlockedInstallerEnvironment() model.FindXAgentInstallEnvironment {
	return blockedInstallEnvironment([]string{
		"PACKAGE_REPOSITORY_MISSING",
		"SIGNATURE_MISSING",
		"SCRIPT_MANIFEST_REF_MISSING",
		"EXECUTOR_REF_MISSING",
		"BUNDLED_INSTALL_ENVIRONMENT_CONTRACT_MISSING",
	})
}

func appendInstallerEnvironmentBlockers(environment model.FindXAgentInstallEnvironment, blockers ...string) model.FindXAgentInstallEnvironment {
	environment.Status = "blocked"
	environment.Blockers = uniquePackageRepositoryBlockers(append(environment.Blockers, blockers...))
	return environment
}

func installerArtifactMatches(artifact packageRepositoryManifestArtifact, packageID, osName string) bool {
	if packageRepositoryArtifactID(artifact) != packageID {
		return false
	}
	artifactOS := strings.ToLower(strings.TrimSpace(artifact.OS))
	return artifactOS == "" || artifactOS == strings.ToLower(osName)
}

func installerEvidenceFromArtifact(manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) (installerPackageEvidence, bool) {
	artifactRef := artifactPathRef(artifact)
	checksumRef := manifestBackedRef(artifact.ChecksumFile, manifest.ChecksumFile)
	signatureRef := manifestBackedRef(artifact.ChecksumSignature, manifest.ChecksumSignature)
	publicKeyRef := manifestBackedRef(artifact.PublicKey, manifest.PublicKey)
	for _, ref := range []string{artifactRef, checksumRef, signatureRef, publicKeyRef} {
		if _, ok := safePackageRepositoryRef(ref); !ok {
			return installerPackageEvidence{}, false
		}
	}
	return installerPackageEvidence{
		PackageID:      packageRepositoryArtifactID(artifact),
		ArtifactRef:    artifactRef,
		ChecksumRef:    checksumRef,
		SignatureRef:   signatureRef,
		PublicKeyRef:   publicKeyRef,
		SignatureScope: strings.TrimSpace(manifest.SignatureScope),
		ToolRefs:       map[string]string{},
	}, true
}

func installerToolRefs(tools []model.FindXAgentInstallToolEvidence) map[string]string {
	refs := map[string]string{}
	for _, tool := range tools {
		if tool.Status == "ready" && strings.TrimSpace(tool.EvidenceRef) != "" {
			refs[tool.Name] = tool.EvidenceRef
		}
	}
	return refs
}

func installerContentType(installer string) string {
	switch installer {
	case "linux.sh":
		return "text/x-shellscript; charset=utf-8"
	case "windows.ps1":
		return "text/plain; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
}

func renderInstallerScript(installer, platform string, evidence installerPackageEvidence) string {
	switch installer {
	case "windows.ps1":
		return renderWindowsPowerShellInstaller(evidence)
	case "windows.bat":
		return renderWindowsBatchInstaller(evidence)
	default:
		return renderLinuxShellInstaller(platform, evidence)
	}
}

func installerArtifactDownloadPath(evidence installerPackageEvidence) string {
	return installerPackageDownloadPath(evidence.PackageID, evidence.ArtifactRef)
}

func installerPackageDownloadPath(packageID, ref string) string {
	return "/api/v1/findx-agents/package-downloads/" + url.PathEscape(packageID) +
		"?artifact=" + url.QueryEscape(ref)
}

func renderLinuxShellInstaller(platform string, evidence installerPackageEvidence) string {
	artifactURL := installerArtifactDownloadPath(evidence)
	checksumURL := installerPackageDownloadPath(evidence.PackageID, evidence.ChecksumRef)
	signatureURL := installerPackageDownloadPath(evidence.PackageID, evidence.SignatureRef)
	publicKeyURL := installerPackageDownloadPath(evidence.PackageID, evidence.PublicKeyRef)
	toolEvidence := installerBundledToolEvidenceLines("#", evidence)
	return fmt.Sprintf(`#!/usr/bin/env sh
set -eu
%s
FINDX_BASE_URL="${FINDX_BASE_URL:-http://127.0.0.1:8080}"
FINDX_TOKEN="${FINDX_TOKEN:-<TOKEN>}"
PACKAGE_URL="${FINDX_BASE_URL}%s"
CHECKSUM_URL="${FINDX_BASE_URL}%s"
SIGNATURE_URL="${FINDX_BASE_URL}%s"
PUBLIC_KEY_URL="${FINDX_BASE_URL}%s"
INSTALL_ROOT="${FINDX_INSTALL_ROOT:-/opt/findx-agent}"
SERVICE_NAME="${FINDX_SERVICE_NAME:-findx-agent}"
TMP_DIR="$(mktemp -d /tmp/findx-agent.XXXXXX)"
TMP_FILE="${TMP_DIR}/package.tar.gz"
CHECKSUM_FILE="${TMP_DIR}/SHA256SUMS"
SIGNATURE_FILE="${TMP_DIR}/SHA256SUMS.sig"
PUBLIC_KEY_FILE="${TMP_DIR}/findx-agent.pub"
curl -fL -H "Authorization: Bearer ${FINDX_TOKEN}" -o "${TMP_FILE}" "${PACKAGE_URL}"
curl -fL -H "Authorization: Bearer ${FINDX_TOKEN}" -o "${CHECKSUM_FILE}" "${CHECKSUM_URL}"
curl -fL -H "Authorization: Bearer ${FINDX_TOKEN}" -o "${SIGNATURE_FILE}" "${SIGNATURE_URL}"
curl -fL -H "Authorization: Bearer ${FINDX_TOKEN}" -o "${PUBLIC_KEY_FILE}" "${PUBLIC_KEY_URL}"
printf 'artifact=%s\nchecksum=%s\nsignature=%s\npublic_key=%s\nsignature_scope=%s\n'
(cd "${TMP_DIR}" && sha256sum -c "${CHECKSUM_FILE}")
gpg --import "${PUBLIC_KEY_FILE}"
gpg --verify "${SIGNATURE_FILE}" "${CHECKSUM_FILE}"
install -d "${INSTALL_ROOT}"
tar -xzf "${TMP_FILE}" -C "${INSTALL_ROOT}"
printf 'install FindX Agent service %%s for %s\n' "${SERVICE_NAME}"
`, toolEvidence, artifactURL, checksumURL, signatureURL, publicKeyURL, evidence.ArtifactRef, evidence.ChecksumRef, evidence.SignatureRef,
		evidence.PublicKeyRef, evidence.SignatureScope, platform)
}

func renderWindowsPowerShellInstaller(evidence installerPackageEvidence) string {
	artifactURL := installerArtifactDownloadPath(evidence)
	checksumURL := installerPackageDownloadPath(evidence.PackageID, evidence.ChecksumRef)
	signatureURL := installerPackageDownloadPath(evidence.PackageID, evidence.SignatureRef)
	publicKeyURL := installerPackageDownloadPath(evidence.PackageID, evidence.PublicKeyRef)
	toolEvidence := installerBundledToolEvidenceLines("#", evidence)
	return fmt.Sprintf(`$ErrorActionPreference = "Stop"
%s
$FindXBaseUrl = if ($env:FINDX_BASE_URL) { $env:FINDX_BASE_URL } else { "http://127.0.0.1:8080" }
$FindXToken = if ($env:FINDX_TOKEN) { $env:FINDX_TOKEN } else { "<TOKEN>" }
$PackageUrl = "$FindXBaseUrl%s"
$ChecksumUrl = "$FindXBaseUrl%s"
$SignatureUrl = "$FindXBaseUrl%s"
$PublicKeyUrl = "$FindXBaseUrl%s"
$InstallRoot = if ($env:FINDX_INSTALL_ROOT) { $env:FINDX_INSTALL_ROOT } else { "C:\Program Files\FindX Agent" }
$ServiceName = if ($env:FINDX_SERVICE_NAME) { $env:FINDX_SERVICE_NAME } else { "FindX Agent" }
$TempDir = Join-Path $env:TEMP ("findx-agent-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null
$TempFile = Join-Path $TempDir "package.zip"
$ChecksumFile = Join-Path $TempDir "SHA256SUMS"
$SignatureFile = Join-Path $TempDir "SHA256SUMS.sig"
$PublicKeyFile = Join-Path $TempDir "findx-agent.pub"
Invoke-WebRequest -Uri $PackageUrl -Headers @{ Authorization = "Bearer $FindXToken" } -OutFile $TempFile
Invoke-WebRequest -Uri $ChecksumUrl -Headers @{ Authorization = "Bearer $FindXToken" } -OutFile $ChecksumFile
Invoke-WebRequest -Uri $SignatureUrl -Headers @{ Authorization = "Bearer $FindXToken" } -OutFile $SignatureFile
Invoke-WebRequest -Uri $PublicKeyUrl -Headers @{ Authorization = "Bearer $FindXToken" } -OutFile $PublicKeyFile
Write-Output "artifact=%s checksum=%s signature=%s public_key=%s signature_scope=%s"
$Hash = Get-FileHash -Algorithm SHA256 $TempFile
$ExpectedHash = Select-String -Path $ChecksumFile -Pattern "[A-Fa-f0-9]{64}" | Select-Object -First 1 | ForEach-Object { $_.Matches[0].Value }
if (!$ExpectedHash) { throw "FindX Agent checksum file has no SHA256 value" }
if ($Hash.Hash.ToLowerInvariant() -ne $ExpectedHash.ToLowerInvariant()) { throw "FindX Agent checksum mismatch" }
if (Get-Command gpg -ErrorAction SilentlyContinue) {
  & gpg --import $PublicKeyFile
  & gpg --verify $SignatureFile $ChecksumFile
  if ($LASTEXITCODE -ne 0) { throw "FindX Agent signature verification failed" }
} else {
  throw "FindX Agent signature verifier is missing"
}
New-Item -ItemType Directory -Force -Path $InstallRoot | Out-Null
Expand-Archive -Force -Path $TempFile -DestinationPath $InstallRoot
Write-Output "install FindX Agent service $ServiceName"
`, toolEvidence, artifactURL, checksumURL, signatureURL, publicKeyURL, evidence.ArtifactRef, evidence.ChecksumRef, evidence.SignatureRef,
		evidence.PublicKeyRef, evidence.SignatureScope)
}

func renderWindowsBatchInstaller(evidence installerPackageEvidence) string {
	artifactURL := installerArtifactDownloadPath(evidence)
	checksumURL := installerPackageDownloadPath(evidence.PackageID, evidence.ChecksumRef)
	signatureURL := installerPackageDownloadPath(evidence.PackageID, evidence.SignatureRef)
	publicKeyURL := installerPackageDownloadPath(evidence.PackageID, evidence.PublicKeyRef)
	toolEvidence := installerBundledToolEvidenceLines("rem", evidence)
	return fmt.Sprintf(`@echo off
setlocal enabledelayedexpansion
%s
if "%%FINDX_BASE_URL%%"=="" set "FINDX_BASE_URL=http://127.0.0.1:8080"
if "%%FINDX_TOKEN%%"=="" set "FINDX_TOKEN=<TOKEN>"
if "%%FINDX_INSTALL_ROOT%%"=="" set "FINDX_INSTALL_ROOT=%%ProgramFiles%%\FindX Agent"
if "%%FINDX_SERVICE_NAME%%"=="" set "FINDX_SERVICE_NAME=FindX Agent"
set "PACKAGE_URL=%%FINDX_BASE_URL%%%s"
set "CHECKSUM_URL=%%FINDX_BASE_URL%%%s"
set "SIGNATURE_URL=%%FINDX_BASE_URL%%%s"
set "PUBLIC_KEY_URL=%%FINDX_BASE_URL%%%s"
set "TMP_DIR=%%TEMP%%\findx-agent-%%RANDOM%%%%RANDOM%%"
mkdir "%%TMP_DIR%%"
set "TMP_FILE=%%TMP_DIR%%\package.zip"
set "CHECKSUM_FILE=%%TMP_DIR%%\SHA256SUMS"
set "SIGNATURE_FILE=%%TMP_DIR%%\SHA256SUMS.sig"
set "PUBLIC_KEY_FILE=%%TMP_DIR%%\findx-agent.pub"
certutil -urlcache -f "%%PACKAGE_URL%%" "%%TMP_FILE%%"
certutil -urlcache -f "%%CHECKSUM_URL%%" "%%CHECKSUM_FILE%%"
certutil -urlcache -f "%%SIGNATURE_URL%%" "%%SIGNATURE_FILE%%"
certutil -urlcache -f "%%PUBLIC_KEY_URL%%" "%%PUBLIC_KEY_FILE%%"
echo artifact=%s checksum=%s signature=%s public_key=%s signature_scope=%s
certutil -hashfile "%%TMP_FILE%%" SHA256
powershell -NoProfile -ExecutionPolicy Bypass -Command "$hash=(Get-FileHash -Algorithm SHA256 '%%TMP_FILE%%').Hash.ToLowerInvariant(); $expected=(Select-String -Path '%%CHECKSUM_FILE%%' -Pattern '[A-Fa-f0-9]{64}' | Select-Object -First 1).Matches[0].Value.ToLowerInvariant(); if ($hash -ne $expected) { throw 'FindX Agent checksum mismatch' }; if (Get-Command gpg -ErrorAction SilentlyContinue) { & gpg --import '%%PUBLIC_KEY_FILE%%'; & gpg --verify '%%SIGNATURE_FILE%%' '%%CHECKSUM_FILE%%'; if ($LASTEXITCODE -ne 0) { throw 'FindX Agent signature verification failed' } } else { throw 'FindX Agent signature verifier is missing' }; New-Item -ItemType Directory -Force -Path '%%FINDX_INSTALL_ROOT%%' | Out-Null; Expand-Archive -Force -Path '%%TMP_FILE%%' -DestinationPath '%%FINDX_INSTALL_ROOT%%'"
echo install FindX Agent service %%FINDX_SERVICE_NAME%%
`, toolEvidence, artifactURL, checksumURL, signatureURL, publicKeyURL, evidence.ArtifactRef, evidence.ChecksumRef, evidence.SignatureRef,
		evidence.PublicKeyRef, evidence.SignatureScope)
}

func installerBundledToolEvidenceLines(comment string, evidence installerPackageEvidence) string {
	names := make([]string, 0, len(evidence.ToolRefs))
	for name := range evidence.ToolRefs {
		names = append(names, name)
	}
	sort.Strings(names)
	var builder strings.Builder
	builder.WriteString(comment)
	builder.WriteString(" bundled_install_environment=ready\n")
	for _, name := range names {
		builder.WriteString(comment)
		builder.WriteString(" bundled_tool ")
		builder.WriteString(name)
		builder.WriteString("=")
		builder.WriteString(evidence.ToolRefs[name])
		builder.WriteString("\n")
	}
	return strings.TrimRight(builder.String(), "\n")
}


type packageRepositoryManifestDiagnostics struct {
	Manifest packageRepositoryManifest
	Blockers []string
	Ready    bool
}

func diagnosePackageRepositoryManifest(root string) packageRepositoryManifestDiagnostics {
	var manifest packageRepositoryManifest
	data, err := os.ReadFile(filepath.Join(root, packageRepositoryManifestPath))
	if err != nil {
		return packageRepositoryManifestDiagnostics{
			Blockers: []string{"PACKAGE_REPOSITORY_MANIFEST_MISSING"},
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return packageRepositoryManifestDiagnostics{
			Blockers: []string{"PACKAGE_REPOSITORY_MANIFEST_INVALID"},
		}
	}
	return packageRepositoryManifestDiagnostics{
		Manifest: manifest,
		Ready:    true,
	}
}

func packageRepositoryManifestDiagnosticBlockers(root, packageID string) []string {
	diagnostics := diagnosePackageRepositoryManifest(root)
	if !diagnostics.Ready {
		return diagnostics.Blockers
	}
	if isTestOnlyPackageRepositoryManifest(diagnostics.Manifest) {
		return packageRepositoryTestOnlyBlockers(root, diagnostics.Manifest, packageID)
	}
	return packageRepositoryProductionEvidenceBlockers(root, diagnostics.Manifest, packageID)
}

func packageRepositoryTestOnlyBlockers(root string, manifest packageRepositoryManifest, packageID string) []string {
	if !manifestHasPackageArtifact(manifest, packageID) {
		return nil
	}
	if !packageRepositoryPackageArtifactRefsSafe(manifest, packageID) {
		return packageRepositoryProductionEvidenceBlockers(root, manifest, packageID)
	}
	blockers := []string{"PACKAGE_REPOSITORY_TEST_ONLY", "SIGNATURE_TEST_ONLY"}
	blockers = append(blockers, packageRepositoryProductionEvidenceBlockers(root, manifest, packageID)...)
	return uniquePackageRepositoryBlockers(blockers)
}

func packageRepositoryProductionEvidenceBlockers(root string, manifest packageRepositoryManifest, packageID string) []string {
	blockers := []string{}
	for _, artifact := range manifest.Artifacts {
		if packageRepositoryArtifactID(artifact) != packageID {
			continue
		}
		blockers = append(blockers, packageRepositoryArtifactEvidenceBlockers(root, manifest, artifact)...)
	}
	return uniquePackageRepositoryBlockers(blockers)
}

func packageRepositoryPackageArtifactRefsSafe(manifest packageRepositoryManifest, packageID string) bool {
	for _, artifact := range manifest.Artifacts {
		if packageRepositoryArtifactID(artifact) == packageID && !packageRepositoryArtifactRefsSafe(manifest, artifact) {
			return false
		}
	}
	return true
}

func packageRepositoryArtifactRefsSafe(manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) bool {
	for _, ref := range packageDownloadAllowedRefs(manifest, artifact) {
		if _, ok := safePackageRepositoryRef(ref); !ok {
			return false
		}
	}
	return true
}

func packageRepositoryArtifactEvidenceBlockers(root string, manifest packageRepositoryManifest, artifact packageRepositoryManifestArtifact) []string {
	blockers := []string{}
	if !safePackageRepositoryFileExists(root, artifactPathRef(artifact)) {
		blockers = append(blockers, "PACKAGE_REPOSITORY_ARTIFACT_MISSING")
	}
	if !safePackageRepositoryFileExists(root, manifestBackedRef(artifact.ChecksumFile, manifest.ChecksumFile)) {
		blockers = append(blockers, "PACKAGE_REPOSITORY_CHECKSUM_MISSING")
	}
	if !safePackageRepositoryFileExists(root, manifestBackedRef(artifact.ChecksumSignature, manifest.ChecksumSignature)) {
		blockers = append(blockers, "PRODUCTION_SIGNATURE_MISSING")
	}
	if !safePackageRepositoryFileExists(root, publicKeyOrFingerprintRef(manifest, artifact)) {
		blockers = append(blockers, "PRODUCTION_PUBLIC_KEY_MISSING")
	}
	if manifest.SignatureScope == "" || isTestOnlyPackageRepositoryManifest(manifest) {
		blockers = append(blockers, "PRODUCTION_SIGNATURE_MISSING")
	}
	return blockers
}

func packageRepositoryTrustChainDiagnosticBlockers(root, packageID string) []string {
	diagnostics := diagnosePackageRepositoryManifest(root)
	if !diagnostics.Ready {
		return diagnostics.Blockers
	}
	blockers := []string{}
	for _, artifact := range diagnostics.Manifest.Artifacts {
		if packageRepositoryArtifactID(artifact) == packageID {
			blockers = append(blockers, packageRepositoryTrustChainBlockers(root, diagnostics.Manifest, artifact)...)
		}
	}
	return uniquePackageRepositoryBlockers(blockers)
}
