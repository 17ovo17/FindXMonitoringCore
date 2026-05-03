package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/monitoring"
	"ai-workbench-api/internal/store"

	"github.com/spf13/viper"
)

func TestSchedulerFiringCreatesCurrentAndSafeEvalLog(t *testing.T) {
	rule := schedulerTestRule("firing", model.MonitorNoDataPolicyOK)
	rule.Query = `up{secret_token="<TOKEN>"} == 0`
	upstream := prometheusTestServer(t, http.StatusOK, vectorPayload("host-a"))
	result := runSchedulerRule(t, upstream.URL, rule)

	if result.Evaluated != 1 || result.Succeeded != 1 || result.EventsCreated != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	current := currentEventsForRule(rule.ID)
	if len(current) != 1 || current[0].TargetIdent != "host-a" {
		t.Fatalf("expected one current event for host-a, got %+v", current)
	}
	rendered := renderEvalLog(t, result)
	if !strings.Contains(rendered, "query_hash") || strings.Contains(rendered, rule.Query) || strings.Contains(rendered, "<TOKEN>") {
		t.Fatalf("eval log leaked raw query or missed hash: %s", rendered)
	}
}

func TestSchedulerFiringRedactsSensitiveLabelValuesBeforePersist(t *testing.T) {
	rule := schedulerTestRule("redact-label-values", model.MonitorNoDataPolicyOK)
	rule.Labels = map[string]string{"service": "checkout"}
	rule.Annotations = map[string]string{"dsn_hint": "user:pass@tcp(db:3306)/prod"}
	upstream := prometheusTestServer(t, http.StatusOK, sensitiveVectorPayload())

	result := runSchedulerRule(t, upstream.URL, rule)
	if result.Evaluated != 1 || result.Succeeded != 1 || result.EventsCreated != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	current := currentEventsForRule(rule.ID)
	if len(current) != 1 || current[0].TargetIdent != monitorAlertRedactedValue {
		t.Fatalf("expected one redacted current event, got %+v", current)
	}
	renderedEvent := renderJSON(t, current[0])
	renderedLog := renderEvalLog(t, result)
	for _, rendered := range []string{renderedEvent, renderedLog} {
		assertNoSensitiveSchedulerFragment(t, rendered)
	}
}

func TestSchedulerConfigDefaultsAndBounds(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	if viper.GetBool("monitoring.alert_scheduler.enabled") {
		t.Fatalf("alert scheduler must be disabled by default")
	}
	defaults := monitorAlertSchedulerOptionsFromConfig()
	if defaults.Timeout != 5*time.Second || defaults.MaxConcurrency != 4 || schedulerIntervalFromConfig() != time.Minute {
		t.Fatalf("unexpected scheduler defaults: opts=%+v interval=%s", defaults, schedulerIntervalFromConfig())
	}
	viper.Set("monitoring.alert_scheduler.interval_seconds", 1)
	viper.Set("monitoring.alert_scheduler.timeout_seconds", 99)
	viper.Set("monitoring.alert_scheduler.max_concurrency", 0)
	if schedulerIntervalFromConfig() != 10*time.Second || timeoutSecondsFromConfig() != 30*time.Second {
		t.Fatalf("interval/timeout clamp failed: interval=%s timeout=%s", schedulerIntervalFromConfig(), timeoutSecondsFromConfig())
	}
	if got := monitorAlertSchedulerOptionsFromConfig().MaxConcurrency; got != 4 {
		t.Fatalf("zero max_concurrency should use fallback 4, got %d", got)
	}
	viper.Set("monitoring.alert_scheduler.max_concurrency", -1)
	if got := monitorAlertSchedulerOptionsFromConfig().MaxConcurrency; got != 1 {
		t.Fatalf("low max_concurrency clamp failed: %d", got)
	}
	viper.Set("monitoring.alert_scheduler.max_concurrency", 99)
	if got := monitorAlertSchedulerOptionsFromConfig().MaxConcurrency; got != 16 {
		t.Fatalf("high max_concurrency clamp failed: %d", got)
	}
}

