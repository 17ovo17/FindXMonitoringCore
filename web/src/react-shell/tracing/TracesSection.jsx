import React, { useCallback, useEffect, useState } from 'react'
import { formatTracingError, tracingApi } from '../api/tracing.js'
import { defaultDuration, displayText, durationMs, fmtTime, orderOptions, toTraceCondition, traceId, traceName, traceStates } from './tracingModel.js'
import { AgentEvidenceNotice, AgentLinkActions, Blocked, Empty, ErrorBox, Field, JsonPreview, Status } from './TracingShared.jsx'

const CONNECTION_HINT = '暂无 Trace 数据。请确认 SkyWalking OAP 已启动并有应用接入。\n\n接入步骤：\n1. 确认 SkyWalking OAP 地址已在平台设置中配置\n2. 在目标应用中集成 SkyWalking Agent 并配置 collector 地址\n3. 对应用发起请求后，Trace 数据将在数秒内出现'

export function TracesSection({ query, onNavigate }) {
  const duration = defaultDuration()
  const [draft, setDraft] = useState({ start: duration.start, end: duration.end, serviceId: query.serviceId || '', instanceId: query.instanceId || '', endpointId: query.endpointId || '', traceId: query.traceId || '', traceState: 'ALL', minDuration: '', maxDuration: '', queryOrder: 'BY_START_TIME', tags: '', pageNum: '1' })
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [services, setServices] = useState([])

  const patch = (key, value) => setDraft(prev => ({ ...prev, [key]: value }))

  // Load service list for the dropdown
  useEffect(() => {
    tracingApi.selectors.services({}).then(list => setServices(list || [])).catch(() => {})
  }, [])

  const queryTraces = useCallback(async () => {
    setLoading(true); setError(''); setBlocked('')
    try {
      const condition = toTraceCondition(draft)
      const result = await tracingApi.traces.query(condition)
      const traceList = Array.isArray(result) ? result : (result?.traces || result?.data || [])
      setRows(traceList)
      setTotal(result?.total || traceList.length || 0)
    } catch (err) {
      setRows([])
      const msg = formatTracingError(err)
      if (msg.startsWith('BLOCKED_BY_CONTRACT')) { setBlocked(msg) } else { setError(msg) }
    } finally { setLoading(false) }
  }, [draft])

  useEffect(() => { if (query.traceId) queryTraces() }, [query.traceId])

  const pageCount = Math.max(1, Math.ceil(total / 20))
  const currentPage = Number(draft.pageNum) || 1

  return (
    <section className='fx-tracing-work'>
      <div className='fx-tracing-filter is-wide'>
        <Field label='开始时间'><input type='datetime-local' value={draft.start} onChange={e => patch('start', e.target.value)} /></Field>
        <Field label='结束时间'><input type='datetime-local' value={draft.end} onChange={e => patch('end', e.target.value)} /></Field>
        <Field label='Trace ID'><input value={draft.traceId} onChange={e => patch('traceId', e.target.value)} placeholder='输入 Trace ID 精确查询' /></Field>
        <Field label='服务'>
          <select value={draft.serviceId} onChange={e => patch('serviceId', e.target.value)}>
            <option value=''>全部服务</option>
            {services.map(s => <option key={s.id || s.value} value={s.id || s.value}>{displayText(s.label || s.name)}</option>)}
          </select>
        </Field>
        <Field label='端点'><input value={draft.endpointId} onChange={e => patch('endpointId', e.target.value)} placeholder='端点 ID 或名称' /></Field>
        <Field label='状态'>
          <select value={draft.traceState} onChange={e => patch('traceState', e.target.value)}>
            {traceStates.map(item => <option key={item} value={item}>{item === 'ALL' ? '全部' : item === 'SUCCESS' ? '成功' : '错误'}</option>)}
          </select>
        </Field>
        <Field label='排序'>
          <select value={draft.queryOrder} onChange={e => patch('queryOrder', e.target.value)}>
            {orderOptions.map(item => <option key={item} value={item}>{item === 'BY_START_TIME' ? '按时间' : '按耗时'}</option>)}
          </select>
        </Field>
        <Field label='最小耗时(ms)'><input type='number' value={draft.minDuration} onChange={e => patch('minDuration', e.target.value)} placeholder='0' /></Field>
        <Field label='最大耗时(ms)'><input type='number' value={draft.maxDuration} onChange={e => patch('maxDuration', e.target.value)} placeholder='不限' /></Field>
        <Field label='标签'><textarea rows='2' value={draft.tags} onChange={e => patch('tags', e.target.value)} placeholder='key=value，每行一个' /></Field>
      </div>
      <div className='fx-tracing-toolbar'>
        <button type='button' onClick={queryTraces}>{loading ? '检索中...' : '检索 Trace'}</button>
        <button type='button' onClick={() => { const d = defaultDuration(); setDraft(prev => ({ ...prev, start: d.start, end: d.end, serviceId: '', instanceId: '', endpointId: '', traceId: '', traceState: 'ALL', minDuration: '', maxDuration: '', tags: '', pageNum: '1' })); setRows([]); setError(''); setBlocked('') }}>重置</button>
        <AgentLinkActions onNavigate={onNavigate} q={draft.serviceId || draft.instanceId || draft.endpointId || draft.traceId} />
      </div>
      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}
      <AgentEvidenceNotice>Trace 检索条件会保留服务、实例、端点或 Trace ID 上下文，用于跳转主机 Agent 过滤；真实探针状态必须等待 APM-Agent 映射契约。</AgentEvidenceNotice>
      <div className='fx-tracing-table'>
        {rows.length > 0 && <p style={{ margin: '0 0 8px', color: 'var(--fx-muted, #66758d)', fontSize: 12 }}>共 {total} 条结果，当前第 {currentPage}/{pageCount} 页</p>}
        <table><thead><tr><th>Trace ID</th><th>服务</th><th>端点</th><th>耗时</th><th>状态</th><th>时间</th><th>操作</th></tr></thead><tbody>
          {rows.map(row => (
            <tr key={traceId(row)}>
              <td style={{ fontFamily: 'monospace', fontSize: 12, cursor: 'pointer', color: 'var(--fx-blue, #1769ff)' }} onClick={() => onNavigate({ section: 'trace-detail', traceId: traceId(row) })}>{displayText(traceId(row))}</td>
              <td>{displayText(row.serviceCode || row.service || row.serviceName)}</td>
              <td>{traceName(row)}</td>
              <td>{durationMs(row)} ms</td>
              <td><Status ok={!row.isError}>{row.isError ? '错误' : '正常'}</Status></td>
              <td>{fmtTime(row.start || row.startTime)}</td>
              <td className='fx-tracing-actions'>
                <button type='button' onClick={() => onNavigate({ section: 'trace-detail', traceId: traceId(row) })}>详情</button>
                <AgentLinkActions onNavigate={onNavigate} q={row.serviceCode || row.service || traceName(row) || traceId(row)} />
              </td>
            </tr>
          ))}
        </tbody></table>
        {!rows.length && !loading && <Empty>{!blocked ? CONNECTION_HINT : '暂无匹配的 Trace 数据'}</Empty>}
        {rows.length > 0 && pageCount > 1 && (
          <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginTop: 12, justifyContent: 'center' }}>
            <button type='button' disabled={currentPage <= 1} onClick={() => { patch('pageNum', String(currentPage - 1)); setTimeout(queryTraces, 0) }} style={{ minHeight: 32, border: '1px solid #d8e1ee', borderRadius: 8, background: '#fff', padding: '0 12px', cursor: 'pointer' }}>上一页</button>
            <span style={{ fontSize: 13, color: 'var(--fx-muted, #66758d)' }}>第 {currentPage} / {pageCount} 页</span>
            <button type='button' disabled={currentPage >= pageCount} onClick={() => { patch('pageNum', String(currentPage + 1)); setTimeout(queryTraces, 0) }} style={{ minHeight: 32, border: '1px solid #d8e1ee', borderRadius: 8, background: '#fff', padding: '0 12px', cursor: 'pointer' }}>下一页</button>
          </div>
        )}
      </div>
      <JsonPreview value={{ condition: toTraceCondition(draft) }} />
    </section>
  )
}
