import { redactText, safeJson } from '../../api/http.js'

export const BLOCKED_BY_CONTRACT = 'BLOCKED_BY_CONTRACT'
export const HISTORY_LIMIT = 100

const brandPattern = (parts) => new RegExp(parts.join(''), 'gi')

const HIDDEN_BRANDS = [
  [brandPattern(['\\bAI Work', 'Bench\\b']), 'FindX'],
  [brandPattern(['\\bAI', 'Ops\\b']), 'FindX AI SRE'],
  [brandPattern(['Night', 'ingale']), 'FindX'],
  [brandPattern(['\\b', 'n', '9', 'e', '\\b']), 'FindX'],
  [brandPattern(['Sky', 'Walking']), 'FindX'],
  [brandPattern(['Sig', 'NoZ']), 'FindX'],
  [brandPattern(['Cate', 'graf']), 'FindX Agent'],
  [brandPattern(['Cat', 'paw']), 'FindX Agent'],
  [brandPattern(['Flash', 'cat']), 'FindX'],
  [brandPattern(['Flash', 'Duty']), 'FindX'],
]

export const displayText = (value) => HIDDEN_BRANDS.reduce(
  (text, [pattern, replacement]) => text.replace(pattern, replacement),
  redactText(value ?? ''),
)

export const displayJson = (value, limit = 10000) => displayText(safeJson(value, limit))

export const blockedContracts = {
  datasource: `${BLOCKED_BY_CONTRACT}: 指标数据源契约未暴露，当前不能伪造查询上下文。`,
  builtInMetrics: `${BLOCKED_BY_CONTRACT}: 内置指标需要指标分类、搜索、单位、输入联动和权限契约。`,
  saveView: `${BLOCKED_BY_CONTRACT}: 保存视图需要新增、更新、权限和审计契约。`,
  shareChart: `${BLOCKED_BY_CONTRACT}: 分享需要临时图表载荷、访问控制和过期策略契约。`,
  graphSettings: `${BLOCKED_BY_CONTRACT}: 图表设置需要面板持久化契约。`,
  aiQuery: `${BLOCKED_BY_CONTRACT}: AI 查询生成需要模型配置、权限和审计契约。`,
  recordingRules: `${BLOCKED_BY_CONTRACT}: 记录规则需要列表、编辑、启停、克隆、删除和权限契约。`,
  objectViews: `${BLOCKED_BY_CONTRACT}: 对象快捷视图需要对象范围和权限契约。`,
}

const promqlTypeFields = ['type', 'plugin_type', 'plugin_type_name', 'datasource_type', 'category', 'name']
const promqlPattern = /(^|[^a-z0-9])(prometheus|promql|victoriametrics|victoria[-_\s]?metrics|vmselect)([^a-z0-9]|$)/i
const nonPromqlPattern = /(^|[^a-z0-9])(mysql|mariadb|postgres|postgresql|clickhouse|elasticsearch|influxdb|loki|tempo|jaeger)([^a-z0-9]|$)/i

const fieldText = (row, field) => {
  const value = row?.[field]
  return value === undefined || value === null ? '' : String(value)
}

const pureInternalIdPattern = /^(\d{10,}|[0-9a-f]{24,}|[0-9a-f]{8}-[0-9a-f-]{27,})$/i
const timestampSuffixPattern = /(?:\s*[/_-]\s*)?\d{10,}$/

const isInternalIdText = (value) => pureInternalIdPattern.test(String(value || '').trim())

const readableDatasourcePart = (value) => {
  const text = displayText(value || '').trim()
  if (!text || isInternalIdText(text)) return ''
  return text.replace(timestampSuffixPattern, '').trim()
}

export const isPromqlDatasource = (row) => {
  const typeText = promqlTypeFields.map((field) => fieldText(row, field)).filter(Boolean).join(' ')
  const hasPromql = promqlPattern.test(typeText)
  const hasNonPromql = nonPromqlPattern.test(typeText)
  return hasPromql && !hasNonPromql
}

export const datasourceId = (row) => row?.id ?? row?.datasource_id ?? row?.value ?? ''
export const datasourceType = (row) => displayText(row?.type ?? row?.plugin_type ?? row?.datasource_type ?? row?.category ?? 'PromQL')
export const datasourceName = (row) => (
  readableDatasourcePart(row?.name)
  || readableDatasourcePart(row?.label)
  || readableDatasourcePart(row?.title)
  || readableDatasourcePart(row?.alias)
  || '新数据源'
)

export const metricName = (metric) => displayText(
  metric?.promql || metric?.raw_name || metric?.standard_name || metric?.name || metric?.metric || '',
)

export const metricDescription = (metric) => displayText(
  metric?.description || metric?.standard_name || metric?.help || metric?.unit || '',
)

export const metricUnit = (metric) => displayText(metric?.unit || metric?.unit_name || '')

