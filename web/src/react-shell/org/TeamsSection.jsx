import React, { useEffect, useMemo, useState } from 'react'
import { formatOrgError, orgApi, ORG_BLOCKERS } from '../api/org.js'
import { useConfirm } from '../shared/ConfirmModal.jsx'
import { blankTeam, buildTree, useSectionLoader, Blocked, SimpleModal, TreeNode } from './OrgShared.jsx'

export function TeamsSection({ q }) {
  const { rows, loading, error, source, reload } = useSectionLoader('teams', q)
  const [selected, setSelected] = useState(null)
  const [modal, setModal] = useState(null)
  const [memberText, setMemberText] = useState('')
  const { confirm, modal: confirmModal } = useConfirm()
  const [actionError, setActionError] = useState('')
  const [actionFeedback, setActionFeedback] = useState('')
  const detail = selected || rows[0] || null
  const tree = useMemo(() => buildTree(rows), [rows])

  useEffect(() => { setSelected(rows[0] || null) }, [rows])

  const save = async (draft) => {
    setActionError('')
    setActionFeedback('')
    try {
      if (modal?.mode === 'edit') {
        await orgApi.teams.update(modal.item.id, draft)
        await reload()
        setActionFeedback('团队已更新。')
      } else {
        await orgApi.teams.create(draft)
        await reload()
        setActionFeedback('团队已创建。')
      }
      setModal(null)
    } catch (err) {
      setActionError(formatOrgError(err))
      throw err
    }
  }

  const remove = async (row) => {
    const ok = await confirm({ title: '删除团队', message: `确认删除团队 ${row.name}？`, confirmText: '删除', danger: true })
    if (!ok) return
    setActionError('')
    setActionFeedback('')
    try {
      await orgApi.teams.remove(row.id)
      await reload()
      setActionFeedback('团队已删除。')
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  const addMembers = async () => {
    const ids = memberText.split(',').map((item) => item.trim()).filter(Boolean)
    if (!detail || !ids.length) return
    setActionError('')
    setActionFeedback('')
    try {
      await orgApi.teams.addMembers(detail.id, { ids })
      setMemberText('')
      const next = await orgApi.teams.get(detail.id)
      setSelected(next)
      await reload()
      setActionFeedback('成员已添加。')
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  const removeMember = async (user) => {
    if (!detail) return
    setActionError('')
    setActionFeedback('')
    try {
      await orgApi.teams.removeMember(detail.id, user.id)
      const next = await orgApi.teams.get(detail.id)
      setSelected(next)
      await reload()
      setActionFeedback('成员已移除。')
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  return (
    <section className='fx-org-split'>
      <div className='fx-org-list'>
        <div className='fx-org-toolbar'><button type='button' onClick={() => setModal({ mode: 'create', item: blankTeam })}>新增团队</button><button type='button' onClick={reload}>刷新</button><span>{loading ? '加载中...' : source}</span></div>
        {error && <Blocked>{error}</Blocked>}
        {actionError && <Blocked>{actionError}</Blocked>}
        {actionFeedback && <div className='fx-org-feedback'>{actionFeedback}</div>}
        {tree.map((node) => <TreeNode key={node.id} node={node} depth={0} selected={detail} onSelect={setSelected} />)}
        {!rows.length && !loading && <div className='fx-org-empty'>暂无团队。</div>}
      </div>
      <aside className='fx-org-detail'>
        {detail ? (
          <>
            <header><h3>{detail.name}</h3><p>{detail.note || '无备注'}{detail.parent_id ? ` | 父级: ${detail.parent_id}` : ''}</p></header>
            <div className='fx-org-actions'><button type='button' onClick={() => setModal({ mode: 'edit', item: detail })}>编辑</button><button type='button' className='danger' onClick={() => remove(detail)}>删除</button></div>
            <div className='fx-org-member-box'>
              <h4>成员（{(detail.members || []).length}）</h4>
              <div className='fx-org-inline'><input value={memberText} onChange={(e) => setMemberText(e.target.value)} placeholder='输入用户 ID，多个用逗号分隔' /><button type='button' onClick={addMembers}>添加成员</button></div>
              {(detail.members || []).map((user) => <span key={user.id} className='fx-org-chip'>{user.username}<button type='button' onClick={() => removeMember(user)}>移除</button></span>)}
              {!(detail.members || []).length && <p>暂无成员。</p>}
            </div>
          </>
        ) : <Blocked>{ORG_BLOCKERS.members}</Blocked>}
      </aside>
      {modal && <SimpleModal title={modal.mode === 'edit' ? '编辑团队' : '新增团队'} fields={['name', 'note', 'parent_id']} labels={{ name: '名称', note: '备注', parent_id: '父级 ID' }} initial={modal.item} onClose={() => setModal(null)} onSave={save} />}
      {confirmModal}
    </section>
  )
}
