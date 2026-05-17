package main

import (
	"ai-workbench-api/internal/embedding"
	"ai-workbench-api/internal/handler"
	"ai-workbench-api/internal/knowledge"
	"ai-workbench-api/internal/scheduler"
	"ai-workbench-api/internal/store"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("读取配置失败: %v", err)
	}

	expandConfigEnvVars()

	os.MkdirAll("../logs", 0755)
	logFile, _ := os.OpenFile("../logs/api.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	logrus.SetOutput(io.MultiWriter(os.Stdout, logFile))
	store.Init()
	embedding.Init()
	go knowledge.LoadExistingDocsToSearchEngine()
	handler.InitAuth()
	handler.InitSandbox()
	scheduler.SetDiagnoseFunc(handler.RunDiagnose)
	scheduler.Start()
	scheduler.StartWorkflowScheduler()
	scheduler.StartMonitorAlertScheduler()
	handler.StartMetricsAutoSync()

	r := gin.Default()
	r.Use(cors.New(corsConfig()))
	registerRoutes(r)

	host := viper.GetString("server.host")
	if host == "" {
		host = "0.0.0.0"
	}
	port := viper.GetString("server.port")
	addr := host + ":" + port
	logrus.Infof("服务启动，监听: %s", addr)
	r.Run(addr)
}

func corsConfig() cors.Config {
	origins := viper.GetStringSlice("server.allowed_origins")
	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Admin-Token", "X-Agent-Token", "X-Test-Batch-Id"},
		ExposeHeaders:    []string{"X-Security-Mode"},
		AllowCredentials: false,
	}
	if len(origins) == 1 && origins[0] == "*" {
		cfg.AllowAllOrigins = true
	} else if len(origins) == 0 {
		cfg.AllowOrigins = defaultLocalhostOrigins()
	} else {
		cfg.AllowOrigins = origins
	}
	return cfg
}

func defaultLocalhostOrigins() []string {
	return []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://localhost:8080",
		"http://127.0.0.1:8080",
	}
}

func expandConfigEnvVars() {
	settings := expandConfigValue(viper.AllSettings())
	if expanded, ok := settings.(map[string]any); ok {
		for key, value := range expanded {
			viper.Set(key, value)
		}
	}
}

func expandConfigValue(value any) any {
	switch typed := value.(type) {
	case string:
		return os.Expand(typed, func(key string) string {
			if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
				return value
			}
			return "${" + key + "}"
		})
	case []any:
		expanded := make([]any, len(typed))
		for i, item := range typed {
			expanded[i] = expandConfigValue(item)
		}
		return expanded
	case map[string]any:
		expanded := make(map[string]any, len(typed))
		for key, item := range typed {
			expanded[key] = expandConfigValue(item)
		}
		return expanded
	default:
		return value
	}
}

func requireAdminToken() gin.HandlerFunc {
	return requireAdminTokenWithAuth(handler.RequireAuth())
}

func requireAdminTokenWithAuth(authRequired gin.HandlerFunc) gin.HandlerFunc {
	configured, allowPermissive := adminTokenConfig()
	logrus.Infof("admin token configured: len=%d empty=%v", len(configured), configured == "")
	return func(c *gin.Context) {
		if configured == "" {
			if allowPermissive {
				c.Header("X-Security-Mode", "permissive-admin")
				c.Next()
				return
			}
			if bearerToken(c) != "" {
				c.Header("X-Security-Mode", "login-token-enforced")
				authRequired(c)
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		provided := strings.TrimSpace(c.GetHeader("X-Admin-Token"))
		if provided != "" && provided == configured {
			c.Header("X-Security-Mode", "admin-token-enforced")
			c.Next()
			return
		}
		bearer := bearerToken(c)
		if bearer == configured {
			c.Header("X-Security-Mode", "admin-token-enforced")
			c.Next()
			return
		}
		if bearer != "" {
			c.Header("X-Security-Mode", "login-token-enforced")
			authRequired(c)
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "admin token required"})
	}
}

func adminTokenConfig() (string, bool) {
	configured := strings.TrimSpace(os.Getenv("AIW_ADMIN_TOKEN"))
	if configured == "" {
		configured = strings.TrimSpace(viper.GetString("security.admin_token"))
	}
	return configured, viper.GetBool("security.allow_permissive_admin")
}

func bearerToken(c *gin.Context) string {
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if !strings.HasPrefix(authorization, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
}
