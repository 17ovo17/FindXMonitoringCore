import React, { useRef, useState } from 'react'
import { datasourceApi } from '../../api/datasources.js'
import { redactText } from '../../api/http.js'

const AUTH_OPTIONS = [
  { value: 'none', label: '无认证' },
  { value: 'basic', label: 'Basic Auth' },
  { value: 'bearer', label: 'Bearer Token' },
]

const TSDB_TYPES = ['Prometheus', 'Thanos', 'VictoriaMetrics', 'M3', 'SLS']

const ES_VERSIONS = ['6.x', '7.x', '8.x']

// --- 表单验证 ---
function validate(form, dsType) {
  const errors = {}
  if (!form.name.trim()) errors.name = '名称为必填项'
  else if (form.name.trim().length < 3) errors.name = '名称至少 3 个字符'
  if (!form.url.trim()) errors.url = 'URL 为必填项'
  else if (/\s/.test(form.url)) errors.url = 'URL 不能包含空格'
  if (form.auth === 'basic' && !form.username.trim()) errors.username = '选择 Basic Auth 时用户名为必填项'
  return errors
}

function FieldLabel({ label, name, errors, children, wide }) {
  const hasError = !!errors[name]
  return (
    <label className={`${wide ? 'is-wide' : ''} ${hasError ? 'fx-ds-field-error' : ''}`} data-field={name}>
      {label}
      {children}
      {hasError && <span className='fx-ds-error-msg'>{errors[name]}</span>}
    </label>
  )
}

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

function SwitchInput({ checked, onChange, label }) {
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
      <span className='fx-ds-switch'>
        <input type='checkbox' checked={checked} onChange={(e) => onChange(e.target.checked)} />
        <span className='fx-ds-switch__track' />
      </span>
      <span style={{ fontSize: 13, color: '#4d5f78' }}>{label}</span>
    </span>
  )
}

