/**
 * 模板中心主页面
 * T01: 图标网格 + 搜索 + URL 联动
 * T02: 组件创建/编辑/删除
 * T03: 抽屉详情入口
 */
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { templateApi } from './templateApi.js'
import { normalizeComponents } from './templateModel.js'
import { useConfirm } from '../../shared/ConfirmModal.jsx'
import { ComponentFormModal } from './ComponentFormModal.jsx'
import { ComponentDrawer } from './ComponentDrawer.jsx'
import './templates.css'

const SEARCH_KEY = 'fx-tpl-search-value'

export function TemplatesPage({ query = {}, onNavigate }) {
  const [components, setComponents] = useState([])
  const [loading, setLoading] = useState(false)
  const [searchValue, setSearchValue] = useState(
    () => localStorage.getItem(SEARCH_KEY) || ''
  )
  const [activeComponent, setActiveComponent] = useState(null)
  const [formState, setFormState] = useState(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const { confirm, modal: confirmModal } = useConfirm()

  const fetchComponents = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const data = await templateApi.listComponents()
      setComponents(normalizeComponents(data))
    } catch (err) {
      setError(err.message || '组件列表加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { fetchComponents() }, [fetchComponents])

  // URL 联动：从 query.component 恢复选中状态
  useEffect(() => {
    if (!components.length) return
    const wanted = query.component
    if (wanted) {
      const found = components.find((c) => c.ident === wanted)
      if (found && found.ident !== activeComponent?.ident) setActiveComponent(found)
    } else if (activeComponent) {
      setActiveComponent(null)
    }
  }, [components, query.component])

  const navigate = (patch = {}) => {
    onNavigate?.({ section: 'templates', ...patch })
  }

  const visibleComponents = useMemo(() => {
    const text = searchValue.trim().toLowerCase()
    if (!text) return components
    return components.filter((c) => c.ident.toLowerCase().includes(text))
  }, [components, searchValue])

  const openComponent = (item) => {
    setActiveComponent(item)
    navigate({ component: item.ident })
  }

  const closeDrawer = () => {
    setActiveComponent(null)
    navigate({ component: undefined })
  }

  // T02: 创建
  const openCreate = () => {
    setFormState({ mode: 'create', draft: { ident: '', logo: '', readme: '', disabled: 0 }, error: '' })
  }

  // T02: 编辑
  const openEdit = (item) => {
    setFormState({ mode: 'edit', draft: { ...item }, error: '' })
  }

  // T02: 删除
  const handleDelete = async (item) => {
    const ok = await confirm({
      title: '删除组件',
      message: `确认删除组件「${item.ident}」？此操作不可恢复。`,
      confirmText: '删除',
      danger: true,
    })
    if (!ok) return
    setSaving(true)
    try {
      await templateApi.deleteComponents([item.id])
      await fetchComponents()
      if (activeComponent?.id === item.id) closeDrawer()
    } catch (err) {
      setError(err.message || '删除失败')
    } finally {
      setSaving(false)
    }
  }

  const submitForm = async () => {
    if (!formState) return
    const { mode, draft } = formState
    if (!draft.ident?.trim()) {
      setFormState({ ...formState, error: '组件标识必填' })
      return
    }
    setSaving(true)
    try {
      if (mode === 'create') {
        await templateApi.createComponents([draft])
      } else {
        await templateApi.updateComponent(draft)
      }
      setFormState(null)
      await fetchComponents()
    } catch (err) {
      setFormState({ ...formState, error: err.message || '保存失败' })
    } finally {
      setSaving(false)
    }
  }

  // 更新 readme（抽屉内保存）
  const updateReadme = async (component, readme) => {
    await templateApi.updateComponent({ ...component, readme })
    await fetchComponents()
  }

  return (
    <main className='fx-tpl-page'>
      <div className='fx-tpl-header'>
        <input
          type='text'
          value={searchValue}
          onChange={(e) => {
            setSearchValue(e.target.value)
            localStorage.setItem(SEARCH_KEY, e.target.value)
          }}
          placeholder='搜索组件标识...'
        />
        <div className='fx-tpl-actions'>
          <button type='button' onClick={fetchComponents} disabled={loading}>
            {loading ? '刷新中...' : '刷新'}
          </button>
          <button type='button' className='is-primary' onClick={openCreate} disabled={saving}>
            新增组件
          </button>
        </div>
      </div>
      {error && <div className='fx-tpl-alert is-error'>{error}</div>}
      <section className='fx-tpl-grid'>
        {visibleComponents.map((item) => (
          <GridItem
            key={item.id}
            item={item}
            selected={item.ident === activeComponent?.ident}
            onOpen={openComponent}
            onEdit={openEdit}
            onDelete={handleDelete}
            saving={saving}
          />
        ))}
        {!loading && visibleComponents.length === 0 && (
          <div className='fx-tpl-empty'>暂无组件</div>
        )}
      </section>
      {activeComponent && (
        <ComponentDrawer
          component={activeComponent}
          query={query}
          onClose={closeDrawer}
          onNavigate={navigate}
          onUpdateReadme={updateReadme}
        />
      )}
      <ComponentFormModal
        state={formState}
        saving={saving}
        onDraft={(draft) => setFormState((s) => s ? { ...s, draft, error: '' } : s)}
        onSubmit={submitForm}
        onClose={() => { if (!saving) setFormState(null) }}
      />
      {confirmModal}
    </main>
  )
}

function GridItem({ item, selected, onOpen, onEdit, onDelete, saving }) {
  return (
    <div
      className={`fx-tpl-grid-item${selected ? ' is-selected' : ''}`}
      onClick={() => onOpen(item)}
    >
      {item.logo
        ? <img src={item.logo} alt={item.ident} />
        : <span style={{ fontSize: 24, fontWeight: 700 }}>{(item.ident || 'FX').slice(0, 2).toUpperCase()}</span>
      }
      <span className='fx-tpl-ident'>{item.ident}</span>
      <span className='fx-tpl-ops'>
        <button type='button' onClick={(e) => { e.stopPropagation(); onEdit(item) }} disabled={saving} title='编辑'>E</button>
        <button type='button' onClick={(e) => { e.stopPropagation(); onDelete(item) }} disabled={saving} title='删除'>D</button>
      </span>
      {item.disabled === 1 && (
        <span className='fx-tpl-disabled-badge' title='已禁用'>⊘</span>
      )}
    </div>
  )
}
