import React, { useEffect, useMemo, useState } from 'react'
import { formatLogError, LOG_BLOCKERS, logsApi } from '../api/logs.js'
import { fieldGroups } from './logsModel.js'
import { Blocked, Field, Status } from './LogsShared.jsx'

/**
 * DEGRADE-051: 字段管理
 * 从 API 获取字段列表（名称/类型/是否索引），支持搜索过滤，支持切换索引状态
 */
export function FieldsSection() {
  const [q, setQ] = useState('')
  const [blocked, setBlocked] = useState('')
  const [fields, setFields] = useState([])
  const [meta, setMeta] = useState(null)
  const [toggling, setToggling] = useState(null)

  useEffect(() => {
    let alive = true
    logsApi.fields().then(resp => {
      if (!alive) return
      setFields(Array.isArray(resp?.fields) ? resp.fields : [])
      setMeta(resp || null)
      setBlocked(resp?.blocker || resp?.live_discovery?.blocker || '')
    }).catch(err => {
      if (!alive) return
      setFields([])
      setMeta(null)
      setBlocked(formatLogError(err))
    })
    return () => { alive = false }
  }, [])

  const toggleIndex = async (field) => {
    setToggling(field.name || field.key)
    try {
      const resp = await logsApi.toggleIndex(field.name || field.key, !field.indexed)
      if (resp?.blocker) {
        setBlocked(resp.blocker)
      } else {
        setFields(prev => prev.map(f =>
          (f.name || f.key) === (field.name || field.key) ? { ...f, indexed: !f.indexed } : f
        ))
        setBlocked('')
      }
    } catch (err) {
      if (err?.status === 404 || err?.status === 501) {
        setBlocked('BLOCKED_BY_CONTRACT: 后端不支持字段索引切换接口。')
      } else {
        setBlocked(formatLogError(err))
      }
    } finally {
      setToggling(null)
    }
  }

  const groups = useMemo(() => {
    const sourceGroups = fields.length ? fields.reduce((acc, item) => {
      const group = item.category || '字段目录'
      if (!acc[group]) acc[group] = { title: groupLabel(group), fields: [] }
      acc[group].fields.push({ ...item, name: item.key || item.name })
      return acc
    }, {}) : fieldGroups.reduce((acc, group) => {
      acc[group.title] = { title: group.title, fields: group.fields.map(field => ({ name: field, type: 'string', indexed: false })) }
      return acc
    }, {})
    return Object.values(sourceGroups).map(group => ({
      ...group,
      fields: group.fields.filter(field => !q || field.name.toLowerCase().includes(q.toLowerCase())),
    })).filter(group => group.fields.length)
  }, [fields, q])

  const indexedFields = useMemo(() => {
    const all = groups.flatMap(g => g.fields)
    return all.filter(f => f.indexed)
  }, [groups])

  return (
    <section className='fx-logs-split'>
      <div className='fx-logs-panel'>
        <Field label='搜索字段'>
          <input value={q} onChange={event => setQ(event.target.value)} placeholder='字段名 / 属性名' />
        </Field>
        {groups.map(group => (
          <div key={group.title} className='fx-logs-field-group'>
            <h3>{group.title}</h3>
            <table className='fx-field-table'>
              <thead>
                <tr><th>字段名</th><th>类型</th><th>索引</th><th>操作</th></tr>
              </thead>
              <tbody>
                {group.fields.map(field => (
                  <tr key={field.name}>
                    <td>{field.name}</td>
                    <td><span className='fx-field-type'>{field.type || 'string'}</span></td>
                    <td>{field.indexed ? <span style={{ color: '#167346' }}>是</span> : <span style={{ color: '#8c8c8c' }}>否</span>}</td>
                    <td>
                      <button
                        type='button'
                        className='fx-qb__btn fx-qb__btn--ghost'
                        style={{ minHeight: 24, fontSize: 11, padding: '0 8px' }}
                        disabled={toggling === field.name}
                        onClick={() => toggleIndex(field)}
                      >
                        {toggling === field.name ? '...' : (field.indexed ? '取消索引' : '添加索引')}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ))}
      </div>
      <div className='fx-logs-panel'>
        <h3>已索引字段（{indexedFields.length}）</h3>
        <Status ok={fields.length > 0}>{fields.length ? '字段目录已加载' : '待契约'}</Status>
        <p>来源：本地字段目录；实时字段发现：{meta?.live_discovery?.status || 'BLOCKED_BY_CONTRACT'}</p>
        {indexedFields.length > 0 && (
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginTop: 8 }}>
            {indexedFields.map(f => (
              <span key={f.name} className='fx-chip is-active'>{f.name}</span>
            ))}
          </div>
        )}
        {!fields.length && <Blocked>{LOG_BLOCKERS.fields}</Blocked>}
        {blocked && <Blocked>{blocked}</Blocked>}
      </div>
    </section>
  )
}

const groupLabel = key => ({
  resource: '资源字段',
  log: '日志字段',
  http: 'HTTP 字段',
  exception: '异常字段',
})[key] || key
