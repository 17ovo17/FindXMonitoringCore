package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestContractMatrixReadyRequiresExecutableContract(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	badReady := performContractMatrixRequest(t, r, http.MethodPost, "/contract-matrix", map[string]any{
		"id":         "FX-CONTRACT-AGENT-INSTALL",
		"domain":     "agent",
		"capability": "install",
		"status":     model.ContractStatusReady,
		"handler":    "InstallFindXAgent",
	})
	if badReady.Code != http.StatusBadRequest {
		t.Fatalf("ready without backend/datasource/executor/evidence should fail, got %d body=%s", badReady.Code, badReady.Body.String())
	}

	ready := performContractMatrixRequest(t, r, http.MethodPost, "/contract-matrix", map[string]any{
		"id":            "FX-CONTRACT-AGENT-INSTALL",
		"domain":        "agent",
		"capability":    "install",
		"status":        model.ContractStatusReady,
		"handler":       "InstallFindXAgent",
		"backend":       "findx_agent_lifecycle",
		"datasource":    "agent_package_repository",
		"executor":      "ssh_executor",
		"evidence_refs": []string{"source:AutoOps serviceDeploy.go"},
	})
	if ready.Code != http.StatusOK {
		t.Fatalf("executable ready contract should pass, got %d body=%s", ready.Code, ready.Body.String())
	}
	var item model.ContractMatrixEntry
	decodeContractMatrixResponse(t, ready, &item)
	if item.Status != model.ContractStatusReady || item.ID != "FX-CONTRACT-AGENT-INSTALL" {
		t.Fatalf("unexpected ready item: %#v", item)
	}
}

func TestContractMatrixGapStatusesAndBlockedResponse(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	tests := []struct {
		name   string
		id     string
		status string
	}{
		{name: "missing backend", id: "FX-CONTRACT-LOGS-LIVE-BACKEND", status: model.ContractStatusMissingBackend},
		{name: "missing datasource", id: "FX-CONTRACT-APM-OAP-DATASOURCE", status: model.ContractStatusMissingDatasource},
		{name: "missing executor", id: "FX-CONTRACT-AGENT-EXECUTOR-SSH", status: model.ContractStatusMissingExecutor},
		{name: "unsafe", id: "FX-CONTRACT-UNSAFE-INPUT", status: model.ContractStatusUnsafe},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performContractMatrixRequest(t, r, http.MethodPost, "/contract-matrix", map[string]any{
				"id":         tt.id,
				"domain":     "agent",
				"capability": tt.name,
				"status":     tt.status,
			})
			if w.Code != http.StatusConflict {
				t.Fatalf("gap registration should be blocked 409, got %d body=%s", w.Code, w.Body.String())
			}
			var payload model.ContractMatrixBlockedResponse
			decodeContractMatrixResponse(t, w, &payload)
			if payload.Code != model.ContractBlockedByContractCode || payload.ContractGapID != tt.id || payload.Status != tt.status || payload.SafeToRetry {
				t.Fatalf("blocked response mismatch: %#v", payload)
			}
			if !strings.Contains(payload.Message, "能力缺少后端、数据源或执行器契约") {
				t.Fatalf("blocked response message must be readable UTF-8 Chinese, got %#v", payload.Message)
			}
			assertContractMatrixNoMojibake(t, payload.Message)
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixNoFakeRuntimeState(t, w.Body.String())
		})
	}
}

