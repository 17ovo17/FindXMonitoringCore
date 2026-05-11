package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentInstallExecutionListCoercesNonBlockedRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	blocked, err := store.SaveFindXAgentInstallExecution(model.FindXAgentInstallExecution{
		PlanID:       "plan-blocked",
		TargetID:     "target-blocked",
		Runner:       "ssh",
		Status:       model.FindXAgentExecutionStatusBlocked,
		Steps:        []model.FindXAgentInstallExecutionStep{{Name: "preflight", Status: model.FindXAgentExecutionStatusBlocked}},
		EvidenceRefs: []string{"install-plan:plan-blocked", "blocked-contract:linux-lifecycle"},
	})
	if err != nil {
		t.Fatalf("save blocked execution: %v", err)
	}
	nonBlocked, err := store.SaveFindXAgentInstallExecution(model.FindXAgentInstallExecution{
		PlanID:       "plan-running",
		TargetID:     "target-running",
		Runner:       "ssh",
		Status:       model.FindXAgentExecutionStatusRunning,
		Steps:        []model.FindXAgentInstallExecutionStep{{Name: "execute", Status: model.FindXAgentExecutionStatusSucceeded}},
		EvidenceRefs: []string{"install-plan:plan-running"},
	})
	if err != nil {
		t.Fatalf("save non-blocked execution: %v", err)
	}
	if nonBlocked.Status != model.FindXAgentExecutionStatusBlocked || nonBlocked.ErrorSummary == "" {
		t.Fatalf("store must coerce install execution to blocked, got %#v", nonBlocked)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/install-executions", ListFindXAgentInstallExecutions)
	if w.Code != http.StatusOK {
		t.Fatalf("install execution list should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, blocked.ID) || !strings.Contains(body, nonBlocked.ID) {
		t.Fatalf("list should include blocked records after store coercion, body=%s", body)
	}
	assertNoInstallExecutionSuccessStates(t, body)
	assertNoInstallExecutionSensitiveEcho(t, body)
}

func TestFindXAgentInstallExecutionDetailReturnsBlockedSafeRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	blocked, err := store.SaveFindXAgentInstallExecution(model.FindXAgentInstallExecution{
		PlanID:   "plan-blocked",
		TargetID: "target-blocked",
		Runner:   "ssh",
		Status:   model.FindXAgentExecutionStatusBlocked,
		Steps: []model.FindXAgentInstallExecutionStep{{
			Name:        "preflight",
			Status:      model.FindXAgentExecutionStatusBlocked,
			EvidenceRef: "install-step-ref",
		}},
		EvidenceRefs: []string{"install-plan:plan-blocked", "blocked-contract:linux-lifecycle"},
		ErrorSummary: "BLOCKED_BY_CONTRACT: executor not enabled / execution protocol not open",
	})
	if err != nil {
		t.Fatalf("save blocked execution: %v", err)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/install-executions?id="+blocked.ID, ListFindXAgentInstallExecutions)
	if w.Code != http.StatusOK {
		t.Fatalf("install execution detail should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	var detail model.FindXAgentInstallExecution
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("invalid detail response: %v body=%s", err, w.Body.String())
	}
	if detail.ID != blocked.ID || detail.Status != model.FindXAgentExecutionStatusBlocked || detail.PlanID != "plan-blocked" {
		t.Fatalf("detail should return the blocked execution record, got %#v", detail)
	}
	if len(detail.Steps) != 1 || detail.Steps[0].Name != "preflight" || detail.Steps[0].Status != model.FindXAgentExecutionStatusBlocked {
		t.Fatalf("detail should preserve blocked steps, got %#v", detail.Steps)
	}
	if len(detail.EvidenceRefs) != 2 || detail.EvidenceRefs[0] != "install-plan:plan-blocked" {
		t.Fatalf("detail should preserve evidence refs, got %#v", detail.EvidenceRefs)
	}
	assertNoInstallExecutionSuccessStates(t, w.Body.String())
	assertNoInstallExecutionSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentInstallExecutionDetailCoercesNonBlockedAndMissingReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	nonBlocked, err := store.SaveFindXAgentInstallExecution(model.FindXAgentInstallExecution{
		PlanID:       "plan-running",
		TargetID:     "target-running",
		Runner:       "ssh",
		Status:       model.FindXAgentExecutionStatusRunning,
		EvidenceRefs: []string{"install-plan:plan-running"},
	})
	if err != nil {
		t.Fatalf("save non-blocked execution: %v", err)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/install-executions?id="+nonBlocked.ID, ListFindXAgentInstallExecutions)
	if w.Code != http.StatusOK {
		t.Fatalf("coerced install execution detail should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"status":"blocked"`) || !strings.Contains(w.Body.String(), "executor not enabled") {
		t.Fatalf("coerced install execution detail should expose blocked gate, body=%s", w.Body.String())
	}
	assertNoInstallExecutionSuccessStates(t, w.Body.String())
	assertNoInstallExecutionSensitiveEcho(t, w.Body.String())

	missing := performAgentLifecycleGet("/api/v1/findx-agents/install-executions?id=missing-execution", ListFindXAgentInstallExecutions)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing install execution detail should return 404, got %d body=%s", missing.Code, missing.Body.String())
	}
	assertNoInstallExecutionSuccessStates(t, missing.Body.String())
	assertNoInstallExecutionSensitiveEcho(t, missing.Body.String())
}

func TestFindXAgentInstallExecutionBlockedDetailsKeepOSSpecificStepsAndAvoidSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range installExecutionBlockedDetailCases() {
		t.Run(tc.name, func(t *testing.T) {
			assertBlockedInstallExecutionScenario(t, tc)
		})
	}
}

type blockedInstallExecutionCase struct {
	name                    string
	body                    string
	wantRunner              string
	assertSteps             func(*testing.T, []model.FindXAgentInstallExecutionStep)
	assertSensitiveEchoGone func(*testing.T, string)
}

func installExecutionBlockedDetailCases() []blockedInstallExecutionCase {
	return []blockedInstallExecutionCase{
		linuxInstallExecutionBlockedDetailCase(),
		windowsInstallExecutionBlockedDetailCase(),
		kubernetesInstallExecutionBlockedDetailCase(),
	}
}

func linuxInstallExecutionBlockedDetailCase() blockedInstallExecutionCase {
	return blockedInstallExecutionCase{
		name: "linux",
		body: `{
				"package_id":"host-collector",
				"os":"linux",
				"method":"linux-curl",
				"mode":"execute",
				"target_ids":["target-a"],
				"credential_ref":"<CREDENTIAL_REF>",
				"metadata":{
					"package_repository_ref":"repo-ref",
					"signature_ref":"sig-ref",
					"checksum":"sha256:abc",
					"script_manifest_ref":"script-manifest-ref",
					"executor_ref":"executor-ref",
					"safety_policy_path":"scripts/testing/safety-policy.json",
					"runner":"ssh",
					"ssh_fingerprint":"SHA256:linux-fingerprint",
					"systemd_unit_ref":"systemd-ref",
					"curl_installer_ref":"curl-ref",
					"env_template_ref":"env-ref",
					"token":"linux-secret-token",
					"password":"linux-secret-password"
				}
			}`,
		wantRunner:              "ssh",
		assertSteps:             assertLinuxInstallSteps,
		assertSensitiveEchoGone: assertNoInstallExecutionSensitiveEcho,
	}
}

func windowsInstallExecutionBlockedDetailCase() blockedInstallExecutionCase {
	return blockedInstallExecutionCase{
		name: "windows",
		body: `{
				"package_id":"host-collector",
				"os":"windows",
				"method":"windows-powershell",
				"mode":"execute",
				"target_ids":["win-target-a"],
				"credential_ref":"<CREDENTIAL_REF>",
				"metadata":{
					"package_repository_ref":"repo-ref",
					"signature_ref":"sig-ref",
					"checksum":"sha256:abc",
					"script_manifest_ref":"script-manifest-ref",
					"executor_ref":"executor-ref",
					"windows_installer_ref":"installer-ref",
					"service_manifest_ref":"service-ref",
					"install_root_policy_ref":"root-policy-ref",
					"rollback_manifest_ref":"rollback-ref",
					"uninstall_manifest_ref":"uninstall-ref",
					"receiver_endpoint_ref":"receiver-ref",
					"token":"windows-secret-token"
				}
			}`,
		wantRunner:              "windows-powershell",
		assertSteps:             assertWindowsInstallSteps,
		assertSensitiveEchoGone: assertNoInstallExecutionSensitiveEcho,
	}
}

func kubernetesInstallExecutionBlockedDetailCase() blockedInstallExecutionCase {
	return blockedInstallExecutionCase{
		name: "kubernetes",
		body: `{
				"package_id":"container-collector",
				"os":"Kubernetes",
				"method":"helm",
				"mode":"execute",
				"target_ids":["cluster-a"],
				"metadata":{
					"cluster_ref":"cluster-ref",
					"namespace_ref":"namespace-ref",
					"workload_selector_ref":"workload-selector-ref",
					"manifest_bundle_ref":"manifest-ref",
					"rbac_ref":"rbac-ref",
					"service_account_ref":"service-account-ref",
					"rollout_strategy_ref":"rollout-strategy-ref",
					"rollout_receipt_ref":"rollout-receipt-ref",
					"data_arrival_validator_ref":"data-arrival-validator-ref",
					"executor_ref":"executor-ref",
					"evidence_chain_ref":"evidence-chain-ref",
					"token":"k8s-secret-token",
					"session":"k8s-secret-session"
				}
			}`,
		wantRunner:              "helm",
		assertSteps:             assertKubernetesInstallSteps,
		assertSensitiveEchoGone: assertNoKubernetesCredentialValueEcho,
	}
}

func assertBlockedInstallExecutionScenario(t *testing.T, tc blockedInstallExecutionCase) {
	t.Helper()
	resetAgentLifecycleRecordsForTest(t)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", strings.NewReader(tc.body), CreateFindXAgentInstallPlan)
	payload := decodeInstallExecutionPayload(t, w.Body.Bytes())
	if w.Code != http.StatusConflict {
		t.Fatalf("%s execute mode should stay blocked, got %d body=%s", tc.name, w.Code, w.Body.String())
	}
	detail := getInstallExecutionDetail(t, payload.Execution.ID)
	if detail.Status != model.FindXAgentExecutionStatusBlocked || detail.Runner != tc.wantRunner {
		t.Fatalf("%s detail should stay blocked, got %#v", tc.name, detail)
	}
	tc.assertSteps(t, detail.Steps)
	tc.assertSensitiveEchoGone(t, w.Body.String())
	tc.assertSensitiveEchoGone(t, mustMarshalInstallExecutionDetail(t, detail))
}

func decodeInstallExecutionPayload(t *testing.T, data []byte) struct {
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

func getInstallExecutionDetail(t *testing.T, id string) model.FindXAgentInstallExecution {
	t.Helper()
	w := performAgentLifecycleGet("/api/v1/findx-agents/install-executions?id="+id, ListFindXAgentInstallExecutions)
	if w.Code != http.StatusOK {
		t.Fatalf("install execution detail should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	var detail model.FindXAgentInstallExecution
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("invalid install execution detail: %v body=%s", err, w.Body.String())
	}
	if detail.ID != id {
		t.Fatalf("expected install execution %s, got %#v", id, detail)
	}
	return detail
}

func mustMarshalInstallExecutionDetail(t *testing.T, detail model.FindXAgentInstallExecution) string {
	t.Helper()
	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("marshal detail: %v", err)
	}
	return string(data)
}

func assertNoInstallExecutionSuccessStates(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{
		`"status":"queued"`,
		`"status":"running"`,
		`"status":"succeeded"`,
		`"status":"success"`,
		`"status":"applied"`,
		`"status":"rolled-back"`,
		`"status":"failed"`,
		`"status":"cancelled"`,
		`"status":"rollback_required"`,
		`"status":"uninstall_verified"`,
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("install execution response must not expose success states: %s", body)
		}
	}
}

func assertNoInstallExecutionSensitiveEcho(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{
		"<CREDENTIAL_REF>",
		"linux-secret-token",
		"linux-secret-password",
		"windows-secret-token",
		"k8s-secret-token",
		"k8s-secret-session",
		"top-secret-token",
		"top-secret-session",
		"mysql://user:pass@host/db",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("install execution response must not echo sensitive values: %s", body)
		}
	}
}
