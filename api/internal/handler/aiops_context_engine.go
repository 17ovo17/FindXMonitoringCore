package handler

import (
	"net/http"

	"ai-workbench-api/internal/aiengine"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// aiengine 集成层 — 将 ContextEngine 接入现有 handler
// ---------------------------------------------------------------------------
//
// 提供 HTTP API 暴露 aiengine 能力：
// - POST /api/v1/ai/context/build — 构建上下文
// - POST /api/v1/ai/intent/classify — 意图分类
// - POST /api/v1/ai/alert/classify — 告警分类
// - GET  /api/v1/ai/trajectory/:id — 获取执行轨迹
// - POST /api/v1/ai/memory/store — 存储记忆
// - GET  /api/v1/ai/memory/recall — 召回记忆

// getAIEngine 获取全局 AI 引擎实例
func getAIEngine() *aiengine.Engine {
	return aiengine.GetEngine()
}

// AIContextBuild 构建 AI 上下文
// POST /api/v1/ai/context/build
func AIContextBuild(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Query     string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	engine := getAIEngine()
	blocks, intent, trajectoryID, err := engine.ProcessQuery(req.SessionID, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"blocks":        blocks,
			"intent":        intent,
			"trajectory_id": trajectoryID,
		},
	})
}

// AIIntentClassify 意图分类
// POST /api/v1/ai/intent/classify
func AIIntentClassify(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	engine := getAIEngine()
	intent := engine.Context.ClassifyIntent(req.Query)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"intent": intent,
			"query":  req.Query,
		},
	})
}

// AIAlertClassify 告警分类
// POST /api/v1/ai/alert/classify
func AIAlertClassify(c *gin.Context) {
	var req struct {
		ID       string            `json:"id" binding:"required"`
		Title    string            `json:"title" binding:"required"`
		Labels   map[string]string `json:"labels"`
		Value    float64           `json:"value"`
		Severity string            `json:"severity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	engine := getAIEngine()
	event := &aiengine.AlertEvent{
		ID:       req.ID,
		Title:    req.Title,
		Labels:   req.Labels,
		Value:    req.Value,
		Severity: req.Severity,
	}
	category := engine.Classifier.Classify(event)
	workflowID := engine.Classifier.RouteToWorkflow(category)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"category":    category,
			"workflow_id": workflowID,
		},
	})
}

// AITrajectoryGet 获取执行轨迹
// GET /api/v1/ai/trajectory/:id
func AITrajectoryGet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "trajectory id required"})
		return
	}

	engine := getAIEngine()
	trajectory := engine.Trajectory.Get(id)
	if trajectory == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": "trajectory not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": trajectory,
	})
}

// AIMemoryStore 存储记忆
// POST /api/v1/ai/memory/store
func AIMemoryStore(c *gin.Context) {
	var req struct {
		Type    string   `json:"type" binding:"required"`
		Content string   `json:"content" binding:"required"`
		Tags    []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	engine := getAIEngine()
	memType := aiengine.MemoryType(req.Type)
	id := engine.Context.GetMemoryManager().Store(memType, req.Content, req.Tags, nil)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"id": id},
	})
}

// AIMemoryRecall 召回记忆
// GET /api/v1/ai/memory/recall?query=xxx&limit=5
func AIMemoryRecall(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "query required"})
		return
	}

	engine := getAIEngine()
	memories := engine.Context.GetMemoryManager().Recall(query, 5)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"memories": memories,
			"total":    len(memories),
		},
	})
}
