package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
)

func callPrometheus(parent context.Context, base, path string, params url.Values, timeout time.Duration) promCallResult {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	endpoint := strings.TrimRight(base, "/") + path
	if encoded := params.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	started := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return promCallResult{StatusCode: http.StatusBadRequest}
	}
	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	out := promCallResult{LatencyMS: time.Since(started).Milliseconds(), StatusCode: http.StatusServiceUnavailable}
	if err != nil {
		return out
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		out.StatusCode = http.StatusServiceUnavailable
		return out
	}
	return decodePrometheusResponse(raw, out)
}

func decodePrometheusResponse(raw []byte, out promCallResult) promCallResult {
	if err := json.Unmarshal(raw, &out.Body); err != nil {
		out.StatusCode = http.StatusInternalServerError
		return out
	}
	if rawWarnings, _ := out.Body["warnings"].([]any); len(rawWarnings) > 0 {
		for _, item := range rawWarnings {
			if s, ok := item.(string); ok {
				out.Warnings = append(out.Warnings, s)
			}
		}
	}
	if status, _ := out.Body["status"].(string); status == "success" {
		out.StatusCode, out.OK = http.StatusOK, true
	} else {
		out.StatusCode = http.StatusServiceUnavailable
	}
	return out
}

func writePrometheusStringList(c *gin.Context, datasourceID, label string, result promCallResult, limit int) {
	if !result.OK {
		writeMonitorError(c, result.StatusCode, "prometheus labels failed")
		return
	}
	values, ok := prometheusStringData(result.Body)
	if !ok {
		writeMonitorError(c, http.StatusInternalServerError, "invalid prometheus response")
		return
	}
	if len(values) > limit {
		values = values[:limit]
	}
	body := gin.H{"datasource_id": datasourceID, "items": values, "limit": limit, "warnings": result.Warnings}
	if label != "" {
		body["label"] = label
	}
	c.JSON(http.StatusOK, body)
}

func prometheusStringData(body map[string]any) ([]string, bool) {
	data, ok := body["data"].([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(data))
	for _, item := range data {
		if value, ok := item.(string); ok && value != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out, true
}

func validatePromQL(query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return fmt.Errorf("query required")
	}
	if len(query) > 4096 {
		return fmt.Errorf("query too long")
	}
	for _, r := range query {
		if unicode.IsControl(r) {
			return fmt.Errorf("query contains invalid control characters")
		}
	}
	lower := strings.ToLower(query)
	for _, term := range []string{"delete_series", "/api/v1/admin", "api/v1/admin", "/admin/", " admin "} {
		if strings.Contains(lower, term) {
			return fmt.Errorf("query rejected by safety policy")
		}
	}
	return nil
}

func sortPrometheusMatrix(body map[string]any) map[string]any {
	data, _ := body["data"].(map[string]any)
	if data["resultType"] != "matrix" {
		return body
	}
	result, _ := data["result"].([]any)
	sort.SliceStable(result, func(i, j int) bool { return metricSortKey(result[i]) < metricSortKey(result[j]) })
	data["result"], body["data"] = result, data
	return body
}

func metricSortKey(item any) string {
	row, _ := item.(map[string]any)
	metric, _ := row["metric"].(map[string]any)
	keys := make([]string, 0, len(metric))
	for key := range metric {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+fmt.Sprint(metric[key]))
	}
	return strings.Join(parts, "\xff")
}
