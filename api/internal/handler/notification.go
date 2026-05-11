package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type NotificationChannel struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Endpoint string `json:"endpoint,omitempty"`
	Receiver string `json:"receiver,omitempty"`
	Secret   string `json:"secret,omitempty"`
	Webhook  string `json:"webhook"`
	Enabled  bool   `json:"enabled"`
}

var (
	notificationChannels   []NotificationChannel
	notificationChannelsMu sync.RWMutex
)

func ListNotificationChannels(c *gin.Context) {
	notificationChannelsMu.RLock()
	defer notificationChannelsMu.RUnlock()
	items := make([]NotificationChannel, 0, len(notificationChannels))
	for _, ch := range notificationChannels {
		items = append(items, redactNotificationChannel(ch))
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
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
	if ch.ID == "" {
		ch.ID = fmt.Sprintf("nc_%d", time.Now().UnixNano())
	}
	if ch.Webhook == "" {
		ch.Webhook = ch.Endpoint
	}
	if ch.Endpoint == "" {
		ch.Endpoint = ch.Webhook
	}

	notificationChannelsMu.Lock()
	found := false
	for i, existing := range notificationChannels {
		if existing.ID == ch.ID {
			ch = mergeNotificationChannelSecrets(ch, existing)
			notificationChannels[i] = ch
			found = true
			break
		}
	}
	if !found {
		notificationChannels = append(notificationChannels, ch)
	}
	notificationChannelsMu.Unlock()

	c.JSON(http.StatusOK, redactNotificationChannel(ch))
}

func DeleteNotificationChannel(c *gin.Context) {
	id := c.Param("id")
	notificationChannelsMu.Lock()
	defer notificationChannelsMu.Unlock()
	for i, ch := range notificationChannels {
		if ch.ID == id {
			notificationChannels = append(notificationChannels[:i], notificationChannels[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func SendNotification(title, content, level string) {
	notificationChannelsMu.RLock()
	defer notificationChannelsMu.RUnlock()
	for _, ch := range notificationChannels {
		if !ch.Enabled {
			continue
		}
		go sendToChannel(ch, title, content, level)
	}
}

func sendToChannel(ch NotificationChannel, title, content, level string) {
	payload := buildChannelPayload(ch.Type, title, content, level)
	if payload == nil {
		return
	}
	resp, err := http.Post(ch.Webhook, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.WithError(err).Warnf("notification: send to %s(%s) failed", ch.Name, ch.Type)
		return
	}
	resp.Body.Close()
}

func buildChannelPayload(chType, title, content, level string) []byte {
	var payload []byte
	switch chType {
	case "dingtalk":
		payload, _ = json.Marshal(map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": title,
				"text":  fmt.Sprintf("### %s\n%s\n> 级别: %s", title, content, level),
			},
		})
	case "feishu":
		payload, _ = json.Marshal(map[string]any{
			"msg_type": "text",
			"content":  map[string]string{"text": fmt.Sprintf("[%s] %s\n%s", level, title, content)},
		})
	case "wecom":
		payload, _ = json.Marshal(map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"content": fmt.Sprintf("### %s\n%s\n> 级别: %s", title, content, level),
			},
		})
	default:
		return nil
	}
	return payload
}
