<template>
  <section class="channels-layout">
    <aside class="type-panel">
      <el-input v-model="typeSearch" clearable placeholder="搜索媒介类型" />
      <div class="type-grid">
        <button v-for="item in filteredTypes" :key="item.ident" class="type-card" type="button" @click="openType(item)">
          <span class="type-icon">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
          <el-tag v-if="!item.supported" size="small" type="warning">BLOCKED</el-tag>
        </button>
      </div>
    </aside>

    <main class="channel-work-area">
      <header class="toolbar">
        <div>
          <div class="kicker">通知媒介</div>
          <h3>媒介配置</h3>
        </div>
        <div class="toolbar-actions">
          <el-input v-model="filter.query" clearable placeholder="搜索名称" class="search" />
          <el-select v-model="filter.status" clearable placeholder="状态" class="w-110">
            <el-option label="启用" value="true" />
            <el-option label="禁用" value="false" />
          </el-select>
          <el-select v-model="filter.types" multiple clearable collapse-tags placeholder="类型" class="w-170">
            <el-option v-for="item in channelTypes" :key="item.ident" :label="item.label" :value="item.ident" />
          </el-select>
          <el-button :loading="loading" @click="load">刷新</el-button>
          <el-button @click="importVisible = true">导入</el-button>
          <el-button @click="exportSelected">导出</el-button>
        </div>
      </header>

      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" class="state-alert" />
      <el-table v-loading="loading" :data="filteredChannels" border class="channel-table" empty-text="暂无真实通知媒介数据" @selection-change="selectedRows = $event">
        <el-table-column type="selection" width="42" />
        <el-table-column label="名称" min-width="180" fixed show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="openEdit(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column label="类型" min-width="140">
          <template #default="{ row }"><span class="type-chip"><span class="type-icon small">{{ typeIcon(row.type) }}</span>{{ typeLabel(row.type) }}</span></template>
        </el-table-column>
        <el-table-column label="接收对象" min-width="160" show-overflow-tooltip><template #default="{ row }">{{ row.receiver || '-' }}</template></el-table-column>
        <el-table-column label="更新人" min-width="100"><template #default="{ row }">{{ row.updatedBy }}</template></el-table-column>
        <el-table-column label="更新时间" min-width="160"><template #default="{ row }">{{ formatDate(row.updatedAt) }}</template></el-table-column>
        <el-table-column label="启用" width="92">
          <template #default="{ row }"><el-switch v-model="row.enabled" size="small" @change="value => toggle(row, value)" /></template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link @click="cloneChannel(row)">克隆</el-button>
            <el-button link type="danger" :disabled="row.enabled" @click="deleteChannel(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </main>

    <el-drawer v-model="formVisible" :title="form.id ? '编辑通知媒介' : '新增通知媒介'" size="620px">
      <el-form label-position="top" @submit.prevent>
        <div class="form-grid">
          <el-form-item label="类型"><el-input :model-value="typeLabel(form.type)" disabled /></el-form-item>
          <el-form-item label="启用"><el-switch v-model="form.enabled" /></el-form-item>
          <el-form-item label="名称" required><el-input v-model.trim="form.name" maxlength="80" /></el-form-item>
          <el-form-item label="接收对象"><el-input v-model.trim="form.receiver" maxlength="120" /></el-form-item>
        </div>
        <el-form-item label="Webhook / Endpoint" required>
          <el-input v-model.trim="form.endpoint" placeholder="凭据不回显；编辑时如需保存请重新输入" show-password />
        </el-form-item>
        <el-form-item label="Secret">
          <el-input v-model.trim="form.secret" placeholder="可选，保存后不回显" show-password />
        </el-form-item>
        <el-alert v-if="formError" :title="formError" type="error" show-icon :closable="false" />
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-drawer>

    <el-dialog v-model="importVisible" title="导入通知媒介" width="720px">
      <el-alert title="仅支持当前已接入投递 contract 的媒介类型；凭据字段请使用部署环境中的引用或重新输入。" type="info" show-icon :closable="false" />
      <el-input v-model="importJson" type="textarea" :rows="14" spellcheck="false" placeholder="[{...}]" class="json-view" />
      <el-alert v-if="importError" :title="importError" type="error" show-icon :closable="false" class="state-alert" />
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitImport">导入</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { normalizeList, notificationsApi, redactText, safeJson } from '../../api/notifications'
import { channelTypes, downloadText, filterChannels, formatDate, normalizeChannel, supportedChannelTypes, typeIcon, typeLabel } from './notificationModel'

const emit = defineEmits(['blocked'])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const formError = ref('')
const importError = ref('')
const channels = ref([])
const selectedRows = ref([])
const formVisible = ref(false)
const importVisible = ref(false)
const typeSearch = ref('')
const importJson = ref('')
const filter = reactive({ query: '', status: '', types: [] })
const form = reactive({ id: '', type: 'dingtalk', name: '', endpoint: '', receiver: '', secret: '', enabled: true })
const filteredTypes = computed(() => channelTypes.filter(item => `${item.ident} ${item.label}`.toLowerCase().includes(typeSearch.value.toLowerCase())))
const filteredChannels = computed(() => filterChannels(channels.value, filter))
const formatError = err => redactText(err?.message || '请求失败')

