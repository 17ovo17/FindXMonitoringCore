package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestCategrafHTTPProviderRejectsMissingToken(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-provider-token", false)
	w := performCategrafProviderRequest("/categraf/configs?agent=categraf&host=machine1", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("provider request without token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafHTTPProviderRejectsMissingTarget(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-provider-token", false)
	w := performCategrafProviderRequest("/categraf/configs?version=v1", "unit-provider-token")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("provider request without target should be 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafHTTPProviderReturnsEmptyConfigsForUnknownAgent(t *testing.T) {
	resetCategrafReceiverTestState(t)
	store.ResetCategrafConfigAssignmentsForTest()
	configureCategrafReceiverTokenTest(t, "unit-provider-token", false)
	path := "/categraf/configs?agent=categraf&host=machine1&version=v1&agent_hostname=node-a"
	w := performCategrafProviderRequest(path, "unit-provider-token")
	if w.Code != http.StatusOK {
		t.Fatalf("provider request should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var payload store.CategrafProviderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("provider response should be json: %v", err)
	}
	if payload.Version != "" {
		t.Fatalf("provider version should be empty for unknown agent, got %q", payload.Version)
	}
	if payload.Configs == nil || len(payload.Configs) != 0 {
		t.Fatalf("provider configs must be empty for unknown agent, got %#v", payload.Configs)
	}
}

func TestCategrafHTTPProviderReturnsAssignedConfigs(t *testing.T) {
	resetCategrafReceiverTestState(t)
	store.ResetCategrafConfigAssignmentsForTest()
	configureCategrafReceiverTokenTest(t, "unit-provider-token", false)

	// 分配配置给 agent
	_, err := store.UpsertCategrafConfigByAgent("node-a", "mysql", "[[instances]]\naddress = \"127.0.0.1:3306\"", "toml")
	if err != nil {
		t.Fatalf("upsert config: %v", err)
	}
	_, err = store.UpsertCategrafConfigByAgent("node-a", "redis", "[[instances]]\naddress = \"127.0.0.1:6379\"", "toml")
	if err != nil {
		t.Fatalf("upsert config: %v", err)
	}

	path := "/categraf/configs?agent_hostname=node-a&version=old-version"
	w := performCategrafProviderRequest(path, "unit-provider-token")
	if w.Code != http.StatusOK {
		t.Fatalf("provider request should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var payload store.CategrafProviderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("provider response should be json: %v", err)
	}
	if payload.Version == "" {
		t.Fatalf("provider version should not be empty when configs exist")
	}
	if len(payload.Configs) != 2 {
		t.Fatalf("expected 2 input configs, got %d: %#v", len(payload.Configs), payload.Configs)
	}
	if _, ok := payload.Configs["mysql"]; !ok {
		t.Fatalf("expected mysql config in response, got %#v", payload.Configs)
	}
	if _, ok := payload.Configs["redis"]; !ok {
		t.Fatalf("expected redis config in response, got %#v", payload.Configs)
	}
}

func TestCategrafHTTPProviderReturnsEmptyWhenVersionMatches(t *testing.T) {
	resetCategrafReceiverTestState(t)
	store.ResetCategrafConfigAssignmentsForTest()
	configureCategrafReceiverTokenTest(t, "unit-provider-token", false)

	_, err := store.UpsertCategrafConfigByAgent("node-b", "cpu", "[interval]\ncollect = 10", "toml")
	if err != nil {
		t.Fatalf("upsert config: %v", err)
	}

	// 先获取当前 version
	resp := store.BuildCategrafProviderResponse("node-b")
	currentVersion := resp.Version

	// 用相同 version 请求
	path := "/categraf/configs?agent_hostname=node-b&version=" + currentVersion
	w := performCategrafProviderRequest(path, "unit-provider-token")
	if w.Code != http.StatusOK {
		t.Fatalf("provider request should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var payload store.CategrafProviderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("provider response should be json: %v", err)
	}
	if payload.Version != currentVersion {
		t.Fatalf("version should match, got %q want %q", payload.Version, currentVersion)
	}
	if len(payload.Configs) != 0 {
		t.Fatalf("configs should be empty when version matches, got %#v", payload.Configs)
	}
}

func TestCategrafHTTPProviderAcceptsSharedToken(t *testing.T) {
	resetCategrafReceiverTestState(t)
	store.ResetCategrafConfigAssignmentsForTest()
	configureCategrafProviderSharedTokenTest(t, "unit-provider-shared-token", false)
	path := "/categraf/configs?agent=categraf&host=machine1&version=v1"
	w := performCategrafProviderRequest(path, "unit-provider-shared-token")
	if w.Code != http.StatusOK {
		t.Fatalf("provider request with shared token should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var payload store.CategrafProviderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("provider response should be json: %v", err)
	}
	if payload.Configs == nil {
		t.Fatalf("provider response should include configs map (even if empty)")
	}
}

func TestCategrafHTTPProviderRejectsAnonymousDebugToken(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafProviderSharedTokenTest(t, "", true)
	w := performCategrafProviderRequest("/categraf/configs?agent=categraf&host=machine1", "any-debug-token")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("provider request with allow_anonymous and no configured token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategrafHTTPProviderRejectsWrongTokenWithoutEcho(t *testing.T) {
	resetCategrafReceiverTestState(t)
	configureCategrafReceiverTokenTest(t, "unit-provider-token", false)
	w := performCategrafProviderRequest("/categraf/configs?agent=categraf&host=machine1", "wrong-provider-token")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("provider request with wrong token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
	for _, forbidden := range []string{"wrong-provider-token", "unit-provider-token", "basic-user", "basic-pass", "synthetic-cookie"} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("provider 401 response must not echo sensitive value %q: %s", forbidden, w.Body.String())
		}
	}
}

func TestCategrafHTTPProviderResponseIsSanitized(t *testing.T) {
	resetCategrafReceiverTestState(t)
	store.ResetCategrafConfigAssignmentsForTest()
	configureCategrafReceiverTokenTest(t, "unit-provider-token", false)
	path := "/categraf/configs?host=machine1&password=synthetic-password&cookie=synthetic-cookie&dsn=synthetic-dsn"
	w := performCategrafProviderRequest(path, "unit-provider-token")
	if w.Code != http.StatusOK {
		t.Fatalf("provider request should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	for _, forbidden := range []string{"unit-provider-token", "synthetic-password", "synthetic-cookie", "synthetic-dsn", "basic-user", "basic-pass"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("provider response must not echo sensitive value %q: %s", forbidden, body)
		}
	}
}

func TestCategrafHTTPProviderRootRouteIsRegistered(t *testing.T) {
	router := gin.New()
	router.GET("/categraf/configs", CategrafHTTPProviderConfigs)
	routes := router.Routes()
	if len(routes) != 1 || routes[0].Method != http.MethodGet || routes[0].Path != "/categraf/configs" {
		t.Fatalf("expected GET /categraf/configs route, got %#v", routes)
	}
}

func performCategrafProviderRequest(path, token string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = "127.0.0.1:51000"
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("basic-user:basic-pass")))
	req.Header.Set("Cookie", "session=synthetic-cookie")
	if token != "" {
		req.Header.Set("X-Agent-Token", token)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	CategrafHTTPProviderConfigs(c)
	return w
}

func configureCategrafProviderSharedTokenTest(t *testing.T, token string, allowAnonymous bool) {
	t.Helper()
	t.Setenv("FINDX_AGENT_TOKEN", "")
	viper.Set("findx_agents.shared_token", token)
	viper.Set("findx_agents.allow_anonymous", allowAnonymous)
	t.Cleanup(func() {
		viper.Set("findx_agents.shared_token", "")
		viper.Set("findx_agents.allow_anonymous", false)
	})
}
