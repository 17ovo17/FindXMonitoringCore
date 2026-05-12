/**
 * 组件创建/编辑 Modal
 * T02: logo URL + ident + readme（Markdown）+ disabled
 */
import React from 'react'

export function ComponentFormModal({ state, saving, onDraft, onSubmit, onClose }) {
  if (!state) return null
  const { draft, mode, error } = state
  const editing = mode === 'edit'
  const patch = (key, value) => onDraft({ ...draft, [key]: value })

  return (
    <div className='fx-tpl-modal' role='dialog' aria-modal='true'>
      <div className='fx-tpl-modal__backdrop' onClick={onClose} />
      <section className='fx-tpl-modal__panel fx-tpl-modal__panel--wide'>
        <header className='fx-tpl-modal__head'>
          <h2>{editing ? '编辑组件' : '新增组件'}</h2>
          <button type='button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        <div className='fx-tpl-form'>
          <label>
            组件标识（ident）
            <input
              value={draft.ident || ''}
              maxLength={128}
              disabled={editing}
              onChange={(e) => patch('ident', e.target.value)}
              placeholder='如 mysql、redis、host'
            />
          </label>
          <label>
            Logo URL
            <input
              value={draft.logo || ''}
              onChange={(e) => patch('logo', e.target.value)}
              placeholder='/image/logos/host.png'
            />
          </label>
          {draft.logo && (
            <div style={{ marginBottom: 12 }}>
              <img src={draft.logo} alt='logo preview' style={{ height: 42, maxWidth: 120 }} />
            </div>
          )}
          <label>
            使用说明（Markdown）
            <textarea
              rows={8}
              value={draft.readme || ''}
              onChange={(e) => patch('readme', e.target.value)}
              placeholder='支持 Markdown 格式'
            />
          </label>
          <label className='fx-tpl-checkline'>
            <input
              type='checkbox'
              checked={Number(draft.disabled) === 1}
              onChange={(e) => patch('disabled', e.target.checked ? 1 : 0)}
            />
            禁用此组件
          </label>
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
