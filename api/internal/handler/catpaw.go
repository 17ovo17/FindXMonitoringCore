package handler

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func cleanIP(value string) (string, bool) {
	ip := strings.TrimSpace(value)
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.String() != ip {
		return "", false
	}
	if parsed.IsUnspecified() || parsed.IsMulticast() || parsed.IsLinkLocalUnicast() || parsed.IsLinkLocalMulticast() {
		return "", false
	}
	return ip, true
}

func CatpawHeartbeat(c *gin.Context) {
	var a model.CatpawAgent
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if ip, ok := cleanIP(a.IP); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid ip is required"})
		return
	} else {
		a.IP = ip
	}
	a.LastSeen = time.Now()
	store.UpsertAgent(&a)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func CatpawReport(c *gin.Context) {
	var body struct {
		IP     string `json:"ip"`
		Report string `json:"report"`
		Title  string `json:"title"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if ip, ok := cleanIP(body.IP); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid ip is required"})
		return
	} else {
		body.IP = ip
	}
	summary := summarizeCatpawReport(body.Report)
	now := time.Now()
	rec := &model.DiagnoseRecord{
		ID:            fmt.Sprintf("%d", now.UnixNano()),
		TargetIP:      body.IP,
		Trigger:       "catpaw",
		Source:        "catpaw",
		Status:        model.StatusDone,
		Report:        summary,
		SummaryReport: summary,
		RawReport:     body.Report,
		AlertTitle:    body.Title,
		CreateTime:    now,
		EndTime:       &now,
	}
	store.AddRecord(rec)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListAgents(c *gin.Context) {
	c.JSON(http.StatusOK, store.ListAgents())
}

func DeleteAgent(c *gin.Context) {
	ip, ok := cleanIP(c.Param("ip"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid ip is required"})
		return
	}
	store.DeleteAgent(ip)
	auditEvent(c, "catpaw.agent.delete", ip, "L2", "allow", "agent record removed by user confirmation", c.Query("test_batch_id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteDiagnose(c *gin.Context) {
	store.DeleteRecord(c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func CleanupDiagnose(c *gin.Context) {
	scope := strings.TrimSpace(c.Query("scope"))
	batchID := strings.TrimSpace(c.Query("test_batch_id"))
	businessID := strings.TrimSpace(c.Query("business_id"))
	if scope == "" && batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope or test_batch_id is required"})
		return
	}
	matchesBusiness := func(r *model.DiagnoseRecord) bool {
		if businessID == "" {
			return true
		}
		joined := strings.Join([]string{r.TargetIP, r.AlertTitle, r.RawReport, r.Report, r.SummaryReport}, " ")
		return strings.Contains(joined, businessID)
	}
	deleted := store.DeleteRecordsByFilter(func(r *model.DiagnoseRecord) bool {
		if batchID != "" {
			joined := strings.Join([]string{r.ID, r.TargetIP, r.Trigger, r.Source, r.DataSource, r.AlertTitle, r.RawReport, r.Report}, " ")
			return strings.Contains(joined, batchID) && matchesBusiness(r)
		}
		switch scope {
		case "business_inspection":
			isInspection := r.Trigger == "business_inspection" || r.Source == "business_inspection" || r.DataSource == "business_inspection"
			return isInspection && matchesBusiness(r)
		case "test":
			joined := strings.ToLower(strings.Join([]string{r.ID, r.TargetIP, r.Trigger, r.Source, r.DataSource, r.AlertTitle}, " "))
			return (strings.Contains(joined, "test") || strings.Contains(joined, "whitebox") || strings.Contains(joined, "aiw-")) && matchesBusiness(r)
		default:
			return false
		}
	})
	scopeLabel := scope
	if businessID != "" {
		scopeLabel = scopeLabel + ":" + businessID
	}
	auditEvent(c, "diagnose.cleanup", scopeLabel, "L3", "allow", fmt.Sprintf("deleted %d diagnose records", deleted), batchID)
	c.JSON(http.StatusOK, gin.H{"ok": true, "deleted": deleted, "scope": scope, "business_id": businessID})
}

func ListDiagnose(c *gin.Context) {
	records := store.ListRecords()
	if businessID := strings.TrimSpace(c.Query("business_id")); businessID != "" {
		filtered := []*model.DiagnoseRecord{}
		for _, r := range records {
			if strings.Contains(r.TargetIP, businessID) || strings.Contains(r.AlertTitle, businessID) || strings.Contains(r.RawReport, businessID) {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}
	if source := strings.TrimSpace(c.Query("source")); source != "" {
		filtered := []*model.DiagnoseRecord{}
		for _, r := range records {
			if r.Source == source || r.DataSource == source || r.Trigger == source {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}
	c.JSON(http.StatusOK, records)
}

func StartDiagnose(c *gin.Context) {
	var req struct {
		IP           string           `json:"ip" binding:"required"`
		Prompt       string           `json:"prompt"`
		CredentialID string           `json:"credential_id"`
		Credential   RemoteCredential `json:"credential"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if ip, ok := cleanIP(req.IP); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid ip is required"})
		return
	} else {
		req.IP = ip
	}
	now := time.Now()
	source := "prometheus"
	if store.HasOnlineAgent(req.IP) {
		source = "catpaw"
	}
	prompt := req.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("请对主机 %s 进行全面的健康诊断，分析 CPU、内存、磁盘、网络等关键指标，给出根因分析和处置建议。", req.IP)
	}
	rec := &model.DiagnoseRecord{
		ID:         fmt.Sprintf("%d", now.UnixNano()),
		TargetIP:   req.IP,
		Trigger:    "manual",
		Source:     source,
		DataSource: source,
		Status:     model.StatusPending,
		CreateTime: now,
	}
	store.AddRecord(rec)
	go RunDiagnoseWithOptions(rec, DiagnoseOptions{Prompt: prompt, CredentialID: req.CredentialID, Credential: req.Credential})
	c.JSON(http.StatusOK, gin.H{"id": rec.ID, "source": source})
}

// --- catpaw report helper: pure data conversion utilities ---

func asMap(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{}
}

func asList(value any) []map[string]any {
	if value == nil {
		return nil
	}
	if typed, ok := value.(map[string]any); ok {
		return []map[string]any{typed}
	}
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]any); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func numberField(row map[string]any, key string) float64 {
	if row == nil {
		return 0
	}
	return toFloat(row[key])
}

func maxNumberField(rows []map[string]any, key string) float64 {
	maxValue := 0.0
	for _, row := range rows {
		if value := numberField(row, key); value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func toFloat(value any) float64 {
	switch typed := value.(type) {
	case json.Number:
		v, _ := typed.Float64()
		return v
	case float64:
		return typed
	case int:
		return float64(typed)
	case string:
		v, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return v
	default:
		return 0
	}
}

func stateCount(tcpInfo map[string]any, state string) int {
	for _, row := range asList(tcpInfo["TCPStateSummary"]) {
		if strings.EqualFold(anyText(row["Name"]), state) {
			return int(toFloat(row["Count"]))
		}
	}
	return 0
}

func anyText(value any) string {
	switch typed := value.(type) {
	case nil:
		return "not collected"
	case json.Number:
		return typed.String()
	case string:
		return typed
	case bool:
		return strconv.FormatBool(typed)
	case []any, map[string]any:
		b, _ := json.Marshal(typed)
		return string(b)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func percentText(value float64) string {
	if value == 0 {
		return "not collected"
	}
	return fmt.Sprintf("%.2f%%", value)
}

func numberText(value float64) string {
	if value == 0 {
		return "not collected"
	}
	return fmt.Sprintf("%.2f", value)
}
