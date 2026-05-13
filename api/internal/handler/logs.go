package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func ListLogFields(c *gin.Context) {
	c.JSON(http.StatusOK, builtinLogFields())
}

func ListLogPipelines(c *gin.Context) {
	items, err := store.ListLogPipelines(c.Param("version"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "log pipeline list failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items":  sanitizeLogPipelines(items),
		"status": "partial",
		"capabilities": gin.H{
			"list":     logsCapabilityOK(),
			"save":     logsCapabilityOK(),
			"preview":  logsCapabilityOK(),
			"deploy":   logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-PIPELINE-DEPLOY", "logs.pipeline.deploy"),
			"rollback": logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-PIPELINE-ROLLBACK", "logs.pipeline.rollback"),
			"update":   logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-PIPELINE-MUTATION", "logs.pipeline.update"),
			"delete":   logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-PIPELINE-MUTATION", "logs.pipeline.delete"),
		},
		"blocker": model.LogsContractBlocked + ": log pipeline deployment, rollback, and index mutation contracts are not connected",
	})
}

func SaveLogPipeline(c *gin.Context) {
	var input model.LogPipelineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log pipeline payload"})
		return
	}
	item := logPipelineFromInput(input)
	out, err := store.SaveLogPipeline(&item, requestActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log pipeline"})
		return
	}
	c.JSON(http.StatusOK, sanitizeLogPipeline(*out))
}

func PreviewLogPipeline(c *gin.Context) {
	var req model.LogPipelinePreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log pipeline preview payload"})
		return
	}
	result := previewLogSamples(req)
	c.JSON(http.StatusOK, result)
}

func UpdateLogPipelineBlocked(c *gin.Context) {
	blockLogsContract(c, http.StatusConflict, "FX-CONTRACT-SIGNOZ-LOGS-PIPELINE-MUTATION", []string{"logs.pipeline.update", "logs.pipeline.deploy", "logs.pipeline.versioning"}, "log pipeline update requires SigNoZ pipeline mutation and deployment contracts")
}

func DeleteLogPipelineBlocked(c *gin.Context) {
	blockLogsContract(c, http.StatusConflict, "FX-CONTRACT-SIGNOZ-LOGS-PIPELINE-MUTATION", []string{"logs.pipeline.delete", "logs.pipeline.deploy", "logs.pipeline.audit"}, "log pipeline delete requires SigNoZ pipeline mutation and deployment contracts")
}

func DeployLogPipelineBlocked(c *gin.Context) {
	blockLogsContract(c, http.StatusConflict, "FX-CONTRACT-SIGNOZ-LOGS-PIPELINE-DEPLOY", []string{"logs.pipeline.deploy", "logs.pipeline.executor", "logs.pipeline.evidence"}, "saved pipeline config is not a deployed processing pipeline")
}

func RollbackLogPipelineBlocked(c *gin.Context) {
	blockLogsContract(c, http.StatusConflict, "FX-CONTRACT-SIGNOZ-LOGS-PIPELINE-ROLLBACK", []string{"logs.pipeline.rollback", "logs.pipeline.history", "logs.pipeline.evidence"}, "log pipeline rollback requires version history and executor contracts")
}

func ListLogsBlocked(c *gin.Context) {
	source := normalizeLogSource(c.Query("source"))
	if source != model.LogsSourceFindXAudit {
		blockLogsContract(c, http.StatusServiceUnavailable, "FX-CONTRACT-SIGNOZ-LOGS-QUERY", []string{"logs.query", "logs.query_service", "logs.clickhouse.datasource"}, "generic logs query datasource contract is not connected; only findx_audit is queryable")
		return
	}
	resp, err := store.QueryFindXAuditLogs(logQueryRequest(c, source))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "FindX audit log query failed"})
		return
	}
	c.JSON(http.StatusOK, sanitizeUserFacingLogQueryResponse(resp))
}

