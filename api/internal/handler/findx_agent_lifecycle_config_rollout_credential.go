package handler

import (
	"encoding/json"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const (
	cmdbAgentPluginCredentialContract        = "cmdb_agent_plugin_credential_contract"
	cmdbCredentialRefResolveContract         = "cmdb_credential_ref_resolve_contract"
	cmdbCredentialScopePolicyContract        = "cmdb_credential_scope_policy_contract"
	cmdbAgentPluginCredentialScopeContract   = "cmdb_agent_plugin_credential_scope_contract"
	cmdbAgentPluginCredentialPolicyContract  = "cmdb_agent_plugin_credential_policy_contract"
	cmdbCredentialScopeAuthorizationContract = "cmdb_credential_scope_authorization_contract"
	cmdbCredentialDispatchPolicyGateContract = "cmdb_credential_dispatch_policy_gate_contract"
)

type cmdbPluginCredentialContext struct {
	Required               bool
	Provided               bool
	Resolved               bool
	Mode                   string
	Protocol               string
	PortPresent            bool
	ScopePolicyRequired    bool
	ScopePolicyResolved    bool
	ScopeAuthorized        bool
	PolicyApprovalRef      string
	HostRefPresent         bool
	AgentRefPresent        bool
	BusinessContextPresent bool
	PluginRefPresent       bool
	ProviderModePresent    bool
	PolicyReasonCode       string
}

func cmdbResolvePluginCredentialContext(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) cmdbPluginCredentialContext {
	if !isCMDBHostPluginRollout(req, metadata) || !cmdbPluginCatalogRequiresCredential(req.PluginID) {
		return cmdbPluginCredentialContext{}
	}
	ref := strings.TrimSpace(req.CredentialRef)
	ctx := cmdbPluginCredentialContext{
		Required:               true,
		Provided:               ref != "",
		Mode:                   "credential_ref",
		ScopePolicyRequired:    true,
		HostRefPresent:         cmdbPluginAssignmentHostRef(req, metadata) != "",
		AgentRefPresent:        cmdbPluginAssignmentAgentRef(req, metadata) != "",
		BusinessContextPresent: cmdbCredentialBusinessContextPresent(metadata),
		PluginRefPresent:       sanitizeRemoteMutationValue("plugin_id", req.PluginID) != "",
		ProviderModePresent:    sanitizeRemoteMutationValue("provider_mode", req.ProviderMode) != "",
	}
	if ref == "" {
		ctx.PolicyReasonCode = "credential_ref_required"
		return ctx
	}
	credential, ok := store.GetCredential(ref)
	if !ok || credential == nil || strings.TrimSpace(credential.ID) == "" {
		ctx.PolicyReasonCode = "credential_ref_unresolved"
		return ctx
	}
	ctx.Resolved = true
	ctx.Protocol = sanitizeRemoteMutationValue("protocol", credential.Protocol)
	ctx.PortPresent = credential.Port > 0
	ctx.PolicyReasonCode = "scope_policy_contract_missing"
	cmdbApplyCredentialPolicyApproval(req, metadata, &ctx)
	return ctx
}

func cmdbPluginCatalogRequiresCredential(pluginID string) bool {
	clean := normalizeFindXPluginCredentialID(pluginID)
	if clean == "" {
		return false
	}
	for _, plugin := range pluginCatalog {
		if normalizeFindXPluginCredentialID(plugin.ID) == clean {
			return plugin.CredentialRequired
		}
	}
	return findXPluginCredentialRequired(clean)
}

func normalizeFindXPluginCredentialID(pluginID string) string {
	clean := strings.ToLower(strings.TrimSpace(pluginID))
	clean = strings.TrimPrefix(clean, "inputs.")
	clean = strings.TrimPrefix(clean, "input.")
	return clean
}

func cmdbApplyCredentialResolveGate(missing []string, ctx cmdbPluginCredentialContext) []string {
	if !ctx.Required {
		return missing
	}
	out := make([]string, 0, len(missing)+2)
	for _, value := range missing {
		if ctx.Resolved && value == cmdbCredentialRefResolveContract {
			continue
		}
		out = append(out, value)
	}
	if !ctx.Provided {
		out = appendMissingConfigRolloutContract(out, "credential_ref")
	}
	out = appendMissingConfigRolloutContract(out, cmdbAgentPluginCredentialContract)
	if !ctx.Resolved {
		out = appendMissingConfigRolloutContract(out, cmdbCredentialRefResolveContract)
	}
	return uniquePackageRepositoryBlockers(out)
}

func cmdbApplyCredentialScopePolicyGate(missing []string, ctx cmdbPluginCredentialContext, mode string) []string {
	if !ctx.Required || !ctx.Resolved {
		return uniquePackageRepositoryBlockers(missing)
	}
	if ctx.ScopePolicyResolved && ctx.ScopeAuthorized {
		return cmdbFilterCredentialScopePolicyMissingContracts(missing, mode)
	}
	out := append([]string{}, missing...)
	for _, contract := range []string{
		cmdbCredentialScopePolicyContract,
		cmdbAgentPluginCredentialScopeContract,
		cmdbAgentPluginCredentialPolicyContract,
		cmdbCredentialScopeAuthorizationContract,
	} {
		out = appendMissingConfigRolloutContract(out, contract)
	}
	if mode == configRolloutPluginOperationDispatch {
		out = appendMissingConfigRolloutContract(out, cmdbCredentialDispatchPolicyGateContract)
	}
	return uniquePackageRepositoryBlockers(out)
}

func cmdbApplyCredentialPolicyApproval(req model.FindXAgentConfigRolloutRequest, metadata map[string]string, ctx *cmdbPluginCredentialContext) {
	if ctx == nil || !ctx.Required || !ctx.Resolved {
		return
	}
	approvalRef := cmdbCredentialPolicyApprovalRef(metadata)
	if approvalRef == "" || !cmdbCredentialPolicyApprovalMatches(approvalRef, req, metadata, *ctx) {
		return
	}
	ctx.ScopePolicyResolved = true
	ctx.ScopeAuthorized = true
	ctx.PolicyApprovalRef = approvalRef
	ctx.PolicyReasonCode = "credential_policy_release_recorded"
}

func cmdbCredentialPolicyApprovalRef(metadata map[string]string) string {
	for _, key := range []string{"scope_policy_approval_ref", "credential_policy_approval_ref", "credential_scope_policy_approval_ref"} {
		if ref := sanitizeRemoteMutationValue("approval_ref", metadata[key]); ref != "" {
			return ref
		}
	}
	return ""
}

func cmdbCredentialPolicyApprovalMatches(ref string, req model.FindXAgentConfigRolloutRequest, metadata map[string]string, ctx cmdbPluginCredentialContext) bool {
	approval, ok := store.GetCmdbResourceApproval(ref)
	if !ok || approval == nil {
		return false
	}
	if approval.Status != "review_accept_recorded" || approval.WorkflowState != "review_accept_recorded" {
		return false
	}
	if approval.ResourceType != "cmdb_agent_plugin_credential_policy" || approval.Action != "cmdb.agent.plugin.credential_policy.release" {
		return false
	}
	if approval.ResourceID != cmdbCredentialPolicyResourceID(req, metadata) {
		return false
	}
	context := map[string]string{}
	if err := json.Unmarshal([]byte(approval.ContextJSON), &context); err != nil {
		return false
	}
	expected := map[string]string{
		"host_ref":            cmdbPluginAssignmentHostRef(req, metadata),
		"agent_ref":           cmdbPluginAssignmentAgentRef(req, metadata),
		"plugin_id":           sanitizeRemoteMutationValue("plugin_id", req.PluginID),
		"provider_mode":       sanitizeRemoteMutationValue("provider_mode", req.ProviderMode),
		"operation_mode":      configRolloutOperationMode(req, metadata),
		"credential_protocol": ctx.Protocol,
	}
	for key, want := range expected {
		if want == "" || cmdbCredentialApprovalContextValue(key, context[key]) != want {
			return false
		}
	}
	for _, key := range []string{"business_group_ref", "team_ref"} {
		if want := cmdbCredentialApprovalContextValue(key, metadata[key]); want != "" && cmdbCredentialApprovalContextValue(key, context[key]) != want {
			return false
		}
	}
	return true
}

func cmdbCredentialApprovalContextValue(key, value string) string {
	if key == "credential_protocol" {
		return sanitizeRemoteMutationValue("protocol", value)
	}
	return sanitizeRemoteMutationValue(key, value)
}

func cmdbCredentialPolicyResourceID(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	parts := []string{
		cmdbPluginAssignmentHostRef(req, metadata),
		cmdbPluginAssignmentAgentRef(req, metadata),
		sanitizeRemoteMutationValue("plugin_id", req.PluginID),
	}
	for _, part := range parts {
		if part == "" {
			return ""
		}
	}
	return strings.Join(parts, ":")
}

func cmdbFilterCredentialScopePolicyMissingContracts(missing []string, mode string) []string {
	out := make([]string, 0, len(missing))
	for _, value := range missing {
		switch value {
		case cmdbCredentialScopePolicyContract,
			cmdbAgentPluginCredentialScopeContract,
			cmdbAgentPluginCredentialPolicyContract,
			cmdbCredentialScopeAuthorizationContract:
			continue
		case cmdbCredentialDispatchPolicyGateContract:
			if mode == configRolloutPluginOperationDispatch {
				continue
			}
		}
		out = append(out, value)
	}
	return uniquePackageRepositoryBlockers(out)
}

func cmdbCredentialBusinessContextPresent(metadata map[string]string) bool {
	for _, key := range []string{"business_group_ref", "business_group_id", "team_ref", "team_id", "tenant_id"} {
		if sanitizeRemoteMutationValue(key, metadata[key]) != "" {
			return true
		}
	}
	return false
}

func appendMissingConfigRolloutContract(values []string, contract string) []string {
	if contract == "" || containsConfigRolloutRef(values, contract) {
		return values
	}
	return append(values, contract)
}

func cmdbPluginCredentialContractResponse(ctx cmdbPluginCredentialContext) gin.H {
	if !ctx.Required {
		return nil
	}
	out := gin.H{
		"contract_id":              "cmdb.agent.plugin.credential_ref.read.v1",
		"mode":                     ctx.Mode,
		"credential_ref_required":  true,
		"credential_ref_present":   ctx.Provided,
		"credential_ref_resolved":  ctx.Resolved,
		"scope_policy_contract_id": "cmdb.credential.scope_policy.read.v1",
		"scope_policy_required":    ctx.ScopePolicyRequired,
		"scope_policy_resolved":    ctx.ScopePolicyResolved,
		"scope_authorized":         ctx.ScopeAuthorized,
		"host_ref_present":         ctx.HostRefPresent,
		"agent_ref_present":        ctx.AgentRefPresent,
		"business_context_present": ctx.BusinessContextPresent,
		"plugin_ref_present":       ctx.PluginRefPresent,
		"provider_mode_present":    ctx.ProviderModePresent,
		"policy_reason_code":       ctx.PolicyReasonCode,
		"status":                   model.FindXAgentExecutionStateBlockedByContract,
	}
	if ctx.Resolved {
		out["protocol"] = ctx.Protocol
		out["port_present"] = ctx.PortPresent
	}
	if ctx.ScopePolicyResolved && ctx.ScopeAuthorized && ctx.PolicyApprovalRef != "" {
		out["policy_approval_ref"] = ctx.PolicyApprovalRef
		out["policy_release_contract_id"] = "cmdb.agent.plugin.credential_policy.release.v1"
	}
	return out
}
