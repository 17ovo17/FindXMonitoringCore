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

func ListMonitorDashboardTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, sanitizeDashboardTemplates(store.ListMonitorDashboardTemplates()))
}

func GetMonitorDashboardTemplate(c *gin.Context) {
	item, ok := store.GetMonitorDashboardTemplate(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard template not found"})
		return
	}
	c.JSON(http.StatusOK, sanitizeDashboardTemplate(*item))
}

func ImportMonitorDashboardTemplate(c *gin.Context) {
	tpl, ok := store.GetMonitorDashboardTemplate(c.Param("id"))
	if !ok {
		cmdbDashboardImportLookupBlocked(c, c.Param("id"))
		return
	}
	var payload model.MonitorDashboardTemplateImportInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dashboard template import payload"})
		return
	}
	item, checks := dashboardFromTemplate(*tpl, payload)
	if len(checks) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dashboard template import", "checks": checks})
		return
	}
	existing, _, err := store.FindMonitorDashboardByTitleWorkspaceResourceGroup(item.Title, item.WorkspaceID, item.ResourceGroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dashboard import dedupe failed"})
		return
	}
	cmdbDashboardImportBlocked(c, *tpl, *item, existing)
}

func cmdbDashboardImportLookupBlocked(c *gin.Context, templateID string) {
	tpl := model.MonitorDashboardTemplate{
		ID:    strings.TrimSpace(templateID),
		Title: strings.TrimSpace(templateID),
	}
	cmdbDashboardImportBlocked(c, tpl, model.MonitorDashboard{
		Title: strings.TrimSpace(templateID),
	}, nil)
}

func cmdbDashboardImportBlocked(c *gin.Context, tpl model.MonitorDashboardTemplate, item model.MonitorDashboard, existing *model.MonitorDashboard) {
	missing := []string{
		"cmdb_dashboard_template_lookup_contract",
		"cmdb_dashboard_import_runtime_contract",
		"cmdb_dashboard_import_executor_contract",
		"cmdb_dashboard_import_dedupe_contract",
		"cmdb_dashboard_import_batch_result_contract",
		"cmdb_dashboard_import_batch_receipt_contract",
		"cmdb_dashboard_import_conflict_rollback_contract",
		"cmdb_dashboard_import_rollback_receipt_contract",
	}
	if existing != nil {
		missing = append(missing, "cmdb_dashboard_import_dedup_contract")
	}
	auditQuery := strings.Join([]string{
		"scope=cmdb",
		"resource_type=cmdb_dashboard_import",
		"action=dashboard.template.import",
		"template_id=" + tpl.ID,
	}, "/")
	auditLog, auditErr := auditDashboardTemplateImportRequest(c, tpl, item, existing)
	if auditErr != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":          "DASHBOARD_IMPORT_AUDIT_UNAVAILABLE",
			"error":         "dashboard import audit unavailable",
			"safe_to_retry": false,
		})
		return
	}
	resp := gin.H{
		"code":              "pending",
		"status":            "pending",
		"contract_id":       "cmdb.dashboard.import.runtime.v1",
		"message":           "PENDING: dashboard import runtime, batch result and conflict rollback contracts are not open",
		"missing_contracts": missing,
		"blockers":          missing,
		"safe_to_retry":     false,
		"findx_audit_query": auditQuery,
		"receipt_contract":  cmdbDashboardImportReceiptContract(missing),
		"template": gin.H{
			"id":    tpl.ID,
			"title": tpl.Title,
		},
		"request_preview": gin.H{
			"title":             item.Title,
			"workspace_id":      item.WorkspaceID,
			"resource_group_id": item.ResourceGroupID,
			"tags":              item.Tags,
		},
		"audit_ref": auditLog.ID,
	}
	if existing != nil {
		resp["dedupe"] = gin.H{
			"reason": "existing_dashboard",
			"existing_dashboard": gin.H{
				"id":                existing.ID,
				"title":             existing.Title,
				"workspace_id":      existing.WorkspaceID,
				"resource_group_id": existing.ResourceGroupID,
			},
		}
	}
	c.JSON(http.StatusConflict, resp)
}

