package handler

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func findXAgentPackages() []model.FindXAgentPackage {
	now := time.Now().Format(time.RFC3339)
	defs := findXAgentPackageDefs()
	rows := make([]model.FindXAgentPackage, 0, len(defs))
	for _, def := range defs {
		rows = append(rows, agentPackage(def, now))
	}
	return rows
}

func agentPackage(def agentPackageDef, updatedAt string) model.FindXAgentPackage {
	state := agentSourceState(def.id)
	installEnvironment := agentPackageInstallEnvironment(def.id)
	blockers := agentPackageBlockers(def.id)
	if state != "LOCAL_SOURCE_PRESENT" {
		blockers = append([]string{"LOCAL_SOURCE_MISSING"}, blockers...)
	}
	if def.pluginConfig != nil {
		blockers = append(blockers, pluginConfigContractBlockers()...)
	}
	return model.FindXAgentPackage{
		ID:                 def.id,
		Name:               def.name,
		CapabilityDomain:   def.domain,
		Runtime:            def.runtime,
		SupportedOS:        def.osList,
		PackageShape:       def.shape,
		TelemetryKinds:     def.telemetryKinds,
		ConfigKeys:         def.configKeys,
		ConfigTemplateIDs:  def.configTemplateIDs,
		PluginConfig:       def.pluginConfig,
		InstallEnvironment: installEnvironment,
		EnvironmentMatrix:  agentPackageEnvironmentMatrix(def, installEnvironment, blockers, state),
		InstallMethods:     []string{"linux-curl", "windows-cmd", "windows-powershell", "ssh", "winrm", "helm"},
		SourceState:        state,
		Status:             "blocked",
		Blockers:           blockers,
		Signature:          "missing",
		UpdatedAt:          updatedAt,
	}
}

func agentSourceState(packageID string) string {
	for _, root := range agentSourceRoots() {
		for _, candidate := range agentSourceCandidates(packageID) {
			if _, err := os.Stat(filepath.Join(root, candidate)); err == nil {
				return "LOCAL_SOURCE_PRESENT"
			}
		}
	}
	return "LOCAL_SOURCE_MISSING"
}

func agentSourceCandidates(packageID string) []string {
	candidates := []string{
		filepath.Join("findx-agent", packageID),
		"findx-agent-" + packageID,
		packageID,
	}
	sourceHints := map[string][]string{
		"host-collector":      {sourceHint("cate", "graf-main"), filepath.Join(sourceHint("cate", "graf-main (1)"), sourceHint("cate", "graf-main"))},
		"container-collector": {sourceHint("cate", "graf-main"), filepath.Join(sourceHint("cate", "graf-main (1)"), sourceHint("cate", "graf-main"))},
		"log-collector":       {sourceHint("cate", "graf-main"), filepath.Join(sourceHint("cate", "graf-main (1)"), sourceHint("cate", "graf-main"))},
		"inspection-runner":   {sourceHint("cat", "paw-master"), filepath.Join(sourceHint("cat", "paw-master"), sourceHint("cat", "paw-master"))},
		"java-app":            {sourceHint("sky", "walking-java"), sourceHint("sky", "walking-java-main")},
		"python-app":          {sourceHint("sky", "walking-python"), sourceHint("sky", "walking-python-main")},
		"nodejs-app":          {sourceHint("sky", "walking-nodejs"), sourceHint("sky", "walking-nodejs-main")},
		"php-app":             {sourceHint("sky", "walking-php"), sourceHint("sky", "walking-php-main")},
		"go-app":              {sourceHint("sky", "walking-go"), sourceHint("sky", "walking-go-main")},
		"rust-app":            {sourceHint("sky", "walking-rust"), sourceHint("sky", "walking-rust-main")},
		"ruby-app":            {sourceHint("sky", "walking-ruby"), sourceHint("sky", "walking-ruby-main")},
		"gateway-probe":       {sourceHint("sky", "walking-nginx-lua"), sourceHint("sky", "walking-kong")},
		"browser-client":      {sourceHint("sky", "walking-client-js"), sourceHint("sky", "walking-client-js-main")},
	}
	return append(candidates, sourceHints[packageID]...)
}

