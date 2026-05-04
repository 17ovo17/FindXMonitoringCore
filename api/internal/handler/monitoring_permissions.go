package handler

import (
	"net/http"
	"os"
	"strings"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RequireMonitorPermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authenticateMonitorRequest(c) {
			return
		}
		result := monitorPermissionChecker().Check(c.GetString("role"), resource, action)
		if !result.Known {
			c.AbortWithStatusJSON(http.StatusBadRequest, result)
			return
		}
		if !result.Allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func authenticateMonitorRequest(c *gin.Context) bool {
	if authenticateLegacyMonitorAdmin(c) {
		return true
	}
	token := extractToken(c)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}
	val, ok := tokenStore.Load(token)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}
	entry := val.(tokenEntry)
	if time.Now().After(entry.expiresAt) {
		tokenStore.Delete(token)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}
	c.Set("userID", entry.userID)
	c.Set("username", entry.username)
	c.Set("role", entry.role)
	return true
}

func authenticateLegacyMonitorAdmin(c *gin.Context) bool {
	mode, ok := legacyMonitorAdminMode(c)
	if !ok {
		return false
	}
	c.Header("X-Security-Mode", mode)
	c.Set("userID", "admin")
	c.Set("username", "admin")
	c.Set("role", "admin")
	return true
}

func legacyMonitorAdminMode(c *gin.Context) (string, bool) {
	configured, allowPermissive := monitorAdminTokenConfig()
	if configured == "" {
		if allowPermissive {
			return "permissive-admin", true
		}
		return "", false
	}
	headerToken := strings.TrimSpace(c.GetHeader("X-Admin-Token"))
	if headerToken != "" && headerToken == configured {
		return "admin-token-enforced", true
	}
	if monitorBearerToken(c) == configured {
		return "admin-token-enforced", true
	}
	return "", false
}

func monitorAdminTokenConfig() (string, bool) {
	configured := strings.TrimSpace(os.Getenv("AIW_ADMIN_TOKEN"))
	if configured == "" {
		configured = strings.TrimSpace(viper.GetString("security.admin_token"))
	}
	return configured, viper.GetBool("security.allow_permissive_admin")
}

func monitorBearerToken(c *gin.Context) string {
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if !strings.HasPrefix(authorization, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
}

func GetMonitorPermissionMe(c *gin.Context) {
	user := model.MonitorPermissionUser{
		ID:       c.GetString("userID"),
		Username: c.GetString("username"),
		Role:     c.GetString("role"),
	}
	scope := model.MonitorPermissionScope{}
	checker := monitorPermissionChecker()
	role := user.Role
	c.JSON(http.StatusOK, model.MonitorPermissionMeResponse{
		Version:     model.MonitorPermissionVersion,
		Mode:        model.MonitorPermissionMode,
		User:        user,
		Scope:       scope,
		Permissions: checker.PermissionsForRole(role),
		Matrix:      checker.MatrixForRole(role),
	})
}

func CheckMonitorPermission(c *gin.Context) {
	var req model.MonitorPermissionCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.MonitorPermissionCheckResult{Known: false, Reason: model.MonitorPermissionReasonUnknown})
		return
	}
	result := monitorPermissionChecker().Check(c.GetString("role"), req.Resource, req.Action)
	result.ResourceID = strings.TrimSpace(req.ResourceID)
	result.TraceID = strings.TrimSpace(req.TraceID)
	if !result.Known {
		c.JSON(http.StatusBadRequest, result)
		return
	}
	c.JSON(http.StatusOK, result)
}

func monitorPermissionChecker() model.MonitorPermissionChecker {
	return model.NewMonitorPermissionChecker()
}
