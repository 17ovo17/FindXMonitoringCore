package handler

import (
	"os"
	"path/filepath"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestFindXAgentPackageRepositoryTestOnlyManifestExposesBlockedEnvironmentEvidence(t *testing.T) {
	repo := t.TempDir()
	for _, ref := range testOnlyPackageRepositoryManifestFileRefs() {
		writePackageRepositoryTestFile(t, repo, ref)
	}
	writeTestOnlyPackageRepositoryManifest(t, repo)
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	environment := agentPackageInstallEnvironment("host-collector")

	if environment.Status != "blocked" {
		t.Fatalf("test-only environment must stay blocked, got %q", environment.Status)
	}
	assertInstallEnvironmentHasPlatform(t, environment.Platforms, "linux/amd64")
	assertInstallEnvironmentHasPlatform(t, environment.Platforms, "windows/amd64")
	assertInstallEnvironmentHasReadyTool(t, environment.Tools, "signature_verifier", "linux")
	assertInstallEnvironmentHasReadyTool(t, environment.Tools, "signature_verifier", "windows")
	for _, blocker := range []string{"PACKAGE_REPOSITORY_TEST_ONLY", "SIGNATURE_TEST_ONLY"} {
		if !containsLifecycleTestString(environment.Blockers, blocker) {
			t.Fatalf("test-only environment missing blocker %s: %#v", blocker, environment.Blockers)
		}
	}
	if containsLifecycleTestString(environment.Blockers, "PACKAGE_REPOSITORY_MANIFEST_INVALID") {
		t.Fatalf("valid test-only manifest must not be invalid: %#v", environment.Blockers)
	}
}

func TestFindXAgentPackageRepositoryInvalidManifestDoesNotSynthesizeEnvironmentEvidence(t *testing.T) {
	repo := t.TempDir()
	manifestPath := filepath.Join(repo, packageRepositoryManifestPath)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, []byte("{invalid-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	environment := agentPackageInstallEnvironment("host-collector")

	if !containsLifecycleTestString(environment.Blockers, "PACKAGE_REPOSITORY_MANIFEST_INVALID") {
		t.Fatalf("invalid manifest should expose precise blocker: %#v", environment.Blockers)
	}
	if len(environment.Tools) != 0 || len(environment.Platforms) != 0 {
		t.Fatalf("invalid manifest must not synthesize evidence, tools=%#v platforms=%#v", environment.Tools, environment.Platforms)
	}
}

func writeTestOnlyPackageRepositoryManifest(t *testing.T, repo string) {
	t.Helper()
	manifest, blockers := buildTestOnlyPackageRepositoryManifest(repo)
	if len(blockers) != 0 {
		t.Fatalf("test-only manifest refs should be complete: %#v", blockers)
	}
	data, err := marshalPackageRepositoryManifest(manifest)
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

func assertInstallEnvironmentHasReadyTool(t *testing.T, tools []model.FindXAgentInstallToolEvidence, name, osName string) {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name && tool.OS == osName && tool.Status == "ready" && tool.Bundled {
			return
		}
	}
	t.Fatalf("missing ready bundled tool %s/%s: %#v", name, osName, tools)
}

func assertInstallEnvironmentHasPlatform(t *testing.T, platforms []string, want string) {
	t.Helper()
	if !containsLifecycleTestString(platforms, want) {
		t.Fatalf("missing platform %s: %#v", want, platforms)
	}
}
