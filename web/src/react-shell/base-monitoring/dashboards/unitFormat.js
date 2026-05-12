/**
 * DEGRADE-014: 完整单位格式化体系
 * 对齐夜莺 UnitPicker 的单位分类和格式化逻辑
 */

export const UNIT_CATEGORIES = [
  {
    label: '无',
    units: [{ key: 'none', label: '无', format: (v) => formatNumber(v) }],
  },
  {
    label: '数据 (IEC)',
    units: [
      { key: 'bytes', label: 'bytes', format: (v) => formatBytes(v, 1024) },
      { key: 'bytesIEC', label: 'bytes (IEC)', format: (v) => formatBytes(v, 1024) },
      { key: 'bitsIEC', label: 'bits (IEC)', format: (v) => formatBits(v) },
    ],
  },
  {
    label: '数据 (SI)',
    units: [
      { key: 'bytesSI', label: 'bytes (SI)', format: (v) => formatBytes(v, 1000) },
    ],
  },
  {
    label: '百分比',
    units: [
      { key: 'percent', label: '% (0-100)', format: (v) => `${formatNumber(v)}%` },
      { key: 'percentUnit', label: '% (0.0-1.0)', format: (v) => `${formatNumber(v * 100)}%` },
    ],
  },
  {
    label: '时间',
    units: [
      { key: 'seconds', label: '秒 (s)', format: (v) => formatDuration(v, 's') },
      { key: 'ms', label: '毫秒 (ms)', format: (v) => formatDuration(v, 'ms') },
      { key: 'us', label: '微秒 (us)', format: (v) => formatDuration(v, 'us') },
      { key: 'ns', label: '纳秒 (ns)', format: (v) => formatDuration(v, 'ns') },
    ],
  },
  {
    label: '吞吐量',
    units: [
      { key: 'reqps', label: 'requests/s', format: (v) => `${formatNumber(v)} req/s` },
      { key: 'ops', label: 'ops/s', format: (v) => `${formatNumber(v)} ops/s` },
      { key: 'rps', label: 'reads/s', format: (v) => `${formatNumber(v)} rd/s` },
      { key: 'wps', label: 'writes/s', format: (v) => `${formatNumber(v)} wr/s` },
      { key: 'iops', label: 'iops', format: (v) => `${formatNumber(v)} iops` },
    ],
  },
  {
    label: '数据速率',
    units: [
      { key: 'bps', label: 'bits/s', format: (v) => `${formatSI(v)}bps` },
      { key: 'Bps', label: 'bytes/s', format: (v) => `${formatSI(v)}B/s` },
    ],
  },
  {
    label: '货币',
    units: [
      { key: 'currencyUSD', label: 'USD ($)', format: (v) => `$${formatNumber(v)}` },
      { key: 'currencyCNY', label: 'CNY (¥)', format: (v) => `¥${formatNumber(v)}` },
      { key: 'currencyEUR', label: 'EUR (€)', format: (v) => `€${formatNumber(v)}` },
    ],
  },
]

function formatNumber(v) {
  if (v === null || v === undefined || !Number.isFinite(v)) return '-'
  if (Math.abs(v) < 0.001 && v !== 0) return v.toExponential(2)
  if (Math.abs(v) >= 1e9) return (v / 1e9).toFixed(2) + 'G'
  if (Math.abs(v) >= 1e6) return (v / 1e6).toFixed(2) + 'M'
  if (Math.abs(v) >= 1e4) return (v / 1e3).toFixed(1) + 'K'
  if (v % 1 === 0) return v.toString()
  return v.toFixed(2)
}

function formatSI(v) {
  if (Math.abs(v) >= 1e12) return (v / 1e12).toFixed(2) + 'T'
  if (Math.abs(v) >= 1e9) return (v / 1e9).toFixed(2) + 'G'
  if (Math.abs(v) >= 1e6) return (v / 1e6).toFixed(2) + 'M'
  if (Math.abs(v) >= 1e3) return (v / 1e3).toFixed(1) + 'K'
  return formatNumber(v)
}

function formatBytes(v, base = 1024) {
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  if (v === 0) return '0 B'
  let i = 0
  let val = Math.abs(v)
  while (val >= base && i < units.length - 1) { val /= base; i++ }
  return `${(v < 0 ? -val : val).toFixed(i === 0 ? 0 : 2)} ${units[i]}`
}

function formatBits(v) {
  return formatBytes(v / 8, 1024).replace('B', 'b')
}

function formatDuration(v, sourceUnit) {
  let seconds = v
  if (sourceUnit === 'ms') seconds = v / 1000
  else if (sourceUnit === 'us') seconds = v / 1e6
  else if (sourceUnit === 'ns') seconds = v / 1e9

  if (Math.abs(seconds) >= 86400) return `${(seconds / 86400).toFixed(1)}d`
  if (Math.abs(seconds) >= 3600) return `${(seconds / 3600).toFixed(1)}h`
  if (Math.abs(seconds) >= 60) return `${(seconds / 60).toFixed(1)}m`
  if (Math.abs(seconds) >= 1) return `${seconds.toFixed(2)}s`
  if (Math.abs(seconds) >= 0.001) return `${(seconds * 1000).toFixed(1)}ms`
  if (Math.abs(seconds) >= 1e-6) return `${(seconds * 1e6).toFixed(1)}us`
  return `${(seconds * 1e9).toFixed(0)}ns`
}

const unitMap = new Map()
for (const cat of UNIT_CATEGORIES) {
  for (const u of cat.units) {
    unitMap.set(u.key, u)
  }
}

/**
 * 根据 unit key 格式化数值
 */
export function formatValue(value, unitKey) {
  if (value === null || value === undefined) return '-'
  const num = Number(value)
  if (!Number.isFinite(num)) return '-'
  const unit = unitMap.get(unitKey)
  if (!unit) return formatNumber(num)
  return unit.format(num)
}

/**
 * 获取所有单位选项（扁平列表）
 */
export function getAllUnits() {
  const result = []
  for (const cat of UNIT_CATEGORIES) {
    for (const u of cat.units) {
      result.push({ key: u.key, label: u.label, category: cat.label })
    }
  }
  return result
}
