package evaluator

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"ai-workbench-api/internal/monitoring"
)

const (
	NoDataKeepState = "keep_state"
	NoDataAlerting  = "alerting"
	NoDataOK        = "ok"
	MaxSamples      = 20
	MaxCandidates   = 20
)

type Rule struct {
	ID           string
	Name         string
	Severity     string
	DatasourceID string
	QueryHash    string
	ForDuration  string
	NoDataPolicy string
	Labels       map[string]string
	Annotations  map[string]string
}

type Result struct {
	PromQLExecuted bool                   `json:"promql_executed"`
	State          string                 `json:"state"`
	Triggered      bool                   `json:"triggered"`
	NoDataPolicy   string                 `json:"no_data_policy"`
	ForDurationMS  int64                  `json:"for_duration_ms"`
	QueryHash      string                 `json:"query_hash,omitempty"`
	DatasourceID   string                 `json:"datasource_id,omitempty"`
	LatencyMS      int64                  `json:"latency_ms"`
	Warnings       []string               `json:"warnings,omitempty"`
	Stats          monitoring.ResultStats `json:"stats"`
	Samples        []Sample               `json:"samples,omitempty"`
	Candidates     []Candidate            `json:"candidates,omitempty"`
	Message        string                 `json:"message,omitempty"`
}

type Sample struct {
	Labels map[string]string `json:"labels,omitempty"`
	Value  string            `json:"value"`
}

type Candidate struct {
	EventKey string            `json:"event_key"`
	Labels   map[string]string `json:"labels,omitempty"`
	Value    string            `json:"value"`
	Reason   string            `json:"reason"`
}

func EvaluateRule(rule Rule, prom monitoring.PrometheusCallResult) (Result, error) {
	durationMS, err := ParsePromDurationMillis(rule.ForDuration)
	if err != nil {
		return baseResult(rule, prom, durationMS), err
	}
	result := baseResult(rule, prom, durationMS)
	resultType := result.Stats.ResultType
	switch resultType {
	case "vector":
		evaluateVector(&result, prom.Data["result"])
	case "scalar", "string":
		evaluateScalarOrString(&result, prom.Data["result"])
	case "matrix":
		result.State = "invalid_response"
		result.Message = "matrix result is unsupported for instant tryrun"
	default:
		result.State = "invalid_response"
		result.Message = "unsupported prometheus result type"
	}
	return result, nil
}

func ParsePromDurationMillis(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	total := int64(0)
	for pos := 0; pos < len(raw); {
		start := pos
		for pos < len(raw) && raw[pos] >= '0' && raw[pos] <= '9' {
			pos++
		}
		if start == pos || pos >= len(raw) {
			return 0, fmt.Errorf("invalid for_duration")
		}
		value, err := strconv.ParseInt(raw[start:pos], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid for_duration")
		}
		unitStart := pos
		for pos < len(raw) && (raw[pos] < '0' || raw[pos] > '9') {
			pos++
		}
		multiplier, ok := durationUnitMillis(raw[unitStart:pos])
		if !ok {
			return 0, fmt.Errorf("invalid for_duration")
		}
		total += value * multiplier
	}
	return total, nil
}

func SafeDetails(result Result) map[string]any {
	return map[string]any{
		"promql_executed": true,
		"query_hash":      result.QueryHash,
		"datasource_id":   result.DatasourceID,
		"state":           result.State,
		"triggered":       result.Triggered,
		"no_data_policy":  result.NoDataPolicy,
		"for_duration_ms": result.ForDurationMS,
		"stats":           result.Stats,
		"candidate_count": len(result.Candidates),
		"sample_count":    len(result.Samples),
	}
}

func SanitizeLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(labels))
	for key, value := range labels {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if sensitiveKey(key) {
			out[key] = "<REDACTED>"
		} else {
			out[key] = strings.TrimSpace(value)
		}
	}
	return out
}

func baseResult(rule Rule, prom monitoring.PrometheusCallResult, durationMS int64) Result {
	policy := strings.TrimSpace(rule.NoDataPolicy)
	if policy == "" {
		policy = NoDataKeepState
	}
	return Result{
		PromQLExecuted: true, State: "ok", NoDataPolicy: policy, ForDurationMS: durationMS,
		QueryHash: rule.QueryHash, DatasourceID: rule.DatasourceID, LatencyMS: prom.LatencyMS,
		Warnings: append([]string{}, prom.Warnings...), Stats: prom.Stats,
	}
}

func evaluateVector(result *Result, raw any) {
	rows, _ := raw.([]any)
	if len(rows) == 0 {
		applyNoDataPolicy(result)
		return
	}
	result.State = "firing"
	result.Triggered = true
	for _, item := range rows {
		row, _ := item.(map[string]any)
		labels := labelsFromMetric(row["metric"])
		value := valueString(row["value"])
		appendSample(result, labels, value)
		appendCandidate(result, labels, value, "vector sample present")
	}
}

func evaluateScalarOrString(result *Result, raw any) {
	value := scalarStringValue(raw)
	appendSample(result, map[string]string{}, value)
	result.State = "ok"
	result.Triggered = false
}

func applyNoDataPolicy(result *Result) {
	switch result.NoDataPolicy {
	case NoDataAlerting:
		result.State = "no_data_alerting"
		result.Triggered = true
		appendCandidate(result, map[string]string{"no_data": "true"}, "no_data", "no data policy alerting")
	case NoDataOK:
		result.State = "no_data_ok"
	default:
		result.State = "no_data_keep_state"
	}
}

func appendSample(result *Result, labels map[string]string, value string) {
	if len(result.Samples) >= MaxSamples {
		return
	}
	result.Samples = append(result.Samples, Sample{Labels: SanitizeLabels(labels), Value: value})
}

func appendCandidate(result *Result, labels map[string]string, value, reason string) {
	if len(result.Candidates) >= MaxCandidates {
		return
	}
	safeLabels := SanitizeLabels(labels)
	result.Candidates = append(result.Candidates, Candidate{
		EventKey: candidateKey(safeLabels), Labels: safeLabels, Value: value, Reason: reason,
	})
}

func labelsFromMetric(raw any) map[string]string {
	metric, _ := raw.(map[string]any)
	labels := make(map[string]string, len(metric))
	for key, value := range metric {
		labels[key] = fmt.Sprint(value)
	}
	return labels
}

func valueString(raw any) string {
	values, ok := raw.([]any)
	if !ok || len(values) < 2 {
		return ""
	}
	return fmt.Sprint(values[1])
}

func scalarStringValue(raw any) string {
	values, ok := raw.([]any)
	if !ok || len(values) < 2 {
		return fmt.Sprint(raw)
	}
	return fmt.Sprint(values[1])
}

func candidateKey(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+labels[key])
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return fmt.Sprintf("%x", sum[:])
}

func durationUnitMillis(unit string) (int64, bool) {
	switch unit {
	case "ms":
		return 1, true
	case "s":
		return int64(time.Second / time.Millisecond), true
	case "m":
		return int64(time.Minute / time.Millisecond), true
	case "h":
		return int64(time.Hour / time.Millisecond), true
	case "d":
		return int64(24 * time.Hour / time.Millisecond), true
	case "w":
		return int64(7 * 24 * time.Hour / time.Millisecond), true
	case "y":
		return int64(365 * 24 * time.Hour / time.Millisecond), true
	default:
		return 0, false
	}
}

func sensitiveKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, marker := range []string{"api_key", "apikey", "auth", "authorization", "cookie", "dsn", "password", "private", "secret", "token"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}
