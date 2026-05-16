package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbMonitorBindingReceiptQueryPartialRemainsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-receipt-partial")
	binding, err := store.SaveCmdbMonitorBinding(&model.CmdbMonitorBinding{
		InstanceID:   instanceID,
		HostID:       "host-receipt-partial",
		TemplateID:   "collect:findx-agent-plugin",
		CmdbAttrID:   "OS001",
		ServerAttrID: "agent.ip",
		AuditRef:     "cmdb-monitor-binding-receipt-partial-ref",
	})
	if err != nil {
		t.Fatalf("save binding: %v", err)
	}
	if _, err := store.SaveCmdbMonitorBindingReceipt(&model.CmdbMonitorBindingReceipt{
		BindingID:   binding.ID,
		InstanceID:  instanceID,
		ReceiptType: "delivery",
		Status:      "PENDING",
		ContractID:  "cmdb_monitor_binding_delivery_receipt_contract",
		MissingJSON: `["cmdb_monitor_binding_delivery_executor"]`,
		AuditRef:    binding.AuditRef,
	}); err != nil {
		t.Fatalf("save partial receipt: %v", err)
	}

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
		"cmdb_monitor_binding_receipt_complete_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"receipts":[`, `"code":0`, `"ready"`})
}

func TestCmdbMonitorBindingReceiptQueryMissingRequestRefsRemainBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-receipt-ref-missing")
	binding, err := store.SaveCmdbMonitorBinding(&model.CmdbMonitorBinding{
		InstanceID:   instanceID,
		HostID:       "host-receipt-ref-missing",
		TemplateID:   "template-receipt-ref-missing",
		CmdbAttrID:   "OS001",
		ServerAttrID: "agent.ip",
		AuditRef:     "cmdb-monitor-binding-receipt-ref-missing-audit",
	})
	if err != nil {
		t.Fatalf("save binding: %v", err)
	}
	for _, receiptType := range []string{"delivery", "effect", "rollback"} {
		if _, err := store.SaveCmdbMonitorBindingReceipt(&model.CmdbMonitorBindingReceipt{
			BindingID:   binding.ID,
			InstanceID:  instanceID,
			ReceiptType: receiptType,
			Status:      "PENDING",
			ContractID:  "cmdb_monitor_binding_" + receiptType + "_receipt_contract",
			MissingJSON: `["cmdb_monitor_binding_request_ref_contract"]`,
			AuditRef:    "cmdb-monitor-binding-receipt-ref-missing-audit",
		}); err != nil {
			t.Fatalf("save %s receipt: %v", receiptType, err)
		}
	}

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
		"cmdb_monitor_binding_request_ref_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"receipts":[`, `"code":0`, `"ready"`})
}
