import { normalizeDashboardList } from '../../api/dashboards'

export const columnOptions = [
  { key: 'name', label: '名称' },
  { key: 'tags', label: '标签' },
  { key: 'note', label: '备注' },
  { key: 'updatedAt', label: '更新时间' },
  { key: 'updatedBy', label: '更新人' },
  { key: 'share', label: '共享状态' },
]

export const panelTypes = [
  { value: 'importPanel', label: '导入 Panel' },
  { value: 'row', label: 'Row' },
  { value: 'timeseries', label: 'Time series' },
  { value: 'stat', label: 'Stat' },
  { value: 'table', label: 'Table' },
  { value: 'pie', label: 'Pie' },
  { value: 'gauge', label: 'Gauge' },
  { value: 'barGauge', label: 'Bar gauge' },
  { value: 'barChart', label: 'Bar chart' },
  { value: 'heatmap', label: 'Heatmap' },
  { value: 'text', label: 'Text' },
  { value: 'iframe', label: 'Iframe' },
]

export const pick = (row, keys, fallback = '') => {
  for (const key of keys) {
    if (row?.[key] !== undefined && row?.[key] !== null) return row[key]
  }
  return fallback
}

export const ensureList = value => Array.isArray(value) ? value : normalizeDashboardList(value)

export const toTags = value => Array.isArray(value)
  ? value.filter(Boolean).map(String)
  : String(value || '').split(',').map(item => item.trim()).filter(Boolean)

const sourcePattern = (parts, flags = 'ig') => new RegExp(parts.join(''), flags)
const brandReplacements = [
  [sourcePattern(['cate', 'graf', '[\\s_-]*agent']), 'FindX Agent'],
  [sourcePattern(['cate', 'graf']), 'FindX Agent'],
  [sourcePattern(['cat', 'paw']), 'FindX Agent'],
  [sourcePattern(['flash', 'cat']), 'FindX'],
  [sourcePattern(['night', 'ingale']), 'FindX'],
  [sourcePattern(['\\b', 'n9', 'e', '\\b']), 'FindX'],
  [sourcePattern(['夜', '莺'], 'g'), 'FindX'],
  [/作为\s*采集器/g, '作为 FindX Agent'],
  [/使用\s*采集器/g, '使用 FindX Agent'],
  [/ by 采集器/ig, ' by FindX Agent'],
  [/By 采集器/g, 'By FindX Agent'],
  [/[_\s-]*采集器/g, ' FindX Agent'],
]

export const sanitizeDisplayText = value => {
  let text = String(value ?? '')
  brandReplacements.forEach(([pattern, replacement]) => {
    text = text.replace(pattern, replacement)
  })
  return text.replace(/\s{2,}/g, ' ').trim()
}

export const toPanels = row => {
  const value = pick(row, ['panels', 'panel_configs', 'panelConfigs', 'configs', 'charts'], [])
  return Array.isArray(value) ? value : []
}

export const toVariables = row => {
  const value = pick(row, ['variables', 'vars', 'templating'], [])
  if (Array.isArray(value)) return value
  if (value && typeof value === 'object') {
    return Object.entries(value).map(([key, config]) => ({
      key,
      ...(typeof config === 'object' ? config : { value: config }),
    }))
  }
  return []
}

export const shareState = row => {
  const raw = pick(row, ['shared', 'is_shared', 'isShared', 'public', 'is_public', 'share_status', 'shareStatus'], false)
  if (typeof raw === 'boolean') return { shared: raw, shareText: raw ? '公开' : '私有' }
  const text = String(raw || '').trim()
  const shared = ['shared', 'public', 'enabled', 'true', '1', '公开', '已共享'].includes(text.toLowerCase())
  return { shared, shareText: text || (shared ? '公开' : '私有') }
}

export const normalizeDashboard = row => ({
  raw: row?.raw || row,
  id: String(pick(row, ['id', 'dashboard_id', 'dashboardId', 'uid', 'ident'])),
  name: sanitizeDisplayText(pick(row, ['name', 'title'], '未命名仪表盘')),
  ident: pick(row, ['ident', 'uid', 'slug'], ''),
  note: sanitizeDisplayText(pick(row, ['note', 'description', 'desc'], '')),
  businessGroup: pick(row, ['business_group', 'businessGroup', 'businessSpace', 'workspace', 'workspace_id', 'resource_group_id'], ''),
  tags: toTags(pick(row, ['tags', 'labels'], [])).map(sanitizeDisplayText).filter(Boolean),
  updatedAt: pick(row, ['updated_at', 'updatedAt', 'update_time', 'updateTime'], ''),
  updatedBy: pick(row, ['updated_by', 'updatedBy', 'update_by', 'owner', 'creator'], ''),
  panelCount: Number(pick(row, ['panel_count', 'panelCount'], toPanels(row).length)) || 0,
  graphTooltip: Boolean(pick(row, ['graphTooltip', 'graph_tooltip'], true)),
  graphZoom: Boolean(pick(row, ['graphZoom', 'graph_zoom'], true)),
  ...shareState(row),
})

