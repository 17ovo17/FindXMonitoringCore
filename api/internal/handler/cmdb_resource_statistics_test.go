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

func TestCmdbResourceStatisticsAggregatesRealStoresAndContractGaps(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/cmdb/resource-statistics", GetCmdbResourceStatistics)

	objectID := "stats-object-128m"
	if err := store.CreateCmdbObject(&model.CmdbObject{ID: objectID, Name: "128M 应用", CategoryID: "stats-cat", ObjectType: 101}); err != nil {
		t.Fatalf("create cmdb object: %v", err)
	}
	for _, inst := range []model.CmdbInstance{
		{ID: "stats-instance-128m-a", ObjectID: objectID, Data: `{"name":"stats-a","ip_address":"10.128.0.1"}`, Creator: "test", Updater: "test"},
		{ID: "stats-instance-128m-b", ObjectID: objectID, Data: `{"name":"stats-b","ip_address":"10.128.0.2"}`, Creator: "test", Updater: "test"},
	} {
		item := inst
		if err := store.CreateCmdbInstance(&item); err != nil {
			t.Fatalf("create cmdb instance %s: %v", item.ID, err)
		}
	}
	if err := store.CreateCmdbRelationType(&model.CmdbRelationType{ID: "stats-rel-128m", Name: "depends_on", Label: "依赖"}); err != nil {
		t.Fatalf("create relation type: %v", err)
	}
	if err := store.CreateCmdbInstanceRelation(&model.CmdbInstanceRelation{ID: "stats-edge-128m", SourceInstanceID: "stats-instance-128m-a", TargetInstanceID: "stats-instance-128m-b", RelationTypeID: "stats-rel-128m"}); err != nil {
		t.Fatalf("create instance relation: %v", err)
	}
	savedGroup, err := store.SaveResourceGroup(model.ResourceGroup{Name: "128M 资源组", Status: "active"})
	if err != nil {
		t.Fatalf("save resource group: %v", err)
	}
	business := store.SaveTopologyBusiness(model.TopologyBusiness{ID: "stats-business-128m", Name: "128M 业务组", Hosts: []string{"10.128.0.1"}, Attributes: map[string]string{"status": "active"}})
	t.Cleanup(func() {
		store.DeleteTopologyBusiness(business.ID)
		if savedGroup.ID != "" {
			_, _ = store.DeleteResourceGroup(savedGroup.ID)
		}
	})
	_, target, err := store.UpsertFindXAgentHeartbeat(model.FindXAgentHeartbeat{
		Ident:         "stats-agent-128m",
		IP:            "10.128.0.1",
		Hostname:      "stats-agent-host",
		Version:       "1.2.8",
		ConfigVersion: "cfg-128m",
	})
	if err != nil {
		t.Fatalf("seed heartbeat: %v", err)
	}
	t.Cleanup(func() {
		if target != nil {
			_, _ = store.DeleteMonitorTarget(target.ID)
		}
		_, _ = store.DeleteFindXAgent("stats-agent-128m")
	})
	target.Labels[hostLabelWorkspaceID] = business.ID
	target.Labels[hostLabelResourceGroupID] = savedGroup.ID
	if _, err := store.UpsertMonitorTarget(target); err != nil {
		t.Fatalf("update target labels: %v", err)
	}
	if _, err := store.SaveCmdbMonitorBinding(&model.CmdbMonitorBinding{
		ID:              "stats-binding-128m",
		InstanceID:      "stats-instance-128m-a",
		HostID:          target.ID,
		TemplateID:      "stats-template-128m",
		CmdbObjectID:    objectID,
		CmdbAttrID:      "ip_address",
		ServerAttrID:    "agent.ip",
		ActiveStatus:    "active",
		ServerModelID:   "server-model-128m",
		ServerModelName: "stats server",
	}); err != nil {
		t.Fatalf("save monitor binding: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/resource-statistics", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("resource statistics should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp model.CmdbResourceStatisticsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode statistics response: %v body=%s", err, w.Body.String())
	}
	if resp.Contract != "cmdb.resource.statistics.read.v1" || resp.Status != "ready_with_contract_gaps" {
		t.Fatalf("unexpected contract status: %+v", resp)
	}
	if resp.Totals.CmdbInstances < 2 || resp.Totals.CmdbModels < 1 || resp.Totals.HostAssets < 1 || resp.Totals.FindXAgents < 1 {
		t.Fatalf("totals should aggregate real stores: %+v", resp.Totals)
	}
	if resp.Totals.MonitorBindingRows < 1 || resp.Totals.MonitorBoundInstances < 1 || resp.Totals.RelationEdges < 1 {
		t.Fatalf("binding/relation totals should be real: %+v", resp.Totals)
	}
	if !statsDimensionHas(resp.ModelDistribution, objectID, 2) {
		t.Fatalf("model distribution should include seeded object: %+v", resp.ModelDistribution)
	}
	if !statsDimensionHas(resp.BusinessGroupDistribution, business.ID, 1) {
		t.Fatalf("business distribution should include seeded business group: %+v", resp.BusinessGroupDistribution)
	}
	if !statsDimensionHas(resp.ResourceGroupDistribution, savedGroup.ID, 1) {
		t.Fatalf("resource group distribution should include seeded resource group: %+v", resp.ResourceGroupDistribution)
	}
	if !statsContractGapHas(resp.BlockedContracts, "cmdb_resource_approval_runtime_contract") ||
		!statsContractGapHas(resp.BlockedContracts, "cmdb_dashboard_import_runtime_contract") ||
		!statsContractGapHas(resp.BlockedContracts, "cmdb_operation_risk_policy_contract") {
		t.Fatalf("blocked contracts should expose missing runtime gaps: %+v", resp.BlockedContracts)
	}
	body := strings.ToLower(w.Body.String())
	for _, forbidden := range []string{"secret-marker", "password", "token", "cookie", "queued", "running", "installed", "data_arrived", "succeeded"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("statistics response should not expose fake state or sensitive marker %q: %s", forbidden, w.Body.String())
		}
	}
}

func statsDimensionHas(rows []model.CmdbResourceStatisticsDimension, key string, minCount int) bool {
	for _, row := range rows {
		if row.Key == key && row.Count >= minCount {
			return true
		}
	}
	return false
}

func statsContractGapHas(rows []model.CmdbResourceStatisticsContractGap, id string) bool {
	for _, row := range rows {
		if row.ID == id && row.Status == "PENDING" {
			return true
		}
	}
	return false
}
