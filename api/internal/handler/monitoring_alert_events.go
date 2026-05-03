package handler

import (
	"errors"
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func ListMonitorEventsCurrent(c *gin.Context) {
	c.JSON(http.StatusOK, filterMonitorEvents(store.ListMonitorAlertEvents(true), c))
}

func ListMonitorEventsHistory(c *gin.Context) {
	c.JSON(http.StatusOK, filterMonitorEvents(store.ListMonitorAlertEvents(false), c))
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
	applyMonitorEventAction(c, "ack")
}

func AssignMonitorEvent(c *gin.Context) {
	applyMonitorEventAction(c, "assign")
}

func ResolveMonitorEvent(c *gin.Context) {
	applyMonitorEventAction(c, "resolve")
}

func ArchiveMonitorEvent(c *gin.Context) {
	applyMonitorEventAction(c, "archive")
}

func applyMonitorEventAction(c *gin.Context, actionName string) {
	var req struct {
		Actor    string `json:"actor"`
		Reason   string `json:"reason"`
		Assignee string `json:"assignee"`
		TraceID  string `json:"trace_id"`
	}
	_ = c.ShouldBindJSON(&req)
	if actionName == "assign" && strings.TrimSpace(req.Assignee) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assignee is required"})
		return
	}
	action := model.MonitorAlertAction{Action: actionName, Actor: firstNonEmpty(req.Actor, requestActor(c)), Reason: req.Reason, Assignee: req.Assignee, TraceID: req.TraceID}
	event, ok, err := store.ApplyMonitorAlertEventAction(c.Param("id"), action)
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

func filterMonitorEvents(events []model.MonitorAlertEvent, c *gin.Context) []model.MonitorAlertEvent {
	status := strings.TrimSpace(c.Query("status"))
	severity := strings.TrimSpace(c.Query("severity"))
	if status == "" && severity == "" {
		return events
	}
	out := []model.MonitorAlertEvent{}
	for _, event := range events {
		if status != "" && event.Status != status {
			continue
		}
		if severity != "" && event.Severity != severity {
			continue
		}
		out = append(out, event)
	}
	return out
}
