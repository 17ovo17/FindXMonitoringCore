package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func callEndpointClassifierAI(endpoints []model.TopologyEndpoint) (string, error) {
	apiKey := strings.TrimSpace(getAPIKey())
	if apiKey == "" || apiKey == "******" || strings.Contains(apiKey, "${") {
		return "", fmt.Errorf("AI provider API key is not configured")
	}
	payload := map[string]any{
		"model": resolveDefaultModel(),
		"messages": []map[string]string{
			{"role": "system", "content": "Classify user-provided endpoints into entry, application, middleware, database, or observability. Return a concise explanation only. Never add endpoints."},
			{"role": "user", "content": fmt.Sprintf("endpoints=%v", endpoints)},
		},
		"stream": false,
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 6 * time.Second}
	req, err := http.NewRequest(http.MethodPost, getBaseURL()+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("AI classifier status %d", resp.StatusCode)
	}
	return "ai-classified", nil
}

func addInferredBusinessEdges(graph *model.TopologyGraph, endpoints []model.TopologyEndpoint, now time.Time) {
	frontends := []string{}
	apps := []string{}
	middleware := []string{}
	databases := []string{}
	for _, endpoint := range endpoints {
		if strings.TrimSpace(endpoint.IP) == "" || endpoint.Port <= 0 {
			continue
		}
		id := fmt.Sprintf("biz-%s-%d", sanitizeID(endpoint.IP), endpoint.Port)
		switch classifyEndpointRole(endpoint) {
		case "frontend":
			frontends = append(frontends, id)
		case "app":
			apps = append(apps, id)
		case "middleware":
			middleware = append(middleware, id)
		case "database":
			databases = append(databases, id)
		}
	}
	for _, frontend := range frontends {
		for _, app := range apps {
			addInferredEdge(graph, frontend, app, "HTTP", "Nginx to application", now)
		}
	}
	for _, app := range apps {
		for _, target := range middleware {
			addInferredEdge(graph, app, target, inferredEdgeProtocol(graph, target), "Application to middleware", now)
		}
		for _, target := range databases {
			addInferredEdge(graph, app, target, inferredEdgeProtocol(graph, target), "Application to database", now)
		}
	}
	if len(frontends) == 0 {
		for _, app := range apps {
			for _, target := range middleware {
				addInferredEdge(graph, app, target, inferredEdgeProtocol(graph, target), "Application to middleware", now)
			}
			for _, target := range databases {
				addInferredEdge(graph, app, target, inferredEdgeProtocol(graph, target), "Application to database", now)
			}
		}
	}
}

func addInferredEdge(graph *model.TopologyGraph, sourceID, targetID, protocol, label string, now time.Time) {
	if sourceID == targetID || !hasNode(graph, sourceID) || !hasNode(graph, targetID) {
		return
	}
	id := "edge-business-" + sourceID + "-" + targetID
	if hasEdge(graph, id) {
		return
	}
	status := "connected"
	errorText := ""
	sourceStatus, targetStatus := nodeStatus(graph, sourceID), nodeStatus(graph, targetID)
	if sourceStatus == "offline" || targetStatus == "offline" {
		status = "disconnected"
		errorText = "One endpoint is unreachable; offline target does not block historical metric discovery"
	}
	graph.Edges = append(graph.Edges, model.TopologyEdge{ID: id, SourceID: sourceID, TargetID: targetID, Protocol: protocol, Direction: "forward", Label: label, Status: status, Error: errorText, CreatedAt: now, UpdatedAt: now})
}

func classifyEndpointType(endpoint model.TopologyEndpoint) string {
	switch classifyEndpointRole(endpoint) {
	case "frontend":
		return "frontend"
	case "app":
		return "application"
	case "middleware":
		return "cache"
	case "database":
		return "database"
	default:
		return "service"
	}
}

func classifyEndpointRole(endpoint model.TopologyEndpoint) string {
	name := strings.ToLower(strings.TrimSpace(endpoint.ServiceName))
	port := endpoint.Port
	if strings.Contains(name, "nginx") || strings.Contains(name, "gateway") || strings.Contains(name, "lb") || port == 80 || port == 443 {
		return "frontend"
	}
	if strings.Contains(name, "jvm") || strings.Contains(name, "app") || strings.Contains(name, "tomcat") || port == 8080 || port == 8081 || port == 8000 {
		return "app"
	}
	if strings.Contains(name, "redis") || strings.Contains(name, "sentinel") || strings.Contains(name, "kafka") || strings.Contains(name, "rabbit") || strings.Contains(name, "mq") || port == 6379 || port == 26379 || port == 5672 || port == 9092 {
		return "middleware"
	}
	if strings.Contains(name, "oracle") || strings.Contains(name, "mysql") || strings.Contains(name, "postgres") || strings.Contains(name, "database") || port == 1521 || port == 3306 || port == 5432 {
		return "database"
	}
	return "other"
}

