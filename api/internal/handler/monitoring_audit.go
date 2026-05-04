package handler

import (
	"net/http"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func ListMonitorAuditLogs(c *gin.Context) {
	page, limit := monitorAuditPagination(c)
	result, err := store.ListMonitorAuditLogs(model.MonitorAuditLogQuery{
		Page:         page,
		Limit:        limit,
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		ResourceID:   c.Query("resource_id"),
		Status:       c.Query("status"),
		TraceID:      c.Query("trace_id"),
		Scope:        c.Query("scope"),
	})
	if err != nil {
		writeMonitorError(c, http.StatusServiceUnavailable, "audit logs unavailable")
		return
	}
	c.JSON(http.StatusOK, result)
}

func GetMonitorAuditLog(c *gin.Context) {
	item, ok := store.GetMonitorAuditLog(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "audit log not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func monitorAuditPagination(c *gin.Context) (int, int) {
	page := boundedPositiveInt(c.DefaultQuery("page", "1"), 1, 100000)
	limit := boundedPositiveInt(c.DefaultQuery("limit", "20"), 20, 100)
	return page, limit
}
