const to = (path, section, extraQuery = {}) => ({ path, query: { section, ...extraQuery } })
const agentMergedSections = new Set(['install', 'templates', 'heartbeat', 'data-arrival', 'config', 'plugins'])

const targetFor = (group, item) => {
  if (group.key === 'agents' && agentMergedSections.has(item.section)) {
    if (item.to) {
      return {
        ...item.to,
        query: { ...(item.to.query || {}), section: 'hosts', legacySection: item.section },
      }
    }
    return to(group.path, 'hosts', { legacySection: item.section })
  }
  if (item.to) return item.to
  return to(group.path, item.section)
}

const attachTargets = (group) => {
  const children = group.children.map((item) => ({
    ...item,
    to: targetFor(group, item),
  }))
  const hiddenChildren = (group.hiddenChildren || []).map((item) => ({
    ...item,
    to: targetFor(group, item),
  }))

  return {
    ...group,
    children,
    hiddenChildren,
    to: group.to || children[0]?.to || to(group.path, group.defaultSection),
  }
}

export const navGroups = [
  {
    key: 'asset-center',
    label: '资产中心',
    path: '/assets',
    defaultSection: 'overview',
    children: [
      { section: 'overview', label: '资产概览' },
      { section: 'models', label: '对象建模' },
      { section: 'instances', label: '实例管理' },
      { section: 'agents', label: 'Agent 状态' },
      { section: 'topology', label: '拓扑视图' },
    ],
    hiddenChildren: [
      { section: 'model-detail', label: '模型详情', to: to('/assets', 'model-detail') },
      { section: 'instance-detail', label: '实例详情', to: to('/assets', 'instance-detail') },
      { section: 'business', label: '业务空间', to: to('/assets', 'business') },
      { section: 'cmdb', label: 'CMDB', to: to('/assets', 'cmdb') },
      { section: 'hosts', label: '主机资产', to: to('/assets', 'hosts') },
      { section: 'packages', label: '能力包', to: to('/assets', 'packages') },
    ],
  },
  {
    key: 'integrations',
    label: '集成中心',
    path: '/integrations',
    defaultSection: 'datasources',
    children: [
      { section: 'datasources', label: '数据源' },
      { section: 'templates', label: '模板中心' },
      { section: 'systems', label: '系统集成' },
    ],
  },
  {
    key: 'explorer',
    label: '数据查询',
    path: '/query',
    defaultSection: 'metrics',
    children: [
      {
        section: 'metrics',
        label: '指标查询',
        matchSections: ['metrics', 'built-in-metrics', 'objects', 'recording-rules'],
      },
      { section: 'dashboards', label: '仪表盘', to: to('/dashboards', 'list') },
    ],
    hiddenChildren: [
      { section: 'built-in-metrics', label: '内置指标', to: to('/query', 'built-in-metrics') },
      { section: 'objects', label: '对象查询', to: to('/query', 'objects') },
      { section: 'recording-rules', label: '记录规则', to: to('/query', 'recording-rules') },
      { section: 'dashboard-templates', label: '仪表盘导入', to: to('/dashboards', 'templates') },
      { section: 'dashboard-detail', label: '仪表盘详情', to: to('/dashboards', 'detail') },
    ],
  },
  {
    key: 'alerts',
    label: '告警',
    path: '/alerts',
    defaultSection: 'rules',
    children: [
      { section: 'rules', label: '规则管理' },
      { section: 'mutes', label: '告警屏蔽' },
      { section: 'subscriptions', label: '告警订阅' },
      { section: 'events', label: '告警事件' },
      { section: 'history-events', label: '历史事件' },
      { section: 'tracing-alarms', label: '链路告警' },
      { section: 'event-pipelines', label: '事件流水线' },
    ],
  },
  {
    key: 'notification',
    label: '通知',
    path: '/notifications',
    defaultSection: 'rules',
    children: [
      { section: 'rules', label: '通知规则' },
      { section: 'channels', label: '通知媒介' },
      { section: 'templates', label: '消息模板' },
    ],
  },
  {
    key: 'tracing',
    label: '链路监控',
    path: '/tracing',
    defaultSection: 'overview',
    children: [
      { section: 'overview', label: '链路总览' },
      { section: 'services', label: '服务目录' },
      { section: 'topology', label: '服务拓扑' },
      { section: 'traces', label: 'Trace 检索' },
      { section: 'profiling', label: 'Profiling' },
      { section: 'settings', label: '链路设置' },
    ],
    hiddenChildren: [
      { section: 'trace-detail', label: 'Trace 详情', to: to('/tracing', 'trace-detail') },
    ],
  },
  {
    key: 'logs',
    label: '日志中心',
    path: '/logs',
    defaultSection: 'query',
    children: [
      { section: 'query', label: '日志检索' },
      { section: 'live', label: '实时日志' },
      { section: 'fields', label: '字段筛选' },
      { section: 'context', label: '上下文' },
      { section: 'aggregate', label: '聚合分析' },
      { section: 'pipelines', label: '接入管道' },
      { section: 'saved-views', label: '保存视图' },
      { section: 'trace-link', label: 'Trace 关联' },
    ],
  },
  {
    key: 'organization',
    label: '人员组织',
    path: '/org',
    defaultSection: 'users',
    children: [
      { section: 'users', label: '用户管理' },
      { section: 'teams', label: '团队组织' },
      { section: 'business', label: '业务组' },
      { section: 'roles', label: '角色管理' },
    ],
  },
  {
    key: 'setting',
    label: '系统配置',
    path: '/platform',
    defaultSection: 'models',
    children: [
      { section: 'models', label: 'AI 模型配置' },
      { section: 'site', label: '站点设置' },
      { section: 'variables', label: '变量设置' },
      { section: 'sso', label: '单点登录' },
      { section: 'alerting-engines', label: '告警引擎' },
      { section: 'health', label: '运行自检' },
      { section: 'audit', label: '审计日志' },
    ],
  },
  {
    key: 'ai-sre',
    label: 'AI SRE',
    path: '/aiops',
    defaultSection: 'diagnosis',
    children: [
      { section: 'diagnosis', label: '诊断会话' },
      { section: 'workflow', label: '工作流' },
      { section: 'health', label: '健康检查' },
      { section: 'report', label: '复盘报告' },
      { section: 'evidence', label: 'Evidence Chain' },
      { section: 'knowledge', label: '知识库' },
      { section: 'remediation', label: '自动修复' },
    ],
  },
].map(attachTargets)

