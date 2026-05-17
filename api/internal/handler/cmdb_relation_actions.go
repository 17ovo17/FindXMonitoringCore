package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func CreateCmdbRelationActionRequest(c *gin.Context) {
	var payload cmdbRelationActionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusConflict, cmdbRelationActionBlockedEnvelope(payload, "cmdb relation action requires a valid JSON payload"))
		return
	}
	payload.normalize()
	if !payload.ready() {
		c.JSON(http.StatusConflict, cmdbRelationActionBlockedEnvelope(payload, "cmdb relation action requires action, instance_id, node_id, object_id, and relation_id"))
		return
	}
	if !cmdbRelationActionAllowed(payload.Action) {
		c.JSON(http.StatusConflict, cmdbRelationActionBlockedEnvelope(payload, "cmdb relation action is not supported by the current contract"))
		return
	}
	if _, ok := store.GetCmdbInstance(payload.InstanceID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb instance not found"})
		return
	}
	if _, ok := store.GetCmdbInstance(payload.NodeID); !ok {
		c.JSON(http.StatusConflict, cmdbRelationActionBlockedEnvelope(payload, "cmdb relation action target instance is missing"))
		return
	}
	if !cmdbRelationActionRelationExists(payload.InstanceID, payload.NodeID, payload.RelationID) {
		c.JSON(http.StatusConflict, cmdbRelationActionBlockedEnvelope(payload, "cmdb relation action requires an existing relation row connected to the instance and target"))
		return
	}

	auditRef := "cmdb-relation-action-audit-" + store.NewID()
	item := payload.toModel(requestActor(c), auditRef)
	saved, err := store.SaveCmdbRelationActionRequest(&item)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb relation action store unavailable"})
		return
	}
	if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
		ID:           auditRef,
		Actor:        requestActor(c),
		Action:       cmdbRelationActionAuditAction(saved.Action),
		ResourceType: "cmdb_relation_action",
		ResourceID:   saved.InstanceID,
		Scope:        "cmdb",
		Status:       "recorded",
		ClientIP:     c.ClientIP(),
		Summary:      "CMDB relation action request recorded",
		Details: map[string]any{
			"action_request_id": saved.ID,
			"action":            saved.Action,
			"instance_id":       saved.InstanceID,
			"node_id":           saved.NodeID,
			"object_id":         saved.ObjectID,
			"relation_id":       saved.RelationID,
			"context":           cmdbRelationActionRawJSON(saved.ContextJSON),
			"delivery_status":   saved.DeliveryStatus,
			"effect_status":     saved.EffectStatus,
		},
	}); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb relation action audit unavailable"})
		return
	}
	receipts, err := cmdbCreateRelationActionReceipts(*saved, auditRef, requestActor(c), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb relation action receipt unavailable"})
		return
	}
	c.JSON(http.StatusAccepted, cmdbRelationActionRecordedEnvelopeWithReceipts(*saved, receipts))
}

func GetCmdbRelationActionRequest(c *gin.Context) {
	item, ok := store.GetCmdbRelationActionRequest(strings.TrimSpace(c.Param("id")))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb relation action request not found"})
		return
	}
	if cmdbRelationActionAuditRequested(c) {
		logs, err := cmdbRelationActionAuditLogs(*item)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb relation action audit unavailable"})
			return
		}
		c.JSON(http.StatusOK, cmdbRelationActionAuditEnvelope(*item, logs))
		return
	}
	receipts := store.ListCmdbRelationActionReceipts(item.ID)
	c.JSON(http.StatusOK, cmdbRelationActionDetailEnvelope(*item, receipts))
}

func ListCmdbRelationActionRequests(c *gin.Context) {
	instanceID := strings.TrimSpace(c.Query("instance_id"))
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instance_id is required"})
		return
	}
	if _, ok := store.GetCmdbInstance(instanceID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb instance not found"})
		return
	}
	items := store.ListCmdbRelationActionRequests(instanceID)
	c.JSON(http.StatusOK, cmdbRelationActionListEnvelope(instanceID, items))
}

type cmdbRelationActionPayload struct {
	Action     string         `json:"action"`
	InstanceID string         `json:"instance_id"`
	NodeID     string         `json:"node_id"`
	ObjectID   string         `json:"object_id"`
	RelationID string         `json:"relation_id"`
	Context    map[string]any `json:"context"`
}

