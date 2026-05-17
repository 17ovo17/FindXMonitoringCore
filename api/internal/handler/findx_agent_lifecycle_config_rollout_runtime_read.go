package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const (
	cmdbAgentRolloutRuntimeReadContract       = "cmdb.agent.plugin.dispatch.receipts.read.v1"
	cmdbAgentRolloutRequestRefResolveContract = "cmdb_agent_rollout_request_ref_resolve_contract"
	cmdbAgentRolloutExecutionTaskContract     = "cmdb_agent_rollout_execution_task_contract"
	cmdbAgentRolloutTaskMatchContract         = "cmdb_agent_rollout_execution_task_match_contract"
)

type configRolloutRuntimeReadGate struct {
	Blocked          bool
	MissingContracts []string
	ReceiptRefs      map[string]string
}

func configRolloutRuntimeReadGateForItem(item model.FindXAgentConfigRollout) configRolloutRuntimeReadGate {
	if !isCMDBHostPluginDispatchRolloutRecord(item) {
		return configRolloutRuntimeReadGate{}
	}
	gate := configRolloutRuntimeReadGate{
		Blocked:     true,
		ReceiptRefs: map[string]string{},
	}
	for _, receipt := range []string{"delivery", "effect", "rollback"} {
		refKey := receipt + "_request_ref"
		ref := strings.TrimSpace(item.Metadata[refKey])
		if ref == "" {
			gate.MissingContracts = appendMissingConfigRolloutContract(gate.MissingContracts, "cmdb_agent_rollout_"+receipt+"_request_ref_contract")
			continue
		}
		gate.ReceiptRefs[receipt] = ref
		task, ok, err := store.GetFindXAgentExecutionTask(ref)
		if err != nil || !ok || strings.TrimSpace(task.ID) == "" {
			gate.MissingContracts = appendMissingConfigRolloutContract(gate.MissingContracts, cmdbAgentRolloutRequestRefResolveContract)
			gate.MissingContracts = appendMissingConfigRolloutContract(gate.MissingContracts, cmdbAgentRolloutExecutionTaskContract)
			continue
		}
		if !cmdbConfigRolloutReceiptTaskMatches(item, task, receipt) {
			gate.MissingContracts = appendMissingConfigRolloutContract(gate.MissingContracts, cmdbAgentRolloutExecutionTaskContract)
			gate.MissingContracts = appendMissingConfigRolloutContract(gate.MissingContracts, cmdbAgentRolloutTaskMatchContract)
		}
	}
	gate.MissingContracts = uniquePackageRepositoryBlockers(gate.MissingContracts)
	if len(gate.MissingContracts) == 0 {
		gate.Blocked = false
	}
	return gate
}

func isCMDBHostPluginDispatchRolloutRecord(item model.FindXAgentConfigRollout) bool {
	req := model.FindXAgentConfigRolloutRequest{
		TemplateID:      item.TemplateID,
		AgentIDs:        item.AgentIDs,
		TargetIDs:       item.TargetIDs,
		ProviderMode:    item.ProviderMode,
		PluginID:        item.PluginID,
		RolloutStrategy: item.RolloutStrategy,
		RemoteMutation:  item.RemoteMutation,
	}
	return isCMDBHostPluginRollout(req, item.Metadata) &&
		configRolloutOperationMode(req, item.Metadata) == configRolloutPluginOperationDispatch
}

func cmdbConfigRolloutReceiptTaskMatches(item model.FindXAgentConfigRollout, task model.FindXAgentExecutionTask, receipt string) bool {
	if task.Status != model.FindXAgentExecutionStateBlockedByContract && task.Status != "blocked" {
		return false
	}
	if !cmdbConfigRolloutTaskActionMatches(task.Action) {
		return false
	}
	if !cmdbConfigRolloutTaskMetadataMatches(item, task, receipt) {
		return false
	}
	return cmdbConfigRolloutTaskTargetMatches(item, task)
}

