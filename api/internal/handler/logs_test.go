package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestLogFieldsExposeBuiltinsAndBlockedDiscovery(t *testing.T) {
	r := logsTestRouter()
	w := performLogsRequest(t, r, http.MethodGet, "/logs/fields", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("fields should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp model.LogFieldsResponse
	decodeLogsResponse(t, w, &resp)
	if len(resp.Fields) < 10 || resp.LiveDiscovery.Status != "blocked" {
		t.Fatalf("unexpected fields response: %+v", resp)
	}
	if !strings.Contains(resp.LiveDiscovery.Blocker, model.LogsContractBlocked) {
		t.Fatalf("live discovery blocker must expose contract state: %+v", resp.LiveDiscovery)
	}
	assertNoLogSensitiveLeak(t, w.Body.String())
}

func TestExplorerSavedViewsCRUDForLogs(t *testing.T) {
	r := logsTestRouter()
	view := map[string]any{
		"sourcePage": "logs",
		"name":       "Errors " + time.Now().Format("150405.000000"),
		"query":      map[string]any{"q": "severity_text:ERROR", "token": "<TOKEN>"},
		"filters":    map[string]any{"service.name": "checkout"},
		"columns":    []string{"timestamp", "body", "trace_id"},
		"timeRange":  map[string]any{"from": "now-1h", "to": "now"},
	}
	create := performLogsRequest(t, r, http.MethodPost, "/explorer/views", view)
	if create.Code != http.StatusOK {
		t.Fatalf("create view should be 200, got %d body=%s", create.Code, create.Body.String())
	}
	var created model.ExplorerSavedView
	decodeLogsResponse(t, create, &created)
	if created.ID == "" || created.SourcePage != "logs" || created.Name == "" {
		t.Fatalf("created view mismatch: %+v", created)
	}
	assertNoLogSensitiveLeak(t, create.Body.String())

	list := performLogsRequest(t, r, http.MethodGet, "/explorer/views?sourcePage=logs", nil)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), created.ID) {
		t.Fatalf("list views should include created view, code=%d body=%s", list.Code, list.Body.String())
	}

	updatePayload := map[string]any{"sourcePage": "logs", "name": "Updated errors", "query": map[string]any{"q": "status:500"}}
	update := performLogsRequest(t, r, http.MethodPut, "/explorer/views/"+created.ID, updatePayload)
	if update.Code != http.StatusOK {
		t.Fatalf("update view should be 200, got %d body=%s", update.Code, update.Body.String())
	}
	var updated model.ExplorerSavedView
	decodeLogsResponse(t, update, &updated)
	if updated.ID != created.ID || updated.Name != "Updated errors" {
		t.Fatalf("updated view mismatch: %+v", updated)
	}

	del := performLogsRequest(t, r, http.MethodDelete, "/explorer/views/"+created.ID, nil)
	if del.Code != http.StatusOK {
		t.Fatalf("delete view should be 200, got %d body=%s", del.Code, del.Body.String())
	}
}

