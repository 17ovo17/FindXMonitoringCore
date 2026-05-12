import React, { useMemo } from 'react'

// Severity metadata -- used by chips, row badges, line borders, chart bars.
export const SEVERITY_META = {
  '':      { key: '',      label: '全部级别', color: '#66758d' },
  debug:   { key: 'debug', label: 'DEBUG',   color: '#8c8c8c' },
  info:    { key: 'info',  label: 'INFO',    color: '#1769ff' },
  warn:    { key: 'warn',  label: 'WARN',    color: '#d4a017' },
  warning: { key: 'warn',  label: 'WARN',    color: '#d4a017' },
  error:   { key: 'error', label: 'ERROR',   color: '#cf1322' },
  fatal:   { key: 'error', label: 'ERROR',   color: '#cf1322' },
}

export const SEVERITY_ORDER = ['', 'debug', 'info', 'warn', 'error']

export const TIME_RANGES = [
  { value: '15m', label: '最近 15 分钟' },
  { value: '1h',  label: '最近 1 小时' },
  { value: '6h',  label: '最近 6 小时' },
  { value: '24h', label: '最近 24 小时' },
  { value: '7d',  label: '最近 7 天' },
  { value: '30d', label: '最近 30 天' },
]

export function normalizeLevel(value) {
  const raw = String(value || '').toLowerCase()
  const meta = SEVERITY_META[raw]
  return meta ? meta.key : 'info'
}

export function formatTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString()
}

export function formatDetailValue(value) {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'object') {
    try { return JSON.stringify(value) } catch (_) { return String(value) }
  }
  return String(value)
}

// Bucket log rows into `bucketCount` time buckets for the histogram.
export function buildHistogram(rows, bucketCount = 20) {
  const valid = rows
    .map(r => ({ ts: Date.parse(r.timestamp || r.time || r.ts || 0), level: normalizeLevel(r.severity_text || r.level) }))
    .filter(r => !Number.isNaN(r.ts))
  if (!valid.length) return { buckets: [], start: 0, end: 0 }
  let start = Infinity
  let end = -Infinity
  for (const r of valid) { if (r.ts < start) start = r.ts; if (r.ts > end) end = r.ts }
  if (start === end) end = start + 60_000
  const span = end - start
  const buckets = Array.from({ length: bucketCount }, () => ({ count: 0, error: 0, warn: 0 }))
  for (const r of valid) {
    const idx = Math.min(bucketCount - 1, Math.floor(((r.ts - start) / span) * bucketCount))
    buckets[idx].count += 1
    if (r.level === 'error') buckets[idx].error += 1
    if (r.level === 'warn')  buckets[idx].warn += 1
  }
  return { buckets, start, end }
}

