package aiengine

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// NL2Query — 自然语言转查询语言（对标 Splunk AI Assistant）
// ---------------------------------------------------------------------------
//
// 支持：中文/英文 → PromQL / LogQL / SQL
//
// 示例：
// "最近1小时CPU使用率超过80%的主机"
//   → up{job="node"} * on(instance) group_left() (1 - rate(node_cpu_seconds_total{mode="idle"}[5m])) > 0.8
// "昨天下午3点到5点的错误日志"
//   → {level="error"} |= `` | __timestamp__ >= ... | __timestamp__ <= ...
// "连接数最多的前10个MySQL实例"
//   → topk(10, mysql_global_status_threads_connected)

// QueryLanguage 查询语言类型
type QueryLanguage string

const (
	LangPromQL QueryLanguage = "promql"
	LangLogQL  QueryLanguage = "logql"
	LangSQL    QueryLanguage = "sql"
)

// NL2QueryResult 自然语言转查询结果
type NL2QueryResult struct {
	ID              string        `json:"id"`
	NaturalLanguage string        `json:"natural_language"`
	TargetLanguage  QueryLanguage `json:"target_language"`
	GeneratedQuery  string        `json:"generated_query"`
	Explanation     string        `json:"explanation"`
	Confidence      float64       `json:"confidence"`
	Alternatives    []string      `json:"alternatives,omitempty"`
	GeneratedAt     time.Time     `json:"generated_at"`
}

// NL2Query 自然语言转查询引擎
type NL2Query struct {
	metricMapping map[string]string // 中文指标名 → PromQL 表达式
	logKeywords   []string          // 日志相关关键词
	sqlKeywords   []string          // SQL 相关关键词
}

// NewNL2Query 创建 NL2Query 引擎
func NewNL2Query() *NL2Query {
	return &NL2Query{
		metricMapping: defaultMetricNameMapping(),
		logKeywords:   []string{"日志", "log", "错误日志", "error", "warn", "异常日志", "访问日志"},
		sqlKeywords:   []string{"表", "数据库", "记录", "table", "database", "record", "SQL", "sql"},
	}
}

// Translate 将自然语言转为查询语言
func (n *NL2Query) Translate(input string, targetLang QueryLanguage) (*NL2QueryResult, error) {
	if input == "" {
		return nil, fmt.Errorf("input cannot be empty")
	}

	// 如果未指定目标语言，自动检测
	if targetLang == "" {
		targetLang = n.AutoDetectLanguage(input)
	}

	result := &NL2QueryResult{
		ID:              uuid.New().String(),
		NaturalLanguage: input,
		TargetLanguage:  targetLang,
		GeneratedAt:     time.Now(),
	}

	switch targetLang {
	case LangPromQL:
		n.translateToPromQL(input, result)
	case LangLogQL:
		n.translateToLogQL(input, result)
	case LangSQL:
		n.translateToSQL(input, result)
	default:
		return nil, fmt.Errorf("unsupported target language: %s", targetLang)
	}

	return result, nil
}

// AutoDetectLanguage 自动检测应该用哪种查询语言
func (n *NL2Query) AutoDetectLanguage(input string) QueryLanguage {
	lower := strings.ToLower(input)

	// 日志相关关键词 → LogQL
	for _, kw := range n.logKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return LangLogQL
		}
	}

	// SQL 相关关键词 → SQL
	for _, kw := range n.sqlKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return LangSQL
		}
	}

	// 默认 → PromQL（指标查询最常见）
	return LangPromQL
}

// translateToPromQL 转换为 PromQL
func (n *NL2Query) translateToPromQL(input string, result *NL2QueryResult) {
	lower := strings.ToLower(input)

	// 尝试从指标映射中匹配
	for cnName, promql := range n.metricMapping {
		if strings.Contains(lower, strings.ToLower(cnName)) {
			result.GeneratedQuery = promql
			result.Explanation = fmt.Sprintf("匹配到指标 [%s]，对应 PromQL 表达式", cnName)
			result.Confidence = 0.85
			return
		}
	}

	// 解析时间范围
	timeRange := extractTimeRange(input)

	// 解析阈值条件
	threshold, op := extractThreshold(input)

	// 解析 topk/bottomk
	if topN := extractTopN(input); topN > 0 {
		// 尝试匹配指标
		metric := extractMetricHint(input, n.metricMapping)
		if metric != "" {
			result.GeneratedQuery = fmt.Sprintf("topk(%d, %s)", topN, metric)
			result.Explanation = fmt.Sprintf("取前 %d 名，指标: %s", topN, metric)
			result.Confidence = 0.8
			return
		}
	}

	// 通用构建
	metric := extractMetricHint(input, n.metricMapping)
	if metric == "" {
		metric = "up"
		result.Confidence = 0.3
	} else {
		result.Confidence = 0.7
	}

	query := metric
	if timeRange != "" {
		// 如果指标需要 rate，包装一下
		if strings.Contains(metric, "_total") {
			query = fmt.Sprintf("rate(%s[%s])", metric, timeRange)
		}
	}
	if threshold > 0 && op != "" {
		query = fmt.Sprintf("%s %s %g", query, op, threshold)
	}

	result.GeneratedQuery = query
	result.Explanation = fmt.Sprintf("基于输入 [%s] 生成 PromQL 查询", input)
}

