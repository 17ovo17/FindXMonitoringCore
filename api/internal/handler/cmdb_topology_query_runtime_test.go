package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCmdbInstanceTopologyRelationQueryRuntimeRemainsBlockedByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, _, _ := createCmdbRecursiveRelationFixture(t, "cmdb-relation-query-runtime")

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
	if body["status"] != "ready_with_contract_gaps" || body["code"] != float64(0) {
		t.Fatalf("real topology graph should remain readable with contract gaps while query executor is blocked: %#v", body)
	}
	runtime, ok := body["relation_query_runtime"].(map[string]any)
	if !ok {
		t.Fatalf("relation_query_runtime missing: %s", w.Body.String())
	}
	if runtime["status"] != "PENDING" || runtime["contract"] != "cmdb.relation_query.runtime.read.v1" {
		t.Fatalf("relation_query_runtime should be blocked contract, got %#v", runtime)
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		`"relation_queries"`,
		`"relation_query_rules"`,
		`"relation_query_runtime"`,
		`"cmdb.relation_query.runtime.read.v1"`,
		`"cmdb_relation_query_executor_contract"`,
		`"cmdb_relation_recursive_path_contract"`,
		`"cmdb_topology_layout_coordinate_contract"`,
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"relation_query_results"`,
		`"recursive_paths":[`,
		`"queued"`,
		`"running"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	})
}

func TestCmdbInstanceTopologyRelationQueryRuntimeExecuteFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, _, _ := createCmdbRecursiveRelationFixture(t, "cmdb-relation-query-runtime-exec")

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/"+rootID+"/topology?relation_query=execute", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	runtime, ok := body["relation_query_runtime"].(map[string]any)
	if !ok {
		t.Fatalf("relation_query_runtime missing: %s", w.Body.String())
	}
	if runtime["status"] != "PENDING" || runtime["contract"] != "cmdb.relation_query.runtime.read.v1" || runtime["execute_requested"] != true {
		t.Fatalf("relation_query_runtime should remain blocked when execute is explicit, got %#v", runtime)
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		`"relation_queries"`,
		`"relation_query_runtime"`,
		`"cmdb.relation_query.runtime.read.v1"`,
		`"execute_requested":true`,
		`"cmdb_relation_query_executor_contract"`,
		`"cmdb_relation_recursive_path_contract"`,
		`"cmdb_topology_layout_coordinate_contract"`,
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"relation_query_results"`,
		`"executor":"findx_cmdb_relation_query_runtime"`,
		`"recursive_paths":[`,
		`"layout_coordinates":[{`,
		`"nodes":[]`,
		`"edges":[]`,
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
		"leaf-token-marker",
	})
}

func TestCmdbInstanceTopologyRecursivePathLayoutExecutorContractFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, middleID, leafID := createCmdbRecursiveRelationFixture(t, "cmdb-relation-layout-executor")

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/"+rootID+"/topology?relation_query=execute", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "ready_with_contract_gaps" {
		t.Fatalf("topology should expose relation rows without claiming final layout readiness: %#v", body["status"])
	}
	if body["code"] != float64(0) {
		t.Fatalf("verified relation rows should remain readable, got code=%#v", body["code"])
	}
	nodes := body["nodes"].([]any)
	edges := body["edges"].([]any)
	if len(nodes) != 3 || len(edges) != 2 {
		t.Fatalf("verified recursive graph rows missing nodes=%#v edges=%#v", nodes, edges)
	}
	runtime := body["relation_query_runtime"].(map[string]any)
	if runtime["status"] != "PENDING" || runtime["contract"] != "cmdb.relation_query.runtime.read.v1" {
		t.Fatalf("runtime should fail closed without executors: %#v", runtime)
	}
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		`"path":["` + rootID + `","` + middleID + `"]`,
		`"path":["` + rootID + `","` + middleID + `","` + leafID + `"]`,
		`"relation_path_levels"`,
		`"ready_recursive_from_rows"`,
		`"topology_layout_coordinate"`,
		`"blocked_by_executor"`,
		`"cmdb_relation_recursive_path_contract"`,
		`"cmdb_topology_layout_coordinate_contract"`,
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"topology_success_response"`,
		`"topology_layout","status":"ready_stable"`,
		`"relation_query_results"`,
		`"recursive_paths":[`,
		`"layout_coordinates":[{`,
		`"nodes":[]`,
		`"edges":[]`,
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"data_arrived"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
		`"success"`,
		"leaf-token-marker",
	})
}
