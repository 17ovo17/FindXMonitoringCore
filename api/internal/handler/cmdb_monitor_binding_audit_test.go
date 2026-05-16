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

func TestCmdbMonitorBindingAuditQueryReturnsFindXAuditRows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-audit-query")
	createCmdbMonitorTargetFixture(t, "host-audit-1001")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-audit-template")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"host":"srv-audit",
		"hostid":"host-audit-1001",
		"templateid":"` + templateID + `",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip",
		"tags":{"env":"prod","secret":"audit-token-marker"},
		"password":"audit-secret-marker"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("write status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	auditReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID+"?include=audit", nil)
	auditW := httptest.NewRecorder()
	router.ServeHTTP(auditW, auditReq)
	if auditW.Code != http.StatusOK {
		t.Fatalf("audit status = %d, want %d, body=%s", auditW.Code, http.StatusOK, auditW.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(auditW.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode audit response: %v", err)
	}
	if body["contract"] != "cmdb.monitor_binding.audit.read.v1" {
		t.Fatalf("contract = %v, want monitor binding audit contract: %s", body["contract"], auditW.Body.String())
	}
	records := body["audit_logs"].([]any)
	if len(records) != 4 {
		t.Fatalf("audit_logs len = %d, want 4: %#v", len(records), records)
	}
	assertCmdbTestStringContainsAll(t, auditW.Body.String(), []string{
		`"status":"ready"`,
		`"source":"findx_audit"`,
		`"cmdb.monitor_binding.save"`,
		`"cmdb.monitor_binding.delivery.blocked"`,
		`"cmdb.monitor_binding.effect.blocked"`,
		`"cmdb.monitor_binding.rollback.blocked"`,
	})
	assertCmdbTestStringExcludesAll(t, auditW.Body.String(), []string{
		"audit-token-marker",
		"audit-secret-marker",
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"data_arrived"`,
		`"rolled_back"`,
		`"delivered"`,
		`"effective"`,
	})
}

func TestCmdbMonitorBindingAuditQueryEmptyRemainsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-audit-empty")
	binding, err := store.SaveCmdbMonitorBinding(&model.CmdbMonitorBinding{
		InstanceID:   instanceID,
		HostID:       "host-audit-empty",
		TemplateID:   "collect:findx-agent-plugin",
		CmdbAttrID:   "OS001",
		ServerAttrID: "agent.ip",
		AuditRef:     "cmdb-monitor-binding-audit-empty-ref",
	})
	if err != nil {
		t.Fatalf("save binding: %v", err)
	}
	if binding.ID == "" {
		t.Fatalf("binding id should be assigned")
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID+"?include=audit", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.audit.read.v1",
		"cmdb_monitor_binding_audit_log_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"audit_logs":[`, `"code":0`, `"ready"`})
}
