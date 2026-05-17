package notifier

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"ai-workbench-api/internal/model"
)

// TemplateData 是通知模板渲染时可用的全部变量。
type TemplateData struct {
	RuleName    string            `json:"rule_name"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Value       string            `json:"value"`
	StartsAt    time.Time         `json:"starts_at"`
	EndsAt      time.Time         `json:"ends_at"`
	Duration    string            `json:"duration"`
	Host        string            `json:"host"`
	IP          string            `json:"ip"`
	Name        string            `json:"name"`
	Target      string            `json:"target"`
	RuleID      string            `json:"rule_id"`
}

// templateFuncMap 提供模板中可用的自定义函数。
var templateFuncMap = template.FuncMap{
	"timeFormat": func(t time.Time, layout string) string {
		if t.IsZero() {
			return ""
		}
		return t.Format(layout)
	},
	"jsonEscape": func(s string) string {
		b, err := json.Marshal(s)
		if err != nil {
			return s
		}
		// 去掉首尾引号
		return string(b[1 : len(b)-1])
	},
	"truncate": func(s string, maxLen int) string {
		if len(s) <= maxLen {
			return s
		}
		return s[:maxLen] + "..."
	},
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
}

// NewTemplateData 从告警事件构建模板数据。
func NewTemplateData(event *model.MonitorAlertEvent) *TemplateData {
	if event == nil {
		return &TemplateData{}
	}
	duration := ""
	if !event.FirstSeen.IsZero() && !event.LastSeen.IsZero() {
		d := event.LastSeen.Sub(event.FirstSeen)
		duration = d.String()
	}
	host := ""
	ip := ""
	if event.Labels != nil {
		host = event.Labels["host"]
		if host == "" {
			host = event.Labels["hostname"]
		}
		ip = event.Labels["ip"]
		if ip == "" {
			ip = event.Labels["instance"]
		}
	}
	return &TemplateData{
		RuleName:    event.Name,
		Severity:    event.Severity,
		Status:      event.Status,
		Labels:      event.Labels,
		Annotations: event.Annotations,
		Value:       event.Value,
		StartsAt:    event.FirstSeen,
		EndsAt:      event.LastSeen,
		Duration:    duration,
		Host:        host,
		IP:          ip,
		Name:        event.Name,
		Target:      event.TargetIdent,
		RuleID:      event.RuleID,
	}
}

// RenderTemplate 使用 Go text/template 渲染通知模板。
// 内置变量：RuleName, Severity, Status, Labels, Annotations, Value, StartsAt, EndsAt, Duration, Host, IP
// 支持自定义函数：timeFormat, jsonEscape, truncate, upper, lower
func RenderTemplate(tplContent string, event *model.MonitorAlertEvent) (string, error) {
	if tplContent == "" {
		return "", fmt.Errorf("template content is empty")
	}
	data := NewTemplateData(event)
	tmpl, err := template.New("notification").Funcs(templateFuncMap).Parse(tplContent)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

// RenderTemplatePreview 渲染模板并返回预览结果，使用示例数据。
func RenderTemplatePreview(tplContent string) (string, error) {
	sampleEvent := &model.MonitorAlertEvent{
		ID:          "preview-001",
		Name:        "CPU 使用率过高",
		Severity:    "warning",
		Status:      "firing",
		Value:       "95.2%",
		TargetIdent: "prod-web-01",
		RuleID:      "rule-cpu-high",
		Labels: map[string]string{
			"host":     "prod-web-01",
			"ip":       "10.0.1.100",
			"service":  "web-api",
			"instance": "10.0.1.100:9090",
		},
		Annotations: map[string]string{
			"summary":     "CPU 使用率超过 90%",
			"description": "节点 prod-web-01 的 CPU 使用率已达 95.2%，持续 5 分钟",
			"content":     "请检查是否有异常进程占用 CPU",
		},
		FirstSeen: time.Now().Add(-5 * time.Minute),
		LastSeen:  time.Now(),
	}
	return RenderTemplate(tplContent, sampleEvent)
}
