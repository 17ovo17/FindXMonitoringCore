package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/tracing"

	"github.com/gin-gonic/gin"
)

func TestAPMMissingOAPReturnsStructuredBlocked(t *testing.T) {
	r := tracingSkyWalkingTestRouter()
	resetSkyWalkingHandlerClient(t, &tracing.SWClient{})
	w := performTracingSkyWalkingRequest(r, http.MethodGet, "/apm/selectors/services", nil)
	assertAPMBlockedEnvelope(t, w, http.StatusServiceUnavailable, "FX-CONTRACT-TRACING-SERVICES")
}

func TestAPMSelectorRequiresServiceID(t *testing.T) {
	r := tracingSkyWalkingTestRouter()
	for _, path := range []string{"/apm/selectors/endpoints", "/apm/selectors/instances"} {
		w := performTracingSkyWalkingRequest(r, http.MethodGet, path, nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s missing serviceId should be 400, got %d body=%s", path, w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), `"endpoints":[]`) || strings.Contains(w.Body.String(), `"instances":[]`) {
			t.Fatalf("%s must not return fake empty success: %s", path, w.Body.String())
		}
	}
}

func TestAPMTraceQueryMalformedJSONReturns400(t *testing.T) {
	r := tracingSkyWalkingTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/apm/traces", strings.NewReader(`{"serviceId":`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("malformed trace query should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestAPMTraceDetailBlocksUntilContractExists(t *testing.T) {
	upstreamCalled := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		_, _ = w.Write([]byte(`{"data":{"trace":{"spans":[{"spanId":1,"status":"success"}]}}}`))
	}))
	defer upstream.Close()
	resetSkyWalkingHandlerClient(t, tracing.NewSWClientForTest(upstream.URL, upstream.Client()))

	r := tracingSkyWalkingTestRouter()
	w := performTracingSkyWalkingRequest(r, http.MethodGet, "/apm/traces/trace-1", nil)
	assertAPMBlockedEnvelope(t, w, http.StatusConflict, "FX-CONTRACT-TRACING-TRACE-DETAIL")
	body := w.Body.String()
	for _, want := range []string{"findx_tracing_trace_detail_query_contract", "findx_trace_log_agent_linkage_contract"} {
		if !strings.Contains(body, want) {
			t.Fatalf("trace detail blocked response missing %s: %s", want, body)
		}
	}
	if strings.Contains(body, `"spans"`) {
		t.Fatalf("trace detail must not return spans as fake success: %s", body)
	}
	if upstreamCalled {
		t.Fatalf("trace detail should block locally until the detail contract exists")
	}
}

func TestAPMUpstreamErrorIsSanitizedAndBlocked(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`token=secret-cookie dsn=mysql://root:pass@db/db`))
	}))
	defer upstream.Close()
	resetSkyWalkingHandlerClient(t, tracing.NewSWClientForTest(upstream.URL, upstream.Client()))

	r := tracingSkyWalkingTestRouter()
	w := performTracingSkyWalkingRequest(r, http.MethodGet, "/apm/selectors/services", nil)
	assertAPMBlockedEnvelope(t, w, http.StatusServiceUnavailable, "FX-CONTRACT-TRACING-SERVICES")
	assertNoAPMSensitiveLeak(t, w.Body.String())
}

