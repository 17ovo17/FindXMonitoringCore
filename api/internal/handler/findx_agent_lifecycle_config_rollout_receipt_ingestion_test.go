package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentConfigRolloutReceiptIngestionRecordsBlockedDeliveryReceipt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	if deliveryRef == "" {
		t.Fatalf("fixture should create delivery_request_ref: %#v", rollout.Metadata)
	}

	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"delivery",
		"request_ref":"` + deliveryRef + `",
		"status":"PENDING",
		"contract_id":"cmdb_agent_rollout_delivery_receipt_contract",
		"missing_contracts":["cmdb_agent_rollout_delivery_executor_contract","cmdb_agent_rollout_delivery_executor_registration_contract","cmdb_agent_rollout_delivery_runner_identity_contract","cmdb_agent_rollout_delivery_attested_receipt_contract","cmdb_agent_rollout_delivery_target_binding_contract","cmdb_agent_rollout_delivery_request_ref_match_contract","password=receipt-secret-marker"],
		"evidence_ref":"receipt-safe-ref"
	}`
	w := performConfigRolloutReceiptIngestionPost(rollout.ID, payload)

	if w.Code != http.StatusConflict {
		t.Fatalf("receipt ingestion should remain blocked, got %d body=%s", w.Code, w.Body.String())
	}
	assertConfigRolloutReceiptIngestionContains(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.agent.plugin.dispatch.receipt.ingest.v1",
		`"receipt_type":"delivery"`,
		`"request_ref":"` + deliveryRef + `"`,
		"findx_agent.config_rollout.delivery.receipt.ingest",
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_delivery_executor_registration_contract",
		"cmdb_agent_rollout_delivery_runner_identity_contract",
		"cmdb_agent_rollout_delivery_attested_receipt_contract",
		"cmdb_agent_rollout_delivery_target_binding_contract",
		"cmdb_agent_rollout_delivery_request_ref_match_contract",
	})
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())

	task, ok, err := store.GetFindXAgentExecutionTask(deliveryRef)
	if err != nil || !ok {
		t.Fatalf("delivery request_ref should still resolve after ingestion, ok=%v err=%v", ok, err)
	}
	if task.Status != "blocked" || task.Action != "config_rollout_receipt" {
		t.Fatalf("receipt ingestion must not turn task into execution success: %#v", task)
	}
	if task.Metadata["receipt_status"] != "PENDING" ||
		task.Metadata["receipt_type"] != "delivery" ||
		task.Metadata["receipt_contract_id"] != "cmdb_agent_rollout_delivery_receipt_contract" ||
		strings.TrimSpace(task.Metadata["receipt_audit_ref"]) == "" {
		t.Fatalf("blocked delivery receipt metadata not persisted safely: %#v", task.Metadata)
	}
	if !containsLifecycleTestString(task.EvidenceRefs, "receipt-safe-ref") {
		t.Fatalf("safe evidence_ref should be linked to task evidence refs: %#v", task.EvidenceRefs)
	}
	rawTask, _ := json.Marshal(task)
	assertNoConfigRolloutReceiptIngestionForbidden(t, string(rawTask))

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("blocked receipt ingestion must keep request_ref read gate resolved, got %d body=%s", detail.Code, detail.Body.String())
	}
	assertNoConfigRolloutReceiptIngestionForbidden(t, detail.Body.String())
}

func TestFindXAgentConfigRolloutReceiptIngestionRecordsBlockedEffectReceiptKeepsDeliveryEvidenceGap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	if effectRef == "" {
		t.Fatalf("fixture should create effect_request_ref: %#v", rollout.Metadata)
	}

	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"effect",
		"request_ref":"` + effectRef + `",
		"status":"PENDING",
		"contract_id":"cmdb_agent_rollout_effect_receipt_contract",
		"missing_contracts":["cmdb_agent_rollout_effect_executor_contract","cmdb_agent_rollout_effect_executor_registration_contract","cmdb_agent_rollout_effect_runner_identity_contract","cmdb_agent_rollout_effect_delivery_evidence_match_contract","cmdb_agent_rollout_effect_attested_receipt_contract","cmdb_agent_rollout_effect_request_ref_match_contract","password=effect-receipt-secret-marker"],
		"evidence_ref":"effect-receipt-safe-ref"
	}`
	w := performConfigRolloutReceiptIngestionPost(rollout.ID, payload)

	if w.Code != http.StatusConflict {
		t.Fatalf("effect receipt ingestion should remain blocked, got %d body=%s", w.Code, w.Body.String())
	}
	assertConfigRolloutReceiptIngestionContains(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.agent.plugin.dispatch.receipt.ingest.v1",
		`"receipt_type":"effect"`,
		`"request_ref":"` + effectRef + `"`,
		"findx_agent.config_rollout.effect.receipt.ingest",
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_effect_executor_registration_contract",
		"cmdb_agent_rollout_effect_runner_identity_contract",
		"cmdb_agent_rollout_effect_delivery_evidence_match_contract",
		"cmdb_agent_rollout_effect_attested_receipt_contract",
		"cmdb_agent_rollout_effect_request_ref_match_contract",
	})
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())

	task, ok, err := store.GetFindXAgentExecutionTask(effectRef)
	if err != nil || !ok {
		t.Fatalf("effect request_ref should still resolve after ingestion, ok=%v err=%v", ok, err)
	}
	if task.Status != "blocked" || task.Action != "config_rollout_receipt" {
		t.Fatalf("effect receipt ingestion must not turn task into execution success: %#v", task)
	}
	if task.Metadata["receipt_status"] != "PENDING" ||
		task.Metadata["receipt_type"] != "effect" ||
		task.Metadata["receipt_contract_id"] != "cmdb_agent_rollout_effect_receipt_contract" ||
		strings.TrimSpace(task.Metadata["receipt_audit_ref"]) == "" {
		t.Fatalf("blocked effect receipt metadata not persisted safely: %#v", task.Metadata)
	}
	if !strings.Contains(task.Metadata["receipt_missing_contracts"], "cmdb_agent_rollout_effect_delivery_evidence_match_contract") {
		t.Fatalf("effect receipt metadata should keep delivery evidence match gap: %#v", task.Metadata)
	}
	if !containsLifecycleTestString(task.EvidenceRefs, "effect-receipt-safe-ref") {
		t.Fatalf("safe evidence_ref should be linked to task evidence refs: %#v", task.EvidenceRefs)
	}
	rawTask, _ := json.Marshal(task)
	assertNoConfigRolloutReceiptIngestionForbidden(t, string(rawTask))
}

