import { normalizeList } from '../../api/integrations'

const sourcePattern = (parts, flags = 'ig') => new RegExp(parts.join(''), flags)
const replacements = [
  [sourcePattern(['cate', 'graf', '[\\s_-]*agent']), 'FindX Agent'],
  [sourcePattern(['cate', 'graf']), 'FindX Agent'],
  [sourcePattern(['cat', 'paw']), 'FindX Agent'],
  [sourcePattern(['night', 'ingale']), 'FindX'],
  [sourcePattern(['\\b', 'n9', 'e', '\\b']), 'FindX'],
  [sourcePattern(['\\bNG\\b']), 'FindX'],
  [sourcePattern(['夜', '莺'], 'g'), 'FindX'],
  [sourcePattern(['作为\\s*', '采', '集', '器'], 'g'), '作为 FindX Agent'],
  [sourcePattern(['使用\\s*', '采', '集', '器'], 'g'), '使用 FindX Agent'],
  [sourcePattern(['[_\\s-]*', '采', '集', '器'], 'g'), ' FindX Agent'],
]

export const sanitizeDisplayText = value => {
  let text = String(value ?? '')
  replacements.forEach(([pattern, replacement]) => {
    text = text.replace(pattern, replacement)
  })
  return text
    .replace(sourcePattern(['使', '用', '\\s*FindX Agent\\s*作', '为', '\\s*FindX Agent'], 'g'), '使用 FindX Agent')
    .replace(sourcePattern(['FindX Agent\\s*作', '为', '\\s*FindX Agent'], 'g'), 'FindX Agent')
    .replace(/FindX Agent_version/g, 'FindX Agent version')
    .replace(/" FindX Agent"/g, '"FindX Agent"')
    .replace(/\s{2,}/g, ' ')
    .trim()
}

export const payloadTabs = [
  { key: 'instructions', label: '使用说明' },
  { key: 'collect', label: '采集模板' },
  { key: 'metric', label: '内置指标' },
  { key: 'dashboard', label: '仪表盘' },
  { key: 'alert', label: '告警规则' },
  { key: 'record', label: '记录规则' },
]

export const blockedContracts = {
  collect: 'BLOCKED_BY_CONTRACT：采集模板 payload 列表、分组选择、下发和回滚 contract 未完整暴露。',
  metric: 'BLOCKED_BY_CONTRACT：内置指标 payload 列表、PromQL 详情和指标浏览器联动 contract 未完整暴露。',
  alert: 'BLOCKED_BY_CONTRACT：告警规则模板分类、导入到业务组、数据源替换和批量导出 contract 未完整暴露。',
  record: 'BLOCKED_BY_CONTRACT：记录规则模板列表、导入、启停和预计算规则 contract 未完整暴露。',
}

export const toTags = value => Array.isArray(value)
  ? value.filter(Boolean).map(item => sanitizeDisplayText(item))
  : String(value || '').split(/[,\s]+/).map(item => sanitizeDisplayText(item)).filter(Boolean)

const pick = (row, keys, fallback = '') => {
  for (const key of keys) {
    if (row?.[key] !== undefined && row?.[key] !== null) return row[key]
  }
  return fallback
}

const componentKey = template => toTags(pick(template, ['tags', 'labels'], []))[0] || sanitizeDisplayText(pick(template, ['kind', 'category'], '通用'))

export const normalizeTemplatePayload = row => {
  const tags = toTags(pick(row, ['tags', 'labels'], []))
  return {
    raw: row,
    id: String(pick(row, ['id', 'template_id', 'templateId', 'uid'])),
    uuid: String(pick(row, ['uuid', 'id', 'uid'])),
    type: 'dashboard',
    name: sanitizeDisplayText(pick(row, ['name', 'title'], '未命名模板')),
    cate: sanitizeDisplayText(componentKey(row)),
    tags,
    note: sanitizeDisplayText(pick(row, ['description', 'desc'], '')),
    updatedBy: sanitizeDisplayText(pick(row, ['updated_by', 'updatedBy'], 'system')),
    content: JSON.stringify(row, null, 2),
  }
}

export const normalizeTemplates = value => normalizeList(value).map(normalizeTemplatePayload).filter(item => item.id)

export const buildComponents = templates => {
  const map = new Map()
  templates.forEach(template => {
    const key = template.cate || '通用'
    const current = map.get(key) || { ident: key, logo: key.slice(0, 2).toUpperCase(), readme: '', dashboardCount: 0, alertCount: 0, metricCount: 0, collectCount: 0, disabled: 0 }
    current.dashboardCount += 1
    current.readme = `${key} 组件包含 ${current.dashboardCount} 个仪表盘模板。告警规则、采集模板、内置指标和记录规则等待对应 contract 接入。`
    map.set(key, current)
  })
  return Array.from(map.values()).sort((a, b) => a.ident.localeCompare(b.ident))
}

export const filterPayloads = (payloads, query, componentIdent) => payloads.filter(item => {
  const words = [item.name, item.note, item.tags.join(' '), item.cate].join(' ').toLowerCase()
  const keywordHit = !query || words.includes(query.toLowerCase())
  const componentHit = !componentIdent || item.cate === componentIdent
  return keywordHit && componentHit
})

export const safeJson = value => {
  try {
    return JSON.stringify(value ?? null, null, 2)
  } catch {
    return String(value ?? '')
  }
}
