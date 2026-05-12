import React, { useEffect, useMemo, useState } from 'react'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/styles.css'
import { dashboardsApi } from '../../api/dashboards.js'
import {
  dashboardExportPayload,
  dashboardPayload,
  displayJson,
  displayText,
  downloadJson,
  normalizeDashboard,
  normalizePanels,
  normalizeTemplate,
  normalizeVariables,
  toTags,
} from './dashboardModel.js'
import ConfirmModal from './ConfirmModal.jsx'
import DashboardList from './DashboardList.jsx'
import { ImportJsonModal, ShareConfirmModal } from './DashboardModals.jsx'
import DetailView from './DetailView.jsx'
import TemplatesView from './TemplatesView.jsx'
import { PermissionProvider, usePermission } from './usePermission.jsx'
import './dashboards.css'

/** debounce hook */
function useDebounce(value, delay = 300) {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])
  return debounced
}

const ALL_COLUMNS = [
  { key: 'title', label: '仪表盘名称' },
  { key: 'tags', label: '分类标签' },
  { key: 'updatedAt', label: '更新时间' },
  { key: 'shared', label: '公开' },
]
const DEFAULT_COLUMNS = ['title', 'tags', 'updatedAt', 'shared']

function loadColumnSettings() {
  try { const s = localStorage.getItem('fx-dash-columns'); if (s) return JSON.parse(s) } catch { /* ignore */ }
  return DEFAULT_COLUMNS
}
function saveColumnSettings(cols) { localStorage.setItem('fx-dash-columns', JSON.stringify(cols)) }

const defaultDraft = { title: '', ident: '', tags: '', description: '' }
const makeError = (error, fallback = '请求失败') => {
  if (error?.status === 401) return '登录已过期，请重新登录。'
  if (error?.status === 403) return '没有仪表盘权限。'
  if (error?.status >= 500) return `HTTP ${error.status}: 仪表盘服务异常。`
  return displayText(error?.message || fallback)
}

function BusinessGroupSidebar({ groups, scope, setScope, collapsed, setCollapsed }) {
  const [search, setSearch] = useState('')
  const filtered = groups.filter((g) => !search || g.label.toLowerCase().includes(search.toLowerCase()))
  if (collapsed) {
    return <aside className="fx-dash-sidebar" style={{ width: 40, padding: '8px 4px', cursor: 'pointer' }} onClick={() => setCollapsed(false)}><span style={{ writingMode: 'vertical-rl', fontSize: 12, color: '#65758d' }}>&gt;</span></aside>
  }
  return (
    <aside className="fx-dash-sidebar" style={{ width: 280, minWidth: 280 }}>
      <strong>预置筛选</strong>
      <button type="button" className={scope === 'public' ? 'is-active' : ''} onClick={() => setScope('public')}>公开仪表盘</button>
      <button type="button" className={scope === 'all' ? 'is-active' : ''} onClick={() => setScope('all')}>所属业务组仪表盘<span title="展示当前用户所属业务组的仪表盘" style={{ cursor: 'help', marginLeft: 4, color: '#8a96aa' }}>ⓘ</span></button>
      <strong>业务组</strong>
      <input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="🔍 请输入搜索关键字" style={{ fontSize: 12 }} />
      <div style={{ maxHeight: 400, overflowY: 'auto' }}>
        {filtered.map((g) => (
          <button type="button" key={g.key} className={scope === g.key ? 'is-active' : ''} onClick={() => setScope(g.key)}><span>{g.label}</span><small>{g.count}</small></button>
        ))}
      </div>
      <button type="button" style={{ marginTop: 'auto', textAlign: 'center', fontSize: 12, color: '#8a96aa' }} onClick={() => setCollapsed(true)}>&lt;</button>
    </aside>
  )
}

