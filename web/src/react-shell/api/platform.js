import { del, get, normalizeList, post, put, redactText, safeJson } from './http.js'

export const PLATFORM_BLOCKERS = {
  llmCrud: 'PENDING: 成熟 LLM 配置的单条详情、extra_config、单条测试和删除禁用契约尚未完整开放。',
  site: 'PENDING: 站点配置读取/保存、权限点、审计和回滚契约尚未开放。',
  variables: 'PENDING: 变量配置 CRUD、加密字段和引用影响分析契约尚未开放。',
  sso: 'PENDING: SSO 配置读取/保存、回调校验、启停和测试登录契约尚未开放。',
  engines: 'PENDING: 告警引擎列表、心跳、启停和删除契约尚未开放。',
  healthAction: 'PENDING: 平台健康修复动作缺少执行、审计和回滚契约。',
}

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== undefined && value !== null),
)

const blockedError = (error, label) => {
  if ([404, 405, 501].includes(error?.status)) {
    const blocked = new Error(`PENDING: ${label} 未接入或未开放。`)
    blocked.status = error.status
    blocked.blocked = true
    return blocked
  }
  return error
}

const normalizeRows = (data) => {
  const rows = normalizeList(data)
  if (rows.length) return rows
  if (Array.isArray(data?.dat)) return data.dat
  if (Array.isArray(data?.events)) return data.events
  if (Array.isArray(data?.providers)) return data.providers
  return []
}

const tryList = async (url, params, label) => {
  try {
    return {
      rows: normalizeRows(await get(url, { params: cleanParams(params) })),
      source: `GET /api/v1${url}`,
      blocked: '',
    }
  } catch (error) {
    throw blockedError(error, label)
  }
}

export const platformApi = {
  async listProviders() {
    const providers = normalizeRows(await get('/ai-providers'))
    const health = await get('/health/ai-providers').catch(() => [])
    return {
      rows: providers.map((item) => {
        const hit = normalizeRows(health).find((h) => String(h.id) === String(item.id))
        return { ...item, health: hit || null }
      }),
      source: 'GET /api/v1/ai-providers + GET /api/v1/health/ai-providers',
    }
  },
  saveProviders: (providers) => post('/ai-providers', providers),
  getEmbedding: () => get('/settings/embedding'),
  saveEmbedding: (body) => put('/settings/embedding', body),
  testEmbedding: (body) => post('/settings/embedding/test', body),
  testProvider: (body) => post('/ai-providers/test', body),
  getReranker: () => get('/settings/reranker'),
  saveReranker: (body) => put('/settings/reranker', body),
  testReranker: (body) => post('/settings/reranker/test', body),
  // 站点设置
  getSiteSettings: () => get('/platform/site'),
  saveSiteSettings: (body) => put('/platform/site', body),
  // 变量设置
  listVariables: (params) => tryList('/platform/variables', params, '变量配置契约'),
  createVariable: (body) => post('/platform/variables', body),
  updateVariable: (id, body) => put(`/platform/variables/${encodeURIComponent(id)}`, body),
  deleteVariable: (id) => del(`/platform/variables/${encodeURIComponent(id)}`),
  // SSO
  getSsoConfig: () => get('/platform/sso'),
  saveSsoConfig: (body) => put('/platform/sso', body),
  testSsoConnection: (body) => post('/platform/sso/test', body),
  // 告警引擎
  listEngines: (params) => tryList('/platform/alerting-engines', params, '告警引擎契约'),
  createEngine: (body) => post('/platform/alerting-engines', body),
  updateEngine: (id, body) => put(`/platform/alerting-engines/${encodeURIComponent(id)}`, body),
  deleteEngine: (id) => del(`/platform/alerting-engines/${encodeURIComponent(id)}`),
  checkEngineHealth: (id) => post(`/platform/alerting-engines/${encodeURIComponent(id)}/health-check`),
  // 兼容旧接口
  site: (params) => tryList('/platform/site', params, '站点配置契约'),
  variables: (params) => tryList('/platform/variables', params, '变量配置契约'),
  sso: (params) => tryList('/platform/sso', params, 'SSO 配置契约'),
  alertingEngines: (params) => tryList('/platform/alerting-engines', params, '告警引擎契约'),
  storageHealth: () => get('/health/storage'),
  async audit(params) {
    try {
      return {
        rows: normalizeRows(await get('/audit/events', { params: cleanParams(params) })),
        source: 'GET /api/v1/audit/events',
      }
    } catch (error) {
      if (error?.status && ![404, 405, 501].includes(error.status)) throw error
      return {
        rows: normalizeRows(await get('/monitor/audit-logs', { params: cleanParams(params) })),
        source: 'fallback GET /api/v1/monitor/audit-logs',
      }
    }
  },
}

export const formatPlatformError = (error) => redactText(error?.message || '请求失败')
export const safePlatformJson = (value, limit = 8000) => safeJson(value, limit)
