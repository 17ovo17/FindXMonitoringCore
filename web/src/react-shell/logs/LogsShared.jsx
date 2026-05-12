import React from 'react'
import { sections } from './logsModel.js'

export function Blocked({ children }) {
  const text = typeof children === 'string' ? children.replace(/^BLOCKED_BY_CONTRACT:\s*/, '') : children
  return <div className='fx-logs-blocked'><strong>BLOCKED_BY_CONTRACT</strong> {text}</div>
}

export function Field({ label, children }) {
  return <label className='fx-logs-field'><span>{label}</span>{children}</label>
}

export function SectionTabs({ section, onNavigate }) {
  return (
    <nav className='fx-logs-tabs'>
      {sections.map(item => (
        <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate({ section: item.value })}>
          {item.label}
        </button>
      ))}
    </nav>
  )
}

export function Status({ ok, children }) {
  return <span className={`fx-logs-status ${ok ? 'is-ok' : ''}`}>{children}</span>
}

export function Empty({ children }) {
  return <div className='fx-logs-empty'>{children}</div>
}

export function JsonPreview({ value }) {
  return <pre className='fx-logs-json'>{JSON.stringify(value, null, 2)}</pre>
}
