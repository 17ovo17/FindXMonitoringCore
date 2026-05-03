package store

import (
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"ai-workbench-api/internal/model"
)

func ListMonitorAlertRules() []model.MonitorAlertRule {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,name,query,severity,datasource_id,COALESCE(target_selector,'{}'),COALESCE(labels,'{}'),COALESCE(annotations,'{}'),enabled,version,for_duration,no_data_policy,status,created_by,updated_by,created_at,updated_at FROM monitor_alert_rules ORDER BY updated_at DESC LIMIT 1000`)
		if err == nil {
			defer rows.Close()
			return scanMonitorAlertRules(rows)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.MonitorAlertRule, 0, len(monitorAlertRules))
	for _, rule := range monitorAlertRules {
		out = append(out, *copyMonitorAlertRule(rule))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func GetMonitorAlertRule(id string) (*model.MonitorAlertRule, bool) {
	if mysqlOK {
		row := db.QueryRow(`SELECT id,name,query,severity,datasource_id,COALESCE(target_selector,'{}'),COALESCE(labels,'{}'),COALESCE(annotations,'{}'),enabled,version,for_duration,no_data_policy,status,created_by,updated_by,created_at,updated_at FROM monitor_alert_rules WHERE id=?`, id)
		if rule, ok := scanMonitorAlertRuleRow(row); ok {
			return rule, true
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	rule, ok := monitorAlertRules[id]
	return copyMonitorAlertRule(rule), ok
}

func SaveMonitorAlertRule(rule *model.MonitorAlertRule, actor string) (*model.MonitorAlertRule, error) {
	now := time.Now()
	existing, exists := GetMonitorAlertRule(rule.ID)
	normalized := normalizeMonitorAlertRule(rule, existing, actor, now)
	if !exists {
		normalized.Version = 1
		normalized.CreatedAt = now
	}
	version := snapshotMonitorRule(normalized, actor, now)
	mu.Lock()
	monitorAlertRules[normalized.ID] = copyMonitorAlertRule(normalized)
	monitorRuleVersions[normalized.ID] = append(monitorRuleVersions[normalized.ID], version)
	mu.Unlock()
	if mysqlOK {
		if err := persistMonitorAlertRule(normalized); err != nil {
			return copyMonitorAlertRule(normalized), err
		}
		if err := persistMonitorAlertRuleVersion(version); err != nil {
			return copyMonitorAlertRule(normalized), err
		}
	}
	return copyMonitorAlertRule(normalized), nil
}

func DeleteMonitorAlertRule(id string) (bool, error) {
	found := false
	mu.Lock()
	if _, ok := monitorAlertRules[id]; ok {
		delete(monitorAlertRules, id)
		found = true
	}
	mu.Unlock()
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM monitor_alert_rules WHERE id=?`, id)
		if err != nil {
			return found, err
		}
		if res == nil {
			return found, fmt.Errorf("delete alert rule returned no result")
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return found, err
		}
		if rows > 0 {
			found = true
		}
	}
	return found, nil
}

func SetMonitorAlertRuleEnabled(id string, enabled bool, actor string) (*model.MonitorAlertRule, bool, error) {
	rule, ok := GetMonitorAlertRule(id)
	if !ok {
		return nil, false, nil
	}
	rule.Enabled = enabled
	rule.Status = model.MonitorAlertRuleStatusActive
	if !enabled {
		rule.Status = model.MonitorAlertRuleStatusDisabled
	}
	rule.Version++
	out, err := SaveMonitorAlertRule(rule, actor)
	return out, true, err
}

func CloneMonitorAlertRule(id, actor string) (*model.MonitorAlertRule, bool, error) {
	rule, ok := GetMonitorAlertRule(id)
	if !ok {
		return nil, false, nil
	}
	rule.ID = ""
	rule.Name = rule.Name + " Copy"
	rule.Version = 0
	rule.CreatedAt = time.Time{}
	out, err := SaveMonitorAlertRule(rule, actor)
	return out, true, err
}

func RollbackMonitorAlertRule(id string, version int, actor string) (*model.MonitorAlertRule, bool, error) {
	versions := ListMonitorAlertRuleVersions(id)
	if len(versions) == 0 {
		return nil, false, nil
	}
	selected, ok := selectMonitorRuleVersion(versions, version)
	if !ok {
		return nil, false, nil
	}
	current, ok := GetMonitorAlertRule(id)
	if !ok {
		return nil, false, nil
	}
	restored := ruleFromVersion(selected, current.Version+1, actor)
	out, err := SaveMonitorAlertRule(restored, actor)
	return out, true, err
}

