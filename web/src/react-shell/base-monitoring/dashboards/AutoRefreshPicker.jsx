import React, { useState, useRef, useEffect } from 'react'

const AUTO_REFRESH_OPTIONS = [
  { key: 'off', label: 'Off', ms: 0 },
  { key: '5s', label: '5s', ms: 5000 },
  { key: '10s', label: '10s', ms: 10000 },
  { key: '30s', label: '30s', ms: 30000 },
  { key: '1m', label: '1m', ms: 60000 },
  { key: '5m', label: '5m', ms: 300000 },
]

/**
 * 自动刷新下拉 (D04)
 * 选中后启动 setInterval 按间隔触发 onRefresh
 * 切换页面或选 Off 时清除 interval
 */
export default function AutoRefreshPicker({ onRefresh }) {
  const [activeKey, setActiveKey] = useState('off')
  const [open, setOpen] = useState(false)
  const ref = useRef(null)
  const timerRef = useRef(null)

  useEffect(() => {
    const handler = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  // 管理 interval
  useEffect(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current)
      timerRef.current = null
    }
    const opt = AUTO_REFRESH_OPTIONS.find((o) => o.key === activeKey)
    if (opt && opt.ms > 0) {
      timerRef.current = setInterval(() => onRefresh?.(), opt.ms)
    }
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current)
        timerRef.current = null
      }
    }
  }, [activeKey, onRefresh])

  const handleSelect = (key) => {
    setActiveKey(key)
    setOpen(false)
  }

  const activeLabel = AUTO_REFRESH_OPTIONS.find((o) => o.key === activeKey)?.label || 'Off'

  return (
    <div className="fx-auto-refresh" ref={ref}>
      <button
        type="button"
        className="fx-auto-refresh__trigger"
        onClick={() => setOpen(!open)}
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <polyline points="23 4 23 10 17 10" />
          <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10" />
        </svg>
        <span>{activeLabel}</span>
      </button>
      {open && (
        <div className="fx-auto-refresh__dropdown">
          {AUTO_REFRESH_OPTIONS.map((opt) => (
            <button
              key={opt.key}
              type="button"
              className={activeKey === opt.key ? 'is-active' : ''}
              onClick={() => handleSelect(opt.key)}
            >
              {opt.label}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
