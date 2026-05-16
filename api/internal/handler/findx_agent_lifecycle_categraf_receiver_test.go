package handler

import (
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCategrafN9EHeartbeatGzipPersistsAgentTargetAndEvidence(t *testing.T) {
	resetCategrafReceiverTestState(t)
	body := gzipTestBody(t, `{
		"agent":"agent-a",
		"host":"host-a",
		"hostname":"host-a",
		"ip":"10.0.0.8",
		"version":"v0.3.90",
		"plugin":"input.cpu",
		"source":"n9e",
		"scope":"linux",
		"labels":{"env":"prod","token":"secret-value"},
		"metadata":{"credential_ref":"<CREDENTIAL_REF>","region":"cn"}
	}`)
	w := performCategrafReceiverPost("/v1/n9e/heartbeat", body, "application/json", "gzip", CategrafN9EHeartbeat)
	if w.Code != http.StatusOK {
		t.Fatalf("gzip heartbeat should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "secret-value") || strings.Contains(w.Body.String(), "<CREDENTIAL_REF>") {
		t.Fatalf("heartbeat response must not echo sensitive data: %s", w.Body.String())
	}
	agents := store.ListFindXAgents()
	if len(agents) != 1 || agents[0].Ident != "agent-a" || agents[0].IP != "10.0.0.8" {
		t.Fatalf("expected heartbeat to upsert FindX agent, got %#v", agents)
	}
	items := findReceiverEvidence(t, model.FindXAgentDataArrivalKindHeartbeat)
	if len(items) != 1 || items[0].Status != model.FindXAgentDataArrivalStatusReported {
		t.Fatalf("expected heartbeat reported evidence, got %#v", items)
	}
	if items[0].Metadata["token"] != "" || items[0].Metadata["credential_ref"] != "" || items[0].Metadata["region"] != "cn" {
		t.Fatalf("heartbeat evidence metadata should be sanitized, got %#v", items[0].Metadata)
	}
}

func TestCategrafN9EHeartbeatPlainJSONPersistsEvidence(t *testing.T) {
	resetCategrafReceiverTestState(t)
	body := strings.NewReader(`{"ident":"agent-plain","hostname":"host-plain","ip":"10.0.0.9","version":"v1"}`)
	w := performCategrafReceiverPost("/v1/n9e/heartbeat", body, "application/json", "", CategrafN9EHeartbeat)
	if w.Code != http.StatusOK {
		t.Fatalf("plain heartbeat should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	items := findReceiverEvidence(t, model.FindXAgentDataArrivalKindHeartbeat)
	if len(items) != 1 || items[0].Metadata["agent"] != "agent-plain" {
		t.Fatalf("expected plain heartbeat evidence, got %#v", items)
	}
}

func TestCategrafPrometheusRemoteWritePersistsIngestionEvidence(t *testing.T) {
	resetCategrafReceiverTestState(t)
	body := strings.NewReader("prometheus-snappy-protobuf-bytes")
	w := performCategrafReceiverPost("/prometheus/v1/write", body, "application/x-protobuf", "snappy", CategrafPrometheusRemoteWrite)
	if w.Code != http.StatusOK {
		t.Fatalf("remote write should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	items := findReceiverEvidence(t, model.FindXAgentDataArrivalKindMetrics)
	if len(items) != 1 || items[0].Status != model.FindXAgentDataArrivalStatusReported {
		t.Fatalf("expected metrics ingestion evidence, got %#v", items)
	}
	if items[0].Metadata["body_bytes"] == "" || items[0].Metadata["content_encoding"] != "snappy" {
		t.Fatalf("expected safe remote write metadata, got %#v", items[0].Metadata)
	}
}

func TestCategrafPrometheusRemoteWriteLinksCmdbDispatchDataArrivalEvidence(t *testing.T) {
	resetCategrafReceiverTestState(t)
	rollout := saveConfigRolloutRuntimeReadRollout(t, map[string]string{
		"delivery_request_ref": "delivery-task",
		"effect_request_ref":   "effect-task",
		"rollback_request_ref": "rollback-task",
	})
	saveConfigRolloutRuntimeReadTask(t, "delivery-task", rollout, "delivery", nil)
	saveConfigRolloutRuntimeReadTask(t, "effect-task", rollout, "effect", nil)
	saveConfigRolloutRuntimeReadTask(t, "rollback-task", rollout, "rollback", nil)

	path := "/prometheus/v1/write?agent_id=agent-a&target_id=host-a&rollout_id=" + rollout.ID + "&request_ref=delivery-task&plugin_id=" + rollout.PluginID
	w := performCategrafReceiverPost(path, strings.NewReader("prometheus-snappy-protobuf-bytes"), "application/x-protobuf", "snappy", CategrafPrometheusRemoteWrite)
	if w.Code != http.StatusOK {
		t.Fatalf("remote write should accept dispatch evidence refs, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(strings.ToLower(w.Body.String()), "token") || strings.Contains(strings.ToLower(w.Body.String()), "cookie") {
		t.Fatalf("receiver response must stay small and sanitized: %s", w.Body.String())
	}

	read := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref=delivery-task", ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadOK(t, read)
	if read.Code != http.StatusOK || payload.Status != "blocked" || payload.Contract != cmdbAgentRolloutDataArrivalReadContract {
		t.Fatalf("runtime read should link receiver evidence while staying blocked, code=%d payload=%#v body=%s", read.Code, payload, read.Body.String())
	}
	if payload.EvidenceCount != 1 || payload.RolloutRef != rollout.ID {
		t.Fatalf("runtime read should expose exactly one receiver evidence row, payload=%#v", payload)
	}
	for _, want := range []string{
		"cmdb_agent_rollout_remote_executor_contract",
		"cmdb_agent_rollout_delivery_receipt_contract",
		"cmdb_agent_rollout_effect_receipt_contract",
		"cmdb_agent_rollout_rollback_receipt_contract",
		"cmdb_agent_rollout_data_arrival_receipt_contract",
		"cmdb_agent_rollout_evidence_chain_contract",
	} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("receiver-linked read should keep real execution gap %q, payload=%#v", want, payload)
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, read.Body.String())
	assertNoDataArrivalRuntimeReadForbidden(t, read.Body.String())
}

func TestCategrafPrometheusRemoteWriteRejectsEmptyBody(t *testing.T) {
	resetCategrafReceiverTestState(t)
	w := performCategrafReceiverPost("/prometheus/v1/write", strings.NewReader(""), "application/x-protobuf", "", CategrafPrometheusRemoteWrite)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty remote write body should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafPrometheusRemoteWriteRejectsTooLargeBody(t *testing.T) {
	resetCategrafReceiverTestState(t)
	w := performCategrafReceiverPost("/prometheus/v1/write", strings.NewReader(strings.Repeat("x", categrafReceiverBodyLimit+1)), "application/x-protobuf", "", CategrafPrometheusRemoteWrite)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized remote write body should be 413, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafN9EHeartbeatRejectsEmptyBody(t *testing.T) {
	resetCategrafReceiverTestState(t)
	w := performCategrafReceiverPost("/v1/n9e/heartbeat", strings.NewReader(""), "application/json", "", CategrafN9EHeartbeat)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty heartbeat body should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafN9EHeartbeatRejectsInvalidJSON(t *testing.T) {
	resetCategrafReceiverTestState(t)
	w := performCategrafReceiverPost("/v1/n9e/heartbeat", strings.NewReader("{"), "application/json", "", CategrafN9EHeartbeat)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid heartbeat json should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafN9EHeartbeatRejectsMissingIdentityFields(t *testing.T) {
	resetCategrafReceiverTestState(t)
	w := performCategrafReceiverPost("/v1/n9e/heartbeat", strings.NewReader(`{"version":"v1"}`), "application/json", "", CategrafN9EHeartbeat)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("heartbeat without ident host hostname or ip should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafN9EHeartbeatRejectsTooLargeBody(t *testing.T) {
	resetCategrafReceiverTestState(t)
	w := performCategrafReceiverPost("/v1/n9e/heartbeat", strings.NewReader(strings.Repeat("x", categrafReceiverBodyLimit+1)), "application/json", "", CategrafN9EHeartbeat)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized heartbeat body should be 413, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafN9EHeartbeatRejectsInvalidGzipBody(t *testing.T) {
	resetCategrafReceiverTestState(t)
	w := performCategrafReceiverPost("/v1/n9e/heartbeat", strings.NewReader("not-gzip"), "application/json", "gzip", CategrafN9EHeartbeat)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid gzip heartbeat body should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafRootRoutesAreRegistered(t *testing.T) {
	router := gin.New()
	router.POST("/v1/n9e/heartbeat", CategrafN9EHeartbeat)
	router.POST("/prometheus/v1/write", CategrafPrometheusRemoteWrite)
	for _, route := range router.Routes() {
		if route.Method == http.MethodPost && route.Path == "/v1/n9e/heartbeat" {
			continue
		}
		if route.Method == http.MethodPost && route.Path == "/prometheus/v1/write" {
			continue
		}
	}
	if len(router.Routes()) != 2 {
		t.Fatalf("expected two root Categraf receiver routes, got %#v", router.Routes())
	}
}
