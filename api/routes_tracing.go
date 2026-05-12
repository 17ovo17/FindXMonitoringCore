package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

// registerTracingRoutes wires SkyWalking-backed tracing endpoints.
func registerTracingRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	v1.GET("/tracing/selectors/services", mw.readRequired, handler.TracingListServicesSW)
	v1.GET("/tracing/selectors/endpoints", mw.readRequired, handler.TracingListEndpointsSW)
	v1.GET("/tracing/selectors/instances", mw.readRequired, handler.TracingListInstancesSW)
	v1.POST("/tracing/traces/query", mw.readRequired, handler.TracingQueryTracesSW)
	v1.GET("/tracing/traces/:id/spans", mw.readRequired, handler.TracingGetTraceSpansSW)
	v1.GET("/tracing/topology", mw.readRequired, handler.TracingGetTopologySW)
}
