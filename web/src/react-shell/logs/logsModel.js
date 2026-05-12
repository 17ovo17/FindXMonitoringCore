export const sections = [
  { value: 'query', label: '日志检索', desc: '按时间、关键词、字段、服务、级别和 Trace 信息检索日志。' },
  { value: 'live', label: '实时日志', desc: '承载实时 tail、暂停、恢复、断线提示和查询条件复用。' },
  { value: 'fields', label: '字段筛选', desc: '管理可用字段、已选字段、字段搜索和字段值联想。' },
  { value: 'context', label: '上下文', desc: '围绕单条日志加载前后文、同实例和同 Trace 日志。' },
  { value: 'aggregate', label: '聚合分析', desc: '按字段、时间粒度和过滤条件聚合日志趋势。' },
  { value: 'pipelines', label: '接入管道', desc: '管理解析、过滤、增强、预览验证和版本审计。' },
  { value: 'saved-views', label: '保存视图', desc: '保存、更新、删除和按来源页面加载查询视图。' },
  { value: 'trace-link', label: 'Trace 关联', desc: '从日志 traceId/spanId 跳转到 FindX 链路详情，并保留主机 Agent 过滤上下文。' },
]

export const sectionSet = new Set(sections.map(item => item.value))

export const fieldGroups = [
  { title: '基础字段', fields: ['timestamp', 'severity_text', 'body', 'service_name', 'host_name'] },
  { title: '资源字段', fields: ['resource.service.name', 'resource.host.name', 'resource.container.name', 'resource.namespace'] },
  { title: 'Trace 字段', fields: ['trace_id', 'span_id', 'trace_flags', 'span_kind'] },
  { title: '自定义属性', fields: ['attributes.http.method', 'attributes.http.status_code', 'attributes.error', 'attributes.route'] },
]

export const pipelineSteps = [
  { id: 'parse', title: '解析', desc: '解析文本、JSON、正则和时间戳字段。' },
  { id: 'filter', title: '过滤', desc: '过滤低价值日志、敏感字段和噪声来源。' },
  { id: 'enrich', title: '增强', desc: '追加资源组、服务、环境、主机和链路字段。' },
  { id: 'route', title: '路由', desc: '按业务组、级别、服务和租户路由到目标索引。' },
  { id: 'preview', title: '预览验证', desc: '使用样本日志验证解析结果和错误位置。' },
]

export const savedViewColumns = ['名称', '来源', '查询条件', '字段', '更新时间', '操作']

const mojibakeTokens = ['\uFFFD', '\u951F', '\u00C3', '\u00C2', '\u6D93', '\u7EE0', '\u9477', '\u93BA', '\u59AF', '\u93CC', '\u935B', '\u95BE', '\u93C3', '\u701B', '\u7EF1', '\u7487', '\u935A', '\u6769', '\u6FA7', '\u68F0', '\u6DC7', '\u93BF', '\u9350']

export const displayText = (value, fallback = '-') => {
  if (value === null || value === undefined || value === '') return fallback
  const text = String(value)
  return mojibakeTokens.some(token => text.includes(token)) ? '内容不可读' : text
}
