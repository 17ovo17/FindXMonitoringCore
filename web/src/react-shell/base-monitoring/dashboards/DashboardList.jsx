import React, { useState } from 'react'

/**
 * 仪表盘列表 — 对齐夜莺 Table 结构
 * columns: checkbox | 名称(Link) | 分类标签(purple Tag) | 更新时间 | 公开 | 操作(Dropdown)
 * pagination: 共N条 + 页码 + N条/页
 */
export default function DashboardList({
  rows, selected, setSelected, onOpen, onRowAction, searchVal, onSearchAppendTag,
  visibleCols, page, pageSize, onPageChange, onPageSizeChange,
}) {
  const show = (key) => visibleCols.includes(key)
  const total = rows.length
  const totalPages = Math.ceil(total / pageSize) || 1
  const pagedRows = rows.slice((page - 1) * pageSize, page * pageSize)
  const allChecked = pagedRows.length > 0 && pagedRows.every((r) => selected.includes(r.id))
  const someChecked = pagedRows.some((r) => selected.includes(r.id)) && !allChecked

  const toggleAll = (checked) => {
    if (checked) {
      const ids = new Set([...selected, ...pagedRows.map((r) => r.id)])
      setSelected([...ids])
    } else {
      const pageIds = new Set(pagedRows.map((r) => r.id))
      setSelected(selected.filter((id) => !pageIds.has(id)))
    }
  }

  const toggleRow = (id, checked) => {
    setSelected(checked ? [...selected, id] : selected.filter((s) => s !== id))
  }

  return (
    <div className="fx-dash-table">
      <table>
        <thead>
          <tr>
            <th style={{ width: 40 }}>
              <input
                type="checkbox"
                checked={allChecked}
                ref={(el) => { if (el) el.indeterminate = someChecked }}
                onChange={(e) => toggleAll(e.target.checked)}
              />
            </th>
            {show('title') && <th>仪表盘名称</th>}
            {show('tags') && <th>分类标签</th>}
            {show('updatedAt') && <th>更新时间</th>}
            {show('shared') && <th>公开</th>}
            <th style={{ width: 80 }}>操作</th>
          </tr>
        </thead>
        <tbody>
          {pagedRows.map((row) => (
            <tr key={row.id}>
              <td>
                <input
                  type="checkbox"
                  checked={selected.includes(row.id)}
                  onChange={(e) => toggleRow(row.id, e.target.checked)}
                />
              </td>
              {show('title') && (
                <td>
                  <button type="button" className="is-link" onClick={() => onOpen(row.id)}>
                    {row.title}
                  </button>
                </td>
              )}
              {show('tags') && (
                <td>
                  {row.tags.length > 0
                    ? row.tags.map((tag) => (
                      <span
                        key={tag}
                        className="fx-dash-tag is-purple"
                        onClick={() => onSearchAppendTag?.(tag)}
                        style={{ cursor: 'pointer' }}
                      >
                        {tag}
                      </span>
                    ))
                    : <span className="muted">-</span>}
                </td>
              )}
              {show('updatedAt') && (
                <td>{row.updatedAt || <span className="muted">-</span>}</td>
              )}
              {show('shared') && (
                <td>
                  <span className={row.shared ? 'fx-dash-state is-on' : 'fx-dash-state'}>
                    {row.shared && <ShareIcon />}
                    {row.shareText}
                  </span>
                </td>
              )}
              <td>
                <RowActions row={row} onRowAction={onRowAction} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {total === 0 && <div className="fx-dash-empty">暂无仪表盘数据</div>}
      {total > 0 && (
        <div className="fx-dash-pagination">
          <span>共 {total} 条</span>
          <button type="button" disabled={page <= 1} onClick={() => onPageChange(page - 1)}>&lt;</button>
          {Array.from({ length: totalPages }, (_, i) => i + 1)
            .filter((p) => p === 1 || p === totalPages || Math.abs(p - page) <= 2)
            .map((p, idx, arr) => {
              const prev = arr[idx - 1]
              const gap = prev && p - prev > 1
              return (
                <React.Fragment key={p}>
                  {gap && <span>...</span>}
                  <button
                    type="button"
                    className={p === page ? 'is-active' : ''}
                    onClick={() => onPageChange(p)}
                  >
                    {p}
                  </button>
                </React.Fragment>
              )
            })}
          <button type="button" disabled={page >= totalPages} onClick={() => onPageChange(page + 1)}>&gt;</button>
          <select value={pageSize} onChange={(e) => onPageSizeChange(Number(e.target.value))}>
            <option value={10}>10 条/页</option>
            <option value={20}>20 条/页</option>
            <option value={50}>50 条/页</option>
            <option value={100}>100 条/页</option>
          </select>
        </div>
      )}
    </div>
  )
}

/** 操作列 — 三点图标 Dropdown */
function RowActions({ row, onRowAction }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="fx-panel-menu" onMouseLeave={() => setOpen(false)}>
      <button type="button" className="fx-panel-menu__trigger" onClick={() => setOpen(!open)}>⋮</button>
      {open && (
        <div className="fx-panel-menu__dropdown">
          <button onClick={() => { onRowAction('edit', row); setOpen(false) }}>编辑</button>
          <button onClick={() => { onRowAction('clone', row); setOpen(false) }}>克隆</button>
          <button onClick={() => { onRowAction('export', row); setOpen(false) }}>导出</button>
          <button className="fx-panel-menu__danger" onClick={() => { onRowAction('delete', row); setOpen(false) }}>删除</button>
        </div>
      )}
    </div>
  )
}

/** ShareAltOutlined 等效 SVG */
function ShareIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 1024 1024" fill="currentColor" style={{ marginRight: 4, verticalAlign: 'middle' }}>
      <path d="M752 664c-28.5 0-54.8 10-75.4 26.7L469.3 540.5a127.5 127.5 0 0 0 0-57l207.3-150.2A127.4 127.4 0 0 0 752 360c70.7 0 128-57.3 128-128s-57.3-128-128-128-128 57.3-128 128c0 9.8 1.1 19.4 3.2 28.5L419.9 410.7a127.9 127.9 0 0 0-147.8 0L64.8 260.5A128 128 0 0 0 272 232c0-70.7-57.3-128-128-128S16 161.3 16 232c0 47.2 25.6 88.4 63.7 110.7l207.3 150.2a127.5 127.5 0 0 0 0 57L79.7 700.1A127.9 127.9 0 0 0 16 792c0 70.7 57.3 128 128 128s128-57.3 128-128c0-9.8-1.1-19.4-3.2-28.5l207.3-150.2A127.9 127.9 0 0 0 752 664z" />
    </svg>
  )
}
