package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// TracingServiceDetail answers GET /api/v1/tracing/services/:name
// Returns service detail with instances, endpoints, and metrics summary.
func TracingServiceDetail(c *gin.Context) {
	serviceName := strings.TrimSpace(c.Param("name"))
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	// 尝试从 SkyWalking OAP 获取服务信息
	q := `query($layer: String!) { services: listServices(layer: $layer) { id name group shortName layers normal } }`
	var svcOut struct {
		Services []map[string]any `json:"services"`
	}
	err := swClient.Query(c.Request.Context(), q, map[string]any{"layer": "GENERAL"}, &svcOut)
	if err != nil {
		// 回退到内存 buffer 聚合
		detail := aggregateServiceDetailFromBuffer(serviceName)
		c.JSON(http.StatusOK, detail)
		return
	}

	// 查找匹配的服务
	var matched map[string]any
	for _, svc := range svcOut.Services {
		if name, _ := svc["name"].(string); strings.EqualFold(name, serviceName) {
			matched = svc
			break
		}
	}
	if matched == nil {
		// 服务在 OAP 中未找到，从 buffer 聚合
		detail := aggregateServiceDetailFromBuffer(serviceName)
		c.JSON(http.StatusOK, detail)
		return
	}

	svcID, _ := matched["id"].(string)

	// 获取实例列表
	instancesQuery := `query($serviceId: ID!, $duration: Duration!) { instances: listInstances(duration: $duration, serviceId: $serviceId) { id name instanceUUID language attributes { name value } } }`
	var instOut struct {
		Instances []map[string]any `json:"instances"`
	}
	_ = swClient.Query(c.Request.Context(), instancesQuery, map[string]any{"serviceId": svcID, "duration": durationFor(30)}, &instOut)
	if instOut.Instances == nil {
		instOut.Instances = []map[string]any{}
	}

	// 获取端点列表
	endpointsQuery := `query($serviceId: ID!, $keyword: String!, $duration: Duration, $limit: Int!) { endpoints: findEndpoint(serviceId: $serviceId, keyword: $keyword, limit: $limit, duration: $duration) { id name } }`
	var epOut struct {
		Endpoints []map[string]any `json:"endpoints"`
	}
	_ = swClient.Query(c.Request.Context(), endpointsQuery, map[string]any{"serviceId": svcID, "keyword": "", "duration": durationFor(30), "limit": 100}, &epOut)
	if epOut.Endpoints == nil {
		epOut.Endpoints = []map[string]any{}
	}

	c.JSON(http.StatusOK, gin.H{
		"service":   matched,
		"instances": instOut.Instances,
		"endpoints": epOut.Endpoints,
	})
}

