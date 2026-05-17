import React, { useCallback, useEffect, useState } from 'react'
import { formatTracingError, tracingApi } from '../api/tracing.js'
import { displayText, fmtTime } from './tracingModel.js'
import { Blocked, Empty, ErrorBox, Field, Status } from './TracingShared.jsx'

const SEVERITY_OPTIONS = [
  { value: '', label: '全部' },
  { value: 'critical', label: '严重' },
  { value: 'warning', label: '警告' },
  { value: 'notice', label: '通知' },
]

function AlarmDetail({ alarm, onClose }) {
  if (!alarm) return null
  return (
    <div className='fx-tracing-modal'>
      <div className='fx-tracing-modal__body'>
        <header>
          <h2>告警详情</h2>
          <button type='button' onClick={onClose}>关闭</button>
        </header>
        <div className='fx-tracing-form'>
          <Field label='告警消息'><div style={{ padding: '6px 0', fontSize: 13 }}>{displayText(alarm.message || alarm.name)}</div></Field>
          <Field label='告警对象'><div style={{ padding: '6px 0', fontSize: 13 }}>{displayText(alarm.scope || alarm.entityName || alarm.serviceName)}</div></Field>
          <Field label='级别'><Status>{displayText(alarm.severity || alarm.level || '-')}</Status></Field>
          <Field label='开始时间'><div style={{ padding: '6px 0', fontSize: 13 }}>{fmtTime(alarm.startTime || alarm.start_time)}</div></Field>
          <Field label='规则名称'><div style={{ padding: '6px 0', fontSize: 13 }}>{displayText(alarm.ruleName || alarm.rule || '-')}</div></Field>
          <Field label='标签'>
            <div style={{ padding: '6px 0', fontSize: 12 }}>
              {(alarm.tags || []).map((t, i) => (
                <span key={i} style={{ display: 'inline-block', marginRight: 6, padding: '2px 6px', background: '#eef3fb', borderRadius: 4 }}>
                  {typeof t === 'string' ? t : `${t.key}=${t.value}`}
                </span>
              ))}
              {(!alarm.tags || !alarm.tags.length) && '-'}
            </div>
          </Field>
        </div>
      </div>
    </div>
  )
}
export function AlarmsSection() {
  const [filters, setFilters] = useState({ keyword: '', severity: '', scope: '' })
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [detail, setDetail] = useState(null)
  const patch = (key, value) => setFilters(prev => ({ ...prev, [key]: value }))

  const load = useCallback(async () => {
    setLoading(true); setError(''); setBlocked('')
    try {
      const params = {}
      if (filters.keyword) params.keyword = filters.keyword
      if (filters.severity) params.severity = filters.severity
      if (filters.scope) params.scope = filters.scope
      const data = await tracingApi.alarms.list(params)
      setRows(data || [])
    } catch (err) {
      setRows([])
      const msg = formatTracingError(err)
      if ([404, 405, 501].includes(err?.status)) {
        setBlocked('需要后端实现 /apm/alarms GET API')
      } else { setError(msg) }
    } finally { setLoading(false) }
  }, [filters.keyword, filters.severity, filters.scope])

  useEffect(() => { load() }, [load])

  const ackAlarm = async (id) => {
    try {
      await tracingApi.alarms.ack(id)
      load()
    } catch (err) {
      const msg = formatTracingError(err)
      if ([404, 405, 501].includes(err?.status)) {
        setBlocked('需要后端实现 /apm/alarms/:id/ack POST API')
      } else { setError(msg) }
    }
  }

  return (
    <section className='fx-tracing-work'>
      <div className='fx-tracing-condition-bar'>
        <Field label='关键字' className='is-flex'>
          <input value={filters.keyword} onChange={e => patch('keyword', e.target.value)} placeholder='搜索告警消息' />
        </Field>
        <Field label='级别'>
          <select value={filters.severity} onChange={e => patch('severity', e.target.value)}>
            {SEVERITY_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
        </Field>
        <Field label='范围'>
          <input value={filters.scope} onChange={e => patch('scope', e.target.value)} placeholder='服务/实例/端点' />
        </Field>
        <div className='fx-tracing-condition-actions'>
          <button type='button' className='is-primary' onClick={load}>{loading ? '查询中...' : '查询告警'}</button>
        </div>
      </div>

      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}

      <div className='fx-tracing-table'>
        <h3>链路告警列表 ({rows.length})</h3>
        <table>
          <thead><tr><th>告警消息</th><th>对象</th><th>级别</th><th>开始时间</th><th>操作</th></tr></thead>
          <tbody>
            {rows.map(row => (
              <tr key={row.id || row.message}>
                <td>{displayText(row.message || row.name)}</td>
                <td>{displayText(row.scope || row.entityName || row.serviceName)}</td>
                <td><Status ok={row.severity !== 'critical'}>{displayText(row.severity || row.level || '-')}</Status></td>
                <td>{fmtTime(row.startTime || row.start_time)}</td>
                <td className='fx-tracing-actions'>
                  <button type='button' onClick={() => setDetail(row)}>详情</button>
                  <button type='button' onClick={() => ackAlarm(row.id)}>确认</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!rows.length && !loading && <Empty>{blocked || '暂无告警数据'}</Empty>}
      </div>

      {detail && <AlarmDetail alarm={detail} onClose={() => setDetail(null)} />}
    </section>
  )
}
