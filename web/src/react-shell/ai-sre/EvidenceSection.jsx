import React, { useEffect, useMemo, useState } from 'react'
import { AISRE_BLOCKERS, aiSreApi, formatAiSreError } from '../api/aiSre.js'
import { Blocked, Empty, ErrorBox, StatusPill } from './AiSreShared.jsx'
import { evidenceCategories } from './aiSreModel.js'

const typeIcons = {
  metric: '📊',
  metrics: '📊',
  alert: '🔔',
  alerts: '🔔',
  log: '📝',
  logs: '📝',
  trace: '🔗',
  traces: '🔗',
  agent: '🤖',
  config: '⚙️',
  inspection: '🔍',
  workflow: '⚡',
  knowledge: '📚',
  diagnosis: '🩺',
  cmdb: '🗄️',
}

const typeLabels = {
  metric: '指标', metrics: '指标', alert: '告警', alerts: '告警',
  log: '日志', logs: '日志', trace: '链路', traces: '链路',
  agent: 'Agent', config: '配置', inspection: '巡检',
  workflow: '工作流', knowledge: '知识库', diagnosis: '诊断', cmdb: 'CMDB',
}

const evidenceTypeOptions = [
  { value: '', label: '全部类型' },
  { value: 'metrics', label: '指标' },
  { value: 'alerts', label: '告警' },
  { value: 'logs', label: '日志' },
  { value: 'traces', label: '链路' },
  { value: 'agent', label: 'Agent' },
  { value: 'config', label: '配置' },
]

function TimelineItem({ item, isExpanded, onToggle }) {
  const icon = typeIcons[item.kind || item.category] || '📋'
  const label = typeLabels[item.kind || item.category] || item.kind || item.category || '未知'
  const time = item.updated_at || item.at || item.timestamp
  return (
    <div className={`fx-aisre-timeline-item ${isExpanded ? 'is-expanded' : ''}`}>
      <div className='fx-aisre-timeline-dot'><span>{icon}</span></div>
      <div className='fx-aisre-timeline-card' onClick={onToggle}>
        <div className='fx-aisre-timeline-card-head'>
          <StatusPill tone={item.status === 'reported' || item.status === 'pass' ? 'ok' : 'blocked'}>
            {label}
          </StatusPill>
          <strong>{item.title || item.id || '未命名'}</strong>
          {time && <time>{new Date(time).toLocaleString('zh-CN', { hour12: false })}</time>}
        </div>
        {isExpanded && (
          <div className='fx-aisre-timeline-detail'>
            {item.detail && <p>{item.detail}</p>}
            {item.content && <p>{item.content}</p>}
            {item.source_type && <p>来源: {item.source_type}</p>}
            {item.evidence_refs?.length > 0 && <p>证据引用: {item.evidence_refs.join('、')}</p>}
            {item.metadata && Object.keys(item.metadata).length > 0 && (
              <dl className='fx-aisre-evidence-meta'>
                {Object.entries(item.metadata).map(([k, v]) => (
                  <React.Fragment key={k}><dt>{k}</dt><dd>{String(v || '')}</dd></React.Fragment>
                ))}
              </dl>
            )}
            {item.blocker && <Blocked>{item.blocker}</Blocked>}
          </div>
        )}
      </div>
    </div>
  )
}

export function EvidenceSection({ evidence, onNavigate }) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [expandedId, setExpandedId] = useState(null)
  const [filterType, setFilterType] = useState('')
  const [filterTarget, setFilterTarget] = useState('')
  const [timeRange, setTimeRange] = useState('')

  useEffect(() => {
    let active = true
    setLoading(true); setError('')
    aiSreApi.evidenceChain({ type: filterType || undefined, target: filterTarget || undefined, range: timeRange || undefined })
      .then(next => { if (active) setData(next || {}) })
      .catch(err => { if (active) setError(formatAiSreError(err)) })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [filterType, filterTarget, timeRange])

  const allItems = useMemo(() => {
    const backendItems = (data?.items || []).map(item => ({
      ...item,
      _source: 'backend',
      _sortTime: item.updated_at || item.created_at || 0,
    }))
    const sessionItems = evidence.map(item => ({
      ...item,
      _source: 'session',
      _sortTime: item.at || 0,
      kind: item.category,
      title: item.title,
    }))
    return [...backendItems, ...sessionItems].sort((a, b) => {
      const ta = new Date(a._sortTime).getTime() || 0
      const tb = new Date(b._sortTime).getTime() || 0
      return tb - ta
    })
  }, [data, evidence])

  const filteredItems = useMemo(() => {
    let items = allItems
    if (filterType) items = items.filter(i => (i.kind || i.category) === filterType)
    return items
  }, [allItems, filterType])

  const handleExport = () => {
    alert('导出报告功能即将上线，敬请期待。')
  }

  return (
    <section className='fx-aisre-panel'>
      <div className='fx-aisre-toolbar'>
        <h2>Evidence Chain 时间线</h2>
        <button type='button' onClick={() => onNavigate({ section: 'diagnosis' })}>回到诊断</button>
        <button type='button' onClick={handleExport}>导出报告</button>
      </div>
      <p className='fx-aisre-note'>证据链按时间倒序展示所有后端和会话证据，点击展开查看详情。</p>
      <div className='fx-aisre-evidence-filters'>
        <select value={filterType} onChange={e => setFilterType(e.target.value)}>
          {evidenceTypeOptions.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
        <input value={filterTarget} onChange={e => setFilterTarget(e.target.value)} placeholder='目标 (主机/服务)' />
        <select value={timeRange} onChange={e => setTimeRange(e.target.value)}>
          <option value=''>全部时间</option>
          <option value='1h'>最近 1 小时</option>
          <option value='6h'>最近 6 小时</option>
          <option value='24h'>最近 24 小时</option>
          <option value='7d'>最近 7 天</option>
        </select>
      </div>
      {loading && <Empty>正在读取 Evidence Chain...</Empty>}
      <ErrorBox>{error}</ErrorBox>
      {!loading && !error && !filteredItems.length && (
        <div className='fx-aisre-empty'>
          <p><strong>暂无证据数据</strong></p>
          <p>Evidence Chain 聚合以下来源的数据：</p>
          <ul>
            <li>Prometheus / 指标采集</li>
            <li>告警系统 (AlertManager)</li>
            <li>日志系统 (Loki / ES)</li>
            <li>链路追踪 (Jaeger / Tempo)</li>
            <li>Agent 上报</li>
            <li>配置变更记录</li>
          </ul>
          <p>当以上数据源接入后，证据将自动出现在时间线中。</p>
        </div>
      )}
      {!loading && filteredItems.length > 0 && (
        <div className='fx-aisre-timeline'>
          <div className='fx-aisre-timeline-line' />
          {filteredItems.map((item, idx) => (
            <TimelineItem
              key={`${item._source}-${item.id || idx}-${item.kind || ''}`}
              item={item}
              isExpanded={expandedId === idx}
              onToggle={() => setExpandedId(expandedId === idx ? null : idx)}
            />
          ))}
        </div>
      )}
    </section>
  )
}
