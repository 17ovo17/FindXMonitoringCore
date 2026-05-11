package handler

import (
	"fmt"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

type businessMetricSpec struct {
	Name, Metric, Unit string
	Warn, Crit         float64
}

func countBusinessAlerts(alerts []*model.AlertRecord, status string) int {
	count := 0
	for _, alert := range alerts {
		if alert.Status == status {
			count++
		}
	}
	return count
}

func limitBusinessMetrics(items []model.BusinessMetricSample, limit int) []model.BusinessMetricSample {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func businessInspectionSuggestions(business model.TopologyBusiness, metrics []model.BusinessMetricSample, processes []model.BusinessProcess, alerts []*model.AlertRecord) []string {
	suggestions := []string{businessLayerSuggestion(business.Endpoints)}
	suggestions = appendMetricSuggestions(suggestions, metrics)
	suggestions = appendProcessSuggestions(suggestions, processes)
	suggestions = appendAlertSuggestions(suggestions, alerts)
	if hasRedisEndpoint(business.Endpoints) {
		suggestions = append(suggestions, "AI suggestion: Redis is registered as middleware; inspect connections, memory, QPS, hit rate, and application-to-Redis path health.")
	}
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "AI suggestion: no blocking risk was found; keep observing SLO, alerts, core process health, and database connection pools.")
	}
	return suggestions
}

func businessLayerSuggestion(endpoints []model.TopologyEndpoint) string {
	layers := map[string]bool{}
	for _, endpoint := range endpoints {
		layers[classifyEndpointRole(endpoint)] = true
	}
	if layers["frontend"] && layers["app"] && layers["middleware"] && layers["database"] {
		return "AI suggestion: all four business layers are registered; validate latency, errors, and port reachability along entry, app, middleware, and database layers."
	}
	return "AI suggestion: business topology is incomplete; register entry, app, middleware, and database hosts and ports before deep inspection."
}

func appendMetricSuggestions(suggestions []string, metrics []model.BusinessMetricSample) []string {
	criticalMetrics, warningMetrics := businessMetricStatusText(metrics)
	if len(criticalMetrics) > 0 {
		return append(suggestions, "AI suggestion: handle critical resource anomalies first: "+strings.Join(limitStrings(criticalMetrics, 4), "; "))
	}
	if len(warningMetrics) > 0 {
		return append(suggestions, "AI suggestion: resource warning trends exist; compare CPU, memory, disk IO, network, JVM, Redis, and database metrics at the same time point.")
	}
	return suggestions
}

func businessMetricStatusText(metrics []model.BusinessMetricSample) ([]string, []string) {
	criticalMetrics := []string{}
	warningMetrics := []string{}
	for _, metric := range metrics {
		text := fmt.Sprintf("%s/%s=%.2f%s", metric.IP, metric.Name, metric.Value, metric.Unit)
		if metric.Status == "critical" {
			criticalMetrics = append(criticalMetrics, text)
		} else if metric.Status == "warning" {
			warningMetrics = append(warningMetrics, text)
		}
	}
	return criticalMetrics, warningMetrics
}

func appendProcessSuggestions(suggestions []string, processes []model.BusinessProcess) []string {
	badProcesses := []string{}
	for _, process := range processes {
		if process.Status != "running" {
			badProcesses = append(badProcesses, fmt.Sprintf("%s %s:%d", process.Name, process.IP, process.Port))
		}
	}
	if len(badProcesses) == 0 {
		return suggestions
	}
	return append(suggestions, "AI suggestion: verify business processes and listening ports: "+strings.Join(limitStrings(badProcesses, 4), "; "))
}

func appendAlertSuggestions(suggestions []string, alerts []*model.AlertRecord) []string {
	firing := 0
	for _, alert := range alerts {
		if alert.Status == "firing" {
			firing++
		}
	}
	if firing == 0 {
		return suggestions
	}
	return append(suggestions, fmt.Sprintf("AI suggestion: %d unresolved alerts exist; map them to impacted business nodes before node-level diagnosis.", firing))
}

