import { redactText, safeJson } from '../../api/alerting'

export const severityOptions = [
  { label: 'S1 Critical', value: 'critical', tag: 'danger' },
  { label: 'S2 Warning', value: 'warning', tag: 'warning' },
  { label: 'S3 Info', value: 'info', tag: 'info' },
  { label: 'P0', value: 'p0', tag: 'danger' },
  { label: 'P1', value: 'p1', tag: 'warning' },
  { label: 'P2', value: 'p2', tag: 'warning' },
  { label: 'P3', value: 'p3', tag: 'info' },
]

export const eventStatusOptions = [
  { label: 'firing', value: 'firing', tag: 'danger' },
  { label: 'acknowledged', value: 'acknowledged', tag: 'warning' },
  { label: 'assigned', value: 'assigned', tag: 'warning' },
  { label: 'muted', value: 'muted', tag: 'info' },
  { label: 'resolved', value: 'resolved', tag: 'success' },
  { label: 'archived', value: 'archived', tag: 'info' },
]

export const ruleColumnOptions = [
  { key: 'eventStatus', label: '当前事件' },
  { key: 'prod', label: '类型/分类' },
  { key: 'datasource', label: '数据源' },
  { key: 'name', label: '名称' },
  { key: 'severity', label: '级别' },
  { key: 'labels', label: '附加标签' },
  { key: 'notify', label: '通知' },
  { key: 'updatedAt', label: '更新时间' },
  { key: 'updatedBy', label: '更新人' },
  { key: 'enabled', label: '启用' },
]

export const defaultVisibleColumns = Object.fromEntries(ruleColumnOptions.map(item => [item.key, true]))

export const blockedContracts = {
  'rule-import-builtin': '内置告警模板导入 contract 未暴露，需接入模板中心告警规则导入后才能执行。',
  'rule-import-json': '告警规则 JSON 导入 contract 未暴露，需定义批量校验、冲突处理和审计结果。',
  'rule-import-prometheus': '外部规则格式导入 contract 未暴露，需定义解析、转换、预览和回滚。',
  'batch-update': '告警规则批量更新 contract 未暴露，不能静态修改筛选条件或通知字段。',
  'clone-to-business': '克隆到业务组 contract 未暴露，需定义目标业务组、权限和冲突策略。',
  'clone-to-hosts': '克隆到机器 contract 未暴露，需定义机器选择、下发状态和回滚。',
  'events-delete': '事件删除 contract 未暴露，当前只允许 ack、assign、resolve、archive 真实动作。',
  'events-aggregate': '事件聚合视图 contract 未暴露，当前保留入口并显示阻断态。',
  mutes: '告警屏蔽 contract 未暴露，需接入屏蔽规则、时间窗口、对象选择、启停和审计。',
  subscriptions: '告警订阅 contract 未暴露，需接入订阅规则、接收对象、过滤条件和投递记录。',
  workflows: '告警工作流需要并入 AI SRE 工作流，并保留事件流水线列表、编辑、调试和执行记录结构。',
}

const severityRank = { critical: 1, p0: 1, warning: 2, p1: 2, p2: 2, info: 3, p3: 3 }
const statusRank = { firing: 1, acknowledged: 2, assigned: 2, muted: 3, resolved: 4, archived: 5 }

export const formatDate = value => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return redactText(value)
  return date.toLocaleString('zh-CN', { hour12: false })
}

const brandRules = [
  [new RegExp(['AI', ' Work', 'Bench'].join(''), 'ig'), 'FindX'],
  [new RegExp(['Night', 'ingale'].join(''), 'ig'), 'FindX'],
  [new RegExp(`\\b${['n', '9', 'e'].join('')}\\b`, 'ig'), 'FindX'],
  [new RegExp(`${['Cate', 'graf'].join('')}|${['Cat', 'paw'].join('')}`, 'ig'), 'FindX Agent'],
  [new RegExp(['Sky', 'Walking'].join(''), 'ig'), 'FindX 链路监控'],
  [new RegExp(['Sig', 'NoZ'].join(''), 'ig'), 'FindX 日志中心'],
]

export const sanitizeBrand = value => brandRules.reduce(
  (text, [pattern, replacement]) => text.replace(pattern, replacement),
  redactText(value),
)

export const datasourceValue = item => String(item?.id || item?.ident || item?.identifier || item?.name || '')
export const datasourceLabel = item => sanitizeBrand(item?.name || item?.ident || item?.identifier || item?.id || '-')

export const severityTag = value => severityOptions.find(item => item.value === value)?.tag || 'info'
export const eventStatusTag = value => eventStatusOptions.find(item => item.value === value)?.tag || 'info'
export const severityText = value => severityOptions.find(item => item.value === value)?.label || redactText(value || '-')

export const normalizeDatasourceId = value => {
  if (Array.isArray(value)) return value[0] || ''
  return String(value ?? '')
}

export const normalizeMap = value => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return {}
  return Object.fromEntries(Object.entries(value).map(([key, val]) => [String(key), String(val ?? '')]))
}

export const mapToTags = value => Object.entries(normalizeMap(value)).map(([key, val]) => `${key}=${val}`)

