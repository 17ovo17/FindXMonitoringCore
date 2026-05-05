package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type tokenEntry struct {
	userID    string
	username  string
	role      string
	expiresAt time.Time
}

var tokenStore sync.Map

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		entry, ok := requireTokenEntry(c)
		if !ok {
			return
		}
		setTokenContext(c, entry)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		entry, ok := requireTokenEntry(c)
		if !ok {
			return
		}
		if entry.role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		setTokenContext(c, entry)
		c.Next()
	}
}

func requireTokenEntry(c *gin.Context) (tokenEntry, bool) {
	token := extractToken(c)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return tokenEntry{}, false
	}
	val, ok := tokenStore.Load(token)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录已过期"})
		return tokenEntry{}, false
	}
	entry := val.(tokenEntry)
	if time.Now().After(entry.expiresAt) {
		tokenStore.Delete(token)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录已过期"})
		return tokenEntry{}, false
	}
	return entry, true
}

func setTokenContext(c *gin.Context, entry tokenEntry) {
	c.Set("userID", entry.userID)
	c.Set("username", entry.username)
	c.Set("role", entry.role)
}
