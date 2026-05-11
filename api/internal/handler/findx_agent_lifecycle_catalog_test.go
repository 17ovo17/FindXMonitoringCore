package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindXAgentSourceRootsIncludeMigrationFallbacksAndDedupe(t *testing.T) {
	t.Setenv("FINDX_AGENT_SOURCE_ROOT", "  ")

	roots := agentSourceRoots()
	assertLifecycleCatalogContains(t, roots, `D:\项目迁移文件\平台源码`)
	assertLifecycleCatalogContains(t, roots, `D:\平台源码`)
	assertLifecycleCatalogContains(t, roots, `/mnt/d/项目迁移文件/平台源码`)
	assertLifecycleCatalogContains(t, roots, `/mnt/d/平台源码`)
	assertLifecycleCatalogUnique(t, roots)
}
func TestFindXAgentSourceRootsUseEnvAndCurrentDirectoryParents(t *testing.T) {
	project := t.TempDir()
	sourceRoot := filepath.Join(project, "平台源码")
	workDir := filepath.Join(project, "api", "internal", "handler")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FINDX_AGENT_SOURCE_ROOT", sourceRoot+string(os.PathListSeparator)+sourceRoot)

	roots := agentSourceRoots()
	assertLifecycleCatalogContains(t, roots, sourceRoot)
	assertLifecycleCatalogUnique(t, roots)
}
func TestFindXAgentMigratedMatureSourcesOnlyMarkMatchingPackagesPresent(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{
		filepath.Join(root, "categraf-main (1)", "categraf-main"),
		filepath.Join(root, "catpaw-master", "catpaw-master"),
		filepath.Join(root, "skywalking-master"),
		filepath.Join(root, "skywalking-booster-ui-main"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("FINDX_AGENT_SOURCE_ROOT", root)

	presentPackages := []string{"host-collector", "container-collector", "log-collector", "inspection-runner"}
	for _, packageID := range presentPackages {
		if state := agentSourceState(packageID); state != "LOCAL_SOURCE_PRESENT" {
			t.Fatalf("package %s should detect migrated local source, got %q", packageID, state)
		}
	}
	missingPackages := []string{"java-app", "python-app", "nodejs-app", "php-app", "go-app", "rust-app", "ruby-app", "gateway-probe", "browser-client"}
	for _, packageID := range missingPackages {
		if state := agentSourceState(packageID); state != "LOCAL_SOURCE_MISSING" {
			t.Fatalf("package %s must not reuse SkyWalking UI/OAP as agent source, got %q", packageID, state)
		}
	}
}
func TestFindXAgentPackagesKeepBlockedContractsWhenSourceIsPresent(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "categraf-main (1)", "categraf-main"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FINDX_AGENT_SOURCE_ROOT", root)
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", t.TempDir())

	pkg, ok := findAgentPackage("host-collector")
	if !ok {
		t.Fatal("expected host collector package")
	}
	if pkg.SourceState != "LOCAL_SOURCE_PRESENT" {
		t.Fatalf("expected local source present, got %q", pkg.SourceState)
	}
	if pkg.Status != "blocked" || pkg.Signature != "missing" {
		t.Fatalf("source presence must not mark package ready, status=%q signature=%q", pkg.Status, pkg.Signature)
	}
	for _, blocker := range []string{"PACKAGE_REPOSITORY_MISSING", "SIGNATURE_MISSING", "INSTALL_PLAN_CONTRACT_MISSING", "CONFIG_ROLLOUT_CONTRACT_MISSING"} {
		if !containsLifecycleTestString(pkg.Blockers, blocker) {
			t.Fatalf("missing blocker %s: %#v", blocker, pkg.Blockers)
		}
	}
}
func TestFindXAgentPackageRepositoryRootsUseExplicitEnvWithoutFallback(t *testing.T) {
	repo := t.TempDir()
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", "  "+repo+"  "+string(os.PathListSeparator)+repo+string(os.PathListSeparator)+"  ")

	roots := agentPackageRepositoryRoots()
	if len(roots) != 1 {
		t.Fatalf("explicit package repository roots should dedupe and ignore fallback: %#v", roots)
	}
	if roots[0] != filepath.Clean(repo) {
		t.Fatalf("explicit package repository root mismatch, got %q want %q", roots[0], filepath.Clean(repo))
	}
}
func TestFindXAgentPackageRepositoryTestOnlyEvidenceDoesNotMarkReady(t *testing.T) {
	repo := t.TempDir()
	for _, ref := range []string{"artifacts/host-collector.tar.gz", "checksums/runtime.sha256", "signatures/runtime.sha256.sig", "keys/runtime-test.pub", "keys/runtime-test.pub.sha256", "artifacts/inspection-runner.tar.gz"} {
		writePackageRepositoryTestFile(t, repo, ref)
	}
	writePackageRepositoryManifest(t, repo, packageRepositoryManifest{
		ChecksumFile:             "checksums/runtime.sha256",
		ChecksumSignature:        "signatures/runtime.sha256.sig",
		PublicKey:                "keys/runtime-test.pub",
		PublicKeyFingerprintFile: "keys/runtime-test.pub.sha256",
		Artifacts: []packageRepositoryManifestArtifact{
			{ID: "host-collector", File: "artifacts/host-collector.tar.gz", OS: "linux", Arch: "amd64"},
			{ID: "inspection-runner", File: "artifacts/inspection-runner.tar.gz", OS: "linux", Arch: "amd64"},
		},
	})

	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	host, ok := findAgentPackage("host-collector")
	if !ok {
		t.Fatal("expected host collector package")
	}
	assertPackageRepositoryTestOnlyBlocked(t, host.Status, host.Signature, host.Blockers)
	javaPkg, ok := findAgentPackage("java-app")
	if !ok {
		t.Fatal("expected java package")
	}
	if containsLifecycleTestString(javaPkg.Blockers, "PACKAGE_REPOSITORY_TEST_ONLY") || containsLifecycleTestString(javaPkg.Blockers, "SIGNATURE_TEST_ONLY") {
		t.Fatalf("java package must not inherit host evidence: %#v", javaPkg.Blockers)
	}
	for _, blocker := range []string{"PACKAGE_REPOSITORY_MISSING", "SIGNATURE_MISSING"} {
		if !containsLifecycleTestString(javaPkg.Blockers, blocker) {
			t.Fatalf("java package should keep generic blocker %s: %#v", blocker, javaPkg.Blockers)
		}
	}
}
func TestFindXAgentPackageRepositoryArtifactEvidenceOverridesManifestEvidence(t *testing.T) {
	repo := t.TempDir()
	writePackageRepositoryTestFile(t, repo, "artifacts/host-collector.tar.gz")
	writePackageRepositoryTestFile(t, repo, "checksums/host-collector.sha256")
	writePackageRepositoryTestFile(t, repo, "signatures/host-collector.sha256.sig")
	writePackageRepositoryTestFile(t, repo, "keys/runtime-test.pub")
	writePackageRepositoryManifest(t, repo, packageRepositoryManifest{
		ChecksumFile:      "checksums/missing-runtime.sha256",
		ChecksumSignature: "signatures/missing-runtime.sha256.sig",
		PublicKey:         "keys/missing-runtime.pub",
		Artifacts: []packageRepositoryManifestArtifact{
			{ID: "host-collector", Artifact: "artifacts/host-collector.tar.gz", ChecksumFile: "checksums/host-collector.sha256", ChecksumSignature: "signatures/host-collector.sha256.sig", PublicKey: "keys/runtime-test.pub"},
		},
	})
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	host, ok := findAgentPackage("host-collector")
	if !ok {
		t.Fatal("expected host collector package")
	}
	assertPackageRepositoryTestOnlyBlocked(t, host.Status, host.Signature, host.Blockers)
}

func TestFindXAgentPackageRepositoryRejectsMissingSignatureEvidence(t *testing.T) {
	repo := t.TempDir()
	writePackageRepositoryTestFile(t, repo, "artifacts/host-collector.tar.gz")
	writePackageRepositoryTestFile(t, repo, "checksums/host-collector.sha256")
	writePackageRepositoryTestFile(t, repo, "keys/runtime-test.pub")
	writeProductionPackageRepositoryManifest(t, repo, packageRepositoryManifest{
		ChecksumFile:      "checksums/host-collector.sha256",
		ChecksumSignature: "signatures/missing-host-collector.sha256.sig",
		PublicKey:         "keys/runtime-test.pub",
		Artifacts: []packageRepositoryManifestArtifact{
			{ID: "host-collector", File: "artifacts/host-collector.tar.gz"},
		},
	})
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	pkg, ok := findAgentPackage("host-collector")
	if !ok {
		t.Fatal("expected host collector package")
	}
	if containsLifecycleTestString(pkg.Blockers, "PACKAGE_REPOSITORY_TEST_ONLY") {
		t.Fatalf("production manifest must not expose test-only blocker: %#v", pkg.Blockers)
	}
	if !containsLifecycleTestString(pkg.InstallEnvironment.Blockers, "PRODUCTION_SIGNATURE_MISSING") {
		t.Fatalf("missing production signature should be precise: %#v", pkg.InstallEnvironment.Blockers)
	}
}

func TestFindXAgentPackageRepositoryInvalidManifestIsNotMissing(t *testing.T) {
	repo := t.TempDir()
	manifestPath := filepath.Join(repo, packageRepositoryManifestPath)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, []byte("{invalid-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	pkg, ok := findAgentPackage("host-collector")
	if !ok {
		t.Fatal("expected host collector package")
	}
	if !containsLifecycleTestString(pkg.InstallEnvironment.Blockers, "PACKAGE_REPOSITORY_MANIFEST_INVALID") {
		t.Fatalf("invalid manifest should expose precise blocker: %#v", pkg.InstallEnvironment.Blockers)
	}
	if containsLifecycleTestString(pkg.InstallEnvironment.Blockers, "PACKAGE_REPOSITORY_MANIFEST_MISSING") {
		t.Fatalf("invalid manifest must not be treated as missing: %#v", pkg.InstallEnvironment.Blockers)
	}
	for _, blocker := range []string{"INSTALL_ENVIRONMENT_MANIFEST_MISSING", "REQUIRED_TOOLS_MISSING", "BUNDLED_TOOLS_MISSING"} {
		if containsLifecycleTestString(pkg.InstallEnvironment.Blockers, blocker) {
			t.Fatalf("invalid manifest must not include generic environment blocker %s: %#v", blocker, pkg.InstallEnvironment.Blockers)
		}
	}
	if len(pkg.InstallEnvironment.Tools) != 0 {
		t.Fatalf("invalid manifest must not synthesize tool missing blockers: %#v", pkg.InstallEnvironment.Tools)
	}
}

func TestFindXAgentPackageRepositoryTestOnlyManifestBuilderAndDiagnostics(t *testing.T) {
	repo := t.TempDir()
	writeTestOnlyPackageRepositoryFiles(t, repo)
	assertTestOnlyPackageRepositoryManifestDiagnostics(t, repo)
	assertTestOnlyPackageRepositoryPackage(t, repo)
}

func TestFindXAgentPackageRepositoryRejectsUnsafeArtifactReference(t *testing.T) {
	repo := t.TempDir()
	writePackageRepositoryTestFile(t, repo, "checksums/host-collector.sha256")
	writePackageRepositoryTestFile(t, repo, "signatures/host-collector.sha256.sig")
	writePackageRepositoryTestFile(t, repo, "keys/runtime-test.pub")
	writePackageRepositoryManifest(t, repo, packageRepositoryManifest{
		ChecksumFile:      "checksums/host-collector.sha256",
		ChecksumSignature: "signatures/host-collector.sha256.sig",
		PublicKey:         "keys/runtime-test.pub",
		Artifacts: []packageRepositoryManifestArtifact{
			{ID: "host-collector", File: "../outside/host-collector.tar.gz"},
		},
	})
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	pkg, ok := findAgentPackage("host-collector")
	if !ok {
		t.Fatal("expected host collector package")
	}
	if containsLifecycleTestString(pkg.Blockers, "PACKAGE_REPOSITORY_TEST_ONLY") {
		t.Fatalf("unsafe artifact reference must not expose test-only evidence: %#v", pkg.Blockers)
	}
}

func writePackageRepositoryManifest(t *testing.T, repo string, manifest packageRepositoryManifest) {
	t.Helper()
	manifest.Repository = "findx-agent-runtime-local"
	manifest.Status = "checksum_ready_test_signature_ready"
	manifest.SignatureScope = "test-only-runtime-generated"
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(repo, packageRepositoryManifestPath)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertPackageRepositoryTestOnlyBlocked(t *testing.T, status, signature string, blockers []string) {
	t.Helper()
	if status != "blocked" || signature != "missing" {
		t.Fatalf("test-only evidence must not mark package ready, status=%q signature=%q", status, signature)
	}
	for _, blocker := range []string{"PACKAGE_REPOSITORY_TEST_ONLY", "SIGNATURE_TEST_ONLY", "PRODUCTION_PACKAGE_REPOSITORY_MISSING", "PRODUCTION_SIGNATURE_MISSING", "INSTALL_PLAN_CONTRACT_MISSING", "CONFIG_ROLLOUT_CONTRACT_MISSING"} {
		if !containsLifecycleTestString(blockers, blocker) {
			t.Fatalf("host collector missing blocker %s: %#v", blocker, blockers)
		}
	}
	for _, blocker := range []string{"PACKAGE_REPOSITORY_MISSING", "SIGNATURE_MISSING"} {
		if containsLifecycleTestString(blockers, blocker) {
			t.Fatalf("test-only evidence should replace generic blocker %s: %#v", blocker, blockers)
		}
	}
}

func assertLifecycleCatalogNoMojibake(t *testing.T, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []rune{'\u93cd', '\u9369', '\u6d93', '\u93b8', '\u9473', '\u5bb8', '\u7487', '\u941c', '\u7039', '\u7f03', '\u59af'} {
		if strings.ContainsRune(string(data), fragment) {
			t.Fatalf("catalog JSON contains mojibake fragment %q: %s", string(fragment), string(data))
		}
	}
}

func writePackageRepositoryTestFile(t *testing.T, repo, ref string) {
	t.Helper()
	path := filepath.Join(repo, ref)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("test-only"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeTestOnlyPackageRepositoryFiles(t *testing.T, repo string) {
	t.Helper()
	for _, ref := range testOnlyPackageRepositoryManifestFileRefs() {
		writePackageRepositoryTestFile(t, repo, ref)
	}
}

func assertTestOnlyPackageRepositoryPackage(t *testing.T, repo string) {
	t.Helper()
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)
	pkg, ok := findAgentPackage("host-collector")
	if !ok {
		t.Fatal("expected host collector package")
	}
	assertPackageRepositoryTestOnlyBlocked(t, pkg.Status, pkg.Signature, pkg.Blockers)
	assertLifecycleCatalogNoMojibake(t, findXAgentPackages())
	assertLifecycleCatalogNoMojibake(t, findXAgentConfigTemplates())
	for _, blocker := range []string{"PACKAGE_REPOSITORY_TEST_ONLY", "SIGNATURE_TEST_ONLY"} {
		if !containsLifecycleTestString(pkg.InstallEnvironment.Blockers, blocker) {
			t.Fatalf("host collector install environment must keep test-only blocker %s: %#v", blocker, pkg.InstallEnvironment.Blockers)
		}
	}
	assertInstallEnvironmentHasReadyTool(t, pkg.InstallEnvironment.Tools, "signature_verifier", "linux")
	assertInstallEnvironmentHasReadyTool(t, pkg.InstallEnvironment.Tools, "signature_verifier", "windows")
	if containsLifecycleTestString(pkg.Blockers, "PACKAGE_REPOSITORY_MANIFEST_INVALID") {
		t.Fatalf("builder output must not be invalid JSON: %#v", pkg.Blockers)
	}
}

func assertLifecycleCatalogContains(t *testing.T, roots []string, want string) {
	t.Helper()
	cleaned := filepath.Clean(want)
	for _, root := range roots {
		if root == cleaned {
			return
		}
	}
	t.Fatalf("roots missing %q: %#v", cleaned, roots)
}

func assertLifecycleCatalogUnique(t *testing.T, roots []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, root := range roots {
		if root == "" {
			t.Fatalf("roots should not contain empty values: %#v", roots)
		}
		key := filepath.Clean(root)
		if seen[key] {
			t.Fatalf("roots should be unique, duplicate %q in %#v", key, roots)
		}
		seen[key] = true
	}
}
