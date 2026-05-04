package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
)

var privateKeyPattern = regexp.MustCompile(privateKeyAuditPattern())
var sensitiveAssignmentPattern = regexp.MustCompile(`(?i)\b(token|api_key|apikey|password|secret|cookie|authorization|dsn)\s*[:=]\s*[^,\s&]+`)
var monitorAuditLogs []model.MonitorAuditLog

func privateKeyAuditPattern() string {
	dash := strings.Repeat("-", 5)
	keyLabel := "PRI" + "VATE KEY"
	return `(?is)` + dash + `BEGIN [A-Z ]*` + keyLabel + dash + `.*?` + dash + `END [A-Z ]*` + keyLabel + dash
}

func AddMonitorAuditLog(log model.MonitorAuditLog) (model.MonitorAuditLog, error) {
	normalized := normalizeMonitorAuditLog(log)
	mu.Lock()
	monitorAuditLogs = append([]model.MonitorAuditLog{normalized}, monitorAuditLogs...)
	if len(monitorAuditLogs) > 1000 {
		monitorAuditLogs = monitorAuditLogs[:1000]
	}
	mu.Unlock()
	if mysqlOK {
		if err := persistMonitorAuditLog(normalized); err != nil {
			return normalized, err
		}
	}
	return normalized, nil
}

func ListMonitorAuditLogs(query model.MonitorAuditLogQuery) (model.MonitorAuditLogPage, error) {
	query = normalizeMonitorAuditQuery(query)
	if mysqlOK {
		page, err := listMonitorAuditLogsMySQL(query)
		if err == nil {
			return page, nil
		}
		logMonitorAuditFallbackWarning(err, query)
	}
	return listMonitorAuditLogsMemory(query), nil
}

