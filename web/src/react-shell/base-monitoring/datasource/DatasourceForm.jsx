import React, { useState } from 'react'
import { datasourceApi } from '../../api/datasources.js'
import { redactText } from '../../api/http.js'

const AUTH_OPTIONS = [
  { value: 'none', label: '无认证' },
  { value: 'basic', label: 'Basic Auth' },
  { value: 'bearer', label: 'Bearer Token' },
]

function HeaderRows({ headers, onChange }) {
  const add = () => onChange([...headers, { key: '', value: '' }])
  const remove = (i) => onChange(headers.filter((_, idx) => idx !== i))
  const update = (i, field, val) => {
    const next = headers.map((h, idx) => idx === i ? { ...h, [field]: val } : h)
    onChange(next)
  }

  return (
    <label className='is-wide'>
      自定义 Header
      {headers.map((h, i) => (
        <div key={i} className='fx-ds-header-row'>
          <input placeholder='Key' value={h.key} onChange={(e) => update(i, 'key', e.target.value)} />
          <input placeholder='Value' value={h.value} onChange={(e) => update(i, 'value', e.target.value)} />
          <button type='button' onClick={() => remove(i)} aria-label='删除'>×</button>
        </div>
      ))}
      <button type='button' onClick={add}>+ 添加 Header</button>
    </label>
  )
}

function buildPayload(form, dsType) {
  const payload = {
    name: form.name,
    type: dsType.type,
    plugin_type: dsType.type,
    url: form.url,
    cluster_name: form.cluster || 'default',
    status: 'enabled',
    settings: {},
  }
  if (form.auth === 'basic') {
    payload.settings.basic_auth_user = form.username
    payload.settings.basic_auth_password = form.password
  } else if (form.auth === 'bearer') {
    payload.settings.bearer_token = form.token
  }
  if (form.tls) {
    payload.settings.tls_enabled = true
    payload.settings.ca_cert = form.caCert
  }
  if (form.headers.length > 0) {
    payload.settings.custom_headers = Object.fromEntries(
      form.headers.filter((h) => h.key).map((h) => [h.key, h.value])
    )
  }
  if (dsType.type === 'prometheus') {
    payload.settings.scrape_interval = form.scrapeInterval || '15s'
  }
  return payload
}

const INITIAL_FORM = { name: '', url: '', auth: 'none', username: '', password: '', token: '', tls: false, caCert: '', headers: [], scrapeInterval: '15s', cluster: '' }

export function DatasourceForm({ dsType, editRow, onSaved, onCancel }) {
  const [form, setForm] = useState(() => {
    if (editRow) {
      return {
        name: editRow.name || '',
        url: editRow.url || editRow.address || editRow.endpoint || '',
        auth: editRow.settings?.basic_auth_user ? 'basic' : editRow.settings?.bearer_token ? 'bearer' : 'none',
        username: editRow.settings?.basic_auth_user || '',
        password: editRow.settings?.basic_auth_password || '',
        token: editRow.settings?.bearer_token || '',
        tls: !!editRow.settings?.tls_enabled,
        caCert: editRow.settings?.ca_cert || '',
        headers: Object.entries(editRow.settings?.custom_headers || {}).map(([key, value]) => ({ key, value })),
        scrapeInterval: editRow.settings?.scrape_interval || '15s',
        cluster: editRow.cluster_name || editRow.cluster || '',
      }
    }
    return { ...INITIAL_FORM }
  })
  const [testing, setTesting] = useState(false)
  const [saving, setSaving] = useState(false)
  const [msg, setMsg] = useState({ type: '', text: '' })

  const set = (key, val) => setForm((prev) => ({ ...prev, [key]: val }))
  const isPrometheus = dsType.type === 'prometheus'
  const valid = form.name.trim() && form.url.trim()

  const testConnection = async () => {
    setTesting(true)
    setMsg({ type: '', text: '' })
    try {
      await datasourceApi.testConnection()
      setMsg({ type: 'success', text: '连接测试成功' })
      return true
    } catch (err) {
      setMsg({ type: 'error', text: `连接测试失败：${redactText(err?.message || '未知错误')}` })
      return false
    } finally {
      setTesting(false)
    }
  }

  const doSave = async () => {
    setSaving(true)
    setMsg({ type: '', text: '' })
    try {
      const payload = buildPayload(form, dsType)
      if (editRow?.id || editRow?.datasource_id) {
        payload.id = editRow.id || editRow.datasource_id
      }
      await datasourceApi.save(payload)
      setMsg({ type: 'success', text: '保存成功' })
      onSaved?.()
    } catch (err) {
      setMsg({ type: 'error', text: `保存失败：${redactText(err?.message || '未知错误')}` })
    } finally {
      setSaving(false)
    }
  }

  const saveAndTest = async () => {
    const ok = await testConnection()
    if (ok) {
      await doSave()
    } else {
      if (window.confirm('连接测试失败，是否仍要保存？')) await doSave()
    }
  }

  return (
    <section className='fx-ds-form'>
      <header className='fx-ds-modal__head'>
        <h2>{editRow ? '编辑' : '新增'} {dsType.name} 数据源</h2>
        <button type='button' className='fx-ds-icon-button' onClick={onCancel} aria-label='返回'>×</button>
      </header>
      {!dsType.supported && (
        <div className='fx-ds-alert is-warning'>该数据源类型的查询功能尚在开发中</div>
      )}
      {msg.text && <div className={`fx-ds-alert ${msg.type === 'error' ? 'is-error' : 'is-success'}`}>{msg.text}</div>}
      <div className='fx-ds-form__grid'>
        <label>名称（必填）<input value={form.name} onChange={(e) => set('name', e.target.value)} placeholder='my-prometheus' /></label>
        <label>URL（必填）<input value={form.url} onChange={(e) => set('url', e.target.value)} placeholder='http://localhost:9090' /></label>
        <label>集群<input value={form.cluster} onChange={(e) => set('cluster', e.target.value)} placeholder='default' /></label>
        <label>认证方式
          <select value={form.auth} onChange={(e) => set('auth', e.target.value)}>
            {AUTH_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
        </label>
        {form.auth === 'basic' && <>
          <label>用户名<input value={form.username} onChange={(e) => set('username', e.target.value)} /></label>
          <label>密码<input type='password' value={form.password} onChange={(e) => set('password', e.target.value)} /></label>
        </>}
        {form.auth === 'bearer' && (
          <label className='is-wide'>Bearer Token<input type='password' value={form.token} onChange={(e) => set('token', e.target.value)} /></label>
        )}
        <label>
          <span><input type='checkbox' checked={form.tls} onChange={(e) => set('tls', e.target.checked)} /> 启用 TLS</span>
        </label>
        {form.tls && <label className='is-wide'>CA 证书<textarea rows='3' value={form.caCert} onChange={(e) => set('caCert', e.target.value)} /></label>}
        {isPrometheus && <label>Scrape Interval<input value={form.scrapeInterval} onChange={(e) => set('scrapeInterval', e.target.value)} placeholder='15s' /></label>}
        <HeaderRows headers={form.headers} onChange={(h) => set('headers', h)} />
      </div>
      <footer>
        <button type='button' onClick={onCancel}>取消</button>
        <button type='button' onClick={testConnection} disabled={!valid || testing}>{testing ? '测试中...' : '测试连接'}</button>
        <button type='button' onClick={saveAndTest} disabled={!valid || saving || testing}>保存并测试</button>
        <button type='button' className='is-primary' onClick={doSave} disabled={!valid || saving}>{saving ? '保存中...' : '强制保存'}</button>
      </footer>
    </section>
  )
}