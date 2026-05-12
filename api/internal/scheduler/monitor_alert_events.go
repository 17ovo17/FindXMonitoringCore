package scheduler

import (
	"errors"
	"net/url"
	"strings"

	"ai-workbench-api/internal/evaluator"
	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/notifier"
	"ai-workbench-api/internal/store"
)

const monitorAlertRedactedValue = "<REDACTED>"

// dispatchAlertEvent is replaced in tests to observe firing notifications
// without touching the real notifier package.
var dispatchAlertEvent = notifier.DispatchAlertEvent

func applyMonitorAlertEvaluation(rule model.MonitorAlertRule, eval evaluator.Result) (int, int, error) {
	if eval.Triggered && len(eval.Candidates) > 0 {
		created, active, err := upsertMonitorAlertCandidates(rule, eval.Candidates)
		resolved, resolveErr := resolveStaleMonitorAlertEvents(rule.ID, active)
		return created, resolved, errors.Join(err, resolveErr)
	}
	if eval.State == "no_data_keep_state" {
		return 0, 0, nil
	}
	resolved, err := resolveAllMonitorAlertEvents(rule.ID)
	return 0, resolved, err
}

func upsertMonitorAlertCandidates(rule model.MonitorAlertRule, candidates []evaluator.Candidate) (int, map[string]bool, error) {
	created := 0
	active := map[string]bool{}
	var joined error
	for _, candidate := range candidates {
		event := eventFromMonitorCandidate(rule, candidate)
		fingerprint := model.GenerateMonitorAlertEventFingerprint(event)
		active[fingerprint] = true
		newlyFiring := !currentMonitorAlertFingerprintExists(fingerprint)
		if newlyFiring {
			created++
		}
		stored, err := store.UpsertMonitorAlertEvent(event)
		joined = errors.Join(joined, err)
		if err == nil && newlyFiring {
			dispatchAlertEvent(stored)
		}
	}
	return created, active, joined
}

func eventFromMonitorCandidate(rule model.MonitorAlertRule, candidate evaluator.Candidate) *model.MonitorAlertEvent {
	labels := sanitizeMonitorAlertStringMap(mergeMonitorStringMaps(rule.Labels, candidate.Labels))
	annotations := sanitizeMonitorAlertStringMap(mergeMonitorStringMaps(rule.Annotations, map[string]string{"reason": candidate.Reason}))
	return &model.MonitorAlertEvent{
		RuleID: rule.ID, RuleVersion: rule.Version, EventKey: candidate.EventKey, Name: rule.Name,
		Severity: rule.Severity, Status: model.MonitorAlertEventStatusFiring, DatasourceID: rule.DatasourceID,
		TargetIdent: monitorCandidateTargetIdent(labels), Labels: labels,
		Annotations: annotations,
		Value:       candidate.Value, Count: 1,
	}
}

func resolveStaleMonitorAlertEvents(ruleID string, active map[string]bool) (int, error) {
	resolved := 0
	var joined error
	for _, event := range store.ListMonitorAlertEvents(true) {
		if event.RuleID != ruleID || active[monitorAlertEventFingerprint(event)] {
			continue
		}
		if resolveMonitorAlertEvent(event.ID) {
			resolved++
		} else {
			joined = errors.Join(joined, errors.New("resolve alert event failed"))
		}
	}
	return resolved, joined
}

func resolveAllMonitorAlertEvents(ruleID string) (int, error) {
	return resolveStaleMonitorAlertEvents(ruleID, map[string]bool{})
}

func monitorAlertEventFingerprint(event model.MonitorAlertEvent) string {
	if event.Fingerprint != "" {
		return event.Fingerprint
	}
	return model.GenerateMonitorAlertEventFingerprint(&event)
}

func resolveMonitorAlertEvent(eventID string) bool {
	_, ok, err := store.ApplyMonitorAlertEventAction(eventID, model.MonitorAlertAction{
		Action: model.MonitorAlertEventActionResolve, Actor: monitorAlertSchedulerActor,
		Reason: "scheduler recovered", TraceID: monitorAlertSchedulerActor,
	})
	return ok && err == nil
}

func currentMonitorAlertFingerprintExists(fingerprint string) bool {
	for _, event := range store.ListMonitorAlertEvents(true) {
		if event.Fingerprint == fingerprint {
			return true
		}
	}
	return false
}

func monitorCandidateTargetIdent(labels map[string]string) string {
	for _, key := range []string{"instance", "target", "host", "hostname", "ip"} {
		if value := labels[key]; value != "" {
			return value
		}
	}
	return ""
}

func mergeMonitorStringMaps(first, second map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range first {
		out[key] = value
	}
	for key, value := range second {
		out[key] = value
	}
	return out
}

func sanitizeMonitorAlertStringMap(in map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range in {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if monitorAlertSensitiveKey(key) || monitorAlertSensitiveValue(value) {
			out[key] = monitorAlertRedactedValue
			continue
		}
		out[key] = strings.TrimSpace(value)
	}
	return out
}

func monitorAlertSensitiveKey(key string) bool {
	return monitorAlertContainsSensitiveMarker(key)
}

func monitorAlertSensitiveValue(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	lower := strings.ToLower(value)
	if monitorAlertContainsSensitiveMarker(lower) || strings.Contains(lower, "mysql://") || strings.Contains(lower, "@tcp(") {
		return true
	}
	return monitorAlertURLHasSensitivePart(value)
}

func monitorAlertContainsSensitiveMarker(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, marker := range []string{"api_key", "apikey", "authorization", "password", "private", "secret", "token", "cookie", "dsn", "auth"} {
		if strings.Contains(value, marker) {
			return true
		}
	}
	return false
}

func monitorAlertURLHasSensitivePart(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" {
		return false
	}
	if parsed.User != nil {
		return true
	}
	for key, values := range parsed.Query() {
		if monitorAlertContainsSensitiveMarker(key) {
			return true
		}
		for _, item := range values {
			if monitorAlertContainsSensitiveMarker(item) {
				return true
			}
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
