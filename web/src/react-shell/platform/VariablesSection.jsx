import React, { useEffect, useState } from 'react'
import { formatPlatformError, platformApi } from '../api/platform.js'
import { useConfirm } from '../shared/ConfirmModal.jsx'
import { Pagination } from '../shared/ConfirmModal.jsx'

function Field({ label, children }) {
  return <label className='fx-platform-field'><span>{label}</span>{children}</label>
}

function Modal({ title, children, onClose }) {
  return (
    <div className='fx-platform-modal'>
      <div className='fx-platform-modal__body'>
        <header><h2>{title}</h2><button type='button' onClick={onClose}>关闭</button></header>
        {children}
      </div>
    </div>
  )
}

const blankVariable = { key: '', value: '', description: '', encrypted: false }

export function VariablesSection() {
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [feedback, setFeedback] = useState('')
  const [modal, setModal] = useState(null)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const pageSize = 20
  const { confirm, modal: confirmModal } = useConfirm()

  const load = async () => {
    setError('')
    try {
      const result = await platformApi.listVariables({ page, limit: pageSize })
      setRows(result.rows || [])
      setTotal(result.rows?.length || 0)
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }

  useEffect(() => { load() }, [page])

  const save = async (draft) => {
    setError('')
    setFeedback('')
    try {
      if (modal?.mode === 'edit') {
        await platformApi.updateVariable(modal.item.id || modal.item.key, draft)
        setFeedback('变量已更新。')
      } else {
        await platformApi.createVariable(draft)
        setFeedback('变量已创建。')
      }
      setModal(null)
      await load()
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }

  const remove = async (row) => {
    const ok = await confirm({ title: '删除变量', message: `确认删除变量 ${row.key}？`, confirmText: '删除', danger: true })
    if (!ok) return
    setError('')
    setFeedback('')
    try {
      await platformApi.deleteVariable(row.id || row.key)
      setFeedback('变量已删除。')
      await load()
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }

  return (
    <section className='fx-platform-contract'>
      <header><h2>变量设置</h2></header>
      <div className='fx-platform-toolbar'>
        <button type='button' onClick={() => setModal({ mode: 'create', item: blankVariable })}>新增变量</button>
        <button type='button' onClick={load}>刷新</button>
      </div>
      {error && <div className='fx-platform-error'>{error}</div>}
      {feedback && <div className='fx-platform-feedback'>{feedback}</div>}
      <div className='fx-platform-table'>
        <table>
          <thead>
            <tr><th>变量名</th><th>值</th><th>描述</th><th>加密</th><th>操作</th></tr>
          </thead>
          <tbody>
            {rows.map((row, idx) => (
              <tr key={row.id || row.key || idx}>
                <td><code>{row.key}</code></td>
                <td>{row.encrypted ? '******' : (row.value || '-')}</td>
                <td>{row.description || '-'}</td>
                <td>{row.encrypted ? '是' : '否'}</td>
                <td className='fx-org-actions'>
                  <button type='button' onClick={() => setModal({ mode: 'edit', item: row })}>编辑</button>
                  <button type='button' className='danger' onClick={() => remove(row)}>删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!rows.length && <div className='fx-platform-empty'>暂无变量配置。变量可在告警模板和通知模板中通过 ${'{'}key{'}'} 引用。</div>}
      </div>
      <Pagination total={total} page={page} pageSize={pageSize} onPageChange={setPage} />
      {modal && <VariableModal modal={modal} onClose={() => setModal(null)} onSave={save} />}
      {confirmModal}
    </section>
  )
}

function VariableModal({ modal, onClose, onSave }) {
  const [draft, setDraft] = useState(modal.item || blankVariable)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const patch = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }))

  const submit = async () => {
    if (!draft.key?.trim()) { setError('变量名不能为空。'); return }
    setSaving(true)
    setError('')
    try {
      await onSave(draft)
    } catch (err) {
      setError(err?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal title={modal.mode === 'edit' ? '编辑变量' : '新增变量'} onClose={onClose}>
      <div className='fx-platform-form'>
        <Field label='变量名 (Key)'>
          <input value={draft.key || ''} disabled={modal.mode === 'edit'} onChange={(e) => patch('key', e.target.value)} placeholder='NOTIFY_WEBHOOK_URL' />
        </Field>
        <Field label='值 (Value)'>
          <input type={draft.encrypted ? 'password' : 'text'} value={draft.value || ''} onChange={(e) => patch('value', e.target.value)} />
        </Field>
        <Field label='描述'>
          <input value={draft.description || ''} onChange={(e) => patch('description', e.target.value)} placeholder='变量用途说明' />
        </Field>
        <Field label='加密存储'>
          <select value={draft.encrypted ? 'true' : 'false'} onChange={(e) => patch('encrypted', e.target.value === 'true')}>
            <option value='false'>明文</option>
            <option value='true'>加密</option>
          </select>
        </Field>
        {error && <div className='fx-platform-error'>{error}</div>}
        <footer><button type='button' disabled={saving} onClick={submit}>{saving ? '保存中...' : '保存'}</button></footer>
      </div>
    </Modal>
  )
}
