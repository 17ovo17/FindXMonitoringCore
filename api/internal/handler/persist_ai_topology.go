package handler

import (
	"fmt"
	"sort"
	"strings"

	"ai-workbench-api/internal/model"
)

func buildAITopologyGraph(req aiTopologyGenerateRequest, planner, plannerError string) model.AITopologyGraph {
	nodes := buildAITopologyNodes(req, buildAITopologyEndpointMap(req.Endpoints))
	links := inferAITopologyLinks(nodes, req.Dependencies)
	risks := detectAITopologyRisks(nodes, links)
	summary := summarizeAITopology(req.ServiceName, planner, plannerError, nodes, links)
	return model.AITopologyGraph{Nodes: nodes, Links: links, Risks: risks, Summary: summary}
}

func buildAITopologyEndpointMap(endpoints []model.TopologyEndpoint) map[string]model.TopologyEndpoint {
	endpointMap := map[string]model.TopologyEndpoint{}
	for _, endpoint := range endpoints {
		if strings.TrimSpace(endpoint.IP) == "" || endpoint.Port <= 0 || isAgentLikeEndpoint(endpoint) {
			continue
		}
		endpoint.ServiceName = normalizeEndpointServiceName(endpoint)
		endpointMap[aiTopologyNodeID(endpoint)] = endpoint
	}
	return endpointMap
}

func buildAITopologyNodes(req aiTopologyGenerateRequest, endpointMap map[string]model.TopologyEndpoint) []model.AITopologyNode {
	ids := make([]string, 0, len(endpointMap))
	for id := range endpointMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	nodes := make([]model.AITopologyNode, 0, len(ids))
	for _, id := range ids {
		endpoint := endpointMap[id]
		layer := classifyAITopologyLayer(endpoint)
		nodes = append(nodes, model.AITopologyNode{
			ID:       id,
			IP:       endpoint.IP,
			Hostname: endpoint.IP,
			Layer:    layer,
			Services: []model.AITopologyService{aiTopologyService(endpoint, layer)},
			Health:   aiTopologyNodeHealth(req, endpoint, id),
			Metrics:  aiTopologyNodeMetrics(req, endpoint, id),
			Alerts:   compactTextList(append(req.Alerts[endpoint.IP], req.Alerts[id]...), 12),
		})
	}
	sortAITopologyNodes(nodes)
	return nodes
}

func aiTopologyService(endpoint model.TopologyEndpoint, layer string) model.AITopologyService {
	serviceName := strings.TrimSpace(endpoint.ServiceName)
	if serviceName == "" {
		serviceName = fmt.Sprintf("port-%d", endpoint.Port)
	}
	return model.AITopologyService{Name: serviceName, Port: endpoint.Port, Role: aiServiceRole(layer, serviceName)}
}

func aiTopologyNodeHealth(req aiTopologyGenerateRequest, endpoint model.TopologyEndpoint, id string) model.AITopologyHealth {
	health := req.HealthStatus[endpoint.IP]
	if health.Status == "" {
		health = req.HealthStatus[id]
	}
	if health.Status == "" {
		health = model.AITopologyHealth{Score: 100, Status: "healthy"}
	}
	return normalizeAITopologyHealth(health)
}

func aiTopologyNodeMetrics(req aiTopologyGenerateRequest, endpoint model.TopologyEndpoint, id string) model.AITopologyMetrics {
	metrics := req.Metrics[endpoint.IP]
	if metrics == (model.AITopologyMetrics{}) {
		metrics = req.Metrics[id]
	}
	return metrics
}

func aiTopologyNodeID(endpoint model.TopologyEndpoint) string {
	return fmt.Sprintf("%s-%s-%d", classifyAITopologyLayer(endpoint), sanitizeID(endpoint.IP), endpoint.Port)
}

func isAgentLikeEndpoint(endpoint model.TopologyEndpoint) bool {
	name := strings.ToLower(strings.TrimSpace(endpoint.ServiceName))
	return strings.Contains(name, "catpaw") || strings.Contains(name, "main agent") || strings.Contains(name, "ai-agent") || strings.Contains(name, "agent")
}

