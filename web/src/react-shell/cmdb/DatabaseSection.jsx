import React, { useCallback, useEffect, useState } from 'react'
import { CMDB_EXECUTION_BLOCKERS, cmdbApi, cmdbContractMessage, isCmdbContractBlocked } from '../api/cmdb.js'
import { displayText, fmtTime } from './assetsModel.js'
import { Blocked, ErrorBox, Feedback, Field, Modal, Status } from './Shared.jsx'

const DB_TYPES = [
  { value: 'mysql', label: 'MySQL', port: 3306 },
  { value: 'postgresql', label: 'PostgreSQL', port: 5432 },
  { value: 'oracle', label: 'Oracle', port: 1521 },
  { value: 'redis', label: 'Redis', port: 6379 },
  { value: 'mongodb', label: 'MongoDB', port: 27017 },
]

const statusLabel = status => {
  const normalized = String(status || '').toLowerCase()
  return {
    online: '在线',
    offline: '离线',
    unknown: '未知',
    blocked_by_contract: '契约阻断',
    blocked: '契约阻断',
  }[normalized] || displayText(status, '未知')
}

const dbTypeLabel = type => DB_TYPES.find(item => item.value === type)?.label || displayText(type)

const safeDetailEntries = row => [
  ['ID', row.id],
  ['名称', row.name],
  ['类型', dbTypeLabel(row.type)],
  ['地址', row.host],
  ['端口', row.port],
  ['版本', row.version || '-'],
  ['状态', statusLabel(row.status)],
  ['创建时间', fmtTime(row.created_at)],
  ['更新时间', fmtTime(row.updated_at)],
]

const receiptText = receipt => {
  if (!receipt) return ''
  if (typeof receipt === 'string') return receipt
  if (receipt.id || receipt.receipt_id) return `执行回执：${displayText(receipt.id || receipt.receipt_id)}`
  if (receipt.message) return `执行回执：${displayText(receipt.message)}`
  return ''
}

export function DatabaseSection({ onRefresh }) {
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [feedback, setFeedback] = useState('')
  const [filterType, setFilterType] = useState('')
  const [modal, setModal] = useState(null)
  const [testing, setTesting] = useState({})
  const [blocked, setBlocked] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const res = await cmdbApi.databases.list(filterType ? { type: filterType } : {})
      setRows(res?.items || [])
    } catch (err) {
      setError(err?.message || '加载数据库资产失败')
    } finally {
      setLoading(false)
    }
  }, [filterType])

  useEffect(() => { load() }, [load])

  const handleCreate = async (draft) => {
    await cmdbApi.databases.create(draft)
    setFeedback('数据库资产已创建')
    setBlocked('')
    setModal(null)
    await load()
    onRefresh?.()
  }

  const handleDelete = async (row) => {
    if (!window.confirm(`确认删除数据库「${displayText(row.name)}」？`)) return
    setError('')
    try {
      await cmdbApi.databases.remove(row.id)
      setFeedback('数据库资产已删除')
      setBlocked('')
      await load()
      onRefresh?.()
    } catch (err) {
      setError(err?.message || '删除数据库资产失败')
    }
  }

  const handleTest = async (row) => {
    setTesting(prev => ({ ...prev, [row.id]: true }))
    setBlocked('')
    setFeedback('')
    setError('')
    try {
      const res = await cmdbApi.databases.test(row.id)
      const message = receiptText(res?.receipt || res?.execution_receipt)
      if (message) {
        setFeedback(`${displayText(row.name)}：${message}`)
      } else {
        setBlocked(`${displayText(row.name)}：${CMDB_EXECUTION_BLOCKERS.databaseTest}；缺少执行回执`)
      }
      await load()
    } catch (err) {
      if (isCmdbContractBlocked(err)) {
        setBlocked(`${displayText(row.name)}：${cmdbContractMessage(err, CMDB_EXECUTION_BLOCKERS.databaseTest)}`)
      } else {
        setError(`${displayText(row.name)}：${err?.message || '连接测试失败'}`)
      }
    } finally {
      setTesting(prev => ({ ...prev, [row.id]: false }))
    }
  }

  return (
    <section className='fx-assets-work'>
      <div className='fx-assets-toolbar'>
        <button type='button' onClick={() => setModal({ mode: 'create' })}>新建数据库</button>
        <select value={filterType} onChange={e => setFilterType(e.target.value)}>
          <option value=''>全部类型</option>
          {DB_TYPES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
        </select>
        <button type='button' onClick={load}>{loading ? '刷新中...' : '刷新'}</button>
      </div>
      <ErrorBox>{error}</ErrorBox>
      <Feedback>{feedback}</Feedback>
      {blocked && <Blocked>{blocked}</Blocked>}
      <div className='fx-assets-table'>
        <table>
          <thead><tr><th>名称</th><th>类型</th><th>地址</th><th>端口</th><th>版本</th><th>状态</th><th>操作</th></tr></thead>
          <tbody>{rows.map(row => (
            <tr key={row.id}>
              <td><strong>{displayText(row.name)}</strong></td>
              <td>{dbTypeLabel(row.type)}</td>
              <td>{displayText(row.host)}</td>
              <td>{displayText(row.port)}</td>
              <td>{displayText(row.version)}</td>
              <td><Status ok={row.status === 'online'}>{statusLabel(row.status)}</Status></td>
              <td className='fx-assets-actions'>
                <button type='button' onClick={() => setModal({ mode: 'detail', row })}>详情</button>
                <button type='button' onClick={() => handleTest(row)} disabled={testing[row.id]}>{testing[row.id] ? '测试中...' : '连接测试'}</button>
                <button type='button' className='is-danger' onClick={() => handleDelete(row)}>删除</button>
              </td>
            </tr>
          ))}</tbody>
        </table>
        {!rows.length && <div className='fx-assets-empty'>{loading ? '加载中...' : '暂无数据库资产'}</div>}
      </div>
      {modal?.mode === 'create' && <DatabaseCreateModal onClose={() => setModal(null)} onSave={handleCreate} />}
      {modal?.mode === 'detail' && <DatabaseDetailModal row={modal.row} onClose={() => setModal(null)} />}
    </section>
  )
}

