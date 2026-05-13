package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbInstancesCompatibleContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	objectID := "cmdb-compatible-os"
	createCmdbCompatibleFixture(t, objectID)

	router := gin.New()
	router.GET("/api/v1/cmdb/objects/:id/instances-compatible", ListCmdbInstancesCompatible)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/objects/"+objectID+"/instances-compatible?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := body["tree"].([]any); !ok {
		t.Fatalf("tree missing or invalid: %#v", body["tree"])
	}
	instances := body["instances"].(map[string]any)
	if instances["total"].(float64) < 1 {
		t.Fatalf("instances.total = %v, want >= 1", instances["total"])
	}
	list := instances["list"].([]any)
	first := list[0].(map[string]any)
	attribute := first["attribute"].(map[string]any)
	for _, code := range []string{"OS001", "OS004", "x5qvHkM1Bz1661218322", "eNuJHg9BTb1682409230"} {
		if _, ok := attribute[code]; !ok {
			t.Fatalf("attribute code %s missing in %#v", code, attribute)
		}
	}
	columns := body["columns"].([]any)
	assertCmdbColumnHasContractFields(t, columns, "OS001")
	assertCmdbColumnHasContractFields(t, columns, "eNuJHg9BTb1682409230")
	if _, ok := body["items"].([]any); !ok {
		t.Fatalf("legacy items missing: %#v", body["items"])
	}
	meta := body["meta"].(map[string]any)
	persistence := meta["persistence"].(map[string]any)
	if persistence["status"] != "blocked_by_persistence" {
		t.Fatalf("persistence.status = %v, want blocked_by_persistence", persistence["status"])
	}
}

func TestCmdbInstanceDetailCompatibleMasksPII(t *testing.T) {
	gin.SetMode(gin.TestMode)
	objectID := "cmdb-compatible-os-detail"
	instanceID := createCmdbCompatibleFixture(t, objectID)

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/detail-compatible", GetCmdbInstanceDetailCompatible)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/"+instanceID+"/detail-compatible", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	base := body["base"].([]any)
	if len(base) == 0 {
		t.Fatal("base groups missing")
	}
	foundTag := false
	foundMaskedOwner := false
	foundMaskedPhone := false
	for _, groupRaw := range base {
		group := groupRaw.(map[string]any)
		if group["tag"] == "基本信息" {
			foundTag = true
		}
		for _, infoRaw := range group["infos"].([]any) {
			info := infoRaw.(map[string]any)
			switch info["attr"] {
			case "x5qvHkM1Bz1661218322":
				foundMaskedOwner = info["value"] == cmdbMaskedValue && info["masked"] == true
			case "eNuJHg9BTb1682409230":
				foundMaskedPhone = info["value"] == cmdbMaskedValue && info["masked"] == true
			}
		}
	}
	if !foundTag {
		t.Fatal("base[].tag did not include 基本信息")
	}
	if !foundMaskedOwner || !foundMaskedPhone {
		t.Fatalf("PII mask missing, owner=%v phone=%v body=%s", foundMaskedOwner, foundMaskedPhone, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "15678908879") || strings.Contains(w.Body.String(), "黎键辉") {
		t.Fatal("detail response leaked raw PII")
	}
}

func TestCmdbCompatibleBlockedGates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/cmdb/instances/inst-1/topology"},
		{http.MethodGet, "/api/v1/cmdb/monitor-bindings/inst-1"},
		{http.MethodPost, "/api/v1/cmdb/monitor-bindings/inst-1"},
	}
	for _, tt := range cases {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusConflict {
				t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body["code"] != cmdbBlockedByContract {
				t.Fatalf("code = %v, want %s", body["code"], cmdbBlockedByContract)
			}
			if body["contract_id"] == "" {
				t.Fatal("contract_id missing")
			}
			if body["safe_to_retry"] != false {
				t.Fatalf("safe_to_retry = %v, want false", body["safe_to_retry"])
			}
			if len(body["missing_contracts"].([]any)) == 0 {
				t.Fatal("missing_contracts missing")
			}
		})
	}
}

