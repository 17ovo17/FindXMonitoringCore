<template>
  <section class="events-layout">
    <aside class="filter-panel">
      <div class="filter-title">事件筛选</div>
      <el-segmented v-model="scope" :options="scopeOptions" size="small" @change="load" />
      <el-divider />
      <label>关键字</label>
      <el-input v-model="filter.query" clearable placeholder="搜索事件、对象、标签" @keyup.enter="load" />
      <label>数据源</label>
      <el-select v-model="filter.datasource" clearable filterable placeholder="全部数据源">
        <el-option v-for="item in datasources" :key="datasourceValue(item)" :label="datasourceLabel(item)" :value="datasourceValue(item)" />
      </el-select>
      <label>级别</label>
      <el-select v-model="filter.severities" multiple clearable collapse-tags placeholder="全部级别">
        <el-option v-for="item in severityOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <label>状态</label>
      <el-select v-model="filter.statuses" multiple clearable collapse-tags placeholder="全部状态">
        <el-option v-for="item in eventStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-button class="full" :loading="loading" @click="load">刷新</el-button>
    </aside>

    <main class="event-work-area">
      <header class="toolbar">
        <div>
          <div class="kicker">事件中心</div>
          <h3>{{ scope === 'current' ? '当前事件' : '历史事件' }}</h3>
        </div>
        <div class="toolbar-actions">
          <el-button :loading="loading" @click="load">刷新</el-button>
          <el-dropdown trigger="click" @command="handleBatch">
            <el-button>批量动作</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="ack">ack</el-dropdown-item>
                <el-dropdown-item command="assign">assign</el-dropdown-item>
                <el-dropdown-item command="resolve">resolve</el-dropdown-item>
                <el-dropdown-item command="archive">archive</el-dropdown-item>
                <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button @click="$emit('blocked', 'events-aggregate', { scope })">聚合规则</el-button>
        </div>
      </header>

      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" class="state-alert" />
      <el-table v-loading="loading" :data="filteredEvents" border class="event-table" empty-text="暂无真实事件数据" @selection-change="selectedRows = $event">
        <el-table-column type="selection" width="42" />
        <el-table-column label="事件" min-width="220" fixed show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">{{ row.name }}</el-button>
            <div class="ident">{{ row.target }}</div>
          </template>
        </el-table-column>
        <el-table-column label="级别" width="128">
          <template #default="{ row }"><el-tag :type="severityTag(row.severity)" size="small">{{ severityText(row.severity) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="状态" width="132">
          <template #default="{ row }"><el-tag :type="eventStatusTag(row.status)" size="small">{{ row.status }}</el-tag></template>
        </el-table-column>
        <el-table-column label="数据源" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.datasourceId }}</template>
        </el-table-column>
        <el-table-column label="次数" prop="count" width="72" />
        <el-table-column label="首次触发" min-width="160"><template #default="{ row }">{{ formatDate(row.firstSeen) }}</template></el-table-column>
        <el-table-column label="最近触发" min-width="160"><template #default="{ row }">{{ formatDate(row.lastSeen) }}</template></el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button link @click="submitAction(row, 'ack')">ack</el-button>
            <el-button link @click="openAssign(row)">assign</el-button>
            <el-button link type="success" @click="submitAction(row, 'resolve')">resolve</el-button>
            <el-button link type="warning" @click="submitAction(row, 'archive')">archive</el-button>
          </template>
        </el-table-column>
      </el-table>
    </main>

    <AlertEventDetailDrawer
      v-model:visible="detailVisible"
      :event="detailEvent"
      :error="detailError"
      @action="submitAction"
      @assign="openAssign"
    />
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { alertingApi, isPermissionError, isUnauthorizedError, normalizeList, redactText } from '../../api/alerting'
import AlertEventDetailDrawer from './AlertEventDetailDrawer.vue'
import { datasourceLabel, datasourceValue, eventStatusOptions, eventStatusTag, filterEvents, formatDate, normalizeEvent, severityOptions, severityTag, severityText, sortEvents } from './alertingModel'

const emit = defineEmits(['blocked'])
const props = defineProps({
  ruleId: { type: String, default: '' },
  datasources: { type: Array, default: () => [] },
})

