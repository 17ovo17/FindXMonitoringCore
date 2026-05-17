import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { formatTracingError, tracingApi } from '../api/tracing.js'
import { displayText, durationPresets, durationToRange, entityOptions, layerOptions, rowText } from './tracingModel.js'
import { AgentEvidenceNotice, AgentLinkActions, Blocked, Empty, ErrorBox, Field, Status } from './TracingShared.jsx'
import { ServiceMetricsCharts } from './ServiceMetricsCharts.jsx'
import { ServiceDetailSection } from './ServiceDetailSection.jsx'

const CONNECTION_HINT = [
  '暂无服务数据。请确认链路监控上游服务已启动并配置了正确的接入地址。',
  '',
  '接入步骤:',
  '1. 在 FindX 平台设置中配置链路查询服务地址',
  '2. 在目标应用中集成 FindX Agent',
  '3. 启动应用后等待 1-2 分钟, 服务将自动注册',
].join('\n')

// Upstream Layer concept: infrastructure dimension such as GENERAL/Kubernetes/VM/Service Mesh.
const LAYER_GROUPS = [
  { group: '通用', layers: ['GENERAL', 'Service'] },
  { group: 'Kubernetes', layers: ['K8S_SERVICE', 'MESH', 'MESH_DP', 'MESH_CP'] },
  { group: '虚拟机/主机', layers: ['OS_LINUX', 'OS_WINDOWS', 'VIRTUAL_MACHINE'] },
  { group: '数据存储', layers: ['Database', 'Cache', 'MQ'] },
  { group: '前端/网关', layers: ['Browser', 'Gateway', 'FAAS'] },
]

function pickNumber(row, keys) {
  for (const k of keys) {
    const v = row && row[k]
    if (v !== undefined && v !== null && v !== '') return Number(v)
  }
  return null
}

function fmtLatency(row) {
  const v = pickNumber(row, ['avgResponseTime', 'avg_latency', 'latency', 'responseTime'])
  return v === null ? '-' : v.toFixed(0) + ' ms'
}
function fmtSuccessRate(row) {
  let v = pickNumber(row, ['successRate', 'success_rate', 'sla'])
  if (v === null) {
    const err = pickNumber(row, ['errorRate', 'error_rate'])
    if (err !== null) v = 1 - err
  }
  if (v === null) return '-'
  if (v <= 1) v = v * 100
  return v.toFixed(2) + '%'
}
function fmtCpm(row) {
  const v = pickNumber(row, ['cpm', 'throughput', 'calls'])
  return v === null ? '-' : v.toFixed(0) + ' cpm'
}
function fmtApdex(row) {
  const v = pickNumber(row, ['apdex', 'apdexScore'])
  if (v === null) return '-'
  return (v > 1 ? v / 10000 : v).toFixed(2)
}
function statusOf(row) {
  if (row.isError || row.status === 'unhealthy') return { ok: false, text: '异常' }
  if (row.status === 'healthy' || row.normal === true) return { ok: true, text: '正常' }
  return { ok: true, text: displayText(row.status || row.health || '正常') }
}