function Header({ keyword, setKeyword, onRefresh, onCreate, onImportJson, selectedCount, onBatch, loading, visibleCols, setVisibleCols }) {
  const [showColMenu, setShowColMenu] = useState(false)
  const [showMore, setShowMore] = useState(false)
  const toggleCol = (key) => { const next = visibleCols.includes(key) ? visibleCols.filter((c) => c !== key) : [...visibleCols, key]; setVisibleCols(next); saveColumnSettings(next) }
  return (
    <div className="fx-dash-toolbar" style={{ marginBottom: 12 }}>
      <div className="fx-dash-actions">
        <button type="button" className="fx-dash-icon-btn" disabled={loading} onClick={onRefresh} title="刷新"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M23 4v6h-6M1 20v-6h6" /><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" /></svg></button>
        <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder="🔍 仪表盘名称、分类标签" style={{ minWidth: 220 }} />
      </div>
      <div className="fx-dash-actions">
        <button type="button" className="is-primary" onClick={onCreate}>新建</button>
        <button type="button" onClick={onImportJson}>导入</button>
        <div style={{ position: 'relative', display: 'inline-block' }}>
          <button type="button" onClick={() => setShowMore(!showMore)}>更多操作 ∨</button>
          {showMore && <div className="fx-panel-menu__dropdown" style={{ top: '100%', left: 0, minWidth: 120 }}><button onClick={() => { onBatch('clone'); setShowMore(false) }}>批量克隆{selectedCount > 0 && ` (${selectedCount})`}</button><button className="fx-panel-menu__danger" onClick={() => { onBatch('delete'); setShowMore(false) }}>批量删除{selectedCount > 0 && ` (${selectedCount})`}</button></div>}
        </div>
        <div style={{ position: 'relative', display: 'inline-block' }}>
          <button type="button" className="fx-dash-icon-btn" onClick={() => setShowColMenu(!showColMenu)} title="列设置">⚙</button>
          {showColMenu && <div className="fx-dash-col-menu">{ALL_COLUMNS.map((col) => <label key={col.key}><input type="checkbox" checked={visibleCols.includes(col.key)} onChange={() => toggleCol(col.key)} />{col.label}</label>)}</div>}
        </div>
      </div>
    </div>
  )
}

function FormModal({ draft, setDraft, onSubmit, onClose, saving, error }) {
  if (!draft) return null
  return (
    <div className="fx-dash-modal"><div className="fx-dash-modal__body" style={{ width: 'min(520px, 100%)' }}>
      <header><h2>{draft.id ? '编辑仪表盘' : '新建仪表盘'}</h2><button type="button" onClick={onClose}>×</button></header>
      <div className="fx-dash-form">
        <label><span><span className="fx-field-required">*</span>仪表盘名称</span><input value={draft.title} onChange={(e) => setDraft({ ...draft, title: e.target.value })} placeholder="请输入仪表盘名称" /></label>
        <label><span>英文标识</span><input value={draft.ident} onChange={(e) => setDraft({ ...draft, ident: e.target.value.replace(/[^a-zA-Z0-9-]/g, '') })} placeholder="仅允许英文、数字、横线" /></label>
        <label><span>分类标签</span><input value={draft.tags} onChange={(e) => setDraft({ ...draft, tags: e.target.value })} placeholder="多个标签用空格分隔" /></label>
        <label><span>备注</span><textarea value={draft.description || ''} onChange={(e) => setDraft({ ...draft, description: e.target.value })} rows={2} /></label>
        {error && <div className="fx-dash-alert is-error">{error}</div>}
        <div className="fx-dash-actions" style={{ justifyContent: 'flex-end' }}><button type="button" onClick={onClose}>取消</button><button type="button" className="is-primary" disabled={saving || !draft.title.trim()} onClick={onSubmit}>{saving ? '保存中...' : '确定'}</button></div>
      </div>
    </div>
    </div>
  )
}

