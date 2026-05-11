package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

type aiTopologyGenerateRequest struct {
	BusinessID   string                             `json:"business_id"`
	ServiceName  string                             `json:"service_name"`
	Hosts        []string                           `json:"hosts"`
	Endpoints    []model.TopologyEndpoint           `json:"endpoints"`
	Dependencies []model.AITopologyLink             `json:"dependencies"`
	HealthStatus map[string]model.AITopologyHealth  `json:"health_status"`
	Metrics      map[string]model.AITopologyMetrics `json:"metrics"`
	Alerts       map[string][]string                `json:"alerts"`
}

type businessInspectionState struct {
	business        model.TopologyBusiness
	generatedAt     time.Time
	metrics         []model.BusinessMetricSample
	processes       []model.BusinessProcess
	resources       []model.BusinessResource
	alerts          []*model.AlertRecord
	findings        []string
	recommendations []string
	aiSuggestions   []string
	aiAnalysis      string
	aiError         string
	planner         string
	score           int
	status          string
}

func buildBusinessInspection(business model.TopologyBusiness) model.BusinessInspection {
	state := newBusinessInspectionState(business)
	applyDeterministicBusinessInspection(&state)
	applyBusinessInspectionAI(&state)
	inspection := buildBusinessInspectionResult(state)
	if detailedReport := renderRichInspectionReport(inspection); detailedReport != "" {
		inspection.AIAnalysis = detailedReport
	}
	return inspection
}

func ListTopologyBusinesses(c *gin.Context) {
	c.JSON(http.StatusOK, store.ListTopologyBusinesses())
}

func GetTopologyBusiness(c *gin.Context) {
	if b, ok := store.GetTopologyBusiness(c.Param("id")); ok {
		c.JSON(http.StatusOK, b)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "business topology not found"})
}

func InspectTopologyBusiness(c *gin.Context) {
	business, ok := store.GetTopologyBusiness(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "business topology not found"})
		return
	}
	inspection := buildBusinessInspection(business)
	persistBusinessInspectionRecord(business, &inspection)
	c.JSON(http.StatusOK, inspection)
}

func SaveTopologyBusiness(c *gin.Context) {
	var req model.TopologyBusiness
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "business name is required"})
		return
	}
	if len(req.Graph.Nodes) == 0 {
		req.Graph = buildBusinessTopologyGraph(req)
	}
	c.JSON(http.StatusOK, store.SaveTopologyBusiness(req))
}

func DeleteTopologyBusiness(c *gin.Context) {
	store.DeleteTopologyBusiness(c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GenerateAITopology(c *gin.Context) {
	var req aiTopologyGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !hydrateAITopologyRequest(&req, c) {
		return
	}
	if len(req.Endpoints) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one business endpoint is required"})
		return
	}
	graph := buildAITopologyGraph(req, "heuristic_fallback", "AI provider was not used; deterministic Topo-Architect rules generated this graph")
	c.JSON(http.StatusOK, graph)
}

func buildBusinessTopologyGraph(req model.TopologyBusiness) model.TopologyGraph {
	graph := model.TopologyGraph{Nodes: []model.TopologyNode{}, Edges: []model.TopologyEdge{}}
	hosts := mergeHosts(normalizeHosts(req.Hosts), prometheusHostsForSelection(normalizeHosts(req.Hosts)))
	for index, host := range hosts {
		addHostDiscovery(&graph, host, index, time.Now(), endpointPortSet(req.Endpoints))
	}
	addUserDefinedEndpoints(&graph, req.Endpoints, time.Now())
	addInferredBusinessEdges(&graph, req.Endpoints, time.Now())
	graph.Discovery = buildTopologyDiscoveryPlan(hosts, req.Endpoints, &graph, false)
	layoutBusinessTree(&graph)
	return graph
}

func hydrateAITopologyRequest(req *aiTopologyGenerateRequest, c *gin.Context) bool {
	if id := strings.TrimSpace(req.BusinessID); id != "" {
		business, ok := store.GetTopologyBusiness(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "business topology not found"})
			return false
		}
		applyBusinessToAITopologyRequest(req, business)
	}
	if len(req.Endpoints) == 0 && len(req.Hosts) > 0 {
		req.Endpoints = endpointsFromHosts(req.Hosts)
	}
	return true
}

