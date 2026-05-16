package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CmdbBatchUpdateGroup 批量更新资源分组
func CmdbBatchUpdateGroup(c *gin.Context) {
	var req struct {
		ResourceIDs []string `json:"resource_ids" binding:"required"`
		Group       string   `json:"group" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: resource_ids 和 group 必填"})
		return
	}
	if len(req.ResourceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource_ids 不能为空"})
		return
	}

	var success, failed int
	for _, id := range req.ResourceIDs {
		inst, ok := store.GetCmdbInstance(id)
		if !ok {
			failed++
			continue
		}
		data := cmdbBatchParseData(inst.Data)
		data["group"] = req.Group
		raw, _ := json.Marshal(data)
		inst.Data = string(raw)
		inst.Updater = requestActor(c)
		if err := store.UpdateCmdbInstance(inst); err != nil {
			failed++
			continue
		}
		success++
	}

	logrus.WithFields(logrus.Fields{
		"action":  "batch_update_group",
		"total":   len(req.ResourceIDs),
		"success": success,
		"failed":  failed,
	}).Info("cmdb: batch update group completed")

	c.JSON(http.StatusOK, gin.H{
		"total":   len(req.ResourceIDs),
		"success": success,
		"failed":  failed,
	})
}

// CmdbBatchUpdateOwner 批量更新资源负责人
func CmdbBatchUpdateOwner(c *gin.Context) {
	var req struct {
		ResourceIDs []string `json:"resource_ids" binding:"required"`
		Owner       string   `json:"owner" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: resource_ids 和 owner 必填"})
		return
	}
	if len(req.ResourceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource_ids 不能为空"})
		return
	}

	var success, failed int
	for _, id := range req.ResourceIDs {
		inst, ok := store.GetCmdbInstance(id)
		if !ok {
			failed++
			continue
		}
		data := cmdbBatchParseData(inst.Data)
		data["owner"] = req.Owner
		raw, _ := json.Marshal(data)
		inst.Data = string(raw)
		inst.Updater = requestActor(c)
		if err := store.UpdateCmdbInstance(inst); err != nil {
			failed++
			continue
		}
		success++
	}

	logrus.WithFields(logrus.Fields{
		"action":  "batch_update_owner",
		"total":   len(req.ResourceIDs),
		"success": success,
		"failed":  failed,
	}).Info("cmdb: batch update owner completed")

	c.JSON(http.StatusOK, gin.H{
		"total":   len(req.ResourceIDs),
		"success": success,
		"failed":  failed,
	})
}

// CmdbBatchUpdateMaintenance 批量设置维护状态
func CmdbBatchUpdateMaintenance(c *gin.Context) {
	var req struct {
		ResourceIDs []string `json:"resource_ids" binding:"required"`
		Maintenance bool     `json:"maintenance"`
		Reason      string   `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: resource_ids 必填"})
		return
	}
	if len(req.ResourceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource_ids 不能为空"})
		return
	}

	var success, failed int
	for _, id := range req.ResourceIDs {
		inst, ok := store.GetCmdbInstance(id)
		if !ok {
			failed++
			continue
		}
		data := cmdbBatchParseData(inst.Data)
		data["maintenance"] = req.Maintenance
		data["maintenance_reason"] = req.Reason
		data["maintenance_at"] = time.Now().Format(time.RFC3339)
		raw, _ := json.Marshal(data)
		inst.Data = string(raw)
		inst.Updater = requestActor(c)
		if err := store.UpdateCmdbInstance(inst); err != nil {
			failed++
			continue
		}
		success++
	}

	logrus.WithFields(logrus.Fields{
		"action":  "batch_update_maintenance",
		"total":   len(req.ResourceIDs),
		"success": success,
		"failed":  failed,
	}).Info("cmdb: batch update maintenance completed")

	c.JSON(http.StatusOK, gin.H{
		"total":   len(req.ResourceIDs),
		"success": success,
		"failed":  failed,
	})
}

// CmdbBatchDelete 批量删除资源
func CmdbBatchDelete(c *gin.Context) {
	var req struct {
		ResourceIDs []string `json:"resource_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: resource_ids 必填"})
		return
	}
	if len(req.ResourceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource_ids 不能为空"})
		return
	}

	var success, failed int
	for _, id := range req.ResourceIDs {
		if err := store.DeleteCmdbInstance(id); err != nil {
			failed++
			continue
		}
		success++
	}

	logrus.WithFields(logrus.Fields{
		"action":  "batch_delete",
		"actor":   requestActor(c),
		"total":   len(req.ResourceIDs),
		"success": success,
		"failed":  failed,
	}).Info("cmdb: batch delete completed")

	c.JSON(http.StatusOK, gin.H{
		"total":   len(req.ResourceIDs),
		"success": success,
		"failed":  failed,
	})
}

// CmdbBatchDeployAgent 批量触发 Agent 部署
func CmdbBatchDeployAgent(c *gin.Context) {
	var req struct {
		ResourceIDs []string `json:"resource_ids" binding:"required"`
		AgentType   string   `json:"agent_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: resource_ids 必填"})
		return
	}
	if len(req.ResourceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource_ids 不能为空"})
		return
	}

	agentType := req.AgentType
	if agentType == "" {
		agentType = "findx-agent"
	}

	var success, failed int
	for _, id := range req.ResourceIDs {
		inst, ok := store.GetCmdbInstance(id)
		if !ok {
			failed++
			continue
		}
		data := cmdbBatchParseData(inst.Data)
		data["agent_deploy_status"] = "pending"
		data["agent_deploy_type"] = agentType
		data["agent_deploy_at"] = time.Now().Format(time.RFC3339)
		raw, _ := json.Marshal(data)
		inst.Data = string(raw)
		inst.Updater = requestActor(c)
		if err := store.UpdateCmdbInstance(inst); err != nil {
			failed++
			continue
		}
		success++
	}

	logrus.WithFields(logrus.Fields{
		"action":     "batch_deploy_agent",
		"agent_type": agentType,
		"total":      len(req.ResourceIDs),
		"success":    success,
		"failed":     failed,
	}).Info("cmdb: batch deploy agent triggered")

	c.JSON(http.StatusOK, gin.H{
		"total":      len(req.ResourceIDs),
		"success":    success,
		"failed":     failed,
		"agent_type": agentType,
	})
}

func cmdbBatchParseData(raw string) map[string]any {
	data := make(map[string]any)
	if raw != "" {
		json.Unmarshal([]byte(raw), &data)
	}
	return data
}
