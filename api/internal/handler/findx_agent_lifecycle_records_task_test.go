package handler

import (
	"net/http"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestFindXAgentTaskRequiresTargetOrAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"action":"uninstall","metadata":{"uninstall_manifest_ref":"manifest-1","executor_ref":"executor-1"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("agent task without target or agent should be 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "target_ids or agent_ids is required") {
		t.Fatalf("empty target error should be explicit, body=%s", w.Body.String())
	}
}

func TestFindXAgentTaskActionSpecificMissingRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range taskActionMissingRefCases() {
		t.Run(tt.name, func(t *testing.T) {
			assertTaskActionMissingRefs(t, tt)
		})
	}
}

type taskActionMissingRefCase struct {
	name        string
	action      string
	metadata    string
	missingRefs []string
}

func taskActionMissingRefCases() []taskActionMissingRefCase {
	return []taskActionMissingRefCase{
		{name: "uninstall", action: "uninstall", metadata: `"executor_ref":"executor-1"`, missingRefs: []string{"uninstall_manifest_ref"}},
		{name: "rollback", action: "rollback", metadata: `"rollback_manifest_ref":"rollback-1","executor_ref":"executor-1"`, missingRefs: []string{"state_snapshot_ref"}},
		{name: "upgrade", action: "upgrade", metadata: `"package_repository_ref":"repo-1","signature_ref":"sig-1","checksum":"sha256:abc","executor_ref":"executor-1"`, missingRefs: []string{"script_manifest_ref"}},
		{name: "restart", action: "restart", metadata: `"executor_ref":"executor-1"`, missingRefs: []string{"service_ref"}},
	}
}

func assertTaskActionMissingRefs(t *testing.T, tt taskActionMissingRefCase) {
	t.Helper()
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"action":"` + tt.action + `","target_ids":["target-a"],"credential_ref":"<CREDENTIAL_REF>","metadata":{` + tt.metadata + `}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
	payload := decodeAgentTaskResponse(t, w)
	if w.Code != http.StatusConflict {
		t.Fatalf("%s missing refs should stay blocked 409, got %d body=%s", tt.action, w.Code, w.Body.String())
	}
	assertTaskMissingRefBlocker(t, tt, payload.Data)
	assertNoCredentialEcho(t, w.Body.String())
}

func assertTaskMissingRefBlocker(t *testing.T, tt taskActionMissingRefCase, task model.FindXAgentExecutionTask) {
	t.Helper()
	if task.Status != "blocked" || task.Blocker == "" {
		t.Fatalf("%s should persist blocked task, got %#v", tt.action, task)
	}
	if !strings.Contains(task.Blocker, "BLOCKED_BY_CONTRACT") {
		t.Fatalf("%s blocker should include contract marker: %s", tt.action, task.Blocker)
	}
	for _, ref := range tt.missingRefs {
		if !strings.Contains(task.Blocker, ref) {
			t.Fatalf("%s blocker should name missing ref %s: %s", tt.action, ref, task.Blocker)
		}
	}
}

