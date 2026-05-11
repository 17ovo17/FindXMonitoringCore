package handler

import (
	"net/http"
	"strconv"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func ListNotificationRules(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": decorateNotificationRules(store.ListNotificationRules())})
}

func GetNotificationRule(c *gin.Context) {
	rule, ok := store.GetNotificationRule(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification rule not found"})
		return
	}
	c.JSON(http.StatusOK, decorateNotificationRule(*rule))
}

func SaveNotificationRules(c *gin.Context) {
	var rules []model.NotificationRule
	if err := bindArrayOrSingle(c, &rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification rule payload"})
		return
	}
	saved, err := store.SaveNotificationRules(rules, notificationActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification rule requires name and notify configs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": decorateNotificationRules(saved)})
}

func UpdateNotificationRule(c *gin.Context) {
	var rule model.NotificationRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification rule payload"})
		return
	}
	rule.ID = c.Param("id")
	saved, err := store.SaveNotificationRule(&rule, notificationActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification rule requires name and notify configs"})
		return
	}
	c.JSON(http.StatusOK, decorateNotificationRule(*saved))
}

func DeleteNotificationRules(c *gin.Context) {
	ids := parseIDsPayload(c)
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": store.DeleteNotificationRules(ids)})
}

func EnableNotificationRule(c *gin.Context) {
	setNotificationRuleEnabled(c, true)
}

func DisableNotificationRule(c *gin.Context) {
	setNotificationRuleEnabled(c, false)
}

func setNotificationRuleEnabled(c *gin.Context, enabled bool) {
	rule, ok, err := store.SetNotificationRuleEnabled(c.Param("id"), enabled, notificationActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification rule cannot be updated"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification rule not found"})
		return
	}
	c.JSON(http.StatusOK, decorateNotificationRule(*rule))
}

func CloneNotificationRule(c *gin.Context) {
	rule, ok, err := store.CloneNotificationRule(c.Param("id"), notificationActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification rule cannot be cloned"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification rule not found"})
		return
	}
	c.JSON(http.StatusOK, decorateNotificationRule(*rule))
}

func TestNotificationRule(c *gin.Context) {
	var req struct {
		EventIDs     []string                 `json:"event_ids"`
		NotifyConfig model.NotificationConfig `json:"notify_config"`
		RuleID       string                   `json:"rule_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification test payload"})
		return
	}
	if req.NotifyConfig.ChannelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notify_config.channel_id required"})
		return
	}
	channel, ok := getNotificationChannel(req.NotifyConfig.ChannelID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification channel not found"})
		return
	}
	if !channel.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "notification channel is disabled"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"dry_run": true,
		"message": "notification contract validated; external delivery is disabled in test endpoint",
		"channel": redactNotificationChannel(channel),
	})
}

func GetNotificationStatistics(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	c.JSON(http.StatusOK, store.NotificationRuleStatistics(c.Param("id"), days))
}

func GetNotificationEvents(c *gin.Context) {
	rule, ok := store.GetNotificationRule(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification rule not found"})
		return
	}
	alertIDs := notificationAlertIDSet(rule.AlertRuleIDs)
	items := []model.MonitorAlertEvent{}
	for _, event := range store.ListMonitorAlertEvents(true) {
		if alertIDs[event.RuleID] {
			items = append(items, event)
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func GetNotificationAlertRules(c *gin.Context) {
	rule, ok := store.GetNotificationRule(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification rule not found"})
		return
	}
	alertIDs := notificationAlertIDSet(rule.AlertRuleIDs)
	items := []model.MonitorAlertRule{}
	for _, alertRule := range store.ListMonitorAlertRules() {
		if alertIDs[alertRule.ID] {
			items = append(items, alertRule)
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func GetNotificationSubAlertRules(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []model.MonitorAlertRule{}, "total": 0})
}
