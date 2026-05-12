import React, { useEffect, useState } from 'react'
import { formatPlatformError, platformApi } from '../api/platform.js'

function Field({ label, children }) {
  return <label className='fx-platform-field'><span>{label}</span>{children}</label>
}

export function SiteSettingsSection() {
  const [form, setForm] = useState({
    site_name: '',
    logo_url: '',
    login_announcement: '',
    session_timeout: 7200,
    site_url: '',
    home_page_url: '',
    business_group_display_mode: 'tree',
    team_display_mode: 'list',
  })
  const [saving, setSaving] = useState(false)
  const [feedback, setFeedback] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    platformApi.getSiteSettings().then((data) => {
      if (data && typeof data === 'object') {
        setForm((prev) => ({ ...prev, ...data }))
      }
    }).catch((err) => {
      setError(formatPlatformError(err))
    })
  }, [])

  const patch = (key, value) => setForm((prev) => ({ ...prev, [key]: value }))

  const save = async () => {
    setSaving(true)
    setFeedback('')
    setError('')
    try {
      await platformApi.saveSiteSettings(form)
      setFeedback('站点设置已保存。')
    } catch (err) {
      setError(formatPlatformError(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <section className='fx-platform-contract'>
      <header><h2>站点设置</h2></header>
      {error && <div className='fx-platform-error'>{error}</div>}
      {feedback && <div className='fx-platform-feedback'>{feedback}</div>}
      <div className='fx-platform-form'>
        <Field label='站点名称'>
          <input value={form.site_name} onChange={(e) => patch('site_name', e.target.value)} placeholder='FindX Monitoring' />
        </Field>
        <Field label='站点 URL'>
          <input value={form.site_url} onChange={(e) => patch('site_url', e.target.value)} placeholder='https://findx.example.com' />
        </Field>
        <Field label='Logo URL'>
          <input value={form.logo_url} onChange={(e) => patch('logo_url', e.target.value)} placeholder='https://cdn.example.com/logo.png' />
        </Field>
        <Field label='首页 URL'>
          <input value={form.home_page_url} onChange={(e) => patch('home_page_url', e.target.value)} placeholder='/dashboards' />
        </Field>
        <Field label='登录页公告'>
          <textarea value={form.login_announcement} onChange={(e) => patch('login_announcement', e.target.value)} rows={3} placeholder='登录页面公告内容' style={{ width: '100%', border: '1px solid #cdd8e8', borderRadius: 7, padding: '9px 10px', fontSize: 13 }} />
        </Field>
        <Field label='会话超时时间（秒）'>
          <input type='number' value={form.session_timeout} onChange={(e) => patch('session_timeout', Number(e.target.value) || 7200)} min={300} />
        </Field>
        <Field label='业务组展示模式'>
          <select value={form.business_group_display_mode} onChange={(e) => patch('business_group_display_mode', e.target.value)}>
            <option value='tree'>树形</option>
            <option value='list'>列表</option>
          </select>
        </Field>
        <Field label='团队展示模式'>
          <select value={form.team_display_mode} onChange={(e) => patch('team_display_mode', e.target.value)}>
            <option value='tree'>树形</option>
            <option value='list'>列表</option>
          </select>
        </Field>
        <footer>
          <button type='button' disabled={saving} onClick={save} className='fx-platform-btn-primary'>{saving ? '保存中...' : '保存'}</button>
        </footer>
      </div>
    </section>
  )
}
