import { redactText, safeJson } from '../../api/http.js'

export const notificationSections = [
  { value: 'rules', label: '通知规则' },
  { value: 'channels', label: '通知媒介' },
  { value: 'templates', label: '消息模板' },
]

export const supportedChannelTypes = new Set(['dingtalk', 'feishu', 'wecom', 'email', 'webhook', 'telegram', 'slack'])

export const channelTypes = [
  { ident: 'dingtalk', label: '钉钉', supported: true },
  { ident: 'feishu', label: '飞书', supported: true },
  { ident: 'wecom', label: '企业微信', supported: true },
  { ident: 'email', label: 'Email', supported: true },
  { ident: 'webhook', label: 'Webhook', supported: true },
  { ident: 'telegram', label: 'Telegram', supported: true },
  { ident: 'slack', label: 'Slack', supported: true },
]

export const channelFormFields = {
  dingtalk: [
    { key: 'endpoint', label: 'Webhook URL', type: 'text', required: true, placeholder: 'https://oapi.dingtalk.com/robot/send?access_token=...' },
    { key: 'secret', label: '签名密钥', type: 'password', placeholder: 'SEC...' },
    { key: 'receiver', label: '接收方', type: 'text', placeholder: '手机号，多个用逗号分隔' },
  ],
  feishu: [
    { key: 'endpoint', label: 'Webhook URL', type: 'text', required: true, placeholder: 'https://open.feishu.cn/open-apis/bot/v2/hook/...' },
    { key: 'secret', label: '签名密钥', type: 'password', placeholder: '签名校验密钥' },
    { key: 'receiver', label: '接收方', type: 'text', placeholder: '用户 ID，多个用逗号分隔' },
  ],
  wecom: [
    { key: 'endpoint', label: 'Webhook URL', type: 'text', required: true, placeholder: 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=...' },
    { key: 'secret', label: 'Corp Secret', type: 'password' },
    { key: 'receiver', label: '接收方', type: 'text', placeholder: '@all 或 userid' },
  ],
  email: [
    { key: 'smtp_host', label: 'SMTP Host', type: 'text', required: true, placeholder: 'smtp.example.com' },
    { key: 'smtp_port', label: 'SMTP Port', type: 'number', required: true, placeholder: '465' },
    { key: 'smtp_from', label: '发件人地址', type: 'text', required: true, placeholder: 'alert@example.com' },
    { key: 'smtp_username', label: '用户名', type: 'text', required: true, placeholder: 'alert@example.com' },
    { key: 'smtp_password', label: '密码', type: 'password', required: true },
    { key: 'smtp_tls', label: '启用 TLS', type: 'checkbox' },
    { key: 'receiver', label: '收件人', type: 'text', required: true, placeholder: '多个邮箱用逗号分隔' },
  ],
  webhook: [
    { key: 'endpoint', label: 'Webhook URL', type: 'text', required: true, placeholder: 'https://your-service.com/webhook' },
    { key: 'http_method', label: 'HTTP Method', type: 'select', options: ['POST', 'PUT', 'GET'], placeholder: 'POST' },
    { key: 'headers', label: '自定义 Headers (JSON)', type: 'textarea', placeholder: '{"Authorization": "Bearer ..."}' },
    { key: 'secret', label: 'Secret (HMAC 签名)', type: 'password' },
    { key: 'timeout', label: '超时 (秒)', type: 'number', placeholder: '10' },
  ],
  telegram: [
    { key: 'bot_token', label: 'Bot Token', type: 'password', required: true, placeholder: '123456:ABC-DEF...' },
    { key: 'chat_id', label: 'Chat ID', type: 'text', required: true, placeholder: '-1001234567890' },
    { key: 'parse_mode', label: '解析模式', type: 'select', options: ['HTML', 'Markdown', 'MarkdownV2'] },
    { key: 'disable_notification', label: '静默发送', type: 'checkbox' },
  ],
  slack: [
    { key: 'endpoint', label: 'Webhook URL', type: 'text', required: true, placeholder: 'https://hooks.slack.com/services/...' },
    { key: 'channel', label: 'Channel', type: 'text', placeholder: '#alerts' },
    { key: 'username', label: 'Bot 名称', type: 'text', placeholder: 'FindX Alert' },
    { key: 'icon_emoji', label: 'Icon Emoji', type: 'text', placeholder: ':warning:' },
  ],
}

export const blockedContracts = {
  channelExport: '通知媒介导出只能输出脱敏视图，原始密钥导出被禁止。',
}

export const makeError = (error, fallback = '请求失败') => redactText(error?.message || fallback)
export const displayJson = (value) => safeJson(value, 9000)
export const filterText = (parts, keyword) => {
  const needle = String(keyword || '').trim().toLowerCase()
  if (!needle) return true
  return parts.filter(Boolean).join(' ').toLowerCase().includes(needle)
}

export const displayDate = (value) => {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? redactText(value) : date.toLocaleString('zh-CN', { hour12: false })
}

export const parseJson = (text, fallback) => {
  try {
    return text ? JSON.parse(text) : fallback
  } catch {
    throw new Error('JSON 格式不合法')
  }
}

export const parseLines = (text) => String(text || '').split(/\r?\n|,/).map((item) => item.trim()).filter(Boolean)
export const channelTypeLabel = (type) => channelTypes.find((item) => item.ident === type)?.label || redactText(type || '-')
export const secretView = (value) => (value ? '<SECRET>' : '')

export const normalizeChannel = (raw = {}) => ({
  raw,
  id: String(raw.id || ''),
  type: raw.type || raw.ident || 'dingtalk',
  name: redactText(raw.name || '-'),
  endpoint: secretView(raw.endpoint || raw.webhook),
  receiver: redactText(raw.receiver || ''),
  enabled: raw.enabled !== false && raw.enable !== false,
  updatedBy: redactText(raw.updated_by || raw.updatedBy || raw.created_by || '-'),
  updatedAt: raw.updated_at || raw.updatedAt || raw.created_at || raw.createdAt || '',
})

export const channelDraftFromRaw = (channel, selectedType) => {
  if (!channel) {
    const type = supportedChannelTypes.has(selectedType) ? selectedType : 'dingtalk'
    const fields = channelFormFields[type] || []
    const draft = { id: '', type, name: '', enabled: true }
    fields.forEach((f) => { draft[f.key] = f.type === 'checkbox' ? false : '' })
    return draft
  }
  const type = channel.type || channel.raw?.type || 'dingtalk'
  const fields = channelFormFields[type] || []
  const draft = { id: channel.id, type, name: channel.raw?.name || channel.name, enabled: channel.enabled }
  const raw = channel.raw || {}
  fields.forEach((f) => {
    if (f.type === 'checkbox') draft[f.key] = Boolean(raw[f.key])
    else if (f.type === 'password') draft[f.key] = raw[f.key] ? '<SECRET>' : ''
    else draft[f.key] = raw[f.key] != null ? String(raw[f.key]) : ''
  })
  return draft
}

export const channelPayloadFromDraft = (draft) => {
  if (!supportedChannelTypes.has(draft.type)) throw new Error('不支持的媒介类型')
  if (!draft.name.trim()) throw new Error('名称不能为空')
  const payload = { id: draft.id || undefined, type: draft.type, name: draft.name.trim(), enabled: draft.enabled }
  const fields = channelFormFields[draft.type] || []
  fields.forEach((f) => {
    const val = draft[f.key]
    if (f.type === 'checkbox') payload[f.key] = Boolean(val)
    else if (f.type === 'password' && val === '<SECRET>') { /* 不覆盖 */ }
    else if (f.type === 'number') payload[f.key] = val ? Number(val) : undefined
    else payload[f.key] = val || undefined
  })
  return payload
}

export const normalizeRule = (raw = {}) => ({
  raw,
  id: String(raw.id || ''),
  name: redactText(raw.name || '-'),
  description: redactText(raw.description || ''),
  enabled: raw.enabled !== false && raw.enable !== false,
  notifyConfigs: Array.isArray(raw.notify_configs) ? raw.notify_configs : [],
  alertRuleIds: Array.isArray(raw.alert_rule_ids) ? raw.alert_rule_ids : [],
  conditions: raw.conditions || {},
  timeWindow: raw.time_window || {},
  updatedBy: redactText(raw.updated_by || raw.updatedBy || raw.created_by || '-'),
  updatedAt: raw.updated_at || raw.updatedAt || raw.created_at || raw.createdAt || '',
})

export const normalizeTemplate = (raw = {}) => ({
  raw,
  id: String(raw.id || ''),
  name: redactText(raw.name || '-'),
  ident: redactText(raw.ident || ''),
  notifyChannelIdent: raw.notify_channel_ident || raw.notifyChannelIdent || 'dingtalk',
  private: Number(raw.private || 0),
  content: raw.content || { title: '', content: '' },
  updatedBy: redactText(raw.updated_by || raw.updatedBy || raw.created_by || '-'),
  updatedAt: raw.updated_at || raw.updatedAt || raw.created_at || raw.createdAt || '',
})
