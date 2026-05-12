import React, { useEffect, useRef, useState, useCallback } from 'react'
import uPlot from 'uplot'
import 'uplot/dist/uPlot.min.css'
import { queryPanel } from './panelQuery.js'
import { StatPanel, BarGaugePanel, GaugePanel } from './chartRenderers.jsx'
import { PiePanel, HexbinPanel, TextPanel, HeatmapPanel, IframePanel, BarChartPanel, TableNGPanel } from './chartRenderersExtra.jsx'
import { formatValue } from './unitFormat.js'

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

function TimeSeriesChart({ series, annotations = [], onBrushEnd, unit }) {
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

    /* DEGRADE-005: Annotations 渲染 hook */
    const drawAnnotations = (u) => {
      if (!annotations || annotations.length === 0) return
      const ctx = u.ctx
      const xMin = u.scales.x.min
      const xMax = u.scales.x.max
      for (const ann of annotations) {
        const ts = Number(ann.time || ann.timestamp)
        if (ts < xMin || ts > xMax) continue
        const px = u.valToPos(ts, 'x', true)
        ctx.save()
        ctx.strokeStyle = ann.color || '#e6550d'
        ctx.lineWidth = 1.5
        ctx.setLineDash([4, 2])
        ctx.beginPath()
        ctx.moveTo(px, u.bbox.top)
        ctx.lineTo(px, u.bbox.top + u.bbox.height)
        ctx.stroke()
        ctx.setLineDash([])
        ctx.fillStyle = ann.color || '#e6550d'
        ctx.beginPath()
        ctx.arc(px, u.bbox.top + 4, 4, 0, Math.PI * 2)
        ctx.fill()
        ctx.restore()
      }
    }

    const opts = {
      width,
      height,
      cursor: { drag: { x: true, y: false } },
      select: { show: true },
      hooks: {
        drawAxes: [drawAnnotations],
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
        {
          stroke: '#8c99a8', grid: { stroke: '#e8ecf0' }, size: 50,
          values: (u, vals) => vals.map((v) => unit ? formatValue(v, unit) : (v >= 1e6 ? (v / 1e6).toFixed(1) + 'M' : v >= 1e3 ? (v / 1e3).toFixed(1) + 'K' : String(v))),
        },
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
  }, [series, annotations, onBrushEnd, unit])

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
  const type = (panel.type || '').toLowerCase()
  const displayOpts = rawPanel.displayOptions || panel.displayOptions || {}
  const unit = rawPanel.unit || panel.unit

  /* text 和 iframe 不需要查询数据 */
  if (type === 'text') {
    return <TextPanel panel={panel} options={displayOpts} />
  }
  if (type === 'iframe') {
    return <IframePanel panel={panel} options={displayOpts} />
  }

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

  if (type === 'stat' || type === 'singlestat') {
    return <StatPanel series={series} options={{ ...displayOpts, unit }} />
  }
  if (type === 'table' || type === 'table-old') {
    return <TablePanel series={series} />
  }
  if (type === 'tableng') {
    return <TableNGPanel series={series} options={{ ...displayOpts, unit }} />
  }
  if (type === 'pie') {
    return <PiePanel series={series} options={{ ...displayOpts, unit }} />
  }
  if (type === 'hexbin') {
    return <HexbinPanel series={series} options={displayOpts} />
  }
  if (type === 'bargauge') {
    return <BarGaugePanel series={series} options={{ ...displayOpts, unit }} />
  }
  if (type === 'gauge') {
    return <GaugePanel series={series} options={{ ...displayOpts, unit }} />
  }
  if (type === 'heatmap') {
    return <HeatmapPanel series={series} options={displayOpts} />
  }
  if (type === 'barchart') {
    return <BarChartPanel series={series} options={{ ...displayOpts, unit }} />
  }
  if (type === 'timeseries' || type === 'graph' || type === '') {
    return <TimeSeriesChart series={series} annotations={annotations} onBrushEnd={onBrushEnd} unit={unit} />
  }

  return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>{panel.type} 暂不支持</div>
}
