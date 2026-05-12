import React from 'react'
import { formatValue } from './unitFormat.js'

/**
 * DEGRADE-013: Stat 面板（增强版）
 * 支持 textMode(valueAndName/value/name)、colorMode、calcMode、direction
 */
export function StatPanel({ series, options = {} }) {
  if (series.length === 0) {
    return <div className="fx-chart-empty">无数据</div>
  }
  const { textMode = 'valueAndName', colorMode = 'value', calc = 'last', direction = 'vertical', unit } = options
  const items = series.map((s) => {
    const val = calcValue(s.values, calc)
    return { name: formatMetricName(s), value: val, formatted: formatValue(val, unit) }
  })

  return (
    <div className={`fx-stat-panel fx-stat-panel--${direction}`}>
      {items.map((item, i) => (
        <div key={i} className="fx-stat-panel__item" style={colorMode === 'background' ? { background: 'var(--fx-blue-light, #eef5ff)' } : undefined}>
          {(textMode === 'valueAndName' || textMode === 'value') && (
            <span className="fx-stat-panel__value" style={colorMode === 'value' ? { color: 'var(--fx-blue)' } : undefined}>
              {item.formatted}
            </span>
          )}
          {(textMode === 'valueAndName' || textMode === 'name') && (
            <span className="fx-stat-panel__name">{item.name}</span>
          )}
        </div>
      ))}
    </div>
  )
}

/**
 * DEGRADE-013: BarGauge 面板（水平条形排行榜）
 */
export function BarGaugePanel({ series, options = {} }) {
  if (series.length === 0) return <div className="fx-chart-empty">无数据</div>
  const { calc = 'last', baseColor = '#1769ff', displayMode = 'gradient', sort = 'desc', unit } = options
  const items = series.map((s) => ({
    name: formatMetricName(s),
    value: calcValue(s.values, calc),
  }))
  if (sort === 'desc') items.sort((a, b) => b.value - a.value)
  else if (sort === 'asc') items.sort((a, b) => a.value - b.value)
  const max = Math.max(...items.map((i) => Math.abs(i.value)), 1)

  return (
    <div className="fx-bargauge-panel">
      {items.map((item, i) => {
        const pct = Math.min(100, (Math.abs(item.value) / max) * 100)
        const bg = displayMode === 'gradient'
          ? `linear-gradient(90deg, ${baseColor}22, ${baseColor})`
          : baseColor
        return (
          <div key={i} className="fx-bargauge-panel__row">
            <span className="fx-bargauge-panel__label">{item.name}</span>
            <div className="fx-bargauge-panel__bar-wrap">
              <div className="fx-bargauge-panel__bar" style={{ width: `${pct}%`, background: bg }} />
            </div>
            <span className="fx-bargauge-panel__value">{formatValue(item.value, unit)}</span>
          </div>
        )
      })}
    </div>
  )
}

/**
 * DEGRADE-013: Gauge 面板（仪表盘圆弧图）
 */
export function GaugePanel({ series, options = {} }) {
  if (series.length === 0) return <div className="fx-chart-empty">无数据</div>
  const { calc = 'last', min = 0, max = 100, unit } = options
  const value = calcValue(series[0].values, calc)
  const pct = Math.min(1, Math.max(0, (value - min) / (max - min || 1)))
  const angle = pct * 180
  const color = pct > 0.8 ? '#e6550d' : pct > 0.5 ? '#f5a623' : '#31a354'

  return (
    <div className="fx-gauge-panel">
      <svg viewBox="0 0 200 120" className="fx-gauge-panel__svg">
        <path d="M 20 100 A 80 80 0 0 1 180 100" fill="none" stroke="#e8ecf0" strokeWidth="12" strokeLinecap="round" />
        <path d="M 20 100 A 80 80 0 0 1 180 100" fill="none" stroke={color} strokeWidth="12" strokeLinecap="round"
          strokeDasharray={`${pct * 251.2} 251.2`} />
        <text x="100" y="90" textAnchor="middle" fontSize="24" fontWeight="700" fill="#17233c">
          {formatValue(value, unit)}
        </text>
        <text x="100" y="110" textAnchor="middle" fontSize="11" fill="#65758d">
          {formatMetricName(series[0])}
        </text>
      </svg>
    </div>
  )
}

function formatMetricName(s) {
  if (s.legendFormat) return s.legendFormat
  const m = s.metric || {}
  const name = m.__name__ || ''
  const labels = Object.entries(m).filter(([k]) => k !== '__name__').map(([k, v]) => `${k}="${v}"`).join(', ')
  return labels ? `${name}{${labels}}` : name || 'series'
}

function calcValue(values, calc) {
  if (!values || values.length === 0) return 0
  const nums = values.map(([, v]) => Number(v)).filter(Number.isFinite)
  if (nums.length === 0) return 0
  if (calc === 'last') return nums[nums.length - 1]
  if (calc === 'first') return nums[0]
  if (calc === 'max') return Math.max(...nums)
  if (calc === 'min') return Math.min(...nums)
  if (calc === 'avg') return nums.reduce((a, b) => a + b, 0) / nums.length
  if (calc === 'sum') return nums.reduce((a, b) => a + b, 0)
  if (calc === 'count') return nums.length
  return nums[nums.length - 1]
}

export { calcValue, formatMetricName }
