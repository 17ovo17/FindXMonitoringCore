import React, { useEffect, useMemo, useState } from 'react'
import { notificationsApi } from '../../api/notifications.js'
import { Modal } from './NotificationsPage.jsx'
import { channelTypes, displayDate, displayJson, filterText, makeError, normalizeTemplate, parseJson } from './notificationModel.js'

const templateVariables = [
  { name: '$event.name', desc: '告警名称' },
  { name: '$event.severity', desc: '告警级别' },
  { name: '$event.status', desc: '告警状态 (firing/resolved)' },
  { name: '$event.target_ident', desc: '监控对象标识' },
  { name: '$event.value', desc: '触发值' },
  { name: '$event.labels', desc: '标签 map' },
  { name: '$event.labels.<key>', desc: '指定标签值' },
  { name: '$event.annotations', desc: '注解 map' },
  { name: '$event.annotations.summary', desc: '摘要注解' },
  { name: '$event.annotations.description', desc: '描述注解' },
  { name: '$event.trigger_time', desc: '触发时间' },
  { name: '$event.recover_time', desc: '恢复时间' },
  { name: '$event.rule_name', desc: '规则名称' },
  { name: '$event.rule_id', desc: '规则 ID' },
  { name: '$event.group_name', desc: '业务组名称' },
  { name: '{{range .Events}}', desc: '遍历事件列表' },
  { name: '{{end}}', desc: '结束遍历' },
  { name: '{{if eq .Status "firing"}}', desc: '条件判断' },
]

const emptyDraft = { id: '', name: '', ident: '', notifyChannelIdent: 'dingtalk', private: 0, content: '{\n  "title": "告警: {{.Name}}",\n  "content": "级别: {{.Severity}}\\n状态: {{.Status}}\\n对象: {{.TargetIdent}}"\n}' }

function toDraft(template) {
  return template ? {
    id: template.id,
    name: template.name,
    ident: template.ident,
    notifyChannelIdent: template.notifyChannelIdent,
    private: template.private,
    content: displayJson(template.content || {}),
  } : emptyDraft
}

function payloadFromDraft(draft) {
  return {
    id: draft.id || undefined,
    name: draft.name.trim(),
    ident: draft.ident.trim(),
    notify_channel_ident: draft.notifyChannelIdent,
    private: Number(draft.private || 0),
    content: parseJson(draft.content, {}),
  }
}

function MonacoTemplateEditor({ value, onChange }) {
  const [Editor, setEditor] = useState(null)
  useEffect(() => { import('@monaco-editor/react').then((mod) => setEditor(() => mod.default)) }, [])
  if (!Editor) return <textarea rows={12} value={value} onChange={(e) => onChange(e.target.value)} style={{ width: '100%', fontFamily: 'monospace' }} />
  return <Editor height='300px' language='handlebars' theme='vs-dark' value={value} onChange={(val) => onChange(val || '')} options={{ minimap: { enabled: false }, wordWrap: 'on', fontSize: 13, scrollBeyondLastLine: false }} />
}

function TemplateForm({ draft, setDraft, saving, onSave, onPreview, onClose }) {
  return (
    <Modal title={draft.id ? '编辑消息模板' : '新建消息模板'} onClose={onClose}>
      <div className='fx-notify-template-editor'>
        <div className='fx-notify-template-editor__left'>
          <div className='fx-notify-form'>
            <label><span>名称</span><input value={draft.name} onChange={(e) => setDraft({ ...draft, name: e.target.value })} /></label>
            <label><span>标识</span><input value={draft.ident} onChange={(e) => setDraft({ ...draft, ident: e.target.value })} /></label>
            <label><span>媒介</span><select value={draft.notifyChannelIdent} onChange={(e) => setDraft({ ...draft, notifyChannelIdent: e.target.value })}>{channelTypes.filter((t) => t.supported).map((t) => <option key={t.ident} value={t.ident}>{t.label}</option>)}</select></label>
            <label><span>私有</span><select value={draft.private} onChange={(e) => setDraft({ ...draft, private: Number(e.target.value) })}><option value={0}>否</option><option value={1}>是</option></select></label>
          </div>
          <div className='fx-notify-template-editor__monaco'>
            <label><span>模板内容</span></label>
            <MonacoTemplateEditor value={draft.content} onChange={(val) => setDraft({ ...draft, content: val })} />
          </div>
        </div>
        <aside className='fx-notify-template-editor__vars'>
          <strong>可用变量</strong>
          <ul>{templateVariables.map((v) => <li key={v.name}><code>{v.name}</code><span>{v.desc}</span></li>)}</ul>
        </aside>
      </div>
      <div className='fx-notify-actions'>
        <button type='button' onClick={onClose}>取消</button>
        <button type='button' onClick={onPreview}>预览</button>
        <button type='button' className='is-primary' disabled={saving} onClick={onSave}>{saving ? '保存中...' : '保存'}</button>
      </div>
    </Modal>
  )
}

