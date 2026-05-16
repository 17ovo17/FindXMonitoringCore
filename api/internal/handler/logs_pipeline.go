package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// In-memory pipeline store for deploy/rollback lifecycle management.
var (
	pipelinesMu       sync.RWMutex
	deployPipelines   = map[string]*DeployLogPipeline{}
	pipelineVersions  = map[string][]*DeployLogPipeline{} // id -> version history
)

// DeployLogPipeline represents a log pipeline with deploy/rollback lifecycle.
type DeployLogPipeline struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Stages      []DeployPipelineStage  `json:"stages"`
	Status      string                 `json:"status"` // draft, deployed, rolled_back
	Version     int                    `json:"version"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// DeployPipelineStage defines a processing stage in the pipeline.
type DeployPipelineStage struct {
	Type   string         `json:"type"`   // regex, json, label, filter, drop
	Config map[string]any `json:"config"`
}

type deployPipelineInput struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Stages      []DeployPipelineStage `json:"stages"`
}

// ListDeployLogPipelines returns all log pipelines.
// GET /api/v1/logs/pipelines
func ListDeployLogPipelines(c *gin.Context) {
	pipelinesMu.RLock()
	items := make([]*DeployLogPipeline, 0, len(deployPipelines))
	for _, p := range deployPipelines {
		items = append(items, p)
	}
	pipelinesMu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"items":  items,
		"total":  len(items),
	})
}

// CreateDeployLogPipeline creates a new log pipeline in draft status.
// POST /api/v1/logs/pipelines/deploy
func CreateDeployLogPipeline(c *gin.Context) {
	var input deployPipelineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pipeline payload"})
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	now := time.Now()
	pipeline := &DeployLogPipeline{
		ID:          newLogPipelineID(),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Stages:      input.Stages,
		Status:      "draft",
		Version:     1,
		CreatedBy:   requestActor(c),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if pipeline.Stages == nil {
		pipeline.Stages = []DeployPipelineStage{}
	}

	pipelinesMu.Lock()
	deployPipelines[pipeline.ID] = pipeline
	pipelineVersions[pipeline.ID] = []*DeployLogPipeline{copyDeployPipeline(pipeline)}
	pipelinesMu.Unlock()

	c.JSON(http.StatusOK, pipeline)
}

// UpdateDeployLogPipeline updates an existing pipeline.
// PUT /api/v1/logs/pipelines/:id/config
func UpdateDeployLogPipeline(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline id is required"})
		return
	}

	var input deployPipelineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pipeline payload"})
		return
	}

	pipelinesMu.Lock()
	pipeline, ok := deployPipelines[id]
	if !ok {
		pipelinesMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
		return
	}

	if strings.TrimSpace(input.Name) != "" {
		pipeline.Name = strings.TrimSpace(input.Name)
	}
	if input.Description != "" {
		pipeline.Description = strings.TrimSpace(input.Description)
	}
	if input.Stages != nil {
		pipeline.Stages = input.Stages
	}
	pipeline.UpdatedAt = time.Now()
	pipelinesMu.Unlock()

	c.JSON(http.StatusOK, pipeline)
}

// DeleteDeployLogPipeline deletes a pipeline.
// DELETE /api/v1/logs/pipelines/:id/config
func DeleteDeployLogPipeline(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline id is required"})
		return
	}

	pipelinesMu.Lock()
	_, ok := deployPipelines[id]
	if !ok {
		pipelinesMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
		return
	}
	delete(deployPipelines, id)
	delete(pipelineVersions, id)
	pipelinesMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DeployLogPipelineAction marks a pipeline as deployed.
// POST /api/v1/logs/pipelines/:id/deploy
func DeployLogPipelineAction(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline id is required"})
		return
	}

	pipelinesMu.Lock()
	pipeline, ok := deployPipelines[id]
	if !ok {
		pipelinesMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
		return
	}

	// Save current version to history before deploying
	pipelineVersions[id] = append(pipelineVersions[id], copyDeployPipeline(pipeline))
	pipeline.Status = "deployed"
	pipeline.Version++
	pipeline.UpdatedAt = time.Now()
	pipelinesMu.Unlock()

	c.JSON(http.StatusOK, pipeline)
}

// RollbackLogPipelineAction rolls back a pipeline to the previous version.
// POST /api/v1/logs/pipelines/:id/rollback
func RollbackLogPipelineAction(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline id is required"})
		return
	}

	pipelinesMu.Lock()
	pipeline, ok := deployPipelines[id]
	if !ok {
		pipelinesMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
		return
	}

	history := pipelineVersions[id]
	if len(history) < 2 {
		pipelinesMu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": "no previous version to rollback to"})
		return
	}

	// Restore previous version
	prev := history[len(history)-2]
	pipeline.Name = prev.Name
	pipeline.Description = prev.Description
	pipeline.Stages = prev.Stages
	pipeline.Status = "rolled_back"
	pipeline.Version++
	pipeline.UpdatedAt = time.Now()

	// Remove the last history entry
	pipelineVersions[id] = history[:len(history)-1]
	pipelinesMu.Unlock()

	c.JSON(http.StatusOK, pipeline)
}

func copyDeployPipeline(p *DeployLogPipeline) *DeployLogPipeline {
	cp := *p
	if p.Stages != nil {
		cp.Stages = make([]DeployPipelineStage, len(p.Stages))
		for i, s := range p.Stages {
			cp.Stages[i] = DeployPipelineStage{Type: s.Type}
			if s.Config != nil {
				raw, _ := json.Marshal(s.Config)
				var cfg map[string]any
				json.Unmarshal(raw, &cfg)
				cp.Stages[i].Config = cfg
			}
		}
	}
	return &cp
}

func newLogPipelineID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "lp_" + hex.EncodeToString(b)
}
