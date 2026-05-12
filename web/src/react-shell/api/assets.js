import { del, get, isPermissionError, normalizeList, post, put, redactText } from './http.js'

export const ASSET_BLOCKERS = {
  hostCreate: 'BLOCKED_BY_CONTRACT: 主机新建、编辑、删除、云导入和 Excel 导入缺少 FindX 审计契约。',
  terminal: 'BLOCKED_BY_CONTRACT: 远程终端、文件上传和命令执行缺少会话、审计和回滚契约。',
  monitor: 'BLOCKED_BY_CONTRACT: 主机监控弹窗缺少进程、端口和全量指标 Adapter 契约。',
  agentLifecycle: 'BLOCKED_BY_CONTRACT: FindX Agent 部署、卸载、重启、删除和进度记录缺少生命周期契约。',
  agentPackage: 'BLOCKED_BY_CONTRACT: Agent 能力包、配置下发和数据到达验证缺少生命周期契约。',
}

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== null && value !== undefined),
)

const normalizeRows = value => normalizeList(value)

export const splitText = value => String(value || '')
  .split(/[\n,]/)
  .map(item => item.trim())
  .filter(Boolean)

export const formatAssetError = error => {
  if (isPermissionError(error)) return error.status === 401 ? '登录状态已过期，请重新登录。' : '当前账号没有资产访问权限。'
  if ([404, 405, 501].includes(error?.status)) return `BLOCKED_BY_CONTRACT: ${redactText(error.message || '接口未开放')}`
  return redactText(error?.message || '资产请求失败')
}

export const assetsApi = {
  workspaces: {
    list: () => get('/workspaces').then(normalizeRows),
    create: body => post('/workspaces', body),
    update: (id, body) => put(`/workspaces/${encodeURIComponent(id)}`, body),
    remove: id => del(`/workspaces/${encodeURIComponent(id)}`),
  },
  resourceGroups: {
    list: params => get('/resource-groups', { params: cleanParams(params) }).then(normalizeRows),
    create: body => post('/resource-groups', body),
    update: (id, body) => put(`/resource-groups/${encodeURIComponent(id)}`, body),
    remove: id => del(`/resource-groups/${encodeURIComponent(id)}`),
  },
  hosts: {
    list: params => get('/host-assets', { params: cleanParams(params) }).then(normalizeRows),
    detail: id => get(`/host-assets/${encodeURIComponent(id)}`),
    updateTags: (id, tags) => put(`/host-assets/${encodeURIComponent(id)}/tags`, { tags }),
    bindResourceGroup: (id, resource_group_id) => put(`/host-assets/${encodeURIComponent(id)}/resource-group`, { resource_group_id }),
    bindWorkspace: (id, workspace_id) => put(`/host-assets/${encodeURIComponent(id)}/workspace`, { workspace_id }),
  },
  agents: {
    list: () => get('/findx-agents').then(normalizeRows),
  },
}
