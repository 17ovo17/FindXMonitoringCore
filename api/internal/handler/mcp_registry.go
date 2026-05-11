package handler

import (
	"net/http"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

// ListMcpServers returns all registered MCP servers.
func ListMcpServers(c *gin.Context) {
	rows := store.ListMcpServers()
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GetMcpServer returns a single MCP server by ID.
func GetMcpServer(c *gin.Context) {
	id := c.Param("id")
	server, ok := store.GetMcpServer(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "mcp server not found"})
		return
	}
	c.JSON(http.StatusOK, server)
}

// CreateMcpServer registers a new MCP server.
func CreateMcpServer(c *gin.Context) {
	var req model.McpServer
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if err := store.CreateMcpServer(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

// UpdateMcpServer updates an existing MCP server.
func UpdateMcpServer(c *gin.Context) {
	id := c.Param("id")
	existing, ok := store.GetMcpServer(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "mcp server not found"})
		return
	}
	var req model.McpServer
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = existing.ID
	req.CreatedAt = existing.CreatedAt
	if err := store.UpdateMcpServer(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

// DeleteMcpServer removes an MCP server by ID.
func DeleteMcpServer(c *gin.Context) {
	id := c.Param("id")
	if _, ok := store.GetMcpServer(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "mcp server not found"})
		return
	}
	if err := store.DeleteMcpServer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// McpServerHealthCheck performs a health check against the server's endpoint.
func McpServerHealthCheck(c *gin.Context) {
	id := c.Param("id")
	server, ok := store.GetMcpServer(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "mcp server not found"})
		return
	}

	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(server.Endpoint + "/health")
	latency := time.Since(start).Milliseconds()

	if err != nil {
		server.Status = "error"
		server.LastCheckAt = time.Now()
		_ = store.UpdateMcpServer(server)
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"latency": latency,
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		server.Status = "online"
	} else {
		server.Status = "offline"
	}
	server.LastCheckAt = time.Now()
	_ = store.UpdateMcpServer(server)

	c.JSON(http.StatusOK, gin.H{
		"status":      server.Status,
		"latency":     latency,
		"status_code": resp.StatusCode,
	})
}
