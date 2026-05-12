import React, { useState } from 'react'
import { sections, sectionSet } from './tracingModel.js'
import { AlarmsSection } from './AlarmsSection.jsx'
import { OverviewSection } from './OverviewSection.jsx'
import { ProfilingSection } from './ProfilingSection.jsx'
import { ServicesSection } from './ServicesSection.jsx'
import { SettingsSection } from './SettingsSection.jsx'
import { TopologySection } from './TopologySection.jsx'
import { TraceDetailSection } from './TraceDetailSection.jsx'
import { TracesSection } from './TracesSection.jsx'
import { Blocked, SectionTabs } from './TracingShared.jsx'
import { TRACING_BLOCKERS } from '../api/tracing.js'
import './tracing.css'

export function TracingPage({ query, params, onNavigate }) {
  const deepTraceId = params?.traceId || query?.traceId || ''
  const selected = deepTraceId ? 'trace-detail' : sectionSet.has(query?.section) ? query.section : 'overview'
  const meta = sections.find(item => item.value === selected) || sections[0]
  const [bannerVisible, setBannerVisible] = useState(true)
  const navigate = next => onNavigate({ ...query, ...next })

  return (
    <main className='fx-tracing-page'>
      <header className='fx-tracing-head'>
        <div><p>链路监控</p><h1>{meta.label}</h1><span>{meta.desc}</span></div>
        <SectionTabs section={selected} onNavigate={navigate} />
      </header>
      {bannerVisible && <section className='fx-tracing-banner'><Blocked>{TRACING_BLOCKERS.agentLinkage}</Blocked><button type='button' onClick={() => setBannerVisible(false)}>收起</button></section>}
      {selected === 'overview' && <OverviewSection onNavigate={navigate} />}
      {selected === 'services' && <ServicesSection onNavigate={navigate} />}
      {selected === 'topology' && <TopologySection query={query || {}} onNavigate={navigate} />}
      {selected === 'traces' && <TracesSection query={query || {}} onNavigate={navigate} />}
      {selected === 'trace-detail' && <TraceDetailSection traceId={deepTraceId} onNavigate={navigate} />}
      {selected === 'profiling' && <ProfilingSection query={query || {}} onNavigate={navigate} />}
      {selected === 'alarms' && <AlarmsSection />}
      {selected === 'settings' && <SettingsSection />}
    </main>
  )
}