function ServiceDetail({ service, onClose, onNavigate }) {
  const [endpoints, setEndpoints] = useState([])
  const [instances, setInstances] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const serviceId = service.id || service.value || ''

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      setLoading(true); setError('')
      try {
        const [ep, inst] = await Promise.all([
          tracingApi.selectors.endpoints({ serviceId }),
          tracingApi.selectors.instances({ serviceId }),
        ])
        if (!cancelled) { setEndpoints(ep || []); setInstances(inst || []) }
      } catch (err) { if (!cancelled) setError(formatTracingError(err)) }
      finally { if (!cancelled) setLoading(false) }
    }
    load()
    return () => { cancelled = true }
  }, [serviceId])

  const avgLat = fmtLatency(service)
  const sla = fmtSuccessRate(service)
  const cpm = fmtCpm(service)
  const apdex = fmtApdex(service)

  return (
    <div className='fx-tracing-modal'>
      <div className='fx-tracing-modal__body'>
        <header>
          <h2>{displayText(service.label || service.name || service.shortName)} · 服务详情</h2>
          <button type='button' onClick={onClose}>关闭</button>
        </header>
        {error && <ErrorBox>{error}</ErrorBox>}
        {loading && <p style={{ color: 'var(--fx-muted, #66758d)', fontSize: 13 }}>加载中...</p>}

        <div className='fx-tracing-metric-grid'>
          <div className='fx-metric'><span className='fx-label'>平均响应时间</span><span className='fx-value'>{avgLat}</span></div>
          <div className='fx-metric'><span className='fx-label'>成功率 SLA</span><span className='fx-value'>{sla}</span></div>
          <div className='fx-metric'><span className='fx-label'>吞吐 CPM</span><span className='fx-value'>{cpm}</span></div>
          <div className='fx-metric'><span className='fx-label'>Apdex</span><span className='fx-value'>{apdex}</span></div>
        </div>

        <ServiceMetricsCharts serviceId={serviceId} />

        <div className='fx-tracing-split'>
          <div>
            <div className='fx-tracing-table'>
              <h3>端点列表 ({endpoints.length})</h3>
              <table><thead><tr><th>端点名称</th><th>平均延迟</th><th>成功率</th><th>吞吐</th><th>操作</th></tr></thead>
                <tbody>{endpoints.map(ep => (
                  <tr key={ep.id || ep.value || ep.label}>
                    <td>{displayText(ep.label || ep.name)}</td>
                    <td>{fmtLatency(ep)}</td>
                    <td>{fmtSuccessRate(ep)}</td>
                    <td>{fmtCpm(ep)}</td>
                    <td className='fx-tracing-actions'>
                      <button type='button' onClick={() => { onClose(); onNavigate({ section: 'traces', serviceId, endpointId: ep.id || ep.value }) }}>查看 Trace</button>
                    </td>
                  </tr>
                ))}</tbody>
              </table>
              {!endpoints.length && !loading && <Empty>暂无端点数据</Empty>}
            </div>
          </div>
          <div>
            <div className='fx-tracing-table'>
              <h3>实例列表 ({instances.length})</h3>
              <table><thead><tr><th>实例名称</th><th>语言</th><th>平均延迟</th><th>成功率</th><th>吞吐</th><th>操作</th></tr></thead>
                <tbody>{instances.map(inst => (
                  <tr key={inst.id || inst.value || inst.label}>
                    <td>{displayText(inst.label || inst.name)}</td>
                    <td>{displayText(inst.language || inst.lang)}</td>
                    <td>{fmtLatency(inst)}</td>
                    <td>{fmtSuccessRate(inst)}</td>
                    <td>{fmtCpm(inst)}</td>
                    <td className='fx-tracing-actions'>
                      <button type='button' onClick={() => { onClose(); onNavigate({ section: 'traces', serviceId, instanceId: inst.id || inst.value }) }}>查看 Trace</button>
                      <AgentLinkActions onNavigate={(q) => { onClose(); onNavigate(q) }} q={inst.label || inst.name || inst.id} />
                    </td>
                  </tr>
                ))}</tbody>
              </table>
              {!instances.length && !loading && <Empty>暂无实例数据</Empty>}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export function ServicesSection({ onNavigate }) {
  const [filters, setFilters] = useState({ layer: 'GENERAL', entity: 'service', keyword: '', duration: '15m' })
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [detail, setDetail] = useState(null)
  const [serviceDetailName, setServiceDetailName] = useState(null)
  const filtered = useMemo(() => rows.filter(row => !filters.keyword || rowText(row).includes(filters.keyword.toLowerCase())), [rows, filters.keyword])

  const load = useCallback(async () => {
    setLoading(true); setError(''); setBlocked('')
    try {
      const range = durationToRange(filters.duration)
      const data = await tracingApi.selectors.services({ layer: filters.layer, ...range })
      setRows(data || [])
    } catch (err) {
      setRows([])
      const msg = formatTracingError(err)
      setError(msg)
    } finally { setLoading(false) }
  }, [filters.layer, filters.duration])

  useEffect(() => { load() }, [load])

  return (
    <section className='fx-tracing-work'>
      {/* Layer concept follows upstream tracing source semantics. */}
      <div className='fx-tracing-condition-bar'>
        <Field label='时间范围'>
          <select value={filters.duration} onChange={e => setFilters(prev => Object.assign({}, prev, { duration: e.target.value }))}>
            {durationPresets.map(p => <option key={p.value} value={p.value}>{p.label}</option>)}
          </select>
        </Field>
        <Field label='层级 Layer'>
          <select value={filters.layer} onChange={e => setFilters(prev => Object.assign({}, prev, { layer: e.target.value }))}>
            {LAYER_GROUPS.map(group => (
              <optgroup key={group.group} label={group.group}>
                {group.layers.map(layer => <option key={layer} value={layer}>{layer}</option>)}
              </optgroup>
            ))}
            {layerOptions.filter(l => !LAYER_GROUPS.some(g => g.layers.includes(l))).map(l => <option key={l} value={l}>{l}</option>)}
          </select>
        </Field>
        <Field label='对象'>
          <select value={filters.entity} onChange={e => setFilters(prev => Object.assign({}, prev, { entity: e.target.value }))}>
            {entityOptions.map(item => <option key={item} value={item}>{item}</option>)}
          </select>
        </Field>
        <Field label='搜索' className='is-flex'>
          <input value={filters.keyword} onChange={e => setFilters(prev => Object.assign({}, prev, { keyword: e.target.value }))} placeholder='按服务名称过滤' />
        </Field>
        <div className='fx-tracing-condition-actions'>
          <button type='button' className='is-primary' onClick={load}>{loading ? '查询中...' : '查询'}</button>
          <button type='button' onClick={() => onNavigate({ section: 'topology', layer: filters.layer })}>查看拓扑</button>
        </div>
      </div>

      <AgentEvidenceNotice>服务目录按 serviceId 下钻实例和进程, 再通过 process.agentId 反查 FindX Agent 覆盖率; 当前 APM-Agent Adapter 未开放时保留阻断说明和主机 Agent 入口。</AgentEvidenceNotice>
      <AgentLinkActions onNavigate={onNavigate} q={filters.keyword} />
      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}

      <div className='fx-tracing-table'>
        <h3>服务指标 ({filtered.length})</h3>
        <table>
          <thead>
            <tr>
              <th>服务名</th>
              <th>语言 / 层级</th>
              <th>实例数</th>
              <th>平均响应时间</th>
              <th>成功率</th>
              <th>吞吐 CPM</th>
              <th>Apdex</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map(row => {
              const st = statusOf(row)
              const label = displayText(row.label || row.name || row.shortName)
              return (
                <tr key={row.id || row.value || row.label}>
                  <td style={{ fontWeight: 600, cursor: 'pointer', color: 'var(--fx-blue, #1769ff)' }} onClick={() => setDetail(row)}>{label}</td>
                  <td>{displayText(row.language || row.lang || row.layer || filters.layer)}</td>
                  <td>{displayText(row.instance_count || row.instances || row.numOfServiceInstance)}</td>
                  <td>{fmtLatency(row)}</td>
                  <td>{fmtSuccessRate(row)}</td>
                  <td>{fmtCpm(row)}</td>
                  <td>{fmtApdex(row)}</td>
                  <td><Status ok={st.ok}>{st.text}</Status></td>
                  <td className='fx-tracing-actions'>
                    <button type='button' onClick={() => setDetail(row)}>概览</button>
                    <button type='button' onClick={() => setServiceDetailName(label)}>详情</button>
                    <button type='button' onClick={() => onNavigate({ section: 'topology', serviceId: row.id || row.value })}>拓扑</button>
                    <button type='button' onClick={() => onNavigate({ section: 'traces', serviceId: row.id || row.value })}>Trace</button>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
        {!filtered.length && !loading && <Empty>{rows.length === 0 && !blocked ? CONNECTION_HINT : '没有匹配的服务'}</Empty>}
      </div>

      {detail && <ServiceDetail service={detail} onClose={() => setDetail(null)} onNavigate={onNavigate} />}
      {serviceDetailName && <ServiceDetailSection serviceName={serviceDetailName} onClose={() => setServiceDetailName(null)} onNavigate={onNavigate} />}
    </section>
  )
}