// ---------------------------------------------------------------------------
// Query builder bar: free-text search + time range + submit + extra buttons.
// ---------------------------------------------------------------------------
export function LogsQueryBuilder({
  query,
  onQueryChange,
  onSubmit,
  loading,
  timeRange,
  onTimeRangeChange,
  hideTimeRange = false,
  submitLabel = '检索',
  submittingLabel = '检索中...',
  extraRight,
  placeholder = '输入关键词检索日志：动作、资源、Trace ID、摘要内容...',
}) {
  const handleKey = (event) => {
    if (event.key === 'Enter') onSubmit()
  }
  return (
    <div className='fx-qb'>
      <input
        className='fx-qb__search'
        value={query}
        onChange={event => onQueryChange(event.target.value)}
        onKeyDown={handleKey}
        placeholder={placeholder}
      />
      {!hideTimeRange && (
        <select className='fx-qb__select' value={timeRange} onChange={event => onTimeRangeChange(event.target.value)}>
          {TIME_RANGES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
        </select>
      )}
      <button type='button' className='fx-qb__btn' onClick={onSubmit} disabled={loading}>
        {loading ? submittingLabel : submitLabel}
      </button>
      <span className='fx-qb__divider' />
      {extraRight}
    </div>
  )
}

// Severity chips + service / host filters.
export function SeverityChips({
  value,
  onChange,
  serviceFilter,
  onServiceChange,
  hostFilter,
  onHostChange,
  onClearAll,
}) {
  return (
    <div className='fx-qb-chips'>
      <span className='fx-qb-chips__label'>级别：</span>
      {SEVERITY_ORDER.map(key => {
        const meta = SEVERITY_META[key]
        const active = value === key
        return (
          <button
            key={key || 'all'}
            type='button'
            className={'fx-chip' + (active ? ' is-active' : '')}
            style={active ? { color: meta.color, borderColor: meta.color } : null}
            onClick={() => onChange(active ? '' : key)}
          >
            {key && <span className='fx-chip__dot' style={{ background: meta.color }} />}
            {meta.label}
          </button>
        )
      })}
      <span className='fx-qb__divider' />
      <input
        className='fx-chip'
        style={{ width: 130, cursor: 'text' }}
        value={serviceFilter}
        onChange={event => onServiceChange(event.target.value)}
        placeholder='服务名'
      />
      <input
        className='fx-chip'
        style={{ width: 130, cursor: 'text' }}
        value={hostFilter}
        onChange={event => onHostChange(event.target.value)}
        placeholder='主机名'
      />
      {onClearAll && (
        <button type='button' className='fx-chip' style={{ color: '#9f3a38' }} onClick={onClearAll}>
          清除过滤
        </button>
      )}
    </div>
  )
}

// ---------------------------------------------------------------------------
// View toolbar: List / Chart / Table tabs + Raw / Default / Column format.
// ---------------------------------------------------------------------------
const PANELS = [
  { value: 'list',  label: 'List view' },
  { value: 'chart', label: 'Chart view' },
  { value: 'table', label: 'Table view' },
]

const FORMATS = [
  { value: 'raw',     label: 'Raw' },
  { value: 'default', label: 'Default' },
  { value: 'column',  label: 'Column' },
]

export function ViewToolbar({
  panel,
  onPanelChange,
  format,
  onFormatChange,
  showFormat = true,
  meta = [],
}) {
  return (
    <div className='fx-view-bar'>
      <div className='fx-view-tabs'>
        {PANELS.map(item => (
          <button
            key={item.value}
            type='button'
            className={'fx-view-tabs__btn' + (panel === item.value ? ' is-active' : '')}
            onClick={() => onPanelChange(item.value)}
          >
            {item.label}
          </button>
        ))}
      </div>
      <div className='fx-view-meta'>
        {meta.map((item, idx) => (
          <React.Fragment key={idx}>
            {idx > 0 && <span className='fx-view-meta__sep' />}
            <span>{item}</span>
          </React.Fragment>
        ))}
        {showFormat && (
          <>
            {meta.length > 0 && <span className='fx-view-meta__sep' />}
            <span style={{ marginRight: 6 }}>Format</span>
            <div className='fx-segment'>
              {FORMATS.map(item => (
                <button
                  key={item.value}
                  type='button'
                  className={'fx-segment__btn' + (format === item.value ? ' is-active' : '')}
                  onClick={() => onFormatChange(item.value)}
                >
                  {item.label}
                </button>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// List body: renders rows in default / raw / column layout.
// ---------------------------------------------------------------------------
export function LogListBody({ rows, format = 'default', onRowClick, activeRow, hasMore, onLoadMore, loadingMore }) {
  return (
    <div className='fx-loglist'>
      <div className='fx-loglist__body'>
        {rows.length === 0
          ? <NoResults />
          : rows.map((row, idx) => (
              <LogRowLine
                key={row.id || idx}
                row={row}
                format={format}
                active={row === activeRow}
                onClick={() => onRowClick && onRowClick(row)}
              />
            ))
        }
      </div>
      {hasMore && (
        <div className='fx-loglist__loadmore'>
          <button type='button' onClick={onLoadMore} disabled={loadingMore}>
            {loadingMore ? '加载中...' : '加载更多'}
          </button>
        </div>
      )}
      {rows.length > 0 && (
        <div className='fx-loglist__foot'>
          <span>{rows.length} 行已加载</span>
          <span>单击行展开详情</span>
        </div>
      )}
    </div>
  )
}

function LogRowLine({ row, format, active, onClick }) {
  const levelKey = normalizeLevel(row.severity_text || row.level)
  const ts = formatTime(row.timestamp)
  const svc = row.source_name || row.service_name || row.source || '-'
  const body = row.body || row.message || '-'
  const trace = row.trace_id ? row.trace_id.slice(0, 12) + '...' : '-'

  if (format === 'raw') {
    const rawLine = `${ts}  ${levelKey.toUpperCase()}  ${svc}  ${body}`
    return (
      <div
        className={`fx-logrow fx-logrow--raw is-${levelKey}${active ? ' is-active' : ''}`}
        onClick={onClick}
      >
        {rawLine}
      </div>
    )
  }

  if (format === 'column') {
    return (
      <div
        className={`fx-logrow fx-logrow--column is-${levelKey}${active ? ' is-active' : ''}`}
        onClick={onClick}
      >
        <span className='fx-logrow__ts'>{ts}</span>
        <span className={`fx-logrow__lvl is-${levelKey}`}>{levelKey}</span>
        <span className='fx-logrow__svc'>{svc}</span>
        <span className='fx-logrow__trace'>{trace}</span>
        <span className='fx-logrow__body'>{body}</span>
      </div>
    )
  }

  return (
    <div
      className={`fx-logrow is-${levelKey}${active ? ' is-active' : ''}`}
      onClick={onClick}
    >
      <span className='fx-logrow__ts'>{ts}</span>
      <span className={`fx-logrow__lvl is-${levelKey}`}>{levelKey}</span>
      <span className='fx-logrow__svc'>{svc}</span>
      <span className='fx-logrow__body'>{body}</span>
      <span className='fx-logrow__trace'>{trace}</span>
    </div>
  )
}

function NoResults() {
  return (
    <div style={{ padding: '48px 16px', textAlign: 'center', color: 'var(--fx-text-weak,#66758d)' }}>
      <div style={{ fontSize: 28, marginBottom: 8 }}>{'{ }'}</div>
      <p style={{ margin: 0, color: 'var(--fx-heading,#17233c)', fontWeight: 700 }}>暂无匹配的日志</p>
      <p style={{ margin: '6px 0 0', fontSize: 13 }}>调整时间范围、级别或关键词后重试。</p>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Chart view: simple SVG bar chart of log counts per bucket.
// ---------------------------------------------------------------------------
export function LogChartView({ histogram, title = '日志数量分布（按时间桶）', height = 200 }) {
  const buckets = histogram?.buckets || []
  if (!buckets.length) {
    return (
      <div className='fx-chart'>
        <h4 className='fx-chart__title'>{title}</h4>
        <div className='fx-chart__empty'>暂无数据可绘制柱状图。</div>
      </div>
    )
  }
  const width = 960
  const pad = { top: 16, right: 16, bottom: 24, left: 28 }
  const innerW = width - pad.left - pad.right
  const innerH = height - pad.top - pad.bottom
  const maxCount = buckets.reduce((m, b) => Math.max(m, b.count), 1)
  const barW = innerW / buckets.length - 2
  const yTicks = 4
  const start = histogram.start ? new Date(histogram.start) : null
  const end = histogram.end ? new Date(histogram.end) : null
  return (
    <div className='fx-chart'>
      <h4 className='fx-chart__title'>{title}</h4>
      <svg className='fx-chart__svg' viewBox={`0 0 ${width} ${height}`} preserveAspectRatio='none'>
        {Array.from({ length: yTicks + 1 }, (_, i) => {
          const y = pad.top + (innerH * i) / yTicks
          const v = Math.round(maxCount * (1 - i / yTicks))
          return (
            <g key={i}>
              <line x1={pad.left} x2={width - pad.right} y1={y} y2={y} className='fx-chart__axis' />
              <text x={pad.left - 6} y={y + 3} className='fx-chart__label' textAnchor='end'>{v}</text>
            </g>
          )
        })}
        {buckets.map((bucket, idx) => {
          const h = (bucket.count / maxCount) * innerH
          const x = pad.left + (innerW * idx) / buckets.length + 1
          const y = pad.top + innerH - h
          const className = bucket.error > 0
            ? 'fx-chart__bar is-error'
            : bucket.warn > 0 ? 'fx-chart__bar is-warn' : 'fx-chart__bar'
          return (
            <rect key={idx} className={className} x={x} y={y} width={Math.max(barW, 1)} height={Math.max(h, 1)}>
              <title>{`${bucket.count} 条 (error=${bucket.error}, warn=${bucket.warn})`}</title>
            </rect>
          )
        })}
        {start && (
          <text x={pad.left} y={height - 6} className='fx-chart__label'>{start.toLocaleTimeString()}</text>
        )}
        {end && (
          <text x={width - pad.right} y={height - 6} className='fx-chart__label' textAnchor='end'>{end.toLocaleTimeString()}</text>
        )}
      </svg>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Table view.
// ---------------------------------------------------------------------------
export function LogTableView({ rows, onRowClick, activeRow }) {
  if (!rows.length) {
    return (
      <div className='fx-logtable__wrap'>
        <div style={{ padding: '32px 16px', textAlign: 'center', color: 'var(--fx-text-weak,#66758d)' }}>暂无匹配数据</div>
      </div>
    )
  }
  return (
    <div className='fx-logtable__wrap'>
      <table className='fx-logtable'>
        <thead>
          <tr>
            <th style={{ width: 170 }}>时间</th>
            <th style={{ width: 64 }}>级别</th>
            <th style={{ width: 140 }}>服务</th>
            <th style={{ width: 140 }}>主机</th>
            <th>内容</th>
            <th style={{ width: 140 }}>Trace</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row, idx) => {
            const levelKey = normalizeLevel(row.severity_text || row.level)
            const active = row === activeRow
            return (
              <tr key={row.id || idx} onClick={() => onRowClick && onRowClick(row)} style={active ? { background: 'var(--fx-primary-weak, rgba(23,105,255,.08))' } : null}>
                <td>{formatTime(row.timestamp)}</td>
                <td style={{ color: SEVERITY_META[levelKey]?.color, fontWeight: 700 }}>{levelKey.toUpperCase()}</td>
                <td>{row.source_name || row.service_name || row.source || '-'}</td>
                <td>{row.host_name || row.host || '-'}</td>
                <td style={{ maxWidth: 420, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{row.body || row.message || '-'}</td>
                <td>{row.trace_id ? row.trace_id.slice(0, 16) + '...' : '-'}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Log detail drawer: slides from right, 40% width.
// ---------------------------------------------------------------------------
export function LogDetailDrawer({ row, onClose, similarRows = [], onSelectSimilar }) {
  const entries = useMemo(() => flattenRow(row), [row])
  const levelKey = normalizeLevel(row.severity_text || row.level)
  const ts = formatTime(row.timestamp)
  const svc = row.source_name || row.service_name || row.source || '-'
  const host = row.host_name || row.host || '-'
  const trace = row.trace_id || ''
  const span = row.span_id || ''

  const openTrace = () => {
    if (!trace) return
    window.location.href = `/tracing?traceId=${encodeURIComponent(trace)}${span ? `&spanId=${encodeURIComponent(span)}` : ''}`
  }

  const copyJson = () => {
    try {
      const text = JSON.stringify(row, null, 2)
      if (navigator?.clipboard) navigator.clipboard.writeText(text)
    } catch (_) { /* noop */ }
  }

  const handleMaskClick = (event) => {
    if (event.target === event.currentTarget) onClose()
  }

  return (
    <>
      <div className='fx-drawer-mask' onClick={handleMaskClick} />
      <aside className='fx-drawer' role='dialog' aria-label='日志详情'>
        <div className='fx-drawer__head'>
          <h3 className='fx-drawer__title'>日志详情</h3>
          <button type='button' className='fx-drawer__close' onClick={onClose} aria-label='关闭详情'>×</button>
        </div>
        <div className='fx-drawer__meta'>
          <span><b>时间</b>{ts}</span>
          <span style={{ color: SEVERITY_META[levelKey]?.color, fontWeight: 700 }}><b>级别</b>{levelKey.toUpperCase()}</span>
          <span><b>服务</b>{svc}</span>
          <span><b>主机</b>{host}</span>
          {trace && <span><b>Trace</b>{trace.slice(0, 16)}...</span>}
        </div>
        <div className='fx-drawer__body'>
          <div className='fx-drawer__section'>
            <h4>消息体</h4>
            <pre className='fx-msg'>{row.body || row.message || '-'}</pre>
          </div>
          <div className='fx-drawer__section'>
            <h4>属性（{entries.length} 项）</h4>
            <dl className='fx-kv'>
              {entries.map(([key, value]) => (
                <React.Fragment key={key}>
                  <dt>{key}</dt>
                  <dd>{value}</dd>
                </React.Fragment>
              ))}
            </dl>
          </div>
          {similarRows.length > 0 && (
            <div className='fx-drawer__section'>
              <h4>相似日志（{similarRows.length}）</h4>
              <div className='fx-loglist' style={{ borderRadius: 6 }}>
                <div className='fx-loglist__body' style={{ maxHeight: 220 }}>
                  {similarRows.map((r, idx) => {
                    const lk = normalizeLevel(r.severity_text || r.level)
                    return (
                      <div key={r.id || idx} className={`fx-logrow is-${lk}`} onClick={() => onSelectSimilar && onSelectSimilar(r)}>
                        <span className='fx-logrow__ts'>{formatTime(r.timestamp)}</span>
                        <span className={`fx-logrow__lvl is-${lk}`}>{lk}</span>
                        <span className='fx-logrow__svc'>{r.source_name || r.service_name || '-'}</span>
                        <span className='fx-logrow__body'>{r.body || r.message || '-'}</span>
                        <span className='fx-logrow__trace'>{r.trace_id ? r.trace_id.slice(0,12)+'...' : '-'}</span>
                      </div>
                    )
                  })}
                </div>
              </div>
            </div>
          )}
        </div>
        <div className='fx-drawer__foot'>
          <button type='button' onClick={copyJson}>复制 JSON</button>
          <button type='button' disabled={!trace} onClick={openTrace} className={trace ? 'is-primary' : ''}>
            {trace ? '查看链路详情' : '无 Trace ID'}
          </button>
        </div>
      </aside>
    </>
  )
}

// Flatten an arbitrary object into key=value entries up to depth 2.
function flattenRow(row) {
  const out = []
  for (const [key, value] of Object.entries(row || {})) {
    if (key === '__proto__') continue
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      for (const [k2, v2] of Object.entries(value)) {
        out.push([`${key}.${k2}`, formatDetailValue(v2)])
      }
    } else {
      out.push([key, formatDetailValue(value)])
    }
  }
  return out
}
