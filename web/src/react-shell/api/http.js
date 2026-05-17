import axios from 'axios'

const applyStoredAuthHeader = (config) => {
  const token = localStorage.getItem('aiw-token')
  const headers = { ...config.headers }
  for (const key of ['common', 'get', 'post', 'put', 'patch', 'delete', 'head', 'options']) {
    delete headers[key]
  }
  delete headers.Authorization
  delete headers.authorization
  if (token) headers.Authorization = `Bearer ${token}`
  config.headers = headers
  return config
}

const SECRET_RE = /((?:api[_-]?key|token|password|passwd|secret|dsn|cookie|authorization)\s*["']?\s*[:=]\s*["']?)[^"',\s}&]+/ig
const BEARER_RE = /(bearer\s+)[^"',\s}]+/ig
const COOKIE_RE = /(cookie\s*[:=]\s*)[^;\n"',}]+/ig
const URL_RE = /https?:\/\/[^\s"'<>]+/ig
const PERMISSION_STATUS = new Set([401, 403])

export const AUTH_EXPIRED_EVENT = 'findx-auth-expired'

export const redactText = (value) => String(value ?? '')
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

export const normalizeList = (value) => {
  if (Array.isArray(value)) return value
  if (Array.isArray(value?.items)) return value.items
  if (Array.isArray(value?.data)) return value.data
  if (Array.isArray(value?.list)) return value.list
  if (Array.isArray(value?.rows)) return value.rows
  if (Array.isArray(value?.value)) return value.value
  return []
}

const http = axios.create({ baseURL: '/api/v1', timeout: 15000 })

http.interceptors.request.use(applyStoredAuthHeader)

const emitAuthExpired = () => {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
}

const safeError = (error) => {
  const status = error?.response?.status
  const data = error?.response?.data
  const raw = data?.error || data?.message || error?.message || '请求失败'
  const message = redactText(raw).slice(0, 220)
  return status ? `HTTP ${status}: ${message}` : message
}

const wrapError = (error) => {
  const status = error?.response?.status
  const body = error?.response?.data
  if (status === 401) emitAuthExpired()
  const wrapped = new Error(safeError(error))
  wrapped.status = status
  wrapped.code = error?.code
  wrapped.body = body
  return wrapped
}

http.interceptors.response.use(
  (resp) => resp.data,
  (error) => Promise.reject(wrapError(error)),
)

export const isPermissionError = (error) => PERMISSION_STATUS.has(error?.status)
export const get = (url, config) => http.get(url, config)
export const post = (url, body = {}, config) => http.post(url, body, config)
export const put = (url, body = {}, config) => http.put(url, body, config)
export const del = (url, config) => http.delete(url, config)
