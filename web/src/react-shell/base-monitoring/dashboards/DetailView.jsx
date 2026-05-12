import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import GridLayout from 'react-grid-layout'
import AddPanelMenu from './AddPanelMenu.jsx'
import AutoRefreshPicker from './AutoRefreshPicker.jsx'
import DashboardSettingsModal from './DashboardSettingsModal.jsx'
import PanelChart from './PanelChart.jsx'
import PanelEditor from './PanelEditor.jsx'
import PanelMenu from './PanelMenu.jsx'
import ShareLinkModal from './ShareLinkModal.jsx'
import TemplateVariablesBar from './TemplateVariablesBar.jsx'
import TimeRangePicker, { QUICK_RANGES, REFRESH_OPTIONS, parseTimeRangeFromURL, syncTimeRangeToURL } from './TimeRangePicker.jsx'

/** D09: 根据折叠状态过滤可见面板 */
function getVisiblePanels(panelList, collapsedRows) {
  const visible = []
  let currentRowId = null
  for (const panel of panelList) {
    if (panel.type === 'row') {
      currentRowId = panel.id
      continue
    }
    if (currentRowId && collapsedRows[currentRowId]) continue
    visible.push(panel)
  }
  return visible
}

export default function DetailView({ dashboard, variables, panels, onBack, onRefresh, onBlocked, onExport, onShare, onFullscreen, detailError, onUpdateDashboard }) {
  const [timeRangeKey, setTimeRangeKey] = useState(() => parseTimeRangeFromURL() || '1h')
  const [refreshKey, setRefreshKey] = useState('off')
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
  const containerRef = useRef(null)
  const [containerWidth, setContainerWidth] = useState(1200)
  const refreshTimerRef = useRef(null)
  const datasourceId = 'prometheus-default'
/* PLACEHOLDER_DETAIL_VIEW_BODY */

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
    setLayout(panels.filter((p) => p.type !== 'row').map((p, i) => ({
      i: p.id, x: (i % 2) * 12, y: Math.floor(i / 2) * 8, w: 12, h: 8,
    })))
  }, [panels])

  useEffect(() => { setTitleValue(dashboard.title) }, [dashboard.title])

  useEffect(() => {
    if (!containerRef.current) return
    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) setContainerWidth(entry.contentRect.width)
    })
    observer.observe(containerRef.current)
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    if (refreshTimerRef.current) clearInterval(refreshTimerRef.current)
    const opt = REFRESH_OPTIONS.find((r) => r.key === refreshKey)
    if (opt && opt.ms) refreshTimerRef.current = setInterval(() => onRefresh(), opt.ms)
    return () => { if (refreshTimerRef.current) clearInterval(refreshTimerRef.current) }
  }, [refreshKey, onRefresh])

  const timeRange = useMemo(() => {
    const now = Math.floor(Date.now() / 1000)
    const range = QUICK_RANGES.find((r) => r.key === timeRangeKey) || QUICK_RANGES[3]
    const step = Math.max(15, Math.floor(range.seconds / 240))
    return { start: now - range.seconds, end: now, step }
  }, [timeRangeKey])

  const handleLayoutChange = useCallback((newLayout) => { setLayout(newLayout) }, [])
  const handleTitleSave = () => { setEditingTitle(false); if (titleValue !== dashboard.title) onUpdateDashboard?.({ title: titleValue }) }

  const toggleFullscreen = () => {
    setIsFullscreen((prev) => {
      const next = !prev
      document.body.classList.toggle('fx-fullscreen', next)
      return next
    })
  }

  useEffect(() => {
    const handleEsc = (e) => { if (e.key === 'Escape' && isFullscreen) { setIsFullscreen(false); document.body.classList.remove('fx-fullscreen') } }
    document.addEventListener('keydown', handleEsc)
    return () => { document.removeEventListener('keydown', handleEsc); document.body.classList.remove('fx-fullscreen') }
  }, [isFullscreen])

  const toggleRow = (rowId) => { setCollapsedRows((prev) => ({ ...prev, [rowId]: !prev[rowId] })) }

  const handlePanelAction = (action, panel) => {
    if (action === 'edit') setEditorPanel(panel)
    else if (action === 'clone') {
      const cloned = { ...panel, id: 'panel_' + Date.now(), title: panel.title + ' (副本)' }
      setPanelList((prev) => [...prev, cloned])
      setLayout((prev) => [...prev, { i: cloned.id, x: 0, y: Infinity, w: 12, h: 8 }])
    } else if (action === 'delete') { if (confirm('确定删除此面板？')) { setPanelList((prev) => prev.filter((p) => p.id !== panel.id)); setLayout((prev) => prev.filter((l) => l.i !== panel.id)) } }
    else if (action === 'inspect') setInspectPanel(panel)
    else if (action === 'fullscreen') setFullscreenPanel(panel)
  }

  const handleEditorSave = (updatedPanel) => {
    const exists = panelList.some((p) => p.id === updatedPanel.id)
    if (exists) setPanelList((prev) => prev.map((p) => p.id === updatedPanel.id ? { ...p, ...updatedPanel, raw: updatedPanel } : p))
    else { setPanelList((prev) => [...prev, { ...updatedPanel, raw: updatedPanel }]); setLayout((prev) => [...prev, { i: updatedPanel.id, x: 0, y: Infinity, w: 12, h: 8 }]) }
    setEditorPanel(null)
  }

  const handleSave = () => { onUpdateDashboard?.({ title: titleValue, panels: panelList, layout }) }
  const handleAddPanel = (type) => {
    const id = 'panel_' + Date.now()
    setEditorPanel({ id, title: '新面板', type, targets: [{ expr: '', legendFormat: '' }], raw: { id, title: '新面板', type, targets: [{ expr: '', legendFormat: '' }] } })
  }
  const handleVariablesChange = useCallback((values) => { setVariableValues(values) }, [])
