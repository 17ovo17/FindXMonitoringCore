package handler

import (
	"encoding/json"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func cmdbCreateMonitorBindingDeliveryRequest(binding model.CmdbMonitorBinding, payload cmdbMonitorBindingPayload, template model.MonitoringBuiltinPayload, actor string) (string, error) {
	metadata := map[string]string{
		"scope":                    configRolloutScopeCMDBHost,
		"cmdb_binding_id":          binding.ID,
		"cmdb_instance_id":         binding.InstanceID,
		"cmdb_hostid":              binding.HostID,
		"cmdb_templateid":          binding.TemplateID,
		"cmdb_attr_id":             binding.CmdbAttrID,
		"server_attr_id":           binding.ServerAttrID,
		"executor_ref":             cmdbMonitorBindingRuntimeStringField(template.Content, "executor_ref"),
		"config_snippet_ref":       cmdbMonitorBindingConfigSnippetRef(template),
		"plugin_config_writer_ref": "cmdb-monitor-binding-writer-contract",
		"reload_command_ref":       "cmdb-monitor-binding-reload-contract",
		"reload_receipt_ref":       "cmdb-monitor-binding-reload-receipt-contract",
		"rollback_receipt_ref":     "cmdb-monitor-binding-rollback-receipt-contract",
		"drift_check_ref":          "cmdb-monitor-binding-drift-contract",
		"evidence_chain_ref":       "cmdb-monitor-binding-evidence-chain-contract",
		"target_os":                firstNonEmpty(binding.ServerPlatformID, "linux"),
		"cmdb_receipt_type":        "delivery",
		"timeout_policy_ref":       "cmdb-monitor-binding-timeout-policy",
		"idempotency_key":          "cmdb-monitor-binding-" + binding.ID,
		"requested_by":             actor,
	}
	cmdbMergeStringMetadata(metadata, cmdbMonitorBindingTaskLogMetadata(binding, "delivery"))
	task := model.FindXAgentExecutionTask{
		Action:        "cmdb-monitor-binding-delivery",
		TargetIDs:     []string{binding.HostID},
		ConfigVersion: firstNonEmpty(payload.Queue, binding.ID),
		Status:        "blocked",
		Blocker:       "PENDING: cmdb monitor binding delivery executor is not enabled",
		Audit:         "findx_agent.task.requested",
		Metadata:      metadata,
	}
	saved, err := store.SaveFindXAgentExecutionTask(task)
	if err != nil {
		return "", err
	}
	return saved.ID, nil
}

func cmdbCreateMonitorBindingExecutionRequest(binding model.CmdbMonitorBinding, receiptType, actor string) (string, error) {
	blocker := "cmdb monitor binding " + receiptType + " executor is not enabled"
	if receiptType == "effect" {
		blocker = "cmdb monitor binding effect probe is not enabled"
	}
	metadata := map[string]string{
		"scope":             "cmdb_monitor_binding",
		"cmdb_binding_id":   binding.ID,
		"cmdb_instance_id":  binding.InstanceID,
		"cmdb_hostid":       binding.HostID,
		"cmdb_templateid":   binding.TemplateID,
		"cmdb_attr_id":      binding.CmdbAttrID,
		"server_attr_id":    binding.ServerAttrID,
		"cmdb_receipt_type": receiptType,
		"effect_probe_ref":  "cmdb-monitor-binding-effect-probe-contract",
		"rollback_ref":      "cmdb-monitor-binding-rollback-contract",
		"evidence_ref":      "cmdb-monitor-binding-evidence-chain-contract",
		"idempotency_key":   "cmdb-monitor-binding-" + receiptType + "-" + binding.ID,
		"requested_by":      actor,
	}
	cmdbMergeStringMetadata(metadata, cmdbMonitorBindingTaskLogMetadata(binding, receiptType))
	task := model.FindXAgentExecutionTask{
		Action:    "cmdb-monitor-binding-" + receiptType,
		TargetIDs: []string{binding.HostID},
		Status:    "blocked",
		Blocker:   "PENDING: " + blocker,
		Audit:     "findx_agent.task.requested",
		Metadata:  metadata,
	}
	saved, err := store.SaveFindXAgentExecutionTask(task)
	if err != nil {
		return "", err
	}
	return saved.ID, nil
}

func cmdbMonitorBindingConfigSnippetRef(template model.MonitoringBuiltinPayload) string {
	if value := cmdbMonitorBindingRuntimeStringField(template.Content, "config_snippet_ref"); value != "" {
		return value
	}
	return "cmdb-monitor-template-" + strings.TrimSpace(template.ID)
}

func cmdbMonitorBindingRuntimeStringField(content json.RawMessage, field string) string {
	var payload map[string]any
	if err := json.Unmarshal(content, &payload); err != nil {
		return ""
	}
	if value := cmdbMonitorRuntimeString(payload[field]); value != "" {
		return value
	}
	runtime, _ := payload["runtime"].(map[string]any)
	return cmdbMonitorRuntimeString(runtime[field])
}

func cmdbMonitorBindingTaskLogMetadata(binding model.CmdbMonitorBinding, receiptType string) map[string]string {
	attr := cmdbMonitorBindingTaskLogAttr(binding)
	attrStru := cmdbMonitorBindingTaskLogAttrStru(binding)
	data := map[string]any{
		"is_only_update": 0,
		"object_id":      binding.CmdbObjectID,
		"instance_id":    binding.InstanceID,
		"hostid":         binding.HostID,
		"templateid":     binding.TemplateID,
		"cmdb_attr_id":   binding.CmdbAttrID,
		"server_attr_id": binding.ServerAttrID,
		"queue":          binding.Queue,
	}
	if binding.Attr != "" {
		data["attr"] = binding.Attr
	}
	if attrStru != "" {
		data["attr_stru"] = attrStru
	}
	return map[string]string{
		"cmdb_task_log_contract":   "cmdb-task-log-status0-55274",
		"cmdb_task_log_status":     "pending",
		"cmdb_task_log_type":       "2",
		"cmdb_task_log_msg":        "cmdb monitor binding " + receiptType + " executor is not enabled",
		"cmdb_task_log_attr":       attr,
		"cmdb_task_log_attr_stru":  attrStru,
		"cmdb_task_log_data":       cmdbMonitorBindingTaskLogJSON(data),
		"cmdb_task_log_queue":      firstNonEmpty(binding.Queue, "cmdb-monitor-binding-"+binding.ID),
		"cmdb_model_id":            binding.CmdbObjectID,
		"cmdb_model_name":          firstNonEmpty(binding.SubTypeLabel, binding.HostTypeLabel, binding.CmdbObjectID),
		"cmdb_object_id":           binding.InstanceID,
		"cmdb_object_name":         firstNonEmpty(binding.Host, binding.HostID),
		"server_model_id":          binding.ServerModelID,
		"server_model_name":        binding.ServerModelName,
		"server_object_id":         binding.ServerObjectID,
		"server_object_name":       firstNonEmpty(binding.Host, binding.HostID),
		"task_id_ref":              "cmdb-monitor-binding-" + receiptType + "-" + binding.ID,
		"task_log_source_evidence": "cmdb-task-log-status0-55274.json",
	}
}

func cmdbMonitorBindingTaskLogAttr(binding model.CmdbMonitorBinding) string {
	if binding.CmdbAttrID == "" && binding.ServerAttrID == "" {
		return ""
	}
	return strings.TrimSpace(binding.CmdbAttrID) + "=>" + strings.TrimSpace(binding.ServerAttrID)
}

func cmdbMonitorBindingTaskLogAttrStru(binding model.CmdbMonitorBinding) string {
	var rows []map[string]any
	if err := json.Unmarshal([]byte(binding.AttrStruJSON), &rows); err != nil || len(rows) == 0 {
		return ""
	}
	parts := make([]string, 0, len(rows))
	for _, row := range rows {
		cmdbAttr := cmdbMonitorRuntimeString(row["cmdb_attr_id"])
		serverAttr := cmdbMonitorRuntimeString(row["server_attr_id"])
		if cmdbAttr == "" || serverAttr == "" {
			continue
		}
		item := cmdbAttr + "=>" + serverAttr
		if macro := cmdbMonitorRuntimeString(row["macro"]); macro != "" {
			item += "@" + macro
		}
		parts = append(parts, item)
	}
	return strings.Join(parts, ",")
}

func cmdbMonitorBindingTaskLogJSON(value map[string]any) string {
	raw, err := json.Marshal(cmdbSanitizeBindingValue("", value))
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func cmdbMergeStringMetadata(dst, src map[string]string) {
	for key, value := range src {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		dst[key] = value
	}
}
