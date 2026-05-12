package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

var monitoringContractSeedIDs = []string{
	"FX-CONTRACT-N9E-DATASOURCE-BRIEF-LIST",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-BY-ID",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-LABELS",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-LABEL-VALUES",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-METRIC-NAMES",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-SERIES",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-BUILDINFO",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-QUERY",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-QUERY-RANGE",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-ES-SEARCH",
	"FX-CONTRACT-N9E-DATASOURCE-TEST-CONNECTION",
	"FX-CONTRACT-N9E-SYSTEM-INTEGRATION-CATALOG",
	"FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT",
	"FX-CONTRACT-N9E-TEMPLATE-CENTER-BUILTIN-BOARD-DETAIL",
	"FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT-BATCH-RESULT",
	"FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT-CONFLICT-ROLLBACK",
	"FX-CONTRACT-N9E-TEMPLATE-CENTER-DOCUMENT-DRAWER",
	"FX-CONTRACT-FINDX-PROMETHEUS-SINGLE-QUERY",
	"FX-CONTRACT-N9E-METRIC-VIEWS-CRUD",
	"FX-CONTRACT-N9E-METRIC-VIEWS-LIST",
	"FX-CONTRACT-N9E-METRIC-VIEWS-CREATE",
	"FX-CONTRACT-N9E-METRIC-VIEWS-UPDATE",
	"FX-CONTRACT-N9E-METRIC-VIEWS-DELETE",
	"FX-CONTRACT-N9E-METRIC-QUERY-BATCH",
	"FX-CONTRACT-N9E-QUERY-RANGE-BATCH",
	"FX-CONTRACT-N9E-QUERY-INSTANT-BATCH",
	"FX-CONTRACT-N9E-PLUS-QUERY-BATCH",
	"FX-CONTRACT-N9E-TAG-PAIRS",
	"FX-CONTRACT-N9E-TAG-METRICS",
	"FX-CONTRACT-N9E-QUERY-DATA",
	"FX-CONTRACT-N9E-QUERY-BENCH",
	"FX-CONTRACT-N9E-PROMETHEUS-COMPAT-API",
	"FX-CONTRACT-N9E-SHARE-CHARTS",
	"FX-CONTRACT-N9E-METRICS-DESC",
	"FX-CONTRACT-N9E-DASHBOARD-CRUD",
	"FX-CONTRACT-N9E-DASHBOARD-LIST-BY-BUSI-GROUP",
	"FX-CONTRACT-N9E-DASHBOARD-LIST-BY-BUSI-GROUPS",
	"FX-CONTRACT-N9E-DASHBOARD-CREATE",
	"FX-CONTRACT-N9E-DASHBOARD-DETAIL",
	"FX-CONTRACT-N9E-DASHBOARD-UPDATE-METADATA",
	"FX-CONTRACT-N9E-DASHBOARD-UPDATE-CONFIGS",
	"FX-CONTRACT-N9E-DASHBOARD-CLONE",
	"FX-CONTRACT-N9E-DASHBOARD-CLONES",
	"FX-CONTRACT-N9E-DASHBOARD-DELETE",
	"FX-CONTRACT-N9E-DASHBOARD-PURE-DETAIL",
	"FX-CONTRACT-N9E-DASHBOARD-NAMES",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-LIST",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-CREATE",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-UPDATE",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-DELETE",
	"FX-CONTRACT-N9E-DASHBOARD-PUBLIC-LIST",
	"FX-CONTRACT-N9E-DASHBOARD-PUBLIC-UPDATE",
	"FX-CONTRACT-N9E-DASHBOARD-EXPORT",
	"FX-CONTRACT-N9E-DASHBOARD-MIGRATE",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUPS",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUP-LIST",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUP-CREATE",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUP-DETAIL",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUP-UPDATE",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUP-DELETE",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUP-RULES-LIST",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUPS-MULTI-RULES-LIST",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITES-LIST",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITE-ADD",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITE-DELETE",
	"FX-CONTRACT-N9E-ALERT-RULE-LIFECYCLE",
	"FX-CONTRACT-N9E-ALERT-RULE-DETAIL",
	"FX-CONTRACT-N9E-ALERT-RULE-PURE-DETAIL",
	"FX-CONTRACT-N9E-ALERT-RULE-CREATE",
	"FX-CONTRACT-N9E-ALERT-RULE-UPDATE",
	"FX-CONTRACT-N9E-ALERT-RULE-DELETE-BY-BUSI-GROUP",
	"FX-CONTRACT-N9E-ALERT-RULE-DELETE-BY-RULE-GROUP",
	"FX-CONTRACT-N9E-ALERT-RULE-IMPORT-JSON",
	"FX-CONTRACT-N9E-ALERT-RULE-IMPORT-PROM-RULE",
	"FX-CONTRACT-N9E-ALERT-RULE-BULK-FIELDS-UPDATE",
	"FX-CONTRACT-N9E-ALERT-RULE-ENABLE-DISABLE",
	"FX-CONTRACT-N9E-ALERT-RULE-STATUS-BATCH",
	"FX-CONTRACT-N9E-ALERT-RULE-CLONE-TO-HOSTS",
	"FX-CONTRACT-N9E-ALERT-RULE-CLONE-TO-BUSI-GROUPS",
	"FX-CONTRACT-N9E-ALERT-RULE-VALIDATE",
	"FX-CONTRACT-N9E-ALERT-RULE-ENABLE-TRYRUN",
	"FX-CONTRACT-N9E-ALERT-RULE-CALLBACKS-LIST",
	"FX-CONTRACT-N9E-ALERT-RULE-TIMEZONES",
	"FX-CONTRACT-N9E-NOTIFICATION-FINDX-ADAPTER",
	"FX-CONTRACT-N9E-ALERT-MUTE-SHIELD",
	"FX-CONTRACT-N9E-ALERT-MUTE-LIST-BY-BUSI-GROUP",
	"FX-CONTRACT-N9E-ALERT-MUTE-LIST-BY-BUSI-GROUPS",
	"FX-CONTRACT-N9E-ALERT-MUTE-DETAIL",
	"FX-CONTRACT-N9E-ALERT-MUTE-CREATE",
	"FX-CONTRACT-N9E-ALERT-MUTE-UPDATE",
	"FX-CONTRACT-N9E-ALERT-MUTE-DELETE",
	"FX-CONTRACT-N9E-ALERT-MUTE-BULK-FIELDS-UPDATE",
	"FX-CONTRACT-N9E-ALERT-MUTE-PREVIEW-EVENTS",
	"FX-CONTRACT-N9E-ALERT-MUTE-TRYRUN",
	"FX-CONTRACT-N9E-BUSI-GROUP-RESOURCE-GROUP-MAP",
}

