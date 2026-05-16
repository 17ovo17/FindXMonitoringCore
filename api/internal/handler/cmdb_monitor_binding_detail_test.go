package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCmdbMonitorBindingDetailKeepsExecutorGapAfterStoredBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-detail")
	createCmdbMonitorTargetFixture(t, "host-detail-1001")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-detail-template")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"host":"srv-detail",
		"hostid":"host-detail-1001",
		"templateid":"` + templateID + `",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip",
		"tags":{"env":"prod","secret":"detail-secret-marker"}
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusOK {
		t.Fatalf("create status = %d, want %d, body=%s", createW.Code, http.StatusOK, createW.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(createW.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	bindingID := created["binding"].(map[string]any)["id"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID+"/"+bindingID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("detail status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.runtime.receipts.read.v1",
		"cmdb_monitor_binding_delivery_executor_contract",
		"cmdb_monitor_binding_effect_executor_contract",
		"cmdb_monitor_binding_rollback_executor_contract",
		"cmdb_monitor_binding_attested_receipt_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		"detail-secret-marker",
		`"code":0`,
		`"status":"ready"`,
		`"binding":{`,
		`"receipts":[`,
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"data_arrived"`,
		`"rolled_back"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
		`"success"`,
	})
}

func TestCmdbMonitorBindingDetailWrongInstanceRemainsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-detail-owner")
	otherID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-detail-other")
	createCmdbMonitorTargetFixture(t, "host-detail-1002")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-detail-owner-template")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"hostid":"host-detail-1002",
		"templateid":"` + templateID + `",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip"
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusOK {
		t.Fatalf("create status = %d, want %d, body=%s", createW.Code, http.StatusOK, createW.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(createW.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	bindingID := created["binding"].(map[string]any)["id"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+otherID+"/"+bindingID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.detail.read.v1",
		"cmdb_monitor_binding_instance_match_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"binding":{`, `"code":0`, `"ready"`})
}
