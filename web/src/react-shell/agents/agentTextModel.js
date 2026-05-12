const mojibakePattern = /[\uFFFD]|\?\?\?/

export const displayText = (value, fallback = '-') => {
  if (value === null || value === undefined || value === '') return fallback
  const text = String(value)
  return mojibakePattern.test(text) ? '内容不可读' : text
}

export const fmtTime = value => {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString('zh-CN', { hour12: false })
}

export const agentKey = row => row?.id || row?.ident || row?.agent_id || `${row?.ip || ''}-${row?.hostname || ''}`
export const agentOnline = row => row?.online === true || row?.status === 'online'
export const sourceStateLabel = state => ({
  LOCAL_SOURCE_MISSING: '源码未接入',
  LOCAL_SOURCE_PRESENT: '源码待打包',
  READY: '能力包可用',
  BLOCKED: '接入阻断',
})[state] || state || '未知'

export const rowText = row => Object.values(row || {}).map(value => displayText(value, '')).join(' ').toLowerCase()
