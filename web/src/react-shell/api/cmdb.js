import { get, post, put, del } from './http.js'

export const CMDB_EXECUTION_BLOCKERS = {
  terminal: 'BLOCKED_BY_CONTRACT: 远程终端缺少会话授权、审计记录、执行回执和安全隔离契约，当前只能显示阻断状态。',
  upload: 'BLOCKED_BY_CONTRACT: 文件上传缺少 SFTP/执行器、目标路径校验、审计记录、回滚和执行回执契约，当前只能显示阻断状态。',
  exec: 'BLOCKED_BY_CONTRACT: 命令执行缺少远程执行器、权限审批、审计记录、超时回执和输出脱敏契约，当前只能显示阻断状态。',
  deploy: 'BLOCKED_BY_CONTRACT: 部署任务缺少执行器、任务回执、日志轮询、回滚和 Evidence Chain 契约，当前只能显示阻断状态。',
  databaseTest: 'BLOCKED_BY_CONTRACT: 数据库连接测试缺少凭据引用、执行器、审计记录和脱敏错误模型，当前只能显示阻断状态。',
}

export const isCmdbContractBlocked = (error) => {
  const message = String(error?.message || error || '')
  return error?.status === 409 || /BLOCKED_BY_CONTRACT|HTTP 409/i.test(message)
}

export const cmdbContractMessage = (error, fallback) => {
  const message = String(error?.message || '')
  if (/BLOCKED_BY_CONTRACT/i.test(message)) return message.replace(/^HTTP 409:\s*/i, '')
  return fallback
}

export const isCmdbBlockedRecord = (record) => {
  const code = String(record?.code || record?.status || record?.error_code || '').toLowerCase()
  return code === 'blocked_by_contract' || code === 'blocked'
}

export const cmdbBlockedRecordMessage = (record, fallback) => {
  const missing = Array.isArray(record?.missing_contracts) ? record.missing_contracts.join('、') : ''
  if (record?.contract_id) {
    return `${fallback} contract_id=${record.contract_id}${missing ? `；缺口：${missing}` : ''}`
  }
  return fallback
}

export const cmdbApi = {
  tree: () => get('/cmdb/tree'),
  objects: {
    list: (categoryId) => get('/cmdb/objects', { params: categoryId ? { category_id: categoryId } : {} }),
    get: (id) => get(`/cmdb/objects/${id}`),
    create: (body) => post('/cmdb/objects', body),
    update: (id, body) => put(`/cmdb/objects/${id}`, body),
    remove: (id) => del(`/cmdb/objects/${id}`),
  },
  attributes: {
    list: (objectId) => get(`/cmdb/objects/${objectId}/attributes`),
    create: (objectId, body) => post(`/cmdb/objects/${objectId}/attributes`, body),
    update: (id, body) => put(`/cmdb/attributes/${id}`, body),
    remove: (id) => del(`/cmdb/attributes/${id}`),
  },
  instances: {
    list: (objectId, params) => get(`/cmdb/objects/${objectId}/instances`, { params }),
    get: (id) => get(`/cmdb/instances/${id}`),
    create: (objectId, body) => post(`/cmdb/objects/${objectId}/instances`, body),
    update: (id, body) => put(`/cmdb/instances/${id}`, body),
    remove: (id) => del(`/cmdb/instances/${id}`),
  },
  // C03: SSH 终端
  terminal: {
    wsUrl: (hostId) => {
      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      return `${proto}//${window.location.host}/api/v1/cmdb/hosts/${encodeURIComponent(hostId)}/terminal`
    },
  },
  // C04: 文件上传
  upload: (hostId, formData) => post(`/cmdb/hosts/${encodeURIComponent(hostId)}/upload`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000,
  }),
  // C05: 命令执行
  exec: (hostId, body) => post(`/cmdb/hosts/${encodeURIComponent(hostId)}/exec`, body),
  // C07: 数据库资产
  databases: {
    list: (params) => get('/cmdb/databases', { params }),
    create: (body) => post('/cmdb/databases', body),
    get: (id) => get(`/cmdb/databases/${encodeURIComponent(id)}`),
    remove: (id) => del(`/cmdb/databases/${encodeURIComponent(id)}`),
    test: (id) => post(`/cmdb/databases/${encodeURIComponent(id)}/test`),
  },
  // C09: 部署任务
  deployTasks: {
    list: () => get('/cmdb/deploy-tasks'),
    create: (body) => post('/cmdb/deploy-tasks', body),
    get: (id) => get(`/cmdb/deploy-tasks/${encodeURIComponent(id)}`),
  },
  // C10: 导入
  import: {
    excel: (formData) => post('/cmdb/import/excel', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),
    confirm: (body) => post('/cmdb/import/confirm', body),
    cloud: (body) => post('/cmdb/import/cloud', body),
  },
}