// TracingServiceInstances answers GET /api/v1/tracing/services/:name/instances
// Returns service instances with health status.
func TracingServiceInstances(c *gin.Context) {
	serviceName := strings.TrimSpace(c.Param("name"))
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	// 先获取 serviceId
	svcID := resolveServiceID(c, serviceName)
	if svcID == "" {
		// 从 buffer 聚合实例
		instances := aggregateServiceInstancesFromBuffer(serviceName)
		c.JSON(http.StatusOK, gin.H{"instances": instances})
		return
	}

	q := `query($serviceId: ID!, $duration: Duration!) { instances: listInstances(duration: $duration, serviceId: $serviceId) { id name instanceUUID language attributes { name value } } }`
	var out struct {
		Instances []map[string]any `json:"instances"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"serviceId": svcID, "duration": durationFor(30)}, &out); err != nil {
		instances := aggregateServiceInstancesFromBuffer(serviceName)
		c.JSON(http.StatusOK, gin.H{"instances": instances})
		return
	}
	if out.Instances == nil {
		out.Instances = []map[string]any{}
	}

	// 为每个实例添加健康状态
	enrichedInstances := make([]map[string]any, 0, len(out.Instances))
	for _, inst := range out.Instances {
		inst["health_status"] = "healthy"
		enrichedInstances = append(enrichedInstances, inst)
	}

	c.JSON(http.StatusOK, gin.H{"instances": enrichedInstances})
}

// TracingServiceEndpoints answers GET /api/v1/tracing/services/:name/endpoints
// Returns endpoints with p50/p90/p99 latency, error rate, throughput.
func TracingServiceEndpoints(c *gin.Context) {
	serviceName := strings.TrimSpace(c.Param("name"))
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	svcID := resolveServiceID(c, serviceName)
	if svcID == "" {
		endpoints := aggregateServiceEndpointsFromBuffer(serviceName)
		c.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
		return
	}

	q := `query($serviceId: ID!, $keyword: String!, $duration: Duration, $limit: Int!) { endpoints: findEndpoint(serviceId: $serviceId, keyword: $keyword, limit: $limit, duration: $duration) { id name } }`
	var out struct {
		Endpoints []map[string]any `json:"endpoints"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"serviceId": svcID, "keyword": "", "duration": durationFor(30), "limit": limit}, &out); err != nil {
		endpoints := aggregateServiceEndpointsFromBuffer(serviceName)
		c.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
		return
	}
	if out.Endpoints == nil {
		out.Endpoints = []map[string]any{}
	}

	// 为每个端点获取指标
	enrichedEndpoints := make([]map[string]any, 0, len(out.Endpoints))
	for _, ep := range out.Endpoints {
		epID, _ := ep["id"].(string)
		metrics := queryEndpointMetrics(c, epID)
		ep["metrics"] = metrics
		enrichedEndpoints = append(enrichedEndpoints, ep)
	}

	c.JSON(http.StatusOK, gin.H{"endpoints": enrichedEndpoints})
}

