package handler

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func MonitorHealth(c *gin.Context) {
	agents := store.ListFindXAgents()
	online := 0
	for _, agent := range agents {
		if agent.Status == "online" {
			online++
		}
	}
	health := model.MonitorHealth{
		Status:      monitorStatus(agents),
		Mode:        "findx-core",
		Storage:     monitorStorageHealth(),
		Targets:     len(store.ListMonitorTargets()),
		Agents:      len(agents),
		AgentOnline: online,
		GeneratedAt: time.Now(),
	}
	c.JSON(http.StatusOK, health)
}

func ListMonitorTargets(c *gin.Context) {
	targets := store.ListMonitorTargets()
	status := strings.TrimSpace(c.Query("status"))
	if status != "" {
		targets = filterMonitorTargetsByStatus(targets, status)
	}
	c.JSON(http.StatusOK, targets)
}

func GetMonitorTarget(c *gin.Context) {
	target, ok := store.GetMonitorTarget(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}
	c.JSON(http.StatusOK, target)
}

func SaveMonitorTarget(c *gin.Context) {
	var target model.MonitorTarget
	if err := c.ShouldBindJSON(&target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target payload"})
		return
	}
	if id := strings.TrimSpace(c.Param("id")); id != "" {
		target.ID = id
	}
	if err := validateMonitorTarget(target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := store.UpsertMonitorTarget(&target)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "monitor storage unavailable"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func DeleteMonitorTarget(c *gin.Context) {
	ok, err := store.DeleteMonitorTarget(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "monitor storage unavailable"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListFindXAgents(c *gin.Context) {
	c.JSON(http.StatusOK, store.ListFindXAgents())
}

func FindXAgentRegister(c *gin.Context) {
	FindXAgentHeartbeat(c)
}

func FindXAgentHeartbeat(c *gin.Context) {
	if !validAgentToken(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent token"})
		return
	}
	var heartbeat model.FindXAgentHeartbeat
	if err := c.ShouldBindJSON(&heartbeat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid heartbeat payload"})
		return
	}
	if err := validateAgentHeartbeat(heartbeat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent, target, err := store.UpsertFindXAgentHeartbeat(heartbeat)
	if err != nil {
		if strings.Contains(err.Error(), "future") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "heartbeat time is too far in the future"})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "monitor storage unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "agent": agent, "target": target})
}

func monitorStatus(agents []*model.FindXAgent) string {
	storage := store.Health()
	if ok, _ := storage["mysql"].(bool); !ok {
		return "degraded"
	}
	if len(agents) == 0 {
		return "empty"
	}
	return "healthy"
}

func monitorStorageHealth() map[string]any {
	health := store.Health()
	out := map[string]any{
		"mysql": health["mysql"],
		"redis": health["redis"],
	}
	if ok, _ := health["mysql"].(bool); !ok {
		out["reason_code"] = "mysql_unavailable"
	}
	return out
}

func filterMonitorTargetsByStatus(targets []*model.MonitorTarget, status string) []*model.MonitorTarget {
	filtered := []*model.MonitorTarget{}
	for _, target := range targets {
		if target.Status == status {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

func validateMonitorTarget(target model.MonitorTarget) error {
	if strings.TrimSpace(target.Ident) == "" && strings.TrimSpace(target.IP) == "" && strings.TrimSpace(target.Hostname) == "" {
		return errBadRequest("ident, ip, or hostname is required")
	}
	if strings.TrimSpace(target.IP) != "" {
		if _, ok := cleanIP(target.IP); !ok {
			return errBadRequest("valid ip is required")
		}
	}
	if !validMonitorTargetStatus(target.Status) {
		return errBadRequest("invalid target status")
	}
	return nil
}

func validateAgentHeartbeat(heartbeat model.FindXAgentHeartbeat) error {
	if strings.TrimSpace(heartbeat.Ident) == "" && strings.TrimSpace(heartbeat.IP) == "" && strings.TrimSpace(heartbeat.Hostname) == "" {
		return errBadRequest("ident, ip, or hostname is required")
	}
	if strings.TrimSpace(heartbeat.IP) != "" {
		if _, ok := cleanIP(heartbeat.IP); !ok {
			return errBadRequest("valid ip is required")
		}
	}
	return nil
}

func validAgentToken(c *gin.Context) bool {
	expected := strings.TrimSpace(os.Getenv("FINDX_AGENT_TOKEN"))
	if expected == "" {
		expected = strings.TrimSpace(viper.GetString("findx_agents.shared_token"))
	}
	if expected == "" {
		return viper.GetBool("findx_agents.allow_anonymous")
	}
	actual := strings.TrimSpace(c.GetHeader("X-Agent-Token"))
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func validMonitorTargetStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "", "online", "warning", "offline", "unknown", "maintenance":
		return true
	default:
		return false
	}
}

type errBadRequest string

func (e errBadRequest) Error() string { return string(e) }
