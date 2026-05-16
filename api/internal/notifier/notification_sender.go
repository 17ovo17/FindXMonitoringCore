package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	"ai-workbench-api/internal/model"

	log "github.com/sirupsen/logrus"
)

// sendDingTalkReal 发送钉钉机器人 Markdown 消息。
func sendDingTalkReal(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	url := firstNonEmpty(ch.Webhook, ch.Endpoint)
	if url == "" {
		return fmt.Errorf("dingtalk channel missing webhook URL")
	}

	title := fmt.Sprintf("[%s] %s", strings.ToUpper(event.Severity), event.Name)
	content := buildMarkdownContent(event)

	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  content,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("dingtalk marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("dingtalk request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warnf("notifier: dingtalk send failed channel=%s", ch.Name)
		return fmt.Errorf("dingtalk send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("dingtalk returned status %d", resp.StatusCode)
	}
	log.Infof("alert event dispatched: channel=%s type=dingtalk name=%q", ch.Name, event.Name)
	return nil
}

// sendWeComReal 发送企业微信机器人 Markdown 消息。
func sendWeComReal(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	url := firstNonEmpty(ch.Webhook, ch.Endpoint)
	if url == "" {
		return fmt.Errorf("wecom channel missing webhook URL")
	}

	content := buildMarkdownContent(event)
	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("wecom marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("wecom request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warnf("notifier: wecom send failed channel=%s", ch.Name)
		return fmt.Errorf("wecom send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("wecom returned status %d", resp.StatusCode)
	}
	log.Infof("alert event dispatched: channel=%s type=wecom name=%q", ch.Name, event.Name)
	return nil
}

// sendFeishuReal 发送飞书机器人卡片消息。
func sendFeishuReal(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	url := firstNonEmpty(ch.Webhook, ch.Endpoint)
	if url == "" {
		return fmt.Errorf("feishu channel missing webhook URL")
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
			"elements": []map[string]any{
				{
					"tag": "markdown",
					"content": content,
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("feishu marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("feishu request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warnf("notifier: feishu send failed channel=%s", ch.Name)
		return fmt.Errorf("feishu send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("feishu returned status %d", resp.StatusCode)
	}
	log.Infof("alert event dispatched: channel=%s type=feishu name=%q", ch.Name, event.Name)
	return nil
}

// sendEmailReal 通过 SMTP 发送邮件通知。
// 环境变量：SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM
func sendEmailReal(ch *model.NotificationChannel, event *model.MonitorAlertEvent) error {
	smtpHost := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	smtpPort := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	smtpUser := strings.TrimSpace(os.Getenv("SMTP_USER"))
	smtpPass := strings.TrimSpace(os.Getenv("SMTP_PASS"))
	smtpFrom := strings.TrimSpace(os.Getenv("SMTP_FROM"))

	if smtpHost == "" {
		return fmt.Errorf("email channel: SMTP_HOST not configured")
	}
	if smtpPort == "" {
		smtpPort = "25"
	}
	if smtpFrom == "" {
		smtpFrom = smtpUser
	}

	receiver := firstNonEmpty(ch.Receiver, ch.Endpoint)
	if receiver == "" {
		return fmt.Errorf("email channel missing receiver address")
	}

	subject := fmt.Sprintf("[FindX Alert][%s] %s", strings.ToUpper(event.Severity), event.Name)
	body := buildPlainContent(event)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		smtpFrom, receiver, subject, body)

	addr := smtpHost + ":" + smtpPort
	var auth smtp.Auth
	if smtpUser != "" && smtpPass != "" {
		auth = smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	}

	recipients := strings.Split(receiver, ",")
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}

	if err := smtp.SendMail(addr, auth, smtpFrom, recipients, []byte(msg)); err != nil {
		log.WithError(err).Warnf("notifier: email send failed channel=%s", ch.Name)
		return fmt.Errorf("email send: %w", err)
	}
	log.Infof("alert event dispatched: channel=%s type=email name=%q to=%s", ch.Name, event.Name, receiver)
	return nil
}

func buildMarkdownContent(event *model.MonitorAlertEvent) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**告警名称**: %s\n\n", event.Name))
	sb.WriteString(fmt.Sprintf("**严重级别**: %s\n\n", event.Severity))
	sb.WriteString(fmt.Sprintf("**状态**: %s\n\n", event.Status))
	if event.TargetIdent != "" {
		sb.WriteString(fmt.Sprintf("**目标**: %s\n\n", event.TargetIdent))
	}
	if event.Value != "" {
		sb.WriteString(fmt.Sprintf("**当前值**: %s\n\n", event.Value))
	}
	if desc, ok := event.Annotations["content"]; ok && desc != "" {
		sb.WriteString(fmt.Sprintf("**详情**: %s\n\n", desc))
	}
	return sb.String()
}

func buildPlainContent(event *model.MonitorAlertEvent) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("告警名称: %s\n", event.Name))
	sb.WriteString(fmt.Sprintf("严重级别: %s\n", event.Severity))
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

func feishuHeaderColor(severity string) string {
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