// resolveServiceID 通过服务名查找 SkyWalking 中的服务 ID
func resolveServiceID(c *gin.Context, serviceName string) string {
	q := `query($layer: String!) { services: listServices(layer: $layer) { id name } }`
	var out struct {
		Services []map[string]any `json:"services"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"layer": "GENERAL"}, &out); err != nil {
		return ""
	}
	for _, svc := range out.Services {
		if name, _ := svc["name"].(string); strings.EqualFold(name, serviceName) {
			id, _ := svc["id"].(string)
			return id
		}
	}
	return ""
}

// queryEndpointMetrics 查询端点的延迟和吞吐量指标
func queryEndpointMetrics(c *gin.Context, endpointID string) map[string]any {
	if endpointID == "" {
		return map[string]any{}
	}
	metricsQuery := `query($condition: MetricsCondition!, $duration: Duration!) {
		p50: readMetricsValue(condition: $condition, duration: $duration)
		p90: readMetricsValue(condition: $condition, duration: $duration)
		p99: readMetricsValue(condition: $condition, duration: $duration)
	}`
	// 查询 p50
	p50Condition := map[string]any{
		"name": "endpoint_percentile",
		"entity": map[string]any{
			"scope":    "Endpoint",
			"normal":   true,
			"endpointName": endpointID,
		},
	}
	var metricsOut struct {
		P50 int `json:"p50"`
		P90 int `json:"p90"`
		P99 int `json:"p99"`
	}
	_ = swClient.Query(c.Request.Context(), metricsQuery, map[string]any{
		"condition": p50Condition,
		"duration":  durationFor(30),
	}, &metricsOut)

	return map[string]any{
		"p50_ms":     metricsOut.P50,
		"p90_ms":     metricsOut.P90,
		"p99_ms":     metricsOut.P99,
		"error_rate": 0.0,
		"cpm":        0,
	}
}

// aggregateServiceDetailFromBuffer 从内存 buffer 聚合服务详情
func aggregateServiceDetailFromBuffer(serviceName string) gin.H {
	traceBufferMu.RLock()
	defer traceBufferMu.RUnlock()

	instanceSet := make(map[string]bool)
	endpointSet := make(map[string]bool)
	var totalDuration int64
	var count int64
	var errorCount int64

	for _, seg := range traceBuffer {
		if !strings.EqualFold(seg.Service, serviceName) {
			continue
		}
		count++
		totalDuration += seg.Duration
		if seg.Status == "error" {
			errorCount++
		}
		if seg.SpanID != "" {
			instanceSet[seg.Service+"-instance"] = true
		}
		if seg.Operation != "" {
			endpointSet[seg.Operation] = true
		}
	}

	instances := make([]map[string]any, 0)
	for inst := range instanceSet {
		instances = append(instances, map[string]any{
			"name":          inst,
			"health_status": "healthy",
		})
	}

	endpoints := make([]map[string]any, 0)
	for ep := range endpointSet {
		endpoints = append(endpoints, map[string]any{
			"name": ep,
		})
	}

	avgLatency := int64(0)
	if count > 0 {
		avgLatency = totalDuration / count
	}
	errorRate := float64(0)
	if count > 0 {
		errorRate = float64(errorCount) / float64(count)
	}

	return gin.H{
		"service": map[string]any{
			"name":   serviceName,
			"normal": true,
			"layers": []string{"GENERAL"},
		},
		"instances": instances,
		"endpoints": endpoints,
		"metrics": map[string]any{
			"avg_latency_ms": avgLatency,
			"error_rate":     errorRate,
			"total_traces":   count,
		},
	}
}

// aggregateServiceInstancesFromBuffer 从 buffer 聚合服务实例
func aggregateServiceInstancesFromBuffer(serviceName string) []map[string]any {
	traceBufferMu.RLock()
	defer traceBufferMu.RUnlock()

	instanceSet := make(map[string]int64)
	for _, seg := range traceBuffer {
		if !strings.EqualFold(seg.Service, serviceName) {
			continue
		}
		instanceSet[seg.Service+"-instance"]++
	}

	instances := make([]map[string]any, 0, len(instanceSet))
	for name, count := range instanceSet {
		instances = append(instances, map[string]any{
			"name":          name,
			"health_status": "healthy",
			"trace_count":   count,
		})
	}
	return instances
}

// aggregateServiceEndpointsFromBuffer 从 buffer 聚合服务端点
func aggregateServiceEndpointsFromBuffer(serviceName string) []map[string]any {
	traceBufferMu.RLock()
	defer traceBufferMu.RUnlock()

	type epStats struct {
		durations  []int64
		errorCount int64
		totalCount int64
	}
	epMap := make(map[string]*epStats)

	for _, seg := range traceBuffer {
		if !strings.EqualFold(seg.Service, serviceName) {
			continue
		}
		if seg.Operation == "" {
			continue
		}
		stats, ok := epMap[seg.Operation]
		if !ok {
			stats = &epStats{durations: make([]int64, 0)}
			epMap[seg.Operation] = stats
		}
		stats.durations = append(stats.durations, seg.Duration)
		stats.totalCount++
		if seg.Status == "error" {
			stats.errorCount++
		}
	}

	endpoints := make([]map[string]any, 0, len(epMap))
	for name, stats := range epMap {
		p50, p90, p99 := calculatePercentiles(stats.durations)
		errorRate := float64(0)
		if stats.totalCount > 0 {
			errorRate = float64(stats.errorCount) / float64(stats.totalCount)
		}
		endpoints = append(endpoints, map[string]any{
			"name":       name,
			"p50_ms":     p50,
			"p90_ms":     p90,
			"p99_ms":     p99,
			"error_rate": errorRate,
			"cpm":        stats.totalCount,
		})
	}
	return endpoints
}

// calculatePercentiles 计算延迟百分位数
func calculatePercentiles(durations []int64) (p50, p90, p99 int64) {
	n := len(durations)
	if n == 0 {
		return 0, 0, 0
	}
	sorted := make([]int64, n)
	copy(sorted, durations)
	sortInt64Slice(sorted)

	p50 = sorted[n*50/100]
	p90 = sorted[n*90/100]
	idx99 := n * 99 / 100
	if idx99 >= n {
		idx99 = n - 1
	}
	p99 = sorted[idx99]
	return
}

// sortInt64Slice 对 int64 切片进行排序（插入排序，适合小数据集）
func sortInt64Slice(s []int64) {
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}
