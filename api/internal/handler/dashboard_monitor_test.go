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

	"github.com/gin-gonic/gin"
)

func TestMonitorDashboardRoutesRequireAuthAndPermission(t *testing.T) {
	r := monitorDashboardTestRouter()
	if w := performDashboardRequest(t, r, http.MethodGet, "/monitor/dashboards", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("GET /monitor/dashboards without auth should be 401, got %d", w.Code)
	}
	seedDashboardToken("dashboard-user-token", "u1", "alice", "user")
	w := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboards", "dashboard-user-token", validDashboardInput())
	if w.Code != http.StatusForbidden {
		t.Fatalf("POST /monitor/dashboards with user token should be 403, got %d body=%s", w.Code, w.Body.String())
	}
	if !model.NewMonitorPermissionChecker().Check("user", "monitor.dashboard", "read").Allowed {
		t.Fatalf("user should have monitor.dashboard read permission")
	}
	if model.NewMonitorPermissionChecker().Check("user", "monitor.dashboard", "create").Allowed {
		t.Fatalf("user must not have monitor.dashboard create permission")
	}
}

func TestAdminMonitorDashboardCRUDCloneShare(t *testing.T) {
	r := monitorDashboardTestRouter()
	seedDashboardToken("dashboard-admin-token", "a1", "root", "admin")

	createResp := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboards", "dashboard-admin-token", validDashboardInput())
	if createResp.Code != http.StatusOK {
		t.Fatalf("admin create should be 200, got %d body=%s", createResp.Code, createResp.Body.String())
	}
	created := decodeDashboard(t, createResp)
	if created.CreatedBy != "root" || created.UpdatedBy != "root" {
		t.Fatalf("server actor fields should come from auth context: %+v", created)
	}
	if created.Shared || created.ShareTokenSet || created.ShareSummary != "" {
		t.Fatalf("create must ignore client-controlled share fields: %+v", created)
	}

	update := validDashboardInput()
	update["title"] = "更新后的仪表盘"
	update["shared"] = true
	update["share_token_set"] = true
	update["share_summary"] = "token=<TOKEN>"
	updateResp := performDashboardRequest(t, r, http.MethodPut, "/monitor/dashboards/"+created.ID, "dashboard-admin-token", update)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("admin update should be 200, got %d body=%s", updateResp.Code, updateResp.Body.String())
	}
	updated := decodeDashboard(t, updateResp)
	if updated.Title != "更新后的仪表盘" || updated.Shared || updated.ShareTokenSet || updated.ShareSummary != "" {
		t.Fatalf("update should keep server-only share fields untouched: %+v", updated)
	}

	shareResp := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboards/"+created.ID+"/share", "dashboard-admin-token", nil)
	if shareResp.Code != http.StatusOK {
		t.Fatalf("admin share should be 200, got %d body=%s", shareResp.Code, shareResp.Body.String())
	}
	assertDashboardResponseSanitized(t, shareResp.Body.String())
	var share model.MonitorDashboardShareResult
	if err := json.Unmarshal(shareResp.Body.Bytes(), &share); err != nil {
		t.Fatalf("decode share response: %v", err)
	}
	if !share.ShareEnabled || share.ShareSummary != "仪表盘分享已启用" {
		t.Fatalf("unexpected share response: %+v", share)
	}
	assertDashboardShareResponseSanitized(t, shareResp.Body.String())

	time.Sleep(time.Millisecond)
	cloneResp := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboards/"+created.ID+"/clone", "dashboard-admin-token", nil)
	if cloneResp.Code != http.StatusOK {
		t.Fatalf("admin clone should be 200, got %d body=%s", cloneResp.Code, cloneResp.Body.String())
	}
	clone := decodeDashboard(t, cloneResp)
	if clone.ID == created.ID {
		t.Fatalf("clone should be a new dashboard id: created=%s clone=%s", created.ID, clone.ID)
	}
	if !strings.Contains(clone.Title, "副本") {
		t.Fatalf("clone title should contain 副本: %q", clone.Title)
	}
	if clone.Shared || clone.ShareTokenSet || clone.ShareSummary != "" {
		t.Fatalf("clone must not copy share state: %+v", clone)
	}
	getSource := performDashboardRequest(t, r, http.MethodGet, "/monitor/dashboards/"+created.ID, "dashboard-admin-token", nil)
	source := decodeDashboard(t, getSource)
	if !source.Shared || source.ShareSummary != "仪表盘分享已启用" {
		t.Fatalf("source dashboard share state should remain enabled: %+v", source)
	}

	deleteResp := performDashboardRequest(t, r, http.MethodDelete, "/monitor/dashboards/"+created.ID, "dashboard-admin-token", nil)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("admin delete should be 200, got %d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
}

