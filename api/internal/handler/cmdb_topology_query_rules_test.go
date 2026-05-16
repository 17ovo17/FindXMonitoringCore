package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCmdbInstanceTopologyReturnsMatureRelationQueryRules(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rootID, middleID, leafID := createCmdbRecursiveRelationFixture(t, "cmdb-relation-query-rules")

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
	queries, ok := body["relation_queries"].([]any)
	if !ok || len(queries) == 0 {
		t.Fatalf("relation_queries missing or empty: %#v", body["relation_queries"])
	}

	leafObjectID := "cmdb-relation-query-rules-cache"
	var leafQuery map[string]any
	for _, raw := range queries {
		query := raw.(map[string]any)
		if query["end_vertex"] == leafObjectID {
			leafQuery = query
			break
		}
	}
	if leafQuery == nil {
		t.Fatalf("leaf query rule missing for %s: %#v", leafObjectID, queries)
	}
	if leafQuery["start_vertex"] != "cmdb-relation-query-rules-app" || leafQuery["end_vertex"] != leafObjectID {
		t.Fatalf("start/end vertex mismatch: %#v", leafQuery)
	}
	for _, key := range []string{"_id", "name", "path_id", "path_label", "authority_type", "authority_data", "paths_hash", "paths_updated_at", "start_vertex_name", "end_vertex_name"} {
		if leafQuery[key] == nil || strings.TrimSpace(anyToString(leafQuery[key])) == "" {
			t.Fatalf("query rule missing %s: %#v", key, leafQuery)
		}
	}
	path := leafQuery["path"].([]any)
	if len(path) != 5 {
		t.Fatalf("path length = %d, want 5 alternating object/relation entries: %#v", len(path), path)
	}
	assertCmdbRelationQueryObjectVertex(t, path[0], "cmdb-relation-query-rules-app", "app")
	assertCmdbRelationQueryEdge(t, path[1], "cmdb-relation-query-rules-app", "cmdb-relation-query-rules-db")
	assertCmdbRelationQueryObjectVertex(t, path[2], "cmdb-relation-query-rules-db", "database")
	assertCmdbRelationQueryEdge(t, path[3], "cmdb-relation-query-rules-db", leafObjectID)
	assertCmdbRelationQueryObjectVertex(t, path[4], leafObjectID, "cache")
	assertCmdbTestStringContainsAll(t, w.Body.String(), []string{
		`"relation_query_rules"`,
		`"cmdb.relation.query_rules.v1"`,
		`"path_label":"app-(depends on)database-(depends on)cache"`,
		`"start_instance_id":"` + rootID + `"`,
		`"end_instance_id":"` + leafID + `"`,
		`"via_vertex":["cmdb-relation-query-rules-db"]`,
	})
	assertCmdbTestStringExcludesAll(t, w.Body.String(), []string{
		`"nodes":[]`,
		`"edges":[]`,
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
		middleID + "-secret",
	})
}

func assertCmdbRelationQueryObjectVertex(t *testing.T, raw any, objectID string, name string) {
	t.Helper()
	item := raw.(map[string]any)
	if item["id"] != "1:"+objectID || item["label"] != "object" || item["type"] != "vertex" {
		t.Fatalf("object vertex mismatch for %s: %#v", objectID, item)
	}
	props := item["properties"].(map[string]any)
	if props["object_id"] != objectID || props["name"] != name || props["type"] != float64(101) || props["is_preset"] != false || props["p_object_id"] != "" {
		t.Fatalf("object vertex properties mismatch for %s: %#v", objectID, props)
	}
}

func assertCmdbRelationQueryEdge(t *testing.T, raw any, sourceObjectID string, targetObjectID string) {
	t.Helper()
	item := raw.(map[string]any)
	if item["label"] != "relation" || item["type"] != "edge" {
		t.Fatalf("relation edge type mismatch: %#v", item)
	}
	if item["outV"] != "1:"+sourceObjectID || item["inV"] != "1:"+targetObjectID || item["outVLabel"] != "object" || item["inVLabel"] != "object" {
		t.Fatalf("relation edge endpoint mismatch: %#v", item)
	}
	props := item["properties"].(map[string]any)
	if props["relation_id"] != "depends_on" || props["asst_id"] != "depends_on" || props["mapping"] != "n:1" || props["visible"] != true {
		t.Fatalf("relation edge properties mismatch: %#v", props)
	}
}
