import React, { useMemo, useState } from 'react'
import { assetsApi, formatAssetError, splitText } from '../api/assets.js'
import { displayTags, displayText, endpointText, fmtTime, normalizeTags, parseEndpoints } from './assetsModel.js'
import { ErrorBox, Feedback, Field, filterRows, Modal, Status, statusLabel, Tags } from './Shared.jsx'
import { useConfirm } from '../shared/ConfirmModal.jsx'

const blankBusiness = { name: '', description: '', owner: '', status: 'active', tagsText: '', hostsText: '', endpointsText: '' }

export function BusinessSection({ rows, error, q, onRefresh }) {
  const [modal, setModal] = useState(null)
  const [feedback, setFeedback] = useState('')
  const [actionError, setActionError] = useState('')
  const { confirm, modal: confirmModal } = useConfirm()
  const filtered = useMemo(() => filterRows(rows, q), [rows, q])

  const save = async draft => {
    setActionError('')
    setFeedback('')
    const payload = {
      name: draft.name.trim(),
      description: draft.description.trim(),
      owner: draft.owner.trim(),
      status: draft.status || 'active',
      tags: splitText(draft.tagsText),
      hosts: splitText(draft.hostsText),
      endpoints: parseEndpoints(draft.endpointsText),
    }
    try {
      if (modal.mode === 'edit') await assetsApi.workspaces.update(modal.item.id, payload)
      else await assetsApi.workspaces.create(payload)
      setFeedback('业务组已保存。')
      setModal(null)
      onRefresh()
    } catch (err) {
      setActionError(formatAssetError(err))
      throw err
    }
  }

  const remove = async row => {
    const ok = await confirm({ title: '删除业务组', message: `确认删除业务组「${displayText(row.name || row.id)}」？`, confirmText: '删除', danger: true })
    if (!ok) return
    setActionError('')
    setFeedback('')
    try {
      await assetsApi.workspaces.remove(row.id)
      setFeedback('业务组已删除。')
      onRefresh()
    } catch (err) {
      setActionError(formatAssetError(err))
    }
  }

  return (
    <section className='fx-assets-work'>
      <div className='fx-assets-toolbar'>
        <button type='button' onClick={() => setModal({ mode: 'create', item: blankBusiness })}>新建业务组</button>
        <button type='button' onClick={onRefresh}>刷新</button>
      </div>
      <ErrorBox>{error || actionError}</ErrorBox><Feedback>{feedback}</Feedback>
      <div className='fx-assets-table'>
        <table><thead><tr><th>名称</th><th>负责人</th><th>状态</th><th>资源</th><th>标签</th><th>更新时间</th><th>操作</th></tr></thead>
          <tbody>{filtered.map(row => (
            <tr key={row.id}>
              <td><strong>{displayText(row.name)}</strong><div className='fx-assets-muted'>{displayText(row.description)}</div></td>
              <td>{displayText(row.owner)}</td><td><Status ok={row.status === 'active'}>{statusLabel(row.status)}</Status></td>
              <td>{row.resource_count ?? ((row.hosts || []).length + (row.endpoints || []).length)} 项</td>
              <td><Tags items={row.tags} /></td><td>{fmtTime(row.updated_at)}</td>
              <td className='fx-assets-actions'><button type='button' onClick={() => setModal({ mode: 'edit', item: businessDraft(row) })}>编辑</button><button type='button' className='is-danger' onClick={() => remove(row)}>删除</button></td>
            </tr>
          ))}</tbody>
        </table>
        {!filtered.length && <div className='fx-assets-empty'>暂无业务组。</div>}
      </div>
      {modal && <BusinessModal modal={modal} onClose={() => setModal(null)} onSave={save} />}
      {confirmModal}
    </section>
  )
}

function BusinessModal({ modal, onClose, onSave }) {
  const [draft, setDraft] = useState(modal.item)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const patch = (key, value) => setDraft(prev => ({ ...prev, [key]: value }))
  const submit = async () => {
    if (!draft.name.trim()) return setError('业务组名称不能为空。')
    setSaving(true); setError('')
    try { await onSave(draft) } catch (err) { setError(formatAssetError(err)) } finally { setSaving(false) }
  }
  return (
    <Modal title={modal.mode === 'edit' ? '编辑业务组' : '新建业务组'} onClose={onClose}>
      <div className='fx-assets-form'>
        <Field label='名称 *'><input value={draft.name} onChange={e => patch('name', e.target.value)} /></Field>
        <Field label='描述'><textarea rows='3' value={draft.description} onChange={e => patch('description', e.target.value)} /></Field>
        <div className='fx-assets-form-grid'><Field label='负责人'><input value={draft.owner} onChange={e => patch('owner', e.target.value)} /></Field><Field label='状态'><select value={draft.status} onChange={e => patch('status', e.target.value)}><option value='active'>启用</option><option value='disabled'>停用</option><option value='archived'>归档</option></select></Field></div>
        <Field label='标签'><input value={draft.tagsText} onChange={e => patch('tagsText', e.target.value)} placeholder='逗号或换行分隔' /></Field>
        <Field label='主机'><textarea rows='3' value={draft.hostsText} onChange={e => patch('hostsText', e.target.value)} placeholder='每行一个主机名或 IP' /></Field>
        <Field label='端点'><textarea rows='3' value={draft.endpointsText} onChange={e => patch('endpointsText', e.target.value)} placeholder='10.0.0.1:8080 Web HTTP' /></Field>
        <ErrorBox>{error}</ErrorBox><footer><button type='button' disabled={saving} onClick={submit}>{saving ? '保存中...' : '保存'}</button></footer>
      </div>
    </Modal>
  )
}

function businessDraft(row) {
  return { name: displayText(row.name, ''), description: displayText(row.description, ''), owner: displayText(row.owner, ''), status: row.status || 'active', tagsText: displayTags(row.tags).join(', '), hostsText: normalizeTags(row.hosts).map(item => displayText(item, '')).join('\n'), endpointsText: (row.endpoints || []).map(endpointText).join('\n') }
}
