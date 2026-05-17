package aiengine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// AutonomousInvestigator — 自主调查引擎（对标 Neubird）
// ---------------------------------------------------------------------------
//
// 核心理念：AI 自己决定下一步查什么，不需要人指定调查路径
//
// 工作流程：
// 1. 接收初始信号（告警/异常/用户提问）
// 2. 生成调查假设（Hypothesis）
// 3. 选择验证工具（从工具注册表中选最相关的）
// 4. 执行验证 → 分析结果
// 5. 根据结果决定：
//    a. 假设成立 → 输出结论
//    b. 假设不成立 → 生成新假设，回到步骤 3
//    c. 需要更多数据 → 选择新工具继续调查
// 6. 最多 MaxRounds 轮调查，超过则输出当前最佳结论

// Investigation 一次完整的调查过程
type Investigation struct {
	ID           string              `json:"id"`
	Trigger      InvestigationTrigger `json:"trigger"`
	Hypotheses   []Hypothesis        `json:"hypotheses"`
	Steps        []InvestigationStep `json:"steps"`
	Conclusion   *Conclusion         `json:"conclusion,omitempty"`
	MaxRounds    int                 `json:"max_rounds"`
	CurrentRound int                 `json:"current_round"`
	Status       string              `json:"status"` // investigating, concluded, escalated
	StartedAt    time.Time           `json:"started_at"`
	FinishedAt   *time.Time          `json:"finished_at,omitempty"`
}

// InvestigationTrigger 调查触发源
type InvestigationTrigger struct {
	Type    string         `json:"type"` // alert, anomaly, user_query
	Data    map[string]any `json:"data"`
	Context map[string]any `json:"context"` // CMDB/告警/指标上下文
}

// Hypothesis 调查假设
type Hypothesis struct {
	ID          string  `json:"id"`
	Description string  `json:"description"` // 如 "MySQL 连接池耗尽导致超时"
	Confidence  float64 `json:"confidence"`  // 0-1
	Evidence    []string `json:"evidence"`   // 支持证据
	Refuted     bool    `json:"refuted"`
}

// InvestigationStep 调查步骤
type InvestigationStep struct {
	Round      int            `json:"round"`
	ToolUsed   string         `json:"tool_used"`
	Query      map[string]any `json:"query"`
	Result     any            `json:"result"`
	Reasoning  string         `json:"reasoning"`   // 为什么选这个工具
	NextAction string         `json:"next_action"` // continue, conclude, escalate
	Timestamp  time.Time      `json:"timestamp"`
}

// Conclusion 调查结论
type Conclusion struct {
	RootCause     string   `json:"root_cause"`
	Confidence    float64  `json:"confidence"`
	Evidence      []string `json:"evidence"`
	Suggestions   []string `json:"suggestions"`
	AffectedScope string   `json:"affected_scope"`
}

// AutonomousInvestigator 自主调查引擎
type AutonomousInvestigator struct {
	mu           sync.RWMutex
	toolRegistry *ToolRegistry
	maxRounds    int
	// investigations 存储进行中和已完成的调查
	investigations map[string]*Investigation
}

// NewAutonomousInvestigator 创建自主调查引擎
func NewAutonomousInvestigator(toolRegistry *ToolRegistry) *AutonomousInvestigator {
	return &AutonomousInvestigator{
		toolRegistry:   toolRegistry,
		maxRounds:      5,
		investigations: make(map[string]*Investigation),
	}
}

