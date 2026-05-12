import React, { useState } from 'react'
import Editor from '@monaco-editor/react'
import { MetricGraph } from './MetricGraph.jsx'
import {
  displayJson,
  displayText,
  metricDescription,
  metricName,
  metricUnit,
  nowLocalInput,
  rangeSeries,
  relativeRanges,
  resultRows,
  UNITS,
} from './metricQueryModel.js'

/* GAP-M15: QueryStats 显示在 Tab 栏右侧 */
function QueryStats({ result }) {
  if (!result) return null
  const loadTime = result?.latency_ms ?? '-'
  const resolution = result?.stats?.resolution ?? null
  const series = result?.stats?.series_count ?? '-'
  return (
    <span className='fx-query-stats-inline'>
      Load time: {displayText(loadTime)}ms
      {resolution ? ` | Resolution: ${resolution}s` : ''}
      {' | '}Result series: {displayText(series)}
    </span>
  )
}

/* GAP-M06: 历史记录浮层（绝对定位 + 标题栏） */
function HistoryPopover({ panel, items, onPick, onSearch, onClose }) {
  const search = String(panel.historySearch || '').toLowerCase()
  const rows = items.filter((item) => !search || item.toLowerCase().includes(search))
  return (
    <div className='fx-query-float-popover'>
      <div className='fx-query-float-popover__head'>
        <strong>历史记录</strong>
        <button type='button' onClick={onClose}>x</button>
      </div>
      <input value={panel.historySearch || ''} onChange={(e) => onSearch(e.target.value)} placeholder='搜索历史记录' />
      <div className='fx-query-float-popover__list'>
        {rows.map((item) => (
          <button type='button' key={item} onClick={() => onPick(item)}>{displayText(item)}</button>
        ))}
        {rows.length === 0 && <span className='fx-query-float-popover__empty'>暂无历史记录</span>}
      </div>
    </div>
  )
}

/* GAP-M02: 内置指标 Dropdown 浮层（Monaco 左侧 addon 触发） */
function MetricsDropdown({ panel, onSearch, onPick, onClose }) {
  return (
    <div className='fx-query-float-popover fx-query-float-popover--wide'>
      <div className='fx-query-float-popover__head'>
        <strong>内置指标</strong>
        <button type='button' onClick={onClose}>x</button>
      </div>
      <input value={panel.metricSearch || ''} onChange={(e) => onSearch(e.target.value)} placeholder='搜索内置指标' />
      <div className='fx-query-float-popover__list'>
        {panel.metricsLoading && <span className='fx-query-float-popover__empty'>正在加载...</span>}
        {!panel.metricsLoading && panel.metrics?.map((metric) => (
          <button type='button' key={metric.id || metricName(metric)} onClick={() => onPick(metric)}>
            <strong>{metricName(metric)}</strong>
            <small>{metricDescription(metric) || metricUnit(metric) || '-'}</small>
          </button>
        ))}
        {!panel.metricsLoading && (!panel.metrics || panel.metrics.length === 0) && (
          <span className='fx-query-float-popover__empty'>暂无指标</span>
        )}
      </div>
    </div>
  )
}

/* GAP-M12: Table 展示改为 List 形式 + metric 语法高亮 */
function MetricLabel({ metricJson }) {
  // 解析 metric JSON 字符串，提取 __name__ 和 labels
  try {
    const obj = typeof metricJson === 'string' ? JSON.parse(metricJson) : metricJson
    if (typeof obj === 'object' && obj !== null) {
      const name = obj.__name__ || ''
      const labels = Object.entries(obj)
        .filter(([k]) => k !== '__name__')
        .map(([k, v], i, arr) => (
          <span key={k}>
            <span className='fx-metric-label-key'>{k}</span>
            ={'"'}{v}{'"'}{i < arr.length - 1 ? ', ' : ''}
          </span>
        ))
      return (
        <span className='fx-metric-label'>
          <span className='fx-metric-name'>{name}</span>
          <span className='fx-metric-bracket'>{'{'}</span>
          {labels}
          <span className='fx-metric-bracket'>{'}'}</span>
        </span>
      )
    }
  } catch { /* fallback */ }
  return <span className='fx-metric-label'>{displayText(String(metricJson))}</span>
}

