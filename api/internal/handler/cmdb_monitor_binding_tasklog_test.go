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

func TestCmdbMonitorBindingBlockedRequestsCarryMatureTaskLogShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-task-log-shape")
	createCmdbMonitorTargetFixture(t, "host-task-log-shape")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-task-log-template")

	router := gin.New()
	router.POST("/api/v1/cmdb/monitor-bindings/*path", CreateCmdbMonitorBindingsBlocked)
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	payload := `{
		"host":"srv-task-log",
		"hostid":"host-task-log-shape",
		"templateid":"` + templateID + `",
		"server_object_id":"server-object-55274",
		"server_platform_id":"linux",
		"cmdb_object_id":"cmdb-monitor-binding-task-log-shape",
		"group":{"name":"safe-group"},
		"tags":{"env":"prod"},
		"active_status":"active",
		"hosttype":"server",
		"subtype":"linux",
		"hosttypeLabel":"服务器",
		"subtypeLabel":"Linux",
		"cmdb_attr_id":"OS001",
		"server_attr_id":"agent.ip",
		"server_model_id":"model-55274",
		"server_model_name":"Linux 主机模板",
		"attr":"OS001",
		"attr_stru":[{"cmdb_attr_parent_id":"interfaces","cmdb_attr_id":"ifName","server_attr_id":"net.if.name","is_discovery":1,"macro":"{#IFNAME}"}],
		"queue":"cmdb-task-log-55274",
		"password":"task-log-secret-marker"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/monitor-bindings/"+instanceID, strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	binding, receipts := cmdbMonitorBindingCreatedBindingAndReceipts(t, w.Body.String())
	deliveryRef := cmdbMonitorBindingReceiptRequestRef(t, receipts, "delivery")
	deliveryTask, ok, err := store.GetFindXAgentExecutionTask(deliveryRef)
	if err != nil {
		t.Fatalf("query delivery task ref: %v", err)
	}
	if !ok {
		t.Fatalf("delivery task ref %q was not stored", deliveryRef)
	}
	for _, want := range []string{
		"cmdb_task_log_contract",
		"cmdb_task_log_attr",
		"cmdb_task_log_attr_stru",
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
		if strings.TrimSpace(deliveryTask.Metadata[want]) == "" {
			t.Fatalf("delivery task metadata missing %s: %#v", want, deliveryTask.Metadata)
		}
	}
	if deliveryTask.Action != "cmdb-monitor-binding-delivery" || deliveryTask.Metadata["cmdb_receipt_type"] != "delivery" {
		t.Fatalf("delivery task should carry delivery action metadata: %#v", deliveryTask)
	}
	if deliveryTask.Metadata["cmdb_task_log_status"] != "PENDING" {
		t.Fatalf("task log status must remain blocked, metadata=%#v", deliveryTask.Metadata)
	}
	if deliveryTask.Metadata["cmdb_task_log_attr"] != "OS001=>agent.ip" {
		t.Fatalf("task log attr mismatch: %#v", deliveryTask.Metadata)
	}
	if !strings.Contains(deliveryTask.Metadata["cmdb_task_log_attr_stru"], "ifName=>net.if.name") {
		t.Fatalf("task log attr_stru missing discovery mapping: %#v", deliveryTask.Metadata)
	}
	if !strings.Contains(deliveryTask.Metadata["cmdb_task_log_data"], "cmdb-task-log-55274") {
		t.Fatalf("task log data missing queue reference: %#v", deliveryTask.Metadata)
	}
	if strings.Contains(strings.Join(mapValues(deliveryTask.Metadata), " "), "task-log-secret-marker") {
		t.Fatalf("delivery task metadata leaked secret marker: %#v", deliveryTask.Metadata)
	}
	if deliveryTask.Status != "blocked" || binding["delivery_status"] != "PENDING" {
		t.Fatalf("blocked delivery should stay contract-blocked: task=%#v binding=%#v", deliveryTask, binding)
	}

	for _, receiptType := range []string{"effect", "rollback"} {
		ref := cmdbMonitorBindingReceiptRequestRef(t, receipts, receiptType)
		task, ok, err := store.GetFindXAgentExecutionTask(ref)
		if err != nil {
			t.Fatalf("query %s task ref: %v", receiptType, err)
		}
		if !ok {
			t.Fatalf("%s task ref %q was not stored", receiptType, ref)
		}
		for _, want := range []string{"cmdb_task_log_contract", "cmdb_task_log_data", "cmdb_task_log_queue", "cmdb_model_id", "server_model_id", "task_id_ref"} {
			if strings.TrimSpace(task.Metadata[want]) == "" {
				t.Fatalf("%s task metadata missing %s: %#v", receiptType, want, task.Metadata)
			}
		}
		if task.Status != "blocked" || task.Metadata["cmdb_task_log_status"] != "PENDING" {
			t.Fatalf("%s task should remain blocked: %#v", receiptType, task)
		}
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
		"task-log-secret-marker",
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

func mapValues(values map[string]string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}

func cmdbMonitorBindingFirstBindingAndReceipts(t *testing.T, body string) (map[string]any, []any) {
	t.Helper()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("decode monitor binding response: %v body=%s", err, body)
	}
	bindings, ok := parsed["bindings"].([]any)
	if !ok || len(bindings) == 0 {
		t.Fatalf("monitor binding response missing bindings: %s", body)
	}
	binding, ok := bindings[0].(map[string]any)
	if !ok {
		t.Fatalf("first binding has unexpected shape: %#v", bindings[0])
	}
	receipts, ok := binding["receipts"].([]any)
	if !ok || len(receipts) == 0 {
		t.Fatalf("first binding missing receipts: %#v", binding)
	}
	return binding, receipts
}

func cmdbMonitorBindingCreatedBindingAndReceipts(t *testing.T, body string) (map[string]any, []any) {
	t.Helper()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("decode monitor binding create response: %v body=%s", err, body)
	}
	binding, ok := parsed["binding"].(map[string]any)
	if !ok {
		t.Fatalf("monitor binding create response missing binding: %s", body)
	}
	receipts, ok := binding["receipts"].([]any)
	if !ok || len(receipts) == 0 {
		t.Fatalf("created binding missing receipts: %#v", binding)
	}
	return binding, receipts
}
