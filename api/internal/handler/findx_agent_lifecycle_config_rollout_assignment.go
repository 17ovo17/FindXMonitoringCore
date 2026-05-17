package handler

import (
	"encoding/json"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const (
	cmdbAgentPluginAssignmentStoreContract = "cmdb_agent_plugin_assignment_store_contract"
	cmdbAgentPluginTargetBindingContract   = "cmdb_agent_plugin_target_binding_contract"
	cmdbAgentPluginAssignmentAuditContract = "cmdb_agent_plugin_assignment_audit_contract"
	cmdbAgentPluginAssignmentRefContract   = "cmdb_agent_plugin_assignment_ref_contract"
)

type cmdbPluginAssignmentContext struct {
	Assignment      model.FindXAgentPluginAssignment
	TargetBinding   model.FindXAgentPluginTargetBinding
	AssignmentReady bool
	BindingReady    bool
	AuditReady      bool
}

func cmdbResolvePluginAssignmentContext(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) cmdbPluginAssignmentContext {
	if !isCMDBHostPluginRollout(req, metadata) {
		return cmdbPluginAssignmentContext{}
	}
	if configRolloutOperationMode(req, metadata) != configRolloutPluginOperationDispatch {
		return cmdbPluginAssignmentContext{}
	}
	assignment, ok := cmdbFindPluginAssignment(req, metadata)
	if !ok {
		return cmdbPluginAssignmentContext{}
	}
	binding, bindingOK := cmdbFindPluginTargetBinding(assignment)
	return cmdbPluginAssignmentContext{
		Assignment:      assignment,
		TargetBinding:   binding,
		AssignmentReady: true,
		BindingReady:    bindingOK,
		AuditReady:      strings.TrimSpace(assignment.AuditRef) != "",
	}
}

func cmdbPersistPluginAssignment(c *gin.Context, req model.FindXAgentConfigRolloutRequest, metadata map[string]string, saved model.FindXAgentConfigRollout, missing []string) (cmdbPluginAssignmentContext, error) {
	if !isCMDBHostPluginRollout(req, metadata) || configRolloutOperationMode(req, metadata) != configRolloutPluginOperationAssign {
		return cmdbPluginAssignmentContext{}, nil
	}
	hostRef := cmdbPluginAssignmentHostRef(req, metadata)
	agentRef := cmdbPluginAssignmentAgentRef(req, metadata)
	pluginID := sanitizeRemoteMutationValue("plugin_id", req.PluginID)
	if hostRef == "" || agentRef == "" || pluginID == "" {
		return cmdbPluginAssignmentContext{}, nil
	}
	auditRef := "cmdb-agent-plugin-assignment-audit-" + store.NewID()
	assignment := model.FindXAgentPluginAssignment{
		SourceRolloutID:      saved.ID,
		HostRef:              hostRef,
		AgentRef:             agentRef,
		PluginID:             pluginID,
		PluginVersion:        sanitizeRemoteMutationValue("plugin_version", req.PluginVersion),
		ConfigSnippetRef:     sanitizeRemoteMutationValue("config_snippet_ref", req.ConfigSnippetRef),
		ConfigFormat:         sanitizeRemoteMutationValue("config_format", req.ConfigFormat),
		ProviderMode:         sanitizeRemoteMutationValue("provider_mode", req.ProviderMode),
		AuditRef:             auditRef,
		AssignmentContract:   "cmdb.agent.plugin.assignment.v1",
		Status:               "accepted",
		Blocker:              "",
		CredentialRefPresent: strings.TrimSpace(req.CredentialRef) != "",
		DashboardRefsJSON:    lifecycleJSONString(cmdbPluginAssignmentDashboardRefs(metadata)),
		DashboardRefsCount:   len(cmdbPluginAssignmentDashboardRefs(metadata)),
		MissingContractsJSON: lifecycleJSONString(cmdbFilterMissingContractsForAssignment(missing, true)),
		MetadataJSON:         lifecycleJSONString(metadata),
	}
	savedAssignment, err := store.SaveFindXAgentPluginAssignment(assignment)
	if err != nil {
		return cmdbPluginAssignmentContext{}, err
	}
	binding := model.FindXAgentPluginTargetBinding{
		AssignmentID:         savedAssignment.ID,
		SourceRolloutID:      saved.ID,
		HostRef:              hostRef,
		TargetID:             cmdbPluginAssignmentTargetID(req, metadata),
		AgentRef:             agentRef,
		PluginID:             pluginID,
		BindingType:          "cmdb_host_plugin",
		Status:               "accepted",
		Blocker:              "",
		CredentialRefPresent: strings.TrimSpace(req.CredentialRef) != "",
		DashboardRefsCount:   len(cmdbPluginAssignmentDashboardRefs(metadata)),
		ContractID:           "cmdb.agent.plugin.target_binding.v1",
		AuditRef:             auditRef,
	}
	savedBinding, err := store.SaveFindXAgentPluginTargetBinding(binding)
	if err != nil {
		return cmdbPluginAssignmentContext{}, err
	}
	savedAssignment.TargetBindingRef = savedBinding.ID
	savedAssignment, err = store.SaveFindXAgentPluginAssignment(savedAssignment)
	if err != nil {
		return cmdbPluginAssignmentContext{}, err
	}
	if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
		ID:           auditRef,
		Actor:        requestActor(c),
		Action:       "cmdb.agent.plugin.assignment.save",
		ResourceType: "cmdb_agent_plugin_assignment",
		ResourceID:   savedAssignment.ID,
		Scope:        "cmdb",
		Status:       "accepted",
		TraceID:      c.GetHeader("X-Test-Batch-Id"),
		ClientIP:     c.ClientIP(),
		Summary:      "CMDB plugin assignment saved",
		Details: map[string]any{
			"assignment_id":          savedAssignment.ID,
			"target_binding_id":      savedBinding.ID,
			"host_ref":               hostRef,
			"agent_ref":              agentRef,
			"plugin_id":              pluginID,
			"credential_ref_present": savedAssignment.CredentialRefPresent,
			"dashboard_refs_count":   savedAssignment.DashboardRefsCount,
			"source_rollout_id":      saved.ID,
		},
	}); err != nil {
		return cmdbPluginAssignmentContext{}, err
	}
	return cmdbPluginAssignmentContext{
		Assignment:      savedAssignment,
		TargetBinding:   savedBinding,
		AssignmentReady: true,
		BindingReady:    true,
		AuditReady:      true,
	}, nil
}

