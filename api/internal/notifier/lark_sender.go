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

// sendLarkReal 通过 Lark（国际版飞书）Webhook 发送告警通知。
// POST webhook_url
// Body: {"msg_type":"interactive","card":{...}}
func sendLarkReal(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	url := firstNonEmpty(ch.Webhook, ch.Endpoint)
	if url == "" {
		return fmt.Errorf("lark channel missing webhook URL")
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
				"template": larkHeaderColor(event.Severity),
			},
			"elements": []map[string]any{
				{
					"tag":     "markdown",
					"content": content,
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("lark marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("lark request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warnf("notifier: lark send failed channel=%s", ch.Name)
		// 重试 1 次
		resp, err = httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("lark send (retry): %w", err)
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("lark returned status %d", resp.StatusCode)
	}
	log.Infof("alert event dispatched: channel=%s type=lark name=%q", ch.Name, event.Name)
	return nil
}

func larkHeaderColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical", "p0":
		return "red"
	case "warning", "p1":
		return "orange"
	case "info", "p2":
		return "blue"
	default:
		return "grey"
	}
}
