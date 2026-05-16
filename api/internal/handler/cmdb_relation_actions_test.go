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

func TestCmdbRelationActionRequestCreatesBlockedReceiptsAndAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-receipt")

	router := gin.New()
	router.POST("/api/v1/cmdb/relation-actions", CreateCmdbRelationActionRequest)

	payload := `{
		"action":"topology",
		"instance_id":"` + rootID + `",
		"node_id":"` + targetID + `",
		"object_id":"cmdb-relation-action-receipt-db",
		"relation_id":"cmdb-relation-action-receipt-rel",
		"context":{"note":"receipt audit","secret":"secret-marker"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/relation-actions", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	action := body["action_request"].(map[string]any)
	receipts := action["receipts"].([]any)
	if len(receipts) != 2 {
		t.Fatalf("receipts len = %d, want 2: %#v", len(receipts), receipts)
	}
	if action["delivery_status"] != "PENDING" || action["effect_status"] != "PENDING" {
		t.Fatalf("action receipt status mismatch: %#v", action)
	}
	deliveryRequestRef := ""
	effectRequestRef := ""
	for _, raw := range receipts {
		receipt := raw.(map[string]any)
		if receipt["receipt_type"] == "delivery" {
			deliveryRequestRef = strings.TrimSpace(receipt["request_ref"].(string))
		}
		if receipt["receipt_type"] == "effect" {
			effectRequestRef = strings.TrimSpace(receipt["request_ref"].(string))
		}
	}
	if deliveryRequestRef == "" {
		t.Fatalf("delivery receipt request_ref is empty: %#v", receipts)
	}
	if effectRequestRef == "" {
		t.Fatalf("effect receipt request_ref is empty: %#v", receipts)
	}
	deliveryTask, ok, err := store.GetFindXAgentExecutionTask(deliveryRequestRef)
	if err != nil {
		t.Fatalf("get delivery task: %v", err)
	}
	if !ok {
		t.Fatalf("delivery request_ref %q does not resolve to a blocked execution task", deliveryRequestRef)
	}
	if deliveryTask.Status != "blocked" {
		t.Fatalf("delivery task status = %q, want blocked: %#v", deliveryTask.Status, deliveryTask)
	}
	if deliveryTask.Metadata["cmdb_relation_action_id"] != action["id"] ||
		deliveryTask.Metadata["cmdb_relation_id"] != "cmdb-relation-action-receipt-rel" ||
		deliveryTask.Metadata["cmdb_instance_id"] != rootID ||
		deliveryTask.Metadata["cmdb_node_id"] != targetID ||
		deliveryTask.Metadata["cmdb_relation_action"] != "topology" {
		t.Fatalf("delivery task metadata mismatch: %#v", deliveryTask.Metadata)
	}
	effectTask, ok, err := store.GetFindXAgentExecutionTask(effectRequestRef)
	if err != nil {
		t.Fatalf("get effect task: %v", err)
	}
	if !ok {
		t.Fatalf("effect request_ref %q does not resolve to a blocked execution task", effectRequestRef)
	}
	if effectTask.Status != "blocked" {
		t.Fatalf("effect task status = %q, want blocked: %#v", effectTask.Status, effectTask)
	}
	if effectTask.Metadata["cmdb_relation_action_id"] != action["id"] ||
		effectTask.Metadata["cmdb_relation_id"] != "cmdb-relation-action-receipt-rel" ||
		effectTask.Metadata["cmdb_instance_id"] != rootID ||
		effectTask.Metadata["cmdb_node_id"] != targetID ||
		effectTask.Metadata["cmdb_relation_action"] != "topology" {
		t.Fatalf("effect task metadata mismatch: %#v", effectTask.Metadata)
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		`"receipts"`,
		`"request_ref":"` + deliveryRequestRef + `"`,
		`"request_ref":"` + effectRequestRef + `"`,
		`"receipt_type":"delivery"`,
		`"receipt_type":"effect"`,
		`"cmdb.relation.topology.delivery.blocked"`,
		`"cmdb.relation.topology.effect.blocked"`,
		`"cmdb_relation_action_delivery_executor"`,
		`"cmdb_relation_action_effect_probe"`,
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		"secret-marker",
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"data_arrived"`,
		`"success":true`,
	})

	requestID := action["id"].(string)
	stored := store.ListCmdbRelationActionReceipts(requestID)
	if len(stored) != 2 {
		t.Fatalf("stored receipts len = %d, want 2: %#v", len(stored), stored)
	}

	for _, auditAction := range []string{
		"cmdb.relation.topology.delivery.blocked",
		"cmdb.relation.topology.effect.blocked",
	} {
		auditResp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
			Source:       model.LogsSourceFindXAudit,
			Scope:        "cmdb",
			ResourceType: "cmdb_relation_action",
			ResourceID:   rootID,
			Action:       auditAction,
			Limit:        10,
		})
		if err != nil {
			t.Fatalf("query audit logs %s: %v", auditAction, err)
		}
		if len(auditResp.Items) == 0 {
			t.Fatalf("missing audit action %s: %+v", auditAction, auditResp)
		}
	}
}

