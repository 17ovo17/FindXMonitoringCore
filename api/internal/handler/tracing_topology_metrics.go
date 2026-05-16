package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TracingTopologyMetrics answers GET /api/v1/tracing/topology/:service/metrics
// Returns metrics for a topology node popup: cpm, avg_latency, error_rate, percentiles.
func TracingTopologyMetrics(c *gin.Context) {
	serviceName := strings.TrimSpace(c.Param("service"))
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	svcID := resolveServiceID(c, serviceName)
	if svcID != "" {
		metrics := queryServiceTopologyMetrics(c, svcID)
		c.JSON(http.StatusOK, metrics)
		return
	}

	// 从 buffer 聚合
	metrics := aggregateTopologyMetricsFromBuffer(serviceName)
	c.JSON(http.StatusOK, metrics)
}

// TracingTopologyDependencies answers GET /api/v1/tracing/topology/:service/dependencies
// Returns upstream/downstream services with call metrics.
func TracingTopologyDependencies(c *gin.Context) {
	serviceName := strings.TrimSpace(c.Param("service"))
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	// 从全局拓扑中提取依赖关系
	q := `query($duration: Duration!) { topology: getGlobalTopology(duration: $duration) { nodes { id name type isReal } calls { id source target detectPoints } } }`
	var out struct {
		Topology struct {
			Nodes []map[string]any `json:"nodes"`
			Calls []map[string]any `json:"calls"`
		} `json:"topology"`
	}

	if err := swClient.Query(c.Request.Context(), q, map[string]any{"duration": durationFor(30)}, &out); err != nil {
		// 从 buffer 聚合依赖
		deps := aggregateTopologyDependenciesFromBuffer(serviceName)
		c.JSON(http.StatusOK, deps)
		return
	}

	// 构建节点 ID -> 名称映射
	nodeNameMap := make(map[string]string)
	var serviceNodeID string
	for _, node := range out.Topology.Nodes {
		id, _ := node["id"].(string)
		name, _ := node["name"].(string)
		nodeNameMap[id] = name
		if strings.EqualFold(name, serviceName) {
			serviceNodeID = id
		}
	}

	upstream := make([]map[string]any, 0)
	downstream := make([]map[string]any, 0)

	for _, call := range out.Topology.Calls {
		source, _ := call["source"].(string)
		target, _ := call["target"].(string)

		if source == serviceNodeID {
			downstream = append(downstream, map[string]any{
				"service":       nodeNameMap[target],
				"service_id":    target,
				"detect_points": call["detectPoints"],
			})
		}
		if target == serviceNodeID {
			upstream = append(upstream, map[string]any{
				"service":       nodeNameMap[source],
				"service_id":    source,
				"detect_points": call["detectPoints"],
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"service":    serviceName,
		"upstream":   upstream,
		"downstream": downstream,
	})
}

// queryServiceTopologyMetrics 从 SkyWalking 查询服务拓扑指标
func queryServiceTopologyMetrics(c *gin.Context, svcID string) gin.H {
	metricsQuery := `query($duration: Duration!, $condition: MetricsCondition!) {
		cpm: readMetricsValue(condition: $condition, duration: $duration)
	}`
	condition := map[string]any{
		"name": "service_cpm",
		"entity": map[string]any{
			"scope":     "Service",
			"serviceName": svcID,
			"normal":    true,
		},
	}
	var metricsOut struct {
		CPM int `json:"cpm"`
	}
	_ = swClient.Query(c.Request.Context(), metricsQuery, map[string]any{
		"condition": condition,
		"duration":  durationFor(30),
	}, &metricsOut)

	// 从 buffer 补充百分位数据
	traceBufferMu.RLock()
	defer traceBufferMu.RUnlock()

	var durations []int64
	var errorCount int64
	var totalCount int64

	for _, seg := range traceBuffer {
		if seg.Service == "" {
			continue
		}
		totalCount++
		durations = append(durations, seg.Duration)
		if seg.Status == "error" {
			errorCount++
		}
	}

	p50, p90, p99 := calculatePercentiles(durations)
	avgLatency := int64(0)
	if totalCount > 0 {
		var sum int64
		for _, d := range durations {
			sum += d
		}
		avgLatency = sum / totalCount
	}
	errorRate := float64(0)
	if totalCount > 0 {
		errorRate = float64(errorCount) / float64(totalCount)
	}

	return gin.H{
		"cpm":            metricsOut.CPM,
		"avg_latency_ms": avgLatency,
		"error_rate":     errorRate,
		"p50_ms":         p50,
		"p75_ms":         int64(0),
		"p90_ms":         p90,
		"p95_ms":         int64(0),
		"p99_ms":         p99,
	}
}

// aggregateTopologyMetricsFromBuffer 从 buffer 聚合拓扑指标
func aggregateTopologyMetricsFromBuffer(serviceName string) gin.H {
	traceBufferMu.RLock()
	defer traceBufferMu.RUnlock()

	var durations []int64
	var errorCount int64
	var totalCount int64

	for _, seg := range traceBuffer {
		if !strings.EqualFold(seg.Service, serviceName) {
			continue
		}
		totalCount++
		durations = append(durations, seg.Duration)
		if seg.Status == "error" {
			errorCount++
		}
	}

	p50, p90, p99 := calculatePercentiles(durations)
	p75 := int64(0)
	p95 := int64(0)
	if len(durations) > 0 {
		sorted := make([]int64, len(durations))
		copy(sorted, durations)
		sortInt64Slice(sorted)
		n := len(sorted)
		p75 = sorted[n*75/100]
		idx95 := n * 95 / 100
		if idx95 >= n {
			idx95 = n - 1
		}
		p95 = sorted[idx95]
	}

	avgLatency := int64(0)
	if totalCount > 0 {
		var sum int64
		for _, d := range durations {
			sum += d
		}
		avgLatency = sum / totalCount
	}
	errorRate := float64(0)
	if totalCount > 0 {
		errorRate = float64(errorCount) / float64(totalCount)
	}

	return gin.H{
		"cpm":            totalCount,
		"avg_latency_ms": avgLatency,
		"error_rate":     errorRate,
		"p50_ms":         p50,
		"p75_ms":         p75,
		"p90_ms":         p90,
		"p95_ms":         p95,
		"p99_ms":         p99,
	}
}

// aggregateTopologyDependenciesFromBuffer 从 buffer 聚合依赖关系
func aggregateTopologyDependenciesFromBuffer(serviceName string) gin.H {
	// 内存 buffer 中没有跨服务调用关系的完整信息
	// 返回空依赖列表
	return gin.H{
		"service":    serviceName,
		"upstream":   []map[string]any{},
		"downstream": []map[string]any{},
	}
}
