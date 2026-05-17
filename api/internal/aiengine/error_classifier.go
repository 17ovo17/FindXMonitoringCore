package aiengine

import (
	"strings"
)

// ---------------------------------------------------------------------------
// ErrorClassifier — 对告警/错误进行分类，自动路由到对应工作流
// ---------------------------------------------------------------------------
//
// 分类维度：
// - 基础设施层（CPU/内存/磁盘/网络）
// - 应用层（OOM/连接池/超时/死锁）
// - 数据层（主从延迟/连接数/慢查询/空间不足）
// - 中间件层（队列堆积/连接断开/配置错误）
//
// 参考 Hermes error_classifier.py：结构化分类 + 恢复策略路由

// ErrorLayer 错误所在层级
type ErrorLayer string

const (
	LayerInfrastructure ErrorLayer = "infrastructure" // 基础设施层
	LayerApplication    ErrorLayer = "application"    // 应用层
	LayerData           ErrorLayer = "data"           // 数据层
	LayerMiddleware     ErrorLayer = "middleware"     // 中间件层
	LayerNetwork        ErrorLayer = "network"        // 网络层
	LayerUnknown        ErrorLayer = "unknown"        // 未知
)

// ErrorSeverity 错误严重程度
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "critical" // P0 - 立即处理
	SeverityMajor    ErrorSeverity = "major"    // P1 - 尽快处理
	SeverityMinor    ErrorSeverity = "minor"    // P2 - 计划处理
	SeverityWarning  ErrorSeverity = "warning"  // P3 - 关注
)

// ErrorCategory 错误分类结果
type ErrorCategory struct {
	Layer    ErrorLayer    `json:"layer"`    // 所在层级
	Type     string        `json:"type"`     // 具体类型（cpu_high, oom, slow_query...）
	Severity ErrorSeverity `json:"severity"` // 严重程度
	Workflow string        `json:"workflow"` // 推荐的自愈工作流 ID
	Tags     []string      `json:"tags"`     // 附加标签
}

// AlertEvent 告警事件（简化版，与 model.MonitorAlertEvent 对应）
type AlertEvent struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Labels   map[string]string `json:"labels"`
	Value    float64           `json:"value"`
	Severity string            `json:"severity"`
	Status   string            `json:"status"`
}

// ErrorClassifier 错误分类器
type ErrorClassifier struct {
	rules []classifyRule
}

// classifyRule 分类规则
type classifyRule struct {
	keywords []string
	labels   map[string]string // label key -> value pattern
	category ErrorCategory
}

// NewErrorClassifier 创建错误分类器并注册默认规则
func NewErrorClassifier() *ErrorClassifier {
	ec := &ErrorClassifier{
		rules: make([]classifyRule, 0, 32),
	}
	ec.registerDefaultRules()
	return ec
}

// Classify 根据告警标签和指标分类错误
func (ec *ErrorClassifier) Classify(event *AlertEvent) ErrorCategory {
	title := strings.ToLower(event.Title)
	labels := event.Labels

	// 按规则优先级匹配
	for _, rule := range ec.rules {
		if ec.matchRule(title, labels, rule) {
			category := rule.category
			// 根据告警原始 severity 调整
			if event.Severity == "critical" || event.Severity == "P0" {
				category.Severity = SeverityCritical
			}
			return category
		}
	}

	// 兜底分类
	return ErrorCategory{
		Layer:    LayerUnknown,
		Type:     "unclassified",
		Severity: ec.inferSeverity(event),
		Workflow: "alert_diagnosis",
		Tags:     []string{"needs_manual_classification"},
	}
}

// RouteToWorkflow 根据分类结果路由到对应工作流
func (ec *ErrorClassifier) RouteToWorkflow(category ErrorCategory) string {
	if category.Workflow != "" {
		return category.Workflow
	}
	// 按层级默认路由
	switch category.Layer {
	case LayerInfrastructure:
		return "health_inspection"
	case LayerApplication:
		return "alert_diagnosis"
	case LayerData:
		return "slow_query_diagnosis"
	case LayerMiddleware:
		return "middleware_inspection"
	case LayerNetwork:
		return "network_check"
	default:
		return "alert_diagnosis"
	}
}

// ClassifyBatch 批量分类（用于告警聚合场景）
func (ec *ErrorClassifier) ClassifyBatch(events []*AlertEvent) map[ErrorLayer][]*AlertEvent {
	result := make(map[ErrorLayer][]*AlertEvent)
	for _, event := range events {
		category := ec.Classify(event)
		result[category.Layer] = append(result[category.Layer], event)
	}
	return result
}

// ---------------------------------------------------------------------------
// 内部方法
// ---------------------------------------------------------------------------

