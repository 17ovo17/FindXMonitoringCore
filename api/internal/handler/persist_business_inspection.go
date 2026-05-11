package handler

import (
	"fmt"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func newBusinessInspectionState(business model.TopologyBusiness) businessInspectionState {
	metrics := ensureBusinessInspectionMetricCoverage(business, businessMetricSamples(business.Hosts, business.Endpoints))
	return businessInspectionState{
		business:    business,
		generatedAt: time.Now(),
		metrics:     metrics,
		processes:   businessProcesses(business),
		resources:   businessResources(business),
		alerts:      businessAlerts(business.Hosts),
		findings:    topologyFindings(business),
		planner:     "ai-workbench-main-agent+deterministic-tools",
		score:       100,
		status:      "healthy",
	}
}

func applyDeterministicBusinessInspection(state *businessInspectionState) {
	appendAITopologyFindings(state)
	for _, alert := range state.alerts {
		applyBusinessAlertScore(state, alert)
	}
	for _, metric := range state.metrics {
		applyBusinessMetricScore(state, metric)
	}
	for _, process := range state.processes {
		applyBusinessProcessScore(state, process)
	}
	state.aiSuggestions = businessInspectionSuggestions(state.business, state.metrics, state.processes, state.alerts)
	state.recommendations = append(state.recommendations, state.aiSuggestions...)
	normalizeBusinessInspectionScore(state)
}

func appendAITopologyFindings(state *businessInspectionState) {
	req := aiTopologyGenerateRequest{ServiceName: state.business.Name, Hosts: state.business.Hosts, Endpoints: state.business.Endpoints}
	graph := buildAITopologyGraph(req, "heuristic_fallback", "business inspection reused deterministic Topo-Architect graph")
	state.findings = append(state.findings, fmt.Sprintf("Topo-Architect graph summary: nodes=%d, links=%d, critical path=%s", graph.Summary.NodeCount, graph.Summary.LinkCount, strings.Join(graph.Summary.CriticalPath, " -> ")))
	for _, risk := range graph.Risks {
		state.findings = append(state.findings, fmt.Sprintf("Topology risk[%s]: %s - %s", risk.Severity, risk.Title, risk.Description))
	}
}

func applyBusinessAlertScore(state *businessInspectionState, alert *model.AlertRecord) {
	if alert.Status != "firing" {
		return
	}
	state.score -= 18
	state.findings = append(state.findings, fmt.Sprintf("Unresolved alert: %s (%s)", alert.Title, alert.TargetIP))
}

func applyBusinessMetricScore(state *businessInspectionState, metric model.BusinessMetricSample) {
	switch metric.Status {
	case "critical":
		state.score -= 12
		state.recommendations = append(state.recommendations, businessMetricRecommendation(metric))
	case "warning":
		state.score -= 6
	case "unknown":
		state.score -= 3
	}
}

func applyBusinessProcessScore(state *businessInspectionState, process model.BusinessProcess) {
	if process.Status == "running" {
		return
	}
	state.score -= 8
	state.findings = append(state.findings, fmt.Sprintf("Process or port abnormal: %s %s:%d", process.Name, process.IP, process.Port))
}

func normalizeBusinessInspectionScore(state *businessInspectionState) {
	if state.score < 0 {
		state.score = 0
	}
	if state.score < 60 {
		state.status = "critical"
	} else if state.score < 85 {
		state.status = "warning"
	}
}

func applyBusinessInspectionAI(state *businessInspectionState) {
	result, err := callBusinessInspectionAI(state.business, state.metrics, state.processes, state.resources, state.alerts, state.findings, state.recommendations, state.score, state.status)
	if err != nil {
		state.aiError = err.Error()
		state.findings = append(state.findings, "External AI inspection is unavailable; deterministic evidence was used.")
		return
	}
	state.planner = "ai-workbench-main-agent+external-ai+deterministic-tools"
	mergeBusinessInspectionAIResult(state, result)
}

func mergeBusinessInspectionAIResult(state *businessInspectionState, result businessInspectionAIResult) {
	state.aiAnalysis = strings.TrimSpace(result.Analysis)
	if state.aiAnalysis == "" {
		state.aiAnalysis = strings.TrimSpace(result.Summary)
	}
	if hasRedisEndpoint(state.business.Endpoints) && !strings.Contains(strings.ToLower(state.aiAnalysis), "redis") {
		state.aiAnalysis = strings.TrimSpace(state.aiAnalysis + "; registered Redis is included in the middleware-layer AI inspection.")
	}
	mergeAIInspectionLists(state, result)
	if result.Score > 0 && result.Score <= 100 {
		state.score = int(float64(state.score)*0.6 + float64(result.Score)*0.4)
	}
	normalizeAIInspectionStatus(state, result.Status)
}

func mergeAIInspectionLists(state *businessInspectionState, result businessInspectionAIResult) {
	if len(result.Findings) > 0 {
		state.findings = compactTextList(result.Findings, 6)
	}
	if len(result.Recommendations) > 0 {
		state.recommendations = compactTextList(result.Recommendations, 7)
		state.aiSuggestions = state.recommendations
	}
}

func normalizeAIInspectionStatus(state *businessInspectionState, aiStatus string) {
	if aiStatus != "healthy" && aiStatus != "warning" && aiStatus != "critical" {
		return
	}
	if state.score >= 85 {
		state.status = "healthy"
	} else if state.score >= 60 {
		state.status = "warning"
	} else {
		state.status = "critical"
	}
}

func buildBusinessInspectionResult(state businessInspectionState) model.BusinessInspection {
	finalizeBusinessInspectionLists(&state)
	summary := businessInspectionSummary(state)
	return model.BusinessInspection{
		BusinessID:        state.business.ID,
		BusinessName:      state.business.Name,
		Status:            state.status,
		Score:             state.score,
		Summary:           summary,
		GeneratedAt:       state.generatedAt,
		Attributes:        state.business.Attributes,
		Metrics:           state.metrics,
		Processes:         state.processes,
		Resources:         state.resources,
		Alerts:            state.alerts,
		TopologyFindings:  state.findings,
		Recommendations:   state.recommendations,
		AISuggestions:     state.aiSuggestions,
		DataSources:       businessInspectionDataSources(state.aiError),
		Planner:           state.planner,
		AIAnalysis:        compactAIInspectionText(state.aiAnalysis),
		AIError:           state.aiError,
		ExecutiveSummary:  summary,
		RiskLevel:         state.status,
		TopFindings:       limitStrings(state.findings, 5),
		AIRecommendations: limitStrings(state.recommendations, 5),
		EvidenceRefs:      businessInspectionEvidenceRefs(state),
	}
}

func finalizeBusinessInspectionLists(state *businessInspectionState) {
	if len(state.recommendations) == 0 {
		state.recommendations = append(state.recommendations, "No blocking risk was found; continue watching SLO, alerts, and core process health.")
	}
	state.findings = compactTextList(state.findings, 8)
	state.recommendations = compactTextList(state.recommendations, 7)
	state.aiSuggestions = compactTextList(state.aiSuggestions, 7)
}

func businessInspectionSummary(state businessInspectionState) string {
	if state.aiAnalysis != "" {
		return compactAIInspectionText(state.aiAnalysis)
	}
	return fmt.Sprintf("Business inspection completed: %d hosts, %d endpoints, %d alerts, %d metric samples.", len(state.business.Hosts), len(state.business.Endpoints), len(state.alerts), len(state.metrics))
}

func businessInspectionDataSources(aiError string) []string {
	dataSources := []string{"topology", "prometheus", "catpaw", "alerts", "business_attributes", "ai_provider"}
	if aiError != "" {
		dataSources = append(dataSources, "ai_provider_unavailable")
	}
	return dataSources
}

func businessInspectionEvidenceRefs(state businessInspectionState) []string {
	return []string{
		fmt.Sprintf("topology:%s", state.business.ID),
		fmt.Sprintf("metrics:%d", len(state.metrics)),
		fmt.Sprintf("processes:%d", len(state.processes)),
		fmt.Sprintf("alerts:%d", len(state.alerts)),
	}
}
