import React, { useEffect, useRef } from 'react'
import { marked } from 'marked'
import sanitizeHtml from 'sanitize-html'
import { calcValue, formatMetricName } from './chartRenderers.jsx'
import { formatValue } from './unitFormat.js'

/**
 * DEGRADE-013: Pie 饼图（SVG d3.pie + d3.arc）
 */
export function PiePanel({ series, options = {} }) {
  if (series.length === 0) return <div className="fx-chart-empty">无数据</div>
  const { calc = 'last', unit } = options
  const items = series.map((s) => ({ name: formatMetricName(s), value: Math.abs(calcValue(s.values, calc)) }))
  const total = items.reduce((a, b) => a + b.value, 0) || 1
  const colors = ['#1769ff', '#e6550d', '#31a354', '#756bb1', '#636363', '#6baed6', '#fd8d3c', '#74c476']
  let startAngle = 0

  return (
    <div className="fx-pie-panel">
      <svg viewBox="0 0 200 200" className="fx-pie-panel__svg">
        {items.map((item, i) => {
          const angle = (item.value / total) * Math.PI * 2
          const endAngle = startAngle + angle
          const x1 = 100 + 80 * Math.cos(startAngle - Math.PI / 2)
          const y1 = 100 + 80 * Math.sin(startAngle - Math.PI / 2)
          const x2 = 100 + 80 * Math.cos(endAngle - Math.PI / 2)
          const y2 = 100 + 80 * Math.sin(endAngle - Math.PI / 2)
          const largeArc = angle > Math.PI ? 1 : 0
          const path = `M 100 100 L ${x1} ${y1} A 80 80 0 ${largeArc} 1 ${x2} ${y2} Z`
          startAngle = endAngle
          return <path key={i} d={path} fill={colors[i % colors.length]} />
        })}
      </svg>
      <div className="fx-pie-panel__legend">
        {items.map((item, i) => (
          <span key={i} className="fx-pie-panel__legend-item">
            <span style={{ width: 10, height: 10, background: colors[i % colors.length], borderRadius: 2, display: 'inline-block' }} />
            <span>{item.name}: {formatValue(item.value, unit)}</span>
          </span>
        ))}
      </div>
    </div>
  )
}

/**
 * DEGRADE-013: Hexbin 蜂窝图
 */
export function HexbinPanel({ series, options = {} }) {
  if (series.length === 0) return <div className="fx-chart-empty">无数据</div>
  const { calc = 'last', colorRange = ['#eef5ff', '#1769ff'] } = options
  const items = series.map((s) => ({ name: formatMetricName(s), value: calcValue(s.values, calc) }))
  const max = Math.max(...items.map((i) => Math.abs(i.value)), 1)
  const hexSize = 30
  const cols = Math.ceil(Math.sqrt(items.length))

  return (
    <div className="fx-hexbin-panel">
      <svg viewBox={`0 0 ${cols * hexSize * 2} ${Math.ceil(items.length / cols) * hexSize * 2}`} className="fx-hexbin-panel__svg">
        {items.map((item, i) => {
          const row = Math.floor(i / cols)
          const col = i % cols
          const cx = col * hexSize * 1.8 + (row % 2 ? hexSize * 0.9 : 0) + hexSize
          const cy = row * hexSize * 1.6 + hexSize
          const pct = Math.abs(item.value) / max
          const color = interpolateColor(colorRange[0], colorRange[1], pct)
          return (
            <g key={i}>
              <polygon points={hexPoints(cx, cy, hexSize * 0.8)} fill={color} stroke="#fff" strokeWidth="1" />
              <text x={cx} y={cy + 4} textAnchor="middle" fontSize="9" fill="#17233c">{item.value.toFixed(1)}</text>
            </g>
          )
        })}
      </svg>
    </div>
  )
}

/**
 * DEGRADE-013: Text 文本卡片（Markdown 渲染）
 */
