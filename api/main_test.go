package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestRequireAdminTokenRejectsWhenNotConfiguredByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Unsetenv("AIW_ADMIN_TOKEN")
	viper.Set("security.admin_token", "")
	viper.Set("security.allow_permissive_admin", false)
	r := gin.New()
	r.POST("/protected", requireAdminToken(), func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected default deny, code=%d header=%q", w.Code, w.Header().Get("X-Security-Mode"))
	}
}

func TestRequireAdminTokenPermissiveWhenExplicitlyAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Unsetenv("AIW_ADMIN_TOKEN")
	viper.Set("security.admin_token", "")
	viper.Set("security.allow_permissive_admin", true)
	defer viper.Set("security.allow_permissive_admin", false)
	r := gin.New()
	r.POST("/protected", requireAdminToken(), func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Header().Get("X-Security-Mode") != "permissive-admin" {
		t.Fatalf("expected explicit permissive mode, code=%d header=%q", w.Code, w.Header().Get("X-Security-Mode"))
	}
}

func TestRequireAdminTokenDelegatesLoginWhenAdminTokenEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Unsetenv("AIW_ADMIN_TOKEN")
	viper.Set("security.admin_token", "")
	viper.Set("security.allow_permissive_admin", false)
	authRequired := func(c *gin.Context) {
		if c.GetHeader("Authorization") != "Bearer login-token" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid login token"})
			return
		}
		c.Next()
	}
	r := gin.New()
	r.POST("/protected", requireAdminTokenWithAuth(authRequired), func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set("Authorization", "Bearer login-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Header().Get("X-Security-Mode") != "login-token-enforced" {
		t.Fatalf("expected login auth delegation, code=%d header=%q", w.Code, w.Header().Get("X-Security-Mode"))
	}
}

func TestRequireAdminTokenEnforcesConfiguredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("AIW_ADMIN_TOKEN", "unit-secret")
	defer os.Unsetenv("AIW_ADMIN_TOKEN")
	viper.Set("security.allow_permissive_admin", false)
	r := gin.New()
	r.POST("/protected", requireAdminToken(), func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	missing := httptest.NewRecorder()
	r.ServeHTTP(missing, httptest.NewRequest(http.MethodPost, "/protected", nil))
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("missing token should be unauthorized, got %d", missing.Code)
	}
	allowed := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set("X-Admin-Token", "unit-secret")
	r.ServeHTTP(allowed, req)
	if allowed.Code != http.StatusOK || allowed.Header().Get("X-Security-Mode") != "admin-token-enforced" {
		t.Fatalf("valid token should pass, code=%d header=%q", allowed.Code, allowed.Header().Get("X-Security-Mode"))
	}
}

func TestRequireAdminTokenAcceptsBearerLoginToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("AIW_ADMIN_TOKEN", "unit-secret")
	defer os.Unsetenv("AIW_ADMIN_TOKEN")
	viper.Set("security.allow_permissive_admin", false)
	authRequired := func(c *gin.Context) {
		if c.GetHeader("Authorization") != "Bearer login-token" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid login token"})
			return
		}
		c.Set("username", "admin")
		c.Next()
	}
	r := gin.New()
	r.POST("/protected", requireAdminTokenWithAuth(authRequired), func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	allowed := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set("Authorization", "Bearer login-token")
	r.ServeHTTP(allowed, req)
	if allowed.Code != http.StatusOK || allowed.Header().Get("X-Security-Mode") != "login-token-enforced" {
		t.Fatalf("valid login bearer should pass, code=%d header=%q", allowed.Code, allowed.Header().Get("X-Security-Mode"))
	}

	denied := httptest.NewRecorder()
	badReq := httptest.NewRequest(http.MethodPost, "/protected", nil)
	badReq.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(denied, badReq)
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("invalid login bearer should be unauthorized, got %d", denied.Code)
	}
}

