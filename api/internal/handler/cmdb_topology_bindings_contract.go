package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetCmdbInstanceTopologyBlocked 明确阻断未接入的拓扑契约，避免返回假拓扑。
func GetCmdbInstanceTopologyBlocked(c *gin.Context) {
	c.JSON(http.StatusConflict, cmdbBlockedContractEnvelope(
		"cmdb.instance.topology.v1",
		[]string{
			"cmdb_relation_graph_contract",
			"cmdb_topology_view_contract",
			"cmdb_instance_relation_store",
			"cmdb_topology_field_mapping_contract",
		},
		cmdbInstanceTopologyBlockedContract(),
	))
}

// GetCmdbMonitorBindingsBlocked 明确阻断未接入的监控绑定查询契约。
func GetCmdbMonitorBindingsBlocked(c *gin.Context) {
	c.JSON(http.StatusConflict, cmdbBlockedContractEnvelope(
		"cmdb.monitor_bindings.read.v1",
		[]string{
			"monitor_host_binding_contract",
			"monitor_template_contract",
			"cmdb_monitor_binding_store",
			"cmdb_monitor_binding_field_mapping_contract",
		},
		cmdbMonitorBindingsReadBlockedContract(),
	))
}

// CreateCmdbMonitorBindingsBlocked 明确阻断未接入的监控绑定写入契约。
func CreateCmdbMonitorBindingsBlocked(c *gin.Context) {
	c.JSON(http.StatusConflict, cmdbBlockedContractEnvelope(
		"cmdb.monitor_bindings.write.v1",
		[]string{
			"monitor_host_binding_contract",
			"monitor_template_contract",
			"binding_audit_contract",
			"binding_rollback_contract",
			"cmdb_monitor_binding_write_receipt_contract",
		},
		cmdbMonitorBindingsWriteBlockedContract(),
	))
}

func cmdbInstanceTopologyBlockedContract() gin.H {
	return gin.H{
		"expected_schema": gin.H{
			"top_level": []string{
				"_id", "object_id", "name", "data", "default", "filter_empty", "status", "sort",
				"instance_id", "creator", "create_time", "business_status", "in_inst_detail", "business_names",
			},
			"data":             []string{"name", "instances"},
			"data.instances[]": []string{"object_id", "object_name", "object_type", "id", "name", "in", "out"},
			"relation_edge":    []string{"related_instance_id", "related_object_id", "instance_relation_id", "relation_id"},
			"children[]":       []string{"name", "instances", "children", "object", "relation", "level"},
			"object":           []string{"id", "name", "type"},
			"relation.in[]":    []string{"asst_id", "asst_name", "side", "position", "relation_id", "direction", "location"},
			"relation.out[]":   []string{"asst_id", "asst_name", "side", "position", "relation_id", "direction", "location"},
		},
		"field_matrix": []gin.H{
			cmdbContractFieldGroup("topology_document", []string{
				"_id", "object_id", "name", "data", "default", "filter_empty", "status", "sort",
				"instance_id", "creator", "create_time", "business_status", "in_inst_detail", "business_names",
			}, "cmdb_topology_view_contract"),
			cmdbContractFieldGroup("topology_instance", []string{
				"object_id", "object_name", "object_type", "id", "name", "in", "out",
			}, "cmdb_instance_relation_store"),
			cmdbContractFieldGroup("topology_relation_edge", []string{
				"related_instance_id", "related_object_id", "instance_relation_id", "relation_id",
			}, "cmdb_relation_graph_contract"),
			cmdbContractFieldGroup("topology_children", []string{
				"name", "instances", "children", "object", "relation", "level",
			}, "cmdb_topology_field_mapping_contract"),
			cmdbContractFieldGroup("topology_relation_descriptor", []string{
				"asst_id", "asst_name", "side", "position", "relation_id", "direction", "location",
			}, "cmdb_relation_graph_contract"),
		},
		"source_evidence": []gin.H{
			{
				"kind": "capture",
				"path": "D:\\测试\\LWOPS_安全测试资料_2026-05-10\\reverse-poc\\public\\captures\\cmdb-instance-default-topology-request.txt",
				"api":  "GET /backend_api/cmdb/instance/{object_id}/instance/{instance_id}/topology/id/{topology_id}",
			},
			{
				"kind": "capture",
				"path": "D:\\测试\\LWOPS_安全测试资料_2026-05-10\\reverse-poc\\public\\captures\\cmdb-instance-default-topology.json",
			},
		},
		"compatibility_notes": []string{
			"blocked response only exposes the source contract shape for React consumption",
			"未接入关系图和字段映射契约前，不合成图节点、图边、结果行或拓扑载荷",
		},
	}
}

