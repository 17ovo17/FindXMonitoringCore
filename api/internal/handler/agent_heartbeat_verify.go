package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/monitoring"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

type heartbeatVerifyResult struct {
	AgentID          string    `json:"agent_id"`
	Ident            string    `json:"ident"`
	Verdict          string    `json:"verdict"`
	HeartbeatCheck   checkItem `json:"heartbeat_check"`
	ProcessCheck     checkItem `json:"process_check"`
	DataArrivalCheck checkItem `json:"data_arrival_check"`
	VerifiedAt       time.Time `json:"verified_at"`
}

type checkItem struct {
	Status  string `json:"status"`
	Detail  string `json:"detail"`
	Latency int64  `json:"latency_ms,omitempty"`
}

func VerifyAgentHeartbeat(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent id is required"})
		return
	}
	agent, ok := store.GetFindXAgent(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	result := verifyAgent(c.Request.Context(), agent)
	c.JSON(http.StatusOK, result)
}

type batchVerifyRequest struct {
	AgentIDs []string `json:"agent_ids"`
}

func BatchVerifyAgentHeartbeat(c *gin.Context) {
	var req batchVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	if len(req.AgentIDs) == 0 {
		agents := store.ListFindXAgents()
		for _, agent := range agents {
			req.AgentIDs = append(req.AgentIDs, agent.ID)
		}
	}
	const maxBatch = 50
	if len(req.AgentIDs) > maxBatch {
		req.AgentIDs = req.AgentIDs[:maxBatch]
	}
	results := make([]heartbeatVerifyResult, 0, len(req.AgentIDs))
	for _, id := range req.AgentIDs {
		agent, ok := store.GetFindXAgent(id)
		if !ok {
			results = append(results, heartbeatVerifyResult{
				AgentID:        id,
				Verdict:        "offline",
				HeartbeatCheck: checkItem{Status: "fail", Detail: "agent not found in registry"},
				ProcessCheck:   checkItem{Status: "skip", Detail: "agent not found"},
				DataArrivalCheck: checkItem{Status: "skip", Detail: "agent not found"},
				VerifiedAt:     time.Now(),
			})
			continue
		}
		results = append(results, verifyAgent(c.Request.Context(), agent))
	}
	c.JSON(http.StatusOK, gin.H{"results": results, "total": len(results)})
}

func verifyAgent(ctx context.Context, agent *model.FindXAgent) heartbeatVerifyResult {
	now := time.Now()
	result := heartbeatVerifyResult{
		AgentID:    agent.ID,
		Ident:      agent.Ident,
		VerifiedAt: now,
	}

	// Check 1: Control plane heartbeat — time since last heartbeat report
	result.HeartbeatCheck = checkHeartbeatAge(agent, now)

	// Check 2: Process status — try to reach agent health endpoint
	result.ProcessCheck = checkAgentProcess(ctx, agent)

	// Check 3: Data arrival — query Prometheus for recent metrics from this agent
	result.DataArrivalCheck = checkDataArrival(ctx, agent)

	// Determine verdict based on checks
	result.Verdict = determineVerdict(result)
	return result
}

func checkHeartbeatAge(agent *model.FindXAgent, now time.Time) checkItem {
	if agent.LastSeen.IsZero() {
		return checkItem{Status: "fail", Detail: "no heartbeat ever received"}
	}
	delta := now.Sub(agent.LastSeen)
	if delta < 2*time.Minute {
		return checkItem{Status: "pass", Detail: fmt.Sprintf("last heartbeat %ds ago", int(delta.Seconds()))}
	}
	if delta < 5*time.Minute {
		return checkItem{Status: "warn", Detail: fmt.Sprintf("last heartbeat %ds ago (stale)", int(delta.Seconds()))}
	}
	return checkItem{Status: "fail", Detail: fmt.Sprintf("last heartbeat %ds ago (expired)", int(delta.Seconds()))}
}

func checkAgentProcess(ctx context.Context, agent *model.FindXAgent) checkItem {
	if agent.IP == "" {
		return checkItem{Status: "skip", Detail: "no IP address available for health check"}
	}
	// Try common agent health endpoint ports
	healthURL := fmt.Sprintf("http://%s:9100/health", agent.IP)
	start := time.Now()
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return checkItem{Status: "fail", Detail: "failed to create health request", Latency: time.Since(start).Milliseconds()}
	}
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return checkItem{Status: "fail", Detail: "agent health endpoint unreachable", Latency: latency}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return checkItem{Status: "pass", Detail: fmt.Sprintf("health endpoint responded %d", resp.StatusCode), Latency: latency}
	}
	return checkItem{Status: "warn", Detail: fmt.Sprintf("health endpoint returned %d", resp.StatusCode), Latency: latency}
}

func checkDataArrival(ctx context.Context, agent *model.FindXAgent) checkItem {
	promBase := resolvePrometheusURL(defaultPrometheusDatasourceID())
	if promBase == "" {
		return checkItem{Status: "skip", Detail: "prometheus datasource not configured"}
	}
	// Query for any metric from this agent in the last 5 minutes
	ident := agent.Ident
	if ident == "" {
		ident = agent.IP
	}
	if ident == "" {
		return checkItem{Status: "skip", Detail: "no agent identifier for prometheus query"}
	}
	query := fmt.Sprintf(`count({ident="%s"}[5m])`, ident)
	start := time.Now()
	result, err := monitoring.NewPrometheusGateway(nil).QueryInstant(ctx, monitoring.PrometheusQueryRequest{
		BaseURL: promBase,
		Query:   query,
		Timeout: 5 * time.Second,
	})
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return checkItem{Status: "fail", Detail: "prometheus query failed", Latency: latency}
	}
	if result.Stats.SeriesCount > 0 || result.Stats.SampleCount > 0 {
		return checkItem{Status: "pass", Detail: fmt.Sprintf("data arriving (%d series)", result.Stats.SeriesCount), Latency: latency}
	}
	// Check if result data has any values
	if hasPrometheusResults(result.Data) {
		return checkItem{Status: "pass", Detail: "data arriving (confirmed via query)", Latency: latency}
	}
	return checkItem{Status: "fail", Detail: "no data from agent in last 5 minutes", Latency: latency}
}

func hasPrometheusResults(data map[string]any) bool {
	if data == nil {
		return false
	}
	resultRaw, ok := data["result"]
	if !ok {
		return false
	}
	results, ok := resultRaw.([]any)
	if !ok {
		return false
	}
	return len(results) > 0
}

func determineVerdict(result heartbeatVerifyResult) string {
	passCount := 0
	failCount := 0
	checks := []checkItem{result.HeartbeatCheck, result.ProcessCheck, result.DataArrivalCheck}
	for _, check := range checks {
		switch check.Status {
		case "pass":
			passCount++
		case "fail":
			failCount++
		}
	}
	if failCount == 0 && passCount >= 2 {
		return "online"
	}
	if passCount >= 1 {
		return "degraded"
	}
	return "offline"
}
