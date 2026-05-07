import { redactText, safeJson } from '../../api/notifications'

export const supportedChannelTypes = new Set(['dingtalk', 'feishu', 'wecom'])

export const channelTypes = [
  { ident: 'dingtalk', label: '钉钉', icon: '钉', supported: true },
  { ident: 'feishu', label: '飞书', icon: '飞', supported: true },
  { ident: 'wecom', label: '企业微信', icon: '企', supported: true },
  { ident: 'callback', label: 'Webhook', icon: 'H', supported: false },
  { ident: 'email', label: 'Email', icon: 'M', supported: false },
  { ident: 'pagerduty', label: 'PagerDuty', icon: 'P', supported: false },
  { ident: 'script', label: 'Script', icon: 'S', supported: false },
  { ident: 'telegram', label: 'Telegram', icon: 'T', supported: false },
  { ident: 'sms', label: 'SMS', icon: '短', supported: false },
  { ident: 'voice', label: 'Voice', icon: '电', supported: false },
]

export const blockedContracts = {
  'rule-list': '通知规则 contract 未暴露，需接入规则列表、启停、克隆、删除、详情和测试投递。',
  'rule-create': '通知规则新增/编辑 contract 未暴露，需接入事件条件、通知配置、接收对象、时间窗口和审计。',
  'rule-toggle': '通知规则启停 contract 未暴露，不能只在前端切换静态状态。',
  'channel-type': '该通知媒介的投递 contract 未接入当前后端，不能伪装为可用媒介。',
  'template-list': '消息模板 contract 未暴露，需接入模板列表、详情、编辑、预览、克隆和删除。',
  'template-save': '消息模板保存/预览 contract 未暴露，不能静态保存模板内容。',
}

export const blockedPayload = (action, context = {}) => safeJson({
  action,
  context,
  status: 'BLOCKED_BY_CONTRACT',
  next_contract_needed: blockedContracts[action] || '后端 contract 未暴露',
})

export const normalizeChannel = raw => ({
  raw,
  id: String(raw?.id ?? ''),
  type: raw?.type || raw?.ident || 'dingtalk',
  name: redactText(raw?.name || '-'),
  endpoint: redactText(raw?.endpoint || raw?.webhook || ''),
  receiver: redactText(raw?.receiver || ''),
  enabled: raw?.enabled !== false,
  updatedBy: redactText(raw?.updated_by || raw?.updatedBy || 'root'),
  updatedAt: raw?.updated_at || raw?.updatedAt || raw?.created_at || '',
})

export const formatDate = value => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return redactText(value)
  return date.toLocaleString('zh-CN', { hour12: false })
}

export const filterChannels = (channels, filter) => channels.filter(item => {
  const text = `${item.name} ${item.type} ${item.receiver}`.toLowerCase()
  if (filter.query && !text.includes(filter.query.toLowerCase())) return false
  if (filter.status && String(item.enabled) !== filter.status) return false
  if (filter.types?.length && !filter.types.includes(item.type)) return false
  return true
})

export const downloadText = (filename, text, type = 'text/plain;charset=utf-8') => {
  const blob = new Blob([text], { type })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export const typeLabel = type => channelTypes.find(item => item.ident === type)?.label || redactText(type || '-')
export const typeIcon = type => channelTypes.find(item => item.ident === type)?.icon || '?'
