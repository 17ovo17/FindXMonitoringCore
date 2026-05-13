package handler

import (
	"encoding/json"
	"regexp"
	"strings"

	"ai-workbench-api/internal/model"
)

func sanitizeLogPipelines(items []model.LogPipeline) []model.LogPipeline {
	out := make([]model.LogPipeline, 0, len(items))
	for _, item := range items {
		out = append(out, sanitizeLogPipeline(item))
	}
	return out
}

func sanitizeLogPipeline(item model.LogPipeline) model.LogPipeline {
	item.Stages = sanitizeLogJSON(item.Stages)
	item.Config = sanitizeLogJSON(item.Config)
	return item
}

func sanitizeExplorerViews(items []model.ExplorerSavedView) []model.ExplorerSavedView {
	out := make([]model.ExplorerSavedView, 0, len(items))
	for _, item := range items {
		out = append(out, sanitizeExplorerView(item))
	}
	return out
}

func sanitizeExplorerView(item model.ExplorerSavedView) model.ExplorerSavedView {
	item.Query = sanitizeLogJSON(item.Query)
	item.Filters = sanitizeLogJSON(item.Filters)
	item.Columns = sanitizeLogJSON(item.Columns)
	item.TimeRange = sanitizeLogJSON(item.TimeRange)
	item.Layout = sanitizeLogJSON(item.Layout)
	return item
}

func sanitizeLogQueryResponse(resp model.LogQueryResponse) model.LogQueryResponse {
	for idx := range resp.Items {
		resp.Items[idx].Body = sanitizeLogString(resp.Items[idx].Body)
		resp.Items[idx].Attributes = sanitizeLogValue(resp.Items[idx].Attributes).(map[string]any)
	}
	return resp
}

func sanitizeLogContextResponse(resp model.LogContextResponse) model.LogContextResponse {
	if resp.Center != nil {
		resp.Center.Body = sanitizeLogString(resp.Center.Body)
		resp.Center.Attributes = sanitizeLogValue(resp.Center.Attributes).(map[string]any)
	}
	for idx := range resp.Before {
		resp.Before[idx].Body = sanitizeLogString(resp.Before[idx].Body)
		resp.Before[idx].Attributes = sanitizeLogValue(resp.Before[idx].Attributes).(map[string]any)
	}
	for idx := range resp.After {
		resp.After[idx].Body = sanitizeLogString(resp.After[idx].Body)
		resp.After[idx].Attributes = sanitizeLogValue(resp.After[idx].Attributes).(map[string]any)
	}
	for idx := range resp.Items {
		resp.Items[idx].Body = sanitizeLogString(resp.Items[idx].Body)
		resp.Items[idx].Attributes = sanitizeLogValue(resp.Items[idx].Attributes).(map[string]any)
	}
	return resp
}

func sanitizeUserFacingLogQueryResponse(resp model.LogQueryResponse) model.LogQueryResponse {
	resp = sanitizeLogQueryResponse(resp)
	resp.SourceName = sanitizeUserFacingLogString(resp.SourceName)
	resp.Meta = sanitizeUserFacingLogValue(resp.Meta).(map[string]any)
	for idx := range resp.Items {
		resp.Items[idx] = sanitizeUserFacingLogRecord(resp.Items[idx])
	}
	return resp
}

func sanitizeUserFacingLogAggregateResponse(resp model.LogAggregateResponse) model.LogAggregateResponse {
	resp.SourceName = sanitizeUserFacingLogString(resp.SourceName)
	resp.GroupBy = sanitizeUserFacingLogString(resp.GroupBy)
	resp.Meta = sanitizeUserFacingLogValue(resp.Meta).(map[string]any)
	for idx := range resp.Buckets {
		resp.Buckets[idx].Key = sanitizeUserFacingLogString(resp.Buckets[idx].Key)
		resp.Buckets[idx].Label = sanitizeUserFacingLogString(resp.Buckets[idx].Label)
	}
	return resp
}

