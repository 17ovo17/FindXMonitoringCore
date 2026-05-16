import React, { useEffect, useRef, useState } from 'react'
import uPlot from 'uplot'
import 'uplot/dist/uPlot.min.css'
import { formatTracingError, tracingApi } from '../api/tracing.js'
import { Blocked, ErrorBox } from './TracingShared.jsx'

const METRICS = [
  { key: 'responseTime', label: '响应时间趋势', unit: 'ms', color: '#1769ff' },
  { key: 'throughput', label: '吞吐趋势', unit: 'cpm', color: '#17a86b' },
  { key: 'sla', label: 'SLA 趋势', unit: '%', color: '#f59e0b' },
  { key: 'apdex', label: 'Apdex 趋势', unit: '', color: '#8b5cf6' },
]

function generateMockTimeSeries(points = 30) {
  const now = Math.floor(Date.now() / 1000)
  const step = 60
  const timestamps = Array.from({ length: points }, (_, i) => now - (points - 1 - i) * step)
  return timestamps
}

function UPlotChart({ title, timestamps, values, unit, color }) {
  const containerRef = useRef(null)
  const chartRef = useRef(null)

  useEffect(() => {
    if (!containerRef.current || !timestamps.length) return
    const width = containerRef.current.clientWidth || 280
    const opts = {
      width,
      height: 160,
      cursor: { show: true },
      scales: { x: { time: true }, y: { auto: true } },
      axes: [
        { stroke: '#66758d', grid: { stroke: '#eef2f8' } },
        { stroke: '#66758d', grid: { stroke: '#eef2f8' }, label: unit },
      ],
      series: [
        {},
        { stroke: color, width: 2, fill: color + '18', label: title },
      ],
    }
    const data = [timestamps, values]
    if (chartRef.current) { chartRef.current.destroy(); chartRef.current = null }
    chartRef.current = new uPlot(opts, data, containerRef.current)
    return () => { if (chartRef.current) { chartRef.current.destroy(); chartRef.current = null } }
  }, [timestamps, values, unit, color, title])

  return (
    <div className='fx-metric-chart'>
      <h4 className='fx-metric-chart__title'>{title}</h4>
      <div ref={containerRef} className='fx-metric-chart__canvas' />
    </div>
  )
}
export function ServiceMetricsCharts({ serviceId }) {
  const [metricsData, setMetricsData] = useState(null)
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      setLoading(true); setError(''); setBlocked('')
      try {
        const resp = await tracingApi.overview({ serviceId, metrics: 'responseTime,throughput,sla,apdex' })
        if (!cancelled) {
          if (resp && resp.metrics) {
            setMetricsData(resp.metrics)
          } else {
            // 后端返回了数据但没有 metrics 字段，使用模拟时间轴展示 BLOCKED
            setBlocked('PENDING: 需要后端实现 /apm/overview?serviceId&metrics API 返回时序指标数据')
          }
        }
      } catch (err) {
        if (cancelled) return
        const msg = formatTracingError(err)
        if (msg.startsWith('PENDING') || [404, 405, 501].includes(err?.status)) {
          setBlocked('PENDING: 需要后端实现 /apm/overview?serviceId&metrics API 返回时序指标数据')
        } else {
          setError(msg)
        }
      } finally { if (!cancelled) setLoading(false) }
    }
    load()
    return () => { cancelled = true }
  }, [serviceId])

  if (loading) return <p style={{ color: 'var(--fx-muted)', fontSize: 13 }}>加载指标趋势...</p>
  if (error) return <ErrorBox>{error}</ErrorBox>
  if (blocked) return <Blocked>{blocked}</Blocked>

  const timestamps = metricsData?.timestamps || generateMockTimeSeries()

  return (
    <div className='fx-metric-charts-grid'>
      {METRICS.map(m => {
        const values = metricsData?.[m.key] || []
        if (!values.length) return (
          <div key={m.key} className='fx-metric-chart'>
            <h4 className='fx-metric-chart__title'>{m.label}</h4>
            <Blocked>PENDING: 需要后端实现 /apm/overview 返回 {m.key} 时序数据</Blocked>
          </div>
        )
        return <UPlotChart key={m.key} title={m.label} timestamps={timestamps} values={values} unit={m.unit} color={m.color} />
      })}
    </div>
  )
}
