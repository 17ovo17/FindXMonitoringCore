package main

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

func TestContractMatrixProductionRoutesAreRegisteredAndProtected(t *testing.T) {
	store.ResetContractMatrixForTest()
	t.Cleanup(store.ResetContractMatrixForTest)
	r := contractMatrixProductionRouter(t)

	unauthRead := performProductionContractMatrixRequest(t, r, http.MethodGet, "/api/v1/contracts/matrix", "", nil)
	if unauthRead.Code != http.StatusUnauthorized {
		t.Fatalf("contract matrix read must require auth, got %d body=%s", unauthRead.Code, unauthRead.Body.String())
	}

	unauthWrite := performProductionContractMatrixRequest(t, r, http.MethodPost, "/api/v1/contracts/matrix", "", map[string]any{
		"id":         "FX-CONTRACT-PROD-UNAUTH",
		"domain":     "agent",
		"capability": "install",
		"status":     model.ContractStatusMissingExecutor,
	})
	if unauthWrite.Code != http.StatusUnauthorized {
		t.Fatalf("contract matrix write must not be anonymous, got %d body=%s", unauthWrite.Code, unauthWrite.Body.String())
	}

	userToken := loginProductionContractMatrixUserToken(t, r)
	userWrite := performProductionContractMatrixBearerRequest(t, r, http.MethodPost, "/api/v1/contracts/matrix", userToken, map[string]any{
		"id":         "FX-CONTRACT-PROD-USER-DENIED",
		"domain":     "agent",
		"capability": "install",
		"status":     model.ContractStatusMissingExecutor,
	})
	if userWrite.Code != http.StatusForbidden {
		t.Fatalf("contract matrix write must reject read-only user, got %d body=%s", userWrite.Code, userWrite.Body.String())
	}

	adminWrite := performProductionContractMatrixRequest(t, r, http.MethodPost, "/api/v1/contracts/matrix", "unit-contract-admin", map[string]any{
		"id":         "FX-CONTRACT-PROD-ROUTE",
		"domain":     "agent",
		"capability": "install",
		"status":     model.ContractStatusMissingExecutor,
	})
	if adminWrite.Code != http.StatusConflict {
		t.Fatalf("contract matrix production write should reach handler blocked response, got %d body=%s", adminWrite.Code, adminWrite.Body.String())
	}
	assertProductionContractMatrixBlockedShape(t, adminWrite.Body.String(), "FX-CONTRACT-PROD-ROUTE")

	adminRead := performProductionContractMatrixRequest(t, r, http.MethodGet, "/api/v1/contracts/matrix/FX-CONTRACT-PROD-ROUTE", "unit-contract-admin", nil)
	if adminRead.Code != http.StatusConflict {
		t.Fatalf("contract matrix production detail should reach handler blocked response, got %d body=%s", adminRead.Code, adminRead.Body.String())
	}
	assertProductionContractMatrixBlockedShape(t, adminRead.Body.String(), "FX-CONTRACT-PROD-ROUTE")
}

func contractMatrixProductionRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("AIW_ADMIN_TOKEN", "unit-contract-admin")
	oldFallbackFile := viper.GetString("storage.fallback_file")
	viper.Set("storage.fallback_file", filepath.Join(t.TempDir(), "memory-store.json"))
	viper.Set("security.admin_token", "")
	viper.Set("security.allow_permissive_admin", false)
	t.Cleanup(func() {
		viper.Set("storage.fallback_file", oldFallbackFile)
		viper.Set("security.admin_token", "")
		viper.Set("security.allow_permissive_admin", false)
	})
	r := gin.New()
	registerRoutes(r)
	return r
}

func performProductionContractMatrixRequest(t *testing.T, r *gin.Engine, method, path, adminToken string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if adminToken != "" {
		req.Header.Set("X-Admin-Token", adminToken)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performProductionContractMatrixBearerRequest(t *testing.T, r *gin.Engine, method, path, bearerToken string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func loginProductionContractMatrixUserToken(t *testing.T, r *gin.Engine) string {
	t.Helper()
	username := "contract-matrix-readonly-user"
	password := "contract-matrix-readonly-pass"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash synthetic user password: %v", err)
	}
	if existing := store.GetUserByUsername(username); existing == nil {
		if err := store.CreateUser(&model.User{
			ID:            store.NewID(),
			Username:      username,
			PasswordHash:  string(hash),
			Role:          "user",
			MustChangePwd: false,
		}); err != nil {
			t.Fatalf("create synthetic read-only user: %v", err)
		}
	}
	w := performProductionContractMatrixRequest(t, r, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"username": username,
		"password": password,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("synthetic read-only user login should be 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if role, _ := findStringField(resp, "role"); role != "user" {
		t.Fatalf("synthetic login should return role=user, got %q body=%s", role, w.Body.String())
	}
	token, _ := findStringField(resp, "token")
	if token == "" {
		token, _ = findStringField(resp, "access_token")
	}
	if token == "" {
		t.Fatalf("login response missing token field: %s", w.Body.String())
	}
	return token
}

func findStringField(value any, key string) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for k, v := range typed {
			if k == key {
				s, ok := v.(string)
				return s, ok
			}
			if s, ok := findStringField(v, key); ok {
				return s, true
			}
		}
	case []any:
		for _, item := range typed {
			if s, ok := findStringField(item, key); ok {
				return s, true
			}
		}
	}
	return "", false
}

func assertProductionContractMatrixBlockedShape(t *testing.T, body, gapID string) {
	t.Helper()
	for _, required := range []string{`"code":"BLOCKED_BY_CONTRACT"`, `"message"`, `"contract_gap_id":"` + gapID + `"`, `"status":"missing_executor"`, `"safe_to_retry":false`} {
		if !strings.Contains(body, required) {
			t.Fatalf("production blocked response missing %s: %s", required, body)
		}
	}
	for _, forbidden := range []string{"queued", "running", "succeeded", "installed", "data_arrived"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("production blocked response exposed fake runtime state %q: %s", forbidden, body)
		}
	}
}