func TestContractMatrixQueryReturnsRegisteredGapID(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	created := performContractMatrixRequest(t, r, http.MethodPost, "/contract-matrix", map[string]any{
		"id":             "FX-CONTRACT-AGENT-EXECUTOR-SSH",
		"domain":         "agent",
		"capability":     "ssh executor",
		"status":         model.ContractStatusMissingExecutor,
		"blocked_reason": "executor contract missing",
	})
	if created.Code != http.StatusConflict {
		t.Fatalf("gap create should be conflict, got %d body=%s", created.Code, created.Body.String())
	}
	w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/FX-CONTRACT-AGENT-EXECUTOR-SSH", nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("gap detail should be blocked 409, got %d body=%s", w.Code, w.Body.String())
	}
	var payload model.ContractMatrixBlockedResponse
	decodeContractMatrixResponse(t, w, &payload)
	if payload.ContractGapID != "FX-CONTRACT-AGENT-EXECUTOR-SSH" || payload.Status != model.ContractStatusMissingExecutor {
		t.Fatalf("frontend must be able to map blocked response to gap id, got %#v", payload)
	}
}

func TestContractMatrixSeededMonitoringGapDetailIsBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/FX-CONTRACT-N9E-DATASOURCE-PROXY-BY-ID", nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("seeded monitoring gap detail should be blocked 409, got %d body=%s", w.Code, w.Body.String())
	}
	var payload model.ContractMatrixBlockedResponse
	decodeContractMatrixResponse(t, w, &payload)
	if payload.ContractGapID != "FX-CONTRACT-N9E-DATASOURCE-PROXY-BY-ID" ||
		payload.Status != model.ContractStatusBlocked ||
		payload.Code != model.ContractBlockedByContractCode ||
		payload.SafeToRetry {
		t.Fatalf("seeded monitoring blocked response mismatch: %#v", payload)
	}
	if !strings.Contains(payload.Message, "能力缺少后端、数据源或执行器契约") {
		t.Fatalf("blocked response message must be readable UTF-8 Chinese, got %#v", payload.Message)
	}
	assertContractMatrixNoMojibake(t, payload.Message)
	assertContractMatrixNoSensitiveLeak(t, w.Body.String())
	assertContractMatrixBlockedShape(t, w.Body.String())
}

func TestContractMatrixSeededDatasourceProxySplitGapsAreBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	for _, id := range []string{
		"FX-CONTRACT-N9E-DATASOURCE-PROXY-LABELS",
		"FX-CONTRACT-N9E-DATASOURCE-PROXY-LABEL-VALUES",
		"FX-CONTRACT-N9E-DATASOURCE-PROXY-METRIC-NAMES",
		"FX-CONTRACT-N9E-DATASOURCE-PROXY-SERIES",
		"FX-CONTRACT-N9E-DATASOURCE-PROXY-BUILDINFO",
		"FX-CONTRACT-N9E-DATASOURCE-PROXY-QUERY",
		"FX-CONTRACT-N9E-DATASOURCE-PROXY-QUERY-RANGE",
		"FX-CONTRACT-N9E-DATASOURCE-PROXY-ES-SEARCH",
	} {
		t.Run(id, func(t *testing.T) {
			w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/"+id, nil)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s should return BLOCKED_BY_CONTRACT 409, got %d body=%s", id, w.Code, w.Body.String())
			}
			var payload model.ContractMatrixBlockedResponse
			decodeContractMatrixResponse(t, w, &payload)
			if payload.Code != model.ContractBlockedByContractCode ||
				payload.ContractGapID != id ||
				payload.Status != model.ContractStatusMissingDatasource ||
				payload.SafeToRetry {
				t.Fatalf("datasource proxy split gap response mismatch for %s: %#v", id, payload)
			}
			assertContractMatrixNoMojibake(t, payload.Message)
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
		})
	}
}

func TestContractMatrixSeededDatasourceBriefListIsReady(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/FX-CONTRACT-N9E-DATASOURCE-BRIEF-LIST", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("datasource brief list should be ready, got %d body=%s", w.Code, w.Body.String())
	}
	var item model.ContractMatrixEntry
	decodeContractMatrixResponse(t, w, &item)
	if item.Status != model.ContractStatusReady ||
		item.Handler == "" || item.Backend == "" || item.Datasource == "" || item.Executor == "" ||
		len(item.EvidenceRefs) == 0 {
		t.Fatalf("ready contract missing executable evidence: %#v", item)
	}
	if item.Metadata["upstream_ref"] != "/monitor/datasources" {
		t.Fatalf("ready contract must use FindX datasource adapter, got %#v", item.Metadata)
	}
	assertContractMatrixNoSensitiveLeak(t, w.Body.String())
}

