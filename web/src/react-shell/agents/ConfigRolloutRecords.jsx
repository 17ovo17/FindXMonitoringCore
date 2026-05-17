import React, { useEffect, useMemo, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { displayText, fmtTime } from './agentModel.js'
import { Blocked, Empty, ErrorBox, Status, Tags } from './AgentShared.jsx'
import { REMOTE_MUTATION_REF_KEYS } from './HostProviderRefs.jsx'

const REF_KEYS = [...REMOTE_MUTATION_REF_KEYS, 'config_version_ref', 'audit_record_ref', 'blocked_record_detail_ref']
const PAYLOAD_KEYS = ['template_id', 'plugin_id', 'target_mode', 'rollout_strategy', 'provider_mode', 'config_format', 'config_version', 'config_snippet_ref', 'credential_ref', 'rollback_ref', 'change_ticket', 'audit_reason', 'cluster_ref', 'namespace_ref', 'workload_ref', 'config_map_ref', 'executor_ref']
const SCOPE_KEYS = ['target_mode', 'scope', 'business_group', 'namespace', 'workload', 'cluster', 'target_ids', 'agent_ids']
const SENSITIVE_REF_RE = /(token|cookie|session|password|secret|dsn|private[_-]?key|credential|bearer|api[_-]?key|access[_-]?key)/i
const MAX_REF_LENGTH = 96

const isBlocked = record => String(record?.status || '').toLowerCase() === 'blocked'
const normalizeRecords = rows => rows.filter(isBlocked)
const recordId = record => record?.id || record?.record_id || record?.rollout_id || ''
const pick = (record, ...keys) => keys.map(key => record?.[key]).find(Boolean)
const metadataOf = record => record?.metadata && typeof record.metadata === 'object' ? record.metadata : {}
const refValue = (record, key) => record?.[key] ?? record?.safe_refs?.[key] ?? metadataOf(record)[key]
const hasRef = (record, key) => Boolean(refValue(record, key))
const blockerText = record => pick(record, 'blocker', 'blocked_reason', 'reason', 'message') || ''
const statusText = record => isBlocked(record) ? 'blocked' : '待处理'
const arrayText = value => Array.isArray(value) ? value.slice(0, 5).join(' / ') : value
const safeValue = (record, key) => {
  const value = record?.[key] ?? metadataOf(record)[key] ?? record?.request_payload?.[key] ?? record?.payload?.[key]
  return redactRef(arrayText(value))
}
const scopeTags = record => SCOPE_KEYS.map(key => {
  const value = safeValue(record, key)
  return value ? `${key}: ${value}` : ''
}).filter(Boolean)

const targetText = record => {
  const direct = pick(record, 'target', 'target_id', 'agent_id', 'host_id')
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
  return text.length > MAX_REF_LENGTH ? `${text.slice(0, MAX_REF_LENGTH)}...` : text
}

function Summary({ records, loading, onRefresh }) {
  return (
    <div className='fx-agent-toolbar'>
      <button type='button' onClick={onRefresh} disabled={loading}>{loading ? '刷新中...' : '刷新 blocked 记录'}</button>
      <span className='fx-agent-selected'>blocked rollout 记录 {records.length} 条</span>
      <span className='fx-agent-muted'>仅展示 status=blocked；非 blocked 响应会在前端丢弃。</span>
    </div>
  )
}

function RecordsTable({ records, loadingDetail, onSelect }) {
  if (!records.length) return <Empty>暂无 blocked rollout 记录；这不是下发完成，只表示当前没有持久化的阻断审计记录。</Empty>
  return (
    <div className='fx-agent-table'>
      <table>
        <thead><tr><th>template</th><th>plugin</th><th>target</th><th>status</th><th>blocker</th><th>updated_at</th><th>操作</th></tr></thead>
        <tbody>{records.map(record => <RecordRow key={recordId(record) || `${pick(record, 'template_id')}-${targetText(record)}`} record={record} loadingDetail={loadingDetail} onSelect={onSelect} />)}</tbody>
      </table>
    </div>
  )
}

function RecordRow({ record, loadingDetail, onSelect }) {
  const id = recordId(record)
  return (
    <tr>
      <td>{displayText(pick(record, 'template', 'template_id', 'template_name'))}</td>
      <td>{displayText(pick(record, 'plugin', 'plugin_id', 'plugin_name'))}</td>
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
  const missing = REF_KEYS.filter(key => !hasRef(record, key))
  const scopes = scopeTags(record)
  return (
    <div className='fx-agent-summary-row'>
      <strong>rollout detail</strong>
      <span>record {displayText(recordId(record))}</span>
      <Tags items={[`template ${displayText(pick(record, 'template_id', 'template'))}`, `plugin ${displayText(pick(record, 'plugin_id', 'plugin'))}`, `target ${displayText(targetText(record))}`]} />
      <Tags items={scopes.length ? scopes : ['scope: 待确认']} />
      {missing.length ? <Blocked>缺少 contract refs：{missing.join(', ')}。配置/插件写入、reload、drift check、Evidence Chain 和 rollback 仍不可执行。</Blocked> : null}
      <RefTable record={record} />
      <PayloadPreview record={record} />
    </div>
  )
}

function RefTable({ record }) {
  return (
    <div className='fx-agent-table'>
      <table>
        <thead><tr><th>safe _ref key</th><th>状态</th><th>脱敏值</th></tr></thead>
        <tbody>{REF_KEYS.map(key => {
          const value = refValue(record, key)
          return <tr key={key}><td>{key}</td><td><Status ok={false}>{value ? '已提供 safe ref' : '缺失'}</Status></td><td><span className='fx-agent-muted'>{value ? redactRef(value) : '待确认'}</span></td></tr>
        })}</tbody>
      </table>
    </div>
  )
}

function PayloadPreview({ record }) {
  const rows = PAYLOAD_KEYS.map(key => ({ key, value: safeValue(record, key) || '' }))
  return (
    <div className='fx-agent-table'>
      <table>
        <thead><tr><th>blocked record field</th><th>脱敏值 / safe ref</th></tr></thead>
        <tbody>{rows.map(row => <tr key={row.key}><td>{row.key}</td><td><span className='fx-agent-muted'>{row.value}</span></td></tr>)}</tbody>
      </table>
    </div>
  )
}

export function ConfigRolloutRecords() {
  const [records, setRecords] = useState([])
  const [detail, setDetail] = useState(null)
  const [loading, setLoading] = useState(false)
  const [loadingDetail, setLoadingDetail] = useState('')
  const [error, setError] = useState('')
  const displayRecords = useMemo(() => normalizeRecords(records), [records])

  const load = () => {
    setLoading(true)
    setError('')
    agentApi.listConfigRollouts()
      .then(value => setRecords(normalizeRecords(value)))
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setLoading(false))
  }
  const openDetail = id => {
    setLoadingDetail(id)
    setError('')
    agentApi.getConfigRollout(id)
      .then(value => {
        if (isBlocked(value)) setDetail(value)
        else { setDetail(null); setError('record detail 不是 blocked 配置/插件记录，已拒绝展示。') }
      })
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setLoadingDetail(''))
  }

  useEffect(() => { load() }, [])

  return (
    <section className='fx-agent-panel'>
      <h3>配置/插件下发阻断记录</h3>
      <Summary records={displayRecords} loading={loading} onRefresh={load} />
      <ErrorBox>{error}</ErrorBox>
      <RecordsTable records={displayRecords} loadingDetail={loadingDetail} onSelect={openDetail} />
      <Detail record={detail} />
    </section>
  )
}
