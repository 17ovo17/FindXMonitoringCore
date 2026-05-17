package aiengine

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Trajectory — 记录 AI 故障自愈的完整执行轨迹
// ---------------------------------------------------------------------------
//
// 用于：审计、复盘、学习
// 每次自愈/诊断过程产生一条 Trajectory，包含多个 Step。

// TrajectoryAction 轨迹步骤动作类型
type TrajectoryAction string

const (
	ActionClassify  TrajectoryAction = "classify"  // 错误分类
	ActionCollect   TrajectoryAction = "collect"   // 证据收集
	ActionDiagnose  TrajectoryAction = "diagnose"  // 诊断分析
	ActionPlan      TrajectoryAction = "plan"      // 制定方案
	ActionExecute   TrajectoryAction = "execute"   // 执行修复
	ActionVerify    TrajectoryAction = "verify"    // 验证结果
	ActionEscalate  TrajectoryAction = "escalate"  // 升级处理
	ActionLearn     TrajectoryAction = "learn"     // 提取经验
)

// TrajectoryStatus 步骤状态
type TrajectoryStatus string

const (
	StepSuccess TrajectoryStatus = "success"
	StepFailed  TrajectoryStatus = "failed"
	StepSkipped TrajectoryStatus = "skipped"
	StepRunning TrajectoryStatus = "running"
)

// TrajectoryOutcome 轨迹最终结果
type TrajectoryOutcome string

const (
	OutcomeResolved  TrajectoryOutcome = "resolved"  // 已解决
	OutcomeEscalated TrajectoryOutcome = "escalated" // 已升级
	OutcomeFailed    TrajectoryOutcome = "failed"    // 失败
	OutcomePartial   TrajectoryOutcome = "partial"   // 部分解决
)

// TrajectoryStep 执行轨迹中的一步
type TrajectoryStep struct {
	StepID    string           `json:"step_id"`
	Action    TrajectoryAction `json:"action"`
	Input     map[string]any   `json:"input,omitempty"`
	Output    map[string]any   `json:"output,omitempty"`
	Duration  int64            `json:"duration_ms"`
	Status    TrajectoryStatus `json:"status"`
	Error     string           `json:"error,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
}

// Trajectory 完整执行轨迹
type Trajectory struct {
	ID         string            `json:"id"`
	AlertID    string            `json:"alert_id,omitempty"`
	SessionID  string            `json:"session_id"`
	Intent     IntentType        `json:"intent"`
	Steps      []TrajectoryStep  `json:"steps"`
	Outcome    TrajectoryOutcome `json:"outcome"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt *time.Time        `json:"finished_at,omitempty"`
}

// TrajectoryRecorder 轨迹记录器
type TrajectoryRecorder struct {
	mu           sync.RWMutex
	trajectories map[string]*Trajectory // id -> Trajectory
	active       map[string]string      // sessionID -> active trajectory ID
}

// NewTrajectoryRecorder 创建轨迹记录器
func NewTrajectoryRecorder() *TrajectoryRecorder {
	return &TrajectoryRecorder{
		trajectories: make(map[string]*Trajectory),
		active:       make(map[string]string),
	}
}

// Start 开始一条新轨迹
func (tr *TrajectoryRecorder) Start(sessionID string, intent IntentType, alertID string) string {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	id := uuid.New().String()
	trajectory := &Trajectory{
		ID:        id,
		AlertID:   alertID,
		SessionID: sessionID,
		Intent:    intent,
		Steps:     make([]TrajectoryStep, 0, 8),
		Outcome:   "",
		Metadata:  make(map[string]any),
		StartedAt: time.Now(),
	}

	tr.trajectories[id] = trajectory
	tr.active[sessionID] = id
	return id
}

// Record 记录一步
func (tr *TrajectoryRecorder) Record(trajectoryID string, action TrajectoryAction, input map[string]any) string {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	trajectory, ok := tr.trajectories[trajectoryID]
	if !ok {
		return ""
	}

	stepID := uuid.New().String()
	step := TrajectoryStep{
		StepID:    stepID,
		Action:    action,
		Input:     input,
		Status:    StepRunning,
		Timestamp: time.Now(),
	}
	trajectory.Steps = append(trajectory.Steps, step)
	return stepID
}

