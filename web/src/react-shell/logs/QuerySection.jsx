import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { formatLogError, LOG_BLOCKERS, LOG_SOURCES, logsApi } from '../api/logs.js'
import { Blocked, Field } from './LogsShared.jsx'
import {
  LogsQueryBuilder,
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

const PAGE_SIZE = 50

export function QuerySection() {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [arrivalLoading, setArrivalLoading] = useState(false)
  const [query, setQuery] = useState('')
  const [source, setSource] = useState('findx_audit')
  const [panel, setPanel] = useState('list')     // list | chart | table
  const [format, setFormat] = useState('default') // raw | default | column
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

  const buildParams = useCallback((pageNum) => ({
    source,
    query,
    view: panel,
    limit: PAGE_SIZE,
    offset: (pageNum - 1) * PAGE_SIZE,
    time_range: timeRange,
    severity: severity || undefined,
    service: serviceFilter || undefined,
    host: hostFilter || undefined,
  }), [source, query, panel, timeRange, severity, serviceFilter, hostFilter])

  const runQuery = async () => {
    setLoading(true)
    setError('')
    setActiveRow(null)
    setPage(1)
    if (source !== 'findx_audit') {
      setRows([])
      setMeta(null)
      setError(LOG_BLOCKERS.query)
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
      if (!logsRow) setError('BLOCKED_BY_CONTRACT: FindX Agent 数据到达契约未返回日志通道。')
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
      <LogsQueryBuilder
        query={query}
        onQueryChange={setQuery}
        onSubmit={runQuery}
        loading={loading}
        timeRange={timeRange}
        onTimeRangeChange={setTimeRange}
        extraRight={
          <>
            <Field label='来源'>
              <select value={source} onChange={e => setSource(e.target.value)} className='fx-qb__select'>
                {LOG_SOURCES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
              </select>
            </Field>
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

      {source !== 'findx_audit' && <Blocked>{LOG_BLOCKERS.query}</Blocked>}
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

function findLogsArrival(rows) {
  return Array.isArray(rows) ? rows.find(r => r?.kind === 'logs') || null : null
}

function firstRowScope(rows) {
  const row = Array.isArray(rows) ? rows[0] : null
  return row?.service_name || row?.source_name || row?.trace_id || row?.scope || ''
}

// re-exports preserved for legacy imports
export { SEVERITY_META, normalizeLevel }
