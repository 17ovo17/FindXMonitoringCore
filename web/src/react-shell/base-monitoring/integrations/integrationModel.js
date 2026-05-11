import { normalizeList, redactText } from '../../api/integrations.js'

const sourcePattern = (parts, flags = 'ig') => new RegExp(parts.join(''), flags)

const SECRET_KEY_RE = /password|passwd|token|secret|cookie|authorization|api[_-]?key|access[_-]?key|private[_-]?key|dsn|(^|[_-])(user(name)?|login[_-]?user|account)([_-]|$)/i
const DSN_CREDENTIAL_RE = /([a-z][a-z0-9+.-]*:\/\/)([^/@\s:]+):([^/@\s]+)@/ig
const TCP_DSN_CREDENTIAL_RE = /(^|[\s"'])([^:\s"'/@]+):([^@\s"']+)@tcp\(/ig
const TEXT_SECRET_KEY_RE = /password|passwd|token|secret|cookie|authorization|api[_-]?key|apikey|access[_-]?key|private[_-]?key|privatekey|dsn|user(name)?|login[_-]?user|account/i
const TEXT_SECRET_ASSIGNMENT_RE = /(["']?)(password|passwd|token|secret|cookie|authorization|api[_-]?key|apikey|access[_-]?key|private[_-]?key|privatekey|dsn|user(?:name)?|login[_-]?user|account)\1(\s*[:=]\s*)(["']?)([^"',\s}\]]+)\4/ig
const TEXT_AUTH_HEADER_RE = /\b(authorization|cookie)\s*:\s*([^\n\r,}]+)/ig
const TEXT_AUTH_SCHEME_RE = /\b(bearer|basic)\s+[a-z0-9._~+/=-]+/ig

const displayReplacements = [
  [sourcePattern(['cate', 'graf', '[\\s_-]*agent']), 'FindX Agent'],
  [sourcePattern(['cate', 'graf']), 'FindX Agent'],
  [sourcePattern(['cat', 'paw']), 'FindX Agent'],
  [sourcePattern(['night', 'ingale']), 'FindX'],
  [sourcePattern(['\\b', 'n9', 'e', '\\b']), 'FindX'],
  [sourcePattern(['flash', 'cat']), 'FindX'],
  [sourcePattern(['flash', 'duty']), 'FindX'],
  [sourcePattern(['sky', 'walking']), 'FindX'],
  [sourcePattern(['sig', 'no', 'z']), 'FindX'],
]

export const sanitizeDisplayText = (value) => {
  let text = String(value ?? '')
  displayReplacements.forEach(([pattern, replacement]) => {
    text = text.replace(pattern, replacement)
  })
  return text
    .replace(DSN_CREDENTIAL_RE, '$1<LOGIN_USER>:<TOKEN>@')
    .replace(TCP_DSN_CREDENTIAL_RE, '$1<LOGIN_USER>:<TOKEN>@tcp(')
    .replace(/\s{2,}/g, ' ')
    .trim()
}

const placeholderForKey = (key) => {
  const lowered = String(key).toLowerCase()
  if (lowered.includes('api') || lowered.includes('access') || lowered.includes('private')) return '<API_KEY>'
  if (lowered.includes('dsn')) return '<DB_DSN>'
  if (lowered.includes('user') || lowered.includes('account')) return '<LOGIN_USER>'
  return '<TOKEN>'
}

const redactObject = (value, key = '') => {
  if (value === null || value === undefined) return value
  if (SECRET_KEY_RE.test(key)) return placeholderForKey(key)
  if (Array.isArray(value)) return value.map((item) => redactObject(item))
  if (typeof value === 'object') {
    return Object.fromEntries(
      Object.entries(value).map(([entryKey, entryValue]) => [entryKey, redactObject(entryValue, entryKey)]),
    )
  }
  if (typeof value === 'string') return sanitizeDisplayText(value)
  return value
}

const redactSensitiveText = (value) => sanitizeDisplayText(value)
  .replace(TEXT_AUTH_HEADER_RE, (match, key) => `${key}: ${placeholderForKey(key)}`)
  .replace(TEXT_AUTH_SCHEME_RE, (match, scheme) => `${scheme} <TOKEN>`)
  .replace(TEXT_SECRET_ASSIGNMENT_RE, (match, keyQuote, key, separator, valueQuote) => (
    TEXT_SECRET_KEY_RE.test(key)
      ? `${keyQuote}${key}${keyQuote}${separator}${valueQuote}${placeholderForKey(key)}${valueQuote}`
      : match
  ))

