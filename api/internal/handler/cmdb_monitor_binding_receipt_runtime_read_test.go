package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbMonitorBindingReceiptRuntimeReadBlocksMissingRequestRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-runtime-missing")
	createCmdbMonitorTargetFixture(t, "host-runtime-missing")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-runtime-missing")
	binding := saveCmdbMonitorBindingRuntimeFixture(t, instanceID, "host-runtime-missing", templateID)
	for _, receiptType := range []string{"delivery", "effect", "rollback"} {
		saveCmdbMonitorBindingRuntimeReceipt(t, binding, receiptType, "cmdb-monitor-binding-runtime-missing-"+receiptType)
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	for _, path := range []string{
		"/api/v1/cmdb/monitor-bindings/" + instanceID,
		"/api/v1/cmdb/monitor-bindings/" + instanceID + "/" + binding.ID,
		"/api/v1/cmdb/monitor-bindings/" + instanceID + "/receipts",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Fatalf("%s status = %d, want %d, body=%s", path, w.Code, http.StatusConflict, w.Body.String())
		}
		assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
			"PENDING",
			"cmdb.monitor_binding.runtime.receipts.read.v1",
			"cmdb_monitor_binding_request_ref_resolve_contract",
			"cmdb_monitor_binding_execution_task_contract",
		})
		assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
			`"code":0`,
			`"status":"ready"`,
			`"bindings":[`,
			`"binding":{`,
			`"receipts":[`,
			`"queued"`,
			`"running"`,
			`"delivered"`,
			`"effective"`,
			`"succeeded"`,
			"binding-runtime-secret-marker",
		})
	}
}

func TestCmdbMonitorBindingReceiptRuntimeReadBlocksMismatchedRequestRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-runtime-mismatch")
	createCmdbMonitorTargetFixture(t, "host-runtime-mismatch")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-runtime-mismatch")
	binding := saveCmdbMonitorBindingRuntimeFixture(t, instanceID, "host-runtime-mismatch", templateID)
	delivery := saveCmdbMonitorBindingRuntimeTask(t, binding, "delivery", map[string]string{"cmdb_hostid": "wrong-host"})
	effect := saveCmdbMonitorBindingRuntimeTask(t, binding, "effect", nil)
	rollback := saveCmdbMonitorBindingRuntimeTask(t, binding, "rollback", nil)
	saveCmdbMonitorBindingRuntimeReceipt(t, binding, "delivery", delivery.ID)
	saveCmdbMonitorBindingRuntimeReceipt(t, binding, "effect", effect.ID)
	saveCmdbMonitorBindingRuntimeReceipt(t, binding, "rollback", rollback.ID)

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
		"cmdb.monitor_binding.runtime.receipts.read.v1",
		"cmdb_monitor_binding_execution_task_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"code":0`,
		`"receipts":[`,
		`"queued"`,
		`"running"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	})
}

func TestCmdbMonitorBindingReceiptRuntimeReadBlocksNonExecutionRequestRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-runtime-nonexecution")
	createCmdbMonitorTargetFixture(t, "host-runtime-nonexecution")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-runtime-nonexecution")
	binding := saveCmdbMonitorBindingRuntimeFixture(t, instanceID, "host-runtime-nonexecution", templateID)
	delivery := saveCmdbMonitorBindingRuntimeRollout(t, binding, nil)
	effect := saveCmdbMonitorBindingRuntimeTask(t, binding, "effect", nil)
	rollback := saveCmdbMonitorBindingRuntimeTask(t, binding, "rollback", nil)
	saveCmdbMonitorBindingRuntimeReceipt(t, binding, "delivery", delivery.ID)
	saveCmdbMonitorBindingRuntimeReceipt(t, binding, "effect", effect.ID)
	saveCmdbMonitorBindingRuntimeReceipt(t, binding, "rollback", rollback.ID)

	router := gin.New()
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID+"/"+binding.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.runtime.receipts.read.v1",
		"cmdb_monitor_binding_execution_task_contract",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"binding":{`,
		`"code":0`,
		`"succeeded"`,
	})
}

func TestCmdbMonitorBindingReceiptRuntimeReadKeepsBlockedAfterCompleteReceiptsWithBlockedTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-runtime-resolved")
	createCmdbMonitorTargetFixture(t, "host-runtime-resolved")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-runtime-resolved")
	binding := saveCmdbMonitorBindingRuntimeFixture(t, instanceID, "host-runtime-resolved", templateID)
	for _, receiptType := range []string{"delivery", "effect", "rollback"} {
		task := saveCmdbMonitorBindingRuntimeTask(t, binding, receiptType, nil)
		saveCmdbMonitorBindingRuntimeReceipt(t, binding, receiptType, task.ID)
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	for _, path := range []string{
		"/api/v1/cmdb/monitor-bindings/" + instanceID,
		"/api/v1/cmdb/monitor-bindings/" + instanceID + "/" + binding.ID,
		"/api/v1/cmdb/monitor-bindings/" + instanceID + "/receipts",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Fatalf("%s status = %d, want %d, body=%s", path, w.Code, http.StatusConflict, w.Body.String())
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
			`"code":0`,
			`"status":"ready"`,
			`"bindings":[`,
			`"binding":{`,
			`"receipts":[`,
			`"queued"`,
			`"running"`,
			`"delivered"`,
			`"effective"`,
			`"succeeded"`,
			`"success"`,
			"binding-runtime-secret-marker",
		})
	}
}

