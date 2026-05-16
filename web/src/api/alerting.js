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

export const safeJson = (value, limit = 12000) => {
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
  if (Array.isArray(value?.rules)) return value.rules
  if (Array.isArray(value?.events)) return value.events
  return []
}

export const isPermissionError = error => PERMISSION_STATUS.has(error?.status || error?.response?.status)
export const isUnauthorizedError = error => error?.status === 401 || error?.response?.status === 401

const safeError = error => {
  if (error?.blocked) return error.message
  const status = error?.response?.status
  const data = error?.response?.data
  const raw = data?.error || data?.message || error?.message || '请求失败'
  const text = redactText(raw).slice(0, 260)
  return status ? `HTTP ${status}: ${text}` : text
}

const wrapError = error => {
  const status = error?.response?.status
  const wrapped = new Error(safeError(error))
  wrapped.status = status
  wrapped.code = error?.code
  wrapped.blocked = error?.blocked
  return wrapped
}

http.interceptors.request.use(applyStoredAuthHeader)

http.interceptors.response.use(
  response => response.data,
  error => Promise.reject(wrapError(error))
)

const qs = params => Object.fromEntries(Object.entries(params || {}).filter(([, value]) => value !== '' && value !== undefined && value !== null))
const rulePath = id => `/monitor/alert-rules/${encodeURIComponent(id)}`
const eventPath = id => `/monitor/events/${encodeURIComponent(id)}`

export const blockedContractError = (action, reason) => {
  const error = new Error(`PENDING：${reason || action}`)
  error.blocked = true
  error.action = action
  return error
}

export const alertingApi = {
  datasources: () => http.get('/monitor/datasources'),
  listRules: params => http.get('/monitor/alert-rules', { params: qs(params) }),
  createRule: body => http.post('/monitor/alert-rules', body),
  getRule: id => http.get(rulePath(id)),
  updateRule: (id, body) => http.put(rulePath(id), body),
  removeRule: id => http.delete(rulePath(id)),
  enableRule: id => http.post(`${rulePath(id)}/enable`),
  disableRule: id => http.post(`${rulePath(id)}/disable`),
  cloneRule: id => http.post(`${rulePath(id)}/clone`),
  rollbackRule: (id, body) => http.post(`${rulePath(id)}/rollback`, body || {}),
  tryRunRule: (id, body) => http.post(`${rulePath(id)}/tryrun`, body || {}),
  eventsCurrent: params => http.get('/monitor/events/current', { params: qs(params) }),
  eventsHistory: params => http.get('/monitor/events/history', { params: qs(params) }),
  getEvent: id => http.get(eventPath(id)),
  eventAction: (id, action, body = {}) => http.post(`${eventPath(id)}/${action}`, body),
}
