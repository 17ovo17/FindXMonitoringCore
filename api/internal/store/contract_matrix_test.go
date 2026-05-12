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
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-LIST",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-CREATE",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-UPDATE",
	"FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-DELETE",
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
		id          string
		status      string
		upstreamRef string
	}{
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-CRUD", status: model.ContractStatusBlocked, upstreamRef: "metric-views-crud-aggregate"},
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-LIST", status: model.ContractStatusMissingBackend, upstreamRef: "GET /api/n9e/metric-views"},
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-CREATE", status: model.ContractStatusMissingBackend, upstreamRef: "POST /api/n9e/metric-views"},
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-UPDATE", status: model.ContractStatusMissingBackend, upstreamRef: "PUT /api/n9e/metric-views"},
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-DELETE", status: model.ContractStatusMissingBackend, upstreamRef: "DELETE /api/n9e/metric-views"},
		{id: "FX-CONTRACT-N9E-METRIC-QUERY-BATCH", status: model.ContractStatusBlocked, upstreamRef: "metric-query-aggregate"},
		{id: "FX-CONTRACT-N9E-QUERY-RANGE-BATCH", status: model.ContractStatusMissingDatasource, upstreamRef: "/api/{N9E_PATHNAME}/query-range-batch"},
		{id: "FX-CONTRACT-N9E-QUERY-INSTANT-BATCH", status: model.ContractStatusMissingDatasource, upstreamRef: "/api/{N9E_PATHNAME}/query-instant-batch"},
		{id: "FX-CONTRACT-N9E-PLUS-QUERY-BATCH", status: model.ContractStatusMissingDatasource, upstreamRef: "/api/n9e-plus/query-batch"},
		{id: "FX-CONTRACT-N9E-TAG-PAIRS", status: model.ContractStatusMissingDatasource, upstreamRef: "/api/n9e/tag-pairs"},
		{id: "FX-CONTRACT-N9E-TAG-METRICS", status: model.ContractStatusMissingDatasource, upstreamRef: "/api/n9e/tag-metrics"},
		{id: "FX-CONTRACT-N9E-QUERY-DATA", status: model.ContractStatusMissingDatasource, upstreamRef: "/api/n9e/query"},
		{id: "FX-CONTRACT-N9E-QUERY-BENCH", status: model.ContractStatusMissingBackend, upstreamRef: "/api/n9e/query-bench"},
		{id: "FX-CONTRACT-N9E-PROMETHEUS-COMPAT-API", status: model.ContractStatusMissingDatasource, upstreamRef: "/api/n9e/prometheus/api/v1/{path}"},
		{id: "FX-CONTRACT-N9E-SHARE-CHARTS", status: model.ContractStatusMissingBackend, upstreamRef: "/api/n9e/share-charts"},
		{id: "FX-CONTRACT-N9E-METRICS-DESC", status: model.ContractStatusMissingBackend, upstreamRef: "/api/n9e/metrics/desc"},
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
			assertMonitoringN9eGapMetadata(t, item, tt.upstreamRef)
		})
	}
}