func TestFindXAgentTaskRemotePreflightMissingRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{"action":"restart","target_ids":["target-a"],"metadata":{"service_ref":"svc-1","executor_ref":"executor-1"}}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
	payload := decodeAgentTaskResponse(t, w)
	if w.Code != http.StatusConflict {
		t.Fatalf("remote preflight gaps should stay blocked 409, got %d body=%s", w.Code, w.Body.String())
	}
	want := "BLOCKED_BY_CONTRACT: missing audit_ref_or_evidence_chain_ref, credential_ref, data_arrival_validator_ref, execution_receipt_ref_or_receipt_ref, idempotency_key, target_os, timeout_policy_ref, transport_or_runner"
	if payload.Data.Blocker != want {
		t.Fatalf("missing refs should be stable sorted, want %q got %q", want, payload.Data.Blocker)
	}
	if payload.Data.CredentialRefPresent {
		t.Fatalf("missing top-level credential_ref should persist credential_ref_present=false: %#v", payload.Data)
	}
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentTaskCompleteRefsStillBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name     string
		action   string
		metadata string
	}{
		{
			name:     "uninstall",
			action:   "uninstall",
			metadata: `"uninstall_manifest_ref":"manifest-1","executor_ref":"executor-1"`,
		},
		{
			name:     "rollback",
			action:   "rollback",
			metadata: `"rollback_manifest_ref":"rollback-1","state_snapshot_ref":"snapshot-1","executor_ref":"executor-1"`,
		},
		{
			name:     "upgrade",
			action:   "upgrade",
			metadata: `"package_repository_ref":"repo-1","signature_ref":"sig-1","checksum":"sha256:abc","script_manifest_ref":"script-1","executor_ref":"executor-1"`,
		},
		{
			name:     "restart",
			action:   "restart",
			metadata: `"service_ref":"svc-1","executor_ref":"executor-1"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetAgentLifecycleRecordsForTest(t)
			body := strings.NewReader(`{"action":"` + tt.action + `","agent_ids":["agent-a"],"credential_ref":"<CREDENTIAL_REF>","metadata":{` + tt.metadata + `,` + completeRemotePreflightMetadata() + `}}`)
			w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
			payload := decodeAgentTaskResponse(t, w)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s complete refs should stay blocked 409, got %d body=%s", tt.action, w.Code, w.Body.String())
			}
			if payload.Data.Status != "blocked" || payload.Data.Blocker != "BLOCKED_BY_CONTRACT: executor not enabled / execution protocol not open" {
				t.Fatalf("%s complete refs should still honest-block, got %#v", tt.action, payload.Data)
			}
			for _, forbidden := range []string{"queued", "running", "succeeded"} {
				if strings.Contains(w.Body.String(), forbidden) {
					t.Fatalf("%s must not fake execution state %s: %s", tt.action, forbidden, w.Body.String())
				}
			}
			assertNoCredentialEcho(t, w.Body.String())
		})
	}
}

func TestFindXAgentTaskKubernetesMissingRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"action":"restart",
		"target_ids":["target-a"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{
			"target_os":"kubernetes",
			"service_ref":"svc-1",
			"executor_ref":"executor-1",
			"transport":"k8s-api",
			"idempotency_key":"idem-1",
			"timeout_policy_ref":"timeout-1",
			"execution_receipt_ref":"receipt-1",
			"audit_ref":"audit-1"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
	payload := decodeAgentTaskResponse(t, w)
	if w.Code != http.StatusConflict {
		t.Fatalf("kubernetes missing refs should stay blocked 409, got %d body=%s", w.Code, w.Body.String())
	}
	want := "BLOCKED_BY_CONTRACT: missing cluster_ref, data_arrival_validator_ref, namespace_ref, rbac_ref, restart_strategy_ref, rollout_receipt_ref, rollout_strategy_ref, service_account_ref, workload_selector_ref"
	if payload.Data.Blocker != want {
		t.Fatalf("kubernetes missing refs should be stable sorted, want %q got %q", want, payload.Data.Blocker)
	}
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentTaskKubernetesCompleteRefsStillExecutorBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"action":"restart",
		"agent_ids":["agent-a"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{
			"target_os":"k8s",
			"workload_kind":"DaemonSet",
			"service_ref":"svc-1",
			"executor_ref":"executor-1",
			"restart_strategy_ref":"restart-strategy-1",
			` + completeRemotePreflightMetadata() + `,
			` + completeKubernetesTaskMetadata() + `
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
	payload := decodeAgentTaskResponse(t, w)
	if w.Code != http.StatusConflict {
		t.Fatalf("complete kubernetes refs should stay blocked 409, got %d body=%s", w.Code, w.Body.String())
	}
	if payload.Data.Blocker != "BLOCKED_BY_CONTRACT: executor not enabled / execution protocol not open" {
		t.Fatalf("complete kubernetes refs should still executor-block, got %#v", payload.Data)
	}
	for _, forbidden := range []string{"queued", "running", "succeeded"} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("kubernetes task must not fake execution state %s: %s", forbidden, w.Body.String())
		}
	}
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentTaskHelmUpgradeMissingChoiceRef(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"action":"upgrade",
		"target_ids":["target-a"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{
			"method":"helm",
			"package_repository_ref":"repo-1",
			"signature_ref":"sig-1",
			"checksum":"sha256:abc",
			"script_manifest_ref":"script-1",
			"executor_ref":"executor-1",
			"helm_release_ref":"release-1",
			"values_ref":"values-1",
			"image_ref":"image-1",
			"config_map_ref":"config-map-1",
			` + completeRemotePreflightMetadata() + `,
			` + completeKubernetesTaskMetadata() + `
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
	payload := decodeAgentTaskResponse(t, w)
	if w.Code != http.StatusConflict {
		t.Fatalf("helm upgrade without chart or manifest should stay blocked 409, got %d body=%s", w.Code, w.Body.String())
	}
	want := "BLOCKED_BY_CONTRACT: missing helm_chart_ref_or_manifest_bundle_ref"
	if payload.Data.Blocker != want {
		t.Fatalf("helm upgrade choice ref should be explicit, want %q got %q", want, payload.Data.Blocker)
	}
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentTaskKubernetesSensitiveMetadataNotEchoed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"action":"uninstall",
		"target_ids":["target-a"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{
			"orchestrator":"kubernetes",
			"uninstall_manifest_ref":"manifest-1",
			"executor_ref":"executor-1",
			"teardown_plan_ref":"teardown-1",
			` + completeRemotePreflightMetadata() + `,
			` + completeKubernetesTaskMetadata() + `,
			"api_token":"token-value",
			"password":"password-value",
			"cookie":"cookie-value",
			"session":"session-value",
			"ticket":"CHG-1"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
	payload := decodeAgentTaskResponse(t, w)
	if w.Code != http.StatusConflict {
		t.Fatalf("kubernetes sensitive metadata task should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	if payload.Data.Blocker != "BLOCKED_BY_CONTRACT: executor not enabled / execution protocol not open" {
		t.Fatalf("complete safe kubernetes metadata should not be blocked by dropped sensitive refs, got %#v", payload.Data)
	}
	if payload.Data.Metadata["ticket"] != "CHG-1" {
		t.Fatalf("safe metadata should be retained, got %#v", payload.Data.Metadata)
	}
	for _, key := range []string{"api_token", "password", "cookie", "session"} {
		if payload.Data.Metadata[key] != "" {
			t.Fatalf("sensitive kubernetes metadata key %s should be dropped, got %#v", key, payload.Data.Metadata)
		}
	}
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentTaskMetadataDropsSensitiveRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	body := strings.NewReader(`{
		"action":"restart",
		"target_ids":["target-a"],
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{
			"service_ref":"svc-1",
			"executor_ref":"executor-1",
			` + completeRemotePreflightMetadata() + `,
			"credential_ref":"cred-1",
			"api_token":"token-value",
			"secret":"secret-value",
			"password":"password-value",
			"cookie":"cookie-value",
			"session":"session-value",
			"dsn":"mysql://user:pass@db/prod",
			"private_key":"-----BEGIN PRIVATE KEY-----",
			"ticket":"CHG-1"
		}
	}`)
	w := performAgentLifecyclePost("/api/v1/findx-agents/tasks", body, CreateFindXAgentTask)
	payload := decodeAgentTaskResponse(t, w)
	if w.Code != http.StatusConflict {
		t.Fatalf("sensitive metadata task should stay blocked, got %d body=%s", w.Code, w.Body.String())
	}
	if payload.Data.Metadata["ticket"] != "CHG-1" || payload.Data.Metadata["service_ref"] != "svc-1" {
		t.Fatalf("safe metadata should be retained, got %#v", payload.Data.Metadata)
	}
	for _, key := range []string{"credential_ref", "api_token", "secret", "password", "cookie", "session", "dsn", "private_key"} {
		if payload.Data.Metadata[key] != "" {
			t.Fatalf("sensitive metadata key %s should be dropped, got %#v", key, payload.Data.Metadata)
		}
	}
	assertNoCredentialEcho(t, w.Body.String())
}

func TestFindXAgentTasksListAndDetailAreBlockedOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetAgentLifecycleRecordsForTest(t)
	payload := createBlockedRollbackTaskForListTest(t)
	nonBlocked := saveNonBlockedTaskForCoercionTest(t)
	listBody := assertTaskListIncludesBlockedRecords(t, payload.Data.ID, nonBlocked.ID)
	detailBody := assertTaskDetailBlockedOnly(t, payload.Data.ID)
	assertTaskResponsesStayBlockedOnly(t, listBody, detailBody)
	assertCoercedTaskDetailBlocked(t, nonBlocked.ID)
	assertMissingTaskDetailNotFound(t)
}

func createBlockedRollbackTaskForListTest(t *testing.T) struct {
	Data model.FindXAgentExecutionTask `json:"data"`
} {
	t.Helper()
	post := performAgentLifecyclePost("/api/v1/findx-agents/tasks", strings.NewReader(`{"action":"rollback","agent_ids":["agent-a"],"credential_ref":"<CREDENTIAL_REF>","metadata":{"rollback_manifest_ref":"rollback-1","state_snapshot_ref":"snapshot-1","executor_ref":"executor-1",`+completeRemotePreflightMetadata()+`,"ticket":"CHG-1","token":"secret"}}`), CreateFindXAgentTask)
	payload := decodeAgentTaskResponse(t, post)
	if post.Code != http.StatusConflict || payload.Data.ID == "" {
		t.Fatalf("blocked task should be created before detail query, code=%d body=%s", post.Code, post.Body.String())
	}
	return payload
}

func saveNonBlockedTaskForCoercionTest(t *testing.T) model.FindXAgentExecutionTask {
	t.Helper()
	nonBlocked, err := store.SaveFindXAgentExecutionTask(model.FindXAgentExecutionTask{Action: "upgrade", TargetIDs: []string{"target-queued"}, Status: "queued", Audit: "findx_agent.task.requested"})
	if err != nil {
		t.Fatalf("save non-blocked task: %v", err)
	}
	if nonBlocked.Status != "blocked" || nonBlocked.Blocker != "BLOCKED_BY_CONTRACT: executor not enabled / execution protocol not open" {
		t.Fatalf("store must coerce execution task to blocked, got %#v", nonBlocked)
	}
	return nonBlocked
}

func assertTaskListIncludesBlockedRecords(t *testing.T, ids ...string) string {
	t.Helper()
	list := performAgentLifecycleGet("/api/v1/findx-agents/tasks", ListFindXAgentTasks)
	if list.Code != http.StatusOK {
		t.Fatalf("task list should return 200, got %d body=%s", list.Code, list.Body.String())
	}
	body := list.Body.String()
	for _, id := range ids {
		if !strings.Contains(body, id) {
			t.Fatalf("task list should include blocked record %s, body=%s", id, body)
		}
	}
	assertNoCredentialEcho(t, body)
	return body
}

func assertCoercedTaskDetailBlocked(t *testing.T, id string) {
	t.Helper()
	coerced := performAgentLifecycleGet("/api/v1/findx-agents/tasks?id="+id, ListFindXAgentTasks)
	if coerced.Code != http.StatusOK {
		t.Fatalf("coerced task detail should return 200, got %d body=%s", coerced.Code, coerced.Body.String())
	}
	if !strings.Contains(coerced.Body.String(), `"status":"blocked"`) || !strings.Contains(coerced.Body.String(), "executor not enabled") {
		t.Fatalf("coerced task detail should expose blocked gate, body=%s", coerced.Body.String())
	}
	assertNoCredentialEcho(t, coerced.Body.String())
}

func assertMissingTaskDetailNotFound(t *testing.T) {
	t.Helper()
	missing := performAgentLifecycleGet("/api/v1/findx-agents/tasks?id=missing-task", ListFindXAgentTasks)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing task detail should return 404, got %d body=%s", missing.Code, missing.Body.String())
	}
	assertNoCredentialEcho(t, missing.Body.String())
}
