import React, { useMemo } from 'react'
import { displayJson, displayText } from './metricQueryModel.js'

const colors = ['#1f78ff', '#14a06f', '#f59f00', '#d6336c', '#7048e8', '#0b7285', '#e67700', '#2b8a3e']

const readSeries = (result) => {
  const rows = result?.data?.result ?? result?.result ?? []
  return Array.isArray(rows) ? rows : []
}

const seriesName = (row, index) => (
  row?.metric && Object.keys(row.metric).length ? displayJson(row.metric, 400) : `series_${index + 1}`
)

export function MetricGraph({ result, mode = 'line', showLegend = true }) {
  const chart = useMemo(() => {
    const rows = readSeries(result).slice(0, 8)
    const series = rows.map((row, index) => ({
      name: seriesName(row, index),
      values: Array.isArray(row?.values)
        ? row.values
          .map((point) => [Number(point?.[0]), Number(point?.[1])])
          .filter(([time, value]) => Number.isFinite(time) && Number.isFinite(value))
        : [],
    })).filter((item) => item.values.length > 0)

    const points = series.flatMap((item) => item.values)
    if (!points.length) return { series, points: [], pathSets: [] }

    const minX = Math.min(...points.map(([time]) => time))
    const maxX = Math.max(...points.map(([time]) => time))
    const minY = Math.min(...points.map(([, value]) => value))
    const maxY = Math.max(...points.map(([, value]) => value))
    const width = 760
    const height = 240
    const pad = 26
    const xRange = Math.max(maxX - minX, 1)
    const yRange = Math.max(maxY - minY, 1)
    const x = (time) => pad + ((time - minX) / xRange) * (width - pad * 2)
    const y = (value) => height - pad - ((value - minY) / yRange) * (height - pad * 2)
    const pathSets = series.map((item, index) => {
      const line = item.values.map(([time, value]) => `${x(time)},${y(value)}`).join(' ')
      const area = `${pad},${height - pad} ${line} ${width - pad},${height - pad}`
      return { ...item, line, area, color: colors[index % colors.length] }
    })
    return { series, points, pathSets, width, height, minY, maxY }
  }, [result])

  if (!result) return <div className='fx-query-graph-empty'>执行区间查询后渲染图表。</div>
  if (!chart.points.length) return <div className='fx-query-graph-empty'>查询契约未返回可绘制的图表点。</div>

  return (
    <div className='fx-query-graph'>
      <svg viewBox={`0 0 ${chart.width} ${chart.height}`} role='img' aria-label='指标查询图表'>
        {[0, 1, 2, 3].map((line) => {
          const y = 26 + line * ((chart.height - 52) / 3)
          return <line key={line} x1='26' x2={chart.width - 26} y1={y} y2={y} className='fx-query-graph__grid' />
        })}
        {chart.pathSets.map((item) => (
          <g key={item.name}>
            {mode === 'area' && <polygon points={item.area} fill={item.color} opacity='0.12' />}
            <polyline points={item.line} fill='none' stroke={item.color} strokeWidth='2' strokeLinejoin='round' strokeLinecap='round' />
          </g>
        ))}
      </svg>
      <div className='fx-query-graph__axis'>
        <span>最小 {displayText(chart.minY)}</span>
        <span>最大 {displayText(chart.maxY)}</span>
      </div>
      {showLegend && (
        <div className='fx-query-legend'>
          {chart.pathSets.map((item) => (
            <span key={item.name}>
              <i style={{ background: item.color }} />
              {displayText(item.name)}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
