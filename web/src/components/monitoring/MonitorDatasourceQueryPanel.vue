<template>
  <div class="query-layout">
    <section class="box">
      <div class="bar"><b>数据源</b><el-button size="small" @click="load">刷新</el-button></div>
      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
      <el-table :data="sources" v-loading="loading" height="210" highlight-current-row @current-change="pick">
        <el-table-column label="名称" min-width="140">
          <template #default="{ row }">{{ safe(row.name) || '-' }}</template>
        </el-table-column>
        <el-table-column label="ID" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ safe(row.id) || '-' }}</template>
        </el-table-column>
        <el-table-column label="地址" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ safe(row.url) || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90">
          <template #default="{ row }"><el-button size="small" link @click.stop="test(row)">测试</el-button></template>
        </el-table-column>
      </el-table>
      <el-alert v-if="testMsg" :title="testMsg" :type="testOK ? 'success' : 'error'" show-icon style="margin-top:10px" />
    </section>

    <section class="box">
      <div class="bar">
        <b>PromQL 查询</b>
        <el-segmented v-model="mode" :options="['instant', 'range']" size="small" />
      </div>
      <el-form label-width="92px" size="small">
        <el-form-item label="数据源"><el-input v-model="form.datasource_id" placeholder="选择或输入 datasource_id" /></el-form-item>
        <el-form-item label="查询语句"><el-input v-model="form.query" type="textarea" :rows="3" placeholder="up" /></el-form-item>
        <el-form-item label="标签名"><el-input v-model="labelName" placeholder="instance" /></el-form-item>
        <template v-if="mode === 'range'">
          <el-form-item label="窗口">
            <el-input-number v-model="rangeMinutes" :min="1" :max="1440" />
            <span class="hint">分钟</span>
            <el-input-number v-model="step" :min="1" :max="3600" />
            <span class="hint">step 秒</span>
          </el-form-item>
        </template>
        <el-form-item>
          <el-button type="primary" :loading="running" @click="run">执行</el-button>
          <el-button @click="loadLabels">labels</el-button>
          <el-button @click="loadLabelValues">label-values</el-button>
          <el-button @click="loadMetrics">metrics</el-button>
        </el-form-item>
      </el-form>
      <el-alert v-if="runError" :title="runError" type="error" show-icon :closable="false" />
      <pre v-if="result" class="result">{{ clipped(result) }}</pre>
      <el-empty v-else description="暂无查询结果" />
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { monitorApi, normalizeList, safeJson, redactText, isPermissionError, isUnauthorizedError } from '../../api/monitor'

const sources = ref([])
const loading = ref(false)
const error = ref('')
const running = ref(false)
const runError = ref('')
const result = ref(null)
const mode = ref('instant')
const rangeMinutes = ref(30)
const step = ref(30)
const labelName = ref('instance')
const testMsg = ref('')
const testOK = ref(false)
const form = reactive({ datasource_id: '', query: 'up', timeout_seconds: 5 })

const safe = value => redactText(value)
const permissionMessage = error => isUnauthorizedError(error) ? '登录已过期，请重新登录后继续访问监控能力' : '无权限访问该监控能力'
const formatError = error => isPermissionError(error) ? permissionMessage(error) : redactText(error?.message || '请求失败')

const load = async () => {
  loading.value = true
  error.value = ''
  try {
    sources.value = normalizeList(await monitorApi.datasources())
    const def = sources.value.find(source => source.default) || sources.value[0]
    if (!form.datasource_id && def) form.datasource_id = def.id
  } catch (e) {
    error.value = formatError(e)
  } finally {
    loading.value = false
  }
}

const pick = row => {
  if (row?.id) form.datasource_id = row.id
}

const test = async row => {
  testMsg.value = ''
  testOK.value = false
  try {
    const data = await monitorApi.testDatasource(row.id)
    testOK.value = true
    testMsg.value = `连接成功，series=${data.stats?.series_count ?? '-'}`
  } catch (e) {
    testMsg.value = formatError(e)
  }
}

const buildQueryBody = () => {
  const now = Math.floor(Date.now() / 1000)
  const body = { ...form }
  if (mode.value === 'range') Object.assign(body, { start: now - rangeMinutes.value * 60, end: now, step: step.value })
  return body
}

const run = async () => {
  running.value = true
  runError.value = ''
  result.value = null
  try {
    result.value = mode.value === 'range' ? await monitorApi.queryRange(buildQueryBody()) : await monitorApi.query(buildQueryBody())
  } catch (e) {
    runError.value = formatError(e)
  } finally {
    running.value = false
  }
}

const loadLabels = async () => {
  try { result.value = await monitorApi.labels({ datasource_id: form.datasource_id, limit: 200 }) } catch (e) { runError.value = formatError(e) }
}
const loadLabelValues = async () => {
  try { result.value = await monitorApi.labelValues({ datasource_id: form.datasource_id, label: labelName.value, limit: 500 }) } catch (e) { runError.value = formatError(e) }
}
const loadMetrics = async () => {
  try { result.value = await monitorApi.metrics({ datasource_id: form.datasource_id, limit: 50 }) } catch (e) { runError.value = formatError(e) }
}
const clipped = data => safeJson(data, 12000)

defineExpose({ load })
onMounted(load)
</script>

<style scoped>
.query-layout { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 12px; }
.box { padding: 14px; border: 1px solid #e4e9f2; border-radius: 8px; background: #fff; }
.bar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
.hint { margin: 0 10px 0 6px; color: #6b778c; font-size: 12px; }
.result { max-height: 360px; overflow: auto; padding: 12px; border-radius: 8px; background: #f5f7fb; color: #263653; white-space: pre-wrap; word-break: break-word; font-size: 12px; }
@media (max-width: 520px) { .query-layout { grid-template-columns: 1fr; } }
</style>
