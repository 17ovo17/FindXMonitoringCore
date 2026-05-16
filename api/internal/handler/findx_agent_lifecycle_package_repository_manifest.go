package handler

import (
	"os"
	"path/filepath"
	"strings"
)

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
