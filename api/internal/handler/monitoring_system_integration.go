package handler

import (
	"errors"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const maxSystemIntegrationNameLen = 120

func ListMonitoringSystemIntegrations(c *gin.Context) {
	filter := model.MonitoringSystemIntegrationFilter{
		Query:      strings.TrimSpace(c.Query("query")),
		Status:     strings.TrimSpace(c.Query("status")),
		Visibility: strings.TrimSpace(c.Query("visibility")),
	}
	items, err := store.ListMonitoringSystemIntegrations(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "system integration list failed"})
		return
	}
	c.JSON(http.StatusOK, sanitizeSystemIntegrations(items))
}

func GetMonitoringSystemIntegration(c *gin.Context) {
	item, ok, err := store.GetMonitoringSystemIntegration(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "system integration get failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "system integration not found"})
		return
	}
	c.JSON(http.StatusOK, sanitizeSystemIntegration(item))
}

func CreateMonitoringSystemIntegration(c *gin.Context) {
	saveMonitoringSystemIntegration(c, "")
}

func UpdateMonitoringSystemIntegration(c *gin.Context) {
	saveMonitoringSystemIntegration(c, c.Param("id"))
}

func DeleteMonitoringSystemIntegration(c *gin.Context) {
	ok, err := store.DeleteMonitoringSystemIntegration(c.Param("id"))
	if errors.Is(err, store.ErrMonitoringSystemIntegrationBuiltin) {
		c.JSON(http.StatusConflict, gin.H{"error": "built-in system integration cannot be deleted"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "system integration delete failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "system integration not found"})
		return
	}
	auditSystemIntegrationWrite(c, "monitor.integration.delete", c.Param("id"), "")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UpdateMonitoringSystemIntegrationWeights(c *gin.Context) {
	var payload []model.MonitoringSystemIntegrationWeightInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid system integration weights payload"})
		return
	}
	if checks := validateSystemIntegrationWeights(payload); len(checks) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid system integration weights", "checks": checks})
		return
	}
	items, ok, err := store.UpdateMonitoringSystemIntegrationWeights(payload, requestActor(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "system integration weights update failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown system integration id"})
		return
	}
	auditSystemIntegrationWrite(c, "monitor.integration.sort", "weights", "")
	c.JSON(http.StatusOK, sanitizeSystemIntegrations(items))
}

func SetMonitoringSystemIntegrationHide(c *gin.Context) {
	var payload model.MonitoringSystemIntegrationHideInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid system integration hide payload"})
		return
	}
	item, ok, err := store.SetMonitoringSystemIntegrationHide(c.Param("id"), payload.Hide, requestActor(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "system integration hide update failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "system integration not found"})
		return
	}
	auditSystemIntegrationWrite(c, "monitor.integration.hide", item.ID, "hide="+boolAuditValue(payload.Hide))
	c.JSON(http.StatusOK, sanitizeSystemIntegration(item))
}

func saveMonitoringSystemIntegration(c *gin.Context, id string) {
	var payload model.MonitoringSystemIntegrationInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid system integration payload"})
		return
	}
	item := systemIntegrationFromInput(payload)
	if id != "" {
		item.ID = id
	}
	if checks := validateSystemIntegrationPayload(item, id == ""); len(checks) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid system integration", "checks": checks})
		return
	}
	var (
		out model.MonitoringSystemIntegration
		ok  = true
		err error
	)
	if id == "" {
		out, err = store.SaveMonitoringSystemIntegration(item, requestActor(c))
	} else {
		out, ok, err = store.UpdateMonitoringSystemIntegration(item, requestActor(c))
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "system integration save failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "system integration not found"})
		return
	}
	action := "monitor.integration.create"
	if id != "" {
		action = "monitor.integration.update"
	}
	auditSystemIntegrationWrite(c, action, out.ID, "name="+safeAuditSystemIntegrationText(out.Name))
	c.JSON(http.StatusOK, sanitizeSystemIntegration(out))
}

func systemIntegrationFromInput(input model.MonitoringSystemIntegrationInput) model.MonitoringSystemIntegration {
	return model.MonitoringSystemIntegration{
		ID:            strings.TrimSpace(input.ID),
		Name:          input.Name,
		URL:           input.URL,
		ConfigPreview: input.ConfigPreview,
		IsPrivate:     input.IsPrivate,
		TeamIDs:       input.TeamIDs,
		Weight:        input.Weight,
		Hide:          input.Hide,
	}
}

