package handler

import (
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentConfigRolloutRejectsMissingTargetsWithoutPersisting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"template_id":"metrics",
		"config_version":"cfg-v1",
		"config_snippet_ref":"snippet-ref",
		"rollback_ref":"rollback-ref",
		"metadata":{"executor_ref":"executor-ref"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing target should be 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "target_ids or agent_ids is required") {
		t.Fatalf("missing target error should be explicit, body=%s", w.Body.String())
	}
	rows, err := store.ListFindXAgentConfigRollouts()
	if err != nil {
		t.Fatalf("list config rollouts: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("400 validation must not persist blocked rollout: %#v", rows)
	}
}

func TestFindXAgentConfigRolloutMissingBaseRefsPersistsBlockedContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"metrics","target_ids":["target-a"]}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("missing refs should persist blocked 409, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{"PENDING", "config_snippet_ref", "config_version", "executor_ref", "rollback_ref"} {
		if !strings.Contains(payload.Error, want) || !strings.Contains(payload.Data.Blocker, want) {
			t.Fatalf("blocker should include %q, payload=%#v", want, payload)
		}
	}
}

func TestFindXAgentConfigRolloutRequiresRemotePreflightEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"template_id":"metrics",
		"target_ids":["target-a"],
		"config_version":"cfg-v1",
		"config_snippet_ref":"snippet-ref",
		"rollback_ref":"rollback-ref",
		"metadata":{"executor_ref":"executor-ref"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("missing remote preflight should persist blocked 409, code=%d payload=%#v", w.Code, payload)
	}
	want := "PENDING: missing audit_ref_or_evidence_chain_ref, credential_ref, execution_receipt_ref_or_receipt_ref, idempotency_key, target_os, timeout_policy_ref, transport_or_runner"
	if payload.Error != want || payload.Data.Blocker != want {
		t.Fatalf("remote preflight blocker should be stable sorted, want=%q payload=%#v", want, payload)
	}
	rows, err := store.ListFindXAgentConfigRollouts()
	if err != nil {
		t.Fatalf("list config rollouts: %v", err)
	}
	if len(rows) != 1 || rows[0].Blocker != want {
		t.Fatalf("blocked remote preflight should be persisted, rows=%#v", rows)
	}
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutPluginRequiresPluginRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"template_id":"host-plugin",
		"target_ids":["target-a"],
		"config_version":"cfg-v1",
		"config_snippet_ref":"snippet-ref",
		"rollback_ref":"rollback-ref",
		"metadata":{"executor_ref":"executor-ref"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("plugin missing refs should persist blocked 409, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{"config_format", "plugin_id", "provider_mode"} {
		if !strings.Contains(payload.Error, want) {
			t.Fatalf("plugin blocker should include %q, error=%s", want, payload.Error)
		}
	}
}

func TestFindXAgentConfigRolloutPluginRequiresReloadRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"template_id":"host-plugin",
		"target_ids":["target-a"],
		"config_version":"cfg-v1",
		"config_snippet_ref":"snippet-ref",
		"config_format":"toml",
		"provider_mode":"http",
		"plugin_id":"input.cpu",
		"reload_strategy":"hup",
		"rollback_ref":"rollback-ref",
		"remote_mutation":true,
		"metadata":{"executor_ref":"executor-ref"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("plugin missing reload refs should persist blocked 409, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{"data_arrival_validator_ref", "drift_check_ref", "evidence_chain_ref", "plugin_config_writer_ref", "reload_command_ref", "reload_receipt_ref", "rollback_receipt_ref"} {
		if !strings.Contains(payload.Error, want) || !strings.Contains(payload.Data.Blocker, want) {
			t.Fatalf("plugin reload blocker should include %q, payload=%#v", want, payload)
		}
	}
}

