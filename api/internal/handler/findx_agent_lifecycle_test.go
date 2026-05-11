package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestFindXAgentPackagesRemainBlockedWithoutRepositoryAndSignature(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "findx-agent", "java-app"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FINDX_AGENT_SOURCE_ROOT", root)

	pkg, ok := findAgentPackage("java-app")
	if !ok {
		t.Fatal("expected java package")
	}
	if pkg.SourceState != "LOCAL_SOURCE_PRESENT" {
		t.Fatalf("source state should record local source presence, got %q", pkg.SourceState)
	}
	if pkg.Status != "blocked" || pkg.Signature != "missing" {
		t.Fatalf("source directory must not imply ready package/signature, status=%q signature=%q", pkg.Status, pkg.Signature)
	}
	if !containsLifecycleTestString(pkg.Blockers, "PACKAGE_REPOSITORY_MISSING") || !containsLifecycleTestString(pkg.Blockers, "SIGNATURE_MISSING") {
		t.Fatalf("missing package/signature blockers: %#v", pkg.Blockers)
	}
}

func TestFindXAgentPackagesUseUnifiedFindXAgentDomains(t *testing.T) {
	packages := findXAgentPackages()
	if len(packages) < 10 {
		t.Fatalf("expected unified FindX Agent capability matrix, got %d packages", len(packages))
	}
	requiredDomains := map[string]bool{
		"基础 Agent": false,
		"基础采集":     false,
		"日志采集":     false,
		"应用链路":     false,
		"网关链路":     false,
		"前端体验":     false,
		"巡检诊断":     false,
	}
	for _, pkg := range packages {
		if strings.Contains(pkg.Name, "探针") {
			t.Fatalf("user-facing package name should use unified capability wording, got %q", pkg.Name)
		}
		if strings.Contains(pkg.Name, "SkyWalking") || strings.Contains(pkg.Name, "Categraf") || strings.Contains(pkg.Name, "Catpaw") {
			t.Fatalf("user-facing package name exposes external source: %q", pkg.Name)
		}
		if _, ok := requiredDomains[pkg.CapabilityDomain]; ok {
			requiredDomains[pkg.CapabilityDomain] = true
		}
		if len(pkg.ConfigTemplateIDs) == 0 {
			t.Fatalf("package %s should declare distributable config templates", pkg.ID)
		}
	}
	for domain, seen := range requiredDomains {
		if !seen {
			t.Fatalf("missing FindX Agent capability domain %q", domain)
		}
	}
}

func TestFindXAgentConfigTemplatesCoverUnifiedRollout(t *testing.T) {
	templates := findXAgentConfigTemplates()
	required := map[string]bool{
		"agent-core":       false,
		"metrics":          false,
		"logs":             false,
		"tracing":          false,
		"profiling":        false,
		"inspection":       false,
		"host-plugin":      false,
		"container-plugin": false,
		"gateway-plugin":   false,
		"browser-probe":    false,
	}
	for _, tpl := range templates {
		if strings.Contains(tpl.Name, "探针") {
			t.Fatalf("template name should not frame config rollout as probe-only, got %q", tpl.Name)
		}
		required[tpl.ID] = true
		if !tpl.RemoteDistribution {
			t.Fatalf("template %s should participate in unified remote distribution", tpl.ID)
		}
		if len(tpl.CapabilityPackages) == 0 || len(tpl.RolloutScopes) == 0 || tpl.RollbackPolicy == "" {
			t.Fatalf("template %s missing rollout metadata: %#v", tpl.ID, tpl)
		}
	}
	for id, seen := range required {
		if !seen {
			t.Fatalf("missing unified config template %q", id)
		}
	}
}

func TestFindXAgentPluginTemplatesExposeRemoteMutationMetadata(t *testing.T) {
	templates := findXAgentConfigTemplates()
	pluginTemplates := map[string]model.FindXAgentConfigTemplate{}
	for _, tpl := range templates {
		if strings.HasSuffix(tpl.ID, "-plugin") {
			pluginTemplates[tpl.ID] = tpl
		}
	}
	for _, id := range []string{"host-plugin", "container-plugin", "gateway-plugin"} {
		tpl, ok := pluginTemplates[id]
		if !ok {
			t.Fatalf("missing plugin template %q", id)
		}
		if tpl.PluginConfig == nil {
			t.Fatalf("plugin template %s should expose plugin config metadata", id)
		}
		spec := tpl.PluginConfig
		for _, field := range []string{"plugin_id", "plugin_version", "config_snippet_ref", "provider_mode", "remote_mutation", "rollback_ref", "audit_reason", "credential_ref"} {
			if !containsLifecycleTestString(tpl.Fields, field) {
				t.Fatalf("template %s missing remote mutation field %q: %#v", id, field, tpl.Fields)
			}
		}
		if spec.ConfigFormat != "toml" || !containsLifecycleTestString(spec.ProviderModes, "local") || !containsLifecycleTestString(spec.ProviderModes, "http") {
			t.Fatalf("template %s should retain TOML local/http provider semantics: %#v", id, spec)
		}
		if !spec.RemoteMutation || spec.RemoteMutationStatus != agentBlocked || !spec.CredentialRefRequired {
			t.Fatalf("template %s should be remotely mutable but blocked by contract: %#v", id, spec)
		}
		if len(spec.SourceEvidence) == 0 || strings.Contains(tpl.Name, "Categraf") {
			t.Fatalf("template %s missing source evidence or exposes source brand: %#v", id, tpl)
		}
	}
}

