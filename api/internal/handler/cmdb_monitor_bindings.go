package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// GetCmdbMonitorBindingsBlocked returns persisted bindings for a real CMDB instance.
func GetCmdbMonitorBindingsBlocked(c *gin.Context) {
	if cmdbMonitorBindingReceiptsPath(c) {
		GetCmdbMonitorBindingReceipts(c)
		return
	}
	if cmdbMonitorBindingAuditRequested(c) {
		GetCmdbMonitorBindingAudit(c)
		return
	}
	if cmdbMonitorBindingDetailPath(c) {
		GetCmdbMonitorBindingDetail(c)
		return
	}
	instanceID := cmdbMonitorBindingInstanceID(c)
	if instanceID != "" {
		if _, ok := store.GetCmdbInstance(instanceID); ok {
			bindings := store.ListCmdbMonitorBindings(instanceID)
			c.JSON(http.StatusOK, cmdbMonitorBindingsReadyEnvelope(instanceID, bindings))
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code":        0,
		"status":      "ready",
		"instance_id": instanceID,
		"bindings":    []gin.H{},
		"total":       0,
		"meta":        cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	})
}

// CreateCmdbMonitorBindingsBlocked persists safe binding references.
func CreateCmdbMonitorBindingsBlocked(c *gin.Context) {
	if cmdbMonitorBindingReceiptIngestionPath(c) {
		IngestCmdbMonitorBindingReceipt(c)
		return
	}
	var payload cmdbMonitorBindingPayload
	if err := c.ShouldBindJSON(&payload); err != nil || !payload.readyForPersist() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cmdb monitor binding requires hostid, templateid, cmdb_attr_id, and server_attr_id"})
		return
	}
	instanceID := firstNonEmpty(payload.InstanceID, cmdbMonitorBindingInstanceID(c))
	if _, ok := store.GetCmdbInstance(instanceID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb instance not found"})
		return
	}
	runtimeTemplate, ok := cmdbMonitorBindingRuntimeTemplate(payload.TemplateID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monitor template not found or not executable"})
		return
	}
	if _, ok := store.GetMonitorTarget(payload.HostID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monitor target host not found"})
		return
	}
	auditRef := "cmdb-monitor-binding-audit-" + store.NewID()
	binding := payload.toModel(instanceID, requestActor(c), auditRef)
	saved, err := store.SaveCmdbMonitorBinding(&binding)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb monitor binding store unavailable"})
		return
	}
	deliveryRequestRef, err := cmdbCreateMonitorBindingDeliveryRequest(*saved, payload, runtimeTemplate, requestActor(c))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb monitor binding delivery request unavailable"})
		return
	}
	receipts, err := cmdbCreateMonitorBindingReceipts(*saved, auditRef, requestActor(c), c.ClientIP(), deliveryRequestRef)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb monitor binding receipt unavailable"})
		return
	}
	if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
		ID:           auditRef,
		Actor:        requestActor(c),
		Action:       "cmdb.monitor_binding.save",
		ResourceType: "cmdb_monitor_binding",
		ResourceID:   instanceID,
		Scope:        "cmdb",
		Status:       "ok",
		ClientIP:     c.ClientIP(),
		Summary:      "CMDB monitor binding saved",
		Details: map[string]any{
			"binding_id":       saved.ID,
			"instance_id":      instanceID,
			"hostid":           saved.HostID,
			"templateid":       saved.TemplateID,
			"cmdb_attr_id":     saved.CmdbAttrID,
			"server_attr_id":   saved.ServerAttrID,
			"server_model_id":  saved.ServerModelID,
			"server_object_id": saved.ServerObjectID,
			"template_type":    runtimeTemplate.Type,
		},
	}); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cmdb monitor binding audit unavailable"})
		return
	}
	c.JSON(http.StatusOK, cmdbMonitorBindingWriteEnvelope(instanceID, *saved, receipts))
}

