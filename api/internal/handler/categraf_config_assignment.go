package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// CategrafConfigAssignRequest 配置分配请求
type CategrafConfigAssignRequest struct {
	AgentIdent string `json:"agent_ident" binding:"required"`
	InputName  string `json:"input_name" binding:"required"`
	Config     string `json:"config" binding:"required"`
	Format     string `json:"format"`
}

// ListCategrafConfigAssignments 列出某个 agent 的所有配置分配
// GET /api/v1/categraf/config-assignments?agent_ident=xxx
func ListCategrafConfigAssignments(c *gin.Context) {
	agentIdent := strings.TrimSpace(c.Query("agent_ident"))
	if agentIdent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_ident is required"})
		return
	}
	entries := store.ListCategrafConfigEntries(agentIdent)
	c.JSON(http.StatusOK, gin.H{
		"items": entries,
		"total": len(entries),
	})
}

// UpsertCategrafConfigAssignment 创建或更新配置分配
// 当 UI 用户"推送"配置时，实际上是更新分配记录
// 下次 agent 轮询时会自动获取新配置
// POST /api/v1/categraf/config-assignments
func UpsertCategrafConfigAssignment(c *gin.Context) {
	var req CategrafConfigAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := store.UpsertCategrafConfigByAgent(req.AgentIdent, req.InputName, req.Config, req.Format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          entry.ID,
		"agent_ident": entry.AgentIdent,
		"input_name":  entry.InputName,
		"checksum":    entry.Checksum,
		"format":      entry.Format,
		"enabled":     entry.Enabled,
		"updated_at":  entry.UpdatedAt,
	})
}

// DeleteCategrafConfigAssignment 删除配置分配（禁用某个插件）
// DELETE /api/v1/categraf/config-assignments/:id
func DeleteCategrafConfigAssignment(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	if err := store.DeleteCategrafConfigEntry(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DisableCategrafConfigAssignment 禁用某个 agent 的某个插件配置
// POST /api/v1/categraf/config-assignments/disable
func DisableCategrafConfigAssignment(c *gin.Context) {
	var req struct {
		AgentIdent string `json:"agent_ident" binding:"required"`
		InputName  string `json:"input_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := store.DisableCategrafConfigByAgent(req.AgentIdent, req.InputName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// PreviewCategrafProviderResponse 预览某个 agent 将收到的 HTTP Provider 响应
// GET /api/v1/categraf/config-assignments/preview?agent_ident=xxx
func PreviewCategrafProviderResponse(c *gin.Context) {
	agentIdent := strings.TrimSpace(c.Query("agent_ident"))
	if agentIdent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_ident is required"})
		return
	}
	resp := store.BuildCategrafProviderResponse(agentIdent)
	c.JSON(http.StatusOK, resp)
}
