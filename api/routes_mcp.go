package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerMcpRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	mcp := v1.Group("/mcp")
	{
		mcp.GET("/servers", handler.ListMcpServers)
		mcp.POST("/servers", mw.adminRequired, handler.CreateMcpServer)
		mcp.GET("/servers/:id", handler.GetMcpServer)
		mcp.PUT("/servers/:id", mw.adminRequired, handler.UpdateMcpServer)
		mcp.DELETE("/servers/:id", mw.adminRequired, handler.DeleteMcpServer)
		mcp.POST("/servers/:id/health-check", handler.McpServerHealthCheck)
		mcp.POST("/servers/:id/list-tools", handler.McpListTools)
		mcp.POST("/servers/:id/call-tool", handler.McpCallTool)
	}
}
