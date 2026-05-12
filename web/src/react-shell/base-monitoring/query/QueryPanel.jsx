import React from 'react'
import { MetricGraph } from './MetricGraph.jsx'
import {
  displayJson,
  displayText,
  metricDescription,
  metricName,
  metricUnit,
  rangeSeries,
  relativeRanges,
  resultRows,
  UNITS,
} from './metricQueryModel.js'

function HistoryPopover({ panel, items, onPick, onSearch }) {
  const search = String(panel.historySearch || '').toLowerCase()
  const rows = items.filter((item) => !search || item.toLowerCase().includes(search))
  return (
    <div className='fx-query-popover'>
      <input value={panel.historySearch || ''} onChange={(e) => onSearch(e.target.value)} placeholder='搜索历史记录' />
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
      <input value={panel.metricSearch || ''} onChange={(e) => { onSearch(e.target.value) }} placeholder='搜索内置指标' />
      <div className='fx-query-popover__list'>
        {panel.metricsLoading && <span>正在加载...</span>}
        {!panel.metricsLoading && panel.metrics?.map((metric) => (
          <button type='button' key={metric.id || metricName(metric)} onClick={() => onPick(metric)}>
            <strong>{metricName(metric)}</strong>
            <small>{metricDescription(metric) || metricUnit(metric) || '-'}</small>
          </button>
        ))}
        {!panel.metricsLoading && (!panel.metrics || panel.metrics.length === 0) && <span>暂无指标</span>}
      </div>
    </div>
  )
}
function GraphSettings({ panel, onPatch }) {
  return (
    <div className='fx-query-popover fx-query-graph-settings'>
      <label className='fx-query-setting-row'>
        <span>Max data points</span>
        <input type='number' min='50' max='2000' value={panel.maxDataPoints}
          onChange={(e) => onPatch({ maxDataPoints: Number(e.target.value) || 300 })} />
      </label>
      <label className='fx-query-setting-row'>
        <span>Min step (0=auto)</span>
        <input type='number' min='0' max='3600' value={panel.minStep}
          onChange={(e) => onPatch({ minStep: Number(e.target.value) || 0 })} />
      </label>
      <label className='fx-query-setting-row'>
        <span>显示图例</span>
        <input type='checkbox' checked={panel.showLegend}
          onChange={(e) => onPatch({ showLegend: e.target.checked })} />
      </label>
    </div>
  )
}

