import React, { useEffect, useMemo, useState } from 'react'
import { dashboardsApi } from '../../api/dashboards.js'
import {
  BLOCKED_BY_CONTRACT,
  blockedContracts,
  dashboardExportPayload,
  dashboardPayload,
  displayJson,
  displayText,
  downloadJson,
  normalizeDashboard,
  normalizePanels,
  normalizeTemplate,
  normalizeVariables,
  toTags,
} from './dashboardModel.js'
import PanelChart from './PanelChart.jsx'
import './dashboards.css'

const defaultDraft = {
  title: '',
  description: '',
  workspaceId: '',
  resourceGroupId: '',
  tags: '',
  variables: {},
  panels: [],
  status: 'active',
}

const makeError = (error, fallback = '请求失败') => {
  if (error?.status === 401) return '登录已过期，请重新登录。'
  if (error?.status === 403) return '没有仪表盘权限。'
  if (error?.status >= 500) return `HTTP ${error.status}: 仪表盘服务异常。`
  return displayText(error?.message || fallback)
}

function Toolbar({ keyword, setKeyword, onRefresh, onCreate, onTemplates, selectedCount, onBatch, loading }) {
  return (
    <header className='fx-dash-toolbar'>
      <div>
        <p>仪表盘</p>
        <h1>仪表盘列表</h1>
      </div>
      <div className='fx-dash-actions'>
        <input value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder='搜索名称、标签、说明' />
        <button type='button' disabled={loading} onClick={onRefresh}>{loading ? '刷新中...' : '刷新'}</button>
        <button type='button' className='is-primary' onClick={onCreate}>新建</button>
        <button type='button' onClick={onTemplates}>导入</button>
        <select disabled={selectedCount === 0} onChange={(event) => { if (event.target.value) onBatch(event.target.value); event.target.value = '' }}>
          <option value=''>批量操作</option>
          <option value='share'>公开配置</option>
          <option value='export'>导出</option>
          <option value='delete'>删除</option>
        </select>
      </div>
    </header>
  )
}

function Sidebar({ groups, scope, setScope }) {
  return (
    <aside className='fx-dash-sidebar'>
      <button type='button' className={scope === 'all' ? 'is-active' : ''} onClick={() => setScope('all')}>全部仪表盘</button>
      <button type='button' className={scope === 'public' ? 'is-active' : ''} onClick={() => setScope('public')}>公开仪表盘</button>
      <strong>业务组</strong>
      {groups.map((group) => (
        <button type='button' key={group.key} className={scope === group.key ? 'is-active' : ''} onClick={() => setScope(group.key)}>
          <span>{group.label}</span><small>{group.count}</small>
        </button>
      ))}
    </aside>
  )
}

function DashboardList({ rows, selected, setSelected, onOpen, onRowAction }) {
  return (
    <div className='fx-dash-table'>
      <table>
        <thead>
          <tr><th><input type='checkbox' checked={rows.length > 0 && selected.length === rows.length} onChange={(event) => setSelected(event.target.checked ? rows.map((row) => row.id) : [])} /></th><th>名称</th><th>标签</th><th>说明</th><th>更新时间</th><th>更新人</th><th>共享</th><th>操作</th></tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.id}>
              <td><input type='checkbox' checked={selected.includes(row.id)} onChange={(event) => setSelected(event.target.checked ? [...selected, row.id] : selected.filter((id) => id !== row.id))} /></td>
              <td><button type='button' className='is-link' onClick={() => onOpen(row.id)}>{row.title}</button><small>{row.id}</small></td>
              <td>{row.tags.length ? row.tags.map((tag) => <span className='fx-dash-tag' key={tag}>{tag}</span>) : <span className='muted'>无</span>}</td>
              <td>{row.description || <span className='muted'>无</span>}</td>
              <td>{row.updatedAt || <span className='muted'>-</span>}</td>
              <td>{row.updatedBy || <span className='muted'>-</span>}</td>
              <td><span className={row.shared ? 'fx-dash-state is-on' : 'fx-dash-state'}>{row.shareText}</span></td>
              <td>
                <select onChange={(event) => { if (event.target.value) onRowAction(event.target.value, row); event.target.value = '' }}>
                  <option value=''>更多</option>
                  <option value='edit'>编辑</option>
                  <option value='clone'>克隆</option>
                  <option value='share'>分享</option>
                  <option value='export'>导出</option>
                  <option value='delete'>删除</option>
                </select>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {rows.length === 0 && <div className='fx-dash-empty'>暂无仪表盘数据</div>}
    </div>
  )
}

