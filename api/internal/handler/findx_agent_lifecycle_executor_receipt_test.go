package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
)

type agentTaskContractResponse struct {
	Blockers         []string                      `json:"blockers"`
	MissingContracts []string                      `json:"missing_contracts"`
	Data             model.FindXAgentExecutionTask `json:"data"`
}

func TestFindXAgentTaskExecutorReceiptMatrixMissingRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name     string
		metadata map[string]string
		wantRefs []string
	}{
		{"local linux", localTaskMetadata("linux"), []string{"local_executor_ref", "linux_installer_ref", "service_receipt_ref", "data_arrival_validator_ref"}},
		{"local windows", localTaskMetadata("windows"), []string{"local_executor_ref", "windows_installer_ref", "service_receipt_ref", "data_arrival_validator_ref"}},
		{"ssh", transportTaskMetadata("ssh"), []string{"ssh_runner_ref", "ssh_host_key", "ssh_fingerprint", "remote_executor_ref"}},
		{"winrm", transportTaskMetadata("winrm"), []string{"winrm_endpoint_ref", "winrm_transport_ref", "remote_executor_ref"}},
		{"systemd", markerTaskMetadata("systemd"), []string{"systemd_unit_ref", "systemd_receipt_ref"}},
		{"windows service", markerTaskMetadata("windows-service"), []string{"windows_service_ref", "windows_service_receipt_ref"}},
		{"iis", markerTaskMetadata("iis"), []string{"iis_site_ref", "iis_app_pool_ref", "iis_receipt_ref"}},
		{"docker", markerTaskMetadata("docker"), []string{"container_ref", "image_ref", "docker_receipt_ref"}},
		{"operator", kubernetesWorkloadMetadata("operator"), []string{"operator_ref", "crd_ref", "controller_receipt_ref"}},
		{"daemonset", kubernetesWorkloadMetadata("daemonset"), []string{"daemonset_ref"}},
		{"sidecar", kubernetesWorkloadMetadata("sidecar"), []string{"sidecar_injection_ref"}},
		{"initcontainer", kubernetesWorkloadMetadata("initcontainer"), []string{"init_container_ref"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			payload := postAgentTaskForContract(t, tt.metadata)
			if payload.Data.Status != "blocked" || !containsLifecycleTestString(payload.Blockers, "MISSING_CONTRACTS") {
				t.Fatalf("%s should be blocked by missing contracts, got %#v", tt.name, payload)
			}
			for _, ref := range tt.wantRefs {
				if !containsLifecycleTestString(payload.MissingContracts, ref) || !strings.Contains(payload.Data.Blocker, ref) {
					t.Fatalf("%s should require %s, missing=%v blocker=%s", tt.name, ref, payload.MissingContracts, payload.Data.Blocker)
				}
			}
			assertNoCredentialEcho(t, mustMarshalAgentTaskContractResponse(t, payload))
		})
	}
}

func TestFindXAgentTaskExecutorReceiptMatrixCompleteRefsStillBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name     string
		metadata map[string]string
	}{
		{"local linux", completeLocalTaskMetadata("linux")},
		{"local windows", completeLocalTaskMetadata("windows")},
		{"ssh", completeTransportTaskMetadata("ssh")},
		{"winrm", completeTransportTaskMetadata("winrm")},
		{"systemd", completeMarkerTaskMetadata("systemd")},
		{"windows service", completeMarkerTaskMetadata("windows-service")},
		{"iis", completeMarkerTaskMetadata("iis")},
		{"docker", completeMarkerTaskMetadata("docker")},
		{"operator", completeKubernetesWorkloadMetadata("operator")},
		{"daemonset", completeKubernetesWorkloadMetadata("daemonset")},
		{"sidecar", completeKubernetesWorkloadMetadata("sidecar")},
		{"initcontainer", completeKubernetesWorkloadMetadata("initcontainer")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			payload := postAgentTaskForContract(t, tt.metadata)
			body := mustMarshalAgentTaskContractResponse(t, payload)
			if payload.Data.Status != "blocked" || payload.Data.Blocker != "PENDING: executor not enabled / execution protocol not open" {
				t.Fatalf("%s complete refs should stay executor-blocked, got %#v", tt.name, payload.Data)
			}
			if !containsLifecycleTestString(payload.Blockers, "EXECUTOR_DISABLED_BY_CONTRACT") || len(payload.MissingContracts) != 0 {
				t.Fatalf("%s should expose executor-disabled blocker only, got blockers=%v missing=%v", tt.name, payload.Blockers, payload.MissingContracts)
			}
			for _, forbidden := range []string{"queued", "running", "succeeded", "success", "applied", "rolled-back"} {
				if strings.Contains(body, forbidden) {
					t.Fatalf("%s must not fake execution state %s: %s", tt.name, forbidden, body)
				}
			}
			assertNoCredentialEcho(t, body)
		})
	}
}

