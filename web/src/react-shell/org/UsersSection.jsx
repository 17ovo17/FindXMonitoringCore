import React, { useEffect, useState } from 'react'
import { formatOrgError, orgApi } from '../api/org.js'
import { useConfirm } from '../shared/ConfirmModal.jsx'
import { Pagination } from '../shared/ConfirmModal.jsx'
import { blankUser, joinRoles, toRoleList, useSectionLoader, Modal, Blocked, Field } from './OrgShared.jsx'

export function UsersSection({ q, reloadKey }) {
  const [page, setPage] = useState(1)
  const pageSize = 20
  const { rows, total, loading, error, source, reload } = useSectionLoader('users', q, page, pageSize)
  const [modal, setModal] = useState(null)
  const [feedback, setFeedback] = useState('')
  const [actionError, setActionError] = useState('')
  const [selected, setSelected] = useState(new Set())
  const { confirm, modal: confirmModal } = useConfirm()

  useEffect(() => { if (reloadKey) reload() }, [reloadKey])

  const toggleSelect = (id) => setSelected((prev) => {
    const next = new Set(prev)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    return next
  })
  const toggleAll = () => {
    if (selected.size === rows.length) setSelected(new Set())
    else setSelected(new Set(rows.map((r) => r.id)))
  }

  const batchSetStatus = async (status) => {
    if (!selected.size) return
    const label = status === 1 ? '启用' : '禁用'
    const ok = await confirm({ title: '批量操作确认', message: `确认批量${label}选中的 ${selected.size} 个用户？`, confirmText: label, danger: status !== 1 })
    if (!ok) return
    setActionError('')
    setFeedback('')
    try {
      await Promise.all(Array.from(selected).map((id) => orgApi.users.setStatus(id, { status })))
      setSelected(new Set())
      await reload()
      setFeedback(`已批量${label} ${selected.size} 个用户。`)
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  const save = async (draft) => {
    setActionError('')
    setFeedback('')
    try {
      const body = { ...draft, roles: toRoleList(draft.roles) }
      if (modal?.mode === 'edit') {
        await orgApi.users.update(modal.item.id, body)
        await reload()
        setFeedback('用户资料已更新。')
      } else {
        await orgApi.users.create(body)
        await reload()
        setFeedback('用户已创建，初始密码不会在页面回显。')
      }
      setModal(null)
    } catch (err) {
      setActionError(formatOrgError(err))
      throw err
    }
  }

  const remove = async (row) => {
    const ok = await confirm({ title: '删除用户', message: `确认删除用户 ${row.username}？`, confirmText: '删除', danger: true })
    if (!ok) return
    setActionError('')
    setFeedback('')
    try {
      await orgApi.users.remove(row.id)
      await reload()
      setFeedback('用户已删除。')
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  const resetPassword = async (row) => {
    const password = window.prompt(`为 ${row.username} 输入新密码，至少 6 位。`)
    if (!password) return
    setActionError('')
    setFeedback('')
    try {
      await orgApi.users.resetPassword(row.id, { password, confirm: password })
      setFeedback('密码已重置；新密码不会在页面回显。')
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  const toggleStatus = async (row) => {
    const nextStatus = row.status === 0 ? 1 : 0
    const label = nextStatus === 1 ? '启用' : '禁用'
    const ok = await confirm({ title: `${label}用户`, message: `确认${label}用户 ${row.username}？`, confirmText: label, danger: nextStatus === 0 })
    if (!ok) return
    setActionError('')
    setFeedback('')
    try {
      await orgApi.users.setStatus(row.id, { status: nextStatus })
      await reload()
      setFeedback(`用户已${label}。`)
    } catch (err) {
      setActionError(formatOrgError(err))
    }
  }

  return (
    <section className='fx-org-work'>
      <div className='fx-org-toolbar'>
        <button type='button' onClick={() => setModal({ mode: 'create', item: blankUser })}>新增用户</button>
        <button type='button' onClick={reload}>刷新</button>
        <button type='button' disabled={!selected.size} onClick={() => batchSetStatus(1)}>批量启用</button>
        <button type='button' disabled={!selected.size} onClick={() => batchSetStatus(0)}>批量禁用</button>
        <span>{loading ? '加载中...' : `${source}${selected.size ? ` | 已选 ${selected.size}` : ''}`}</span>
      </div>
      {error && <Blocked>{error}</Blocked>}
      {actionError && <Blocked>{actionError}</Blocked>}
      {feedback && <div className='fx-org-feedback'>{feedback}</div>}
      <div className='fx-org-table'>
        <table>
          <thead><tr><th><input type='checkbox' checked={rows.length > 0 && selected.size === rows.length} onChange={toggleAll} /></th><th>用户名</th><th>显示名</th><th>角色</th><th>团队</th><th>创建时间</th><th>最后活跃</th><th>状态</th><th>操作</th></tr></thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td><input type='checkbox' checked={selected.has(row.id)} onChange={() => toggleSelect(row.id)} /></td>
                <td><strong>{row.username}</strong>{row.must_change_pwd && <em>需改密</em>}</td>
                <td>{row.nickname || '-'}</td>
                <td>{joinRoles(row.roles) || '-'}</td>
                <td>{Array.isArray(row.teams) ? row.teams.map((t) => t.name || t).join(', ') : (row.teams || '-')}</td>
                <td>{row.create_at ? new Date(row.create_at * 1000).toLocaleString('zh-CN') : '-'}</td>
                <td>{row.last_active_time ? new Date(row.last_active_time * 1000).toLocaleString('zh-CN') : '-'}</td>
                <td>{row.status === 0 ? '禁用' : '正常'}</td>
                <td className='fx-org-actions'>
                  <button type='button' onClick={() => setModal({ mode: 'edit', item: { ...row, roles: joinRoles(row.roles) } })}>编辑</button>
                  <button type='button' onClick={() => resetPassword(row)}>修改密码</button>
                  <button type='button' onClick={() => toggleStatus(row)}>{row.status === 0 ? '启用' : '禁用'}</button>
                  <button type='button' className='danger' onClick={() => remove(row)}>删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!rows.length && !loading && <div className='fx-org-empty'>暂无用户。</div>}
      </div>
      <Pagination total={total} page={page} pageSize={pageSize} onPageChange={setPage} />
      {modal && <UserModal modal={modal} onClose={() => setModal(null)} onSave={save} />}
      {confirmModal}
    </section>
  )
}

function UserModal({ modal, onClose, onSave }) {
  const [draft, setDraft] = useState(modal.item || blankUser)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const patch = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }))
  const submit = async () => {
    if (!draft.username?.trim() && modal.mode !== 'edit') { setError('用户名不能为空。'); return }
    if (modal.mode !== 'edit' && !draft.password?.trim()) { setError('初始密码不能为空。'); return }
    setSaving(true)
    setError('')
    try { await onSave(draft) } catch (err) { setError(formatOrgError(err)) } finally { setSaving(false) }
  }
  return (
    <Modal title={modal.mode === 'edit' ? '编辑用户' : '新增用户'} onClose={onClose}>
      <div className='fx-org-form'>
        <Field label='用户名'><input disabled={modal.mode === 'edit'} value={draft.username || ''} onChange={(e) => patch('username', e.target.value)} /></Field>
        {modal.mode !== 'edit' && <Field label='初始密码'><input type='password' value={draft.password || ''} onChange={(e) => patch('password', e.target.value)} /></Field>}
        {modal.mode !== 'edit' && <Field label='确认密码'><input type='password' value={draft.confirm || ''} onChange={(e) => patch('confirm', e.target.value)} /></Field>}
        <Field label='显示名'><input value={draft.nickname || ''} onChange={(e) => patch('nickname', e.target.value)} placeholder='用户显示名称' /></Field>
        <Field label='邮箱'><input value={draft.email || ''} onChange={(e) => patch('email', e.target.value)} /></Field>
        <Field label='电话'><input value={draft.phone || ''} onChange={(e) => patch('phone', e.target.value)} /></Field>
        <Field label='角色'><input value={draft.roles || ''} onChange={(e) => patch('roles', e.target.value)} placeholder='admin, ops, viewer' /></Field>
        <Field label='团队'><input value={draft.teams || ''} onChange={(e) => patch('teams', e.target.value)} placeholder='团队名称，多个用逗号分隔' /></Field>
        {error && <Blocked>{error}</Blocked>}
        <footer><button type='button' disabled={saving} onClick={submit}>{saving ? '保存中...' : '保存'}</button></footer>
      </div>
    </Modal>
  )
}
