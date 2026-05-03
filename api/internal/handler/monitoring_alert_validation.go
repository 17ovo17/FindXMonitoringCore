package handler

import (
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/monitoring"
)

func validateMonitorAlertRule(rule *model.MonitorAlertRule) []model.MonitorTryCheck {
	checks := []model.MonitorTryCheck{}
	checks = append(checks, monitorCheck("name", strings.TrimSpace(rule.Name) != "", "name is required"))
	checks = append(checks, monitorCheck("datasource_id", validMonitorDatasource(rule.DatasourceID), "valid prometheus datasource_id is required"))
	checks = append(checks, monitorCheck("severity", validMonitorSeverity(rule.Severity), "severity must be critical, warning, info, p0, p1, p2, or p3"))
	checks = append(checks, monitorCheck("query", validMonitorPromQL(rule.Query), "query is required and must not contain dangerous characters"))
	checks = append(checks, monitorCheck("no_data_policy", validNoDataPolicy(rule.NoDataPolicy), "no_data_policy must be keep_state, alerting, or ok"))
	return checks
}

func monitorCheck(name string, ok bool, message string) model.MonitorTryCheck {
	status := "pass"
	if !ok {
		status = "fail"
	}
	return model.MonitorTryCheck{Name: name, Status: status, Message: message}
}

func monitorChecksOK(checks []model.MonitorTryCheck) bool {
	for _, check := range checks {
		if check.Status != "pass" {
			return false
		}
	}
	return true
}

func validMonitorSeverity(severity string) bool {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case model.MonitorAlertSeverityCritical, model.MonitorAlertSeverityWarning,
		model.MonitorAlertSeverityInfo, model.MonitorAlertSeverityP0,
		model.MonitorAlertSeverityP1, model.MonitorAlertSeverityP2,
		model.MonitorAlertSeverityP3:
		return true
	default:
		return false
	}
}

func validNoDataPolicy(policy string) bool {
	switch strings.TrimSpace(policy) {
	case "", model.MonitorNoDataPolicyKeepState, model.MonitorNoDataPolicyAlerting, model.MonitorNoDataPolicyOK:
		return true
	default:
		return false
	}
}

func validMonitorPromQL(query string) bool {
	return monitoring.ValidatePromQL(query) == nil
}

func validMonitorDatasource(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}
	for _, ds := range monitoring.PrometheusDatasourcesFromConfig() {
		if ds.ID == id && monitoring.DatasourceURLReady(ds.URL) {
			return true
		}
	}
	_, _, err := monitoring.ResolvePrometheusDatasourceFromConfig(id)
	return err == nil
}

func monitorTryRunDetails(checks []model.MonitorTryCheck) map[string]any {
	return map[string]any{"checks": checks}
}

func storeMonitorQueryHash(query string) string {
	return monitoring.QueryHash(query)
}

func requestActor(c interface{ GetHeader(string) string }) string {
	if actor := strings.TrimSpace(c.GetHeader("X-Admin-Token")); actor != "" {
		return "admin-token"
	}
	return "system"
}