func TestSchedulerRepeatedFiringDeduplicatesCurrentEvent(t *testing.T) {
	rule := schedulerTestRule("dedupe", model.MonitorNoDataPolicyOK)
	upstream := prometheusTestServer(t, http.StatusOK, vectorPayload("host-a"))

	first := runSchedulerRule(t, upstream.URL, rule)
	second := runSchedulerRule(t, upstream.URL, rule)
	current := currentEventsForRule(rule.ID)
	if first.EventsCreated != 1 || second.EventsCreated != 0 {
		t.Fatalf("expected first create and second merge, first=%+v second=%+v", first, second)
	}
	if len(current) != 1 || current[0].Count < 2 {
		t.Fatalf("expected one merged current event with count>=2, got %+v", current)
	}
}

func TestSchedulerRecoveryResolvesOnlySameRuleCurrentEvents(t *testing.T) {
	rule := schedulerTestRule("recover", model.MonitorNoDataPolicyOK)
	other := schedulerTestRule("recover-other", model.MonitorNoDataPolicyOK)
	createCurrentEvent(t, rule.ID, "stale-a")
	createCurrentEvent(t, other.ID, "other-a")
	upstream := prometheusTestServer(t, http.StatusOK, emptyVectorPayload())

	result := runSchedulerRule(t, upstream.URL, rule)
	if result.EventsResolved != 1 || len(currentEventsForRule(rule.ID)) != 0 {
		t.Fatalf("same rule current should be resolved, result=%+v current=%+v", result, currentEventsForRule(rule.ID))
	}
	if len(historyEventsForRule(rule.ID)) != 1 || len(currentEventsForRule(other.ID)) != 1 {
		t.Fatalf("history or other rule current mismatch, history=%+v other=%+v", historyEventsForRule(rule.ID), currentEventsForRule(other.ID))
	}
}

func TestSchedulerRecoveryUsesActiveFingerprint(t *testing.T) {
	rule := schedulerTestRule("recover-fingerprint", model.MonitorNoDataPolicyOK)
	upstream := prometheusTestServer(t, http.StatusOK, vectorPayload("host-a"))
	first := runSchedulerRule(t, upstream.URL, rule)
	active := currentEventsForRule(rule.ID)
	if first.EventsCreated != 1 || len(active) != 1 {
		t.Fatalf("expected initial active event, result=%+v current=%+v", first, active)
	}
	createCurrentEventWithLabels(t, rule.ID, active[0].EventKey, map[string]string{"instance": "host-b"})

	second := runSchedulerRule(t, upstream.URL, rule)
	current := currentEventsForRule(rule.ID)
	history := historyEventsForRule(rule.ID)
	if second.EventsResolved != 1 || len(current) != 1 || len(history) != 1 {
		t.Fatalf("expected stale fingerprint recovered, result=%+v current=%+v history=%+v", second, current, history)
	}
	if current[0].TargetIdent != "host-a" || history[0].TargetIdent != "host-b" {
		t.Fatalf("unexpected active/history targets, current=%+v history=%+v", current, history)
	}
}

func TestSchedulerKeepStateDoesNotRecoverExistingCurrent(t *testing.T) {
	rule := schedulerTestRule("keep-state", model.MonitorNoDataPolicyKeepState)
	createCurrentEvent(t, rule.ID, "stale-a")
	upstream := prometheusTestServer(t, http.StatusOK, emptyVectorPayload())

	result := runSchedulerRule(t, upstream.URL, rule)
	if result.EventsResolved != 0 || len(currentEventsForRule(rule.ID)) != 1 {
		t.Fatalf("keep_state should keep existing current, result=%+v current=%+v", result, currentEventsForRule(rule.ID))
	}
}

func TestSchedulerNoDataAlertingCreatesNoDataEvent(t *testing.T) {
	rule := schedulerTestRule("no-data-alerting", model.MonitorNoDataPolicyAlerting)
	upstream := prometheusTestServer(t, http.StatusOK, emptyVectorPayload())

	result := runSchedulerRule(t, upstream.URL, rule)
	current := currentEventsForRule(rule.ID)
	if result.EventsCreated != 1 || len(current) != 1 || current[0].Value != "no_data" {
		t.Fatalf("expected no_data current event, result=%+v current=%+v", result, current)
	}
}