// CompleteStep 完成一步
func (tr *TrajectoryRecorder) CompleteStep(trajectoryID, stepID string, output map[string]any, status TrajectoryStatus, err string) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	trajectory, ok := tr.trajectories[trajectoryID]
	if !ok {
		return
	}

	for i := range trajectory.Steps {
		if trajectory.Steps[i].StepID == stepID {
			trajectory.Steps[i].Output = output
			trajectory.Steps[i].Status = status
			trajectory.Steps[i].Error = err
			trajectory.Steps[i].Duration = time.Since(trajectory.Steps[i].Timestamp).Milliseconds()
			break
		}
	}
}

// Complete 标记轨迹完成
func (tr *TrajectoryRecorder) Complete(trajectoryID string, outcome TrajectoryOutcome) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	trajectory, ok := tr.trajectories[trajectoryID]
	if !ok {
		return
	}

	now := time.Now()
	trajectory.Outcome = outcome
	trajectory.FinishedAt = &now

	// 清理 active 映射
	if tr.active[trajectory.SessionID] == trajectoryID {
		delete(tr.active, trajectory.SessionID)
	}
}

// GetActive 获取会话当前活跃的轨迹
func (tr *TrajectoryRecorder) GetActive(sessionID string) *Trajectory {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	id, ok := tr.active[sessionID]
	if !ok {
		return nil
	}
	return tr.trajectories[id]
}

// Get 获取指定轨迹
func (tr *TrajectoryRecorder) Get(trajectoryID string) *Trajectory {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return tr.trajectories[trajectoryID]
}

// Export 导出轨迹为 JSON 报告
func (tr *TrajectoryRecorder) Export(trajectoryID string) (string, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	trajectory, ok := tr.trajectories[trajectoryID]
	if !ok {
		return "", nil
	}

	data, err := json.MarshalIndent(trajectory, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ListBySession 列出会话的所有轨迹
func (tr *TrajectoryRecorder) ListBySession(sessionID string) []*Trajectory {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	result := make([]*Trajectory, 0)
	for _, t := range tr.trajectories {
		if t.SessionID == sessionID {
			result = append(result, t)
		}
	}
	return result
}

// ExtractLearnings 从已完成的轨迹中提取经验（供 MemoryManager 学习）
func (tr *TrajectoryRecorder) ExtractLearnings(trajectoryID string) []TrajectoryLearning {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	trajectory, ok := tr.trajectories[trajectoryID]
	if !ok || trajectory.Outcome == "" {
		return nil
	}

	learnings := make([]TrajectoryLearning, 0)

	// 从成功的诊断中提取
	if trajectory.Outcome == OutcomeResolved {
		for _, step := range trajectory.Steps {
			if step.Action == ActionDiagnose && step.Status == StepSuccess {
				if rootCause, ok := step.Output["root_cause"]; ok {
					learnings = append(learnings, TrajectoryLearning{
						Type:    MemoryDiagnosis,
						Content: toString(rootCause),
						Tags:    extractTags(trajectory),
					})
				}
			}
			if step.Action == ActionExecute && step.Status == StepSuccess {
				if plan, ok := step.Input["plan"]; ok {
					learnings = append(learnings, TrajectoryLearning{
						Type:    MemoryRunbook,
						Content: toString(plan),
						Tags:    extractTags(trajectory),
					})
				}
			}
		}
	}

	return learnings
}

// TrajectoryLearning 从轨迹中提取的经验
type TrajectoryLearning struct {
	Type    MemoryType `json:"type"`
	Content string     `json:"content"`
	Tags    []string   `json:"tags"`
}

// ---------------------------------------------------------------------------
// 辅助函数
// ---------------------------------------------------------------------------

func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	data, _ := json.Marshal(v)
	return string(data)
}

func extractTags(t *Trajectory) []string {
	tags := make([]string, 0, 3)
	if t.AlertID != "" {
		tags = append(tags, "alert:"+t.AlertID)
	}
	tags = append(tags, "intent:"+string(t.Intent))
	if t.Metadata != nil {
		if host, ok := t.Metadata["host"]; ok {
			tags = append(tags, "host:"+toString(host))
		}
		if service, ok := t.Metadata["service"]; ok {
			tags = append(tags, "service:"+toString(service))
		}
	}
	return tags
}
