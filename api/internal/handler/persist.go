package handler

import (
	"net/http"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func HealthStorage(c *gin.Context) {
	c.JSON(http.StatusOK, store.Health())
}

func ListChatSessions(c *gin.Context) {
	c.JSON(http.StatusOK, store.ListChatSessions())
}

func CreateChatSession(c *gin.Context) {
	var req struct {
		Title    string `json:"title"`
		Model    string `json:"model"`
		TargetIP string `json:"target_ip"`
	}
	_ = c.ShouldBindJSON(&req)
	now := time.Now()
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "New session"
	}
	s := &model.ChatSession{ID: store.NewID(), Title: title, Model: req.Model, TargetIP: req.TargetIP, CreatedAt: now, UpdatedAt: now}
	store.SaveChatSession(s)
	c.JSON(http.StatusOK, s)
}

func GetChatSession(c *gin.Context) {
	if s, ok := store.GetChatSession(c.Param("id")); ok {
		c.JSON(http.StatusOK, s)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
}

func RenameChatSession(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	s, ok := store.GetChatSession(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	s.Title = strings.TrimSpace(req.Title)
	store.SaveChatSession(s)
	c.JSON(http.StatusOK, s)
}

func DeleteChatSession(c *gin.Context) {
	store.DeleteChatSession(c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetTopology(c *gin.Context) {
	g := store.GetTopology()
	agents := store.ListAgents()
	for i := range g.Nodes {
		if g.Nodes[i].Type == "host" && g.Nodes[i].IP != "" {
			g.Nodes[i].Status = "offline"
			for _, a := range agents {
				if a.IP == g.Nodes[i].IP && a.Online {
					g.Nodes[i].Status = "online"
				}
			}
		}
	}
	c.JSON(http.StatusOK, g)
}

func SaveTopology(c *gin.Context) {
	var g model.TopologyGraph
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	seen := map[string]bool{}
	for _, n := range g.Nodes {
		if strings.TrimSpace(n.ID) == "" || strings.TrimSpace(n.Name) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "node id and name are required"})
			return
		}
		seen[n.ID] = true
	}
	edges := map[string]bool{}
	for _, e := range g.Edges {
		if !seen[e.SourceID] || !seen[e.TargetID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "edge source/target must exist"})
			return
		}
		key := e.SourceID + "->" + e.TargetID + ":" + e.Protocol
		if edges[key] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate edge"})
			return
		}
		edges[key] = true
	}
	store.SaveTopology(g)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func TopologyResources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"agents": store.ListAgents(), "platform": platformResources()})
}
