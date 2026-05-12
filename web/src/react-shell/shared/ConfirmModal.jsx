import React from 'react'

/**
 * 通用确认弹窗（替代 window.confirm）
 * 支持两按钮（确认/取消）和三按钮（保存/放弃/取消）模式
 */
export function ConfirmModal({ title, message, onConfirm, onCancel, onDiscard, confirmText = '确定', cancelText = '取消', discardText, danger = false }) {
  return (
    <div className='fx-confirm-overlay'>
      <div className='fx-confirm-dialog'>
        <header>
          <h3>{title || '确认操作'}</h3>
          <button type='button' className='fx-confirm-close' onClick={onCancel} aria-label='关闭'>×</button>
        </header>
        <div className='fx-confirm-content'>
          <p>{message}</p>
        </div>
        <footer className='fx-confirm-footer'>
          <button type='button' onClick={onCancel}>{cancelText}</button>
          {onDiscard && <button type='button' className='fx-confirm-discard' onClick={onDiscard}>{discardText || '放弃修改'}</button>}
          <button type='button' className={danger ? 'fx-confirm-danger' : 'fx-confirm-primary'} onClick={onConfirm}>{confirmText}</button>
        </footer>
      </div>
    </div>
  )
}

/**
 * 通用分页组件（默认 20 条/页）
 */
export function Pagination({ total, page, pageSize = 20, onPageChange, onPageSizeChange }) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  if (total <= pageSize && !onPageSizeChange) return null

  const getPages = () => {
    const pages = []
    const max = 7
    if (totalPages <= max) {
      for (let i = 1; i <= totalPages; i++) pages.push(i)
    } else {
      pages.push(1)
      if (page > 3) pages.push('...')
      const start = Math.max(2, page - 1)
      const end = Math.min(totalPages - 1, page + 1)
      for (let i = start; i <= end; i++) pages.push(i)
      if (page < totalPages - 2) pages.push('...')
      pages.push(totalPages)
    }
    return pages
  }
  return (
    <div className='fx-shared-pagination'>
      <span>共 {total} 条</span>
      <div className='fx-shared-pagination__controls'>
        {onPageSizeChange && (
          <select value={pageSize} onChange={(e) => onPageSizeChange(Number(e.target.value))} aria-label='每页条数'>
            {[10, 20, 50].map((size) => <option key={size} value={size}>{size} 条/页</option>)}
          </select>
        )}
        <button type='button' disabled={page <= 1} onClick={() => onPageChange(page - 1)} aria-label='上一页'>&lt;</button>
        {getPages().map((p, idx) =>
          p === '...' ? <span key={`e-${idx}`}>...</span> : (
            <button key={p} type='button' className={p === page ? 'is-active' : ''} onClick={() => onPageChange(p)}>{p}</button>
          )
        )}
        <button type='button' disabled={page >= totalPages} onClick={() => onPageChange(page + 1)} aria-label='下一页'>&gt;</button>
      </div>
    </div>
  )
}

/**
 * useConfirm hook — 管理确认弹窗状态
 */
export function useConfirm() {
  const [state, setState] = React.useState(null)
  const confirm = ({ title, message, confirmText, danger }) => {
    return new Promise((resolve) => {
      setState({ title, message, confirmText, danger, resolve })
    })
  }
  const handleConfirm = () => { state?.resolve(true); setState(null) }
  const handleCancel = () => { state?.resolve(false); setState(null) }
  const modal = state ? (
    <ConfirmModal title={state.title} message={state.message} confirmText={state.confirmText} danger={state.danger} onConfirm={handleConfirm} onCancel={handleCancel} />
  ) : null
  return { confirm, modal }
}
