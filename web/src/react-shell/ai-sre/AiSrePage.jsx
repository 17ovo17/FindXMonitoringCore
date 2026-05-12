import React, { useCallback, useMemo, useState } from 'react'
import { DiagnosisSection } from './DiagnosisSection.jsx'
import { EvidenceSection } from './EvidenceSection.jsx'
import { HealthSection } from './HealthSection.jsx'
import { KnowledgeSection } from './KnowledgeSection.jsx'
import { RemediationSection } from './RemediationSection.jsx'
import { ReportSection } from './ReportSection.jsx'
import { WorkflowSection } from './WorkflowSection.jsx'
import { sections, sectionSet } from './aiSreModel.js'
import { Blocked, SectionTabs } from './AiSreShared.jsx'
import { AISRE_BLOCKERS } from '../api/aiSre.js'
import './aiSre.css'

export function AiSrePage({ query, onNavigate }) {
  const section = sectionSet.has(query?.section) ? query.section : 'diagnosis'
  const meta = sections.find(item => item.value === section) || sections[0]
  const [evidence, setEvidence] = useState([])
  const [bannerVisible, setBannerVisible] = useState(true)
  const navigate = next => onNavigate({ ...query, ...next })

  const addEvidence = useCallback(item => {
    setEvidence(prev => [{
      id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
      at: new Date().toISOString(),
      ...item,
    }, ...prev].slice(0, 80))
  }, [])

  const evidenceCount = useMemo(() => evidence.length, [evidence])

  return (
    <main className='fx-aisre-page'>
      <header className='fx-aisre-head'>
        <div>
          <p>AI SRE</p>
          <h1>{meta.label}</h1>
          <span>{meta.desc}</span>
        </div>
        <SectionTabs section={section} onNavigate={navigate} />
      </header>
      {bannerVisible && (
        <section className='fx-aisre-banner'>
          <Blocked>{AISRE_BLOCKERS.evidence}</Blocked>
          <button type='button' onClick={() => setBannerVisible(false)}>收起</button>
          <span>当前会话临时证据 {evidenceCount} 条</span>
        </section>
      )}
      {section === 'diagnosis' && <DiagnosisSection query={query || {}} onNavigate={navigate} addEvidence={addEvidence} />}
      {section === 'workflow' && <WorkflowSection query={query || {}} onNavigate={navigate} addEvidence={addEvidence} />}
      {section === 'health' && <HealthSection addEvidence={addEvidence} />}
      {section === 'report' && <ReportSection query={query || {}} onNavigate={navigate} addEvidence={addEvidence} />}
      {section === 'evidence' && <EvidenceSection evidence={evidence} onNavigate={navigate} />}
      {section === 'knowledge' && <KnowledgeSection query={query || {}} onNavigate={navigate} addEvidence={addEvidence} />}
      {section === 'remediation' && <RemediationSection onNavigate={navigate} evidence={evidence} />}
    </main>
  )
}
