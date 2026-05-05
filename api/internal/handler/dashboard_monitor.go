package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const maxDashboardTitleLen = 120

func ListMonitorDashboards(c *gin.Context) {
	items, err := store.ListMonitorDashboards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dashboard list failed"})
		return
	}
	c.JSON(http.StatusOK, sanitizeDashboards(items))
}

func CreateMonitorDashboard(c *gin.Context) {
	saveMonitorDashboard(c, "")
}

func GetMonitorDashboard(c *gin.Context) {
	item, ok, err := store.GetMonitorDashboard(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dashboard get failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}
	c.JSON(http.StatusOK, sanitizeDashboard(*item))
}

func UpdateMonitorDashboard(c *gin.Context) {
	saveMonitorDashboard(c, c.Param("id"))
}

func DeleteMonitorDashboard(c *gin.Context) {
	ok, err := store.DeleteMonitorDashboard(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dashboard delete failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func CloneMonitorDashboard(c *gin.Context) {
	item, ok, err := store.CloneMonitorDashboard(c.Param("id"), requestActor(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dashboard clone failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}
	c.JSON(http.StatusOK, sanitizeDashboard(*item))
}

func ShareMonitorDashboard(c *gin.Context) {
	result, ok, err := store.ShareMonitorDashboard(c.Param("id"), requestActor(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dashboard share failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func saveMonitorDashboard(c *gin.Context, id string) {
	var payload model.MonitorDashboardInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dashboard payload"})
		return
	}
	item := monitorDashboardFromInput(payload)
	if id != "" {
		item.ID = id
	}
	if checks := validateMonitorDashboardPayload(&item); len(checks) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dashboard", "checks": checks})
		return
	}
	var (
		out *model.MonitorDashboard
		ok  = true
		err error
	)
	if id == "" {
		out, err = store.SaveMonitorDashboard(&item, requestActor(c))
	} else {
		out, ok, err = store.UpdateMonitorDashboard(&item, requestActor(c))
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dashboard save failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}
	c.JSON(http.StatusOK, sanitizeDashboard(*out))
}

func monitorDashboardFromInput(input model.MonitorDashboardInput) model.MonitorDashboard {
	return model.MonitorDashboard{
		Title:           input.Title,
		Description:     input.Description,
		WorkspaceID:     input.WorkspaceID,
		ResourceGroupID: input.ResourceGroupID,
		Tags:            input.Tags,
		Variables:       input.Variables,
		Panels:          input.Panels,
		Status:          input.Status,
	}
}

func validateMonitorDashboardPayload(item *model.MonitorDashboard) []string {
	checks := []string{}
	title := strings.TrimSpace(item.Title)
	if title == "" {
		checks = append(checks, "title is required")
	}
	if len([]rune(title)) > maxDashboardTitleLen {
		checks = append(checks, "title is too long")
	}
	if !validMonitorDashboardStatus(item.Status) {
		checks = append(checks, "status must be active, draft, or archived")
	}
	if !validDashboardJSON(item.Variables, true) {
		checks = append(checks, "variables must be a json object")
	}
	if !validDashboardJSON(item.Panels, false) {
		checks = append(checks, "panels must be a json array")
	}
	return checks
}

func validMonitorDashboardStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "", model.MonitorDashboardStatusActive, model.MonitorDashboardStatusDraft, model.MonitorDashboardStatusArchived:
		return true
	default:
		return false
	}
}

func validDashboardJSON(raw json.RawMessage, wantObject bool) bool {
	if len(raw) == 0 {
		return true
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	if wantObject {
		_, ok := value.(map[string]any)
		return ok
	}
	_, ok := value.([]any)
	return ok
}

func sanitizeDashboards(items []model.MonitorDashboard) []model.MonitorDashboard {
	out := make([]model.MonitorDashboard, 0, len(items))
	for _, item := range items {
		out = append(out, sanitizeDashboard(item))
	}
	return out
}

func sanitizeDashboard(item model.MonitorDashboard) model.MonitorDashboard {
	item.Variables = sanitizeJSONPayload(item.Variables)
	item.Panels = sanitizeJSONPayload(item.Panels)
	return item
}

func sanitizeJSONPayload(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return json.RawMessage(`null`)
	}
	sanitized := sanitizeJSONValue(value)
	data, err := json.Marshal(sanitized)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return data
}

func sanitizeJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for key, val := range typed {
			if dashboardSensitiveKey(key) {
				out[key] = "REDACTED"
				continue
			}
			out[key] = sanitizeJSONValue(val)
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, val := range typed {
			out = append(out, sanitizeJSONValue(val))
		}
		return out
	case string:
		return sanitizeDashboardString(typed)
	default:
		return typed
	}
}

func dashboardSensitiveKey(key string) bool {
	lower := strings.ToLower(strings.TrimSpace(key))
	for _, marker := range []string{"token", "cookie", "dsn", "password", "secret", "api_key", "apikey", "authorization"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func sanitizeDashboardString(value string) string {
	lower := strings.ToLower(value)
	if strings.Contains(lower, "://") && (strings.Contains(lower, "@") || dashboardSensitiveURL(lower)) {
		return "REDACTED_URL"
	}
	return value
}

func dashboardSensitiveURL(lower string) bool {
	for _, marker := range []string{"token=", "cookie=", "dsn=", "password=", "secret=", "api_key=", "apikey=", "authorization="} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