function DetailView({ dashboard, variables, panels, onBack, onRefresh, onBlocked, onExport, onShare, onFullscreen, detailError }) {
  const [timeRangeKey, setTimeRangeKey] = useState('1h')
  const datasourceId = 'prometheus-default'

  const timeRange = useMemo(() => {
    const now = Math.floor(Date.now() / 1000)
    const durations = { '1h': 3600, '6h': 21600, '24h': 86400 }
    const duration = durations[timeRangeKey] || 3600
    const step = Math.max(15, Math.floor(duration / 240))
    return { start: now - duration, end: now, step }
  }, [timeRangeKey])

  return (
    <main className='fx-dash-detail'>
      <header className='fx-dash-detail__head'>
        <div><button type='button' onClick={onBack}>返回列表</button><h1>{dashboard.title}</h1><p>{dashboard.description || '无说明'}</p></div>
        <div className='fx-dash-actions'><button type='button' onClick={onRefresh}>刷新</button><button type='button' onClick={onShare}>分享</button><button type='button' onClick={onExport}>导出</button><button type='button' onClick={onFullscreen}>放大</button><button type='button' onClick={() => onBlocked('设置', blockedContracts.panelEditor)}>设置</button></div>
      </header>
      {detailError && <div className='fx-dash-alert is-warning'>{detailError}</div>}
      <section className='fx-dash-runtime'>
        <label><span>时间范围</span><select value={timeRangeKey} onChange={(e) => setTimeRangeKey(e.target.value)}><option value='1h'>近 1 小时</option><option value='6h'>近 6 小时</option><option value='24h'>近 24 小时</option></select></label>
        <label><span>刷新</span><select disabled value='off'><option value='off'>关闭</option><option value='30s'>30s</option><option value='1m'>1m</option><option value='5m'>5m</option></select></label>
        <label><span>时区</span><select disabled value='local'><option value='local'>本地时区</option><option value='utc'>UTC</option></select></label>
        <button type='button' onClick={() => onBlocked('查询参数', blockedContracts.inspect)}>查询参数</button>
      </section>
      <section className='fx-dash-vars'>
        {variables.map((item) => (
          <label key={item.key}><span>{item.label}</span>{item.options.length ? <select defaultValue={item.value}>{item.options.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}</select> : <input defaultValue={item.value} />}{item.blocked && <em>{blockedContracts.variableOptions}</em>}</label>
        ))}
        {variables.length === 0 && <span className='muted'>暂无变量</span>}
      </section>
      <div className='fx-dash-panel-actions'>{['Time series', 'Stat', 'Table', 'Text', 'Pie', 'Gauge', 'Bar chart', 'Heatmap', '导入 Panel'].map((type) => <button key={type} type='button' onClick={() => onBlocked(`添加图表：${type}`, type === '导入 Panel' ? blockedContracts.importPanel : blockedContracts.panelEditor)}>添加 {type}</button>)}</div>
      <section className='fx-dash-grid'>
        {panels.map((panel) => (
          <article className='fx-dash-panel' key={panel.id}>
            <header><strong>{panel.title}</strong><select onChange={(event) => { if (event.target.value) onBlocked(`${event.target.value}：${panel.title}`, event.target.value === '复制' ? displayJson(panel.raw) : blockedContracts.panelEditor); event.target.value = '' }}><option value=''>...</option><option value='编辑'>编辑</option><option value='复制'>复制配置</option><option value='检查'>检查</option><option value='删除'>删除</option></select></header>
            <PanelChart panel={panel} timeRange={timeRange} datasourceId={datasourceId} />
          </article>
        ))}
        {panels.length === 0 && <div className='fx-dash-empty'>暂无面板</div>}
      </section>
    </main>
  )
}

