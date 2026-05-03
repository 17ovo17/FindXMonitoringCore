package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"ai-workbench-api/internal/model"
)

func persistMonitorAlertRule(rule *model.MonitorAlertRule) error {
	selector, err := json.Marshal(rule.TargetSelector)
	if err != nil {
		return fmt.Errorf("marshal alert rule selector: %w", err)
	}
	labels, err := json.Marshal(rule.Labels)
	if err != nil {
		return fmt.Errorf("marshal alert rule labels: %w", err)
	}
	annotations, err := json.Marshal(rule.Annotations)
	if err != nil {
		return fmt.Errorf("marshal alert rule annotations: %w", err)
	}
	_, err = db.Exec(`REPLACE INTO monitor_alert_rules (id,name,query,severity,datasource_id,target_selector,labels,annotations,enabled,version,for_duration,no_data_policy,status,created_by,updated_by,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, rule.ID, rule.Name, rule.Query, rule.Severity, rule.DatasourceID, string(selector), string(labels), string(annotations), rule.Enabled, rule.Version, rule.ForDuration, rule.NoDataPolicy, rule.Status, rule.CreatedBy, rule.UpdatedBy, rule.CreatedAt, rule.UpdatedAt)
	return err
}

func persistMonitorAlertRuleVersion(v model.MonitorAlertRuleVersion) error {
	selector, err := json.Marshal(v.TargetSelector)
	if err != nil {
		return fmt.Errorf("marshal alert rule version selector: %w", err)
	}
	labels, err := json.Marshal(v.Labels)
	if err != nil {
		return fmt.Errorf("marshal alert rule version labels: %w", err)
	}
	annotations, err := json.Marshal(v.Annotations)
	if err != nil {
		return fmt.Errorf("marshal alert rule version annotations: %w", err)
	}
	_, err = db.Exec(`REPLACE INTO monitor_alert_rule_versions (id,rule_id,version,name,query,severity,datasource_id,target_selector,labels,annotations,enabled,for_duration,no_data_policy,status,created_by,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.RuleID, v.Version, v.Name, v.Query, v.Severity, v.DatasourceID, string(selector), string(labels), string(annotations), v.Enabled, v.ForDuration, v.NoDataPolicy, v.Status, v.CreatedBy, v.CreatedAt)
	return err
}

func scanMonitorRuleVersions(rows *sql.Rows) []model.MonitorAlertRuleVersion {
	out := []model.MonitorAlertRuleVersion{}
	for rows.Next() {
		item := model.MonitorAlertRuleVersion{}
		var selector, labels, annotations string
		if err := rows.Scan(&item.ID, &item.RuleID, &item.Version, &item.Name, &item.Query, &item.Severity, &item.DatasourceID, &selector, &labels, &annotations, &item.Enabled, &item.ForDuration, &item.NoDataPolicy, &item.Status, &item.CreatedBy, &item.CreatedAt); err != nil {
			continue
		}
		if !unmarshalStringMap(selector, &item.TargetSelector) ||
			!unmarshalStringMap(labels, &item.Labels) ||
			!unmarshalStringMap(annotations, &item.Annotations) {
			continue
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return out
	}
	return out
}

func unmarshalMonitorRuleJSON(rule *model.MonitorAlertRule, selector, labels, annotations string) bool {
	return unmarshalStringMap(selector, &rule.TargetSelector) &&
		unmarshalStringMap(labels, &rule.Labels) &&
		unmarshalStringMap(annotations, &rule.Annotations)
}

func unmarshalStringMap(raw string, dest *map[string]string) bool {
	if raw == "" {
		*dest = map[string]string{}
		return true
	}
	return json.Unmarshal([]byte(raw), dest) == nil
}

func copyMonitorAlertRule(in *model.MonitorAlertRule) *model.MonitorAlertRule {
	if in == nil {
		return nil
	}
	cp := *in
	cp.TargetSelector = copyStringMap(in.TargetSelector)
	cp.Labels = copyStringMap(in.Labels)
	cp.Annotations = copyStringMap(in.Annotations)
	return &cp
}
