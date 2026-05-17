package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"ai-workbench-api/internal/tracing"

	"github.com/gin-gonic/gin"
)

var swClient = tracing.NewSWClient()

func bindOptionalJSON(c *gin.Context, body *map[string]any) bool {
	if c.Request.Body == nil {
		*body = map[string]any{}
		return true
	}
	if c.Request.ContentLength == 0 {
		*body = map[string]any{}
		return true
	}
	if err := c.ShouldBindJSON(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return false
	}
	if *body == nil {
		*body = map[string]any{}
	}
	return true
}

func queryStringValue(c *gin.Context, body map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			return value
		}
		if raw, ok := body[key]; ok {
			if value := strings.TrimSpace(toString(raw)); value != "" {
				return value
			}
		}
	}
	return ""
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	default:
		return ""
	}
}

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
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking services query failed: " + err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "serviceId is required"})
		return
	}
	q := `query($serviceId: ID!, $keyword: String!, $duration: Duration, $limit: Int!) { endpoints: findEndpoint(serviceId: $serviceId, keyword: $keyword, limit: $limit, duration: $duration) { id name value: name label: name } }`
	var out struct {
		Endpoints []map[string]any `json:"endpoints"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"serviceId": svcID, "keyword": keyword, "duration": durationFor(30), "limit": limit}, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking endpoints query failed: " + err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "serviceId is required"})
		return
	}
	q := `query($serviceId: ID!, $duration: Duration!) { instances: listInstances(duration: $duration, serviceId: $serviceId) { id name value: name label: name instanceUUID language attributes { name value } } }`
	var out struct {
		Instances []map[string]any `json:"instances"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"serviceId": svcID, "duration": durationFor(30)}, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking instances query failed: " + err.Error()})
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
	if !bindOptionalJSON(c, &body) {
		return
	}

	condition := map[string]any{
		"queryDuration": durationFor(30),
		"queryOrder":    "BY_START_TIME",
		"traceState":    "ALL",
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
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking traces query failed: " + err.Error()})
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
	if strings.TrimSpace(traceID) == "" {
		traceID = strings.TrimSpace(c.Param("traceId"))
	}
	if strings.TrimSpace(traceID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trace id is required"})
		return
	}
	q := `query($traceId: ID!) { trace: queryTrace(traceId: $traceId) { spans { traceId segmentId spanId parentSpanId refs { traceId parentSegmentId parentSpanId type } serviceCode serviceInstanceName endpointName startTime endTime type peer component isError layer tags { key value } logs { time data { key value } } attachedEvents { startTime { seconds nanos } event endTime { seconds nanos } tags { key value } summary { key value } } } } }`
	var out struct {
		Trace struct {
			Spans []map[string]any `json:"spans"`
		} `json:"trace"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"traceId": traceID}, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking trace spans query failed: " + err.Error()})
		return
	}
	if out.Trace.Spans == nil {
		out.Trace.Spans = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"spans": out.Trace.Spans})
}

// TracingGetTopologySW answers GET /api/v1/tracing/topology.
func TracingGetTopologySW(c *gin.Context) {
	svcID := strings.TrimSpace(c.Query("serviceId"))
	endpointID := strings.TrimSpace(c.Query("endpointId"))
	instanceID := strings.TrimSpace(c.Query("instanceId"))

	var q string
	vars := map[string]any{"duration": durationFor(30)}

	switch {
	case svcID != "" && endpointID != "":
		q = `query($endpointId: ID!, $duration: Duration!) { topology: getEndpointDependencies(endpointId: $endpointId, duration: $duration) { nodes { id name type isReal } calls { id source target detectPoints } } }`
		vars["endpointId"] = endpointID
	case svcID != "" && instanceID != "":
		q = `query($clientServiceId: ID!, $serverServiceId: ID!, $duration: Duration!) { topology: getServiceInstanceTopology(clientServiceId: $clientServiceId, serverServiceId: $serverServiceId, duration: $duration) { nodes { id name type isReal } calls { id source target detectPoints } } }`
		vars["clientServiceId"] = svcID
		vars["serverServiceId"] = svcID
	case svcID != "":
		q = `query($serviceId: ID!, $duration: Duration!) { topology: getServiceTopology(serviceId: $serviceId, duration: $duration) { nodes { id name type isReal } calls { id source target detectPoints } } }`
		vars["serviceId"] = svcID
	default:
		q = `query($duration: Duration!) { topology: getGlobalTopology(duration: $duration) { nodes { id name type isReal } calls { id source target detectPoints } } }`
	}

	var out struct {
		Topology struct {
			Nodes []map[string]any `json:"nodes"`
			Calls []map[string]any `json:"calls"`
		} `json:"topology"`
	}
	if err := swClient.Query(c.Request.Context(), q, vars, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking topology query failed: " + err.Error()})
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

// APMQueryTracesSW answers GET/POST /api/v1/apm/traces.
func APMQueryTracesSW(c *gin.Context) {
	TracingQueryTracesSW(c)
}

// APMGetTraceSW answers GET /api/v1/apm/traces/:traceId.
func APMGetTraceSW(c *gin.Context) {
	traceID := strings.TrimSpace(c.Param("traceId"))
	if traceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "traceId is required"})
		return
	}
	q := `query($traceId: ID!) { trace: queryTrace(traceId: $traceId) { spans { traceId segmentId spanId parentSpanId refs { traceId parentSegmentId parentSpanId type } serviceCode serviceInstanceName endpointName startTime endTime type peer component isError layer tags { key value } logs { time data { key value } } } } }`
	var out struct {
		Trace struct {
			Spans []map[string]any `json:"spans"`
		} `json:"trace"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"traceId": traceID}, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking trace detail query failed: " + err.Error()})
		return
	}
	if out.Trace.Spans == nil {
		out.Trace.Spans = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"spans": out.Trace.Spans})
}

