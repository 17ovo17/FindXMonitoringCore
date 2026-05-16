package handler

import "ai-workbench-api/internal/model"

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
