package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestMonitoringContractMatrixAlertEventActionPipelineQuerySplit(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-ALERT-EVENT-ACTION-PIPELINE-QUERY")
	if err != nil {
		t.Fatalf("get alert event action pipeline query aggregate: %v", err)
	}
	if !ok {
		t.Fatal("alert event action pipeline query aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked || aggregate.SafeToRetry {
		t.Fatalf("aggregate must be blocked and non-retryable: %#v", aggregate)
	}
	if aggregate.Metadata["findx_route"] != "/alerts?section=events" {
		t.Fatalf("aggregate findx_route = %q, want /alerts?section=events", aggregate.Metadata["findx_route"])
	}
	if aggregate.Metadata["gap_type"] != "alert_event_action_pipeline_query_aggregate" {
		t.Fatalf("aggregate gap_type = %q", aggregate.Metadata["gap_type"])
	}
	for _, want := range []string{"ack", "unack", "share credential", "shared detail", "event pipeline crud", "tryrun", "tag lookup", "enrich preview", "executions", "execution detail", "event query/test selector"} {
		if !strings.Contains(aggregate.Metadata["upstream_scope"], want) {
			t.Fatalf("aggregate upstream_scope missing %q: %#v", want, aggregate.Metadata)
		}
	}
	for _, forbidden := range []string{"shield", "delete owned by 119b17", "query-timeseries", "query-log"} {
		if strings.Contains(strings.ToLower(aggregate.Metadata["upstream_scope"]), forbidden) {
			t.Fatalf("aggregate upstream_scope should not own %q: %#v", forbidden, aggregate.Metadata)
		}
	}
	if aggregate.Handler != "" || aggregate.Backend != "" || aggregate.Datasource != "" || aggregate.Executor != "" || len(aggregate.EvidenceRefs) != 0 {
		t.Fatalf("aggregate must not expose executable contract fields: %#v", aggregate)
	}

	counts := map[string]int{}
	for _, tt := range alertEventActionPipelineQueryGapExpectations() {
		t.Run(tt.id, func(t *testing.T) {
			item, ok, err := GetContractMatrixEntry(tt.id)
			if err != nil {
				t.Fatalf("get %s: %v", tt.id, err)
			}
			if !ok {
				t.Fatalf("%s seed missing", tt.id)
			}
			counts[item.Status]++
			if item.Status != tt.status {
				t.Fatalf("%s status = %s, want %s", tt.id, item.Status, tt.status)
			}
			if item.SafeToRetry || item.Handler != "" || item.Backend != "" || item.Datasource != "" || item.Executor != "" || len(item.EvidenceRefs) != 0 {
				t.Fatalf("%s should remain non-ready without executable evidence: %#v", tt.id, item)
			}
			if item.Metadata["findx_route"] != "/alerts?section=events" {
				t.Fatalf("%s findx_route = %q", tt.id, item.Metadata["findx_route"])
			}
			if item.Metadata["gap_type"] != tt.gapType {
				t.Fatalf("%s gap_type = %q, want %q", tt.id, item.Metadata["gap_type"], tt.gapType)
			}
			assertMonitoringN9eGapMetadata(t, item, tt.upstreamRef)
			for _, want := range alertEventActionPipelineQuerySourceRefs() {
				if !contractListContains(item.SourceRefs, want) {
					t.Fatalf("%s source_refs missing %s: %#v", tt.id, want, item.SourceRefs)
				}
			}
			for _, ref := range item.SourceRefs {
				if !strings.HasPrefix(ref, `D:\项目迁移文件\平台源码\`) {
					t.Fatalf("%s source_ref must use authoritative source root: %q", tt.id, ref)
				}
			}
			assertAlertEventActionGapNoSensitiveOrFakeRuntimeText(t, item)
		})
	}
	if counts[model.ContractStatusBlocked] != 2 ||
		counts[model.ContractStatusMissingBackend] != 10 ||
		counts[model.ContractStatusMissingDatasource] != 3 ||
		counts[model.ContractStatusMissingExecutor] != 2 {
		t.Fatalf("status distribution mismatch: %#v", counts)
	}
	for _, oldID := range []string{
		"FX-CONTRACT-N9E-ALERT-EVENT-ACTION-SHIELD-LINK",
		"FX-CONTRACT-N9E-ALERT-EVENT-ACTION-DELETE",
		"FX-CONTRACT-N9E-ALERT-EVENT-UNACK",
		"FX-CONTRACT-N9E-EVENT-PIPELINE-EDIT",
		"FX-CONTRACT-N9E-EVENT-PIPELINE-TEST",
		"FX-CONTRACT-N9E-EVENT-PIPELINE-ENABLE-DISABLE",
		"FX-CONTRACT-N9E-EVENT-PIPELINE-PROCESSOR-SORT",
		"FX-CONTRACT-N9E-EVENT-PIPELINE-EXECUTIONS",
		"FX-CONTRACT-N9E-EVENT-QUERY-TIMESERIES",
		"FX-CONTRACT-N9E-EVENT-QUERY-LOG",
		"FX-CONTRACT-N9E-EVENT-RULE-TESTER-REFERENCE",
	} {
		if _, ok, err := GetContractMatrixEntry(oldID); err != nil {
			t.Fatalf("get old gap %s: %v", oldID, err)
		} else if ok {
			t.Fatalf("old 119B17 gap should not be seeded: %s", oldID)
		}
	}
}

func alertEventActionPipelineQueryGapExpectations() []struct {
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
		{"FX-CONTRACT-N9E-ALERT-EVENT-ACK", model.ContractStatusMissingBackend, "alert_event_ack", "POST /api/n9e-plus/alert-cur-events/{action}"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-SHARE-CREDENTIAL", model.ContractStatusMissingBackend, "alert_event_share_credential", "share-credential-issue"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-SHARED-DETAIL", model.ContractStatusMissingBackend, "alert_event_shared_detail", "GET /api/n9e/alert-his-event/{eventId}; GET /api/n9e-plus/alert-his-event/{eventId}"},
		{"FX-CONTRACT-N9E-EVENT-PIPELINE-CRUD", model.ContractStatusBlocked, "event_pipeline_crud_aggregate", "event-pipeline-crud-aggregate"},
		{"FX-CONTRACT-N9E-EVENT-PIPELINE-LIST", model.ContractStatusMissingBackend, "event_pipeline_list", "GET /api/n9e/event-pipelines"},
		{"FX-CONTRACT-N9E-EVENT-PIPELINE-DETAIL", model.ContractStatusMissingBackend, "event_pipeline_detail", "GET /api/n9e/event-pipeline/{id}"},
		{"FX-CONTRACT-N9E-EVENT-PIPELINE-CREATE", model.ContractStatusMissingBackend, "event_pipeline_create", "POST /api/n9e/event-pipeline"},
		{"FX-CONTRACT-N9E-EVENT-PIPELINE-UPDATE", model.ContractStatusMissingBackend, "event_pipeline_update", "PUT /api/n9e/event-pipeline"},
		{"FX-CONTRACT-N9E-EVENT-PIPELINE-DELETE", model.ContractStatusMissingBackend, "event_pipeline_delete", "DELETE /api/n9e/event-pipelines"},
		{"FX-CONTRACT-N9E-EVENT-PROCESSOR-TRYRUN", model.ContractStatusMissingExecutor, "event_processor_tryrun", "POST /api/n9e/event-processor-tryrun"},
		{"FX-CONTRACT-N9E-EVENT-PIPELINE-TRYRUN", model.ContractStatusMissingExecutor, "event_pipeline_tryrun", "POST /api/n9e/event-pipeline-tryrun"},
		{"FX-CONTRACT-N9E-EVENT-TAGKEYS", model.ContractStatusMissingDatasource, "event_tagkeys", "GET /api/n9e/event-tagkeys"},
		{"FX-CONTRACT-N9E-EVENT-TAGVALUES", model.ContractStatusMissingDatasource, "event_tagvalues", "GET /api/n9e/event-tagvalues?key={key}"},
		{"FX-CONTRACT-N9E-EVENT-ENRICH-DATA-PREVIEW", model.ContractStatusMissingDatasource, "event_enrich_data_preview", "POST /api/n9e-plus/event-enrich-data-preview"},
		{"FX-CONTRACT-N9E-EVENT-PIPELINE-EXECUTIONS-LIST", model.ContractStatusMissingBackend, "event_pipeline_executions_list", "GET /api/n9e/event-pipeline-executions"},
		{"FX-CONTRACT-N9E-EVENT-PIPELINE-EXECUTION-DETAIL", model.ContractStatusMissingBackend, "event_pipeline_execution_detail", "GET /api/n9e/event-pipeline-execution/{id}"},
		{"FX-CONTRACT-N9E-ALERT-EVENT-RULE-TESTER", model.ContractStatusBlocked, "alert_event_rule_tester", "AlertEventRuleTesterWithButton onClick/onTest event selector"},
	}
}

func assertAlertEventActionGapNoSensitiveOrFakeRuntimeText(t *testing.T, item model.ContractMatrixEntry) {
	t.Helper()
	values := []string{item.Capability, item.BlockedReason, item.Handler, item.Backend, item.Datasource, item.Executor}
	values = append(values, item.SourceRefs...)
	values = append(values, item.EvidenceRefs...)
	for key, value := range item.Metadata {
		values = append(values, key, value)
	}
	body := strings.ToLower(strings.Join(values, " "))
	for _, forbidden := range []string{"token", "__token", "cookie", "dsn", "secret", "private key", "api_key", "password", "basic_auth_pass", "header.authorization", "integration-key", "queued", "running", "succeeded", "success", "applied", "rolled-back", "installed", "data_arrived"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("%s exposed forbidden marker %q: %#v", item.ID, forbidden, item)
		}
	}
}