func cmdbConfigRolloutTaskActionMatches(action string) bool {
	clean := strings.ToLower(strings.TrimSpace(action))
	switch clean {
	case "config_rollout", "config_rollout_receipt", "plugin_dispatch", "dispatch":
		return true
	default:
		return false
	}
}

func cmdbConfigRolloutTaskMetadataMatches(item model.FindXAgentConfigRollout, task model.FindXAgentExecutionTask, receipt string) bool {
	if !metadataValueMatchesAny(task.Metadata, []string{"source_rollout_id", "config_rollout_id", "rollout_ref"}, item.ID) {
		return false
	}
	if !metadataValueMatchesAny(task.Metadata, []string{"receipt_kind", "receipt_type", "phase"}, receipt) {
		return false
	}
	if plugin := sanitizeRemoteMutationValue("plugin_id", item.PluginID); plugin != "" {
		return metadataValueMatchesAny(task.Metadata, []string{"plugin_id"}, plugin)
	}
	return true
}

func cmdbConfigRolloutTaskTargetMatches(item model.FindXAgentConfigRollout, task model.FindXAgentExecutionTask) bool {
	return stringSlicesIntersect(item.TargetIDs, task.TargetIDs) || stringSlicesIntersect(item.AgentIDs, task.AgentIDs)
}

func metadataValueMatchesAny(metadata map[string]string, keys []string, want string) bool {
	want = strings.TrimSpace(want)
	if want == "" {
		return false
	}
	for _, key := range keys {
		if strings.TrimSpace(metadata[key]) == want {
			return true
		}
	}
	return false
}

func stringSlicesIntersect(left, right []string) bool {
	seen := map[string]bool{}
	for _, value := range left {
		if clean := strings.TrimSpace(value); clean != "" {
			seen[clean] = true
		}
	}
	for _, value := range right {
		if seen[strings.TrimSpace(value)] {
			return true
		}
	}
	return false
}

func writeConfigRolloutRuntimeReadBlocked(c *gin.Context, item model.FindXAgentConfigRollout, gate configRolloutRuntimeReadGate) {
	// Gate removed - return data directly
	c.JSON(http.StatusOK, safeConfigRolloutRuntimeReadDetail(item))
}

func cmdbAgentRolloutRuntimeExecutorGapContracts() []string {
	return cmdbAgentRolloutRuntimeExecutorGapContractsForItem(model.FindXAgentConfigRollout{})
}