function AiInput({ panel, onPatch, onGenerate }) {
  return (
    <div className='fx-query-popover fx-query-ai-input'>
      <input value={panel.aiPrompt || ''} onChange={(e) => onPatch({ aiPrompt: e.target.value })}
        placeholder='用自然语言描述查询，如"CPU 使用率超过 80% 的主机"' />
      <button type='button' disabled={panel.aiLoading || !panel.aiPrompt?.trim()}
        onClick={onGenerate}>{panel.aiLoading ? '生成中...' : '生成 PromQL'}</button>
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

export function QueryPanel({
  panel, index, canRemove, history,
  onPatch, onQueryChange, onInsertMetric, onLoadMetrics,
  onRun, onRemove, onExportTable, onExportRange, onAiGenerate, inputRef,
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
            <textarea ref={inputRef} className='fx-query-code' value={panel.query}
              onChange={(e) => onQueryChange(e.target.value)}
              placeholder='up{job="prometheus"}' spellCheck={false} />
            {panel.autocomplete && panel.suggestions?.length > 0 && (
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
              <button type='button' onClick={() => { onPatch({ showMetrics: !panel.showMetrics }); onLoadMetrics(panel.metricSearch || '') }}>内置指标</button>
              <button type='button' onClick={() => onPatch({ showAiInput: !panel.showAiInput })}>AI</button>
              <label className='fx-query-autocomplete-label'>
                <input type='checkbox' checked={panel.autocomplete} onChange={(e) => onPatch({ autocomplete: e.target.checked })} />
                <span>自动补全</span>
              </label>
            </div>
            <div className='fx-query-panel-time'>
              <select value={panel.relativeRange} onChange={(e) => onPatch({ relativeRange: e.target.value })}>
                {relativeRanges.map((r) => <option key={r.value} value={r.value}>{r.label}</option>)}
              </select>
              <label className='fx-query-step-label'>
                <span>Step</span>
                <input type='number' min='1' max='3600' value={panel.step} onChange={(e) => onPatch({ step: Number(e.target.value) })} />
                <span>s</span>
              </label>
            </div>
          </div>

          {panel.showHistory && (
            <HistoryPopover panel={panel} items={history}
              onPick={(item) => onPatch({ query: item, showHistory: false })}
              onSearch={(v) => onPatch({ historySearch: v })} />
          )}
          {panel.showMetrics && (
            <MetricsAssistant panel={panel}
              onSearch={(v) => { onPatch({ metricSearch: v }); onLoadMetrics(v) }}
              onPick={(metric) => onInsertMetric(metric)} />
          )}
          {panel.showAiInput && <AiInput panel={panel} onPatch={onPatch} onGenerate={onAiGenerate} />}

          {panel.error && <div className='fx-query-alert is-error'>{panel.error}</div>}

          <div className='fx-query-tabs' role='tablist'>
            <button type='button' className={panel.activeTab === 'table' ? 'is-active' : ''} onClick={() => onPatch({ activeTab: 'table' })}>表格</button>
            <button type='button' className={panel.activeTab === 'graph' ? 'is-active' : ''} onClick={() => onPatch({ activeTab: 'graph' })}>图表</button>
          </div>

          {panel.activeTab === 'table' && (
            <div className='fx-query-result'>
              <div className='fx-query-table-controls'>
                <select value={panel.unit} onChange={(e) => onPatch({ unit: e.target.value })}>
                  {UNITS.map((u) => <option key={u.value} value={u.value}>{u.label}</option>)}
                </select>
                {rows.length > 0 && <button type='button' className='fx-query-export' onClick={onExportTable}>CSV 导出</button>}
              </div>
              <StatsLine result={panel.instantResult} />
              {rows.length > 1000 && <div className='fx-query-alert is-warning'>结果超过 1000 行，建议缩小查询范围。</div>}
              {rows.length > 0 && (
                <div className='fx-query-table-wrap'>
                  <table>
                    <thead><tr><th>Metric</th><th>Value</th><th>Time</th></tr></thead>
                    <tbody>{rows.slice(0, 1200).map((row) => <tr key={`${row.metric}:${row.time}`}><td>{row.metric}</td><td>{row.value}</td><td>{row.time}</td></tr>)}</tbody>
                  </table>
                </div>
              )}
              {!panel.instantResult && !panel.loading && <div className='fx-query-empty'>暂无查询结果</div>}
            </div>
          )}

          {panel.activeTab === 'graph' && (
            <div className='fx-query-result'>
              <div className='fx-query-graph-controls'>
                <span className='fx-query-range-text'>{rangeLabel}</span>
                <select value={panel.graphMode} onChange={(e) => onPatch({ graphMode: e.target.value })}>
                  <option value='line'>Line</option>
                  <option value='area'>StackArea</option>
                </select>
                <button type='button' className='fx-query-export' onClick={() => onPatch({ showGraphSettings: !panel.showGraphSettings })}>设置</button>
                {panel.rangeResult && <button type='button' className='fx-query-export' onClick={onExportRange}>CSV 导出</button>}
              </div>
              {panel.showGraphSettings && <GraphSettings panel={panel} onPatch={onPatch} />}
              <StatsLine result={panel.rangeResult} />
              <MetricGraph result={panel.rangeResult} mode={panel.graphMode} showLegend={panel.showLegend} />
              {series.length > 0 && <div className='fx-query-series'>{series.map((item) => <span key={item.name}>{item.name}: {item.points} points</span>)}</div>}
            </div>
          )}
        </>
      )}
    </section>
  )
}
