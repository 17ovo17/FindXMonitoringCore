import React, { useCallback, useEffect, useState } from 'react'
import { formatTracingError, tracingApi } from '../api/tracing.js'
import { defaultDuration, displayText, durationMs, fmtTime, orderOptions, toTraceCondition, traceId, traceName, traceStates } from './tracingModel.js'
import { AgentEvidenceNotice, AgentLinkActions, Blocked, Empty, ErrorBox, Field, Status } from './TracingShared.jsx'
import { TraceWaterfall } from './TraceWaterfall.jsx'
import { TagsFilter } from './TagsFilter.jsx'

const CONNECTION_HINT = [
  '暂无 Trace 数据。请确认链路监控上游服务已启动并有应用接入。',
  '',
  '接入步骤:',
  '1. 确认链路查询服务地址已在平台设置中配置',
  '2. 在目标应用中集成 FindX Agent 并配置采集地址',
  '3. 对应用发起请求后, Trace 数据将在数秒内出现',
].join('\n')

export function TracesSection(props) {
  const query = props.query || {}
  const onNavigate = props.onNavigate
  const duration = defaultDuration()
  const [draft, setDraft] = useState({
    start: duration.start, end: duration.end,
    serviceId: query.serviceId || '', instanceId: query.instanceId || '', endpointId: query.endpointId || '',
    traceId: query.traceId || '', traceState: 'ALL',
    minDuration: '', maxDuration: '',
    queryOrder: 'BY_START_TIME', tags: '', pageNum: '1',
  })
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [services, setServices] = useState([])
  const [selectedTraceId, setSelectedTraceId] = useState(query.traceId || '')
  const [selectedRow, setSelectedRow] = useState(null)
  const [detail, setDetail] = useState({ spans: [], loading: false, error: '', blocked: '' })

  const patch = (key, value) => setDraft(prev => Object.assign({}, prev, { [key]: value }))

  useEffect(() => {
    tracingApi.selectors.services({}).then(list => setServices(list || [])).catch(() => {})
  }, [])

  const queryTraces = useCallback(async () => {
    setLoading(true); setError(''); setBlocked('')
    try {
      const condition = toTraceCondition(draft)
      const result = await tracingApi.traces.query(condition)
      const traceList = Array.isArray(result) ? result : ((result && (result.traces || result.data)) || [])
      setRows(traceList)
      setTotal((result && result.total) || traceList.length || 0)
    } catch (err) {
      setRows([])
      const msg = formatTracingError(err)
      if (msg.startsWith('PENDING')) setBlocked(msg); else setError(msg)
    } finally { setLoading(false) }
  }, [draft])

  useEffect(() => { if (query.traceId) queryTraces() }, [query.traceId])

  const loadSpans = useCallback(async (tid) => {
    if (!tid) return
    setDetail({ spans: [], loading: true, error: '', blocked: '' })
    try {
      const spans = await tracingApi.traces.spans(tid)
      setDetail({ spans: spans || [], loading: false, error: '', blocked: '' })
    } catch (err) {
      const msg = formatTracingError(err)
      if (msg.startsWith('PENDING')) {
        setDetail({ spans: [], loading: false, error: '', blocked: msg })
      } else {
        setDetail({ spans: [], loading: false, error: msg, blocked: '' })
      }
    }
  }, [])

  const selectTrace = (row) => {
    const tid = traceId(row)
    setSelectedTraceId(tid)
    setSelectedRow(row)
    loadSpans(tid)
  }

  useEffect(() => {
    if (query.traceId && !selectedTraceId) {
      setSelectedTraceId(query.traceId)
      loadSpans(query.traceId)
    }
  }, [query.traceId, selectedTraceId, loadSpans])

  const pageCount = Math.max(1, Math.ceil(total / 20))
  const currentPage = Number(draft.pageNum) || 1
  const hasSelection = Boolean(selectedTraceId)

  const reset = () => {
    const d = defaultDuration()
    setDraft(prev => Object.assign({}, prev, { start: d.start, end: d.end, serviceId: '', instanceId: '', endpointId: '', traceId: '', traceState: 'ALL', minDuration: '', maxDuration: '', tags: '', pageNum: '1' }))
    setRows([]); setError(''); setBlocked('')
    setSelectedTraceId(''); setSelectedRow(null); setDetail({ spans: [], loading: false, error: '', blocked: '' })
  }

  const goPage = (delta) => {
    const next = Math.max(1, Math.min(pageCount, currentPage + delta))
    setDraft(prev => Object.assign({}, prev, { pageNum: String(next) }))
    setTimeout(queryTraces, 0)
  }

  const layoutCls = hasSelection ? 'fx-tracing-trace-layout' : 'fx-tracing-trace-layout is-single'

  return (
    <section className='fx-tracing-work'>
      <div className='fx-tracing-condition-bar'>
        <Field label='开始时间'><input type='datetime-local' value={draft.start} onChange={e => patch('start', e.target.value)} /></Field>
        <Field label='结束时间'><input type='datetime-local' value={draft.end} onChange={e => patch('end', e.target.value)} /></Field>
        <Field label='服务'>
          <select value={draft.serviceId} onChange={e => patch('serviceId', e.target.value)}>
            <option value=''>全部服务</option>
            {services.map(s => <option key={s.id || s.value} value={s.id || s.value}>{displayText(s.label || s.name)}</option>)}
          </select>
        </Field>
        <Field label='端点'><input value={draft.endpointId} onChange={e => patch('endpointId', e.target.value)} placeholder='端点 ID 或名称' /></Field>
        <Field label='Trace ID'><input value={draft.traceId} onChange={e => patch('traceId', e.target.value)} placeholder='精确匹配' /></Field>
        <Field label='状态'>
          <select value={draft.traceState} onChange={e => patch('traceState', e.target.value)}>
            {traceStates.map(item => <option key={item} value={item}>{item === 'ALL' ? '全部' : item === 'SUCCESS' ? '成功' : '错误'}</option>)}
          </select>
        </Field>
        <Field label='最小耗时(ms)'><input type='number' value={draft.minDuration} onChange={e => patch('minDuration', e.target.value)} placeholder='0' /></Field>
        <Field label='最大耗时(ms)'><input type='number' value={draft.maxDuration} onChange={e => patch('maxDuration', e.target.value)} placeholder='不限' /></Field>
        <Field label='排序'>
          <select value={draft.queryOrder} onChange={e => patch('queryOrder', e.target.value)}>
            {orderOptions.map(item => <option key={item} value={item}>{item === 'BY_START_TIME' ? '按时间' : '按耗时'}</option>)}
          </select>
        </Field>
        <div className='fx-tracing-condition-actions'>
          <button type='button' className='is-primary' onClick={queryTraces}>{loading ? '检索中...' : '检索'}</button>
          <button type='button' onClick={reset}>重置</button>
        </div>
      </div>

      <TagsFilter value={draft.tags} onChange={v => patch('tags', v)} />

      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}
      <AgentEvidenceNotice>Trace 检索条件会保留服务、实例、端点或 Trace ID 上下文, 用于跳转主机 Agent 过滤; 真实探针状态必须等待 APM-Agent 映射契约。</AgentEvidenceNotice>
      <AgentLinkActions onNavigate={onNavigate} q={draft.serviceId || draft.instanceId || draft.endpointId || draft.traceId} />

      <div className={layoutCls}>
        <div className='fx-tracing-trace-list'>
          <div className='fx-tracing-trace-list__head'>
            <div><strong>Trace Segments</strong>{total ? '共 ' + total + ' 条' : ''}</div>
            {total > 0 && <span>第 {currentPage}/{pageCount} 页</span>}
          </div>
          <div className='fx-tracing-trace-list__body'>
            {rows.map(row => {
              const tid = traceId(row)
              const isActive = tid === selectedTraceId
              const itemCls = 'fx-tracing-trace-list__item' + (isActive ? ' is-active' : '') + (row.isError ? ' is-error' : '')
              return (
                <button type='button' key={tid} className={itemCls} onClick={() => selectTrace(row)}>
                  <span className='fx-endpoint'>{traceName(row)}</span>
                  <span className='fx-meta'>
                    <span className='fx-tag'>{durationMs(row)} ms</span>
                    {displayText(row.serviceCode || row.service || row.serviceName)} · {fmtTime(row.start || row.startTime)}
                  </span>
                </button>
              )
            })}
            {!rows.length && !loading && <Empty>{!blocked ? CONNECTION_HINT : '暂无匹配的 Trace 数据'}</Empty>}
          </div>
          {rows.length > 0 && pageCount > 1 && (
            <div className='fx-tracing-trace-list__pager'>
              <button type='button' disabled={currentPage <= 1} onClick={() => goPage(-1)}>上一页</button>
              <span>{currentPage} / {pageCount}</span>
              <button type='button' disabled={currentPage >= pageCount} onClick={() => goPage(1)}>下一页</button>
            </div>
          )}
        </div>

        {hasSelection && (
          <div className='fx-tracing-trace-detail'>
            <div className='fx-tracing-trace-detail__head'>
              <div style={{ minWidth: 0, flex: 1 }}>
                <h3>{(selectedRow && traceName(selectedRow)) || selectedTraceId}</h3>
                <div className='fx-meta'>
                  <b>Trace ID</b><span style={{ fontFamily: 'monospace' }}>{displayText(selectedTraceId)}</span>
                  <br />
                  <b>服务</b>{displayText(selectedRow && (selectedRow.serviceCode || selectedRow.service || selectedRow.serviceName))}
                  <b>耗时</b>{durationMs(selectedRow || {})} ms
                  <b>Span 数</b>{detail.spans.length}
                  <b>状态</b><Status ok={!(selectedRow && selectedRow.isError)}>{selectedRow && selectedRow.isError ? '错误' : '正常'}</Status>
                  <b>开始</b>{fmtTime(selectedRow && (selectedRow.start || selectedRow.startTime))}
                </div>
              </div>
              <div style={{ display: 'flex', gap: 6 }}>
                <button type='button' onClick={() => loadSpans(selectedTraceId)}>{detail.loading ? '加载中...' : '刷新'}</button>
                <button type='button' onClick={() => { setSelectedTraceId(''); setSelectedRow(null); setDetail({ spans: [], loading: false, error: '', blocked: '' }) }}>关闭</button>
              </div>
            </div>
            {detail.error && <ErrorBox>{detail.error}</ErrorBox>}
            {detail.blocked && <Blocked>{detail.blocked}</Blocked>}
            {!detail.blocked && !detail.error && <TraceWaterfall spans={detail.spans} />}
            <AgentLinkActions onNavigate={onNavigate} q={(selectedRow && (selectedRow.serviceCode || selectedRow.service)) || selectedTraceId} />
          </div>
        )}
      </div>
    </section>
  )
}
