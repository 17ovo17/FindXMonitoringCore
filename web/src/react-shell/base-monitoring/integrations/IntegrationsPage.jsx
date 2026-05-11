import React, { useEffect, useMemo, useState } from 'react'
import { DatasourcePage } from '../datasource/DatasourcePage.jsx'
import { integrationsApi } from '../../api/integrations.js'
import { orgApi } from '../../api/org.js'
import {
  blockedContracts,
  blankComponentDraft,
  blankPayloadDraft,
  buildExportBody,
  buildFallbackComponents,
  blankSystemIntegrationDraft,
  filterPayloads,
  filterSystemIntegrations,
  formatContentForPreview,
  isWritablePayloadType,
  normalizeComponents,
  normalizeDashboardTemplateFallback,
  normalizePayloads,
  normalizeSystemIntegrations,
  parseComponentDraft,
  parsePayloadDraft,
  parseSystemIntegrationDraft,
  safeErrorText,
  safeJson,
  systemIntegrationOrderPayload,
} from './integrationModel.js'
import {
  ComponentCard,
  ComponentDrawer,
  ComponentFormModal,
  ImportModal,
  PayloadFormModal,
  PreviewModal,
  SystemsSection,
} from './IntegrationParts.jsx'
import './integrations.css'

const validSections = new Set(['datasources', 'templates', 'systems'])
const validTabs = new Set(['instructions', 'collect', 'metric', 'dashboard', 'alert', 'record'])

function asSection(value) {
  return validSections.has(value) ? value : 'datasources'
}

function asTab(value) {
  return validTabs.has(value) ? value : 'instructions'
}

