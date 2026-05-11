import React, { useState, useRef, useEffect } from 'react'

const QUICK_RANGES = [
  { key: '5m', label: '近 5 分钟', seconds: 300 },
  { key: '15m', label: '近 15 分钟', seconds: 900 },
  { key: '30m', label: '近 30 分钟', seconds: 1800 },
  { key: '1h', label: '近 1 小时', seconds: 3600 },
  { key: '3h', label: '近 3 小时', seconds: 10800 },
  { key: '6h', label: '近 6 小时', seconds: 21600 },
  { key: '12h', label: '近 12 小时', seconds: 43200 },
  { key: '24h', label: '近 24 小时', seconds: 86400 },
  { key: '2d', label: '近 2 天', seconds: 172800 },
  { key: '7d', label: '近 7 天', seconds: 604800 },
]

const REFRESH_OPTIONS = [
  { key: 'off', label: '关闭' },
  { key: '10s', label: '10s', ms: 10000 },
  { key: '30s', label: '30s', ms: 30000 },
  { key: '1m', label: '1m', ms: 60000 },
  { key: '5m', label: '5m', ms: 300000 },
]

export default function TimeRangePicker({ rangeKey, refreshKey, onRangeChange, onRefreshChange }) {
  const [open, setOpen] = useState(false)
  const ref = useRef(null)

  useEffect(() => {
    const handler = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const activeRange = QUICK_RANGES.find((r) => r.key === rangeKey) || QUICK_RANGES[3]
  const activeRefresh = REFRESH_OPTIONS.find((r) => r.key === refreshKey) || REFRESH_OPTIONS[0]

  return (
    <div className="fx-time-picker" ref={ref}>
      <button type="button" className="fx-time-picker__trigger" onClick={() => setOpen(!open)}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" />
          <polyline points="12 6 12 12 16 14" />
        </svg>
        <span>{activeRange.label}</span>
        {activeRefresh.key !== 'off' && (
          <span className="fx-time-picker__refresh-badge">{activeRefresh.label}</span>
        )}
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>
      {open && (
        <div className="fx-time-picker__dropdown">
          <div className="fx-time-picker__section">
            <strong>时间范围</strong>
            <div className="fx-time-picker__grid">
              {QUICK_RANGES.map((r) => (
                <button
                  key={r.key}
                  type="button"
                  className={r.key === rangeKey ? 'is-active' : ''}
                  onClick={() => { onRangeChange(r.key); setOpen(false) }}
                >
                  {r.label}
                </button>
              ))}
            </div>
          </div>
          <div className="fx-time-picker__section">
            <strong>自动刷新</strong>
            <div className="fx-time-picker__grid">
              {REFRESH_OPTIONS.map((r) => (
                <button
                  key={r.key}
                  type="button"
                  className={r.key === refreshKey ? 'is-active' : ''}
                  onClick={() => { onRefreshChange(r.key); setOpen(false) }}
                >
                  {r.label}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export { QUICK_RANGES, REFRESH_OPTIONS }
