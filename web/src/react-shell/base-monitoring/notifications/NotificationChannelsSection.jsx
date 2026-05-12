import React, { useEffect, useMemo, useState } from 'react'
import { notificationsApi } from '../../api/notifications.js'
import { Modal } from './NotificationsPage.jsx'
import { channelDraftFromRaw, channelFormFields, channelPayloadFromDraft, channelTypeLabel, channelTypes, displayDate, displayJson, filterText, makeError, normalizeChannel, secretView, supportedChannelTypes } from './notificationModel.js'

function ChannelForm({ draft, setDraft, saving, onSubmit, onClose }) {
  const fields = channelFormFields[draft.type] || []
  const switchType = (type) => {
    const newFields = channelFormFields[type] || []
    const next = { id: draft.id, type, name: draft.name, enabled: draft.enabled }
    newFields.forEach((f) => { next[f.key] = f.type === 'checkbox' ? false : '' })
    setDraft(next)
  }
  return (
    <Modal title={draft.id ? '编辑媒介' : '新建媒介'} onClose={onClose}>
      <div className='fx-notify-form'>
        <label><span>名称</span><input value={draft.name} onChange={(e) => setDraft({ ...draft, name: e.target.value })} /></label>
        <label><span>类型</span>
          <select value={draft.type} onChange={(e) => switchType(e.target.value)}>
            {channelTypes.filter((t) => t.supported).map((t) => <option key={t.ident} value={t.ident}>{t.label}</option>)}
          </select>
        </label>
        {fields.map((f) => {
          if (f.type === 'checkbox') {
            return <label key={f.key} className='fx-notify-check'><input type='checkbox' checked={!!draft[f.key]} onChange={(e) => setDraft({ ...draft, [f.key]: e.target.checked })} />{f.label}</label>
          }
          if (f.type === 'select') {
            return <label key={f.key}><span>{f.label}</span><select value={draft[f.key] || ''} onChange={(e) => setDraft({ ...draft, [f.key]: e.target.value })}><option value=''>请选择</option>{(f.options || []).map((o) => <option key={o} value={o}>{o}</option>)}</select></label>
          }
          if (f.type === 'textarea') {
            return <label key={f.key} className='is-wide'><span>{f.label}</span><textarea rows={3} value={draft[f.key] || ''} placeholder={f.placeholder} onChange={(e) => setDraft({ ...draft, [f.key]: e.target.value })} /></label>
          }
          return <label key={f.key}><span>{f.label}{f.required && ' *'}</span><input type={f.type === 'password' ? 'password' : f.type === 'number' ? 'number' : 'text'} value={draft[f.key] || ''} placeholder={f.placeholder} onChange={(e) => setDraft({ ...draft, [f.key]: e.target.value })} /></label>
        })}
        <label className='fx-notify-check'><input type='checkbox' checked={draft.enabled} onChange={(e) => setDraft({ ...draft, enabled: e.target.checked })} />启用</label>
      </div>
      <div className='fx-notify-actions'>
        <button type='button' onClick={onClose}>取消</button>
        <button type='button' className='is-primary' disabled={saving} onClick={onSubmit}>{saving ? '保存中...' : '保存'}</button>
      </div>
    </Modal>
  )
}

function ImportModal({ onClose, onImport }) {
  const [text, setText] = useState('')
  const [preview, setPreview] = useState(null)
  const [error, setError] = useState('')
  const fileRef = React.useRef(null)

  const parse = (raw) => {
    setError('')
    try {
      const data = JSON.parse(raw)
      const items = Array.isArray(data) ? data : [data]
      if (items.length === 0) { setError('JSON 为空数组'); return }
      setPreview(items)
    } catch { setError('JSON 格式不合法') }
  }

  const handleFile = (e) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = (ev) => { const content = ev.target.result; setText(content); parse(content) }
    reader.readAsText(file)
  }

  return (
    <Modal title='导入通知媒介' onClose={onClose}>
      <div className='fx-notify-form'>
        <label className='is-wide'><span>粘贴 JSON 配置</span>
          <textarea rows={8} value={text} placeholder='[{"type":"email","name":"...","smtp_host":"..."}]' onChange={(e) => setText(e.target.value)} />
        </label>
        <label className='is-wide'>
          <span>或上传 JSON 文件</span>
          <input ref={fileRef} type='file' accept='.json' onChange={handleFile} />
        </label>
      </div>
      {error && <div className='fx-notify-message is-error'>{error}</div>}
      {preview && <div style={{ margin: '8px 0' }}><strong>预览 ({preview.length} 条)</strong><pre>{displayJson(preview)}</pre></div>}
      <div className='fx-notify-actions'>
        <button type='button' onClick={onClose}>取消</button>
        {!preview && <button type='button' className='is-primary' onClick={() => parse(text)}>解析预览</button>}
        {preview && <button type='button' className='is-primary' onClick={() => onImport(preview)}>确认导入</button>}
      </div>
    </Modal>
  )
}

function Pagination({ total, page, pageSize, onChange }) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  if (total <= pageSize) return null
  return (
    <div className='fx-notify-pagination'>
      <button type='button' disabled={page <= 1} onClick={() => onChange(page - 1)}>上一页</button>
      <span>{page} / {totalPages}（共 {total} 条）</span>
      <button type='button' disabled={page >= totalPages} onClick={() => onChange(page + 1)}>下一页</button>
    </div>
  )
}

