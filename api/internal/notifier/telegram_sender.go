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

// sendTelegramReal 通过 Telegram Bot API 发送告警通知。
// POST https://api.telegram.org/bot{token}/sendMessage
func sendTelegramReal(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	botToken := firstNonEmpty(ch.Secret, ch.Endpoint)
	if botToken == "" {
		return fmt.Errorf("telegram channel missing bot token")
	}
	chatID := firstNonEmpty(ch.Receiver)
	if chatID == "" {
		return fmt.Errorf("telegram channel missing chat_id")
	}

	text := buildTelegramText(event)
	payload := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram marshal: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", strings.TrimSpace(botToken))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warnf("notifier: telegram send failed channel=%s", ch.Name)
		// 重试 1 次
		resp, err = httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("telegram send (retry): %w", err)
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}
	log.Infof("alert event dispatched: channel=%s type=telegram name=%q", ch.Name, event.Name)
	return nil
}

func buildTelegramText(event *model.MonitorAlertEvent) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>[%s] %s</b>\n", strings.ToUpper(event.Severity), event.Name))
	sb.WriteString(fmt.Sprintf("状态: %s\n", event.Status))
	if event.TargetIdent != "" {
		sb.WriteString(fmt.Sprintf("目标: %s\n", event.TargetIdent))
	}
	if event.Value != "" {
		sb.WriteString(fmt.Sprintf("当前值: %s\n", event.Value))
	}
	if desc, ok := event.Annotations["content"]; ok && desc != "" {
		sb.WriteString(fmt.Sprintf("详情: %s\n", desc))
	}
	return sb.String()
}
