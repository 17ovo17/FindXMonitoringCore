import React, { useEffect, useState } from 'react'
import { formatLogError, LOG_BLOCKERS, logsApi } from '../api/logs.js'
import { pipelineSteps } from './logsModel.js'
import { Blocked, JsonPreview, Status } from './LogsShared.jsx'

const RULE_TYPES = [
  { value: 'grok', label: 'Grok' },
  { value: 'json', label: 'JSON' },
  { value: 'regex', label: 'Regex' },
  { value: 'rename', label: 'Rename' },
  { value: 'remove', label: 'Remove' },
]

const emptyRule = () => ({ type: 'grok', config: '' })
const emptyPipeline = () => ({ id: '', name: '', description: '', rules: [emptyRule()] })

/**
 * DEGRADE-051: Pipeline 管理
 * Pipeline 列表 + 创建/编辑表单（名称/描述/处理规则列表）
 */
export function PipelinesSection() {
  const [pipelines, setPipelines] = useState([])
  const [selected, setSelected] = useState(null)
  const [blocked, setBlocked] = useState('')
  const [preview, setPreview] = useState(null)
  const [editing, setEditing] = useState(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let alive = true
    logsApi.pipelines.list('latest').then(resp => {
      if (!alive) return
      const items = Array.isArray(resp) ? resp : Array.isArray(resp?.items) ? resp.items : []
      setPipelines(items)
      setSelected(items[0] || null)
      setBlocked(resp?.blocker || '')
    }).catch(err => {
      if (!alive) return
      setPipelines([])
      setSelected(null)
      setBlocked(formatLogError(err))
    })
    return () => { alive = false }
  }, [])

  const visible = pipelines.length
    ? pipelines
    : pipelineSteps.map(step => ({ ...step, name: step.title, description: step.desc, status: 'blocked', rules: [] }))

  const startCreate = () => {
    setEditing(emptyPipeline())
    setSelected(null)
  }

  const startEdit = (pipeline) => {
    setEditing({ ...pipeline, rules: pipeline.rules?.length ? pipeline.rules : [emptyRule()] })
  }

  const cancelEdit = () => setEditing(null)

  const savePipeline = async () => {
    if (!editing) return
    setSaving(true)
    try {
      if (editing.id) {
        const resp = await logsApi.pipelines.update(editing.id, editing)
        if (resp?.blocker) { setBlocked(resp.blocker) }
        else {
          setPipelines(prev => prev.map(p => p.id === editing.id ? { ...p, ...editing } : p))
          setSelected(editing)
          setEditing(null)
          setBlocked('')
        }
      } else {
        const resp = await logsApi.pipelines.save(editing)
        if (resp?.blocker) { setBlocked(resp.blocker) }
        else {
          const newPipeline = resp?.id ? resp : { ...editing, id: Date.now().toString() }
          setPipelines(prev => [...prev, newPipeline])
          setSelected(newPipeline)
          setEditing(null)
          setBlocked('')
        }
      }
    } catch (err) {
      if (err?.status === 404 || err?.status === 501) {
        setBlocked('后端不支持 Pipeline 保存接口。')
      } else {
        setBlocked(formatLogError(err))
      }
    } finally {
      setSaving(false)
    }
  }

  const deletePipeline = async (pipeline) => {
    if (!pipeline?.id) return
    try {
      await logsApi.pipelines.remove(pipeline.id)
      setPipelines(prev => prev.filter(p => p.id !== pipeline.id))
      if (selected?.id === pipeline.id) setSelected(null)
    } catch (err) {
      if (err?.status === 404 || err?.status === 501) {
        setBlocked('后端不支持 Pipeline 删除接口。')
      } else {
        setBlocked(formatLogError(err))
      }
    }
  }

  const previewPipeline = async () => {
    try {
      const target = editing || selected || visible[0]
      const resp = await logsApi.pipelines.preview({ parser: 'json', sample: '{"message":"demo"}', pipeline: target })
      setPreview(resp)
      setBlocked(resp?.blocker || '')
    } catch (err) {
      setPreview(null)
      setBlocked(formatLogError(err))
    }
  }

  return (
    <section className='fx-logs-split'>
      <div className='fx-logs-list'>
        <button type='button' className='fx-qb__btn' style={{ marginBottom: 8 }} onClick={startCreate}>+ 新建 Pipeline</button>
        {visible.map(step => (
          <button
            key={step.id}
            className={selected?.id === step.id ? 'is-active' : ''}
            type='button'
            onClick={() => { setSelected(step); setEditing(null) }}
          >
            <strong>{step.name || step.title}</strong>
            <span>{step.description || step.desc}</span>
          </button>
        ))}
      </div>
      <div className='fx-logs-panel'>
        {editing ? (
          <PipelineForm
            pipeline={editing}
            onChange={setEditing}
            onSave={savePipeline}
            onCancel={cancelEdit}
            saving={saving}
          />
        ) : (
          <PipelineDetail
            pipeline={selected}
            pipelines={pipelines}
            onEdit={startEdit}
            onDelete={deletePipeline}
            onPreview={previewPipeline}
            preview={preview}
          />
        )}
        <Status ok={pipelines.length > 0}>{pipelines.length ? '配置契约已加载' : '待契约'}</Status>
        {!pipelines.length && <Blocked>{LOG_BLOCKERS.pipeline}</Blocked>}
        {blocked && <Blocked>{blocked}</Blocked>}
      </div>
    </section>
  )
}