function Modal({ title, children, onClose }) {
  return <div className='fx-dash-modal'><div className='fx-dash-modal__body'><header><h2>{title}</h2><button type='button' onClick={onClose}>×</button></header>{children}</div></div>
}

function DashboardForm({ draft, setDraft, onSubmit, onClose, saving, error }) {
  return (
    <Modal title={draft.id ? '编辑仪表盘' : '新建仪表盘'} onClose={onClose}>
      <div className='fx-dash-form'>
        <label><span>名称</span><input value={draft.title} onChange={(event) => setDraft({ ...draft, title: event.target.value })} /></label>
        <label><span>说明</span><textarea value={draft.description} onChange={(event) => setDraft({ ...draft, description: event.target.value })} /></label>
        <label><span>业务组</span><input value={draft.resourceGroupId} onChange={(event) => setDraft({ ...draft, resourceGroupId: event.target.value })} /></label>
        <label><span>标签</span><input value={draft.tags} onChange={(event) => setDraft({ ...draft, tags: event.target.value })} /></label>
        {error && <div className='fx-dash-alert is-error'>{error}</div>}
        <div className='fx-dash-actions'><button type='button' onClick={onClose}>取消</button><button type='button' className='is-primary' disabled={saving} onClick={onSubmit}>{saving ? '保存中...' : '保存'}</button></div>
      </div>
    </Modal>
  )
}

