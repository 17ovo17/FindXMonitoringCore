package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type topologyDiscoverRequest struct {
	Hosts           []string                 `json:"hosts"`
	Endpoints       []model.TopologyEndpoint `json:"endpoints"`
	IncludePlatform bool                     `json:"include_platform"`
	UseAI           bool                     `json:"use_ai"`
	Attributes      map[string]string        `json:"attributes"`
}

type topologyServiceCandidate struct {
	Name     string
	Type     string
	Port     int
	Protocol string
	Required bool
}

func DiscoverTopology(c *gin.Context) {
	var req topologyDiscoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hosts := normalizeHosts(req.Hosts)
	hosts = mergeHosts(hosts, prometheusHostsForSelection(hosts))
	if !req.IncludePlatform && len(hosts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hosts or include_platform is required"})
		return
	}
	now := time.Now()
	graph := model.TopologyGraph{Nodes: []model.TopologyNode{}, Edges: []model.TopologyEdge{}}
	if req.IncludePlatform {
		addPlatformDiscovery(&graph, now)
	}
	declaredPorts := endpointPortSet(req.Endpoints)
	for index, host := range hosts {
		addHostDiscovery(&graph, host, index, now, declaredPorts)
	}
	classified := classifyEndpointsWithAI(req.Endpoints, req.UseAI)
	addUserDefinedEndpoints(&graph, classified, now)
	addInferredBusinessEdges(&graph, classified, now)
	graph.Discovery = buildTopologyDiscoveryPlan(hosts, classified, &graph, req.UseAI)
	layoutBusinessTree(&graph)
	c.JSON(http.StatusOK, graph)
}

func normalizeHosts(hosts []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		host, _ = url.QueryUnescape(host)
		host = strings.TrimPrefix(strings.TrimPrefix(host, "http://"), "https://")
		if i := strings.Index(host, "/"); i >= 0 {
			host = host[:i]
		}
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if host == "" || seen[host] {
			continue
		}
		seen[host] = true
		out = append(out, host)
	}
	sort.Strings(out)
	return out
}

func platformResources() []gin.H {
	ip := PlatformIP()
	return []gin.H{
		{"id": "platform-web", "name": "AI WorkBench \u524d\u7aef", "type": "frontend", "ip": ip, "port": 3000},
		{"id": "platform-api", "name": "AI WorkBench \u540e\u7aef", "type": "backend", "ip": ip, "port": 8080},
		{"id": "platform-mysql", "name": "MySQL", "type": "database", "ip": "127.0.0.1", "port": 3306},
		{"id": "platform-redis", "name": "Redis", "type": "cache", "ip": "127.0.0.1", "port": 6379},
		{"id": "platform-prom", "name": "Prometheus", "type": "monitor", "ip": ip, "port": 9090},
	}
}

func addPlatformDiscovery(graph *model.TopologyGraph, now time.Time) {
	baseX := 90.0
	baseY := 110.0
	components := []model.TopologyNode{
		{ID: "platform-web", Name: "AI WorkBench \u524d\u7aef", Type: "frontend", IP: "127.0.0.1", ServiceName: "web", Port: 3000, X: baseX, Y: baseY, CreatedAt: now, UpdatedAt: now},
		{ID: "platform-api", Name: "AI WorkBench \u540e\u7aef", Type: "backend", IP: "127.0.0.1", ServiceName: "api", Port: 8080, X: baseX + 250, Y: baseY, CreatedAt: now, UpdatedAt: now},
		{ID: "platform-mysql", Name: "MySQL", Type: "database", IP: "127.0.0.1", ServiceName: "mysql", Port: 3306, X: baseX + 520, Y: baseY - 55, CreatedAt: now, UpdatedAt: now},
		{ID: "platform-redis", Name: "Redis", Type: "cache", IP: "127.0.0.1", ServiceName: "redis", Port: 6379, X: baseX + 520, Y: baseY + 80, CreatedAt: now, UpdatedAt: now},
		{ID: "platform-prom", Name: "Prometheus", Type: "monitor", IP: "127.0.0.1", ServiceName: "prometheus", Port: 9090, X: baseX + 250, Y: baseY + 190, CreatedAt: now, UpdatedAt: now},
	}
	for i := range components {
		components[i].Status = statusFromDial("127.0.0.1", components[i].Port)
		graph.Nodes = append(graph.Nodes, components[i])
	}
	addCheckedEdge(graph, "edge-web-api", "platform-web", "platform-api", "HTTP", "API \u8c03\u7528", "127.0.0.1", 8080, now)
	addCheckedEdge(graph, "edge-api-mysql", "platform-api", "platform-mysql", "MySQL", "\u6301\u4e45\u5316", "127.0.0.1", 3306, now)
	addCheckedEdge(graph, "edge-api-redis", "platform-api", "platform-redis", "Redis", "\u7f13\u5b58/\u5728\u7ebf\u72b6\u6001", "127.0.0.1", 6379, now)
	addCheckedEdge(graph, "edge-api-prom", "platform-api", "platform-prom", "HTTP", "\u76d1\u63a7\u67e5\u8be2", "127.0.0.1", 9090, now)
}

