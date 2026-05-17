const to = (path, section, extraQuery = {}) => ({ path, query: { section, ...extraQuery } })
const agentMergedSections = new Set(['install', 'templates', 'heartbeat', 'data-arrival', 'config', 'plugins'])

const targetFor = (group, item) => {
  if (item.kind === 'label') return item.to
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

const attachItemTargets = (group, item) => {
  const next = {
    ...item,
    to: targetFor(group, item),
  }
  if (Array.isArray(item.children)) {
    next.children = item.children.map((child) => attachItemTargets(group, child))
  }
  return next
}

const attachTargets = (group) => {
  const children = group.children.map((item) => attachItemTargets(group, item))
  const hiddenChildren = (group.hiddenChildren || []).map((item) => attachItemTargets(group, item))

  return {
    ...group,
    children,
    hiddenChildren,
    to: group.to || children[0]?.to || to(group.path, group.defaultSection),
  }
}

export const navGroups = [
  {
    key: 'ai-sre',
    label: 'AI助手',
    icon: 'ai-assistant',
    path: '/aiops',
    defaultSection: 'diagnosis',
    children: [
      { section: 'diagnosis', label: '诊断会话' },
      { section: 'health', label: '健康检查' },
      { section: 'report', label: '复盘报告' },
      { section: 'evidence', label: 'Evidence Chain' },
      { section: 'remediation', label: '自动修复' },
    ],
  },
  {
    key: 'overview',
    label: '监控总览',
    icon: 'monitoring',
    path: '/overview',
    defaultSection: 'dashboard',
    children: [
      { section: 'dashboard', label: '全局概览' },
      {
        section: 'metrics',
        label: '指标查询',
        to: to('/query', 'metrics'),
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
    label: '告警中心',
    icon: 'alert',
    path: '/alerts',
    defaultSection: 'events',
    children: [
      { section: 'events', label: '告警事件' },
      { section: 'rules', label: '规则管理' },
      { section: 'mutes', label: '告警屏蔽' },
      { section: 'subscriptions', label: '告警订阅' },
      { section: 'history-events', label: '历史事件' },
      { section: 'tracing-alarms', label: '链路告警' },
      { section: 'event-pipelines', label: '事件流水线' },
      { section: 'notify-rules', label: '通知规则', to: to('/notifications', 'rules') },
      { section: 'notify-channels', label: '通知媒介', to: to('/notifications', 'channels') },
      { section: 'notify-templates', label: '消息模板', to: to('/notifications', 'templates') },
    ],
  },
  {
    key: 'tracing',
    label: '链路监控',
    icon: 'trace',
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
    icon: 'logs',
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
    key: 'asset-center',
    label: 'CMDB',
    icon: 'cmdb',
    path: '/assets',
    defaultSection: 'overview',
    children: [
      { section: 'overview', label: '概况' },
      { section: 'search', label: '全文检索', to: to('/assets', 'overview', { focus: 'search' }) },
      {
        section: 'resource-management',
        label: '资源管理',
        children: [
          { section: 'hosts', label: '资源列表', to: to('/assets', 'hosts') },
          { section: 'business', label: '业务视图', to: to('/assets', 'business') },
          { section: 'room-view', label: '机房视图', to: to('/assets', 'room-view') },
          { section: 'recycle-bin', label: '回收站', to: to('/assets', 'recycle-bin') },
        ],
      },
      {
        section: 'model-config',
        label: '模型配置',
        children: [
          { section: 'models', label: '模型管理', to: to('/assets', 'models') },
          { section: 'model-relations', label: '关联类型', to: to('/assets', 'model-relations') },
          { section: 'attribute-units', label: '属性单位', to: to('/assets', 'attribute-units') },
        ],
      },
      {
        section: 'discovery-management',
        label: '发现管理',
        children: [
          { section: 'auto-discovery', label: '自动发现', to: to('/assets', 'auto-discovery') },
          { section: 'auto-mapping', label: '自动化映射', to: to('/assets', 'auto-mapping') },
          { section: 'discovery-records', label: '执行记录', to: to('/assets', 'discovery-records') },
        ],
      },
      {
        section: 'resource-approval',
        label: '资源审批',
        children: [
          { section: 'approval-mine', label: '我的申请', to: to('/assets', 'approval-mine') },
          { section: 'approval-todo', label: '我的待办', to: to('/assets', 'approval-todo') },
          { section: 'approval-archive', label: '已归档', to: to('/assets', 'approval-archive') },
        ],
      },
      {
        section: 'resource-reports',
        label: '资源报表',
        children: [
          { section: 'resource-stats', label: '资源统计', to: to('/assets', 'resource-stats') },
          { section: 'change-stats', label: '变更统计', to: to('/assets', 'change-stats') },
          { section: 'model-change', label: '模型变更', to: to('/assets', 'model-change') },
          { section: 'instance-change-top', label: '实例变更TOP', to: to('/assets', 'instance-change-top') },
          { section: 'custom-report', label: '自定义报表', to: to('/assets', 'custom-report') },
          { section: 'cloud-bill', label: '云平台账单', to: to('/assets', 'cloud-bill') },
        ],
      },
      {
        section: 'asset-consumption',
        label: '资产消费',
        children: [
          { section: 'compliance-check', label: '合规性检查', to: to('/assets', 'compliance-check') },
          { section: 'compliance-stats', label: '合规性统计', to: to('/assets', 'compliance-stats') },
          { section: 'change-notice', label: '变更通知', to: to('/assets', 'change-notice') },
          { section: 'spare-management', label: '备件管理', to: to('/assets', 'spare-management') },
          { section: 'inspection-config', label: '巡检配置', to: to('/assets', 'inspection-config') },
          { section: 'relationship-query', label: '关系查询', to: to('/assets', 'relationship-query') },
          { section: 'event-subscription', label: '事件订阅', to: to('/assets', 'event-subscription') },
        ],
      },
      {
        section: 'audit-records',
        label: '审计记录',
        children: [
          { section: 'notice-records', label: '通知记录', to: to('/assets', 'notice-records') },
          { section: 'change-records', label: '变更记录', to: to('/assets', 'change-records') },
          { section: 'subscription-records', label: '订阅记录', to: to('/assets', 'subscription-records') },
        ],
      },
    ],
    hiddenChildren: [
      { section: 'model-detail', label: '模型详情', to: to('/assets', 'model-detail') },
      { section: 'instance-detail', label: '实例详情', to: to('/assets', 'instance-detail') },
      { section: 'topology', label: '实例拓扑', to: to('/assets', 'topology') },
      { section: 'business', label: '业务空间', to: to('/assets', 'business') },
      { section: 'cmdb', label: 'CMDB', to: to('/assets', 'cmdb') },
      { section: 'hosts', label: '主机资产', to: to('/assets', 'hosts') },
      { section: 'databases', label: '数据库资产', to: to('/assets', 'databases') },
      { section: 'deploy-tasks', label: '部署任务', to: to('/assets', 'deploy-tasks') },
      { section: 'packages', label: '能力包', to: to('/assets', 'packages') },
    ],
  },
  {
    key: 'agents',
    label: 'Agent管理',
    icon: 'agent',
    path: '/agents',
    defaultSection: 'hosts',
    children: [
      { section: 'hosts', label: '主机列表' },
      { section: 'packages', label: '能力包' },
      { section: 'plugins', label: '插件目录' },
      { section: 'environment', label: '环境适配' },
    ],
  },
  {
    key: 'knowledge',
    label: '知识库',
    icon: 'knowledge',
    path: '/knowledge',
    defaultSection: 'docs',
    children: [
      { section: 'docs', label: '知识管理', to: to('/aiops', 'knowledge') },
      { section: 'runbooks', label: '运维手册', to: to('/aiops', 'runbooks') },
    ],
  },
  {
    key: 'integrations',
    label: '集成中心',
    icon: 'integration',
    path: '/integrations',
    defaultSection: 'datasources',
    children: [
      { section: 'datasources', label: '数据源' },
      { section: 'templates', label: '模板中心' },
      { section: 'systems', label: '系统集成' },
    ],
  },
  {
    key: 'business-probes',
    label: '业务拨测',
    icon: 'probe',
    path: '/status',
    defaultSection: 'public',
    children: [
      { section: 'public', label: '状态页' },
      { section: 'config', label: '拨测配置' },
    ],
    hiddenChildren: [
      { section: 'incidents', label: '事故维护', to: to('/status', 'incidents') },
    ],
  },
  {
    key: 'organization',
    label: '人员组织',
    icon: 'org',
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
    icon: 'settings',
    path: '/platform',
    defaultSection: 'models',
    children: [
      { section: 'models', label: 'AI 模型配置' },
      { section: 'mcp', label: 'MCP 服务' },
      { section: 'site', label: '站点设置' },
      { section: 'variables', label: '变量设置' },
      { section: 'sso', label: '单点登录' },
      { section: 'alerting-engines', label: '告警引擎' },
      { section: 'sandbox', label: 'AI 执行沙箱' },
      { section: 'health', label: '运行自检' },
      { section: 'audit', label: '审计日志' },
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

const matchesItem = (route, item, group) => item.kind !== 'label' && (sameTarget(route, item.to) || matchesSection(route, item, group))

const findChildByRoute = (route, items, group) => {
  for (const item of items) {
    const child = findChildByRoute(route, item.children || [], group)
    if (child) return child
  }
  return items.find((item) => matchesItem(route, item, group))
}

const flattenNavItems = (items = []) => items.flatMap((item) => (
  item.children?.length ? flattenNavItems(item.children) : item.kind === 'label' ? [] : [item]
))

export const quickOptions = navGroups.flatMap((group) =>
  [...flattenNavItems(group.children), ...(group.hiddenChildren || []).filter((item) => item.quick)].map((item) => ({
    value: `${group.key}:${item.section}`,
    label: `${group.label} / ${item.label}`,
    to: item.to,
  })),
)

export const findNavByRoute = (route) => {
  for (const group of navGroups) {
    const child = findChildByRoute(route, group.children, group)
    if (child) return { group, child }
  }

  for (const group of navGroups) {
    const child = findChildByRoute(route, group.hiddenChildren, group)
    if (child) return { group, child }
  }

  const group = navGroups.find((item) => samePath(route.path, item.path)) || navGroups[0]
  const section = String(route.query?.section || group.defaultSection)
  const childOptions = flattenNavItems([...group.children, ...group.hiddenChildren])
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
