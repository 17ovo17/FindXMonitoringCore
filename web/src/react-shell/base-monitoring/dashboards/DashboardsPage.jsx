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
import { ImportJsonModal, ShareConfirmModal } from './DashboardModals.jsx'
import DetailView from './DetailView.jsx'
import TemplatesView from './TemplatesView.jsx'
import './dashboards.css'

/** D12: debounce hook */
function useDebounce(value, delay = 300) {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])
  return debounced
}

const ALL_COLUMNS = [
  { key: 'title', label: '名称' },
  { key: 'tags', label: '标签' },
  { key: 'description', label: '说明' },
  { key: 'updatedAt', label: '更新时间' },
  { key: 'updatedBy', label: '更新人' },
  { key: 'shared', label: '共享' },
]
const DEFAULT_COLUMNS = ['title', 'tags', 'updatedAt', 'updatedBy', 'shared']

function loadColumnSettings() {
  try {
    const saved = localStorage.getItem('fx-dash-columns')
    if (saved) return JSON.parse(saved)
  } catch { /* ignore */ }
  return DEFAULT_COLUMNS
}
function saveColumnSettings(cols) {
  localStorage.setItem('fx-dash-columns', JSON.stringify(cols))
}

const defaultDraft = { title: '', description: '', workspaceId: '', resourceGroupId: '', tags: '', variables: {}, panels: [], status: 'active' }

const makeError = (error, fallback = '请求失败') => {
  if (error?.status === 401) return '登录已过期，请重新登录。'
  if (error?.status === 403) return '没有仪表盘权限。'
  if (error?.status >= 500) return `HTTP ${error.status}: 仪表盘服务异常。`
  return displayText(error?.message || fallback)
}

function Toolbar({ keyword, setKeyword, onRefresh, onCreate, onTemplates, onImportJson, selectedCount, onBatch, loading, visibleCols, setVisibleCols }) {
  const [showColMenu, setShowColMenu] = useState(false)
  const toggleCol = (key) => {
    const next = visibleCols.includes(key) ? visibleCols.filter((c) => c !== key) : [...visibleCols, key]
    setVisibleCols(next)
    saveColumnSettings(next)
  }
  return (
    <header className='fx-dash-toolbar'>
      <div>
        <p>仪表盘</p>
        <h1>仪表盘列表</h1>
      </div>
      <div className='fx-dash-actions'>
        <input value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder='搜索名称、标签、说明' />
        <button type='button' disabled={loading} onClick={onRefresh}>{loading ? '刷新中...' : '刷新'}</button>
        <button type='button' className='is-primary' onClick={onCreate}>新建</button>
        <button type='button' onClick={onTemplates}>模板导入</button>
        <button type='button' onClick={onImportJson}>导入 JSON</button>
        <div style={{ position: 'relative', display: 'inline-block' }}>
          <button type='button' onClick={() => setShowColMenu(!showColMenu)}>列设置</button>
          {showColMenu && (
            <div className='fx-dash-col-menu'>
              {ALL_COLUMNS.map((col) => (
                <label key={col.key}><input type='checkbox' checked={visibleCols.includes(col.key)} onChange={() => toggleCol(col.key)} />{col.label}</label>
              ))}
            </div>
          )}
        </div>
        <select disabled={selectedCount === 0} onChange={(event) => { if (event.target.value) onBatch(event.target.value); event.target.value = '' }}>
          <option value=''>批量操作</option>
          <option value='clone'>克隆</option>
          <option value='share'>公开配置</option>
          <option value='export'>导出</option>
          <option value='delete'>删除</option>
        </select>
      </div>
    </header>
  )
}

function Sidebar({ groups, scope, setScope }) {
  return (
    <aside className='fx-dash-sidebar'>
      <button type='button' className={scope === 'all' ? 'is-active' : ''} onClick={() => setScope('all')}>全部仪表盘</button>
      <button type='button' className={scope === 'public' ? 'is-active' : ''} onClick={() => setScope('public')}>公开仪表盘</button>
      <strong>业务组</strong>
      {groups.map((group) => (
        <button type='button' key={group.key} className={scope === group.key ? 'is-active' : ''} onClick={() => setScope(group.key)}>
          <span>{group.label}</span><small>{group.count}</small>
        </button>
      ))}
    </aside>
  )
}