func TestCmdbMonitorBindingReceiptRuntimeReadKeepsExecutorGapAfterBlockedReceipts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-executor-gap")
	createCmdbMonitorTargetFixture(t, "host-executor-gap")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-executor-gap")
	binding := saveCmdbMonitorBindingRuntimeFixture(t, instanceID, "host-executor-gap", templateID)
	for _, receiptType := range []string{"delivery", "effect", "rollback"} {
		task := saveCmdbMonitorBindingRuntimeTask(t, binding, receiptType, nil)
		saveCmdbMonitorBindingRuntimeReceipt(t, binding, receiptType, task.ID)
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	for _, path := range []string{
		"/api/v1/cmdb/monitor-bindings/" + instanceID,
		"/api/v1/cmdb/monitor-bindings/" + instanceID + "/" + binding.ID,
		"/api/v1/cmdb/monitor-bindings/" + instanceID + "/receipts",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Fatalf("%s status = %d, want %d, body=%s", path, w.Code, http.StatusConflict, w.Body.String())
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
			`"code":0`,
			`"status":"ready"`,
			`"bindings":[`,
			`"binding":{`,
			`"receipts":[`,
			`"queued"`,
			`"running"`,
			`"applied"`,
			`"installed"`,
			`"data_arrived"`,
			`"service_registered"`,
			`"rolled_back"`,
			`"uninstalled"`,
			`"delivered"`,
			`"effective"`,
			`"succeeded"`,
			`"success"`,
			`"imported"`,
			"binding-runtime-secret-marker",
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

func saveCmdbMonitorBindingRuntimeFixture(t *testing.T, instanceID, hostID, templateID string) model.CmdbMonitorBinding {
	t.Helper()
	binding, err := store.SaveCmdbMonitorBinding(&model.CmdbMonitorBinding{
		InstanceID:   instanceID,
		HostID:       hostID,
		TemplateID:   templateID,
		CmdbAttrID:   "OS001",
		ServerAttrID: "agent.ip",
		AuditRef:     "cmdb-monitor-binding-runtime-audit-" + hostID,
		Attr:         "OS001=>agent.ip",
		Queue:        "cmdb-monitor-binding-runtime-queue",
	})
	if err != nil {
		t.Fatalf("save monitor binding: %v", err)
	}
	return *binding
}

func saveCmdbMonitorBindingRuntimeReceipt(t *testing.T, binding model.CmdbMonitorBinding, receiptType, requestRef string) {
	t.Helper()
	contract := "cmdb_monitor_binding_" + receiptType + "_receipt_contract"
	if receiptType == "rollback" {
		contract = "binding_rollback_contract"
	}
	if _, err := store.SaveCmdbMonitorBindingReceipt(&model.CmdbMonitorBindingReceipt{
		BindingID:   binding.ID,
		InstanceID:  binding.InstanceID,
		ReceiptType: receiptType,
		Status:      "PENDING",
		ContractID:  contract,
		MissingJSON: `["cmdb_monitor_binding_` + receiptType + `_executor"]`,
		RequestRef:  requestRef,
		AuditRef:    binding.AuditRef,
	}); err != nil {
		t.Fatalf("save %s receipt: %v", receiptType, err)
	}
}

func saveCmdbMonitorBindingRuntimeRollout(t *testing.T, binding model.CmdbMonitorBinding, overrides map[string]string) model.FindXAgentConfigRollout {
	t.Helper()
	metadata := cmdbMonitorBindingRuntimeMetadata(binding, "delivery")
	for key, value := range overrides {
		metadata[key] = value
	}
	rollout, err := store.SaveFindXAgentConfigRollout(model.FindXAgentConfigRollout{
		TemplateID:    "host-plugin",
		TargetIDs:     []string{binding.HostID},
		ConfigVersion: binding.ID,
		Status:        "blocked",
		Blocker:       "PENDING: test delivery runtime",
		Audit:         "findx_agent.config_rollout.requested",
		Metadata:      metadata,
	})
	if err != nil {
		t.Fatalf("save delivery rollout: %v", err)
	}
	return rollout
}

func saveCmdbMonitorBindingRuntimeTask(t *testing.T, binding model.CmdbMonitorBinding, receiptType string, overrides map[string]string) model.FindXAgentExecutionTask {
	t.Helper()
	metadata := cmdbMonitorBindingRuntimeMetadata(binding, receiptType)
	for key, value := range overrides {
		metadata[key] = value
	}
	task, err := store.SaveFindXAgentExecutionTask(model.FindXAgentExecutionTask{
		Action:    "cmdb-monitor-binding-" + receiptType,
		TargetIDs: []string{binding.HostID},
		Status:    "blocked",
		Blocker:   "PENDING: test " + receiptType + " runtime",
		Audit:     "findx_agent.task.requested",
		Metadata:  metadata,
	})
	if err != nil {
		t.Fatalf("save %s task: %v", receiptType, err)
	}
	return task
}

func cmdbMonitorBindingRuntimeMetadata(binding model.CmdbMonitorBinding, receiptType string) map[string]string {
	return map[string]string{
		"scope":             "cmdb_monitor_binding",
		"cmdb_binding_id":   binding.ID,
		"cmdb_instance_id":  binding.InstanceID,
		"cmdb_hostid":       binding.HostID,
		"cmdb_templateid":   binding.TemplateID,
		"cmdb_attr_id":      binding.CmdbAttrID,
		"server_attr_id":    binding.ServerAttrID,
		"cmdb_receipt_type": receiptType,
	}
}
