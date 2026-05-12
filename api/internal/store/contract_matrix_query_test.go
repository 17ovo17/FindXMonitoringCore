package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

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
