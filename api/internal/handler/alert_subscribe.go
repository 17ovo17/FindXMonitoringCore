package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ListAlertSubscribes returns all subscriptions, optionally filtered by user_id.
func ListAlertSubscribes(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("user_id"))
	var items []model.AlertSubscribe
	if userID != "" {
		items = store.ListAlertSubscribesByUser(userID)
	} else {
		items = store.ListAlertSubscribes()
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

// CreateAlertSubscribe creates a new alert subscription.
func CreateAlertSubscribe(c *gin.Context) {
	var sub model.AlertSubscribe
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert subscribe payload"})
		return
	}
	created, err := store.CreateAlertSubscribe(sub)
	if err != nil {
		log.WithError(err).Warn("alert subscribe creation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert subscribe requires name, user_id, and channel_ids"})
		return
	}
	c.JSON(http.StatusOK, created)
}

// UpdateAlertSubscribe updates an existing alert subscription.
func UpdateAlertSubscribe(c *gin.Context) {
	id := c.Param("id")
	var sub model.AlertSubscribe
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert subscribe payload"})
		return
	}
	updated, err := store.UpdateAlertSubscribe(id, sub)
	if err == store.ErrAlertSubscribeNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert subscribe not found"})
		return
	}
	if err != nil {
		log.WithError(err).Warn("alert subscribe update failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert subscribe requires name and channel_ids"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// DeleteAlertSubscribe removes an alert subscription.
func DeleteAlertSubscribe(c *gin.Context) {
	id := c.Param("id")
	if err := store.DeleteAlertSubscribe(id); err == store.ErrAlertSubscribeNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert subscribe not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
