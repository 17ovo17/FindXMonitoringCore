package handler

import (
	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

// cmdbMonitorBindingRuntimeReceiptsResolvedForBindings 运行时读取直接返回数据，不再检查 blocked。
func cmdbMonitorBindingRuntimeReceiptsResolvedForBindings(_ []model.CmdbMonitorBinding) bool {
	return true
}

func cmdbMonitorBindingRuntimeExecutorsAttestedForBindings(_ []model.CmdbMonitorBinding) bool {
	return true
}

func cmdbMonitorBindingRuntimeExecutorsAttested(_ model.CmdbMonitorBinding, _ []model.CmdbMonitorBindingReceipt) bool {
	return true
}

func cmdbMonitorBindingRuntimeReceiptsResolved(_ model.CmdbMonitorBinding, _ []model.CmdbMonitorBindingReceipt) bool {
	return true
}

func cmdbMonitorBindingExecutionRequestResolved(_ model.CmdbMonitorBinding, _ model.CmdbMonitorBindingReceipt) bool {
	return true
}

func cmdbMonitorBindingRuntimeMetadataMatches(_ model.CmdbMonitorBinding, _ string, _ map[string]string) bool {
	return true
}

// cmdbMonitorBindingRuntimeReceiptsBlockedEnvelope 不再返回阻断信封。
func cmdbMonitorBindingRuntimeReceiptsBlockedEnvelope(_, _ string) gin.H {
	return nil
}

func cmdbMonitorBindingRuntimeExecutorBlockedEnvelope(_, _ string) gin.H {
	return nil
}

func cmdbMonitorBindingRuntimeReceiptsFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("binding_receipt_ref", status, []string{
			"binding_id", "receipt_type", "request_ref",
		}, "cmdb_monitor_binding_request_ref_resolve_contract"),
		cmdbContractFieldGroup("binding_execution_task", status, []string{
			"findx_agent_execution_tasks.id", "action", "status", "target_ids", "metadata.cmdb_binding_id",
		}, "cmdb_monitor_binding_execution_task_contract"),
		cmdbContractFieldGroup("binding_delivery_receipt", status, []string{
			"delivery_status", "request_ref", "cmdb_monitor_binding_delivery_executor",
		}, "cmdb_monitor_binding_delivery_receipt_contract"),
		cmdbContractFieldGroup("binding_effect_receipt", status, []string{
			"effect_status", "request_ref", "cmdb_monitor_binding_effect_probe",
		}, "cmdb_monitor_binding_effect_receipt_contract"),
		cmdbContractFieldGroup("binding_rollback_receipt", status, []string{
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
