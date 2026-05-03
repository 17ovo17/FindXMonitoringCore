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
		table = "monitor_alert_events_current"
	}
	_, err = db.Exec(`REPLACE INTO `+table+` (id,rule_id,rule_version,event_key,name,severity,status,datasource_id,target_id,target_ident,labels,annotations,value,fingerprint,count,first_seen,last_seen,ack_by,assignee,resolution,archived_at,resolved_at,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, event.ID, event.RuleID, event.RuleVersion, event.EventKey, event.Name, event.Severity, event.Status, event.DatasourceID, event.TargetID, event.TargetIdent, string(labels), string(annotations), event.Value, event.Fingerprint, event.Count, event.FirstSeen, event.LastSeen, event.AckBy, event.Assignee, event.Resolution, nullableTime(event.ArchivedAt), nullableTime(event.ResolvedAt), event.CreatedAt, event.UpdatedAt)
	return err
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
