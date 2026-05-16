import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { quickOptions } from '../navigation.js'

/**
 * 模糊匹配：将 query 拆为字符序列，检查 text 中是否按顺序包含这些字符。
 */
function fuzzyMatch(text, query) {
  const lower = text.toLowerCase()
  const q = query.toLowerCase()
  let idx = 0
  for (let i = 0; i < lower.length && idx < q.length; i++) {
    if (lower[i] === q[idx]) idx++
  }
  return idx === q.length
}

const defaultActions = [
  { id: 'action:create-alert-rule', label: '创建告警规则', section: 'actions', to: { path: '/alerts', query: { section: 'rules', action: 'create' } } },
  { id: 'action:create-dashboard', label: '创建仪表盘', section: 'actions', to: { path: '/dashboards', query: { section: 'list', action: 'create' } } },
  { id: 'action:create-mute', label: '创建告警屏蔽', section: 'actions', to: { path: '/alerts', query: { section: 'mutes', action: 'create' } } },
  { id: 'action:ai-diagnose', label: '发起 AI 诊断', section: 'actions', to: { path: '/aiops', query: { section: 'diagnosis' } } },
]

// PLACEHOLDER_APPEND

export function CommandPalette({ open, onClose, onNavigate }) {
  const [search, setSearch] = useState('')
  const [activeIndex, setActiveIndex] = useState(0)
  const inputRef = useRef(null)
  const listRef = useRef(null)

  const allItems = useMemo(() => {
    const navItems = quickOptions.map((opt) => ({
      id: `nav:${opt.value}`,
      label: opt.label,
      section: 'pages',
      to: opt.to,
    }))
    return [...navItems, ...defaultActions]
  }, [])

  const filtered = useMemo(() => {
    if (!search.trim()) return allItems.slice(0, 20)
    return allItems.filter((item) => fuzzyMatch(item.label, search))
  }, [search, allItems])

  useEffect(() => {
    setActiveIndex(0)
  }, [filtered])

  useEffect(() => {
    if (open) {
      setSearch('')
      setActiveIndex(0)
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open])

  const select = useCallback((item) => {
    if (item?.to) {
      onNavigate?.(item.to)
    }
    onClose?.()
  }, [onNavigate, onClose])

  const handleKeyDown = useCallback((e) => {
    if (e.key === 'Escape') {
      e.preventDefault()
      onClose?.()
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setActiveIndex((i) => Math.min(i + 1, filtered.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setActiveIndex((i) => Math.max(i - 1, 0))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      select(filtered[activeIndex])
    }
  }, [filtered, activeIndex, select, onClose])

  useEffect(() => {
    if (!open) return
    const handler = (e) => {
      if (e.key === 'Escape') {
        e.preventDefault()
        onClose?.()
      }
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [open, onClose])

  // 滚动活跃项到可见区域
  useEffect(() => {
    if (!listRef.current) return
    const active = listRef.current.querySelector('.fx-cmd-item.is-active')
    if (active) active.scrollIntoView({ block: 'nearest' })
  }, [activeIndex])

  if (!open) return null

  return (
    <div className="fx-cmd-overlay" role="presentation" onMouseDown={(e) => { if (e.target === e.currentTarget) onClose?.() }}>
      <div className="fx-cmd-palette" role="dialog" aria-modal="true" aria-label="命令面板">
        <div className="fx-cmd-input-wrap">
          <input
            ref={inputRef}
            className="fx-cmd-input"
            type="text"
            placeholder="搜索页面、操作..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={handleKeyDown}
            aria-label="搜索命令"
          />
        </div>
        <ul className="fx-cmd-list" ref={listRef} role="listbox">
          {filtered.length === 0 && (
            <li className="fx-cmd-empty">无匹配结果</li>
          )}
          {filtered.map((item, idx) => (
            <li
              key={item.id}
              className={`fx-cmd-item ${idx === activeIndex ? 'is-active' : ''}`}
              role="option"
              aria-selected={idx === activeIndex}
              onMouseEnter={() => setActiveIndex(idx)}
              onClick={() => select(item)}
            >
              <span className="fx-cmd-item-label">{item.label}</span>
              <span className="fx-cmd-item-section">{item.section === 'pages' ? '页面' : '操作'}</span>
            </li>
          ))}
        </ul>
        <div className="fx-cmd-footer">
          <kbd>↑↓</kbd> 导航 <kbd>Enter</kbd> 选择 <kbd>Esc</kbd> 关闭
        </div>
      </div>
    </div>
  )
}