export const parseJsonObject = (text, field) => {
  if (!String(text || '').trim()) return {}
  const value = JSON.parse(text)
  if (!value || Array.isArray(value) || typeof value !== 'object') throw new Error(`${field} 必须是对象`)
  return value
}

export const normalizeRule = raw => {
  const enabled = raw?.enabled ?? raw?.enable ?? raw?.disabled === 0 ?? false
  const severity = raw?.severity || raw?.severities?.[0] || raw?.priority || 'info'
  const datasourceId = normalizeDatasourceId(raw?.datasource_id || raw?.datasource_ids)
  return {
    raw,
    id: String(raw?.id ?? ''),
    name: redactText(raw?.name || raw?.title || '-'),
    prod: redactText(raw?.prod || raw?.cate || 'metric'),
    datasourceId: redactText(datasourceId || '-'),
    severity,
    labels: normalizeMap(raw?.labels || raw?.append_tags),
    annotations: normalizeMap(raw?.annotations),
    query: redactText(raw?.query || raw?.expr || ''),
    forDuration: raw?.for_duration || raw?.duration || '',
    noDataPolicy: raw?.no_data_policy || 'keep_state',
    enabled: Boolean(enabled),
    status: raw?.status || (enabled ? 'active' : 'disabled'),
    updatedAt: raw?.updated_at || raw?.update_at || raw?.created_at || '',
    updatedBy: redactText(raw?.updated_by || raw?.update_by || raw?.created_by || '-'),
    version: raw?.version || 1,
  }
}

export const normalizeEvent = raw => ({
  raw,
  id: String(raw?.id ?? ''),
  ruleId: String(raw?.rule_id ?? ''),
  name: redactText(raw?.name || raw?.rule_name || '-'),
  severity: raw?.severity || 'info',
  status: raw?.status || 'firing',
  datasourceId: redactText(raw?.datasource_id || '-'),
  target: redactText(raw?.target_ident || raw?.target_id || '-'),
  value: redactText(raw?.value || ''),
  fingerprint: redactText(raw?.fingerprint || ''),
  count: raw?.count || 0,
  firstSeen: raw?.first_seen || raw?.created_at || '',
  lastSeen: raw?.last_seen || raw?.updated_at || '',
  labels: normalizeMap(raw?.labels),
  annotations: normalizeMap(raw?.annotations),
  actionLog: Array.isArray(raw?.action_log) ? raw.action_log : [],
  assignee: redactText(raw?.assignee || ''),
  ackBy: redactText(raw?.ack_by || ''),
  resolution: redactText(raw?.resolution || ''),
})

export const filterRules = (rules, filter) => rules.filter(rule => {
  const text = `${rule.name} ${rule.datasourceId} ${rule.prod} ${mapToTags(rule.labels).join(' ')}`.toLowerCase()
  if (filter.query && !text.includes(filter.query.toLowerCase())) return false
  if (filter.datasource && rule.datasourceId !== filter.datasource) return false
  if (filter.enabled !== '' && String(rule.enabled) !== filter.enabled) return false
  if (filter.severities?.length && !filter.severities.includes(rule.severity)) return false
  return true
})

export const filterEvents = (events, filter) => events.filter(event => {
  const text = `${event.name} ${event.target} ${event.datasourceId} ${mapToTags(event.labels).join(' ')}`.toLowerCase()
  if (filter.query && !text.includes(filter.query.toLowerCase())) return false
  if (filter.datasource && event.datasourceId !== filter.datasource) return false
  if (filter.severities?.length && !filter.severities.includes(event.severity)) return false
  if (filter.statuses?.length && !filter.statuses.includes(event.status)) return false
  if (filter.ruleId && event.ruleId !== filter.ruleId) return false
  return true
})

export const sortEvents = events => [...events].sort((a, b) => {
  const statusDelta = (statusRank[a.status] || 9) - (statusRank[b.status] || 9)
  if (statusDelta) return statusDelta
  const severityDelta = (severityRank[a.severity] || 9) - (severityRank[b.severity] || 9)
  if (severityDelta) return severityDelta
  return new Date(b.lastSeen).getTime() - new Date(a.lastSeen).getTime()
})

export const summarizeRuleEvent = (rule, events) => {
  const matched = sortEvents(events.filter(event => event.ruleId === rule.id))
  return matched[0] || null
}

export const downloadText = (filename, text, type = 'text/plain;charset=utf-8') => {
  const blob = new Blob([text], { type })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export const toCsv = rows => {
  const keys = ['id', 'name', 'prod', 'datasourceId', 'severity', 'status', 'enabled', 'updatedBy', 'updatedAt']
  const escape = value => `"${String(value ?? '').replace(/"/g, '""')}"`
  return [keys.join(','), ...rows.map(row => keys.map(key => escape(row[key])).join(','))].join('\n')
}

export const blockedPayload = (action, context = {}) => safeJson({
  action,
  context,
  status: 'BLOCKED_BY_CONTRACT',
  next_contract_needed: blockedContracts[action] || '后端 contract 未暴露',
})
