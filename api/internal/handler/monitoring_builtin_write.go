package handler

import (
	"errors"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const (
	maxBuiltinIdentLen   = 128
	maxBuiltinNameLen    = 255
	maxBuiltinCateLen    = 128
	maxBuiltinReadmeLen  = 20000
	maxBuiltinContentLen = 200000
)

var writableBuiltinPayloadTypes = map[string]bool{
	"dashboard": true,
	"collect":   true,
	"alert":     true,
}

func CreateMonitoringBuiltinComponents(c *gin.Context) {
	inputs, err := bindBuiltinComponentInputs(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid builtin component payload"})
		return
	}
	created := make([]model.MonitoringBuiltinComponent, 0, len(inputs))
	for _, input := range inputs {
		item := monitoringBuiltinComponentFromInput(input)
		if code, checks := validateBuiltinComponent(item, true); code != http.StatusOK {
			c.JSON(code, gin.H{"error": "invalid builtin component", "checks": checks})
			return
		}
		out, saveErr := store.SaveMonitoringBuiltinComponent(item, requestActor(c))
		if saveErr != nil {
			respondBuiltinStoreError(c, saveErr, "builtin component create failed")
			return
		}
		auditBuiltinWrite(c, "monitor.builtin.component.create", out.ID, "ident="+safeBuiltinAuditText(out.Ident))
		created = append(created, sanitizeBuiltinComponent(out))
	}
	c.JSON(http.StatusOK, created)
}

func UpdateMonitoringBuiltinComponent(c *gin.Context) {
	var input model.MonitoringBuiltinComponentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid builtin component payload"})
		return
	}
	item := monitoringBuiltinComponentFromInput(input)
	if code, checks := validateBuiltinComponent(item, false); code != http.StatusOK {
		c.JSON(code, gin.H{"error": "invalid builtin component", "checks": checks})
		return
	}
	out, ok, err := store.UpdateMonitoringBuiltinComponent(item, requestActor(c))
	if err != nil {
		respondBuiltinStoreError(c, err, "builtin component update failed")
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "builtin component not found"})
		return
	}
	auditBuiltinWrite(c, "monitor.builtin.component.update", out.ID, "ident="+safeBuiltinAuditText(out.Ident))
	c.JSON(http.StatusOK, sanitizeBuiltinComponent(out))
}

func DeleteMonitoringBuiltinComponents(c *gin.Context) {
	ids := parseBuiltinIDsPayload(c)
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids required"})
		return
	}
	ok, err := store.DeleteMonitoringBuiltinComponents(ids)
	if err != nil {
		respondBuiltinStoreError(c, err, "builtin component delete failed")
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "builtin component not found"})
		return
	}
	auditBuiltinWrite(c, "monitor.builtin.component.delete", strings.Join(ids, ","), "count="+intAuditValue(len(ids)))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func CreateMonitoringBuiltinPayloads(c *gin.Context) {
	inputs, err := bindBuiltinPayloadInputs(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid builtin payload"})
		return
	}
	created := make([]model.MonitoringBuiltinPayload, 0, len(inputs))
	for _, input := range inputs {
		item, checks := monitoringBuiltinPayloadFromInput(input)
		if len(checks) == 0 {
			checks = validateBuiltinPayload(item, true)
		}
		if len(checks) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid builtin payload", "checks": checks})
			return
		}
		out, saveErr := store.SaveMonitoringBuiltinPayload(item, requestActor(c))
		if saveErr != nil {
			respondBuiltinStoreError(c, saveErr, "builtin payload create failed")
			return
		}
		auditBuiltinWrite(c, "monitor.builtin.payload.create", out.ID, "type="+safeBuiltinAuditText(out.Type))
		created = append(created, sanitizeBuiltinPayload(out))
	}
	c.JSON(http.StatusOK, created)
}

func UpdateMonitoringBuiltinPayload(c *gin.Context) {
	var input model.MonitoringBuiltinPayloadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid builtin payload"})
		return
	}
	item, checks := monitoringBuiltinPayloadFromInput(input)
	if len(checks) == 0 {
		checks = validateBuiltinPayload(item, false)
	}
	if len(checks) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid builtin payload", "checks": checks})
		return
	}
	out, ok, err := store.UpdateMonitoringBuiltinPayload(item, requestActor(c))
	if err != nil {
		respondBuiltinStoreError(c, err, "builtin payload update failed")
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "builtin payload not found"})
		return
	}
	auditBuiltinWrite(c, "monitor.builtin.payload.update", out.ID, "type="+safeBuiltinAuditText(out.Type))
	c.JSON(http.StatusOK, sanitizeBuiltinPayload(out))
}

