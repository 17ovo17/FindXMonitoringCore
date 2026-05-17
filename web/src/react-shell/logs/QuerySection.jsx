import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { formatLogError, LOG_BLOCKERS, LOG_SOURCES, logsApi } from '../api/logs.js'
import { Blocked, Field } from './LogsShared.jsx'
import {
  SeverityChips,
  ViewToolbar,
  LogListBody,
  LogTableView,
  LogChartView,
  LogDetailDrawer,
  SEVERITY_META,
  buildHistogram,
  normalizeLevel,
} from './LogsViewKit.jsx'
import { QueryBuilder } from './QueryBuilder.jsx'

const PAGE_SIZE = 50

export function QuerySection() {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [arrivalLoading, setArrivalLoading] = useState(false)
  const [query, setQuery] = useState('')
  const [source, setSource] = useState('findx_audit')
  const [panel, setPanel] = useState('list')
  const [format, setFormat] = useState('default')
  const [timeRange, setTimeRange] = useState('24h')
  const [severity, setSeverity] = useState('')
  const [serviceFilter, setServiceFilter] = useState('')
  const [hostFilter, setHostFilter] = useState('')
  const [rows, setRows] = useState([])
  const [meta, setMeta] = useState(null)
  const [arrival, setArrival] = useState(null)
  const [activeRow, setActiveRow] = useState(null)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(false)
  const [conditions, setConditions] = useState([])
  const [queryMode, setQueryMode] = useState('builder')

  const buildParams = useCallback((pageNum) => ({
    source,
    query: queryMode === 'raw' ? query : buildQueryFromConditions(conditions, query),
    view: panel,
    limit: PAGE_SIZE,
    offset: (pageNum - 1) * PAGE_SIZE,
    time_range: timeRange,
    severity: severity || undefined,
    service: serviceFilter || undefined,
    host: hostFilter || undefined,
  }), [source, query, panel, timeRange, severity, serviceFilter, hostFilter, conditions, queryMode])

  const runQuery = async () => {
    setLoading(true)
    setError('')
    setActiveRow(null)
    setPage(1)
    if (source !== 'findx_audit') {
      setRows([])
      setMeta(null)
      setError(`需要部署 ${sourceLabel(source)} 并配置 /logs/query 代理`)
      setLoading(false)
      return
    }
    try {
      const resp = await logsApi.query(buildParams(1))
      const items = Array.isArray(resp?.items) ? resp.items : []
      setRows(items)
      setMeta(resp || null)
      setHasMore(items.length >= PAGE_SIZE)
    } catch (err) {
      setRows([])
      setMeta(null)
      setError(formatLogError(err))
      setHasMore(false)
    } finally {
      setLoading(false)
    }
  }

  const loadMore = async () => {
    if (loadingMore || !hasMore) return
    setLoadingMore(true)
    const nextPage = page + 1
    try {
      const resp = await logsApi.query(buildParams(nextPage))
      const items = Array.isArray(resp?.items) ? resp.items : []
      setRows(prev => [...prev, ...items])
      setMeta(resp || null)
      setPage(nextPage)
      setHasMore(items.length >= PAGE_SIZE)
    } catch (err) {
      setError(formatLogError(err))
    } finally {
      setLoadingMore(false)
    }
  }

  const openAgent = () => {
    const scope = query || firstRowScope(rows)
    window.location.href = `/agents?section=hosts&package=logs&q=${encodeURIComponent(scope || '')}`
  }

  const loadArrival = async () => {
    setError('')
    setArrivalLoading(true)
    try {
      const rowsValue = await agentApi.dataArrival()
      const logsRow = findLogsArrival(rowsValue)
      setArrival(logsRow)
      if (!logsRow) setError('FindX Agent 数据到达契约未返回日志通道。')
      if (logsRow && logsRow.status !== 'reported') setError(logsRow.blocker || LOG_BLOCKERS.agentLinkage)
    } catch (err) {
      setArrival(null)
      setError(formatAgentError(err))
    } finally {
      setArrivalLoading(false)
    }
  }

  const clearFilters = () => {
    setSeverity('')
    setServiceFilter('')
    setHostFilter('')
  }

  useEffect(() => { runQuery() }, [])

  const histogram = useMemo(() => buildHistogram(rows, 20), [rows])
  const activeFilters = [severity, serviceFilter, hostFilter].filter(Boolean)
  const similarRows = useMemo(() => {
    if (!activeRow) return []
    const sig = (activeRow.body || activeRow.message || '').slice(0, 40)
    if (!sig) return []
    return rows.filter(r => r !== activeRow && ((r.body || r.message || '').indexOf(sig) === 0)).slice(0, 5)
  }, [activeRow, rows])

  return (
    <section className='fx-logs-work'>
      {/* DEGRADE-047: 数据源选择器 */}
      <DataSourceSelector source={source} onChange={setSource} />

      {/* DEGRADE-048: 查询构建器 / 原始查询切换 */}
      <QueryBuilder
        query={query}
        onQueryChange={setQuery}
        conditions={conditions}
        onConditionsChange={setConditions}
        mode={queryMode}
        onModeChange={setQueryMode}
        onSubmit={runQuery}
        loading={loading}
        timeRange={timeRange}
        onTimeRangeChange={setTimeRange}
        extraRight={
          <>
            <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={openAgent}>进入主机 Agent</button>
            <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={loadArrival} disabled={arrivalLoading}>{arrivalLoading ? '读取中' : '数据到达'}</button>
          </>
        }
      />

      <SeverityChips
        value={severity}
        onChange={setSeverity}
        serviceFilter={serviceFilter}
        onServiceChange={setServiceFilter}
        hostFilter={hostFilter}
        onHostChange={setHostFilter}
        onClearAll={activeFilters.length ? clearFilters : null}
      />

      {source !== 'findx_audit' && <Blocked>{`需要部署 ${sourceLabel(source)} 并配置 /logs/query 代理`}</Blocked>}
      {error && <Blocked>{error}</Blocked>}

      <ViewToolbar
        panel={panel}
        onPanelChange={setPanel}
        format={format}
        onFormatChange={setFormat}
        showFormat={panel === 'list'}
        meta={[
          meta?.total != null ? `共 ${meta.total} 条` : null,
          loading ? '检索中...' : null,
        ].filter(Boolean)}
      />

      {panel === 'list' && (
        <LogListBody
          rows={rows}
          format={format}
          onRowClick={setActiveRow}
          activeRow={activeRow}
          hasMore={hasMore}
          onLoadMore={loadMore}
          loadingMore={loadingMore}
        />
      )}
      {panel === 'chart' && <LogChartView histogram={histogram} />}
      {panel === 'table' && <LogTableView rows={rows} onRowClick={setActiveRow} activeRow={activeRow} />}

      {activeRow && (
        <LogDetailDrawer
          row={activeRow}
          onClose={() => setActiveRow(null)}
          similarRows={similarRows}
          onSelectSimilar={setActiveRow}
        />
      )}

      {arrival && (
        <div className='fx-logs-panel' style={{ marginTop: 12 }}>
          <h3>FindX Agent 数据到达</h3>
          <p>通道 {arrival.kind} 状态：{arrival.status}。{arrival.note || ''}</p>
        </div>
      )}
    </section>
  )
}

