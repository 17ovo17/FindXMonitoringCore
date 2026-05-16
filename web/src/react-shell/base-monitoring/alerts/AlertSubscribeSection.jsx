import React, { useEffect, useMemo, useState } from 'react'
import { get, post, put, del } from '../../api/http.js'
import { normalizeList } from '../../api/http.js'
import { filterText, makeError, mapToPairs, severities } from './alertModel.js'
import { useConfirm } from '../../shared/ConfirmModal.jsx'

const PAGE_SIZE = 20

const emptyForm = () => ({
  id: '',
  name: '',
  labels: [{ key: '', value: '' }],
  severities: [],
  rule_ids: '',
  channel_ids: [],
  enabled: true,
})

const labelsToMap = (labels) => {
  const map = {}
  labels.forEach(({ key, value }) => { if (key.trim()) map[key.trim()] = value.trim() })
  return map
}

const mapToLabels = (map) => {
  if (!map || typeof map !== 'object') return [{ key: '', value: '' }]
  const entries = Object.entries(map)
  return entries.length ? entries.map(([key, value]) => ({ key, value: String(value) })) : [{ key: '', value: '' }]
}

function SubscribeModal({ form, setForm, onSave, onCancel, saving, channels }) {
  const addLabel = () => setForm({ ...form, labels: [...form.labels, { key: '', value: '' }] })
  const removeLabel = (idx) => setForm({ ...form, labels: form.labels.filter((_, i) => i !== idx) })
  const updateLabel = (idx, field, val) => {
    const next = [...form.labels]
    next[idx] = { ...next[idx], [field]: val }
    setForm({ ...form, labels: next })
  }
  const toggleSeverity = (sev) => {
    const has = form.severities.includes(sev)
    setForm({ ...form, severities: has ? form.severities.filter((s) => s !== sev) : [...form.severities, sev] })
  }
  const toggleChannel = (id) => {
    const has = form.channel_ids.includes(id)
    setForm({ ...form, channel_ids: has ? form.channel_ids.filter((c) => c !== id) : [...form.channel_ids, id] })
  }

  return (
    <div className='fx-alert-modal'>
      <div className='fx-alert-modal__body'>
        <header>
          <h2>{form.id ? '编辑订阅规则' : '新建订阅规则'}</h2>
          <button type='button' onClick={onCancel}>关闭</button>
        </header>
        <div className='fx-alert-form'>
          <label>
            <span>名称</span>
            <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder='订阅规则名称' />
          </label>
          <label>
            <span>规则 ID（逗号分隔）</span>
            <input value={form.rule_ids} onChange={(e) => setForm({ ...form, rule_ids: e.target.value })} placeholder='rule1,rule2' />
          </label>
          <div className='is-wide'>
            <span style={{ fontSize: 13, color: '#475569', marginBottom: 6, display: 'block' }}>级别</span>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
              {severities.map((s) => (
                <label key={s.value} className='fx-alert-check-inline'>
                  <input type='checkbox' checked={form.severities.includes(s.value)} onChange={() => toggleSeverity(s.value)} />
                  {s.label}
                </label>
              ))}
            </div>
          </div>
          <div className='is-wide'>
            <span style={{ fontSize: 13, color: '#475569', marginBottom: 6, display: 'block' }}>通知渠道</span>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
              {channels.map((ch) => (
                <label key={ch.id} className='fx-alert-check-inline'>
                  <input type='checkbox' checked={form.channel_ids.includes(String(ch.id))} onChange={() => toggleChannel(String(ch.id))} />
                  {ch.name || ch.id}
                </label>
              ))}
              {channels.length === 0 && <span style={{ color: '#64748b', fontSize: 12 }}>暂无可用渠道</span>}
            </div>
          </div>
          <div className='is-wide'>
            <span style={{ fontSize: 13, color: '#475569', marginBottom: 6, display: 'block' }}>标签匹配（key=value）</span>
            {form.labels.map((lbl, idx) => (
              <div key={idx} style={{ display: 'flex', gap: 6, marginBottom: 4 }}>
                <input value={lbl.key} onChange={(e) => updateLabel(idx, 'key', e.target.value)} placeholder='key' style={{ flex: 1 }} />
                <input value={lbl.value} onChange={(e) => updateLabel(idx, 'value', e.target.value)} placeholder='value' style={{ flex: 1 }} />
                <button type='button' onClick={() => removeLabel(idx)} disabled={form.labels.length <= 1}>-</button>
              </div>
            ))}
            <button type='button' onClick={addLabel}>+ 添加标签</button>
          </div>
          <label className='fx-alert-check'>
            <span>启用</span>
            <input type='checkbox' checked={form.enabled} onChange={(e) => setForm({ ...form, enabled: e.target.checked })} />
          </label>
        </div>
        <div className='fx-alert-actions' style={{ marginTop: 12, display: 'flex', gap: 8 }}>
          <button type='button' className='is-primary' disabled={saving || !form.name.trim()} onClick={onSave}>{saving ? '保存中...' : '保存'}</button>
          <button type='button' onClick={onCancel}>取消</button>
        </div>
      </div>
    </div>
  )
}
export function AlertSubscribeSection() {
  const [subscribes, setSubscribes] = useState([])
  const [channels, setChannels] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [modal, setModal] = useState(null)
  const [form, setForm] = useState(emptyForm())
  const [saving, setSaving] = useState(false)
  const { confirm, modal: confirmModal } = useConfirm()

  const loadSubscribes = async () => {
    setLoading(true); setError('')
    try {
      const data = await get('/alert-subscribes')
      setSubscribes(normalizeList(data))
      setPage(1)
    } catch (err) {
      setError(makeError(err, '加载订阅规则失败'))
    } finally {
      setLoading(false)
    }
  }

  const loadChannels = async () => {
    try {
      const data = await get('/notification-channels')
      setChannels(normalizeList(data))
    } catch (_) { /* 渠道加载失败不阻塞主流程 */ }
  }

  useEffect(() => { loadSubscribes(); loadChannels() }, [])

  const filtered = useMemo(() => subscribes.filter((s) =>
    filterText([s.name, ...mapToPairs(s.labels)], keyword)
  ), [subscribes, keyword])

  const total = filtered.length
  const paged = useMemo(() => filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE), [filtered, page])
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const openCreate = () => { setForm(emptyForm()); setModal('edit') }
  const openEdit = (item) => {
    setForm({
      id: item.id || '',
      name: item.name || '',
      labels: mapToLabels(item.labels),
      severities: Array.isArray(item.severities) ? item.severities : [],
      rule_ids: Array.isArray(item.rule_ids) ? item.rule_ids.join(',') : (item.rule_ids || ''),
      channel_ids: Array.isArray(item.channel_ids) ? item.channel_ids.map(String) : [],
      enabled: item.enabled !== false,
    })
    setModal('edit')
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = {
        name: form.name.trim(),
        labels: labelsToMap(form.labels),
        severities: form.severities,
        rule_ids: form.rule_ids ? form.rule_ids.split(',').map((s) => s.trim()).filter(Boolean) : [],
        channel_ids: form.channel_ids,
        enabled: form.enabled,
      }
      if (form.id) {
        await put(`/alert-subscribes/${encodeURIComponent(form.id)}`, payload)
      } else {
        await post('/alert-subscribes', payload)
      }
      setModal(null)
      await loadSubscribes()
    } catch (err) {
      setError(makeError(err, '保存失败'))
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (item) => {
    const ok = await confirm({ title: '删除订阅规则', message: `确认删除「${item.name}」？`, confirmText: '删除', danger: true })
    if (!ok) return
    try {
      await del(`/alert-subscribes/${encodeURIComponent(item.id)}`)
      await loadSubscribes()
    } catch (err) {
      setError(makeError(err, '删除失败'))
    }
  }

  const handleToggle = async (item) => {
    try {
      await put(`/alert-subscribes/${encodeURIComponent(item.id)}`, { ...item, enabled: !item.enabled })
      await loadSubscribes()
    } catch (err) {
      setError(makeError(err, '切换状态失败'))
    }
  }

  const channelName = (id) => {
    const ch = channels.find((c) => String(c.id) === String(id))
    return ch ? (ch.name || ch.id) : id
  }

  return (
    <section className='fx-alert-section'>
      <div className='fx-alert-filterbar'>
        <button type='button' disabled={loading} onClick={loadSubscribes}>{loading ? '刷新中...' : '刷新'}</button>
        <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder='搜索名称、标签' />
        <button type='button' className='is-primary' onClick={openCreate}>新建订阅</button>
      </div>
      {error && <div className='fx-alert-message is-error'>{error}</div>}
      <div className='fx-alert-table'>
        <table>
          <thead>
            <tr>
              <th>名称</th>
              <th>标签</th>
              <th>级别</th>
              <th>规则 ID</th>
              <th>通知渠道</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {paged.map((item) => (
              <tr key={item.id}>
                <td>{item.name}</td>
                <td>{mapToPairs(item.labels).map((t) => <span className='fx-alert-tag' key={t}>{t}</span>)}</td>
                <td>{(item.severities || []).map((s) => <span className={`fx-alert-severity is-${s}`} key={s}>{s}</span>)}</td>
                <td>{(Array.isArray(item.rule_ids) ? item.rule_ids : []).join(', ') || '-'}</td>
                <td>{(Array.isArray(item.channel_ids) ? item.channel_ids : []).map((id) => <span className='fx-alert-tag' key={id}>{channelName(id)}</span>)}</td>
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
        {!loading && paged.length === 0 && <div className='fx-alert-empty'>暂无订阅规则</div>}
      </div>
      {total > PAGE_SIZE && (
        <div className='fx-alert-pagination'>
          <span>共 {total} 条，第 {page}/{totalPages} 页</span>
          <button type='button' disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</button>
          <button type='button' disabled={page >= totalPages} onClick={() => setPage(page + 1)}>下一页</button>
        </div>
      )}
      {modal === 'edit' && <SubscribeModal form={form} setForm={setForm} onSave={handleSave} onCancel={() => setModal(null)} saving={saving} channels={channels} />}
      {confirmModal}
    </section>
  )
}
