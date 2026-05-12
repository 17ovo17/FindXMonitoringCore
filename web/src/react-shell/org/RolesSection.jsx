import React, { useEffect, useState } from 'react'
import { formatOrgError, orgApi } from '../api/org.js'
import { useConfirm } from '../shared/ConfirmModal.jsx'
import { blankRole, useSectionLoader, Blocked, SimpleModal } from './OrgShared.jsx'

export function RolesSection({ q }) {
  const { rows, loading, error, source, reload } = useSectionLoader('roles', q)
  const [operations, setOperations] = useState([])
  const [selected, setSelected] = useState(null)
  const [checked, setChecked] = useState(new Set())
  const [modal, setModal] = useState(null)
  const [actionError, setActionError] = useState('')
  const [actionFeedback, setActionFeedback] = useState('')
  const { confirm, modal: confirmModal } = useConfirm()
  const detail = selected || rows[0] || null

  useEffect(() => {
    orgApi.roles.operations().then((res) => setOperations(res.rows || [])).catch(() => setOperations([]))
  }, [])
  useEffect(() => { setSelected(rows[0] || null) }, [rows])
  useEffect(() => {
    if (!detail?.id) return
    orgApi.roles.roleOperations(detail.id).then((ops) => setChecked(new Set(Array.isArray(ops) ? ops : []))).catch(() => setChecked(new Set()))
  }, [detail?.id])

  const toggle = (op) => setChecked((prev) => {
    const next = new Set(prev)
    if (next.has(op)) next.delete(op)
    else next.add(op)
    return next
  })

  const saveOps = async () => {
    if (!detail || detail.builtin) return
    setActionError('')
    setActionFeedback('')
    try {
      await orgApi.roles.saveOperations(detail.id, Array.from(checked))
      setActionFeedback('角色权限已保存。')
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  const saveRole = async (draft) => {
    setActionError('')
    setActionFeedback('')
    try {
      if (modal?.mode === 'edit') {
        await orgApi.roles.update(modal.item.id, draft)
        await reload()
        setActionFeedback('角色已更新。')
      } else {
        await orgApi.roles.create(draft)
        await reload()
        setActionFeedback('角色已创建。')
      }
      setModal(null)
    } catch (err) {
      setActionError(formatOrgError(err))
      throw err
    }
  }

  const removeRole = async (role) => {
    if (role.builtin) return
    const ok = await confirm({ title: '删除角色', message: `确认删除角色 ${role.name}？`, confirmText: '删除', danger: true })
    if (!ok) return
    setActionError('')
    setActionFeedback('')
    try {
      await orgApi.roles.remove(role.id)
      await reload()
      setActionFeedback('角色已删除。')
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  return (
    <section className='fx-org-split'>
      <div className='fx-org-list'>
        <div className='fx-org-toolbar'><button type='button' onClick={() => setModal({ mode: 'create', item: blankRole })}>新增角色</button><button type='button' onClick={reload}>刷新</button><span>{loading ? '加载中...' : source}</span></div>
        {error && <Blocked>{error}</Blocked>}
        {actionError && <Blocked>{actionError}</Blocked>}
        {actionFeedback && <div className='fx-org-feedback'>{actionFeedback}</div>}
        {rows.map((row) => <button key={row.id} type='button' className={detail?.id === row.id ? 'fx-org-row is-active' : 'fx-org-row'} onClick={() => setSelected(row)}><strong>{row.name}</strong><span>{row.builtin ? '内置角色' : row.note || '自定义角色'}</span></button>)}
      </div>
      <aside className='fx-org-detail is-wide'>
        {detail ? (
          <>
            <header><h3>{detail.name}</h3><p>{detail.builtin ? '内置角色受保护，不允许编辑权限或删除。' : detail.note || '无备注'}</p></header>
            <div className='fx-org-actions'><button type='button' disabled={detail.builtin} onClick={() => setModal({ mode: 'edit', item: detail })}>编辑</button><button type='button' disabled={detail.builtin} className='danger' onClick={() => removeRole(detail)}>删除</button><button type='button' disabled={detail.builtin} onClick={saveOps}>保存权限</button></div>
            {detail.builtin && <Blocked>内置角色权限由平台保护，不允许从页面修改。</Blocked>}
            <div className='fx-org-ops'>
              {operations.map((group) => (
                <section key={group.name}>
                  <h4>{group.cname || group.name}</h4>
                  {(group.ops || []).map((op) => (
                    <label key={op.name}>
                      <input type='checkbox' checked={checked.has(op.name)} disabled={detail.builtin} onChange={() => toggle(op.name)} />
                      {op.cname || op.name}<small>{op.name}</small>
                    </label>
                  ))}
                </section>
              ))}
            </div>
          </>
        ) : <div className='fx-org-empty'>请选择角色。</div>}
      </aside>
      {modal && <SimpleModal title={modal.mode === 'edit' ? '编辑角色' : '新增角色'} fields={['name', 'note']} labels={{ name: '名称', note: '备注' }} initial={modal.item} onClose={() => setModal(null)} onSave={saveRole} />}
      {confirmModal}
    </section>
  )
}