func TestContractMatrixSeededPrometheusSingleQueryAdapterIsReady(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/FX-CONTRACT-FINDX-PROMETHEUS-SINGLE-QUERY", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("single query adapter should be ready, got %d body=%s", w.Code, w.Body.String())
	}
	var item model.ContractMatrixEntry
	decodeContractMatrixResponse(t, w, &item)
	if item.Status != model.ContractStatusReady ||
		item.Handler == "" || item.Backend == "" || item.Datasource == "" || item.Executor == "" ||
		len(item.EvidenceRefs) == 0 {
		t.Fatalf("single query ready contract missing executable evidence: %#v", item)
	}
	body := w.Body.String()
	for _, want := range []string{"/monitor/query", "/monitor/query-range", "/monitor/labels", "/monitor/label-values", "upstream_scope"} {
		if !strings.Contains(body, want) {
			t.Fatalf("single query adapter response missing %s: %s", want, body)
		}
	}
	for _, forbidden := range []string{"query-range-batch", "query-instant-batch", "/api/n9e-plus/query-batch", "/api/n9e/tag-pairs", "/api/n9e/tag-metrics", "/api/n9e/query", "/api/n9e/metric-views", "/api/n9e/prometheus/api/v1", "/api/n9e/share-charts", "/api/n9e/metrics/desc"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("single query adapter must not claim Nightingale batch/query view endpoint %s: %s", forbidden, body)
		}
	}
	assertContractMatrixNoSensitiveLeak(t, body)
}

func TestContractMatrixSeededMetricQueryBatchAndMetricViewsStayBlocked(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	tests := []struct {
		id     string
		status string
	}{
		{id: "FX-CONTRACT-N9E-METRIC-QUERY-BATCH", status: model.ContractStatusBlocked},
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-CRUD", status: model.ContractStatusBlocked},
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-LIST", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-CREATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-UPDATE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-METRIC-VIEWS-DELETE", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-QUERY-RANGE-BATCH", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-QUERY-INSTANT-BATCH", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-PLUS-QUERY-BATCH", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-TAG-PAIRS", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-TAG-METRICS", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-QUERY-DATA", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-QUERY-BENCH", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-PROMETHEUS-COMPAT-API", status: model.ContractStatusMissingDatasource},
		{id: "FX-CONTRACT-N9E-SHARE-CHARTS", status: model.ContractStatusMissingBackend},
		{id: "FX-CONTRACT-N9E-METRICS-DESC", status: model.ContractStatusMissingBackend},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/"+tt.id, nil)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s should remain blocked 409, got %d body=%s", tt.id, w.Code, w.Body.String())
			}
			var payload model.ContractMatrixBlockedResponse
			decodeContractMatrixResponse(t, w, &payload)
			if payload.ContractGapID != tt.id || payload.Status != tt.status || payload.Code != model.ContractBlockedByContractCode || payload.SafeToRetry {
				t.Fatalf("blocked response mismatch for %s: %#v", tt.id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
		})
	}
}

func TestContractMatrixSeededMetricQuerySplitGapsDoNotLeakRuntimeState(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	for _, id := range []string{
		"FX-CONTRACT-N9E-QUERY-INSTANT-BATCH",
		"FX-CONTRACT-N9E-PROMETHEUS-COMPAT-API",
		"FX-CONTRACT-N9E-METRICS-DESC",
	} {
		t.Run(id, func(t *testing.T) {
			w := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix/"+id, nil)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s should return BLOCKED_BY_CONTRACT 409, got %d body=%s", id, w.Code, w.Body.String())
			}
			var payload model.ContractMatrixBlockedResponse
			decodeContractMatrixResponse(t, w, &payload)
			if payload.Code != model.ContractBlockedByContractCode || payload.ContractGapID != id || payload.Status == model.ContractStatusReady || payload.SafeToRetry {
				t.Fatalf("blocked split gap response mismatch for %s: %#v", id, payload)
			}
			assertContractMatrixNoSensitiveLeak(t, w.Body.String())
			assertContractMatrixBlockedShape(t, w.Body.String())
		})
	}
}