func sanitizeUserFacingLogContextResponse(resp model.LogContextResponse) model.LogContextResponse {
	resp = sanitizeLogContextResponse(resp)
	resp.SourceName = sanitizeUserFacingLogString(resp.SourceName)
	resp.Meta = sanitizeUserFacingLogValue(resp.Meta).(map[string]any)
	if resp.Center != nil {
		item := sanitizeUserFacingLogRecord(*resp.Center)
		resp.Center = &item
	}
	for idx := range resp.Before {
		resp.Before[idx] = sanitizeUserFacingLogRecord(resp.Before[idx])
	}
	for idx := range resp.After {
		resp.After[idx] = sanitizeUserFacingLogRecord(resp.After[idx])
	}
	for idx := range resp.Items {
		resp.Items[idx] = sanitizeUserFacingLogRecord(resp.Items[idx])
	}
	return resp
}

func sanitizeUserFacingLogRecord(item model.LogRecord) model.LogRecord {
	item.SourceName = sanitizeUserFacingLogString(item.SourceName)
	item.ServiceName = sanitizeUserFacingLogString(item.ServiceName)
	item.Body = sanitizeUserFacingLogString(item.Body)
	item.Attributes = sanitizeUserFacingLogValue(item.Attributes).(map[string]any)
	return item
}

func sanitizeUserFacingLogValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for key, val := range typed {
			out[key] = sanitizeUserFacingLogValue(val)
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, val := range typed {
			out = append(out, sanitizeUserFacingLogValue(val))
		}
		return out
	case string:
		return sanitizeUserFacingLogString(typed)
	default:
		return typed
	}
}

func sanitizeUserFacingLogString(value string) string {
	value = sanitizeLogString(value)
	if strings.TrimSpace(value) == "" {
		return value
	}
	if userFacingLogPrivateKeyPattern.MatchString(value) {
		return "REDACTED"
	}
	for _, item := range userFacingLogBrandReplacements {
		value = strings.ReplaceAll(value, item.from, item.to)
	}
	return value
}

func sanitizeLogJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return json.RawMessage(`null`)
	}
	data, err := json.Marshal(sanitizeLogValue(value))
	if err != nil {
		return json.RawMessage(`null`)
	}
	return data
}

func sanitizeLogValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for key, val := range typed {
			if isSensitiveLogKey(key) {
				out[key] = "REDACTED"
			} else {
				out[key] = sanitizeLogValue(val)
			}
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, val := range typed {
			out = append(out, sanitizeLogValue(val))
		}
		return out
	case string:
		return sanitizeLogString(typed)
	default:
		return typed
	}
}

func isSensitiveLogKey(key string) bool {
	lower := strings.ToLower(strings.TrimSpace(key))
	for _, marker := range []string{"token", "cookie", "dsn", "password", "secret", "api_key", "apikey", "authorization"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func sanitizeLogString(value string) string {
	if strings.TrimSpace(value) == "" {
		return value
	}
	if strings.Contains(strings.ToLower(value), "<redacted>") {
		return "REDACTED"
	}
	if logSecretLinePattern.MatchString(value) || logSecretQueryPattern.MatchString(value) || logBearerPattern.MatchString(value) {
		return "REDACTED"
	}
	return value
}

var (
	logSecretLinePattern  = regexp.MustCompile(`(?i)\b(authorization|token|api[-_]?key|password|cookie|dsn|secret)\s*[:=]\s*(bearer\s+)?[^,\s&;]+`)
	logSecretQueryPattern = regexp.MustCompile(`(?i)[?&](token|access_token|refresh_token|api[-_]?key|apikey|password|cookie|dsn|secret)=`)
	logBearerPattern      = regexp.MustCompile(`(?i)\bbearer\s+[^,\s]+`)

	userFacingLogPrivateKeyPattern = regexp.MustCompile(`(?i)\bprivate\s+key\b`)
	userFacingLogBrandReplacements = []struct {
		from string
		to   string
	}{
		{"prometheus-default", "findx-datasource-default"},
		{"Prometheus", "FindX datasource"},
		{"prometheus", "findx-datasource"},
		{"Nightingale", "FindX"},
		{"SkyWalking", "FindX tracing"},
		{"SigNoZ", "FindX logs"},
		{"AutoOps", "FindX ops"},
		{"Categraf", "FindX Agent"},
		{"Catpaw", "FindX Agent"},
		{"Grafana", "FindX dashboard"},
	}
)
