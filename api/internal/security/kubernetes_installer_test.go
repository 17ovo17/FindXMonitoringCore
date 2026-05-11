package security

import (
	"sort"
	"strings"
	"testing"
)

func TestKubernetesInstallerPrerequisitesRequireStableSortedContractRefs(t *testing.T) {
	result := EvaluateKubernetesInstallerPrerequisites(KubernetesInstallerPrerequisites{
		Method: "helm",
		OS:     "Kubernetes",
	})
	if result.Allowed || result.Status != "blocked" || result.Runner != "helm" {
		t.Fatalf("expected blocked helm gate, got %#v", result)
	}
	if !strings.HasPrefix(result.Reason, "BLOCKED_BY_CONTRACT: missing ") {
		t.Fatalf("expected missing contract reason, got %q", result.Reason)
	}
	got := strings.Split(strings.TrimPrefix(result.Reason, "BLOCKED_BY_CONTRACT: missing "), ", ")
	if !sort.StringsAreSorted(got) {
		t.Fatalf("missing refs must be stable sorted, got %#v", got)
	}
	for _, want := range []string{
		"credential_ref",
		"cluster_ref",
		"namespace_ref",
		"workload_selector_ref",
		"helm_chart_ref_or_manifest_bundle_ref",
		"values_ref",
		"rbac_ref",
		"service_account_ref",
		"image_ref",
		"package_repository_ref",
		"signature_ref",
		"checksum",
		"config_map_ref",
		"secret_ref_or_credential_ref",
		"rollout_strategy_ref",
		"rollout_receipt_ref",
		"data_arrival_validator_ref",
		"executor_ref",
		"audit_ref_or_evidence_chain_ref",
	} {
		if !strings.Contains(result.Reason, want) {
			t.Fatalf("expected missing ref %s in %q", want, result.Reason)
		}
	}
}

func TestKubernetesInstallerPrerequisitesStillBlockWhenComplete(t *testing.T) {
	result := EvaluateKubernetesInstallerPrerequisites(completeKubernetesPrerequisites("daemonset"))
	if result.Allowed || result.Status != "blocked" {
		t.Fatalf("complete Kubernetes contract must still block execution, got %#v", result)
	}
	if result.Runner != "kubernetes-daemonset" {
		t.Fatalf("expected daemonset runner, got %q", result.Runner)
	}
	if result.Reason != "BLOCKED_BY_CONTRACT: Kubernetes executor not enabled / lifecycle protocol not open" {
		t.Fatalf("unexpected complete-block reason %q", result.Reason)
	}
}

func TestKubernetesInstallerRunnerDetection(t *testing.T) {
	tests := []struct {
		name   string
		method string
		osName string
		want   string
	}{
		{name: "helm", method: "helm", want: "helm"},
		{name: "kubernetes os", osName: "Kubernetes", want: "kubernetes"},
		{name: "k8s", method: "k8s", want: "kubernetes"},
		{name: "daemonset", method: "k8s-daemonset", want: "kubernetes-daemonset"},
		{name: "sidecar", method: "sidecar", want: "kubernetes-sidecar"},
		{name: "initcontainer", method: "initcontainer", want: "kubernetes-initcontainer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeKubernetesInstallerRunner(tt.method, tt.osName); got != tt.want {
				t.Fatalf("expected runner %q, got %q", tt.want, got)
			}
		})
	}
}

func completeKubernetesPrerequisites(method string) KubernetesInstallerPrerequisites {
	return KubernetesInstallerPrerequisites{
		CredentialRefPresent:    true,
		ClusterRef:              "cluster-ref",
		NamespaceRef:            "namespace-ref",
		WorkloadSelectorRef:     "workload-selector-ref",
		HelmChartRef:            "chart-ref",
		ValuesRef:               "values-ref",
		RBACRef:                 "rbac-ref",
		ServiceAccountRef:       "service-account-ref",
		ImageRef:                "image-ref",
		PackageRepositoryRef:    "repo-ref",
		SignatureRef:            "signature-ref",
		Checksum:                "sha256:abc",
		ConfigMapRef:            "config-map-ref",
		RolloutStrategyRef:      "rollout-strategy-ref",
		RolloutReceiptRef:       "rollout-receipt-ref",
		DataArrivalValidatorRef: "data-arrival-validator-ref",
		ExecutorRef:             "executor-ref",
		EvidenceChainRef:        "evidence-chain-ref",
		Method:                  method,
		OS:                      "Kubernetes",
	}
}
