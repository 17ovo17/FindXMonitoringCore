import React, { useEffect, useMemo, useState } from 'react'
import { displayText } from './dashboardModel.js'

export default function TemplatesView({ templates, onBack, onPreview, onImport, loading, error }) {
  const [keyword, setKeyword] = useState('')
  const [activeTag, setActiveTag] = useState('全部')
  const [selectedId, setSelectedId] = useState('')
  const tags = useMemo(() => {
    const values = templates.flatMap((tpl) => tpl.tags || []).filter(Boolean)
    return ['全部', ...Array.from(new Set(values)).slice(0, 12)]
  }, [templates])
  const filtered = useMemo(() => templates.filter((tpl) => {
    const words = [tpl.title, tpl.description, (tpl.tags || []).join(' ')].join(' ').toLowerCase()
    const byKeyword = !keyword || words.includes(keyword.trim().toLowerCase())
    const byTag = activeTag === '全部' || (tpl.tags || []).includes(activeTag)
    return byKeyword && byTag
  }), [templates, keyword, activeTag])
  const selectedTemplate = useMemo(
    () => filtered.find((tpl) => tpl.id === selectedId) || filtered[0] || null,
    [filtered, selectedId],
  )
  const variableCount = selectedTemplate ? Object.keys(selectedTemplate.variables || {}).length : 0
  const previewPanels = selectedTemplate?.panels || []

  useEffect(() => {
    if (!selectedId || !filtered.some((tpl) => tpl.id === selectedId)) {
      setSelectedId(filtered[0]?.id || '')
    }
  }, [filtered, selectedId])

  return (
    <main className='fx-dash-templates'>
      <header className='fx-dash-template-head'>
        <div>
          <p>导入</p>
          <h1>仪表盘导入</h1>
          <span>按模板预览变量、Panel 和标签后导入到仪表盘，导入失败会保留错误态。</span>
        </div>
        <button type='button' onClick={onBack}>返回列表</button>
      </header>
      <section className='fx-dash-template-filter'>
        <input value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder='搜索模板名称、标签、说明' />
        <div className='fx-dash-template-tags'>
          {tags.map((tag) => <button type='button' key={tag} className={activeTag === tag ? 'is-active' : ''} onClick={() => setActiveTag(tag)}>{tag}</button>)}
        </div>
      </section>
      {error && <div className='fx-dash-alert is-error'>{error}</div>}
      {loading && <div className='fx-dash-empty'>加载中...</div>}
      <section className='fx-dash-template-workbench'>
        <div className='fx-dash-template-grid' aria-label='仪表盘模板列表'>
          {filtered.map((tpl) => (
            <article className={`fx-dash-template-card${selectedTemplate?.id === tpl.id ? ' is-selected' : ''}`} key={tpl.id}>
              <button type='button' className='fx-dash-template-card__pick' onClick={() => setSelectedId(tpl.id)}>
                <header>
                  <div>
                    <strong>{tpl.title}</strong>
                    <p>{tpl.description || '无说明'}</p>
                  </div>
                  <span className='fx-dash-template-count'>{tpl.panelCount} Panel</span>
                </header>
                <div className='fx-dash-template-meta'>
                  <span>{Object.keys(tpl.variables || {}).length} 变量</span>
                  <span>{(tpl.tags || []).length || 0} 标签</span>
                </div>
                <div className='fx-dash-template-card__tags'>
                  {(tpl.tags || []).length ? tpl.tags.map((tag) => <span className='fx-dash-tag' key={tag}>{tag}</span>) : <span className='muted'>无标签</span>}
                </div>
              </button>
              <footer>
                <button type='button' onClick={() => onPreview(tpl)}>JSON 预览</button>
                <button type='button' className='is-primary' onClick={() => onImport(tpl)}>导入</button>
              </footer>
            </article>
          ))}
        </div>
        <aside className='fx-dash-template-preview' aria-label='模板预览'>
          {selectedTemplate ? (
            <>
              <header>
                <div>
                  <p>预览</p>
                  <h2>{selectedTemplate.title}</h2>
                  <span>{selectedTemplate.description || '无说明'}</span>
                </div>
                <button type='button' className='is-primary' onClick={() => onImport(selectedTemplate)}>导入</button>
              </header>
              <div className='fx-dash-template-summary'>
                <span><b>{selectedTemplate.panelCount}</b><small>Panel</small></span>
                <span><b>{variableCount}</b><small>变量</small></span>
                <span><b>{(selectedTemplate.tags || []).length}</b><small>标签</small></span>
              </div>
              <section>
                <h3>标签</h3>
                <div className='fx-dash-template-card__tags'>
                  {(selectedTemplate.tags || []).length ? selectedTemplate.tags.map((tag) => <span className='fx-dash-tag' key={tag}>{tag}</span>) : <span className='muted'>无标签</span>}
                </div>
              </section>
              <section>
                <h3>Panel</h3>
                <div className='fx-dash-template-panel-list'>
                  {previewPanels.slice(0, 6).map((panel, index) => (
                    <span key={panel.id || panel.title || index}>
                      <b>{displayText(panel.title || panel.name || `Panel ${index + 1}`)}</b>
                      <small>{displayText(panel.type || panel.chartType || 'panel')}</small>
                    </span>
                  ))}
                  {previewPanels.length === 0 && <span className='muted'>暂无 Panel</span>}
                  {previewPanels.length > 6 && <span className='muted'>另有 {previewPanels.length - 6} 个 Panel</span>}
                </div>
              </section>
              <button type='button' onClick={() => onPreview(selectedTemplate)}>查看脱敏 JSON</button>
            </>
          ) : (
            <div className='fx-dash-empty'>选择模板后查看预览</div>
          )}
        </aside>
      </section>
      {!loading && filtered.length === 0 && <div className='fx-dash-empty'>暂无匹配模板</div>}
    </main>
  )
}