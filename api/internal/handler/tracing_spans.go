package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TracingTraceSpans answers GET /api/v1/tracing/traces/:traceId/spans
// Returns spans with logs for a given trace. Proxies to SkyWalking or falls back to buffer.
func TracingTraceSpans(c *gin.Context) {
	traceID := strings.TrimSpace(c.Param("id"))
	if traceID == "" {
		traceID = strings.TrimSpace(c.Param("traceId"))
	}
	if traceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "traceId is required"})
		return
	}

	// 尝试从 SkyWalking 获取
	q := `query($traceId: ID!) {
		trace: queryTrace(traceId: $traceId) {
			spans {
				traceId segmentId spanId parentSpanId
				refs { traceId parentSegmentId parentSpanId type }
				serviceCode serviceInstanceName endpointName
				startTime endTime type peer component isError layer
				tags { key value }
				logs { time data { key value } }
			}
		}
	}`
	var out struct {
		Trace struct {
			Spans []map[string]any `json:"spans"`
		} `json:"trace"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"traceId": traceID}, &out); err == nil && len(out.Trace.Spans) > 0 {
		// 转换为统一格式
		spans := transformSWSpans(out.Trace.Spans)
		c.JSON(http.StatusOK, gin.H{"trace_id": traceID, "spans": spans})
		return
	}

	// 从 buffer 查找
	spans := findSpansFromBuffer(traceID)
	c.JSON(http.StatusOK, gin.H{"trace_id": traceID, "spans": spans})
}

// transformSWSpans 将 SkyWalking span 格式转换为统一输出格式
func transformSWSpans(swSpans []map[string]any) []map[string]any {
	spans := make([]map[string]any, 0, len(swSpans))
	for _, sw := range swSpans {
		span := map[string]any{
			"span_id":      sw["spanId"],
			"parent_id":    sw["parentSpanId"],
			"service":      sw["serviceCode"],
			"operation":    sw["endpointName"],
			"start_time":   sw["startTime"],
			"end_time":     sw["endTime"],
			"type":         sw["type"],
			"peer":         sw["peer"],
			"component":    sw["component"],
			"is_error":     sw["isError"],
			"layer":        sw["layer"],
			"tags":         sw["tags"],
			"segment_id":   sw["segmentId"],
			"trace_id":     sw["traceId"],
		}

		// 计算 duration_ms
		startTime, startOk := sw["startTime"].(float64)
		endTime, endOk := sw["endTime"].(float64)
		if startOk && endOk {
			span["duration_ms"] = int64(endTime - startTime)
		}

		// 转换 logs 格式
		if rawLogs, ok := sw["logs"]; ok {
			span["logs"] = transformSpanLogs(rawLogs)
		} else {
			span["logs"] = []map[string]any{}
		}

		spans = append(spans, span)
	}
	return spans
}

// transformSpanLogs 转换 SkyWalking span logs 为统一格式
func transformSpanLogs(rawLogs any) []map[string]any {
	logs, ok := rawLogs.([]any)
	if !ok {
		return []map[string]any{}
	}

	result := make([]map[string]any, 0, len(logs))
	for _, rawLog := range logs {
		logEntry, ok := rawLog.(map[string]any)
		if !ok {
			continue
		}

		entry := map[string]any{
			"timestamp": logEntry["time"],
		}

		// 转换 data 字段为 fields map
		fields := make(map[string]any)
		if data, ok := logEntry["data"].([]any); ok {
			for _, item := range data {
				kv, ok := item.(map[string]any)
				if !ok {
					continue
				}
				key, _ := kv["key"].(string)
				value := kv["value"]
				if key != "" {
					fields[key] = value
				}
			}
		}
		entry["fields"] = fields
		result = append(result, entry)
	}
	return result
}

// findSpansFromBuffer 从内存 buffer 中查找指定 traceId 的 spans
func findSpansFromBuffer(traceID string) []map[string]any {
	traceBufferMu.RLock()
	defer traceBufferMu.RUnlock()

	spans := make([]map[string]any, 0)
	for _, seg := range traceBuffer {
		if seg.TraceID != traceID {
			continue
		}
		span := map[string]any{
			"span_id":     seg.SpanID,
			"parent_id":   "",
			"service":     seg.Service,
			"operation":   seg.Operation,
			"duration_ms": seg.Duration,
			"start_time":  seg.Timestamp.UnixMilli(),
			"end_time":    seg.Timestamp.UnixMilli() + seg.Duration,
			"is_error":    seg.Status == "error",
			"tags":        seg.Tags,
			"logs":        []map[string]any{},
			"trace_id":    seg.TraceID,
		}
		spans = append(spans, span)
	}
	return spans
}
