package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/notifier"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// NotificationChannel is kept as a handler-level alias so existing callers
// (tests, helpers) compile while storage lives in the store package.
type NotificationChannel = model.NotificationChannel

func ListNotificationChannels(c *gin.Context) {
	items := store.ListNotificationChannels()
	out := make([]NotificationChannel, 0, len(items))
	for _, ch := range items {
		out = append(out, redactNotificationChannel(ch))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func SaveNotificationChannel(c *gin.Context) {
	var ch NotificationChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification channel payload"})
		return
	}
	if strings.TrimSpace(ch.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification channel name required"})
		return
	}
	now := time.Now()
	if ch.ID == "" {
		ch.ID = fmt.Sprintf("nc_%d", now.UnixNano())
		ch.CreatedAt = now
	}
	if ch.Webhook == "" {
		ch.Webhook = ch.Endpoint
	}
	if ch.Endpoint == "" {
		ch.Endpoint = ch.Webhook
	}
	if existing, ok := store.GetNotificationChannel(ch.ID); ok {
		ch = mergeNotificationChannelSecrets(ch, *existing)
		if ch.CreatedAt.IsZero() {
			ch.CreatedAt = existing.CreatedAt
		}
	}
	ch.UpdatedAt = now
	store.PutNotificationChannel(&ch)
	c.JSON(http.StatusOK, redactNotificationChannel(ch))
}

func DeleteNotificationChannel(c *gin.Context) {
	if store.DeleteNotificationChannel(c.Param("id")) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

// TestNotificationChannel dispatches a mock alert event to a single channel
// and reports whether the delivery succeeded. This lets the UI verify a
// newly created channel without waiting for a real alert to fire.
func TestNotificationChannel(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	ch, ok := store.GetNotificationChannel(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification channel not found"})
		return
	}
	if !ch.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "notification channel is disabled"})
		return
	}
	event := notifier.BuildMockAlertEvent(ch.Name)
	if err := notifier.SendToChannel(ch, event); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"ok":      false,
			"channel": redactNotificationChannel(*ch),
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"channel": redactNotificationChannel(*ch),
		"message": "notification channel delivered mock event",
	})
}

// SendNotification is kept as a convenience wrapper used by legacy
// diagnosis and alert handlers. It broadcasts a text notification to
// every enabled channel using a synthetic MonitorAlertEvent.
func SendNotification(title, content, level string) {
	event := &model.MonitorAlertEvent{
		Name:     title,
		Severity: level,
		Status:   model.MonitorAlertEventStatusFiring,
		Value:    content,
		Annotations: map[string]string{
			"title":   title,
			"content": content,
		},
	}
	for _, ch := range store.ListActiveNotificationChannels() {
		chCopy := ch
		go func(c model.NotificationChannel) {
			if err := notifier.SendToChannel(&c, event); err != nil {
				// SendToChannel already logs; swallow to keep signature compatible.
				_ = err
			}
		}(chCopy)
	}
}
