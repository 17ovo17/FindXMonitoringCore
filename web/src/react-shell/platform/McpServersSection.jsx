import React, { useEffect, useState } from 'react'
import { get, post, put, del } from '../api/http.js'

const MCP_TYPES = [
  { value: 'nightingale', label: '夜莺监控' },
  { value: 'cmdb', label: 'CMDB' },
  { value: 'agent', label: 'Agent' },
  { value: 'knowledge', label: '知识库' },
  { value: 'prometheus', label: 'Prometheus' },
  { value: 'alertmanager', label: 'AlertManager' },
]

const STATUS_BADGE = { online: '\u{1F7E2}', offline: '\u{1F534}', error: '⚠️' }
const STATUS_LABEL = { online: '在线', offline: '离线', error: '异常' }

function McpModal({ mode, item, onClose, onSave }) {
  const [draft, setDraft] = useState(item || { name: '', type: 'nightingale', endpoint: '', description: '' })
  const [saving, setSaving] = useState(false)
  const patch = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }))

  const handleSave = async () => {
    setSaving(true)
    await onSave(draft)
    setSaving(false)
  }

  return (
    <div className='fx-platform-modal'>
      <div className='fx-platform-modal__body'>
        <header>
          <h2>{mode === 'edit' ? '编辑 MCP Server' : '注册 MCP Server'}</h2>
          <button type='button' onClick={onClose}>关闭</button>
        </header>
        <div className='fx-platform-form'>
          <label className='fx-platform-field'><span>名称</span>
            <input value={draft.name || ''} onChange={(e) => patch('name', e.target.value)} placeholder='MCP 服务名称' />
          </label>
          <label className='fx-platform-field'><span>类型</span>
            <select value={draft.type || 'nightingale'} onChange={(e) => patch('type', e.target.value)}>
              {MCP_TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
            </select>
          </label>
          <label className='fx-platform-field'><span>Endpoint</span>
            <input value={draft.endpoint || ''} onChange={(e) => patch('endpoint', e.target.value)} placeholder='http://localhost:8080/mcp/...' />
          </label>
          <label className='fx-platform-field'><span>描述</span>
            <input value={draft.description || ''} onChange={(e) => patch('description', e.target.value)} placeholder='服务描述' />
          </label>
          <footer>
            <button type='button' disabled={saving} onClick={handleSave}>{saving ? '保存中...' : '保存'}</button>
          </footer>
        </div>
      </div>
    </div>
  )
}

export function McpServersSection() {
  const [servers, setServers] = useState([])
  const [error, setError] = useState('')
  const [modal, setModal] = useState(null)
  const [healthResults, setHealthResults] = useState({})

  const load = async () => {
    setError('')
    try {
      const result = await get('/mcp/servers')
      setServers(result.data || result || [])
    } catch (err) {
      setError(err?.message || '加载 MCP 服务列表失败')
    }
  }

  useEffect(() => { load() }, [])

  const handleSave = async (draft) => {
    setError('')
    try {
      if (modal.mode === 'edit') {
        await put(`/mcp/servers/${draft.id}`, draft)
      } else {
        await post('/mcp/servers', draft)
      }
      setModal(null)
      await load()
    } catch (err) {
      setError(err?.message || '保存失败')
    }
  }

  const handleDelete = async (server) => {
    if (!window.confirm(`确认删除 MCP 服务 "${server.name}"？`)) return
    setError('')
    try {
      await del(`/mcp/servers/${server.id}`)
      await load()
    } catch (err) {
      setError(err?.message || '删除失败')
    }
  }

  const handleHealthCheck = async (server) => {
    try {
      const result = await post(`/mcp/servers/${server.id}/health-check`)
      setHealthResults((prev) => ({ ...prev, [server.id]: result }))
      await load()
    } catch (err) {
      setHealthResults((prev) => ({ ...prev, [server.id]: { status: 'error', error: err?.message } }))
    }
  }

  const typeIcon = (type) => {
    const icons = { nightingale: '\u{1F426}', cmdb: '\u{1F4E6}', agent: '\u{1F916}', knowledge: '\u{1F4DA}', prometheus: '\u{1F525}', alertmanager: '\u{1F514}' }
    return icons[type] || '\u{1F50C}'
  }

  return (
    <section className='fx-platform-models'>
      <div className='fx-platform-toolbar'>
        <button type='button' onClick={() => setModal({ mode: 'create', item: { name: '', type: 'nightingale', endpoint: '', description: '' } })}>注册 MCP Server</button>
        <button type='button' onClick={load}>刷新</button>
      </div>
      {error && <div className='fx-platform-error'>{error}</div>}
      <div className='fx-platform-provider-grid'>
        {servers.map((server) => (
          <article key={server.id} className='fx-platform-card'>
            <header>
              <h3>{typeIcon(server.type)} {server.name}</h3>
              <span className={server.status === 'online' ? 'is-ok' : 'is-unknown'}>
                {STATUS_BADGE[server.status] || STATUS_BADGE.offline} {STATUS_LABEL[server.status] || '离线'}
              </span>
            </header>
            <p>{server.endpoint || '未配置 Endpoint'}</p>
            <p>{server.description || '-'}</p>
            {healthResults[server.id] && (
              <p style={{ fontSize: '12px', color: healthResults[server.id].status === 'online' ? '#16a34a' : '#dc2626' }}>
                {healthResults[server.id].status === 'online' ? `健康 (${healthResults[server.id].latency}ms)` : `异常: ${healthResults[server.id].error || 'status ' + healthResults[server.id].status_code}`}
              </p>
            )}
            <footer>
              <button type='button' onClick={() => handleHealthCheck(server)}>健康检查</button>
              <button type='button' onClick={() => setModal({ mode: 'edit', item: server })}>编辑</button>
              <button type='button' className='danger' onClick={() => handleDelete(server)}>删除</button>
            </footer>
          </article>
        ))}
        {!servers.length && <div className='fx-platform-empty'>暂无 MCP 服务注册。</div>}
      </div>
      {modal && <McpModal mode={modal.mode} item={modal.item} onClose={() => setModal(null)} onSave={handleSave} />}
    </section>
  )
}