func cmdbFindPluginAssignment(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) (model.FindXAgentPluginAssignment, bool) {
	if ref := strings.TrimSpace(metadata["assignment_ref"]); ref != "" {
		assignment, ok, err := store.GetFindXAgentPluginAssignment(ref)
		if err != nil || !ok || !cmdbPluginAssignmentMatchesRequest(assignment, req, metadata) {
			return model.FindXAgentPluginAssignment{}, false
		}
		return assignment, true
	}
	assignment, ok, err := store.FindFindXAgentPluginAssignment(
		cmdbPluginAssignmentHostRef(req, metadata),
		cmdbPluginAssignmentAgentRef(req, metadata),
		sanitizeRemoteMutationValue("plugin_id", req.PluginID),
	)
	if err != nil || !ok {
		return model.FindXAgentPluginAssignment{}, false
	}
	return assignment, true
}

func cmdbFindPluginTargetBinding(assignment model.FindXAgentPluginAssignment) (model.FindXAgentPluginTargetBinding, bool) {
	if strings.TrimSpace(assignment.TargetBindingRef) != "" {
		binding, ok, err := store.GetFindXAgentPluginTargetBinding(assignment.TargetBindingRef)
		if err == nil && ok {
			return binding, true
		}
	}
	bindings, err := store.ListFindXAgentPluginTargetBindings(assignment.ID)
	if err != nil || len(bindings) == 0 {
		return model.FindXAgentPluginTargetBinding{}, false
	}
	return bindings[0], true
}

func cmdbPluginAssignmentMatchesRequest(assignment model.FindXAgentPluginAssignment, req model.FindXAgentConfigRolloutRequest, metadata map[string]string) bool {
	return assignment.HostRef == cmdbPluginAssignmentHostRef(req, metadata) &&
		assignment.AgentRef == cmdbPluginAssignmentAgentRef(req, metadata) &&
		assignment.PluginID == sanitizeRemoteMutationValue("plugin_id", req.PluginID)
}

func cmdbPluginAssignmentHostRef(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	for _, value := range []string{metadata["cmdb_host_ref"], metadata["cmdb_host_id"]} {
		if clean := sanitizeRemoteMutationValue("cmdb_host_ref", value); clean != "" {
			return clean
		}
	}
	targets := cleanAgentLifecycleValues(req.TargetIDs)
	if len(targets) > 0 {
		return sanitizeRemoteMutationValue("target_id", targets[0])
	}
	return ""
}

