package handler

import (
	"net/http"
	"strings"
	"testing"

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
	for _, want := range []string{"BLOCKED_BY_CONTRACT", "config_snippet_ref", "config_version", "executor_ref", "rollback_ref"} {
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
	want := "BLOCKED_BY_CONTRACT: missing audit_ref_or_evidence_chain_ref, credential_ref, execution_receipt_ref_or_receipt_ref, idempotency_key, target_os, timeout_policy_ref, transport_or_runner"
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
	for _, want := range []string{"drift_check_ref", "evidence_chain_ref", "plugin_config_writer_ref", "reload_command_ref", "reload_receipt_ref", "rollback_receipt_ref"} {
		if !strings.Contains(payload.Error, want) || !strings.Contains(payload.Data.Blocker, want) {
			t.Fatalf("plugin reload blocker should include %q, payload=%#v", want, payload)
		}
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
	if payload.Error != "BLOCKED_BY_CONTRACT: executor not enabled / config rollout protocol not open" {
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
	want := "BLOCKED_BY_CONTRACT: missing cluster_ref, config_map_ref, data_arrival_validator_ref, drift_check_ref, executor_ref, helm_chart_ref_or_manifest_bundle_ref, namespace_ref, reload_receipt_ref, rollout_receipt_ref, rollout_strategy_ref, workload_selector_ref"
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
	want := "BLOCKED_BY_CONTRACT: missing helm_chart_ref_or_manifest_bundle_ref, helm_release_ref"
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
	if payload.Error != "BLOCKED_BY_CONTRACT: executor not enabled / config rollout protocol not open" {
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
	body := strings.NewReader(`{"template_id":"host-plugin","target_ids":["target-a"],"remote_mutation":true}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)
	if w.Code != http.StatusConflict || payload.Status != "blocked" {
		t.Fatalf("config rollout should return blocked envelope, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{"BLOCKED_BY_CONTRACT", "MISSING_CONTRACTS"} {
		if !containsLifecycleTestString(payload.Blockers, want) {
			t.Fatalf("blocked envelope should include %q, payload=%#v", want, payload)
		}
	}
	for _, want := range []string{"config_snippet_ref", "config_version", "executor_ref", "reload_receipt_ref"} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("missing_contracts should include %q, payload=%#v", want, payload)
		}
	}
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
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
	if !containsLifecycleTestString(payload.Blockers, "EXECUTOR_DISABLED_BY_CONTRACT") {
		t.Fatalf("complete refs should expose executor disabled blocker, payload=%#v", payload)
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
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
