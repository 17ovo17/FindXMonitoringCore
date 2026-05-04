<template>
  <div class="rules-grid">
    <section class="box">
      <div class="bar">
        <b>告警规则</b>
        <el-select v-model="status" clearable size="small" placeholder="状态" @change="load">
          <el-option value="active" label="active" />
          <el-option value="draft" label="draft" />
          <el-option value="disabled" label="disabled" />
        </el-select>
      </div>
      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
      <el-table :data="rules" v-loading="loading" height="420" highlight-current-row @current-change="detail">
        <el-table-column label="规则" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ safe(row.name) || '-' }}</template>
        </el-table-column>
        <el-table-column prop="severity" label="级别" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="sev(row.severity)">{{ safe(row.severity) || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="启用" width="70">
          <template #default="{ row }">{{ row.enabled ? '是' : '否' }}</template>
        </el-table-column>
        <el-table-column label="数据源" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ safe(row.datasource_id) || '-' }}</template>
        </el-table-column>
      </el-table>
    </section>

    <section class="box">
      <div class="bar">
        <b>详情 / TryRun</b>
        <el-button size="small" :disabled="!selected" :loading="tryLoading" @click="tryrun">tryrun</el-button>
      </div>
      <el-alert title="tryrun 只执行校验和 PromQL 评估，不生成正式事件。" type="info" show-icon :closable="false" />
      <el-alert v-if="detailError" :title="detailError" type="error" show-icon :closable="false" style="margin-top:10px" />
      <el-empty v-if="!selected" description="请选择规则" />
      <template v-else>
        <el-descriptions :column="2" border size="small" class="desc">
          <el-descriptions-item label="名称">{{ safe(selected.name) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="版本">v{{ safe(selected.version) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="级别">{{ safe(selected.severity) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="持续">{{ safe(selected.for_duration) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="无数据策略">{{ safe(selected.no_data_policy) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ safe(selected.status) || '-' }}</el-descriptions-item>
        </el-descriptions>
        <pre class="code">{{ safe(selected.query || '-') }}</pre>
        <div class="minor">历史版本：{{ versions.length }}</div>
        <pre v-if="tryResult" class="code">{{ safeJson(tryResult, 10000) }}</pre>
      </template>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { monitorApi, normalizeList, safeJson, redactText, isPermissionError, isUnauthorizedError } from '../../api/monitor'

const rules = ref([])
const selected = ref(null)
const versions = ref([])
const loading = ref(false)
const tryLoading = ref(false)
const error = ref('')
const detailError = ref('')
const status = ref('')
const tryResult = ref(null)

const sev = value => ({ critical: 'danger', p0: 'danger', p1: 'warning', warning: 'warning', info: 'info' }[value] || 'info')
const safe = value => redactText(value)
const permissionMessage = error => isUnauthorizedError(error) ? '登录已过期，请重新登录后继续访问监控能力' : '无权限访问该监控能力'
const formatError = error => isPermissionError(error) ? permissionMessage(error) : redactText(error?.message || '请求失败')

const setError = (target, error) => {
  target.value = formatError(error)
}

const load = async () => {
  loading.value = true
  error.value = ''
  try {
    const data = await monitorApi.alertRules({ status: status.value })
    rules.value = normalizeList(data)
  } catch (e) {
    setError(error, e)
  } finally {
    loading.value = false
  }
}

const detail = async row => {
  if (!row?.id) return
  selected.value = row
  versions.value = []
  tryResult.value = null
  detailError.value = ''
  try {
    const data = await monitorApi.alertRuleDetail(row.id)
    selected.value = data.rule || row
    versions.value = normalizeList(data.versions)
  } catch (e) {
    setError(detailError, e)
  }
}

const tryrun = async () => {
  if (!selected.value) return
  tryLoading.value = true
  detailError.value = ''
  tryResult.value = null
  try {
    tryResult.value = await monitorApi.tryRunRule(selected.value.id, selected.value)
  } catch (e) {
    setError(detailError, e)
  } finally {
    tryLoading.value = false
  }
}

defineExpose({ load })
onMounted(load)
</script>

<style scoped>
.rules-grid { display: grid; grid-template-columns: minmax(320px, 1fr) minmax(320px, 430px); gap: 12px; }
.box { padding: 14px; border: 1px solid #e4e9f2; border-radius: 8px; background: #fff; }
.bar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
.desc { margin-top: 12px; }
.code { max-height: 240px; overflow: auto; padding: 12px; border-radius: 8px; background: #f5f7fb; white-space: pre-wrap; word-break: break-word; font-size: 12px; }
.minor { color: #6b778c; font-size: 12px; margin: 10px 0; }
@media (max-width: 1120px) { .rules-grid { grid-template-columns: 1fr; } }
</style>