func TestSchedulerEvalLogFailureDoesNotWriteEvents(t *testing.T) {
	original := addMonitorAlertEvalLog
	addMonitorAlertEvalLog = func(model.MonitorAlertEvalLog) (model.MonitorAlertEvalLog, error) {
		return model.MonitorAlertEvalLog{}, errors.New("forced eval log failure")
	}
	t.Cleanup(func() { addMonitorAlertEvalLog = original })
	rule := schedulerTestRule("eval-log-fail", model.MonitorNoDataPolicyOK)
	upstream := prometheusTestServer(t, http.StatusOK, vectorPayload("host-a"))

	result := runSchedulerRule(t, upstream.URL, rule)
	if result.Failed != 1 || result.Succeeded != 0 || result.EventsCreated != 0 || len(result.EvalLogs) != 0 {
		t.Fatalf("eval log failure should fail before events, result=%+v", result)
	}
	if len(currentEventsForRule(rule.ID)) != 0 || len(historyEventsForRule(rule.ID)) != 0 {
		t.Fatalf("eval log failure must not write events, current=%+v history=%+v", currentEventsForRule(rule.ID), historyEventsForRule(rule.ID))
	}
}

func TestSchedulerUpstreamFailuresOnlyWriteEvalLog(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"http-502", http.StatusBadGateway, `authorization token secret`},
		{"status-error", http.StatusOK, `{"status":"error","error":"token secret"}`},
		{"invalid-json", http.StatusOK, `not-json token secret`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rule := schedulerTestRule("upstream-"+tc.name, model.MonitorNoDataPolicyOK)
			upstream := prometheusTestServer(t, tc.status, tc.body)
			result := runSchedulerRule(t, upstream.URL, rule)
			if result.Failed != 1 || len(currentEventsForRule(rule.ID)) != 0 || len(historyEventsForRule(rule.ID)) != 0 {
				t.Fatalf("upstream failure must not write events: result=%+v current=%+v history=%+v", result, currentEventsForRule(rule.ID), historyEventsForRule(rule.ID))
			}
			rendered := strings.ToLower(renderEvalLog(t, result))
			if strings.Contains(rendered, "token secret") || strings.Contains(rendered, "authorization") {
				t.Fatalf("eval log leaked upstream body: %s", rendered)
			}
		})
	}
}

func TestSchedulerDisabledRuleIsSkipped(t *testing.T) {
	requests := 0
	rule := schedulerTestRule("disabled", model.MonitorNoDataPolicyOK)
	rule.Enabled = false
	rule.Status = model.MonitorAlertRuleStatusDisabled
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(vectorPayload("host-a")))
	}))
	defer upstream.Close()

	result := runSchedulerRule(t, upstream.URL, rule)
	if result.Evaluated != 0 || result.Skipped != 1 || requests != 0 {
		t.Fatalf("disabled rule should be skipped, result=%+v requests=%d", result, requests)
	}
}

func TestSchedulerDatasourceNotFoundDoesNotCreateEvents(t *testing.T) {
	rule := schedulerTestRule("missing-ds", model.MonitorNoDataPolicyOK)
	rule.DatasourceID = "missing-prometheus"
	result := RunMonitorAlertEvaluationOnce(context.Background(), MonitorAlertSchedulerOptions{
		Rules: []model.MonitorAlertRule{rule}, Datasources: []monitoring.Datasource{},
		DefaultDatasourceID: monitoring.DefaultPrometheusDatasourceID, Timeout: time.Second, MaxConcurrency: 1,
	})
	if result.Failed != 1 || len(currentEventsForRule(rule.ID)) != 0 {
		t.Fatalf("missing datasource should fail without events, result=%+v", result)
	}
	if len(result.EvalLogs) != 1 || result.EvalLogs[0].Status != "datasource_not_found" {
		t.Fatalf("expected datasource_not_found eval log, logs=%+v", result.EvalLogs)
	}
}

func TestSchedulerConcurrentRunOnceIsLocked(t *testing.T) {
	rule := schedulerTestRule("locked", model.MonitorNoDataPolicyOK)
	requested := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(requested) })
		<-release
		_, _ = w.Write([]byte(vectorPayload("host-a")))
	}))
	defer upstream.Close()

	done := make(chan MonitorAlertRunResult, 1)
	go func() { done <- runSchedulerRule(t, upstream.URL, rule) }()
	<-requested
	locked := runSchedulerRule(t, upstream.URL, rule)
	close(release)
	first := <-done
	if !locked.Locked || locked.Skipped != 1 || first.Evaluated != 1 {
		t.Fatalf("expected second run locked and first evaluated, first=%+v locked=%+v", first, locked)
	}
}

