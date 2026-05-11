import React, { useEffect, useRef, useState } from 'react'
import { metricQueryApi } from '../../api/metrics.js'
import { isPermissionError, normalizeList, redactText } from '../../api/http.js'
import { MetricGraph } from './MetricGraph.jsx'
import {
  blockedContracts,
  createPanel,
  csvEscape,
  datasourceId,
  datasourceName,
  datasourceType,
  displayJson,
  displayText,
  historyKey,
  isPromqlDatasource,
  metricDescription,
  metricName,
  metricUnit,
  rangeSeries,
  readHistory,
  relativeRanges,
  resolveRelativeRange,
  resultRows,
  saveHistory,
} from './metricQueryModel.js'
import './metricExplorer.css'

const validSections = new Set(['metrics', 'built-in-metrics', 'objects', 'recording-rules'])

const formatError = (error) => {
  if (isPermissionError(error)) return error.status === 401 ? '登录已过期，请重新登录。' : '没有指标查询权限。'
  if (error?.status === 503) return 'HTTP 503: 下游指标查询不可用或查询语句无效。'
  if (error?.status >= 500) return `HTTP ${error.status}: 指标查询服务异常。`
  return redactText(error?.message || '指标查询失败')
}

const cleanPanel = (panel) => ({
  datasource_id: panel.datasourceId,
  query: panel.query.trim(),
  timeout_seconds: 10,
})

const lastMetricToken = (query) => {
  const match = String(query || '').match(/[a-zA-Z_:][a-zA-Z0-9_:]*$/)
  return match?.[0] || ''
}

const datasourceLabel = (row) => {
  const name = datasourceName(row)
  const type = datasourceType(row)
  const normalizedName = String(name || '').toLowerCase()
  const normalizedType = String(type || '').toLowerCase()
  if (normalizedType && normalizedName.endsWith(` / ${normalizedType}`)) return name
  return type ? `${name} / ${type}` : name
}

function QueryBlockedSection({ section }) {
  const messages = {
    'built-in-metrics': blockedContracts.builtInMetrics,
    objects: blockedContracts.objectViews,
    'recording-rules': blockedContracts.recordingRules,
  }
  const titles = {
    'built-in-metrics': '内置指标',
    objects: '对象快捷视图',
    'recording-rules': '记录规则',
  }
  return (
    <main className='fx-query-page'>
      <section className='fx-query-blocked'>
        <strong>BLOCKED_BY_CONTRACT</strong>
        <h1>{titles[section] || '指标查询'}</h1>
        <p>{messages[section] || blockedContracts.datasource}</p>
      </section>
    </main>
  )
}

function HistoryPopover({ panel, items, onPick, onSearch }) {
  const search = String(panel.historySearch || '').toLowerCase()
  const rows = items.filter((item) => !search || item.toLowerCase().includes(search))
  return (
    <div className='fx-query-popover'>
      <input value={panel.historySearch || ''} onChange={(event) => onSearch(event.target.value)} placeholder='搜索 PromQL 历史记录' />
      <div className='fx-query-popover__list'>
        {rows.map((item) => (
          <button type='button' key={item} onClick={() => onPick(item)}>{displayText(item)}</button>
        ))}
        {rows.length === 0 && <span>暂无历史记录</span>}
      </div>
    </div>
  )
}

function MetricsAssistant({ panel, onSearch, onPick }) {
  return (
    <div className='fx-query-popover fx-query-popover--wide'>
      <div className='fx-query-popover__search'>
        <input value={panel.metricSearch || ''} onChange={(event) => onSearch(event.target.value)} placeholder='搜索指标名称或说明' />
      </div>
      <div className='fx-query-popover__list'>
        {panel.metricsLoading && <span>正在加载指标...</span>}
        {!panel.metricsLoading && panel.metrics?.map((metric) => (
          <button type='button' key={metric.id || metricName(metric)} onClick={() => onPick(metric)}>
            <strong>{metricName(metric)}</strong>
            <small>{metricDescription(metric) || metricUnit(metric) || '-'}</small>
          </button>
        ))}
        {!panel.metricsLoading && (!panel.metrics || panel.metrics.length === 0) && <span>暂无指标记录</span>}
      </div>
    </div>
  )
}

