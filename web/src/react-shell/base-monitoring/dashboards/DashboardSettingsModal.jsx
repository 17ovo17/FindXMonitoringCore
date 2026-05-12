import React, { useState } from 'react'
import { dashboardsApi } from '../../api/dashboards.js'
import { displayJson } from './dashboardModel.js'

function VariableEditor({ variables, onChange }) {
  const addVariable = () => {
    onChange([...variables, { name: '', label: '', type: 'custom', query: '', options: [], multi: false }])
  }
  const updateVariable = (index, field, value) => {
    const next = variables.map((v, i) => i === index ? { ...v, [field]: value } : v)
    onChange(next)
  }
  const removeVariable = (index) => {
    onChange(variables.filter((_, i) => i !== index))
  }

  return (
    <div className="fx-settings-section">
      <strong>变量配置</strong>
      {variables.map((v, i) => (
        <div key={i} className="fx-settings-var-row">
          <input value={v.name} onChange={(e) => updateVariable(i, 'name', e.target.value)} placeholder="name" />
          <input value={v.label} onChange={(e) => updateVariable(i, 'label', e.target.value)} placeholder="label" />
          <select value={v.type} onChange={(e) => updateVariable(i, 'type', e.target.value)}>
            <option value="custom">custom</option>
            <option value="query">query</option>
            <option value="textbox">textbox</option>
            <option value="constant">constant</option>
          </select>
          <input value={v.query || ''} onChange={(e) => updateVariable(i, 'query', e.target.value)} placeholder="query" />
          <label className="fx-settings-var-multi">
            <input type="checkbox" checked={v.multi || false} onChange={(e) => updateVariable(i, 'multi', e.target.checked)} />
            <span>多选</span>
          </label>
          <button type="button" onClick={() => removeVariable(i)}>x</button>
        </div>
      ))}
      <button type="button" className="fx-settings-add-btn" onClick={addVariable}>+ 添加变量</button>
    </div>
  )
}

export default function DashboardSettingsModal({ dashboard, onClose, onSaved }) {
  const [tab, setTab] = useState('basic')
  const [title, setTitle] = useState(dashboard.title || '')
  const [tags, setTags] = useState((dashboard.tags || []).join(', '))
  const [variables, setVariables] = useState(() => {
    const raw = dashboard.raw?.variables || dashboard.variables || {}
    if (Array.isArray(raw)) return raw
    return Object.entries(raw).map(([key, val]) => {
      if (typeof val === 'object' && val !== null) return { name: key, ...val }
      return { name: key, type: 'custom', query: '', options: [], multi: false, label: key }
    })
  })
  const [jsonText, setJsonText] = useState(() => displayJson(dashboard.raw || dashboard))
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const handleSave = async () => {
    setSaving(true)
    setError('')
    try {
      let body = {}
      if (tab === 'json') {
        body = JSON.parse(jsonText)
      } else {
        const tagList = tags.split(/[,，]/).map((t) => t.trim()).filter(Boolean)
        body = { ...dashboard.raw, title, tags: tagList, variables }
      }
      await dashboardsApi.update(dashboard.id, body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fx-dash-modal">
      <div className="fx-dash-modal__body fx-settings-modal">
        <header>
          <h2>仪表盘设置</h2>
          <button type="button" onClick={onClose}>x</button>
        </header>
        <nav className="fx-settings-tabs">
          <button type="button" className={tab === 'basic' ? 'is-active' : ''} onClick={() => setTab('basic')}>基本</button>
          <button type="button" className={tab === 'variables' ? 'is-active' : ''} onClick={() => setTab('variables')}>变量</button>
          <button type="button" className={tab === 'json' ? 'is-active' : ''} onClick={() => setTab('json')}>JSON</button>
        </nav>
        <div className="fx-settings-body">
          {tab === 'basic' && (
            <div className="fx-settings-section">
              <label className="fx-pe-field"><span>名称</span><input value={title} onChange={(e) => setTitle(e.target.value)} /></label>
              <label className="fx-pe-field"><span>标签（逗号分隔）</span><input value={tags} onChange={(e) => setTags(e.target.value)} /></label>
            </div>
          )}
          {tab === 'variables' && <VariableEditor variables={variables} onChange={setVariables} />}
          {tab === 'json' && (
            <div className="fx-settings-section">
              <textarea className="fx-settings-json" value={jsonText} onChange={(e) => setJsonText(e.target.value)} rows={16} />
            </div>
          )}
        </div>
        {error && <div className="fx-dash-alert is-error">{error}</div>}
        <footer className="fx-settings-footer">
          <button type="button" onClick={onClose}>取消</button>
          <button type="button" className="is-primary" disabled={saving} onClick={handleSave}>{saving ? '保存中...' : '保存'}</button>
        </footer>
      </div>
    </div>
  )
}
