import React, { useEffect, useRef, useState } from 'react'
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
 * Heatmap 热力图（时间桶 x 值范围桶）
 * - X 轴：时间桶
 * - Y 轴：值范围（自动分桶）
 * - 颜色强度基于计数
 * - 悬停显示桶详情
 */
export function HeatmapPanel({ series, options = {} }) {
  const [tooltip, setTooltip] = useState(null)
  if (series.length === 0) return <div className="fx-chart-empty">无数据</div>
  const { colorScheme = ['#eef5ff', '#1769ff', '#e6550d'], bucketCount = 8, timeBuckets = 30 } = options

  // 收集所有数据点
  const allPoints = []
  series.forEach(s => {
    (s.values || []).forEach(([ts, v]) => {
      const num = Number(v)
      if (Number.isFinite(num)) allPoints.push({ ts: Number(ts), value: num })
    })
  })
  if (allPoints.length === 0) return <div className="fx-chart-empty">无数据</div>

  // 计算时间范围和值范围
  const tsMin = Math.min(...allPoints.map(p => p.ts))
  const tsMax = Math.max(...allPoints.map(p => p.ts))
  const valMin = Math.min(...allPoints.map(p => p.value))
  const valMax = Math.max(...allPoints.map(p => p.value))
  const valRange = valMax - valMin || 1
  const tsRange = tsMax - tsMin || 1

  // 构建桶矩阵
  const matrix = Array.from({ length: bucketCount }, () => Array(timeBuckets).fill(0))
  allPoints.forEach(p => {
    const col = Math.min(Math.floor(((p.ts - tsMin) / tsRange) * timeBuckets), timeBuckets - 1)
    const row = Math.min(Math.floor(((p.value - valMin) / valRange) * bucketCount), bucketCount - 1)
    matrix[row][col]++
  })

  const maxCount = Math.max(...matrix.flat(), 1)
  const cellW = 12
  const cellH = 14
  const svgWidth = timeBuckets * cellW + 60
  const svgHeight = bucketCount * cellH + 30

  function getColor(count) {
    if (count === 0) return colorScheme[0]
    const pct = count / maxCount
    if (pct > 0.66) return colorScheme[2] || '#e6550d'
    if (pct > 0.33) return colorScheme[1] || '#1769ff'
    return interpolateColor(colorScheme[0], colorScheme[1], pct * 3)
  }

  function formatTs(idx) {
    const ts = tsMin + (idx / timeBuckets) * tsRange
    const d = new Date(ts * 1000)
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
  }

  return (
    <div className="fx-heatmap-panel" style={{ overflow: 'auto', position: 'relative' }}>
      <svg width={svgWidth} height={svgHeight}>
        {matrix.map((row, ri) =>
          row.map((count, ci) => (
            <rect
              key={`${ri}-${ci}`}
              x={ci * cellW + 50}
              y={(bucketCount - 1 - ri) * cellH}
              width={cellW - 1}
              height={cellH - 1}
              fill={getColor(count)}
              rx="1"
              onMouseEnter={(e) => {
                const bucketLow = valMin + (ri / bucketCount) * valRange
                const bucketHigh = valMin + ((ri + 1) / bucketCount) * valRange
                setTooltip({
                  x: e.clientX, y: e.clientY,
                  time: formatTs(ci),
                  range: `${bucketLow.toFixed(1)} - ${bucketHigh.toFixed(1)}`,
                  count,
                })
              }}
              onMouseLeave={() => setTooltip(null)}
            />
          ))
        )}
        {/* Y 轴标签 */}
        {[0, Math.floor(bucketCount / 2), bucketCount - 1].map(ri => (
          <text key={ri} x="0" y={(bucketCount - 1 - ri) * cellH + cellH - 3} fontSize="8" fill="#9ca3af">
            {(valMin + (ri / bucketCount) * valRange).toFixed(0)}
          </text>
        ))}
        {/* X 轴标签 */}
        {[0, Math.floor(timeBuckets / 2), timeBuckets - 1].map(ci => (
          <text key={ci} x={ci * cellW + 50} y={bucketCount * cellH + 12} fontSize="8" fill="#9ca3af">
            {formatTs(ci)}
          </text>
        ))}
      </svg>
      {tooltip && (
        <div style={{ position: 'fixed', left: tooltip.x + 10, top: tooltip.y - 30, background: 'rgba(0,0,0,0.85)', color: '#fff', padding: '4px 8px', borderRadius: 4, fontSize: 11, pointerEvents: 'none', zIndex: 999 }}>
          <div>时间: {tooltip.time}</div>
          <div>范围: {tooltip.range}</div>
          <div>计数: {tooltip.count}</div>
        </div>
      )}
    </div>
  )
}

