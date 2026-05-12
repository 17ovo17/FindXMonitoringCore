import React, { useEffect, useRef, useState, useCallback } from 'react'
import uPlot from 'uplot'
import 'uplot/dist/uPlot.min.css'
import { queryPanel } from './panelQuery.js'

const COLORS = [
  '#1769ff', '#e6550d', '#31a354', '#756bb1', '#636363',
  '#6baed6', '#fd8d3c', '#74c476', '#9e9ac8', '#969696',
]

function formatMetricLabel(series) {
  if (series.legendFormat) {
    let label = series.legendFormat
    for (const [key, val] of Object.entries(series.metric || {})) {
      label = label.replace(new RegExp(`\\{\\{\\s*${key}\\s*\\}\\}`, 'g'), val)
    }
    if (!label.includes('{{')) return label
  }
  const m = series.metric || {}
  const name = m.__name__ || ''
  const labels = Object.entries(m)
    .filter(([k]) => k !== '__name__')
    .map(([k, v]) => `${k}="${v}"`)
    .join(', ')
  return labels ? `${name}{${labels}}` : name || 'series'
}

function TimeSeriesChart({ series, onBrushEnd }) {
  const containerRef = useRef(null)
  const chartRef = useRef(null)

  useEffect(() => {
    if (!containerRef.current || series.length === 0) return
    if (chartRef.current) { chartRef.current.destroy(); chartRef.current = null }

    const allTimestamps = new Set()
    series.forEach((s) => s.values.forEach(([t]) => allTimestamps.add(Number(t))))
    const timestamps = [...allTimestamps].sort((a, b) => a - b)
    if (!timestamps.length) return

    const timeMap = new Map(timestamps.map((t, i) => [t, i]))
    const uSeries = series.map((s, i) => {
      const values = new Float64Array(timestamps.length).fill(NaN)
      s.values.forEach(([t, v]) => {
        const idx = timeMap.get(Number(t))
        if (idx !== undefined) values[idx] = Number(v)
      })
      return { name: formatMetricLabel(s), values, color: COLORS[i % COLORS.length] }
    })

    const uData = [timestamps, ...uSeries.map((s) => Array.from(s.values))]
    const width = containerRef.current.clientWidth || 400
    const height = Math.min(Math.max(containerRef.current.clientHeight - 30, 120), 300)

    const opts = {
      width,
      height,
      cursor: { drag: { x: true, y: false } },
      select: { show: true },
      hooks: {
        setSelect: [
          (u) => {
            if (!u.select.width || u.select.width < 5) return
            const min = u.posToVal(u.select.left, 'x')
            const max = u.posToVal(u.select.left + u.select.width, 'x')
            u.setSelect({ left: 0, width: 0, top: 0, height: 0 }, false)
            if (max - min > 1 && onBrushEnd) {
              onBrushEnd({ start: Math.floor(min), end: Math.floor(max) })
            }
          },
        ],
      },
      series: [
        { label: 'Time' },
        ...uSeries.map((s) => ({ label: s.name, stroke: s.color, width: 1.5 })),
      ],
      axes: [
        { stroke: '#8c99a8', grid: { stroke: '#e8ecf0' }, size: 40 },
        { stroke: '#8c99a8', grid: { stroke: '#e8ecf0' }, size: 50 },
      ],
    }

    chartRef.current = new uPlot(opts, uData, containerRef.current)

    const handleResize = () => {
      if (chartRef.current && containerRef.current) {
        chartRef.current.setSize({ width: containerRef.current.clientWidth, height })
      }
    }
    window.addEventListener('resize', handleResize)
    return () => {
      window.removeEventListener('resize', handleResize)
      if (chartRef.current) { chartRef.current.destroy(); chartRef.current = null }
    }
  }, [series, onBrushEnd])

  if (series.length === 0) {
    return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>无数据</div>
  }

  return (
    <div style={{ width: '100%', height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div ref={containerRef} style={{ flex: '1 1 auto', minHeight: '120px' }} />
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', padding: '4px 0', fontSize: '11px', flexShrink: 0 }}>
        {series.map((s, i) => (
          <span key={i} style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <span style={{ width: 10, height: 3, background: COLORS[i % COLORS.length], display: 'inline-block' }} />
            <span style={{ color: 'var(--fx-ink)' }}>{formatMetricLabel(s)}</span>
          </span>
        ))}
      </div>
    </div>
  )
}

function StatPanel({ series }) {
  if (series.length === 0) {
    return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>无数据</div>
  }
  const lastValues = series[0].values
  const value = lastValues.length > 0 ? lastValues[lastValues.length - 1][1] : 0
  const formatted = Number.isFinite(value)
    ? (value % 1 === 0 ? value.toString() : value.toFixed(2))
    : '-'
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '160px' }}>
      <span style={{ fontSize: '42px', fontWeight: 700, color: 'var(--fx-blue)' }}>{formatted}</span>
      <span style={{ fontSize: '12px', color: 'var(--fx-muted)', marginTop: '4px' }}>{formatMetricLabel(series[0])}</span>
    </div>
  )
}

