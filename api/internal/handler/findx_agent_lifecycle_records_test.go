package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentConfigRolloutPersistsBlockedRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"template_id":"host-plugin",
		"target_ids":["target-a"],
		"plugin_id":"input.cpu",
		"config_snippet_ref":"<CONFIG_SNIPPET_REF>",
		"provider_mode":"http",
		"remote_mutation":true,
		"canary_percent":25,
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{"change":"CHG-1","cookie":"secret"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	if w.Code != http.StatusConflict {
		t.Fatalf("config rollout should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	var payload struct {
		Data model.FindXAgentConfigRollout `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if payload.Data.ID == "" || payload.Data.Status != "blocked" || !payload.Data.RemoteMutation || !payload.Data.CredentialRefPresent {
		t.Fatalf("blocked rollout should persist request metadata: %#v", payload.Data)
	}
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutListCoercesNonBlockedRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	blocked, err := store.SaveFindXAgentConfigRollout(model.FindXAgentConfigRollout{
		TemplateID: "host-plugin",
		TargetIDs:  []string{"target-blocked"},
		Status:     "blocked",
		Blocker:    "PENDING: executor not enabled",
		Metadata:   map[string]string{"evidence_chain_ref": "evidence-ref"},
	})
	if err != nil {
		t.Fatalf("save blocked rollout: %v", err)
	}
	nonBlocked, err := store.SaveFindXAgentConfigRollout(model.FindXAgentConfigRollout{
		TemplateID: "host-plugin",
		TargetIDs:  []string{"target-queued"},
		Status:     "queued",
		Metadata:   map[string]string{"evidence_chain_ref": "queued-evidence-ref"},
	})
	if err != nil {
		t.Fatalf("save non-blocked rollout: %v", err)
	}
	if nonBlocked.Status != "blocked" || nonBlocked.Blocker == "" {
		t.Fatalf("store must coerce config rollout to blocked, got %#v", nonBlocked)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts", ListFindXAgentConfigRollouts)
	if w.Code != http.StatusOK {
		t.Fatalf("config rollout list should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, blocked.ID) || !strings.Contains(body, nonBlocked.ID) {
		t.Fatalf("list should include store-coerced blocked rollouts, body=%s", body)
	}
	assertNoConfigRolloutExecutionStates(t, body)
	assertNoConfigRolloutSensitiveEcho(t, body)
}

func TestFindXAgentConfigRolloutDetailReturnsBlockedSafeRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"template_id":"host-plugin","target_ids":["target-a"],"config_version":"cfg-v1","config_snippet_ref":"snippet-ref","config_format":"toml","provider_mode":"http","plugin_id":"input.cpu","reload_strategy":"hup","rollback_ref":"rollback-ref","remote_mutation":true,"change_ticket":"CHG-1","audit_reason":"planned rollout","credential_ref":"<CREDENTIAL_REF>","metadata":{"executor_ref":"executor-ref",` + completeRemotePreflightMetadata() + `,` + completePluginConfigRolloutMetadata() + `,` + completeHTTPProviderConfigRolloutMetadata() + `,"token":"secret","cookie":"session-secret","credential_ref":"credential-secret","dsn":"mysql://user:pass@host/db","private_key":"secret-private-key","provider_auth_token":"secret-token","provider_password_ref":"secret-password"}}`)
	post := performAgentLifecyclePost("/api/v1/findx-agents/config-rollouts", body, CreateFindXAgentConfigRollout)
	payload := decodeConfigRolloutResponse(t, post)
	if post.Code != http.StatusConflict || payload.Data.ID == "" {
		t.Fatalf("blocked rollout should be created before detail query, code=%d body=%s", post.Code, post.Body.String())
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+payload.Data.ID, ListFindXAgentConfigRollouts)
	if w.Code != http.StatusOK {
		t.Fatalf("config rollout detail should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	var detail model.FindXAgentConfigRollout
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("invalid detail response: %v body=%s", err, w.Body.String())
	}
	if detail.ID != payload.Data.ID || detail.Status != "blocked" || !strings.Contains(detail.Blocker, "PENDING") {
		t.Fatalf("detail should return the blocked rollout record, got %#v", detail)
	}
	for _, key := range []string{
		"plugin_config_writer_ref",
		"reload_command_ref",
		"reload_receipt_ref",
		"drift_check_ref",
		"evidence_chain_ref",
		"rollback_receipt_ref",
		"provider_auth_ref",
		"checksum_registry_ref",
		"config_serving_receipt_ref",
	} {
		if detail.Metadata[key] == "" {
			t.Fatalf("detail should retain safe metadata ref %s, metadata=%#v", key, detail.Metadata)
		}
	}
	for _, key := range []string{"token", "cookie", "session", "dsn", "password", "private_key", "credential_ref"} {
		if detail.Metadata[key] != "" {
			t.Fatalf("detail must not expose sensitive metadata key %s, metadata=%#v", key, detail.Metadata)
		}
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentConfigRolloutDetailCoercesNonBlockedAndMissingReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	nonBlocked, err := store.SaveFindXAgentConfigRollout(model.FindXAgentConfigRollout{
		TemplateID: "host-plugin",
		TargetIDs:  []string{"target-running"},
		Status:     "running",
		Metadata:   map[string]string{"evidence_chain_ref": "running-evidence-ref"},
	})
	if err != nil {
		t.Fatalf("save non-blocked rollout: %v", err)
	}

	w := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id="+nonBlocked.ID, ListFindXAgentConfigRollouts)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"status":"blocked"`) || !strings.Contains(w.Body.String(), "executor not enabled") {
		t.Fatalf("coerced rollout detail should return blocked gate, got %d body=%s", w.Code, w.Body.String())
	}
	assertNoConfigRolloutExecutionStates(t, w.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, w.Body.String())

	missing := performAgentLifecycleGet("/api/v1/findx-agents/config-rollouts?id=missing-rollout", ListFindXAgentConfigRollouts)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing rollout detail should return 404, got %d body=%s", missing.Code, missing.Body.String())
	}
	assertNoConfigRolloutExecutionStates(t, missing.Body.String())
	assertNoConfigRolloutSensitiveEcho(t, missing.Body.String())
}

func TestFindXAgentTaskPersistsBlockedRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"action":"uninstall",
		"target_ids":["target-a"],
		"package_id":"agent-core",
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{"reason":"cleanup"}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
	if w.Code != http.StatusConflict {
		t.Fatalf("agent task should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	var payload struct {
		Data model.FindXAgentExecutionTask `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if payload.Data.ID == "" || payload.Data.Action != "uninstall" || payload.Data.Status != "blocked" || !payload.Data.CredentialRefPresent {
		t.Fatalf("blocked task should be persisted: %#v", payload.Data)
	}
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentLifecycleListsReturnEmptySlices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	for _, tt := range []struct {
		path    string
		handler gin.HandlerFunc
	}{
		{path: "/api/v1/findx-agents/install-plans", handler: ListFindXAgentInstallPlans},
		{path: "/api/v1/findx-agents/config-rollouts", handler: ListFindXAgentConfigRollouts},
		{path: "/api/v1/findx-agents/tasks", handler: ListFindXAgentTasks},
		{path: "/api/v1/findx-agents/data-arrival-evidence", handler: ListFindXAgentDataArrivalEvidence},
	} {
		w := performAgentLifecycleGet(tt.path, tt.handler)
		if w.Code != http.StatusOK {
			t.Fatalf("%s should return 200, got %d body=%s", tt.path, w.Code, w.Body.String())
		}
		if strings.TrimSpace(w.Body.String()) != "[]" {
			t.Fatalf("%s should return empty list, got %s", tt.path, w.Body.String())
		}
	}
}

func TestFindXAgentLifecyclePostsAreReadableFromLists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	for _, tt := range lifecyclePostReadCases() {
		w := performAgentLifecyclePost(tt.path, strings.NewReader(tt.body), tt.handler)
		if w.Code != http.StatusConflict {
			t.Fatalf("%s should stay blocked, got %d body=%s", tt.path, w.Code, w.Body.String())
		}
		assertNoCredentialEcho(t, w.Body.String())
	}

	for _, tt := range lifecycleListReadCases() {
		assertLifecycleListReadsBlockedRecord(t, tt)
	}
}

type lifecyclePostReadCase struct {
	path    string
	body    string
	handler gin.HandlerFunc
}

func lifecyclePostReadCases() []lifecyclePostReadCase {
	return []lifecyclePostReadCase{
		{
			path:    "/api/v1/findx-agents/install-plans",
			body:    `{"package_id":"agent-core","os":"linux","method":"linux-curl","target_ids":["target-a"],"credential_ref":"<CREDENTIAL_REF>","metadata":{"ticket":"CHG-1","token":"secret"}}`,
			handler: CreateFindXAgentInstallPlan,
		},
		{
			path:    "/api/v1/findx-agents/config-rollouts",
			body:    `{"template_id":"host-plugin","target_ids":["target-a"],"plugin_id":"input.cpu","config_snippet_ref":"<CONFIG_SNIPPET_REF>","credential_ref":"<CREDENTIAL_REF>","metadata":{"change":"CHG-1","cookie":"secret"}}`,
			handler: CreateFindXAgentConfigRollout,
		},
		{
			path:    "/api/v1/findx-agents/tasks",
			body:    `{"action":"uninstall","target_ids":["target-a"],"package_id":"agent-core","credential_ref":"<CREDENTIAL_REF>","metadata":{"reason":"cleanup","session":"secret"}}`,
			handler: CreateFindXAgentTask,
		},
	}
}

type lifecycleListReadCase struct {
	path       string
	handler    gin.HandlerFunc
	wantStatus string
}

func lifecycleListReadCases() []lifecycleListReadCase {
	return []lifecycleListReadCase{
		{path: "/api/v1/findx-agents/install-plans", handler: ListFindXAgentInstallPlans, wantStatus: `"status":"blocked"`},
		{path: "/api/v1/findx-agents/config-rollouts", handler: ListFindXAgentConfigRollouts, wantStatus: `"status":"blocked"`},
		{path: "/api/v1/findx-agents/tasks", handler: ListFindXAgentTasks, wantStatus: `"status":"blocked"`},
	}
}

func assertLifecycleListReadsBlockedRecord(t *testing.T, tt lifecycleListReadCase) {
	t.Helper()
	w := performAgentLifecycleGet(tt.path, tt.handler)
	if w.Code != http.StatusOK {
		t.Fatalf("%s should return 200, got %d body=%s", tt.path, w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, tt.wantStatus) || !strings.Contains(body, "PENDING") {
		t.Fatalf("%s should read back blocked record, body=%s", tt.path, body)
	}
	assertNoCredentialEcho(t, body)
}

func TestFindXAgentDataArrivalEvidenceListDoesNotFakeSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)

	w := performAgentLifecycleGet("/api/v1/findx-agents/data-arrival-evidence", ListFindXAgentDataArrivalEvidence)
	if w.Code != http.StatusOK {
		t.Fatalf("data arrival evidence should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(body, `"status":"reported"`) || strings.Contains(body, `"status":"success"`) {
		t.Fatalf("data arrival evidence list must not fake success: %s", body)
	}
	assertNoCredentialEcho(t, body)
}

func performAgentLifecyclePost(path string, body *strings.Reader, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler(c)
	return w
}

func performAgentLifecycleGet(path string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler(c)
	return w
}

func decodeAgentTaskResponse(t *testing.T, w *httptest.ResponseRecorder) struct {
	Data model.FindXAgentExecutionTask `json:"data"`
} {
	t.Helper()
	var payload struct {
		Data model.FindXAgentExecutionTask `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid task response: %v body=%s", err, w.Body.String())
	}
	return payload
}

func resetAgentLifecycleRecordsForTest(t *testing.T) {
	t.Helper()
	store.ResetFindXAgentLifecycleForTest()
	store.ResetFindXAgentInstallExecutionsForTest()
}

func assertNoCredentialEcho(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{`"credential_ref":`, "<CREDENTIAL_REF>", "secret"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response must not echo credential or sensitive metadata: %s", body)
		}
	}
}

func assertNoInstallPlanSuccessStates(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{
		`"status":"queued"`,
		`"status":"running"`,
		`"status":"succeeded"`,
		`"status":"success"`,
		`"status":"installed"`,
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("install plan response must not expose execution success states: %s", body)
		}
	}
}

func assertNoInstallPlanSensitiveEcho(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{
		`"credential_ref":`,
		"<CREDENTIAL_REF>",
		"secret",
		`"token":`,
		`"cookie":`,
		`"session":`,
		`"dsn":`,
		`"password":`,
		`"private_key":`,
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("install plan response must not echo credential or sensitive metadata: %s", body)
		}
	}
}

func assertNoConfigRolloutExecutionStates(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{
		`"status":"queued"`,
		`"status":"running"`,
		`"status":"succeeded"`,
		`"status":"success"`,
		`"status":"applied"`,
		`"status":"rolled-back"`,
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("config rollout response must not expose execution success states: %s", body)
		}
	}
}
