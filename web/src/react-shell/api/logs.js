import { get, isPermissionError, normalizeList, post, put, del, redactText } from './http.js'

export const LOG_BLOCKERS = {
  query: 'PENDING: 通用 OTel 日志查询、分页、详情、导出和查询审计缺少 Logs Adapter 契约；当前仅 FindX 审计日志可查询。',
  fields: 'PENDING: 字段发现、已选字段保存、字段值联想和字段类型统计缺少 Logs Adapter 契约。',
  context: 'PENDING: 日志上下文、前后文游标、同实例和同 Trace 窗口缺少 Logs Adapter 契约。',
  aggregate: 'PENDING: 通用生产日志聚合、分桶、分组、趋势图和取消请求缺少 Logs Adapter 契约；当前仅 FindX 审计日志可聚合。',
  live: 'PENDING: 实时日志流、断线重连、背压、暂停和权限续期缺少 Logs Adapter 契约。',
  pipeline: 'PENDING: 生产管道部署、生效、远程下发、回滚和审计契约尚未接入。',
  savedViews: 'PENDING: 保存视图分享和跨来源加载仍缺少 Logs Adapter 契约。',
  traceLink: 'PENDING: 日志 Trace 字段提取、链路详情深链和权限校验缺少 Logs Adapter 与 APM 联动契约。',
  agentLinkage: 'PENDING: 日志采集能力包、主机 FindX Agent 心跳、数据到达和配置版本缺少 Logs-Agent 映射契约。',
}

export const LOG_SOURCES = [
  { value: 'findx_audit', label: 'FindX 审计日志', real: true },
  { value: 'loki', label: 'Loki', real: false },
  { value: 'elasticsearch', label: 'Elasticsearch', real: false },
  { value: 'custom', label: '自定义', real: false },
  { value: 'otel', label: '通用 OTel 日志', real: false },
]

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== null && value !== undefined),
)

export const formatLogError = error => {
  if (isPermissionError(error)) return error.status === 401 ? '登录状态已过期，请重新登录。' : '当前账号没有日志中心权限。'
  if ([404, 405, 409, 501, 503].includes(error?.status)) return redactText(error.message || 'PENDING: 接口未开放。')
  return redactText(error?.message || '日志中心请求失败')
}

export const logsApi = {
  query: params => get('/logs', { params: cleanParams(params) }),
  fields: () => get('/logs/fields'),
  addField: body => post('/logs/fields', body),
  removeField: body => del('/logs/fields', { data: body }),
  toggleIndex: (fieldName, indexed) => put(`/logs/fields/${encodeURIComponent(fieldName)}/index`, { indexed }),
  aggregate: params => get('/logs/aggregate', { params: cleanParams(params) }),
  context: params => get('/logs/context', { params: cleanParams(params) }),
  contextById: (id, params) => get(`/api/v1/logs/context`, { params: cleanParams({ id, ...params }) }),
  pipelines: {
    list: version => get(`/logs/pipelines/${encodeURIComponent(version || 'latest')}`),
    save: body => post('/logs/pipelines', body),
    update: (id, body) => put(`/logs/pipelines/${encodeURIComponent(id)}`, body),
    remove: id => del(`/logs/pipelines/${encodeURIComponent(id)}`),
    preview: body => post('/logs/pipelines/preview', body),
  },
  views: {
    list: sourcePage => get('/explorer/views', { params: cleanParams({ sourcePage }) }).then(normalizeList),
    save: body => post('/explorer/views', body),
    update: (id, body) => put(`/explorer/views/${encodeURIComponent(id)}`, body),
    remove: id => del(`/explorer/views/${encodeURIComponent(id)}`),
  },
  trace: traceId => get(`/traces/${encodeURIComponent(traceId)}`),
}
