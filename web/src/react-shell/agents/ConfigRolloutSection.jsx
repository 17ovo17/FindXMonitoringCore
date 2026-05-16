import React, { useEffect, useMemo, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { fmtTime } from './agentModel.js'
import { Blocked, Empty, ErrorBox, Field, Status, Tags } from './AgentShared.jsx'

const MODES = [
  { value: 'single', label: '单机' },
  { value: 'batch', label: '批量' },
  { value: 'canary', label: '灰度' },
]

export function ConfigRolloutSection({ agents }) {
  const [mode, setMode] = useState('single')
  const [targetId, setTargetId] = useState('')
  const [configJson, setConfigJson] = useState('{}')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [result, setResult] = useState(null)
  const [rolloutHistory, setRolloutHistory] = useState([])
  const [historyLoading, setHistoryLoading] = useState(false)
  const [driftItems, setDriftItems] = useState([])

  const agentList = useMemo(() => (agents || []).map(a => ({
    id: a.id || a.ident,
    label: a.hostname || a.ip || a.ident || a.id,
  })), [agents])

  const loadHistory = () => {
    setHistoryLoading(true)
    agentApi.listConfigRollouts()
      .then(setRolloutHistory)
      .catch(() => setRolloutHistory([]))
      .finally(() => setHistoryLoading(false))
  }

  useEffect(() => { loadHistory() }, [])

  const handleSubmit = (e) => {
    e.preventDefault()
    setError('')
    setResult(null)
    if (!targetId) { setError('请选择目标 Agent'); return }
    let config
    try { config = JSON.parse(configJson) } catch { setError('配置 JSON 格式错误'); return }
    setSubmitting(true)
    agentApi.configPushAgent(targetId, { config, mode })
      .then(res => { setResult(res); loadHistory() })
      .catch(err => setResult(err?.body || { status: 'PENDING', message: formatAgentError(err) }))
      .finally(() => setSubmitting(false))
  }

  return (
    <section className='fx-agent-work'>
      <div className='fx-agent-banner'>
        <div>
          <h3>配置下发</h3>
          <p>单机、批量和灰度配置只提交契约预检；没有执行器、回执和审计闭环前不显示真实下发。</p>
        </div>
      </div>

      <form className='fx-agent-panel' onSubmit={handleSubmit}>
        <h4>配置推送</h4>
        <ErrorBox>{error}</ErrorBox>
        <div className='fx-agent-filter fx-agent-filter-3'>
          <Field label='下发模式'>
            <select value={mode} onChange={e => setMode(e.target.value)}>
              {MODES.map(m => <option key={m.value} value={m.value}>{m.label}</option>)}
            </select>
          </Field>
          <Field label='目标 Agent'>
            <select value={targetId} onChange={e => setTargetId(e.target.value)}>
              <option value=''>请选择</option>
              {agentList.map(a => <option key={a.id} value={a.id}>{a.label}</option>)}
            </select>
          </Field>
        </div>
        <Field label='配置 JSON'>
          <textarea value={configJson} onChange={e => setConfigJson(e.target.value)} rows={6} style={{ width: '100%', fontFamily: 'monospace' }} />
        </Field>
        <div className='fx-agent-actions'>
          <button type='submit' disabled={submitting}>{submitting ? '预检中...' : '提交配置预检'}</button>
        </div>
        {result && (
          <div className='fx-agent-summary-row'>
            <Status ok={false}>{result.status || result.code || 'PENDING'}</Status>
            <Tags items={[result.message, `mode=${result.mode || mode}`, result.contract_id, ...(result.missing_contracts || [])]} />
          </div>
        )}
      </form>

      <DriftDetection agents={agentList} items={driftItems} onRefresh={() => setDriftItems([])} />
      <RollbackHistory records={rolloutHistory} loading={historyLoading} onRefresh={loadHistory} />
    </section>
  )
}

function DriftDetection({ agents, items, onRefresh }) {
  return (
    <div className='fx-agent-panel'>
      <div className='fx-agent-toolbar'>
        <h4>配置漂移检测</h4>
        <button type='button' onClick={onRefresh}>刷新</button>
      </div>
      {!items.length ? (
        <Empty>暂无漂移记录。配置漂移检测需要 Agent 上报当前配置版本后对比。</Empty>
      ) : (
        <div className='fx-agent-table'>
          <table>
            <thead><tr><th>Agent</th><th>期望版本</th><th>实际版本</th><th>状态</th></tr></thead>
            <tbody>{items.map((item, i) => (
              <tr key={i}>
                <td>{item.agent_id}</td>
                <td>{item.expected_version}</td>
                <td>{item.actual_version}</td>
                <td><Status ok={item.expected_version === item.actual_version}>{item.expected_version === item.actual_version ? '一致' : '漂移'}</Status></td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}
      <Blocked>配置漂移检测需要 Agent 定期上报 config_version，当前仅展示 UI 框架。</Blocked>
    </div>
  )
}

function RollbackHistory({ records, loading, onRefresh }) {
  const blocked = records.filter(r => r.status === 'blocked')
  return (
    <div className='fx-agent-panel'>
      <div className='fx-agent-toolbar'>
        <h4>下发/回滚历史</h4>
        <button type='button' onClick={onRefresh} disabled={loading}>{loading ? '刷新中...' : '刷新'}</button>
      </div>
      {!blocked.length ? (
        <Empty>暂无配置下发记录。</Empty>
      ) : (
        <div className='fx-agent-table'>
          <table>
            <thead><tr><th>模板</th><th>插件</th><th>策略</th><th>状态</th><th>阻断原因</th><th>时间</th></tr></thead>
            <tbody>{blocked.slice(0, 50).map(record => (
              <tr key={record.id || record.rollout_id}>
                <td>{record.template_id || '-'}</td>
                <td>{record.plugin_id || '-'}</td>
                <td>{record.rollout_strategy || record.rolloutStrategy || '-'}</td>
                <td><Status ok={false}>{record.status}</Status></td>
                <td><span className='fx-agent-muted'>{(record.blocker || '').slice(0, 60)}</span></td>
                <td>{fmtTime(record.updated_at || record.updatedAt || record.created_at)}</td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}
    </div>
  )
}
