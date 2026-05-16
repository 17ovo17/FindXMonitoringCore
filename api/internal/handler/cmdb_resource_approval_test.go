package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbResourceApprovalsFailCloseWithoutRuntime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetCmdbApprovalRuntimeForTest(t)
	router := gin.New()
	router.GET("/approvals", GetCmdbResourceApprovals)

	for _, view := range []string{"mine", "todo", "archive"} {
		t.Run(view, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/approvals?view="+view, nil))

			if w.Code != http.StatusConflict {
				t.Fatalf("status = %d, want 409, body=%s", w.Code, w.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body["contract_id"] != "cmdb.resource.approval.runtime.v1" || body["status"] != cmdbBlockedByContract {
				t.Fatalf("unexpected approval contract response: %#v", body)
			}
			if _, ok := body["approval_requests"]; ok {
				t.Fatalf("blocked approval response must not expose fake approval_requests: %s", w.Body.String())
			}
			assertCmdbMissingContract(t, body, "cmdb_resource_approval_store_contract")
			assertCmdbMissingContract(t, body, "cmdb_resource_approval_workflow_contract")
			assertCmdbMissingContract(t, body, "cmdb_operation_risk_policy_contract")
			assertCmdbMissingContract(t, body, "cmdb_approval_audit_receipt_contract")
			bodyText := strings.ToLower(w.Body.String())
			for _, forbidden := range []string{"approved", "succeeded", "success", "queued", "running", "installed", "password", "token", "cookie"} {
				if strings.Contains(bodyText, forbidden) {
					t.Fatalf("approval blocked response leaked fake state or sensitive word %q: %s", forbidden, w.Body.String())
				}
			}
		})
	}
}

func TestCmdbHighRiskOperationsExposeApprovalAndRiskPolicyContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetCmdbApprovalRuntimeForTest(t)
	hostID := createCmdbHostOpsFixture(t)
	router := gin.New()
	router.GET("/hosts/:id/terminal", CmdbHostTerminal)
	router.POST("/hosts/:id/upload", CmdbHostUpload)
	router.POST("/hosts/:id/exec", CmdbHostExec)
	router.POST("/deploy-tasks", CmdbCreateDeployTask)

	uploadBody, uploadContentType := cmdbHostOpsMultipartBody(t, "deploy.sh", "token=<TOKEN>\n")
	cases := []struct {
		name        string
		method      string
		path        string
		body        *bytes.Reader
		contentType string
		contractID  string
	}{
		{name: "terminal", method: http.MethodGet, path: "/hosts/" + hostID + "/terminal", body: bytes.NewReader(nil), contractID: "cmdb.host.terminal.v1"},
		{name: "upload", method: http.MethodPost, path: "/hosts/" + hostID + "/upload", body: bytes.NewReader(uploadBody.Bytes()), contentType: uploadContentType, contractID: "cmdb.host.file_upload.v1"},
		{name: "exec", method: http.MethodPost, path: "/hosts/" + hostID + "/exec", body: bytes.NewReader([]byte(`{"command":"echo password=<SECRET> token=<TOKEN>","timeout":10}`)), contentType: "application/json", contractID: "cmdb.host.command_exec.v1"},
		{name: "deploy", method: http.MethodPost, path: "/deploy-tasks", body: bytes.NewReader([]byte(`{"name":"deploy-128n","target_hosts":["` + hostID + `"],"script":"echo token=<TOKEN>"}`)), contentType: "application/json", contractID: "cmdb.deploy.executor.v1"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, tt.body)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			body := assertCmdbHighRiskBlockedEnvelope(t, w, tt.contractID)
			assertCmdbMissingContract(t, body, "cmdb_resource_approval_runtime_contract")
			assertCmdbMissingContract(t, body, "cmdb_operation_risk_policy_contract")
			assertCmdbMissingContract(t, body, "cmdb_action_preflight_contract")
			assertCmdbMissingContract(t, body, "cmdb_action_audit_receipt_contract")
			text := strings.ToLower(w.Body.String())
			for _, forbidden := range []string{"password=<secret>", "token=<token>", `"success":true`, `"exit_code":0`, "succeeded", "running", "queued", "installed"} {
				if strings.Contains(text, forbidden) {
					t.Fatalf("high risk blocked response leaked fake state or sensitive marker %q: %s", forbidden, w.Body.String())
				}
			}
		})
	}
}