func TestFindXAgentConfigRolloutWindowsHUPRequiresRestartReceiptRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"host-plugin","target_ids":["target-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"local","plugin_id":"input.cpu","reload_strategy":"hup","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"target_os":"windows","transport":"windows_service"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("windows hup rollout should stay blocked, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{"data_arrival_validator_ref", "restart_strategy_ref", "service_restart_receipt_ref"} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("windows hup missing_contracts should include %q, payload=%#v", want, payload)
		}
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutWindowsServiceTransportRequiresRestartReceiptRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"host-plugin","target_ids":["target-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"local","plugin_id":"input.cpu","reload_strategy":"local-reload","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"executor_ref":"executor-ref",` + completeRemotePreflightMetadataWithoutTargetOSAndTransport() + `,` + completePluginConfigRolloutMetadata() + `,"transport":"windows_service"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("windows_service transport rollout should stay blocked, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{"data_arrival_validator_ref", "restart_strategy_ref", "service_restart_receipt_ref"} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("windows_service transport missing_contracts should include %q, payload=%#v", want, payload)
		}
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutUnsafeExecPluginBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, pluginID := range []string{"inputs.exec", "input.exec", "exec"} {
		t.Run(pluginID, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			body := strings.NewReader(`{"template_id":"host-plugin","target_ids":["target-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"local","plugin_id":"` + pluginID + `","reload_strategy":"hup","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"data_arrival_validator_ref":"validator-ref"}}`)
			w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
			payload := decodeConfigRolloutEnvelope(t, w)
			if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
				t.Fatalf("unsafe exec plugin should stay blocked, code=%d payload=%#v", w.Code, payload)
			}
			if !containsLifecycleTestString(payload.Blockers, "UNSAFE_PLUGIN_PENDING") ||
				!containsLifecycleTestString(payload.MissingContracts, "unsafe_plugin_policy_ref") {
				t.Fatalf("unsafe exec plugin should expose unsafe policy blocker, payload=%#v", payload)
			}
			assertNoConfigRolloutExecutionStates(t, w.Body.String())
			assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
		})
	}
}

func TestFindXAgentConfigRolloutHTTPProviderRequiresProviderRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"template_id":"host-plugin",
		"target_ids":["target-a"],
		"config_version":"cfg-v1",
		"config_snippet_ref":"snippet-ref",
		"config_format":"toml",
		"provider_mode":"http",
		"plugin_id":"input.cpu",
		"reload_strategy":"hup",
		"rollback_ref":"rollback-ref",
		"remote_mutation":true,
		"metadata":{
			"executor_ref":"executor-ref",
			"plugin_config_writer_ref":"plugin-writer-ref",
			"reload_command_ref":"reload-command-ref",
			"reload_receipt_ref":"reload-receipt-ref",
			"drift_check_ref":"drift-check-ref",
			"evidence_chain_ref":"evidence-chain-ref",
			"rollback_receipt_ref":"rollback-receipt-ref"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("http provider missing refs should persist blocked 409, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range requiredHTTPProviderConfigRolloutRefs() {
		if !strings.Contains(payload.Error, want) || !strings.Contains(payload.Data.Blocker, want) {
			t.Fatalf("http provider blocker should include %q, payload=%#v", want, payload)
		}
	}
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutLocalProviderSkipsHTTPProviderRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"host-plugin","target_ids":["target-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"local","plugin_id":"input.cpu","reload_strategy":"local-reload","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("local provider complete refs should stay blocked 409, code=%d payload=%#v", w.Code, payload)
	}
	if payload.Error != "PENDING: executor not enabled / config rollout protocol not open" {
		t.Fatalf("local provider should skip http provider refs, payload=%#v", payload)
	}
	for _, forbidden := range requiredHTTPProviderConfigRolloutRefs() {
		if strings.Contains(payload.Error, forbidden) || strings.Contains(payload.Data.Blocker, forbidden) {
			t.Fatalf("local provider blocker must not require %q, payload=%#v", forbidden, payload)
		}
	}
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutPluginRequiresReloadStrategy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"template_id":"host-plugin",
		"target_ids":["target-a"],
		"config_version":"cfg-v1",
		"config_snippet_ref":"snippet-ref",
		"config_format":"toml",
		"provider_mode":"http",
		"plugin_id":"input.cpu",
		"rollback_ref":"rollback-ref",
		"remote_mutation":true,
		"metadata":{
			"executor_ref":"executor-ref",
			"plugin_config_writer_ref":"plugin-writer-ref",
			"reload_command_ref":"reload-command-ref",
			"reload_receipt_ref":"reload-receipt-ref",
			"drift_check_ref":"drift-check-ref",
			"evidence_chain_ref":"evidence-chain-ref",
			"rollback_receipt_ref":"rollback-receipt-ref"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("plugin missing reload_strategy should persist blocked 409, code=%d payload=%#v", w.Code, payload)
	}
	if !strings.Contains(payload.Error, "reload_strategy") || !strings.Contains(payload.Data.Blocker, "reload_strategy") {
		t.Fatalf("plugin reload blocker should include reload_strategy, payload=%#v", payload)
	}
}

