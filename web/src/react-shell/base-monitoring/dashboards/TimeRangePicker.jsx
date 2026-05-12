import React, { useState, useEffect, useRef, useMemo } from 'react'
import {
  QUICK_RANGES, REFRESH_OPTIONS,
  isValidTimeStr, describeRange, getHistoryCache, setHistoryCache
} from './timeRangeUtils.js'

// 重新导出供外部使用
export { QUICK_RANGES, REFRESH_OPTIONS }

/* ===== URL 同步 ===== */
export function parseTimeRangeFromURL() {
  const params = new URLSearchParams(window.location.search)
  const from = params.get('from')
  const to = params.get('to')
  if (!from) return null
  const match = QUICK_RANGES.find((r) => r.start === from && r.end === (to || 'now'))
  return match || (from ? { start: from, end: to || 'now' } : null)
}

export function syncTimeRangeToURL(range) {
  if (!range) return
  const params = new URLSearchParams(window.location.search)
  params.set('from', range.start)
  params.set('to', range.end)
  const newUrl = `${window.location.pathname}?${params.toString()}${window.location.hash}`
  window.history.replaceState(null, '', newUrl)
}

/* ===== 时区选项 ===== */
const TIMEZONE_OPTIONS = [
  { key: 'local', label: 'Browser Time', abbr: 'Local', offset: null },
  { key: 'utc', label: 'UTC', abbr: 'UTC', offset: 0 },
  { key: 'utc+8', label: 'China, CST', abbr: 'UTC+08:00', offset: 8 },
  { key: 'utc+9', label: 'Japan, JST', abbr: 'UTC+09:00', offset: 9 },
  { key: 'utc-5', label: 'US Eastern, EST', abbr: 'UTC-05:00', offset: -5 },
  { key: 'utc-8', label: 'US Pacific, PST', abbr: 'UTC-08:00', offset: -8 },
  { key: 'utc+1', label: 'Central Europe, CET', abbr: 'UTC+01:00', offset: 1 },
  { key: 'utc+5.5', label: 'India, IST', abbr: 'UTC+05:30', offset: 5.5 },
]

/**
 * TimeRangePicker — 对齐夜莺 Popover 面板结构
 *
 * 结构：
 * Popover (trigger=click, placement=bottomRight):
 *   左侧: 开始/结束时间输入 + 日历 + 确定 + 历史
 *   右侧: 搜索 + 快捷选项列表
 *   底部: 时区选择
 *
 * Props:
 *   value: { start, end } — 当前时间范围
 *   onChange: (range) => void
 *   timezone: string
 *   onTimezoneChange: (tz) => void
 */
