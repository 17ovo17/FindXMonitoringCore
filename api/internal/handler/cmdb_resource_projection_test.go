package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbResourceProjectionHostAssetUsesLiveMonitorAndCmdbModel(t *testing.T) {
	r := cmdbResourceProjectionTestRouter(t)
	seedResourceAssetToken("projection-token", "a1", "root", "admin")
	_, target, err := store.UpsertFindXAgentHeartbeat(model.FindXAgentHeartbeat{
		Ident:    "projection-host-1",
		IP:       "10.134.1.9",
		Hostname: "projection-host-one",
		OS:       "linux",
		Arch:     "amd64",
		Version:  "3.4.5",
	})
	if err != nil {
		t.Fatalf("seed heartbeat failed: %v", err)
	}
	attrs := []model.CmdbAttribute{
		{ID: "projection-attr-name", ObjectID: "obj-os", Label: "name", Attr: "name", ValueType: "char", Tag: "basic", Discovery: true, Sort: 1},
		{ID: "projection-attr-ip", ObjectID: "obj-os", Label: "management ip", Attr: "ip_address", ValueType: "ip", Tag: "basic", Discovery: true, Sort: 2},
		{ID: "projection-attr-cpu", ObjectID: "obj-os", Label: "cpu cores", Attr: "cpu_cores", ValueType: "int", Tag: "resource", Discovery: true, Sort: 3},
		{ID: "projection-attr-owner", ObjectID: "obj-os", Label: "owner", Attr: "owner", ValueType: "char", Tag: "basic", Sort: 4},
	}
	for i := range attrs {
		if err := store.CreateCmdbAttribute(&attrs[i]); err != nil {
			t.Fatalf("create cmdb attr %s: %v", attrs[i].Attr, err)
		}
	}
	if err := store.CreateCmdbInstance(&model.CmdbInstance{
		ID:       "projection-cmdb-instance",
		ObjectID: "obj-os",
		Data:     `{"name":"projection-host-one","ip_address":"10.134.1.9","cpu_cores":24,"owner":"sensitive-owner","password":"sensitive-password","api_token":"sensitive-token","cookie":"sensitive-cookie","dsn":"sensitive-dsn"}`,
		Creator:  "test",
		Updater:  "test",
	}); err != nil {
		t.Fatalf("create cmdb instance: %v", err)
	}

	resp := performResourceAssetRequest(t, r, http.MethodGet, "/cmdb/resource-projections?resource_type=host_asset", "projection-token", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("resource projection should be 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode projection response: %v body=%s", err, resp.Body.String())
	}
	if body["contract_id"] != "cmdb.resource.projection.read.v1" || body["contract"] != "cmdb.resource.projection.read.v1" {
		t.Fatalf("projection contract mismatch: %#v", body)
	}
	if body["code"] != float64(0) || body["status"] != "ready_with_contract_gaps" {
		t.Fatalf("projection should expose ready_with_contract_gaps status: %#v", body)
	}
	assertStringSliceContains(t, body["identity_fields"], "host_id", "cmdb_instance_id", "cmdb_object_id")
	auditQuery, ok := body["findx_audit_query"].(map[string]any)
	if !ok || auditQuery["source"] != "findx_audit" || auditQuery["scope"] != "cmdb" ||
		auditQuery["resource_type"] != "cmdb_resource_projection" || auditQuery["action"] != "cmdb.resource_projection.read" {
		t.Fatalf("unexpected audit query: %#v", body["findx_audit_query"])
	}
	assertStringSliceContains(t, body["missing_contracts"], "cmdb_resource_projection_runtime_contract", "cmdb_resource_approval_runtime_contract", "cmdb_dashboard_import_runtime_contract")
	assertStringSliceNotContains(t, body["missing_contracts"], "cmdb_resource_projection_source_contract", "cmdb_resource_projection_model_contract")
	columns, ok := body["columns"].([]any)
	if !ok || len(columns) < len(attrs) {
		t.Fatalf("projection should expose dynamic CMDB columns: %#v", body["columns"])
	}
	assertProjectionColumn(t, columns, "cpu_cores", "cpu cores")
	assertProjectionColumn(t, columns, "owner", "owner")

	rows, ok := body["rows"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("projection should include one projected row, got %#v", body["rows"])
	}
	row, ok := rows[0].(map[string]any)
	if !ok {
		t.Fatalf("projection row should be object: %#v", rows[0])
	}
	if row["host_id"] != target.ID || row["cmdb_instance_id"] != "projection-cmdb-instance" || row["cmdb_object_id"] != "obj-os" {
		t.Fatalf("projection row identity mismatch: %#v", row)
	}
	values, ok := row["values"].(map[string]any)
	if !ok {
		t.Fatalf("projection row should expose values map: %#v", row)
	}
	cmdbRef, ok := row["cmdb_instance"].(map[string]any)
	if !ok || cmdbRef["instance_id"] != "projection-cmdb-instance" || cmdbRef["object_id"] != "obj-os" {
		t.Fatalf("projection row should expose cmdb_instance ref: %#v", row)
	}
	if row["name"] != "projection-host-one" || row["ip_address"] != "10.134.1.9" || row["cpu_cores"] != float64(24) {
		t.Fatalf("projection row values mismatch: %#v", row)
	}
	if values["name"] != "projection-host-one" || values["ip_address"] != "10.134.1.9" || values["cpu_cores"] != float64(24) {
		t.Fatalf("projection row values map mismatch: %#v", values)
	}
	if row["owner"] != cmdbMaskedValue || values["owner"] != cmdbMaskedValue {
		t.Fatalf("projection should mask sensitive owner, got row=%#v values=%#v", row["owner"], values["owner"])
	}
	lowerBody := strings.ToLower(resp.Body.String())
	for _, forbidden := range []string{"sensitive-owner", "sensitive-password", "sensitive-token", "sensitive-cookie", "sensitive-dsn", "password", "api_token", "cookie", "dsn", "fake"} {
		if strings.Contains(lowerBody, forbidden) {
			t.Fatalf("projection response should not expose %q: %s", forbidden, resp.Body.String())
		}
	}
}