const samePath = (routePath, targetPath) => routePath === targetPath || routePath.startsWith(`${targetPath}/`)

const sameTarget = (route, target) => {
  if (!target) return false
  if (!samePath(route.path, target.path)) return false
  if (target.query?.section && String(route.query?.section || '') !== target.query.section) return false
  return true
}

const matchesSection = (route, item, group) => {
  if (!samePath(route.path, item.to?.path || group.path) && !samePath(route.path, group.path)) return false
  const section = String(route.query?.section || '')
  return item.matchSections?.includes(section) || section === item.to?.query?.section
}

const matchesItem = (route, item, group) => sameTarget(route, item.to) || matchesSection(route, item, group)

export const quickOptions = navGroups.flatMap((group) =>
  [...group.children, ...(group.hiddenChildren || []).filter((item) => item.quick)].map((item) => ({
    value: `${group.key}:${item.section}`,
    label: `${group.label} / ${item.label}`,
    to: item.to,
  })),
)

export const findNavByRoute = (route) => {
  for (const group of navGroups) {
    const child = group.children.find((item) => matchesItem(route, item, group))
    if (child) return { group, child }
  }

  for (const group of navGroups) {
    const child = group.hiddenChildren.find((item) => matchesItem(route, item, group))
    if (child) return { group, child }
  }

  const group = navGroups.find((item) => samePath(route.path, item.path)) || navGroups[0]
  const section = String(route.query?.section || group.defaultSection)
  const childOptions = [...group.children, ...group.hiddenChildren]
  const child = childOptions.find((item) => item.section === section) || group.children[0]
  return { group, child }
}

export const createNavigationRegistry = (items) => {
  if (!Array.isArray(items)) return navGroups
  return items.map(attachTargets)
}

export const flattenNavigationItems = (items) => {
  const routes = []

  const visit = (item) => {
    if (item.path) routes.push(item)
    ;(item.children || []).forEach(visit)
  }

  items.forEach(visit)
  return routes
}