export function IframePanel({ panel, options = {} }) {
	const url = options.url || panel?.raw?.url || panel?.url || ''
	if (!url) return <div className="fx-chart-empty">请配置 URL</div>
	const sanitizedUrl = (url.startsWith('http://') || url.startsWith('https://')) ? url : ''
	if (!sanitizedUrl) return <div className="fx-chart-empty">仅支持 http/https 协议</div>
	return (
		<div className="fx-iframe-panel" style={{ width: '100%', height: '100%' }}>
			<iframe
				src={sanitizedUrl}
				style={{ width: '100%', height: '100%', border: 'none' }}
				sandbox="allow-scripts allow-same-origin"
				title="嵌入面板"
			/>
		</div>
	)
}

/**
 * BarChart 柱状图（SVG 渲染，支持堆叠模式）
 * - 接受时间序列数据，渲染垂直柱状图
 * - 支持 stacked 模式
 * - 颜色使用 CSS 变量
 */
export function BarChartPanel({ series, options = {} }) {
  const [tooltip, setTooltip] = useState(null)
  if (series.length === 0) return <div className="fx-chart-empty">无数据</div>
  const { calc = 'last', unit, stacked = false } = options
  const colors = ['var(--fx-chart-1, #1769ff)', 'var(--fx-chart-2, #e6550d)', 'var(--fx-chart-3, #31a354)', 'var(--fx-chart-4, #756bb1)', 'var(--fx-chart-5, #636363)', 'var(--fx-chart-6, #6baed6)']

  const items = series.map((s) => ({ name: formatMetricName(s), value: calcValue(s.values, calc) }))

  const svgWidth = 400
  const svgHeight = 200
  const padding = { top: 20, right: 20, bottom: 40, left: 50 }
  const chartWidth = svgWidth - padding.left - padding.right
  const chartHeight = svgHeight - padding.top - padding.bottom

  let maxVal
  if (stacked) {
    maxVal = items.reduce((sum, item) => sum + Math.abs(item.value), 0) || 1
  } else {
    maxVal = Math.max(...items.map((i) => Math.abs(i.value)), 1)
  }

  const barWidth = Math.min(40, chartWidth / items.length - 4)
  const barGap = (chartWidth - barWidth * items.length) / (items.length + 1)

  let stackY = chartHeight

  return (
    <div className="fx-barchart-panel" style={{ position: 'relative' }}>
      <svg viewBox={`0 0 ${svgWidth} ${svgHeight}`} style={{ width: '100%', height: '100%' }}>
        {/* Y 轴刻度 */}
        {[0, 0.25, 0.5, 0.75, 1].map((pct) => {
          const y = padding.top + chartHeight * (1 - pct)
          return (
            <g key={pct}>
              <line x1={padding.left} y1={y} x2={svgWidth - padding.right} y2={y} stroke="#e5e7eb" strokeWidth="0.5" />
              <text x={padding.left - 6} y={y + 3} textAnchor="end" fontSize="9" fill="#9ca3af">
                {formatValue(maxVal * pct, unit)}
              </text>
            </g>
          )
        })}
        {/* 柱状图 */}
        {stacked ? (
          (() => {
            let accY = 0
            return items.map((item, i) => {
              const barH = (Math.abs(item.value) / maxVal) * chartHeight
              const y = padding.top + chartHeight - accY - barH
              accY += barH
              return (
                <g key={i}
                  onMouseEnter={() => setTooltip({ name: item.name, value: item.value })}
                  onMouseLeave={() => setTooltip(null)}
                >
                  <rect
                    x={padding.left + chartWidth / 2 - barWidth / 2}
                    y={y}
                    width={barWidth}
                    height={barH}
                    fill={colors[i % colors.length]}
                    rx="2"
                  />
                </g>
              )
            })
          })()
        ) : (
          items.map((item, i) => {
            const barH = (Math.abs(item.value) / maxVal) * chartHeight
            const x = padding.left + barGap + i * (barWidth + barGap)
            const y = padding.top + chartHeight - barH
            return (
              <g key={i}
                onMouseEnter={() => setTooltip({ name: item.name, value: item.value })}
                onMouseLeave={() => setTooltip(null)}
              >
                <rect x={x} y={y} width={barWidth} height={barH} fill={colors[i % colors.length]} rx="2" />
                <text x={x + barWidth / 2} y={svgHeight - padding.bottom + 14} textAnchor="middle" fontSize="8" fill="#6b7280">
                  {item.name.slice(0, 8)}
                </text>
              </g>
            )
          })
        )}
      </svg>
      {tooltip && (
        <div style={{ position: 'absolute', top: 4, right: 4, background: 'rgba(0,0,0,0.8)', color: '#fff', padding: '4px 8px', borderRadius: 4, fontSize: 11 }}>
          {tooltip.name}: {formatValue(tooltip.value, unit)}
        </div>
      )}
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
