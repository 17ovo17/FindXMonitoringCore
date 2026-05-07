<template>
  <section class="rules-panel">
    <header class="toolbar">
      <div>
        <div class="kicker">规则管理</div>
        <h3>告警规则</h3>
      </div>
      <div class="toolbar-actions">
        <el-button :loading="loading" @click="load">刷新</el-button>
        <el-select v-model="filter.datasource" clearable filterable placeholder="数据源" class="w-150">
          <el-option v-for="item in datasources" :key="datasourceValue(item)" :label="datasourceLabel(item)" :value="datasourceValue(item)" />
        </el-select>
        <el-select v-model="filter.severities" multiple clearable collapse-tags placeholder="级别" class="w-170">
          <el-option v-for="item in severityOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-input v-model="filter.query" clearable placeholder="搜索规则、标签" class="search" @keyup.enter="load" />
        <el-select v-model="filter.enabled" clearable placeholder="启用" class="w-110">
          <el-option label="启用" value="true" />
          <el-option label="禁用" value="false" />
        </el-select>
        <el-button type="primary" @click="openCreate">新增</el-button>
        <el-button @click="importVisible = true">导入</el-button>
        <el-dropdown trigger="click" @command="handleBatch">
          <el-button>更多</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="export-json">批量导出 JSON</el-dropdown-item>
              <el-dropdown-item command="delete">批量删除</el-dropdown-item>
              <el-dropdown-item command="batch-update">批量更新</el-dropdown-item>
              <el-dropdown-item command="clone-to-business">克隆到业务组</el-dropdown-item>
              <el-dropdown-item command="clone-to-hosts">克隆到机器</el-dropdown-item>
              <el-dropdown-item command="export-csv" divided>导出 CSV</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-dropdown trigger="click" :hide-on-click="false">
          <el-button>列设置</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-for="col in ruleColumnOptions" :key="col.key">
                <el-checkbox v-model="visibleColumns[col.key]" @change="persistColumns">{{ col.label }}</el-checkbox>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" class="state-alert" />
    <el-table v-loading="loading" :data="filteredRules" border class="rules-table" empty-text="暂无真实告警规则数据" @selection-change="selectedRows = $event">
      <el-table-column type="selection" width="42" />
      <el-table-column v-if="visibleColumns.eventStatus" label="当前事件" width="116">
        <template #default="{ row }">
          <el-button v-if="ruleEvent(row)" link type="danger" @click="openRuleEvents(row)">
            <el-tag :type="eventStatusTag(ruleEvent(row).status)" size="small">{{ ruleEvent(row).status }}</el-tag>
          </el-button>
          <el-button v-else link @click="openRuleEvents(row)">无事件</el-button>
        </template>
      </el-table-column>
      <el-table-column v-if="visibleColumns.prod" label="类型/分类" min-width="120">
        <template #default="{ row }"><el-tag size="small" type="info">{{ row.prod }}</el-tag></template>
      </el-table-column>
      <el-table-column v-if="visibleColumns.datasource" label="数据源" min-width="130" show-overflow-tooltip>
        <template #default="{ row }">{{ row.datasourceId }}</template>
      </el-table-column>
      <el-table-column v-if="visibleColumns.name" label="名称" min-width="210" fixed show-overflow-tooltip>
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">{{ row.name }}</el-button>
          <div class="ident">v{{ row.version }} · {{ row.status }}</div>
        </template>
      </el-table-column>
      <el-table-column v-if="visibleColumns.severity" label="级别" width="120">
        <template #default="{ row }"><el-tag :type="severityTag(row.severity)" size="small">{{ severityText(row.severity) }}</el-tag></template>
      </el-table-column>
      <el-table-column v-if="visibleColumns.labels" label="附加标签" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="tag in mapToTags(row.labels)" :key="tag" size="small" class="tag">{{ tag }}</el-tag>
          <span v-if="mapToTags(row.labels).length === 0" class="muted">无</span>
        </template>
      </el-table-column>
      <el-table-column v-if="visibleColumns.notify" label="通知" min-width="140">
        <template #default><el-tag size="small" type="info">规则内配置</el-tag></template>
      </el-table-column>
      <el-table-column v-if="visibleColumns.updatedAt" label="更新时间" min-width="160">
        <template #default="{ row }">{{ formatDate(row.updatedAt) }}</template>
      </el-table-column>
      <el-table-column v-if="visibleColumns.updatedBy" label="更新人" min-width="110">
        <template #default="{ row }">{{ row.updatedBy }}</template>
      </el-table-column>
      <el-table-column v-if="visibleColumns.enabled" label="启用" width="92">
        <template #default="{ row }"><el-switch v-model="row.enabled" size="small" @change="value => toggleEnabled(row, value)" /></template>
      </el-table-column>
      <el-table-column label="操作" width="210" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link @click="cloneRule(row)">克隆</el-button>
          <el-button link @click="openRuleEvents(row)">事件</el-button>
          <el-button link type="danger" @click="deleteRule(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <AlertRuleFormDrawer v-model:visible="formVisible" :rule="editingRule" :datasources="datasources" :saving="saving" @save="saveRule" />
    <AlertEventsDrawer v-model:visible="eventsVisible" :rule="eventsRule" :datasources="datasources" @blocked="openBlocked" />

    <el-dialog v-model="importVisible" title="导入告警规则" width="720px">
      <el-tabs v-model="importTab">
        <el-tab-pane label="内置规则" name="rule-import-builtin" />
        <el-tab-pane label="JSON 导入" name="rule-import-json" />
        <el-tab-pane label="外部规则" name="rule-import-prometheus" />
      </el-tabs>
      <el-alert :title="blockedContracts[importTab]" type="warning" show-icon :closable="false" />
      <el-input :model-value="blockedPayload(importTab)" type="textarea" :rows="10" readonly spellcheck="false" class="json-view" />
      <template #footer>
        <el-button @click="importVisible = false">关闭</el-button>
        <el-button type="primary" @click="openBlocked(importTab)">查看阻断详情</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { alertingApi, isPermissionError, isUnauthorizedError, normalizeList, redactText, safeJson } from '../../api/alerting'
