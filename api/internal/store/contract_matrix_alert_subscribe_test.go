package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestMonitoringContractMatrixAlertSubscribeSplit(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-ALERT-SUBSCRIBE")
	if err != nil {
		t.Fatalf("get alert subscribe aggregate: %v", err)
	}
	if !ok {
		t.Fatal("alert subscribe aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked || aggregate.SafeToRetry {
		t.Fatalf("alert subscribe aggregate must stay blocked and non-ready: %#v", aggregate)
	}
	if aggregate.Metadata["findx_route"] != "/alerts?section=subscribes" {
		t.Fatalf("aggregate findx_route = %q, want /alerts?section=subscribes", aggregate.Metadata["findx_route"])
	}
	if aggregate.Metadata["gap_type"] != "alert_subscribe_aggregate" {
		t.Fatalf("aggregate gap_type = %q, want alert_subscribe_aggregate", aggregate.Metadata["gap_type"])
	}
	if aggregate.Metadata["upstream_ref"] != "alert-subscribe-aggregate" {
		t.Fatalf("aggregate upstream_ref = %q, want alert-subscribe-aggregate", aggregate.Metadata["upstream_ref"])
	}

	scope := aggregate.Metadata["upstream_scope"]
	for _, want := range []string{
		"list",
		"detail",
		"create",
		"update",
		"delete",
		"tryrun",
	} {
		if !strings.Contains(scope, want) {
			t.Fatalf("aggregate upstream_scope missing %q: %#v", want, aggregate.Metadata)
		}
	}
	for _, forbidden := range []string{
		"/api/n9e/",
		"alert_subscribe_list",
		"alert_subscribe_detail",
		"alert_subscribe_create",
		"alert_subscribe_update",
		"alert_subscribe_delete",
	} {
		if strings.Contains(scope, forbidden) || strings.Contains(aggregate.Metadata["upstream_ref"], forbidden) {
			t.Fatalf("aggregate must not own child or concrete refs %s: %#v", forbidden, aggregate.Metadata)
		}
	}
	for _, wantExclusion := range []string{
		"alert rule groups",
		"lifecycle",
		"mute",
		"shield",
		"notification",
		"dashboard",
		"template",
		"metric",
		"event",
	} {
		if !strings.Contains(scope, wantExclusion) {
			t.Fatalf("aggregate upstream_scope missing exclusion %q: %#v", wantExclusion, aggregate.Metadata)
		}
	}

	for _, tt := range alertSubscribeGapExpectations() {
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
			if item.Metadata["findx_route"] != "/alerts?section=subscribes" {
				t.Fatalf("%s findx_route = %q, want /alerts?section=subscribes", tt.id, item.Metadata["findx_route"])
			}
			if item.Metadata["gap_type"] != tt.gapType {
				t.Fatalf("%s gap_type = %q, want %q", tt.id, item.Metadata["gap_type"], tt.gapType)
			}
			assertMonitoringN9eGapMetadata(t, item, tt.upstreamRef)
			for _, want := range alertSubscribeSourceRefs() {
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
				"alert_rule_group",
				"alert_rule_lifecycle",
				"alert_mute",
				"notification",
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

func alertSubscribeGapExpectations() []struct {
	id          string
	status      string
	gapType     string
	upstreamRef string
} {
	return []struct {
		id          string
		status      string
		gapType     string
		upstreamRef string
	}{
		{"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-LIST-BY-BUSI-GROUP", model.ContractStatusMissingBackend, "alert_subscribe_list_by_busi_group", "GET /api/n9e/busi-group/{id}/alert-subscribes"},
		{"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-LIST-BY-BUSI-GROUPS", model.ContractStatusMissingBackend, "alert_subscribe_list_by_busi_groups", "GET /api/n9e/busi-groups/alert-subscribes"},
		{"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-DETAIL", model.ContractStatusMissingBackend, "alert_subscribe_detail", "GET /api/n9e/alert-subscribe/{subscribeId}"},
		{"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-CREATE", model.ContractStatusMissingBackend, "alert_subscribe_create", "POST /api/n9e/busi-group/{busiId}/alert-subscribes"},
		{"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-UPDATE", model.ContractStatusMissingBackend, "alert_subscribe_update", "PUT /api/n9e/busi-group/{busiId}/alert-subscribes"},
		{"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-DELETE", model.ContractStatusMissingBackend, "alert_subscribe_delete", "DELETE /api/n9e/busi-group/{busiId}/alert-subscribes"},
		{"FX-CONTRACT-N9E-ALERT-SUBSCRIBE-TRYRUN", model.ContractStatusMissingDatasource, "alert_subscribe_tryrun", "POST /api/n9e/alert-subscribe/alert-subscribes-tryrun"},
	}
}