func TestFindXAgentConfigRolloutReceiptIngestionRecordsBlockedRollbackReceiptKeepsOperationContextGap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	rollbackRef := strings.TrimSpace(rollout.Metadata["rollback_request_ref"])
	if rollbackRef == "" {
		t.Fatalf("fixture should create rollback_request_ref: %#v", rollout.Metadata)
	}

	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"rollback",
		"request_ref":"` + rollbackRef + `",
		"status":"PENDING",
		"contract_id":"cmdb_agent_rollout_rollback_receipt_contract",
		"missing_contracts":["cmdb_agent_rollout_rollback_executor_contract","cmdb_agent_rollout_rollback_executor_registration_contract","cmdb_agent_rollout_rollback_runner_identity_contract","cmdb_agent_rollout_rollback_operation_context_contract","cmdb_agent_rollout_rollback_attested_receipt_contract","cmdb_agent_rollout_rollback_request_ref_match_contract","password=rollback-receipt-secret-marker"],
		"evidence_ref":"rollback-receipt-safe-ref"
	}`
	w := performConfigRolloutReceiptIngestionPost(rollout.ID, payload)

	if w.Code != http.StatusConflict {
		t.Fatalf("rollback receipt ingestion should remain blocked, got %d body=%s", w.Code, w.Body.String())
	}
	assertConfigRolloutReceiptIngestionContains(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.agent.plugin.dispatch.receipt.ingest.v1",
		`"receipt_type":"rollback"`,
		`"request_ref":"` + rollbackRef + `"`,
		"findx_agent.config_rollout.rollback.receipt.ingest",
		"cmdb_agent_rollout_rollback_executor_contract",
		"cmdb_agent_rollout_rollback_executor_registration_contract",
		"cmdb_agent_rollout_rollback_runner_identity_contract",
		"cmdb_agent_rollout_rollback_operation_context_contract",
		"cmdb_agent_rollout_rollback_attested_receipt_contract",
		"cmdb_agent_rollout_rollback_request_ref_match_contract",
	})
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())

	task, ok, err := store.GetFindXAgentExecutionTask(rollbackRef)
	if err != nil || !ok {
		t.Fatalf("rollback request_ref should still resolve after ingestion, ok=%v err=%v", ok, err)
	}
	if task.Status != "blocked" || task.Action != "config_rollout_receipt" {
		t.Fatalf("rollback receipt ingestion must not turn task into execution success: %#v", task)
	}
	if task.Metadata["receipt_status"] != "PENDING" ||
		task.Metadata["receipt_type"] != "rollback" ||
		task.Metadata["receipt_contract_id"] != "cmdb_agent_rollout_rollback_receipt_contract" ||
		strings.TrimSpace(task.Metadata["receipt_audit_ref"]) == "" {
		t.Fatalf("blocked rollback receipt metadata not persisted safely: %#v", task.Metadata)
	}
	if !strings.Contains(task.Metadata["receipt_missing_contracts"], "cmdb_agent_rollout_rollback_operation_context_contract") {
		t.Fatalf("rollback receipt metadata should keep operation context gap: %#v", task.Metadata)
	}
	if !containsLifecycleTestString(task.EvidenceRefs, "rollback-receipt-safe-ref") {
		t.Fatalf("safe evidence_ref should be linked to task evidence refs: %#v", task.EvidenceRefs)
	}
	rawTask, _ := json.Marshal(task)
	assertNoConfigRolloutReceiptIngestionForbidden(t, string(rawTask))
}

func TestFindXAgentConfigRolloutReceiptIngestionRejectsMismatchedRequestRef(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	if effectRef == "" {
		t.Fatalf("fixture should create effect_request_ref: %#v", rollout.Metadata)
	}

	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"delivery",
		"request_ref":"` + effectRef + `",
		"status":"PENDING"
	}`
	w := performConfigRolloutReceiptIngestionPost(rollout.ID, payload)

	if w.Code != http.StatusConflict {
		t.Fatalf("mismatched request_ref should be blocked, got %d body=%s", w.Code, w.Body.String())
	}
	assertConfigRolloutReceiptIngestionContains(t, w.Body.String(), []string{
		"PENDING",
		"cmdb_agent_rollout_request_ref_resolve_contract",
		"cmdb_agent_rollout_execution_task_match_contract",
	})
	assertConfigRolloutReceiptIngestionExcludes(t, w.Body.String(), []string{`"receipt":{`, `"code":0`, `"status":"ready"`})
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())
}

