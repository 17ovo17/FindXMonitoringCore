import React from 'react'
import { compactJson, displayText, sections } from './aiSreModel.js'

export function Blocked({ children }) {
  return <div className='fx-aisre-blocked'><strong>BLOCKED_BY_CONTRACT</strong> {children}</div>
}

export function ErrorBox({ children }) {
  return children ? <div className='fx-aisre-error'>{children}</div> : null
}

export function Empty({ children }) {
  return <div className='fx-aisre-empty'>{children || '数据缺失'}</div>
}

export function Field({ label, children }) {
  return <label className='fx-aisre-field'><span>{label}</span>{children}</label>
}

export function JsonPreview({ value }) {
  return <pre className='fx-aisre-json'>{compactJson(value)}</pre>
}

export function SectionTabs({ section, onNavigate }) {
  return (
    <nav className='fx-aisre-tabs'>
      {sections.map(item => (
        <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate({ section: item.value })}>
          {item.label}
        </button>
      ))}
    </nav>
  )
}

export function StatusPill({ children, tone = 'blocked' }) {
  return <span className={`fx-aisre-status is-${tone}`}>{children}</span>
}

export function TextPreview({ value }) {
  return <pre className='fx-aisre-text'>{displayText(value || '数据缺失')}</pre>
}
