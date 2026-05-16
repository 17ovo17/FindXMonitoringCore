import React, { useEffect, useMemo, useRef, useState } from 'react'
import { LOG_SOURCES } from '../api/logs.js'
import { Blocked, Field } from './LogsShared.jsx'
import {
  LogsQueryBuilder,
  SeverityChips,
  LogChartView,
  buildHistogram,
  normalizeLevel,
  formatTime,
} from './LogsViewKit.jsx'
import { useLiveFollow, LIVE_STATES, WS_STATES } from './useLiveFollow.js'

const INTERVALS = [
  { value: 3, label: '3 秒' },
  { value: 5, label: '5 秒' },
  { value: 10, label: '10 秒' },
  { value: 30, label: '30 秒' },
]

export function LiveSection() {
  const live = useLiveFollow()
  const histogram = useMemo(() => buildHistogram(live.rows, 24), [live.rows])
  return (
    <section className='fx-logs-work'>
      <LogsQueryBuilder
        query={live.query}
        onQueryChange={value => live.setControlValue({ query: value })}
        onSubmit={live.startFollow}
        loading={live.loading}
        hideTimeRange
        submitLabel='开始跟随'
        submittingLabel='连接中...'
        placeholder='输入关键词过滤实时日志（动作 / 资源 / Trace / 摘要）'
        extraRight={
          <>
            <Field label='来源'>
              <select value={live.source} onChange={event => live.changeSource(event.target.value)} className='fx-qb__select'>
                {LOG_SOURCES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
              </select>
            </Field>
            <IntervalSelect value={live.intervalSeconds} onChange={v => live.setControlValue({ intervalSeconds: v })} />
            <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={live.openAgent}>进入主机 Agent</button>
            <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={live.loadArrival} disabled={live.arrivalLoading}>
              {live.arrivalLoading ? '读取中' : '数据到达'}
            </button>
          </>
        }
      />

      <SeverityChips
        value={live.severityFilter}
        onChange={value => live.setControlValue({ severityFilter: value })}
        serviceFilter={live.serviceFilter}
        onServiceChange={value => live.setControlValue({ serviceFilter: value })}
        hostFilter={live.hostFilter}
        onHostChange={value => live.setControlValue({ hostFilter: value })}
        onClearAll={live.severityFilter || live.serviceFilter || live.hostFilter
          ? () => live.setControlValue({ severityFilter: '', serviceFilter: '', hostFilter: '' })
          : null}
      />

      <LiveToolbar live={live} />

      {live.source === 'findx_audit' && live.blocked && <Blocked>{live.blocked}</Blocked>}
      {live.wsNotice && (
        <div className='fx-logs-blocked' style={{ borderColor: '#ffd591', background: '#fffbe6', color: '#ad6800' }}>
          {live.wsNotice}
        </div>
      )}

      <LogChartView
        histogram={histogram}
        title={`实时日志速率（近 ${live.rows.length} 条 / 每桶数量）`}
        height={160}
      />
      <LiveStreamView live={live} />
    </section>
  )
}

function IntervalSelect({ value, onChange }) {
  return (
    <Field label='刷新间隔'>
      <select value={value} onChange={event => onChange(Number(event.target.value))} className='fx-qb__select'>
        {INTERVALS.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
      </select>
    </Field>
  )
}

function LiveToolbar({ live }) {
  const playing = live.state === LIVE_STATES.playing
  return (
    <div className='fx-qb' style={{ gap: 8 }}>
      {!playing ? (
        <button type='button' className='fx-qb__btn' onClick={live.startFollow} disabled={live.loading}>
          {live.loading ? '连接中...' : '开始跟随'}
        </button>
      ) : (
        <button type='button' className='fx-qb__btn fx-qb__btn--warn' onClick={live.pauseFollow}>暂停</button>
      )}
      <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={live.stopFollow} disabled={live.state === LIVE_STATES.stopped && !live.rows.length}>停止</button>
      <button type='button' className='fx-qb__btn fx-qb__btn--danger' onClick={live.clearLogs} disabled={!live.rows.length}>清空</button>
      <span className='fx-qb__divider' />
      <LiveIndicator state={live.state} />
      <TransportBadge transport={live.transport} />
      {live.updatedAt && <span style={{ fontSize: 12, color: 'var(--fx-text-weak,#66758d)' }}>最近刷新：{live.updatedAt}</span>}
      <span style={{ fontSize: 12, color: 'var(--fx-text-weak,#66758d)' }}>已接收 {live.rows.length} 条 · 轮询 {live.pollCount} 次</span>
    </div>
  )
}

function TransportBadge({ transport }) {
  const isWs = transport === WS_STATES.connected
  const label = isWs ? 'WebSocket' : transport === WS_STATES.connecting ? '连接中...' : '轮询'
  const style = isWs ? { background: '#e8f8ef', color: '#167346' } : { background: '#f5f5f5', color: '#595959' }
  return <span style={{ ...style, fontSize: 11, fontWeight: 700, padding: '2px 8px', borderRadius: 10 }}>{label}</span>
}

function LiveIndicator({ state }) {
  const cls = state === LIVE_STATES.playing ? 'is-playing' : state === LIVE_STATES.paused ? 'is-paused' : 'is-stopped'
  const text = state === LIVE_STATES.playing ? '实时跟随中' : state === LIVE_STATES.paused ? '已暂停' : '未启动'
  return (
    <span className={`fx-live-indicator ${cls}`}>
      <span className='fx-live-indicator__dot' />
      {text}
    </span>
  )
}

function LiveStreamView({ live }) {
  const containerRef = useRef(null)
  const [autoScroll, setAutoScroll] = useState(true)

  useEffect(() => {
    if (autoScroll && containerRef.current) containerRef.current.scrollTop = containerRef.current.scrollHeight
  }, [live.rows.length, autoScroll])

  const handleScroll = () => {
    if (!containerRef.current) return
    const { scrollTop, scrollHeight, clientHeight } = containerRef.current
    setAutoScroll(scrollHeight - scrollTop - clientHeight < 40)
  }

  if (live.source !== 'findx_audit') {
    return (
      <div className='fx-logs-panel'>
        <p style={{ margin: 0, color: 'var(--fx-text-weak,#66758d)' }}>通用 OTel 实时日志仍为 PENDING，未发起实时请求。</p>
      </div>
    )
  }

  return (
    <div className='fx-live'>
      {!autoScroll && live.state === LIVE_STATES.playing && (
        <button type='button' className='fx-live__notice' onClick={() => setAutoScroll(true)}>有新日志到达，点击回到底部</button>
      )}
      <div ref={containerRef} onScroll={handleScroll} className='fx-live__scroll'>
        {live.rows.length === 0 ? (
          <div className='fx-live__empty'>
            <big>{'>_'}</big>
            <p style={{ margin: 0 }}>{live.state === LIVE_STATES.playing ? '等待日志到达...' : '点击「开始跟随」启动实时日志流'}</p>
          </div>
        ) : (
          live.rows.map((row, idx) => <LiveLogLine key={row.id || idx} row={row} />)
        )}
      </div>
      <div className='fx-live__foot'>
        <span>模式：{live.transport === WS_STATES.connected ? 'WebSocket' : `轮询（间隔 ${live.intervalSeconds}s）`}</span>
        <span>{live.rows.length} 条日志 · {autoScroll ? '自动滚动中' : '手动浏览'}</span>
      </div>
    </div>
  )
}

function LiveLogLine({ row }) {
  const level = normalizeLevel(row.severity_text || row.level)
  const svc = row.source_name || row.service_name || row.source || '-'
  return (
    <div className={`fx-live-line is-${level}`}>
      <span className='fx-live-line__ts'>{formatTime(row.timestamp)}</span>
      <span className={`fx-live-line__lvl is-${level}`}>{level}</span>
      <span className='fx-live-line__svc'>{svc}</span>
      <span className='fx-live-line__body'>{row.body || row.message || '-'}</span>
    </div>
  )
}