export function TextPanel({ panel, options = {} }) {
  const { fontSize = 14, textColor = '#17233c', backgroundColor, textAlign = 'left' } = options
  const content = panel?.raw?.content || panel?.content || options.content || ''
  const html = sanitizeHtml(marked.parse(content || ''), {
    allowedTags: sanitizeHtml.defaults.allowedTags.concat(['img', 'h1', 'h2', 'h3']),
  })

  return (
    <div
      className="fx-text-panel"
      style={{ fontSize, color: textColor, background: backgroundColor, textAlign, padding: 12 }}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}

/**
 * DEGRADE-013: Heatmap 热力图
 */
export function HeatmapPanel({ series, options = {} }) {
  if (series.length === 0) return <div className="fx-chart-empty">无数据</div>
  const { colorScheme = ['#eef5ff', '#1769ff', '#e6550d'] } = options
  const cellSize = 16
  const maxPoints = 50
  const displaySeries = series.slice(0, 10)

  return (
    <div className="fx-heatmap-panel" style={{ overflow: 'auto' }}>
      <svg width={maxPoints * cellSize + 60} height={displaySeries.length * cellSize + 30}>
        {displaySeries.map((s, row) => {
          const vals = s.values.slice(-maxPoints)
          const max = Math.max(...vals.map(([, v]) => Math.abs(Number(v))), 1)
          return vals.map(([, v], col) => {
            const pct = Math.abs(Number(v)) / max
            const color = pct > 0.66 ? colorScheme[2] : pct > 0.33 ? colorScheme[1] : colorScheme[0]
            return (
              <rect key={`${row}-${col}`} x={col * cellSize + 60} y={row * cellSize} width={cellSize - 1} height={cellSize - 1} fill={color} rx="2" />
            )
          })
        })}
        {displaySeries.map((s, row) => (
          <text key={row} x="0" y={row * cellSize + cellSize - 3} fontSize="9" fill="#65758d">
            {formatMetricName(s).slice(0, 8)}
          </text>
        ))}
      </svg>
    </div>
  )
}

/**
 * DEGRADE-013: Iframe 内嵌文档
 */
export function IframePanel({ panel, options = {} }) {
  const url = options.url || panel?.raw?.url || panel?.url || ''
  if (!url) return <div className="fx-chart-empty">请配置 URL</div>
  return (
    <iframe src={url} className="fx-iframe-panel" title="嵌入内容" sandbox="allow-scripts allow-same-origin" />
  )
}

/**
 * DEGRADE-013: BarChart 柱状图（uPlot bars 模式）
 */
export function BarChartPanel({ series, options = {} }) {
  const containerRef = useRef(null)
  if (series.length === 0) return <div className="fx-chart-empty">无数据</div>
  const { calc = 'last', unit } = options
  const items = series.map((s) => ({ name: formatMetricName(s), value: calcValue(s.values, calc) }))
  const max = Math.max(...items.map((i) => Math.abs(i.value)), 1)
  const colors = ['#1769ff', '#e6550d', '#31a354', '#756bb1', '#636363', '#6baed6']

  return (
    <div className="fx-barchart-panel" ref={containerRef}>
      <div className="fx-barchart-panel__bars">
        {items.map((item, i) => {
          const pct = (Math.abs(item.value) / max) * 100
          return (
            <div key={i} className="fx-barchart-panel__col">
              <div className="fx-barchart-panel__bar" style={{ height: `${pct}%`, background: colors[i % colors.length] }} />
              <span className="fx-barchart-panel__label">{item.name.slice(0, 10)}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

/**
 * DEGRADE-013: TableNG 增强表格
 */
export function TableNGPanel({ series, options = {} }) {
  if (series.length === 0) return <div className="fx-chart-empty">无数据</div>
  const { showHeader = true, unit } = options

  return (
    <div className="fx-tableng-panel">
      <table>
        {showHeader && (
          <thead><tr><th>指标</th><th>最新值</th><th>最小值</th><th>最大值</th><th>平均值</th></tr></thead>
        )}
        <tbody>
          {series.map((s, i) => {
            const nums = s.values.map(([, v]) => Number(v)).filter(Number.isFinite)
            const last = nums.length > 0 ? nums[nums.length - 1] : 0
            const min = nums.length > 0 ? Math.min(...nums) : 0
            const max = nums.length > 0 ? Math.max(...nums) : 0
            const avg = nums.length > 0 ? nums.reduce((a, b) => a + b, 0) / nums.length : 0
            return (
              <tr key={i}>
                <td>{formatMetricName(s)}</td>
                <td>{formatValue(last, unit)}</td>
                <td>{formatValue(min, unit)}</td>
                <td>{formatValue(max, unit)}</td>
                <td>{formatValue(avg, unit)}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

function hexPoints(cx, cy, size) {
  const points = []
  for (let i = 0; i < 6; i++) {
    const angle = (Math.PI / 3) * i - Math.PI / 6
    points.push(`${cx + size * Math.cos(angle)},${cy + size * Math.sin(angle)}`)
  }
  return points.join(' ')
}

function interpolateColor(c1, c2, t) {
  const hex = (c) => parseInt(c.slice(1), 16)
  const r1 = (hex(c1) >> 16) & 255, g1 = (hex(c1) >> 8) & 255, b1 = hex(c1) & 255
  const r2 = (hex(c2) >> 16) & 255, g2 = (hex(c2) >> 8) & 255, b2 = hex(c2) & 255
  const r = Math.round(r1 + (r2 - r1) * t)
  const g = Math.round(g1 + (g2 - g1) * t)
  const b = Math.round(b1 + (b2 - b1) * t)
  return `rgb(${r},${g},${b})`
}
