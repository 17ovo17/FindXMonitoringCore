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

func TestFindXAgentPluginCatalogCredentialPluginsExposeClosedContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := performAgentPluginCatalogRequest()
	if w.Code != http.StatusOK {
		t.Fatalf("plugin catalog should be 200, got %d body=%s", w.Code, w.Body.String())
	}

	var plugins []model.FindXAgentPlugin
	if err := json.Unmarshal(w.Body.Bytes(), &plugins); err != nil {
		t.Fatalf("decode plugin catalog: %v body=%s", err, w.Body.String())
	}
	byID := make(map[string]model.FindXAgentPlugin, len(plugins))
	for _, plugin := range plugins {
		byID[plugin.ID] = plugin
	}

	for _, pluginID := range []string{
		"redis",
		"mysql",
		"postgresql",
		"mongodb",
		"oracle",
		"kafka",
		"rabbitmq",
		"clickhouse",
		"elasticsearch",
		"nginx",
		"etcd",
	} {
		t.Run(pluginID, func(t *testing.T) {
			plugin, ok := byID[pluginID]
			if !ok {
				t.Fatalf("credential plugin %s should exist in catalog", pluginID)
			}
			if !plugin.CredentialRequired {
				t.Fatalf("%s should require credential_ref", pluginID)
			}
			wantCredentialSchema := map[string]string{
				"contract": "cmdb.agent.plugin.credential.v1",
				"mode":     "credential_ref",
				"fields":   "credential_ref",
			}
			for key, want := range wantCredentialSchema {
				if plugin.CredentialSchema[key] != want {
					t.Fatalf("%s credential_schema[%s] got %q want %q schema=%v", pluginID, key, plugin.CredentialSchema[key], want, plugin.CredentialSchema)
				}
			}
			if len(plugin.DashboardRefs) == 0 {
				t.Fatalf("%s should expose non-empty dashboard_refs", pluginID)
			}
			for _, want := range []string{
				"cmdb_agent_plugin_credential_contract",
				"cmdb_credential_ref_resolve_contract",
				"cmdb_dashboard_import_runtime_contract",
			} {
				if !findXAgentPluginConfigContainsString(plugin.MissingContracts, want) {
					t.Fatalf("%s missing_contracts should include %s, got %#v", pluginID, want, plugin.MissingContracts)
				}
			}
		})
	}
}

