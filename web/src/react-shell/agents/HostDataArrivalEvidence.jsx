import React from 'react'
import { displayText, fmtTime } from './agentModel.js'
import { Status } from './AgentShared.jsx'

export const DATA_ARRIVAL_SIGNALS = [
  {
    kind: 'heartbeat',
    label: 'heartbeat / 心跳',
    acceptance: 'Agent 心跳 receiver evidence，不能替代指标、日志或链路数据',
    blocker: 'BLOCKED_BY_CONTRACT: heartbeat receiver evidence 缺失，且不能替代其他数据到达信号。',
  },
  {
    kind: 'metrics',
    label: 'metrics / 指标',
    acceptance: '主机、容器、进程和运行时指标 receiver evidence',
    blocker: 'BLOCKED_BY_CONTRACT: metrics 数据到达缺少 receiver evidence，heartbeat 不能替代 metrics。',
  },
  {
    kind: 'logs',
    label: 'logs / 日志',
    acceptance: '日志 receiver evidence，至少一条有效日志记录',
    blocker: 'BLOCKED_BY_CONTRACT: logs receiver evidence 缺失，未证明日志数据到达。',
  },
  {
    kind: 'tracing',
    label: 'tracing / 链路',
    acceptance: '链路 receiver evidence，至少一个 trace/span',
    blocker: 'BLOCKED_BY_CONTRACT: tracing receiver evidence 缺失，未证明 span/trace 数据到达。',
  },
  {
    kind: 'profiling',
    label: 'profiling / 性能分析',
    acceptance: '剖析任务、segment 和结果回执',
    blocker: 'BLOCKED_BY_CONTRACT: profiling 任务、segment 和结果回执契约未开放。',
  },
  {
    kind: 'inspection',
    label: 'inspection / 巡检',
    acceptance: '巡检执行回执、报告和 Evidence Chain',
    blocker: 'BLOCKED_BY_CONTRACT: inspection 执行回执、报告和 Evidence Chain 契约未开放。',
  },
  {
    kind: 'rum',
    label: 'RUM / 前端体验',
    acceptance: 'RUM 事件接收、会话关联和回执',
    blocker: 'BLOCKED_BY_CONTRACT: RUM 事件接收、会话关联和回执契约未开放。',
  },
  {
    kind: 'gateway_trace',
    label: 'gateway trace / 网关链路',
    acceptance: '网关入口/出口 trace 接收、关联和回执',
    blocker: 'BLOCKED_BY_CONTRACT: gateway trace 接收、入口出口关联和回执契约未开放。',
  },
]

const signalMap = Object.fromEntries(DATA_ARRIVAL_SIGNALS.map(item => [item.kind, item]))
const aliasMap = {
  metric: 'metrics',
  log: 'logs',
  trace: 'tracing',
  traces: 'tracing',
  gateway: 'gateway_trace',
  gateway_trace: 'gateway_trace',
  gatewaytrace: 'gateway_trace',
  gateway_tracing: 'gateway_trace',
}

const norm = value => String(value || '').trim().toLowerCase()
const safeCount = value => {
  const number = Number(value || 0)
  return Number.isFinite(number) ? number : 0
}
const newest = (left, right) => new Date(right || 0) > new Date(left || 0) ? right : left
const reportedWithEvidence = item => item?.status === 'reported' && safeCount(item?.evidence_count) > 0

export const normalizeDataArrivalKind = value => {
  const clean = norm(value).replace(/[-\s]+/g, '_')
  return signalMap[clean] ? clean : aliasMap[clean] || ''
}

export const buildDataArrivalRows = (rows = []) => {
  const byKind = new Map()
  ;(rows || []).forEach(item => {
    const kind = normalizeDataArrivalKind(item?.kind)
    if (!kind) return
    const signal = signalMap[kind]
    const ok = reportedWithEvidence(item)
    byKind.set(kind, {
      ...signal,
      status: ok ? 'reported' : 'blocked',
      count: safeCount(item?.evidence_count),
      lastSeen: item?.last_seen,
      blocker: ok ? '' : displayText(item?.blocker || signal.blocker),
      acceptance: signal.acceptance,
    })
  })
  return DATA_ARRIVAL_SIGNALS.map(signal => byKind.get(signal.kind) || {
    ...signal,
    status: 'blocked',
    count: 0,
    lastSeen: '',
    blocker: signal.blocker,
  })
}

export function HostDataArrivalEvidence({ evidenceRows = [], identityKeys = [], installed = false, capabilities = [] }) {
  const matrix = buildHostEvidenceMatrix(evidenceRows, identityKeys, installed, capabilities)
  return (
    <div className='fx-agent-provider-refs'>
      <ul>
        {matrix.map(item => (
          <li key={item.kind}>
            <span><Status ok={item.status === 'reported'}>{item.status === 'reported' ? 'reported' : 'blocked'}</Status> {item.label}</span>
            <code>{item.detail}</code>
          </li>
        ))}
      </ul>
    </div>
  )
}

function buildHostEvidenceMatrix(evidenceRows, identityKeys, installed, capabilities) {
  const keys = new Set((identityKeys || []).map(norm).filter(Boolean))
  const grouped = groupEvidenceBySignal(evidenceRows, keys)
  const capabilitySet = new Set((capabilities || []).map(value => normalizeDataArrivalKind(value)).filter(Boolean))
  return DATA_ARRIVAL_SIGNALS.map(signal => {
    const evidence = grouped[signal.kind]
    if (evidence?.reportedCount > 0) return reportedSignal(signal, evidence)
    const capabilityHint = capabilitySet.has(signal.kind) ? '仅有 Agent 能力线索，缺 receiver evidence。' : ''
    const installHint = installed ? capabilityHint : '未安装 FindX Agent，数据到达未验证。'
    return {
      ...signal,
      status: 'blocked',
      detail: [installHint, evidence?.blocker || signal.blocker].filter(Boolean).join(' '),
    }
  })
}

function groupEvidenceBySignal(evidenceRows, keys) {
  const grouped = {}
  if (!keys.size) return grouped
  ;(evidenceRows || []).forEach(item => {
    const kind = normalizeDataArrivalKind(item?.kind)
    if (!kind || !evidenceMatchesHost(item, keys)) return
    const current = grouped[kind] || { reportedCount: 0, lastSeen: '', refs: [], blocker: '' }
    const increment = item?.status === 'reported' ? Math.max(1, safeCount(item?.evidence_count) || 1) : 0
    current.reportedCount += increment
    current.lastSeen = newest(current.lastSeen, item?.updated_at || item?.created_at || item?.last_seen)
    current.blocker = current.blocker || item?.blocker || ''
    current.refs = [...new Set([...current.refs, ...(Array.isArray(item?.evidence_refs) ? item.evidence_refs : [])])]
    grouped[kind] = current
  })
  return grouped
}

function evidenceMatchesHost(item, keys) {
  return evidenceIdentityKeys(item).some(key => keys.has(key))
}

function evidenceIdentityKeys(item) {
  return [
    item?.target_id,
    item?.agent_id,
    item?.metadata?.target_id,
    item?.metadata?.agent_id,
    item?.metadata?.ip,
    item?.metadata?.host,
    item?.metadata?.hostname,
  ].map(norm).filter(Boolean)
}

function reportedSignal(signal, evidence) {
  const refs = evidence.refs.length ? ` · refs ${evidence.refs.slice(0, 2).map(displayText).join(', ')}` : ''
  return {
    ...signal,
    status: 'reported',
    detail: `证据 ${evidence.reportedCount} · 最近 ${fmtTime(evidence.lastSeen)}${refs}`,
  }
}
