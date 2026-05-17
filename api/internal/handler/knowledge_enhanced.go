package handler

import (
	"net/http"
	"time"

	"ai-workbench-api/internal/embedding"
	"ai-workbench-api/internal/knowledge"

	"github.com/gin-gonic/gin"
)

// --- HyDE 增强搜索 ---

type hydeSearchRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k"`
}

// SearchWithHyDE POST /api/v1/knowledge/search/hyde
func SearchWithHyDE(c *gin.Context) {
	var req hydeSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TopK <= 0 {
		req.TopK = 10
	}

	searcher := embedding.GetSearcher()
	result, err := searcher.SearchWithHyDE(c.Request.Context(), req.Query, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 应用反馈调整
	if result != nil && len(result.SearchResults) > 0 {
		fl := embedding.GetFeedbackLearner()
		result.SearchResults = fl.AdjustScores(req.Query, result.SearchResults)
	}

	c.JSON(http.StatusOK, result)
}

// --- 多查询检索 ---

type multiQueryRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k"`
}

// SearchMultiQuery POST /api/v1/knowledge/search/multi
func SearchMultiQuery(c *gin.Context) {
	var req multiQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TopK <= 0 {
		req.TopK = 10
	}

	searcher := embedding.GetSearcher()
	result, err := searcher.SearchMultiQuery(c.Request.Context(), req.Query, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 应用反馈调整
	if result != nil && len(result.SearchResults) > 0 {
		fl := embedding.GetFeedbackLearner()
		result.SearchResults = fl.AdjustScores(req.Query, result.SearchResults)
	}

	c.JSON(http.StatusOK, result)
}

// --- 反馈 ---

type feedbackRequest struct {
	Query  string `json:"query" binding:"required"`
	DocID  string `json:"doc_id" binding:"required"`
	Type   string `json:"type" binding:"required"` // like, dislike, click, skip
	UserID string `json:"user_id"`
}

// SubmitKnowledgeFeedback POST /api/v1/knowledge/feedback
func SubmitKnowledgeFeedback(c *gin.Context) {
	var req feedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 校验反馈类型
	fbType := embedding.FeedbackType(req.Type)
	switch fbType {
	case embedding.FeedbackLike, embedding.FeedbackDislike, embedding.FeedbackClick, embedding.FeedbackSkip:
		// valid
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feedback type, must be: like, dislike, click, skip"})
		return
	}

	fl := embedding.GetFeedbackLearner()
	fl.Record(embedding.FeedbackEntry{
		Query:     req.Query,
		DocID:     req.DocID,
		Type:      fbType,
		UserID:    req.UserID,
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetFeedbackStats GET /api/v1/knowledge/feedback/stats
func GetFeedbackStats(c *gin.Context) {
	fl := embedding.GetFeedbackLearner()
	stats := fl.GetStats()
	c.JSON(http.StatusOK, stats)
}

// --- 知识图谱 ---

// QueryGraph GET /api/v1/knowledge/graph
func QueryGraph(c *gin.Context) {
	entityID := c.Query("entity_id")
	entityType := c.Query("type")

	graph := knowledge.GetGraph()

	if entityID != "" {
		// 查询实体邻居
		entities, relations := graph.QueryNeighbors(entityID, 2)
		c.JSON(http.StatusOK, gin.H{
			"entity_id": entityID,
			"neighbors": entities,
			"relations": relations,
		})
		return
	}

	// 列出实体
	entities := graph.ListEntities(knowledge.EntityType(entityType))
	stats := graph.Stats()
	c.JSON(http.StatusOK, gin.H{
		"entities": entities,
		"stats":    stats,
	})
}

// ExtractGraphEntities POST /api/v1/knowledge/graph/extract
type graphExtractRequest struct {
	DocID   string `json:"doc_id" binding:"required"`
	Title   string `json:"title"`
	Content string `json:"content" binding:"required"`
}

func ExtractGraphEntities(c *gin.Context) {
	var req graphExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	graph := knowledge.GetGraph()
	entities, relations := graph.ExtractFromDocument(req.DocID, req.Title, req.Content)

	c.JSON(http.StatusOK, gin.H{
		"entities_extracted":  len(entities),
		"relations_extracted": len(relations),
		"entities":            entities,
		"relations":           relations,
	})
}

// FindGraphPath GET /api/v1/knowledge/graph/path
func FindGraphPath(c *gin.Context) {
	source := c.Query("source")
	target := c.Query("target")
	if source == "" || target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source and target required"})
		return
	}

	graph := knowledge.GetGraph()
	path := graph.FindPath(source, target)
	if path == nil {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": "no path found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"found":     true,
		"path":      path,
	})
}

// GraphRAGSearch POST /api/v1/knowledge/graph/search
type graphSearchRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k"`
}

func GraphRAGSearch(c *gin.Context) {
	var req graphSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TopK <= 0 {
		req.TopK = 10
	}

	graph := knowledge.GetGraph()
	results := graph.GraphRAG(c.Request.Context(), req.Query, req.TopK)

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"results": results,
	})
}

// --- 知识库统计 ---

// KnowledgeEnhancedStats GET /api/v1/knowledge/stats
func KnowledgeEnhancedStats(c *gin.Context) {
	graph := knowledge.GetGraph()
	graphStats := graph.Stats()

	fl := embedding.GetFeedbackLearner()
	fbStats := fl.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"graph": graphStats,
		"feedback": gin.H{
			"total":        fbStats.TotalFeedback,
			"by_type":      fbStats.ByType,
			"top_liked":    fbStats.TopLiked,
			"top_disliked": fbStats.TopDisliked,
		},
	})
}
