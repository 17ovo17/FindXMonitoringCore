package handler

import (
	"ai-workbench-api/internal/model"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	status := "ready"
	if state != "LOCAL_SOURCE_PRESENT" {
		status = "pending"
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
		Status:             status,
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
		Status:             "ready",
		Blocker:            "",
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


const (
	environmentToolMissing        = "TOOL_EVIDENCE_MISSING"
	environmentPackageMissing     = "PACKAGE_MISSING"
	environmentSourceMissing      = "LOCAL_SOURCE_MISSING"
	environmentTestOnlyRepository = "PACKAGE_REPOSITORY_TEST_ONLY"
	environmentTestOnlySignature  = "SIGNATURE_TEST_ONLY"
)

type packageEnvironmentMethod struct {
	platform string
	method   string
}

var packageEnvironmentMethods = []packageEnvironmentMethod{
	{platform: "linux", method: "curl -kfsSL"},
	{platform: "windows", method: "certutil -urlcache -f"},
	{platform: "windows", method: "Invoke-WebRequest"},
	{platform: "linux", method: "SSH"},
	{platform: "windows", method: "WinRM"},
	{platform: "kubernetes", method: "Helm"},
	{platform: "kubernetes", method: "Operator"},
	{platform: "kubernetes", method: "DaemonSet"},
	{platform: "kubernetes", method: "Sidecar"},
	{platform: "kubernetes", method: "InitContainer"},
}

func agentPackageEnvironmentMatrix(
	def agentPackageDef,
	environment model.FindXAgentInstallEnvironment,
	blockers []string,
	sourceState string,
) []model.FindXAgentPackageEnvironmentMatrixRow {
	matrix := make([]model.FindXAgentPackageEnvironmentMatrixRow, 0, len(packageEnvironmentMethods))
	for _, method := range packageEnvironmentMethods {
		row := packageEnvironmentRow(def, environment, blockers, sourceState, method)
		matrix = append(matrix, row)
	}
	return matrix
}

func packageEnvironmentRow(
	def agentPackageDef,
	environment model.FindXAgentInstallEnvironment,
	blockers []string,
	sourceState string,
	method packageEnvironmentMethod,
) model.FindXAgentPackageEnvironmentMatrixRow {
	state := packageEnvironmentSourceState(def.id, sourceState)
	return model.FindXAgentPackageEnvironmentMatrixRow{
		Platform:            method.platform,
		InstallMethod:       method.method,
		ToolEvidence:        packageEnvironmentToolEvidence(environment, method.platform),
		SourceState:         state,
		PackageState:        packageEnvironmentPackageState(def.id, blockers, state),
		Executor:            packageEnvironmentBlocked("EXECUTOR_CONTRACT_MISSING", blockers),
		ServiceRegistration: packageEnvironmentBlocked("SERVICE_REGISTRATION_CONTRACT_MISSING", blockers),
		ConfigDelivery:      packageEnvironmentConfigDelivery(def.pluginConfig != nil),
		Uninstall:           packageEnvironmentBlocked("UNINSTALL_CONTRACT_MISSING", blockers),
		Rollback:            packageEnvironmentBlocked("ROLLBACK_CONTRACT_MISSING", blockers),
		DataArrival:         packageEnvironmentBlocked("DATA_ARRIVAL_CONTRACT_MISSING", blockers),
		Blocker:             packageEnvironmentBlocker(def.id, blockers, state),
	}
}

func packageEnvironmentSourceState(packageID, sourceState string) string {
	if isProbePackageWithoutLocalSource(packageID) {
		return environmentSourceMissing
	}
	if strings.TrimSpace(sourceState) == "" {
		return environmentSourceMissing
	}
	return sourceState
}

func isProbePackageWithoutLocalSource(packageID string) bool {
	return strings.HasSuffix(packageID, "-app") ||
		packageID == "gateway-probe" ||
		packageID == "browser-client"
}

func packageEnvironmentToolEvidence(environment model.FindXAgentInstallEnvironment, platform string) string {
	if packageEnvironmentHasReadyTool(environment, platform) &&
		containsLifecycleString(environment.Blockers, environmentTestOnlyRepository) {
		return strings.Join([]string{
			environmentTestOnlyRepository,
			environmentTestOnlySignature,
			"TOOL_EVIDENCE_TEST_ONLY",
		}, ";")
	}
	return packageEnvironmentBlocked(environmentToolMissing, environment.Blockers)
}

func packageEnvironmentHasReadyTool(environment model.FindXAgentInstallEnvironment, platform string) bool {
	for _, tool := range environment.Tools {
		if strings.EqualFold(tool.Status, "ready") && strings.EqualFold(tool.OS, platform) {
			return true
		}
	}
	return false
}

func packageEnvironmentPackageState(packageID string, blockers []string, sourceState string) string {
	if sourceState == environmentSourceMissing || isProbePackageWithoutLocalSource(packageID) {
		return environmentPackageMissing
	}
	if containsLifecycleString(blockers, environmentTestOnlyRepository) {
		return strings.Join([]string{
			environmentTestOnlyRepository,
			environmentTestOnlySignature,
			"PRODUCTION_PACKAGE_REPOSITORY_MISSING",
			"PRODUCTION_SIGNATURE_MISSING",
		}, ";")
	}
	return packageEnvironmentBlocked(environmentPackageMissing, blockers)
}

func packageEnvironmentConfigDelivery(hasPluginConfig bool) string {
	if !hasPluginConfig {
		return packageEnvironmentBlocked("CONFIG_DELIVERY_CONTRACT_MISSING", nil)
	}
	return strings.Join([]string{
		"FINDX_AGENT_CONTROL_PLANE_ENTRY",
		"REMOTE_MUTATION_PENDING",
		"RELOAD_PENDING",
		"DRIFT_PENDING",
		"ROLLBACK_PENDING",
		"RECEIPT_PENDING",
	}, ";")
}

func packageEnvironmentBlocker(packageID string, blockers []string, sourceState string) string {
	if sourceState == environmentSourceMissing || isProbePackageWithoutLocalSource(packageID) {
		return agentBlocked + ": LOCAL_SOURCE_MISSING;PACKAGE_MISSING"
	}
	return packageEnvironmentBlocked("ENVIRONMENT_MATRIX_PENDING", blockers)
}

func packageEnvironmentBlocked(reason string, blockers []string) string {
	values := []string{agentBlocked}
	values = append(values, reason)
	values = append(values, blockers...)
	return strings.Join(uniquePackageRepositoryBlockers(values), ":")
}

func containsLifecycleString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}


const (
	pluginConfigBlockedStatus = agentBlocked
	pluginConfigFormatTOML    = "toml"
)

type pluginSourceCatalogDef struct {
	id         string
	category   string
	configFile string
	platforms  []string
	security   string
	unsafe     bool
	blockers   []string
	evidence   []string
	policyRef  string
	sourceDir  string
	configPath string
	summaryRef string
}

func pluginSourceMapForTemplate(templateID string) []model.FindXAgentPluginSourceSpec {
	defs := pluginSourceDefsForTemplate(templateID)
	rows := make([]model.FindXAgentPluginSourceSpec, 0, len(defs))
	for _, def := range defs {
		rows = append(rows, model.FindXAgentPluginSourceSpec{
			PluginID:                 def.id,
			PluginCategory:           def.category,
			SourceDirectories:        []string{def.sourceDir},
			ConfigPaths:              []string{def.configPath},
			ConfigFormat:             pluginConfigFormatTOML,
			SupportedPlatforms:       def.platforms,
			SecurityLevel:            def.security,
			UnsafePlugin:             def.unsafe,
			UnsafePluginPolicyRef:    def.policyRef,
			RemoteMutationStatus:     pluginConfigBlockedStatus,
			Blockers:                 pluginSourceBlockers(def),
			SourceEvidence:           def.evidence,
			SourceEvidenceSummaryRef: def.summaryRef,
		})
	}
	return rows
}

func pluginPlatformMatrixForTemplate(templateID string) []model.FindXAgentPluginPlatformSpec {
	if templateID == "gateway-plugin" {
		return nil
	}
	return []model.FindXAgentPluginPlatformSpec{
		{
			Platform:      "Linux",
			ConfigPath:    "/etc/findx-agent/conf/input.<plugin>/<plugin>.toml",
			ConfigFormat:  pluginConfigFormatTOML,
			ReloadSupport: "HUP reload evidence exists in upstream service code; FindX writer and receipt contracts are absent",
			ReceiptRequirements: []string{
				"plugin_config_writer_ref",
				"reload_receipt_ref",
				"drift_check_ref",
				"data_arrival_receipt_ref",
			},
			Status: pluginConfigBlockedStatus,
			Blockers: []string{
				"PLUGIN_CONFIG_WRITER_MISSING",
				"RELOAD_RECEIPT_MISSING",
				"DATA_ARRIVAL_RECEIPT_MISSING",
			},
		},
		{
			Platform:      "Windows",
			ConfigPath:    "C:\\ProgramData\\FindX\\Agent\\conf\\input.<plugin>\\<plugin>.toml",
			ConfigFormat:  pluginConfigFormatTOML,
			ReloadSupport: "Windows HUP is not accepted; service restart receipt is required",
			ReceiptRequirements: []string{
				"windows_service_restart_receipt_ref",
				"drift_check_ref",
				"data_arrival_receipt_ref",
			},
			Status: pluginConfigBlockedStatus,
			Blockers: []string{
				"WINDOWS_HUP_NOT_SUPPORTED",
				"SERVICE_RESTART_RECEIPT_MISSING",
				"DATA_ARRIVAL_RECEIPT_MISSING",
			},
		},
		{
			Platform:      "Kubernetes",
			ConfigPath:    "ConfigMap: findx-agent-plugin-config",
			ConfigFormat:  pluginConfigFormatTOML,
			ReloadSupport: "ConfigMap and DaemonSet rollout receipts are required",
			Selectors:     []string{"namespace_selector", "workload_selector", "agent_selector"},
			ReceiptRequirements: []string{
				"configmap_write_receipt_ref",
				"daemonset_rollout_receipt_ref",
				"rollback_receipt_ref",
				"data_arrival_receipt_ref",
			},
			Status: pluginConfigBlockedStatus,
			Blockers: []string{
				"CONFIGMAP_WRITER_MISSING",
				"DAEMONSET_ROLLOUT_RECEIPT_MISSING",
				"DATA_ARRIVAL_RECEIPT_MISSING",
			},
		},
	}
}

func pluginSecurityProfileForTemplate(templateID string) model.FindXAgentPluginSecurityProfile {
	defs := pluginSourceDefsForTemplate(templateID)
	profile := model.FindXAgentPluginSecurityProfile{
		SecurityLevel: "medium",
		Blockers: []string{
			"REMOTE_MUTATION_PENDING",
			"CREDENTIAL_REF_GATE_MISSING",
			"URL_ALLOWLIST_GATE_MISSING",
			"TLS_AND_TIMEOUT_GATE_MISSING",
			"ERROR_SANITIZATION_GATE_MISSING",
			"DATA_ARRIVAL_RECEIPT_MISSING",
		},
		EvidenceRefs: []string{
			"inputs/http_provider.go",
			"inputs/local_provider.go",
			"inputs/provider_manager.go",
		},
	}
	for _, def := range defs {
		if def.unsafe {
			profile.SecurityLevel = "critical"
			profile.UnsafePluginPolicyRef = def.policyRef
			profile.UnsafePluginIDs = append(profile.UnsafePluginIDs, def.id)
			profile.BlockedPluginIDs = append(profile.BlockedPluginIDs, def.id)
		}
	}
	return profile
}

func pluginSourceDefsForTemplate(templateID string) []pluginSourceCatalogDef {
	defs := allCategrafPluginSourceDefs()
	switch templateID {
	case "host-plugin":
		return defs
	case "container-plugin":
		return filterPluginSourceDefs(defs, map[string]bool{"input.docker": true})
	default:
		return nil
	}
}

func allCategrafPluginSourceDefs() []pluginSourceCatalogDef {
	defs := []pluginSourceCatalogDef{
		pluginSourceDef("input.cpu", "host-metrics", "conf/input.cpu", "conf/input.cpu/cpu.toml", "low", []string{"Linux", "Windows"}, nil),
		pluginSourceDef("input.mem", "host-metrics", "conf/input.mem", "conf/input.mem/mem.toml", "low", []string{"Linux", "Windows"}, nil),
		pluginSourceDef("input.disk", "host-metrics", "conf/input.disk", "conf/input.disk/disk.toml", "medium", []string{"Linux", "Windows"}, nil),
		pluginSourceDef("input.diskio", "host-metrics", "conf/input.diskio", "conf/input.diskio/diskio.toml", "medium", []string{"Linux", "Windows"}, nil),
		pluginSourceDef("input.net", "host-metrics", "conf/input.net", "conf/input.net/net.toml", "medium", []string{"Linux", "Windows"}, nil),
		pluginSourceDef("input.processes", "process-metrics", "conf/input.processes", "conf/input.processes/processes.toml", "medium", []string{"Linux", "Windows"}, nil),
		pluginSourceDef("input.procstat", "process-metrics", "conf/input.procstat", "conf/input.procstat/procstat.toml", "high", []string{"Linux", "Windows"}, []string{"PROCESS_SELECTOR_GATE_MISSING"}),
		pluginSourceDef("input.docker", "container-metrics", "conf/input.docker", "conf/input.docker/docker.toml", "high", []string{"Linux", "Kubernetes"}, []string{"CONTAINER_SOCKET_ACCESS_GATE_MISSING"}),
		pluginSourceDef("input.mysql", "database-metrics", "conf/input.mysql", "conf/input.mysql/mysql.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.postgresql", "database-metrics", "conf/input.postgresql", "conf/input.postgresql/postgresql.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.redis", "cache-metrics", "conf/input.redis", "conf/input.redis/redis.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.mongodb", "database-metrics", "conf/input.mongodb", "conf/input.mongodb/mongodb.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.elasticsearch", "search-metrics", "conf/input.elasticsearch", "conf/input.elasticsearch/elasticsearch.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, httpScrapePluginBlockers()),
		pluginSourceDef("input.nginx", "http-service-metrics", "conf/input.nginx", "conf/input.nginx/nginx.toml", "high", []string{"Linux", "Kubernetes"}, httpScrapePluginBlockers()),
		pluginSourceDef("input.prometheus", "prometheus-scrape", "conf/input.prometheus", "conf/input.prometheus/prometheus.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, httpScrapePluginBlockers()),
		unsafeExecPluginSourceDef(),
	}
	defs = append(defs, expandedPluginSourceDefs()...)
	return dedupePluginSourceDefs(defs)
}

func expandedPluginSourceDefs() []pluginSourceCatalogDef {
	return []pluginSourceCatalogDef{
		pluginSourceDef("input.aliyun", "cloud-metrics", "conf/input.aliyun", "conf/input.aliyun/aliyun.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.apache", "http-service-metrics", "conf/input.apache", "conf/input.apache/apache.toml", "high", []string{"Linux", "Windows"}, httpScrapePluginBlockers()),
		pluginSourceDef("input.bind", "dns-service-metrics", "conf/input.bind", "conf/input.bind/bind.toml", "high", []string{"Linux", "Windows"}, httpScrapePluginBlockers()),
		pluginSourceDef("input.cadvisor", "container-metrics", "conf/input.cadvisor", "conf/input.cadvisor/cadvisor.toml", "high", []string{"Linux", "Kubernetes"}, httpScrapePluginBlockers()),
		pluginSourceDef("input.chrony", "availability-metrics", "conf/input.chrony", "conf/input.chrony/chrony.toml", "medium", []string{"Linux"}, nil),
		pluginSourceDef("input.clickhouse", "database-metrics", "conf/input.clickhouse", "conf/input.clickhouse/clickhouse.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.cloudwatch", "cloud-metrics", "conf/input.cloudwatch", "conf/input.cloudwatch/cloudwatch.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.conntrack", "network-metrics", "conf/input.conntrack", "conf/input.conntrack/conntrack.toml", "medium", []string{"Linux"}, nil),
		pluginSourceDef("input.consul", "service-registry-metrics", "conf/input.consul", "conf/input.consul/consul.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, httpScrapePluginBlockers()),
		pluginSourceDef("input.dcgm", "accelerator-metrics", "conf/input.dcgm", "conf/input.dcgm/dcgm.toml", "medium", []string{"Linux", "Kubernetes"}, nil),
		pluginSourceDef("input.dns_query", "availability-metrics", "conf/input.dns_query", "conf/input.dns_query/dns_query.toml", "high", []string{"Linux", "Windows"}, networkProbePluginBlockers()),
		pluginSourceDef("input.emc_unity", "device-metrics", "conf/input.emc_unity", "conf/input.emc_unity/emc_unity.toml", "high", []string{"Linux", "Windows"}, credentialPluginBlockers()),
		pluginSourceDef("input.ethtool", "network-metrics", "conf/input.ethtool", "conf/input.ethtool/ethtool.toml", "medium", []string{"Linux"}, nil),
		pluginSourceDef("input.filecount", "file-metrics", "conf/input.filecount", "conf/input.filecount/filecount.toml", "high", []string{"Linux", "Windows"}, pathScopedPluginBlockers()),
		pluginSourceDef("input.gnmi", "network-metrics", "conf/input.gnmi", "conf/input.gnmi/gnmi.toml", "high", []string{"Linux", "Windows"}, credentialPluginBlockers()),
		pluginSourceDef("input.greenplum", "database-metrics", "conf/input.greenplum", "conf/input.greenplum/greenplum.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.hadoop", "data-platform-metrics", "conf/input.hadoop", "conf/input.hadoop/hadoop.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, httpScrapePluginBlockers()),
		pluginSourceDef("input.haproxy", "load-balancer-metrics", "conf/input.haproxy", "conf/input.haproxy/haproxy.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, httpScrapePluginBlockers()),
		pluginSourceDef("input.http_response", "availability-metrics", "conf/input.http_response", "conf/input.http_response/http_response.toml", "high", []string{"Linux", "Windows"}, networkProbePluginBlockers()),
		pluginSourceDef("input.ipmi", "device-metrics", "conf/input.ipmi", "conf/input.ipmi/conf.toml", "critical", []string{"Linux"}, deviceAccessPluginBlockers()),
		pluginSourceDef("input.kafka", "queue-metrics", "conf/input.kafka", "conf/input.kafka/kafka.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.kernel", "host-metrics", "conf/input.kernel", "conf/input.kernel/kernel.toml", "medium", []string{"Linux"}, nil),
		pluginSourceDef("input.kernel_vmstat", "host-metrics", "conf/input.kernel_vmstat", "conf/input.kernel_vmstat/kernel_vmstat.toml", "medium", []string{"Linux"}, nil),
		pluginSourceDef("input.kubernetes", "workload-metrics", "conf/input.kubernetes", "conf/input.kubernetes/kubernetes.toml", "high", []string{"Linux", "Kubernetes"}, httpScrapePluginBlockers()),
		pluginSourceDef("input.linux_sysctl_fs", "host-metrics", "conf/input.linux_sysctl_fs", "conf/input.linux_sysctl_fs/linux_sysctl_fs.toml", "medium", []string{"Linux"}, nil),
		pluginSourceDef("input.netstat", "network-metrics", "conf/input.netstat", "conf/input.netstat/netstat.toml", "medium", []string{"Linux", "Windows"}, nil),
		pluginSourceDef("input.oracle", "database-metrics", "conf/input.oracle", "conf/input.oracle/oracle.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.rabbitmq", "queue-metrics", "conf/input.rabbitmq", "conf/input.rabbitmq/rabbitmq.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.redfish", "device-metrics", "conf/input.redfish", "conf/input.redfish/redfish.toml", "high", []string{"Linux", "Windows"}, credentialPluginBlockers()),
		pluginSourceDef("input.redis_sentinel", "cache-metrics", "conf/input.redis_sentinel", "conf/input.redis_sentinel/redis_sentinel.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.smart", "device-metrics", "conf/input.smart", "conf/input.smart/smart.toml", "critical", []string{"Linux"}, deviceAccessPluginBlockers()),
		pluginSourceDef("input.snmp", "network-device-metrics", "conf/input.snmp", "conf/input.snmp/snmp.toml", "high", []string{"Linux", "Windows"}, credentialPluginBlockers()),
		pluginSourceDef("input.sqlserver", "database-metrics", "conf/input.sqlserver", "conf/input.sqlserver/sqlserver.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, credentialPluginBlockers()),
		pluginSourceDef("input.system", "host-metrics", "conf/input.system", "conf/input.system/system.toml", "low", []string{"Linux", "Windows"}, nil),
		pluginSourceDef("input.tomcat", "http-service-metrics", "conf/input.tomcat", "conf/input.tomcat/tomcat.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, httpScrapePluginBlockers()),
		pluginSourceDef("input.vsphere", "virtualization-metrics", "conf/input.vsphere", "conf/input.vsphere/vsphere.toml", "high", []string{"Linux", "Windows"}, credentialPluginBlockers()),
		pluginSourceDef("input.whois", "availability-metrics", "conf/input.whois", "conf/input.whois/whois.toml", "high", []string{"Linux", "Windows"}, networkProbePluginBlockers()),
		pluginSourceDef("input.x509_cert", "availability-metrics", "conf/input.x509_cert", "conf/input.x509_cert/x509_cert.toml", "high", []string{"Linux", "Windows"}, networkProbePluginBlockers()),
		pluginSourceDef("input.zookeeper", "coordination-service-metrics", "conf/input.zookeeper", "conf/input.zookeeper/zookeeper.toml", "high", []string{"Linux", "Windows", "Kubernetes"}, httpScrapePluginBlockers()),
	}
}

func pluginSourceDef(id, category, sourceDir, configPath, security string, platforms []string, blockers []string) pluginSourceCatalogDef {
	return pluginSourceCatalogDef{
		id:         id,
		category:   category,
		sourceDir:  sourceDir,
		configPath: configPath,
		platforms:  platforms,
		security:   security,
		blockers:   blockers,
		evidence: []string{
			configPath,
			"inputs/local_provider.go",
			"inputs/http_provider.go",
			"inputs/provider_manager.go",
			"agent/install/service_linux.go",
			"agent/install/service_windows.go",
		},
		summaryRef: "findx-agent-plugin-source-map/" + id,
	}
}

func unsafeExecPluginSourceDef() pluginSourceCatalogDef {
	def := pluginSourceDef(
		"input.exec",
		"command-metrics",
		"conf/input.exec",
		"conf/input.exec/exec.toml",
		"critical",
		[]string{"Linux", "Windows", "Kubernetes"},
		[]string{
			"UNSAFE_PLUGIN_PENDING",
			"REMOTE_COMMAND_EXECUTION_NOT_ALLOWED",
			"STRUCTURED_TOOL_POLICY_REQUIRED",
		},
	)
	def.unsafe = true
	def.policyRef = "findx-agent/security/remote-command-disabled"
	return def
}

func pluginSourceBlockers(def pluginSourceCatalogDef) []string {
	blockers := []string{
		"REMOTE_MUTATION_PENDING",
		"PLUGIN_CONFIG_WRITER_MISSING",
		"RELOAD_RECEIPT_MISSING",
		"DRIFT_CHECK_MISSING",
		"ROLLBACK_CONTRACT_MISSING",
		"DATA_ARRIVAL_RECEIPT_MISSING",
	}
	blockers = append(blockers, def.blockers...)
	return uniquePackageRepositoryBlockers(blockers)
}

func credentialPluginBlockers() []string {
	return []string{
		"CREDENTIAL_REF_REQUIRED",
		"CREDENTIAL_REF_GATE_MISSING",
		"URL_ALLOWLIST_GATE_MISSING",
		"TLS_AND_TIMEOUT_GATE_MISSING",
		"TLS_GATE_MISSING",
		"TIMEOUT_GATE_MISSING",
		"ERROR_SANITIZATION_GATE_MISSING",
	}
}

func httpScrapePluginBlockers() []string {
	return []string{
		"CREDENTIAL_REF_GATE_MISSING",
		"URL_ALLOWLIST_GATE_MISSING",
		"TLS_AND_TIMEOUT_GATE_MISSING",
		"TLS_GATE_MISSING",
		"TIMEOUT_GATE_MISSING",
		"ERROR_SANITIZATION_GATE_MISSING",
	}
}

func networkProbePluginBlockers() []string {
	return []string{
		"URL_ALLOWLIST_GATE_MISSING",
		"TLS_AND_TIMEOUT_GATE_MISSING",
		"TLS_GATE_MISSING",
		"TIMEOUT_GATE_MISSING",
		"ERROR_SANITIZATION_GATE_MISSING",
	}
}

func pathScopedPluginBlockers() []string {
	return []string{
		"PATH_ALLOWLIST_GATE_MISSING",
		"PATH_SCOPE_POLICY_MISSING",
		"ERROR_SANITIZATION_GATE_MISSING",
	}
}

func deviceAccessPluginBlockers() []string {
	return []string{
		"DEVICE_ACCESS_POLICY_MISSING",
		"CREDENTIAL_REF_REQUIRED",
		"CREDENTIAL_REF_GATE_MISSING",
		"NETWORK_TARGET_ALLOWLIST_MISSING",
		"ERROR_SANITIZATION_GATE_MISSING",
	}
}

func dedupePluginSourceDefs(defs []pluginSourceCatalogDef) []pluginSourceCatalogDef {
	seen := map[string]bool{}
	out := make([]pluginSourceCatalogDef, 0, len(defs))
	for _, def := range defs {
		if def.id == "" || seen[def.id] {
			continue
		}
		seen[def.id] = true
		out = append(out, def)
	}
	return out
}

func filterPluginSourceDefs(defs []pluginSourceCatalogDef, allowed map[string]bool) []pluginSourceCatalogDef {
	rows := make([]pluginSourceCatalogDef, 0, len(allowed))
	for _, def := range defs {
		if allowed[def.id] {
			rows = append(rows, def)
		}
	}
	return rows
}


func findXAgentPackageDefs() []agentPackageDef {
	defs := baseFindXAgentPackageDefs()
	defs = append(defs, applicationProbePackageDefs()...)
	return append(defs, edgeFindXAgentPackageDefs()...)
}

func baseFindXAgentPackageDefs() []agentPackageDef {
	return []agentPackageDef{
		{
			id: "agent-core", name: "FindX Agent 核心", domain: "基础 Agent", runtime: "Agent Core",
			osList: []string{"Linux", "Windows", "Kubernetes"}, shape: "service / daemon / sidecar",
			telemetryKinds: []string{"heartbeat", "task", "config"}, configKeys: []string{"agent_id", "heartbeat_interval", "task_channel", "global_labels", "credential_ref"},
			configTemplateIDs: []string{"agent-core"},
		},
		{
			id: "host-collector", name: "主机采集能力包", domain: "基础采集", runtime: "Host",
			osList: []string{"Linux", "Windows"}, shape: "collector plugins / service",
			telemetryKinds: []string{"metrics", "process", "host"}, configKeys: []string{"scrape_interval", "plugin_set", "plugin_id", "plugin_version", "config_snippet_ref", "provider_mode", "reload_strategy", "global_labels", "credential_ref"},
			configTemplateIDs: []string{"metrics", "host-plugin"},
			pluginConfig:      pluginConfigSpec("host-plugin", pluginIDForTemplate("host-plugin"), "local-reload"),
		},
		{
			id: "container-collector", name: "容器采集能力包", domain: "基础采集", runtime: "Container",
			osList: []string{"Linux", "Kubernetes"}, shape: "daemonset / sidecar / container plugin",
			telemetryKinds: []string{"metrics", "container", "workload"}, configKeys: []string{"cluster_ref", "namespace_selector", "workload_selector", "plugin_id", "config_snippet_ref", "provider_mode", "restart_strategy", "credential_ref"},
			configTemplateIDs: []string{"metrics", "container-plugin"},
			pluginConfig:      pluginConfigSpec("container-plugin", pluginIDForTemplate("container-plugin"), "rolling-restart"),
		},
		{
			id: "log-collector", name: "日志采集能力包", domain: "日志采集", runtime: "Logs",
			osList: []string{"Linux", "Windows", "Kubernetes"}, shape: "file tailer / stdout collector / parser",
			telemetryKinds: []string{"logs", "pipeline"}, configKeys: []string{"paths", "parser", "pipeline_ref", "labels", "credential_ref"},
			configTemplateIDs: []string{"logs"},
		},
		{
			id: "inspection-runner", name: "巡检诊断能力包", domain: "巡检诊断", runtime: "Inspection",
			osList: []string{"Linux", "Windows", "Kubernetes"}, shape: "check runner / diagnostic plugin",
			telemetryKinds: []string{"inspection", "evidence"}, configKeys: []string{"check_set", "schedule", "risk_level", "evidence_chain", "credential_ref"},
			configTemplateIDs: []string{"inspection", "host-plugin"},
		},
	}
}

func applicationProbePackageDefs() []agentPackageDef {
	return []agentPackageDef{
		applicationProbe("java-app", "Java 应用能力包", "Java"),
		applicationProbe("python-app", "Python 应用能力包", "Python"),
		applicationProbe("nodejs-app", "Node.js 应用能力包", "Node.js"),
		applicationProbe("php-app", "PHP 应用能力包", "PHP"),
		applicationProbe("go-app", "Go 应用能力包", "Go"),
		applicationProbe("rust-app", "Rust 应用能力包", "Rust"),
		applicationProbe("ruby-app", "Ruby 应用能力包", "Ruby"),
	}
}

func edgeFindXAgentPackageDefs() []agentPackageDef {
	return []agentPackageDef{
		{
			id: "gateway-probe", name: "网关链路能力包", domain: "网关链路", runtime: "Gateway",
			osList: []string{"Linux", "Kubernetes"}, shape: "gateway plugin / reverse proxy module",
			telemetryKinds: []string{"tracing", "topology", "logs"}, configKeys: []string{"gateway_id", "route_selector", "collector_endpoint", "sampling", "reload_policy"},
			configTemplateIDs: []string{"gateway-plugin", "tracing", "logs"},
			pluginConfig:      pluginConfigSpec("gateway-plugin", pluginIDForTemplate("gateway-plugin"), "reload"),
		},
		{
			id: "browser-client", name: "前端体验能力包", domain: "前端体验", runtime: "Web",
			osList: []string{"Web"}, shape: "JavaScript SDK",
			telemetryKinds: []string{"rum", "tracing", "errors"}, configKeys: []string{"app_id", "domain", "collector_endpoint", "trace_context", "sampling"},
			configTemplateIDs: []string{"browser-probe"},
		},
	}
}

func applicationProbe(id, name, runtime string) agentPackageDef {
	return agentPackageDef{
		id: id, name: name, domain: "应用链路", runtime: runtime,
		osList: []string{"Linux", "Windows", "Kubernetes"}, shape: "runtime package / SDK / sidecar / init container",
		telemetryKinds:    []string{"tracing", "profiling", "topology"},
		configKeys:        []string{"service_name", "collector_endpoint", "instance_name", "environment", "sampling"},
		configTemplateIDs: []string{"tracing", "profiling"},
	}
}

func sourceHint(parts ...string) string {
	return strings.Join(parts, "")
}

func findXAgentConfigTemplateDefs() []agentConfigTemplateDef {
	defs := coreFindXAgentConfigTemplateDefs()
	defs = append(defs, diagnosticFindXAgentConfigTemplateDefs()...)
	defs = append(defs, pluginFindXAgentConfigTemplateDefs()...)
	return append(defs, applicationProbeConfigTemplateDefs()...)
}

func coreFindXAgentConfigTemplateDefs() []agentConfigTemplateDef {
	return []agentConfigTemplateDef{
		{
			id: "agent-core", name: "Agent 基础配置", kind: "基础配置", scope: "注册 / 心跳 / 任务 / 标签",
			fields:             []string{"agent_id", "heartbeat_interval", "global_labels", "task_channel", "credential_ref"},
			targetScopes:       []string{"全部 Agent", "业务组", "主机", "能力包"},
			rolloutScopes:      []string{"注册参数", "心跳周期", "任务通道", "全局标签"},
			rollbackPolicy:     "按配置版本回滚到上一套稳定基础配置",
			capabilityPackages: []string{"FindX Agent 核心"},
		},
		{
			id: "metrics", name: "指标采集配置", kind: "采集配置", scope: "主机 / 容器 / 进程",
			fields:             []string{"scrape_interval", "labels", "resource_group", "credential_ref"},
			targetScopes:       []string{"全部 Agent", "业务组", "主机", "能力包"},
			rolloutScopes:      []string{"主机指标", "容器指标", "进程指标"},
			rollbackPolicy:     "按采集配置版本回滚到上一套稳定版本",
			capabilityPackages: []string{"主机采集能力包", "容器采集能力包"},
		},
		{
			id: "logs", name: "日志采集配置", kind: "日志配置", scope: "文件 / 标准输出 / 系统日志",
			fields:             []string{"paths", "parser", "pipeline_ref", "labels"},
			targetScopes:       []string{"全部 Agent", "业务组", "主机", "能力包"},
			rolloutScopes:      []string{"文件日志", "容器标准输出", "系统日志"},
			rollbackPolicy:     "按管道引用和采集路径回滚",
			capabilityPackages: []string{"日志采集能力包"},
		},
		{
			id: "tracing", name: "应用链路配置", kind: "链路配置", scope: "应用 / 服务 / 网关",
			fields:             []string{"service_name", "collector_endpoint", "sampling", "propagation"},
			targetScopes:       []string{"业务组", "主机", "能力包", "服务"},
			rolloutScopes:      []string{"应用链路", "服务调用", "网关入口"},
			rollbackPolicy:     "保留采样率、传播协议和接入端点版本快照",
			capabilityPackages: []string{"应用链路能力包", "网关链路能力包"},
		},
	}
}

func diagnosticFindXAgentConfigTemplateDefs() []agentConfigTemplateDef {
	return []agentConfigTemplateDef{
		{
			id: "profiling", name: "性能分析配置", kind: "性能分析配置", scope: "应用运行时",
			fields:             []string{"target_runtime", "duration", "cpu", "memory"},
			targetScopes:       []string{"业务组", "主机", "能力包"},
			rolloutScopes:      []string{"CPU 分析", "内存分析", "运行时分析"},
			rollbackPolicy:     "到期自动关闭并恢复默认采样策略",
			capabilityPackages: []string{"应用链路能力包"},
		},
		{
			id: "inspection", name: "巡检诊断配置", kind: "巡检配置", scope: "主机 / 服务 / 端口",
			fields:             []string{"check_set", "schedule", "risk_level", "evidence_chain"},
			targetScopes:       []string{"全部 Agent", "业务组", "主机", "能力包"},
			rolloutScopes:      []string{"主机巡检", "服务巡检", "端口巡检"},
			rollbackPolicy:     "回滚到上一套巡检集和调度周期",
			capabilityPackages: []string{"巡检诊断能力包"},
		},
	}
}

func pluginFindXAgentConfigTemplateDefs() []agentConfigTemplateDef {
	return []agentConfigTemplateDef{
		pluginTemplate("host-plugin", "主机插件配置", "Linux / Windows 主机插件", "主机采集能力包", "巡检诊断能力包"),
		pluginTemplate("container-plugin", "容器插件配置", "Kubernetes / Docker 插件", "容器采集能力包"),
		pluginTemplate("gateway-plugin", "网关插件配置", "网关 / 反向代理 / 边缘节点", "网关链路能力包"),
	}
}

func applicationProbeConfigTemplateDefs() []agentConfigTemplateDef {
	return []agentConfigTemplateDef{
		{
			id: "browser-probe", name: "前端体验配置", kind: "前端体验配置", scope: "Web 应用 / 页面 / 域名",
			fields:             []string{"app_id", "domain", "collector_endpoint", "trace_context", "sampling"},
			targetScopes:       []string{"业务组", "Web 应用", "域名", "能力包"},
			rolloutScopes:      []string{"前端链路", "页面性能", "错误采集"},
			rollbackPolicy:     "回滚到上一套应用标识、采样率和上下文传播配置",
			capabilityPackages: []string{"前端体验能力包"},
		},
	}
}

func pluginTemplate(id, name, scope string, packages ...string) agentConfigTemplateDef {
	return agentConfigTemplateDef{
		id: id, name: name, kind: "插件配置", scope: scope,
		fields:             []string{"plugin_id", "plugin_version", "config_format", "config_snippet_ref", "provider_mode", "reload_strategy", "remote_mutation", "canary_percent", "rollback_ref", "audit_reason", "change_ticket", "credential_ref"},
		targetScopes:       []string{"全部 Agent", "业务组", "主机", "能力包"},
		rolloutScopes:      []string{"插件配置", "HTTP Provider 配置", "远程修改"},
		rollbackPolicy:     "按插件配置版本和 TOML 片段引用回滚，并保留远程修改审计",
		capabilityPackages: packages,
		pluginConfig:       pluginConfigSpec(id, pluginIDForTemplate(id), reloadStrategyForTemplate(id)),
	}
}

func pluginIDForTemplate(id string) string {
	switch id {
	case "host-plugin":
		return "findx-agent.host-plugin-source-map"
	case "container-plugin":
		return "findx-agent.container-plugin-source-map"
	default:
		return "findx-agent.gateway-plugin-source-map"
	}
}

func reloadStrategyForTemplate(id string) string {
	if id == "container-plugin" {
		return "rolling-restart"
	}
	if id == "gateway-plugin" {
		return "reload"
	}
	return "local-reload"
}
