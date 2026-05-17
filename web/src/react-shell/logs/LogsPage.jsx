import React from 'react'
import { AggregateSection } from './AggregateSection.jsx'
import { ContextSection } from './ContextSection.jsx'
import { FieldsSection } from './FieldsSection.jsx'
import { LiveSection } from './LiveSection.jsx'
import { LogsIndexToFields } from './LogsIndexToFields.jsx'
import { PipelinesSection } from './PipelinesSection.jsx'
import { QuerySection } from './QuerySection.jsx'
import { SavedViewsSection } from './SavedViewsSection.jsx'
import { TraceLinkSection } from './TraceLinkSection.jsx'
import { sectionSet, sections } from './logsModel.js'
import { SectionTabs } from './LogsShared.jsx'
import './logs.css'

export function LogsPage({ query, onNavigate, onOpenTrace, onOpenAgent }) {
  const section = sectionSet.has(query?.section) ? query.section : 'query'
  const meta = sections.find(item => item.value === section) || sections[0]

  return (
    <main className='fx-logs-page'>
      <header className='fx-logs-head'>
        <div><p>日志中心</p><h1>{meta.label}</h1><span>{meta.desc}</span></div>
        <SectionTabs section={section} onNavigate={onNavigate} />
      </header>
      {section === 'query' && <QuerySection />}
      {section === 'live' && <LiveSection />}
      {section === 'fields' && <FieldsSection />}
      {section === 'indexes' && <LogsIndexToFields />}
      {section === 'context' && <ContextSection query={query} onOpenTrace={onOpenTrace} onOpenAgent={onOpenAgent} />}
      {section === 'aggregate' && <AggregateSection />}
      {section === 'pipelines' && <PipelinesSection />}
      {section === 'saved-views' && <SavedViewsSection />}
      {section === 'trace-link' && <TraceLinkSection query={query} onOpenTrace={onOpenTrace} onOpenAgent={onOpenAgent} />}
    </main>
  )
}
