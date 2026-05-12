import React, { useEffect, useMemo, useState } from 'react'
import { formatOrgError, orgApi, ORG_BLOCKERS } from '../api/org.js'
import { useConfirm } from '../shared/ConfirmModal.jsx'
import './org.css'

const sections = [
  { value: 'users', label: '用户管理', desc: '按用户资料、最后活跃、团队和业务组关系维护登录主体。' },
  { value: 'teams', label: '团队组织', desc: '按列表/树结构维护团队，并管理团队成员。' },
  { value: 'business', label: '业务组', desc: '维护业务组树和团队授权关系，作为监控资源范围基础。' },
  { value: 'roles', label: '角色管理', desc: '维护角色列表、内置角色保护和操作权限树。' },
  { value: 'audit', label: '审计日志', desc: '敏感操作、权限和配置变更留痕，支持筛选和导出。' },
]

const sectionSet = new Set(sections.map((item) => item.value))
const blankUser = { username: '', password: '', confirm: '', nickname: '', email: '', phone: '', roles: 'viewer' }
const blankTeam = { name: '', note: '', parent_id: '' }
const blankBusiness = { name: '', note: '', parent_id: '' }
const blankRole = { name: '', note: '' }

const asRows = (value) => Array.isArray(value) ? value : []
const joinRoles = (roles) => Array.isArray(roles) ? roles.join(', ') : String(roles || '')
const toRoleList = (value) => String(value || '').split(',').map((item) => item.trim()).filter(Boolean)
const buildTree = (rows) => {
  const map = {}
  const roots = []
  rows.forEach((row) => { map[row.id] = { ...row, children: [] } })
  rows.forEach((row) => {
    if (row.parent_id && map[row.parent_id]) map[row.parent_id].children.push(map[row.id])
    else roots.push(map[row.id])
  })
  return roots
}

function Modal({ title, children, onClose }) {
  return (
    <div className='fx-org-modal'>
      <div className='fx-org-modal__body'>
        <header>
          <h2>{title}</h2>
          <button type='button' onClick={onClose}>关闭</button>
        </header>
        {children}
      </div>
    </div>
  )
}

function Blocked({ children }) {
  return <div className='fx-org-blocked'><strong>BLOCKED_BY_CONTRACT</strong><span>{children}</span></div>
}

function Field({ label, children }) {
  return <label className='fx-org-field'><span>{label}</span>{children}</label>
}

function useSectionLoader(section, q) {
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [source, setSource] = useState('')

  const load = async () => {
    setLoading(true)
    setError('')
    try {
      const result = await orgApi[section].list({ q, page: 1, limit: 100 })
      setRows(asRows(result.rows))
      setSource(result.source || '')
    } catch (err) {
      setRows([])
      setError(formatOrgError(err))
      setSource('')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [section, q])

  return { rows, loading, error, source, reload: load }
}

function UsersSection({ q, reloadKey }) {
  const { rows, loading, error, source, reload } = useSectionLoader('users', q)
  const [modal, setModal] = useState(null)
  const [feedback, setFeedback] = useState('')
  const [actionError, setActionError] = useState('')
  const [selected, setSelected] = useState(new Set())
  const { confirm, modal: confirmModal } = useConfirm()

  useEffect(() => {
    if (reloadKey) reload()
  }, [reloadKey])

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
          <thead><tr><th><input type='checkbox' checked={rows.length > 0 && selected.size === rows.length} onChange={toggleAll} /></th><th>用户名</th><th>角色</th><th>邮箱</th><th>手机</th><th>最后活跃</th><th>状态</th><th>操作</th></tr></thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td><input type='checkbox' checked={selected.has(row.id)} onChange={() => toggleSelect(row.id)} /></td>
                <td><strong>{row.username}</strong>{row.must_change_pwd && <em>需改密</em>}</td>
                <td>{joinRoles(row.roles) || '-'}</td>
                <td>{row.email || '-'}</td>
                <td>{row.phone || '-'}</td>
                <td>{row.last_active_time ? new Date(row.last_active_time * 1000).toLocaleString('zh-CN') : '-'}</td>
                <td>{row.status === 0 ? '禁用' : '正常'}</td>
                <td className='fx-org-actions'>
                  <button type='button' onClick={() => setModal({ mode: 'edit', item: { ...row, roles: joinRoles(row.roles) } })}>编辑</button>
                  <button type='button' onClick={() => resetPassword(row)}>重置密码</button>
                  <button type='button' className='danger' onClick={() => remove(row)}>删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!rows.length && !loading && <div className='fx-org-empty'>暂无用户。新增用户会进入真实组织契约，不创建临时列表。</div>}
      </div>
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
    if (!draft.username?.trim() && modal.mode !== 'edit') {
      setError('用户名不能为空。')
      return
    }
    if (modal.mode !== 'edit' && !draft.password?.trim()) {
      setError('初始密码不能为空。')
      return
    }
    setSaving(true)
    setError('')
    try {
      await onSave(draft)
    } catch (err) {
      setError(formatOrgError(err))
    } finally {
      setSaving(false)
    }
  }
  return (
    <Modal title={modal.mode === 'edit' ? '编辑用户' : '新增用户'} onClose={onClose}>
      <div className='fx-org-form'>
        <Field label='用户名'><input disabled={modal.mode === 'edit'} value={draft.username || ''} onChange={(e) => patch('username', e.target.value)} /></Field>
        {modal.mode !== 'edit' && <Field label='初始密码'><input type='password' value={draft.password || ''} onChange={(e) => patch('password', e.target.value)} /></Field>}
        {modal.mode !== 'edit' && <Field label='确认密码'><input type='password' value={draft.confirm || ''} onChange={(e) => patch('confirm', e.target.value)} /></Field>}
        <Field label='昵称'><input value={draft.nickname || ''} onChange={(e) => patch('nickname', e.target.value)} /></Field>
        <Field label='邮箱'><input value={draft.email || ''} onChange={(e) => patch('email', e.target.value)} /></Field>
        <Field label='电话'><input value={draft.phone || ''} onChange={(e) => patch('phone', e.target.value)} /></Field>
        <Field label='角色'><input value={draft.roles || ''} onChange={(e) => patch('roles', e.target.value)} placeholder='viewer, ops' /></Field>
        {error && <Blocked>{error}</Blocked>}
        <footer><button type='button' disabled={saving} onClick={submit}>{saving ? '保存中...' : '保存'}</button></footer>
      </div>
    </Modal>
  )
}

