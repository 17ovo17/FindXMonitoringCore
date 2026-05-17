package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const cmdbAgentRolloutReceiptIngestContract = "cmdb.agent.plugin.dispatch.receipt.ingest.v1"
const cmdbAgentRolloutReceiptAuditPersistenceContract = "cmdb_agent_rollout_receipt_audit_persistence_contract"

type configRolloutReceiptIngestionPayload struct {
	RolloutID        string   `json:"rollout_id"`
	ReceiptType      string   `json:"receipt_type"`
	RequestRef       string   `json:"request_ref"`
	Status           string   `json:"status"`
	ContractID       string   `json:"contract_id"`
	MissingContracts []string `json:"missing_contracts"`
	AuditRef         string   `json:"audit_ref"`
	EvidenceRef      string   `json:"evidence_ref"`
}

// IngestFindXAgentConfigRolloutReceipt records blocked dispatch receipt evidence
// without claiming the remote executor or data-arrival finalization completed.
func IngestFindXAgentConfigRolloutReceipt(c *gin.Context) {
	ingestFindXAgentConfigRolloutReceipt(c, store.AddMonitorAuditLog)
}

func ingestFindXAgentConfigRolloutReceipt(c *gin.Context, addAuditLog func(model.MonitorAuditLog) (model.MonitorAuditLog, error)) {
	rolloutID := sanitizeRemoteMutationValue("rollout_id", c.Param("id"))
	if rolloutID == "" {
		c.JSON(http.StatusConflict, configRolloutReceiptIngestionBlockedEnvelope("", "", "", "config rollout receipt ingestion requires rollout id", []string{
			"cmdb_agent_rollout_receipt_ingest_path_contract",
		}))
		return
	}
	rollout, ok, err := store.GetFindXAgentConfigRollout(rolloutID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config rollout receipt ingestion unavailable"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "config rollout not found"})
		return
	}
	if !isCMDBHostPluginDispatchRolloutRecord(rollout) {
		c.JSON(http.StatusConflict, configRolloutReceiptIngestionBlockedEnvelope(rollout.ID, "", "", "config rollout receipt ingestion only supports cmdb host plugin dispatch rollouts", []string{
			"cmdb_agent_rollout_dispatch_contract",
			cmdbAgentRolloutTaskMatchContract,
		}))
		return
	}
	var payload configRolloutReceiptIngestionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusConflict, configRolloutReceiptIngestionBlockedEnvelope(rollout.ID, "", "", "config rollout receipt ingestion requires a JSON payload", []string{
			"cmdb_agent_rollout_receipt_ingest_payload_contract",
		}))
		return
	}
	receiptType := strings.TrimSpace(payload.ReceiptType)
	if !configRolloutReceiptIngestionTypeSupported(receiptType) {
		c.JSON(http.StatusConflict, configRolloutReceiptIngestionBlockedEnvelope(rollout.ID, "", receiptType, "unsupported config rollout receipt type", []string{
			"cmdb_agent_rollout_receipt_type_contract",
		}))
		return
	}
	if requestRolloutID := sanitizeRemoteMutationValue("rollout_id", payload.RolloutID); requestRolloutID != "" && requestRolloutID != rollout.ID {
		c.JSON(http.StatusConflict, configRolloutReceiptIngestionBlockedEnvelope(rollout.ID, "", receiptType, "config rollout receipt payload rollout_id does not match path", []string{
			cmdbAgentRolloutTaskMatchContract,
		}))
		return
	}
	if strings.TrimSpace(payload.Status) != "pending" {
		c.JSON(http.StatusConflict, configRolloutReceiptIngestionBlockedEnvelope(rollout.ID, sanitizeRemoteMutationValue("request_ref", payload.RequestRef), receiptType, "config rollout receipt ingestion only accepts blocked contract receipts until delivery/effect/rollback executors are enabled", []string{
			"cmdb_agent_rollout_receipt_status_contract",
		}))
		return
	}
	requestRef := sanitizeRemoteMutationValue("request_ref", payload.RequestRef)
	receiptPhase, phaseOK := configRolloutReceiptIngestionRequestPhase(rollout, receiptType, requestRef)
	if requestRef == "" || !phaseOK {
		c.JSON(http.StatusConflict, configRolloutReceiptIngestionBlockedEnvelope(rollout.ID, requestRef, receiptType, "config rollout receipt request_ref does not match rollout receipt phase", []string{
			cmdbAgentRolloutRequestRefResolveContract,
			cmdbAgentRolloutExecutionTaskContract,
			cmdbAgentRolloutTaskMatchContract,
		}))
		return
	}
	task, ok, err := store.GetFindXAgentExecutionTask(requestRef)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config rollout receipt task unavailable"})
		return
	}
	if !ok || !cmdbConfigRolloutReceiptTaskMatches(rollout, task, receiptPhase) {
		c.JSON(http.StatusConflict, configRolloutReceiptIngestionBlockedEnvelope(rollout.ID, requestRef, receiptType, "config rollout receipt request_ref must resolve to a matching blocked receipt task", []string{
			cmdbAgentRolloutRequestRefResolveContract,
			cmdbAgentRolloutExecutionTaskContract,
			cmdbAgentRolloutTaskMatchContract,
		}))
		return
	}
	if (receiptType == "data_arrival" || receiptType == "evidence_chain") && !configRolloutReceiptIngestionHasReceiverEvidence(rollout, requestRef) {
		c.JSON(http.StatusConflict, configRolloutReceiptIngestionBlockedEnvelope(rollout.ID, requestRef, receiptType, "config rollout data-arrival receipt requires receiver-backed evidence before finalization", []string{
			cmdbAgentRolloutDataArrivalEvidenceContract,
			cmdbAgentRolloutReceiverEvidenceContract,
			cmdbAgentRolloutDataArrivalRequestRefContract,
		}))
		return
	}
	if receiptType == "evidence_chain" && !configRolloutReceiptIngestionHasDataArrivalReceipt(task) {
		c.JSON(http.StatusConflict, configRolloutReceiptIngestionBlockedEnvelope(rollout.ID, requestRef, receiptType, "config rollout evidence-chain attestation requires blocked data-arrival receipt finalization evidence", []string{
			"cmdb_agent_rollout_data_arrival_receipt_contract",
			"cmdb_agent_rollout_evidence_chain_contract",
		}))
		return
	}

	auditRef := "cmdb-agent-rollout-" + receiptType + "-receipt-ingest-" + store.NewID()
	contractID := configRolloutReceiptIngestionContractID(receiptType, payload.ContractID)
	missing := configRolloutReceiptIngestionMissing(receiptType, payload.MissingContracts)
	candidate := task
	if candidate.Metadata == nil {
		candidate.Metadata = map[string]string{}
	}
	candidate.Status = "blocked"
	candidate.Blocker = "PENDING: cmdb plugin dispatch " + receiptType + " receipt executor is not enabled"
	candidate.Audit = configRolloutReceiptIngestionAuditAction(receiptType)
	configRolloutReceiptIngestionApplyMetadata(candidate.Metadata, receiptType, receiptPhase, contractID, auditRef, missing)
	if evidenceRef := sanitizeRemoteMutationValue("evidence_ref", payload.EvidenceRef); evidenceRef != "" {
		candidate.EvidenceRefs = appendConfigRolloutReceiptIngestionRef(candidate.EvidenceRefs, evidenceRef)
	}
	saved, err := store.SaveFindXAgentExecutionTask(candidate)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config rollout receipt task persistence unavailable"})
		return
	}
	auditMissing := []string{}
	if _, err := addAuditLog(model.MonitorAuditLog{
		ID:           auditRef,
		Actor:        requestActor(c),
		Action:       configRolloutReceiptIngestionAuditAction(receiptType),
		ResourceType: "findx_agent_config_rollout",
		ResourceID:   rollout.ID,
		Scope:        "cmdb",
		Status:       "accepted",
		ClientIP:     c.ClientIP(),
		Summary:      "CMDB plugin dispatch receipt ingestion accepted",
		Details: map[string]any{
			"rollout_id":        rollout.ID,
			"task_ref":          saved.ID,
			"receipt_type":      receiptType,
			"contract_id":       contractID,
			"missing_contracts": missing,
			"request_ref":       requestRef,
			"evidence_ref":      sanitizeRemoteMutationValue("evidence_ref", payload.EvidenceRef),
		},
	}); err != nil {
		auditMissing = append(auditMissing, cmdbAgentRolloutReceiptAuditPersistenceContract)
	}
	c.JSON(http.StatusOK, gin.H{
		"code":        http.StatusOK,
		"status":      "accepted",
		"contract":    cmdbAgentRolloutReceiptIngestContract,
		"rollout_ref": rollout.ID,
		"request_ref": task.ID,
		"receipt": gin.H{
			"task_ref":     saved.ID,
			"receipt_type": receiptType,
			"request_ref":  task.ID,
			"status":       "accepted",
			"contract_id":  contractID,
			"audit_ref":    auditRef,
		},
	})
}

