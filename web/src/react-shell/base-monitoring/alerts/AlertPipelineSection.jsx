import React, { useEffect, useMemo, useState } from 'react'
import { get, post, put, del } from '../../api/http.js'
import { normalizeList } from '../../api/http.js'
import { filterText, makeError } from './alertModel.js'
import { useConfirm } from '../../shared/ConfirmModal.jsx'
import { FxDrawer } from '../../shared/FxDrawer.jsx'

const PAGE_SIZE = 20

const PROCESSOR_TYPES = [
  { value: 'relabel', label: '重标记', icon: '🏷' },
  { value: 'drop', label: '丢弃', icon: '🗑' },
  { value: 'callback', label: '回调', icon: '🔗' },
  { value: 'enrich', label: '富化', icon: '📝' },
  { value: 'aisummary', label: 'AI 摘要', icon: '🤖' },
]

const emptyCondition = () => ({ key: '', operator: '=', value: '' })

const emptyProcessor = (type = 'relabel') => ({
  type,
  enabled: true,
  conditions: [emptyCondition()],
  config: type === 'relabel' ? { action: 'set', key: '', value: '' }
    : type === 'callback' ? { url: '', method: 'POST', headers: {} }
    : type === 'enrich' ? { annotations: {} }
    : type === 'aisummary' ? { prompt_template: '', model: 'default', target_field: 'ai_summary' }
    : {},
})

const emptyForm = () => ({
  id: '',
  name: '',
  priority: 0,
  enabled: true,
  processors: [],
})

const OPERATORS = ['=', '!=', '=~', '!~']

function ConditionsEditor({ conditions, onChange }) {
  const add = () => onChange([...conditions, emptyCondition()])
  const remove = (idx) => onChange(conditions.filter((_, i) => i !== idx))
  const update = (idx, field, val) => {
    const next = [...conditions]
    next[idx] = { ...next[idx], [field]: val }
    onChange(next)
  }
  return (
    <div style={{ marginTop: 8 }}>
      <span style={{ fontSize: 12, color: '#475569' }}>条件</span>
      {conditions.map((cond, idx) => (
        <div key={idx} className='fx-alert-pipeline-row' style={{ marginTop: 4 }}>
          <input value={cond.key} onChange={(e) => update(idx, 'key', e.target.value)} placeholder='key' style={{ width: 120 }} />
          <select value={cond.operator} onChange={(e) => update(idx, 'operator', e.target.value)}>
            {OPERATORS.map((op) => <option key={op} value={op}>{op}</option>)}
          </select>
          <input value={cond.value} onChange={(e) => update(idx, 'value', e.target.value)} placeholder='value' style={{ flex: 1 }} />
          <button type='button' onClick={() => remove(idx)} disabled={conditions.length <= 1}>-</button>
        </div>
      ))}
      <button type='button' onClick={add} style={{ marginTop: 4 }}>+ 条件</button>
    </div>
  )
}

