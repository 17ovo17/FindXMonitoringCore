package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbRelationActionRuntimeReadBlocksMissingRequestRefTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-runtime-missing")
	item := saveCmdbRelationActionRuntimeFixture(t, rootID, targetID, "topology", "cmdb-relation-action-runtime-missing-rel")
	saveCmdbRelationActionRuntimeReceipt(t, item, "delivery", "cmdb-relation-action-runtime-missing-delivery-task")
	saveCmdbRelationActionRuntimeReceipt(t, item, "effect", "cmdb-relation-action-runtime-missing-effect-task")

	router := gin.New()
	router.GET("/api/v1/cmdb/relation-actions", ListCmdbRelationActionRequests)
	router.GET("/api/v1/cmdb/relation-actions/:id", GetCmdbRelationActionRequest)
	router.GET("/api/v1/cmdb/relation-actions/:id/receipts", GetCmdbRelationActionReceipts)

	for _, path := range []string{
		"/api/v1/cmdb/relation-actions/" + item.ID,
		"/api/v1/cmdb/relation-actions/" + item.ID + "/receipts",
		"/api/v1/cmdb/relation-actions?instance_id=" + rootID,
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Fatalf("%s status = %d, want %d, body=%s", path, w.Code, http.StatusConflict, w.Body.String())
		}
		assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
			"PENDING",
			"cmdb.relation_action.runtime.read.v1",
			"cmdb_relation_action_request_ref_resolve_contract",
			"cmdb_relation_action_execution_task_contract",
		})
		assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
			`"code":0`,
			`"status":"ready"`,
			`"status":"recorded"`,
			`"action_request":{`,
			`"receipts":[`,
			`"action_requests":[`,
			`"queued"`,
			`"running"`,
			`"delivered"`,
			`"effective"`,
			`"succeeded"`,
			"runtime-secret-marker",
		})
	}
}

func TestCmdbRelationActionRuntimeReadBlocksMismatchedRequestRefTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-runtime-mismatch")
	item := saveCmdbRelationActionRuntimeFixture(t, rootID, targetID, "relation", "cmdb-relation-action-runtime-mismatch-rel")
	deliveryTask := saveCmdbRelationActionRuntimeTask(t, item, "delivery", map[string]string{
		"cmdb_relation_id": "wrong-relation-id",
	})
	effectTask := saveCmdbRelationActionRuntimeTask(t, item, "effect", nil)
	saveCmdbRelationActionRuntimeReceipt(t, item, "delivery", deliveryTask.ID)
	saveCmdbRelationActionRuntimeReceipt(t, item, "effect", effectTask.ID)

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
		"cmdb.relation_action.runtime.read.v1",
		"cmdb_relation_action_execution_task_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"action_request":{`,
		`"receipts":[`,
		`"code":0`,
		`"queued"`,
		`"running"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	})
}

func TestCmdbRelationActionRuntimeReadBlocksWrongExecutionRequestRefTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-runtime-wrong-action")
	item := saveCmdbRelationActionRuntimeFixture(t, rootID, targetID, "topology", "cmdb-relation-action-runtime-wrong-action-rel")
	deliveryTask := saveCmdbRelationActionRuntimeTask(t, item, "delivery", map[string]string{
		"cmdb_relation_action": "wrong-action",
	})
	effectTask := saveCmdbRelationActionRuntimeTask(t, item, "effect", nil)
	saveCmdbRelationActionRuntimeReceipt(t, item, "delivery", deliveryTask.ID)
	saveCmdbRelationActionRuntimeReceipt(t, item, "effect", effectTask.ID)

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
		"cmdb.relation_action.runtime.read.v1",
		"cmdb_relation_action_execution_task_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"receipts":[`,
		`"code":0`,
		`"succeeded"`,
	})
}

func TestCmdbRelationActionRuntimeReadKeepsBlockedAfterCompleteReceiptsWithBlockedTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-runtime-resolved")
	item := saveCmdbRelationActionRuntimeFixture(t, rootID, targetID, "topology", "cmdb-relation-action-runtime-resolved-rel")
	deliveryTask := saveCmdbRelationActionRuntimeTask(t, item, "delivery", nil)
	effectTask := saveCmdbRelationActionRuntimeTask(t, item, "effect", nil)
	saveCmdbRelationActionRuntimeReceipt(t, item, "delivery", deliveryTask.ID)
	saveCmdbRelationActionRuntimeReceipt(t, item, "effect", effectTask.ID)

	router := gin.New()
	router.GET("/api/v1/cmdb/relation-actions", ListCmdbRelationActionRequests)
	router.GET("/api/v1/cmdb/relation-actions/:id", GetCmdbRelationActionRequest)
	router.GET("/api/v1/cmdb/relation-actions/:id/receipts", GetCmdbRelationActionReceipts)

	for _, path := range []string{
		"/api/v1/cmdb/relation-actions/" + item.ID,
		"/api/v1/cmdb/relation-actions/" + item.ID + "/receipts",
		"/api/v1/cmdb/relation-actions?instance_id=" + rootID,
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Fatalf("%s status = %d, want %d, body=%s", path, w.Code, http.StatusConflict, w.Body.String())
		}
		assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
			"PENDING",
			"cmdb.relation_action.runtime.read.v1",
			"cmdb_relation_action_action_executor_contract",
			"cmdb_relation_action_attested_receipt_contract",
		})
		assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
			`"code":0`,
			`"status":"ready"`,
			`"status":"recorded"`,
			`"action_request":{`,
			`"receipts":[`,
			`"action_requests":[`,
			`"queued"`,
			`"running"`,
			`"delivered"`,
			`"effective"`,
			`"succeeded"`,
			`"success"`,
			"runtime-secret-marker",
		})
	}
}