func AggregateLogsBlocked(c *gin.Context) {
	source := normalizeLogSource(c.Query("source"))
	if source != model.LogsSourceFindXAudit {
		blockLogsContract(c, http.StatusConflict, "FX-CONTRACT-SIGNOZ-LOGS-AGGREGATE", []string{"logs.aggregate", "logs.query_builder", "logs.clickhouse.datasource"}, "generic logs aggregate datasource contract is not connected; only findx_audit can aggregate")
		return
	}
	resp, err := store.AggregateFindXAuditLogs(logAggregateRequest(c, source))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "FindX audit log aggregate failed"})
		return
	}
	c.JSON(http.StatusOK, sanitizeUserFacingLogAggregateResponse(resp))
}

func GetLogContext(c *gin.Context) {
	source := normalizeLogSource(c.Query("source"))
	if source != model.LogsSourceFindXAudit {
		blockLogsContract(c, http.StatusConflict, "FX-CONTRACT-SIGNOZ-LOGS-CONTEXT", []string{"logs.context", "logs.trace_correlation", "logs.query_service"}, "generic logs context datasource contract is not connected; only findx_audit context is queryable")
		return
	}
	req := logContextRequest(c, source)
	if req.LogID == "" && req.TraceID == "" && req.Scope == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "log_id, trace_id or scope is required"})
		return
	}
	resp, err := store.ContextFindXAuditLogs(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "FindX audit log context failed"})
		return
	}
	c.JSON(http.StatusOK, sanitizeUserFacingLogContextResponse(resp))
}

func RealtimeLogsBlocked(c *gin.Context) {
	blockLogsContract(c, http.StatusServiceUnavailable, "FX-CONTRACT-SIGNOZ-LOGS-LIVE", []string{"logs.live_tail", "logs.websocket", "logs.query_service"}, "realtime log backend is not connected")
}

func ListExplorerViews(c *gin.Context) {
	items, err := store.ListExplorerSavedViews(c.Query("sourcePage"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "saved view list failed"})
		return
	}
	c.JSON(http.StatusOK, sanitizeExplorerViews(items))
}

func GetExplorerView(c *gin.Context) {
	item, ok, err := store.GetExplorerSavedView(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "saved view get failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "saved view not found"})
		return
	}
	c.JSON(http.StatusOK, sanitizeExplorerView(*item))
}

func CreateExplorerView(c *gin.Context) {
	saveExplorerView(c, "")
}

func UpdateExplorerView(c *gin.Context) {
	saveExplorerView(c, c.Param("id"))
}

func DeleteExplorerView(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		var req struct {
			ID string `json:"id"`
		}
		_ = c.ShouldBindJSON(&req)
		id = strings.TrimSpace(req.ID)
	}
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "saved view id is required"})
		return
	}
	ok, err := store.DeleteExplorerSavedView(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "saved view delete failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "saved view not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func saveExplorerView(c *gin.Context, id string) {
	var input model.ExplorerSavedViewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid saved view payload"})
		return
	}
	item := explorerViewFromInput(input)
	if id != "" {
		item.ID = id
	}
	var out *model.ExplorerSavedView
	var ok = true
	var err error
	if id == "" {
		out, err = store.SaveExplorerSavedView(&item, requestActor(c))
	} else {
		out, ok, err = store.UpdateExplorerSavedView(&item, requestActor(c))
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid saved view"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "saved view not found"})
		return
	}
	c.JSON(http.StatusOK, sanitizeExplorerView(*out))
}

func logPipelineFromInput(input model.LogPipelineInput) model.LogPipeline {
	enabled := false
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	return model.LogPipeline{
		ID:          input.ID,
		Name:        input.Name,
		Version:     input.Version,
		Description: input.Description,
		Enabled:     enabled,
		Stages:      defaultRaw(input.Stages, `[]`),
		Config:      defaultRaw(input.Config, `{}`),
	}
}

