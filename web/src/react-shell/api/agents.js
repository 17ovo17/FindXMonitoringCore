import { del, get, isPermissionError, normalizeList, post, redactText } from './http.js'

export const AGENT_BLOCKERS = {
  packageLifecycle: 'BLOCKED_BY_CONTRACT: 能力包下载、内置包仓库、签名校验、发布、停用和删除缺少 Agent Adapter 契约。',
  installLifecycle: 'BLOCKED_BY_CONTRACT: 本机安装、远程安装、Kubernetes 安装、执行回执和审计记录缺少 Agent Adapter 契约。',
  configLifecycle: 'BLOCKED_BY_CONTRACT: 配置模板保存、远程 writer、reload 回执、drift 检测、灰度/全量下发、回滚、审计记录和 Evidence Chain 缺少 Agent Adapter 契约。',
  heartbeat: 'BLOCKED_BY_CONTRACT: 心跳详情、丢包检测、版本漂移、进程状态和服务注册缺少 Agent Adapter 契约。',
  dataArrival: 'BLOCKED_BY_CONTRACT: 心跳、指标、日志、链路、性能分析、巡检、拓扑、前端体验和网关链路数据到达验证缺少 Agent Adapter 契约。',
  traceLinkage: 'BLOCKED_BY_CONTRACT: 服务目录覆盖率、链路详情反查 Agent 状态、拓扑节点安装证据缺少 APM 与 Agent 联动契约。',
}

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== null && value !== undefined),
)

export const formatAgentError = error => {
  if (isPermissionError(error)) return error.status === 401 ? '登录状态已过期，请重新登录。' : '当前账号没有 Agent 管理权限。'
  if ([404, 405, 501].includes(error?.status)) return `BLOCKED_BY_CONTRACT: ${redactText(error.message || '接口未开放')}`
  return redactText(error?.message || 'Agent 管理请求失败')
}

export const agentApi = {
  list: params => get('/findx-agents', { params: cleanParams(params) }).then(normalizeList),
  packages: () => get('/findx-agents/packages').then(normalizeList),
  lifecycle: () => get('/findx-agents/lifecycle'),
  createInstallPlan: body => post('/findx-agents/install-plans', body),
  listInstallPlans: () => get('/findx-agents/install-plans').then(normalizeList),
  getInstallPlan: id => get('/findx-agents/install-plans', { params: cleanParams({ id }) }),
  listInstallExecutions: () => get('/findx-agents/install-executions').then(normalizeList),
  getInstallExecution: id => get('/findx-agents/install-executions', { params: cleanParams({ id }) }),
  packageAction: (action, packageId = 'all') => post('/findx-agents/tasks', {
    action,
    package_id: packageId,
    target_ids: ['package-repository-control-plane'],
    metadata: {
      package_id: packageId,
      target_os: 'control-plane',
      transport: 'local-control-plane',
      audit_ref: 'package-action-audit',
    },
  }),
  templates: () => get('/findx-agents/config-templates').then(normalizeList),
  rolloutConfig: body => post('/findx-agents/config-rollouts', body),
  listConfigRollouts: () => get('/findx-agents/config-rollouts').then(normalizeList),
  getConfigRollout: id => get('/findx-agents/config-rollouts', { params: cleanParams({ id }) }),
  listTasks: () => get('/findx-agents/tasks').then(normalizeList),
  getTask: id => get('/findx-agents/tasks', { params: cleanParams({ id }) }),
  dataArrival: () => get('/findx-agents/data-arrival').then(normalizeList),
  dataArrivalEvidence: () => get('/findx-agents/data-arrival/evidence').then(normalizeList),
  task: body => post('/findx-agents/tasks', body),

  // P5: Agent lifecycle package management
  listAgentPackagesV2: () => get('/agent-packages').then(normalizeList),
  registerAgentPackage: body => post('/agent-packages', body),
  deleteAgentPackage: id => del(`/agent-packages/${id}`),

  // P5: Agent lifecycle actions
  installAgent: (id, body) => post(`/findx-agents/${id}/install`, body),
  upgradeAgent: (id, body) => post(`/findx-agents/${id}/upgrade`, body),
  rollbackAgent: id => post(`/findx-agents/${id}/rollback`),
  uninstallAgent: id => post(`/findx-agents/${id}/uninstall`),
  configPushAgent: (id, body) => post(`/findx-agents/${id}/config-push`, body),
  evidenceChain: id => get(`/findx-agents/${id}/evidence-chain`),
}
