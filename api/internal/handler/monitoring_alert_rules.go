package handler

import (
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/evaluator"
	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/monitoring"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func ListMonitorAlertRules(c *gin.Context) {
	rules := store.ListMonitorAlertRules()
	status := strings.TrimSpace(c.Query("status"))
	if status != "" {
		rules = filterMonitorRules(rules, status)
	}
	c.JSON(http.StatusOK, rules)
}

func CreateMonitorAlertRule(c *gin.Context) {
	saveMonitorAlertRule(c, "")
}

func GetMonitorAlertRule(c *gin.Context) {
	rule, ok := store.GetMonitorAlertRule(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rule": rule, "versions": store.ListMonitorAlertRuleVersions(rule.ID)})
}

func UpdateMonitorAlertRule(c *gin.Context) {
	saveMonitorAlertRule(c, c.Param("id"))
}

func DeleteMonitorAlertRule(c *gin.Context) {
	ok, err := store.DeleteMonitorAlertRule(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alert rule delete failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func EnableMonitorAlertRule(c *gin.Context) {
	setMonitorRuleEnabled(c, true)
}

func DisableMonitorAlertRule(c *gin.Context) {
	setMonitorRuleEnabled(c, false)
}

func CloneMonitorAlertRule(c *gin.Context) {
	rule, ok, err := store.CloneMonitorAlertRule(c.Param("id"), requestActor(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alert rule clone failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func RollbackMonitorAlertRule(c *gin.Context) {
	var req struct {
		Version int    `json:"version"`
		Actor   string `json:"actor"`
	}
	_ = c.ShouldBindJSON(&req)
	rule, ok, err := store.RollbackMonitorAlertRule(c.Param("id"), req.Version, firstNonEmpty(req.Actor, requestActor(c)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alert rule rollback failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert rule version not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func TryRunMonitorAlertRule(c *gin.Context) {
	rule, ok := monitorRuleForTryRun(c)
	if !ok {
		return
	}
	started := time.Now()
	checks := validateMonitorAlertRuleForTryRun(rule)
	if !monitorChecksOK(checks) {
		writeTryRunValidationResult(c, rule, checks, started)
		return
	}
	dsID, base, ok := resolveMonitoringPrometheus(c, rule.DatasourceID)
	if !ok {
		return
	}
	rule.DatasourceID = dsID
	prom, err := monitoring.NewPrometheusGateway(nil).QueryInstant(c.Request.Context(), monitoring.PrometheusQueryRequest{
		BaseURL: base, Query: rule.Query, Timeout: monitoring.DefaultInstantTimeout,
	})
	if err != nil {
		writeTryRunUpstreamError(c, rule, checks, started, prom, err)
		return
	}
	eval, evalErr := evaluator.EvaluateRule(evaluatorRule(rule, prom.QueryHash), prom)
	writeTryRunEvaluation(c, rule, checks, started, eval, evalErr)
}

func saveMonitorAlertRule(c *gin.Context, id string) {
	var rule model.MonitorAlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert rule payload"})
		return
	}
	if id != "" {
		rule.ID = id
	}
	checks := validateMonitorAlertRule(&rule)
	if !monitorChecksOK(checks) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert rule", "checks": checks})
		return
	}
	out, err := store.SaveMonitorAlertRule(&rule, requestActor(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alert rule save failed"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func setMonitorRuleEnabled(c *gin.Context, enabled bool) {
	rule, ok, err := store.SetMonitorAlertRuleEnabled(c.Param("id"), enabled, requestActor(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alert rule update failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func monitorRuleForTryRun(c *gin.Context) (*model.MonitorAlertRule, bool) {
	var payload model.MonitorAlertRule
	if err := c.ShouldBindJSON(&payload); err == nil && monitorTryRunPayloadProvided(payload) {
		if payload.ID == "" {
			payload.ID = c.Param("id")
		}
		return &payload, true
	}
	rule, ok := store.GetMonitorAlertRule(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
		return nil, false
	}
	return rule, true
}

func monitorTryRunPayloadProvided(payload model.MonitorAlertRule) bool {
	return strings.TrimSpace(payload.ID) != "" || strings.TrimSpace(payload.Name) != "" ||
		strings.TrimSpace(payload.Query) != "" || strings.TrimSpace(payload.DatasourceID) != "" ||
		strings.TrimSpace(payload.Severity) != ""
}

func filterMonitorRules(rules []model.MonitorAlertRule, status string) []model.MonitorAlertRule {
	out := []model.MonitorAlertRule{}
	for _, rule := range rules {
		if rule.Status == status {
			out = append(out, rule)
		}
	}
	return out
}

func validateMonitorAlertRuleForTryRun(rule *model.MonitorAlertRule) []model.MonitorTryCheck {
	checks := []model.MonitorTryCheck{}
	checks = append(checks, monitorCheck("name", strings.TrimSpace(rule.Name) != "", "name is required"))
	checks = append(checks, monitorCheck("datasource_id", strings.TrimSpace(rule.DatasourceID) != "", "datasource_id is required"))
	checks = append(checks, monitorCheck("severity", validMonitorSeverity(rule.Severity), "severity must be critical, warning, info, p0, p1, p2, or p3"))
	checks = append(checks, monitorCheck("query", validMonitorPromQL(rule.Query), "query is required and must not contain dangerous characters"))
	checks = append(checks, monitorCheck("no_data_policy", validNoDataPolicy(rule.NoDataPolicy), "no_data_policy must be keep_state, alerting, or ok"))
	return checks
}

func writeTryRunValidationResult(c *gin.Context, rule *model.MonitorAlertRule, checks []model.MonitorTryCheck, started time.Time) {
	finished := time.Now()
	details := monitorTryRunDetails(checks)
	details["mode"] = "validation"
	details["promql_executed"] = false
	log, err := store.AddMonitorAlertEvalLog(model.MonitorAlertEvalLog{
		RuleID: rule.ID, RuleVersion: rule.Version, Status: "invalid",
		Message: "validation failed; PromQL was not executed", StartedAt: started, FinishedAt: finished,
		DurationMs: finished.Sub(started).Milliseconds(), DatasourceID: rule.DatasourceID,
		QueryHash: storeMonitorQueryHash(rule.Query), Details: details,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alert rule evaluation log write failed"})
		return
	}
	c.JSON(http.StatusOK, model.MonitorAlertTryRunResult{OK: false, Status: "invalid", Checks: checks, Rule: safeMonitorTryRunRule(rule), EvalLog: log})
}

func writeTryRunUpstreamError(c *gin.Context, rule *model.MonitorAlertRule, checks []model.MonitorTryCheck, started time.Time, prom monitoring.PrometheusCallResult, upstreamErr error) {
	finished := time.Now()
	details := monitorTryRunDetails(checks)
	details["promql_executed"] = true
	details["query_hash"] = storeMonitorQueryHash(rule.Query)
	details["upstream_status"] = monitoring.HTTPStatus(upstreamErr)
	log, err := store.AddMonitorAlertEvalLog(model.MonitorAlertEvalLog{
		RuleID: rule.ID, RuleVersion: rule.Version, Status: "upstream_error",
		Message: "prometheus query failed", StartedAt: started, FinishedAt: finished,
		DurationMs: finished.Sub(started).Milliseconds(), DatasourceID: rule.DatasourceID,
		QueryHash: firstNonEmpty(prom.QueryHash, storeMonitorQueryHash(rule.Query)), Details: details,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alert rule evaluation log write failed"})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "prometheus query failed", "eval_log": log})
}

func writeTryRunEvaluation(c *gin.Context, rule *model.MonitorAlertRule, checks []model.MonitorTryCheck, started time.Time, eval evaluator.Result, evalErr error) {
	status, ok := tryRunEvalStatus(eval, evalErr)
	finished := time.Now()
	details := monitorTryRunDetails(checks)
	for key, value := range evaluator.SafeDetails(eval) {
		details[key] = value
	}
	message := "prometheus query evaluated"
	if evalErr != nil {
		message = "prometheus query evaluated with invalid rule settings"
	}
	log, err := store.AddMonitorAlertEvalLog(model.MonitorAlertEvalLog{
		RuleID: rule.ID, RuleVersion: rule.Version, Status: status, Message: message,
		StartedAt: started, FinishedAt: finished, DurationMs: finished.Sub(started).Milliseconds(),
		DatasourceID: rule.DatasourceID, QueryHash: eval.QueryHash, Details: details,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alert rule evaluation log write failed"})
		return
	}
	c.JSON(http.StatusOK, model.MonitorAlertTryRunResult{OK: ok, Status: status, Checks: checks, Rule: safeMonitorTryRunRule(rule), EvalLog: log, Eval: eval})
}

func tryRunEvalStatus(eval evaluator.Result, evalErr error) (string, bool) {
	if evalErr != nil {
		return "invalid", false
	}
	if eval.State == "invalid_response" {
		return "invalid_response", false
	}
	return "valid", true
}

func evaluatorRule(rule *model.MonitorAlertRule, queryHash string) evaluator.Rule {
	return evaluator.Rule{
		ID: rule.ID, Name: rule.Name, Severity: rule.Severity, DatasourceID: rule.DatasourceID,
		QueryHash: queryHash, ForDuration: rule.ForDuration, NoDataPolicy: rule.NoDataPolicy,
		Labels: rule.Labels, Annotations: rule.Annotations,
	}
}

func safeMonitorTryRunRule(rule *model.MonitorAlertRule) *model.MonitorAlertRule {
	if rule == nil {
		return nil
	}
	cp := *rule
	cp.Query = ""
	return &cp
}
