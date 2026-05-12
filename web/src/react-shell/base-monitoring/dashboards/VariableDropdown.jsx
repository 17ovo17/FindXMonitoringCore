import React, { useState, useRef, useEffect, useMemo } from 'react'

/**
 * 变量下拉选择器 (D02)
 * 支持搜索过滤、多选（tag 展示）、单选
 */
export default function VariableDropdown({ name, label, options, value, multi, onChange }) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const ref = useRef(null)

  useEffect(() => {
    const handler = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const selected = useMemo(() => {
    if (!value) return []
    if (Array.isArray(value)) return value
    return [value]
  }, [value])

  const filtered = useMemo(() => {
    if (!search.trim()) return options
    const keyword = search.trim().toLowerCase()
    return options.filter((opt) => {
      const text = typeof opt === 'string' ? opt : (opt.label || opt.value || '')
      return text.toLowerCase().includes(keyword)
    })
  }, [options, search])

  const getOptionValue = (opt) => typeof opt === 'string' ? opt : (opt.value || opt.label || '')
  const getOptionLabel = (opt) => typeof opt === 'string' ? opt : (opt.label || opt.value || '')

  const handleSelect = (optValue) => {
    if (multi) {
      const next = selected.includes(optValue)
        ? selected.filter((v) => v !== optValue)
        : [...selected, optValue]
      onChange(name, next)
    } else {
      onChange(name, optValue)
      setOpen(false)
    }
  }

  const handleRemoveTag = (e, tagValue) => {
    e.stopPropagation()
    onChange(name, selected.filter((v) => v !== tagValue))
  }

  const displayValue = multi
    ? (selected.length > 0 ? `${selected.length} 项已选` : '请选择')
    : (selected[0] || '请选择')

  return (
    <div className="fx-var-dropdown" ref={ref}>
      <span className="fx-var-dropdown__label">{label || name}</span>
      <button
        type="button"
        className="fx-var-dropdown__trigger"
        onClick={() => setOpen(!open)}
      >
        {multi && selected.length > 0 ? (
          <span className="fx-var-dropdown__tags">
            {selected.slice(0, 3).map((v) => (
              <span key={v} className="fx-var-dropdown__tag">
                {v}
                <span className="fx-var-dropdown__tag-x" onClick={(e) => handleRemoveTag(e, v)}>x</span>
              </span>
            ))}
            {selected.length > 3 && <span className="fx-var-dropdown__tag">+{selected.length - 3}</span>}
          </span>
        ) : (
          <span>{displayValue}</span>
        )}
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>
      {open && (
        <div className="fx-var-dropdown__panel">
          <input
            className="fx-var-dropdown__search"
            placeholder="搜索..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            autoFocus
          />
          <ul className="fx-var-dropdown__list">
            {filtered.map((opt) => {
              const optVal = getOptionValue(opt)
              const optLabel = getOptionLabel(opt)
              const isSelected = selected.includes(optVal)
              return (
                <li key={optVal}>
                  <button
                    type="button"
                    className={isSelected ? 'is-active' : ''}
                    onClick={() => handleSelect(optVal)}
                  >
                    {multi && <span className="fx-var-dropdown__check">{isSelected ? '✓' : ''}</span>}
                    {optLabel}
                  </button>
                </li>
              )
            })}
            {filtered.length === 0 && <li className="fx-var-dropdown__empty">无匹配项</li>}
          </ul>
        </div>
      )}
    </div>
  )
}
