package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbInstanceTopologyReturnsRealRelationGraph(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-success")

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/"+rootID+"/topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["code"].(float64) != 0 {
		t.Fatalf("code = %v, want 0", body["code"])
	}
	if body["status"] != "ready_with_contract_gaps" {
		t.Fatalf("status = %v, want ready_with_contract_gaps", body["status"])
	}
	if body["creator"] != "test" || body["create_time"] == "" {
		t.Fatalf("topology envelope missing creator/create_time: %#v", body)
	}
	nodes := body["nodes"].([]any)
	edges := body["edges"].([]any)
	if len(nodes) != 2 || len(edges) != 1 {
		t.Fatalf("graph shape mismatch nodes=%#v edges=%#v", nodes, edges)
	}
	edge := edges[0].(map[string]any)
	if edge["source"] != rootID || edge["target"] != targetID {
		t.Fatalf("edge source/target mismatch: %#v", edge)
	}
	if edge["instance_relation_id"] != "cmdb-relation-success-rel" || edge["relation_id"] != "depends_on" || edge["relation_name"] != "depends on" {
		t.Fatalf("edge relation metadata mismatch: %#v", edge)
	}
	if edge["related_instance_id"] != targetID || edge["related_object_id"] != "cmdb-relation-success-db" {
		t.Fatalf("edge related metadata mismatch: %#v", edge)
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"expected_schema",
		"field_matrix",
		"source_evidence",
		"audit_context",
		"cmdb_instance_relation",
	})
	assertCmdbContractMatrixStatuses(t, body["contract_matrix"].([]any), map[string]string{
		"relation_in_edges":          "ready",
		"relation_out_edges":         "ready",
		"relation_descriptor_matrix": "ready_from_relation_type",
		"topology_layout_coordinate": "blocked_by_executor",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		"root-password-marker",
		"target-token-marker",
		"secret-dsn-marker",
		"missing_relation_store",
		"cmdb_topology_layout_contract",
		`"success":true`,
		`"status":"success"`,
		`"running"`,
		`"queued"`,
	})
}

func TestCmdbInstanceTopologyReturnsRecursivePathAndStableLayout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, middleID, leafID := createCmdbRecursiveRelationFixture(t, "cmdb-relation-recursive")

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/"+rootID+"/topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	nodes := body["nodes"].([]any)
	edges := body["edges"].([]any)
	if len(nodes) != 3 || len(edges) != 2 {
		t.Fatalf("recursive graph shape mismatch nodes=%#v edges=%#v", nodes, edges)
	}
	nodeByID := map[string]map[string]any{}
	for _, raw := range nodes {
		node := raw.(map[string]any)
		id := node["id"].(string)
		nodeByID[id] = node
		if node["x"] == nil || node["y"] == nil || node["layer"] == nil || node["layout"] == nil {
			t.Fatalf("node %s missing stable layout fields: %#v", id, node)
		}
		layout := node["layout"].(map[string]any)
		if layout["x"] != node["x"] || layout["y"] != node["y"] || layout["layer"] != node["layer"] {
			t.Fatalf("node %s layout mirror mismatch: node=%#v layout=%#v", id, node, layout)
		}
	}
	if nodeByID[rootID]["level"] != float64(1) || nodeByID[rootID]["layer"] != float64(1) {
		t.Fatalf("root level/layer mismatch: %#v", nodeByID[rootID])
	}
	if nodeByID[middleID]["level"] != float64(2) || nodeByID[middleID]["layer"] != float64(2) {
		t.Fatalf("middle level/layer mismatch: %#v", nodeByID[middleID])
	}
	if nodeByID[leafID]["level"] != float64(3) || nodeByID[leafID]["layer"] != float64(3) {
		t.Fatalf("leaf level/layer mismatch: %#v", nodeByID[leafID])
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		`"path":["` + rootID + `","` + middleID + `"]`,
		`"path":["` + rootID + `","` + middleID + `","` + leafID + `"]`,
		`"relation_path_levels"`,
		`"ready_recursive_from_rows"`,
		`"topology_layout_coordinate"`,
		`"blocked_by_executor"`,
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		"cmdb_topology_layout_contract",
		`"nodes":[]`,
		`"edges":[]`,
		`"queued"`,
		`"running"`,
		`"applied"`,
	})
}

