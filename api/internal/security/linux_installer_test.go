package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxInstallerPrerequisitesRequireArtifactSafetyAndHostKey(t *testing.T) {
	result := EvaluateLinuxInstallerPrerequisites(LinuxInstallerPrerequisites{
		Runner:           "ssh",
		SafetyPolicyPath: filepath.Join(t.TempDir(), "missing-policy.json"),
	})
	if result.Allowed || result.Status != "blocked" {
		t.Fatalf("expected blocked prerequisites, got %#v", result)
	}
	for _, want := range []string{
		"package_repository_ref",
		"signature_ref",
		"checksum",
		"script_manifest_ref",
		"executor_ref",
		"systemd_unit_ref",
		"systemd_unit_name_ref_or_systemd_unit_path_ref",
		"systemd_mode",
		"curl_installer_ref",
		"curl_command_ref",
		"curl_manifest_ref",
		"env_template_ref",
		"service_receipt_ref",
		"heartbeat_validator_ref",
		"data_arrival_validator_ref",
		"audit_ref_or_evidence_chain_ref",
		"runner_whitelist_ref",
		"reload_receipt_ref",
		"service_status_receipt_ref",
		"linux_installer_safety_policy",
		"ssh_host_key_or_fingerprint",
	} {
		if !strings.Contains(result.Reason, want) {
			t.Fatalf("expected blocker %s in %q", want, result.Reason)
		}
	}
}

func TestLinuxInstallerPrerequisitesStillBlockWhenExecutorDisabled(t *testing.T) {
	ensureDefaultLinuxInstallerSafetyPolicy(t)

	result := EvaluateLinuxInstallerPrerequisites(LinuxInstallerPrerequisites{
		PackageRepositoryRef:    "repo-ref",
		SignatureRef:            "sig-ref",
		Checksum:                "sha256:abc",
		ScriptManifestRef:       "script-manifest-ref",
		ExecutorRef:             "executor-ref",
		SafetyPolicyPath:        defaultLinuxInstallerSafetyPolicyPath,
		Runner:                  "ssh",
		SSHFingerprint:          "SHA256:example",
		SystemdUnitRef:          "systemd-ref",
		SystemdUnitNameRef:      "unit-name-ref",
		SystemdMode:             "system",
		CurlInstallerRef:        "curl-ref",
		CurlCommandRef:          "curl-command-ref",
		CurlManifestRef:         "curl-manifest-ref",
		EnvTemplateRef:          "env-ref",
		ServiceReceiptRef:       "service-receipt-ref",
		HeartbeatValidatorRef:   "heartbeat-validator-ref",
		DataArrivalValidatorRef: "data-arrival-validator-ref",
		EvidenceChainRef:        "evidence-chain-ref",
		RunnerWhitelistRef:      "runner-whitelist-ref",
		ReloadReceiptRef:        "reload-receipt-ref",
		ServiceStatusReceiptRef: "service-status-receipt-ref",
	})
	if result.Allowed || result.Status != "blocked" {
		t.Fatalf("executor should remain blocked, got %#v", result)
	}
	if !strings.Contains(result.Reason, "executor is not enabled") {
		t.Fatalf("expected executor disabled blocker, got %q", result.Reason)
	}
}

func TestLinuxInstallerPrerequisitesLocalRunnerDoesNotRequireSSHHostKey(t *testing.T) {
	ensureDefaultLinuxInstallerSafetyPolicy(t)

	result := EvaluateLinuxInstallerPrerequisites(LinuxInstallerPrerequisites{
		PackageRepositoryRef:    "repo-ref",
		SignatureRef:            "sig-ref",
		Checksum:                "sha256:abc",
		ScriptManifestRef:       "script-manifest-ref",
		ExecutorRef:             "executor-ref",
		SafetyPolicyPath:        defaultLinuxInstallerSafetyPolicyPath,
		Runner:                  "local",
		SystemdUnitRef:          "systemd-ref",
		SystemdUnitPathRef:      "unit-path-ref",
		SystemdMode:             "user",
		CurlInstallerRef:        "curl-ref",
		CurlCommandRef:          "curl-command-ref",
		CurlManifestRef:         "curl-manifest-ref",
		EnvTemplateRef:          "env-ref",
		ServiceReceiptRef:       "service-receipt-ref",
		HeartbeatValidatorRef:   "heartbeat-validator-ref",
		DataArrivalValidatorRef: "data-arrival-validator-ref",
		AuditRef:                "audit-ref",
		RunnerWhitelistRef:      "runner-whitelist-ref",
		ReloadReceiptRef:        "reload-receipt-ref",
		ServiceStatusReceiptRef: "service-status-receipt-ref",
	})
	if strings.Contains(result.Reason, "ssh_host_key_or_fingerprint") {
		t.Fatalf("local runner should not require ssh host key, got %q", result.Reason)
	}
}

