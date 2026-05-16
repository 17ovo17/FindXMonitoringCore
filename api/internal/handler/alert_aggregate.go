package handler

import (
	"fmt"
	"net/http"
	"time"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// AlertAggregateGroup 表示一组聚合后的告警事件。
type AlertAggregateGroup struct {
	RuleID    uint      `json:"rule_id"`
	RuleName  string    `json:"rule_name"`
	Severity  int       `json:"severity"`
	Status    string    `json:"status"`
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	SampleID  uint      `json:"sample_id"`
	Labels    string    `json:"labels"`
}

// ListAggregatedAlertEvents 返回按 rule + target 聚合的告警事件。
// 同一规则在 5 分钟窗口内的事件合并为一组，减少通知噪音。
// GET /api/v1/alert-events/aggregated?status=&window_minutes=5
func ListAggregatedAlertEvents(c *gin.Context) {
	if !store.GormOK() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
		return
	}

	status := c.Query("status")
	windowStr := c.DefaultQuery("window_minutes", "5")
	var windowMinutes int
	if _, err := fmt.Sscanf(windowStr, "%d", &windowMinutes); err != nil || windowMinutes <= 0 {
		windowMinutes = 5
	}

	// 查询最近 24 小时的事件用于聚合
	since := time.Now().Add(-24 * time.Hour)

	db := store.GetDB()
	query := db.Table("alert_events").
		Select(`rule_id, rule_name, severity, status,
			COUNT(*) as count,
			MIN(created_at) as first_seen,
			MAX(created_at) as last_seen,
			MAX(id) as sample_id,
			labels`).
		Where("created_at >= ?", since).
		Group(fmt.Sprintf("rule_id, rule_name, severity, status, labels, FLOOR(UNIX_TIMESTAMP(created_at) / %d)", windowMinutes*60))

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query = query.Order("last_seen DESC")

	var groups []AlertAggregateGroup
	if err := query.Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  len(groups),
	})
}
