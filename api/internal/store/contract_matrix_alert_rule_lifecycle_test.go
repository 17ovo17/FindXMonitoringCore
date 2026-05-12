package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestMonitoringContractMatrixAlertRuleLifecycleSplit(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-ALERT-RULE-LIFECYCLE")
	if err != nil {
		t.Fatalf("get alert rule lifecycle aggregate: %v", err)
	}
	if !ok {
		t.Fatal("alert rule lifecycle aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked || aggregate.SafeToRetry {
		t.Fatalf("alert rule lifecycle aggregate must stay blocked and non-ready: %#v", aggregate)
	}
	if aggregate.Metadata["findx_route"] != "/alerts?section=rules" {
		t.Fatalf("aggregate findx_route = %q, want /alerts?section=rules", aggregate.Metadata["findx_route"])
	}
	if aggregate.Metadata["gap_type"] != "alert_rule_lifecycle_aggregate" {
		t.Fatalf("aggregate gap_type = %q, want alert_rule_lifecycle_aggregate", aggregate.Metadata["gap_type"])
	}
	if aggregate.Metadata["upstream_ref"] != "alert-rule-lifecycle-aggregate" {
		t.Fatalf("aggregate upstream_ref = %q, want alert-rule-lifecycle-aggregate", aggregate.Metadata["upstream_ref"])
	}
	scope := aggregate.Metadata["upstream_scope"]
	for _, want := range []string{
		"detail",
		"pure",
		"create",
		"update",
		"delete",
		"import",
		"clone",
		"bulk-fields",
		"enable",
		"validate",
		"tryrun",
		"timezones",
		"callbacks",
	} {
		if !strings.Contains(scope, want) {
			t.Fatalf("aggregate upstream_scope missing %q: %#v", want, aggregate.Metadata)
		}
	}
	for _, forbidden := range []string{
		"/api/n9e/",
		"alert_rule_detail",
		"alert_rule_create",
		"alert_rule_group",
		"favorite",
	} {
		if strings.Contains(scope, forbidden) || strings.Contains(aggregate.Metadata["upstream_ref"], forbidden) {
			t.Fatalf("aggregate must not own child refs %s: %#v", forbidden, aggregate.Metadata)
		}
	}

	for _, tt := range alertRuleLifecycleGapExpectations() {
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
			if item.SafeToRetry || item.Handler != "" || item.Backend != "" || item.Datasource != "" || item.Executor != "" || len(item.EvidenceRefs) != 0 {
				t.Fatalf("%s should remain non-ready without executable evidence: %#v", tt.id, item)
			}
			if item.Metadata["findx_route"] != "/alerts?section=rules" {
				t.Fatalf("%s findx_route = %q, want /alerts?section=rules", tt.id, item.Metadata["findx_route"])
			}
			if item.Metadata["gap_type"] != tt.gapType {
				t.Fatalf("%s gap_type = %q, want %q", tt.id, item.Metadata["gap_type"], tt.gapType)
			}
			assertMonitoringN9eGapMetadata(t, item, tt.upstreamRef)
			for _, want := range tt.sourceRefs {
				if !contractListContains(item.SourceRefs, want) {
					t.Fatalf("%s source_refs missing %s: %#v", tt.id, want, item.SourceRefs)
				}
			}
			for _, ref := range item.SourceRefs {
				if !strings.HasPrefix(ref, `D:\项目迁移文件\平台源码\`) {
					t.Fatalf("%s source_ref must use authoritative migrated source root: %q", tt.id, ref)
				}
			}
			for _, forbidden := range []string{
				"favorite",
				"alert-mutes",
				"alert-subscribes",
				"notify-rules",
				"message-templates",
				"notify-channel-configs",
				"notify-contact",
				"dashboard",
				"template",
				"metric_view",
				"metric_query",
				"event_lifecycle",
			} {
				if strings.Contains(item.Metadata["gap_type"], forbidden) || strings.Contains(item.Metadata["upstream_ref"], forbidden) {
					t.Fatalf("%s must not duplicate unrelated ownership %q: %#v", tt.id, forbidden, item.Metadata)
				}
			}
		})
	}
}

func alertRuleLifecycleGapExpectations() []struct {
	id          string
	status      string
	gapType     string
	upstreamRef string
	sourceRefs  []string
} {
	return []struct {
		id          string
		status      string
		gapType     string
		upstreamRef string
		sourceRefs  []string
	}{
		{"FX-CONTRACT-N9E-ALERT-RULE-DETAIL", model.ContractStatusMissingBackend, "alert_rule_detail", "GET /api/n9e/alert-rule/{id}", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`, `fe-main\src\pages\alertRules\Edit.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-PURE-DETAIL", model.ContractStatusMissingBackend, "alert_rule_pure_detail", "GET /api/n9e/alert-rule/{id}/pure", monitoringMatureSourceRefs(`fe-main\src\pages\alertRules\services.ts`, `fe-main\src\pages\alertRules\Edit.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-CREATE", model.ContractStatusMissingBackend, "alert_rule_create", "POST /api/n9e/busi-group/{busiId}/alert-rules", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`, `fe-main\src\pages\alertRules\Form\index.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-UPDATE", model.ContractStatusMissingBackend, "alert_rule_update", "PUT /api/n9e/busi-group/{busiId}/alert-rule/{strategyId}", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`, `fe-main\src\pages\alertRules\Form\index.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-DELETE-BY-BUSI-GROUP", model.ContractStatusMissingBackend, "alert_rule_delete_by_busi_group", "DELETE /api/n9e/busi-group/{busiId}/alert-rules", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`, `fe-main\src\pages\alertRules\List\MoreOperations.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-DELETE-BY-RULE-GROUP", model.ContractStatusMissingBackend, "alert_rule_delete_by_rule_group", "DELETE /api/n9e/alert-rule-group/{ruleGroupId}/alert-rules", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-IMPORT-JSON", model.ContractStatusMissingBackend, "alert_rule_import_json", "POST /api/n9e/busi-group/{busiId}/alert-rules/import", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`, `fe-main\src\pages\alertRules\List\Import\ImportBase.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-IMPORT-PROM-RULE", model.ContractStatusMissingBackend, "alert_rule_import_prom_rule", "POST /api/n9e/busi-group/{busiId}/alert-rules/import-prom-rule", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`, `fe-main\src\pages\alertRules\List\Import\ImportPrometheus.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-BULK-FIELDS-UPDATE", model.ContractStatusMissingBackend, "alert_rule_bulk_fields_update", "PUT /api/n9e/busi-group/{busiId}/alert-rules/fields", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`, `fe-main\src\pages\alertRules\List\EditModal.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-ENABLE-DISABLE", model.ContractStatusMissingBackend, "alert_rule_enable_disable", "PUT /api/n9e/busi-group/{busiId}/alert-rules/fields", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`, `fe-main\src\pages\alertRules\List\ListNG.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-STATUS-BATCH", model.ContractStatusMissingBackend, "alert_rule_status_batch", "PUT /api/n9e/alert-rules/status", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-CLONE-TO-HOSTS", model.ContractStatusMissingBackend, "alert_rule_clone_to_hosts", "POST /api/n9e/busi-group/{gid}/alert-rules/clone", monitoringMatureSourceRefs(`fe-main\src\pages\alertRules\services.ts`, `fe-main\src\pages\alertRules\List\CloneToHosts\index.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-CLONE-TO-BUSI-GROUPS", model.ContractStatusMissingBackend, "alert_rule_clone_to_busi_groups", "POST /api/n9e/busi-groups/alert-rules/clones", monitoringMatureSourceRefs(`fe-main\src\pages\alertRules\services.ts`, `fe-main\src\pages\alertRules\List\CloneToBgids\index.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-VALIDATE", model.ContractStatusMissingBackend, "alert_rule_validate", "PUT /api/n9e/busi-group/alert-rule/validate", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`, `fe-main\src\pages\alertRules\Form\index.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-ENABLE-TRYRUN", model.ContractStatusMissingDatasource, "alert_rule_enable_tryrun", "POST /api/n9e/busi-group/alert-rules/enable-tryrun", monitoringMatureSourceRefs(`fe-main\src\pages\alertRules\services.ts`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-CALLBACKS-LIST", model.ContractStatusMissingBackend, "alert_rule_callbacks_list", "GET /api/n9e/alert-rules/callbacks", monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`, `fe-main\src\pages\alertRules\Form\Notify\index.tsx`)},
		{"FX-CONTRACT-N9E-ALERT-RULE-TIMEZONES", model.ContractStatusMissingBackend, "alert_rule_timezones", "GET /api/n9e/timezones", monitoringMatureSourceRefs(`fe-main\src\pages\alertRules\services.ts`, `fe-main\src\pages\alertRules\Form\Effective\index.tsx`)},
	}
}
