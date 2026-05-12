import React, { useMemo, useState } from 'react'
import { assetsApi, formatAssetError, splitText } from '../api/assets.js'
import { displayTags, displayText, fmtTime } from './assetsModel.js'
import { ErrorBox, Feedback, Field, filterRows, Modal, Status, statusLabel, Tags } from './Shared.jsx'
import { useConfirm } from '../shared/ConfirmModal.jsx'

const blankGroup = { name: '', description: '', workspace_id: '', parent_id: '', status: 'active', tagsText: '' }

export function ResourceGroupsSection({ rows, workspaces, error, q, onRefresh }) {
  const [selected, setSelected] = useState('')
  const [modal, setModal] = useState(null)
  const [feedback, setFeedback] = useState('')
  const [actionError, setActionError] = useState('')
  const { confirm, modal: confirmModal } = useConfirm()
  const filtered = useMemo(() => filterRows(rows, q), [rows, q])
  const save = async draft => {
    setActionError(''); setFeedback('')
    const payload = { name: draft.name.trim(), description: draft.description.trim(), workspace_id: draft.workspace_id.trim(), parent_id: draft.parent_id.trim(), status: draft.status || 'active', tags: splitText(draft.tagsText) }
    try {
      if (modal.mode === 'edit') await assetsApi.resourceGroups.update(modal.item.id, payload)
      else await assetsApi.resourceGroups.create(payload)
      setFeedback('资源组已保存。'); setModal(null); onRefresh()
    } catch (err) { setActionError(formatAssetError(err)); throw err }
  }
  const remove = async row => {
    const ok = await confirm({ title: '删除资源组', message: `确认删除资源组「${displayText(row.name || row.id)}」？`, confirmText: '删除', danger: true })
    if (!ok) return
    try { await assetsApi.resourceGroups.remove(row.id); setFeedback('资源组已删除。'); onRefresh() } catch (err) { setActionError(formatAssetError(err)) }
  }
  return (
    <section className='fx-assets-split'>
      <GroupTree rows={filtered} selected={selected} onSelect={setSelected} onCreateRoot={() => setModal({ mode: 'create', item: blankGroup })} onCreateChild={row => setModal({ mode: 'create', item: { ...blankGroup, parent_id: row.id } })} onEdit={row => setModal({ mode: 'edit', item: groupDraft(row) })} onDelete={remove} />
      <div className='fx-assets-detail'>
        <div className='fx-assets-toolbar'><button type='button' onClick={() => setModal({ mode: 'create', item: blankGroup })}>新建资源组</button><button type='button' onClick={onRefresh}>刷新</button></div>
        <ErrorBox>{error || actionError}</ErrorBox><Feedback>{feedback}</Feedback>
        <ResourceGroupTable rows={filtered} workspaces={workspaces} onEdit={row => setModal({ mode: 'edit', item: groupDraft(row) })} onDelete={remove} />
      </div>
      {modal && <GroupModal modal={modal} groups={rows} workspaces={workspaces} onClose={() => setModal(null)} onSave={save} />}
      {confirmModal}
    </section>
  )
}

export function GroupTree({ rows, selected, onSelect, onCreateRoot, onCreateChild, onEdit, onDelete, readonly = false }) {
  return (
    <aside className='fx-assets-tree'>
      <div className='fx-assets-toolbar'><button type='button' disabled={readonly} onClick={onCreateRoot}>创建根资源组</button></div>
      {rows.map(row => (
        <button type='button' key={row.id} className={`fx-assets-tree-row ${selected === row.id ? 'is-active' : ''}`} onClick={() => onSelect(row.id)}>
          <strong>{displayText(row.name)}</strong><span>{row.parent_id ? `父级 ${displayText(row.parent_id)}` : '根资源组'} · {statusLabel(row.status)}</span>
          {!readonly && <div className='fx-assets-actions'><button type='button' onClick={e => { e.stopPropagation(); onCreateChild(row) }}>子组</button><button type='button' onClick={e => { e.stopPropagation(); onEdit(row) }}>编辑</button><button type='button' className='is-danger' onClick={e => { e.stopPropagation(); onDelete(row) }}>删除</button></div>}
        </button>
      ))}
      {!rows.length && <div className='fx-assets-empty'>暂无资源组。</div>}
    </aside>
  )
}

function ResourceGroupTable({ rows, workspaces, onEdit, onDelete }) {
  const workspaceName = id => displayText(workspaces.find(item => item.id === id)?.name || id)
  return <div className='fx-assets-table'><table><thead><tr><th>名称</th><th>业务组</th><th>父级</th><th>状态</th><th>标签</th><th>更新时间</th><th>操作</th></tr></thead><tbody>{rows.map(row => <tr key={row.id}><td><strong>{displayText(row.name)}</strong><div className='fx-assets-muted'>{displayText(row.description)}</div></td><td>{workspaceName(row.workspace_id)}</td><td>{displayText(row.parent_id)}</td><td><Status ok={row.status === 'active'}>{statusLabel(row.status)}</Status></td><td><Tags items={row.tags} /></td><td>{fmtTime(row.updated_at)}</td><td className='fx-assets-actions'><button type='button' onClick={() => onEdit(row)}>编辑</button><button type='button' className='is-danger' onClick={() => onDelete(row)}>删除</button></td></tr>)}</tbody></table>{!rows.length && <div className='fx-assets-empty'>暂无资源组。</div>}</div>
}

function GroupModal({ modal, groups, workspaces, onClose, onSave }) {
  const [draft, setDraft] = useState(modal.item)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const patch = (key, value) => setDraft(prev => ({ ...prev, [key]: value }))
  const submit = async () => {
    if (!draft.name.trim()) return setError('资源组名称不能为空。')
    setSaving(true); setError('')
    try { await onSave(draft) } catch (err) { setError(formatAssetError(err)) } finally { setSaving(false) }
  }
  return <Modal title={modal.mode === 'edit' ? '编辑资源组' : '新建资源组'} onClose={onClose}><div className='fx-assets-form'><Field label='名称 *'><input value={draft.name} onChange={e => patch('name', e.target.value)} /></Field><Field label='描述'><textarea rows='3' value={draft.description} onChange={e => patch('description', e.target.value)} /></Field><div className='fx-assets-form-grid'><Field label='业务组'><select value={draft.workspace_id} onChange={e => patch('workspace_id', e.target.value)}><option value=''>未绑定</option>{workspaces.map(row => <option key={row.id} value={row.id}>{displayText(row.name)}</option>)}</select></Field><Field label='父级资源组'><select value={draft.parent_id} onChange={e => patch('parent_id', e.target.value)}><option value=''>根资源组</option>{groups.filter(row => row.id !== draft.id).map(row => <option key={row.id} value={row.id}>{displayText(row.name)}</option>)}</select></Field></div><Field label='状态'><select value={draft.status} onChange={e => patch('status', e.target.value)}><option value='active'>启用</option><option value='disabled'>停用</option><option value='archived'>归档</option></select></Field><Field label='标签'><input value={draft.tagsText} onChange={e => patch('tagsText', e.target.value)} /></Field><ErrorBox>{error}</ErrorBox><footer><button type='button' disabled={saving} onClick={submit}>{saving ? '保存中...' : '保存'}</button></footer></div></Modal>
}

function groupDraft(row) {
  return { id: row.id, name: displayText(row.name, ''), description: displayText(row.description, ''), workspace_id: row.workspace_id || '', parent_id: row.parent_id || '', status: row.status || 'active', tagsText: displayTags(row.tags).join(', ') }
}
