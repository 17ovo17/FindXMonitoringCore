package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"ai-workbench-api/internal/model"
)

var (
	logsMu             sync.RWMutex
	logPipelines       = map[string]*model.LogPipeline{}
	explorerSavedViews = map[string]*model.ExplorerSavedView{}
)

var ErrLogsValidation = errors.New("logs contract validation failed")

func ListLogPipelines(version string) ([]model.LogPipeline, error) {
	version = normalizeLogVersion(version)
	if mysqlOK {
		rows, err := db.Query(`SELECT id,name,version_id,description,enabled,COALESCE(stages,'[]'),COALESCE(config,'{}'),created_by,updated_by,created_at,updated_at FROM log_pipelines WHERE version_id=? ORDER BY updated_at DESC`, version)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		items, err := scanLogPipelines(rows)
		if err != nil || len(items) > 0 {
			return items, err
		}
	}
	logsMu.RLock()
	out := make([]model.LogPipeline, 0, len(logPipelines))
	for _, item := range logPipelines {
		if item.Version == version {
			out = append(out, *copyLogPipeline(item))
		}
	}
	logsMu.RUnlock()
	sortLogPipelines(out)
	if len(out) == 0 {
		out = []model.LogPipeline{DefaultLogPipeline(version)}
	}
	return out, nil
}

func SaveLogPipeline(input *model.LogPipeline, actor string) (*model.LogPipeline, error) {
	item, err := normalizeLogPipeline(input, actor, time.Now())
	if err != nil {
		return nil, err
	}
	logsMu.Lock()
	logPipelines[item.ID] = copyLogPipeline(item)
	logsMu.Unlock()
	if mysqlOK {
		if err := persistLogPipeline(item); err != nil {
			return copyLogPipeline(item), err
		}
	}
	return copyLogPipeline(item), nil
}

