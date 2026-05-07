<template>
  <div class="integration-page">
    <section class="integration-panel">
      <header class="page-head">
        <div>
          <div class="kicker">集成中心</div>
          <h2>{{ currentCopy.title }}</h2>
          <p>{{ currentCopy.desc }}</p>
        </div>
        <div class="head-actions">
          <el-input v-model="searchValue" class="search" clearable placeholder="搜索组件" />
          <el-button :loading="loading" @click="loadTemplates">刷新</el-button>
          <el-button type="primary" @click="openBlocked('component-create')">新增组件</el-button>
        </div>
      </header>

      <el-alert v-if="errorText" :title="errorText" type="error" show-icon :closable="false" class="state-alert" />
      <IntegrationGrid :components="filteredComponents" @open-component="openComponent" />
      <el-empty v-if="!loading && filteredComponents.length === 0" description="暂无组件模板" />
    </section>

    <IntegrationDrawer
      v-model:visible="drawerVisible"
      v-model:active-tab="activeTab"
      v-model:query="payloadQuery"
      :component="activeComponent"
      :tabs="payloadTabs"
      :dashboard-payloads="filteredDashboardPayloads"
      @select="selectedRows = $event"
      @blocked="openBlocked"
      @batch-import="batchImport"
      @batch-export="batchExport"
      @preview="previewPayload"
      @tag-search="tagSearch"
      @row-command="handleRowCommand"
    />

    <el-dialog v-model="previewVisible" :title="previewTitle" width="760px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="类型">仪表盘</el-descriptions-item>
        <el-descriptions-item label="分类">{{ selectedPayload?.cate || '-' }}</el-descriptions-item>
        <el-descriptions-item label="说明" :span="2">{{ selectedPayload?.note || '暂无说明' }}</el-descriptions-item>
      </el-descriptions>
      <el-input :model-value="previewJson" type="textarea" :rows="15" readonly spellcheck="false" class="json-view" />
      <template #footer>
        <el-button @click="previewVisible = false">关闭</el-button>
        <el-button type="primary" @click="openImport(selectedPayload)">导入到业务组</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="导入仪表盘模板" width="720px">
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="名称" required><el-input v-model.trim="importForm.title" maxlength="120" /></el-form-item>
        <div class="form-grid">
          <el-form-item label="业务组"><el-input v-model.trim="importForm.workspace_id" maxlength="80" /></el-form-item>
          <el-form-item label="资源组"><el-input v-model.trim="importForm.resource_group_id" maxlength="80" /></el-form-item>
        </div>
        <el-form-item label="标签"><el-input v-model.trim="importTagText" placeholder="多个标签用逗号分隔" /></el-form-item>
        <el-form-item label="变量 JSON"><el-input v-model="variablesJson" type="textarea" :rows="8" spellcheck="false" placeholder="{}" /></el-form-item>
        <el-alert v-if="importError" :title="importError" type="error" show-icon :closable="false" />
      </el-form>
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" :loading="importing" @click="submitImport">导入</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="blockedVisible" title="能力阻断" size="520px">
      <el-alert :title="blockedMessage" type="warning" show-icon :closable="false" />
      <el-input :model-value="blockedJson" type="textarea" :rows="16" readonly spellcheck="false" class="json-view" />
    </el-drawer>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { integrationsApi, normalizeList, redactText } from '../api/integrations'
import IntegrationDrawer from '../components/integrations/IntegrationDrawer.vue'
import IntegrationGrid from '../components/integrations/IntegrationGrid.vue'
import { blockedContracts, buildComponents, filterPayloads, normalizeTemplates, payloadTabs, safeJson, sanitizeDisplayText, toTags } from '../components/integrations/integrationModel'

const route = useRoute()
const router = useRouter()
const section = computed(() => ['overview', 'collectors', 'templates'].includes(route.query.section) ? route.query.section : 'overview')
const copy = {
  overview: { title: '模板中心', desc: '按成熟基础监控结构管理组件、使用说明、采集模板、内置指标、仪表盘、告警规则和记录规则。' },
  collectors: { title: '采集接入', desc: '采集模板入口保留在组件抽屉内，未暴露 contract 的动作显示阻断态。' },
  templates: { title: '模板导入', desc: '模板列表、预览、导入、导出按组件抽屉和 payload 表格组织。' },
}
const currentCopy = computed(() => copy[section.value] || copy.overview)
const loading = ref(false), importing = ref(false), drawerVisible = ref(false)
const previewVisible = ref(false), importVisible = ref(false), blockedVisible = ref(false)
const searchValue = ref(''), payloadQuery = ref(''), activeTab = ref('instructions'), errorText = ref('')
const templates = ref([]), selectedRows = ref([]), selectedPayload = ref(null), previewJson = ref('')
const importTagText = ref(''), variablesJson = ref('{}'), importError = ref('')
const blockedMessage = ref(''), blockedJson = ref('{}'), activeComponent = ref(null)
const importForm = reactive({ title: '', workspace_id: '', resource_group_id: '' })

