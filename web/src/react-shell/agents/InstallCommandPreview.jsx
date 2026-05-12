import React from 'react'
import { installCommands, renderInstallCommand, resolveFindxOrigin } from './agentTemplateModel.js'
import { Blocked, CopyBlock } from './AgentShared.jsx'

const BLOCKED_TEXT = '当前内置包仓库和安装环境仍是测试证据，生产安装执行必须等包签名、一次性 token、审计和回滚契约开放后才能解除阻断。'

const kubernetesPreviewIds = new Set(['helm', 'kubernetes-daemonset', 'kubernetes-operator', 'kubernetes-sidecar', 'kubernetes-initcontainer'])

const commandText = {
  'linux-curl': {
    label: 'Linux curl -kfsSL',
    desc: '使用当前站点安装器预览 Linux 本机安装命令，token 保持占位符。',
  },
  'windows-cmd': {
    label: 'Windows CMD certutil',
    desc: '使用当前站点安装器预览 Windows CMD 批处理安装命令，token 保持占位符。',
  },
  'windows-powershell': {
    label: 'PowerShell Invoke-WebRequest',
    desc: '使用当前站点安装器预览 PowerShell 脚本安装命令，token 保持占位符。',
  },
  helm: {
    label: 'Kubernetes Helm',
    desc: '预览 FindX Agent Helm 渲染命令，保留 <TOKEN>、<NAMESPACE>、<RELEASE_NAME> 等安全占位符。',
  },
  'kubernetes-daemonset': {
    label: 'Kubernetes DaemonSet',
    desc: '预览节点级 FindX Agent DaemonSet manifest；复制入口不代表集群执行。',
  },
  'kubernetes-operator': {
    label: 'Kubernetes Operator',
    desc: '预览 FindX Agent Operator 自定义资源；CRD、控制器、权限和回执契约未开放前保持阻断。',
  },
  'kubernetes-sidecar': {
    label: 'Kubernetes Sidecar',
    desc: '预览工作负载 sidecar patch；缺少执行器、凭据、回执和数据到达契约时保持阻断。',
  },
  'kubernetes-initcontainer': {
    label: 'Kubernetes InitContainer',
    desc: '预览初始化容器 patch；真实注入、升级和回滚仍由 BLOCKED_BY_CONTRACT 阻断。',
  },
}

export function InstallCommandPreview({ packageId, selectedMethod, onSelectMethod, origin = resolveFindxOrigin() }) {
  return (
    <div className='fx-agent-command-list'>
      {installCommands.map(item => {
        const text = commandText[item.id] || { label: item.label, desc: item.desc }
        return (
          <article key={item.id} className='fx-agent-panel'>
            <h3>{text.label}</h3>
            <p>{text.desc}</p>
            <p>能力包：<strong>{packageId || 'agent-core'}</strong>；安装器和包下载地址按当前站点 origin 渲染。</p>
            {selectedMethod === item.id ? <span className='fx-agent-tag'>当前计划方式</span> : null}
            {onSelectMethod && selectedMethod !== item.id ? <button type='button' onClick={() => onSelectMethod(item.id)}>设为计划方式</button> : null}
            {kubernetesPreviewIds.has(item.id) ? <span className='fx-agent-tag'>BLOCKED_BY_CONTRACT</span> : null}
            <CopyBlock>{renderInstallCommand(item, { packageId, origin })}</CopyBlock>
          </article>
        )
      })}
      <Blocked>{BLOCKED_TEXT}</Blocked>
    </div>
  )
}
