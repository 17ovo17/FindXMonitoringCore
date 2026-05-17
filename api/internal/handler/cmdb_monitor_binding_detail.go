package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func GetCmdbMonitorBindingDetail(c *gin.Context) {
	parts := cmdbMonitorBindingPathParts(c)
	if len(parts) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cmdb monitor binding detail requires instance_id and binding_id"})
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
	if binding.InstanceID != instanceID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cmdb monitor binding does not belong to the requested instance"})
		return
	}
	c.JSON(http.StatusOK, cmdbMonitorBindingDetailReadyEnvelope(instanceID, *binding))
}

func cmdbMonitorBindingDetailPath(c *gin.Context) bool {
	parts := cmdbMonitorBindingPathParts(c)
	if len(parts) < 2 {
		return false
	}
	switch strings.TrimSpace(parts[1]) {
	case "receipts":
		return false
	default:
		return true
	}
}

func cmdbMonitorBindingDetailReadyEnvelope(instanceID string, binding model.CmdbMonitorBinding) gin.H {
	return gin.H{
		"code":              0,
		"status":            "ready",
		"contract":          "cmdb.monitor_binding.detail.read.v1",
		"instance_id":       instanceID,
		"binding":           cmdbMonitorBindingDTO(binding),
		"expected_schema":   cmdbMonitorBindingDetailExpectedSchema(),
		"field_matrix":      cmdbMonitorBindingDetailFieldMatrix("ready"),
		"source_evidence":   cmdbMonitorBindingsReadBlockedContract()["source_evidence"],
		"missing_contracts": []string{"cmdb_monitor_binding_delivery_executor", "cmdb_monitor_binding_effect_probe", "cmdb_monitor_binding_rollback_executor"},
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_monitor_binding",
			"resource_id":   instanceID,
			"action":        "cmdb.monitor_binding.save",
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingDetailBlockedEnvelope(instanceID, bindingID, message string) gin.H {
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  message,
		"contract": "cmdb.monitor_binding.detail.read.v1",
		"missing_contracts": []string{
			"cmdb_monitor_binding_detail_contract",
			"cmdb_monitor_binding_instance_match_contract",
		},
		"instance_id":     strings.TrimSpace(instanceID),
		"binding_id":      strings.TrimSpace(bindingID),
		"expected_schema": cmdbMonitorBindingDetailExpectedSchema(),
		"field_matrix":    cmdbMonitorBindingDetailFieldMatrix("blocked"),
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingDetailExpectedSchema() gin.H {
	return gin.H{
		"path":     []string{"instance_id", "binding_id"},
		"binding":  []string{"id", "instance_id", "hostid", "templateid", "cmdb_attr_id", "server_attr_id", "receipts", "audit_ref"},
		"receipts": []string{"delivery", "effect", "rollback"},
	}
}

func cmdbMonitorBindingDetailFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("binding_detail_store", status, []string{
			"id", "instance_id", "hostid", "templateid", "audit_ref",
		}, "cmdb_monitor_binding_detail_contract"),
		cmdbContractFieldGroup("binding_instance_match", status, []string{
			"instance_id", "binding_id",
		}, "cmdb_monitor_binding_instance_match_contract"),
		cmdbContractFieldGroup("binding_detail_receipts", "blocked", []string{
			"delivery_status", "effect_status", "rollback_status",
		}, "cmdb_monitor_binding_receipt_complete_contract"),
	}
}
