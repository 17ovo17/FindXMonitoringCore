package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
)

type businessInspectionAIResult struct {
	Summary         string   `json:"summary"`
	Analysis        string   `json:"analysis"`
	Status          string   `json:"status"`
	Score           int      `json:"score"`
	Findings        []string `json:"findings"`
	Recommendations []string `json:"recommendations"`
}

func callBusinessInspectionAI(business model.TopologyBusiness, metrics []model.BusinessMetricSample, processes []model.BusinessProcess, resources []model.BusinessResource, alerts []*model.AlertRecord, findings []string, recommendations []string, score int, status string) (businessInspectionAIResult, error) {
	apiKey := strings.TrimSpace(getAPIKey())
	if apiKey == "" || apiKey == "******" || strings.Contains(apiKey, "${") {
		return businessInspectionAIResult{}, fmt.Errorf("AI provider API key is not configured")
	}
	payload, err := businessInspectionAIPayload(business, metrics, processes, resources, alerts, findings, recommendations, score, status)
	if err != nil {
		return businessInspectionAIResult{}, err
	}
	content, err := postBusinessInspectionAI(apiKey, payload)
	if err != nil {
		return businessInspectionAIResult{}, err
	}
	result := parseBusinessInspectionAIContent(content, status, score, findings, recommendations)
	return normalizeBusinessInspectionAIResult(result), nil
}

func businessInspectionAIPayload(business model.TopologyBusiness, metrics []model.BusinessMetricSample, processes []model.BusinessProcess, resources []model.BusinessResource, alerts []*model.AlertRecord, findings []string, recommendations []string, score int, status string) ([]byte, error) {
	evidenceJSON, err := json.Marshal(businessInspectionEvidence(business, metrics, processes, resources, alerts, findings, recommendations, score, status))
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"model": resolveDefaultModel(),
		"messages": []map[string]string{
			{"role": "system", "content": businessInspectionSystemPrompt()},
			{"role": "user", "content": string(evidenceJSON)},
		},
		"stream": false,
	}
	return json.Marshal(payload)
}

func businessInspectionEvidence(business model.TopologyBusiness, metrics []model.BusinessMetricSample, processes []model.BusinessProcess, resources []model.BusinessResource, alerts []*model.AlertRecord, findings []string, recommendations []string, score int, status string) map[string]any {
	return map[string]any{
		"business":                      map[string]any{"id": business.ID, "name": business.Name, "hosts": business.Hosts, "endpoints": classifyEndpointsWithAI(business.Endpoints, false), "attributes": business.Attributes},
		"metrics":                       metrics,
		"processes":                     processes,
		"resources":                     resources,
		"alerts":                        alerts,
		"deterministic_findings":        findings,
		"deterministic_recommendations": recommendations,
		"deterministic_score":           score,
		"deterministic_status":          status,
	}
}

func businessInspectionSystemPrompt() string {
	return "You are the FindX platform main agent for business inspection. Analyze only the supplied tool evidence; do not invent hosts, metrics, or ports. Classify entry, application, middleware, and database layers. If Redis is registered, inspect it as middleware. Return JSON only with summary, analysis, status, score, findings, and recommendations."
}

func postBusinessInspectionAI(apiKey string, payload []byte) (string, error) {
	client := &http.Client{Timeout: chatUpstreamTimeout}
	req, err := http.NewRequest(http.MethodPost, getBaseURL()+"/chat/completions", bytes.NewReader(payload))
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
		return "", fmt.Errorf("AI business inspection status %d", resp.StatusCode)
	}
	return businessInspectionAIMessageContent(respBody)
}

func businessInspectionAIMessageContent(respBody []byte) (string, error) {
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
		return "", fmt.Errorf("AI business inspection returned empty content")
	}
	return trimJSONFence(parsed.Choices[0].Message.Content), nil
}

func trimJSONFence(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}

func parseBusinessInspectionAIContent(content, status string, score int, findings, recommendations []string) businessInspectionAIResult {
	var result businessInspectionAIResult
	var raw map[string]any
	rawOK := json.Unmarshal([]byte(content), &raw) == nil
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return fallbackBusinessInspectionAIResult(content, raw, rawOK, status, score, findings, recommendations)
	}
	if rawOK && strings.TrimSpace(result.Summary) == "" && strings.TrimSpace(result.Analysis) == "" {
		result = aiResultFromRaw(raw)
	}
	return result
}

