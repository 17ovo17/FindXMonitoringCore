<template>
  <section class="box">
    <div class="bar">
      <el-segmented v-model="scope" :options="scopeOptions" size="small" @change="load" />
      <div class="filters">
        <el-select v-model="query.severity" clearable size="small" placeholder="级别" @change="load">
          <el-option v-for="s in severities" :key="s" :value="s" :label="s" />
        </el-select>
        <el-select v-model="query.status" clearable size="small" placeholder="状态" @change="load">
          <el-option v-for="s in statuses" :key="s" :value="s" :label="s" />
        </el-select>
        <el-button size="small" @click="load">刷新</el-button>
      </div>
    </div>
    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
    <el-table :data="events" v-loading="loading" height="470" empty-text="暂无事件">
      <el-table-column label="事件" min-width="220" show-overflow-tooltip>
        <template #default="{ row }">{{ safe(row.name) || '-' }}</template>
      </el-table-column>
      <el-table-column label="级别" width="90">
        <template #default="{ row }"><el-tag size="small" :type="sev(row.severity)">{{ safe(row.severity) || '-' }}</el-tag></template>
      </el-table-column>
      <el-table-column label="状态" width="110"><template #default="{ row }">{{ safe(row.status) || '-' }}</template></el-table-column>
      <el-table-column label="对象" min-width="130" show-overflow-tooltip>
        <template #default="{ row }">{{ safe(row.target_ident) || '-' }}</template>
      </el-table-column>
      <el-table-column prop="count" label="次数" width="70" />
      <el-table-column label="最近触发" width="170"><template #default="{ row }">{{ fmt(row.last_seen) }}</template></el-table-column>
      <el-table-column label="操作" width="230" fixed="right">
        <template #default="{ row }">
          <el-button size="small" link @click="open(row)">详情</el-button>
          <el-button size="small" link @click="act(row, 'ack')">ack</el-button>
          <el-button size="small" link @click="assign(row)">assign</el-button>
          <el-button size="small" link type="success" @click="act(row, 'resolve')">resolve</el-button>
          <el-button size="small" link type="warning" @click="act(row, 'archive')">archive</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-drawer v-model="drawer" title="事件详情" size="520px">
      <el-alert v-if="actionError" :title="actionError" type="error" show-icon :closable="false" />
      <el-empty v-if="!detailData" description="无详情" />
      <template v-else>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="名称">{{ safe(detailData.name) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ safe(detailData.status) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="对象">{{ safe(detailData.target_ident || detailData.target_id) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="值">{{ safe(detailData.value) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="指纹">{{ safe(detailData.fingerprint) || '-' }}</el-descriptions-item>
        </el-descriptions>
        <div class="drawer-actions">
          <el-button size="small" @click="act(detailData, 'ack')">ack</el-button>
          <el-button size="small" @click="assign(detailData)">assign</el-button>
          <el-button size="small" type="success" plain @click="act(detailData, 'resolve')">resolve</el-button>
          <el-button size="small" type="warning" plain @click="act(detailData, 'archive')">archive</el-button>
        </div>
        <h4>Actions</h4>
        <el-timeline>
          <el-timeline-item v-for="a in detailData.action_log || []" :key="a.id || a.created_at" :timestamp="fmt(a.created_at)">
            <b>{{ safe(a.action) || '-' }}</b> {{ safe(a.actor) || '-' }} {{ safe(a.from) || '' }} -> {{ safe(a.to) || '' }}
            <div class="minor">{{ safe(a.assignee) || '' }} {{ safe(a.reason) || '' }}</div>
          </el-timeline-item>
        </el-timeline>
        <h4>Labels</h4><pre>{{ safeJson(detailData.labels || {}) }}</pre>
        <h4>Annotations</h4><pre>{{ safeJson(detailData.annotations || {}) }}</pre>
      </template>
    </el-drawer>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { monitorApi, normalizeList, safeJson, redactText, isPermissionError, isUnauthorizedError } from '../../api/monitor'

const scope = ref('current')
const scopeOptions = [{ label: '当前', value: 'current' }, { label: '历史', value: 'history' }]
const events = ref([])
const loading = ref(false)
const error = ref('')
const drawer = ref(false)
const detailData = ref(null)
const actionError = ref('')
const query = reactive({ severity: '', status: '' })
const severities = ['critical', 'warning', 'info', 'p0', 'p1', 'p2', 'p3']
const statuses = ['firing', 'acknowledged', 'assigned', 'resolved', 'archived']

const sev = value => ({ critical: 'danger', p0: 'danger', p1: 'warning', warning: 'warning', info: 'info' }[value] || 'info')
const fmt = time => time ? new Date(time).toLocaleString('zh-CN', { hour12: false }) : '-'
const safe = value => redactText(value)
const permissionMessage = error => isUnauthorizedError(error) ? '登录已过期，请重新登录后继续访问监控能力' : '无权限访问该监控能力'
const formatError = error => isPermissionError(error) ? permissionMessage(error) : redactText(error?.message || '请求失败')

const setActionError = error => {
  actionError.value = formatError(error)
}

const load = async () => {
  loading.value = true
  error.value = ''
  try {
    const data = scope.value === 'current' ? await monitorApi.eventsCurrent(query) : await monitorApi.eventsHistory(query)
    events.value = normalizeList(data)
  } catch (e) {
    error.value = formatError(e)
  } finally {
    loading.value = false
  }
}

const open = async row => {
  drawer.value = true
  actionError.value = ''
  detailData.value = row
  try {
    detailData.value = await monitorApi.eventDetail(row.id)
  } catch (e) {
    setActionError(e)
  }
}

const act = async (row, action, extra = {}) => {
  actionError.value = ''
  try {
    const data = await monitorApi.eventAction(row.id, action, { actor: '当前用户', reason: action, ...extra })
    detailData.value = data
    ElMessage.success('操作已提交')
    await load()
  } catch (e) {
    setActionError(e)
    ElMessage.error(formatError(e))
  }
}

const assign = async row => {
  try {
    const { value } = await ElMessageBox.prompt('请输入处理人或值班组', 'assign', { inputPattern: /^.{1,40}$/, inputErrorMessage: '1-40 个字符' })
    await act(row, 'assign', { assignee: value.trim(), reason: 'assign' })
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(formatError(e))
  }
}

defineExpose({ load })
onMounted(load)
</script>

<style scoped>
.box { padding: 14px; border: 1px solid #e4e9f2; border-radius: 8px; background: #fff; }
.bar, .filters { display: flex; align-items: center; gap: 10px; justify-content: space-between; margin-bottom: 12px; }
.filters { justify-content: flex-end; margin-bottom: 0; }
.drawer-actions { display: flex; gap: 8px; margin: 12px 0; flex-wrap: wrap; }
.minor { color: #6b778c; font-size: 12px; margin-top: 4px; }
pre { max-height: 180px; overflow: auto; padding: 10px; border-radius: 8px; background: #f5f7fb; white-space: pre-wrap; word-break: break-word; font-size: 12px; }
@media (max-width: 760px) { .bar { align-items: stretch; flex-direction: column; } .filters { justify-content: flex-start; flex-wrap: wrap; } }
</style>
