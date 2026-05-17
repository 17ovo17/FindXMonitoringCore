package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"ai-workbench-api/internal/model"

	log "github.com/sirupsen/logrus"
)

// sendCallbackReal POST 到用户自定义 URL，支持自定义 Headers 和 Body 模板。
// 支持 Go template 变量：{{.RuleName}} {{.Severity}} {{.Labels}} {{.Value}} {{.StartsAt}}
func sendCallbackReal(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	url := firstNonEmpty(ch.Webhook, ch.Endpoint)
	if url == "" {
		return fmt.Errorf("callback channel missing URL")
	}

	// 构建模板数据
	data := callbackTemplateData(event)

	// 渲染 body 模板（如果有自定义模板则使用，否则使用默认 JSON）
	var bodyBytes []byte
	if ch.Secret != "" {
		// Secret 字段复用为 body 模板
		rendered, err := renderCallbackTemplate(ch.Secret, data)
		if err != nil {
			log.Warnf("notifier: callback template render failed: %v, using default", err)
			bodyBytes, _ = json.Marshal(data)
		} else {
			bodyBytes = []byte(rendered)
		}
	} else {
		bodyBytes, _ = json.Marshal(data)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("callback request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warnf("notifier: callback send failed channel=%s", ch.Name)
		// 重试 1 次
		resp, err = httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("callback send (retry): %w", err)
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("callback returned status %d", resp.StatusCode)
	}
	log.Infof("alert event dispatched: channel=%s type=callback name=%q", ch.Name, event.Name)
	return nil
}

// CallbackData 是回调模板可用的变量集合。
type CallbackData struct {
	RuleName string            `json:"rule_name"`
	Severity string            `json:"severity"`
	Status   string            `json:"status"`
	Labels   map[string]string `json:"labels"`
	Value    string            `json:"value"`
	StartsAt string            `json:"starts_at"`
	EndsAt   string            `json:"ends_at"`
	Name     string            `json:"name"`
	Target   string            `json:"target"`
}

func callbackTemplateData(event *model.MonitorAlertEvent) *CallbackData {
	return &CallbackData{
		RuleName: event.Name,
		Severity: event.Severity,
		Status:   event.Status,
		Labels:   event.Labels,
		Value:    event.Value,
		StartsAt: event.FirstSeen.Format(time.RFC3339),
		EndsAt:   event.LastSeen.Format(time.RFC3339),
		Name:     event.Name,
		Target:   event.TargetIdent,
	}
}

func renderCallbackTemplate(tplStr string, data *CallbackData) (string, error) {
	tmpl, err := template.New("callback").Parse(tplStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
