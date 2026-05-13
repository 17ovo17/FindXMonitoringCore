const sectionRedirect = (path, section, extra = {}) => ({ path, query: { section, ...extra } })
const agentMergedSections = new Set(['install', 'templates', 'heartbeat', 'data-arrival', 'config', 'plugins'])

const keepAgentFilters = (query = {}) => {
  const keys = ['q', 'os', 'runtime', 'status', 'package']
  const next = Object.fromEntries(keys.filter((key) => query[key] !== undefined).map((key) => [key, query[key]]))
  const section = String(query.section || '')
  const legacySection = String(query.legacySection || '')
  if (agentMergedSections.has(section)) next.legacySection = section
  else if (agentMergedSections.has(legacySection)) next.legacySection = legacySection
  return next
}

const keepIntegrationTemplateQuery = (query = {}) => {
  const keys = ['component', 'tab']
  return Object.fromEntries(keys.filter((key) => query[key] !== undefined).map((key) => [key, query[key]]))
}

const monitorSections = {
  overview: ['/query', 'metrics'],
  datasource: ['/integrations', 'datasources'],
  datasources: ['/integrations', 'datasources'],
  query: ['/query', 'metrics'],
  metrics: ['/query', 'metrics'],
  dashboard: ['/dashboards', 'list'],
  dashboards: ['/dashboards', 'list'],
  collection: ['/agents', 'hosts'],
  templates: ['/integrations', 'templates'],
  agent: ['/agents', 'hosts'],
  targets: ['/assets', 'hosts'],
  rules: ['/alerts', 'rules'],
  events: ['/alerts', 'events'],
}

const settingsTabs = {
  credentials: ['/agents', 'hosts'],
  datasource: ['/integrations', 'datasources'],
  metrics: ['/query', 'metrics'],
  oncall: ['/org', 'teams'],
  'health-audit': ['/platform', 'health'],
  profiles: ['/org', 'users'],
}

const alertSections = {
  notification: ['/notifications', 'channels'],
  notify: ['/notifications', 'channels'],
  oncall: ['/org', 'teams'],
  'recording-rules': ['/query', 'recording-rules'],
  'aggr-views': ['/alerts', 'events'],
  subscribes: ['/alerts', 'subscriptions'],
  pipeline: ['/alerts', 'event-pipelines'],
  workflows: ['/alerts', 'event-pipelines'],
  'event-pipelines': ['/alerts', 'event-pipelines'],
  audit: ['/platform', 'audit'],
}

const plainSectionMaps = {
  notifications: {
    oncall: ['/org', 'teams'],
    pipeline: ['/alerts', 'event-pipelines'],
    configs: ['/notifications', 'rules'],
    records: ['/notifications', 'rules'],
  },
  assets: {
    installs: ['/agents', 'hosts'],
    credentials: ['/agents', 'hosts'],
    'agent-packages': ['/agents', 'packages'],
    'agent-install': ['/agents', 'hosts'],
  },
  agents: {
    install: ['/agents', 'hosts'],
    templates: ['/agents', 'hosts'],
    heartbeat: ['/agents', 'hosts'],
    'data-arrival': ['/agents', 'hosts'],
    config: ['/agents', 'hosts'],
    plugins: ['/agents', 'hosts'],
  },
  query: {
    'metric-mapping': ['/aiops', 'knowledge'],
    dashboard: ['/dashboards', 'list'],
    dashboards: ['/dashboards', 'list'],
    list: ['/dashboards', 'list'],
  },
  dashboards: {
    shares: ['/dashboards', 'list'],
  },
  org: {
    permissions: ['/org', 'roles'],
    tokens: ['/org', 'users'],
    business: ['/org', 'business'],
  },
  aiops: {
    'asset-chat': ['/aiops', 'diagnosis'],
  },
  integrations: {
    collection: ['/integrations', 'templates'],
    collectors: ['/integrations', 'templates'],
    components: ['/integrations', 'templates'],
    templates: ['/integrations', 'templates'],
    datasource: ['/integrations', 'datasources'],
    datasources: ['/integrations', 'datasources'],
    embedded: ['/integrations', 'systems'],
    'embedded-products': ['/integrations', 'systems'],
    systems: ['/integrations', 'systems'],
    labels: ['/aiops', 'knowledge'],
    system: ['/platform', 'site'],
    cmdb: ['/assets', 'cmdb'],
  },
  tracing: {
    alarms: ['/alerts', 'tracing-alarms'],
    alerting: ['/alerts', 'tracing-alarms'],
  },
}