func TestLogPipelineSaveListAndPreview(t *testing.T) {
	r := logsTestRouter()
	version := "test-" + time.Now().Format("150405.000000")
	payload := map[string]any{
		"name":    "JSON parser",
		"version": version,
		"enabled": true,
		"stages":  []any{map[string]any{"id": "json", "type": "json"}},
		"config":  map[string]any{"parser": "json", "api_key": "<API_KEY>"},
	}
	create := performLogsRequest(t, r, http.MethodPost, "/logs/pipelines", payload)
	if create.Code != http.StatusOK {
		t.Fatalf("save pipeline should be 200, got %d body=%s", create.Code, create.Body.String())
	}
	var pipeline model.LogPipeline
	decodeLogsResponse(t, create, &pipeline)
	if pipeline.ID == "" || pipeline.Version != version || !pipeline.Enabled {
		t.Fatalf("saved pipeline mismatch: %+v", pipeline)
	}
	assertNoLogSensitiveLeak(t, create.Body.String())

	list := performLogsRequest(t, r, http.MethodGet, "/logs/pipelines/"+version, nil)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), pipeline.ID) {
		t.Fatalf("list pipeline should include saved item, code=%d body=%s", list.Code, list.Body.String())
	}

	before := list.Body.String()
	preview := performLogsRequest(t, r, http.MethodPost, "/logs/pipelines/preview", map[string]any{
		"parser": "json",
		"sample": `{"severity_text":"ERROR","body":"failed","password":"<PASSWORD>"}`,
	})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview should be 200, got %d body=%s", preview.Code, preview.Body.String())
	}
	if !strings.Contains(preview.Body.String(), `"severity_text"`) || strings.Contains(preview.Body.String(), "<PASSWORD>") {
		t.Fatalf("preview did not parse or leaked secret: %s", preview.Body.String())
	}
	after := performLogsRequest(t, r, http.MethodGet, "/logs/pipelines/"+version, nil)
	if after.Body.String() != before {
		t.Fatalf("preview must not mutate pipelines, before=%s after=%s", before, after.Body.String())
	}
}

func TestFindXAuditLogQueryAndAggregateUseRealAuditRows(t *testing.T) {
	r := logsTestRouter()
	scope := "logs-test-" + time.Now().Format("150405.000000")
	_, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
		ID:           "logs-audit-query-" + scope,
		CreatedAt:    time.Now().Add(-2 * time.Hour),
		Actor:        "codex",
		Action:       "log.query",
		ResourceType: "logs",
		ResourceID:   "audit-source",
		Scope:        scope,
		Status:       "ok",
		TraceID:      "trace-" + scope,
		Summary:      "FindX audit source token=<TOKEN> dsn=mysql://user:pass@host/db",
		Details: map[string]any{
			"token": "<TOKEN>",
			"url":   "https://example.test/audit?api_key=<API_KEY>",
		},
	})
	if err != nil {
		t.Fatalf("seed audit log: %v", err)
	}

	query := performLogsRequest(t, r, http.MethodGet, "/logs?source=findx_audit&scope="+scope+"&query=log.query&limit=10", nil)
	if query.Code != http.StatusOK {
		t.Fatalf("findx audit query should be 200, got %d body=%s", query.Code, query.Body.String())
	}
	var queryResp model.LogQueryResponse
	decodeLogsResponse(t, query, &queryResp)
	if queryResp.Source != model.LogsSourceFindXAudit || queryResp.SourceName != "FindX 审计日志" || len(queryResp.Items) != 1 {
		t.Fatalf("unexpected findx audit query response: %+v", queryResp)
	}
	if queryResp.Items[0].ServiceName != "findx-audit" || queryResp.Items[0].Attributes["action"] != "log.query" {
		t.Fatalf("audit row was not mapped into log shape: %+v", queryResp.Items[0])
	}
	assertNoLogSensitiveLeak(t, query.Body.String())

	aggregate := performLogsRequest(t, r, http.MethodGet, "/logs/aggregate?source=findx_audit&scope="+scope+"&group_by=status", nil)
	if aggregate.Code != http.StatusOK {
		t.Fatalf("findx audit aggregate should be 200, got %d body=%s", aggregate.Code, aggregate.Body.String())
	}
	var aggregateResp model.LogAggregateResponse
	decodeLogsResponse(t, aggregate, &aggregateResp)
	if aggregateResp.Source != model.LogsSourceFindXAudit || len(aggregateResp.Buckets) != 1 || aggregateResp.Buckets[0].Key != "ok" {
		t.Fatalf("unexpected aggregate response: %+v", aggregateResp)
	}
	assertNoLogSensitiveLeak(t, aggregate.Body.String())
}