func (p *cmdbRelationActionPayload) normalize() {
	p.Action = strings.TrimSpace(p.Action)
	p.InstanceID = strings.TrimSpace(p.InstanceID)
	p.NodeID = strings.TrimSpace(p.NodeID)
	p.ObjectID = strings.TrimSpace(p.ObjectID)
	p.RelationID = strings.TrimSpace(p.RelationID)
}

func (p cmdbRelationActionPayload) ready() bool {
	return p.Action != "" && p.InstanceID != "" && p.NodeID != "" && p.ObjectID != "" && p.RelationID != ""
}

func (p cmdbRelationActionPayload) toModel(actor, auditRef string) model.CmdbRelationActionRequest {
	return model.CmdbRelationActionRequest{
		Action:         p.Action,
		InstanceID:     p.InstanceID,
		NodeID:         p.NodeID,
		ObjectID:       p.ObjectID,
		RelationID:     p.RelationID,
		Actor:          actor,
		Status:         "recorded",
		DeliveryStatus: "pending",
		EffectStatus:   "pending",
		ContextJSON:    cmdbRelationActionJSON(p.Context),
		AuditRef:       auditRef,
	}
}

func cmdbRelationActionAllowed(action string) bool {
	switch strings.TrimSpace(action) {
	case "expand", "detail", "relation", "topology":
		return true
	default:
		return false
	}
}

func cmdbRelationActionRelationExists(instanceID, nodeID, relationID string) bool {
	for _, rel := range store.ListCmdbInstanceRelations(instanceID) {
		if rel.ID != relationID {
			continue
		}
		if rel.SourceInstanceID == instanceID && rel.TargetInstanceID == nodeID {
			return true
		}
		if rel.TargetInstanceID == instanceID && rel.SourceInstanceID == nodeID {
			return true
		}
	}
	return false
}

func cmdbRelationActionJSON(value map[string]any) string {
	if value == nil {
		return "{}"
	}
	raw, err := json.Marshal(cmdbSanitizeBindingValue("", value))
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func cmdbRelationActionRawJSON(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return gin.H{}
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return gin.H{}
	}
	return cmdbSanitizeBindingValue("", value)
}

func cmdbRelationActionRecordedEnvelope(item model.CmdbRelationActionRequest) gin.H {
	return cmdbRelationActionRecordedEnvelopeWithReceipts(item, store.ListCmdbRelationActionReceipts(item.ID))
}