function TreeNode({ node, depth, selected, onSelect }) {
  return (
    <>
      <button type='button' className={selected?.id === node.id ? 'fx-org-row is-active' : 'fx-org-row'} style={{ paddingLeft: `${10 + depth * 18}px` }} onClick={() => onSelect(node)}>
        <strong>{depth > 0 ? '└ ' : ''}{node.name}</strong><span>{node.note || '无备注'}</span>
      </button>
      {node.children && node.children.map((child) => <TreeNode key={child.id} node={child} depth={depth + 1} selected={selected} onSelect={onSelect} />)}
    </>
  )
}

function TeamsSection({ q }) {
  const { rows, loading, error, source, reload } = useSectionLoader('teams', q)
  const [selected, setSelected] = useState(null)
  const [modal, setModal] = useState(null)
  const [memberText, setMemberText] = useState('')
  const { confirm, modal: confirmModal } = useConfirm()
  const [actionError, setActionError] = useState('')
  const [actionFeedback, setActionFeedback] = useState('')
  const detail = selected || rows[0] || null
  const tree = useMemo(() => buildTree(rows), [rows])

  useEffect(() => {
    setSelected(rows[0] || null)
  }, [rows])

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
        {!rows.length && !loading && <div className='fx-org-empty'>暂无团队。团队创建走真实组织契约。</div>}
      </div>
      <aside className='fx-org-detail'>
        {detail ? (
          <>
            <header><h3>{detail.name}</h3><p>{detail.note || '无备注'}{detail.parent_id ? ` | 父级: ${detail.parent_id}` : ''}</p></header>
            <div className='fx-org-actions'><button type='button' onClick={() => setModal({ mode: 'edit', item: detail })}>编辑</button><button type='button' className='danger' onClick={() => remove(detail)}>删除</button></div>
            <div className='fx-org-member-box'>
              <h4>成员</h4>
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

function BusinessSection({ q }) {
  const { rows, loading, error, source, reload } = useSectionLoader('business', q)
  const [selected, setSelected] = useState(null)
  const [modal, setModal] = useState(null)
  const [teamID, setTeamID] = useState('')
  const [actionError, setActionError] = useState('')
  const [actionFeedback, setActionFeedback] = useState('')
  const { confirm, modal: confirmModal } = useConfirm()
  const detail = selected || rows[0] || null

  useEffect(() => {
    setSelected(rows[0] || null)
  }, [rows])

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
        {!rows.length && !loading && <div className='fx-org-empty'>暂无业务组。业务组创建走真实组织契约。</div>}
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

function RolesSection({ q }) {
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
  useEffect(() => {
    setSelected(rows[0] || null)
  }, [rows])
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
              {operations.map((group) => <section key={group.name}><h4>{group.cname || group.name}</h4>{(group.ops || []).map((op) => <label key={op.name}><input type='checkbox' checked={checked.has(op.name)} disabled={detail.builtin} onChange={() => toggle(op.name)} />{op.cname || op.name}<small>{op.name}</small></label>)}</section>)}
            </div>
          </>
        ) : <div className='fx-org-empty'>请选择角色。</div>}
      </aside>
      {modal && <SimpleModal title={modal.mode === 'edit' ? '编辑角色' : '新增角色'} fields={['name', 'note']} labels={{ name: '名称', note: '备注' }} initial={modal.item} onClose={() => setModal(null)} onSave={saveRole} />}
      {confirmModal}
    </section>
  )
}

function SimpleModal({ title, fields, labels, initial, onClose, onSave }) {
  const [draft, setDraft] = useState(initial || {})
  const [error, setError] = useState('')
  const submit = async () => {
    setError('')
    try {
      await onSave(draft)
    } catch (err) {
      setError(formatOrgError(err))
    }
  }
  return (
    <Modal title={title} onClose={onClose}>
      <div className='fx-org-form'>
        {fields.map((field) => <Field key={field} label={labels[field] || field}><input value={draft[field] || ''} onChange={(e) => setDraft((prev) => ({ ...prev, [field]: e.target.value }))} /></Field>)}
        {error && <Blocked>{error}</Blocked>}
        <footer><button type='button' onClick={submit}>保存</button></footer>
      </div>
    </Modal>
  )
}

function AuditSection({ q }) {
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [source, setSource] = useState('')
  const [expanded, setExpanded] = useState(null)
  const [filters, setFilters] = useState({ action: '', operator: '', target: '', start: '', end: '' })

  const load = async () => {
    try {
      const params = { q, ...Object.fromEntries(Object.entries(filters).filter(([, v]) => v)) }
      const result = await orgApi.audit(params)
      setRows(result.rows || [])
      setSource(result.source || '')
      setError('')
    } catch (err) {
      setError(formatOrgError(err))
    }
  }
  useEffect(() => { load() }, [q])

  const doFilter = () => load()
  const exportPlaceholder = () => window.alert('导出功能待后端契约开放后接入。')

  return (
    <section className='fx-org-work'>
      <div className='fx-org-toolbar'>
        <button type='button' onClick={doFilter}>筛选</button>
        <button type='button' onClick={load}>刷新</button>
        <button type='button' onClick={exportPlaceholder}>导出</button>
        <span>{source}</span>
      </div>
      <div className='fx-org-filter'>
        <input value={filters.action} onChange={(e) => setFilters((p) => ({ ...p, action: e.target.value }))} placeholder='动作' />
        <input value={filters.operator} onChange={(e) => setFilters((p) => ({ ...p, operator: e.target.value }))} placeholder='操作人' />
        <input value={filters.target} onChange={(e) => setFilters((p) => ({ ...p, target: e.target.value }))} placeholder='对象' />
        <input type='date' value={filters.start} onChange={(e) => setFilters((p) => ({ ...p, start: e.target.value }))} />
        <input type='date' value={filters.end} onChange={(e) => setFilters((p) => ({ ...p, end: e.target.value }))} />
      </div>
      {error && <Blocked>{error}</Blocked>}
      <div className='fx-org-table'>
        <table>
          <thead><tr><th>时间</th><th>动作</th><th>操作人</th><th>对象</th><th>风险</th><th>决策</th><th>详情</th></tr></thead>
          <tbody>
            {rows.map((row, idx) => (
              <React.Fragment key={row.id || idx}>
                <tr onClick={() => setExpanded(expanded === idx ? null : idx)} style={{ cursor: 'pointer' }}>
                  <td>{row.timestamp ? new Date(row.timestamp).toLocaleString('zh-CN') : row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}</td>
                  <td>{row.action || '-'}</td>
                  <td>{row.operator || '-'}</td>
                  <td>{row.target || '-'}</td>
                  <td>{row.risk || '-'}</td>
                  <td>{row.decision || '-'}</td>
                  <td>{expanded === idx ? '收起' : '展开'}</td>
                </tr>
                {expanded === idx && (
                  <tr><td colSpan={7}><pre style={{ whiteSpace: 'pre-wrap', margin: 0, fontSize: 12, color: '#475569' }}>{row.description || row.detail || JSON.stringify(row, null, 2)}</pre></td></tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
        {!rows.length && <div className='fx-org-empty'>暂无审计日志。</div>}
      </div>
    </section>
  )
}

export function OrgPage({ query = {}, onNavigate }) {
  const section = sectionSet.has(query.section) ? query.section : 'users'
  const current = useMemo(() => sections.find((item) => item.value === section), [section])
  const [q, setQ] = useState(query.q || '')
  const [reloadKey, setReloadKey] = useState(0)

  useEffect(() => {
    setQ(query.q || '')
  }, [query.q])

  const commitSearch = () => onNavigate?.({ section, q })

  return (
    <main className='fx-org-page'>
      <header className='fx-org-head'>
        <div><p>FindX 组织治理</p><h1>人员组织</h1><span>{current?.desc}</span></div>
        <nav>{sections.map((item) => <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate?.({ section: item.value })}>{item.label}</button>)}</nav>
      </header>
      <section className='fx-org-filter'>
        <input value={q} onChange={(e) => setQ(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') commitSearch() }} placeholder='搜索名称、用户或备注' />
        <button type='button' onClick={commitSearch}>搜索</button>
        <button type='button' onClick={() => setReloadKey((value) => value + 1)}>刷新</button>
      </section>
      {section === 'users' && <UsersSection q={q} reloadKey={reloadKey} />}
      {section === 'teams' && <TeamsSection q={q} />}
      {section === 'business' && <BusinessSection q={q} />}
      {section === 'roles' && <RolesSection q={q} />}
      {section === 'audit' && <AuditSection q={q} />}
    </main>
  )
}
