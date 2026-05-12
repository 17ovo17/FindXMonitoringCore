export const sections = [
  { value: 'overview', label: '链路总览', desc: '查看采集覆盖、服务健康、Trace 查询和链路告警契约状态。' },
  { value: 'services', label: '服务目录', desc: '按服务、实例、端点和层级组织链路对象。' },
  { value: 'topology', label: '服务拓扑', desc: '查看服务、实例、端点调用关系和节点指标。' },
  { value: 'traces', label: 'Trace 检索', desc: '按时间、服务、端点、标签、耗时和状态检索 Trace。' },
  { value: 'trace-detail', label: 'Trace 详情', desc: '查看 Span 树、时间轴、标签、日志和错误事件。', hidden: true },
  { value: 'profiling', label: 'Profiling', desc: '管理按需剖析任务、任务时间线和结果视图。' },
  { value: 'alarms', label: '链路告警', desc: '查看链路异常、慢调用、错误率和告警事件详情。' },
  { value: 'settings', label: '链路设置', desc: '管理保留期、采样、标签索引和查询保护。' },
]

export const visibleSections = sections.filter(item => !item.hidden)
export const sectionSet = new Set(sections.map(item => item.value))
export const layerOptions = ['GENERAL', 'Service', 'Database', 'Cache', 'MQ', 'Browser', 'Gateway']
export const entityOptions = ['service', 'instance', 'endpoint']
export const traceStates = ['ALL', 'SUCCESS', 'ERROR']
export const orderOptions = ['BY_START_TIME', 'BY_DURATION']

export const durationPresets = [
  { value: '15m', label: '最近 15 分钟', ms: 15 * 60 * 1000 },
  { value: '30m', label: '最近 30 分钟', ms: 30 * 60 * 1000 },
  { value: '1h', label: '最近 1 小时', ms: 60 * 60 * 1000 },
  { value: '3h', label: '最近 3 小时', ms: 3 * 60 * 60 * 1000 },
  { value: '6h', label: '最近 6 小时', ms: 6 * 60 * 60 * 1000 },
  { value: '12h', label: '最近 12 小时', ms: 12 * 60 * 60 * 1000 },
  { value: '24h', label: '最近 24 小时', ms: 24 * 60 * 60 * 1000 },
  { value: '7d', label: '最近 7 天', ms: 7 * 24 * 60 * 60 * 1000 },
]

export const durationToRange = (preset) => {
  const found = durationPresets.find(p => p.value === preset)
  if (!found) return defaultDuration()
  const end = Date.now()
  const start = end - found.ms
  return {
    start: new Date(start).toISOString().slice(0, 16),
    end: new Date(end).toISOString().slice(0, 16),
  }
}

export const defaultDuration = () => {
  const end = Date.now()
  const start = end - 15 * 60 * 1000
  return {
    start: new Date(start).toISOString().slice(0, 16),
    end: new Date(end).toISOString().slice(0, 16),
  }
}

export const displayText = (value, fallback = '-') => {
  if (value === null || value === undefined || value === '') return fallback
  const text = String(value)
  return /[\uFFFD]|\?\?\?/.test(text) ? '内容不可读' : text
}

export const fmtTime = value => {
  if (!value) return '-'
  const date = new Date(Number(value) || value)
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString('zh-CN', { hour12: false })
}

export const durationMs = row => {
  if (row?.duration !== undefined) return Number(row.duration) || 0
  if (row?.endTime && row?.startTime) return Math.max(0, Number(row.endTime) - Number(row.startTime))
  return 0
}

export const traceId = row => row?.traceId || row?.trace_id || row?.traceIds?.[0]?.value || row?.traceIds?.[0] || row?.id || ''
export const traceName = row => displayText(row?.endpointName || row?.endpointNames?.[0] || row?.label || row?.serviceCode || traceId(row))
export const rowText = row => Object.values(row || {}).map(value => displayText(value, '')).join(' ').toLowerCase()

export const toTraceCondition = draft => ({
  duration: {
    start: draft.start,
    end: draft.end,
  },
  serviceId: draft.serviceId.trim(),
  serviceInstanceId: draft.instanceId.trim(),
  endpointId: draft.endpointId.trim(),
  traceId: draft.traceId.trim(),
  traceState: draft.traceState,
  minTraceDuration: Number(draft.minDuration) || undefined,
  maxTraceDuration: Number(draft.maxDuration) || undefined,
  queryOrder: draft.queryOrder,
  tags: String(draft.tags || '').split('\n').map(item => item.trim()).filter(Boolean),
  paging: { pageNum: Number(draft.pageNum) || 1, pageSize: 20 },
})
