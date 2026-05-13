package security

import (
	"strings"
	"testing"
)

func TestWindowsInstallerPrerequisitesRequireArtifactServiceAndReceiver(t *testing.T) {
	result := EvaluateWindowsInstallerPrerequisites(WindowsInstallerPrerequisites{
		Method: "windows-powershell",
	})
	if result.Allowed || result.Status != "blocked" || result.Runner != "windows-powershell" {
		t.Fatalf("expected blocked Windows prerequisites, got %#v", result)
	}
	for _, want := range []string{
		"package_repository_ref",
		"signature_ref",
		"checksum",
		"script_manifest_ref",
		"executor_ref",
		"windows_installer_ref",
		"service_manifest_ref",
		"install_root_policy_ref",
		"rollback_manifest_ref",
		"uninstall_manifest_ref",
		"receiver_endpoint_ref",
	} {
		if !strings.Contains(result.Reason, want) {
			t.Fatalf("expected blocker %s in %q", want, result.Reason)
		}
	}
}

func TestWindowsInstallerPrerequisitesRequirePowerShellAndCertutilRefs(t *testing.T) {
	powershell := EvaluateWindowsInstallerPrerequisites(WindowsInstallerPrerequisites{
		PackageRepositoryRef:    "repo-ref",
		SignatureRef:            "sig-ref",
		Checksum:                "sha256:abc",
		ScriptManifestRef:       "script-manifest-ref",
		ExecutorRef:             "executor-ref",
		WindowsInstallerRef:     "installer-ref",
		ServiceManifestRef:      "service-ref",
		InstallRootPolicyRef:    "root-policy-ref",
		WindowsServiceNameRef:   "service-name-ref",
		WindowsServicePolicyRef: "service-policy-ref",
		ServiceReceiptRef:       "service-receipt-ref",
		ServiceStatusReceiptRef: "service-status-receipt-ref",
		InstallReceiptRef:       "install-receipt-ref",
		UninstallManifestRef:    "uninstall-manifest-ref",
		UninstallReceiptRef:     "uninstall-receipt-ref",
		RollbackManifestRef:     "rollback-manifest-ref",
		RollbackReceiptRef:      "rollback-receipt-ref",
		HeartbeatValidatorRef:   "heartbeat-validator-ref",
		DataArrivalValidatorRef: "data-arrival-validator-ref",
		AuditRef:                "audit-ref",
		EvidenceChainRef:        "evidence-chain-ref",
		ReceiverEndpointRef:     "receiver-ref",
		Method:                  "Invoke-WebRequest -UseBasicParsing",
	})
	if powershell.Allowed || powershell.Status != "blocked" || powershell.Runner != "windows-powershell" {
		t.Fatalf("expected blocked PowerShell gate, got %#v", powershell)
	}
	if !strings.Contains(powershell.Reason, "powershell_installer_ref") {
		t.Fatalf("expected missing PowerShell installer ref, got %q", powershell.Reason)
	}

	certutil := EvaluateWindowsInstallerPrerequisites(WindowsInstallerPrerequisites{
		PackageRepositoryRef:    "repo-ref",
		SignatureRef:            "sig-ref",
		Checksum:                "sha256:abc",
		ScriptManifestRef:       "script-manifest-ref",
		ExecutorRef:             "executor-ref",
		WindowsInstallerRef:     "installer-ref",
		PowerShellInstallerRef:  "powershell-installer-ref",
		ServiceManifestRef:      "service-ref",
		InstallRootPolicyRef:    "root-policy-ref",
		WindowsServiceNameRef:   "service-name-ref",
		WindowsServicePolicyRef: "service-policy-ref",
		ServiceReceiptRef:       "service-receipt-ref",
		ServiceStatusReceiptRef: "service-status-receipt-ref",
		InstallReceiptRef:       "install-receipt-ref",
		UninstallManifestRef:    "uninstall-manifest-ref",
		UninstallReceiptRef:     "uninstall-receipt-ref",
		RollbackManifestRef:     "rollback-manifest-ref",
		RollbackReceiptRef:      "rollback-receipt-ref",
		HeartbeatValidatorRef:   "heartbeat-validator-ref",
		DataArrivalValidatorRef: "data-arrival-validator-ref",
		AuditRef:                "audit-ref",
		EvidenceChainRef:        "evidence-chain-ref",
		ReceiverEndpointRef:     "receiver-ref",
		Method:                  "CERTUTIL.EXE",
	})
	if certutil.Allowed || certutil.Status != "blocked" || certutil.Runner != "windows-cmd" {
		t.Fatalf("expected blocked certutil gate, got %#v", certutil)
	}
	if !strings.Contains(certutil.Reason, "certutil_installer_ref_or_windows_cmd_installer_ref") {
		t.Fatalf("expected missing certutil/windows-cmd ref, got %q", certutil.Reason)
	}
}