func TestLinuxInstallerPrerequisitesSystemdRunnerRequiresSafeModeAndUnitIdentity(t *testing.T) {
	ensureDefaultLinuxInstallerSafetyPolicy(t)

	result := EvaluateLinuxInstallerPrerequisites(LinuxInstallerPrerequisites{
		PackageRepositoryRef:    "repo-ref",
		SignatureRef:            "sig-ref",
		Checksum:                "sha256:abc",
		ScriptManifestRef:       "script-manifest-ref",
		ExecutorRef:             "executor-ref",
		SafetyPolicyPath:        defaultLinuxInstallerSafetyPolicyPath,
		Runner:                  "local-systemd",
		SystemdUnitRef:          "systemd-ref",
		SystemdUnitNameRef:      "unit-name-ref",
		SystemdMode:             "system",
		CurlInstallerRef:        "curl-ref",
		CurlCommandRef:          "curl-command-ref",
		CurlManifestRef:         "curl-manifest-ref",
		EnvTemplateRef:          "env-ref",
		ServiceReceiptRef:       "service-receipt-ref",
		HeartbeatValidatorRef:   "heartbeat-validator-ref",
		DataArrivalValidatorRef: "data-arrival-validator-ref",
		EvidenceChainRef:        "evidence-chain-ref",
		RunnerWhitelistRef:      "runner-whitelist-ref",
		ReloadReceiptRef:        "reload-receipt-ref",
		ServiceStatusReceiptRef: "service-status-receipt-ref",
	})
	for _, want := range []string{"systemd_mode", "systemd_unit_name_ref_or_systemd_unit_path_ref", "ssh_host_key_or_fingerprint"} {
		if strings.Contains(result.Reason, want) {
			t.Fatalf("systemd local runner should not report %s, got %q", want, result.Reason)
		}
	}
	if !strings.Contains(result.Reason, "executor is not enabled") {
		t.Fatalf("complete systemd preflight must remain executor-blocked, got %q", result.Reason)
	}
}

func TestLinuxInstallerPrerequisitesSystemdRunnerRequiresModeAndUnitIdentity(t *testing.T) {
	ensureDefaultLinuxInstallerSafetyPolicy(t)

	result := EvaluateLinuxInstallerPrerequisites(LinuxInstallerPrerequisites{
		PackageRepositoryRef:    "repo-ref",
		SignatureRef:            "sig-ref",
		Checksum:                "sha256:abc",
		ScriptManifestRef:       "script-manifest-ref",
		ExecutorRef:             "executor-ref",
		SafetyPolicyPath:        defaultLinuxInstallerSafetyPolicyPath,
		Runner:                  "local-systemd",
		SystemdUnitRef:          "systemd-ref",
		CurlInstallerRef:        "curl-ref",
		CurlCommandRef:          "curl-command-ref",
		CurlManifestRef:         "curl-manifest-ref",
		EnvTemplateRef:          "env-ref",
		ServiceReceiptRef:       "service-receipt-ref",
		HeartbeatValidatorRef:   "heartbeat-validator-ref",
		DataArrivalValidatorRef: "data-arrival-validator-ref",
		EvidenceChainRef:        "evidence-chain-ref",
		RunnerWhitelistRef:      "runner-whitelist-ref",
		ReloadReceiptRef:        "reload-receipt-ref",
		ServiceStatusReceiptRef: "service-status-receipt-ref",
	})
	for _, want := range []string{"systemd_mode", "systemd_unit_name_ref_or_systemd_unit_path_ref"} {
		if !strings.Contains(result.Reason, want) {
			t.Fatalf("systemd local runner should require %s, got %q", want, result.Reason)
		}
	}
	if strings.Contains(result.Reason, "ssh_host_key_or_fingerprint") {
		t.Fatalf("systemd local runner should not require ssh host key, got %q", result.Reason)
	}
}

func ensureDefaultLinuxInstallerSafetyPolicy(t *testing.T) {
	t.Helper()

	path := filepath.Join("..", "..", "..", defaultLinuxInstallerSafetyPolicyPath)
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
