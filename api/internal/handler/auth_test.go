package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthFallbackInitDefaultAdminLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("mysql.dsn", "")
	viper.Set("storage.fallback_file", filepath.Join(t.TempDir(), "memory-store.json"))

	store.Init()
	InitAuth()

	router := authFallbackTestRouter()
	w := performAuthRequest(router, http.MethodPost, "/login", "", map[string]string{
		"username": "admin",
		"password": "admin",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("default fallback admin login should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "password_hash") {
		t.Fatalf("login response must not expose password hash: %s", w.Body.String())
	}

	wrong := performAuthRequest(router, http.MethodPost, "/login", "", map[string]string{
		"username": "admin",
		"password": "wrong-password",
	})
	if wrong.Code != http.StatusUnauthorized {
		t.Fatalf("wrong fallback admin password should be 401, got %d", wrong.Code)
	}
}

func TestAuthFallbackChangePasswordUpdatesLoginSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	viper.Set("mysql.dsn", "")
	viper.Set("storage.fallback_file", filepath.Join(t.TempDir(), "memory-store.json"))

	hash, err := bcrypt.GenerateFromPassword([]byte("old-pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash test password failed: %v", err)
	}
	if err := store.CreateUser(&model.User{
		ID:            store.NewID(),
		Username:      "fallback-change-user",
		PasswordHash:  string(hash),
		Role:          "admin",
		MustChangePwd: true,
	}); err != nil {
		t.Fatalf("create fallback test user failed: %v", err)
	}
	InitAuth()
	router := authFallbackTestRouter()

	login := performAuthRequest(router, http.MethodPost, "/login", "", map[string]string{
		"username": "fallback-change-user",
		"password": "old-pass",
	})
	if login.Code != http.StatusOK {
		t.Fatalf("fallback login should be 200, got %d body=%s", login.Code, login.Body.String())
	}
	token := decodeAuthToken(t, login)

	change := performAuthRequest(router, http.MethodPost, "/change-password", token, map[string]string{
		"old_password": "old-pass",
		"new_password": "new-pass",
	})
	if change.Code != http.StatusOK {
		t.Fatalf("fallback change password should be 200, got %d body=%s", change.Code, change.Body.String())
	}

	oldLogin := performAuthRequest(router, http.MethodPost, "/login", "", map[string]string{
		"username": "fallback-change-user",
		"password": "old-pass",
	})
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("old fallback password should be rejected, got %d", oldLogin.Code)
	}

	newLogin := performAuthRequest(router, http.MethodPost, "/login", "", map[string]string{
		"username": "fallback-change-user",
		"password": "new-pass",
	})
	if newLogin.Code != http.StatusOK {
		t.Fatalf("new fallback password should be accepted, got %d body=%s", newLogin.Code, newLogin.Body.String())
	}
}

func authFallbackTestRouter() *gin.Engine {
	router := gin.New()
	router.POST("/login", Login)
	router.POST("/change-password", RequireAuth(), ChangePassword)
	return router
}

func performAuthRequest(router *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func decodeAuthToken(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode login response failed: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected login token in response")
	}
	return resp.Token
}
