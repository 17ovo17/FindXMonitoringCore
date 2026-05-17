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

// sendFeishuCardReal 通过飞书 Webhook 发送交互卡片（带按钮）。
// POST webhook_url
// Body: {"msg_type":"interactive","card":{"header":{...},"elements":[...]}}
func sendFeishuCardReal(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	url := firstNonEmpty(ch.Webhook, ch.Endpoint)
	if url == "" {
		return fmt.Errorf("feishucard channel missing webhook URL")
	}

	title := fmt.Sprintf("[%s] %s", strings.ToUpper(event.Severity), event.Name)
	content := buildPlainContent(event)

	payload := map[string]any{
		"msg_type": "interactive",
		"card": map[string]any{
			"header": map[string]any{
				"title": map[string]string{
					"tag":     "plain_text",
					"content": title,
				},
				"template": feishuHeaderColor(event.Severity),
			},
			"elements": []any{
				map[string]any{
					"tag":     "markdown",
					"content": content,
				},
				map[string]any{
					"tag": "action",
					"actions": []map[string]any{
						{
							"tag": "button",
							"text": map[string]string{
								"tag":     "plain_text",
								"content": "AI 分析",
							},
							"type":  "primary",
							"value": map[string]string{"action": "ai_analyze", "event_id": event.ID},
						},
						{
							"tag": "button",
							"text": map[string]string{
								"tag":     "plain_text",
								"content": "屏蔽 1h",
							},
							"type":  "danger",
							"value": map[string]string{"action": "mute_1h", "event_id": event.ID},
						},
						{
							"tag": "button",
							"text": map[string]string{
								"tag":     "plain_text",
								"content": "确认",
							},
							"type":  "default",
							"value": map[string]string{"action": "acknowledge", "event_id": event.ID},
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("feishucard marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("feishucard request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warnf("notifier: feishucard send failed channel=%s", ch.Name)
		// 重试 1 次
		resp, err = httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("feishucard send (retry): %w", err)
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("feishucard returned status %d", resp.StatusCode)
	}
	log.Infof("alert event dispatched: channel=%s type=feishucard name=%q", ch.Name, event.Name)
	return nil
}