func applyBusinessToAITopologyRequest(req *aiTopologyGenerateRequest, business model.TopologyBusiness) {
	if req.ServiceName == "" {
		req.ServiceName = business.Name
	}
	if len(req.Hosts) == 0 {
		req.Hosts = business.Hosts
	}
	if len(req.Endpoints) == 0 {
		req.Endpoints = business.Endpoints
	}
}

func endpointsFromHosts(hosts []string) []model.TopologyEndpoint {
	endpoints := make([]model.TopologyEndpoint, 0, len(hosts))
	for _, h := range hosts {
		endpoints = append(endpoints, model.TopologyEndpoint{IP: h, Port: 80, ServiceName: h, Protocol: "HTTP"})
	}
	return endpoints
}

func persistBusinessInspectionRecord(business model.TopologyBusiness, inspection *model.BusinessInspection) {
	now := time.Now()
	target := "business:" + strings.TrimSpace(business.ID)
	if strings.TrimSpace(business.ID) == "" {
		target = "business:" + strings.TrimSpace(business.Name)
	}
	recordID := "biz-inspect-" + store.NewID()
	inspection.DiagnoseRecordID = recordID
	raw, _ := json.MarshalIndent(inspection, "", "  ")
	md := renderBusinessInspectionMarkdown(*inspection)
	store.AddRecord(&model.DiagnoseRecord{
		ID:            recordID,
		TargetIP:      target,
		Trigger:       "business_inspection",
		Source:        "business_inspection",
		DataSource:    "business_inspection",
		Status:        model.StatusDone,
		Report:        md,
		SummaryReport: md,
		RawReport:     string(raw),
		AlertTitle:    fmt.Sprintf("Business inspection: %s, %d hosts", business.Name, len(business.Hosts)),
		CreateTime:    now,
		EndTime:       &now,
	})
}

func renderBusinessInspectionMarkdown(inspection model.BusinessInspection) string {
	summary := cleanInspectionSummary(inspection.Summary)
	lines := []string{
		fmt.Sprintf("# %s business inspection", inspection.BusinessName),
		fmt.Sprintf("- Status: %s", inspection.Status),
		fmt.Sprintf("- Score: %d", inspection.Score),
		fmt.Sprintf("- Data sources: %s", strings.Join(inspection.DataSources, ", ")),
		"",
		"## Summary",
		summary,
	}
	appendInspectionMarkdownSections(&lines, inspection)
	return strings.Join(lines, "\n")
}

func appendInspectionMarkdownSections(lines *[]string, inspection model.BusinessInspection) {
	if strings.TrimSpace(inspection.AIAnalysis) != "" {
		*lines = append(*lines, "", "## AI analysis", "", inspection.AIAnalysis)
	}
	if len(inspection.TopologyFindings) > 0 {
		*lines = append(*lines, "", "## Key findings")
		for _, item := range compactTextList(inspection.TopologyFindings, 8) {
			*lines = append(*lines, "- "+item)
		}
	}
	if len(inspection.AISuggestions) > 0 {
		*lines = append(*lines, "", "## AI suggestions")
		for _, item := range compactTextList(inspection.AISuggestions, 7) {
			*lines = append(*lines, "- "+item)
		}
	}
	appendInspectionMetricMarkdown(lines, inspection)
}

func appendInspectionMetricMarkdown(lines *[]string, inspection model.BusinessInspection) {
	if len(inspection.Alerts) > 0 {
		*lines = append(*lines, "", fmt.Sprintf("## Alert summary\n- firing: %d, resolved: %d", countBusinessAlerts(inspection.Alerts, "firing"), countBusinessAlerts(inspection.Alerts, "resolved")))
	}
	if len(inspection.Metrics) == 0 {
		return
	}
	*lines = append(*lines, "", "## Metric summary")
	for _, metric := range limitBusinessMetrics(inspection.Metrics, 10) {
		*lines = append(*lines, fmt.Sprintf("- %s %s: %.2f%s (%s)", metric.IP, metric.Name, metric.Value, metric.Unit, metric.Status))
	}
}

func cleanInspectionSummary(raw string) string {
	if !strings.Contains(raw, "{") && !strings.Contains(raw, "\"evidence\"") {
		return raw
	}
	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "{") || strings.HasPrefix(t, "\"") || strings.Contains(t, "{\"") || strings.Contains(t, "\"evidence\"") {
			continue
		}
		lines = append(lines, t)
	}
	if len(lines) > 0 {
		return strings.Join(lines, " ")
	}
	return raw
}
