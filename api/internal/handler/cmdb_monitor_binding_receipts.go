package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

type cmdbMonitorBindingReceiptIngestionPayload struct {
	BindingID        string   `json:"binding_id"`
	ReceiptType      string   `json:"receipt_type"`
	Status           string   `json:"status"`
	ContractID       string   `json:"contract_id"`
	MissingContracts []string `json:"missing_contracts"`
	RequestRef       string   `json:"request_ref"`
	AuditRef         string   `json:"audit_ref"`
	EvidenceRef      string   `json:"evidence_ref"`
}

// GetCmdbMonitorBindingReceipts returns stored receipts for an instance's bindings.
func GetCmdbMonitorBindingReceipts(c *gin.Context) {
	instanceID := cmdbMonitorBindingInstanceID(c)
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instance_id is required"})
		return
	}
	if _, ok := store.GetCmdbInstance(instanceID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb instance not found"})
		return
	}
	bindings := store.ListCmdbMonitorBindings(instanceID)
	receipts := store.ListCmdbMonitorBindingReceiptsForInstance(instanceID)
	c.JSON(http.StatusOK, cmdbMonitorBindingReceiptsReadyEnvelope(instanceID, bindings, receipts))
}

func IngestCmdbMonitorBindingReceipt(c *gin.Context) {
	parts := cmdbMonitorBindingPathParts(c)
	if len(parts) < 3 {
		c.JSON(http.StatusConflict, cmdbMonitorBindingReceiptIngestionBlockedEnvelope("", "", "", "cmdb monitor binding receipt ingestion requires instance_id, binding_id and receipts path", []string{
			"cmdb_monitor_binding_receipt_ingest_path_contract",
		}))
		return
	}
	instanceID, bindingID := parts[0], parts[1]
	if _, ok := store.GetCmdbInstance(instanceID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb instance not found"})
		return
	}
	binding, ok := store.GetCmdbMonitorBinding(bindingID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb monitor binding not found"})
		return
	}
	if strings.TrimSpace(binding.InstanceID) != strings.TrimSpace(instanceID) {
		c.JSON(http.StatusConflict, cmdbMonitorBindingReceiptIngestionBlockedEnvelope(instanceID, bindingID, "", "cmdb monitor binding does not belong to the requested instance", []string{
			"cmdb_monitor_binding_instance_match_contract",
		}))
		return
	}
	var payload cmdbMonitorBindingReceiptIngestionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusConflict, cmdbMonitorBindingReceiptIngestionBlockedEnvelope(instanceID, bindingID, "", "cmdb monitor binding receipt ingestion requires a JSON payload", []string{
			"cmdb_monitor_binding_receipt_ingest_payload_contract",
		}))
		return
	}
	receiptType := strings.TrimSpace(payload.ReceiptType)
	if !cmdbMonitorBindingReceiptTypeSupported(receiptType) {
		c.JSON(http.StatusConflict, cmdbMonitorBindingReceiptIngestionBlockedEnvelope(instanceID, bindingID, receiptType, "unsupported cmdb monitor binding receipt type", []string{
			"cmdb_monitor_binding_receipt_type_contract",
		}))
		return
	}
	if strings.TrimSpace(payload.BindingID) != "" && strings.TrimSpace(payload.BindingID) != bindingID {
		c.JSON(http.StatusConflict, cmdbMonitorBindingReceiptIngestionBlockedEnvelope(instanceID, bindingID, receiptType, "cmdb monitor binding receipt payload binding_id does not match path", []string{
			"cmdb_monitor_binding_instance_match_contract",
		}))
		return
	}
	if strings.TrimSpace(payload.Status) != "pending" {
		c.JSON(http.StatusConflict, cmdbMonitorBindingReceiptIngestionBlockedEnvelope(instanceID, bindingID, receiptType, "cmdb monitor binding receipt ingestion only accepts blocked contract receipts until delivery/effect/rollback executors are enabled", []string{
			"cmdb_monitor_binding_receipt_status_contract",
		}))
		return
	}
	existing, ok := store.GetCmdbMonitorBindingReceipt(bindingID, receiptType)
	if !ok {
		c.JSON(http.StatusConflict, cmdbMonitorBindingReceiptIngestionBlockedEnvelope(instanceID, bindingID, receiptType, "cmdb monitor binding receipt ingestion updates existing receipt rows only", []string{
			"cmdb_monitor_binding_receipt_store",
			"cmdb_monitor_binding_receipt_complete_contract",
		}))
		return
	}
	if strings.TrimSpace(existing.RequestRef) != "" && strings.TrimSpace(payload.RequestRef) != strings.TrimSpace(existing.RequestRef) {
		c.JSON(http.StatusConflict, cmdbMonitorBindingReceiptIngestionBlockedEnvelope(instanceID, bindingID, receiptType, "cmdb monitor binding receipt request_ref does not match stored receipt", []string{
			"cmdb_monitor_binding_request_ref_resolve_contract",
			"cmdb_monitor_binding_execution_task_contract",
		}))
		return
	}
	candidate := *existing
	candidate.RequestRef = strings.TrimSpace(payload.RequestRef)
	candidate.Status = "pending"
	candidate.ContractID = cmdbMonitorBindingReceiptContractID(receiptType, payload.ContractID)
	candidate.MissingJSON = cmdbMonitorBindingReceiptMissingJSON(cmdbMonitorBindingReceiptIngestionMissing(receiptType, payload.MissingContracts))
	if !cmdbMonitorBindingExecutionRequestResolved(*binding, candidate) {
		c.JSON(http.StatusConflict, cmdbMonitorBindingReceiptIngestionBlockedEnvelope(instanceID, bindingID, receiptType, "cmdb monitor binding receipt request_ref must resolve to a matching blocked execution task", []string{
			"cmdb_monitor_binding_request_ref_resolve_contract",
			"cmdb_monitor_binding_execution_task_contract",
			cmdbMonitorBindingReceiptContractID(receiptType, ""),
		}))
		return
	}
	auditRef := "cmdb-monitor-binding-" + receiptType + "-ingest-" + store.NewID()
	candidate.AuditRef = auditRef
	saved, err := store.UpdateCmdbMonitorBindingReceipt(&candidate)
	if err != nil || saved == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb monitor binding receipt store unavailable"})
		return
	}
	if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
		ID:           auditRef,
		Actor:        requestActor(c),
		Action:       cmdbMonitorBindingReceiptIngestionAuditAction(receiptType),
		ResourceType: "cmdb_monitor_binding",
		ResourceID:   binding.InstanceID,
		Scope:        "cmdb",
		Status:       "blocked",
		ClientIP:     c.ClientIP(),
		Summary:      "CMDB monitor binding receipt ingestion blocked by contract",
		Details: map[string]any{
			"binding_id":        binding.ID,
			"receipt_id":        saved.ID,
			"receipt_type":      receiptType,
			"contract_id":       saved.ContractID,
			"missing_contracts": cmdbMonitorBindingReceiptMissingList(saved.MissingJSON),
			"request_ref":       saved.RequestRef,
			"evidence_ref":      cmdbSanitizeBindingValue("evidence_ref", payload.EvidenceRef),
		},
	}); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb monitor binding receipt audit unavailable"})
		return
	}
	c.JSON(http.StatusConflict, cmdbMonitorBindingReceiptIngestionEnvelope(binding.InstanceID, binding.ID, *saved))
}

