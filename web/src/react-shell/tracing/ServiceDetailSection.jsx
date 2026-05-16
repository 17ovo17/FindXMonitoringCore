import React, { useCallback, useEffect, useState } from 'react'
import { get } from '../api/http.js'
import { formatTracingError } from '../api/tracing.js'
import { displayText } from './tracingModel.js'
import { Blocked, Empty, ErrorBox, Field, Status } from './TracingShared.jsx'
import { ServiceMetricsCharts } from './ServiceMetricsCharts.jsx'
import { FlameGraph } from './FlameGraph.jsx'

const TABS = [
  { key: 'instances', label: '实例' },
  { key: 'endpoints', label: '端点' },
  { key: 'metrics', label: '指标趋势' },
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

export function ServiceDetailSection({ serviceName, onClose, onNavigate }) {
  const [tab, setTab] = useState('instances')
  const [detail, setDetail] = useState(null)
  const [instances, setInstances] = useState([])
  const [endpoints, setEndpoints] = useState([])
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
        const resp = await get(`/tracing/services/${encodeURIComponent(serviceName)}`)
        if (!cancelled) setDetail(resp)
      } catch (err) {
        if (!cancelled) setError(formatTracingError(err))
      } finally { if (!cancelled) setLoading(false) }
    }
    load()
    return () => { cancelled = true }
  }, [serviceName])

  const loadEndpoints = useCallback(async () => {
    try {
      const resp = await get(`/tracing/services/${encodeURIComponent(serviceName)}/endpoints`)
      const list = Array.isArray(resp) ? resp : resp?.items || resp?.data || resp?.list || []
      setEndpoints(list)
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
    if (tab === 'endpoints') loadEndpoints()
    if (tab === 'profiling') loadProfiling()
  }, [tab, loadEndpoints, loadProfiling])

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

        {tab === 'instances' && <InstancesTab instances={instances} />}
        {tab === 'endpoints' && <EndpointsTab endpoints={endpoints} />}
        {tab === 'metrics' && <MetricsTab serviceName={serviceName} />}
        {tab === 'profiling' && <ProfilingTab tasks={profilingTasks} flameData={flameData} onLoadFlame={loadFlameGraph} />}
      </div>
    </div>
  )
}

function InstancesTab({ instances }) {
  if (!instances.length) return <Empty>暂无实例数据</Empty>
  return (
    <div className='fx-tracing-table'>
      <table>
        <thead>
          <tr><th>实例 ID</th><th>主机</th><th>状态</th><th>延迟</th><th>吞吐</th></tr>
        </thead>
        <tbody>
          {instances.map((inst, i) => (
            <tr key={inst.instance_id || inst.id || i}>
              <td>{inst.instance_id || inst.id || '-'}</td>
              <td>{inst.host || inst.ip || '-'}</td>
              <td><Status ok={inst.status === 'healthy' || inst.status === 'running'}>{inst.status || '-'}</Status></td>
              <td>{fmtMs(inst.latency ?? inst.avg_latency)}</td>
              <td>{fmtCpm(inst.throughput ?? inst.cpm)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function EndpointsTab({ endpoints }) {
  if (!endpoints.length) return <Empty>暂无端点数据</Empty>
  return (
    <div className='fx-tracing-table'>
      <table>
        <thead>
          <tr><th>端点名称</th><th>P50</th><th>P90</th><th>P99</th><th>错误率</th><th>CPM</th></tr>
        </thead>
        <tbody>
          {endpoints.map((ep, i) => (
            <tr key={ep.id || ep.name || i}>
              <td style={{ fontWeight: 600 }}>{ep.name || ep.endpoint || '-'}</td>
              <td>{fmtMs(ep.p50 ?? ep.latency_p50)}</td>
              <td>{fmtMs(ep.p90 ?? ep.latency_p90)}</td>
              <td>{fmtMs(ep.p99 ?? ep.latency_p99)}</td>
              <td>{fmtRate(ep.error_rate ?? ep.errorRate)}</td>
              <td>{fmtCpm(ep.cpm ?? ep.throughput)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function MetricsTab({ serviceName }) {
  return <ServiceMetricsCharts serviceId={serviceName} />
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