func TestCmdbResourceProjectionReusesHostAssetFilters(t *testing.T) {
	r := cmdbResourceProjectionTestRouter(t)
	seedResourceAssetToken("projection-filter-token", "a1", "root", "admin")
	_, target, err := store.UpsertFindXAgentHeartbeat(model.FindXAgentHeartbeat{
		Ident:    "projection-filter-1",
		IP:       "10.134.2.9",
		Hostname: "projection-filter-one",
		OS:       "linux",
		Arch:     "amd64",
		Version:  "3.4.6",
	})
	if err != nil {
		t.Fatalf("seed heartbeat failed: %v", err)
	}
	target.Labels[hostLabelWorkspaceID] = "projection-ws"
	target.Labels[hostLabelResourceGroupID] = "projection-rg"
	target.Labels[hostLabelTags] = "ops,prod"
	if _, err := store.UpsertMonitorTarget(target); err != nil {
		t.Fatalf("update target labels: %v", err)
	}
	attrs := []model.CmdbAttribute{
		{ID: "projection-filter-attr-name", ObjectID: "obj-os", Label: "name", Attr: "name", ValueType: "char", Tag: "basic", Discovery: true, Sort: 1},
		{ID: "projection-filter-attr-ip", ObjectID: "obj-os", Label: "management ip", Attr: "ip_address", ValueType: "ip", Tag: "basic", Discovery: true, Sort: 2},
	}
	for i := range attrs {
		if err := store.CreateCmdbAttribute(&attrs[i]); err != nil {
			t.Fatalf("create cmdb attr %s: %v", attrs[i].Attr, err)
		}
	}
	if err := store.CreateCmdbInstance(&model.CmdbInstance{
		ID:       "projection-filter-instance",
		ObjectID: "obj-os",
		Data:     `{"name":"projection-filter-one","ip_address":"10.134.2.9"}`,
		Creator:  "test",
		Updater:  "test",
	}); err != nil {
		t.Fatalf("create cmdb instance: %v", err)
	}

	okResp := performResourceAssetRequest(t, r, http.MethodGet, "/cmdb/resource-projections?resource_type=host_asset&workspace_id=projection-ws&resource_group_id=projection-rg&tag=ops&keyword=projection-filter-one", "projection-filter-token", nil)
	if okResp.Code != http.StatusOK {
		t.Fatalf("filtered projection should be 200, got %d body=%s", okResp.Code, okResp.Body.String())
	}
	missResp := performResourceAssetRequest(t, r, http.MethodGet, "/cmdb/resource-projections?resource_type=host_asset&workspace_id=missing-ws", "projection-filter-token", nil)
	if missResp.Code != http.StatusConflict {
		t.Fatalf("projection with no filtered rows should be 409, got %d body=%s", missResp.Code, missResp.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(missResp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode filtered blocked response: %v body=%s", err, missResp.Body.String())
	}
	if _, ok := body["rows"]; ok {
		t.Fatalf("filtered empty projection must not return successful rows body: %#v", body)
	}
	assertStringSliceContains(t, body["missing_contracts"], "cmdb_resource_projection_source_contract", "cmdb_resource_projection_model_contract")
}

func TestCmdbResourceProjectionUnsupportedTypeBlockedByContract(t *testing.T) {
	r := cmdbResourceProjectionTestRouter(t)
	seedResourceAssetToken("projection-block-token", "a1", "root", "admin")

	resp := performResourceAssetRequest(t, r, http.MethodGet, "/cmdb/resource-projections?resource_type=database_asset", "projection-block-token", nil)
	if resp.Code != http.StatusConflict {
		t.Fatalf("unsupported projection should be 409, got %d body=%s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode blocked response: %v body=%s", err, resp.Body.String())
	}
	if body["code"] != cmdbBlockedByContract || body["contract_id"] != "cmdb.resource.projection.read.v1" || body["contract"] != "cmdb.resource.projection.read.v1" {
		t.Fatalf("blocked contract envelope mismatch: %#v", body)
	}
	assertStringSliceContains(t, body["missing_contracts"], "cmdb_resource_projection_source_contract", "cmdb_resource_projection_model_contract")
	if _, ok := body["rows"]; ok {
		t.Fatalf("unsupported type must not return successful rows body: %#v", body)
	}
}

func cmdbResourceProjectionTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	r := resourceAssetTestRouter(t)
	r.GET("/cmdb/resource-projections", RequireAuth(), GetCmdbResourceProjection)
	return r
}

func assertProjectionColumn(t *testing.T, columns []any, attr string, label string) {
	t.Helper()
	for _, item := range columns {
		col, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if col["attr"] == attr && col["label"] == label {
			return
		}
	}
	t.Fatalf("missing projection column attr=%s label=%s in %#v", attr, label, columns)
}

func assertStringSliceContains(t *testing.T, raw any, want ...string) {
	t.Helper()
	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected string slice, got %#v", raw)
	}
	seen := map[string]bool{}
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("expected string item, got %#v", item)
		}
		seen[text] = true
	}
	for _, item := range want {
		if !seen[item] {
			t.Fatalf("missing %q in %#v", item, raw)
		}
	}
}

func assertStringSliceNotContains(t *testing.T, raw any, forbidden ...string) {
	t.Helper()
	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected string slice, got %#v", raw)
	}
	seen := map[string]bool{}
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("expected string item, got %#v", item)
		}
		seen[text] = true
	}
	for _, item := range forbidden {
		if seen[item] {
			t.Fatalf("unexpected %q in %#v", item, raw)
		}
	}
}
