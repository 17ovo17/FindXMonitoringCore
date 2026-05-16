package handler

import (
	"encoding/json"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func cmdbCreateRelationActionDeliveryRequest(item model.CmdbRelationActionRequest, actor string) (string, error) {
	return cmdbCreateRelationActionExecutionRequest(item, actor, "delivery", "cmdb relation action delivery executor is not enabled")
}

func cmdbCreateRelationActionEffectRequest(item model.CmdbRelationActionRequest, actor string) (string, error) {
	return cmdbCreateRelationActionExecutionRequest(item, actor, "effect", "cmdb relation action effect probe is not enabled")
}

func cmdbCreateRelationActionExecutionRequest(item model.CmdbRelationActionRequest, actor, receiptType, blocker string) (string, error) {
	metadata := map[string]string{
		"scope":                   "cmdb_relation_action",
		"cmdb_relation_action_id": item.ID,
		"cmdb_relation_action":    item.Action,
		"cmdb_receipt_type":       receiptType,
		"cmdb_instance_id":        item.InstanceID,
		"cmdb_node_id":            item.NodeID,
		"cmdb_object_id":          item.ObjectID,
		"cmdb_relation_id":        item.RelationID,
		"delivery_receipt_ref":    "cmdb-relation-action-delivery-receipt-contract",
		"effect_receipt_ref":      "cmdb-relation-action-effect-receipt-contract",
		"executor_ref":            "cmdb-relation-action-executor-contract",
		"idempotency_key":         "cmdb-relation-action-" + receiptType + "-" + item.ID,
		"requested_by":            actor,
	}
	if schema, ok := cmdbRelationSchemaFromAction(item); ok {
		for key, value := range cmdbRelationSchemaMetadata(schema) {
			metadata[key] = value
		}
		cmdbMergeStringMetadata(metadata, cmdbRelationActionTaskLogMetadata(item, receiptType, schema))
	} else {
		cmdbMergeStringMetadata(metadata, cmdbRelationActionTaskLogMetadata(item, receiptType, nil))
	}
	task := model.FindXAgentExecutionTask{
		Action:    "cmdb-relation-" + item.Action + "-" + receiptType,
		TargetIDs: []string{item.NodeID},
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

func cmdbRelationActionTaskLogMetadata(item model.CmdbRelationActionRequest, receiptType string, schema map[string]any) map[string]string {
	data := map[string]any{
		"is_only_update": 0,
		"action":         item.Action,
		"object_id":      item.ObjectID,
		"instance_id":    item.NodeID,
		"relation_id":    item.RelationID,
		"root_id":        item.InstanceID,
		"request_id":     item.ID,
		"receipt_type":   receiptType,
	}
	if schema != nil {
		data["relation_schema"] = schema
	}
	return map[string]string{
		"cmdb_task_log_contract":   "cmdb-task-log-status0-55274",
		"cmdb_task_log_status":     "pending",
		"cmdb_task_log_type":       "2",
		"cmdb_task_log_msg":        "cmdb relation action " + receiptType + " executor is not enabled",
		"cmdb_task_log_data":       cmdbRelationActionTaskLogJSON(data),
		"cmdb_task_log_queue":      "cmdb-relation-action-" + item.Action,
		"cmdb_model_id":            item.ObjectID,
		"cmdb_model_name":          firstNonEmptyRelationTaskLogValue(anyToString(schemaValue(schema, "left_object_name")), item.ObjectID),
		"cmdb_object_id":           item.NodeID,
		"cmdb_object_name":         firstNonEmptyRelationTaskLogValue(anyToString(schemaValue(schema, "left_name")), item.NodeID),
		"server_model_id":          firstNonEmptyRelationTaskLogValue(anyToString(schemaValue(schema, "right_object_id")), item.ObjectID),
		"server_model_name":        firstNonEmptyRelationTaskLogValue(anyToString(schemaValue(schema, "right_object_name")), item.ObjectID),
		"server_object_id":         item.InstanceID,
		"server_object_name":       firstNonEmptyRelationTaskLogValue(anyToString(schemaValue(schema, "right_name")), item.InstanceID),
		"task_id_ref":              "cmdb-relation-action-" + receiptType + "-" + item.ID,
		"task_log_source_evidence": "cmdb-task-log-status0-55274.json",
	}
}

func cmdbRelationActionTaskLogJSON(value map[string]any) string {
	raw, err := json.Marshal(cmdbSanitizeBindingValue("", value))
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func schemaValue(schema map[string]any, key string) any {
	if schema == nil {
		return nil
	}
	return schema[key]
}

func firstNonEmptyRelationTaskLogValue(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
