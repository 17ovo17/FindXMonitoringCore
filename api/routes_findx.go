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
	v1.POST("/findx-agents/config-rollouts/:id/receipts", mw.monitorRequired("findx_agent", "write"), handler.IngestFindXAgentConfigRolloutReceipt)
	v1.GET("/findx-agents/data-arrival", mw.monitorRequired("findx_agent", "read"), handler.FindXAgentDataArrival)
	v1.GET("/findx-agents/data-arrival/evidence", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgentDataArrivalEvidence)
	v1.GET("/findx-agents/tasks", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgentTasks)
	v1.POST("/findx-agents/tasks", mw.monitorRequired("findx_agent", "write"), handler.CreateFindXAgentTask)
	v1.GET("/n9e/agents", handler.ListN9eAgents)
	v1.GET("/n9e/alerts", handler.ListN9eAlerts)
	v1.POST("/findx-agents/:id/verify-heartbeat", mw.monitorRequired("findx_agent", "read"), handler.VerifyAgentHeartbeat)
	v1.POST("/findx-agents/batch-verify", mw.monitorRequired("findx_agent", "read"), handler.BatchVerifyAgentHeartbeat)

	// 插件配置管理
	v1.GET("/findx-agents/plugins", mw.monitorRequired("findx_agent", "read"), handler.ListFindXAgentPlugins)
	v1.GET("/findx-agents/:id/config", mw.monitorRequired("findx_agent", "read"), handler.GetFindXAgentConfig)
	v1.PUT("/findx-agents/:id/config", mw.monitorRequired("findx_agent", "write"), handler.UpdateFindXAgentConfig)
	v1.PATCH("/findx-agents/:id/plugins/:pluginId", mw.monitorRequired("findx_agent", "write"), handler.PatchFindXAgentPlugin)
	v1.POST("/findx-agents/:id/plugins/:pluginId/config", mw.monitorRequired("findx_agent", "write"), handler.UpdateFindXAgentPluginConfig)
	v1.POST("/findx-agents/config-push", mw.monitorRequired("findx_agent", "write"), handler.FindXAgentConfigPushBatch)
	v1.GET("/findx-agents/:id/environment", mw.monitorRequired("findx_agent", "read"), handler.GetFindXAgentEnvironment)
	v1.POST("/findx-agents/:id/auto-adapt", mw.monitorRequired("findx_agent", "write"), handler.FindXAgentAutoAdapt)
	v1.POST("/findx-agents/:id/plugins/:pluginId/start", mw.monitorRequired("findx_agent", "write"), handler.StartFindXAgentPlugin)
	v1.POST("/findx-agents/:id/plugins/:pluginId/stop", mw.monitorRequired("findx_agent", "write"), handler.StopFindXAgentPlugin)

	// P5: Agent complete lifecycle endpoints
	v1.GET("/agent-packages", mw.monitorRequired("findx_agent", "read"), handler.ListAgentPackagesV2)
	v1.POST("/agent-packages", mw.monitorRequired("findx_agent", "write"), handler.RegisterAgentPackage)
	v1.DELETE("/agent-packages/:id", mw.monitorRequired("findx_agent", "write"), handler.DeleteAgentPackageHandler)
	v1.POST("/findx-agents/:id/install", mw.monitorRequired("findx_agent", "write"), handler.InstallFindXAgent)
	v1.POST("/findx-agents/:id/upgrade", mw.monitorRequired("findx_agent", "write"), handler.UpgradeFindXAgent)
	v1.POST("/findx-agents/:id/rollback", mw.monitorRequired("findx_agent", "write"), handler.RollbackFindXAgent)
	v1.POST("/findx-agents/:id/uninstall", mw.monitorRequired("findx_agent", "write"), handler.UninstallFindXAgent)
	v1.POST("/findx-agents/:id/config-push", mw.monitorRequired("findx_agent", "write"), handler.ConfigPushFindXAgent)
	v1.GET("/findx-agents/:id/evidence-chain", mw.monitorRequired("findx_agent", "read"), handler.GetAgentEvidenceChain)

	// Categraf HTTP Provider 配置分配管理
	v1.GET("/categraf/config-assignments", mw.monitorRequired("findx_agent", "read"), handler.ListCategrafConfigAssignments)
	v1.POST("/categraf/config-assignments", mw.monitorRequired("findx_agent", "write"), handler.UpsertCategrafConfigAssignment)
	v1.DELETE("/categraf/config-assignments/:id", mw.monitorRequired("findx_agent", "write"), handler.DeleteCategrafConfigAssignment)
	v1.POST("/categraf/config-assignments/disable", mw.monitorRequired("findx_agent", "write"), handler.DisableCategrafConfigAssignment)
	v1.GET("/categraf/config-assignments/preview", mw.monitorRequired("findx_agent", "read"), handler.PreviewCategrafProviderResponse)
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