func validateSystemIntegrationPayload(item model.MonitoringSystemIntegration, allowOptionalID bool) []string {
	checks := []string{}
	if item.ID != "" && !validSystemIntegrationID(item.ID) {
		checks = append(checks, "id contains unsupported characters")
	}
	if !allowOptionalID && strings.TrimSpace(item.ID) == "" {
		checks = append(checks, "id is required")
	}
	name := strings.TrimSpace(item.Name)
	if name == "" {
		checks = append(checks, "name is required")
	}
	if len([]rune(name)) > maxSystemIntegrationNameLen {
		checks = append(checks, "name is too long")
	}
	if unsafeSystemIntegrationText(name) {
		checks = append(checks, "name contains blocked content")
	}
	if !validSystemIntegrationRoute(item.URL) {
		checks = append(checks, "url must be a safe FindX route")
	}
	if !validSystemIntegrationRoute(item.ConfigPreview) {
		checks = append(checks, "config_preview must be a safe FindX route")
	}
	if !validSystemIntegrationTeamIDs(item.TeamIDs) {
		checks = append(checks, "team_ids must contain unique positive ids")
	}
	return checks
}

func validateSystemIntegrationWeights(inputs []model.MonitoringSystemIntegrationWeightInput) []string {
	checks := []string{}
	if len(inputs) == 0 {
		return []string{"weights payload is empty"}
	}
	seen := map[string]bool{}
	for _, input := range inputs {
		id := strings.TrimSpace(input.ID)
		if !validSystemIntegrationID(id) {
			checks = append(checks, "id contains unsupported characters")
			continue
		}
		if seen[id] {
			checks = append(checks, "duplicate id")
		}
		seen[id] = true
	}
	return checks
}

func validSystemIntegrationID(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" || len(id) > 64 {
		return false
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == ':' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func validSystemIntegrationRoute(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 500 {
		return false
	}
	if !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return false
	}
	return !unsafeSystemIntegrationText(value)
}

func validSystemIntegrationTeamIDs(ids []int) bool {
	seen := map[int]bool{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}

func unsafeSystemIntegrationText(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"nightingale", "n9e", "embeddedproduct", "embedded-product", "categraf", "skywalking", "signoz",
		"findx.local", "token=", "password", "api_key", "apikey", "secret", "cookie", "mysql://", "postgres://",
		"authorization=", "private key",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func sanitizeSystemIntegrations(items []model.MonitoringSystemIntegration) []model.MonitoringSystemIntegration {
	out := make([]model.MonitoringSystemIntegration, 0, len(items))
	for _, item := range items {
		out = append(out, sanitizeSystemIntegration(item))
	}
	return out
}

func sanitizeSystemIntegration(item model.MonitoringSystemIntegration) model.MonitoringSystemIntegration {
	item.Name = sanitizeSystemIntegrationString(item.Name)
	item.URL = sanitizeSystemIntegrationString(item.URL)
	item.ConfigPreview = sanitizeSystemIntegrationString(item.ConfigPreview)
	item.CreateBy = safeAuditSystemIntegrationText(item.CreateBy)
	item.UpdateBy = safeAuditSystemIntegrationText(item.UpdateBy)
	item.TeamIDs = append([]int{}, item.TeamIDs...)
	item.BlockedActions = append([]model.MonitoringSystemIntegrationBlockedAction{}, item.BlockedActions...)
	return item
}

func sanitizeSystemIntegrationString(value string) string {
	if unsafeSystemIntegrationText(value) {
		return "REDACTED"
	}
	return strings.TrimSpace(value)
}

func safeAuditSystemIntegrationText(value string) string {
	if unsafeSystemIntegrationText(value) {
		return "REDACTED"
	}
	return strings.TrimSpace(value)
}

func auditSystemIntegrationWrite(c *gin.Context, action, target, detail string) {
	auditEvent(c, action, safeAuditSystemIntegrationText(target), "medium", "ok", safeAuditSystemIntegrationText(detail), c.GetHeader("X-Test-Batch-Id"))
}

func boolAuditValue(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
