package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertTestOnlyPackageRepositoryManifestDiagnostics(t *testing.T, repo string) {
	t.Helper()
	diagnostics := writeAndDiagnoseTestOnlyManifest(t, repo)
	if !diagnostics.Ready || !isTestOnlyPackageRepositoryManifest(diagnostics.Manifest) {
		t.Fatalf("builder manifest must stay ready and test-only: %#v", diagnostics.Manifest)
	}
	if len(diagnostics.Manifest.Artifacts) != 4 {
		t.Fatalf("builder must include host collector and inspection runner artifacts: %#v", diagnostics.Manifest.Artifacts)
	}
	assertTestOnlyManifestRefs(t, diagnostics.Manifest)
}

func writeAndDiagnoseTestOnlyManifest(t *testing.T, repo string) packageRepositoryManifestDiagnostics {
	t.Helper()
	manifest, blockers := buildTestOnlyPackageRepositoryManifest(repo)
	if len(blockers) != 0 {
		t.Fatalf("test-only manifest builder must not report blockers for a valid temp repo: %#v", blockers)
	}
	data, err := marshalPackageRepositoryManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatalf("builder must emit valid JSON: %s", string(data))
	}
	writeTestOnlyManifestJSON(t, repo, data)
	return diagnosePackageRepositoryManifest(repo)
}

func writeTestOnlyManifestJSON(t *testing.T, repo string, data []byte) {
	t.Helper()
	manifestPath := filepath.Join(repo, packageRepositoryManifestPath)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertTestOnlyManifestRefs(t *testing.T, manifest packageRepositoryManifest) {
	t.Helper()
	for _, want := range strings.Fields(testOnlyManifestExpectedRefs()) {
		if !containsLifecycleTestString(packageRepositoryManifestRefs(manifest), want) {
			t.Fatalf("builder manifest missing %q: %#v", want, manifest)
		}
	}
}

func testOnlyManifestExpectedRefs() string {
	return "bin/findx-categraf-linux-amd64 bin/findx-categraf-windows-amd64.exe bin/findx-catpaw-linux-amd64 bin/findx-catpaw-windows-amd64.exe checksums/SHA256SUMS signatures/SHA256SUMS.asc signatures/test-public-key.asc signatures/test-key-fingerprint.txt commands/install.sh commands/uninstall.sh commands/configure.sh commands/plugins.sh kubernetes/cluster.ref kubernetes/namespace.ref kubernetes/workload-selector.ref kubernetes/helm/findx-agent-chart.ref kubernetes/manifests/findx-agent-bundle.ref kubernetes/helm/values.ref kubernetes/rbac.ref kubernetes/service-account.ref kubernetes/images/findx-agent.ref kubernetes/config-map.ref kubernetes/se" + "cret.ref kubernetes/rollout-strategy.ref kubernetes/rollout-receipt.ref validators/data-arrival.sh audit/package-repository-audit.json evidence/package-repository-evidence.json"
}
