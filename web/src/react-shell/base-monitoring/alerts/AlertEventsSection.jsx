import React, { useEffect, useMemo, useState } from 'react'
import { alertsApi } from '../../api/alerts.js'
import {
  displayDate,
  eventStatuses,
  filterText,
  formatDuration,
  getUniqueOptions,
  makeError,
  mapToPairs,
  normalizeEvent,
  severities,
  severityLabel,
  statusLabel,
} from './alertModel.js'
import { EventDrawer } from './form/EventDrawer.jsx'
import { AggregatedView } from './form/AggregatedView.jsx'
import { useConfirm } from '../../shared/ConfirmModal.jsx'

const rangeOptions = [
  { value: '', label: '全部时间' },
  { value: '1h', label: '最近 1 小时' },
  { value: '6h', label: '最近 6 小时' },
  { value: '24h', label: '最近 24 小时' },
  { value: '7d', label: '最近 7 天' },
]

const rangeToMs = {
  '1h': 60 * 60 * 1000,
  '6h': 6 * 60 * 60 * 1000,
  '24h': 24 * 60 * 60 * 1000,
  '7d': 7 * 24 * 60 * 60 * 1000,
}

const PAGE_SIZE = 20
const VIEW_MODES = [
  { value: 'list', label: '列表视图' },
  { value: 'rule', label: '规则聚合' },
  { value: 'target', label: '目标聚合' },
]

function buildCards(events) {
  const active = events.filter((e) => e.status === 'firing').length
  const acknowledged = events.filter((e) => ['acknowledged', 'assigned'].includes(e.status)).length
  const resolved = events.filter((e) => ['resolved', 'archived'].includes(e.status)).length
  const critical = events.filter((e) => ['critical', 'p0'].includes(e.severity)).length
  return [
    { key: 'active', label: '触发中', value: active },
    { key: 'ack', label: '处理中', value: acknowledged },
    { key: 'resolved', label: '已恢复/归档', value: resolved },
    { key: 'critical', label: '高优先级', value: critical },
  ]
}

function Pagination({ total, page, pageSize, onPageChange }) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  return (
    <div className='fx-alert-pagination'>
      <span>共 {total} 条，第 {page}/{totalPages} 页</span>
      <button type='button' disabled={page <= 1} onClick={() => onPageChange(page - 1)}>上一页</button>
      <button type='button' disabled={page >= totalPages} onClick={() => onPageChange(page + 1)}>下一页</button>
    </div>
  )
}

