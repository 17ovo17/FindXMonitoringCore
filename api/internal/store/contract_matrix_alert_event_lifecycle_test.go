package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestMonitoringContractMatrixAlertEventLifecycleSplit(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-ALERT-EVENT-LIFECYCLE")
	if err != nil {
		t.Fatalf("get alert event lifecycle aggregate: %v", err)
	}
	if !ok {
		t.Fatal("alert event lifecycle aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked || aggregate.SafeToRetry {
		t.Fatalf("alert event lifecycle aggregate must stay blocked and non-ready: %#v", aggregate)
	}
	if aggregate.Metadata["findx_route"] != "/alerts?section=events" {
		t.Fatalf("aggregate findx_route = %q, want /alerts?section=events", aggregate.Metadata["findx_route"])
	}
	if aggregate.Metadata["gap_type"] != "alert_event_lifecycle_aggregate" {
		t.Fatalf("aggregate gap_type = %q, want alert_event_lifecycle_aggregate", aggregate.Metadata["gap_type"])
	}
	if aggregate.Metadata["upstream_ref"] != "alert-event-lifecycle-aggregate" {
		t.Fatalf("aggregate upstream_ref = %q, want alert-event-lifecycle-aggregate", aggregate.Metadata["upstream_ref"])
	}

	scope := aggregate.Metadata["upstream_scope"]
	for _, want := range []string{
		"current",
		"history",
		"list",
		"detail",
		"delete",
		"notify",
		"card",
		"ds",
		"csv",
		"share-read",
	} {
		if !strings.Contains(scope, want) {
			t.Fatalf("aggregate upstream_scope missing %q: %#v", want, aggregate.Metadata)
		}
	}
	for _, wantExclusion := range []string{
		"rule",
		"mute",
		"shield",
		"subscribe",
		"notif",
		"pipeline",
		"query",
		"dashboard",
		"template",
		"metric",
		"ack",
		"share-cred",
	} {
		if !strings.Contains(scope, wantExclusion) {
			t.Fatalf("aggregate upstream_scope missing exclusion %q: %#v", wantExclusion, aggregate.Metadata)
		}
	}
	for _, forbidden := range []string{
		"/api/n9e/",
		"alert_event_current_list",
		"alert_event_detail",
		"alert_event_history_list",
	} {
		if strings.Contains(scope, forbidden) || strings.Contains(aggregate.Metadata["upstream_ref"], forbidden) {
			t.Fatalf("aggregate must not own child refs or sensitive endpoint markers %s: %#v", forbidden, aggregate.Metadata)
		}
	}

	for _, tt := range alertEventLifecycleGapExpectations() {
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
			if item.Metadata["findx_route"] != "/alerts?section=events" {
				t.Fatalf("%s findx_route = %q, want /alerts?section=events", tt.id, item.Metadata["findx_route"])
			}
			if item.Metadata["gap_type"] != tt.gapType {
				t.Fatalf("%s gap_type = %q, want %q", tt.id, item.Metadata["gap_type"], tt.gapType)
			}
			assertMonitoringN9eGapMetadata(t, item, tt.upstreamRef)
			for _, want := range alertEventLifecycleSourceRefs() {
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
				"alert_rule_lifecycle",
				"alert_mute",
				"alert_subscribe",
				"notification_adapter",
				"dashboard",
				"template",
				"metric_view",
				"metric_query",
				"event_pipeline",
				"share_credential",
				"ack_action",
			} {
				if strings.Contains(item.Metadata["gap_type"], forbidden) || strings.Contains(item.Metadata["upstream_ref"], forbidden) {
					t.Fatalf("%s must not duplicate unrelated ownership %q: %#v", tt.id, forbidden, item.Metadata)
				}
			}
		})
	}
}

func alertEventLifecycleGapExpectations() []struct {
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
		{"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-LIST", model.ContractStatusMissingBackend, "alert_event_current_list", "GET /api/n9e/alert-cur-events/list; GET /api/n9e-plus/alert-cur-events/list"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-DATASOURCES", model.ContractStatusMissingBackend, "alert_event_current_datasources", "GET /api/n9e/alert-cur-events-datasources"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-CARD-LIST", model.ContractStatusMissingDatasource, "alert_event_current_card_list", "GET /api/n9e/alert-cur-events/card"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-CARD-DETAILS", model.ContractStatusMissingDatasource, "alert_event_current_card_details", "POST /api/n9e/alert-cur-events/card/details; POST /api/n9e-plus/alert-cur-events/card/details"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-DETAIL", model.ContractStatusMissingBackend, "alert_event_detail", "GET /api/n9e/alert-his-event/{eventId}; GET /api/n9e-plus/alert-his-event/{eventId}"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-CURRENT-DELETE", model.ContractStatusMissingBackend, "alert_event_current_delete", "DELETE /api/n9e/alert-cur-events"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-LIST", model.ContractStatusMissingBackend, "alert_event_history_list", "GET /api/n9e/alert-his-events/list"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-BY-IDS", model.ContractStatusMissingBackend, "alert_event_history_by_ids", "GET /api/n9e-plus/alert-his-events/{ids}"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-HISTORY-CLEANUP", model.ContractStatusMissingBackend, "alert_event_history_cleanup", "DELETE /api/n9e/alert-his-events"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-NOTIFY-RECORDS", model.ContractStatusMissingBackend, "alert_event_notify_records", "GET /api/n9e/event-notify-records/{eventId}"},
	}
}
