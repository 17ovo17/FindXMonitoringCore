import React, { useEffect, useState } from 'react'
import { formatOrgError, orgApi } from '../api/org.js'

export const blankUser = { username: '', password: '', confirm: '', nickname: '', email: '', phone: '', roles: 'viewer', teams: '' }
export const blankTeam = { name: '', note: '', parent_id: '' }
export const blankBusiness = { name: '', note: '', parent_id: '' }
export const blankRole = { name: '', note: '' }

export const asRows = (value) => Array.isArray(value) ? value : []
export const joinRoles = (roles) => Array.isArray(roles) ? roles.join(', ') : String(roles || '')
export const toRoleList = (value) => String(value || '').split(',').map((item) => item.trim()).filter(Boolean)

export const buildTree = (rows) => {
  const map = {}
  const roots = []
  rows.forEach((row) => { map[row.id] = { ...row, children: [] } })
  rows.forEach((row) => {
    if (row.parent_id && map[row.parent_id]) map[row.parent_id].children.push(map[row.id])
    else roots.push(map[row.id])
  })
  return roots
}

export function Modal({ title, children, onClose }) {
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

export function Blocked({ children }) {
  return <div className='fx-org-blocked'><strong>PENDING</strong><span>{children}</span></div>
}

export function Field({ label, children }) {
  return <label className='fx-org-field'><span>{label}</span>{children}</label>
}

export function useSectionLoader(section, q, page = 1, pageSize = 20) {
  const [rows, setRows] = useState([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [source, setSource] = useState('')

  const load = async () => {
    setLoading(true)
    setError('')
    try {
      const result = await orgApi[section].list({ q, page, limit: pageSize })
      setRows(asRows(result.rows))
      setTotal(result.total || result.rows?.length || 0)
      setSource(result.source || '')
    } catch (err) {
      setRows([])
      setTotal(0)
      setError(formatOrgError(err))
      setSource('')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [section, q, page])

  return { rows, total, loading, error, source, reload: load }
}

export function SimpleModal({ title, fields, labels, initial, onClose, onSave }) {
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

export function TreeNode({ node, depth, selected, onSelect }) {
  return (
    <>
      <button type='button' className={selected?.id === node.id ? 'fx-org-row is-active' : 'fx-org-row'} style={{ paddingLeft: `${10 + depth * 18}px` }} onClick={() => onSelect(node)}>
        <strong>{depth > 0 ? '└ ' : ''}{node.name}</strong><span>{node.note || '无备注'}</span>
      </button>
      {node.children && node.children.map((child) => <TreeNode key={child.id} node={child} depth={depth + 1} selected={selected} onSelect={onSelect} />)}
    </>
  )
}