func TestCmdbRelationActionRuntimeReadKeepsActionExecutorGapAfterBlockedReceipts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-executor-gap")
	item := saveCmdbRelationActionRuntimeFixture(t, rootID, targetID, "topology", "cmdb-relation-action-executor-gap-rel")
	deliveryTask := saveCmdbRelationActionRuntimeTask(t, item, "delivery", nil)
	effectTask := saveCmdbRelationActionRuntimeTask(t, item, "effect", nil)
	saveCmdbRelationActionRuntimeReceipt(t, item, "delivery", deliveryTask.ID)
	saveCmdbRelationActionRuntimeReceipt(t, item, "effect", effectTask.ID)

	router := gin.New()
	router.GET("/api/v1/cmdb/relation-actions", ListCmdbRelationActionRequests)
	router.GET("/api/v1/cmdb/relation-actions/:id", GetCmdbRelationActionRequest)
	router.GET("/api/v1/cmdb/relation-actions/:id/receipts", GetCmdbRelationActionReceipts)

	for _, path := range []string{
		"/api/v1/cmdb/relation-actions/" + item.ID,
		"/api/v1/cmdb/relation-actions/" + item.ID + "/receipts",
		"/api/v1/cmdb/relation-actions?instance_id=" + rootID,
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Fatalf("%s status = %d, want %d, body=%s", path, w.Code, http.StatusConflict, w.Body.String())
		}
		assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
			"PENDING",
			"cmdb.relation_action.runtime.read.v1",
			"cmdb_relation_action_store_contract",
			"cmdb_relation_action_action_executor_contract",
			"cmdb_relation_action_delivery_executor_contract",
			"cmdb_relation_action_effect_executor_contract",
			"cmdb_relation_action_attested_receipt_contract",
		})
		assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
			`"code":0`,
			`"status":"ready"`,
			`"status":"recorded"`,
			`"action_request":{`,
			`"receipts":[`,
			`"action_requests":[`,
			`"actions":[]`,
			`"queued"`,
			`"running"`,
			`"applied"`,
			`"installed"`,
			`"data_arrived"`,
			`"delivered"`,
			`"effective"`,
			`"succeeded"`,
			`"success"`,
			`"imported"`,
			"runtime-secret-marker",
			"password",
			"token",
			"cookie",
			"private_key",
			"mysql://",
			"postgres://",
			"Nightingale",
			"SkyWalking",
			"SigNoZ",
			"Categraf",
			"Catpaw",
			"Grafana",
			"Prometheus",
		})
	}
}

func saveCmdbRelationActionRuntimeFixture(t *testing.T, rootID, targetID, action, relationID string) model.CmdbRelationActionRequest {
	t.Helper()
	item, err := store.SaveCmdbRelationActionRequest(&model.CmdbRelationActionRequest{
		Action:         action,
		InstanceID:     rootID,
		NodeID:         targetID,
		ObjectID:       relationID + "-object",
		RelationID:     relationID,
		Actor:          "tester",
		Status:         "recorded",
		DeliveryStatus: "PENDING",
		EffectStatus:   "PENDING",
		ContextJSON:    `{"source":"runtime read","secret":"runtime-secret-marker"}`,
		AuditRef:       relationID + "-audit",
	})
	if err != nil {
		t.Fatalf("save relation action request: %v", err)
	}
	return *item
}

func saveCmdbRelationActionRuntimeReceipt(t *testing.T, item model.CmdbRelationActionRequest, receiptType, requestRef string) {
	t.Helper()
	contract := "cmdb_relation_action_" + receiptType + "_receipt_contract"
	if _, err := store.SaveCmdbRelationActionReceipt(&model.CmdbRelationActionReceipt{
		ActionRequestID: item.ID,
		Action:          item.Action,
		InstanceID:      item.InstanceID,
		NodeID:          item.NodeID,
		RelationID:      item.RelationID,
		ReceiptType:     receiptType,
		Status:          "PENDING",
		ContractID:      contract,
		MissingJSON:     `["cmdb_relation_action_` + receiptType + `_executor"]`,
		RequestRef:      requestRef,
		AuditRef:        item.AuditRef,
	}); err != nil {
		t.Fatalf("save %s receipt: %v", receiptType, err)
	}
}

func saveCmdbRelationActionRuntimeTask(t *testing.T, item model.CmdbRelationActionRequest, receiptType string, overrides map[string]string) model.FindXAgentExecutionTask {
	t.Helper()
	metadata := map[string]string{
		"scope":                   "cmdb_relation_action",
		"cmdb_relation_action_id": item.ID,
		"cmdb_relation_action":    item.Action,
		"cmdb_receipt_type":       receiptType,
		"cmdb_instance_id":        item.InstanceID,
		"cmdb_node_id":            item.NodeID,
		"cmdb_object_id":          item.ObjectID,
		"cmdb_relation_id":        item.RelationID,
	}
	for key, value := range overrides {
		metadata[key] = value
	}
	task, err := store.SaveFindXAgentExecutionTask(model.FindXAgentExecutionTask{
		Action:    "cmdb-relation-" + item.Action + "-" + receiptType,
		TargetIDs: []string{item.NodeID},
		Status:    "blocked",
		Blocker:   "PENDING: test runtime read",
		Audit:     "findx_agent.task.requested",
		Metadata:  metadata,
	})
	if err != nil {
		t.Fatalf("save %s runtime task: %v", receiptType, err)
	}
	return task
}
