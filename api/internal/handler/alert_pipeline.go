package handler

import (
	"net/http"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ListAlertPipelines 返回所有告警事件流水线。
func ListAlertPipelines(c *gin.Context) {
	items := store.ListAlertPipelines()
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateAlertPipeline 创建告警事件流水线。
func CreateAlertPipeline(c *gin.Context) {
	var req model.AlertPipeline
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pipeline payload"})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline name required"})
		return
	}
	for i, proc := range req.Processors {
		if !isValidPipelineProcessorType(proc.Type) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid processor type at index", "index": i, "type": proc.Type})
			return
		}
	}
	created, err := store.CreateAlertPipeline(req)
	if err != nil {
		logrus.WithError(err).Error("create alert pipeline failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create pipeline failed"})
		return
	}
	c.JSON(http.StatusOK, created)
}

// UpdateAlertPipeline 更新指定 ID 的告警事件流水线。
func UpdateAlertPipeline(c *gin.Context) {
	id := c.Param("id")
	var req model.AlertPipeline
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pipeline payload"})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline name required"})
		return
	}
	for i, proc := range req.Processors {
		if !isValidPipelineProcessorType(proc.Type) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid processor type at index", "index": i, "type": proc.Type})
			return
		}
	}
	updated, err := store.UpdateAlertPipeline(id, req)
	if err != nil {
		if err == store.ErrAlertPipelineNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
			return
		}
		logrus.WithError(err).Error("update alert pipeline failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update pipeline failed"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// DeleteAlertPipeline 删除指定 ID 的告警事件流水线。
func DeleteAlertPipeline(c *gin.Context) {
	id := c.Param("id")
	if err := store.DeleteAlertPipeline(id); err != nil {
		if err == store.ErrAlertPipelineNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
			return
		}
		logrus.WithError(err).Error("delete alert pipeline failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete pipeline failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func isValidPipelineProcessorType(t string) bool {
	switch t {
	case "relabel", "drop", "callback", "enrich":
		return true
	default:
		return false
	}
}