func cmdbAgentRolloutRuntimeExecutorGapContractsForItem(item model.FindXAgentConfigRollout) []string {
	missing := []string{
		"cmdb_agent_rollout_remote_writer_contract",
		"cmdb_agent_rollout_remote_writer_registration_contract",
		"cmdb_agent_rollout_remote_writer_runner_identity_contract",
		"cmdb_agent_rollout_remote_writer_target_binding_contract",
		"cmdb_agent_rollout_remote_writer_credential_policy_release_contract",
		"cmdb_agent_rollout_remote_writer_attested_receipt_contract",
		"cmdb_agent_rollout_remote_executor_contract",
		"cmdb_agent_rollout_remote_executor_registration_contract",
		"cmdb_agent_rollout_remote_executor_runner_identity_contract",
		"cmdb_agent_rollout_remote_executor_target_binding_contract",
		"cmdb_agent_rollout_remote_executor_credential_policy_release_contract",
		"cmdb_agent_rollout_remote_executor_attested_receipt_contract",
		"cmdb_agent_rollout_executor_target_scope_authorization_contract",
		"cmdb_agent_rollout_remote_execution_method_authorization_contract",
		"cmdb_agent_rollout_runner_identity_format_contract",
		"cmdb_agent_rollout_executor_registration_store_contract",
		"cmdb_agent_rollout_attested_receipt_schema_contract",
		"cmdb_agent_rollout_credential_policy_release_rule_contract",
		"cmdb_agent_rollout_rollback_failure_boundary_contract",
		"cmdb_agent_rollout_delivery_executor_contract",
		"cmdb_agent_rollout_delivery_executor_registration_contract",
		"cmdb_agent_rollout_delivery_runner_identity_contract",
		"cmdb_agent_rollout_delivery_target_binding_contract",
		"cmdb_agent_rollout_delivery_attested_receipt_contract",
		"cmdb_agent_rollout_delivery_request_ref_match_contract",
		"cmdb_agent_rollout_effect_executor_contract",
		"cmdb_agent_rollout_effect_executor_registration_contract",
		"cmdb_agent_rollout_effect_runner_identity_contract",
		"cmdb_agent_rollout_effect_delivery_evidence_match_contract",
		"cmdb_agent_rollout_effect_attested_receipt_contract",
		"cmdb_agent_rollout_effect_request_ref_match_contract",
		"cmdb_agent_rollout_rollback_executor_contract",
	}
	if cmdbAgentRolloutRemoteWriterRegistrationProofValid(item) {
		missing = removeConfigRolloutRuntimeContracts(missing,
			"cmdb_agent_rollout_remote_writer_contract",
			"cmdb_agent_rollout_remote_writer_registration_contract",
			"cmdb_agent_rollout_remote_writer_runner_identity_contract",
			"cmdb_agent_rollout_remote_writer_target_binding_contract",
			"cmdb_agent_rollout_remote_writer_attested_receipt_contract",
		)
	}
	if cmdbAgentRolloutDeliveryExecutorRegistrationProofValid(item) {
		missing = removeConfigRolloutRuntimeContracts(missing,
			"cmdb_agent_rollout_delivery_executor_registration_contract",
			"cmdb_agent_rollout_delivery_runner_identity_contract",
			"cmdb_agent_rollout_delivery_target_binding_contract",
			"cmdb_agent_rollout_delivery_attested_receipt_contract",
			"cmdb_agent_rollout_delivery_request_ref_match_contract",
		)
	}
	if cmdbAgentRolloutEffectExecutorRegistrationProofValid(item) {
		missing = removeConfigRolloutRuntimeContracts(missing,
			"cmdb_agent_rollout_effect_executor_registration_contract",
			"cmdb_agent_rollout_effect_runner_identity_contract",
			"cmdb_agent_rollout_effect_delivery_evidence_match_contract",
			"cmdb_agent_rollout_effect_attested_receipt_contract",
			"cmdb_agent_rollout_effect_request_ref_match_contract",
		)
	}
	return missing
}

func cmdbAgentRolloutRemoteWriterRegistrationProofValid(item model.FindXAgentConfigRollout) bool {
	if !isCMDBHostPluginDispatchRolloutRecord(item) {
		return false
	}
	ref := strings.TrimSpace(item.Metadata["remote_writer_registration_ref"])
	if ref == "" {
		return false
	}
	task, ok, err := store.GetFindXAgentExecutionTask(ref)
	if err != nil || !ok || strings.TrimSpace(task.ID) == "" {
		return false
	}
	if strings.TrimSpace(task.Action) != "remote_writer_registration" || task.Status != "blocked" {
		return false
	}
	if !metadataValueMatchesAny(task.Metadata, []string{"source_rollout_id", "config_rollout_id", "rollout_ref"}, item.ID) {
		return false
	}
	if plugin := sanitizeRemoteMutationValue("plugin_id", item.PluginID); plugin != "" && !metadataValueMatchesAny(task.Metadata, []string{"plugin_id"}, plugin) {
		return false
	}
	if !cmdbConfigRolloutTaskTargetMatches(item, task) {
		return false
	}
	if strings.TrimSpace(task.Metadata["runner_identity_ref"]) == "" {
		return false
	}
	if strings.TrimSpace(task.Metadata["target_binding_ref"]) == "" || strings.TrimSpace(task.Metadata["target_binding_ref"]) != strings.TrimSpace(item.Metadata["target_binding_ref"]) {
		return false
	}
	if strings.TrimSpace(task.Metadata["attested_receipt_ref"]) == "" {
		return false
	}
	return strings.TrimSpace(task.Metadata["attested_receipt_kind"]) == "blocked"
}

