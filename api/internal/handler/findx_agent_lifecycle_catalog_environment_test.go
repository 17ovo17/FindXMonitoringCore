package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestFindXAgentPackageEnvironmentMatrixFieldsAndMethods(t *testing.T) {
	for _, pkg := range findXAgentPackages() {
		if len(pkg.EnvironmentMatrix) != len(packageEnvironmentMethods) {
			t.Fatalf("package %s matrix rows=%d", pkg.ID, len(pkg.EnvironmentMatrix))
		}
		seen := map[string]bool{}
		for _, row := range pkg.EnvironmentMatrix {
			assertEnvironmentMatrixRowComplete(t, pkg.ID, row)
			seen[row.InstallMethod] = true
		}
		for _, method := range packageEnvironmentMethods {
			if !seen[method.method] {
				t.Fatalf("package %s missing method %s", pkg.ID, method.method)
			}
		}
	}
}

func TestFindXAgentPackageEnvironmentMatrixProbeSourcesStayMissing(t *testing.T) {
	for _, id := range []string{
		"java-app", "python-app", "nodejs-app", "php-app", "go-app",
		"rust-app", "ruby-app", "gateway-probe", "browser-client",
	} {
		pkg := mustFindEnvironmentPackage(t, id)
		for _, row := range pkg.EnvironmentMatrix {
			if row.SourceState != environmentSourceMissing ||
				row.PackageState != environmentPackageMissing ||
				!strings.Contains(row.Blocker, agentBlocked) {
				t.Fatalf("probe package %s must stay missing and blocked: %#v", id, row)
			}
		}
	}
}

func TestFindXAgentPackageEnvironmentMatrixTestOnlyRepositoryPropagation(t *testing.T) {
	repo := t.TempDir()
	for _, ref := range testOnlyPackageRepositoryManifestFileRefs() {
		writePackageRepositoryTestFile(t, repo, ref)
	}
	manifest, blockers := buildTestOnlyPackageRepositoryManifest(repo)
	if len(blockers) != 0 {
		t.Fatalf("test repo fixture invalid: %#v", blockers)
	}
	writePackageRepositoryManifestData(t, repo, manifest)
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	pkg := mustFindEnvironmentPackage(t, "host-collector")
	for _, row := range pkg.EnvironmentMatrix {
		for _, want := range []string{
			environmentTestOnlyRepository,
			environmentTestOnlySignature,
			"PRODUCTION_PACKAGE_REPOSITORY_MISSING",
			"PRODUCTION_SIGNATURE_MISSING",
		} {
			if !environmentRowContains(row, want) {
				t.Fatalf("host matrix missing %s: %#v", want, row)
			}
		}
		if !strings.Contains(row.ToolEvidence, "TOOL_EVIDENCE_TEST_ONLY") &&
			(row.Platform == "linux" || row.Platform == "windows") {
			t.Fatalf("host matrix should expose test-only tool evidence: %#v", row)
		}
		assertEnvironmentRowHasNoFakeSuccess(t, row)
	}
}

func TestFindXAgentPackageEnvironmentMatrixPluginDeliveryContract(t *testing.T) {
	for _, id := range []string{"host-collector", "container-collector"} {
		pkg := mustFindEnvironmentPackage(t, id)
		for _, row := range pkg.EnvironmentMatrix {
			for _, want := range []string{
				"FINDX_AGENT_CONTROL_PLANE_ENTRY",
				"REMOTE_MUTATION_BLOCKED_BY_CONTRACT",
				"RELOAD_BLOCKED_BY_CONTRACT",
				"DRIFT_BLOCKED_BY_CONTRACT",
				"ROLLBACK_BLOCKED_BY_CONTRACT",
				"RECEIPT_BLOCKED_BY_CONTRACT",
			} {
				if !strings.Contains(row.ConfigDelivery, want) {
					t.Fatalf("package %s config delivery missing %s: %#v", id, want, row)
				}
			}
		}
	}
}

func TestFindXAgentPackageEnvironmentMatrixInvalidManifestNotReady(t *testing.T) {
	repo := t.TempDir()
	manifestPath := filepath.Join(repo, packageRepositoryManifestPath)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, []byte("{invalid-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FINDX_AGENT_PACKAGE_REPOSITORY_ROOT", repo)

	pkg := mustFindEnvironmentPackage(t, "host-collector")
	for _, row := range pkg.EnvironmentMatrix {
		if !environmentRowContains(row, "PACKAGE_REPOSITORY_MANIFEST_INVALID") {
			t.Fatalf("invalid manifest blocker must propagate to matrix: %#v", row)
		}
		if environmentRowContains(row, environmentTestOnlyRepository) {
			t.Fatalf("invalid manifest must not synthesize test-only state: %#v", row)
		}
		assertEnvironmentRowHasNoFakeSuccess(t, row)
	}
}

func TestFindXAgentPackageEnvironmentMatrixSanitizedAndBlocked(t *testing.T) {
	rows := []any{}
	for _, pkg := range findXAgentPackages() {
		for _, row := range pkg.EnvironmentMatrix {
			rows = append(rows, row)
			if !environmentRowContains(row, agentBlocked) {
				t.Fatalf("matrix row must stay blocked: package=%s row=%#v", pkg.ID, row)
			}
			assertEnvironmentRowHasNoFakeSuccess(t, row)
		}
	}
	data, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	payload := string(data)
	for _, forbidden := range []string{
		"Categraf", "Catpaw", "SkyWalking", "Nightingale",
		"AKIA", "BEGIN PRIVATE KEY", "Cookie", "Set-Cookie", "DB_DSN",
		"\u93cd", "\u9369", "\u6d93", "\u93bb", "\u95b0", "\u7ecb", "\u7039",
	} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("environment matrix exposes forbidden fragment %q: %s", forbidden, payload)
		}
	}
}

func assertEnvironmentMatrixRowComplete(t *testing.T, packageID string, row any) {
	t.Helper()
	data, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	if err := json.Unmarshal(data, &values); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{
		"platform", "install_method", "tool_evidence", "source_state",
		"package_state", "executor", "service_registration", "config_delivery",
		"uninstall", "rollback", "data_arrival", "blocker",
	} {
		if strings.TrimSpace(values[field]) == "" {
			t.Fatalf("package %s row missing field %s: %s", packageID, field, string(data))
		}
	}
}

func assertEnvironmentRowHasNoFakeSuccess(t *testing.T, row any) {
	t.Helper()
	data, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"queued", "running", "succeeded", "success", "applied", "rolled-back", "ready"} {
		if strings.Contains(strings.ToLower(string(data)), forbidden) {
			t.Fatalf("matrix row must not expose fake success state %q: %s", forbidden, string(data))
		}
	}
}

func mustFindEnvironmentPackage(t *testing.T, id string) model.FindXAgentPackage {
	t.Helper()
	pkg, ok := findAgentPackage(id)
	if !ok {
		t.Fatalf("missing package %s", id)
	}
	return pkg
}

func environmentRowContains(row any, want string) bool {
	data, _ := json.Marshal(row)
	return strings.Contains(string(data), want)
}

func writePackageRepositoryManifestData(t *testing.T, repo string, manifest packageRepositoryManifest) {
	t.Helper()
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
