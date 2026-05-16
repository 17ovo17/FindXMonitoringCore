import React, { useEffect, useMemo, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { displayText, fmtTime } from './agentModel.js'
import { Blocked, Empty, ErrorBox, Status, Tags } from './AgentShared.jsx'

const SENSITIVE_REF_RE = /(token|cookie|session|password|secret|dsn|private[_-]?key|credential|bearer|api[_-]?key|access[_-]?key)/i
const VISUAL_REF_RE = /(?:_ref$|^checksum$|^ssh_(?:host_key|fingerprint)$)/i
const BLOCKED_NOTE = 'install plan 只读审计面板仅展示被阻断的持久化记录，以及 safe refs、blocked reason 和 status；不会展示 credential、token 或其他敏感值。'

const isBlockedPlan = record => String(record?.status || '').toLowerCase() === 'blocked'
const normalizeRecords = rows => rows.filter(isBlockedPlan)
const recordId = record => record?.id || record?.record_id || ''
const pick = (record, ...keys) => keys.map(key => record?.[key]).find(value => value !== undefined && value !== null && value !== '')
const metadataOf = record => record?.metadata && typeof record.metadata === 'object' ? record.metadata : {}
const safeRefsOf = record => record?.safe_refs && typeof record.safe_refs === 'object' ? record.safe_refs : {}
const visibleRefKeys = record => {
  const keys = new Set([
    ...Object.keys(record || {}),
    ...Object.keys(metadataOf(record)),
    ...Object.keys(safeRefsOf(record)),
  ])
  return Array.from(keys).filter(key => VISUAL_REF_RE.test(key) && !SENSITIVE_REF_RE.test(key))
}
const refValue = (record, key) => record?.[key] ?? safeRefsOf(record)[key] ?? metadataOf(record)[key]
const hasRef = (record, key) => Boolean(String(refValue(record, key) ?? '').trim())
const blockerText = record => pick(record, 'blocker', 'blocked_reason', 'reason', 'message') || 'PENDING'
const statusText = record => isBlockedPlan(record) ? 'blocked' : 'PENDING'

const targetText = record => {
  const direct = pick(record, 'target', 'target_id', 'host_id', 'agent_id')
  const values = [
    direct,
    ...(Array.isArray(record?.target_ids) ? record.target_ids : []),
    ...(Array.isArray(record?.agent_ids) ? record.agent_ids : []),
  ].filter(Boolean)
  return values.length ? values.slice(0, 3).join(' / ') : '-'
}

const redactRef = value => {
  const text = String(value ?? '').trim()
  if (!text) return ''
  if (SENSITIVE_REF_RE.test(text)) return '<REDACTED_REF>'
  return text.length > 96 ? `${text.slice(0, 96)}...` : text
}

function Summary({ records, loading, onRefresh }) {
  return (
    <div className='fx-agent-toolbar'>
      <button type='button' onClick={onRefresh} disabled={loading}>{loading ? '刷新中...' : '刷新 blocked install plans'}</button>
      <span className='fx-agent-selected'>blocked install plans {records.length} 条</span>
      <span className='fx-agent-muted'>只读取 status=blocked 的持久化审计记录，非 blocked 响应会被前端丢弃。</span>
    </div>
  )
}

function RecordsTable({ records, loadingDetail, onSelect }) {
  if (!records.length) return <Empty>暂无 blocked install plans 审计记录；这不是安装成功，只表示当前没有持久化的阻断计划。</Empty>
  return (
    <div className='fx-agent-table'>
      <table>
        <thead><tr><th>package</th><th>os</th><th>method</th><th>target</th><th>status</th><th>blocked reason</th><th>updated_at</th><th>操作</th></tr></thead>
        <tbody>{records.map(record => <RecordRow key={recordId(record) || `${pick(record, 'package_id')}-${targetText(record)}`} record={record} loadingDetail={loadingDetail} onSelect={onSelect} />)}</tbody>
      </table>
    </div>
  )
}

function RecordRow({ record, loadingDetail, onSelect }) {
  const id = recordId(record)
  return (
    <tr>
      <td>{displayText(pick(record, 'package_id', 'package', 'package_name'))}</td>
      <td>{displayText(pick(record, 'os'))}</td>
      <td>{displayText(pick(record, 'method'))}</td>
      <td>{displayText(targetText(record))}</td>
      <td><Status ok={false}>{statusText(record)}</Status></td>
      <td><span className='fx-agent-muted'>{displayText(blockerText(record))}</span></td>
      <td>{fmtTime(pick(record, 'updated_at', 'updatedAt', 'created_at'))}</td>
      <td className='fx-agent-actions'><button type='button' disabled={!id || loadingDetail === id} onClick={() => onSelect(id)}>{loadingDetail === id ? '加载中...' : '详情'}</button></td>
    </tr>
  )
}

function Detail({ record }) {
  if (!record) return null
  const refs = visibleRefKeys(record)
  return (
    <div className='fx-agent-panel'>
      <h3>blocked install plan detail</h3>
      <p>{BLOCKED_NOTE}</p>
      <div className='fx-agent-summary-row'>
        <Status ok={false}>{statusText(record)}</Status>
        <span>record {displayText(recordId(record))}</span>
      </div>
      <div className='fx-agent-summary-row'>
        <strong>blocked reason</strong>
        <span>{displayText(blockerText(record))}</span>
      </div>
      {refs.length ? <RefTable record={record} refs={refs} /> : <Blocked>当前记录没有可见 safe refs，仍然保持 PENDING。</Blocked>}
    </div>
  )
}

function RefTable({ record, refs }) {
  return (
    <div className='fx-agent-table'>
      <table>
        <thead><tr><th>safe ref key</th><th>状态</th><th>脱敏值</th></tr></thead>
        <tbody>{refs.map(key => {
          const value = refValue(record, key)
          return (
            <tr key={key}>
              <td>{key}</td>
              <td><Status ok={hasRef(record, key)}>{value ? 'present' : 'missing'}</Status></td>
              <td><span className='fx-agent-muted'>{value ? redactRef(value) : 'PENDING'}</span></td>
            </tr>
          )
        })}</tbody>
      </table>
    </div>
  )
}

export function InstallPlanRecords() {
  const [records, setRecords] = useState([])
  const [detail, setDetail] = useState(null)
  const [loading, setLoading] = useState(false)
  const [loadingDetail, setLoadingDetail] = useState('')
  const [error, setError] = useState('')
  const displayRecords = useMemo(() => normalizeRecords(records), [records])

  const load = () => {
    setLoading(true)
    setError('')
    agentApi.listInstallPlans()
      .then(value => setRecords(normalizeRecords(value)))
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setLoading(false))
  }

  const openDetail = id => {
    setLoadingDetail(id)
    setError('')
    agentApi.getInstallPlan(id)
      .then(value => setDetail(isBlockedPlan(value) ? value : null))
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setLoadingDetail(''))
  }

  useEffect(() => { load() }, [])

  return (
    <section className='fx-agent-panel'>
      <h3>blocked install plans 审计记录</h3>
      <p>{BLOCKED_NOTE}</p>
      <Summary records={displayRecords} loading={loading} onRefresh={load} />
      <ErrorBox>{error}</ErrorBox>
      <RecordsTable records={displayRecords} loadingDetail={loadingDetail} onSelect={openDetail} />
      <Detail record={detail} />
    </section>
  )
}
