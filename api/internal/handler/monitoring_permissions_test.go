package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestMonitorPermissionEndpointsRequireAuth(t *testing.T) {
	r := monitorPermissionTestRouter()
	if w := performPermissionRequest(r, http.MethodGet, "/monitor/permissions/me", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("me without auth should be 401, got %d", w.Code)
	}
	body := map[string]string{"resource": "monitor.health", "action": "read"}
	if w := performPermissionRequest(r, http.MethodPost, "/monitor/permissions/check", "", body); w.Code != http.StatusUnauthorized {
		t.Fatalf("check without auth should be 401, got %d", w.Code)
	}
}

func TestMonitorPermissionMeForUserIsReadOnly(t *testing.T) {
	r := monitorPermissionTestRouter()
	seedPermissionToken("user-token", "u1", "alice", "user")
	w := performPermissionRequest(r, http.MethodGet, "/monitor/permissions/me", "user-token", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("me should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp model.MonitorPermissionMeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	if resp.Version != model.MonitorPermissionVersion || resp.Mode != model.MonitorPermissionMode {
		t.Fatalf("unexpected version/mode: %+v", resp)
	}
	if resp.User.Username != "alice" || resp.User.Role != "user" {
		t.Fatalf("unexpected user payload: %+v", resp.User)
	}
	if !permissionListed(resp.Permissions, "monitor.query", "execute") {
		t.Fatalf("user should include query execute permission: %+v", resp.Permissions)
	}
	if !permissionListed(resp.Permissions, "cmdb.resource.projection", "read") {
		t.Fatalf("user should include cmdb resource projection read permission: %+v", resp.Permissions)
	}
	if permissionListed(resp.Permissions, "monitor.alert_rule", "tryrun") {
		t.Fatalf("user must not include alert_rule tryrun permission")
	}
	if resp.Matrix["monitor.alert_rule"]["tryrun"] {
		t.Fatalf("user matrix should deny alert_rule tryrun")
	}
	if permissionListed(resp.Permissions, "credential", "read") {
		t.Fatalf("user must not include credential read permission")
	}
	if resp.Matrix["credential"]["read"] {
		t.Fatalf("user matrix should deny credential read")
	}
	for _, forbidden := range []string{"password", "token", "admin_token"} {
		if strings.Contains(strings.ToLower(w.Body.String()), forbidden) {
			t.Fatalf("me response leaked sensitive field %q: %s", forbidden, w.Body.String())
		}
	}
}

func TestMonitorPermissionCheckResults(t *testing.T) {
	r := monitorPermissionTestRouter()
	seedPermissionToken("user-token", "u1", "alice", "user")
	seedPermissionToken("admin-token", "a1", "root", "admin")
	req := map[string]string{"resource": "monitor.alert_rule", "action": "tryrun", "resource_id": "r1", "trace_id": "t1"}
	userResp := performPermissionRequest(r, http.MethodPost, "/monitor/permissions/check", "user-token", req)
	assertPermissionCheck(t, userResp, http.StatusOK, true, false, model.MonitorPermissionReasonRoleDenied)
	adminResp := performPermissionRequest(r, http.MethodPost, "/monitor/permissions/check", "admin-token", req)
	assertPermissionCheck(t, adminResp, http.StatusOK, true, true, model.MonitorPermissionReasonRoleAllowed)
	unknown := map[string]string{"resource": "monitor.alert_rule", "action": "drop"}
	unknownResp := performPermissionRequest(r, http.MethodPost, "/monitor/permissions/check", "user-token", unknown)
	assertPermissionCheck(t, unknownResp, http.StatusBadRequest, false, false, model.MonitorPermissionReasonUnknown)
}

func TestRequireMonitorPermissionMiddleware(t *testing.T) {
	r := monitorPermissionTestRouter()
	seedPermissionToken("user-token", "u1", "alice", "user")
	seedPermissionToken("admin-token", "a1", "root", "admin")
	userResp := performPermissionRequest(r, http.MethodPost, "/monitor/alert-rules/r1/tryrun", "user-token", map[string]string{})
	if userResp.Code != http.StatusForbidden {
		t.Fatalf("user write/action should be 403, got %d body=%s", userResp.Code, userResp.Body.String())
	}
	adminResp := performPermissionRequest(r, http.MethodPost, "/monitor/alert-rules/r1/tryrun", "admin-token", map[string]string{})
	if adminResp.Code != http.StatusOK || !strings.Contains(adminResp.Body.String(), `"entered":true`) {
		t.Fatalf("admin should enter handler, got %d body=%s", adminResp.Code, adminResp.Body.String())
	}
}

func TestCredentialListRequiresMonitorReadPermissionAndMasksSecrets(t *testing.T) {
	r := monitorPermissionTestRouter()
	seedPermissionToken("user-token", "u1", "alice", "user")
	seedPermissionToken("admin-token", "a1", "root", "admin")
	credentialID := "unit-monitor-credential-read"
	store.SaveCredential(&model.Credential{
		ID:       credentialID,
		Name:     "unit credential",
		Protocol: "ssh",
		Username: "ops-user",
		Password: "unit-secret-password",
		SSHKey:   "unit-secret-ssh-key",
		Port:     22,
		Remark:   "credential read gate",
	})
	t.Cleanup(func() {
		store.DeleteCredential(credentialID)
	})

	if w := performPermissionRequest(r, http.MethodGet, "/credentials", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("credential list without auth should be 401, got %d body=%s", w.Code, w.Body.String())
	}
	if w := performPermissionRequest(r, http.MethodGet, "/credentials", "user-token", nil); w.Code != http.StatusForbidden {
		t.Fatalf("credential list for normal user should be 403, got %d body=%s", w.Code, w.Body.String())
	}
	adminResp := performPermissionRequest(r, http.MethodGet, "/credentials", "admin-token", nil)
	assertCredentialListMasked(t, adminResp, credentialID)
}

func TestMonitorPermissionAllowsLegacyAdminTokenHeader(t *testing.T) {
	configureLegacyMonitorAdmin(t, "unit-monitor-admin")
	r := monitorPermissionTestRouter()
	resp := performPermissionRequestWithHeaders(r, http.MethodPost, "/monitor/alert-rules/r1/tryrun", "", map[string]string{}, map[string]string{
		"X-Admin-Token": "unit-monitor-admin",
	})
	assertLegacyAdminMonitorResponse(t, resp)
}

func TestMonitorPermissionAllowsLegacyAdminBearer(t *testing.T) {
	configureLegacyMonitorAdmin(t, "unit-monitor-admin")
	r := monitorPermissionTestRouter()
	resp := performPermissionRequest(r, http.MethodPost, "/monitor/alert-rules/r1/tryrun", "unit-monitor-admin", map[string]string{})
	assertLegacyAdminMonitorResponse(t, resp)
}

func TestRequestActorPrefersContextUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "alice")
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Request.Header.Set("X-Admin-Token", "<TOKEN>")
	if got := requestActor(c); got != "alice" {
		t.Fatalf("requestActor should prefer context username, got %q", got)
	}
}

