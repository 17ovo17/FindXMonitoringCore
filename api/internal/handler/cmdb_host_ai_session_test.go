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

func TestCmdbHostAISessionPreflightBlocksWithoutRuntimeAndPreservesContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	target := createCmdbHostOpsMonitorTargetFixture(t)
	router := gin.New()
	router.GET("/hosts/:id/ai-session", GetCmdbHostAISessionPreflight)

	req := httptest.NewRequest(http.MethodGet, "/hosts/"+target.ID+"/ai-session", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := assertCmdbBlockedResponse(t, w, cmdbAIHostSessionRuntimeContract)
	assertCmdbAIHostSessionMissing(t, body)
	assertCmdbMissingContract(t, body, cmdbHostInstanceMappingContract)
	if _, ok := body["session"]; ok {
		t.Fatalf("blocked host AI session must not expose fake session: %s", w.Body.String())
	}
	if _, ok := body["messages"]; ok {
		t.Fatalf("blocked host AI session must not expose fake messages: %s", w.Body.String())
	}
	context := body["host_context"].(map[string]any)
	if _, ok := context["monitor_target"].(map[string]any); !ok {
		t.Fatalf("blocked host AI session should include monitor target context: %#v", context)
	}
	if _, ok := context["cmdb_instance"]; ok {
		t.Fatalf("unmapped monitor target must not pretend CMDB instance context exists: %#v", context)
	}
	text := strings.ToLower(w.Body.String())
	for _, forbidden := range []string{"password", "token", "cookie", "queued", "running", "installed", "succeeded", `"success":true`} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("blocked host AI session leaked fake state or sensitive word %q: %s", forbidden, w.Body.String())
		}
	}
}

func TestCmdbHostAISessionCreateDoesNotCreateChatSessionOrEchoSensitiveMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hostID := createCmdbHostOpsFixture(t)
	before := len(store.ListChatSessions())
	router := gin.New()
	router.POST("/hosts/:id/ai-session", CreateCmdbHostAISessionPreflight)

	req := httptest.NewRequest(http.MethodPost, "/hosts/"+hostID+"/ai-session", bytes.NewReader([]byte(`{"message":"check password=<SECRET> token=<TOKEN>","tool":"shell","session_id":"session-sensitive-ref-1234567890","metadata":{"password":"raw"}}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := assertCmdbBlockedResponse(t, w, cmdbAIHostSessionRuntimeContract)
	assertCmdbAIHostSessionMissing(t, body)
	if got := len(store.ListChatSessions()); got != before {
		t.Fatalf("blocked host AI session must not create chat session, before=%d after=%d", before, got)
	}
	preview := body["request_preview"].(map[string]any)
	if preview["message_length"].(float64) == 0 || preview["tool_requested"] != true {
		t.Fatalf("blocked host AI session should expose safe request preview: %#v", preview)
	}
	text := strings.ToLower(w.Body.String())
	for _, forbidden := range []string{"password=<secret>", "token=<token>", "raw", "queued", "running", "installed", "succeeded", `"success":true`} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("blocked host AI session leaked fake state or sensitive marker %q: %s", forbidden, w.Body.String())
		}
	}
}

func TestCmdbHostAISessionContextIncludesMappedCmdbAndAgentWithoutReadySession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	agent, target, err := store.UpsertFindXAgentHeartbeat(model.FindXAgentHeartbeat{
		Ident:    "cmdb-ai-host-agent",
		IP:       "10.128.15.30",
		Hostname: "cmdb-ai-host-agent",
		Version:  "1.2.8",
	})
	if err != nil {
		t.Fatalf("seed heartbeat: %v", err)
	}
	t.Cleanup(func() {
		if agent != nil {
			_, _ = store.DeleteFindXAgent(agent.ID)
			_, _ = store.DeleteFindXAgent(agent.Ident)
		}
		_, _ = store.DeleteMonitorTarget(target.ID)
	})
	inst := &model.CmdbInstance{
		ObjectID: "obj-os",
		Data:     `{"name":"cmdb-ai-host-agent","ip_address":"10.128.15.30","password":"must-not-leak"}`,
		Creator:  "unit-test",
		Updater:  "unit-test",
	}
	if err := store.CreateCmdbInstance(inst); err != nil {
		t.Fatalf("create cmdb instance: %v", err)
	}

	router := gin.New()
	router.GET("/hosts/:id/ai-session", GetCmdbHostAISessionPreflight)
	req := httptest.NewRequest(http.MethodGet, "/hosts/"+target.ID+"/ai-session", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := assertCmdbBlockedResponse(t, w, cmdbAIHostSessionRuntimeContract)
	assertCmdbAIHostSessionMissing(t, body)
	context := body["host_context"].(map[string]any)
	if _, ok := context["cmdb_instance"].(map[string]any); !ok {
		t.Fatalf("mapped host AI session should include CMDB instance context: %#v", context)
	}
	if _, ok := context["agent"].(map[string]any); !ok {
		t.Fatalf("mapped host AI session should include Agent context: %#v", context)
	}
	if strings.Contains(strings.ToLower(w.Body.String()), "must-not-leak") {
		t.Fatalf("host AI context leaked CMDB sensitive data: %s", w.Body.String())
	}
}

func assertCmdbAIHostSessionMissing(t *testing.T, body map[string]any) {
	t.Helper()
	for _, want := range append([]string{
		"cmdb_resource_approval_runtime_contract",
		"cmdb_operation_risk_policy_contract",
		"cmdb_action_preflight_contract",
		"cmdb_action_audit_receipt_contract",
	}, cmdbAIHostSessionMissingContracts...) {
		assertCmdbMissingContract(t, body, want)
	}
	if query, ok := body["findx_audit_query"].(map[string]any); !ok || query["resource_type"] != "cmdb_host_ai_session" {
		raw, _ := json.Marshal(body["findx_audit_query"])
		t.Fatalf("host AI blocked response missing findx_audit query: %s", raw)
	}
}