func TestFindXAuditLogContextUsesRealAuditRows(t *testing.T) {
	r := logsTestRouter()
	scope := "logs-context-" + time.Now().Format("150405.000000")
	traceID := "trace-" + scope
	base := time.Now().Add(-2 * time.Hour)
	rows := []model.MonitorAuditLog{
		{ID: "ctx-new-" + scope, CreatedAt: base.Add(2 * time.Minute), Actor: "codex", Action: "newer", ResourceType: "logs", Scope: scope, Status: "ok", TraceID: traceID, Summary: "newer context"},
		{ID: "ctx-center-" + scope, CreatedAt: base.Add(time.Minute), Actor: "codex", Action: "center", ResourceType: "logs", Scope: scope, Status: "ok", TraceID: traceID, Summary: "center context"},
		{ID: "ctx-old-" + scope, CreatedAt: base, Actor: "codex", Action: "older", ResourceType: "logs", Scope: scope, Status: "ok", TraceID: traceID, Summary: "older context"},
	}
	for _, row := range rows {
		if _, err := store.AddMonitorAuditLog(row); err != nil {
			t.Fatalf("seed context row: %v", err)
		}
	}

	w := performLogsRequest(t, r, http.MethodGet, "/logs/context?source=findx_audit&log_id=ctx-center-"+scope+"&before=2&after=2", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("findx audit context should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp model.LogContextResponse
	decodeLogsResponse(t, w, &resp)
	if resp.Center == nil || resp.Center.ID != "ctx-center-"+scope || len(resp.Before) != 1 || len(resp.After) != 1 {
		t.Fatalf("unexpected context response: %+v", resp)
	}
	assertNoLogSensitiveLeak(t, w.Body.String())
}

func TestFindXAuditLogResponsesRedactBearerColonAndQuerySecrets(t *testing.T) {
	r := logsTestRouter()
	scope := "logs-redact-" + time.Now().Format("150405.000000")
	_, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
		ID:           "logs-redact-" + scope,
		CreatedAt:    time.Now().Add(-3 * time.Hour),
		Actor:        "codex",
		Action:       "log.redact",
		ResourceType: "logs",
		ResourceID:   "redact-source",
		Scope:        scope,
		Status:       "ok",
		TraceID:      "trace-" + scope,
		Summary:      "Authorization: Bearer <TOKEN>",
		Details: map[string]any{
			"body":    "Bearer <TOKEN>",
			"headers": "api-key: <API_KEY> cookie: <COOKIE>",
			"dsn":     "<DB_DSN>",
			"nested": map[string]any{
				"password_hint": "password: <PASSWORD>",
				"url":           "https://example.test/path?token=<TOKEN>&safe=1",
			},
		},
	})
	if err != nil {
		t.Fatalf("seed redaction audit log: %v", err)
	}

	w := performLogsRequest(t, r, http.MethodGet, "/logs?source=findx_audit&scope="+scope+"&limit=10", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("findx audit query should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	assertNoLogSensitiveLeak(t, body)
	for _, forbidden := range []string{"Authorization:", "Bearer ", "api-key:", "cookie:", "password:", "token="} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("logs response retained sensitive format %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, "REDACTED") {
		t.Fatalf("logs response should include generic redaction marker: %s", body)
	}
}

func TestFindXAuditLogContextCrowdedWindowKeepsCenter(t *testing.T) {
	r := logsTestRouter()
	scope := "logs-crowded-" + time.Now().Format("150405.000000")
	traceID := "trace-" + scope
	base := time.Now().Add(-4 * time.Hour)
	centerID := "ctx-crowded-center-" + scope

	if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{ID: centerID, CreatedAt: base, Actor: "codex", Action: "center", ResourceType: "logs", Scope: scope, Status: "ok", TraceID: traceID, Summary: "center"}); err != nil {
		t.Fatalf("seed crowded center: %v", err)
	}
	for i := 0; i < 125; i++ {
		if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
			ID:           "ctx-crowded-newer-" + strconv.Itoa(i) + "-" + scope,
			CreatedAt:    base.Add(time.Duration(i+1) * time.Minute),
			Actor:        "codex",
			Action:       "newer",
			ResourceType: "logs",
			Scope:        scope,
			Status:       "ok",
			TraceID:      traceID,
			Summary:      "newer",
		}); err != nil {
			t.Fatalf("seed crowded newer row: %v", err)
		}
	}
	for i := 0; i < 3; i++ {
		if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
			ID:           "ctx-crowded-older-" + strconv.Itoa(i) + "-" + scope,
			CreatedAt:    base.Add(-time.Duration(i+1) * time.Minute),
			Actor:        "codex",
			Action:       "older",
			ResourceType: "logs",
			Scope:        scope,
			Status:       "ok",
			TraceID:      traceID,
			Summary:      "older",
		}); err != nil {
			t.Fatalf("seed crowded older row: %v", err)
		}
	}

	w := performLogsRequest(t, r, http.MethodGet, "/logs/context?source=findx_audit&log_id="+centerID+"&before=2&after=2", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("findx crowded context should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp model.LogContextResponse
	decodeLogsResponse(t, w, &resp)
	if resp.Center == nil || resp.Center.ID != centerID {
		t.Fatalf("crowded context should keep center, got %+v", resp.Center)
	}
	if len(resp.Before) != 2 || len(resp.After) != 2 || len(resp.Items) != 5 {
		t.Fatalf("crowded context window mismatch: before=%d after=%d items=%d", len(resp.Before), len(resp.After), len(resp.Items))
	}
	if resp.Items[2].ID != centerID {
		t.Fatalf("crowded context should order before + center + after, got %+v", resp.Items)
	}
	assertNoLogSensitiveLeak(t, w.Body.String())
}

func TestLogQueryEndpointsBlockedByContract(t *testing.T) {
	r := logsTestRouter()
	for _, path := range []string{"/logs", "/logs?source=otel", "/logs/aggregate", "/logs/aggregate?source=otel", "/logs/context?source=otel&trace_id=x", "/logs/realtime", "/logs/stream"} {
		w := performLogsRequest(t, r, http.MethodGet, path, nil)
		if w.Code != http.StatusConflict && w.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s should be blocked 409/503, got %d body=%s", path, w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), model.LogsContractBlocked) {
			t.Fatalf("%s missing contract blocker: %s", path, w.Body.String())
		}
		assertNoLogSensitiveLeak(t, w.Body.String())
	}

	empty := performLogsRequest(t, r, http.MethodGet, "/logs/context?source=findx_audit", nil)
	if empty.Code != http.StatusBadRequest {
		t.Fatalf("context without target should be 400, got %d body=%s", empty.Code, empty.Body.String())
	}
}

func logsTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/logs/fields", ListLogFields)
	r.GET("/logs/pipelines/:version", ListLogPipelines)
	r.POST("/logs/pipelines", SaveLogPipeline)
	r.POST("/logs/pipelines/preview", PreviewLogPipeline)
	r.GET("/logs", ListLogsBlocked)
	r.GET("/logs/aggregate", AggregateLogsBlocked)
	r.GET("/logs/context", GetLogContext)
	r.GET("/logs/realtime", RealtimeLogsBlocked)
	r.GET("/logs/stream", RealtimeLogsBlocked)
	r.GET("/explorer/views", ListExplorerViews)
	r.POST("/explorer/views", CreateExplorerView)
	r.GET("/explorer/views/:id", GetExplorerView)
	r.PUT("/explorer/views/:id", UpdateExplorerView)
	r.DELETE("/explorer/views/:id", DeleteExplorerView)
	return r
}

func performLogsRequest(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal logs request: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeLogsResponse(t *testing.T, w *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
		t.Fatalf("decode logs response: %v body=%s", err, w.Body.String())
	}
}

func assertNoLogSensitiveLeak(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{"<token>", "<api_key>", "<password>", "<cookie>", "<db_dsn>", "mysql://", "password="} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("logs response leaked sensitive marker %q: %s", forbidden, body)
		}
	}
}
