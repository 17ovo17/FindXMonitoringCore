package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProfilingTask 表示一个性能分析任务
type ProfilingTask struct {
	ID              string              `json:"id"`
	Service         string              `json:"service"`
	ServiceInstance string              `json:"service_instance"`
	Duration        int                 `json:"duration_seconds"`
	StartTime       time.Time           `json:"start_time"`
	EndTime         time.Time           `json:"end_time"`
	Status          string              `json:"status"`
	CreatedAt       time.Time           `json:"created_at"`
	FlameGraph      *FlameGraphNode     `json:"flame_graph,omitempty"`
}

// FlameGraphNode 表示火焰图中的一个节点
type FlameGraphNode struct {
	Name     string           `json:"name"`
	Value    int64            `json:"value"`
	Children []*FlameGraphNode `json:"children,omitempty"`
}

var (
	profilingTasks   = make(map[string]*ProfilingTask)
	profilingTasksMu sync.RWMutex
)

// TracingListProfilingTasks answers GET /api/v1/tracing/profiling/tasks
// Returns list of profiling tasks.
func TracingListProfilingTasks(c *gin.Context) {
	service := strings.TrimSpace(c.Query("service"))

	profilingTasksMu.RLock()
	defer profilingTasksMu.RUnlock()

	tasks := make([]*ProfilingTask, 0)
	for _, task := range profilingTasks {
		if service != "" && !strings.EqualFold(task.Service, service) {
			continue
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks, "total": len(tasks)})
}

// TracingCreateProfilingTask answers POST /api/v1/tracing/profiling/tasks
// Creates a new profiling task for a service instance.
func TracingCreateProfilingTask(c *gin.Context) {
	var req struct {
		Service         string `json:"service" binding:"required"`
		ServiceInstance string `json:"service_instance"`
		Duration        int    `json:"duration_seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service is required"})
		return
	}
	if req.Duration <= 0 {
		req.Duration = 60
	}
	if req.Duration > 300 {
		req.Duration = 300
	}

	taskID := generateTaskID()
	now := time.Now()
	task := &ProfilingTask{
		ID:              taskID,
		Service:         req.Service,
		ServiceInstance: req.ServiceInstance,
		Duration:        req.Duration,
		StartTime:       now,
		EndTime:         now.Add(time.Duration(req.Duration) * time.Second),
		Status:          "running",
		CreatedAt:       now,
	}

	profilingTasksMu.Lock()
	profilingTasks[taskID] = task
	profilingTasksMu.Unlock()

	logrus.WithFields(logrus.Fields{
		"task_id": taskID,
		"service": req.Service,
		"duration": req.Duration,
	}).Info("profiling task created")

	// 异步模拟任务完成并生成火焰图数据
	go completeProfilingTask(taskID, req.Service, req.Duration)

	c.JSON(http.StatusCreated, gin.H{"task": task})
}

// TracingGetProfilingTask answers GET /api/v1/tracing/profiling/tasks/:id
// Returns task result with flame graph data.
func TracingGetProfilingTask(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id is required"})
		return
	}

	profilingTasksMu.RLock()
	task, exists := profilingTasks[taskID]
	profilingTasksMu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "profiling task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// completeProfilingTask 模拟任务完成并生成火焰图数据
func completeProfilingTask(taskID string, service string, durationSec int) {
	// 等待任务持续时间结束
	time.Sleep(time.Duration(durationSec) * time.Second)

	// 基于 buffer 中的实际 span 数据生成火焰图
	flameGraph := buildFlameGraphFromBuffer(service)

	profilingTasksMu.Lock()
	if task, ok := profilingTasks[taskID]; ok {
		task.Status = "completed"
		task.FlameGraph = flameGraph
	}
	profilingTasksMu.Unlock()

	logrus.WithField("task_id", taskID).Info("profiling task completed")
}

// buildFlameGraphFromBuffer 从 buffer 中的 span 数据构建火焰图
func buildFlameGraphFromBuffer(service string) *FlameGraphNode {
	traceBufferMu.RLock()
	defer traceBufferMu.RUnlock()

	// 按 operation 聚合耗时
	opDurations := make(map[string]int64)
	for _, seg := range traceBuffer {
		if !strings.EqualFold(seg.Service, service) {
			continue
		}
		if seg.Operation != "" {
			opDurations[seg.Operation] += seg.Duration
		}
	}

	// 构建火焰图树
	root := &FlameGraphNode{
		Name:     service,
		Value:    0,
		Children: make([]*FlameGraphNode, 0),
	}

	for op, duration := range opDurations {
		root.Value += duration
		child := &FlameGraphNode{
			Name:     op,
			Value:    duration,
			Children: []*FlameGraphNode{},
		}
		root.Children = append(root.Children, child)
	}

	// 如果没有数据，返回基础结构
	if root.Value == 0 {
		root.Value = 1
	}

	return root
}

// generateTaskID 生成唯一任务 ID
func generateTaskID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405")
	}
	return hex.EncodeToString(b)
}
