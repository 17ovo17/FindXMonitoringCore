import React, { useEffect, useRef, useState } from 'react'
import { metricQueryApi } from '../../api/metrics.js'
import { isPermissionError, normalizeList, redactText } from '../../api/http.js'
import { QueryPanel } from './QueryPanel.jsx'
import {
  blockedContracts,
  createPanel,
  csvEscape,
  datasourceId,
  datasourceName,
  datasourceType,
  deleteView,
  displayJson,
  displayText,
  historyKey,
  isPromqlDatasource,
  metricName,
  readHistory,
  readViews,
  relativeRanges,
  resolveRelativeRange,
  resultRows,
  saveHistory,
  saveView,
} from './metricQueryModel.js'
import './metricExplorer.css'

const validSections = new Set(['metrics', 'built-in-metrics', 'objects', 'recording-rules'])

const formatError = (error) => {
  if (isPermissionError(error)) return error.status === 401 ? '登录已过期，请重新登录。' : '没有指标查询权限。'
  if (error?.status === 503) return 'HTTP 503: 下游指标查询不可用或查询语句无效。'
  if (error?.status >= 500) return `HTTP ${error.status}: 指标查询服务异常。`
  return redactText(error?.message || '指标查询失败')
}

const cleanPanel = (panel) => ({ datasource_id: panel.datasourceId, query: panel.query.trim(), timeout_seconds: 10 })
const lastMetricToken = (query) => { const m = String(query || '').match(/[a-zA-Z_:][a-zA-Z0-9_:]*$/); return m?.[0] || '' }

const datasourceLabel = (row) => {
  const name = datasourceName(row)
  const type = datasourceType(row)
  if (String(name || '').toLowerCase().endsWith(` / ${String(type || '').toLowerCase()}`)) return name
  return type ? `${name} / ${type}` : name
}

function QueryBlockedSection({ section }) {
  const titles = { 'built-in-metrics': '内置指标', objects: '对象快捷视图', 'recording-rules': '记录规则' }
  return (
    <main className='fx-query-page'>
      <section className='fx-query-blocked'>
        <strong>BLOCKED_BY_CONTRACT</strong>
        <h1>{titles[section] || '指标查询'}</h1>
        <p>{blockedContracts[section === 'built-in-metrics' ? 'builtInMetrics' : section] || blockedContracts.datasource}</p>
      </section>
    </main>
  )
}
function MetricBrowserModal({ open, onClose, onPick }) {
  const [allMetrics, setAllMetrics] = useState([])
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!open) return
    setLoading(true)
    metricQueryApi.labelValues({ label: '__name__' })
      .then((rows) => setAllMetrics(Array.isArray(rows) ? rows : []))
      .catch(() => setAllMetrics([]))
      .finally(() => setLoading(false))
  }, [open])

  if (!open) return null
  const filtered = search ? allMetrics.filter((m) => String(m).toLowerCase().includes(search.toLowerCase())) : allMetrics

  return (
    <div className='fx-query-modal-overlay' onClick={onClose}>
      <div className='fx-query-modal' onClick={(e) => e.stopPropagation()}>
        <header className='fx-query-modal__head'>
          <h2>指标浏览</h2>
          <button type='button' className='fx-query-close' onClick={onClose}>x</button>
        </header>
        <input className='fx-query-modal__search' value={search} onChange={(e) => setSearch(e.target.value)} placeholder='搜索指标名称' />
        <div className='fx-query-modal__list'>
          {loading && <span className='fx-query-modal__hint'>加载中...</span>}
          {!loading && filtered.length === 0 && <span className='fx-query-modal__hint'>暂无指标</span>}
          {!loading && filtered.slice(0, 200).map((m) => (
            <button type='button' key={m} onClick={() => { onPick(String(m)); onClose() }}>{displayText(String(m))}</button>
          ))}
        </div>
      </div>
    </div>
  )
}

function ViewManager({ onSave, onLoad }) {
  const [showList, setShowList] = useState(false)
  const [views, setViews] = useState([])
  const [saveName, setSaveName] = useState('')
  const [showSave, setShowSave] = useState(false)

  const openList = () => { setViews(readViews()); setShowList(true) }

  return (
    <div className='fx-query-view-manager'>
      <button type='button' onClick={() => setShowSave(!showSave)}>保存视图</button>
      <button type='button' onClick={openList}>加载视图</button>
      {showSave && (
        <div className='fx-query-popover fx-query-view-save'>
          <input value={saveName} onChange={(e) => setSaveName(e.target.value)} placeholder='视图名称' />
          <button type='button' disabled={!saveName.trim()} onClick={() => { onSave(saveName.trim()); setSaveName(''); setShowSave(false) }}>确认保存</button>
        </div>
      )}
      {showList && (
        <div className='fx-query-popover fx-query-view-list'>
          <div className='fx-query-popover__list'>
            {views.length === 0 && <span>暂无已保存视图</span>}
            {views.map((v) => (
              <div key={v.name} className='fx-query-view-item'>
                <button type='button' onClick={() => { onLoad(v); setShowList(false) }}>{v.name}</button>
                <button type='button' className='fx-query-view-del' onClick={() => setViews(deleteView(v.name))}>x</button>
              </div>
            ))}
          </div>
          <button type='button' className='fx-query-view-close' onClick={() => setShowList(false)}>关闭</button>
        </div>
      )}
    </div>
  )
}

