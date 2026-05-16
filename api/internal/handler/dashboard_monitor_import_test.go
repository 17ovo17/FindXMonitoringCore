package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func TestAdminImportsMonitorDashboardTemplateBlocksWithoutRuntimeExecutor(t *testing.T) {
	r := monitorDashboardTestRouter()
	seedDashboardToken("dashboard-template-admin-token", "a1", "root", "admin")
	before, err := store.ListMonitorDashboards()
	if err != nil {
		t.Fatalf("list dashboards before import: %v", err)
	}
	payload := map[string]any{
		"title":             "导入 Linux 主机",
		"workspace_id":      "ws-template",
		"resource_group_id": "rg-template",
		"variables":         map[string]any{"ident": "host-a", "instance": "node-a:9100"},
		"tags":              []string{"ops", "prod"},
	}
	resp := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboard-templates/linux-host-basic/import", "dashboard-template-admin-token", payload)
	if resp.Code != http.StatusConflict {
		t.Fatalf("admin template import should fail close with 409, got %d body=%s", resp.Code, resp.Body.String())
	}
	payloadResp := decodeDashboardImportBlocked(t, resp)
	if payloadResp.Code != "PENDING" ||
		payloadResp.ContractID != "cmdb.dashboard.import.runtime.v1" ||
		payloadResp.Status != "PENDING" ||
		payloadResp.SafeToRetry {
		t.Fatalf("template import blocked envelope mismatch: %#v", payloadResp)
	}
	for _, want := range []string{
		"cmdb_dashboard_template_lookup_contract",
		"cmdb_dashboard_import_runtime_contract",
		"cmdb_dashboard_import_batch_result_contract",
		"cmdb_dashboard_import_conflict_rollback_contract",
	} {
		if !containsDashboardTestString(payloadResp.MissingContracts, want) {
			t.Fatalf("blocked template import should include %q: %#v", want, payloadResp)
		}
	}
	after, err := store.ListMonitorDashboards()
	if err != nil {
		t.Fatalf("list dashboards after import: %v", err)
	}
	if len(after) != len(before) {
		t.Fatalf("blocked template import must not create dashboards, before=%d after=%d body=%s", len(before), len(after), resp.Body.String())
	}
	for _, forbidden := range []string{`"status":"active"`, `"shared":true`, `"share_token_set":true`} {
		if strings.Contains(strings.ToLower(resp.Body.String()), forbidden) {
			t.Fatalf("blocked import must not expose dashboard success shape %q: %s", forbidden, resp.Body.String())
		}
	}
	assertDashboardResponseSanitized(t, resp.Body.String())
}

func TestMonitorDashboardTemplateImportKeepsExecutorDedupeBatchRollbackReceiptContracts(t *testing.T) {
	r := monitorDashboardTestRouter()
	seedDashboardToken("dashboard-template-contract-admin-token", "a1", "root", "admin")
	payload := map[string]any{
		"title":             "FX-NIGHT-157 import contract inventory",
		"workspace_id":      "ws-fx-night-157",
		"resource_group_id": "rg-fx-night-157",
		"variables":         map[string]any{"ident": "host-a", "instance": "node-a:9100"},
		"tags":              []string{"ops", "contract"},
	}
	resp := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboard-templates/linux-host-basic/import", "dashboard-template-contract-admin-token", payload)
	if resp.Code != http.StatusConflict {
		t.Fatalf("template import should remain blocked, got %d body=%s", resp.Code, resp.Body.String())
	}
	payloadResp := decodeDashboardImportBlocked(t, resp)
	for _, want := range []string{
		"cmdb_dashboard_import_executor_contract",
		"cmdb_dashboard_import_dedupe_contract",
		"cmdb_dashboard_import_batch_receipt_contract",
		"cmdb_dashboard_import_rollback_receipt_contract",
	} {
		if !containsDashboardTestString(payloadResp.MissingContracts, want) {
			t.Fatalf("dashboard import blocked gate should expose %q, payload=%#v body=%s", want, payloadResp, resp.Body.String())
		}
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode import contract body: %v body=%s", err, resp.Body.String())
	}
	receiptContract, ok := body["receipt_contract"].(map[string]any)
	if !ok {
		t.Fatalf("dashboard import blocked gate should expose receipt_contract: %s", resp.Body.String())
	}
	if receiptContract["status"] != "PENDING" || receiptContract["contract_id"] != "cmdb.dashboard.import.receipts.v1" {
		t.Fatalf("dashboard import receipt_contract mismatch: %#v", receiptContract)
	}
	required, ok := receiptContract["required_receipts"].([]any)
	if !ok {
		t.Fatalf("dashboard import receipt_contract should expose required_receipts: %#v", receiptContract)
	}
	for _, want := range []string{"dedupe_receipt", "batch_result_receipt", "rollback_receipt"} {
		if !dashboardAnySliceContains(required, want) {
			t.Fatalf("dashboard import required_receipts missing %q: %#v", want, receiptContract)
		}
	}
	lowerBody := strings.ToLower(resp.Body.String())
	for _, forbidden := range []string{`"status":"imported"`, `"status":"success"`, `"success"`, `"imported"`, `"panels":[]`, `"receipts":[]`} {
		if strings.Contains(lowerBody, forbidden) {
			t.Fatalf("dashboard import blocked gate must not expose fake success marker %q: %s", forbidden, resp.Body.String())
		}
	}
}

