import React, { useState, useEffect, useRef } from 'react'
import { REFRESH_OPTIONS } from './timeRangeUtils.js'

/**
 * AutoRefreshPicker — 对齐夜莺 AutoRefresh 组件结构
 *
 * 结构：
 * div.fx-auto-refresh:
 *   ├── Tooltip > Button icon=SyncOutlined(旋转动画) onClick=立即刷新
 *   └── Dropdown overlay=Menu[Off/5s/10s/20s/30s/1m/2m/3m/5m/10m]:
 *       └── Button: "{当前间隔}" + UpOutlined/DownOutlined
 *
 * Props:
 *   onRefresh: () => void — 立即刷新回调
 *   localKey: string — localStorage 缓存 key
 */
export default function AutoRefreshPicker({ onRefresh, localKey = 'fx-auto-refresh-interval' }) {
  const cachedSeconds = (() => {
    try { return parseInt(localStorage.getItem(localKey), 10) || 0 } catch { return 0 }
  })()
  const [intervalSeconds, setIntervalSeconds] = useState(cachedSeconds)
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const intervalRef = useRef(null)
  const removedRef = useRef(false)
  const containerRef = useRef(null)

  // 定时刷新逻辑（对齐夜莺 loop 模式）
  useEffect(() => {
    let cancelled = false
    if (intervalRef.current) {
      clearTimeout(intervalRef.current)
      intervalRef.current = null
    }
    if (intervalSeconds > 0) {
      const loop = () => {
        if (removedRef.current || cancelled) return
        intervalRef.current = setTimeout(() => {
          if (removedRef.current || cancelled) return
          onRefresh?.()
          loop()
        }, intervalSeconds * 1000)
      }
      loop()
    }
    return () => {
      cancelled = true
      if (intervalRef.current) { clearTimeout(intervalRef.current); intervalRef.current = null }
    }
  }, [intervalSeconds, onRefresh])

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      removedRef.current = true
      if (intervalRef.current) { clearTimeout(intervalRef.current); intervalRef.current = null }
    }
  }, [])

  // 点击外部关闭下拉
  useEffect(() => {
    const handler = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setDropdownOpen(false)
      }
    }
    if (dropdownOpen) document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [dropdownOpen])

  const handleSelect = (seconds) => {
    setIntervalSeconds(seconds)
    try { localStorage.setItem(localKey, String(seconds)) } catch { /* ignore */ }
    setDropdownOpen(false)
  }

  const currentLabel = REFRESH_OPTIONS.find((r) => r.seconds === intervalSeconds)?.label || 'Off'
  const isSpinning = intervalSeconds > 0

  return (
    <div className="fx-auto-refresh" ref={containerRef}>
      {/* 立即刷新按钮（带旋转动画） */}
      <button
        type="button"
        className="fx-auto-refresh__sync-btn"
        title="立即刷新"
        onClick={() => onRefresh?.()}
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          className={isSpinning ? 'fx-spin' : ''}
        >
          <polyline points="23 4 23 10 17 10" />
          <polyline points="1 20 1 14 7 14" />
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
        </svg>
      </button>
      {/* 间隔选择下拉 */}
      <div className="fx-auto-refresh__dropdown-wrap">
        <button
          type="button"
          className="fx-auto-refresh__trigger"
          onClick={() => setDropdownOpen(!dropdownOpen)}
        >
          <span>{currentLabel}</span>
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            {dropdownOpen
              ? <polyline points="18 15 12 9 6 15" />
              : <polyline points="6 9 12 15 18 9" />}
          </svg>
        </button>
        {dropdownOpen && (
          <div className="fx-auto-refresh__dropdown">
            {REFRESH_OPTIONS.map((opt) => (
              <button
                key={opt.key}
                type="button"
                className={intervalSeconds === opt.seconds ? 'is-active' : ''}
                onClick={() => handleSelect(opt.seconds)}
              >
                {opt.label}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
