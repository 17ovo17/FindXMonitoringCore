import { del, get, normalizeList, post, put } from './http.js'

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== undefined && value !== null),
)

const dashboardPath = (id) => `/monitor/dashboards/${encodeURIComponent(id)}`
const templatePath = (id) => `/monitor/dashboard-templates/${encodeURIComponent(id)}`

export const normalizeDashboardList = (value) => {
  const direct = normalizeList(value)
  if (direct.length) return direct
  if (Array.isArray(value?.dashboards)) return value.dashboards
  return []
}

export const dashboardsApi = {
  list: async (params) => normalizeDashboardList(await get('/monitor/dashboards', { params: cleanParams(params) })),
  create: (body) => post('/monitor/dashboards', body),
  detail: (id) => get(dashboardPath(id)),
  update: (id, body) => put(dashboardPath(id), body),
  remove: (id) => del(dashboardPath(id)),
  clone: (id) => post(`${dashboardPath(id)}/clone`),
  share: (id) => post(`${dashboardPath(id)}/share`),
  listTemplates: async (params) => normalizeDashboardList(await get('/monitor/dashboard-templates', { params: cleanParams(params) })),
  getTemplate: (id) => get(templatePath(id)),
  importTemplate: (id, body) => post(`${templatePath(id)}/import`, body),
}
