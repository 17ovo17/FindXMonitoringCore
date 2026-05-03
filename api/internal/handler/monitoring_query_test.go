package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func performMonitorRequest(method, path string, handler gin.HandlerFunc, body any) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	routePath := path
	if idx := strings.Index(routePath, "?"); idx >= 0 {
		routePath = routePath[:idx]
	}
	r.Handle(method, routePath, handler)
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

func configureMonitorPrometheus(t *testing.T, upstream string) {
	t.Helper()
	viper.Reset()
	viper.Set("data_sources", []map[string]any{{
		"id": "prometheus-default", "name": "Prometheus", "type": "prometheus", "url": upstream,
	}})
	viper.Set("prometheus.url", upstream)
	t.Cleanup(viper.Reset)
}

func TestMonitorQueryRejectsEmptyQuery(t *testing.T) {
	configureMonitorPrometheus(t, "http://127.0.0.1:9090")
	w := performMonitorRequest(http.MethodPost, "/", MonitorQuery, map[string]any{"query": " "})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty query should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestMonitorLabelValuesRejectsInvalidLabel(t *testing.T) {
	configureMonitorPrometheus(t, "http://127.0.0.1:9090")
	w := performMonitorRequest(http.MethodGet, "/?label=bad-name", ListMonitorLabelValues, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid label should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestMonitorQueryRangeRejectsEndBeforeStart(t *testing.T) {
	configureMonitorPrometheus(t, "http://127.0.0.1:9090")
	w := performMonitorRequest(http.MethodPost, "/", MonitorQueryRange, map[string]any{"query": "up", "start": 100, "end": 100, "step": 10})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("end<=start should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestMonitorQueryRangeRejectsTooManyPoints(t *testing.T) {
	configureMonitorPrometheus(t, "http://127.0.0.1:9090")
	w := performMonitorRequest(http.MethodPost, "/", MonitorQueryRange, map[string]any{"query": "up", "start": 0, "end": 11001, "step": 1})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("too many range points should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestSanitizeDatasourceURLRedactsUserinfoAndSecrets(t *testing.T) {
	raw := "http://login-user:login-value@example.test:9090/prom?token=%3CTOKEN%3E&x=1&api_key=%3CAPI_KEY%3E&password=%3CPASSWORD%3E"
	got := sanitizeDatasourceURL(raw)
	if strings.Contains(got, "login-value") || strings.Contains(got, "%3CTOKEN%3E") || strings.Contains(got, "%3CAPI_KEY%3E") || strings.Contains(got, "%3CPASSWORD%3E") {
		t.Fatalf("sanitized url leaked secret: %s", got)
	}
	if !strings.Contains(got, "token=%3CREDACTED%3E") || !strings.Contains(got, "api_key=%3CREDACTED%3E") || !strings.Contains(got, "password=%3CREDACTED%3E") || !strings.Contains(got, "x=1") {
		t.Fatalf("sanitized url missing expected query handling: %s", got)
	}
}

func TestMonitorQueryUsesTimeoutMS(t *testing.T) {
	req := monitoringPromRequest{Query: "up", TimeoutMS: 2500, TimeoutSeconds: 9}
	path, params, timeout := monitorPromRequestTarget(req, false)
	if path != "/api/v1/query" || params.Get("query") != "up" {
		t.Fatalf("unexpected instant target path=%s params=%v", path, params)
	}
	if timeout != 2500*time.Millisecond {
		t.Fatalf("timeout_ms should take priority, got %s", timeout)
	}
}

func TestMonitorQueryRangeUsesTimeoutMS(t *testing.T) {
	req := monitoringPromRequest{Query: "up", Start: 1, End: 10, Step: 1, TimeoutMS: 1500}
	path, params, timeout := monitorPromRequestTarget(req, true)
	if path != "/api/v1/query_range" || params.Get("start") != "1" || params.Get("end") != "10" || params.Get("step") != "1" {
		t.Fatalf("unexpected range target path=%s params=%v", path, params)
	}
	if timeout != 1500*time.Millisecond {
		t.Fatalf("timeout_ms should be used, got %s", timeout)
	}
}

func TestMonitorQueryReturnsRawPrometheusResult(t *testing.T) {
	var requestedPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","job":"node"},"value":[1710000000,"1"]}]},"warnings":["partial"]}`))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	w := performMonitorRequest(http.MethodPost, "/", MonitorQuery, map[string]any{"query": "up"})
	if w.Code != http.StatusOK {
		t.Fatalf("query should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	if requestedPath != "/api/v1/query" {
		t.Fatalf("unexpected upstream path: %s", requestedPath)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"resultType":"vector"`) || !strings.Contains(body, `"result"`) || !strings.Contains(body, `"query_hash"`) || !strings.Contains(body, `"warnings":["partial"]`) {
		t.Fatalf("response missing raw prometheus fields: %s", body)
	}
	if !strings.Contains(body, `"stats"`) || !strings.Contains(body, `"result_type":"vector"`) || !strings.Contains(body, `"series_count":1`) || !strings.Contains(body, `"sample_count":1`) {
		t.Fatalf("response missing query stats: %s", body)
	}
	events := store.ListAuditEvents(5)
	if len(events) == 0 || events[0].Action != "monitor.query" || !strings.Contains(events[0].Detail, "query_hash=") || strings.Contains(events[0].Detail, "query=up") {
		t.Fatalf("query audit not recorded or leaked raw query: %+v", events)
	}
}

func TestMonitorQuerySuccessWarningsAreSanitized(t *testing.T) {
	dsnWarning := "upstream dsn=<DB_DSN>"
	urlWarning := "remote http://prom/api?" + "auth" + "=<TOKEN>&" + "private" + "=<PRIVATE>"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := `{"status":"success","data":{"resultType":"vector","result":[]},"warnings":["partial scrape","` +
			dsnWarning + `","` + urlWarning + `"]}`
		_, _ = w.Write([]byte(payload))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	w := performMonitorRequest(http.MethodPost, "/", MonitorQuery, map[string]any{"query": "up"})
	if w.Code != http.StatusOK {
		t.Fatalf("query should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "partial scrape") || !strings.Contains(body, "REDACTED") {
		t.Fatalf("response lost safe warning or redaction marker: %s", body)
	}
	for _, forbidden := range []string{"<DB_DSN>", "<TOKEN>", "<PRIVATE>", "auth", "private", "dsn"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			t.Fatalf("query response leaked sensitive warning fragment %q: %s", forbidden, body)
		}
	}
}

func TestMonitorQueryMapsPrometheusFailureToServiceUnavailable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway " + "token" + "=<TOKEN>"))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	w := performMonitorRequest(http.MethodPost, "/", MonitorQuery, map[string]any{"query": "up"})
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("non-2xx upstream should be 503, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(strings.ToLower(w.Body.String()), "secret") || strings.Contains(strings.ToLower(w.Body.String()), "token") {
		t.Fatalf("error response leaked upstream secret: %s", w.Body.String())
	}
}

func TestMonitorQueryMapsPrometheusNonSuccessToServiceUnavailable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"error","error":"upstream token secret"}`))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	w := performMonitorRequest(http.MethodPost, "/", MonitorQuery, map[string]any{"query": "up"})
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("non-success upstream should be 503, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(strings.ToLower(w.Body.String()), "secret") || strings.Contains(strings.ToLower(w.Body.String()), "token") {
		t.Fatalf("error response leaked upstream secret: %s", w.Body.String())
	}
}

func TestMonitorQueryRangeSortsMatrixSeries(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"instance":"b"},"values":[[1,"2"]]},{"metric":{"instance":"a"},"values":[[1,"1"]]}]}}`))
	}))
	defer upstream.Close()
	configureMonitorPrometheus(t, upstream.URL)

	w := performMonitorRequest(http.MethodPost, "/", MonitorQueryRange, map[string]any{"query": "up", "start": 1, "end": 10, "step": 1})
	if w.Code != http.StatusOK {
		t.Fatalf("query_range should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Index(w.Body.String(), `"instance":"a"`) > strings.Index(w.Body.String(), `"instance":"b"`) {
		t.Fatalf("matrix series not sorted: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"result_type":"matrix"`) || !strings.Contains(w.Body.String(), `"series_count":2`) || !strings.Contains(w.Body.String(), `"sample_count":2`) {
		t.Fatalf("matrix stats missing or wrong: %s", w.Body.String())
	}
}

func TestMonitorMetricsFiltersMappings(t *testing.T) {
	dsID := "prometheus-default"
	configureMonitorPrometheus(t, "http://127.0.0.1:9090")
	suffix := time.Now().Format("150405.000000000")
	nodeMetric := &model.MetricsMapping{ID: "monitor-query-node-" + suffix, DatasourceID: dsID, RawName: "mysql_global_status_queries", StandardName: "mysql.qps", Exporter: "mysqld_exporter", Status: "confirmed"}
	redisMetric := &model.MetricsMapping{ID: "monitor-query-redis-" + suffix, DatasourceID: dsID, RawName: "redis_up", StandardName: "redis.up", Status: "unmapped"}
	store.SaveMetricsMapping(nodeMetric)
	store.SaveMetricsMapping(redisMetric)

	w := performMonitorRequest(http.MethodGet, "/?datasource_id="+dsID+"&q=queries&status=confirmed&category=mysql&exporter=mysqld_exporter&limit=500", ListMonitorMetrics, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("metrics should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, nodeMetric.ID) || strings.Contains(body, redisMetric.ID) || !strings.Contains(body, `"promql"`) {
		t.Fatalf("metrics filters did not apply: %s", body)
	}
}