export function NotificationChannelsSection({ initialType, onTypeChange }) {
  const [channels, setChannels] = useState([])
  const [selectedType, setSelectedType] = useState(supportedChannelTypes.has(initialType) ? initialType : 'dingtalk')
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [modal, setModal] = useState(null)
  const [draft, setDraft] = useState(null)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [showImport, setShowImport] = useState(false)
  const pageSize = 20

  const load = async () => {
    setLoading(true); setError('')
    try {
      setChannels((await notificationsApi.listChannels()).map(normalizeChannel).filter((item) => item.id))
    } catch (err) {
      setError(makeError(err, '媒介加载失败'))
    } finally { setLoading(false) }
  }

  useEffect(() => { load() }, [])

  const filtered = useMemo(() => channels.filter((item) => {
    if (selectedType && item.type !== selectedType) return false
    if (status && String(item.enabled) !== status) return false
    return filterText([item.name, item.type, item.receiver], keyword)
  }), [channels, selectedType, status, keyword])

  const paged = useMemo(() => {
    const start = (page - 1) * pageSize
    return filtered.slice(start, start + pageSize)
  }, [filtered, page])

  useEffect(() => { setPage(1) }, [selectedType, status, keyword])

  const chooseType = (type) => { setSelectedType(type); onTypeChange?.(type) }

  const submit = async () => {
    setSaving(true); setError('')
    try {
      await notificationsApi.saveChannel(channelPayloadFromDraft(draft))
      setDraft(null); await load()
    } catch (err) { setError(makeError(err, '媒介保存失败')) }
    finally { setSaving(false) }
  }

  const toggle = async (channel) => {
    try {
      await notificationsApi.saveChannel({ ...channel.raw, id: channel.id, type: channel.type, name: channel.name, enabled: !channel.enabled })
      await load()
    } catch (err) { setModal({ title: '启停失败', body: makeError(err) }) }
  }

  const remove = async () => {
    if (!deleteTarget) return
    try { await notificationsApi.deleteChannel(deleteTarget.id); setDeleteTarget(null); await load() }
    catch (err) { setModal({ title: '删除失败', body: makeError(err) }) }
  }

  const handleImport = async (items) => {
    try {
      for (const item of items) await notificationsApi.saveChannel(item)
      setShowImport(false); await load()
    } catch (err) { setError(makeError(err, '导入失败')) }
  }

  const sanitizedExport = () => {
    const data = filtered.map((item) => ({ id: item.id, type: item.type, name: item.name, receiver: item.receiver, enabled: item.enabled, endpoint: secretView(item.raw.endpoint || item.raw.webhook) }))
    setModal({ title: '脱敏导出预览', body: displayJson(data) })
  }

  return (
    <section className='fx-notify-section is-split'>
      <aside className='fx-notify-catalog'>
        {channelTypes.map((type) => (
          <button key={type.ident} type='button' className={selectedType === type.ident ? 'is-active' : ''} onClick={() => chooseType(type.ident)}>
            <span>{type.label}</span>
          </button>
        ))}
      </aside>
      <div className='fx-notify-main'>
        <div className='fx-notify-filterbar'>
          <button type='button' disabled={loading} onClick={load}>{loading ? '刷新中...' : '刷新'}</button>
          <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder='搜索名称、接收方' />
          <select value={status} onChange={(e) => setStatus(e.target.value)}><option value=''>全部状态</option><option value='true'>启用</option><option value='false'>停用</option></select>
          <button type='button' className='is-primary' onClick={() => setDraft(channelDraftFromRaw(null, selectedType))}>新建</button>
          <button type='button' onClick={() => setShowImport(true)}>导入</button>
          <button type='button' onClick={sanitizedExport}>导出</button>
        </div>
        {error && <div className='fx-notify-message is-error'>{error}</div>}
        <div className='fx-notify-table'>
          <table>
            <thead><tr><th>状态</th><th>名称</th><th>类型</th><th>接收方</th><th>Endpoint</th><th>更新人</th><th>更新时间</th><th>操作</th></tr></thead>
            <tbody>{paged.map((item) => (
              <tr key={item.id}>
                <td><button type='button' className={item.enabled ? 'fx-notify-state is-on' : 'fx-notify-state'} onClick={() => toggle(item)}>{item.enabled ? '启用' : '停用'}</button></td>
                <td>{item.name}<small>{item.id}</small></td><td>{channelTypeLabel(item.type)}</td><td>{item.receiver || '-'}</td><td>{item.endpoint || '<SECRET>'}</td><td>{item.updatedBy}</td><td>{displayDate(item.updatedAt)}</td>
                <td><select onChange={(e) => { const v = e.target.value; e.target.value = ''; if (v === 'edit') setDraft(channelDraftFromRaw(item, item.type)); if (v === 'delete') setDeleteTarget(item) }}><option value=''>更多</option><option value='edit'>编辑</option><option value='delete'>删除</option></select></td>
              </tr>
            ))}</tbody>
          </table>
          {!loading && filtered.length === 0 && <div className='fx-notify-empty'>暂无媒介</div>}
        </div>
        <Pagination total={filtered.length} page={page} pageSize={pageSize} onChange={setPage} />
      </div>
      {draft && <ChannelForm draft={draft} setDraft={setDraft} saving={saving} onSubmit={submit} onClose={() => setDraft(null)} />}
      {showImport && <ImportModal onClose={() => setShowImport(false)} onImport={handleImport} />}
      {deleteTarget && <Modal title='确认删除媒介' narrow onClose={() => setDeleteTarget(null)}><p>确认删除 {deleteTarget.name}？</p><div className='fx-notify-actions'><button type='button' onClick={() => setDeleteTarget(null)}>取消</button><button type='button' className='is-danger' onClick={remove}>确认删除</button></div></Modal>}
      {modal && <Modal title={modal.title} onClose={() => setModal(null)}><pre>{modal.body}</pre></Modal>}
    </section>
  )
}