// Investigate 启动自主调查
func (inv *AutonomousInvestigator) Investigate(ctx context.Context, trigger InvestigationTrigger) (*Investigation, error) {
	investigation := &Investigation{
		ID:        uuid.New().String(),
		Trigger:   trigger,
		MaxRounds: inv.maxRounds,
		Status:    "investigating",
		StartedAt: time.Now(),
	}

	inv.mu.Lock()
	inv.investigations[investigation.ID] = investigation
	inv.mu.Unlock()

	for investigation.CurrentRound < investigation.MaxRounds {
		select {
		case <-ctx.Done():
			investigation.Status = "escalated"
			now := time.Now()
			investigation.FinishedAt = &now
			return investigation, ctx.Err()
		default:
		}

		investigation.CurrentRound++

		// 1. 生成/更新假设（基于当前证据）
		hypotheses := inv.generateHypotheses(investigation)
		investigation.Hypotheses = hypotheses

		// 2. 选择最佳验证工具
		tool, params, reasoning := inv.selectNextTool(investigation)
		if tool == "" {
			// 无更多工具可用，输出结论
			break
		}

		// 3. 执行工具
		result, err := inv.executeTool(ctx, tool, params)

		// 4. 记录步骤
		step := InvestigationStep{
			Round:     investigation.CurrentRound,
			ToolUsed:  tool,
			Query:     params,
			Result:    result,
			Reasoning: reasoning,
			Timestamp: time.Now(),
		}

		if err != nil {
			step.Result = map[string]any{"error": err.Error()}
			step.NextAction = "continue"
			investigation.Steps = append(investigation.Steps, step)
			continue
		}

		// 5. 分析结果，决定下一步
		action := inv.analyzeAndDecide(investigation, step)
		step.NextAction = action
		investigation.Steps = append(investigation.Steps, step)

		if action == "conclude" {
			investigation.Conclusion = inv.formConclusion(investigation)
			investigation.Status = "concluded"
			now := time.Now()
			investigation.FinishedAt = &now
			break
		}
		if action == "escalate" {
			investigation.Status = "escalated"
			now := time.Now()
			investigation.FinishedAt = &now
			break
		}
	}

	if investigation.Status == "investigating" {
		// 达到最大轮次，输出当前最佳结论
		investigation.Conclusion = inv.formConclusion(investigation)
		investigation.Status = "concluded"
		now := time.Now()
		investigation.FinishedAt = &now
	}

	return investigation, nil
}

// GetInvestigation 获取调查详情
func (inv *AutonomousInvestigator) GetInvestigation(id string) (*Investigation, bool) {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	i, ok := inv.investigations[id]
	return i, ok
}

// ListInvestigations 列出所有调查
func (inv *AutonomousInvestigator) ListInvestigations() []*Investigation {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	result := make([]*Investigation, 0, len(inv.investigations))
	for _, i := range inv.investigations {
		result = append(result, i)
	}
	return result
}

// generateHypotheses 基于当前证据生成假设
func (inv *AutonomousInvestigator) generateHypotheses(investigation *Investigation) []Hypothesis {
	hypotheses := make([]Hypothesis, 0)

	// 基于触发类型生成初始假设
	triggerType := investigation.Trigger.Type
	triggerData := investigation.Trigger.Data

	switch triggerType {
	case "alert":
		// 从告警数据推断可能的根因
		if alertName, ok := triggerData["alert_name"].(string); ok {
			hypotheses = append(hypotheses, inv.hypothesesFromAlert(alertName, triggerData)...)
		}
	case "anomaly":
		// 从异常指标推断
		if metricName, ok := triggerData["metric_name"].(string); ok {
			hypotheses = append(hypotheses, Hypothesis{
				ID:          uuid.New().String(),
				Description: fmt.Sprintf("指标 %s 异常可能由资源瓶颈导致", metricName),
				Confidence:  0.5,
			})
		}
	case "user_query":
		// 从用户问题推断
		hypotheses = append(hypotheses, Hypothesis{
			ID:          uuid.New().String(),
			Description: "用户描述的问题需要进一步数据验证",
			Confidence:  0.3,
		})
	}

	// 如果已有步骤结果，根据结果更新假设置信度
	for i := range hypotheses {
		for _, step := range investigation.Steps {
			if step.Result != nil {
				hypotheses[i].Confidence = inv.adjustConfidence(hypotheses[i], step)
			}
		}
	}

	return hypotheses
}

