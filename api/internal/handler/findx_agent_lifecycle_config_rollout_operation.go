package handler

import (
	"strings"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

const (
	configRolloutPluginOperationAssign   = "assign"
	configRolloutPluginOperationDispatch = "dispatch"
	configRolloutPluginOperationConflict = "cmdb_agent_plugin_operation_identity_contract"
)

func configRolloutOperationContract(req model.FindXAgentConfigRolloutRequest, metadata map[string]string, missing []string) gin.H {
	if !isCMDBHostPluginRollout(req, metadata) {
		return gin.H{
			"mode":              configRolloutOperationMode(req, metadata),
			"contract":          "findx.agent.config_rollout.runtime.v1",
			"remote_mutation":   req.RemoteMutation,
			"required_receipts": []string{"writer_receipt", "reload_receipt", "restart_receipt", "drift_receipt", "rollback_receipt", "data_arrival_receipt", "evidence_chain"},
			"missing_contracts": configRolloutReceiptMissingContracts(missing),
			"status":            model.FindXAgentExecutionStateBlockedByContract,
		}
	}
	mode := configRolloutOperationMode(req, metadata)
	return gin.H{
		"mode":              mode,
		"contract":          configRolloutOperationContractID(mode),
		"remote_mutation":   mode == configRolloutPluginOperationDispatch,
		"required_receipts": configRolloutOperationRequiredReceipts(mode),
		"missing_contracts": configRolloutOperationMissingContracts(req, missing, mode),
		"status":            model.FindXAgentExecutionStateBlockedByContract,
	}
}

func configRolloutOperationMode(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	for _, value := range []string{
		req.TemplateID,
		req.RolloutStrategy,
	} {
		if mode, ok := configRolloutOperationFromValue(value); ok {
			return mode
		}
	}
	if req.RemoteMutation {
		return configRolloutPluginOperationDispatch
	}
	return configRolloutPluginOperationAssign
}

func configRolloutOperationFromValue(value string) (string, bool) {
	clean := strings.ToLower(strings.TrimSpace(value))
	switch {
	case clean == configRolloutPluginOperationAssign || strings.Contains(clean, "plugin-assign"):
		return configRolloutPluginOperationAssign, true
	case clean == configRolloutPluginOperationDispatch || strings.Contains(clean, "plugin-dispatch"):
		return configRolloutPluginOperationDispatch, true
	default:
		return "", false
	}
}

func configRolloutOperationIdentityConflict(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) bool {
	if strings.EqualFold(strings.TrimSpace(metadata["plugin_action_conflict"]), "blocked") {
		return true
	}
	expected := ""
	for _, value := range []string{
		req.TemplateID,
		req.RolloutStrategy,
		metadata["plugin_action"],
	} {
		mode, ok := configRolloutOperationFromValue(value)
		if !ok {
			continue
		}
		if expected == "" {
			expected = mode
			continue
		}
		if mode != expected {
			return true
		}
	}
	if expected == "" {
		return false
	}
	return (expected == configRolloutPluginOperationAssign && req.RemoteMutation) ||
		(expected == configRolloutPluginOperationDispatch && !req.RemoteMutation)
}

func configRolloutOperationContractID(mode string) string {
	if mode == configRolloutPluginOperationAssign {
		return "cmdb.agent.plugin.assignment.v1"
	}
	return "cmdb.agent.plugin.dispatch.v1"
}

func configRolloutOperationRequiredReceipts(mode string) []string {
	if mode == configRolloutPluginOperationAssign {
		return []string{"assignment_record", "target_binding_ref", "credential_policy_ref", "audit_ref"}
	}
	return []string{"assignment_ref", "writer_receipt", "delivery_receipt", "effect_receipt", "rollback_receipt", "data_arrival_receipt", "evidence_chain"}
}

func configRolloutOperationMissingContracts(req model.FindXAgentConfigRolloutRequest, missing []string, mode string) []string {
	required := cmdbHostPluginOperationMissingContractsFromExisting(req, missing, mode)
	for _, value := range missing {
		if !containsConfigRolloutRef(required, value) && cmdbHostPluginOperationMissingContractAllowed(value, mode) {
			required = append(required, value)
		}
	}
	return uniquePackageRepositoryBlockers(required)
}

func cmdbHostPluginOperationMissingContracts(req model.FindXAgentConfigRolloutRequest, mode string) []string {
	return cmdbHostPluginOperationMissingContractsFromExisting(req, nil, mode)
}

func cmdbHostPluginOperationMissingContractsFromExisting(req model.FindXAgentConfigRolloutRequest, existing []string, mode string) []string {
	missing := cmdbHostPluginBaseMissingContracts(existing)
	if strings.TrimSpace(req.CredentialRef) == "" {
		missing = append(missing, "credential_ref")
	}
	if mode == configRolloutPluginOperationAssign && req.RemoteMutation {
		missing = append(missing, configRolloutPluginOperationConflict)
	}
	if mode == configRolloutPluginOperationAssign {
		for _, value := range []string{
			cmdbAgentPluginAssignmentStoreContract,
			cmdbAgentPluginTargetBindingContract,
			cmdbAgentPluginAssignmentAuditContract,
		} {
			if len(existing) == 0 || containsConfigRolloutRef(existing, value) {
				missing = append(missing, value)
			}
		}
		return missing
	}
	if len(existing) == 0 || containsConfigRolloutRef(existing, cmdbAgentPluginAssignmentRefContract) {
		missing = append(missing, cmdbAgentPluginAssignmentRefContract)
	}
	return append(missing,
		"cmdb_agent_rollout_writer_receipt_contract",
		"cmdb_agent_rollout_delivery_receipt_contract",
		"cmdb_agent_rollout_effect_receipt_contract",
		"cmdb_agent_rollout_rollback_receipt_contract",
		"cmdb_agent_rollout_data_arrival_contract",
		"cmdb_agent_rollout_evidence_chain_contract",
	)
}

func cmdbHostPluginBaseMissingContracts(existing []string) []string {
	base := []string{
		"cmdb_agent_plugin_credential_contract",
		"cmdb_credential_ref_resolve_contract",
		"cmdb_plugin_config_schema_contract",
		"cmdb_dashboard_template_lookup_contract",
		"cmdb_dashboard_import_runtime_contract",
	}
	if existing == nil {
		return base
	}
	missing := make([]string, 0, len(base))
	for _, value := range base {
		if containsConfigRolloutRef(existing, value) {
			missing = append(missing, value)
		}
	}
	return missing
}

func cmdbHostPluginOperationMissingContractAllowed(value, mode string) bool {
	if mode == configRolloutPluginOperationAssign {
		return !strings.Contains(value, "rollout_delivery") &&
			!strings.Contains(value, "rollout_effect") &&
			!strings.Contains(value, "rollout_writer") &&
			!strings.Contains(value, "rollout_rollback") &&
			!strings.Contains(value, "rollout_data_arrival") &&
			!strings.Contains(value, "rollout_evidence_chain") &&
			!strings.Contains(value, "assignment_ref")
	}
	return !strings.Contains(value, "assignment_store") &&
		!strings.Contains(value, "target_binding") &&
		!strings.Contains(value, "assignment_audit")
}
