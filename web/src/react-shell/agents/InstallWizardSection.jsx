import React, { useMemo, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { Blocked, ErrorBox, Field, Status } from './AgentShared.jsx'

const STEPS = [
  { key: 'host', label: '选择主机' },
  { key: 'package', label: '选择能力包' },
  { key: 'template', label: '配置模板' },
  { key: 'precheck', label: '安装前检查' },
  { key: 'execute', label: '执行' },
]

const METHODS = [
  { value: 'linux-curl', label: 'Linux curl', os: 'Linux' },
  { value: 'windows-cmd', label: 'Windows certutil', os: 'Windows' },
  { value: 'helm', label: 'K8s Helm', os: 'Kubernetes' },
]

export function InstallWizardSection({ agents, packages }) {
  const [step, setStep] = useState(0)
  const [selectedHosts, setSelectedHosts] = useState([])
  const [selectedPackage, setSelectedPackage] = useState('')
  const [method, setMethod] = useState('linux-curl')
  const [templateConfig, setTemplateConfig] = useState('')
  const [progress, setProgress] = useState(null)
  const [error, setError] = useState('')

  const hostList = useMemo(() => (agents || []).map(a => ({
    id: a.id || a.ident,
    label: `${a.hostname || a.ip || a.ident} (${a.os || '?'}/${a.arch || '?'})`,
  })), [agents])

  const packageList = useMemo(() => (packages || []).map(p => ({
    id: p.id,
    label: `${p.name} [${p.capability_domain || p.capabilityDomain || ''}]`,
  })), [packages])

  const currentMethod = METHODS.find(m => m.value === method) || METHODS[0]

  const canNext = () => {
    if (step === 0) return selectedHosts.length > 0
    if (step === 1) return selectedPackage !== ''
    if (step === 2) return true
    if (step === 3) return true
    return false
  }

  const handleNext = () => {
    if (step < STEPS.length - 1) setStep(step + 1)
  }
  const handlePrev = () => {
    if (step > 0) setStep(step - 1)
  }

  const handleExecute = () => {
    setError('')
    setProgress({ status: 'running', message: '正在下发安装任务...' })
    const targetIDs = selectedHosts.map(h => h.id || h)
    agentApi.createInstallPlan({
      package_id: selectedPackage,
      method,
      os: currentMethod.os,
      target_ids: targetIDs,
      mode: 'execute',
      metadata: { config_template: templateConfig },
    })
      .then(() => setProgress({ status: 'done', message: '安装任务已提交' }))
      .catch(err => {
        const msg = formatAgentError(err)
        if (msg.includes('pending')) {
          setProgress({ status: 'blocked', message: msg })
        } else {
          setError(msg)
          setProgress(null)
        }
      })
  }

  const toggleHost = (host) => {
    setSelectedHosts(prev =>
      prev.find(h => h.id === host.id)
        ? prev.filter(h => h.id !== host.id)
        : [...prev, host]
    )
  }

  return (
    <section className='fx-agent-work'>
      <div className='fx-agent-banner'>
        <div>
          <h3>安装向导</h3>
          <p>多步骤引导完成 Agent 安装：选择主机、能力包、配置模板、安装前检查、执行。</p>
        </div>
      </div>

      <WizardSteps steps={STEPS} current={step} />
      <ErrorBox>{error}</ErrorBox>

      {step === 0 && (
        <StepHosts hosts={hostList} selected={selectedHosts} onToggle={toggleHost} />
      )}
      {step === 1 && (
        <StepPackage packages={packageList} selected={selectedPackage} onSelect={setSelectedPackage} method={method} onMethodChange={setMethod} />
      )}
      {step === 2 && (
        <StepTemplate value={templateConfig} onChange={setTemplateConfig} />
      )}
      {step === 3 && (
        <StepPrecheck hosts={selectedHosts} packageId={selectedPackage} method={method} />
      )}
      {step === 4 && (
        <StepExecute progress={progress} onExecute={handleExecute} />
      )}

      <div className='fx-agent-actions'>
        {step > 0 && <button type='button' onClick={handlePrev}>上一步</button>}
        {step < STEPS.length - 1 && <button type='button' disabled={!canNext()} onClick={handleNext}>下一步</button>}
        {step === STEPS.length - 1 && !progress && <button type='button' onClick={handleExecute}>执行安装</button>}
      </div>
    </section>
  )
}

function WizardSteps({ steps, current }) {
  return (
    <div className='fx-agent-filter'>
      {steps.map((s, i) => (
        <span key={s.key} className={`fx-agent-tag ${i === current ? 'is-active' : ''} ${i < current ? 'is-done' : ''}`}>
          {i + 1}. {s.label}
        </span>
      ))}
    </div>
  )
}

function StepHosts({ hosts, selected, onToggle }) {
  return (
    <div className='fx-agent-panel'>
      <h4>选择目标主机</h4>
      {!hosts.length && <p className='fx-agent-muted'>暂无已注册主机</p>}
      <div className='fx-agent-package-grid'>
        {hosts.map(host => {
          const active = selected.find(h => h.id === host.id)
          return (
            <label key={host.id} className={`fx-agent-package ${active ? 'is-selected' : ''}`} style={{ cursor: 'pointer' }}>
              <input type='checkbox' checked={!!active} onChange={() => onToggle(host)} style={{ marginRight: 8 }} />
              {host.label}
            </label>
          )
        })}
      </div>
    </div>
  )
}

function StepPackage({ packages, selected, onSelect, method, onMethodChange }) {
  return (
    <div className='fx-agent-panel'>
      <h4>选择能力包和安装方式</h4>
      <div className='fx-agent-filter fx-agent-filter-3'>
        <Field label='能力包'>
          <select value={selected} onChange={e => onSelect(e.target.value)}>
            <option value=''>请选择</option>
            {packages.map(p => <option key={p.id} value={p.id}>{p.label}</option>)}
          </select>
        </Field>
        <Field label='安装方式'>
          <select value={method} onChange={e => onMethodChange(e.target.value)}>
            {METHODS.map(m => <option key={m.value} value={m.value}>{m.label}</option>)}
          </select>
        </Field>
      </div>
    </div>
  )
}

function StepTemplate({ value, onChange }) {
  return (
    <div className='fx-agent-panel'>
      <h4>配置模板</h4>
      <p className='fx-agent-muted'>可选：填写自定义配置模板 ID 或留空使用默认配置。</p>
      <Field label='模板 ID'>
        <input value={value} onChange={e => onChange(e.target.value)} placeholder='留空使用默认' />
      </Field>
    </div>
  )
}

function StepPrecheck({ hosts, packageId, method }) {
  return (
    <div className='fx-agent-panel'>
      <h4>安装前检查</h4>
      <div className='fx-agent-table'>
        <table>
          <thead><tr><th>检查项</th><th>状态</th></tr></thead>
          <tbody>
            <tr><td>目标主机数</td><td><Status ok={hosts.length > 0}>{hosts.length} 台</Status></td></tr>
            <tr><td>能力包</td><td><Status ok={!!packageId}>{packageId || '未选择'}</Status></td></tr>
            <tr><td>安装方式</td><td><Status ok>{method}</Status></td></tr>
            <tr><td>网络连通性</td><td><Status ok={false}>待验证</Status></td></tr>
            <tr><td>磁盘空间</td><td><Status ok={false}>待验证</Status></td></tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}

function StepExecute({ progress, onExecute }) {
  if (!progress) {
    return (
      <div className='fx-agent-panel'>
        <h4>准备执行</h4>
        <p>所有检查已完成，点击"执行安装"开始部署。</p>
      </div>
    )
  }
  return (
    <div className='fx-agent-panel'>
      <h4>执行进度</h4>
      <Status ok={progress.status === 'done'}>{progress.status}</Status>
      <p>{progress.message}</p>
      {progress.status === 'blocked' && <Blocked>{progress.message}</Blocked>}
    </div>
  )
}