type prometheusTargetInfo struct {
	Address string
	IP      string
	Port    int
	Health  string
	Error   string
}

func prometheusHostsForSelection(hosts []string) []string {
	if len(hosts) == 0 {
		return nil
	}
	out := []string{}
	for _, target := range listPrometheusTargets() {
		for _, host := range hosts {
			if target.IP == host {
				out = append(out, target.IP)
			}
		}
	}
	return out
}

func listPrometheusTargets() []prometheusTargetInfo {
	base := strings.TrimRight(viper.GetString("prometheus.url"), "/")
	if base == "" {
		return nil
	}
	resp, err := http.Get(base + "/api/v1/targets")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status string `json:"status"`
		Data   struct {
			ActiveTargets []struct {
				DiscoveredLabels map[string]string `json:"discoveredLabels"`
				Labels           map[string]string `json:"labels"`
				Health           string            `json:"health"`
				LastError        string            `json:"lastError"`
			} `json:"activeTargets"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &result) != nil || result.Status != "success" {
		return nil
	}
	out := []prometheusTargetInfo{}
	for _, item := range result.Data.ActiveTargets {
		address := item.DiscoveredLabels["__address__"]
		if address == "" {
			address = item.Labels["instance"]
		}
		if address == "" {
			continue
		}
		host, portText, err := net.SplitHostPort(address)
		if err != nil {
			host = address
		}
		port := 0
		if portText != "" {
			fmt.Sscanf(portText, "%d", &port)
		}
		out = append(out, prometheusTargetInfo{Address: address, IP: host, Port: port, Health: item.Health, Error: item.LastError})
	}
	return out
}

func mergeHosts(hosts []string, extra []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, host := range append(hosts, extra...) {
		host = strings.TrimSpace(host)
		if host == "" || seen[host] {
			continue
		}
		seen[host] = true
		out = append(out, host)
	}
	sort.Strings(out)
	return out
}

func endpointPortSet(endpoints []model.TopologyEndpoint) map[string]map[int]bool {
	out := map[string]map[int]bool{}
	for _, endpoint := range endpoints {
		ip := strings.TrimSpace(endpoint.IP)
		if ip == "" || endpoint.Port <= 0 {
			continue
		}
		if out[ip] == nil {
			out[ip] = map[int]bool{}
		}
		out[ip][endpoint.Port] = true
	}
	return out
}

func addHostDiscovery(graph *model.TopologyGraph, host string, index int, now time.Time, declaredPorts map[string]map[int]bool) {
	hostID := "host-" + sanitizeID(host)
	hostName := host
	status := "offline"
	for _, agent := range store.ListAgents() {
		if agent.IP == host {
			if strings.TrimSpace(agent.Hostname) != "" {
				hostName = agent.Hostname + " (" + host + ")"
			}
			if agent.Online {
				status = "online"
			}
		}
	}
	x := 90.0 + float64(index%2)*420
	y := 430.0 + float64(index/2)*250
	graph.Nodes = append(graph.Nodes, model.TopologyNode{ID: hostID, Name: hostName, Type: "host", IP: host, Status: status, Layer: 0, X: x, Y: y, Meta: "User-defined business host", CreatedAt: now, UpdatedAt: now})
	serviceIndex := 0
	for _, target := range listPrometheusTargets() {
		if target.IP != host || target.Port == 0 {
			continue
		}
		if declaredPorts[host] != nil && declaredPorts[host][target.Port] {
			continue
		}
		serviceID := fmt.Sprintf("prom-target-%s-%d", sanitizeID(host), target.Port)
		status := "online"
		edgeStatus := "connected"
		if target.Health != "up" {
			status = "offline"
			edgeStatus = "disconnected"
		}
		name := prometheusTargetName(host, target.Port)
		graph.Nodes = append(graph.Nodes, model.TopologyNode{ID: serviceID, Name: name, Type: "monitor", IP: host, ServiceName: prometheusTargetServiceName(target.Port), Port: target.Port, Status: status, Layer: 5, X: x + 230 + float64(serviceIndex%2)*230, Y: y - 30 + float64(serviceIndex/2)*85, Meta: "Prometheus target discovery; offline is allowed in test mode", CreatedAt: now, UpdatedAt: now})
		graph.Edges = append(graph.Edges, model.TopologyEdge{ID: "edge-" + hostID + "-" + serviceID, SourceID: hostID, TargetID: serviceID, Protocol: "Metrics", Direction: "forward", Label: "Prometheus target", Status: edgeStatus, Error: target.Error, CreatedAt: now, UpdatedAt: now})
		serviceIndex++
	}
	if hasNode(graph, "platform-api") {
		addCheckedEdge(graph, "edge-"+hostID+"-platform-api", hostID, "platform-api", "HTTP", "Catpaw/\u4e1a\u52a1\u56de\u4f20\u5e73\u53f0", "127.0.0.1", 8080, now)
	}
}

func prometheusTargetName(host string, port int) string {
	switch port {
	case 9090:
		return fmt.Sprintf("%s:%d Prometheus", host, port)
	case 9091:
		return fmt.Sprintf("%s:%d Pushgateway / Observability", host, port)
	case 9100:
		return fmt.Sprintf("%s:%d Node Exporter", host, port)
	case 9101, 9102, 9103:
		return fmt.Sprintf("%s:%d Categraf / Exporter", host, port)
	}
	return fmt.Sprintf("%s:%d Prometheus Target", host, port)
}

func prometheusTargetServiceName(port int) string {
	switch port {
	case 9090:
		return "Prometheus"
	case 9091:
		return "Pushgateway / Observability"
	case 9100:
		return "Node Exporter"
	case 9101, 9102, 9103:
		return "Categraf / Exporter"
	}
	return "Prometheus Target"
}

func addUserDefinedEndpoints(graph *model.TopologyGraph, endpoints []model.TopologyEndpoint, now time.Time) {
	for _, endpoint := range endpoints {
		endpoint.IP = strings.TrimSpace(endpoint.IP)
		if endpoint.IP == "" || endpoint.Port <= 0 {
			continue
		}
		hostID := "host-" + sanitizeID(endpoint.IP)
		if !hasNode(graph, hostID) {
			graph.Nodes = append(graph.Nodes, model.TopologyNode{ID: hostID, Name: endpoint.IP, Type: "host", IP: endpoint.IP, Status: "offline", Layer: 0, Meta: "User-defined business host", CreatedAt: now, UpdatedAt: now})
		}
		serviceName := strings.TrimSpace(endpoint.ServiceName)
		if serviceName == "" {
			serviceName = fmt.Sprintf("Business port %d", endpoint.Port)
		}
		protocol := semanticEndpointProtocol(endpoint)
		serviceID := fmt.Sprintf("biz-%s-%d", sanitizeID(endpoint.IP), endpoint.Port)
		if hasNode(graph, serviceID) {
			continue
		}
		latency, errText := checkTCP(endpoint.IP, endpoint.Port)
		status, edgeStatus := "online", "connected"
		if errText != "" {
			status, edgeStatus = "offline", "disconnected"
		}
		graph.Nodes = append(graph.Nodes, model.TopologyNode{ID: serviceID, Name: fmt.Sprintf("%s:%d %s", endpoint.IP, endpoint.Port, serviceName), Type: classifyEndpointType(endpoint), IP: endpoint.IP, Port: endpoint.Port, ServiceName: serviceName, Status: status, Layer: endpointLayer(endpoint), Meta: "User-defined business endpoint with automatic connectivity check", CreatedAt: now, UpdatedAt: now})
		graph.Edges = append(graph.Edges, model.TopologyEdge{ID: "edge-" + hostID + "-" + serviceID, SourceID: hostID, TargetID: serviceID, Protocol: protocol, Direction: "forward", Label: serviceName, Status: edgeStatus, LatencyMs: latency, Error: errText, CreatedAt: now, UpdatedAt: now})
	}
}

func classifyEndpointsWithAI(endpoints []model.TopologyEndpoint, useAI bool) []model.TopologyEndpoint {
	classified := make([]model.TopologyEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpoint.ServiceName = normalizeEndpointServiceName(endpoint)
		endpoint.Protocol = semanticEndpointProtocol(endpoint)
		classified = append(classified, endpoint)
	}
	if !useAI {
		return classified
	}
	// External AI is an enhancement, not a trust boundary. The backend still applies
	// deterministic classification and never accepts out-of-scope IPs or ports.
	if summary, err := callEndpointClassifierAI(classified); err == nil && summary != "" {
		for i := range classified {
			classified[i].ServiceName = strings.TrimSpace(classified[i].ServiceName)
		}
	}
	return classified
}

func normalizeEndpointServiceName(endpoint model.TopologyEndpoint) string {
	name := strings.TrimSpace(endpoint.ServiceName)
	lower := strings.ToLower(name)
	switch {
	case name == "" && (endpoint.Port == 80 || endpoint.Port == 443):
		return "nginx"
	case name == "" && (endpoint.Port == 8080 || endpoint.Port == 8081):
		return "jvm-app"
	case name == "" && endpoint.Port == 6379:
		return "redis"
	case name == "" && endpoint.Port == 1521:
		return "oracle"
	case strings.Contains(lower, "nginx") || strings.Contains(lower, "gateway"):
		return name
	case strings.Contains(lower, "redis") || strings.Contains(lower, "sentinel"):
		return name
	case strings.Contains(lower, "oracle") || strings.Contains(lower, "mysql") || strings.Contains(lower, "postgres"):
		return name
	case strings.Contains(lower, "jvm") || strings.Contains(lower, "app") || strings.Contains(lower, "tomcat"):
		return name
	case endpoint.Port == 6379:
		return "redis " + name
	case endpoint.Port == 1521:
		return "oracle " + name
	case endpoint.Port == 8080 || endpoint.Port == 8081:
		return "jvm-app " + name
	default:
		return name
	}
}
