import axios from 'axios'

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

export const normalizeList = value => {
  if (Array.isArray(value)) return value
  if (Array.isArray(value?.items)) return value.items
  if (Array.isArray(value?.data)) return value.data
  if (Array.isArray(value?.list)) return value.list
  if (Array.isArray(value?.rows)) return value.rows
  return []
}

export const isPermissionError = error => PERMISSION_STATUS.has(error?.status || error?.response?.status)

export const safeError = error => {
  const status = error?.response?.status || error?.status
  const data = error?.response?.data
  const raw = data?.error || data?.message || error?.message || '请求失败'
  const text = redactText(raw).slice(0, 180)
  if (status === 404) return '业务空间接口暂不可用，请确认后端已提供 /api/v1/workspaces'
  if (status === 501) return '业务空间接口尚未实现，请等待后端发布后重试'
  if (status === 401) return '登录状态已失效，请重新登录后再试'
  if (status === 403) return '当前账号没有业务空间操作权限'
  return status ? `HTTP ${status}: ${text}` : text
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
  resp => resp.data,
  error => Promise.reject(wrapError(error))
)

export const workspaceApi = {
  list: () => http.get('/workspaces'),
  create: payload => http.post('/workspaces', payload),
  update: (id, payload) => http.put(`/workspaces/${encodeURIComponent(id)}`, payload),
  remove: id => http.delete(`/workspaces/${encodeURIComponent(id)}`),
}