func TestMonitorReadRequiresAuthAndAllowsBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	readRequired := func(c *gin.Context) {
		if c.GetHeader("Authorization") != "Bearer login-token" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("username", "admin")
		c.Next()
	}
	r := gin.New()
	v1 := r.Group("/api/v1")
	v1.GET("/monitor/health", readRequired, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	denied := httptest.NewRecorder()
	r.ServeHTTP(denied, httptest.NewRequest(http.MethodGet, "/api/v1/monitor/health", nil))
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("monitor read without auth should be unauthorized, got %d", denied.Code)
	}

	allowed := httptest.NewRecorder()
	authed := httptest.NewRequest(http.MethodGet, "/api/v1/monitor/health", nil)
	authed.Header.Set("Authorization", "Bearer login-token")
	r.ServeHTTP(allowed, authed)
	if allowed.Code != http.StatusOK {
		t.Fatalf("monitor read with bearer auth should pass, got %d body=%s", allowed.Code, allowed.Body.String())
	}
}

func TestCORSDefaultsAreLocalhostOnly(t *testing.T) {
	viper.Set("server.allowed_origins", []string{})
	cfg := corsConfig()
	if cfg.AllowAllOrigins {
		t.Fatal("default CORS must not allow all origins")
	}
	want := defaultLocalhostOrigins()
	if !reflect.DeepEqual(cfg.AllowOrigins, want) {
		t.Fatalf("default CORS origins=%v, want %v", cfg.AllowOrigins, want)
	}
}

func TestCORSAllowsWildcardOnlyWhenExplicit(t *testing.T) {
	viper.Set("server.allowed_origins", []string{"*"})
	cfg := corsConfig()
	if !cfg.AllowAllOrigins {
		t.Fatal("explicit wildcard CORS should allow all origins")
	}
	if len(cfg.AllowOrigins) != 0 {
		t.Fatalf("wildcard CORS should not also set explicit origins: %v", cfg.AllowOrigins)
	}
}

func TestExpandConfigValueExpandsEnvPlaceholders(t *testing.T) {
	const (
		dsnEnv     = "AI_WORKBENCH_TEST_DSN"
		apiKeyEnv  = "AI_WORKBENCH_TEST_API_KEY"
		emptyEnv   = "AI_WORKBENCH_TEST_EMPTY"
		missingEnv = "AI_WORKBENCH_TEST_MISSING"
	)

	t.Setenv(dsnEnv, "expanded-dsn-from-env")
	t.Setenv(apiKeyEnv, "expanded-api-key-from-env")
	t.Setenv(emptyEnv, "")
	unsetEnvForExpandConfigTest(t, missingEnv)

	input := map[string]any{
		"mysql": map[string]any{
			"dsn": "${AI_WORKBENCH_TEST_DSN}",
			"backup": map[string]any{
				"dsn": "$AI_WORKBENCH_TEST_DSN",
			},
			"literal": "localhost:3306/no-placeholder",
		},
		"providers": []any{
			map[string]any{"apikey": "$AI_WORKBENCH_TEST_API_KEY"},
			map[string]any{"apikey": "${AI_WORKBENCH_TEST_MISSING}"},
			map[string]any{"apikey": "${AI_WORKBENCH_TEST_EMPTY}"},
		},
	}

	want := map[string]any{
		"mysql": map[string]any{
			"dsn": "expanded-dsn-from-env",
			"backup": map[string]any{
				"dsn": "expanded-dsn-from-env",
			},
			"literal": "localhost:3306/no-placeholder",
		},
		"providers": []any{
			map[string]any{"apikey": "expanded-api-key-from-env"},
			map[string]any{"apikey": "${AI_WORKBENCH_TEST_MISSING}"},
			map[string]any{"apikey": "${AI_WORKBENCH_TEST_EMPTY}"},
		},
	}

	got := expandConfigValue(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expanded config mismatch:\ngot:  %#v\nwant: %#v", got, want)
	}
}

func unsetEnvForExpandConfigTest(t *testing.T, key string) {
	t.Helper()

	previous, hadPrevious := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset test env %s: %v", key, err)
	}
	t.Cleanup(func() {
		if hadPrevious {
			if err := os.Setenv(key, previous); err != nil {
				t.Fatalf("restore test env %s: %v", key, err)
			}
			return
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("cleanup test env %s: %v", key, err)
		}
	})
}