func TestCmdbInstanceTopologyCycleDeduplicatesNodesAndEdges(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, middleID, leafID := createCmdbRecursiveRelationFixture(t, "cmdb-relation-cycle")
	if err := store.CreateCmdbInstanceRelation(&model.CmdbInstanceRelation{
		ID:               "cmdb-relation-cycle-rel-leaf-root",
		SourceInstanceID: leafID,
		TargetInstanceID: rootID,
		RelationTypeID:   "depends_on",
	}); err != nil {
		t.Fatalf("create cycle relation: %v", err)
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/"+rootID+"/topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	nodes := body["nodes"].([]any)
	edges := body["edges"].([]any)
	if len(nodes) != 3 || len(edges) != 3 {
		t.Fatalf("cycle graph should keep unique nodes and relation rows once: nodes=%#v edges=%#v", nodes, edges)
	}
	seenNodes := map[string]bool{}
	for _, raw := range nodes {
		node := raw.(map[string]any)
		id := node["id"].(string)
		if seenNodes[id] {
			t.Fatalf("duplicate node in cycle graph: %s nodes=%#v", id, nodes)
		}
		seenNodes[id] = true
	}
	for _, id := range []string{rootID, middleID, leafID} {
		if !seenNodes[id] {
			t.Fatalf("cycle graph missing node %s: %#v", id, nodes)
		}
	}
	seenEdges := map[string]bool{}
	for _, raw := range edges {
		edge := raw.(map[string]any)
		id := edge["id"].(string)
		if seenEdges[id] {
			t.Fatalf("duplicate edge in cycle graph: %s edges=%#v", id, edges)
		}
		seenEdges[id] = true
	}
}

func TestCmdbInstanceTopologyDanglingRecursiveRelationIsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, middleID, _ := createCmdbRecursiveRelationFixture(t, "cmdb-relation-recursive-dangling")
	if err := store.CreateCmdbInstanceRelation(&model.CmdbInstanceRelation{
		ID:               "cmdb-relation-recursive-dangling-rel-bad",
		SourceInstanceID: middleID,
		TargetInstanceID: "cmdb-relation-recursive-dangling-missing-leaf",
		RelationTypeID:   "depends_on",
	}); err != nil {
		t.Fatalf("create dangling recursive relation: %v", err)
	}

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/"+rootID+"/topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{"PENDING", "cmdb_relation_graph_contract"})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"nodes":[`, `"edges":[`, `"code":0`, `"success":true`})
}

func TestCmdbRelationActionRequestPersistsStoreAndAuditLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-real")

	router := gin.New()
	router.POST("/api/v1/cmdb/relation-actions", CreateCmdbRelationActionRequest)
	router.GET("/api/v1/cmdb/relation-actions/:id", GetCmdbRelationActionRequest)

	payload := `{
		"action":"relation",
		"instance_id":"` + rootID + `",
		"node_id":"` + targetID + `",
		"object_id":"cmdb-relation-action-real-db",
		"relation_id":"cmdb-relation-action-real-rel",
		"context":{"password":"secret-marker","dsn":"mysql://user:pass@example/db","note":"relation action audit"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/relation-actions", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "recorded" {
		t.Fatalf("status = %v, want recorded: %s", body["status"], w.Body.String())
	}
	action := body["action_request"].(map[string]any)
	if action["status"] != "recorded" || action["delivery_status"] != "PENDING" || action["effect_status"] != "PENDING" {
		t.Fatalf("action request status mismatch: %#v", action)
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		`"action":"relation"`,
		`"relation_id":"cmdb-relation-action-real-rel"`,
		`"audit_ref"`,
		`"cmdb_relation_action"`,
		`"cmdb.relation.relation.request"`,
		`"cmdb_relation_action_delivery_receipt_contract"`,
		`"cmdb_relation_action_effect_receipt_contract"`,
		`"findx_audit"`,
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		"secret-marker",
		"mysql://user:pass@example/db",
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"data_arrived"`,
		`"success":true`,
	})

	id := action["id"].(string)
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/relation-actions/"+id, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusConflict {
		t.Fatalf("read status = %d, want %d, body=%s", getW.Code, http.StatusConflict, getW.Body.String())
	}
	assertCmdbTestStringContainsAll(t, getW.Body.String(), []string{
		"PENDING",
		"cmdb.relation_action.runtime.read.v1",
		"cmdb_relation_action_action_executor_contract",
		"cmdb_relation_action_attested_receipt_contract",
	})
	assertCmdbTestStringExcludesAll(t, getW.Body.String(), []string{
		"secret-marker",
		"mysql://user:pass@example/db",
		`"code":0`,
		`"status":"recorded"`,
		`"action_request":{`,
		`"receipts":[`,
	})

	auditResp, err := store.QueryFindXAuditLogs(model.LogQueryRequest{
		Source:       model.LogsSourceFindXAudit,
		Scope:        "cmdb",
		ResourceType: "cmdb_relation_action",
		ResourceID:   rootID,
		Action:       "cmdb.relation.relation.request",
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	if len(auditResp.Items) == 0 {
		t.Fatalf("relation action audit log missing: %+v", auditResp)
	}
	for _, row := range auditResp.Items {
		details, _ := row.Attributes["details"].(map[string]any)
		if details["action_request_id"] == "" || details["context"] == nil {
			t.Fatalf("audit row missing action details: %+v", row)
		}
	}
}

func TestCmdbRelationActionRequestRequiresExistingRelation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, targetID := createCmdbRelationGraphFixture(t, "cmdb-relation-action-missing")

	router := gin.New()
	router.POST("/api/v1/cmdb/relation-actions", CreateCmdbRelationActionRequest)

	payload := `{
		"action":"detail",
		"instance_id":"` + rootID + `",
		"node_id":"` + targetID + `",
		"object_id":"cmdb-relation-action-missing-db",
		"relation_id":"missing-relation-id"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/relation-actions", strings.NewReader(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb_relation_action_store",
		"cmdb_relation_action_target_contract",
		"missing-relation-id",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"audit_ref":"`,
		`"cmdb.relation.detail.request"`,
		`"queued"`,
		`"running"`,
	})
}

func TestCmdbInstanceTopologyEmptyRelationsRemainBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-compatible-empty-relation")

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/"+instanceID+"/topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		"PENDING",
		"cmdb_relation_graph_contract",
		"safe_to_retry",
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"nodes":[`, `"edges":[`, `"code":0`, `"success":true`})
}

func TestCmdbInstanceTopologyMissingInstanceIs404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/cmdb-missing-inst/topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusNotFound, w.Body.String())
	}
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"nodes":[`, `"edges":[`, `"code":0`, `"success":true`})
}

func TestCmdbInstanceTopologyDanglingRelationIsBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID := createCmdbDanglingRelationFixture(t, "cmdb-relation-dangling")

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/"+rootID+"/topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{"PENDING", "cmdb_relation_graph_contract"})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{`"nodes":[`, `"edges":[`, `"code":0`, `"success":true`})
}
