import React, { useEffect, useMemo, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { displayText, fmtTime } from './agentModel.js'
import { Blocked, Empty, ErrorBox, Status, Tags } from './AgentShared.jsx'

const SENSITIVE_REF_RE = /(token|cookie|session|password|secret|dsn|private[_-]?key|credential|bearer|api[_-]?key|access[_-]?key)/i
const MAX_REF_LENGTH = 96
const BLOCKED_NOTE = 'install execution 只展示 status=blocked 的阻断审计记录，不表示真实安装成功或已执行完成。'

const STEP_LABELS = {
  preflight: '预检',
  download_script: '下载脚本',
  verify_package: '校验能力包',
  install_files: '安装文件',
  register_service: '注册服务',
  start_service: '启动服务',
  verify_heartbeat: '校验心跳',
  verify_metrics: '校验指标',
  rollback_or_cleanup: '回滚或清理',
  resolve_cluster: '解析集群',
  validate_namespace: '校验命名空间',
  verify_rbac: '校验 RBAC',
  render_manifest: '渲染清单',
  prepare_rollout: '准备下发',
  verify_data_arrival: '校验数据到达',
  capture_evidence: '采集证据',
}

const RUNNER_LABELS = {
  ssh: 'Linux / SSH',
  local: 'Linux / local',
  'linux-curl': 'Linux / curl',
  'windows-cmd': 'Windows / CMD',
  'windows-installer': 'Windows / Installer',
  'windows-powershell': 'Windows / PowerShell',
  powershell: 'Windows / PowerShell',
  helm: 'Kubernetes / Helm',
  kubernetes: 'Kubernetes / Kubernetes',
  'kubernetes-daemonset': 'Kubernetes / DaemonSet',
  'kubernetes-initcontainer': 'Kubernetes / InitContainer',
  'kubernetes-sidecar': 'Kubernetes / Sidecar',
}

const isBlocked = record => String(record?.status || '').toLowerCase() === 'blocked'
const normalizeRecords = rows => rows.filter(isBlocked)
const recordId = record => record?.id || record?.record_id || ''
const pick = (record, ...keys) => keys.map(key => record?.[key]).find(value => value !== undefined && value !== null && value !== '')
const safeObject = value => value && typeof value === 'object' ? value : {}
const metadataOf = record => safeObject(record?.metadata)
const safeList = value => Array.isArray(value) ? value.filter(Boolean) : []
const blockedReason = record => pick(record, 'blocked_reason', 'blocker', 'reason', 'error_summary', 'message') || 'PENDING'
const statusText = record => isBlocked(record) ? 'blocked' : 'PENDING'
const runnerText = record => {
  const raw = String(pick(record, 'runner', 'runner_name', 'execution_runner') || '').trim()
  if (!raw) return 'PENDING'
  const label = RUNNER_LABELS[raw.toLowerCase()]
  return label ? `${label} / ${raw}` : raw
}
const targetText = record => displayText(pick(record, 'target_id', 'target', 'host_id', 'agent_id'), 'PENDING')
const planText = record => displayText(pick(record, 'plan_id', 'plan', 'install_plan_id'), 'PENDING')
const evidenceRefs = record => safeList(pick(record, 'evidence_refs') || metadataOf(record).evidence_refs)
const safeText = value => {
  const text = String(value ?? '').trim()
  if (!text) return ''
  if (SENSITIVE_REF_RE.test(text)) return '<REDACTED_REF>'
  return text.length > MAX_REF_LENGTH ? `${text.slice(0, MAX_REF_LENGTH)}...` : text
}
const stepNameText = step => displayText(step?.name, 'PENDING')
const stepLabelText = step => {
  const name = String(step?.name || '').trim().toLowerCase()
  if (!name) return 'PENDING'
  const label = STEP_LABELS[name]
  return label ? `${label} / ${name}` : name
}
const stepStatusText = step => String(step?.status || '').toLowerCase() === 'blocked' ? 'blocked' : 'PENDING'
const stepReasonText = step => safeText(step?.error_summary || step?.blocked_reason || step?.reason || blockedReason(step))
const evidenceTags = record => evidenceRefs(record).map(item => safeText(item)).filter(Boolean)

function Summary({ records, loading, onRefresh }) {
  return (
    <div className='fx-agent-toolbar'>
      <button type='button' onClick={onRefresh} disabled={loading}>{loading ? '刷新中...' : '刷新 blocked install executions'}</button>
      <span className='fx-agent-selected'>blocked install executions {records.length} 条</span>
      <span className='fx-agent-muted'>仅展示 status=blocked；非 blocked 响应会在前端丢弃。</span>
    </div>
  )
}

function RecordsTable({ records, loadingDetail, onSelect }) {
  if (!records.length) return <Empty>暂无 blocked install execution 记录；这不是安装成功，只表示当前没有持久化的阻断执行审计。</Empty>
  return (
    <div className='fx-agent-table'>
      <table>
        <thead><tr><th>plan_id</th><th>target_id</th><th>runner</th><th>status</th><th>blocked reason</th><th>evidence_refs</th><th>updated_at</th><th>操作</th></tr></thead>
        <tbody>{records.map(record => <RecordRow key={recordId(record) || `${planText(record)}-${targetText(record)}-${runnerText(record)}`} record={record} loadingDetail={loadingDetail} onSelect={onSelect} />)}</tbody>
      </table>
    </div>
  )
}

function RecordRow({ record, loadingDetail, onSelect }) {
  const id = recordId(record)
  const refs = evidenceTags(record)
  return (
    <tr>
      <td>{planText(record)}</td>
      <td>{targetText(record)}</td>
      <td>{runnerText(record)}</td>
      <td><Status ok={false}>{statusText(record)}</Status></td>
      <td><span className='fx-agent-muted'>{displayText(blockedReason(record))}</span></td>
      <td>{refs.length ? <Tags items={refs} /> : <Blocked>PENDING: 缺少 evidence_refs。</Blocked>}</td>
      <td>{fmtTime(pick(record, 'updated_at', 'updatedAt', 'created_at'))}</td>
      <td className='fx-agent-actions'><button type='button' disabled={!id || loadingDetail === id} onClick={() => onSelect(id)}>{loadingDetail === id ? '加载中...' : '详情'}</button></td>
    </tr>
  )
}

function EvidenceTable({ record }) {
  const refs = evidenceTags(record)
  if (!refs.length) return <Blocked>PENDING: 当前记录没有可见 evidence_refs。</Blocked>
  return (
    <div className='fx-agent-table'>
      <table>
        <thead><tr><th>evidence_ref</th><th>状态</th></tr></thead>
        <tbody>{refs.map(item => <tr key={item}><td><span className='fx-agent-muted'>{item}</span></td><td><Status ok={false}>blocked</Status></td></tr>)}</tbody>
      </table>
    </div>
  )
}

function StepsTable({ record }) {
  const steps = Array.isArray(record?.steps) ? record.steps : []
  if (!steps.length) return <Blocked>PENDING: 当前记录没有可见 steps。</Blocked>
  return (
    <div className='fx-agent-table'>
      <table>
        <thead><tr><th>step name</th><th>step label</th><th>status</th><th>blocked reason</th><th>updated_at</th></tr></thead>
        <tbody>{steps.map((step, index) => (
          <tr key={`${stepNameText(step)}-${index}`}>
            <td>{stepNameText(step)}</td>
            <td><span className='fx-agent-muted'>{stepLabelText(step)}</span></td>
            <td><Status ok={false}>{stepStatusText(step)}</Status></td>
            <td><span className='fx-agent-muted'>{displayText(stepReasonText(step))}</span></td>
            <td>{fmtTime(step?.updated_at)}</td>
          </tr>
        ))}</tbody>
      </table>
    </div>
  )
}

function Detail({ record }) {
  if (!record) return null
  return (
    <div className='fx-agent-panel'>
      <h3>blocked install execution detail</h3>
      <p>{BLOCKED_NOTE}</p>
      <div className='fx-agent-summary-row'>
        <Status ok={false}>{statusText(record)}</Status>
        <span>execution {displayText(recordId(record))}</span>
      </div>
      <div className='fx-agent-summary-row'><strong>plan_id</strong><span>{planText(record)}</span></div>
      <div className='fx-agent-summary-row'><strong>target_id</strong><span>{targetText(record)}</span></div>
      <div className='fx-agent-summary-row'><strong>runner</strong><span>{runnerText(record)}</span></div>
      <div className='fx-agent-summary-row'><strong>blocked reason</strong><span>{displayText(blockedReason(record))}</span></div>
      <div className='fx-agent-summary-row'><strong>evidence_refs</strong><Tags items={evidenceTags(record)} /></div>
      <EvidenceTable record={record} />
      <StepsTable record={record} />
    </div>
  )
}

export function InstallExecutionRecords() {
  const [records, setRecords] = useState([])
  const [detail, setDetail] = useState(null)
  const [loading, setLoading] = useState(false)
  const [loadingDetail, setLoadingDetail] = useState('')
  const [error, setError] = useState('')
  const displayRecords = useMemo(() => normalizeRecords(records), [records])

  const load = () => {
    setLoading(true)
    setError('')
    agentApi.listInstallExecutions()
      .then(value => setRecords(normalizeRecords(value)))
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setLoading(false))
  }

  const openDetail = id => {
    setLoadingDetail(id)
    setError('')
    agentApi.getInstallExecution(id)
      .then(value => setDetail(isBlocked(value) ? value : null))
      .catch(err => setError(formatAgentError(err)))
      .finally(() => setLoadingDetail(''))
  }

  useEffect(() => { load() }, [])

  return (
    <section className='fx-agent-panel'>
      <h3>blocked install executions 审计记录</h3>
      <p>{BLOCKED_NOTE}</p>
      <Summary records={displayRecords} loading={loading} onRefresh={load} />
      <ErrorBox>{error}</ErrorBox>
      <RecordsTable records={displayRecords} loadingDetail={loadingDetail} onSelect={openDetail} />
      <Detail record={detail} />
    </section>
  )
}
