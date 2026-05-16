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

func TestCmdbRelationActionBlockedTasksCarryMatureTaskLogShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prefix := "cmdb-relation-action-task-log"
	rootID, targetID := createCmdbRelationDescriptorFixture(t, prefix)

	router := gin.New()
	router.POST("/api/v1/cmdb/relation-actions", CreateCmdbRelationActionRequest)

	payload := `{
		"action":"topology",
		"instance_id":"` + rootID + `",
		"node_id":"` + targetID + `",
		"object_id":"` + prefix + `-db",
		"relation_id":"` + prefix + `-rel",
		"context":{"source":"task-log-shape","secret":"relation-task-log-secret-marker"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/relation-actions", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("create status = %d, want %d, body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}

	action, receipts := cmdbRelationActionCreatedRequestAndReceipts(t, w.Body.String())
	for _, receiptType := range []string{"delivery", "effect"} {
		ref := cmdbRelationActionReceiptRequestRef(t, receipts, receiptType)
		task, ok, err := store.GetFindXAgentExecutionTask(ref)
		if err != nil {
			t.Fatalf("query %s task ref: %v", receiptType, err)
		}
		if !ok {
			t.Fatalf("%s task ref %q was not stored", receiptType, ref)
		}
		for _, want := range []string{
			"cmdb_task_log_contract",
			"cmdb_task_log_status",
			"cmdb_task_log_type",
			"cmdb_task_log_data",
			"cmdb_task_log_queue",
			"cmdb_model_id",
			"cmdb_model_name",
			"cmdb_object_id",
			"cmdb_object_name",
			"server_model_id",
			"server_model_name",
			"server_object_id",
			"server_object_name",
			"task_id_ref",
		} {
			if strings.TrimSpace(task.Metadata[want]) == "" {
				t.Fatalf("%s task metadata missing %s: %#v", receiptType, want, task.Metadata)
			}
		}
		if task.Metadata["cmdb_task_log_status"] != "PENDING" || task.Status != "blocked" {
			t.Fatalf("%s task should remain blocked: %#v", receiptType, task)
		}
		if task.Metadata["cmdb_relation_action_id"] != action["id"] ||
			task.Metadata["cmdb_relation_id"] != prefix+"-rel" ||
			task.Metadata["cmdb_relation_asst_id"] != "belong" ||
			task.Metadata["cmdb_relation_mapping"] != "n:1" {
			t.Fatalf("%s task lost relation schema metadata: %#v", receiptType, task.Metadata)
		}
		if !strings.Contains(task.Metadata["cmdb_task_log_data"], prefix+"-rel") ||
			!strings.Contains(task.Metadata["cmdb_task_log_data"], "topology") {
			t.Fatalf("%s task log data missing relation/action context: %#v", receiptType, task.Metadata)
		}
		if strings.Contains(strings.Join(mapValues(task.Metadata), " "), "relation-task-log-secret-marker") {
			t.Fatalf("%s task metadata leaked secret marker: %#v", receiptType, task.Metadata)
		}
	}
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		"relation-task-log-secret-marker",
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	})
}

func cmdbRelationActionCreatedRequestAndReceipts(t *testing.T, body string) (map[string]any, []any) {
	t.Helper()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("decode relation action response: %v body=%s", err, body)
	}
	action, ok := parsed["action_request"].(map[string]any)
	if !ok {
		t.Fatalf("response missing action_request: %s", body)
	}
	receipts, ok := action["receipts"].([]any)
	if !ok || len(receipts) == 0 {
		t.Fatalf("action_request missing receipts: %#v", action)
	}
	return action, receipts
}

func cmdbRelationActionReceiptRequestRef(t *testing.T, receipts []any, receiptType string) string {
	t.Helper()
	for _, item := range receipts {
		receipt := item.(map[string]any)
		if receipt["receipt_type"] == receiptType {
			ref := strings.TrimSpace(receipt["request_ref"].(string))
			if ref == "" {
				t.Fatalf("%s receipt request_ref is empty: %#v", receiptType, receipts)
			}
			return ref
		}
	}
	t.Fatalf("%s receipt not found: %#v", receiptType, receipts)
	return ""
}
