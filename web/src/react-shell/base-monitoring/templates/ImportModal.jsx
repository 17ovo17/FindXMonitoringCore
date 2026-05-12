/**
 * 导入 Modal
 * 支持仪表盘模板导入到业务组
 */
import React, { useEffect, useState } from 'react'
import { templateApi } from './templateApi.js'
import { get } from '../../api/http.js'

export function ImportModal({ rows, componentIdent, onClose }) {
  const [groups, setGroups] = useState([])
  const [groupId, setGroupId] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState(null)

  useEffect(() => {
    if (!rows.length) return
    setLoading(true)
    setError('')
    setResult(null)
    get('/monitor/busi-groups', { params: { limit: 200 } })
      .then((data) => {
        const list = Array.isArray(data) ? data : (data?.dat || data?.list || data?.items || [])
        setGroups(list)
      })
      .catch(() => setGroups([]))
      .finally(() => setLoading(false))
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
      const results = []
      for (const row of rows) {
        const id = row.id || row.uuid
        const resp = await templateApi.importDashboardTemplate(id, {
          resource_group_id: groupId,
          title: row.name,
        })
        results.push(resp)
      }
      setResult({ count: results.length })
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
          <h2>导入仪表盘模板</h2>
          <button type='button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        <div className='fx-tpl-form'>
          <div style={{ marginBottom: 12 }}>
            <strong>已选模板：</strong>
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
        </div>
        {error && <div className='fx-tpl-alert is-error'>{error}</div>}
        {result && <div className='fx-tpl-alert is-success'>成功导入 {result.count} 个仪表盘模板</div>}
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
