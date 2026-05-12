import React, { useState, useRef, useEffect } from 'react'

const AUTO_REFRESH_OPTIONS = [
  { key: 'off', label: 'Off', ms: 0 },
  { key: '5s', label: '5s', ms: 5000 },
  { key: '10s', label: '10s', ms: 10000 },
  { key: '30s', label: '30s', ms: 30000 },
  { key: '1m', label: '1m', ms: 60000 },
  { key: '5m', label: '5m', ms: 300000 },
  { key: '15m', label: '15m', ms: 900000 },
  { key: '30m', label: '30m', ms: 1800000 },
  { key: '1h', label: '1h', ms: 3600000 },
  { key: '2h', label: '2h', ms: 7200000 },
  { key: '1d', label: '1d', ms: 86400000 },
]

/**
 * 自动刷新下拉（对齐夜莺 11 种选项）
 * 选中后启动 setInterval 按间隔触发 onRefresh
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
        <span>{activeLabel}</span>
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <polyline points="6 9 12 15 18 9" />
        </svg>
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
