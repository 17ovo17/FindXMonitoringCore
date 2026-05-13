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

func TestProbeCheckCRUDAndUnknownStatusPage(t *testing.T) {
	store.ResetProbeStoreForTest()
	r := probeTestRouter()

	create := performProbeRequest(t, r, http.MethodPost, "/probes/checks", map[string]any{
		"name":             "API 网关",
		"type":             "http",
		"url":              "https://status.example.local/health",
		"interval_seconds": 60,
		"timeout_seconds":  5,
		"enabled":          true,
		"labels": map[string]string{
			"owner": "sre",
			"token": "<TOKEN>",
		},
	})
	if create.Code != http.StatusOK {
		t.Fatalf("create probe check should be 200, got %d body=%s", create.Code, create.Body.String())
	}
	var check model.ProbeCheck
	decodeProbeResponse(t, create, &check)
	if check.ID == "" || check.Status != model.ProbeStatusUnknown || check.Labels["token"] != "" {
		t.Fatalf("created check should be unknown and sanitized: %+v", check)
	}

	list := performProbeRequest(t, r, http.MethodGet, "/probes/checks", nil)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), check.ID) {
		t.Fatalf("list should include check, code=%d body=%s", list.Code, list.Body.String())
	}

	status := performProbeRequest(t, r, http.MethodGet, "/probes/status-pages/main", nil)
	if status.Code != http.StatusOK {
		t.Fatalf("status page should be 200, got %d body=%s", status.Code, status.Body.String())
	}
	var view model.ProbeStatusPageView
	decodeProbeResponse(t, status, &view)
	if view.Status != model.ProbeStatusUnknown || view.Summary.HasRunEvidence || view.Summary.Uptime90d != nil {
		t.Fatalf("status page without runs must be unknown/no-data: %+v", view)
	}
	if strings.Contains(strings.ToLower(status.Body.String()), "all operational") || strings.Contains(status.Body.String(), "99.9") {
		t.Fatalf("status page must not fake green uptime: %s", status.Body.String())
	}
}

func TestProbeExecutionAndBindingsExposeBlockedContract(t *testing.T) {
	store.ResetProbeStoreForTest()
	r := probeTestRouter()
	create := performProbeRequest(t, r, http.MethodPost, "/probes/checks", map[string]any{
		"name": "DB TCP", "type": "tcp", "target": "db.internal", "port": 3306,
		"interval_seconds": 60, "timeout_seconds": 5, "enabled": true,
	})
	var check model.ProbeCheck
	decodeProbeResponse(t, create, &check)

	testRun := performProbeRequest(t, r, http.MethodPost, "/probes/checks/"+check.ID+"/test", map[string]any{"password": "<PASSWORD>"})
	if testRun.Code != http.StatusConflict {
		t.Fatalf("test run should be blocked 409, got %d body=%s", testRun.Code, testRun.Body.String())
	}
	var blocked model.ProbeBlockedResponse
	decodeProbeResponse(t, testRun, &blocked)
	if blocked.Code != model.ProbeContractBlockedCode || blocked.ContractGapID != "FX-CONTRACT-BUSINESS-PROBE-EXECUTOR" ||
		blocked.SafeToRetry || len(blocked.MissingContracts) == 0 {
		t.Fatalf("blocked envelope mismatch: %+v", blocked)
	}
	assertProbeNoSensitiveOrFakeSuccess(t, testRun.Body.String())

	saveNotify := performProbeRequest(t, r, http.MethodPut, "/probes/checks/"+check.ID+"/notification-bindings", map[string]any{
		"items": []map[string]any{{"channel_id": "notify-console", "enabled": true, "labels": map[string]string{"cookie": "<COOKIE>", "team": "sre"}}},
	})
	if saveNotify.Code != http.StatusOK {
		t.Fatalf("save notification binding should be 200, got %d body=%s", saveNotify.Code, saveNotify.Body.String())
	}
	if !strings.Contains(saveNotify.Body.String(), "blocked_by_contract") || strings.Contains(saveNotify.Body.String(), "<COOKIE>") {
		t.Fatalf("notification binding must show blocked receipt and sanitize labels: %s", saveNotify.Body.String())
	}

	saveAlert := performProbeRequest(t, r, http.MethodPut, "/probes/checks/"+check.ID+"/alert-bindings", map[string]any{
		"items": []map[string]any{{"alert_rule_id": "rule-1", "enabled": true}},
	})
	if saveAlert.Code != http.StatusOK || !strings.Contains(saveAlert.Body.String(), "FX-CONTRACT-BUSINESS-PROBE-ALERT-LIFECYCLE") {
		t.Fatalf("alert binding should save but expose blocked lifecycle: code=%d body=%s", saveAlert.Code, saveAlert.Body.String())
	}
}

func TestProbeIncidentLifecycleAndValidation(t *testing.T) {
	store.ResetProbeStoreForTest()
	r := probeTestRouter()
	bad := performProbeRequest(t, r, http.MethodPost, "/probes/incidents", map[string]any{"title": "bad", "status": "success"})
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("invalid incident status should be 400, got %d body=%s", bad.Code, bad.Body.String())
	}
	create := performProbeRequest(t, r, http.MethodPost, "/probes/incidents", map[string]any{
		"title": "支付网关访问异常", "status": "investigating", "severity": "p1", "message": "人工确认中",
	})
	if create.Code != http.StatusOK {
		t.Fatalf("create incident should be 200, got %d body=%s", create.Code, create.Body.String())
	}
	var incident model.ProbeIncident
	decodeProbeResponse(t, create, &incident)
	update := performProbeRequest(t, r, http.MethodPut, "/probes/incidents/"+incident.ID, map[string]any{
		"title": "支付网关访问异常", "status": "resolved", "severity": "p1",
	})
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), "resolved") {
		t.Fatalf("update incident should resolve, code=%d body=%s", update.Code, update.Body.String())
	}
}

func probeTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/probes/status-pages/:slug", GetProbeStatusPage)
	r.GET("/probes/checks", ListProbeChecks)
	r.POST("/probes/checks", CreateProbeCheck)
	r.GET("/probes/checks/:id", GetProbeCheck)
	r.PUT("/probes/checks/:id", UpdateProbeCheck)
	r.DELETE("/probes/checks/:id", DeleteProbeCheck)
	r.POST("/probes/checks/:id/test", TestProbeCheckBlocked)
	r.PUT("/probes/checks/:id/notification-bindings", SaveProbeNotificationBindings)
	r.PUT("/probes/checks/:id/alert-bindings", SaveProbeAlertBindings)
	r.POST("/probes/incidents", CreateProbeIncident)
	r.PUT("/probes/incidents/:id", UpdateProbeIncident)
	return r
}

func performProbeRequest(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", "probe-test-admin")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeProbeResponse(t *testing.T, w *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
		t.Fatalf("decode response failed: %v body=%s", err, w.Body.String())
	}
}

func assertProbeNoSensitiveOrFakeSuccess(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, marker := range []string{"<password>", "<cookie>", "password", "cookie", `"success":true`, `"status":"success"`, `"status":"succeeded"`, `"status":"running"`, `"status":"queued"`} {
		if strings.Contains(lower, marker) {
			t.Fatalf("probe response leaked sensitive or fake state marker %q: %s", marker, body)
		}
	}
}