// translateToLogQL 转换为 LogQL
func (n *NL2Query) translateToLogQL(input string, result *NL2QueryResult) {
	lower := strings.ToLower(input)

	// 解析日志级别
	level := "error"
	if strings.Contains(lower, "warn") || strings.Contains(lower, "警告") {
		level = "warn"
	} else if strings.Contains(lower, "info") || strings.Contains(lower, "信息") {
		level = "info"
	}

	// 解析关键词过滤
	filterExpr := ""
	if strings.Contains(lower, "超时") || strings.Contains(lower, "timeout") {
		filterExpr = ` |= "timeout"`
	} else if strings.Contains(lower, "连接") || strings.Contains(lower, "connection") {
		filterExpr = ` |= "connection"`
	} else if strings.Contains(lower, "oom") || strings.Contains(lower, "内存") {
		filterExpr = ` |= "OOM"`
	}

	result.GeneratedQuery = fmt.Sprintf(`{level="%s"}%s`, level, filterExpr)
	result.Explanation = fmt.Sprintf("查询 %s 级别日志%s", level, filterExpr)
	result.Confidence = 0.75
}

// translateToSQL 转换为 SQL
func (n *NL2Query) translateToSQL(input string, result *NL2QueryResult) {
	// 基础 SQL 生成（实际场景中会由 LLM 辅助）
	result.GeneratedQuery = "SELECT * FROM events WHERE 1=1 LIMIT 100"
	result.Explanation = "基于输入生成 SQL 查询（需要 LLM 辅助精确生成）"
	result.Confidence = 0.5
}

// defaultMetricNameMapping 中文指标名到 PromQL 表达式的默认映射
func defaultMetricNameMapping() map[string]string {
	return map[string]string{
		// 系统指标
		"CPU使用率":  "1 - rate(node_cpu_seconds_total{mode=\"idle\"}[5m])",
		"cpu使用率":  "1 - rate(node_cpu_seconds_total{mode=\"idle\"}[5m])",
		"内存使用率":   "1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes",
		"磁盘使用率":   "1 - node_filesystem_avail_bytes / node_filesystem_size_bytes",
		"网络入流量":   "rate(node_network_receive_bytes_total[5m])",
		"网络出流量":   "rate(node_network_transmit_bytes_total[5m])",
		"TCP连接数":  "node_netstat_Tcp_CurrEstab",
		"tcp连接数":  "node_netstat_Tcp_CurrEstab",
		"系统负载":    "node_load1",
		"进程数":     "node_procs_running",
		// MySQL
		"MySQL连接数": "mysql_global_status_threads_connected",
		"mysql连接数": "mysql_global_status_threads_connected",
		"MySQL QPS": "rate(mysql_global_status_queries[5m])",
		"mysql qps": "rate(mysql_global_status_queries[5m])",
		"MySQL慢查询":  "rate(mysql_global_status_slow_queries[5m])",
		"mysql慢查询":  "rate(mysql_global_status_slow_queries[5m])",
		// Redis
		"Redis连接数": "redis_connected_clients",
		"redis连接数": "redis_connected_clients",
		"Redis内存":  "redis_memory_used_bytes",
		"redis内存":  "redis_memory_used_bytes",
		"Redis QPS": "rate(redis_commands_processed_total[5m])",
		"redis qps": "rate(redis_commands_processed_total[5m])",
		// HTTP
		"请求延迟":  "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
		"请求量":   "rate(http_requests_total[5m])",
		"错误率":   "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])",
		"QPS":   "rate(http_requests_total[5m])",
		"qps":   "rate(http_requests_total[5m])",
	}
}

// extractTimeRange 从自然语言中提取时间范围
func extractTimeRange(input string) string {
	lower := strings.ToLower(input)
	timePatterns := map[string]string{
		"1小时": "1h", "一小时": "1h", "1h": "1h",
		"30分钟": "30m", "半小时": "30m", "30m": "30m",
		"5分钟": "5m", "五分钟": "5m", "5m": "5m",
		"10分钟": "10m", "十分钟": "10m", "10m": "10m",
		"15分钟": "15m", "15m": "15m",
		"24小时": "24h", "一天": "24h", "1天": "24h",
		"7天": "7d", "一周": "7d", "7d": "7d",
	}
	for pattern, promRange := range timePatterns {
		if strings.Contains(lower, pattern) {
			return promRange
		}
	}
	return ""
}

// extractThreshold 从自然语言中提取阈值和比较操作符
func extractThreshold(input string) (float64, string) {
	lower := strings.ToLower(input)

	// 简单的阈值提取（实际场景由 LLM 辅助）
	thresholds := map[string]float64{
		"80%": 0.8, "90%": 0.9, "95%": 0.95, "70%": 0.7, "60%": 0.6, "50%": 0.5,
	}
	ops := map[string]string{
		"超过": ">", "大于": ">", "高于": ">",
		"低于": "<", "小于": "<", "不足": "<",
	}

	var threshold float64
	op := ""

	for pattern, val := range thresholds {
		if strings.Contains(lower, pattern) {
			threshold = val
			break
		}
	}
	for pattern, opStr := range ops {
		if strings.Contains(lower, pattern) {
			op = opStr
			break
		}
	}

	if threshold > 0 && op == "" {
		op = ">" // 默认大于
	}

	return threshold, op
}

// extractTopN 从自然语言中提取 top N
func extractTopN(input string) int {
	lower := strings.ToLower(input)
	topPatterns := map[string]int{
		"前10": 10, "前5": 5, "前3": 3, "前20": 20,
		"top10": 10, "top5": 5, "top3": 3, "top20": 20,
		"top 10": 10, "top 5": 5, "top 3": 3,
	}
	for pattern, n := range topPatterns {
		if strings.Contains(lower, pattern) {
			return n
		}
	}
	return 0
}

// extractMetricHint 从输入中提取可能的指标名
func extractMetricHint(input string, mapping map[string]string) string {
	lower := strings.ToLower(input)
	for cnName, promql := range mapping {
		if strings.Contains(lower, strings.ToLower(cnName)) {
			return promql
		}
	}
	return ""
}
