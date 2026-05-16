package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentConfigRolloutReceiptIngestionRecordsBlockedEvidenceChainAttestation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	if deliveryRef == "" {
		t.Fatalf("fixture should create delivery_request_ref: %#v", rollout.Metadata)
	}
	saveConfigRolloutReceiptReceiverEvidence(t, rollout, deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "delivery", deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "data_arrival", deliveryRef)

	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"evidence_chain",
		"request_ref":"` + deliveryRef + `",
		"status":"PENDING",
		"contract_id":"cmdb_agent_rollout_evidence_chain_contract",
		"missing_contracts":["cmdb_agent_rollout_evidence_chain_contract"],
		"evidence_ref":"evidence-chain-safe-ref"
	}`
	w := performConfigRolloutReceiptIngestionPost(rollout.ID, payload)

	if w.Code != http.StatusConflict {
		t.Fatalf("evidence-chain attestation receipt should remain blocked, got %d body=%s", w.Code, w.Body.String())
	}
	assertConfigRolloutReceiptIngestionContains(t, w.Body.String(), []string{
		"PENDING",
		"cmdb.agent.plugin.dispatch.receipt.ingest.v1",
		`"receipt_type":"evidence_chain"`,
		"cmdb_agent_rollout_evidence_chain_contract",
		"findx_agent.config_rollout.evidence_chain.receipt.ingest",
	})
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())

	task, ok, err := store.GetFindXAgentExecutionTask(deliveryRef)
	if err != nil || !ok {
		t.Fatalf("delivery request_ref should still resolve after evidence-chain ingestion, ok=%v err=%v", ok, err)
	}
	if task.Status != "blocked" || task.Metadata["receipt_type"] != "delivery" || task.Metadata["phase"] != "delivery" {
		t.Fatalf("evidence-chain receipt must not rewrite canonical delivery task phase: %#v", task)
	}
	if task.Metadata["evidence_chain_receipt_status"] != "PENDING" ||
		task.Metadata["evidence_chain_receipt_contract_id"] != "cmdb_agent_rollout_evidence_chain_contract" ||
		task.Metadata["evidence_chain_receipt_ingestion_contract"] != cmdbAgentRolloutReceiptIngestContract ||
		strings.TrimSpace(task.Metadata["evidence_chain_receipt_audit_ref"]) == "" {
		t.Fatalf("blocked evidence-chain metadata not persisted safely: %#v", task.Metadata)
	}
	if !containsLifecycleTestString(task.EvidenceRefs, "evidence-chain-safe-ref") {
		t.Fatalf("safe evidence ref should be linked to task evidence refs: %#v", task.EvidenceRefs)
	}
	rawTask, _ := json.Marshal(task)
	assertNoConfigRolloutReceiptIngestionForbidden(t, string(rawTask))

	detail := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+rollout.ID, ListFindXAgentConfigRollouts)
	if detail.Code != http.StatusOK {
		t.Fatalf("evidence-chain receipt ingestion must not break request_ref read gate, got %d body=%s", detail.Code, detail.Body.String())
	}
	evidence := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival/evidence?rollout_id="+rollout.ID+"&request_ref="+deliveryRef, ListFindXAgentDataArrivalEvidence)
	readPayload := decodeDataArrivalRuntimeReadOK(t, evidence)
	if evidence.Code != http.StatusOK || readPayload.Status != "blocked" || readPayload.EvidenceCount != 1 {
		t.Fatalf("evidence-chain receipt must not turn evidence read into success, code=%d payload=%#v body=%s", evidence.Code, readPayload, evidence.Body.String())
	}
	if !containsLifecycleTestString(readPayload.MissingContracts, "cmdb_agent_rollout_evidence_chain_contract") {
		t.Fatalf("data-arrival read should still keep evidence-chain gap, payload=%#v", readPayload)
	}
	assertNoConfigRolloutReceiptIngestionForbidden(t, detail.Body.String())
	assertNoDataArrivalRuntimeReadForbidden(t, evidence.Body.String())
}

func TestFindXAgentConfigRolloutReceiptIngestionRejectsEvidenceChainWithoutDataArrivalReceipt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rollout := createConfigRolloutReceiptIngestionFixture(t)
	deliveryRef := strings.TrimSpace(rollout.Metadata["delivery_request_ref"])
	if deliveryRef == "" {
		t.Fatalf("fixture should create delivery_request_ref: %#v", rollout.Metadata)
	}
	saveConfigRolloutReceiptReceiverEvidence(t, rollout, deliveryRef)
	postBlockedConfigRolloutReceiptForTest(t, rollout, "delivery", deliveryRef)

	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"evidence_chain",
		"request_ref":"` + deliveryRef + `",
		"status":"PENDING"
	}`
	w := performConfigRolloutReceiptIngestionPost(rollout.ID, payload)

	if w.Code != http.StatusConflict {
		t.Fatalf("evidence-chain receipt without data-arrival receipt should be blocked, got %d body=%s", w.Code, w.Body.String())
	}
	assertConfigRolloutReceiptIngestionContains(t, w.Body.String(), []string{
		"PENDING",
		"cmdb_agent_rollout_data_arrival_receipt_contract",
		"cmdb_agent_rollout_evidence_chain_contract",
	})
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())
	task, ok, err := store.GetFindXAgentExecutionTask(deliveryRef)
	if err != nil || !ok {
		t.Fatalf("delivery request_ref should resolve after rejected evidence-chain receipt, ok=%v err=%v", ok, err)
	}
	if strings.TrimSpace(task.Metadata["evidence_chain_receipt_status"]) != "" {
		t.Fatalf("rejected evidence-chain receipt must not persist attestation metadata: %#v", task.Metadata)
	}
}

func saveConfigRolloutReceiptReceiverEvidence(t *testing.T, rollout model.FindXAgentConfigRollout, requestRef string) {
	t.Helper()
	if _, err := store.SaveFindXAgentDataArrivalEvidence(model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindMetrics,
		AgentID:      firstNonEmpty(rollout.AgentIDs...),
		TargetID:     firstNonEmpty(rollout.TargetIDs...),
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/metrics/write-compatible"},
		Metadata: map[string]string{
			"source_rollout_id": rollout.ID,
			"request_ref":       requestRef,
			"plugin_id":         rollout.PluginID,
			"password":          "evidence-chain-secret-marker",
		},
	}); err != nil {
		t.Fatalf("save receiver evidence: %v", err)
	}
}

func postBlockedConfigRolloutReceiptForTest(t *testing.T, rollout model.FindXAgentConfigRollout, receiptType, requestRef string) {
	t.Helper()
	payload := `{
		"rollout_id":"` + rollout.ID + `",
		"receipt_type":"` + receiptType + `",
		"request_ref":"` + requestRef + `",
		"status":"PENDING",
		"contract_id":"` + configRolloutReceiptIngestionContractID(receiptType, "") + `",
		"missing_contracts":["` + configRolloutReceiptIngestionContractID(receiptType, "") + `"],
		"evidence_ref":"` + receiptType + `-safe-ref"
	}`
	w := performConfigRolloutReceiptIngestionPost(rollout.ID, payload)
	if w.Code != http.StatusConflict {
		t.Fatalf("%s receipt ingestion should remain blocked, got %d body=%s", receiptType, w.Code, w.Body.String())
	}
	assertNoConfigRolloutReceiptIngestionForbidden(t, w.Body.String())
}
