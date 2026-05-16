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

func TestCmdbMonitorBindingReadRevalidatesRuntimeTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-read-runtime")
	createCmdbMonitorTargetFixture(t, "host-read-runtime-1001")
	bindingID := saveCmdbMonitorBindingRuntimeReadFixture(t, instanceID, "host-read-runtime-1001", "collect:missing-read-runtime-template")

	router := gin.New()
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID, nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusConflict {
		t.Fatalf("list status = %d, want %d, body=%s", listW.Code, http.StatusConflict, listW.Body.String())
	}
	assertCmdbBlockedTopLevelStatus(t, listW.Body.Bytes())
	assertCmdbTestStringContainsAll(t, listW.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.runtime.read.v1",
		"monitor_template_runtime_contract",
		"collect:missing-read-runtime-template",
	})
	assertCmdbTestStringExcludesAll(t, listW.Body.String(), []string{
		`"bindings":[`,
		`"binding":{`,
		`"code":0`,
		"runtime-read-secret-marker",
		`"queued"`,
		`"running"`,
		`"applied"`,
		`"installed"`,
		`"data_arrived"`,
		`"delivered"`,
		`"effective"`,
		`"succeeded"`,
	})

	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID+"/"+bindingID, nil)
	detailW := httptest.NewRecorder()
	router.ServeHTTP(detailW, detailReq)
	if detailW.Code != http.StatusConflict {
		t.Fatalf("detail status = %d, want %d, body=%s", detailW.Code, http.StatusConflict, detailW.Body.String())
	}
	assertCmdbBlockedTopLevelStatus(t, detailW.Body.Bytes())
	assertCmdbTestStringContainsAll(t, detailW.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.runtime.read.v1",
		"cmdb_monitor_binding_runtime_read_contract",
		"monitor_template_runtime_contract",
	})
	assertCmdbTestStringExcludesAll(t, detailW.Body.String(), []string{`"binding":{`, `"code":0`, "runtime-read-secret-marker"})
}

func TestCmdbMonitorBindingReadRevalidatesMonitorTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	instanceID := createCmdbCompatibleFixture(t, "cmdb-monitor-binding-read-host")
	templateID := createCmdbMonitorRuntimeTemplateFixture(t, "cmdb-monitor-binding-read-host-template")
	bindingID := saveCmdbMonitorBindingRuntimeReadFixture(t, instanceID, "missing-read-host-target", templateID)

	router := gin.New()
	router.GET("/api/v1/cmdb/monitor-bindings/*path", GetCmdbMonitorBindingsBlocked)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID, nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusConflict {
		t.Fatalf("list status = %d, want %d, body=%s", listW.Code, http.StatusConflict, listW.Body.String())
	}
	assertCmdbBlockedTopLevelStatus(t, listW.Body.Bytes())
	assertCmdbTestStringContainsAll(t, listW.Body.String(), []string{
		"PENDING",
		"cmdb.monitor_binding.runtime.read.v1",
		"monitor_host_binding_contract",
		"missing-read-host-target",
	})
	assertCmdbTestStringExcludesAll(t, listW.Body.String(), []string{
		`"bindings":[`,
		`"code":0`,
		"runtime-read-secret-marker",
	})

	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/cmdb/monitor-bindings/"+instanceID+"/"+bindingID, nil)
	detailW := httptest.NewRecorder()
	router.ServeHTTP(detailW, detailReq)
	if detailW.Code != http.StatusConflict {
		t.Fatalf("detail status = %d, want %d, body=%s", detailW.Code, http.StatusConflict, detailW.Body.String())
	}
	assertCmdbBlockedTopLevelStatus(t, detailW.Body.Bytes())
	assertCmdbTestStringContainsAll(t, detailW.Body.String(), []string{
		"PENDING",
		"cmdb_monitor_binding_runtime_read_contract",
		"monitor_host_binding_contract",
	})
	assertCmdbTestStringExcludesAll(t, detailW.Body.String(), []string{`"binding":{`, `"code":0`, "runtime-read-secret-marker"})
}

func assertCmdbBlockedTopLevelStatus(t *testing.T, raw []byte) {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("decode blocked response: %v", err)
	}
	if body["status"] != "PENDING" || body["code"] == float64(0) {
		t.Fatalf("response should be top-level blocked, body=%#v", body)
	}
}

func saveCmdbMonitorBindingRuntimeReadFixture(t *testing.T, instanceID, hostID, templateID string) string {
	t.Helper()
	saved, err := store.SaveCmdbMonitorBinding(&model.CmdbMonitorBinding{
		InstanceID:   strings.TrimSpace(instanceID),
		Host:         "runtime-read-host",
		HostID:       strings.TrimSpace(hostID),
		TemplateID:   strings.TrimSpace(templateID),
		CmdbObjectID: "cmdb-monitor-binding-runtime-read",
		CmdbAttrID:   "OS001",
		ServerAttrID: "agent.ip",
		TagsJSON:     `{"token":"runtime-read-secret-marker"}`,
		AuditRef:     "cmdb-monitor-binding-runtime-read-audit",
		Creator:      "test",
		Updater:      "test",
	})
	if err != nil {
		t.Fatalf("save stale cmdb monitor binding: %v", err)
	}
	return saved.ID
}
