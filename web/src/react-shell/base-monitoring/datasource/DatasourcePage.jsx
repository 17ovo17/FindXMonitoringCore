import React, { useEffect, useMemo, useState } from 'react'
import { datasourceApi } from '../../api/datasources.js'
import { isPermissionError, redactText } from '../../api/http.js'
import {
  datasourceId,
  datasourceTypes,
  displayAddress,
  displayCluster,
  displayName,
  displayType,
  rowKey,
  safeDisplayText,
  statusText,
} from './datasourceModel.js'
import { SourceTypeModal } from './SourceTypeModal.jsx'
import { DatasourceForm } from './DatasourceForm.jsx'
import { DatasourceDrawer } from './DatasourceDrawer.jsx'
import './datasource.css'

function formatError(error) {
  if (isPermissionError(error)) {
    return error.status === 401 ? '登录已过期，请重新登录后访问数据源能力。' : '无权限访问该数据源能力。'
  }
  return redactText(error?.message || '请求失败')
}

function useDebounce(value, delay) {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])
  return debounced
}

function StatusDot({ row }) {
  const text = statusText(row)
  const color = text === '启用' ? '#22c55e' : text === '禁用' ? '#9ca3af' : '#eab308'
  return (
    <span className='fx-ds-status'>
      <span style={{ width: 8, height: 8, borderRadius: '50%', background: color, marginRight: 6, display: 'inline-block' }} />
      {text}
    </span>
  )
}

function TypeIcon({ row }) {
  const type = String(displayType(row) || '').toLowerCase()
  const matched = datasourceTypes.find((t) => t.type === type)
  if (matched?.logo) {
    return <img src={matched.logo} alt={matched.name} width='20' height='20' style={{ verticalAlign: 'middle', marginRight: 6 }} />
  }
  return null
}
export function DatasourcePage() {
  const [rows, setRows] = useState([])
  const [apiSource, setApiSource] = useState('')
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [selected, setSelected] = useState(null)
  const [typeModalVisible, setTypeModalVisible] = useState(false)
  const [formState, setFormState] = useState(null)
  const [sortAsc, setSortAsc] = useState(true)

  const debouncedKeyword = useDebounce(keyword, 300)

  const load = async () => {
    setLoading(true)
    setError('')
    setNotice('')
    try {
      const out = await datasourceApi.list()
      setRows(out.rows)
      setApiSource(out.source)
    } catch (err) {
      setRows([])
      setError(formatError(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const typeOptions = useMemo(() => [...new Set(rows.map(displayType).filter(Boolean))], [rows])

  const filteredRows = useMemo(() => {
    const text = debouncedKeyword.trim().toLowerCase()
    const filtered = rows.filter((row) => {
      const haystack = [displayType(row), displayName(row), displayCluster(row), displayAddress(row)]
        .map((v) => safeDisplayText(v).toLowerCase()).join(' ')
      return (!text || haystack.includes(text))
        && (!typeFilter || displayType(row) === typeFilter)
    })
    return filtered.sort((a, b) => {
      const na = (displayName(a) || '').toLowerCase()
      const nb = (displayName(b) || '').toLowerCase()
      return sortAsc ? na.localeCompare(nb) : nb.localeCompare(na)
    })
  }, [rows, debouncedKeyword, typeFilter, sortAsc])

  const chooseType = (item) => {
    setTypeModalVisible(false)
    setFormState({ dsType: item, editRow: null })
  }

  const handleEdit = (row) => {
    const type = String(displayType(row) || 'prometheus').toLowerCase()
    const matched = datasourceTypes.find((t) => t.type === type) || datasourceTypes[0]
    setFormState({ dsType: matched, editRow: row })
    setSelected(null)
  }

  const handleDelete = async (row) => {
    const id = datasourceId(row)
    if (!id) { setError('该记录缺少 ID，无法删除'); return }
    if (!window.confirm(`确认删除数据源「${displayName(row)}」？此操作不可恢复。`)) return
    try {
      await datasourceApi.remove(id)
      setNotice('删除成功')
      load()
    } catch (err) {
      setError(`删除失败：${redactText(err?.message || '未知错误')}`)
    }
  }

  const handleToggle = async (row) => {
    const id = datasourceId(row)
    if (!id) { setError('该记录缺少 ID'); return }
    const current = statusText(row)
    const newStatus = current === '启用' ? 'disabled' : 'enabled'
    try {
      await datasourceApi.save({ ...row, id, status: newStatus })
      setNotice(`已${newStatus === 'enabled' ? '启用' : '禁用'}`)
      load()
    } catch (err) {
      setError(`操作失败：${redactText(err?.message || '未知错误')}`)
    }
  }

  if (formState) {
    return (
      <main className='fx-ds-page'>
        <DatasourceForm
          dsType={formState.dsType}
          editRow={formState.editRow}
          onSaved={() => { setFormState(null); load() }}
          onCancel={() => setFormState(null)}
        />
      </main>
    )
  }

  return (
    <main className='fx-ds-page'>
      <header className='fx-ds-header'>
        <div>
          <p>集成中心</p>
          <h1>数据源</h1>
          <span>管理和配置监控数据源连接。</span>
        </div>
        <div className='fx-ds-actions'>
          <button type='button' onClick={load} disabled={loading}>{loading ? '刷新中' : '刷新'}</button>
          <button type='button' className='is-primary' onClick={() => setTypeModalVisible(true)}>新增数据源</button>
        </div>
      </header>

      <section className='fx-ds-panel'>
        <div className='fx-ds-toolbar'>
          <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder='搜索数据源名称、类型或集群...' />
          <select value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)}>
            <option value=''>全部类型</option>
            {typeOptions.map((t) => <option key={t} value={t}>{safeDisplayText(t)}</option>)}
          </select>
          <button type='button' onClick={() => setSortAsc(!sortAsc)}>名称 {sortAsc ? '↑' : '↓'}</button>
        </div>

        {apiSource && <div className='fx-ds-source'>数据来源：{apiSource}</div>}
        {error && <div className='fx-ds-alert is-error'>{error}</div>}
        {notice && <div className='fx-ds-alert is-warning'>{notice}</div>}

        <div className='fx-ds-table-wrap'>
          <table className='fx-ds-table'>
            <thead>
              <tr>
                <th>类型</th>
                <th>名称</th>
                <th>集群</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {filteredRows.map((row, index) => {
                const key = rowKey(row) || `row-${index}`
                return (
                  <tr key={key}>
                    <td><TypeIcon row={row} />{safeDisplayText(displayType(row)) || '-'}</td>
                    <td><button type='button' className='fx-ds-link' onClick={() => setSelected(row)}>{safeDisplayText(displayName(row)) || '-'}</button></td>
                    <td>{safeDisplayText(displayCluster(row)) || 'default'}</td>
                    <td><StatusDot row={row} /></td>
                    <td className='fx-ds-row-actions'>
                      <button type='button' onClick={() => handleEdit(row)}>编辑</button>
                      <button type='button' onClick={() => handleToggle(row)}>{statusText(row) === '启用' ? '禁用' : '启用'}</button>
                      <button type='button' className='is-danger' onClick={() => handleDelete(row)}>删除</button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
          {!loading && filteredRows.length === 0 && <div className='fx-ds-empty'>暂无数据源</div>}
        </div>
      </section>

      <DatasourceDrawer row={selected} onClose={() => setSelected(null)} onEdit={handleEdit} />
      <SourceTypeModal visible={typeModalVisible} onClose={() => setTypeModalVisible(false)} onChoose={chooseType} />
    </main>
  )
}
