import React from 'react'

/**
 * 自定义确认弹窗（替代 window.confirm）
 * 支持两按钮（确认/取消）和三按钮（保存/放弃/取消）模式
 */
export default function ConfirmModal({ title, message, onConfirm, onCancel, onDiscard, confirmText = '确定', cancelText = '取消', discardText, danger = false }) {
  return (
    <div className="fx-dash-modal">
      <div className="fx-dash-modal__body fx-confirm-modal">
        <header>
          <h2>{title || '确认操作'}</h2>
          <button type="button" onClick={onCancel}>x</button>
        </header>
        <div className="fx-confirm-modal__content">
          <p>{message}</p>
        </div>
        <footer className="fx-confirm-modal__footer">
          <button type="button" onClick={onCancel}>{cancelText}</button>
          {onDiscard && (
            <button type="button" className="fx-confirm-modal__discard" onClick={onDiscard}>
              {discardText || '放弃修改'}
            </button>
          )}
          <button
            type="button"
            className={danger ? 'fx-confirm-modal__danger' : 'is-primary'}
            onClick={onConfirm}
          >
            {confirmText}
          </button>
        </footer>
      </div>
    </div>
  )
}
