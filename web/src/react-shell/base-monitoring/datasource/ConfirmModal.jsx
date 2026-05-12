import React from 'react'

export function ConfirmModal({ visible, title, content, onConfirm, onCancel }) {
  if (!visible) return null

  return (
    <div className='fx-ds-confirm' role='dialog' aria-modal='true' aria-labelledby='fx-ds-confirm-title'>
      <div className='fx-ds-confirm__backdrop' onClick={onCancel} />
      <div className='fx-ds-confirm__panel'>
        <h3 id='fx-ds-confirm-title'>{title}</h3>
        <p>{content}</p>
        <div className='fx-ds-confirm__actions'>
          <button type='button' className='is-cancel' onClick={onCancel}>取消</button>
          <button type='button' className='is-delete' onClick={onConfirm}>确认</button>
        </div>
      </div>
    </div>
  )
}
