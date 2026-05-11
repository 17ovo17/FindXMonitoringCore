import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { formatTracingError, tracingApi } from '../api/tracing.js'
import { displayText, entityOptions, layerOptions, rowText } from './tracingModel.js'
import { AgentEvidenceNotice, AgentLinkActions, Blocked, Empty, ErrorBox, Field, Status } from './TracingShared.jsx'

const CONNECTION_HINT = '暂无服务数据。请确认 SkyWalking OAP 已启动并配置了正确的接入地址。\n\n接入步骤：\n1. 在 FindX 平台设置中配置 SkyWalking OAP 地址（默认 http://localhost:12800）\n2. 在目标应用中集成 SkyWalking Agent\n3. 启动应用后等待 1-2 分钟，服务将自动注册'

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
          tracingApi.selectors.endpoints({ serviceId }).catch(() => []),
          tracingApi.selectors.instances({ serviceId }).catch(() => []),
        ])
        if (!cancelled) { setEndpoints(ep); setInstances(inst) }
      } catch (err) { if (!cancelled) setError(formatTracingError(err)) }
      finally { if (!cancelled) setLoading(false) }
    }
    load()
    return () => { cancelled = true }
  }, [serviceId])

  return (
    <div className='fx-tracing-modal'>
      <div className='fx-tracing-modal__body'>
        <header>
          <h2>{displayText(service.label || service.name || service.shortName)} - 详情</h2>
          <button type='button' onClick={onClose}>关闭</button>
        </header>
        {error && <ErrorBox>{error}</ErrorBox>}
        {loading && <p style={{ color: 'var(--fx-muted, #66758d)', fontSize: 13 }}>加载中...</p>}
        <div className='fx-tracing-split'>
          <div>
            <div className='fx-tracing-table'>
              <h3>端点列表（{endpoints.length}）</h3>
              <table><thead><tr><th>端点名称</th><th>操作</th></tr></thead>
                <tbody>{endpoints.map(ep => (
                  <tr key={ep.id || ep.value || ep.label}>
                    <td>{displayText(ep.label || ep.name)}</td>
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
              <h3>实例列表（{instances.length}）</h3>
              <table><thead><tr><th>实例名称</th><th>语言</th><th>操作</th></tr></thead>
                <tbody>{instances.map(inst => (
                  <tr key={inst.id || inst.value || inst.label}>
                    <td>{displayText(inst.label || inst.name)}</td>
                    <td>{displayText(inst.language || inst.lang)}</td>
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
  const [filters, setFilters] = useState({ layer: 'GENERAL', entity: 'service', keyword: '' })
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [detail, setDetail] = useState(null)
  const filtered = useMemo(() => rows.filter(row => !filters.keyword || rowText(row).includes(filters.keyword.toLowerCase())), [rows, filters.keyword])

  const load = useCallback(async () => {
    setLoading(true); setError(''); setBlocked('')
    try {
      const data = await tracingApi.selectors.services({ layer: filters.layer })
      setRows(data || [])
    } catch (err) {
      setRows([])
      const msg = formatTracingError(err)
      if (msg.startsWith('BLOCKED_BY_CONTRACT')) { setBlocked(msg) } else { setError(msg) }
    } finally { setLoading(false) }
  }, [filters.layer])

  useEffect(() => { load() }, [load])

  const getLatency = row => {
    const v = row.avgResponseTime || row.avg_latency || row.latency
    return v !== undefined && v !== null ? `${Number(v).toFixed(0)} ms` : '-'
  }
  const getErrorRate = row => {
    const v = row.errorRate || row.error_rate
    return v !== undefined && v !== null ? `${(Number(v) * 100).toFixed(2)}%` : '-'
  }
  const getThroughput = row => {
    const v = row.cpm || row.throughput || row.calls
    return v !== undefined && v !== null ? `${Number(v).toFixed(0)} cpm` : '-'
  }
  const getStatus = row => {
    if (row.isError || row.status === 'unhealthy') return { ok: false, text: '异常' }
    if (row.status === 'healthy' || row.normal === true) return { ok: true, text: '正常' }
    return { ok: true, text: displayText(row.status || row.health || '正常') }
  }

  return (
    <section className='fx-tracing-work'>
      <div className='fx-tracing-filter'>
        <Field label='层级'><select value={filters.layer} onChange={e => setFilters(prev => ({ ...prev, layer: e.target.value }))}>{layerOptions.map(item => <option key={item}>{item}</option>)}</select></Field>
        <Field label='对象'><select value={filters.entity} onChange={e => setFilters(prev => ({ ...prev, entity: e.target.value }))}>{entityOptions.map(item => <option key={item} value={item}>{item}</option>)}</select></Field>
        <Field label='搜索'><input value={filters.keyword} onChange={e => setFilters(prev => ({ ...prev, keyword: e.target.value }))} placeholder='服务名称' /></Field>
        <button type='button' onClick={load}>{loading ? '查询中...' : '查询'}</button>
      </div>
      <AgentEvidenceNotice>服务目录按 serviceId 下钻实例和进程，再通过 process.agentId 反查 FindX Agent 覆盖率；当前 APM-Agent Adapter 未开放时保留阻断说明和主机 Agent 入口。</AgentEvidenceNotice>
      <AgentLinkActions onNavigate={onNavigate} q={filters.keyword} />
      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}
      <div className='fx-tracing-table'>
        <table><thead><tr><th>服务名</th><th>语言/框架</th><th>实例数</th><th>平均延迟</th><th>错误率</th><th>吞吐量</th><th>状态</th><th>操作</th></tr></thead>
          <tbody>{filtered.map(row => {
            const st = getStatus(row)
            return (
              <tr key={row.id || row.value || row.label}>
                <td style={{ fontWeight: 600, cursor: 'pointer', color: 'var(--fx-blue, #1769ff)' }} onClick={() => setDetail(row)}>{displayText(row.label || row.name || row.shortName)}</td>
                <td>{displayText(row.language || row.lang || row.layer || filters.layer)}</td>
                <td>{displayText(row.instance_count || row.instances || row.numOfServiceInstance)}</td>
                <td>{getLatency(row)}</td>
                <td>{getErrorRate(row)}</td>
                <td>{getThroughput(row)}</td>
                <td><Status ok={st.ok}>{st.text}</Status></td>
                <td className='fx-tracing-actions'>
                  <button type='button' onClick={() => setDetail(row)}>详情</button>
                  <button type='button' onClick={() => onNavigate({ section: 'topology', serviceId: row.id || row.value })}>拓扑</button>
                  <button type='button' onClick={() => onNavigate({ section: 'traces', serviceId: row.id || row.value })}>Trace</button>
                </td>
              </tr>
            )
          })}</tbody>
        </table>
        {!filtered.length && !loading && <Empty>{rows.length === 0 && !blocked ? CONNECTION_HINT : '没有匹配的服务'}</Empty>}
      </div>
      {detail && <ServiceDetail service={detail} onClose={() => setDetail(null)} onNavigate={onNavigate} />}
    </section>
  )
}