// --- 类型专属配置 ---
function PrometheusSettings({ form, set }) {
  return (
    <div className='fx-ds-type-settings'>
      <h4>Prometheus 专属配置</h4>
      <div className='fx-ds-type-settings__grid'>
        <label>Remote Write URL
          <input value={form.writeAddr} onChange={(e) => set('writeAddr', e.target.value)} placeholder='http://localhost:9090/api/v1/write' />
        </label>
        <label>时序库类型
          <select value={form.tsdbType} onChange={(e) => set('tsdbType', e.target.value)}>
            <option value=''>请选择</option>
            {TSDB_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
          </select>
        </label>
        <label>Scrape Interval
          <input value={form.scrapeInterval} onChange={(e) => set('scrapeInterval', e.target.value)} placeholder='15s' />
        </label>
        <label>External Labels
          <input value={form.externalLabels} onChange={(e) => set('externalLabels', e.target.value)} placeholder='key=value,key2=value2' />
        </label>
        <label className='is-wide'>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
            <input type='checkbox' checked={form.isDefault} onChange={(e) => set('isDefault', e.target.checked)} />
            设为默认数据源
          </span>
        </label>
        <label className='is-wide'>备注
          <textarea rows='2' value={form.description} onChange={(e) => set('description', e.target.value)} placeholder='可选备注信息' />
        </label>
      </div>
    </div>
  )
}

function ElasticsearchSettings({ form, set }) {
  return (
    <div className='fx-ds-type-settings'>
      <h4>Elasticsearch 专属配置</h4>
      <div className='fx-ds-type-settings__grid'>
        <label>Version
          <select value={form.esVersion} onChange={(e) => set('esVersion', e.target.value)}>
            <option value=''>请选择版本</option>
            {ES_VERSIONS.map((v) => <option key={v} value={v}>{v}</option>)}
          </select>
        </label>
        <label>Max Concurrent Shard Requests
          <input type='number' min='0' value={form.esMaxShard} onChange={(e) => set('esMaxShard', e.target.value)} placeholder='5' />
        </label>
        <label>Min Interval (秒)
          <input type='number' min='0' value={form.esMinInterval} onChange={(e) => set('esMinInterval', e.target.value)} placeholder='10' />
        </label>
        <label>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
            <input type='checkbox' checked={form.esEnableWrite} onChange={(e) => set('esEnableWrite', e.target.checked)} />
            启用写入
          </span>
        </label>
        <label className='is-wide'>备注
          <textarea rows='2' value={form.description} onChange={(e) => set('description', e.target.value)} placeholder='可选备注信息' />
        </label>
      </div>
    </div>
  )
}

function LokiSettings({ form, set }) {
  return (
    <div className='fx-ds-type-settings'>
      <h4>Loki 专属配置</h4>
      <div className='fx-ds-type-settings__grid'>
        <label>Max Lines
          <input type='number' min='1' value={form.lokiMaxLines} onChange={(e) => set('lokiMaxLines', e.target.value)} placeholder='1000' />
        </label>
        <label className='is-wide'>备注
          <textarea rows='2' value={form.description} onChange={(e) => set('description', e.target.value)} placeholder='可选备注信息' />
        </label>
      </div>
    </div>
  )
}

function JaegerSettings({ form, set }) {
  return (
    <div className='fx-ds-type-settings'>
      <h4>Jaeger 专属配置</h4>
      <div className='fx-ds-type-settings__grid'>
        <label>Version
          <select value={form.jaegerVersion} onChange={(e) => set('jaegerVersion', e.target.value)}>
            <option value='v3'>v3</option>
          </select>
        </label>
        <label className='is-wide'>备注
          <textarea rows='2' value={form.description} onChange={(e) => set('description', e.target.value)} placeholder='可选备注信息' />
        </label>
      </div>
    </div>
  )
}

function TypeSpecificSettings({ dsType, form, set }) {
  switch (dsType.type) {
    case 'prometheus': return <PrometheusSettings form={form} set={set} />
    case 'elasticsearch': return <ElasticsearchSettings form={form} set={set} />
    case 'loki': return <LokiSettings form={form} set={set} />
    case 'jaeger': return <JaegerSettings form={form} set={set} />
    default: return null
  }
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
  // TLS
  payload.settings.skip_tls_verify = form.skipTlsVerify
  if (form.caCert) payload.settings.ca_cert = form.caCert
  if (form.serverName) payload.settings.server_name = form.serverName
  if (form.clientCert) payload.settings.client_cert = form.clientCert
  if (form.clientKey) payload.settings.client_key = form.clientKey
  if (form.headers.length > 0) {
    payload.settings.custom_headers = Object.fromEntries(
      form.headers.filter((h) => h.key).map((h) => [h.key, h.value])
    )
  }
  // 类型专属
  if (dsType.type === 'prometheus') {
    payload.settings.scrape_interval = form.scrapeInterval || '15s'
    if (form.writeAddr) payload.settings.write_addr = form.writeAddr
    if (form.tsdbType) payload.settings['prometheus.tsdb_type'] = form.tsdbType
    if (form.externalLabels) payload.settings.external_labels = form.externalLabels
    if (form.isDefault) payload.is_default = true
    if (form.description) payload.description = form.description
  } else if (dsType.type === 'elasticsearch') {
    if (form.esVersion) payload.settings.version = form.esVersion
    payload.settings.max_shard = Number(form.esMaxShard) || 5
    payload.settings.min_interval = Number(form.esMinInterval) || 10
    payload.settings.enable_write = form.esEnableWrite
    if (form.description) payload.description = form.description
  } else if (dsType.type === 'loki') {
    payload.settings.max_lines = Number(form.lokiMaxLines) || 1000
    if (form.description) payload.description = form.description
  } else if (dsType.type === 'jaeger') {
    payload.settings.version = form.jaegerVersion || 'v3'
    if (form.description) payload.description = form.description
  }
  return payload
}

const INITIAL_FORM = {
  name: '', url: '', auth: 'none', username: '', password: '', token: '',
  skipTlsVerify: false, caCert: '', serverName: '', clientCert: '', clientKey: '',
  headers: [], scrapeInterval: '15s', cluster: '',
  writeAddr: '', tsdbType: '', externalLabels: '', isDefault: false, description: '',
  esVersion: '', esMaxShard: '5', esMinInterval: '10', esEnableWrite: false,
  lokiMaxLines: '1000', jaegerVersion: 'v3',
}

export function DatasourceForm({ dsType, editRow, onSaved, onCancel }) {
  const formRef = useRef(null)
  const [form, setForm] = useState(() => {
    if (editRow) {
      return {
        name: editRow.name || '',
        url: editRow.url || editRow.address || editRow.endpoint || '',
        auth: editRow.settings?.basic_auth_user ? 'basic' : editRow.settings?.bearer_token ? 'bearer' : 'none',
        username: editRow.settings?.basic_auth_user || '',
        password: editRow.settings?.basic_auth_password || '',
        token: editRow.settings?.bearer_token || '',
        skipTlsVerify: !!editRow.settings?.skip_tls_verify,
        caCert: editRow.settings?.ca_cert || '',
        serverName: editRow.settings?.server_name || '',
        clientCert: editRow.settings?.client_cert || '',
        clientKey: editRow.settings?.client_key || '',
        headers: Object.entries(editRow.settings?.custom_headers || {}).map(([key, value]) => ({ key, value })),
        scrapeInterval: editRow.settings?.scrape_interval || '15s',
        cluster: editRow.cluster_name || editRow.cluster || '',
        writeAddr: editRow.settings?.write_addr || '',
        tsdbType: editRow.settings?.['prometheus.tsdb_type'] || '',
        externalLabels: editRow.settings?.external_labels || '',
        isDefault: !!editRow.is_default,
        description: editRow.description || '',
        esVersion: editRow.settings?.version || '',
        esMaxShard: String(editRow.settings?.max_shard ?? '5'),
        esMinInterval: String(editRow.settings?.min_interval ?? '10'),
        esEnableWrite: !!editRow.settings?.enable_write,
        lokiMaxLines: String(editRow.settings?.max_lines ?? '1000'),
        jaegerVersion: editRow.settings?.version || 'v3',
      }
    }
    return { ...INITIAL_FORM }
  })
  const [errors, setErrors] = useState({})
  const [testing, setTesting] = useState(false)
  const [saving, setSaving] = useState(false)
  const [msg, setMsg] = useState({ type: '', text: '' })
  const [advancedOpen, setAdvancedOpen] = useState(false)

  const set = (key, val) => {
    setForm((prev) => ({ ...prev, [key]: val }))
    setErrors((prev) => { const next = { ...prev }; delete next[key]; return next })
  }

  const scrollToFirstError = (errs) => {
    const firstKey = Object.keys(errs)[0]
    if (!firstKey || !formRef.current) return
    const el = formRef.current.querySelector(`[data-field="${firstKey}"]`)
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }

  const runValidation = () => {
    const errs = validate(form, dsType)
    setErrors(errs)
    if (Object.keys(errs).length > 0) { scrollToFirstError(errs); return false }
    return true
  }

  const testConnection = async () => {
    if (!runValidation()) return false
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
    if (!runValidation()) return
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
    } else if (msg.text) {
      // 测试失败但用户可选择强制保存
    }
  }

  return (
    <section className='fx-ds-form' ref={formRef}>
      <header className='fx-ds-modal__head'>
        <h2>{editRow ? '编辑' : '新增'} {dsType.name} 数据源</h2>
        <button type='button' className='fx-ds-icon-button' onClick={onCancel} aria-label='返回'>×</button>
      </header>
      {!dsType.supported && (
        <div className='fx-ds-alert is-warning'>该数据源类型的查询功能尚在开发中</div>
      )}
      {msg.text && <div className={`fx-ds-alert ${msg.type === 'error' ? 'is-error' : 'is-success'}`}>{msg.text}</div>}
      <div className='fx-ds-form__grid'>
        <FieldLabel label='名称（必填）' name='name' errors={errors}>
          <input value={form.name} onChange={(e) => set('name', e.target.value)} placeholder='my-prometheus' />
        </FieldLabel>
        <FieldLabel label='URL（必填）' name='url' errors={errors}>
          <input value={form.url} onChange={(e) => set('url', e.target.value)} placeholder='http://localhost:9090' />
        </FieldLabel>
        <label>集群<input value={form.cluster} onChange={(e) => set('cluster', e.target.value)} placeholder='default' /></label>
        <label>认证方式
          <select value={form.auth} onChange={(e) => set('auth', e.target.value)}>
            {AUTH_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
        </label>
        {form.auth === 'basic' && <>
          <FieldLabel label='用户名' name='username' errors={errors}>
            <input value={form.username} onChange={(e) => set('username', e.target.value)} />
          </FieldLabel>
          <label>密码<input type='password' value={form.password} onChange={(e) => set('password', e.target.value)} /></label>
        </>}
        {form.auth === 'bearer' && (
          <label className='is-wide'>Bearer Token<input type='password' value={form.token} onChange={(e) => set('token', e.target.value)} /></label>
        )}
        {/* Skip TLS Verify */}
        <label className='is-wide'>
          <SwitchInput checked={form.skipTlsVerify} onChange={(v) => set('skipTlsVerify', v)} label='Skip TLS Verify' />
        </label>
        {/* 高级设置折叠 - mTLS */}
        <div className='is-wide' style={{ gridColumn: '1 / -1' }}>
          <div className='fx-ds-collapse-trigger' onClick={() => setAdvancedOpen(!advancedOpen)}>
            <span>{advancedOpen ? '▼' : '▶'}</span>
            <span>高级设置</span>
          </div>
          {advancedOpen && (
            <div className='fx-ds-collapse-content'>
              <label>CA 证书<textarea rows='7' value={form.caCert} onChange={(e) => set('caCert', e.target.value)} placeholder='Begins with -----BEGIN CERTIFICATE-----' /></label>
              <label>服务器名称<input value={form.serverName} onChange={(e) => set('serverName', e.target.value)} placeholder='domain.example.com' /></label>
              <label>客户端证书<textarea rows='7' value={form.clientCert} onChange={(e) => set('clientCert', e.target.value)} placeholder='Begins with -----BEGIN CERTIFICATE-----' /></label>
              <label>客户端密钥<textarea rows='7' value={form.clientKey} onChange={(e) => set('clientKey', e.target.value)} placeholder='Begins with -----BEGIN RSA PRIVATE KEY-----' /></label>
            </div>
          )}
        </div>
        <HeaderRows headers={form.headers} onChange={(h) => set('headers', h)} />
        {/* 类型专属配置 */}
        <TypeSpecificSettings dsType={dsType} form={form} set={set} />
      </div>
      <footer>
        <button type='button' onClick={onCancel}>取消</button>
        <button type='button' onClick={testConnection} disabled={testing}>{testing ? '测试中...' : '测试连接'}</button>
        <button type='button' onClick={saveAndTest} disabled={saving || testing}>保存并测试</button>
        <button type='button' className='is-primary' onClick={doSave} disabled={saving}>{saving ? '保存中...' : '强制保存'}</button>
      </footer>
    </section>
  )
}