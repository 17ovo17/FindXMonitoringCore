import React, { useEffect, useState } from 'react'
import { pluginApi } from '../api/plugins.js'
import { Blocked, ErrorBox } from './AgentShared.jsx'

export function EnvironmentAdaptSection({ agentId }) {
  const [env, setEnv] = useState(null)
  const [recommendations, setRecommendations] = useState([])
  const [selected, setSelected] = useState(new Set())
  const [loading, setLoading] = useState(false)
  const [adapting, setAdapting] = useState(false)
  const [error, setError] = useState('')
  const [feedback, setFeedback] = useState('')

  useEffect(() => {
    if (!agentId) return
    setLoading(true)
    pluginApi.getEnvironment(agentId)
      .then(setEnv)
      .catch(err => setError(err?.message || '获取环境信息失败'))
      .finally(() => setLoading(false))
  }, [agentId])

  const handleAutoAdapt = async () => {
    if (!agentId) return
    setAdapting(true)
    setError('')
    try {
      const res = await pluginApi.autoAdapt(agentId)
      const recs = res?.recommendations || []
      setRecommendations(recs)
      setSelected(new Set(recs.filter(r => r.suggested_on).map(r => r.plugin_id)))
    } catch (err) {
      setError(err?.message || '自动适配失败')
    } finally { setAdapting(false) }
  }

  const toggleSelect = (pluginId) => {
    setSelected(prev => {
      const next = new Set(prev)
      if (next.has(pluginId)) next.delete(pluginId)
      else next.add(pluginId)
      return next
    })
  }

  const handleApply = async () => {
    if (!agentId || selected.size === 0) return
    setError('')
    setFeedback('')
    try {
      const body = {
        agent_ids: [agentId],
        plugins: [...selected].map(id => ({ id, enabled: true, config: '' })),
        strategy: 'incremental',
      }
      const res = await pluginApi.configPush(body)
      setFeedback(`PENDING: 配置预检返回 ${res?.status || 'non-blocked'}；没有真实执行器、投递回执和效果回执前，不显示完成态。`)
    } catch (err) {
      setFeedback(`PENDING: ${err?.body?.message || err?.message || '配置预检被后端契约阻断'}`)
    }
  }

  if (!agentId) {
    return <section className='fx-agent-work'><div className='fx-agent-muted'>请先选择一个 Agent</div></section>
  }

  return (
    <section className='fx-agent-work'>
      <h3>环境探测与自动适配</h3>
      <ErrorBox>{error}</ErrorBox>
      {feedback && <Blocked>{feedback}</Blocked>}

      {loading && <div className='fx-agent-muted'>加载环境信息...</div>}

      {env && (
        <div className='fx-plugin-env-info'>
          <h4>主机环境</h4>
          <dl>
            <dt>操作系统</dt><dd>{env.os} / {env.arch}</dd>
            <dt>主机名</dt><dd>{env.hostname}</dd>
            <dt>内核版本</dt><dd>{env.kernel_version}</dd>
            <dt>CPU 核心</dt><dd>{env.cpu_cores}</dd>
            <dt>内存</dt><dd>{env.memory_mb} MB</dd>
            <dt>磁盘</dt><dd>{env.disk_gb} GB</dd>
          </dl>
          <h4>探测线索</h4>
          <div className='fx-plugin-service-tags'>
            {(env.detected_services || []).map(s => (
              <span key={s} className='fx-agent-tag'>{s}</span>
            ))}
          </div>
        </div>
      )}

      <div className='fx-plugin-adapt-actions'>
        <button type='button' onClick={handleAutoAdapt} disabled={adapting}>
          {adapting ? '分析中...' : '自动适配'}
        </button>
      </div>

      {recommendations.length > 0 && (
        <div className='fx-plugin-recommendations'>
          <h4>推荐插件</h4>
          <div className='fx-plugin-select-list'>
            {recommendations.map(rec => (
              <label key={rec.plugin_id} className='fx-plugin-select-item'>
                <input type='checkbox'
                  checked={selected.has(rec.plugin_id)}
                  onChange={() => toggleSelect(rec.plugin_id)} />
                <span>{rec.plugin_name}</span>
                <small>{rec.reason} (置信度: {rec.confidence}%)</small>
              </label>
            ))}
          </div>
          <button type='button' onClick={handleApply}
            disabled={selected.size === 0}>
            提交配置预检
          </button>
        </div>
      )}
    </section>
  )
}
