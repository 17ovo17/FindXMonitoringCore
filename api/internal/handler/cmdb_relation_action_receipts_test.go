package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbRelationActionReceiptQueryPartialRemainsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-receipt-partial")
	item, err := store.SaveCmdbRelationActionRequest(&model.CmdbRelationActionRequest{
		Action:         "topology",
		InstanceID:     rootID,
		NodeID:         targetID,
		ObjectID:       "cmdb-relation-action-receipt-partial-db",
		RelationID:     "cmdb-relation-action-receipt-partial-rel",
		Actor:          "tester",
		Status:         "recorded",
		DeliveryStatus: "PENDING",
		EffectStatus:   "PENDING",
		AuditRef:       "cmdb-relation-action-receipt-partial-ref",
	})
	if err != nil {
		t.Fatalf("save action request: %v", err)
	}
	if _, err := store.SaveCmdbRelationActionReceipt(&model.CmdbRelationActionReceipt{
		ActionRequestID: item.ID,
		Action:          "topology",
		InstanceID:      rootID,
		NodeID:          targetID,
		RelationID:      "cmdb-relation-action-receipt-partial-rel",
		ReceiptType:     "delivery",
		Status:          "PENDING",
		ContractID:      "cmdb_relation_action_delivery_receipt_contract",
		MissingJSON:     `["cmdb_relation_action_delivery_executor"]`,
		AuditRef:        item.AuditRef,
	}); err != nil {
		t.Fatalf("save partial receipt: %v", err)
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/relation-actions/:id/receipts", GetCmdbRelationActionReceipts)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions/"+item.ID+"/receipts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.relation_action.receipts.read.v1",
		"cmdb_relation_action_receipt_complete_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"receipts":[`, `"code":0`, `"ready"`, "success"})
}

func TestCmdbRelationActionReceiptQueryMissingRequestRefsRemainBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-receipt-ref-missing")
	item, err := store.SaveCmdbRelationActionRequest(&model.CmdbRelationActionRequest{
		Action:         "topology",
		InstanceID:     rootID,
		NodeID:         targetID,
		ObjectID:       "cmdb-relation-action-receipt-ref-missing-db",
		RelationID:     "cmdb-relation-action-receipt-ref-missing-rel",
		Actor:          "tester",
		Status:         "recorded",
		DeliveryStatus: "PENDING",
		EffectStatus:   "PENDING",
		AuditRef:       "cmdb-relation-action-receipt-ref-missing-ref",
	})
	if err != nil {
		t.Fatalf("save action request: %v", err)
	}
	for _, receiptType := range []string{"delivery", "effect"} {
		if _, err := store.SaveCmdbRelationActionReceipt(&model.CmdbRelationActionReceipt{
			ActionRequestID: item.ID,
			Action:          "topology",
			InstanceID:      rootID,
			NodeID:          targetID,
			RelationID:      "cmdb-relation-action-receipt-ref-missing-rel",
			ReceiptType:     receiptType,
			Status:          "PENDING",
			ContractID:      "cmdb_relation_action_" + receiptType + "_receipt_contract",
			MissingJSON:     `["cmdb_relation_action_request_ref_contract"]`,
			AuditRef:        item.AuditRef,
		}); err != nil {
			t.Fatalf("save %s receipt: %v", receiptType, err)
		}
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/relation-actions/:id/receipts", GetCmdbRelationActionReceipts)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions/"+item.ID+"/receipts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.relation_action.receipts.read.v1",
		"cmdb_relation_action_request_ref_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"receipts":[`, `"code":0`, `"ready"`})
}
