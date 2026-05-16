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

func TestFindXAgentConfigRolloutDispatchDetailRequiresReceiptRequestRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := saveConfigRolloutRuntimeReadRollout(t, map[string]string{})

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	payload := decodeConfigRolloutRuntimeReadBlocked(t, w)

	if w.Code != http.StatusConflict || payload.Status != "blocked" || payload.Contract != cmdbAgentRolloutRuntimeReadContract {
		t.Fatalf("dispatch detail without receipt request refs should be blocked, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{
		"cmdb_agent_rollout_delivery_request_ref_contract",
		"cmdb_agent_rollout_effect_request_ref_contract",
		"cmdb_agent_rollout_rollback_request_ref_contract",
	} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("missing request refs should include %q, payload=%#v", want, payload)
		}
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailRejectsStaleReceiptRequestRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := saveConfigRolloutRuntimeReadRollout(t, map[string]string{
		"delivery_request_ref": "missing-delivery-task",
		"effect_request_ref":   "missing-effect-task",
		"rollback_request_ref": "missing-rollback-task",
	})

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	payload := decodeConfigRolloutRuntimeReadBlocked(t, w)

	if w.Code != http.StatusConflict || payload.Status != "blocked" {
		t.Fatalf("dispatch detail with stale request refs should be blocked, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{
		cmdbAgentRolloutRequestRefResolveContract,
		cmdbAgentRolloutExecutionTaskContract,
	} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("stale request refs should include %q, payload=%#v", want, payload)
		}
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailRejectsMismatchedReceiptTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := saveConfigRolloutRuntimeReadRollout(t, map[string]string{
		"delivery_request_ref": "delivery-task",
		"effect_request_ref":   "effect-task",
		"rollback_request_ref": "rollback-task",
	})
	saveConfigRolloutRuntimeReadTask(t, "delivery-task", rollout, "delivery", map[string]string{"plugin_id": "mysql"})
	saveConfigRolloutRuntimeReadTask(t, "effect-task", rollout, "effect", map[string]string{"source_rollout_id": "other-rollout"})
	saveConfigRolloutRuntimeReadTask(t, "rollback-task", rollout, "rollback", map[string]string{"receipt_kind": "delivery"})

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	payload := decodeConfigRolloutRuntimeReadBlocked(t, w)

	if w.Code != http.StatusConflict || !containsLifecycleTestString(payload.MissingContracts, cmdbAgentRolloutTaskMatchContract) {
		t.Fatalf("mismatched receipt tasks should be blocked by task match contract, code=%d payload=%#v", w.Code, payload)
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailReturnsBlockedRecordWhenReceiptTasksResolve(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := saveConfigRolloutRuntimeReadRollout(t, map[string]string{
		"delivery_request_ref": "delivery-task",
		"effect_request_ref":   "effect-task",
		"rollback_request_ref": "rollback-task",
	})
	saveConfigRolloutRuntimeReadTask(t, "delivery-task", rollout, "delivery", nil)
	saveConfigRolloutRuntimeReadTask(t, "effect-task", rollout, "effect", nil)
	saveConfigRolloutRuntimeReadTask(t, "rollback-task", rollout, "rollback", nil)

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if w.Code != http.StatusOK {
		t.Fatalf("dispatch detail with resolved blocked receipt tasks should return record, got %d body=%s", w.Code, w.Body.String())
	}
	var detail model.FindXAgentConfigRollout
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("invalid detail response: %v body=%s", err, w.Body.String())
	}
	if detail.ID != rollout.ID || detail.Status != "blocked" {
		t.Fatalf("detail should return blocked rollout after receipt task read gate resolves, detail=%#v", detail)
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailDoesNotRequireWriterRequestRef(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := saveConfigRolloutRuntimeReadRollout(t, map[string]string{
		"delivery_request_ref": "delivery-task",
		"effect_request_ref":   "effect-task",
		"rollback_request_ref": "rollback-task",
	})
	saveConfigRolloutRuntimeReadTask(t, "delivery-task", rollout, "delivery", nil)
	saveConfigRolloutRuntimeReadTask(t, "effect-task", rollout, "effect", nil)
	saveConfigRolloutRuntimeReadTask(t, "rollback-task", rollout, "rollback", nil)

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if w.Code != http.StatusOK {
		t.Fatalf("dispatch detail should not require writer_request_ref, got %d body=%s", w.Code, w.Body.String())
	}
	var detail model.FindXAgentConfigRollout
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("invalid detail response: %v body=%s", err, w.Body.String())
	}
	if detail.ID != rollout.ID || detail.Status != "blocked" {
		t.Fatalf("detail should return blocked rollout without writer_request_ref, detail=%#v", detail)
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, w.Body.String())
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailIgnoresStaleWriterRequestRef(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := saveConfigRolloutRuntimeReadRollout(t, map[string]string{
		"delivery_request_ref": "delivery-task",
		"effect_request_ref":   "effect-task",
		"rollback_request_ref": "rollback-task",
		"writer_request_ref":   "stale-writer-task",
	})
	saveConfigRolloutRuntimeReadTask(t, "delivery-task", rollout, "delivery", nil)
	saveConfigRolloutRuntimeReadTask(t, "effect-task", rollout, "effect", nil)
	saveConfigRolloutRuntimeReadTask(t, "rollback-task", rollout, "rollback", nil)

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if w.Code != http.StatusOK {
		t.Fatalf("dispatch detail should ignore stale writer_request_ref, got %d body=%s", w.Code, w.Body.String())
	}
	var detail model.FindXAgentConfigRollout
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("invalid detail response: %v body=%s", err, w.Body.String())
	}
	if detail.ID != rollout.ID || detail.Status != "blocked" {
		t.Fatalf("detail should return blocked rollout while ignoring stale writer_request_ref, detail=%#v", detail)
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, w.Body.String())
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutRuntimeReadGateSkipsNonDispatchRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout, err := store.SaveFindXAgentConfigRollout(model.FindXAgentConfigRollout{
		TemplateID:      "cmdb-host-plugin-assign",
		TargetIDs:       []string{"host-a"},
		AgentIDs:        []string{"agent-a"},
		ProviderMode:    "cmdb_host_probe",
		PluginID:        "redis",
		RolloutStrategy: "assign",
		RemoteMutation:  false,
		Status:          "blocked",
		Blocker:         "PENDING: assignment read",
		Metadata: map[string]string{
			"scope":         "cmdb_host",
			"cmdb_host_ref": "host-a",
			"agent_ref":     "agent-a",
		},
	})
	if err != nil {
		t.Fatalf("save assign rollout: %v", err)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if w.Code != http.StatusOK {
		t.Fatalf("assign detail should not use dispatch runtime read gate, got %d body=%s", w.Code, w.Body.String())
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutDispatchPostCreatesReceiptRequestRefTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	body := strings.NewReader(`{"template_id":"cmdb-host-plugin-dispatch","target_ids":["host-a"],"agent_ids":["agent-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"cmdb_host_probe","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"dispatch","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"cmdb_host","cmdb_host_ref":"host-a","agent_ref":"agent-a","assignment_ref":"missing-assignment-ref","target_binding_ref":"binding-a","dashboard_refs":"dashboard:redis-overview","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref","writer_request_ref":"stale-writer-task","token":"secret-token","cookie":"secret-cookie","dsn":"mysql://user:pass@host/db","private_key":"secret-key"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("dispatch post should remain blocked with receipt request refs, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	for _, receipt := range []string{"delivery", "effect", "rollback"} {
		refKey := receipt + "_request_ref"
		ref := strings.TrimSpace(payload.Data.Metadata[refKey])
		if ref == "" {
			t.Fatalf("dispatch rollout metadata should include %s, metadata=%#v", refKey, payload.Data.Metadata)
		}
		task, ok, err := store.GetFindXAgentExecutionTask(ref)
		if err != nil || !ok {
			t.Fatalf("receipt request ref %s should resolve to blocked execution task, ok=%v err=%v", ref, ok, err)
		}
		if task.Action != "config_rollout_receipt" || task.Status != "blocked" {
			t.Fatalf("receipt task should be blocked config_rollout_receipt, task=%#v", task)
		}
		if !metadataValueMatchesAny(task.Metadata, []string{"source_rollout_id", "config_rollout_id", "rollout_ref"}, payload.Data.ID) ||
			!metadataValueMatchesAny(task.Metadata, []string{"receipt_kind", "receipt_type", "phase"}, receipt) ||
			task.Metadata["plugin_id"] != payload.Data.PluginID {
			t.Fatalf("receipt task metadata should match rollout and receipt=%s, task=%#v rollout=%#v", receipt, task, payload.Data)
		}
		if !stringSlicesIntersect(task.TargetIDs, payload.Data.TargetIDs) && !stringSlicesIntersect(task.AgentIDs, payload.Data.AgentIDs) {
			t.Fatalf("receipt task should target same host or agent, task=%#v rollout=%#v", task, payload.Data)
		}
		if strings.Contains(strings.ToLower(task.Metadata["requested_by"]), "secret") {
			t.Fatalf("receipt task metadata should not carry sensitive actor, task=%#v", task)
		}
	}

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+payload.Data.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("dispatch detail should pass request_ref read gate after POST-created tasks, got %d body=%s", detail.Code, detail.Body.String())
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, w.Body.String())
	assertNoWriterRequestRefRuntimeReadLeak(t, detail.Body.String())
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
	assertNoConfigRolloutExecutionStates(t, detail.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, detail.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDeliveryRequestRefKeepsExecutorRegistrationGaps(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	if deliveryRef == "" {
		t.Fatalf("fixture should create delivery_request_ref: %#v", rollout.Metadata)
	}

	task, ok, err := store.GetFindXAgentExecutionTask(deliveryRef)
	if err != nil || !ok {
		t.Fatalf("delivery request_ref should resolve to blocked execution task, ok=%v err=%v", ok, err)
	}
	if task.Action != "config_rollout_receipt" || task.Status != "blocked" {
		t.Fatalf("delivery request task must stay blocked config_rollout_receipt, task=%#v", task)
	}
	if task.Metadata["delivery_executor_contract_status"] != "blocked" {
		t.Fatalf("delivery task should expose blocked executor boundary status: %#v", task.Metadata)
	}
	for key, want := range map[string]string{
		"delivery_executor_registration_contract": "cmdb_agent_rollout_delivery_executor_registration_contract",
		"delivery_runner_identity_contract":       "cmdb_agent_rollout_delivery_runner_identity_contract",
		"delivery_attested_receipt_contract":      "cmdb_agent_rollout_delivery_attested_receipt_contract",
		"delivery_target_binding_contract":        "cmdb_agent_rollout_delivery_target_binding_contract",
		"delivery_request_ref_match_contract":     "cmdb_agent_rollout_delivery_request_ref_match_contract",
	} {
		if task.Metadata[key] != want {
			t.Fatalf("delivery task metadata should pin %s=%s, metadata=%#v", key, want, task.Metadata)
		}
	}
	missing := task.Metadata["delivery_executor_missing_contracts"]
	for _, want := range []string{
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_delivery_executor_registration_contract",
		"cmdb_agent_rollout_delivery_runner_identity_contract",
		"cmdb_agent_rollout_delivery_attested_receipt_contract",
		"cmdb_agent_rollout_delivery_target_binding_contract",
		"cmdb_agent_rollout_delivery_request_ref_match_contract",
	} {
		if !strings.Contains(missing, want) {
			t.Fatalf("delivery task missing contracts should include %q, metadata=%#v", want, task.Metadata)
		}
	}
	rawTask, _ := json.Marshal(task)
	assertNoWriterRequestRefRuntimeReadLeak(t, string(rawTask))
	assertNoConfigRolloutExecutionStates(t, string(rawTask))
	assertNoConfigRolloutSensitiveEcho(t, string(rawTask))
}

func TestFindXAgentConfigRolloutDispatchEffectRequestRefKeepsExecutorRegistrationGaps(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	if effectRef == "" {
		t.Fatalf("fixture should create effect_request_ref: %#v", rollout.Metadata)
	}

	task, ok, err := store.GetFindXAgentExecutionTask(effectRef)
	if err != nil || !ok {
		t.Fatalf("effect request_ref should resolve to blocked execution task, ok=%v err=%v", ok, err)
	}
	if task.Action != "config_rollout_receipt" || task.Status != "blocked" {
		t.Fatalf("effect request task must stay blocked config_rollout_receipt, task=%#v", task)
	}
	if task.Metadata["effect_executor_contract_status"] != "blocked" {
		t.Fatalf("effect task should expose blocked executor boundary status: %#v", task.Metadata)
	}
	for key, want := range map[string]string{
		"effect_executor_registration_contract": "cmdb_agent_rollout_effect_executor_registration_contract",
		"effect_runner_identity_contract":       "cmdb_agent_rollout_effect_runner_identity_contract",
		"effect_delivery_evidence_contract":     "cmdb_agent_rollout_effect_delivery_evidence_match_contract",
		"effect_attested_receipt_contract":      "cmdb_agent_rollout_effect_attested_receipt_contract",
		"effect_request_ref_match_contract":     "cmdb_agent_rollout_effect_request_ref_match_contract",
	} {
		if task.Metadata[key] != want {
			t.Fatalf("effect task metadata should pin %s=%s, metadata=%#v", key, want, task.Metadata)
		}
	}
	missing := task.Metadata["effect_executor_missing_contracts"]
	for _, want := range []string{
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_effect_executor_registration_contract",
		"cmdb_agent_rollout_effect_runner_identity_contract",
		"cmdb_agent_rollout_effect_delivery_evidence_match_contract",
		"cmdb_agent_rollout_effect_attested_receipt_contract",
		"cmdb_agent_rollout_effect_request_ref_match_contract",
	} {
		if !strings.Contains(missing, want) {
			t.Fatalf("effect task missing contracts should include %q, metadata=%#v", want, task.Metadata)
		}
	}
	rawTask, _ := json.Marshal(task)
	assertNoWriterRequestRefRuntimeReadLeak(t, string(rawTask))
	assertNoConfigRolloutExecutionStates(t, string(rawTask))
	assertNoConfigRolloutSensitiveEcho(t, string(rawTask))
}

func TestFindXAgentConfigRolloutDispatchRollbackRequestRefKeepsExecutorRegistrationGaps(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	rollbackRef := strings.TrimSpace(rollout.Metadata["rollback_request_ref"])
	if rollbackRef == "" {
		t.Fatalf("fixture should create rollback_request_ref: %#v", rollout.Metadata)
	}

	task, ok, err := store.GetFindXAgentExecutionTask(rollbackRef)
	if err != nil || !ok {
		t.Fatalf("rollback request_ref should resolve to blocked execution task, ok=%v err=%v", ok, err)
	}
	if task.Action != "config_rollout_receipt" || task.Status != "blocked" {
		t.Fatalf("rollback request task must stay blocked config_rollout_receipt, task=%#v", task)
	}
	if task.Metadata["rollback_executor_contract_status"] != "blocked" {
		t.Fatalf("rollback task should expose blocked executor boundary status: %#v", task.Metadata)
	}
	for key, want := range map[string]string{
		"rollback_executor_registration_contract": "cmdb_agent_rollout_rollback_executor_registration_contract",
		"rollback_runner_identity_contract":       "cmdb_agent_rollout_rollback_runner_identity_contract",
		"rollback_operation_context_contract":     "cmdb_agent_rollout_rollback_operation_context_contract",
		"rollback_attested_receipt_contract":      "cmdb_agent_rollout_rollback_attested_receipt_contract",
		"rollback_request_ref_match_contract":     "cmdb_agent_rollout_rollback_request_ref_match_contract",
	} {
		if task.Metadata[key] != want {
			t.Fatalf("rollback task metadata should pin %s=%s, metadata=%#v", key, want, task.Metadata)
		}
	}
	missing := task.Metadata["rollback_executor_missing_contracts"]
	for _, want := range []string{
		"cmdb_agent_rollout_rollback_executor_contract",
		"cmdb_agent_rollout_rollback_executor_registration_contract",
		"cmdb_agent_rollout_rollback_runner_identity_contract",
		"cmdb_agent_rollout_rollback_operation_context_contract",
		"cmdb_agent_rollout_rollback_attested_receipt_contract",
		"cmdb_agent_rollout_rollback_request_ref_match_contract",
	} {
		if !strings.Contains(missing, want) {
			t.Fatalf("rollback task missing contracts should include %q, metadata=%#v", want, task.Metadata)
		}
	}
	rawTask, _ := json.Marshal(task)
	assertNoWriterRequestRefRuntimeReadLeak(t, string(rawTask))
	assertNoConfigRolloutExecutionStates(t, string(rawTask))
	assertNoConfigRolloutSensitiveEcho(t, string(rawTask))
}

func TestFindXAgentConfigRolloutAssignPostDoesNotCreateDispatchReceiptTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	body := strings.NewReader(`{"template_id":"cmdb-host-plugin-assign","target_ids":["host-a"],"agent_ids":["agent-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"cmdb_host_probe","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"assign","rollback_ref":"rollback-ref","remote_mutation":false,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"cmdb_host","cmdb_host_ref":"host-a","agent_ref":"agent-a","dashboard_refs":"dashboard:redis-overview","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("assign post should stay blocked, code=%d payload=%#v", w.Code, payload)
	}
	for _, refKey := range []string{"delivery_request_ref", "effect_request_ref", "rollback_request_ref"} {
		if payload.Data.Metadata[refKey] != "" {
			t.Fatalf("assign rollout must not create dispatch receipt ref %s, metadata=%#v", refKey, payload.Data.Metadata)
		}
	}
	tasks, err := store.ListFindXAgentExecutionTasks()
	if err != nil {
		t.Fatalf("list execution tasks: %v", err)
	}
	for _, task := range tasks {
		if task.Action == "config_rollout_receipt" {
			t.Fatalf("assign rollout must not create dispatch receipt task, task=%#v", task)
		}
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutListKeepsBlockedRecordsWithoutOpeningReceipts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := saveConfigRolloutRuntimeReadRollout(t, map[string]string{})

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts", ListFindXAgentConfigRollouts)
	if w.Code != http.StatusOK {
		t.Fatalf("list should keep blocked records without opening receipt read success, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), rollout.ID) || !strings.Contains(w.Body.String(), `"status":"blocked"`) {
		t.Fatalf("list should return blocked rollout record only, body=%s", w.Body.String())
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailKeepsExecutorGapsAfterEvidenceChainReceipt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	rollbackRef := strings.TrimSpace(rollout.Metadata["rollback_request_ref"])
	if deliveryRef == "" || effectRef == "" || rollbackRef == "" {
		t.Fatalf("fixture should create delivery/effect/rollback request refs: %#v", rollout.Metadata)
	}
	saveConfigRolloutReceiptReceiverEvidence(t, rollout, deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "delivery", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "effect", effectRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "rollback", rollbackRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "data_arrival", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "evidence_chain", deliveryRef)

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail should keep resolved runtime read gate after full blocked receipt chain, got %d body=%s", detail.Code, detail.Body.String())
	}
	var item model.FindXAgentConfigRollout
	if err := json.Unmarshal(detail.Body.Bytes(), &item); err != nil {
		t.Fatalf("invalid rollout detail after full blocked chain: %v body=%s", err, detail.Body.String())
	}
	if item.Status != "blocked" {
		t.Fatalf("detail must stay blocked after evidence-chain receipt, item=%#v", item)
	}
	for _, want := range []string{
		"cmdb_agent_rollout_remote_writer_contract",
		"cmdb_agent_rollout_remote_executor_contract",
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_rollback_executor_contract",
	} {
		if !strings.Contains(detail.Body.String(), want) {
			t.Fatalf("detail after full blocked chain should expose executor gap %q, body=%s", want, detail.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, detail.Body.String())
	assertNoConfigRolloutExecutionStates(t, detail.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, detail.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailExposesRemoteExecutorAuthorizationEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	rollbackRef := strings.TrimSpace(rollout.Metadata["rollback_request_ref"])
	if deliveryRef == "" || effectRef == "" || rollbackRef == "" {
		t.Fatalf("fixture should create delivery/effect/rollback request refs: %#v", rollout.Metadata)
	}
	saveConfigRolloutReceiptReceiverEvidence(t, rollout, deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "delivery", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "effect", effectRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "rollback", rollbackRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "data_arrival", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "evidence_chain", deliveryRef)

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail should stay readable but blocked after full receipt chain, got %d body=%s", detail.Code, detail.Body.String())
	}
	for _, want := range cmdbAgentRolloutRemoteExecutorAuthorizationContractsForTest() {
		if !strings.Contains(detail.Body.String(), want) {
			t.Fatalf("detail should expose remote executor authorization gap %q, body=%s", want, detail.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, detail.Body.String())
	assertNoConfigRolloutExecutionStates(t, detail.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, detail.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailKeepsRealExecutorImplementationBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	rollbackRef := strings.TrimSpace(rollout.Metadata["rollback_request_ref"])
	if deliveryRef == "" || effectRef == "" || rollbackRef == "" {
		t.Fatalf("fixture should create delivery/effect/rollback request refs: %#v", rollout.Metadata)
	}
	saveConfigRolloutReceiptReceiverEvidence(t, rollout, deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "delivery", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "effect", effectRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "rollback", rollbackRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "data_arrival", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "evidence_chain", deliveryRef)

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail should stay readable but blocked until real executor authorization is approved, got %d body=%s", detail.Code, detail.Body.String())
	}
	for _, want := range cmdbAgentRolloutRealExecutorImplementationBoundaryContractsForTest() {
		if !strings.Contains(detail.Body.String(), want) {
			t.Fatalf("detail should expose real executor implementation boundary %q, body=%s", want, detail.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, detail.Body.String())
	assertNoConfigRolloutExecutionStates(t, detail.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, detail.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailAcceptsBlockedRemoteWriterRegistrationProof(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	saveConfigRolloutRemoteWriterRegistrationProof(t, rollout, "remote-writer-registration-ref")

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail should stay readable but blocked with remote writer registration proof, got %d body=%s", detail.Code, detail.Body.String())
	}
	var item model.FindXAgentConfigRollout
	if err := json.Unmarshal(detail.Body.Bytes(), &item); err != nil {
		t.Fatalf("invalid rollout detail with remote writer registration proof: %v body=%s", err, detail.Body.String())
	}
	missing := decodeConfigRolloutRuntimeMissingContracts(t, item.Metadata["runtime_read_missing_contracts"])
	for _, cleared := range []string{
		"cmdb_agent_rollout_remote_writer_contract",
		"cmdb_agent_rollout_remote_writer_registration_contract",
		"cmdb_agent_rollout_remote_writer_runner_identity_contract",
		"cmdb_agent_rollout_remote_writer_target_binding_contract",
		"cmdb_agent_rollout_remote_writer_attested_receipt_contract",
	} {
		if containsLifecycleTestString(missing, cleared) {
			t.Fatalf("remote writer registration proof should clear %q, missing=%#v detail=%s", cleared, missing, detail.Body.String())
		}
	}
	for _, stillMissing := range []string{
		"cmdb_agent_rollout_remote_writer_credential_policy_release_contract",
		"cmdb_agent_rollout_remote_executor_contract",
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_rollback_executor_contract",
	} {
		if !containsLifecycleTestString(missing, stillMissing) {
			t.Fatalf("remote writer registration proof must not clear executor gap %q, missing=%#v detail=%s", stillMissing, missing, detail.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, detail.Body.String())
	assertNoConfigRolloutExecutionStates(t, detail.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, detail.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailAcceptsBlockedDeliveryExecutorRegistrationProof(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	if deliveryRef == "" {
		t.Fatalf("fixture should create delivery_request_ref: %#v", rollout.Metadata)
	}
	saveConfigRolloutDeliveryExecutorRegistrationProof(t, rollout, deliveryRef, "delivery-executor-registration-ref")

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail should stay readable but blocked with delivery executor registration proof, got %d body=%s", detail.Code, detail.Body.String())
	}
	var item model.FindXAgentConfigRollout
	if err := json.Unmarshal(detail.Body.Bytes(), &item); err != nil {
		t.Fatalf("invalid rollout detail with delivery executor registration proof: %v body=%s", err, detail.Body.String())
	}
	missing := decodeConfigRolloutRuntimeMissingContracts(t, item.Metadata["runtime_read_missing_contracts"])
	for _, cleared := range []string{
		"cmdb_agent_rollout_delivery_executor_registration_contract",
		"cmdb_agent_rollout_delivery_runner_identity_contract",
		"cmdb_agent_rollout_delivery_target_binding_contract",
		"cmdb_agent_rollout_delivery_attested_receipt_contract",
		"cmdb_agent_rollout_delivery_request_ref_match_contract",
	} {
		if containsLifecycleTestString(missing, cleared) {
			t.Fatalf("delivery executor registration proof should clear %q, missing=%#v detail=%s", cleared, missing, detail.Body.String())
		}
	}
	for _, stillMissing := range []string{
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_rollback_executor_contract",
		"cmdb_agent_rollout_remote_writer_credential_policy_release_contract",
	} {
		if !containsLifecycleTestString(missing, stillMissing) {
			t.Fatalf("delivery executor registration proof must not clear executor gap %q, missing=%#v detail=%s", stillMissing, missing, detail.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, detail.Body.String())
	assertNoConfigRolloutExecutionStates(t, detail.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, detail.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailExposesDeliveryExecutorRegistrationBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail should stay readable but blocked without delivery executor registration, got %d body=%s", detail.Code, detail.Body.String())
	}
	var item model.FindXAgentConfigRollout
	if err := json.Unmarshal(detail.Body.Bytes(), &item); err != nil {
		t.Fatalf("invalid rollout detail without delivery executor registration proof: %v body=%s", err, detail.Body.String())
	}
	missing := decodeConfigRolloutRuntimeMissingContracts(t, item.Metadata["runtime_read_missing_contracts"])
	for _, want := range []string{
		"cmdb_agent_rollout_delivery_executor_registration_contract",
		"cmdb_agent_rollout_delivery_runner_identity_contract",
		"cmdb_agent_rollout_delivery_target_binding_contract",
		"cmdb_agent_rollout_delivery_attested_receipt_contract",
		"cmdb_agent_rollout_delivery_request_ref_match_contract",
	} {
		if !containsLifecycleTestString(missing, want) {
			t.Fatalf("detail should expose delivery executor registration boundary %q, missing=%#v detail=%s", want, missing, detail.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, detail.Body.String())
	assertNoConfigRolloutExecutionStates(t, detail.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, detail.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailAcceptsBlockedEffectExecutorRegistrationProof(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	if effectRef == "" || deliveryRef == "" {
		t.Fatalf("fixture should create delivery/effect request refs: %#v", rollout.Metadata)
	}
	saveConfigRolloutEffectExecutorRegistrationProof(t, rollout, effectRef, deliveryRef, "effect-executor-registration-ref")

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail should stay readable but blocked with effect executor registration proof, got %d body=%s", detail.Code, detail.Body.String())
	}
	var item model.FindXAgentConfigRollout
	if err := json.Unmarshal(detail.Body.Bytes(), &item); err != nil {
		t.Fatalf("invalid rollout detail with effect executor registration proof: %v body=%s", err, detail.Body.String())
	}
	missing := decodeConfigRolloutRuntimeMissingContracts(t, item.Metadata["runtime_read_missing_contracts"])
	for _, cleared := range []string{
		"cmdb_agent_rollout_effect_executor_registration_contract",
		"cmdb_agent_rollout_effect_runner_identity_contract",
		"cmdb_agent_rollout_effect_delivery_evidence_match_contract",
		"cmdb_agent_rollout_effect_attested_receipt_contract",
		"cmdb_agent_rollout_effect_request_ref_match_contract",
	} {
		if containsLifecycleTestString(missing, cleared) {
			t.Fatalf("effect executor registration proof should clear %q, missing=%#v detail=%s", cleared, missing, detail.Body.String())
		}
	}
	for _, stillMissing := range []string{
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_rollback_executor_contract",
		"cmdb_agent_rollout_remote_writer_credential_policy_release_contract",
	} {
		if !containsLifecycleTestString(missing, stillMissing) {
			t.Fatalf("effect executor registration proof must not clear executor gap %q, missing=%#v detail=%s", stillMissing, missing, detail.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, detail.Body.String())
	assertNoConfigRolloutExecutionStates(t, detail.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, detail.Body.String())
}

func TestFindXAgentConfigRolloutDispatchDetailExposesEffectExecutorRegistrationBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail should stay readable but blocked without effect executor registration, got %d body=%s", detail.Code, detail.Body.String())
	}
	var item model.FindXAgentConfigRollout
	if err := json.Unmarshal(detail.Body.Bytes(), &item); err != nil {
		t.Fatalf("invalid rollout detail without effect executor registration proof: %v body=%s", err, detail.Body.String())
	}
	missing := decodeConfigRolloutRuntimeMissingContracts(t, item.Metadata["runtime_read_missing_contracts"])
	for _, want := range []string{
		"cmdb_agent_rollout_effect_executor_registration_contract",
		"cmdb_agent_rollout_effect_runner_identity_contract",
		"cmdb_agent_rollout_effect_delivery_evidence_match_contract",
		"cmdb_agent_rollout_effect_attested_receipt_contract",
		"cmdb_agent_rollout_effect_request_ref_match_contract",
	} {
		if !containsLifecycleTestString(missing, want) {
			t.Fatalf("detail should expose effect executor registration boundary %q, missing=%#v detail=%s", want, missing, detail.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, detail.Body.String())
	assertNoConfigRolloutExecutionStates(t, detail.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, detail.Body.String())
}

func saveConfigRolloutRuntimeReadRollout(t *testing.T, metadata map[string]string) model.FindXAgentConfigRollout {
	t.Helper()
	base := map[string]string{
		"scope":         "cmdb_host",
		"cmdb_host_ref": "host-a",
		"agent_ref":     "agent-a",
	}
	for key, value := range metadata {
		base[key] = value
	}
	rollout, err := store.SaveFindXAgentConfigRollout(model.FindXAgentConfigRollout{
		TemplateID:      "cmdb-host-plugin-dispatch",
		TargetIDs:       []string{"host-a"},
		AgentIDs:        []string{"agent-a"},
		ConfigVersion:   "cfg-v1",
		ProviderMode:    "cmdb_host_probe",
		PluginID:        "redis",
		RolloutStrategy: "dispatch",
		RemoteMutation:  true,
		Status:          "blocked",
		Blocker:         "PENDING: dispatch read",
		Metadata:        base,
	})
	if err != nil {
		t.Fatalf("save dispatch rollout: %v", err)
	}
	return rollout
}

func saveConfigRolloutRemoteWriterRegistrationProof(t *testing.T, rollout model.FindXAgentConfigRollout, ref string) {
	t.Helper()
	rollout.Metadata["remote_writer_registration_ref"] = ref
	if _, err := store.SaveFindXAgentConfigRollout(rollout); err != nil {
		t.Fatalf("save rollout remote writer registration ref: %v", err)
	}
	_, err := store.SaveFindXAgentExecutionTask(model.FindXAgentExecutionTask{
		ID:        ref,
		Action:    "remote_writer_registration",
		AgentIDs:  rollout.AgentIDs,
		TargetIDs: rollout.TargetIDs,
		Status:    "blocked",
		Blocker:   "PENDING: remote writer registration is attested but dispatch execution is not authorized",
		Audit:     "findx_agent.remote_writer.registration.requested",
		Metadata: map[string]string{
			"source_rollout_id":     rollout.ID,
			"config_rollout_id":     rollout.ID,
			"rollout_ref":           rollout.ID,
			"phase":                 "remote_writer_registration",
			"receipt_kind":          "remote_writer_registration",
			"plugin_id":             rollout.PluginID,
			"runner_identity_ref":   "runner-identity-ref",
			"target_binding_ref":    rollout.Metadata["target_binding_ref"],
			"attested_receipt_ref":  "attested-blocked-receipt-ref",
			"attested_receipt_kind": "blocked",
		},
	})
	if err != nil {
		t.Fatalf("save remote writer registration proof: %v", err)
	}
}

func saveConfigRolloutDeliveryExecutorRegistrationProof(t *testing.T, rollout model.FindXAgentConfigRollout, requestRef, ref string) {
	t.Helper()
	rollout.Metadata["delivery_executor_registration_ref"] = ref
	if _, err := store.SaveFindXAgentConfigRollout(rollout); err != nil {
		t.Fatalf("save rollout delivery executor registration ref: %v", err)
	}
	_, err := store.SaveFindXAgentExecutionTask(model.FindXAgentExecutionTask{
		ID:        ref,
		Action:    "delivery_executor_registration",
		AgentIDs:  rollout.AgentIDs,
		TargetIDs: rollout.TargetIDs,
		Status:    "blocked",
		Blocker:   "PENDING: delivery executor registration is attested but delivery execution is not authorized",
		Audit:     "findx_agent.delivery_executor.registration.requested",
		Metadata: map[string]string{
			"source_rollout_id":     rollout.ID,
			"config_rollout_id":     rollout.ID,
			"rollout_ref":           rollout.ID,
			"phase":                 "delivery_executor_registration",
			"receipt_kind":          "delivery",
			"plugin_id":             rollout.PluginID,
			"request_ref":           requestRef,
			"runner_identity_ref":   "delivery-runner-identity-ref",
			"target_binding_ref":    rollout.Metadata["target_binding_ref"],
			"attested_receipt_ref":  "delivery-attested-blocked-receipt-ref",
			"attested_receipt_kind": "blocked",
		},
	})
	if err != nil {
		t.Fatalf("save delivery executor registration proof: %v", err)
	}
}

func saveConfigRolloutEffectExecutorRegistrationProof(t *testing.T, rollout model.FindXAgentConfigRollout, requestRef, deliveryRef, ref string) {
	t.Helper()
	rollout.Metadata["effect_executor_registration_ref"] = ref
	if _, err := store.SaveFindXAgentConfigRollout(rollout); err != nil {
		t.Fatalf("save rollout effect executor registration ref: %v", err)
	}
	_, err := store.SaveFindXAgentExecutionTask(model.FindXAgentExecutionTask{
		ID:        ref,
		Action:    "effect_executor_registration",
		AgentIDs:  rollout.AgentIDs,
		TargetIDs: rollout.TargetIDs,
		Status:    "blocked",
		Blocker:   "PENDING: effect executor registration is attested but effect execution is not authorized",
		Audit:     "findx_agent.effect_executor.registration.requested",
		Metadata: map[string]string{
			"source_rollout_id":     rollout.ID,
			"config_rollout_id":     rollout.ID,
			"rollout_ref":           rollout.ID,
			"phase":                 "effect_executor_registration",
			"receipt_kind":          "effect",
			"plugin_id":             rollout.PluginID,
			"request_ref":           requestRef,
			"delivery_evidence_ref": deliveryRef,
			"runner_identity_ref":   "effect-runner-identity-ref",
			"target_binding_ref":    rollout.Metadata["target_binding_ref"],
			"attested_receipt_ref":  "effect-attested-blocked-receipt-ref",
			"attested_receipt_kind": "blocked",
		},
	})
	if err != nil {
		t.Fatalf("save effect executor registration proof: %v", err)
	}
}

func cmdbAgentRolloutRemoteExecutorAuthorizationContractsForTest() []string {
	return []string{
		"cmdb_agent_rollout_remote_writer_registration_contract",
		"cmdb_agent_rollout_remote_writer_runner_identity_contract",
		"cmdb_agent_rollout_remote_writer_target_binding_contract",
		"cmdb_agent_rollout_remote_writer_credential_policy_release_contract",
		"cmdb_agent_rollout_remote_writer_attested_receipt_contract",
		"cmdb_agent_rollout_remote_executor_registration_contract",
		"cmdb_agent_rollout_remote_executor_runner_identity_contract",
		"cmdb_agent_rollout_remote_executor_target_binding_contract",
		"cmdb_agent_rollout_remote_executor_credential_policy_release_contract",
		"cmdb_agent_rollout_remote_executor_attested_receipt_contract",
	}
}

func cmdbAgentRolloutRealExecutorImplementationBoundaryContractsForTest() []string {
	return []string{
		"cmdb_agent_rollout_executor_target_scope_authorization_contract",
		"cmdb_agent_rollout_remote_execution_method_authorization_contract",
		"cmdb_agent_rollout_runner_identity_format_contract",
		"cmdb_agent_rollout_executor_registration_store_contract",
		"cmdb_agent_rollout_attested_receipt_schema_contract",
		"cmdb_agent_rollout_credential_policy_release_rule_contract",
		"cmdb_agent_rollout_rollback_failure_boundary_contract",
	}
}

func decodeConfigRolloutRuntimeMissingContracts(t *testing.T, raw string) []string {
	t.Helper()
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		t.Fatalf("invalid runtime_read_missing_contracts JSON: %v raw=%s", err, raw)
	}
	return values
}

func saveConfigRolloutRuntimeReadTask(t *testing.T, id string, rollout model.FindXAgentConfigRollout, receipt string, overrides map[string]string) {
	t.Helper()
	metadata := map[string]string{
		"source_rollout_id": rollout.ID,
		"receipt_kind":      receipt,
		"plugin_id":         rollout.PluginID,
	}
	for key, value := range overrides {
		metadata[key] = value
	}
	_, err := store.SaveFindXAgentExecutionTask(model.FindXAgentExecutionTask{
		ID:        id,
		Action:    "config_rollout_receipt",
		AgentIDs:  []string{"agent-a"},
		TargetIDs: []string{"host-a"},
		Status:    "blocked",
		Blocker:   "PENDING: receipt task blocked",
		Metadata:  metadata,
	})
	if err != nil {
		t.Fatalf("save receipt task %s: %v", id, err)
	}
}

func assertNoWriterRequestRefRuntimeReadLeak(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{
		"writer_request_ref",
		"cmdb_agent_rollout_writer_request_ref_contract",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("runtime read response leaked writer request ref marker %q: %s", forbidden, body)
		}
	}
}

func decodeConfigRolloutRuntimeReadBlocked(t *testing.T, w *httptest.ResponseRecorder) struct {
	Error            string         `json:"error"`
	Status           string         `json:"status"`
	Contract         string         `json:"contract"`
	MissingContracts []string       `json:"missing_contracts"`
	ReceiptContract  map[string]any `json:"receipt_contract"`
	RolloutRef       string         `json:"rollout_ref"`
	SafeToRetry      bool           `json:"safe_to_retry"`
} {
	t.Helper()
	var payload struct {
		Error            string         `json:"error"`
		Status           string         `json:"status"`
		Contract         string         `json:"contract"`
		MissingContracts []string       `json:"missing_contracts"`
		ReceiptContract  map[string]any `json:"receipt_contract"`
		RolloutRef       string         `json:"rollout_ref"`
		SafeToRetry      bool           `json:"safe_to_retry"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid runtime read blocked envelope: %v body=%s", err, w.Body.String())
	}
	return payload
}
