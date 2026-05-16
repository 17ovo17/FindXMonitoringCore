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

func TestFindXAgentDataArrivalEvidenceCmdbDispatchReadRequiresResolvedRequestRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := saveConfigRolloutRuntimeReadRollout(t, map[string]string{
		"delivery_request_ref": "missing-delivery-task",
		"effect_request_ref":   "missing-effect-task",
		"rollback_request_ref": "missing-rollback-task",
	})

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID, ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadBlocked(t, w)

	if w.Code != http.StatusConflict || payload.Status != "blocked" || payload.Contract != cmdbAgentRolloutDataArrivalReadContract {
		t.Fatalf("data-arrival runtime read with stale rollout refs should be blocked, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	for _, want := range []string{
		cmdbAgentRolloutRequestRefResolveContract,
		cmdbAgentRolloutExecutionTaskContract,
		cmdbAgentRolloutDataArrivalEvidenceContract,
		cmdbAgentRolloutReceiverEvidenceContract,
	} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("blocked runtime read should include %q, payload=%#v", want, payload)
		}
	}
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceCmdbDispatchReadBlocksWithoutReceiverEvidence(t *testing.T) {
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

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID, ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadBlocked(t, w)

	if w.Code != http.StatusConflict || payload.Status != "blocked" {
		t.Fatalf("data-arrival runtime read without receiver evidence should be blocked, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	for _, want := range []string{
		cmdbAgentRolloutDataArrivalEvidenceContract,
		cmdbAgentRolloutReceiverEvidenceContract,
		cmdbAgentRolloutDataArrivalRequestRefContract,
	} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("missing receiver evidence should include %q, payload=%#v", want, payload)
		}
	}
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceCmdbDispatchReadRejectsMismatchedRequestRef(t *testing.T) {
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
	saveConfigRolloutRuntimeReadTask(t, "foreign-task", rollout, "delivery", map[string]string{"source_rollout_id": "other-rollout"})

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref=foreign-task", ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadBlocked(t, w)

	if w.Code != http.StatusConflict || !containsLifecycleTestString(payload.MissingContracts, cmdbAgentRolloutTaskMatchContract) {
		t.Fatalf("mismatched request_ref should be blocked by task match contract, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceCmdbDispatchReadReturnsReceiverEvidenceWithContractGaps(t *testing.T) {
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
	if _, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindMetrics,
		AgentID:      "agent-a",
		TargetID:     "host-a",
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/metrics/write-compatible"},
		Metadata: map[string]string{
			"source_rollout_id": rollout.ID,
			"request_ref":       "delivery-task",
			"plugin_id":         rollout.PluginID,
			"state":             "succeeded",
			"password":          "do-not-store",
		},
	}); err != nil {
		t.Fatalf("save data-arrival evidence: %v", err)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref=delivery-task", ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadOK(t, w)

	if w.Code != http.StatusOK || payload.Status != "blocked" || payload.Contract != cmdbAgentRolloutDataArrivalReadContract {
		t.Fatalf("matching receiver evidence should be returned with blocked contract gaps, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	if payload.EvidenceCount != 1 || payload.RolloutRef != rollout.ID {
		t.Fatalf("matching runtime read should expose exactly one evidence row for the rollout, payload=%#v", payload)
	}
	for _, want := range []string{
		"cmdb_agent_rollout_remote_executor_contract",
		"cmdb_agent_rollout_delivery_receipt_contract",
		"cmdb_agent_rollout_effect_receipt_contract",
		"cmdb_agent_rollout_rollback_receipt_contract",
	} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("receiver evidence response should keep execution gaps %q, payload=%#v", want, payload)
		}
	}
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
	if !strings.Contains(w.Body.String(), `"request_ref":"delivery-task"`) {
		t.Fatalf("safe request_ref should be visible for audit correlation, body=%s", w.Body.String())
	}
}