func TestFindXAgentConfigRolloutReceiptIngestionRejectsSuccessStatusAndSensitiveEcho(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	if effectRef == "" {
		t.Fatalf("fixture should create effect_request_ref: %#v", rollout.Metadata)
	}

	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"effect",
		"request_ref":"` + effectRef + `",
		"status":"succeeded",
		"missing_contracts":["cmdb_agent_rollout_effect_executor_contract"],
		"result":{"password":"receipt-status-secret-marker","dsn":"mysql://user:pass@example/db"}
	}`
	w := performConfigRolloutReceiptIngestionPost(rollout.ID, payload)

	if w.Code != http.StatusConflict {
		t.Fatalf("fake success receipt should be blocked, got %d body=%s", w.Code, w.Body.String())
	}
	assertConfigRolloutReceiptIngestionContains(t, w.Body.String(), []string{
		"PENDING",
		"cmdb_agent_rollout_receipt_status_contract",
	})
	assertConfigRolloutReceiptIngestionExcludes(t, w.Body.String(), []string{
		"receipt-status-secret-marker",
		"mysql://user:pass@example/db",
		`"succeeded"`,
		`"success"`,
		`"receipt":{`,
		`"code":0`,
		`"status":"ready"`,
	})
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())
}

func TestFindXAgentConfigRolloutReceiptIngestionKeepsBlockedReceiptWhenAuditUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	rollbackRef := strings.TrimSpace(rollout.Metadata["rollback_request_ref"])
	if rollbackRef == "" {
		t.Fatalf("fixture should create rollback_request_ref: %#v", rollout.Metadata)
	}

	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"rollback",
		"request_ref":"` + rollbackRef + `",
		"status":"PENDING",
		"contract_id":"cmdb_agent_rollout_rollback_receipt_contract",
		"missing_contracts":["cmdb_agent_rollout_rollback_executor_contract"],
		"evidence_ref":"receipt-audit-unavailable-safe-ref"
	}`
	w := performConfigRolloutReceiptIngestionPostWithAudit(rollout.ID, payload, func(model.MonitorAuditLog) (model.MonitorAuditLog, error) {
		return model.MonitorAuditLog{}, errors.New("audit backend unavailable")
	})

	if w.Code != http.StatusConflict {
		t.Fatalf("audit backend failure must not turn blocked receipt ingestion into 503, got %d body=%s", w.Code, w.Body.String())
	}
	assertConfigRolloutReceiptIngestionContains(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.agent.plugin.dispatch.receipt.ingest.v1",
		"cmdb_agent_rollout_receipt_audit_persistence_contract",
		`"receipt_type":"rollback"`,
	})
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())

	task, ok, err := store.GetFindXAgentExecutionTask(rollbackRef)
	if err != nil || !ok {
		t.Fatalf("rollback request_ref should remain persisted even when audit is unavailable, ok=%v err=%v", ok, err)
	}
	if task.Metadata["receipt_status"] != "PENDING" || task.Metadata["receipt_type"] != "rollback" {
		t.Fatalf("blocked rollback receipt metadata should still be persisted: %#v", task.Metadata)
	}
	if !containsLifecycleTestString(task.EvidenceRefs, "receipt-audit-unavailable-safe-ref") {
		t.Fatalf("safe evidence ref should still be linked when audit is unavailable: %#v", task.EvidenceRefs)
	}
}