function DatabaseCreateModal({ onClose, onSave }) {
  const [draft, setDraft] = useState({ name: '', type: 'mysql', host: '', port: 3306, version: '' })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const patch = (key, value) => setDraft(prev => ({ ...prev, [key]: value }))

  const handleTypeChange = (type) => {
    const preset = DB_TYPES.find(item => item.value === type)
    setDraft(prev => ({ ...prev, type, port: preset?.port || prev.port }))
  }

  const submit = async () => {
    if (!draft.name.trim()) return setError('名称不能为空')
    if (!draft.host.trim()) return setError('地址不能为空')
    setSaving(true)
    setError('')
    try {
      await onSave({ ...draft, name: draft.name.trim(), host: draft.host.trim(), port: Number(draft.port) || 0 })
    } catch (err) {
      setError(err?.message || '创建数据库资产失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal title='新建数据库资产' onClose={onClose}>
      <div className='fx-assets-form'>
        <Field label='名称 *'><input value={draft.name} onChange={e => patch('name', e.target.value)} /></Field>
        <div className='fx-assets-form-grid'>
          <Field label='类型'><select value={draft.type} onChange={e => handleTypeChange(e.target.value)}>{DB_TYPES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}</select></Field>
          <Field label='版本'><input value={draft.version} onChange={e => patch('version', e.target.value)} placeholder='例如 8.0.32' /></Field>
        </div>
        <div className='fx-assets-form-grid'>
          <Field label='地址 *'><input value={draft.host} onChange={e => patch('host', e.target.value)} placeholder='10.0.0.1' /></Field>
          <Field label='端口'><input type='number' value={draft.port} onChange={e => patch('port', e.target.value)} /></Field>
        </div>
        <ErrorBox>{error}</ErrorBox>
        <footer><button type='button' disabled={saving} onClick={submit}>{saving ? '创建中...' : '创建'}</button></footer>
      </div>
    </Modal>
  )
}

function DatabaseDetailModal({ row, onClose }) {
  return (
    <Modal title='数据库详情' onClose={onClose}>
      <div className='fx-assets-form'>
        <div className='fx-assets-table'>
          <table><tbody>{safeDetailEntries(row).map(([key, value]) => <tr key={key}><th style={{ width: 100 }}>{key}</th><td>{displayText(value)}</td></tr>)}</tbody></table>
        </div>
        <p className='fx-assets-muted'>连接私密字段不会在详情中展示。</p>
        <footer><button type='button' onClick={onClose}>关闭</button></footer>
      </div>
    </Modal>
  )
}
