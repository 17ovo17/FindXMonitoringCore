package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TracingServiceMetrics answers GET /api/v1/tracing/services/:name/metrics
// Returns time series metrics: p50, p75, p90, p95, p99 latency, throughput, error rate.
func TracingServiceMetrics(c *gin.Context) {
	serviceName := strings.TrimSpace(c.Param("name"))
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	// 解析时间范围参数
	startStr := c.Query("start")
	endStr := c.Query("end")
	stepStr := c.DefaultQuery("step", "60")

	stepSeconds, _ := strconv.ParseInt(stepStr, 10, 64)
	if stepSeconds <= 0 {
		stepSeconds = 60
	}

	now := time.Now()
	var startTime, endTime time.Time

	if startStr != "" {
		if ts, err := strconv.ParseInt(startStr, 10, 64); err == nil {
			startTime = time.Unix(ts, 0)
		} else if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}
	if endStr != "" {
		if ts, err := strconv.ParseInt(endStr, 10, 64); err == nil {
			endTime = time.Unix(ts, 0)
		} else if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	if startTime.IsZero() {
		startTime = now.Add(-30 * time.Minute)
	}
	if endTime.IsZero() {
		endTime = now
	}

	// 尝试从 SkyWalking 获取指标
	svcID := resolveServiceID(c, serviceName)
	if svcID != "" {
		metrics := queryServiceMetricsFromSW(c, svcID, startTime, endTime, stepSeconds)
		if metrics != nil {
			c.JSON(http.StatusOK, metrics)
			return
		}
	}

	// 从 buffer 聚合时间序列指标
	metrics := aggregateServiceMetricsFromBuffer(serviceName, startTime, endTime, stepSeconds)
	c.JSON(http.StatusOK, metrics)
}

// queryServiceMetricsFromSW 从 SkyWalking 查询服务指标时间序列
func queryServiceMetricsFromSW(c *gin.Context, svcID string, start, end time.Time, stepSec int64) gin.H {
	duration := map[string]any{
		"start": start.UTC().Format("2006-01-02 1504"),
		"end":   end.UTC().Format("2006-01-02 1504"),
		"step":  "MINUTE",
	}

	q := `query($duration: Duration!, $condition: MetricsCondition!) {
		values: readMetricsValues(condition: $condition, duration: $duration) {
			values { values { id value } }
		}
	}`
	condition := map[string]any{
		"name": "service_percentile",
		"entity": map[string]any{
			"scope":       "Service",
			"serviceName": svcID,
			"normal":      true,
		},
	}

	var out struct {
		Values struct {
			Values []struct {
				Values []map[string]any `json:"values"`
			} `json:"values"`
		} `json:"values"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{
		"condition": condition,
		"duration":  duration,
	}, &out); err != nil {
		return nil
	}

	return gin.H{
		"service":    svcID,
		"start":      start.Unix(),
		"end":        end.Unix(),
		"step":       stepSec,
		"raw_values": out.Values,
	}
}

// metricsDataPoint 表示一个时间序列数据点
type metricsDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// aggregateServiceMetricsFromBuffer 从 buffer 聚合服务指标时间序列
func aggregateServiceMetricsFromBuffer(serviceName string, start, end time.Time, stepSec int64) gin.H {
	traceBufferMu.RLock()
	defer traceBufferMu.RUnlock()

	// 按时间窗口分桶
	type bucket struct {
		durations  []int64
		errorCount int64
		totalCount int64
	}

	stepDuration := time.Duration(stepSec) * time.Second
	numBuckets := int(end.Sub(start) / stepDuration)
	if numBuckets <= 0 {
		numBuckets = 1
	}
	if numBuckets > 1000 {
		numBuckets = 1000
	}

	buckets := make([]bucket, numBuckets)

	for _, seg := range traceBuffer {
		if !strings.EqualFold(seg.Service, serviceName) {
			continue
		}
		if seg.Timestamp.Before(start) || seg.Timestamp.After(end) {
			continue
		}
		idx := int(seg.Timestamp.Sub(start) / stepDuration)
		if idx < 0 {
			idx = 0
		}
		if idx >= numBuckets {
			idx = numBuckets - 1
		}
		buckets[idx].durations = append(buckets[idx].durations, seg.Duration)
		buckets[idx].totalCount++
		if seg.Status == "error" {
			buckets[idx].errorCount++
		}
	}

	// 构建时间序列
	p50Series := make([]metricsDataPoint, 0, numBuckets)
	p75Series := make([]metricsDataPoint, 0, numBuckets)
	p90Series := make([]metricsDataPoint, 0, numBuckets)
	p95Series := make([]metricsDataPoint, 0, numBuckets)
	p99Series := make([]metricsDataPoint, 0, numBuckets)
	throughputSeries := make([]metricsDataPoint, 0, numBuckets)
	errorRateSeries := make([]metricsDataPoint, 0, numBuckets)

	for i, b := range buckets {
		ts := start.Add(time.Duration(i) * stepDuration).Unix()

		var p50Val, p75Val, p90Val, p95Val, p99Val int64
		if len(b.durations) > 0 {
			sorted := make([]int64, len(b.durations))
			copy(sorted, b.durations)
			sortInt64Slice(sorted)
			n := len(sorted)
			p50Val = sorted[n*50/100]
			p75Val = sorted[n*75/100]
			p90Val = sorted[n*90/100]
			idx95 := n * 95 / 100
			if idx95 >= n {
				idx95 = n - 1
			}
			p95Val = sorted[idx95]
			idx99 := n * 99 / 100
			if idx99 >= n {
				idx99 = n - 1
			}
			p99Val = sorted[idx99]
		}

		p50Series = append(p50Series, metricsDataPoint{Timestamp: ts, Value: float64(p50Val)})
		p75Series = append(p75Series, metricsDataPoint{Timestamp: ts, Value: float64(p75Val)})
		p90Series = append(p90Series, metricsDataPoint{Timestamp: ts, Value: float64(p90Val)})
		p95Series = append(p95Series, metricsDataPoint{Timestamp: ts, Value: float64(p95Val)})
		p99Series = append(p99Series, metricsDataPoint{Timestamp: ts, Value: float64(p99Val)})
		throughputSeries = append(throughputSeries, metricsDataPoint{Timestamp: ts, Value: float64(b.totalCount)})

		errRate := float64(0)
		if b.totalCount > 0 {
			errRate = float64(b.errorCount) / float64(b.totalCount)
		}
		errorRateSeries = append(errorRateSeries, metricsDataPoint{Timestamp: ts, Value: errRate})
	}

	return gin.H{
		"service": serviceName,
		"start":   start.Unix(),
		"end":     end.Unix(),
		"step":    stepSec,
		"series": map[string]any{
			"p50_ms":     p50Series,
			"p75_ms":     p75Series,
			"p90_ms":     p90Series,
			"p95_ms":     p95Series,
			"p99_ms":     p99Series,
			"throughput": throughputSeries,
			"error_rate": errorRateSeries,
		},
	}
}
