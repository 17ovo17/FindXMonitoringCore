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

func TestCmdbMonitorBindingReceiptIngestionUpdatesExistingDeliveryReceipt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID, bindingID, refs := createCmdbMonitorBindingReceiptIngestionFixture(t, "cmdb-monitor-binding-ingest-delivery", "host-ingest-delivery")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)

	payload := `{
		"binding_id":"` + bindingID + `",
		"receipt_type":"delivery",
		"request_ref":"` + refs["delivery"] + `",
		"status":"PENDING",
		"missing_contracts":["cmdb_monitor_binding_delivery_executor","password=ingest-secret-marker"],
		"evidence_ref":"token=ingest-secret-marker"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID+"/"+bindingID+"/receipts", strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.receipt.ingest.v1",
		`"receipt_type":"delivery"`,
		`"request_ref":"` + refs["delivery"] + `"`,
		"cmdb.monitor_binding.delivery.receipt.ingest",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		"ingest-secret-marker",
		`"code":0`,
		`"status":"ready"`,
		`"success"`,
		`"succeeded"`,
		`"delivered"`,
		`"effective"`,
	})

	receipts := store.ListCmdbMonitorBindingReceipts(bindingID)
	if len(receipts) != 3 {
		t.Fatalf("receipt ingestion must update existing rows, got %d receipts: %#v", len(receipts), receipts)
	}
	var delivery model.CmdbMonitorBindingReceipt
	for _, receipt := range receipts {
		if receipt.ReceiptType == "delivery" {
			delivery = receipt
			break
		}
	}
	if delivery.ID == "" || delivery.AuditRef == "" || !strings.Contains(delivery.AuditRef, "ingest") {
		t.Fatalf("delivery receipt should be updated with ingestion audit ref: %#v", delivery)
	}
	if delivery.Status != "PENDING" || delivery.RequestRef != refs["delivery"] {
		t.Fatalf("delivery receipt should stay blocked and keep request_ref: %#v", delivery)
	}
	if strings.Contains(delivery.MissingJSON, "ingest-secret-marker") || strings.Contains(delivery.MissingJSON, "password=") {
		t.Fatalf("delivery receipt missing contracts should not store sensitive payload: %s", delivery.MissingJSON)
	}

	auditResp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
		Source:       model.LogsSourceFindXAudit,
		Scope:        "cmdb",
		ResourceType: "cmdb_monitor_binding",
		ResourceID:   instanceID,
		Action:       "cmdb.monitor_binding.delivery.receipt.ingest",
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("query receipt ingestion audit: %v", err)
	}
	if len(auditResp.Items) == 0 {
		t.Fatalf("receipt ingestion audit log missing: %+v", auditResp)
	}
	for _, row := range auditResp.Items {
		raw, _ := json.Marshal(row.Attributes)
		if strings.Contains(string(raw), "ingest-secret-marker") || strings.Contains(string(raw), "password=") || strings.Contains(string(raw), "token=") {
			t.Fatalf("receipt ingestion audit leaked sensitive evidence: %s", string(raw))
		}
	}
}

func TestCmdbMonitorBindingReceiptIngestionRejectsMismatchedRequestRef(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID, bindingID, refs := createCmdbMonitorBindingReceiptIngestionFixture(t, "cmdb-monitor-binding-ingest-mismatch", "host-ingest-mismatch")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)

	payload := `{
		"binding_id":"` + bindingID + `",
		"receipt_type":"delivery",
		"request_ref":"` + refs["effect"] + `",
		"status":"PENDING"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID+"/"+bindingID+"/receipts", strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.receipt.ingest.v1",
		"cmdb_monitor_binding_request_ref_resolve_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"receipt":{`, `"code":0`, `"status":"ready"`})
}

func TestCmdbMonitorBindingReceiptIngestionRejectsWrongInstanceOrBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID, bindingID, refs := createCmdbMonitorBindingReceiptIngestionFixture(t, "cmdb-monitor-binding-ingest-instance-a", "host-ingest-instance-a")
	otherInstanceID, _, _ := createCmdbMonitorBindingReceiptIngestionFixture(t, "cmdb-monitor-binding-ingest-instance-b", "host-ingest-instance-b")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)

	payload := `{
		"binding_id":"` + bindingID + `",
		"receipt_type":"delivery",
		"request_ref":"` + refs["delivery"] + `",
		"status":"PENDING"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+otherInstanceID+"/"+bindingID+"/receipts", strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{"PENDING", "cmdb_monitor_binding_instance_match_contract"})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{instanceID + `","receipt"`, `"code":0`, `"status":"ready"`})
}

func TestCmdbMonitorBindingReceiptIngestionRejectsUnsupportedStatusAndSensitiveEcho(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID, bindingID, refs := createCmdbMonitorBindingReceiptIngestionFixture(t, "cmdb-monitor-binding-ingest-status", "host-ingest-status")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)

	payload := `{
		"binding_id":"` + bindingID + `",
		"receipt_type":"effect",
		"request_ref":"` + refs["effect"] + `",
		"status":"succeeded",
		"result":{"password":"ingest-status-secret-marker","dsn":"mysql://user:pass@example/db"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID+"/"+bindingID+"/receipts", strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb_monitor_binding_receipt_status_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		"ingest-status-secret-marker",
		"mysql://user:pass@example/db",
		`"succeeded"`,
		`"success"`,
		`"code":0`,
		`"status":"ready"`,
		`"receipt":{`,
	})
}

func TestCmdbMonitorBindingReceiptIngestionIsIdempotent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID, bindingID, refs := createCmdbMonitorBindingReceiptIngestionFixture(t, "cmdb-monitor-binding-ingest-idempotent", "host-ingest-idempotent")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)

	payload := `{
		"binding_id":"` + bindingID + `",
		"receipt_type":"rollback",
		"request_ref":"` + refs["rollback"] + `",
		"status":"PENDING"
	}`
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID+"/"+bindingID+"/receipts", strings.NewReader(payload))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Fatalf("attempt %d status = %d, want %d, body=%s", i+1, w.Code, http.StatusConflict, w.Body.String())
		}
		assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"code":0`, `"status":"ready"`, `"rolled_back"`, `"succeeded"`})
	}
	receipts := store.ListCmdbMonitorBindingReceipts(bindingID)
	if len(receipts) != 3 {
		t.Fatalf("idempotent ingestion must not append duplicate receipts, got %d receipts: %#v", len(receipts), receipts)
	}
}

func createCmdbMonitorBindingReceiptIngestionFixture(t *testing.T, objectID, hostID string) (string, string, map[string]string) {
	t.Helper()
	instanceID := createCmdbCompatibleFixture(t, objectID)
	createCmdbMonitorTargetFixture(t, hostID)
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, objectID+"-template")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"host":"` + hostID + `",
		"hostid":"` + hostID + `",
		"templateid":"` + templateID + `",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip",
		"queue":"` + objectID + `-queue"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create binding status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode binding create response: %v", err)
	}
	binding, _ := created["binding"].(map[string]any)
	bindingID, _ := binding["id"].(string)
	if strings.TrimSpace(bindingID) == "" {
		t.Fatalf("binding id missing from create response: %#v", created)
	}
	receipts := binding["receipts"].([]any)
	refs := map[string]string{}
	for _, receiptType := range []string{"delivery", "effect", "rollback"} {
		refs[receiptType] = cmdbMonitorBindingReceiptRequestRef(t, receipts, receiptType)
	}
	return instanceID, bindingID, refs
}
