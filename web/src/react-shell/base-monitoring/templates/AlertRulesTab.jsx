/**
 * 告警规则模板 Tab (T07)
 * 列表 + 分类筛选 + 导入（支持数据源替换）+ 批量操作
 */
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { templateApi } from './templateApi.js'
import { normalizePayloads, TypeEnum, formatBeautifyJson, formatBeautifyJsons } from './templateModel.js'
import { useConfirm, Pagination } from '../../shared/ConfirmModal.jsx'
import { PayloadEditorModal } from './PayloadEditorModal.jsx'
import { AlertImportModal } from './AlertImportModal.jsx'

const PAGE_SIZE = 20

export function AlertRulesTab({ component }) {
  const [data, setData] = useState([])
  const [cates, setCates] = useState([])
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState('')
  const [cate, setCate] = useState('')
  const [page, setPage] = useState(1)
  const [selectedIds, setSelectedIds] = useState(new Set())
  const [importRows, setImportRows] = useState([])
  const [formState, setFormState] = useState(null)
  const [saving, setSaving] = useState(false)
  const { confirm, modal: confirmModal } = useConfirm()
  const debounceRef = useRef(null)
  const selectedRowsRef = useRef([])

  const fetchCates = useCallback(async () => {
    try {
      const result = await templateApi.listPayloadCates({
        component_id: component.id,
        type: TypeEnum.alert,
      })
      const list = Array.isArray(result) ? result : (result?.dat || [])
      setCates(list)
      if (list.length && !cate) setCate(list[0])
    } catch {
      setCates([])
    }
  }, [component.id])

  const fetchData = useCallback(async (q = query, c = cate) => {
    setLoading(true)
    try {
      const result = await templateApi.listPayloads({
        component_id: component.id,
        type: TypeEnum.alert,
        cate: c || undefined,
        query: q || undefined,
      })
      setData(normalizePayloads(result))
    } catch {
      setData([])
    } finally {
      setLoading(false)
    }
  }, [component.id, cate])

  useEffect(() => { fetchCates() }, [fetchCates])

  useEffect(() => {
    if (cate || !cates.length) fetchData(query, cate)
  }, [cate, fetchData])

  const handleSearch = (value) => {
    setQuery(value)
    clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => fetchData(value, cate), 400)
  }

  const paged = data.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)
  const allSelected = paged.length > 0 && paged.every((r) => selectedIds.has(r.id))

  const toggleAll = (checked) => {
    const next = new Set(selectedIds)
    paged.forEach((r) => { checked ? next.add(r.id) : next.delete(r.id) })
    setSelectedIds(next)
    selectedRowsRef.current = data.filter((r) => next.has(r.id))
  }

  const toggleRow = (id, checked) => {
    const next = new Set(selectedIds)
    checked ? next.add(id) : next.delete(id)
    setSelectedIds(next)
    selectedRowsRef.current = data.filter((r) => next.has(r.id))
  }

  const handleImportSingle = (row) => setImportRows([row])
  const handleBatchImport = () => {
    if (selectedRowsRef.current.length) setImportRows(selectedRowsRef.current)
  }

  const handleBatchExport = () => {
    const rows = selectedRowsRef.current
    if (!rows.length) return
    const blob = new Blob([formatBeautifyJsons(rows.map((r) => r.content))], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${component.ident}-alert-rules.json`
    link.click()
    URL.revokeObjectURL(url)
  }

  const handleBatchDelete = async () => {
    const rows = selectedRowsRef.current.filter((r) => r.updated_by !== 'system')
    if (!rows.length) return
    const ok = await confirm({ title: '批量删除', message: `确认删除 ${rows.length} 个告警规则模板？`, confirmText: '删除', danger: true })
    if (!ok) return
    setSaving(true)
    try {
      await templateApi.deletePayloads(rows.map((r) => r.id))
      setSelectedIds(new Set())
      selectedRowsRef.current = []
      await fetchData()
    } finally {
      setSaving(false)
    }
  }

  const openCreate = () => {
    setFormState({
      mode: 'create',
      draft: { name: '', cate: cate || '', type: TypeEnum.alert, component_id: component.id, content: '{\n  "expr": "",\n  "severity": "warning"\n}' },
      error: '',
    })
  }

  const openEdit = (row) => {
    setFormState({ mode: 'edit', draft: { ...row, content: formatBeautifyJson(row.content) }, error: '' })
  }

  const submitForm = async () => {
    if (!formState) return
    const { mode, draft } = formState
    if (!draft.name?.trim()) {
      setFormState({ ...formState, error: '名称必填' })
      return
    }
    setSaving(true)
    try {
      if (mode === 'create') {
        await templateApi.createPayloads([draft])
      } else {
        await templateApi.updatePayload(draft)
      }
      setFormState(null)
      await fetchData()
      await fetchCates()
    } catch (err) {
      setFormState({ ...formState, error: err.message || '保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (row) => {
    const ok = await confirm({ title: '删除模板', message: `确认删除「${row.name}」？`, confirmText: '删除', danger: true })
    if (!ok) return
    setSaving(true)
    try {
      await templateApi.deletePayloads([row.id])
      await fetchData()
      await fetchCates()
    } finally {
      setSaving(false)
    }
  }

  const parseTags = (tags) => String(tags || '').split(/\s+/).filter(Boolean)

  // 从 content 解析 severity
  const getSeverity = (row) => {
    try {
      const parsed = typeof row.content === 'string' ? JSON.parse(row.content) : row.content
      if (Array.isArray(parsed)) return parsed[0]?.severity || ''
      return parsed?.severity || ''
    } catch {
      return ''
    }
  }

  return (
    <div>
      <div className='fx-tpl-toolbar'>
        <div className='fx-tpl-toolbar-left'>
          <select value={cate} onChange={(e) => setCate(e.target.value)} style={{ width: 180 }}>
            <option value=''>全部分类</option>
            {cates.map((c) => <option key={c} value={c}>{c}</option>)}
          </select>
          <input type='text' value={query} onChange={(e) => handleSearch(e.target.value)} placeholder='搜索名称...' />
        </div>
        <div className='fx-tpl-toolbar-right'>
          <button type='button' className='is-primary' onClick={openCreate} disabled={saving}>新增</button>
          <button type='button' onClick={handleBatchImport}>导入到业务组</button>
          <button type='button' onClick={handleBatchExport}>批量导出</button>
          <button type='button' className='is-danger' onClick={handleBatchDelete} disabled={saving}>批量删除</button>
        </div>
      </div>
      <div className='fx-tpl-table-wrap'>
        <table className='fx-tpl-table'>
          <thead>
            <tr>
              <th className='is-check'>
                <input type='checkbox' checked={allSelected} onChange={(e) => toggleAll(e.target.checked)} aria-label='全选' />
              </th>
              <th>名称</th>
              <th>标签</th>
              <th>级别</th>
              <th>更新人</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {paged.map((row) => (
              <tr key={row.id}>
                <td className='is-check'>
                  <input type='checkbox' checked={selectedIds.has(row.id)} onChange={(e) => toggleRow(row.id, e.target.checked)} />
                </td>
                <td>{row.name}</td>
                <td>{parseTags(row.tags).map((tag) => <span key={tag} className='fx-tpl-tag'>{tag}</span>)}</td>
                <td>{getSeverity(row) || '-'}</td>
                <td>{row.updated_by === 'system' ? <span className='fx-tpl-system-tag'>系统内置</span> : (row.updated_by || '-')}</td>
                <td>
                  <span className='fx-tpl-row-actions'>
                    <button type='button' onClick={() => handleImportSingle(row)}>导入</button>
                    <button type='button' onClick={() => openEdit(row)} disabled={saving}>编辑</button>
                    {row.updated_by !== 'system' && (
                      <button type='button' className='is-danger' onClick={() => handleDelete(row)} disabled={saving}>删除</button>
                    )}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {loading && <div className='fx-tpl-empty'>加载中...</div>}
        {!loading && data.length === 0 && <div className='fx-tpl-empty'>暂无告警规则模板</div>}
      </div>
      <Pagination total={data.length} page={page} pageSize={PAGE_SIZE} onPageChange={setPage} />
      <AlertImportModal rows={importRows} componentIdent={component.ident} onClose={() => setImportRows([])} onDone={() => fetchData()} />
      <PayloadEditorModal state={formState} saving={saving} contentMode='json' onDraft={(draft) => setFormState((s) => s ? { ...s, draft, error: '' } : s)} onSubmit={submitForm} onClose={() => { if (!saving) setFormState(null) }} />
      {confirmModal}
    </div>
  )
}