func classifyAITopologyLayer(endpoint model.TopologyEndpoint) string {
	name := strings.ToLower(strings.TrimSpace(endpoint.ServiceName))
	port := endpoint.Port
	switch {
	case strings.Contains(name, "nginx") || strings.Contains(name, "haproxy") || strings.Contains(name, "traefik") || strings.Contains(name, "kong") || strings.Contains(name, "envoy") || port == 80 || port == 443 || port == 8443:
		return "gateway"
	case strings.Contains(name, "redis") || strings.Contains(name, "memcached") || port == 6379 || port == 11211:
		return "cache"
	case strings.Contains(name, "kafka") || strings.Contains(name, "rabbitmq") || strings.Contains(name, "rocketmq") || strings.Contains(name, "mq") || port == 9092 || port == 5672 || port == 9876:
		return "mq"
	case strings.Contains(name, "mysql") || strings.Contains(name, "postgres") || strings.Contains(name, "oracle") || strings.Contains(name, "mongo") || strings.Contains(name, "elasticsearch") || strings.Contains(name, "database") || port == 3306 || port == 5432 || port == 1521 || port == 9200:
		return "db"
	case strings.Contains(name, "etcd") || strings.Contains(name, "zookeeper") || strings.Contains(name, "consul") || strings.Contains(name, "zk") || port == 2379 || port == 2181 || port == 8300:
		return "infra"
	case strings.Contains(name, "prometheus") || strings.Contains(name, "categraf") || strings.Contains(name, "exporter") || strings.Contains(name, "grafana") || port == 9090 || port == 9100 || port == 9101:
		return "monitor"
	case strings.Contains(name, "jvm") || strings.Contains(name, "java") || strings.Contains(name, "python") || strings.Contains(name, "node") || strings.Contains(name, "service") || strings.Contains(name, "api") || strings.Contains(name, "app") || (port >= 8000 && port <= 9000):
		return "app"
	default:
		return "app"
	}
}

func normalizeAITopologyHealth(health model.AITopologyHealth) model.AITopologyHealth {
	if health.Score < 0 {
		health.Score = 0
	}
	if health.Score > 100 {
		health.Score = 100
	}
	if health.Status == "" || health.Status == "unknown" {
		if health.Score >= 85 {
			health.Status = "healthy"
		} else if health.Score >= 70 {
			health.Status = "warning"
		} else if health.Score > 0 {
			health.Status = "danger"
		} else {
			health.Status = "unknown"
		}
	}
	if health.Status == "critical" {
		health.Status = "danger"
	}
	return health
}

func aiServiceRole(layer, service string) string {
	switch layer {
	case "gateway":
		return "入口网关"
	case "app":
		return "业务服务"
	case "cache":
		return "缓存"
	case "mq":
		return "消息队列"
	case "db":
		if strings.Contains(strings.ToLower(service), "slave") || strings.Contains(strings.ToLower(service), "standby") {
			return "从库"
		}
		return "数据库"
	case "infra":
		return "注册/配置中心"
	case "monitor":
		return "监控采集"
	default:
		return "业务组件"
	}
}

func inferAITopologyLinks(nodes []model.AITopologyNode, explicit []model.AITopologyLink) []model.AITopologyLink {
	links := []model.AITopologyLink{}
	nodeIDs := map[string]bool{}
	byLayer := map[string][]model.AITopologyNode{}
	for _, node := range nodes {
		nodeIDs[node.ID] = true
		byLayer[node.Layer] = append(byLayer[node.Layer], node)
	}
	for _, link := range explicit {
		links = appendAITopologyLink(links, nodeIDs, link)
	}
	add := func(link model.AITopologyLink) {
		links = appendAITopologyLink(links, nodeIDs, link)
	}
	for _, gateway := range byLayer["gateway"] {
		for _, app := range byLayer["app"] {
			add(model.AITopologyLink{Source: gateway.ID, Target: app.ID, Type: "HTTP", Label: "负载均衡"})
		}
	}
	for _, app := range byLayer["app"] {
		for _, cache := range byLayer["cache"] {
			add(model.AITopologyLink{Source: app.ID, Target: cache.ID, Type: "Redis", Label: "缓存读写"})
		}
		for _, mq := range byLayer["mq"] {
			add(model.AITopologyLink{Source: app.ID, Target: mq.ID, Type: "MQ", Label: "消息生产/消费"})
		}
		for _, db := range byLayer["db"] {
			add(model.AITopologyLink{Source: app.ID, Target: db.ID, Type: "DB", Label: "数据读写"})
		}
	}
	for _, infra := range byLayer["infra"] {
		for _, app := range byLayer["app"] {
			add(model.AITopologyLink{Source: infra.ID, Target: app.ID, Type: "Discovery", Label: "服务注册/配置发现", Dashed: true})
		}
	}
	for i, source := range byLayer["db"] {
		for j, target := range byLayer["db"] {
			if i >= j || !sameDBFamily(source, target) {
				continue
			}
			add(model.AITopologyLink{Source: source.ID, Target: target.ID, Type: "Replication", Label: "主从同步", Dashed: true, Relation: "replication"})
		}
	}
	return links
}

func appendAITopologyLink(links []model.AITopologyLink, nodeIDs map[string]bool, link model.AITopologyLink) []model.AITopologyLink {
	if link.Source == link.Target || !nodeIDs[link.Source] || !nodeIDs[link.Target] {
		return links
	}
	if link.Type == "" {
		link.Type = "TCP"
	}
	if link.Label == "" {
		link.Label = link.Type
	}
	for _, existing := range links {
		if existing.Source == link.Source && existing.Target == link.Target && existing.Label == link.Label {
			return links
		}
	}
	return append(links, link)
}