// APMTraceTagKeysSW proxies the mature trace tag-key autocomplete query.
func APMTraceTagKeysSW(c *gin.Context) {
	q := `query($duration: Duration!) { tagKeys: queryTraceTagAutocompleteKeys(duration: $duration) }`
	var out struct {
		TagKeys []string `json:"tagKeys"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"duration": durationFor(30)}, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking trace tag keys query failed: " + err.Error()})
		return
	}
	if out.TagKeys == nil {
		out.TagKeys = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"tagKeys": out.TagKeys})
}

// APMTraceTagValuesSW proxies the mature trace tag-value autocomplete query.
func APMTraceTagValuesSW(c *gin.Context) {
	tagKey := strings.TrimSpace(c.Query("tagKey"))
	if tagKey == "" {
		tagKey = strings.TrimSpace(c.Query("key"))
	}
	if tagKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tagKey is required"})
		return
	}
	q := `query($tagKey: String!, $duration: Duration!) { tagValues: queryTraceTagAutocompleteValues(tagKey: $tagKey, duration: $duration) }`
	var out struct {
		TagValues []string `json:"tagValues"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"tagKey": tagKey, "duration": durationFor(30)}, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking trace tag values query failed: " + err.Error()})
		return
	}
	if out.TagValues == nil {
		out.TagValues = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"tagValues": out.TagValues})
}

// APMGetSpanDetailSW answers GET /api/v1/apm/traces/:traceId/spans/:spanId.
func APMGetSpanDetailSW(c *gin.Context) {
	traceID := strings.TrimSpace(c.Param("traceId"))
	spanID := strings.TrimSpace(c.Param("spanId"))
	if traceID == "" || spanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "traceId and spanId are required"})
		return
	}
	// 查询完整 trace 后按 spanId 过滤
	q := `query($traceId: ID!) { trace: queryTrace(traceId: $traceId) { spans { traceId segmentId spanId parentSpanId refs { traceId parentSegmentId parentSpanId type } serviceCode serviceInstanceName endpointName startTime endTime type peer component isError layer tags { key value } logs { time data { key value } } } } }`
	var out struct {
		Trace struct {
			Spans []map[string]any `json:"spans"`
		} `json:"trace"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"traceId": traceID}, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking span detail query failed: " + err.Error()})
		return
	}
	for _, span := range out.Trace.Spans {
		if toString(span["spanId"]) == spanID {
			c.JSON(http.StatusOK, gin.H{"span": span})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "span not found in trace"})
}

// APMListProfilingTasksSW queries profiling tasks from SkyWalking OAP.
func APMListProfilingTasksSW(c *gin.Context) {
	svcID := strings.TrimSpace(c.Query("serviceId"))
	q := `query($serviceId: ID, $endpointName: String) { tasks: getProfileTaskList(serviceId: $serviceId, endpointName: $endpointName) { id serviceId endpointName startTime duration minDurationThreshold dumpPeriod maxSamplingCount } }`
	var out struct {
		Tasks []map[string]any `json:"tasks"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"serviceId": svcID, "endpointName": ""}, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking profiling tasks query failed: " + err.Error()})
		return
	}
	if out.Tasks == nil {
		out.Tasks = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"tasks": out.Tasks})
}