func TestContractMatrixDropsSensitiveInput(t *testing.T) {
	store.ResetContractMatrixForTest()
	r := contractMatrixTestRouter()
	w := performContractMatrixRequest(t, r, http.MethodPost, "/contract-matrix", map[string]any{
		"id":             "FX-CONTRACT-LOGS-DATASOURCE",
		"domain":         "logs",
		"capability":     "query",
		"status":         model.ContractStatusMissingDatasource,
		"blocked_reason": "missing password passwd secret api_key access_key session privatekey token cookie DSN private key",
		"metadata": map[string]string{
			"password":   "secret",
			"api_key":    "api-key-value",
			"access_key": "access-key-value",
			"session":    "session-value",
			"privatekey": "private-key-value",
			"passwd":     "passwd-value",
			"note":       "safe source evidence",
		},
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("sensitive blocked gap should still be represented safely, got %d body=%s", w.Code, w.Body.String())
	}
	assertContractMatrixNoSensitiveLeak(t, w.Body.String())
	assertContractMatrixBlockedShape(t, w.Body.String())
	list := performContractMatrixRequest(t, r, http.MethodGet, "/contract-matrix", nil)
	assertContractMatrixNoSensitiveLeak(t, list.Body.String())
}

func contractMatrixTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/contract-matrix", ListContractMatrixEntries)
	r.POST("/contract-matrix", RegisterContractMatrixEntry)
	r.GET("/contract-matrix/:id", GetContractMatrixEntry)
	return r
}

func performContractMatrixRequest(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeContractMatrixResponse(t *testing.T, w *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
}

func assertContractMatrixNoSensitiveLeak(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{"token", "password", "passwd", "cookie", "dsn", "private key", "private_key", "privatekey", "bearer", "secret", "api_key", "apikey", "access_key", "session"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("contract matrix response leaked sensitive marker %q: %s", forbidden, body)
		}
	}
}

func assertContractMatrixBlockedShape(t *testing.T, body string) {
	t.Helper()
	for _, required := range []string{`"code"`, `"message"`, `"contract_gap_id"`, `"status"`, `"safe_to_retry":false`} {
		if !strings.Contains(body, required) {
			t.Fatalf("blocked response missing required field %s: %s", required, body)
		}
	}
	assertContractMatrixNoMojibake(t, body)
	assertContractMatrixNoFakeRuntimeState(t, body)
}

func assertContractMatrixNoFakeRuntimeState(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{"queued", "running", "succeeded", "success", "installed", "data_arrived", "applied", "rolled-back"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("blocked path must not expose fake runtime state %q: %s", forbidden, body)
		}
	}
}

func assertContractMatrixNoMojibake(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range contractMatrixMojibakeDenylist() {
		if strings.Contains(body, forbidden) {
			t.Fatalf("contract matrix response contains mojibake %q: %s", forbidden, body)
		}
	}
}

func contractMatrixMojibakeDenylist() []string {
	return []string{
		string([]rune{0x9473, 0x85c9, 0x59cf}),
		string([]rune{0x7f02, 0x54c4, 0x76af}),
		string([]rune{0x6924, 0x572d, 0x6d30}),
		string([]rune{0x9a9e, 0x51b2, 0x5f74}),
		string([]rune{0x0044, 0x951b}),
		string([]rune{0x0044, 0xf03a}),
		"\ufffd",
	}
}
