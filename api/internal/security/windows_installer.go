package security

import (
	"sort"
	"strings"
)

type WindowsInstallerPrerequisites struct {
	PackageRepositoryRef    string
	SignatureRef            string
	Checksum                string
	ScriptManifestRef       string
	ExecutorRef             string
	WindowsInstallerRef     string
	PowerShellInstallerRef  string
	CertutilInstallerRef    string
	WindowsCmdInstallerRef  string
	ServiceManifestRef      string
	InstallRootPolicyRef    string
	WindowsServiceNameRef   string
	WindowsServicePolicyRef string
	ServiceReceiptRef       string
	ServiceStatusReceiptRef string
	InstallReceiptRef       string
	UninstallManifestRef    string
	UninstallReceiptRef     string
	RollbackManifestRef     string
	RollbackReceiptRef      string
	HeartbeatValidatorRef   string
	DataArrivalValidatorRef string
	AuditRef                string
	EvidenceChainRef        string
	ReceiverEndpointRef     string
	Method                  string
}

type WindowsInstallerGateResult struct {
	Allowed bool
	Status  string
	Reason  string
	Runner  string
}

func EvaluateWindowsInstallerPrerequisites(input WindowsInstallerPrerequisites) WindowsInstallerGateResult {
	runner := normalizeWindowsInstallerRunner(input.Method)
	missing := windowsInstallerMissingPrerequisites(input)
	if len(missing) > 0 {
		return WindowsInstallerGateResult{
			Allowed: false,
			Status:  "blocked",
			Reason:  "BLOCKED_BY_CONTRACT: Windows installer prerequisites missing: " + strings.Join(missing, ", "),
			Runner:  runner,
		}
	}
	return WindowsInstallerGateResult{
		Allowed: false,
		Status:  "blocked",
		Reason:  "BLOCKED_BY_CONTRACT: Windows executor/service lifecycle protocol not enabled",
		Runner:  runner,
	}
}

func windowsInstallerMissingPrerequisites(input WindowsInstallerPrerequisites) []string {
	required := []struct {
		name  string
		value string
	}{
		{"package_repository_ref", input.PackageRepositoryRef},
		{"signature_ref", input.SignatureRef},
		{"checksum", input.Checksum},
		{"script_manifest_ref", input.ScriptManifestRef},
		{"executor_ref", input.ExecutorRef},
		{"windows_installer_ref", input.WindowsInstallerRef},
		{"service_manifest_ref", input.ServiceManifestRef},
		{"install_root_policy_ref", input.InstallRootPolicyRef},
		{"windows_service_name_ref", input.WindowsServiceNameRef},
		{"windows_service_policy_ref", input.WindowsServicePolicyRef},
		{"service_receipt_ref", input.ServiceReceiptRef},
		{"service_status_receipt_ref", input.ServiceStatusReceiptRef},
		{"install_receipt_ref", input.InstallReceiptRef},
		{"rollback_manifest_ref", input.RollbackManifestRef},
		{"rollback_receipt_ref", input.RollbackReceiptRef},
		{"uninstall_manifest_ref", input.UninstallManifestRef},
		{"uninstall_receipt_ref", input.UninstallReceiptRef},
		{"heartbeat_validator_ref", input.HeartbeatValidatorRef},
		{"data_arrival_validator_ref", input.DataArrivalValidatorRef},
		{"receiver_endpoint_ref", input.ReceiverEndpointRef},
	}
	missing := []string{}
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			missing = append(missing, item.name)
		}
	}
	runner := normalizeWindowsInstallerRunner(input.Method)
	if runner == "windows-powershell" && strings.TrimSpace(input.PowerShellInstallerRef) == "" {
		missing = append(missing, "powershell_installer_ref")
	}
	if runner == "windows-cmd" && strings.TrimSpace(input.CertutilInstallerRef) == "" &&
		strings.TrimSpace(input.WindowsCmdInstallerRef) == "" {
		missing = append(missing, "certutil_installer_ref_or_windows_cmd_installer_ref")
	}
	if strings.TrimSpace(input.AuditRef) == "" && strings.TrimSpace(input.EvidenceChainRef) == "" {
		missing = append(missing, "audit_ref_or_evidence_chain_ref")
	}
	sort.Strings(missing)
	return missing
}

func normalizeWindowsInstallerRunner(method string) string {
	clean := strings.ToLower(strings.TrimSpace(method))
	switch {
	case strings.Contains(clean, "powershell"), strings.Contains(clean, "invoke-webrequest"):
		return "windows-powershell"
	case strings.Contains(clean, "cmd"), strings.Contains(clean, "certutil"):
		return "windows-cmd"
	default:
		return "windows-installer"
	}
}
