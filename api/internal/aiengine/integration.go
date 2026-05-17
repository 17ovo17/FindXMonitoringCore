package aiengine

// ---------------------------------------------------------------------------
// Integration — 将 aiengine 各模块组装为统一入口
// ---------------------------------------------------------------------------
//
// 提供给 handler 层的统一接口，替代原有的硬编码 context 构建逻辑。

import (
	"sync"
)

// Engine 是 aiengine 包的统一入口，聚合所有子模块
type Engine struct {
	mu          sync.RWMutex
	Context     *ContextEngine
	Classifier  *ErrorClassifier
	Trajectory  *TrajectoryRecorder
	initialized bool
}

// EngineConfig 引擎配置
type EngineConfig struct {
	TokenBudget int // Context Engine 的 token 预算
}

var (
	globalEngine     *Engine
	globalEngineOnce sync.Once
)

// GetEngine 获取全局引擎单例
func GetEngine() *Engine {
	globalEngineOnce.Do(func() {
		globalEngine = NewEngine(EngineConfig{})
	})
	return globalEngine
}

// NewEngine 创建引擎实例
func NewEngine(cfg EngineConfig) *Engine {
	ce := NewContextEngine(ContextEngineConfig{
		TokenBudget: cfg.TokenBudget,
	})

	return &Engine{
		Context:     ce,
		Classifier:  NewErrorClassifier(),
		Trajectory:  NewTrajectoryRecorder(),
		initialized: true,
	}
}

// ProcessQuery 处理用户查询的完整流程
// 1. 意图分类
// 2. 构建上下文
// 3. 记录轨迹
// 返回：上下文块列表、意图类型、轨迹 ID
func (e *Engine) ProcessQuery(sessionID string, query string) ([]ContextBlock, IntentType, string, error) {
	// 1. 意图分类
	intent := e.Context.ClassifyIntent(query)

	// 2. 构建上下文
	blocks, err := e.Context.BuildContext(sessionID, query, intent)
	if err != nil {
		return nil, intent, "", err
	}

	// 3. 开始轨迹记录（仅诊断和自愈类）
	var trajectoryID string
	if intent == IntentDiagnose || intent == IntentSelfHeal {
		trajectoryID = e.Trajectory.Start(sessionID, intent, "")
		e.Trajectory.Record(trajectoryID, ActionClassify, map[string]any{
			"query":  query,
			"intent": string(intent),
		})
	}

	return blocks, intent, trajectoryID, nil
}

// ProcessAlert 处理告警事件的完整流程
// 1. 错误分类
// 2. 路由到工作流
// 3. 开始轨迹记录
func (e *Engine) ProcessAlert(sessionID string, event *AlertEvent) (ErrorCategory, string, string) {
	// 1. 错误分类
	category := e.Classifier.Classify(event)

	// 2. 路由到工作流
	workflowID := e.Classifier.RouteToWorkflow(category)

	// 3. 开始轨迹记录
	trajectoryID := e.Trajectory.Start(sessionID, IntentSelfHeal, event.ID)
	stepID := e.Trajectory.Record(trajectoryID, ActionClassify, map[string]any{
		"alert_id": event.ID,
		"title":    event.Title,
		"category": category,
	})
	e.Trajectory.CompleteStep(trajectoryID, stepID, map[string]any{
		"layer":    string(category.Layer),
		"type":     category.Type,
		"workflow": workflowID,
	}, StepSuccess, "")

	return category, workflowID, trajectoryID
}

// LearnFromTrajectory 从已完成的轨迹中提取经验并存储
func (e *Engine) LearnFromTrajectory(trajectoryID string) {
	learnings := e.Trajectory.ExtractLearnings(trajectoryID)
	mm := e.Context.GetMemoryManager()
	for _, l := range learnings {
		mm.Learn(l.Type, l.Content, l.Tags)
	}
}

// RecordDiagnosisMemory 记录诊断结论到记忆
func (e *Engine) RecordDiagnosisMemory(content string, tags []string) string {
	mm := e.Context.GetMemoryManager()
	return mm.Learn(MemoryDiagnosis, content, tags)
}

// RecordRunbookMemory 记录修复方案到记忆
func (e *Engine) RecordRunbookMemory(content string, tags []string) string {
	mm := e.Context.GetMemoryManager()
	return mm.Learn(MemoryRunbook, content, tags)
}

// AddConversationTurn 记录对话轮次
func (e *Engine) AddConversationTurn(sessionID, role, content string) {
	intent := e.Context.ClassifyIntent(content)
	e.Context.curator.AddTurn(sessionID, role, content, intent)
}

// IsInitialized 检查引擎是否已初始化
func (e *Engine) IsInitialized() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.initialized
}
