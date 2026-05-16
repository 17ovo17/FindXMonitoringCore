package security

import (
	"sort"
	"strings"
)

type KubernetesInstallerPrerequisites struct {
	CredentialRefPresent    bool
	ClusterRef              string
	NamespaceRef            string
	WorkloadSelectorRef     string
	HelmChartRef            string
	ManifestBundleRef       string
	ValuesRef               string
	RBACRef                 string
	ServiceAccountRef       string
	ImageRef                string
	PackageRepositoryRef    string
	SignatureRef            string
	Checksum                string
	ConfigMapRef            string
	SecretRef               string
	RolloutStrategyRef      string
	RolloutReceiptRef       string
	DataArrivalValidatorRef string
	ExecutorRef             string
	AuditRef                string
	EvidenceChainRef        string
	Method                  string
	OS                      string
}

type KubernetesInstallerGateResult struct {
	Allowed bool
	Status  string
	Reason  string
	Runner  string
}

func EvaluateKubernetesInstallerPrerequisites(input KubernetesInstallerPrerequisites) KubernetesInstallerGateResult {
	runner := NormalizeKubernetesInstallerRunner(input.Method, input.OS)
	missing := kubernetesInstallerMissingPrerequisites(input)
	if len(missing) > 0 {
		return KubernetesInstallerGateResult{
			Allowed: false,
			Status:  "blocked",
			Reason:  "PENDING: missing " + strings.Join(missing, ", "),
			Runner:  runner,
		}
	}
	return KubernetesInstallerGateResult{
		Allowed: false,
		Status:  "blocked",
		Reason:  "PENDING: Kubernetes executor not enabled / lifecycle protocol not open",
		Runner:  runner,
	}
}

func kubernetesInstallerMissingPrerequisites(input KubernetesInstallerPrerequisites) []string {
	required := []struct {
		name  string
		value string
	}{
		{"cluster_ref", input.ClusterRef},
		{"namespace_ref", input.NamespaceRef},
		{"workload_selector_ref", input.WorkloadSelectorRef},
		{"values_ref", input.ValuesRef},
		{"rbac_ref", input.RBACRef},
		{"service_account_ref", input.ServiceAccountRef},
		{"image_ref", input.ImageRef},
		{"package_repository_ref", input.PackageRepositoryRef},
		{"signature_ref", input.SignatureRef},
		{"checksum", input.Checksum},
		{"config_map_ref", input.ConfigMapRef},
		{"rollout_strategy_ref", input.RolloutStrategyRef},
		{"rollout_receipt_ref", input.RolloutReceiptRef},
		{"data_arrival_validator_ref", input.DataArrivalValidatorRef},
		{"executor_ref", input.ExecutorRef},
	}
	missing := []string{}
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			missing = append(missing, item.name)
		}
	}
	if !input.CredentialRefPresent {
		missing = append(missing, "credential_ref")
	}
	if strings.TrimSpace(input.HelmChartRef) == "" && strings.TrimSpace(input.ManifestBundleRef) == "" {
		missing = append(missing, "helm_chart_ref_or_manifest_bundle_ref")
	}
	if strings.TrimSpace(input.SecretRef) == "" && !input.CredentialRefPresent {
		missing = append(missing, "secret_ref_or_credential_ref")
	}
	if strings.TrimSpace(input.AuditRef) == "" && strings.TrimSpace(input.EvidenceChainRef) == "" {
		missing = append(missing, "audit_ref_or_evidence_chain_ref")
	}
	sort.Strings(missing)
	return missing
}

func NormalizeKubernetesInstallerRunner(method, osName string) string {
	clean := strings.ToLower(strings.TrimSpace(method))
	switch {
	case strings.Contains(clean, "helm"):
		return "helm"
	case strings.Contains(clean, "daemonset"):
		return "kubernetes-daemonset"
	case strings.Contains(clean, "sidecar"):
		return "kubernetes-sidecar"
	case strings.Contains(clean, "initcontainer"), strings.Contains(clean, "init-container"):
		return "kubernetes-initcontainer"
	case strings.Contains(clean, "k8s"), strings.Contains(clean, "kubernetes"):
		return "kubernetes"
	case strings.EqualFold(strings.TrimSpace(osName), "kubernetes"):
		return "kubernetes"
	default:
		return "kubernetes"
	}
}