// hypothesesFromAlert 从告警名称生成假设列表
func (inv *AutonomousInvestigator) hypothesesFromAlert(alertName string, data map[string]any) []Hypothesis {
	hypotheses := []Hypothesis{
		{
			ID:          uuid.New().String(),
			Description: fmt.Sprintf("告警 [%s] 可能由资源耗尽导致", alertName),
			Confidence:  0.6,
		},
		{
			ID:          uuid.New().String(),
			Description: fmt.Sprintf("告警 [%s] 可能由上游服务异常引起", alertName),
			Confidence:  0.4,
		},
		{
			ID:          uuid.New().String(),
			Description: fmt.Sprintf("告警 [%s] 可能由配置变更触发", alertName),
			Confidence:  0.3,
		},
	}
	// 如果有 severity 信息，提高高严重度假设的置信度
	if sev, ok := data["severity"].(string); ok && (sev == "critical" || sev == "P0") {
		hypotheses[0].Confidence = 0.8
	}
	return hypotheses
}

// adjustConfidence 根据步骤结果调整假设置信度
func (inv *AutonomousInvestigator) adjustConfidence(h Hypothesis, step InvestigationStep) float64 {
	// 如果工具返回了异常数据，提高相关假设的置信度
	if resultMap, ok := step.Result.(map[string]any); ok {
		if _, hasError := resultMap["error"]; hasError {
			return h.Confidence * 0.8 // 工具执行失败，略微降低置信度
		}
		if anomalous, ok := resultMap["anomalous"].(bool); ok && anomalous {
			return h.Confidence * 1.3 // 发现异常，提高置信度（上限 1.0）
		}
	}
	return h.Confidence
}

// selectNextTool 选择最能验证/推翻当前假设的工具
func (inv *AutonomousInvestigator) selectNextTool(investigation *Investigation) (string, map[string]any, string) {
	// 收集已使用的工具，避免重复
	usedTools := make(map[string]bool)
	for _, step := range investigation.Steps {
		usedTools[step.ToolUsed] = true
	}

	// 根据触发类型和当前轮次选择工具策略
	var toolPriority []string
	switch investigation.Trigger.Type {
	case "alert":
		toolPriority = []string{
			"prometheus_query", "alert_correlate", "logs_query",
			"cmdb_topology", "metrics_snapshot", "knowledge_search",
		}
	case "anomaly":
		toolPriority = []string{
			"prometheus_query_range", "metrics_snapshot", "logs_query",
			"alert_list", "cmdb_query", "logs_pattern_analysis",
		}
	default:
		toolPriority = []string{
			"prometheus_query", "logs_query", "alert_list",
			"cmdb_query", "knowledge_search", "metrics_snapshot",
		}
	}

	// 选择第一个未使用的工具
	for _, toolName := range toolPriority {
		if usedTools[toolName] {
			continue
		}
		params := inv.buildToolParams(toolName, investigation)
		reasoning := fmt.Sprintf("第 %d 轮调查：选择 %s 验证当前假设", investigation.CurrentRound, toolName)
		return toolName, params, reasoning
	}

	return "", nil, ""
}

// buildToolParams 根据工具名和调查上下文构建工具参数
func (inv *AutonomousInvestigator) buildToolParams(toolName string, investigation *Investigation) map[string]any {
	params := make(map[string]any)
	triggerData := investigation.Trigger.Data
	triggerCtx := investigation.Trigger.Context

	switch toolName {
	case "prometheus_query", "prometheus_query_range":
		if metric, ok := triggerData["metric_name"].(string); ok {
			params["query"] = metric
		} else if alertName, ok := triggerData["alert_name"].(string); ok {
			params["query"] = fmt.Sprintf("ALERTS{alertname=%q}", alertName)
		}
		params["time_range"] = "30m"
	case "logs_query":
		if host, ok := triggerCtx["host"].(string); ok {
			params["host"] = host
		}
		params["level"] = "error"
		params["time_range"] = "30m"
	case "alert_correlate", "alert_list":
		params["time_range"] = "1h"
		if host, ok := triggerCtx["host"].(string); ok {
			params["host"] = host
		}
	case "cmdb_query", "cmdb_topology":
		if host, ok := triggerCtx["host"].(string); ok {
			params["target"] = host
		}
		if service, ok := triggerCtx["service"].(string); ok {
			params["service"] = service
		}
	case "metrics_snapshot":
		if host, ok := triggerCtx["host"].(string); ok {
			params["host"] = host
		}
	case "knowledge_search":
		if alertName, ok := triggerData["alert_name"].(string); ok {
			params["keyword"] = alertName
		}
	case "logs_pattern_analysis":
		if host, ok := triggerCtx["host"].(string); ok {
			params["host"] = host
		}
		params["time_range"] = "1h"
	}

	return params
}