func TestCmdbRelationActionReceiptQueryKeepsExecutorGapAfterStoredReceipts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-receipt-query")

	router := gin.New()
	router.POST("/api/v1/cmdb/relation-actions", CreateCmdbRelationActionRequest)
	router.GET("/api/v1/cmdb/relation-actions/:id/receipts", GetCmdbRelationActionReceipts)

	payload := `{
		"action":"expand",
		"instance_id":"` + rootID + `",
		"node_id":"` + targetID + `",
		"object_id":"cmdb-relation-action-receipt-query-db",
		"relation_id":"cmdb-relation-action-receipt-query-rel",
		"context":{"source":"receipt query","secret":"receipt-query-secret-marker"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/relation-actions", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("create status = %d, want %d, body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	requestID := created["action_request"].(map[string]any)["id"].(string)

	receiptsReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions/"+requestID+"/receipts", nil)
	receiptsW := httptest.NewRecorder()
	router.ServeHTTP(receiptsW, receiptsReq)
	if receiptsW.Code != http.StatusConflict {
		t.Fatalf("receipt status = %d, want %d, body=%s", receiptsW.Code, http.StatusConflict, receiptsW.Body.String())
	}
	assertCmdbTestStringContainsAll(t, receiptsW.Body.String(), []string{
		"PENDING",
		"cmdb.relation_action.runtime.read.v1",
		"cmdb_relation_action_store_contract",
		"cmdb_relation_action_action_executor_contract",
		"cmdb_relation_action_delivery_executor_contract",
		"cmdb_relation_action_effect_executor_contract",
		"cmdb_relation_action_attested_receipt_contract",
	})
	assertCmdbTestStringExcludesAll(t, receiptsW.Body.String(), []string{
		"receipt-query-secret-marker",
		`"code":0`,
		`"receipts":[`,
		`"status":"ready"`,
		`"status":"recorded"`,
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
		`"success"`,
	})
}

func TestCmdbRelationActionAuditQueryReturnsFindXAuditRows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-audit-query")

	router := gin.New()
	router.POST("/api/v1/cmdb/relation-actions", CreateCmdbRelationActionRequest)
	router.GET("/api/v1/cmdb/relation-actions/:id", GetCmdbRelationActionRequest)

	payload := `{
		"action":"relation",
		"instance_id":"` + rootID + `",
		"node_id":"` + targetID + `",
		"object_id":"cmdb-relation-action-audit-query-db",
		"relation_id":"cmdb-relation-action-audit-query-rel",
		"context":{"source":"audit query","secret":"audit-query-secret-marker"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/relation-actions", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("create status = %d, want %d, body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	requestID := created["action_request"].(map[string]any)["id"].(string)

	auditReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions/"+requestID+"?include=audit", nil)
	auditW := httptest.NewRecorder()
	router.ServeHTTP(auditW, auditReq)
	if auditW.Code != http.StatusOK {
		t.Fatalf("audit status = %d, want %d, body=%s", auditW.Code, http.StatusOK, auditW.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(auditW.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode audit response: %v", err)
	}
	if body["contract"] != "cmdb.relation_action.audit.read.v1" {
		t.Fatalf("contract = %v, want relation action audit contract: %s", body["contract"], auditW.Body.String())
	}
	records := body["audit_logs"].([]any)
	if len(records) != 3 {
		t.Fatalf("audit_logs len = %d, want 3: %#v", len(records), records)
	}
	assertCmdbTestStringContainsAll(t, auditW.Body.String(), []string{
		`"status":"ready"`,
		`"source":"findx_audit"`,
		`"cmdb.relation.relation.request"`,
		`"cmdb.relation.relation.delivery.blocked"`,
		`"cmdb.relation.relation.effect.blocked"`,
	})
	assertCmdbTestStringExcludesAll(t, auditW.Body.String(), []string{
		"audit-query-secret-marker",
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"delivered"`,
		`"effective"`,
	})
}

func TestCmdbRelationActionAuditQueryEmptyRemainsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-audit-empty")
	item, err := store.SaveCmdbRelationActionRequest(&model.CmdbRelationActionRequest{
		Action:         "detail",
		InstanceID:     rootID,
		NodeID:         targetID,
		ObjectID:       "cmdb-relation-action-audit-empty-db",
		RelationID:     "cmdb-relation-action-audit-empty-rel",
		Actor:          "tester",
		Status:         "recorded",
		DeliveryStatus: "PENDING",
		EffectStatus:   "PENDING",
		AuditRef:       "cmdb-relation-action-audit-empty-ref",
	})
	if err != nil {
		t.Fatalf("save action request: %v", err)
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/relation-actions/:id", GetCmdbRelationActionRequest)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions/"+item.ID+"?include=audit", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.relation_action.audit.read.v1",
		"cmdb_relation_action_audit_log_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"audit_logs":[`, `"code":0`, `"ready"`})
}

func TestCmdbRelationActionDetailReadPartialReceiptsRemainBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-detail-partial")
	item, err := store.SaveCmdbRelationActionRequest(&model.CmdbRelationActionRequest{
		Action:         "detail",
		InstanceID:     rootID,
		NodeID:         targetID,
		ObjectID:       "cmdb-relation-action-detail-partial-db",
		RelationID:     "cmdb-relation-action-detail-partial-rel",
		Actor:          "tester",
		Status:         "recorded",
		DeliveryStatus: "PENDING",
		EffectStatus:   "PENDING",
		ContextJSON:    `{"source":"detail partial","secret":"detail-partial-secret-marker"}`,
		AuditRef:       "cmdb-relation-action-detail-partial-ref",
	})
	if err != nil {
		t.Fatalf("save action request: %v", err)
	}
	if _, err := store.SaveCmdbRelationActionReceipt(&model.CmdbRelationActionReceipt{
		ActionRequestID: item.ID,
		Action:          "detail",
		InstanceID:      rootID,
		NodeID:          targetID,
		RelationID:      "cmdb-relation-action-detail-partial-rel",
		ReceiptType:     "delivery",
		Status:          "PENDING",
		ContractID:      "cmdb_relation_action_delivery_receipt_contract",
		MissingJSON:     `["cmdb_relation_action_delivery_executor"]`,
		AuditRef:        "cmdb-relation-action-detail-partial-ref-delivery",
	}); err != nil {
		t.Fatalf("save partial receipt: %v", err)
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/relation-actions/:id", GetCmdbRelationActionRequest)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions/"+item.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.relation_action.detail.read.v1",
		"cmdb_relation_action_receipt_complete_contract",
		"cmdb_relation_action_effect_receipt_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"code":0`,
		`"status":"recorded"`,
		`"action_request":{`,
		"detail-partial-secret-marker",
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	})
}

func TestCmdbRelationActionDetailReadKeepsExecutorGapAfterCompleteStoredReceipts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-detail-read")

	router := gin.New()
	router.POST("/api/v1/cmdb/relation-actions", CreateCmdbRelationActionRequest)
	router.GET("/api/v1/cmdb/relation-actions/:id", GetCmdbRelationActionRequest)

	payload := `{
		"action":"topology",
		"instance_id":"` + rootID + `",
		"node_id":"` + targetID + `",
		"object_id":"cmdb-relation-action-detail-read-db",
		"relation_id":"cmdb-relation-action-detail-read-rel",
		"context":{"source":"detail read","secret":"detail-read-secret-marker"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/relation-actions", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("create status = %d, want %d, body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	requestID := created["action_request"].(map[string]any)["id"].(string)

	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions/"+requestID, nil)
	detailW := httptest.NewRecorder()
	router.ServeHTTP(detailW, detailReq)
	if detailW.Code != http.StatusConflict {
		t.Fatalf("detail status = %d, want %d, body=%s", detailW.Code, http.StatusConflict, detailW.Body.String())
	}
	assertCmdbTestStringContainsAll(t, detailW.Body.String(), []string{
		"PENDING",
		"cmdb.relation_action.runtime.read.v1",
		"cmdb_relation_action_store_contract",
		"cmdb_relation_action_action_executor_contract",
		"cmdb_relation_action_delivery_executor_contract",
		"cmdb_relation_action_effect_executor_contract",
		"cmdb_relation_action_attested_receipt_contract",
	})
	assertCmdbTestStringExcludesAll(t, detailW.Body.String(), []string{
		"detail-read-secret-marker",
		`"code":0`,
		`"action_request":{`,
		`"receipts":[`,
		`"status":"ready"`,
		`"status":"recorded"`,
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
		`"success"`,
	})
}
