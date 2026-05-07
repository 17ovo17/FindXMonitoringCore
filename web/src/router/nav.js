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
      { section: 'agents', label: 'FindX Agent' },
    ],
  },
  {
    key: 'integrations',
    label: '集成中心',
    icon: 'Connection',
    path: '/integrations',
    defaultSection: 'overview',
    children: [
      { section: 'overview', label: '集成总览' },
      { section: 'collectors', label: '采集接入' },
      { section: 'templates', label: '模板导入' },
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
    ],
  },
  {
    key: 'dashboards',
    label: '仪表盘',
    icon: 'TrendCharts',
    path: '/dashboards',
    defaultSection: 'list',
    children: [
      { section: 'list', label: '仪表盘列表' },
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
      { section: 'events', label: '事件中心' },
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
    ],
  },
  {
    key: 'tracing',
    label: '链路监控',
    icon: 'Share',
    path: '/tracing',
    defaultSection: 'overview',
    children: [
      { section: 'overview', label: '链路总览' },
      { section: 'services', label: '服务拓扑' },
      { section: 'traces', label: 'Trace 检索' },
    ],
  },
  {
    key: 'logs',
    label: '日志中心',
    icon: 'Document',
    path: '/logs',
    defaultSection: 'overview',
    children: [
      { section: 'overview', label: '日志总览' },
      { section: 'query', label: '日志检索' },
      { section: 'pipelines', label: '接入管道' },
    ],
  },
  {
    key: 'agents',
    label: 'Agent 管理中心',
    icon: 'Cpu',
    path: '/agents',
    defaultSection: 'overview',
    children: [
      { section: 'overview', label: 'FindX Agent' },
      { section: 'install', label: '安装计划' },
      { section: 'command', label: '生成命令' },
      { section: 'ops', label: '远程运维' },
      { section: 'chat', label: '探针对话' },
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
    ],
  },
  {
    key: 'org',
    label: '组织权限',
    icon: 'User',
    path: '/org',
    defaultSection: 'users',
    children: [
      { section: 'users', label: '用户管理' },
      { section: 'teams', label: '团队组织' },
      { section: 'roles', label: '角色管理' },
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
