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

func TestCmdbMonitorBindingWriteReadAndAuditLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-real")
	createCmdbMonitorTargetFixture(t, "host-1001")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-real-template")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"host":"srv-1",
		"hostid":"host-1001",
		"templateid":"` + templateID + `",
		"server_object_id":"server-object-linux",
		"server_platform_id":"linux",
		"cmdb_object_id":"cmdb-monitor-binding-real",
		"group":{"name":"业务系统A","token":"token-marker"},
		"tags":{"env":"prod","dsn":"mysql://user:pass@example/db"},
		"active_status":"active",
		"hosttype":"server",
		"subtype":"linux",
		"hosttypeLabel":"服务器",
		"subtypeLabel":"Linux",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip",
		"server_model_id":"model-linux",
		"server_model_name":"Linux 主机",
		"attr":"OS001",
		"attr_stru":[{"cmdb_attr_id":"OS001","server_attr_id":"agent.ip"}],
		"queue":"monitor-binding-queue",
		"password":"secret-marker",
		"status":"success"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		`"status":"ready"`,
		`"hostid":"host-1001"`,
		`"templateid":"` + templateID + `"`,
		`"audit_ref"`,
		`"receipts"`,
		`"delivery_status":"PENDING"`,
		`"effect_status":"PENDING"`,
		`"rollback_status":"PENDING"`,
		`"findx_audit"`,
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		"secret-marker",
		"token-marker",
		"mysql://user:pass@example/db",
		`"success"`,
		`"queued"`,
		`"running"`,
		`"installed"`,
		`"data_arrived"`,
	})

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusConflict {
		t.Fatalf("read status = %d, want %d, body=%s", getW.Code, http.StatusConflict, getW.Body.String())
	}
	created := map[string]any{}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	binding := created["binding"].(map[string]any)
	receipts := binding["receipts"].([]any)
	deliveryRequestRef := cmdbMonitorBindingReceiptRequestRef(t, receipts, "delivery")
	if strings.TrimSpace(deliveryRequestRef) == "" {
		t.Fatalf("delivery receipt should reference a stored execution task request: %#v", receipts)
	}
	deliveryTask, ok, err := store.GetFindXAgentExecutionTask(deliveryRequestRef)
	if err != nil {
		t.Fatalf("query delivery task ref: %v", err)
	}
	if !ok || deliveryTask.Status != "blocked" || deliveryTask.Action != "cmdb-monitor-binding-delivery" ||
		deliveryTask.Metadata["cmdb_binding_id"] != binding["id"] || deliveryTask.Metadata["cmdb_receipt_type"] != "delivery" {
		t.Fatalf("delivery request_ref should resolve to blocked CMDB execution task, ok=%v task=%#v binding=%#v", ok, deliveryTask, binding)
	}
	assertCmdbTestStringContainsAll(t, getW.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.runtime.receipts.read.v1",
		"cmdb_monitor_binding_delivery_executor_contract",
		"cmdb_monitor_binding_effect_executor_contract",
		"cmdb_monitor_binding_rollback_executor_contract",
		"cmdb_monitor_binding_attested_receipt_contract",
	})
	assertCmdbTestStringExcludesAll(t, getW.Body.String(), []string{
		"secret-marker",
		"token-marker",
		"mysql://user:pass@example/db",
		`"code":0`,
		`"status":"ready"`,
		`"bindings":[`,
		`"receipts":[`,
	})

	auditResp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
		Source:       model.LogsSourceFindXAudit,
		Scope:        "cmdb",
		ResourceType: "cmdb_monitor_binding",
		ResourceID:   instanceID,
		Action:       "cmdb.monitor_binding.save",
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	if len(auditResp.Items) == 0 {
		t.Fatalf("monitor binding audit log missing: %+v", auditResp)
	}
	for _, row := range auditResp.Items {
		details, _ := row.Attributes["details"].(map[string]any)
		if details["binding_id"] == "" || details["hostid"] == "secret-marker" {
			t.Fatalf("audit row missing binding details or leaked sensitive data: %+v", row)
		}
	}
}