function DataSourceSelector({ source, onChange }) {
  return (
    <div className='fx-qb' style={{ marginBottom: 8 }}>
      <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--fx-log-heading,#17233c)' }}>数据源：</span>
      {LOG_SOURCES.map(item => (
        <button
          key={item.value}
          type='button'
          className={'fx-chip' + (source === item.value ? ' is-active' : '')}
          onClick={() => onChange(item.value)}
        >
          {item.label}
        </button>
      ))}
    </div>
  )
}

function sourceLabel(value) {
  const item = LOG_SOURCES.find(s => s.value === value)
  return item ? item.label : value
}

function buildQueryFromConditions(conditions, freeText) {
  const parts = []
  if (freeText) parts.push(freeText)
  if (!conditions.length) return parts.join(' ')
  const grouped = conditions.reduce((acc, c, idx) => {
    const expr = `${c.field} ${c.operator} "${c.value}"`
    if (idx === 0) return expr
    return `${acc} ${c.logic || 'AND'} ${expr}`
  }, '')
  if (grouped) parts.push(grouped)
  return parts.join(' ')
}

function findLogsArrival(rows) {
  return Array.isArray(rows) ? rows.find(r => r?.kind === 'logs') || null : null
}

function firstRowScope(rows) {
  const row = Array.isArray(rows) ? rows[0] : null
  return row?.service_name || row?.source_name || row?.trace_id || row?.scope || ''
}

// re-exports preserved for legacy imports
export { SEVERITY_META, normalizeLevel }