func cmdbMonitorBindingReceiptsPath(c *gin.Context) bool {
	path := strings.Trim(c.Param("path"), "/")
	parts := strings.Split(path, "/")
	return len(parts) >= 2 && strings.TrimSpace(parts[1]) == "receipts"
}

func cmdbMonitorBindingReceiptIngestionPath(c *gin.Context) bool {
	parts := cmdbMonitorBindingPathParts(c)
	return len(parts) == 3 && strings.TrimSpace(parts[2]) == "receipts"
}

func cmdbMonitorBindingReceiptTypeSupported(receiptType string) bool {
	switch strings.TrimSpace(receiptType) {
	case "delivery", "effect", "rollback":
		return true
	default:
		return false
	}
}

func cmdbMonitorBindingReceiptContractID(receiptType, requested string) string {
	if strings.TrimSpace(requested) != "" {
		switch strings.TrimSpace(receiptType) {
		case "delivery":
			if strings.TrimSpace(requested) == "cmdb_monitor_binding_delivery_receipt_contract" {
				return strings.TrimSpace(requested)
			}
		case "effect":
			if strings.TrimSpace(requested) == "cmdb_monitor_binding_effect_receipt_contract" {
				return strings.TrimSpace(requested)
			}
		case "rollback":
			if strings.TrimSpace(requested) == "binding_rollback_contract" || strings.TrimSpace(requested) == "cmdb_monitor_binding_rollback_receipt_contract" {
				return strings.TrimSpace(requested)
			}
		}
	}
	switch strings.TrimSpace(receiptType) {
	case "delivery":
		return "cmdb_monitor_binding_delivery_receipt_contract"
	case "effect":
		return "cmdb_monitor_binding_effect_receipt_contract"
	case "rollback":
		return "binding_rollback_contract"
	default:
		return "cmdb_monitor_binding_receipt_complete_contract"
	}
}

