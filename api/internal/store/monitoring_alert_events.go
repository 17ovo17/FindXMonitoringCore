package store

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"ai-workbench-api/internal/model"
)

var ErrTerminalMonitorAlertEvent = errors.New("terminal alert event cannot be changed")

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
	mu.Lock()
	if existing := monitorEventsCurrent[normalized.ID]; existing != nil {
		existing.Count++
		existing.LastSeen = normalized.LastSeen
		existing.UpdatedAt = normalized.UpdatedAt
		normalized = copyMonitorAlertEvent(existing)
	} else {
		monitorEventsCurrent[normalized.ID] = copyMonitorAlertEvent(normalized)
	}
	mu.Unlock()
	if mysqlOK {
		if err := persistMonitorAlertEvent(normalized, true); err != nil {
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
	if isTerminalMonitorAlertStatus(event.Status) {
		return nil, true, ErrTerminalMonitorAlertEvent
	}
	now := time.Now()
	action = normalizeMonitorEventAction(event, action, now)
	action = applyMonitorEventMutation(event, action, now)
	mu.Lock()
	delete(monitorEventsCurrent, id)
	if event.Status == "resolved" || event.Status == "archived" {
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
		current := event.Status != "resolved" && event.Status != "archived"
		if err := persistMonitorAlertEvent(event, current); err != nil {
			return event, true, err
		}
		if event.Status == "resolved" || event.Status == "archived" {
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

func normalizeMonitorAlertEvent(event *model.MonitorAlertEvent) *model.MonitorAlertEvent {
	now := time.Now()
	cp := copyMonitorAlertEvent(event)
	if cp.ID == "" {
		cp.ID = NewID()
	}
	if cp.Status == "" {
		cp.Status = "firing"
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
	cp.Fingerprint = monitorEventFingerprint(cp)
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

func isTerminalMonitorAlertStatus(status string) bool {
	return status == "resolved" || status == "archived"
}

func applyMonitorEventMutation(event *model.MonitorAlertEvent, action model.MonitorAlertAction, now time.Time) model.MonitorAlertAction {
	switch action.Action {
	case "ack":
		event.Status = "acknowledged"
		event.AckBy = firstNonEmpty(action.Actor, "admin")
	case "assign":
		event.Status = "assigned"
		event.Assignee = firstNonEmpty(action.Assignee, action.Actor)
	case "resolve":
		event.Status = "resolved"
		event.Resolution = action.Reason
		event.ResolvedAt = &now
	case "archive":
		event.Status = "archived"
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

func monitorEventFingerprint(event *model.MonitorAlertEvent) string {
	if event.Fingerprint != "" {
		return event.Fingerprint
	}
	key := event.RuleID + "|" + event.EventKey + "|" + event.TargetID + "|" + event.TargetIdent + "|" + event.Name
	return fmt.Sprintf("%x", sha1.Sum([]byte(key)))
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