func TestCmdbMonitorBindingReceiptQueryKeepsExecutorGapAfterStoredReceipts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-receipt-query")
	createCmdbMonitorTargetFixture(t, "host-receipt-1001")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-receipt-template")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"host":"srv-receipt",
		"hostid":"host-receipt-1001",
		"templateid":"` + templateID + `",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip",
		"tags":{"env":"prod","token":"receipt-token-marker"},
		"password":"receipt-secret-marker"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("write status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	receiptsReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID+"/receipts", nil)
	receiptsW := httptest.NewRecorder()
	router.ServeHTTP(receiptsW, receiptsReq)
	if receiptsW.Code != http.StatusConflict {
		t.Fatalf("receipt query status = %d, want %d, body=%s", receiptsW.Code, http.StatusConflict, receiptsW.Body.String())
	}
	assertCmdbTestStringContainsAll(t, receiptsW.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.runtime.receipts.read.v1",
		"cmdb_monitor_binding_delivery_executor_contract",
		"cmdb_monitor_binding_effect_executor_contract",
		"cmdb_monitor_binding_rollback_executor_contract",
		"cmdb_monitor_binding_attested_receipt_contract",
	})
	assertCmdbTestStringExcludesAll(t, receiptsW.Body.String(), []string{
		"receipt-token-marker",
		"receipt-secret-marker",
		`"code":0`,
		`"status":"ready"`,
		`"receipts":[`,
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"data_arrived"`,
		`"rolled_back"`,
	})
}

func TestCmdbMonitorBindingRejectsUnknownMonitorTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-missing-host")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-missing-host-template")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"hostid":"missing-monitor-target",
		"templateid":"` + templateID + `",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip",
		"password":"missing-host-secret-marker"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"monitor_host_binding_contract",
		"cmdb.monitor_binding.host_target.v1",
		"missing-monitor-target",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"audit_ref":"`,
		`"cmdb.monitor_binding.save"`,
		"missing-host-secret-marker",
	})

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusConflict {
		t.Fatalf("read status = %d, want %d, body=%s", getW.Code, http.StatusConflict, getW.Body.String())
	}

	auditResp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
		Source:       model.LogsSourceFindXAudit,
		Scope:        "cmdb",
		ResourceType: "cmdb_monitor_binding",
		ResourceID:   instanceID,
		Action:       "cmdb.monitor_binding.save",
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	if len(auditResp.Items) != 0 {
		t.Fatalf("unexpected audit rows for rejected monitor target: %+v", auditResp.Items)
	}
}

func TestCmdbMonitorBindingReceiptQueryEmptyRemainsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-receipt-empty")

	router := gin.New()
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID+"/receipts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.receipts.read.v1",
		"cmdb_monitor_binding_receipt_store",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"receipts":[`, `"code":0`, `"ready"`})
}

func TestCmdbMonitorBindingReadEmptyRemainsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-empty")
	router := gin.New()
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{"PENDING", "cmdb_monitor_binding_store"})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"bindings":[`, `"code":0`, `"success":true`})
}

func TestCmdbMonitorBindingRejectsUnknownRuntimeTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-template-runtime")
	createCmdbMonitorTargetFixture(t, "host-1002")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"hostid":"host-1002",
		"templateid":"missing-runtime-template",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode blocked response: %v", err)
	}
	if body["status"] != "PENDING" {
		t.Fatalf("top-level status = %v, want PENDING: %s", body["status"], w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"monitor_template_runtime_contract",
		"missing-runtime-template",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"audit_ref":"`,
		`"cmdb.monitor_binding.save"`,
	})

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusConflict {
		t.Fatalf("read status = %d, want %d, body=%s", getW.Code, http.StatusConflict, getW.Body.String())
	}

	auditResp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
		Source:       model.LogsSourceFindXAudit,
		Scope:        "cmdb",
		ResourceType: "cmdb_monitor_binding",
		ResourceID:   instanceID,
		Action:       "cmdb.monitor_binding.save",
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	if len(auditResp.Items) != 0 {
		t.Fatalf("unexpected audit rows for rejected template: %+v", auditResp.Items)
	}
}

