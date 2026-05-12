import React, { useEffect, useState } from 'react'
import { formatOrgError, orgApi } from '../api/org.js'
import { Pagination } from '../shared/ConfirmModal.jsx'
import { Blocked } from './OrgShared.jsx'

export function OrgAuditSection({ q }) {
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [source, setSource] = useState('')
  const [expanded, setExpanded] = useState(null)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const pageSize = 20
  const [filters, setFilters] = useState({ action: '', operator: '', target: '', start: '', end: '' })

  const load = async () => {
    try {
      const params = { q, page, limit: pageSize, ...Object.fromEntries(Object.entries(filters).filter(([, v]) => v)) }
      const result = await orgApi.audit(params)
      setRows(result.rows || [])
      setTotal(result.total || result.rows?.length || 0)
      setSource(result.source || '')
      setError('')
    } catch (err) {
      setError(formatOrgError(err))
    }
  }

  useEffect(() => { load() }, [q, page])

  const doFilter = () => { setPage(1); load() }

  const maskSensitive = (text) => {
    if (!text) return '-'
    const str = typeof text === 'string' ? text : JSON.stringify(text, null, 2)
    return str
      .replace(/(password|secret|token|api_key)\s*[:=]\s*["']?[^"',\s}]+/gi, '$1: ******')
      .replace(/(Authorization:\s*Bearer\s+)\S+/gi, '$1******')
  }

  return (
    <section className='fx-org-work'>
      <div className='fx-org-toolbar'>
        <button type='button' onClick={doFilter}>筛选</button>
        <button type='button' onClick={load}>刷新</button>
        <span>{source}</span>
      </div>
      <div className='fx-org-filter'>
        <select value={filters.action} onChange={(e) => setFilters((p) => ({ ...p, action: e.target.value }))}>
          <option value=''>全部操作</option>
          <option value='create'>创建</option>
          <option value='update'>更新</option>
          <option value='delete'>删除</option>
          <option value='login'>登录</option>
          <option value='config_change'>配置变更</option>
        </select>
        <input value={filters.operator} onChange={(e) => setFilters((p) => ({ ...p, operator: e.target.value }))} placeholder='操作人' />
        <input value={filters.target} onChange={(e) => setFilters((p) => ({ ...p, target: e.target.value }))} placeholder='对象' />
        <input type='datetime-local' value={filters.start} onChange={(e) => setFilters((p) => ({ ...p, start: e.target.value }))} />
        <input type='datetime-local' value={filters.end} onChange={(e) => setFilters((p) => ({ ...p, end: e.target.value }))} />
      </div>
      {error && <Blocked>{error}</Blocked>}
      <div className='fx-org-table'>
        <table>
          <thead><tr><th>时间</th><th>用户</th><th>操作</th><th>对象</th><th>IP</th><th>结果</th><th>详情</th></tr></thead>
          <tbody>
            {rows.map((row, idx) => (
              <React.Fragment key={row.id || idx}>
                <tr onClick={() => setExpanded(expanded === idx ? null : idx)} style={{ cursor: 'pointer' }}>
                  <td>{row.timestamp ? new Date(row.timestamp).toLocaleString('zh-CN') : row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}</td>
                  <td>{row.operator || row.username || '-'}</td>
                  <td>{row.action || '-'}</td>
                  <td>{row.target || row.resource || '-'}</td>
                  <td>{row.ip || row.client_ip || '-'}</td>
                  <td>{row.result || row.decision || '-'}</td>
                  <td>{expanded === idx ? '收起' : '展开'}</td>
                </tr>
                {expanded === idx && (
                  <tr className='fx-audit-detail-row'>
                    <td colSpan={7}>
                      <div className='fx-audit-detail'>
                        <h4>请求详情</h4>
                        <pre>{maskSensitive(row.request || row.description || row.detail)}</pre>
                        {row.response && (
                          <>
                            <h4>响应详情</h4>
                            <pre>{maskSensitive(row.response)}</pre>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
        {!rows.length && <div className='fx-org-empty'>暂无审计日志。</div>}
      </div>
      <Pagination total={total} page={page} pageSize={pageSize} onPageChange={setPage} />
    </section>
  )
}
