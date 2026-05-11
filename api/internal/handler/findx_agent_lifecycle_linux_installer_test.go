package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

const handlerLinuxInstallerSafetyPolicyPath = "scripts/testing/safety_policy.json"

func TestFindXAgentInstallPlanExecuteCreatesBlockedExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"linux",
		"method":"linux-curl",
		"mode":"execute",
		"target_ids":["target-a"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{"ticket":"CHG-1","password":"secret"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("execute mode should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	var payload struct {
		Data      model.FindXAgentInstallPlan      `json:"data"`
		Execution model.FindXAgentInstallExecution `json:"execution"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if payload.Data.ID == "" || payload.Execution.PlanID != payload.Data.ID {
		t.Fatalf("expected execution linked to install plan, got %#v", payload)
	}
	if payload.Execution.Status != model.FindXAgentExecutionStatusBlocked || payload.Execution.TargetID != "target-a" {
		t.Fatalf("expected blocked target execution, got %#v", payload.Execution)
	}
	if !strings.Contains(payload.Execution.ErrorSummary, "package_repository_ref") ||
		!strings.Contains(payload.Execution.ErrorSummary, "script_manifest_ref") ||
		!strings.Contains(payload.Execution.ErrorSummary, "executor_ref") {
		t.Fatalf("expected prerequisite blockers, got %q", payload.Execution.ErrorSummary)
	}
	assertLinuxInstallSteps(t, payload.Execution.Steps)
	assertNoCredentialEcho(t, w.Body.String())
	assertNoLinuxInstallSuccessStates(t, w.Body.String())

	list := performAgentLifecycleGet("/api/v1/findx-agents/install-executions", ListFindXAgentInstallExecutions)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), payload.Execution.ID) {
		t.Fatalf("expected execution list readback, got %d body=%s", list.Code, list.Body.String())
	}
}

func TestFindXAgentInstallPlanExecuteBooleanAndBadMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"linux",
		"method":"linux-curl",
		"execute":true,
		"target_ids":["target-a"]
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), `"execution"`) {
		t.Fatalf("execute boolean should create blocked execution, got %d body=%s", w.Code, w.Body.String())
	}

	bad := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", strings.NewReader(`{
		"package_id":"host-collector",
		"mode":"now",
		"target_ids":["target-a"]
	}`), CreateFindXAgentInstallPlan)
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("bad mode should return 400, got %d body=%s", bad.Code, bad.Body.String())
	}
}

func TestFindXAgentInstallExecutionListEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	w := performAgentLifecycleGet("/api/v1/findx-agents/install-executions", ListFindXAgentInstallExecutions)
	if w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("install execution list should be empty, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestLinuxInstallerMetadataCanUseFingerprintWithoutEchoingSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	ensureHandlerDefaultLinuxInstallerSafetyPolicy(t)

	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"linux",
		"method":"linux-curl",
		"mode":"execute",
		"target_ids":["target-a"],
		"metadata":{
			"package_repository_ref":"repo-ref",
			"signature_ref":"sig-ref",
			"checksum":"sha256:abc",
			"script_manifest_ref":"script-manifest-ref",
			"executor_ref":"executor-ref",
			"safety_policy_path":"scripts/testing/safety_policy.json",
			"runner":"ssh",
			"ssh_fingerprint":"SHA256:example",
			"systemd_unit_ref":"systemd-ref",
			"curl_installer_ref":"curl-ref",
			"env_template_ref":"env-ref",
			"token":"secret"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("execute mode should still be blocked, got %d body=%s", w.Code, w.Body.String())
	}
	bodyText := w.Body.String()
	if strings.Contains(bodyText, "ssh_host_key_or_fingerprint") {
		t.Fatalf("fingerprint should satisfy host-key gate, body=%s", bodyText)
	}
	if !strings.Contains(bodyText, "executor is not enabled") {
		t.Fatalf("executor must remain blocked, body=%s", bodyText)
	}
	payload := decodeLinuxInstallExecutionPayload(t, w.Body.Bytes())
	assertLinuxInstallSteps(t, payload.Execution.Steps)
	assertNoLinuxInstallSuccessStates(t, bodyText)
	assertNoCredentialEcho(t, bodyText)
}

func decodeLinuxInstallExecutionPayload(t *testing.T, data []byte) struct {
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

func assertLinuxInstallSteps(t *testing.T, steps []model.FindXAgentInstallExecutionStep) {
	t.Helper()
	want := []string{
		"preflight",
		"download_package",
		"verify_package",
		"register_service",
		"enable_or_start_service",
		"verify_heartbeat",
		"verify_data_arrival",
		"capture_evidence",
	}
	if len(steps) != len(want) {
		t.Fatalf("expected Linux blocked chain steps %#v, got %#v", want, steps)
	}
	for i, name := range want {
		if steps[i].Name != name || steps[i].Status != model.FindXAgentExecutionStatusBlocked {
			t.Fatalf("unexpected step[%d], got %#v", i, steps[i])
		}
	}
}

func ensureHandlerDefaultLinuxInstallerSafetyPolicy(t *testing.T) {
	t.Helper()

	path := filepath.Join("..", "..", "..", handlerLinuxInstallerSafetyPolicyPath)
	if _, err := os.Stat(path); err == nil {
		return
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create safety policy dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"version":1}`), 0o600); err != nil {
		t.Fatalf("write safety policy: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(path)
	})
}

func assertNoLinuxInstallSuccessStates(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{
		`"status":"queued"`,
		`"status":"running"`,
		`"status":"succeeded"`,
		`"status":"success"`,
		`"status":"applied"`,
		`"status":"rolled-back"`,
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("Linux installer response must not expose %s: %s", forbidden, body)
		}
	}
}