export const nowLocalInput = (offsetMs = 0) => {
  const date = new Date(Date.now() + offsetMs)
  const pad = (value) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

export const localInputToUnix = (value, fallback = Date.now()) => {
  const parsed = value ? new Date(value).getTime() : fallback
  return Math.floor((Number.isFinite(parsed) ? parsed : fallback) / 1000)
}

export const relativeRanges = [
  { label: '近 5 分钟', value: '5m', ms: 5 * 60 * 1000 },
  { label: '近 15 分钟', value: '15m', ms: 15 * 60 * 1000 },
  { label: '近 30 分钟', value: '30m', ms: 30 * 60 * 1000 },
  { label: '近 1 小时', value: '1h', ms: 60 * 60 * 1000 },
  { label: '近 6 小时', value: '6h', ms: 6 * 60 * 60 * 1000 },
  { label: '近 24 小时', value: '24h', ms: 24 * 60 * 60 * 1000 },
  { label: '近 7 天', value: '7d', ms: 7 * 24 * 60 * 60 * 1000 },
]

export const resolveRelativeRange = (value) => {
  const range = relativeRanges.find((r) => r.value === value)
  if (!range) return { start: Math.floor((Date.now() - 3600000) / 1000), end: Math.floor(Date.now() / 1000) }
  const end = Math.floor(Date.now() / 1000)
  const start = Math.floor((Date.now() - range.ms) / 1000)
  return { start, end }
}

export const createPanel = () => ({
  id: `panel-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
  query: 'up',
  activeTab: 'table',
  instantTime: nowLocalInput(),
  rangeStart: nowLocalInput(-60 * 60 * 1000),
  rangeEnd: nowLocalInput(),
  relativeRange: '1h',
  unit: '',
  graphMode: 'line',
  maxDataPoints: 300,
  minStep: 0,
  step: 15,
  showLegend: true,
  autocomplete: true,
  collapsed: false,
  loading: false,
  error: '',
  notice: '',
  instantResult: null,
  rangeResult: null,
  suggestions: [],
  metrics: [],
  metricSearch: '',
  historySearch: '',
  showHistory: false,
  showMetrics: false,
  showGraphSettings: false,
  showAiInput: false,
  aiPrompt: '',
  aiLoading: false,
})

export const UNITS = [
  { value: '', label: 'none' },
  { value: 'bytes', label: 'bytes' },
  { value: '%', label: 'percent' },
  { value: 's', label: 'seconds' },
  { value: 'ms', label: 'ms' },
]

export const VIEW_STORAGE_KEY = 'findx-metric-explorer-views'

export const readViews = () => {
  try {
    const raw = JSON.parse(localStorage.getItem(VIEW_STORAGE_KEY) || '[]')
    return Array.isArray(raw) ? raw : []
  } catch { return [] }
}

export const saveView = (name, datasource, panels) => {
  const snapshot = panels.map((p) => ({
    query: p.query, relativeRange: p.relativeRange, unit: p.unit,
    maxDataPoints: p.maxDataPoints, minStep: p.minStep, step: p.step,
  }))
  const views = readViews().filter((v) => v.name !== name)
  views.unshift({ name, datasource, panels: snapshot, time: Date.now() })
  localStorage.setItem(VIEW_STORAGE_KEY, JSON.stringify(views.slice(0, 20)))
  return views
}

export const deleteView = (name) => {
  const views = readViews().filter((v) => v.name !== name)
  localStorage.setItem(VIEW_STORAGE_KEY, JSON.stringify(views))
  return views
}

export const historyKey = (datasource) => `findx-query-promql-history-${datasource || 'default'}`

export const readHistory = (key) => {
  try {
    const raw = JSON.parse(localStorage.getItem(key) || '[]')
    if (Array.isArray(raw)) {
      return raw
        .map((item) => (Array.isArray(item) ? item[0] : item))
        .filter(Boolean)
        .map(displayText)
        .slice(0, HISTORY_LIMIT)
    }
  } catch {
    return []
  }
  return []
}

export const saveHistory = (key, query) => {
  const value = String(query || '').trim()
  if (!value) return []
  const next = [value, ...readHistory(key).filter((item) => item !== value)].slice(0, HISTORY_LIMIT)
  localStorage.setItem(key, JSON.stringify(next))
  return next
}

export const resultRows = (result, unit = '') => {
  const data = result?.data?.result ?? result?.result ?? []
  if (!Array.isArray(data)) return []
  return data.map((row, index) => {
    const value = Array.isArray(row?.value) ? row.value : []
    const metric = row?.metric && Object.keys(row.metric).length ? displayJson(row.metric, 1200) : `series_${index + 1}`
    return {
      metric,
      value: `${displayText(value[1] ?? '')}${unit ? ` ${unit}` : ''}`,
      time: value[0] ? new Date(Number(value[0]) * 1000).toLocaleString() : '-',
    }
  })
}

export const rangeSeries = (result) => {
  const data = result?.data?.result ?? result?.result ?? []
  if (!Array.isArray(data)) return []
  return data.map((row, index) => ({
    name: row?.metric && Object.keys(row.metric).length ? displayJson(row.metric, 800) : `series_${index + 1}`,
    points: Array.isArray(row?.values) ? row.values.length : 0,
  }))
}

export const csvEscape = (value) => `"${String(value ?? '').replaceAll('"', '""')}"`
