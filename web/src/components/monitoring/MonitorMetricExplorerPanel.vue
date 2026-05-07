<template>
  <div class="metric-explorer">
    <section class="explorer-toolbar">
      <div class="toolbar-main">
        <el-select
          v-model="datasourceId"
          class="datasource-select"
          filterable
          placeholder="选择数据源"
          :loading="loadingSources"
          @change="handleDatasourceChange"
        >
          <el-option
            v-for="source in datasources"
            :key="source.id"
            :label="sourceLabel(source)"
            :value="source.id"
          />
        </el-select>
        <el-button size="small" :loading="loadingSources" @click="loadDatasources">刷新</el-button>
        <el-switch
          v-model="autocompleteEnabled"
          active-text="Enable autocomplete"
          inactive-text="Autocomplete off"
          inline-prompt
        />
      </div>
      <div class="toolbar-actions">
        <el-button size="small" plain @click="showBlocked('保存视图')">保存视图</el-button>
        <el-button size="small" plain @click="showBlocked('分享图表')">分享</el-button>
        <el-button size="small" plain @click="showBlocked('AI query 生成')">AI query</el-button>
        <el-tag type="warning" effect="plain">BLOCKED_BY_CONTRACT</el-tag>
      </div>
    </section>

    <el-alert
      v-if="sourceError"
      :title="sourceError"
      type="error"
      show-icon
      :closable="false"
      class="top-alert"
    />

    <section
      v-for="(panel, index) in panels"
      :key="panel.id"
      class="query-card"
    >
      <div class="card-head">
        <div>
          <strong>Panel {{ index + 1 }}</strong>
          <span class="muted">PromQL</span>
        </div>
        <div class="card-actions">
          <el-popover width="360" trigger="click" @show="panel.historySearch = ''">
            <template #reference>
              <el-button size="small">历史记录</el-button>
            </template>
            <div class="popover-title">历史记录</div>
            <el-input v-model="panel.historySearch" size="small" placeholder="搜索 PromQL" clearable />
            <div class="history-list">
              <button
                v-for="item in filteredHistory(panel)"
                :key="item"
                class="plain-row"
                type="button"
                @click="setQuery(panel, item)"
              >
                {{ item }}
              </button>
              <el-empty v-if="filteredHistory(panel).length === 0" description="暂无历史" :image-size="56" />
            </div>
          </el-popover>
          <el-popover width="440" trigger="click" @show="loadBuiltinMetrics(panel)">
            <template #reference>
              <el-button size="small">内置指标</el-button>
            </template>
            <div class="popover-title">内置指标</div>
            <div class="metric-search">
              <el-input
                v-model="panel.metricSearch"
                size="small"
                placeholder="搜索指标名、说明或 exporter"
                clearable
                @input="debouncedBuiltin(panel)"
              />
              <el-button size="small" :loading="panel.metricsLoading" @click="loadBuiltinMetrics(panel)">搜索</el-button>
            </div>
            <div class="metric-list">
              <button
                v-for="metric in panel.builtinMetrics"
                :key="metric.id || metric.raw_name || metric.standard_name"
                class="metric-row"
                type="button"
                @click="insertMetric(panel, metric)"
              >
                <b>{{ safe(metric.raw_name || metric.standard_name || metric.promql) }}</b>
                <span>{{ safe(metric.standard_name || metric.description || '-') }}</span>
              </button>
              <el-empty v-if="!panel.metricsLoading && panel.builtinMetrics.length === 0" description="暂无指标" :image-size="56" />
            </div>
          </el-popover>
          <el-button v-if="panels.length > 1" size="small" type="danger" plain @click="removePanel(panel.id)">关闭</el-button>
        </div>
      </div>

      <div class="query-editor">
        <el-input
          v-model="panel.query"
          type="textarea"
          :rows="3"
          resize="vertical"
          placeholder="up"
          @input="handleQueryInput(panel)"
        />
        <div v-if="autocompleteEnabled && panel.suggestions.length > 0" class="suggestions">
          <button
            v-for="item in panel.suggestions"
            :key="item.id || item.raw_name || item.standard_name || item.promql"
            class="suggestion"
            type="button"
            @click="insertMetric(panel, item)"
          >
            <b>{{ safe(item.raw_name || item.standard_name || item.promql) }}</b>
            <span>{{ safe(item.description || item.standard_name || '') }}</span>
          </button>
        </div>
      </div>

      <el-tabs v-model="panel.activeTab" class="result-tabs">
        <el-tab-pane label="Table" name="table">
          <div class="mode-bar">
            <label>
              <span>Time</span>
              <input v-model="panel.instantTime" class="native-input" type="datetime-local" />
            </label>
            <label>
              <span>Unit</span>
              <el-select v-model="panel.unit" size="small" class="unit-select">
                <el-option label="none" value="" />
                <el-option label="percent" value="%" />
                <el-option label="seconds" value="s" />
                <el-option label="bytes" value="bytes" />
                <el-option label="requests/sec" value="req/s" />
              </el-select>
            </label>
            <el-button type="primary" size="small" :loading="panel.loading" @click="runInstant(panel)">查询</el-button>
            <el-button size="small" :disabled="tableRows(panel).length === 0" @click="exportCsv(panel)">CSV export</el-button>
          </div>
          <el-alert v-if="panel.error" :title="panel.error" type="error" show-icon :closable="false" />
          <div v-if="panel.instantResult" class="stats-line">
            <el-tag size="small">series {{ panel.instantResult?.stats?.series_count ?? '-' }}</el-tag>
            <el-tag size="small">samples {{ panel.instantResult?.stats?.sample_count ?? '-' }}</el-tag>
            <el-tag size="small">latency {{ panel.instantResult?.latency_ms ?? '-' }} ms</el-tag>
          </div>
          <el-table v-if="tableRows(panel).length" :data="tableRows(panel)" size="small" border max-height="280">
            <el-table-column prop="metric" label="Metric" min-width="220" show-overflow-tooltip />
            <el-table-column prop="value" label="Value" width="180" show-overflow-tooltip />
            <el-table-column prop="time" label="Time" width="180" show-overflow-tooltip />
          </el-table>
          <pre v-if="panel.instantResult" class="json-box">{{ clipped(panel.instantResult) }}</pre>
          <el-empty v-if="!panel.instantResult && !panel.loading" description="暂无查询结果" />
        </el-tab-pane>

        <el-tab-pane label="Graph" name="graph">
          <div class="mode-bar graph-controls">
            <label>
              <span>Start</span>
              <input v-model="panel.rangeStart" class="native-input" type="datetime-local" />
            </label>
            <label>
              <span>End</span>
              <input v-model="panel.rangeEnd" class="native-input" type="datetime-local" />
            </label>
            <label>
              <span>Max data points</span>
              <el-input-number v-model="panel.maxDataPoints" size="small" :min="10" :max="11000" />
            </label>
            <label>
              <span>Min step</span>
              <el-input-number v-model="panel.minStep" size="small" :min="1" :max="3600" />
            </label>
            <el-segmented v-model="panel.graphMode" :options="graphOptions" size="small" />
            <el-button type="primary" size="small" :loading="panel.loading" @click="runRange(panel)">查询</el-button>
            <el-button size="small" plain @click="showBlocked('完整图表设置')">设置</el-button>
          </div>
          <el-alert
            title="BLOCKED_BY_CONTRACT: 完整时序图渲染等待图表组件契约；当前仅展示真实 query_range 返回摘要和脱敏 JSON。"
            type="warning"
            show-icon
            :closable="false"
          />
          <el-alert v-if="panel.error" :title="panel.error" type="error" show-icon :closable="false" class="inline-alert" />
          <div v-if="panel.rangeResult" class="stats-line">
            <el-tag size="small">series {{ panel.rangeResult?.stats?.series_count ?? '-' }}</el-tag>
            <el-tag size="small">samples {{ panel.rangeResult?.stats?.sample_count ?? '-' }}</el-tag>
            <el-tag size="small">latency {{ panel.rangeResult?.latency_ms ?? '-' }} ms</el-tag>
            <el-tag size="small">{{ panel.graphMode }}</el-tag>
          </div>
          <pre v-if="panel.rangeResult" class="json-box">{{ clipped(panel.rangeResult) }}</pre>
          <el-empty v-if="!panel.rangeResult && !panel.loading" description="暂无区间查询结果" />
        </el-tab-pane>
      </el-tabs>
    </section>

    <div class="add-panel">
      <el-button type="primary" plain @click="addPanel">添加面板</el-button>
    </div>
  </div>