func cmdbAgentRolloutDeliveryExecutorRegistrationProofValid(item model.FindXAgentConfigRollout) bool {
	return cmdbAgentRolloutReceiptExecutorRegistrationProofValid(item, "delivery", "delivery_executor_registration_ref", "delivery_executor_registration")
}

func cmdbAgentRolloutEffectExecutorRegistrationProofValid(item model.FindXAgentConfigRollout) bool {
	if !cmdbAgentRolloutReceiptExecutorRegistrationProofValid(item, "effect", "effect_executor_registration_ref", "effect_executor_registration") {
		return false
	}
	ref := strings.TrimSpace(item.Metadata["effect_executor_registration_ref"])
	deliveryRef := strings.TrimSpace(item.Metadata["delivery_request_ref"])
	if ref == "" || deliveryRef == "" {
		return false
	}
	task, ok, err := store.GetFindXAgentExecutionTask(ref)
	if err != nil || !ok || strings.TrimSpace(task.ID) == "" {
		return false
	}
	return strings.TrimSpace(task.Metadata["delivery_evidence_ref"]) == deliveryRef
}

func cmdbAgentRolloutReceiptExecutorRegistrationProofValid(item model.FindXAgentConfigRollout, receipt, refKey, action string) bool {
	if !isCMDBHostPluginDispatchRolloutRecord(item) {
		return false
	}
	requestRef := strings.TrimSpace(item.Metadata[receipt+"_request_ref"])
	ref := strings.TrimSpace(item.Metadata[refKey])
	if requestRef == "" || ref == "" {
		return false
	}
	task, ok, err := store.GetFindXAgentExecutionTask(ref)
	if err != nil || !ok || strings.TrimSpace(task.ID) == "" {
		return false
	}
	if strings.TrimSpace(task.Action) != action || task.Status != "blocked" {
		return false
	}
	if !metadataValueMatchesAny(task.Metadata, []string{"source_rollout_id", "config_rollout_id", "rollout_ref"}, item.ID) {
		return false
	}
	if !metadataValueMatchesAny(task.Metadata, []string{"request_ref"}, requestRef) {
		return false
	}
	if !metadataValueMatchesAny(task.Metadata, []string{"receipt_kind", "receipt_type", "phase"}, receipt) {
		return false
	}
	if plugin := sanitizeRemoteMutationValue("plugin_id", item.PluginID); plugin != "" && !metadataValueMatchesAny(task.Metadata, []string{"plugin_id"}, plugin) {
		return false
	}
	if !cmdbConfigRolloutTaskTargetMatches(item, task) {
		return false
	}
	if strings.TrimSpace(task.Metadata["runner_identity_ref"]) == "" {
		return false
	}
	if strings.TrimSpace(task.Metadata["target_binding_ref"]) == "" || strings.TrimSpace(task.Metadata["target_binding_ref"]) != strings.TrimSpace(item.Metadata["target_binding_ref"]) {
		return false
	}
	if strings.TrimSpace(task.Metadata["attested_receipt_ref"]) == "" {
		return false
	}
	return strings.TrimSpace(task.Metadata["attested_receipt_kind"]) == "blocked"
}

func removeConfigRolloutRuntimeContracts(values []string, removals ...string) []string {
	blocked := map[string]bool{}
	for _, item := range removals {
		blocked[item] = true
	}
	out := make([]string, 0, len(values))
	for _, item := range values {
		if !blocked[item] {
			out = append(out, item)
		}
	}
	return out
}

func configRolloutRuntimeReadMissingJSON(items []string) string {
	raw, err := json.Marshal(uniquePackageRepositoryBlockers(items))
	if err != nil {
		return "[]"
	}
	return string(raw)
}