func sameDBFamily(a, b model.AITopologyNode) bool {
	if len(a.Services) == 0 || len(b.Services) == 0 {
		return false
	}
	an := strings.ToLower(a.Services[0].Name)
	bn := strings.ToLower(b.Services[0].Name)
	families := []string{"oracle", "mysql", "postgres", "mongo", "elasticsearch"}
	for _, family := range families {
		if strings.Contains(an, family) && strings.Contains(bn, family) {
			return true
		}
	}
	return false
}

func detectAITopologyRisks(nodes []model.AITopologyNode, links []model.AITopologyLink) []model.AITopologyRisk {
	risks := []model.AITopologyRisk{}
	layerCounts := map[string]int{}
	nodeMap := map[string]model.AITopologyNode{}
	degree := map[string]int{}
	for _, node := range nodes {
		layerCounts[node.Layer]++
		nodeMap[node.ID] = node
		degree[node.ID] = 0
	}
	for _, link := range links {
		degree[link.Source]++
		degree[link.Target]++
		if isCrossLayerRisk(nodeMap[link.Source].Layer, nodeMap[link.Target].Layer) {
			risks = append(risks, model.AITopologyRisk{Type: "cross_layer_direct", Severity: "medium", Title: "跨层直连", Description: fmt.Sprintf("%s -> %s 跨层级直连，需确认是否绕过标准业务链路", link.Source, link.Target), Nodes: []string{link.Source, link.Target}, Suggestion: "核对调用链配置；如为真实链路，应补充限流、超时、鉴权和监控。"})
		}
	}
	for _, layer := range []string{"gateway", "cache", "mq", "db", "infra"} {
		if layerCounts[layer] == 1 {
			for _, node := range nodes {
				if node.Layer == layer {
					risks = append(risks, model.AITopologyRisk{Type: "single_point", Severity: "high", Title: "单点风险", Description: fmt.Sprintf("%s 层仅 1 个节点：%s", layer, node.ID), Nodes: []string{node.ID}, Suggestion: "评估主备/集群化改造，并先补齐健康检查、自动拉起和容量预警。"})
					break
				}
			}
		}
	}
	for _, node := range nodes {
		if degree[node.ID] == 0 {
			risks = append(risks, model.AITopologyRisk{Type: "island", Severity: "medium", Title: "孤岛节点", Description: fmt.Sprintf("%s 无入边也无出边，可能为配置遗漏或未纳入业务链路", node.ID), Nodes: []string{node.ID}, Suggestion: "补充真实依赖关系或从业务拓扑中移除无关端口。"})
		}
		if node.Health.Status == "danger" {
			risks = append(risks, model.AITopologyRisk{Type: "blast_radius", Severity: "high", Title: "故障扩散风险", Description: fmt.Sprintf("%s 处于危险状态，可能影响其上下游链路", node.ID), Nodes: []string{node.ID}, Suggestion: "优先核查该节点进程、端口连通、连接池、慢查询/慢请求和未恢复告警。"})
		}
		if node.Health.Status == "unknown" {
			risks = append(risks, model.AITopologyRisk{Type: "observability_gap", Severity: "medium", Title: "监控盲区", Description: fmt.Sprintf("%s 缺少健康状态或指标证据", node.ID), Nodes: []string{node.ID}, Suggestion: "补齐 Categraf/Prometheus 指标、Catpaw 探针状态和告警路由。"})
		}
	}
	return risks
}

func isCrossLayerRisk(source, target string) bool {
	if source == "infra" && target == "app" {
		return false
	}
	if source == "gateway" && target == "app" {
		return false
	}
	if source == "app" && (target == "cache" || target == "mq" || target == "db") {
		return false
	}
	if source == "db" && target == "db" {
		return false
	}
	order := map[string]int{"gateway": 0, "app": 1, "cache": 2, "mq": 2, "db": 3, "infra": 4, "monitor": 5}
	return absInt(order[source]-order[target]) > 1
}

func summarizeAITopology(serviceName, planner, plannerError string, nodes []model.AITopologyNode, links []model.AITopologyLink) model.AITopologySummary {
	layers := map[string]int{}
	health := map[string]int{}
	for _, node := range nodes {
		layers[node.Layer]++
		health[node.Health.Status]++
	}
	critical := []string{}
	for _, layer := range []string{"gateway", "app", "cache", "db"} {
		for _, node := range nodes {
			if node.Layer == layer {
				critical = append(critical, node.ID)
				break
			}
		}
	}
	return model.AITopologySummary{ServiceName: serviceName, Planner: planner, NodeCount: len(nodes), LinkCount: len(links), LayerCounts: layers, HealthDistribution: health, CriticalPath: critical, Error: plannerError}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