</template>

<script setup>
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, onMounted, ref, watch } from 'vue'
import { monitorApi, normalizeList, safeJson, redactText, isPermissionError, isUnauthorizedError } from '../../api/monitor'

const graphOptions = ['Line', 'StackArea']
const datasources = ref([])
const datasourceId = ref('')
const loadingSources = ref(false)
const sourceError = ref('')
const autocompleteEnabled = ref(true)
const panels = ref([createPanel()])
const builtinTimers = new Map()
const suggestTimers = new Map()
const historyKey = computed(() => `n9e-query-promql-history-${datasourceId.value || 'default'}`)
const historyItems = ref([])
const promqlTypeFields = ['type', 'plugin_type', 'plugin_type_name', 'datasource_type', 'category']
const promqlSourceFields = [...promqlTypeFields, 'id', 'name']
const promqlSourcePattern = /(^|[^a-z0-9])(prometheus|promql|victoriametrics|victoria[-_\s]?metrics|vmselect)([^a-z0-9]|$)/i
const nonPromqlSourcePattern = /(^|[^a-z0-9])(mysql|mariadb|postgres|postgresql|clickhouse|elasticsearch|influxdb|loki|tempo|jaeger|skywalking|signoz)([^a-z0-9]|$)/i
const brandDisplayRules = [
  [/AI\s*WorkBench/gi, 'FindX'],
  [/AIOps/gi, 'FindX AI SRE'],
  [/Nightingale/gi, 'FindX AI SRE'],
  [/\bn9e\b/gi, 'FindX AI SRE'],
  [/SkyWalking/gi, 'FindX AI SRE'],
  [/SigNoZ/gi, 'FindX AI SRE'],
  [/Categraf/gi, 'FindX Agent'],
  [/Catpaw/gi, 'FindX Agent'],
]