func agentSourceRoots() []string {
	roots := []string{}
	for _, configured := range filepath.SplitList(os.Getenv("FINDX_AGENT_SOURCE_ROOT")) {
		roots = appendSourceRoot(roots, configured)
	}
	if cwd, err := os.Getwd(); err == nil {
		for dir := cwd; ; dir = filepath.Dir(dir) {
			roots = appendSourceRoot(roots, filepath.Join(dir, "..", "平台源码"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}
	for _, fallback := range []string{
		`D:\项目迁移文件\平台源码`,
		`D:\平台源码`,
		`/mnt/d/项目迁移文件/平台源码`,
		`/mnt/d/平台源码`,
	} {
		roots = appendSourceRoot(roots, fallback)
	}
	return roots
}

func appendSourceRoot(roots []string, root string) []string {
	root = strings.TrimSpace(root)
	if root == "" {
		return roots
	}
	cleaned := filepath.Clean(root)
	for _, existing := range roots {
		if strings.EqualFold(existing, cleaned) {
			return roots
		}
	}
	return append(roots, cleaned)
}

func findXAgentConfigTemplates() []model.FindXAgentConfigTemplate {
	now := time.Now().Format(time.RFC3339)
	defs := findXAgentConfigTemplateDefs()
	rows := make([]model.FindXAgentConfigTemplate, 0, len(defs))
	for _, def := range defs {
		rows = append(rows, configTemplate(def, now))
	}
	return rows
}

func configTemplate(def agentConfigTemplateDef, updatedAt string) model.FindXAgentConfigTemplate {
	return model.FindXAgentConfigTemplate{
		ID:                 def.id,
		Name:               def.name,
		ConfigKind:         def.kind,
		Scope:              def.scope,
		Fields:             def.fields,
		TargetScopes:       def.targetScopes,
		RolloutScopes:      def.rolloutScopes,
		RolloutStrategies:  []string{"保存模板", "灰度下发", "全量下发", "回滚"},
		RemoteDistribution: true,
		RollbackPolicy:     def.rollbackPolicy,
		CapabilityPackages: def.capabilityPackages,
		PluginConfig:       def.pluginConfig,
		Status:             "blocked",
		Blocker:            "BLOCKED_BY_CONTRACT: 模板保存、凭据引用、远程修改、下发、回滚和审计协议未开放",
		UpdatedAt:          updatedAt,
	}
}

func pluginConfigSpec(templateID, pluginID, reloadStrategy string) *model.FindXAgentPluginConfigSpec {
	return &model.FindXAgentPluginConfigSpec{
		PluginID:              pluginID,
		PluginVersion:         "<PLUGIN_VERSION>",
		ConfigFormat:          "toml",
		ConfigSnippetRef:      "<CONFIG_SNIPPET_REF>",
		ProviderModes:         []string{"local", "http"},
		ReloadStrategy:        reloadStrategy,
		RestartStrategy:       "restart-if-plugin-requires",
		RemoteMutation:        true,
		RemoteMutationStatus:  agentBlocked,
		RolloutMetadata:       pluginConfigRolloutMetadata(),
		CredentialRefRequired: true,
		AuditEvent:            "findx_agent.plugin_config.remote_mutation.requested",
		SourceEvidence:        pluginSourceEvidence(templateID),
		PluginSourceMap:       pluginSourceMapForTemplate(templateID),
		PlatformMatrix:        pluginPlatformMatrixForTemplate(templateID),
		SecurityProfile:       pluginSecurityProfileForTemplate(templateID),
		Blockers:              pluginConfigContractBlockers(),
	}
}

func pluginConfigRolloutMetadata() []string {
	return []string{
		"config_version",
		"canary_percent",
		"rollback_ref",
		"change_ticket",
		"audit_reason",
		"plugin_config_writer_ref",
		"reload_command_ref",
		"reload_receipt_ref",
		"drift_check_ref",
		"evidence_chain_ref",
		"rollback_receipt_ref",
		"provider_endpoint_ref",
		"provider_response_version_ref",
		"provider_checksum_ref",
		"checksum_registry_ref",
		"provider_headers_ref",
		"provider_auth_ref",
		"provider_tls_ref",
		"reload_interval_ref",
		"timeout_ref",
		"config_serving_receipt_ref",
	}
}

func pluginConfigContractBlockers() []string {
	return []string{
		"PLUGIN_CONFIG_WRITER_MISSING",
		"RELOAD_RECEIPT_MISSING",
		"DRIFT_CHECK_MISSING",
		"ROLLBACK_CONTRACT_MISSING",
	}
}

func pluginSourceEvidence(templateID string) []string {
	base := []string{
		"conf/config.toml: [global].providers supports local/http, [global.labels], [heartbeat]",
		"doc/provider.toml: http_provider remote_url, headers, auth, timeout, reload_interval",
		"inputs/http_provider.go: GET remote_url with version, response version plus configs[input][checksum]{config,format}, checksum diff reloads input",
	}
	switch templateID {
	case "host-plugin":
		return append(base, "conf/input.cpu, conf/input.mem, conf/input.disk, conf/input.net TOML plugin directories")
	case "container-plugin":
		return append(base, "conf/input.docker, conf/input.cadvisor, conf/input.kubernetes TOML plugin directories")
	case "gateway-plugin":
		return append(base, "gateway plugin template keeps FindX routing and reload metadata")
	default:
		return base
	}
}