func TestFindXAgentConfigRolloutKubernetesMissingRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"metrics","target_ids":["target-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","rollback_ref":"rollback-ref","credential_ref":"<CREDENTIAL_REF>","metadata":{"target_os":"kubernetes","transport":"k8s-api","idempotency_key":"idem-1","timeout_policy_ref":"timeout-1","execution_receipt_ref":"receipt-1","audit_ref":"audit-1"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, w)
	if w.Code != http.StatusConflict {
		t.Fatalf("kubernetes missing refs should stay blocked 409, got %d body=%s", w.Code, w.Body.String())
	}
	want := "PENDING: missing cluster_ref, config_map_ref, data_arrival_validator_ref, drift_check_ref, executor_ref, helm_chart_ref_or_manifest_bundle_ref, namespace_ref, reload_receipt_ref, rollout_receipt_ref, rollout_strategy_ref, workload_selector_ref"
	if payload.Error != want || payload.Data.Blocker != want {
		t.Fatalf("kubernetes missing refs should be stable sorted, want=%q payload=%#v", want, payload)
	}
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutHelmMissingChoiceAndReleaseRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"metrics","target_ids":["target-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","rollback_ref":"rollback-ref","credential_ref":"<CREDENTIAL_REF>","metadata":{"method":"helm","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completeKubernetesConfigRolloutMetadata() + `}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, w)
	if w.Code != http.StatusConflict {
		t.Fatalf("helm missing refs should stay blocked 409, got %d body=%s", w.Code, w.Body.String())
	}
	want := "PENDING: missing helm_chart_ref_or_manifest_bundle_ref, helm_release_ref"
	if payload.Error != want || payload.Data.Blocker != want {
		t.Fatalf("helm missing refs should be stable sorted, want=%q payload=%#v", want, payload)
	}
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCompleteRefsStillBlockedByExecutorGate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"host-plugin","target_ids":["target-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"http","plugin_id":"input.cpu","reload_strategy":"hup","rollback_ref":"rollback-ref","remote_mutation":true,"change_ticket":"CHG-1","audit_reason":"planned rollout","credential_ref":"<CREDENTIAL_REF>","metadata":{"method":"helm","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,` + completeHTTPProviderConfigRolloutMetadata() + `,` + completeKubernetesConfigRolloutMetadata() + `,"helm_release_ref":"release-ref","helm_chart_ref":"chart-ref","ticket":"CHG-1","token":"secret","cookie":"session-secret","credential_ref":"credential-secret","dsn":"mysql://user:pass@host/db","provider_auth_token":"secret-token","provider_password_ref":"secret-password"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, w)
	if w.Code != http.StatusConflict {
		t.Fatalf("complete refs should still be blocked 409, got %d body=%s", w.Code, w.Body.String())
	}
	if payload.Error != "PENDING: executor not enabled / config rollout protocol not open" {
		t.Fatalf("unexpected executor gate blocker: %#v", payload)
	}
	if payload.Data.ReloadStrategy != "hup" {
		t.Fatalf("reload_strategy should be retained as a non-sensitive request ref: %#v", payload.Data)
	}
	retainedRefs := []string{"plugin_config_writer_ref", "reload_command_ref", "reload_receipt_ref", "drift_check_ref", "evidence_chain_ref", "rollback_receipt_ref"}
	for _, key := range append(retainedRefs, requiredHTTPProviderConfigRolloutRefs()...) {
		if payload.Data.Metadata[key] == "" {
			t.Fatalf("safe rollout metadata ref %s should be retained, metadata=%#v", key, payload.Data.Metadata)
		}
	}
	for _, key := range []string{"token", "cookie", "credential_ref", "dsn", "provider_auth_token", "provider_password_ref"} {
		if payload.Data.Metadata[key] != "" {
			t.Fatalf("sensitive metadata key %s should be dropped, metadata=%#v", key, payload.Data.Metadata)
		}
	}
	for _, forbidden := range []string{`"status":"queued"`, `"status":"running"`, `"status":"succeeded"`, `"status":"success"`, `"status":"applied"`, `"status":"rolled-back"`} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("config rollout must not expose execution success states: %s", w.Body.String())
		}
	}
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutResponseIncludesSafeContractEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"host-plugin","target_ids":["target-a"],"remote_mutation":true,"metadata":{"marker":"marker-secret","sensitive":"sensitive-secret"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Status != "blocked" {
		t.Fatalf("config rollout should return blocked envelope, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{"PENDING", "MISSING_CONTRACTS"} {
		if !containsLifecycleTestString(payload.Blockers, want) {
			t.Fatalf("blocked envelope should include %q, payload=%#v", want, payload)
		}
	}
	for _, want := range []string{"config_snippet_ref", "config_version", "executor_ref", "reload_receipt_ref"} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("missing_contracts should include %q, payload=%#v", want, payload)
		}
	}
	if payload.ReceiptContract.ID == "" ||
		payload.ReceiptContract.Scope != "findx_agent_plugin_config_rollout" ||
		payload.ReceiptContract.Status != "blocked_by_contract" ||
		!payload.ReceiptContract.CredentialRequired ||
		payload.ReceiptContract.CredentialProvided {
		t.Fatalf("blocked envelope should include config rollout receipt_contract, payload=%#v", payload)
	}
	for _, want := range []string{"writer_receipt", "reload_receipt", "restart_receipt", "drift_receipt", "rollback_receipt", "data_arrival_receipt", "evidence_chain"} {
		if !containsLifecycleTestString(payload.ReceiptContract.RequiredReceipts, want) {
			t.Fatalf("receipt_contract required_receipts should include %q, payload=%#v", want, payload)
		}
	}
	if !containsLifecycleTestString(payload.ReceiptContract.MissingContracts, "config_snippet_ref") ||
		!containsLifecycleTestString(payload.ReceiptContract.MissingContracts, "credential_ref") {
		t.Fatalf("receipt_contract missing_contracts should mirror missing refs, payload=%#v", payload)
	}
	assertConfigRolloutReceiptMatrixBlocked(t, payload.ReceiptMatrix)
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutReceiptScopeRejectsUnknownMetadataScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, scope := range []string{"../token", "business group", "custom", "agent-token", "   "} {
		t.Run(scope, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			body := strings.NewReader(`{"template_id":"host-plugin","target_ids":["target-a"],"remote_mutation":true,"metadata":{"scope":"` + scope + `"}}`)
			w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
			payload := decodeConfigRolloutEnvelope(t, w)
			if w.Code != http.StatusConflict || payload.Status != "blocked" {
				t.Fatalf("unknown scope should stay blocked, code=%d payload=%#v", w.Code, payload)
			}
			if payload.ReceiptContract.Scope != "findx_agent_plugin_config_rollout" {
				t.Fatalf("unknown scope must not be echoed into receipt contract, scope=%q payload=%#v", scope, payload)
			}
			if strings.Contains(payload.ReceiptContract.Scope, "token") ||
				strings.Contains(payload.ReceiptContract.Scope, "..") ||
				strings.Contains(payload.ReceiptContract.Scope, "business group") ||
				strings.Contains(payload.ReceiptContract.Scope, "custom") {
				t.Fatalf("receipt contract scope contains unsafe metadata scope: %#v", payload.ReceiptContract)
			}
			assertNoConfigRolloutExecutionStates(t, w.Body.String())
			assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
		})
	}
}

