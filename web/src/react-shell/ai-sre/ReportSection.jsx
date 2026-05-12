import React, { useState } from 'react'
import { AISRE_BLOCKERS, aiSreApi, formatAiSreError } from '../api/aiSre.js'
import { Blocked, Empty, ErrorBox, Field, JsonPreview, StatusPill } from './AiSreShared.jsx'

export function ReportSection({ query, onNavigate, addEvidence }) {
  const [name, setName] = useState('AI SRE 巡检')
  const [inspection, setInspection] = useState(query.inspectionId ? { inspectionId: query.inspectionId } : null)
  const [progress, setProgress] = useState(null)
  const [report, setReport] = useState(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const refreshInspection = async id => {
    const [nextProgress, nextReport] = await Promise.all([
      aiSreApi.inspections.progress(id).catch(err => ({ error: formatAiSreError(err) })),
      aiSreApi.inspections.report(id).catch(err => ({ error: formatAiSreError(err) })),
    ])
    setProgress(nextProgress); setReport(nextReport)
    addEvidence({ category: 'inspection', title: '巡检报告已读取', detail: id })
  }

  const createInspection = async () => {
    setLoading(true); setError('')
    try {
      const created = await aiSreApi.inspections.create({ name, scope: {}, context: {} })
      setInspection(created)
      onNavigate({ section: 'report', inspectionId: created.inspectionId || created.id })
      await refreshInspection(created.inspectionId || created.id)
    } catch (err) {
      setError(formatAiSreError(err))
    } finally {
      setLoading(false)
    }
  }

  const refresh = () => {
    const id = inspection?.inspectionId || inspection?.id
    if (!id) return
    setLoading(true); setError('')
    refreshInspection(id).catch(err => setError(formatAiSreError(err))).finally(() => setLoading(false))
  }

  return (
    <section className='fx-aisre-grid'>
      <div className='fx-aisre-panel'>
        <h2>巡检任务</h2>
        <Field label='名称'><input value={name} onChange={event => setName(event.target.value)} /></Field>
        <div className='fx-aisre-toolbar'>
          <button type='button' onClick={createInspection} disabled={loading || !name.trim()}>{loading ? '请求中...' : '创建巡检'}</button>
          <button type='button' onClick={refresh} disabled={!inspection}>刷新报告</button>
        </div>
        <ErrorBox>{error}</ErrorBox>
        <Blocked>{AISRE_BLOCKERS.reportExport}</Blocked>
        {inspection ? <JsonPreview value={inspection} /> : <Empty>数据缺失：尚未创建或选择巡检任务。</Empty>}
      </div>
      <div className='fx-aisre-panel'>
        <h2>报告详情</h2>
        <StatusPill tone={progress?.status ? 'ok' : 'blocked'}>{progress?.status || '数据缺失'}</StatusPill>
        {progress && <JsonPreview value={progress} />}
        {report && <JsonPreview value={report} />}
        {!progress && !report && <Empty>数据缺失：暂无报告详情。</Empty>}
      </div>
    </section>
  )
}
