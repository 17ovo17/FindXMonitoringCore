package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerFindXAgentRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	v1.GET("/findx-agents", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgents)
	v1.POST("/findx-agents/register", handler.FindXAgentRegister)
	v1.POST("/findx-agents/heartbeat", handler.FindXAgentHeartbeat)
	v1.GET("/findx-agents/packages", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgentPackages)
	v1.GET("/findx-agents/lifecycle", mw.monitorRequired("findx_agent", "read"), handler.FindXAgentLifecycle)
	v1.GET("/findx-agents/install-plans", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgentInstallPlans)
	v1.POST("/findx-agents/install-plans", mw.monitorRequired("findx_agent", "write"), handler.CreateFindXAgentInstallPlan)
	v1.GET("/findx-agents/install-executions", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgentInstallExecutions)
	v1.GET("/findx-agents/installers/linux.sh", mw.monitorRequired("findx_agent", "read"), handler.DownloadFindXAgentLinuxInstaller)
	v1.GET("/findx-agents/installers/windows.ps1", mw.monitorRequired("findx_agent", "read"), handler.DownloadFindXAgentWindowsPowerShellInstaller)
	v1.GET("/findx-agents/installers/windows.bat", mw.monitorRequired("findx_agent", "read"), handler.DownloadFindXAgentWindowsBatchInstaller)
	v1.GET("/findx-agents/package-downloads/:package", mw.monitorRequired("findx_agent", "read"), handler.DownloadFindXAgentPackageArtifact)
	v1.GET("/findx-agents/config-templates", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgentConfigTemplates)
	v1.GET("/findx-agents/config-rollouts", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgentConfigRollouts)
	v1.POST("/findx-agents/config-rollouts", mw.monitorRequired("findx_agent", "write"), handler.CreateFindXAgentConfigRollout)
	v1.GET("/findx-agents/data-arrival", mw.monitorRequired("findx_agent", "read"), handler.FindXAgentDataArrival)
	v1.GET("/findx-agents/data-arrival/evidence", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgentDataArrivalEvidence)
	v1.GET("/findx-agents/tasks", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgentTasks)
	v1.POST("/findx-agents/tasks", mw.monitorRequired("findx_agent", "write"), handler.CreateFindXAgentTask)
	v1.GET("/n9e/agents", handler.ListN9eAgents)
	v1.GET("/n9e/alerts", handler.ListN9eAlerts)
}

func registerKnowledgeRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	v1.GET("/knowledge/cases", mw.adminRequired, handler.ListCases)
	v1.GET("/knowledge/cases/export", mw.adminRequired, handler.ExportCases)
	v1.POST("/knowledge/cases", mw.adminRequired, handler.CreateCase)
	v1.POST("/knowledge/cases/import", mw.adminRequired, handler.ImportCases)
	v1.GET("/knowledge/cases/:id", handler.GetCase)
	v1.PUT("/knowledge/cases/:id", mw.adminRequired, handler.UpdateCase)
	v1.DELETE("/knowledge/cases/:id", mw.adminRequired, handler.DeleteCase)
	v1.POST("/knowledge/documents/upload", mw.adminRequired, handler.UploadDocument)
	v1.GET("/knowledge/documents", handler.ListDocumentsHandler)
	v1.GET("/knowledge/documents/:id", handler.GetDocumentHandler)
	v1.DELETE("/knowledge/documents/:id", mw.adminRequired, handler.DeleteDocumentHandler)
	v1.POST("/knowledge/documents/:id/reindex", mw.adminRequired, handler.ReindexDocumentHandler)
	v1.POST("/knowledge/search", handler.SearchKnowledge)
	v1.GET("/knowledge/search/stats", handler.KnowledgeSearchStats)
	v1.POST("/knowledge/search/badcase", mw.adminRequired, handler.SubmitSearchBadcase)
	v1.POST("/knowledge/reindex-all", mw.adminRequired, handler.ReindexAllHandler)
}