func fallbackBusinessInspectionAIResult(content string, raw map[string]any, rawOK bool, status string, score int, findings, recommendations []string) businessInspectionAIResult {
	if rawOK {
		return aiResultFromRaw(raw)
	}
	return businessInspectionAIResult{Summary: content, Analysis: content, Status: status, Score: score, Findings: findings, Recommendations: recommendations}
}

func aiResultFromRaw(raw map[string]any) businessInspectionAIResult {
	result := businessInspectionAIResult{
		Summary:         aiText(raw["summary"]),
		Analysis:        aiText(raw["analysis"]),
		Status:          aiText(raw["status"]),
		Score:           aiInt(raw["score"]),
		Findings:        aiTextList(raw["findings"]),
		Recommendations: aiTextList(raw["recommendations"]),
	}
	if strings.TrimSpace(result.Summary) == "" && strings.TrimSpace(result.Analysis) == "" {
		result.Summary = summarizeAIInspectionMap(raw)
		result.Analysis = result.Summary
	}
	return result
}

func normalizeBusinessInspectionAIResult(result businessInspectionAIResult) businessInspectionAIResult {
	if strings.TrimSpace(result.Analysis) == "" {
		result.Analysis = result.Summary
	}
	result.Summary = compactAIInspectionText(result.Summary)
	result.Analysis = compactAIInspectionText(result.Analysis)
	result.Findings = compactTextList(result.Findings, 6)
	result.Recommendations = compactTextList(result.Recommendations, 7)
	return result
}

func compactAIInspectionText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	var raw map[string]any
	if json.Unmarshal([]byte(text), &raw) == nil {
		if summary := summarizeAIInspectionMap(raw); summary != "" {
			return summary
		}
	}
	if len([]rune(text)) > 600 {
		return string([]rune(text)[:600]) + "..."
	}
	return text
}

func summarizeAIInspectionMap(raw map[string]any) string {
	parts := []string{}
	if conclusion := firstAISectionText(raw, "business_health_conclusion"); conclusion != "" {
		parts = append(parts, "Conclusion: "+conclusion)
	}
	for _, key := range []string{"topology", "middleware_and_database_health", "middleware_and_database", "database_and_middleware_health", "process_and_port_status", "resource_metrics", "alerts", "alerts_assessment", "data_consistency", "data_consistency_observation"} {
		if text := firstAISectionText(raw, key); text != "" {
			parts = append(parts, humanAISectionName(key)+": "+text)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(limitStrings(parts, 6), "\n")
}

func firstAISectionText(raw map[string]any, key string) string {
	value, ok := raw[key]
	if !ok {
		return ""
	}
	if text := aiText(value); text != "" && !strings.HasPrefix(strings.TrimSpace(text), "{") {
		return text
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	if text := firstAIObjectText(obj); text != "" {
		return text
	}
	if evidence := aiTextList(obj["evidence"]); len(evidence) > 0 {
		return strings.Join(limitStrings(evidence, 2), "; ")
	}
	return firstNestedAISectionText(obj)
}

func firstAIObjectText(obj map[string]any) string {
	for _, field := range []string{"summary", "result", "status", "conclusion", "assessment", "action", "priority", "note"} {
		if text := aiText(obj[field]); text != "" {
			return text
		}
	}
	return ""
}

func firstNestedAISectionText(obj map[string]any) string {
	for _, nested := range []string{"middleware", "database", "entry_and_app"} {
		if child, ok := obj[nested].(map[string]any); ok {
			if text := firstAISectionText(map[string]any{nested: child}, nested); text != "" {
				return text
			}
		}
	}
	return ""
}

func humanAISectionName(key string) string {
	switch key {
	case "topology":
		return "Topology"
	case "middleware_and_database_health", "middleware_and_database", "database_and_middleware_health":
		return "Middleware/Database"
	case "process_and_port_status":
		return "Process/Port"
	case "resource_metrics":
		return "Resources"
	case "alerts", "alerts_assessment":
		return "Alerts"
	case "data_consistency", "data_consistency_observation":
		return "Data consistency"
	default:
		return key
	}
}
