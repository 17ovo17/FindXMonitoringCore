import { del, get, normalizeList, post, put, redactText } from './http.js'

export const ORG_BLOCKERS = {
  create: 'PENDING: 缺少该 section 的创建、输入校验、权限点和审计契约，已禁止保存。',
  edit: 'PENDING: 缺少该 section 的单条详情或更新契约，已禁止保存。',
  delete: 'PENDING: 缺少该 section 的删除、幂等和审计契约，已禁止删除。',
  members: 'PENDING: 缺少成员增删、批量选择和回滚契约，已禁止成员变更。',
  status: 'PENDING: 缺少启停状态、权限点和审计契约，已禁止状态变更。',
  roleOps: 'PENDING: 缺少角色操作点保存契约时仅展示权限矩阵，不执行保存。',
}

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== undefined && value !== null),
)

const contractBlocked = (error, label) => {
  if ([404, 405, 501].includes(error?.status)) {
    const blocked = new Error(`PENDING: ${label} 未接入或未开放。`)
    blocked.status = error.status
    blocked.blocked = true
    return blocked
  }
  return error
}

const normalizeRows = (data) => {
  const rows = normalizeList(data)
  if (rows.length) return rows
  if (Array.isArray(data?.dat)) return data.dat
  if (Array.isArray(data?.data?.items)) return data.data.items
  return []
}

const callList = async (url, params, label) => {
  try {
    return {
      rows: normalizeRows(await get(url, { params: cleanParams(params) })),
      source: `GET /api/v1${url}`,
      blocked: '',
    }
  } catch (error) {
    throw contractBlocked(error, label)
  }
}

export const orgApi = {
  users: {
    list: (params) => callList('/org/users', params, '人员用户列表契约'),
    get: (id) => get(`/org/users/${encodeURIComponent(id)}`),
    create: (body) => post('/org/users', body),
    update: (id, body) => put(`/org/users/${encodeURIComponent(id)}`, body),
    remove: (id) => del(`/org/users/${encodeURIComponent(id)}`),
    resetPassword: (id, body) => put(`/org/users/${encodeURIComponent(id)}/password`, body),
    setStatus: (id, body) => put(`/org/users/${encodeURIComponent(id)}/status`, body),
  },
  teams: {
    list: (params) => callList('/org/teams', params, '团队组织列表契约'),
    get: (id) => get(`/org/teams/${encodeURIComponent(id)}`),
    create: (body) => post('/org/teams', body),
    update: (id, body) => put(`/org/teams/${encodeURIComponent(id)}`, body),
    remove: (id) => del(`/org/teams/${encodeURIComponent(id)}`),
    addMembers: (id, body) => post(`/org/teams/${encodeURIComponent(id)}/members`, body),
    removeMember: (id, userId) => del(`/org/teams/${encodeURIComponent(id)}/members/${encodeURIComponent(userId)}`),
  },
  business: {
    list: (params) => callList('/org/business-groups', params, '业务组列表契约'),
    get: (id) => get(`/org/business-groups/${encodeURIComponent(id)}`),
    create: (body) => post('/org/business-groups', body),
    update: (id, body) => put(`/org/business-groups/${encodeURIComponent(id)}`, body),
    remove: (id) => del(`/org/business-groups/${encodeURIComponent(id)}`),
    addTeams: (id, body) => post(`/org/business-groups/${encodeURIComponent(id)}/teams`, body),
    removeTeam: (id, teamId) => del(`/org/business-groups/${encodeURIComponent(id)}/teams/${encodeURIComponent(teamId)}`),
  },
  roles: {
    list: (params) => callList('/org/roles', params, '角色列表契约'),
    operations: () => callList('/org/permissions/operations', {}, '权限操作点契约'),
    roleOperations: (id) => get(`/org/roles/${encodeURIComponent(id)}/operations`),
    create: (body) => post('/org/roles', body),
    update: (id, body) => put(`/org/roles/${encodeURIComponent(id)}`, body),
    remove: (id) => del(`/org/roles/${encodeURIComponent(id)}`),
    saveOperations: (id, ops) => put(`/org/roles/${encodeURIComponent(id)}/operations`, ops),
  },
  async audit(params) {
    try {
      return {
        rows: normalizeRows(await get('/audit/events', { params: cleanParams(params) })),
        source: 'GET /api/v1/audit/events',
      }
    } catch (error) {
      if (error?.status && ![404, 405, 501].includes(error.status)) throw error
      return {
        rows: normalizeRows(await get('/monitor/audit-logs', { params: cleanParams(params) })),
        source: 'fallback GET /api/v1/monitor/audit-logs',
      }
    }
  },
}

export const formatOrgError = (error) => redactText(error?.message || '请求失败')