function ProcessorCard({ processor, index, total, onChange, onRemove, onMove }) {
  const typeInfo = PROCESSOR_TYPES.find((t) => t.value === processor.type) || PROCESSOR_TYPES[0]
  const updateConfig = (key, val) => onChange({ ...processor, config: { ...processor.config, [key]: val } })
  const updateConditions = (conditions) => onChange({ ...processor, conditions })

  const condSummary = (processor.conditions || [])
    .filter((c) => c.key)
    .map((c) => `${c.key}${c.operator}${c.value}`)
    .join(', ')

  return (
    <div className='fx-alert-trigger-item'>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
        <span style={{ fontSize: 16 }}>{typeInfo.icon}</span>
        <strong style={{ fontSize: 13 }}>{typeInfo.label}</strong>
        {condSummary && <span style={{ fontSize: 11, color: '#64748b' }}>({condSummary})</span>}
        <span style={{ flex: 1 }} />
        <button type='button' disabled={index === 0} onClick={() => onMove(index, index - 1)} style={{ padding: '0 6px' }}>↑</button>
        <button type='button' disabled={index === total - 1} onClick={() => onMove(index, index + 1)} style={{ padding: '0 6px' }}>↓</button>
        <label className='fx-alert-check-inline'>
          <input type='checkbox' checked={processor.enabled} onChange={(e) => onChange({ ...processor, enabled: e.target.checked })} />
          启用
        </label>
        <button type='button' onClick={onRemove} style={{ color: '#dc2626' }}>删除</button>
      </div>

      {processor.type === 'relabel' && (
        <div className='fx-alert-pipeline-row'>
          <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 12 }}>
            <span>动作</span>
            <select value={processor.config.action || 'set'} onChange={(e) => updateConfig('action', e.target.value)}>
              <option value='set'>set</option>
              <option value='delete'>delete</option>
            </select>
          </label>
          <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 12, flex: 1 }}>
            <span>Key</span>
            <input value={processor.config.key || ''} onChange={(e) => updateConfig('key', e.target.value)} />
          </label>
          {processor.config.action !== 'delete' && (
            <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 12, flex: 1 }}>
              <span>Value</span>
              <input value={processor.config.value || ''} onChange={(e) => updateConfig('value', e.target.value)} />
            </label>
          )}
        </div>
      )}

      {processor.type === 'callback' && (
        <div className='fx-alert-pipeline-row' style={{ flexDirection: 'column', alignItems: 'stretch' }}>
          <div style={{ display: 'flex', gap: 6 }}>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 12 }}>
              <span>Method</span>
              <select value={processor.config.method || 'POST'} onChange={(e) => updateConfig('method', e.target.value)}>
                <option value='GET'>GET</option>
                <option value='POST'>POST</option>
                <option value='PUT'>PUT</option>
              </select>
            </label>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 12, flex: 1 }}>
              <span>URL</span>
              <input value={processor.config.url || ''} onChange={(e) => updateConfig('url', e.target.value)} placeholder='https://...' />
            </label>
          </div>
          <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 12, marginTop: 6 }}>
            <span>Headers (JSON)</span>
            <textarea value={typeof processor.config.headers === 'string' ? processor.config.headers : JSON.stringify(processor.config.headers || {}, null, 2)} onChange={(e) => updateConfig('headers', e.target.value)} rows={2} />
          </label>
        </div>
      )}

      {processor.type === 'enrich' && (
        <EnrichEditor annotations={processor.config.annotations || {}} onChange={(annotations) => updateConfig('annotations', annotations)} />
      )}

      {processor.type === 'aisummary' && (
        <div className='fx-alert-pipeline-row' style={{ flexDirection: 'column', alignItems: 'stretch' }}>
          <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 12 }}>
            <span>Prompt 模板</span>
            <textarea
              value={processor.config.prompt_template || ''}
              onChange={(e) => updateConfig('prompt_template', e.target.value)}
              rows={3}
              placeholder='请根据以下告警事件生成摘要：{{.Labels}} {{.Annotations}}'
            />
          </label>
          <div style={{ display: 'flex', gap: 6, marginTop: 6 }}>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 12, flex: 1 }}>
              <span>模型</span>
              <select value={processor.config.model || 'default'} onChange={(e) => updateConfig('model', e.target.value)}>
                <option value='default'>默认</option>
                <option value='gpt-4'>GPT-4</option>
                <option value='gpt-3.5-turbo'>GPT-3.5</option>
                <option value='claude'>Claude</option>
              </select>
            </label>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 12, flex: 1 }}>
              <span>目标字段</span>
              <input value={processor.config.target_field || 'ai_summary'} onChange={(e) => updateConfig('target_field', e.target.value)} />
            </label>
          </div>
        </div>
      )}

      <ConditionsEditor conditions={processor.conditions || [emptyCondition()]} onChange={updateConditions} />
    </div>
  )
}
function EnrichEditor({ annotations, onChange }) {
  const pairs = Object.entries(annotations || {})
  const addPair = () => onChange({ ...annotations, '': '' })
  const removePair = (key) => {
    const next = { ...annotations }
    delete next[key]
    onChange(next)
  }
  const updatePair = (oldKey, newKey, newVal) => {
    const entries = Object.entries(annotations)
    const updated = entries.map(([k, v]) => k === oldKey ? [newKey, newVal] : [k, v])
    onChange(Object.fromEntries(updated))
  }
  return (
    <div style={{ marginTop: 4 }}>
      <span style={{ fontSize: 12, color: '#475569' }}>Annotations (key-value)</span>
      {pairs.map(([key, val], idx) => (
        <div key={idx} className='fx-alert-pipeline-row' style={{ marginTop: 4 }}>
          <input value={key} onChange={(e) => updatePair(key, e.target.value, val)} placeholder='key' style={{ flex: 1 }} />
          <input value={val} onChange={(e) => updatePair(key, key, e.target.value)} placeholder='value' style={{ flex: 1 }} />
          <button type='button' onClick={() => removePair(key)}>-</button>
        </div>
      ))}
      <button type='button' onClick={addPair} style={{ marginTop: 4 }}>+ 添加</button>
    </div>
  )
}

