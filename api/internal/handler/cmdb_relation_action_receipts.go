package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func GetCmdbRelationActionReceipts(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	item, ok := store.GetCmdbRelationActionRequest(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb relation action request not found"})
		return
	}
	receipts := store.ListCmdbRelationActionReceipts(item.ID)
	c.JSON(http.StatusOK, cmdbRelationActionReceiptsEnvelope(*item, receipts))
}

func cmdbRelationActionReceiptsEnvelope(item model.CmdbRelationActionRequest, receipts []model.CmdbRelationActionReceipt) gin.H {
	return gin.H{
		"code":              0,
		"status":            "ready",
		"contract":          "cmdb.relation_action.receipts.read.v1",
		"action_request_id": item.ID,
		"instance_id":       item.InstanceID,
		"action":            item.Action,
		"receipts":          cmdbRelationActionReceiptDTOs(receipts),
		"total":             len(receipts),
		"expected_schema":   cmdbRelationActionReceiptsExpectedSchema(),
		"field_matrix":      cmdbRelationActionReceiptsFieldMatrix("ready"),
		"missing_contracts": []string{"cmdb_relation_action_delivery_executor", "cmdb_relation_action_effect_probe"},
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_relation_action",
			"resource_id":   item.InstanceID,
			"actions": []string{
				cmdbRelationActionReceiptAuditAction(item.Action, "delivery"),
				cmdbRelationActionReceiptAuditAction(item.Action, "effect"),
			},
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbCreateRelationActionReceipts(item model.CmdbRelationActionRequest, auditRef, actor, clientIP string) ([]model.CmdbRelationActionReceipt, error) {
	deliveryRequestRef, err := cmdbCreateRelationActionDeliveryRequest(item, actor)
	if err != nil {
		return nil, err
	}
	effectRequestRef, err := cmdbCreateRelationActionEffectRequest(item, actor)
	if err != nil {
		return nil, err
	}
	defs := []struct {
		typ      string
		contract string
		missing  []string
		request  string
	}{
		{"delivery", "cmdb_relation_action_delivery_receipt_contract", []string{"cmdb_relation_action_delivery_executor", "cmdb_relation_action_delivery_receipt_contract"}, deliveryRequestRef},
		{"effect", "cmdb_relation_action_effect_receipt_contract", []string{"cmdb_relation_action_effect_probe", "cmdb_relation_action_effect_receipt_contract"}, effectRequestRef},
	}
	out := make([]model.CmdbRelationActionReceipt, 0, len(defs))
	for _, def := range defs {
		receipt := model.CmdbRelationActionReceipt{
			ActionRequestID: item.ID,
			Action:          item.Action,
			InstanceID:      item.InstanceID,
			NodeID:          item.NodeID,
			RelationID:      item.RelationID,
			ReceiptType:     def.typ,
			Status:          "pending",
			ContractID:      def.contract,
			MissingJSON:     cmdbMonitorBindingReceiptMissingJSON(def.missing),
			RequestRef:      def.request,
			AuditRef:        auditRef,
		}
		saved, err := store.SaveCmdbRelationActionReceipt(&receipt)
		if err != nil {
			return nil, err
		}
		if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
			ID:           auditRef + "-" + def.typ,
			Actor:        actor,
			Action:       cmdbRelationActionReceiptAuditAction(item.Action, def.typ),
			ResourceType: "cmdb_relation_action",
			ResourceID:   item.InstanceID,
			Scope:        "cmdb",
			Status:       "blocked",
			ClientIP:     clientIP,
			Summary:      "CMDB relation action receipt blocked by contract",
			Details: map[string]any{
				"action_request_id": item.ID,
				"receipt_id":        saved.ID,
				"receipt_type":      def.typ,
				"action":            item.Action,
				"node_id":           item.NodeID,
				"relation_id":       item.RelationID,
				"contract_id":       def.contract,
				"missing_contracts": def.missing,
				"request_ref":       saved.RequestRef,
			},
		}); err != nil {
			return nil, err
		}
		out = append(out, *saved)
	}
	return out, nil
}