func TestMonitoringContractMatrixSeedContainsRequiredGaps(t *testing.T) {
	ResetContractMatrixForTest()

	items, err := ListContractMatrixEntries("", "monitoring")
	if err != nil {
		t.Fatalf("list monitoring contract matrix: %v", err)
	}
	got := map[string]model.ContractMatrixEntry{}
	for _, item := range items {
		got[item.ID] = item
	}

	for _, id := range monitoringContractSeedIDs {
		item, ok := got[id]
		if !ok {
			t.Fatalf("missing monitoring contract seed %s", id)
		}
		assertMonitoringContractSeedEntry(t, item)
	}
}

func TestMonitoringContractMatrixDomainFilter(t *testing.T) {
	ResetContractMatrixForTest()
	_, err := SaveContractMatrixEntry(model.ContractMatrixRegisterRequest{
		ID:         "FX-CONTRACT-AGENT-EXECUTOR-SSH",
		Domain:     "agent",
		Capability: "ssh executor",
		Status:     model.ContractStatusMissingExecutor,
	})
	if err != nil {
		t.Fatalf("save non-monitoring entry: %v", err)
	}

	items, err := ListContractMatrixEntries("", "monitoring")
	if err != nil {
		t.Fatalf("filter monitoring contract matrix: %v", err)
	}
	for _, item := range items {
		if item.Domain != "monitoring" {
			t.Fatalf("domain filter returned non-monitoring entry: %#v", item)
		}
	}
	if len(items) < len(monitoringContractSeedIDs) {
		t.Fatalf("domain filter did not include monitoring seeds, got %d", len(items))
	}
}

