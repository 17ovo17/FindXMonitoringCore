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
		Data             model.FindXAgentInstallPlan      `json:"data"`
		Execution        model.FindXAgentInstallExecution `json:"execution"`
		MissingContracts []string                         `json:"missing_contracts"`
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
	if !strings.Contains(payload.Execution.ErrorSummary, "checksum") ||
		!strings.Contains(payload.Execution.ErrorSummary, "executor_ref") {
		t.Fatalf("expected prerequisite blockers, got %q", payload.Execution.ErrorSummary)
	}
	for _, missing := range []string{
		"package_repository_ref",
		"service_receipt_ref",
		"heartbeat_validator_ref",
		"data_arrival_validator_ref",
		"audit_ref_or_evidence_chain_ref",
		"service_status_receipt_ref",
	} {
		if !containsString(payload.MissingContracts, missing) {
			t.Fatalf("expected missing contract %q in %#v", missing, payload.MissingContracts)
		}
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
			"systemd_unit_name_ref":"unit-name-ref",
			"systemd_mode":"system",
			"curl_installer_ref":"curl-ref",
			"curl_command_ref":"curl-command-ref",
			"curl_manifest_ref":"curl-manifest-ref",
			"env_template_ref":"env-ref",
			"service_receipt_ref":"service-receipt-ref",
			"heartbeat_validator_ref":"heartbeat-validator-ref",
			"data_arrival_validator_ref":"data-arrival-validator-ref",
			"evidence_chain_ref":"evidence-chain-ref",
			"runner_whitelist_ref":"runner-whitelist-ref",
			"reload_receipt_ref":"reload-receipt-ref",
			"service_status_receipt_ref":"service-status-receipt-ref",
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

func TestLinuxInstallerLocalSystemdPreflightContractsStayBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	ensureHandlerDefaultLinuxInstallerSafetyPolicy(t)

	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"linux",
		"method":"linux-curl-systemd",
		"mode":"execute",
		"target_ids":["target-a"],
		"metadata":{
			"package_repository_ref":"repo-ref",
			"signature_ref":"sig-ref",
			"checksum":"sha256:abc",
			"script_manifest_ref":"script-manifest-ref",
			"executor_ref":"executor-ref",
			"safety_policy_path":"scripts/testing/safety_policy.json",
			"runner":"local-systemd",
			"systemd_unit_ref":"systemd-ref",
			"systemd_unit_path_ref":"unit-path-ref",
			"systemd_mode":"user",
			"curl_installer_ref":"curl-ref",
			"curl_command_ref":"curl-command-ref",
			"curl_manifest_ref":"curl-manifest-ref",
			"env_template_ref":"env-ref",
			"service_receipt_ref":"service-receipt-ref",
			"heartbeat_validator_ref":"heartbeat-validator-ref",
			"data_arrival_validator_ref":"data-arrival-validator-ref",
			"audit_ref":"audit-ref",
			"runner_whitelist_ref":"runner-whitelist-ref",
			"reload_receipt_ref":"reload-receipt-ref",
			"service_status_receipt_ref":"service-status-receipt-ref",
			"bearer":"secret",
			"private_key":"secret"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("execute mode should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	bodyText := w.Body.String()
	var response struct {
		Error            string   `json:"error"`
		MissingContracts []string `json:"missing_contracts"`
		SafeToRetry      bool     `json:"safe_to_retry"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	for _, forbidden := range []string{
		"package_repository_ref",
		"service_receipt_ref",
		"heartbeat_validator_ref",
		"data_arrival_validator_ref",
		"audit_ref_or_evidence_chain_ref",
		"systemd_unit_name_ref_or_systemd_unit_path_ref",
		"systemd_mode",
		"ssh_host_key_or_fingerprint",
	} {
		if strings.Contains(response.Error, forbidden) || containsString(response.MissingContracts, forbidden) {
			t.Fatalf("complete local systemd preflight should not miss %s: %#v", forbidden, response)
		}
	}
	if !strings.Contains(bodyText, "executor is not enabled") || !strings.Contains(bodyText, `"safe_to_retry":false`) {
		t.Fatalf("executor must remain blocked and non-retryable, body=%s", bodyText)
	}
	assertLinuxInstallSteps(t, decodeLinuxInstallExecutionPayload(t, w.Body.Bytes()).Execution.Steps)
	assertNoLinuxInstallSuccessStates(t, bodyText)
	assertNoCredentialEcho(t, bodyText)
}

func TestLinuxInstallerLocalSystemdPlanDoesNotEchoSensitiveMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"linux",
		"method":"linux-curl-systemd",
		"mode":"plan",
		"target_ids":["target-a"],
		"metadata":{
			"runner":"local-systemd",
			"ticket":"CHG-2",
			"bearer":"Bearer abc",
			"private_key":"-----BEGIN PRIVATE KEY-----",
			"token":"plain-token"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("plan mode should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	bodyText := w.Body.String()
	for _, forbidden := range []string{"Bearer abc", "BEGIN PRIVATE KEY", "plain-token", "local-systemd"} {
		if strings.Contains(bodyText, forbidden) {
			t.Fatalf("install plan response must not echo sensitive or raw runner metadata %q: %s", forbidden, bodyText)
		}
	}
	if !strings.Contains(bodyText, `"runner":"systemd"`) || !strings.Contains(bodyText, "CHG-2") {
		t.Fatalf("expected safe metadata to preserve normalized runner and ticket, body=%s", bodyText)
	}
	assertNoCredentialEcho(t, bodyText)
}

func TestLinuxInstallerPlanDropsUnknownRunnerMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"linux",
		"method":"linux-curl",
		"mode":"plan",
		"target_ids":["target-a"],
		"metadata":{"runner":"custom-lab-runner"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("plan mode should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "custom-lab-runner") {
		t.Fatalf("unknown runner metadata must not echo, body=%s", w.Body.String())
	}
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
		`"status":"installed"`,
		`"status":"data_arrived"`,
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("Linux installer response must not expose %s: %s", forbidden, body)
		}
	}
}