func TestFindXAgentConfigRolloutReceiptIngestionRecordsBlockedDataArrivalReceiptWithReceiverEvidence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	if deliveryRef == "" {
		t.Fatalf("fixture should create delivery_request_ref: %#v", rollout.Metadata)
	}
	if _, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindMetrics,
		AgentID:      "agent-a",
		TargetID:     "host-a",
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/metrics/write-compatible"},
		Metadata: map[string]string{
			"source_rollout_id": rollout.ID,
			"request_ref":       deliveryRef,
			"plugin_id":         rollout.PluginID,
			"password":          "data-arrival-secret-marker",
		},
	}); err != nil {
		t.Fatalf("save receiver evidence: %v", err)
	}

	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"data_arrival",
		"request_ref":"` + deliveryRef + `",
		"status":"PENDING",
		"contract_id":"cmdb_agent_rollout_data_arrival_receipt_contract",
		"missing_contracts":["cmdb_agent_rollout_data_arrival_receipt_contract","cmdb_agent_rollout_evidence_chain_contract"],
		"evidence_ref":"data-arrival-receipt-safe-ref"
	}`
	w := performConfigRolloutReceiptIngestionPost(rollout.ID, payload)

	if w.Code != http.StatusConflict {
		t.Fatalf("data-arrival receipt ingestion should remain blocked, got %d body=%s", w.Code, w.Body.String())
	}
	assertConfigRolloutReceiptIngestionContains(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.agent.plugin.dispatch.receipt.ingest.v1",
		`"receipt_type":"data_arrival"`,
		"cmdb_agent_rollout_data_arrival_receipt_contract",
		"cmdb_agent_rollout_evidence_chain_contract",
		"findx_agent.config_rollout.data_arrival.receipt.ingest",
	})
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())

	task, ok, err := store.GetFindXAgentExecutionTask(deliveryRef)
	if err != nil || !ok {
		t.Fatalf("delivery request_ref should still resolve after data-arrival ingestion, ok=%v err=%v", ok, err)
	}
	if task.Status != "blocked" || task.Metadata["receipt_type"] != "delivery" || task.Metadata["phase"] != "delivery" {
		t.Fatalf("data-arrival receipt must not rewrite canonical delivery task phase: %#v", task)
	}
	if task.Metadata["data_arrival_receipt_status"] != "PENDING" ||
		task.Metadata["data_arrival_receipt_contract_id"] != "cmdb_agent_rollout_data_arrival_receipt_contract" ||
		task.Metadata["data_arrival_receipt_ingestion_contract"] != cmdbAgentRolloutReceiptIngestContract ||
		strings.TrimSpace(task.Metadata["data_arrival_receipt_audit_ref"]) == "" {
		t.Fatalf("blocked data-arrival receipt metadata not persisted safely: %#v", task.Metadata)
	}
	if !containsLifecycleTestString(task.EvidenceRefs, "data-arrival-receipt-safe-ref") {
		t.Fatalf("safe evidence_ref should be linked to task evidence refs: %#v", task.EvidenceRefs)
	}
	rawTask, _ := json.Marshal(task)
	assertNoConfigRolloutReceiptIngestionForbidden(t, string(rawTask))

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("data-arrival receipt ingestion must not break request_ref read gate, got %d body=%s", detail.Code, detail.Body.String())
	}
	evidence := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref="+deliveryRef, ListFindXAgentDataArrivalEvidence)
	readPayload := decodeDataArrivalRuntimeReadOK(t, evidence)
	if evidence.Code != http.StatusOK || readPayload.Status != "blocked" || readPayload.EvidenceCount != 1 {
		t.Fatalf("data-arrival receipt must not turn evidence read into success, code=%d payload=%#v body=%s", evidence.Code, readPayload, evidence.Body.String())
	}
	for _, want := range []string{
		"cmdb_agent_rollout_data_arrival_receipt_contract",
		"cmdb_agent_rollout_evidence_chain_contract",
	} {
		if !containsLifecycleTestString(readPayload.MissingContracts, want) {
			t.Fatalf("data-arrival read should keep finalization gap %q, payload=%#v", want, readPayload)
		}
	}
	assertNoConfigRolloutReceiptIngestionForbidden(t, detail.Body.String())
	assertNoDataArrivalRuntimeReadForbidden(t, evidence.Body.String())
}

func TestFindXAgentConfigRolloutReceiptIngestionRejectsDataArrivalReceiptWithoutReceiverEvidence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	if deliveryRef == "" {
		t.Fatalf("fixture should create delivery_request_ref: %#v", rollout.Metadata)
	}

	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"data_arrival",
		"request_ref":"` + deliveryRef + `",
		"status":"PENDING"
	}`
	w := performConfigRolloutReceiptIngestionPost(rollout.ID, payload)

	if w.Code != http.StatusConflict {
		t.Fatalf("data-arrival receipt without receiver evidence should be blocked, got %d body=%s", w.Code, w.Body.String())
	}
	assertConfigRolloutReceiptIngestionContains(t, w.Body.String(), []string{
		"PENDING",
		cmdbAgentRolloutDataArrivalEvidenceContract,
		cmdbAgentRolloutReceiverEvidenceContract,
	})
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())
	task, ok, err := store.GetFindXAgentExecutionTask(deliveryRef)
	if err != nil || !ok {
		t.Fatalf("delivery request_ref should resolve after rejected data-arrival receipt, ok=%v err=%v", ok, err)
	}
	if strings.TrimSpace(task.Metadata["data_arrival_receipt_status"]) != "" {
		t.Fatalf("rejected data-arrival receipt must not persist finalization metadata: %#v", task.Metadata)
	}
}