func TestFindXAgentPluginCatalogCoversExpandedInputDirectoryEvidence(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := performAgentPluginCatalogRequest()
	if w.Code != http.StatusOK {
		t.Fatalf("plugin catalog should be 200, got %d body=%s", w.Code, w.Body.String())
	}

	var plugins []model.FindXAgentPlugin
	if err := json.Unmarshal(w.Body.Bytes(), &plugins); err != nil {
		t.Fatalf("decode plugin catalog: %v body=%s", err, w.Body.String())
	}
	collect := 0
	byID := make(map[string]model.FindXAgentPlugin, len(plugins))
	for _, plugin := range plugins {
		byID[plugin.ID] = plugin
		if plugin.Category == "collect" {
			collect++
		}
	}
	if collect < 80 {
		t.Fatalf("collect plugin catalog should cover expanded input directory evidence, got %d collect rows out of %d", collect, len(plugins))
	}
	for _, pluginID := range []string{
		"aliyun", "apache", "bind", "cadvisor", "chrony", "conntrack",
		"consul", "dcgm", "dns_query", "elasticsearch", "emc_unity",
		"ethtool", "filecount", "gnmi", "greenplum", "hadoop", "haproxy",
		"ipmi", "iptables", "ipvs", "jenkins", "kafka", "keepalived",
		"kubernetes", "ldap", "logstash", "mongodb", "mysql", "nats",
		"nginx", "node_exporter", "oracle", "phpfpm", "postgresql",
		"prometheus", "rabbitmq", "redfish", "redis", "rocketmq_offset",
		"smart", "snmp", "snmp_trap", "sqlserver", "supervisor",
		"system", "systemd", "tomcat", "vsphere", "x509_cert", "zookeeper",
	} {
		t.Run(pluginID, func(t *testing.T) {
			plugin, ok := byID[pluginID]
			if !ok {
				t.Fatalf("expanded input plugin %s should exist in catalog", pluginID)
			}
			if plugin.Category != "collect" || plugin.ConfigFormat != "toml" || len(plugin.SupportedOS) == 0 {
				t.Fatalf("expanded plugin %s should be collect toml with supported_os: %#v", pluginID, plugin)
			}
			if !findXAgentPluginConfigContainsString(plugin.Blockers, "PENDING") {
				t.Fatalf("expanded plugin %s must remain blocked by contract: %#v", pluginID, plugin.Blockers)
			}
			if len(plugin.ConfigSchemaContracts) == 0 ||
				!findXAgentPluginConfigContainsString(plugin.MissingContracts, "cmdb_plugin_config_schema_contract") ||
				!findXAgentPluginConfigContainsString(plugin.MissingContracts, "cmdb_agent_rollout_delivery_receipt_contract") ||
				!findXAgentPluginConfigContainsString(plugin.MissingContracts, "cmdb_agent_rollout_effect_receipt_contract") {
				t.Fatalf("expanded plugin %s missing runtime contracts: %#v", pluginID, plugin.MissingContracts)
			}
		})
	}
	assertFindXAgentPluginConfigNoFakeRuntimeState(t, w.Body.String())
	assertFindXAgentPluginConfigNoSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentPluginCatalogExpandedSecurityClasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := performAgentPluginCatalogRequest()
	if w.Code != http.StatusOK {
		t.Fatalf("plugin catalog should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var plugins []model.FindXAgentPlugin
	if err := json.Unmarshal(w.Body.Bytes(), &plugins); err != nil {
		t.Fatalf("decode plugin catalog: %v body=%s", err, w.Body.String())
	}
	byID := make(map[string]model.FindXAgentPlugin, len(plugins))
	for _, plugin := range plugins {
		byID[plugin.ID] = plugin
	}

	for _, pluginID := range []string{"aliyun", "cloudwatch", "greenplum", "sqlserver", "snmp", "vsphere"} {
		plugin := byID[pluginID]
		if !plugin.CredentialRequired || plugin.SecurityLevel != "credential-gated" {
			t.Fatalf("%s should stay credential gated, got %#v", pluginID, plugin)
		}
		for _, want := range []string{"cmdb_agent_plugin_credential_contract", "cmdb_credential_ref_resolve_contract"} {
			if !findXAgentPluginConfigContainsString(plugin.MissingContracts, want) {
				t.Fatalf("%s missing credential contract %s: %#v", pluginID, want, plugin.MissingContracts)
			}
		}
	}
	for _, pluginID := range []string{"dns_query", "http_response", "net_response", "ping", "prometheus", "x509_cert", "whois"} {
		plugin := byID[pluginID]
		if plugin.SecurityLevel != "network-gated" {
			t.Fatalf("%s should stay network gated, got %#v", pluginID, plugin)
		}
		if !findXAgentPluginConfigContainsString(plugin.MissingContracts, "cmdb_agent_plugin_network_policy_contract") {
			t.Fatalf("%s missing network policy contract: %#v", pluginID, plugin.MissingContracts)
		}
	}
	for _, pluginID := range []string{"gnmi", "redfish", "snmp", "vsphere"} {
		plugin := byID[pluginID]
		if !plugin.CredentialRequired || plugin.SecurityLevel != "credential-gated" {
			t.Fatalf("%s should keep credential gate while also requiring network policy, got %#v", pluginID, plugin)
		}
		if !findXAgentPluginConfigContainsString(plugin.MissingContracts, "cmdb_agent_plugin_network_policy_contract") {
			t.Fatalf("%s missing network policy contract: %#v", pluginID, plugin.MissingContracts)
		}
	}
	exec := byID["exec"]
	if exec.SecurityLevel != "high-risk" ||
		!findXAgentPluginConfigContainsString(exec.MissingContracts, "unsafe_plugin_policy_ref") {
		t.Fatalf("exec plugin must remain high-risk and policy-gated: %#v", exec)
	}
	assertFindXAgentPluginConfigNoFakeRuntimeState(t, w.Body.String())
	assertFindXAgentPluginConfigNoSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentPluginConfigPushBatchBlocksRuntimeExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := strings.NewReader(`{"agent_ids":["agent-a"],"plugins":[{"id":"redis","enabled":true,"config":"password=secret"}],"strategy":"all"}`)
	w := performAgentPluginConfigRequest(http.MethodPost, "/findx-agents/config-push", body, FindXAgentConfigPushBatch)
	if w.Code != http.StatusConflict {
		t.Fatalf("config push should fail close with 409, got %d body=%s", w.Code, w.Body.String())
	}
	for _, want := range []string{
		"PENDING",
		"cmdb.agent.plugin.config_push.runtime.v1",
		"cmdb_agent_config_push_executor_contract",
		"cmdb_agent_rollout_delivery_receipt_contract",
		"cmdb_agent_rollout_effect_receipt_contract",
		`"status":"blocked"`,
	} {
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("blocked config push response missing %q: %s", want, w.Body.String())
		}
	}
	assertFindXAgentPluginConfigNoFakeRuntimeState(t, w.Body.String())
	assertFindXAgentPluginConfigNoSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentPluginSingleConfigPushBlocksRuntimeExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := strings.NewReader(`{"mode":"single","config":{"redis":"password=secret"}}`)
	w := performSingleAgentConfigPushRequest(body)
	if w.Code != http.StatusConflict {
		t.Fatalf("single-agent config push should fail close with 409, got %d body=%s", w.Code, w.Body.String())
	}
	for _, want := range []string{
		"PENDING",
		"cmdb.agent.plugin.config_push.runtime.v1",
		"cmdb_agent_config_rollout_contract",
		"cmdb_agent_config_push_executor_contract",
		"cmdb_agent_rollout_delivery_receipt_contract",
		"cmdb_agent_rollout_effect_receipt_contract",
	} {
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("blocked single config push response missing %q: %s", want, w.Body.String())
		}
	}
	for _, forbidden := range []string{"config push triggered", `"event"`, `"status":"pending"`, "password=secret"} {
		if strings.Contains(strings.ToLower(w.Body.String()), forbidden) {
			t.Fatalf("single config push response leaked fake runtime/sensitive marker %q: %s", forbidden, w.Body.String())
		}
	}
	assertFindXAgentPluginConfigNoFakeRuntimeState(t, w.Body.String())
	assertFindXAgentPluginConfigNoSensitiveEcho(t, w.Body.String())
}

