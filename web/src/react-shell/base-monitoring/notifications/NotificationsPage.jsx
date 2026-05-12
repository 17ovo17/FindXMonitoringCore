import React, { useMemo } from 'react'
import { notificationSections } from './notificationModel.js'
import { NotificationChannelsSection } from './NotificationChannelsSection.jsx'
import { NotificationRulesSection } from './NotificationRulesSection.jsx'
import { NotificationTemplatesSection } from './NotificationTemplatesSection.jsx'
import './notifications.css'

const validSections = new Set(notificationSections.map((item) => item.value))

export function Modal({ title, children, onClose, narrow = false }) {
  return (
    <div className='fx-notify-modal'>
      <div className={narrow ? 'fx-notify-modal__body is-narrow' : 'fx-notify-modal__body'}>
        <header><h2>{title}</h2><button type='button' onClick={onClose}>关闭</button></header>
        {children}
      </div>
    </div>
  )
}

export function NotificationsPage({ query = {}, onNavigate }) {
  const section = validSections.has(query.section) ? query.section : 'rules'
  const current = useMemo(() => notificationSections.find((item) => item.value === section), [section])

  return (
    <main className='fx-notify-page'>
      <header className='fx-notify-head'>
        <div>
          <p>FindX 通知</p>
          <h1>通知工作台</h1>
        </div>
        <nav>
          {notificationSections.map((item) => (
            <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate?.({ section: item.value })}>
              {item.label}
            </button>
          ))}
        </nav>
      </header>
      <section className='fx-notify-title'><h2>{current?.label}</h2></section>
      {section === 'rules' && <NotificationRulesSection />}
      {section === 'channels' && <NotificationChannelsSection initialType={query.type} onTypeChange={(type) => onNavigate?.({ section: 'channels', type })} />}
      {section === 'templates' && <NotificationTemplatesSection />}
    </main>
  )
}
