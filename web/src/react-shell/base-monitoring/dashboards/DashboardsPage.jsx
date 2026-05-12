import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import GridLayout from 'react-grid-layout'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/styles.css'
import { dashboardsApi } from '../../api/dashboards.js'
import {
  BLOCKED_BY_CONTRACT,
  blockedContracts,
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
import AddPanelMenu from './AddPanelMenu.jsx'
import AutoRefreshPicker from './AutoRefreshPicker.jsx'
import PanelChart from './PanelChart.jsx'
import PanelEditor from './PanelEditor.jsx'
import TemplateVariablesBar from './TemplateVariablesBar.jsx'
import TimeRangePicker, { QUICK_RANGES, REFRESH_OPTIONS } from './TimeRangePicker.jsx'
import './dashboards.css'

const defaultDraft = {
  title: '',
  description: '',
  workspaceId: '',
  resourceGroupId: '',
  tags: '',
  variables: {},
  panels: [],
  status: 'active',
}

const makeError = (error, fallback = '请求失败') => {
  if (error?.status === 401) return '登录已过期，请重新登录。'
  if (error?.status === 403) return '没有仪表盘权限。'
  if (error?.status >= 500) return `HTTP ${error.status}: 仪表盘服务异常。`
  return displayText(error?.message || fallback)
}

function Toolbar({ keyword, setKeyword, onRefresh, onCreate, onTemplates, selectedCount, onBatch, loading }) {
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
        <button type='button' onClick={onTemplates}>导入</button>
        <select disabled={selectedCount === 0} onChange={(event) => { if (event.target.value) onBatch(event.target.value); event.target.value = '' }}>
          <option value=''>批量操作</option>
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

function DashboardList({ rows, selected, setSelected, onOpen, onRowAction }) {
  return (
    <div className='fx-dash-table'>
      <table>
        <thead>
          <tr><th><input type='checkbox' checked={rows.length > 0 && selected.length === rows.length} onChange={(event) => setSelected(event.target.checked ? rows.map((row) => row.id) : [])} /></th><th>名称</th><th>标签</th><th>说明</th><th>更新时间</th><th>更新人</th><th>共享</th><th>操作</th></tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.id}>
              <td><input type='checkbox' checked={selected.includes(row.id)} onChange={(event) => setSelected(event.target.checked ? [...selected, row.id] : selected.filter((id) => id !== row.id))} /></td>
              <td><button type='button' className='is-link' onClick={() => onOpen(row.id)}>{row.title}</button><small>{row.id}</small></td>
              <td>{row.tags.length ? row.tags.map((tag) => <span className='fx-dash-tag' key={tag}>{tag}</span>) : <span className='muted'>无</span>}</td>
              <td>{row.description || <span className='muted'>无</span>}</td>
              <td>{row.updatedAt || <span className='muted'>-</span>}</td>
              <td>{row.updatedBy || <span className='muted'>-</span>}</td>
              <td><span className={row.shared ? 'fx-dash-state is-on' : 'fx-dash-state'}>{row.shareText}</span></td>
              <td>
                <select onChange={(event) => { if (event.target.value) onRowAction(event.target.value, row); event.target.value = '' }}>
                  <option value=''>更多</option>
                  <option value='edit'>编辑</option>
                  <option value='clone'>克隆</option>
                  <option value='share'>分享</option>
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

function DetailView({ dashboard, variables, panels, onBack, onRefresh, onBlocked, onExport, onShare, onFullscreen, detailError, onUpdateDashboard }) {
  const [timeRangeKey, setTimeRangeKey] = useState('1h')
  const [refreshKey, setRefreshKey] = useState('off')
  const [editingTitle, setEditingTitle] = useState(false)
  const [titleValue, setTitleValue] = useState(dashboard.title)
  const [layout, setLayout] = useState(() => panels.map((p, i) => ({
    i: p.id,
    x: (i % 2) * 12,
    y: Math.floor(i / 2) * 8,
    w: 12,
    h: 8,
  })))
  const [editorPanel, setEditorPanel] = useState(null)
  const [panelList, setPanelList] = useState(panels)
  const [variableValues, setVariableValues] = useState({})
  const containerRef = useRef(null)
  const [containerWidth, setContainerWidth] = useState(1200)
  const refreshTimerRef = useRef(null)
  const datasourceId = 'prometheus-default'

  // 解析仪表盘 JSON 中的 variables 定义
  const dashboardVariables = useMemo(() => {
    const raw = dashboard.raw?.variables || dashboard.variables || {}
    if (Array.isArray(raw)) return raw
    return Object.entries(raw).map(([key, val]) => {
      if (typeof val === 'object' && val !== null) return { name: key, ...val }
      return { name: key, type: 'custom', options: [], current: val }
    })
  }, [dashboard])

  useEffect(() => {
    setPanelList(panels)
    setLayout(panels.map((p, i) => ({
      i: p.id,
      x: (i % 2) * 12,
      y: Math.floor(i / 2) * 8,
      w: 12,
      h: 8,
    })))
  }, [panels])

  useEffect(() => {
    setTitleValue(dashboard.title)
  }, [dashboard.title])

  useEffect(() => {
    if (!containerRef.current) return
    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setContainerWidth(entry.contentRect.width)
      }
    })
    observer.observe(containerRef.current)
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    if (refreshTimerRef.current) clearInterval(refreshTimerRef.current)
    const opt = REFRESH_OPTIONS.find((r) => r.key === refreshKey)
    if (opt && opt.ms) {
      refreshTimerRef.current = setInterval(() => onRefresh(), opt.ms)
    }
    return () => { if (refreshTimerRef.current) clearInterval(refreshTimerRef.current) }
  }, [refreshKey, onRefresh])

  const timeRange = useMemo(() => {
    const now = Math.floor(Date.now() / 1000)
    const range = QUICK_RANGES.find((r) => r.key === timeRangeKey) || QUICK_RANGES[3]
    const duration = range.seconds
    const step = Math.max(15, Math.floor(duration / 240))
    return { start: now - duration, end: now, step }
  }, [timeRangeKey])

  const handleLayoutChange = useCallback((newLayout) => {
    setLayout(newLayout)
  }, [])

  const handleTitleSave = () => {
    setEditingTitle(false)
    if (titleValue !== dashboard.title) {
      onUpdateDashboard?.({ title: titleValue })
    }
  }

  const handlePanelAction = (action, panel) => {
    if (action === 'edit') {
      setEditorPanel(panel)
    } else if (action === 'clone') {
      const cloned = { ...panel, id: 'panel_' + Date.now(), title: panel.title + ' (副本)' }
      const newPanels = [...panelList, cloned]
      setPanelList(newPanels)
      setLayout((prev) => [...prev, { i: cloned.id, x: 0, y: Infinity, w: 12, h: 8 }])
    } else if (action === 'delete') {
      if (confirm('确定删除此面板？')) {
        setPanelList((prev) => prev.filter((p) => p.id !== panel.id))
        setLayout((prev) => prev.filter((l) => l.i !== panel.id))
      }
    }
  }

  const handleEditorSave = (updatedPanel) => {
    const exists = panelList.some((p) => p.id === updatedPanel.id)
    if (exists) {
      setPanelList((prev) => prev.map((p) => p.id === updatedPanel.id ? { ...p, ...updatedPanel, raw: updatedPanel } : p))
    } else {
      setPanelList((prev) => [...prev, { ...updatedPanel, raw: updatedPanel }])
      setLayout((prev) => [...prev, { i: updatedPanel.id, x: 0, y: Infinity, w: 12, h: 8 }])
    }
    setEditorPanel(null)
  }

  const handleSave = () => {
    onUpdateDashboard?.({ title: titleValue, panels: panelList, layout })
  }

  const handleAddPanel = (type) => {
    const newPanel = {
      id: 'panel_' + Date.now(),
      title: '新面板',
      type,
      targets: [{ expr: '', legendFormat: '' }],
      raw: { id: 'panel_' + Date.now(), title: '新面板', type, targets: [{ expr: '', legendFormat: '' }] },
    }
    setEditorPanel(newPanel)
  }

  const handleVariablesChange = useCallback((values) => {
    setVariableValues(values)
  }, [])

  return (
    <main className="fx-dash-detail" ref={containerRef}>
      <header className="fx-dash-detail__head fx-dash-detail__head--sticky">
        <div className="fx-dash-detail__head-left">
          <button type="button" className="fx-dash-back-btn" onClick={onBack}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="15 18 9 12 15 6" /></svg>
            返回
          </button>
          {editingTitle ? (
            <input
              className="fx-dash-title-input"
              value={titleValue}
              onChange={(e) => setTitleValue(e.target.value)}
              onBlur={handleTitleSave}
              onKeyDown={(e) => { if (e.key === 'Enter') handleTitleSave() }}
              autoFocus
            />
          ) : (
            <h1 className="fx-dash-detail__title" onClick={() => setEditingTitle(true)}>{titleValue}</h1>
          )}
          <div className="fx-dash-detail__tags">
            {dashboard.tags.map((tag) => <span className="fx-dash-tag" key={tag}>{tag}</span>)}
          </div>
        </div>
        <div className="fx-dash-detail__head-right">
          <AutoRefreshPicker onRefresh={onRefresh} />
          <TimeRangePicker
            rangeKey={timeRangeKey}
            refreshKey={refreshKey}
            onRangeChange={setTimeRangeKey}
            onRefreshChange={setRefreshKey}
          />
          <AddPanelMenu onSelect={handleAddPanel} />
          <button type="button" className="is-primary" onClick={handleSave}>保存</button>
        </div>
      </header>
      {detailError && <div className="fx-dash-alert is-warning">{detailError}</div>}
      <TemplateVariablesBar
        variables={dashboardVariables}
        onVariablesChange={handleVariablesChange}
      />
      <section className="fx-dash-grid-container">
        {panelList.length > 0 ? (
          <GridLayout
            className="fx-dash-rgl"
            layout={layout}
            cols={24}
            rowHeight={40}
            width={containerWidth - 32}
            draggableHandle=".fx-panel-drag"
            onLayoutChange={handleLayoutChange}
            isResizable={true}
            isDraggable={true}
            margin={[12, 12]}
          >
            {panelList.map((panel) => (
              <div key={panel.id} className="fx-dash-panel">
                <header className="fx-panel-drag">
                  <strong>{panel.title}</strong>
                  <select onChange={(e) => { if (e.target.value) { handlePanelAction(e.target.value, panel); e.target.value = '' } }}>
                    <option value="">...</option>
                    <option value="edit">编辑</option>
                    <option value="clone">克隆</option>
                    <option value="delete">删除</option>
                  </select>
                </header>
                <div className="fx-dash-panel__body">
                  <PanelChart panel={panel} timeRange={timeRange} datasourceId={datasourceId} />
                </div>
              </div>
            ))}
          </GridLayout>
        ) : (
          <div className="fx-dash-empty">暂无面板</div>
        )}
      </section>
      {editorPanel && (
        <PanelEditor
          panel={editorPanel.raw || editorPanel}
          timeRange={timeRange}
          datasourceId={datasourceId}
          onSave={handleEditorSave}
          onClose={() => setEditorPanel(null)}
        />
      )}
    </main>
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

function TemplatesView({ templates, onBack, onPreview, onImport, loading, error }) {
  const [keyword, setKeyword] = useState('')
  const [activeTag, setActiveTag] = useState('全部')
  const [selectedId, setSelectedId] = useState('')
  const tags = useMemo(() => {
    const values = templates.flatMap((tpl) => tpl.tags || []).filter(Boolean)
    return ['全部', ...Array.from(new Set(values)).slice(0, 12)]
  }, [templates])
  const filtered = useMemo(() => templates.filter((tpl) => {
    const words = [tpl.title, tpl.description, (tpl.tags || []).join(' ')].join(' ').toLowerCase()
    const byKeyword = !keyword || words.includes(keyword.trim().toLowerCase())
    const byTag = activeTag === '全部' || (tpl.tags || []).includes(activeTag)
    return byKeyword && byTag
  }), [templates, keyword, activeTag])
  const selectedTemplate = useMemo(
    () => filtered.find((tpl) => tpl.id === selectedId) || filtered[0] || null,
    [filtered, selectedId],
  )
  const variableCount = selectedTemplate ? Object.keys(selectedTemplate.variables || {}).length : 0
  const previewPanels = selectedTemplate?.panels || []

  useEffect(() => {
    if (!selectedId || !filtered.some((tpl) => tpl.id === selectedId)) {
      setSelectedId(filtered[0]?.id || '')
    }
  }, [filtered, selectedId])

  return (
    <main className='fx-dash-templates'>
      <header className='fx-dash-template-head'>
        <div>
          <p>导入</p>
          <h1>仪表盘导入</h1>
          <span>按模板预览变量、Panel 和标签后导入到仪表盘，导入失败会保留错误态。</span>
        </div>
        <button type='button' onClick={onBack}>返回列表</button>
      </header>
      <section className='fx-dash-template-filter'>
        <input value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder='搜索模板名称、标签、说明' />
        <div className='fx-dash-template-tags'>
          {tags.map((tag) => <button type='button' key={tag} className={activeTag === tag ? 'is-active' : ''} onClick={() => setActiveTag(tag)}>{tag}</button>)}
        </div>
      </section>
      {error && <div className='fx-dash-alert is-error'>{error}</div>}
      {loading && <div className='fx-dash-empty'>加载中...</div>}
      <section className='fx-dash-template-workbench'>
        <div className='fx-dash-template-grid' aria-label='仪表盘模板列表'>
          {filtered.map((tpl) => (
            <article className={`fx-dash-template-card${selectedTemplate?.id === tpl.id ? ' is-selected' : ''}`} key={tpl.id}>
              <button type='button' className='fx-dash-template-card__pick' onClick={() => setSelectedId(tpl.id)}>
                <header>
                  <div>
                    <strong>{tpl.title}</strong>
                    <p>{tpl.description || '无说明'}</p>
                  </div>
                  <span className='fx-dash-template-count'>{tpl.panelCount} Panel</span>
                </header>
                <div className='fx-dash-template-meta'>
                  <span>{Object.keys(tpl.variables || {}).length} 变量</span>
                  <span>{(tpl.tags || []).length || 0} 标签</span>
                </div>
                <div className='fx-dash-template-card__tags'>
                  {(tpl.tags || []).length ? tpl.tags.map((tag) => <span className='fx-dash-tag' key={tag}>{tag}</span>) : <span className='muted'>无标签</span>}
                </div>
              </button>
              <footer>
                <button type='button' onClick={() => onPreview(tpl)}>JSON 预览</button>
                <button type='button' className='is-primary' onClick={() => onImport(tpl)}>导入</button>
              </footer>
            </article>
          ))}
        </div>
        <aside className='fx-dash-template-preview' aria-label='模板预览'>
          {selectedTemplate ? (
            <>
              <header>
                <div>
                  <p>预览</p>
                  <h2>{selectedTemplate.title}</h2>
                  <span>{selectedTemplate.description || '无说明'}</span>
                </div>
                <button type='button' className='is-primary' onClick={() => onImport(selectedTemplate)}>导入</button>
              </header>
              <div className='fx-dash-template-summary'>
                <span><b>{selectedTemplate.panelCount}</b><small>Panel</small></span>
                <span><b>{variableCount}</b><small>变量</small></span>
                <span><b>{(selectedTemplate.tags || []).length}</b><small>标签</small></span>
              </div>
              <section>
                <h3>标签</h3>
                <div className='fx-dash-template-card__tags'>
                  {(selectedTemplate.tags || []).length ? selectedTemplate.tags.map((tag) => <span className='fx-dash-tag' key={tag}>{tag}</span>) : <span className='muted'>无标签</span>}
                </div>
              </section>
              <section>
                <h3>Panel</h3>
                <div className='fx-dash-template-panel-list'>
                  {previewPanels.slice(0, 6).map((panel, index) => (
                    <span key={panel.id || panel.title || index}>
                      <b>{displayText(panel.title || panel.name || `Panel ${index + 1}`)}</b>
                      <small>{displayText(panel.type || panel.chartType || 'panel')}</small>
                    </span>
                  ))}
                  {previewPanels.length === 0 && <span className='muted'>暂无 Panel</span>}
                  {previewPanels.length > 6 && <span className='muted'>另有 {previewPanels.length - 6} 个 Panel</span>}
                </div>
              </section>
              <button type='button' onClick={() => onPreview(selectedTemplate)}>查看脱敏 JSON</button>
            </>
          ) : (
            <div className='fx-dash-empty'>选择模板后查看预览</div>
          )}
        </aside>
      </section>
      {!loading && filtered.length === 0 && <div className='fx-dash-empty'>暂无匹配模板</div>}
    </main>
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
  const [scope, setScope] = useState('all')
  const [selected, setSelected] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [detailError, setDetailError] = useState('')
  const [draft, setDraft] = useState(null)
  const [saving, setSaving] = useState(false)
  const [modal, setModal] = useState(null)
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
    const byKeyword = !keyword || text.includes(keyword.toLowerCase())
    const byScope = scope === 'all' || (scope === 'public' ? row.shared : `group:${row.resourceGroupId || '未分组'}` === scope)
    return byKeyword && byScope
  }), [dashboards, keyword, scope])
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
      if (action === 'share') await dashboardsApi.share(row.id)
      if (action === 'delete') await dashboardsApi.remove(row.id)
      if (action === 'export') downloadJson(`${row.title || row.id}.json`, dashboardExportPayload(row))
      if (['share', 'delete'].includes(action)) await loadDashboards()
    } catch (err) { setModal({ title: '操作失败', body: makeError(err) }) }
  }
  const batchAction = async (action) => { for (const id of selected) { const row = dashboards.find((item) => item.id === id); if (row) await rowAction(action, row) } setSelected([]) }
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
            <Toolbar keyword={keyword} setKeyword={setKeyword} loading={loading} selectedCount={selected.length} onRefresh={loadDashboards} onCreate={() => openForm(null)} onTemplates={() => onNavigate?.({ section: 'templates' })} onBatch={batchAction} />
            {error && <div className='fx-dash-alert is-error'>{error}</div>}
            <DashboardList rows={filtered} selected={selected} setSelected={setSelected} onOpen={openDetail} onRowAction={rowAction} />
          </section>
        </div>
      )}
      <DashboardOverlays draft={draft} setDraft={setDraft} saving={saving} error={error} onSubmit={saveDraft} onCloseDraft={() => setDraft(null)} modal={modal} onCloseModal={() => setModal(null)} />
    </main>
  )
}
