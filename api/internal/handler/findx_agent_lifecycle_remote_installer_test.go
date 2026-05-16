package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentRemoteSSHInstallExecutePreflightBlocksWithScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"linux",
		"method":"ssh",
		"mode":"execute",
		"target_ids":["target-ssh-a"],
		"metadata":{"transport":"ssh","runner":"custom-lab-runner","token":"secret"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("ssh remote execute should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	payload := decodeBlockedExecutionMatrixResponse(t, w.Body.String())
	assertRemoteInstallBlockedEnvelope(t, payload, "ssh", w.Body.String())
	for _, want := range []string{"credential_ref", "ssh_runner_ref", "ssh_host_key_or_fingerprint", "remote_executor_ref", "timeout_policy_ref", "idempotency_key", "install_receipt_ref", "execution_receipt_ref", "service_receipt_ref", "heartbeat_validator_ref", "data_arrival_validator_ref", "audit_ref_or_evidence_chain_ref"} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("ssh missing contracts should include %s, got %#v", want, payload.MissingContracts)
		}
	}
	for _, forbidden := range []string{"custom-lab-runner", "secret", "token", "<CREDENTIAL_REF>"} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("ssh blocked gate must not echo %q: %s", forbidden, w.Body.String())
		}
	}
}

func TestFindXAgentRemoteWinRMInstallExecutePreflightBlocksWithScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"package_id":"host-collector",
		"os":"windows",
		"method":"remote-install",
		"mode":"execute",
		"target_ids":["target-winrm-a"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{"transport":"winrm","runner":"winrm","session":"secret"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
	if w.Code != http.StatusConflict {
		t.Fatalf("winrm remote execute should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	payload := decodeBlockedExecutionMatrixResponse(t, w.Body.String())
	assertRemoteInstallBlockedEnvelope(t, payload, "winrm", w.Body.String())
	for _, want := range []string{"winrm_endpoint_ref", "winrm_transport_ref", "remote_executor_ref", "timeout_policy_ref", "idempotency_key"} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("winrm missing contracts should include %s, got %#v", want, payload.MissingContracts)
		}
	}
	for _, forbidden := range []string{"ssh_runner_ref", "ssh_host_key_or_fingerprint", "secret", "session", "<CREDENTIAL_REF>"} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("winrm blocked gate must not echo or require %q: %s", forbidden, w.Body.String())
		}
	}
}

func TestFindXAgentRemoteInstallCompleteRefsStillExecutorBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		scope    string
		os       string
		metadata string
	}{
		{scope: "ssh", os: "linux", metadata: completeSSHRemoteInstallMetadata()},
		{scope: "winrm", os: "windows", metadata: completeWinRMRemoteInstallMetadata()},
	} {
		t.Run(tt.scope, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			body := strings.NewReader(`{
				"package_id":"host-collector",
				"os":"` + tt.os + `",
				"method":"` + tt.scope + `",
				"mode":"execute",
				"target_ids":["target-a"],
				"credential_ref":"<CREDENTIAL_REF>",
				"metadata":{` + tt.metadata + `}
			}`)
			w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", body, CreateFindXAgentInstallPlan)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s complete remote execute should stay blocked, got %d body=%s", tt.scope, w.Code, w.Body.String())
			}
			payload := decodeBlockedExecutionMatrixResponse(t, w.Body.String())
			assertRemoteInstallBlockedEnvelope(t, payload, tt.scope, w.Body.String())
			want := "PENDING: " + tt.scope + " remote executor contract is not enabled"
			if payload.Execution.ErrorSummary != want {
				t.Fatalf("%s complete refs should still executor-block, want %q got %#v", tt.scope, want, payload.Execution)
			}
			for _, forbidden := range []string{"credential_ref", "remote_executor_ref", "timeout_policy_ref", "idempotency_key", "install_receipt_ref", "execution_receipt_ref", "ssh_runner_ref", "ssh_host_key_or_fingerprint", "winrm_endpoint_ref", "winrm_transport_ref", "systemd_unit_contract"} {
				if containsLifecycleTestString(payload.MissingContracts, forbidden) {
					t.Fatalf("%s complete refs should not report missing prerequisite %s, got %#v", tt.scope, forbidden, payload.MissingContracts)
				}
			}
			if tt.scope == "ssh" && (!containsLifecycleTestString(payload.MissingContracts, "ssh_runner_contract") || !containsLifecycleTestString(payload.MissingContracts, "host_key_contract")) {
				t.Fatalf("ssh complete refs should still expose closed ssh execution contracts, got %#v", payload.MissingContracts)
			}
			if tt.scope == "winrm" && !containsLifecycleTestString(payload.MissingContracts, "winrm_transport_contract") {
				t.Fatalf("winrm complete refs should still expose closed winrm transport contract, got %#v", payload.MissingContracts)
			}
		})
	}
}

