export const TEST_BATCH_ID = 'aiw-a3f7b2c1'

export const channelTypes = [
  { value: 'console', label: '平台内通知' },
  { value: 'webhook', label: 'Webhook' },
  { value: 'email', label: 'Email' },
  { value: 'flashduty', label: 'Flashduty' },
  { value: 'pagerduty', label: 'PagerDuty' },
  { value: 'dingtalk', label: '钉钉' },
  { value: 'feishu', label: '飞书' },
  { value: 'wecom', label: '企业微信' },
]

export const channelTypeLabel = type => channelTypes.find(item => item.value === type)?.label || type || '未知渠道'
export const formatTime = value => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'
export const isBuiltinChannel = channel => ['console', 'webhook', 'flashduty', 'pagerduty'].includes(channel.id)
export const formatMembers = members => (members || []).join('、') || '未配置成员'

export const displayEndpoint = channel => {
  if (channel.type === 'console') return '平台内留痕'
  return channel.endpoint || channel.webhook || '未配置 Endpoint'
}

export const normalizeChannels = items => {
  const defaults = [
    { id: 'console', name: 'Console', type: 'console', receiver: '平台值班人员', enabled: true, endpoint: '' },
    { id: 'webhook', name: 'Webhook', type: 'webhook', receiver: '', enabled: false, endpoint: '' },
    { id: 'flashduty', name: 'Flashduty', type: 'flashduty', receiver: '', enabled: false, endpoint: '' },
    { id: 'pagerduty', name: 'PagerDuty', type: 'pagerduty', receiver: '', enabled: false, endpoint: '' },
  ]
  const byID = new Map(defaults.map(item => [item.id, item]))
  ;(items || []).forEach(item => byID.set(item.id || `${item.type}-${item.name}`, { ...byID.get(item.id), ...item }))
  return [...byID.values()]
}
