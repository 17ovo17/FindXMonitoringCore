import React, { useEffect, useRef, useMemo } from 'react'
import uPlot from 'uplot'
import 'uplot/dist/uPlot.min.css'
import { displayJson, displayText } from './metricQueryModel.js'

const colors = ['#1f78ff', '#14a06f', '#f59f00', '#d6336c', '#7048e8', '#0b7285', '#e67700', '#2b8a3e']

const readSeries = (result) => {
  const rows = result?.data?.result ?? result?.result ?? []
  return Array.isArray(rows) ? rows : []
}

const seriesName = (row, index) => (
  row?.metric && Object.keys(row.metric).length ? displayJson(row.metric, 60) : `series_${index + 1}`
)

export function MetricGraph({ result, mode = 'line', showLegend = true, onBrush }) {
  const containerRef = useRef(null)
  const chartRef = useRef(null)

  const data = useMemo(() => {
    const rows = readSeries(result)
    if (!rows.length) return null

    const allTimestamps = new Set()
    rows.forEach((row) => {
      if (Array.isArray(row?.values)) {
        row.values.forEach(([t]) => allTimestamps.add(Number(t)))
      }
    })
    const timestamps = [...allTimestamps].sort((a, b) => a - b)
    if (!timestamps.length) return null

    const timeMap = new Map(timestamps.map((t, i) => [t, i]))
    const series = rows.map((row, index) => {
      const values = new Float64Array(timestamps.length).fill(NaN)
      if (Array.isArray(row?.values)) {
        row.values.forEach(([t, v]) => {
          const idx = timeMap.get(Number(t))
          if (idx !== undefined) values[idx] = Number(v)
        })
      }
      return { name: seriesName(row, index), values, color: colors[index % colors.length] }
    })

    return { timestamps, series }
  }, [result])

  useEffect(() => {
    if (!data || !containerRef.current) return
    if (chartRef.current) { chartRef.current.destroy(); chartRef.current = null }

    const { timestamps, series } = data
    const uData = [timestamps, ...series.map((s) => Array.from(s.values))]

    const opts = {
      width: containerRef.current.clientWidth || 760,
      height: 280,
      cursor: { drag: { x: true, y: false } },
      select: { show: true },
      hooks: {
        setSelect: [
          (u) => {
            const min = u.posToVal(u.select.left, 'x')
            const max = u.posToVal(u.select.left + u.select.width, 'x')
            if (max - min > 1 && onBrush) onBrush(min, max)
            u.setSelect({ left: 0, width: 0, top: 0, height: 0 }, false)
          },
        ],
      },
      series: [
        { label: 'Time' },
        ...series.map((s) => ({
          label: s.name,
          stroke: s.color,
          width: 2,
          fill: mode === 'area' ? s.color + '1a' : undefined,
        })),
      ],
      axes: [
        { stroke: '#8c99a8', grid: { stroke: '#e8ecf0' } },
        { stroke: '#8c99a8', grid: { stroke: '#e8ecf0' } },
      ],
    }

    chartRef.current = new uPlot(opts, uData, containerRef.current)

    const handleResize = () => {
      if (chartRef.current && containerRef.current) {
        chartRef.current.setSize({ width: containerRef.current.clientWidth, height: 280 })
      }
    }
    window.addEventListener('resize', handleResize)
    return () => {
      window.removeEventListener('resize', handleResize)
      if (chartRef.current) { chartRef.current.destroy(); chartRef.current = null }
    }
  }, [data, mode, onBrush])

  if (!result) return <div className='fx-query-graph-empty'>执行区间查询后渲染图表。</div>
  if (!data) return <div className='fx-query-graph-empty'>查询契约未返回可绘制的图表点。</div>

  return (
    <div className='fx-query-graph'>
      <div ref={containerRef} style={{ width: '100%' }} />
      {showLegend && data && (
        <div className='fx-query-legend'>
          {data.series.map((item) => (
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
