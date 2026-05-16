package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func cmdbRelationActionAuditRequested(c *gin.Context) bool {
	include := strings.ToLower(strings.TrimSpace(c.Query("include")))
	if include == "audit" || strings.Contains(include, "audit") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(c.Query("audit")), "true")
}

func cmdbRelationActionAuditLogs(item model.CmdbRelationActionRequest) ([]model.LogRecord, error) {
	actions := cmdbRelationActionAuditActions(item.Action)
	out := make([]model.LogRecord, 0, len(actions))
	for _, action := range actions {
		resp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
			Source:       model.LogsSourceFindXAudit,
			Scope:        "cmdb",
			ResourceType: "cmdb_relation_action",
			ResourceID:   item.InstanceID,
			Action:       action,
			Limit:        10,
		})
		if err != nil {
			return nil, err
		}
		if len(resp.Items) == 0 {
			return nil, nil
		}
		out = append(out, resp.Items[0])
	}
	return out, nil
}

func cmdbRelationActionAuditEnvelope(item model.CmdbRelationActionRequest, logs []model.LogRecord) gin.H {
	return gin.H{
		"code":              0,
		"status":            "ready",
		"contract":          "cmdb.relation_action.audit.read.v1",
		"action_request":    cmdbRelationActionDTO(item),
		"action_request_id": item.ID,
		"instance_id":       item.InstanceID,
		"audit_logs":        logs,
		"total":             len(logs),
		"expected_schema":   cmdbRelationActionAuditExpectedSchema(),
		"field_matrix":      cmdbRelationActionAuditFieldMatrix("ready"),
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

func cmdbRelationActionAuditBlockedEnvelope(item model.CmdbRelationActionRequest) gin.H {
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  "cmdb relation action audit requires stored findx_audit rows for request, delivery receipt, and effect receipt",
		"contract": "cmdb.relation_action.audit.read.v1",
		"missing_contracts": []string{
			"cmdb_relation_action_audit_log_contract",
			"cmdb_relation_action_delivery_receipt_audit_contract",
			"cmdb_relation_action_effect_receipt_audit_contract",
		},
		"action_request_id": item.ID,
		"instance_id":       item.InstanceID,
		"expected_schema":   cmdbRelationActionAuditExpectedSchema(),
		"field_matrix":      cmdbRelationActionAuditFieldMatrix("blocked"),
		"safe_to_retry":     false,
		"meta":              cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbRelationActionAuditExpectedSchema() gin.H {
	return gin.H{
		"query":        []string{"include=audit"},
		"audit_logs[]": []string{"id", "timestamp", "source", "body", "attributes"},
		"log_query":    []string{"source", "scope", "resource_type", "resource_id", "actions"},
	}
}

func cmdbRelationActionAuditFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("relation_action_request_audit", status, []string{
			"findx_audit", "cmdb.relation.<action>.request", "resource_type=cmdb_relation_action",
		}, "cmdb_relation_action_audit_log_contract"),
		cmdbContractFieldGroup("relation_action_delivery_receipt_audit", status, []string{
			"findx_audit", "cmdb.relation.<action>.delivery.blocked",
		}, "cmdb_relation_action_delivery_receipt_audit_contract"),
		cmdbContractFieldGroup("relation_action_effect_receipt_audit", status, []string{
			"findx_audit", "cmdb.relation.<action>.effect.blocked",
		}, "cmdb_relation_action_effect_receipt_audit_contract"),
	}
}

func cmdbRelationActionAuditActions(action string) []string {
	return []string{
		cmdbRelationActionAuditAction(action),
		cmdbRelationActionReceiptAuditAction(action, "delivery"),
		cmdbRelationActionReceiptAuditAction(action, "effect"),
	}
}
