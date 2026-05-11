package handler

import (
	"encoding/json"
	"fmt"
	"strings"
)

func compactTextList(items []string, limit int) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		item = strings.TrimSpace(compactAIInspectionText(item))
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func aiText(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]any:
		return aiMapText(typed)
	case []any:
		return strings.Join(aiTextList(typed), "; ")
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%.0f", typed)
	default:
		if value == nil {
			return ""
		}
		return fmt.Sprintf("%v", value)
	}
}

func aiMapText(value map[string]any) string {
	if text := aiBusinessSummaryText(value); text != "" {
		return text
	}
	parts := []string{}
	for _, key := range []string{"action", "priority", "expected_benefit", "result", "status", "overall_assessment", "business_health_conclusion", "overall", "summary", "detail", "severity", "target", "business_name", "topology_complete"} {
		if text := aiText(value[key]); text != "" {
			parts = append(parts, key+": "+text)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, "; ")
	}
	data, _ := json.Marshal(value)
	return string(data)
}

func aiBusinessSummaryText(value map[string]any) string {
	layers, ok := value["layers"].(map[string]any)
	if !ok {
		return ""
	}
	parts := []string{}
	for _, key := range []string{"entry", "application", "middleware", "database"} {
		if text := aiLayerText(key, layers[key]); text != "" {
			parts = append(parts, text)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "AI business inspection completed: " + strings.Join(parts, "; ") + aiAlertSuffix(value)
}

func aiAlertSuffix(value map[string]any) string {
	alerts, ok := value["alerts"].(map[string]any)
	if !ok {
		return ""
	}
	if firing, ok := alerts["firing"].([]any); ok && len(firing) > 0 {
		return fmt.Sprintf("; firing alerts=%d", len(firing))
	}
	return ""
}

func aiLayerText(layer string, value any) string {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return ""
	}
	names := []string{}
	for _, item := range items {
		if text := aiLayerItemText(item); text != "" {
			names = append(names, text)
		}
	}
	if len(names) == 0 {
		return ""
	}
	return layer + "=" + strings.Join(names, ", ")
}

func aiLayerItemText(item any) string {
	obj, ok := item.(map[string]any)
	if !ok {
		return ""
	}
	service := aiText(obj["service"])
	ip := aiText(obj["ip"])
	port := aiText(obj["port"])
	if service == "" || ip == "" || port == "" {
		return ""
	}
	return fmt.Sprintf("%s %s:%s", service, ip, port)
}

func aiTextList(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := []string{}
		for _, item := range typed {
			if text := aiText(item); text != "" {
				out = append(out, text)
			}
		}
		return out
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		return []string{strings.TrimSpace(typed)}
	default:
		return nil
	}
}

func aiInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case float64:
		return int(typed)
	case string:
		var out int
		fmt.Sscanf(typed, "%d", &out)
		return out
	default:
		return 0
	}
}
