package handler

import (
	"net/http"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ExportMonitorAlertRules 导出选中的告警规则为 JSON
func ExportMonitorAlertRules(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids 参数必填"})
		return
	}

	rules := make([]model.MonitorAlertRule, 0, len(req.IDs))
	for _, id := range req.IDs {
		rule, ok := store.GetMonitorAlertRule(id)
		if ok {
			rules = append(rules, *rule)
		}
	}

	logrus.WithFields(logrus.Fields{
		"requested": len(req.IDs),
		"exported":  len(rules),
		"action":    "alert_rules_export",
	}).Info("monitor: alert rules exported")

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"total": len(rules),
	})
}

// ImportMonitorAlertRules 导入 JSON 告警规则
func ImportMonitorAlertRules(c *gin.Context) {
	var req struct {
		Rules []model.MonitorAlertRule `json:"rules"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Rules) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rules 参数必填"})
		return
	}

	actor := requestActor(c)
	imported := 0
	var errors []string

	for i := range req.Rules {
		rule := &req.Rules[i]
		// 清空 ID 以创建新规则
		rule.ID = ""
		if _, err := store.SaveMonitorAlertRule(rule, actor); err != nil {
			errors = append(errors, rule.Name+": "+err.Error())
		} else {
			imported++
		}
	}

	logrus.WithFields(logrus.Fields{
		"total":    len(req.Rules),
		"imported": imported,
		"errors":   len(errors),
		"action":   "alert_rules_import",
	}).Info("monitor: alert rules imported")

	result := gin.H{
		"imported": imported,
		"total":    len(req.Rules),
	}
	if len(errors) > 0 {
		result["errors"] = errors
	}
	c.JSON(http.StatusOK, result)
}
