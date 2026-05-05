<template>
  <div class="dash-page">
    <section class="dash-shell">
      <header class="dash-head">
        <div>
          <div class="kicker">监控仪表盘</div>
          <h2>仪表盘管理</h2>
          <p>基于真实监控接口维护业务看板、资源分组、面板配置和共享状态。</p>
        </div>
        <div class="head-actions">
          <el-button :loading="templatesLoading" @click="loadTemplates">刷新模板</el-button>
          <el-button :loading="loading" @click="loadDashboards">刷新</el-button>
          <el-button type="primary" @click="openCreate">新建仪表盘</el-button>
        </div>
      </header>

      <section class="template-box">
        <div class="section-title">
          <strong>内置模板</strong>
          <span>Linux / Windows / MySQL / Redis / Kubernetes</span>
        </div>
        <el-alert v-if="templatesError" :title="templatesError" type="error" show-icon :closable="false" class="state-alert" />
        <el-table v-loading="templatesLoading" :data="templates" border empty-text="后端暂未返回内置模板" class="dash-table">
          <el-table-column prop="name" label="模板" min-width="150" show-overflow-tooltip />
          <el-table-column prop="kind" label="类型" width="130" />
          <el-table-column prop="description" label="说明" min-width="220" show-overflow-tooltip />
          <el-table-column prop="panelCount" label="Panel 数" width="92" />
          <el-table-column label="标签" min-width="170">
            <template #default="{ row }">
              <el-tag v-for="tag in row.tags" :key="tag" size="small" class="tag">{{ tag }}</el-tag>
              <span v-if="row.tags.length === 0" class="muted">无</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="170" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="previewTemplate(row)">预览</el-button>
              <el-button size="small" type="primary" @click="openImport(row)">导入</el-button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <el-alert v-if="errorText" :title="errorTitle" :description="errorText" :type="permissionError ? 'warning' : 'error'" show-icon :closable="false" class="state-alert" />
      <div v-if="!loading && !errorText && dashboards.length === 0" class="empty-state">
        <el-empty description="当前接口返回的仪表盘列表为空">
          <el-button type="primary" @click="openCreate">创建第一个仪表盘</el-button>
        </el-empty>
      </div>
      <el-table v-else v-loading="loading" :data="dashboards" border class="dash-table" empty-text="暂无真实仪表盘数据">
        <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
        <el-table-column prop="description" label="描述" min-width="220" show-overflow-tooltip />
        <el-table-column prop="businessSpace" label="业务空间" min-width="120" />
        <el-table-column prop="resourceGroup" label="资源组" min-width="120" />
        <el-table-column label="标签" min-width="180">
          <template #default="{ row }"><el-tag v-for="tag in row.tags" :key="tag" size="small" class="tag">{{ tag }}</el-tag><span v-if="row.tags.length === 0" class="muted">无</span></template>
        </el-table-column>
        <el-table-column prop="panelCount" label="Panel 数" width="92" />
        <el-table-column prop="version" label="版本" width="86" />
        <el-table-column label="状态" width="96"><template #default="{ row }"><el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">{{ row.status || '未知' }}</el-tag></template></el-table-column>
        <el-table-column label="共享状态" width="108"><template #default="{ row }"><el-tag :type="row.shared ? 'primary' : 'info'" size="small">{{ row.shareText }}</el-tag></template></el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" min-width="160" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" @click="cloneDashboard(row)">克隆</el-button>
            <el-button size="small" :disabled="row.shared" @click="shareDashboard(row)">{{ row.shared ? '已共享' : '分享' }}</el-button>
            <el-button size="small" type="danger" @click="deleteDashboard(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </section>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑仪表盘' : '新建仪表盘'" width="720px">
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="标题" required><el-input v-model.trim="form.title" maxlength="80" /></el-form-item>
        <el-form-item label="描述"><el-input v-model.trim="form.description" type="textarea" :rows="2" maxlength="300" /></el-form-item>
        <div class="form-grid">
          <el-form-item label="业务空间"><el-input v-model.trim="form.workspace_id" maxlength="80" /></el-form-item>
          <el-form-item label="资源组"><el-input v-model.trim="form.resource_group_id" maxlength="80" /></el-form-item>
          <el-form-item label="状态"><el-select v-model="form.status"><el-option label="启用" value="active" /><el-option label="草稿" value="draft" /><el-option label="归档" value="archived" /></el-select></el-form-item>
          <el-form-item label="标签"><el-input v-model.trim="tagText" placeholder="多个标签用逗号分隔" /></el-form-item>
        </div>
        <el-form-item label="Panel 配置 JSON"><el-input v-model="panelJson" type="textarea" :rows="10" spellcheck="false" placeholder="[]" /></el-form-item>
        <el-alert v-if="formError" :title="formError" type="error" show-icon :closable="false" />
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="saveDashboard">保存</el-button></template>
    </el-dialog>

    <el-dialog v-model="previewVisible" :title="previewTitle" width="760px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="类型">{{ selectedTemplate?.kind || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Panel 数">{{ selectedTemplate?.panelCount ?? 0 }}</el-descriptions-item>
        <el-descriptions-item label="说明" :span="2">{{ selectedTemplate?.description || '无' }}</el-descriptions-item>
      </el-descriptions>
      <el-input :model-value="previewJson" type="textarea" :rows="14" readonly spellcheck="false" class="json-view" />
      <template #footer><el-button @click="previewVisible = false">关闭</el-button><el-button type="primary" @click="openImport(selectedTemplate)">导入</el-button></template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="导入内置模板" width="720px">
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="标题" required><el-input v-model.trim="importForm.title" maxlength="80" /></el-form-item>
        <div class="form-grid">
          <el-form-item label="业务空间"><el-input v-model.trim="importForm.workspace_id" maxlength="80" /></el-form-item>
          <el-form-item label="资源组"><el-input v-model.trim="importForm.resource_group_id" maxlength="80" /></el-form-item>
        </div>
        <el-form-item label="标签"><el-input v-model.trim="importTagText" placeholder="多个标签用逗号分隔" /></el-form-item>
        <el-form-item label="变量 JSON"><el-input v-model="variablesJson" type="textarea" :rows="8" spellcheck="false" placeholder="{}" /></el-form-item>
        <el-alert v-if="importError" :title="importError" type="error" show-icon :closable="false" />
      </el-form>
      <template #footer><el-button @click="importVisible = false">取消</el-button><el-button type="primary" :loading="importing" @click="submitImport">导入为仪表盘</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { dashboardsApi, normalizeDashboardList } from '../api/dashboards'

