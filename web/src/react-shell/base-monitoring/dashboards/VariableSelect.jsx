import React, { useState, useRef, useEffect, useMemo } from 'react'

/**
 * VariableSelect — 对齐夜莺 InputGroupWithFormItem + Select 结构
 * 左侧标签 + 右侧下拉框，支持多选/单选/All/搜索/清除/maxTagCount
 */
export default function VariableSelect({
  name, label, options, value, multi, allOption, hide, errorMsg, onChange,
}) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const ref = useRef(null)
  const openValueRef = useRef(null)

  // 点击外部关闭
  useEffect(() => {
    const handler = (e) => {
      if (ref.current && !ref.current.contains(e.target)) {
        if (open) handleDropdownClose()
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open, value])

  const filtered = useMemo(() => {
    if (!search.trim()) return options || []
    const kw = search.trim().toLowerCase()
    return (options || []).filter((opt) => (opt.label || opt.value || '').toLowerCase().includes(kw))
  }, [options, search])

  const handleDropdownOpen = () => {
    openValueRef.current = value
    setOpen(true)
  }

  const handleDropdownClose = () => {
    setOpen(false)
    setSearch('')
    // 多选模式关闭时提交
    if (multi && !isEqual(openValueRef.current, value)) {
      onChange(name, value)
    }
  }

  const handleSelect = (optValue) => {
    if (multi) {
      if (optValue === 'all') {
        onChange(name, ['all'], true) // partial=true 不提交
      } else {
        const current = Array.isArray(value) ? value : []
        const next = current.filter((v) => v !== 'all')
        if (next.includes(optValue)) return // 不在 select 中去重
        onChange(name, [...next, optValue], true)
      }
    } else {
      setSearch('')
      setOpen(false)
      onChange(name, optValue)
    }
  }

  const handleDeselect = (optValue) => {
    if (!multi) return
    let next = []
    if (optValue === 'all') {
      next = []
    } else {
      next = (Array.isArray(value) ? value : []).filter((v) => v !== optValue)
    }
    onChange(name, next, true)
    // Tag 关闭按钮（dropdown 未打开时）直接提交
    if (!open) {
      setSearch('')
      onChange(name, next)
    }
  }

  const handleClear = (e) => {
    e.stopPropagation()
    if (multi) {
      onChange(name, [])
    } else {
      onChange(name, '')
    }
  }

  const selected = useMemo(() => {
    if (!value) return []
    return Array.isArray(value) ? value : [value]
  }, [value])

  const displayTags = useMemo(() => {
    if (!multi) return []
    const arr = Array.isArray(value) ? value : []
    return arr.slice(0, 3)
  }, [value, multi])

  const overflowCount = useMemo(() => {
    if (!multi) return 0
    const arr = Array.isArray(value) ? value : []
    return Math.max(0, arr.length - 3)
  }, [value, multi])

  if (hide) return null

  return (
    <div className="fx-var-item" ref={ref}>
      <div className="fx-var-item__label">
        {errorMsg && (
          <span className="fx-var-item__error" title={errorMsg}>&#9888;</span>
        )}
        <span>{label || name}</span>
      </div>
      <div className={`fx-var-item__select ${open ? 'is-open' : ''}`}>
        <div className="fx-var-item__trigger" onClick={open ? handleDropdownClose : handleDropdownOpen}>
          {multi && selected.length > 0 ? (
            <span className="fx-var-item__tags">
              {displayTags.map((v) => (
                <span key={v} className="fx-var-item__tag">
                  {v}
                  <span className="fx-var-item__tag-x" onClick={(e) => { e.stopPropagation(); handleDeselect(v) }}>×</span>
                </span>
              ))}
              {overflowCount > 0 && (
                <span className="fx-var-item__tag is-overflow" title={(Array.isArray(value) ? value.slice(3) : []).join(', ')}>
                  +{overflowCount}...
                </span>
              )}
            </span>
          ) : (
            <span className="fx-var-item__value">{!multi && value ? value : '请选择'}</span>
          )}
          {value && (multi ? selected.length > 0 : value !== '') && (
            <span className="fx-var-item__clear" onClick={handleClear}>×</span>
          )}
          <svg className="fx-var-item__arrow" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="6 9 12 15 18 9" /></svg>
        </div>
        {open && (
          <div className="fx-var-item__dropdown" style={{ minWidth: (options || []).length > 100 ? undefined : '180px' }}>
            <input
              className="fx-var-item__search"
              placeholder="搜索..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              autoFocus
            />
            <ul className="fx-var-item__options">
              {allOption && (
                <li key="all">
                  <button type="button" className={selected.includes('all') ? 'is-active' : ''} onClick={() => handleSelect('all')}>
                    {multi && <span className="fx-var-item__check">{selected.includes('all') ? '✓' : ''}</span>}
                    All
                  </button>
                </li>
              )}
              {filtered.map((opt) => {
                const isActive = selected.includes(opt.value)
                return (
                  <li key={opt.value}>
                    <button type="button" className={isActive ? 'is-active' : ''} onClick={() => isActive && multi ? handleDeselect(opt.value) : handleSelect(opt.value)}>
                      {multi && <span className="fx-var-item__check">{isActive ? '✓' : ''}</span>}
                      {opt.label}
                    </button>
                  </li>
                )
              })}
              {filtered.length === 0 && <li className="fx-var-item__empty">无匹配项</li>}
            </ul>
          </div>
        )}
      </div>
    </div>
  )
}

function isEqual(a, b) {
  if (a === b) return true
  if (Array.isArray(a) && Array.isArray(b)) {
    return a.length === b.length && a.every((v, i) => v === b[i])
  }
  return false
}
