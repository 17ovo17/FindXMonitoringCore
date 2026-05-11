import React, { useEffect, useMemo, useState } from 'react'
import { AISRE_BLOCKERS, aiSreApi, formatAiSreError } from '../api/aiSre.js'
import { Blocked, Empty, ErrorBox, Field, StatusPill } from './AiSreShared.jsx'

const categories = ['全部', '故障处理', '最佳实践', '架构设计', '运维手册', '告警规则', '其他']

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

export function KnowledgeSection({ query, addEvidence }) {
  const [docs, setDocs] = useState([])
  const [runbooks, setRunbooks] = useState([])
  const [searchQ, setSearchQ] = useState('')
  const [searchResults, setSearchResults] = useState([])
  const [activeCategory, setActiveCategory] = useState('全部')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [editDoc, setEditDoc] = useState(null)
  const [confirmDelete, setConfirmDelete] = useState(null)

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

  const handleSearch = async () => {
    if (!searchQ.trim()) return
    setLoading(true); setError('')
    try {
      const data = await aiSreApi.knowledge.search({ query: searchQ, top_k: 10 })
      const items = data.items || data.results || []
      setSearchResults(items)
      addEvidence({ category: 'knowledge', title: '知识搜索已执行', detail: `${items.length} 条命中` })
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
    if (!window.confirm('确认删除此 Runbook？')) return
    try {
      await aiSreApi.knowledge.removeRunbook(id)
      loadRunbooks()
    } catch (err) {
      setError(formatAiSreError(err))
    }
  }

  useEffect(() => { loadDocs(); loadRunbooks() }, [])

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
        <div className='fx-aisre-category-chips'>
          {categories.map(c => (
            <button key={c} type='button' className={activeCategory === c ? 'is-active' : ''} onClick={() => setActiveCategory(c)}>{c}</button>
          ))}
        </div>
        {showForm && <DocForm initial={editDoc} onSave={handleSaveDoc} onCancel={() => { setShowForm(false); setEditDoc(null) }} />}
        {!showForm && (
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
      </div>
    </section>
  )
}