func runSchedulerRule(t *testing.T, upstream string, rule model.MonitorAlertRule) MonitorAlertRunResult {
	t.Helper()
	return RunMonitorAlertEvaluationOnce(context.Background(), MonitorAlertSchedulerOptions{
		Rules: []model.MonitorAlertRule{rule},
		Datasources: []monitoring.Datasource{{
			ID: monitoring.DefaultPrometheusDatasourceID, Type: "prometheus", URL: upstream,
		}},
		DefaultDatasourceID: monitoring.DefaultPrometheusDatasourceID,
		Timeout:             2 * time.Second,
		MaxConcurrency:      2,
	})
}

func schedulerTestRule(name, policy string) model.MonitorAlertRule {
	id := "rule-scheduler-" + name + "-" + time.Now().Format("150405.000000000")
	return model.MonitorAlertRule{
		ID: id, Name: "Scheduler " + name, Query: "up == 0", Severity: model.MonitorAlertSeverityWarning,
		DatasourceID: monitoring.DefaultPrometheusDatasourceID, Enabled: true, Version: 1,
		NoDataPolicy: policy, Status: model.MonitorAlertRuleStatusActive,
	}
}

func prometheusTestServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query" {
			t.Fatalf("unexpected prometheus path: %s", r.URL.Path)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func createCurrentEvent(t *testing.T, ruleID, eventKey string) {
	t.Helper()
	createCurrentEventWithLabels(t, ruleID, eventKey, map[string]string{"instance": eventKey})
}

func createCurrentEventWithLabels(t *testing.T, ruleID, eventKey string, labels map[string]string) {
	t.Helper()
	_, err := store.UpsertMonitorAlertEvent(&model.MonitorAlertEvent{
		ID: ruleID + ":" + eventKey, RuleID: ruleID, RuleVersion: 1, EventKey: eventKey, Name: eventKey,
		Severity: model.MonitorAlertSeverityWarning, DatasourceID: monitoring.DefaultPrometheusDatasourceID,
		TargetIdent: labels["instance"], Labels: labels, Count: 1,
	})
	if err != nil {
		t.Fatalf("create current event failed: %v", err)
	}
}

func currentEventsForRule(ruleID string) []model.MonitorAlertEvent {
	return eventsForRule(ruleID, true)
}

func historyEventsForRule(ruleID string) []model.MonitorAlertEvent {
	return eventsForRule(ruleID, false)
}

func eventsForRule(ruleID string, current bool) []model.MonitorAlertEvent {
	out := []model.MonitorAlertEvent{}
	for _, event := range store.ListMonitorAlertEvents(current) {
		if event.RuleID == ruleID {
			out = append(out, event)
		}
	}
	return out
}

func renderEvalLog(t *testing.T, result MonitorAlertRunResult) string {
	t.Helper()
	if len(result.EvalLogs) == 0 {
		t.Fatalf("expected eval log in result: %+v", result)
	}
	data, err := json.Marshal(result.EvalLogs[0])
	if err != nil {
		t.Fatalf("marshal eval log: %v", err)
	}
	return string(data)
}

func renderJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(data)
}

func assertNoSensitiveSchedulerFragment(t *testing.T, rendered string) {
	t.Helper()
	lower := strings.ToLower(rendered)
	for _, forbidden := range []string{
		"http://user:pass@example:9090/metrics?token=abc", "pass", "token=abc",
		"mysql://", "@tcp", "cookie", "secret", "authorization",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("scheduler output leaked sensitive fragment %q in %s", forbidden, rendered)
		}
	}
}

func vectorPayload(instance string) string {
	return `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"instance":"` + instance + `","token":"<TOKEN>"},"value":[1,"1"]}]}}`
}

func sensitiveVectorPayload() string {
	return `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"instance":"http://user:pass@example:9090/metrics?token=abc","endpoint":"mysql://user:pass@db/prod","token":"<TOKEN>"},"value":[1,"1"]}]}}`
}

func emptyVectorPayload() string {
	return `{"status":"success","data":{"resultType":"vector","result":[]}}`
}
