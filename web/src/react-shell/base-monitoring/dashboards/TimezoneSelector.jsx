import React, { useState } from 'react'

const TIMEZONE_OPTIONS = [
  { key: 'local', label: '本地时区', offset: null },
  { key: 'utc', label: 'UTC', offset: 0 },
  { key: 'utc+8', label: 'UTC+8 (中国)', offset: 8 },
  { key: 'utc+9', label: 'UTC+9 (日本)', offset: 9 },
  { key: 'utc-5', label: 'UTC-5 (美东)', offset: -5 },
  { key: 'utc-8', label: 'UTC-8 (美西)', offset: -8 },
  { key: 'utc+1', label: 'UTC+1 (中欧)', offset: 1 },
  { key: 'utc+5.5', label: 'UTC+5:30 (印度)', offset: 5.5 },
]

function getStoredTimezone() {
  try {
    return localStorage.getItem('fx-dash-timezone') || 'local'
  } catch { return 'local' }
}

/**
 * DEGRADE-007: 时区选择器
 * 在时间范围选择器旁添加时区下拉
 */
export default function TimezoneSelector({ value, onChange }) {
  const [open, setOpen] = useState(false)
  const current = TIMEZONE_OPTIONS.find((tz) => tz.key === (value || 'local')) || TIMEZONE_OPTIONS[0]

  const handleSelect = (tz) => {
    try { localStorage.setItem('fx-dash-timezone', tz.key) } catch { /* ignore */ }
    onChange?.(tz.key)
    setOpen(false)
  }

  return (
    <div className="fx-tz-selector">
      <button type="button" className="fx-tz-selector__trigger" onClick={() => setOpen(!open)}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" />
          <line x1="2" y1="12" x2="22" y2="12" />
          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
        </svg>
        <span>{current.label}</span>
      </button>
      {open && (
        <div className="fx-tz-selector__dropdown">
          {TIMEZONE_OPTIONS.map((tz) => (
            <button
              key={tz.key}
              type="button"
              className={tz.key === current.key ? 'is-active' : ''}
              onClick={() => handleSelect(tz)}
            >
              {tz.label}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

/**
 * 根据时区偏移计算时间范围
 */
export function applyTimezoneOffset(timeRange, timezoneKey) {
  if (!timezoneKey || timezoneKey === 'local') return timeRange
  const tz = TIMEZONE_OPTIONS.find((t) => t.key === timezoneKey)
  if (!tz || tz.offset === null) return timeRange
  const localOffset = new Date().getTimezoneOffset() / 60
  const diff = (tz.offset + localOffset) * 3600
  return { ...timeRange, start: timeRange.start + diff, end: timeRange.end + diff }
}

export { getStoredTimezone, TIMEZONE_OPTIONS }
