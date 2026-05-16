package main

import (
	"ai-workbench-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerAIOpsEnhancedRoutes(v1 *gin.RouterGroup, mw routeMiddleware) {
	// G1: Tool Registry
	v1.GET("/ai/tools", mw.readRequired, handler.AIToolsList)
	v1.POST("/ai/tools/:name/execute", mw.adminRequired, handler.AIToolExecute)

	// G4: Session persistence
	v1.GET("/ai/sessions", mw.readRequired, handler.AISessionList)
	v1.GET("/ai/sessions/:id", mw.readRequired, handler.AISessionGet)
	v1.POST("/ai/sessions", mw.readRequired, handler.AISessionCreate)
	v1.DELETE("/ai/sessions/:id", mw.adminRequired, handler.AISessionDelete)

	// G8: Knowledge self-learning
	v1.POST("/ai/learn", mw.adminRequired, handler.AILearnSubmit)
	v1.GET("/ai/learned", mw.readRequired, handler.AILearnList)

	// G6: Feishu Bot Gateway
	v1.POST("/bot/feishu/webhook", handler.BotFeishuWebhook)

	// G7: WeCom Bot Gateway
	v1.POST("/bot/wecom/webhook", handler.BotWeComWebhook)
	v1.GET("/bot/wecom/webhook", handler.BotWeComVerify)
}