func TestCmdbResourceApprovalsListStoredHighRiskRequestsWithContractGaps(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetCmdbApprovalRuntimeForTest(t)
	actor := "approval-list-actor"
	hostID := createCmdbHostOpsFixture(t)
	router := cmdbApprovalTestRouter(actor)
	router.GET("/hosts/:id/terminal", CmdbHostTerminal)
	router.GET("/approvals", GetCmdbResourceApprovals)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/hosts/"+hostID+"/terminal", nil))
	gate := assertCmdbHighRiskBlockedEnvelope(t, w, "cmdb.host.terminal.v1")
	approvalID := cmdbApprovalIDFromGate(t, gate)
	if approvalID == "" {
		t.Fatal("high-risk gate did not return an approval id")
	}
	riskID := cmdbRiskIDFromGate(t, gate)
	if riskID == "" {
		t.Fatal("high-risk gate did not return a risk id")
	}
	if runtime, ok := gate["approval_runtime"].(map[string]any); !ok || runtime["status"] != "ready_with_contract_gaps" || runtime["execution_state"] != cmdbBlockedByContract {
		t.Fatalf("approval runtime must be queryable but execution-blocked, got %#v", gate["approval_runtime"])
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/approvals?view=mine", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode approval list: %v", err)
	}
	if body["status"] != "ready_with_contract_gaps" || body["contract_id"] != cmdbResourceApprovalRuntimeContract {
		t.Fatalf("unexpected approval list envelope: %#v", body)
	}
	requests, ok := body["approval_requests"].([]any)
	if !ok || len(requests) == 0 {
		t.Fatalf("approval list must expose stored approval_requests, body=%s", w.Body.String())
	}
	first, ok := requests[0].(map[string]any)
	if !ok {
		t.Fatalf("approval request has unexpected type: %#v", requests[0])
	}
	if first["id"] != approvalID || first["risk_record_id"] != riskID || first["requester"] != actor || first["status"] != "pending_review" {
		t.Fatalf("approval request does not match stored high-risk request: %#v", first)
	}
	assertCmdbMissingContract(t, body, "cmdb_resource_approval_workflow_contract")
	assertCmdbMissingContract(t, body, "cmdb_operation_risk_policy_contract")
	assertCmdbMissingContract(t, body, "cmdb_action_preflight_contract")
	assertCmdbMissingContract(t, body, "cmdb_action_audit_receipt_contract")
	assertNoCmdbApprovalFakeStateOrSecret(t, w.Body.String())
}

