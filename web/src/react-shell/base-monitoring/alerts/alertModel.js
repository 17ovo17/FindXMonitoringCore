import { redactText, safeJson } from '../../api/http.js'

export const BLOCKED_BY_CONTRACT = 'BLOCKED_BY_CONTRACT'

export const alertSections = [
  { value: 'rules', label: '规则' },
  { value: 'events', label: '事件' },
  { value: 'history-events', label: '历史事件' },
  { value: 'tracing-alarms', label: '链路告警' },
  { value: 'mutes', label: '屏蔽' },
  { value: 'subscriptions', label: '订阅' },
  { value: 'job', label: '告警自愈' },
  { value: 'event-pipelines', label: '事件流水线' },
]

export const severities = [
  { value: 'critical', label: 'S1 Critical' },
  { value: 'warning', label: 'S2 Warning' },
  { value: 'info', label: 'S3 Info' },
  { value: 'p0', label: 'P0' },
  { value: 'p1', label: 'P1' },
  { value: 'p2', label: 'P2' },
  { value: 'p3', label: 'P3' },
]

export const eventStatuses = [
  { value: 'firing', label: '触发中' },
  { value: 'acknowledged', label: '已确认' },
  { value: 'assigned', label: '已分派' },
  { value: 'muted', label: '已屏蔽' },
  { value: 'resolved', label: '已恢复' },
  { value: 'archived', label: '已归档' },
]

export const noDataPolicies = [
  { value: 'keep_state', label: '保持状态' },
  { value: 'alerting', label: '转为告警' },
  { value: 'ok', label: '转为正常' },
]

export const blockedContracts = {
  job: 'BLOCKED_BY_CONTRACT: 缺少告警自愈模板、任务列表、执行参数、审批、审计、回滚和执行记录契约。',
  'tracing-alarms': 'BLOCKED_BY_CONTRACT: 缺少链路告警规则、链路告警事件、服务/实例/端点关联、Trace 反查和复盘证据契约。',
  mutes: 'BLOCKED_BY_CONTRACT: 缺少屏蔽规则列表、时间窗口、对象选择、启停、克隆、删除与审计契约。',
  subscriptions: 'BLOCKED_BY_CONTRACT: 缺少订阅规则列表、接收对象、级别重定义、过滤条件、启停、克隆、删除与投递记录契约。',
  'event-pipelines': 'BLOCKED_BY_CONTRACT: 缺少事件流水线列表、触发方式、用途、启停、编辑、克隆、删除、执行记录与调试契约。',
  businessGroup: 'BLOCKED_BY_CONTRACT: 缺少业务组列表、我的业务组、跨组权限和按业务组过滤契约。',
  category: 'BLOCKED_BY_CONTRACT: 缺少规则分类/采集类型枚举及后端过滤契约。',
  columnSettings: 'BLOCKED_BY_CONTRACT: 缺少用户级列配置保存、恢复默认和服务端持久化契约。',
  batchRuleUpdate: 'BLOCKED_BY_CONTRACT: 缺少规则批量编辑、批量克隆到目标范围、批量删除、导入和导出服务端契约。',
  notificationConfig: 'BLOCKED_BY_CONTRACT: 缺少通知组、通知规则、通知渠道、模板和投递预览契约。',
  pipelineConfig: 'BLOCKED_BY_CONTRACT: 缺少事件流水线绑定、执行顺序、调试和执行记录契约。',
  effectiveTime: 'BLOCKED_BY_CONTRACT: 缺少生效时间、周期窗口、时区和节假日策略契约。',
  eventTimeRange: 'BLOCKED_BY_CONTRACT: 缺少事件服务端时间范围查询、自动刷新、时区和刷新间隔契约；当前只在已加载事件上做本地筛选。',
  eventAggregation: 'BLOCKED_BY_CONTRACT: 缺少服务端聚合规则、聚合卡片钻取和按聚合结果回填事件筛选契约；当前仅展示本地统计。',
  eventBatch: 'BLOCKED_BY_CONTRACT: 缺少事件批量删除、批量确认/反确认、批量屏蔽和批量分派契约。',
  eventMute: 'BLOCKED_BY_CONTRACT: 缺少从事件生成屏蔽规则的后端契约。',
  eventDelete: 'BLOCKED_BY_CONTRACT: 缺少事件删除契约；当前只接入确认、分派、恢复、归档。',
}

export const sanitizeDisplay = (value) => redactText(value ?? '').slice(0, 500)

export const displayDate = (value) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return sanitizeDisplay(value)
  return date.toLocaleString('zh-CN', { hour12: false })
}

export const displayJson = (value) => safeJson(value, 10000)

export const normalizeMap = (value) => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return {}
  return Object.fromEntries(Object.entries(value).map(([key, val]) => [String(key), sanitizeDisplay(val)]))
}

export const mapToPairs = (value) => Object.entries(normalizeMap(value)).map(([key, val]) => `${key}=${val}`)

export const normalizeRuleDetail = (value) => value?.rule || value?.data?.rule || value?.data || value

