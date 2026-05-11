package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerWorkflowAndNotificationRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	v1.POST("/diagnosis/start", mw.adminRequired, handler.StartDiagnosisWorkflow)
	v1.GET("/workflows", handler.ListWorkflows)
	v1.GET("/workflows/:id", handler.GetWorkflow)
	v1.POST("/workflows", mw.adminRequired, handler.CreateWorkflow)
	v1.PUT("/workflows/:id", mw.adminRequired, handler.UpdateWorkflow)
	v1.DELETE("/workflows/:id", mw.adminRequired, handler.DeleteWorkflow)
	v1.POST("/workflows/:id/run", handler.RunWorkflowAPI)
	v1.POST("/workflows/:id/stream", handler.StreamWorkflowAPI)
	v1.GET("/workflows/:id/runs", handler.ListWorkflowRuns)
	v1.GET("/schedules", handler.ListSchedulesHandler)
	v1.POST("/schedules", mw.adminRequired, handler.CreateScheduleHandler)
	v1.DELETE("/schedules/:id", mw.adminRequired, handler.DeleteScheduleHandler)
	v1.GET("/notifications/channels", handler.ListNotificationChannels)
	v1.POST("/notifications/channels", mw.adminRequired, handler.SaveNotificationChannel)
	v1.DELETE("/notifications/channels/:id", mw.adminRequired, handler.DeleteNotificationChannel)
	v1.GET("/notifications/rules", handler.ListNotificationRules)
	v1.POST("/notifications/rules", mw.adminRequired, handler.SaveNotificationRules)
	v1.DELETE("/notifications/rules", mw.adminRequired, handler.DeleteNotificationRules)
	v1.POST("/notifications/rules/test", mw.adminRequired, handler.TestNotificationRule)
	v1.GET("/notifications/rules/:id/statistics", handler.GetNotificationStatistics)
	v1.GET("/notifications/rules/:id/events", handler.GetNotificationEvents)
	v1.GET("/notifications/rules/:id/alert-rules", handler.GetNotificationAlertRules)
	v1.GET("/notifications/rules/:id/sub-alert-rules", handler.GetNotificationSubAlertRules)
	v1.POST("/notifications/rules/:id/enable", mw.adminRequired, handler.EnableNotificationRule)
	v1.POST("/notifications/rules/:id/disable", mw.adminRequired, handler.DisableNotificationRule)
	v1.POST("/notifications/rules/:id/clone", mw.adminRequired, handler.CloneNotificationRule)
	v1.GET("/notifications/rules/:id", handler.GetNotificationRule)
	v1.PUT("/notifications/rules/:id", mw.adminRequired, handler.UpdateNotificationRule)
	v1.DELETE("/notifications/rules/:id", mw.adminRequired, handler.DeleteNotificationRules)
	registerNotificationTemplateRoutes(v1, mw)
}

func registerNotificationTemplateRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	v1.GET("/notifications/templates", handler.ListNotificationTemplates)
	v1.POST("/notifications/templates", mw.adminRequired, handler.SaveNotificationTemplates)
	v1.DELETE("/notifications/templates", mw.adminRequired, handler.DeleteNotificationTemplates)
	v1.POST("/notifications/templates/preview", mw.adminRequired, handler.PreviewNotificationTemplate)
	v1.POST("/notifications/templates/:id/preview", mw.adminRequired, handler.PreviewNotificationTemplateByID)
	v1.POST("/notifications/templates/:id/clone", mw.adminRequired, handler.CloneNotificationTemplate)
	v1.GET("/notifications/templates/:id", handler.GetNotificationTemplate)
	v1.PUT("/notifications/templates/:id", mw.adminRequired, handler.UpdateNotificationTemplate)
	v1.DELETE("/notifications/templates/:id", mw.adminRequired, handler.DeleteNotificationTemplates)
}

func registerMetricsAndRunbookRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	v1.POST("/metrics/scan", mw.adminRequired, handler.ScanMetrics)
	v1.POST("/metrics/auto-adapt", mw.adminRequired, handler.AutoAdaptMetrics)
	v1.GET("/metrics/mappings", handler.ListMetricsMappings)
	v1.PUT("/metrics/mappings/:id", mw.adminRequired, handler.UpdateMetricMapping)
	v1.POST("/metrics/mappings/confirm", mw.adminRequired, handler.ConfirmMappings)
	v1.GET("/knowledge/runbooks", handler.ListRunbooks)
	v1.GET("/knowledge/runbooks/:id", handler.GetRunbook)
	v1.POST("/knowledge/runbooks", mw.adminRequired, handler.CreateRunbook)
	v1.PUT("/knowledge/runbooks/:id", mw.adminRequired, handler.UpdateRunbook)
	v1.DELETE("/knowledge/runbooks/:id", mw.adminRequired, handler.DeleteRunbook)
	v1.POST("/knowledge/runbooks/:id/execute", mw.adminRequired, handler.ExecuteRunbook)
	v1.GET("/knowledge/runbooks/:id/history", handler.ListRunbookHistory)
	v1.POST("/diagnosis/feedback", handler.SubmitFeedback)
	v1.GET("/diagnosis/feedback", handler.ListFeedbacksByDiagnosisHandler)
	v1.GET("/diagnosis/feedback/all", handler.ListAllFeedbacks)
	v1.GET("/diagnosis/feedback/stats", handler.FeedbackStats)
	v1.GET("/diagnosis/verifications", handler.ListVerifications)
	v1.POST("/diagnosis/archive", mw.adminRequired, handler.ArchiveDiagnosis)
}

func registerSettingsRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	v1.GET("/prompts", handler.ListPromptTemplates)
	v1.GET("/prompts/:name", handler.GetPromptTemplate)
	v1.POST("/prompts", mw.adminRequired, handler.CreatePromptTemplate)
	v1.PUT("/prompts/:name", mw.adminRequired, handler.UpdatePromptTemplate)
	v1.DELETE("/prompts/:name", mw.adminRequired, handler.DeletePromptTemplate)
	v1.GET("/settings/ai", handler.GetAISettings)
	v1.PUT("/settings/ai", mw.adminRequired, handler.UpdateAISettings)
	v1.GET("/settings/ai/:key", handler.GetAISetting)
	v1.PUT("/settings/ai/:key", mw.adminRequired, handler.UpdateAISettingHandler)
	v1.DELETE("/settings/ai/:key", mw.adminRequired, handler.DeleteAISetting)
	v1.GET("/settings/embedding", handler.GetEmbeddingSettings)
	v1.PUT("/settings/embedding", mw.adminRequired, handler.UpdateEmbeddingSettings)
	v1.POST("/settings/embedding/test", mw.adminRequired, handler.TestEmbeddingConnection)
	v1.GET("/settings/reranker", handler.GetRerankerSettings)
	v1.PUT("/settings/reranker", mw.adminRequired, handler.UpdateRerankerSettings)
}

func registerDownloadRoute(r *gin.Engine) {
	r.Static("/download", "./assets")
}
