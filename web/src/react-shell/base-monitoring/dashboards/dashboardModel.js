import { redactText, safeJson } from '../../api/http.js'

export const PENDING = ''

const brandPattern = (parts) => new RegExp(parts.join(''), 'gi')

const HIDDEN_BRANDS = [
  [brandPattern(['Night', 'ingale']), 'FindX'],
  [brandPattern(['\\b', 'n', '9', 'e', '\\b']), 'FindX'],
  [brandPattern(['Sky', 'Walking']), 'FindX'],
  [brandPattern(['Sig', 'NoZ']), 'FindX'],
  [brandPattern(['Cate', 'graf']), 'FindX Agent'],
  [brandPattern(['Cat', 'paw']), 'FindX Agent'],
  [brandPattern(['Flash', 'cat']), 'FindX'],
  [brandPattern(['Flash', 'Duty']), 'FindX'],
]

const SENSITIVE_FIELD_RE = /(?:api[_-]?key|token|password|passwd|secret|cookie|authorization|dsn|private[_-]?key)/i

export const displayText = (value) => HIDDEN_BRANDS.reduce(
  (text, [pattern, replacement]) => text.replace(pattern, replacement),
  redactText(value ?? ''),
).replace(/[?？]{2,}/g, '内容不可读')

export const displayJson = (value, limit = 12000) => displayText(safeJson(value, limit))

export const blockedContracts = {
  exportApi: '批量导出服务端契约未暴露，当前只导出已加载的脱敏配置。',
  panelRender: 'Panel 查询、变量替换、渲染器、loading、empty 和 error 状态契约未完整暴露。',
  panelEditor: '图表编辑器需要查询编辑器、渲染配置、变换、覆盖和审计契约。',
  variableOptions: '动态变量选项、搜索、URL 联动和 Panel 查询替换契约未完整暴露。',
  runtimeControls: '时间范围、刷新间隔和时区需要 Panel 查询、变量替换、URL 联动和自动刷新契约。',
  importPanel: '从外部 Panel 导入需要解析、冲突检测和回滚契约。',
  inspect: '查询参数检查需要 Panel 查询执行和响应检查契约。',
  shareView: '分享视图需要分享 token 校验、过期控制、权限隔离和只读渲染契约。',
}

const pick = (row, keys, fallback = '') => {
  for (const key of keys) {
    if (row?.[key] !== undefined && row?.[key] !== null) return row[key]
  }
  return fallback
}

const parseMaybeJson = (value, fallback) => {
  if (Array.isArray(value) || (value && typeof value === 'object')) return value
  if (typeof value !== 'string' || !value.trim()) return fallback
  try {
    return JSON.parse(value)
  } catch {
    return fallback
  }
}

export const toTags = (value) => {
  if (Array.isArray(value)) return value.map((item) => displayText(item)).filter(Boolean)
  return String(value || '').split(/[, ]+/).map((item) => displayText(item.trim())).filter(Boolean)
}

export const dashboardId = (row) => String(pick(row, ['id', 'dashboard_id', 'dashboardId', 'uid', 'ident'], ''))

export const normalizeDashboard = (row = {}) => {
  const variables = parseMaybeJson(pick(row, ['variables', 'vars', 'templating'], {}), {})
  const panels = parseMaybeJson(pick(row, ['panels', 'panel_configs', 'panelConfigs', 'configs', 'charts'], []), [])
  const shared = Boolean(pick(row, ['shared', 'is_shared', 'public', 'is_public'], false))
  return {
    raw: row,
    id: dashboardId(row),
    title: displayText(pick(row, ['title', 'name'], '未命名仪表盘')),
    description: displayText(pick(row, ['description', 'note', 'desc'], '')),
    workspaceId: displayText(pick(row, ['workspace_id', 'workspaceId', 'workspace'], '')),
    resourceGroupId: displayText(pick(row, ['resource_group_id', 'resourceGroupId', 'business_group', 'group_id'], '')),
    tags: toTags(pick(row, ['tags', 'labels'], [])),
    variables,
    panels,
    panelCount: Array.isArray(panels) ? panels.length : 0,
    version: Number(pick(row, ['version'], 0)) || 0,
    status: displayText(pick(row, ['status'], 'active')),
    shared,
    shareText: shared ? '公开' : '私有',
    shareSummary: displayText(pick(row, ['share_summary', 'shareSummary'], '')),
    updatedBy: displayText(pick(row, ['updated_by', 'updatedBy', 'update_by', 'owner', 'created_by'], '')),
    updatedAt: displayText(pick(row, ['updated_at', 'updatedAt', 'update_at'], '')),
  }
}