func TestCmdbResourceApprovalDetailReturnsStoredRequestWithoutExecutionState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetCmdbApprovalRuntimeForTest(t)
	hostID := createCmdbHostOpsFixture(t)
	router := cmdbApprovalTestRouter("approval-detail-actor")
	router.POST("/hosts/:id/exec", CmdbHostExec)
	router.GET("/approvals/:id", GetCmdbResourceApproval)

	req := httptest.NewRequest(http.MethodPost, "/hosts/"+hostID+"/exec", bytes.NewReader([]byte(`{"command":"echo password=<SECRET> token=<TOKEN>","timeout":10}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	gate := assertCmdbHighRiskBlockedEnvelope(t, w, "cmdb.host.command_exec.v1")
	approvalID := cmdbApprovalIDFromGate(t, gate)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/approvals/"+approvalID, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode approval detail: %v", err)
	}
	request, ok := body["approval_request"].(map[string]any)
	if !ok {
		t.Fatalf("approval detail missing stored approval_request: %s", w.Body.String())
	}
	if request["id"] != approvalID || request["action"] != "exec" || request["status"] != "pending_review" {
		t.Fatalf("unexpected approval detail request: %#v", request)
	}
	context, ok := request["context"].(map[string]any)
	if !ok || context["command_digest"] == "" || context["command_length"] == "" {
		t.Fatalf("approval detail must keep command digest/length evidence, got %#v", request["context"])
	}
	assertCmdbMissingContract(t, body, "cmdb_operation_risk_policy_contract")
	assertCmdbMissingContract(t, body, "cmdb_action_preflight_contract")
	assertCmdbMissingContract(t, body, "cmdb_action_audit_receipt_contract")
	assertNoCmdbApprovalFakeStateOrSecret(t, w.Body.String())

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/approvals/missing-approval-id", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing approval status = %d, want 404, body=%s", w.Code, w.Body.String())
	}
}

func TestCmdbResourceApprovalReviewRecordsDecisionButKeepsExecutionBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetCmdbApprovalRuntimeForTest(t)
	hostID := createCmdbHostOpsFixture(t)
	router := cmdbApprovalTestRouter("cmdb-operator")
	router.GET("/hosts/:id/terminal", CmdbHostTerminal)
	router.POST("/approvals/:id/review", ReviewCmdbResourceApproval)
	router.GET("/approvals", GetCmdbResourceApprovals)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/hosts/"+hostID+"/terminal", nil))
	gate := assertCmdbHighRiskBlockedEnvelope(t, w, "cmdb.host.terminal.v1")
	approvalID := cmdbApprovalIDFromGate(t, gate)

	w = httptest.NewRecorder()
	reviewBody := bytes.NewReader([]byte(`{"decision":"accept","note":"approved after checking password=<SECRET> token=<TOKEN>"}`))
	req := httptest.NewRequest(http.MethodPost, "/approvals/"+approvalID+"/review", reviewBody)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode approval review: %v", err)
	}
	if body["code"] != cmdbBlockedByContract || body["status"] != cmdbBlockedByContract {
		t.Fatalf("review must remain execution-blocked, body=%#v", body)
	}
	request, ok := body["approval_request"].(map[string]any)
	if !ok || request["id"] != approvalID || request["status"] != "review_accept_recorded" || request["workflow_state"] != "review_accept_recorded" {
		t.Fatalf("review did not record decision on approval request: %#v", body["approval_request"])
	}
	assertCmdbMissingContract(t, body, "cmdb_approval_execution_release_contract")
	assertNoCmdbApprovalFakeStateOrSecret(t, w.Body.String())

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/approvals?view=archive", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("archive status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), approvalID) || !strings.Contains(w.Body.String(), "review_accept_recorded") {
		t.Fatalf("archive list must include reviewed approval, body=%s", w.Body.String())
	}
	assertNoCmdbApprovalFakeStateOrSecret(t, w.Body.String())

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/approvals/"+approvalID+"/review", bytes.NewReader([]byte(`{"decision":"reject","note":"second write"}`)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("second review status = %d, want 409, body=%s", w.Code, w.Body.String())
	}
	var second map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &second); err != nil {
		t.Fatalf("decode second review: %v", err)
	}
	if _, exists := second["approval_request"]; exists {
		t.Fatalf("second review must not overwrite or return a fake updated approval_request: %s", w.Body.String())
	}
	assertCmdbMissingContract(t, second, "cmdb_approval_decision_contract")
}

func TestCmdbCredentialPolicyApprovalRequestRecordsPendingReviewWithoutRelease(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	resetCmdbApprovalRuntimeForTest(t)
	credentialID := "credential-fx-night-147-request"
	saveFindXAgentLifecycleCredentialForTest(credentialID)
	router := cmdbApprovalTestRouter("credential-policy-requester")
	router.POST("/approvals", CreateCmdbResourceApprovalRequest)
	router.POST("/approvals/:id/review", ReviewCmdbResourceApproval)

	requestBody := bytes.NewReader([]byte(`{
		"resource_type":"cmdb_agent_plugin_credential_policy",
		"action":"cmdb.agent.plugin.credential_policy.release",
		"host_ref":"host-a",
		"agent_ref":"agent-a",
		"plugin_id":"redis",
		"provider_mode":"cmdb_host_probe",
		"operation_mode":"assign",
		"credential_protocol":"ssh",
		"business_group_ref":"finance-core",
		"team_ref":"sre-team",
		"reason":"release credential policy for host plugin; password=<SECRET> token=<TOKEN>",
		"credential_ref":"` + credentialID + `"
	}`))
	req := httptest.NewRequest(http.MethodPost, "/approvals", requestBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("approval request status = %d, want 409 blocked, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode approval request: %v", err)
	}
	if body["code"] != cmdbBlockedByContract || body["status"] != cmdbBlockedByContract || body["contract_id"] != cmdbResourceApprovalRuntimeContract {
		t.Fatalf("approval request must return blocked contract envelope, body=%#v", body)
	}
	approvalID := cmdbApprovalIDFromGate(t, body)
	request, ok := body["approval_request"].(map[string]any)
	if !ok || request["resource_type"] != "cmdb_agent_plugin_credential_policy" ||
		request["resource_id"] != "host-a:agent-a:redis" ||
		request["action"] != "cmdb.agent.plugin.credential_policy.release" ||
		request["status"] != "pending_review" ||
		request["workflow_state"] != "pending_review" {
		t.Fatalf("credential policy approval request not persisted as pending review: %#v", request)
	}
	context, ok := request["context"].(map[string]any)
	if !ok ||
		context["host_ref"] != "host-a" ||
		context["agent_ref"] != "agent-a" ||
		context["plugin_id"] != "redis" ||
		context["provider_mode"] != "cmdb_host_probe" ||
		context["operation_mode"] != "assign" ||
		context["credential_protocol"] != "ssh" ||
		context["business_group_ref"] != "finance-core" ||
		context["team_ref"] != "sre-team" {
		t.Fatalf("credential policy approval context mismatch: %#v", request["context"])
	}
	assertCmdbMissingContract(t, body, "cmdb_approval_decision_contract")
	assertCmdbMissingContract(t, body, "cmdb_approval_execution_release_contract")
	assertNoCmdbApprovalFakeStateOrSecret(t, w.Body.String())
	if strings.Contains(w.Body.String(), credentialID) {
		t.Fatalf("approval request must not echo credential ref/id value %q: %s", credentialID, w.Body.String())
	}

	pendingRollout := strings.NewReader(cmdbCredentialRolloutBodyWithMetadata("assign", credentialID, "", false,
		`"scope_policy_approval_ref":"`+approvalID+`","business_group_ref":"finance-core","team_ref":"sre-team"`))
	pendingResp := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", pendingRollout, CreateFindXAgentConfigRollout)
	pendingPayload := decodeConfigRolloutEnvelope(t, pendingResp)
	assertConfigRolloutCredentialPolicyBlocked(t, pendingPayload)

	reviewReq := httptest.NewRequest(http.MethodPost, "/approvals/"+approvalID+"/review", bytes.NewReader([]byte(`{"decision":"accept","note":"policy checked"}`)))
	reviewReq.Header.Set("Content-Type", "application/json")
	reviewResp := httptest.NewRecorder()
	router.ServeHTTP(reviewResp, reviewReq)
	if reviewResp.Code != http.StatusConflict {
		t.Fatalf("review status = %d, want 409 blocked, body=%s", reviewResp.Code, reviewResp.Body.String())
	}

	releasedRollout := strings.NewReader(cmdbCredentialRolloutBodyWithMetadata("assign", credentialID, "", false,
		`"scope_policy_approval_ref":"`+approvalID+`","business_group_ref":"finance-core","team_ref":"sre-team"`))
	releasedResp := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", releasedRollout, CreateFindXAgentConfigRollout)
	releasedPayload := decodeConfigRolloutEnvelope(t, releasedResp)
	assertConfigRolloutCredentialPolicyReleased(t, releasedPayload, false)
	assertNoCmdbApprovalFakeStateOrSecret(t, releasedResp.Body.String())
	if strings.Contains(releasedResp.Body.String(), credentialID) {
		t.Fatalf("released rollout must not echo credential ref/id value %q: %s", credentialID, releasedResp.Body.String())
	}
}

func TestCmdbCredentialPolicyApprovalRequestRejectsInvalidContextAndSensitiveEcho(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetCmdbApprovalRuntimeForTest(t)
	router := cmdbApprovalTestRouter("credential-policy-requester")
	router.POST("/approvals", CreateCmdbResourceApprovalRequest)

	req := httptest.NewRequest(http.MethodPost, "/approvals", bytes.NewReader([]byte(`{
		"resource_type":"cmdb_agent_plugin_credential_policy",
		"action":"cmdb.agent.plugin.credential_policy.release",
		"host_ref":"host-a",
		"agent_ref":"",
		"plugin_id":"redis",
		"provider_mode":"cmdb_host_probe",
		"operation_mode":"delivered",
		"credential_protocol":"ssh",
		"reason":"bad token=<TOKEN> password=<SECRET>"
	}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("invalid approval request status = %d, want 409, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode invalid approval request: %v", err)
	}
	if _, exists := body["approval_request"]; exists {
		t.Fatalf("invalid approval request must not persist or return approval_request: %s", w.Body.String())
	}
	assertCmdbMissingContract(t, body, "cmdb_credential_policy_approval_context_contract")
	assertNoCmdbApprovalFakeStateOrSecret(t, w.Body.String())
}

func assertCmdbHighRiskBlockedEnvelope(t *testing.T, w *httptest.ResponseRecorder, contractID string) map[string]any {
	t.Helper()
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	var top map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &top); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if gate, ok := top["gate"].(map[string]any); ok {
		if gate["code"] != cmdbBlockedByContract {
			t.Fatalf("gate.code = %v, want %s, body=%s", gate["code"], cmdbBlockedByContract, w.Body.String())
		}
		if gate["contract_id"] != contractID {
			t.Fatalf("gate.contract_id = %v, want %s, body=%s", gate["contract_id"], contractID, w.Body.String())
		}
		return gate
	}
	if top["code"] != cmdbBlockedByContract {
		t.Fatalf("code = %v, want %s, body=%s", top["code"], cmdbBlockedByContract, w.Body.String())
	}
	if top["contract_id"] != contractID {
		t.Fatalf("contract_id = %v, want %s, body=%s", top["contract_id"], contractID, w.Body.String())
	}
	return top
}

