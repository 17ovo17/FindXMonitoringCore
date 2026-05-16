package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// --- Agent Package Management ---

func ListAgentPackagesV2(c *gin.Context) {
	rows, err := store.ListAgentPackages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list agent packages"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func RegisterAgentPackage(c *gin.Context) {
	var req model.AgentPackage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid package payload"})
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Version) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and version are required"})
		return
	}
	if req.ID == "" {
		req.ID = store.NewID()
	}
	saved, err := store.SaveAgentPackage(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "package persistence unavailable"})
		return
	}
	c.JSON(http.StatusCreated, saved)
}

func DeleteAgentPackageHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	_, found, _ := store.GetAgentPackage(id)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}
	if err := store.DeleteAgentPackage(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- Agent Lifecycle Actions ---

type installRequest struct {
	PackageID string `json:"package_id"`
	Method    string `json:"method"` // curl/certutil/helm
}

func InstallFindXAgent(c *gin.Context) {
	agentID := c.Param("id")
	var req installRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid install payload"})
		return
	}
	if req.PackageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package_id is required"})
		return
	}
	method := strings.ToLower(strings.TrimSpace(req.Method))
	if method == "" {
		method = "curl"
	}
	if method != "curl" && method != "certutil" && method != "helm" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "method must be curl, certutil, or helm"})
		return
	}
	event := model.AgentLifecycleEvent{
		ID:       store.NewID(),
		AgentID:  agentID,
		Action:   "install",
		Status:   "pending",
		Detail:   "package=" + req.PackageID + " method=" + method,
		Operator: c.GetString("username"),
	}
	saved, _ := store.SaveAgentLifecycleEvent(event)
	c.JSON(http.StatusOK, gin.H{"message": "install triggered", "event": saved})
}

type upgradeRequest struct {
	TargetVersion string `json:"target_version"`
}

func UpgradeFindXAgent(c *gin.Context) {
	agentID := c.Param("id")
	var req upgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upgrade payload"})
		return
	}
	if req.TargetVersion == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_version is required"})
		return
	}
	event := model.AgentLifecycleEvent{
		ID:       store.NewID(),
		AgentID:  agentID,
		Action:   "upgrade",
		Status:   "pending",
		Detail:   "target_version=" + req.TargetVersion,
		Operator: c.GetString("username"),
	}
	saved, _ := store.SaveAgentLifecycleEvent(event)
	c.JSON(http.StatusOK, gin.H{"message": "upgrade triggered", "event": saved})
}

func RollbackFindXAgent(c *gin.Context) {
	agentID := c.Param("id")
	event := model.AgentLifecycleEvent{
		ID:       store.NewID(),
		AgentID:  agentID,
		Action:   "rollback",
		Status:   "pending",
		Detail:   "rollback to previous version",
		Operator: c.GetString("username"),
	}
	saved, _ := store.SaveAgentLifecycleEvent(event)
	c.JSON(http.StatusOK, gin.H{"message": "rollback triggered", "event": saved})
}

func UninstallFindXAgent(c *gin.Context) {
	agentID := c.Param("id")
	event := model.AgentLifecycleEvent{
		ID:       store.NewID(),
		AgentID:  agentID,
		Action:   "uninstall",
		Status:   "pending",
		Detail:   "uninstall agent",
		Operator: c.GetString("username"),
	}
	saved, _ := store.SaveAgentLifecycleEvent(event)
	c.JSON(http.StatusOK, gin.H{"message": "uninstall triggered", "event": saved})
}

type configPushRequest struct {
	Config map[string]interface{} `json:"config"`
	Mode   string                 `json:"mode"` // single/batch/canary
}

func ConfigPushFindXAgent(c *gin.Context) {
	agentID := c.Param("id")
	var req configPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config push payload"})
		return
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "single"
	}
	if mode != "single" && mode != "batch" && mode != "canary" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be single, batch, or canary"})
		return
	}
	c.JSON(http.StatusConflict, gin.H{
		"code":        "pending",
		"status":      "pending",
		"contract_id": "cmdb.agent.plugin.config_push.runtime.v1",
		"message":     "PENDING: single-agent config push must use config-rollouts runtime contracts before execution",
		"mode":        mode,
		"agent_id":    agentID,
		"missing_contracts": []string{
			"cmdb_agent_config_rollout_contract",
			"cmdb_agent_config_push_executor_contract",
			"cmdb_agent_rollout_delivery_receipt_contract",
			"cmdb_agent_rollout_effect_receipt_contract",
			"cmdb_action_audit_receipt_contract",
		},
		"safe_to_retry":     false,
		"findx_audit_query": "scope=cmdb/resource_type=cmdb_agent_plugin/action=config_push/agent_id=" + agentID,
	})
}

func GetAgentEvidenceChain(c *gin.Context) {
	agentID := c.Param("id")
	events, err := store.ListAgentLifecycleEvents(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load evidence chain"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"agent_id": agentID, "events": events})
}
