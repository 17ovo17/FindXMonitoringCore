import React, { useEffect, useState } from 'react'
import { alertsApi } from '../../../api/alerts.js'
import { makeError } from '../alertModel.js'

/**
 * 业务组树组件
 * 对齐夜莺左侧业务组树：从 API 获取业务组列表，树形展示，点击过滤规则
 */
export function BusinessGroupTree({ selectedGroup, onSelect }) {
  const [groups, setGroups] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    setLoading(true)
    alertsApi.listBusinessGroups()
      .then((list) => setGroups(Array.isArray(list) ? list : []))
      .catch((err) => setError(makeError(err, '加载业务组失败')))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className='fx-alert-biz-tree'>
      <div className='fx-alert-biz-tree-head'>
        <span>业务组</span>
      </div>
      {error && <div className='fx-alert-message is-error'>{error}</div>}
      {loading && <div className='fx-alert-hint-text'>加载中...</div>}
      <div className='fx-alert-biz-tree-list'>
        <button
          type='button'
          className={!selectedGroup ? 'is-active' : ''}
          onClick={() => onSelect?.('')}
        >
          全部
        </button>
        {groups.map((group) => (
          <button
            key={group.id || group.name}
            type='button'
            className={selectedGroup === String(group.id || group.name) ? 'is-active' : ''}
            onClick={() => onSelect?.(String(group.id || group.name))}
          >
            {group.name || group.label || group.id}
          </button>
        ))}
      </div>
    </div>
  )
}