const loading = ref(false), saving = ref(false), templatesLoading = ref(false), importing = ref(false)
const dashboards = ref([]), templates = ref([]), errorText = ref(''), templatesError = ref(''), errorStatus = ref(0)
const dialogVisible = ref(false), previewVisible = ref(false), importVisible = ref(false)
const editingId = ref(''), tagText = ref(''), panelJson = ref('[]'), formError = ref('')
const selectedTemplate = ref(null), previewJson = ref(''), variablesJson = ref('{}'), importTagText = ref(''), importError = ref('')
const validStatuses = ['active', 'draft', 'archived']
const form = reactive({ title: '', description: '', workspace_id: '', resource_group_id: '', status: 'active' })
const importForm = reactive({ title: '', workspace_id: '', resource_group_id: '' })

const permissionError = computed(() => [401, 403].includes(errorStatus.value))
const errorTitle = computed(() => permissionError.value ? '权限或登录状态异常' : '仪表盘接口请求失败')
const previewTitle = computed(() => selectedTemplate.value ? `预览：${selectedTemplate.value.name}` : '预览')
const pick = (row, keys, fallback = '') => { for (const key of keys) if (row?.[key] !== undefined && row?.[key] !== null) return row[key]; return fallback }
const toTags = value => Array.isArray(value) ? value.filter(Boolean).map(String) : String(value || '').split(',').map(v => v.trim()).filter(Boolean)
const toPanels = row => { const panels = pick(row, ['panels', 'panel_configs', 'panelConfigs'], []); return Array.isArray(panels) ? panels : [] }
const ensureList = value => Array.isArray(value) ? value : normalizeDashboardList(value)
const parseObjectJson = text => {
  try {
    const value = text.trim() ? JSON.parse(text) : {}
    if (!value || Array.isArray(value) || typeof value !== 'object') throw new Error('变量 JSON 必须是对象')
    return value
  } catch (error) {
    throw new Error(error.message.includes('必须是对象') ? error.message : '变量 JSON 不是合法对象，请检查格式')
  }
}
const toShareState = row => {
  const raw = pick(row, ['shared', 'is_shared', 'isShared', 'share_status', 'shareStatus'], false)
  if (typeof raw === 'boolean') return { shared: raw, shareText: raw ? '已共享' : '未共享' }
  const text = String(raw || '').trim(), shared = ['shared', 'public', 'enabled', 'true', '1', '已共享'].includes(text.toLowerCase())
  return { shared, shareText: text || (shared ? '已共享' : '未共享') }
}
const normalizeRow = row => ({
  raw: row,
  id: String(pick(row, ['id', 'dashboard_id', 'dashboardId', 'uid'])),
  title: pick(row, ['title', 'name'], '未命名仪表盘'),
  description: pick(row, ['description', 'desc'], ''),
  businessSpace: pick(row, ['workspace_id', 'business_space', 'businessSpace', 'workspace'], ''),
  resourceGroup: pick(row, ['resource_group_id', 'resource_group', 'resourceGroup', 'group'], ''),
  tags: toTags(pick(row, ['tags', 'labels'], [])),
  panelCount: Number(pick(row, ['panel_count', 'panelCount'], toPanels(row).length)) || 0,
  version: pick(row, ['version'], '-'),
  status: pick(row, ['status'], ''),
  ...toShareState(row),
  updatedAt: pick(row, ['updated_at', 'updatedAt', 'update_time', 'updateTime'], ''),
})
const normalizeTemplate = row => ({
  raw: row,
  id: String(pick(row, ['id', 'template_id', 'templateId', 'uid'])),
  name: pick(row, ['name', 'title'], '未命名模板'),
  kind: pick(row, ['kind', 'type', 'category'], '-'),
  description: pick(row, ['description', 'desc'], ''),
  tags: toTags(pick(row, ['tags', 'labels'], [])),
  panelCount: Number(pick(row, ['panel_count', 'panelCount'], toPanels(row).length)) || 0,
})

