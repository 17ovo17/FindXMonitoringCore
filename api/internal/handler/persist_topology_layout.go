package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func hasRedisEndpoint(endpoints []model.TopologyEndpoint) bool {
	for _, endpoint := range endpoints {
		name := strings.ToLower(endpoint.ServiceName)
		if strings.Contains(name, "redis") || endpoint.Port == 6379 || endpoint.Port == 6375 || endpoint.Port == 26379 {
			return true
		}
	}
	return false
}

func callTopologyPlannerAI(hosts []string, endpoints []model.TopologyEndpoint, graph *model.TopologyGraph) (string, error) {
	apiKey := strings.TrimSpace(getAPIKey())
	if apiKey == "" || apiKey == "******" || strings.Contains(apiKey, "${") {
		return "", fmt.Errorf("AI provider API key is not configured")
	}
	payload := map[string]any{
		"model": resolveDefaultModel(),
		"messages": []map[string]string{
			{"role": "system", "content": "You are the AI WorkBench platform main agent. Given user-scoped hosts, ports, Prometheus/Catpaw findings, return one concise topology planning sentence. Never add IPs not supplied by the user."},
			{"role": "user", "content": fmt.Sprintf("hosts=%v endpoints=%v nodes=%d edges=%d", hosts, endpoints, len(graph.Nodes), len(graph.Edges))},
		},
		"stream": false,
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 8 * time.Second}
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
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("AI planner status %d", resp.StatusCode)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("AI planner returned empty content")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func layoutBusinessTree(graph *model.TopologyGraph) {
	layers := map[int][]int{}
	for i := range graph.Nodes {
		layer := graph.Nodes[i].Layer
		if layer == 0 {
			layer = defaultTopologyLayer(graph.Nodes[i])
			graph.Nodes[i].Layer = layer
		}
		layers[layer] = append(layers[layer], i)
	}
	orderedLayers := []int{}
	for layer := range layers {
		orderedLayers = append(orderedLayers, layer)
	}
	sort.Ints(orderedLayers)
	for _, layer := range orderedLayers {
		items := layers[layer]
		for order, nodeIndex := range items {
			graph.Nodes[nodeIndex].X = 70 + float64(layer)*230
			graph.Nodes[nodeIndex].Y = 120 + float64(order)*140
		}
	}
}

func defaultTopologyLayer(node model.TopologyNode) int {
	switch node.Type {
	case "host":
		return 0
	case "frontend":
		return 1
	case "application", "backend", "service":
		return 2
	case "cache":
		return 3
	case "database":
		return 4
	case "monitor", "management":
		return 5
	default:
		return 2
	}
}

func endpointLayer(endpoint model.TopologyEndpoint) int {
	switch classifyEndpointRole(endpoint) {
	case "frontend":
		return 1
	case "app":
		return 2
	case "middleware":
		return 3
	case "database":
		return 4
	default:
		return 2
	}
}

func hasPrometheusTargetPort(host string, port int) bool {
	for _, target := range listPrometheusTargets() {
		if target.IP == host && target.Port == port {
			return true
		}
	}
	return false
}

func addCheckedEdge(graph *model.TopologyGraph, id, sourceID, targetID, protocol, label, host string, port int, now time.Time) {
	status := "connected"
	latency := 0
	errText := ""
	if !hasRecentPrometheusData(host) && !store.HasOnlineAgent(host) {
		status = "unknown"
		errText = "离线目标不影响历史指标发现"
	}
	graph.Edges = append(graph.Edges, model.TopologyEdge{ID: id, SourceID: sourceID, TargetID: targetID, Protocol: protocol, Direction: "forward", Label: label, Status: status, LatencyMs: latency, Error: errText, CreatedAt: now, UpdatedAt: now})
}

func statusFromDial(host string, port int) string {
	_, errText := checkTCP(host, port)
	if errText != "" {
		return "offline"
	}
	return "online"
}

func checkTCP(host string, port int) (int, string) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 220*time.Millisecond)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		return latency, err.Error()
	}
	_ = conn.Close()
	return latency, ""
}

func sanitizeID(value string) string {
	replacer := strings.NewReplacer(".", "-", ":", "-", "_", "-", " ", "-", "[", "", "]", "")
	return replacer.Replace(strings.ToLower(value))
}

func hasNode(graph *model.TopologyGraph, id string) bool {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return true
		}
	}
	return false
}
