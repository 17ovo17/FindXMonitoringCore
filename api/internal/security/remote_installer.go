package security

import (
	"sort"
	"strings"
)

type RemoteInstallerPreflightInput struct {
	Scope                   string
	CredentialRef           string
	RemoteExecutorRef       string
	SSHRunnerRef            string
	SSHHostKey              string
	SSHFingerprint          string
	WinRMEndpointRef        string
	WinRMTransportRef       string
	TimeoutPolicyRef        string
	IdempotencyKey          string
	InstallReceiptRef       string
	ExecutionReceiptRef     string
	ServiceReceiptRef       string
	HeartbeatValidatorRef   string
	DataArrivalValidatorRef string
	AuditRef                string
	EvidenceChainRef        string
	PackageRepositoryRef    string
	SignatureRef            string
	Checksum                string
	ScriptManifestRef       string
}

type RemoteInstallerPreflightResult struct {
	Allowed bool
	Status  string
	Reason  string
	Scope   string
	Runner  string
}

func EvaluateRemoteInstallerPreflight(input RemoteInstallerPreflightInput) RemoteInstallerPreflightResult {
	scope := NormalizeRemoteInstallerScope(input.Scope)
	missing := remoteInstallerMissingPrerequisites(input, scope)
	if len(missing) > 0 {
		return RemoteInstallerPreflightResult{
			Allowed: false,
			Status:  "blocked",
			Reason:  "PENDING: remote installer prerequisites missing: " + strings.Join(missing, ", "),
			Scope:   scope,
			Runner:  scope,
		}
	}
	return RemoteInstallerPreflightResult{
		Allowed: false,
		Status:  "blocked",
		Reason:  "PENDING: " + scope + " remote executor contract is not enabled",
		Scope:   scope,
		Runner:  scope,
	}
}

func NormalizeRemoteInstallerScope(value string) string {
	clean := strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.Contains(clean, "winrm"):
		return "winrm"
	case strings.Contains(clean, "ssh"):
		return "ssh"
	default:
		return ""
	}
}

func remoteInstallerMissingPrerequisites(input RemoteInstallerPreflightInput, scope string) []string {
	required := []struct {
		name  string
		value string
	}{
		{"credential_ref", input.CredentialRef},
		{"remote_executor_ref", input.RemoteExecutorRef},
		{"timeout_policy_ref", input.TimeoutPolicyRef},
		{"idempotency_key", input.IdempotencyKey},
		{"install_receipt_ref", input.InstallReceiptRef},
		{"execution_receipt_ref", input.ExecutionReceiptRef},
		{"service_receipt_ref", input.ServiceReceiptRef},
		{"heartbeat_validator_ref", input.HeartbeatValidatorRef},
		{"data_arrival_validator_ref", input.DataArrivalValidatorRef},
		{"package_repository_ref", input.PackageRepositoryRef},
		{"signature_ref", input.SignatureRef},
		{"checksum", input.Checksum},
		{"script_manifest_ref", input.ScriptManifestRef},
	}
	missingSet := map[string]bool{}
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			missingSet[item.name] = true
		}
	}
	if strings.TrimSpace(input.AuditRef) == "" && strings.TrimSpace(input.EvidenceChainRef) == "" {
		missingSet["audit_ref_or_evidence_chain_ref"] = true
	}
	switch scope {
	case "ssh":
		if strings.TrimSpace(input.SSHRunnerRef) == "" {
			missingSet["ssh_runner_ref"] = true
		}
		if strings.TrimSpace(input.SSHHostKey) == "" && strings.TrimSpace(input.SSHFingerprint) == "" {
			missingSet["ssh_host_key_or_fingerprint"] = true
		}
	case "winrm":
		if strings.TrimSpace(input.WinRMEndpointRef) == "" {
			missingSet["winrm_endpoint_ref"] = true
		}
		if strings.TrimSpace(input.WinRMTransportRef) == "" {
			missingSet["winrm_transport_ref"] = true
		}
	default:
		missingSet["transport_or_runner"] = true
	}
	missing := make([]string, 0, len(missingSet))
	for key := range missingSet {
		missing = append(missing, key)
	}
	sort.Strings(missing)
	return missing
}
