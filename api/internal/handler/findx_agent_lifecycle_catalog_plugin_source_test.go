package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestFindXAgentCategrafPluginSourceMapCatalog(t *testing.T) {
	spec := mustFindPluginConfig(t, "host-collector")
	if len(spec.PluginSourceMap) < 15 {
		t.Fatalf("plugin source map should expose at least 15 plugins, got %d", len(spec.PluginSourceMap))
	}

	plugins := mapPluginSourceSpecs(spec.PluginSourceMap)
	for _, id := range []string{
		"input.cpu", "input.mem", "input.disk", "input.net", "input.processes",
		"input.procstat", "input.docker", "input.mysql", "input.postgresql",
		"input.redis", "input.mongodb", "input.elasticsearch", "input.nginx",
		"input.prometheus", "input.exec",
	} {
		row, ok := plugins[id]
		if !ok {
			t.Fatalf("missing plugin source map row %s", id)
		}
		assertPluginSourceMapRowComplete(t, row)
	}
	if strings.Contains(spec.PluginID, " / ") {
		t.Fatalf("host plugin id should not remain a combined plugin string: %q", spec.PluginID)
	}
}

func TestFindXAgentCategrafPluginSourceMapSecurityGates(t *testing.T) {
	spec := mustFindPluginConfig(t, "host-collector")
	plugins := mapPluginSourceSpecs(spec.PluginSourceMap)

	exec := plugins["input.exec"]
	if !exec.UnsafePlugin || exec.SecurityLevel != "critical" || exec.UnsafePluginPolicyRef == "" {
		t.Fatalf("input.exec must be critical unsafe plugin with policy ref: %#v", exec)
	}
	for _, want := range []string{"UNSAFE_PLUGIN_BLOCKED_BY_CONTRACT", "REMOTE_COMMAND_EXECUTION_NOT_ALLOWED"} {
		if !containsLifecycleTestString(exec.Blockers, want) {
			t.Fatalf("input.exec missing blocker %s: %#v", want, exec.Blockers)
		}
	}
	if spec.SecurityProfile.SecurityLevel != "critical" ||
		!containsLifecycleTestString(spec.SecurityProfile.UnsafePluginIDs, "input.exec") ||
		!containsLifecycleTestString(spec.SecurityProfile.BlockedPluginIDs, "input.exec") {
		t.Fatalf("security profile must classify unsafe exec plugin: %#v", spec.SecurityProfile)
	}

	for _, id := range []string{"input.mysql", "input.postgresql", "input.redis", "input.mongodb"} {
		assertPluginHasBlockers(t, plugins[id],
			"CREDENTIAL_REF_REQUIRED",
			"CREDENTIAL_REF_GATE_MISSING",
			"URL_ALLOWLIST_GATE_MISSING",
			"TLS_AND_TIMEOUT_GATE_MISSING",
			"TLS_GATE_MISSING",
			"TIMEOUT_GATE_MISSING",
			"ERROR_SANITIZATION_GATE_MISSING",
			"DATA_ARRIVAL_RECEIPT_MISSING",
		)
	}
	for _, id := range []string{"input.elasticsearch", "input.nginx", "input.prometheus"} {
		assertPluginHasBlockers(t, plugins[id],
			"CREDENTIAL_REF_GATE_MISSING",
			"URL_ALLOWLIST_GATE_MISSING",
			"TLS_AND_TIMEOUT_GATE_MISSING",
			"TLS_GATE_MISSING",
			"TIMEOUT_GATE_MISSING",
			"ERROR_SANITIZATION_GATE_MISSING",
			"DATA_ARRIVAL_RECEIPT_MISSING",
		)
	}
}