func TestMonitoringContractMatrixMetricViewsCrudIsSplitIntoPreciseGaps(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-METRIC-VIEWS-CRUD")
	if err != nil {
		t.Fatalf("get metric views aggregate: %v", err)
	}
	if !ok {
		t.Fatal("metric views aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked {
		t.Fatalf("metric views aggregate status = %s, want blocked", aggregate.Status)
	}
	if aggregate.Metadata["upstream_ref"] != "metric-views-crud-aggregate" {
		t.Fatalf("aggregate must not own endpoint directly: %#v", aggregate.Metadata)
	}
	assertMonitoringMetricViewsSplitGap(t, "FX-CONTRACT-N9E-METRIC-VIEWS-LIST", "GET /api/n9e/metric-views")
	assertMonitoringMetricViewsSplitGap(t, "FX-CONTRACT-N9E-METRIC-VIEWS-CREATE", "POST /api/n9e/metric-views")
	assertMonitoringMetricViewsSplitGap(t, "FX-CONTRACT-N9E-METRIC-VIEWS-UPDATE", "PUT /api/n9e/metric-views")
	assertMonitoringMetricViewsSplitGap(t, "FX-CONTRACT-N9E-METRIC-VIEWS-DELETE", "DELETE /api/n9e/metric-views")
}

func TestMonitoringContractMatrixDatasourceProxyIsSplitIntoPreciseGaps(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-DATASOURCE-PROXY-BY-ID")
	if err != nil {
		t.Fatalf("get datasource proxy aggregate: %v", err)
	}
	if !ok {
		t.Fatal("datasource proxy aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked {
		t.Fatalf("datasource proxy aggregate status = %s, want blocked", aggregate.Status)
	}
	if aggregate.Metadata["upstream_ref"] != "datasource-proxy-aggregate" {
		t.Fatalf("aggregate must not own datasource proxy endpoint directly: %#v", aggregate.Metadata)
	}
	if strings.Contains(aggregate.Metadata["upstream_ref"], "/api/n9e/proxy/{datasource_id}") {
		t.Fatalf("aggregate must point to child gaps, not direct endpoint ownership: %#v", aggregate.Metadata)
	}
	if !strings.Contains(aggregate.Metadata["upstream_scope"], "labels") ||
		!strings.Contains(aggregate.Metadata["upstream_scope"], "es_search") {
		t.Fatalf("aggregate scope must enumerate datasource proxy child gaps: %#v", aggregate.Metadata)
	}

	tests := []struct {
		id          string
		route       string
		upstreamRef string
		sourceRefs  []string
	}{
		{id: "FX-CONTRACT-N9E-DATASOURCE-PROXY-LABELS", route: "/query?section=metrics", upstreamRef: "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/labels", sourceRefs: monitoringMatureSourceRefs(`fe-main\src\services\dashboardV2.ts`, `fe-main\src\services\metricViews.ts`)},
		{id: "FX-CONTRACT-N9E-DATASOURCE-PROXY-LABEL-VALUES", route: "/query?section=metrics", upstreamRef: "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/label/{label}/values", sourceRefs: monitoringMatureSourceRefs(`fe-main\src\services\dashboardV2.ts`, `fe-main\src\services\metricViews.ts`)},
		{id: "FX-CONTRACT-N9E-DATASOURCE-PROXY-METRIC-NAMES", route: "/query?section=metrics", upstreamRef: "GET /api/{N9E_PATHNAME}/proxy/{datasource_id}/api/v1/label/__name__/values", sourceRefs: monitoringMatureSourceRefs(`fe-main\src\services\dashboardV2.ts`, `fe-main\src\services\metricViews.ts`)},
		{id: "FX-CONTRACT-N9E-DATASOURCE-PROXY-SERIES", route: "/query?section=metrics", upstreamRef: "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/series", sourceRefs: monitoringMatureSourceRefs(`fe-main\src\services\dashboardV2.ts`)},
		{id: "FX-CONTRACT-N9E-DATASOURCE-PROXY-BUILDINFO", route: "/query?section=metrics", upstreamRef: "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/status/buildinfo", sourceRefs: monitoringMatureSourceRefs(`fe-main\src\services\dashboardV2.ts`)},
		{id: "FX-CONTRACT-N9E-DATASOURCE-PROXY-QUERY", route: "/query?section=metrics", upstreamRef: "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/query", sourceRefs: monitoringMatureSourceRefs(`fe-main\src\services\dashboardV2.ts`)},
		{id: "FX-CONTRACT-N9E-DATASOURCE-PROXY-QUERY-RANGE", route: "/query?section=metric-views", upstreamRef: "GET /api/{N9E_PATHNAME}/proxy/{datasourceValue}/api/v1/query_range", sourceRefs: monitoringMatureSourceRefs(`fe-main\src\services\metricViews.ts`)},
		{id: "FX-CONTRACT-N9E-DATASOURCE-PROXY-ES-SEARCH", route: "/query?section=metrics", upstreamRef: "POST /api/{N9E_PATHNAME}/proxy/{datasourceValue}/{index}/_search", sourceRefs: monitoringMatureSourceRefs(`fe-main\src\services\dashboardV2.ts`)},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			item := assertMonitoringDatasourceProxySplitGap(t, tt.id, tt.route, tt.upstreamRef)
			for _, want := range tt.sourceRefs {
				if !contractListContains(item.SourceRefs, want) {
					t.Fatalf("%s source_refs missing mature source %s: %#v", tt.id, want, item.SourceRefs)
				}
			}
		})
	}
}