func cmdbMonitorBindingReceiptIngestionMissing(receiptType string, requested []string) []string {
	allowed := map[string]bool{
		cmdbMonitorBindingReceiptContractID(receiptType, ""): true,
	}
	switch strings.TrimSpace(receiptType) {
	case "delivery":
		allowed["cmdb_monitor_binding_delivery_executor"] = true
	case "effect":
		allowed["cmdb_monitor_binding_effect_probe"] = true
	case "rollback":
		allowed["cmdb_monitor_binding_rollback_executor"] = true
		allowed["cmdb_monitor_binding_rollback_receipt_contract"] = true
	}
	out := make([]string, 0, len(requested)+1)
	seen := map[string]bool{}
	for _, item := range requested {
		clean := strings.TrimSpace(item)
		if clean == "" || !allowed[clean] || seen[clean] || cmdbSensitiveBindingText(clean) || cmdbSensitiveBindingKey(clean) {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	for required := range allowed {
		if !seen[required] && !strings.Contains(required, "rollback_receipt") {
			out = append(out, required)
		}
	}
	return out
}

func cmdbMonitorBindingReceiptIngestionAuditAction(receiptType string) string {
	return "cmdb.monitor_binding." + strings.TrimSpace(receiptType) + ".receipt.ingest"
}

func cmdbMonitorBindingReceiptIngestionEnvelope(instanceID, bindingID string, receipt model.CmdbMonitorBindingReceipt) gin.H {
	return gin.H{
		"code":        http.StatusConflict,
		"status":      "pending",
		"error":       "pending",
		"message":     "CMDB monitor binding receipt ingestion recorded a blocked contract receipt; delivery, effect and rollback executors are still not enabled.",
		"contract":    "cmdb.monitor_binding.receipt.ingest.v1",
		"instance_id": strings.TrimSpace(instanceID),
		"binding_id":  strings.TrimSpace(bindingID),
		"receipt":     cmdbMonitorBindingReceiptDTOs([]model.CmdbMonitorBindingReceipt{receipt})[0],
		"missing_contracts": []string{
			"cmdb_monitor_binding_delivery_receipt_contract",
			"cmdb_monitor_binding_effect_receipt_contract",
			"binding_rollback_contract",
			"cmdb_monitor_binding_execution_task_contract",
		},
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_monitor_binding",
			"resource_id":   strings.TrimSpace(instanceID),
			"action":        cmdbMonitorBindingReceiptIngestionAuditAction(receipt.ReceiptType),
		},
		"expected_schema": cmdbMonitorBindingReceiptIngestionExpectedSchema(),
		"field_matrix":    cmdbMonitorBindingReceiptIngestionFieldMatrix("blocked"),
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingReceiptIngestionBlockedEnvelope(instanceID, bindingID, receiptType, message string, missing []string) gin.H {
	return gin.H{
		"code":              http.StatusConflict,
		"status":            "pending",
		"error":             "pending",
		"message":           message,
		"contract":          "cmdb.monitor_binding.receipt.ingest.v1",
		"missing_contracts": missing,
		"instance_id":       strings.TrimSpace(instanceID),
		"binding_id":        strings.TrimSpace(bindingID),
		"receipt_type":      strings.TrimSpace(receiptType),
		"expected_schema":   cmdbMonitorBindingReceiptIngestionExpectedSchema(),
		"field_matrix":      cmdbMonitorBindingReceiptIngestionFieldMatrix("blocked"),
		"safe_to_retry":     false,
		"meta":              cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingReceiptIngestionExpectedSchema() gin.H {
	return gin.H{
		"path": []string{"instance_id", "binding_id", "receipts"},
		"request": []string{
			"binding_id", "receipt_type", "request_ref", "status", "missing_contracts", "evidence_ref",
		},
		"receipt": []string{
			"id", "binding_id", "instance_id", "receipt_type", "status", "contract_id", "missing_contracts", "request_ref", "audit_ref",
		},
		"log_query": []string{"source", "scope", "resource_type", "resource_id", "action"},
	}
}

func cmdbMonitorBindingReceiptIngestionFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("binding_receipt_ingest_path", status, []string{
			"instance_id", "binding_id", "receipts",
		}, "cmdb_monitor_binding_receipt_ingest_path_contract"),
		cmdbContractFieldGroup("binding_receipt_ref", status, []string{
			"receipt_type", "request_ref", "binding_id",
		}, "cmdb_monitor_binding_request_ref_resolve_contract"),
		cmdbContractFieldGroup("binding_execution_task", status, []string{
			"findx_agent_execution_tasks.id", "action", "status", "target_ids", "metadata.cmdb_binding_id",
		}, "cmdb_monitor_binding_execution_task_contract"),
		cmdbContractFieldGroup("binding_receipt_status", "blocked", []string{
			"pending", "cmdb_monitor_binding_delivery_executor", "cmdb_monitor_binding_effect_probe", "cmdb_monitor_binding_rollback_executor",
		}, "cmdb_monitor_binding_receipt_status_contract"),
	}
}

func cmdbMonitorBindingReceiptsReadyEnvelope(instanceID string, bindings []model.CmdbMonitorBinding, receipts []model.CmdbMonitorBindingReceipt) gin.H {
	return gin.H{
		"code":              0,
		"status":            "ready",
		"contract":          "cmdb.monitor_binding.receipts.read.v1",
		"instance_id":       instanceID,
		"binding_ids":       cmdbMonitorBindingIDs(bindings),
		"receipts":          cmdbMonitorBindingReceiptDTOs(receipts),
		"total":             len(receipts),
		"expected_schema":   cmdbMonitorBindingReceiptsReadContract()["expected_schema"],
		"field_matrix":      cmdbMonitorBindingReceiptsReadContract()["field_matrix"],
		"source_evidence":   cmdbMonitorBindingReceiptsReadContract()["source_evidence"],
		"missing_contracts": []string{"cmdb_monitor_binding_delivery_executor", "cmdb_monitor_binding_effect_probe", "cmdb_monitor_binding_rollback_executor"},
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_monitor_binding",
			"resource_id":   instanceID,
			"actions": []string{
				"cmdb.monitor_binding.delivery.blocked",
				"cmdb.monitor_binding.effect.blocked",
				"cmdb.monitor_binding.rollback.blocked",
			},
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingReceiptQueryBlockedEnvelope(instanceID string) gin.H {
	contract := cmdbMonitorBindingReceiptsReadContract()
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  "CMDB monitor binding receipts require a stored binding and stored receipt rows; empty receipts are not delivery evidence.",
		"contract": "cmdb.monitor_binding.receipts.read.v1",
		"missing_contracts": []string{
			"cmdb_monitor_binding_store",
			"cmdb_monitor_binding_receipt_store",
			"cmdb_monitor_binding_receipt_complete_contract",
			"cmdb_monitor_binding_request_ref_contract",
			"cmdb_monitor_binding_delivery_receipt_contract",
			"cmdb_monitor_binding_effect_receipt_contract",
			"binding_rollback_contract",
		},
		"instance_id":     strings.TrimSpace(instanceID),
		"expected_schema": contract["expected_schema"],
		"field_matrix":    contract["blocked_field_matrix"],
		"source_evidence": contract["source_evidence"],
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingReceiptsReadContract() gin.H {
	return gin.H{
		"expected_schema": gin.H{
			"receipts[]": []string{
				"id", "binding_id", "instance_id", "receipt_type", "status", "audit_action",
				"contract_id", "missing_contracts", "request_ref", "audit_ref", "created_at",
			},
			"log_query": []string{"source", "scope", "resource_type", "resource_id", "actions"},
		},
		"field_matrix": []gin.H{
			cmdbContractFieldGroup("binding_receipt_store", "ready", []string{
				"binding_id", "instance_id", "receipt_type", "status", "contract_id", "missing_contracts", "request_ref",
			}, "cmdb_monitor_binding_receipt_store"),
			cmdbContractFieldGroup("binding_delivery_receipt", "blocked", []string{
				"delivery_status", "request_ref", "cmdb_monitor_binding_delivery_executor",
			}, "cmdb_monitor_binding_delivery_receipt_contract"),
			cmdbContractFieldGroup("binding_effect_receipt", "blocked", []string{
				"effect_status", "request_ref", "cmdb_monitor_binding_effect_probe",
			}, "cmdb_monitor_binding_effect_receipt_contract"),
			cmdbContractFieldGroup("binding_rollback_receipt", "blocked", []string{
				"rollback_status", "request_ref", "cmdb_monitor_binding_rollback_executor",
			}, "binding_rollback_contract"),
			cmdbContractFieldGroup("binding_receipt_audit", "ready", []string{
				"findx_audit", "cmdb.monitor_binding.delivery.blocked", "cmdb.monitor_binding.effect.blocked", "cmdb.monitor_binding.rollback.blocked",
			}, "binding_audit_contract"),
		},
		"blocked_field_matrix": []gin.H{
			cmdbContractFieldGroup("binding_receipt_store", "missing_backend", []string{
				"binding_id", "instance_id", "receipt_type", "status", "contract_id", "missing_contracts", "request_ref",
			}, "cmdb_monitor_binding_receipt_store"),
			cmdbContractFieldGroup("binding_delivery_receipt", "blocked", []string{
				"delivery_status", "request_ref", "cmdb_monitor_binding_delivery_executor",
			}, "cmdb_monitor_binding_delivery_receipt_contract"),
			cmdbContractFieldGroup("binding_effect_receipt", "blocked", []string{
				"effect_status", "request_ref", "cmdb_monitor_binding_effect_probe",
			}, "cmdb_monitor_binding_effect_receipt_contract"),
			cmdbContractFieldGroup("binding_rollback_receipt", "blocked", []string{
				"rollback_status", "request_ref", "cmdb_monitor_binding_rollback_executor",
			}, "binding_rollback_contract"),
		},
		"source_evidence": cmdbMonitorBindingsWriteBlockedContract()["source_evidence"],
	}
}

func cmdbMonitorBindingIDs(bindings []model.CmdbMonitorBinding) []string {
	out := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if strings.TrimSpace(binding.ID) != "" {
			out = append(out, binding.ID)
		}
	}
	return out
}

func cmdbMonitorBindingReceiptsComplete(receipts []model.CmdbMonitorBindingReceipt) bool {
	seen := map[string]bool{}
	for _, receipt := range receipts {
		receiptType := strings.TrimSpace(receipt.ReceiptType)
		if receiptType != "" && strings.TrimSpace(receipt.RequestRef) != "" {
			seen[receiptType] = true
		}
	}
	for _, required := range []string{"delivery", "effect", "rollback"} {
		if !seen[required] {
			return false
		}
	}
	return true
}

func cmdbCreateMonitorBindingReceipts(binding model.CmdbMonitorBinding, auditRef, actor, clientIP, deliveryRequestRef string) ([]model.CmdbMonitorBindingReceipt, error) {
	effectRequestRef, err := cmdbCreateMonitorBindingExecutionRequest(binding, "effect", actor)
	if err != nil {
		return nil, err
	}
	rollbackRequestRef, err := cmdbCreateMonitorBindingExecutionRequest(binding, "rollback", actor)
	if err != nil {
		return nil, err
	}
	defs := []struct {
		typ        string
		contract   string
		missing    []string
		requestRef string
	}{
		{"delivery", "cmdb_monitor_binding_delivery_receipt_contract", []string{"cmdb_monitor_binding_delivery_executor", "cmdb_monitor_binding_delivery_receipt_contract"}, deliveryRequestRef},
		{"effect", "cmdb_monitor_binding_effect_receipt_contract", []string{"cmdb_monitor_binding_effect_probe", "cmdb_monitor_binding_effect_receipt_contract"}, effectRequestRef},
		{"rollback", "binding_rollback_contract", []string{"cmdb_monitor_binding_rollback_executor", "binding_rollback_contract"}, rollbackRequestRef},
	}
	out := make([]model.CmdbMonitorBindingReceipt, 0, len(defs))
	for _, def := range defs {
		receipt := model.CmdbMonitorBindingReceipt{
			BindingID:   binding.ID,
			InstanceID:  binding.InstanceID,
			ReceiptType: def.typ,
			Status:      "pending",
			ContractID:  def.contract,
			MissingJSON: cmdbMonitorBindingReceiptMissingJSON(def.missing),
			RequestRef:  def.requestRef,
			AuditRef:    auditRef,
		}
		saved, err := store.SaveCmdbMonitorBindingReceipt(&receipt)
		if err != nil {
			return nil, err
		}
		if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
			ID:           auditRef + "-" + def.typ,
			Actor:        actor,
			Action:       cmdbMonitorBindingReceiptAuditAction(def.typ),
			ResourceType: "cmdb_monitor_binding",
			ResourceID:   binding.InstanceID,
			Scope:        "cmdb",
			Status:       "blocked",
			ClientIP:     clientIP,
			Summary:      "CMDB monitor binding receipt blocked by contract",
			Details: map[string]any{
				"binding_id":        binding.ID,
				"receipt_id":        saved.ID,
				"receipt_type":      def.typ,
				"contract_id":       def.contract,
				"missing_contracts": def.missing,
				"request_ref":       def.requestRef,
			},
		}); err != nil {
			return nil, err
		}
		out = append(out, *saved)
	}
	return out, nil
}

func cmdbMonitorBindingReceiptDTOs(receipts []model.CmdbMonitorBindingReceipt) []gin.H {
	out := make([]gin.H, 0, len(receipts))
	for _, receipt := range receipts {
		out = append(out, gin.H{
			"id":                receipt.ID,
			"binding_id":        receipt.BindingID,
			"instance_id":       receipt.InstanceID,
			"receipt_type":      receipt.ReceiptType,
			"status":            receipt.Status,
			"audit_action":      cmdbMonitorBindingReceiptAuditAction(receipt.ReceiptType),
			"contract_id":       receipt.ContractID,
			"missing_contracts": cmdbMonitorBindingReceiptMissingList(receipt.MissingJSON),
			"request_ref":       receipt.RequestRef,
			"audit_ref":         receipt.AuditRef,
			"created_at":        receipt.CreatedAt,
		})
	}
	return out
}

func cmdbMonitorBindingReceiptAuditAction(receiptType string) string {
	return "cmdb.monitor_binding." + strings.TrimSpace(receiptType) + ".blocked"
}

func cmdbMonitorBindingReceiptStatus(receipts []model.CmdbMonitorBindingReceipt, receiptType string) string {
	for _, receipt := range receipts {
		if receipt.ReceiptType == receiptType {
			return receipt.Status
		}
	}
	return "pending"
}

func cmdbMonitorBindingReceiptMissingJSON(items []string) string {
	raw, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func cmdbMonitorBindingReceiptMissingList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}