const scope = ref('current')
const scopeOptions = [{ label: '当前', value: 'current' }, { label: '历史', value: 'history' }]
const loading = ref(false)
const error = ref('')
const detailError = ref('')
const events = ref([])
const selectedRows = ref([])
const detailVisible = ref(false)
const detailEvent = ref(null)
const filter = reactive({ query: '', datasource: '', severities: [], statuses: [], ruleId: props.ruleId })

const filteredEvents = computed(() => sortEvents(filterEvents(events.value, filter)))

const formatError = error => {
  if (isPermissionError(error)) return isUnauthorizedError(error) ? '登录已过期，请重新登录后继续访问告警能力' : '无权限访问该告警能力'
  return redactText(error?.message || '请求失败')
}

const load = async () => {
  loading.value = true
  error.value = ''
  try {
    const data = scope.value === 'current' ? await alertingApi.eventsCurrent({}) : await alertingApi.eventsHistory({})
    events.value = normalizeList(data).map(normalizeEvent)
  } catch (err) {
    events.value = []
    error.value = formatError(err)
  } finally {
    loading.value = false
  }
}

const openDetail = async row => {
  detailVisible.value = true
  detailEvent.value = row
  detailError.value = ''
  try {
    detailEvent.value = normalizeEvent(await alertingApi.getEvent(row.id))
  } catch (err) {
    detailError.value = formatError(err)
  }
}

const applyAction = async (row, action, extra = {}) => {
  const updated = normalizeEvent(await alertingApi.eventAction(row.id, action, { actor: '当前用户', reason: action, ...extra }))
  detailEvent.value = detailEvent.value?.id === updated.id ? updated : detailEvent.value
  await load()
  ElMessage.success('操作已提交')
}

const submitAction = async (row, action, extra = {}) => {
  try {
    await applyAction(row, action, extra)
  } catch (err) {
    const message = formatError(err)
    detailError.value = message
    ElMessage.error(message)
  }
}

const openAssign = async row => {
  try {
    const { value } = await ElMessageBox.prompt('请输入处理人或团队', 'assign', { inputPattern: /^.{1,40}$/, inputErrorMessage: '1-40 个字符' })
    await submitAction(row, 'assign', { assignee: value.trim(), reason: 'assign' })
  } catch (err) {
    if (err !== 'cancel') ElMessage.error(formatError(err))
  }
}

const handleBatch = async command => {
  if (command === 'delete') {
    emit('blocked', 'events-delete', { selected: selectedRows.value.map(item => item.id) })
    return
  }
  if (!selectedRows.value.length) {
    ElMessage.warning('请先选择事件')
    return
  }
  if (command === 'assign') {
    try {
      const { value } = await ElMessageBox.prompt('请输入处理人或团队', '批量 assign', { inputPattern: /^.{1,40}$/, inputErrorMessage: '1-40 个字符' })
      await Promise.all(selectedRows.value.map(row => applyAction(row, 'assign', { assignee: value.trim(), reason: 'batch assign' })))
    } catch (err) {
      if (err !== 'cancel') ElMessage.error(formatError(err))
    }
    return
  }
  try {
    await Promise.all(selectedRows.value.map(row => applyAction(row, command)))
  } catch (err) {
    ElMessage.error(formatError(err))
  }
}

defineExpose({ load })
onMounted(load)
</script>

<style scoped>
.events-layout { display: grid; grid-template-columns: 220px minmax(0, 1fr); gap: 16px; min-height: calc(100vh - 126px); }
.filter-panel, .event-work-area { border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.filter-panel { padding: 14px; }
.filter-title { margin-bottom: 12px; color: #17233c; font-size: 17px; font-weight: 700; }
.filter-panel label { display: block; margin: 12px 0 6px; color: #63718a; font-size: 12px; font-weight: 700; }
.filter-panel .el-select, .filter-panel .el-input, .full { width: 100%; }
.full { margin-top: 14px; }
.event-work-area { min-width: 0; padding: 16px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.toolbar-actions { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.kicker, .ident { color: #63718a; font-size: 12px; }
h3 { margin: 4px 0 0; color: #17233c; font-size: 20px; }
.state-alert { margin-bottom: 14px; }
.event-table { width: 100%; }
@media (max-width: 900px) {
  .events-layout { grid-template-columns: 1fr; }
  .toolbar { align-items: flex-start; flex-direction: column; }
}
</style>
