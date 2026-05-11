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

func TestMonitoringSystemIntegrationRoutesRequireAuth(t *testing.T) {
	r := monitoringSystemIntegrationTestRouter()
	resp := performSystemIntegrationRequest(r, http.MethodGet, "/monitor/system-integrations", "", nil)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("system integrations without auth should be 401, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestMonitoringSystemIntegrationListAndDetail(t *testing.T) {
	r := monitoringSystemIntegrationTestRouter()
	seedDashboardToken("system-integration-user-token", "u1", "alice", "user")

	listResp := performSystemIntegrationRequest(r, http.MethodGet, "/monitor/system-integrations?visibility=public&status=active", "system-integration-user-token", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("system integration list should be 200, got %d body=%s", listResp.Code, listResp.Body.String())
	}
	items := decodeSystemIntegrations(t, listResp)
	if len(items) == 0 {
		t.Fatalf("filtered system integration list should not be empty")
	}
	for _, item := range items {
		if item.IsPrivate || item.Status != model.MonitoringSystemIntegrationStatusActive {
			t.Fatalf("filters returned unexpected item: %+v", item)
		}
		assertSystemIntegrationWriteContract(t, item)
	}

	detailResp := performSystemIntegrationRequest(r, http.MethodGet, "/monitor/system-integrations/"+items[0].ID, "system-integration-user-token", nil)
	if detailResp.Code != http.StatusOK {
		t.Fatalf("system integration detail should be 200, got %d body=%s", detailResp.Code, detailResp.Body.String())
	}
	detail := decodeSystemIntegration(t, detailResp)
	if detail.ID != items[0].ID || detail.Name == "" || detail.ConfigPreview == "" {
		t.Fatalf("detail should include base fields: %+v", detail)
	}
	assertSystemIntegrationWriteContract(t, detail)
	assertSystemIntegrationResponseSanitized(t, detailResp.Body.String())
}

func TestMonitoringSystemIntegrationQueryFilter(t *testing.T) {
	r := monitoringSystemIntegrationTestRouter()
	seedDashboardToken("system-integration-query-token", "u1", "alice", "user")

	resp := performSystemIntegrationRequest(r, http.MethodGet, "/monitor/system-integrations?query=knowledge", "system-integration-query-token", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("query filter should be 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	items := decodeSystemIntegrations(t, resp)
	if len(items) != 1 || items[0].ID != "findx-operations-knowledge" {
		t.Fatalf("query filter returned unexpected items: %+v", items)
	}
}

func TestMonitoringSystemIntegrationUserWriteDenied(t *testing.T) {
	r := monitoringSystemIntegrationTestRouter()
	seedDashboardToken("system-integration-readonly-user-token", "u1", "alice", "user")

	resp := performSystemIntegrationRequest(r, http.MethodPost, "/monitor/system-integrations", "system-integration-readonly-user-token", validSystemIntegrationInput("user-denied"))
	if resp.Code != http.StatusForbidden {
		t.Fatalf("normal user write should be 403, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestMonitoringSystemIntegrationAdminCreateUpdateDeletePersistedAndAudited(t *testing.T) {
	r := monitoringSystemIntegrationTestRouter()
	seedDashboardToken("system-integration-admin-token", "a1", "root", "admin")
	id := "fx-test-create-update-delete"

	createResp := performSystemIntegrationRequest(r, http.MethodPost, "/monitor/system-integrations", "system-integration-admin-token", validSystemIntegrationInput(id), map[string]string{"X-Test-Batch-Id": "fx-system-create"})
	if createResp.Code != http.StatusOK {
		t.Fatalf("admin create should be 200, got %d body=%s", createResp.Code, createResp.Body.String())
	}
	created := decodeSystemIntegration(t, createResp)
	if created.ID != id || created.CreateBy != "root" || created.UpdateBy != "root" || !created.ShowInMenu {
		t.Fatalf("created integration not persisted with actor/menu state: %+v", created)
	}
	assertSystemIntegrationResponseSanitized(t, createResp.Body.String())

	update := validSystemIntegrationInput(id)
	update["name"] = "FindX Updated Integration"
	update["url"] = "/platform?section=updated"
	update["config_preview"] = "/platform?section=updated"
	updateResp := performSystemIntegrationRequest(r, http.MethodPut, "/monitor/system-integrations/"+id, "system-integration-admin-token", update, map[string]string{"X-Test-Batch-Id": "fx-system-update"})
	if updateResp.Code != http.StatusOK {
		t.Fatalf("admin update should be 200, got %d body=%s", updateResp.Code, updateResp.Body.String())
	}
	updated := decodeSystemIntegration(t, updateResp)
	if updated.Name != "FindX Updated Integration" || updated.URL != "/platform?section=updated" {
		t.Fatalf("updated integration not readable: %+v", updated)
	}

	getResp := performSystemIntegrationRequest(r, http.MethodGet, "/monitor/system-integrations/"+id, "system-integration-admin-token", nil)
	if getResp.Code != http.StatusOK || decodeSystemIntegration(t, getResp).Name != "FindX Updated Integration" {
		t.Fatalf("persisted integration should be readable, got %d body=%s", getResp.Code, getResp.Body.String())
	}

	deleteResp := performSystemIntegrationRequest(r, http.MethodDelete, "/monitor/system-integrations/"+id, "system-integration-admin-token", nil, map[string]string{"X-Test-Batch-Id": "fx-system-delete"})
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("admin delete should be 200, got %d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	missingResp := performSystemIntegrationRequest(r, http.MethodGet, "/monitor/system-integrations/"+id, "system-integration-admin-token", nil)
	if missingResp.Code != http.StatusNotFound {
		t.Fatalf("deleted integration should be missing, got %d body=%s", missingResp.Code, missingResp.Body.String())
	}
	assertSystemIntegrationAuditRecorded(t, "monitor.integration.create", id)
	assertSystemIntegrationAuditRecorded(t, "monitor.integration.update", id)
	assertSystemIntegrationAuditRecorded(t, "monitor.integration.delete", id)
}

func TestMonitoringSystemIntegrationValidationErrors(t *testing.T) {
	r := monitoringSystemIntegrationTestRouter()
	seedDashboardToken("system-integration-validation-admin-token", "a1", "root", "admin")

	cases := []struct {
		name  string
		input map[string]any
	}{
		{name: "empty name", input: mergeSystemIntegrationInput(validSystemIntegrationInput("fx-test-invalid-empty-name"), map[string]any{"name": ""})},
		{name: "external url", input: mergeSystemIntegrationInput(validSystemIntegrationInput("fx-test-invalid-external-url"), map[string]any{"url": "https://example.invalid"})},
		{name: "secret route", input: mergeSystemIntegrationInput(validSystemIntegrationInput("fx-test-invalid-secret"), map[string]any{"url": "/platform?token=abc", "config_preview": "/platform?token=abc"})},
		{name: "duplicate team", input: mergeSystemIntegrationInput(validSystemIntegrationInput("fx-test-invalid-team"), map[string]any{"team_ids": []int{1, 1}})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := performSystemIntegrationRequest(r, http.MethodPost, "/monitor/system-integrations", "system-integration-validation-admin-token", tc.input)
			if resp.Code != http.StatusBadRequest {
				t.Fatalf("invalid payload should be 400, got %d body=%s", resp.Code, resp.Body.String())
			}
			assertSystemIntegrationResponseSanitized(t, resp.Body.String())
		})
	}
}

func TestMonitoringSystemIntegrationWeightsAndHide(t *testing.T) {
	r := monitoringSystemIntegrationTestRouter()
	seedDashboardToken("system-integration-sort-admin-token", "a1", "root", "admin")
	first := "fx-test-sort-first"
	second := "fx-test-sort-second"
	performSystemIntegrationRequest(r, http.MethodPost, "/monitor/system-integrations", "system-integration-sort-admin-token", validSystemIntegrationInput(first))
	performSystemIntegrationRequest(r, http.MethodPost, "/monitor/system-integrations", "system-integration-sort-admin-token", validSystemIntegrationInput(second))

	weights := []map[string]any{{"id": first, "weight": 90}, {"id": second, "weight": 80}}
	weightResp := performSystemIntegrationRequest(r, http.MethodPut, "/monitor/system-integrations/weights", "system-integration-sort-admin-token", weights)
	if weightResp.Code != http.StatusOK {
		t.Fatalf("weights update should be 200, got %d body=%s", weightResp.Code, weightResp.Body.String())
	}
	items := decodeSystemIntegrations(t, weightResp)
	firstIndex, secondIndex := systemIntegrationIndex(items, first), systemIntegrationIndex(items, second)
	if firstIndex < 0 || secondIndex < 0 || secondIndex >= firstIndex {
		t.Fatalf("weights should reorder items by weight, firstIndex=%d secondIndex=%d items=%+v", firstIndex, secondIndex, items)
	}
	unknownResp := performSystemIntegrationRequest(r, http.MethodPut, "/monitor/system-integrations/weights", "system-integration-sort-admin-token", []map[string]any{{"id": "missing-system-integration", "weight": 1}})
	if unknownResp.Code != http.StatusBadRequest {
		t.Fatalf("unknown weight id should be 400, got %d body=%s", unknownResp.Code, unknownResp.Body.String())
	}

	hideResp := performSystemIntegrationRequest(r, http.MethodPut, "/monitor/system-integrations/"+first+"/hide", "system-integration-sort-admin-token", map[string]any{"hide": true})
	if hideResp.Code != http.StatusOK {
		t.Fatalf("hide toggle should be 200, got %d body=%s", hideResp.Code, hideResp.Body.String())
	}
	hidden := decodeSystemIntegration(t, hideResp)
	if !hidden.Hide || hidden.ShowInMenu {
		t.Fatalf("hide toggle should update show_in_menu: %+v", hidden)
	}
}

func TestMonitoringSystemIntegrationNotFoundAndBuiltinDeleteProtected(t *testing.T) {
	r := monitoringSystemIntegrationTestRouter()
	seedDashboardToken("system-integration-missing-token", "a1", "root", "admin")

	missingResp := performSystemIntegrationRequest(r, http.MethodGet, "/monitor/system-integrations/not-found", "system-integration-missing-token", nil)
	if missingResp.Code != http.StatusNotFound {
		t.Fatalf("missing system integration should be 404, got %d body=%s", missingResp.Code, missingResp.Body.String())
	}
	deleteMissingResp := performSystemIntegrationRequest(r, http.MethodDelete, "/monitor/system-integrations/not-found", "system-integration-missing-token", nil)
	if deleteMissingResp.Code != http.StatusNotFound {
		t.Fatalf("missing delete should be 404, got %d body=%s", deleteMissingResp.Code, deleteMissingResp.Body.String())
	}
	builtinDeleteResp := performSystemIntegrationRequest(r, http.MethodDelete, "/monitor/system-integrations/findx-console-overview", "system-integration-missing-token", nil)
	if builtinDeleteResp.Code != http.StatusConflict {
		t.Fatalf("builtin delete should be 409, got %d body=%s", builtinDeleteResp.Code, builtinDeleteResp.Body.String())
	}
}

func TestMonitoringSystemIntegrationPermissionMatrix(t *testing.T) {
	checker := model.NewMonitorPermissionChecker()
	read := checker.Check("user", "monitor.integration", "read")
	if !read.Known || !read.Allowed {
		t.Fatalf("user should have monitor.integration read permission: %+v", read)
	}
	adminWrite := checker.Check("admin", "monitor.integration", "create")
	if !adminWrite.Known || !adminWrite.Allowed {
		t.Fatalf("admin should have monitor.integration write permission: %+v", adminWrite)
	}
	write := checker.Check("user", "monitor.integration", "create")
	if !write.Known || write.Allowed || write.Reason != model.MonitorPermissionReasonRoleDenied {
		t.Fatalf("monitor.integration create should be known/denied for user: %+v", write)
	}
	unknown := checker.Check("user", "monitor.integration.unknown", "read")
	if unknown.Known || unknown.Allowed {
		t.Fatalf("unknown integration permission should be denied: %+v", unknown)
	}
}

func TestMonitoringSystemIntegrationResponseDoesNotLeakBrandsOrCredentials(t *testing.T) {
	r := monitoringSystemIntegrationTestRouter()
	seedDashboardToken("system-integration-sanitize-token", "u1", "alice", "user")

	resp := performSystemIntegrationRequest(r, http.MethodGet, "/monitor/system-integrations", "system-integration-sanitize-token", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("system integration list should be 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	assertSystemIntegrationResponseSanitized(t, resp.Body.String())
}

func monitoringSystemIntegrationTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	clearDashboardTokens()
	r := gin.New()
	r.GET("/monitor/system-integrations", RequireMonitorPermission("monitor.integration", "read"), ListMonitoringSystemIntegrations)
	r.POST("/monitor/system-integrations", RequireMonitorPermission("monitor.integration", "create"), CreateMonitoringSystemIntegration)
	r.PUT("/monitor/system-integrations/weights", RequireMonitorPermission("monitor.integration", "sort"), UpdateMonitoringSystemIntegrationWeights)
	r.GET("/monitor/system-integrations/:id", RequireMonitorPermission("monitor.integration", "read"), GetMonitoringSystemIntegration)
	r.PUT("/monitor/system-integrations/:id", RequireMonitorPermission("monitor.integration", "update"), UpdateMonitoringSystemIntegration)
	r.DELETE("/monitor/system-integrations/:id", RequireMonitorPermission("monitor.integration", "delete"), DeleteMonitoringSystemIntegration)
	r.PUT("/monitor/system-integrations/:id/hide", RequireMonitorPermission("monitor.integration", "hide"), SetMonitoringSystemIntegrationHide)
	return r
}

func performSystemIntegrationRequest(r *gin.Engine, method, path, token string, body any, headers ...map[string]string) *httptest.ResponseRecorder {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
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

func decodeSystemIntegrations(t *testing.T, w *httptest.ResponseRecorder) []model.MonitoringSystemIntegration {
	t.Helper()
	var items []model.MonitoringSystemIntegration
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("decode system integrations: %v body=%s", err, w.Body.String())
	}
	return items
}

func decodeSystemIntegration(t *testing.T, w *httptest.ResponseRecorder) model.MonitoringSystemIntegration {
	t.Helper()
	var item model.MonitoringSystemIntegration
	if err := json.Unmarshal(w.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode system integration: %v body=%s", err, w.Body.String())
	}
	return item
}

func validSystemIntegrationInput(id string) map[string]any {
	return map[string]any{
		"id":             id,
		"name":           "FindX Test Integration " + id,
		"url":            "/platform?section=" + id,
		"config_preview": "/platform?section=" + id,
		"is_private":     false,
		"team_ids":       []int{101, 102},
		"weight":         55,
		"hide":           false,
	}
}

func mergeSystemIntegrationInput(base map[string]any, patch map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range base {
		out[key] = value
	}
	for key, value := range patch {
		out[key] = value
	}
	return out
}

func assertSystemIntegrationWriteContract(t *testing.T, item model.MonitoringSystemIntegration) {
	t.Helper()
	if !item.Capabilities.Read || !item.Capabilities.List || !item.Capabilities.Detail || !item.Capabilities.Write || !item.Capabilities.Sort {
		t.Fatalf("read/write capabilities should be enabled except embedded open: %+v", item.Capabilities)
	}
	if item.Capabilities.OpenEmbedded || item.Capabilities.MenuEmbedding {
		t.Fatalf("embedded open/menu embedding must remain blocked: %+v", item.Capabilities)
	}
	blocked := map[string]bool{}
	for _, action := range item.BlockedActions {
		if action.Status == model.MonitoringSystemIntegrationStatusBlockedByContract {
			blocked[action.Action] = true
		}
	}
	if !blocked["open_embedded"] || !blocked["menu_embedding"] {
		t.Fatalf("embedded blocked actions missing: %+v", item.BlockedActions)
	}
}

func assertSystemIntegrationResponseSanitized(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{
		"nightingale",
		"n9e",
		"embeddedproduct",
		"embedded-product",
		"categraf",
		"skywalking",
		"signoz",
		"findx.local",
		"token=",
		"password",
		"api_key",
		"apikey",
		"secret",
		"cookie",
		"mysql://",
		"postgres://",
		"authorization=",
		"private key",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("system integration response leaked forbidden value %q: %s", forbidden, body)
		}
	}
}

func assertSystemIntegrationAuditRecorded(t *testing.T, action, target string) {
	t.Helper()
	events := store.ListAuditEvents(1000)
	for _, event := range events {
		if event.Action == action && event.Target == target {
			body, _ := json.Marshal(event)
			assertSystemIntegrationResponseSanitized(t, string(body))
			return
		}
	}
	t.Fatalf("audit event %s target %s not found in %+v", action, target, events)
}

func systemIntegrationIndex(items []model.MonitoringSystemIntegration, id string) int {
	for index, item := range items {
		if item.ID == id {
			return index
		}
	}
	return -1
}
