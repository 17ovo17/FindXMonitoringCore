import React from 'react'

const PAGE_SIZE_OPTIONS = [10, 20, 50]

export function Pagination({ total, current, pageSize, onPageChange, onPageSizeChange }) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const getPageNumbers = () => {
    const pages = []
    const maxVisible = 7
    if (totalPages <= maxVisible) {
      for (let i = 1; i <= totalPages; i++) pages.push(i)
    } else {
      pages.push(1)
      if (current > 3) pages.push('...')
      const start = Math.max(2, current - 1)
      const end = Math.min(totalPages - 1, current + 1)
      for (let i = start; i <= end; i++) pages.push(i)
      if (current < totalPages - 2) pages.push('...')
      pages.push(totalPages)
    }
    return pages
  }

  return (
    <div className='fx-ds-pagination'>
      <span>共 {total} 条</span>
      <div className='fx-ds-pagination__controls'>
        <select
          value={pageSize}
          onChange={(e) => onPageSizeChange(Number(e.target.value))}
          aria-label='每页条数'
        >
          {PAGE_SIZE_OPTIONS.map((size) => (
            <option key={size} value={size}>{size} 条/页</option>
          ))}
        </select>
        <button
          type='button'
          disabled={current <= 1}
          onClick={() => onPageChange(current - 1)}
          aria-label='上一页'
        >
          &lt;
        </button>
        {getPageNumbers().map((page, idx) =>
          page === '...' ? (
            <span key={`ellipsis-${idx}`}>...</span>
          ) : (
            <button
              key={page}
              type='button'
              className={page === current ? 'is-active' : ''}
              onClick={() => onPageChange(page)}
            >
              {page}
            </button>
          )
        )}
        <button
          type='button'
          disabled={current >= totalPages}
          onClick={() => onPageChange(current + 1)}
          aria-label='下一页'
        >
          &gt;
        </button>
      </div>
    </div>
  )
}