func cmdbRelationActionReceiptDTOs(receipts []model.CmdbRelationActionReceipt) []gin.H {
	out := make([]gin.H, 0, len(receipts))
	for _, receipt := range receipts {
		out = append(out, gin.H{
			"id":                receipt.ID,
			"action_request_id": receipt.ActionRequestID,
			"action":            receipt.Action,
			"instance_id":       receipt.InstanceID,
			"node_id":           receipt.NodeID,
			"relation_id":       receipt.RelationID,
			"receipt_type":      receipt.ReceiptType,
			"status":            receipt.Status,
			"audit_action":      cmdbRelationActionReceiptAuditAction(receipt.Action, receipt.ReceiptType),
			"contract_id":       receipt.ContractID,
			"missing_contracts": cmdbMonitorBindingReceiptMissingList(receipt.MissingJSON),
			"request_ref":       receipt.RequestRef,
			"audit_ref":         receipt.AuditRef,
			"created_at":        receipt.CreatedAt,
		})
	}
	return out
}

func cmdbRelationActionReceiptsComplete(receipts []model.CmdbRelationActionReceipt) bool {
	seen := map[string]bool{}
	for _, receipt := range receipts {
		receiptType := strings.TrimSpace(receipt.ReceiptType)
		if receiptType != "" && strings.TrimSpace(receipt.RequestRef) != "" {
			seen[receiptType] = true
		}
	}
	for _, required := range []string{"delivery", "effect"} {
		if !seen[required] {
			return false
		}
	}
	return true
}

func cmdbRelationActionReceiptsBlockedEnvelope(item model.CmdbRelationActionRequest) gin.H {
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  "cmdb relation action receipts require stored receipt rows; empty receipts cannot satisfy delivery or effect contracts",
		"contract": "cmdb.relation_action.receipts.read.v1",
		"missing_contracts": []string{
			"cmdb_relation_action_receipt_complete_contract",
			"cmdb_relation_action_request_ref_contract",
			"cmdb_relation_action_delivery_receipt_contract",
			"cmdb_relation_action_effect_receipt_contract",
		},
		"action_request_id": item.ID,
		"instance_id":       item.InstanceID,
		"expected_schema":   cmdbRelationActionReceiptsExpectedSchema(),
		"field_matrix":      cmdbRelationActionReceiptsFieldMatrix("blocked"),
		"safe_to_retry":     false,
		"meta":              cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbRelationActionReceiptsExpectedSchema() gin.H {
	return gin.H{
		"path":       []string{"action_request_id"},
		"receipts[]": []string{"id", "action_request_id", "receipt_type", "status", "audit_action", "contract_id", "missing_contracts", "request_ref", "audit_ref", "created_at"},
		"log_query":  []string{"source", "scope", "resource_type", "resource_id", "actions"},
	}
}

func cmdbRelationActionReceiptsFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("relation_action_receipt_store", status, []string{
			"action_request_id", "receipt_type", "status", "contract_id", "missing_contracts", "request_ref",
		}, "cmdb_relation_action_delivery_receipt_contract"),
		cmdbContractFieldGroup("relation_action_delivery_receipt", "blocked", []string{
			"delivery_status", "request_ref", "cmdb_relation_action_delivery_executor",
		}, "cmdb_relation_action_delivery_receipt_contract"),
		cmdbContractFieldGroup("relation_action_effect_receipt", "blocked", []string{
			"effect_status", "request_ref", "cmdb_relation_action_effect_probe",
		}, "cmdb_relation_action_effect_receipt_contract"),
		cmdbContractFieldGroup("relation_action_receipt_audit", status, []string{
			"findx_audit", "cmdb.relation.<action>.delivery.blocked", "cmdb.relation.<action>.effect.blocked",
		}, "binding_audit_contract"),
	}
}