function createPanel() {
  const now = new Date()
  const start = new Date(now.getTime() - 60 * 60 * 1000)
  return {
    id: `panel-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    query: 'up',
    activeTab: 'table',
    instantTime: toLocalInput(now),
    rangeStart: toLocalInput(start),
    rangeEnd: toLocalInput(now),
    unit: '',
    graphMode: 'Line',
    maxDataPoints: 600,
    minStep: 15,
    loading: false,
    error: '',
    instantResult: null,
    rangeResult: null,
    suggestions: [],
    builtinMetrics: [],
    metricsLoading: false,
    metricSearch: '',
    historySearch: '',
  }
}

const displayText = value => brandDisplayRules.reduce(
  (text, [pattern, replacement]) => text.replace(pattern, replacement),
  String(value ?? ''),
)
const safe = value => displayText(redactText(value ?? ''))
const clipped = data => displayText(safeJson(data, 12000))
const permissionMessage = error => isUnauthorizedError(error)
  ? '登录已过期，请重新登录后继续访问监控能力'
  : '无权限访问该监控能力'
const formatError = error => isPermissionError(error)
  ? permissionMessage(error)
  : safe(error?.message || '请求失败')
const sourceLabel = source => [source.name, source.type, source.id].filter(Boolean).map(safe).join(' / ')

function sourceFieldValue(source, field) {
  const value = source?.[field]
  return value === null || value === undefined ? '' : String(value)
}

function sourceSearchText(source) {
  return promqlSourceFields.map(field => sourceFieldValue(source, field)).filter(Boolean).join(' ')
}

function isPromqlDatasource(source) {
  const typeText = promqlTypeFields.map(field => sourceFieldValue(source, field)).filter(Boolean).join(' ')
  const hasPromqlType = promqlSourcePattern.test(typeText)
  const hasPromqlSignal = hasPromqlType || promqlSourcePattern.test(sourceSearchText(source))
  const hasNonPromqlType = nonPromqlSourcePattern.test(typeText)
  const hasNonPromqlNameOnly = !hasPromqlType && nonPromqlSourcePattern.test(sourceSearchText(source))
  return hasPromqlSignal && !hasNonPromqlType && !hasNonPromqlNameOnly
}

function toLocalInput(date) {
  const pad = value => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function fromLocalInput(value, fallback = Date.now()) {
  const time = value ? new Date(value).getTime() : fallback
  return Math.floor((Number.isFinite(time) ? time : fallback) / 1000)
}

function handleDatasourceChange() {
  loadHistory()
  panels.value.forEach(panel => {
    panel.suggestions = []
    panel.builtinMetrics = []
    panel.instantResult = null
    panel.rangeResult = null
    panel.error = ''
  })
}

async function loadDatasources() {
  loadingSources.value = true
  sourceError.value = ''
  try {
    datasources.value = normalizeList(await monitorApi.datasources()).filter(isPromqlDatasource)
    const current = datasources.value.find(item => item.id === datasourceId.value)
    const def = current || datasources.value.find(item => item.default) || datasources.value[0]
    if (def?.id) {
      datasourceId.value = def.id
    } else {
      datasourceId.value = ''
      sourceError.value = '未发现 Prometheus 或 VictoriaMetrics 等 PromQL 兼容数据源，请先配置指标查询数据源'
    }
    loadHistory()
  } catch (error) {
    sourceError.value = formatError(error)
  } finally {
    loadingSources.value = false
  }
}

function loadHistory() {
  try {
    const raw = JSON.parse(localStorage.getItem(historyKey.value) || '[]')
    historyItems.value = Array.isArray(raw) ? raw.filter(Boolean).slice(0, 100).map(item => redactText(item)) : []
  } catch {
    historyItems.value = []
  }
}

function saveHistory(query) {
  const value = String(query || '').trim()
  if (!value || !datasourceId.value) return
  const next = [value, ...historyItems.value.filter(item => item !== value)].slice(0, 100)
  historyItems.value = next
  localStorage.setItem(historyKey.value, JSON.stringify(next))
}

function filteredHistory(panel) {
  const q = panel.historySearch.trim().toLowerCase()
  return historyItems.value.filter(item => !q || item.toLowerCase().includes(q)).slice(0, 100)
}

function setQuery(panel, query) {
  panel.query = query
  panel.suggestions = []
}

function insertMetric(panel, metric) {
  const query = metric.promql || metric.raw_name || metric.standard_name
  if (query) setQuery(panel, redactText(query))
}

function addPanel() {
  panels.value.push(createPanel())
}

function removePanel(id) {
  if (panels.value.length > 1) panels.value = panels.value.filter(panel => panel.id !== id)
}

function requestBody(panel) {
  return {
    datasource_id: datasourceId.value,
    query: panel.query.trim(),
    timeout_seconds: 10,
  }
}

function ensureRunnable(panel) {
  panel.error = ''
  if (!datasourceId.value) {
    panel.error = '请先选择数据源'
    return false
  }
  if (!panel.query.trim()) {
    panel.error = '请输入 PromQL'
    return false
  }
  return true
}

async function runInstant(panel) {
  if (!ensureRunnable(panel)) return
  panel.loading = true
  panel.instantResult = null
  try {
    panel.instantResult = await monitorApi.query({
      ...requestBody(panel),
      time: fromLocalInput(panel.instantTime),
    })
    saveHistory(panel.query)
  } catch (error) {
    panel.error = formatError(error)
  } finally {
    panel.loading = false
  }
}

async function runRange(panel) {
  if (!ensureRunnable(panel)) return
  const start = fromLocalInput(panel.rangeStart, Date.now() - 60 * 60 * 1000)
  const end = fromLocalInput(panel.rangeEnd)
  const step = Math.max(panel.minStep, Math.ceil((end - start) / Math.max(panel.maxDataPoints, 1)))
  panel.loading = true
  panel.rangeResult = null
  try {
    panel.rangeResult = await monitorApi.queryRange({
      ...requestBody(panel),
      start,
      end,
      step,
    })
    saveHistory(panel.query)
  } catch (error) {
    panel.error = formatError(error)
  } finally {
    panel.loading = false
  }
}

function handleQueryInput(panel) {
  if (!autocompleteEnabled.value) {
    panel.suggestions = []
    return
  }
  window.clearTimeout(suggestTimers.get(panel.id))
  suggestTimers.set(panel.id, window.setTimeout(() => loadSuggestions(panel), 260))
}

async function loadSuggestions(panel) {
  if (!autocompleteEnabled.value || !datasourceId.value) return
  const q = panel.query.trim().split(/[\s{(\[]/).pop() || ''
  if (q.length < 2) {
    panel.suggestions = []
    return
  }
  try {
    panel.suggestions = normalizeList(await monitorApi.metrics({ datasource_id: datasourceId.value, q, limit: 8 }))
  } catch (error) {
    panel.error = formatError(error)
    panel.suggestions = []
  }
}

function debouncedBuiltin(panel) {
  window.clearTimeout(builtinTimers.get(panel.id))
  builtinTimers.set(panel.id, window.setTimeout(() => loadBuiltinMetrics(panel), 260))
}

async function loadBuiltinMetrics(panel) {
  if (!datasourceId.value) {
    panel.error = '请先选择数据源'
    return
  }
  panel.metricsLoading = true
  try {
    panel.builtinMetrics = normalizeList(await monitorApi.metrics({
      datasource_id: datasourceId.value,
      q: panel.metricSearch,
      limit: 20,
    }))
  } catch (error) {
    panel.error = formatError(error)
    panel.builtinMetrics = []
  } finally {
    panel.metricsLoading = false
  }
}

function tableRows(panel) {
  const result = panel.instantResult?.data?.result
  if (!Array.isArray(result)) return []
  return result.map(row => {
    const value = Array.isArray(row.value) ? row.value : []
    return {
      metric: displayText(safeJson(row.metric || {}, 1200)),
      value: `${safe(value[1] ?? '')}${panel.unit ? ` ${panel.unit}` : ''}`,
      time: value[0] ? new Date(Number(value[0]) * 1000).toLocaleString() : '-',
    }
  })
}

function exportCsv(panel) {
  const rows = tableRows(panel)
  if (!rows.length) return
  const header = ['Metric', 'Value', 'Time']
  const csv = [header, ...rows.map(row => [row.metric, row.value, row.time])]
    .map(cols => cols.map(value => `"${String(value ?? '').replaceAll('"', '""')}"`).join(','))
    .join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `metric-query-${Date.now()}.csv`
  link.click()
  URL.revokeObjectURL(url)
}

function showBlocked(name) {
  ElMessageBox.alert(
    `${name} BLOCKED_BY_CONTRACT：当前后端契约未提供保存、分享、AI 生成或完整图表设置能力，不能伪造成功。`,
    'BLOCKED_BY_CONTRACT',
    { type: 'warning' }
  ).catch(() => ElMessage.warning('BLOCKED_BY_CONTRACT'))
}

watch(autocompleteEnabled, enabled => {
  if (!enabled) panels.value.forEach(panel => { panel.suggestions = [] })
})

onMounted(loadDatasources)
</script>

<style scoped>
.metric-explorer { display: flex; flex-direction: column; gap: 14px; }
.explorer-toolbar,
.query-card { border: 1px solid #e1e7f0; border-radius: 8px; background: #fff; }
.explorer-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 12px; }
.toolbar-main,
.toolbar-actions,
.card-head,
.card-actions,
.mode-bar,
.metric-search,
.stats-line { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.datasource-select { width: min(420px, 100%); }
.toolbar-actions { justify-content: flex-end; }
.top-alert { margin-top: -4px; }
.query-card { padding: 12px; }
.card-head { justify-content: space-between; margin-bottom: 10px; }
.card-head strong { color: #1f3557; font-size: 14px; }
.muted { margin-left: 8px; color: #75839a; font-size: 12px; }
.query-editor { position: relative; }
.suggestions { margin-top: 6px; border: 1px solid #dce4ef; border-radius: 8px; background: #fbfdff; overflow: hidden; }
.suggestion,
.plain-row,
.metric-row { width: 100%; border: 0; background: transparent; text-align: left; cursor: pointer; }
.suggestion { display: grid; grid-template-columns: minmax(140px, 240px) 1fr; gap: 8px; padding: 7px 10px; color: #243553; }
.plain-row { padding: 8px 6px; color: #243553; border-bottom: 1px solid #edf1f7; font-family: ui-monospace, SFMono-Regular, Consolas, monospace; font-size: 12px; }
.metric-row { display: grid; gap: 4px; padding: 9px 6px; border-bottom: 1px solid #edf1f7; color: #243553; }
.metric-row span,
.suggestion span { color: #75839a; font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.suggestion:hover,
.plain-row:hover,
.metric-row:hover { background: #eef5ff; }
.result-tabs { margin-top: 8px; }
.mode-bar { min-height: 34px; margin-bottom: 10px; }
.mode-bar label { display: inline-flex; align-items: center; gap: 6px; color: #60728e; font-size: 12px; }
.native-input { height: 24px; border: 1px solid #dcdfe6; border-radius: 4px; padding: 2px 6px; color: #243553; font-size: 12px; }
.unit-select { width: 130px; }
.graph-controls { align-items: center; }
.stats-line { margin: 10px 0; }
.json-box { max-height: 300px; overflow: auto; padding: 10px; border-radius: 8px; background: #f6f8fb; color: #243553; font-size: 12px; line-height: 1.55; white-space: pre-wrap; word-break: break-word; }
.history-list,
.metric-list { max-height: 280px; overflow: auto; margin-top: 8px; }
.popover-title { color: #1f3557; font-weight: 700; margin-bottom: 8px; }
.inline-alert { margin-top: 10px; }
.add-panel { display: flex; justify-content: center; padding: 6px 0 2px; }
@media (max-width: 720px) {
  .explorer-toolbar,
  .card-head { align-items: stretch; flex-direction: column; }
  .toolbar-main,
  .toolbar-actions,
  .card-actions,
  .mode-bar { align-items: stretch; flex-direction: column; }
  .datasource-select,
  .unit-select { width: 100%; }
  .suggestion { grid-template-columns: 1fr; }
}
</style>