func TestFindXAgentConfigRolloutPluginClassScopeAndStrategyMatrixStaysBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name       string
		templateID string
		scope      string
		strategy   string
		want       []string
	}{
		{"host agent canary", "host-plugin", "agent", "canary", []string{"agent_ref", "rollout_receipt_ref"}},
		{"container namespace full", "container-plugin", "namespace", "full", []string{"namespace_ref", "rollout_receipt_ref", "rollout_strategy_ref"}},
		{"container workload reload", "container-plugin", "workload", "reload", []string{"namespace_ref", "workload_selector_ref", "reload_receipt_ref"}},
		{"gateway cmdb rollback", "gateway-plugin", "cmdb_host", "rollback", []string{"cmdb_host_ref", "rollback_receipt_ref"}},
		{"gateway business drift", "gateway-plugin", "business_group", "drift", []string{"business_group_ref", "drift_check_ref"}},
		{"host evidence", "host-plugin", "agent", "evidence", []string{"agent_ref", "evidence_chain_ref"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			body := strings.NewReader(configRolloutBlockedMatrixBody(tt.templateID, tt.scope, tt.strategy))
			w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
			payload := decodeConfigRolloutEnvelope(t, w)
			if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
				t.Fatalf("matrix case should stay blocked, code=%d payload=%#v", w.Code, payload)
			}
			for _, want := range tt.want {
				if !containsLifecycleTestString(payload.MissingContracts, want) {
					t.Fatalf("matrix missing_contracts should include %q, payload=%#v", want, payload)
				}
			}
			assertNoConfigRolloutExecutionStates(t, w.Body.String())
			assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
		})
	}
}

