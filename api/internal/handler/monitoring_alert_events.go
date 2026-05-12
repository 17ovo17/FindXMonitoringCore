package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func ListMonitorEventsCurrent(c *gin.Context) {
	listMonitorEventsPaged(c, true)
}

func ListMonitorEventsHistory(c *gin.Context) {
	listMonitorEventsPaged(c, false)
}

func listMonitorEventsPaged(c *gin.Context, current bool) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := strings.TrimSpace(c.Query("status"))
	severity := strings.TrimSpace(c.Query("severity"))
	search := strings.TrimSpace(c.Query("search"))

	events, total, err := store.ListMonitorAlertEventsPaged(current, page, pageSize, status, severity, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query alert events failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": events,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// CreateMonitorEvent 手动创建告警事件（用于测试和外部集成）。
func CreateMonitorEvent(c *gin.Context) {
	var req model.MonitorAlertEvent
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if req.Severity == "" {
		req.Severity = model.MonitorAlertSeverityWarning
	}
	event, err := store.UpsertMonitorAlertEvent(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create alert event failed"})
		return
	}
	c.JSON(http.StatusCreated, event)
}

func GetMonitorEvent(c *gin.Context) {
	event, ok := store.GetMonitorAlertEvent(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func AckMonitorEvent(c *gin.Context) {
	applyMonitorEventAction(c, model.MonitorAlertEventActionAck)
}

func AssignMonitorEvent(c *gin.Context) {
	applyMonitorEventAction(c, model.MonitorAlertEventActionAssign)
}

func ResolveMonitorEvent(c *gin.Context) {
	applyMonitorEventAction(c, model.MonitorAlertEventActionResolve)
}

func ArchiveMonitorEvent(c *gin.Context) {
	applyMonitorEventAction(c, model.MonitorAlertEventActionArchive)
}

func applyMonitorEventAction(c *gin.Context, actionName string) {
	var req struct {
		Actor    string `json:"actor"`
		Reason   string `json:"reason"`
		Assignee string `json:"assignee"`
		TraceID  string `json:"trace_id"`
	}
	_ = c.ShouldBindJSON(&req)
	if actionName == model.MonitorAlertEventActionAssign && strings.TrimSpace(req.Assignee) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assignee is required"})
		return
	}
	current, ok := store.GetMonitorAlertEvent(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert event not found"})
		return
	}
	if _, err := model.ValidateMonitorAlertEventTransition(current.Status, actionName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	action := model.MonitorAlertAction{Action: actionName, Actor: firstNonEmpty(req.Actor, requestActor(c)), Reason: req.Reason, Assignee: req.Assignee, TraceID: req.TraceID}
	event, ok, err := store.ApplyMonitorAlertEventAction(current.ID, action)
	if err != nil {
		if errors.Is(err, store.ErrTerminalMonitorAlertEvent) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alert event action failed"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