func TestCmdbImportConfirmRequiresApprovalRuntime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetCmdbApprovalRuntimeForTest(t)
	router := gin.New()
	router.POST("/import/confirm", CmdbImportConfirm)

	req := httptest.NewRequest(http.MethodPost, "/import/confirm", bytes.NewReader([]byte(`{"hosts":[{"name":"128n-host","ip_address":"10.128.9.1","password":"raw-secret"}]}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := assertCmdbBlockedResponse(t, w, "cmdb.import.confirm.v1")
	assertCmdbMissingContract(t, body, "cmdb_resource_approval_runtime_contract")
	assertCmdbMissingContract(t, body, "cmdb_operation_risk_policy_contract")
	assertCmdbMissingContract(t, body, "cmdb_action_preflight_contract")
	assertCmdbMissingContract(t, body, "cmdb_action_audit_receipt_contract")
	assertCmdbMissingContract(t, body, "cmdb_import_preview_diff_contract")
	assertCmdbMissingContract(t, body, "cmdb_import_rollback_snapshot_contract")
	assertCmdbMissingContract(t, body, "cmdb_import_write_audit_contract")
	if _, ok := store.GetCmdbInstance("128n-host"); ok {
		t.Fatal("blocked import confirm must not create a fake CMDB instance")
	}
	for _, forbidden := range []string{"raw-secret", `"created":1`, `"success":true`} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("blocked import confirm leaked fake success or sensitive marker %q: %s", forbidden, w.Body.String())
		}
	}
}

func resetCmdbApprovalRuntimeForTest(t *testing.T) {
	t.Helper()
	store.ResetCmdbApprovalRiskForTest()
	store.ResetCmdbDeployTasksForTest()
	t.Cleanup(func() {
		store.ResetCmdbApprovalRiskForTest()
		store.ResetCmdbDeployTasksForTest()
	})
}

func cmdbApprovalTestRouter(actor string) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("username", actor)
		c.Next()
	})
	return router
}

func cmdbApprovalIDFromGate(t *testing.T, gate map[string]any) string {
	t.Helper()
	request, ok := gate["approval_request"].(map[string]any)
	if !ok {
		t.Fatalf("gate missing approval_request: %#v", gate)
	}
	id, ok := request["id"].(string)
	if !ok || id == "" {
		t.Fatalf("approval_request.id missing: %#v", request)
	}
	return id
}

func cmdbRiskIDFromGate(t *testing.T, gate map[string]any) string {
	t.Helper()
	record, ok := gate["risk_record"].(map[string]any)
	if !ok {
		t.Fatalf("gate missing risk_record: %#v", gate)
	}
	id, ok := record["id"].(string)
	if !ok || id == "" {
		t.Fatalf("risk_record.id missing: %#v", record)
	}
	return id
}

func assertNoCmdbApprovalFakeStateOrSecret(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{
		`"success":true`,
		`"exit_code":0`,
		"queued",
		"running",
		"installed",
		"applied",
		"data_arrived",
		"service_registered",
		"rolled_back",
		"uninstalled",
		"delivered",
		"effective",
		"succeeded",
		"password=<secret>",
		"token=<token>",
		"cookie",
		"private_key",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("approval response leaked fake state or sensitive marker %q: %s", forbidden, body)
		}
	}
}

func TestMonitorPermissionApprovalReviewKnownAndUserDenied(t *testing.T) {
	checker := model.NewMonitorPermissionChecker()
	for _, action := range []string{"approve", "review"} {
		user := checker.Check("user", "cmdb.approval", action)
		if !user.Known || user.Allowed || user.Reason != model.MonitorPermissionReasonRoleDenied {
			t.Fatalf("user cmdb.approval:%s = known:%v allowed:%v reason:%s; want denied known permission", action, user.Known, user.Allowed, user.Reason)
		}
		admin := checker.Check("admin", "cmdb.approval", action)
		if !admin.Known || !admin.Allowed || admin.Reason != model.MonitorPermissionReasonRoleAllowed {
			t.Fatalf("admin cmdb.approval:%s = known:%v allowed:%v reason:%s; want allowed", action, admin.Known, admin.Allowed, admin.Reason)
		}
	}
}
