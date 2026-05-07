import axios from 'axios'

const http = axios.create({ baseURL: '/api/v1', timeout: 15000 })
const BEARER_RE = /(bearer\s+)[^"',\s}]+/ig
const COOKIE_RE = /(cookie\s*[:=]\s*)[^;\n"',}]+/ig
const SECRET_RE = /((?:api[_-]?key|token|password|passwd|secret|dsn|cookie|authorization|webhook)\s*["']?\s*[:=]\s*["']?)[^"',\s}&]+/ig
const URL_RE = /https?:\/\/[^\s"'<>]+/ig

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
  return []
}

const wrapError = error => {
  const status = error?.response?.status
  const data = error?.response?.data
  const raw = data?.error || data?.message || error?.message || '请求失败'
  const wrapped = new Error(status ? `HTTP ${status}: ${redactText(raw).slice(0, 260)}` : redactText(raw).slice(0, 260))
  wrapped.status = status
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

const channelPath = id => `/notifications/channels/${encodeURIComponent(id)}`

export const notificationsApi = {
  listChannels: () => http.get('/notifications/channels'),
  saveChannel: body => http.post('/notifications/channels', body),
  deleteChannel: id => http.delete(channelPath(id)),
}
