package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	traceBufferMaxSize   = 1000
	traceForwardTimeout  = 10 * time.Second
)

// traceSegment 表示一条 OTLP JSON 格式的 trace segment。
type traceSegment struct {
	TraceID   string         `json:"trace_id"`
	SpanID    string         `json:"span_id"`
	Service   string         `json:"service"`
	Operation string         `json:"operation"`
	Duration  int64          `json:"duration_ms"`
	Status    string         `json:"status"`
	Tags      map[string]any `json:"tags,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

var (
	traceBuffer   []traceSegment
	traceBufferMu sync.RWMutex
)

func init() {
	traceBuffer = make([]traceSegment, 0, traceBufferMaxSize)
}

func getSkywalkingOAPURL() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("SKYWALKING_OAP_URL")), "/")
}

// APMTraceReceive POST /findx-agent/v1/traces — 接收 OTLP JSON trace 并转发/缓存
// 此 handler 增强现有的 FindXAgentTracesCompatibleReceiver，增加真实转发和内存缓存。
func APMTraceReceive(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body failed"})
		return
	}
	defer c.Request.Body.Close()

	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty trace payload"})
		return
	}

	// 解析 OTLP JSON 格式的 traces
	segments := parseTraceSegments(body)
	if len(segments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid trace segments found"})
		return
	}

	// 尝试转发到 SkyWalking OAP
	oapURL := getSkywalkingOAPURL()
	if oapURL != "" {
		go forwardToOAP(oapURL, body)
	}

	// 存入内存 buffer
	bufferTraceSegments(segments)

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"accepted":       len(segments),
		"forwarded":      oapURL != "",
		"buffer_size":    getTraceBufferSize(),
	})
}

// APMTraceQuery POST /api/v1/tracing/traces/query — 从内存 buffer 查询 traces
// 增强现有的 TracingQueryTracesSW，当 SkyWalking 不可用时从 buffer 查询。
func APMTraceQueryBuffer(c *gin.Context) {
	var req struct {
		TraceID string `json:"trace_id"`
		Service string `json:"service"`
		Limit   int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Limit = 50
	}
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}

	traceBufferMu.RLock()
	defer traceBufferMu.RUnlock()

	results := make([]traceSegment, 0)
	for i := len(traceBuffer) - 1; i >= 0 && len(results) < req.Limit; i-- {
		seg := traceBuffer[i]
		if req.TraceID != "" && seg.TraceID != req.TraceID {
			continue
		}
		if req.Service != "" && !strings.Contains(strings.ToLower(seg.Service), strings.ToLower(req.Service)) {
			continue
		}
		results = append(results, seg)
	}
	c.JSON(http.StatusOK, gin.H{"traces": results, "total": len(results)})
}

func parseTraceSegments(body []byte) []traceSegment {
	// 尝试解析为数组
	var segments []traceSegment
	if err := json.Unmarshal(body, &segments); err == nil && len(segments) > 0 {
		for i := range segments {
			if segments[i].Timestamp.IsZero() {
				segments[i].Timestamp = time.Now()
			}
		}
		return segments
	}

	// 尝试解析为单个 segment
	var single traceSegment
	if err := json.Unmarshal(body, &single); err == nil && single.TraceID != "" {
		if single.Timestamp.IsZero() {
			single.Timestamp = time.Now()
		}
		return []traceSegment{single}
	}

	// 尝试解析 OTLP resourceSpans 格式
	var otlp struct {
		ResourceSpans []struct {
			ScopeSpans []struct {
				Spans []struct {
					TraceID            string `json:"traceId"`
					SpanID             string `json:"spanId"`
					Name               string `json:"name"`
					Kind               int    `json:"kind"`
					StartTimeUnixNano  int64  `json:"startTimeUnixNano"`
					EndTimeUnixNano    int64  `json:"endTimeUnixNano"`
					Status             struct {
						Code int `json:"code"`
					} `json:"status"`
				} `json:"spans"`
			} `json:"scopeSpans"`
			Resource struct {
				Attributes []struct {
					Key   string `json:"key"`
					Value struct {
						StringValue string `json:"stringValue"`
					} `json:"value"`
				} `json:"attributes"`
			} `json:"resource"`
		} `json:"resourceSpans"`
	}
	if err := json.Unmarshal(body, &otlp); err == nil && len(otlp.ResourceSpans) > 0 {
		for _, rs := range otlp.ResourceSpans {
			svcName := ""
			for _, attr := range rs.Resource.Attributes {
				if attr.Key == "service.name" {
					svcName = attr.Value.StringValue
					break
				}
			}
			for _, ss := range rs.ScopeSpans {
				for _, span := range ss.Spans {
					durationMs := int64(0)
					if span.EndTimeUnixNano > span.StartTimeUnixNano {
						durationMs = (span.EndTimeUnixNano - span.StartTimeUnixNano) / 1_000_000
					}
					status := "ok"
					if span.Status.Code == 2 {
						status = "error"
					}
					segments = append(segments, traceSegment{
						TraceID:   span.TraceID,
						SpanID:    span.SpanID,
						Service:   svcName,
						Operation: span.Name,
						Duration:  durationMs,
						Status:    status,
						Timestamp: time.Now(),
					})
				}
			}
		}
		return segments
	}

	return nil
}

func bufferTraceSegments(segments []traceSegment) {
	traceBufferMu.Lock()
	defer traceBufferMu.Unlock()

	traceBuffer = append(traceBuffer, segments...)
	// 保持 buffer 不超过最大容量
	if len(traceBuffer) > traceBufferMaxSize {
		excess := len(traceBuffer) - traceBufferMaxSize
		traceBuffer = traceBuffer[excess:]
	}
}

func getTraceBufferSize() int {
	traceBufferMu.RLock()
	defer traceBufferMu.RUnlock()
	return len(traceBuffer)
}

func forwardToOAP(oapURL string, body []byte) {
	target := oapURL + "/v3/segments"
	ctx, cancel := context.WithTimeout(context.Background(), traceForwardTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		logrus.WithError(err).Warn("apm: create OAP forward request failed")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: traceForwardTimeout}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("apm: forward to OAP failed")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		logrus.Warnf("apm: OAP returned status %d", resp.StatusCode)
	}
}
