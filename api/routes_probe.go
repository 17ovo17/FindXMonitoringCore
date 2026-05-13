package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerProbeRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	v1.GET("/probes/status-pages/main", mw.readRequired, handler.GetProbeStatusPage)
	v1.GET("/probes/status-pages/:slug", mw.readRequired, handler.GetProbeStatusPage)
	v1.GET("/probes/status-pages", mw.readRequired, handler.ListProbeStatusPages)
	v1.POST("/probes/status-pages", mw.adminRequired, handler.SaveProbeStatusPage)
	v1.PUT("/probes/status-pages/:id", mw.adminRequired, handler.SaveProbeStatusPage)

	v1.GET("/probes/checks", mw.readRequired, handler.ListProbeChecks)
	v1.POST("/probes/checks", mw.adminRequired, handler.CreateProbeCheck)
	v1.GET("/probes/checks/:id", mw.readRequired, handler.GetProbeCheck)
	v1.PUT("/probes/checks/:id", mw.adminRequired, handler.UpdateProbeCheck)
	v1.DELETE("/probes/checks/:id", mw.adminRequired, handler.DeleteProbeCheck)
	v1.POST("/probes/checks/:id/test", mw.adminRequired, handler.TestProbeCheckBlocked)
	v1.POST("/probes/checks/:id/enable", mw.adminRequired, handler.EnableProbeCheck)
	v1.POST("/probes/checks/:id/disable", mw.adminRequired, handler.DisableProbeCheck)
	v1.GET("/probes/checks/:id/notification-bindings", mw.readRequired, handler.ListProbeNotificationBindings)
	v1.PUT("/probes/checks/:id/notification-bindings", mw.adminRequired, handler.SaveProbeNotificationBindings)
	v1.GET("/probes/checks/:id/alert-bindings", mw.readRequired, handler.ListProbeAlertBindings)
	v1.PUT("/probes/checks/:id/alert-bindings", mw.adminRequired, handler.SaveProbeAlertBindings)

	v1.GET("/probes/incidents", mw.readRequired, handler.ListProbeIncidents)
	v1.POST("/probes/incidents", mw.adminRequired, handler.CreateProbeIncident)
	v1.PUT("/probes/incidents/:id", mw.adminRequired, handler.UpdateProbeIncident)
	v1.DELETE("/probes/incidents/:id", mw.adminRequired, handler.DeleteProbeIncident)
}
