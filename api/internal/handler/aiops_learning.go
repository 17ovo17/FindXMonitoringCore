package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LearnedPattern represents a pattern learned from resolved incidents.
type LearnedPattern struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Symptoms        []string `json:"symptoms"`
	RootCause       string   `json:"root_cause"`
	ResolutionSteps []string `json:"resolution_steps"`
	RelatedMetrics  []string `json:"related_metrics"`
	RelatedLogs     []string `json:"related_logs"`
	Tags            []string `json:"tags"`
	Confidence      float64  `json:"confidence"`
	UsageCount      int      `json:"usage_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedBy       string   `json:"created_by"`
}

var (
	learnedPatternsMu sync.RWMutex
	learnedPatterns   = map[string]*LearnedPattern{}
)

func generateLearnID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "learn-" + hex.EncodeToString(b)
}

// AILearnSubmit handles POST /api/v1/ai/learn
func AILearnSubmit(c *gin.Context) {
	var req struct {
		Title           string   `json:"title"`
		Symptoms        []string `json:"symptoms"`
		RootCause       string   `json:"root_cause"`
		ResolutionSteps []string `json:"resolution_steps"`
		RelatedMetrics  []string `json:"related_metrics"`
		RelatedLogs     []string `json:"related_logs"`
		Tags            []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "title is required"})
		return
	}
	if req.RootCause == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "root_cause is required"})
		return
	}

	now := time.Now()
	pattern := &LearnedPattern{
		ID:              generateLearnID(),
		Title:           req.Title,
		Symptoms:        req.Symptoms,
		RootCause:       req.RootCause,
		ResolutionSteps: req.ResolutionSteps,
		RelatedMetrics:  req.RelatedMetrics,
		RelatedLogs:     req.RelatedLogs,
		Tags:            req.Tags,
		Confidence:      0.8,
		UsageCount:      0,
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       c.GetString("user_id"),
	}

	// Auto-extract tags from symptoms and root cause if not provided
	if len(pattern.Tags) == 0 {
		pattern.Tags = extractLearnTags(pattern)
	}

	learnedPatternsMu.Lock()
	learnedPatterns[pattern.ID] = pattern
	learnedPatternsMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": pattern})
}

// AILearnList handles GET /api/v1/ai/learned
func AILearnList(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))

	learnedPatternsMu.RLock()
	defer learnedPatternsMu.RUnlock()

	patterns := make([]*LearnedPattern, 0, len(learnedPatterns))
	for _, p := range learnedPatterns {
		if keyword != "" && !learnedPatternMatchesKeyword(p, keyword) {
			continue
		}
		patterns = append(patterns, p)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"patterns": patterns, "total": len(patterns)}})
}

// SearchLearnedPatterns searches learned patterns by symptoms (internal use).
func SearchLearnedPatterns(symptoms string) []*LearnedPattern {
	learnedPatternsMu.RLock()
	defer learnedPatternsMu.RUnlock()

	lower := strings.ToLower(symptoms)
	results := make([]*LearnedPattern, 0)
	for _, p := range learnedPatterns {
		for _, s := range p.Symptoms {
			if strings.Contains(lower, strings.ToLower(s)) {
				results = append(results, p)
				break
			}
		}
		if len(results) >= 5 {
			break
		}
	}
	return results
}

func learnedPatternMatchesKeyword(p *LearnedPattern, keyword string) bool {
	lower := strings.ToLower(keyword)
	if strings.Contains(strings.ToLower(p.Title), lower) {
		return true
	}
	if strings.Contains(strings.ToLower(p.RootCause), lower) {
		return true
	}
	for _, s := range p.Symptoms {
		if strings.Contains(strings.ToLower(s), lower) {
			return true
		}
	}
	for _, t := range p.Tags {
		if strings.Contains(strings.ToLower(t), lower) {
			return true
		}
	}
	return false
}

func extractLearnTags(p *LearnedPattern) []string {
	tagSet := map[string]bool{}
	keywords := []string{"cpu", "memory", "disk", "network", "timeout", "oom", "crash", "slow", "error", "connection", "database", "redis", "mysql", "kafka"}
	combined := strings.ToLower(p.Title + " " + p.RootCause + " " + strings.Join(p.Symptoms, " "))
	for _, kw := range keywords {
		if strings.Contains(combined, kw) {
			tagSet[kw] = true
		}
	}
	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}
	return tags
}