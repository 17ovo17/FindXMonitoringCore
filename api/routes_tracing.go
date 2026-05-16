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
	v1.GET("/tracing/traces/:id/spans", mw.readRequired, handler.TracingTraceSpans)
	v1.GET("/tracing/topology", mw.readRequired, handler.TracingGetTopologySW)

	// D1: Service detail
	v1.GET("/tracing/services/:name", mw.readRequired, handler.TracingServiceDetail)
	v1.GET("/tracing/services/:name/instances", mw.readRequired, handler.TracingServiceInstances)
	v1.GET("/tracing/services/:name/endpoints", mw.readRequired, handler.TracingServiceEndpoints)

	// D2: Topology metrics
	v1.GET("/tracing/topology/:service/metrics", mw.readRequired, handler.TracingTopologyMetrics)
	v1.GET("/tracing/topology/:service/dependencies", mw.readRequired, handler.TracingTopologyDependencies)

	// D4: Profiling
	v1.GET("/tracing/profiling/tasks", mw.readRequired, handler.TracingListProfilingTasks)
	v1.POST("/tracing/profiling/tasks", mw.readRequired, handler.TracingCreateProfilingTask)
	v1.GET("/tracing/profiling/tasks/:id", mw.readRequired, handler.TracingGetProfilingTask)

	// D5: Service metrics aggregation
	v1.GET("/tracing/services/:name/metrics", mw.readRequired, handler.TracingServiceMetrics)

	v1.GET("/apm/selectors/services", mw.readRequired, handler.TracingListServicesSW)
	v1.GET("/apm/selectors/instances", mw.readRequired, handler.TracingListInstancesSW)
	v1.GET("/apm/selectors/endpoints", mw.readRequired, handler.TracingListEndpointsSW)
	v1.GET("/apm/topology", mw.readRequired, handler.TracingGetTopologySW)
	v1.GET("/apm/traces", mw.readRequired, handler.APMQueryTracesSW)
	v1.POST("/apm/traces", mw.readRequired, handler.APMQueryTracesSW)
	v1.GET("/apm/traces/:traceId", mw.readRequired, handler.APMGetTraceSW)
	v1.GET("/apm/traces/:traceId/spans/:spanId", mw.readRequired, handler.APMGetSpanDetailSW)
	v1.GET("/apm/trace-tags/keys", mw.readRequired, handler.APMTraceTagKeysSW)
	v1.GET("/apm/trace-tags/values", mw.readRequired, handler.APMTraceTagValuesSW)
	v1.GET("/apm/profiling/tasks", mw.readRequired, handler.APMListProfilingTasksSW)
	v1.POST("/apm/profiling/tasks", mw.readRequired, handler.APMCreateProfilingTaskSW)
	v1.POST("/apm/profiling/tasks/:id/cancel", mw.readRequired, handler.APMCancelProfilingTaskSW)
	v1.GET("/apm/alarms", mw.readRequired, handler.APMListAlarmsSW)
	v1.POST("/apm/alarms/:id/ack", mw.readRequired, handler.APMAckAlarmSW)
	v1.GET("/apm/settings", mw.readRequired, handler.APMGetSettingsSW)
	v1.PUT("/apm/settings", mw.readRequired, handler.APMPutSettingsSW)
	v1.GET("/apm/agent-linkage", mw.readRequired, handler.APMAgentLinkageSW)
}
