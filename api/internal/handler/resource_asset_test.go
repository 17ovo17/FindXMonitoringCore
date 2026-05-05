package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestResourceAndHostAssetAuth(t *testing.T) {
	r := resourceAssetTestRouter(t)
	if w := performResourceAssetRequest(t, r, http.MethodGet, "/resource-groups", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("GET /resource-groups without auth should be 401, got %d", w.Code)
	}
	seedResourceAssetToken("resource-user-token", "u1", "alice", "user")
	w := performResourceAssetRequest(t, r, http.MethodPost, "/resource-groups", "resource-user-token", validResourceGroupInput())
	if w.Code != http.StatusForbidden {
		t.Fatalf("POST /resource-groups with user token should be 403, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestAdminResourceGroupCRUDAndStatusValidation(t *testing.T) {
	r := resourceAssetTestRouter(t)
	seedResourceAssetToken("resource-admin-token", "a1", "root", "admin")

	createResp := performResourceAssetRequest(t, r, http.MethodPost, "/resource-groups", "resource-admin-token", validResourceGroupInput())
	if createResp.Code != http.StatusOK {
		t.Fatalf("create resource group should be 200, got %d body=%s", createResp.Code, createResp.Body.String())
	}
	created := decodeResourceGroup(t, createResp)
	if created.ID == "" || created.Name != "Core hosts" || created.Status != "active" {
		t.Fatalf("created resource group mismatch: %+v", created)
	}
	if !reflect.DeepEqual(created.Tags, []string{"linux", "prod"}) {
		t.Fatalf("resource group tags should be normalized: %+v", created.Tags)
	}

	invalid := validResourceGroupInput()
	invalid.Status = "maintenance"
	invalidResp := performResourceAssetRequest(t, r, http.MethodPut, "/resource-groups/"+created.ID, "resource-admin-token", invalid)
	if invalidResp.Code != http.StatusBadRequest {
		t.Fatalf("invalid status should be 400, got %d body=%s", invalidResp.Code, invalidResp.Body.String())
	}

	getResp := performResourceAssetRequest(t, r, http.MethodGet, "/resource-groups/"+created.ID, "resource-admin-token", nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get resource group should be 200, got %d", getResp.Code)
	}
	target, err := store.UpsertMonitorTarget(&model.MonitorTarget{Ident: "group-bound", IP: "10.20.2.1", Labels: map[string]string{hostLabelResourceGroupID: created.ID}})
	if err != nil {
		t.Fatalf("seed bound target failed: %v", err)
	}
	inUseResp := performResourceAssetRequest(t, r, http.MethodDelete, "/resource-groups/"+created.ID, "resource-admin-token", nil)
	if inUseResp.Code != http.StatusConflict || strings.Contains(inUseResp.Body.String(), created.ID) {
		t.Fatalf("delete in-use group should be sanitized 409, got %d body=%s", inUseResp.Code, inUseResp.Body.String())
	}
	if deleted, err := store.DeleteMonitorTarget(target.ID); err != nil || !deleted {
		t.Fatalf("cleanup bound target failed: deleted=%v err=%v", deleted, err)
	}
	deleteResp := performResourceAssetRequest(t, r, http.MethodDelete, "/resource-groups/"+created.ID, "resource-admin-token", nil)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("delete resource group should be 200, got %d", deleteResp.Code)
	}
}

func TestHostAssetsFromHeartbeatAndAssignments(t *testing.T) {
	r := resourceAssetTestRouter(t)
	seedResourceAssetToken("host-admin-token", "a1", "root", "admin")
	_, target, err := store.UpsertFindXAgentHeartbeat(model.FindXAgentHeartbeat{
		Ident:    "asset-1",
		IP:       "10.20.1.2",
		Hostname: "host-one",
		OS:       "linux",
		Arch:     "amd64",
		Version:  "1.2.3",
		GlobalLabels: map[string]string{
			"env":      "prod",
			"password": "secret-value",
		},
	})
	if err != nil {
		t.Fatalf("seed heartbeat failed: %v", err)
	}

	listResp := performResourceAssetRequest(t, r, http.MethodGet, "/host-assets", "host-admin-token", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list host assets should be 200, got %d body=%s", listResp.Code, listResp.Body.String())
	}
	assets := decodeHostAssets(t, listResp)
	if len(assets) != 1 || assets[0].AgentVersion != "1.2.3" || assets[0].Labels["env"] != "prod" {
		t.Fatalf("host asset should include agent summary and safe labels: %+v", assets)
	}
	if _, ok := assets[0].Labels["password"]; ok {
		t.Fatalf("host asset should not expose sensitive label keys: %+v", assets[0].Labels)
	}

	group := createResourceGroupForTest(t, r, "host-admin-token")
	updateTags := performResourceAssetRequest(t, r, http.MethodPut, "/host-assets/"+target.ID+"/tags", "host-admin-token", model.HostTagsInput{Tags: []string{" prod ", "db", "prod"}})
	if updateTags.Code != http.StatusOK {
		t.Fatalf("update tags should be 200, got %d body=%s", updateTags.Code, updateTags.Body.String())
	}
	tagged := decodeHostAsset(t, updateTags)
	if !reflect.DeepEqual(tagged.Tags, []string{"db", "prod"}) {
		t.Fatalf("host tags should be normalized, got %+v", tagged.Tags)
	}

	bindGroup := performResourceAssetRequest(t, r, http.MethodPut, "/host-assets/"+target.ID+"/resource-group", "host-admin-token", model.HostResourceGroupInput{ResourceGroupID: group.ID})
	if bindGroup.Code != http.StatusOK {
		t.Fatalf("bind resource group should be 200, got %d body=%s", bindGroup.Code, bindGroup.Body.String())
	}
	boundGroup := decodeHostAsset(t, bindGroup)
	if boundGroup.ResourceGroupID != group.ID {
		t.Fatalf("resource group binding mismatch: %+v", boundGroup)
	}

	bindWorkspace := performResourceAssetRequest(t, r, http.MethodPut, "/host-assets/"+target.ID+"/workspace", "host-admin-token", model.HostWorkspaceInput{WorkspaceID: " ws-1 "})
	if bindWorkspace.Code != http.StatusOK {
		t.Fatalf("bind workspace should be 200, got %d body=%s", bindWorkspace.Code, bindWorkspace.Body.String())
	}
	boundWorkspace := decodeHostAsset(t, bindWorkspace)
	if boundWorkspace.WorkspaceID != "ws-1" {
		t.Fatalf("workspace binding should be trimmed: %+v", boundWorkspace)
	}

	other, err := store.UpsertMonitorTarget(&model.MonitorTarget{
		Ident:  "asset-2",
		IP:     "10.20.9.9",
		Status: "offline",
		Labels: map[string]string{
			hostLabelWorkspaceID:     "ws-2",
			hostLabelResourceGroupID: "rg-other",
			hostLabelTags:            "cache,stage",
		},
	})
	if err != nil || other.ID == "" {
		t.Fatalf("seed second target failed: target=%+v err=%v", other, err)
	}
	assertHostAssetIDs(t, r, "host-admin-token", "/host-assets?workspace_id=ws-1&resource_group_id="+group.ID+"&tag=db", []string{target.ID})
	assertHostAssetIDs(t, r, "host-admin-token", "/host-assets?tags=stage&status=offline&online=false", []string{other.ID})
	assertHostAssetIDs(t, r, "host-admin-token", "/host-assets?online=true&keyword=host-one", []string{target.ID})
	assertHostAssetIDs(t, r, "host-admin-token", "/host-assets?keyword=10.20.9", []string{other.ID})
}

func TestHostAssetMissingAndUnknownResourceGroup(t *testing.T) {
	r := resourceAssetTestRouter(t)
	seedResourceAssetToken("missing-admin-token", "a1", "root", "admin")
	resp := performResourceAssetRequest(t, r, http.MethodGet, "/host-assets/missing", "missing-admin-token", nil)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("missing host should be 404, got %d", resp.Code)
	}
	target, err := store.UpsertMonitorTarget(&model.MonitorTarget{Ident: "asset-404", IP: "10.20.1.9"})
	if err != nil {
		t.Fatalf("seed target failed: %v", err)
	}
	bindResp := performResourceAssetRequest(t, r, http.MethodPut, "/host-assets/"+target.ID+"/resource-group", "missing-admin-token", model.HostResourceGroupInput{ResourceGroupID: "rg-missing"})
	if bindResp.Code != http.StatusNotFound {
		t.Fatalf("unknown resource group should be 404, got %d", bindResp.Code)
	}
}