export function AlertEventsSection({ historyMode = false }) {
  const [events, setEvents] = useState([])
  const [history, setHistory] = useState(historyMode)
  const [keyword, setKeyword] = useState('')
  const [severity, setSeverity] = useState('')
  const { confirm, modal: confirmModal } = useConfirm()
  const [status, setStatus] = useState('')
  const [datasourceId, setDatasourceId] = useState('')
  const [timeRange, setTimeRange] = useState('')
  const [selectedIds, setSelectedIds] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [active, setActive] = useState(null)
  const [modal, setModal] = useState(null)
  const [page, setPage] = useState(1)
  const [viewMode, setViewMode] = useState('list')

  const loadEvents = async () => {
    setLoading(true); setError('')
    try {
      const params = { severity, status }
      const rows = history ? await alertsApi.listHistoryEvents(params) : await alertsApi.listCurrentEvents(params)
      setEvents(rows.map(normalizeEvent).filter((row) => row.id))
      setSelectedIds([])
      setPage(1)
    } catch (err) {
      setError(makeError(err, '事件加载失败'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { setHistory(historyMode) }, [historyMode])
  useEffect(() => { loadEvents() }, [history, severity, status])

  const datasourceOptions = useMemo(() => getUniqueOptions(events, 'datasourceId'), [events])
  const filtered = useMemo(() => {
    const now = Date.now()
    return events.filter((event) => {
      if (datasourceId && event.datasourceId !== datasourceId) return false
      if (timeRange) {
        const lastSeen = new Date(event.lastSeen || event.firstSeen).getTime()
        if (Number.isNaN(lastSeen) || now - lastSeen > rangeToMs[timeRange]) return false
      }
      return filterText([event.name, event.target, event.datasourceId, event.businessGroup, event.category, event.fingerprint, ...mapToPairs(event.labels)], keyword)
    })
  }, [events, keyword, datasourceId, timeRange])

  const total = filtered.length
  const paged = useMemo(() => filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE), [filtered, page])
  const cards = useMemo(() => buildCards(filtered), [filtered])

  const openDetail = async (event) => {
    try {
      setActive(normalizeEvent(await alertsApi.getEvent(event.id)))
    } catch (err) {
      setModal({ title: '详情加载失败', body: makeError(err) })
    }
  }

  const eventAction = async (action, event, body = {}) => {
    try {
      if (action === 'ack') await alertsApi.ackEvent(event.id, body)
      else if (action === 'assign') await alertsApi.assignEvent(event.id, body)
      else if (action === 'resolve') await alertsApi.resolveEvent(event.id, body)
      else if (action === 'archive') await alertsApi.archiveEvent(event.id, body)
      else if (action === 'mute') await alertsApi.batchMuteEvents([event.id], body)
      else if (action === 'delete') await alertsApi.deleteEvent(event.id)
      setActive(null)
      await loadEvents()
    } catch (err) {
      setModal({ title: '处置失败', body: makeError(err) })
    }
  }

  const batchAck = async () => {
    if (!selectedIds.length) return
    try { await alertsApi.batchAckEvents(selectedIds, {}); await loadEvents() }
    catch (err) { setModal({ title: '批量确认失败', body: makeError(err) }) }
  }

  const batchMute = async () => {
    if (!selectedIds.length) return
    try { await alertsApi.batchMuteEvents(selectedIds, {}); await loadEvents() }
    catch (err) { setModal({ title: '批量屏蔽失败', body: makeError(err) }) }
  }

  const batchDelete = async () => {
    if (!selectedIds.length) return
    const ok = await confirm({ title: '批量删除事件', message: `确认删除 ${selectedIds.length} 个事件？`, confirmText: '删除', danger: true })
    if (!ok) return
    try { await alertsApi.batchDeleteEvents(selectedIds); await loadEvents() }
    catch (err) { setModal({ title: '批量删除失败', body: makeError(err) }) }
  }

  const toggleSelect = (id, checked) => setSelectedIds((ids) => checked ? [...ids, id] : ids.filter((i) => i !== id))
  const toggleSelectAll = (checked) => setSelectedIds(checked ? paged.map((e) => e.id) : [])

  return (
    <section className='fx-alert-section'>
      <div className='fx-alert-filterbar'>
        <button type='button' disabled={loading} onClick={loadEvents}>{loading ? '刷新中...' : '刷新'}</button>
        <div className='fx-alert-segment'>
          <button type='button' className={!history ? 'is-active' : ''} onClick={() => setHistory(false)}>当前事件</button>
          <button type='button' className={history ? 'is-active' : ''} onClick={() => setHistory(true)}>历史事件</button>
        </div>
        <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder='搜索事件、目标、业务组、标签' />
        <select value={timeRange} onChange={(e) => setTimeRange(e.target.value)}>{rangeOptions.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select>
        <select value={datasourceId} onChange={(e) => setDatasourceId(e.target.value)}>
          <option value=''>全部数据源</option>
          {datasourceOptions.map((item) => <option key={item} value={item}>{item}</option>)}
        </select>
        <select value={severity} onChange={(e) => setSeverity(e.target.value)}><option value=''>全部级别</option>{severities.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select>
        <select value={status} onChange={(e) => setStatus(e.target.value)}><option value=''>全部状态</option>{eventStatuses.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select>
        <div className='fx-alert-segment'>
          {VIEW_MODES.map((mode) => (
            <button key={mode.value} type='button' className={viewMode === mode.value ? 'is-active' : ''} onClick={() => setViewMode(mode.value)}>{mode.label}</button>
          ))}
        </div>
      </div>
      <div className='fx-alert-cards'>
        {cards.map((card) => <article key={card.key}><span>{card.label}</span><strong>{card.value}</strong></article>)}
      </div>
      {selectedIds.length > 0 && (
        <div className='fx-alert-batchbar'>
          <span>已选择 {selectedIds.length} 个事件</span>
          <button type='button' onClick={batchAck}>批量确认</button>
          <button type='button' onClick={batchMute}>批量屏蔽</button>
          <button type='button' className='is-danger' onClick={batchDelete}>批量删除</button>
        </div>
      )}
      {error && <div className='fx-alert-message is-error'>{error}</div>}
      {viewMode === 'list' && (
        <>
          <div className='fx-alert-table'>
            <table>
              <thead><tr><th><input type='checkbox' checked={selectedIds.length === paged.length && paged.length > 0} onChange={(e) => toggleSelectAll(e.target.checked)} /></th><th>事件</th><th>级别</th><th>状态</th><th>目标</th><th>业务组</th><th>数据源</th><th>次数</th><th>持续时间</th><th>最近触发</th><th>操作</th></tr></thead>
              <tbody>
                {paged.map((event) => (
                  <tr key={event.id}>
                    <td><input type='checkbox' checked={selectedIds.includes(event.id)} onChange={(e) => toggleSelect(event.id, e.target.checked)} /></td>
                    <td><button type='button' className='fx-alert-link' onClick={() => openDetail(event)}>{event.name}</button><small>{event.id}</small>{mapToPairs(event.labels).slice(0, 3).map((item) => <span className='fx-alert-tag' key={item}>{item}</span>)}</td>
                    <td><span className={`fx-alert-severity is-${event.severity}`}>{severityLabel(event.severity)}</span></td>
                    <td>{statusLabel(event.status)}</td>
                    <td>{event.target}</td>
                    <td>{event.businessGroup}</td>
                    <td>{event.datasourceId}</td>
                    <td>{event.count}</td>
                    <td>{formatDuration(event.firstSeen, event.lastSeen || Date.now())}</td>
                    <td>{displayDate(event.lastSeen)}</td>
                    <td><select onChange={(e) => { if (e.target.value) eventAction(e.target.value, event, {}); e.target.value = '' }}><option value=''>更多</option><option value='ack'>确认</option><option value='resolve'>恢复</option><option value='archive'>归档</option><option value='mute'>屏蔽</option><option value='delete'>删除</option></select></td>
                  </tr>
                ))}
              </tbody>
            </table>
            {!loading && paged.length === 0 && <div className='fx-alert-empty'>暂无事件</div>}
          </div>
          <Pagination total={total} page={page} pageSize={PAGE_SIZE} onPageChange={setPage} />
        </>
      )}
      {viewMode === 'rule' && <AggregatedView events={filtered} groupKey='ruleId' onOpenDetail={openDetail} />}
      {viewMode === 'target' && <AggregatedView events={filtered} groupKey='target' onOpenDetail={openDetail} />}
      {active && <EventDrawer event={active} onClose={() => setActive(null)} onAction={eventAction} />}
      {modal && <div className='fx-alert-modal'><div className='fx-alert-modal__body'><header><h2>{modal.title}</h2><button type='button' onClick={() => setModal(null)}>关闭</button></header><pre>{modal.body}</pre></div></div>}
      {confirmModal}
    </section>
  )
}

