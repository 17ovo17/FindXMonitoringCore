package handler

import "ai-workbench-api/internal/model"

func expandedFindXAgentCollectPlugins() []model.FindXAgentPlugin {
	allOS := []string{"linux", "windows", "darwin"}
	linuxOnly := []string{"linux"}
	linuxKubernetes := []string{"linux", "kubernetes"}

	return []model.FindXAgentPlugin{
		{ID: "aliyun", Name: "Cloud Resource Metrics", Category: "collect", Description: "Cloud account resource and billing-side metric collection via credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "amd_rocm_smi", Name: "AMD GPU Metrics", Category: "collect", Description: "AMD GPU utilization, memory and temperature metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "apache", Name: "HTTP Service Metrics", Category: "collect", Description: "HTTP service availability, worker and request metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "appdynamics", Name: "APM Bridge Metrics", Category: "collect", Description: "Application metric bridge with credential and endpoint policy gates", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "arp_packet", Name: "ARP Packet Metrics", Category: "collect", Description: "ARP packet and neighbor discovery metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "bind", Name: "DNS Service Metrics", Category: "collect", Description: "DNS server query, cache and response metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "cadvisor", Name: "Container Runtime Metrics", Category: "collect", Description: "Container CPU, memory, filesystem and network metrics", ConfigFormat: "toml", SupportedOS: linuxKubernetes, Enabled: false},
		{ID: "chrony", Name: "Time Sync Metrics", Category: "collect", Description: "Time synchronization offset and tracking metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "cloudwatch", Name: "Cloud Monitoring Metrics", Category: "collect", Description: "Cloud monitoring API collection through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "consul", Name: "Service Registry Metrics", Category: "collect", Description: "Service registry health, catalog and runtime metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "dcgm", Name: "GPU Fabric Metrics", Category: "collect", Description: "GPU fabric, memory and accelerator health metrics", ConfigFormat: "toml", SupportedOS: linuxKubernetes, Enabled: false},
		{ID: "dns_query", Name: "DNS Query Probe", Category: "collect", Description: "DNS query latency and response validation with target allowlist", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "emc_unity", Name: "Storage Array Metrics", Category: "collect", Description: "Storage array pool, disk and service metrics through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "ethtool", Name: "NIC Driver Metrics", Category: "collect", Description: "Network adapter driver, queue and link metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "filecount", Name: "File Count Metrics", Category: "collect", Description: "File count and path inventory metrics with path allowlist gates", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "gnmi", Name: "Network Telemetry Metrics", Category: "collect", Description: "Network telemetry collection through target allowlist and credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "googlecloud", Name: "Cloud Platform Metrics", Category: "collect", Description: "Cloud platform metric collection through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "greenplum", Name: "Database Cluster Metrics", Category: "collect", Description: "MPP database cluster, query and storage metrics through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "hadoop", Name: "Data Platform Metrics", Category: "collect", Description: "Distributed data platform runtime and service metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "haproxy", Name: "Load Balancer Metrics", Category: "collect", Description: "Load balancer frontend, backend and session metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "http_response", Name: "HTTP Response Probe", Category: "collect", Description: "HTTP response code, latency and content probe with allowlist gates", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "ipmi", Name: "Out-of-band Device Metrics", Category: "collect", Description: "Hardware management metrics through credential and device policy gates", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "iptables", Name: "Firewall Rule Metrics", Category: "collect", Description: "Firewall chain and rule counters", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "ipvs", Name: "Virtual Server Metrics", Category: "collect", Description: "Virtual server connection and scheduler metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "jenkins", Name: "Build Service Metrics", Category: "collect", Description: "Build service job, queue and node metrics through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "jolokia_agent_kafka", Name: "JMX Kafka Metrics", Category: "collect", Description: "JMX bridge metrics for messaging workloads", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "jolokia_agent_misc", Name: "JMX Service Metrics", Category: "collect", Description: "JMX bridge metrics for Java services", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "keepalived", Name: "HA Service Metrics", Category: "collect", Description: "High availability service state and VIP metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "kernel", Name: "Kernel Metrics", Category: "collect", Description: "Kernel counters and system runtime metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "kernel_vmstat", Name: "Kernel VM Metrics", Category: "collect", Description: "Virtual memory kernel counters", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "kubernetes", Name: "Workload Metrics", Category: "collect", Description: "Cluster workload and node metrics through scoped selectors", ConfigFormat: "toml", SupportedOS: linuxKubernetes, Enabled: false},
		{ID: "ldap", Name: "Directory Service Metrics", Category: "collect", Description: "Directory service availability and operation metrics through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "linux_sysctl_fs", Name: "Filesystem Kernel Metrics", Category: "collect", Description: "Filesystem kernel parameter metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "logstash", Name: "Pipeline Service Metrics", Category: "collect", Description: "Data pipeline node, queue and event metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "mtail", Name: "Log-derived Metrics", Category: "collect", Description: "Log-derived metric collection with path allowlist gates", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "nats", Name: "Messaging Service Metrics", Category: "collect", Description: "Messaging server connection, route and queue metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "net_response", Name: "Network Response Probe", Category: "collect", Description: "TCP/UDP response probe with target allowlist", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "netstat", Name: "Network Socket Metrics", Category: "collect", Description: "Network socket and protocol counters", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "netstat_filter", Name: "Filtered Network Metrics", Category: "collect", Description: "Filtered network socket metrics with selector gates", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "nfsclient", Name: "NFS Client Metrics", Category: "collect", Description: "NFS client operation and latency metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "nginx_upstream_check", Name: "HTTP Upstream Metrics", Category: "collect", Description: "HTTP upstream health and availability metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "node_exporter", Name: "Node Exporter Scrape", Category: "collect", Description: "Node metrics scrape through network policy gates", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "nsq", Name: "Messaging Queue Metrics", Category: "collect", Description: "Messaging queue topic, channel and node metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "nvidia_smi", Name: "NVIDIA GPU Metrics", Category: "collect", Description: "GPU utilization, memory and temperature metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "phpfpm", Name: "PHP Runtime Metrics", Category: "collect", Description: "PHP process manager pool, worker and request metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "prometheus", Name: "Metrics Scrape Probe", Category: "collect", Description: "Metric endpoint scrape with target allowlist, timeout and TLS policy gates", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "redfish", Name: "Hardware API Metrics", Category: "collect", Description: "Hardware management API metrics through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "rocketmq_offset", Name: "Queue Offset Metrics", Category: "collect", Description: "Messaging offset and lag metrics through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "self_metrics", Name: "Agent Self Metrics", Category: "collect", Description: "Agent runtime self metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "smart", Name: "Disk Health Metrics", Category: "collect", Description: "Disk health and device metrics with device access policy gates", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "snmp", Name: "Network Device Metrics", Category: "collect", Description: "Network device metrics through credential and target allowlist gates", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "snmp_trap", Name: "Network Trap Metrics", Category: "collect", Description: "Network trap receiver metrics with target policy gates", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "snmp_zabbix", Name: "Network Template Metrics", Category: "collect", Description: "Network device template metrics through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "sqlserver", Name: "SQL Database Metrics", Category: "collect", Description: "SQL database connection, query and storage metrics through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "supervisor", Name: "Process Supervisor Metrics", Category: "collect", Description: "Process supervisor state and program metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "system", Name: "System Metrics", Category: "collect", Description: "System runtime, load and uptime metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "tengine", Name: "HTTP Runtime Metrics", Category: "collect", Description: "HTTP runtime status and request metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "tomcat", Name: "Java Web Runtime Metrics", Category: "collect", Description: "Java web runtime thread, request and JVM metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "traffic_server", Name: "HTTP Cache Metrics", Category: "collect", Description: "HTTP cache and proxy metrics", ConfigFormat: "toml", SupportedOS: linuxOnly, Enabled: false},
		{ID: "vsphere", Name: "Virtualization Metrics", Category: "collect", Description: "Virtualization platform metrics through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "whois", Name: "Domain Registry Probe", Category: "collect", Description: "Domain registry probe with target allowlist gates", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "x509_cert", Name: "Certificate Probe", Category: "collect", Description: "Certificate expiry and TLS metadata probe with target allowlist", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "xskyapi", Name: "Storage API Metrics", Category: "collect", Description: "Storage API metrics through credential_ref", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
		{ID: "zookeeper", Name: "Coordination Service Metrics", Category: "collect", Description: "Coordination service health, request and connection metrics", ConfigFormat: "toml", SupportedOS: allOS, Enabled: false},
	}
}

func findXPluginExpandedCredentialRequired(pluginID string) bool {
	switch pluginID {
	case "aliyun", "appdynamics", "cloudwatch", "emc_unity", "gnmi", "googlecloud",
		"greenplum", "ipmi", "jenkins", "ldap", "redfish", "rocketmq_offset",
		"snmp", "snmp_trap", "snmp_zabbix", "sqlserver", "vsphere", "xskyapi":
		return true
	default:
		return false
	}
}

func findXPluginNetworkPolicyRequired(pluginID string) bool {
	switch pluginID {
	case "apache", "appdynamics", "bind", "cadvisor", "cloudwatch", "consul",
		"dns_query", "elasticsearch", "emc_unity", "gnmi", "hadoop", "haproxy",
		"http", "http_response", "jolokia_agent_kafka",
		"jolokia_agent_misc", "kafka", "kubernetes", "logstash", "nats",
		"net_response", "nginx", "nginx_upstream_check", "node_exporter", "nsq",
		"ntp", "phpfpm", "ping", "prometheus", "rabbitmq", "redfish", "redis",
		"redis_sentinel", "snmp", "snmp_trap", "snmp_zabbix", "tengine", "tomcat",
		"traffic_server", "vsphere", "whois", "x509_cert", "zookeeper":
		return true
	default:
		return false
	}
}

func findXPluginPathPolicyRequired(pluginID string) bool {
	switch pluginID {
	case "filecheck", "filecount", "logfile", "mtail", "journaltail":
		return true
	default:
		return false
	}
}

func findXPluginExpandedDashboardRefs(pluginID string) []string {
	switch pluginID {
	case "aliyun", "cloudwatch", "googlecloud":
		return []string{"dashboard:cloud-resource-overview"}
	case "apache", "bind", "consul", "hadoop", "haproxy", "jenkins", "logstash", "nats", "nginx_upstream_check", "nsq", "phpfpm", "supervisor", "tengine", "tomcat", "traffic_server", "zookeeper":
		return []string{"dashboard:service-overview"}
	case "cadvisor", "kubernetes", "node_exporter":
		return []string{"dashboard:workload-overview"}
	case "amd_rocm_smi", "dcgm", "nvidia_smi":
		return []string{"dashboard:accelerator-overview"}
	case "appdynamics", "jolokia_agent_kafka", "jolokia_agent_misc":
		return []string{"dashboard:application-runtime-overview"}
	case "arp_packet", "conntrack", "dns_query", "ethtool", "gnmi", "iptables", "ipvs", "net_response", "netstat", "netstat_filter", "snmp", "snmp_trap", "snmp_zabbix":
		return []string{"dashboard:network-overview"}
	case "chrony", "http_response", "ntp", "whois", "x509_cert":
		return []string{"dashboard:availability-overview"}
	case "emc_unity", "redfish", "smart", "xskyapi":
		return []string{"dashboard:device-health-overview"}
	case "filecount", "mtail", "nfsclient":
		return []string{"dashboard:file-overview"}
	case "greenplum", "influxdb", "sqlserver":
		return []string{"dashboard:database-overview"}
	case "kernel", "kernel_vmstat", "linux_sysctl_fs", "self_metrics", "system":
		return []string{"dashboard:host-basic"}
	case "rocketmq_offset":
		return []string{"dashboard:queue-overview"}
	case "vsphere":
		return []string{"dashboard:virtualization-overview"}
	default:
		return nil
	}
}
