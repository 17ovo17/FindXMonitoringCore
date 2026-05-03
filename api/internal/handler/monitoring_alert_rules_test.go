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

func performTryRunRequest(path string, body any) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/:id/tryrun", TryRunMonitorAlertRule)
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestTryRunMonitorAlertRuleExecutesPrometheus(t *testing.T) {
	var requestedPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"instance":"host-a","token":"<TOKEN>"},"value":[1,"1"]}]}}`))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	beforeCurrent, beforeHistory := len(store.ListMonitorAlertEvents(true)), len(store.ListMonitorAlertEvents(false))
	query := `sum(rate(http_requests_total{secret_token="<TOKEN>"}[5m])) > 0`
	w := performTryRunRequest("/rule-tryrun/tryrun", tryRunRule(query))
	if w.Code != http.StatusOK {
		t.Fatalf("tryrun should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	if requestedPath != "/api/v1/query" {
		t.Fatalf("prometheus query endpoint not called, path=%s", requestedPath)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"promql_executed":true`) || !strings.Contains(body, `"stats"`) || !strings.Contains(body, `"series_count":1`) {
		t.Fatalf("response missing eval execution fields: %s", body)
	}
	if strings.Contains(body, query) || strings.Contains(body, "<TOKEN>") {
		t.Fatalf("tryrun response leaked raw query or secret label: %s", body)
	}
	if got := len(store.ListMonitorAlertEvents(true)); got != beforeCurrent {
		t.Fatalf("tryrun wrote current events, before=%d after=%d", beforeCurrent, got)
	}
	if got := len(store.ListMonitorAlertEvents(false)); got != beforeHistory {
		t.Fatalf("tryrun wrote history events, before=%d after=%d", beforeHistory, got)
	}
}

func TestTryRunMonitorAlertRuleSuccessWarningsDoNotLeak(t *testing.T) {
	sensitiveWarning := "remote http://prom/api?" + "api_key" + "=<API_KEY>&" + "cookie" + "=<COOKIE>"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := `{"status":"success","data":{"resultType":"vector","result":[]},"warnings":["` + sensitiveWarning + `","partial scrape"]}`
		_, _ = w.Write([]byte(payload))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	w := performTryRunRequest("/rule-tryrun/tryrun", tryRunRule("up == 0"))
	if w.Code != http.StatusOK {
		t.Fatalf("tryrun should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	for _, forbidden := range []string{"<API_KEY>", "<COOKIE>", "api_key", "cookie"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			t.Fatalf("tryrun response leaked sensitive warning fragment %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, "REDACTED") || !strings.Contains(body, "partial scrape") {
		t.Fatalf("tryrun response missing sanitized warnings: %s", body)
	}
}

func TestTryRunMonitorAlertRuleUpstreamSecretDoesNotLeak(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`authorization token secret`))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	w := performTryRunRequest("/rule-tryrun/tryrun", tryRunRule("up == 0"))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("upstream failure should be 503, got %d body=%s", w.Code, w.Body.String())
	}
	body := strings.ToLower(w.Body.String())
	for _, marker := range []string{"authorization", "token secret", "badgateway"} {
		if strings.Contains(body, marker) {
			t.Fatalf("upstream secret leaked in response: %s", w.Body.String())
		}
	}
}

func TestTryRunMonitorAlertRuleStatusErrorIsServiceUnavailable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"error","error":"upstream token secret"}`))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	w := performTryRunRequest("/rule-tryrun/tryrun", tryRunRule("up == 0"))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status:error should be 503, got %d body=%s", w.Code, w.Body.String())
	}
	body := strings.ToLower(w.Body.String())
	for _, marker := range []string{"token", "secret"} {
		if strings.Contains(body, marker) {
			t.Fatalf("status:error leaked upstream body: %s", w.Body.String())
		}
	}
}

func TestTryRunMonitorAlertRuleInvalidJSONIsServiceUnavailable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json token secret`))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	w := performTryRunRequest("/rule-tryrun/tryrun", tryRunRule("up == 0"))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("invalid json should be 503, got %d body=%s", w.Code, w.Body.String())
	}
	body := strings.ToLower(w.Body.String())
	for _, marker := range []string{"token", "secret", "not-json"} {
		if strings.Contains(body, marker) {
			t.Fatalf("invalid json leaked upstream body: %s", w.Body.String())
		}
	}
}

func TestTryRunMonitorAlertRuleInvalidDoesNotRequestPrometheus(t *testing.T) {
	requests := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	rule := tryRunRule("")
	w := performTryRunRequest("/rule-tryrun/tryrun", rule)
	if w.Code != http.StatusOK {
		t.Fatalf("invalid tryrun should stay 200, got %d body=%s", w.Code, w.Body.String())
	}
	if requests != 0 {
		t.Fatalf("invalid rule requested upstream %d times", requests)
	}
	if !strings.Contains(w.Body.String(), `"ok":false`) || !strings.Contains(w.Body.String(), `"status":"invalid"`) {
		t.Fatalf("invalid response contract changed: %s", w.Body.String())
	}
}

func TestTryRunMonitorAlertRuleEvalLogDetailsDoNotContainRawQuery(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	query := `up{password="<PASSWORD>"} == 0`
	w := performTryRunRequest("/rule-tryrun/tryrun", tryRunRule(query))
	if w.Code != http.StatusOK {
		t.Fatalf("tryrun should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(body, query) || strings.Contains(body, "<PASSWORD>") {
		t.Fatalf("eval response leaked raw query in details: %s", body)
	}
	if !strings.Contains(body, `"query_hash"`) || strings.Contains(body, `"query":"`) && !strings.Contains(body, `"query":""`) {
		t.Fatalf("expected query hash and redacted rule query: %s", body)
	}
}

func TestTryRunMonitorAlertRuleDatasourceNotFound(t *testing.T) {
	configureMonitorPrometheus(t, "http://127.0.0.1:9090")
	rule := tryRunRule("up == 0")
	rule.DatasourceID = "missing-prometheus"
	w := performTryRunRequest("/rule-tryrun/tryrun", rule)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing datasource should be 404, got %d body=%s", w.Code, w.Body.String())
	}
}

func tryRunRule(query string) model.MonitorAlertRule {
	return model.MonitorAlertRule{
		ID: "rule-tryrun", Name: "TryRun Rule", Query: query, Severity: model.MonitorAlertSeverityWarning,
		DatasourceID: "prometheus-default", Enabled: true, Version: 1,
		ForDuration: "5m", NoDataPolicy: model.MonitorNoDataPolicyOK,
	}
}
