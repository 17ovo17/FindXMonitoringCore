import React, { useEffect, useMemo, useState } from 'react'
import { AISRE_BLOCKERS, aiSreApi, formatAiSreError } from '../api/aiSre.js'
import { Blocked, Empty, ErrorBox, Field, StatusPill } from './AiSreShared.jsx'
import { useConfirm } from '../shared/ConfirmModal.jsx'

const categories = ['全部', '故障处理', '最佳实践', '架构设计', '运维手册', '告警规则', '其他']
const searchModes = [
  { key: 'normal', label: '普通搜索' },
  { key: 'hyde', label: 'HyDE 增强' },
  { key: 'multi', label: '多查询' },
  { key: 'graph', label: '图谱搜索' },
]

function MarkdownPreview({ content }) {
  if (!content) return <Empty>无内容</Empty>
  try {
    const marked = window.marked || { parse: t => t }
    const html = typeof marked.parse === 'function' ? marked.parse(content) : content
    return <div className='fx-chat-md fx-aisre-md-preview' dangerouslySetInnerHTML={{ __html: html }} />
  } catch {
    return <pre className='fx-aisre-text'>{content}</pre>
  }
}

function DocForm({ initial, onSave, onCancel }) {
  const [title, setTitle] = useState(initial?.title || '')
  const [category, setCategory] = useState(initial?.category || '其他')
  const [content, setContent] = useState(initial?.content || '')
  const [tags, setTags] = useState(initial?.tags?.join(', ') || '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [preview, setPreview] = useState(false)

  const handleSubmit = async e => {
    e.preventDefault()
    if (!title.trim()) { setError('请输入标题'); return }
    setSaving(true); setError('')
    try {
      await onSave({ title, category, content, tags: tags.split(',').map(s => s.trim()).filter(Boolean) })
    } catch (err) {
      setError(formatAiSreError(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <form className='fx-aisre-doc-form' onSubmit={handleSubmit}>
      <h3>{initial ? '编辑文档' : '创建文档'}</h3>
      <ErrorBox>{error}</ErrorBox>
      <Field label='标题'><input value={title} onChange={e => setTitle(e.target.value)} placeholder='文档标题' /></Field>
      <Field label='分类'>
        <select value={category} onChange={e => setCategory(e.target.value)}>
          {categories.filter(c => c !== '全部').map(c => <option key={c} value={c}>{c}</option>)}
        </select>
      </Field>
      <Field label='标签 (逗号分隔)'><input value={tags} onChange={e => setTags(e.target.value)} placeholder='k8s, 内存泄漏' /></Field>
      <Field label='内容 (支持 Markdown)'>
        <div className='fx-aisre-toolbar'>
          <button type='button' onClick={() => setPreview(!preview)}>{preview ? '编辑' : '预览'}</button>
        </div>
        {preview ? <MarkdownPreview content={content} /> : <textarea value={content} onChange={e => setContent(e.target.value)} rows={8} placeholder='# 标题&#10;正文内容...' />}
      </Field>
      <div className='fx-aisre-toolbar'>
        <button type='submit' disabled={saving}>{saving ? '保存中...' : '保存'}</button>
        <button type='button' onClick={onCancel}>取消</button>
      </div>
    </form>
  )
}

function RunbookList({ runbooks, onDelete }) {
  if (!runbooks.length) return <Empty>暂无 Runbook</Empty>
  return (
    <div className='fx-aisre-runbook-list'>
      {runbooks.map(rb => (
        <article key={rb.id} className='fx-aisre-runbook-card'>
          <div className='fx-aisre-runbook-head'>
            <strong>{rb.title || rb.name || rb.id}</strong>
            <StatusPill tone={rb.auto_executable ? 'ok' : 'blocked'}>
              {rb.auto_executable ? '可自动执行' : '手动执行'}
            </StatusPill>
            <button type='button' className='fx-aisre-btn-sm fx-aisre-btn-danger' onClick={() => onDelete(rb.id)}>删除</button>
          </div>
          {rb.trigger_conditions && <p className='fx-aisre-note'>触发条件: {rb.trigger_conditions}</p>}
          {rb.steps?.length > 0 && (
            <ol className='fx-aisre-runbook-steps'>
              {rb.steps.map((step, i) => <li key={i}>{step.description || step.action || step}</li>)}
            </ol>
          )}
        </article>
      ))}
    </div>
  )
}

// 搜索结果反馈按钮
function FeedbackButtons({ query, docId }) {
  const [sent, setSent] = useState(null)
  const handleFeedback = async type => {
    try {
      await aiSreApi.knowledge.feedback({ query, doc_id: docId, type })
      setSent(type)
    } catch { /* 静默失败 */ }
  }
  if (sent) return <span className='fx-aisre-feedback-sent'>{sent === 'like' ? '已点赞' : '已点踩'}</span>
  return (
    <span className='fx-aisre-feedback-btns'>
      <button type='button' className='fx-aisre-btn-sm' onClick={() => handleFeedback('like')} title='有帮助'>&#x1F44D;</button>
      <button type='button' className='fx-aisre-btn-sm' onClick={() => handleFeedback('dislike')} title='无帮助'>&#x1F44E;</button>
    </span>
  )
}

// 简单知识图谱可视化（SVG）
function GraphVisualization({ entities, relations }) {
  if (!entities?.length) return <Empty>暂无图谱数据</Empty>
  const width = 600
  const height = 400
  const cx = width / 2
  const cy = height / 2
  const radius = Math.min(width, height) * 0.35

  // 环形布局
  const positions = entities.map((e, i) => {
    const angle = (2 * Math.PI * i) / entities.length - Math.PI / 2
    return { x: cx + radius * Math.cos(angle), y: cy + radius * Math.sin(angle), entity: e }
  })
  const posMap = {}
  positions.forEach(p => { posMap[p.entity.id] = p })

  return (
    <svg viewBox={`0 0 ${width} ${height}`} className='fx-aisre-graph-svg' style={{ width: '100%', maxHeight: 400 }}>
      {/* 关系线 */}
      {(relations || []).map((rel, i) => {
        const src = posMap[rel.source]
        const tgt = posMap[rel.target]
        if (!src || !tgt) return null
        return <line key={i} x1={src.x} y1={src.y} x2={tgt.x} y2={tgt.y} stroke='#94a3b8' strokeWidth={1} opacity={0.6} />
      })}
      {/* 实体节点 */}
      {positions.map((p, i) => (
        <g key={i}>
          <circle cx={p.x} cy={p.y} r={18} fill={entityColor(p.entity.type)} stroke='#475569' strokeWidth={1.5} />
          <text x={p.x} y={p.y + 30} textAnchor='middle' fontSize={10} fill='#334155'>{p.entity.name?.slice(0, 12)}</text>
        </g>
      ))}
    </svg>
  )
}

function entityColor(type) {
  const colors = { service: '#3b82f6', host: '#10b981', metric: '#f59e0b', alert: '#ef4444', runbook: '#8b5cf6', incident: '#ec4899', config: '#6b7280' }
  return colors[type] || '#6b7280'
}

// 搜索统计面板
function StatsPanel({ stats }) {
  if (!stats) return null
  return (
    <div className='fx-aisre-stats-panel'>
      <h4>知识库统计</h4>
      {stats.graph && (
        <div className='fx-aisre-stats-row'>
          <span>图谱实体: {stats.graph.entity_count}</span>
          <span>图谱关系: {stats.graph.relation_count}</span>
        </div>
      )}
      {stats.feedback && (
        <div className='fx-aisre-stats-row'>
          <span>反馈总数: {stats.feedback.total}</span>
          {stats.feedback.top_liked?.length > 0 && <span>最受欢迎: {stats.feedback.top_liked[0]?.doc_id}</span>}
        </div>
      )}
    </div>
  )
}

export function KnowledgeSection({ query, addEvidence }) {
  const [docs, setDocs] = useState([])
  const [runbooks, setRunbooks] = useState([])
  const [searchQ, setSearchQ] = useState('')
  const [searchResults, setSearchResults] = useState([])
  const [activeCategory, setActiveCategory] = useState('全部')
  const [searchMode, setSearchMode] = useState('normal')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [editDoc, setEditDoc] = useState(null)
  const [confirmDelete, setConfirmDelete] = useState(null)
  const [stats, setStats] = useState(null)
  const [graphData, setGraphData] = useState(null)
  const { confirm, modal: confirmModal } = useConfirm()

  const loadDocs = async () => {
    setLoading(true); setError('')
    try {
      const items = await aiSreApi.knowledge.list()
      setDocs(items)
      addEvidence({ category: 'knowledge', title: '知识库文档已读取', detail: `${items.length} 篇文档` })
    } catch (err) {
      setError(formatAiSreError(err))
    } finally {
      setLoading(false)
    }
  }

  const loadRunbooks = async () => {
    try {
      const items = await aiSreApi.knowledge.runbooks()
      setRunbooks(items)
    } catch { setRunbooks([]) }
  }

  const loadStats = async () => {
    try {
      const data = await aiSreApi.knowledge.stats()
      setStats(data)
    } catch { /* 静默 */ }
  }

  const handleSearch = async () => {
    if (!searchQ.trim()) return
    setLoading(true); setError('')
    try {
      let data
      switch (searchMode) {
        case 'hyde':
          data = await aiSreApi.knowledge.searchHyDE({ query: searchQ, top_k: 10 })
          break
        case 'multi':
          data = await aiSreApi.knowledge.searchMulti({ query: searchQ, top_k: 10 })
          break
        case 'graph':
          data = await aiSreApi.knowledge.graphSearch({ query: searchQ, top_k: 10 })
          break
        default:
          data = await aiSreApi.knowledge.search({ query: searchQ, top_k: 10 })
      }
      const items = data?.search_results || data?.items || data?.results || []
      setSearchResults(items)
      addEvidence({ category: 'knowledge', title: `知识搜索已执行 (${searchMode})`, detail: `${items.length} 条命中` })
    } catch (err) {
      setError(formatAiSreError(err))
    } finally {
      setLoading(false)
    }
  }

  const handleSaveDoc = async data => {
    if (editDoc?.id) {
      await aiSreApi.knowledge.update(editDoc.id, data)
    } else {
      await aiSreApi.knowledge.create(data)
    }
    setShowForm(false); setEditDoc(null)
    addEvidence({ category: 'knowledge', title: editDoc ? '文档已更新' : '文档已创建', detail: data.title })
    loadDocs()
  }

  const handleDelete = async id => {
    try {
      await aiSreApi.knowledge.remove(id)
      setConfirmDelete(null)
      addEvidence({ category: 'knowledge', title: '文档已删除', detail: id })
      loadDocs()
    } catch (err) {
      setError(formatAiSreError(err))
    }
  }

  const handleDeleteRunbook = async id => {
    const ok = await confirm({ title: '删除 Runbook', message: '确认删除此 Runbook？', confirmText: '删除', danger: true })
    if (!ok) return
    try {
      await aiSreApi.knowledge.removeRunbook(id)
      loadRunbooks()
    } catch (err) {
      setError(formatAiSreError(err))
    }
  }

  const loadGraph = async () => {
    try {
      const data = await aiSreApi.knowledge.graph({})
      setGraphData(data)
    } catch { setGraphData(null) }
  }

  useEffect(() => { loadDocs(); loadRunbooks(); loadStats(); loadGraph() }, [])

  const filteredDocs = useMemo(() => {
    let items = docs
    if (activeCategory !== '全部') items = items.filter(d => d.category === activeCategory)
    if (searchQ.trim() && !searchResults.length) {
      const q = searchQ.toLowerCase()
      items = items.filter(d => (d.title || '').toLowerCase().includes(q) || (d.content || '').toLowerCase().includes(q))
    }
    return items
  }, [docs, activeCategory, searchQ, searchResults])

  return (
    <section className='fx-aisre-grid'>
      <div className='fx-aisre-panel'>
        <div className='fx-aisre-toolbar'>
          <h2>知识库</h2>
          <button type='button' onClick={() => { setShowForm(true); setEditDoc(null) }}>+ 创建文档</button>
          <button type='button' onClick={loadDocs}>{loading ? '读取中...' : '刷新'}</button>
        </div>
        <ErrorBox>{error}</ErrorBox>
        <div className='fx-aisre-knowledge-search'>
          <input value={searchQ} onChange={e => setSearchQ(e.target.value)} placeholder='搜索标题或内容...' onKeyDown={e => e.key === 'Enter' && handleSearch()} />
          <button type='button' onClick={handleSearch} disabled={loading || !searchQ.trim()}>搜索</button>
        </div>
        {/* 搜索模式切换 */}
        <div className='fx-aisre-category-chips'>
          {searchModes.map(m => (
            <button key={m.key} type='button' className={searchMode === m.key ? 'is-active' : ''} onClick={() => setSearchMode(m.key)}>{m.label}</button>
          ))}
        </div>
        <div className='fx-aisre-category-chips'>
          {categories.map(c => (
            <button key={c} type='button' className={activeCategory === c ? 'is-active' : ''} onClick={() => setActiveCategory(c)}>{c}</button>
          ))}
        </div>
        {/* 搜索结果（带反馈按钮） */}
        {searchResults.length > 0 && (
          <div className='fx-aisre-doc-list'>
            <h4>搜索结果</h4>
            {searchResults.map((r, i) => (
              <article key={r.doc_id || i} className='fx-aisre-doc-item'>
                <div className='fx-aisre-doc-item-head'>
                  <strong>{r.title || r.doc_id}</strong>
                  <span className='fx-aisre-score'>得分: {(r.score || 0).toFixed(3)}</span>
                  <FeedbackButtons query={searchQ} docId={r.doc_id} />
                </div>
                {r.content && <p className='fx-aisre-note'>{r.content.slice(0, 200)}</p>}
              </article>
            ))}
          </div>
        )}
        {showForm && <DocForm initial={editDoc} onSave={handleSaveDoc} onCancel={() => { setShowForm(false); setEditDoc(null) }} />}
        {!showForm && !searchResults.length && (
          <div className='fx-aisre-doc-list'>
            {!filteredDocs.length && <Empty>暂无文档</Empty>}
            {filteredDocs.map(doc => (
              <article key={doc.id} className='fx-aisre-doc-item'>
                <div className='fx-aisre-doc-item-head'>
                  <strong>{doc.title}</strong>
                  <span>{doc.category || '未分类'}</span>
                </div>
                <div className='fx-aisre-doc-item-meta'>
                  {doc.tags?.length > 0 && <span className='fx-aisre-doc-tags'>{doc.tags.join(', ')}</span>}
                  {doc.updated_at && <time>{new Date(doc.updated_at).toLocaleString('zh-CN', { hour12: false })}</time>}
                  {doc.author && <span>{doc.author}</span>}
                </div>
                <div className='fx-aisre-doc-item-actions'>
                  <button type='button' className='fx-aisre-btn-sm' onClick={() => { setEditDoc(doc); setShowForm(true) }}>编辑</button>
                  {confirmDelete === doc.id ? (
                    <>
                      <button type='button' className='fx-aisre-btn-sm fx-aisre-btn-danger' onClick={() => handleDelete(doc.id)}>确认删除</button>
                      <button type='button' className='fx-aisre-btn-sm' onClick={() => setConfirmDelete(null)}>取消</button>
                    </>
                  ) : (
                    <button type='button' className='fx-aisre-btn-sm' onClick={() => setConfirmDelete(doc.id)}>删除</button>
                  )}
                </div>
              </article>
            ))}
          </div>
        )}
      </div>
      <div className='fx-aisre-panel'>
        <h2>Runbook</h2>
        <RunbookList runbooks={runbooks} onDelete={handleDeleteRunbook} />
        <Blocked>{AISRE_BLOCKERS.knowledgeWrite}</Blocked>
        {/* 知识图谱可视化 */}
        <h3 style={{ marginTop: 16 }}>知识图谱</h3>
        <GraphVisualization entities={graphData?.entities} relations={graphData?.relations} />
        {/* 统计面板 */}
        <StatsPanel stats={stats} />
      </div>
      {confirmModal}
    </section>
  )
}
