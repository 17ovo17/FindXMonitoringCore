import React from 'react'
import { SEVERITY_META, normalizeLevel, formatTime } from './LogsViewKit.jsx'

/**
 * Chart view: simple SVG bar chart of log counts per bucket.
 */
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

/**
 * Table view.
 */
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