func TestFindXAgentPluginConfigRejectsAndRedactsSensitiveConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := strings.NewReader(`{"config":"password=secret\nurl=mysql://user:pass@db/app"}`)
	update := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/findx-agents/agent-sensitive/plugins/redis/config", body)
	req.Header.Set("Content-Type", "application/json")
	c, _ := gin.CreateTestContext(update)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "agent-sensitive"}, {Key: "pluginId", Value: "redis"}}
	UpdateFindXAgentPluginConfig(c)
	if update.Code != http.StatusConflict {
		t.Fatalf("sensitive plugin config should fail close with 409, got %d body=%s", update.Code, update.Body.String())
	}
	if !strings.Contains(update.Body.String(), "cmdb.agent.plugin.credential.v1") {
		t.Fatalf("sensitive plugin config response should include credential contract: %s", update.Body.String())
	}
	assertFindXAgentPluginConfigNoSensitiveEcho(t, update.Body.String())

	_, err := store.SavePluginConfig(model.FindXAgentPluginConfig{
		ID:       "agent-sensitive-read:redis",
		AgentID:  "agent-sensitive-read",
		PluginID: "redis",
		Enabled:  true,
		Config:   "token=secret",
		Status:   "blocked",
	})
	if err != nil {
		t.Fatalf("seed plugin config: %v", err)
	}
	read := performGetAgentPluginConfigRequest("agent-sensitive-read")
	if read.Code != http.StatusOK {
		t.Fatalf("read plugin config should be 200, got %d body=%s", read.Code, read.Body.String())
	}
	if !strings.Contains(read.Body.String(), "REDACTED_CONFIG") {
		t.Fatalf("sensitive stored config should be redacted: %s", read.Body.String())
	}
	assertFindXAgentPluginConfigNoSensitiveEcho(t, read.Body.String())
}

func TestFindXAgentPluginStartStopBlockRuntimeExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name    string
		path    string
		handler gin.HandlerFunc
		action  string
	}{
		{name: "start", path: "/findx-agents/agent-a/plugins/redis/start", handler: StartFindXAgentPlugin, action: "start"},
		{name: "stop", path: "/findx-agents/agent-a/plugins/redis/stop", handler: StopFindXAgentPlugin, action: "stop"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := performAgentPluginRuntimeActionRequest(tt.path, tt.handler)
			if w.Code != http.StatusConflict {
				t.Fatalf("%s should fail close with 409, got %d body=%s", tt.name, w.Code, w.Body.String())
			}
			for _, want := range []string{
				"PENDING",
				"cmdb.agent.plugin.runtime.action.v1",
				"cmdb_agent_plugin_runtime_executor_contract",
				"cmdb_agent_plugin_runtime_receipt_contract",
				`"action":"` + tt.action + `"`,
			} {
				if !strings.Contains(w.Body.String(), want) {
					t.Fatalf("blocked plugin runtime action response missing %q: %s", want, w.Body.String())
				}
			}
			assertFindXAgentPluginConfigNoFakeRuntimeState(t, w.Body.String())
			assertFindXAgentPluginConfigNoSensitiveEcho(t, w.Body.String())
		})
	}
}

func performSingleAgentConfigPushRequest(body *strings.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/findx-agents/agent-a/config-push", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "agent-a"}}
	ConfigPushFindXAgent(c)
	return w
}

func performAgentPluginCatalogRequest() *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/findx-agents/plugins", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	ListFindXAgentPlugins(c)
	return w
}

func performGetAgentPluginConfigRequest(agentID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/findx-agents/"+agentID+"/config", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: agentID}}
	GetFindXAgentConfig(c)
	return w
}

func performAgentPluginConfigRequest(method, path string, body *strings.Reader, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler(c)
	return w
}

func performAgentPluginRuntimeActionRequest(path string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "agent-a"},
		{Key: "pluginId", Value: "redis"},
	}
	handler(c)
	return w
}

func assertFindXAgentPluginConfigNoFakeRuntimeState(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{`"status":"success"`, `"status":"running"`, `"status":"succeeded"`, `"status":"applied"`, `"status":"installed"`, `"status":"delivered"`, `"status":"effective"`, `"status":"uninstalled"`, `"status":"rolled_back"`} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("plugin runtime response must not expose fake state %q: %s", forbidden, body)
		}
	}
}

func findXAgentPluginConfigContainsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func assertFindXAgentPluginConfigNoSensitiveEcho(t *testing.T, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{"password=secret", "token=", "cookie=", "mysql://", "private_key", "authorization"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("plugin runtime response leaked sensitive marker %q: %s", forbidden, body)
		}
	}
}
