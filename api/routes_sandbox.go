package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerSandboxRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	sandbox := v1.Group("/sandbox")
	sandbox.GET("/policy", mw.readRequired, handler.SandboxGetPolicy)
	sandbox.PUT("/policy", mw.adminRequired, handler.SandboxSetPolicy)
	sandbox.GET("/approvals", mw.readRequired, handler.SandboxListApprovals)
	sandbox.POST("/approvals/:id/resolve", mw.adminRequired, handler.SandboxResolveApproval)
	sandbox.GET("/audit", mw.readRequired, handler.SandboxListAudit)
}