func createConfigRolloutReceiptIngestionFixture(t *testing.T) model.FindXAgentConfigRollout {
	t.Helper()
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"cmdb-host-plugin-dispatch","target_ids":["host-a"],"agent_ids":["agent-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"cmdb_host_probe","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"dispatch","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"cmdb_host","cmdb_host_ref":"host-a","agent_ref":"agent-a","assignment_ref":"missing-assignment-ref","target_binding_ref":"binding-a","dashboard_refs":"dashboard:redis-overview","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref","token":"secret-token","cookie":"secret-cookie","dsn":"mysql://user:pass@host/db","private_key":"secret-key"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Data.ID == "" {
		t.Fatalf("create dispatch rollout fixture failed, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	return payload.Data
}

func performConfigRolloutReceiptIngestionPost(rolloutID, payload string) *httptest.ResponseRecorder {
	return performConfigRolloutReceiptIngestionPostWithAudit(rolloutID, payload, store.AddMonitorAuditLog)
}

func performConfigRolloutReceiptIngestionPostWithAudit(rolloutID, payload string, addAuditLog func(model.MonitorAuditLog) (model.MonitorAuditLog, error)) *httptest.ResponseRecorder {
	router := gin.New()
	router.POST("/api/v1/findx-agents/config-rollouts/:id/receipts", func(c *gin.Context) {
		ingestFindXAgentConfigRolloutReceipt(c, addAuditLog)
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/findx-agents/config-rollouts/"+rolloutID+"/receipts", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func assertConfigRolloutReceiptIngestionContains(t *testing.T, body string, wants []string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(body, want) {
			t.Fatalf("receipt ingestion response missing %q: %s", want, body)
		}
	}
}

func assertConfigRolloutReceiptIngestionExcludes(t *testing.T, body string, forbidden []string) {
	t.Helper()
	for _, item := range forbidden {
		if strings.Contains(body, item) {
			t.Fatalf("receipt ingestion response leaked forbidden marker %q: %s", item, body)
		}
	}
}

func assertNoConfigRolloutReceiptIngestionForbidden(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{
		`"status":"queued"`,
		`"status":"running"`,
		`"status":"applied"`,
		`"status":"installed"`,
		`"status":"data_arrived"`,
		`"status":"service_registered"`,
		`"status":"rolled_back"`,
		`"status":"uninstalled"`,
		`"status":"delivered"`,
		`"status":"effective"`,
		`"status":"succeeded"`,
		`"status":"success"`,
		"receipt-secret-marker",
		"receipt-status-secret-marker",
		"mysql://",
		"postgres://",
		"password",
		"token",
		"cookie",
		"private_key",
		"dsn",
		"nightingale",
		"skywalking",
		"signoz",
		"categraf",
		"catpaw",
		"grafana",
		"prometheus",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("receipt ingestion leaked forbidden marker %q: %s", forbidden, body)
		}
	}
}
