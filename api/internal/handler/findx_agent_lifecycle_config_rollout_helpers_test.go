package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func configRolloutBlockedMatrixBody(templateID, scope, strategy string) string {
	return `{"template_id":"` + templateID + `","target_ids":["target-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"local","plugin_id":"input.cpu","reload_strategy":"hup","rollout_strategy":"` + strategy + `","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"` + scope + `","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `}}`
}

func decodeConfigRolloutEnvelope(t *testing.T, w *httptest.ResponseRecorder) struct {
	Error             string                                     `json:"error"`
	Status            string                                     `json:"status"`
	Blockers          []string                                   `json:"blockers"`
	MissingContracts  []string                                   `json:"missing_contracts"`
	StateMachine      model.FindXAgentExecutionStateMachine      `json:"state_machine"`
	ReceiptContract   model.FindXAgentReceiptContract            `json:"receipt_contract"`
	ReceiptMatrix     []model.FindXAgentReceiptContractMatrixRow `json:"receipt_matrix"`
	OperationContract map[string]any                             `json:"operation_contract"`
	SafeToRetry       bool                                       `json:"safe_to_retry"`
	Data              model.FindXAgentConfigRollout              `json:"data"`
} {
	t.Helper()
	var payload struct {
		Error             string                                     `json:"error"`
		Status            string                                     `json:"status"`
		Blockers          []string                                   `json:"blockers"`
		MissingContracts  []string                                   `json:"missing_contracts"`
		StateMachine      model.FindXAgentExecutionStateMachine      `json:"state_machine"`
		ReceiptContract   model.FindXAgentReceiptContract            `json:"receipt_contract"`
		ReceiptMatrix     []model.FindXAgentReceiptContractMatrixRow `json:"receipt_matrix"`
		OperationContract map[string]any                             `json:"operation_contract"`
		SafeToRetry       bool                                       `json:"safe_to_retry"`
		Data              model.FindXAgentConfigRollout              `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid config rollout envelope: %v body=%s", err, w.Body.String())
	}
	return payload
}

func containsLifecycleTestStringFragment(values []string, want string) bool {
	for _, value := range values {
		if strings.Contains(value, want) {
			return true
		}
	}
	return false
}

func completeHTTPProviderConfigRolloutMetadata() string {
	return `"provider_endpoint_ref":"provider-endpoint-ref","provider_response_version_ref":"provider-version-ref","provider_checksum_ref":"provider-checksum-ref","checksum_registry_ref":"checksum-registry-ref","provider_headers_ref":"provider-headers-ref","provider_auth_ref":"provider-auth-ref","provider_tls_ref":"provider-tls-ref","reload_interval_ref":"reload-interval-ref","timeout_ref":"timeout-ref","config_serving_receipt_ref":"config-serving-receipt-ref"`
}

func completeRemotePreflightMetadataWithoutTargetOSAndTransport() string {
	return `"idempotency_key":"idem-1","timeout_policy_ref":"timeout-1","execution_receipt_ref":"receipt-1","audit_ref":"audit-1","data_arrival_validator_ref":"validator-1","ssh_runner_ref":"ssh-runner-1","ssh_host_key":"host-key-1","ssh_fingerprint":"fingerprint-1","remote_executor_ref":"remote-executor-1"`
}

func completePluginConfigRolloutMetadata() string {
	return `"plugin_config_writer_ref":"plugin-writer-ref","reload_command_ref":"reload-command-ref","reload_receipt_ref":"reload-receipt-ref","drift_check_ref":"drift-check-ref","evidence_chain_ref":"evidence-chain-ref","rollback_receipt_ref":"rollback-receipt-ref"`
}

func completeKubernetesConfigRolloutMetadata() string {
	return `"cluster_ref":"cluster-ref","namespace_ref":"namespace-ref","workload_selector_ref":"workload-ref","config_map_ref":"config-map-ref","rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref","reload_receipt_ref":"reload-receipt-ref","drift_check_ref":"drift-check-ref","data_arrival_validator_ref":"validator-ref"`
}

func decodeConfigRolloutResponse(t *testing.T, w *httptest.ResponseRecorder) struct {
	Error string                        `json:"error"`
	Data  model.FindXAgentConfigRollout `json:"data"`
} {
	t.Helper()
	var payload struct {
		Error string                        `json:"error"`
		Data  model.FindXAgentConfigRollout `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid config rollout response: %v body=%s", err, w.Body.String())
	}
	return payload
}

func assertNoConfigRolloutSensitiveEcho(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{`"credential_ref":`, "<credential_ref>", "secret", "token", "cookie", "session", "password", "private key", "dsn", "marker-secret", "sensitive-secret"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("response must not echo config rollout sensitive values: %s", body)
		}
	}
}
