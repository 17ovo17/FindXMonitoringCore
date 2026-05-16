import React, { useEffect, useState } from 'react'
import { pluginApi } from '../api/plugins.js'
import { agentApi } from '../api/agents.js'
import { ErrorBox } from './AgentShared.jsx'

const STEPS = ['选择目标 Agent', '选择插件和配置', '选择策略', '确认并执行']

export function ConfigPushWizard({ onClose }) {
  const [step, setStep] = useState(0)
  const [agents, setAgents] = useState([])
  const [plugins, setPlugins] = useState([])
  const [selectedAgents, setSelectedAgents] = useState([])
  const [selectedPlugins, setSelectedPlugins] = useState([])
  const [strategy, setStrategy] = useState('all')
  const [results, setResults] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      agentApi.list().catch(() => []),
      pluginApi.listPlugins().catch(() => []),
    ]).then(([a, p]) => { setAgents(a); setPlugins(p) })
  }, [])

  const toggleAgent = (id) => {
    setSelectedAgents(prev =>
      prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]
    )
  }

  const togglePlugin = (id) => {
    setSelectedPlugins(prev =>
      prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]
    )
  }

  const handleExecute = async () => {
    setLoading(true)
    setError('')
    try {
      const body = {
        agent_ids: selectedAgents,
        plugins: selectedPlugins.map(id => {
          const p = plugins.find(x => x.id === id)
          return { id, enabled: true, config: p?.default_config || '' }
        }),
        strategy,
      }
      const res = await pluginApi.configPush(body)
      setResults(res?.results || [])
      setStep(4)
    } catch (err) {
      setResults((err?.body?.results || []).length ? err.body.results : [{
        agent_id: selectedAgents.join(', '),
        plugin_id: selectedPlugins.join(', '),
        status: err?.body?.status || 'PENDING',
        message: err?.body?.message || err?.message || '配置预检被后端契约阻断',
      }])
      setStep(4)
    } finally { setLoading(false) }
  }

  const canNext = () => {
    if (step === 0) return selectedAgents.length > 0
    if (step === 1) return selectedPlugins.length > 0
    return true
  }

  return (
    <section className='fx-agent-work'>
      <div className='fx-plugin-wizard-head'>
        <h3>配置下发向导</h3>
        <button type='button' onClick={onClose}>关闭</button>
      </div>
      <nav className='fx-plugin-wizard-steps'>
        {STEPS.map((s, i) => (
          <span key={i} className={i === step ? 'is-active' : i < step ? 'is-done' : ''}>
            {i + 1}. {s}
          </span>
        ))}
      </nav>
      <ErrorBox>{error}</ErrorBox>

      {step === 0 && (
        <div className='fx-plugin-wizard-body'>
          <p>选择要下发配置的目标 Agent（可多选）：</p>
          <div className='fx-plugin-select-list'>
            {agents.map(a => (
              <label key={a.ident || a.id || a.ip} className='fx-plugin-select-item'>
                <input type='checkbox'
                  checked={selectedAgents.includes(a.ident || a.id || a.ip)}
                  onChange={() => toggleAgent(a.ident || a.id || a.ip)} />
                <span>{a.hostname || a.ident || a.ip}</span>
                <small>{a.ip} / {a.os}</small>
              </label>
            ))}
            {agents.length === 0 && <div className='fx-agent-muted'>暂无在线 Agent</div>}
          </div>
        </div>
      )}

      {step === 1 && (
        <div className='fx-plugin-wizard-body'>
          <p>选择要下发的插件（可多选）：</p>
          <div className='fx-plugin-select-list'>
            {plugins.map(p => (
              <label key={p.id} className='fx-plugin-select-item'>
                <input type='checkbox'
                  checked={selectedPlugins.includes(p.id)}
                  onChange={() => togglePlugin(p.id)} />
                <span>{p.name}</span>
                <small>{p.category} / {p.config_format}</small>
              </label>
            ))}
          </div>
        </div>
      )}

      {step === 2 && (
        <div className='fx-plugin-wizard-body'>
          <p>选择下发策略：</p>
          <label className='fx-plugin-select-item'>
            <input type='radio' name='strategy' value='all'
              checked={strategy === 'all'} onChange={() => setStrategy('all')} />
            <span>全量下发</span>
            <small>覆盖目标 Agent 上的所有插件配置</small>
          </label>
          <label className='fx-plugin-select-item'>
            <input type='radio' name='strategy' value='incremental'
              checked={strategy === 'incremental'}
              onChange={() => setStrategy('incremental')} />
            <span>增量下发</span>
            <small>仅更新选中的插件配置，保留其他配置不变</small>
          </label>
        </div>
      )}

      {step === 3 && (
        <div className='fx-plugin-wizard-body'>
          <h4>确认下发信息</h4>
          <dl>
            <dt>目标 Agent</dt>
            <dd>{selectedAgents.join(', ')}</dd>
            <dt>下发插件</dt>
            <dd>{selectedPlugins.join(', ')}</dd>
            <dt>策略</dt>
            <dd>{strategy === 'all' ? '全量下发' : '增量下发'}</dd>
          </dl>
        </div>
      )}

      {step === 4 && results && (
        <div className='fx-plugin-wizard-body'>
          <h4>契约预检结果</h4>
          <table className='fx-plugin-result-table'>
            <thead><tr><th>Agent</th><th>插件</th><th>状态</th><th>信息</th></tr></thead>
            <tbody>
              {results.map((r, i) => (
                <tr key={i}>
                  <td>{r.agent_id}</td>
                  <td>{r.plugin_id}</td>
                  <td className='is-fail'>{r.status}</td>
                  <td>{r.message || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {step < 4 && (
        <div className='fx-plugin-wizard-nav'>
          {step > 0 && <button type='button' onClick={() => setStep(s => s - 1)}>上一步</button>}
          {step < 3 && (
            <button type='button' disabled={!canNext()} onClick={() => setStep(s => s + 1)}>
              下一步
            </button>
          )}
          {step === 3 && (
            <button type='button' disabled={loading} onClick={handleExecute}>
              {loading ? '预检中...' : '确认预检'}
            </button>
          )}
        </div>
      )}
    </section>
  )
}
