import React, { useEffect, useState } from 'react'
import { cmdbApi } from '../api/cmdb.js'

const VALUE_TYPES = [
  { value: 'char', label: '字符串' },
  { value: 'int', label: '整数' },
  { value: 'float', label: '浮点数' },
  { value: 'ip', label: 'IP地址' },
  { value: 'boolean', label: '布尔' },
  { value: 'enum', label: '枚举' },
  { value: 'array', label: '数组' },
  { value: 'struct', label: '结构体' },
  { value: 'file', label: '文件' },
  { value: 'image', label: '图片' },
]

const TAG_OPTIONS = ['基本信息', '系统资源', '系统信息', 'Agent', '自定义']

const emptyForm = () => ({
  label: '', attr: '', value_type: 'char', tag: '基本信息',
  required: false, unique: false, discovery: false,
  sort: 0, unit: '', options: '', default_val: '',
})

export function ModelDetailSection({ query, onNavigate }) {
  const modelId = query?.id
  const [model, setModel] = useState(null)
  const [attributes, setAttributes] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState(null)
  const [form, setForm] = useState(emptyForm())

  const load = async () => {
    if (!modelId) return
    setLoading(true)
    setError('')
    try {
      const [modelData, attrData] = await Promise.all([
        cmdbApi.objects.get(modelId),
        cmdbApi.attributes.list(modelId),
      ])
      setModel(modelData)
      setAttributes(Array.isArray(attrData) ? attrData : (attrData?.items || attrData?.list || []))
    } catch (err) {
      setError(err?.message || '加载失败')
    }
    setLoading(false)
  }

  useEffect(() => { load() }, [modelId])

  const openCreate = () => {
    setEditingId(null)
    setForm(emptyForm())
    setShowForm(true)
  }

  const openEdit = (attr) => {
    setEditingId(attr.id)
    setForm({
      label: attr.label || '',
      attr: attr.attr || '',
      value_type: attr.value_type || 'char',
      tag: attr.tag || '基本信息',
      required: !!attr.required,
      unique: !!attr.unique,
      discovery: !!attr.discovery,
      sort: attr.sort || 0,
      unit: attr.unit || '',
      options: attr.options ? JSON.stringify(attr.options) : '',
      default_val: attr.default_val || '',
    })
    setShowForm(true)
  }

  const handleSave = async () => {
    if (!form.label || !form.attr) return
    const body = {
      ...form,
      sort: Number(form.sort) || 0,
      options: form.options ? (() => { try { return JSON.parse(form.options) } catch { return form.options } })() : undefined,
    }
    try {
      if (editingId) {
        await cmdbApi.attributes.update(editingId, body)
      } else {
        await cmdbApi.attributes.create(modelId, body)
      }
      setShowForm(false)
      load()
    } catch (err) {
      setError(err?.message || '保存失败')
    }
  }

  const handleDelete = async (attrId) => {
    if (!confirm('确定删除该属性？')) return
    try {
      await cmdbApi.attributes.remove(attrId)
      load()
    } catch (err) {
      setError(err?.message || '删除失败')
    }
  }

  const updateField = (key, value) => setForm(prev => ({ ...prev, [key]: value }))

  return (
    <section style={{ padding: 16 }}>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
        <button
          type="button"
          onClick={() => onNavigate({ section: 'models' })}
          style={{ padding: '6px 14px', borderRadius: 6, border: '1px solid var(--fx-border)', background: 'transparent', color: 'var(--fx-ink)', cursor: 'pointer', fontSize: 13 }}
        >返回</button>
        <h2 style={{ margin: 0, fontSize: 18, color: 'var(--fx-ink)' }}>{model?.name || '模型详情'}</h2>
      </div>

      {error && <p style={{ color: '#e53e3e', fontSize: 13, marginBottom: 12 }}>{error}</p>}
      {loading && <p style={{ color: 'var(--fx-muted)', fontSize: 13 }}>加载中...</p>}

      {/* Toolbar */}
      <div style={{ marginBottom: 12 }}>
        <button
          type="button"
          onClick={openCreate}
          style={{ background: 'var(--fx-blue)', color: '#fff', border: 'none', borderRadius: 6, padding: '6px 14px', cursor: 'pointer', fontSize: 13 }}
        >添加属性</button>
      </div>

      {/* Attributes table */}
      <div className="fx-assets-table">
        <table>
          <thead>
            <tr>
              <th>属性名</th><th>标识符</th><th>类型</th><th>分组</th>
              <th>必填</th><th>唯一</th><th>自动发现</th><th>排序</th><th>操作</th>
            </tr>
          </thead>
          <tbody>
            {attributes.map(attr => (
              <tr key={attr.id}>
                <td>{attr.label}</td>
                <td><code style={{ fontSize: 12 }}>{attr.attr}</code></td>
                <td>{VALUE_TYPES.find(t => t.value === attr.value_type)?.label || attr.value_type}</td>
                <td>{attr.tag || '-'}</td>
                <td>{attr.required ? '是' : '否'}</td>
                <td>{attr.unique ? '是' : '否'}</td>
                <td>{attr.discovery ? '是' : '否'}</td>
                <td>{attr.sort ?? 0}</td>
                <td>
                  <button type="button" onClick={() => openEdit(attr)} style={{ marginRight: 6, padding: '2px 8px', borderRadius: 4, border: '1px solid var(--fx-border)', background: 'transparent', color: 'var(--fx-blue)', cursor: 'pointer', fontSize: 12 }}>编辑</button>
                  <button type="button" onClick={() => handleDelete(attr.id)} style={{ padding: '2px 8px', borderRadius: 4, border: '1px solid var(--fx-border)', background: 'transparent', color: '#e53e3e', cursor: 'pointer', fontSize: 12 }}>删除</button>
                </td>
              </tr>
            ))}
            {!attributes.length && !loading && (
              <tr><td colSpan="9" style={{ color: 'var(--fx-muted)' }}>暂无属性，点击"添加属性"开始定义模型字段。</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Attribute form modal */}
      {showForm && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.4)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000,
        }} onClick={() => setShowForm(false)}>
          <div style={{
            background: 'var(--fx-panel)', border: '1px solid var(--fx-border)',
            borderRadius: 8, padding: 24, width: 480, maxHeight: '80vh', overflowY: 'auto',
            boxShadow: 'var(--fx-shadow)',
          }} onClick={e => e.stopPropagation()}>
            <h3 style={{ margin: '0 0 16px', fontSize: 16, color: 'var(--fx-ink)' }}>
              {editingId ? '编辑属性' : '添加属性'}
            </h3>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
              <FormField label="属性名" value={form.label} onChange={v => updateField('label', v)} />
              <FormField label="标识符" value={form.attr} onChange={v => updateField('attr', v)} />
              <label style={{ display: 'block', fontSize: 13, color: 'var(--fx-ink)' }}>
                类型
                <select value={form.value_type} onChange={e => updateField('value_type', e.target.value)} style={inputStyle}>
                  {VALUE_TYPES.map(t => <option key={t.value} value={t.value}>{t.label}</option>)}
                </select>
              </label>
              <label style={{ display: 'block', fontSize: 13, color: 'var(--fx-ink)' }}>
                分组
                <select value={form.tag} onChange={e => updateField('tag', e.target.value)} style={inputStyle}>
                  {TAG_OPTIONS.map(t => <option key={t} value={t}>{t}</option>)}
                </select>
              </label>
              <FormField label="排序" value={form.sort} onChange={v => updateField('sort', v)} type="number" />
              <FormField label="单位" value={form.unit} onChange={v => updateField('unit', v)} />
              <FormField label="默认值" value={form.default_val} onChange={v => updateField('default_val', v)} />
              <div style={{ display: 'flex', gap: 16, alignItems: 'center', gridColumn: '1 / -1' }}>
                <CheckField label="必填" checked={form.required} onChange={v => updateField('required', v)} />
                <CheckField label="唯一" checked={form.unique} onChange={v => updateField('unique', v)} />
                <CheckField label="自动发现" checked={form.discovery} onChange={v => updateField('discovery', v)} />
              </div>
              {form.value_type === 'enum' && (
                <label style={{ display: 'block', fontSize: 13, color: 'var(--fx-ink)', gridColumn: '1 / -1' }}>
                  选项 (JSON数组)
                  <textarea
                    value={form.options} onChange={e => updateField('options', e.target.value)}
                    rows={3}
                    style={{ ...inputStyle, resize: 'vertical' }}
                    placeholder='["选项1","选项2"]'
                  />
                </label>
              )}
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 16 }}>
              <button type="button" onClick={() => setShowForm(false)} style={{ padding: '6px 14px', borderRadius: 6, border: '1px solid var(--fx-border)', background: 'transparent', color: 'var(--fx-ink)', cursor: 'pointer', fontSize: 13 }}>取消</button>
              <button type="button" onClick={handleSave} style={{ padding: '6px 14px', borderRadius: 6, border: 'none', background: 'var(--fx-blue)', color: '#fff', cursor: 'pointer', fontSize: 13 }}>保存</button>
            </div>
          </div>
        </div>
      )}
    </section>
  )
}

const inputStyle = {
  display: 'block', width: '100%', marginTop: 4, padding: '6px 8px',
  borderRadius: 4, border: '1px solid var(--fx-border)',
  background: 'var(--fx-bg)', color: 'var(--fx-ink)', fontSize: 13,
}

function FormField({ label, value, onChange, type = 'text' }) {
  return (
    <label style={{ display: 'block', fontSize: 13, color: 'var(--fx-ink)' }}>
      {label}
      <input
        type={type} value={value} onChange={e => onChange(e.target.value)}
        style={inputStyle}
      />
    </label>
  )
}

function CheckField({ label, checked, onChange }) {
  return (
    <label style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 13, color: 'var(--fx-ink)', cursor: 'pointer' }}>
      <input type="checkbox" checked={checked} onChange={e => onChange(e.target.checked)} />
      {label}
    </label>
  )
}
