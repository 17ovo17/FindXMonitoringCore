export const sections = [
  { value: 'overview', label: '资产总览', desc: '统一查看业务空间、资源组、主机资产和 FindX Agent 存活状态。' },
  { value: 'models', label: '对象建模', desc: '管理 CMDB 模型分类、模型定义和属性配置。' },
  { value: 'instances', label: '实例管理', desc: '管理 CMDB 资产实例数据。' },
  { value: 'business', label: '业务组', desc: '按业务边界管理主机、端点、负责人和状态。' },
  { value: 'cmdb', label: 'CMDB', desc: '按资源组树和主机表管理资产关系。' },
  { value: 'hosts', label: '主机资产', desc: '按业务空间、资源组、标签、在线状态筛选和绑定主机。' },
  { value: 'resource-groups', label: '资源组', desc: '维护分层资源组树和主机关联边界。' },
  { value: 'agents', label: 'FindX Agent', desc: '查看 Agent 在线、心跳、版本、能力和生命周期状态。' },
  { value: 'model-detail', label: '模型详情', desc: '查看和编辑模型属性定义。' },
]

export const sectionSet = new Set(sections.map(item => item.value))

export const fmtTime = value => {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString('zh-CN', { hour12: false })
}

export const normalizeTags = value => {
  if (Array.isArray(value)) return value.filter(Boolean)
  return String(value || '').split(/[\n,]/).map(item => item.trim()).filter(Boolean)
}

const mojibakePattern = /[\uFFFD]|\?\?\?/

export const displayText = (value, fallback = '-') => {
  if (value === null || value === undefined || value === '') return fallback
  const text = String(value)
  return mojibakePattern.test(text) ? '内容不可读' : text
}

export const displayTags = value => normalizeTags(value).map(tag => displayText(tag)).filter(Boolean)
export const safeRowText = row => Object.values(row || {}).map(value => displayText(value, '')).join(' ').toLowerCase()

export const hostKey = row => row?.host_id || row?.id || row?.ident || ''
export const hostName = row => displayText(row?.hostname || row?.host_name || row?.name || row?.ident || row?.id)
export const hostIp = row => displayText(Array.isArray(row?.ip_list) ? row.ip_list.join(', ') : (row?.ip || row?.host_ip || row?.address || '-'))
export const isHostOnline = row => row?.online === true || row?.status === 'online' || row?.agent_status === 'online'
export const agentKey = row => row?.id || row?.ident || `${row?.ip || ''}-${row?.hostname || ''}`
export const agentOnline = row => row?.online === true || row?.status === 'online'
export const rowText = safeRowText

export const endpointText = endpoint => {
  if (!endpoint) return ''
  const port = endpoint.port ? `:${endpoint.port}` : ''
  return [endpoint.ip ? `${endpoint.ip}${port}` : '', endpoint.service_name || endpoint.service || '', endpoint.protocol || ''].map(item => displayText(item, '')).filter(Boolean).join(' ')
}

export const parseEndpoints = text => String(text || '').split('\n').map(line => {
  const parts = line.trim().split(/\s+/).filter(Boolean)
  if (!parts.length) return null
  const match = parts[0].match(/^(.+):(\d{1,5})$/)
  return {
    ip: match ? match[1] : parts[0],
    port: match ? Number(match[2]) : 0,
    service_name: parts[1] || '',
    protocol: parts[2] || '',
  }
}).filter(Boolean)
