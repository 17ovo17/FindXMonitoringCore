import { get, normalizeList, post } from './http.js'

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== undefined && value !== null),
)

export const metricQueryApi = {
  datasources: async () => normalizeList(await get('/monitor/datasources')),
  instantQuery: (body) => post('/monitor/query', body),
  rangeQuery: (body) => post('/monitor/query-range', body),
  metrics: async (params) => normalizeList(await get('/monitor/metrics', { params: cleanParams(params) })),
  labels: async (params) => normalizeList(await get('/monitor/labels', { params: cleanParams(params) })),
  labelValues: async (params) => normalizeList(await get('/monitor/label-values', { params: cleanParams(params) })),
  aiGenerateQuery: async (message) => {
    const resp = await post('/ai-sre/chat', { message: `生成 PromQL: ${message}`, stream: false })
    const text = resp?.data?.content || resp?.content || resp?.message || ''
    const match = text.match(/```[a-z]*\n?([\s\S]*?)```/) || text.match(/`([^`]+)`/)
    return match ? match[1].trim() : text.trim()
  },
}
