package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func cmdbRelationActionRuntimeReceiptsResolved(item model.CmdbRelationActionRequest, receipts []model.CmdbRelationActionReceipt) bool {
	for _, receipt := range receipts {
		receiptType := strings.TrimSpace(receipt.ReceiptType)
		if receiptType != "delivery" && receiptType != "effect" {
			continue
		}
		if !cmdbRelationActionRuntimeReceiptResolved(item, receipt) {
			return false
		}
	}
	return true
}

func cmdbRelationActionRuntimeReceiptResolved(item model.CmdbRelationActionRequest, receipt model.CmdbRelationActionReceipt) bool {
	task, ok, err := store.GetFindXAgentExecutionTask(strings.TrimSpace(receipt.RequestRef))
	if err != nil || !ok {
		return false
	}
	if strings.TrimSpace(task.Status) != "blocked" {
		return false
	}
	receiptType := strings.TrimSpace(receipt.ReceiptType)
	expectedAction := "cmdb-relation-" + strings.TrimSpace(item.Action) + "-" + receiptType
	if strings.TrimSpace(task.Action) != expectedAction {
		return false
	}
	if len(task.TargetIDs) != 1 || strings.TrimSpace(task.TargetIDs[0]) != strings.TrimSpace(item.NodeID) {
		return false
	}
	return cmdbRelationActionRuntimeMetadataMatches(item, receiptType, task.Metadata)
}

func cmdbRelationActionRuntimeMetadataMatches(item model.CmdbRelationActionRequest, receiptType string, metadata map[string]string) bool {
	required := map[string]string{
		"cmdb_relation_action_id": item.ID,
		"cmdb_relation_action":    item.Action,
		"cmdb_instance_id":        item.InstanceID,
		"cmdb_node_id":            item.NodeID,
		"cmdb_relation_id":        item.RelationID,
	}
	for key, expected := range required {
		if strings.TrimSpace(metadata[key]) != strings.TrimSpace(expected) {
			return false
		}
	}
	return true
}

func cmdbRelationActionListRuntimeReceiptsResolved(items []model.CmdbRelationActionRequest) bool {
	for _, item := range items {
		receipts := store.ListCmdbRelationActionReceipts(item.ID)
		if !cmdbRelationActionRuntimeReceiptsResolved(item, receipts) {
			return false
		}
	}
	return true
}

func cmdbRelationActionRuntimeExecutorsAttested(item model.CmdbRelationActionRequest, receipts []model.CmdbRelationActionReceipt) bool {
	return false
}

func cmdbRelationActionListExecutorsAttested(items []model.CmdbRelationActionRequest) bool {
	for _, item := range items {
		if !cmdbRelationActionRuntimeExecutorsAttested(item, store.ListCmdbRelationActionReceipts(item.ID)) {
			return false
		}
	}
	return true
}

func cmdbRelationActionRuntimeReadBlockedEnvelope(item model.CmdbRelationActionRequest, message string) gin.H {
	return cmdbRelationActionRuntimeBlockedEnvelope(item.ID, item.InstanceID, message, cmdbRelationActionDetailExpectedSchema(), cmdbRelationActionRuntimeFieldMatrix("blocked"))
}

func cmdbRelationActionRuntimeReceiptsBlockedEnvelope(item model.CmdbRelationActionRequest) gin.H {
	return cmdbRelationActionRuntimeBlockedEnvelope(item.ID, item.InstanceID, "cmdb relation action receipts require request_ref rows to resolve to blocked execution tasks", cmdbRelationActionReceiptsExpectedSchema(), cmdbRelationActionRuntimeFieldMatrix("blocked"))
}

func cmdbRelationActionRuntimeListBlockedEnvelope(instanceID, message string) gin.H {
	return cmdbRelationActionRuntimeBlockedEnvelope("", instanceID, message, cmdbRelationActionListExpectedSchema(), cmdbRelationActionRuntimeFieldMatrix("blocked"))
}

