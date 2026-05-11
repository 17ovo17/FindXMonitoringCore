package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func ListMonitoringBuiltinComponents(c *gin.Context) {
	items := store.ListMonitoringBuiltinComponents()
	query := strings.ToLower(strings.TrimSpace(c.Query("query")))
	if query != "" {
		filtered := make([]model.MonitoringBuiltinComponent, 0, len(items))
		for _, item := range items {
			if monitoringBuiltinComponentMatches(item, query) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	c.JSON(http.StatusOK, items)
}

func ListMonitoringBuiltinPayloadTypes(c *gin.Context) {
	c.JSON(http.StatusOK, store.ListMonitoringBuiltinPayloadTypes())
}

func ListMonitoringBuiltinPayloads(c *gin.Context) {
	filter := model.MonitoringBuiltinPayloadFilter{
		ComponentID: strings.TrimSpace(c.Query("component_id")),
		Type:        strings.TrimSpace(c.Query("type")),
		Query:       strings.TrimSpace(c.Query("query")),
	}
	c.JSON(http.StatusOK, store.ListMonitoringBuiltinPayloads(filter))
}

func GetMonitoringBuiltinPayload(c *gin.Context) {
	item, ok := store.GetMonitoringBuiltinPayload(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "builtin payload not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func monitoringBuiltinComponentMatches(item model.MonitoringBuiltinComponent, query string) bool {
	fields := []string{item.ID, item.Ident, item.Name, item.Readme}
	fields = append(fields, item.Tags...)
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}