func TestResourceAssetResponseDoesNotExposeSensitiveText(t *testing.T) {
	r := resourceAssetTestRouter(t)
	seedResourceAssetToken("sanitize-admin-token", "a1", "root", "admin")
	_, _, err := store.UpsertFindXAgentHeartbeat(model.FindXAgentHeartbeat{
		Ident:        "asset-sensitive",
		IP:           "10.20.1.10",
		Hostname:     "sensitive-host",
		GlobalLabels: map[string]string{"api_token": "super-secret-token"},
	})
	if err != nil {
		t.Fatalf("seed heartbeat failed: %v", err)
	}
	resp := performResourceAssetRequest(t, r, http.MethodGet, "/host-assets", "sanitize-admin-token", nil)
	body := strings.ToLower(resp.Body.String())
	for _, forbidden := range []string{"super-secret-token", "authorization", "cookie", "dsn"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response should not expose %q: %s", forbidden, resp.Body.String())
		}
	}
}

func TestResourceGroupStorageErrorMessageIsSanitized(t *testing.T) {
	body, err := json.Marshal(gin.H{"error": resourceGroupStorageError})
	if err != nil {
		t.Fatalf("marshal storage error response: %v", err)
	}
	lower := strings.ToLower(string(body))
	if !strings.Contains(lower, "storage unavailable") {
		t.Fatalf("storage error should be actionable and sanitized: %s", string(body))
	}
	for _, forbidden := range []string{"select ", "replace into", "delete from", "mysql", "dsn", "password", "token"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("storage error should not expose %q: %s", forbidden, string(body))
		}
	}
}