const components = computed(() => buildComponents(templates.value))
const filteredComponents = computed(() => components.value.filter(item => !searchValue.value || item.ident.toLowerCase().includes(searchValue.value.toLowerCase())))
const filteredDashboardPayloads = computed(() => filterPayloads(templates.value, payloadQuery.value, activeComponent.value?.ident))
const previewTitle = computed(() => selectedPayload.value ? `预览：${selectedPayload.value.name}` : '预览')
const safeDisplayJson = value => sanitizeDisplayText(safeJson(value))

const loadTemplates = async () => {
  loading.value = true
  errorText.value = ''
  try {
    templates.value = normalizeTemplates(await integrationsApi.listDashboardTemplates())
    if (!activeComponent.value && components.value.length > 0 && section.value === 'templates') openComponent(components.value[0], 'dashboard')
  } catch (error) {
    templates.value = []
    errorText.value = error.message || '模板列表加载失败'
  } finally {
    loading.value = false
  }
}

const openComponent = (item, tab = section.value === 'collectors' ? 'collect' : 'instructions') => {
  activeComponent.value = item
  activeTab.value = tab
  payloadQuery.value = ''
  drawerVisible.value = true
}

const openBlocked = action => {
  const message = blockedContracts[action] || 'BLOCKED_BY_CONTRACT：组件新增、编辑、删除、使用说明保存和非仪表盘 payload contract 未完整暴露。'
  blockedMessage.value = message
  blockedJson.value = safeJson({ action, component: activeComponent.value?.ident, reason: message })
  blockedVisible.value = true
}

const previewPayload = async row => {
  selectedPayload.value = row
  previewVisible.value = true
  previewJson.value = '加载中...'
  try {
    const detail = await integrationsApi.getDashboardTemplate(row.id)
    previewJson.value = safeDisplayJson(detail)
  } catch (error) {
    previewJson.value = error.message || '模板详情加载失败'
  }
}

const openImport = row => {
  if (!row?.id) return
  selectedPayload.value = row
  Object.assign(importForm, { title: row.name, workspace_id: '', resource_group_id: '' })
  importTagText.value = row.tags.join(', ')
  variablesJson.value = '{}'
  importError.value = ''
  importVisible.value = true
}

const submitImport = async () => {
  importError.value = ''
  if (!selectedPayload.value?.id) { importError.value = '请先选择模板'; return }
  let variables
  try {
    variables = variablesJson.value.trim() ? JSON.parse(variablesJson.value) : {}
    if (!variables || Array.isArray(variables) || typeof variables !== 'object') throw new Error('变量 JSON 必须是对象')
  } catch (error) {
    importError.value = error.message || '变量 JSON 不是合法对象'
    return
  }
  importing.value = true
  try {
    const out = await integrationsApi.importDashboardTemplate(selectedPayload.value.id, { ...importForm, tags: toTags(importTagText.value), variables })
    ElMessage.success('已导入仪表盘模板')
    importVisible.value = false
    previewVisible.value = false
    if (out?.id) router.push({ path: '/dashboards', query: { section: 'detail', id: out.id } })
  } catch (error) {
    importError.value = redactText(error.message || '导入失败')
  } finally {
    importing.value = false
  }
}

const downloadJson = (name, rows) => {
  const blob = new Blob([safeDisplayJson(normalizeList(rows).map(item => item.raw || item))], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${name || 'templates'}.json`
  link.click()
  URL.revokeObjectURL(url)
}

const batchImport = () => {
  if (selectedRows.value.length !== 1) {
    ElMessage.warning(selectedRows.value.length === 0 ? '请先选择模板' : '当前导入 contract 仅暴露单模板导入')
    return
  }
  openImport(selectedRows.value[0])
}

const batchExport = () => {
  if (selectedRows.value.length === 0) { ElMessage.warning('请先选择模板'); return }
  downloadJson(`${activeComponent.value?.ident || 'dashboard'}-templates`, selectedRows.value)
}

const tagSearch = tag => { payloadQuery.value = payloadQuery.value ? `${payloadQuery.value} ${tag}` : tag }
const handleRowCommand = (cmd, row) => {
  if (cmd === 'import') openImport(row)
  else if (cmd === 'export') downloadJson(row.name, [row])
  else openBlocked(cmd)
}

watch(section, value => {
  if (value === 'templates' && components.value.length > 0) openComponent(components.value[0], 'dashboard')
  if (value === 'collectors' && components.value.length > 0) openComponent(components.value[0], 'collect')
})

onMounted(loadTemplates)
</script>

<style scoped>
.integration-page { min-height: 100%; padding: 18px; color: #25324a; background: #f5f7fb; }
.integration-panel { min-height: calc(100vh - 104px); padding: 16px; border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 16px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2, p { margin: 0; }
h2 { margin-top: 4px; color: #17233c; font-size: 22px; }
p { margin-top: 8px; color: #66758d; font-size: 13px; line-height: 1.6; }
.head-actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.search { width: 260px; }
.state-alert { margin-bottom: 14px; }
.json-view { margin-top: 12px; }
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 14px; }
@media (max-width: 760px) {
  .page-head { align-items: flex-start; flex-direction: column; }
  .search { width: 100%; }
  .form-grid { grid-template-columns: 1fr; }
}
</style>
