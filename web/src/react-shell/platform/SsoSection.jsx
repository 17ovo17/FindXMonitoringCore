import React, { useEffect, useState } from 'react'
import { formatPlatformError, platformApi } from '../api/platform.js'

const SSO_TABS = [
  { value: 'oauth2', label: 'OAuth2' },
  { value: 'ldap', label: 'LDAP' },
  { value: 'saml', label: 'SAML' },
]

const blankLdap = {
  host: '',
  port: 389,
  base_dn: '',
  bind_dn: '',
  bind_password: '',
  user_filter: '(uid=%s)',
  group_filter: '(memberUid=%s)',
  tls_enabled: false,
  tls_skip_verify: false,
}

const blankOauth2 = {
  client_id: '',
  client_secret: '',
  auth_url: '',
  token_url: '',
  userinfo_url: '',
  redirect_uri: '',
  scopes: 'openid profile email',
  default_role: 'viewer',
}

const blankSaml = {
  entity_id: '',
  sso_url: '',
  certificate: '',
  name_id_format: 'urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress',
  attribute_mapping: '{"email": "email", "name": "displayName"}',
}

function Field({ label, children }) {
  return <label className='fx-platform-field'><span>{label}</span>{children}</label>
}

export function SsoSection() {
  const [tab, setTab] = useState('oauth2')
  const [ldap, setLdap] = useState(blankLdap)
  const [oauth2, setOauth2] = useState(blankOauth2)
  const [saml, setSaml] = useState(blankSaml)
  const [feedback, setFeedback] = useState('')
  const [error, setError] = useState('')
  const [testing, setTesting] = useState(false)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    platformApi.getSsoConfig().then((data) => {
      if (!data) return
      const configs = Array.isArray(data) ? data : (data.rows || [data])
      const ldapRow = configs.find((r) => r.type === 'ldap')
      const oauth2Row = configs.find((r) => r.type === 'oauth2')
      const samlRow = configs.find((r) => r.type === 'saml')
      if (ldapRow?.config) setLdap((prev) => ({ ...prev, ...ldapRow.config }))
      if (oauth2Row?.config) setOauth2((prev) => ({ ...prev, ...oauth2Row.config }))
      if (samlRow?.config) setSaml((prev) => ({ ...prev, ...samlRow.config }))
    }).catch(() => {
      platformApi.sso({}).then((result) => {
        const rows = result.rows || []
        const ldapRow = rows.find((r) => r.type === 'ldap')
        const oauth2Row = rows.find((r) => r.type === 'oauth2')
        const samlRow = rows.find((r) => r.type === 'saml')
        if (ldapRow?.config) setLdap((prev) => ({ ...prev, ...ldapRow.config }))
        if (oauth2Row?.config) setOauth2((prev) => ({ ...prev, ...oauth2Row.config }))
        if (samlRow?.config) setSaml((prev) => ({ ...prev, ...samlRow.config }))
      }).catch(() => {})
    })
  }, [])

  const getConfigForTab = () => {
    if (tab === 'ldap') return ldap
    if (tab === 'oauth2') return oauth2
    return saml
  }

  const testConnection = async () => {
    setTesting(true)
    setFeedback('')
    setError('')
    try {
      const result = await platformApi.testSsoConnection({ type: tab, config: getConfigForTab() })
      setFeedback(result?.message || `${tab.toUpperCase()} 测试连接成功。`)
    } catch (err) {
      setError(formatPlatformError(err))
    } finally {
      setTesting(false)
    }
  }

  const saveConfig = async () => {
    setSaving(true)
    setFeedback('')
    setError('')
    try {
      await platformApi.saveSsoConfig({ type: tab, config: getConfigForTab() })
      setFeedback(`${tab.toUpperCase()} 配置已保存。`)
    } catch (err) {
      setError(formatPlatformError(err))
    } finally {
      setSaving(false)
    }
  }

  const patchLdap = (key, value) => setLdap((prev) => ({ ...prev, [key]: value }))
  const patchOauth2 = (key, value) => setOauth2((prev) => ({ ...prev, [key]: value }))
  const patchSaml = (key, value) => setSaml((prev) => ({ ...prev, [key]: value }))

  return (
    <section className='fx-platform-models'>
      <div className='fx-platform-toolbar'>
        {SSO_TABS.map((t) => (
          <button key={t.value} type='button' className={tab === t.value ? 'is-active' : ''} style={tab === t.value ? { background: '#1769ff', borderColor: '#1769ff', color: '#fff' } : {}} onClick={() => setTab(t.value)}>{t.label}</button>
        ))}
      </div>
      {error && <div className='fx-platform-error'>{error}</div>}
      {feedback && <div className='fx-platform-feedback'>{feedback}</div>}

      {tab === 'oauth2' && (
        <div className='fx-platform-form'>
          <Field label='Client ID'><input value={oauth2.client_id} onChange={(e) => patchOauth2('client_id', e.target.value)} /></Field>
          <Field label='Client Secret'><input type='password' value={oauth2.client_secret} onChange={(e) => patchOauth2('client_secret', e.target.value)} /></Field>
          <Field label='授权地址 (Auth URL)'><input value={oauth2.auth_url} onChange={(e) => patchOauth2('auth_url', e.target.value)} placeholder='https://provider.com/oauth2/authorize' /></Field>
          <Field label='Token 地址'><input value={oauth2.token_url} onChange={(e) => patchOauth2('token_url', e.target.value)} placeholder='https://provider.com/oauth2/token' /></Field>
          <Field label='用户信息地址 (Userinfo URL)'><input value={oauth2.userinfo_url} onChange={(e) => patchOauth2('userinfo_url', e.target.value)} placeholder='https://provider.com/oauth2/userinfo' /></Field>
          <Field label='回调地址 (Redirect URI)'><input value={oauth2.redirect_uri} onChange={(e) => patchOauth2('redirect_uri', e.target.value)} placeholder='https://findx.example.com/callback' /></Field>
          <Field label='Scopes'><input value={oauth2.scopes} onChange={(e) => patchOauth2('scopes', e.target.value)} placeholder='openid profile email' /></Field>
          <Field label='默认角色'><input value={oauth2.default_role} onChange={(e) => patchOauth2('default_role', e.target.value)} placeholder='viewer' /></Field>
          <footer className='fx-platform-form-footer'>
            <button type='button' disabled={testing} onClick={testConnection}>{testing ? '测试中...' : '测试连接'}</button>
            <button type='button' disabled={saving} onClick={saveConfig} className='fx-platform-btn-primary'>{saving ? '保存中...' : '保存配置'}</button>
          </footer>
        </div>
      )}

      {tab === 'ldap' && (
        <div className='fx-platform-form'>
          <Field label='主机地址 (Host)'><input value={ldap.host} onChange={(e) => patchLdap('host', e.target.value)} placeholder='ldap.example.com' /></Field>
          <Field label='端口 (Port)'><input type='number' value={ldap.port} onChange={(e) => patchLdap('port', Number(e.target.value) || 389)} /></Field>
          <Field label='Base DN'><input value={ldap.base_dn} onChange={(e) => patchLdap('base_dn', e.target.value)} placeholder='ou=users,dc=example,dc=com' /></Field>
          <Field label='Bind DN'><input value={ldap.bind_dn} onChange={(e) => patchLdap('bind_dn', e.target.value)} placeholder='cn=admin,dc=example,dc=com' /></Field>
          <Field label='Bind 密码'><input type='password' value={ldap.bind_password} onChange={(e) => patchLdap('bind_password', e.target.value)} /></Field>
          <Field label='用户过滤器 (User Filter)'><input value={ldap.user_filter} onChange={(e) => patchLdap('user_filter', e.target.value)} placeholder='(uid=%s)' /></Field>
          <Field label='组过滤器 (Group Filter)'><input value={ldap.group_filter} onChange={(e) => patchLdap('group_filter', e.target.value)} placeholder='(memberUid=%s)' /></Field>
          <Field label='启用 TLS'>
            <select value={ldap.tls_enabled ? 'true' : 'false'} onChange={(e) => patchLdap('tls_enabled', e.target.value === 'true')}>
              <option value='false'>关闭</option>
              <option value='true'>启用</option>
            </select>
          </Field>
          <Field label='跳过 TLS 验证'>
            <select value={ldap.tls_skip_verify ? 'true' : 'false'} onChange={(e) => patchLdap('tls_skip_verify', e.target.value === 'true')}>
              <option value='false'>否</option>
              <option value='true'>是</option>
            </select>
          </Field>
          <footer className='fx-platform-form-footer'>
            <button type='button' disabled={testing} onClick={testConnection}>{testing ? '测试中...' : '测试连接'}</button>
            <button type='button' disabled={saving} onClick={saveConfig} className='fx-platform-btn-primary'>{saving ? '保存中...' : '保存配置'}</button>
          </footer>
        </div>
      )}

      {tab === 'saml' && (
        <div className='fx-platform-form'>
          <Field label='Entity ID'><input value={saml.entity_id} onChange={(e) => patchSaml('entity_id', e.target.value)} placeholder='https://findx.example.com/saml/metadata' /></Field>
          <Field label='SSO 地址'><input value={saml.sso_url} onChange={(e) => patchSaml('sso_url', e.target.value)} placeholder='https://idp.example.com/saml/sso' /></Field>
          <Field label='证书 (PEM)'>
            <textarea value={saml.certificate} onChange={(e) => patchSaml('certificate', e.target.value)} rows={5} style={{ width: '100%', border: '1px solid #cdd8e8', borderRadius: 7, padding: '9px 10px', fontFamily: 'monospace', fontSize: 12 }} placeholder='-----BEGIN CERTIFICATE-----' />
          </Field>
          <Field label='NameID 格式'><input value={saml.name_id_format} onChange={(e) => patchSaml('name_id_format', e.target.value)} /></Field>
          <Field label='属性映射 (JSON)'>
            <textarea value={saml.attribute_mapping} onChange={(e) => patchSaml('attribute_mapping', e.target.value)} rows={3} style={{ width: '100%', border: '1px solid #cdd8e8', borderRadius: 7, padding: '9px 10px', fontFamily: 'monospace', fontSize: 12 }} placeholder='{"email": "email", "name": "displayName"}' />
          </Field>
          <footer className='fx-platform-form-footer'>
            <button type='button' disabled={testing} onClick={testConnection}>{testing ? '测试中...' : '测试连接'}</button>
            <button type='button' disabled={saving} onClick={saveConfig} className='fx-platform-btn-primary'>{saving ? '保存中...' : '保存配置'}</button>
          </footer>
        </div>
      )}
    </section>
  )
}
