import { del, get, normalizeList, post, put } from './http.js'

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== undefined && value !== null),
)

const channelPath = (id) => `/notifications/channels/${encodeURIComponent(id)}`
const rulePath = (id) => `/notifications/rules/${encodeURIComponent(id)}`
const templatePath = (id) => `/notifications/templates/${encodeURIComponent(id)}`

const firstItem = (value) => {
  const items = normalizeList(value)
  return items[0] || value
}

export const notificationsApi = {
  listChannels: async () => normalizeList(await get('/notifications/channels')),
  saveChannel: (body) => post('/notifications/channels', body),
  deleteChannel: (id) => del(channelPath(id)),

  listRules: async (params) => normalizeList(await get('/notifications/rules', { params: cleanParams(params) })),
  getRule: (id) => get(rulePath(id)),
  saveRule: async (body) => firstItem(await post('/notifications/rules', body)),
  updateRule: (id, body) => put(rulePath(id), body),
  deleteRules: (ids) => del('/notifications/rules', { data: { ids } }),
  enableRule: (id) => post(`${rulePath(id)}/enable`),
  disableRule: (id) => post(`${rulePath(id)}/disable`),
  cloneRule: (id) => post(`${rulePath(id)}/clone`),
  testRule: (body) => post('/notifications/rules/test', body),
  getRuleStatistics: (id, days = 7) => get(`${rulePath(id)}/statistics`, { params: { days } }),
  getRuleEvents: async (id) => normalizeList(await get(`${rulePath(id)}/events`)),
  getRuleAlertRules: async (id) => normalizeList(await get(`${rulePath(id)}/alert-rules`)),
  getRuleSubAlertRules: async (id) => normalizeList(await get(`${rulePath(id)}/sub-alert-rules`)),

  listTemplates: async (notifyChannelIdent) => normalizeList(await get('/notifications/templates', { params: cleanParams({ notify_channel_ident: notifyChannelIdent }) })),
  getTemplate: (id) => get(templatePath(id)),
  saveTemplate: async (body) => firstItem(await post('/notifications/templates', body)),
  updateTemplate: (id, body) => put(templatePath(id), body),
  deleteTemplates: (ids) => del('/notifications/templates', { data: { ids } }),
  cloneTemplate: (id) => post(`${templatePath(id)}/clone`),
  previewTemplate: (body) => post('/notifications/templates/preview', body),
  previewTemplateById: (id, body = {}) => post(`${templatePath(id)}/preview`, body),
}
