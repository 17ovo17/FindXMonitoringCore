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

func TestMonitoringBuiltinRoutesRequireAuth(t *testing.T) {
	r := monitoringBuiltinTestRouter()
	resp := performBuiltinRequest(r, http.MethodGet, "/monitor/builtin-components", "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("builtin components without auth should be 401, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestMonitoringBuiltinListAndFilters(t *testing.T) {
	r := monitoringBuiltinTestRouter()
	seedDashboardToken("builtin-user-token", "u1", "alice", "user")

	componentsResp := performBuiltinRequest(r, http.MethodGet, "/monitor/builtin-components", "builtin-user-token")
	if componentsResp.Code != http.StatusOK {
		t.Fatalf("builtin components should be 200, got %d body=%s", componentsResp.Code, componentsResp.Body.String())
	}
	var components []model.MonitoringBuiltinComponent
	if err := json.Unmarshal(componentsResp.Body.Bytes(), &components); err != nil {
		t.Fatalf("decode components: %v", err)
	}
	if len(components) == 0 || components[0].ID == "" || components[0].Ident == "" || components[0].Name == "" || components[0].DashboardCount == 0 {
		t.Fatalf("components should include normalized fields and dashboard count: %+v", components)
	}

	payloadsResp := performBuiltinRequest(r, http.MethodGet, "/monitor/builtin-payloads?component_id=findx-monitor-core&type=dashboard&query=linux", "builtin-user-token")
	if payloadsResp.Code != http.StatusOK {
		t.Fatalf("filtered payload list should be 200, got %d body=%s", payloadsResp.Code, payloadsResp.Body.String())
	}
	payloads := decodeBuiltinPayloads(t, payloadsResp)
	if len(payloads) == 0 {
		t.Fatalf("filtered dashboard payloads should not be empty")
	}
	for _, payload := range payloads {
		if payload.ComponentID != "findx-monitor-core" || payload.Type != "dashboard" {
			t.Fatalf("payload filter returned unexpected item: %+v", payload)
		}
	}
}

func TestMonitoringBuiltinPayloadCatesAndDetail(t *testing.T) {
	r := monitoringBuiltinTestRouter()
	seedDashboardToken("builtin-detail-token", "u1", "alice", "user")

	catesResp := performBuiltinRequest(r, http.MethodGet, "/monitor/builtin-payloads/cates", "builtin-detail-token")
	if catesResp.Code != http.StatusOK {
		t.Fatalf("payload cates should be 200, got %d body=%s", catesResp.Code, catesResp.Body.String())
	}
	var cates []string
	if err := json.Unmarshal(catesResp.Body.Bytes(), &cates); err != nil {
		t.Fatalf("decode cates: %v", err)
	}
	if !containsBuiltinString(cates, "dashboard") || !containsBuiltinString(cates, "alert") || !containsBuiltinString(cates, "collect") {
		t.Fatalf("cates should include dashboard, alert, collect: %+v", cates)
	}

	detailResp := performBuiltinRequest(r, http.MethodGet, "/monitor/builtin-payload/dashboard:linux-host-basic", "builtin-detail-token")
	if detailResp.Code != http.StatusOK {
		t.Fatalf("payload detail should be 200, got %d body=%s", detailResp.Code, detailResp.Body.String())
	}
	detail := decodeBuiltinPayload(t, detailResp)
	if detail.ID != "dashboard:linux-host-basic" || detail.Content == nil {
		t.Fatalf("detail should include content and base fields: %+v", detail)
	}
	var content map[string]any
	if err := json.Unmarshal(detail.Content, &content); err != nil {
		t.Fatalf("decode dashboard payload content: %v content=%s", err, string(detail.Content))
	}
	if _, ok := content["variables"]; !ok {
		t.Fatalf("dashboard payload content should include variables: %+v", content)
	}
	if template, ok := content["template"].(string); !ok || template == "" {
		t.Fatalf("dashboard payload content should include template id: %+v", content)
	}
	if panels, ok := content["panels"].([]any); !ok || len(panels) == 0 {
		t.Fatalf("dashboard payload content should include panels: %+v", content)
	}
}

func TestMonitoringBuiltinCollectPayloadUsesFindXAgentNaming(t *testing.T) {
	r := monitoringBuiltinTestRouter()
	seedDashboardToken("builtin-collect-token", "u1", "alice", "user")

	resp := performBuiltinRequest(r, http.MethodGet, "/monitor/builtin-payloads?component_id=findx-monitor-core&type=collect", "builtin-collect-token")
	if resp.Code != http.StatusOK {
		t.Fatalf("collect payload list should be 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	body := strings.ToLower(resp.Body.String())
	sourceBrand := "cate" + "graf"
	if strings.Contains(body, sourceBrand) {
		t.Fatalf("collect payload should use FindX Agent naming and not expose upstream collector brand: %s", resp.Body.String())
	}
	payloads := decodeBuiltinPayloads(t, resp)
	if len(payloads) == 0 {
		t.Fatalf("collect payload list should not be empty")
	}
	if !strings.Contains(strings.ToLower(string(payloads[0].Content)), "findx_agent_collector_plugin") {
		t.Fatalf("collect payload content should identify FindX Agent collector plugin: %s", string(payloads[0].Content))
	}
}

func TestMonitoringBuiltinPayloadNotFoundAndWritesNotRegistered(t *testing.T) {
	r := monitoringBuiltinTestRouter()
	seedDashboardToken("builtin-missing-token", "u1", "alice", "user")

	missingResp := performBuiltinRequest(r, http.MethodGet, "/monitor/builtin-payload/not-found", "builtin-missing-token")
	if missingResp.Code != http.StatusNotFound {
		t.Fatalf("missing builtin payload should be 404, got %d body=%s", missingResp.Code, missingResp.Body.String())
	}
}

func TestMonitoringBuiltinUserWriteDenied(t *testing.T) {
	r := monitoringBuiltinTestRouter()
	seedDashboardToken("builtin-user-write-token", "u1", "alice", "user")

	resp := performBuiltinJSONRequest(r, http.MethodPost, "/monitor/builtin-components", "builtin-user-write-token", []map[string]any{validBuiltinComponentInput("fx-user-denied")})
	if resp.Code != http.StatusForbidden {
		t.Fatalf("normal user write should be 403, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestMonitoringBuiltinComponentAdminCreateUpdateDeleteAudited(t *testing.T) {
	r := monitoringBuiltinTestRouter()
	seedDashboardToken("builtin-admin-component-token", "a1", "root", "admin")
	id := "fx-test-builtin-component"

	createResp := performBuiltinJSONRequest(r, http.MethodPost, "/monitor/builtin-components", "builtin-admin-component-token", []map[string]any{validBuiltinComponentInput(id)}, map[string]string{"X-Test-Batch-Id": "fx-builtin-component-create"})
	if createResp.Code != http.StatusOK {
		t.Fatalf("admin create component should be 200, got %d body=%s", createResp.Code, createResp.Body.String())
	}
	var created []model.MonitoringBuiltinComponent
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil || len(created) != 1 {
		t.Fatalf("decode created component: %v body=%s", err, createResp.Body.String())
	}
	if created[0].ID != id || created[0].UpdatedBy != "root" || created[0].Disabled != 0 {
		t.Fatalf("created component not normalized: %+v", created[0])
	}

	update := validBuiltinComponentInput(id)
	update["ident"] = "fx-test-builtin-component-updated"
	update["name"] = "FindX Builtin Updated"
	update["readme"] = "Updated README"
	updateResp := performBuiltinJSONRequest(r, http.MethodPut, "/monitor/builtin-components", "builtin-admin-component-token", update, map[string]string{"X-Test-Batch-Id": "fx-builtin-component-update"})
	if updateResp.Code != http.StatusOK {
		t.Fatalf("admin update component should be 200, got %d body=%s", updateResp.Code, updateResp.Body.String())
	}
	updated := decodeBuiltinComponent(t, updateResp)
	if updated.Ident != "fx-test-builtin-component-updated" || updated.Readme != "Updated README" {
		t.Fatalf("component update not readable: %+v", updated)
	}

	deleteResp := performBuiltinJSONRequest(r, http.MethodDelete, "/monitor/builtin-components", "builtin-admin-component-token", map[string]any{"ids": []string{id}}, map[string]string{"X-Test-Batch-Id": "fx-builtin-component-delete"})
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("admin delete component should be 200, got %d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	if componentListed(id) {
		t.Fatalf("deleted component %s should not be listed", id)
	}
	assertBuiltinAuditRecorded(t, "monitor.builtin.component.create", id)
	assertBuiltinAuditRecorded(t, "monitor.builtin.component.update", id)
	assertBuiltinAuditRecorded(t, "monitor.builtin.component.delete", id)
}

func TestMonitoringBuiltinPayloadAdminCreateUpdateDeleteAudited(t *testing.T) {
	r := monitoringBuiltinTestRouter()
	seedDashboardToken("builtin-admin-payload-token", "a1", "root", "admin")
	componentID := "fx-test-builtin-payload-component"
	payloadID := "fx-test-builtin-payload"
	performBuiltinJSONRequest(r, http.MethodPost, "/monitor/builtin-components", "builtin-admin-payload-token", []map[string]any{validBuiltinComponentInput(componentID)})

	createResp := performBuiltinJSONRequest(r, http.MethodPost, "/monitor/builtin-payloads", "builtin-admin-payload-token", []map[string]any{validBuiltinPayloadInput(payloadID, componentID)}, map[string]string{"X-Test-Batch-Id": "fx-builtin-payload-create"})
	if createResp.Code != http.StatusOK {
		t.Fatalf("admin create payload should be 200, got %d body=%s", createResp.Code, createResp.Body.String())
	}
	created := decodeBuiltinPayloads(t, createResp)
	if len(created) != 1 || created[0].ID != payloadID || created[0].UpdatedBy != "root" || !json.Valid(created[0].Content) {
		t.Fatalf("created payload not normalized: %+v", created)
	}

	update := validBuiltinPayloadInput(payloadID, componentID)
	update["name"] = "FindX Payload Updated"
	update["content"] = map[string]any{"expr": "up == 0", "severity": "critical"}
	updateResp := performBuiltinJSONRequest(r, http.MethodPut, "/monitor/builtin-payloads", "builtin-admin-payload-token", update, map[string]string{"X-Test-Batch-Id": "fx-builtin-payload-update"})
	if updateResp.Code != http.StatusOK {
		t.Fatalf("admin update payload should be 200, got %d body=%s", updateResp.Code, updateResp.Body.String())
	}
	updated := decodeBuiltinPayload(t, updateResp)
	if updated.Name != "FindX Payload Updated" {
		t.Fatalf("payload update not readable: %+v", updated)
	}

	deleteResp := performBuiltinJSONRequest(r, http.MethodDelete, "/monitor/builtin-payloads", "builtin-admin-payload-token", map[string]any{"ids": []string{payloadID}}, map[string]string{"X-Test-Batch-Id": "fx-builtin-payload-delete"})
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("admin delete payload should be 200, got %d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	missingResp := performBuiltinRequest(r, http.MethodGet, "/monitor/builtin-payload/"+payloadID, "builtin-admin-payload-token")
	if missingResp.Code != http.StatusNotFound {
		t.Fatalf("deleted payload should be missing, got %d body=%s", missingResp.Code, missingResp.Body.String())
	}
	assertBuiltinAuditRecorded(t, "monitor.builtin.payload.create", payloadID)
	assertBuiltinAuditRecorded(t, "monitor.builtin.payload.update", payloadID)
	assertBuiltinAuditRecorded(t, "monitor.builtin.payload.delete", payloadID)
}

func TestMonitoringBuiltinValidationAndProtection(t *testing.T) {
	r := monitoringBuiltinTestRouter()
	seedDashboardToken("builtin-validation-admin-token", "a1", "root", "admin")

	invalidComponent := validBuiltinComponentInput("fx-invalid-component")
	invalidComponent["logo"] = "https://example.invalid/logo.png"
	assertBuiltinStatus(t, r, http.MethodPost, "/monitor/builtin-components", "builtin-validation-admin-token", []map[string]any{invalidComponent}, http.StatusBadRequest)
	protectedComponent := validBuiltinComponentInput("findx-monitor-core")
	protectedComponent["ident"] = "findx-monitor-core"
	assertBuiltinStatus(t, r, http.MethodPost, "/monitor/builtin-components", "builtin-validation-admin-token", []map[string]any{protectedComponent}, http.StatusConflict)
	assertBuiltinStatus(t, r, http.MethodDelete, "/monitor/builtin-components", "builtin-validation-admin-token", map[string]any{"ids": []string{"findx-monitor-core"}}, http.StatusConflict)

	componentID := "fx-test-builtin-protected-component"
	payloadID := "fx-test-builtin-protected-payload"
	assertBuiltinStatus(t, r, http.MethodPost, "/monitor/builtin-components", "builtin-validation-admin-token", []map[string]any{validBuiltinComponentInput(componentID)}, http.StatusOK)
	assertBuiltinStatus(t, r, http.MethodPost, "/monitor/builtin-payloads", "builtin-validation-admin-token", []map[string]any{validBuiltinPayloadInput(payloadID, componentID)}, http.StatusOK)
	assertBuiltinStatus(t, r, http.MethodDelete, "/monitor/builtin-components", "builtin-validation-admin-token", map[string]any{"ids": []string{componentID}}, http.StatusConflict)

	invalidPayload := validBuiltinPayloadInput("fx-invalid-payload", componentID)
	invalidPayload["type"] = "metric"
	assertBuiltinStatus(t, r, http.MethodPost, "/monitor/builtin-payloads", "builtin-validation-admin-token", []map[string]any{invalidPayload}, http.StatusBadRequest)
	secretPayload := validBuiltinPayloadInput("fx-secret-payload", componentID)
	secretPayload["content"] = map[string]any{"password": "not-allowed"}
	assertBuiltinStatus(t, r, http.MethodPost, "/monitor/builtin-payloads", "builtin-validation-admin-token", []map[string]any{secretPayload}, http.StatusBadRequest)
	assertBuiltinStatus(t, r, http.MethodDelete, "/monitor/builtin-payloads", "builtin-validation-admin-token", map[string]any{"ids": []string{"dashboard:linux-host-basic"}}, http.StatusConflict)
}

func TestMonitoringBuiltinPermissionMatrix(t *testing.T) {
	checker := model.NewMonitorPermissionChecker()
	read := checker.Check("user", "monitor.builtin", "read")
	if !read.Known || !read.Allowed {
		t.Fatalf("user should have monitor.builtin read permission: %+v", read)
	}
	create := checker.Check("user", "monitor.builtin", "create")
	if !create.Known || create.Allowed || create.Reason != model.MonitorPermissionReasonRoleDenied {
		t.Fatalf("user should be denied monitor.builtin create: %+v", create)
	}
	adminCreate := checker.Check("admin", "monitor.builtin", "create")
	if !adminCreate.Known || !adminCreate.Allowed {
		t.Fatalf("admin should have monitor.builtin create: %+v", adminCreate)
	}
}

func monitoringBuiltinTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	clearDashboardTokens()
	r := gin.New()
	r.GET("/monitor/builtin-components", RequireMonitorPermission("monitor.builtin", "read"), ListMonitoringBuiltinComponents)
	r.POST("/monitor/builtin-components", RequireMonitorPermission("monitor.builtin", "create"), CreateMonitoringBuiltinComponents)
	r.PUT("/monitor/builtin-components", RequireMonitorPermission("monitor.builtin", "update"), UpdateMonitoringBuiltinComponent)
	r.DELETE("/monitor/builtin-components", RequireMonitorPermission("monitor.builtin", "delete"), DeleteMonitoringBuiltinComponents)
	r.GET("/monitor/builtin-payloads/cates", RequireMonitorPermission("monitor.builtin", "read"), ListMonitoringBuiltinPayloadTypes)
	r.GET("/monitor/builtin-payloads", RequireMonitorPermission("monitor.builtin", "read"), ListMonitoringBuiltinPayloads)
	r.POST("/monitor/builtin-payloads", RequireMonitorPermission("monitor.builtin", "create"), CreateMonitoringBuiltinPayloads)
	r.PUT("/monitor/builtin-payloads", RequireMonitorPermission("monitor.builtin", "update"), UpdateMonitoringBuiltinPayload)
	r.DELETE("/monitor/builtin-payloads", RequireMonitorPermission("monitor.builtin", "delete"), DeleteMonitoringBuiltinPayloads)
	r.GET("/monitor/builtin-payload/:id", RequireMonitorPermission("monitor.builtin", "read"), GetMonitoringBuiltinPayload)
	return r
}

func performBuiltinRequest(r *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performBuiltinJSONRequest(r *gin.Engine, method, path, token string, body any, headers ...map[string]string) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeBuiltinPayloads(t *testing.T, w *httptest.ResponseRecorder) []model.MonitoringBuiltinPayload {
	t.Helper()
	var items []model.MonitoringBuiltinPayload
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("decode builtin payloads: %v body=%s", err, w.Body.String())
	}
	return items
}

func decodeBuiltinPayload(t *testing.T, w *httptest.ResponseRecorder) model.MonitoringBuiltinPayload {
	t.Helper()
	var item model.MonitoringBuiltinPayload
	if err := json.Unmarshal(w.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode builtin payload: %v body=%s", err, w.Body.String())
	}
	return item
}

func decodeBuiltinComponent(t *testing.T, w *httptest.ResponseRecorder) model.MonitoringBuiltinComponent {
	t.Helper()
	var item model.MonitoringBuiltinComponent
	if err := json.Unmarshal(w.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode builtin component: %v body=%s", err, w.Body.String())
	}
	return item
}

func validBuiltinComponentInput(id string) map[string]any {
	return map[string]any{
		"id":       id,
		"ident":    id,
		"name":     "FindX Builtin " + id,
		"logo":     "/image/logos/host.png",
		"readme":   "FindX builtin component for contract tests.",
		"disabled": 0,
	}
}

func validBuiltinPayloadInput(id, componentID string) map[string]any {
	return map[string]any{
		"id":           id,
		"uuid":         id,
		"type":         "alert",
		"component_id": componentID,
		"cate":         "host",
		"name":         "FindX Builtin Payload " + id,
		"content":      map[string]any{"expr": "up == 0", "severity": "warning"},
	}
}

func assertBuiltinStatus(t *testing.T, r *gin.Engine, method, path, token string, body any, want int) {
	t.Helper()
	resp := performBuiltinJSONRequest(r, method, path, token, body)
	if resp.Code != want {
		t.Fatalf("%s %s want %d got %d body=%s", method, path, want, resp.Code, resp.Body.String())
	}
	assertBuiltinResponseSanitized(t, resp.Body.String())
}

func assertBuiltinAuditRecorded(t *testing.T, action, target string) {
	t.Helper()
	for _, event := range store.ListAuditEvents(1000) {
		if event.Action == action && event.Target == target {
			body, _ := json.Marshal(event)
			assertBuiltinResponseSanitized(t, string(body))
			return
		}
	}
	t.Fatalf("audit event %s target %s not found", action, target)
}

func assertBuiltinResponseSanitized(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{"nightingale", "n9e", "categraf", "catpaw", "skywalking", "signoz", "findx.local", "token=", "password=", "api_key", "apikey", "secret=", "cookie:", "authorization:", "private key"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("builtin response leaked forbidden value %q: %s", forbidden, body)
		}
	}
}