func cmdbDashboardImportReceiptContract(missing []string) gin.H {
	return gin.H{
		"contract_id":       "cmdb.dashboard.import.receipts.v1",
		"status":            "pending",
		"required_receipts": []string{"dedupe_receipt", "batch_result_receipt", "rollback_receipt"},
		"missing_contracts": uniquePackageRepositoryBlockers(missing),
		"safe_to_retry":     false,
	}
}

func auditDashboardTemplateImportRequest(c *gin.Context, tpl model.MonitorDashboardTemplate, item model.MonitorDashboard, existing *model.MonitorDashboard) (model.MonitorAuditLog, error) {
	details := map[string]any{
		"template_id":       tpl.ID,
		"requested_title":   item.Title,
		"workspace_id":      item.WorkspaceID,
		"resource_group_id": item.ResourceGroupID,
		"tag_count":         len(item.Tags),
		"blocked_contract":  "cmdb.dashboard.import.runtime.v1",
	}
	status := "blocked"
	summary := "dashboard template import blocked by contract"
	if existing != nil {
		details["dedupe"] = map[string]any{
			"reason":                "existing_dashboard",
			"existing_dashboard_id": existing.ID,
		}
		summary = "dashboard template import blocked by dedupe and contract"
	}
	return store.AddMonitorAuditLog(model.MonitorAuditLog{
		Actor:        requestActor(c),
		Action:       "dashboard.template.import",
		ResourceType: "cmdb_dashboard_import",
		ResourceID:   tpl.ID,
		Scope:        "cmdb",
		Status:       status,
		TraceID:      c.GetHeader("X-Test-Batch-Id"),
		ClientIP:     c.ClientIP(),
		Summary:      summary,
		Details:      details,
	})
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

func dashboardFromTemplate(tpl model.MonitorDashboardTemplate, input model.MonitorDashboardTemplateImportInput) (*model.MonitorDashboard, []string) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = tpl.Title
	}
	item := &model.MonitorDashboard{
		Title:           title,
		Description:     tpl.Description,
		WorkspaceID:     input.WorkspaceID,
		ResourceGroupID: input.ResourceGroupID,
		Tags:            templateImportTags(tpl.Tags, input.Tags),
		Variables:       mergeTemplateVariables(tpl.Variables, input.Variables),
		Panels:          append([]byte{}, tpl.Panels...),
		Status:          model.MonitorDashboardStatusActive,
	}
	checks := validateMonitorDashboardPayload(item)
	if strings.TrimSpace(item.ResourceGroupID) == "" {
		checks = append(checks, "resource_group_id is required")
	}
	return item, checks
}

func templateImportTags(defaults, input []string) []string {
	if input != nil {
		return input
	}
	return defaults
}

func mergeTemplateVariables(base, override json.RawMessage) json.RawMessage {
	if len(override) == 0 {
		return append([]byte{}, base...)
	}
	var baseObj, overrideObj map[string]any
	if err := json.Unmarshal(base, &baseObj); err != nil {
		return override
	}
	if err := json.Unmarshal(override, &overrideObj); err != nil {
		return override
	}
	for key, value := range overrideObj {
		baseObj[key] = value
	}
	out, err := json.Marshal(baseObj)
	if err != nil {
		return override
	}
	return out
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

func sanitizeDashboardTemplates(items []model.MonitorDashboardTemplate) []model.MonitorDashboardTemplate {
	out := make([]model.MonitorDashboardTemplate, 0, len(items))
	for _, item := range items {
		out = append(out, sanitizeDashboardTemplate(item))
	}
	return out
}

func sanitizeDashboardTemplate(item model.MonitorDashboardTemplate) model.MonitorDashboardTemplate {
	item.Tags = append([]string{}, item.Tags...)
	item.Variables = sanitizeJSONPayload(item.Variables)
	item.Panels = sanitizeJSONPayload(item.Panels)
	return item
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
	for _, marker := range []string{"token", "cookie", "dsn", "password", "secret", "api_key", "apikey", "authorization", "hash"} {
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
