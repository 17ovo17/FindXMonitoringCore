import React, { useEffect, useState } from 'react'
import { pluginApi } from '../api/plugins.js'
import { ErrorBox } from './AgentShared.jsx'

const TABS = [
  { key: 'collect', label: '采集插件' },
  { key: 'diagnose', label: '诊断插件' },
  { key: 'apm', label: 'APM 探针' },
]

const CATEGORY_ICONS = {
  collect: '\u{1F4CA}',
  diagnose: '\u{1F527}',
  apm: '\u{1F50D}',
}

export function PluginCatalogSection({ agentId, onEditPlugin }) {
  const [plugins, setPlugins] = useState([])
  const [tab, setTab] = useState('collect')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    setLoading(true)
    pluginApi.listPlugins()
      .then(setPlugins)
      .catch(err => setError(err?.message || '加载插件列表失败'))
      .finally(() => setLoading(false))
  }, [])

  const filtered = plugins.filter(p => p.category === tab)

  const handleToggle = async (plugin) => {
    if (!agentId) return
    try {
      await pluginApi.patchPlugin(agentId, plugin.id, !plugin.enabled)
      setPlugins(prev => prev.map(p =>
        p.id === plugin.id ? { ...p, enabled: !p.enabled } : p
      ))
    } catch (err) {
      setError(err?.message || '操作失败')
    }
  }

  return (
    <section className='fx-agent-work'>
      <nav className='fx-agent-tabs'>
        {TABS.map(t => (
          <button key={t.key} type='button'
            className={tab === t.key ? 'is-active' : ''}
            onClick={() => setTab(t.key)}>
            {t.label}
          </button>
        ))}
      </nav>
      <ErrorBox>{error}</ErrorBox>
      {loading && <div className='fx-agent-muted'>加载中...</div>}
      <div className='fx-plugin-grid'>
        {filtered.map(plugin => (
          <article key={plugin.id} className='fx-plugin-card'>
            <div className='fx-plugin-card-head'>
              <span className='fx-plugin-icon'>
                {CATEGORY_ICONS[plugin.category] || ''}
              </span>
              <strong>{plugin.name}</strong>
            </div>
            <p className='fx-plugin-desc'>{plugin.description}</p>
            <div className='fx-plugin-card-foot'>
              <span className='fx-plugin-os'>
                {(plugin.supported_os || []).join(' / ')}
              </span>
              <div className='fx-plugin-actions'>
                {agentId && (
                  <label className='fx-plugin-toggle'>
                    <input type='checkbox' checked={plugin.enabled}
                      onChange={() => handleToggle(plugin)} />
                    <span>{plugin.enabled ? '已启用' : '已停用'}</span>
                  </label>
                )}
                <button type='button' className='fx-plugin-edit-btn'
                  onClick={() => onEditPlugin && onEditPlugin(plugin)}>
                  配置
                </button>
              </div>
            </div>
          </article>
        ))}
        {!loading && filtered.length === 0 && (
          <div className='fx-agent-empty'>暂无插件</div>
        )}
      </div>
    </section>
  )
}