function DashboardList({ rows, selected, setSelected, onOpen, onRowAction, visibleCols }) {
  const show = (key) => visibleCols.includes(key)
  return (
    <div className='fx-dash-table'>
      <table>
        <thead>
          <tr>
            <th><input type='checkbox' checked={rows.length > 0 && selected.length === rows.length} onChange={(event) => setSelected(event.target.checked ? rows.map((row) => row.id) : [])} /></th>
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
          {rows.map((row) => (
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
    </div>
  )
}

function Modal({ title, children, onClose }) {
  return <div className='fx-dash-modal'><div className='fx-dash-modal__body'><header><h2>{title}</h2><button type='button' onClick={onClose}>×</button></header>{children}</div></div>
}

function DashboardForm({ draft, setDraft, onSubmit, onClose, saving, error }) {
  return (
    <Modal title={draft.id ? '编辑仪表盘' : '新建仪表盘'} onClose={onClose}>
      <div className='fx-dash-form'>
        <label><span>名称</span><input value={draft.title} onChange={(event) => setDraft({ ...draft, title: event.target.value })} /></label>
        <label><span>说明</span><textarea value={draft.description} onChange={(event) => setDraft({ ...draft, description: event.target.value })} /></label>
        <label><span>业务组</span><input value={draft.resourceGroupId} onChange={(event) => setDraft({ ...draft, resourceGroupId: event.target.value })} /></label>
        <label><span>标签</span><input value={draft.tags} onChange={(event) => setDraft({ ...draft, tags: event.target.value })} /></label>
        {error && <div className='fx-dash-alert is-error'>{error}</div>}
        <div className='fx-dash-actions'><button type='button' onClick={onClose}>取消</button><button type='button' className='is-primary' disabled={saving} onClick={onSubmit}>{saving ? '保存中...' : '保存'}</button></div>
      </div>
    </Modal>
  )
}

function DashboardOverlays({ draft, setDraft, saving, error, onSubmit, onCloseDraft, modal, onCloseModal }) {
  return (
    <>
      {draft && <DashboardForm draft={draft} setDraft={setDraft} saving={saving} error={error} onSubmit={onSubmit} onClose={onCloseDraft} />}
      {modal && <Modal title={modal.title} onClose={onCloseModal}><pre>{displayText(modal.body)}</pre></Modal>}
    </>
  )
}

export function DashboardsPage({ query = {}, onNavigate }) {
  const [dashboards, setDashboards] = useState([])
  const [templates, setTemplates] = useState([])
  const [keyword, setKeyword] = useState('')
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
  const section = query.section === 'detail' || query.section === 'templates' ? query.section : 'list'
  const active = useMemo(() => dashboards.find((row) => row.id === String(query.id)) || dashboards[0], [dashboards, query.id])
  const groups = useMemo(() => Object.values(dashboards.reduce((acc, row) => {
    const label = row.resourceGroupId || '未分组'
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
    try { setDashboards((await dashboardsApi.list()).map(normalizeDashboard).filter((row) => row.id)) } catch (err) { setError(makeError(err, '仪表盘列表加载失败')) } finally { setLoading(false) }
  }
  const loadTemplates = async () => {
    try { setTemplates((await dashboardsApi.listTemplates()).map(normalizeTemplate).filter((row) => row.id)) } catch (err) { setError(makeError(err, '模板加载失败')) }
  }
  const openDetail = async (id) => { onNavigate?.({ section: 'detail', id }); setDetailError(''); try { const detail = normalizeDashboard(await dashboardsApi.detail(id)); setDashboards((rows) => [detail, ...rows.filter((row) => row.id !== detail.id)]) } catch (err) { setDetailError(makeError(err, '详情加载失败')) } }
  const openForm = (row) => setDraft(row ? { ...defaultDraft, id: row.id, title: row.title, description: row.description, resourceGroupId: row.resourceGroupId, tags: row.tags.join(', '), variables: row.variables, panels: row.panels, status: row.status } : { ...defaultDraft })
  const saveDraft = async () => { setSaving(true); setError(''); try { const body = dashboardPayload(draft); const saved = normalizeDashboard(draft.id ? await dashboardsApi.update(draft.id, body) : await dashboardsApi.create(body)); setDashboards((rows) => [saved, ...rows.filter((row) => row.id !== saved.id)]); setDraft(null) } catch (err) { setError(makeError(err, '保存失败')) } finally { setSaving(false) } }
  const rowAction = async (action, row) => {
    try {
      if (action === 'edit') return openForm(row)
      if (action === 'clone') {
        const cloned = normalizeDashboard(await dashboardsApi.clone(row.id))
        setDashboards((rows) => [cloned, ...rows])
      }
      if (action === 'share') { setShareTarget(row); return }
      if (action === 'delete') { if (!confirm(`确定删除仪表盘「${row.title}」？`)) return; await dashboardsApi.remove(row.id) }
      if (action === 'export') downloadJson(`${row.title || row.id}.json`, dashboardExportPayload(row))
      if (['delete'].includes(action)) await loadDashboards()
    } catch (err) { setModal({ title: '操作失败', body: makeError(err) }) }
  }
  const confirmShare = async () => {
    if (!shareTarget) return
    try { await dashboardsApi.share(shareTarget.id); setShareTarget(null); await loadDashboards() } catch (err) { setModal({ title: '公开失败', body: makeError(err) }); setShareTarget(null) }
  }
  const batchAction = async (action) => {
    if (action === 'delete' && !confirm(`确定删除选中的 ${selected.length} 个仪表盘？`)) return
    for (const id of selected) { const row = dashboards.find((item) => item.id === id); if (row) { try { if (action === 'clone') await dashboardsApi.clone(id); else if (action === 'share') await dashboardsApi.share(id); else if (action === 'delete') await dashboardsApi.remove(id); else if (action === 'export') downloadJson(`${row.title || row.id}.json`, dashboardExportPayload(row)) } catch { /* continue */ } } }
    setSelected([]); await loadDashboards()
  }
  const importJsonData = async (data) => {
    try {
      const body = { title: data.title || data.name || '导入仪表盘', description: data.description || '', tags: toTags(data.tags || []), variables: data.variables || {}, panels: Array.isArray(data.panels) ? data.panels : [], status: 'active' }
      const saved = normalizeDashboard(await dashboardsApi.create(body))
      setDashboards((rows) => [saved, ...rows])
      setShowImportJson(false)
    } catch (err) { setModal({ title: '导入失败', body: makeError(err) }) }
  }
  const previewTemplate = (tpl) => setModal({ title: `预览：${tpl.title}`, body: displayJson({ variables: tpl.variables, panels: tpl.panels }) })
  const importTemplate = async (tpl) => { try { const saved = normalizeDashboard(await dashboardsApi.importTemplate(tpl.id, { title: tpl.title, variables: tpl.variables, tags: tpl.tags })); setDashboards((rows) => [saved, ...rows]); onNavigate?.({ section: 'detail', id: saved.id }) } catch (err) { setModal({ title: '导入失败', body: makeError(err) }) } }
  const blocked = (title, body) => setModal({ title, body })

  useEffect(() => { loadDashboards(); loadTemplates() }, [])

  if (section === 'templates') {
    return (
      <>
        <TemplatesView templates={templates} loading={loading} error={error} onBack={() => onNavigate?.({ section: 'list' })} onPreview={previewTemplate} onImport={importTemplate} />
        <DashboardOverlays draft={draft} setDraft={setDraft} saving={saving} error={error} onSubmit={saveDraft} onCloseDraft={() => setDraft(null)} modal={modal} onCloseModal={() => setModal(null)} />
      </>
    )
  }
  return (
    <main className='fx-dash-page'>
      {section === 'detail' && active ? (
        <DetailView dashboard={active} variables={variables} panels={panels} detailError={detailError} onBack={() => onNavigate?.({ section: 'list' })} onRefresh={() => openDetail(active.id)} onExport={() => rowAction('export', active)} onShare={() => rowAction('share', active)} onFullscreen={() => document.documentElement.requestFullscreen?.()} onBlocked={blocked} onUpdateDashboard={async (updates) => { try { const body = dashboardPayload({ ...active, ...updates }); const saved = normalizeDashboard(await dashboardsApi.update(active.id, body)); setDashboards((rows) => [saved, ...rows.filter((row) => row.id !== saved.id)]) } catch (err) { setDetailError(makeError(err, '保存失败')) } }} />
      ) : (
        <div className='fx-dash-layout'>
          <Sidebar groups={groups} scope={scope} setScope={setScope} />
          <section className='fx-dash-main'>
            <Toolbar keyword={keyword} setKeyword={setKeyword} loading={loading} selectedCount={selected.length} visibleCols={visibleCols} setVisibleCols={setVisibleCols} onRefresh={loadDashboards} onCreate={() => openForm(null)} onTemplates={() => onNavigate?.({ section: 'templates' })} onImportJson={() => setShowImportJson(true)} onBatch={batchAction} />
            {error && <div className='fx-dash-alert is-error'>{error}</div>}
            <DashboardList rows={filtered} selected={selected} setSelected={setSelected} onOpen={openDetail} onRowAction={rowAction} visibleCols={visibleCols} />
          </section>
        </div>
      )}
      <DashboardOverlays draft={draft} setDraft={setDraft} saving={saving} error={error} onSubmit={saveDraft} onCloseDraft={() => setDraft(null)} modal={modal} onCloseModal={() => setModal(null)} />
      {showImportJson && <ImportJsonModal onClose={() => setShowImportJson(false)} onImport={importJsonData} />}
      {shareTarget && <ShareConfirmModal row={shareTarget} onClose={() => setShareTarget(null)} onConfirm={confirmShare} />}
    </main>
  )
}
