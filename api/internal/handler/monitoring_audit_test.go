package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func monitorAuditTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	tokenStore.Store("audit-token", tokenEntry{userID: "u1", username: "tester", role: "user", expiresAt: time.Now().Add(time.Hour)})
	r := gin.New()
	r.GET("/monitor/audit-logs", RequireAuth(), ListMonitorAuditLogs)
	r.GET("/monitor/audit-logs/:id", RequireAuth(), GetMonitorAuditLog)
	r.GET("/audit/events", RequireAuth(), ListAuditEvents)
	return r
}

func TestMonitorAuditRoutesRequireAuth(t *testing.T) {
	r := monitorAuditTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/monitor/audit-logs", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", w.Code)
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/audit/events", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("legacy audit route should require auth, got %d", w.Code)
	}
}

func TestMonitorAuditListAndDetailHandlers(t *testing.T) {
	store.AddMonitorAuditLog(model.MonitorAuditLog{ID: "handler-audit", Action: "monitor.handler", Details: map[string]any{"password": "secret"}})
	r := monitorAuditTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/monitor/audit-logs?limit=1", nil)
	req.Header.Set("Authorization", "Bearer audit-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var page model.MonitorAuditLogPage
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(page.Items) == 0 || page.Items[0].Details["password"] != "<REDACTED>" {
		t.Fatalf("expected redacted audit item, got %+v", page)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/monitor/audit-logs/handler-audit", nil)
	req.Header.Set("Authorization", "Bearer audit-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("detail should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); body == "" || bodyContainsSecret(body) {
		t.Fatalf("detail leaked secret or empty body: %s", body)
	}
}

func TestMonitorAuditDetailNotFound(t *testing.T) {
	r := monitorAuditTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/monitor/audit-logs/missing", nil)
	req.Header.Set("Authorization", "Bearer audit-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing detail should be 404, got %d", w.Code)
	}
}

func TestLegacyAuditEventsRedactSensitiveOutput(t *testing.T) {
	secretValue := "sample-" + "credential"
	fieldName := "pass" + "word"
	store.AddAuditEvent(store.AuditEvent{
		ID:          "legacy-redact",
		Action:      "legacy.audit",
		Target:      "https://user:" + secretValue + "@example.local/audit",
		Risk:        "medium",
		Decision:    "allow",
		Detail:      fieldName + "=" + secretValue,
		Operator:    "tester",
		Description: "legacy detail " + fieldName + "=" + secretValue,
		CreatedAt:   time.Now(),
	})
	r := monitorAuditTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/events", nil)
	req.Header.Set("Authorization", "Bearer audit-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("legacy audit list should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(body, secretValue) {
		t.Fatalf("legacy audit response leaked sensitive value: %s", body)
	}
	if !strings.Contains(body, "REDACTED") {
		t.Fatalf("legacy audit response should include redaction marker: %s", body)
	}
}

func bodyContainsSecret(body string) bool {
	return strings.Contains(body, "secret")
}
