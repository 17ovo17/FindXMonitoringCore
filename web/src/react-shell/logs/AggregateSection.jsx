import React, { useEffect, useState } from 'react'
import { formatLogError, LOG_BLOCKERS, LOG_SOURCES, logsApi } from '../api/logs.js'
import { Blocked, Empty, Field, JsonPreview } from './LogsShared.jsx'

export function AggregateSection() {
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [source, setSource] = useState('findx_audit')
  const [groupBy, setGroupBy] = useState('status')
  const [buckets, setBuckets] = useState([])
  const [meta, setMeta] = useState(null)

  const runAggregate = async () => {
    setLoading(true)
    setBlocked('')
    if (source !== 'findx_audit') {
      setBuckets([])
      setMeta(null)
      setBlocked(LOG_BLOCKERS.aggregate)
      setLoading(false)
      return
    }
    try {
      const resp = await logsApi.aggregate({ source, group_by: groupBy, limit: 200 })
      setBuckets(Array.isArray(resp?.buckets) ? resp.buckets : [])
      setMeta(resp || null)
    } catch (err) {
      setBuckets([])
      setMeta(null)
      setBlocked(formatLogError(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    runAggregate()
  }, [])

  return (
    <section className='fx-logs-work'>
      <div className='fx-logs-filter'>
        <Field label='来源'>
          <select value={source} onChange={event => setSource(event.target.value)}>
            {LOG_SOURCES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
          </select>
        </Field>
        <Field label='分组'><select value={groupBy} onChange={event => setGroupBy(event.target.value)}><option value='status'>状态</option><option value='action'>动作</option><option value='resource_type'>资源类型</option><option value='actor'>操作者</option></select></Field>
        <Field label='函数'><select disabled><option>count()</option></select></Field>
        <button type='button' onClick={runAggregate} disabled={loading}>{loading ? '分析中' : '分析'}</button>
      </div>
      {source !== 'findx_audit' && <Blocked>{LOG_BLOCKERS.aggregate}</Blocked>}
      {blocked && <Blocked>{blocked}</Blocked>}
      {buckets.length ? (
        <div className='fx-logs-table'><table><thead><tr><th>分组</th><th>计数</th><th>来源</th></tr></thead><tbody>{buckets.map(bucket => <tr key={bucket.key}><td>{bucket.label || bucket.key}</td><td>{bucket.count}</td><td>{meta?.source_name || 'FindX 审计日志'}</td></tr>)}</tbody></table></div>
      ) : (
        <Empty>{source === 'findx_audit' && !blocked ? '暂无可聚合的 FindX 审计日志。' : '通用生产日志聚合仍为 BLOCKED_BY_CONTRACT，未绘制静态趋势图。'}</Empty>
      )}
      <JsonPreview value={{
        source,
        group_by: groupBy,
        total: meta?.total ?? 0,
        buckets: buckets.length,
        blocked: source === 'findx_audit' ? meta?.blocker || null : LOG_BLOCKERS.aggregate,
        note: '当前聚合只覆盖 FindX 审计日志；生产 OTel 日志聚合未接入。',
      }} />
    </section>
  )
}
