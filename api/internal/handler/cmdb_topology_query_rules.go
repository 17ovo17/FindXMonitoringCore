package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
)

func cmdbTopologyRelationQueries(graph cmdbTopologyTraversalGraph) []gin.H {
	out := make([]gin.H, 0, len(graph.nodes)-1)
	for _, node := range graph.nodes {
		if node.inst.ID == graph.rootID {
			continue
		}
		chain, ok := cmdbTopologyPathEdges(graph, node.inst.ID)
		if !ok || len(chain) == 0 {
			continue
		}
		out = append(out, cmdbTopologyRelationQueryRule(graph, node, chain))
	}
	return out
}

func cmdbTopologyPathEdges(graph cmdbTopologyTraversalGraph, nodeID string) ([]cmdbTopologyTraversalEdge, bool) {
	chain := []cmdbTopologyTraversalEdge{}
	currentID := nodeID
	for currentID != graph.rootID {
		edge, ok := graph.parentEdge[currentID]
		if !ok {
			return nil, false
		}
		chain = append([]cmdbTopologyTraversalEdge{edge}, chain...)
		parentID, parentOK := graph.parent[currentID]
		if !parentOK {
			return nil, false
		}
		currentID = parentID
	}
	return chain, true
}

func cmdbTopologyRelationQueryRule(graph cmdbTopologyTraversalGraph, node cmdbTopologyTraversalNode, chain []cmdbTopologyTraversalEdge) gin.H {
	startObjectID := chain[0].from.ObjectID
	endObjectID := node.inst.ObjectID
	pathLabel := cmdbTopologyRelationQueryPathLabel(chain)
	pathID := cmdbStableSHA1("cmdb-relation-query:path:" + startObjectID + ":" + endObjectID + ":" + pathLabel)
	return gin.H{
		"_id":               "findx-relation-query-" + pathID[:24],
		"name":              pathLabel,
		"start_vertex":      startObjectID,
		"end_vertex":        endObjectID,
		"filter_vertex":     []string{},
		"via_vertex":        cmdbTopologyRelationQueryViaVertices(chain),
		"path_id":           pathID[:32],
		"path_label":        pathLabel,
		"path":              cmdbTopologyRelationQueryPath(chain),
		"authority_type":    1,
		"authority_data":    gin.H{"users": []int{}, "roles": []int{}},
		"paths_hash":        cmdbStableSHA1("cmdb-relation-query:hash:" + pathLabel),
		"paths_updated_at":  0,
		"created_at":        0,
		"creator_id":        0,
		"creator_name":      "FindX",
		"start_vertex_name": cmdbObjectName(startObjectID),
		"end_vertex_name":   cmdbObjectName(endObjectID),
		"start_instance_id": graph.rootID,
		"end_instance_id":   node.inst.ID,
	}
}

func cmdbTopologyRelationQueryPath(chain []cmdbTopologyTraversalEdge) []gin.H {
	path := make([]gin.H, 0, len(chain)*2+1)
	path = append(path, cmdbTopologyRelationQueryObjectVertex(chain[0].from.ObjectID))
	for _, edge := range chain {
		path = append(path, cmdbTopologyRelationQueryEdge(edge))
		path = append(path, cmdbTopologyRelationQueryObjectVertex(edge.to.ObjectID))
	}
	return path
}

func cmdbTopologyRelationQueryObjectVertex(objectID string) gin.H {
	return gin.H{
		"id":    "1:" + objectID,
		"label": "object",
		"type":  "vertex",
		"properties": gin.H{
			"object_id":   objectID,
			"name":        cmdbObjectName(objectID),
			"type":        cmdbObjectType(objectID),
			"is_preset":   false,
			"p_object_id": "",
		},
	}
}

