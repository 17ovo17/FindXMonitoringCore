import React, { useState } from 'react'

/**
 * 列配置组件
 * 对齐夜莺 OrganizeColumns：可选择显示哪些列，存 localStorage
 */

const STORAGE_KEY = 'fx-alert-rules-columns'

const allColumns = [
  { key: 'status', label: '状态' },
  { key: 'name', label: '名称' },
  { key: 'businessGroup', label: '业务组' },
  { key: 'category', label: '分类' },
  { key: 'datasourceId', label: '数据源' },
  { key: 'severity', label: '级别' },
  { key: 'labels', label: '标签' },
  { key: 'updatedAt', label: '更新时间' },
  { key: 'updatedBy', label: '更新人' },
]

export function getVisibleColumns() {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) return JSON.parse(stored)
  } catch {}
  return allColumns.map((col) => col.key)
}

export function saveVisibleColumns(columns) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(columns))
  } catch {}
}

export function ColumnConfigModal({ onClose, onSave }) {
  const [visible, setVisible] = useState(() => getVisibleColumns())

  const toggle = (key) => {
    setVisible((prev) => prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key])
  }

  const handleSave = () => {
    saveVisibleColumns(visible)
    onSave?.(visible)
    onClose?.()
  }

  const resetDefault = () => {
    const all = allColumns.map((col) => col.key)
    setVisible(all)
  }

  return (
    <div className='fx-alert-modal'>
      <div className='fx-alert-modal__body is-confirm'>
        <header>
          <h2>列配置</h2>
          <button type='button' onClick={onClose}>关闭</button>
        </header>
        <div className='fx-alert-column-list'>
          {allColumns.map((col) => (
            <label key={col.key} className='fx-alert-check-inline'>
              <input
                type='checkbox'
                checked={visible.includes(col.key)}
                onChange={() => toggle(col.key)}
              />
              {col.label}
            </label>
          ))}
        </div>
        <div className='fx-alert-actions'>
          <button type='button' onClick={resetDefault}>恢复默认</button>
          <button type='button' className='is-primary' onClick={handleSave}>确认</button>
        </div>
      </div>
    </div>
  )
}

export { allColumns }
