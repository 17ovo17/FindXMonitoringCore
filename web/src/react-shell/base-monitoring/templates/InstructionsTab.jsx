/**
 * 采集说明 Tab
 * Markdown 渲染 + 编辑模式
 */
import React, { useState } from 'react'
import { marked } from 'marked'
import sanitizeHtml from 'sanitize-html'

export function InstructionsTab({ component, onUpdateReadme }) {
  const [editing, setEditing] = useState(false)
  const [value, setValue] = useState(component.readme || '')
  const [saving, setSaving] = useState(false)

  const handleSave = async () => {
    setSaving(true)
    try {
      await onUpdateReadme(component, value)
      setEditing(false)
    } catch {
      // 错误由上层处理
    } finally {
      setSaving(false)
    }
  }

  const handleCancel = () => {
    setValue(component.readme || '')
    setEditing(false)
  }

  const renderMarkdown = (source) => {
    const raw = marked.parse(source || '暂无说明', { breaks: true })
    return sanitizeHtml(raw, {
      allowedTags: sanitizeHtml.defaults.allowedTags.concat(['img', 'h1', 'h2', 'h3']),
      allowedAttributes: {
        ...sanitizeHtml.defaults.allowedAttributes,
        img: ['src', 'alt', 'width', 'height'],
      },
    })
  }

  if (editing) {
    return (
      <div>
        <textarea
          rows={20}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          style={{ width: '100%', fontFamily: 'monospace', fontSize: 13, padding: 12, border: '1px solid var(--fx-border, #d9d9d9)', borderRadius: 4 }}
        />
        <div style={{ marginTop: 12, display: 'flex', gap: 8 }}>
          <button type='button' className='is-primary' onClick={handleSave} disabled={saving}>
            {saving ? '保存中...' : '保存'}
          </button>
          <button type='button' onClick={handleCancel} disabled={saving}>取消</button>
        </div>
      </div>
    )
  }

  return (
    <div>
      <div
        className='fx-tpl-markdown'
        dangerouslySetInnerHTML={{ __html: renderMarkdown(component.readme) }}
      />
      <div style={{ marginTop: 12 }}>
        <button type='button' className='is-primary' onClick={() => setEditing(true)}>
          编辑说明
        </button>
      </div>
    </div>
  )
}