func cmdbPluginAssignmentAgentRef(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	if clean := sanitizeRemoteMutationValue("agent_ref", metadata["agent_ref"]); clean != "" {
		return clean
	}
	agents := cleanAgentLifecycleValues(req.AgentIDs)
	if len(agents) > 0 {
		return sanitizeRemoteMutationValue("agent_id", agents[0])
	}
	return ""
}

func cmdbPluginAssignmentTargetID(req model.FindXAgentConfigRolloutRequest, metadata map[string]string) string {
	if clean := sanitizeRemoteMutationValue("target_id", metadata["target_id"]); clean != "" {
		return clean
	}
	return cmdbPluginAssignmentHostRef(req, metadata)
}

func cmdbPluginAssignmentDashboardRefs(metadata map[string]string) []string {
	raw := strings.TrimSpace(metadata["dashboard_refs"])
	if raw == "" {
		return nil
	}
	values := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '/'
	})
	out := make([]string, 0, len(values))
	for _, value := range values {
		if clean := sanitizeRemoteMutationValue("dashboard_ref", value); clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

func cmdbFilterMissingContractsForAssignment(missing []string, removeAssignmentContracts bool) []string {
	out := make([]string, 0, len(missing))
	for _, value := range missing {
		if removeAssignmentContracts && (value == cmdbAgentPluginAssignmentStoreContract ||
			value == cmdbAgentPluginTargetBindingContract ||
			value == cmdbAgentPluginAssignmentAuditContract) {
			continue
		}
		out = append(out, value)
	}
	return uniquePackageRepositoryBlockers(out)
}

func cmdbFilterMissingContractsForDispatch(missing []string, ctx cmdbPluginAssignmentContext) []string {
	out := make([]string, 0, len(missing))
	for _, value := range missing {
		if ctx.AssignmentReady && value == cmdbAgentPluginAssignmentRefContract {
			continue
		}
		out = append(out, value)
	}
	return uniquePackageRepositoryBlockers(out)
}

func cmdbPluginAssignmentResponse(ctx cmdbPluginAssignmentContext) gin.H {
	if !ctx.AssignmentReady {
		return nil
	}
	return gin.H{
		"assignment_ref":          ctx.Assignment.ID,
		"target_binding_ref":      ctx.TargetBinding.ID,
		"contract_id":             "cmdb.agent.plugin.assignment.store.read.v1",
		"status":                  "accepted",
		"credential_ref_present":  ctx.Assignment.CredentialRefPresent,
		"dashboard_refs_count":    ctx.Assignment.DashboardRefsCount,
		"audit_ref":               ctx.Assignment.AuditRef,
		"source_rollout_id":       ctx.Assignment.SourceRolloutID,
		"remote_mutation":         false,
		"dispatch_receipts_state": "accepted",
		"findx_audit_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_agent_plugin_assignment",
			"resource_id":   ctx.Assignment.ID,
			"action":        "cmdb.agent.plugin.assignment.save",
		},
	}
}

func cmdbOperationContractWithAssignment(base gin.H, ctx cmdbPluginAssignmentContext) gin.H {
	if !ctx.AssignmentReady {
		return base
	}
	base["assignment_ref"] = ctx.Assignment.ID
	base["target_binding_ref"] = ctx.TargetBinding.ID
	base["assignment_contract"] = cmdbPluginAssignmentResponse(ctx)
	return base
}

func cmdbOperationContractWithAssignmentAndCredential(base gin.H, assignmentCtx cmdbPluginAssignmentContext, credentialCtx cmdbPluginCredentialContext) gin.H {
	base = cmdbOperationContractWithAssignment(base, assignmentCtx)
	if credentialContract := cmdbPluginCredentialContractResponse(credentialCtx); credentialContract != nil {
		base["credential_contract"] = credentialContract
		if credentialCtx.Resolved {
			base["credential_ref_resolved"] = true
		}
		if credentialCtx.ScopePolicyResolved && credentialCtx.ScopeAuthorized {
			base["scope_policy_resolved"] = true
			base["scope_authorized"] = true
			base["policy_approval_ref"] = credentialCtx.PolicyApprovalRef
		}
	}
	return base
}

func lifecycleJSONString(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}
