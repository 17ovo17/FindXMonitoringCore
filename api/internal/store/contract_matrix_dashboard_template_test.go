package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestMonitoringContractMatrixDashboardTemplateBuiltinImportSplit(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT")
	if err != nil {
		t.Fatalf("get dashboard template import aggregate: %v", err)
	}
	if !ok {
		t.Fatal("dashboard template import aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked || aggregate.SafeToRetry {
		t.Fatalf("dashboard template import aggregate must remain non-ready: %#v", aggregate)
	}
	if aggregate.Metadata["upstream_ref"] != "dashboard-template-import-aggregate" {
		t.Fatalf("aggregate must not own concrete builtin/template endpoints: %#v", aggregate.Metadata)
	}
	if strings.Contains(aggregate.Metadata["upstream_ref"], "/api/n9e/builtin-boards-detail") ||
		strings.Contains(aggregate.Metadata["upstream_ref"], "/api/n9e/busi-group/{id}/boards") {
		t.Fatalf("aggregate must point to child gaps, not direct endpoint ownership: %#v", aggregate.Metadata)
	}
	scope := aggregate.Metadata["upstream_scope"]
	for _, want := range []string{"builtin detail", "batch result", "conflict rollback", "docs drawer", "excludes CRUD/public/export/migrate/annotations"} {
		if !strings.Contains(scope, want) {
			t.Fatalf("aggregate scope missing %q: %#v", want, aggregate.Metadata)
		}
	}

	for _, tt := range dashboardTemplateBuiltinGapCases() {
		t.Run(tt.id, func(t *testing.T) {
			item := assertMonitoringDashboardTemplateBuiltinGap(t, tt.id, tt.gapType, tt.upstreamRef)
			for _, want := range tt.sourceRefs {
				if !contractListContains(item.SourceRefs, want) {
					t.Fatalf("%s source_refs missing mature source %s: %#v", tt.id, want, item.SourceRefs)
				}
			}
			for _, forbidden := range []string{
				"GET /api/n9e/dashboard-annotations",
				"POST /api/n9e/dashboard-annotations",
				"PUT /api/n9e/dashboard-annotation/{id}",
				"DELETE /api/n9e/dashboard-annotation/{id}",
				"GET /api/n9e/busi-groups/public-boards",
				"PUT /api/n9e/board/{id}/public",
				"POST /api/n9e/busi-group/{busiId}/dashboards/export",
				"PUT /api/n9e/dashboard/{id}/migrate",
				"GET /api/n9e/busi-group/{id}/boards",
				"GET /api/n9e/busi-groups/boards",
				"PUT /api/n9e/board/{id}/configs",
				"GET /api/n9e/boards?bids={ids}",
			} {
				if strings.Contains(item.Metadata["upstream_ref"], forbidden) {
					t.Fatalf("%s must not duplicate 119B7/119B8/119B9 endpoint ownership %s: %#v", tt.id, forbidden, item.Metadata)
				}
			}
		})
	}
}

func TestMonitoringContractMatrixDashboardTemplateBuiltinDoesNotExpandReadySingleQuery(t *testing.T) {
	ResetContractMatrixForTest()

	item, ok, err := GetContractMatrixEntry("FX-CONTRACT-FINDX-PROMETHEUS-SINGLE-QUERY")
	if err != nil {
		t.Fatalf("get single query adapter: %v", err)
	}
	if !ok {
		t.Fatal("single query adapter seed missing")
	}
	if item.Status != model.ContractStatusReady || !item.SafeToRetry {
		t.Fatalf("single query adapter must stay the only ready monitoring query regression target: %#v", item)
	}
	for _, forbidden := range []string{
		"/api/n9e/builtin-boards-detail",
		"template_center_builtin_board_detail",
		"dashboard_template_import_batch_result",
		"dashboard_template_import_conflict_rollback",
		"template_center_document_drawer",
	} {
		body := strings.Join(append(item.SourceRefs, item.EvidenceRefs...), " ") + item.Metadata["upstream_ref"] + item.Metadata["upstream_scope"]
		if strings.Contains(body, forbidden) {
			t.Fatalf("single query ready adapter must not claim dashboard template builtin contract %s: %#v", forbidden, item)
		}
	}
}

func dashboardTemplateBuiltinGapCases() []struct {
	id          string
	gapType     string
	upstreamRef string
	sourceRefs  []string
} {
	return []struct {
		id          string
		gapType     string
		upstreamRef string
		sourceRefs  []string
	}{
		{
			id:          "FX-CONTRACT-N9E-TEMPLATE-CENTER-BUILTIN-BOARD-DETAIL",
			gapType:     "template_center_builtin_board_detail",
			upstreamRef: "POST /api/n9e/builtin-boards-detail",
			sourceRefs: dashboardTemplateMatureSourceRefs(
				`fe-main\src\services\dashboardV2.ts`,
				`fe-main\src\pages\builtInComponents\Dashboards\Detail.tsx`,
				`fe-main\src\pages\dashboard\Detail\Detail.tsx`,
			),
		},
		{
			id:          "FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT-BATCH-RESULT",
			gapType:     "dashboard_template_import_batch_result",
			upstreamRef: "template-import-batch-result",
			sourceRefs:  dashboardTemplateMatureSourceRefs(`fe-main\src\pages\builtInComponents\Dashboards\Import.tsx`, `fe-main\src\pages\builtInComponents\Dashboards\services.ts`),
		},
		{
			id:          "FX-CONTRACT-N9E-TEMPLATE-CENTER-DASHBOARD-IMPORT-CONFLICT-ROLLBACK",
			gapType:     "dashboard_template_import_conflict_rollback",
			upstreamRef: "template-import-conflict-rollback",
			sourceRefs:  dashboardTemplateMatureSourceRefs(`fe-main\src\pages\builtInComponents\Dashboards\Import.tsx`, `fe-main\src\services\dashboardV2.ts`),
		},
		{
			id:          "FX-CONTRACT-N9E-TEMPLATE-CENTER-DOCUMENT-DRAWER",
			gapType:     "template_center_document_drawer",
			upstreamRef: "/n9e-docs/{path}/{language}.md",
			sourceRefs:  dashboardTemplateMatureSourceRefs(`fe-main\src\components\DocumentDrawer\index.tsx`, `fe-main\src\components\DocumentDrawer\Document.tsx`),
		},
	}
}

func dashboardTemplateMatureSourceRefs(relativePaths ...string) []string {
	refs := make([]string, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		refs = append(refs, `D:\项目迁移文件\平台源码\`+relativePath)
	}
	return refs
}

func assertMonitoringDashboardTemplateBuiltinGap(t *testing.T, id, gapType, upstreamRef string) model.ContractMatrixEntry {
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
	if item.Metadata["findx_route"] != "/integrations?section=templates" {
		t.Fatalf("%s findx_route = %q, want /integrations?section=templates", id, item.Metadata["findx_route"])
	}
	if item.Metadata["gap_type"] != gapType {
		t.Fatalf("%s gap_type = %q, want %q", id, item.Metadata["gap_type"], gapType)
	}
	assertMonitoringN9eGapMetadata(t, item, upstreamRef)
	return item
}