func DeleteMonitoringBuiltinPayloads(c *gin.Context) {
	ids := parseBuiltinIDsPayload(c)
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids required"})
		return
	}
	ok, err := store.DeleteMonitoringBuiltinPayloads(ids)
	if err != nil {
		respondBuiltinStoreError(c, err, "builtin payload delete failed")
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "builtin payload not found"})
		return
	}
	auditBuiltinWrite(c, "monitor.builtin.payload.delete", strings.Join(ids, ","), "count="+intAuditValue(len(ids)))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func monitoringBuiltinComponentFromInput(input model.MonitoringBuiltinComponentInput) model.MonitoringBuiltinComponent {
	name := strings.TrimSpace(input.Name)
	ident := strings.TrimSpace(input.Ident)
	if name == "" {
		name = ident
	}
	return model.MonitoringBuiltinComponent{
		ID:       strings.TrimSpace(input.ID),
		Ident:    ident,
		Name:     name,
		Logo:     strings.TrimSpace(input.Logo),
		Readme:   strings.TrimSpace(input.Readme),
		Disabled: input.Disabled,
	}
}

func monitoringBuiltinPayloadFromInput(input model.MonitoringBuiltinPayloadInput) (model.MonitoringBuiltinPayload, []string) {
	content, checks := normalizeBuiltinPayloadContent(input.Type, input.Content)
	item := model.MonitoringBuiltinPayload{
		ID:          strings.TrimSpace(input.ID),
		UUID:        strings.TrimSpace(input.UUID),
		Type:        strings.TrimSpace(input.Type),
		ComponentID: strings.TrimSpace(input.ComponentID),
		Cate:        strings.TrimSpace(input.Cate),
		Name:        strings.TrimSpace(input.Name),
		Title:       strings.TrimSpace(input.Name),
		Content:     content,
	}
	return item, checks
}

func parseBuiltinIDsPayload(c *gin.Context) []string {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil
	}
	seen := map[string]bool{}
	ids := []string{}
	for _, value := range req.IDs {
		id := strings.TrimSpace(value)
		if id != "" && validBuiltinID(id) && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	return ids
}

func builtinComponentExists(id string) bool {
	needle := strings.TrimSpace(id)
	if needle == "" {
		return false
	}
	for _, component := range store.ListMonitoringBuiltinComponents() {
		if component.ID == needle {
			return true
		}
	}
	return false
}

func respondBuiltinStoreError(c *gin.Context, err error, fallback string) {
	if errors.Is(err, store.ErrMonitoringBuiltinProtected) {
		c.JSON(http.StatusConflict, gin.H{"error": "built-in catalog row is protected"})
		return
	}
	if errors.Is(err, store.ErrMonitoringBuiltinComponentHasPayloads) {
		c.JSON(http.StatusConflict, gin.H{"error": "component still has payloads"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
}

func sanitizeBuiltinComponent(item model.MonitoringBuiltinComponent) model.MonitoringBuiltinComponent {
	item.Ident = safeBuiltinAuditText(item.Ident)
	item.Name = safeBuiltinAuditText(item.Name)
	item.Logo = safeBuiltinAuditText(item.Logo)
	item.Readme = safeBuiltinAuditText(item.Readme)
	return item
}

func sanitizeBuiltinPayload(item model.MonitoringBuiltinPayload) model.MonitoringBuiltinPayload {
	item.Type = safeBuiltinAuditText(item.Type)
	item.Cate = safeBuiltinAuditText(item.Cate)
	item.Name = safeBuiltinAuditText(item.Name)
	item.Title = safeBuiltinAuditText(item.Title)
	return item
}

func auditBuiltinWrite(c *gin.Context, action, target, detail string) {
	auditEvent(c, action, safeBuiltinAuditText(target), "medium", "ok", safeBuiltinAuditText(detail), c.GetHeader("X-Test-Batch-Id"))
}

func validBuiltinID(id string) bool {
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

func validBuiltinIdent(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxBuiltinIdentLen {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func safeBuiltinRoute(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= 500 && strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") && !unsafeBuiltinText(value)
}

func unsafeBuiltinText(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"nightingale", "n9e", "categraf", "catpaw", "skywalking", "signoz", "flashcat",
		"findx.local", "token=", "password=", "password:", "api_key", "apikey", "secret=", "secret:",
		"cookie:", "authorization:", "bearer ", "basic ", "mysql://", "postgres://", "private key",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func safeBuiltinAuditText(value string) string {
	if unsafeBuiltinText(value) {
		return "REDACTED"
	}
	return strings.TrimSpace(value)
}
