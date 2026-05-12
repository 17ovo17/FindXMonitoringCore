import React, { useState, useRef, useEffect } from 'react'

const QUICK_RANGES = [
  { key: '5m', label: '近 5 分钟', seconds: 300, from: 'now-5m' },
  { key: '15m', label: '近 15 分钟', seconds: 900, from: 'now-15m' },
  { key: '30m', label: '近 30 分钟', seconds: 1800, from: 'now-30m' },
  { key: '1h', label: '近 1 小时', seconds: 3600, from: 'now-1h' },
  { key: '3h', label: '近 3 小时', seconds: 10800, from: 'now-3h' },
  { key: '6h', label: '近 6 小时', seconds: 21600, from: 'now-6h' },
  { key: '12h', label: '近 12 小时', seconds: 43200, from: 'now-12h' },
  { key: '24h', label: '近 24 小时', seconds: 86400, from: 'now-24h' },
  { key: '2d', label: '近 2 天', seconds: 172800, from: 'now-2d' },
  { key: '7d', label: '近 7 天', seconds: 604800, from: 'now-7d' },
  { key: '30d', label: '近 30 天', seconds: 2592000, from: 'now-30d' },
]

/** 从 URL search params 中恢复时间范围 key */
export function parseTimeRangeFromURL() {
  const params = new URLSearchParams(window.location.search)
  const from = params.get('from')
  if (!from) return null
  const match = QUICK_RANGES.find((r) => r.from === from)
  return match ? match.key : null
}

/** 将时间范围同步到 URL params（不刷新页面） */
export function syncTimeRangeToURL(rangeKey) {
  const range = QUICK_RANGES.find((r) => r.key === rangeKey)
  if (!range) return
  const params = new URLSearchParams(window.location.search)
  params.set('from', range.from)
  params.set('to', 'now')
  const newUrl = `${window.location.pathname}?${params.toString()}${window.location.hash}`
  window.history.replaceState(null, '', newUrl)
}

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

  const handleRangeChange = (key) => {
    onRangeChange(key)
    syncTimeRangeToURL(key)
    setOpen(false)
  }

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
                  onClick={() => handleRangeChange(r.key)}
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