func TestWindowsInstallerPrerequisitesStillBlockWhenExecutorDisabled(t *testing.T) {
	result := EvaluateWindowsInstallerPrerequisites(WindowsInstallerPrerequisites{
		PackageRepositoryRef:    "repo-ref",
		SignatureRef:            "sig-ref",
		Checksum:                "sha256:abc",
		ScriptManifestRef:       "script-manifest-ref",
		ExecutorRef:             "executor-ref",
		WindowsInstallerRef:     "installer-ref",
		PowerShellInstallerRef:  "powershell-installer-ref",
		CertutilInstallerRef:    "certutil-installer-ref",
		WindowsCmdInstallerRef:  "windows-cmd-installer-ref",
		ServiceManifestRef:      "service-ref",
		InstallRootPolicyRef:    "root-policy-ref",
		WindowsServiceNameRef:   "service-name-ref",
		WindowsServicePolicyRef: "service-policy-ref",
		ServiceReceiptRef:       "service-receipt-ref",
		ServiceStatusReceiptRef: "service-status-receipt-ref",
		InstallReceiptRef:       "install-receipt-ref",
		UninstallManifestRef:    "uninstall-ref",
		UninstallReceiptRef:     "uninstall-receipt-ref",
		RollbackManifestRef:     "rollback-ref",
		RollbackReceiptRef:      "rollback-receipt-ref",
		HeartbeatValidatorRef:   "heartbeat-validator-ref",
		DataArrivalValidatorRef: "data-arrival-validator-ref",
		AuditRef:                "audit-ref",
		EvidenceChainRef:        "evidence-chain-ref",
		ReceiverEndpointRef:     "receiver-ref",
		Method:                  "windows-cmd",
	})
	if result.Allowed || result.Status != "blocked" || result.Runner != "windows-cmd" {
		t.Fatalf("executor should remain blocked, got %#v", result)
	}
	if result.Reason != "BLOCKED_BY_CONTRACT: Windows executor/service lifecycle protocol not enabled" {
		t.Fatalf("expected executor disabled blocker, got %q", result.Reason)
	}
}

func TestWindowsInstallerPrerequisitesDoNotEchoSensitiveValues(t *testing.T) {
	result := EvaluateWindowsInstallerPrerequisites(WindowsInstallerPrerequisites{
		PackageRepositoryRef:    "repo-ref",
		SignatureRef:            "sig-ref",
		Checksum:                "sha256:abc",
		ScriptManifestRef:       "script-manifest-ref",
		ExecutorRef:             "executor-ref",
		WindowsInstallerRef:     "installer-ref",
		PowerShellInstallerRef:  "powershell-installer-ref",
		CertutilInstallerRef:    "certutil-installer-ref",
		WindowsCmdInstallerRef:  "windows-cmd-installer-ref",
		ServiceManifestRef:      "service-ref",
		InstallRootPolicyRef:    "root-policy-ref",
		WindowsServiceNameRef:   "service-name-ref",
		WindowsServicePolicyRef: "service-policy-ref",
		ServiceReceiptRef:       "service-receipt-ref",
		ServiceStatusReceiptRef: "service-status-receipt-ref",
		InstallReceiptRef:       "install-receipt-ref",
		UninstallManifestRef:    "uninstall-ref",
		UninstallReceiptRef:     "uninstall-receipt-ref",
		RollbackManifestRef:     "rollback-ref",
		RollbackReceiptRef:      "rollback-receipt-ref",
		HeartbeatValidatorRef:   "heartbeat-validator-ref",
		DataArrivalValidatorRef: "data-arrival-validator-ref",
		AuditRef:                "audit-ref",
		EvidenceChainRef:        "evidence-chain-ref",
		ReceiverEndpointRef:     "https://example.invalid?token=secret",
		Method:                  "PowerShell",
	})
	if strings.Contains(result.Reason, "secret") || strings.Contains(result.Reason, "token") {
		t.Fatalf("reason must not echo sensitive metadata values: %q", result.Reason)
	}
}