func TestFindXAgentCategrafPluginPlatformMatrixBlocked(t *testing.T) {
	spec := mustFindPluginConfig(t, "host-collector")
	platforms := mapPluginPlatformSpecs(spec.PlatformMatrix)
	for _, platform := range []string{"Linux", "Windows", "Kubernetes"} {
		row, ok := platforms[platform]
		if !ok {
			t.Fatalf("missing platform matrix row %s", platform)
		}
		if row.Status != agentBlocked {
			t.Fatalf("platform row %s must stay blocked: %#v", platform, row)
		}
		assertEnvironmentRowHasNoFakeSuccess(t, row)
	}
	if !containsLifecycleTestString(platforms["Windows"].Blockers, "WINDOWS_HUP_NOT_SUPPORTED") ||
		!containsLifecycleTestString(platforms["Windows"].Blockers, "SERVICE_RESTART_RECEIPT_MISSING") {
		t.Fatalf("windows matrix must reject HUP and require service restart receipt: %#v", platforms["Windows"])
	}
	if !containsLifecycleTestString(platforms["Kubernetes"].Blockers, "CONFIGMAP_WRITER_MISSING") ||
		!containsLifecycleTestString(platforms["Kubernetes"].Blockers, "DAEMONSET_ROLLOUT_RECEIPT_MISSING") {
		t.Fatalf("kubernetes matrix must block on ConfigMap and rollout receipts: %#v", platforms["Kubernetes"])
	}
	if !containsLifecycleTestString(platforms["Linux"].Blockers, "PLUGIN_CONFIG_WRITER_MISSING") ||
		!containsLifecycleTestString(platforms["Linux"].Blockers, "RELOAD_RECEIPT_MISSING") {
		t.Fatalf("linux matrix must stay blocked without writer and reload receipt: %#v", platforms["Linux"])
	}
}

func TestFindXAgentCategrafPluginSourceMapSanitizedAndBlocked(t *testing.T) {
	spec := mustFindPluginConfig(t, "host-collector")
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	payload := string(data)
	for _, forbidden := range []string{
		"Categraf", "Catpaw", "SkyWalking", "Nightingale",
		"AKIA", "BEGIN PRIVATE KEY", "Cookie", "Set-Cookie", "DB_DSN",
		"ready", "success", "applied", "installed", "data_arrived", "queued", "running", "succeeded",
		"\uFFFD", "\u93cd", "\u9369", "\u6d93", "\u93bb", "\u95b0", "\u7ecb", "\u7039",
	} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("plugin source map exposes forbidden fragment %q: %s", forbidden, payload)
		}
	}
	assertLifecycleCatalogNoMojibake(t, spec)
}

func mustFindPluginConfig(t *testing.T, packageID string) *model.FindXAgentPluginConfigSpec {
	t.Helper()
	pkg, ok := findAgentPackage(packageID)
	if !ok {
		t.Fatalf("missing package %s", packageID)
	}
	if pkg.PluginConfig == nil {
		t.Fatalf("package %s missing plugin config", packageID)
	}
	return pkg.PluginConfig
}

func mapPluginSourceSpecs(rows []model.FindXAgentPluginSourceSpec) map[string]model.FindXAgentPluginSourceSpec {
	plugins := make(map[string]model.FindXAgentPluginSourceSpec, len(rows))
	for _, row := range rows {
		plugins[row.PluginID] = row
	}
	return plugins
}

func mapPluginPlatformSpecs(rows []model.FindXAgentPluginPlatformSpec) map[string]model.FindXAgentPluginPlatformSpec {
	platforms := make(map[string]model.FindXAgentPluginPlatformSpec, len(rows))
	for _, row := range rows {
		platforms[row.Platform] = row
	}
	return platforms
}

func assertPluginSourceMapRowComplete(t *testing.T, row model.FindXAgentPluginSourceSpec) {
	t.Helper()
	if row.PluginID == "" ||
		row.PluginCategory == "" ||
		len(row.SourceDirectories) == 0 ||
		len(row.ConfigPaths) == 0 ||
		row.ConfigFormat != pluginConfigFormatTOML ||
		len(row.SupportedPlatforms) == 0 ||
		row.SecurityLevel == "" ||
		row.RemoteMutationStatus != agentBlocked ||
		len(row.Blockers) == 0 ||
		len(row.SourceEvidence) == 0 {
		t.Fatalf("plugin source map row incomplete: %#v", row)
	}
	if !containsLifecycleTestString(row.Blockers, "REMOTE_MUTATION_BLOCKED_BY_CONTRACT") ||
		!containsLifecycleTestString(row.Blockers, "DATA_ARRIVAL_RECEIPT_MISSING") {
		t.Fatalf("plugin source map row missing common gates: %#v", row)
	}
}

func assertPluginHasBlockers(t *testing.T, row model.FindXAgentPluginSourceSpec, blockers ...string) {
	t.Helper()
	for _, blocker := range blockers {
		if !containsLifecycleTestString(row.Blockers, blocker) {
			t.Fatalf("plugin %s missing blocker %s: %#v", row.PluginID, blocker, row.Blockers)
		}
	}
}
