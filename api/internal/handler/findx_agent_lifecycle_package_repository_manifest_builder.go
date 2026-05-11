package handler

import (
	"encoding/json"
	"os"
	"strings"
)

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
