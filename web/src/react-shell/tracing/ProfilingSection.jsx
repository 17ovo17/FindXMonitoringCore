import React, { useCallback, useEffect, useState } from 'react'
import { formatTracingError, tracingApi } from '../api/tracing.js'
import { displayText, fmtTime } from './tracingModel.js'
import { AgentEvidenceNotice, AgentLinkActions, Blocked, Empty, ErrorBox, Field, Status } from './TracingShared.jsx'

const defaultForm = {
  serviceId: '',
  endpointName: '',
  duration: '5',
  minDurationThreshold: '0',
  dumpPeriod: '10',
  maxSamplingCount: '5',
}

function CreateTaskForm({ onCreated }) {
  const [form, setForm] = useState(defaultForm)
  const [services, setServices] = useState([])
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')

  const patch = (key, value) => setForm(prev => ({ ...prev, [key]: value }))

  useEffect(() => {
    tracingApi.selectors.services({}).then(list => setServices(list || [])).catch(() => {})
  }, [])

  const submit = async () => {
    if (!form.serviceId) { setError('请选择服务'); return }
    setSubmitting(true); setError(''); setBlocked('')
    try {
      await tracingApi.profiling.create({
        serviceId: form.serviceId,
        endpointName: form.endpointName,
        duration: Number(form.duration) || 5,
        minDurationThreshold: Number(form.minDurationThreshold) || 0,
        dumpPeriod: Number(form.dumpPeriod) || 10,
        maxSamplingCount: Number(form.maxSamplingCount) || 5,
      })
      setForm(defaultForm)
      onCreated && onCreated()
    } catch (err) {
      const msg = formatTracingError(err)
      if (msg.startsWith('BLOCKED_BY_CONTRACT') || [404, 405, 501].includes(err?.status)) {
        setBlocked('BLOCKED_BY_CONTRACT: 需要后端实现 /apm/profiling/tasks POST API')
      } else { setError(msg) }
    } finally { setSubmitting(false) }
  }

  return (
    <div style={{ border: '1px solid var(--fx-border)', borderRadius: 8, padding: 12, marginBottom: 12 }}>
      <h3 style={{ margin: '0 0 10px', fontSize: 14 }}>创建 Profiling 任务</h3>
      <div className='fx-tracing-form'>
        <Field label='服务'>
          <select value={form.serviceId} onChange={e => patch('serviceId', e.target.value)}>
            <option value=''>选择服务</option>
            {services.map(s => <option key={s.id || s.value} value={s.id || s.value}>{displayText(s.label || s.name)}</option>)}
          </select>
        </Field>
        <Field label='端点名称'><input value={form.endpointName} onChange={e => patch('endpointName', e.target.value)} placeholder='可选，留空表示全部端点' /></Field>
      </div>
      <div className='fx-tracing-form'>
        <Field label='持续时间(分钟)'><input type='number' value={form.duration} onChange={e => patch('duration', e.target.value)} min='1' max='60' /></Field>
        <Field label='最小耗时阈值(ms)'><input type='number' value={form.minDurationThreshold} onChange={e => patch('minDurationThreshold', e.target.value)} min='0' /></Field>
        <Field label='采样间隔(ms)'><input type='number' value={form.dumpPeriod} onChange={e => patch('dumpPeriod', e.target.value)} min='1' /></Field>
        <Field label='最大采样数'><input type='number' value={form.maxSamplingCount} onChange={e => patch('maxSamplingCount', e.target.value)} min='1' max='100' /></Field>
      </div>
      <div className='fx-tracing-toolbar'>
        <button type='button' onClick={submit} disabled={submitting}>{submitting ? '创建中...' : '创建任务'}</button>
      </div>
      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}
    </div>
  )
}
export function ProfilingSection({ query = {}, onNavigate }) {
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [showForm, setShowForm] = useState(false)

  const load = useCallback(async () => {
    setLoading(true); setError(''); setBlocked('')
    try {
      const data = await tracingApi.profiling.list({})
      setRows(data || [])
    } catch (err) {
      setRows([])
      const msg = formatTracingError(err)
      if (msg.startsWith('BLOCKED_BY_CONTRACT') || [404, 405, 501].includes(err?.status)) {
        setBlocked('BLOCKED_BY_CONTRACT: 需要后端实现 /apm/profiling/tasks GET API')
      } else { setError(msg) }
    } finally { setLoading(false) }
  }, [])

  useEffect(() => { load() }, [load])

  const cancelTask = async (id) => {
    try {
      await tracingApi.profiling.cancel(id)
      load()
    } catch (err) {
      const msg = formatTracingError(err)
      if (msg.startsWith('BLOCKED_BY_CONTRACT') || [404, 405, 501].includes(err?.status)) {
        setBlocked('BLOCKED_BY_CONTRACT: 需要后端实现 /apm/profiling/tasks/:id/cancel API')
      } else { setError(msg) }
    }
  }

  return (
    <section className='fx-tracing-work'>
      <div className='fx-tracing-toolbar' style={{ marginBottom: 12 }}>
        <button type='button' onClick={() => setShowForm(!showForm)}>{showForm ? '收起表单' : '创建任务'}</button>
        <button type='button' onClick={load}>{loading ? '加载中...' : '刷新列表'}</button>
      </div>

      {showForm && <CreateTaskForm onCreated={() => { setShowForm(false); load() }} />}

      <AgentEvidenceNotice>Profiling 任务需要实例进程和 Agent 能力支持状态；当前只保留主机 Agent 入口和阻断说明，不伪造任务可执行状态。</AgentEvidenceNotice>
      <AgentLinkActions onNavigate={onNavigate} q={query.q} packageName='profiling' />
      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}

      <div className='fx-tracing-table'>
        <h3>Profiling 任务列表 ({rows.length})</h3>
        <table>
          <thead><tr><th>任务 ID</th><th>服务</th><th>端点</th><th>持续时间</th><th>采样间隔</th><th>状态</th><th>创建时间</th><th>操作</th></tr></thead>
          <tbody>
            {rows.map(row => (
              <tr key={row.id || row.taskId}>
                <td style={{ fontFamily: 'monospace', fontSize: 11 }}>{displayText(row.id || row.taskId)}</td>
                <td>{displayText(row.serviceId || row.serviceName)}</td>
                <td>{displayText(row.endpointName || '-')}</td>
                <td>{displayText(row.duration ? row.duration + ' min' : '-')}</td>
                <td>{displayText(row.dumpPeriod ? row.dumpPeriod + ' ms' : '-')}</td>
                <td><Status ok={row.status === 'FINISHED' || row.status === 'SUCCESS'}>{displayText(row.status || 'UNKNOWN')}</Status></td>
                <td>{fmtTime(row.createTime || row.created_at)}</td>
                <td className='fx-tracing-actions'>
                  <button type='button' onClick={() => cancelTask(row.id || row.taskId)}>取消</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!rows.length && !loading && <Empty>{blocked || '暂无 Profiling 任务'}</Empty>}
      </div>
    </section>
  )
}
