package store

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

var ErrTerminalMonitorAlertEvent = errors.New("terminal alert event cannot be changed")
var ErrInvalidMonitorAlertEventAction = errors.New("invalid alert event action")

func ListMonitorAlertEvents(current bool) []model.MonitorAlertEvent {
	if mysqlOK {
		table := "monitor_alert_events_history"
		if current {
			table = "monitor_alert_events_current"
		}
		rows, err := db.Query(monitorEventSelectSQL(table) + ` ORDER BY last_seen DESC LIMIT 1000`)
		if err == nil {
			defer rows.Close()
			return scanMonitorAlertEvents(rows)
		}
	}
	return listMonitorEventsMemory(current)
}

// ListMonitorAlertEventsPaged 分页查询告警事件，支持状态过滤和搜索。
func ListMonitorAlertEventsPaged(current bool, page, pageSize int, status, severity, search string) ([]model.MonitorAlertEvent, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if mysqlOK {
		return listMonitorAlertEventsPagedMySQL(current, offset, pageSize, status, severity, search)
	}
	return listMonitorAlertEventsPagedMemory(current, offset, pageSize, status, severity, search)
}

func listMonitorAlertEventsPagedMySQL(current bool, offset, limit int, status, severity, search string) ([]model.MonitorAlertEvent, int64, error) {
	table := "monitor_alert_events_history"
	if current {
		table = "monitor_alert_events_current"
	}

	where, args := buildMonitorEventWhereClause(status, severity, search)

	countSQL := `SELECT COUNT(*) FROM ` + table + where
	var total int64
	if err := db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count alert events: %w", err)
	}

	querySQL := monitorEventSelectSQL(table) + where + ` ORDER BY last_seen DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	rows, err := db.Query(querySQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query alert events: %w", err)
	}
	defer rows.Close()
	events := scanMonitorAlertEvents(rows)
	return events, total, nil
}

func buildMonitorEventWhereClause(status, severity, search string) (string, []any) {
	conditions := []string{}
	args := []any{}

	if status != "" {
		conditions = append(conditions, "status=?")
		args = append(args, status)
	}
	if severity != "" {
		conditions = append(conditions, "severity=?")
		args = append(args, severity)
	}
	if search != "" {
		conditions = append(conditions, "(name LIKE ? OR event_key LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func listMonitorAlertEventsPagedMemory(current bool, offset, limit int, status, severity, search string) ([]model.MonitorAlertEvent, int64, error) {
	all := listMonitorEventsMemory(current)
	filtered := filterMonitorEventsInMemory(all, status, severity, search)
	total := int64(len(filtered))

	if offset >= len(filtered) {
		return []model.MonitorAlertEvent{}, total, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], total, nil
}

func filterMonitorEventsInMemory(events []model.MonitorAlertEvent, status, severity, search string) []model.MonitorAlertEvent {
	if status == "" && severity == "" && search == "" {
		return events
	}
	out := []model.MonitorAlertEvent{}
	for _, event := range events {
		if status != "" && event.Status != status {
			continue
		}
		if severity != "" && event.Severity != severity {
			continue
		}
		if search != "" && !strings.Contains(event.Name, search) && !strings.Contains(event.EventKey, search) {
			continue
		}
		out = append(out, event)
	}
	return out
}

func GetMonitorAlertEvent(id string) (*model.MonitorAlertEvent, bool) {
	if mysqlOK {
		for _, table := range []string{"monitor_alert_events_current", "monitor_alert_events_history"} {
			row := db.QueryRow(monitorEventSelectSQL(table)+` WHERE id=?`, id)
			if event, ok := scanMonitorAlertEventRow(row); ok {
				event.ActionLog = ListMonitorAlertEventActions(id)
				return event, true
			}
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	if event, ok := monitorEventsCurrent[id]; ok {
		return copyMonitorAlertEvent(event), true
	}
	event, ok := monitorEventsHistory[id]
	return copyMonitorAlertEvent(event), ok
}

func UpsertMonitorAlertEvent(event *model.MonitorAlertEvent) (*model.MonitorAlertEvent, error) {
	normalized := normalizeMonitorAlertEvent(event)
	countDelta := max(1, normalized.Count)
	var existing *model.MonitorAlertEvent
	if mysqlOK {
		if found, ok := getCurrentMonitorAlertEventByFingerprint(normalized.Fingerprint); ok {
			existing = found
		}
	}
	mu.Lock()
	if existing == nil {
		existing = findCurrentMonitorAlertEventLocked(normalized)
	}
	if existing != nil {
		merged := mergeMonitorAlertEvent(existing, normalized)
		if existing.ID != merged.ID {
			delete(monitorEventsCurrent, existing.ID)
		}
		monitorEventsCurrent[merged.ID] = copyMonitorAlertEvent(merged)
		normalized = merged
	} else {
		monitorEventsCurrent[normalized.ID] = copyMonitorAlertEvent(normalized)
	}
	mu.Unlock()
	if mysqlOK {
		if err := persistMonitorAlertEventCurrent(normalized, countDelta, true); err != nil {
			return normalized, err
		}
	}
	return normalized, nil
}

func ApplyMonitorAlertEventAction(id string, action model.MonitorAlertAction) (*model.MonitorAlertEvent, bool, error) {
	event, ok := GetMonitorAlertEvent(id)
	if !ok {
		return nil, false, nil
	}
	if _, err := model.ValidateMonitorAlertEventTransition(event.Status, action.Action); err != nil {
		if model.IsTerminalMonitorAlertEventStatus(event.Status) {
			return nil, true, ErrTerminalMonitorAlertEvent
		}
		return nil, true, fmt.Errorf("%w: %v", ErrInvalidMonitorAlertEventAction, err)
	}
	if model.IsTerminalMonitorAlertEventStatus(event.Status) {
		return nil, true, ErrTerminalMonitorAlertEvent
	}
	now := time.Now()
	action = normalizeMonitorEventAction(event, action, now)
	action = applyMonitorEventMutation(event, action, now)
	mu.Lock()
	delete(monitorEventsCurrent, id)
	if model.IsTerminalMonitorAlertEventStatus(event.Status) {
		monitorEventsHistory[id] = copyMonitorAlertEvent(event)
	} else {
		monitorEventsCurrent[id] = copyMonitorAlertEvent(event)
	}
	monitorEventActions[id] = append(monitorEventActions[id], action)
	mu.Unlock()
	if mysqlOK {
		if err := persistMonitorAlertAction(action); err != nil {
			return event, true, err
		}
		current := !model.IsTerminalMonitorAlertEventStatus(event.Status)
		if err := persistMonitorAlertEvent(event, current); err != nil {
			return event, true, err
		}
		if model.IsTerminalMonitorAlertEventStatus(event.Status) {
			res, err := db.Exec(`DELETE FROM monitor_alert_events_current WHERE id=?`, id)
			if err != nil {
				return event, true, err
			}
			if res == nil {
				return event, true, fmt.Errorf("delete current alert event returned no result")
			}
		}
	}
	event.ActionLog = append(event.ActionLog, action)
	return event, true, nil
}

func ListMonitorAlertEventActions(eventID string) []model.MonitorAlertAction {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,event_id,action,actor,`+"`from`"+`,`+"`to`"+`,reason,assignee,trace_id,created_at FROM monitor_alert_event_actions WHERE event_id=? ORDER BY created_at ASC`, eventID)
		if err == nil {
			defer rows.Close()
			return scanMonitorAlertActions(rows)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	return append([]model.MonitorAlertAction{}, monitorEventActions[eventID]...)
}

