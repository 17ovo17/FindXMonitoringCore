package handler

import (
	"net/http"
	"strconv"
	"time"

	"ai-workbench-api/internal/tracing"

	"github.com/gin-gonic/gin"
)

var swClient = tracing.NewSWClient()

// durationFor returns a SkyWalking Duration struct for the last N minutes.
func durationFor(minutes int) map[string]any {
	end := time.Now().UTC()
	start := end.Add(-time.Duration(minutes) * time.Minute)
	return map[string]any{
		"start": start.Format("2006-01-02 1504"),
		"end":   end.Format("2006-01-02 1504"),
		"step":  "MINUTE",
	}
}

// TracingListServicesSW answers GET /api/v1/tracing/selectors/services.
func TracingListServicesSW(c *gin.Context) {
	layer := c.DefaultQuery("layer", "GENERAL")
	q := `query($layer: String!) { services: listServices(layer: $layer) { id name group shortName layers normal } }`
	var out struct {
		Services []map[string]any `json:"services"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"layer": layer}, &out); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "services": []any{}})
		return
	}
	if out.Services == nil {
		out.Services = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"services": out.Services})
}

// TracingListEndpointsSW answers GET /api/v1/tracing/selectors/endpoints.
func TracingListEndpointsSW(c *gin.Context) {
	svcID := c.Query("serviceId")
	keyword := c.Query("keyword")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit <= 0 {
		limit = 100
	}
	if svcID == "" {
		c.JSON(http.StatusOK, gin.H{"endpoints": []any{}})
		return
	}
	q := `query($serviceId: ID!, $keyword: String!, $limit: Int!) { endpoints: getEndpoints(serviceId: $serviceId, keyword: $keyword, limit: $limit) { id name } }`
	var out struct {
		Endpoints []map[string]any `json:"endpoints"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"serviceId": svcID, "keyword": keyword, "limit": limit}, &out); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "endpoints": []any{}})
		return
	}
	if out.Endpoints == nil {
		out.Endpoints = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"endpoints": out.Endpoints})
}

// TracingListInstancesSW answers GET /api/v1/tracing/selectors/instances.
func TracingListInstancesSW(c *gin.Context) {
	svcID := c.Query("serviceId")
	if svcID == "" {
		c.JSON(http.StatusOK, gin.H{"instances": []any{}})
		return
	}
	q := `query($serviceId: ID!, $duration: Duration!) { instances: listInstances(duration: $duration, serviceId: $serviceId) { id name instanceUUID language attributes { name value } } }`
	var out struct {
		Instances []map[string]any `json:"instances"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"serviceId": svcID, "duration": durationFor(30)}, &out); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "instances": []any{}})
		return
	}
	if out.Instances == nil {
		out.Instances = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"instances": out.Instances})
}

// TracingQueryTracesSW answers POST /api/v1/tracing/traces/query.
func TracingQueryTracesSW(c *gin.Context) {
	var body map[string]any
	_ = c.ShouldBindJSON(&body)

	condition := map[string]any{
		"queryDuration": durationFor(30),
		"queryOrder":    "BY_START_TIME",
		"traceState":   "ALL",
		"paging":        map[string]any{"pageNum": 1, "pageSize": 20},
	}
	for _, k := range []string{"serviceId", "endpointId", "traceId", "minDuration", "maxDuration", "queryOrder", "traceState"} {
		if v, ok := body[k]; ok && v != nil && v != "" {
			condition[k] = v
		}
	}
	if p, ok := body["paging"]; ok && p != nil {
		condition["paging"] = p
	}
	if qd, ok := body["queryDuration"]; ok && qd != nil {
		condition["queryDuration"] = qd
	}

	q := `query($condition: TraceQueryCondition) { result: queryBasicTraces(condition: $condition) { traces { segmentId endpointNames duration start isError traceIds } } }`
	var out struct {
		Result struct {
			Traces []map[string]any `json:"traces"`
			Total  int              `json:"total"`
		} `json:"result"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"condition": condition}, &out); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "traces": []any{}, "total": 0})
		return
	}
	if out.Result.Traces == nil {
		out.Result.Traces = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"traces": out.Result.Traces, "total": out.Result.Total})
}

// TracingGetTraceSpansSW answers GET /api/v1/tracing/traces/:id/spans.
func TracingGetTraceSpansSW(c *gin.Context) {
	traceID := c.Param("id")
	q := `query($traceId: ID!) { trace: queryTrace(traceId: $traceId) { spans { traceId segmentId spanId parentSpanId refs { traceId parentSegmentId parentSpanId type } serviceCode endpointName startTime endTime type peer component isError tags { key value } logs { time data { key value } } } } }`
	var out struct {
		Trace struct {
			Spans []map[string]any `json:"spans"`
		} `json:"trace"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"traceId": traceID}, &out); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "spans": []any{}})
		return
	}
	if out.Trace.Spans == nil {
		out.Trace.Spans = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"spans": out.Trace.Spans})
}

// TracingGetTopologySW answers GET /api/v1/tracing/topology.
func TracingGetTopologySW(c *gin.Context) {
	q := `query($duration: Duration!) { topology: getGlobalTopology(duration: $duration) { nodes { id name type isReal } calls { id source target detectPoints } } }`
	var out struct {
		Topology struct {
			Nodes []map[string]any `json:"nodes"`
			Calls []map[string]any `json:"calls"`
		} `json:"topology"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"duration": durationFor(30)}, &out); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "nodes": []any{}, "calls": []any{}})
		return
	}
	if out.Topology.Nodes == nil {
		out.Topology.Nodes = []map[string]any{}
	}
	if out.Topology.Calls == nil {
		out.Topology.Calls = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"nodes": out.Topology.Nodes, "calls": out.Topology.Calls})
}
