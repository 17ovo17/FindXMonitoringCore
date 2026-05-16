import React, { useCallback, useEffect, useState } from 'react'
import { formatTracingError, tracingApi, TRACING_BLOCKERS } from '../api/tracing.js'
import { displayText, durationMs, fmtTime } from './tracingModel.js'
import { AgentEvidenceNotice, AgentLinkActions, Blocked, ErrorBox, Field, JsonPreview, Status } from './TracingShared.jsx'
import { TraceWaterfall } from './TraceWaterfall.jsx'

export function TraceDetailSection({ traceId, onNavigate }) {
  const [detail, setDetail] = useState(null)
  const [spans, setSpans] = useState([])
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)

  const load = useCallback(async () => {
    if (!traceId) { setBlocked('PENDING: Trace 详情需要 traceId 和详情查询契约。'); return }
    setLoading(true); setError(''); setBlocked('')
    try {
      const d = await tracingApi.traces.detail(traceId)
      const s = await tracingApi.traces.spans(traceId)
      const nextSpans = s || []
      setDetail(d || null)
      setSpans(nextSpans)
      if (!nextSpans.length) setBlocked(TRACING_BLOCKERS.traces)
    } catch (err) {
      setDetail(null)
      setSpans([])
      const msg = formatTracingError(err)
      if (msg.startsWith('PENDING')) setBlocked(msg); else setError(msg)
    } finally { setLoading(false) }
  }, [traceId])

  useEffect(() => { load() }, [load])

  return (
    <section className='fx-tracing-work'>
      <div className='fx-tracing-condition-bar'>
        <Field label='Trace ID' className='is-flex'>
          <input value={traceId || ''} readOnly style={{ fontFamily: 'monospace' }} />
        </Field>
        <div className='fx-tracing-condition-actions'>
          <button type='button' className='is-primary' onClick={load}>{loading ? '加载中...' : '刷新'}</button>
          <button type='button' onClick={() => setBlocked('PENDING: Span 日志和日志中心上下文关联缺少跨域证据契约。')}>关联日志</button>
          <button type='button' onClick={() => onNavigate({ section: 'traces' })}>返回列表</button>
        </div>
      </div>

      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}
      <AgentEvidenceNotice>Trace 详情需要从 Span 的服务、实例和端点回查进程 agentId, 再关联主机 FindX Agent、心跳、数据到达和配置版本; 当前契约缺失时只提供可追踪入口, 不伪造探针状态。</AgentEvidenceNotice>
      <AgentLinkActions onNavigate={onNavigate} q={traceId} />

      <div className='fx-tracing-trace-detail'>
        <div className='fx-tracing-trace-detail__head'>
          <div style={{ minWidth: 0, flex: 1 }}>
            <h3>{displayText((detail && (detail.endpointName || detail.endpoint)) || traceId)}</h3>
            <div className='fx-meta'>
              <b>Trace ID</b><span style={{ fontFamily: 'monospace' }}>{displayText(traceId)}</span>
              <br />
              <b>服务</b>{displayText(detail && (detail.serviceCode || detail.service))}
              <b>耗时</b>{durationMs(detail)} ms
              <b>Span 数</b>{spans.length}
              <b>状态</b><Status ok={!(detail && detail.isError)}>{detail && detail.isError ? '错误' : '正常'}</Status>
              <b>开始</b>{fmtTime(detail && (detail.start || detail.startTime))}
            </div>
          </div>
        </div>
        {!blocked && !error && <TraceWaterfall spans={spans} />}
      </div>

      <JsonPreview value={{ traceId, span_count: spans.length }} />
    </section>
  )
}