func TestCmdbMonitorBindingRejectsPreviewOnlyTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-preview-template")
	createCmdbMonitorTargetFixture(t, "host-1003")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)

	for _, templateID := range []string{
		"collect:findx-agent-plugin",
		"alert:host-availability-preview",
		"dashboard:linux-host-basic",
		"metric:host-core",
		"record:dashboard-import-preview",
	} {
		payload := `{
			"hostid":"host-1003",
			"templateid":"` + templateID + `",
			"cmdb_attr_id":"OS001",
			"server_attr_id":"agent.ip"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("template %s status = %d, want %d, body=%s", templateID, w.Code, http.StatusConflict, w.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode blocked response for %s: %v", templateID, err)
		}
		if body["status"] != "PENDING" {
			t.Fatalf("template %s top-level status = %v, want PENDING: %s", templateID, body["status"], w.Body.String())
		}
		assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
			"PENDING",
			"monitor_template_runtime_contract",
			templateID,
		})
		assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"audit_ref":"`})
	}
}

func TestCmdbMonitorBindingRejectsBlockedRuntimeTemplateContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-blocked-runtime-template")
	createCmdbMonitorTargetFixture(t, "host-blocked-template-1001")
	templateID := "collect:cmdb-blocked-runtime-template"
	if _, err := store.SaveMonitoringBuiltinPayload(model.MonitoringBuiltinPayload{
		ID:          templateID,
		ComponentID: "findx-monitor-core",
		Type:        "collect",
		Name:        "blocked runtime template",
		Content: []byte(`{
			"status":"PENDING",
			"reason":"runtime-missing",
			"preview":{"token":"blocked-runtime-secret-marker"}
		}`),
	}, "cmdb-monitor-binding-test"); err != nil {
		t.Fatalf("save blocked runtime template: %v", err)
	}

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"hostid":"host-blocked-template-1001",
		"templateid":"` + templateID + `",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip",
		"tags":{"token":"request-secret-marker"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_template.runtime.v1",
		"cmdb_monitor_template_runtime_content_contract",
		templateID,
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"audit_ref":"`,
		`"cmdb.monitor_binding.save"`,
		"blocked-runtime-secret-marker",
		"request-secret-marker",
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"data_arrived"`,
		`"service_registered"`,
		`"rolled_back"`,
		`"uninstalled"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	})

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusConflict {
		t.Fatalf("read status = %d, want %d, body=%s", getW.Code, http.StatusConflict, getW.Body.String())
	}

	auditResp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
		Source:       model.LogsSourceFindXAudit,
		Scope:        "cmdb",
		ResourceType: "cmdb_monitor_binding",
		ResourceID:   instanceID,
		Action:       "cmdb.monitor_binding.save",
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	if len(auditResp.Items) != 0 {
		t.Fatalf("unexpected audit rows for blocked runtime template: %+v", auditResp.Items)
	}
}

func createCmdbMonitorTargetFixture(t *testing.T, ident string) {
	t.Helper()
	target := &model.MonitorTarget{
		Ident:    ident,
		Name:     "monitor target " + ident,
		IP:       "10.126.0.10",
		Status:   "online",
		Source:   "cmdb-monitor-binding-test",
		Labels:   map[string]string{"env": "test"},
		Metadata: map[string]string{"contract": "monitor_host_binding_contract"},
	}
	if _, err := store.UpsertMonitorTarget(target); err != nil {
		t.Fatalf("create monitor target %s: %v", ident, err)
	}
}

func createCmdbMonitorRuntimeTemplateFixture(t *testing.T, id string) string {
	t.Helper()
	templateID := "collect:" + id
	_, err := store.SaveMonitoringBuiltinPayload(model.MonitoringBuiltinPayload{
		ID:          templateID,
		ComponentID: "findx-monitor-core",
		Type:        "collect",
		Name:        "runtime " + id,
		Content: []byte(`{
			"runtime": {"executor_ref":"findx-agent-config-rollout", "config_snippet_ref":"cmdb-runtime-template"},
			"plugin_id":"input.cpu",
			"config_format":"toml"
		}`),
	}, "cmdb-monitor-binding-test")
	if err != nil {
		t.Fatalf("create monitor runtime template %s: %v", templateID, err)
	}
	return templateID
}
