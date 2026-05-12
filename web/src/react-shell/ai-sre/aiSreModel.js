export const sections = [
  { value: 'diagnosis', label: '诊断会话', desc: '会话、推理链、证据来源和只读建议动作。' },
  { value: 'workflow', label: '工作流', desc: '工作流列表、DSL、运行记录和执行契约边界。' },
  { value: 'health', label: '健康检查', desc: '数据源、模型、存储和数据到达健康视图。' },
  { value: 'report', label: '复盘报告', desc: '巡检任务、报告详情、进度和导出边界。' },
  { value: 'evidence', label: 'Evidence Chain', desc: '只展示本页真实获取的数据证据，缺失即标记数据缺失。' },
  { value: 'knowledge', label: '知识库', desc: '知识搜索、诊断案例和向量索引状态边界。' },
  { value: 'remediation', label: '自动修复', desc: '处置计划、审批、执行、回滚和审计契约边界。' },
]

export const sectionSet = new Set(sections.map(item => item.value))

export const evidenceCategories = [
  { key: 'diagnosis', label: '诊断会话' },
  { key: 'metrics', label: '指标证据' },
  { key: 'logs', label: '日志证据' },
  { key: 'trace', label: '链路证据' },
  { key: 'alerts', label: '告警证据' },
  { key: 'cmdb', label: 'CMDB 证据' },
  { key: 'agent', label: 'Agent 证据' },
  { key: 'inspection', label: '巡检证据' },
  { key: 'workflow', label: '工作流证据' },
  { key: 'knowledge', label: '知识库证据' },
]

export const displayText = value => String(value ?? '')
  .replace(/https?:\/\/[^\s"'<>]+/ig, '<URL>')
  .replace(/(bearer\s+)[^"',\s}]+/ig, '$1<TOKEN>')
  .replace(/((?:token|cookie|authorization|api[_-]?key|secret|dsn|password)\s*[:=]\s*)[^;\n"',}]+/ig, '$1<SECRET>')

export const compactJson = value => {
  try {
    return displayText(JSON.stringify(value ?? {}, null, 2))
  } catch {
    return displayText(value)
  }
}
