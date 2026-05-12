import React from 'react'
import { AISRE_BLOCKERS } from '../api/aiSre.js'
import { Blocked, Empty, JsonPreview } from './AiSreShared.jsx'

const steps = ['证据确认', '处置计划', '审批', '执行', '回滚', '审计留痕']

export function RemediationSection({ onNavigate, evidence }) {
  return (
    <section className='fx-aisre-panel'>
      <div className='fx-aisre-toolbar'>
        <h2>自动修复</h2>
        <button type='button' onClick={() => onNavigate({ section: 'workflow' })}>查看工作流</button>
        <button type='button' onClick={() => onNavigate({ section: 'evidence' })}>查看证据链</button>
      </div>
      <Blocked>{AISRE_BLOCKERS.remediation}</Blocked>
      <div className='fx-aisre-cards'>
        {steps.map(step => (
          <article key={step} className='fx-aisre-card'>
            <h3>{step}</h3>
            <Empty>数据缺失：该阶段需要真实契约和审计证据。</Empty>
          </article>
        ))}
      </div>
      {!!evidence.length && <JsonPreview value={evidence.slice(0, 12)} />}
    </section>
  )
}