const load = async () => {
  loading.value = true
  error.value = ''
  try { channels.value = normalizeList(await notificationsApi.listChannels()).map(normalizeChannel) } catch (err) { error.value = formatError(err); channels.value = [] } finally { loading.value = false }
}

const resetForm = type => Object.assign(form, { id: '', type, name: '', endpoint: '', receiver: '', secret: '', enabled: true })
const openType = item => {
  if (!item.supported) { emit('blocked', 'channel-type', { type: item.ident }); return }
  resetForm(item.ident)
  formVisible.value = true
}
const openEdit = row => { Object.assign(form, { id: row.id, type: row.type, name: row.name, endpoint: '', receiver: row.receiver, secret: '', enabled: row.enabled }); formVisible.value = true }
const cloneChannel = row => { Object.assign(form, { id: '', type: row.type, name: `${row.name}-copy`, endpoint: '', receiver: row.receiver, secret: '', enabled: false }); formVisible.value = true }

const buildPayload = source => {
  if (!supportedChannelTypes.has(source.type)) throw new Error('该媒介类型未接入投递 contract')
  if (!source.name?.trim()) throw new Error('名称不能为空')
  if (!source.endpoint?.trim()) throw new Error('Webhook / Endpoint 不能为空')
  return { id: source.id || undefined, type: source.type, name: source.name.trim(), endpoint: source.endpoint.trim(), webhook: source.endpoint.trim(), receiver: source.receiver || '', secret: source.secret || '', enabled: source.enabled !== false }
}

const save = async () => {
  formError.value = ''
  saving.value = true
  try {
    await notificationsApi.saveChannel(buildPayload(form))
    ElMessage.success('通知媒介已保存')
    formVisible.value = false
    await load()
  } catch (err) {
    formError.value = formatError(err)
  } finally {
    saving.value = false
  }
}

const toggle = async (row, value) => {
  try {
    await notificationsApi.saveChannel({ ...row.raw, enabled: value })
    ElMessage.success(value ? '媒介已启用' : '媒介已禁用')
    await load()
  } catch (err) {
    row.enabled = !value
    ElMessage.error(formatError(err))
  }
}

const deleteChannel = async row => {
  if (row.enabled) return
  try {
    await ElMessageBox.confirm(`确认删除媒介「${row.name}」？`, '删除确认', { type: 'warning' })
    await notificationsApi.deleteChannel(row.id)
    ElMessage.success('媒介已删除')
    await load()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error(formatError(err))
  }
}

const exportSelected = () => {
  if (!selectedRows.value.length) { ElMessage.warning('请先选择媒介'); return }
  downloadText('notification-channels.json', safeJson(selectedRows.value), 'application/json;charset=utf-8')
}

const submitImport = async () => {
  importError.value = ''
  saving.value = true
  try {
    const parsed = JSON.parse(importJson.value)
    const rows = Array.isArray(parsed) ? parsed : [parsed]
    for (const row of rows) await notificationsApi.saveChannel(buildPayload(row))
    ElMessage.success('导入已完成')
    importVisible.value = false
    await load()
  } catch (err) {
    importError.value = formatError(err)
  } finally {
    saving.value = false
  }
}

defineExpose({ load })
onMounted(load)
</script>

<style scoped>
.channels-layout { display: grid; grid-template-columns: 360px minmax(0, 1fr); gap: 16px; min-height: calc(100vh - 126px); }
.type-panel, .channel-work-area { border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.type-panel { padding: 14px; overflow: auto; }
.type-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; margin-top: 14px; }
.type-card { min-height: 88px; border: 1px solid #e3e8f1; border-radius: 8px; background: #fbfcff; color: #23324d; cursor: pointer; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px; }
.type-card:hover { border-color: #1769ff; }
.type-icon { width: 34px; height: 34px; border-radius: 8px; background: #eef4ff; color: #1769ff; display: inline-grid; place-items: center; font-weight: 800; }
.type-icon.small { width: 24px; height: 24px; font-size: 12px; }
.type-chip { display: inline-flex; align-items: center; gap: 8px; }
.channel-work-area { min-width: 0; padding: 16px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.toolbar-actions { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h3 { margin: 4px 0 0; color: #17233c; font-size: 20px; }
.search { width: 200px; }
.w-110 { width: 110px; }
.w-170 { width: 170px; }
.state-alert { margin: 12px 0; }
.json-view { margin-top: 12px; }
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 14px; }
@media (max-width: 980px) {
  .channels-layout { grid-template-columns: 1fr; }
  .toolbar { align-items: flex-start; flex-direction: column; }
  .toolbar-actions, .search, .w-110, .w-170 { width: 100%; }
}
</style>