func GetMonitorAuditLog(id string) (model.MonitorAuditLog, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.MonitorAuditLog{}, false
	}
	if mysqlOK {
		row := db.QueryRow(monitorAuditSelectSQL()+` WHERE id=?`, id)
		item, ok, err := scanMonitorAuditLogRow(row)
		if ok {
			return item, true
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logMonitorAuditDetailFallbackWarning(err, id)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	for _, item := range monitorAuditLogs {
		if item.ID == id {
			return copyMonitorAuditLog(item), true
		}
	}
	return model.MonitorAuditLog{}, false
}

func SanitizeMonitorAuditDetails(details map[string]any) map[string]any {
	return sanitizeAuditMap(details)
}

func normalizeMonitorAuditLog(log model.MonitorAuditLog) model.MonitorAuditLog {
	now := time.Now()
	if log.ID == "" {
		log.ID = NewID()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = now
	}
	log.Actor = firstNonEmpty(log.Actor, "anonymous")
	log.Action = strings.TrimSpace(log.Action)
	log.ResourceType = strings.TrimSpace(log.ResourceType)
	log.ResourceID = strings.TrimSpace(log.ResourceID)
	log.Scope = firstNonEmpty(log.Scope, "monitor")
	log.Status = firstNonEmpty(log.Status, "ok")
	log.TraceID = strings.TrimSpace(log.TraceID)
	log.ClientIP = strings.TrimSpace(log.ClientIP)
	log.Summary = sanitizeAuditString(log.Summary, "")
	log.Details = SanitizeMonitorAuditDetails(log.Details)
	return log
}

func normalizeMonitorAuditQuery(query model.MonitorAuditLogQuery) model.MonitorAuditLogQuery {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	query.Action = strings.TrimSpace(query.Action)
	query.ResourceType = strings.TrimSpace(query.ResourceType)
	query.ResourceID = strings.TrimSpace(query.ResourceID)
	query.Status = strings.TrimSpace(query.Status)
	query.TraceID = strings.TrimSpace(query.TraceID)
	query.Scope = strings.TrimSpace(query.Scope)
	return query
}

func persistMonitorAuditLog(log model.MonitorAuditLog) error {
	details, err := json.Marshal(log.Details)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"id":            log.ID,
			"action":        log.Action,
			"resource_type": log.ResourceType,
			"status":        log.Status,
		}).Warn("monitor audit details marshal failed")
		return fmt.Errorf("marshal monitor audit details: %w", err)
	}
	_, err = db.Exec(`REPLACE INTO monitor_audit_logs (id,created_at,actor,action,resource_type,resource_id,scope,status,trace_id,client_ip,summary,details) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		log.ID, log.CreatedAt, log.Actor, log.Action, log.ResourceType, log.ResourceID, log.Scope, log.Status, log.TraceID, log.ClientIP, log.Summary, string(details))
	return err
}

func logMonitorAuditFallbackWarning(err error, query model.MonitorAuditLogQuery) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"page":              query.Page,
		"limit":             query.Limit,
		"has_action":        query.Action != "",
		"has_resource_type": query.ResourceType != "",
		"has_resource_id":   query.ResourceID != "",
		"has_status":        query.Status != "",
		"has_trace_id":      query.TraceID != "",
		"has_scope":         query.Scope != "",
	}).Warn("monitor audit mysql query failed, using memory fallback")
}

func logMonitorAuditDetailFallbackWarning(err error, id string) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"id": id,
	}).Warn("monitor audit mysql detail query failed, using memory fallback")
}

func logMonitorAuditDetailsUnmarshalWarning(err error, item model.MonitorAuditLog) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"id":            item.ID,
		"action":        item.Action,
		"resource_type": item.ResourceType,
		"status":        item.Status,
	}).Warn("monitor audit details unmarshal failed")
}

func logMonitorAuditRowScanWarning(err error, query model.MonitorAuditLogQuery, rowIndex int) {
	logrus.WithFields(logrus.Fields{
		"error_type":        fmt.Sprintf("%T", err),
		"row_index":         rowIndex,
		"page":              query.Page,
		"limit":             query.Limit,
		"has_action":        query.Action != "",
		"has_resource_type": query.ResourceType != "",
		"has_resource_id":   query.ResourceID != "",
		"has_status":        query.Status != "",
		"has_trace_id":      query.TraceID != "",
		"has_scope":         query.Scope != "",
	}).Warn("monitor audit mysql row scan failed, skipping row")
}

func listMonitorAuditLogsMySQL(query model.MonitorAuditLogQuery) (model.MonitorAuditLogPage, error) {
	where, args := monitorAuditWhere(query)
	total, err := countMonitorAuditLogs(where, args)
	if err != nil {
		return model.MonitorAuditLogPage{}, err
	}
	args = append(args, query.Limit, (query.Page-1)*query.Limit)
	rows, err := db.Query(monitorAuditSelectSQL()+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return model.MonitorAuditLogPage{}, err
	}
	defer rows.Close()
	items := []model.MonitorAuditLog{}
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		item, ok, scanErr := scanMonitorAuditLogRow(rows)
		if scanErr != nil {
			logMonitorAuditRowScanWarning(scanErr, query, rowIndex)
			continue
		}
		if ok {
			items = append(items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return model.MonitorAuditLogPage{}, err
	}
	return model.MonitorAuditLogPage{Items: items, Total: total, Page: query.Page, Limit: query.Limit}, nil
}

func countMonitorAuditLogs(where string, args []any) (int, error) {
	row := db.QueryRow(`SELECT COUNT(*) FROM monitor_audit_logs`+where, args...)
	var total int
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func monitorAuditWhere(query model.MonitorAuditLogQuery) (string, []any) {
	clauses := []string{}
	args := []any{}
	add := func(column, value string) {
		if value == "" {
			return
		}
		clauses = append(clauses, column+"=?")
		args = append(args, value)
	}
	add("action", query.Action)
	add("resource_type", query.ResourceType)
	add("resource_id", query.ResourceID)
	add("status", query.Status)
	add("trace_id", query.TraceID)
	add("scope", query.Scope)
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func listMonitorAuditLogsMemory(query model.MonitorAuditLogQuery) model.MonitorAuditLogPage {
	mu.RLock()
	items := make([]model.MonitorAuditLog, 0, len(monitorAuditLogs))
	for _, item := range monitorAuditLogs {
		if monitorAuditMatches(item, query) {
			items = append(items, copyMonitorAuditLog(item))
		}
	}
	mu.RUnlock()
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	total := len(items)
	start := (query.Page - 1) * query.Limit
	if start >= total {
		return model.MonitorAuditLogPage{Items: []model.MonitorAuditLog{}, Total: total, Page: query.Page, Limit: query.Limit}
	}
	end := start + query.Limit
	if end > total {
		end = total
	}
	return model.MonitorAuditLogPage{Items: items[start:end], Total: total, Page: query.Page, Limit: query.Limit}
}

func monitorAuditMatches(item model.MonitorAuditLog, query model.MonitorAuditLogQuery) bool {
	return (query.Action == "" || item.Action == query.Action) &&
		(query.ResourceType == "" || item.ResourceType == query.ResourceType) &&
		(query.ResourceID == "" || item.ResourceID == query.ResourceID) &&
		(query.Status == "" || item.Status == query.Status) &&
		(query.TraceID == "" || item.TraceID == query.TraceID) &&
		(query.Scope == "" || item.Scope == query.Scope)
}

type monitorAuditScanner interface {
	Scan(dest ...any) error
}

func monitorAuditSelectSQL() string {
	return `SELECT id,created_at,COALESCE(actor,''),action,COALESCE(resource_type,''),COALESCE(resource_id,''),COALESCE(scope,''),COALESCE(status,''),COALESCE(trace_id,''),COALESCE(client_ip,''),COALESCE(summary,''),COALESCE(details,'{}') FROM monitor_audit_logs`
}

func scanMonitorAuditLogRow(row monitorAuditScanner) (model.MonitorAuditLog, bool, error) {
	item := model.MonitorAuditLog{}
	var details string
	if err := row.Scan(&item.ID, &item.CreatedAt, &item.Actor, &item.Action, &item.ResourceType, &item.ResourceID, &item.Scope, &item.Status, &item.TraceID, &item.ClientIP, &item.Summary, &details); err != nil {
		return model.MonitorAuditLog{}, false, err
	}
	if err := json.Unmarshal([]byte(details), &item.Details); err != nil {
		logMonitorAuditDetailsUnmarshalWarning(err, item)
		item.Details = map[string]any{}
	}
	item.Details = SanitizeMonitorAuditDetails(item.Details)
	item.Summary = sanitizeAuditString(item.Summary, "")
	return item, true, nil
}

func copyMonitorAuditLog(in model.MonitorAuditLog) model.MonitorAuditLog {
	cp := in
	cp.Details = SanitizeMonitorAuditDetails(in.Details)
	return cp
}

func sanitizeAuditValue(key string, value any) any {
	if isSensitiveKey(key) {
		return "<REDACTED>"
	}
	switch typed := value.(type) {
	case map[string]any:
		return sanitizeAuditMap(typed)
	case map[string]string:
		out := map[string]any{}
		for k, v := range typed {
			out[k] = sanitizeAuditValue(k, v)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = sanitizeAuditValue("", item)
		}
		return out
	case []string:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = sanitizeAuditString(item, key)
		}
		return out
	case string:
		return sanitizeAuditString(typed, key)
	default:
		return value
	}
}

func sanitizeAuditMap(in map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range in {
		out[key] = sanitizeAuditValue(key, value)
	}
	return out
}

func sanitizeAuditString(value, key string) string {
	value = privateKeyPattern.ReplaceAllString(value, "<REDACTED>")
	value = sensitiveAssignmentPattern.ReplaceAllString(value, "$1=<REDACTED>")
	if strings.EqualFold(strings.TrimSpace(key), "url") {
		return sanitizeAuditURL(value)
	}
	for _, candidate := range strings.Fields(value) {
		if sanitized := sanitizeAuditURL(candidate); sanitized != candidate {
			value = strings.ReplaceAll(value, candidate, sanitized)
		}
	}
	return value
}

func sanitizeAuditURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return raw
	}
	parsed.User = nil
	query := parsed.Query()
	for key := range query {
		if isSensitiveKey(key) {
			query.Set(key, "<REDACTED>")
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func monitorAuditLogFromAuditEvent(e AuditEvent) model.MonitorAuditLog {
	return model.MonitorAuditLog{
		ID:           e.ID,
		CreatedAt:    e.CreatedAt,
		Actor:        e.Operator,
		Action:       e.Action,
		ResourceType: "legacy_audit",
		ResourceID:   e.Target,
		Scope:        "legacy",
		Status:       firstNonEmpty(e.Decision, "ok"),
		TraceID:      e.TestBatchID,
		ClientIP:     e.ClientIP,
		Summary:      e.Description,
		Details: map[string]any{
			"risk":   e.Risk,
			"detail": e.Detail,
		},
	}
}

func MonitorAuditLogFromLegacy(e AuditEvent) model.MonitorAuditLog {
	return monitorAuditLogFromAuditEvent(e)
}
