import React, { useEffect, useMemo, useState } from 'react'
import { formatPlatformError, platformApi, PLATFORM_BLOCKERS, safePlatformJson } from '../api/platform.js'
import { McpServersSection } from './McpServersSection.jsx'
import { SsoSection } from './SsoSection.jsx'
import { useConfirm } from '../shared/ConfirmModal.jsx'
import './platform.css'

const sections = [
  { value: 'models', label: 'AI 模型配置', desc: '统一管理大模型、Embedding、Reranker 与健康检测。' },
  { value: 'mcp', label: 'MCP 服务', desc: 'MCP Server 注册、健康检查与状态管理。' },
  { value: 'site', label: '站点设置', desc: '站点 URL、首页和展示模式配置。' },
  { value: 'variables', label: '变量设置', desc: '统一变量、密文变量和引用影响分析。' },
  { value: 'sso', label: '单点登录', desc: '登录协议、默认角色和团队映射。' },
  { value: 'alerting-engines', label: '告警引擎', desc: '告警执行节点和心跳状态。' },
  { value: 'health', label: '运行自检', desc: '平台依赖、存储与配置链路状态。' },
  { value: 'audit', label: '审计日志', desc: '敏感操作、权限和配置变更留痕。' },
]

const sectionSet = new Set(sections.map((item) => item.value))
const emptyProvider = { id: '', name: '', base_url: '', api_key: '', models: '', default: false }
const sanitizeEndpoint = (value) => String(value || '').replace(/\$\{[^}]+\}/g, '<FINDX_URL>').replace(/https?:\/\/[^\s]+/g, '<URL>')

function Blocked({ children }) {
  return <div className='fx-platform-blocked'><strong>BLOCKED_BY_CONTRACT</strong><span>{children}</span></div>
}

function Modal({ title, children, onClose }) {
  return (
    <div className='fx-platform-modal'>
      <div className='fx-platform-modal__body'>
        <header><h2>{title}</h2><button type='button' onClick={onClose}>关闭</button></header>
        {children}
      </div>
    </div>
  )
}

function Field({ label, children }) {
  return <label className='fx-platform-field'><span>{label}</span>{children}</label>
}

const maskProvider = (item) => ({
  ...item,
  base_url: sanitizeEndpoint(item.base_url),
  models: Array.isArray(item.models) ? item.models : [],
  api_key: item.api_key ? '<SECRET>' : '',
})

const providerPayload = (providers) => providers.map((item) => ({
  id: item.id || `${Date.now()}`,
  name: item.name,
  base_url: item.base_url,
  api_key: item.api_key,
  models: String(Array.isArray(item.models) ? item.models.join(',') : item.models || '').split(',').map((model) => model.trim()).filter(Boolean),
  default: !!item.default,
}))

