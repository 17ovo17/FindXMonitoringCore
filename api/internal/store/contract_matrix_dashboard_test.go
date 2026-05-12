package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

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

func TestMonitoringContractMatrixDashboardPublicExportMigrateStayBlocked(t *testing.T) {
	ResetContractMatrixForTest()

	templateImport, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT")
	if err != nil {
		t.Fatalf("get template center dashboard import: %v", err)
	}
	if !ok {
		t.Fatal("template center dashboard import seed missing")
	}
	if templateImport.Metadata["upstream_ref"] == "PUT /api/n9e/dashboard/{id}/migrate" ||
		templateImport.Metadata["upstream_ref"] == "/api/n9e/dashboard/{id}/migrate" {
		t.Fatalf("template center import must not directly own migrate endpoint: %#v", templateImport.Metadata)
	}
	if templateImport.Metadata["upstream_ref"] != "dashboard-template-import-aggregate" {
		t.Fatalf("template center import should be aggregate scoped: %#v", templateImport.Metadata)
	}
	if strings.Contains(templateImport.Metadata["upstream_ref"], "/api/n9e/dashboard/{id}/migrate") {
		t.Fatalf("template center import upstream_ref must not duplicate dashboard migrate ownership: %#v", templateImport.Metadata)
	}
	if !strings.Contains(templateImport.Metadata["upstream_scope"], "migrate") ||
		!strings.Contains(templateImport.Metadata["upstream_scope"], "rollback") ||
		!strings.Contains(templateImport.Metadata["upstream_scope"], "conflict") {
		t.Fatalf("template center import scope must describe aggregate flow dependencies: %#v", templateImport.Metadata)
	}

	tests := []struct {
		id          string
		upstreamRef string
	}{
		{id: "FX-CONTRACT-N9E-DASHBOARD-PUBLIC-LIST", upstreamRef: "GET /api/n9e/busi-groups/public-boards"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-PUBLIC-UPDATE", upstreamRef: "PUT /api/n9e/board/{id}/public"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-EXPORT", upstreamRef: "POST /api/n9e/busi-group/{busiId}/dashboards/export"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-MIGRATE", upstreamRef: "PUT /api/n9e/dashboard/{id}/migrate"},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			item := assertMonitoringDashboardActionsSplitGap(t, tt.id, tt.upstreamRef)
			for _, want := range monitoringMatureSourceRefs(`fe-main\src\services\dashboardV2.ts`) {
				if !contractListContains(item.SourceRefs, want) {
					t.Fatalf("%s source_refs missing mature source %s: %#v", tt.id, want, item.SourceRefs)
				}
			}
			for _, forbidden := range []string{"/api/n9e/dashboard-annotations", "/api/n9e/dashboard-annotation", "/api/n9e/metric-views", "/api/{N9E_PATHNAME}/proxy"} {
				if strings.Contains(item.Metadata["upstream_ref"], forbidden) {
					t.Fatalf("%s must not duplicate unrelated endpoint ownership %s: %#v", tt.id, forbidden, item.Metadata)
				}
			}
		})
	}
}

func assertMonitoringDashboardAnnotationsSplitGap(t *testing.T, id, upstreamRef string) model.ContractMatrixEntry {
	t.Helper()
	return assertMonitoringDashboardActionsSplitGap(t, id, upstreamRef)
}

func assertMonitoringDashboardActionsSplitGap(t *testing.T, id, upstreamRef string) model.ContractMatrixEntry {
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