func ListMonitorAlertRuleVersions(ruleID string) []model.MonitorAlertRuleVersion {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,rule_id,version,name,query,severity,datasource_id,COALESCE(target_selector,'{}'),COALESCE(labels,'{}'),COALESCE(annotations,'{}'),enabled,for_duration,no_data_policy,status,created_by,created_at FROM monitor_alert_rule_versions WHERE rule_id=? ORDER BY version DESC`, ruleID)
		if err == nil {
			defer rows.Close()
			return scanMonitorRuleVersions(rows)
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := append([]model.MonitorAlertRuleVersion{}, monitorRuleVersions[ruleID]...)
	sort.Slice(out, func(i, j int) bool { return out[i].Version > out[j].Version })
	return out
}

func AddMonitorAlertEvalLog(log model.MonitorAlertEvalLog) (model.MonitorAlertEvalLog, error) {
	if log.ID == "" {
		log.ID = NewID()
	}
	if mysqlOK {
		details, err := json.Marshal(log.Details)
		if err != nil {
			return log, fmt.Errorf("marshal monitor alert eval log details: %w", err)
		}
		_, err = db.Exec(`INSERT INTO monitor_alert_rule_eval_logs (id,rule_id,rule_version,status,message,details,started_at,finished_at,duration_ms,datasource_id,query_hash) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, log.ID, log.RuleID, log.RuleVersion, log.Status, log.Message, string(details), log.StartedAt, log.FinishedAt, log.DurationMs, log.DatasourceID, log.QueryHash)
		if err != nil {
			return log, fmt.Errorf("insert monitor alert eval log: %w", err)
		}
	}
	return log, nil
}

func normalizeMonitorAlertRule(rule, existing *model.MonitorAlertRule, actor string, now time.Time) *model.MonitorAlertRule {
	cp := copyMonitorAlertRule(rule)
	if cp.ID == "" {
		cp.ID = NewID()
	}
	if existing != nil {
		cp.CreatedAt = existing.CreatedAt
		cp.Version = max(existing.Version+1, cp.Version)
	}
	cp.TargetSelector = sanitizeStringMap(cp.TargetSelector)
	cp.Labels = sanitizeStringMap(cp.Labels)
	cp.Annotations = sanitizeStringMap(cp.Annotations)
	cp.UpdatedBy = actor
	cp.UpdatedAt = now
	cp.Status = monitorRuleStatus(cp.Enabled, cp.Status)
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
		cp.CreatedBy = actor
	}
	if cp.NoDataPolicy == "" {
		cp.NoDataPolicy = model.MonitorNoDataPolicyKeepState
	}
	return cp
}

func snapshotMonitorRule(rule *model.MonitorAlertRule, actor string, now time.Time) model.MonitorAlertRuleVersion {
	return model.MonitorAlertRuleVersion{
		ID: rule.ID + ":v" + fmt.Sprint(rule.Version), RuleID: rule.ID, Version: rule.Version,
		Name: rule.Name, Query: rule.Query, Severity: rule.Severity, DatasourceID: rule.DatasourceID,
		TargetSelector: copyStringMap(rule.TargetSelector), Labels: copyStringMap(rule.Labels),
		Annotations: copyStringMap(rule.Annotations), Enabled: rule.Enabled,
		ForDuration: rule.ForDuration, NoDataPolicy: rule.NoDataPolicy, Status: rule.Status,
		CreatedBy: actor, CreatedAt: now,
	}
}

func selectMonitorRuleVersion(versions []model.MonitorAlertRuleVersion, version int) (model.MonitorAlertRuleVersion, bool) {
	for _, item := range versions {
		if version <= 0 || item.Version == version {
			return item, true
		}
	}
	return model.MonitorAlertRuleVersion{}, false
}

func ruleFromVersion(v model.MonitorAlertRuleVersion, version int, actor string) *model.MonitorAlertRule {
	return &model.MonitorAlertRule{ID: v.RuleID, Name: v.Name, Query: v.Query, Severity: v.Severity, DatasourceID: v.DatasourceID, TargetSelector: v.TargetSelector, Labels: v.Labels, Annotations: v.Annotations, Enabled: v.Enabled, Version: version, ForDuration: v.ForDuration, NoDataPolicy: v.NoDataPolicy, Status: v.Status, UpdatedBy: actor}
}

func monitorRuleStatus(enabled bool, status string) string {
	if !enabled {
		return model.MonitorAlertRuleStatusDisabled
	}
	if status == "" || status == model.MonitorAlertRuleStatusDisabled {
		return model.MonitorAlertRuleStatusActive
	}
	return status
}

func monitorQueryHash(query string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(query)))
}

type monitorRuleScanner interface {
	Scan(dest ...any) error
}

func scanMonitorAlertRuleRow(row monitorRuleScanner) (*model.MonitorAlertRule, bool) {
	rule := model.MonitorAlertRule{}
	var selector, labels, annotations string
	err := row.Scan(&rule.ID, &rule.Name, &rule.Query, &rule.Severity, &rule.DatasourceID, &selector, &labels, &annotations, &rule.Enabled, &rule.Version, &rule.ForDuration, &rule.NoDataPolicy, &rule.Status, &rule.CreatedBy, &rule.UpdatedBy, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return nil, false
	}
	if !unmarshalMonitorRuleJSON(&rule, selector, labels, annotations) {
		return nil, false
	}
	return &rule, true
}

func scanMonitorAlertRules(rows *sql.Rows) []model.MonitorAlertRule {
	out := []model.MonitorAlertRule{}
	for rows.Next() {
		if rule, ok := scanMonitorAlertRuleRow(rows); ok {
			out = append(out, *rule)
		}
	}
	if err := rows.Err(); err != nil {
		return out
	}
	return out
}
