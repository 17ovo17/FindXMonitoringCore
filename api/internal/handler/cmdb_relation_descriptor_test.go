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

func TestCmdbRelationDescriptorPreservesMatureCaptureFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prefix := "cmdb-relation-descriptor"
	rootID, targetID := createCmdbRelationDescriptorFixture(t, prefix)

	router := gin.New()
	router.GET("/api/v1/cmdb/instances/:id/topology", GetCmdbInstanceTopologyBlocked)
	router.POST("/api/v1/cmdb/relation-actions", CreateCmdbRelationActionRequest)

	topologyReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/instances/"+rootID+"/topology", nil)
	topologyW := httptest.NewRecorder()
	router.ServeHTTP(topologyW, topologyReq)
	if topologyW.Code != http.StatusOK {
		t.Fatalf("topology status = %d, want %d, body=%s", topologyW.Code, http.StatusOK, topologyW.Body.String())
	}
	assertCmdbTestStringContainsAll(t, topologyW.Body.String(), []string{
		`"asst_id":"belong"`,
		`"asst_name":"belongs to"`,
		`"mapping":"n:1"`,
		`"visible":true`,
		`"rule_logic":"and"`,
		`"rule_expression":"A"`,
		`"left_attr":"host_ip"`,
		`"right_attr":"IP"`,
		`"path_label":"descriptor-app-(belongs to)descriptor-database"`,
		`"relation_schema"`,
	})
	assertCmdbTestStringExcludesAll(t, topologyW.Body.String(), []string{
		`"nodes":[]`,
		`"edges":[]`,
		`"queued"`,
		`"running"`,
		`"applied"`,
	})

	payload := `{
		"action":"relation",
		"instance_id":"` + rootID + `",
		"node_id":"` + targetID + `",
		"object_id":"` + prefix + `-db",
		"relation_id":"` + prefix + `-rel",
		"context":{"source":"descriptor parity","secret":"descriptor-secret-marker"}
	}`
	actionReq := httptest.NewRequest(http.MethodPost, "/api/v1/cmdb/relation-actions", strings.NewReader(payload))
	actionReq.Header.Set("Content-Type", "application/json")
	actionW := httptest.NewRecorder()
	router.ServeHTTP(actionW, actionReq)
	if actionW.Code != http.StatusAccepted {
		t.Fatalf("action status = %d, want %d, body=%s", actionW.Code, http.StatusAccepted, actionW.Body.String())
	}
	assertCmdbTestStringContainsAll(t, actionW.Body.String(), []string{
		`"relation_schema"`,
		`"instance_relation_id":"` + prefix + `-rel"`,
		`"asst_id":"belong"`,
		`"mapping":"n:1"`,
		`"rule_expression":"A"`,
	})
	assertCmdbTestStringExcludesAll(t, actionW.Body.String(), []string{
		"descriptor-secret-marker",
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	})

	var body map[string]any
	if err := json.Unmarshal(actionW.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode action response: %v", err)
	}
	action := body["action_request"].(map[string]any)
	receipts := action["receipts"].([]any)
	if len(receipts) != 2 {
		t.Fatalf("receipts len = %d, want 2: %#v", len(receipts), receipts)
	}
	for _, raw := range receipts {
		receipt := raw.(map[string]any)
		ref := strings.TrimSpace(receipt["request_ref"].(string))
		task, ok, err := store.GetFindXAgentExecutionTask(ref)
		if err != nil {
			t.Fatalf("get execution task %s: %v", ref, err)
		}
		if !ok {
			t.Fatalf("execution task %s not found", ref)
		}
		if task.Metadata["cmdb_relation_asst_id"] != "belong" ||
			task.Metadata["cmdb_relation_mapping"] != "n:1" ||
			task.Metadata["cmdb_relation_visible"] != "true" {
			t.Fatalf("task metadata missing mature relation fields: %#v", task.Metadata)
		}
	}
}

func createCmdbRelationDescriptorFixture(t *testing.T, prefix string) (string, string) {
	t.Helper()
	rootObjectID := prefix + "-app"
	targetObjectID := prefix + "-db"
	for _, obj := range []model.CmdbObject{
		{ID: rootObjectID, Name: "descriptor-app", CategoryID: prefix + "-cat", ObjectType: 101},
		{ID: targetObjectID, Name: "descriptor-database", CategoryID: prefix + "-cat", ObjectType: 101},
	} {
		obj := obj
		if err := store.CreateCmdbObject(&obj); err != nil {
			t.Fatalf("create object %s: %v", obj.ID, err)
		}
	}
	root := &model.CmdbInstance{ObjectID: rootObjectID, Data: `{"name":"api-service"}`, Creator: "test", Updater: "test"}
	if err := store.CreateCmdbInstance(root); err != nil {
		t.Fatalf("create root instance: %v", err)
	}
	target := &model.CmdbInstance{ObjectID: targetObjectID, Data: `{"name":"mysql-primary"}`, Creator: "test", Updater: "test"}
	if err := store.CreateCmdbInstance(target); err != nil {
		t.Fatalf("create target instance: %v", err)
	}
	visible := true
	if err := store.CreateCmdbRelationType(&model.CmdbRelationType{
		ID:             prefix + "-object-relation",
		Name:           "belong",
		Label:          "belongs to",
		Mapping:        "n:1",
		Visible:        &visible,
		Description:    "mature cmdb relation descriptor",
		RuleLogic:      "and",
		RuleExpression: "A",
		RulesJSON:      `[{"left_attr":"host_ip","logic":"=","right_attr":"IP","tag":"A"}]`,
		LeftMin:        0,
		LeftMax:        1,
		RightMin:       0,
		RightMax:       -1,
		Source:         2,
		LeftAsstName:   "belongs to",
		RightAsstName:  "belongs to",
	}); err != nil {
		t.Fatalf("create relation type: %v", err)
	}
	if err := store.CreateCmdbInstanceRelation(&model.CmdbInstanceRelation{
		ID:               prefix + "-rel",
		SourceInstanceID: root.ID,
		TargetInstanceID: target.ID,
		RelationTypeID:   prefix + "-object-relation",
	}); err != nil {
		t.Fatalf("create instance relation: %v", err)
	}
	return root.ID, target.ID
}
