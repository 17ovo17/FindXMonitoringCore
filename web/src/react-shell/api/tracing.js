import { get, isPermissionError, normalizeList, post, put, redactText } from './http.js'

export const TRACING_BLOCKERS = {
  overview: 'BLOCKED_BY_CONTRACT: 链路总览、采集覆盖率和错误率聚合缺少 APM Adapter 契约。',
  selectors: 'BLOCKED_BY_CONTRACT: 服务、实例、端点和标签选择器缺少 APM Adapter 契约。',
  topology: 'BLOCKED_BY_CONTRACT: 服务拓扑、层级拓扑、节点指标和调用边指标缺少 APM Adapter 契约。',
  traces: 'BLOCKED_BY_CONTRACT: Trace 检索、标签联想、Span 树和日志关联缺少 APM Adapter 契约。',
  profiling: 'BLOCKED_BY_CONTRACT: Profiling 任务创建、列表、取消、重试和结果查看缺少 APM Adapter 契约。',
  alarms: 'BLOCKED_BY_CONTRACT: 链路告警列表、详情、确认和规则关联缺少 APM Adapter 契约。',
  settings: 'BLOCKED_BY_CONTRACT: 链路保留期、采样、标签索引和查询保护保存缺少 APM Adapter 契约。',
  agentLinkage: 'BLOCKED_BY_CONTRACT: 服务覆盖率、Trace 详情反查探针状态和拓扑节点探针证据缺少 Agent 联动契约。',
}

export const formatTracingError = error => {
  if (isPermissionError(error)) return error.status === 401 ? '登录状态已过期，请重新登录。' : '当前账号没有链路监控权限。'
  if ([404, 405, 501].includes(error?.status)) return `BLOCKED_BY_CONTRACT: ${redactText(error.message || '接口未开放')}`
  return redactText(error?.message || '链路监控请求失败')
}

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== null && value !== undefined),
)

const normalizeRows = value => normalizeList(value)

export const tracingApi = {
  overview: params => get('/apm/overview', { params: cleanParams(params) }),
  selectors: {
    services: params => get('/tracing/selectors/services', { params: cleanParams(params) }).then(normalizeRows),
    instances: params => get('/tracing/selectors/instances', { params: cleanParams(params) }).then(normalizeRows),
    endpoints: params => get('/tracing/selectors/endpoints', { params: cleanParams(params) }).then(normalizeRows),
    tagKeys: params => get('/apm/trace-tags/keys', { params: cleanParams(params) }).then(normalizeRows),
    tagValues: params => get('/apm/trace-tags/values', { params: cleanParams(params) }).then(normalizeRows),
  },
  topology: params => get('/tracing/topology', { params: cleanParams(params) }),
  traces: {
    query: body => post('/tracing/traces/query', body),
    detail: traceId => get(`/apm/traces/${encodeURIComponent(traceId)}`),
    spans: traceId => get(`/tracing/traces/${encodeURIComponent(traceId)}/spans`).then(normalizeRows),
  },
  profiling: {
    list: params => get('/apm/profiling/tasks', { params: cleanParams(params) }).then(normalizeRows),
    create: body => post('/apm/profiling/tasks', body),
    cancel: id => post(`/apm/profiling/tasks/${encodeURIComponent(id)}/cancel`),
  },
  alarms: {
    list: params => get('/apm/alarms', { params: cleanParams(params) }).then(normalizeRows),
    ack: id => post(`/apm/alarms/${encodeURIComponent(id)}/ack`),
  },
  settings: {
    get: () => get('/apm/settings'),
    save: body => put('/apm/settings', body),
  },
}
