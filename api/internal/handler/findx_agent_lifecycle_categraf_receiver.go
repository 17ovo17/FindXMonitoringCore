package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const categrafReceiverBodyLimit = 2 << 20

type categrafHeartbeatPayload struct {
	Ident        string            `json:"ident"`
	Agent        string            `json:"agent"`
	AgentID      string            `json:"agent_id"`
	Host         string            `json:"host"`
	Hostname     string            `json:"hostname"`
	IP           string            `json:"ip"`
	HostIP       string            `json:"host_ip"`
	Version      string            `json:"version"`
	AgentVersion string            `json:"agent_version"`
	OS           string            `json:"os"`
	Arch         string            `json:"arch"`
	Plugin       string            `json:"plugin"`
	Source       string            `json:"source"`
	Scope        string            `json:"scope"`
	Collector    string            `json:"collector"`
	Tags         map[string]string `json:"tags"`
	Labels       map[string]string `json:"labels"`
	GlobalLabels map[string]string `json:"global_labels"`
	Metadata     map[string]string `json:"metadata"`
	UnixTime     int64             `json:"unixtime"`
	Timestamp    int64             `json:"timestamp"`
	RemoteAddr   string            `json:"-"`
}

func CategrafN9EHeartbeat(c *gin.Context) {
	if !validCategrafReceiverSource(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent receiver token"})
		return
	}
	payload, err := readCategrafHeartbeat(c)
	if err != nil {
		writeCategrafReceiverError(c, err)
		return
	}
	heartbeat := payload.toFindXAgentHeartbeat()
	agent, target, err := store.UpsertFindXAgentHeartbeat(heartbeat)
	if err != nil {
		if strings.Contains(err.Error(), "future") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "heartbeat time is too far in the future"})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent heartbeat persistence unavailable"})
		return
	}
	evidence := payload.toEvidence(agent.ID, target.ID)
	saved, err := store.SaveFindXAgentDataArrivalEvidence(evidence)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "heartbeat evidence persistence unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "status": saved.Status, "agent_id": agent.ID, "target_id": target.ID, "evidence_id": saved.ID})
}

func CategrafPrometheusRemoteWrite(c *gin.Context) {
	if !validCategrafReceiverSource(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent receiver token"})
		return
	}
	body, err := readLimitedReceiverBody(c.Request.Body, categrafReceiverBodyLimit)
	if err != nil {
		writeCategrafReceiverError(c, err)
		return
	}
	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "remote write body is required"})
		return
	}
	metadata := remoteWriteEvidenceMetadata(c, len(body))
	evidence := model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindMetrics,
		AgentID:      sanitizeReceiverValue("agent_id", firstReceiverValue(c.Query("agent_id"), c.GetHeader("X-FindX-Agent-Id"))),
		TargetID:     sanitizeReceiverValue("target_id", firstReceiverValue(c.Query("target_id"), c.GetHeader("X-FindX-Target-Id"))),
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/metrics/write-compatible"},
		Metadata:     metadata,
	}
	saved, err := store.SaveFindXAgentDataArrivalEvidence(evidence)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metrics evidence persistence unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "status": saved.Status, "evidence_id": saved.ID})
}


func readCategrafHeartbeat(c *gin.Context) (categrafHeartbeatPayload, error) {
	body, err := readReceiverBody(c.Request.Body, c.GetHeader("Content-Encoding"), categrafReceiverBodyLimit)
	if err != nil {
		return categrafHeartbeatPayload{}, err
	}
	var payload categrafHeartbeatPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return categrafHeartbeatPayload{}, errBadRequest("invalid heartbeat payload")
	}
	payload.RemoteAddr = clientHost(c.Request.RemoteAddr)
	if err := validateCategrafHeartbeat(payload); err != nil {
		return categrafHeartbeatPayload{}, err
	}
	return payload, nil
}

func validateCategrafHeartbeat(payload categrafHeartbeatPayload) error {
	if strings.TrimSpace(payload.ident()) == "" && strings.TrimSpace(payload.agentIP()) == "" && strings.TrimSpace(payload.Hostname) == "" && strings.TrimSpace(payload.Host) == "" {
		return errBadRequest("agent, host, hostname, or ip is required")
	}
	if strings.TrimSpace(payload.agentIP()) != "" {
		if _, ok := cleanIP(payload.agentIP()); !ok {
			return errBadRequest("valid ip is required")
		}
	}
	return nil
}

func (p categrafHeartbeatPayload) toFindXAgentHeartbeat() model.FindXAgentHeartbeat {
	return model.FindXAgentHeartbeat{
		Ident:        firstReceiverValue(p.ident(), p.Host, p.Hostname, p.IP),
		IP:           firstReceiverValue(p.agentIP(), p.RemoteAddr),
		Hostname:     firstReceiverValue(p.Hostname, p.Host),
		OS:           sanitizeReceiverValue("os", p.OS),
		Arch:         sanitizeReceiverValue("arch", p.Arch),
		Version:      sanitizeReceiverValue("version", firstReceiverValue(p.Version, p.AgentVersion)),
		Collector:    "findx-agent-host-collector",
		Capabilities: []string{"metrics", "heartbeat"},
		GlobalLabels: safeReceiverMetadata(p.Labels, p.Tags, p.GlobalLabels, p.Metadata),
		UnixTime:     firstReceiverUnixTime(p.UnixTime, p.Timestamp),
	}
}

func (p categrafHeartbeatPayload) toEvidence(agentID, targetID string) model.FindXAgentDataArrivalEvidence {
	metadata := safeReceiverMetadata(p.Labels, p.Tags, p.GlobalLabels, p.Metadata)
	for key, value := range map[string]string{
		"agent":     p.ident(),
		"host":      firstReceiverValue(p.Host, p.Hostname),
		"hostname":  firstReceiverValue(p.Hostname, p.Host),
		"ip":        firstReceiverValue(p.agentIP(), p.RemoteAddr),
		"os":        p.OS,
		"arch":      p.Arch,
		"version":   firstReceiverValue(p.Version, p.AgentVersion),
		"plugin":    p.Plugin,
		"source":    "findx-agent-compatible",
		"scope":     p.Scope,
		"collector": "findx-agent-host-collector",
	} {
		if clean := sanitizeReceiverValue(key, value); clean != "" {
			metadata[key] = clean
		}
	}
	return model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindHeartbeat,
		AgentID:      agentID,
		TargetID:     targetID,
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/agent/heartbeat-compatible"},
		Metadata:     metadata,
	}
}

func (p categrafHeartbeatPayload) ident() string {
	return firstReceiverValue(p.Ident, p.AgentID, p.Agent)
}

func (p categrafHeartbeatPayload) agentIP() string {
	return firstReceiverValue(p.IP, p.HostIP)
}
