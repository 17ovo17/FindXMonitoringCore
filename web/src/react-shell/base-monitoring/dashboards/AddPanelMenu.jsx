import React, { useState, useRef, useEffect } from 'react'

const PANEL_TYPE_OPTIONS = [
  { key: 'timeseries', label: '时序图', icon: '📈' },
  { key: 'stat', label: '单值面板', icon: '🔢' },
  { key: 'table', label: '表格', icon: '📋' },
  { key: 'pie', label: '饼图', icon: '🥧' },
  { key: 'bar', label: '柱状图', icon: '📊' },
  { key: 'text', label: '文本', icon: '📝' },
]

/**
 * 添加图表按钮 + 类型菜单 (D03)
 * 点击后弹出下拉菜单，选择图表类型后回调 onSelect(type)
 */
export default function AddPanelMenu({ onSelect }) {
  const [open, setOpen] = useState(false)
  const ref = useRef(null)

  useEffect(() => {
    const handler = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const handleSelect = (type) => {
    setOpen(false)
    onSelect?.(type)
  }

  return (
    <div className="fx-add-panel-menu" ref={ref}>
      <button
        type="button"
        className="fx-add-panel-menu__btn is-primary"
        onClick={() => setOpen(!open)}
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
        添加图表
      </button>
      {open && (
        <div className="fx-add-panel-menu__dropdown">
          {PANEL_TYPE_OPTIONS.map((item) => (
            <button
              key={item.key}
              type="button"
              className="fx-add-panel-menu__item"
              onClick={() => handleSelect(item.key)}
            >
              <span className="fx-add-panel-menu__icon">{item.icon}</span>
              <span>{item.label}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
