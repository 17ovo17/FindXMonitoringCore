import React, { useCallback, useEffect, useRef, useState } from 'react'
import { marked } from 'marked'
import sanitizeHtml from 'sanitize-html'
import { aiSreApi, formatAiSreError } from '../api/aiSre.js'
import { ErrorBox } from './AiSreShared.jsx'

/* ─── 配置 ─── */
const QUICK_QUESTIONS = [
  'CPU 使用率过高怎么办',
  '磁盘空间不足',
  '服务响应慢',
  '网络连接异常',
  '内存使用率告警',
  '数据库连接池满',
  '容器 OOMKilled',
  '证书即将过期',
]

/* ─── NL2Query 快捷示例 ─── */
const NL2QUERY_EXAMPLES = [
  '最近1小时CPU使用率超过80%的主机',
  '连接数最多的前10个MySQL实例',
  '最近30分钟的错误日志',
  'Redis内存使用情况',
]

const SANITIZE_OPTS = {
  allowedTags: sanitizeHtml.defaults.allowedTags.concat(['img', 'h1', 'h2', 'h3', 'h4', 'details', 'summary', 'code', 'pre', 'table', 'thead', 'tbody', 'tr', 'th', 'td']),
  allowedAttributes: { ...sanitizeHtml.defaults.allowedAttributes, img: ['src', 'alt'], code: ['class'], span: ['class'] },
}

const renderMarkdown = text => {
  if (!text) return ''
  const html = marked.parse(text, { breaks: true, gfm: true })
  return sanitizeHtml(html, SANITIZE_OPTS)
}

/* ─── 会话列表面板 ─── */
function SessionPanel({ sessions, activeId, onSelect, onCreate, onRename, onDelete, loading }) {
  const [editingId, setEditingId] = useState(null)
  const [editTitle, setEditTitle] = useState('')

  const startRename = (e, s) => {
    e.stopPropagation()
    setEditingId(s.sessionId || s.id)
    setEditTitle(s.title || '')
  }

  const confirmRename = (e, s) => {
    e.stopPropagation()
    const id = s.sessionId || s.id
    if (editTitle.trim() && editTitle.trim() !== (s.title || '')) {
      onRename(id, editTitle.trim())
    }
    setEditingId(null)
  }

  return (
    <aside className='fx-chat-sidebar'>
      <div className='fx-chat-sidebar-head'>
        <h3>历史会话</h3>
        <button type='button' onClick={onCreate} disabled={loading} title='新建会话'>+</button>
      </div>
      <div className='fx-chat-session-list'>
        {!sessions.length && <p className='fx-chat-empty-hint'>暂无会话，点击 + 新建</p>}
        {sessions.map(s => {
          const id = s.sessionId || s.id
          const isActive = id === activeId
          return (
            <div key={id} className={`fx-chat-session-item${isActive ? ' is-active' : ''}`} onClick={() => onSelect(s)}>
              {editingId === id ? (
                <input
                  className='fx-chat-rename-input'
                  value={editTitle}
                  onChange={e => setEditTitle(e.target.value)}
                  onBlur={e => confirmRename(e, s)}
                  onKeyDown={e => { if (e.key === 'Enter') confirmRename(e, s); if (e.key === 'Escape') setEditingId(null) }}
                  onClick={e => e.stopPropagation()}
                  autoFocus
                />
              ) : (
                <span className='fx-chat-session-title'>{s.title || '未命名会话'}</span>
              )}
              <span className='fx-chat-session-time'>{s.created_at || s.createdAt || ''}</span>
              <div className='fx-chat-session-actions'>
                <button type='button' onClick={e => startRename(e, s)} title='重命名'>&#9998;</button>
                <button type='button' onClick={e => { e.stopPropagation(); onDelete(id) }} title='删除'>&times;</button>
              </div>
            </div>
          )
        })}
      </div>
    </aside>
  )
}

/* ─── 消息气泡 ─── */
function ChatBubble({ message }) {
  const isUser = message.role === 'user'
  const htmlContent = isUser ? null : renderMarkdown(message.content)
  return (
    <div className={`fx-chat-bubble ${isUser ? 'is-user' : 'is-ai'}`}>
      <div className='fx-chat-bubble-label'>{isUser ? '我' : 'AI 助手'}</div>
      {isUser ? (
        <div className='fx-chat-bubble-body'>{message.content}</div>
      ) : (
        <div className='fx-chat-bubble-body fx-chat-md' dangerouslySetInnerHTML={{ __html: htmlContent }} />
      )}
    </div>
  )
}

/* ─── 流式指示器 ─── */
function StreamingIndicator() {
  return (
    <div className='fx-chat-bubble is-ai'>
      <div className='fx-chat-bubble-label'>AI 助手</div>
      <div className='fx-chat-bubble-body fx-chat-typing'>
        <span className='fx-chat-dot' /><span className='fx-chat-dot' /><span className='fx-chat-dot' />
      </div>
    </div>
  )
}

