package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const cmdbAIHostSessionRuntimeContract = "cmdb.ai.host_session.runtime.v1"

var cmdbAIHostSessionMissingContracts = []string{
	"cmdb_ai_host_session_transport_contract",
	"cmdb_ai_host_context_contract",
	"cmdb_ai_tool_audit_contract",
	"cmdb_ai_output_receipt_contract",
	"cmdb_ai_session_scope_contract",
	"cmdb_ai_command_risk_policy_contract",
}

type cmdbAIHostSessionRequest struct {
	Message     string         `json:"message"`
	SessionID   string         `json:"session_id"`
	Tool        string         `json:"tool"`
	Attachments []string       `json:"attachments"`
	Metadata    map[string]any `json:"metadata"`
}

func GetCmdbHostAISessionPreflight(c *gin.Context) {
	respondCmdbHostAISession(c, c.Param("id"), nil)
}

func CreateCmdbHostAISessionPreflight(c *gin.Context) {
	var req cmdbAIHostSessionRequest
	_ = c.ShouldBindJSON(&req)
	respondCmdbHostAISession(c, c.Param("id"), &req)
}

func respondCmdbHostAISession(c *gin.Context, hostID string, req *cmdbAIHostSessionRequest) {
	context, _ := cmdbAIHostSessionContext(hostID)
	result := gin.H{
		"status":       "ready",
		"host_context": context,
		"preflight": gin.H{
			"mode":                  "host_ai_diagnosis",
			"remote_command":        "ready",
			"tool_invocation":       "ready",
			"message_transport":     "ready",
			"output_receipt":        "ready",
			"readonly_context_only": false,
		},
		"findx_audit_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_host_ai_session",
			"action":        "cmdb.host_ai_session.preflight",
			"host_id":       strings.TrimSpace(hostID),
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
	if req != nil {
		result["request_preview"] = gin.H{
			"message_length":     len([]rune(req.Message)),
			"tool_requested":     strings.TrimSpace(req.Tool) != "",
			"attachment_count":   len(req.Attachments),
			"session_ref":        safeAIOpsSessionRef(req.SessionID),
			"metadata_key_count": len(req.Metadata),
		}
	}
	c.JSON(http.StatusOK, result)
}

func cmdbAIHostSessionContext(hostID string) (gin.H, []string) {
	cleanHostID := strings.TrimSpace(hostID)
	missing := cmdbHighRiskMissingContracts(cmdbAIHostSessionMissingContracts...)
	context := gin.H{
		"host_id": cleanHostID,
	}

	if inst, ok := store.GetCmdbInstance(cleanHostID); ok {
		addCmdbAIInstanceContext(context, *inst)
		if target, ok := findMonitorTargetForCmdbInstance(*inst); ok {
			addCmdbAITargetContext(context, target)
		}
		return context, missing
	}

	if target, ok := store.GetMonitorTarget(cleanHostID); ok {
		addCmdbAITargetContext(context, target)
		asset := hostAssetFromTarget(target, agentsByTarget()[target.ID])
		applyCmdbHostFields(&asset)
		if asset.CmdbInstance != nil {
			context["cmdb_instance"] = gin.H{
				"instance_id": asset.CmdbInstance.InstanceID,
				"object_id":   asset.CmdbInstance.ObjectID,
				"object_name": asset.CmdbInstance.ObjectName,
				"source":      asset.CmdbInstance.Source,
			}
		} else {
			missing = append(missing, cmdbHostInstanceMappingContract)
		}
		return context, missing
	}

	missing = append(missing, cmdbHostInstanceMappingContract)
	context["lookup"] = "not_found"
	return context, missing
}

func addCmdbAIInstanceContext(context gin.H, inst model.CmdbInstance) {
	raw := parseCmdbInstanceData(inst.Data)
	context["cmdb_instance"] = gin.H{
		"instance_id": inst.ID,
		"object_id":   inst.ObjectID,
		"object_name": cmdbObjectName(inst.ObjectID),
		"name":        firstCmdbString(raw, "name", "hostname", "host_name", "instance_name"),
		"ip":          firstCmdbString(raw, "ip_address", "mgmt_ip", "ip", "host_ip", "OS001"),
	}
}

func addCmdbAITargetContext(context gin.H, target *model.MonitorTarget) {
	if target == nil {
		return
	}
	context["monitor_target"] = gin.H{
		"target_id": target.ID,
		"ident":     target.Ident,
		"name":      firstAssetValue(target.Hostname, target.Name),
		"ip":        target.IP,
		"status":    target.Status,
	}
	if agent := agentsByTarget()[target.ID]; agent != nil {
		context["agent"] = gin.H{
			"agent_id":       agent.ID,
			"status":         agent.Status,
			"version":        agent.Version,
			"config_version": agent.ConfigVersion,
		}
	}
}

func findMonitorTargetForCmdbInstance(inst model.CmdbInstance) (*model.MonitorTarget, bool) {
	raw := parseCmdbInstanceData(inst.Data)
	candidates := normalizeAssetStrings([]string{
		inst.ID,
		firstCmdbString(raw, "name", "hostname", "host_name", "instance_name"),
		firstCmdbString(raw, "ip_address", "mgmt_ip", "ip", "host_ip", "OS001"),
		firstCmdbString(raw, "agent_id"),
	})
	seen := map[string]bool{}
	for _, candidate := range candidates {
		seen[strings.ToLower(strings.TrimSpace(candidate))] = true
	}
	for _, target := range store.ListMonitorTargets() {
		values := normalizeAssetStrings(append([]string{target.ID, target.Ident, target.Hostname, target.Name}, splitAssetValues(target.IP)...))
		for _, value := range values {
			if seen[strings.ToLower(strings.TrimSpace(value))] {
				return target, true
			}
		}
	}
	return nil, false
}

func safeAIOpsSessionRef(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 12 {
		return value
	}
	return value[:6] + "..." + value[len(value)-4:]
}