func configRolloutReceiptIngestionTypeSupported(receiptType string) bool {
	switch strings.TrimSpace(receiptType) {
	case "delivery", "effect", "rollback", "data_arrival", "evidence_chain":
		return true
	default:
		return false
	}
}

func configRolloutReceiptIngestionRequestPhase(rollout model.FindXAgentConfigRollout, receiptType, requestRef string) (string, bool) {
	requestRef = strings.TrimSpace(requestRef)
	if requestRef == "" {
		return "", false
	}
	if receiptType != "data_arrival" && receiptType != "evidence_chain" {
		return receiptType, strings.TrimSpace(rollout.Metadata[receiptType+"_request_ref"]) == requestRef
	}
	for _, phase := range []string{"delivery", "effect", "rollback"} {
		if strings.TrimSpace(rollout.Metadata[phase+"_request_ref"]) == requestRef {
			return phase, true
		}
	}
	return "", false
}

func configRolloutReceiptIngestionContractID(receiptType, requested string) string {
	requested = strings.TrimSpace(requested)
	if requested != "" {
		switch strings.TrimSpace(receiptType) {
		case "delivery":
			if requested == "cmdb_agent_rollout_delivery_receipt_contract" {
				return requested
			}
		case "effect":
			if requested == "cmdb_agent_rollout_effect_receipt_contract" {
				return requested
			}
		case "rollback":
			if requested == "cmdb_agent_rollout_rollback_receipt_contract" {
				return requested
			}
		case "data_arrival":
			if requested == "cmdb_agent_rollout_data_arrival_receipt_contract" {
				return requested
			}
		case "evidence_chain":
			if requested == "cmdb_agent_rollout_evidence_chain_contract" {
				return requested
			}
		}
	}
	switch strings.TrimSpace(receiptType) {
	case "delivery":
		return "cmdb_agent_rollout_delivery_receipt_contract"
	case "effect":
		return "cmdb_agent_rollout_effect_receipt_contract"
	case "rollback":
		return "cmdb_agent_rollout_rollback_receipt_contract"
	case "data_arrival":
		return "cmdb_agent_rollout_data_arrival_receipt_contract"
	case "evidence_chain":
		return "cmdb_agent_rollout_evidence_chain_contract"
	default:
		return "cmdb_agent_rollout_receipt_contract"
	}
}

