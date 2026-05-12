package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestMonitoringContractMatrixDashboardCRUDIsSplitIntoPreciseGaps(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-DASHBOARD-CRUD")
	if err != nil {
		t.Fatalf("get dashboard CRUD aggregate: %v", err)
	}
	if !ok {
		t.Fatal("dashboard CRUD aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked {
		t.Fatalf("dashboard CRUD aggregate status = %s, want blocked", aggregate.Status)
	}
	if aggregate.SafeToRetry || aggregate.Handler != "" || aggregate.Backend != "" || aggregate.Datasource != "" || aggregate.Executor != "" || len(aggregate.EvidenceRefs) != 0 {
		t.Fatalf("dashboard CRUD aggregate should remain non-ready without executable evidence: %#v", aggregate)
	}
	if aggregate.Metadata["upstream_ref"] != "dashboard-crud-aggregate" {
		t.Fatalf("aggregate must not own dashboard CRUD endpoint directly: %#v", aggregate.Metadata)
	}
	for _, forbidden := range []string{
		"/api/n9e/busi-group/",
		"/api/n9e/busi-groups/boards",
		"/api/n9e/board/",
		"/api/n9e/boards",
		"/api/n9e/dashboard-annotations",
		"/api/n9e/dashboard-annotation",
		"/api/n9e/dashboard/{id}/migrate",
	} {
		if strings.Contains(aggregate.Metadata["upstream_ref"], forbidden) {
			t.Fatalf("aggregate must point to child gaps, not direct endpoint ownership %s: %#v", forbidden, aggregate.Metadata)
		}
	}
	scope := aggregate.Metadata["upstream_scope"]
	for _, want := range []string{
		"dashboard CRUD child gaps",
		"list",
		"create",
		"detail",
		"update",
		"clone",
		"delete",
		"pure",
		"names",
		"excludes public/export/migrate/annotations",
	} {
		if !strings.Contains(scope, want) {
			t.Fatalf("aggregate scope missing %q: %#v", want, aggregate.Metadata)
		}
	}
	assertMonitoringN9eGapMetadata(t, aggregate, "dashboard-crud-aggregate")

	for _, tt := range dashboardCRUDGapCases() {
		t.Run(tt.id, func(t *testing.T) {
			item := assertMonitoringDashboardCRUDSplitGap(t, tt.id, tt.upstreamRef)
			if item.Metadata["gap_type"] != tt.gapType {
				t.Fatalf("%s gap_type = %q, want %q", tt.id, item.Metadata["gap_type"], tt.gapType)
			}
			for _, want := range monitoringMatureSourceRefs(`fe-main\src\services\dashboardV2.ts`) {
				if !contractListContains(item.SourceRefs, want) {
					t.Fatalf("%s source_refs missing mature source %s: %#v", tt.id, want, item.SourceRefs)
				}
			}
			for _, forbidden := range []string{
				"/api/n9e/dashboard-annotations",
				"/api/n9e/dashboard-annotation",
				"/api/n9e/busi-groups/public-boards",
				"/api/n9e/board/{id}/public",
				"/api/n9e/busi-group/{busiId}/dashboards/export",
				"/api/n9e/dashboard/{id}/migrate",
				"/api/n9e/metric-views",
				"/api/{N9E_PATHNAME}/proxy",
			} {
				if strings.Contains(item.Metadata["upstream_ref"], forbidden) {
					t.Fatalf("%s must not duplicate unrelated endpoint ownership %s: %#v", tt.id, forbidden, item.Metadata)
				}
			}
		})
	}
}

func dashboardCRUDGapCases() []struct {
	id          string
	gapType     string
	upstreamRef string
} {
	return []struct {
		id          string
		gapType     string
		upstreamRef string
	}{
		{id: "FX-CONTRACT-N9E-DASHBOARD-LIST-BY-BUSI-GROUP", gapType: "dashboard_list_by_busi_group", upstreamRef: "GET /api/n9e/busi-group/{id}/boards"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-LIST-BY-BUSI-GROUPS", gapType: "dashboard_list_by_busi_groups", upstreamRef: "GET /api/n9e/busi-groups/boards"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-CREATE", gapType: "dashboard_create", upstreamRef: "POST /api/n9e/busi-group/{id}/boards"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-DETAIL", gapType: "dashboard_detail", upstreamRef: "GET /api/n9e/board/{id}"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-UPDATE-METADATA", gapType: "dashboard_update_metadata", upstreamRef: "PUT /api/n9e/board/{id}"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-UPDATE-CONFIGS", gapType: "dashboard_update_configs", upstreamRef: "PUT /api/n9e/board/{id}/configs"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-CLONE", gapType: "dashboard_clone", upstreamRef: "POST /api/n9e/busi-group/{busiId}/board/{id}/clone"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-CLONES", gapType: "dashboard_clones", upstreamRef: "POST /api/n9e/busi-groups/boards/clones"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-DELETE", gapType: "dashboard_delete", upstreamRef: "DELETE /api/n9e/boards"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-PURE-DETAIL", gapType: "dashboard_pure_detail", upstreamRef: "GET /api/n9e/board/{id}/pure"},
		{id: "FX-CONTRACT-N9E-DASHBOARD-NAMES", gapType: "dashboard_names", upstreamRef: "GET /api/n9e/boards?bids={ids}"},
	}
}

func assertMonitoringDashboardCRUDSplitGap(t *testing.T, id, upstreamRef string) model.ContractMatrixEntry {
	t.Helper()
	item := assertMonitoringDashboardActionsSplitGap(t, id, upstreamRef)
	if !strings.Contains(item.BlockedReason, "backend contract is missing") {
		t.Fatalf("%s blocked reason should stay missing backend oriented: %#v", id, item.BlockedReason)
	}
	return item
}