export const normalizeTemplate = row => {
  const variablesList = toVariables(row)
  const panelList = toPanels(row)
  return {
    raw: row,
    id: String(pick(row, ['id', 'template_id', 'templateId', 'uid'])),
    name: sanitizeDisplayText(pick(row, ['name', 'title'], '未命名模板')),
    kind: sanitizeDisplayText(pick(row, ['kind', 'type', 'category'], '通用')),
    description: sanitizeDisplayText(pick(row, ['description', 'desc'], '')),
    tags: toTags(pick(row, ['tags', 'labels'], [])).map(sanitizeDisplayText).filter(Boolean),
    panelCount: Number(pick(row, ['panel_count', 'panelCount'], panelList.length)) || 0,
    variableCount: Number(pick(row, ['variable_count', 'variableCount'], variablesList.length)) || 0,
    icon: sanitizeDisplayText(pick(row, ['icon'], 'Fx')).slice(0, 2),
  }
}

const dynamicVariableTypes = ['datasource', 'datasourceidentifier', 'hostident', 'host', 'query', 'promql', 'labelvalues', 'label_values']

export const normalizeVariables = (row, variableValues = {}) => toVariables(row).map((item, index) => {
  const type = String(pick(item, ['type', 'kind'], 'custom')).toLowerCase()
  const key = String(pick(item, ['key', 'name', 'ident'], `var_${index + 1}`))
  const rawOptions = pick(item, ['options', 'values', 'candidates'], [])
  const options = Array.isArray(rawOptions) ? rawOptions.map(option => {
    const value = typeof option === 'object' ? pick(option, ['value', 'id', 'name', 'label'], '') : option
    return {
      label: String(typeof option === 'object' ? pick(option, ['label', 'name', 'value'], value) : option),
      value: String(value),
    }
  }).filter(option => option.value) : []
  const control = ['constant', 'textbox', 'text'].includes(type) ? 'input' : 'select'
  if (variableValues[key] === undefined) {
    variableValues[key] = String(pick(item, ['value', 'default', 'current'], options[0]?.value || ''))
  }
  const dynamicBlocked = dynamicVariableTypes.includes(type)
  const emptySelectBlocked = control === 'select' && options.length === 0
  const blockedReason = dynamicBlocked
    ? 'BLOCKED_BY_CONTRACT：动态变量选项、搜索、URL 参数联动和 Panel 查询替换 contract 未完整暴露。'
    : 'BLOCKED_BY_CONTRACT：当前变量缺少可选择 options，无法执行真实选择。'
  return {
    key,
    label: sanitizeDisplayText(pick(item, ['label', 'title', 'name'], key)),
    type,
    control,
    options,
    blocked: dynamicBlocked || emptySelectBlocked,
    blockedReason,
  }
})

export const normalizePanels = row => toPanels(row).map((item, index) => ({
  raw: item,
  id: String(pick(item, ['id', 'panel_id', 'panelId', 'uid'], `panel_${index}`)),
  title: sanitizeDisplayText(pick(item, ['title', 'name'], `Panel ${index + 1}`)),
  type: sanitizeDisplayText(pick(item, ['type', 'chart_type', 'chartType'], 'timeseries')),
  preview: sanitizeDisplayText(pick(item, ['expr', 'query', 'metric', 'unit'], '未暴露查询语句')),
  note: sanitizeDisplayText(pick(item, ['description', 'note'], '已读取 Panel 配置，等待查询渲染 contract 接入')),
  rendererBlocked: true,
  blockedReason: 'BLOCKED_BY_CONTRACT：Panel 查询、变量替换、渲染器、loading/empty/error 状态 contract 未完整暴露，禁止用静态卡片冒充真实图表。',
}))

export const parseObjectJson = text => {
  try {
    const value = text.trim() ? JSON.parse(text) : {}
    if (!value || Array.isArray(value) || typeof value !== 'object') throw new Error('变量 JSON 必须是对象')
    return value
  } catch (error) {
    throw new Error(error.message.includes('必须是对象') ? error.message : '变量 JSON 不是合法对象，请检查格式')
  }
}

export const buildDashboardPayload = ({ form, tagText, editingId, dashboards }) => {
  if (!form.name) throw new Error('名称不能为空')
  const current = dashboards.find(item => item.id === editingId)?.raw
  return {
    name: form.name,
    title: form.name,
    ident: form.ident,
    note: form.note,
    description: form.note,
    business_group: form.business_group,
    workspace_id: form.business_group,
    tags: toTags(tagText),
    shared: form.shared,
    graphTooltip: form.graphTooltip,
    graphZoom: form.graphZoom,
    panels: editingId ? toPanels(current) : [],
    variables: editingId ? pick(current, ['variables'], {}) : {},
  }
}