func limitStrings(items []string, limit int) []string {
	if len(items) <= limit {
		return items
	}
	return append(items[:limit], fmt.Sprintf("More items: %d", len(items)-limit))
}

func businessAlerts(hosts []string) []*model.AlertRecord {
	hostSet := map[string]bool{}
	for _, host := range hosts {
		hostSet[host] = true
	}
	out := []*model.AlertRecord{}
	for _, alert := range store.ListAlerts() {
		if hostSet[alert.TargetIP] {
			out = append(out, alert)
		}
	}
	return out
}

func businessMetricSamples(hosts []string, endpoints []model.TopologyEndpoint) []model.BusinessMetricSample {
	samples := []model.BusinessMetricSample{}
	hostRoles := businessHostRoles(endpoints)
	for _, host := range hosts {
		samples = append(samples, businessHostMetricSamples(host, hostRoles[host])...)
	}
	samples = append(samples, businessEndpointMetricSamples(endpoints)...)
	return samples
}

func businessHostRoles(endpoints []model.TopologyEndpoint) map[string]string {
	hostRoles := map[string]string{}
	for _, ep := range endpoints {
		if role := classifyEndpointRole(ep); role != "" {
			hostRoles[ep.IP] = role
		}
	}
	return hostRoles
}

func businessHostMetricSamples(host, role string) []model.BusinessMetricSample {
	target := discoverPromTarget(host)
	if target.LabelKey == "" {
		return []model.BusinessMetricSample{{IP: host, Name: "Prometheus metrics", Status: "unknown", Source: "prometheus", Detail: "No Prometheus label or historical metric was found for this IP."}}
	}
	samples := []model.BusinessMetricSample{}
	selector := fmt.Sprintf(`%s="%s"`, target.LabelKey, target.LabelVal)
	metricSet := stringSet(target.Metrics)
	for _, spec := range businessHostMetricSpecs(role) {
		if !metricSet[spec.Metric] {
			continue
		}
		if sample, ok := queryBusinessMetricSample(host, selector, spec); ok {
			samples = append(samples, sample)
		}
	}
	return samples
}

func businessHostMetricSpecs(role string) []businessMetricSpec {
	cpuWarn, cpuCrit := 75.0, 90.0
	memWarn, memCrit := 80.0, 92.0
	switch role {
	case "database":
		cpuWarn, cpuCrit, memWarn, memCrit = 80, 95, 90, 97
	case "frontend":
		cpuWarn, cpuCrit, memWarn, memCrit = 65, 85, 75, 90
	case "middleware":
		cpuWarn, cpuCrit, memWarn, memCrit = 50, 80, 80, 95
	}
	return []businessMetricSpec{
		{"CPU usage", "cpu_usage_active", "%", cpuWarn, cpuCrit},
		{"Memory usage", "mem_used_percent", "%", memWarn, memCrit},
		{"Disk usage", "disk_used_percent", "%", 80, 90},
		{"System load 1m", "system_load1", "", 8, 16},
		{"TCP established", "netstat_tcp_established", "", 3000, 8000},
	}
}

func queryBusinessMetricSample(host, selector string, spec businessMetricSpec) (model.BusinessMetricSample, bool) {
	query := fmt.Sprintf(`%s{%s}`, spec.Metric, selector)
	text, err := queryProm(query)
	if err != nil || strings.TrimSpace(text) == "" || strings.Contains(text, "no data") {
		return model.BusinessMetricSample{}, false
	}
	value := parseFloatValue(text)
	status := "healthy"
	if value >= spec.Crit {
		status = "critical"
	} else if value >= spec.Warn {
		status = "warning"
	}
	sample := model.BusinessMetricSample{IP: host, Name: spec.Name, Value: value, Unit: spec.Unit, Status: status, Source: "prometheus", Query: query, Detail: text}
	return sample, true
}

func stringSet(items []string) map[string]bool {
	out := map[string]bool{}
	for _, item := range items {
		out[item] = true
	}
	return out
}

func businessEndpointMetricSamples(endpoints []model.TopologyEndpoint) []model.BusinessMetricSample {
	samples := []model.BusinessMetricSample{}
	for _, endpoint := range classifyEndpointsWithAI(endpoints, false) {
		if sample, ok := businessEndpointMetricSample(endpoint); ok {
			samples = append(samples, sample...)
		}
	}
	return samples
}

