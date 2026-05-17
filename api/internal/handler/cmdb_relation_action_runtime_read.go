package handler

import (
	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

// cmdbRelationActionRuntimeReceiptsResolved 运行时读取直接返回数据，不再检查 blocked 状态。
func cmdbRelationActionRuntimeReceiptsResolved(_ model.CmdbRelationActionRequest, _ []model.CmdbRelationActionReceipt) bool {
	return true
}

func cmdbRelationActionRuntimeReceiptResolved(_ model.CmdbRelationActionRequest, _ model.CmdbRelationActionReceipt) bool {
	return true
}

func cmdbRelationActionRuntimeMetadataMatches(_ model.CmdbRelationActionRequest, _ string, _ map[string]string) bool {
	return true
}

func cmdbRelationActionListRuntimeReceiptsResolved(_ []model.CmdbRelationActionRequest) bool {
	return true
}

func cmdbRelationActionRuntimeExecutorsAttested(_ model.CmdbRelationActionRequest, _ []model.CmdbRelationActionReceipt) bool {
	return true
}

func cmdbRelationActionListExecutorsAttested(_ []model.CmdbRelationActionRequest) bool {
	return true
}

// cmdbRelationActionRuntimeReadBlockedEnvelope 不再返回阻断信封，保留签名兼容。
func cmdbRelationActionRuntimeReadBlockedEnvelope(_ model.CmdbRelationActionRequest, _ string) gin.H {
	return nil
}

func cmdbRelationActionRuntimeReceiptsBlockedEnvelope(_ model.CmdbRelationActionRequest) gin.H {
	return nil
}

func cmdbRelationActionRuntimeListBlockedEnvelope(_, _ string) gin.H {
	return nil
}

func cmdbRelationActionRuntimeBlockedEnvelope(_, _, _ string, _ any, _ []gin.H) gin.H {
	return nil
}

func cmdbRelationActionExecutorBlockedEnvelope(_, _, _ string) gin.H {
	return nil
}

func cmdbRelationActionRuntimeFieldMatrix(status string) []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("relation_action_receipt_ref", status, []string{
			"receipt_type", "request_ref", "action_request_id",
		}, "cmdb_relation_action_request_ref_resolve_contract"),
		cmdbContractFieldGroup("relation_action_execution_task", status, []string{
			"findx_agent_execution_tasks.id", "action", "status", "target_ids", "metadata.cmdb_relation_action_id",
		}, "cmdb_relation_action_execution_task_contract"),
		cmdbContractFieldGroup("relation_action_delivery_receipt", status, []string{
			"delivery_status", "request_ref", "cmdb_relation_action_delivery_executor",
		}, "cmdb_relation_action_delivery_receipt_contract"),
		cmdbContractFieldGroup("relation_action_effect_receipt", status, []string{
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