// executeTool 执行工具调用
func (inv *AutonomousInvestigator) executeTool(ctx context.Context, toolName string, params map[string]any) (any, error) {
	// 通过 ToolRegistry 验证工具存在
	inv.mu.RLock()
	registry := inv.toolRegistry
	inv.mu.RUnlock()

	if registry == nil {
		return nil, fmt.Errorf("tool registry not initialized")
	}

	tools := registry.GetAll()
	found := false
	for _, t := range tools {
		if t.Name == toolName {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("tool %q not found in registry", toolName)
	}

	// 实际执行委托给 handler 层的工具执行器
	// 这里返回占位结果，实际集成时由 handler 层注入执行函数
	return map[string]any{
		"tool":      toolName,
		"params":    params,
		"status":    "executed",
		"timestamp": time.Now().Unix(),
	}, nil
}

// analyzeAndDecide 分析工具结果，决定继续/结论/升级
func (inv *AutonomousInvestigator) analyzeAndDecide(investigation *Investigation, step InvestigationStep) string {
	// 检查是否有高置信度假设
	for _, h := range investigation.Hypotheses {
		if h.Confidence >= 0.85 && !h.Refuted {
			return "conclude"
		}
	}

	// 检查是否所有假设都被推翻
	allRefuted := true
	for _, h := range investigation.Hypotheses {
		if !h.Refuted {
			allRefuted = false
			break
		}
	}
	if allRefuted {
		return "escalate"
	}

	// 如果已经是最后一轮，输出结论
	if investigation.CurrentRound >= investigation.MaxRounds {
		return "conclude"
	}

	return "continue"
}

// formConclusion 综合所有证据形成结论
func (inv *AutonomousInvestigator) formConclusion(investigation *Investigation) *Conclusion {
	// 找到置信度最高的未推翻假设
	var bestHypothesis *Hypothesis
	for i := range investigation.Hypotheses {
		h := &investigation.Hypotheses[i]
		if h.Refuted {
			continue
		}
		if bestHypothesis == nil || h.Confidence > bestHypothesis.Confidence {
			bestHypothesis = h
		}
	}

	conclusion := &Conclusion{
		RootCause:     "未能确定根因",
		Confidence:    0.0,
		Evidence:      make([]string, 0),
		Suggestions:   []string{"建议人工介入进一步排查"},
		AffectedScope: "unknown",
	}

	if bestHypothesis != nil {
		conclusion.RootCause = bestHypothesis.Description
		conclusion.Confidence = bestHypothesis.Confidence
		conclusion.Evidence = bestHypothesis.Evidence
		conclusion.Suggestions = inv.generateSuggestions(investigation, bestHypothesis)
	}

	// 从调查步骤中收集证据
	for _, step := range investigation.Steps {
		if step.Result != nil {
			evidence := fmt.Sprintf("[%s] %s", step.ToolUsed, step.Reasoning)
			conclusion.Evidence = append(conclusion.Evidence, evidence)
		}
	}

	// 确定影响范围
	if scope, ok := investigation.Trigger.Context["service"].(string); ok {
		conclusion.AffectedScope = scope
	} else if host, ok := investigation.Trigger.Context["host"].(string); ok {
		conclusion.AffectedScope = host
	}

	return conclusion
}

// generateSuggestions 根据调查结果生成修复建议
func (inv *AutonomousInvestigator) generateSuggestions(_ *Investigation, h *Hypothesis) []string {
	suggestions := []string{
		fmt.Sprintf("根因假设：%s（置信度 %.0f%%）", h.Description, h.Confidence*100),
		"建议查看相关 Runbook 获取修复步骤",
		"如需自动修复，请确认后触发自愈工作流",
	}
	return suggestions
}

