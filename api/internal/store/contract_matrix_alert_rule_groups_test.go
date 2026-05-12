package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestMonitoringContractMatrixAlertRuleGroupsSplit(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-ALERT-RULE-GROUPS")
	if err != nil {
		t.Fatalf("get alert rule groups aggregate: %v", err)
	}
	if !ok {
		t.Fatal("alert rule groups aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked || aggregate.SafeToRetry {
		t.Fatalf("alert rule groups aggregate must stay blocked and non-ready: %#v", aggregate)
	}
	if aggregate.Metadata["findx_route"] != "/alerts?section=rules" {
		t.Fatalf("aggregate findx_route = %q, want /alerts?section=rules", aggregate.Metadata["findx_route"])
	}
	if aggregate.Metadata["gap_type"] != "alert_rule_group_aggregate" {
		t.Fatalf("aggregate gap_type = %q, want alert_rule_group_aggregate", aggregate.Metadata["gap_type"])
	}
	if aggregate.Metadata["upstream_ref"] != "alert-rule-groups-aggregate" {
		t.Fatalf("aggregate upstream_ref = %q, want alert-rule-groups-aggregate", aggregate.Metadata["upstream_ref"])
	}
	scope := aggregate.Metadata["upstream_scope"]
	for _, want := range []string{
		"list",
		"detail",
		"create",
		"update",
		"delete",
		"favorite",
		"group rules",
	} {
		if !strings.Contains(scope, want) {
			t.Fatalf("aggregate upstream_scope missing %q: %#v", want, aggregate.Metadata)
		}
	}
	for _, forbidden := range []string{
		"/api/n9e/alert-rule-groups",
		"/api/n9e/alert-rule-group/{id}",
		"/api/n9e/busi-group/{id}/alert-rules",
		"alert_rule_group_list",
		"alert_rule_group_detail",
		"alert_rule_group_rules_list",
	} {
		if strings.Contains(scope, forbidden) || strings.Contains(aggregate.Metadata["upstream_ref"], forbidden) {
			t.Fatalf("aggregate must not own child refs %s: %#v", forbidden, aggregate.Metadata)
		}
	}

	for _, tt := range []struct {
		id          string
		gapType     string
		upstreamRef string
		sourceRefs  []string
	}{
		{
			id:          "FX-CONTRACT-N9E-ALERT-RULE-GROUP-LIST",
			gapType:     "alert_rule_group_list",
			upstreamRef: "GET /api/n9e/alert-rule-groups",
			sourceRefs:  monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`),
		},
		{
			id:          "FX-CONTRACT-N9E-ALERT-RULE-GROUP-CREATE",
			gapType:     "alert_rule_group_create",
			upstreamRef: "POST /api/n9e/alert-rule-groups",
			sourceRefs:  monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`),
		},
		{
			id:          "FX-CONTRACT-N9E-ALERT-RULE-GROUP-DETAIL",
			gapType:     "alert_rule_group_detail",
			upstreamRef: "GET /api/n9e/alert-rule-group/{id}",
			sourceRefs:  monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`),
		},
		{
			id:          "FX-CONTRACT-N9E-ALERT-RULE-GROUP-UPDATE",
			gapType:     "alert_rule_group_update",
			upstreamRef: "PUT /api/n9e/alert-rule-group/{id}",
			sourceRefs:  monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`),
		},
		{
			id:          "FX-CONTRACT-N9E-ALERT-RULE-GROUP-DELETE",
			gapType:     "alert_rule_group_delete",
			upstreamRef: "DELETE /api/n9e/alert-rule-group/{id}",
			sourceRefs:  monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`),
		},
		{
			id:          "FX-CONTRACT-N9E-ALERT-RULE-GROUP-RULES-LIST",
			gapType:     "alert_rule_group_rules_list",
			upstreamRef: "GET /api/n9e/busi-group/{id}/alert-rules",
			sourceRefs: monitoringMatureSourceRefs(
				`fe-main\src\services\warning.ts`,
				`fe-main\src\pages\warning\subscribe\components\ruleModal.tsx`,
			),
		},
		{
			id:          "FX-CONTRACT-N9E-ALERT-RULE-GROUPS-MULTI-RULES-LIST",
			gapType:     "alert_rule_group_multi_rules_list",
			upstreamRef: "GET /api/n9e/busi-groups/alert-rules",
			sourceRefs:  monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`),
		},
		{
			id:          "FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITES-LIST",
			gapType:     "alert_rule_group_favorites_list",
			upstreamRef: "GET /api/n9e/alert-rule-groups/favorites",
			sourceRefs:  monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`),
		},
		{
			id:          "FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITE-ADD",
			gapType:     "alert_rule_group_favorite_add",
			upstreamRef: "POST /api/n9e/alert-rule-group/{id}/favorites",
			sourceRefs:  monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`),
		},
		{
			id:          "FX-CONTRACT-N9E-ALERT-RULE-GROUP-FAVORITE-DELETE",
			gapType:     "alert_rule_group_favorite_delete",
			upstreamRef: "DELETE /api/n9e/alert-rule-group/{id}/favorites",
			sourceRefs:  monitoringMatureSourceRefs(`fe-main\src\services\warning.ts`),
		},
	} {
		t.Run(tt.id, func(t *testing.T) {
			item, ok, err := GetContractMatrixEntry(tt.id)
			if err != nil {
				t.Fatalf("get %s: %v", tt.id, err)
			}
			if !ok {
				t.Fatalf("%s seed missing", tt.id)
			}
			if item.Status != model.ContractStatusMissingBackend {
				t.Fatalf("%s status = %s, want missing_backend", tt.id, item.Status)
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
			for _, forbidden := range []string{
				"dashboard",
				"template",
				"metric",
				"alert-rules/import",
				"alert-rules/clone",
				"alert-rules/status",
				"alert-rule/{id}/pure",
				"alert-mutes",
				"alert-subscribes",
				"notify-rules",
				"message-templates",
				"notify-channel-configs",
				"notify-contact",
			} {
				if strings.Contains(item.Metadata["upstream_ref"], forbidden) {
					t.Fatalf("%s must not duplicate unrelated endpoint ownership %q: %#v", tt.id, forbidden, item.Metadata)
				}
			}
		})
	}
}