func TestFindXAgentTaskRemoteSSHAndWinRMUsePreflightSemantics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		scope          string
		metadata       map[string]string
		wantTransport  string
		forbiddenValue string
	}{
		{
			scope: "ssh",
			metadata: map[string]string{
				"target_os": "linux", "transport": "ssh", "runner": "ssh", "service_ref": "svc-1", "executor_ref": "executor-1",
				"idempotency_key": "idem-1", "timeout_policy_ref": "timeout-1", "execution_receipt_ref": "receipt-1", "audit_ref": "audit-1",
				"data_arrival_validator_ref": "validator-1", "ssh_runner_ref": "ssh-runner-1", "ssh_fingerprint": "SHA256:example", "remote_executor_ref": "remote-executor-1",
			},
		},
		{
			scope: "winrm",
			metadata: map[string]string{
				"target_os": "windows", "transport": "winrm", "runner": "custom-lab-runner", "service_ref": "svc-1", "executor_ref": "executor-1",
				"idempotency_key": "idem-1", "timeout_policy_ref": "timeout-1", "execution_receipt_ref": "receipt-1", "audit_ref": "audit-1",
				"data_arrival_validator_ref": "validator-1", "winrm_endpoint_ref": "winrm-endpoint-1", "winrm_transport_ref": "winrm-transport-1", "remote_executor_ref": "remote-executor-1",
			},
			wantTransport:  "winrm",
			forbiddenValue: "custom-lab-runner",
		},
	} {
		t.Run(tt.scope, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			if tt.wantTransport == "" {
				tt.wantTransport = tt.scope
			}
			req := model.FindXAgentTaskRequest{
				Action:        "restart",
				TargetIDs:     []string{"target-a"},
				CredentialRef: "<CREDENTIAL_REF>",
				Metadata:      tt.metadata,
			}
			raw, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("marshal task request: %v", err)
			}
			w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", strings.NewReader(string(raw)), CreateFindXAgentTask)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s task should stay blocked 409, got %d body=%s", tt.scope, w.Code, w.Body.String())
			}
			payload := decodeTaskMatrixResponse(t, w.Body.String())
			if payload.ReceiptContract.Scope != tt.scope || payload.ReceiptContract.Transport != tt.wantTransport || payload.ReceiptContract.Runner != tt.scope {
				t.Fatalf("%s task receipt should keep standard remote scope, got %#v", tt.scope, payload.ReceiptContract)
			}
			if payload.SafeToRetry || payload.Data.Blocker != "PENDING: executor not enabled / execution protocol not open" {
				t.Fatalf("%s task must be non-retryable executor-blocked, got %#v", tt.scope, payload)
			}
			if tt.forbiddenValue != "" && strings.Contains(w.Body.String(), tt.forbiddenValue) {
				t.Fatalf("%s task must not echo custom runner %q: %s", tt.scope, tt.forbiddenValue, w.Body.String())
			}
			assertRemoteNoFakeState(t, w.Body.String())
			assertNoCredentialEcho(t, w.Body.String())
		})
	}
}

func assertRemoteInstallBlockedEnvelope(t *testing.T, payload blockedExecutionMatrixResponse, scope string, body string) {
	t.Helper()
	if payload.Status != "blocked" || payload.Execution.Status != model.FindXAgentExecutionStatusBlocked || payload.Execution.Runner != scope ||
		payload.ReceiptContract.Scope != scope || payload.ReceiptContract.Transport != scope || payload.ReceiptContract.Runner != scope ||
		payload.StateMachine.CurrentState != model.FindXAgentExecutionStateBlockedByContract || payload.SafeToRetry {
		t.Fatalf("unexpected %s blocked envelope: %#v", scope, payload)
	}
	assertRemoteNoFakeState(t, body)
	assertNoCredentialEcho(t, body)
}

func assertRemoteNoFakeState(t *testing.T, body string) {
	t.Helper()
	var raw any
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatalf("decode body for fake state scan: %v body=%s", err, body)
	}
	for _, field := range []string{"status", "current_state", "state", "allowed_states"} {
		for _, value := range collectJSONFieldStrings(raw, field) {
			switch value {
			case "queued", "running", "succeeded", "success", "installed", "service_registered", "heartbeat_seen", "data_arrival_seen", "uninstalled", "rolled_back":
				t.Fatalf("remote blocked gate must not expose fake %s=%s: %s", field, value, body)
			}
		}
	}
}

func decodeTaskMatrixResponse(t *testing.T, body string) blockedTaskMatrixResponse {
	t.Helper()
	var payload blockedTaskMatrixResponse
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode task matrix response: %v body=%s", err, body)
	}
	return payload
}

func completeSSHRemoteInstallMetadata() string {
	return `"transport":"ssh","runner":"ssh","package_repository_ref":"repo-ref","signature_ref":"sig-ref","checksum":"sha256:abc","script_manifest_ref":"script-ref","remote_executor_ref":"remote-executor-ref","timeout_policy_ref":"timeout-ref","idempotency_key":"idem-1","install_receipt_ref":"install-receipt-ref","execution_receipt_ref":"execution-receipt-ref","service_receipt_ref":"service-receipt-ref","heartbeat_validator_ref":"heartbeat-ref","data_arrival_validator_ref":"data-arrival-ref","audit_ref":"audit-ref","ssh_runner_ref":"ssh-runner-ref","ssh_fingerprint":"SHA256:example","ticket":"CHG-1","token":"secret"`
}

func completeWinRMRemoteInstallMetadata() string {
	return `"transport":"winrm","runner":"winrm","package_repository_ref":"repo-ref","signature_ref":"sig-ref","checksum":"sha256:abc","script_manifest_ref":"script-ref","remote_executor_ref":"remote-executor-ref","timeout_policy_ref":"timeout-ref","idempotency_key":"idem-1","install_receipt_ref":"install-receipt-ref","execution_receipt_ref":"execution-receipt-ref","service_receipt_ref":"service-receipt-ref","heartbeat_validator_ref":"heartbeat-ref","data_arrival_validator_ref":"data-arrival-ref","audit_ref":"audit-ref","winrm_endpoint_ref":"winrm-endpoint-ref","winrm_transport_ref":"winrm-transport-ref","ticket":"CHG-1","token":"secret"`
}
