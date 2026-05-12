import React from 'react'
import { displayText, visibleSections } from './tracingModel.js'

export const FINDX_AGENT_LINKAGE_BLOCKER = 'BLOCKED_BY_CONTRACT: 服务覆盖率、Trace 详情反查 FindX Agent 状态、拓扑节点探针覆盖缺少 APM-Agent Adapter 联动契约。'

export function Blocked({ children }) {
  return <div className='fx-tracing-blocked'><strong>BLOCKED_BY_CONTRACT</strong> {children}</div>
}

export function ErrorBox({ children }) {
  return children ? <div className='fx-tracing-error'>{children}</div> : null
}

export function Field({ label, children, className }) {
  const cls = className ? ('fx-tracing-field ' + className) : 'fx-tracing-field'
  return <label className={cls}><span>{label}</span>{children}</label>
}

export function SectionTabs({ section, onNavigate }) {
  return (
    <nav className='fx-tracing-tabs'>
      {visibleSections.map(item => (
        <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate({ section: item.value })}>
          {item.label}
        </button>
      ))}
    </nav>
  )
}

export function Status({ ok, children }) {
  return <span className={`fx-tracing-status ${ok ? 'is-ok' : ''}`}>{children}</span>
}

export function Empty({ children }) {
  return <div className='fx-tracing-empty'>{children}</div>
}

const cleanFilter = value => String(value || '').trim()

export function AgentLinkActions({ onNavigate, q, runtime, status, packageName = 'tracing' }) {
  const openHosts = next => {
    if (!onNavigate) return
    const query = { action: 'agent-hosts', section: 'hosts', package: packageName, ...next }
    Object.keys(query).forEach(key => {
      if (!cleanFilter(query[key])) delete query[key]
    })
    onNavigate(query)
  }
  const keyword = cleanFilter(q)
  const common = {
    q: keyword,
    runtime: cleanFilter(runtime),
    status: cleanFilter(status),
  }
  return (
    <div className='fx-tracing-agent-links'>
      <button type='button' onClick={() => openHosts(common)}>进入主机 FindX Agent</button>
      <button type='button' onClick={() => openHosts({ ...common, package: packageName })}>查看能力包探针覆盖</button>
    </div>
  )
}

export function AgentEvidenceNotice({ children }) {
  return (
    <div className='fx-tracing-agent-evidence'>
      <strong>FindX Agent 证据链</strong>
      <span>{children || '按服务、实例、端点或 Trace 上下文进入主机 Agent；缺少 serviceInstanceId -> process.agentId -> FindX Agent 的只读映射契约时保持 BLOCKED_BY_CONTRACT。'}</span>
    </div>
  )
}

export function Drawer({ title, payload, onClose }) {
  if (!payload) return null
  return (
    <div className='fx-tracing-modal'>
      <div className='fx-tracing-modal__body'>
        <header><h2>{title}</h2><button type='button' onClick={onClose}>关闭</button></header>
        <pre>{displayText(payload)}</pre>
      </div>
    </div>
  )
}

export function JsonPreview({ value }) {
  return <pre className='fx-tracing-json'>{displayText(JSON.stringify(value || {}, null, 2))}</pre>
}