/* ─── NL2Query 面板 ─── */
function NL2QueryPanel({ onClose }) {
  const [input, setInput] = useState('')
  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleTranslate = async (text) => {
    const query = (text || input).trim()
    if (!query) return
    setLoading(true)
    setError('')
    setResult(null)
    try {
      const resp = await fetch('/api/v1/ai/nl2query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ input: query }),
      })
      const data = await resp.json()
      if (data.code !== 0) throw new Error(data.error || '查询转换失败')
      setResult(data.data)
    } catch (err) {
      setError(err.message || '请求失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className='fx-nl2query-panel'>
      <div className='fx-nl2query-header'>
        <h4>自然语言查询</h4>
        <button type='button' onClick={onClose} title='关闭'>&times;</button>
      </div>
      <div className='fx-nl2query-body'>
        <div className='fx-nl2query-input-row'>
          <input
            type='text'
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') handleTranslate() }}
            placeholder='输入中文问题，如：最近1小时CPU使用率超过80%的主机'
          />
          <button type='button' onClick={() => handleTranslate()} disabled={loading || !input.trim()}>
            {loading ? '转换中...' : '转换'}
          </button>
        </div>
        <div className='fx-nl2query-examples'>
          {NL2QUERY_EXAMPLES.map(ex => (
            <button key={ex} type='button' className='fx-chat-chip' onClick={() => { setInput(ex); handleTranslate(ex) }}>{ex}</button>
          ))}
        </div>
        {error && <div className='fx-nl2query-error'>{error}</div>}
        {result && (
          <div className='fx-nl2query-result'>
            <div className='fx-nl2query-result-lang'>
              <span className='fx-nl2query-badge'>{result.target_language}</span>
              <span className='fx-nl2query-confidence'>置信度: {Math.round((result.confidence || 0) * 100)}%</span>
            </div>
            <pre className='fx-nl2query-code'>{result.generated_query}</pre>
            <p className='fx-nl2query-explanation'>{result.explanation}</p>
          </div>
        )}
      </div>
    </div>
  )
}

/* ─── 主组件 ─── */
export function DiagnosisSection({ query, onNavigate, addEvidence }) {
  const [sessions, setSessions] = useState([])
  const [activeSession, setActiveSession] = useState(null)
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [streaming, setStreaming] = useState(false)
  const [streamContent, setStreamContent] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [showNL2Query, setShowNL2Query] = useState(false)
  const messagesEndRef = useRef(null)
  const abortRef = useRef(null)

  const activeId = activeSession?.sessionId || activeSession?.id || null

  /* 自动滚动到底部 */
  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [])

  useEffect(() => { scrollToBottom() }, [messages, streamContent, scrollToBottom])

  /* 加载会话列表 */
  const loadSessions = useCallback(async () => {
    try {
      const result = await aiSreApi.chat.listSessions({ pageSize: 50 })
      setSessions(result.list || [])
    } catch (err) {
      setError(formatAiSreError(err))
    }
  }, [])

  useEffect(() => { loadSessions() }, [loadSessions])

  /* 选择会话 */
  const selectSession = useCallback(async (session) => {
    const id = session.sessionId || session.id
    setActiveSession(session)
    setError('')
    setStreamContent('')
    onNavigate({ section: 'diagnosis', sessionId: id })
    try {
      const msgs = await aiSreApi.chat.getMessages(id)
      setMessages(msgs)
      addEvidence({ category: 'diagnosis', title: '会话消息已加载', detail: `${msgs.length} 条消息` })
    } catch (err) {
      setError(formatAiSreError(err))
      setMessages([])
    }
  }, [onNavigate, addEvidence])

  /* 新建会话 */
  const createSession = useCallback(async () => {
    setLoading(true); setError('')
    try {
      const created = await aiSreApi.chat.createSession({ title: 'AI 助手会话', mode: 'diagnostic', scope: {}, context: {} })
      await loadSessions()
      setActiveSession(created)
      setMessages([])
      onNavigate({ section: 'diagnosis', sessionId: created.sessionId || created.id })
      addEvidence({ category: 'diagnosis', title: '新建会话', detail: created.title || created.sessionId || created.id })
    } catch (err) {
      setError(formatAiSreError(err))
    } finally {
      setLoading(false)
    }
  }, [loadSessions, onNavigate, addEvidence])

  /* 删除会话 */
  const deleteSession = useCallback(async (id) => {
    try {
      await aiSreApi.chat.deleteSession(id)
      if (activeId === id) { setActiveSession(null); setMessages([]) }
      await loadSessions()
    } catch (err) {
      setError(formatAiSreError(err))
    }
  }, [activeId, loadSessions])

  /* 重命名会话 */
  const renameSession = useCallback(async (id, title) => {
    try {
      await aiSreApi.chat.renameSession(id, title)
      await loadSessions()
    } catch (err) {
      setError(formatAiSreError(err))
    }
  }, [loadSessions])
  /* 发送消息 (SSE 流式) */
  const sendMessage = useCallback(async (text) => {
    const content = (text || input).trim()
    if (!content || streaming) return
    setError('')

    /* 确保有活跃会话 */
    let session = activeSession
    if (!session) {
      try {
        session = await aiSreApi.chat.createSession({ title: content.slice(0, 30), mode: 'diagnostic', scope: {}, context: {} })
        setActiveSession(session)
        await loadSessions()
        onNavigate({ section: 'diagnosis', sessionId: session.sessionId || session.id })
      } catch (err) {
        setError(formatAiSreError(err)); return
      }
    }

    const sessionId = session.sessionId || session.id
    const userMsg = { role: 'user', content, id: `u-${Date.now()}` }
    setMessages(prev => [...prev, userMsg])
    setInput('')
    setStreaming(true)
    setStreamContent('')

    let accumulated = ''
    const { promise, abort } = aiSreApi.chat.sendMessageSSE(sessionId, {
      role: 'user', content, audience: 'ops', attachments: [],
    }, event => {
      if (event.type === 'content') {
        accumulated += (event.content || '')
        setStreamContent(accumulated)
      } else if (event.type === 'done') {
        /* handled after promise resolves */
      }
    })

    abortRef.current = abort

    try {
      await promise
      if (accumulated) {
        setMessages(prev => [...prev, { role: 'assistant', content: accumulated, id: `a-${Date.now()}` }])
      }
      setStreamContent('')
      addEvidence({ category: 'diagnosis', title: 'AI 响应', detail: accumulated.slice(0, 80) || '流式响应完成' })
    } catch (err) {
      if (err.name !== 'AbortError') {
        setError(formatAiSreError(err))
      }
    } finally {
      setStreaming(false)
      abortRef.current = null
    }
  }, [input, streaming, activeSession, loadSessions, onNavigate, addEvidence])

  const handleKeyDown = e => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage() }
  }

  const handleQuickQuestion = q => {
    setInput(q)
    sendMessage(q)
  }

  return (
    <section className='fx-chat-layout'>
      <SessionPanel
        sessions={sessions}
        activeId={activeId}
        onSelect={selectSession}
        onCreate={createSession}
        onRename={renameSession}
        onDelete={deleteSession}
        loading={loading}
      />
      <div className='fx-chat-main'>
        {/* 消息区域 */}
        <div className='fx-chat-messages'>
          {!messages.length && !streaming && (
            <div className='fx-chat-welcome'>
              <h2>AI 助手</h2>
              <p>您好，我是 AI 运维助手。请选择一个快捷问题或直接输入您的问题。</p>
              <div className='fx-chat-chips'>
                {QUICK_QUESTIONS.map(q => (
                  <button key={q} type='button' className='fx-chat-chip' onClick={() => handleQuickQuestion(q)}>{q}</button>
                ))}
              </div>
            </div>
          )}
          {messages.map(msg => <ChatBubble key={msg.id || msg.messageId || msg.createdAt} message={msg} />)}
          {streaming && streamContent && (
            <div className='fx-chat-bubble is-ai'>
              <div className='fx-chat-bubble-label'>AI 助手</div>
              <div className='fx-chat-bubble-body fx-chat-md' dangerouslySetInnerHTML={{ __html: renderMarkdown(streamContent) }} />
            </div>
          )}
          {streaming && !streamContent && <StreamingIndicator />}
          <div ref={messagesEndRef} />
        </div>
        {/* 输入区域 */}
        <div className='fx-chat-input-area'>
          <ErrorBox>{error}</ErrorBox>
          {showNL2Query && <NL2QueryPanel onClose={() => setShowNL2Query(false)} />}
          {!messages.length && !streaming && null}
          {(messages.length > 0 || streaming) && (
            <div className='fx-chat-chips fx-chat-chips-inline'>
              {QUICK_QUESTIONS.slice(0, 4).map(q => (
                <button key={q} type='button' className='fx-chat-chip' onClick={() => handleQuickQuestion(q)} disabled={streaming}>{q}</button>
              ))}
            </div>
          )}
          <div className='fx-chat-input-row'>
            <textarea
              className='fx-chat-textarea'
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder='输入您的问题... (Enter 发送, Shift+Enter 换行)'
              rows={2}
              disabled={streaming}
            />
            <button
              type='button'
              className='fx-chat-nl2query-btn'
              onClick={() => setShowNL2Query(v => !v)}
              title='自然语言查询'
            >
              NL
            </button>
            <button
              type='button'
              className='fx-chat-send-btn'
              onClick={() => sendMessage()}
              disabled={streaming || !input.trim()}
            >
              {streaming ? '回答中...' : '发送'}
            </button>
          </div>
        </div>
      </div>
    </section>
  )
}
