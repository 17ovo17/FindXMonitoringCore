/**
 * 指标说明 Tab (T05)
 * 表格显示：指标名 + 类型 + 单位 + 说明
 * 支持搜索过滤 + 导入/导出
 */
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { templateApi } from './templateApi.js'
import { normalizePayloads, TypeEnum } from './templateModel.js'
import { Pagination } from '../../shared/ConfirmModal.jsx'

const PAGE_SIZE = 20

export function MetricsTab({ component }) {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [selectedIds, setSelectedIds] = useState(new Set())
  const debounceRef = useRef(null)

  const fetchData = useCallback(async (q = query) => {
    setLoading(true)
    try {
      const result = await templateApi.listPayloads({
        component_id: component.id,
        type: TypeEnum.metric,
        query: q || undefined,
      })
      setData(normalizePayloads(result))
    } catch {
      setData([])
    } finally {
      setLoading(false)
    }
  }, [component.id])

  useEffect(() => { fetchData() }, [fetchData])

  const handleSearch = (value) => {
    setQuery(value)
    clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => fetchData(value), 400)
  }

  // 解析 metric content 为表格行
  const parseMetricContent = (row) => {
    try {
      const parsed = typeof row.content === 'string' ? JSON.parse(row.content) : row.content
      if (Array.isArray(parsed)) return parsed
      if (parsed?.name) return [parsed]
      return [{ name: row.name, type: parsed?.type || '', unit: parsed?.unit || '', note: parsed?.note || row.note || '' }]
    } catch {
      return [{ name: row.name, type: '', unit: '', note: row.note || '' }]
    }
  }

  // 展开所有 metric 行
  const metricRows = data.flatMap((row) => {
    const metrics = parseMetricContent(row)
    return metrics.map((m, idx) => ({ ...m, _id: `${row.id}-${idx}`, _payloadId: row.id }))
  })

  const filteredRows = metricRows.filter((m) => {
    if (!query.trim()) return true
    const text = `${m.name} ${m.type} ${m.unit} ${m.note}`.toLowerCase()
    return text.includes(query.trim().toLowerCase())
  })

  const paged = filteredRows.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)
  const allSelected = paged.length > 0 && paged.every((r) => selectedIds.has(r._id))

  const toggleAll = (checked) => {
    const next = new Set(selectedIds)
    paged.forEach((r) => { checked ? next.add(r._id) : next.delete(r._id) })
    setSelectedIds(next)
  }

  const toggleRow = (id, checked) => {
    const next = new Set(selectedIds)
    checked ? next.add(id) : next.delete(id)
    setSelectedIds(next)
  }

  const handleExport = () => {
    const selected = filteredRows.filter((r) => selectedIds.has(r._id))
    const exportData = selected.length > 0 ? selected : filteredRows
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${component.ident}-metrics.json`
    link.click()
    URL.revokeObjectURL(url)
  }

  const handleImport = () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.json'
    input.onchange = async (e) => {
      const file = e.target.files?.[0]
      if (!file) return
      try {
        const text = await file.text()
        const items = JSON.parse(text)
        const payloads = (Array.isArray(items) ? items : [items]).map((item) => ({
          type: TypeEnum.metric,
          component_id: component.id,
          name: item.name || '导入指标',
          cate: item.cate || '',
          content: JSON.stringify(item),
        }))
        await templateApi.createPayloads(payloads)
        await fetchData()
      } catch {
        // 静默处理
      }
    }
    input.click()
  }

  return (
    <div>
      <div className='fx-tpl-toolbar'>
        <div className='fx-tpl-toolbar-left'>
          <input type='text' value={query} onChange={(e) => handleSearch(e.target.value)} placeholder='搜索指标名称...' />
        </div>
        <div className='fx-tpl-toolbar-right'>
          <button type='button' onClick={handleImport}>导入</button>
          <button type='button' onClick={handleExport}>导出</button>
        </div>
      </div>
      <div className='fx-tpl-table-wrap'>
        <table className='fx-tpl-table'>
          <thead>
            <tr>
              <th className='is-check'>
                <input type='checkbox' checked={allSelected} onChange={(e) => toggleAll(e.target.checked)} aria-label='全选' />
              </th>
              <th>指标名</th>
              <th>类型</th>
              <th>单位</th>
              <th>说明</th>
            </tr>
          </thead>
          <tbody>
            {paged.map((row) => (
              <tr key={row._id}>
                <td className='is-check'>
                  <input type='checkbox' checked={selectedIds.has(row._id)} onChange={(e) => toggleRow(row._id, e.target.checked)} />
                </td>
                <td style={{ fontFamily: 'monospace', fontSize: 12 }}>{row.name}</td>
                <td>{row.type || '-'}</td>
                <td>{row.unit || '-'}</td>
                <td>{row.note || '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {loading && <div className='fx-tpl-empty'>加载中...</div>}
        {!loading && filteredRows.length === 0 && <div className='fx-tpl-empty'>暂无指标数据</div>}
      </div>
      <Pagination total={filteredRows.length} page={page} pageSize={PAGE_SIZE} onPageChange={setPage} />
    </div>
  )
}