const loadDashboards = async () => {
  loading.value = true; errorText.value = ''; errorStatus.value = 0
  try { dashboards.value = normalizeDashboardList(await dashboardsApi.list()).map(normalizeRow).filter(row => row.id) }
  catch (error) { dashboards.value = []; errorStatus.value = error.status || 0; errorText.value = error.message || '请求失败' }
  finally { loading.value = false }
}
const loadTemplates = async () => {
  templatesLoading.value = true; templatesError.value = ''
  try { templates.value = ensureList(await dashboardsApi.listTemplates()).map(normalizeTemplate).filter(row => row.id) }
  catch (error) { templates.value = []; templatesError.value = error.message || '模板加载失败' }
  finally { templatesLoading.value = false }
}
const resetForm = () => { Object.assign(form, { title: '', description: '', workspace_id: '', resource_group_id: '', status: 'active' }); tagText.value = ''; panelJson.value = '[]'; formError.value = '' }
const openCreate = () => { editingId.value = ''; resetForm(); dialogVisible.value = true }
const openEdit = row => {
  editingId.value = row.id
  Object.assign(form, { title: row.title, description: row.description, workspace_id: row.businessSpace, resource_group_id: row.resourceGroup, status: validStatuses.includes(row.status) ? row.status : 'active' })
  tagText.value = row.tags.join(', '); panelJson.value = JSON.stringify(toPanels(row.raw), null, 2); formError.value = ''; dialogVisible.value = true
}
const buildPayload = () => {
  if (!form.title) throw new Error('标题不能为空')
  let panels
  try { panels = panelJson.value.trim() ? JSON.parse(panelJson.value) : [] } catch (error) { throw new Error(`Panel 配置不是合法 JSON：${error.message}`) }
  if (!Array.isArray(panels)) throw new Error('Panel 配置 JSON 必须是数组')
  return { title: form.title, description: form.description, workspace_id: form.workspace_id, resource_group_id: form.resource_group_id, tags: toTags(tagText.value), variables: {}, panels, status: validStatuses.includes(form.status) ? form.status : 'active' }
}
const saveDashboard = async () => {
  formError.value = ''
  let payload
  try { payload = buildPayload() } catch (error) { formError.value = error.message; return }
  saving.value = true
  try { editingId.value ? await dashboardsApi.update(editingId.value, payload) : await dashboardsApi.create(payload); ElMessage.success('已保存仪表盘'); dialogVisible.value = false; await loadDashboards() }
  catch (error) { formError.value = error.message || '保存失败' }
  finally { saving.value = false }
}
const previewTemplate = async row => {
  selectedTemplate.value = row; previewVisible.value = true; previewJson.value = '加载中...'
  try {
    const detail = await dashboardsApi.getTemplate(row.id)
    selectedTemplate.value = normalizeTemplate({ ...row.raw, ...detail })
    previewJson.value = JSON.stringify(detail, null, 2)
  } catch (error) { previewJson.value = error.message || '模板详情加载失败' }
}
const openImport = row => {
  if (!row?.id) return
  selectedTemplate.value = row; Object.assign(importForm, { title: `${row.name}`, workspace_id: '', resource_group_id: '' })
  importTagText.value = row.tags.join(', '); variablesJson.value = '{}'; importError.value = ''; importVisible.value = true
}
const submitImport = async () => {
  importError.value = ''
  if (!selectedTemplate.value?.id) { importError.value = '请先选择模板'; return }
  if (!importForm.title) { importError.value = '标题不能为空'; return }
  let variables
  try { variables = parseObjectJson(variablesJson.value) } catch (error) { importError.value = error.message; return }
  importing.value = true
  try {
    await dashboardsApi.importTemplate(selectedTemplate.value.id, { title: importForm.title, workspace_id: importForm.workspace_id, resource_group_id: importForm.resource_group_id, variables, tags: toTags(importTagText.value) })
    ElMessage.success('已导入为仪表盘'); importVisible.value = false; previewVisible.value = false; await loadDashboards()
  } catch (error) { importError.value = error.message || '导入失败' }
  finally { importing.value = false }
}
const deleteDashboard = async row => {
  try { await ElMessageBox.confirm(`确认删除“${row.title}”？`, '删除仪表盘', { type: 'warning' }) } catch (error) { if (error !== 'cancel' && error !== 'close') ElMessage.warning('删除确认未完成'); return }
  try { await dashboardsApi.remove(row.id); ElMessage.success('已删除仪表盘'); await loadDashboards() } catch (error) { ElMessage.error(error.message || '删除失败') }
}
const cloneDashboard = async row => { try { await dashboardsApi.clone(row.id); ElMessage.success('已克隆仪表盘'); await loadDashboards() } catch (error) { ElMessage.error(error.message || '克隆失败') } }
const shareDashboard = async row => { if (row.shared) return; try { await dashboardsApi.share(row.id, { shared: true }); ElMessage.success('已分享仪表盘'); await loadDashboards() } catch (error) { ElMessage.error(error.message || '分享状态更新失败') } }