const directRedirects = {
  '/findx-agent': ['/agents', 'overview'],
  '/settings/ai': ['/platform', 'models'],
  '/settings/profiles': ['/org', 'users'],
  '/settings/oncall': ['/org', 'teams'],
  '/workbench': ['/aiops', 'diagnosis'],
  '/knowledge': ['/aiops', 'knowledge'],
  '/workflows': ['/aiops', 'workflow'],
  '/topology': ['/assets', 'business'],
  '/diagnose': ['/aiops', 'knowledge'],
  '/alert-mutes': ['/alerts', 'mutes'],
  '/alert-subscribes': ['/alerts', 'subscriptions'],
  '/alert-his-events': ['/alerts', 'history-events'],
  '/users': ['/org', 'users'],
  '/user-groups': ['/org', 'teams'],
  '/busi-groups': ['/org', 'business'],
  '/roles': ['/org', 'roles'],
  '/system/site-settings': ['/platform', 'site'],
  '/system/variable-settings': ['/platform', 'variables'],
  '/system/sso-settings': ['/platform', 'sso'],
  '/system/alerting-engines': ['/platform', 'alerting-engines'],
  '/datasources': ['/integrations', 'datasources'],
  '/components': ['/integrations', 'templates'],
  '/embedded-products': ['/integrations', 'systems'],
  '/metric/explorer': ['/query', 'metrics'],
  '/log/explorer': ['/logs', 'query'],
  '/business-probes': ['/status', 'config'],
  '/status-page': ['/status', 'public'],
  '/metrics-built-in': ['/query', 'built-in-metrics'],
  '/object/explorer': ['/query', 'objects'],
  '/recording-rules': ['/query', 'recording-rules'],
  '/trace/explorer': ['/tracing', 'traces'],
  '/trace/dependencies': ['/tracing', 'topology'],
}

const fromPair = (pair, extra = {}) => sectionRedirect(pair[0], pair[1], extra)
const trimTrailingSlash = (path) => (path.length > 1 ? path.replace(/\/+$/, '') : path)

export const normalizeLegacyRoute = ({ path, query }) => {
  path = trimTrailingSlash(path)
  if (path === '/') return sectionRedirect('/query', 'metrics')

  const traceMatch = path.match(/^\/traces\/([^/]+)$/)
  if (traceMatch) return sectionRedirect(`/tracing/${encodeURIComponent(traceMatch[1])}`, 'trace-detail')

  if (directRedirects[path]) return fromPair(directRedirects[path])

  if (path === '/monitor') return fromPair(monitorSections[String(query.section || 'overview')] || monitorSections.overview)
  if (path === '/settings') return fromPair(settingsTabs[String(query.tab || '')] || ['/platform', 'site'])
  if (path === '/alerts' && alertSections[String(query.section || '')]) return fromPair(alertSections[String(query.section || '')])

  const sectionMapKey = path.slice(1)
  const hit = plainSectionMaps[sectionMapKey]?.[String(query.section || '')]
  if (!hit) return null
  if (sectionMapKey === 'agents') return fromPair(hit, keepAgentFilters(query))
  if (sectionMapKey === 'integrations' && hit[1] === 'templates') return fromPair(hit, keepIntegrationTemplateQuery(query))
  return fromPair(hit)
}
