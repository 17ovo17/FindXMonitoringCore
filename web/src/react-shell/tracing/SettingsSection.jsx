import React, { useCallback, useEffect, useState } from 'react'
import { formatTracingError, tracingApi } from '../api/tracing.js'
import { displayText } from './tracingModel.js'
import { Blocked, Empty, ErrorBox, Field } from './TracingShared.jsx'

const defaultSettings = {
  recordDataTTL: '',
  metricsDataTTL: '',
  otherMetricsDataTTL: '',
}

const isBlockedResponse = resp => {
  if (!resp || typeof resp !== 'object') return false
  const code = String(resp.code || resp.error_code || '').toUpperCase()
  const status = String(resp.status || '').toLowerCase()
  const message = String(resp.message || resp.error || resp.reason || '')
  return code === '' || status === 'blocked'
}

export function SettingsSection() {
  const [draft, setDraft] = useState(defaultSettings)
  const [nodes, setNodes] = useState([])
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [saveMsg, setSaveMsg] = useState('')
  const patch = (key, value) => setDraft(prev => ({ ...prev, [key]: value }))

  const load = useCallback(async () => {
    setLoading(true); setError(''); setBlocked('')
    try {
      const resp = await tracingApi.settings.get()
      if (resp && typeof resp === 'object') {
        setDraft({
          recordDataTTL: resp.recordDataTTL || resp.retention_days || '',
          metricsDataTTL: resp.metricsDataTTL || resp.metrics_ttl || '',
          otherMetricsDataTTL: resp.otherMetricsDataTTL || resp.other_ttl || '',
        })
        if (resp.clusterNodes || resp.nodes) {
          setNodes(resp.clusterNodes || resp.nodes || [])
        }
      } else {
        setBlocked('需要后端实现 /apm/settings GET API')
      }
    } catch (err) {
      const msg = formatTracingError(err)
      if ([404, 405, 501].includes(err?.status)) {
        setBlocked('需要后端实现 /apm/settings GET API')
      } else { setError(msg) }
    } finally { setLoading(false) }
  }, [])

  useEffect(() => { load() }, [load])
  const save = async () => {
    setSaving(true); setSaveMsg(''); setError(''); setBlocked('')
    try {
      const resp = await tracingApi.settings.save({
        recordDataTTL: Number(draft.recordDataTTL) || undefined,
        metricsDataTTL: Number(draft.metricsDataTTL) || undefined,
        otherMetricsDataTTL: Number(draft.otherMetricsDataTTL) || undefined,
      })
      if (isBlockedResponse(resp)) {
        setBlocked(formatTracingError(new Error(resp.message || resp.error || '链路查询服务设置保存被契约阻断')))
        return
      }
      setSaveMsg('设置已保存')
    } catch (err) {
      const msg = formatTracingError(err)
      if ([404, 405, 501].includes(err?.status)) {
        setBlocked('需要后端实现 /apm/settings PUT API')
      } else { setError(msg) }
    } finally { setSaving(false) }
  }

  return (
    <section className='fx-tracing-work'>
      <h3 style={{ margin: '0 0 12px', fontSize: 15 }}>TTL 配置</h3>
      <div className='fx-tracing-form'>
        <Field label='Trace 记录保留期（天）'>
          <input type='number' value={draft.recordDataTTL} onChange={e => patch('recordDataTTL', e.target.value)} placeholder='例如: 3' min='1' />
        </Field>
        <Field label='指标数据保留期（天）'>
          <input type='number' value={draft.metricsDataTTL} onChange={e => patch('metricsDataTTL', e.target.value)} placeholder='例如: 7' min='1' />
        </Field>
        <Field label='其他指标保留期（天）'>
          <input type='number' value={draft.otherMetricsDataTTL} onChange={e => patch('otherMetricsDataTTL', e.target.value)} placeholder='例如: 30' min='1' />
        </Field>
      </div>
      <div className='fx-tracing-toolbar'>
        <button type='button' onClick={save} disabled={saving}>{saving ? '保存中...' : '保存设置'}</button>
        <button type='button' onClick={load}>{loading ? '加载中...' : '重新加载'}</button>
        {saveMsg && <span style={{ color: '#167346', fontSize: 13 }}>{saveMsg}</span>}
      </div>
      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}

      <h3 style={{ margin: '16px 0 8px', fontSize: 15 }}>集群节点列表</h3>
      <div className='fx-tracing-table'>
        <table>
          <thead><tr><th>节点名称</th><th>地址</th><th>角色</th><th>状态</th></tr></thead>
          <tbody>
            {nodes.map((node, idx) => (
              <tr key={node.id || node.name || idx}>
                <td>{displayText(node.name || node.host || node.id)}</td>
                <td style={{ fontFamily: 'monospace', fontSize: 12 }}>{displayText(node.address || node.host || '-')}</td>
                <td>{displayText(node.role || node.type || '链路查询节点')}</td>
                <td>{displayText(node.status || node.state || '未知')}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {!nodes.length && !loading && <Empty>{blocked || '需要后端实现 /apm/settings 返回 clusterNodes 字段'}</Empty>}
      </div>
    </section>
  )
}
