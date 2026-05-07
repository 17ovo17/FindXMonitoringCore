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
  const text = redactText(raw).slice(0, 220)
  return status ? `HTTP ${status}: ${text}` : text
}

http.interceptors.request.use(config => {
  const token = localStorage.getItem('aiw-token')
  config.headers = config.headers || {}
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

http.interceptors.response.use(
  response => response.data,
  error => Promise.reject(Object.assign(new Error(safeError(error)), { status: error?.response?.status }))
)

const qs = params => Object.fromEntries(Object.entries(params || {}).filter(([, value]) => value !== '' && value !== undefined && value !== null))

export const normalizeList = value => {
  if (Array.isArray(value)) return value
  if (Array.isArray(value?.items)) return value.items
  if (Array.isArray(value?.data)) return value.data
  if (Array.isArray(value?.list)) return value.list
  if (Array.isArray(value?.rows)) return value.rows
  return []
}

const templatePath = id => `/monitor/dashboard-templates/${encodeURIComponent(id)}`

export const integrationsApi = {
  listDashboardTemplates: params => http.get('/monitor/dashboard-templates', { params: qs(params) }),
  getDashboardTemplate: id => http.get(templatePath(id)),
  importDashboardTemplate: (id, body) => http.post(`${templatePath(id)}/import`, body),
}