onMounted(() => { loadDashboards(); loadTemplates() })
</script>

<style scoped>
.dash-page { min-height: 100%; padding: 24px; color: #243553; background: linear-gradient(180deg, #f6faff 0%, #fff 64%); }
.dash-shell { min-height: calc(100vh - 114px); padding: 22px; border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.94); box-shadow: 0 12px 34px rgba(31,45,61,.06); }
.dash-head, .section-title { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.section-title { align-items: center; margin-bottom: 12px; }
.section-title span { color: #60728e; font-size: 12px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.head-actions { display: flex; gap: 10px; flex-wrap: wrap; justify-content: flex-end; }
.template-box { margin-bottom: 18px; padding-bottom: 18px; border-bottom: 1px solid #e8edf5; }
.state-alert { margin-bottom: 16px; }
.empty-state { min-height: 320px; display: grid; place-items: center; border: 1px dashed #d8e1ee; border-radius: 8px; background: #f8fbff; }
.dash-table { width: 100%; }
.tag { margin: 2px 4px 2px 0; }
.muted { color: #8a9ab1; }
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 14px; }
.json-view { margin-top: 14px; }
@media (max-width: 760px) {
  .dash-page { padding: 14px; }
  .dash-shell { padding: 16px; }
  .dash-head { flex-direction: column; }
  .section-title { align-items: flex-start; flex-direction: column; gap: 6px; }
  .form-grid { grid-template-columns: 1fr; }
}
</style>
