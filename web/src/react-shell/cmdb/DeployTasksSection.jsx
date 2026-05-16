import React, { useCallback, useEffect, useState } from 'react'
import { CMDB_EXECUTION_BLOCKERS, cmdbApi, cmdbBlockedRecordMessage, cmdbContractMessage, isCmdbBlockedRecord, isCmdbContractBlocked } from '../api/cmdb.js'
import { displayText, fmtTime } from './assetsModel.js'
import { Blocked, ErrorBox, Field, Modal, Status } from './Shared.jsx'

const PENDING_STATUSES = new Set(['pending', 'unavailable'])
const EXECUTION_WITHOUT_RECEIPT_STATUSES = new Set([
  'queued',
  'running',
  'succeeded',
  'success',
  'applied',
  'rolled-back',
  'installed',
  'data_arrived',
  'service_registered',
  'heartbeat_seen',
])

const deployStatusLabel = status => {
  const normalized = String(status || '').toLowerCase()
  if (PENDING_STATUSES.has(normalized)) return '待处理'
  if (EXECUTION_WITHOUT_RECEIPT_STATUSES.has(normalized)) return '缺少执行回执'
  return '未知回执状态'
}

const deployProgress = task => {
  if (isCmdbBlockedRecord(task)) return 0
  return 0
}

const missingContracts = task => Array.isArray(task?.missing_contracts) ? task.missing_contracts.filter(Boolean) : []

const blockedMessage = task => {
  if (!task) return CMDB_EXECUTION_BLOCKERS.deploy
  return cmdbBlockedRecordMessage(task, CMDB_EXECUTION_BLOCKERS.deploy)
}

export function DeployTasksSection() {
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [modal, setModal] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const res = await cmdbApi.deployTasks.list()
      setRows(res?.items || [])
    } catch (err) {
      setError(err?.message || '加载部署任务失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handleCreate = async (draft) => {
    try {
      const res = await cmdbApi.deployTasks.create(draft)
      const task = res?.task || res
      setModal(null)
      if (isCmdbBlockedRecord(task)) {
        setBlocked(blockedMessage(task))
      } else {
        setBlocked(CMDB_EXECUTION_BLOCKERS.deploy)
      }
      await load()
    } catch (err) {
      setModal(null)
      if (isCmdbContractBlocked(err)) {
        setBlocked(cmdbContractMessage(err, CMDB_EXECUTION_BLOCKERS.deploy))
        await load()
        return
      }
      setError(err?.message || '创建契约预检失败')
    }
  }

  return (
    <section className='fx-assets-work'>
      <div className='fx-assets-toolbar'>
        <button type='button' onClick={() => setModal({ mode: 'create' })}>创建契约预检</button>
        <button type='button' onClick={load}>{loading ? '刷新中...' : '刷新'}</button>
      </div>
      <Blocked>{CMDB_EXECUTION_BLOCKERS.deploy}</Blocked>
      {blocked && <Blocked>{blocked}</Blocked>}
      <ErrorBox>{error}</ErrorBox>
      <div className='fx-assets-table'>
        <table>
          <thead><tr><th>任务名称</th><th>目标主机</th><th>状态</th><th>进度</th><th>契约</th><th>缺失契约</th><th>创建时间</th><th>操作</th></tr></thead>
          <tbody>{rows.map(row => (
            <tr key={row.id}>
              <td><strong>{displayText(row.name)}</strong></td>
              <td>{(row.target_hosts || []).length} 台</td>
              <td><Status ok={false}>{deployStatusLabel(row.status)}</Status></td>
              <td>
                <div className='fx-deploy-progress-wrap'>
                  <div className='fx-deploy-progress-bar' style={{ width: `${deployProgress(row)}%` }} />
                  <span>{deployProgress(row)}%</span>
                </div>
              </td>
              <td>{displayText(row.contract_id)}</td>
              <td>{missingContracts(row).length ? missingContracts(row).map(displayText).join('、') : '-'}</td>
              <td>{fmtTime(row.created_at)}</td>
              <td className='fx-assets-actions'>
                <button type='button' onClick={() => setModal({ mode: 'detail', taskId: row.id })}>详情</button>
              </td>
            </tr>
          ))}</tbody>
        </table>
        {!rows.length && <div className='fx-assets-empty'>{loading ? '加载中...' : '暂无契约预检记录'}</div>}
      </div>
      {modal?.mode === 'create' && <DeployCreateModal onClose={() => setModal(null)} onSave={handleCreate} />}
      {modal?.mode === 'detail' && <DeployDetailModal taskId={modal.taskId} onClose={() => setModal(null)} />}
    </section>
  )
}

