import React from 'react'

/**
 * DEGRADE-018: 仪表盘列表（含分页）
 */
export default function DashboardList({ rows, selected, setSelected, onOpen, onRowAction, visibleCols, page, pageSize, onPageChange, onPageSizeChange }) {
  const show = (key) => visibleCols.includes(key)
  const totalPages = Math.ceil(rows.length / pageSize) || 1
  const pagedRows = rows.slice((page - 1) * pageSize, page * pageSize)
  return (
    <div className='fx-dash-table'>
      <table>
        <thead>
          <tr>
            <th><input type='checkbox' checked={pagedRows.length > 0 && selected.length === pagedRows.length} onChange={(event) => setSelected(event.target.checked ? pagedRows.map((row) => row.id) : [])} /></th>
            {show('title') && <th>名称</th>}
            {show('tags') && <th>标签</th>}
            {show('description') && <th>说明</th>}
            {show('updatedAt') && <th>更新时间</th>}
            {show('updatedBy') && <th>更新人</th>}
            {show('shared') && <th>共享</th>}
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {pagedRows.map((row) => (
            <tr key={row.id}>
              <td><input type='checkbox' checked={selected.includes(row.id)} onChange={(event) => setSelected(event.target.checked ? [...selected, row.id] : selected.filter((id) => id !== row.id))} /></td>
              {show('title') && <td><button type='button' className='is-link' onClick={() => onOpen(row.id)}>{row.title}</button>{row.shared && <svg width="12" height="12" viewBox="0 0 24 24" fill="#1769ff" style={{marginLeft:4}}><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/></svg>}</td>}
              {show('tags') && <td>{row.tags.length ? row.tags.map((tag) => <span className='fx-dash-tag' key={tag}>{tag}</span>) : <span className='muted'>无</span>}</td>}
              {show('description') && <td>{row.description || <span className='muted'>无</span>}</td>}
              {show('updatedAt') && <td>{row.updatedAt || <span className='muted'>-</span>}</td>}
              {show('updatedBy') && <td>{row.updatedBy || <span className='muted'>-</span>}</td>}
              {show('shared') && <td><span className={row.shared ? 'fx-dash-state is-on' : 'fx-dash-state'}>{row.shareText}</span></td>}
              <td>
                <select onChange={(event) => { if (event.target.value) onRowAction(event.target.value, row); event.target.value = '' }}>
                  <option value=''>更多</option>
                  <option value='edit'>编辑</option>
                  <option value='clone'>克隆</option>
                  <option value='share'>公开</option>
                  <option value='export'>导出</option>
                  <option value='delete'>删除</option>
                </select>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {rows.length === 0 && <div className='fx-dash-empty'>暂无仪表盘数据</div>}
      {/* DEGRADE-018: 分页组件 */}
      {rows.length > 0 && (
        <div className='fx-dash-pagination'>
          <span>共 {rows.length} 条</span>
          <select value={pageSize} onChange={(e) => onPageSizeChange(Number(e.target.value))}>
            <option value={10}>10 条/页</option>
            <option value={20}>20 条/页</option>
            <option value={50}>50 条/页</option>
            <option value={100}>100 条/页</option>
          </select>
          <button type='button' disabled={page <= 1} onClick={() => onPageChange(page - 1)}>上一页</button>
          <span>{page} / {totalPages}</span>
          <button type='button' disabled={page >= totalPages} onClick={() => onPageChange(page + 1)}>下一页</button>
        </div>
      )}
    </div>
  )
}
