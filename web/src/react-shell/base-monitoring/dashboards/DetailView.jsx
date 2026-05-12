import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import GridLayout from 'react-grid-layout'
import AddPanelMenu from './AddPanelMenu.jsx'
import AutoRefreshPicker from './AutoRefreshPicker.jsx'
import ConfirmModal from './ConfirmModal.jsx'
import DashboardLinks from './DashboardLinks.jsx'
import DashboardSettingsModal from './DashboardSettingsModal.jsx'
import PanelChart from './PanelChart.jsx'
import PanelEditor from './PanelEditor.jsx'
import PanelMenu from './PanelMenu.jsx'
import ShareLinkModal from './ShareLinkModal.jsx'
import TemplateVariablesBar, { replaceVariables } from './TemplateVariablesBar.jsx'
import TimeRangePicker, { QUICK_RANGES, REFRESH_OPTIONS, parseTimeRangeFromURL, syncTimeRangeToURL } from './TimeRangePicker.jsx'
import { applyTimezoneOffset, getStoredTimezone } from './TimezoneSelector.jsx'
import { useBeforeUnload, useSaveMode } from './useDashboardHooks.js'
import { usePermission } from './usePermission.jsx'

/** 根据折叠状态过滤可见面板 */
function getVisiblePanels(panelList, collapsedRows) {
  const visible = []
  let currentRowId = null
  for (const panel of panelList) {
    if (panel.type === 'row') { currentRowId = panel.id; continue }
    if (currentRowId && collapsedRows[currentRowId]) continue
    visible.push(panel)
  }
  return visible
}

/** 根据变量值重复 Panel */
function expandRepeatedPanels(panelList, variableValues) {
  const expanded = []
  for (const panel of panelList) {
    const repeatVar = panel.repeat || panel.raw?.repeat
    if (!repeatVar || !variableValues[repeatVar]) { expanded.push(panel); continue }
    const values = Array.isArray(variableValues[repeatVar]) ? variableValues[repeatVar] : [variableValues[repeatVar]]
    if (values.length === 0) { expanded.push(panel); continue }
    for (const val of values) {
      expanded.push({ ...panel, id: `${panel.id}_repeat_${val}`, title: `${panel.title} [${val}]`, _repeatValue: { [repeatVar]: val } })
    }
  }
  return expanded
}

/* DEGRADE-004: 全局 Loading */
function GlobalLoading() {
  return (
    <div className="fx-dash-global-loading">
      <div className="fx-dash-global-loading__spinner" />
      <span>加载中...</span>
    </div>
  )
}

