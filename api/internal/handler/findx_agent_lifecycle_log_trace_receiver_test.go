package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentLogsCompatibleReceiverPersistsEvidence(t *testing.T) {
	resetCategrafReceiverTestState(t)
	body := gzipTestBody(t, `{
		"agent_id":"agent-log-1",
		"source":"filelog",
		"scope":"linux",
		"service":"orders",
		"trace_id":"trace-log-1",
		"records":[{"message":"user login failed"}],
		"metadata":{"region":"cn","password":"do-not-store"},
		"labels":{"credential_ref":"<CREDENTIAL_REF>","env":"prod"}
	}`)
	w := performCategrafReceiverPost("/findx-agent/v1/logs", body, "application/json", "gzip", FindXAgentLogsCompatibleReceiver)
	if w.Code != http.StatusOK {
		t.Fatalf("logs receiver should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	assertFindXAgentReceiverSmallResponse(t, w.Body.String(), model.FindXAgentDataArrivalKindLogs)
	items := findReceiverEvidence(t, model.FindXAgentDataArrivalKindLogs)
	if len(items) != 1 || items[0].Status != model.FindXAgentDataArrivalStatusReported {
		t.Fatalf("expected reported logs evidence, got %#v", items)
	}
	assertLogsEvidenceMetadata(t, items[0].Metadata)
}

func TestFindXAgentTracesCompatibleReceiverPersistsEvidence(t *testing.T) {
	resetCategrafReceiverTestState(t)
	body := strings.NewReader(`{
		"target_id":"target-trace-1",
		"trace_id":"trace-1",
		"source":"findx-agent",
		"scope":"kubernetes",
		"service":"checkout",
		"spans":[{"span_id":"span-1"},{"span_id":"span-2"}],
		"metadata":{"region":"cn","session_id":"should-not-store"}
	}`)
	w := performCategrafReceiverPost("/findx-agent/v1/traces", body, "application/json", "", FindXAgentTracesCompatibleReceiver)
	if w.Code != http.StatusOK {
		t.Fatalf("traces receiver should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	assertFindXAgentReceiverSmallResponse(t, w.Body.String(), model.FindXAgentDataArrivalKindTracing)
	items := findReceiverEvidence(t, model.FindXAgentDataArrivalKindTracing)
	if len(items) != 1 || items[0].Status != model.FindXAgentDataArrivalStatusReported {
		t.Fatalf("expected reported tracing evidence, got %#v", items)
	}
	assertTracingEvidenceMetadata(t, items[0].Metadata)
}

func TestFindXAgentLogTraceReceiversRejectInvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		body    string
		handler gin.HandlerFunc
	}{
		{"logs empty body", "/findx-agent/v1/logs", "", FindXAgentLogsCompatibleReceiver},
		{"logs invalid json", "/findx-agent/v1/logs", "{", FindXAgentLogsCompatibleReceiver},
		{"logs missing identity", "/findx-agent/v1/logs", `{"records":[{"message":"x"}]}`, FindXAgentLogsCompatibleReceiver},
		{"logs missing record body", "/findx-agent/v1/logs", `{"agent_id":"a","records":[{}]}`, FindXAgentLogsCompatibleReceiver},
		{"logs missing records", "/findx-agent/v1/logs", `{"agent_id":"a"}`, FindXAgentLogsCompatibleReceiver},
		{"traces empty body", "/findx-agent/v1/traces", "", FindXAgentTracesCompatibleReceiver},
		{"traces invalid json", "/findx-agent/v1/traces", "{", FindXAgentTracesCompatibleReceiver},
		{"traces missing identity", "/findx-agent/v1/traces", `{"trace_id":"t","spans":[{"span_id":"s"}]}`, FindXAgentTracesCompatibleReceiver},
		{"traces missing trace id", "/findx-agent/v1/traces", `{"agent_id":"a","spans":[{"span_id":"s"}]}`, FindXAgentTracesCompatibleReceiver},
		{"traces missing span id", "/findx-agent/v1/traces", `{"agent_id":"a","trace_id":"t","spans":[{}]}`, FindXAgentTracesCompatibleReceiver},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCategrafReceiverTestState(t)
			w := performCategrafReceiverPost(tt.path, strings.NewReader(tt.body), "application/json", "", tt.handler)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("invalid receiver input should be 400, got %d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestFindXAgentLogTraceReceiversRejectTooLargeBody(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		handler gin.HandlerFunc
	}{
		{"logs too large", "/findx-agent/v1/logs", FindXAgentLogsCompatibleReceiver},
		{"traces too large", "/findx-agent/v1/traces", FindXAgentTracesCompatibleReceiver},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCategrafReceiverTestState(t)
			body := strings.NewReader(strings.Repeat("x", categrafReceiverBodyLimit+1))
			w := performCategrafReceiverPost(tt.path, body, "application/json", "", tt.handler)
			if w.Code != http.StatusRequestEntityTooLarge {
				t.Fatalf("oversized receiver body should be 413, got %d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestFindXAgentLogTraceReceiversRejectInvalidGzipBody(t *testing.T) {
	for _, tt := range []struct {
		name    string
		path    string
		handler gin.HandlerFunc
	}{
		{"logs", "/findx-agent/v1/logs", FindXAgentLogsCompatibleReceiver},
		{"traces", "/findx-agent/v1/traces", FindXAgentTracesCompatibleReceiver},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetCategrafReceiverTestState(t)
			w := performCategrafReceiverPost(tt.path, strings.NewReader("not-gzip"), "application/json", "gzip", tt.handler)
			if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "invalid gzip receiver payload") {
				t.Fatalf("bad gzip should return generic receiver 400, got %d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestFindXAgentLogTraceReceiverErrorsDoNotEchoInput(t *testing.T) {
	resetCategrafReceiverTestState(t)
	w := performCategrafReceiverPost(
		"/findx-agent/v1/logs",
		strings.NewReader(`{"records":[{"message":"sensitive-log-body"}],"metadata":{"password":"do-not-store"}}`),
		"application/json",
		"",
		FindXAgentLogsCompatibleReceiver,
	)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid logs payload should be 400, got %d body=%s", w.Code, w.Body.String())
	}
	assertBodyDoesNotContainReceiverSecrets(t, w.Body.String())
}

func TestFindXAgentLogTraceReceiversReuseReceiverTokenGate(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-receiver-token", false)
	assertReceiverTokenGate(t, "logs", "/findx-agent/v1/logs",
		`{"agent_id":"agent-log-token","records":[{"body":"ok"}]}`,
		FindXAgentLogsCompatibleReceiver)
	assertReceiverTokenGate(t, "traces", "/findx-agent/v1/traces",
		`{"agent_id":"agent-trace-token","trace_id":"trace-token","spans":[{"span_id":"span-token"}]}`,
		FindXAgentTracesCompatibleReceiver)
}

func TestFindXAgentLogTraceEvidenceMergesIntoDataArrivalSummary(t *testing.T) {
	items := agentDataArrival([]*model.FindXAgent{{
		ID:           "agent-merge",
		Status:       "online",
		Capabilities: []string{"logs", "tracing"},
	}})
	merged := mergeDataArrivalEvidence(items, map[string]model.FindXAgentDataArrival{
		model.FindXAgentDataArrivalKindLogs: {
			Kind:          model.FindXAgentDataArrivalKindLogs,
			Status:        model.FindXAgentDataArrivalStatusReported,
			EvidenceCount: 1,
		},
		model.FindXAgentDataArrivalKindTracing: {
			Kind:          model.FindXAgentDataArrivalKindTracing,
			Status:        model.FindXAgentDataArrivalStatusReported,
			EvidenceCount: 1,
		},
	})
	assertDataArrivalReported(t, merged, model.FindXAgentDataArrivalKindLogs)
	assertDataArrivalReported(t, merged, model.FindXAgentDataArrivalKindTracing)
}

func TestFindXAgentLogTraceEvidenceListDoesNotLeakPayload(t *testing.T) {
	resetCategrafReceiverTestState(t)
	logs := performCategrafReceiverPost(
		"/findx-agent/v1/logs",
		strings.NewReader(`{"agent_id":"agent-log-list","records":[{"message":"sensitive-log-body"}],"metadata":{"password":"do-not-store"}}`),
		"application/json",
		"",
		FindXAgentLogsCompatibleReceiver,
	)
	if logs.Code != http.StatusOK {
		t.Fatalf("logs receiver should be 200, got %d body=%s", logs.Code, logs.Body.String())
	}
	traces := performCategrafReceiverPost(
		"/findx-agent/v1/traces",
		strings.NewReader(`{"agent_id":"agent-trace-list","trace_id":"trace-list","spans":[{"span_id":"span-secret"}],"metadata":{"session_id":"should-not-store"}}`),
		"application/json",
		"",
		FindXAgentTracesCompatibleReceiver,
	)
	if traces.Code != http.StatusOK {
		t.Fatalf("traces receiver should be 200, got %d body=%s", traces.Code, traces.Body.String())
	}
	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival-evidence", ListFindXAgentDataArrivalEvidence)
	if w.Code != http.StatusOK {
		t.Fatalf("data arrival evidence list should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	assertBodyDoesNotContainReceiverSecrets(t, w.Body.String())
}

func TestFindXAgentLogTraceRootRoutesAreRegistered(t *testing.T) {
	router := gin.New()
	router.POST("/findx-agent/v1/logs", FindXAgentLogsCompatibleReceiver)
	router.POST("/findx-agent/v1/traces", FindXAgentTracesCompatibleReceiver)
	paths := map[string]bool{}
	for _, route := range router.Routes() {
		paths[route.Path] = route.Method == http.MethodPost
	}
	if !paths["/findx-agent/v1/logs"] || !paths["/findx-agent/v1/traces"] {
		t.Fatalf("expected logs/traces root receiver routes, got %#v", router.Routes())
	}
}

func performFindXAgentLogTraceTokenRequest(path, body, token string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	return performCategrafReceiverRequest(categrafReceiverRequest{
		path:        path,
		body:        strings.NewReader(body),
		contentType: "application/json",
		remoteAddr:  "198.51.100.10:51000",
		token:       token,
		handler:     handler,
	})
}

func assertReceiverTokenGate(t *testing.T, name, path, body string, handler gin.HandlerFunc) {
	t.Helper()
	t.Run(name+" loopback no token", func(t *testing.T) {
		w := performCategrafReceiverPost(path, strings.NewReader(body), "application/json", "", handler)
		if w.Code != http.StatusOK {
			t.Fatalf("loopback %s request without token should be 200, got %d body=%s", name, w.Code, w.Body.String())
		}
	})
	t.Run(name+" non-loopback no token", func(t *testing.T) {
		w := performFindXAgentLogTraceTokenRequest(path, body, "", handler)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("non-loopback %s request without token should be 401, got %d body=%s", name, w.Code, w.Body.String())
		}
	})
	t.Run(name+" non-loopback with token", func(t *testing.T) {
		w := performFindXAgentLogTraceTokenRequest(path, body, "unit-receiver-token", handler)
		if w.Code != http.StatusOK {
			t.Fatalf("non-loopback %s request with token should be 200, got %d body=%s", name, w.Code, w.Body.String())
		}
	})
}

func assertDataArrivalReported(t *testing.T, rows []model.FindXAgentDataArrival, kind string) {
	t.Helper()
	for _, row := range rows {
		if row.Kind == kind {
			if row.Status != model.FindXAgentDataArrivalStatusReported || row.EvidenceCount != 1 || row.Blocker != "" {
				t.Fatalf("expected %s reported with evidence, got %#v", kind, row)
			}
			return
		}
	}
	t.Fatalf("missing data-arrival kind %s in %#v", kind, rows)
}

func assertLogsEvidenceMetadata(t *testing.T, metadata map[string]string) {
	t.Helper()
	if metadata["count"] != "1" || metadata["source"] != "filelog" || metadata["service"] != "orders" {
		t.Fatalf("unexpected logs evidence metadata: %#v", metadata)
	}
	if metadata["trace_id"] != "trace-log-1" || metadata["region"] != "cn" || metadata["env"] != "prod" {
		t.Fatalf("logs evidence metadata missing safe values: %#v", metadata)
	}
	assertMetadataDoesNotContainSensitiveValues(t, metadata)
}

func assertTracingEvidenceMetadata(t *testing.T, metadata map[string]string) {
	t.Helper()
	if metadata["span_count"] != "2" || metadata["trace_id"] != "trace-1" {
		t.Fatalf("unexpected tracing evidence metadata: %#v", metadata)
	}
	if metadata["service"] != "checkout" || metadata["scope"] != "kubernetes" || metadata["region"] != "cn" {
		t.Fatalf("tracing evidence metadata missing safe values: %#v", metadata)
	}
	assertMetadataDoesNotContainSensitiveValues(t, metadata)
}

func assertFindXAgentReceiverSmallResponse(t *testing.T, body, kind string) {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("response should be json: %v body=%s", err, body)
	}
	if payload["ok"] != true || payload["status"] != model.FindXAgentDataArrivalStatusReported || payload["kind"] != kind {
		t.Fatalf("unexpected receiver response: %#v", payload)
	}
	for _, key := range []string{"message", "body", "log", "metadata", "span_id", "trace_id"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("receiver response leaked key %q: %#v", key, payload)
		}
	}
	for _, marker := range []string{"trace-log-1", "user login failed"} {
		if strings.Contains(body, marker) {
			t.Fatalf("receiver response leaked value %q: %s", marker, body)
		}
	}
}

func assertMetadataDoesNotContainSensitiveValues(t *testing.T, metadata map[string]string) {
	t.Helper()
	for key, value := range metadata {
		combined := key + "=" + value
		for _, marker := range []string{"password", "credential", "session", "do-not-store", "<CREDENTIAL_REF>", "should-not-store"} {
			if strings.Contains(strings.ToLower(combined), strings.ToLower(marker)) {
				t.Fatalf("metadata leaked sensitive marker %q: %#v", marker, metadata)
			}
		}
	}
}

func assertBodyDoesNotContainReceiverSecrets(t *testing.T, body string) {
	t.Helper()
	for _, marker := range []string{
		"sensitive-log-body",
		"span-secret",
		"span_id",
		"password",
		"credential",
		"session",
		"do-not-store",
		"should-not-store",
	} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(marker)) {
			t.Fatalf("receiver API response leaked sensitive marker %q: %s", marker, body)
		}
	}
}
