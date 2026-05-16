package security

import (
	"strings"
	"testing"
)

func TestRemoteInstallerPreflightSSHRequiresRemoteContractsAndHostIdentity(t *testing.T) {
	result := EvaluateRemoteInstallerPreflight(RemoteInstallerPreflightInput{Scope: "ssh"})
	if result.Allowed || result.Status != "blocked" || result.Scope != "ssh" || result.Runner != "ssh" {
		t.Fatalf("expected blocked ssh preflight, got %#v", result)
	}
	for _, want := range []string{
		"credential_ref",
		"ssh_runner_ref",
		"ssh_host_key_or_fingerprint",
		"remote_executor_ref",
		"timeout_policy_ref",
		"idempotency_key",
		"install_receipt_ref",
		"execution_receipt_ref",
		"service_receipt_ref",
		"heartbeat_validator_ref",
		"data_arrival_validator_ref",
		"audit_ref_or_evidence_chain_ref",
		"package_repository_ref",
		"signature_ref",
		"checksum",
		"script_manifest_ref",
	} {
		if !strings.Contains(result.Reason, want) {
			t.Fatalf("expected missing %s in %q", want, result.Reason)
		}
	}
}

func TestRemoteInstallerPreflightWinRMRequiresTransportContracts(t *testing.T) {
	result := EvaluateRemoteInstallerPreflight(RemoteInstallerPreflightInput{Scope: "winrm"})
	if result.Allowed || result.Status != "blocked" || result.Scope != "winrm" || result.Runner != "winrm" {
		t.Fatalf("expected blocked winrm preflight, got %#v", result)
	}
	for _, want := range []string{"credential_ref", "winrm_endpoint_ref", "winrm_transport_ref", "remote_executor_ref", "timeout_policy_ref", "idempotency_key"} {
		if !strings.Contains(result.Reason, want) {
			t.Fatalf("expected missing %s in %q", want, result.Reason)
		}
	}
	if strings.Contains(result.Reason, "ssh_runner_ref") || strings.Contains(result.Reason, "ssh_host_key_or_fingerprint") {
		t.Fatalf("winrm preflight should not require ssh refs, got %q", result.Reason)
	}
}

func TestRemoteInstallerPreflightCompleteRefsStillBlocksExecutor(t *testing.T) {
	for _, scope := range []string{"ssh", "winrm"} {
		t.Run(scope, func(t *testing.T) {
			input := completeRemoteInstallerPreflightInput(scope)
			result := EvaluateRemoteInstallerPreflight(input)
			if result.Allowed || result.Status != "blocked" {
				t.Fatalf("complete %s refs must remain blocked, got %#v", scope, result)
			}
			want := "PENDING: " + scope + " remote executor contract is not enabled"
			if result.Reason != want {
				t.Fatalf("expected executor blocker %q, got %q", want, result.Reason)
			}
		})
	}
}

func completeRemoteInstallerPreflightInput(scope string) RemoteInstallerPreflightInput {
	input := RemoteInstallerPreflightInput{
		Scope:                   scope,
		CredentialRef:           "credential-ref",
		RemoteExecutorRef:       "remote-executor-ref",
		TimeoutPolicyRef:        "timeout-ref",
		IdempotencyKey:          "idem-1",
		InstallReceiptRef:       "install-receipt-ref",
		ExecutionReceiptRef:     "execution-receipt-ref",
		ServiceReceiptRef:       "service-receipt-ref",
		HeartbeatValidatorRef:   "heartbeat-validator-ref",
		DataArrivalValidatorRef: "data-arrival-validator-ref",
		EvidenceChainRef:        "evidence-chain-ref",
		PackageRepositoryRef:    "repo-ref",
		SignatureRef:            "signature-ref",
		Checksum:                "sha256:abc",
		ScriptManifestRef:       "script-manifest-ref",
	}
	if scope == "ssh" {
		input.SSHRunnerRef = "ssh-runner-ref"
		input.SSHFingerprint = "SHA256:example"
		return input
	}
	input.WinRMEndpointRef = "winrm-endpoint-ref"
	input.WinRMTransportRef = "winrm-transport-ref"
	return input
}