/* PLACEHOLDER_DETAIL_VIEW_RENDER */

  return (
    <main className="fx-dash-detail" ref={containerRef}>
      <header className="fx-dash-detail__head fx-dash-detail__head--sticky">
        <div className="fx-dash-detail__head-left">
          <button type="button" className="fx-dash-back-btn" onClick={onBack}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="15 18 9 12 15 6" /></svg>
            返回
          </button>
          {editingTitle ? (
            <input className="fx-dash-title-input" value={titleValue} onChange={(e) => setTitleValue(e.target.value)} onBlur={handleTitleSave} onKeyDown={(e) => { if (e.key === 'Enter') handleTitleSave() }} autoFocus />
          ) : (
            <h1 className="fx-dash-detail__title" onClick={() => setEditingTitle(true)}>{titleValue}</h1>
          )}
          <div className="fx-dash-detail__tags">{dashboard.tags.map((tag) => <span className="fx-dash-tag" key={tag}>{tag}</span>)}</div>
        </div>
        <div className="fx-dash-detail__head-right">
          <AutoRefreshPicker onRefresh={onRefresh} />
          <TimeRangePicker rangeKey={timeRangeKey} refreshKey={refreshKey} onRangeChange={setTimeRangeKey} onRefreshChange={setRefreshKey} />
          <AddPanelMenu onSelect={handleAddPanel} />
          <button type="button" className="fx-dash-icon-btn" title="全屏" onClick={toggleFullscreen}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>
          </button>
          <button type="button" className="fx-dash-icon-btn" title="分享" onClick={() => setShowShare(true)}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
          </button>
          <button type="button" className="fx-dash-icon-btn" title="设置" onClick={() => setShowSettings(true)}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          </button>
          <button type="button" className="is-primary" onClick={handleSave}>保存</button>
        </div>
      </header>
      {detailError && <div className="fx-dash-alert is-warning">{detailError}</div>}
      <TemplateVariablesBar variables={dashboardVariables} onVariablesChange={handleVariablesChange} />
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
        {(() => {
          const visiblePanels = getVisiblePanels(panelList, collapsedRows)
          if (visiblePanels.length === 0) return <div className="fx-dash-empty">暂无面板</div>
          return (
            <GridLayout className="fx-dash-rgl" layout={layout.filter((l) => visiblePanels.some((p) => p.id === l.i))} cols={24} rowHeight={40} width={containerWidth - 32} draggableHandle=".fx-panel-drag" onLayoutChange={handleLayoutChange} isResizable isDraggable margin={[12, 12]}>
              {visiblePanels.map((panel) => (
                <div key={panel.id} className="fx-dash-panel">
                  <header className="fx-panel-drag"><strong>{panel.title}</strong><PanelMenu panel={panel} onAction={handlePanelAction} /></header>
                  <div className="fx-dash-panel__body"><PanelChart panel={panel} timeRange={timeRange} datasourceId={datasourceId} /></div>
                </div>
              ))}
            </GridLayout>
          )
        })()}
      </section>
      {editorPanel && <PanelEditor panel={editorPanel.raw || editorPanel} timeRange={timeRange} datasourceId={datasourceId} onSave={handleEditorSave} onClose={() => setEditorPanel(null)} />}
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
          <div className="fx-panel-fullscreen__body"><PanelChart panel={fullscreenPanel} timeRange={timeRange} datasourceId={datasourceId} /></div>
        </div>
      )}
    </main>
  )
}
