package aiengine

import (
	"sync"
)

// ---------------------------------------------------------------------------
// ToolRegistry — 按意图过滤的工具注册表
// ---------------------------------------------------------------------------
//
// 核心理念：不注入全部工具，根据意图只注入相关工具。
// 减少 token 消耗，提高 LLM 工具选择准确率。

// ToolDef 工具定义（轻量版，用于 context 注入）
type ToolDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	RiskLevel   int        `json:"risk_level"` // 0=安全, 1=低风险, 2=中风险, 3=高风险
	Intents     []IntentType `json:"intents"`  // 适用的意图类型
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]*ToolDef
}

// NewToolRegistry 创建工具注册表并注册默认工具
func NewToolRegistry() *ToolRegistry {
	tr := &ToolRegistry{
		tools: make(map[string]*ToolDef),
	}
	tr.registerDefaults()
	return tr
}

// Register 注册一个工具
func (tr *ToolRegistry) Register(tool *ToolDef) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.tools[tool.Name] = tool
}

// GetToolsByIntent 根据意图获取相关工具列表
func (tr *ToolRegistry) GetToolsByIntent(intent IntentType) []*ToolDef {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	result := make([]*ToolDef, 0)
	for _, tool := range tr.tools {
		for _, i := range tool.Intents {
			if i == intent {
				result = append(result, tool)
				break
			}
		}
	}
	return result
}

// GetToolsByCategory 根据分类获取工具列表
func (tr *ToolRegistry) GetToolsByCategory(category string) []*ToolDef {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	result := make([]*ToolDef, 0)
	for _, tool := range tr.tools {
		if tool.Category == category {
			result = append(result, tool)
		}
	}
	return result
}

// GetAll 获取所有工具
func (tr *ToolRegistry) GetAll() []*ToolDef {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	result := make([]*ToolDef, 0, len(tr.tools))
	for _, tool := range tr.tools {
		result = append(result, tool)
	}
	return result
}

// registerDefaults 注册默认工具集
func (tr *ToolRegistry) registerDefaults() {
	// --- 指标类工具 ---
	tr.tools["prometheus_query"] = &ToolDef{
		Name:        "prometheus_query",
		Description: "查询 Prometheus 即时指标",
		Category:    "metrics",
		RiskLevel:   0,
		Intents:     []IntentType{IntentQuery, IntentDiagnose, IntentSelfHeal, IntentTopology},
	}
	tr.tools["prometheus_query_range"] = &ToolDef{
		Name:        "prometheus_query_range",
		Description: "查询 Prometheus 时间范围指标",
		Category:    "metrics",
		RiskLevel:   0,
		Intents:     []IntentType{IntentQuery, IntentDiagnose, IntentSelfHeal},
	}
	tr.tools["metrics_snapshot"] = &ToolDef{
		Name:        "metrics_snapshot",
		Description: "获取主机/服务关键指标快照",
		Category:    "metrics",
		RiskLevel:   0,
		Intents:     []IntentType{IntentDiagnose, IntentSelfHeal, IntentTopology},
	}

	// --- 日志类工具 ---
	tr.tools["logs_query"] = &ToolDef{
		Name:        "logs_query",
		Description: "查询日志（支持关键词、时间范围、级别过滤）",
		Category:    "logs",
		RiskLevel:   0,
		Intents:     []IntentType{IntentQuery, IntentDiagnose},
	}
	tr.tools["logs_pattern_analysis"] = &ToolDef{
		Name:        "logs_pattern_analysis",
		Description: "日志模式分析（聚类、异常检测）",
		Category:    "logs",
		RiskLevel:   0,
		Intents:     []IntentType{IntentDiagnose},
	}

	// --- CMDB 类工具 ---
	tr.tools["cmdb_query"] = &ToolDef{
		Name:        "cmdb_query",
		Description: "查询 CMDB 资产信息（主机、服务、业务组）",
		Category:    "cmdb",
		RiskLevel:   0,
		Intents:     []IntentType{IntentQuery, IntentDiagnose, IntentSelfHeal, IntentTopology},
	}
	tr.tools["cmdb_topology"] = &ToolDef{
		Name:        "cmdb_topology",
		Description: "查询 CMDB 拓扑关系（上下游依赖）",
		Category:    "cmdb",
		RiskLevel:   0,
		Intents:     []IntentType{IntentTopology, IntentDiagnose},
	}

	// --- 告警类工具 ---
	tr.tools["alert_list"] = &ToolDef{
		Name:        "alert_list",
		Description: "查询当前活跃告警列表",
		Category:    "alert",
		RiskLevel:   0,
		Intents:     []IntentType{IntentQuery, IntentDiagnose, IntentSelfHeal, IntentTopology},
	}
	tr.tools["alert_correlate"] = &ToolDef{
		Name:        "alert_correlate",
		Description: "告警关联分析（时间窗口内相关告警）",
		Category:    "alert",
		RiskLevel:   0,
		Intents:     []IntentType{IntentDiagnose, IntentSelfHeal},
	}
	tr.tools["alert_silence"] = &ToolDef{
		Name:        "alert_silence",
		Description: "静默告警（需确认）",
		Category:    "alert",
		RiskLevel:   2,
		Intents:     []IntentType{IntentSelfHeal},
	}

	// --- 执行类工具 ---
	tr.tools["script_execute"] = &ToolDef{
		Name:        "script_execute",
		Description: "在目标主机执行脚本（沙箱模式）",
		Category:    "agent",
		RiskLevel:   3,
		Intents:     []IntentType{IntentScript, IntentSelfHeal},
	}
	tr.tools["service_restart"] = &ToolDef{
		Name:        "service_restart",
		Description: "重启服务（需人工确认）",
		Category:    "agent",
		RiskLevel:   3,
		Intents:     []IntentType{IntentSelfHeal},
	}

	// --- 知识库类工具 ---
	tr.tools["knowledge_search"] = &ToolDef{
		Name:        "knowledge_search",
		Description: "搜索知识库（Runbook、历史案例）",
		Category:    "knowledge",
		RiskLevel:   0,
		Intents:     []IntentType{IntentQuery, IntentDiagnose, IntentScript, IntentSelfHeal},
	}
	tr.tools["runbook_match"] = &ToolDef{
		Name:        "runbook_match",
		Description: "匹配推荐 Runbook",
		Category:    "knowledge",
		RiskLevel:   0,
		Intents:     []IntentType{IntentDiagnose, IntentSelfHeal},
	}
}
