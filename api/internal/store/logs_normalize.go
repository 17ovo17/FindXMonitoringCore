package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func normalizeLogPipeline(input *model.LogPipeline, actor string, now time.Time) (*model.LogPipeline, error) {
	item := copyLogPipeline(input)
	if item == nil {
		return nil, ErrLogsValidation
	}
	item.ID = strings.TrimSpace(item.ID)
	if item.ID == "" {
		item.ID = NewID()
	}
	item.Name = strings.TrimSpace(item.Name)
	item.Version = normalizeLogVersion(item.Version)
	if item.Name == "" || !validLogJSON(item.Stages, false) || !validLogJSON(item.Config, true) {
		return nil, ErrLogsValidation
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
		item.CreatedBy = actor
	}
	item.UpdatedAt = now
	item.UpdatedBy = actor
	return item, nil
}

func normalizeExplorerSavedView(input *model.ExplorerSavedView, actor string, now time.Time, keepCreated bool) (*model.ExplorerSavedView, error) {
	item := copyExplorerSavedView(input)
	if item == nil {
		return nil, ErrLogsValidation
	}
	item.ID = strings.TrimSpace(item.ID)
	if item.ID == "" {
		item.ID = NewID()
	}
	item.SourcePage = strings.TrimSpace(item.SourcePage)
	item.Name = strings.TrimSpace(item.Name)
	if item.SourcePage != model.LogsSourcePage || item.Name == "" || !validSavedViewJSON(item) {
		return nil, ErrLogsValidation
	}
	if keepCreated {
		existing, ok, _ := GetExplorerSavedView(item.ID)
		if ok {
			item.CreatedAt = existing.CreatedAt
			item.CreatedBy = existing.CreatedBy
		}
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
		item.CreatedBy = actor
	}
	item.UpdatedAt = now
	item.UpdatedBy = actor
	return item, nil
}

func normalizeLogQueryRequest(req model.LogQueryRequest) model.LogQueryRequest {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	req.Source = strings.TrimSpace(req.Source)
	req.Query = strings.ToLower(strings.TrimSpace(req.Query))
	req.Status = strings.TrimSpace(req.Status)
	req.Action = strings.TrimSpace(req.Action)
	req.ResourceType = strings.TrimSpace(req.ResourceType)
	req.ResourceID = strings.TrimSpace(req.ResourceID)
	req.TraceID = strings.TrimSpace(req.TraceID)
	req.Scope = strings.TrimSpace(req.Scope)
	return req
}

func normalizeLogAggregateRequest(req model.LogAggregateRequest) model.LogAggregateRequest {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 500 {
		req.Limit = 500
	}
	req.Source = strings.TrimSpace(req.Source)
	req.GroupBy = strings.TrimSpace(req.GroupBy)
	if req.GroupBy == "" {
		req.GroupBy = "status"
	}
	req.Status = strings.TrimSpace(req.Status)
	req.Action = strings.TrimSpace(req.Action)
	req.ResourceType = strings.TrimSpace(req.ResourceType)
	req.Scope = strings.TrimSpace(req.Scope)
	return req
}

func normalizeLogContextRequest(req model.LogContextRequest) model.LogContextRequest {
	req.Source = strings.TrimSpace(req.Source)
	req.LogID = strings.TrimSpace(req.LogID)
	req.TraceID = strings.TrimSpace(req.TraceID)
	req.Scope = strings.TrimSpace(req.Scope)
	if req.Before <= 0 {
		req.Before = 5
	}
	if req.After <= 0 {
		req.After = 5
	}
	if req.Before > 50 {
		req.Before = 50
	}
	if req.After > 50 {
		req.After = 50
	}
	return req
}

func auditLogRecord(item model.MonitorAuditLog) model.LogRecord {
	body := strings.TrimSpace(item.Summary)
	if body == "" {
		body = strings.TrimSpace(item.Action)
	}
	return model.LogRecord{
		ID: item.ID, Timestamp: item.CreatedAt, Source: model.LogsSourceFindXAudit, SourceName: "FindX 审计日志",
		SeverityText: auditSeverity(item.Status), SeverityNumber: auditSeverityNumber(item.Status),
		ServiceName: "findx-audit", Body: body, TraceID: item.TraceID,
		Attributes: map[string]any{
			"actor": item.Actor, "action": item.Action, "resource_type": item.ResourceType,
			"resource_id": item.ResourceID, "scope": item.Scope, "status": item.Status,
			"client_ip": item.ClientIP, "details": SanitizeMonitorAuditDetails(item.Details),
		},
	}
}

func auditLogMatchesText(item model.MonitorAuditLog, query string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		item.ID, item.Actor, item.Action, item.ResourceType, item.ResourceID,
		item.Scope, item.Status, item.TraceID, item.Summary,
	}, " "))
	return strings.Contains(haystack, query)
}

func auditAggregateKey(item model.MonitorAuditLog, groupBy string) string {
	switch strings.ToLower(strings.TrimSpace(groupBy)) {
	case "action":
		return item.Action
	case "resource_type", "resource.type":
		return item.ResourceType
	case "actor":
		return item.Actor
	case "scope":
		return item.Scope
	case "severity_text", "status":
		return item.Status
	default:
		return item.Status
	}
}

func auditSeverity(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "error", "failed", "fail", "denied", "blocked":
		return "ERROR"
	case "warn", "warning", "risk":
		return "WARN"
	default:
		return "INFO"
	}
}

func auditSeverityNumber(status string) int {
	switch auditSeverity(status) {
	case "ERROR":
		return 17
	case "WARN":
		return 13
	default:
		return 9
	}
}

func validSavedViewJSON(item *model.ExplorerSavedView) bool {
	return validLogJSON(item.Query, true) && validLogJSON(item.Filters, true) &&
		validLogJSON(item.Columns, false) && validLogJSON(item.TimeRange, true) &&
		validLogJSON(item.Layout, true)
}

func validLogJSON(raw json.RawMessage, wantObject bool) bool {
	if len(raw) == 0 {
		return true
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	if wantObject {
		_, ok := value.(map[string]any)
		return ok
	}
	_, ok := value.([]any)
	return ok
}

func normalizeLogVersion(version string) string {
	if trimmed := strings.TrimSpace(version); trimmed != "" {
		return trimmed
	}
	return "draft"
}

func sortLogPipelines(items []model.LogPipeline) {
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
}

func sortExplorerViews(items []model.ExplorerSavedView) {
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
}

func copyLogPipeline(in *model.LogPipeline) *model.LogPipeline {
	if in == nil {
		return nil
	}
	cp := *in
	cp.Stages = append([]byte{}, in.Stages...)
	cp.Config = append([]byte{}, in.Config...)
	return &cp
}

func copyExplorerSavedView(in *model.ExplorerSavedView) *model.ExplorerSavedView {
	if in == nil {
		return nil
	}
	cp := *in
	cp.Query = append([]byte{}, in.Query...)
	cp.Filters = append([]byte{}, in.Filters...)
	cp.Columns = append([]byte{}, in.Columns...)
	cp.TimeRange = append([]byte{}, in.TimeRange...)
	cp.Layout = append([]byte{}, in.Layout...)
	return &cp
}

func (e ErrLogValidation) Error() string { return string(e) }

type ErrLogValidation string

func logsValidationError(reason string) error {
	if reason == "" {
		return ErrLogsValidation
	}
	return fmt.Errorf("%w: %s", ErrLogsValidation, reason)
}
