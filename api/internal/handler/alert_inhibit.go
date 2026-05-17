package handler

import (
	"net/http"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// ListAlertInhibits 列出所有抑制规则。
// GET /api/v1/alert-inhibits
func ListAlertInhibits(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": store.ListAlertInhibitRules()})
}

// CreateAlertInhibit 创建一条抑制规则。
// POST /api/v1/alert-inhibits
func CreateAlertInhibit(c *gin.Context) {
	var rule model.AlertInhibitRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert inhibit payload"})
		return
	}
	rule.CreatedBy = alertInhibitActor(c)
	saved, err := store.CreateAlertInhibitRule(rule)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, saved)
}

// UpdateAlertInhibit 更新一条抑制规则。
// PUT /api/v1/alert-inhibits/:id
func UpdateAlertInhibit(c *gin.Context) {
	var rule model.AlertInhibitRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert inhibit payload"})
		return
	}
	id := c.Param("id")
	saved, err := store.UpdateAlertInhibitRule(id, rule)
	if err == store.ErrAlertInhibitNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert inhibit rule not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, saved)
}

// DeleteAlertInhibit 删除一条抑制规则。
// DELETE /api/v1/alert-inhibits/:id
func DeleteAlertInhibit(c *gin.Context) {
	id := c.Param("id")
	err := store.DeleteAlertInhibitRule(id)
	if err == store.ErrAlertInhibitNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert inhibit rule not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func alertInhibitActor(c *gin.Context) string {
	if user := c.GetString("username"); user != "" {
		return user
	}
	return "system"
}