export function NotificationTemplatesSection() {
  const [templates, setTemplates] = useState([])
  const [selectedId, setSelectedId] = useState('')
  const [keyword, setKeyword] = useState('')
  const [channelIdent, setChannelIdent] = useState('')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [draft, setDraft] = useState(null)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [modal, setModal] = useState(null)

  const load = async () => {
    setLoading(true); setError('')
    try {
      const rows = (await notificationsApi.listTemplates(channelIdent)).map(normalizeTemplate).filter((item) => item.id)
      setTemplates(rows)
      if (!selectedId && rows[0]) setSelectedId(rows[0].id)
    } catch (err) { setError(makeError(err, '模板加载失败')) }
    finally { setLoading(false) }
  }

  useEffect(() => { load() }, [channelIdent])

  const filtered = useMemo(() => templates.filter((item) => filterText([item.name, item.ident, item.notifyChannelIdent], keyword)), [templates, keyword])
  const selected = useMemo(() => templates.find((item) => item.id === selectedId) || filtered[0], [templates, filtered, selectedId])

  const save = async () => {
    setSaving(true); setError('')
    try {
      const body = payloadFromDraft(draft)
      if (draft.id) await notificationsApi.updateTemplate(draft.id, body)
      else await notificationsApi.saveTemplate(body)
      setDraft(null); await load()
    } catch (err) { setError(makeError(err, '模板保存失败')) }
    finally { setSaving(false) }
  }

  const preview = async (source) => {
    try {
      const body = source?.id ? { event_ids: [], event: sampleEvent() } : { tpl: { content: payloadFromDraft(draft).content }, event_ids: [], event: sampleEvent() }
      const result = source?.id ? await notificationsApi.previewTemplateById(source.id, body) : await notificationsApi.previewTemplate(body)
      setModal({ title: '模板预览', body: displayJson(result) })
    } catch (err) { setModal({ title: '预览失败', body: makeError(err) }) }
  }

  const clone = async (template) => {
    try { await notificationsApi.cloneTemplate(template.id); await load() }
    catch (err) { setModal({ title: '克隆失败', body: makeError(err) }) }
  }

  const remove = async () => {
    try { await notificationsApi.deleteTemplates([deleteTarget.id]); setDeleteTarget(null); await load() }
    catch (err) { setModal({ title: '删除失败', body: makeError(err) }) }
  }

  return (
    <section className='fx-notify-section is-split'>
      <aside className='fx-notify-list'>
        <div className='fx-notify-filterbar is-list'>
          <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder='搜索模板' />
          <select value={channelIdent} onChange={(e) => setChannelIdent(e.target.value)}><option value=''>全部媒介</option>{channelTypes.filter((t) => t.supported).map((t) => <option key={t.ident} value={t.ident}>{t.label}</option>)}</select>
          <button type='button' className='is-primary' onClick={() => setDraft(toDraft(null))}>新建</button>
        </div>
        {filtered.map((item) => <button key={item.id} type='button' className={selected?.id === item.id ? 'is-active' : ''} onClick={() => setSelectedId(item.id)}><strong>{item.name}</strong><span>{item.notifyChannelIdent} · {item.ident || item.id}</span></button>)}
        {!loading && filtered.length === 0 && <div className='fx-notify-empty'>暂无模板</div>}
      </aside>
      <div className='fx-notify-main'>
        <div className='fx-notify-filterbar'>
          <button type='button' disabled={loading} onClick={load}>{loading ? '刷新中...' : '刷新'}</button>
          <button type='button' disabled={!selected} onClick={() => setDraft(toDraft(selected))}>编辑</button>
          <button type='button' disabled={!selected} onClick={() => preview(selected)}>预览</button>
          <button type='button' disabled={!selected} onClick={() => clone(selected)}>克隆</button>
          <button type='button' disabled={!selected} className='is-danger' onClick={() => setDeleteTarget(selected)}>删除</button>
        </div>
        {error && <div className='fx-notify-message is-error'>{error}</div>}
        {selected ? <div className='fx-notify-detail'>
          <dl><dt>名称</dt><dd>{selected.name}</dd><dt>标识</dt><dd>{selected.ident || '-'}</dd><dt>媒介</dt><dd>{selected.notifyChannelIdent}</dd><dt>私有</dt><dd>{selected.private ? '是' : '否'}</dd><dt>更新人</dt><dd>{selected.updatedBy}</dd><dt>更新时间</dt><dd>{displayDate(selected.updatedAt)}</dd></dl>
          <pre>{displayJson(selected.content)}</pre>
        </div> : <div className='fx-notify-empty'>请选择模板</div>}
      </div>
      {draft && <TemplateForm draft={draft} setDraft={setDraft} saving={saving} onSave={save} onPreview={() => preview()} onClose={() => setDraft(null)} />}
      {deleteTarget && <Modal title='确认删除消息模板' narrow onClose={() => setDeleteTarget(null)}><p>确认删除 {deleteTarget.name}？</p><div className='fx-notify-actions'><button type='button' onClick={() => setDeleteTarget(null)}>取消</button><button type='button' className='is-danger' onClick={remove}>确认删除</button></div></Modal>}
      {modal && <Modal title={modal.title} onClose={() => setModal(null)}><pre>{modal.body}</pre></Modal>}
    </section>
  )
}

function sampleEvent() {
  return { name: 'FindX notification preview', severity: 'warning', status: 'firing', value: 'preview', target_ident: 'demo-node', labels: { service: 'findx' }, annotations: { summary: 'contract validation' } }
}
