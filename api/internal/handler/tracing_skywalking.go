package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ai-workbench-api/internal/tracing"

	"github.com/gin-gonic/gin"
)

var swClient = tracing.NewSWClient()

const apmBlockedCode = "BLOCKED_BY_CONTRACT"

type apmBlockedContract struct {
	ContractID       string   `json:"contract_id"`
	MissingContracts []string `json:"missing_contracts"`
}

func writeAPMBlocked(c *gin.Context, httpStatus int, contractID string, missing []string, err error) {
	if len(missing) == 0 {
		missing = []string{"findx_tracing_query_upstream_contract"}
	}
	message := "FindX 链路监控上游契约不可用"
	if err != nil && !errors.Is(err, tracing.ErrNotConfigured) {
		message = "FindX 链路监控上游请求失败"
	}
	c.JSON(httpStatus, gin.H{
		"error":             message,
		"code":              apmBlockedCode,
		"status":            "blocked",
		"contract_id":       contractID,
		"missing_contracts": missing,
		"safe_to_retry":     false,
		"data": apmBlockedContract{
			ContractID:       contractID,
			MissingContracts: missing,
		},
	})
}

func writeAPMContractBlocked(c *gin.Context, contractID string, missing []string) {
	writeAPMBlocked(c, http.StatusConflict, contractID, missing, tracing.ErrNotConfigured)
}

func writeAPMUpstreamBlocked(c *gin.Context, contractID string, err error) {
	writeAPMBlocked(c, http.StatusServiceUnavailable, contractID, []string{"findx_tracing_query_upstream_contract"}, err)
}

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
		writeAPMUpstreamBlocked(c, "FX-CONTRACT-TRACING-SERVICES", err)
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
		writeAPMUpstreamBlocked(c, "FX-CONTRACT-TRACING-ENDPOINTS", err)
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
		writeAPMUpstreamBlocked(c, "FX-CONTRACT-TRACING-INSTANCES", err)
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
		writeAPMUpstreamBlocked(c, "FX-CONTRACT-TRACING-TRACES", err)
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
		writeAPMUpstreamBlocked(c, "FX-CONTRACT-TRACING-TRACE-SPANS", err)
		return
	}
	if out.Trace.Spans == nil {
		out.Trace.Spans = []map[string]any{}
	}
	c.JSON(http.StatusOK, gin.H{"spans": out.Trace.Spans})
}

