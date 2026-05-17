package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	log "github.com/sirupsen/logrus"
)

// httpTimeout is applied to outbound webhook deliveries.
const httpTimeout = 10 * time.Second

// httpClient is shared by webhook-style dispatches so we don't
// leak connections on every alert fire.
var httpClient = &http.Client{Timeout: httpTimeout}

// DispatchAlertEvent fans an alert event out to all matching
// notification rules and their configured channels. Sending is
// performed in background goroutines so the caller (scheduler) is
// never blocked by external endpoints.
func DispatchAlertEvent(event *model.MonitorAlertEvent) {
	if event == nil {
		return
	}
	log.Infof("notifier: dispatch called for event name=%q severity=%s", event.Name, event.Severity)
	if model.IsAlertMuted(event) {
		log.Infof("alert event muted: name=%q", event.Name)
		return
	}
	rules := store.ListActiveNotificationRules()
	log.Infof("notifier: active rules count=%d", len(rules))
	for _, rule := range rules {
		if !matchesRule(rule, event) {
			continue
		}
		for _, cfg := range rule.NotifyConfigs {
			id := strings.TrimSpace(cfg.ChannelID)
			if id == "" {
				continue
			}
			ch, ok := store.GetNotificationChannel(id)
			if !ok || ch == nil || !ch.Enabled {
				continue
			}
			channel := *ch
			evCopy := *event
			go sendToChannel(&channel, &evCopy)
		}
	}

	// Dispatch to subscribers
	subs := store.MatchingSubscribes(event)
	for _, sub := range subs {
		for _, chID := range sub.ChannelIDs {
			ch, ok := store.GetNotificationChannel(chID)
			if !ok || ch == nil || !ch.Enabled {
				continue
			}
			channel := *ch
			evCopy := *event
			go sendToChannel(&channel, &evCopy)
		}
	}
}

// matchesRule checks whether a notification rule applies to the given event.
// Rules without alert_rule_ids match all events; otherwise the event's
// RuleID must appear in the rule's AlertRuleIDs. Severity filters are
// honoured when any notify config declares them.
func matchesRule(rule model.NotificationRule, event *model.MonitorAlertEvent) bool {
	if len(rule.AlertRuleIDs) > 0 {
		found := false
		for _, id := range rule.AlertRuleIDs {
			if strings.TrimSpace(id) == event.RuleID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if !matchesSeverity(rule, event.Severity) {
		return false
	}
	return true
}

func matchesSeverity(rule model.NotificationRule, severity string) bool {
	severity = strings.TrimSpace(strings.ToLower(severity))
	hasFilter := false
	for _, cfg := range rule.NotifyConfigs {
		if len(cfg.Severities) == 0 {
			continue
		}
		hasFilter = true
		for _, wanted := range cfg.Severities {
			if strings.ToLower(strings.TrimSpace(wanted)) == severity {
				return true
			}
		}
	}
	return !hasFilter
}

// SendToChannel exposes a single-channel dispatch for test endpoints.
func SendToChannel(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	if ch == nil || event == nil {
		return fmt.Errorf("channel and event required")
	}
	return sendToChannel(ch, event)
}

func sendToChannel(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	switch strings.ToLower(strings.TrimSpace(ch.Type)) {
	case "webhook":
		return sendWebhook(ch, event)
	case "log", "":
		log.Infof("alert event: name=%q severity=%q status=%q rule_id=%s target=%s value=%s channel=%s",
			event.Name, event.Severity, event.Status, event.RuleID, event.TargetIdent, event.Value, ch.Name)
		return nil
	case "dingtalk":
		return sendDingTalk(ch, event)
	case "wecom":
		return sendWeCom(ch, event)
	case "feishu":
		return sendFeishu(ch, event)
	case "email":
		return sendEmail(ch, event)
	case "telegram":
		return sendTelegram(ch, event)
	case "lark":
		return sendLark(ch, event)
	case "feishucard":
		return sendFeishuCard(ch, event)
	case "callback":
		return sendCallback(ch, event)
	case "mattermost":
		return sendMattermost(ch, event)
	default:
		log.Warnf("notifier: unsupported channel type: %s", ch.Type)
		return fmt.Errorf("unsupported channel type: %s", ch.Type)
	}
}

// sendWebhook performs a real HTTP POST with a JSON body describing the event.
// The channel's Webhook (fallback: Endpoint) is used as the destination.
func sendWebhook(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	url := firstNonEmpty(ch.Webhook, ch.Endpoint)
	if url == "" {
		log.Warnf("notifier: webhook channel %s missing URL", ch.Name)
		return fmt.Errorf("webhook channel missing URL")
	}
	payload, err := json.Marshal(webhookPayload(event))
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		// Never echo the URL; it may carry a secret token.
		log.WithError(err).Warnf("notifier: webhook send failed channel=%s", ch.Name)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Warnf("notifier: webhook channel=%s returned status=%d", ch.Name, resp.StatusCode)
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	log.Infof("alert event dispatched: channel=%s type=webhook name=%q", ch.Name, event.Name)
	return nil
}

func webhookPayload(event *model.MonitorAlertEvent) map[string]any {
	return map[string]any{
		"id":           event.ID,
		"name":         event.Name,
		"severity":     event.Severity,
		"status":       event.Status,
		"rule_id":      event.RuleID,
		"target_ident": event.TargetIdent,
		"value":        event.Value,
		"labels":       event.Labels,
		"annotations":  event.Annotations,
		"first_seen":   event.FirstSeen,
		"last_seen":    event.LastSeen,
	}
}

func sendDingTalk(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	return sendDingTalkReal(ch, event)
}

func sendWeCom(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	return sendWeComReal(ch, event)
}

func sendFeishu(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	return sendFeishuReal(ch, event)
}

func sendEmail(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	return sendEmailReal(ch, event)
}

func sendTelegram(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	return sendTelegramReal(ch, event)
}

func sendLark(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	return sendLarkReal(ch, event)
}

func sendFeishuCard(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	return sendFeishuCardReal(ch, event)
}

func sendCallback(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	return sendCallbackReal(ch, event)
}

func sendMattermost(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	return sendMattermostReal(ch, event)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// BuildMockAlertEvent creates an in-memory event suitable for channel test endpoints.
func BuildMockAlertEvent(channelName string) *model.MonitorAlertEvent {
	now := time.Now()
	return &model.MonitorAlertEvent{
		ID:          "test-" + fmt.Sprint(now.UnixNano()),
		Name:        "FindX notification channel test",
		Severity:    model.MonitorAlertSeverityInfo,
		Status:      model.MonitorAlertEventStatusFiring,
		Value:       "1",
		TargetIdent: "findx-test-target",
		Labels: map[string]string{
			"channel": channelName,
			"source":  "manual-test",
		},
		Annotations: map[string]string{
			"reason": "notification channel connectivity test",
		},
		FirstSeen: now,
		LastSeen:  now,
	}
}