export const safeJson = (value, limit = 12000) => {
  try {
    const text = sanitizeDisplayText(JSON.stringify(redactObject(value), null, 2))
    return text.length > limit ? `${text.slice(0, limit)}\n...内容已截断` : text
  } catch {
    return sanitizeDisplayText(String(value ?? ''))
  }
}

export const payloadTabs = [
  { key: 'instructions', label: '使用说明' },
  { key: 'collect', label: '采集模板' },
  { key: 'metric', label: '内置指标' },
  { key: 'dashboard', label: '仪表盘模板' },
  { key: 'alert', label: '告警规则' },
  { key: 'record', label: '记录规则' },
]

export const blockedContracts = {
  components: 'BLOCKED_BY_CONTRACT：组件实体列表契约未暴露，当前仅展示仪表盘模板降级视图，不能判定组件实体完成。',
  systems: '系统集成新增、编辑、删除、排序和菜单显示状态已接入 FindX 后端写契约；嵌入打开和动态菜单嵌入仍保持 BLOCKED_BY_CONTRACT。',
  componentCreate: '组件新增已接入 FindX 后端写契约；保存后会重新读取列表确认。',
  componentEdit: '组件编辑、图标、说明和停用状态已接入 FindX 后端写契约；内置静态行由后端 409 保护。',
  componentDelete: '组件删除已接入 FindX 后端写契约；内置静态行或仍有关联 payload 的组件由后端 409 保护。',
  collect: 'BLOCKED_BY_CONTRACT：采集模板列表、分类、下发和回滚契约未暴露。',
  metric: 'BLOCKED_BY_CONTRACT：内置指标列表、指标详情和查询联动契约未暴露。',
  alert: '告警规则 payload 写入已接入 FindX 后端契约；导入、数据源替换和批量导出业务落地仍未暴露。',
  record: 'BLOCKED_BY_CONTRACT：记录规则模板列表、导入、启停和预计算契约未暴露。',
  payloadCreate: 'dashboard、collect、alert payload 新增已接入 FindX 后端写契约；metric、firemap、record 仍保持 BLOCKED_BY_CONTRACT。',
  payloadEdit: 'dashboard、collect、alert payload 编辑已接入 FindX 后端写契约；内置静态行由后端 409 保护。',
  payloadDelete: 'payload 删除已接入 FindX 后端写契约；内置静态行由后端 409 保护。',
  payloadImport: 'BLOCKED_BY_CONTRACT：payload 导入到业务组契约未暴露，当前不打开导入表单，也不调用非同源导入接口。',
  payloadContent: 'BLOCKED_BY_CONTRACT：当前记录缺少可用 payload.content，不能预览、导出或导入。',
  systemCreate: '系统集成新增已接入 FindX 后端写契约。',
  systemEdit: '系统集成编辑已接入 FindX 后端写契约。',
  systemDelete: '系统集成删除已接入 FindX 后端写契约；内置项删除由后端保护。',
  systemSort: '系统集成排序权重写入已接入 FindX 后端写契约。',
  systemMenu: '系统集成菜单显示状态已接入 FindX 后端 hide 契约；动态菜单嵌入仍阻断。',
  systemOpen: 'BLOCKED_BY_CONTRACT：系统集成嵌入打开契约未开放，当前不打开 iframe、WebView 或外部窗口。',
}

export const writablePayloadTypes = new Set(['dashboard', 'collect', 'alert'])

const baseLogoMap = {
  prometheus: '/image/logos/prometheus.png',
  elasticsearch: '/image/logos/elasticsearch.png',
  tdengine: '/image/logos/tdengine.png',
  loki: '/image/logos/loki.png',
  jaeger: '/image/logos/jaeger.png',
  ck: '/image/logos/ck.png',
  opensearch: '/image/logos/opensearch.png',
  doris: '/image/logos/doris.png',
  mysql: '/image/logos/mysql.png',
  pgsql: '/image/logos/pgsql.png',
  postgresql: '/image/logos/pgsql.png',
  postgres: '/image/logos/pgsql.png',
  linux: '/image/logos/host.png',
  windows: '/image/logos/host.png',
  host: '/image/logos/host.png',
  hosts: '/image/logos/host.png',
  kubernetes: '/image/logos/host.png',
  k8s: '/image/logos/host.png',
  oracle: '/image/logos/oracle.png',
  sqlserver: '/image/logos/sqlserver.png',
  influxdb: '/image/logos/influxdb.png',
  zabbix: '/image/logos/zabbix.png',
  victorialogs: '/image/logos/victorialogs.png',
}

