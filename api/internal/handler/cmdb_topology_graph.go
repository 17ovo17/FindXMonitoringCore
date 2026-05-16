package handler

import (
	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

type cmdbTopologyTraversalNode struct {
	inst  model.CmdbInstance
	level int
	path  []string
}

type cmdbTopologyTraversalEdge struct {
	rel          model.CmdbInstanceRelation
	relationType model.CmdbRelationType
	from         model.CmdbInstance
	to           model.CmdbInstance
	location     string
	path         []string
}

type cmdbTopologyTraversalGraph struct {
	rootID         string
	nodes          []cmdbTopologyTraversalNode
	edges          []cmdbTopologyTraversalEdge
	parent         map[string]string
	parentEdge     map[string]cmdbTopologyTraversalEdge
	childrenByNode map[string][]string
}

func buildCmdbInstanceTopologyGraph(instanceID string, executeRelationQuery bool) (gin.H, bool) {
	root, ok := store.GetCmdbInstance(instanceID)
	if !ok {
		return nil, false
	}
	graph, ok := cmdbBuildTopologyTraversal(*root)
	if !ok || len(graph.edges) == 0 {
		return nil, false
	}

	nodes := cmdbApplyStableTopologyLayout(graph.nodes)
	edges := cmdbTopologyEdges(graph.edges)
	relationQueries := cmdbTopologyRelationQueries(graph)
	relationQueryRuntime := cmdbTopologyRelationQueryRuntime(graph, executeRelationQuery)
	rootInstance := cmdbTopologyInstanceWithRelations(*root, 1, graph)
	children := cmdbTopologyChildrenForRoot(graph, root.ID)

	contract := cmdbInstanceTopologyBlockedContract()
	readyContract := cmdbInstanceTopologyReadyContract(contract)
	data := gin.H{
		"name":      nodes[0]["object_name"],
		"object":    gin.H{"id": root.ObjectID, "name": nodes[0]["object_name"], "type": nodes[0]["object_type"]},
		"instances": []gin.H{rootInstance},
		"children":  children,
		"relation": gin.H{
			"in":  cmdbTopologyRootRelationDescriptors(graph.edges, root.ID, "target"),
			"out": cmdbTopologyRootRelationDescriptors(graph.edges, root.ID, "source"),
		},
		"level":                  1,
		"nodes":                  nodes,
		"edges":                  edges,
		"relation_queries":       relationQueries,
		"relation_query_rules":   relationQueries,
		"relation_query_runtime": relationQueryRuntime,
		"layout": gin.H{
			"engine": "findx_cmdb_stable_layered",
			"mode":   "deterministic",
		},
	}
	topology := gin.H{
		"code":                   0,
		"message":                "",
		"status":                 "ready_with_contract_gaps",
		"_id":                    "findx-topology-" + root.ID,
		"object_id":              root.ObjectID,
		"name":                   "default topology",
		"data":                   data,
		"default":                true,
		"filter_empty":           false,
		"sort":                   0,
		"instance_id":            root.ID,
		"creator":                root.Creator,
		"create_time":            root.CreatedAt,
		"business_status":        false,
		"in_inst_detail":         false,
		"business_names":         []string{},
		"nodes":                  nodes,
		"edges":                  edges,
		"relation_queries":       relationQueries,
		"relation_query_rules":   relationQueries,
		"relation_query_runtime": relationQueryRuntime,
		"contract_gap_id":        readyContract["contract_gap_id"],
		"contract_matrix":        readyContract["contract_matrix"],
		"expected_schema":        readyContract["expected_schema"],
		"field_matrix":           readyContract["field_matrix"],
		"source_evidence":        readyContract["source_evidence"],
		"missing_contracts":      cmdbTopologyReadyMissingContracts(),
		"audit_context": gin.H{
			"scope":          "cmdb",
			"resource_type":  "cmdb_instance_relation",
			"resource_id":    root.ID,
			"relation_count": len(edges),
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	}
	topology["data"] = data
	return topology, true
}

func cmdbBuildTopologyTraversal(root model.CmdbInstance) (cmdbTopologyTraversalGraph, bool) {
	graph := cmdbTopologyTraversalGraph{
		rootID:         root.ID,
		nodes:          []cmdbTopologyTraversalNode{{inst: root, level: 1, path: []string{root.ID}}},
		edges:          []cmdbTopologyTraversalEdge{},
		parent:         map[string]string{},
		parentEdge:     map[string]cmdbTopologyTraversalEdge{},
		childrenByNode: map[string][]string{},
	}
	seenNodes := map[string]bool{root.ID: true}
	seenEdges := map[string]bool{}
	nodeByID := map[string]cmdbTopologyTraversalNode{root.ID: graph.nodes[0]}
	queue := []string{root.ID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		currentNode := nodeByID[currentID]
		relations := store.ListCmdbInstanceRelations(currentID)
		for _, rel := range relations {
			relatedID, location := cmdbRelatedInstanceID(rel, currentID)
			if relatedID == "" {
				return cmdbTopologyTraversalGraph{}, false
			}
			related, found := store.GetCmdbInstance(relatedID)
			if !found {
				return cmdbTopologyTraversalGraph{}, false
			}
			relationType, relationTypeOK := cmdbRelationType(rel.RelationTypeID)
			if !relationTypeOK {
				return cmdbTopologyTraversalGraph{}, false
			}
			source, sourceOK := store.GetCmdbInstance(rel.SourceInstanceID)
			target, targetOK := store.GetCmdbInstance(rel.TargetInstanceID)
			if !sourceOK || !targetOK {
				return cmdbTopologyTraversalGraph{}, false
			}

			edgePath := cmdbAppendTopologyPath(currentNode.path, relatedID)
			if !seenEdges[rel.ID] {
				graph.edges = append(graph.edges, cmdbTopologyTraversalEdge{
					rel:          rel,
					relationType: relationType,
					from:         *source,
					to:           *target,
					location:     location,
					path:         edgePath,
				})
				seenEdges[rel.ID] = true
			}
			if !seenNodes[relatedID] {
				nextNode := cmdbTopologyTraversalNode{
					inst:  *related,
					level: currentNode.level + 1,
					path:  edgePath,
				}
				graph.nodes = append(graph.nodes, nextNode)
				nodeByID[relatedID] = nextNode
				seenNodes[relatedID] = true
				graph.parent[relatedID] = currentID
				graph.parentEdge[relatedID] = cmdbTopologyTraversalEdge{
					rel:          rel,
					relationType: relationType,
					from:         *source,
					to:           *target,
					location:     location,
					path:         edgePath,
				}
				graph.childrenByNode[currentID] = append(graph.childrenByNode[currentID], relatedID)
				queue = append(queue, relatedID)
			}
		}
	}
	return graph, true
}

func cmdbAppendTopologyPath(path []string, id string) []string {
	out := make([]string, 0, len(path)+1)
	out = append(out, path...)
	return append(out, id)
}

func cmdbApplyStableTopologyLayout(nodes []cmdbTopologyTraversalNode) []gin.H {
	out := make([]gin.H, 0, len(nodes))
	layerIndex := map[int]int{}
	for _, node := range nodes {
		index := layerIndex[node.level]
		layerIndex[node.level] = index + 1
		x := (node.level - 1) * 320
		y := index * 140
		item := cmdbTopologyNode(node.inst, node.level)
		item["layer"] = node.level
		item["x"] = x
		item["y"] = y
		item["path"] = node.path
		item["layout"] = gin.H{
			"engine": "findx_cmdb_stable_layered",
			"layer":  node.level,
			"order":  index,
			"x":      x,
			"y":      y,
		}
		out = append(out, item)
	}
	return out
}

func cmdbTopologyEdges(edges []cmdbTopologyTraversalEdge) []gin.H {
	out := make([]gin.H, 0, len(edges))
	for _, edge := range edges {
		out = append(out, cmdbTopologyEdgeFromTraversal(edge))
	}
	return out
}

func cmdbTopologyEdgeFromTraversal(edge cmdbTopologyTraversalEdge) gin.H {
	schema := cmdbRelationSchema(edge.rel, edge.relationType, edge.from, edge.to)
	pathNodes := []string{cmdbObjectName(edge.from.ObjectID), cmdbObjectName(edge.to.ObjectID)}
	return gin.H{
		"id":                   edge.rel.ID,
		"source":               edge.rel.SourceInstanceID,
		"target":               edge.rel.TargetInstanceID,
		"relation_id":          edge.rel.RelationTypeID,
		"relation_name":        cmdbRelationDisplayName(edge.relationType),
		"instance_relation_id": edge.rel.ID,
		"asst_id":              schema["asst_id"],
		"asst_name":            schema["asst_name"],
		"mapping":              schema["mapping"],
		"visible":              schema["visible"],
		"rule_logic":           schema["rule_logic"],
		"rule_expression":      schema["rule_expression"],
		"rules":                schema["rules"],
		"path_label":           cmdbRelationPathLabel(pathNodes, cmdbRelationDisplayName(edge.relationType)),
		"relation_schema":      schema,
		"direction":            cmdbTopologyDirection(edge.location),
		"location":             edge.location,
		"related_instance_id":  cmdbTraversalRelated(edge).ID,
		"related_object_id":    cmdbTraversalRelated(edge).ObjectID,
		"path":                 edge.path,
	}
}

func cmdbTraversalRelated(edge cmdbTopologyTraversalEdge) model.CmdbInstance {
	if edge.location == "target" {
		return edge.from
	}
	return edge.to
}

func cmdbTopologyChildrenForRoot(graph cmdbTopologyTraversalGraph, rootID string) []gin.H {
	children := make([]gin.H, 0, len(graph.childrenByNode[rootID]))
	for _, childID := range graph.childrenByNode[rootID] {
		child, ok := cmdbTopologyChildForNode(graph, childID)
		if ok {
			children = append(children, child)
		}
	}
	return children
}

func cmdbTopologyChildForNode(graph cmdbTopologyTraversalGraph, nodeID string) (gin.H, bool) {
	edge, ok := graph.parentEdge[nodeID]
	if !ok {
		return nil, false
	}
	inst := cmdbTraversalRelated(edge)
	level := len(edge.path)
	relation := gin.H{"in": []gin.H{}, "out": []gin.H{}}
	if edge.rel.SourceInstanceID == graph.parent[nodeID] {
		relation["in"] = []gin.H{cmdbTopologyRelationDescriptor(edge.rel.RelationTypeID, edge.relationType, "source")}
	} else {
		relation["out"] = []gin.H{cmdbTopologyRelationDescriptor(edge.rel.RelationTypeID, edge.relationType, "target")}
	}
	children := make([]gin.H, 0, len(graph.childrenByNode[nodeID]))
	for _, childID := range graph.childrenByNode[nodeID] {
		child, childOK := cmdbTopologyChildForNode(graph, childID)
		if childOK {
			children = append(children, child)
		}
	}
	return gin.H{
		"name":      cmdbObjectName(inst.ObjectID),
		"object":    gin.H{"id": inst.ObjectID, "name": cmdbObjectName(inst.ObjectID), "type": cmdbObjectType(inst.ObjectID)},
		"instances": []gin.H{cmdbTopologyInstanceWithRelations(inst, level, graph)},
		"children":  children,
		"relation":  relation,
		"level":     level,
	}, true
}

func cmdbInstanceRelationsForNode(inst model.CmdbInstance, graph cmdbTopologyTraversalGraph) ([]gin.H, []gin.H) {
	in := make([]gin.H, 0)
	out := make([]gin.H, 0)
	for _, edge := range graph.edges {
		if edge.rel.SourceInstanceID == inst.ID {
			out = append(out, cmdbTopologyInstanceRelation(edge.rel, edge.to, edge.relationType, "out"))
		}
		if edge.rel.TargetInstanceID == inst.ID {
			in = append(in, cmdbTopologyInstanceRelation(edge.rel, edge.from, edge.relationType, "in"))
		}
	}
	return in, out
}

func cmdbInstanceTopologyReadyContract(blockedContract gin.H) gin.H {
	return gin.H{
		"contract_gap_id": blockedContract["contract_gap_id"],
		"contract_matrix": []gin.H{
			cmdbContractMatrixItem("topology_request_path", "ready", "cmdb.instance.topology.path.v1", []string{"object_id", "instance_id", "topology_id"}, "FindX resolves the current instance_id route and preserves the mature route shape in expected_schema"),
			cmdbContractMatrixItem("topology_document_envelope", "ready", "cmdb.topology.document.envelope.v1", []string{"_id", "object_id", "name", "data", "default", "filter_empty", "status", "sort", "instance_id", "creator", "create_time", "business_status", "in_inst_detail", "business_names"}, "relation-backed envelope is available with deterministic layout metadata"),
			cmdbContractMatrixItem("topology_instance_identity", "ready", "cmdb.instance.identity.read.v1", []string{"object_id", "object_name", "object_type", "id", "name"}, "CmdbInstance and CmdbObject are read from the current store"),
			cmdbContractMatrixItem("relation_path_levels", "ready_recursive_from_rows", "cmdb.relation.path.levels.v1", []string{"children", "level", "object", "relation", "instances", "path"}, "relation rows are traversed recursively with cycle-safe node and edge de-duplication; final recursive path execution remains blocked by runtime contract"),
			cmdbContractMatrixItem("relation_in_edges", "ready", "cmdb.relation.in.edges.v1", []string{"in", "related_instance_id", "related_object_id", "instance_relation_id", "relation_id"}, "backed by CmdbInstanceRelation rows and CmdbRelationType rows"),
			cmdbContractMatrixItem("relation_out_edges", "ready", "cmdb.relation.out.edges.v1", []string{"out", "related_instance_id", "related_object_id", "instance_relation_id", "relation_id"}, "backed by CmdbInstanceRelation rows and CmdbRelationType rows"),
			cmdbContractMatrixItem("relation_descriptor_matrix", "ready_from_relation_type", "cmdb.relation.descriptor.matrix.v1", []string{"asst_id", "asst_name", "side", "position", "direction", "location"}, "descriptor fields are derived from relation type rows and mature topology field names"),
			cmdbContractMatrixItem("relation_query_rules", "ready_from_relation_path", "cmdb.relation.query_rules.v1", []string{"start_vertex", "end_vertex", "path_id", "path_label", "path", "authority_type", "authority_data", "paths_hash"}, "verified relation rows are projected into mature relation-query path rule DTOs"),
			cmdbContractMatrixItem("topology_layout_coordinate", "blocked_by_executor", "cmdb.topology.layout_coordinate.v1", []string{"x", "y", "layer", "layout"}, "deterministic row-derived coordinates are exposed for graph readability, but final layout coordinates require a verified layout executor"),
		},
		"expected_schema": blockedContract["expected_schema"],
		"field_matrix": []gin.H{
			cmdbContractFieldGroup("topology_document", "ready", []string{
				"_id", "object_id", "name", "data", "default", "filter_empty", "status", "sort",
				"instance_id", "creator", "create_time", "business_status", "in_inst_detail", "business_names",
			}, "none"),
			cmdbContractFieldGroup("topology_instance", "ready", []string{
				"object_id", "object_name", "object_type", "id", "name", "in", "out",
			}, "cmdb_instance_relation_store"),
			cmdbContractFieldGroup("topology_relation_edge", "ready", []string{
				"related_instance_id", "related_object_id", "instance_relation_id", "relation_id", "path",
			}, "cmdb_relation_graph_contract"),
			cmdbContractFieldGroup("topology_children", "ready_recursive_from_rows", []string{
				"name", "instances", "children", "object", "relation", "level", "path",
			}, "cmdb_topology_field_mapping_contract"),
			cmdbContractFieldGroup("topology_relation_descriptor", "ready_from_relation_type", []string{
				"asst_id", "asst_name", "side", "position", "relation_id", "direction", "location",
			}, "cmdb_relation_graph_contract"),
			cmdbContractFieldGroup("topology_relation_query_rules", "ready_from_relation_path", []string{
				"start_vertex", "end_vertex", "path_id", "path_label", "path", "authority_type", "authority_data", "paths_hash",
			}, "cmdb_relation_query_rules_contract"),
			cmdbContractFieldGroup("topology_layout_coordinate", "blocked_by_executor", []string{
				"x", "y", "layer", "layout.engine", "layout.order",
			}, "cmdb_topology_layout_coordinate_contract"),
		},
		"source_evidence": blockedContract["source_evidence"],
	}
}

func cmdbTopologyNode(inst model.CmdbInstance, level int) gin.H {
	attrs := store.ListCmdbAttributes(inst.ObjectID)
	name := cmdbDefaultDisplay(inst, attrs)["value"]
	return gin.H{
		"id":          inst.ID,
		"name":        name,
		"object_id":   inst.ObjectID,
		"object_name": cmdbObjectName(inst.ObjectID),
		"object_type": cmdbObjectType(inst.ObjectID),
		"kind":        "instance",
		"level":       level,
	}
}

func cmdbTopologyInstance(inst model.CmdbInstance) gin.H {
	node := cmdbTopologyNode(inst, 1)
	return gin.H{
		"object_id":   node["object_id"],
		"object_name": node["object_name"],
		"object_type": node["object_type"],
		"id":          node["id"],
		"name":        node["name"],
		"in":          []gin.H{},
		"out":         []gin.H{},
	}
}

func cmdbTopologyInstanceWithRelations(inst model.CmdbInstance, level int, graph cmdbTopologyTraversalGraph) gin.H {
	instance := cmdbTopologyInstance(inst)
	in, out := cmdbInstanceRelationsForNode(inst, graph)
	instance["in"] = in
	instance["out"] = out
	instance["level"] = level
	return instance
}

func cmdbTopologyInstanceRelation(rel model.CmdbInstanceRelation, related model.CmdbInstance, relationType model.CmdbRelationType, direction string) gin.H {
	schema, schemaOK := cmdbRelationSchemaForRelation(rel, relationType)
	if !schemaOK {
		schema = gin.H{}
	}
	return gin.H{
		"related_instance_id":  related.ID,
		"related_object_id":    related.ObjectID,
		"instance_relation_id": rel.ID,
		"relation_id":          rel.RelationTypeID,
		"asst_id":              cmdbRelationAssocID(relationType),
		"asst_name":            cmdbRelationDisplayName(relationType),
		"mapping":              schema["mapping"],
		"visible":              schema["visible"],
		"rule_logic":           schema["rule_logic"],
		"rule_expression":      schema["rule_expression"],
		"rules":                schema["rules"],
		"relation_schema":      schema,
		"direction":            direction,
		"location":             direction,
	}
}

func cmdbTopologyRootRelationDescriptors(edges []cmdbTopologyTraversalEdge, instanceID, location string) []gin.H {
	descriptors := make([]gin.H, 0, len(edges))
	for _, edge := range edges {
		if edge.location != location {
			continue
		}
		if location == "source" && edge.rel.SourceInstanceID != instanceID {
			continue
		}
		if location == "target" && edge.rel.TargetInstanceID != instanceID {
			continue
		}
		descriptors = append(descriptors, cmdbTopologyRelationDescriptor(edge.rel.RelationTypeID, edge.relationType, location))
	}
	return descriptors
}

func cmdbTopologyRelationDescriptor(relationID string, relationType model.CmdbRelationType, location string) gin.H {
	return gin.H{
		"asst_id":     cmdbRelationAssocID(relationType),
		"asst_name":   cmdbRelationDisplayName(relationType),
		"mapping":     firstNonEmptyRelationString(relationType.Mapping, "n:1"),
		"visible":     cmdbRelationVisible(relationType),
		"side":        "right",
		"position":    "right",
		"relation_id": relationID,
		"direction":   cmdbTopologyDirection(location),
		"location":    location,
	}
}

func cmdbRelatedInstanceID(rel model.CmdbInstanceRelation, instanceID string) (string, string) {
	if rel.SourceInstanceID == instanceID {
		return rel.TargetInstanceID, "source"
	}
	if rel.TargetInstanceID == instanceID {
		return rel.SourceInstanceID, "target"
	}
	return "", ""
}

func cmdbRelationType(id string) (model.CmdbRelationType, bool) {
	if rel, ok := store.GetCmdbRelationType(id); ok {
		return *rel, true
	}
	return model.CmdbRelationType{}, false
}

func cmdbRelationAssocID(rel model.CmdbRelationType) string {
	if rel.Name != "" {
		return rel.Name
	}
	return rel.ID
}

func cmdbRelationDisplayName(rel model.CmdbRelationType) string {
	if rel.Label != "" {
		return rel.Label
	}
	if rel.Name != "" {
		return rel.Name
	}
	return rel.ID
}

func cmdbTopologyDirection(location string) int {
	if location == "target" {
		return 1
	}
	return 3
}

func cmdbObjectType(objectID string) int {
	if obj, ok := store.GetCmdbObject(objectID); ok {
		return obj.ObjectType
	}
	return 101
}
