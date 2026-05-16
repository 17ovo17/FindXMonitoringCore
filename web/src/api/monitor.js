import axios from 'axios'
import { applyStoredAuthHeader } from './authHeaders'

const http = axios.create({ baseURL: '/api/v1', timeout: 15000 })
const BEARER_RE = /(bearer\s+)[^"',\s}]+/ig
const COOKIE_RE = /(cookie\s*[:=]\s*)[^;\n"',}]+/ig
const SECRET_RE = /((?:api[_-]?key|token|password|passwd|secret|dsn|cookie|authorization)\s*["']?\s*[:=]\s*["']?)[^"',\s}&]+/ig
const URL_RE = /https?:\/\/[^\s"'<>]+/ig
const PERMISSION_STATUS = new Set([401, 403])

export const redactText = value => String(value ?? '')
  .replace(URL_RE, '<URL>')
  .replace(BEARER_RE, '$1<TOKEN>')
  .replace(COOKIE_RE, '$1<COOKIE>')
  .replace(SECRET_RE, '$1<SECRET>')

export const safeJson = (value, limit = 8000) => {
  try {
    const text = redactText(JSON.stringify(value ?? null, null, 2))
    return text.length > limit ? `${text.slice(0, limit)}\n...内容已截断` : text
  } catch {
    return redactText(String(value ?? ''))
  }
}

export const normalizeList = value => {
  if (Array.isArray(value)) return value
  if (Array.isArray(value?.items)) return value.items
  if (Array.isArray(value?.data)) return value.data
  if (Array.isArray(value?.list)) return value.list
  if (Array.isArray(value?.rows)) return value.rows
  return []
}

export const isPermissionError = error => PERMISSION_STATUS.has(error?.status || error?.response?.status)
export const isUnauthorizedError = error => error?.status === 401 || error?.response?.status === 401

const wrapError = error => {
  const status = error?.response?.status
  const wrapped = new Error(safeError(error))
  wrapped.status = status
  wrapped.code = error?.code
  return wrapped
}

export const safeError = error => {
  const status = error?.response?.status
  const data = error?.response?.data
  const raw = data?.error || data?.message || error?.message || '请求失败'
  const text = redactText(raw).slice(0, 220)
  return status ? `HTTP ${status}: ${text}` : text
}

http.interceptors.request.use(applyStoredAuthHeader)

http.interceptors.response.use(
  resp => resp.data,
  error => Promise.reject(wrapError(error))
)

const qs = params => Object.fromEntries(Object.entries(params || {}).filter(([, v]) => v !== '' && v !== undefined && v !== null))
const post = (url, body = {}) => http.post(url, body)

export const monitorApi = {
  health: () => http.get('/monitor/health'),
  targets: params => http.get('/monitor/targets', { params: qs(params) }),
  agents: () => http.get('/findx-agents'),
  datasources: () => http.get('/monitor/datasources'),
  testDatasource: datasource_id => post('/monitor/query', { datasource_id, query: 'up', timeout_seconds: 3 }),
  query: body => post('/monitor/query', body),
  queryRange: body => post('/monitor/query-range', body),
  metrics: params => http.get('/monitor/metrics', { params: qs(params) }),
  labels: params => http.get('/monitor/labels', { params: qs(params) }),
  labelValues: params => http.get('/monitor/label-values', { params: qs(params) }),
  alertRules: params => http.get('/monitor/alert-rules', { params: qs(params) }),
  alertRuleDetail: id => http.get(`/monitor/alert-rules/${encodeURIComponent(id)}`),
  tryRunRule: (id, body) => post(`/monitor/alert-rules/${encodeURIComponent(id)}/tryrun`, body || {}),
  eventsCurrent: params => http.get('/monitor/events/current', { params: qs(params) }),
  eventsHistory: params => http.get('/monitor/events/history', { params: qs(params) }),
  eventDetail: id => http.get(`/monitor/events/${encodeURIComponent(id)}`),
  eventAction: (id, action, body = {}) => post(`/monitor/events/${encodeURIComponent(id)}/${action}`, body),
}
