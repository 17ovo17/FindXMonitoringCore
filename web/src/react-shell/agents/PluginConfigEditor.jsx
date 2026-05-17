import React, { useEffect, useState, useRef } from 'react'
import { pluginApi } from '../api/plugins.js'
import { ErrorBox } from './AgentShared.jsx'

function LazyMonacoEditor({ language, value, onChange, options }) {
  const [Editor, setEditor] = useState(null)

  useEffect(() => {
    import('@monaco-editor/react').then((mod) => {
      setEditor(() => mod.default)
    }).catch(() => { /* 降级 */ })
  }, [])

  if (!Editor) {
    return (
      <textarea rows={20} value={value || ''}
        onChange={(e) => onChange(e.target.value)}
        style={{ width: '100%', height: '400px', fontFamily: 'monospace', fontSize: 13, padding: 12, border: '1px solid #e0e0e0', resize: 'vertical', borderRadius: 4 }}
      />
    )
  }

  return (
    <Editor height='400px' language={language}
      value={value || ''} onChange={(val) => onChange(val || '')}
      options={options} />
  )
}

export function PluginConfigEditor({ plugin, agentId, onBack }) {
  const [config, setConfig] = useState(plugin?.default_config || '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [feedback, setFeedback] = useState('')
  const timerRef = useRef(null)

  useEffect(() => {
    if (!agentId || !plugin) return
    pluginApi.getConfig(agentId).then(data => {
      const found = (data?.plugins || []).find(p => p.plugin_id === plugin.id)
      if (found?.config) setConfig(found.config)
    }).catch(() => { /* 使用默认配置 */ })
  }, [agentId, plugin])

  const showFeedback = (msg) => {
    setFeedback(msg)
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => setFeedback(''), 3000)
  }

  const handleSave = async () => {
    if (!agentId) { setError('请先选择 Agent'); return }
    setSaving(true)
    setError('')
    try {
      await pluginApi.updatePluginConfig(agentId, plugin.id, config)
      showFeedback('配置已保存为控制面记录；真实下发仍需 config-rollouts 回执契约。')
    } catch (err) {
      setError(err?.message || '保存失败')
    } finally { setSaving(false) }
  }

  const handlePush = async () => {
    if (!agentId) { setError('请先选择 Agent'); return }
    setSaving(true)
    setError('')
    try {
      const res = await pluginApi.configPush({
        agent_ids: [agentId],
        plugins: [{ id: plugin.id, enabled: true, config }],
        strategy: 'all',
      })
      showFeedback(`配置预检返回 ${res?.status || 'non-blocked'}；不显示完成态。`)
    } catch (err) {
      showFeedback(`${err?.body?.message || err?.message || '配置预检被后端契约阻断'}`)
    } finally { setSaving(false) }
  }

  const language = plugin?.config_format === 'yaml' ? 'yaml' : 'plaintext'

  return (
    <section className='fx-agent-work'>
      <div className='fx-plugin-editor-head'>
        <button type='button' onClick={onBack}>返回</button>
        <h3>{plugin?.name} - 配置编辑</h3>
      </div>
      <ErrorBox>{error}</ErrorBox>
      {feedback && <div className='fx-agent-blocked'>{feedback}</div>}
      <div className='fx-plugin-editor-layout'>
        <div className='fx-plugin-editor-main'>
          <LazyMonacoEditor language={language} value={config}
            onChange={setConfig}
            options={{ minimap: { enabled: false }, lineNumbers: 'on', scrollBeyondLastLine: false, wordWrap: 'on', tabSize: 2 }}
          />
        </div>
        <aside className='fx-plugin-editor-side'>
          <h4>插件信息</h4>
          <dl>
            <dt>ID</dt><dd>{plugin?.id}</dd>
            <dt>分类</dt><dd>{plugin?.category}</dd>
            <dt>格式</dt><dd>{plugin?.config_format}</dd>
            <dt>支持系统</dt>
            <dd>{(plugin?.supported_os || []).join(', ')}</dd>
          </dl>
          <h4>说明</h4>
          <p className='fx-plugin-desc'>{plugin?.description}</p>
        </aside>
      </div>
      <div className='fx-plugin-editor-actions'>
        <button type='button' onClick={handleSave} disabled={saving}>
          {saving ? '保存中...' : '保存配置'}
        </button>
        <button type='button' onClick={handlePush} disabled={saving}>
          下发到 Agent
        </button>
      </div>
    </section>
  )
}
