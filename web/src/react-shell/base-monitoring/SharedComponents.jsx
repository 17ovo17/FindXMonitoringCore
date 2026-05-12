import React from 'react'

export function Pagination({ total, page, pageSize, onPageChange, onPageSizeChange }) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const pages = []
  for (let i = 1; i <= totalPages; i++) {
    if (i === 1 || i === totalPages || (i >= page - 2 && i <= page + 2)) {
      pages.push(i)
    } else if (pages[pages.length - 1] !== '...') {
      pages.push('...')
    }
  }

  return (
    <div className='fx-pagination'>
      <span className='fx-pagination__total'>共 {total} 条</span>
      <div className='fx-pagination__pages'>
        <button type='button' disabled={page <= 1} onClick={() => onPageChange(page - 1)}>&lt;</button>
        {pages.map((p, i) => (
          p === '...'
            ? <span key={`ellipsis-${i}`} className='fx-pagination__ellipsis'>...</span>
            : <button key={p} type='button' className={p === page ? 'is-active' : ''} onClick={() => onPageChange(p)}>{p}</button>
        ))}
        <button type='button' disabled={page >= totalPages} onClick={() => onPageChange(page + 1)}>&gt;</button>
      </div>
      {onPageSizeChange && (
        <select className='fx-pagination__size' value={pageSize} onChange={(e) => onPageSizeChange(Number(e.target.value))}>
          <option value={10}>10 条/页</option>
          <option value={20}>20 条/页</option>
          <option value={50}>50 条/页</option>
          <option value={100}>100 条/页</option>
        </select>
      )}
    </div>
  )
}

export function ConfirmModal({ title, message, confirmText = '确定', cancelText = '取消', danger = false, onConfirm, onCancel }) {
  return (
    <div className='fx-modal-overlay' onClick={onCancel}>
      <div className='fx-modal fx-modal--sm' onClick={(e) => e.stopPropagation()}>
        <div className='fx-modal__header'>
          <h3>{title || '确认操作'}</h3>
          <button type='button' className='fx-modal__close' onClick={onCancel}>×</button>
        </div>
        <div className='fx-modal__body'>
          <p>{message}</p>
        </div>
        <div className='fx-modal__footer'>
          <button type='button' className='fx-btn' onClick={onCancel}>{cancelText}</button>
          <button type='button' className={`fx-btn ${danger ? 'is-danger' : 'is-primary'}`} onClick={onConfirm}>{confirmText}</button>
        </div>
      </div>
    </div>
  )
}

export function UnsavedChangesModal({ onDiscard, onSave, onCancel }) {
  return (
    <div className='fx-modal-overlay' onClick={onCancel}>
      <div className='fx-modal fx-modal--sm' onClick={(e) => e.stopPropagation()}>
        <div className='fx-modal__header'>
          <h3>未保存的修改</h3>
          <button type='button' className='fx-modal__close' onClick={onCancel}>×</button>
        </div>
        <div className='fx-modal__body'>
          <p>当前仪表盘有未保存的修改，是否保存？</p>
        </div>
        <div className='fx-modal__footer'>
          <button type='button' className='fx-btn' onClick={onCancel}>取消</button>
          <button type='button' className='fx-btn is-danger' onClick={onDiscard}>放弃修改</button>
          <button type='button' className='fx-btn is-primary' onClick={onSave}>保存</button>
        </div>
      </div>
    </div>
  )
}
