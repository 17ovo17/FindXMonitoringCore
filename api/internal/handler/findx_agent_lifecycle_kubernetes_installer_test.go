package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentInstallPlanKubernetesExecuteCreatesBlockedExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"container-collector",
		"os":"Kubernetes",
		"method":"helm",
		"mode":"execute",
		"target_ids":["cluster-a"],
		"metadata":{"token":"top-secret-token","cluster_ref":"cluster-ref"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("Kubernetes execute mode should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	payload := decodeKubernetesInstallExecutionPayload(t, w.Body.Bytes())
	if payload.Execution.Status != model.FindXAgentExecutionStatusBlocked ||
		payload.Execution.Runner != "helm" ||
		payload.Execution.TargetID != "cluster-a" {
		t.Fatalf("expected blocked Kubernetes helm execution, got %#v", payload.Execution)
	}
	if !strings.Contains(w.Body.String(), "credential_ref") ||
		!strings.Contains(w.Body.String(), "namespace_ref") ||
		!strings.Contains(w.Body.String(), "helm_chart_ref_or_manifest_bundle_ref") {
		t.Fatalf("expected Kubernetes missing refs, got %s", w.Body.String())
	}
	assertKubernetesInstallSteps(t, payload.Execution.Steps)
	assertNoKubernetesCredentialValueEcho(t, w.Body.String())

	list := performAgentLifecycleGet("/api/v1/findx-agents/install-executions", ListFindXAgentInstallExecutions)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), payload.Execution.ID) {
		t.Fatalf("expected Kubernetes execution readback, got %d body=%s", list.Code, list.Body.String())
	}
}

func TestFindXAgentInstallPlanKubernetesCompleteRefsStillBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"container-collector",
		"os":"Kubernetes",
		"method":"k8s-daemonset",
		"mode":"execute",
		"target_ids":["cluster-b"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{
			"cluster_ref":"cluster-ref",
			"namespace_ref":"namespace-ref",
			"workload_selector_ref":"workload-selector-ref",
			"manifest_bundle_ref":"manifest-ref",
			"values_ref":"values-ref",
			"rbac_ref":"rbac-ref",
			"service_account_ref":"service-account-ref",
			"image_ref":"image-ref",
			"package_repository_ref":"repo-ref",
			"signature_ref":"sig-ref",
			"checksum":"sha256:abc",
			"config_map_ref":"config-map-ref",
			"rollout_strategy_ref":"rollout-strategy-ref",
			"rollout_receipt_ref":"rollout-receipt-ref",
			"data_arrival_validator_ref":"data-arrival-validator-ref",
			"executor_ref":"executor-ref",
			"evidence_chain_ref":"evidence-chain-ref",
			"session":"top-secret-session"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("Kubernetes daemonset execute mode should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	payload := decodeKubernetesInstallExecutionPayload(t, w.Body.Bytes())
	if payload.Execution.Runner != "kubernetes-daemonset" {
		t.Fatalf("expected daemonset runner, got %#v", payload.Execution)
	}
	if payload.Execution.ErrorSummary != "BLOCKED_BY_CONTRACT: Kubernetes executor not enabled / lifecycle protocol not open" {
		t.Fatalf("executor must remain blocked, got %q", payload.Execution.ErrorSummary)
	}
	for _, forbidden := range []string{"queued", "running", "succeeded"} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("Kubernetes installer response must not contain %s: %s", forbidden, w.Body.String())
		}
	}
	assertNoKubernetesCredentialValueEcho(t, w.Body.String())
}

func TestFindXAgentInstallPlanKubernetesRunnerDetectionByMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   string
	}{
		{name: "kubernetes", method: "kubernetes", want: "kubernetes"},
		{name: "k8s", method: "k8s", want: "kubernetes"},
		{name: "sidecar", method: "sidecar", want: "kubernetes-sidecar"},
		{name: "initcontainer", method: "initcontainer", want: "kubernetes-initcontainer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			resetAgentLifecycleRecordsForTest(t)
			body := strings.NewReader(`{
				"package_id":"container-collector",
				"os":"linux",
				"method":"` + tt.method + `",
				"mode":"execute",
				"target_ids":["cluster-c"]
			}`)
			w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
			if w.Code != http.StatusConflict {
				t.Fatalf("method %s should stay blocked, got %d body=%s", tt.method, w.Code, w.Body.String())
			}
			payload := decodeKubernetesInstallExecutionPayload(t, w.Body.Bytes())
			if payload.Execution.Runner != tt.want {
				t.Fatalf("expected runner %q, got %#v", tt.want, payload.Execution)
			}
		})
	}
}

func TestFindXAgentInstallPlanKubernetesPlanModeDoesNotCreateExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"container-collector",
		"os":"Kubernetes",
		"method":"helm",
		"mode":"plan",
		"target_ids":["cluster-plan"]
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict || strings.Contains(w.Body.String(), `"execution"`) {
		t.Fatalf("plan mode should return blocked plan only, got %d body=%s", w.Code, w.Body.String())
	}
	list := performAgentLifecycleGet("/api/v1/findx-agents/install-executions", ListFindXAgentInstallExecutions)
	if list.Code != http.StatusOK || strings.TrimSpace(list.Body.String()) != "[]" {
		t.Fatalf("plan mode must not persist execution, got %d body=%s", list.Code, list.Body.String())
	}
}

func decodeKubernetesInstallExecutionPayload(t *testing.T, data []byte) struct {
	Data      model.FindXAgentInstallPlan      `json:"data"`
	Execution model.FindXAgentInstallExecution `json:"execution"`
} {
	t.Helper()
	var payload struct {
		Data      model.FindXAgentInstallPlan      `json:"data"`
		Execution model.FindXAgentInstallExecution `json:"execution"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if payload.Data.ID == "" || payload.Execution.PlanID != payload.Data.ID {
		t.Fatalf("expected execution linked to install plan, got %#v", payload)
	}
	return payload
}

func assertKubernetesInstallSteps(t *testing.T, steps []model.FindXAgentInstallExecutionStep) {
	t.Helper()
	want := []string{
		"resolve_cluster",
		"validate_namespace",
		"verify_package",
		"verify_rbac",
		"render_manifest",
		"prepare_rollout",
		"verify_data_arrival",
		"capture_evidence",
	}
	if len(steps) != len(want) {
		t.Fatalf("expected Kubernetes blocked chain steps %#v, got %#v", want, steps)
	}
	for i, name := range want {
		if steps[i].Name != name || steps[i].Status != model.FindXAgentExecutionStatusBlocked {
			t.Fatalf("unexpected step[%d], got %#v", i, steps[i])
		}
	}
}

func assertNoKubernetesCredentialValueEcho(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{`"credential_ref":`, "<CREDENTIAL_REF>", `"token":`, `"session":`, "top-secret-token", "top-secret-session"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response must not echo credential or sensitive metadata values: %s", body)
		}
	}
}
