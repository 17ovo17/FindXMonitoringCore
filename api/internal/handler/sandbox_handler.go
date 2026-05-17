package handler

import (
	"net/http"
	"strconv"

	"ai-workbench-api/internal/sandbox"

	"github.com/gin-gonic/gin"
)

// 全局沙箱实例
var globalSandbox *sandbox.Sandbox

// InitSandbox 初始化全局沙箱单例
func InitSandbox() {
	globalSandbox = sandbox.New(sandbox.DefaultPolicy)
}

// GetSandbox 获取全局沙箱实例
func GetSandbox() *sandbox.Sandbox {
	if globalSandbox == nil {
		InitSandbox()
	}
	return globalSandbox
}

// SandboxGetPolicy GET /api/v1/sandbox/policy — 获取当前沙箱策略
func SandboxGetPolicy(c *gin.Context) {
	sb := GetSandbox()
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": sb.GetPolicy()})
}

// SandboxSetPolicy PUT /api/v1/sandbox/policy — 更新沙箱策略
func SandboxSetPolicy(c *gin.Context) {
	var req sandbox.Policy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	// 校验 mode 合法性
	switch req.Mode {
	case sandbox.ModeReadonly, sandbox.ModeAutoReview, sandbox.ModeFullAccess:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "无效的策略模式，可选: readonly, auto_review, full_access"})
		return
	}
	if req.MaxTimeout <= 0 {
		req.MaxTimeout = 300
	}
	sb := GetSandbox()
	sb.SetPolicy(req)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "策略已更新"})
}

// SandboxListApprovals GET /api/v1/sandbox/approvals — 获取待确认的请求列表
func SandboxListApprovals(c *gin.Context) {
	sb := GetSandbox()
	pending := sb.ListPendingApprovals()
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"approvals": pending, "total": len(pending)}})
}

// SandboxResolveApproval POST /api/v1/sandbox/approvals/:id/resolve — 确认/拒绝请求
func SandboxResolveApproval(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Action string `json:"action"` // approve / deny
		User   string `json:"user"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	var status sandbox.ApprovalStatus
	switch req.Action {
	case "approve":
		status = sandbox.ApprovalApproved
	case "deny":
		status = sandbox.ApprovalDenied
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "action 必须为 approve 或 deny"})
		return
	}

	sb := GetSandbox()
	if err := sb.Resolve(id, status, req.User); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "操作已处理"})
}

// SandboxListAudit GET /api/v1/sandbox/audit — 查询审计日志
func SandboxListAudit(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 50
	}
	sb := GetSandbox()
	entries := sb.ListAudit(limit)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"entries": entries, "total": len(entries)}})
}