export const normalizeRule = (raw = {}) => {
  const enabled = raw.enabled ?? raw.enable ?? raw.disabled === 0
  const severity = raw.severity || raw.severities?.[0] || 'info'
  return {
    raw,
    id: String(raw.id ?? ''),
    name: sanitizeDisplay(raw.name || raw.title || '-'),
    query: sanitizeDisplay(raw.query || raw.expr || raw.prom_ql || ''),
    severity: String(severity),
    datasourceId: sanitizeDisplay(raw.datasource_id || raw.datasourceId || raw.datasource_ids?.[0] || '-'),
    businessGroup: sanitizeDisplay(raw.group_name || raw.businessGroup || raw.business_group || raw.group_id || '-'),
    category: sanitizeDisplay(raw.cate || raw.category || raw.prod || '-'),
    targetSelector: normalizeMap(raw.target_selector || raw.targetSelector),
    labels: normalizeMap(raw.labels || raw.append_tags),
    annotations: normalizeMap(raw.annotations),
    enabled: Boolean(enabled),
    status: raw.status || (enabled ? 'active' : 'disabled'),
    version: Number(raw.version || 1),
    forDuration: sanitizeDisplay(raw.for_duration || raw.forDuration || ''),
    noDataPolicy: raw.no_data_policy || raw.noDataPolicy || 'keep_state',
    createdAt: raw.created_at || raw.createdAt,
    updatedAt: raw.updated_at || raw.updatedAt || raw.update_at,
    updatedBy: sanitizeDisplay(raw.updated_by || raw.updatedBy || raw.update_by || raw.create_by || raw.created_by || '-'),
  }
}

export const normalizeEvent = (raw = {}) => ({
  raw,
  id: String(raw.id ?? ''),
  ruleId: String(raw.rule_id ?? raw.ruleId ?? ''),
  name: sanitizeDisplay(raw.name || raw.rule_name || '-'),
  severity: String(raw.severity || 'info'),
  status: raw.status || 'firing',
  datasourceId: sanitizeDisplay(raw.datasource_id || raw.datasourceId || '-'),
  businessGroup: sanitizeDisplay(raw.group_name || raw.businessGroup || raw.business_group || raw.group_id || '-'),
  category: sanitizeDisplay(raw.cate || raw.category || raw.rule_prod || raw.prod || '-'),
  target: sanitizeDisplay(raw.target_ident || raw.target_id || raw.target || '-'),
  value: sanitizeDisplay(raw.value || ''),
  fingerprint: sanitizeDisplay(raw.fingerprint || ''),
  count: Number(raw.count || 0),
  firstSeen: raw.first_seen || raw.firstSeen || raw.first_trigger_time || raw.created_at,
  lastSeen: raw.last_seen || raw.lastSeen || raw.trigger_time || raw.updated_at,
  labels: normalizeMap(raw.labels || raw.tags),
  annotations: normalizeMap(raw.annotations),
  actionLog: Array.isArray(raw.action_log) ? raw.action_log : [],
  ackBy: sanitizeDisplay(raw.ack_by || raw.ackBy || raw.claimant || ''),
  assignee: sanitizeDisplay(raw.assignee || raw.claimant || ''),
  resolution: sanitizeDisplay(raw.resolution || ''),
})

export const parseJsonMap = (text, field) => {
  if (!String(text || '').trim()) return {}
  const value = JSON.parse(text)
  if (!value || Array.isArray(value) || typeof value !== 'object') throw new Error(`${field}必须是 JSON 对象`)
  return Object.fromEntries(Object.entries(value).map(([key, val]) => [String(key), String(val ?? '')]))
}

export const rulePayload = (draft) => ({
  id: draft.id || undefined,
  name: draft.name.trim(),
  datasource_id: draft.datasourceId.trim(),
  query: draft.query.trim(),
  severity: draft.severity,
  for_duration: draft.forDuration.trim(),
  no_data_policy: draft.noDataPolicy,
  enabled: draft.enabled,
  status: draft.enabled ? 'active' : 'disabled',
  target_selector: parseJsonMap(draft.targetSelector, '目标选择器'),
  labels: parseJsonMap(draft.labels, '标签'),
  annotations: parseJsonMap(draft.annotations, '注解'),
  effective_time: draft.effective_time || undefined,
  notify_config: draft.notify_config || undefined,
  pipeline_config: draft.pipeline_config || undefined,
  triggers_config: draft.triggers_config || undefined,
})

export const filterText = (parts, query) => {
  if (!query) return true
  return parts.join(' ').toLowerCase().includes(query.toLowerCase())
}

export const severityLabel = (value) => severities.find((item) => item.value === value)?.label || sanitizeDisplay(value || '-')

export const statusLabel = (value) => eventStatuses.find((item) => item.value === value)?.label || sanitizeDisplay(value || '-')

export const makeError = (error, fallback = '请求失败') => {
  if (error?.status === 401) return '登录已过期，请重新登录。'
  if (error?.status === 403) return '没有当前操作权限。'
  if (error?.status >= 500) return `HTTP ${error.status}: 服务暂时不可用。`
  return sanitizeDisplay(error?.message || fallback)
}

export const getUniqueOptions = (rows, key) => Array.from(
  new Set(rows.map((row) => row[key]).filter((value) => value && value !== '-').map(String)),
).sort((a, b) => a.localeCompare(b, 'zh-CN'))

export const formatDuration = (start, end = Date.now()) => {
  const startDate = new Date(start)
  const endDate = end ? new Date(end) : new Date()
  if (Number.isNaN(startDate.getTime())) return '-'
  const safeEnd = Number.isNaN(endDate.getTime()) ? Date.now() : endDate.getTime()
  const diff = Math.max(0, safeEnd - startDate.getTime())
  const minutes = Math.floor(diff / 60000)
  if (minutes < 60) return `${minutes || 1} 分钟`
  const hours = Math.floor(minutes / 60)
  if (hours < 48) return `${hours} 小时 ${minutes % 60} 分钟`
  const days = Math.floor(hours / 24)
  return `${days} 天 ${hours % 24} 小时`
}
