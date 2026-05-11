import React, { useMemo, useState, useCallback } from 'react'
import { ASSET_BLOCKERS } from '../api/assets.js'
import { post } from '../api/http.js'
import { agentKey, agentOnline, displayText, fmtTime } from './assetsModel.js'
import { Blocked, ErrorBox, filterRows, Status, Tags } from './Shared.jsx'

const VERDICT_LABEL = { online: '✅ 已验证在线', degraded: '⚠️ 降级', offline: '❌ 离线' }
const VERDICT_CLASS = { online: 'is-success', degraded: 'is-warning', offline: 'is-danger' }

export function AgentsSection({ rows, hosts, error, q, onRefresh }) {
  const [selected, setSelected] = useState([])
  const [blocked, setBlocked] = useState('')
  const [verifying, setVerifying] = useState({})
  const [verifyResults, setVerifyResults] = useState({})
  const filtered = useMemo(() => filterRows(rows, q), [rows, q])
  const toggle = id => setSelected(prev => prev.includes(id) ? prev.filter(item => item !== id) : [...prev, id])

  const verifySingle = useCallback(async (agentId) => {
    setVerifying(prev => ({ ...prev, [agentId]: true }))
    try {
      const result = await post(`/findx-agents/${encodeURIComponent(agentId)}/verify-heartbeat`)
      setVerifyResults(prev => ({ ...prev, [agentId]: result }))
    } catch (err) {
      setVerifyResults(prev => ({ ...prev, [agentId]: { verdict: 'offline', error: err.message } }))
    } finally {
      setVerifying(prev => ({ ...prev, [agentId]: false }))
    }
  }, [])

  const batchVerify = useCallback(async () => {
    const ids = selected.length ? selected : filtered.map(row => row.id)
    if (!ids.length) return
    const batchIds = ids.slice(0, 50)
    batchIds.forEach(id => setVerifying(prev => ({ ...prev, [id]: true })))
    try {
      const resp = await post('/findx-agents/batch-verify', { agent_ids: batchIds })
      const results = resp?.results || []
      const map = {}
      results.forEach(r => { map[r.agent_id] = r })
      setVerifyResults(prev => ({ ...prev, ...map }))
    } catch (err) {
      batchIds.forEach(id => setVerifyResults(prev => ({ ...prev, [id]: { verdict: 'offline', error: err.message } })))
    } finally {
      batchIds.forEach(id => setVerifying(prev => ({ ...prev, [id]: false })))
    }
  }, [selected, filtered])

  return (
    <section className='fx-assets-work'>
      <div className='fx-assets-toolbar'>
        <button type='button' onClick={() => setBlocked(ASSET_BLOCKERS.agentLifecycle)}>部署 FindX Agent</button>
        <button type='button' disabled={!selected.length} onClick={() => setBlocked(ASSET_BLOCKERS.agentLifecycle)}>批量卸载({selected.length})</button>
        <button type='button' onClick={batchVerify}>批量验证({selected.length || filtered.length})</button>
        <button type='button' onClick={onRefresh}>刷新</button>
      </div>
      <div className='fx-assets-grid'>
        <article className='fx-assets-card'><strong>{filtered.length}</strong><span>Agent 总数</span></article>
        <article className='fx-assets-card'><strong>{filtered.filter(agentOnline).length}</strong><span>在线 Agent</span></article>
        <article className='fx-assets-card'><strong>{hosts.length}</strong><span>关联主机</span></article>
        <article className='fx-assets-card'><strong>{new Set(filtered.map(row => row.version).filter(Boolean)).size}</strong><span>版本数</span></article>
      </div>
      <ErrorBox>{error}</ErrorBox>{blocked && <Blocked>{blocked}</Blocked>}<Blocked>{ASSET_BLOCKERS.agentPackage}</Blocked>
      <div className='fx-assets-table'>
        <table><thead><tr><th></th><th>主机</th><th>IP</th><th>版本</th><th>状态</th><th>验证结果</th><th>能力</th><th>配置版本</th><th>最后心跳</th><th>操作</th></tr></thead>
          <tbody>{filtered.map(row => {
            const key = agentKey(row)
            const vr = verifyResults[row.id]
            const isVerifying = verifying[row.id]
            return <tr key={key}><td><input type='checkbox' checked={selected.includes(key)} onChange={() => toggle(key)} /></td><td>{displayText(row.hostname || row.ident)}</td><td>{displayText(row.ip)}</td><td>{displayText(row.version)}</td><td><Status ok={agentOnline(row)}>{agentOnline(row) ? '在线' : '离线'}</Status></td><td>{isVerifying ? <span className='fx-verify-loading'>验证中...</span> : vr ? <span className={`fx-verify-badge ${VERDICT_CLASS[vr.verdict] || ''}`} title={vr.verified_at ? `验证时间: ${fmtTime(vr.verified_at)}` : ''}>{VERDICT_LABEL[vr.verdict] || vr.verdict}</span> : <span className='fx-verify-none'>未验证</span>}</td><td><Tags items={row.capabilities} /></td><td>{displayText(row.config_version)}</td><td>{fmtTime(row.last_seen)}</td><td className='fx-assets-actions'><button type='button' onClick={() => verifySingle(row.id)} disabled={isVerifying}>验证心跳</button><button type='button' onClick={() => setBlocked(ASSET_BLOCKERS.agentLifecycle)}>重启</button><button type='button' onClick={() => setBlocked(ASSET_BLOCKERS.agentLifecycle)}>重新部署</button><button type='button' className='is-danger' onClick={() => setBlocked(ASSET_BLOCKERS.agentLifecycle)}>卸载</button></td></tr>
          })}</tbody>
        </table>
        {!filtered.length && <div className='fx-assets-empty'>暂无 FindX Agent 心跳记录。</div>}
      </div>
    </section>
  )
}
