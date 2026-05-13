package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

type blockedExecutionMatrixResponse struct {
	Status           string                                     `json:"status"`
	StateMachine     model.FindXAgentExecutionStateMachine      `json:"state_machine"`
	ReceiptContract  model.FindXAgentReceiptContract            `json:"receipt_contract"`
	ReceiptMatrix    []model.FindXAgentReceiptContractMatrixRow `json:"receipt_matrix"`
	MissingContracts []string                                   `json:"missing_contracts"`
	SafeToRetry      bool                                       `json:"safe_to_retry"`
	Data             model.FindXAgentInstallPlan                `json:"data"`
	Execution        model.FindXAgentInstallExecution           `json:"execution"`
}

type blockedTaskMatrixResponse struct {
	Status           string                                     `json:"status"`
	StateMachine     model.FindXAgentExecutionStateMachine      `json:"state_machine"`
	ReceiptContract  model.FindXAgentReceiptContract            `json:"receipt_contract"`
	ReceiptMatrix    []model.FindXAgentReceiptContractMatrixRow `json:"receipt_matrix"`
	MissingContracts []string                                   `json:"missing_contracts"`
	SafeToRetry      bool                                       `json:"safe_to_retry"`
	Data             model.FindXAgentExecutionTask              `json:"data"`
}