function DeployCreateModal({ onClose, onSave }) {
  const [step, setStep] = useState(1)
  const [draft, setDraft] = useState({ name: '', target_hosts: '', script: '' })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const patch = (key, value) => setDraft(prev => ({ ...prev, [key]: value }))

  const next = () => {
    if (step === 1 && !draft.target_hosts.trim()) return setError('请输入目标主机')
    if (step === 2 && !draft.script.trim()) return setError('请输入预检脚本')
    setError('')
    setStep(prev => prev + 1)
  }

  const submit = async () => {
    if (!draft.name.trim()) return setError('任务名称不能为空')
    setSaving(true)
    setError('')
    try {
      const hosts = draft.target_hosts.split(/[\n,]/).map(host => host.trim()).filter(Boolean)
      await onSave({ name: draft.name.trim(), target_hosts: hosts, script: draft.script.trim() })
    } catch (err) {
      setError(err?.message || '创建契约预检失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal title='创建契约预检' onClose={onClose}>
      <div className='fx-assets-form'>
        <div className='fx-assets-muted' style={{ marginBottom: 8 }}>步骤 {step}/3：{['选择目标', '编写脚本', '确认预检'][step - 1]}</div>
        {step === 1 && (
          <Field label='目标主机（每行一个 IP 或主机名）'>
            <textarea rows='5' value={draft.target_hosts} onChange={e => patch('target_hosts', e.target.value)} placeholder='10.0.0.1&#10;10.0.0.2&#10;web-server-01' />
          </Field>
        )}
        {step === 2 && (
          <Field label='预检脚本'>
            <textarea rows='8' value={draft.script} onChange={e => patch('script', e.target.value)} placeholder='#!/bin/bash&#10;echo "precheck"' style={{ fontFamily: 'monospace' }} />
          </Field>
        )}
        {step === 3 && (
          <>
            <Field label='任务名称 *'>
              <input value={draft.name} onChange={e => patch('name', e.target.value)} placeholder='例如 应用契约预检' />
            </Field>
            <div className='fx-assets-table'><table><tbody>
              <tr><th>目标主机</th><td>{draft.target_hosts.split(/[\n,]/).filter(host => host.trim()).length} 台</td></tr>
              <tr><th>脚本长度</th><td>{draft.script.length} 字符</td></tr>
            </tbody></table></div>
          </>
        )}
        <ErrorBox>{error}</ErrorBox>
        <footer>
          {step > 1 && <button type='button' onClick={() => setStep(value => value - 1)}>上一步</button>}
          {step < 3 && <button type='button' onClick={next}>下一步</button>}
          {step === 3 && <button type='button' disabled={saving} onClick={submit}>{saving ? '校验中...' : '发起预检'}</button>}
        </footer>
      </div>
    </Modal>
  )
}

function DeployDetailModal({ taskId, onClose }) {
  const [task, setTask] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true
    const poll = async () => {
      try {
        const res = await cmdbApi.deployTasks.get(taskId)
        if (active) setTask(res)
      } catch (err) {
        if (active) setError(err?.message || '加载契约预检详情失败')
      }
    }
    poll()
    const timer = setInterval(poll, 3000)
    return () => { active = false; clearInterval(timer) }
  }, [taskId])

  if (!task && !error) return <Modal title='契约预检详情' onClose={onClose}><div className='fx-assets-form'><p>加载中...</p></div></Modal>

  return (
    <Modal title={`契约预检详情 - ${displayText(task?.name || taskId)}`} onClose={onClose}>
      <div className='fx-assets-form'>
        <ErrorBox>{error}</ErrorBox>
        {task && (
          <>
            <div className='fx-assets-table'><table><tbody>
              <tr><th>状态</th><td><Status ok={false}>{deployStatusLabel(task.status)}</Status></td></tr>
              <tr><th>进度</th><td>{deployProgress(task)}%</td></tr>
              <tr><th>目标数量</th><td>{(task.target_hosts || []).length} 台</td></tr>
              <tr><th>创建时间</th><td>{fmtTime(task.created_at)}</td></tr>
              <tr><th>缺失契约数量</th><td>{missingContracts(task).length}</td></tr>
            </tbody></table></div>
            <Blocked>{blockedMessage(task)}</Blocked>
            <h4 style={{ margin: '12px 0 8px', fontSize: 13, color: '#526984' }}>阻断记录</h4>
            <pre className='fx-exec-stdout' style={{ maxHeight: 200, overflow: 'auto' }}>
              {[
                `status=${deployStatusLabel(task.status)}`,
                `progress=${deployProgress(task)}%`,
                `missing_contracts_count=${missingContracts(task).length}`,
              ].join('\n')}
            </pre>
          </>
        )}
        <footer><button type='button' onClick={onClose}>关闭</button></footer>
      </div>
    </Modal>
  )
}