func explorerViewFromInput(input model.ExplorerSavedViewInput) model.ExplorerSavedView {
	return model.ExplorerSavedView{
		ID:          input.ID,
		SourcePage:  input.SourcePage,
		Name:        input.Name,
		Description: input.Description,
		Query:       defaultRaw(input.Query, `{}`),
		Filters:     defaultRaw(input.Filters, `{}`),
		Columns:     defaultRaw(input.Columns, `[]`),
		TimeRange:   defaultRaw(input.TimeRange, `{}`),
		Layout:      defaultRaw(input.Layout, `{}`),
	}
}

func blockLogsContract(c *gin.Context, status int, contractGapID string, missingContracts []string, reason string) {
	c.JSON(status, logsBlockedEnvelope(contractGapID, missingContracts, reason))
}

func logsBlockedEnvelope(contractGapID string, missingContracts []string, reason string) model.LogsBlockedEnvelope {
	return model.LogsBlockedEnvelope{
		Code:             model.LogsContractBlocked,
		Status:           "blocked",
		ContractID:       contractGapID,
		ContractGapID:    contractGapID,
		MissingContracts: missingContracts,
		SafeToRetry:      false,
		Error:            model.LogsContractBlocked + ": " + sanitizeLogString(reason),
		Blocker:          model.LogsContractBlocked,
	}
}

func logsCapabilityOK() model.LogCapabilityState {
	return model.LogCapabilityState{Status: "ready", SafeToRetry: true}
}

func logsCapabilityBlocked(contractGapID string, missingContracts ...string) model.LogCapabilityState {
	return model.LogCapabilityState{
		Status:           "blocked",
		ContractGapID:    contractGapID,
		MissingContracts: missingContracts,
		SafeToRetry:      false,
	}
}

func normalizeLogSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case model.LogsSourceFindXAudit:
		return model.LogsSourceFindXAudit
	case model.LogsSourceOtel:
		return model.LogsSourceOtel
	default:
		return model.LogsSourceOtel
	}
}

func logQueryRequest(c *gin.Context, source string) model.LogQueryRequest {
	return model.LogQueryRequest{
		Source: source, Query: firstNonEmptyQuery(c, "query", "q"),
		Page: queryInt(c, "page", 1), Limit: queryInt(c, "limit", 50),
		Status: c.Query("status"), Action: c.Query("action"), ResourceType: c.Query("resource_type"),
		ResourceID: c.Query("resource_id"), TraceID: c.Query("trace_id"), Scope: c.Query("scope"),
	}
}

func logAggregateRequest(c *gin.Context, source string) model.LogAggregateRequest {
	return model.LogAggregateRequest{
		Source: source, GroupBy: firstNonEmptyQuery(c, "group_by", "groupBy"),
		Page: queryInt(c, "page", 1), Limit: queryInt(c, "limit", 100),
		Status: c.Query("status"), Action: c.Query("action"), ResourceType: c.Query("resource_type"),
		Scope: c.Query("scope"),
	}
}

func logContextRequest(c *gin.Context, source string) model.LogContextRequest {
	return model.LogContextRequest{
		Source: source, LogID: firstNonEmptyQuery(c, "log_id", "logId", "id"),
		TraceID: firstNonEmptyQuery(c, "trace_id", "traceId"), Scope: c.Query("scope"),
		Before: queryInt(c, "before", 5), After: queryInt(c, "after", 5),
	}
}