func TestFindXAgentTaskExecutorReceiptSensitiveMetadataNotEchoed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	metadata := completeTransportTaskMetadata("ssh")
	metadata["api_token"] = "token-value"
	metadata["password"] = "password-value"
	metadata["cookie"] = "cookie-value"
	metadata["session"] = "session-value"
	metadata["private_key"] = "-----BEGIN PRIVATE KEY-----"
	metadata["ticket"] = "CHG-104"
	payload := postAgentTaskForContract(t, metadata)
	if payload.Data.Metadata["ticket"] != "CHG-104" {
		t.Fatalf("safe metadata should be retained, got %#v", payload.Data.Metadata)
	}
	for _, key := range []string{"api_token", "password", "cookie", "session", "private_key"} {
		if payload.Data.Metadata[key] != "" {
			t.Fatalf("sensitive metadata key %s should be dropped, got %#v", key, payload.Data.Metadata)
		}
	}
	assertNoCredentialEcho(t, mustMarshalAgentTaskContractResponse(t, payload))
}

func postAgentTaskForContract(t *testing.T, metadata map[string]string) agentTaskContractResponse {
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
	var payload agentTaskContractResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode task response: %v body=%s", err, w.Body.String())
	}
	return payload
}

func localTaskMetadata(targetOS string) map[string]string {
	return baseAgentTaskMetadata(targetOS)
}

func transportTaskMetadata(transport string) map[string]string {
	metadata := baseAgentTaskMetadata("linux")
	metadata["transport"] = transport
	return metadata
}

func markerTaskMetadata(marker string) map[string]string {
	metadata := transportTaskMetadata("ssh")
	metadata["service_type"] = marker
	return metadata
}

func kubernetesWorkloadMetadata(workload string) map[string]string {
	metadata := completeKubernetesBaseTaskMetadata()
	metadata["workload_kind"] = workload
	return metadata
}

func completeLocalTaskMetadata(targetOS string) map[string]string {
	metadata := localTaskMetadata(targetOS)
	metadata["data_arrival_validator_ref"] = "validator-1"
	metadata["local_executor_ref"] = "local-executor-1"
	metadata["service_receipt_ref"] = "service-receipt-1"
	if targetOS == "windows" {
		metadata["windows_installer_ref"] = "windows-installer-1"
	} else {
		metadata["linux_installer_ref"] = "linux-installer-1"
	}
	return metadata
}

func completeTransportTaskMetadata(transport string) map[string]string {
	metadata := transportTaskMetadata(transport)
	metadata["data_arrival_validator_ref"] = "validator-1"
	metadata["remote_executor_ref"] = "remote-executor-1"
	if transport == "winrm" {
		metadata["winrm_endpoint_ref"] = "winrm-endpoint-1"
		metadata["winrm_transport_ref"] = "winrm-transport-1"
		return metadata
	}
	metadata["ssh_runner_ref"] = "ssh-runner-1"
	metadata["ssh_host_key"] = "host-key-1"
	metadata["ssh_fingerprint"] = "fingerprint-1"
	return metadata
}

func completeMarkerTaskMetadata(marker string) map[string]string {
	metadata := completeTransportTaskMetadata("ssh")
	metadata["service_type"] = marker
	for key, value := range map[string]string{
		"systemd_unit_ref":            "systemd-unit-1",
		"systemd_receipt_ref":         "systemd-receipt-1",
		"windows_service_ref":         "windows-service-1",
		"windows_service_receipt_ref": "windows-service-receipt-1",
		"iis_site_ref":                "iis-site-1",
		"iis_app_pool_ref":            "iis-pool-1",
		"iis_receipt_ref":             "iis-receipt-1",
		"container_ref":               "container-1",
		"image_ref":                   "image-1",
		"docker_receipt_ref":          "docker-receipt-1",
	} {
		metadata[key] = value
	}
	return metadata
}

func completeKubernetesWorkloadMetadata(workload string) map[string]string {
	metadata := kubernetesWorkloadMetadata(workload)
	for key, value := range map[string]string{
		"operator_ref":           "operator-1",
		"crd_ref":                "crd-1",
		"controller_receipt_ref": "controller-receipt-1",
		"daemonset_ref":          "daemonset-1",
		"sidecar_injection_ref":  "sidecar-injection-1",
		"init_container_ref":     "init-container-1",
	} {
		metadata[key] = value
	}
	return metadata
}

func baseAgentTaskMetadata(targetOS string) map[string]string {
	return map[string]string{
		"target_os":             targetOS,
		"service_ref":           "svc-1",
		"executor_ref":          "executor-1",
		"idempotency_key":       "idem-1",
		"timeout_policy_ref":    "timeout-1",
		"execution_receipt_ref": "receipt-1",
		"audit_ref":             "audit-1",
	}
}

func completeKubernetesBaseTaskMetadata() map[string]string {
	metadata := completeTransportTaskMetadata("ssh")
	for key, value := range map[string]string{
		"target_os":                  "k8s",
		"cluster_ref":                "cluster-1",
		"namespace_ref":              "namespace-1",
		"workload_selector_ref":      "workload-1",
		"rbac_ref":                   "rbac-1",
		"service_account_ref":        "service-account-1",
		"rollout_strategy_ref":       "rollout-strategy-1",
		"rollout_receipt_ref":        "rollout-receipt-1",
		"restart_strategy_ref":       "restart-strategy-1",
		"data_arrival_validator_ref": "validator-1",
	} {
		metadata[key] = value
	}
	return metadata
}

func mustMarshalAgentTaskContractResponse(t *testing.T, payload agentTaskContractResponse) string {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal task response: %v", err)
	}
	return string(raw)
}
