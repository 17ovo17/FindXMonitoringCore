import React, { useMemo, useState } from 'react'
import { displayDate, severityLabel, statusLabel } from '../alertModel.js'

/**
 * 事件聚合视图组件
 * 对齐夜莺事件聚合：按规则聚合 / 按目标聚合
 */
export function AggregatedView({ events, groupKey, onOpenDetail }) {
  const groups = useMemo(() => {
    const map = {}
    for (const event of events) {
      const key = event[groupKey] || '-'
      if (!map[key]) map[key] = { key, events: [] }
      map[key].events.push(event)
    }
    return Object.values(map).sort((a, b) => b.events.length - a.events.length)
  }, [events, groupKey])

  const [expanded, setExpanded] = useState({})
  const toggle = (key) => setExpanded((prev) => ({ ...prev, [key]: !prev[key] }))

  return (
    <div className='fx-alert-aggregated'>
      {groups.map((group) => (
        <div key={group.key} className='fx-alert-agg-group'>
          <button type='button' className='fx-alert-agg-header' onClick={() => toggle(group.key)}>
            <span>{expanded[group.key] ? '▼' : '▶'}</span>
            <strong>{group.key}</strong>
            <span className='fx-alert-tag'>{group.events.length} 个事件</span>
          </button>
          {expanded[group.key] && (
            <div className='fx-alert-agg-list'>
              {group.events.map((event) => (
                <div key={event.id} className='fx-alert-agg-item'>
                  <button type='button' className='fx-alert-link' onClick={() => onOpenDetail(event)}>{event.name}</button>
                  <span className={`fx-alert-severity is-${event.severity}`}>{severityLabel(event.severity)}</span>
                  <span>{statusLabel(event.status)}</span>
                  <span>{displayDate(event.lastSeen)}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      ))}
      {groups.length === 0 && <div className='fx-alert-empty'>暂无事件</div>}
    </div>
  )
}
