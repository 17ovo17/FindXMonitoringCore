import { redactText } from '../../api/http.js'
import { DATASOURCE_CONTRACT_BLOCKERS } from '../../api/datasources.js'

const BRAND_RULES = [
  [/\bAI WorkBench\b/gi, 'FindX'],
  [/\bAIOps\b/gi, 'FindX AI SRE'],
  [/\bNightingale\b/gi, 'FindX'],
  [/\bn9e\b/gi, 'FindX'],
  [/\bSkyWalking\b/gi, 'FindX'],
  [/\bSigNoZ\b/gi, 'FindX'],
  [/\bCategraf\b/gi, 'FindX Agent'],
  [/\bCatpaw\b/gi, 'FindX Agent'],
]

const SECRET_KEY_RE = /password|passwd|token|secret|cookie|authorization|api[_-]?key|access[_-]?key|private[_-]?key|dsn|(^|[_-])(user(name)?|login[_-]?user|account)([_-]|$)/i
const DSN_CREDENTIAL_RE = /([a-z][a-z0-9+.-]*:\/\/)([^/@\s:]+):([^/@\s]+)@/ig
const TCP_DSN_CREDENTIAL_RE = /(^|[\s"'])([^:\s"'/@]+):([^@\s"']+)@tcp\(/ig

const PLACEHOLDERS = [
  ['password', '<TOKEN>'],
  ['passwd', '<TOKEN>'],
  ['token', '<TOKEN>'],
  ['secret', '<TOKEN>'],
  ['cookie', '<TOKEN>'],
  ['authorization', '<TOKEN>'],
  ['api_key', '<API_KEY>'],
  ['apikey', '<API_KEY>'],
  ['access_key', '<API_KEY>'],
  ['private_key', '<API_KEY>'],
  ['dsn', '<DB_DSN>'],
  ['user', '<LOGIN_USER>'],
  ['username', '<LOGIN_USER>'],
]

const METRIC_DATASOURCE_TYPES = new Set([
  'prom',
  'prometheus',
  'victoriametrics',
  'vm',
  'vmselect',
  'vmstorage',
  'vminsert',
])

export const datasourceTypes = [
  {
    name: 'Prometheus',
    type: 'prometheus',
    category: 'metrics',
    supported: true,
    logo: '/image/logos/prometheus.png',
    description: '开源监控系统，支持 PromQL 查询和多维时序数据。',
  },
  {
    name: 'VictoriaMetrics',
    type: 'victoriametrics',
    category: 'metrics',
    supported: false,
    logo: '/image/logos/prometheus.png',
    description: '高性能时序数据库，兼容 PromQL 查询语法。',
  },
  {
    name: 'Elasticsearch',
    type: 'elasticsearch',
    category: 'logs',
    supported: false,
    logo: '/image/logos/elasticsearch.png',
    description: '分布式搜索和分析引擎，适用于日志和全文检索。',
  },
  {
    name: 'Loki',
    type: 'loki',
    category: 'logs',
    supported: false,
    logo: '/image/logos/loki.png',
    description: '轻量级日志聚合系统，使用标签索引日志流。',
  },
  {
    name: 'TDengine',
    type: 'tdengine',
    category: 'metrics',
    supported: false,
    logo: '/image/logos/tdengine.png',
    description: '高性能物联网时序数据库，支持 SQL 查询。',
  },
  {
    name: 'ClickHouse',
    type: 'clickhouse',
    category: 'logs',
    supported: false,
    logo: '/image/logos/ck.png',
    description: '列式存储分析数据库，适用于大规模数据分析。',
  },
]

const blocked = (text) => text

export const blockedContracts = {
  add: blocked(DATASOURCE_CONTRACT_BLOCKERS.create),
  edit: blocked(DATASOURCE_CONTRACT_BLOCKERS.edit),
  delete: blocked(DATASOURCE_CONTRACT_BLOCKERS.delete),
  toggle: blocked(DATASOURCE_CONTRACT_BLOCKERS.toggle),
  unsupportedType: blocked(DATASOURCE_CONTRACT_BLOCKERS.create),
  pluginList: blocked(DATASOURCE_CONTRACT_BLOCKERS.pluginList),
  missingTestId: blocked('测试连接需要 datasource_id 或 id，当前列表记录缺少可用标识。'),
  nonMetricTest: blocked(DATASOURCE_CONTRACT_BLOCKERS.testNonMetric),
}

export const safeDisplayText = (value) => BRAND_RULES.reduce(
  (text, [pattern, replacement]) => text.replace(pattern, replacement),
  redactText(value),
)

const placeholderForKey = (key) => {
  const lowered = String(key).toLowerCase()
  const matched = PLACEHOLDERS.find(([name]) => lowered.includes(name))
  return matched?.[1] || '<TOKEN>'
}

const redactSecrets = (value, key = '') => {
  if (value === null || value === undefined) return value
  if (SECRET_KEY_RE.test(key)) return placeholderForKey(key)
  if (Array.isArray(value)) return value.map((item) => redactSecrets(item))
  if (typeof value === 'object') {
    return Object.fromEntries(
      Object.entries(value).map(([entryKey, entryValue]) => [entryKey, redactSecrets(entryValue, entryKey)]),
    )
  }
  if (typeof value === 'string') {
    return value
      .replace(DSN_CREDENTIAL_RE, '$1<LOGIN_USER>:<TOKEN>@')
      .replace(TCP_DSN_CREDENTIAL_RE, '$1<LOGIN_USER>:<TOKEN>@tcp(')
  }
  return value
}

export const safeDisplayJson = (value, limit = 10000) => {
  try {
    const text = safeDisplayText(JSON.stringify(redactSecrets(value), null, 2))
    return text.length > limit ? `${text.slice(0, limit)}\n...内容已截断` : text
  } catch {
    return safeDisplayText(value)
  }
}

export const firstValue = (row, keys) => keys
  .map((key) => row?.[key])
  .find((value) => value !== undefined && value !== null && value !== '')

export const rowKey = (row) => String(row?.id ?? row?.datasource_id ?? row?.name ?? row?.url ?? '')
export const datasourceId = (row) => firstValue(row, ['datasource_id', 'id'])
export const displayType = (row) => firstValue(row, ['plugin_type_name', 'plugin_type', 'type', 'datasource_type', 'category'])
export const displayName = (row) => firstValue(row, ['name', 'display_name', 'id', 'datasource_id'])
export const displayCluster = (row) => firstValue(row, ['cluster_name', 'cluster', 'source', 'origin', 'engine'])
export const displayAddress = (row) => firstValue(row, ['url', 'address', 'endpoint', 'base_url', 'remote_write_url'])
export const displayDescription = (row) => firstValue(row, ['description', 'desc', 'remark', 'note'])

const normalizeDatasourceType = (value) => String(value ?? '')
  .trim()
  .toLowerCase()
  .replace(/[^a-z0-9]/g, '')

export const isMetricDatasource = (row) => {
  const candidates = [
    row?.plugin_type,
    row?.type,
    row?.datasource_type,
    row?.plugin_type_name,
    row?.category,
    row?.engine,
  ]
  return candidates.some((value) => METRIC_DATASOURCE_TYPES.has(normalizeDatasourceType(value)))
}

export const testContractBlockedMessage = (row) => {
  const type = safeDisplayText(displayType(row)) || '未知类型'
  return `${blockedContracts.nonMetricTest} 类型：${type}。`
}

export const normalizedStatus = (row) => String(firstValue(row, ['status', 'state', 'enabled']) ?? 'unknown').toLowerCase()

export const statusText = (row) => {
  const status = normalizedStatus(row)
  if (status === 'enabled' || status === 'true' || status === 'active') return '启用'
  if (status === 'disabled' || status === 'false' || status === 'inactive') return '禁用'
  return '未知'
}