func TestMonitoringContractMatrixDashboardAnnotationsIsSplitIntoPreciseGaps(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS")
	if err != nil {
		t.Fatalf("get dashboard annotations aggregate: %v", err)
	}
	if !ok {
		t.Fatal("dashboard annotations aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked {
		t.Fatalf("dashboard annotations aggregate status = %s, want blocked", aggregate.Status)
	}
	if aggregate.SafeToRetry || aggregate.Handler != "" || aggregate.Backend != "" || aggregate.Datasource != "" || aggregate.Executor != "" || len(aggregate.EvidenceRefs) != 0 {
		t.Fatalf("dashboard annotations aggregate should remain non-ready without executable evidence: %#v", aggregate)
	}
	if aggregate.Metadata["upstream_ref"] != "dashboard-annotations-aggregate" {
		t.Fatalf("aggregate must not own dashboard annotations endpoint directly: %#v", aggregate.Metadata)
	}
	if strings.Contains(aggregate.Metadata["upstream_ref"], "/api/n9e/dashboard-annotations") ||
		strings.Contains(aggregate.Metadata["upstream_ref"], "/api/n9e/dashboard-annotation") {
		t.Fatalf("aggregate must point to child gaps, not direct endpoint ownership: %#v", aggregate.Metadata)
	}
	if !strings.Contains(aggregate.Metadata["upstream_scope"], "list") ||
		!strings.Contains(aggregate.Metadata["upstream_scope"], "create") ||
		!strings.Contains(aggregate.Metadata["upstream_scope"], "update") ||
		!strings.Contains(aggregate.Metadata["upstream_scope"], "delete") {
		t.Fatalf("aggregate scope must enumerate annotations child gaps: %#v", aggregate.Metadata)
	}
	assertMonitoringN9eGapMetadata(t, aggregate, "dashboard-annotations-aggregate")

	tests := []struct {
		id          string
		upstreamRef string
	}{
		{id: "FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-LIST", upstreamRef: "GET /api/n9e/dashboard-annotations"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-CREATE", upstreamRef: "POST /api/n9e/dashboard-annotations"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-UPDATE", upstreamRef: "PUT /api/n9e/dashboard-annotation/{id}"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-ANNOTATIONS-DELETE", upstreamRef: "DELETE /api/n9e/dashboard-annotation/{id}"},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			item := assertMonitoringDashboardAnnotationsSplitGap(t, tt.id, tt.upstreamRef)
			for _, want := range monitoringMatureSourceRefs(`fe-main\src\services\dashboardV2.ts`) {
				if !contractListContains(item.SourceRefs, want) {
					t.Fatalf("%s source_refs missing mature source %s: %#v", tt.id, want, item.SourceRefs)
				}
			}
		})
	}
}

func assertMonitoringMetricViewsSplitGap(t *testing.T, id, upstreamRef string) {
	t.Helper()
	item, ok, err := GetContractMatrixEntry(id)
	if err != nil {
		t.Fatalf("get %s: %v", id, err)
	}
	if !ok {
		t.Fatalf("%s seed missing", id)
	}
	if item.Status == model.ContractStatusReady || item.SafeToRetry || item.Handler != "" || item.Backend != "" || item.Datasource != "" || item.Executor != "" || len(item.EvidenceRefs) != 0 {
		t.Fatalf("%s should remain non-ready without executable evidence: %#v", id, item)
	}
	assertMonitoringN9eGapMetadata(t, item, upstreamRef)
}

func assertMonitoringDatasourceProxySplitGap(t *testing.T, id, findxRoute, upstreamRef string) model.ContractMatrixEntry {
	t.Helper()
	item, ok, err := GetContractMatrixEntry(id)
	if err != nil {
		t.Fatalf("get %s: %v", id, err)
	}
	if !ok {
		t.Fatalf("%s seed missing", id)
	}
	if item.Status != model.ContractStatusMissingDatasource {
		t.Fatalf("%s status = %s, want missing_datasource", id, item.Status)
	}
	if item.SafeToRetry || item.Handler != "" || item.Backend != "" || item.Datasource != "" || item.Executor != "" || len(item.EvidenceRefs) != 0 {
		t.Fatalf("%s should remain non-ready without executable evidence: %#v", id, item)
	}
	if item.Metadata["findx_route"] != findxRoute {
		t.Fatalf("%s findx_route = %q, want %q", id, item.Metadata["findx_route"], findxRoute)
	}
	if strings.Contains(item.Metadata["upstream_ref"], "query-range-batch") ||
		strings.Contains(item.Metadata["upstream_ref"], "query-instant-batch") ||
		strings.Contains(item.Metadata["upstream_ref"], "/api/n9e-plus/query-batch") {
		t.Fatalf("%s must not duplicate batch/query compatibility gaps: %#v", id, item.Metadata)
	}
	assertMonitoringN9eGapMetadata(t, item, upstreamRef)
	return item
}

func assertMonitoringDashboardAnnotationsSplitGap(t *testing.T, id, upstreamRef string) model.ContractMatrixEntry {
	t.Helper()
	item, ok, err := GetContractMatrixEntry(id)
	if err != nil {
		t.Fatalf("get %s: %v", id, err)
	}
	if !ok {
		t.Fatalf("%s seed missing", id)
	}
	if item.Status != model.ContractStatusMissingBackend {
		t.Fatalf("%s status = %s, want missing_backend", id, item.Status)
	}
	if item.SafeToRetry || item.Handler != "" || item.Backend != "" || item.Datasource != "" || item.Executor != "" || len(item.EvidenceRefs) != 0 {
		t.Fatalf("%s should remain non-ready without executable evidence: %#v", id, item)
	}
	if item.Metadata["findx_route"] != "/dashboards?section=list" {
		t.Fatalf("%s findx_route = %q, want /dashboards?section=list", id, item.Metadata["findx_route"])
	}
	assertMonitoringN9eGapMetadata(t, item, upstreamRef)
	return item
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