func TestFindXAgentConfigRolloutCompleteRefsEnvelopeStillExecutorDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"container-plugin","target_ids":["target-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"http","plugin_id":"input.kubernetes","reload_strategy":"rolling-restart","rollout_strategy":"full","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"workload","namespace_ref":"namespace-ref","workload_selector_ref":"workload-ref","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,` + completeHTTPProviderConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || len(payload.MissingContracts) != 0 {
		t.Fatalf("complete refs should only hit executor disabled gate, code=%d payload=%#v", w.Code, payload)
	}
	if payload.SafeToRetry || !payload.StateMachine.Terminal ||
		payload.StateMachine.CurrentState != "blocked_by_contract" ||
		len(payload.StateMachine.AllowedStates) != 1 ||
		payload.StateMachine.AllowedStates[0] != "blocked_by_contract" {
		t.Fatalf("complete refs should expose terminal blocked state machine, payload=%#v", payload)
	}
	if !containsLifecycleTestString(payload.Blockers, "EXECUTOR_DISABLED_BY_CONTRACT") {
		t.Fatalf("complete refs should expose executor disabled blocker, payload=%#v", payload)
	}
	if payload.ReceiptContract.Status != "blocked_by_contract" ||
		len(payload.ReceiptContract.MissingContracts) != 1 ||
		payload.ReceiptContract.MissingContracts[0] != "executor_disabled_contract" ||
		!payload.ReceiptContract.CredentialProvided {
		t.Fatalf("complete refs should expose only executor-disabled receipt contract, payload=%#v", payload)
	}
	assertConfigRolloutReceiptMatrixBlocked(t, payload.ReceiptMatrix)
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbPluginCredentialAndDashboardContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"cmdb-host-plugin-dispatch","target_ids":["host-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"local","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"full","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"cmdb_host","cmdb_host_ref":"host-a","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref","dashboard_refs":"dashboard:redis-overview","credential_ref":"password=secret","token":"secret-token","cookie":"secret-cookie"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("cmdb plugin rollout should stay blocked, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{
		"cmdb_agent_plugin_credential_contract",
		"cmdb_credential_ref_resolve_contract",
		"cmdb_plugin_config_schema_contract",
		"cmdb_dashboard_template_lookup_contract",
		"cmdb_dashboard_import_runtime_contract",
	} {
		if !containsLifecycleTestString(payload.MissingContracts, want) ||
			!containsLifecycleTestString(payload.ReceiptContract.MissingContracts, want) {
			t.Fatalf("cmdb plugin missing_contracts should include %q, payload=%#v", want, payload)
		}
	}
	if !payload.Data.CredentialRefPresent {
		t.Fatalf("top-level credential_ref should be recorded as present without echoing the value: %#v", payload.Data)
	}
	if payload.Data.Metadata["credential_ref"] != "" || payload.Data.Metadata["token"] != "" || payload.Data.Metadata["cookie"] != "" {
		t.Fatalf("sensitive metadata must be dropped, metadata=%#v", payload.Data.Metadata)
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbPluginAssignAndDispatchHaveDifferentContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name           string
		templateID     string
		strategy       string
		remoteMutation bool
		wantMode       string
		wantContract   string
		wantRequired   []string
		wantMissing    []string
		forbidRequired []string
		forbidMissing  []string
		wantRemote     bool
	}{
		{
			name:           "assign records target plugin binding intent only",
			templateID:     "cmdb-host-plugin-assign",
			strategy:       "assign",
			remoteMutation: false,
			wantMode:       "assign",
			wantContract:   "cmdb.agent.plugin.assignment.v1",
			wantRequired:   []string{"assignment_record", "target_binding_ref", "credential_policy_ref", "audit_ref"},
			wantMissing:    []string{"cmdb_agent_plugin_credential_contract", "cmdb_dashboard_import_runtime_contract"},
			forbidRequired: []string{"writer_receipt", "reload_receipt", "effect_receipt", "data_arrival_receipt"},
			forbidMissing:  []string{"cmdb_agent_plugin_assignment_store_contract", "cmdb_agent_plugin_target_binding_contract", "cmdb_agent_plugin_assignment_audit_contract", "cmdb_agent_rollout_delivery_receipt_contract", "cmdb_agent_rollout_effect_receipt_contract"},
			wantRemote:     false,
		},
		{
			name:           "dispatch records remote delivery intent",
			templateID:     "cmdb-host-plugin-dispatch",
			strategy:       "dispatch",
			remoteMutation: true,
			wantMode:       "dispatch",
			wantContract:   "cmdb.agent.plugin.dispatch.v1",
			wantRequired:   []string{"assignment_ref", "writer_receipt", "delivery_receipt", "effect_receipt", "rollback_receipt", "data_arrival_receipt", "evidence_chain"},
			wantMissing:    []string{"cmdb_agent_rollout_delivery_receipt_contract", "cmdb_agent_rollout_effect_receipt_contract"},
			forbidRequired: []string{"assignment_record"},
			wantRemote:     true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			body := strings.NewReader(`{"template_id":"` + tt.templateID + `","target_ids":["host-a"],"agent_ids":["agent-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"cmdb_host_probe","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"` + tt.strategy + `","rollback_ref":"rollback-ref","remote_mutation":` + boolString(tt.remoteMutation) + `,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"cmdb_host","cmdb_host_ref":"host-a","agent_ref":"agent-a","dashboard_refs":"dashboard:redis-overview","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref"}}`)
			w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
			payload := decodeConfigRolloutEnvelope(t, w)
			if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
				t.Fatalf("cmdb plugin %s should stay blocked, code=%d payload=%#v", tt.strategy, w.Code, payload)
			}
			op := payload.OperationContract
			if op["mode"] != tt.wantMode || op["contract"] != tt.wantContract || op["remote_mutation"] != tt.wantRemote {
				t.Fatalf("operation_contract mismatch for %s: %#v", tt.strategy, op)
			}
			required := anySliceToStrings(t, op["required_receipts"])
			missing := anySliceToStrings(t, op["missing_contracts"])
			for _, want := range tt.wantRequired {
				if !containsLifecycleTestString(required, want) {
					t.Fatalf("%s required_receipts missing %q: %#v", tt.strategy, want, op)
				}
			}
			for _, want := range tt.wantMissing {
				if !containsLifecycleTestString(missing, want) ||
					!containsLifecycleTestString(payload.MissingContracts, want) ||
					!containsLifecycleTestString(payload.ReceiptContract.MissingContracts, want) {
					t.Fatalf("%s missing_contracts missing %q: op=%#v payload=%#v", tt.strategy, want, op, payload)
				}
			}
			for _, forbidden := range tt.forbidRequired {
				if containsLifecycleTestString(required, forbidden) {
					t.Fatalf("%s required_receipts should not include %q: %#v", tt.strategy, forbidden, op)
				}
			}
			for _, forbidden := range tt.forbidMissing {
				if containsLifecycleTestString(missing, forbidden) {
					t.Fatalf("%s missing_contracts should not include %q: %#v", tt.strategy, forbidden, op)
				}
			}
			if payload.Data.Metadata["plugin_action"] != tt.wantMode || payload.Data.Metadata["cmdb_host_ref"] != "host-a" {
				t.Fatalf("rollout metadata should preserve safe action/host refs: %#v", payload.Data.Metadata)
			}
			assertNoConfigRolloutExecutionStates(t, w.Body.String())
			assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
		})
	}
}