func TestFindXAgentDataArrivalEvidenceCmdbDispatchReadDoesNotRequireWriterRequestRef(t *testing.T) {
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
	if _, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindMetrics,
		AgentID:      "agent-a",
		TargetID:     "host-a",
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/metrics/write-compatible"},
		Metadata: map[string]string{
			"source_rollout_id": rollout.ID,
			"request_ref":       "delivery-task",
			"plugin_id":         rollout.PluginID,
		},
	}); err != nil {
		t.Fatalf("save data-arrival evidence: %v", err)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref=delivery-task", ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadOK(t, w)

	if w.Code != http.StatusOK || payload.Status != "blocked" || payload.Contract != cmdbAgentRolloutDataArrivalReadContract {
		t.Fatalf("data-arrival read should not require writer_request_ref, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	if payload.EvidenceCount != 1 || payload.RolloutRef != rollout.ID {
		t.Fatalf("data-arrival read should return receiver evidence for rollout, payload=%#v", payload)
	}
	for _, want := range []string{
		"cmdb_agent_rollout_remote_executor_contract",
		"cmdb_agent_rollout_delivery_receipt_contract",
		"cmdb_agent_rollout_effect_receipt_contract",
		"cmdb_agent_rollout_rollback_receipt_contract",
		"cmdb_agent_rollout_data_arrival_receipt_contract",
	} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("data-arrival read should keep real execution gap %q, payload=%#v", want, payload)
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, w.Body.String())
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceCmdbDispatchReadKeepsExecutorGapsAfterBlockedReceipts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	store.ResetCmdbApprovalRiskForTest()
	credentialID := "credential-fx-night-148"
	approvalID := "approval-fx-night-148"
	saveFindXAgentLifecycleCredentialForTest(credentialID)
	assignBody := strings.NewReader(cmdbCredentialRolloutBody("assign", credentialID, "", false))
	assignResp := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", assignBody, CreateFindXAgentConfigRollout)
	assignPayload := decodeConfigRolloutEnvelope(t, assignResp)
	assignmentRef, _ := assignPayload.OperationContract["assignment_ref"].(string)
	if assignmentRef == "" {
		t.Fatalf("assign should create assignment_ref before dispatch: %#v", assignPayload.OperationContract)
	}
	saveCmdbCredentialPolicyApprovalForTest(t, approvalID, "dispatch", nil, "review_accept_recorded")
	dispatchBody := strings.NewReader(cmdbCredentialRolloutBodyWithMetadata("dispatch", credentialID, assignmentRef, true, `"scope_policy_approval_ref":"`+approvalID+`"`))
	dispatchResp := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", dispatchBody, CreateFindXAgentConfigRollout)
	dispatchPayload := decodeConfigRolloutEnvelope(t, dispatchResp)
	if dispatchResp.Code != http.StatusConflict || dispatchPayload.Data.Status != "blocked" {
		t.Fatalf("dispatch fixture should stay blocked while refs resolve, code=%d payload=%#v body=%s", dispatchResp.Code, dispatchPayload, dispatchResp.Body.String())
	}
	assertConfigRolloutCredentialPolicyReleased(t, dispatchPayload, true)
	rollout := dispatchPayload.Data
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	rollbackRef := strings.TrimSpace(rollout.Metadata["rollback_request_ref"])
	for receipt, ref := range map[string]string{"delivery": deliveryRef, "effect": effectRef, "rollback": rollbackRef} {
		if ref == "" {
			t.Fatalf("dispatch rollout should create %s request_ref: %#v", receipt, rollout.Metadata)
		}
		receiptPayload := `{
			"rollout_id":"` + rollout.ID + `",
			"receipt_type":"` + receipt + `",
			"request_ref":"` + ref + `",
			"status":"PENDING",
			"contract_id":"` + configRolloutReceiptIngestionContractID(receipt, "") + `",
			"missing_contracts":["cmdb_agent_rollout_` + receipt + `_executor_contract"],
			"evidence_ref":"fx-night-148-` + receipt + `-receipt"
		}`
		receiptResp := performConfigRolloutReceiptIngestionPost(rollout.ID, receiptPayload)
		if receiptResp.Code != http.StatusConflict {
			t.Fatalf("%s receipt ingestion should remain blocked, code=%d body=%s", receipt, receiptResp.Code, receiptResp.Body.String())
		}
		assertNoConfigRolloutReceiptIngestionForbidden(t, receiptResp.Body.String())
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
		},
	}); err != nil {
		t.Fatalf("save data-arrival evidence: %v", err)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref="+deliveryRef, ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadOK(t, w)

	if w.Code != http.StatusOK || payload.Status != "blocked" || payload.EvidenceCount != 1 {
		t.Fatalf("blocked receipts and receiver evidence must still return blocked data-arrival envelope, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	for _, want := range []string{
		"cmdb_agent_rollout_remote_executor_contract",
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_rollback_executor_contract",
		"cmdb_agent_rollout_delivery_receipt_contract",
		"cmdb_agent_rollout_effect_receipt_contract",
		"cmdb_agent_rollout_rollback_receipt_contract",
		"cmdb_agent_rollout_data_arrival_receipt_contract",
		"cmdb_agent_rollout_evidence_chain_contract",
	} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("data-arrival read after blocked receipts should keep execution gap %q, payload=%#v body=%s", want, payload, w.Body.String())
		}
	}
	if strings.Contains(w.Body.String(), credentialID) {
		t.Fatalf("data-arrival read must not echo credential ref/id value %q: %s", credentialID, w.Body.String())
	}
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceKeepsExecutorGapsAfterEvidenceChainReceipt(t *testing.T) {
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

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref="+deliveryRef, ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadOK(t, w)

	if w.Code != http.StatusOK || payload.Status != "blocked" || payload.EvidenceCount != 1 {
		t.Fatalf("evidence-chain receipt must still read as blocked evidence, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	for _, want := range []string{
		"cmdb_agent_rollout_remote_writer_contract",
		"cmdb_agent_rollout_remote_executor_contract",
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_rollback_executor_contract",
	} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("data-arrival read after evidence-chain receipt should keep executor gap %q, payload=%#v body=%s", want, payload, w.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, w.Body.String())
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceExposesRemoteExecutorAuthorizationEnvelope(t *testing.T) {
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

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref="+deliveryRef, ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadOK(t, w)
	if w.Code != http.StatusOK || payload.Status != "blocked" || payload.EvidenceCount != 1 {
		t.Fatalf("data-arrival evidence should stay readable but blocked, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	for _, want := range cmdbAgentRolloutRemoteExecutorAuthorizationContractsForTest() {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("data-arrival evidence should expose remote executor authorization gap %q, payload=%#v body=%s", want, payload, w.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, w.Body.String())
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceKeepsRealExecutorImplementationBoundary(t *testing.T) {
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

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref="+deliveryRef, ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadOK(t, w)
	if w.Code != http.StatusOK || payload.Status != "blocked" || payload.EvidenceCount != 1 {
		t.Fatalf("data-arrival evidence should stay readable but blocked, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	for _, want := range cmdbAgentRolloutRealExecutorImplementationBoundaryContractsForTest() {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("data-arrival evidence should expose real executor implementation boundary %q, payload=%#v body=%s", want, payload, w.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, w.Body.String())
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceAcceptsBlockedRemoteWriterRegistrationProof(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	rollbackRef := strings.TrimSpace(rollout.Metadata["rollback_request_ref"])
	if deliveryRef == "" || effectRef == "" || rollbackRef == "" {
		t.Fatalf("fixture should create delivery/effect/rollback request refs: %#v", rollout.Metadata)
	}
	saveConfigRolloutRemoteWriterRegistrationProof(t, rollout, "remote-writer-registration-ref")
	rollout, _, _ = store.GetFindXAgentConfigRollout(rollout.ID)
	saveConfigRolloutReceiptReceiverEvidence(t, rollout, deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "delivery", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "effect", effectRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "rollback", rollbackRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "data_arrival", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "evidence_chain", deliveryRef)

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref="+deliveryRef, ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadOK(t, w)
	if w.Code != http.StatusOK || payload.Status != "blocked" || payload.EvidenceCount != 1 {
		t.Fatalf("data-arrival evidence should stay readable but blocked, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	for _, cleared := range []string{
		"cmdb_agent_rollout_remote_writer_contract",
		"cmdb_agent_rollout_remote_writer_registration_contract",
		"cmdb_agent_rollout_remote_writer_runner_identity_contract",
		"cmdb_agent_rollout_remote_writer_target_binding_contract",
		"cmdb_agent_rollout_remote_writer_attested_receipt_contract",
	} {
		if containsLifecycleTestString(payload.MissingContracts, cleared) {
			t.Fatalf("remote writer registration proof should clear %q, payload=%#v body=%s", cleared, payload, w.Body.String())
		}
	}
	for _, stillMissing := range []string{
		"cmdb_agent_rollout_remote_writer_credential_policy_release_contract",
		"cmdb_agent_rollout_remote_executor_contract",
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_rollback_executor_contract",
	} {
		if !containsLifecycleTestString(payload.MissingContracts, stillMissing) {
			t.Fatalf("remote writer registration proof must not clear executor gap %q, payload=%#v body=%s", stillMissing, payload, w.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, w.Body.String())
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceAcceptsBlockedDeliveryExecutorRegistrationProof(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	rollbackRef := strings.TrimSpace(rollout.Metadata["rollback_request_ref"])
	if deliveryRef == "" || effectRef == "" || rollbackRef == "" {
		t.Fatalf("fixture should create delivery/effect/rollback request refs: %#v", rollout.Metadata)
	}
	saveConfigRolloutDeliveryExecutorRegistrationProof(t, rollout, deliveryRef, "delivery-executor-registration-ref")
	rollout, _, _ = store.GetFindXAgentConfigRollout(rollout.ID)
	saveConfigRolloutReceiptReceiverEvidence(t, rollout, deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "delivery", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "effect", effectRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "rollback", rollbackRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "data_arrival", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "evidence_chain", deliveryRef)

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref="+deliveryRef, ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadOK(t, w)
	if w.Code != http.StatusOK || payload.Status != "blocked" || payload.EvidenceCount != 1 {
		t.Fatalf("data-arrival evidence should stay readable but blocked, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	for _, cleared := range []string{
		"cmdb_agent_rollout_delivery_executor_registration_contract",
		"cmdb_agent_rollout_delivery_runner_identity_contract",
		"cmdb_agent_rollout_delivery_target_binding_contract",
		"cmdb_agent_rollout_delivery_attested_receipt_contract",
		"cmdb_agent_rollout_delivery_request_ref_match_contract",
	} {
		if containsLifecycleTestString(payload.MissingContracts, cleared) {
			t.Fatalf("delivery executor registration proof should clear %q, payload=%#v body=%s", cleared, payload, w.Body.String())
		}
	}
	for _, stillMissing := range []string{
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_rollback_executor_contract",
	} {
		if !containsLifecycleTestString(payload.MissingContracts, stillMissing) {
			t.Fatalf("delivery executor registration proof must not clear executor gap %q, payload=%#v body=%s", stillMissing, payload, w.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, w.Body.String())
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceAcceptsBlockedEffectExecutorRegistrationProof(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	effectRef := strings.TrimSpace(rollout.Metadata["effect_request_ref"])
	rollbackRef := strings.TrimSpace(rollout.Metadata["rollback_request_ref"])
	if deliveryRef == "" || effectRef == "" || rollbackRef == "" {
		t.Fatalf("fixture should create delivery/effect/rollback request refs: %#v", rollout.Metadata)
	}
	saveConfigRolloutEffectExecutorRegistrationProof(t, rollout, effectRef, deliveryRef, "effect-executor-registration-ref")
	rollout, _, _ = store.GetFindXAgentConfigRollout(rollout.ID)
	saveConfigRolloutReceiptReceiverEvidence(t, rollout, deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "delivery", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "effect", effectRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "rollback", rollbackRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "data_arrival", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "evidence_chain", deliveryRef)

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref="+deliveryRef, ListFindXAgentDataArrivalEvidence)
	payload := decodeDataArrivalRuntimeReadOK(t, w)
	if w.Code != http.StatusOK || payload.Status != "blocked" || payload.EvidenceCount != 1 {
		t.Fatalf("data-arrival evidence should stay readable but blocked, code=%d payload=%#v body=%s", w.Code, payload, w.Body.String())
	}
	for _, cleared := range []string{
		"cmdb_agent_rollout_effect_executor_registration_contract",
		"cmdb_agent_rollout_effect_runner_identity_contract",
		"cmdb_agent_rollout_effect_delivery_evidence_match_contract",
		"cmdb_agent_rollout_effect_attested_receipt_contract",
		"cmdb_agent_rollout_effect_request_ref_match_contract",
	} {
		if containsLifecycleTestString(payload.MissingContracts, cleared) {
			t.Fatalf("effect executor registration proof should clear %q, payload=%#v body=%s", cleared, payload, w.Body.String())
		}
	}
	for _, stillMissing := range []string{
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_rollback_executor_contract",
	} {
		if !containsLifecycleTestString(payload.MissingContracts, stillMissing) {
			t.Fatalf("effect executor registration proof must not clear executor gap %q, payload=%#v body=%s", stillMissing, payload, w.Body.String())
		}
	}
	assertNoWriterRequestRefRuntimeReadLeak(t, w.Body.String())
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func TestFindXAgentDataArrivalEvidenceRuntimeReadKeepsLegacyList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	if _, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindLogs,
		AgentID:      "agent-legacy-list",
		TargetID:     "host-legacy-list",
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/findx-agent/logs-compatible"},
		Metadata:     map[string]string{"source": "legacy-list"},
	}); err != nil {
		t.Fatalf("save legacy evidence: %v", err)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence", ListFindXAgentDataArrivalEvidence)
	if w.Code != http.StatusOK {
		t.Fatalf("legacy evidence list should stay available, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "agent-legacy-list") || strings.Contains(w.Body.String(), cmdbAgentRolloutDataArrivalReadContract) {
		t.Fatalf("legacy evidence list should return raw list without runtime-read envelope, body=%s", w.Body.String())
	}
	assertNoDataArrivalRuntimeReadForbidden(t, w.Body.String())
}

func decodeDataArrivalRuntimeReadBlocked(t *testing.T, w *httptest.ResponseRecorder) struct {
	Status           string   `json:"status"`
	Contract         string   `json:"contract"`
	MissingContracts []string `json:"missing_contracts"`
	RolloutRef       string   `json:"rollout_ref"`
} {
	t.Helper()
	var payload struct {
		Status           string   `json:"status"`
		Contract         string   `json:"contract"`
		MissingContracts []string `json:"missing_contracts"`
		RolloutRef       string   `json:"rollout_ref"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid data-arrival runtime read blocked envelope: %v body=%s", err, w.Body.String())
	}
	return payload
}

func decodeDataArrivalRuntimeReadOK(t *testing.T, w *httptest.ResponseRecorder) struct {
	Status           string   `json:"status"`
	Contract         string   `json:"contract"`
	MissingContracts []string `json:"missing_contracts"`
	RolloutRef       string   `json:"rollout_ref"`
	EvidenceCount    int      `json:"evidence_count"`
} {
	t.Helper()
	var payload struct {
		Status           string   `json:"status"`
		Contract         string   `json:"contract"`
		MissingContracts []string `json:"missing_contracts"`
		RolloutRef       string   `json:"rollout_ref"`
		EvidenceCount    int      `json:"evidence_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid data-arrival runtime read envelope: %v body=%s", err, w.Body.String())
	}
	return payload
}

func assertNoDataArrivalRuntimeReadForbidden(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{
		`"status":"queued"`,
		`"status":"running"`,
		`"status":"succeeded"`,
		`"status":"success"`,
		`"status":"applied"`,
		`"status":"installed"`,
		`"status":"delivered"`,
		`"status":"effective"`,
		`"status":"data_arrived"`,
		`"status":"service_registered"`,
		"password",
		"do-not-store",
		"token",
		"cookie",
		"dsn",
		"private_key",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("data-arrival runtime read leaked forbidden marker %q: %s", forbidden, body)
		}
	}
}