func businessEndpointMetricSample(endpoint model.TopologyEndpoint) ([]model.BusinessMetricSample, bool) {
	role := classifyEndpointRole(endpoint)
	if role != "middleware" && role != "database" && role != "frontend" && role != "app" {
		return nil, false
	}
	target := discoverPromTarget(endpoint.IP)
	if target.LabelKey == "" {
		return nil, false
	}
	return queryEndpointPriorityMetrics(endpoint, target), true
}

func queryEndpointPriorityMetrics(endpoint model.TopologyEndpoint, target promTarget) []model.BusinessMetricSample {
	selector := fmt.Sprintf(`%s="%s"`, target.LabelKey, target.LabelVal)
	metricNames := endpointPriorityMetrics(endpoint, target.Metrics)
	if len(metricNames) == 0 {
		detail := "Prometheus has no dedicated metric for this endpoint; endpoint connectivity and process inspection are still retained"
		return []model.BusinessMetricSample{{IP: endpoint.IP, Name: endpoint.ServiceName + " metrics", Status: "unknown", Source: "prometheus", Query: selector, Detail: detail}}
	}
	samples := []model.BusinessMetricSample{}
	for _, metricName := range metricNames {
		if sample, ok := queryEndpointMetricSample(endpoint, selector, metricName); ok {
			samples = append(samples, sample)
		}
	}
	return samples
}

func queryEndpointMetricSample(endpoint model.TopologyEndpoint, selector, metricName string) (model.BusinessMetricSample, bool) {
	query := fmt.Sprintf(`%s{%s}`, metricName, selector)
	text, err := queryProm(query)
	if err != nil || strings.TrimSpace(text) == "" || strings.Contains(text, "no data") {
		return model.BusinessMetricSample{}, false
	}
	sample := model.BusinessMetricSample{IP: endpoint.IP, Name: endpoint.ServiceName + " / " + metricName, Value: parseFloatValue(text), Status: "healthy", Source: "prometheus", Query: query, Detail: text}
	return sample, true
}

func endpointPriorityMetrics(endpoint model.TopologyEndpoint, metrics []string) []string {
	name := strings.ToLower(endpoint.ServiceName)
	want := []string{}
	switch {
	case strings.Contains(name, "redis") || endpoint.Port == 6379 || endpoint.Port == 6375:
		want = []string{"redis_connected_clients", "redis_used_memory", "redis_mem_used", "redis_instantaneous_ops_per_sec", "redis_keyspace_hits", "redis_keyspace_misses", "redis_uptime_in_seconds"}
	case strings.Contains(name, "oracle") || endpoint.Port == 1521:
		want = []string{"oracle_up", "oracle_sessions", "oracle_tablespace_used_percent", "oracle_process_count"}
	case strings.Contains(name, "nginx") || endpoint.Port == 80 || endpoint.Port == 443:
		want = []string{"nginx_connections_active", "nginx_requests_total", "nginx_up"}
	case strings.Contains(name, "jvm") || strings.Contains(name, "app") || endpoint.Port == 8080 || endpoint.Port == 8081:
		want = []string{"jvm_memory_used_bytes", "jvm_threads_live_threads", "jvm_gc_pause_seconds_count", "process_cpu_seconds_total"}
	}
	metricSet := stringSet(metrics)
	out := []string{}
	for _, metric := range want {
		if metricSet[metric] {
			out = append(out, metric)
		}
	}
	return out
}

func businessProcesses(business model.TopologyBusiness) []model.BusinessProcess {
	processes := []model.BusinessProcess{}
	for _, endpoint := range classifyEndpointsWithAI(business.Endpoints, false) {
		processes = append(processes, businessProcess(endpoint))
	}
	return processes
}

