import { del, get, normalizeList, post, put } from './http.js'

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => {
    if (Array.isArray(value)) return value.length > 0
    return value !== '' && value !== undefined && value !== null
  }),
)

const rulePath = (id) => `/monitor/alert-rules/${encodeURIComponent(id)}`
const eventPath = (id) => `/monitor/events/${encodeURIComponent(id)}`

export const normalizeAlertList = (value, key) => {
  const direct = normalizeList(value)
  if (direct.length) return direct
  if (Array.isArray(value?.[key])) return value[key]
  return []
}

export const alertsApi = {
  listRules: async (params) => normalizeAlertList(await get('/monitor/alert-rules', { params: cleanParams(params) }), 'rules'),
  createRule: (body) => post('/monitor/alert-rules', body),
  getRule: (id) => get(rulePath(id)),
  updateRule: (id, body) => put(rulePath(id), body),
  removeRule: (id) => del(rulePath(id)),
  enableRule: (id) => post(`${rulePath(id)}/enable`),
  disableRule: (id) => post(`${rulePath(id)}/disable`),
  cloneRule: (id) => post(`${rulePath(id)}/clone`),
  rollbackRule: (id, body) => post(`${rulePath(id)}/rollback`, body || {}),
  tryRunRule: (id, body) => post(`${rulePath(id)}/tryrun`, body || {}),
  batchEnableRules: (ids) => post('/monitor/alert-rules/batch-enable', { ids }),
  batchDisableRules: (ids) => post('/monitor/alert-rules/batch-disable', { ids }),
  batchDeleteRules: (ids) => post('/monitor/alert-rules/batch-delete', { ids }),
  importRules: (body) => post('/monitor/alert-rules/import', body),
  listNotificationChannels: async () => normalizeAlertList(await get('/monitor/notification-channels'), 'channels'),
  listNotificationTemplates: async () => normalizeAlertList(await get('/monitor/notification-templates'), 'templates'),
  listBusinessGroups: async () => normalizeAlertList(await get('/monitor/business-groups'), 'groups'),
  listCurrentEvents: async (params) => normalizeAlertList(await get('/monitor/events/current', { params: cleanParams(params) }), 'events'),
  listHistoryEvents: async (params) => normalizeAlertList(await get('/monitor/events/history', { params: cleanParams(params) }), 'events'),
  getEvent: (id) => get(eventPath(id)),
  ackEvent: (id, body) => post(`${eventPath(id)}/ack`, body || {}),
  assignEvent: (id, body) => post(`${eventPath(id)}/assign`, body || {}),
  resolveEvent: (id, body) => post(`${eventPath(id)}/resolve`, body || {}),
  archiveEvent: (id, body) => post(`${eventPath(id)}/archive`, body || {}),
  deleteEvent: (id) => del(eventPath(id)),
  batchAckEvents: (ids, body) => post('/monitor/events/batch-ack', { ids, ...body }),
  batchMuteEvents: (ids, body) => post('/monitor/events/batch-mute', { ids, ...body }),
  batchDeleteEvents: (ids) => post('/monitor/events/batch-delete', { ids }),
}