func cmdbTopologyRelationQueryEdge(edge cmdbTopologyTraversalEdge) gin.H {
	sourceObjectID := edge.from.ObjectID
	targetObjectID := edge.to.ObjectID
	schema := cmdbRelationSchema(edge.rel, edge.relationType, edge.from, edge.to)
	return gin.H{
		"id":        "S1:" + sourceObjectID + ">1>" + edge.rel.RelationTypeID + ">S1:" + targetObjectID,
		"label":     "relation",
		"type":      "edge",
		"outV":      "1:" + sourceObjectID,
		"outVLabel": "object",
		"inV":       "1:" + targetObjectID,
		"inVLabel":  "object",
		"properties": gin.H{
			"relation_id": edge.rel.RelationTypeID,
			"asst_id":     schema["asst_id"],
			"mapping":     schema["mapping"],
			"visible":     schema["visible"],
		},
	}
}

func cmdbTopologyRelationQueryViaVertices(chain []cmdbTopologyTraversalEdge) []string {
	if len(chain) <= 1 {
		return []string{}
	}
	via := make([]string, 0, len(chain)-1)
	for i := 0; i < len(chain)-1; i++ {
		via = append(via, chain[i].to.ObjectID)
	}
	return via
}

func cmdbTopologyRelationQueryPathLabel(chain []cmdbTopologyTraversalEdge) string {
	if len(chain) == 0 {
		return ""
	}
	label := cmdbObjectName(chain[0].from.ObjectID)
	for _, edge := range chain {
		label += "-(" + cmdbRelationDisplayName(edge.relationType) + ")" + cmdbObjectName(edge.to.ObjectID)
	}
	return label
}

func cmdbTopologyRelationQueryRuntimeBlocked(instanceID string) gin.H {
	return gin.H{
		"status":   "pending",
		"error":    "pending",
		"contract": "cmdb.relation_query.runtime.read.v1",
		"message":  "relation query rules are source-compatible DTOs; recursive path execution and final layout coordinates require runtime contracts.",
		"missing_contracts": []string{
			"cmdb_relation_query_executor_contract",
			"cmdb_relation_recursive_path_contract",
			"cmdb_topology_layout_coordinate_contract",
		},
		"instance_id": strings.TrimSpace(instanceID),
		"expected_schema": gin.H{
			"query":              []string{"start_vertex", "end_vertex", "via_vertex", "path_id"},
			"recursive_paths[]":  []string{"nodes", "edges", "depth", "authority_data", "layout_coordinates"},
			"layout_coordinates": []string{"instance_id", "x", "y", "layer", "order", "engine"},
		},
		"field_matrix": []gin.H{
			cmdbContractFieldGroup("relation_query_executor", "blocked", []string{
				"start_vertex", "end_vertex", "filter_vertex", "via_vertex", "authority_data",
			}, "cmdb_relation_query_executor_contract"),
			cmdbContractFieldGroup("relation_recursive_path", "blocked", []string{
				"recursive_paths", "depth", "cycle_guard", "path_hash",
			}, "cmdb_relation_recursive_path_contract"),
			cmdbContractFieldGroup("topology_layout_coordinate", "blocked", []string{
				"x", "y", "layer", "order", "engine",
			}, "cmdb_topology_layout_coordinate_contract"),
		},
		"safe_to_retry": false,
	}
}

func cmdbTopologyRelationQueryRuntime(graph cmdbTopologyTraversalGraph, execute bool) gin.H {
	blocked := cmdbTopologyRelationQueryRuntimeBlocked(graph.rootID)
	if execute {
		blocked["execute_requested"] = true
		blocked["message"] = "relation_query=execute was requested, but no verified relation query executor or persisted layout-coordinate contract is available; source-compatible query rules remain DTO-only."
	}
	return blocked
}

func cmdbTopologyReadyMissingContracts() []string {
	return []string{
		"cmdb_monitor_binding_executor_contract",
		"cmdb_relation_action_delivery_executor_contract",
		"cmdb_relation_query_executor_contract",
		"cmdb_relation_recursive_path_contract",
		"cmdb_topology_layout_coordinate_contract",
	}
}

func cmdbStableSHA1(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}
