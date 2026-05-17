package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func cmdbMonitorBindingAuditRequested(c *gin.Context) bool {
	include := strings.ToLower(strings.TrimSpace(c.Query("include")))
	if include == "audit" || strings.Contains(include, "audit") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(c.Query("audit")), "true")
}

func GetCmdbMonitorBindingAudit(c *gin.Context) {
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
	logs, err := cmdbMonitorBindingAuditLogs(instanceID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb monitor binding audit unavailable"})
		return
	}
	c.JSON(http.StatusOK, cmdbMonitorBindingAuditReadyEnvelope(instanceID, bindings, logs))
}

func cmdbMonitorBindingAuditLogs(instanceID string) ([]model.LogRecord, error) {
	actions := cmdbMonitorBindingAuditActions()
	out := make([]model.LogRecord, 0, len(actions))
	for _, action := range actions {
		resp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
			Source:       model.LogsSourceFindXAudit,
			Scope:        "cmdb",
			ResourceType: "cmdb_monitor_binding",
			ResourceID:   instanceID,
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

func cmdbMonitorBindingAuditReadyEnvelope(instanceID string, bindings []model.CmdbMonitorBinding, logs []model.LogRecord) gin.H {
	return gin.H{
		"code":              0,
		"status":            "ready",
		"contract":          "cmdb.monitor_binding.audit.read.v1",
		"instance_id":       instanceID,
		"binding_ids":       cmdbMonitorBindingIDs(bindings),
		"audit_logs":        logs,
		"total":             len(logs),
		"expected_schema":   cmdbMonitorBindingAuditExpectedSchema(),
		"field_matrix":      cmdbMonitorBindingAuditFieldMatrix("ready"),
		"source_evidence":   cmdbMonitorBindingsWriteBlockedContract()["source_evidence"],
		"missing_contracts": []string{"cmdb_monitor_binding_delivery_executor", "cmdb_monitor_binding_effect_probe", "cmdb_monitor_binding_rollback_executor"},
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_monitor_binding",
			"resource_id":   instanceID,
			"actions":       cmdbMonitorBindingAuditActions(),
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingAuditBlockedEnvelope(instanceID string) gin.H {
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  "CMDB monitor binding audit requires stored binding rows and findx_audit rows for save, delivery, effect, and rollback.",
		"contract": "cmdb.monitor_binding.audit.read.v1",
		"missing_contracts": []string{
			"cmdb_monitor_binding_store",
			"cmdb_monitor_binding_audit_log_contract",
			"cmdb_monitor_binding_delivery_receipt_audit_contract",
			"cmdb_monitor_binding_effect_receipt_audit_contract",
			"cmdb_monitor_binding_rollback_receipt_audit_contract",
		},
		"instance_id":     strings.TrimSpace(instanceID),
		"expected_schema": cmdbMonitorBindingAuditExpectedSchema(),
		"field_matrix":    cmdbMonitorBindingAuditFieldMatrix("blocked"),
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingAuditExpectedSchema() gin.H {
	return gin.H{
		"query":        []string{"include=audit"},
		"audit_logs[]": []string{"id", "timestamp", "source", "body", "attributes"},
		"log_query":    []string{"source", "scope", "resource_type", "resource_id", "actions"},
	}
}

func cmdbMonitorBindingAuditFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("binding_save_audit", status, []string{
			"findx_audit", "cmdb.monitor_binding.save", "resource_type=cmdb_monitor_binding",
		}, "cmdb_monitor_binding_audit_log_contract"),
		cmdbContractFieldGroup("binding_delivery_receipt_audit", status, []string{
			"findx_audit", "cmdb.monitor_binding.delivery.blocked",
		}, "cmdb_monitor_binding_delivery_receipt_audit_contract"),
		cmdbContractFieldGroup("binding_effect_receipt_audit", status, []string{
			"findx_audit", "cmdb.monitor_binding.effect.blocked",
		}, "cmdb_monitor_binding_effect_receipt_audit_contract"),
		cmdbContractFieldGroup("binding_rollback_receipt_audit", status, []string{
			"findx_audit", "cmdb.monitor_binding.rollback.blocked",
		}, "cmdb_monitor_binding_rollback_receipt_audit_contract"),
	}
}

func cmdbMonitorBindingAuditActions() []string {
	return []string{
		"cmdb.monitor_binding.save",
		"cmdb.monitor_binding.delivery.blocked",
		"cmdb.monitor_binding.effect.blocked",
		"cmdb.monitor_binding.rollback.blocked",
	}
}
