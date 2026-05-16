package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// AIOpsSessionRecord represents a persisted AI session.
type AIOpsSessionRecord struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Title     string                 `json:"title"`
	Mode      string                 `json:"mode"`
	Status    string                 `json:"status"`
	Context   map[string]any         `json:"context,omitempty"`
	Messages  []AIOpsSessionMessage  `json:"messages,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// AIOpsSessionMessage represents a message within a session.
type AIOpsSessionMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	aiSessionsMu sync.RWMutex
	aiSessions   = map[string]*AIOpsSessionRecord{}
)

func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "aisess-" + hex.EncodeToString(b)
}

func generateMessageID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "msg-" + hex.EncodeToString(b)
}

// AISessionList handles GET /api/v1/ai/sessions
func AISessionList(c *gin.Context) {
	aiSessionsMu.RLock()
	defer aiSessionsMu.RUnlock()
	sessions := make([]*AIOpsSessionRecord, 0, len(aiSessions))
	for _, s := range aiSessions {
		sessions = append(sessions, s)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"sessions": sessions, "total": len(sessions)}})
}

// AISessionGet handles GET /api/v1/ai/sessions/:id
func AISessionGet(c *gin.Context) {
	id := c.Param("id")
	aiSessionsMu.RLock()
	session, ok := aiSessions[id]
	aiSessionsMu.RUnlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": "session not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": session})
}

// AISessionCreate handles POST /api/v1/ai/sessions
func AISessionCreate(c *gin.Context) {
	var req struct {
		Title   string         `json:"title"`
		Mode    string         `json:"mode"`
		Context map[string]any `json:"context"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	now := time.Now()
	session := &AIOpsSessionRecord{
		ID:        generateSessionID(),
		UserID:    c.GetString("user_id"),
		Title:     req.Title,
		Mode:      req.Mode,
		Status:    "active",
		Context:   req.Context,
		Messages:  []AIOpsSessionMessage{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if session.Title == "" {
		session.Title = "AI 会话 " + now.Format("01-02 15:04")
	}
	if session.Mode == "" {
		session.Mode = "diagnostic"
	}
	aiSessionsMu.Lock()
	aiSessions[session.ID] = session
	aiSessionsMu.Unlock()
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": session})
}

// AISessionDelete handles DELETE /api/v1/ai/sessions/:id
func AISessionDelete(c *gin.Context) {
	id := c.Param("id")
	aiSessionsMu.Lock()
	_, ok := aiSessions[id]
	if ok {
		delete(aiSessions, id)
	}
	aiSessionsMu.Unlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": "session not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"deleted": id}})
}

// AISessionAddMessage adds a message to a session (internal use).
func AISessionAddMessage(sessionID, role, content string) {
	aiSessionsMu.Lock()
	defer aiSessionsMu.Unlock()
	session, ok := aiSessions[sessionID]
	if !ok {
		return
	}
	session.Messages = append(session.Messages, AIOpsSessionMessage{
		ID:        generateMessageID(),
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	})
	session.UpdatedAt = time.Now()
}