func mergeMonitorAlertEvent(existing, incoming *model.MonitorAlertEvent) *model.MonitorAlertEvent {
	merged := copyMonitorAlertEvent(existing)
	merged.Count += max(1, incoming.Count)
	merged.LastSeen = laterMonitorAlertEventTime(merged.LastSeen, incoming.LastSeen)
	merged.UpdatedAt = laterMonitorAlertEventTime(merged.UpdatedAt, incoming.UpdatedAt)
	merged.Value = firstNonEmpty(incoming.Value, merged.Value)
	merged.Labels = mergeStringMap(merged.Labels, incoming.Labels)
	merged.Annotations = mergeStringMap(merged.Annotations, incoming.Annotations)
	return merged
}

func laterMonitorAlertEventTime(existing, incoming time.Time) time.Time {
	if incoming.IsZero() || existing.After(incoming) {
		return existing
	}
	return incoming
}

func findCurrentMonitorAlertEventLocked(event *model.MonitorAlertEvent) *model.MonitorAlertEvent {
	if existing := monitorEventsCurrent[event.ID]; existing != nil {
		return existing
	}
	for _, existing := range monitorEventsCurrent {
		if existing.Fingerprint == event.Fingerprint && event.Fingerprint != "" {
			return existing
		}
	}
	return nil
}

func getCurrentMonitorAlertEventByFingerprint(fingerprint string) (*model.MonitorAlertEvent, bool) {
	if fingerprint == "" || !mysqlOK {
		return nil, false
	}
	row := db.QueryRow(monitorEventSelectSQL("monitor_alert_events_current")+` WHERE fingerprint=?`, fingerprint)
	return scanMonitorAlertEventRow(row)
}

func isSensitiveKey(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(k, "api_key") || strings.Contains(k, "apikey") || strings.Contains(k, "auth") || strings.Contains(k, "token") || strings.Contains(k, "secret") || strings.Contains(k, "password") || strings.Contains(k, "cookie") || strings.Contains(k, "private") || strings.Contains(k, "dsn")
}

