import axios from 'axios'
import { applyStoredAuthHeader } from './authHeaders'

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

http.interceptors.request.use(applyStoredAuthHeader)

http.interceptors.response.use(
  response => response.data,
  error => Promise.reject(wrapError(error))
)

const channelPath = id => `/notifications/channels/${encodeURIComponent(id)}`
const rulePath = id => `/notifications/rules/${encodeURIComponent(id)}`
const templatePath = id => `/notifications/templates/${encodeURIComponent(id)}`

export const notificationsApi = {
  listChannels: () => http.get('/notifications/channels'),
  saveChannel: body => http.post('/notifications/channels', body),
  deleteChannel: id => http.delete(channelPath(id)),
  listRules: () => http.get('/notifications/rules'),
  getRule: id => http.get(rulePath(id)),
  saveRules: body => http.post('/notifications/rules', Array.isArray(body) ? body : [body]),
  updateRule: body => http.put(rulePath(body.id), body),
  deleteRules: ids => http.delete('/notifications/rules', { data: { ids } }),
  toggleRule: (id, enabled) => http.post(`${rulePath(id)}/${enabled ? 'enable' : 'disable'}`),
  cloneRule: id => http.post(`${rulePath(id)}/clone`),
  testRule: body => http.post('/notifications/rules/test', body),
  getRuleStatistics: (id, days = 7) => http.get(`${rulePath(id)}/statistics`, { params: { days } }),
  getRuleEvents: id => http.get(`${rulePath(id)}/events`),
  getRuleAlertRules: id => http.get(`${rulePath(id)}/alert-rules`),
  listTemplates: notifyChannelIdent => http.get('/notifications/templates', { params: notifyChannelIdent ? { notify_channel_ident: notifyChannelIdent } : {} }),
  getTemplate: id => http.get(templatePath(id)),
  saveTemplates: body => http.post('/notifications/templates', Array.isArray(body) ? body : [body]),
  updateTemplate: body => http.put(templatePath(body.id), body),
  deleteTemplates: ids => http.delete('/notifications/templates', { data: { ids } }),
  cloneTemplate: id => http.post(`${templatePath(id)}/clone`),
  previewTemplate: body => http.post('/notifications/templates/preview', body),
  previewTemplateByID: (id, body = {}) => http.post(`${templatePath(id)}/preview`, body),
}
