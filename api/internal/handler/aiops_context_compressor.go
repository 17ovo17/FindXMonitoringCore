package handler

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	compressMaxChars    = 2000
	compressMaxItems    = 20
	compressSummarySize = 10
)

// CompressToolOutput compresses large tool outputs into structured summaries
// before sending to LLM. If output > 2000 chars, summarize; if array > 20 items,
// take top 10 + summary. Preserves key fields (error messages, metric values, timestamps).
func CompressToolOutput(output any) any {
	if output == nil {
		return output
	}
	switch v := output.(type) {
	case map[string]any:
		return compressMap(v)
	case []any:
		return compressSlice(v)
	default:
		text := fmt.Sprint(output)
		if len(text) > compressMaxChars {
			return compressText(text)
		}
		return output
	}
}

func compressMap(m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		if isKeyField(k) {
			result[k] = v
			continue
		}
		switch typed := v.(type) {
		case []any:
			result[k] = compressSlice(typed)
		case map[string]any:
			result[k] = compressMap(typed)
		case string:
			if len(typed) > compressMaxChars {
				result[k] = compressText(typed)
			} else {
				result[k] = typed
			}
		default:
			result[k] = v
		}
	}
	return result
}

func compressSlice(items []any) any {
	if len(items) <= compressMaxItems {
		return items
	}
	summary := make([]any, 0, compressSummarySize+1)
	for i := 0; i < compressSummarySize && i < len(items); i++ {
		summary = append(summary, items[i])
	}
	summary = append(summary, map[string]any{
		"_compressed": true,
		"_total":      len(items),
		"_shown":      compressSummarySize,
		"_omitted":    len(items) - compressSummarySize,
		"_note":       fmt.Sprintf("显示前 %d 条，共 %d 条结果", compressSummarySize, len(items)),
	})
	return summary
}

func compressText(text string) string {
	runes := []rune(text)
	if len(runes) <= compressMaxChars {
		return text
	}
	// Preserve first 1500 chars and last 300 chars with a separator
	head := string(runes[:1500])
	tail := string(runes[len(runes)-300:])
	omitted := len(runes) - 1800
	return head + fmt.Sprintf("\n\n... [省略 %d 字符] ...\n\n", omitted) + tail
}

// isKeyField returns true for fields that should always be preserved in full.
func isKeyField(key string) bool {
	preserveKeys := []string{
		"error", "err", "message", "msg",
		"timestamp", "time", "created_at", "updated_at",
		"value", "metric", "score", "severity",
		"status", "id", "name", "title",
		"ip", "host", "target", "source",
		"total", "count",
	}
	lower := strings.ToLower(key)
	for _, k := range preserveKeys {
		if lower == k {
			return true
		}
	}
	return false
}

// CompressForLLM compresses a full context payload (multiple tool results)
// into a size suitable for LLM context windows.
func CompressForLLM(data any, maxLen int) string {
	if maxLen <= 0 {
		maxLen = compressMaxChars
	}
	compressed := CompressToolOutput(data)
	b, err := json.Marshal(compressed)
	if err != nil {
		return fmt.Sprint(data)
	}
	text := string(b)
	if len(text) <= maxLen {
		return text
	}
	return compressText(text)
}
