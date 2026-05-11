package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

type findXAgentLogsPayload struct {
	AgentID  string                `json:"agent_id"`
	TargetID string                `json:"target_id"`
	Source   string                `json:"source"`
	Scope    string                `json:"scope"`
	Service  string                `json:"service"`
	TraceID  string                `json:"trace_id"`
	Records  []findXAgentLogRecord `json:"records"`
	Logs     []findXAgentLogRecord `json:"logs"`
	Metadata map[string]string     `json:"metadata"`
	Labels   map[string]string     `json:"labels"`
}

type findXAgentLogRecord struct {
	Body    string `json:"body"`
	Message string `json:"message"`
	Log     string `json:"log"`
}

type findXAgentTracesPayload struct {
	AgentID  string                  `json:"agent_id"`
	TargetID string                  `json:"target_id"`
	TraceID  string                  `json:"trace_id"`
	Source   string                  `json:"source"`
	Scope    string                  `json:"scope"`
	Service  string                  `json:"service"`
	Spans    []findXAgentSpanSummary `json:"spans"`
	Metadata map[string]string       `json:"metadata"`
	Labels   map[string]string       `json:"labels"`
}

type findXAgentSpanSummary struct {
	SpanID string `json:"span_id"`
}

func FindXAgentLogsCompatibleReceiver(c *gin.Context) {
	if !validCategrafReceiverSource(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent receiver token"})
		return
	}
	payload, err := readFindXAgentLogsPayload(c)
	if err != nil {
		writeCategrafReceiverError(c, err)
		return
	}
	saved, err := store.SaveFindXAgentDataArrivalEvidence(payload.toEvidence(c))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "logs evidence persistence unavailable"})
		return
	}
	writeFindXAgentReceiverOK(c, saved)
}

func FindXAgentTracesCompatibleReceiver(c *gin.Context) {
	if !validCategrafReceiverSource(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent receiver token"})
		return
	}
	payload, err := readFindXAgentTracesPayload(c)
	if err != nil {
		writeCategrafReceiverError(c, err)
		return
	}
	saved, err := store.SaveFindXAgentDataArrivalEvidence(payload.toEvidence(c))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "tracing evidence persistence unavailable"})
		return
	}
	writeFindXAgentReceiverOK(c, saved)
}

func readFindXAgentLogsPayload(c *gin.Context) (findXAgentLogsPayload, error) {
	body, err := readReceiverBody(c.Request.Body, c.GetHeader("Content-Encoding"), categrafReceiverBodyLimit)
	if err != nil {
		return findXAgentLogsPayload{}, err
	}
	if len(body) == 0 {
		return findXAgentLogsPayload{}, errBadRequest("logs payload body is required")
	}
	var payload findXAgentLogsPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return findXAgentLogsPayload{}, errBadRequest("invalid logs payload")
	}
	if err := validateFindXAgentLogsPayload(payload); err != nil {
		return findXAgentLogsPayload{}, err
	}
	return payload, nil
}

func readFindXAgentTracesPayload(c *gin.Context) (findXAgentTracesPayload, error) {
	body, err := readReceiverBody(c.Request.Body, c.GetHeader("Content-Encoding"), categrafReceiverBodyLimit)
	if err != nil {
		return findXAgentTracesPayload{}, err
	}
	if len(body) == 0 {
		return findXAgentTracesPayload{}, errBadRequest("traces payload body is required")
	}
	var payload findXAgentTracesPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return findXAgentTracesPayload{}, errBadRequest("invalid traces payload")
	}
	if err := validateFindXAgentTracesPayload(payload); err != nil {
		return findXAgentTracesPayload{}, err
	}
	return payload, nil
}

func validateFindXAgentLogsPayload(payload findXAgentLogsPayload) error {
	if firstReceiverValue(payload.AgentID, payload.TargetID) == "" {
		return errBadRequest("agent_id or target_id is required")
	}
	if findXAgentLogRecordCount(payload) == 0 {
		return errBadRequest("at least one log record is required")
	}
	for _, record := range append(payload.Records, payload.Logs...) {
		if firstReceiverValue(record.Body, record.Message, record.Log) != "" {
			return nil
		}
	}
	return errBadRequest("at least one log record body, message, or log is required")
}

func validateFindXAgentTracesPayload(payload findXAgentTracesPayload) error {
	if firstReceiverValue(payload.AgentID, payload.TargetID) == "" {
		return errBadRequest("agent_id or target_id is required")
	}
	if strings.TrimSpace(payload.TraceID) == "" {
		return errBadRequest("trace_id is required")
	}
	if len(payload.Spans) == 0 {
		return errBadRequest("at least one span is required")
	}
	for _, span := range payload.Spans {
		if strings.TrimSpace(span.SpanID) != "" {
			return nil
		}
	}
	return errBadRequest("at least one span_id is required")
}

func (p findXAgentLogsPayload) toEvidence(c *gin.Context) model.FindXAgentDataArrivalEvidence {
	return model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindLogs,
		AgentID:      sanitizeReceiverValue("agent_id", p.AgentID),
		TargetID:     sanitizeReceiverValue("target_id", p.TargetID),
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/findx-agent/logs-compatible"},
		Metadata:     p.metadata(c),
	}
}

func (p findXAgentTracesPayload) toEvidence(c *gin.Context) model.FindXAgentDataArrivalEvidence {
	return model.FindXAgentDataArrivalEvidence{
		Kind:         model.FindXAgentDataArrivalKindTracing,
		AgentID:      sanitizeReceiverValue("agent_id", p.AgentID),
		TargetID:     sanitizeReceiverValue("target_id", p.TargetID),
		Status:       model.FindXAgentDataArrivalStatusReported,
		EvidenceRefs: []string{"receiver:/findx-agent/traces-compatible"},
		Metadata:     p.metadata(c),
	}
}

func (p findXAgentLogsPayload) metadata(c *gin.Context) map[string]string {
	metadata := safeReceiverMetadata(p.Labels, p.Metadata)
	for key, value := range map[string]string{
		"count":     strconv.Itoa(findXAgentLogRecordCount(p)),
		"source":    firstReceiverValue(p.Source, "findx-agent-compatible"),
		"scope":     p.Scope,
		"service":   p.Service,
		"trace_id":  p.TraceID,
		"remote_ip": clientHost(c.Request.RemoteAddr),
	} {
		if clean := sanitizeReceiverValue(key, value); clean != "" {
			metadata[key] = clean
		}
	}
	return metadata
}

func (p findXAgentTracesPayload) metadata(c *gin.Context) map[string]string {
	metadata := safeReceiverMetadata(p.Labels, p.Metadata)
	for key, value := range map[string]string{
		"span_count": strconv.Itoa(len(p.Spans)),
		"trace_id":   p.TraceID,
		"source":     firstReceiverValue(p.Source, "findx-agent-compatible"),
		"scope":      p.Scope,
		"service":    p.Service,
		"remote_ip":  clientHost(c.Request.RemoteAddr),
	} {
		if clean := sanitizeReceiverValue(key, value); clean != "" {
			metadata[key] = clean
		}
	}
	return metadata
}

func findXAgentLogRecordCount(payload findXAgentLogsPayload) int {
	return len(payload.Records) + len(payload.Logs)
}

func writeFindXAgentReceiverOK(c *gin.Context, saved model.FindXAgentDataArrivalEvidence) {
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"status":      saved.Status,
		"evidence_id": saved.ID,
		"kind":        saved.Kind,
	})
}
