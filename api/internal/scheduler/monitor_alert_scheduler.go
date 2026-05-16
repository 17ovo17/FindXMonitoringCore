package scheduler

import (
	"context"
	"sync"
	"time"

	"ai-workbench-api/internal/evaluator"
	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/monitoring"
	"ai-workbench-api/internal/store"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const monitorAlertSchedulerActor = "monitor-alert-scheduler"

type MonitorAlertSchedulerOptions struct {
	Rules                    []model.MonitorAlertRule
	Datasources              []monitoring.Datasource
	DefaultDatasourceID      string
	FallbackPrometheusURL    string
	Gateway                  *monitoring.PrometheusGateway
	Timeout                  time.Duration
	MaxConcurrency           int
	UseConfiguredRules       bool
	UseConfiguredDatasources bool
}

type MonitorAlertRunResult struct {
	Evaluated      int
	Succeeded      int
	Failed         int
	EventsCreated  int
	EventsResolved int
	Skipped        int
	Locked         bool
	EvalLogs       []model.MonitorAlertEvalLog
}

type monitorRuleOutcome struct {
	Evaluated      int
	Succeeded      int
	Failed         int
	EventsCreated  int
	EventsResolved int
	EvalLog        model.MonitorAlertEvalLog
}

var monitorAlertRunMu sync.Mutex
var addMonitorAlertEvalLog = store.AddMonitorAlertEvalLog

// globalWorkerManager 持有全局 worker 管理器实例，供 Stop 时使用。
var globalWorkerManager *AlertWorkerManager

func StartMonitorAlertScheduler() {
	if !viper.GetBool("monitoring.alert_scheduler.enabled") {
		log.Info("scheduler: monitor alert scheduler disabled")
		return
	}
	manager := newAlertWorkerManager()
	globalWorkerManager = manager
	manager.Start()
	log.Info("scheduler: monitor alert scheduler started (per-rule worker mode)")
}

// StopMonitorAlertScheduler 停止所有告警规则 worker。
func StopMonitorAlertScheduler() {
	if globalWorkerManager != nil {
		globalWorkerManager.Stop()
	}
}

func RunMonitorAlertEvaluationOnce(ctx context.Context, opts MonitorAlertSchedulerOptions) MonitorAlertRunResult {
	if !monitorAlertRunMu.TryLock() {
		return MonitorAlertRunResult{Skipped: 1, Locked: true}
	}
	defer monitorAlertRunMu.Unlock()
	opts = normalizeMonitorAlertOptions(opts)
	result := MonitorAlertRunResult{}
	rules := monitorAlertRulesForRun(opts)
	active := activeMonitorAlertRules(rules, &result)
	outcomes := runMonitorAlertRules(ctx, active, opts)
	for outcome := range outcomes {
		result.Evaluated += outcome.Evaluated
		result.Succeeded += outcome.Succeeded
		result.Failed += outcome.Failed
		result.EventsCreated += outcome.EventsCreated
		result.EventsResolved += outcome.EventsResolved
		if outcome.EvalLog.ID != "" {
			result.EvalLogs = append(result.EvalLogs, outcome.EvalLog)
		}
	}
	return result
}

func monitorAlertSchedulerOptionsFromConfig() MonitorAlertSchedulerOptions {
	return MonitorAlertSchedulerOptions{
		DefaultDatasourceID:      monitoring.DefaultPrometheusDatasourceID,
		FallbackPrometheusURL:    viper.GetString("prometheus.url"),
		Timeout:                  timeoutSecondsFromConfig(),
		MaxConcurrency:           boundedConfigInt("monitoring.alert_scheduler.max_concurrency", 4, 1, 16),
		UseConfiguredRules:       true,
		UseConfiguredDatasources: true,
	}
}

func schedulerIntervalFromConfig() time.Duration {
	seconds := boundedConfigInt("monitoring.alert_scheduler.interval_seconds", 60, 10, 86400)
	return time.Duration(seconds) * time.Second
}

func timeoutSecondsFromConfig() time.Duration {
	seconds := boundedConfigInt("monitoring.alert_scheduler.timeout_seconds", 5, 1, 30)
	return time.Duration(seconds) * time.Second
}

func boundedConfigInt(key string, fallback, min, max int) int {
	value := viper.GetInt(key)
	if value == 0 {
		value = fallback
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func runMonitorAlertSchedulerLoop(interval time.Duration, opts MonitorAlertSchedulerOptions) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		result := RunMonitorAlertEvaluationOnce(context.Background(), opts)
		log.Infof("scheduler: monitor alert run evaluated=%d succeeded=%d failed=%d created=%d resolved=%d skipped=%d locked=%v",
			result.Evaluated, result.Succeeded, result.Failed, result.EventsCreated, result.EventsResolved, result.Skipped, result.Locked)
	}
}

func normalizeMonitorAlertOptions(opts MonitorAlertSchedulerOptions) MonitorAlertSchedulerOptions {
	if opts.DefaultDatasourceID == "" {
		opts.DefaultDatasourceID = monitoring.DefaultPrometheusDatasourceID
	}
	if opts.UseConfiguredDatasources || opts.Datasources == nil {
		opts.Datasources = monitoring.PrometheusDatasourcesFromConfig()
		opts.FallbackPrometheusURL = viper.GetString("prometheus.url")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = monitoring.DefaultInstantTimeout
	}
	opts.Timeout = monitoring.TimeoutOrDefault(opts.Timeout, monitoring.DefaultInstantTimeout)
	if opts.MaxConcurrency <= 0 {
		opts.MaxConcurrency = 4
	}
	if opts.MaxConcurrency > 16 {
		opts.MaxConcurrency = 16
	}
	if opts.Gateway == nil {
		opts.Gateway = monitoring.NewPrometheusGateway(nil)
	}
	return opts
}

func monitorAlertRulesForRun(opts MonitorAlertSchedulerOptions) []model.MonitorAlertRule {
	if opts.UseConfiguredRules || opts.Rules == nil {
		return store.ListMonitorAlertRules()
	}
	return append([]model.MonitorAlertRule{}, opts.Rules...)
}

func activeMonitorAlertRules(rules []model.MonitorAlertRule, result *MonitorAlertRunResult) []model.MonitorAlertRule {
	active := make([]model.MonitorAlertRule, 0, len(rules))
	for _, rule := range rules {
		if rule.Enabled && rule.Status == model.MonitorAlertRuleStatusActive {
			active = append(active, rule)
			continue
		}
		result.Skipped++
	}
	return active
}

func runMonitorAlertRules(ctx context.Context, rules []model.MonitorAlertRule, opts MonitorAlertSchedulerOptions) <-chan monitorRuleOutcome {
	out := make(chan monitorRuleOutcome, len(rules))
	sem := make(chan struct{}, opts.MaxConcurrency)
	var wg sync.WaitGroup
	for _, rule := range rules {
		rule := rule
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out <- evaluateMonitorAlertRule(ctx, rule, opts)
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func evaluateMonitorAlertRule(ctx context.Context, rule model.MonitorAlertRule, opts MonitorAlertSchedulerOptions) monitorRuleOutcome {
	started := time.Now()
	queryHash := monitoring.QueryHash(rule.Query)
	if err := monitoring.ValidatePromQL(rule.Query); err != nil {
		return failedOutcome(writeMonitorAlertEvalLog(rule, started, "invalid", "validation failed", invalidDetails(rule, queryHash)))
	}
	dsID, base, err := monitoring.ResolvePrometheusDatasource(opts.Datasources, rule.DatasourceID, opts.DefaultDatasourceID, opts.FallbackPrometheusURL)
	if err != nil {
		return failedOutcome(writeMonitorAlertEvalLog(rule, started, "datasource_not_found", "prometheus datasource not found", datasourceDetails(rule, queryHash)))
	}
	rule.DatasourceID = dsID
	prom, err := queryMonitorAlertPrometheus(ctx, opts, base, rule.Query)
	if err != nil {
		return failedOutcome(writeMonitorAlertEvalLog(rule, started, "upstream_error", "prometheus query failed", upstreamDetails(rule, prom, queryHash, err)))
	}
	eval, evalErr := evaluator.EvaluateRule(evaluatorRuleFromMonitorRule(rule, prom.QueryHash), prom)
	if evalErr != nil || eval.State == "invalid_response" {
		return failedOutcome(writeMonitorAlertEvalLog(rule, started, evalFailureStatus(eval, evalErr), "alert rule evaluation invalid", evaluator.SafeDetails(eval)))
	}
	return successfulEvaluationOutcome(rule, eval, started)
}

func queryMonitorAlertPrometheus(ctx context.Context, opts MonitorAlertSchedulerOptions, base, query string) (monitoring.PrometheusCallResult, error) {
	queryCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	return opts.Gateway.QueryInstant(queryCtx, monitoring.PrometheusQueryRequest{
		BaseURL: base, Query: query, Timeout: opts.Timeout,
	})
}

func successfulEvaluationOutcome(rule model.MonitorAlertRule, eval evaluator.Result, started time.Time) monitorRuleOutcome {
	details := evaluator.SafeDetails(eval)
	logItem := writeMonitorAlertEvalLog(rule, started, eval.State, "alert rule evaluated", details)
	out := monitorRuleOutcome{Evaluated: 1, EvalLog: logItem}
	if logItem.ID == "" {
		out.Failed = 1
		return out
	}
	created, resolved, eventErr := applyMonitorAlertEvaluation(rule, eval)
	out.EventsCreated = created
	out.EventsResolved = resolved
	if eventErr != nil {
		out.Failed = 1
		return out
	}
	out.Succeeded = 1
	return out
}

func failedOutcome(logItem model.MonitorAlertEvalLog) monitorRuleOutcome {
	return monitorRuleOutcome{Evaluated: 1, Failed: 1, EvalLog: logItem}
}

func writeMonitorAlertEvalLog(rule model.MonitorAlertRule, started time.Time, status, message string, details map[string]any) model.MonitorAlertEvalLog {
	finished := time.Now()
	logItem, err := addMonitorAlertEvalLog(model.MonitorAlertEvalLog{
		RuleID: rule.ID, RuleVersion: rule.Version, Status: status, Message: message,
		Details: safeMonitorAlertDetails(details), StartedAt: started, FinishedAt: finished,
		DurationMs: finished.Sub(started).Milliseconds(), DatasourceID: rule.DatasourceID,
		QueryHash: monitoring.QueryHash(rule.Query),
	})
	if err != nil {
		log.WithError(err).Warnf("scheduler: monitor alert eval log failed rule_id=%s datasource_id=%s query_hash=%s status=%s",
			rule.ID, rule.DatasourceID, monitoring.QueryHash(rule.Query), status)
		return model.MonitorAlertEvalLog{}
	}
	return logItem
}

func invalidDetails(rule model.MonitorAlertRule, queryHash string) map[string]any {
	return map[string]any{"promql_executed": false, "query_hash": queryHash, "datasource_id": rule.DatasourceID, "state": "invalid_query"}
}

func datasourceDetails(rule model.MonitorAlertRule, queryHash string) map[string]any {
	return map[string]any{"promql_executed": false, "query_hash": queryHash, "datasource_id": rule.DatasourceID, "state": "datasource_not_found"}
}

func upstreamDetails(rule model.MonitorAlertRule, prom monitoring.PrometheusCallResult, queryHash string, err error) map[string]any {
	return map[string]any{
		"promql_executed": true, "query_hash": firstNonEmpty(prom.QueryHash, queryHash),
		"datasource_id": rule.DatasourceID, "state": "upstream_error",
		"upstream_status": monitoring.HTTPStatus(err), "latency_ms": prom.LatencyMS,
	}
}

func evalFailureStatus(eval evaluator.Result, evalErr error) string {
	if evalErr != nil {
		return "invalid"
	}
	return eval.State
}

func safeMonitorAlertDetails(details map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range details {
		if monitorAlertSensitiveKey(key) {
			out[key] = monitorAlertRedactedValue
			continue
		}
		out[key] = safeMonitorAlertDetailValue(value)
	}
	return out
}

func safeMonitorAlertDetailValue(value any) any {
	switch item := value.(type) {
	case string:
		if monitorAlertSensitiveValue(item) {
			return monitorAlertRedactedValue
		}
		return item
	case map[string]any:
		return safeMonitorAlertDetails(item)
	case map[string]string:
		return sanitizeMonitorAlertStringMap(item)
	case []any:
		out := make([]any, 0, len(item))
		for _, nested := range item {
			out = append(out, safeMonitorAlertDetailValue(nested))
		}
		return out
	case []string:
		out := make([]string, 0, len(item))
		for _, nested := range item {
			out = append(out, safeMonitorAlertDetailValue(nested).(string))
		}
		return out
	default:
		return item
	}
}

func evaluatorRuleFromMonitorRule(rule model.MonitorAlertRule, queryHash string) evaluator.Rule {
	return evaluator.Rule{
		ID: rule.ID, Name: rule.Name, Severity: rule.Severity, DatasourceID: rule.DatasourceID,
		QueryHash: queryHash, ForDuration: rule.ForDuration, NoDataPolicy: rule.NoDataPolicy,
		Labels: rule.Labels, Annotations: rule.Annotations,
	}
}