func TestFindXAgentPackageReferencesPluginTemplates(t *testing.T) {
	packages := findXAgentPackages()
	required := map[string]string{
		"host-collector":      "host-plugin",
		"container-collector": "container-plugin",
		"gateway-probe":       "gateway-plugin",
	}
	for _, pkg := range packages {
		if want, ok := required[pkg.ID]; ok {
			if !containsLifecycleTestString(pkg.ConfigTemplateIDs, want) {
				t.Fatalf("package %s should reference plugin template %s: %#v", pkg.ID, want, pkg.ConfigTemplateIDs)
			}
		}
	}
}

func TestFindXAgentPackagesExposePluginConfigMetadata(t *testing.T) {
	packages := findXAgentPackages()
	required := map[string]bool{
		"host-collector":      false,
		"container-collector": false,
		"gateway-probe":       false,
	}
	for _, pkg := range packages {
		if _, ok := required[pkg.ID]; !ok {
			continue
		}
		required[pkg.ID] = true
		if pkg.PluginConfig == nil {
			t.Fatalf("package %s should expose plugin_config metadata on package response", pkg.ID)
		}
		spec := pkg.PluginConfig
		if spec.ConfigFormat != "toml" || !containsLifecycleTestString(spec.ProviderModes, "local") || !containsLifecycleTestString(spec.ProviderModes, "http") {
			t.Fatalf("package %s should expose TOML local/http provider metadata: %#v", pkg.ID, spec)
		}
		if spec.ConfigSnippetRef == "" || !spec.RemoteMutation || spec.RemoteMutationStatus != agentBlocked {
			t.Fatalf("package %s should expose blocked remote mutation metadata: %#v", pkg.ID, spec)
		}
		if spec.CredentialRefRequired && containsLifecycleTestString(spec.RolloutMetadata, "credential_ref") {
			t.Fatalf("package %s should require credential reference without echoing it as rollout metadata: %#v", pkg.ID, spec.RolloutMetadata)
		}
	}
	for id, seen := range required {
		if !seen {
			t.Fatalf("missing package %s", id)
		}
	}
}