function PipelineDetail({ pipeline, pipelines, onEdit, onDelete, onPreview, preview }) {
  if (!pipeline) {
    return <p style={{ color: 'var(--fx-text-weak,#66758d)' }}>选择一个 Pipeline 或新建。</p>
  }
  return (
    <>
      <h3>{pipeline.name || pipeline.title || '接入管道'}</h3>
      <p>{pipeline.description || pipeline.desc || '管道配置契约未加载。'}</p>
      {pipeline.rules?.length > 0 && (
        <div style={{ marginBottom: 12 }}>
          <h4 style={{ fontSize: 13, margin: '8px 0 4px' }}>处理规则（{pipeline.rules.length}）</h4>
          {pipeline.rules.map((rule, idx) => (
            <div key={idx} style={{ fontSize: 12, padding: '4px 8px', background: '#f8fbff', borderRadius: 4, marginBottom: 4 }}>
              <strong>{rule.type}</strong>: {rule.config || '-'}
            </div>
          ))}
        </div>
      )}
      <JsonPreview value={preview || { pipeline: pipeline, sample_log: '<LOG_SAMPLE>', output: '<PIPELINE_PREVIEW>' }} />
      <div className='fx-logs-toolbar' style={{ display: 'flex', gap: 8, marginTop: 12 }}>
        {pipelines.length > 0 && <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={() => onEdit(pipeline)}>编辑</button>}
        {pipelines.length > 0 && <button type='button' className='fx-qb__btn fx-qb__btn--danger' onClick={() => onDelete(pipeline)}>删除</button>}
        <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={onPreview}>预览验证</button>
      </div>
    </>
  )
}

function PipelineForm({ pipeline, onChange, onSave, onCancel, saving }) {
  const updateField = (key, value) => onChange({ ...pipeline, [key]: value })

  const addRule = () => onChange({ ...pipeline, rules: [...pipeline.rules, emptyRule()] })
  const removeRule = (idx) => onChange({ ...pipeline, rules: pipeline.rules.filter((_, i) => i !== idx) })
  const updateRule = (idx, patch) => onChange({
    ...pipeline,
    rules: pipeline.rules.map((r, i) => i === idx ? { ...r, ...patch } : r),
  })

  return (
    <div>
      <h3>{pipeline.id ? '编辑 Pipeline' : '新建 Pipeline'}</h3>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginBottom: 12 }}>
        <label className='fx-logs-field'>
          <span>名称</span>
          <input value={pipeline.name} onChange={e => updateField('name', e.target.value)} placeholder='Pipeline 名称' />
        </label>
        <label className='fx-logs-field'>
          <span>描述</span>
          <input value={pipeline.description} onChange={e => updateField('description', e.target.value)} placeholder='Pipeline 描述' />
        </label>
      </div>
      <h4 style={{ fontSize: 13, margin: '8px 0' }}>处理规则</h4>
      {pipeline.rules.map((rule, idx) => (
        <div key={idx} style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 6 }}>
          <select className='fx-qb__select' value={rule.type} onChange={e => updateRule(idx, { type: e.target.value })}>
            {RULE_TYPES.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
          </select>
          <input
            className='fx-qb__search'
            style={{ flex: 1 }}
            value={rule.config}
            onChange={e => updateRule(idx, { config: e.target.value })}
            placeholder={`${rule.type} 配置（如 pattern、字段名等）`}
          />
          <button type='button' className='fx-qb__btn fx-qb__btn--danger' style={{ minWidth: 32, padding: 0 }} onClick={() => removeRule(idx)}>×</button>
        </div>
      ))}
      <button type='button' className='fx-qb__btn fx-qb__btn--ghost' style={{ marginTop: 4 }} onClick={addRule}>+ 添加规则</button>
      <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
        <button type='button' className='fx-qb__btn' onClick={onSave} disabled={saving || !pipeline.name}>
          {saving ? '保存中...' : '保存'}
        </button>
        <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={onCancel}>取消</button>
      </div>
    </div>
  )
}
