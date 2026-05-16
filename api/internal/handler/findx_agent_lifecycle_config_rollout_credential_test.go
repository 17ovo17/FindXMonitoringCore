package handler

import (
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentConfigRolloutCmdbCredentialGatedPluginMissingCredentialRef(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(cmdbCredentialRolloutBody("assign", "", "", false))

	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" || payload.Data.CredentialRefPresent {
		t.Fatalf("missing credential_ref should stay blocked without credential presence, code=%d payload=%#v", w.Code, payload)
	}
	for _, want := range []string{"credential_ref", cmdbAgentPluginCredentialContract, cmdbCredentialRefResolveContract} {
		if !containsLifecycleTestString(payload.MissingContracts, want) ||
			!containsLifecycleTestString(payload.ReceiptContract.MissingContracts, want) ||
			!containsLifecycleTestString(anySliceToStrings(t, payload.OperationContract["missing_contracts"]), want) {
			t.Fatalf("missing credential_ref path should include %q, payload=%#v", want, payload)
		}
	}
	contract, ok := payload.OperationContract["credential_contract"].(map[string]any)
	if !ok || contract["credential_ref_present"] != false || contract["credential_ref_resolved"] != false {
		t.Fatalf("credential contract should expose safe unresolved state: %#v", payload.OperationContract)
	}
	assertNoConfigRolloutCredentialSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbCredentialGatedPluginStaleCredentialRef(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(cmdbCredentialRolloutBody("assign", "stale-credential-ref", "", false))

	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" || !payload.Data.CredentialRefPresent {
		t.Fatalf("stale credential_ref should stay blocked with credential presence only, code=%d payload=%#v", w.Code, payload)
	}
	if !containsLifecycleTestString(payload.MissingContracts, cmdbCredentialRefResolveContract) {
		t.Fatalf("stale credential_ref must keep resolver contract missing, payload=%#v", payload)
	}
	contract, ok := payload.OperationContract["credential_contract"].(map[string]any)
	if !ok || contract["credential_ref_present"] != true || contract["credential_ref_resolved"] != false {
		t.Fatalf("stale credential contract should expose safe unresolved state: %#v", payload.OperationContract)
	}
	assertNoConfigRolloutCredentialSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbCredentialGatedPluginAssignResolvesCredentialRef(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	credentialID := "credential-fx-night-138"
	saveFindXAgentLifecycleCredentialForTest(credentialID)
	body := strings.NewReader(cmdbCredentialRolloutBody("assign", credentialID, "", false))

	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" || !payload.Data.CredentialRefPresent {
		t.Fatalf("resolved credential assign should remain blocked with safe credential presence, code=%d payload=%#v", w.Code, payload)
	}
	assertConfigRolloutCredentialResolved(t, payload.OperationContract, credentialID)
	assertCredentialResolveContractAbsent(t, payload.MissingContracts, payload.ReceiptContract.MissingContracts, anySliceToStrings(t, payload.OperationContract["missing_contracts"]))
	assertConfigRolloutCredentialPolicyBlocked(t, payload)
	for _, want := range []string{cmdbAgentPluginCredentialContract, "cmdb_plugin_config_schema_contract", "cmdb_dashboard_import_runtime_contract"} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("resolved credential should retain non-resolver blocker %q, payload=%#v", want, payload)
		}
	}
	assignmentRef, _ := payload.OperationContract["assignment_ref"].(string)
	if assignmentRef == "" {
		t.Fatalf("assign should still persist and expose assignment_ref: %#v", payload.OperationContract)
	}
	auditResp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
		Source:       "findx_audit",
		Scope:        "cmdb",
		ResourceType: "cmdb_agent_plugin_assignment",
		ResourceID:   assignmentRef,
		Action:       "cmdb.agent.plugin.assignment.save",
		Limit:        5,
	})
	if err != nil || len(auditResp.Items) == 0 {
		t.Fatalf("assignment audit should be queryable, items=%#v err=%v", auditResp.Items, err)
	}
	assertNoConfigRolloutCredentialSensitiveEcho(t, w.Body.String())
	if strings.Contains(w.Body.String(), credentialID) {
		t.Fatalf("resolved credential response must not echo credential_ref/id value %q: %s", credentialID, w.Body.String())
	}
	for _, item := range auditResp.Items {
		assertNoConfigRolloutCredentialSensitiveEcho(t, item.Body)
		for _, value := range item.Attributes {
			assertNoConfigRolloutCredentialSensitiveEcho(t, stringifyConfigRolloutCredentialAuditValue(value))
		}
	}
}

