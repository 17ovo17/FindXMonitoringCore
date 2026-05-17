import React from 'react'
import { displayTags, rowText, sections } from './assetsModel.js'

export function Blocked({ children }) {
  return children ? <div className='fx-assets-blocked'>{children}</div> : null
}

export function ErrorBox({ children }) {
  return children ? <div className='fx-assets-error'>{children}</div> : null
}

export function Feedback({ children }) {
  return children ? <div className='fx-assets-feedback'>{children}</div> : null
}

export function Field({ label, children }) {
  return <label className='fx-assets-field'><span>{label}</span>{children}</label>
}

export function Modal({ title, children, onClose }) {
  return (
    <div className='fx-assets-modal'>
      <div className='fx-assets-modal__body'>
        <header>
          <h2>{title}</h2>
          <button className='fx-assets-button' type='button' onClick={onClose}>关闭</button>
        </header>
        {children}
      </div>
    </div>
  )
}

export function SectionTabs({ section, onNavigate }) {
  return (
    <nav className='fx-assets-tabs'>
      {sections.map(item => (
        <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate({ section: item.value })}>
          {item.label}
        </button>
      ))}
    </nav>
  )
}

export function Status({ ok, children }) {
  return <span className={`fx-assets-status ${ok ? 'is-ok' : ''}`}>{children}</span>
}

export function Tags({ items }) {
  const tags = displayTags(items)
  return tags.length ? tags.map(tag => <span key={tag} className='fx-assets-tag'>{tag}</span>) : <span className='fx-assets-muted'>-</span>
}

export function filterRows(rows, q) {
  const keyword = String(q || '').trim().toLowerCase()
  if (!keyword) return rows
  return rows.filter(row => rowText(row).includes(keyword))
}

export function statusLabel(status) {
  return { active: '启用', disabled: '停用', archived: '归档', online: '在线', offline: '离线' }[status] || status || '未知'
}
