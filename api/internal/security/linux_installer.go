package security

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const defaultLinuxInstallerSafetyPolicyPath = "scripts/testing/safety_policy.json"

type LinuxInstallerPrerequisites struct {
	PackageRepositoryRef    string
	SignatureRef            string
	Checksum                string
	ScriptManifestRef       string
	ExecutorRef             string
	SafetyPolicyPath        string
	Runner                  string
	SSHHostKey              string
	SSHFingerprint          string
	SystemdUnitRef          string
	SystemdUnitNameRef      string
	SystemdUnitPathRef      string
	SystemdMode             string
	CurlInstallerRef        string
	CurlCommandRef          string
	CurlManifestRef         string
	EnvTemplateRef          string
	ServiceReceiptRef       string
	HeartbeatValidatorRef   string
	DataArrivalValidatorRef string
	AuditRef                string
	EvidenceChainRef        string
	RunnerWhitelistRef      string
	ReloadReceiptRef        string
	ServiceStatusReceiptRef string
}

type LinuxInstallerGateResult struct {
	Allowed bool
	Status  string
	Reason  string
	Runner  string
}

func EvaluateLinuxInstallerPrerequisites(input LinuxInstallerPrerequisites) LinuxInstallerGateResult {
	runner := normalizeInstallerRunner(input.Runner)
	missing := linuxInstallerMissingPrerequisites(input, runner)
	if len(missing) > 0 {
		return LinuxInstallerGateResult{
			Allowed: false,
			Status:  "blocked",
			Reason:  "PENDING: Linux installer prerequisites missing: " + strings.Join(missing, ", "),
			Runner:  runner,
		}
	}
	return LinuxInstallerGateResult{
		Allowed: false,
		Status:  "blocked",
		Reason:  "PENDING: Linux installer executor is not enabled; execution record captured only",
		Runner:  runner,
	}
}

func linuxInstallerMissingPrerequisites(input LinuxInstallerPrerequisites, runner string) []string {
	required := []struct {
		name  string
		value string
	}{
		{"package_repository_ref", input.PackageRepositoryRef},
		{"signature_ref", input.SignatureRef},
		{"checksum", input.Checksum},
		{"script_manifest_ref", input.ScriptManifestRef},
		{"executor_ref", input.ExecutorRef},
		{"systemd_unit_ref", input.SystemdUnitRef},
		{"curl_installer_ref", input.CurlInstallerRef},
		{"curl_command_ref", input.CurlCommandRef},
		{"curl_manifest_ref", input.CurlManifestRef},
		{"env_template_ref", input.EnvTemplateRef},
		{"service_receipt_ref", input.ServiceReceiptRef},
		{"heartbeat_validator_ref", input.HeartbeatValidatorRef},
		{"data_arrival_validator_ref", input.DataArrivalValidatorRef},
		{"runner_whitelist_ref", input.RunnerWhitelistRef},
		{"reload_receipt_ref", input.ReloadReceiptRef},
		{"service_status_receipt_ref", input.ServiceStatusReceiptRef},
	}
	missing := []string{}
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			missing = append(missing, item.name)
		}
	}
	if strings.TrimSpace(input.SystemdUnitNameRef) == "" && strings.TrimSpace(input.SystemdUnitPathRef) == "" {
		missing = append(missing, "systemd_unit_name_ref_or_systemd_unit_path_ref")
	}
	if !isSafeLinuxInstallerRunner(runner) {
		missing = append(missing, "runner_whitelist")
	}
	if !isSafeSystemdMode(input.SystemdMode) {
		missing = append(missing, "systemd_mode")
	}
	if strings.TrimSpace(input.AuditRef) == "" && strings.TrimSpace(input.EvidenceChainRef) == "" {
		missing = append(missing, "audit_ref_or_evidence_chain_ref")
	}
	if !safetyPolicyExists(input.SafetyPolicyPath) {
		missing = append(missing, "linux_installer_safety_policy")
	}
	if runner == "ssh" && strings.TrimSpace(input.SSHHostKey) == "" && strings.TrimSpace(input.SSHFingerprint) == "" {
		missing = append(missing, "ssh_host_key_or_fingerprint")
	}
	sort.Strings(missing)
	return missing
}

func normalizeInstallerRunner(runner string) string {
	clean := strings.ToLower(strings.TrimSpace(runner))
	if clean == "" {
		return "ssh"
	}
	if clean == "local" || clean == "controlled-local" {
		return "local"
	}
	if clean == "systemd" || clean == "local-systemd" || clean == "linux-systemd" {
		return "systemd"
	}
	return "ssh"
}

func isSafeLinuxInstallerRunner(runner string) bool {
	switch runner {
	case "local", "systemd", "ssh":
		return true
	default:
		return false
	}
}

func isSafeSystemdMode(mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "system", "user":
		return true
	default:
		return false
	}
}

func safetyPolicyExists(path string) bool {
	candidate := strings.TrimSpace(path)
	if candidate == "" {
		candidate = defaultLinuxInstallerSafetyPolicyPath
	}
	if filepath.ToSlash(filepath.Clean(candidate)) != defaultLinuxInstallerSafetyPolicyPath {
		return false
	}
	if _, err := os.Stat(candidate); err == nil {
		return true
	}
	if filepath.IsAbs(candidate) {
		return false
	}
	prefix := ".."
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(prefix, candidate)); err == nil {
			return true
		}
		prefix = filepath.Join(prefix, "..")
	}
	return false
}