func DefaultLogPipeline(version string) model.LogPipeline {
	now := time.Now()
	return model.LogPipeline{
		ID:          "default-" + normalizeLogVersion(version),
		Name:        "默认日志解析模板",
		Version:     normalizeLogVersion(version),
		Description: "仅提供内置解析草稿；日志数据源和生产管道生效契约尚未接入。",
		Enabled:     false,
		Stages:      json.RawMessage(`[{"id":"parse-json","type":"json","on_error":"keep_raw"},{"id":"map-severity","type":"attribute_map","from":"severity_text","to":"severity"}]`),
		Config:      json.RawMessage(`{"status":"template","data_source":"BLOCKED_BY_CONTRACT"}`),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func ListExplorerSavedViews(sourcePage string) ([]model.ExplorerSavedView, error) {
	sourcePage = strings.TrimSpace(sourcePage)
	if mysqlOK {
		query := `SELECT id,source_page,name,description,COALESCE(query_json,'{}'),COALESCE(filters,'{}'),COALESCE(columns_json,'[]'),COALESCE(time_range,'{}'),COALESCE(layout,'{}'),created_by,updated_by,created_at,updated_at FROM explorer_saved_views`
		args := []any{}
		if sourcePage != "" {
			query += ` WHERE source_page=?`
			args = append(args, sourcePage)
		}
		rows, err := db.Query(query+` ORDER BY updated_at DESC`, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanExplorerSavedViews(rows)
	}
	logsMu.RLock()
	defer logsMu.RUnlock()
	out := make([]model.ExplorerSavedView, 0, len(explorerSavedViews))
	for _, item := range explorerSavedViews {
		if sourcePage == "" || item.SourcePage == sourcePage {
			out = append(out, *copyExplorerSavedView(item))
		}
	}
	sortExplorerViews(out)
	return out, nil
}

func GetExplorerSavedView(id string) (*model.ExplorerSavedView, bool, error) {
	if mysqlOK {
		row := db.QueryRow(`SELECT id,source_page,name,description,COALESCE(query_json,'{}'),COALESCE(filters,'{}'),COALESCE(columns_json,'[]'),COALESCE(time_range,'{}'),COALESCE(layout,'{}'),created_by,updated_by,created_at,updated_at FROM explorer_saved_views WHERE id=?`, id)
		item, err := scanExplorerSavedViewRow(row)
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return item, err == nil, err
	}
	logsMu.RLock()
	defer logsMu.RUnlock()
	item, ok := explorerSavedViews[id]
	return copyExplorerSavedView(item), ok, nil
}

func SaveExplorerSavedView(input *model.ExplorerSavedView, actor string) (*model.ExplorerSavedView, error) {
	return saveExplorerSavedView(input, actor, false)
}

func UpdateExplorerSavedView(input *model.ExplorerSavedView, actor string) (*model.ExplorerSavedView, bool, error) {
	_, ok, err := GetExplorerSavedView(input.ID)
	if err != nil || !ok {
		return nil, ok, err
	}
	item, err := saveExplorerSavedView(input, actor, true)
	return item, true, err
}

func DeleteExplorerSavedView(id string) (bool, error) {
	found := false
	logsMu.Lock()
	if _, ok := explorerSavedViews[id]; ok {
		delete(explorerSavedViews, id)
		found = true
	}
	logsMu.Unlock()
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM explorer_saved_views WHERE id=?`, id)
		if err != nil {
			return found, err
		}
		rows, _ := res.RowsAffected()
		found = found || rows > 0
	}
	return found, nil
}

func QueryFindXAuditLogs(req model.LogQueryRequest) (model.LogQueryResponse, error) {
	req = normalizeLogQueryRequest(req)
	page, err := ListMonitorAuditLogs(model.MonitorAuditLogQuery{
		Page: req.Page, Limit: req.Limit, Action: req.Action, ResourceType: req.ResourceType,
		ResourceID: req.ResourceID, Status: req.Status, TraceID: req.TraceID, Scope: req.Scope,
	})
	if err != nil {
		return model.LogQueryResponse{}, err
	}
	items := make([]model.LogRecord, 0, len(page.Items))
	for _, item := range page.Items {
		if req.Query != "" && !auditLogMatchesText(item, req.Query) {
			continue
		}
		items = append(items, auditLogRecord(item))
	}
	return model.LogQueryResponse{
		Source: model.LogsSourceFindXAudit, SourceName: "FindX 审计日志", Status: "ok",
		Items: items, Total: page.Total, Page: page.Page, Limit: page.Limit,
		Meta: map[string]any{
			"contract":    "findx_audit",
			"description": "来自 monitor_audit_logs 的真实 FindX 审计事件；不代表通用 OTel 日志已接入。",
		},
	}, nil
}

func AggregateFindXAuditLogs(req model.LogAggregateRequest) (model.LogAggregateResponse, error) {
	req = normalizeLogAggregateRequest(req)
	page, err := ListMonitorAuditLogs(model.MonitorAuditLogQuery{
		Page: req.Page, Limit: req.Limit, Action: req.Action, ResourceType: req.ResourceType,
		Status: req.Status, Scope: req.Scope,
	})
	if err != nil {
		return model.LogAggregateResponse{}, err
	}
	counts := map[string]int{}
	for _, item := range page.Items {
		key := auditAggregateKey(item, req.GroupBy)
		if key == "" {
			key = "unknown"
		}
		counts[key]++
	}
	buckets := make([]model.LogAggregateBucket, 0, len(counts))
	for key, count := range counts {
		buckets = append(buckets, model.LogAggregateBucket{Key: key, Label: key, Count: count})
	}
	sort.Slice(buckets, func(i, j int) bool {
		if buckets[i].Count == buckets[j].Count {
			return buckets[i].Key < buckets[j].Key
		}
		return buckets[i].Count > buckets[j].Count
	})
	return model.LogAggregateResponse{
		Source: model.LogsSourceFindXAudit, SourceName: "FindX 审计日志", Status: "ok",
		GroupBy: req.GroupBy, Buckets: buckets, Total: page.Total,
		Meta: map[string]any{
			"contract":        "findx_audit",
			"blocked_generic": model.LogsContractBlocked + ": 通用 OTel 日志聚合数据源契约未接入。",
		},
	}, nil
}

func ContextFindXAuditLogs(req model.LogContextRequest) (model.LogContextResponse, error) {
	req = normalizeLogContextRequest(req)
	query := model.MonitorAuditLogQuery{Page: 1, Limit: req.Before + req.After + 1, TraceID: req.TraceID, Scope: req.Scope}
	var center *model.LogRecord
	centerID := ""
	var centerTime time.Time
	if req.LogID != "" {
		item, ok := GetMonitorAuditLog(req.LogID)
		if ok {
			record := auditLogRecord(item)
			center = &record
			centerID = item.ID
			centerTime = item.CreatedAt
			query.TraceID = firstNonEmpty(req.TraceID, item.TraceID)
			query.Scope = firstNonEmpty(req.Scope, item.Scope)
		}
	}
	if query.TraceID == "" && query.Scope == "" {
		return model.LogContextResponse{
			Source: model.LogsSourceFindXAudit, SourceName: "FindX 审计日志", Status: "ok",
			Before: []model.LogRecord{}, After: []model.LogRecord{}, Items: []model.LogRecord{},
			Meta: map[string]any{"contract": "findx_audit", "empty_reason": "log_id_or_trace_id_required"},
		}, nil
	}
	page, before, after, items, err := findXAuditContextWindow(query, req, centerID, centerTime)
	if err != nil {
		return model.LogContextResponse{}, err
	}
	if center != nil {
		window := make([]model.LogRecord, 0, len(before)+1+len(after))
		window = append(window, before...)
		window = append(window, *center)
		window = append(window, after...)
		items = window
	}
	return model.LogContextResponse{
		Source: model.LogsSourceFindXAudit, SourceName: "FindX 审计日志", Status: "ok",
		Center: center, Before: before, After: after, Items: items, Total: page.Total,
		Meta: map[string]any{
			"contract":    "findx_audit",
			"description": "来自 monitor_audit_logs 的真实 FindX 审计上下文；不代表通用 OTel 日志上下文已接入。",
		},
	}, nil
}

func findXAuditContextWindow(query model.MonitorAuditLogQuery, req model.LogContextRequest, centerID string, centerTime time.Time) (model.MonitorAuditLogPage, []model.LogRecord, []model.LogRecord, []model.LogRecord, error) {
	if centerID == "" {
		page, err := ListMonitorAuditLogs(query)
		if err != nil {
			return model.MonitorAuditLogPage{}, nil, nil, nil, err
		}
		items := make([]model.LogRecord, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, auditLogRecord(item))
		}
		return page, []model.LogRecord{}, []model.LogRecord{}, items, nil
	}

	query.Page = 1
	query.Limit = 100
	firstPage := model.MonitorAuditLogPage{Page: 1, Limit: query.Limit}
	before := []model.LogRecord{}
	after := []model.LogRecord{}
	for {
		page, err := ListMonitorAuditLogs(query)
		if err != nil {
			return model.MonitorAuditLogPage{}, nil, nil, nil, err
		}
		if query.Page == 1 {
			firstPage = page
		}
		for _, item := range page.Items {
			if item.ID == centerID {
				continue
			}
			record := auditLogRecord(item)
			if item.CreatedAt.After(centerTime) && len(before) < req.Before {
				before = append(before, record)
				continue
			}
			if !item.CreatedAt.After(centerTime) && len(after) < req.After {
				after = append(after, record)
			}
		}
		if len(after) >= req.After || len(page.Items) == 0 || query.Page*query.Limit >= page.Total {
			break
		}
		query.Page++
	}
	return firstPage, before, after, nil, nil
}

func saveExplorerSavedView(input *model.ExplorerSavedView, actor string, requireExisting bool) (*model.ExplorerSavedView, error) {
	item, err := normalizeExplorerSavedView(input, actor, time.Now(), requireExisting)
	if err != nil {
		return nil, err
	}
	logsMu.Lock()
	explorerSavedViews[item.ID] = copyExplorerSavedView(item)
	logsMu.Unlock()
	if mysqlOK {
		if err := persistExplorerSavedView(item); err != nil {
			return copyExplorerSavedView(item), err
		}
	}
	return copyExplorerSavedView(item), nil
}
