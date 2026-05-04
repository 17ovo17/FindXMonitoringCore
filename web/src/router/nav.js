export const navGroups = [
  {
    key: 'assets',
    label: '基础设施',
    icon: 'FolderOpened',
    path: '/assets',
    defaultSection: 'overview',
    children: [
      { section: 'overview', label: '资产中心' },
      { section: 'business', label: '业务空间' },
      { section: 'hosts', label: '主机资产' },
      { section: 'agents', label: '探针与采集' },
      { section: 'installs', label: '接入任务' },
      { section: 'credentials', label: '凭证资产' },
    ],
  },
  {
    key: 'query',
    label: '数据查询',
    icon: 'DataLine',
    path: '/query',
    defaultSection: 'datasources',
    children: [
      { section: 'datasources', label: '数据源' },
      { section: 'metrics', label: '指标查询' },
      { section: 'logs', label: '日志查询' },
      { section: 'traces', label: 'Trace 查询' },
      { section: 'metric-mapping', label: '指标映射' },
    ],
  },
  {
    key: 'dashboards',
    label: '监控仪表盘',
    icon: 'TrendCharts',
    path: '/dashboards',
    defaultSection: 'list',
    children: [
      { section: 'list', label: '仪表盘列表' },
      { section: 'templates', label: '仪表盘模板' },
      { section: 'shares', label: '图表分享' },
    ],
  },
  {
    key: 'alerts',
    label: '告警',
    icon: 'Bell',
    path: '/alerts',
    defaultSection: 'events',
    children: [
      { section: 'rules', label: '告警规则' },
      { section: 'recording-rules', label: '记录规则' },
      { section: 'events', label: '事件中心' },
      { section: 'aggr-views', label: '聚合视图' },
      { section: 'mutes', label: '告警静默' },
      { section: 'subscribes', label: '告警订阅' },
      { section: 'pipeline', label: '事件流水线' },
    ],
  },
  {
    key: 'notifications',
    label: '通知',
    icon: 'Message',
    path: '/notifications',
    defaultSection: 'rules',
    children: [
      { section: 'rules', label: '通知规则' },
      { section: 'channels', label: '通知媒介' },
      { section: 'templates', label: '消息模板' },
      { section: 'records', label: '通知记录' },
      { section: 'configs', label: '通知配置' },
    ],
  },
  {
    key: 'aiops',
    label: 'AI SRE',
    icon: 'ChatDotRound',
    path: '/aiops',
    defaultSection: 'diagnosis',
    children: [
      { section: 'diagnosis', label: 'AI 问答' },
      { section: 'knowledge', label: '知识库' },
      { section: 'workflow', label: '工作流' },
      { section: 'remediation', label: '自动修复' },
      { section: 'asset-chat', label: '单机对话' },
    ],
  },
  {
    key: 'org',
    label: '人员组织',
    icon: 'User',
    path: '/org',
    defaultSection: 'users',
    children: [
      { section: 'users', label: '用户管理' },
      { section: 'teams', label: '团队组织' },
      { section: 'roles', label: '角色管理' },
      { section: 'permissions', label: '操作权限' },
      { section: 'tokens', label: 'Token 管理' },
    ],
  },
  {
    key: 'platform',
    label: '平台治理',
    icon: 'Setting',
    path: '/platform',
    defaultSection: 'models',
    children: [
      { section: 'models', label: '模型配置' },
      { section: 'settings', label: '系统配置' },
      { section: 'health', label: '平台运行自检' },
      { section: 'audit', label: '审计日志' },
    ],
  },
]

navGroups.forEach(group => {
  group.to = { path: group.path, query: { section: group.defaultSection } }
})

export const quickOptions = navGroups.flatMap(group =>
  group.children.map(item => ({
    value: `${group.key}:${item.section}`,
    label: `${group.label} / ${item.label}`,
    to: { path: group.path, query: { section: item.section } },
  }))
)

export const findNavByRoute = route => {
  const group = navGroups.find(item => route.path === item.path || route.path.startsWith(`${item.path}/`)) || navGroups[0]
  const section = String(route.query?.section || group.defaultSection)
  const child = group.children.find(item => item.section === section) || group.children[0]
  return { group, child }
}