func semanticEndpointProtocol(endpoint model.TopologyEndpoint) string {
	if protocol := strings.TrimSpace(endpoint.Protocol); protocol != "" && strings.ToUpper(protocol) != "TCP" {
		return protocol
	}
	name := strings.ToLower(strings.TrimSpace(endpoint.ServiceName))
	switch {
	case strings.Contains(name, "nginx") || strings.Contains(name, "gateway") || endpoint.Port == 80 || endpoint.Port == 443:
		return "HTTP health"
	case strings.Contains(name, "jvm") || strings.Contains(name, "tomcat") || strings.Contains(name, "app") || endpoint.Port == 8080 || endpoint.Port == 8081:
		return "JVM app probe"
	case strings.Contains(name, "redis") || endpoint.Port == 6379:
		return "Redis PING"
	case strings.Contains(name, "oracle") || endpoint.Port == 1521:
		return "Oracle listener probe"
	case strings.Contains(name, "mysql") || endpoint.Port == 3306:
		return "MySQL probe"
	case strings.Contains(name, "postgres") || endpoint.Port == 5432:
		return "Postgres probe"
	case endpoint.Port == 9090 || endpoint.Port == 9091 || endpoint.Port == 9100 || endpoint.Port == 9101:
		return "Prometheus scrape"
	default:
		return "TCP connect fallback"
	}
}

func inferredEdgeProtocol(graph *model.TopologyGraph, nodeID string) string {
	for _, node := range graph.Nodes {
		if node.ID == nodeID {
			return semanticEndpointProtocol(model.TopologyEndpoint{IP: node.IP, Port: node.Port, ServiceName: node.ServiceName})
		}
	}
	return "TCP connect fallback"
}

func nodeStatus(graph *model.TopologyGraph, id string) string {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node.Status
		}
	}
	return ""
}

func hasEdge(graph *model.TopologyGraph, id string) bool {
	for _, edge := range graph.Edges {
		if edge.ID == id {
			return true
		}
	}
	return false
}

func nodeLinkedStatus(graph *model.TopologyGraph, id string) string {
	status := nodeStatus(graph, id)
	if status == "online" || status == "connected" {
		return "connected"
	}
	return "disconnected"
}

func buildTopologyDiscoveryPlan(hosts []string, endpoints []model.TopologyEndpoint, graph *model.TopologyGraph, useAI bool) *model.TopologyDiscovery {
	plan := &model.TopologyDiscovery{
		Planner:       "ai-workbench-main-agent",
		Status:        "heuristic",
		Summary:       "Main Agent builds a layered business topology from user scope, Prometheus labels, Catpaw agent status, and service-aware probes. Agents are metadata, not topology nodes.",
		DataSources:   []string{"user_scope", "prometheus", "catpaw_agents", "port_connectivity"},
		ScopeHosts:    hosts,
		BusinessChain: inferBusinessChain(endpoints),
		Notes:         []string{"Only user-provided IPs and ports are discovered; unrelated targets are ignored.", "Offline targets affect link color only; historical metrics still participate in discovery.", "Catpaw/Main Agent status is described in host metadata and discovery notes, not drawn as business nodes."},
	}
	if !useAI {
		return plan
	}
	localSummary := localTopologyPlannerSummary(hosts, endpoints)
	plan.Status = "ai_assisted"
	plan.Summary = localSummary
	summary, err := callTopologyPlannerAI(hosts, endpoints, graph)
	if err != nil {
		plan.Error = err.Error()
		plan.Notes = append(plan.Notes, "External AI provider is unavailable; AI WorkBench main-agent deterministic planner generated the topology.")
		return plan
	}
	plan.Summary = summary
	plan.Notes = append(plan.Notes, "External AI provider enhanced the main-agent topology plan.")
	return plan
}

func localTopologyPlannerSummary(hosts []string, endpoints []model.TopologyEndpoint) string {
	chain := inferBusinessChain(endpoints)
	if len(chain) == 0 {
		return fmt.Sprintf("AI WorkBench main agent planned a scoped tree for %d user-selected hosts. Catpaw agent status is metadata; observability targets stay in the observability layer.", len(hosts))
	}
	return fmt.Sprintf("AI WorkBench main agent planned a layered business tree for %d scoped hosts: %s. Catpaw agents are metadata, and Prometheus/Pushgateway/exporters are observability nodes, not business services.", len(hosts), strings.Join(chain, " -> "))
}

func inferBusinessChain(endpoints []model.TopologyEndpoint) []string {
	frontends, apps, middleware, databases := []string{}, []string{}, []string{}, []string{}
	for _, endpoint := range endpoints {
		label := fmt.Sprintf("%s:%d %s", endpoint.IP, endpoint.Port, strings.TrimSpace(endpoint.ServiceName))
		switch classifyEndpointRole(endpoint) {
		case "frontend":
			frontends = append(frontends, label)
		case "app":
			apps = append(apps, label)
		case "middleware":
			middleware = append(middleware, label)
		case "database":
			databases = append(databases, label)
		}
	}
	chain := []string{}
	if len(frontends) > 0 {
		chain = append(chain, "Entry layer: "+strings.Join(frontends, ", "))
	}
	if len(apps) > 0 {
		chain = append(chain, "Application layer: "+strings.Join(apps, ", "))
	}
	if len(middleware) > 0 {
		chain = append(chain, "Middleware layer: "+strings.Join(middleware, ", "))
	}
	if len(databases) > 0 {
		chain = append(chain, "Database layer: "+strings.Join(databases, ", "))
	}
	if len(chain) == 0 {
		chain = append(chain, "No explicit business chain recognized; add service names or ports.")
	}
	return chain
}