// ─── 主页面 ───
export function DashboardsPage({ query = {}, onNavigate }) {
  const [dashboards, setDashboards] = useState([])
  const [templates, setTemplates] = useState([])
  const [keyword, setKeyword] = useState(localStorage.getItem('fx-dash-search') || '')
  const debouncedKeyword = useDebounce(keyword, 300)
  const [scope, setScope] = useState('all')
  const [selected, setSelected] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [detailError, setDetailError] = useState('')
  const [draft, setDraft] = useState(null)
  const [saving, setSaving] = useState(false)
  const [modal, setModal] = useState(null)
  const [showImportJson, setShowImportJson] = useState(false)
  const [shareTarget, setShareTarget] = useState(null)
  const [visibleCols, setVisibleCols] = useState(loadColumnSettings)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const [batchDeleteConfirm, setBatchDeleteConfirm] = useState(false)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const section = query.section === 'detail' || query.section === 'templates' ? query.section : 'list'
  const active = useMemo(() => dashboards.find((r) => r.id === String(query.id)) || dashboards[0], [dashboards, query.id])
  const groups = useMemo(() => Object.values(dashboards.reduce((acc, r) => {
    const label = r.resourceGroupId || '未分组'
    const key = `group:${label}`
    acc[key] = { key, label, count: (acc[key]?.count || 0) + 1 }
    return acc
  }, {})), [dashboards])
  const filtered = useMemo(() => dashboards.filter((row) => {
    const text = [row.title, row.description, row.tags.join(' '), row.updatedBy].join(' ').toLowerCase()
    const byKeyword = !debouncedKeyword || text.includes(debouncedKeyword.toLowerCase())
    const byScope = scope === 'all' || (scope === 'public' ? row.shared : `group:${row.resourceGroupId || '未分组'}` === scope)
    return byKeyword && byScope
  }), [dashboards, debouncedKeyword, scope])
  const variables = useMemo(() => normalizeVariables(active?.variables), [active])
  const panels = useMemo(() => normalizePanels(active?.panels), [active])

  const loadDashboards = async () => {
    setLoading(true); setError('')
    try { setDashboards((await dashboardsApi.list()).map(normalizeDashboard).filter((r) => r.id)) } catch (err) { setError(makeError(err, '仪表盘列表加载失败')) } finally { setLoading(false) }
  }
  const loadTemplates = async () => {
    try { setTemplates((await dashboardsApi.listTemplates()).map(normalizeTemplate).filter((r) => r.id)) } catch (err) { setError(makeError(err, '模板加载失败')) }
  }
  const openDetail = async (id) => { onNavigate?.({ section: 'detail', id }); setDetailError(''); try { const d = normalizeDashboard(await dashboardsApi.detail(id)); setDashboards((rows) => [d, ...rows.filter((r) => r.id !== d.id)]) } catch (err) { setDetailError(makeError(err, '详情加载失败')) } }
  const openForm = (row) => setDraft(row ? { id: row.id, title: row.title, ident: row.raw?.ident || '', tags: row.tags.join(' '), description: row.description } : { ...defaultDraft })
  const saveDraft = async () => {
    setSaving(true); setError('')
    try {
      const body = dashboardPayload({ ...draft, tags: draft.tags, resourceGroupId: '', workspaceId: '', variables: {}, panels: [], status: 'active' })
      const saved = normalizeDashboard(draft.id ? await dashboardsApi.update(draft.id, body) : await dashboardsApi.create(body))
      setDashboards((rows) => [saved, ...rows.filter((r) => r.id !== saved.id)]); setDraft(null)
    } catch (err) { setError(makeError(err, '保存失败')) } finally { setSaving(false) }
  }

  const rowAction = async (action, row) => {
    try {
      if (action === 'edit') return openForm(row)
      if (action === 'clone') { const c = normalizeDashboard(await dashboardsApi.clone(row.id)); setDashboards((rows) => [c, ...rows]) }
      if (action === 'share') { setShareTarget(row); return }
      if (action === 'delete') { setDeleteConfirm(row); return }
      if (action === 'export') downloadJson(`${row.title || row.id}.json`, dashboardExportPayload(row))
    } catch (err) { setModal({ title: '操作失败', body: makeError(err) }) }
  }
  const confirmDeleteRow = async () => { if (!deleteConfirm) return; try { await dashboardsApi.remove(deleteConfirm.id); setDeleteConfirm(null); await loadDashboards() } catch (err) { setModal({ title: '删除失败', body: makeError(err) }); setDeleteConfirm(null) } }
  const confirmShare = async () => { if (!shareTarget) return; try { await dashboardsApi.share(shareTarget.id); setShareTarget(null); await loadDashboards() } catch (err) { setModal({ title: '公开失败', body: makeError(err) }); setShareTarget(null) } }
  const batchAction = async (action) => {
    if (!selected.length) { setModal({ title: '提示', body: '请先选择仪表盘' }); return }
    if (action === 'delete') { setBatchDeleteConfirm(true); return }
    for (const id of selected) { const row = dashboards.find((r) => r.id === id); if (row) { try { if (action === 'clone') await dashboardsApi.clone(id); else if (action === 'export') downloadJson(`${row.title || row.id}.json`, dashboardExportPayload(row)) } catch { /* continue */ } } }
    setSelected([]); await loadDashboards()
  }
  const confirmBatchDelete = async () => { for (const id of selected) { try { await dashboardsApi.remove(id) } catch { /* continue */ } }; setBatchDeleteConfirm(false); setSelected([]); await loadDashboards() }
  const importJsonData = async (data) => { try { const body = { title: data.title || data.name || '导入仪表盘', description: data.description || '', tags: toTags(data.tags || []), variables: data.variables || {}, panels: Array.isArray(data.panels) ? data.panels : [], status: 'active' }; const saved = normalizeDashboard(await dashboardsApi.create(body)); setDashboards((rows) => [saved, ...rows]); setShowImportJson(false) } catch (err) { setModal({ title: '导入失败', body: makeError(err) }) } }
  const previewTemplate = (tpl) => setModal({ title: `预览：${tpl.title}`, body: displayJson({ variables: tpl.variables, panels: tpl.panels }) })
  const importTemplate = async (tpl) => { try { const saved = normalizeDashboard(await dashboardsApi.importTemplate(tpl.id, { title: tpl.title, variables: tpl.variables, tags: tpl.tags })); setDashboards((rows) => [saved, ...rows]); onNavigate?.({ section: 'detail', id: saved.id }) } catch (err) { setModal({ title: '导入失败', body: makeError(err) }) } }
  const blocked = (title, body) => setModal({ title, body })
  const onSearchAppendTag = (tag) => { const parts = keyword ? keyword.split(' ') : []; if (parts.includes(tag)) return; const next = keyword ? `${keyword} ${tag}` : tag; setKeyword(next); localStorage.setItem('fx-dash-search', next) }

  useEffect(() => { loadDashboards(); loadTemplates() }, [])

  if (section === 'templates') {
    return (<>
      <TemplatesView templates={templates} loading={loading} error={error} onBack={() => onNavigate?.({ section: 'list' })} onPreview={previewTemplate} onImport={importTemplate} />
      <FormModal draft={draft} setDraft={setDraft} saving={saving} error={error} onSubmit={saveDraft} onClose={() => setDraft(null)} />
      {modal && <ModalOverlay title={modal.title} body={modal.body} onClose={() => setModal(null)} />}
    </>)
  }

  return (
    <main className="fx-dash-page">
      {section === 'detail' && active ? (
        <DetailView dashboard={active} variables={variables} panels={panels} detailError={detailError} loading={loading && !active} onBack={() => onNavigate?.({ section: 'list' })} onRefresh={() => openDetail(active.id)} onExport={() => rowAction('export', active)} onShare={() => rowAction('share', active)} onFullscreen={() => document.documentElement.requestFullscreen?.()} onBlocked={blocked} onUpdateDashboard={async (updates) => { try { const body = dashboardPayload({ ...active, ...updates }); const saved = normalizeDashboard(await dashboardsApi.update(active.id, body)); setDashboards((rows) => [saved, ...rows.filter((r) => r.id !== saved.id)]) } catch (err) { setDetailError(makeError(err, '保存失败')) } }} />
      ) : (
        <div className="fx-dash-layout" style={{ gridTemplateColumns: sidebarCollapsed ? '40px minmax(0,1fr)' : '280px minmax(0,1fr)' }}>
          <BusinessGroupSidebar groups={groups} scope={scope} setScope={setScope} collapsed={sidebarCollapsed} setCollapsed={setSidebarCollapsed} />
          <section className="fx-dash-main">
            <Header keyword={keyword} setKeyword={(v) => { setKeyword(v); localStorage.setItem('fx-dash-search', v) }} loading={loading} selectedCount={selected.length} visibleCols={visibleCols} setVisibleCols={setVisibleCols} onRefresh={loadDashboards} onCreate={() => openForm(null)} onImportJson={() => setShowImportJson(true)} onBatch={batchAction} />
            {error && <div className="fx-dash-alert is-error">{error}</div>}
            <DashboardList rows={filtered} selected={selected} setSelected={setSelected} onOpen={openDetail} onRowAction={rowAction} searchVal={keyword} onSearchAppendTag={onSearchAppendTag} visibleCols={visibleCols} page={page} pageSize={pageSize} onPageChange={setPage} onPageSizeChange={(s) => { setPageSize(s); setPage(1) }} />
          </section>
        </div>
      )}
      <FormModal draft={draft} setDraft={setDraft} saving={saving} error={error} onSubmit={saveDraft} onClose={() => setDraft(null)} />
      {showImportJson && <ImportJsonModal onClose={() => setShowImportJson(false)} onImport={importJsonData} />}
      {shareTarget && <ShareConfirmModal row={shareTarget} onClose={() => setShareTarget(null)} onConfirm={confirmShare} />}
      {deleteConfirm && <ConfirmModal title="删除仪表盘" message={`确定删除仪表盘「${deleteConfirm.title}」？此操作不可撤销。`} confirmText="删除" danger onConfirm={confirmDeleteRow} onCancel={() => setDeleteConfirm(null)} />}
      {batchDeleteConfirm && <ConfirmModal title="批量删除" message={`确定删除选中的 ${selected.length} 个仪表盘？此操作不可撤销。`} confirmText="删除" danger onConfirm={confirmBatchDelete} onCancel={() => setBatchDeleteConfirm(false)} />}
      {modal && <ModalOverlay title={modal.title} body={modal.body} onClose={() => setModal(null)} />}
      {loading && dashboards.length === 0 && <div className="fx-dash-global-loading"><div className="fx-dash-global-loading__spinner" /><span>加载中...</span></div>}
    </main>
  )
}

function ModalOverlay({ title, body, onClose }) {
  return <div className="fx-dash-modal"><div className="fx-dash-modal__body"><header><h2>{title}</h2><button type="button" onClick={onClose}>×</button></header><pre>{displayText(body)}</pre></div></div>
}

export function DashboardsPageWithPermission(props) {
  return <PermissionProvider><DashboardsPage {...props} /></PermissionProvider>
}
