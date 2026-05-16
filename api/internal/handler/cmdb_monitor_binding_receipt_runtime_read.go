package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func cmdbMonitorBindingRuntimeReceiptsResolvedForBindings(bindings []model.CmdbMonitorBinding) bool {
	for _, binding := range bindings {
		receipts := store.ListCmdbMonitorBindingReceipts(binding.ID)
		if !cmdbMonitorBindingRuntimeReceiptsResolved(binding, receipts) {
			return false
		}
	}
	return true
}

func cmdbMonitorBindingRuntimeExecutorsAttestedForBindings(bindings []model.CmdbMonitorBinding) bool {
	for _, binding := range bindings {
		if !cmdbMonitorBindingRuntimeExecutorsAttested(binding, store.ListCmdbMonitorBindingReceipts(binding.ID)) {
			return false
		}
	}
	return true
}

func cmdbMonitorBindingRuntimeExecutorsAttested(binding model.CmdbMonitorBinding, receipts []model.CmdbMonitorBindingReceipt) bool {
	return false
}

func cmdbMonitorBindingRuntimeReceiptsResolved(binding model.CmdbMonitorBinding, receipts []model.CmdbMonitorBindingReceipt) bool {
	if len(receipts) == 0 || !cmdbMonitorBindingReceiptsComplete(receipts) {
		return false
	}
	for _, receipt := range receipts {
		switch strings.TrimSpace(receipt.ReceiptType) {
		case "delivery", "effect", "rollback":
			if !cmdbMonitorBindingExecutionRequestResolved(binding, receipt) {
				return false
			}
		}
	}
	return true
}

func cmdbMonitorBindingExecutionRequestResolved(binding model.CmdbMonitorBinding, receipt model.CmdbMonitorBindingReceipt) bool {
	receiptType := strings.TrimSpace(receipt.ReceiptType)
	task, ok, err := store.GetFindXAgentExecutionTask(strings.TrimSpace(receipt.RequestRef))
	if err != nil || !ok {
		return false
	}
	if strings.TrimSpace(task.Status) != "blocked" {
		return false
	}
	if strings.TrimSpace(task.Action) != "cmdb-monitor-binding-"+receiptType {
		return false
	}
	if len(task.TargetIDs) != 1 || strings.TrimSpace(task.TargetIDs[0]) != strings.TrimSpace(binding.HostID) {
		return false
	}
	return cmdbMonitorBindingRuntimeMetadataMatches(binding, receiptType, task.Metadata)
}

func cmdbMonitorBindingRuntimeMetadataMatches(binding model.CmdbMonitorBinding, receiptType string, metadata map[string]string) bool {
	required := map[string]string{
		"cmdb_binding_id":   binding.ID,
		"cmdb_instance_id":  binding.InstanceID,
		"cmdb_hostid":       binding.HostID,
		"cmdb_templateid":   binding.TemplateID,
		"cmdb_attr_id":      binding.CmdbAttrID,
		"server_attr_id":    binding.ServerAttrID,
		"cmdb_receipt_type": receiptType,
	}
	for key, expected := range required {
		if strings.TrimSpace(metadata[key]) != strings.TrimSpace(expected) {
			return false
		}
	}
	return true
}

func cmdbMonitorBindingRuntimeReceiptsBlockedEnvelope(instanceID, bindingID string) gin.H {
	contract := cmdbMonitorBindingReceiptsReadContract()
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  "CMDB monitor binding receipts require request_ref rows to resolve to blocked delivery/effect/rollback tasks.",
		"contract": "cmdb.monitor_binding.runtime.receipts.read.v1",
		"missing_contracts": []string{
			"cmdb_monitor_binding_request_ref_resolve_contract",
			"cmdb_monitor_binding_execution_task_contract",
			"cmdb_monitor_binding_delivery_receipt_contract",
			"cmdb_monitor_binding_effect_receipt_contract",
			"binding_rollback_contract",
		},
		"instance_id":     strings.TrimSpace(instanceID),
		"binding_id":      strings.TrimSpace(bindingID),
		"expected_schema": contract["expected_schema"],
		"field_matrix":    cmdbMonitorBindingRuntimeReceiptsFieldMatrix("blocked"),
		"source_evidence": contract["source_evidence"],
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingRuntimeExecutorBlockedEnvelope(instanceID, bindingID string) gin.H {
	contract := cmdbMonitorBindingReceiptsReadContract()
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  "CMDB monitor binding receipts require registered delivery, effect and rollback executors with attested receipts.",
		"contract": "cmdb.monitor_binding.runtime.receipts.read.v1",
		"missing_contracts": uniquePackageRepositoryBlockers([]string{
			"cmdb_monitor_binding_delivery_executor_contract",
			"cmdb_monitor_binding_effect_executor_contract",
			"cmdb_monitor_binding_rollback_executor_contract",
			"cmdb_monitor_binding_attested_receipt_contract",
			"cmdb_monitor_binding_delivery_receipt_contract",
			"cmdb_monitor_binding_effect_receipt_contract",
			"binding_rollback_contract",
		}),
		"instance_id":     strings.TrimSpace(instanceID),
		"binding_id":      strings.TrimSpace(bindingID),
		"expected_schema": contract["expected_schema"],
		"field_matrix":    cmdbMonitorBindingRuntimeExecutorFieldMatrix("blocked"),
		"source_evidence": contract["source_evidence"],
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingRuntimeReceiptsFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("binding_receipt_ref", status, []string{
			"binding_id", "receipt_type", "request_ref",
		}, "cmdb_monitor_binding_request_ref_resolve_contract"),
		cmdbContractFieldGroup("binding_execution_task", status, []string{
			"findx_agent_execution_tasks.id", "action", "status", "target_ids", "metadata.cmdb_binding_id",
		}, "cmdb_monitor_binding_execution_task_contract"),
		cmdbContractFieldGroup("binding_delivery_receipt", "blocked", []string{
			"delivery_status", "request_ref", "cmdb_monitor_binding_delivery_executor",
		}, "cmdb_monitor_binding_delivery_receipt_contract"),
		cmdbContractFieldGroup("binding_effect_receipt", "blocked", []string{
			"effect_status", "request_ref", "cmdb_monitor_binding_effect_probe",
		}, "cmdb_monitor_binding_effect_receipt_contract"),
		cmdbContractFieldGroup("binding_rollback_receipt", "blocked", []string{
			"rollback_status", "request_ref", "cmdb_monitor_binding_rollback_executor",
		}, "binding_rollback_contract"),
	}
}

func cmdbMonitorBindingRuntimeExecutorFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("binding_delivery_executor", status, []string{
			"delivery_executor_registration", "delivery_runner_identity", "delivery_request_ref", "delivery_attested_receipt",
		}, "cmdb_monitor_binding_delivery_executor_contract"),
		cmdbContractFieldGroup("binding_effect_executor", status, []string{
			"effect_executor_registration", "effect_runner_identity", "effect_request_ref", "effect_attested_receipt",
		}, "cmdb_monitor_binding_effect_executor_contract"),
		cmdbContractFieldGroup("binding_rollback_executor", status, []string{
			"rollback_executor_registration", "rollback_runner_identity", "rollback_operation_context", "rollback_attested_receipt",
		}, "cmdb_monitor_binding_rollback_executor_contract"),
		cmdbContractFieldGroup("binding_attested_receipt", status, []string{
			"attestation", "receipt_signature", "runner_identity", "target_binding",
		}, "cmdb_monitor_binding_attested_receipt_contract"),
	}
}
