<template>
  <div class="panel-grid">
    <section class="box">
      <div class="bar">
        <b>监控对象</b>
        <el-select v-model="status" clearable size="small" placeholder="状态" @change="loadTargets">
          <el-option v-for="s in statuses" :key="s" :value="s" :label="s" />
        </el-select>
      </div>
      <el-alert v-if="targetError" :title="targetError" type="error" show-icon :closable="false" />
      <el-table :data="targets" v-loading="targetLoading" height="360" empty-text="暂无监控对象">
        <el-table-column label="名称" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ safe(row.name) || '-' }}</template>
        </el-table-column>
        <el-table-column label="IP" width="140">
          <template #default="{ row }">{{ safe(row.ip) || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }"><el-tag size="small" :type="tag(row.status)">{{ safe(row.status) || 'unknown' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="来源" width="110">
          <template #default="{ row }">{{ safe(row.source) || '-' }}</template>
        </el-table-column>
        <el-table-column label="最近心跳" width="170"><template #default="{ row }">{{ fmt(row.last_seen) }}</template></el-table-column>
      </el-table>
    </section>

    <section class="box">
      <div class="bar"><b>FindX Agents</b><el-button size="small" @click="loadAgents">刷新</el-button></div>
      <el-alert v-if="agentError" :title="agentError" type="error" show-icon :closable="false" />
      <el-table :data="agents" v-loading="agentLoading" height="360" empty-text="暂无 Agent">
        <el-table-column label="标识" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ safe(row.ident) || '-' }}</template>
        </el-table-column>
        <el-table-column label="IP" width="130">
          <template #default="{ row }">{{ safe(row.ip) || '-' }}</template>
        </el-table-column>
        <el-table-column label="版本" width="100">
          <template #default="{ row }">{{ safe(row.version) || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }"><el-tag size="small" :type="tag(row.status)">{{ safe(row.status) || '-' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="能力" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ safe((row.capabilities || []).join(', ')) || '-' }}</template>
        </el-table-column>
      </el-table>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { monitorApi, normalizeList, redactText, isPermissionError, isUnauthorizedError } from '../../api/monitor'

const statuses = ['online', 'warning', 'offline', 'unknown', 'maintenance']
const status = ref('')
const targets = ref([])
const agents = ref([])
const targetLoading = ref(false)
const agentLoading = ref(false)
const targetError = ref('')
const agentError = ref('')

const tag = value => ({ online: 'success', warning: 'warning', offline: 'danger', maintenance: 'info' }[value] || 'info')
const fmt = time => time ? new Date(time).toLocaleString('zh-CN', { hour12: false }) : '-'
const safe = value => redactText(value)
const permissionMessage = error => isUnauthorizedError(error) ? '登录已过期，请重新登录后继续访问监控能力' : '无权限访问该监控能力'
const formatError = error => isPermissionError(error) ? permissionMessage(error) : redactText(error?.message || '请求失败')

const loadTargets = async () => {
  targetLoading.value = true
  targetError.value = ''
  try {
    targets.value = normalizeList(await monitorApi.targets({ status: status.value }))
  } catch (e) {
    targetError.value = formatError(e)
  } finally {
    targetLoading.value = false
  }
}

const loadAgents = async () => {
  agentLoading.value = true
  agentError.value = ''
  try {
    agents.value = normalizeList(await monitorApi.agents())
  } catch (e) {
    agentError.value = formatError(e)
  } finally {
    agentLoading.value = false
  }
}

defineExpose({ load: () => Promise.all([loadTargets(), loadAgents()]) })
onMounted(() => Promise.all([loadTargets(), loadAgents()]))
</script>

<style scoped>
.panel-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 12px; }
.box { padding: 14px; border: 1px solid #e4e9f2; border-radius: 8px; background: #fff; }
.bar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
@media (max-width: 520px) { .panel-grid { grid-template-columns: 1fr; } }
</style>
