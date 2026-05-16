package handler

import (
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func cmdbAttachConfigRolloutReceiptRequestRefs(saved model.FindXAgentConfigRollout, req model.FindXAgentConfigRolloutRequest, metadata map[string]string, actor string) (model.FindXAgentConfigRollout, error) {
	if !isCMDBHostPluginRollout(req, metadata) || configRolloutOperationMode(req, metadata) != configRolloutPluginOperationDispatch {
		return saved, nil
	}
	if saved.Metadata == nil {
		saved.Metadata = map[string]string{}
	}
	for _, receipt := range []string{"delivery", "effect", "rollback"} {
		refKey := receipt + "_request_ref"
		if strings.TrimSpace(saved.Metadata[refKey]) != "" {
			continue
		}
		task, err := cmdbCreateConfigRolloutReceiptRequest(saved, req, metadata, receipt, actor)
		if err != nil {
			return saved, err
		}
		saved.Metadata[refKey] = task.ID
	}
	return store.SaveFindXAgentConfigRollout(saved)
}

func cmdbCreateConfigRolloutReceiptRequest(saved model.FindXAgentConfigRollout, req model.FindXAgentConfigRolloutRequest, metadata map[string]string, receipt, actor string) (model.FindXAgentExecutionTask, error) {
	taskMetadata := cmdbConfigRolloutReceiptRequestMetadata(saved, req, metadata, receipt, actor)
	task := model.FindXAgentExecutionTask{
		Action:               "config_rollout_receipt",
		AgentIDs:             cleanAgentLifecycleValues(saved.AgentIDs),
		TargetIDs:            cleanAgentLifecycleValues(saved.TargetIDs),
		ConfigVersion:        sanitizeRemoteMutationValue("config_version", saved.ConfigVersion),
		Status:               "blocked",
		Blocker:              "PENDING: cmdb plugin dispatch " + receipt + " executor is not enabled",
		Audit:                "findx_agent.task.requested",
		CredentialRefPresent: strings.TrimSpace(req.CredentialRef) != "",
		Metadata:             taskMetadata,
	}
	return store.SaveFindXAgentExecutionTask(task)
}

func cmdbConfigRolloutReceiptRequestMetadata(saved model.FindXAgentConfigRollout, req model.FindXAgentConfigRolloutRequest, metadata map[string]string, receipt, actor string) map[string]string {
	out := map[string]string{
		"scope":              configRolloutScopeCMDBHost,
		"source_rollout_id":  saved.ID,
		"config_rollout_id":  saved.ID,
		"rollout_ref":        saved.ID,
		"receipt_kind":       receipt,
		"receipt_type":       receipt,
		"phase":              receipt,
		"plugin_id":          sanitizeRemoteMutationValue("plugin_id", saved.PluginID),
		"provider_mode":      sanitizeRemoteMutationValue("provider_mode", saved.ProviderMode),
		"config_version":     sanitizeRemoteMutationValue("config_version", saved.ConfigVersion),
		"config_snippet_ref": sanitizeRemoteMutationValue("config_snippet_ref", saved.ConfigSnippetRef),
		"rollback_ref":       sanitizeRemoteMutationValue("rollback_ref", saved.RollbackRef),
		"idempotency_key":    sanitizeRemoteMutationValue("idempotency_key", "cmdb-agent-rollout-"+receipt+"-"+saved.ID),
		"requested_by":       sanitizeRemoteMutationValue("requested_by", actor),
	}
	for dst, src := range map[string]string{
		"cmdb_host_ref":      firstNonEmpty(metadata["cmdb_host_ref"], cmdbPluginAssignmentHostRef(req, metadata)),
		"agent_ref":          firstNonEmpty(metadata["agent_ref"], cmdbPluginAssignmentAgentRef(req, metadata)),
		"assignment_ref":     metadata["assignment_ref"],
		"target_binding_ref": metadata["target_binding_ref"],
	} {
		if clean := sanitizeRemoteMutationValue(dst, src); clean != "" {
			out[dst] = clean
		}
	}
	if receipt == "delivery" {
		cmdbAttachDeliveryExecutorBoundaryMetadata(out)
	}
	if receipt == "effect" {
		cmdbAttachEffectExecutorBoundaryMetadata(out)
	}
	if receipt == "rollback" {
		cmdbAttachRollbackExecutorBoundaryMetadata(out)
	}
	return out
}

func cmdbAttachDeliveryExecutorBoundaryMetadata(metadata map[string]string) {
	metadata["delivery_executor_contract_status"] = "blocked"
	metadata["delivery_executor_registration_contract"] = "cmdb_agent_rollout_delivery_executor_registration_contract"
	metadata["delivery_runner_identity_contract"] = "cmdb_agent_rollout_delivery_runner_identity_contract"
	metadata["delivery_attested_receipt_contract"] = "cmdb_agent_rollout_delivery_attested_receipt_contract"
	metadata["delivery_target_binding_contract"] = "cmdb_agent_rollout_delivery_target_binding_contract"
	metadata["delivery_request_ref_match_contract"] = "cmdb_agent_rollout_delivery_request_ref_match_contract"
	metadata["delivery_executor_missing_contracts"] = configRolloutReceiptIngestionMissingJSON([]string{
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_delivery_executor_registration_contract",
		"cmdb_agent_rollout_delivery_runner_identity_contract",
		"cmdb_agent_rollout_delivery_attested_receipt_contract",
		"cmdb_agent_rollout_delivery_target_binding_contract",
		"cmdb_agent_rollout_delivery_request_ref_match_contract",
	})
}

func cmdbAttachEffectExecutorBoundaryMetadata(metadata map[string]string) {
	metadata["effect_executor_contract_status"] = "blocked"
	metadata["effect_executor_registration_contract"] = "cmdb_agent_rollout_effect_executor_registration_contract"
	metadata["effect_runner_identity_contract"] = "cmdb_agent_rollout_effect_runner_identity_contract"
	metadata["effect_delivery_evidence_contract"] = "cmdb_agent_rollout_effect_delivery_evidence_match_contract"
	metadata["effect_attested_receipt_contract"] = "cmdb_agent_rollout_effect_attested_receipt_contract"
	metadata["effect_request_ref_match_contract"] = "cmdb_agent_rollout_effect_request_ref_match_contract"
	metadata["effect_executor_missing_contracts"] = configRolloutReceiptIngestionMissingJSON([]string{
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_effect_executor_registration_contract",
		"cmdb_agent_rollout_effect_runner_identity_contract",
		"cmdb_agent_rollout_effect_delivery_evidence_match_contract",
		"cmdb_agent_rollout_effect_attested_receipt_contract",
		"cmdb_agent_rollout_effect_request_ref_match_contract",
	})
}

func cmdbAttachRollbackExecutorBoundaryMetadata(metadata map[string]string) {
	metadata["rollback_executor_contract_status"] = "blocked"
	metadata["rollback_executor_registration_contract"] = "cmdb_agent_rollout_rollback_executor_registration_contract"
	metadata["rollback_runner_identity_contract"] = "cmdb_agent_rollout_rollback_runner_identity_contract"
	metadata["rollback_operation_context_contract"] = "cmdb_agent_rollout_rollback_operation_context_contract"
	metadata["rollback_attested_receipt_contract"] = "cmdb_agent_rollout_rollback_attested_receipt_contract"
	metadata["rollback_request_ref_match_contract"] = "cmdb_agent_rollout_rollback_request_ref_match_contract"
	metadata["rollback_executor_missing_contracts"] = configRolloutReceiptIngestionMissingJSON([]string{
		"cmdb_agent_rollout_rollback_executor_contract",
		"cmdb_agent_rollout_rollback_executor_registration_contract",
		"cmdb_agent_rollout_rollback_runner_identity_contract",
		"cmdb_agent_rollout_rollback_operation_context_contract",
		"cmdb_agent_rollout_rollback_attested_receipt_contract",
		"cmdb_agent_rollout_rollback_request_ref_match_contract",
	})
}
