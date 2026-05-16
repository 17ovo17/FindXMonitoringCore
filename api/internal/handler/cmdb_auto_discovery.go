package handler

import (
	"net/http"
	"sort"
	"sync"
	"time"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// cmdbDiscoveryRule 自动发现规则
type cmdbDiscoveryRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Target      string `json:"target"`
	Schedule    string `json:"schedule"`
	Enabled     bool   `json:"enabled"`
	LastRunAt   string `json:"last_run_at,omitempty"`
	LastRunStatus string `json:"last_run_status,omitempty"`
	Creator     string `json:"creator"`
	CreatedAt   string `json:"created_at"`
}

// cmdbDiscoveryResult 发现结果
type cmdbDiscoveryResult struct {
	ID           string         `json:"id"`
	RuleID       string         `json:"rule_id"`
	ResourceType string         `json:"resource_type"`
	ResourceName string         `json:"resource_name"`
	Data         map[string]any `json:"data"`
	Status       string         `json:"status"`
	ApprovedBy   string         `json:"approved_by,omitempty"`
	ApprovedAt   string         `json:"approved_at,omitempty"`
	DiscoveredAt string         `json:"discovered_at"`
}

var (
	discoveryMu      sync.RWMutex
	discoveryRules   []cmdbDiscoveryRule
	discoveryResults []cmdbDiscoveryResult
)

// CmdbListDiscoveryRules 列出发现规则
func CmdbListDiscoveryRules(c *gin.Context) {
	discoveryMu.RLock()
	out := make([]cmdbDiscoveryRule, len(discoveryRules))
	copy(out, discoveryRules)
	discoveryMu.RUnlock()

	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
}

// CmdbCreateDiscoveryRule 创建发现规则
func CmdbCreateDiscoveryRule(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Target   string `json:"target" binding:"required"`
		Schedule string `json:"schedule"`
		Enabled  bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: name, type, target 必填"})
		return
	}

	schedule := req.Schedule
	if schedule == "" {
		schedule = "0 */6 * * *"
	}

	rule := cmdbDiscoveryRule{
		ID:        "rule-" + store.NewID(),
		Name:      req.Name,
		Type:      req.Type,
		Target:    req.Target,
		Schedule:  schedule,
		Enabled:   req.Enabled,
		Creator:   requestActor(c),
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	discoveryMu.Lock()
	discoveryRules = append(discoveryRules, rule)
	discoveryMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"rule_id": rule.ID,
		"name":    rule.Name,
		"type":    rule.Type,
	}).Info("cmdb: discovery rule created")

	c.JSON(http.StatusCreated, rule)
}

// CmdbRunDiscoveryRule 触发发现规则执行
func CmdbRunDiscoveryRule(c *gin.Context) {
	ruleID := c.Param("id")

	discoveryMu.Lock()
	var found bool
	for i := range discoveryRules {
		if discoveryRules[i].ID == ruleID {
			found = true
			discoveryRules[i].LastRunAt = time.Now().Format(time.RFC3339)
			discoveryRules[i].LastRunStatus = "running"
			break
		}
	}
	discoveryMu.Unlock()

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "发现规则不存在"})
		return
	}

	// 模拟发现结果生成
	result := cmdbDiscoveryResult{
		ID:           "result-" + store.NewID(),
		RuleID:       ruleID,
		ResourceType: "operating_system",
		ResourceName: "discovered-host-" + store.NewID()[:8],
		Data: map[string]any{
			"ip":         "192.168.1." + store.NewID()[:2],
			"hostname":   "host-discovered",
			"os_version": "Linux 5.15",
		},
		Status:       "pending",
		DiscoveredAt: time.Now().Format(time.RFC3339),
	}

	discoveryMu.Lock()
	discoveryResults = append(discoveryResults, result)
	// 更新规则状态
	for i := range discoveryRules {
		if discoveryRules[i].ID == ruleID {
			discoveryRules[i].LastRunStatus = "completed"
			break
		}
	}
	discoveryMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"rule_id":   ruleID,
		"result_id": result.ID,
	}).Info("cmdb: discovery rule executed")

	c.JSON(http.StatusOK, gin.H{
		"message":    "发现任务已执行",
		"rule_id":    ruleID,
		"discovered": 1,
		"results":    []cmdbDiscoveryResult{result},
	})
}

// CmdbListDiscoveryResults 列出发现结果
func CmdbListDiscoveryResults(c *gin.Context) {
	status := c.DefaultQuery("status", "")

	discoveryMu.RLock()
	var out []cmdbDiscoveryResult
	for _, r := range discoveryResults {
		if status == "" || r.Status == status {
			out = append(out, r)
		}
	}
	discoveryMu.RUnlock()

	if out == nil {
		out = []cmdbDiscoveryResult{}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DiscoveredAt > out[j].DiscoveredAt })
	c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
}

// CmdbApproveDiscoveryResult 审批发现结果
func CmdbApproveDiscoveryResult(c *gin.Context) {
	resultID := c.Param("id")
	actor := requestActor(c)

	discoveryMu.Lock()
	var target *cmdbDiscoveryResult
	for i := range discoveryResults {
		if discoveryResults[i].ID == resultID {
			target = &discoveryResults[i]
			break
		}
	}

	if target == nil {
		discoveryMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "发现结果不存在"})
		return
	}

	if target.Status != "pending" {
		discoveryMu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": "该发现结果已处理", "status": target.Status})
		return
	}

	target.Status = "approved"
	target.ApprovedBy = actor
	target.ApprovedAt = time.Now().Format(time.RFC3339)
	approved := *target
	discoveryMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"result_id": resultID,
		"actor":     actor,
	}).Info("cmdb: discovery result approved")

	c.JSON(http.StatusOK, gin.H{
		"message": "发现结果已审批通过",
		"result":  approved,
	})
}