func TestFindXAgentConfigRolloutCmdbResolvedCredentialRequiresScopePolicyGate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	credentialID := "credential-fx-night-139"
	saveFindXAgentLifecycleCredentialForTest(credentialID)
	body := strings.NewReader(cmdbCredentialRolloutBody("assign", credentialID, "", false))

	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Status != "blocked" || payload.Data.Status != "blocked" {
		t.Fatalf("resolved credential without scope/policy must stay blocked, code=%d payload=%#v", w.Code, payload)
	}
	assertCredentialResolveContractAbsent(t, payload.MissingContracts, payload.ReceiptContract.MissingContracts, anySliceToStrings(t, payload.OperationContract["missing_contracts"]))
	assertConfigRolloutCredentialPolicyBlocked(t, payload)
	contract, ok := payload.OperationContract["credential_contract"].(map[string]any)
	if !ok {
		t.Fatalf("operation contract should include credential_contract: %#v", payload.OperationContract)
	}
	for key, want := range map[string]any{
		"credential_ref_resolved":  true,
		"scope_policy_required":    true,
		"scope_policy_resolved":    false,
		"scope_authorized":         false,
		"host_ref_present":         true,
		"agent_ref_present":        true,
		"business_context_present": false,
		"plugin_ref_present":       true,
		"provider_mode_present":    true,
		"policy_reason_code":       "scope_policy_contract_missing",
	} {
		if got := contract[key]; got != want {
			t.Fatalf("credential scope/policy summary %s=%#v, want %#v contract=%#v", key, got, want, contract)
		}
	}
	if payload.OperationContract["scope_authorized"] == true {
		t.Fatalf("operation contract must not mark resolved credential as authorized: %#v", payload.OperationContract)
	}
	assertNoConfigRolloutCredentialSensitiveEcho(t, w.Body.String())
	if strings.Contains(w.Body.String(), credentialID) {
		t.Fatalf("scope/policy response must not echo credential_ref/id value %q: %s", credentialID, w.Body.String())
	}
}

