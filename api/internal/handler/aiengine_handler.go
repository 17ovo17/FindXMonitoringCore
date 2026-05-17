package handler

import (
	"net/http"

	"ai-workbench-api/internal/aiengine"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// AI Engine Handler — 高级 AI 能力模块的 HTTP 端点
// ---------------------------------------------------------------------------
//
// POST /api/v1/ai/investigate       — 启动自主调查
// POST /api/v1/ai/nl2query          — 自然语言转查询
// GET  /api/v1/ai/anomalies         — 获取检测到的异常列表
// POST /api/v1/ai/drilldown         — 维度下钻分析
// GET  /api/v1/ai/incidents         — 事件生命周期列表
// GET  /api/v1/ai/incidents/:id     — 事件详情（含时间线）
// POST /api/v1/ai/incidents/:id/advance — 手动推进事件阶段
// GET  /api/v1/ai/incidents/:id/postmortem — 获取复盘报告

// 全局 AI 引擎模块实例
var (
	globalInvestigator *aiengine.AutonomousInvestigator
	globalDetector     *aiengine.AnomalyDetector
	globalDrilldown    *aiengine.DimensionDrilldown
	globalLifecycle    *aiengine.IncidentLifecycle
	globalNL2Query     *aiengine.NL2Query
)

func init() {
	registry := aiengine.NewToolRegistry()
	globalInvestigator = aiengine.NewAutonomousInvestigator(registry)
	globalDetector = aiengine.NewAnomalyDetector()
	globalDrilldown = aiengine.NewDimensionDrilldown()
	globalLifecycle = aiengine.NewIncidentLifecycle()
	globalNL2Query = aiengine.NewNL2Query()
}

// --- 自主调查 ---

type investigateRequest struct {
	Type    string         `json:"type" binding:"required"` // alert, anomaly, user_query
	Data    map[string]any `json:"data"`
	Context map[string]any `json:"context"`
}

// AIInvestigate handles POST /api/v1/ai/investigate
func AIInvestigate(c *gin.Context) {
	var req investigateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	trigger := aiengine.InvestigationTrigger{
		Type:    req.Type,
		Data:    req.Data,
		Context: req.Context,
	}

	investigation, err := globalInvestigator.Investigate(c.Request.Context(), trigger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": investigation})
}

// --- 自然语言转查询 ---

type nl2queryRequest struct {
	Input          string `json:"input" binding:"required"`
	TargetLanguage string `json:"target_language"` // promql, logql, sql（可选，自动检测）
}

// AINL2Query handles POST /api/v1/ai/nl2query
func AINL2Query(c *gin.Context) {
	var req nl2queryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	targetLang := aiengine.QueryLanguage(req.TargetLanguage)
	result, err := globalNL2Query.Translate(req.Input, targetLang)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// --- 异常检测 ---

// AIAnomaliesList handles GET /api/v1/ai/anomalies
func AIAnomaliesList(c *gin.Context) {
	anomalies := globalDetector.ListAnomalies()
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"anomalies": anomalies, "total": len(anomalies)}})
}

// --- 维度下钻 ---

type drilldownRequest struct {
	Metric       string                          `json:"metric" binding:"required"`
	CurrentData  []aiengine.DimensionDataPoint   `json:"current_data"`
	BaselineData []aiengine.DimensionDataPoint   `json:"baseline_data"`
	TimeRange    *aiengine.TimeRange             `json:"time_range"`
}

// AIDrilldown handles POST /api/v1/ai/drilldown
func AIDrilldown(c *gin.Context) {
	var req drilldownRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	tr := aiengine.TimeRange{}
	if req.TimeRange != nil {
		tr = *req.TimeRange
	}

	result, err := globalDrilldown.Drilldown(req.Metric, req.CurrentData, req.BaselineData, tr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// --- 事件生命周期 ---

// AIIncidentsList handles GET /api/v1/ai/incidents
func AIIncidentsList(c *gin.Context) {
	incidents := globalLifecycle.ListIncidents()
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"incidents": incidents, "total": len(incidents)}})
}

// AIIncidentGet handles GET /api/v1/ai/incidents/:id
func AIIncidentGet(c *gin.Context) {
	id := c.Param("id")
	incident, ok := globalLifecycle.GetIncident(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": "incident not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": incident})
}

// AIIncidentAdvance handles POST /api/v1/ai/incidents/:id/advance
func AIIncidentAdvance(c *gin.Context) {
	id := c.Param("id")
	incident, ok := globalLifecycle.GetIncident(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": "incident not found"})
		return
	}

	if err := globalLifecycle.AdvancePhase(incident); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": incident})
}

// AIIncidentPostMortem handles GET /api/v1/ai/incidents/:id/postmortem
func AIIncidentPostMortem(c *gin.Context) {
	id := c.Param("id")
	incident, ok := globalLifecycle.GetIncident(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": "incident not found"})
		return
	}

	if incident.PostMortem == nil {
		pm, err := globalLifecycle.GeneratePostMortem(incident)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": pm})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": incident.PostMortem})
}