function TablePanel({ series }) {
  if (series.length === 0) {
    return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>无数据</div>
  }
  return (
    <div style={{ overflow: 'auto', maxHeight: '200px', fontSize: '12px' }}>
      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead>
          <tr style={{ borderBottom: '1px solid var(--fx-border)' }}>
            <th style={{ textAlign: 'left', padding: '4px 8px', color: 'var(--fx-ink)' }}>指标</th>
            <th style={{ textAlign: 'right', padding: '4px 8px', color: 'var(--fx-ink)' }}>最新值</th>
          </tr>
        </thead>
        <tbody>
          {series.map((s, i) => {
            const last = s.values.length > 0 ? s.values[s.values.length - 1][1] : '-'
            const val = Number.isFinite(last) ? (last % 1 === 0 ? last.toString() : last.toFixed(4)) : '-'
            return (
              <tr key={i} style={{ borderBottom: '1px solid var(--fx-border)' }}>
                <td style={{ padding: '4px 8px', color: 'var(--fx-ink)' }}>{formatMetricLabel(s)}</td>
                <td style={{ padding: '4px 8px', textAlign: 'right', color: 'var(--fx-ink)' }}>{val}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

export default function PanelChart({ panel, timeRange, datasourceId, annotations, onBrushEnd }) {
  const [series, setSeries] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    const rawPanel = panel.raw || panel
    const hasTargets = (Array.isArray(rawPanel.targets) && rawPanel.targets.some((t) => t.expr || t.expression || t.query))
      || rawPanel.expr || rawPanel.query || rawPanel.expression || rawPanel.metric

    if (!hasTargets) {
      setError(null)
      setSeries([])
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    queryPanel(rawPanel, timeRange, datasourceId).then((result) => {
      if (cancelled) return
      if (result.error) setError(result.error)
      setSeries(result.series)
    }).finally(() => {
      if (!cancelled) setLoading(false)
    })

    return () => { cancelled = true }
  }, [panel, timeRange, datasourceId])

  const rawPanel = panel.raw || panel
  const hasTargets = (Array.isArray(rawPanel.targets) && rawPanel.targets.some((t) => t.expr || t.expression || t.query))
    || rawPanel.expr || rawPanel.query || rawPanel.expression || rawPanel.metric
  if (!hasTargets) {
    return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>未配置查询</div>
  }

  if (loading) {
    return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>加载中...</div>
  }

  if (error) {
    return <div style={{ color: '#c0392b', padding: '12px', fontSize: '12px', background: 'rgba(192,57,43,0.06)', borderRadius: '4px' }}>{error}</div>
  }

  const type = (panel.type || '').toLowerCase()

  if (type === 'stat' || type === 'singlestat') {
    return <StatPanel series={series} />
  }
  if (type === 'table' || type === 'table-old' || type === 'tableng') {
    return <TablePanel series={series} />
  }
  if (type === 'timeseries' || type === 'graph' || type === '') {
    return <TimeSeriesChart series={series} annotations={annotations} onBrushEnd={onBrushEnd} />
  }

  return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>{panel.type} 暂不支持</div>
}
