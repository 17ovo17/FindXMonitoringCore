import React, { useState } from 'react'

/**
 * DEGRADE-006: DashboardLinks 组件
 * 从仪表盘 JSON 的 links 字段读取链接列表，渲染为可点击的链接图标/文字
 */
export default function DashboardLinks({ links = [] }) {
  const [expanded, setExpanded] = useState(false)

  if (!links || links.length === 0) return null

  return (
    <div className="fx-dash-links">
      <button
        type="button"
        className="fx-dash-links__toggle"
        onClick={() => setExpanded(!expanded)}
        title="仪表盘链接"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
          <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
        </svg>
        <span>{links.length}</span>
      </button>
      {expanded && (
        <div className="fx-dash-links__dropdown">
          {links.map((link, i) => (
            <a
              key={i}
              href={link.url || link.href || '#'}
              target={link.targetBlank ? '_blank' : '_self'}
              rel="noopener noreferrer"
              className="fx-dash-links__item"
            >
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                <polyline points="15 3 21 3 21 9" />
                <line x1="10" y1="14" x2="21" y2="3" />
              </svg>
              <span>{link.title || link.label || link.url || '链接'}</span>
              {link.tooltip && <small>{link.tooltip}</small>}
            </a>
          ))}
        </div>
      )}
    </div>
  )
}
