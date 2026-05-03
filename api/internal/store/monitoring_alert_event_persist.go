package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"ai-workbench-api/internal/model"
)

func persistMonitorAlertEvent(event *model.MonitorAlertEvent, current bool) error {
	labels, err := json.Marshal(event.Labels)
	if err != nil {
		return fmt.Errorf("marshal alert event labels: %w", err)
	}
	annotations, err := json.Marshal(event.Annotations)
	if err != nil {
		return fmt.Errorf("marshal alert event annotations: %w", err)
	}
	table := "monitor_alert_events_history"
	if current {
		return persistMonitorAlertEventCurrentPayload(event, string(labels), string(annotations), event.Count, false)
	}
	_, err = db.Exec(`REPLACE INTO `+table+` (id,rule_id,rule_version,event_key,name,severity,status,datasource_id,target_id,target_ident,labels,annotations,value,fingerprint,count,first_seen,last_seen,ack_by,assignee,resolution,archived_at,resolved_at,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, event.ID, event.RuleID, event.RuleVersion, event.EventKey, event.Name, event.Severity, event.Status, event.DatasourceID, event.TargetID, event.TargetIdent, string(labels), string(annotations), event.Value, event.Fingerprint, event.Count, event.FirstSeen, event.LastSeen, event.AckBy, event.Assignee, event.Resolution, nullableTime(event.ArchivedAt), nullableTime(event.ResolvedAt), event.CreatedAt, event.UpdatedAt)
	return err
}

func persistMonitorAlertEventCurrent(event *model.MonitorAlertEvent, countDelta int, increment bool) error {
	labels, err := json.Marshal(event.Labels)
	if err != nil {
		return fmt.Errorf("marshal alert event labels: %w", err)
	}
	annotations, err := json.Marshal(event.Annotations)
	if err != nil {
		return fmt.Errorf("marshal alert event annotations: %w", err)
	}
	if countDelta <= 0 {
		countDelta = event.Count
	}
	return persistMonitorAlertEventCurrentPayload(event, string(labels), string(annotations), countDelta, increment)
}

func persistMonitorAlertEventCurrentPayload(event *model.MonitorAlertEvent, labels, annotations string, count int, increment bool) error {
	_, err := db.Exec(monitorCurrentEventUpsertSQL(increment), event.ID, event.RuleID, event.RuleVersion, event.EventKey, event.Name, event.Severity, event.Status, event.DatasourceID, event.TargetID, event.TargetIdent, labels, annotations, event.Value, event.Fingerprint, count, event.FirstSeen, event.LastSeen, event.AckBy, event.Assignee, event.Resolution, nullableTime(event.ArchivedAt), nullableTime(event.ResolvedAt), event.CreatedAt, event.UpdatedAt)
	return err
}

func monitorCurrentEventUpsertSQL(increment bool) string {
	countUpdate := "count=VALUES(count)"
	if increment {
		countUpdate = "count=count+VALUES(count)"
	}
	updates := `labels=VALUES(labels),annotations=VALUES(annotations),value=VALUES(value),` + countUpdate + `,last_seen=GREATEST(last_seen,VALUES(last_seen)),updated_at=GREATEST(updated_at,VALUES(updated_at))`
	if !increment {
		updates += `,status=VALUES(status),ack_by=VALUES(ack_by),assignee=VALUES(assignee),resolution=VALUES(resolution),archived_at=VALUES(archived_at),resolved_at=VALUES(resolved_at)`
	}
	return `INSERT INTO monitor_alert_events_current (id,rule_id,rule_version,event_key,name,severity,status,datasource_id,target_id,target_ident,labels,annotations,value,fingerprint,count,first_seen,last_seen,ack_by,assignee,resolution,archived_at,resolved_at,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE ` + updates
}

func persistMonitorAlertAction(action model.MonitorAlertAction) error {
	_, err := db.Exec(`INSERT INTO monitor_alert_event_actions (id,event_id,action,actor,`+"`from`"+`,`+"`to`"+`,reason,assignee,trace_id,created_at) VALUES (?,?,?,?,?,?,?,?,?,?)`, action.ID, action.EventID, action.Action, action.Actor, action.From, action.To, action.Reason, action.Assignee, action.TraceID, action.CreatedAt)
	return err
}

func scanMonitorAlertActions(rows *sql.Rows) []model.MonitorAlertAction {
	out := []model.MonitorAlertAction{}
	for rows.Next() {
		action := model.MonitorAlertAction{}
		if err := rows.Scan(&action.ID, &action.EventID, &action.Action, &action.Actor, &action.From, &action.To, &action.Reason, &action.Assignee, &action.TraceID, &action.CreatedAt); err != nil {
			continue
		}
		out = append(out, action)
	}
	if err := rows.Err(); err != nil {
		return out
	}
	return out
}