function PipelinePreview({ processors }) {
  const [input, setInput] = useState(JSON.stringify({
    labels: { alertname: 'HighCPU', instance: 'server-01', severity: 'critical' },
    annotations: { summary: 'CPU 使用率超过 90%' },
    status: 'firing',
    value: '95.2',
  }, null, 2))
  const [output, setOutput] = useState(null)
  const [previewError, setPreviewError] = useState(null)

  const runPreview = () => {
    try {
      const event = JSON.parse(input)
      let result = { ...event }

      for (const proc of processors) {
        if (!proc.enabled) continue

        // 检查条件
        const conditions = (proc.conditions || []).filter((c) => c.key)
        if (conditions.length > 0) {
          const matched = conditions.every((cond) => {
            const val = result.labels?.[cond.key] || ''
            if (cond.operator === '=') return val === cond.value
            if (cond.operator === '!=') return val !== cond.value
            if (cond.operator === '=~') { try { return new RegExp(cond.value).test(val) } catch { return false } }
            if (cond.operator === '!~') { try { return !new RegExp(cond.value).test(val) } catch { return true } }
            return false
          })
          if (!matched) continue
        }

        if (proc.type === 'relabel') {
          if (proc.config.action === 'set') {
            result = { ...result, labels: { ...result.labels, [proc.config.key]: proc.config.value } }
          } else if (proc.config.action === 'delete') {
            const labels = { ...result.labels }
            delete labels[proc.config.key]
            result = { ...result, labels }
          }
        } else if (proc.type === 'drop') {
          result = { ...result, _dropped: true }
          break
        } else if (proc.type === 'enrich') {
          result = { ...result, annotations: { ...result.annotations, ...proc.config.annotations } }
        } else if (proc.type === 'aisummary') {
          result = { ...result, annotations: { ...result.annotations, [proc.config.target_field || 'ai_summary']: '[AI 摘要将在运行时生成]' } }
        }
      }

      setOutput(result)
      setPreviewError(null)
    } catch (err) {
      setPreviewError(err.message)
      setOutput(null)
    }
  }

  return (
    <div className='fx-alert-pipeline-section' style={{ marginTop: 16, borderTop: '1px solid var(--fx-border, #e5e7eb)', paddingTop: 16 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
        <strong style={{ fontSize: 14 }}>事件预览</strong>
        <button type='button' onClick={runPreview}>运行预览</button>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
        <div>
          <span style={{ fontSize: 12, color: '#475569', display: 'block', marginBottom: 4 }}>输入事件 (JSON)</span>
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            rows={8}
            style={{ width: '100%', fontFamily: 'monospace', fontSize: 11, padding: 8, border: '1px solid var(--fx-border, #e5e7eb)', borderRadius: 4 }}
          />
        </div>
        <div>
          <span style={{ fontSize: 12, color: '#475569', display: 'block', marginBottom: 4 }}>输出结果</span>
          {previewError && <div style={{ color: '#dc2626', fontSize: 12 }}>{previewError}</div>}
          {output && (
            <pre style={{ fontSize: 11, padding: 8, background: 'var(--fx-bg-alt, #f8fafc)', border: '1px solid var(--fx-border, #e5e7eb)', borderRadius: 4, overflow: 'auto', maxHeight: 200 }}>
              {JSON.stringify(output, null, 2)}
            </pre>
          )}
          {!output && !previewError && <div style={{ color: '#9ca3af', fontSize: 12 }}>点击"运行预览"查看结果</div>}
        </div>
      </div>
    </div>
  )
}

function PipelineDrawer({ form, setForm, onSave, onClose, saving }) {
  const addProcessor = (type) => {
    setForm({ ...form, processors: [...form.processors, emptyProcessor(type)] })
  }
  const removeProcessor = (idx) => {
    setForm({ ...form, processors: form.processors.filter((_, i) => i !== idx) })
  }
  const updateProcessor = (idx, proc) => {
    const next = [...form.processors]
    next[idx] = proc
    setForm({ ...form, processors: next })
  }
  const moveProcessor = (from, to) => {
    const next = [...form.processors]
    const [item] = next.splice(from, 1)
    next.splice(to, 0, item)
    setForm({ ...form, processors: next })
  }

  return (
    <FxDrawer open title={form.id ? '编辑事件流水线' : '新建事件流水线'} width='xl' onClose={onClose}
      footer={
        <div style={{ display: 'flex', gap: 8 }}>
          <button type='button' className='is-primary' disabled={saving || !form.name.trim()} onClick={onSave}>{saving ? '保存中...' : '保存'}</button>
          <button type='button' onClick={onClose}>取消</button>
        </div>
      }
    >
      <div className='fx-alert-pipeline'>
        <div className='fx-alert-form' style={{ marginBottom: 16 }}>
          <label>
            <span>名称</span>
            <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder='流水线名称' />
          </label>
          <label>
            <span>优先级</span>
            <input type='number' value={form.priority} onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })} />
          </label>
          <label className='fx-alert-check'>
            <span>启用</span>
            <input type='checkbox' checked={form.enabled} onChange={(e) => setForm({ ...form, enabled: e.target.checked })} />
          </label>
        </div>

        <div className='fx-alert-pipeline-section'>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <strong style={{ fontSize: 14 }}>处理器 ({form.processors.length})</strong>
            <span style={{ flex: 1 }} />
            {PROCESSOR_TYPES.map((pt) => (
              <button key={pt.value} type='button' onClick={() => addProcessor(pt.value)}>+ {pt.label}</button>
            ))}
          </div>
          {form.processors.map((proc, idx) => (
            <ProcessorCard
              key={idx}
              processor={proc}
              index={idx}
              total={form.processors.length}
              onChange={(p) => updateProcessor(idx, p)}
              onRemove={() => removeProcessor(idx)}
              onMove={moveProcessor}
            />
          ))}
          {form.processors.length === 0 && <div className='fx-alert-empty'>暂无处理器，点击上方按钮添加</div>}
        </div>
        <PipelinePreview processors={form.processors} />
      </div>
    </FxDrawer>
  )
}
export function AlertPipelineSection() {
  const [pipelines, setPipelines] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [drawer, setDrawer] = useState(false)
  const [form, setForm] = useState(emptyForm())
  const [saving, setSaving] = useState(false)
  const { confirm, modal: confirmModal } = useConfirm()

  const loadPipelines = async () => {
    setLoading(true); setError('')
    try {
      const data = await get('/alert-pipelines')
      setPipelines(normalizeList(data))
      setPage(1)
    } catch (err) {
      setError(makeError(err, '加载事件流水线失败'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { loadPipelines() }, [])

  const filtered = useMemo(() => pipelines.filter((p) =>
    filterText([p.name], keyword)
  ), [pipelines, keyword])

  const total = filtered.length
  const paged = useMemo(() => filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE), [filtered, page])
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const openCreate = () => { setForm(emptyForm()); setDrawer(true) }
  const openEdit = (item) => {
    setForm({
      id: item.id || '',
      name: item.name || '',
      priority: item.priority || 0,
      enabled: item.enabled !== false,
      processors: Array.isArray(item.processors) ? item.processors.map((p) => ({
        type: p.type || 'relabel',
        enabled: p.enabled !== false,
        conditions: Array.isArray(p.conditions) && p.conditions.length ? p.conditions : [emptyCondition()],
        config: p.config || {},
      })) : [],
    })
    setDrawer(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = {
        name: form.name.trim(),
        priority: form.priority,
        enabled: form.enabled,
        processors: form.processors.map((p) => ({
          type: p.type,
          enabled: p.enabled,
          conditions: (p.conditions || []).filter((c) => c.key),
          config: p.type === 'callback' && typeof p.config.headers === 'string'
            ? { ...p.config, headers: JSON.parse(p.config.headers || '{}') }
            : p.config,
        })),
      }
      if (form.id) {
        await put(`/alert-pipelines/${encodeURIComponent(form.id)}`, payload)
      } else {
        await post('/alert-pipelines', payload)
      }
      setDrawer(false)
      await loadPipelines()
    } catch (err) {
      setError(makeError(err, '保存失败'))
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (item) => {
    const ok = await confirm({ title: '删除事件流水线', message: `确认删除「${item.name}」？`, confirmText: '删除', danger: true })
    if (!ok) return
    try {
      await del(`/alert-pipelines/${encodeURIComponent(item.id)}`)
      await loadPipelines()
    } catch (err) {
      setError(makeError(err, '删除失败'))
    }
  }

  const handleToggle = async (item) => {
    try {
      await put(`/alert-pipelines/${encodeURIComponent(item.id)}`, { ...item, enabled: !item.enabled })
      await loadPipelines()
    } catch (err) {
      setError(makeError(err, '切换状态失败'))
    }
  }

  return (
    <section className='fx-alert-section'>
      <div className='fx-alert-filterbar'>
        <button type='button' disabled={loading} onClick={loadPipelines}>{loading ? '刷新中...' : '刷新'}</button>
        <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder='搜索流水线名称' />
        <button type='button' className='is-primary' onClick={openCreate}>新建流水线</button>
      </div>
      {error && <div className='fx-alert-message is-error'>{error}</div>}
      <div className='fx-alert-table'>
        <table>
          <thead>
            <tr>
              <th>名称</th>
              <th>优先级</th>
              <th>处理器数</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {paged.map((item) => (
              <tr key={item.id}>
                <td>{item.name}</td>
                <td>{item.priority ?? 0}</td>
                <td>{Array.isArray(item.processors) ? item.processors.length : 0}</td>
                <td>
                  <button type='button' className={item.enabled ? 'fx-alert-state is-on' : 'fx-alert-state'} onClick={() => handleToggle(item)}>
                    {item.enabled ? '启用' : '禁用'}
                  </button>
                </td>
                <td>
                  <button type='button' className='fx-alert-link' onClick={() => openEdit(item)}>编辑</button>
                  <button type='button' className='fx-alert-link' style={{ color: '#dc2626', marginLeft: 8 }} onClick={() => handleDelete(item)}>删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!loading && paged.length === 0 && <div className='fx-alert-empty'>暂无事件流水线</div>}
      </div>
      {total > PAGE_SIZE && (
        <div className='fx-alert-pagination'>
          <span>共 {total} 条，第 {page}/{totalPages} 页</span>
          <button type='button' disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</button>
          <button type='button' disabled={page >= totalPages} onClick={() => setPage(page + 1)}>下一页</button>
        </div>
      )}
      {drawer && <PipelineDrawer form={form} setForm={setForm} onSave={handleSave} onClose={() => setDrawer(false)} saving={saving} />}
      {confirmModal}
    </section>
  )
}