func TestAdminImportsMonitorDashboardTemplateAuditsBlockedRequestIntent(t *testing.T) {
	r := monitorDashboardTestRouter()
	seedDashboardToken("dashboard-template-audit-admin-token", "a1", "root", "admin")
	traceID := "fx-night-133-audit-trace"
	payload := map[string]any{
		"title":             "FX-NIGHT-133 audit import",
		"workspace_id":      "ws-fx-night-133-audit",
		"resource_group_id": "rg-fx-night-133-audit",
		"variables": map[string]any{
			"ident":    "host-a",
			"instance": "node-a:9100",
			"password": "should-not-leak",
			"token":    "<TOKEN>",
			"dsn":      "mysql://user:password@example.local/db",
		},
		"tags": []string{"ops", "audit"},
	}
	resp := performDashboardRequestWithHeaders(t, r, http.MethodPost, "/monitor/dashboard-templates/linux-host-basic/import", "dashboard-template-audit-admin-token", payload, map[string]string{
		"X-Test-Batch-Id": traceID,
	})
	if resp.Code != http.StatusConflict {
		t.Fatalf("admin template import should fail close with 409, got %d body=%s", resp.Code, resp.Body.String())
	}
	page, err := store.ListMonitorAuditLogs(model.MonitorAuditLogQuery{
		Action:       "dashboard.template.import",
		ResourceType: "cmdb_dashboard_import",
		ResourceID:   "linux-host-basic",
		Scope:        "cmdb",
		TraceID:      traceID,
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("list monitor audit logs: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("blocked import request intent should write exactly one audit log, page=%+v", page)
	}
	audit := page.Items[0]
	if audit.Status != "blocked" && audit.Status != "requested" {
		t.Fatalf("audit status should record blocked/requested intent, got %+v", audit)
	}
	auditBody := strings.ToLower(mustMarshalDashboardTestJSON(t, audit))
	for _, forbidden := range []string{"panels", "variables", "should-not-leak", "<token>", "mysql://", "password", "cookie", "dsn", "secret"} {
		if strings.Contains(auditBody, forbidden) {
			t.Fatalf("dashboard import audit must contain only safe summary, leaked %q in %s", forbidden, auditBody)
		}
	}
}

func TestMonitorDashboardTemplateImportDedupeBlocksExistingDashboard(t *testing.T) {
	r := monitorDashboardTestRouter()
	seedDashboardToken("dashboard-template-dedupe-admin-token", "a1", "root", "admin")
	title := "FX-NIGHT-133 dedupe import"
	workspaceID := "ws-fx-night-133-dedupe"
	resourceGroupID := "rg-fx-night-133-dedupe"
	existing, err := store.SaveMonitorDashboard(&model.MonitorDashboard{
		Title:           title,
		WorkspaceID:     workspaceID,
		ResourceGroupID: resourceGroupID,
		Variables:       json.RawMessage(`{"env":"prod"}`),
		Panels:          json.RawMessage(`[{"title":"CPU","query":"up"}]`),
		Status:          model.MonitorDashboardStatusActive,
	}, "seed")
	if err != nil {
		t.Fatalf("seed existing dashboard: %v", err)
	}
	before, err := store.ListMonitorDashboards()
	if err != nil {
		t.Fatalf("list dashboards before import: %v", err)
	}
	payload := map[string]any{
		"title":             title,
		"workspace_id":      workspaceID,
		"resource_group_id": resourceGroupID,
		"variables":         map[string]any{"ident": "host-a", "instance": "node-a:9100"},
	}
	resp := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboard-templates/linux-host-basic/import", "dashboard-template-dedupe-admin-token", payload)
	if resp.Code != http.StatusConflict {
		t.Fatalf("duplicate template import should fail close with 409, got %d body=%s", resp.Code, resp.Body.String())
	}
	payloadResp := decodeDashboardImportBlocked(t, resp)
	if !containsDashboardTestString(payloadResp.MissingContracts, "cmdb_dashboard_import_dedup_contract") {
		t.Fatalf("duplicate import should include dedupe missing contract, payload=%#v body=%s", payloadResp, resp.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode duplicate import body: %v body=%s", err, resp.Body.String())
	}
	dedupe, ok := body["dedupe"].(map[string]any)
	if !ok {
		t.Fatalf("duplicate import should include dedupe locator, body=%s", resp.Body.String())
	}
	if dedupe["reason"] != "existing_dashboard" {
		t.Fatalf("duplicate import dedupe reason mismatch: %#v", dedupe)
	}
	existingLocator, ok := dedupe["existing_dashboard"].(map[string]any)
	if !ok || existingLocator["id"] != existing.ID || existingLocator["title"] != title || existingLocator["workspace_id"] != workspaceID || existingLocator["resource_group_id"] != resourceGroupID {
		t.Fatalf("duplicate import should expose safe existing dashboard locator, locator=%#v existing=%+v", existingLocator, existing)
	}
	after, err := store.ListMonitorDashboards()
	if err != nil {
		t.Fatalf("list dashboards after import: %v", err)
	}
	if len(after) != len(before) {
		t.Fatalf("duplicate blocked import must not create dashboards, before=%d after=%d body=%s", len(before), len(after), resp.Body.String())
	}
	lowerBody := strings.ToLower(resp.Body.String())
	for _, forbidden := range []string{`"panels"`, `"variables"`, `"status":"active"`, `"status":"ok"`, `"status":"success"`, `"status":"imported"`, `"success"`, `"imported"`} {
		if strings.Contains(lowerBody, forbidden) {
			t.Fatalf("duplicate blocked import must not expose success dashboard shape %q: %s", forbidden, resp.Body.String())
		}
	}
}

func TestMonitorDashboardTemplateImportRejectsInvalidRequests(t *testing.T) {
	r := monitorDashboardTestRouter()
	seedDashboardToken("dashboard-template-invalid-admin-token", "a1", "root", "admin")
	missing := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboard-templates/not-found/import", "dashboard-template-invalid-admin-token", nil)
	if missing.Code != http.StatusConflict {
		t.Fatalf("missing template import should fail close with 409, got %d body=%s", missing.Code, missing.Body.String())
	}
	missingPayload := decodeDashboardImportBlocked(t, missing)
	if missingPayload.ContractID != "cmdb.dashboard.import.runtime.v1" ||
		!containsDashboardTestString(missingPayload.MissingContracts, "cmdb_dashboard_template_lookup_contract") ||
		missingPayload.SafeToRetry {
		t.Fatalf("missing template import should return dashboard import contract, payload=%#v", missingPayload)
	}
	for _, forbidden := range []string{`"status":"active"`, `"dashboard"`, `"panels"`, `"uid"`} {
		if strings.Contains(strings.ToLower(missing.Body.String()), forbidden) {
			t.Fatalf("missing template import must not expose success dashboard shape %q: %s", forbidden, missing.Body.String())
		}
	}
	emptyGroup := map[string]any{"resource_group_id": " "}
	resp := performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboard-templates/linux-host-basic/import", "dashboard-template-invalid-admin-token", emptyGroup)
	if resp.Code != http.StatusBadRequest || !strings.Contains(resp.Body.String(), "resource_group_id is required") {
		t.Fatalf("empty template import resource group should be 400 with clear check, got %d body=%s", resp.Code, resp.Body.String())
	}
	badVariables := map[string]any{"resource_group_id": "rg-template", "variables": []any{"not-object"}}
	resp = performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboard-templates/linux-host-basic/import", "dashboard-template-invalid-admin-token", badVariables)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("non-object template variables should be 400, got %d body=%s", resp.Code, resp.Body.String())
	}
	longTitle := map[string]any{"title": strings.Repeat("x", maxDashboardTitleLen+1), "resource_group_id": "rg-template"}
	resp = performDashboardRequest(t, r, http.MethodPost, "/monitor/dashboard-templates/linux-host-basic/import", "dashboard-template-invalid-admin-token", longTitle)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("too long template import title should be 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func decodeDashboardImportBlocked(t *testing.T, w *httptest.ResponseRecorder) struct {
	Code             string   `json:"code"`
	Status           string   `json:"status"`
	ContractID       string   `json:"contract_id"`
	MissingContracts []string `json:"missing_contracts"`
	SafeToRetry      bool     `json:"safe_to_retry"`
} {
	t.Helper()
	var payload struct {
		Code             string   `json:"code"`
		Status           string   `json:"status"`
		ContractID       string   `json:"contract_id"`
		MissingContracts []string `json:"missing_contracts"`
		SafeToRetry      bool     `json:"safe_to_retry"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode dashboard import blocked response: %v body=%s", err, w.Body.String())
	}
	return payload
}

func containsDashboardTestString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func dashboardAnySliceContains(values []any, want string) bool {
	for _, value := range values {
		if text, ok := value.(string); ok && text == want {
			return true
		}
	}
	return false
}

func mustMarshalDashboardTestJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal dashboard test json: %v", err)
	}
	return string(data)
}
