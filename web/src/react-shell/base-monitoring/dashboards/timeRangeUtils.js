/**
 * timeRangeUtils.js — 时间范围解析工具
 * 对齐夜莺 TimeRangePicker/utils.ts 的核心逻辑
 */

/** 快捷时间范围选项（对齐夜莺 rangeOptions） */
export const QUICK_RANGES = [
  { start: 'now-1m', end: 'now', label: '最近 1 分钟' },
  { start: 'now-2m', end: 'now', label: '最近 2 分钟' },
  { start: 'now-3m', end: 'now', label: '最近 3 分钟' },
  { start: 'now-5m', end: 'now', label: '最近 5 分钟' },
  { start: 'now-15m', end: 'now', label: '最近 15 分钟' },
  { start: 'now-30m', end: 'now', label: '最近 30 分钟' },
  { start: 'now-1h', end: 'now', label: '最近 1 小时' },
  { start: 'now-3h', end: 'now', label: '最近 3 小时' },
  { start: 'now-6h', end: 'now', label: '最近 6 小时' },
  { start: 'now-12h', end: 'now', label: '最近 12 小时' },
  { start: 'now-24h', end: 'now', label: '最近 24 小时' },
  { start: 'now-2d', end: 'now', label: '最近 2 天' },
  { start: 'now-7d', end: 'now', label: '最近 7 天' },
  { start: 'now-30d', end: 'now', label: '最近 30 天' },
]

/** 自动刷新选项 */
export const REFRESH_OPTIONS = [
  { key: 'off', label: 'Off', seconds: 0 },
  { key: '5s', label: '5s', seconds: 5 },
  { key: '10s', label: '10s', seconds: 10 },
  { key: '20s', label: '20s', seconds: 20 },
  { key: '30s', label: '30s', seconds: 30 },
  { key: '1m', label: '1m', seconds: 60 },
  { key: '2m', label: '2m', seconds: 120 },
  { key: '3m', label: '3m', seconds: 180 },
  { key: '5m', label: '5m', seconds: 300 },
  { key: '10m', label: '10m', seconds: 600 },
]

const DURATION_RE = /^now(?:([+-])(\d+)([smhdwMy]))?(?:\/([smhdwMy]))?$/

function unitToSeconds(amount, unit) {
  switch (unit) {
    case 's': return amount
    case 'm': return amount * 60
    case 'h': return amount * 3600
    case 'd': return amount * 86400
    case 'w': return amount * 604800
    case 'M': return amount * 2592000
    case 'y': return amount * 31536000
    default: return 0
  }
}

/** 解析 now-Xm 语法为 unix 秒 */
export function parseMathExpr(expr) {
  if (!expr || typeof expr !== 'string') return null
  const m = DURATION_RE.exec(expr.trim())
  if (!m) return null
  let ts = Math.floor(Date.now() / 1000)
  if (m[1] && m[2] && m[3]) {
    const amount = parseInt(m[2], 10)
    const unit = m[3]
    const secs = unitToSeconds(amount, unit)
    ts = m[1] === '-' ? ts - secs : ts + secs
  }
  return ts
}

/** 判断是否为 now-Xm 数学表达式 */
export function isMathString(text) {
  return typeof text === 'string' && text.startsWith('now')
}

/** 判断时间字符串是否有效（数学表达式或 ISO 日期） */
export function isValidTimeStr(text) {
  if (!text) return false
  if (isMathString(text)) return parseMathExpr(text) !== null
  return !isNaN(new Date(text).getTime())
}

/** 将时间范围描述为可读文本 */
export function describeRange(range) {
  if (!range || !range.start || !range.end) return ''
  const match = QUICK_RANGES.find((r) => r.start === range.start && r.end === range.end)
  if (match) return match.label
  return `${range.start} ~ ${range.end}`
}

/** 格式化日期为 YYYY-MM-DD HH:mm */
export function formatDate(date) {
  const d = new Date(date)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/* ===== localStorage 历史缓存 ===== */
const HISTORY_KEY = 'fx-timeRangePicker-history'

export function getHistoryCache() {
  try {
    const raw = localStorage.getItem(HISTORY_KEY)
    return raw ? JSON.parse(raw).slice(0, 4) : []
  } catch { return [] }
}

export function setHistoryCache(range) {
  try {
    const list = getHistoryCache()
    const entry = { start: range.start, end: range.end }
    const filtered = list.filter((r) => r.start !== entry.start || r.end !== entry.end)
    const updated = [entry, ...filtered].slice(0, 4)
    localStorage.setItem(HISTORY_KEY, JSON.stringify(updated))
  } catch { /* ignore */ }
}