func TestMonitoringContractMatrixSeedDoesNotOverwriteUserEntry(t *testing.T) {
	ResetContractMatrixForTest()
	const customReason = "operator registered custom contract gap"
	_, err := SaveContractMatrixEntry(model.ContractMatrixRegisterRequest{
		ID:            "FX-CONTRACT-N9E-METRIC-VIEWS-CRUD",
		Domain:        "monitoring",
		Capability:    "custom metric views",
		Status:        model.ContractStatusBlocked,
		SourceRefs:    []string{"custom-source-ref"},
		BlockedReason: customReason,
		Metadata:      map[string]string{"findx_route": "/custom-monitoring"},
	})
	if err != nil {
		t.Fatalf("save user entry: %v", err)
	}

	item, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-METRIC-VIEWS-CRUD")
	if err != nil {
		t.Fatalf("get seeded id after user entry: %v", err)
	}
	if !ok {
		t.Fatal("expected user entry to remain available")
	}
	if item.BlockedReason != customReason || item.Metadata["findx_route"] != "/custom-monitoring" {
		t.Fatalf("seed overwrote user entry: %#v", item)
	}
}

func TestMonitoringContractMatrixAlertMuteShieldSourceRefs(t *testing.T) {
	got := alertMuteShieldSourceRefs()
	want := []string{
		`D:\项目迁移文件\平台源码\fe-main\src\services\shield.ts`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\add.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\edit.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\components\operateForm.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\components\PreviewMutedEvents.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\components\utils.ts`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\warning\shield\components\CateSelect\index.tsx`,
	}
	if len(got) != len(want) {
		t.Fatalf("alert mute shield source refs count = %d, want %d: %#v", len(got), len(want), got)
	}
	for _, ref := range want {
		if !contractListContains(got, ref) {
			t.Fatalf("alert mute shield source refs missing %q: %#v", ref, got)
		}
	}
}

func assertMonitoringN9eGapMetadata(t *testing.T, item model.ContractMatrixEntry, upstreamRef string) {
	t.Helper()
	if item.Metadata["upstream_ref"] != upstreamRef {
		t.Fatalf("%s upstream_ref = %q, want %q", item.ID, item.Metadata["upstream_ref"], upstreamRef)
	}
	for _, key := range []string{"findx_route", "gap_type", "upstream_ref"} {
		if strings.TrimSpace(item.Metadata[key]) == "" {
			t.Fatalf("%s missing metadata %s: %#v", item.ID, key, item.Metadata)
		}
	}
	if strings.TrimSpace(item.BlockedReason) == "" {
		t.Fatalf("%s missing blocked_reason", item.ID)
	}
	if len(item.SourceRefs) == 0 {
		t.Fatalf("%s missing source_refs", item.ID)
	}
}

func TestContractMatrixMetadataDropsSensitiveAndFakeSuccessText(t *testing.T) {
	ResetContractMatrixForTest()
	item, err := SaveContractMatrixEntry(model.ContractMatrixRegisterRequest{
		ID:         "FX-CONTRACT-N9E-SAFE-METADATA",
		Domain:     "monitoring",
		Capability: "metadata sanitize",
		Status:     model.ContractStatusMissingBackend,
		Metadata: map[string]string{
			"findx_route": "/monitoring/safe",
			"password":    "hidden",
			"phase":       "running",
			"result":      "succeeded",
			"note":        "queued for install",
		},
	})
	if err != nil {
		t.Fatalf("save sanitized metadata entry: %v", err)
	}
	if item.Metadata["findx_route"] != "/monitoring/safe" {
		t.Fatalf("safe metadata was dropped: %#v", item.Metadata)
	}
	for _, key := range []string{"password", "phase", "result", "note"} {
		if _, ok := item.Metadata[key]; ok {
			t.Fatalf("unsafe metadata key/value survived: %s=%q in %#v", key, item.Metadata[key], item.Metadata)
		}
	}
}

