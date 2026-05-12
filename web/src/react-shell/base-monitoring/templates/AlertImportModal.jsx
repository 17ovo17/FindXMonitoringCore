/**
 * 告警规则导入 Modal (T07)
 * 支持选择数据源替换 + 业务组选择
 */
import React, { useEffect, useState } from 'react'
import { get, post } from '../../api/http.js'

export function AlertImportModal({ rows, componentIdent, onClose, onDone }) {
  const [groups, setGroups] = useState([])
  const [datasources, setDatasources] = useState([])
  const [groupId, setGroupId] = useState('')
  const [datasourceId, setDatasourceId] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState(null)

  useEffect(() => {
    if (!rows.length) return
    setLoading(true)
    setError('')
    setResult(null)
    Promise.all([
      get('/monitor/busi-groups', { params: { limit: 200 } }).catch(() => []),
      get('/monitor/datasources', { params: { limit: 200 } }).catch(() => []),
    ]).then(([groupData, dsData]) => {
      const groupList = Array.isArray(groupData) ? groupData : (groupData?.dat || groupData?.list || [])
      setGroups(groupList)
      const dsList = Array.isArray(dsData) ? dsData : (dsData?.dat || dsData?.list || [])
      setDatasources(dsList)
    }).finally(() => setLoading(false))
  }, [rows.length])

  if (!rows.length) return null

  const handleSubmit = async () => {
    if (!groupId) {
      setError('请选择业务组')
      return
    }
    setSubmitting(true)
    setError('')
    try {
      const rules = rows.map((row) => {
        let content = row.content
        try {
          const parsed = typeof content === 'string' ? JSON.parse(content) : content
          const rule = Array.isArray(parsed) ? parsed[0] : parsed
          return {
            ...rule,
            group_id: Number(groupId),
            ...(datasourceId ? { datasource_ids: [Number(datasourceId)] } : {}),
          }
        } catch {
          return { name: row.name, group_id: Number(groupId) }
        }
      })
      await post('/monitor/busi-group/' + groupId + '/alert-rules', rules)
      setResult({ count: rules.length })
      onDone?.()
    } catch (err) {
      setError(err.message || '导入失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className='fx-tpl-modal' role='dialog' aria-modal='true'>
      <div className='fx-tpl-modal__backdrop' onClick={onClose} />
      <section className='fx-tpl-modal__panel'>
        <header className='fx-tpl-modal__head'>
          <h2>导入告警规则</h2>
          <button type='button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        <div className='fx-tpl-form'>
          <div style={{ marginBottom: 12 }}>
            <strong>已选规则：</strong>
            {rows.map((r) => <span key={r.id} className='fx-tpl-tag'>{r.name}</span>)}
          </div>
          <label>
            目标业务组
            <select value={groupId} onChange={(e) => setGroupId(e.target.value)} disabled={loading}>
              <option value=''>{loading ? '加载中...' : '请选择业务组'}</option>
              {groups.map((g) => (
                <option key={g.id} value={g.id}>{g.name || g.id}</option>
              ))}
            </select>
          </label>
          <label>
            数据源替换（可选）
            <select value={datasourceId} onChange={(e) => setDatasourceId(e.target.value)} disabled={loading}>
              <option value=''>使用模板默认数据源</option>
              {datasources.map((ds) => (
                <option key={ds.id} value={ds.id}>{ds.name || ds.id} ({ds.plugin_type || ds.type || ''})</option>
              ))}
            </select>
          </label>
        </div>
        {error && <div className='fx-tpl-alert is-error'>{error}</div>}
        {result && <div className='fx-tpl-alert is-success'>成功导入 {result.count} 条告警规则</div>}
        <footer className='fx-tpl-modal__foot'>
          <button type='button' onClick={onClose}>关闭</button>
          <button type='button' className='is-primary' onClick={handleSubmit} disabled={submitting || loading || !groupId}>
            {submitting ? '导入中...' : '确认导入'}
          </button>
        </footer>
      </section>
    </div>
  )
}
