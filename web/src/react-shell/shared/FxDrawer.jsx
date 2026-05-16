import React, { useCallback, useEffect, useRef, useState } from 'react'
import './drawer.css'

const WIDTH_MAP = { sm: 400, md: 600, lg: 800, xl: 1000, full: '100%' }

let drawerStack = []

export function FxDrawer({
  open,
  onClose,
  title,
  width = 'md',
  footer,
  maskClosable = true,
  destroyOnClose = false,
  dirty = false,
  urlSync = false,
  steps,
  currentStep = 0,
  onStepChange,
  children,
}) {
  const drawerRef = useRef(null)
  const idRef = useRef(`fx-drawer-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`)
  const [visible, setVisible] = useState(false)
  const [animating, setAnimating] = useState(false)
  const [showDirtyWarning, setShowDirtyWarning] = useState(false)
  const [stackIndex, setStackIndex] = useState(0)

  const resolvedWidth = WIDTH_MAP[width] || WIDTH_MAP.md
  const widthStyle = typeof resolvedWidth === 'number' ? `${resolvedWidth}px` : resolvedWidth

  const handleClose = useCallback(() => {
    if (dirty) {
      setShowDirtyWarning(true)
      return
    }
    onClose?.()
  }, [dirty, onClose])

  const confirmClose = useCallback(() => {
    setShowDirtyWarning(false)
    onClose?.()
  }, [onClose])

  useEffect(() => {
    if (open) {
      setVisible(true)
      requestAnimationFrame(() => setAnimating(true))
      drawerStack.push(idRef.current)
      setStackIndex(drawerStack.length - 1)
    } else {
      setAnimating(false)
      const timer = setTimeout(() => setVisible(false), 200)
      drawerStack = drawerStack.filter((id) => id !== idRef.current)
      return () => clearTimeout(timer)
    }
    return () => {
      drawerStack = drawerStack.filter((id) => id !== idRef.current)
    }
  }, [open])

  useEffect(() => {
    if (!open) return
    const onKeyDown = (e) => {
      if (e.key === 'Escape') handleClose()
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [open, handleClose])

  useEffect(() => {
    if (!urlSync) return
    if (open) {
      const url = new URL(window.location.href)
      url.searchParams.set('drawer', 'open')
      window.history.pushState({ drawer: idRef.current }, '', url.toString())
    }
    const onPopState = (e) => {
      if (!window.location.search.includes('drawer=open')) {
        onClose?.()
      }
    }
    window.addEventListener('popstate', onPopState)
    return () => window.removeEventListener('popstate', onPopState)
  }, [open, urlSync, onClose])

  useEffect(() => {
    const idx = drawerStack.indexOf(idRef.current)
    if (idx >= 0) setStackIndex(idx)
  })

  if (!visible && !open) return null
  if (destroyOnClose && !open && !visible) return null

  const isNested = stackIndex < drawerStack.length - 1
  const hasSteps = Array.isArray(steps) && steps.length > 0

  const stepContent = hasSteps ? steps[currentStep]?.content : null
  const stepTitle = hasSteps ? steps[currentStep]?.title : null

  return (
    <div
      className={`fx-drawer-overlay ${animating ? 'is-open' : ''}`}
      onMouseDown={(e) => {
        if (maskClosable && e.target === e.currentTarget) handleClose()
      }}
    >
      <aside
        ref={drawerRef}
        className={`fx-drawer-panel ${animating ? 'is-open' : ''} ${isNested ? 'is-nested' : ''}`}
        style={{ width: widthStyle }}
        role="dialog"
        aria-modal="true"
        aria-label={title || '抽屉'}
      >
        {hasSteps && (
          <div className="fx-drawer-steps">
            {steps.map((s, i) => (
              <div
                key={i}
                className={`fx-drawer-step ${i === currentStep ? 'is-active' : ''} ${i < currentStep ? 'is-done' : ''}`}
              >
                <span className="fx-drawer-step-num">{i + 1}</span>
                <span className="fx-drawer-step-title">{s.title}</span>
              </div>
            ))}
          </div>
        )}
        <header className="fx-drawer-header">
          <h3 className="fx-drawer-title">{stepTitle || title}</h3>
          <button type="button" className="fx-drawer-close" onClick={handleClose} aria-label="关闭">
            &times;
          </button>
        </header>
        <div className="fx-drawer-body">
          {hasSteps ? stepContent : children}
        </div>
        {(footer || hasSteps) && (
          <footer className="fx-drawer-footer">
            {hasSteps ? (
              <>
                {currentStep > 0 && (
                  <button type="button" className="fx-drawer-btn fx-drawer-btn--prev" onClick={() => onStepChange?.(currentStep - 1)}>
                    上一步
                  </button>
                )}
                {currentStep < steps.length - 1 && (
                  <button type="button" className="fx-drawer-btn fx-drawer-btn--next" onClick={() => onStepChange?.(currentStep + 1)}>
                    下一步
                  </button>
                )}
                {footer}
              </>
            ) : footer}
          </footer>
        )}
        {showDirtyWarning && (
          <div className="fx-drawer-dirty-overlay">
            <div className="fx-drawer-dirty-dialog">
              <p>有未保存的更改，确定关闭吗？</p>
              <div className="fx-drawer-dirty-actions">
                <button type="button" onClick={() => setShowDirtyWarning(false)}>取消</button>
                <button type="button" className="fx-drawer-btn--danger" onClick={confirmClose}>确定关闭</button>
              </div>
            </div>
          </div>
        )}
      </aside>
    </div>
  )
}
