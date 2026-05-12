/**
 * Payload 编辑 Modal（含 Monaco Editor）
 * 支持 yaml/json 模式切换
 */
import React, { useRef, useState, useEffect } from 'react'

function LazyMonacoEditor({ language, value, onChange, options }) {
  const [Editor, setEditor] = useState(null)
  const containerRef = useRef(null)

  useEffect(() => {
    import('@monaco-editor/react').then((mod) => {
      setEditor(() => mod.default)
    }).catch(() => {
      // Monaco 不可用，降级
    })
  }, [])

  if (!Editor) {
    return (
      <textarea
        rows={16}
        value={value || ''}
        onChange={(e) => onChange(e.target.value)}
        style={{ width: '100%', height: '100%', fontFamily: 'monospace', fontSize: 13, padding: 12, border: 'none', resize: 'none' }}
      />
    )
  }

  return (
    <Editor
      height='100%'
      language={language}
      value={value || ''}
      onChange={(val) => onChange(val || '')}
      options={options}
    />
  )
}

export function PayloadEditorModal({ state, saving, contentMode = 'json', onDraft, onSubmit, onClose }) {
  if (!state) return null
  const { draft, mode, error } = state
  const editing = mode === 'edit'
  const patch = (key, value) => onDraft({ ...draft, [key]: value })

  const language = contentMode === 'yaml' ? 'yaml' : 'json'

  return (
    <div className='fx-tpl-modal' role='dialog' aria-modal='true'>
      <div className='fx-tpl-modal__backdrop' onClick={onClose} />
      <section className='fx-tpl-modal__panel fx-tpl-modal__panel--wide'>
        <header className='fx-tpl-modal__head'>
          <h2>{editing ? '编辑 Payload' : '新增 Payload'}</h2>
          <button type='button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        <div className='fx-tpl-form'>
          <label>
            名称
            <input
              value={draft.name || ''}
              maxLength={255}
              onChange={(e) => patch('name', e.target.value)}
            />
          </label>
          <label>
            分类
            <input
              value={draft.cate || ''}
              maxLength={128}
              onChange={(e) => patch('cate', e.target.value)}
            />
          </label>
          <label>
            内容（{language.toUpperCase()}）
          </label>
          <div className='fx-tpl-editor-wrap'>
            <LazyMonacoEditor
              language={language}
              value={draft.content || ''}
              onChange={(value) => patch('content', value)}
              options={{
                minimap: { enabled: false },
                fontSize: 13,
                lineNumbers: 'on',
                scrollBeyondLastLine: false,
                wordWrap: 'on',
                tabSize: 2,
              }}
            />
          </div>
        </div>
        {error && <div className='fx-tpl-alert is-error'>{error}</div>}
        <footer className='fx-tpl-modal__foot'>
          <button type='button' onClick={onClose} disabled={saving}>取消</button>
          <button type='button' className='is-primary' onClick={onSubmit} disabled={saving}>
            {saving ? '保存中...' : '保存'}
          </button>
        </footer>
      </section>
    </div>
  )
}