func (ec *ErrorClassifier) matchRule(title string, labels map[string]string, rule classifyRule) bool {
	// 关键词匹配
	keywordMatch := false
	if len(rule.keywords) == 0 {
		keywordMatch = true
	} else {
		for _, kw := range rule.keywords {
			if strings.Contains(title, kw) {
				keywordMatch = true
				break
			}
		}
	}
	if !keywordMatch {
		return false
	}

	// 标签匹配
	for key, pattern := range rule.labels {
		val, ok := labels[key]
		if !ok {
			return false
		}
		if !strings.Contains(strings.ToLower(val), strings.ToLower(pattern)) {
			return false
		}
	}

	return true
}

func (ec *ErrorClassifier) inferSeverity(event *AlertEvent) ErrorSeverity {
	switch strings.ToLower(event.Severity) {
	case "critical", "p0":
		return SeverityCritical
	case "major", "p1", "high":
		return SeverityMajor
	case "minor", "p2", "medium":
		return SeverityMinor
	default:
		return SeverityWarning
	}
}

func (ec *ErrorClassifier) registerDefaultRules() {
	// --- 基础设施层 ---
	ec.rules = append(ec.rules,
		classifyRule{
			keywords: []string{"cpu", "load", "负载"},
			category: ErrorCategory{Layer: LayerInfrastructure, Type: "cpu_high", Severity: SeverityMajor, Workflow: "health_inspection"},
		},
		classifyRule{
			keywords: []string{"memory", "内存", "mem_used", "swap"},
			category: ErrorCategory{Layer: LayerInfrastructure, Type: "memory_high", Severity: SeverityMajor, Workflow: "health_inspection"},
		},
		classifyRule{
			keywords: []string{"disk", "磁盘", "空间", "inode"},
			category: ErrorCategory{Layer: LayerInfrastructure, Type: "disk_full", Severity: SeverityMajor, Workflow: "storage_health_check"},
		},
		classifyRule{
			keywords: []string{"network", "网络", "丢包", "延迟", "unreachable"},
			category: ErrorCategory{Layer: LayerNetwork, Type: "network_issue", Severity: SeverityMajor, Workflow: "network_check"},
		},
	)

	// --- 应用层 ---
	ec.rules = append(ec.rules,
		classifyRule{
			keywords: []string{"oom", "out of memory", "killed"},
			category: ErrorCategory{Layer: LayerApplication, Type: "oom", Severity: SeverityCritical, Workflow: "alert_diagnosis"},
		},
		classifyRule{
			keywords: []string{"connection pool", "连接池", "too many connections"},
			category: ErrorCategory{Layer: LayerApplication, Type: "connection_pool_exhausted", Severity: SeverityMajor, Workflow: "alert_diagnosis"},
		},
		classifyRule{
			keywords: []string{"timeout", "超时", "timed out"},
			category: ErrorCategory{Layer: LayerApplication, Type: "timeout", Severity: SeverityMinor, Workflow: "alert_diagnosis"},
		},
		classifyRule{
			keywords: []string{"deadlock", "死锁"},
			category: ErrorCategory{Layer: LayerApplication, Type: "deadlock", Severity: SeverityCritical, Workflow: "db_lock_analysis"},
		},
		classifyRule{
			keywords: []string{"jvm", "gc", "heap", "fullgc"},
			category: ErrorCategory{Layer: LayerApplication, Type: "jvm_issue", Severity: SeverityMajor, Workflow: "jvm_diagnosis"},
		},
	)

	// --- 数据层 ---
	ec.rules = append(ec.rules,
		classifyRule{
			keywords: []string{"replication", "主从", "slave", "replica", "延迟"},
			labels:   map[string]string{"component": "mysql"},
			category: ErrorCategory{Layer: LayerData, Type: "replication_lag", Severity: SeverityMajor, Workflow: "alert_diagnosis"},
		},
		classifyRule{
			keywords: []string{"slow query", "慢查询", "slow_queries"},
			category: ErrorCategory{Layer: LayerData, Type: "slow_query", Severity: SeverityMinor, Workflow: "slow_query_diagnosis"},
		},
		classifyRule{
			keywords: []string{"tablespace", "数据库空间", "db_size"},
			category: ErrorCategory{Layer: LayerData, Type: "db_space_full", Severity: SeverityMajor, Workflow: "storage_health_check"},
		},
	)

	// --- 中间件层 ---
	ec.rules = append(ec.rules,
		classifyRule{
			keywords: []string{"queue", "队列", "堆积", "lag", "consumer"},
			category: ErrorCategory{Layer: LayerMiddleware, Type: "queue_backlog", Severity: SeverityMajor, Workflow: "middleware_inspection"},
		},
		classifyRule{
			keywords: []string{"redis", "cache", "缓存"},
			category: ErrorCategory{Layer: LayerMiddleware, Type: "cache_issue", Severity: SeverityMinor, Workflow: "middleware_inspection"},
		},
		classifyRule{
			keywords: []string{"nginx", "upstream", "502", "503"},
			category: ErrorCategory{Layer: LayerMiddleware, Type: "proxy_error", Severity: SeverityMajor, Workflow: "alert_diagnosis"},
		},
	)
}