func contractListContains(values []string, needle string) bool {
	for _, value := range values {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func monitoringMatureSourceRefs(relativePaths ...string) []string {
	refs := make([]string, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		refs = append(refs, `D:\项目迁移文件\平台源码\`+relativePath)
	}
	return refs
}

func monitoringMojibakeDenylist() []string {
	return []string{
		string([]rune{0x6924, 0x572D, 0x6D30}),
		string([]rune{0x6769, 0x4F7A, 0x0429}),
		string([]rune{0x9A9E, 0x51B2, 0x5F74}),
		string([]rune{0x5A67, 0x612E, 0x721C}),
		string([]rune{0x9473, 0x85C9, 0x59CF}),
		string([]rune{0x7F02, 0x54C4, 0x76AF}),
		string([]rune{0x0044, 0x951B}),
		string([]rune{0x0044, 0xF03A}),
		"\ufffd",
	}
}

func assertMonitoringContractSeedEntry(t *testing.T, item model.ContractMatrixEntry) {
	t.Helper()
	assertMonitoringContractSeedHasNoMojibake(t, item)
	if item.ID == "FX-CONTRACT-N9E-DATASOURCE-BRIEF-LIST" || item.ID == "FX-CONTRACT-FINDX-PROMETHEUS-SINGLE-QUERY" {
		assertReadyMonitoringContractSeedEntry(t, item)
		return
	}
	if !model.IsContractMatrixStatus(item.Status) {
		t.Fatalf("monitoring seed used invalid status: %#v", item)
	}
	if item.Status == model.ContractStatusReady {
		t.Fatalf("unexpected ready monitoring seed: %#v", item)
	}
	if len(item.SourceRefs) == 0 {
		t.Fatalf("monitoring seed missing source refs: %#v", item)
	}
	if item.Metadata["findx_route"] == "" {
		t.Fatalf("monitoring seed missing findx_route metadata: %#v", item)
	}
	if item.Metadata["gap_type"] == "" && item.Metadata["findx_adapter"] == "" {
		t.Fatalf("monitoring seed missing gap metadata: %#v", item)
	}
	body := strings.ToLower(strings.Join(append(item.SourceRefs, item.BlockedReason), " "))
	for _, forbidden := range []string{"queued", "running", "succeeded", "success", "applied", "rolled-back", "installed", "data_arrived"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("monitoring seed exposed fake success state %q: %#v", forbidden, item)
		}
	}
}

func assertMonitoringContractSeedHasNoMojibake(t *testing.T, item model.ContractMatrixEntry) {
	t.Helper()
	values := []string{item.Capability, item.BlockedReason, item.Handler, item.Backend, item.Datasource, item.Executor}
	values = append(values, item.SourceRefs...)
	values = append(values, item.EvidenceRefs...)
	for key, value := range item.Metadata {
		values = append(values, key, value)
	}
	for _, value := range values {
		for _, mojibake := range monitoringMojibakeDenylist() {
			if strings.Contains(value, mojibake) {
				t.Fatalf("%s contains mojibake %q in value %q: %#v", item.ID, mojibake, value, item)
			}
		}
	}
}

func assertReadyMonitoringContractSeedEntry(t *testing.T, item model.ContractMatrixEntry) {
	t.Helper()
	if item.Status != model.ContractStatusReady || !item.SafeToRetry {
		t.Fatalf("ready monitoring contract should be ready and retryable: %#v", item)
	}
	for name, value := range map[string]string{
		"handler":    item.Handler,
		"backend":    item.Backend,
		"datasource": item.Datasource,
		"executor":   item.Executor,
	} {
		if strings.TrimSpace(value) == "" {
			t.Fatalf("ready contract missing %s: %#v", name, item)
		}
	}
	if len(item.EvidenceRefs) == 0 {
		t.Fatalf("ready contract missing evidence refs: %#v", item)
	}
	switch item.ID {
	case "FX-CONTRACT-N9E-DATASOURCE-BRIEF-LIST":
		if item.Metadata["upstream_ref"] != "/monitor/datasources" {
			t.Fatalf("ready datasource brief must point to FindX adapter, got %#v", item.Metadata)
		}
	case "FX-CONTRACT-FINDX-PROMETHEUS-SINGLE-QUERY":
		if item.Metadata["upstream_ref"] != "/monitor/query,/monitor/query-range,/monitor/labels,/monitor/label-values" {
			t.Fatalf("single query adapter must point only to FindX monitor query endpoints, got %#v", item.Metadata)
		}
		scope := strings.ToLower(item.Metadata["upstream_scope"])
		if !strings.Contains(scope, "single prometheus query") || !strings.Contains(scope, "not batch query") || !strings.Contains(scope, "metric views crud") {
			t.Fatalf("single query adapter scope must not imply Nightingale batch or metric views readiness: %#v", item.Metadata)
		}
		for _, forbidden := range []string{"query-range-batch", "query-instant-batch", "/api/n9e-plus/query-batch", "/api/n9e/tag-pairs", "/api/n9e/tag-metrics", "/api/n9e/query", "/api/n9e/metric-views", "/api/n9e/prometheus/api/v1", "/api/n9e/share-charts", "/api/n9e/metrics/desc"} {
			if strings.Contains(item.Metadata["upstream_ref"], forbidden) {
				t.Fatalf("single query adapter upstream_ref must not claim batch/query view endpoint %s: %#v", forbidden, item.Metadata)
			}
		}
	}
}