func TestFindXAgentInstallExecutionBlockedGateExposesReceiptMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name      string
		body      string
		wantScope string
	}{
		{
			name:      "linux ssh",
			wantScope: "ssh",
			body:      `{"package_id":"host-collector","os":"linux","method":"linux-curl","mode":"execute","target_ids":["target-a"],"metadata":{"runner":"ssh","ticket":"CHG-1","token":"secret"}}`,
		},
		{
			name:      "windows local service",
			wantScope: "windows_local",
			body:      `{"package_id":"host-collector","os":"windows","method":"windows-powershell","mode":"execute","target_ids":["target-a"],"credential_ref":"<CREDENTIAL_REF>","metadata":{"ticket":"CHG-1","session":"secret"}}`,
		},
		{
			name:      "kubernetes helm",
			wantScope: "helm",
			body:      `{"package_id":"container-collector","os":"kubernetes","method":"helm","mode":"execute","target_ids":["cluster-a"],"credential_ref":"<CREDENTIAL_REF>","metadata":{"cluster_ref":"cluster-ref","password":"secret"}}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", strings.NewReader(tt.body), CreateFindXAgentInstallPlan)
			if w.Code != http.StatusConflict {
				t.Fatalf("execute gate should be 409, got %d body=%s", w.Code, w.Body.String())
			}
			payload := decodeBlockedExecutionMatrixResponse(t, w.Body.String())
			assertBlockedExecutionMatrixEnvelope(t, payload, tt.wantScope, w.Body.String())
		})
	}
}

func TestFindXAgentInstallExecutionBlockedGateSanitizesMetadataTransport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name          string
		method        string
		transport     string
		wantTransport string
	}{
		{
			name:          "secret token transport falls back to local",
			method:        "linux-curl",
			transport:     "secret-token",
			wantTransport: "local",
		},
		{
			name:          "bearer token transport falls back to ssh method",
			method:        "ssh",
			transport:     "bearer-token",
			wantTransport: "ssh",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			req := model.FindXAgentInstallPlanRequest{
				PackageID: "host-collector",
				OS:        "linux",
				Method:    tt.method,
				Mode:      "execute",
				TargetIDs: []string{"target-a"},
				Metadata: map[string]string{
					"transport": tt.transport,
					"ticket":    "CHG-1",
				},
			}
			raw, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("marshal install request: %v", err)
			}
			w := performAgentLifecyclePost("/api/v1/findx-agents/install-plans", strings.NewReader(string(raw)), CreateFindXAgentInstallPlan)
			if w.Code != http.StatusConflict {
				t.Fatalf("execute gate should be 409, got %d body=%s", w.Code, w.Body.String())
			}
			payload := decodeBlockedExecutionMatrixResponse(t, w.Body.String())
			if payload.SafeToRetry {
				t.Fatalf("blocked transport gate must not be retryable, got %#v", payload)
			}
			if payload.ReceiptContract.Status != model.FindXAgentExecutionStateBlockedByContract {
				t.Fatalf("receipt contract should stay blocked, got %#v", payload.ReceiptContract)
			}
			if payload.ReceiptContract.Transport != tt.wantTransport {
				t.Fatalf("receipt transport should be sanitized to %s, got %#v", tt.wantTransport, payload.ReceiptContract)
			}
			if strings.Contains(w.Body.String(), tt.transport) {
				t.Fatalf("blocked gate must not echo metadata.transport %q: %s", tt.transport, w.Body.String())
			}
		})
	}
}

func TestFindXAgentTaskBlockedGateExposesAllReceiptSurfaces(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	envelope, body := postAgentTaskForMatrixGate(t, completeKubernetesWorkloadMetadata("sidecar"))
	assertStateMachine121States(t, envelope.StateMachine)
	if envelope.Status != "blocked" || envelope.StateMachine.CurrentState != model.FindXAgentExecutionStateBlockedByContract || envelope.SafeToRetry {
		t.Fatalf("task must be blocked_by_contract and not retryable, got %#v", envelope)
	}
	for _, scope := range []string{"linux_local", "windows_local", "ssh", "winrm", "systemd", "windows_service", "iis", "docker", "helm", "operator", "daemonset", "sidecar", "initcontainer"} {
		if !receiptMatrixHasScope(envelope.ReceiptMatrix, scope) {
			t.Fatalf("receipt matrix should contain %s, matrix=%#v", scope, envelope.ReceiptMatrix)
		}
	}
	assertNoFakeExecutionStatusInMatrixResponse(t, body)
	for _, forbidden := range []string{`"status":"installed"`, `"status":"data_arrived"`, "<CREDENTIAL_REF>", "secret"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("task blocked gate must not expose fake state or sensitive value %s: %s", forbidden, body)
		}
	}
}

func TestFindXAgentTaskBlockedGateSanitizesReceiptTransport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name          string
		metadata      map[string]string
		forbidden     []string
		wantTransport string
	}{
		{
			name: "custom transport falls back to local",
			metadata: func() map[string]string {
				metadata := completeLocalTaskMetadata("linux")
				metadata["transport"] = "custom-lab-runner"
				return metadata
			}(),
			forbidden:     []string{"custom-lab-runner"},
			wantTransport: "local",
		},
		{
			name: "secret transport falls back to local",
			metadata: func() map[string]string {
				metadata := completeLocalTaskMetadata("linux")
				metadata["transport"] = "bearer-token"
				return metadata
			}(),
			forbidden:     []string{"bearer-token"},
			wantTransport: "local",
		},
		{
			name: "secret runner falls back to sidecar scope",
			metadata: func() map[string]string {
				metadata := completeKubernetesWorkloadMetadata("sidecar")
				delete(metadata, "transport")
				metadata["runner"] = "secret-token"
				return metadata
			}(),
			forbidden:     []string{"secret-token"},
			wantTransport: "sidecar",
		},
		{
			name:          "standard ssh is preserved",
			metadata:      completeTransportTaskMetadata("ssh"),
			forbidden:     nil,
			wantTransport: "ssh",
		},
		{
			name:          "standard winrm is preserved",
			metadata:      completeTransportTaskMetadata("winrm"),
			forbidden:     nil,
			wantTransport: "winrm",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			envelope, body := postAgentTaskForMatrixGate(t, tt.metadata)
			assertNoFakeExecutionStatusInMatrixResponse(t, body)
			if envelope.ReceiptContract.Status != model.FindXAgentExecutionStateBlockedByContract {
				t.Fatalf("receipt contract should stay blocked, got %#v", envelope.ReceiptContract)
			}
			if envelope.ReceiptContract.Transport != tt.wantTransport {
				t.Fatalf("receipt transport should be sanitized to %s, got %#v", tt.wantTransport, envelope.ReceiptContract)
			}
			for _, forbidden := range tt.forbidden {
				if strings.Contains(body, forbidden) {
					t.Fatalf("task blocked gate must not echo transport %q: %s", forbidden, body)
				}
			}
		})
	}
}

func TestFindXAgentExecutionStateStatusValidatorAccepts121States(t *testing.T) {
	for _, state := range []string{
		model.FindXAgentExecutionStatePlanned,
		model.FindXAgentExecutionStatePreflightFailed,
		model.FindXAgentExecutionStateBlockedByContract,
		model.FindXAgentExecutionStateDispatching,
		model.FindXAgentExecutionStateRunning,
		model.FindXAgentExecutionStateReceiptPending,
		model.FindXAgentExecutionStateServiceRegistered,
		model.FindXAgentExecutionStateHeartbeatSeen,
		model.FindXAgentExecutionStateDataArrivalSeen,
		model.FindXAgentExecutionStateFailed,
		model.FindXAgentExecutionStateRolledBack,
		model.FindXAgentExecutionStateUninstalled,
	} {
		if !model.IsFindXAgentInstallExecutionStatus(state) {
			t.Fatalf("121 state should be accepted by execution validator: %s", state)
		}
	}
}

func decodeBlockedExecutionMatrixResponse(t *testing.T, body string) blockedExecutionMatrixResponse {
	t.Helper()
	var payload blockedExecutionMatrixResponse
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode blocked gate: %v body=%s", err, body)
	}
	return payload
}

func postAgentTaskForMatrixGate(t *testing.T, metadata map[string]string) (blockedTaskMatrixResponse, string) {
	t.Helper()
	req := model.FindXAgentTaskRequest{
		Action:        "restart",
		TargetIDs:     []string{"target-a"},
		CredentialRef: "<CREDENTIAL_REF>",
		Metadata:      metadata,
	}
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal task request: %v", err)
	}
	w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", strings.NewReader(string(raw)), CreateFindXAgentTask)
	if w.Code != http.StatusConflict {
		t.Fatalf("agent task should stay blocked 409, got %d body=%s", w.Code, w.Body.String())
	}
	var payload blockedTaskMatrixResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode task matrix response: %v body=%s", err, w.Body.String())
	}
	return payload, w.Body.String()
}

func assertBlockedExecutionMatrixEnvelope(t *testing.T, payload blockedExecutionMatrixResponse, scope string, body string) {
	t.Helper()
	assertStateMachine121States(t, payload.StateMachine)
	if payload.Status != "blocked" || payload.Execution.Status != model.FindXAgentExecutionStatusBlocked ||
		payload.StateMachine.CurrentState != model.FindXAgentExecutionStateBlockedByContract || payload.SafeToRetry {
		t.Fatalf("install response must be blocked and not retryable, got %#v", payload)
	}
	if payload.ReceiptContract.Scope != scope || payload.ReceiptContract.Status != model.FindXAgentExecutionStateBlockedByContract {
		t.Fatalf("unexpected receipt contract scope/status, got %#v", payload.ReceiptContract)
	}
	for _, required := range []string{"executor_contract", "execution_receipt_contract", "service_receipt_contract", "heartbeat_receipt_contract", "data_arrival_receipt_contract", "evidence_chain_contract"} {
		if !containsLifecycleTestString(payload.MissingContracts, required) {
			t.Fatalf("missing contracts should include %s, got %#v", required, payload.MissingContracts)
		}
	}
	for _, scope := range []string{"linux_local", "windows_local", "ssh", "winrm", "systemd", "windows_service", "iis", "docker", "helm", "operator", "daemonset", "sidecar", "initcontainer"} {
		if !receiptMatrixHasScope(payload.ReceiptMatrix, scope) {
			t.Fatalf("receipt matrix should contain %s, matrix=%#v", scope, payload.ReceiptMatrix)
		}
	}
	assertNoFakeExecutionStatusInMatrixResponse(t, body)
	for _, forbidden := range []string{`"status":"installed"`, `"status":"data_arrived"`, "<CREDENTIAL_REF>", "secret"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("install blocked gate must not expose fake state or sensitive value %s: %s", forbidden, body)
		}
	}
}

func assertStateMachine121States(t *testing.T, stateMachine model.FindXAgentExecutionStateMachine) {
	t.Helper()
	for _, state := range []string{
		model.FindXAgentExecutionStatePlanned,
		model.FindXAgentExecutionStatePreflightFailed,
		model.FindXAgentExecutionStateBlockedByContract,
		model.FindXAgentExecutionStateDispatching,
		model.FindXAgentExecutionStateRunning,
		model.FindXAgentExecutionStateReceiptPending,
		model.FindXAgentExecutionStateServiceRegistered,
		model.FindXAgentExecutionStateHeartbeatSeen,
		model.FindXAgentExecutionStateDataArrivalSeen,
		model.FindXAgentExecutionStateFailed,
		model.FindXAgentExecutionStateRolledBack,
		model.FindXAgentExecutionStateUninstalled,
	} {
		if !containsLifecycleTestString(stateMachine.AllowedStates, state) {
			t.Fatalf("state machine missing %s: %#v", state, stateMachine)
		}
	}
}

func receiptMatrixHasScope(matrix []model.FindXAgentReceiptContractMatrixRow, scope string) bool {
	for _, row := range matrix {
		if row.Scope == scope && row.Status == model.FindXAgentExecutionStateBlockedByContract {
			return true
		}
	}
	return false
}

func assertNoFakeExecutionStatusInMatrixResponse(t *testing.T, body string) {
	t.Helper()
	var raw any
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatalf("decode body for fake status scan: %v body=%s", err, body)
	}
	for _, status := range collectJSONFieldStrings(raw, "status") {
		switch status {
		case "queued", "running", "succeeded", "success", "applied", "rolled-back", "installed", "data_arrived":
			t.Fatalf("blocked gate must not expose fake status %s: %s", status, body)
		}
	}
}

func collectJSONFieldStrings(value any, field string) []string {
	switch typed := value.(type) {
	case map[string]any:
		values := []string{}
		for key, item := range typed {
			if key == field {
				if text, ok := item.(string); ok {
					values = append(values, text)
				}
				continue
			}
			values = append(values, collectJSONFieldStrings(item, field)...)
		}
		return values
	case []any:
		values := []string{}
		for _, item := range typed {
			values = append(values, collectJSONFieldStrings(item, field)...)
		}
		return values
	default:
		return nil
	}
}
