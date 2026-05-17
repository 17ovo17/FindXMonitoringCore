package aiengine

import "strings"

// ---------------------------------------------------------------------------
// PromptBuilder — 根据意图和上下文动态构建 system prompt
// ---------------------------------------------------------------------------
//
// 不同意图有不同的角色定义和输出格式要求。
// 参考 Hermes prompt_builder.py 的设计：模板化 + 动态组装。

// PromptTemplate 定义一个意图对应的 prompt 模板
type PromptTemplate struct {
	Role         string // 角色定义
	Constraints  string // 约束条件
	OutputFormat string // 输出格式要求
	Examples     string // 少量示例（可选）
}

// Build 将模板组装为完整的 system prompt 文本
func (t *PromptTemplate) Build() string {
	var sb strings.Builder
	sb.WriteString("## 角色\n")
	sb.WriteString(t.Role)
	sb.WriteString("\n\n## 约束\n")
	sb.WriteString(t.Constraints)
	sb.WriteString("\n\n## 输出格式\n")
	sb.WriteString(t.OutputFormat)
	if t.Examples != "" {
		sb.WriteString("\n\n## 示例\n")
		sb.WriteString(t.Examples)
	}
	return sb.String()
}

// PromptBuilder 管理所有意图的 prompt 模板
type PromptBuilder struct {
	templates map[IntentType]*PromptTemplate
}

// NewPromptBuilder 创建 PromptBuilder 并注册默认模板
func NewPromptBuilder() *PromptBuilder {
	pb := &PromptBuilder{
		templates: make(map[IntentType]*PromptTemplate),
	}
	pb.registerDefaults()
	return pb
}

// GetTemplate 获取指定意图的模板
func (pb *PromptBuilder) GetTemplate(intent IntentType) *PromptTemplate {
	if tmpl, ok := pb.templates[intent]; ok {
		return tmpl
	}
	// 兜底返回通用问答模板
	return pb.templates[IntentQuery]
}

// RegisterTemplate 注册或覆盖一个意图模板
func (pb *PromptBuilder) RegisterTemplate(intent IntentType, tmpl *PromptTemplate) {
	pb.templates[intent] = tmpl
}

// BuildForIntent 快捷方法：直接返回指定意图的完整 prompt 文本
func (pb *PromptBuilder) BuildForIntent(intent IntentType) string {
	return pb.GetTemplate(intent).Build()
}

// registerDefaults 注册所有默认模板
func (pb *PromptBuilder) registerDefaults() {
	pb.templates[IntentQuery] = &PromptTemplate{
		Role: "你是 FindX 智能运维助手，负责回答运维相关问题。" +
			"你拥有对 CMDB、监控、告警、日志等平台数据的访问能力。",
		Constraints: "1. 只使用平台数据回答，不编造信息。如果数据不足，明确告知用户。\n" +
			"2. 涉及敏感操作时提醒用户确认。\n" +
			"3. 回答需包含数据来源和时间范围。",
		OutputFormat: "简洁回答，关键数据用表格展示。不确定的信息标注置信度。",
	}

	pb.templates[IntentScript] = &PromptTemplate{
		Role: "你是运维脚本专家，根据用户需求生成安全可靠的运维脚本。" +
			"你熟悉 Bash、Python、Ansible 等运维工具链。",
		Constraints: "1. 脚本必须包含错误处理、日志输出、可回滚机制。\n" +
			"2. 禁止生成破坏性命令（rm -rf /、dd 等）。\n" +
			"3. 涉及生产环境操作必须有 dry-run 模式。\n" +
			"4. 密码/密钥使用变量占位符，不硬编码。",
		OutputFormat: "输出格式：\n1. 脚本代码（带注释）\n2. 使用说明\n3. 注意事项\n4. 回滚方法",
		Examples: "用户: 写一个清理 /tmp 下 7 天前文件的脚本\n" +
			"输出: [包含 find + -mtime +7 + 日志 + dry-run 的 bash 脚本]",
	}

	pb.templates[IntentTopology] = &PromptTemplate{
		Role: "你是业务拓扑分析专家，识别架构中的风险和优化点。" +
			"你可以访问 CMDB 关联关系、监控数据和告警状态。",
		Constraints: "1. 基于 CMDB 关联关系和监控数据分析，不猜测未知依赖。\n" +
			"2. 标注数据来源和分析时间窗口。\n" +
			"3. 风险评估需给出影响范围和概率。",
		OutputFormat: "输出格式：\n1. 拓扑概览（节点数、层级）\n2. 单点故障识别\n" +
			"3. 集群异常检测\n4. 性能瓶颈定位\n5. 监控覆盖短板\n6. 优化建议（按优先级排序）",
	}

	pb.templates[IntentSelfHeal] = &PromptTemplate{
		Role: "你是故障自愈引擎，负责从告警到恢复的全链路自动化。" +
			"你可以执行预定义的修复操作，但高风险操作需要人工确认。",
		Constraints: "1. 执行前必须评估影响范围（受影响服务数、用户数）。\n" +
			"2. 高风险操作（重启服务、切换主从、扩容）需要人工确认。\n" +
			"3. 每步操作必须有验证环节，失败立即停止。\n" +
			"4. 全程记录执行轨迹，支持审计和复盘。",
		OutputFormat: "输出格式：\n1. 告警摘要\n2. 根因分析（含置信度）\n" +
			"3. 恢复方案（含风险等级 1-5）\n4. 执行计划（步骤 + 预估耗时）\n5. 验证方法",
	}

	pb.templates[IntentDiagnose] = &PromptTemplate{
		Role: "你是故障诊断专家，通过多维数据关联定位根因。" +
			"你擅长从指标、日志、链路追踪中提取证据链。",
		Constraints: "1. 按证据链推理，不猜测。每个结论必须有数据支撑。\n" +
			"2. 区分「确认」和「疑似」结论，标注置信度。\n" +
			"3. 考虑时间相关性：告警时间 vs 指标异变时间 vs 变更时间。\n" +
			"4. 排除法：列出已排除的可能性及排除依据。",
		OutputFormat: "输出格式：\n1. 现象描述（时间线）\n2. 数据收集结果\n" +
			"3. 关联分析（时间/空间/因果）\n4. 根因定位（含置信度）\n5. 修复建议（按优先级）",
	}
}
