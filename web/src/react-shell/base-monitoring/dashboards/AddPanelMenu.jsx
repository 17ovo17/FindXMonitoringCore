import React, { useState, useRef, useEffect } from 'react'

const PANEL_TYPE_OPTIONS = [
  { key: 'row', label: '分组', icon: '📁' },
  { key: '_divider', label: '', icon: '' },
  { key: 'timeseries', label: '时序图', icon: '📈' },
  { key: 'barchart', label: '柱状图', icon: '📊' },
  { key: 'stat', label: '指标值', icon: '🔢' },
  { key: 'tableNG', label: '表格 NG (Beta)', icon: '📋' },
  { key: 'table', label: '表格', icon: '📄' },
  { key: 'pie', label: '饼图', icon: '🥧' },
  { key: 'hexbin', label: '蜂窝图', icon: '⬡' },
  { key: 'barGauge', label: '排行榜', icon: '📶' },
  { key: 'text', label: '文本卡片', icon: '📝' },
  { key: 'gauge', label: '仪表图', icon: '⏱' },
  { key: 'heatmap', label: '色块图', icon: '🟧' },
  { key: 'iframe', label: '内嵌文档 (iframe)', icon: '🌐' },
]

/**
 * 添加图表按钮 + 类型菜单
 * 对齐夜莺：分组 + 分隔线 + 12 种图表类型 = 13 项
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
          {PANEL_TYPE_OPTIONS.map((item) => {
            if (item.key === '_divider') {
              return <div key="_divider" className="fx-add-panel-menu__divider" />
            }
            return (
              <button
                key={item.key}
                type="button"
                className="fx-add-panel-menu__item"
                onClick={() => handleSelect(item.key)}
              >
                <span className="fx-add-panel-menu__icon">{item.icon}</span>
                <span>{item.label}</span>
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
