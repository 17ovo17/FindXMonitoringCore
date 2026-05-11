package handler

import (
	"net/http"
	"strings"
	"testing"
)

func assertTaskDetailBlockedOnly(t *testing.T, id string) string {
	t.Helper()
	w := performAgentLifecycleGet("/api/v1/findx-agents/tasks?id="+id, ListFindXAgentTasks)
	if w.Code != http.StatusOK {
		t.Fatalf("task detail should return 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	for _, want := range []string{id, `"action":"rollback"`, `"status":"blocked"`, "executor not enabled", `"ticket":"CHG-1"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("task detail should include %s, body=%s", want, body)
		}
	}
	return body
}

func assertTaskResponsesStayBlockedOnly(t *testing.T, bodies ...string) {
	t.Helper()
	for _, body := range bodies {
		assertTaskResponseHasNoExecutionState(t, body)
		assertNoCredentialEcho(t, body)
	}
}

func assertTaskResponseHasNoExecutionState(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range taskForbiddenExecutionStates() {
		if strings.Contains(body, forbidden) {
			t.Fatalf("task response must not expose fake execution state: %s", body)
		}
	}
}

func taskForbiddenExecutionStates() []string {
	return []string{
		`"status":"queued"`, `"status":"running"`, `"status":"succeeded"`, `"status":"success"`,
		`"status":"applied"`, `"status":"rolled-back"`, `"status":"failed"`, `"status":"cancelled"`,
		`"status":"rollback_required"`, `"status":"uninstall_verified"`,
	}
}