export default function DetailView({ dashboard, variables, panels, onBack, onRefresh, onBlocked, onExport, onShare, onFullscreen, detailError, onUpdateDashboard, loading: externalLoading }) {
  const { canEdit } = usePermission()
  const [timeRange, setTimeRange] = useState(() => parseTimeRangeFromURL() || { start: 'now-1h', end: 'now' })
  const [timezone, setTimezone] = useState(getStoredTimezone)
  const [editingTitle, setEditingTitle] = useState(false)
  const [titleValue, setTitleValue] = useState(dashboard.title)
  const [layout, setLayout] = useState(() => panels.filter((p) => p.type !== 'row').map((p, i) => ({
    i: p.id, x: (i % 2) * 12, y: Math.floor(i / 2) * 8, w: 12, h: 8,
  })))
  const [editorPanel, setEditorPanel] = useState(null)
  const [panelList, setPanelList] = useState(panels)
  const [variableValues, setVariableValues] = useState({})
  const [showSettings, setShowSettings] = useState(false)
  const [showShare, setShowShare] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [collapsedRows, setCollapsedRows] = useState({})
  const [inspectPanel, setInspectPanel] = useState(null)
  const [fullscreenPanel, setFullscreenPanel] = useState(null)
  const [hasChanges, setHasChanges] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState(null)
  const containerRef = useRef(null)
  const [containerWidth, setContainerWidth] = useState(1200)
  const datasourceId = 'prometheus-default'

  /* DEGRADE-001: 保存模式 */
  const doSave = useCallback(() => {
    onUpdateDashboard?.({ title: titleValue, panels: panelList, layout })
    setHasChanges(false)
  }, [titleValue, panelList, layout, onUpdateDashboard])
  const { saveMode, toggleSaveMode, triggerAutoSave } = useSaveMode(doSave)

  /* DEGRADE-002: 离开确认 */
  const { showLeaveConfirm, confirmLeave, handleDiscard, handleCancel } = useBeforeUnload(hasChanges)

  const dashboardVariables = useMemo(() => {
    const raw = dashboard.raw?.variables || dashboard.variables || {}
    if (Array.isArray(raw)) return raw
    return Object.entries(raw).map(([key, val]) => {
      if (typeof val === 'object' && val !== null) return { name: key, ...val }
      return { name: key, type: 'custom', options: [], current: val }
    })
  }, [dashboard])
  const dashboardLinks = useMemo(() => dashboard.raw?.links || [], [dashboard])

  useEffect(() => { setPanelList(panels); setLayout(panels.filter((p) => p.type !== 'row').map((p, i) => ({ i: p.id, x: (i % 2) * 12, y: Math.floor(i / 2) * 8, w: 12, h: 8 }))) }, [panels])
  useEffect(() => { setTitleValue(dashboard.title) }, [dashboard.title])
  useEffect(() => {
    if (!containerRef.current) return
    const observer = new ResizeObserver((entries) => { for (const entry of entries) setContainerWidth(entry.contentRect.width) })
    observer.observe(containerRef.current)
    return () => observer.disconnect()
  }, [])
  useEffect(() => {
    const handleEsc = (e) => { if (e.key === 'Escape' && isFullscreen) { setIsFullscreen(false); document.body.classList.remove('fx-fullscreen') } }
    document.addEventListener('keydown', handleEsc)
    return () => { document.removeEventListener('keydown', handleEsc); document.body.classList.remove('fx-fullscreen') }
  }, [isFullscreen])

  const computedTimeRange = useMemo(() => {
    if (!timeRange) return { start: Math.floor(Date.now() / 1000) - 3600, end: Math.floor(Date.now() / 1000), step: 15 }
    const now = Math.floor(Date.now() / 1000)
    let start = now - 3600
    let end = now
    // 解析 start
    if (typeof timeRange.start === 'string' && timeRange.start.startsWith('now')) {
      const m = /^now(?:-(\d+)([smhdwMy]))?$/.exec(timeRange.start)
      if (m && m[1] && m[2]) {
        const amount = parseInt(m[1], 10)
        const unitMap = { s: 1, m: 60, h: 3600, d: 86400, w: 604800, M: 2592000, y: 31536000 }
        start = now - amount * (unitMap[m[2]] || 60)
      }
    } else if (typeof timeRange.start === 'number') {
      start = timeRange.start
    }
    // 解析 end
    if (typeof timeRange.end === 'string' && timeRange.end === 'now') {
      end = now
    } else if (typeof timeRange.end === 'number') {
      end = timeRange.end
    }
    const step = Math.max(15, Math.floor((end - start) / 240))
    return applyTimezoneOffset({ start, end, step }, timezone)
  }, [timeRange, timezone])

  const markChanged = useCallback(() => { if (!hasChanges) setHasChanges(true); triggerAutoSave() }, [hasChanges, triggerAutoSave])
  const handleLayoutChange = useCallback((newLayout) => { setLayout(newLayout); markChanged() }, [markChanged])
  const handleTitleSave = () => { setEditingTitle(false); if (titleValue !== dashboard.title) { onUpdateDashboard?.({ title: titleValue }); markChanged() } }
  const toggleFullscreen = () => { setIsFullscreen((prev) => { const next = !prev; document.body.classList.toggle('fx-fullscreen', next); return next }) }
  const toggleRow = (rowId) => { setCollapsedRows((prev) => ({ ...prev, [rowId]: !prev[rowId] })) }
  const handleBrushEnd = useCallback((range) => { setTimeRange(range) }, [])
  const handleBack = () => { confirmLeave(onBack) }

  const handlePanelAction = (action, panel) => {
    if (action === 'edit') setEditorPanel(panel)
    else if (action === 'clone') {
      const cloned = { ...panel, id: 'panel_' + Date.now(), title: panel.title + ' (副本)' }
      setPanelList((prev) => [...prev, cloned])
      setLayout((prev) => [...prev, { i: cloned.id, x: 0, y: Infinity, w: 12, h: 8 }])
      markChanged()
    } else if (action === 'delete') {
      /* DEGRADE-019: 自定义确认弹窗 */
      setDeleteConfirm(panel)
    } else if (action === 'inspect') setInspectPanel(panel)
    else if (action === 'fullscreen') setFullscreenPanel(panel)
  }

  const confirmDeletePanel = () => {
    if (!deleteConfirm) return
    setPanelList((prev) => prev.filter((p) => p.id !== deleteConfirm.id))
    setLayout((prev) => prev.filter((l) => l.i !== deleteConfirm.id))
    setDeleteConfirm(null)
    markChanged()
  }

  const handleEditorSave = (updatedPanel) => {
    const exists = panelList.some((p) => p.id === updatedPanel.id)
    if (exists) { setPanelList((prev) => prev.map((p) => p.id === updatedPanel.id ? { ...p, ...updatedPanel, raw: updatedPanel } : p)) }
    else { setPanelList((prev) => [...prev, { ...updatedPanel, raw: updatedPanel }]); setLayout((prev) => [...prev, { i: updatedPanel.id, x: 0, y: Infinity, w: 12, h: 8 }]) }
    setEditorPanel(null)
    markChanged()
  }

  const handleSave = () => { doSave() }
  const handleAddPanel = (type) => {
    if (type === 'row') { setPanelList((prev) => [...prev, { id: 'row_' + Date.now(), title: '新分组', type: 'row' }]); return }
    const id = 'panel_' + Date.now()
    setEditorPanel({ id, title: '新面板', type, targets: [{ expr: '', legendFormat: '' }], raw: { id, title: '新面板', type, targets: [{ expr: '', legendFormat: '' }] } })
  }
  const handleVariablesChange = useCallback((values) => { setVariableValues(values) }, [])
  const handleVariablesUpdate = (updatedVars) => { onUpdateDashboard?.({ variables: updatedVars }) }
  const handleTimeRangeChange = (range) => { setTimeRange(range) }
  /* DEGRADE-004: 全局 Loading */
  if (externalLoading) return <GlobalLoading />

  return (
    <main className="fx-dash-detail" ref={containerRef}>
      <header className="fx-dash-detail__head fx-dash-detail__head--sticky">
        <div className="fx-dash-detail__head-left">
          <button type="button" className="fx-dash-back-btn" onClick={handleBack}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="15 18 9 12 15 6" /></svg>
          </button>
          <span className="fx-dash-breadcrumb">
            <a className="fx-dash-breadcrumb__link" onClick={handleBack}>仪表盘列表</a>
            <span className="fx-dash-breadcrumb__sep">/</span>
          </span>
          {editingTitle ? (
            <input className="fx-dash-title-input" value={titleValue} onChange={(e) => setTitleValue(e.target.value)} onBlur={handleTitleSave} onKeyDown={(e) => { if (e.key === 'Enter') handleTitleSave() }} autoFocus />
          ) : (
            <h1 className="fx-dash-detail__title" onClick={() => canEdit && setEditingTitle(true)}>
              {titleValue}
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="6 9 12 15 18 9" /></svg>
            </h1>
          )}
          <DashboardLinks links={dashboardLinks} />
        </div>
        <div className="fx-dash-detail__head-right">
          {/* 1. 添加面板 */}
          {canEdit && <AddPanelMenu onSelect={handleAddPanel} />}
          {/* 2. 时间范围选择器（含时区） */}
          <TimeRangePicker value={timeRange} onChange={handleTimeRangeChange} timezone={timezone} onTimezoneChange={setTimezone} />
          {/* 3. 自动刷新 */}
          <AutoRefreshPicker onRefresh={onRefresh} />
          {/* 4. 保存模式下拉 */}
          <select className="fx-dash-save-mode" value={saveMode} onChange={(e) => toggleSaveMode(e.target.value)} title="保存模式">
            <option value="manual">手动保存</option>
            <option value="auto">自动保存</option>
          </select>
          {/* 5. 全屏 */}
          <button type="button" className="fx-dash-icon-btn" title="全屏" onClick={toggleFullscreen}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>
          </button>
          {/* 6. 分享 */}
          <button type="button" className="fx-dash-icon-btn" title="分享" onClick={() => setShowShare(true)}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
          </button>
          {/* 7. 保存按钮（仅 hasChanges 时显示） */}
          {hasChanges && saveMode === 'manual' && <button type="button" className="is-primary" onClick={handleSave}>保存</button>}
        </div>
      </header>
      {detailError && <div className="fx-dash-alert is-warning">{detailError}</div>}
      <TemplateVariablesBar variables={dashboardVariables} onVariablesChange={handleVariablesChange} onVariablesUpdate={canEdit ? handleVariablesUpdate : undefined} />
      <section className="fx-dash-grid-container">
        {panelList.some((p) => p.type === 'row') && (
          <div className="fx-dash-rows">
            {panelList.filter((p) => p.type === 'row').map((panel) => (
              <div key={panel.id} className="fx-dash-row-header" onClick={() => toggleRow(panel.id)}>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={collapsedRows[panel.id] ? '' : 'fx-row-expanded'}><polyline points="9 18 15 12 9 6" /></svg>
                <span>{panel.title}</span>
              </div>
            ))}
          </div>
        )}
        {renderGrid()}
      </section>
      {editorPanel && <PanelEditor panel={editorPanel.raw || editorPanel} timeRange={computedTimeRange} datasourceId={datasourceId} dashboardVariables={dashboardVariables} onSave={handleEditorSave} onClose={() => setEditorPanel(null)} />}
      {showSettings && <DashboardSettingsModal dashboard={dashboard} onClose={() => setShowSettings(false)} onSaved={() => onRefresh()} />}
      {showShare && <ShareLinkModal onClose={() => setShowShare(false)} />}
      {inspectPanel && (
        <div className="fx-dash-modal"><div className="fx-dash-modal__body">
          <header><h2>查看数据 - {inspectPanel.title}</h2><button type="button" onClick={() => setInspectPanel(null)}>x</button></header>
          <pre>{JSON.stringify(inspectPanel.raw || inspectPanel, null, 2)}</pre>
        </div></div>
      )}
      {fullscreenPanel && (
        <div className="fx-panel-fullscreen">
          <header><strong>{fullscreenPanel.title}</strong><button type="button" onClick={() => setFullscreenPanel(null)}>x</button></header>
          <div className="fx-panel-fullscreen__body">
            <PanelChart panel={fullscreenPanel} timeRange={computedTimeRange} datasourceId={datasourceId} onBrushEnd={handleBrushEnd} />
          </div>
        </div>
      )}
      {/* DEGRADE-019: 删除确认 Modal */}
      {deleteConfirm && (
        <ConfirmModal title="删除面板" message={`确定删除面板「${deleteConfirm.title}」？此操作不可撤销。`} confirmText="删除" danger onConfirm={confirmDeletePanel} onCancel={() => setDeleteConfirm(null)} />
      )}
      {/* DEGRADE-002: 离开确认 Modal */}
      {showLeaveConfirm && (
        <ConfirmModal title="未保存的修改" message="您有未保存的修改，是否保存后离开？" confirmText="保存" discardText="放弃修改" onConfirm={() => { doSave(); handleDiscard() }} onDiscard={handleDiscard} onCancel={handleCancel} />
      )}
    </main>
  )

  function renderGrid() {
    const visiblePanels = getVisiblePanels(panelList, collapsedRows)
    const expandedPanels = expandRepeatedPanels(visiblePanels, variableValues)
    const annotations = dashboard.raw?.annotations || dashboard.annotations || []
    const expandedLayout = expandedPanels.map((p, i) => layout.find((l) => l.i === p.id) || { i: p.id, x: (i % 2) * 12, y: Math.floor(i / 2) * 8, w: 12, h: 8 })
    if (expandedPanels.length === 0) return <div className="fx-dash-empty">暂无面板</div>
    return (
      <GridLayout className="fx-dash-rgl" layout={expandedLayout} cols={24} rowHeight={40} width={containerWidth - 32} draggableHandle=".fx-panel-drag-handle" onLayoutChange={handleLayoutChange} isResizable isDraggable={canEdit} margin={[12, 12]}>
        {expandedPanels.map((panel) => (
          <div key={panel.id} className="fx-dash-panel">
            <header className="fx-dash-panel__header">
              {canEdit && <span className="fx-panel-drag-handle" title="拖拽移动">&#x2807;&#x2807;</span>}
              <strong>{panel.title}</strong>
              {canEdit && <PanelMenu panel={panel} onAction={handlePanelAction} />}
            </header>
            <div className="fx-dash-panel__body">
              <PanelChart panel={panel} timeRange={computedTimeRange} datasourceId={datasourceId} annotations={annotations} onBrushEnd={handleBrushEnd} />
            </div>
          </div>
        ))}
      </GridLayout>
    )
  }
}