function TemplatesView({ templates, onBack, onPreview, onImport, loading, error }) {
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

function DashboardOverlays({ draft, setDraft, saving, error, onSubmit, onCloseDraft, modal, onCloseModal }) {
  return (
    <>
      {draft && <DashboardForm draft={draft} setDraft={setDraft} saving={saving} error={error} onSubmit={onSubmit} onClose={onCloseDraft} />}
      {modal && <Modal title={modal.title} onClose={onCloseModal}><pre>{displayText(modal.body)}</pre></Modal>}
    </>
  )
}

export function DashboardsPage({ query = {}, onNavigate }) {
  const [dashboards, setDashboards] = useState([])
  const [templates, setTemplates] = useState([])
  const [keyword, setKeyword] = useState('')
  const [scope, setScope] = useState('all')
  const [selected, setSelected] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [detailError, setDetailError] = useState('')
  const [draft, setDraft] = useState(null)
  const [saving, setSaving] = useState(false)
  const [modal, setModal] = useState(null)
  const section = query.section === 'detail' || query.section === 'templates' ? query.section : 'list'
  const active = useMemo(() => dashboards.find((row) => row.id === String(query.id)) || dashboards[0], [dashboards, query.id])
  const groups = useMemo(() => Object.values(dashboards.reduce((acc, row) => {
    const label = row.resourceGroupId || '未分组'
    const key = `group:${label}`
    acc[key] = { key, label, count: (acc[key]?.count || 0) + 1 }
    return acc
  }, {})), [dashboards])
  const filtered = useMemo(() => dashboards.filter((row) => {
    const text = [row.title, row.description, row.tags.join(' '), row.updatedBy].join(' ').toLowerCase()
    const byKeyword = !keyword || text.includes(keyword.toLowerCase())
    const byScope = scope === 'all' || (scope === 'public' ? row.shared : `group:${row.resourceGroupId || '未分组'}` === scope)
    return byKeyword && byScope
  }), [dashboards, keyword, scope])
  const variables = useMemo(() => normalizeVariables(active?.variables), [active])
  const panels = useMemo(() => normalizePanels(active?.panels), [active])

  const loadDashboards = async () => {
    setLoading(true); setError('')
    try { setDashboards((await dashboardsApi.list()).map(normalizeDashboard).filter((row) => row.id)) } catch (err) { setError(makeError(err, '仪表盘列表加载失败')) } finally { setLoading(false) }
  }
  const loadTemplates = async () => {
    try { setTemplates((await dashboardsApi.listTemplates()).map(normalizeTemplate).filter((row) => row.id)) } catch (err) { setError(makeError(err, '模板加载失败')) }
  }
  const openDetail = async (id) => { onNavigate?.({ section: 'detail', id }); setDetailError(''); try { const detail = normalizeDashboard(await dashboardsApi.detail(id)); setDashboards((rows) => [detail, ...rows.filter((row) => row.id !== detail.id)]) } catch (err) { setDetailError(makeError(err, '详情加载失败')) } }
  const openForm = (row) => setDraft(row ? { ...defaultDraft, id: row.id, title: row.title, description: row.description, resourceGroupId: row.resourceGroupId, tags: row.tags.join(', '), variables: row.variables, panels: row.panels, status: row.status } : { ...defaultDraft })
  const saveDraft = async () => { setSaving(true); setError(''); try { const body = dashboardPayload(draft); const saved = normalizeDashboard(draft.id ? await dashboardsApi.update(draft.id, body) : await dashboardsApi.create(body)); setDashboards((rows) => [saved, ...rows.filter((row) => row.id !== saved.id)]); setDraft(null) } catch (err) { setError(makeError(err, '保存失败')) } finally { setSaving(false) } }
  const rowAction = async (action, row) => {
    try {
      if (action === 'edit') return openForm(row)
      if (action === 'clone') {
        const cloned = normalizeDashboard(await dashboardsApi.clone(row.id))
        setDashboards((rows) => [cloned, ...rows])
      }
      if (action === 'share') await dashboardsApi.share(row.id)
      if (action === 'delete') await dashboardsApi.remove(row.id)
      if (action === 'export') downloadJson(`${row.title || row.id}.json`, dashboardExportPayload(row))
      if (['share', 'delete'].includes(action)) await loadDashboards()
    } catch (err) { setModal({ title: '操作失败', body: makeError(err) }) }
  }
  const batchAction = async (action) => { for (const id of selected) { const row = dashboards.find((item) => item.id === id); if (row) await rowAction(action, row) } setSelected([]) }
  const previewTemplate = (tpl) => setModal({ title: `预览：${tpl.title}`, body: displayJson({ variables: tpl.variables, panels: tpl.panels }) })
  const importTemplate = async (tpl) => { try { const saved = normalizeDashboard(await dashboardsApi.importTemplate(tpl.id, { title: tpl.title, variables: tpl.variables, tags: tpl.tags })); setDashboards((rows) => [saved, ...rows]); onNavigate?.({ section: 'detail', id: saved.id }) } catch (err) { setModal({ title: '导入失败', body: makeError(err) }) } }
  const blocked = (title, body) => setModal({ title, body })

  useEffect(() => { loadDashboards(); loadTemplates() }, [])

  if (section === 'templates') {
    return (
      <>
        <TemplatesView templates={templates} loading={loading} error={error} onBack={() => onNavigate?.({ section: 'list' })} onPreview={previewTemplate} onImport={importTemplate} />
        <DashboardOverlays draft={draft} setDraft={setDraft} saving={saving} error={error} onSubmit={saveDraft} onCloseDraft={() => setDraft(null)} modal={modal} onCloseModal={() => setModal(null)} />
      </>
    )
  }
  return (
    <main className='fx-dash-page'>
      {section === 'detail' && active ? (
        <DetailView dashboard={active} variables={variables} panels={panels} detailError={detailError} onBack={() => onNavigate?.({ section: 'list' })} onRefresh={() => openDetail(active.id)} onExport={() => rowAction('export', active)} onShare={() => rowAction('share', active)} onFullscreen={() => document.documentElement.requestFullscreen?.()} onBlocked={blocked} />
      ) : (
        <div className='fx-dash-layout'>
          <Sidebar groups={groups} scope={scope} setScope={setScope} />
          <section className='fx-dash-main'>
            <Toolbar keyword={keyword} setKeyword={setKeyword} loading={loading} selectedCount={selected.length} onRefresh={loadDashboards} onCreate={() => openForm(null)} onTemplates={() => onNavigate?.({ section: 'templates' })} onBatch={batchAction} />
            {error && <div className='fx-dash-alert is-error'>{error}</div>}
            <DashboardList rows={filtered} selected={selected} setSelected={setSelected} onOpen={openDetail} onRowAction={rowAction} />
          </section>
        </div>
      )}
      <DashboardOverlays draft={draft} setDraft={setDraft} saving={saving} error={error} onSubmit={saveDraft} onCloseDraft={() => setDraft(null)} modal={modal} onCloseModal={() => setModal(null)} />
    </main>
  )
}
