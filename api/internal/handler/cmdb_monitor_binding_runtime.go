package handler

import (
	"encoding/json"
	"strings"
)

func cmdbMonitorBindingRuntimeContentReady(content json.RawMessage) bool {
	if len(content) == 0 {
		return false
	}
	var payload map[string]any
	if err := json.Unmarshal(content, &payload); err != nil || len(payload) == 0 {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(cmdbMonitorRuntimeString(payload["status"])), "pending") {
		return false
	}
	if _, hasPreview := payload["preview"]; hasPreview && len(payload) <= 2 {
		return false
	}
	for _, key := range []string{"runtime", "executor_ref", "config_snippet_ref", "plugin_id", "rule", "expr"} {
		if runtimeContentValuePresent(payload[key]) {
			return true
		}
	}
	return false
}

func cmdbMonitorRuntimeString(value any) string {
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}

func runtimeContentValuePresent(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(typed) != ""
	case []any:
		return len(typed) > 0
	case map[string]any:
		return len(typed) > 0
	default:
		return true
	}
}

func cmdbSanitizeBindingValue(key string, value any) any {
	if cmdbSensitiveBindingKey(key) {
		return "<REDACTED>"
	}
	switch typed := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, v := range typed {
			out[k] = cmdbSanitizeBindingValue(k, v)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = cmdbSanitizeBindingValue("", item)
		}
		return out
	case string:
		if cmdbSensitiveBindingText(typed) {
			return "<REDACTED>"
		}
		return typed
	default:
		return value
	}
}

func cmdbSensitiveBindingKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, token := range []string{"password", "token", "secret", "cookie", "authorization", "private_key", "dsn", "api_key", "apikey"} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func cmdbSensitiveBindingText(value string) bool {
	lower := strings.ToLower(value)
	for _, token := range []string{"password=", "token=", "secret=", "cookie=", "authorization=", "private_key=", "dsn=", "api_key="} {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}
