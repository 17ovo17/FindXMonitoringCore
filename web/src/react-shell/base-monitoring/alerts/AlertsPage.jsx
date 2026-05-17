import React, { useMemo } from 'react'
import { alertSections } from './alertModel.js'
import { AlertEventsSection } from './AlertEventsSection.jsx'
import { AlertRulesSection } from './AlertRulesSection.jsx'
import { AlertMuteSection } from './AlertMuteSection.jsx'
import { AlertSubscribeSection } from './AlertSubscribeSection.jsx'
import { AlertPipelineSection } from './AlertPipelineSection.jsx'
import './alerts.css'

const validSections = new Set(alertSections.map((item) => item.value))

export function AlertsPage({ query = {}, onNavigate }) {
  const section = validSections.has(query.section) ? query.section : 'events'
  const current = useMemo(() => alertSections.find((item) => item.value === section), [section])

  return (
    <main className='fx-alert-page'>
      <header className='fx-alert-head'>
        <div>
          <p>告警</p>
          <h1>告警工作台</h1>
        </div>
        <nav>
          {alertSections.map((item) => (
            <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate?.({ section: item.value })}>
              {item.label}
            </button>
          ))}
        </nav>
      </header>
      <section className='fx-alert-title'><h2>{current?.label}</h2></section>
      {section === 'rules' && <AlertRulesSection />}
      {section === 'events' && <AlertEventsSection historyMode={query.history === 'true'} />}
      {section === 'history-events' && <AlertEventsSection historyMode />}
      {section === 'mutes' && <AlertMuteSection />}
      {section === 'subscriptions' && <AlertSubscribeSection />}
      {section === 'event-pipelines' && <AlertPipelineSection />}
    </main>
  )
}
