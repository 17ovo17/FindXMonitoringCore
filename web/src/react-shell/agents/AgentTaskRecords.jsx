import React, { useEffect, useMemo, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { displayText, fmtTime } from './agentModel.js'
import { Blocked, Empty, ErrorBox, Status, Tags } from './AgentShared.jsx'

const TASK_ACTION_RE = /(uninstall|upgrade|rollback|restart|package)/i
const SENSITIVE_KEY_RE = /(token|cookie|session|password|passwd|secret|dsn|private[_-]?key|credential|bearer|api[_-]?key|access[_-]?key|authorization)/i
const SAFE_REF_RE = /(^|_)(audit|evidence|receipt|checksum|fingerprint|ref|refs|safe_refs?)$/i
const MAX_REF_LENGTH = 96
const BLOCKED_NOTE = '任务审计面板只展示 status=blocked 的 uninstall / upgrade / rollback / restart / package 记录；真实执行仍为 BLOCKED_BY_CONTRACT。'

const isBlocked = record => String(record?.status || '').toLowerCase() === 'blocked'
const recordId = record => record?.id || record?.task_id || record?.record_id || ''
const safeObject = value => value && typeof value === 'object' && !Array.isArray(value) ? value : {}
const metadataOf = record => safeObject(record?.metadata)
const pick = (record, ...keys) => keys.map(key => record?.[key]).find(value => value !== undefined && value !== null && value !== '')
const actionText = record => String(pick(record, 'action', 'task_action', 'type') || '').toLowerCase()
const isVisibleAction = record => TASK_ACTION_RE.test(actionText(record))
const normalizeRecords = rows => rows.filter(record => isBlocked(record) && isVisibleAction(record))
const blockerText = record => pick(record, 'blocker', 'blocked_reason', 'reason', 'message', 'error_summary') || 'BLOCKED_BY_CONTRACT'
const statusText = record => isBlocked(record) ? 'blocked' : 'BLOCKED_BY_CONTRACT'
const packageText = record => displayText(pick(record, 'package_id', 'package', 'package_name'), '-')
const configText = record => displayText(pick(record, 'config_version', 'version', 'template_version'), '-')

const listText = value => {
  if (Array.isArray(value)) return value.filter(Boolean).slice(0, 4).join(' / ') || '-'
  return displayText(value, '-')
}

const targetText = record => {
  const values = [
    pick(record, 'target_id', 'target', 'host_id'),
    ...(Array.isArray(record?.target_ids) ? record.target_ids : []),
    ...(Array.isArray(record?.agent_ids) ? record.agent_ids : []),
  ].filter(Boolean)
  return values.length ? values.slice(0, 4).join(' / ') : '-'
}

const redactValue = (key, value) => {
  if (SENSITIVE_KEY_RE.test(String(key))) return '<REDACTED_REF>'
  const text = Array.isArray(value) ? value.join(', ') : String(value ?? '').trim()
  if (!text) return ''
  if (SENSITIVE_KEY_RE.test(text)) return '<REDACTED_REF>'
  return text.length > MAX_REF_LENGTH ? `${text.slice(0, MAX_REF_LENGTH)}...` : text
}

const collectSafeRefs = record => {
  const rows = []
  const pushRef = (scope, key, value) => {
    if (!SAFE_REF_RE.test(key) || SENSITIVE_KEY_RE.test(key)) return
    if (value && typeof value === 'object' && !Array.isArray(value)) return
    rows.push({ scope, key, value: redactValue(key, value) || 'BLOCKED_BY_CONTRACT' })
  }
  Object.entries(safeObject(record)).forEach(([key, value]) => pushRef('record', key, value))
  Object.entries(safeObject(record?.safe_refs)).forEach(([key, value]) => pushRef('safe_refs', key, value))
  Object.entries(metadataOf(record)).forEach(([key, value]) => pushRef('metadata', key, value))
  return rows
}

function Summary({ records, loading, onRefresh }) {
  return (
    <div className='fx-agent-toolbar'>
      <button type='button' onClick={onRefresh} disabled={loading}>{loading ? '刷新中...' : '刷新 blocked tasks'}</button>
      <span className='fx-agent-selected'>blocked tasks {records.length} 条</span>
      <span className='fx-agent-muted'>非 blocked 或非生命周期任务响应会被前端丢弃，不显示假成功状态。</span>
    </div>
  )
}

function RecordsTable({ records, loadingDetail, onSelect }) {
  if (!records.length) return <Empty>暂无 blocked task 审计记录；这不表示 uninstall、upgrade、rollback、restart 或 package 动作已经执行成功。</Empty>
  return (
    <div className='fx-agent-table'>
      <table>
        <thead><tr><th>action</th><th>target / agent ids</th><th>package</th><th>config_version</th><th>status</th><th>blocker</th><th>updated_at</th><th>操作</th></tr></thead>
        <tbody>{records.map(record => <RecordRow key={recordId(record) || `${actionText(record)}-${targetText(record)}`} record={record} loadingDetail={loadingDetail} onSelect={onSelect} />)}</tbody>
      </table>
    </div>
  )
}

function RecordRow({ record, loadingDetail, onSelect }) {
  const id = recordId(record)
  return (
    <tr>
      <td>{displayText(actionText(record), 'BLOCKED_BY_CONTRACT')}</td>
      <td>{targetText(record)}</td>
      <td>{packageText(record)}</td>
      <td>{configText(record)}</td>
      <td><Status ok={false}>{statusText(record)}</Status></td>
      <td><span className='fx-agent-muted'>{displayText(blockerText(record))}</span></td>
      <td>{fmtTime(pick(record, 'updated_at', 'updatedAt', 'created_at'))}</td>
      <td className='fx-agent-actions'><button type='button' disabled={!id || loadingDetail === id} onClick={() => onSelect(id)}>{loadingDetail === id ? '加载中...' : '详情'}</button></td>
    </tr>
  )
}

function Detail({ record }) {
  if (!record) return null
  const refs = collectSafeRefs(record)
  return (
    <div className='fx-agent-panel'>
      <h3>blocked task detail</h3>
      <p>{BLOCKED_NOTE}</p>
      <div className='fx-agent-summary-row'><Status ok={false}>{statusText(record)}</Status><span>task {displayText(recordId(record))}</span></div>
      <div className='fx-agent-summary-row'><strong>action</strong><span>{displayText(actionText(record), 'BLOCKED_BY_CONTRACT')}</span></div>
      <div className='fx-agent-summary-row'><strong>target_ids</strong><span>{listText(record?.target_ids || pick(record, 'target_id', 'target'))}</span></div>
      <div className='fx-agent-summary-row'><strong>agent_ids</strong><span>{listText(record?.agent_ids || pick(record, 'agent_id'))}</span></div>
      <div className='fx-agent-summary-row'><strong>package_id</strong><span>{packageText(record)}</span></div>
      <div className='fx-agent-summary-row'><strong>config_version</strong><span>{configText(record)}</span></div>
      <div className='fx-agent-summary-row'><strong>blocker</strong><span>{displayText(blockerText(record))}</span></div>
      <div className='fx-agent-summary-row'><strong>audit / evidence refs</strong><Tags items={refs.map(item => `${item.scope}.${item.key}`)} /></div>
      <RefTable refs={refs} />
    </div>
  )
}

function RefTable({ refs }) {
  if (!refs.length) return <Blocked>当前 task detail 没有可展示的安全 audit/evidence refs；真实执行仍保持阻断。</Blocked>
  return (
    <div className='fx-agent-table'>
      <table>
        <thead><tr><th>scope</th><th>safe ref key</th><th>脱敏值</th></tr></thead>
        <tbody>{refs.map(item => <tr key={`${item.scope}-${item.key}`}><td>{item.scope}</td><td>{item.key}</td><td><span className='fx-agent-muted'>{item.value}</span></td></tr>)}</tbody>
      </table>
    </div>
  )
}

export function AgentTaskRecords() {
  const [records, setRecords] = useState([])
  const [detail, setDetail] = useState(null)
  const [loading, setLoading] = useState(false)
  const [loadingDetail, setLoadingDetail] = useState('')
  const [error, setError] = useState('')
  const displayRecords = useMemo(() => normalizeRecords(records), [records])

  const load = () => {
    setLoading(true)
    setError('')
    setDetail(null)
    agentApi.listTasks()
      .then(value => setRecords(normalizeRecords(value)))
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setLoading(false))
  }

  const openDetail = id => {
    setLoadingDetail(id)
    setError('')
    setDetail(null)
    agentApi.getTask(id)
      .then(value => {
        if (isBlocked(value) && isVisibleAction(value)) setDetail(value)
        else setError('BLOCKED_BY_CONTRACT: task detail 不是 blocked 生命周期任务，已拒绝展示。')
      })
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setLoadingDetail(''))
  }

  useEffect(() => { load() }, [])

  return (
    <section className='fx-agent-panel'>
      <h3>blocked task 审计记录</h3>
      <p>{BLOCKED_NOTE}</p>
      <Summary records={displayRecords} loading={loading} onRefresh={load} />
      <ErrorBox>{error}</ErrorBox>
      <RecordsTable records={displayRecords} loadingDetail={loadingDetail} onSelect={openDetail} />
      <Detail record={detail} />
    </section>
  )
}
