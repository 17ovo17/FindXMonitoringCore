import { del, get, isPermissionError, normalizeList, post, put, redactText } from './http.js'

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== null && value !== undefined),
)

export const PROBE_BLOCKERS = {
  contract: 'PENDING: 业务拨测接口尚未返回可验收的状态页、检查项、运行证据或事故维护契约。',
  subscription: 'PENDING: 订阅入口需要 Webhook、邮件和 RSS 的后端订阅契约。',
  dryRun: 'PENDING: 测试拨测需要后端执行器返回真实探测结果、耗时、错误和 evidence。',
  incidents: 'PENDING: 事故维护入口需要事故创建、更新、关闭和审计契约。',
}

export const formatProbeError = (error) => {
  if (isPermissionError(error)) return error.status === 401 ? '登录状态已过期，请重新登录。' : '当前账号没有业务拨测权限。'
  if ([404, 405, 409, 501].includes(error?.status)) return `PENDING: ${redactText(error.message || '业务拨测接口未开放。')}`
  return redactText(error?.message || '业务拨测请求失败。')
}

export const probesApi = {
  statusPage: (slug = 'main') => get(`/probes/status-pages/${encodeURIComponent(slug)}`),
  statusPages: () => get('/probes/status-pages').then(normalizeList),
  saveStatusPage: (body) => post('/probes/status-pages', body),
  checks: (params = {}) => get('/probes/checks', { params: cleanParams(params) }).then(normalizeList),
  createCheck: (body) => post('/probes/checks', body),
  updateCheck: (id, body) => put(`/probes/checks/${encodeURIComponent(id)}`, body),
  deleteCheck: (id) => del(`/probes/checks/${encodeURIComponent(id)}`),
  testCheck: (id) => post(`/probes/checks/${encodeURIComponent(id)}/test`),
  enableCheck: (id) => post(`/probes/checks/${encodeURIComponent(id)}/enable`),
  disableCheck: (id) => post(`/probes/checks/${encodeURIComponent(id)}/disable`),
  notificationBindings: (id) => get(`/probes/checks/${encodeURIComponent(id)}/notification-bindings`),
  saveNotificationBindings: (id, body) => put(`/probes/checks/${encodeURIComponent(id)}/notification-bindings`, body),
  alertBindings: (id) => get(`/probes/checks/${encodeURIComponent(id)}/alert-bindings`),
  saveAlertBindings: (id, body) => put(`/probes/checks/${encodeURIComponent(id)}/alert-bindings`, body),
  incidents: () => get('/probes/incidents').then(normalizeList),
  createIncident: (body) => post('/probes/incidents', body),
  updateIncident: (id, body) => put(`/probes/incidents/${encodeURIComponent(id)}`, body),
  deleteIncident: (id) => del(`/probes/incidents/${encodeURIComponent(id)}`),
}
