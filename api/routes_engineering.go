package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerEngineeringRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	// J2: 告警聚合
	v1.GET("/alert-events/aggregated", mw.readRequired, handler.ListAggregatedAlertEvents)

	// J3: 值班日历
	handler.RegisterOncallRoutes(v1, mw.readRequired, mw.adminRequired)

	// J4: SLA 报告
	handler.RegisterSLARoutes(v1, mw.readRequired, mw.adminRequired)

	// J5: 容量预测
	v1.GET("/capacity/forecast", mw.readRequired, handler.GetCapacityForecast)
}