func cmdbMonitorBindingsReadBlockedContract() gin.H {
	fields := []string{
		"host", "hostid", "templateid", "server_object_id", "server_platform_id", "cmdb_object_id",
		"group", "tags", "active_status", "hosttype", "subtype", "hosttypeLabel", "subtypeLabel",
		"cmdb_attr_id", "server_attr_id", "server_model_id", "server_model_name", "attr", "attr_stru", "queue",
	}
	return gin.H{
		"expected_schema": gin.H{
			"binding": fields,
			"attr[]":  []string{"cmdb_attr_id", "server_attr_id"},
			"attr_stru[]": []string{
				"cmdb_attr_id", "server_attr_id", "cmdb_attr_parent_id", "is_discovery", "macro",
			},
		},
		"field_matrix": []gin.H{
			cmdbContractFieldGroup("monitor_host_binding", []string{
				"host", "hostid", "server_object_id", "server_platform_id", "cmdb_object_id",
				"group", "tags", "active_status", "hosttype", "subtype", "hosttypeLabel", "subtypeLabel",
			}, "monitor_host_binding_contract"),
			cmdbContractFieldGroup("monitor_template", []string{
				"templateid", "server_model_id", "server_model_name",
			}, "monitor_template_contract"),
			cmdbContractFieldGroup("cmdb_monitor_field_mapping", []string{
				"cmdb_attr_id", "server_attr_id", "attr", "attr_stru", "queue",
			}, "cmdb_monitor_binding_field_mapping_contract"),
		},
		"source_evidence": []gin.H{
			{
				"kind": "capture",
				"path": "D:\\测试\\LWOPS_安全测试资料_2026-05-10\\reverse-poc\\public\\captures\\cmdb-task-log-status0-55274.json",
			},
			{
				"kind": "source",
				"path": "D:\\项目迁移文件\\平台源码\\AutoOps-main\\AutoOps-main\\web\\src\\api\\cmdb.js",
			},
			{
				"kind": "source",
				"path": "D:\\项目迁移文件\\平台源码\\AutoOps-main\\AutoOps-main\\api\\api\\monitor\\service\\monitorService.go",
			},
		},
		"compatibility_notes": []string{
			"blocked response only exposes monitor binding field expectations",
			"no host, template, queue, or binding rows are synthesized until read contracts and CMDB mapping store are available",
		},
	}
}

func cmdbMonitorBindingsWriteBlockedContract() gin.H {
	return gin.H{
		"expected_schema": gin.H{
			"request": []string{
				"host", "hostid", "templateid", "server_object_id", "server_platform_id", "cmdb_object_id",
				"group", "tags", "active_status", "hosttype", "subtype", "cmdb_attr_id", "server_attr_id",
				"server_model_id", "server_model_name", "attr", "attr_stru", "queue",
			},
			"receipt": []string{
				"binding_id", "audit_ref", "rollback_ref", "idempotency_key", "created_at", "operator",
			},
		},
		"field_matrix": []gin.H{
			cmdbContractFieldGroup("write_binding_contract", []string{
				"host", "hostid", "templateid", "server_object_id", "server_platform_id", "cmdb_object_id",
				"group", "tags", "active_status", "hosttype", "subtype",
			}, "monitor_host_binding_contract"),
			cmdbContractFieldGroup("write_mapping_contract", []string{
				"cmdb_attr_id", "server_attr_id", "server_model_id", "server_model_name", "attr", "attr_stru", "queue",
			}, "cmdb_monitor_binding_write_receipt_contract"),
			cmdbContractFieldGroup("write_safety_contract", []string{
				"audit_ref", "rollback_ref", "idempotency_key",
			}, "binding_audit_contract"),
		},
		"source_evidence": []gin.H{
			{
				"kind": "capture",
				"path": "D:\\测试\\LWOPS_安全测试资料_2026-05-10\\reverse-poc\\public\\captures\\cmdb-task-log-status0-55274.json",
			},
			{
				"kind": "source",
				"path": "D:\\项目迁移文件\\平台源码\\AutoOps-main\\AutoOps-main\\web\\src\\api\\cmdb.js",
			},
		},
		"compatibility_notes": []string{
			"request bodies are not echoed to avoid leaking credentials or sensitive binding values",
			"未接入写入回执契约前，不合成任何执行中、已应用、已安装、已到达、已回滚或已卸载状态",
		},
	}
}

func cmdbContractFieldGroup(name string, fields []string, missingContract string) gin.H {
	return gin.H{
		"name":             name,
		"fields":           fields,
		"missing_contract": missingContract,
		"status":           "blocked_by_contract",
	}
}