func TestFindXAgentConfigRolloutCmdbPluginActionIdentityConflictIsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"cmdb-host-plugin-assign","target_ids":["host-a"],"agent_ids":["agent-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"cmdb_host_probe","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"assign","rollback_ref":"rollback-ref","remote_mutation":false,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"cmdb_host","plugin_action":"dispatch","cmdb_host_ref":"host-a","agent_ref":"agent-a","dashboard_refs":"dashboard:redis-overview","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("conflicting plugin action should stay blocked, code=%d payload=%#v", w.Code, payload)
	}
	op := payload.OperationContract
	if op["mode"] != "assign" || op["contract"] != "cmdb.agent.plugin.assignment.v1" || op["remote_mutation"] != false {
		t.Fatalf("metadata plugin_action must not override template identity: %#v", op)
	}
	for _, values := range [][]string{
		payload.MissingContracts,
		payload.ReceiptContract.MissingContracts,
		anySliceToStrings(t, op["missing_contracts"]),
	} {
		if !containsLifecycleTestString(values, configRolloutPluginOperationConflict) {
			t.Fatalf("conflicting action should include identity contract, values=%#v payload=%#v", values, payload)
		}
	}
	if payload.Data.Metadata["plugin_action"] != "assign" || payload.Data.Metadata["plugin_action_conflict"] != "blocked" {
		t.Fatalf("metadata should persist derived action and conflict marker: %#v", payload.Data.Metadata)
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbPluginMissingCredentialRefKeepsRuntimeContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"cmdb-host-plugin-assign","target_ids":["host-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"local","plugin_id":"mysql","reload_strategy":"local-reload","rollback_ref":"rollback-ref","remote_mutation":true,"metadata":{"scope":"cmdb_host","cmdb_host_ref":"host-a","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"dashboard_refs":"dashboard:mysql-overview"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" || payload.Data.CredentialRefPresent {
		t.Fatalf("missing credential_ref should persist blocked without credential present, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{"credential_ref", "cmdb_agent_plugin_credential_contract", "cmdb_credential_ref_resolve_contract", "cmdb_dashboard_import_runtime_contract"} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("missing credential rollout should include %q, payload=%#v", want, payload)
		}
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbPluginAssignPersistsAssignmentStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"cmdb-host-plugin-assign","target_ids":["host-a"],"agent_ids":["agent-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"cmdb_host_probe","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"assign","rollback_ref":"rollback-ref","remote_mutation":false,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"cmdb_host","cmdb_host_ref":"host-a","agent_ref":"agent-a","dashboard_refs":"dashboard:redis-overview","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref","token":"secret-token","dsn":"mysql://user:pass@host/db"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("assign should still return blocked 409, code=%d payload=%#v", w.Code, payload)
	}
	assignmentRef, _ := payload.OperationContract["assignment_ref"].(string)
	bindingRef, _ := payload.OperationContract["target_binding_ref"].(string)
	if assignmentRef == "" || bindingRef == "" {
		t.Fatalf("assign should expose safe assignment and target binding refs: %#v", payload.OperationContract)
	}
	assignment, ok, err := store.GetFindXAgentPluginAssignment(assignmentRef)
	if err != nil || !ok {
		t.Fatalf("assignment should be persisted, ok=%v err=%v ref=%s", ok, err, assignmentRef)
	}
	if assignment.HostRef != "host-a" || assignment.AgentRef != "agent-a" || assignment.PluginID != "redis" || !assignment.CredentialRefPresent {
		t.Fatalf("assignment identity mismatch: %#v", assignment)
	}
	bindings, err := store.ListFindXAgentPluginTargetBindings(assignment.ID)
	if err != nil || len(bindings) != 1 || bindings[0].ID != bindingRef {
		t.Fatalf("target binding should be persisted, bindings=%#v err=%v", bindings, err)
	}
	for _, forbidden := range []string{
		"cmdb_agent_plugin_assignment_store_contract",
		"cmdb_agent_plugin_target_binding_contract",
		"cmdb_agent_plugin_assignment_audit_contract",
	} {
		if containsLifecycleTestString(payload.MissingContracts, forbidden) ||
			containsLifecycleTestString(payload.ReceiptContract.MissingContracts, forbidden) ||
			containsLifecycleTestString(anySliceToStrings(t, payload.OperationContract["missing_contracts"]), forbidden) {
			t.Fatalf("persisted assignment should remove %q from missing contracts: %#v", forbidden, payload)
		}
	}
	if containsLifecycleTestString(anySliceToStrings(t, payload.OperationContract["missing_contracts"]), "cmdb_credential_ref_resolve_contract") == false {
		t.Fatalf("credential resolver must remain blocked after assignment persistence: %#v", payload.OperationContract)
	}
	auditResp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
		Source:       "findx_audit",
		Scope:        "cmdb",
		ResourceType: "cmdb_agent_plugin_assignment",
		ResourceID:   assignment.ID,
		Action:       "cmdb.agent.plugin.assignment.save",
		Limit:        5,
	})
	if err != nil || len(auditResp.Items) == 0 {
		t.Fatalf("assignment audit should be queryable, items=%#v err=%v", auditResp.Items, err)
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbPluginDispatchRequiresResolvableAssignment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"cmdb-host-plugin-dispatch","target_ids":["host-a"],"agent_ids":["agent-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"cmdb_host_probe","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"dispatch","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"cmdb_host","cmdb_host_ref":"host-a","agent_ref":"agent-a","assignment_ref":"missing-assignment-ref","dashboard_refs":"dashboard:redis-overview","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("dispatch with stale assignment should stay blocked, code=%d payload=%#v", w.Code, payload)
	}
	for _, values := range [][]string{
		payload.MissingContracts,
		payload.ReceiptContract.MissingContracts,
		anySliceToStrings(t, payload.OperationContract["missing_contracts"]),
	} {
		if !containsLifecycleTestString(values, "cmdb_agent_plugin_assignment_ref_contract") {
			t.Fatalf("stale assignment_ref should keep assignment ref blocker, values=%#v payload=%#v", values, payload)
		}
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbPluginDispatchResolvesAssignmentRefButKeepsReceiptsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	assignBody := strings.NewReader(`{"template_id":"cmdb-host-plugin-assign","target_ids":["host-a"],"agent_ids":["agent-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"cmdb_host_probe","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"assign","rollback_ref":"rollback-ref","remote_mutation":false,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"cmdb_host","cmdb_host_ref":"host-a","agent_ref":"agent-a","dashboard_refs":"dashboard:redis-overview","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref"}}`)
	assignResp := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", assignBody, CreateFindXAgentConfigRollout)
	assignPayload := decodeConfigRolloutEnvelope(t, assignResp)
	assignmentRef, _ := assignPayload.OperationContract["assignment_ref"].(string)
	if assignmentRef == "" {
		t.Fatalf("assign should create assignment_ref before dispatch: %#v", assignPayload.OperationContract)
	}

	dispatchBody := strings.NewReader(`{"template_id":"cmdb-host-plugin-dispatch","target_ids":["host-a"],"agent_ids":["agent-a"],"config_version":"cfg-v2","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"cmdb_host_probe","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"dispatch","rollback_ref":"rollback-ref","remote_mutation":true,"credential_ref":"<CREDENTIAL_REF>","metadata":{"scope":"cmdb_host","cmdb_host_ref":"host-a","agent_ref":"agent-a","assignment_ref":"` + assignmentRef + `","dashboard_refs":"dashboard:redis-overview","executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", dispatchBody, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("dispatch should still be blocked, code=%d payload=%#v", w.Code, payload)
	}
	if got, _ := payload.OperationContract["assignment_ref"].(string); got != assignmentRef {
		t.Fatalf("dispatch should expose resolved assignment_ref, got=%q want=%q contract=%#v", got, assignmentRef, payload.OperationContract)
	}
	for _, values := range [][]string{
		payload.MissingContracts,
		payload.ReceiptContract.MissingContracts,
		anySliceToStrings(t, payload.OperationContract["missing_contracts"]),
	} {
		if containsLifecycleTestString(values, "cmdb_agent_plugin_assignment_ref_contract") {
			t.Fatalf("resolved assignment_ref should be removed from missing contracts, values=%#v payload=%#v", values, payload)
		}
		for _, want := range []string{
			"cmdb_agent_rollout_delivery_receipt_contract",
			"cmdb_agent_rollout_effect_receipt_contract",
			"cmdb_agent_rollout_rollback_receipt_contract",
			"cmdb_agent_rollout_data_arrival_contract",
			"cmdb_agent_rollout_evidence_chain_contract",
		} {
			if !containsLifecycleTestString(values, want) {
				t.Fatalf("dispatch must keep remote receipt blocker %q, values=%#v payload=%#v", want, values, payload)
			}
		}
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func assertConfigRolloutReceiptMatrixBlocked(t *testing.T, matrix []model.FindXAgentReceiptContractMatrixRow) {
	t.Helper()
	if len(matrix) == 0 {
		t.Fatalf("receipt_matrix is required")
	}
	wantScopes := []string{"writer", "reload", "restart", "drift", "rollback", "data_arrival", "evidence_chain"}
	for _, want := range wantScopes {
		found := false
		for _, row := range matrix {
			if row.Status != "blocked_by_contract" {
				t.Fatalf("receipt_matrix row must stay blocked_by_contract: %#v", row)
			}
			if containsLifecycleTestString(row.MissingContracts, "executor_disabled_contract") {
				t.Fatalf("receipt_matrix must not expose success or executor-disabled shortcut: %#v", row)
			}
			if row.Scope == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("receipt_matrix missing scope %q: %#v", want, matrix)
		}
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func anySliceToStrings(t *testing.T, raw any) []string {
	t.Helper()
	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected []any, got %#v", raw)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("expected string item, got %#v", item)
		}
		out = append(out, text)
	}
	return out
}

func TestFindXAgentPackagesExposePluginConfigMetadataIncludesHTTPProviderRefs(t *testing.T) {
	for _, pkg := range findXAgentPackages() {
		if pkg.PluginConfig == nil {
			continue
		}
		spec := pkg.PluginConfig
		for _, want := range []string{"plugin_config_writer_ref", "reload_command_ref", "reload_receipt_ref", "drift_check_ref", "evidence_chain_ref", "rollback_ref", "rollback_receipt_ref"} {
			if !containsLifecycleTestString(spec.RolloutMetadata, want) {
				t.Fatalf("package %s plugin_config rollout metadata missing lifecycle ref %q: %#v", pkg.ID, want, spec.RolloutMetadata)
			}
		}
		for _, forbidden := range []string{"plugin_config_writer", "reload_receipt", "drift_check", "rollback"} {
			if containsLifecycleTestString(spec.RolloutMetadata, forbidden) {
				t.Fatalf("package %s plugin_config rollout metadata exposes non-ref key %q: %#v", pkg.ID, forbidden, spec.RolloutMetadata)
			}
		}
		for _, want := range requiredHTTPProviderConfigRolloutRefs() {
			if !containsLifecycleTestString(spec.RolloutMetadata, want) {
				t.Fatalf("package %s plugin_config rollout metadata missing %q: %#v", pkg.ID, want, spec.RolloutMetadata)
			}
		}
		if len(spec.SourceEvidence) == 0 ||
			!containsLifecycleTestStringFragment(spec.SourceEvidence, "doc/provider.toml") ||
			!containsLifecycleTestStringFragment(spec.SourceEvidence, "inputs/http_provider.go") {
			t.Fatalf("package %s plugin_config missing provider source evidence: %#v", pkg.ID, spec.SourceEvidence)
		}
	}
}
