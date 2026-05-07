import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { dashboardsApi, normalizeDashboardList, redactText } from '../../api/dashboards'
import {
  buildDashboardPayload,
  columnOptions,
  ensureList,
  normalizeDashboard,
  normalizePanels,
  normalizeTemplate,
  normalizeVariables,
  panelTypes,
  parseObjectJson,
  toTags,
} from './dashboardModel'

export function useDashboardWorkbench() {
  const route = useRoute()
  const router = useRouter()
  const loading = ref(false), detailLoading = ref(false), templatesLoading = ref(false)
  const saving = ref(false), importing = ref(false), dashboards = ref([]), templates = ref([])
  const selectedRows = ref([]), keyword = ref(''), scopeFilter = ref('all')
  const errorText = ref(''), errorStatus = ref(0), detailError = ref(''), templatesError = ref('')
  const activeDashboardId = ref(''), detailRaw = ref(null), formVisible = ref(false)
  const previewVisible = ref(false), importVisible = ref(false), panelDrawerVisible = ref(false)
  const editingId = ref(''), tagText = ref(''), formError = ref('')
  const selectedTemplate = ref(null), previewJson = ref(''), variablesJson = ref('{}')
  const importTagText = ref(''), importError = ref(''), panelDrawerTitle = ref('Panel 配置')
  const panelDraftJson = ref(''), autoRefresh = ref('off'), timeRange = ref('1h'), timezone = ref('local')
  const variableValues = reactive({})
  const form = reactive({ name: '', ident: '', note: '', business_group: '', shared: false, graphTooltip: true, graphZoom: true })
  const importForm = reactive({ name: '', business_group: '' })
  const visibleColumns = reactive({ name: true, tags: true, note: true, updatedAt: true, updatedBy: true, share: true })

  const section = computed(() => ['list', 'detail', 'templates'].includes(String(route.query.section || 'list')) ? String(route.query.section || 'list') : 'list')
  const permissionError = computed(() => [401, 403].includes(errorStatus.value))
  const errorTitle = computed(() => permissionError.value ? '权限或登录状态异常' : '仪表盘接口请求失败')
  const activeDashboard = computed(() => normalizeDashboard(detailRaw.value || dashboards.value.find(item => item.id === activeDashboardId.value) || {}))
  const previewTitle = computed(() => selectedTemplate.value ? `预览：${selectedTemplate.value.name}` : '预览')
  const businessGroups = computed(() => {
    const map = new Map()
    dashboards.value.forEach(item => {
      const label = item.businessGroup || '未分组'
      const key = `group:${label}`
      map.set(key, { key, label, count: (map.get(key)?.count || 0) + 1 })
    })
    return Array.from(map.values())
  })
  const filteredDashboards = computed(() => dashboards.value.filter(item => {
    const words = [item.name, item.ident, item.note, item.updatedBy, item.tags.join(' ')].join(' ').toLowerCase()
    const hitKeyword = !keyword.value || words.includes(keyword.value.toLowerCase())
    const groupKey = `group:${item.businessGroup || '未分组'}`
    const hitScope = scopeFilter.value === 'all' || (scopeFilter.value === 'public' ? item.shared : groupKey === scopeFilter.value)
    return hitKeyword && hitScope
  }))
  const variables = computed(() => normalizeVariables(activeDashboard.value.raw, variableValues))
  const panels = computed(() => normalizePanels(activeDashboard.value.raw))
  const goList = () => router.push({ path: '/dashboards', query: { section: 'list' } })
  const goTemplates = () => router.push({ path: '/dashboards', query: { section: 'templates' } })
  const openDetail = id => router.push({ path: '/dashboards', query: { section: 'detail', id } })
  const refreshDetail = async () => loadDetail(activeDashboardId.value)

  const loadDashboards = async () => {
    loading.value = true
    errorText.value = ''
    errorStatus.value = 0
    try {
      dashboards.value = normalizeDashboardList(await dashboardsApi.list({ query: keyword.value })).map(normalizeDashboard).filter(item => item.id)
    } catch (error) {
      dashboards.value = []
      errorStatus.value = error.status || 0
      errorText.value = error.message || '请求失败'
    } finally {
      loading.value = false
    }
  }
  const loadTemplates = async () => {
    templatesLoading.value = true
    templatesError.value = ''
    try {
      templates.value = ensureList(await dashboardsApi.listTemplates()).map(normalizeTemplate).filter(item => item.id)
    } catch (error) {
      templates.value = []
      templatesError.value = error.message || '模板加载失败'
    } finally {
      templatesLoading.value = false
    }
  }
  const loadDetail = async id => {
    if (!id) return
    activeDashboardId.value = String(id)
    detailLoading.value = true
    detailError.value = ''
    try {
      detailRaw.value = await dashboardsApi.detail(id)
    } catch (error) {
      detailRaw.value = dashboards.value.find(item => item.id === String(id))?.raw || null
      detailError.value = `BLOCKED_BY_CONTRACT：详情接口不可用，已使用列表返回数据降级展示。${error.message || ''}`
    } finally {
      detailLoading.value = false
    }
  }
  const resetForm = () => {
    Object.assign(form, { name: '', ident: '', note: '', business_group: '', shared: false, graphTooltip: true, graphZoom: true })
    tagText.value = ''; formError.value = ''
  }
  const openCreate = () => { editingId.value = ''; resetForm(); formVisible.value = true }
  const openEdit = row => {
    editingId.value = row.id
    Object.assign(form, { name: row.name, ident: row.ident, note: row.note, business_group: row.businessGroup, shared: row.shared, graphTooltip: row.graphTooltip, graphZoom: row.graphZoom })
    tagText.value = row.tags.join(', ')
    formError.value = ''
    formVisible.value = true
  }
  const saveDashboard = async () => {
    formError.value = ''
    let payload
    try {
      payload = buildDashboardPayload({ form, tagText: tagText.value, editingId: editingId.value, dashboards: dashboards.value })
    } catch (error) {
      formError.value = error.message
      return
    }
    saving.value = true
    try {
      editingId.value ? await dashboardsApi.update(editingId.value, payload) : await dashboardsApi.create(payload)
      ElMessage.success('已保存仪表盘')
      formVisible.value = false
      await loadDashboards()
    } catch (error) {
      formError.value = error.message || '保存失败'
    } finally {
      saving.value = false
    }
  }
  const handleRowCommand = async (command, row) => {
    if (command === 'edit') openEdit(row)
    if (command === 'clone') await cloneDashboard(row)
    if (command === 'export') await exportDashboard(row)
    if (command === 'share') await shareDashboard(row)
    if (command === 'delete') await deleteDashboard(row)
  }
  const handleBatchCommand = async command => {
    if (selectedRows.value.length === 0) {
      ElMessage.warning('请先选择仪表盘')
      return
    }
    for (const row of selectedRows.value) {
      if (command === 'share') await shareDashboard(row, true)
      if (command === 'export') await exportDashboard(row, true)
      if (command === 'delete') await deleteDashboard(row, true)
    }
    if (command !== 'export') await loadDashboards()
  }
  const cloneDashboard = async row => {
    try { await dashboardsApi.clone(row.id); ElMessage.success('已克隆仪表盘'); await loadDashboards() } catch (error) { ElMessage.error(error.message || '克隆失败') }
  }
  const shareDashboard = async (row, silent = false) => {
    try {
      await dashboardsApi.share(row.id, { shared: !row.shared, public: !row.shared })
      if (!silent) ElMessage.success('共享状态已更新')
      await loadDashboards()
    } catch (error) {
      ElMessage.error(error.message || '共享状态更新失败')
    }
  }
  const deleteDashboard = async (row, skipConfirm = false) => {
    if (!skipConfirm) {
      try { await ElMessageBox.confirm(`确认删除“${row.name}”？`, '删除仪表盘', { type: 'warning' }) } catch (error) { return }
    }
    try {
      await dashboardsApi.remove(row.id)
      if (!skipConfirm) ElMessage.success('已删除仪表盘')
      await loadDashboards()
    } catch (error) {
      ElMessage.error(error.message || '删除失败')
    }
  }
  const downloadJson = (name, data) => {
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${name || 'dashboard'}.json`
    link.click()
    URL.revokeObjectURL(url)
  }
  const exportDashboard = async (row, silent = false) => {
    try {
      downloadJson(row.ident || row.name, await dashboardsApi.export(row.id))
      if (!silent) ElMessage.success('已导出仪表盘')
    } catch (error) {
      downloadJson(row.ident || row.name, { blocked: 'BLOCKED_BY_CONTRACT', reason: redactText(error.message), dashboard: row.raw })
      if (!silent) ElMessage.warning('导出接口不可用，已导出当前脱敏 JSON')
    }
  }
  const previewTemplate = async row => {
    selectedTemplate.value = row
    previewVisible.value = true
    previewJson.value = '加载中...'
    try {
      const detail = await dashboardsApi.getTemplate(row.id)
      selectedTemplate.value = normalizeTemplate({ ...row.raw, ...detail })
      previewJson.value = JSON.stringify(detail, null, 2)
    } catch (error) {
      previewJson.value = error.message || '模板详情加载失败'
    }
  }
  const openImport = row => {
    if (!row?.id) return
    selectedTemplate.value = row
    Object.assign(importForm, { name: row.name, business_group: '' })
    importTagText.value = row.tags.join(', ')
    variablesJson.value = '{}'
    importError.value = ''
    importVisible.value = true
  }
  const submitImport = async () => {
    importError.value = ''
    if (!selectedTemplate.value?.id) { importError.value = '请先选择模板'; return }
    if (!importForm.name) { importError.value = '名称不能为空'; return }
    let variablesPayload
    try { variablesPayload = parseObjectJson(variablesJson.value) } catch (error) { importError.value = error.message; return }
    importing.value = true
    try {
      await dashboardsApi.importTemplate(selectedTemplate.value.id, { name: importForm.name, title: importForm.name, business_group: importForm.business_group, workspace_id: importForm.business_group, variables: variablesPayload, tags: toTags(importTagText.value) })
      ElMessage.success('已导入为仪表盘')
      importVisible.value = false
      previewVisible.value = false
      await loadDashboards()
      goList()
    } catch (error) {
      importError.value = error.message || '导入失败'
    } finally {
      importing.value = false
    }
  }
  const openPanelEditor = type => {
    panelDrawerTitle.value = `添加图表：${panelTypes.find(item => item.value === type)?.label || type}`
    panelDraftJson.value = JSON.stringify({ type, dashboard_id: activeDashboardId.value, blocked: 'BLOCKED_BY_CONTRACT' }, null, 2)
    panelDrawerVisible.value = true
  }
  const handlePanelCommand = (command, panel) => {
    if (command === 'copy') {
      navigator.clipboard?.writeText(JSON.stringify(panel.raw, null, 2))
      ElMessage.success('已复制 Panel 配置')
      return
    }
    panelDrawerTitle.value = `${command}：${panel.title}`
    panelDraftJson.value = JSON.stringify({ action: command, blocked: 'BLOCKED_BY_CONTRACT', panel: panel.raw }, null, 2)
    panelDrawerVisible.value = true
  }
  const openSettings = () => {
    panelDrawerTitle.value = '仪表盘设置'
    panelDraftJson.value = JSON.stringify({ dashboard: activeDashboard.value.raw, blocked: 'BLOCKED_BY_CONTRACT：设置保存 contract 未完整暴露' }, null, 2)
    panelDrawerVisible.value = true
  }
  const copyDetailLink = async () => {
    const url = `${location.origin}${location.pathname}?section=detail&id=${encodeURIComponent(activeDashboardId.value)}`
    await navigator.clipboard?.writeText(url)
    ElMessage.success('详情链接已复制')
  }
  const toggleFullscreen = async () => {
    if (!document.fullscreenElement) await document.documentElement.requestFullscreen?.()
    else await document.exitFullscreen?.()
  }

  watch(() => route.query, async query => {
    if (query.section === 'detail') await loadDetail(query.id || activeDashboardId.value || dashboards.value[0]?.id)
    if (query.section === 'templates' && templates.value.length === 0) await loadTemplates()
  }, { deep: true })
  onMounted(async () => {
    await loadDashboards()
    await loadTemplates()
    if (section.value === 'detail') await nextTick(() => loadDetail(route.query.id || dashboards.value[0]?.id))
  })

  return {
    activeDashboard, activeDashboardId, autoRefresh, businessGroups, columnOptions, copyDetailLink,
    dashboards, detailError, detailLoading, editingId, errorText, errorTitle, filteredDashboards,
    form, formError, formVisible, goList, goTemplates, handleBatchCommand, handlePanelCommand,
    handleRowCommand, importError, importForm, importTagText, importVisible, importing, keyword,
    loadDashboards, loadTemplates, loading, openCreate, openDetail, openImport, openPanelEditor,
    openSettings, panelDrawerTitle, panelDrawerVisible, panelDraftJson, panelTypes, panels,
    permissionError, previewJson, previewTemplate, previewTitle, previewVisible, refreshDetail,
    saveDashboard, saving, scopeFilter, section, selectedRows, selectedTemplate, submitImport,
    tagText, templates, templatesError, templatesLoading, timeRange, timezone, toggleFullscreen,
    variableValues, variables, variablesJson, visibleColumns,
  }
}
