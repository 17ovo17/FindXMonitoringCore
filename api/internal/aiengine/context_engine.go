package aiengine

import (
	"sort"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------------
// ContextEngine — 动态构建 AI 对话的上下文
// ---------------------------------------------------------------------------
//
// 核心理念：Context Engineering — 不堆积无效历史，动态选择最相关的上下文注入
//
// 工作流程：
// 1. 接收用户意图（query）
// 2. 意图分类（问答/脚本/拓扑分析/故障自愈/诊断分析）
// 3. 根据意图动态组装 context：
//    - 系统 prompt（角色定义 + 能力边界）
//    - 相关 CMDB 上下文（当前主机/服务/业务组）
//    - 相关告警上下文（当前 firing 的告警）
//    - 相关指标快照（最近 5 分钟关键指标）
//    - 历史诊断摘要（不是完整历史，是压缩后的摘要）
//    - 可用工具列表（根据意图过滤，不注入全部工具）
// 4. 输出结构化 context（不超过 token budget）

// IntentType 表示用户意图类型
type IntentType string

const (
	IntentQuery    IntentType = "query"     // 智能问答
	IntentScript   IntentType = "script"    // 运维脚本
	IntentTopology IntentType = "topology"  // 拓扑分析
	IntentSelfHeal IntentType = "self_heal" // 故障自愈
	IntentDiagnose IntentType = "diagnose"  // 诊断分析
)

// ContextBlockRole 定义上下文块的角色
type ContextBlockRole string

const (
	BlockRoleSystem  ContextBlockRole = "system"  // 系统 prompt
	BlockRoleContext ContextBlockRole = "context" // 业务上下文
	BlockRoleHistory ContextBlockRole = "history" // 历史摘要
	BlockRoleTools   ContextBlockRole = "tools"   // 可用工具
)

// ContextBlock 表示一个上下文片段
type ContextBlock struct {
	Role     ContextBlockRole `json:"role"`
	Content  string           `json:"content"`
	Priority int              `json:"priority"` // 越高越优先保留
	Tokens   int              `json:"tokens"`   // 预估 token 数
	Source   string           `json:"source"`   // 来源标识（用于审计）
}

// defaultTokenBudget 默认 token 预算
const defaultTokenBudget = 8000

// defaultCharsPerTokenEst 每 token 约 4 字符（中文约 2 字符/token，取平均）
const defaultCharsPerTokenEst = 3

// ContextEngine 动态构建 AI 对话的上下文
type ContextEngine struct {
	mu            sync.RWMutex
	toolRegistry  *ToolRegistry
	memoryManager *MemoryManager
	curator       *Curator
	compressor    *ContextCompressor
	promptBuilder *PromptBuilder
	tokenBudget   int
}

// ContextEngineConfig 配置项
type ContextEngineConfig struct {
	TokenBudget int
}

// NewContextEngine 创建上下文引擎实例
func NewContextEngine(cfg ContextEngineConfig) *ContextEngine {
	budget := cfg.TokenBudget
	if budget <= 0 {
		budget = defaultTokenBudget
	}
	mm := NewMemoryManager()
	tr := NewToolRegistry()
	pb := NewPromptBuilder()
	cur := NewCurator(mm, 5)
	comp := NewContextCompressor(budget)

	return &ContextEngine{
		toolRegistry:  tr,
		memoryManager: mm,
		curator:       cur,
		compressor:    comp,
		promptBuilder: pb,
		tokenBudget:   budget,
	}
}

// BuildContext 根据意图动态构建上下文
func (e *ContextEngine) BuildContext(sessionID string, query string, intent IntentType) ([]ContextBlock, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	blocks := make([]ContextBlock, 0, 8)

	// 1. 系统 prompt（最高优先级，始终保留）
	blocks = append(blocks, e.buildSystemPrompt(intent))

	// 2. 根据意图注入相关上下文
	switch intent {
	case IntentQuery:
		blocks = append(blocks, e.buildQueryContext(query)...)
	case IntentScript:
		blocks = append(blocks, e.buildScriptContext(query)...)
	case IntentTopology:
		blocks = append(blocks, e.buildTopologyContext(query)...)
	case IntentSelfHeal:
		blocks = append(blocks, e.buildSelfHealContext(query)...)
	case IntentDiagnose:
		blocks = append(blocks, e.buildDiagnoseContext(query)...)
	}

	// 3. 注入历史摘要（压缩后的，不是完整历史）
	blocks = append(blocks, e.curator.SelectRelevantHistory(sessionID, query)...)

	// 4. 注入可用工具（根据意图过滤）
	blocks = append(blocks, e.buildToolContext(intent))

	// 5. Token budget 裁剪（按优先级保留）
	return e.compressor.FitBudget(blocks, e.tokenBudget), nil
}

// ClassifyIntent 意图分类 — 基于关键词 + 模式匹配的快速分类（不调用 LLM）
func (e *ContextEngine) ClassifyIntent(query string) IntentType {
	lower := strings.ToLower(query)

	// 自愈类关键词优先级最高（涉及自动操作）
	selfHealKeywords := []string{"自愈", "修复", "恢复", "重启", "回滚", "扩容", "缩容", "self-heal", "recover", "restart", "rollback"}
	for _, kw := range selfHealKeywords {
		if strings.Contains(lower, kw) {
			return IntentSelfHeal
		}
	}

	// 诊断类
	diagnoseKeywords := []string{"诊断", "分析", "为什么", "根因", "排查", "定位", "故障", "异常", "diagnose", "root cause", "why"}
	for _, kw := range diagnoseKeywords {
		if strings.Contains(lower, kw) {
			return IntentDiagnose
		}
	}

	// 拓扑类
	topologyKeywords := []string{"拓扑", "架构", "依赖", "上下游", "链路", "调用关系", "topology", "dependency", "upstream", "downstream"}
	for _, kw := range topologyKeywords {
		if strings.Contains(lower, kw) {
			return IntentTopology
		}
	}

	// 脚本类
	scriptKeywords := []string{"脚本", "命令", "怎么写", "shell", "script", "bash", "ansible", "playbook", "执行"}
	for _, kw := range scriptKeywords {
		if strings.Contains(lower, kw) {
			return IntentScript
		}
	}

	// 默认为问答
	return IntentQuery
}

// SetTokenBudget 动态调整 token 预算
func (e *ContextEngine) SetTokenBudget(budget int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if budget > 0 {
		e.tokenBudget = budget
	}
}

// GetToolRegistry 获取工具注册表（供外部集成）
func (e *ContextEngine) GetToolRegistry() *ToolRegistry {
	return e.toolRegistry
}

// GetMemoryManager 获取记忆管理器（供外部集成）
func (e *ContextEngine) GetMemoryManager() *MemoryManager {
	return e.memoryManager
}

// ---------------------------------------------------------------------------
// 内部方法：构建各类上下文
// ---------------------------------------------------------------------------

func (e *ContextEngine) buildSystemPrompt(intent IntentType) ContextBlock {
	tmpl := e.promptBuilder.GetTemplate(intent)
	content := tmpl.Build()
	return ContextBlock{
		Role:     BlockRoleSystem,
		Content:  content,
		Priority: 100, // 最高优先级
		Tokens:   estimateTokens(content),
		Source:   "system_prompt",
	}
}

func (e *ContextEngine) buildQueryContext(query string) []ContextBlock {
	blocks := make([]ContextBlock, 0, 2)
	// 注入相关 CMDB 上下文
	blocks = append(blocks, ContextBlock{
		Role:     BlockRoleContext,
		Content:  "[CMDB 上下文] 根据查询关键词匹配相关资产信息",
		Priority: 60,
		Tokens:   estimateTokens("[CMDB 上下文]"),
		Source:   "cmdb",
	})
	// 注入相关记忆
	memories := e.memoryManager.Recall(query, 3)
	if len(memories) > 0 {
		var sb strings.Builder
		sb.WriteString("[历史记忆]\n")
		for _, m := range memories {
			sb.WriteString("- ")
			sb.WriteString(m.Content)
			sb.WriteString("\n")
		}
		content := sb.String()
		blocks = append(blocks, ContextBlock{
			Role:     BlockRoleContext,
			Content:  content,
			Priority: 40,
			Tokens:   estimateTokens(content),
			Source:   "memory",
		})
	}
	return blocks
}

func (e *ContextEngine) buildScriptContext(query string) []ContextBlock {
	blocks := make([]ContextBlock, 0, 2)
	// 注入平台约束（OS 类型、可用工具等）
	blocks = append(blocks, ContextBlock{
		Role:     BlockRoleContext,
		Content:  "[平台约束] 目标环境: Linux (CentOS/Ubuntu), 可用工具: bash/python/ansible",
		Priority: 70,
		Tokens:   estimateTokens("[平台约束]"),
		Source:   "platform",
	})
	// 注入相关 Runbook 模板
	memories := e.memoryManager.RecallByType(query, MemoryRunbook, 2)
	if len(memories) > 0 {
		var sb strings.Builder
		sb.WriteString("[相关 Runbook]\n")
		for _, m := range memories {
			sb.WriteString("- ")
			sb.WriteString(m.Content)
			sb.WriteString("\n")
		}
		content := sb.String()
		blocks = append(blocks, ContextBlock{
			Role:     BlockRoleContext,
			Content:  content,
			Priority: 50,
			Tokens:   estimateTokens(content),
			Source:   "runbook",
		})
	}
	return blocks
}

func (e *ContextEngine) buildTopologyContext(query string) []ContextBlock {
	blocks := make([]ContextBlock, 0, 2)
	// 注入 CMDB 拓扑关系
	blocks = append(blocks, ContextBlock{
		Role:     BlockRoleContext,
		Content:  "[拓扑上下文] CMDB 服务依赖关系、上下游调用链",
		Priority: 80,
		Tokens:   estimateTokens("[拓扑上下文]"),
		Source:   "cmdb_topology",
	})
	// 注入当前告警状态（影响拓扑分析）
	blocks = append(blocks, ContextBlock{
		Role:     BlockRoleContext,
		Content:  "[告警状态] 当前 firing 告警列表（用于标注拓扑异常节点）",
		Priority: 60,
		Tokens:   estimateTokens("[告警状态]"),
		Source:   "alert_state",
	})
	return blocks
}

func (e *ContextEngine) buildSelfHealContext(query string) []ContextBlock {
	blocks := make([]ContextBlock, 0, 3)
	// 注入告警详情
	blocks = append(blocks, ContextBlock{
		Role:     BlockRoleContext,
		Content:  "[告警详情] 触发自愈的告警事件详细信息",
		Priority: 90,
		Tokens:   estimateTokens("[告警详情]"),
		Source:   "alert_detail",
	})
	// 注入 CMDB 上下文（影响范围评估）
	blocks = append(blocks, ContextBlock{
		Role:     BlockRoleContext,
		Content:  "[CMDB 上下文] 目标主机/服务的业务归属、集群信息、SLA 等级",
		Priority: 80,
		Tokens:   estimateTokens("[CMDB 上下文]"),
		Source:   "cmdb",
	})
	// 注入历史修复方案
	memories := e.memoryManager.RecallByType(query, MemoryRunbook, 3)
	if len(memories) > 0 {
		var sb strings.Builder
		sb.WriteString("[历史修复方案]\n")
		for _, m := range memories {
			sb.WriteString("- ")
			sb.WriteString(m.Content)
			sb.WriteString("\n")
		}
		content := sb.String()
		blocks = append(blocks, ContextBlock{
			Role:     BlockRoleContext,
			Content:  content,
			Priority: 70,
			Tokens:   estimateTokens(content),
			Source:   "history_fix",
		})
	}
	return blocks
}

func (e *ContextEngine) buildDiagnoseContext(query string) []ContextBlock {
	blocks := make([]ContextBlock, 0, 3)
	// 注入指标快照
	blocks = append(blocks, ContextBlock{
		Role:     BlockRoleContext,
		Content:  "[指标快照] 最近 5 分钟关键指标（CPU/内存/磁盘/网络/连接数）",
		Priority: 80,
		Tokens:   estimateTokens("[指标快照]"),
		Source:   "metrics_snapshot",
	})
	// 注入告警关联
	blocks = append(blocks, ContextBlock{
		Role:     BlockRoleContext,
		Content:  "[告警关联] 同时段相关告警事件（时间窗口 10 分钟）",
		Priority: 70,
		Tokens:   estimateTokens("[告警关联]"),
		Source:   "alert_correlate",
	})
	// 注入历史诊断结论
	memories := e.memoryManager.RecallByType(query, MemoryDiagnosis, 3)
	if len(memories) > 0 {
		var sb strings.Builder
		sb.WriteString("[历史诊断]\n")
		for _, m := range memories {
			sb.WriteString("- ")
			sb.WriteString(m.Content)
			sb.WriteString("\n")
		}
		content := sb.String()
		blocks = append(blocks, ContextBlock{
			Role:     BlockRoleContext,
			Content:  content,
			Priority: 60,
			Tokens:   estimateTokens(content),
			Source:   "history_diagnosis",
		})
	}
	return blocks
}

func (e *ContextEngine) buildToolContext(intent IntentType) ContextBlock {
	tools := e.toolRegistry.GetToolsByIntent(intent)
	var sb strings.Builder
	sb.WriteString("[可用工具]\n")
	for _, t := range tools {
		sb.WriteString("- ")
		sb.WriteString(t.Name)
		sb.WriteString(": ")
		sb.WriteString(t.Description)
		sb.WriteString("\n")
	}
	content := sb.String()
	return ContextBlock{
		Role:     BlockRoleTools,
		Content:  content,
		Priority: 50,
		Tokens:   estimateTokens(content),
		Source:   "tool_registry",
	}
}

// estimateTokens 粗略估算 token 数
func estimateTokens(s string) int {
	if len(s) == 0 {
		return 0
	}
	// 中英文混合场景：按字符数 / 3 估算
	return len([]rune(s))/defaultCharsPerTokenEst + 1
}

// ---------------------------------------------------------------------------
// ContextCompressor — token budget 裁剪器
// ---------------------------------------------------------------------------

// ContextCompressor 按优先级裁剪上下文块以适应 token 预算
type ContextCompressor struct {
	maxTokens int
}

// NewContextCompressor 创建裁剪器
func NewContextCompressor(maxTokens int) *ContextCompressor {
	return &ContextCompressor{maxTokens: maxTokens}
}

// FitBudget 按优先级保留上下文块，总 token 不超过 budget
func (c *ContextCompressor) FitBudget(blocks []ContextBlock, budget int) []ContextBlock {
	if budget <= 0 {
		budget = c.maxTokens
	}

	// 按优先级降序排列
	sorted := make([]ContextBlock, len(blocks))
	copy(sorted, blocks)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})

	// 贪心选择：按优先级从高到低，直到超出预算
	result := make([]ContextBlock, 0, len(sorted))
	totalTokens := 0
	for _, block := range sorted {
		if totalTokens+block.Tokens > budget {
			// 尝试截断内容以适应剩余预算
			remaining := budget - totalTokens
			if remaining > 50 { // 至少保留 50 token 才有意义
				runes := []rune(block.Content)
				maxChars := remaining * defaultCharsPerTokenEst
				if maxChars < len(runes) {
					block.Content = string(runes[:maxChars]) + "\n...[已截断]"
					block.Tokens = remaining
				}
				result = append(result, block)
			}
			break
		}
		result = append(result, block)
		totalTokens += block.Tokens
	}

	// 恢复原始顺序（system > context > history > tools）
	sort.Slice(result, func(i, j int) bool {
		return blockRoleOrder(result[i].Role) < blockRoleOrder(result[j].Role)
	})

	return result
}

// blockRoleOrder 定义角色的输出顺序
func blockRoleOrder(role ContextBlockRole) int {
	switch role {
	case BlockRoleSystem:
		return 0
	case BlockRoleContext:
		return 1
	case BlockRoleHistory:
		return 2
	case BlockRoleTools:
		return 3
	default:
		return 9
	}
}