func TestFindXAgentConfigRolloutCmdbCredentialPolicyApprovalReleasesAssignScopeGate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	store.ResetCmdbApprovalRiskForTest()
	credentialID := "credential-fx-night-146-assign"
	approvalID := "approval-fx-night-146-assign"
	saveFindXAgentLifecycleCredentialForTest(credentialID)
	saveCmdbCredentialPolicyApprovalForTest(t, approvalID, "assign", nil, "review_accept_recorded")
	body := strings.NewReader(cmdbCredentialRolloutBodyWithMetadata("assign", credentialID, "", false, `"scope_policy_approval_ref":"`+approvalID+`"`))

	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Status != "blocked" || payload.Data.Status != "blocked" {
		t.Fatalf("approved credential policy should still fail closed on remaining contracts, code=%d payload=%#v", w.Code, payload)
	}
	assertConfigRolloutCredentialPolicyReleased(t, payload, false)
	for _, want := range []string{cmdbAgentPluginCredentialContract, "cmdb_plugin_config_schema_contract", "cmdb_dashboard_import_runtime_contract"} {
		if !containsLifecycleTestString(payload.MissingContracts, want) {
			t.Fatalf("policy approval must not remove non-policy blocker %q, payload=%#v", want, payload)
		}
	}
	contract, ok := payload.OperationContract["credential_contract"].(map[string]any)
	if !ok || contract["policy_approval_ref"] != approvalID {
		t.Fatalf("credential contract should expose safe approval ref only: %#v", payload.OperationContract)
	}
	if strings.Contains(w.Body.String(), credentialID) {
		t.Fatalf("policy approval response must not echo credential_ref/id value %q: %s", credentialID, w.Body.String())
	}
	assertNoConfigRolloutCredentialSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbCredentialPolicyApprovalRejectsMismatchedContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	store.ResetCmdbApprovalRiskForTest()
	credentialID := "credential-fx-night-146-mismatch"
	approvalID := "approval-fx-night-146-mismatch"
	saveFindXAgentLifecycleCredentialForTest(credentialID)
	saveCmdbCredentialPolicyApprovalForTest(t, approvalID, "assign", map[string]string{"plugin_id": "nginx"}, "review_accept_recorded")
	body := strings.NewReader(cmdbCredentialRolloutBodyWithMetadata("assign", credentialID, "", false, `"scope_policy_approval_ref":"`+approvalID+`"`))

	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Status != "blocked" || payload.Data.Status != "blocked" {
		t.Fatalf("mismatched approval should stay blocked, code=%d payload=%#v", w.Code, payload)
	}
	assertConfigRolloutCredentialPolicyBlocked(t, payload)
	contract, ok := payload.OperationContract["credential_contract"].(map[string]any)
	if !ok || contract["scope_policy_resolved"] != false || contract["scope_authorized"] != false {
		t.Fatalf("mismatched approval must not release scope policy: %#v", payload.OperationContract)
	}
	if strings.Contains(w.Body.String(), credentialID) {
		t.Fatalf("mismatched approval response must not echo credential_ref/id value %q: %s", credentialID, w.Body.String())
	}
	assertNoConfigRolloutCredentialSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbCredentialPolicyApprovalRejectsNonAcceptedDecision(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	store.ResetCmdbApprovalRiskForTest()
	credentialID := "credential-fx-night-146-denied"
	approvalID := "approval-fx-night-146-denied"
	saveFindXAgentLifecycleCredentialForTest(credentialID)
	saveCmdbCredentialPolicyApprovalForTest(t, approvalID, "assign", nil, "review_deny_recorded")
	body := strings.NewReader(cmdbCredentialRolloutBodyWithMetadata("assign", credentialID, "", false, `"scope_policy_approval_ref":"`+approvalID+`"`))

	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Status != "blocked" || payload.Data.Status != "blocked" {
		t.Fatalf("denied approval should stay blocked, code=%d payload=%#v", w.Code, payload)
	}
	assertConfigRolloutCredentialPolicyBlocked(t, payload)
	contract, ok := payload.OperationContract["credential_contract"].(map[string]any)
	if !ok || contract["scope_policy_resolved"] != false || contract["scope_authorized"] != false {
		t.Fatalf("denied approval must not release scope policy: %#v", payload.OperationContract)
	}
	assertNoConfigRolloutCredentialSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbCredentialPolicyApprovalReleasesDispatchScopeGateButKeepsReceiptsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	store.ResetCmdbApprovalRiskForTest()
	credentialID := "credential-fx-night-146-dispatch"
	approvalID := "approval-fx-night-146-dispatch"
	saveFindXAgentLifecycleCredentialForTest(credentialID)
	assignBody := strings.NewReader(cmdbCredentialRolloutBody("assign", credentialID, "", false))
	assignResp := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", assignBody, CreateFindXAgentConfigRollout)
	assignPayload := decodeConfigRolloutEnvelope(t, assignResp)
	assignmentRef, _ := assignPayload.OperationContract["assignment_ref"].(string)
	if assignmentRef == "" {
		t.Fatalf("assign should create assignment_ref before dispatch: %#v", assignPayload.OperationContract)
	}
	saveCmdbCredentialPolicyApprovalForTest(t, approvalID, "dispatch", nil, "review_accept_recorded")
	dispatchBody := strings.NewReader(cmdbCredentialRolloutBodyWithMetadata("dispatch", credentialID, assignmentRef, true, `"scope_policy_approval_ref":"`+approvalID+`"`))

	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", dispatchBody, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Status != "blocked" || payload.Data.Status != "blocked" {
		t.Fatalf("dispatch with approved credential policy should still fail closed on receipts, code=%d payload=%#v", w.Code, payload)
	}
	assertConfigRolloutCredentialPolicyReleased(t, payload, true)
	for _, values := range [][]string{
		payload.MissingContracts,
		payload.ReceiptContract.MissingContracts,
		anySliceToStrings(t, payload.OperationContract["missing_contracts"]),
	} {
		for _, want := range []string{
			"cmdb_agent_rollout_writer_receipt_contract",
			"cmdb_agent_rollout_delivery_receipt_contract",
			"cmdb_agent_rollout_effect_receipt_contract",
			"cmdb_agent_rollout_rollback_receipt_contract",
			"cmdb_agent_rollout_data_arrival_contract",
			"cmdb_agent_rollout_evidence_chain_contract",
		} {
			if !containsLifecycleTestString(values, want) {
				t.Fatalf("dispatch policy approval must keep receipt blocker %q, values=%#v payload=%#v", want, values, payload)
			}
		}
	}
	if strings.Contains(w.Body.String(), credentialID) {
		t.Fatalf("dispatch policy approval response must not echo credential_ref/id value %q: %s", credentialID, w.Body.String())
	}
	assertNoConfigRolloutCredentialSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutCmdbDispatchResolvesAssignmentAndCredentialButKeepsReceiptsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	credentialID := "credential-fx-night-138-dispatch"
	saveFindXAgentLifecycleCredentialForTest(credentialID)
	assignBody := strings.NewReader(cmdbCredentialRolloutBody("assign", credentialID, "", false))
	assignResp := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", assignBody, CreateFindXAgentConfigRollout)
	assignPayload := decodeConfigRolloutEnvelope(t, assignResp)
	assignmentRef, _ := assignPayload.OperationContract["assignment_ref"].(string)
	if assignmentRef == "" {
		t.Fatalf("assign should create assignment_ref before dispatch: %#v", assignPayload.OperationContract)
	}

	dispatchBody := strings.NewReader(cmdbCredentialRolloutBody("dispatch", credentialID, assignmentRef, true))
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", dispatchBody, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutEnvelope(t, w)

	if w.Code != http.StatusConflict || payload.Data.Status != "blocked" {
		t.Fatalf("dispatch should stay blocked after resolving safe refs, code=%d payload=%#v", w.Code, payload)
	}
	if got, _ := payload.OperationContract["assignment_ref"].(string); got != assignmentRef {
		t.Fatalf("dispatch should expose resolved assignment_ref, got=%q want=%q contract=%#v", got, assignmentRef, payload.OperationContract)
	}
	assertConfigRolloutCredentialResolved(t, payload.OperationContract, credentialID)
	if strings.Contains(w.Body.String(), credentialID) {
		t.Fatalf("resolved dispatch response must not echo credential_ref/id value %q: %s", credentialID, w.Body.String())
	}
	assertConfigRolloutCredentialPolicyBlocked(t, payload)
	for _, values := range [][]string{
		payload.MissingContracts,
		payload.ReceiptContract.MissingContracts,
		anySliceToStrings(t, payload.OperationContract["missing_contracts"]),
	} {
		if containsLifecycleTestString(values, cmdbAgentPluginAssignmentRefContract) {
			t.Fatalf("resolved assignment_ref should be removed from missing contracts, values=%#v payload=%#v", values, payload)
		}
		if containsLifecycleTestString(values, cmdbCredentialRefResolveContract) {
			t.Fatalf("resolved credential_ref should remove resolver contract, values=%#v payload=%#v", values, payload)
		}
		for _, want := range []string{
			cmdbCredentialDispatchPolicyGateContract,
			"cmdb_agent_rollout_writer_receipt_contract",
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
	assertNoConfigRolloutCredentialSensitiveEcho(t, w.Body.String())
}

func assertConfigRolloutCredentialPolicyReleased(t *testing.T, payload struct {
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
}, dispatch bool) {
	t.Helper()
	for _, values := range [][]string{
		payload.MissingContracts,
		payload.ReceiptContract.MissingContracts,
		anySliceToStrings(t, payload.OperationContract["missing_contracts"]),
	} {
		for _, forbidden := range []string{
			cmdbCredentialScopePolicyContract,
			cmdbAgentPluginCredentialScopeContract,
			cmdbAgentPluginCredentialPolicyContract,
			cmdbCredentialScopeAuthorizationContract,
		} {
			if containsLifecycleTestString(values, forbidden) {
				t.Fatalf("approved policy should release scope blocker %q, values=%#v payload=%#v", forbidden, values, payload)
			}
		}
		if dispatch && containsLifecycleTestString(values, cmdbCredentialDispatchPolicyGateContract) {
			t.Fatalf("approved dispatch policy should release dispatch policy gate, values=%#v payload=%#v", values, payload)
		}
	}
	contract, ok := payload.OperationContract["credential_contract"].(map[string]any)
	if !ok {
		t.Fatalf("operation contract should include credential_contract: %#v", payload.OperationContract)
	}
	for key, want := range map[string]any{
		"scope_policy_required": true,
		"scope_policy_resolved": true,
		"scope_authorized":      true,
		"policy_reason_code":    "credential_policy_release_recorded",
	} {
		if got := contract[key]; got != want {
			t.Fatalf("released credential policy summary %s=%#v, want %#v contract=%#v", key, got, want, contract)
		}
	}
	if payload.OperationContract["scope_authorized"] != true || payload.OperationContract["scope_policy_resolved"] != true {
		t.Fatalf("operation contract should expose safe scope authorization summary: %#v", payload.OperationContract)
	}
}

func assertConfigRolloutCredentialPolicyBlocked(t *testing.T, payload struct {
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
}) {
	t.Helper()
	for _, values := range [][]string{
		payload.MissingContracts,
		payload.ReceiptContract.MissingContracts,
		anySliceToStrings(t, payload.OperationContract["missing_contracts"]),
	} {
		for _, want := range []string{
			cmdbCredentialScopePolicyContract,
			cmdbAgentPluginCredentialScopeContract,
			cmdbAgentPluginCredentialPolicyContract,
			cmdbCredentialScopeAuthorizationContract,
		} {
			if !containsLifecycleTestString(values, want) {
				t.Fatalf("resolved credential must keep scope/policy blocker %q, values=%#v payload=%#v", want, values, payload)
			}
		}
	}
	contract, ok := payload.OperationContract["credential_contract"].(map[string]any)
	if !ok {
		t.Fatalf("operation contract should include credential_contract: %#v", payload.OperationContract)
	}
	if contract["scope_policy_required"] != true ||
		contract["scope_policy_resolved"] != false ||
		contract["scope_authorized"] != false ||
		contract["scope_policy_contract_id"] != "cmdb.credential.scope_policy.read.v1" {
		t.Fatalf("credential contract must expose blocked scope/policy summary only: %#v", contract)
	}
}

func cmdbCredentialRolloutBody(mode, credentialRef, assignmentRef string, remoteMutation bool) string {
	return cmdbCredentialRolloutBodyWithMetadata(mode, credentialRef, assignmentRef, remoteMutation, "")
}

func cmdbCredentialRolloutBodyWithMetadata(mode, credentialRef, assignmentRef string, remoteMutation bool, extraMetadata string) string {
	metadata := `"scope":"cmdb_host","cmdb_host_ref":"host-a","agent_ref":"agent-a","dashboard_refs":"dashboard:redis-overview","executor_ref":"executor-ref",` +
		completeRemotePreflightMetadata() + `,` +
		completePluginConfigRolloutMetadata() + `,"rollout_strategy_ref":"rollout-strategy-ref","rollout_receipt_ref":"rollout-receipt-ref"` +
		`,"token":"credential-token-value","dsn":"mysql://user:pass@host/db","ssh_key":"-----BEGIN PRIVATE KEY-----"`
	if assignmentRef != "" {
		metadata += `,"assignment_ref":"` + assignmentRef + `"`
	}
	if extraMetadata != "" {
		metadata += `,` + extraMetadata
	}
	body := `{"template_id":"cmdb-host-plugin-` + mode + `","target_ids":["host-a"],"agent_ids":["agent-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"cmdb_host_probe","plugin_id":"redis","reload_strategy":"local-reload","rollout_strategy":"` + mode + `","rollback_ref":"rollback-ref","remote_mutation":` + boolString(remoteMutation)
	if credentialRef != "" {
		body += `,"credential_ref":"` + credentialRef + `"`
	}
	body += `,"metadata":{` + metadata + `}}`
	return body
}

func saveFindXAgentLifecycleCredentialForTest(id string) {
	store.SaveCredential(&model.Credential{
		ID:       id,
		Name:     "FX-NIGHT-138 Credential",
		Protocol: "ssh",
		Username: "root-user",
		Password: "credential-password-value",
		SSHKey:   "credential-ssh-key-value",
		Port:     22,
		Remark:   "credential-token-value mysql://user:pass@host/db",
	})
}

func saveCmdbCredentialPolicyApprovalForTest(t *testing.T, id, mode string, overrides map[string]string, status string) {
	t.Helper()
	context := map[string]string{
		"host_ref":            "host-a",
		"agent_ref":           "agent-a",
		"plugin_id":           "redis",
		"provider_mode":       "cmdb_host_probe",
		"operation_mode":      mode,
		"credential_protocol": "ssh",
	}
	for key, value := range overrides {
		context[key] = value
	}
	raw := "{"
	first := true
	for key, value := range context {
		if !first {
			raw += ","
		}
		first = false
		raw += `"` + key + `":"` + value + `"`
	}
	raw += "}"
	_, err := store.CreateCmdbResourceApproval(&model.CmdbResourceApproval{
		ID:            id,
		View:          "archive",
		Requester:     "cmdb-operator",
		Approver:      "cmdb-operator",
		ResourceType:  "cmdb_agent_plugin_credential_policy",
		ResourceID:    "host-a:agent-a:redis",
		Action:        "cmdb.agent.plugin.credential_policy.release",
		RiskLevel:     "medium",
		Status:        status,
		WorkflowState: status,
		Title:         "CMDB credential policy release",
		Summary:       "Credential scope policy review recorded",
		ContextJSON:   raw,
		DecisionActor: "cmdb-operator",
		DecisionNote:  "review recorded",
		BusinessGroup: "",
		AuditRef:      "approval-audit-" + id,
		RiskRecordID:  "risk-" + id,
	})
	if err != nil {
		t.Fatalf("save approval fixture: %v", err)
	}
}

func assertConfigRolloutCredentialResolved(t *testing.T, operationContract map[string]any, credentialID string) {
	t.Helper()
	if operationContract["credential_ref_resolved"] != true {
		t.Fatalf("operation contract should expose safe credential resolution only: %#v", operationContract)
	}
	if _, ok := operationContract["credential_id"]; ok {
		t.Fatalf("operation contract must not echo credential_ref/id value %q: %#v", credentialID, operationContract)
	}
	contract, ok := operationContract["credential_contract"].(map[string]any)
	if !ok {
		t.Fatalf("operation contract should include credential_contract: %#v", operationContract)
	}
	if contract["credential_ref_resolved"] != true {
		t.Fatalf("credential contract should expose safe resolved state: %#v", contract)
	}
	if _, ok := contract["credential_id"]; ok {
		t.Fatalf("credential contract must not echo credential_ref/id value %q: %#v", credentialID, contract)
	}
	if contract["protocol"] != "ssh" || contract["port_present"] != true {
		t.Fatalf("credential contract should expose only safe protocol/port summary: %#v", contract)
	}
}

func assertCredentialResolveContractAbsent(t *testing.T, groups ...[]string) {
	t.Helper()
	for _, values := range groups {
		if containsLifecycleTestString(values, cmdbCredentialRefResolveContract) {
			t.Fatalf("resolved credential_ref should remove %s from missing contracts: %#v", cmdbCredentialRefResolveContract, values)
		}
	}
}

func assertNoConfigRolloutCredentialSensitiveEcho(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{
		`"credential_ref":`,
		`"credential_id":`,
		"credential-password-value",
		"credential-ssh-key-value",
		"credential-token-value",
		"mysql://",
		"root-user",
		"ssh_key",
		"password",
		"token",
		"dsn",
		"private key",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("response/audit must not echo credential sensitive data %q: %s", forbidden, body)
		}
	}
}

func stringifyConfigRolloutCredentialAuditValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case map[string]any:
		parts := make([]string, 0, len(typed))
		for key, item := range typed {
			parts = append(parts, key+"="+stringifyConfigRolloutCredentialAuditValue(item))
		}
		return strings.Join(parts, " ")
	default:
		return ""
	}
}