func TestResourceGroupStoreDoesNotIgnoreCriticalErrors(t *testing.T) {
	sourcePath := filepath.Clean("../store/resource_asset.go")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read resource group store source: %v", err)
	}
	content := string(source)
	for _, forbidden := range []string{
		"_ = json.Unmarshal",
		"_, _ = db.Exec",
		"return model.ResourceGroup{}, false, nil\n\t\t\t}",
		"continue\n\t\t}",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("resource group store must not ignore critical errors with %q", forbidden)
		}
	}
}

func resourceAssetTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	clearResourceAssetState(t)
	t.Cleanup(func() {
		clearResourceAssetState(t)
	})

	r := gin.New()
	r.GET("/resource-groups", RequireAuth(), ListResourceGroups)
	r.POST("/resource-groups", RequireRole("admin"), CreateResourceGroup)
	r.GET("/resource-groups/:id", RequireAuth(), GetResourceGroup)
	r.PUT("/resource-groups/:id", RequireRole("admin"), UpdateResourceGroup)
	r.DELETE("/resource-groups/:id", RequireRole("admin"), DeleteResourceGroup)
	r.GET("/host-assets", RequireAuth(), ListHostAssets)
	r.GET("/host-assets/:id", RequireAuth(), GetHostAsset)
	r.PUT("/host-assets/:id/tags", RequireRole("admin"), UpdateHostAssetTags)
	r.PUT("/host-assets/:id/resource-group", RequireRole("admin"), UpdateHostAssetResourceGroup)
	r.PUT("/host-assets/:id/workspace", RequireRole("admin"), UpdateHostAssetWorkspace)
	return r
}

func validResourceGroupInput() model.ResourceGroupInput {
	return model.ResourceGroupInput{
		Name:        "Core hosts",
		Description: "FindX Core host grouping",
		WorkspaceID: "ws-core",
		Status:      "active",
		Tags:        []string{" prod ", "linux", "prod"},
	}
}

func performResourceAssetRequest(t *testing.T, r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func seedResourceAssetToken(token, userID, username, role string) {
	tokenStore.Store(token, tokenEntry{userID: userID, username: username, role: role, expiresAt: time.Now().Add(time.Hour)})
}

func clearResourceAssetState(t *testing.T) {
	t.Helper()
	tokenStore.Range(func(key, _ any) bool {
		tokenStore.Delete(key)
		return true
	})
	for _, group := range store.ListResourceGroups() {
		if _, err := store.DeleteResourceGroup(group.ID); err != nil {
			t.Fatalf("delete resource group %s during cleanup: %v", group.ID, err)
		}
	}
	for _, target := range store.ListMonitorTargets() {
		if _, err := store.DeleteMonitorTarget(target.ID); err != nil {
			t.Fatalf("delete monitor target %s during cleanup: %v", target.ID, err)
		}
	}
}

func decodeResourceGroup(t *testing.T, w *httptest.ResponseRecorder) model.ResourceGroup {
	t.Helper()
	var resp model.ResourceGroup
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode resource group: %v body=%s", err, w.Body.String())
	}
	return resp
}

func decodeHostAssets(t *testing.T, w *httptest.ResponseRecorder) []model.HostAsset {
	t.Helper()
	var resp []model.HostAsset
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode host assets: %v body=%s", err, w.Body.String())
	}
	return resp
}

func decodeHostAsset(t *testing.T, w *httptest.ResponseRecorder) model.HostAsset {
	t.Helper()
	var resp model.HostAsset
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode host asset: %v body=%s", err, w.Body.String())
	}
	return resp
}

func createResourceGroupForTest(t *testing.T, r *gin.Engine, token string) model.ResourceGroup {
	t.Helper()
	resp := performResourceAssetRequest(t, r, http.MethodPost, "/resource-groups", token, validResourceGroupInput())
	if resp.Code != http.StatusOK {
		t.Fatalf("create resource group failed: %d %s", resp.Code, resp.Body.String())
	}
	return decodeResourceGroup(t, resp)
}

func assertHostAssetIDs(t *testing.T, r *gin.Engine, token, path string, want []string) {
	t.Helper()
	resp := performResourceAssetRequest(t, r, http.MethodGet, path, token, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("list host assets %s should be 200, got %d body=%s", path, resp.Code, resp.Body.String())
	}
	assets := decodeHostAssets(t, resp)
	got := make([]string, 0, len(assets))
	for _, asset := range assets {
		got = append(got, asset.HostID)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("host asset filter %s mismatch: got=%+v want=%+v assets=%+v", path, got, want, assets)
	}
}
