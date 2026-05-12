import React from 'react'
import {
  displayAddress,
  displayCluster,
  displayName,
  displayType,
  safeDisplayJson,
  safeDisplayText,
  statusText,
} from './datasourceModel.js'

const SENSITIVE_KEYS = ['password', 'passwd', 'token', 'secret', 'bearer_token', 'basic_auth_password', 'api_key']

function maskValue(key, value) {
  const lower = String(key).toLowerCase()
  if (SENSITIVE_KEYS.some((s) => lower.includes(s))) return '••••••'
  return value
}

function DetailField({ label, value }) {
  return (
    <div>
      <dt>{label}</dt>
      <dd>{value || '-'}</dd>
    </div>
  )
}

export function DatasourceDrawer({ row, onClose, onEdit }) {
  if (!row) return null

  const settings = row.settings || {}
  const headers = settings.custom_headers || {}

  return (
    <aside className='fx-ds-drawer' aria-label='数据源详情'>
      <header>
        <h2>数据源详情</h2>
        <button type='button' className='fx-ds-icon-button' onClick={onClose} aria-label='关闭'>×</button>
      </header>
      <dl className='fx-ds-detail-list'>
        <DetailField label='名称' value={safeDisplayText(displayName(row))} />
        <DetailField label='类型' value={safeDisplayText(displayType(row))} />
        <DetailField label='集群' value={safeDisplayText(displayCluster(row)) || 'default'} />
        <DetailField label='状态' value={statusText(row)} />
        <DetailField label='地址' value={safeDisplayText(displayAddress(row))} />
        {settings.basic_auth_user && <DetailField label='用户名' value={settings.basic_auth_user} />}
        {settings.basic_auth_password && <DetailField label='密码' value={maskValue('password', settings.basic_auth_password)} />}
        {settings.bearer_token && <DetailField label='Token' value={maskValue('token', settings.bearer_token)} />}
        {settings.tls_enabled && <DetailField label='TLS' value='已启用' />}
        {settings.scrape_interval && <DetailField label='Scrape Interval' value={settings.scrape_interval} />}
        {Object.keys(headers).length > 0 && (
          <DetailField label='自定义 Header' value={Object.entries(headers).map(([k, v]) => `${k}: ${v}`).join(', ')} />
        )}
      </dl>
      <div className='fx-ds-drawer__actions'>
        <button type='button' className='is-primary' onClick={() => onEdit(row)}>编辑</button>
        <button type='button' onClick={onClose}>关闭</button>
      </div>
      <pre className='fx-ds-json'>{safeDisplayJson(row, 10000)}</pre>
    </aside>
  )
}