const pick = (row, keys, fallback = '') => {
  for (const key of keys) {
    if (row?.[key] !== undefined && row?.[key] !== null) return row[key]
  }
  return fallback
}

const normalizeRows = (value) => {
  if (Array.isArray(value)) return value
  if (Array.isArray(value?.dat)) return value.dat
  if (Array.isArray(value?.result)) return value.result
  const rows = normalizeList(value)
  if (rows.length) return rows
  if (value && typeof value === 'object') return [value]
  return []
}

const toTimestampMs = (value) => {
  const number = Number(value)
  if (!Number.isFinite(number) || number <= 0) return 0
  return number > 100000000000 ? number : number * 1000
}

const formatTime = (value) => {
  const timestamp = toTimestampMs(value)
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  if (Number.isNaN(date.getTime())) return '-'
  const pad = (item) => String(item).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

const previewValue = (row) => {
  const config = pick(row, ['config', 'settings', 'options', 'payload', 'extra'], '')
  const url = pick(row, ['url', 'href', 'link', 'endpoint', 'embed_url', 'embedUrl'], '')
  if (config && typeof config === 'object') return safeJson(config, 2000)
  if (typeof config === 'string' && config.trim()) return sanitizeDisplayText(config).slice(0, 2000)
  return sanitizeDisplayText(url).slice(0, 2000)
}

const extractContent = (row) => {
  const content = pick(row, ['content', 'payload', 'dashboard', 'template_content', 'templateContent'], '')
  if (typeof content === 'string' && content.trim()) return content
  if (content && typeof content === 'object') return safeJson(content, 500000)
  if (row?.variables || row?.panels) {
    return safeJson({
      id: pick(row, ['id', 'template_id', 'templateId', 'uid']),
      title: pick(row, ['name', 'title']),
      description: pick(row, ['description', 'desc', 'note']),
      tags: pick(row, ['tags', 'labels'], []),
      variables: row.variables || {},
      panels: row.panels || [],
    }, 500000)
  }
  return ''
}

export const toTags = (value) => {
  if (Array.isArray(value)) return value.filter(Boolean).map(sanitizeDisplayText)
  return String(value || '').split(/[,\\s]+/).map(sanitizeDisplayText).filter(Boolean)
}

export const logoForIdent = (ident) => {
  const key = String(ident || '').toLowerCase()
  return baseLogoMap[key] || `/image/logos/${encodeURIComponent(key)}.png`
}

export const initialLetters = (value) => sanitizeDisplayText(value).slice(0, 2).toUpperCase() || 'FX'

export const normalizeComponents = (value) => normalizeRows(value)
  .map((row) => {
    const id = String(pick(row, ['id', 'component_id', 'componentId', 'ident']))
    const ident = sanitizeDisplayText(pick(row, ['ident', 'name', 'title'], id || '通用组件'))
    const updatedBy = sanitizeDisplayText(pick(row, ['updated_by', 'updatedBy', 'create_by', 'created_by', 'createdBy'], 'system'))
    return {
      raw: row,
      id,
      ident,
      key: ident.toLowerCase(),
      logo: sanitizeDisplayText(pick(row, ['logo', 'icon'], logoForIdent(ident))),
      readme: sanitizeDisplayText(pick(row, ['readme', 'description', 'desc', 'note'], '')),
      disabled: Number(pick(row, ['disabled', 'is_disabled'], 0)) === 1,
      dashboardCount: Number(pick(row, ['dashboard_count', 'dashboardCount'], 0)) || 0,
      alertCount: Number(pick(row, ['alert_count', 'alertCount'], 0)) || 0,
      metricCount: Number(pick(row, ['metric_count', 'metricCount'], 0)) || 0,
      collectCount: Number(pick(row, ['collect_count', 'collectCount'], 0)) || 0,
      recordCount: Number(pick(row, ['record_count', 'recordCount'], 0)) || 0,
      updatedBy,
      protected: isBuiltinCatalogActor(updatedBy),
      contractSource: 'components',
    }
  })
  .filter((item) => item.id && item.id !== 'undefined' && item.ident)

export const normalizePayloads = (value, type = 'dashboard') => normalizeRows(value)
  .map((row) => {
    const id = String(pick(row, ['id', 'payload_id', 'payloadId', 'uuid']))
    const content = extractContent(row)
    const payloadType = sanitizeDisplayText(pick(row, ['type'], type))
    const cate = sanitizeDisplayText(pick(row, ['cate', 'category', 'kind'], ''))
    const tags = toTags(pick(row, ['tags', 'labels'], []))
    const mergedTags = cate && !tags.includes(cate) ? [cate, ...tags] : tags
    const updatedBy = sanitizeDisplayText(pick(row, ['updated_by', 'updatedBy', 'create_by', 'created_by', 'createdBy'], 'system'))
    return {
      raw: row,
      id,
      uuid: String(pick(row, ['uuid', 'uid', 'id'], id)),
      type: payloadType,
      componentId: String(pick(row, ['component_id', 'componentId'], '')),
      name: sanitizeDisplayText(pick(row, ['name', 'title'], '未命名 payload')),
      cate,
      tags: mergedTags,
      note: sanitizeDisplayText(pick(row, ['description', 'desc', 'note'], '')),
      updatedBy,
      content,
      missingContent: !content,
      fallbackOnly: Boolean(row?.fallbackOnly),
      protected: isBuiltinCatalogActor(updatedBy),
    }
  })
  .filter((item) => item.id && item.id !== 'undefined')

const isBuiltinCatalogActor = (value) => {
  const actor = String(value || '').trim().toLowerCase()
  return actor === '' || actor === 'system' || actor === 'findx-builtin-catalog'
}

export const normalizeDashboardTemplateFallback = (value) => normalizeRows(value)
  .map((row) => {
    const tags = toTags(pick(row, ['tags', 'labels'], []))
    const id = String(pick(row, ['id', 'template_id', 'templateId', 'uid']))
    const cate = sanitizeDisplayText(tags[0] || pick(row, ['kind', 'category', 'cate'], '通用组件'))
    const content = extractContent(row)
    return {
      raw: row,
      id,
      uuid: String(pick(row, ['uuid', 'uid', 'id'], id)),
      type: 'dashboard',
      componentId: '',
      name: sanitizeDisplayText(pick(row, ['name', 'title'], '未命名仪表盘模板')),
      cate,
      tags,
      note: sanitizeDisplayText(pick(row, ['description', 'desc', 'note'], '')),
      updatedBy: sanitizeDisplayText(pick(row, ['updated_by', 'updatedBy'], 'system')),
      content,
      missingContent: !content,
      fallbackOnly: true,
      protected: true,
    }
  })
  .filter((item) => item.id && item.id !== 'undefined')

export const blankComponentDraft = (row) => ({
  id: row?.id || '',
  ident: row?.ident || '',
  name: row?.raw?.name || row?.ident || '',
  logo: row?.logo || (row?.ident ? logoForIdent(row.ident) : '/image/logos/host.png'),
  readme: row?.readme || '',
  disabled: Boolean(row?.disabled),
})

export const parseComponentDraft = (draft, editing = false) => {
  const id = String(draft?.id || '').trim()
  const ident = String(draft?.ident || '').trim()
  const name = String(draft?.name || ident).trim()
  const logo = String(draft?.logo || '').trim()
  const readme = String(draft?.readme || '').trim()
  if (!editing && id && !/^[A-Za-z0-9_.:-]{1,64}$/.test(id)) return { error: 'ID 只能包含字母、数字、点、下划线、冒号或短横线，最长 64 位。' }
  if (!/^[A-Za-z0-9_.-]{1,128}$/.test(ident)) return { error: '组件标识必填，只能包含字母、数字、点、下划线或短横线，最长 128 位。' }
  if (!name || [...name].length > 255) return { error: '组件名称必填且不能超过 255 个字符。' }
  if (!isFindXRoute(logo)) return { error: 'Logo 必须是以 / 开头的 FindX 同源相对路径。' }
  if ([...readme].length > 20000) return { error: '使用说明不能超过 20000 个字符。' }
  if (hasSensitiveText(`${id} ${ident} ${name} ${logo} ${readme}`)) return { error: '组件字段不能包含凭据、Token、Cookie、DSN 或外部品牌来源内容。' }
  return {
    payload: {
      ...(editing || id ? { id } : {}),
      ident,
      name,
      logo,
      readme,
      disabled: draft?.disabled ? 1 : 0,
    },
  }
}

export const blankPayloadDraft = (row, component, type) => ({
  id: row?.id || '',
  uuid: row?.uuid || row?.id || '',
  type: row?.type || type || 'dashboard',
  componentId: row?.componentId || component?.id || '',
  cate: row?.cate || component?.ident || '',
  name: row?.name || '',
  content: row?.content || defaultPayloadContent(type),
})

export const parsePayloadDraft = (draft, editing = false) => {
  const id = String(draft?.id || '').trim()
  const uuid = String(draft?.uuid || id || buildPayloadId()).trim()
  const type = String(draft?.type || '').trim()
  const componentId = String(draft?.componentId || '').trim()
  const cate = String(draft?.cate || '').trim()
  const name = String(draft?.name || '').trim()
  const contentText = String(draft?.content || '').trim()
  if (!writablePayloadTypes.has(type)) return { error: `payload 类型 ${type || '-'} 暂未开放写入，保持 BLOCKED_BY_CONTRACT。` }
  if (!editing && id && !/^[A-Za-z0-9_.:-]{1,64}$/.test(id)) return { error: 'ID 只能包含字母、数字、点、下划线、冒号或短横线，最长 64 位。' }
  if (!/^[A-Za-z0-9_.:-]{1,64}$/.test(uuid)) return { error: 'UUID 只能包含字母、数字、点、下划线、冒号或短横线，最长 64 位。' }
  if (!componentId) return { error: 'component_id 缺失，不能保存 payload。' }
  if (!name || [...name].length > 255) return { error: 'payload 名称必填且不能超过 255 个字符。' }
  if ([...cate].length > 128) return { error: '分类不能超过 128 个字符。' }
  if (!contentText) return { error: 'content 必填。' }
  if (contentText.length > 200000) return { error: 'content 不能超过 200000 字符。' }
  if (hasSensitiveText(`${type} ${componentId} ${cate} ${name} ${contentText}`)) return { error: 'payload 不能包含密码、Token、Cookie、Authorization、DSN、私钥或完整账号字段。' }
  const content = parsePayloadContent(type, contentText)
  if (content.error) return content
  return {
    payload: {
      ...(editing || id ? { id } : {}),
      uuid,
      type,
      component_id: componentId,
      cate,
      name,
      content: content.value,
    },
  }
}

const defaultPayloadContent = (type) => {
  if (type === 'dashboard') return '{\n  "template": "",\n  "variables": {},\n  "panels": []\n}'
  if (type === 'alert') return '{\n  "expr": "",\n  "severity": "warning"\n}'
  return ''
}

const buildPayloadId = () => `fx-${Date.now()}`

const parsePayloadContent = (type, contentText) => {
  if (type === 'dashboard' || type === 'alert') {
    try {
      const value = JSON.parse(contentText)
      if (!value || (typeof value !== 'object' && !Array.isArray(value))) {
        return { error: 'dashboard/alert content 必须是 JSON 对象或数组。' }
      }
      if (hasSensitiveObject(value)) return { error: 'content JSON 中不能包含敏感字段或凭据内容。' }
      return { value }
    } catch {
      return { error: 'dashboard/alert content 必须是合法 JSON。' }
    }
  }
  return { value: contentText }
}

const testPattern = (pattern, value) => {
  pattern.lastIndex = 0
  return pattern.test(value)
}

const hasSensitiveText = (value) => (
  testPattern(TEXT_SECRET_KEY_RE, value)
  || testPattern(TEXT_AUTH_HEADER_RE, value)
  || testPattern(TEXT_AUTH_SCHEME_RE, value)
  || testPattern(DSN_CREDENTIAL_RE, value)
  || testPattern(TCP_DSN_CREDENTIAL_RE, value)
  || /nightingale|n9e|categraf|catpaw|skywalking|signoz|flashcat|findx\.local|mysql:\/\/|postgres:\/\//i.test(value)
)

const hasSensitiveObject = (value, key = '') => {
  if (SECRET_KEY_RE.test(key)) return true
  if (Array.isArray(value)) return value.some((item) => hasSensitiveObject(item))
  if (value && typeof value === 'object') {
    return Object.entries(value).some(([entryKey, entryValue]) => hasSensitiveObject(entryValue, entryKey))
  }
  return typeof value === 'string' && hasSensitiveText(value)
}

export const isWritablePayloadType = (type) => writablePayloadTypes.has(String(type || '').trim())

export const normalizeSystemIntegrations = (value) => normalizeRows(value)
  .map((row) => {
    const id = String(pick(row, ['id', 'product_id', 'productId', 'ident', 'uid']))
    const hide = Boolean(pick(row, ['hide', 'hidden'], false))
    const showInMenu = Boolean(pick(row, ['show_in_menu', 'showInMenu'], !hide))
    const isPrivate = Boolean(pick(row, ['is_private', 'isPrivate', 'private'], false))
    const teamIds = pick(row, ['team_ids', 'teamIds', 'teams'], [])
    const normalizedTeamIds = Array.isArray(teamIds) ? teamIds.map((item) => Number(item)).filter((item) => Number.isInteger(item) && item > 0) : toPositiveTeamIds(teamIds)
    return {
      raw: row,
      id,
      name: sanitizeDisplayText(pick(row, ['name', 'title', 'ident'], '未命名系统集成')),
      url: sanitizeDisplayText(pick(row, ['url', 'href', 'link', 'endpoint', 'embed_url', 'embedUrl'], '')),
      configPreview: previewValue(row),
      isPrivate,
      teamIds: normalizedTeamIds,
      hide,
      showInMenu,
      weight: Number(pick(row, ['weight', 'sort', 'order'], 0)) || 0,
      builtin: Boolean(pick(row, ['builtin', 'built_in', 'is_builtin'], false)),
      status: sanitizeDisplayText(pick(row, ['status'], 'active')),
      capabilities: row?.capabilities || {},
      blockedActions: Array.isArray(row?.blocked_actions) ? row.blocked_actions : [],
      createBy: sanitizeDisplayText(pick(row, ['create_by', 'createBy', 'created_by', 'createdBy'], '')),
      updateBy: sanitizeDisplayText(pick(row, ['update_by', 'updateBy', 'updated_by', 'updatedBy'], '')),
      createAt: formatTime(pick(row, ['create_at', 'createAt', 'created_at', 'createdAt'], 0)),
      updateAt: formatTime(pick(row, ['update_at', 'updateAt', 'updated_at', 'updatedAt'], 0)),
    }
  })
  .filter((item) => item.id && item.id !== 'undefined')
  .sort((a, b) => a.weight - b.weight || a.name.localeCompare(b.name))

export const toPositiveTeamIds = (value) => {
  const source = Array.isArray(value) ? value : String(value || '').split(/[,，\s]+/)
  const ids = []
  const seen = new Set()
  for (const item of source) {
    const id = Number(String(item).trim())
    if (!Number.isInteger(id) || id <= 0 || seen.has(id)) continue
    ids.push(id)
    seen.add(id)
  }
  return ids
}

export const blankSystemIntegrationDraft = (row, fallbackWeight = 0) => ({
  id: row?.id || '',
  name: row?.name || '',
  url: row?.url || '/',
  configPreview: row?.configPreview || row?.url || '/',
  isPrivate: Boolean(row?.isPrivate),
  teamIdsText: (row?.teamIds || []).join(', '),
  weight: Number(row?.weight ?? fallbackWeight) || 0,
  hide: Boolean(row?.hide),
})

export const parseSystemIntegrationDraft = (draft, editing = false) => {
  const name = String(draft?.name || '').trim()
  const url = String(draft?.url || '').trim()
  const configPreview = String(draft?.configPreview || '').trim()
  const id = String(draft?.id || '').trim()
  const teamIds = toPositiveTeamIds(draft?.teamIdsText)
  const weight = Number(draft?.weight)
  if (!editing && id && !/^[A-Za-z0-9_.:-]{1,64}$/.test(id)) return { error: 'ID 只能包含字母、数字、点、下划线、冒号或短横线，最长 64 位。' }
  if (!name) return { error: '名称必填。' }
  if ([...name].length > 120) return { error: '名称不能超过 120 个字符。' }
  if (!isFindXRoute(url)) return { error: 'FindX route 必须是以 / 开头的同源相对路径，不能是外部地址。' }
  if (!isFindXRoute(configPreview)) return { error: '配置预览必须是以 / 开头的同源相对路径。' }
  if (draft?.isPrivate && teamIds.length === 0) return { error: '团队可见时至少填写一个正整数团队 ID。' }
  if (!Number.isFinite(weight)) return { error: '权重必须是数字。' }
  return {
    payload: {
      ...(editing ? {} : id ? { id } : {}),
      name,
      url,
      config_preview: configPreview,
      is_private: Boolean(draft?.isPrivate),
      team_ids: teamIds,
      weight,
      hide: Boolean(draft?.hide),
    },
  }
}

export const systemIntegrationOrderPayload = (rows) => rows.map((row, index) => ({
  id: row.id,
  weight: Number(row.weightDraft ?? row.weight ?? index),
}))

const isFindXRoute = (value) => {
  const text = String(value || '').trim()
  return text.length > 0 && text.length <= 500 && text.startsWith('/') && !text.startsWith('//')
}

export const filterSystemIntegrations = (rows, query) => {
  const text = String(query || '').trim().toLowerCase()
  if (!text) return rows
  return rows.filter((row) => [
    row.name,
    row.url,
    row.configPreview,
    row.updateBy,
    row.createBy,
    row.teamIds.join(' '),
  ].join(' ').toLowerCase().includes(text))
}

export const systemDetailJson = (row) => safeJson({
  id: row.id,
  name: row.name,
  url: row.url,
  config_preview: row.configPreview,
  is_private: row.isPrivate,
  team_ids: row.teamIds,
  hide: row.hide,
  show_in_menu: row.showInMenu,
  weight: row.weight,
  builtin: row.builtin,
  status: row.status,
  capabilities: row.capabilities,
  create_by: row.createBy,
  create_at: row.createAt,
  update_by: row.updateBy,
  update_at: row.updateAt,
  blocked_actions: {
    open_embedded: blockedContracts.systemOpen,
    menu_embedding: 'BLOCKED_BY_CONTRACT：菜单嵌入仍由后端契约阻断，当前仅允许切换是否显示在菜单。',
  },
}, 20000)

export const buildFallbackComponents = (payloads) => {
  const map = new Map()
  payloads.forEach((payload) => {
    const ident = payload.cate || '通用组件'
    const key = ident.toLowerCase()
    const current = map.get(key) || {
      id: `fallback:${key}`,
      ident,
      key,
      logo: logoForIdent(key),
      readme: '当前仅展示仪表盘模板的降级视图。组件实体契约未暴露，新增、编辑、删除、导入与系统集成仍保持 BLOCKED_BY_CONTRACT。',
      disabled: false,
      dashboardCount: 0,
      alertCount: 0,
      metricCount: 0,
      collectCount: 0,
      contractSource: 'dashboard_fallback',
    }
    current.dashboardCount += 1
    map.set(key, current)
  })
  return Array.from(map.values()).sort((a, b) => a.ident.localeCompare(b.ident))
}

export const filterPayloads = (payloads, query, component) => {
  const text = String(query || '').trim().toLowerCase()
  return payloads.filter((item) => {
    const sameComponent = component?.contractSource === 'components'
      ? !item.componentId || String(item.componentId) === String(component.id)
      : item.cate === component?.ident
    const words = [item.name, item.note, item.tags.join(' '), item.cate].join(' ').toLowerCase()
    return sameComponent && (!text || words.includes(text))
  })
}

export const formatContentForPreview = (row) => {
  if (!row?.content) return blockedContracts.payloadContent
  try {
    return safeJson(JSON.parse(row.content), 500000)
  } catch {
    return redactSensitiveText(row.content)
  }
}

const redactExportContent = (content) => {
  if (!content) return ''
  if (typeof content !== 'string') return safeJson(content, 500000)
  try {
    return safeJson(JSON.parse(content), 500000)
  } catch {
    return redactSensitiveText(content)
  }
}

export const buildExportBody = (rows) => rows.map((row) => redactExportContent(row.content)).filter(Boolean)
export const safeErrorText = (err, fallback) => sanitizeDisplayText(redactText(err?.message || fallback))
