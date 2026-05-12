import React, { useEffect, useState } from 'react'
import { formatOrgError, orgApi, ORG_BLOCKERS } from '../api/org.js'
import { useConfirm } from '../shared/ConfirmModal.jsx'
import { blankBusiness, useSectionLoader, Blocked, SimpleModal } from './OrgShared.jsx'

export function BusinessSection({ q }) {
  const { rows, loading, error, source, reload } = useSectionLoader('business', q)
  const [selected, setSelected] = useState(null)
  const [modal, setModal] = useState(null)
  const [teamID, setTeamID] = useState('')
  const [actionError, setActionError] = useState('')
  const [actionFeedback, setActionFeedback] = useState('')
  const { confirm, modal: confirmModal } = useConfirm()
  const detail = selected || rows[0] || null

  useEffect(() => { setSelected(rows[0] || null) }, [rows])

  const save = async (draft) => {
    setActionError('')
    setActionFeedback('')
    try {
      if (modal?.mode === 'edit') {
        await orgApi.business.update(modal.item.id, draft)
        await reload()
        setActionFeedback('业务组已更新。')
      } else {
        await orgApi.business.create(draft)
        await reload()
        setActionFeedback('业务组已创建。')
      }
      setModal(null)
    } catch (err) {
      setActionError(formatOrgError(err))
      throw err
    }
  }

  const remove = async (row) => {
    const ok = await confirm({ title: '删除业务组', message: `确认删除业务组 ${row.name}？`, confirmText: '删除', danger: true })
    if (!ok) return
    setActionError('')
    setActionFeedback('')
    try {
      await orgApi.business.remove(row.id)
      await reload()
      setActionFeedback('业务组已删除。')
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  const addTeam = async () => {
    if (!detail || !teamID.trim()) return
    setActionError('')
    setActionFeedback('')
    try {
      await orgApi.business.addTeams(detail.id, { team_id: teamID.trim(), perm_flag: 'rw' })
      setTeamID('')
      const next = await orgApi.business.get(detail.id)
      setSelected(next)
      await reload()
      setActionFeedback('授权团队已添加。')
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  const removeTeam = async (team) => {
    if (!detail) return
    setActionError('')
    setActionFeedback('')
    try {
      await orgApi.business.removeTeam(detail.id, team.id)
      const next = await orgApi.business.get(detail.id)
      setSelected(next)
      await reload()
      setActionFeedback('授权团队已移除。')
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  return (
    <section className='fx-org-split'>
      <div className='fx-org-list'>
        <div className='fx-org-toolbar'><button type='button' onClick={() => setModal({ mode: 'create', item: blankBusiness })}>新增业务组</button><button type='button' onClick={reload}>刷新</button><span>{loading ? '加载中...' : source}</span></div>
        {error && <Blocked>{error}</Blocked>}
        {actionError && <Blocked>{actionError}</Blocked>}
        {actionFeedback && <div className='fx-org-feedback'>{actionFeedback}</div>}
        {rows.map((row) => <button key={row.id} type='button' className={detail?.id === row.id ? 'fx-org-row is-active' : 'fx-org-row'} onClick={() => setSelected(row)}><strong>{row.name}</strong><span>{row.note || '无备注'}</span></button>)}
        {!rows.length && !loading && <div className='fx-org-empty'>暂无业务组。</div>}
      </div>
      <aside className='fx-org-detail'>
        {detail ? (
          <>
            <header><h3>{detail.name}</h3><p>{detail.note || '无备注'}</p></header>
            <div className='fx-org-actions'><button type='button' onClick={() => setModal({ mode: 'edit', item: detail })}>编辑</button><button type='button' className='danger' onClick={() => remove(detail)}>删除</button></div>
            <div className='fx-org-member-box'>
              <h4>授权团队</h4>
              <div className='fx-org-inline'><input value={teamID} onChange={(e) => setTeamID(e.target.value)} placeholder='输入团队 ID' /><button type='button' onClick={addTeam}>添加团队</button></div>
              {(detail.teams || []).map((team) => <span key={team.id} className='fx-org-chip'>{team.name || team.id}<small>{team.perm_flag}</small><button type='button' onClick={() => removeTeam(team)}>移除</button></span>)}
              {!(detail.teams || []).length && <p>暂无授权团队。</p>}
            </div>
          </>
        ) : <Blocked>{ORG_BLOCKERS.members}</Blocked>}
      </aside>
      {modal && <SimpleModal title={modal.mode === 'edit' ? '编辑业务组' : '新增业务组'} fields={['name', 'note', 'parent_id']} labels={{ name: '名称', note: '备注', parent_id: '父级 ID' }} initial={modal.item} onClose={() => setModal(null)} onSave={save} />}
      {confirmModal}
    </section>
  )
}