import AlertEventsDrawer from './AlertEventsDrawer.vue'
import AlertRuleFormDrawer from './AlertRuleFormDrawer.vue'
import {
  blockedContracts,
  blockedPayload,
  datasourceLabel,
  datasourceValue,
  defaultVisibleColumns,
  downloadText,
  eventStatusTag,
  filterRules,
  formatDate,
  mapToTags,
  normalizeEvent,
  normalizeRule,
  ruleColumnOptions,
  severityOptions,
  severityTag,
  severityText,
  summarizeRuleEvent,
  toCsv,
} from './alertingModel'

const emit = defineEmits(['blocked'])
const rules = ref([])
const currentEvents = ref([])
const datasources = ref([])
const selectedRows = ref([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const formVisible = ref(false)
const eventsVisible = ref(false)
const importVisible = ref(false)
const importTab = ref('rule-import-builtin')
const editingRule = ref(null)
const eventsRule = ref(null)
const filter = reactive({ query: '', datasource: '', severities: [], enabled: '' })
const visibleColumns = reactive(loadColumns())
const filteredRules = computed(() => filterRules(rules.value, filter))

function loadColumns() {
  try { return { ...defaultVisibleColumns, ...JSON.parse(localStorage.getItem('findx-alert-rule-columns') || '{}') } } catch { return { ...defaultVisibleColumns } }
}

const persistColumns = () => localStorage.setItem('findx-alert-rule-columns', JSON.stringify(visibleColumns))
const formatError = err => isPermissionError(err) ? (isUnauthorizedError(err) ? '登录已过期，请重新登录后继续访问告警能力' : '无权限访问该告警能力') : redactText(err?.message || '请求失败')
const ruleEvent = row => summarizeRuleEvent(row, currentEvents.value)

const load = async () => {
  loading.value = true
  error.value = ''
  try {
    const [ruleData, eventData, dsData] = await Promise.all([
      alertingApi.listRules({}),
      alertingApi.eventsCurrent({}),
      alertingApi.datasources().catch(() => []),
    ])
    rules.value = normalizeList(ruleData).map(normalizeRule)
    currentEvents.value = normalizeList(eventData).map(normalizeEvent)
    datasources.value = normalizeList(dsData)
  } catch (err) {
    error.value = formatError(err)
  } finally {
    loading.value = false
  }
}

const openCreate = () => { editingRule.value = null; formVisible.value = true }
const openEdit = async row => {
  editingRule.value = row
  formVisible.value = true
  try {
    const detail = await alertingApi.getRule(row.id)
    editingRule.value = normalizeRule(detail.rule || detail)
  } catch (err) {
    ElMessage.error(formatError(err))
  }
}

const saveRule = async payload => {
  saving.value = true
  try {
    if (payload.id) await alertingApi.updateRule(payload.id, payload)
    else await alertingApi.createRule(payload)
    ElMessage.success('规则已保存')
    formVisible.value = false
    await load()
  } catch (err) {
    ElMessage.error(formatError(err))
  } finally {
    saving.value = false
  }
}

const toggleEnabled = async (row, value) => {
  try {
    if (value) await alertingApi.enableRule(row.id)
    else await alertingApi.disableRule(row.id)
    ElMessage.success(value ? '规则已启用' : '规则已禁用')
    await load()
  } catch (err) {
    row.enabled = !value
    ElMessage.error(formatError(err))
  }
}

const cloneRule = async row => {
  try {
    await alertingApi.cloneRule(row.id)
    ElMessage.success('规则已克隆')
    await load()
  } catch (err) {
    ElMessage.error(formatError(err))
  }
}

const deleteRule = async row => {
  try {
    await ElMessageBox.confirm(`确认删除规则「${row.name}」？`, '删除确认', { type: 'warning' })
    await alertingApi.removeRule(row.id)
    ElMessage.success('规则已删除')
    await load()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error(formatError(err))
  }
}

const openRuleEvents = row => { eventsRule.value = row; eventsVisible.value = true }
const openBlocked = (action, context = {}) => emit('blocked', action, context)

const requireSelection = () => {
  if (selectedRows.value.length) return true
  ElMessage.warning('请先选择规则')
  return false
}

const handleBatch = async command => {
  if (!requireSelection()) return
  if (command === 'export-json') {
    downloadText('alert-rules.json', safeJson(selectedRows.value.map(row => row.raw || row)), 'application/json;charset=utf-8')
  } else if (command === 'export-csv') {
    downloadText('alert-rules.csv', toCsv(selectedRows.value), 'text/csv;charset=utf-8')
  } else if (command === 'delete') {
    try {
      await ElMessageBox.confirm(`确认删除 ${selectedRows.value.length} 条规则？`, '批量删除确认', { type: 'warning' })
      await Promise.all(selectedRows.value.map(row => alertingApi.removeRule(row.id)))
      ElMessage.success('批量删除已完成')
      await load()
    } catch (err) {
      if (err !== 'cancel') ElMessage.error(formatError(err))
    }
  } else {
    openBlocked(command, { selected: selectedRows.value.map(row => row.id) })
  }
}

defineExpose({ load })
onMounted(load)
</script>

<style scoped>
.rules-panel { min-width: 0; padding: 16px; border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.toolbar-actions { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.kicker, .ident, .muted { color: #63718a; font-size: 12px; }
h3 { margin: 4px 0 0; color: #17233c; font-size: 20px; }
.w-110 { width: 110px; }
.w-150 { width: 150px; }
.w-170 { width: 170px; }
.search { width: 220px; }
.state-alert { margin-bottom: 14px; }
.rules-table { width: 100%; }
.tag { margin: 2px 4px 2px 0; }
.json-view { margin-top: 12px; }
@media (max-width: 900px) {
  .toolbar { align-items: flex-start; flex-direction: column; }
  .toolbar-actions, .search, .w-110, .w-150, .w-170 { width: 100%; }
}
</style>
