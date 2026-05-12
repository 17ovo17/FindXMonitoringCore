import React, { useMemo, useState } from 'react'
import { AGENT_BLOCKERS, agentApi, formatAgentError } from '../api/agents.js'
import { Blocked, Field } from './AgentShared.jsx'
import { InstallCommandPreview } from './InstallCommandPreview.jsx'

const normalizePackage = item => ({
  ...item,
  name: item.name || item.id,
  capabilityDomain: item.capabilityDomain || item.capability_domain || '未分组',
})

export function InstallSection({ packages }) {
  const fallbackPackage = { id: 'agent-core', name: 'FindX Agent 核心', capabilityDomain: '基础 Agent' }
  const packageRows = useMemo(() => (packages?.length ? packages : [fallbackPackage]).map(normalizePackage), [packages])
  const [blocked, setBlocked] = useState('')
  const [packageId, setPackageId] = useState('agent-core')
  const [method, setMethod] = useState('linux-curl')
  const [os, setOs] = useState('Linux')
  const selectedPackage = packageRows.find(item => item.id === packageId) || packageRows[0]

  const requestPlan = () => {
    agentApi.createInstallPlan({ package_id: selectedPackage.id, method, os, credential_ref: '<CREDENTIAL_REF>' })
      .then(() => setBlocked(AGENT_BLOCKERS.installLifecycle))
      .catch(err => setBlocked(formatAgentError(err)))
  }

  return (
    <section className='fx-agent-work'>
      <div className='fx-agent-banner'>
        <div>
          <h3>安装计划</h3>
          <p>本机、远程和容器化安装都通过 FindX Agent 能力包统一生成计划；真实脚本 URL、一次性 token、包签名和审计回执必须由 Agent Adapter 生成。</p>
        </div>
        <button type='button' onClick={requestPlan}>生成真实安装计划</button>
      </div>
      <div className='fx-agent-filter fx-agent-filter-3'>
        <Field label='能力包'>
          <select value={selectedPackage.id} onChange={event => setPackageId(event.target.value)}>
            {packageRows.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}
          </select>
        </Field>
        <Field label='系统'>
          <select value={os} onChange={event => setOs(event.target.value)}>
            <option>Linux</option>
            <option>Windows</option>
            <option>Kubernetes</option>
          </select>
        </Field>
        <Field label='方式'>
          <select value={method} onChange={event => setMethod(event.target.value)}>
            <option value='linux-curl'>curl -kfsSL</option>
            <option value='windows-cmd'>certutil</option>
            <option value='windows-powershell'>PowerShell</option>
            <option value='ssh'>SSH</option>
            <option value='winrm'>WinRM</option>
            <option value='helm'>Helm</option>
          </select>
        </Field>
      </div>
      {blocked && <Blocked>{blocked}</Blocked>}
      <InstallCommandPreview packageId={selectedPackage.id} selectedMethod={method} onSelectMethod={setMethod} />
      <div className='fx-agent-actions'><button type='button' onClick={requestPlan}>执行安装</button></div>
      <Blocked>{AGENT_BLOCKERS.installLifecycle}</Blocked>
    </section>
  )
}
