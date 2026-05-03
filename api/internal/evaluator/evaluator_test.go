package evaluator

import (
	"testing"

	"ai-workbench-api/internal/monitoring"
)

func TestEvaluateRuleVectorFiringCandidates(t *testing.T) {
	result, err := EvaluateRule(testRule(NoDataKeepState), monitoring.PrometheusCallResult{
		Data: map[string]any{"resultType": "vector", "result": []any{
			map[string]any{"metric": map[string]any{"instance": "host-a", "job": "node"}, "value": []any{float64(1), "1"}},
		}},
		Stats: monitoring.ResultStats{ResultType: "vector", SeriesCount: 1, SampleCount: 1},
	})
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}
	if result.State != "firing" || !result.Triggered || len(result.Candidates) != 1 {
		t.Fatalf("unexpected firing result: %+v", result)
	}
}

func TestEvaluateRuleEmptyVectorNoDataPolicies(t *testing.T) {
	cases := []struct {
		policy     string
		wantState  string
		candidates int
	}{
		{NoDataKeepState, "no_data_keep_state", 0},
		{NoDataAlerting, "no_data_alerting", 1},
		{NoDataOK, "no_data_ok", 0},
	}
	for _, tc := range cases {
		result, err := EvaluateRule(testRule(tc.policy), emptyVector())
		if err != nil {
			t.Fatalf("evaluate %s failed: %v", tc.policy, err)
		}
		if result.State != tc.wantState || len(result.Candidates) != tc.candidates {
			t.Fatalf("policy %s got state=%s candidates=%d", tc.policy, result.State, len(result.Candidates))
		}
	}
}

func TestEvaluateRuleSensitiveLabelsAreRedacted(t *testing.T) {
	result, err := EvaluateRule(testRule(NoDataKeepState), monitoring.PrometheusCallResult{
		Data: map[string]any{"resultType": "vector", "result": []any{
			map[string]any{"metric": map[string]any{"token": "<TOKEN>", "instance": "host-a"}, "value": []any{float64(1), "1"}},
		}},
		Stats: monitoring.ResultStats{ResultType: "vector", SeriesCount: 1, SampleCount: 1},
	})
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}
	if got := result.Candidates[0].Labels["token"]; got != "<REDACTED>" {
		t.Fatalf("sensitive label not redacted: %q", got)
	}
}

func TestEvaluateRuleScalarAndStringStats(t *testing.T) {
	for _, resultType := range []string{"scalar", "string"} {
		result, err := EvaluateRule(testRule(NoDataOK), monitoring.PrometheusCallResult{
			Data:  map[string]any{"resultType": resultType, "result": []any{float64(1), "value"}},
			Stats: monitoring.ResultStats{ResultType: resultType, SeriesCount: 1, SampleCount: 1},
		})
		if err != nil {
			t.Fatalf("evaluate %s failed: %v", resultType, err)
		}
		if result.State != "ok" || result.Stats.ResultType != resultType || len(result.Samples) != 1 {
			t.Fatalf("unexpected %s result: %+v", resultType, result)
		}
	}
}

func TestEvaluateRuleMatrixDoesNotPanic(t *testing.T) {
	result, err := EvaluateRule(testRule(NoDataOK), monitoring.PrometheusCallResult{
		Data: map[string]any{"resultType": "matrix", "result": []any{
			map[string]any{"metric": map[string]any{"instance": "host-a"}, "values": []any{[]any{float64(1), "1"}}},
		}},
		Stats: monitoring.ResultStats{ResultType: "matrix", SeriesCount: 1, SampleCount: 1},
	})
	if err != nil {
		t.Fatalf("matrix should be represented as invalid response, not error: %v", err)
	}
	if result.State != "invalid_response" {
		t.Fatalf("expected invalid_response for matrix, got %+v", result)
	}
}

func TestEvaluateRuleForDurationMillis(t *testing.T) {
	result, err := EvaluateRule(testRule(NoDataOK), emptyVector())
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}
	if result.ForDurationMS != 5*60*1000 {
		t.Fatalf("unexpected for duration ms: %d", result.ForDurationMS)
	}
}

func emptyVector() monitoring.PrometheusCallResult {
	return monitoring.PrometheusCallResult{
		Data:  map[string]any{"resultType": "vector", "result": []any{}},
		Stats: monitoring.ResultStats{ResultType: "vector"},
	}
}

func testRule(policy string) Rule {
	return Rule{ID: "rule-a", DatasourceID: "prometheus-default", QueryHash: "hash-a", ForDuration: "5m", NoDataPolicy: policy}
}