export function MetricExplorerPage({ query = {}, onNavigate }) {
  const section = validSections.has(query.section) ? query.section : 'metrics'
  const [datasources, setDatasources] = useState([])
  const [datasource, setDatasource] = useState(String(query.data_source_id || ''))
  const [loadingSources, setLoadingSources] = useState(false)
  const [sourceError, setSourceError] = useState('')
  const [panels, setPanels] = useState([createPanel()])
  const [historyItems, setHistoryItems] = useState([])
  const [showBrowser, setShowBrowser] = useState(false)
  const [activePanel, setActivePanel] = useState(0)
  const inputRefs = useRef(new Map())
  const suggestTimers = useRef(new Map())
  const metricTimers = useRef(new Map())

  const loadHistory = (id = datasource) => setHistoryItems(readHistory(historyKey(id)))

  const loadDatasources = async () => {
    setLoadingSources(true); setSourceError('')
    try {
      const rows = normalizeList(await metricQueryApi.datasources()).filter(isPromqlDatasource)
      setDatasources(rows)
      const queryId = String(query.data_source_id || '')
      const current = rows.find((r) => String(datasourceId(r)) === String(datasource || queryId))
      const next = current || rows.find((r) => r.default) || rows[0]
      if (next) { const id = String(datasourceId(next)); setDatasource(id); loadHistory(id) }
      else { setDatasource(''); setSourceError(blockedContracts.datasource) }
    } catch (e) { setSourceError(formatError(e)) }
    finally { setLoadingSources(false) }
  }

  useEffect(() => { if (section === 'metrics') loadDatasources() }, [section])
  useEffect(() => { loadHistory(datasource) }, [datasource])

  if (section !== 'metrics') return <QueryBlockedSection section={section} />

  const patchPanel = (id, patch) => setPanels((c) => c.map((p) => (p.id === id ? { ...p, ...patch } : p)))
  const queryBody = (panel) => cleanPanel({ ...panel, datasourceId: datasource })

  const runQuery = async (panel) => {
    if (!datasource) { patchPanel(panel.id, { error: '请先选择数据源。' }); return }
    if (!panel.query.trim()) { patchPanel(panel.id, { error: '请先输入 PromQL。' }); return }
    const { start, end } = resolveRelativeRange(panel.relativeRange)
    const minStep = Number(panel.minStep) || 0
    const step = Math.max(minStep || Number(panel.step) || 15, Math.ceil((end - start) / Math.max(Number(panel.maxDataPoints) || 300, 1)))
    patchPanel(panel.id, { loading: true, error: '', instantResult: null, rangeResult: null })
    try {
      const [instantResult, rangeResult] = await Promise.all([
        metricQueryApi.instantQuery({ ...queryBody(panel), time: end }),
        metricQueryApi.rangeQuery({ ...queryBody(panel), start, end, step }),
      ])
      patchPanel(panel.id, { instantResult, rangeResult })
      setHistoryItems(saveHistory(historyKey(datasource), panel.query))
    } catch (e) { patchPanel(panel.id, { error: formatError(e) }) }
    finally { patchPanel(panel.id, { loading: false }) }
  }

  const loadMetrics = (panel, value, asSuggestion = false) => {
    if (!datasource) return
    const timers = asSuggestion ? suggestTimers.current : metricTimers.current
    window.clearTimeout(timers.get(panel.id))
    timers.set(panel.id, window.setTimeout(async () => {
      patchPanel(panel.id, asSuggestion ? {} : { metricsLoading: true })
      try {
        const rows = normalizeList(await metricQueryApi.metrics({ datasource_id: datasource, q: value, limit: asSuggestion ? 8 : 24 }))
        patchPanel(panel.id, asSuggestion ? { suggestions: rows } : { metrics: rows, metricsLoading: false })
      } catch (e) { patchPanel(panel.id, { error: formatError(e), metricsLoading: false, suggestions: [] }) }
    }, asSuggestion ? 260 : 180))
  }

  const onQueryChange = (panel, value) => {
    patchPanel(panel.id, { query: value })
    if (!panel.autocomplete) return
    const token = lastMetricToken(value)
    if (token.length < 1) { patchPanel(panel.id, { suggestions: [] }); return }
    loadMetrics(panel, token, true)
  }

  const insertMetric = (panel, metric) => {
    const text = typeof metric === 'string' ? metric : metricName(metric)
    if (!text) return
    const input = inputRefs.current.get(panel.id)
    const current = panel.query || ''
    const start = input?.selectionStart ?? current.length
    const end = input?.selectionEnd ?? current.length
    const next = `${current.slice(0, start)}${text}${current.slice(end)}`
    patchPanel(panel.id, { query: next, suggestions: [], showMetrics: false })
  }

  const aiGenerate = async (panel) => {
    if (!panel.aiPrompt?.trim()) return
    patchPanel(panel.id, { aiLoading: true, error: '' })
    try {
      const promql = await metricQueryApi.aiGenerateQuery(panel.aiPrompt)
      patchPanel(panel.id, { query: promql, showAiInput: false, aiPrompt: '' })
    } catch (e) {
      const msg = e?.status === 404 || e?.status === 501 ? 'BLOCKED: AI 接口不可用' : formatError(e)
      patchPanel(panel.id, { error: msg })
    } finally { patchPanel(panel.id, { aiLoading: false }) }
  }

  const downloadCsv = (filename, header, rows) => {
    if (!rows.length) return
    const body = [header, ...rows].map((r) => r.map(csvEscape).join(',')).join('\n')
    const url = URL.createObjectURL(new Blob([body], { type: 'text/csv;charset=utf-8' }))
    const a = document.createElement('a'); a.href = url; a.download = filename; a.click()
    URL.revokeObjectURL(url)
  }
  const exportTable = (panel) => downloadCsv(`findx-metric-table-${Date.now()}.csv`, ['metric', 'value', 'time'], resultRows(panel.instantResult, panel.unit).map((r) => [r.metric, r.value, r.time]))
  const exportRange = (panel) => {
    const rows = (panel.rangeResult?.data?.result ?? panel.rangeResult?.result ?? [])
      .flatMap((s, i) => (s.values || []).map((p) => [displayJson(s.metric || `series_${i + 1}`, 600), p?.[0], p?.[1]]))
    downloadCsv(`findx-metric-range-${Date.now()}.csv`, ['metric', 'time', 'value'], rows)
  }

  const chooseDatasource = (value) => {
    setDatasource(value)
    setPanels((c) => c.map((p) => ({ ...p, suggestions: [], metrics: [], instantResult: null, rangeResult: null, error: '' })))
    onNavigate?.({ section: 'metrics', data_source_id: value })
  }

  const handleSaveView = (name) => saveView(name, datasource, panels)
  const handleLoadView = (view) => {
    if (view.datasource) setDatasource(view.datasource)
    if (Array.isArray(view.panels)) setPanels(view.panels.map((snap) => ({ ...createPanel(), ...snap })))
  }

  return (
    <main className='fx-query-page'>
      <header className='fx-query-header'>
        <h1>指标查询</h1>
        <div className='fx-query-header__right'>
          <button type='button' onClick={() => setShowBrowser(true)}>指标浏览</button>
          <ViewManager onSave={handleSaveView} onLoad={handleLoadView} />
          <select className='fx-query-ds-select' value={datasource} onChange={(e) => chooseDatasource(e.target.value)}>
            <option value=''>选择数据源</option>
            {datasources.map((row) => { const id = String(datasourceId(row)); return <option value={id} key={id}>{datasourceLabel(row)}</option> })}
          </select>
          <button type='button' className='fx-query-refresh' disabled={loadingSources} onClick={loadDatasources}>{loadingSources ? '...' : '刷新'}</button>
        </div>
      </header>

      {sourceError && <div className='fx-query-alert is-error'>{sourceError}</div>}

      <MetricBrowserModal open={showBrowser} onClose={() => setShowBrowser(false)}
        onPick={(name) => { const p = panels[activePanel] || panels[0]; if (p) insertMetric(p, name) }} />

      {panels.map((panel, index) => (
        <QueryPanel key={panel.id} panel={panel} index={index} canRemove={panels.length > 1} history={historyItems}
          inputRef={(node) => { if (node) inputRefs.current.set(panel.id, node); else inputRefs.current.delete(panel.id) }}
          onPatch={(patch) => patchPanel(panel.id, patch)}
          onQueryChange={(value) => { setActivePanel(index); onQueryChange(panel, value) }}
          onInsertMetric={(metric) => insertMetric(panel, metric)}
          onLoadMetrics={(value) => loadMetrics(panel, value)}
          onRun={() => runQuery(panel)}
          onRemove={() => setPanels((c) => c.filter((p) => p.id !== panel.id))}
          onExportTable={() => exportTable(panel)}
          onExportRange={() => exportRange(panel)}
          onAiGenerate={() => aiGenerate(panel)}
        />
      ))}

      <button type='button' className='fx-query-add-panel' onClick={() => setPanels((c) => [...c, createPanel()])}>+ 添加面板</button>
    </main>
  )
}