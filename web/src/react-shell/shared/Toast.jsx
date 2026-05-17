import React, { useCallback, useEffect, useRef, useState } from 'react'

/**
 * 全局 Toast 通知系统
 * 类型：success / error / warning / info
 * 自动消失（3s），右上角堆叠显示
 * 通过全局函数调用：toast.success('保存成功')
 */

let toastId = 0
let addToastFn = null

const DURATION = 3000

const ICONS = {
  success: '✓',
  error: '✗',
  warning: '⚠',
  info: 'ℹ',
}

export const toast = {
  success(message) { addToastFn?.({ type: 'success', message }) },
  error(message) { addToastFn?.({ type: 'error', message }) },
  warning(message) { addToastFn?.({ type: 'warning', message }) },
  info(message) { addToastFn?.({ type: 'info', message }) },
}

function ToastItem({ item, onRemove }) {
  const [exiting, setExiting] = useState(false)
  const timerRef = useRef(null)

  useEffect(() => {
    timerRef.current = setTimeout(() => {
      setExiting(true)
      setTimeout(() => onRemove(item.id), 300)
    }, DURATION)
    return () => clearTimeout(timerRef.current)
  }, [item.id, onRemove])

  return (
    <div
      className={`fx-toast-item fx-toast-item--${item.type} ${exiting ? 'is-exiting' : ''}`}
      role="alert"
      aria-live="polite"
    >
      <span className="fx-toast-icon">{ICONS[item.type] || ICONS.info}</span>
      <span className="fx-toast-msg">{item.message}</span>
    </div>
  )
}

export function ToastContainer() {
  const [items, setItems] = useState([])

  const remove = useCallback((id) => {
    setItems((prev) => prev.filter((t) => t.id !== id))
  }, [])

  useEffect(() => {
    addToastFn = ({ type, message }) => {
      const id = ++toastId
      setItems((prev) => [...prev, { id, type, message }])
    }
    return () => { addToastFn = null }
  }, [])

  if (!items.length) return null

  return (
    <div className="fx-toast-container" aria-label="通知区域">
      {items.map((item) => (
        <ToastItem key={item.id} item={item} onRemove={remove} />
      ))}
    </div>
  )
}
