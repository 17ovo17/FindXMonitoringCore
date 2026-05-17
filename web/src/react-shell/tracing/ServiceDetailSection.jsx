import React, { useCallback, useEffect, useRef, useState } from 'react'
import uPlot from 'uplot'
import 'uplot/dist/uPlot.min.css'
import { get } from '../api/http.js'
import { formatTracingError } from '../api/tracing.js'
import { displayText } from './tracingModel.js'
import { Blocked, Empty, ErrorBox, Field, Status } from './TracingShared.jsx'
import { ServiceMetricsCharts } from './ServiceMetricsCharts.jsx'
import { FlameGraph } from './FlameGraph.jsx'

const TABS = [
  { key: 'instances', label: '实例列表' },
  { key: 'endpoints', label: '端点列表' },
  { key: 'metrics', label: '服务指标' },
  { key: 'profiling', label: 'Profiling' },
]

function slaBadge(sla) {
  if (sla === null || sla === undefined) return null
  const v = sla <= 1 ? sla * 100 : sla
  const cls = v >= 99.9 ? 'is-good' : v >= 99 ? 'is-warn' : 'is-bad'
  return <span className={`fx-sla-badge ${cls}`}>{v.toFixed(2)}%</span>
}

function fmtMs(v) { return v === null || v === undefined ? '-' : Number(v).toFixed(1) + ' ms' }
function fmtRate(v) { return v === null || v === undefined ? '-' : (Number(v) * 100).toFixed(2) + '%' }
function fmtCpm(v) { return v === null || v === undefined ? '-' : Number(v).toFixed(0) + ' cpm' }
function fmtApdex(v) { return v === null || v === undefined ? '-' : Number(v).toFixed(3) }

export function ServiceDetailSection({ serviceName, onClose, onNavigate }) {
  const [tab, setTab] = useState('instances')
  const [detail, setDetail] = useState(null)
  const [instances, setInstances] = useState([])
  const [endpoints, setEndpoints] = useState([])
  const [metrics, setMetrics] = useState(null)
  const [profilingTasks, setProfilingTasks] = useState([])
  const [flameData, setFlameData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      setLoading(true)
      setError('')
      try {
        const resp = await get(`/apm/services/${encodeURIComponent(serviceName)}`)
        if (!cancelled) setDetail(resp)
      } catch (err) {
        if (!cancelled) setError(formatTracingError(err))
      } finally { if (!cancelled) setLoading(false) }
    }
    load()
    return () => { cancelled = true }
  }, [serviceName])

  const loadInstances = useCallback(async () => {
    try {
      const resp = await get(`/apm/services/${encodeURIComponent(serviceName)}/instances`)
      const list = Array.isArray(resp) ? resp : resp?.items || resp?.data || resp?.list || []
      setInstances(list)
    } catch (err) { setError(formatTracingError(err)) }
  }, [serviceName])

  const loadEndpoints = useCallback(async () => {
    try {
      const resp = await get(`/apm/services/${encodeURIComponent(serviceName)}/endpoints`)
      const list = Array.isArray(resp) ? resp : resp?.items || resp?.data || resp?.list || []
      setEndpoints(list)
    } catch (err) { setError(formatTracingError(err)) }
  }, [serviceName])

  const loadMetrics = useCallback(async () => {
    try {
      const resp = await get(`/apm/services/${encodeURIComponent(serviceName)}/metrics`)
      setMetrics(resp)
    } catch (err) { setError(formatTracingError(err)) }
  }, [serviceName])

  const loadProfiling = useCallback(async () => {
    try {
      const resp = await get('/apm/profiling/tasks', { params: { service: serviceName } })
      const list = Array.isArray(resp) ? resp : resp?.items || resp?.data || resp?.list || []
      setProfilingTasks(list)
    } catch (err) { setError(formatTracingError(err)) }
  }, [serviceName])

  const loadFlameGraph = useCallback(async (taskId) => {
    setFlameData(null)
    try {
      const resp = await get(`/tracing/profiling/tasks/${encodeURIComponent(taskId)}`)
      setFlameData(resp?.flame_graph || resp?.flameGraph || resp?.data || null)
    } catch (err) { setError(formatTracingError(err)) }
  }, [])

  useEffect(() => {
    if (tab === 'instances') loadInstances()
    if (tab === 'endpoints') loadEndpoints()
    if (tab === 'metrics') loadMetrics()
    if (tab === 'profiling') loadProfiling()
  }, [tab, loadInstances, loadEndpoints, loadMetrics, loadProfiling])

  useEffect(() => {
    if (detail?.instances) setInstances(detail.instances)
  }, [detail])

  const sla = detail?.sla ?? detail?.success_rate ?? detail?.successRate ?? null

  return (
    <div className='fx-tracing-modal'>
      <div className='fx-tracing-modal__body'>
        <header>
          <h2>
            {displayText(serviceName)} · 服务详情
            {detail?.type && <small style={{ marginLeft: 8, fontWeight: 400, color: 'var(--fx-text-weak, #66758d)' }}>{detail.type}</small>}
            {slaBadge(sla)}
          </h2>
          <button type='button' onClick={onClose}>关闭</button>
        </header>
        <ErrorBox>{error}</ErrorBox>
        {loading && <p style={{ color: 'var(--fx-muted, #66758d)', fontSize: 13 }}>加载中...</p>}

        <nav className='fx-tracing-tabs'>
          {TABS.map(t => (
            <button key={t.key} type='button' className={tab === t.key ? 'is-active' : ''} onClick={() => setTab(t.key)}>
              {t.label}
            </button>
          ))}
        </nav>

        {tab === 'instances' && <InstancesTab instances={instances} serviceName={serviceName} />}
        {tab === 'endpoints' && <EndpointsTab endpoints={endpoints} />}
        {tab === 'metrics' && <ServiceMetricsTab serviceName={serviceName} metrics={metrics} />}
        {tab === 'profiling' && <ProfilingTab tasks={profilingTasks} flameData={flameData} onLoadFlame={loadFlameGraph} />}
      </div>
    </div>
  )
}