func cmdbRelationActionRecordedEnvelopeWithReceipts(item model.CmdbRelationActionRequest, receipts []model.CmdbRelationActionReceipt) gin.H {
	return gin.H{
		"code":              0,
		"status":            "recorded",
		"action_request":    cmdbRelationActionDTOWithReceipts(item, receipts),
		"audit_ref":         item.AuditRef,
		"expected_schema":   cmdbRelationActionExpectedSchema(),
		"field_matrix":      cmdbRelationActionFieldMatrix("recorded"),
		"missing_contracts": []string{"cmdb_relation_action_delivery_receipt_contract", "cmdb_relation_action_effect_receipt_contract"},
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_relation_action",
			"resource_id":   item.InstanceID,
			"action":        cmdbRelationActionAuditAction(item.Action),
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbRelationActionDetailEnvelope(item model.CmdbRelationActionRequest, receipts []model.CmdbRelationActionReceipt) gin.H {
	return gin.H{
		"code":              0,
		"status":            "recorded",
		"contract":          "cmdb.relation_action.detail.read.v1",
		"action_request":    cmdbRelationActionDTOWithReceipts(item, receipts),
		"action_request_id": item.ID,
		"instance_id":       item.InstanceID,
		"audit_ref":         item.AuditRef,
		"expected_schema":   cmdbRelationActionDetailExpectedSchema(),
		"field_matrix":      cmdbRelationActionFieldMatrix("recorded"),
		"missing_contracts": []string{"cmdb_relation_action_delivery_executor", "cmdb_relation_action_effect_probe"},
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_relation_action",
			"resource_id":   item.InstanceID,
			"actions":       cmdbRelationActionAuditActions(item.Action),
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbRelationActionListEnvelope(instanceID string, items []model.CmdbRelationActionRequest) gin.H {
	out := make([]gin.H, 0, len(items))
	actions := make([]string, 0, len(items)*3)
	seen := map[string]bool{}
	for _, item := range items {
		out = append(out, cmdbRelationActionDTO(item))
		for _, action := range []string{
			cmdbRelationActionAuditAction(item.Action),
			cmdbRelationActionReceiptAuditAction(item.Action, "delivery"),
			cmdbRelationActionReceiptAuditAction(item.Action, "effect"),
		} {
			if !seen[action] {
				seen[action] = true
				actions = append(actions, action)
			}
		}
	}
	return gin.H{
		"code":              0,
		"status":            "ready",
		"contract":          "cmdb.relation_actions.read.v1",
		"instance_id":       instanceID,
		"action_requests":   out,
		"total":             len(out),
		"expected_schema":   cmdbRelationActionListExpectedSchema(),
		"field_matrix":      cmdbRelationActionListFieldMatrix("ready"),
		"missing_contracts": []string{"cmdb_relation_action_delivery_executor", "cmdb_relation_action_effect_probe"},
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_relation_action",
			"resource_id":   instanceID,
			"actions":       actions,
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbRelationActionListReceiptsComplete(items []model.CmdbRelationActionRequest) bool {
	for _, item := range items {
		receipts := store.ListCmdbRelationActionReceipts(item.ID)
		if len(receipts) == 0 || !cmdbRelationActionReceiptsComplete(receipts) {
			return false
		}
	}
	return true
}

func cmdbRelationActionDTO(item model.CmdbRelationActionRequest) gin.H {
	return cmdbRelationActionDTOWithReceipts(item, store.ListCmdbRelationActionReceipts(item.ID))
}

func cmdbRelationActionDTOWithReceipts(item model.CmdbRelationActionRequest, receipts []model.CmdbRelationActionReceipt) gin.H {
	dto := gin.H{
		"id":              item.ID,
		"action":          item.Action,
		"instance_id":     item.InstanceID,
		"node_id":         item.NodeID,
		"object_id":       item.ObjectID,
		"relation_id":     item.RelationID,
		"actor":           item.Actor,
		"status":          item.Status,
		"delivery_status": item.DeliveryStatus,
		"effect_status":   item.EffectStatus,
		"receipts":        cmdbRelationActionReceiptDTOs(receipts),
		"context":         cmdbRelationActionRawJSON(item.ContextJSON),
		"audit_ref":       item.AuditRef,
		"created_at":      item.CreatedAt,
		"updated_at":      item.UpdatedAt,
	}
	if schema, ok := cmdbRelationSchemaFromAction(item); ok {
		dto["relation_schema"] = schema
	}
	return dto
}

func cmdbRelationActionBlockedEnvelope(payload cmdbRelationActionPayload, message string) gin.H {
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  message,
		"contract": "cmdb.relation_action.request.v1",
		"missing_contracts": []string{
			"cmdb_relation_action_store",
			"cmdb_relation_action_target_contract",
			"cmdb_relation_action_delivery_receipt_contract",
			"cmdb_relation_action_effect_receipt_contract",
		},
		"target": gin.H{
			"action":      payload.Action,
			"instance_id": payload.InstanceID,
			"node_id":     payload.NodeID,
			"object_id":   payload.ObjectID,
			"relation_id": payload.RelationID,
		},
		"expected_schema": cmdbRelationActionExpectedSchema(),
		"field_matrix":    cmdbRelationActionFieldMatrix("blocked"),
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbRelationActionListBlockedEnvelope(instanceID, message string) gin.H {
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  message,
		"contract": "cmdb.relation_actions.read.v1",
		"missing_contracts": []string{
			"cmdb_relation_action_store",
			"cmdb_relation_action_receipt_complete_contract",
			"cmdb_relation_action_delivery_receipt_contract",
			"cmdb_relation_action_effect_receipt_contract",
		},
		"instance_id":     strings.TrimSpace(instanceID),
		"expected_schema": cmdbRelationActionListExpectedSchema(),
		"field_matrix":    cmdbRelationActionListFieldMatrix("blocked"),
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbRelationActionDetailBlockedEnvelope(item model.CmdbRelationActionRequest) gin.H {
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  "cmdb relation action detail requires complete stored delivery and effect receipts",
		"contract": "cmdb.relation_action.detail.read.v1",
		"missing_contracts": []string{
			"cmdb_relation_action_receipt_complete_contract",
			"cmdb_relation_action_delivery_receipt_contract",
			"cmdb_relation_action_effect_receipt_contract",
		},
		"action_request_id": item.ID,
		"instance_id":       item.InstanceID,
		"expected_schema":   cmdbRelationActionDetailExpectedSchema(),
		"field_matrix":      cmdbRelationActionFieldMatrix("blocked"),
		"safe_to_retry":     false,
		"meta":              cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbRelationActionExpectedSchema() gin.H {
	return gin.H{
		"request":         []string{"action", "instance_id", "node_id", "object_id", "relation_id", "context"},
		"record":          []string{"id", "action", "instance_id", "node_id", "object_id", "relation_id", "status", "delivery_status", "effect_status", "receipts.request_ref", "audit_ref", "relation_schema"},
		"relation_schema": []string{"left_object_id", "right_object_id", "asst_id", "asst_name", "mapping", "visible", "rule_logic", "rule_expression", "rules"},
		"audit":           []string{"source", "scope", "resource_type", "resource_id", "action"},
	}
}

func cmdbRelationActionListExpectedSchema() gin.H {
	return gin.H{
		"query":             []string{"instance_id"},
		"action_requests[]": []string{"id", "action", "instance_id", "node_id", "object_id", "relation_id", "relation_schema", "status", "delivery_status", "effect_status", "receipts", "audit_ref"},
		"receipts[]":        []string{"id", "action_request_id", "receipt_type", "status", "audit_action", "contract_id", "missing_contracts", "request_ref", "audit_ref", "created_at"},
		"log_query":         []string{"source", "scope", "resource_type", "resource_id", "actions"},
	}
}

func cmdbRelationActionDetailExpectedSchema() gin.H {
	return gin.H{
		"path":              []string{"action_request_id"},
		"action_request":    []string{"id", "action", "instance_id", "node_id", "object_id", "relation_id", "relation_schema", "status", "delivery_status", "effect_status", "receipts", "audit_ref"},
		"required_receipts": []string{"delivery", "effect"},
		"log_query":         []string{"source", "scope", "resource_type", "resource_id", "actions"},
	}
}

func cmdbRelationActionFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("relation_action_target", status, []string{
			"action", "instance_id", "node_id", "object_id", "relation_id",
		}, "cmdb_relation_action_target_contract"),
		cmdbContractFieldGroup("relation_action_store", status, []string{
			"id", "status", "audit_ref", "created_at", "updated_at",
		}, "cmdb_relation_action_store"),
		cmdbContractFieldGroup("relation_action_relation_schema", status, []string{
			"relation_schema.asst_id", "relation_schema.asst_name", "relation_schema.mapping", "relation_schema.visible", "relation_schema.rules",
		}, "cmdb_relation_action_store"),
		cmdbContractFieldGroup("relation_action_delivery_receipt", "blocked", []string{
			"delivery_status", "delivery_receipt_id", "request_ref",
		}, "cmdb_relation_action_delivery_receipt_contract"),
		cmdbContractFieldGroup("relation_action_effect_receipt", "blocked", []string{
			"effect_status", "effect_receipt_id",
		}, "cmdb_relation_action_effect_receipt_contract"),
	}
}

func cmdbRelationActionListFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("relation_action_store", status, []string{
			"id", "action", "instance_id", "node_id", "relation_id", "relation_schema", "status", "audit_ref",
		}, "cmdb_relation_action_store"),
		cmdbContractFieldGroup("relation_action_receipts", status, []string{
			"receipts", "delivery_status", "effect_status", "contract_id", "missing_contracts", "request_ref",
		}, "cmdb_relation_action_delivery_receipt_contract"),
		cmdbContractFieldGroup("relation_action_audit", status, []string{
			"findx_audit", "cmdb.relation.<action>.request", "cmdb.relation.<action>.delivery.blocked", "cmdb.relation.<action>.effect.blocked",
		}, "binding_audit_contract"),
	}
}

func cmdbRelationActionAuditAction(action string) string {
	return "cmdb.relation." + strings.TrimSpace(action) + ".request"
}

func cmdbRelationActionReceiptAuditAction(action, receiptType string) string {
	return "cmdb.relation." + strings.TrimSpace(action) + "." + strings.TrimSpace(receiptType) + ".blocked"
}
