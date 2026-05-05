import axios from 'axios'

const http = axios.create({ baseURL: '/api/v1', timeout: 15000 })
const BEARER_RE = /(bearer\s+)[^"',\s}]+/ig
const COOKIE_RE = /(cookie\s*[:=]\s*)[^;\n"',}]+/ig
const SECRET_RE = /((?:api[_-]?key|token|password|passwd|secret|dsn|cookie|authorization)\s*["']?\s*[:=]\s*["']?)[^"',\s}&]+/ig
const URL_RE = /https?:\/\/[^\s"'<>]+/ig
const INTERNAL_PATH_RE = /(?:[A-Za-z]:\\|\/(?:opt|var|etc|home|tmp)\/)[^\s"',}]+/g
const PERMISSION_STATUS = new Set([401, 403])

export const redactText = value => String(value ?? '')
  .replace(URL_RE, '<URL>')
  .replace(BEARER_RE, '$1<TOKEN>')
  .replace(COOKIE_RE, '$1<COOKIE>')
  .replace(SECRET_RE, '$1<SECRET>')
  .replace(INTERNAL_PATH_RE, '<PATH>')
  .slice(0, 220)

export const normalizeList = value => {
  if (Array.isArray(value)) return value
  if (Array.isArray(value?.items)) return value.items
  if (Array.isArray(value?.data)) return value.data
  if (Array.isArray(value?.list)) return value.list
  if (Array.isArray(value?.rows)) return value.rows
  return []
}

export const normalizeArray = value => {
  if (Array.isArray(value)) return value.filter(item => item !== '' && item !== null && item !== undefined)
  if (typeof value === 'string') return value.split(/[\n,]/).map(item => item.trim()).filter(Boolean)
  return []
}

export const splitTags = text => String(text || '').split(/[\n,]/).map(item => item.trim()).filter(Boolean)

export const firstValue = (row, keys, fallback = '') => {
  for (const key of keys) {
    const value = row?.[key]
    if (value !== undefined && value !== null && value !== '') return value
  }
  return fallback
}

export const isPermissionError = error => PERMISSION_STATUS.has(error?.status || error?.response?.status)
export const isUnauthorizedError = error => (error?.status || error?.response?.status) === 401

export const formatAssetError = (error, objectName = '资产') => {
  const status = error?.response?.status || error?.status
  if (status === 401) return '登录状态已过期，请重新登录后继续访问'
  if (status === 403) return `当前账号没有${objectName}访问权限`
  if (status === 404) return `${objectName}接口暂不可用，请确认后端已发布对应 API`
  if (status === 409) return `${objectName}失败：${redactText(error?.response?.data?.message || error?.response?.data?.error || error?.message || '存在关联数据，请解除关联后重试')}`
  if (status === 501) return `${objectName}接口尚未实现，请等待后端发布后重试`
  const data = error?.response?.data
  const raw = data?.error || data?.message || error?.message || '请求失败'
  return status ? `HTTP ${status}: ${redactText(raw)}` : redactText(raw)
}

const wrapError = error => {
  const wrapped = new Error(formatAssetError(error))
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

const cleanParams = params => Object.fromEntries(
  Object.entries(params || {}).filter(([, value]) => value !== '' && value !== null && value !== undefined)
)

export const assetsApi = {
  listResourceGroups: params => http.get('/resource-groups', { params: cleanParams(params) }),
  createResourceGroup: payload => http.post('/resource-groups', payload),
  updateResourceGroup: (id, payload) => http.put(`/resource-groups/${encodeURIComponent(id)}`, payload),
  deleteResourceGroup: id => http.delete(`/resource-groups/${encodeURIComponent(id)}`),
  listHostAssets: params => http.get('/host-assets', { params: cleanParams(params) }),
  getHostAsset: id => http.get(`/host-assets/${encodeURIComponent(id)}`),
  updateHostTags: (id, tags) => http.put(`/host-assets/${encodeURIComponent(id)}/tags`, { tags }),
  updateHostResourceGroup: (id, resource_group_id) => http.put(`/host-assets/${encodeURIComponent(id)}/resource-group`, { resource_group_id }),
  updateHostWorkspace: (id, workspace_id) => http.put(`/host-assets/${encodeURIComponent(id)}/workspace`, { workspace_id }),
  listAgents: () => http.get('/findx-agents'),
}
