import axios from 'axios'

const http = axios.create({ baseURL: '/api/v1', timeout: 15000 })
const SECRET_RE = /((?:api[_-]?key|token|password|passwd|secret|dsn|cookie|authorization)\s*["']?\s*[:=]\s*["']?)[^"',\s}&]+/ig
const BEARER_RE = /(bearer\s+)[^"',\s}]+/ig
const COOKIE_RE = /(cookie\s*[:=]\s*)[^;\n"',}]+/ig
const URL_RE = /https?:\/\/[^\s"'<>]+/ig

export const redactText = value => String(value ?? '')
  .replace(URL_RE, '<URL>')
  .replace(BEARER_RE, '$1<TOKEN>')
  .replace(COOKIE_RE, '$1<COOKIE>')
  .replace(SECRET_RE, '$1<SECRET>')

const safeError = error => {
  const status = error?.response?.status
  const data = error?.response?.data
  const raw = data?.error || data?.message || error?.message || '请求失败'
  const message = redactText(raw).slice(0, 220)
  return status ? `HTTP ${status}: ${message}` : message
}

const wrapError = error => {
  const wrapped = new Error(safeError(error))
  wrapped.status = error?.response?.status
  wrapped.code = error?.code
  return wrapped
}

http.interceptors.request.use(config => {
  const token = localStorage.getItem('aiw-token')
  config.headers = config.headers || {}
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

http.interceptors.response.use(
  response => response.data,
  error => Promise.reject(wrapError(error))
)

const qs = params => Object.fromEntries(
  Object.entries(params || {}).filter(([, value]) => value !== '' && value !== undefined && value !== null)
)

const dashboardPath = id => `/monitor/dashboards/${encodeURIComponent(id)}`

export const normalizeDashboardList = value => {
  if (Array.isArray(value)) return value
  if (Array.isArray(value?.items)) return value.items
  if (Array.isArray(value?.data)) return value.data
  if (Array.isArray(value?.list)) return value.list
  if (Array.isArray(value?.rows)) return value.rows
  if (Array.isArray(value?.dashboards)) return value.dashboards
  return []
}

export const dashboardsApi = {
  list: params => http.get('/monitor/dashboards', { params: qs(params) }),
  create: body => http.post('/monitor/dashboards', body),
  update: (id, body) => http.put(dashboardPath(id), body),
  remove: id => http.delete(dashboardPath(id)),
  clone: id => http.post(`${dashboardPath(id)}/clone`),
  share: (id, body) => http.post(`${dashboardPath(id)}/share`, body),
}
