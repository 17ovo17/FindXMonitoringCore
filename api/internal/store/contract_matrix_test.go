package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

var monitoringContractSeedIDs = []string{
	"FX-CONTRACT-N9E-DATASOURCE-BRIEF-LIST",
	"FX-CONTRACT-N9E-DATASOURCE-PROXY-BY-ID",
	"FX-CONTRACT-N9E-DATASOURCE-TEST-CONNECTION",
	"FX-CONTRACT-N9E-SYSTEM-INTEGRATION-CATALOG",
	"FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT",
	"FX-CONTRACT-FINDX-PROMETHEUS-SINGLE-QUERY",
	"FX-CONTRACT-N9E-METRIC-VIEWS-CRUD",
	"FX-CONTRACT-N9E-METRIC-QUERY-BATCH",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS",
	"FX-CONTRACT-N9E-ALERT-RULE-GROUPS",
	"FX-CONTRACT-N9E-NOTIFICATION-FINDX-ADAPTER",
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

func TestMonitoringContractMatrixPrometheusSingleQueryReadyIsScoped(t *testing.T) {
	ResetContractMatrixForTest()

	item, ok, err := GetContractMatrixEntry("FX-CONTRACT-FINDX-PROMETHEUS-SINGLE-QUERY")
	if err != nil {
		t.Fatalf("get single query adapter seed: %v", err)
	}
	if !ok {
		t.Fatal("single query adapter seed missing")
	}
	assertReadyMonitoringContractSeedEntry(t, item)
	for name, value := range map[string]string{
		"handler":        item.Handler,
		"backend":        item.Backend,
		"datasource":     item.Datasource,
		"executor":       item.Executor,
		"upstream_scope": item.Metadata["upstream_scope"],
	} {
		if strings.TrimSpace(value) == "" {
			t.Fatalf("single query adapter missing %s: %#v", name, item)
		}
	}
	if len(item.EvidenceRefs) < 4 {
		t.Fatalf("single query adapter evidence should include route, handler, gateway, and tests: %#v", item)
	}
	wantSourceRefs := monitoringMatureSourceRefs(
		`fe-main\src\services\dashboardV2.ts`,
		`fe-main\src\services\metricViews.ts`,
		`fe-main\src\services\metric.ts`,
	)
	for _, want := range wantSourceRefs {
		if !contractListContains(item.SourceRefs, want) {
			t.Fatalf("single query adapter source_refs missing mature source %s: %#v", want, item.SourceRefs)
		}
	}
	for _, ref := range item.SourceRefs {
		for _, mojibake := range monitoringMojibakeDenylist() {
			if strings.Contains(ref, mojibake) {
				t.Fatalf("single query adapter source_refs contains mojibake %q: %#v", mojibake, item.SourceRefs)
			}
		}
	}
	for _, want := range []string{"api/routes_monitor.go", "monitoring_query.go", "monitoring_query_prometheus.go", "monitoring_query_test.go"} {
		if !contractListContains(item.EvidenceRefs, want) {
			t.Fatalf("single query adapter evidence_refs missing FindX evidence %s: %#v", want, item.EvidenceRefs)
		}
	}
}

func TestMonitoringContractMatrixNightingaleBatchAndMetricViewsStayGaps(t *testing.T) {
	ResetContractMatrixForTest()

	tests := []struct {
		id     string
		status string
	}{
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-CRUD", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-METRIC-QUERY-BATCH", status: model.ContractStatusMissingDatasource},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			item, ok, err := GetContractMatrixEntry(tt.id)
			if err != nil {
				t.Fatalf("get %s: %v", tt.id, err)
			}
			if !ok {
				t.Fatalf("%s seed missing", tt.id)
			}
			if item.Status != tt.status {
				t.Fatalf("%s status = %s, want %s", tt.id, item.Status, tt.status)
			}
			if item.Status == model.ContractStatusReady || item.SafeToRetry || item.Handler != "" || item.Backend != "" || item.Datasource != "" || item.Executor != "" || len(item.EvidenceRefs) != 0 {
				t.Fatalf("%s was incorrectly upgraded by single query adapter: %#v", tt.id, item)
			}
		})
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
	}
}

func assertMonitoringContractSeedEntry(t *testing.T, item model.ContractMatrixEntry) {
	t.Helper()
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
		if !strings.Contains(scope, "single prometheus query") || !strings.Contains(scope, "not nightingale batch") || !strings.Contains(scope, "metric views crud") {
			t.Fatalf("single query adapter scope must not imply Nightingale batch or metric views readiness: %#v", item.Metadata)
		}
		for _, forbidden := range []string{"query-range-batch", "query-instant-batch", "/api/n9e/tag-metrics", "/api/n9e/metric-views"} {
			if strings.Contains(item.Metadata["upstream_ref"], forbidden) {
				t.Fatalf("single query adapter upstream_ref must not claim batch/query view endpoint %s: %#v", forbidden, item.Metadata)
			}
		}
	}
}