func TestAPMGraphQLErrorIsSanitizedAndBlocked(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"token=secret-cookie dsn=mysql://root:pass@db/db"}]}`))
	}))
	defer upstream.Close()
	resetSkyWalkingHandlerClient(t, tracing.NewSWClientForTest(upstream.URL, upstream.Client()))

	r := tracingSkyWalkingTestRouter()
	w := performTracingSkyWalkingRequest(r, http.MethodGet, "/apm/selectors/services", nil)
	assertAPMBlockedEnvelope(t, w, http.StatusServiceUnavailable, "FX-CONTRACT-TRACING-SERVICES")
	assertNoAPMSensitiveLeak(t, w.Body.String())
}

func TestAPMUnimplementedContractsReturn409Blocked(t *testing.T) {
	r := tracingSkyWalkingTestRouter()
	tests := []struct {
		method string
		path   string
		body   any
	}{
		{method: http.MethodGet, path: "/apm/traces/trace-1/spans/span-1"},
		{method: http.MethodGet, path: "/apm/profiling/tasks"},
		{method: http.MethodPost, path: "/apm/profiling/tasks", body: map[string]any{"serviceId": "svc-1"}},
		{method: http.MethodPost, path: "/apm/profiling/tasks/task-1/cancel"},
		{method: http.MethodGet, path: "/apm/alarms"},
		{method: http.MethodPost, path: "/apm/alarms/alarm-1/ack"},
		{method: http.MethodGet, path: "/apm/settings"},
		{method: http.MethodPut, path: "/apm/settings", body: map[string]any{"endpoint": "http://example/graphql"}},
		{method: http.MethodGet, path: "/apm/agent-linkage"},
		{method: http.MethodGet, path: "/apm/topology?serviceId=svc-1"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			w := performTracingSkyWalkingRequest(r, tt.method, tt.path, tt.body)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s should return 409 blocked, got %d body=%s", tt.path, w.Code, w.Body.String())
			}
			assertAPMBlockedShape(t, w.Body.String())
			assertNoAPMFakeSuccessState(t, w.Body.String())
		})
	}
}

func TestAPMTraceTagRoutesValidateInputAndBlockMissingOAP(t *testing.T) {
	r := tracingSkyWalkingTestRouter()
	resetSkyWalkingHandlerClient(t, &tracing.SWClient{})
	keys := performTracingSkyWalkingRequest(r, http.MethodGet, "/apm/trace-tags/keys", nil)
	assertAPMBlockedEnvelope(t, keys, http.StatusServiceUnavailable, "FX-CONTRACT-TRACING-TRACE-TAG-KEYS")

	values := performTracingSkyWalkingRequest(r, http.MethodGet, "/apm/trace-tags/values", nil)
	if values.Code != http.StatusBadRequest {
		t.Fatalf("trace tag values missing tagKey should be 400, got %d body=%s", values.Code, values.Body.String())
	}
}

func TestTracingCompatibilityRoutesStayRegisteredAndBlocked(t *testing.T) {
	r := tracingSkyWalkingTestRouter()
	resetSkyWalkingHandlerClient(t, &tracing.SWClient{})
	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/tracing/selectors/services"},
		{method: http.MethodPost, path: "/tracing/traces/query"},
		{method: http.MethodGet, path: "/tracing/topology"},
	}
	for _, tt := range tests {
		w := performTracingSkyWalkingRequest(r, tt.method, tt.path, nil)
		if w.Code == http.StatusNotFound {
			t.Fatalf("%s should be registered", tt.path)
		}
		assertAPMBlockedEnvelope(t, w, http.StatusServiceUnavailable, "")
	}
}

func tracingSkyWalkingTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/tracing/selectors/services", TracingListServicesSW)
	r.POST("/tracing/traces/query", TracingQueryTracesSW)
	r.GET("/tracing/topology", TracingGetTopologySW)
	r.GET("/apm/selectors/services", TracingListServicesSW)
	r.GET("/apm/selectors/instances", TracingListInstancesSW)
	r.GET("/apm/selectors/endpoints", TracingListEndpointsSW)
	r.GET("/apm/topology", TracingGetTopologySW)
	r.GET("/apm/traces", APMQueryTracesSW)
	r.POST("/apm/traces", APMQueryTracesSW)
	r.GET("/apm/traces/:traceId", APMGetTraceSW)
	r.GET("/apm/traces/:traceId/spans/:spanId", APMGetSpanDetailSW)
	r.GET("/apm/trace-tags/keys", APMTraceTagKeysSW)
	r.GET("/apm/trace-tags/values", APMTraceTagValuesSW)
	r.GET("/apm/profiling/tasks", APMListProfilingTasksSW)
	r.POST("/apm/profiling/tasks", APMCreateProfilingTaskSW)
	r.POST("/apm/profiling/tasks/:id/cancel", APMCancelProfilingTaskSW)
	r.GET("/apm/alarms", APMListAlarmsSW)
	r.POST("/apm/alarms/:id/ack", APMAckAlarmSW)
	r.GET("/apm/settings", APMGetSettingsSW)
	r.PUT("/apm/settings", APMPutSettingsSW)
	r.GET("/apm/agent-linkage", APMAgentLinkageSW)
	return r
}

func resetSkyWalkingHandlerClient(t *testing.T, client *tracing.SWClient) {
	t.Helper()
	old := swClient
	swClient = client
	t.Cleanup(func() { swClient = old })
}

func performTracingSkyWalkingRequest(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func assertAPMBlockedEnvelope(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, contractID string) {
	t.Helper()
	if w.Code != wantStatus {
		t.Fatalf("blocked response status mismatch, got %d want %d body=%s", w.Code, wantStatus, w.Body.String())
	}
	assertAPMBlockedShape(t, w.Body.String())
	if contractID != "" && !strings.Contains(w.Body.String(), contractID) {
		t.Fatalf("blocked response missing contract id %s: %s", contractID, w.Body.String())
	}
	assertNoAPMSensitiveLeak(t, w.Body.String())
	assertNoAPMFakeSuccessState(t, w.Body.String())
}

func assertAPMBlockedShape(t *testing.T, body string) {
	t.Helper()
	for _, want := range []string{`"code":"BLOCKED_BY_CONTRACT"`, `"status":"blocked"`, `"contract_id"`, `"missing_contracts"`, `"safe_to_retry":false`} {
		if !strings.Contains(body, want) {
			t.Fatalf("blocked response missing %s: %s", want, body)
		}
	}
}

func assertNoAPMSensitiveLeak(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{"secret-cookie", "root:pass", "mysql://", "dsn=", "token="} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response leaked sensitive marker %q: %s", forbidden, body)
		}
	}
}

func assertNoAPMFakeSuccessState(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{"success", "succeeded", "queued", "running", "installed", "data_arrived", `"ok":true`, `"status":"ready"`} {
		if strings.Contains(strings.ToLower(body), forbidden) {
			t.Fatalf("response contains fake success state %q: %s", forbidden, body)
		}
	}
}