func TestFindXAgentConfigRolloutBlocksPluginRemoteMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	longReason := strings.Repeat("safe-audit-", 30)
	body := strings.NewReader(`{
		"template_id":"host-plugin",
		"target_ids":["target-a"],
		"plugin_id":"input.cpu",
		"plugin_version":"<PLUGIN_VERSION>",
		"config_version":"cfg-v1",
		"config_snippet_ref":"  <CONFIG_SNIPPET_REF>\r\n  ",
		"config_format":"toml",
		"provider_mode":"http",
		"remote_mutation":true,
		"canary_percent":10,
		"rollout_strategy":"灰度下发",
		"rollback_ref":"<ROLLBACK_REF>",
		"audit_reason":"` + longReason + `",
		"change_ticket":"token=secret-value",
		"credential_ref":"<CREDENTIAL_REF>",
		"metadata":{"executor_ref":"executor-ref"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/findx-agents/config-rollouts", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	CreateFindXAgentConfigRollout(c)
	if w.Code != http.StatusConflict {
		t.Fatalf("plugin remote mutation rollout should be blocked with 409, got %d body=%s", w.Code, w.Body.String())
	}
	assertPluginRemoteMutationRolloutResponse(t, w)
}

func assertPluginRemoteMutationRolloutResponse(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	payload := decodePluginRemoteMutationRollout(t, w)
	assertPluginRemoteMutationMetadata(t, payload)
	assertPluginRemoteMutationSanitized(t, payload, w.Body.String())
}

type pluginRemoteMutationRolloutPayload struct {
	Error string `json:"error"`
	Data  struct {
		TemplateID       string `json:"template_id"`
		PluginID         string `json:"plugin_id"`
		ConfigSnippetRef string `json:"config_snippet_ref"`
		ProviderMode     string `json:"provider_mode"`
		RemoteMutation   bool   `json:"remote_mutation"`
		CanaryPercent    int    `json:"canary_percent"`
		AuditReason      string `json:"audit_reason"`
		ChangeTicket     string `json:"change_ticket"`
		Status           string `json:"status"`
	} `json:"data"`
}

func decodePluginRemoteMutationRollout(t *testing.T, w *httptest.ResponseRecorder) pluginRemoteMutationRolloutPayload {
	t.Helper()
	var payload pluginRemoteMutationRolloutPayload
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid rollout response json: %v", err)
	}
	return payload
}

func assertPluginRemoteMutationMetadata(t *testing.T, payload pluginRemoteMutationRolloutPayload) {
	t.Helper()
	if !strings.Contains(payload.Error, agentBlocked) || payload.Data.Status != "blocked" {
		t.Fatalf("rollout should keep BLOCKED_BY_CONTRACT semantics: %#v", payload)
	}
	if payload.Data.TemplateID != "host-plugin" || payload.Data.PluginID != "input.cpu" || payload.Data.ProviderMode != "http" || !payload.Data.RemoteMutation || payload.Data.CanaryPercent != 10 {
		t.Fatalf("rollout response should echo plugin remote mutation metadata: %#v", payload.Data)
	}
	if payload.Data.ConfigSnippetRef != "<CONFIG_SNIPPET_REF>" {
		t.Fatalf("rollout response should keep config snippet reference, got %q", payload.Data.ConfigSnippetRef)
	}
}

func assertPluginRemoteMutationSanitized(t *testing.T, payload pluginRemoteMutationRolloutPayload, body string) {
	t.Helper()
	if strings.Contains(body, `"credential_ref":`) || strings.Contains(body, "<CREDENTIAL_REF>") {
		t.Fatalf("rollout response must not echo credential_ref: %s", body)
	}
	if strings.Contains(body, "secret-value") || payload.Data.ChangeTicket != "" {
		t.Fatalf("rollout response must drop sensitive-looking metadata: %#v body=%s", payload.Data, body)
	}
	if strings.ContainsAny(payload.Data.AuditReason, "\r\n\t") || len([]rune(payload.Data.AuditReason)) > 120 {
		t.Fatalf("rollout response should strip controls and cap audit metadata, got %q", payload.Data.AuditReason)
	}
}

func TestFindXAgentLifecyclePhasesDoNotPromoteInventoryToReady(t *testing.T) {
	phases := agentLifecyclePhases(3, 2, 1)
	for _, phase := range phases {
		if phase.Status == "ready" {
			t.Fatalf("phase %s should stay blocked until contract evidence exists", phase.Key)
		}
		if !strings.Contains(phase.Blocker, "BLOCKED_BY_CONTRACT") {
			t.Fatalf("phase %s should expose blocked contract, got %q", phase.Key, phase.Blocker)
		}
	}
}

func TestFindXAgentDataArrivalRequiresValidatorEvidence(t *testing.T) {
	agents := []*model.FindXAgent{{
		ID:           "agent-1",
		Status:       "online",
		Capabilities: []string{"metrics", "tracing", "logs"},
		LastSeen:     time.Now(),
	}}
	rows := agentDataArrival(agents)
	for _, row := range rows {
		if row.Status == "reported" {
			t.Fatalf("data arrival %s must not be reported from capability heartbeat only", row.Kind)
		}
		if row.EvidenceCount != 0 {
			t.Fatalf("data arrival %s must not synthesize evidence count, got %d", row.Kind, row.EvidenceCount)
		}
		if !strings.Contains(row.Blocker, "BLOCKED_BY_CONTRACT") {
			t.Fatalf("data arrival %s should be blocked, got %q", row.Kind, row.Blocker)
		}
	}
}

func TestFindXAgentHeartbeatRequiresSharedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("FINDX_AGENT_TOKEN", "")
	viper.Set("findx_agents.shared_token", "unit-agent-token")
	viper.Set("findx_agents.allow_anonymous", false)
	t.Cleanup(func() {
		viper.Set("findx_agents.shared_token", "")
		viper.Set("findx_agents.allow_anonymous", false)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/findx-agents/heartbeat", strings.NewReader(`{"ident":"agent-1","ip":"127.0.0.1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	FindXAgentHeartbeat(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("heartbeat without token should be 401, got %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/findx-agents/heartbeat", strings.NewReader(`{"ident":"agent-1","ip":"127.0.0.1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Token", "wrong-token")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	FindXAgentHeartbeat(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("heartbeat with wrong token should be 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func containsLifecycleTestString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
