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

const blankEngine = { name: '', address: '', cluster: '', datasource: '' }

export function AlertingEnginesSection() {
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [feedback, setFeedback] = useState('')
  const [modal, setModal] = useState(null)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [checking, setChecking] = useState(null)
  const pageSize = 20
  const { confirm, modal: confirmModal } = useConfirm()

  const load = async () => {
    setError('')
    try {
      const result = await platformApi.listEngines({ page, limit: pageSize })
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
        await platformApi.updateEngine(modal.item.id, draft)
        setFeedback('告警引擎已更新。')
      } else {
        await platformApi.createEngine(draft)
        setFeedback('告警引擎已添加。')
      }
      setModal(null)
      await load()
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }

  const remove = async (row) => {
    const ok = await confirm({ title: '删除告警引擎', message: `确认删除引擎 ${row.name || row.address}？`, confirmText: '删除', danger: true })
    if (!ok) return
    setError('')
    setFeedback('')
    try {
      await platformApi.deleteEngine(row.id)
      setFeedback('告警引擎已删除。')
      await load()
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }

  const healthCheck = async (row) => {
    setChecking(row.id)
    setFeedback('')
    setError('')
    try {
      const result = await platformApi.checkEngineHealth(row.id)
      setFeedback(`引擎 ${row.name || row.address} 健康检查：${result?.status || result?.message || '正常'}`)
    } catch (err) {
      setError(formatPlatformError(err))
    } finally {
      setChecking(null)
    }
  }

  const formatTime = (ts) => {
    if (!ts) return '-'
    const d = typeof ts === 'number' && ts < 1e12 ? new Date(ts * 1000) : new Date(ts)
    return d.toLocaleString('zh-CN')
  }

  return (
    <section className='fx-platform-contract'>
      <header><h2>告警引擎</h2></header>
      <div className='fx-platform-toolbar'>
        <button type='button' onClick={() => setModal({ mode: 'create', item: blankEngine })}>添加引擎</button>
        <button type='button' onClick={load}>刷新</button>
      </div>
      {error && <div className='fx-platform-error'>{error}</div>}
      {feedback && <div className='fx-platform-feedback'>{feedback}</div>}
      <div className='fx-platform-table'>
        <table>
          <thead>
            <tr><th>名称</th><th>地址</th><th>集群</th><th>数据源</th><th>状态</th><th>最后心跳</th><th>操作</th></tr>
          </thead>
          <tbody>
            {rows.map((row, idx) => (
              <tr key={row.id || idx}>
                <td>{row.name || row.instance || '-'}</td>
                <td><code>{row.address || row.endpoint || '-'}</code></td>
                <td>{row.cluster || '-'}</td>
                <td>{row.datasource || row.datasource_id || '-'}</td>
                <td><span className={row.status === 'healthy' || row.alive ? 'fx-status-ok' : 'fx-status-warn'}>{row.status || (row.alive ? '健康' : '未知')}</span></td>
                <td>{formatTime(row.last_heartbeat || row.clock)}</td>
                <td className='fx-org-actions'>
                  <button type='button' onClick={() => setModal({ mode: 'edit', item: row })}>编辑</button>
                  <button type='button' disabled={checking === row.id} onClick={() => healthCheck(row)}>{checking === row.id ? '检查中...' : '健康检查'}</button>
                  <button type='button' className='danger' onClick={() => remove(row)}>删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!rows.length && <div className='fx-platform-empty'>暂无告警引擎。</div>}
      </div>
      <Pagination total={total} page={page} pageSize={pageSize} onPageChange={setPage} />
      {modal && <EngineModal modal={modal} onClose={() => setModal(null)} onSave={save} />}
      {confirmModal}
    </section>
  )
}

function EngineModal({ modal, onClose, onSave }) {
  const [draft, setDraft] = useState(modal.item || blankEngine)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const patch = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }))

  const submit = async () => {
    if (!draft.name?.trim() && !draft.address?.trim()) { setError('名称或地址不能为空。'); return }
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
    <Modal title={modal.mode === 'edit' ? '编辑告警引擎' : '添加告警引擎'} onClose={onClose}>
      <div className='fx-platform-form'>
        <Field label='名称'><input value={draft.name || ''} onChange={(e) => patch('name', e.target.value)} placeholder='engine-01' /></Field>
        <Field label='地址'><input value={draft.address || ''} onChange={(e) => patch('address', e.target.value)} placeholder='http://engine:9090' /></Field>
        <Field label='集群'><input value={draft.cluster || ''} onChange={(e) => patch('cluster', e.target.value)} placeholder='default' /></Field>
        <Field label='数据源'><input value={draft.datasource || ''} onChange={(e) => patch('datasource', e.target.value)} placeholder='prometheus' /></Field>
        {error && <div className='fx-platform-error'>{error}</div>}
        <footer><button type='button' disabled={saving} onClick={submit}>{saving ? '保存中...' : '保存'}</button></footer>
      </div>
    </Modal>
  )
}