function InstancesTab({ instances, serviceName }) {
  const [expanded, setExpanded] = useState({})
  const [instanceMetrics, setInstanceMetrics] = useState({})

  const toggleExpand = async (inst) => {
    const id = inst.instance_id || inst.id
    if (expanded[id]) {
      setExpanded(prev => ({ ...prev, [id]: false }))
      return
    }
    setExpanded(prev => ({ ...prev, [id]: true }))
    if (!instanceMetrics[id]) {
      try {
        const resp = await get(`/apm/services/${encodeURIComponent(serviceName)}/instances/${encodeURIComponent(id)}/metrics`)
        setInstanceMetrics(prev => ({ ...prev, [id]: resp }))
      } catch (_) {
        setInstanceMetrics(prev => ({ ...prev, [id]: { error: true } }))
      }
    }
  }

  if (!instances.length) return <Empty>暂无实例数据</Empty>
  return (
    <div className='fx-tracing-table'>
      <table>
        <thead>
          <tr>
            <th style={{ width: 32 }}></th>
            <th>实例名</th>
            <th>语言</th>
            <th>OS</th>
            <th>JVM 版本</th>
            <th>启动时间</th>
            <th>状态</th>
          </tr>
        </thead>
        <tbody>
          {instances.map((inst, i) => {
            const id = inst.instance_id || inst.id || i
            const isExpanded = expanded[id]
            const metrics = instanceMetrics[id]
            return (
              <React.Fragment key={id}>
                <tr onClick={() => toggleExpand(inst)} style={{ cursor: 'pointer' }}>
                  <td style={{ textAlign: 'center', fontSize: 11 }}>{isExpanded ? '▼' : '▶'}</td>
                  <td style={{ fontWeight: 600 }}>{inst.name || inst.instance_id || inst.id || '-'}</td>
                  <td>{inst.language || inst.lang || '-'}</td>
                  <td>{inst.os_name || inst.os || '-'}</td>
                  <td>{inst.jvm_version || inst.runtime_version || '-'}</td>
                  <td>{inst.start_time || inst.startTime || inst.created_at || '-'}</td>
                  <td><Status ok={inst.status === 'healthy' || inst.status === 'running'}>{inst.status || '-'}</Status></td>
                </tr>
                {isExpanded && (
                  <tr>
                    <td colSpan={7} style={{ padding: 0 }}>
                      <InstanceMetricsPanel metrics={metrics} />
                    </td>
                  </tr>
                )}
              </React.Fragment>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

function InstanceMetricsPanel({ metrics }) {
  if (!metrics) return <p style={{ padding: '8px 16px', color: 'var(--fx-muted)', fontSize: 12 }}>加载实例指标...</p>
  if (metrics.error) return <p style={{ padding: '8px 16px', color: 'var(--fx-danger, #c45656)', fontSize: 12 }}>加载实例指标失败</p>

  const cards = [
    { label: 'CPU 使用率', value: metrics.cpu != null ? (metrics.cpu * 100).toFixed(1) + '%' : '-' },
    { label: 'Heap 使用', value: metrics.heap_used != null ? (metrics.heap_used / 1024 / 1024).toFixed(0) + ' MB' : '-' },
    { label: 'Heap 最大', value: metrics.heap_max != null ? (metrics.heap_max / 1024 / 1024).toFixed(0) + ' MB' : '-' },
    { label: 'GC 次数', value: metrics.gc_count ?? '-' },
    { label: 'GC 耗时', value: metrics.gc_time != null ? metrics.gc_time + ' ms' : '-' },
    { label: '线程数', value: metrics.thread_count ?? metrics.threads ?? '-' },
  ]

  return (
    <div style={{ display: 'flex', gap: 12, padding: '10px 16px', background: 'var(--fx-bg-subtle, #f8fbff)', borderTop: '1px solid var(--fx-border, #e3e8f1)' }}>
      {cards.map(c => (
        <div key={c.label} style={{ flex: '1 1 0', textAlign: 'center' }}>
          <div style={{ fontSize: 11, color: 'var(--fx-text-weak, #66758d)' }}>{c.label}</div>
          <div style={{ fontSize: 14, fontWeight: 700, marginTop: 2 }}>{c.value}</div>
        </div>
      ))}
    </div>
  )
}

function EndpointsTab({ endpoints }) {
  const [search, setSearch] = useState('')
  const [sortKey, setSortKey] = useState('cpm')
  const [sortDir, setSortDir] = useState('desc')

  const toggleSort = (key) => {
    if (sortKey === key) setSortDir(prev => prev === 'asc' ? 'desc' : 'asc')
    else { setSortKey(key); setSortDir('desc') }
  }

  const sortArrow = (key) => sortKey === key ? (sortDir === 'asc' ? ' ↑' : ' ↓') : ''

  const filtered = endpoints.filter(ep => {
    if (!search) return true
    const name = (ep.name || ep.endpoint || '').toLowerCase()
    return name.includes(search.toLowerCase())
  })

  const sorted = [...filtered].sort((a, b) => {
    const getVal = (item) => {
      switch (sortKey) {
        case 'cpm': return Number(item.cpm ?? item.throughput ?? 0)
        case 'avg': return Number(item.avg_latency ?? item.latency ?? 0)
        case 'p95': return Number(item.p95 ?? item.latency_p95 ?? 0)
        case 'error': return Number(item.error_rate ?? item.errorRate ?? 0)
        default: return 0
      }
    }
    const va = getVal(a), vb = getVal(b)
    return sortDir === 'asc' ? va - vb : vb - va
  })

  if (!endpoints.length) return <Empty>暂无端点数据</Empty>
  return (
    <div className='fx-tracing-table'>
      <div style={{ marginBottom: 8 }}>
        <input
          style={{ padding: '6px 10px', border: '1px solid var(--fx-border, #e3e8f1)', borderRadius: 6, fontSize: 13, width: 260 }}
          value={search}
          onChange={e => setSearch(e.target.value)}
          placeholder='搜索端点名称...'
        />
      </div>
      <table>
        <thead>
          <tr>
            <th>端点名称</th>
            <th style={{ cursor: 'pointer' }} onClick={() => toggleSort('cpm')}>CPM{sortArrow('cpm')}</th>
            <th style={{ cursor: 'pointer' }} onClick={() => toggleSort('avg')}>平均延迟{sortArrow('avg')}</th>
            <th style={{ cursor: 'pointer' }} onClick={() => toggleSort('p95')}>P95 延迟{sortArrow('p95')}</th>
            <th style={{ cursor: 'pointer' }} onClick={() => toggleSort('error')}>错误率{sortArrow('error')}</th>
          </tr>
        </thead>
        <tbody>
          {sorted.map((ep, i) => (
            <tr key={ep.id || ep.name || i}>
              <td style={{ fontWeight: 600 }}>{ep.name || ep.endpoint || '-'}</td>
              <td>{fmtCpm(ep.cpm ?? ep.throughput)}</td>
              <td>{fmtMs(ep.avg_latency ?? ep.latency ?? ep.p50 ?? ep.latency_p50)}</td>
              <td>{fmtMs(ep.p95 ?? ep.latency_p95)}</td>
              <td>{fmtRate(ep.error_rate ?? ep.errorRate)}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {sorted.length === 0 && search && <Empty>未找到匹配的端点</Empty>}
    </div>
  )
}

function ServiceMetricsTab({ serviceName, metrics }) {
  return (
    <div>
      <ServiceMetricsCharts serviceId={serviceName} />
      {metrics && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12, marginTop: 16 }}>
          <MetricCard label='CPM' value={fmtCpm(metrics.cpm)} />
          <MetricCard label='平均延迟' value={fmtMs(metrics.avg_latency ?? metrics.latency)} />
          <MetricCard label='错误率' value={fmtRate(metrics.error_rate ?? metrics.errorRate)} />
          <MetricCard label='Apdex' value={fmtApdex(metrics.apdex)} />
        </div>
      )}
      {metrics && <LatencyPercentilesChart metrics={metrics} />}
    </div>
  )
}

function MetricCard({ label, value }) {
  return (
    <div style={{ padding: '12px 16px', background: 'var(--fx-bg-subtle, #f8fbff)', borderRadius: 8, border: '1px solid var(--fx-border, #e3e8f1)', textAlign: 'center' }}>
      <div style={{ fontSize: 11, color: 'var(--fx-text-weak, #66758d)', marginBottom: 4 }}>{label}</div>
      <div style={{ fontSize: 18, fontWeight: 700 }}>{value}</div>
    </div>
  )
}

function LatencyPercentilesChart({ metrics }) {
  const containerRef = useRef(null)
  const chartRef = useRef(null)

  useEffect(() => {
    if (!containerRef.current) return
    const timestamps = metrics.timestamps || []
    const p50 = metrics.p50 || metrics.latency_p50 || []
    const p75 = metrics.p75 || metrics.latency_p75 || []
    const p95 = metrics.p95 || metrics.latency_p95 || []
    const p99 = metrics.p99 || metrics.latency_p99 || []
    if (!timestamps.length) return

    const width = containerRef.current.clientWidth || 600
    const opts = {
      width,
      height: 200,
      cursor: { show: true },
      scales: { x: { time: true }, y: { auto: true } },
      axes: [
        { stroke: '#66758d', grid: { stroke: '#eef2f8' } },
        { stroke: '#66758d', grid: { stroke: '#eef2f8' }, label: 'ms' },
      ],
      series: [
        {},
        { stroke: '#60a5fa', width: 2, label: 'P50' },
        { stroke: '#34d399', width: 2, label: 'P75' },
        { stroke: '#f59e0b', width: 2, label: 'P95' },
        { stroke: '#ef4444', width: 2, label: 'P99' },
      ],
    }
    const data = [timestamps, p50, p75, p95, p99]
    if (chartRef.current) { chartRef.current.destroy(); chartRef.current = null }
    chartRef.current = new uPlot(opts, data, containerRef.current)
    return () => { if (chartRef.current) { chartRef.current.destroy(); chartRef.current = null } }
  }, [metrics])

  if (!metrics.timestamps?.length) return null
  return (
    <div style={{ marginTop: 16 }}>
      <h4 style={{ fontSize: 13, marginBottom: 8 }}>延迟百分位趋势 (P50/P75/P95/P99)</h4>
      <div ref={containerRef} />
    </div>
  )
}

function ProfilingTab({ tasks, flameData, onLoadFlame }) {
  const [selectedTask, setSelectedTask] = useState(null)

  const handleSelect = (task) => {
    setSelectedTask(task)
    onLoadFlame(task.id || task.task_id)
  }

  return (
    <div className='fx-profiling-section'>
      <div className='fx-tracing-table'>
        <h4>Profiling 任务列表</h4>
        <table>
          <thead>
            <tr><th>任务 ID</th><th>类型</th><th>状态</th><th>创建时间</th><th>操作</th></tr>
          </thead>
          <tbody>
            {tasks.map((t, i) => (
              <tr key={t.id || t.task_id || i}>
                <td>{t.id || t.task_id || '-'}</td>
                <td>{t.type || t.profile_type || '-'}</td>
                <td><Status ok={t.status === 'completed' || t.status === 'success'}>{t.status || '-'}</Status></td>
                <td>{t.created_at || t.createdAt || '-'}</td>
                <td>
                  <button type='button' onClick={() => handleSelect(t)} disabled={t.status !== 'completed' && t.status !== 'success'}>
                    查看火焰图
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!tasks.length && <Empty>暂无 Profiling 任务</Empty>}
      </div>

      {selectedTask && (
        <div className='fx-flame-section'>
          <h4>火焰图 - {selectedTask.id || selectedTask.task_id}</h4>
          {flameData ? (
            <FlameGraph data={flameData} width={760} />
          ) : (
            <p style={{ color: 'var(--fx-text-weak, #66758d)', fontSize: 13 }}>加载火焰图数据...</p>
          )}
        </div>
      )}
    </div>
  )
}
