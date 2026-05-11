import { del, get, normalizeList, post, put, redactText } from './http.js'

const compactParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== undefined && value !== null),
)

const withRedactedError = async (action, request) => {
  try {
    return await request()
  } catch (err) {
    const safe = redactText(err?.message || `${action} failed`)
    const wrapped = new Error(safe)
    wrapped.status = err?.status
    wrapped.code = err?.code
    throw wrapped
  }
}

const firstAvailable = async (action, paths, config) => {
  let lastError
  for (const path of paths) {
    try {
      const data = await get(path, config)
      return { data, path }
    } catch (err) {
      lastError = err
    }
  }
  throw lastError || new Error(`${action} contract unavailable`)
}

const templatePath = (id) => `/monitor/dashboard-templates/${encodeURIComponent(id)}`

export const integrationsApi = {
  listBuiltinComponents(params) {
    return withRedactedError('list components', () => firstAvailable('list components', [
      '/monitor/builtin-components',
      '/builtin-components',
    ], { params: compactParams(params) }))
  },

  listBuiltinPayloadCates(params) {
    return withRedactedError('list payload categories', () => firstAvailable('list payload categories', [
      '/monitor/builtin-payloads/cates',
      '/builtin-payloads/cates',
    ], { params: compactParams(params) }))
  },

  listBuiltinPayloads(params) {
    return withRedactedError('list payloads', () => firstAvailable('list payloads', [
      '/monitor/builtin-payloads',
      '/builtin-payloads',
    ], { params: compactParams(params) }))
  },

  getBuiltinPayload(id) {
    return withRedactedError('get payload', () => firstAvailable('get payload', [
      `/monitor/builtin-payload/${encodeURIComponent(id)}`,
      `/builtin-payload/${encodeURIComponent(id)}`,
    ]))
  },

  createBuiltinComponents(items) {
    return withRedactedError('create components', () => post('/monitor/builtin-components', items))
  },

  updateBuiltinComponent(body) {
    return withRedactedError('update component', () => put('/monitor/builtin-components', body))
  },

  deleteBuiltinComponents(ids) {
    return withRedactedError('delete components', () => del('/monitor/builtin-components', { data: { ids } }))
  },

  createBuiltinPayloads(items) {
    return withRedactedError('create payloads', () => post('/monitor/builtin-payloads', items))
  },

  updateBuiltinPayload(body) {
    return withRedactedError('update payload', () => put('/monitor/builtin-payloads', body))
  },

  deleteBuiltinPayloads(ids) {
    return withRedactedError('delete payloads', () => del('/monitor/builtin-payloads', { data: { ids } }))
  },

  listDashboardTemplates(params) {
    return withRedactedError('list dashboard template fallback', () => get('/monitor/dashboard-templates', { params: compactParams(params) }))
  },

  getDashboardTemplate(id) {
    return withRedactedError('get dashboard template fallback', () => get(templatePath(id)))
  },

  importDashboardTemplate(id, body) {
    return withRedactedError('import dashboard template', () => post(`${templatePath(id)}/import`, body))
  },

  listSystemIntegrations(params) {
    return withRedactedError('list system integrations', () => get('/monitor/system-integrations', { params: compactParams(params) }))
  },

  getSystemIntegration(id) {
    return withRedactedError('get system integration', () => get(`/monitor/system-integrations/${encodeURIComponent(id)}`))
  },

  createSystemIntegration(body) {
    return withRedactedError('create system integration', () => post('/monitor/system-integrations', body))
  },

  updateSystemIntegration(id, body) {
    return withRedactedError('update system integration', () => put(`/monitor/system-integrations/${encodeURIComponent(id)}`, body))
  },

  deleteSystemIntegration(id) {
    return withRedactedError('delete system integration', () => del(`/monitor/system-integrations/${encodeURIComponent(id)}`))
  },

  updateSystemIntegrationWeights(items) {
    return withRedactedError('update system integration weights', () => put('/monitor/system-integrations/weights', items))
  },

  setSystemIntegrationHide(id, hide) {
    return withRedactedError('set system integration menu visibility', () => put(`/monitor/system-integrations/${encodeURIComponent(id)}/hide`, { hide: Boolean(hide) }))
  },
}

export { normalizeList, redactText }
