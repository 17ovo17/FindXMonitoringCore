package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/monitoring"

	"github.com/gin-gonic/gin"
)

func callPrometheus(parent context.Context, base, path string, params url.Values, timeout time.Duration) promCallResult {
	result, err := monitoring.NewPrometheusGateway(nil).Call(parent, monitoring.PrometheusCallRequest{
		BaseURL: base, Path: path, Params: params, Timeout: timeout,
	})
	out := promCallResult{
		Body: result.Body, Data: result.Data, Stats: result.Stats, LatencyMS: result.LatencyMS,
		StatusCode: result.StatusCode, Warnings: result.Warnings, OK: err == nil,
	}
	if err != nil {
		out.StatusCode = monitoring.HTTPStatus(err)
		return out
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
