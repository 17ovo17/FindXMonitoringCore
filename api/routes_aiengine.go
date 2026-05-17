package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerAIEngineRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	// 自主调查引擎
	v1.POST("/ai/investigate", mw.readRequired, handler.AIInvestigate)

	// 自然语言转查询
	v1.POST("/ai/nl2query", mw.readRequired, handler.AINL2Query)

	// 异常检测
	v1.GET("/ai/anomalies", mw.readRequired, handler.AIAnomaliesList)

	// 维度下钻
	v1.POST("/ai/drilldown", mw.readRequired, handler.AIDrilldown)

	// 事件生命周期
	v1.GET("/ai/incidents", mw.readRequired, handler.AIIncidentsList)
	v1.GET("/ai/incidents/:id", mw.readRequired, handler.AIIncidentGet)
	v1.POST("/ai/incidents/:id/advance", mw.adminRequired, handler.AIIncidentAdvance)
	v1.GET("/ai/incidents/:id/postmortem", mw.readRequired, handler.AIIncidentPostMortem)
}
