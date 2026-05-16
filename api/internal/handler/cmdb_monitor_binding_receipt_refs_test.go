package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbMonitorBindingEffectAndRollbackReceiptsReferenceBlockedTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-receipt-refs")
	createCmdbMonitorTargetFixture(t, "host-receipt-refs")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-receipt-refs-template")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"host":"srv-receipt-refs",
		"hostid":"host-receipt-refs",
		"templateid":"` + templateID + `",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip",
		"password":"receipt-refs-secret-marker"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	binding := created["binding"].(map[string]any)
	receipts := binding["receipts"].([]any)
	for _, receiptType := range []string{"delivery", "effect", "rollback"} {
		requestRef := cmdbMonitorBindingReceiptRequestRef(t, receipts, receiptType)
		task, ok, err := store.GetFindXAgentExecutionTask(requestRef)
		if err != nil {
			t.Fatalf("query %s task ref: %v", receiptType, err)
		}
		if !ok || task.Status != "blocked" || task.Metadata["cmdb_binding_id"] != binding["id"] || task.Metadata["cmdb_receipt_type"] != receiptType {
			t.Fatalf("%s task ref should resolve to blocked CMDB execution task, ok=%v task=%#v binding=%#v", receiptType, ok, task, binding)
		}
		assertCmdbTestStringContainsAll(t, w.Body.String(), []string{`"request_ref":"` + requestRef + `"`})
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusConflict {
		t.Fatalf("read status = %d, want %d, body=%s", getW.Code, http.StatusConflict, getW.Body.String())
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
		"receipt-refs-secret-marker",
		`"code":0`,
		`"status":"ready"`,
		`"bindings":[`,
		`"receipts":[`,
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	})
}

func cmdbMonitorBindingReceiptRequestRef(t *testing.T, receipts []any, receiptType string) string {
	t.Helper()
	for _, item := range receipts {
		receipt := item.(map[string]any)
		if receipt["receipt_type"] == receiptType {
			requestRef, _ := receipt["request_ref"].(string)
			requestRef = strings.TrimSpace(requestRef)
			if requestRef == "" {
				t.Fatalf("%s receipt should reference a stored blocked request: %#v", receiptType, receipts)
			}
			return requestRef
		}
	}
	t.Fatalf("%s receipt not found: %#v", receiptType, receipts)
	return ""
}
