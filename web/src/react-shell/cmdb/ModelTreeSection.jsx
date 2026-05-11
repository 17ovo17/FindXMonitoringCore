import React, { useEffect, useState } from 'react'
import { cmdbApi } from '../api/cmdb.js'

const defaultCategories = [
  { id: 'compute', name: '计算资源', icon: '🖥' },
  { id: 'software', name: '系统软件', icon: '📦' },
  { id: 'network', name: '网络设备', icon: '🌐' },
  { id: 'storage', name: '存储设备', icon: '💾' },
  { id: 'custom', name: '自定义', icon: '⚙' },
]

export function ModelTreeSection({ onNavigate }) {
  const [categories, setCategories] = useState(defaultCategories)
  const [objects, setObjects] = useState([])
  const [counts, setCounts] = useState({})
  const [selectedCategory, setSelectedCategory] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: '', identifier: '', category_id: '' })

  const loadTree = async () => {
    setLoading(true)
    setError('')
    try {
      const data = await cmdbApi.tree()
      if (data?.categories?.length) setCategories(data.categories)
      setObjects(Array.isArray(data?.objects) ? data.objects : [])
      setCounts(data?.counts || {})
    } catch (err) {
      setError(err?.message || '加载失败')
    }
    setLoading(false)
  }

  useEffect(() => { loadTree() }, [])

  const filtered = selectedCategory
    ? objects.filter(obj => obj.category_id === selectedCategory)
    : objects

  const handleCreate = async () => {
    if (!form.name || !form.identifier) return
    try {
      await cmdbApi.objects.create({
        name: form.name,
        identifier: form.identifier,
        category_id: form.category_id || (selectedCategory || ''),
      })
      setShowCreate(false)
      setForm({ name: '', identifier: '', category_id: '' })
      loadTree()
    } catch (err) {
      setError(err?.message || '创建失败')
    }
  }

  return (
    <section style={{ display: 'flex', gap: 16, padding: 16, minHeight: 400 }}>
      {/* Left panel: category tree */}
      <aside style={{
        width: 200, flexShrink: 0,
        background: 'var(--fx-panel)', border: '1px solid var(--fx-border)',
        borderRadius: 8, padding: 12,
      }}>
        <h3 style={{ margin: '0 0 12px', fontSize: 14, color: 'var(--fx-ink)' }}>模型分类</h3>
        <ul style={{ listStyle: 'none', margin: 0, padding: 0 }}>
          <li
            style={{
              padding: '6px 10px', borderRadius: 4, cursor: 'pointer', marginBottom: 4,
              background: !selectedCategory ? 'var(--fx-blue)' : 'transparent',
              color: !selectedCategory ? '#fff' : 'var(--fx-ink)',
            }}
            onClick={() => setSelectedCategory(null)}
          >全部</li>
          {categories.map(cat => (
            <li
              key={cat.id}
              style={{
                padding: '6px 10px', borderRadius: 4, cursor: 'pointer', marginBottom: 4,
                background: selectedCategory === cat.id ? 'var(--fx-blue)' : 'transparent',
                color: selectedCategory === cat.id ? '#fff' : 'var(--fx-ink)',
              }}
              onClick={() => setSelectedCategory(cat.id)}
            >{cat.icon || ''} {cat.name}</li>
          ))}
        </ul>
      </aside>

      {/* Right panel: model cards */}
      <div style={{ flex: 1 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <h2 style={{ margin: 0, fontSize: 18, color: 'var(--fx-ink)' }}>模型列表</h2>
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            style={{
              background: 'var(--fx-blue)', color: '#fff', border: 'none',
              borderRadius: 6, padding: '6px 14px', cursor: 'pointer', fontSize: 13,
            }}
          >新建模型</button>
        </div>

        {error && <p style={{ color: '#e53e3e', fontSize: 13 }}>{error}</p>}
        {loading && <p style={{ color: 'var(--fx-muted)', fontSize: 13 }}>加载中...</p>}

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 12 }}>
          {filtered.map(obj => (
            <article
              key={obj.id}
              onClick={() => onNavigate({ section: 'model-detail', id: obj.id })}
              style={{
                background: 'var(--fx-panel)', border: '1px solid var(--fx-border)',
                borderRadius: 8, padding: 16, cursor: 'pointer',
                boxShadow: 'var(--fx-shadow)', transition: 'border-color 0.15s',
              }}
            >
              <div style={{ fontSize: 28, marginBottom: 8 }}>{obj.icon || '📋'}</div>
              <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--fx-ink)', marginBottom: 4 }}>{obj.name}</div>
              <div style={{ fontSize: 12, color: 'var(--fx-muted)' }}>实例数: {counts[obj.id] ?? 0}</div>
            </article>
          ))}
          {!loading && !filtered.length && (
            <p style={{ color: 'var(--fx-muted)', fontSize: 13, gridColumn: '1 / -1' }}>暂无模型数据</p>
          )}
        </div>
      </div>

      {/* Create modal */}
      {showCreate && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.4)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000,
        }} onClick={() => setShowCreate(false)}>
          <div style={{
            background: 'var(--fx-panel)', border: '1px solid var(--fx-border)',
            borderRadius: 8, padding: 24, width: 360, boxShadow: 'var(--fx-shadow)',
          }} onClick={e => e.stopPropagation()}>
            <h3 style={{ margin: '0 0 16px', fontSize: 16, color: 'var(--fx-ink)' }}>新建模型</h3>
            <label style={{ display: 'block', marginBottom: 12, fontSize: 13, color: 'var(--fx-ink)' }}>
              模型名称
              <input
                value={form.name} onChange={e => setForm({ ...form, name: e.target.value })}
                style={{ display: 'block', width: '100%', marginTop: 4, padding: '6px 8px', borderRadius: 4, border: '1px solid var(--fx-border)', background: 'var(--fx-bg)', color: 'var(--fx-ink)' }}
              />
            </label>
            <label style={{ display: 'block', marginBottom: 12, fontSize: 13, color: 'var(--fx-ink)' }}>
              标识符
              <input
                value={form.identifier} onChange={e => setForm({ ...form, identifier: e.target.value })}
                style={{ display: 'block', width: '100%', marginTop: 4, padding: '6px 8px', borderRadius: 4, border: '1px solid var(--fx-border)', background: 'var(--fx-bg)', color: 'var(--fx-ink)' }}
              />
            </label>
            <label style={{ display: 'block', marginBottom: 16, fontSize: 13, color: 'var(--fx-ink)' }}>
              分类
              <select
                value={form.category_id} onChange={e => setForm({ ...form, category_id: e.target.value })}
                style={{ display: 'block', width: '100%', marginTop: 4, padding: '6px 8px', borderRadius: 4, border: '1px solid var(--fx-border)', background: 'var(--fx-bg)', color: 'var(--fx-ink)' }}
              >
                <option value="">请选择</option>
                {categories.map(cat => <option key={cat.id} value={cat.id}>{cat.name}</option>)}
              </select>
            </label>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button type="button" onClick={() => setShowCreate(false)} style={{ padding: '6px 14px', borderRadius: 6, border: '1px solid var(--fx-border)', background: 'transparent', color: 'var(--fx-ink)', cursor: 'pointer', fontSize: 13 }}>取消</button>
              <button type="button" onClick={handleCreate} style={{ padding: '6px 14px', borderRadius: 6, border: 'none', background: 'var(--fx-blue)', color: '#fff', cursor: 'pointer', fontSize: 13 }}>确定</button>
            </div>
          </div>
        </div>
      )}
    </section>
  )
}