function ModelsSection() {
  const [providers, setProviders] = useState([])
  const [source, setSource] = useState('')
  const [error, setError] = useState('')
  const [modal, setModal] = useState(null)
  const [saving, setSaving] = useState(false)
  const [embedding, setEmbedding] = useState({ provider: 'builtin', api_url: '', api_key: '', model: '', dimensions: 1536 })
  const [reranker, setReranker] = useState({ enabled: false, provider: 'llm', api_url: '', api_key: '', model: '', top_k: 5 })
  const [feedback, setFeedback] = useState('')
  const { confirm, modal: confirmModal } = useConfirm()

  const load = async () => {
    setError('')
    try {
      const result = await platformApi.listProviders()
      setProviders((result.rows || []).map(maskProvider))
      setSource(result.source)
      setEmbedding(await platformApi.getEmbedding().catch(() => ({ provider: 'builtin' })))
      setReranker(await platformApi.getReranker().catch(() => ({ enabled: false, provider: 'llm', top_k: 5 })))
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }

  useEffect(() => {
    load()
  }, [])

  const saveProviders = async (nextProviders) => {
    if (nextProviders.some((item) => item.api_key === '<SECRET>')) {
      setError('BLOCKED_BY_CONTRACT: 当前批量保存契约无法保留既有密钥引用。请为要保存的条目重新输入密钥，或等待凭据引用契约。')
      return
    }
    setSaving(true)
    setError('')
    try {
      await platformApi.saveProviders(providerPayload(nextProviders))
      setFeedback('AI 服务配置已保存。')
      await load()
    } catch (err) {
      setError(formatPlatformError(err))
    } finally {
      setSaving(false)
    }
  }

  const saveProviderModal = async (draft) => {
    const next = modal.mode === 'edit'
      ? providers.map((item) => item.id === modal.item.id ? draft : item)
      : [...providers, { ...draft, id: draft.id || `${Date.now()}` }]
    await saveProviders(next)
    if (!error) setModal(null)
  }

  const removeProvider = async (provider) => {
    const ok = await confirm({ title: '删除 AI 服务', message: `确认删除 AI 服务 ${provider.name}？`, confirmText: '删除', danger: true })
    if (!ok) return
    await saveProviders(providers.filter((item) => item.id !== provider.id))
  }

  const setDefault = async (provider) => {
    await saveProviders(providers.map((item) => ({ ...item, default: item.id === provider.id })))
  }

  const saveEmbedding = async () => {
    setFeedback('')
    try {
      await platformApi.saveEmbedding(embedding)
      setFeedback('Embedding 配置已保存。')
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }

  const testEmbedding = async () => {
    setFeedback('')
    try {
      const result = await platformApi.testEmbedding({ api_url: embedding.api_url, api_key: embedding.api_key, model: embedding.model })
      setFeedback(result?.message || 'Embedding 测试通过。')
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }

  const saveReranker = async () => {
    setFeedback('')
    try {
      await platformApi.saveReranker(reranker)
      setFeedback('Reranker 配置已保存。')
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }

  return (
    <section className='fx-platform-models'>
      <div className='fx-platform-toolbar'><button type='button' onClick={() => setModal({ mode: 'create', item: emptyProvider })}>新增 AI 服务</button><button type='button' onClick={load}>刷新</button><span>{source}</span></div>
      <Blocked>{PLATFORM_BLOCKERS.llmCrud}</Blocked>
      {error && <div className='fx-platform-error'>{error}</div>}
      {feedback && <div className='fx-platform-feedback'>{feedback}</div>}
      <div className='fx-platform-provider-grid'>
        {providers.map((provider) => (
          <article key={provider.id} className='fx-platform-card'>
            <header><h3>{provider.name || '未命名服务'}</h3><span className={provider.health?.alive ? 'is-ok' : 'is-unknown'}>{provider.health ? (provider.health.alive ? '健康' : '异常') : '待检测'}</span></header>
            <p>{provider.base_url || '未配置 API 地址'}</p>
            <p>模型：{(provider.models || []).join(', ') || '-'}</p>
            {provider.default && <strong>默认服务</strong>}
            <footer><button type='button' onClick={() => setDefault(provider)}>设为默认</button><button type='button' onClick={() => setModal({ mode: 'edit', item: { ...provider, models: (provider.models || []).join(', ') } })}>编辑</button><button type='button' className='danger' onClick={() => removeProvider(provider)}>删除</button></footer>
          </article>
        ))}
        {!providers.length && <div className='fx-platform-empty'>暂无 AI 服务配置。</div>}
      </div>
      <div className='fx-platform-settings-grid'>
        <ConfigCard title='Embedding 配置'>
          <Field label='Provider'><select value={embedding.provider || 'builtin'} onChange={(e) => setEmbedding((prev) => ({ ...prev, provider: e.target.value }))}><option value='builtin'>内置 BM25</option><option value='api'>外部 API</option><option value='hybrid'>混合模式</option></select></Field>
          <Field label='API URL'><input value={embedding.api_url || ''} onChange={(e) => setEmbedding((prev) => ({ ...prev, api_url: e.target.value }))} /></Field>
          <Field label='API Key'><input type='password' value={embedding.api_key || ''} onChange={(e) => setEmbedding((prev) => ({ ...prev, api_key: e.target.value }))} /></Field>
          <Field label='模型'><input value={embedding.model || ''} onChange={(e) => setEmbedding((prev) => ({ ...prev, model: e.target.value }))} /></Field>
          <footer><button type='button' onClick={saveEmbedding}>保存</button><button type='button' onClick={testEmbedding}>测试连接</button></footer>
        </ConfigCard>
        <ConfigCard title='Reranker 配置'>
          <Field label='启用'><select value={reranker.enabled ? 'true' : 'false'} onChange={(e) => setReranker((prev) => ({ ...prev, enabled: e.target.value === 'true' }))}><option value='false'>关闭</option><option value='true'>启用</option></select></Field>
          <Field label='Provider'><select value={reranker.provider || 'llm'} onChange={(e) => setReranker((prev) => ({ ...prev, provider: e.target.value }))}><option value='llm'>LLM 重排</option><option value='api'>外部 API</option></select></Field>
          <Field label='Top-K'><input value={reranker.top_k || 5} onChange={(e) => setReranker((prev) => ({ ...prev, top_k: Number(e.target.value) || 5 }))} /></Field>
          <Field label='API URL'><input value={reranker.api_url || ''} onChange={(e) => setReranker((prev) => ({ ...prev, api_url: e.target.value }))} /></Field>
          <footer><button type='button' onClick={saveReranker}>保存</button></footer>
        </ConfigCard>
      </div>
      {modal && <ProviderModal modal={modal} onClose={() => setModal(null)} onSave={saveProviderModal} saving={saving} />}
      {confirmModal}
    </section>
  )
}

function ProviderModal({ modal, onClose, onSave, saving }) {
  const [draft, setDraft] = useState(modal.item || emptyProvider)
  const patch = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }))
  return (
    <Modal title={modal.mode === 'edit' ? '编辑 AI 服务' : '新增 AI 服务'} onClose={onClose}>
      <div className='fx-platform-form'>
        <Field label='名称'><input value={draft.name || ''} onChange={(e) => patch('name', e.target.value)} /></Field>
        <Field label='API 地址'><input value={draft.base_url || ''} onChange={(e) => patch('base_url', e.target.value)} /></Field>
        <Field label='API Key'><input type='password' value={draft.api_key || ''} onChange={(e) => patch('api_key', e.target.value)} placeholder='<SECRET>' /></Field>
        <Field label='模型列表'><input value={draft.models || ''} onChange={(e) => patch('models', e.target.value)} placeholder='model-a, model-b' /></Field>
        <footer><button type='button' disabled={saving} onClick={() => onSave(draft)}>{saving ? '保存中...' : '保存'}</button></footer>
      </div>
    </Modal>
  )
}

function ConfigCard({ title, children }) {
  return <article className='fx-platform-config'><h3>{title}</h3>{children}</article>
}

function ContractSection({ type }) {
  const [rows, setRows] = useState([])
  const [message, setMessage] = useState('')
  const configs = {
    site: { title: '站点设置', blocker: PLATFORM_BLOCKERS.site, loader: platformApi.site, fields: ['站点 URL', '首页 URL', '团队展示模式', '业务组展示模式', '访问日志开关'] },
    variables: { title: '变量设置', blocker: PLATFORM_BLOCKERS.variables, loader: platformApi.variables, fields: ['变量名', '说明', '加密开关', '引用范围', '更新人'] },
    sso: { title: '单点登录', blocker: PLATFORM_BLOCKERS.sso, loader: platformApi.sso, fields: ['协议类型', '启用状态', '默认角色', '默认团队', '回调地址'] },
    'alerting-engines': { title: '告警引擎', blocker: PLATFORM_BLOCKERS.engines, loader: platformApi.alertingEngines, fields: ['集群', '实例', '数据源', '心跳', '时钟偏移'] },
  }
  const config = configs[type]

  useEffect(() => {
    let alive = true
    config.loader({}).then((result) => {
      if (!alive) return
      setRows(result.rows || [])
      setMessage(result.source || '')
    }).catch((err) => {
      if (!alive) return
      setRows([])
      setMessage(formatPlatformError(err))
    })
    return () => { alive = false }
  }, [type])

  return (
    <section className='fx-platform-contract'>
      <header><h2>{config.title}</h2><button type='button' disabled>保存</button></header>
      <Blocked>{message || config.blocker}</Blocked>
      <div className='fx-platform-form is-disabled'>
        {config.fields.map((field) => <Field key={field} label={field}><input disabled value='' placeholder='等待 FindX 契约开放' /></Field>)}
      </div>
      {rows.length > 0 && <pre>{safePlatformJson(rows)}</pre>}
    </section>
  )
}

function HealthSection() {
  const [health, setHealth] = useState(null)
  const [error, setError] = useState('')
  const load = async () => {
    try {
      setHealth(await platformApi.storageHealth())
      setError('')
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }
  useEffect(() => { load() }, [])
  return (
    <section className='fx-platform-health'>
      <div className='fx-platform-toolbar'><button type='button' onClick={load}>刷新</button><button type='button' disabled>自动修复</button></div>
      <Blocked>{PLATFORM_BLOCKERS.healthAction}</Blocked>
      {error && <div className='fx-platform-error'>{error}</div>}
      <div className='fx-platform-health-grid'>
        <article><span>MySQL</span><strong>{String(health?.mysql ?? '待确认')}</strong></article>
        <article><span>Redis</span><strong>{String(health?.redis ?? '待确认')}</strong></article>
      </div>
      {health && <pre>{safePlatformJson(health, 1600)}</pre>}
    </section>
  )
}

function AuditSection({ q }) {
  const [rows, setRows] = useState([])
  const [error, setError] = useState('')
  const [source, setSource] = useState('')
  const load = async () => {
    try {
      const result = await platformApi.audit({ q })
      setRows(result.rows || [])
      setSource(result.source || '')
      setError('')
    } catch (err) {
      setError(formatPlatformError(err))
    }
  }
  useEffect(() => { load() }, [q])
  return (
    <section className='fx-platform-audit'>
      <div className='fx-platform-toolbar'><button type='button' onClick={load}>刷新</button><span>{source}</span></div>
      {error && <div className='fx-platform-error'>{error}</div>}
      <div className='fx-platform-table'>
        <table>
          <thead><tr><th>动作</th><th>操作人</th><th>对象</th><th>风险</th><th>决策</th><th>描述</th><th>时间</th></tr></thead>
          <tbody>
            {rows.map((row) => <tr key={row.id || `${row.action}-${row.timestamp || row.created_at}`}><td>{row.action || '-'}</td><td>{row.operator || '-'}</td><td>{row.target || '-'}</td><td>{row.risk || '-'}</td><td>{row.decision || '-'}</td><td>{row.description || row.detail || '-'}</td><td>{row.timestamp ? new Date(row.timestamp).toLocaleString('zh-CN') : row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}</td></tr>)}
          </tbody>
        </table>
        {!rows.length && <div className='fx-platform-empty'>暂无审计日志。</div>}
      </div>
    </section>
  )
}

export function PlatformPage({ query = {}, onNavigate }) {
  const section = sectionSet.has(query.section) ? query.section : 'models'
  const current = useMemo(() => sections.find((item) => item.value === section), [section])
  const [q, setQ] = useState(query.q || '')

  useEffect(() => setQ(query.q || ''), [query.q])

  return (
    <main className='fx-platform-page'>
      <header className='fx-platform-head'>
        <div><p>FindX 平台治理</p><h1>系统配置</h1><span>{current?.desc}</span></div>
        <nav>{sections.map((item) => <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate?.({ section: item.value })}>{item.label}</button>)}</nav>
      </header>
      <section className='fx-platform-filter'>
        <input value={q} onChange={(e) => setQ(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') onNavigate?.({ section, q }) }} placeholder='搜索配置、操作或对象' />
        <button type='button' onClick={() => onNavigate?.({ section, q })}>搜索</button>
      </section>
      {section === 'models' && <ModelsSection />}
      {section === 'mcp' && <McpServersSection />}
      {section === 'sso' && <SsoSection />}
      {['site', 'variables', 'alerting-engines'].includes(section) && <ContractSection type={section} />}
      {section === 'health' && <HealthSection />}
      {section === 'audit' && <AuditSection q={q} />}
    </main>
  )
}
