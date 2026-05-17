package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const cmdbResourceApprovalRuntimeContract = "cmdb.resource.approval.runtime.v1"

var cmdbHighRiskCommonMissingContracts = []string{
	"cmdb_resource_approval_runtime_contract",
	"cmdb_operation_risk_policy_contract",
	"cmdb_action_preflight_contract",
	"cmdb_action_audit_receipt_contract",
}

func GetCmdbResourceApprovals(c *gin.Context) {
	view := cmdbApprovalView(c.Query("view"))
	actor := requestActor(c)
	items := store.ListCmdbResourceApprovals(view, actor)
	c.JSON(http.StatusOK, cmdbResourceApprovalListEnvelope(view, actor, items))
}

func GetCmdbResourceApproval(c *gin.Context) {
	item, ok := store.GetCmdbResourceApproval(strings.TrimSpace(c.Param("id")))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb approval request not found"})
		return
	}
	c.JSON(http.StatusOK, cmdbResourceApprovalDetailEnvelope(*item))
}

func ReviewCmdbResourceApproval(c *gin.Context) {
	var payload struct {
		Decision string `json:"decision" binding:"required"`
		Note     string `json:"note"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval review payload"})
		return
	}
	item, ok, err := store.DecideCmdbResourceApproval(strings.TrimSpace(c.Param("id")), payload.Decision, requestActor(c), cmdbSafeAuditText(payload.Note))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cmdb approval decision is not valid"})
		return
	}
	if !ok || item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb approval request not found"})
		return
	}
	if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
		ID:           "cmdb-approval-review-audit-" + store.NewID(),
		Actor:        requestActor(c),
		Action:       "cmdb.resource_approval.review",
		ResourceType: "cmdb_resource_approval",
		ResourceID:   item.ID,
		Scope:        "cmdb",
		Status:       "ok",
		ClientIP:     c.ClientIP(),
		Summary:      "CMDB approval review completed",
		Details: map[string]any{
			"approval_id":    item.ID,
			"decision_state": item.Status,
			"resource_type":  item.ResourceType,
			"resource_id":    item.ResourceID,
			"action":         item.Action,
			"note":           item.DecisionNote,
		},
	}); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb approval review audit unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":             0,
		"status":           "ok",
		"approval_request": cmdbResourceApprovalDTO(*item),
	})
}

func cmdbResourceApprovalBlockedEnvelope(view string) gin.H {
	envelope := cmdbBlockedContractEnvelope(
		cmdbResourceApprovalRuntimeContract,
		[]string{
			"cmdb_resource_approval_store_contract",
			"cmdb_resource_approval_workflow_contract",
			"cmdb_operation_risk_policy_contract",
			"cmdb_approval_audit_receipt_contract",
		},
		gin.H{
			"view": view,
			"findx_audit_query": gin.H{
				"source":        "findx_audit",
				"scope":         "cmdb",
				"resource_type": "cmdb_resource_approval",
				"action":        "cmdb.resource_approval.read",
				"view":          view,
			},
		},
	)
	envelope["status"] = cmdbBlockedByContract
	return envelope
}

func cmdbHighRiskMissingContracts(extra ...string) []string {
	missing := make([]string, 0, len(cmdbHighRiskCommonMissingContracts)+len(extra))
	missing = append(missing, cmdbHighRiskCommonMissingContracts...)
	missing = append(missing, extra...)
	return missing
}

type cmdbHighRiskApprovalInput struct {
	ContractID     string
	ResourceType   string
	ResourceID     string
	Action         string
	RiskLevel      string
	Title          string
	Summary        string
	BusinessGroup  string
	Reason         string
	PolicyID       string
	Context        map[string]any
	Diff           map[string]any
	Missing        []string
	ExecutionState string
}

type cmdbResourceApprovalRequestPayload struct {
	ResourceType       string `json:"resource_type"`
	ResourceID         string `json:"resource_id"`
	Action             string `json:"action"`
	RiskLevel          string `json:"risk_level"`
	Title              string `json:"title"`
	Summary            string `json:"summary"`
	BusinessGroup      string `json:"business_group"`
	Reason             string `json:"reason"`
	PolicyID           string `json:"policy_id"`
	HostRef            string `json:"host_ref"`
	AgentRef           string `json:"agent_ref"`
	PluginID           string `json:"plugin_id"`
	ProviderMode       string `json:"provider_mode"`
	OperationMode      string `json:"operation_mode"`
	CredentialProtocol string `json:"credential_protocol"`
	BusinessGroupRef   string `json:"business_group_ref"`
	TeamRef            string `json:"team_ref"`
}

func CreateCmdbResourceApprovalRequest(c *gin.Context) {
	var payload cmdbResourceApprovalRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusConflict, cmdbResourceApprovalRequestBlockedEnvelope("cmdb approval request requires a JSON payload", []string{
			"cmdb_resource_approval_request_payload_contract",
		}))
		return
	}
	input, missing := cmdbCredentialPolicyApprovalInputFromPayload(payload)
	if len(missing) > 0 {
		c.JSON(http.StatusConflict, cmdbResourceApprovalRequestBlockedEnvelope("credential policy approval request context is incomplete or unsupported", missing))
		return
	}
	gate := cmdbHighRiskApprovalGate(c, input)
	gate["code"] = 0
	gate["status"] = "recorded"
	gate["message"] = "CMDB credential policy approval request recorded for review."
	gate["safe_to_retry"] = true
	c.JSON(http.StatusCreated, gate)
}

func cmdbCredentialPolicyApprovalInputFromPayload(payload cmdbResourceApprovalRequestPayload) (cmdbHighRiskApprovalInput, []string) {
	context := map[string]any{}
	missing := []string{}
	resourceType := cmdbCredentialPolicyApprovalEnum(payload.ResourceType, "cmdb_agent_plugin_credential_policy")
	action := cmdbCredentialPolicyApprovalEnum(payload.Action, "cmdb.agent.plugin.credential_policy.release")
	hostRef := sanitizeRemoteMutationValue("host_ref", payload.HostRef)
	agentRef := sanitizeRemoteMutationValue("agent_ref", payload.AgentRef)
	pluginID := sanitizeRemoteMutationValue("plugin_id", payload.PluginID)
	providerMode := sanitizeRemoteMutationValue("provider_mode", payload.ProviderMode)
	operationMode := cmdbCredentialPolicyApprovalOperationMode(payload.OperationMode)
	credentialProtocol := sanitizeRemoteMutationValue("protocol", payload.CredentialProtocol)
	if resourceType != "cmdb_agent_plugin_credential_policy" {
		missing = append(missing, "cmdb_credential_policy_resource_type_contract")
	}
	if action != "cmdb.agent.plugin.credential_policy.release" {
		missing = append(missing, "cmdb_credential_policy_release_action_contract")
	}
	for key, value := range map[string]string{
		"host_ref":            hostRef,
		"agent_ref":           agentRef,
		"plugin_id":           pluginID,
		"provider_mode":       providerMode,
		"operation_mode":      operationMode,
		"credential_protocol": credentialProtocol,
	} {
		if value == "" {
			missing = append(missing, "cmdb_credential_policy_approval_context_contract")
			continue
		}
		context[key] = value
	}
	if ref := sanitizeRemoteMutationValue("business_group_ref", payload.BusinessGroupRef); ref != "" {
		context["business_group_ref"] = ref
	}
	if ref := sanitizeRemoteMutationValue("team_ref", payload.TeamRef); ref != "" {
		context["team_ref"] = ref
	}
	resourceID := sanitizeRemoteMutationValue("resource_id", payload.ResourceID)
	expectedResourceID := cmdbCredentialPolicyApprovalRequestResourceID(hostRef, agentRef, pluginID)
	if resourceID == "" {
		resourceID = expectedResourceID
	}
	if expectedResourceID == "" || resourceID != expectedResourceID {
		missing = append(missing, "cmdb_credential_policy_resource_identity_contract")
	}
	missing = uniquePackageRepositoryBlockers(missing)
	if len(missing) > 0 {
		return cmdbHighRiskApprovalInput{}, missing
	}
	return cmdbHighRiskApprovalInput{
		ContractID:     cmdbResourceApprovalRuntimeContract,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		Action:         action,
		RiskLevel:      firstNonEmpty(sanitizeRemoteMutationValue("risk_level", payload.RiskLevel), "medium"),
		Title:          firstNonEmpty(payload.Title, "CMDB credential policy release review"),
		Summary:        firstNonEmpty(payload.Summary, "Credential scope policy release requires review"),
		BusinessGroup:  sanitizeRemoteMutationValue("business_group", payload.BusinessGroup),
		Reason:         "Credential scope policy release requires review",
		PolicyID:       firstNonEmpty(cmdbCredentialPolicyApprovalPolicyID(payload.PolicyID), "cmdb.agent.plugin.credential_policy.release"),
		Context:        context,
		Diff:           map[string]any{"release_state": cmdbBlockedByContract},
		Missing:        []string{"cmdb_agent_plugin_credential_policy_contract", "cmdb_approval_execution_release_contract"},
		ExecutionState: "credential policy release is pending review and executor contracts are not enabled",
	}, nil
}

func cmdbCredentialPolicyApprovalEnum(value, allowed string) string {
	if strings.TrimSpace(value) == allowed {
		return allowed
	}
	return ""
}

func cmdbCredentialPolicyApprovalPolicyID(value string) string {
	clean := strings.TrimSpace(removeControlRunes(value))
	switch clean {
	case "", "cmdb.agent.plugin.credential_policy.release", "cmdb.agent.plugin.credential_policy.release.v1":
		return clean
	default:
		return ""
	}
}

func cmdbCredentialPolicyApprovalOperationMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case configRolloutPluginOperationAssign:
		return configRolloutPluginOperationAssign
	case configRolloutPluginOperationDispatch:
		return configRolloutPluginOperationDispatch
	default:
		return ""
	}
}

func cmdbCredentialPolicyApprovalRequestResourceID(hostRef, agentRef, pluginID string) string {
	if hostRef == "" || agentRef == "" || pluginID == "" {
		return ""
	}
	return strings.Join([]string{hostRef, agentRef, pluginID}, ":")
}

func cmdbResourceApprovalRequestBlockedEnvelope(message string, missing []string) gin.H {
	return cmdbBlockedContractEnvelope(cmdbResourceApprovalRuntimeContract, uniquePackageRepositoryBlockers(missing), gin.H{
		"message":       cmdbSafeAuditText(message),
		"safe_to_retry": false,
	})
}

func cmdbHighRiskApprovalGate(c *gin.Context, input cmdbHighRiskApprovalInput) gin.H {
	actor := requestActor(c)
	missing := cmdbHighRiskMissingContracts(input.Missing...)
	gate := cmdbBlockedContractEnvelope(input.ContractID, missing, gin.H{
		"risk_policy": gin.H{
			"status":        cmdbBlockedByContract,
			"policy_id":     firstNonEmpty(input.PolicyID, "cmdb.high_risk.default"),
			"risk_level":    firstNonEmpty(input.RiskLevel, "high"),
			"missing":       "cmdb_operation_risk_policy_contract",
			"record_source": "cmdb_approval_risk_store",
		},
		"execution": gin.H{
			"status":  cmdbBlockedByContract,
			"message": firstNonEmpty(input.ExecutionState, "executor, receipt and audit contracts are not enabled"),
		},
		"findx_audit_query": cmdbApprovalAuditQuery("cmdb.resource_approval.requested", "todo"),
	})

	auditRef := "cmdb-approval-audit-" + store.NewID()
	contextJSON := cmdbSanitizedJSON(input.Context)
	risk, riskErr := store.CreateCmdbOperationRiskRecord(&model.CmdbOperationRiskRecord{
		ResourceType:  strings.TrimSpace(input.ResourceType),
		ResourceID:    strings.TrimSpace(input.ResourceID),
		Action:        strings.TrimSpace(input.Action),
		RiskLevel:     firstNonEmpty(input.RiskLevel, "high"),
		PolicyID:      firstNonEmpty(input.PolicyID, "cmdb.high_risk.default"),
		Status:        "risk_recorded",
		Actor:         actor,
		BusinessGroup: strings.TrimSpace(input.BusinessGroup),
		Reason:        cmdbSafeAuditText(firstNonEmpty(input.Reason, input.Summary)),
		ContextJSON:   contextJSON,
		AuditRef:      auditRef,
	})
	if riskErr != nil {
		gate["approval_runtime"] = gin.H{"status": cmdbBlockedByContract, "missing": "cmdb_resource_approval_store_contract"}
		return gate
	}
	approval, approvalErr := store.CreateCmdbResourceApproval(&model.CmdbResourceApproval{
		View:          "todo",
		Requester:     actor,
		Approver:      "cmdb-operator",
		ResourceType:  strings.TrimSpace(input.ResourceType),
		ResourceID:    strings.TrimSpace(input.ResourceID),
		Action:        strings.TrimSpace(input.Action),
		RiskLevel:     firstNonEmpty(input.RiskLevel, "high"),
		Status:        "pending_review",
		Title:         cmdbSafeAuditText(firstNonEmpty(input.Title, "CMDB high-risk operation review")),
		Summary:       cmdbSafeAuditText(firstNonEmpty(input.Summary, input.Reason)),
		BusinessGroup: strings.TrimSpace(input.BusinessGroup),
		WorkflowState: "pending_review",
		RiskRecordID:  risk.ID,
		ContextJSON:   contextJSON,
		DiffJSON:      cmdbSanitizedJSON(input.Diff),
		AuditRef:      auditRef,
	})
	if approvalErr != nil {
		_, _ = store.DeleteCmdbOperationRiskRecord(risk.ID)
		gate["approval_runtime"] = gin.H{"status": cmdbBlockedByContract, "missing": "cmdb_resource_approval_store_contract", "risk_record_id": risk.ID}
		return gate
	}
	if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
		ID:           auditRef,
		Actor:        actor,
		Action:       "cmdb.resource_approval.requested",
		ResourceType: "cmdb_resource_approval",
		ResourceID:   approval.ID,
		Scope:        "cmdb",
		Status:       "blocked",
		ClientIP:     c.ClientIP(),
		Summary:      "CMDB high-risk approval request recorded; execution remains blocked by contract",
		Details: map[string]any{
			"approval_id":    approval.ID,
			"risk_record_id": risk.ID,
			"resource_type":  approval.ResourceType,
			"resource_id":    approval.ResourceID,
			"action":         approval.Action,
			"risk_level":     approval.RiskLevel,
			"context":        cmdbJSONMapForOutput(contextJSON),
			"diff":           cmdbJSONMapForOutput(approval.DiffJSON),
		},
	}); err != nil {
		_, _ = store.DeleteCmdbResourceApproval(approval.ID)
		_, _ = store.DeleteCmdbOperationRiskRecord(risk.ID)
		gate["approval_runtime"] = gin.H{"status": cmdbBlockedByContract, "missing": "cmdb_approval_audit_receipt_contract"}
		return gate
	}
	gate["approval_request"] = cmdbResourceApprovalDTO(*approval)
	gate["risk_record"] = cmdbOperationRiskDTO(*risk)
	gate["approval_runtime"] = gin.H{
		"status":          "ready_with_contract_gaps",
		"approval_id":     approval.ID,
		"risk_record_id":  risk.ID,
		"workflow_state":  approval.WorkflowState,
		"execution_state": cmdbBlockedByContract,
	}
	return gate
}

func cmdbResourceApprovalListEnvelope(view, actor string, items []model.CmdbResourceApproval) gin.H {
	requests := make([]gin.H, 0, len(items))
	for _, item := range items {
		requests = append(requests, cmdbResourceApprovalDTO(item))
	}
	return gin.H{
		"status":            "ready_with_contract_gaps",
		"contract_id":       cmdbResourceApprovalRuntimeContract,
		"view":              view,
		"actor":             actor,
		"total":             len(requests),
		"approval_requests": requests,
		"workflow_state": gin.H{
			"status":  "ready_with_contract_gaps",
			"missing": "cmdb_resource_approval_workflow_contract",
		},
		"risk_policy": gin.H{
			"status":  cmdbBlockedByContract,
			"missing": "cmdb_operation_risk_policy_contract",
		},
		"audit_receipt": gin.H{
			"status": "ready_with_contract_gaps",
			"query":  cmdbApprovalAuditQuery("cmdb.resource_approval.requested", view),
		},
		"missing_contracts": []string{
			"cmdb_resource_approval_workflow_contract",
			"cmdb_operation_risk_policy_contract",
			"cmdb_action_preflight_contract",
			"cmdb_action_audit_receipt_contract",
		},
		"blocked_contracts": []gin.H{
			{"id": "cmdb_resource_approval_workflow_contract", "status": cmdbBlockedByContract},
			{"id": "cmdb_operation_risk_policy_contract", "status": cmdbBlockedByContract},
			{"id": "cmdb_action_preflight_contract", "status": cmdbBlockedByContract},
			{"id": "cmdb_action_audit_receipt_contract", "status": cmdbBlockedByContract},
		},
		"findx_audit_query": cmdbApprovalAuditQuery("cmdb.resource_approval.read", view),
	}
}

func cmdbResourceApprovalDetailEnvelope(item model.CmdbResourceApproval) gin.H {
	return gin.H{
		"status":            "ready_with_contract_gaps",
		"contract_id":       cmdbResourceApprovalRuntimeContract,
		"approval_request":  cmdbResourceApprovalDTO(item),
		"missing_contracts": []string{"cmdb_operation_risk_policy_contract", "cmdb_action_preflight_contract", "cmdb_action_audit_receipt_contract"},
		"findx_audit_query": cmdbApprovalAuditQuery("cmdb.resource_approval.read", item.View),
	}
}

func cmdbResourceApprovalReviewEnvelope(item model.CmdbResourceApproval) gin.H {
	return gin.H{
		"code":              cmdbBlockedByContract,
		"status":            cmdbBlockedByContract,
		"contract_id":       cmdbResourceApprovalRuntimeContract,
		"approval_request":  cmdbResourceApprovalDTO(item),
		"message":           "CMDB approval review was recorded, but high-risk execution remains blocked until executor and receipt contracts are ready.",
		"missing_contracts": cmdbHighRiskMissingContracts("cmdb_approval_execution_release_contract"),
		"safe_to_retry":     false,
		"findx_audit_query": cmdbApprovalAuditQuery("cmdb.resource_approval.review", item.View),
	}
}

func cmdbResourceApprovalReviewBlockedEnvelope(id, message string) gin.H {
	return cmdbBlockedContractEnvelope(cmdbResourceApprovalRuntimeContract, []string{
		"cmdb_resource_approval_workflow_contract",
		"cmdb_approval_decision_contract",
		"cmdb_approval_audit_receipt_contract",
	}, gin.H{"approval_id": id, "message": message})
}

func cmdbResourceApprovalDTO(item model.CmdbResourceApproval) gin.H {
	return gin.H{
		"id":             item.ID,
		"view":           cmdbApprovalView(item.View),
		"requester":      item.Requester,
		"approver":       item.Approver,
		"resource_type":  item.ResourceType,
		"resource_id":    item.ResourceID,
		"action":         item.Action,
		"risk_level":     item.RiskLevel,
		"status":         item.Status,
		"title":          cmdbSafeAuditText(item.Title),
		"summary":        cmdbSafeAuditText(item.Summary),
		"business_group": item.BusinessGroup,
		"workflow_state": item.WorkflowState,
		"risk_record_id": item.RiskRecordID,
		"context":        cmdbJSONMapForOutput(item.ContextJSON),
		"diff":           cmdbJSONMapForOutput(item.DiffJSON),
		"audit_ref":      item.AuditRef,
		"created_at":     item.CreatedAt,
		"updated_at":     item.UpdatedAt,
		"decided_at":     item.DecidedAt,
		"decision_actor": item.DecisionActor,
		"decision_note":  cmdbSafeAuditText(item.DecisionNote),
	}
}

func cmdbOperationRiskDTO(item model.CmdbOperationRiskRecord) gin.H {
	return gin.H{
		"id":             item.ID,
		"resource_type":  item.ResourceType,
		"resource_id":    item.ResourceID,
		"action":         item.Action,
		"risk_level":     item.RiskLevel,
		"policy_id":      item.PolicyID,
		"status":         item.Status,
		"actor":          item.Actor,
		"business_group": item.BusinessGroup,
		"reason":         cmdbSafeAuditText(item.Reason),
		"context":        cmdbJSONMapForOutput(item.ContextJSON),
		"audit_ref":      item.AuditRef,
		"created_at":     item.CreatedAt,
		"updated_at":     item.UpdatedAt,
	}
}

func cmdbApprovalView(view string) string {
	switch strings.ToLower(strings.TrimSpace(view)) {
	case "todo", "archive":
		return strings.ToLower(strings.TrimSpace(view))
	default:
		return "mine"
	}
}

func cmdbApprovalAuditQuery(action, view string) gin.H {
	return gin.H{
		"source":        "findx_audit",
		"scope":         "cmdb",
		"resource_type": "cmdb_resource_approval",
		"action":        action,
		"view":          cmdbApprovalView(view),
	}
}

func cmdbSanitizedJSON(value map[string]any) string {
	if value == nil {
		return "{}"
	}
	raw, err := json.Marshal(store.SanitizeMonitorAuditDetails(value))
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func cmdbJSONMapForOutput(raw string) map[string]any {
	out := map[string]any{}
	if strings.TrimSpace(raw) == "" {
		return out
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return map[string]any{}
	}
	return store.SanitizeMonitorAuditDetails(out)
}

func cmdbSafeAuditText(value string) string {
	return stringFromSanitizedAudit(store.SanitizeMonitorAuditDetails(map[string]any{"value": value})["value"])
}

func stringFromSanitizedAudit(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(raw)
}
