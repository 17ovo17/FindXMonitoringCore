package handler

import (
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
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
	checks := validateMonitorAlertRule(rule)
	ok = monitorChecksOK(checks)
	status := "valid"
	if !ok {
		status = "invalid"
	}
	finished := time.Now()
	details := monitorTryRunDetails(checks)
	details["mode"] = "dry_validation"
	details["promql_executed"] = false
	log, err := store.AddMonitorAlertEvalLog(model.MonitorAlertEvalLog{
		RuleID: rule.ID, RuleVersion: rule.Version, Status: status,
		Message: "dry_validation completed; PromQL was not executed", StartedAt: started, FinishedAt: finished,
		DurationMs: finished.Sub(started).Milliseconds(), DatasourceID: rule.DatasourceID,
		QueryHash: storeMonitorQueryHash(rule.Query), Details: details,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alert rule evaluation log write failed"})
		return
	}
	c.JSON(http.StatusOK, model.MonitorAlertTryRunResult{OK: ok, Status: status, Checks: checks, Rule: rule, EvalLog: log})
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
	if err := c.ShouldBindJSON(&payload); err == nil && strings.TrimSpace(payload.Query) != "" {
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

func filterMonitorRules(rules []model.MonitorAlertRule, status string) []model.MonitorAlertRule {
	out := []model.MonitorAlertRule{}
	for _, rule := range rules {
		if rule.Status == status {
			out = append(out, rule)
		}
	}
	return out
}
