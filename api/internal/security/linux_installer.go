package security

import (
	"os"
	"path/filepath"
	"strings"
)

const defaultLinuxInstallerSafetyPolicyPath = "scripts/testing/safety_policy.json"

type LinuxInstallerPrerequisites struct {
	PackageRepositoryRef string
	SignatureRef         string
	Checksum             string
	ScriptManifestRef    string
	ExecutorRef          string
	SafetyPolicyPath     string
	Runner               string
	SSHHostKey           string
	SSHFingerprint       string
	SystemdUnitRef       string
	CurlInstallerRef     string
	EnvTemplateRef       string
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
			Reason:  "BLOCKED_BY_CONTRACT: Linux installer prerequisites missing: " + strings.Join(missing, ", "),
			Runner:  runner,
		}
	}
	return LinuxInstallerGateResult{
		Allowed: false,
		Status:  "blocked",
		Reason:  "BLOCKED_BY_CONTRACT: Linux installer executor is not enabled; execution record captured only",
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
		{"env_template_ref", input.EnvTemplateRef},
	}
	missing := []string{}
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			missing = append(missing, item.name)
		}
	}
	if !safetyPolicyExists(input.SafetyPolicyPath) {
		missing = append(missing, "linux_installer_safety_policy")
	}
	if runner != "local" && strings.TrimSpace(input.SSHHostKey) == "" && strings.TrimSpace(input.SSHFingerprint) == "" {
		missing = append(missing, "ssh_host_key_or_fingerprint")
	}
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
	return "ssh"
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
