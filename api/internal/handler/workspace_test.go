package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func TestWorkspaceEndpointsRequireAuthAndAdminRole(t *testing.T) {
	r := workspaceTestRouter(t)

	if w := performWorkspaceRequest(r, http.MethodGet, "/workspaces", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("GET /workspaces without auth should be 401, got %d body=%s", w.Code, w.Body.String())
	}

	seedWorkspaceToken("workspace-user-token", "u1", "alice", "user")
	w := performWorkspaceRequest(r, http.MethodPost, "/workspaces", "workspace-user-token", validWorkspaceInput())
	if w.Code != http.StatusForbidden {
		t.Fatalf("POST /workspaces with user token should be 403, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestWorkspaceAdminCreateValidateGetAndDelete(t *testing.T) {
	r := workspaceTestRouter(t)
	seedWorkspaceToken("workspace-admin-token", "a1", "root", "admin")

	createResp := performWorkspaceRequest(r, http.MethodPost, "/workspaces", "workspace-admin-token", validWorkspaceInput())
	if createResp.Code != http.StatusOK {
		t.Fatalf("admin POST /workspaces should be 200, got %d body=%s", createResp.Code, createResp.Body.String())
	}

	created := decodeWorkspaceResponse(t, createResp)
	t.Cleanup(func() {
		if created.ID != "" {
			store.DeleteTopologyBusiness(created.ID)
		}
	})
	assertCreatedWorkspace(t, created)

	invalidUpdate := validWorkspaceInput()
	invalidUpdate.Status = "maintenance"
	updateResp := performWorkspaceRequest(r, http.MethodPut, "/workspaces/"+created.ID, "workspace-admin-token", invalidUpdate)
	if updateResp.Code != http.StatusBadRequest {
		t.Fatalf("admin PUT invalid status should be 400, got %d body=%s", updateResp.Code, updateResp.Body.String())
	}

	missingResp := performWorkspaceRequest(r, http.MethodGet, "/workspaces/workspace-missing-id", "workspace-admin-token", nil)
	if missingResp.Code != http.StatusNotFound {
		t.Fatalf("GET missing workspace should be 404, got %d body=%s", missingResp.Code, missingResp.Body.String())
	}

	deleteResp := performWorkspaceRequest(r, http.MethodDelete, "/workspaces/"+created.ID, "workspace-admin-token", nil)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("admin DELETE /workspaces/:id should be 200, got %d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	if _, ok := store.GetTopologyBusiness(created.ID); ok {
		t.Fatalf("workspace %q should be removed from store", created.ID)
	}
	created.ID = ""
}

func workspaceTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	clearWorkspaceTokens()
	t.Cleanup(clearWorkspaceTokens)

	r := gin.New()
	r.GET("/workspaces", RequireAuth(), ListWorkspaces)
	r.POST("/workspaces", RequireRole("admin"), CreateWorkspace)
	r.GET("/workspaces/:id", RequireAuth(), GetWorkspace)
	r.PUT("/workspaces/:id", RequireRole("admin"), UpdateWorkspace)
	r.DELETE("/workspaces/:id", RequireRole("admin"), DeleteWorkspace)
	return r
}

func validWorkspaceInput() model.WorkspaceInput {
	return model.WorkspaceInput{
		Name:        "P0-T1 Test Workspace",
		Description: "workspace handler test",
		Owner:       "qa",
		Status:      "active",
		Tags:        []string{"prod", "critical", "prod"},
		Hosts:       []string{"10.10.1.10", "10.10.1.11", "10.10.1.10"},
		Endpoints: []model.TopologyEndpoint{
			{IP: "10.10.1.10", Port: 8080, ServiceName: "api", Protocol: "http"},
			{IP: "10.10.1.12", Port: 3306, ServiceName: "mysql", Protocol: "tcp"},
		},
	}
}

func performWorkspaceRequest(r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
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

func seedWorkspaceToken(token, userID, username, role string) {
	tokenStore.Store(token, tokenEntry{userID: userID, username: username, role: role, expiresAt: time.Now().Add(time.Hour)})
}

func clearWorkspaceTokens() {
	tokenStore.Range(func(key, _ any) bool {
		tokenStore.Delete(key)
		return true
	})
}

func decodeWorkspaceResponse(t *testing.T, w *httptest.ResponseRecorder) model.Workspace {
	t.Helper()
	var resp model.Workspace
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode workspace response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func assertCreatedWorkspace(t *testing.T, got model.Workspace) {
	t.Helper()
	if got.ID == "" {
		t.Fatalf("created workspace should include id: %+v", got)
	}
	if got.Name != "P0-T1 Test Workspace" {
		t.Fatalf("created workspace name mismatch: %+v", got)
	}
	if got.Status != "active" {
		t.Fatalf("created workspace status mismatch: %+v", got)
	}
	if !reflect.DeepEqual(got.Tags, []string{"critical", "prod"}) {
		t.Fatalf("created workspace tags should be normalized, got %+v", got.Tags)
	}
	if !reflect.DeepEqual(got.Hosts, []string{"10.10.1.10", "10.10.1.11"}) {
		t.Fatalf("created workspace hosts should be normalized, got %+v", got.Hosts)
	}
	if got.ResourceCount != 5 {
		t.Fatalf("created workspace resource_count should be 5, got %d", got.ResourceCount)
	}
}