type cmdbMonitorBindingPayload struct {
	InstanceID       string `json:"instance_id"`
	Host             string `json:"host"`
	HostID           string `json:"hostid"`
	TemplateID       string `json:"templateid"`
	ServerObjectID   string `json:"server_object_id"`
	ServerPlatformID string `json:"server_platform_id"`
	CmdbObjectID     string `json:"cmdb_object_id"`
	Group            any    `json:"group"`
	Tags             any    `json:"tags"`
	ActiveStatus     string `json:"active_status"`
	HostType         string `json:"hosttype"`
	SubType          string `json:"subtype"`
	HostTypeLabel    string `json:"hosttypeLabel"`
	SubTypeLabel     string `json:"subtypeLabel"`
	CmdbAttrID       string `json:"cmdb_attr_id"`
	ServerAttrID     string `json:"server_attr_id"`
	ServerModelID    string `json:"server_model_id"`
	ServerModelName  string `json:"server_model_name"`
	Attr             string `json:"attr"`
	AttrStru         any    `json:"attr_stru"`
	Queue            string `json:"queue"`
}

func (p cmdbMonitorBindingPayload) readyForPersist() bool {
	return strings.TrimSpace(p.HostID) != "" &&
		strings.TrimSpace(p.TemplateID) != "" &&
		strings.TrimSpace(p.CmdbAttrID) != "" &&
		strings.TrimSpace(p.ServerAttrID) != ""
}

func (p cmdbMonitorBindingPayload) toModel(instanceID, actor, auditRef string) model.CmdbMonitorBinding {
	return model.CmdbMonitorBinding{
		InstanceID:       strings.TrimSpace(instanceID),
		Host:             strings.TrimSpace(p.Host),
		HostID:           strings.TrimSpace(p.HostID),
		TemplateID:       strings.TrimSpace(p.TemplateID),
		ServerObjectID:   strings.TrimSpace(p.ServerObjectID),
		ServerPlatformID: strings.TrimSpace(p.ServerPlatformID),
		CmdbObjectID:     strings.TrimSpace(firstNonEmpty(p.CmdbObjectID, instanceID)),
		GroupJSON:        cmdbMonitorBindingJSON(p.Group),
		TagsJSON:         cmdbMonitorBindingJSON(p.Tags),
		ActiveStatus:     strings.TrimSpace(p.ActiveStatus),
		HostType:         strings.TrimSpace(p.HostType),
		SubType:          strings.TrimSpace(p.SubType),
		HostTypeLabel:    strings.TrimSpace(p.HostTypeLabel),
		SubTypeLabel:     strings.TrimSpace(p.SubTypeLabel),
		CmdbAttrID:       strings.TrimSpace(p.CmdbAttrID),
		ServerAttrID:     strings.TrimSpace(p.ServerAttrID),
		ServerModelID:    strings.TrimSpace(p.ServerModelID),
		ServerModelName:  strings.TrimSpace(p.ServerModelName),
		Attr:             strings.TrimSpace(p.Attr),
		AttrStruJSON:     cmdbMonitorBindingJSON(p.AttrStru),
		Queue:            strings.TrimSpace(p.Queue),
		Creator:          actor,
		Updater:          actor,
		AuditRef:         auditRef,
	}
}

func cmdbMonitorBindingInstanceID(c *gin.Context) string {
	path := strings.Trim(c.Param("path"), "/")
	if path == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(path, "/")[0])
}