func businessProcess(endpoint model.TopologyEndpoint) model.BusinessProcess {
	name := strings.TrimSpace(endpoint.ServiceName)
	if name == "" {
		name = fmt.Sprintf("port-%d", endpoint.Port)
	}
	status, alert := businessProcessStatus(endpoint.IP)
	return model.BusinessProcess{IP: endpoint.IP, Name: name, Description: processDescription(endpoint), Path: processPath(endpoint), Port: endpoint.Port, Layer: classifyEndpointRole(endpoint), Status: status, Alert: alert}
}

func businessProcessStatus(ip string) (string, string) {
	if hasRecentPrometheusData(ip) || store.HasOnlineAgent(ip) {
		return "running", ""
	}
	return "unknown", "No monitoring data was detected for this host; check FindX Agent collection status."
}

func processDescription(endpoint model.TopologyEndpoint) string {
	switch classifyEndpointRole(endpoint) {
	case "frontend":
		return "Business entry or reverse proxy process"
	case "app":
		return "Business application process or JVM service"
	case "middleware":
		return "Middleware process"
	case "database":
		return "Database listener or instance process"
	default:
		return "User-defined business port"
	}
}

func processPath(endpoint model.TopologyEndpoint) string {
	name := strings.ToLower(endpoint.ServiceName)
	switch {
	case strings.Contains(name, "nginx"):
		return "/usr/sbin/nginx"
	case strings.Contains(name, "redis"):
		return "/usr/bin/redis-server"
	case strings.Contains(name, "oracle"):
		return "$ORACLE_HOME/bin/tnslsnr"
	case strings.Contains(name, "jvm") || strings.Contains(name, "app"):
		return "java -jar /opt/app/*.jar"
	default:
		return "Complete with Catpaw process inspection"
	}
}

func businessResources(business model.TopologyBusiness) []model.BusinessResource {
	resources := []model.BusinessResource{}
	owner := business.Attributes["owner"]
	purpose := business.Attributes["purpose"]
	for _, host := range business.Hosts {
		resources = append(resources, businessHostResource(host, owner, purpose))
	}
	for _, endpoint := range business.Endpoints {
		resources = append(resources, businessEndpointResource(endpoint, owner, purpose))
	}
	return resources
}

func businessHostResource(host, owner, purpose string) model.BusinessResource {
	status := "offline"
	for _, agent := range store.ListAgents() {
		if agent.IP == host && agent.Online {
			status = "online"
		}
	}
	return model.BusinessResource{IP: host, Name: host, Type: "host", Owner: owner, Purpose: purpose, Status: status, Attrs: map[string]string{"source": "user_scope+prometheus+catpaw"}}
}

func businessEndpointResource(endpoint model.TopologyEndpoint, owner, purpose string) model.BusinessResource {
	epStatus := "online"
	if !hasRecentPrometheusData(endpoint.IP) && !store.HasOnlineAgent(endpoint.IP) {
		epStatus = "unknown"
	}
	attrs := map[string]string{"port": fmt.Sprintf("%d", endpoint.Port), "protocol": semanticEndpointProtocol(endpoint)}
	return model.BusinessResource{IP: endpoint.IP, Name: endpoint.ServiceName, Type: classifyEndpointRole(endpoint), Owner: owner, Purpose: purpose, Status: epStatus, Attrs: attrs}
}

func topologyFindings(business model.TopologyBusiness) []string {
	findings := []string{}
	roles := map[string]int{}
	for _, endpoint := range business.Endpoints {
		roles[classifyEndpointRole(endpoint)]++
	}
	if roles["frontend"] == 0 {
		findings = append(findings, "No entry layer was identified; add gateway, proxy, or load balancer ports.")
	}
	if roles["app"] == 0 {
		findings = append(findings, "No application layer was identified; add JVM, app, or Tomcat ports.")
	}
	if roles["middleware"] == 0 {
		findings = append(findings, "No middleware layer was identified; add Redis, MQ, or similar endpoints when applicable.")
	}
	if roles["database"] == 0 {
		findings = append(findings, "No database layer was identified; add MySQL, PostgreSQL, Oracle, or MongoDB endpoints when applicable.")
	}
	if len(findings) == 0 {
		findings = append(findings, "Business topology is complete: entry, app, middleware, and database layers were identified.")
	}
	return findings
}
