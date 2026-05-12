/**
 * 采集模板 Tab (T04)
 * YAML/TOML 模板列表 + Monaco 编辑器
 */
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { templateApi } from './templateApi.js'
import { normalizePayloads, TypeEnum } from './templateModel.js'
import { useConfirm, Pagination } from '../../shared/ConfirmModal.jsx'
import { PayloadEditorModal } from './PayloadEditorModal.jsx'

const PAGE_SIZE = 20

export function CollectTplsTab({ component }) {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [formState, setFormState] = useState(null)
  const [saving, setSaving] = useState(false)
  const { confirm, modal: confirmModal } = useConfirm()
  const debounceRef = useRef(null)

  const fetchData = useCallback(async (q = query) => {
    setLoading(true)
    try {
      const result = await templateApi.listPayloads({
        component_id: component.id,
        type: TypeEnum.collect,
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

  const openCreate = () => {
    setFormState({
      mode: 'create',
      draft: { name: '', cate: '', type: TypeEnum.collect, component_id: component.id, content: '' },
      error: '',
    })
  }

  const openEdit = (row) => {
    setFormState({ mode: 'edit', draft: { ...row }, error: '' })
  }

  const handleDelete = async (row) => {
    const ok = await confirm({ title: '删除采集模板', message: `确认删除「${row.name}」？`, confirmText: '删除', danger: true })
    if (!ok) return
    setSaving(true)
    try {
      await templateApi.deletePayloads([row.id])
      await fetchData()
    } finally {
      setSaving(false)
    }
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
    } catch (err) {
      setFormState({ ...formState, error: err.message || '保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const paged = data.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  return (
    <div>
      <div className='fx-tpl-toolbar'>
        <div className='fx-tpl-toolbar-left'>
          <input type='text' value={query} onChange={(e) => handleSearch(e.target.value)} placeholder='搜索模板名称...' />
        </div>
        <div className='fx-tpl-toolbar-right'>
          <button type='button' className='is-primary' onClick={openCreate} disabled={saving}>新增</button>
        </div>
      </div>
      <div className='fx-tpl-table-wrap'>
        <table className='fx-tpl-table'>
          <thead>
            <tr>
              <th>分类</th>
              <th>名称</th>
              <th>更新人</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {paged.map((row) => (
              <tr key={row.id}>
                <td>{row.cate || '-'}</td>
                <td>{row.name}</td>
                <td>{row.updated_by === 'system' ? <span className='fx-tpl-system-tag'>系统内置</span> : (row.updated_by || '-')}</td>
                <td>
                  <span className='fx-tpl-row-actions'>
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
        {!loading && data.length === 0 && <div className='fx-tpl-empty'>暂无采集模板</div>}
      </div>
      <Pagination total={data.length} page={page} pageSize={PAGE_SIZE} onPageChange={setPage} />
      <PayloadEditorModal
        state={formState}
        saving={saving}
        contentMode='yaml'
        onDraft={(draft) => setFormState((s) => s ? { ...s, draft, error: '' } : s)}
        onSubmit={submitForm}
        onClose={() => { if (!saving) setFormState(null) }}
      />
      {confirmModal}
    </div>
  )
}