// TracingGetTopologySW answers GET /api/v1/tracing/topology.
func TracingGetTopologySW(c *gin.Context) {
	if strings.TrimSpace(c.Query("serviceId")) != "" ||
		strings.TrimSpace(c.Query("endpointId")) != "" ||
		strings.TrimSpace(c.Query("instanceId")) != "" ||
		strings.TrimSpace(c.Query("processId")) != "" ||
		strings.TrimSpace(c.Query("hierarchy")) != "" ||
		strings.TrimSpace(c.Query("layerLevels")) != "" {
		writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-TOPOLOGY-SCOPE", []string{"findx_tracing_scoped_topology_contract"})
		return
	}
	q := `query($duration: Duration!) { topology: getGlobalTopology(duration: $duration) { nodes { id name type isReal } calls { id source target detectPoints } } }`
	var out struct {
		Topology struct {
			Nodes []map[string]any `json:"nodes"`
			Calls []map[string]any `json:"calls"`
		} `json:"topology"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"duration": durationFor(30)}, &out); err != nil {
		writeAPMUpstreamBlocked(c, "FX-CONTRACT-TRACING-TOPOLOGY", err)
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
	if strings.TrimSpace(c.Param("traceId")) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "traceId is required"})
		return
	}
	writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-TRACE-DETAIL", []string{"findx_tracing_trace_detail_query_contract", "findx_trace_log_agent_linkage_contract"})
}

// APMTraceTagKeysSW proxies the mature trace tag-key autocomplete query.
func APMTraceTagKeysSW(c *gin.Context) {
	q := `query($duration: Duration!) { tagKeys: queryTraceTagAutocompleteKeys(duration: $duration) }`
	var out struct {
		TagKeys []string `json:"tagKeys"`
	}
	if err := swClient.Query(c.Request.Context(), q, map[string]any{"duration": durationFor(30)}, &out); err != nil {
		writeAPMUpstreamBlocked(c, "FX-CONTRACT-TRACING-TRACE-TAG-KEYS", err)
		return
	}
	if out.TagKeys == nil {
		writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-TRACE-TAG-KEYS-DATA", []string{"findx_tracing_trace_tag_keys_data_contract"})
		return
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
		writeAPMUpstreamBlocked(c, "FX-CONTRACT-TRACING-TRACE-TAG-VALUES", err)
		return
	}
	if out.TagValues == nil {
		writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-TRACE-TAG-VALUES-DATA", []string{"findx_tracing_trace_tag_values_data_contract"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tagValues": out.TagValues})
}

// APMGetSpanDetailSW blocks span detail until the OAP span-detail contract exists.
func APMGetSpanDetailSW(c *gin.Context) {
	if strings.TrimSpace(c.Param("traceId")) == "" || strings.TrimSpace(c.Param("spanId")) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "traceId and spanId are required"})
		return
	}
	writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-SPAN-DETAIL", []string{"findx_tracing_span_detail_query_contract"})
}

// APMListProfilingTasksSW blocks profiling reads until FindX maps the real OAP contract.
func APMListProfilingTasksSW(c *gin.Context) {
	writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-PROFILING-TASKS", []string{"findx_tracing_profile_task_contract"})
}

// APMCreateProfilingTaskSW blocks profiling mutation without faking lifecycle status.
func APMCreateProfilingTaskSW(c *gin.Context) {
	var body map[string]any
	if !bindOptionalJSON(c, &body) {
		return
	}
	if strings.TrimSpace(queryStringValue(c, body, "serviceId", "service_id")) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "serviceId is required"})
		return
	}
	writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-PROFILING-CREATE", []string{"findx_tracing_profile_task_mutation_contract", "findx_profiling_audit_receipt_contract"})
}

// APMCancelProfilingTaskSW blocks profiling cancellation without faking lifecycle status.
func APMCancelProfilingTaskSW(c *gin.Context) {
	if strings.TrimSpace(c.Param("id")) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profiling task id is required"})
		return
	}
	writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-PROFILING-CANCEL", []string{"findx_tracing_profile_cancel_mutation_contract", "findx_profiling_audit_receipt_contract"})
}

// APMListAlarmsSW blocks alarm reads until the OAP alarm contract is mapped.
func APMListAlarmsSW(c *gin.Context) {
	writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-ALARMS", []string{"findx_tracing_alarm_query_contract"})
}

// APMAckAlarmSW blocks alarm ack mutation until a real upstream mutation exists.
func APMAckAlarmSW(c *gin.Context) {
	if strings.TrimSpace(c.Param("id")) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alarm id is required"})
		return
	}
	writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-ALARM-ACK", []string{"findx_tracing_alarm_ack_mutation_contract", "findx_alarm_audit_receipt_contract"})
}

// APMGetSettingsSW returns safe adapter settings without exposing endpoint secrets.
func APMGetSettingsSW(c *gin.Context) {
	writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-SETTINGS", []string{"findx_tracing_settings_read_contract", "findx_settings_audit_receipt_contract"})
}

// APMPutSettingsSW blocks settings writes until persistence and audit contracts exist.
func APMPutSettingsSW(c *gin.Context) {
	var body map[string]any
	if !bindOptionalJSON(c, &body) {
		return
	}
	writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-SETTINGS-WRITE", []string{"findx_tracing_settings_store_contract", "findx_settings_audit_receipt_contract"})
}

// APMAgentLinkageSW blocks linkage until Agent lifecycle evidence is contractually connected.
func APMAgentLinkageSW(c *gin.Context) {
	writeAPMContractBlocked(c, "FX-CONTRACT-TRACING-AGENT-LINKAGE", []string{"findx_agent_installation_linkage_contract", "findx_tracing_service_instance_correlation_contract", "data_arrival_evidence_contract"})
}
