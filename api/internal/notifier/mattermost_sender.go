package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"

	log "github.com/sirupsen/logrus"
)

// sendMattermostReal 通过 Mattermost Incoming Webhook 发送告警通知。
// POST webhook_url
// Body: {"text":"...", "channel":"...", "username":"FindX Alert"}
func sendMattermostReal(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	url := firstNonEmpty(ch.Webhook, ch.Endpoint)
	if url == "" {
		return fmt.Errorf("mattermost channel missing webhook URL")
	}

	text := buildMattermostText(event)
	payload := map[string]any{
		"text":     text,
		"username": "FindX Alert",
	}
	// 如果配置了接收方（channel），则指定
	if ch.Receiver != "" {
		payload["channel"] = strings.TrimSpace(ch.Receiver)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mattermost marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("mattermost request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warnf("notifier: mattermost send failed channel=%s", ch.Name)
		// 重试 1 次
		resp, err = httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("mattermost send (retry): %w", err)
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("mattermost returned status %d", resp.StatusCode)
	}
	log.Infof("alert event dispatched: channel=%s type=mattermost name=%q", ch.Name, event.Name)
	return nil
}

func buildMattermostText(event *model.MonitorAlertEvent) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### [%s] %s\n", strings.ToUpper(event.Severity), event.Name))
	sb.WriteString(fmt.Sprintf("| 字段 | 值 |\n|---|---|\n"))
	sb.WriteString(fmt.Sprintf("| 状态 | %s |\n", event.Status))
	if event.TargetIdent != "" {
		sb.WriteString(fmt.Sprintf("| 目标 | %s |\n", event.TargetIdent))
	}
	if event.Value != "" {
		sb.WriteString(fmt.Sprintf("| 当前值 | %s |\n", event.Value))
	}
	if desc, ok := event.Annotations["content"]; ok && desc != "" {
		sb.WriteString(fmt.Sprintf("| 详情 | %s |\n", desc))
	}
	return sb.String()
}