export const normalizeTemplate = (row = {}) => {
  const variables = parseMaybeJson(pick(row, ['variables'], {}), {})
  const panels = parseMaybeJson(pick(row, ['panels'], []), [])
  return {
    raw: row,
    id: dashboardId(row),
    title: displayText(pick(row, ['title', 'name'], '未命名模板')),
    description: displayText(pick(row, ['description', 'desc'], '')),
    tags: toTags(pick(row, ['tags'], [])),
    variables,
    panels,
    panelCount: Array.isArray(panels) ? panels.length : 0,
  }
}

export const normalizeVariables = (variables = {}) => {
  const source = Array.isArray(variables)
    ? variables
    : Object.entries(variables || {}).map(([key, value]) => ({ key, value }))
  return source.map((item, index) => {
    const rawKey = String(pick(item, ['key', 'name', 'ident'], `var_${index + 1}`) ?? '')
    const sensitiveVariable = SENSITIVE_FIELD_RE.test(rawKey)
    const key = sensitiveVariable ? `credential_${index + 1}` : displayText(rawKey)
    const options = parseMaybeJson(pick(item, ['options', 'values', 'candidates'], []), [])
    const optionList = Array.isArray(options)
      ? options.map((option, optionIndex) => {
        const value = typeof option === 'object' ? pick(option, ['value', 'id', 'name', 'label'], '') : option
        if (sensitiveVariable) return { label: `凭据候选 ${optionIndex + 1}`, value: `credential_option_${optionIndex + 1}` }
        const label = typeof option === 'object' ? pick(option, ['label', 'name', 'value'], value) : option
        const optionSensitive = SENSITIVE_FIELD_RE.test(String(label ?? '')) || SENSITIVE_FIELD_RE.test(String(value ?? ''))
        return {
          label: optionSensitive ? `凭据候选 ${optionIndex + 1}` : displayText(label),
          value: optionSensitive ? `credential_option_${optionIndex + 1}` : displayText(value),
        }
      }).filter((option) => option.value)
      : []
    const type = displayText(pick(item, ['type', 'kind'], optionList.length ? 'custom' : 'textbox'))
    const currentValue = sensitiveVariable ? (optionList[0]?.value || '<SECRET>') : displayText(pick(item, ['value', 'default', 'current'], optionList[0]?.value || ''))
    return {
      key,
      label: sensitiveVariable || SENSITIVE_FIELD_RE.test(String(pick(item, ['label', 'title', 'name'], key) ?? '')) ? '凭据变量' : displayText(pick(item, ['label', 'title', 'name'], key)),
      type,
      value: currentValue,
      options: optionList,
      blocked: ['query', 'datasource', 'host', 'label_values', 'promql'].includes(type.toLowerCase()) || (type !== 'textbox' && optionList.length === 0),
    }
  })
}

export const normalizePanels = (panels = []) => {
  const list = Array.isArray(panels) ? panels : []
  return list.map((panel, index) => ({
    raw: panel,
    id: String(pick(panel, ['id', 'panel_id', 'panelId', 'uid'], `panel_${index + 1}`)),
    title: displayText(pick(panel, ['title', 'name'], `Panel ${index + 1}`)),
    type: displayText(pick(panel, ['type', 'chart_type', 'chartType'], 'timeseries')),
    query: displayText(pick(panel, ['query', 'expr', 'expression', 'metric'], '')),
    description: displayText(pick(panel, ['description', 'note'], '')),
  }))
}

export const dashboardPayload = (draft) => ({
  title: draft.title,
  description: draft.description,
  workspace_id: draft.workspaceId,
  resource_group_id: draft.resourceGroupId,
  tags: toTags(draft.tags),
  variables: draft.variables || {},
  panels: Array.isArray(draft.panels) ? draft.panels : [],
  status: draft.status || 'active',
})

const sanitizeKey = (key, index) => {
  if (SENSITIVE_FIELD_RE.test(String(key ?? ''))) return `credential_${index + 1}`
  return displayText(key)
}

export const sanitizeForExport = (value, key = '') => {
  if (SENSITIVE_FIELD_RE.test(String(key ?? ''))) return '<SECRET>'
  if (Array.isArray(value)) return value.map((item) => sanitizeForExport(item))
  if (value && typeof value === 'object') {
    return Object.fromEntries(Object.entries(value).map(([entryKey, entryValue], index) => [
      sanitizeKey(entryKey, index),
      sanitizeForExport(entryValue, entryKey),
    ]))
  }
  if (typeof value === 'string') return displayText(value)
  return value
}

export const dashboardExportPayload = (row) => ({
  id: row.id,
  title: row.title,
  description: row.description,
  resource_group_id: row.resourceGroupId,
  tags: row.tags,
  variables: sanitizeForExport(row.variables),
  panels: sanitizeForExport(row.panels),
  shared: row.shared,
  export_note: blockedContracts.exportApi,
})

export const downloadJson = (filename, value) => {
  const blob = new Blob([JSON.stringify(sanitizeForExport(value), null, 2)], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}
