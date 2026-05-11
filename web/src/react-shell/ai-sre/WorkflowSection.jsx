import React, { useEffect, useState } from 'react'
import { AISRE_BLOCKERS, aiSreApi, formatAiSreError } from '../api/aiSre.js'
import { Blocked, Empty, ErrorBox, Field, StatusPill } from './AiSreShared.jsx'

const triggerTypes = [
  { value: 'manual', label: '手动触发' },
  { value: 'schedule', label: '定时调度 (Cron)' },
  { value: 'alert', label: '告警触发' },
  { value: 'event', label: '事件触发' },
]

const stepTypes = [
  { value: 'notify', label: '通知' },
  { value: 'execute', label: '执行' },
  { value: 'diagnose', label: '诊断' },
  { value: 'remediate', label: '修复' },
]

const emptyStep = () => ({ type: 'notify', config: '' })

function WorkflowForm({ onSave, onCancel }) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [trigger, setTrigger] = useState('manual')
  const [cron, setCron] = useState('')
  const [steps, setSteps] = useState([emptyStep()])
  const [targets, setTargets] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const addStep = () => setSteps(prev => [...prev, emptyStep()])
  const removeStep = idx => setSteps(prev => prev.filter((_, i) => i !== idx))
  const updateStep = (idx, field, value) => setSteps(prev => prev.map((s, i) => i === idx ? { ...s, [field]: value } : s))
  const moveStep = (idx, dir) => {
    const next = [...steps]
    const target = idx + dir
    if (target < 0 || target >= next.length) return
    ;[next[idx], next[target]] = [next[target], next[idx]]
    setSteps(next)
  }

  const handleSubmit = async e => {
    e.preventDefault()
    if (!name.trim()) { setError('请输入工作流名称'); return }
    setSaving(true); setError('')
    try {
      await onSave({ name, description, trigger, cron: trigger === 'schedule' ? cron : '', steps, notifyTargets: targets.split(',').map(s => s.trim()).filter(Boolean) })
    } catch (err) {
      setError(formatAiSreError(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <form className='fx-aisre-wf-form' onSubmit={handleSubmit}>
      <h3>创建工作流</h3>
      <ErrorBox>{error}</ErrorBox>
      <Field label='名称'><input value={name} onChange={e => setName(e.target.value)} placeholder='工作流名称' /></Field>
      <Field label='描述'><textarea value={description} onChange={e => setDescription(e.target.value)} rows={2} placeholder='工作流描述' /></Field>
      <Field label='触发方式'>
        <select value={trigger} onChange={e => setTrigger(e.target.value)}>
          {triggerTypes.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
        </select>
      </Field>
      {trigger === 'schedule' && <Field label='Cron 表达式'><input value={cron} onChange={e => setCron(e.target.value)} placeholder='0 */5 * * *' /></Field>}
      <div className='fx-aisre-wf-steps'>
        <div className='fx-aisre-wf-steps-head'><strong>步骤编辑器</strong><button type='button' onClick={addStep}>+ 添加步骤</button></div>
        {steps.map((step, idx) => (
          <div key={idx} className='fx-aisre-wf-step-card'>
            <span className='fx-aisre-wf-step-num'>{idx + 1}</span>
            <select value={step.type} onChange={e => updateStep(idx, 'type', e.target.value)}>
              {stepTypes.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
            </select>
            <input value={step.config} onChange={e => updateStep(idx, 'config', e.target.value)} placeholder='配置参数 (JSON 或文本)' />
            <button type='button' onClick={() => moveStep(idx, -1)} disabled={idx === 0}>↑</button>
            <button type='button' onClick={() => moveStep(idx, 1)} disabled={idx === steps.length - 1}>↓</button>
            <button type='button' onClick={() => removeStep(idx)} disabled={steps.length <= 1}>×</button>
          </div>
        ))}
      </div>
      <Field label='通知目标 (逗号分隔)'><input value={targets} onChange={e => setTargets(e.target.value)} placeholder='email@example.com, webhook-url' /></Field>
      <div className='fx-aisre-toolbar'>
        <button type='submit' disabled={saving}>{saving ? '保存中...' : '保存'}</button>
        <button type='button' onClick={onCancel}>取消</button>
      </div>
    </form>
  )
}

function RunHistory({ runs }) {
  const [expanded, setExpanded] = useState(null)
  if (!runs.length) return <Empty>暂无运行记录</Empty>
  return (
    <div className='fx-aisre-wf-runs'>
      {runs.map((run, idx) => (
        <div key={run.id || idx} className='fx-aisre-wf-run-item'>
          <div className='fx-aisre-wf-run-head' onClick={() => setExpanded(expanded === idx ? null : idx)}>
            <StatusPill tone={run.status === 'success' ? 'ok' : run.status === 'running' ? 'warn' : 'blocked'}>
              {run.status === 'success' ? '成功' : run.status === 'running' ? '运行中' : '失败'}
            </StatusPill>
            <span>{run.started_at ? new Date(run.started_at).toLocaleString('zh-CN', { hour12: false }) : '-'}</span>
            <span>{run.duration ? `${run.duration}s` : ''}</span>
            <span className='fx-aisre-wf-run-toggle'>{expanded === idx ? '▼' : '▶'}</span>
          </div>
          {expanded === idx && (
            <pre className='fx-aisre-text'>{run.log || run.output || '无日志输出'}</pre>
          )}
        </div>
      ))}
    </div>
  )
}

export function WorkflowSection({ query, onNavigate, addEvidence }) {
  const [rows, setRows] = useState([])
  const [runs, setRuns] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [showForm, setShowForm] = useState(false)

  const loadList = async () => {
    setLoading(true); setError('')
    try {
      const items = await aiSreApi.workflows.list()
      setRows(items)
      addEvidence({ category: 'workflow', title: '工作流列表已读取', detail: `${items.length} 个工作流` })
    } catch (err) {
      setError(formatAiSreError(err))
    } finally {
      setLoading(false)
    }
  }

  const loadRuns = async id => {
    if (!id) return
    try {
      const items = await aiSreApi.workflows.runs(id)
      setRuns(items)
    } catch { setRuns([]) }
  }

  const handleToggle = async (item) => {
    try {
      await aiSreApi.workflows.toggle(item.id, !item.enabled)
      setRows(prev => prev.map(r => r.id === item.id ? { ...r, enabled: !item.enabled } : r))
      addEvidence({ category: 'workflow', title: `工作流${item.enabled ? '已暂停' : '已启用'}`, detail: item.name || item.id })
    } catch (err) {
      setError(formatAiSreError(err))
    }
  }

  const handleCreate = async data => {
    await aiSreApi.workflows.create(data)
    setShowForm(false)
    addEvidence({ category: 'workflow', title: '工作流已创建', detail: data.name })
    loadList()
  }

  useEffect(() => { loadList() }, [])
  useEffect(() => { if (query.workflowId) loadRuns(query.workflowId) }, [query.workflowId])

  const selected = rows.find(r => r.id === query.workflowId)

  return (
    <section className='fx-aisre-split'>
      <div className='fx-aisre-panel'>
        <div className='fx-aisre-toolbar'>
          <h2>工作流列表</h2>
          <button type='button' onClick={loadList}>{loading ? '读取中...' : '刷新'}</button>
          <button type='button' onClick={() => setShowForm(true)}>+ 创建</button>
        </div>
        <ErrorBox>{error}</ErrorBox>
        {showForm && <WorkflowForm onSave={handleCreate} onCancel={() => setShowForm(false)} />}
        {!rows.length && !error && !showForm && <Empty>暂无工作流</Empty>}
        <table className='fx-aisre-table'>
          <thead>
            <tr><th>名称</th><th>触发</th><th>状态</th><th>最近运行</th><th>操作</th></tr>
          </thead>
          <tbody>
            {rows.map(item => (
              <tr key={item.id} className={query.workflowId === item.id ? 'is-active' : ''} onClick={() => onNavigate({ section: 'workflow', workflowId: item.id })}>
                <td><strong>{item.name || item.id}</strong></td>
                <td>{item.trigger || item.trigger_type || '手动'}</td>
                <td><StatusPill tone={item.enabled !== false ? 'ok' : 'blocked'}>{item.enabled !== false ? '运行中' : '已暂停'}</StatusPill></td>
                <td>{item.last_run ? new Date(item.last_run).toLocaleString('zh-CN', { hour12: false }) : '-'}</td>
                <td>
                  <button type='button' className='fx-aisre-btn-sm' onClick={e => { e.stopPropagation(); handleToggle(item) }}>
                    {item.enabled !== false ? '暂停' : '启用'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className='fx-aisre-panel'>
        <h2>执行历史</h2>
        {!selected && <Empty>请选择工作流查看执行历史</Empty>}
        {selected && (
          <>
            <p className='fx-aisre-note'>工作流: {selected.name || selected.id}</p>
            <Blocked>{AISRE_BLOCKERS.workflowRun}</Blocked>
            <RunHistory runs={runs} />
          </>
        )}
      </div>
    </section>
  )
}