func TestCmdbInstanceTopologyBlockedContractFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/inst-1/topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["contract_id"] != "cmdb.instance.topology.v1" {
		t.Fatalf("contract_id = %v", body["contract_id"])
	}
	if body["safe_to_retry"] != false {
		t.Fatalf("safe_to_retry = %v, want false", body["safe_to_retry"])
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"cmdb_topology_field_mapping_contract",
		"expected_schema",
		"field_matrix",
		"source_evidence",
		"object_id",
		"instances",
		"children",
		"relation_id",
		"asst_id",
		"asst_name",
		"direction",
		"location",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"nodes"`, `"edges"`})
}

func TestCmdbMonitorBindingsReadBlockedContractFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/inst-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["contract_id"] != "cmdb.monitor_bindings.read.v1" {
		t.Fatalf("contract_id = %v", body["contract_id"])
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"cmdb_monitor_binding_field_mapping_contract",
		"host",
		"hostid",
		"templateid",
		"server_object_id",
		"server_platform_id",
		"cmdb_object_id",
		"group",
		"tags",
		"active_status",
		"hosttype",
		"subtype",
	})
}

func TestCmdbMonitorBindingsWriteBlockedContractAvoidsFakeStateAndSensitiveEcho(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)

	payload := `{"host":"srv-1","password":"secret-marker","token":"token-marker","status":"success"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/inst-1", strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["contract_id"] != "cmdb.monitor_bindings.write.v1" {
		t.Fatalf("contract_id = %v", body["contract_id"])
	}
	if body["safe_to_retry"] != false {
		t.Fatalf("safe_to_retry = %v, want false", body["safe_to_retry"])
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"binding_audit_contract",
		"binding_rollback_contract",
		"cmdb_monitor_binding_write_receipt_contract",
		"expected_schema",
		"field_matrix",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		"secret-marker",
		"token-marker",
		`"success"`,
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"data_arrived"`,
		`"service_registered"`,
		`"rolled_back"`,
		`"uninstalled"`,
	})
}

func createCmdbCompatibleFixture(t *testing.T, objectID string) string {
	t.Helper()
	err := store.CreateCmdbObject(&model.CmdbObject{
		ID:         objectID,
		Name:       "操作系统",
		CategoryID: "cmdb-compatible-category",
		ObjectType: 101,
	})
	if err != nil {
		t.Fatalf("create object: %v", err)
	}
	attrs := []model.CmdbAttribute{
		{ID: objectID + "-name", ObjectID: objectID, Label: "操作系统名称", Attr: "name", ValueType: "char", Tag: "基本信息", Required: true, Sort: 1},
		{ID: objectID + "-ip", ObjectID: objectID, Label: "IP地址", Attr: "OS001", ValueType: "ip", Tag: "基本信息", Required: true, Unique: true, Sort: 2},
		{ID: objectID + "-os", ObjectID: objectID, Label: "系统版本", Attr: "OS004", ValueType: "char", Tag: "基本信息", Sort: 3},
		{ID: objectID + "-owner", ObjectID: objectID, Label: "资产负责人", Attr: "x5qvHkM1Bz1661218322", ValueType: "char", Tag: "基本信息", Sort: 4},
		{ID: objectID + "-phone", ObjectID: objectID, Label: "联系电话", Attr: "eNuJHg9BTb1682409230", ValueType: "char", Tag: "基本信息", Sort: 5},
	}
	for i := range attrs {
		if err := store.CreateCmdbAttribute(&attrs[i]); err != nil {
			t.Fatalf("create attr %s: %v", attrs[i].Attr, err)
		}
	}

	data := `{"name":"debian兼容测试","OS001":"192.168.3.163","OS004":"Debian GNU/Linux 12 (bookworm)","x5qvHkM1Bz1661218322":"黎键辉","eNuJHg9BTb1682409230":"15678908879"}`
	inst := &model.CmdbInstance{
		ObjectID: objectID,
		Data:     data,
		Creator:  "自动发现",
		Updater:  "测试",
	}
	if err := store.CreateCmdbInstance(inst); err != nil {
		t.Fatalf("create instance: %v", err)
	}
	return inst.ID
}

func assertCmdbColumnHasContractFields(t *testing.T, columns []any, attr string) {
	t.Helper()
	for _, raw := range columns {
		column := raw.(map[string]any)
		if column["attr"] != attr {
			continue
		}
		for _, key := range []string{"attr", "label", "value_type", "tag", "required", "unique", "visible", "conversion", "sensitive", "mask_policy"} {
			if _, ok := column[key]; !ok {
				t.Fatalf("column %s missing key %s: %#v", attr, key, column)
			}
		}
		return
	}
	t.Fatalf("column attr %s missing in %#v", attr, columns)
}

func assertCmdbTestStringContainsAll(t *testing.T, text string, wants []string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Fatalf("response missing %q: %s", want, text)
		}
	}
}

func assertCmdbTestStringExcludesAll(t *testing.T, text string, blocked []string) {
	t.Helper()
	for _, token := range blocked {
		if strings.Contains(text, token) {
			t.Fatalf("response contains forbidden token %q: %s", token, text)
		}
	}
}