func TestMonitorPermissionLegacyAdminSetsActorContext(t *testing.T) {
	configureLegacyMonitorAdmin(t, "unit-monitor-admin")
	r := monitorPermissionTestRouter()
	resp := performPermissionRequestWithHeaders(r, http.MethodPost, "/monitor/alert-rules/r1/tryrun", "", map[string]string{}, map[string]string{
		"X-Admin-Token": "unit-monitor-admin",
	})
	if !strings.Contains(resp.Body.String(), `"actor":"admin"`) {
		t.Fatalf("legacy admin token should set audit actor context, body=%s", resp.Body.String())
	}
}

func monitorPermissionTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	clearPermissionTokens()
	r := gin.New()
	r.GET("/monitor/permissions/me", RequireAuth(), GetMonitorPermissionMe)
	r.POST("/monitor/permissions/check", RequireAuth(), CheckMonitorPermission)
	r.GET("/credentials", RequireMonitorPermission("credential", "read"), ListCredentials)
	r.POST("/monitor/alert-rules/:id/tryrun", RequireMonitorPermission("monitor.alert_rule", "tryrun"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"actor":    requestActor(c),
			"entered":  true,
			"role":     c.GetString("role"),
			"username": c.GetString("username"),
		})
	})
	return r
}

func performPermissionRequest(r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
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

func performPermissionRequestWithHeaders(r *gin.Engine, method, path, token string, body any, headers map[string]string) *httptest.ResponseRecorder {
	req := permissionRequest(method, path, token, body)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func permissionRequest(method, path, token string, body any) *http.Request {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func seedPermissionToken(token, userID, username, role string) {
	tokenStore.Store(token, tokenEntry{userID: userID, username: username, role: role, expiresAt: time.Now().Add(time.Hour)})
}

func clearPermissionTokens() {
	tokenStore.Range(func(key, _ any) bool {
		tokenStore.Delete(key)
		return true
	})
}

func permissionListed(perms []model.MonitorPermission, resource, action string) bool {
	for _, perm := range perms {
		if perm.Resource == resource && perm.Action == action {
			return true
		}
	}
	return false
}

func assertPermissionCheck(t *testing.T, w *httptest.ResponseRecorder, code int, known, allowed bool, reason string) {
	t.Helper()
	if w.Code != code {
		t.Fatalf("permission check status want %d got %d body=%s", code, w.Code, w.Body.String())
	}
	var resp model.MonitorPermissionCheckResult
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode check response: %v", err)
	}
	if resp.Known != known || resp.Allowed != allowed || resp.Reason != reason {
		t.Fatalf("unexpected check response: %+v", resp)
	}
}

func assertCredentialListMasked(t *testing.T, w *httptest.ResponseRecorder, credentialID string) {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("credential list should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, credentialID) {
		t.Fatalf("credential list should include seeded credential reference, body=%s", body)
	}
	for _, forbidden := range []string{"unit-secret-password", "unit-secret-ssh-key", "password_hash", "cookie", "dsn"} {
		if strings.Contains(strings.ToLower(body), forbidden) {
			t.Fatalf("credential list leaked sensitive marker %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, `"password":"******"`) || !strings.Contains(body, `"ssh_key":"******"`) {
		t.Fatalf("credential list should mask password and ssh_key, body=%s", body)
	}
}

func assertLegacyAdminMonitorResponse(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("legacy admin token should enter handler, got %d body=%s", w.Code, w.Body.String())
	}
	if w.Header().Get("X-Security-Mode") != "admin-token-enforced" {
		t.Fatalf("legacy admin token should keep security mode, got %q", w.Header().Get("X-Security-Mode"))
	}
	body := strings.ToLower(w.Body.String())
	for _, forbidden := range []string{"unit-monitor-admin", "token", "admin_token", "password", "password_hash"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("legacy admin response leaked sensitive value or field %q: %s", forbidden, w.Body.String())
		}
	}
	for _, expected := range []string{`"actor":"admin"`, `"role":"admin"`, `"username":"admin"`} {
		if !strings.Contains(w.Body.String(), expected) {
			t.Fatalf("legacy admin context missing %s: %s", expected, w.Body.String())
		}
	}
}

func configureLegacyMonitorAdmin(t *testing.T, token string) {
	t.Helper()
	t.Setenv("AIW_ADMIN_TOKEN", token)
	viper.Set("security.admin_token", "")
	viper.Set("security.allow_permissive_admin", false)
	t.Cleanup(func() {
		viper.Set("security.admin_token", "")
		viper.Set("security.allow_permissive_admin", false)
	})
}
