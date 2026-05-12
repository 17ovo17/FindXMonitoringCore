import React, { useState } from 'react'
import { dashboardsApi } from '../../api/dashboards.js'

const TOOLTIP_OPTIONS = [
  { key: 'default', label: '默认' },
  { key: 'sharedCrosshair', label: '共享十字线' },
  { key: 'sharedTooltip', label: '共享提示信息 (Tooltip)' },
]

const ZOOM_OPTIONS = [
  { key: 'default', label: '默认' },
  { key: 'updateTimeRange', label: '更新时间范围' },
]

/**
 * 编辑仪表盘 Modal（对齐夜莺 FormModal）
 * 字段：仪表盘名称 / 英文标识 / 分类标签 / 提示信息 / 缩放行为
 */
export default function DashboardSettingsModal({ dashboard, onClose, onSaved }) {
  const [title, setTitle] = useState(dashboard.title || '')
  const [ident, setIdent] = useState(dashboard.raw?.ident || '')
  const [tagInput, setTagInput] = useState('')
  const [tags, setTags] = useState(() => {
    const raw = dashboard.tags || []
    return Array.isArray(raw) ? raw : []
  })
  const [tooltipMode, setTooltipMode] = useState(
    dashboard.raw?.graphTooltip || 'default'
  )
  const [zoomBehavior, setZoomBehavior] = useState(
    dashboard.raw?.graphZoom || 'default'
  )
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const handleTagKeyDown = (e) => {
    if (e.key === 'Enter' && tagInput.trim()) {
      e.preventDefault()
      const newTag = tagInput.trim()
      if (!tags.includes(newTag)) {
        setTags([...tags, newTag])
      }
      setTagInput('')
    }
  }

  const removeTag = (tagToRemove) => {
    setTags(tags.filter((t) => t !== tagToRemove))
  }

  const handleSave = async () => {
    if (!title.trim()) {
      setError('仪表盘名称不能为空')
      return
    }
    setSaving(true)
    setError('')
    try {
      const body = {
        ...dashboard.raw,
        title: title.trim(),
        ident,
        tags,
        graphTooltip: tooltipMode,
        graphZoom: zoomBehavior,
      }
      await dashboardsApi.update(dashboard.id, body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fx-dash-modal">
      <div className="fx-dash-modal__body fx-settings-modal">
        <header>
          <h2>编辑仪表盘</h2>
          <button type="button" onClick={onClose}>x</button>
        </header>
        <div className="fx-settings-body">
          <div className="fx-settings-section">
            <label className="fx-pe-field">
              <span>
                <span className="fx-field-required">*</span>
                仪表盘名称
              </span>
              <input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="请输入仪表盘名称"
              />
            </label>
            <label className="fx-pe-field">
              <span>英文标识</span>
              <input
                value={ident}
                onChange={(e) => setIdent(e.target.value)}
                placeholder="请输入英文标识 (ident)"
              />
            </label>
            <div className="fx-pe-field">
              <span>分类标签</span>
              <div className="fx-tag-input-wrap">
                <div className="fx-tag-input-tags">
                  {tags.map((tag) => (
                    <span key={tag} className="fx-tag-input-tag">
                      {tag}
                      <button
                        type="button"
                        className="fx-tag-input-tag__close"
                        onClick={() => removeTag(tag)}
                      >
                        x
                      </button>
                    </span>
                  ))}
                  <input
                    className="fx-tag-input-field"
                    value={tagInput}
                    onChange={(e) => setTagInput(e.target.value)}
                    onKeyDown={handleTagKeyDown}
                    placeholder={tags.length === 0 ? '输入后回车添加标签' : ''}
                  />
                </div>
              </div>
            </div>
            <div className="fx-pe-field">
              <span>提示信息 (Tooltip)</span>
              <div className="fx-btn-group">
                {TOOLTIP_OPTIONS.map((opt) => (
                  <button
                    key={opt.key}
                    type="button"
                    className={tooltipMode === opt.key ? 'is-active' : ''}
                    onClick={() => setTooltipMode(opt.key)}
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
            </div>
            <div className="fx-pe-field">
              <span>缩放行为</span>
              <div className="fx-btn-group">
                {ZOOM_OPTIONS.map((opt) => (
                  <button
                    key={opt.key}
                    type="button"
                    className={zoomBehavior === opt.key ? 'is-active' : ''}
                    onClick={() => setZoomBehavior(opt.key)}
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
        {error && <div className="fx-dash-alert is-error">{error}</div>}
        <footer className="fx-settings-footer">
          <button type="button" onClick={onClose}>取消</button>
          <button
            type="button"
            className="is-primary"
            disabled={saving}
            onClick={handleSave}
          >
            {saving ? '保存中...' : '确定'}
          </button>
        </footer>
      </div>
    </div>
  )
}
