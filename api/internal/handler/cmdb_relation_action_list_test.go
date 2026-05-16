package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbRelationActionRequestListKeepsExecutorGapAfterStoredActionsAndReceipts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-list")

	router := gin.New()
	router.POST("/api/v1/cmdb/relation-actions", CreateCmdbRelationActionRequest)
	router.GET("/api/v1/cmdb/relation-actions", ListCmdbRelationActionRequests)

	payload := `{
		"action":"detail",
		"instance_id":"` + rootID + `",
		"node_id":"` + targetID + `",
		"object_id":"cmdb-relation-action-list-db",
		"relation_id":"cmdb-relation-action-list-rel",
		"context":{"source":"list query","token":"list-secret-marker"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/relation-actions", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("create status = %d, want %d, body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions?instance_id="+rootID, nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusConflict {
		t.Fatalf("list status = %d, want %d, body=%s", listW.Code, http.StatusConflict, listW.Body.String())
	}
	assertCmdbTestStringContainsAll(t, listW.Body.String(), []string{
		"PENDING",
		"cmdb.relation_action.runtime.read.v1",
		"cmdb_relation_action_store_contract",
		"cmdb_relation_action_action_executor_contract",
		"cmdb_relation_action_delivery_executor_contract",
		"cmdb_relation_action_effect_executor_contract",
		"cmdb_relation_action_attested_receipt_contract",
	})
	assertCmdbTestStringExcludesAll(t, listW.Body.String(), []string{
		"list-secret-marker",
		`"code":0`,
		`"action_requests":[`,
		`"receipts":[`,
		`"status":"ready"`,
		`"status":"recorded"`,
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"data_arrived"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
		`"success"`,
	})
}

func TestCmdbRelationActionRequestListEmptyRemainsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbDanglingRelationFixture(t, "cmdb-relation-action-list-empty")

	router := gin.New()
	router.GET("/api/v1/cmdb/relation-actions", ListCmdbRelationActionRequests)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions?instance_id="+instanceID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.relation_actions.read.v1",
		"cmdb_relation_action_store",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"action_requests":[`, `"code":0`, `"ready"`})
}

func TestCmdbRelationActionRequestListPartialReceiptsRemainBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-list-partial")
	item, err := store.SaveCmdbRelationActionRequest(&model.CmdbRelationActionRequest{
		Action:         "relation",
		InstanceID:     rootID,
		NodeID:         targetID,
		ObjectID:       "cmdb-relation-action-list-partial-db",
		RelationID:     "cmdb-relation-action-list-partial-rel",
		Actor:          "tester",
		Status:         "recorded",
		DeliveryStatus: "PENDING",
		EffectStatus:   "PENDING",
		ContextJSON:    `{"source":"list partial","secret":"list-partial-secret-marker"}`,
		AuditRef:       "cmdb-relation-action-list-partial-ref",
	})
	if err != nil {
		t.Fatalf("save action request: %v", err)
	}
	if _, err := store.SaveCmdbRelationActionReceipt(&model.CmdbRelationActionReceipt{
		ActionRequestID: item.ID,
		Action:          "relation",
		InstanceID:      rootID,
		NodeID:          targetID,
		RelationID:      "cmdb-relation-action-list-partial-rel",
		ReceiptType:     "delivery",
		Status:          "PENDING",
		ContractID:      "cmdb_relation_action_delivery_receipt_contract",
		MissingJSON:     `["cmdb_relation_action_delivery_executor"]`,
		AuditRef:        "cmdb-relation-action-list-partial-ref-delivery",
	}); err != nil {
		t.Fatalf("save partial receipt: %v", err)
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/relation-actions", ListCmdbRelationActionRequests)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions?instance_id="+rootID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.relation_actions.read.v1",
		"cmdb_relation_action_receipt_complete_contract",
		"cmdb_relation_action_effect_receipt_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"code":0`,
		`"action_requests":[`,
		`"status":"ready"`,
		"list-partial-secret-marker",
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	})
}
