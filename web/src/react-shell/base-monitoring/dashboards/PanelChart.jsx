import React, { useEffect, useRef, useState, useCallback } from 'react'
import * as d3 from 'd3'
import { queryPanel } from './panelQuery.js'

const COLORS = [
  'var(--fx-blue)',
  '#e6550d',
  '#31a354',
  '#756bb1',
  '#636363',
  '#6baed6',
  '#fd8d3c',
  '#74c476',
  '#9e9ac8',
  '#969696',
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

function TimeSeriesChart({ series }) {
  const containerRef = useRef(null)
  const svgRef = useRef(null)

  const draw = useCallback(() => {
    const container = containerRef.current
    const svg = svgRef.current
    if (!container || !svg || series.length === 0) return

    const width = container.clientWidth || 400
    const height = 200
    const margin = { top: 10, right: 12, bottom: 30, left: 50 }
    const innerW = width - margin.left - margin.right
    const innerH = height - margin.top - margin.bottom

    const allValues = series.flatMap((s) => s.values)
    const xExtent = d3.extent(allValues, (d) => d[0] * 1000)
    const yExtent = d3.extent(allValues, (d) => d[1])
    if (!xExtent[0] || !yExtent[0] === undefined) return

    const yPad = (yExtent[1] - yExtent[0]) * 0.1 || 1
    const xScale = d3.scaleTime().domain(xExtent).range([0, innerW])
    const yScale = d3.scaleLinear().domain([yExtent[0] - yPad, yExtent[1] + yPad]).range([innerH, 0])

    const sel = d3.select(svg)
    sel.selectAll('*').remove()
    sel.attr('width', width).attr('height', height)

    const g = sel.append('g').attr('transform', `translate(${margin.left},${margin.top})`)

    // Grid lines
    g.append('g')
      .attr('transform', `translate(0,${innerH})`)
      .call(d3.axisBottom(xScale).ticks(5).tickSize(-innerH).tickFormat(d3.timeFormat('%H:%M')))
      .selectAll('line').attr('stroke', 'var(--fx-border)')
    g.append('g')
      .call(d3.axisLeft(yScale).ticks(5).tickSize(-innerW))
      .selectAll('line').attr('stroke', 'var(--fx-border)')

    // Style axis text
    sel.selectAll('.domain').attr('stroke', 'var(--fx-border)')
    sel.selectAll('text').attr('fill', 'var(--fx-ink)').style('font-size', '10px')

    // Lines
    const line = d3.line().x((d) => xScale(d[0] * 1000)).y((d) => yScale(d[1])).curve(d3.curveMonotoneX)
    series.forEach((s, i) => {
      g.append('path')
        .datum(s.values)
        .attr('fill', 'none')
        .attr('stroke', COLORS[i % COLORS.length])
        .attr('stroke-width', 1.5)
        .attr('d', line)
    })
  }, [series])

  useEffect(() => {
    draw()
    const observer = new ResizeObserver(() => draw())
    if (containerRef.current) observer.observe(containerRef.current)
    return () => observer.disconnect()
  }, [draw])

  if (series.length === 0) {
    return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>无数据</div>
  }

  return (
    <div ref={containerRef} style={{ width: '100%' }}>
      <svg ref={svgRef} style={{ width: '100%', height: '200px', display: 'block' }} />
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', padding: '6px 0', fontSize: '11px' }}>
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
  if (series.length === 0) return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>无数据</div>
  const lastValues = series[0].values
  const value = lastValues.length > 0 ? lastValues[lastValues.length - 1][1] : 0
  const formatted = Number.isFinite(value) ? (value % 1 === 0 ? value.toString() : value.toFixed(2)) : '-'
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '160px' }}>
      <span style={{ fontSize: '42px', fontWeight: 700, color: 'var(--fx-blue)' }}>{formatted}</span>
      <span style={{ fontSize: '12px', color: 'var(--fx-muted)', marginTop: '4px' }}>{formatMetricLabel(series[0])}</span>
    </div>
  )
}

function TablePanel({ series }) {
  if (series.length === 0) return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>无数据</div>
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
export default function PanelChart({ panel, timeRange, datasourceId }) {
  const [series, setSeries] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    // panel.raw holds the original panel data with targets
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

  // No query configured
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
  if (type === 'table' || type === 'table-old') {
    return <TablePanel series={series} />
  }
  if (type === 'timeseries' || type === 'graph' || type === '') {
    return <TimeSeriesChart series={series} />
  }

  return <div style={{ color: 'var(--fx-muted)', padding: '20px', textAlign: 'center' }}>{panel.type} 暂不支持</div>
}
