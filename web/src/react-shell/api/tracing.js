import { AUTH_EXPIRED_EVENT, get, isPermissionError, normalizeList, post, redactText } from './http.js'

export const TRACING_BLOCKERS = {
  overview: '链路总览、采集覆盖率和错误率聚合缺少 APM Adapter 契约。',
  selectors: '服务、实例、端点和标签选择器缺少 APM Adapter 契约。',
  topology: '服务拓扑、层级拓扑、节点指标和调用边指标缺少 APM Adapter 契约。',
  traces: 'Trace 检索、标签联想、Span 树和日志关联缺少 APM Adapter 契约。',
  profiling: 'Profiling 任务创建、列表、取消、重试和结果查看缺少 APM Adapter 契约。',
  alarms: '链路告警列表、详情、确认和规则关联缺少 APM Adapter 契约。',
  settings: '链路保留期、采样、标签索引和查询保护保存缺少 APM Adapter 契约。',
  agentLinkage: '服务覆盖率、Trace 详情反查探针状态和拓扑节点探针证据缺少 Agent 联动契约。',
}

const blockedCode = value => String(value || '').toUpperCase() === 'PENDING'
const blockedStatus = value => String(value || '').toLowerCase() === 'blocked'

const blockedMessageFrom = value => {
  if (!value || typeof value !== 'object') return ''
  return value.message || value.error || value.reason || value.blocked_reason || value.detail || ''
}

const isBlockedEnvelope = value => {
  if (!value || typeof value !== 'object') return false
  return blockedCode(value.code) || blockedStatus(value.status) || blockedCode(value.error_code) || /PENDING/i.test(blockedMessageFrom(value))
}

const blockedErrorFromEnvelope = (value, fallback = '链路查询服务契约未开放') => {
  const parts = []
  if (value?.contract_id) parts.push(`contract_id=${redactText(value.contract_id)}`)
  if (value?.code) parts.push(`code=${redactText(value.code)}`)
  const message = redactText(blockedMessageFrom(value) || fallback)
  const error = new Error(`${parts.length ? parts.join(' ') + ' ' : ''}${message}`)
  error.status = value?.status_code || value?.http_status
  error.code = value?.code || 'PENDING'
  error.contract_id = value?.contract_id
  return error
}

const ensureNotBlocked = (value, fallback) => {
  if (isBlockedEnvelope(value)) throw blockedErrorFromEnvelope(value, fallback)
  return value
}

const requestJson = async (url, options = {}) => {
  const headers = { Accept: 'application/json', ...(options.headers || {}) }
  if (options.body !== undefined) headers['Content-Type'] = 'application/json'
  const token = localStorage.getItem('aiw-token')
  if (token) headers.Authorization = `Bearer ${token}`
  const resp = await fetch(`/api/v1${url}`, { ...options, headers, body: options.body === undefined ? undefined : JSON.stringify(options.body) })
  const text = await resp.text()
  let data = null
  if (text) {
    try { data = JSON.parse(text) } catch { data = { message: text } }
  }
  if (!resp.ok) {
    if (resp.status === 401 && typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
    }
    if (isBlockedEnvelope(data) || resp.status === 409) {
      const err = blockedErrorFromEnvelope(data || {}, `HTTP ${resp.status}: 链路查询服务契约阻断`)
      err.status = resp.status
      throw err
    }
    const err = new Error(redactText((data && blockedMessageFrom(data)) || `HTTP ${resp.status}: 链路监控请求失败`))
    err.status = resp.status
    throw err
  }
  return ensureNotBlocked(data, '链路查询服务返回 blocked 响应')
}

export const formatTracingError = error => {
  if (isPermissionError(error)) return error.status === 401 ? '登录状态已过期，请重新登录。' : '当前账号没有链路监控权限。'
  if (error?.contract_id || blockedCode(error?.code) || /PENDING/i.test(error?.message || '')) return redactText(error.message || '链路查询服务契约阻断')
  if ([404, 405, 409, 501].includes(error?.status)) return `${redactText(error.message || '接口未开放')}`
  return redactText(error?.message || '链路监控请求失败')
}

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== null && value !== undefined),
)

const normalizeRows = value => normalizeList(value)

export const tracingApi = {
  overview: params => get('/apm/overview', { params: cleanParams(params) }),
  selectors: {
    services: params => get('/tracing/selectors/services', { params: cleanParams(params) }).then(normalizeRows),
    instances: params => get('/tracing/selectors/instances', { params: cleanParams(params) }).then(normalizeRows),
    endpoints: params => get('/tracing/selectors/endpoints', { params: cleanParams(params) }).then(normalizeRows),
    tagKeys: params => get('/apm/trace-tags/keys', { params: cleanParams(params) }).then(normalizeRows),
    tagValues: params => get('/apm/trace-tags/values', { params: cleanParams(params) }).then(normalizeRows),
  },
  topology: params => get('/tracing/topology', { params: cleanParams(params) }),
  traces: {
    query: body => post('/tracing/traces/query', body),
    detail: traceId => requestJson(`/apm/traces/${encodeURIComponent(traceId)}`),
    spans: traceId => requestJson(`/tracing/traces/${encodeURIComponent(traceId)}/spans`).then(normalizeRows),
  },
  profiling: {
    list: params => get('/apm/profiling/tasks', { params: cleanParams(params) }).then(normalizeRows),
    create: body => post('/apm/profiling/tasks', body),
    cancel: id => post(`/apm/profiling/tasks/${encodeURIComponent(id)}/cancel`),
  },
  alarms: {
    list: params => get('/apm/alarms', { params: cleanParams(params) }).then(normalizeRows),
    ack: id => post(`/apm/alarms/${encodeURIComponent(id)}/ack`),
  },
  settings: {
    get: () => get('/apm/settings'),
    save: body => requestJson('/apm/settings', { method: 'PUT', body }),
  },
}