func normalizeMonitorAlertEvent(event *model.MonitorAlertEvent) *model.MonitorAlertEvent {
	now := time.Now()
	cp := copyMonitorAlertEvent(event)
	if cp.ID == "" {
		cp.ID = NewID()
	}
	if cp.Status == "" {
		cp.Status = model.MonitorAlertEventStatusFiring
	}
	if cp.Count <= 0 {
		cp.Count = 1
	}
	if cp.FirstSeen.IsZero() {
		cp.FirstSeen = now
	}
	if cp.LastSeen.IsZero() {
		cp.LastSeen = now
	}
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
	}
	cp.UpdatedAt = now
	cp.Labels = sanitizeStringMap(cp.Labels)
	cp.Annotations = sanitizeStringMap(cp.Annotations)
	cp.Fingerprint = model.GenerateMonitorAlertEventFingerprint(cp)
	return cp
}

func normalizeMonitorEventAction(event *model.MonitorAlertEvent, action model.MonitorAlertAction, now time.Time) model.MonitorAlertAction {
	if action.ID == "" {
		action.ID = NewID()
	}
	action.EventID = event.ID
	action.From = event.Status
	if action.CreatedAt.IsZero() {
		action.CreatedAt = now
	}
	return action
}

func applyMonitorEventMutation(event *model.MonitorAlertEvent, action model.MonitorAlertAction, now time.Time) model.MonitorAlertAction {
	next, _ := model.ValidateMonitorAlertEventTransition(event.Status, action.Action)
	switch action.Action {
	case model.MonitorAlertEventActionAck:
		event.Status = next
		event.AckBy = firstNonEmpty(action.Actor, "admin")
	case model.MonitorAlertEventActionAssign:
		event.Status = next
		event.Assignee = firstNonEmpty(action.Assignee, action.Actor)
	case model.MonitorAlertEventActionMute:
		event.Status = next
	case model.MonitorAlertEventActionResolve:
		event.Status = next
		event.Resolution = action.Reason
		event.ResolvedAt = &now
	case model.MonitorAlertEventActionArchive:
		event.Status = next
		event.ArchivedAt = &now
	}
	action.To = event.Status
	event.UpdatedAt = now
	return action
}

func listMonitorEventsMemory(current bool) []model.MonitorAlertEvent {
	mu.RLock()
	defer mu.RUnlock()
	source := monitorEventsHistory
	if current {
		source = monitorEventsCurrent
	}
	out := make([]model.MonitorAlertEvent, 0, len(source))
	for _, event := range source {
		cp := copyMonitorAlertEvent(event)
		cp.ActionLog = append([]model.MonitorAlertAction{}, monitorEventActions[event.ID]...)
		out = append(out, *cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastSeen.After(out[j].LastSeen) })
	return out
}

func copyMonitorAlertEvent(in *model.MonitorAlertEvent) *model.MonitorAlertEvent {
	if in == nil {
		return nil
	}
	cp := *in
	cp.Labels = copyStringMap(in.Labels)
	cp.Annotations = copyStringMap(in.Annotations)
	cp.ActionLog = append([]model.MonitorAlertAction{}, in.ActionLog...)
	return &cp
}

type monitorEventScanner interface {
	Scan(dest ...any) error
}

func monitorEventSelectSQL(table string) string {
	return `SELECT id,rule_id,rule_version,event_key,name,severity,status,datasource_id,target_id,target_ident,COALESCE(labels,'{}'),COALESCE(annotations,'{}'),value,fingerprint,count,first_seen,last_seen,ack_by,assignee,resolution,archived_at,resolved_at,created_at,updated_at FROM ` + table
}

func scanMonitorAlertEventRow(row monitorEventScanner) (*model.MonitorAlertEvent, bool) {
	event := model.MonitorAlertEvent{}
	var labels, annotations string
	var archived, resolved sql.NullTime
	err := row.Scan(&event.ID, &event.RuleID, &event.RuleVersion, &event.EventKey, &event.Name, &event.Severity, &event.Status, &event.DatasourceID, &event.TargetID, &event.TargetIdent, &labels, &annotations, &event.Value, &event.Fingerprint, &event.Count, &event.FirstSeen, &event.LastSeen, &event.AckBy, &event.Assignee, &event.Resolution, &archived, &resolved, &event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		return nil, false
	}
	if !unmarshalStringMap(labels, &event.Labels) || !unmarshalStringMap(annotations, &event.Annotations) {
		return nil, false
	}
	if archived.Valid {
		event.ArchivedAt = &archived.Time
	}
	if resolved.Valid {
		event.ResolvedAt = &resolved.Time
	}
	return &event, true
}

func scanMonitorAlertEvents(rows *sql.Rows) []model.MonitorAlertEvent {
	out := []model.MonitorAlertEvent{}
	for rows.Next() {
		if event, ok := scanMonitorAlertEventRow(rows); ok {
			event.ActionLog = ListMonitorAlertEventActions(event.ID)
			out = append(out, *event)
		}
	}
	if err := rows.Err(); err != nil {
		return out
	}
	return out
}
