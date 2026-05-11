package main

import (
	"ai-workbench-api/internal/handler"
	"ai-workbench-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

type routeMiddleware struct {
	adminRequired     gin.HandlerFunc
	readRequired      gin.HandlerFunc
	roleAdminRequired gin.HandlerFunc
	monitorRequired   func(string, string) gin.HandlerFunc
}

func registerRoutes(r *gin.Engine) {
	r.Use(middleware.RateLimit(100))
	registerRootRoutes(r)

	mw := routeMiddleware{
		adminRequired:     requireAdminToken(),
		readRequired:      handler.RequireAuth(),
		roleAdminRequired: handler.RequireRole("admin"),
		monitorRequired:   handler.RequireMonitorPermission,
	}

	v1 := r.Group("/api/v1")
	registerAPIV1Routes(r, v1, mw)
}

func registerRootRoutes(r *gin.Engine) {
	r.GET("/metrics", handler.Metrics)
	r.GET("/categraf/configs", handler.CategrafHTTPProviderConfigs)
	r.POST("/v1/n9e/heartbeat", handler.CategrafN9EHeartbeat)
	r.POST("/prometheus/v1/write", handler.CategrafPrometheusRemoteWrite)
	r.POST("/findx-agent/v1/logs", handler.FindXAgentLogsCompatibleReceiver)
	r.POST("/findx-agent/v1/traces", handler.FindXAgentTracesCompatibleReceiver)
}

func registerAPIV1Routes(r *gin.Engine, v1 *gin.RouterGroup, mw routeMiddleware) {
	registerAIOpsRoutes(v1, mw)
	registerDiagnoseAndAgentRoutes(v1, mw)
	registerPlatformRoutes(v1, mw)
	registerAssetsAndLogsRoutes(v1, mw)
	registerAuthAndUserRoutes(v1, mw)
	registerMonitorCoreRoutes(v1, mw)
	registerMonitorCatalogRoutes(v1, mw)
	registerMonitorAlertRoutes(v1, mw)
	registerFindXAgentRoutes(v1, mw)
	registerKnowledgeRoutes(v1, mw)
	registerWorkflowAndNotificationRoutes(v1, mw)
	registerMetricsAndRunbookRoutes(v1, mw)
	registerSettingsRoutes(v1, mw)
	registerDownloadRoute(r)
}