func cmdbMonitorBindingPathParts(c *gin.Context) []string {
	path := strings.Trim(c.Param("path"), "/")
	if path == "" {
		return nil
	}
	raw := strings.Split(path, "/")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func cmdbMonitorBindingJSON(value any) string {
	if value == nil {
		return "null"
	}
	clean := cmdbSanitizeBindingValue("", value)
	raw, err := json.Marshal(clean)
	if err != nil {
		return "null"
	}
	return string(raw)
}

func cmdbMonitorBindingsReadyEnvelope(instanceID string, bindings []model.CmdbMonitorBinding) gin.H {
	items := make([]gin.H, 0, len(bindings))
	for _, binding := range bindings {
		items = append(items, cmdbMonitorBindingDTO(binding))
	}
	return gin.H{
		"code":              0,
		"status":            "ready",
		"instance_id":       instanceID,
		"bindings":          items,
		"total":             len(items),
		"expected_schema":   cmdbMonitorBindingsReadBlockedContract()["expected_schema"],
		"field_matrix":      cmdbMonitorBindingReadyFieldMatrix(),
		"source_evidence":   cmdbMonitorBindingsReadBlockedContract()["source_evidence"],
		"missing_contracts": []string{"binding_rollback_contract"},
		"audit_context": gin.H{
			"scope":         "cmdb",
			"resource_type": "cmdb_monitor_binding",
			"resource_id":   instanceID,
			"log_source":    "findx_audit",
		},
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_monitor_binding",
			"resource_id":   instanceID,
			"action":        "cmdb.monitor_binding.save",
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingWriteEnvelope(instanceID string, binding model.CmdbMonitorBinding, receipts []model.CmdbMonitorBindingReceipt) gin.H {
	return gin.H{
		"code":              0,
		"status":            "ready",
		"instance_id":       instanceID,
		"binding":           cmdbMonitorBindingDTOWithReceipts(binding, receipts),
		"audit_ref":         binding.AuditRef,
		"expected_schema":   cmdbMonitorBindingsWriteBlockedContract()["expected_schema"],
		"field_matrix":      cmdbMonitorBindingReadyFieldMatrix(),
		"source_evidence":   cmdbMonitorBindingsWriteBlockedContract()["source_evidence"],
		"missing_contracts": []string{"binding_rollback_contract", "cmdb_monitor_binding_delivery_receipt_contract"},
		"log_query": gin.H{
			"source":        "findx_audit",
			"scope":         "cmdb",
			"resource_type": "cmdb_monitor_binding",
			"resource_id":   instanceID,
			"action":        "cmdb.monitor_binding.save",
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingDTO(binding model.CmdbMonitorBinding) gin.H {
	return cmdbMonitorBindingDTOWithReceipts(binding, store.ListCmdbMonitorBindingReceipts(binding.ID))
}

func cmdbMonitorBindingDTOWithReceipts(binding model.CmdbMonitorBinding, receipts []model.CmdbMonitorBindingReceipt) gin.H {
	return gin.H{
		"id":                 binding.ID,
		"instance_id":        binding.InstanceID,
		"host":               binding.Host,
		"hostid":             binding.HostID,
		"templateid":         binding.TemplateID,
		"server_object_id":   binding.ServerObjectID,
		"server_platform_id": binding.ServerPlatformID,
		"cmdb_object_id":     binding.CmdbObjectID,
		"group":              cmdbMonitorBindingRawJSON(binding.GroupJSON),
		"tags":               cmdbMonitorBindingRawJSON(binding.TagsJSON),
		"active_status":      binding.ActiveStatus,
		"hosttype":           binding.HostType,
		"subtype":            binding.SubType,
		"hosttypeLabel":      binding.HostTypeLabel,
		"subtypeLabel":       binding.SubTypeLabel,
		"cmdb_attr_id":       binding.CmdbAttrID,
		"server_attr_id":     binding.ServerAttrID,
		"server_model_id":    binding.ServerModelID,
		"server_model_name":  binding.ServerModelName,
		"attr":               binding.Attr,
		"attr_stru":          cmdbMonitorBindingRawJSON(binding.AttrStruJSON),
		"queue":              binding.Queue,
		"audit_ref":          binding.AuditRef,
		"receipts":           cmdbMonitorBindingReceiptDTOs(receipts),
		"delivery_status":    cmdbMonitorBindingReceiptStatus(receipts, "delivery"),
		"effect_status":      cmdbMonitorBindingReceiptStatus(receipts, "effect"),
		"rollback_status":    cmdbMonitorBindingReceiptStatus(receipts, "rollback"),
		"created_at":         binding.CreatedAt,
		"updated_at":         binding.UpdatedAt,
	}
}

func cmdbMonitorBindingRawJSON(raw string) any {
	if strings.TrimSpace(raw) == "" || strings.TrimSpace(raw) == "null" {
		return nil
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil
	}
	return cmdbSanitizeBindingValue("", value)
}

func cmdbMonitorBindingReadyFieldMatrix() []gin.H {
	return []gin.H{
		cmdbContractFieldGroup("monitor_host_binding", "ready", []string{
			"host", "hostid", "server_object_id", "server_platform_id", "cmdb_object_id",
			"group", "tags", "active_status", "hosttype", "subtype", "hosttypeLabel", "subtypeLabel",
		}, "monitor_host_binding_contract"),
		cmdbContractFieldGroup("monitor_template", "runtime_validated", []string{
			"templateid", "server_model_id", "server_model_name",
		}, "monitor_template_runtime_contract"),
		cmdbContractFieldGroup("cmdb_monitor_field_mapping", "ready", []string{
			"cmdb_attr_id", "server_attr_id", "attr", "attr_stru", "queue",
		}, "cmdb_monitor_binding_field_mapping_contract"),
		cmdbContractFieldGroup("binding_audit_log", "ready", []string{
			"audit_ref", "scope", "resource_type", "resource_id", "action",
		}, "binding_audit_contract"),
		cmdbContractFieldGroup("binding_receipts", "blocked", []string{
			"delivery_status", "effect_status", "rollback_status", "receipts",
		}, "cmdb_monitor_binding_delivery_receipt_contract"),
	}
}

func cmdbMonitorBindingRuntimeTemplate(templateID string) (model.MonitoringBuiltinPayload, bool) {
	template, ok := store.GetMonitoringBuiltinPayload(strings.TrimSpace(templateID))
	if !ok {
		return model.MonitoringBuiltinPayload{}, false
	}
	switch strings.TrimSpace(template.Type) {
	case "collect", "alert":
		if cmdbMonitorBindingRuntimeContentReady(template.Content) {
			return template, true
		}
		return model.MonitoringBuiltinPayload{}, false
	default:
		return model.MonitoringBuiltinPayload{}, false
	}
}

func cmdbMonitorBindingRuntimeBlockedEnvelope(templateID string) gin.H {
	contract := cmdbMonitorBindingsWriteBlockedContract()
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  "CMDB monitor binding requires an existing monitor runtime template; dashboard/readme/instruction previews are not executable binding templates.",
		"reason":   "monitor template runtime contract is missing or not executable",
		"contract": "cmdb.monitor_template.runtime.v1",
		"missing_contracts": []string{
			"monitor_template_runtime_contract",
			"cmdb_monitor_template_runtime_content_contract",
			"cmdb_monitor_binding_delivery_receipt_contract",
			"binding_rollback_contract",
		},
		"templateid":      strings.TrimSpace(templateID),
		"expected_schema": contract["expected_schema"],
		"field_matrix": []gin.H{
			cmdbContractFieldGroup("monitor_template", "blocked", []string{
				"templateid", "component_id", "type", "content", "updated_by",
			}, "monitor_template_runtime_contract"),
			cmdbContractFieldGroup("monitor_template_runtime_content", "blocked", []string{
				"runtime", "executor_ref", "config_snippet_ref", "plugin_id", "rule", "expr",
			}, "cmdb_monitor_template_runtime_content_contract"),
			cmdbContractFieldGroup("monitor_template_runtime_allowed_types", "ready", []string{
				"collect", "alert",
			}, "monitor_template_runtime_contract"),
		},
		"source_evidence": contract["source_evidence"],
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}

func cmdbMonitorBindingHostTargetBlockedEnvelope(hostID string) gin.H {
	contract := cmdbMonitorBindingsWriteBlockedContract()
	return gin.H{
		"code":     http.StatusConflict,
		"status":   "pending",
		"error":    "pending",
		"message":  "CMDB monitor binding requires hostid to match an existing monitor target; unknown hosts are not persisted as bindings.",
		"reason":   "monitor host binding contract is missing",
		"contract": "cmdb.monitor_binding.host_target.v1",
		"missing_contracts": []string{
			"monitor_host_binding_contract",
			"cmdb_monitor_binding_store",
			"cmdb_monitor_binding_delivery_receipt_contract",
			"binding_rollback_contract",
		},
		"hostid":          strings.TrimSpace(hostID),
		"expected_schema": contract["expected_schema"],
		"field_matrix": []gin.H{
			cmdbContractFieldGroup("monitor_host_binding", "blocked", []string{
				"hostid", "monitor_target.id", "monitor_target.ident", "monitor_target.ip", "monitor_target.status",
			}, "monitor_host_binding_contract"),
			cmdbContractFieldGroup("monitor_template", "runtime_validated", []string{
				"templateid", "collect", "alert",
			}, "monitor_template_runtime_contract"),
			cmdbContractFieldGroup("binding_audit_log", "blocked", []string{
				"audit_ref", "audit_action",
			}, "binding_audit_contract"),
		},
		"source_evidence": contract["source_evidence"],
		"safe_to_retry":   false,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
}
