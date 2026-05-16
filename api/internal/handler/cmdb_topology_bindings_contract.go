package handler

import (
	"net/http"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// GetCmdbInstanceTopologyBlocked blocks topology responses unless real relation rows can build a graph.
func GetCmdbInstanceTopologyBlocked(c *gin.Context) {
	instanceID := c.Param("id")
	if _, ok := store.GetCmdbInstance(instanceID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "cmdb instance not found"})
		return
	}
	if graph, ok := buildCmdbInstanceTopologyGraph(instanceID, c.Query("relation_query") == "execute"); ok {
		c.JSON(http.StatusOK, graph)
		return
	}
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
func cmdbInstanceTopologyBlockedContract() gin.H {
	return gin.H{
		"contract_gap_id": "FX-NIGHT-126C-CMDB-RELATION-GRAPH-TOPOLOGY-UI-CONTRACT",
		"contract_matrix": []gin.H{
			cmdbContractMatrixItem("topology_request_path", "ready", "cmdb.instance.topology.path.v1", []string{"object_id", "instance_id", "topology_id"}, "captured route shape is known; current FindX route only carries instance_id"),
			cmdbContractMatrixItem("topology_document_envelope", "blocked", "cmdb.topology.document.envelope.v1", []string{"_id", "object_id", "name", "data", "default", "filter_empty", "status", "sort", "instance_id", "creator", "create_time", "business_status", "in_inst_detail", "business_names"}, "source envelope is known but cannot be returned as code:0 without relation-backed data"),
			cmdbContractMatrixItem("topology_instance_identity", "ready", "cmdb.instance.identity.read.v1", []string{"object_id", "object_name", "object_type", "id", "name"}, "CmdbInstance and CmdbObject are readable through the current store"),
			cmdbContractMatrixItem("relation_path_levels", "missing_backend", "cmdb.relation.path.levels.v1", []string{"children", "level", "object", "relation", "instances"}, "no backend traversal contract maps relation paths into nested topology levels"),
			cmdbContractMatrixItem("relation_in_edges", "missing_relation_store", "cmdb.relation.in.edges.v1", []string{"in", "related_instance_id", "related_object_id", "instance_relation_id", "relation_id"}, "model has CmdbInstanceRelation but current store exposes no relation read API"),
			cmdbContractMatrixItem("relation_out_edges", "missing_relation_store", "cmdb.relation.out.edges.v1", []string{"out", "related_instance_id", "related_object_id", "instance_relation_id", "relation_id"}, "model has CmdbInstanceRelation but current store exposes no relation read API"),
			cmdbContractMatrixItem("relation_descriptor_matrix", "missing_datasource", "cmdb.relation.descriptor.matrix.v1", []string{"asst_id", "asst_name", "side", "position", "direction", "location"}, "FindX has no datasource for AutoOps relation descriptor semantics"),
			cmdbContractMatrixItem("topology_success_response", "unsafe", "cmdb.topology.success.response.v1", []string{"code:0", "data", "nodes", "edges"}, "returning success, nodes, or edges without a relation-backed graph would be fake data"),
		},
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
			cmdbContractFieldGroup("topology_document", "blocked", []string{
				"_id", "object_id", "name", "data", "default", "filter_empty", "status", "sort",
				"instance_id", "creator", "create_time", "business_status", "in_inst_detail", "business_names",
			}, "cmdb_topology_view_contract"),
			cmdbContractFieldGroup("topology_instance", "ready", []string{
				"object_id", "object_name", "object_type", "id", "name", "in", "out",
			}, "cmdb_instance_relation_store"),
			cmdbContractFieldGroup("topology_relation_edge", "missing_relation_store", []string{
				"related_instance_id", "related_object_id", "instance_relation_id", "relation_id",
			}, "cmdb_relation_graph_contract"),
			cmdbContractFieldGroup("topology_children", "missing_backend", []string{
				"name", "instances", "children", "object", "relation", "level",
			}, "cmdb_topology_field_mapping_contract"),
			cmdbContractFieldGroup("topology_relation_descriptor", "missing_datasource", []string{
				"asst_id", "asst_name", "side", "position", "relation_id", "direction", "location",
			}, "cmdb_relation_graph_contract"),
		},
		"source_evidence": []gin.H{
			{
				"kind": "capture",
				"path": "D:\\娴嬭瘯\\LWOPS_瀹夊叏娴嬭瘯璧勬枡_2026-05-10\\reverse-poc\\public\\captures\\cmdb-instance-default-topology-request.txt",
				"api":  "GET /backend_api/cmdb/instance/{object_id}/instance/{instance_id}/topology/id/{topology_id}",
			},
			{
				"kind": "capture",
				"path": "D:\\娴嬭瘯\\LWOPS_瀹夊叏娴嬭瘯璧勬枡_2026-05-10\\reverse-poc\\public\\captures\\cmdb-instance-default-topology.json",
			},
		},
		"compatibility_notes": []string{
			"blocked response only exposes the source contract shape for React consumption",
			"relation graph rows and field mapping must be verified before returning topology nodes or edges",
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
			cmdbContractFieldGroup("monitor_host_binding", "missing_datasource", []string{
				"host", "hostid", "server_object_id", "server_platform_id", "cmdb_object_id",
				"group", "tags", "active_status", "hosttype", "subtype", "hosttypeLabel", "subtypeLabel",
			}, "monitor_host_binding_contract"),
			cmdbContractFieldGroup("monitor_template", "missing_datasource", []string{
				"templateid", "server_model_id", "server_model_name",
			}, "monitor_template_contract"),
			cmdbContractFieldGroup("cmdb_monitor_field_mapping", "missing_backend", []string{
				"cmdb_attr_id", "server_attr_id", "attr", "attr_stru", "queue",
			}, "cmdb_monitor_binding_field_mapping_contract"),
		},
		"source_evidence": []gin.H{
			{
				"kind": "capture",
				"path": "D:\\娴嬭瘯\\LWOPS_瀹夊叏娴嬭瘯璧勬枡_2026-05-10\\reverse-poc\\public\\captures\\cmdb-task-log-status0-55274.json",
			},
			{
				"kind": "source",
				"path": "D:\\椤圭洰杩佺Щ鏂囦欢\\骞冲彴婧愮爜\\AutoOps-main\\AutoOps-main\\web\\src\\api\\cmdb.js",
			},
			{
				"kind": "source",
				"path": "D:\\椤圭洰杩佺Щ鏂囦欢\\骞冲彴婧愮爜\\AutoOps-main\\AutoOps-main\\api\\api\\monitor\\service\\monitorService.go",
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
			cmdbContractFieldGroup("write_binding_contract", "missing_backend", []string{
				"host", "hostid", "templateid", "server_object_id", "server_platform_id", "cmdb_object_id",
				"group", "tags", "active_status", "hosttype", "subtype",
			}, "monitor_host_binding_contract"),
			cmdbContractFieldGroup("write_mapping_contract", "missing_backend", []string{
				"cmdb_attr_id", "server_attr_id", "server_model_id", "server_model_name", "attr", "attr_stru", "queue",
			}, "cmdb_monitor_binding_write_receipt_contract"),
			cmdbContractFieldGroup("write_safety_contract", "missing_backend", []string{
				"audit_ref", "rollback_ref", "idempotency_key",
			}, "binding_audit_contract"),
		},
		"source_evidence": []gin.H{
			{
				"kind": "capture",
				"path": "D:\\娴嬭瘯\\LWOPS_瀹夊叏娴嬭瘯璧勬枡_2026-05-10\\reverse-poc\\public\\captures\\cmdb-task-log-status0-55274.json",
			},
			{
				"kind": "source",
				"path": "D:\\椤圭洰杩佺Щ鏂囦欢\\骞冲彴婧愮爜\\AutoOps-main\\AutoOps-main\\web\\src\\api\\cmdb.js",
			},
		},
		"compatibility_notes": []string{
			"request bodies are not echoed to avoid leaking credentials or sensitive binding values",
			"binding writes must not imply execution, applied config, installation, data arrival, rollback, or uninstall receipts",
		},
	}
}

func cmdbContractFieldGroup(name, status string, fields []string, missingContract string) gin.H {
	return gin.H{
		"name":             name,
		"fields":           fields,
		"missing_contract": missingContract,
		"status":           status,
	}
}

func cmdbContractMatrixItem(name, status, gapID string, fields []string, reason string) gin.H {
	return gin.H{
		"name":   name,
		"status": status,
		"gap_id": gapID,
		"fields": fields,
		"reason": reason,
	}
}