// APMCreateProfilingTaskSW creates a profiling task in SkyWalking OAP.
func APMCreateProfilingTaskSW(c *gin.Context) {
	var body map[string]any
	if !bindOptionalJSON(c, &body) {
		return
	}
	svcID := strings.TrimSpace(queryStringValue(c, body, "serviceId", "service_id"))
	if svcID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "serviceId is required"})
		return
	}
	q := `mutation($request: ProfileTaskCreationRequest!) { task: createProfileTask(creationRequest: $request) { id } }`
	request := map[string]any{
		"serviceId": svcID,
	}
	for _, k := range []string{"endpointName", "startTime", "duration", "minDurationThreshold", "dumpPeriod", "maxSamplingCount"} {
		if v, ok := body[k]; ok && v != nil {
			request[k] = v
		}
	}
	var out struct {
		Task map[string]any `json:"task"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"request": request}, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking profiling task creation failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": out.Task})
}

// APMCancelProfilingTaskSW cancels a profiling task (SkyWalking OAP does not support cancel, return 501).
func APMCancelProfilingTaskSW(c *gin.Context) {
	if strings.TrimSpace(c.Param("id")) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profiling task id is required"})
		return
	}
	c.JSON(http.StatusNotImplemented, gin.H{"error": "SkyWalking OAP 不支持取消 profiling 任务，任务将在到期后自动结束"})
}

// APMListAlarmsSW queries alarms from SkyWalking OAP.
func APMListAlarmsSW(c *gin.Context) {
	q := `query($duration: Duration!, $paging: Pagination!) { alarms: getAlarm(duration: $duration, paging: $paging) { msgs { id message startTime scope { scope0: scope } tags { key value } } total } }`
	var out struct {
		Alarms struct {
			Msgs  []map[string]any `json:"msgs"`
			Total int              `json:"total"`
		} `json:"alarms"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"duration": durationFor(60), "paging": map[string]any{"pageNum": 1, "pageSize": 50}}, &out); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SkyWalking alarms query failed: " + err.Error()})
		return
	}
	if out.Alarms.Msgs == nil {
		out.Alarms.Msgs = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"alarms": out.Alarms.Msgs, "total": out.Alarms.Total})
}

// APMAckAlarmSW acknowledges an alarm (SkyWalking OAP 无原生 ack 接口，返回 501).
func APMAckAlarmSW(c *gin.Context) {
	if strings.TrimSpace(c.Param("id")) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alarm id is required"})
		return
	}
	c.JSON(http.StatusNotImplemented, gin.H{"error": "SkyWalking OAP 不支持告警确认操作，请在告警规则中配置静默策略"})
}

// APMGetSettingsSW returns SkyWalking adapter connection status.
func APMGetSettingsSW(c *gin.Context) {
	configured := swClient != nil
	c.JSON(http.StatusOK, gin.H{
		"configured": configured,
		"adapter":    "skywalking",
		"message":    "SkyWalking OAP 连接配置通过环境变量管理，此接口仅返回连接状态",
	})
}

// APMPutSettingsSW updates SkyWalking adapter settings (managed via env, return 501).
func APMPutSettingsSW(c *gin.Context) {
	var body map[string]any
	if !bindOptionalJSON(c, &body) {
		return
	}
	c.JSON(http.StatusNotImplemented, gin.H{"error": "SkyWalking OAP 连接配置通过环境变量管理，不支持运行时修改"})
}

// APMAgentLinkageSW returns agent linkage status from evidence chain.
func APMAgentLinkageSW(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Agent 安装关联需要通过 FindX Agent 管理模块配置，请参考 /api/v1/findx-agent/install-plans"})
}
