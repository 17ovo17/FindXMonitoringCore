import { useEffect, useCallback, useRef, useState } from 'react'

/**
 * DEGRADE-002: 离开确认 hook
 * - 浏览器关闭/刷新前弹出原生确认
 * - 提供 showLeaveConfirm 状态供路由拦截时显示自定义 Modal
 */
export function useBeforeUnload(hasChanges) {
  const [showLeaveConfirm, setShowLeaveConfirm] = useState(false)
  const pendingNavRef = useRef(null)

  useEffect(() => {
    if (!hasChanges) return
    const handler = (e) => {
      e.preventDefault()
      e.returnValue = '您有未保存的修改，确定要离开吗？'
      return e.returnValue
    }
    window.addEventListener('beforeunload', handler)
    return () => window.removeEventListener('beforeunload', handler)
  }, [hasChanges])

  const confirmLeave = useCallback((navigateFn) => {
    if (!hasChanges) {
      navigateFn()
      return
    }
    pendingNavRef.current = navigateFn
    setShowLeaveConfirm(true)
  }, [hasChanges])

  const handleDiscard = useCallback(() => {
    setShowLeaveConfirm(false)
    if (pendingNavRef.current) {
      pendingNavRef.current()
      pendingNavRef.current = null
    }
  }, [])

  const handleCancel = useCallback(() => {
    setShowLeaveConfirm(false)
    pendingNavRef.current = null
  }, [])

  return { showLeaveConfirm, confirmLeave, handleDiscard, handleCancel }
}

/**
 * DEGRADE-001: 保存模式 hook
 * - manual: 手动点保存
 * - auto: 修改即保存（debounce 2s）
 */
export function useSaveMode(onSave) {
  const [saveMode, setSaveMode] = useState(() => {
    try {
      return localStorage.getItem('fx-dash-save-mode') || 'manual'
    } catch { return 'manual' }
  })
  const timerRef = useRef(null)

  const toggleSaveMode = useCallback((mode) => {
    setSaveMode(mode)
    try { localStorage.setItem('fx-dash-save-mode', mode) } catch { /* ignore */ }
  }, [])

  const triggerAutoSave = useCallback(() => {
    if (saveMode !== 'auto') return
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => {
      onSave?.()
    }, 2000)
  }, [saveMode, onSave])

  useEffect(() => {
    return () => { if (timerRef.current) clearTimeout(timerRef.current) }
  }, [])

  return { saveMode, toggleSaveMode, triggerAutoSave }
}
