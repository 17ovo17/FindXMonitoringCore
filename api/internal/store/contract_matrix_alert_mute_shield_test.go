package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestMonitoringContractMatrixAlertMuteShieldSplit(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-ALERT-MUTE-SHIELD")
	if err != nil {
		t.Fatalf("get alert mute shield aggregate: %v", err)
	}
	if !ok {
		t.Fatal("alert mute shield aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked || aggregate.SafeToRetry {
		t.Fatalf("alert mute shield aggregate must stay blocked and non-ready: %#v", aggregate)
	}
	if aggregate.Metadata["findx_route"] != "/alerts?section=mutes" {
		t.Fatalf("aggregate findx_route = %q, want /alerts?section=mutes", aggregate.Metadata["findx_route"])
	}
	if aggregate.Metadata["gap_type"] != "alert_mute_shield_aggregate" {
		t.Fatalf("aggregate gap_type = %q, want alert_mute_shield_aggregate", aggregate.Metadata["gap_type"])
	}
	if aggregate.Metadata["upstream_ref"] != "alert-mute-shield-aggregate" {
		t.Fatalf("aggregate upstream_ref = %q, want alert-mute-shield-aggregate", aggregate.Metadata["upstream_ref"])
	}

	scope := aggregate.Metadata["upstream_scope"]
	for _, want := range []string{
		"list",
		"detail",
		"create",
		"update",
		"delete",
		"bulk-fields",
		"preview",
		"tryrun",
	} {
		if !strings.Contains(scope, want) {
			t.Fatalf("aggregate upstream_scope missing %q: %#v", want, aggregate.Metadata)
		}
	}
	for _, forbidden := range []string{
		"/api/n9e/",
		"alert_mute_list",
		"alert_mute_detail",
		"alert_mute_create",
		"alert_mute_update",
		"alert_mute_delete",
	} {
		if strings.Contains(scope, forbidden) || strings.Contains(aggregate.Metadata["upstream_ref"], forbidden) {
			t.Fatalf("aggregate must not own child or unrelated refs %s: %#v", forbidden, aggregate.Metadata)
		}
	}
	for _, wantExclusion := range []string{
		"alert rule groups",
		"lifecycle",
		"subscribe",
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

	for _, tt := range alertMuteShieldGapExpectations() {
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
			if item.Metadata["findx_route"] != "/alerts?section=mutes" {
				t.Fatalf("%s findx_route = %q, want /alerts?section=mutes", tt.id, item.Metadata["findx_route"])
			}
			if item.Metadata["gap_type"] != tt.gapType {
				t.Fatalf("%s gap_type = %q, want %q", tt.id, item.Metadata["gap_type"], tt.gapType)
			}
			assertMonitoringN9eGapMetadata(t, item, tt.upstreamRef)
			for _, want := range alertMuteShieldSourceRefs() {
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
				"alert_subscribe",
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

func alertMuteShieldGapExpectations() []struct {
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
		{"FX-CONTRACT-N9E-ALERT-MUTE-LIST-BY-BUSI-GROUP", model.ContractStatusMissingBackend, "alert_mute_list_by_busi_group", "GET /api/n9e/busi-group/{id}/alert-mutes"},
		{"FX-CONTRACT-N9E-ALERT-MUTE-LIST-BY-BUSI-GROUPS", model.ContractStatusMissingBackend, "alert_mute_list_by_busi_groups", "GET /api/n9e/busi-groups/alert-mutes"},
		{"FX-CONTRACT-N9E-ALERT-MUTE-DETAIL", model.ContractStatusMissingBackend, "alert_mute_detail", "GET /api/n9e/busi-group/{busiId}/alert-mute/{id}"},
		{"FX-CONTRACT-N9E-ALERT-MUTE-CREATE", model.ContractStatusMissingBackend, "alert_mute_create", "POST /api/n9e/busi-group/{busiId}/alert-mutes"},
		{"FX-CONTRACT-N9E-ALERT-MUTE-UPDATE", model.ContractStatusMissingBackend, "alert_mute_update", "PUT /api/n9e/busi-group/{busiId}/alert-mute/{muteId}"},
		{"FX-CONTRACT-N9E-ALERT-MUTE-DELETE", model.ContractStatusMissingBackend, "alert_mute_delete", "DELETE /api/n9e/busi-group/{busiId}/alert-mutes"},
		{"FX-CONTRACT-N9E-ALERT-MUTE-BULK-FIELDS-UPDATE", model.ContractStatusMissingBackend, "alert_mute_bulk_fields_update", "PUT /api/n9e/busi-group/{busiId}/alert-mutes/fields"},
		{"FX-CONTRACT-N9E-ALERT-MUTE-PREVIEW-EVENTS", model.ContractStatusMissingDatasource, "alert_mute_preview_events", "POST /api/n9e/busi-group/{busiId}/alert-mutes/preview"},
		{"FX-CONTRACT-N9E-ALERT-MUTE-TRYRUN", model.ContractStatusMissingDatasource, "alert_mute_tryrun", "POST /api/n9e/alert-mute-tryrun"},
	}
}