export function QueryPanel({
  panel, canRemove, history, datasources, datasource,
  onPatch, onQueryChange, onInsertMetric, onLoadMetrics,
  onRun, onRemove, onExportTable, onExportRange, onAiGenerate,
  onDatasourceChange, onOpenBrowser, inputRef,
}) {
  const [showMetrics, setShowMetrics] = useState(false)
  const [showHistory, setShowHistory] = useState(false)
  const rows = resultRows(panel.instantResult, panel.unit)
  const series = rangeSeries(panel.rangeResult)
  const activeResult = panel.activeTab === 'table' ? panel.instantResult : panel.rangeResult

  return (
    <section className='fx-query-card' style={{ maxHeight: 650, overflow: 'auto' }}>
      {/* GAP-M14: 关闭按钮右上角绝对定位 */}
      {canRemove && (
        <button type='button' className='fx-query-card__close' onClick={onRemove} aria-label='关闭面板'>
          <svg width='14' height='14' viewBox='0 0 14 14'><path d='M1 1l12 12M13 1L1 13' stroke='currentColor' strokeWidth='1.5' strokeLinecap='round'/></svg>
        </button>
      )}

      {/* GAP-M14 第一行: DatasourceSelect + autocomplete */}
      <div className='fx-query-row-ds'>
        <select className='fx-query-ds-select' value={datasource} onChange={(e) => onDatasourceChange(e.target.value)}>
          <option value=''>选择数据源</option>
          {datasources.map((row) => {
            const id = String(row.id ?? row.datasource_id ?? row.value ?? '')
            const label = row._label || id
            return <option value={id} key={id}>{label}</option>
          })}
        </select>
        <label className='fx-query-autocomplete-label'>
          <input type='checkbox' checked={panel.autocomplete} onChange={(e) => onPatch({ autocomplete: e.target.checked })} />
          <span>autocomplete</span>
        </label>
      </div>

      {/* GAP-M14 第二行: 内置指标addon + Monaco + 全局指标icon + Execute */}
      <div className='fx-query-row-editor'>
        {/* GAP-M02: 内置指标按钮作为左侧 addon */}
        <div className='fx-query-editor-addon'>
          <button type='button' className='fx-query-addon-btn' title='内置指标'
            onClick={() => { setShowMetrics(!showMetrics); onLoadMetrics(panel.metricSearch || '') }}>
            <svg width='16' height='16' viewBox='0 0 16 16'><path d='M2 12V4l4 3 4-5 4 4v6H2z' fill='none' stroke='currentColor' strokeWidth='1.2'/></svg>
          </button>
          {showMetrics && (
            <MetricsDropdown panel={panel}
              onSearch={(v) => { onPatch({ metricSearch: v }); onLoadMetrics(v) }}
              onPick={(metric) => { onInsertMetric(metric); setShowMetrics(false) }}
              onClose={() => setShowMetrics(false)} />
          )}
        </div>

        {/* Monaco Editor */}
        <div className='fx-query-editor-main'>
          <Editor
            height='80px'
            defaultLanguage='promql'
            value={panel.query}
            onChange={(val) => onQueryChange(val || '')}
            options={{
              minimap: { enabled: false },
              lineNumbers: 'off',
              scrollBeyondLastLine: false,
              wordWrap: 'on',
              fontSize: 13,
              fontFamily: 'Menlo, Monaco, Consolas, monospace',
              overviewRulerLanes: 0,
              renderLineHighlight: 'none',
              scrollbar: { vertical: 'hidden', horizontal: 'auto' },
              padding: { top: 8, bottom: 8 },
            }}
            theme='vs-light'
          />
        </div>

        {/* GAP-M03: 全局指标浏览器 suffix icon */}
        <button type='button' className='fx-query-addon-btn' title='指标浏览' onClick={onOpenBrowser}>
          <svg width='16' height='16' viewBox='0 0 16 16'><circle cx='8' cy='8' r='6.5' fill='none' stroke='currentColor' strokeWidth='1.2'/><path d='M1.5 8h13M8 1.5c-2 2-2 11 0 13M8 1.5c2 2 2 11 0 13' fill='none' stroke='currentColor' strokeWidth='1'/></svg>
        </button>

        {/* GAP-M04: Execute 按钮与 Monaco 同行右侧 */}
        <button type='button' className='fx-query-execute-btn' disabled={panel.loading} onClick={onRun}>
          {panel.loading ? '查询中...' : 'Execute'}
        </button>
      </div>

      {/* GAP-M06: 历史记录浮层 */}
      {showHistory && (
        <HistoryPopover panel={panel} items={history}
          onPick={(item) => { onPatch({ query: item }); setShowHistory(false) }}
          onSearch={(v) => onPatch({ historySearch: v })}
          onClose={() => setShowHistory(false)} />
      )}

      {/* 辅助按钮行: 历史 + AI */}
      <div className='fx-query-row-aux'>
        <button type='button' className='fx-query-aux-btn' onClick={() => setShowHistory(!showHistory)}>历史</button>
        <button type='button' className='fx-query-aux-btn' onClick={() => onPatch({ showAiInput: !panel.showAiInput })}>AI</button>
      </div>

      {panel.showAiInput && (
        <div className='fx-query-ai-input'>
          <input value={panel.aiPrompt || ''} onChange={(e) => onPatch({ aiPrompt: e.target.value })}
            placeholder='用自然语言描述查询' />
          <button type='button' disabled={panel.aiLoading || !panel.aiPrompt?.trim()}
            onClick={onAiGenerate}>{panel.aiLoading ? '生成中...' : '生成 PromQL'}</button>
        </div>
      )}

      {panel.error && <div className='fx-query-alert is-error'>{panel.error}</div>}

      {/* GAP-M05 + M15: Card 类型 Tabs + QueryStats 在右侧 */}
      <div className='fx-query-tabs-card'>
        <div className='fx-query-tabs-card__bar'>
          <button type='button' className={panel.activeTab === 'table' ? 'is-active' : ''} onClick={() => onPatch({ activeTab: 'table' })}>Table</button>
          <button type='button' className={panel.activeTab === 'graph' ? 'is-active' : ''} onClick={() => onPatch({ activeTab: 'graph' })}>Graph</button>
          <div className='fx-query-tabs-card__extra'>
            <QueryStats result={activeResult} />
          </div>
        </div>

        <div className='fx-query-tabs-card__body'>
          {/* TABLE TAB */}
          {panel.activeTab === 'table' && (
            <div className='fx-query-table-panel'>
              <div className='fx-query-table-controls'>
                <label className='fx-query-input-group'>
                  <span className='fx-query-input-group__addon'>Time</span>
                  <input type='datetime-local' value={panel.instantTime || nowLocalInput()}
                    onChange={(e) => onPatch({ instantTime: e.target.value })} />
                </label>
                <label className='fx-query-input-group'>
                  <span className='fx-query-input-group__addon'>Unit</span>
                  <select value={panel.unit} onChange={(e) => onPatch({ unit: e.target.value })}>
                    {UNITS.map((u) => <option key={u.value} value={u.value}>{u.label}</option>)}
                  </select>
                </label>
                {rows.length > 0 && <button type='button' className='fx-query-export' onClick={onExportTable}>CSV</button>}
              </div>
              {rows.length > 1000 && <div className='fx-query-alert is-warning'>结果超过 1000 行，仅显示前 1000 条。</div>}
              {/* GAP-M12: List 形式展示 */}
              <div className='fx-query-metric-list'>
                {rows.slice(0, 1000).map((row, i) => (
                  <div className='fx-query-metric-list__item' key={i}>
                    <div className='fx-query-metric-list__label'><MetricLabel metricJson={row.metric} /></div>
                    <div className='fx-query-metric-list__value'>{row.value}</div>
                  </div>
                ))}
                {!panel.instantResult && !panel.loading && <div className='fx-query-empty'>暂无查询结果</div>}
              </div>
            </div>
          )}

          {/* GRAPH TAB */}
          {panel.activeTab === 'graph' && (
            <div className='fx-query-graph-panel'>
              {/* GAP-M11 + M08 + M18: Graph 控制栏 */}
              <div className='fx-query-graph-controls'>
                <select className='fx-query-time-select' value={panel.relativeRange} onChange={(e) => onPatch({ relativeRange: e.target.value })}>
                  {relativeRanges.map((r) => <option key={r.value} value={r.value}>{r.label}</option>)}
                </select>
                <label className='fx-query-input-group'>
                  <span className='fx-query-input-group__addon'>Max data points</span>
                  <input type='number' min='50' max='2000' style={{ width: 70 }} value={panel.maxDataPoints}
                    onChange={(e) => onPatch({ maxDataPoints: Number(e.target.value) || 300 })} />
                </label>
                <label className='fx-query-input-group'>
                  <span className='fx-query-input-group__addon'>Min step</span>
                  <input type='number' min='0' max='3600' style={{ width: 60 }} value={panel.minStep}
                    onChange={(e) => onPatch({ minStep: Number(e.target.value) || 0 })} />
                </label>
                {/* GAP-M11: Line/Area 图标按钮组 */}
                <div className='fx-query-chart-type-group'>
                  <button type='button' className={panel.graphMode === 'line' ? 'is-active' : ''} onClick={() => onPatch({ graphMode: 'line' })} title='Line'>
                    <svg width='16' height='16' viewBox='0 0 16 16'><polyline points='1,12 5,6 9,9 15,3' fill='none' stroke='currentColor' strokeWidth='1.5'/></svg>
                  </button>
                  <button type='button' className={panel.graphMode === 'area' ? 'is-active' : ''} onClick={() => onPatch({ graphMode: 'area' })} title='StackArea'>
                    <svg width='16' height='16' viewBox='0 0 16 16'><path d='M1,12 L5,6 L9,9 L15,3 V12 H1Z' fill='currentColor' opacity='0.3' stroke='currentColor' strokeWidth='1'/></svg>
                  </button>
                </div>
                {panel.rangeResult && <button type='button' className='fx-query-export' onClick={onExportRange}>CSV</button>}
              </div>
              <MetricGraph result={panel.rangeResult} mode={panel.graphMode} showLegend={panel.showLegend} />
              {series.length > 0 && (
                <div className='fx-query-series'>{series.map((item) => <span key={item.name}>{item.name}: {item.points} pts</span>)}</div>
              )}
            </div>
          )}
        </div>
      </div>
    </section>
  )
}