func TestUpdateMissingMonitorDashboardReturnsNotFound(t *testing.T) {
	r := monitorDashboardTestRouter()
	seedDashboardToken("dashboard-missing-admin-token", "a1", "root", "admin")
	resp := performDashboardRequest(t, r, http.MethodPut, "/monitor/dashboards/dashboard-missing-id", "dashboard-missing-admin-token", validDashboardInput())
	if resp.Code != http.StatusNotFound {
		t.Fatalf("update missing dashboard should be 404, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestMonitorDashboardRejectsInvalidPanelsAndVariables(t *testing.T) {
	r := monitorDashboardTestRouter()
	seedDashboardToken("dashboard-invalid-admin-token", "a1", "root", "admin")
	badPanels := validDashboardInput()
	badPanels["panels"] = map[string]any{"not": "array"}
	if w := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboards", "dashboard-invalid-admin-token", badPanels); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid panels should be 400, got %d body=%s", w.Code, w.Body.String())
	}
	badVariables := validDashboardInput()
	badVariables["variables"] = []any{"not-object"}
	if w := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboards", "dashboard-invalid-admin-token", badVariables); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid variables should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestMonitorDashboardResponsesRedactSensitivePayload(t *testing.T) {
	r := monitorDashboardTestRouter()
	seedDashboardToken("dashboard-sanitize-admin-token", "a1", "root", "admin")
	input := validDashboardInput()
	input["variables"] = map[string]any{
		"api_key": "super-secret",
		"url":     "https://example.local/path?token=<TOKEN>",
	}
	input["panels"] = []any{map[string]any{
		"title":  "panel",
		"cookie": "<COOKIE>",
		"dsn":    "mysql://user:password@example.local/db",
	}}
	resp := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboards", "dashboard-sanitize-admin-token", input)
	if resp.Code != http.StatusOK {
		t.Fatalf("create sanitized dashboard should be 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	assertDashboardResponseSanitized(t, resp.Body.String())
}

func monitorDashboardTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	clearDashboardTokens()
	r := gin.New()
	r.GET("/monitor/dashboards", RequireMonitorPermission("monitor.dashboard", "read"), ListMonitorDashboards)
	r.POST("/monitor/dashboards", RequireMonitorPermission("monitor.dashboard", "create"), CreateMonitorDashboard)
	r.GET("/monitor/dashboards/:id", RequireMonitorPermission("monitor.dashboard", "read"), GetMonitorDashboard)
	r.PUT("/monitor/dashboards/:id", RequireMonitorPermission("monitor.dashboard", "update"), UpdateMonitorDashboard)
	r.DELETE("/monitor/dashboards/:id", RequireMonitorPermission("monitor.dashboard", "delete"), DeleteMonitorDashboard)
	r.POST("/monitor/dashboards/:id/clone", RequireMonitorPermission("monitor.dashboard", "clone"), CloneMonitorDashboard)
	r.POST("/monitor/dashboards/:id/share", RequireMonitorPermission("monitor.dashboard", "share"), ShareMonitorDashboard)
	return r
}

func performDashboardRequest(t *testing.T, r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal dashboard request: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func seedDashboardToken(token, userID, username, role string) {
	tokenStore.Store(token, tokenEntry{userID: userID, username: username, role: role, expiresAt: time.Now().Add(time.Hour)})
}

func clearDashboardTokens() {
	tokenStore.Range(func(key, _ any) bool {
		tokenStore.Delete(key)
		return true
	})
}

func validDashboardInput() map[string]any {
	return map[string]any{
		"title":             "核心监控仪表盘",
		"description":       "单元测试仪表盘",
		"workspace_id":      "ws-dashboard",
		"resource_group_id": "rg-dashboard",
		"tags":              []string{"prod", "prod", "ops"},
		"variables":         map[string]any{"env": "prod"},
		"panels":            []any{map[string]any{"title": "CPU", "query": "up"}},
		"status":            model.MonitorDashboardStatusActive,
		"shared":            true,
		"share_token_set":   true,
		"share_summary":     "token=<TOKEN>",
		"created_by":        "client",
		"updated_by":        "client",
	}
}

func decodeDashboard(t *testing.T, w *httptest.ResponseRecorder) model.MonitorDashboard {
	t.Helper()
	var item model.MonitorDashboard
	if err := json.Unmarshal(w.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode dashboard response: %v body=%s", err, w.Body.String())
	}
	return item
}

func assertDashboardResponseSanitized(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{"super-secret", "<token>", "<cookie>", "mysql://", "password@example"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("dashboard response leaked sensitive marker %q: %s", forbidden, body)
		}
	}
}

func assertDashboardShareResponseSanitized(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{"token", "hash", "cookie", "dsn", "://", "<token>", "<cookie>", "<db_dsn>"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("dashboard response leaked sensitive marker %q: %s", forbidden, body)
		}
	}
}
