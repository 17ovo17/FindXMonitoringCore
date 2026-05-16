package handler

import (
	"net/http"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// ListAlertMutes returns all alert mute rules.
func ListAlertMutes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": store.ListAlertMutes()})
}

// CreateAlertMute creates a new alert mute rule.
func CreateAlertMute(c *gin.Context) {
	var mute model.AlertMute
	if err := c.ShouldBindJSON(&mute); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert mute payload"})
		return
	}
	mute.CreatedBy = alertMuteActor(c)
	saved, err := store.CreateAlertMute(mute)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, saved)
}

// UpdateAlertMute updates an existing alert mute rule.
func UpdateAlertMute(c *gin.Context) {
	var mute model.AlertMute
	if err := c.ShouldBindJSON(&mute); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert mute payload"})
		return
	}
	id := c.Param("id")
	saved, err := store.UpdateAlertMute(id, mute)
	if err == store.ErrAlertMuteNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert mute not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, saved)
}

// DeleteAlertMute removes an alert mute rule by ID.
func DeleteAlertMute(c *gin.Context) {
	id := c.Param("id")
	err := store.DeleteAlertMute(id)
	if err == store.ErrAlertMuteNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert mute not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func alertMuteActor(c *gin.Context) string {
	if user := c.GetString("username"); user != "" {
		return user
	}
	return "system"
}