function StatsLine({ result }) {
  if (!result) return null
  return (
    <div className='fx-query-stats'>
      <span>序列 {displayText(result?.stats?.series_count ?? '-')}</span>
      <span>样本 {displayText(result?.stats?.sample_count ?? '-')}</span>
      <span>耗时 {displayText(result?.latency_ms ?? '-')} ms</span>
    </div>
  )
}
function QueryPanel({
  panel,
  index,
  canRemove,
  history,
  onPatch,
  onQueryChange,
  onInsertMetric,
  onLoadMetrics,
  onRun,
  onRemove,
  onExportTable,
  onExportRange,
  inputRef,
}) {
  const rows = resultRows(panel.instantResult, panel.unit)
  const series = rangeSeries(panel.rangeResult)
  const rangeLabel = relativeRanges.find((r) => r.value === panel.relativeRange)?.label || '近 1 小时'

  return (
    <section className='fx-query-card'>
      <header className='fx-query-card__head'>
        <div className='fx-query-card__title'>
          <span className='fx-query-badge'>{index + 1}</span>
          <button type='button' className='fx-query-collapse' onClick={() => onPatch({ collapsed: !panel.collapsed })}>
            {panel.collapsed ? '展开' : '收起'}
          </button>
        </div>
        {canRemove && <button type='button' className='fx-query-close' onClick={onRemove} aria-label='关闭面板'>x</button>}
      </header>

      {!panel.collapsed && (
        <>
          <div className='fx-query-code-wrap'>
            <textarea
              ref={inputRef}
              className='fx-query-code'
              value={panel.query}
              onChange={(event) => onQueryChange(event.target.value)}
              placeholder='up{job="prometheus"}'
              spellCheck={false}
            />
            {panel.suggestions?.length > 0 && (
              <div className='fx-query-suggestions'>
                {panel.suggestions.map((metric) => (
                  <button type='button' key={metric.id || metricName(metric)} onClick={() => onInsertMetric(metric)}>
                    <strong>{metricName(metric)}</strong>
                    <span>{metricDescription(metric)}</span>
                  </button>
                ))}
              </div>
            )}
          </div>

          <div className='fx-query-panel-toolbar'>
            <div className='fx-query-panel-actions'>
              <button type='button' className='is-primary' disabled={panel.loading} onClick={onRun}>
                {panel.loading ? '查询中...' : '执行'}
              </button>
              <button type='button' onClick={() => onPatch({ showHistory: !panel.showHistory, historySearch: '' })}>历史</button>
              <button type='button' onClick={() => { onPatch({ showMetrics: !panel.showMetrics }); onLoadMetrics(panel.metricSearch || '') }}>指标</button>
            </div>
            <div className='fx-query-panel-time'>
              <select value={panel.relativeRange} onChange={(event) => onPatch({ relativeRange: event.target.value })}>
                {relativeRanges.map((r) => <option key={r.value} value={r.value}>{r.label}</option>)}
              </select>
              <label className='fx-query-step-label'>
                <span>Step</span>
                <input type='number' min='1' max='3600' value={panel.step} onChange={(event) => onPatch({ step: Number(event.target.value) })} />
                <span>s</span>
              </label>
            </div>
          </div>

          {panel.showHistory && (
            <HistoryPopover
              panel={panel}
              items={history}
              onPick={(item) => onPatch({ query: item, showHistory: false })}
              onSearch={(value) => onPatch({ historySearch: value })}
            />
          )}
          {panel.showMetrics && (
            <MetricsAssistant
              panel={panel}
              onSearch={(value) => { onPatch({ metricSearch: value }); onLoadMetrics(value) }}
              onPick={(metric) => onInsertMetric(metric)}
            />
          )}

          {panel.error && <div className='fx-query-alert is-error'>{panel.error}</div>}

          <div className='fx-query-tabs' role='tablist'>
            <button type='button' className={panel.activeTab === 'table' ? 'is-active' : ''} onClick={() => onPatch({ activeTab: 'table' })}>表格</button>
            <button type='button' className={panel.activeTab === 'graph' ? 'is-active' : ''} onClick={() => onPatch({ activeTab: 'graph' })}>图表</button>
          </div>

          {panel.activeTab === 'table' && (
            <div className='fx-query-result'>
              <StatsLine result={panel.instantResult} />
              {rows.length > 0 && (
                <div className='fx-query-table-wrap'>
                  <table>
                    <thead><tr><th>Metric</th><th>Value</th><th>Time</th></tr></thead>
                    <tbody>{rows.map((row) => <tr key={`${row.metric}:${row.time}`}><td>{row.metric}</td><td>{row.value}</td><td>{row.time}</td></tr>)}</tbody>
                  </table>
                </div>
              )}
              {rows.length > 0 && <button type='button' className='fx-query-export' onClick={onExportTable}>CSV 导出</button>}
              {!panel.instantResult && !panel.loading && <div className='fx-query-empty'>暂无查询结果</div>}
            </div>
          )}

          {panel.activeTab === 'graph' && (
            <div className='fx-query-result'>
              <div className='fx-query-graph-controls'>
                <span className='fx-query-range-text'>{rangeLabel}</span>
                <select value={panel.graphMode} onChange={(event) => onPatch({ graphMode: event.target.value })}>
                  <option value='line'>Line</option>
                  <option value='area'>StackArea</option>
                </select>
                {panel.rangeResult && <button type='button' className='fx-query-export' onClick={onExportRange}>CSV 导出</button>}
              </div>
              <StatsLine result={panel.rangeResult} />
              <MetricGraph result={panel.rangeResult} mode={panel.graphMode} />
              {series.length > 0 && <div className='fx-query-series'>{series.map((item) => <span key={item.name}>{item.name}: {item.points} points</span>)}</div>}
            </div>
          )}
        </>
      )}
    </section>
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
  const inputRefs = useRef(new Map())
  const suggestTimers = useRef(new Map())
  const metricTimers = useRef(new Map())

  const loadHistory = (id = datasource) => setHistoryItems(readHistory(historyKey(id)))

  const loadDatasources = async () => {
    setLoadingSources(true)
    setSourceError('')
    try {
      const rows = normalizeList(await metricQueryApi.datasources()).filter(isPromqlDatasource)
      setDatasources(rows)
      const queryId = String(query.data_source_id || '')
      const current = rows.find((row) => String(datasourceId(row)) === String(datasource || queryId))
      const next = current || rows.find((row) => row.default) || rows[0]
      if (next) {
        const id = String(datasourceId(next))
        setDatasource(id)
        loadHistory(id)
      } else {
        setDatasource('')
        setSourceError(blockedContracts.datasource)
      }
    } catch (error) {
      setSourceError(formatError(error))
    } finally {
      setLoadingSources(false)
    }
  }

  useEffect(() => { if (section === 'metrics') loadDatasources() }, [section])
  useEffect(() => { loadHistory(datasource) }, [datasource])

  if (section !== 'metrics') return <QueryBlockedSection section={section} />

  const patchPanel = (id, patch) => {
    setPanels((current) => current.map((panel) => (panel.id === id ? { ...panel, ...patch } : panel)))
  }

  const ensureRunnable = (panel) => {
    if (!datasource) {
      patchPanel(panel.id, { error: '请先选择数据源。' })
      return false
    }
    if (!panel.query.trim()) {
      patchPanel(panel.id, { error: '请先输入 PromQL。' })
      return false
    }
    return true
  }
  const queryBody = (panel) => cleanPanel({ ...panel, datasourceId: datasource })

  const runQuery = async (panel) => {
    if (!ensureRunnable(panel)) return
    const { start, end } = resolveRelativeRange(panel.relativeRange)
    const step = Math.max(Number(panel.step) || 15, Math.ceil((end - start) / Math.max(Number(panel.maxDataPoints) || 600, 1)))
    patchPanel(panel.id, { loading: true, error: '', instantResult: null, rangeResult: null })
    try {
      const [instantResult, rangeResult] = await Promise.all([
        metricQueryApi.instantQuery({ ...queryBody(panel), time: end }),
        metricQueryApi.rangeQuery({ ...queryBody(panel), start, end, step }),
      ])
      patchPanel(panel.id, { instantResult, rangeResult })
      setHistoryItems(saveHistory(historyKey(datasource), panel.query))
    } catch (error) {
      patchPanel(panel.id, { error: formatError(error) })
    } finally {
      patchPanel(panel.id, { loading: false })
    }
  }

  const loadMetrics = (panel, value, asSuggestion = false) => {
    if (!datasource) {
      patchPanel(panel.id, { error: '请先选择数据源。' })
      return
    }
    const timers = asSuggestion ? suggestTimers.current : metricTimers.current
    window.clearTimeout(timers.get(panel.id))
    timers.set(panel.id, window.setTimeout(async () => {
      patchPanel(panel.id, asSuggestion ? {} : { metricsLoading: true })
      try {
        const rows = normalizeList(await metricQueryApi.metrics({ datasource_id: datasource, q: value, limit: asSuggestion ? 8 : 24 }))
        patchPanel(panel.id, asSuggestion ? { suggestions: rows } : { metrics: rows, metricsLoading: false })
      } catch (error) {
        patchPanel(panel.id, { error: formatError(error), metricsLoading: false, suggestions: [] })
      }
    }, asSuggestion ? 260 : 180))
  }

  const onQueryChange = (panel, value) => {
    patchPanel(panel.id, { query: value })
    const token = lastMetricToken(value)
    if (token.length < 1) {
      patchPanel(panel.id, { suggestions: [] })
      return
    }
    loadMetrics(panel, token, true)
  }

  const insertMetric = (panel, metric) => {
    const text = metricName(metric)
    if (!text) return
    const input = inputRefs.current.get(panel.id)
    const current = panel.query || ''
    const start = input?.selectionStart ?? current.length
    const end = input?.selectionEnd ?? current.length
    const next = `${current.slice(0, start)}${text}${current.slice(end)}`
    patchPanel(panel.id, { query: next, suggestions: [], showMetrics: false })
    window.requestAnimationFrame(() => {
      const nextInput = inputRefs.current.get(panel.id)
      nextInput?.focus()
      nextInput?.setSelectionRange(start + text.length, start + text.length)
    })
  }
  const downloadCsv = (filename, header, rows) => {
    if (!rows.length) return
    const body = [header, ...rows].map((row) => row.map(csvEscape).join(',')).join('\n')
    const url = URL.createObjectURL(new Blob([body], { type: 'text/csv;charset=utf-8' }))
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    URL.revokeObjectURL(url)
  }

  const exportTable = (panel) => downloadCsv(
    `findx-metric-table-${Date.now()}.csv`,
    ['metric', 'value', 'time'],
    resultRows(panel.instantResult, panel.unit).map((row) => [row.metric, row.value, row.time]),
  )

  const exportRange = (panel) => {
    const rows = (panel.rangeResult?.data?.result ?? panel.rangeResult?.result ?? [])
      .flatMap((series, index) => (series.values || []).map((point) => [displayJson(series.metric || `series_${index + 1}`, 600), point?.[0], point?.[1]]))
    downloadCsv(`findx-metric-range-${Date.now()}.csv`, ['metric', 'time', 'value'], rows)
  }

  const chooseDatasource = (value) => {
    setDatasource(value)
    setPanels((current) => current.map((panel) => ({ ...panel, suggestions: [], metrics: [], instantResult: null, rangeResult: null, error: '' })))
    onNavigate?.({ section: 'metrics', data_source_id: value })
  }

  return (
    <main className='fx-query-page'>
      <header className='fx-query-header'>
        <h1>指标查询</h1>
        <div className='fx-query-header__right'>
          <select className='fx-query-ds-select' value={datasource} onChange={(event) => chooseDatasource(event.target.value)}>
            <option value=''>选择数据源</option>
            {datasources.map((row) => {
              const id = String(datasourceId(row))
              return <option value={id} key={id}>{datasourceLabel(row)}</option>
            })}
          </select>
          <button type='button' className='fx-query-refresh' disabled={loadingSources} onClick={loadDatasources}>
            {loadingSources ? '...' : '刷新'}
          </button>
        </div>
      </header>

      {sourceError && <div className='fx-query-alert is-error'>{sourceError}</div>}

      {panels.map((panel, index) => (
        <QueryPanel
          key={panel.id}
          panel={panel}
          index={index}
          canRemove={panels.length > 1}
          history={historyItems}
          inputRef={(node) => { if (node) inputRefs.current.set(panel.id, node); else inputRefs.current.delete(panel.id) }}
          onPatch={(patch) => patchPanel(panel.id, patch)}
          onQueryChange={(value) => onQueryChange(panel, value)}
          onInsertMetric={(metric) => insertMetric(panel, metric)}
          onLoadMetrics={(value) => loadMetrics(panel, value)}
          onRun={() => runQuery(panel)}
          onRemove={() => setPanels((current) => current.filter((item) => item.id !== panel.id))}
          onExportTable={() => exportTable(panel)}
          onExportRange={() => exportRange(panel)}
        />
      ))}

      <button type='button' className='fx-query-add-panel' onClick={() => setPanels((current) => [...current, createPanel()])}>
        + 添加查询
      </button>
    </main>
  )
}