func cmdbRelationActionRuntimeBlockedEnvelope(actionRequestID, instanceID, message string, expectedSchema any, fieldMatrix []gin.H) gin.H {
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  message,
		"contract": "cmdb.relation_action.runtime.read.v1",
		"missing_contracts": []string{
			"cmdb_relation_action_request_ref_resolve_contract",
			"cmdb_relation_action_execution_task_contract",
			"cmdb_relation_action_delivery_receipt_contract",
			"cmdb_relation_action_effect_receipt_contract",
		},
		"action_request_id": strings.TrimSpace(actionRequestID),
		"instance_id":       strings.TrimSpace(instanceID),
		"expected_schema":   expectedSchema,
		"field_matrix":      fieldMatrix,
		"safe_to_retry":     false,
		"meta":              cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbRelationActionExecutorBlockedEnvelope(actionRequestID, instanceID, message string) gin.H {
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  message,
		"contract": "cmdb.relation_action.runtime.read.v1",
		"missing_contracts": uniquePackageRepositoryBlockers([]string{
			"cmdb_relation_action_store_contract",
			"cmdb_relation_action_action_executor_contract",
			"cmdb_relation_action_delivery_executor_contract",
			"cmdb_relation_action_effect_executor_contract",
			"cmdb_relation_action_attested_receipt_contract",
			"cmdb_relation_action_delivery_receipt_contract",
			"cmdb_relation_action_effect_receipt_contract",
		}),
		"action_request_id": strings.TrimSpace(actionRequestID),
		"instance_id":       strings.TrimSpace(instanceID),
		"expected_schema": gin.H{
			"action_request":    []string{"id", "action", "instance_id", "node_id", "object_id", "relation_id"},
			"required_receipts": []string{"delivery", "effect"},
			"required_executor": []string{"executor_registration", "runner_identity", "target_binding", "request_ref_match", "attested_receipt"},
		},
		"field_matrix":  cmdbRelationActionExecutorFieldMatrix("blocked"),
		"safe_to_retry": false,
		"meta":          cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbRelationActionRuntimeFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("relation_action_receipt_ref", status, []string{
			"receipt_type", "request_ref", "action_request_id",
		}, "cmdb_relation_action_request_ref_resolve_contract"),
		cmdbContractFieldGroup("relation_action_execution_task", status, []string{
			"findx_agent_execution_tasks.id", "action", "status", "target_ids", "metadata.cmdb_relation_action_id",
		}, "cmdb_relation_action_execution_task_contract"),
		cmdbContractFieldGroup("relation_action_delivery_receipt", "blocked", []string{
			"delivery_status", "request_ref", "cmdb_relation_action_delivery_executor",
		}, "cmdb_relation_action_delivery_receipt_contract"),
		cmdbContractFieldGroup("relation_action_effect_receipt", "blocked", []string{
			"effect_status", "request_ref", "cmdb_relation_action_effect_probe",
		}, "cmdb_relation_action_effect_receipt_contract"),
	}
}

func cmdbRelationActionExecutorFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("relation_action_store", status, []string{
			"cmdb_relation_action_requests.id", "action", "instance_id", "node_id", "relation_id",
		}, "cmdb_relation_action_store_contract"),
		cmdbContractFieldGroup("relation_action_executor", status, []string{
			"executor_registration", "runner_identity", "target_binding", "request_ref_match",
		}, "cmdb_relation_action_action_executor_contract"),
		cmdbContractFieldGroup("relation_action_delivery_executor", status, []string{
			"delivery_executor", "delivery_request_ref", "delivery_attested_receipt",
		}, "cmdb_relation_action_delivery_executor_contract"),
		cmdbContractFieldGroup("relation_action_effect_executor", status, []string{
			"effect_executor", "effect_request_ref", "effect_attested_receipt",
		}, "cmdb_relation_action_effect_executor_contract"),
		cmdbContractFieldGroup("relation_action_attested_receipt", status, []string{
			"attestation", "receipt_signature", "runner_identity", "operation_context",
		}, "cmdb_relation_action_attested_receipt_contract"),
	}
}
