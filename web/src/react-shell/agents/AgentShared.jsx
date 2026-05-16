import React, { useRef, useState } from 'react'
import { sections } from './agentModel.js'

export function Blocked({ children }) {
  const text = typeof children === 'string' ? children.replace(/^PENDING:\s*/, '') : children
  return <div className='fx-agent-blocked'><strong>PENDING</strong> {text}</div>
}

export function ErrorBox({ children }) {
  return children ? <div className='fx-agent-error'>{children}</div> : null
}

export function Field({ label, children }) {
  return <label className='fx-agent-field'><span>{label}</span>{children}</label>
}

export function SectionTabs({ section, onNavigate }) {
  return (
    <nav className='fx-agent-tabs'>
      {sections.map(item => (
        <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate({ section: item.value })}>
          {item.label}
        </button>
      ))}
    </nav>
  )
}

export function Status({ ok, children }) {
  return <span className={`fx-agent-status ${ok ? 'is-ok' : ''}`}>{children}</span>
}

export function Tags({ items }) {
  const rows = Array.isArray(items) ? items.filter(Boolean) : String(items || '').split(/[\n,]/).map(item => item.trim()).filter(Boolean)
  return rows.length ? rows.map(item => <span key={item} className='fx-agent-tag'>{item}</span>) : <span className='fx-agent-muted'>-</span>
}

export function Empty({ children }) {
  return <div className='fx-agent-empty'>{children}</div>
}

export function CopyBlock({ children }) {
  const [copyState, setCopyState] = useState('')
  const timerRef = useRef(null)
  const text = typeof children === 'string' ? children : String(children ?? '')

  const setTimedState = (nextState) => {
    setCopyState(nextState)
    if (timerRef.current) {
      window.clearTimeout(timerRef.current)
    }
    timerRef.current = window.setTimeout(() => {
      setCopyState('')
      timerRef.current = null
    }, 1800)
  }

  const copyWithTextarea = () => {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.setAttribute('readonly', '')
    textarea.style.position = 'fixed'
    textarea.style.top = '-9999px'
    textarea.style.left = '-9999px'
    document.body.appendChild(textarea)
    textarea.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(textarea)
    if (!ok) {
      throw new Error('copy failed')
    }
  }

  const handleCopy = async () => {
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text)
      } else {
        copyWithTextarea()
      }
      setTimedState('已复制')
    } catch {
      setTimedState('复制失败')
    }
  }

  return (
    <div className='fx-agent-copyblock'>
      <div className='fx-agent-copybar'>
        <button type='button' onClick={handleCopy}>复制</button>
        {copyState ? <span className={copyState === '复制失败' ? 'is-error' : ''}>{copyState}</span> : null}
      </div>
      <pre className='fx-agent-code'>{children}</pre>
    </div>
  )
}