func firstNonEmptyQuery(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			return value
		}
	}
	return ""
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(c.Query(key)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func builtinLogFields() model.LogFieldsResponse {
	categories := []model.LogFieldCategory{
		logFieldCategory("resource", "Resource", "Resource identity fields.", []model.LogField{
			logField("service.name", "string", "resource", "Logical service name.", true, "checkout-api"),
			logField("service.namespace", "string", "resource", "Service namespace.", true, "production"),
			logField("service.instance.id", "string", "resource", "Service instance identity.", true, "instance-01"),
			logField("host.name", "string", "resource", "Host name.", true, "app-01"),
			logField("host.id", "string", "resource", "Host identity.", true, "i-001"),
			logField("k8s.namespace.name", "string", "resource", "Kubernetes namespace.", true, "default"),
			logField("k8s.pod.name", "string", "resource", "Kubernetes pod name.", true, "api-0"),
			logField("container.name", "string", "resource", "Container name.", true, "api"),
		}),
		logFieldCategory("log", "Log record", "Core log record fields.", []model.LogField{
			logField("timestamp", "time", "log", "Event timestamp.", true, "2026-05-08T12:00:00Z"),
			logField("severity_text", "string", "log", "Severity label.", true, "ERROR"),
			logField("severity_number", "number", "log", "Numeric severity.", true, "17"),
			logField("body", "string", "log", "Log message body.", false, "request failed"),
			logField("trace_id", "string", "log", "Trace correlation id.", true, "4bf92f3577b34da6a3ce929d0e0e4736"),
			logField("span_id", "string", "log", "Span correlation id.", true, "00f067aa0ba902b7"),
			logField("event.name", "string", "log", "Named event.", true, "exception"),
		}),
		logFieldCategory("http", "HTTP", "HTTP semantic fields.", []model.LogField{
			logField("http.request.method", "string", "http", "HTTP method.", true, "GET"),
			logField("url.path", "string", "http", "Request path.", true, "/api/v1/orders"),
			logField("http.response.status_code", "number", "http", "HTTP response status.", true, "500"),
			logField("client.address", "string", "http", "Client address without secrets.", true, "10.0.0.1"),
		}),
		logFieldCategory("exception", "Exception", "Exception fields.", []model.LogField{
			logField("exception.type", "string", "exception", "Exception type.", true, "RuntimeError"),
			logField("exception.message", "string", "exception", "Exception message.", false, "operation failed"),
			logField("exception.stacktrace", "string", "exception", "Exception stack trace.", false, "stack trace"),
		}),
	}
	fields := []model.LogField{}
	for _, category := range categories {
		fields = append(fields, category.Fields...)
	}
	return model.LogFieldsResponse{
		Status:     "partial",
		Categories: categories,
		Fields:     fields,
		LiveDiscovery: model.LogLiveDiscoveryState{
			Status:  "blocked",
			Blocker: model.LogsContractBlocked + ": live field discovery requires a connected log datasource contract",
		},
		Capabilities: map[string]model.LogCapabilityState{
			"builtin_fields":     logsCapabilityOK(),
			"live_discovery":     logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-FIELD-DISCOVERY", "logs.fields.discovery", "logs.query_service"),
			"field_values":       logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-FIELD-VALUES", "logs.fields.values", "logs.query_service"),
			"index_mutation":     logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-INDEX-MUTATION", "logs.index.create", "logs.index.delete", "logs.index.sync"),
			"query_service":      logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-QUERY", "logs.query_service", "logs.clickhouse.datasource"),
			"pipeline_deploy":    logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-PIPELINE-DEPLOY", "logs.pipeline.deploy", "logs.pipeline.executor"),
			"pipeline_rollback":  logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-PIPELINE-ROLLBACK", "logs.pipeline.rollback", "logs.pipeline.history"),
			"trace_linkage":      logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-TRACE-LINK", "logs.trace_link", "apm.trace_detail"),
			"agent_data_arrival": logsCapabilityBlocked("FX-CONTRACT-SIGNOZ-LOGS-AGENT-LINKAGE", "logs.agent_linkage", "findx_agent.data_arrival.logs"),
		},
	}
}

func logFieldCategory(key, name, desc string, fields []model.LogField) model.LogFieldCategory {
	return model.LogFieldCategory{Key: key, Name: name, Description: desc, Fields: fields}
}

func logField(key, typ, category, desc string, indexed bool, examples ...string) model.LogField {
	return model.LogField{Key: key, Type: typ, Category: category, Description: desc, Indexed: indexed, Examples: examples}
}

func defaultRaw(raw json.RawMessage, fallback string) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(fallback)
	}
	return raw
}