export default function TimeRangePicker({ value, onChange, timezone, onTimezoneChange }) {
  const [visible, setVisible] = useState(false)
  const [startVal, setStartVal] = useState('')
  const [endVal, setEndVal] = useState('')
  const [searchValue, setSearchValue] = useState('')
  const [startInvalid, setStartInvalid] = useState(false)
  const [endInvalid, setEndInvalid] = useState(false)
  const popoverRef = useRef(null)

  // 同步外部 value 到内部输入框
  useEffect(() => {
    if (value) {
      setStartVal(value.start || '')
      setEndVal(value.end || '')
    }
  }, [value?.start, value?.end, visible])

  // 点击外部关闭 Popover
  useEffect(() => {
    const handler = (e) => {
      if (popoverRef.current && !popoverRef.current.contains(e.target)) {
        setVisible(false)
      }
    }
    if (visible) document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [visible])

  const label = useMemo(() => describeRange(value), [value?.start, value?.end])
  const tzInfo = TIMEZONE_OPTIONS.find((t) => t.key === (timezone || 'local')) || TIMEZONE_OPTIONS[0]
  const historyCache = useMemo(() => getHistoryCache(), [visible])

  const handleConfirm = () => {
    if (!isValidTimeStr(startVal)) { setStartInvalid(true); return }
    if (!isValidTimeStr(endVal)) { setEndInvalid(true); return }
    const range = { start: startVal, end: endVal }
    onChange?.(range)
    setHistoryCache(range)
    syncTimeRangeToURL(range)
    setVisible(false)
  }

  const handleQuickSelect = (item) => {
    const range = { start: item.start, end: item.end }
    onChange?.(range)
    setHistoryCache(range)
    syncTimeRangeToURL(range)
    setVisible(false)
  }

  const handleHistorySelect = (range) => {
    onChange?.(range)
    syncTimeRangeToURL(range)
    setVisible(false)
  }

  const filteredRanges = QUICK_RANGES.filter((r) =>
    r.label.includes(searchValue) || r.start.includes(searchValue)
  )
  return (
    <div className="fx-trp" ref={popoverRef}>
      {/* 触发按钮 */}
      <button
        type="button"
        className="fx-trp__trigger"
        onClick={() => setVisible(!visible)}
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" />
          <polyline points="12 6 12 12 16 14" />
        </svg>
        <span>{label || '选择时间范围'}</span>
        {timezone && timezone !== 'local' && (
          <span className="fx-trp__tz-badge">{tzInfo.abbr}</span>
        )}
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          {visible
            ? <polyline points="18 15 12 9 6 15" />
            : <polyline points="6 9 12 15 18 9" />}
        </svg>
        {value && (
          <span
            className="fx-trp__clear"
            onClick={(e) => { e.stopPropagation(); onChange?.(null); setVisible(false) }}
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10" /><line x1="15" y1="9" x2="9" y2="15" /><line x1="9" y1="9" x2="15" y2="15" />
            </svg>
          </span>
        )}
      </button>

      {/* Popover 面板 */}
      {visible && (
        <div className="fx-trp__popover">
          <div className="fx-trp__body">
            {/* 左侧：绝对时间 */}
            <div className="fx-trp__left">
              <div className="fx-trp__field">
                <label>开始时间</label>
                <input
                  className={startInvalid ? 'is-error' : ''}
                  value={startVal}
                  placeholder="now-30m"
                  onChange={(e) => { setStartVal(e.target.value); setStartInvalid(!isValidTimeStr(e.target.value)) }}
                />
                {startInvalid && <span className="fx-trp__error">无效时间</span>}
              </div>
              <div className="fx-trp__field">
                <label>结束时间</label>
                <input
                  className={endInvalid ? 'is-error' : ''}
                  value={endVal}
                  placeholder="now"
                  onChange={(e) => { setEndVal(e.target.value); setEndInvalid(!isValidTimeStr(e.target.value)) }}
                />
                {endInvalid && <span className="fx-trp__error">无效时间</span>}
              </div>
              <button type="button" className="is-primary fx-trp__confirm" onClick={handleConfirm}>
                确定
              </button>
              {historyCache.length > 0 && (
                <div className="fx-trp__history">
                  <span className="fx-trp__history-title">最近使用的时间范围</span>
                  <ul>
                    {historyCache.map((r, i) => (
                      <li key={i} onClick={() => handleHistorySelect(r)}>
                        {describeRange(r)}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
            {/* 右侧：快捷选项 */}
            <div className="fx-trp__right">
              <div className="fx-trp__search">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
                </svg>
                <input
                  placeholder="搜索快捷选项"
                  value={searchValue}
                  onChange={(e) => setSearchValue(e.target.value)}
                />
              </div>
              <ul className="fx-trp__quick-list">
                {filteredRanges.map((item) => (
                  <li
                    key={item.start}
                    className={value?.start === item.start && value?.end === item.end ? 'is-active' : ''}
                    onClick={() => handleQuickSelect(item)}
                  >
                    {item.label}
                  </li>
                ))}
              </ul>
            </div>
          </div>
          {/* 底部：时区选择 */}
          <div className="fx-trp__footer">
            <div className="fx-trp__tz-picker">
              <span className="fx-trp__tz-label">{tzInfo.label}</span>
              <select
                value={timezone || 'local'}
                onChange={(e) => onTimezoneChange?.(e.target.value)}
              >
                {TIMEZONE_OPTIONS.map((tz) => (
                  <option key={tz.key} value={tz.key}>{tz.label} ({tz.abbr})</option>
                ))}
              </select>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

