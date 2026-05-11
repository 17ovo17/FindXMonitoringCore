package handler

import (
	"encoding/json"
	"fmt"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func mergeNotificationChannelSecrets(next, existing NotificationChannel) NotificationChannel {
	if next.Webhook == "" || isSecretPlaceholder(next.Webhook) {
		next.Webhook = existing.Webhook
	}
	if next.Endpoint == "" || isSecretPlaceholder(next.Endpoint) {
		next.Endpoint = existing.Endpoint
	}
	if next.Secret == "" || isSecretPlaceholder(next.Secret) {
		next.Secret = existing.Secret
	}
	return next
}

func isSecretPlaceholder(value string) bool {
	value = strings.TrimSpace(value)
	return value == "<SECRET>" || value == "******" || value == "********"
}

func redactNotificationChannel(ch NotificationChannel) NotificationChannel {
	ch.Secret = ""
	if ch.Webhook != "" {
		ch.Webhook = "<SECRET>"
	}
	if ch.Endpoint != "" {
		ch.Endpoint = "<SECRET>"
	}
	return ch
}

func getNotificationChannel(id string) (NotificationChannel, bool) {
	notificationChannelsMu.RLock()
	defer notificationChannelsMu.RUnlock()
	for _, ch := range notificationChannels {
		if ch.ID == id {
			return ch, true
		}
	}
	return NotificationChannel{}, false
}

func decorateNotificationRules(items []model.NotificationRule) []model.NotificationRule {
	out := make([]model.NotificationRule, 0, len(items))
	for _, item := range items {
		out = append(out, decorateNotificationRule(item))
	}
	return out
}

func decorateNotificationRule(rule model.NotificationRule) model.NotificationRule {
	for i, cfg := range rule.NotifyConfigs {
		if ch, ok := getNotificationChannel(cfg.ChannelID); ok {
			rule.NotifyConfigs[i].Channel = ch.Name
		}
	}
	return rule
}

func bindArrayOrSingle[T any](c *gin.Context, out *[]T) error {
	var raw json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		return err
	}
	trimmed := strings.TrimSpace(string(raw))
	if strings.HasPrefix(trimmed, "[") {
		return json.Unmarshal(raw, out)
	}
	var item T
	if err := json.Unmarshal(raw, &item); err != nil {
		return err
	}
	*out = []T{item}
	return nil
}

func parseIDsPayload(c *gin.Context) []string {
	id := strings.TrimSpace(c.Param("id"))
	if id != "" {
		return []string{id}
	}
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil
	}
	return req.IDs
}

func notificationActor(c *gin.Context) string {
	if value, ok := c.Get("username"); ok {
		if text := fmt.Sprint(value); text != "" {
			return text
		}
	}
	return "system"
}

func notificationAlertIDSet(ids []string) map[string]bool {
	out := map[string]bool{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			out[id] = true
		}
	}
	return out
}

func firstNotificationPreviewEvent(ids []string) map[string]any {
	events := store.ListMonitorAlertEvents(true)
	for _, id := range ids {
		for _, event := range events {
			if event.ID == id {
				return notificationEventMap(event)
			}
		}
	}
	if len(events) > 0 {
		return notificationEventMap(events[0])
	}
	return map[string]any{
		"name":        "FindX notification preview",
		"severity":    "info",
		"status":      "firing",
		"value":       "preview",
		"labels":      map[string]any{},
		"annotations": map[string]any{},
	}
}

func notificationEventMap(event model.MonitorAlertEvent) map[string]any {
	return map[string]any{
		"id":           event.ID,
		"name":         event.Name,
		"severity":     event.Severity,
		"status":       event.Status,
		"value":        event.Value,
		"rule_id":      event.RuleID,
		"target_ident": event.TargetIdent,
		"labels":       event.Labels,
		"annotations":  event.Annotations,
	}
}