function TemplatesSection({ query, onNavigate }) {
  const [components, setComponents] = useState([])
  const [componentLoading, setComponentLoading] = useState(false)
  const [componentError, setComponentError] = useState('')
  const [fallbackPayloads, setFallbackPayloads] = useState([])
  const [payloads, setPayloads] = useState([])
  const [payloadLoading, setPayloadLoading] = useState(false)
  const [payloadError, setPayloadError] = useState('')
  const [notice, setNotice] = useState('')
  const [searchValue, setSearchValue] = useState('')
  const [activeComponent, setActiveComponent] = useState(null)
  const [activeTab, setActiveTab] = useState(asTab(query.tab))
  const [payloadQuery, setPayloadQuery] = useState('')
  const [selectedIds, setSelectedIds] = useState(new Set())
  const [previewRow, setPreviewRow] = useState(null)
  const [previewContent, setPreviewContent] = useState('')
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewError, setPreviewError] = useState('')
  const [importRows, setImportRows] = useState([])
  const [importDraft, setImportDraft] = useState(blankImportDraft())
  const [importError, setImportError] = useState('')
  const [importResult, setImportResult] = useState(null)
  const [importSubmitting, setImportSubmitting] = useState(false)
  const [businessGroups, setBusinessGroups] = useState([])
  const [businessLoading, setBusinessLoading] = useState(false)
  const [businessError, setBusinessError] = useState('')
  const [componentForm, setComponentForm] = useState(null)
  const [componentSaving, setComponentSaving] = useState(false)
  const [payloadForm, setPayloadForm] = useState(null)
  const [payloadSaving, setPayloadSaving] = useState(false)

  const visibleComponents = useMemo(() => {
    const text = searchValue.trim().toLowerCase()
    return components.filter((item) => !text || item.ident.toLowerCase().includes(text))
  }, [components, searchValue])

  const visiblePayloads = useMemo(
    () => filterPayloads(payloads, payloadQuery, activeComponent),
    [payloads, payloadQuery, activeComponent],
  )

  const navigate = (patch = {}) => {
    const next = { section: 'templates', ...patch }
    if (next.component === undefined) delete next.component
    if (next.tab === undefined) delete next.tab
    onNavigate?.(next)
  }

  const loadFallbackTemplates = async () => {
    const rows = normalizeDashboardTemplateFallback(await integrationsApi.listDashboardTemplates())
    setFallbackPayloads(rows)
    return rows
  }

  const loadComponents = async (options = {}) => {
    setComponentLoading(true)
    setComponentError('')
    if (!options.keepNotice) setNotice('')
    try {
      const result = await integrationsApi.listBuiltinComponents()
      const rows = normalizeComponents(result.data)
      if (rows.length > 0) {
        setComponents(rows)
        return rows
      }
      const fallbackRows = await loadFallbackTemplates().catch(() => [])
      const fallbackComponents = buildFallbackComponents(fallbackRows)
      setComponents(fallbackComponents)
      setComponentError(blockedContracts.components)
      return fallbackComponents
    } catch (err) {
      const fallbackRows = await loadFallbackTemplates().catch(() => [])
      const fallbackComponents = buildFallbackComponents(fallbackRows)
      setComponents(fallbackComponents)
      setComponentError(`${blockedContracts.components} ${safeErrorText(err, '组件契约不可用')}`)
      return fallbackComponents
    } finally {
      setComponentLoading(false)
    }
  }

  const loadPayloads = async (component, tab, q = payloadQuery) => {
    if (!component) return
    setPayloadLoading(true)
    setPayloadError('')
    try {
      if (component.contractSource === 'dashboard_fallback') {
        setPayloads(fallbackPayloads)
        setPayloadError(blockedContracts.components)
        return
      }
      const result = await integrationsApi.listBuiltinPayloads({
        component_id: component.id,
        type: tab,
        query: q,
      })
      const rows = normalizePayloads(result.data, tab)
      setPayloads(rows)
      return rows
    } catch (err) {
      setPayloads([])
      setPayloadError(`${blockedContracts[tab] || blockedContracts.payloadContent} ${safeErrorText(err, 'payload 契约不可用')}`)
      return []
    } finally {
      setPayloadLoading(false)
    }
  }

  useEffect(() => { loadComponents() }, [])

  useEffect(() => {
    if (components.length === 0) return
    const wanted = String(query.component || '')
    if (!wanted) {
      if (activeComponent) setActiveComponent(null)
      return
    }
    const matched = components.find((item) => item.ident === wanted)
    if (matched && matched.ident !== activeComponent?.ident) setActiveComponent(matched)
    if (!matched && activeComponent) setActiveComponent(null)
  }, [components, query.component])

  useEffect(() => {
    setActiveTab(asTab(query.tab))
  }, [query.tab])

  useEffect(() => {
    if (activeComponent && activeTab !== 'instructions') {
      loadPayloads(activeComponent, activeTab)
    }
    if (activeTab === 'instructions') {
      setPayloads([])
      setPayloadError('')
    }
  }, [activeComponent?.id, activeComponent?.contractSource, activeTab])

  const openComponent = (item) => {
    setActiveComponent(item)
    setPayloadQuery('')
    setSelectedIds(new Set())
    navigate({ component: item.ident, tab: activeTab })
  }

  const changeTab = (tab) => {
    setActiveTab(tab)
    setSelectedIds(new Set())
    if (activeComponent) navigate({ component: activeComponent.ident, tab })
  }

  const showBlocked = (key) => {
    setNotice(blockedContracts[key] || blockedContracts.componentEdit)
  }

  const openComponentCreate = () => {
    setNotice('')
    setComponentForm({ mode: 'create', draft: blankComponentDraft(), error: '' })
  }

  const openComponentEdit = (row) => {
    if (!row || row.contractSource === 'dashboard_fallback' || row.protected) {
      showBlocked('componentEdit')
      return
    }
    setNotice('')
    setComponentForm({ mode: 'edit', row, draft: blankComponentDraft(row), error: '' })
  }

  const submitComponentForm = async () => {
    if (!componentForm) return
    const editing = componentForm.mode === 'edit'
    const parsed = parseComponentDraft(componentForm.draft, editing)
    if (parsed.error) {
      setComponentForm({ ...componentForm, error: parsed.error })
      return
    }
    setComponentSaving(true)
    try {
      const response = editing
        ? await integrationsApi.updateBuiltinComponent(parsed.payload)
        : await integrationsApi.createBuiltinComponents([parsed.payload])
      const [saved] = normalizeComponents(Array.isArray(response) ? response : [response])
      const reloaded = await loadComponents({ keepNotice: true })
      const confirmed = reloaded.find((item) => item.id === saved?.id || item.ident === saved?.ident)
      if (!saved?.id || !confirmed) throw new Error('component save confirmation missing')
      setComponentForm(null)
      setActiveComponent(confirmed)
      navigate({ component: confirmed.ident, tab: activeTab })
      setNotice(editing ? '组件已更新，并已从后端重新读取确认。' : '组件已创建，并已从后端重新读取确认。')
    } catch (err) {
      setComponentForm({ ...componentForm, error: safeErrorText(err, '组件保存失败') })
    } finally {
      setComponentSaving(false)
    }
  }

  const deleteComponent = async (row) => {
    if (!row || row.contractSource === 'dashboard_fallback' || row.protected) {
      showBlocked('componentDelete')
      return
    }
    if (typeof window !== 'undefined' && !window.confirm(`确认删除组件“${row.ident}”？`)) return
    setComponentSaving(true)
    try {
      await integrationsApi.deleteBuiltinComponents([row.id])
      const reloaded = await loadComponents({ keepNotice: true })
      if (reloaded.some((item) => item.id === row.id)) throw new Error('component delete confirmation missing')
      if (activeComponent?.id === row.id) {
        setActiveComponent(null)
        navigate()
      }
      setNotice('组件已删除，并已从后端重新读取确认。')
    } catch (err) {
      setNotice(safeErrorText(err, '组件删除失败'))
    } finally {
      setComponentSaving(false)
    }
  }

  const openPayloadCreate = () => {
    if (!activeComponent) return
    if (!isWritablePayloadType(activeTab)) {
      showBlocked('payloadCreate')
      return
    }
    setNotice('')
    setPayloadForm({ mode: 'create', draft: blankPayloadDraft(null, activeComponent, activeTab), error: '' })
  }

  const openPayloadEdit = (row) => {
    if (!row || !isWritablePayloadType(row.type) || row.protected || row.fallbackOnly) {
      showBlocked(isWritablePayloadType(row?.type) ? 'payloadEdit' : 'payloadCreate')
      return
    }
    setNotice('')
    setPayloadForm({ mode: 'edit', row, draft: blankPayloadDraft(row, activeComponent, activeTab), error: '' })
  }

  const submitPayloadForm = async () => {
    if (!payloadForm || !activeComponent) return
    const editing = payloadForm.mode === 'edit'
    const parsed = parsePayloadDraft(payloadForm.draft, editing)
    if (parsed.error) {
      setPayloadForm({ ...payloadForm, error: parsed.error })
      return
    }
    setPayloadSaving(true)
    try {
      const response = editing
        ? await integrationsApi.updateBuiltinPayload(parsed.payload)
        : await integrationsApi.createBuiltinPayloads([parsed.payload])
      const [saved] = normalizePayloads(Array.isArray(response) ? response : [response], parsed.payload.type)
      const reloaded = await loadPayloads(activeComponent, activeTab, payloadQuery)
      if (!saved?.id || !reloaded.some((item) => item.id === saved.id)) throw new Error('payload save confirmation missing')
      setPayloadForm(null)
      setNotice(editing ? 'payload 已更新，并已从后端重新读取确认。' : 'payload 已创建，并已从后端重新读取确认。')
    } catch (err) {
      setPayloadForm({ ...payloadForm, error: safeErrorText(err, 'payload 保存失败') })
    } finally {
      setPayloadSaving(false)
    }
  }

  const deletePayload = async (row) => {
    if (!row || !isWritablePayloadType(row.type) || row.protected || row.fallbackOnly) {
      showBlocked('payloadDelete')
      return
    }
    if (typeof window !== 'undefined' && !window.confirm(`确认删除 payload“${row.name}”？`)) return
    setPayloadSaving(true)
    try {
      await integrationsApi.deleteBuiltinPayloads([row.id])
      const reloaded = await loadPayloads(activeComponent, activeTab, payloadQuery)
      if (reloaded.some((item) => item.id === row.id)) throw new Error('payload delete confirmation missing')
      setSelectedIds((current) => {
        const next = new Set(current)
        next.delete(row.id)
        return next
      })
      setNotice('payload 已删除，并已从后端重新读取确认。')
    } catch (err) {
      setNotice(safeErrorText(err, 'payload 删除失败'))
    } finally {
      setPayloadSaving(false)
    }
  }

  const preview = async (row) => {
    setPreviewRow(row)
    setPreviewContent('')
    setPreviewError('')
    setPreviewLoading(true)
    try {
      let detail = row
      if (!row.content && !row.fallbackOnly && row.id) {
        const result = await integrationsApi.getBuiltinPayload(row.id)
        detail = normalizePayloads([result.data], row.type)[0] || row
      }
      if (!detail.content && row.fallbackOnly && row.id) {
        detail = normalizeDashboardTemplateFallback([await integrationsApi.getDashboardTemplate(row.id)])[0] || row
      }
      setPreviewContent(formatContentForPreview(detail))
      if (!detail.content) setPreviewError(blockedContracts.payloadContent)
    } catch (err) {
      setPreviewError(safeErrorText(err, 'payload 内容加载失败'))
      setPreviewContent(blockedContracts.payloadContent)
    } finally {
      setPreviewLoading(false)
    }
  }

  const exportRows = (rows) => {
    if (!rows.length) {
      setNotice('请先选择 payload')
      return
    }
    if (rows.some((item) => !item.content)) {
      setNotice(blockedContracts.payloadContent)
      return
    }
    const blob = new Blob([safeJson(buildExportBody(rows), 500000)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${activeComponent?.ident || 'dashboard'}-payload-content.json`
    link.click()
    URL.revokeObjectURL(url)
  }

  const openImport = (rows) => {
    if (!rows.length) {
      setNotice('请先选择 payload')
      return
    }
    const dashboardRows = rows.filter(isDashboardTemplateRow)
    if (dashboardRows.length !== rows.length) {
      setNotice(blockedContracts.payloadImport)
      setPreviewError(blockedContracts.payloadImport)
      return
    }
    setNotice('')
    setPreviewError('')
    setImportRows(dashboardRows)
    setImportDraft(buildImportDraft(dashboardRows))
    setImportError('')
    setImportResult(null)
    loadBusinessGroups()
  }

  const loadBusinessGroups = async () => {
    setBusinessLoading(true)
    setBusinessError('')
    try {
      const result = await orgApi.business.list({ limit: 200 })
      setBusinessGroups(normalizeBusinessGroups(result.rows))
    } catch (err) {
      setBusinessGroups([])
      setBusinessError(safeErrorText(err, 'BLOCKED_BY_CONTRACT: 业务组列表契约不可用'))
    } finally {
      setBusinessLoading(false)
    }
  }

  const submitImport = async () => {
    setImportError('')
    setImportResult(null)
    const parsed = parseImportDraft(importDraft, importRows)
    if (parsed.error) {
      setImportError(parsed.error)
      return
    }
    setImportSubmitting(true)
    try {
      const saved = []
      for (const row of importRows) {
        const response = await integrationsApi.importDashboardTemplate(templateIdForImport(row), parsed.bodyFor(row))
        saved.push(validateImportedDashboardResponse(response))
      }
      setImportResult({ dashboards: saved.map(normalizeImportedDashboard) })
      setSelectedIds(new Set())
      if (saved.length === 1) setPreviewRow(null)
    } catch (err) {
      setImportError(safeErrorText(err, '仪表盘模板导入失败'))
    } finally {
      setImportSubmitting(false)
    }
  }

  const openImportedDashboard = (dashboard) => {
    if (!dashboard?.id) return
    onNavigate?.({ path: '/dashboards', query: { section: 'detail', id: dashboard.id } })
  }

  return (
    <main className='fx-int-page'>
      <header className='fx-int-header'>
        <div>
          <p>FindX 集成中心</p>
          <h1>模板中心</h1>
          <span>按组件实体管理说明、采集模板、指标、仪表盘、告警规则和记录规则；契约缺失时仅展示降级视图，不会伪造新增、编辑、删除或导入成功。</span>
        </div>
        <div className='fx-int-actions'>
          <input value={searchValue} onChange={(event) => setSearchValue(event.target.value)} placeholder='搜索组件' />
          <button type='button' onClick={loadComponents} disabled={componentLoading}>{componentLoading ? '刷新中...' : '刷新'}</button>
          <button type='button' className='is-primary' onClick={openComponentCreate} disabled={componentSaving}>新增组件</button>
        </div>
      </header>
      {componentError && <div className='fx-int-alert is-warning'><strong>BLOCKED_BY_CONTRACT</strong><span>{componentError.replace(/^BLOCKED_BY_CONTRACT[：:]\s*/, '')}</span></div>}
      {notice && <div className='fx-int-alert is-warning'>{notice}</div>}
      <section className='fx-int-grid'>
        {visibleComponents.map((item) => (
          <ComponentCard
            key={`${item.contractSource}:${item.id}`}
            item={item}
            selected={item.ident === activeComponent?.ident}
            href={`/integrations?section=templates&component=${encodeURIComponent(item.ident)}&tab=${encodeURIComponent(activeTab)}`}
            saving={componentSaving}
            onOpen={openComponent}
            onEdit={openComponentEdit}
            onDelete={deleteComponent}
          />
        ))}
        {componentLoading && <div className='fx-int-empty'>组件加载中...</div>}
        {!componentLoading && visibleComponents.length === 0 && <div className='fx-int-empty'>暂无组件实体</div>}
      </section>
      <ComponentDrawer
        component={activeComponent}
        activeTab={activeTab}
        payloads={visiblePayloads}
        payloadLoading={payloadLoading}
        payloadSaving={payloadSaving}
        payloadError={payloadError}
        query={payloadQuery}
        selectedIds={selectedIds}
        onClose={() => {
          setActiveComponent(null)
          setPayloadQuery('')
          setSelectedIds(new Set())
          navigate()
        }}
        onTab={changeTab}
        onQuery={(value) => {
          setPayloadQuery(value)
          if (activeComponent && activeTab !== 'instructions') loadPayloads(activeComponent, activeTab, value)
        }}
        onSelect={(ids) => setSelectedIds(new Set(ids))}
        onPreview={preview}
        onPayloadCreate={openPayloadCreate}
        onPayloadEdit={openPayloadEdit}
        onPayloadDelete={deletePayload}
        onComponentEdit={openComponentEdit}
        onImport={openImport}
        onExport={exportRows}
        onBlocked={showBlocked}
        onTag={(tag) => setPayloadQuery((value) => (value ? `${value} ${tag}` : tag))}
      />
      <ComponentFormModal
        state={componentForm}
        saving={componentSaving}
        onDraft={(draft) => setComponentForm((current) => current ? { ...current, draft, error: '' } : current)}
        onSubmit={submitComponentForm}
        onClose={() => {
          if (!componentSaving) setComponentForm(null)
        }}
      />
      <PayloadFormModal
        state={payloadForm}
        saving={payloadSaving}
        onDraft={(draft) => setPayloadForm((current) => current ? { ...current, draft, error: '' } : current)}
        onSubmit={submitPayloadForm}
        onClose={() => {
          if (!payloadSaving) setPayloadForm(null)
        }}
      />
      <PreviewModal
        row={previewRow}
        content={previewContent}
        loading={previewLoading}
        error={previewError}
        onClose={() => setPreviewRow(null)}
        onImport={openImport}
      />
      <ImportModal
        rows={importRows}
        draft={importDraft}
        groups={businessGroups}
        groupsLoading={businessLoading}
        groupsError={businessError}
        submitting={importSubmitting}
        error={importError}
        result={importResult}
        onDraft={setImportDraft}
        onClose={() => {
          setImportRows([])
          setImportError('')
          setImportResult(null)
        }}
        onSubmit={submitImport}
        onOpenDashboard={openImportedDashboard}
      />
    </main>
  )
}

function SystemsContainer() {
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [searchValue, setSearchValue] = useState('')
  const [selected, setSelected] = useState(null)
  const [formState, setFormState] = useState(null)
  const [orderOpen, setOrderOpen] = useState(false)
  const [orderRows, setOrderRows] = useState([])
  const [orderError, setOrderError] = useState('')

  const visibleRows = useMemo(
    () => filterSystemIntegrations(rows, searchValue),
    [rows, searchValue],
  )

  const loadRows = async (options = {}) => {
    setLoading(true)
    setError('')
    if (!options.keepNotice) setNotice('')
    try {
      const result = await integrationsApi.listSystemIntegrations()
      const normalized = normalizeSystemIntegrations(result)
      setRows(normalized)
      if (selected) {
        setSelected(normalized.find((row) => row.id === selected.id) || null)
      }
      return normalized
    } catch (err) {
      setRows([])
      setSelected(null)
      setError(`${blockedContracts.systems} ${safeErrorText(err, '系统集成列表契约不可用')}`)
      return []
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { loadRows() }, [])

  const preview = async (row) => {
    setSelected(row)
    setNotice('')
    try {
      const result = await integrationsApi.getSystemIntegration(row.id)
      const [detail] = normalizeSystemIntegrations([result])
      if (detail) setSelected(detail)
    } catch (err) {
      setNotice(`${blockedContracts.systems} ${safeErrorText(err, '系统集成详情契约不可用')}`)
    }
  }

  const showBlocked = (key) => {
    setNotice(blockedContracts[key] || blockedContracts.systems)
  }

  const openCreate = () => {
    const nextWeight = Math.max(0, ...rows.map((row) => Number(row.weight) || 0)) + 10
    setNotice('')
    setFormState({ mode: 'create', draft: blankSystemIntegrationDraft(null, nextWeight), error: '' })
  }

  const openEdit = (row) => {
    setNotice('')
    setFormState({ mode: 'edit', row, draft: blankSystemIntegrationDraft(row), error: '' })
  }

  const patchFormDraft = (draft) => {
    if (!formState) return
    setFormState({ ...formState, draft, error: '' })
  }

  const closeForm = () => {
    if (!saving) setFormState(null)
  }

  const submitForm = async () => {
    if (!formState) return
    const editing = formState.mode === 'edit'
    const parsed = parseSystemIntegrationDraft(formState.draft, editing)
    if (parsed.error) {
      setFormState({ ...formState, error: parsed.error })
      return
    }
    setSaving(true)
    setNotice('')
    try {
      const response = editing
        ? await integrationsApi.updateSystemIntegration(formState.row.id, parsed.payload)
        : await integrationsApi.createSystemIntegration(parsed.payload)
      const [saved] = normalizeSystemIntegrations([response])
      const reloaded = await loadRows({ keepNotice: true })
      if (!saved?.id || !reloaded.some((row) => row.id === saved.id)) {
        throw new Error('save confirmation missing')
      }
      setSelected(null)
      setFormState(null)
      setNotice(editing ? '系统集成已更新，并已从后端重新读取确认。' : '系统集成已创建，并已从后端重新读取确认。')
    } catch (err) {
      setFormState({ ...formState, error: safeErrorText(err, '系统集成保存失败') })
    } finally {
      setSaving(false)
    }
  }

  const deleteRow = async (row) => {
    if (!row) return
    if (typeof window !== 'undefined' && !window.confirm(`确认删除系统集成「${row.name}」？`)) return
    setSaving(true)
    setNotice('')
    try {
      await integrationsApi.deleteSystemIntegration(row.id)
      const reloaded = await loadRows({ keepNotice: true })
      if (reloaded.some((item) => item.id === row.id)) {
        throw new Error('delete confirmation missing')
      }
      if (selected?.id === row.id) setSelected(null)
      setNotice('系统集成已删除，并已从后端重新读取确认。')
    } catch (err) {
      setNotice(safeErrorText(err, '系统集成删除失败'))
    } finally {
      setSaving(false)
    }
  }

  const toggleMenu = async (row) => {
    if (!row) return
    setSaving(true)
    setNotice('')
    try {
      const response = await integrationsApi.setSystemIntegrationHide(row.id, row.showInMenu)
      const [updated] = normalizeSystemIntegrations([response])
      const reloaded = await loadRows({ keepNotice: true })
      const confirmed = reloaded.find((item) => item.id === row.id)
      if (!updated?.id || !confirmed || confirmed.showInMenu === row.showInMenu) {
        throw new Error('visibility confirmation missing')
      }
      setSelected(null)
      setNotice(confirmed.showInMenu ? '系统集成已显示在菜单元数据中。' : '系统集成已从菜单元数据中隐藏。')
    } catch (err) {
      setNotice(safeErrorText(err, '菜单显示状态更新失败'))
    } finally {
      setSaving(false)
    }
  }

  const openOrder = () => {
    setOrderRows(rows.map((row) => ({ ...row, weightDraft: row.weight })))
    setOrderError('')
    setNotice('')
    setOrderOpen(true)
  }

  const changeOrder = (id, weightDraft) => {
    setOrderRows((items) => items.map((item) => item.id === id ? { ...item, weightDraft } : item))
    setOrderError('')
  }

  const closeOrder = () => {
    if (!saving) setOrderOpen(false)
  }

  const submitOrder = async () => {
    const payload = systemIntegrationOrderPayload(orderRows)
    if (payload.some((item) => !item.id || !Number.isFinite(item.weight))) {
      setOrderError('排序权重必须是数字。')
      return
    }
    setSaving(true)
    setOrderError('')
    setNotice('')
    try {
      const response = await integrationsApi.updateSystemIntegrationWeights(payload)
      const updatedRows = normalizeSystemIntegrations(response)
      const reloaded = await loadRows({ keepNotice: true })
      const confirmed = payload.every((item) => reloaded.some((row) => row.id === item.id && row.weight === item.weight))
      if (!updatedRows.length || !confirmed) {
        throw new Error('weight confirmation missing')
      }
      setOrderOpen(false)
      setNotice('系统集成排序已保存，并已从后端重新读取确认。')
    } catch (err) {
      setOrderError(safeErrorText(err, '系统集成排序保存失败'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <SystemsSection
      rows={visibleRows}
      orderRows={orderRows}
      query={searchValue}
      loading={loading}
      saving={saving}
      error={error}
      notice={notice}
      selected={selected}
      formState={formState}
      orderOpen={orderOpen}
      orderError={orderError}
      onQuery={setSearchValue}
      onReload={loadRows}
      onPreview={preview}
      onCreate={openCreate}
      onEdit={openEdit}
      onDelete={deleteRow}
      onToggleMenu={toggleMenu}
      onFormDraft={patchFormDraft}
      onFormSubmit={submitForm}
      onFormClose={closeForm}
      onOrderOpen={openOrder}
      onOrderClose={closeOrder}
      onOrderChange={changeOrder}
      onOrderSubmit={submitOrder}
      onBlocked={showBlocked}
      onClosePreview={() => setSelected(null)}
    />
  )
}

function blankImportDraft() {
  return {
    title: '',
    workspaceId: '',
    resourceGroupId: '',
    businessGroupId: '',
    tagsText: '',
    variablesText: '{}',
  }
}

function isDashboardTemplateRow(row) {
  return row?.type === 'dashboard' && Boolean(templateIdForImport(row))
}

function templateIdForImport(row) {
  if (!row || row.type !== 'dashboard') return ''
  const rawTemplate = row.raw?.template || row.raw?.template_id || row.raw?.templateId
  const id = rawTemplate || row.id || ''
  return String(id).replace(/^dashboard:/, '').trim()
}

function variablesForImport(row) {
  if (row?.raw?.variables && typeof row.raw.variables === 'object' && !Array.isArray(row.raw.variables)) {
    return row.raw.variables
  }
  if (row?.content) {
    try {
      const parsed = typeof row.content === 'string' ? JSON.parse(row.content) : row.content
      if (parsed?.variables && typeof parsed.variables === 'object' && !Array.isArray(parsed.variables)) {
        return parsed.variables
      }
    } catch {
      return {}
    }
  }
  return {}
}

function normalizeBusinessGroups(rows = []) {
  return rows.map((row) => ({
    id: String(row?.id || ''),
    name: safeErrorText({ message: row?.name || row?.title || row?.id || '' }, ''),
    parentId: String(row?.parent_id || row?.parentId || ''),
  })).filter((row) => row.id)
}

function buildImportDraft(rows) {
  const [row] = rows
  const draft = blankImportDraft()
  if (rows.length === 1) {
    draft.title = row.name || ''
    draft.tagsText = (row.tags || []).join(', ')
    draft.variablesText = safeJson(variablesForImport(row), 500000)
  }
  return draft
}

function parseImportDraft(draft, rows) {
  const businessGroupId = String(draft.businessGroupId || '').trim()
  if (!businessGroupId) {
    return { error: 'BLOCKED_BY_CONTRACT：仪表盘模板导入必须选择业务组；成熟源码要求业务组必选，当前不会提交未绑定业务组的导入请求。' }
  }
  let variables
  const text = draft.variablesText.trim()
  if (text) {
    try {
      variables = JSON.parse(text)
    } catch {
      return { error: '变量 JSON 格式不正确，请修正后再提交。' }
    }
    if (!variables || Array.isArray(variables) || typeof variables !== 'object') {
      return { error: '变量 JSON 必须是对象。' }
    }
  }
  const tags = draft.tagsText.trim()
    ? draft.tagsText.split(/[,，\s]+/).map((item) => item.trim()).filter(Boolean)
    : undefined
  const shared = {
    workspace_id: draft.workspaceId.trim(),
    resource_group_id: businessGroupId,
    ...(variables ? { variables } : {}),
    ...(tags ? { tags } : {}),
  }
  return {
    bodyFor: (row) => ({
      ...shared,
      ...(rows.length === 1 && draft.title.trim() ? { title: draft.title.trim() } : {}),
      ...(rows.length > 1 ? { title: row.name } : {}),
    }),
  }
}

function normalizeImportedDashboard(row = {}) {
  return {
    id: String(row.id || row.dashboard_id || row.dashboardId || row.uid || ''),
    title: safeErrorText({ message: row.title || row.name || '已导入仪表盘' }, '已导入仪表盘'),
  }
}

function validateImportedDashboardResponse(response) {
  if (!response || typeof response !== 'object' || Array.isArray(response)) {
    throw new Error('导入接口未返回有效仪表盘标识，未进入成功态')
  }
  if (response.error || response.err) {
    throw new Error('导入接口返回错误，未进入成功态')
  }
  if (response.ok === false || response.success === false) {
    throw new Error('导入接口返回失败状态，未进入成功态')
  }
  const dashboard = response.id || response.dashboard_id || response.dashboardId || response.uid
    ? response
    : response.data
  const normalized = normalizeImportedDashboard(dashboard)
  if (!normalized.id) {
    throw new Error('导入接口未返回有效仪表盘标识，未进入成功态')
  }
  return dashboard
}

export function IntegrationsPage({ query = {}, onNavigate }) {
  const section = asSection(query.section)
  if (section === 'datasources') return <DatasourcePage />
  if (section === 'systems') return <SystemsContainer />
  return <TemplatesSection query={query} onNavigate={onNavigate} />
}
