package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCmdbContractProbeSeedReturnsReadyMatureTopology(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store.SeedCmdbContractProbeTopology()

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/contract-probe/topology", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["code"] != float64(0) || body["status"] != "ready_with_contract_gaps" {
		t.Fatalf("contract-probe should expose relation-backed topology with contract gaps: %#v", body)
	}
	nodes := body["nodes"].([]any)
	edges := body["edges"].([]any)
	if len(nodes) < 2 || len(edges) != 1 {
		t.Fatalf("contract-probe graph shape mismatch nodes=%#v edges=%#v", nodes, edges)
	}
	edge := edges[0].(map[string]any)
	if edge["relation_id"] != "OperatingSystem1_default_j6p8Wb2xkV1666171515" {
		t.Fatalf("edge relation id mismatch: %#v", edge)
	}
	if edge["asst_id"] != "default" || edge["mapping"] != "n:1" || edge["visible"] != true {
		t.Fatalf("edge mature relation fields missing: %#v", edge)
	}
	schema := edge["relation_schema"].(map[string]any)
	for _, key := range []string{"left_object_id", "right_object_id", "left_id", "right_id", "asst_id", "asst_name", "mapping", "visible", "rule_logic", "rule_expression", "rules"} {
		if _, ok := schema[key]; !ok {
			t.Fatalf("relation_schema missing %s: %#v", key, schema)
		}
	}
	text := w.Body.String()
	for _, want := range []string{
		`"path_label"`,
		`"contract_gap_id"`,
		`"relation_descriptor_matrix"`,
		`"ready_from_relation_type"`,
		`"topology_layout_coordinate"`,
		`"blocked_by_executor"`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("contract-probe topology missing %s in %s", want, text)
		}
	}
	for _, forbidden := range []string{
		`"nodes":[]`,
		`"edges":[]`,
		"password",
		"token",
		"dsn",
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("contract-probe topology leaked forbidden marker %q in %s", forbidden, text)
		}
	}
}