func configRolloutReceiptIngestionMissing(receiptType string, requested []string) []string {
	allowed := map[string]bool{
		"cmdb_agent_rollout_remote_executor_contract":            true,
		configRolloutReceiptIngestionContractID(receiptType, ""): true,
	}
	switch strings.TrimSpace(receiptType) {
	case "delivery":
		allowed["cmdb_agent_rollout_delivery_executor_contract"] = true
		allowed["cmdb_agent_rollout_delivery_executor_registration_contract"] = true
		allowed["cmdb_agent_rollout_delivery_runner_identity_contract"] = true
		allowed["cmdb_agent_rollout_delivery_attested_receipt_contract"] = true
		allowed["cmdb_agent_rollout_delivery_target_binding_contract"] = true
		allowed["cmdb_agent_rollout_delivery_request_ref_match_contract"] = true
	case "effect":
		allowed["cmdb_agent_rollout_effect_executor_contract"] = true
		allowed["cmdb_agent_rollout_effect_executor_registration_contract"] = true
		allowed["cmdb_agent_rollout_effect_runner_identity_contract"] = true
		allowed["cmdb_agent_rollout_effect_delivery_evidence_match_contract"] = true
		allowed["cmdb_agent_rollout_effect_attested_receipt_contract"] = true
		allowed["cmdb_agent_rollout_effect_request_ref_match_contract"] = true
	case "rollback":
		allowed["cmdb_agent_rollout_rollback_executor_contract"] = true
		allowed["cmdb_agent_rollout_rollback_executor_registration_contract"] = true
		allowed["cmdb_agent_rollout_rollback_runner_identity_contract"] = true
		allowed["cmdb_agent_rollout_rollback_operation_context_contract"] = true
		allowed["cmdb_agent_rollout_rollback_attested_receipt_contract"] = true
		allowed["cmdb_agent_rollout_rollback_request_ref_match_contract"] = true
	case "data_arrival":
		allowed["cmdb_agent_rollout_data_arrival_contract"] = true
		allowed["cmdb_agent_rollout_data_arrival_evidence_contract"] = true
		allowed["cmdb_agent_rollout_receiver_evidence_contract"] = true
		allowed["cmdb_agent_rollout_evidence_chain_contract"] = true
	case "evidence_chain":
		allowed["cmdb_agent_rollout_data_arrival_receipt_contract"] = true
		allowed["cmdb_agent_rollout_evidence_chain_contract"] = true
		allowed["cmdb_agent_rollout_evidence_chain_attestation_contract"] = true
	}
	out := make([]string, 0, len(requested)+len(allowed))
	seen := map[string]bool{}
	for _, item := range requested {
		clean := sanitizeRemoteMutationValue("missing_contract", item)
		if clean == "" || !allowed[clean] || seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	for required := range allowed {
		if !seen[required] {
			out = append(out, required)
		}
	}
	return uniquePackageRepositoryBlockers(out)
}

func configRolloutReceiptIngestionApplyMetadata(metadata map[string]string, receiptType, receiptPhase, contractID, auditRef string, missing []string) {
	if receiptType == "data_arrival" || receiptType == "evidence_chain" {
		prefix := "data_arrival"
		if receiptType == "evidence_chain" {
			prefix = "evidence_chain"
		}
		metadata[prefix+"_receipt_status"] = "pending"
		metadata[prefix+"_receipt_type"] = receiptType
		metadata[prefix+"_receipt_phase_ref"] = receiptPhase
		metadata[prefix+"_receipt_contract_id"] = contractID
		metadata[prefix+"_receipt_audit_ref"] = auditRef
		metadata[prefix+"_receipt_ingestion_contract"] = cmdbAgentRolloutReceiptIngestContract
		metadata[prefix+"_receipt_missing_contracts"] = configRolloutReceiptIngestionMissingJSON(missing)
		return
	}
	metadata["receipt_status"] = "pending"
	metadata["receipt_type"] = receiptType
	metadata["receipt_contract_id"] = contractID
	metadata["receipt_audit_ref"] = auditRef
	metadata["receipt_ingestion_contract"] = cmdbAgentRolloutReceiptIngestContract
	metadata["receipt_missing_contracts"] = configRolloutReceiptIngestionMissingJSON(missing)
}

func configRolloutReceiptIngestionHasReceiverEvidence(rollout model.FindXAgentConfigRollout, requestRef string) bool {
	items, err := store.ListFindXAgentDataArrivalEvidence()
	if err != nil {
		return false
	}
	gate := configRolloutRuntimeReadGateForItem(rollout)
	query := dataArrivalRuntimeReadQuery{RolloutID: rollout.ID, RequestRef: requestRef, PluginID: rollout.PluginID}
	return len(dataArrivalRuntimeEvidenceForRollout(items, rollout, query, gate)) > 0
}

func configRolloutReceiptIngestionHasDataArrivalReceipt(task model.FindXAgentExecutionTask) bool {
	return task.Metadata["data_arrival_receipt_status"] == "pending" &&
		task.Metadata["data_arrival_receipt_contract_id"] == "cmdb_agent_rollout_data_arrival_receipt_contract"
}

func configRolloutReceiptIngestionMissingJSON(items []string) string {
	raw, err := json.Marshal(uniquePackageRepositoryBlockers(items))
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func appendConfigRolloutReceiptIngestionRef(values []string, value string) []string {
	clean := sanitizeRemoteMutationValue("evidence_ref", value)
	if clean == "" {
		return values
	}
	for _, item := range values {
		if strings.TrimSpace(item) == clean {
			return values
		}
	}
	return append(values, clean)
}

func configRolloutReceiptIngestionAuditAction(receiptType string) string {
	return "findx_agent.config_rollout." + strings.TrimSpace(receiptType) + ".receipt.ingest"
}

func configRolloutReceiptIngestionEnvelope(rollout model.FindXAgentConfigRollout, task model.FindXAgentExecutionTask, receiptType, contractID string, missing []string, auditRef string) gin.H {
	return gin.H{
		"code":        http.StatusConflict,
		"status":      "pending",
		"error":       "pending",
		"message":     "CMDB plugin dispatch receipt ingestion recorded a blocked contract receipt; delivery, effect and rollback executors are still not enabled.",
		"contract":    cmdbAgentRolloutReceiptIngestContract,
		"rollout_ref": rollout.ID,
		"request_ref": task.ID,
		"receipt": gin.H{
			"task_ref":          task.ID,
			"receipt_type":      receiptType,
			"request_ref":       task.ID,
			"status":            "pending",
			"contract_id":       contractID,
			"missing_contracts": uniquePackageRepositoryBlockers(missing),
			"audit_ref":         auditRef,
		},
		"missing_contracts": uniquePackageRepositoryBlockers(append([]string{
			"cmdb_agent_rollout_delivery_receipt_contract",
			"cmdb_agent_rollout_effect_receipt_contract",
			"cmdb_agent_rollout_rollback_receipt_contract",
			"cmdb_agent_rollout_remote_executor_contract",
		}, missing...)),
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "findx_agent_config_rollout",
			"resource_id":   rollout.ID,
			"action":        configRolloutReceiptIngestionAuditAction(receiptType),
			"request_ref":   task.ID,
		},
		"expected_schema": configRolloutReceiptIngestionExpectedSchema(),
		"safe_to_retry":   false,
	}
}

func configRolloutReceiptIngestionBlockedEnvelope(rolloutID, requestRef, receiptType, message string, missing []string) gin.H {
	return gin.H{
		"code":              http.StatusConflict,
		"status":            "pending",
		"error":             "pending",
		"message":           message,
		"contract":          cmdbAgentRolloutReceiptIngestContract,
		"missing_contracts": uniquePackageRepositoryBlockers(missing),
		"rollout_ref":       sanitizeRemoteMutationValue("rollout_id", rolloutID),
		"request_ref":       sanitizeRemoteMutationValue("request_ref", requestRef),
		"receipt_type":      strings.TrimSpace(receiptType),
		"expected_schema":   configRolloutReceiptIngestionExpectedSchema(),
		"safe_to_retry":     false,
	}
}

func configRolloutReceiptIngestionExpectedSchema() gin.H {
	return gin.H{
		"path": []string{"rollout_id", "receipts"},
		"request": []string{
			"rollout_id", "receipt_type", "request_ref", "status", "contract_id", "missing_contracts", "evidence_ref",
		},
		"receipt": []string{
			"task_ref", "receipt_type", "request_ref", "status", "contract_id", "missing_contracts", "audit_ref",
		},
		"log_query": []string{"source", "scope", "resource_type", "resource_id", "action", "request_ref"},
	}
}
