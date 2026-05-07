<template>
  <div class="datasource-panel">
    <section class="datasource-box">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="keyword"
            class="search-input"
            clearable
            placeholder="搜索数据源名称、类型、集群或地址"
            :prefix-icon="Search"
          />
          <el-select v-model="typeFilter" clearable placeholder="类型" class="filter-select">
            <el-option v-for="type in typeOptions" :key="type" :label="safe(type)" :value="type" />
          </el-select>
          <el-select v-model="statusFilter" clearable placeholder="状态" class="filter-select">
            <el-option label="启用" value="enabled" />
            <el-option label="禁用" value="disabled" />
            <el-option label="未知" value="unknown" />
          </el-select>
        </div>
        <div class="toolbar-actions">
          <el-tooltip content="刷新">
            <el-button :icon="Refresh" :loading="loading" @click="load" />
          </el-tooltip>
          <el-button type="primary" :icon="Plus" @click="typeDialogVisible = true">新增</el-button>
        </div>
      </div>

      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" class="state-alert" />
      <el-alert
        v-if="blockedMessage"
        :title="blockedMessage"
        type="warning"
        show-icon
        closable
        class="state-alert"
        @close="blockedMessage = ''"
      />

      <el-table
        :data="pagedSources"
        v-loading="loading"
        row-key="__key"
        height="460"
        empty-text="暂无真实数据源"
        border
        stripe
      >
        <el-table-column label="类型" width="170" sortable>
          <template #default="{ row }">
            <div class="type-cell">
              <el-icon><Connection /></el-icon>
              <span>{{ safe(displayType(row)) || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="名称" min-width="180" sortable show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" class="name-link" @click="openDetail(row)">
              {{ safe(displayName(row)) || '-' }}
            </el-button>
            <el-tag v-if="row.default || row.is_default" size="small" effect="plain">默认</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="集群/来源" min-width="150" sortable show-overflow-tooltip>
          <template #default="{ row }">{{ safe(displayCluster(row)) || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="120" sortable>
          <template #default="{ row }">
            <el-tag size="small" :type="statusTag(row)">{{ statusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="地址/说明" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">{{ safe(displayAddress(row)) || safe(displayDescription(row)) || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-tooltip content="测试连接">
              <el-button link type="primary" :loading="testingId === rowKey(row)" @click="test(row)">
                测试
              </el-button>
            </el-tooltip>
            <el-tooltip content="编辑">
              <el-button link :icon="Edit" @click="showBlocked('编辑数据源')" />
            </el-tooltip>
            <el-tooltip :content="isEnabled(row) ? '停用' : '启用'">
              <el-button link :icon="SwitchButton" @click="showBlocked(isEnabled(row) ? '停用数据源' : '启用数据源')" />
            </el-tooltip>
            <el-tooltip content="删除">
              <el-button link type="danger" :icon="Delete" @click="showBlocked('删除数据源')" />
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>

      <div class="table-footer">
        <span>共 {{ filteredSources.length }} 条</span>
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          layout="sizes, prev, pager, next"
          :page-sizes="[10, 20, 50]"
          :total="filteredSources.length"
          size="small"
          background
        />
      </div>

      <el-alert v-if="testMsg" :title="testMsg" :type="testOK ? 'success' : 'error'" show-icon class="state-alert" />
    </section>

    <el-drawer v-model="detailVisible" title="数据源详情" size="520px" destroy-on-close>
      <template v-if="selected">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="名称">{{ safe(displayName(selected)) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ safe(displayType(selected)) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="集群/来源">{{ safe(displayCluster(selected)) || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ statusText(selected) }}</el-descriptions-item>
          <el-descriptions-item label="地址/说明">
            {{ safe(displayAddress(selected)) || safe(displayDescription(selected)) || '-' }}
          </el-descriptions-item>
        </el-descriptions>
        <div class="drawer-actions">
          <el-button :icon="View" :loading="testingId === rowKey(selected)" @click="test(selected)">测试连接</el-button>
          <el-button :icon="Edit" @click="showBlocked('编辑数据源')">编辑</el-button>
        </div>
        <pre class="json-view">{{ displaySafeJson(selected, 10000) }}</pre>
      </template>
    </el-drawer>

    <el-dialog v-model="typeDialogVisible" title="选择数据源类型" width="760px" destroy-on-close>
      <div class="type-grid">
        <button v-for="item in datasourceTypes" :key="item.type" class="type-card" type="button" @click="chooseType(item)">
          <el-icon><component :is="item.icon" /></el-icon>
          <span class="card-title">{{ item.name }}</span>
          <span class="card-desc">{{ item.desc }}</span>
          <el-tag size="small" :type="item.supported ? 'success' : 'info'" effect="plain">
            {{ item.supported ? '支持计划' : '待 Adapter 契约' }}
          </el-tag>
        </button>
      </div>
      <el-alert
        title="BLOCKED_BY_CONTRACT：后端尚未提供 plugin/list、upsert、status/update、delete、server-clusters 等写入契约，新增流程只能展示类型选择，不能提交保存。"
        type="warning"
        show-icon
        :closable="false"
        class="dialog-alert"
      />
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import {
  Connection,
  Delete,
  Document,
  Edit,
  Histogram,
  Plus,
  Refresh,
  Search,
  SwitchButton,
  View,
  Warning,
} from '@element-plus/icons-vue'
import { monitorApi, normalizeList, redactText, safeJson, isPermissionError, isUnauthorizedError } from '../../api/monitor'

const sources = ref([])
const loading = ref(false)
const error = ref('')
const keyword = ref('')
const typeFilter = ref('')
const statusFilter = ref('')
const page = ref(1)
const pageSize = ref(10)
const selected = ref(null)
const detailVisible = ref(false)
const typeDialogVisible = ref(false)
const blockedMessage = ref('')
const testMsg = ref('')
const testOK = ref(false)
const testingId = ref('')

const datasourceTypes = [
  { name: 'Prometheus', type: 'prometheus', desc: '指标查询与连接测试能力按真实读契约接入。', supported: true, icon: Histogram },
  { name: 'VictoriaMetrics', type: 'victoriametrics', desc: '兼容 PromQL 查询，新增保存等待写契约。', supported: true, icon: Histogram },
  { name: '日志存储', type: 'logs', desc: '等待日志 Adapter、鉴权和写入契约。', supported: false, icon: Document },
  { name: '链路存储', type: 'traces', desc: '等待 Trace Adapter、索引和写入契约。', supported: false, icon: Connection },
  { name: '自定义 Adapter', type: 'custom', desc: '等待 Adapter 描述、校验和生命周期契约。', supported: false, icon: Warning },
]

const brandRules = [
  [/\bAI WorkBench\b/gi, 'FindX'],
  [/\bAIOps\b/gi, 'FindX AI SRE'],
  [/\bNightingale\b/gi, 'FindX'],
  [/\bn9e\b/gi, 'FindX'],
  [/\bSkyWalking\b/gi, 'FindX'],
  [/\bSigNoZ\b/gi, 'FindX'],
  [/\bCategraf\b/gi, 'FindX Agent'],
  [/\bCatpaw\b/gi, 'FindX Agent'],
]
const replaceLegacyBrand = value => brandRules.reduce((text, [pattern, replacement]) => text.replace(pattern, replacement), String(value ?? ''))
const displaySafeText = value => replaceLegacyBrand(redactText(value))
const displaySafeJson = (value, maxLength) => replaceLegacyBrand(safeJson(value, maxLength))
const safe = value => displaySafeText(value)
const permissionMessage = error => isUnauthorizedError(error) ? '登录已过期，请重新登录后继续访问监控能力' : '无权限访问该监控能力'
const formatError = error => isPermissionError(error) ? permissionMessage(error) : redactText(error?.message || '请求失败')
const rowKey = row => String(row?.id ?? row?.datasource_id ?? row?.name ?? row?.url ?? '')
const firstValue = (row, keys) => keys.map(key => row?.[key]).find(value => value !== undefined && value !== null && value !== '')
const displayType = row => firstValue(row, ['plugin_type_name', 'plugin_type', 'type', 'datasource_type', 'category'])
const displayName = row => firstValue(row, ['name', 'display_name', 'id', 'datasource_id'])
const displayCluster = row => firstValue(row, ['cluster_name', 'cluster', 'source', 'origin', 'engine'])
const displayAddress = row => firstValue(row, ['url', 'address', 'endpoint', 'base_url', 'remote_write_url'])
const displayDescription = row => firstValue(row, ['description', 'desc', 'remark', 'note'])
const normalizedStatus = row => String(firstValue(row, ['status', 'state', 'enabled']) ?? 'unknown').toLowerCase()
const isEnabled = row => normalizedStatus(row) === 'enabled' || normalizedStatus(row) === 'true' || normalizedStatus(row) === 'active'
const statusText = row => {
  const status = normalizedStatus(row)
  if (status === 'enabled' || status === 'true' || status === 'active') return '启用'
  if (status === 'disabled' || status === 'false' || status === 'inactive') return '禁用'
  return '未知'
}
const statusTag = row => isEnabled(row) ? 'success' : statusText(row) === '禁用' ? 'warning' : 'info'

const typeOptions = computed(() => [...new Set(sources.value.map(displayType).filter(Boolean))])
const filteredSources = computed(() => {
  const text = keyword.value.trim().toLowerCase()
  return sources.value.filter(row => {
    const haystack = [
      displayType(row),
      displayName(row),
      displayCluster(row),
      displayAddress(row),
      displayDescription(row),
      rowKey(row),
    ].map(value => safe(value).toLowerCase()).join(' ')
    const status = statusText(row)
    return (!text || haystack.includes(text))
      && (!typeFilter.value || displayType(row) === typeFilter.value)
      && (!statusFilter.value || (statusFilter.value === 'enabled' && status === '启用') || (statusFilter.value === 'disabled' && status === '禁用') || (statusFilter.value === 'unknown' && status === '未知'))
  })
})
const pagedSources = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return filteredSources.value.slice(start, start + pageSize.value).map((row, index) => ({ ...row, __key: rowKey(row) || `${start + index}` }))
})

watch([keyword, typeFilter, statusFilter, pageSize], () => {
  page.value = 1
})

const load = async () => {
  loading.value = true
  error.value = ''
  testMsg.value = ''
  try {
    sources.value = normalizeList(await monitorApi.datasources())
  } catch (e) {
    error.value = formatError(e)
  } finally {
    loading.value = false
  }
}

const openDetail = row => {
  selected.value = row
  detailVisible.value = true
}

const showBlocked = action => {
  blockedMessage.value = `BLOCKED_BY_CONTRACT：${action} 需要后端提供数据源写入契约，当前不会假成功。`
  ElMessageBox.alert(blockedMessage.value, '契约缺失', { type: 'warning', confirmButtonText: '知道了' })
}

const chooseType = item => {
  showBlocked(`新增 ${item.name}`)
}

const test = async row => {
  const id = row?.id ?? row?.datasource_id
  testMsg.value = ''
  testOK.value = false
  if (!id) {
    testMsg.value = 'BLOCKED_BY_CONTRACT：测试连接需要 datasource_id，当前数据缺少可用标识。'
    return
  }
  testingId.value = rowKey(row)
  try {
    const data = await monitorApi.testDatasource(id)
    testOK.value = true
    testMsg.value = `连接成功，series=${safe(data?.stats?.series_count ?? data?.data?.stats?.series_count ?? '-')}`
  } catch (e) {
    testMsg.value = formatError(e)
  } finally {
    testingId.value = ''
  }
}

defineExpose({ load })
onMounted(load)
</script>

<style scoped>
.datasource-panel { min-width: 0; }
.datasource-box { padding: 14px; border: 1px solid #e4e9f2; border-radius: 8px; background: #fff; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
.toolbar-left { display: flex; align-items: center; gap: 10px; min-width: 0; flex: 1; }
.toolbar-actions { display: flex; align-items: center; gap: 8px; }
.search-input { width: 300px; flex: 0 0 300px; }
.filter-select { width: 140px; }
.state-alert { margin: 10px 0; }
.type-cell { display: inline-flex; align-items: center; gap: 8px; min-width: 0; }
.name-link { max-width: 150px; padding: 0; vertical-align: middle; }
.table-footer { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-top: 12px; color: #6b778c; font-size: 12px; }
.drawer-actions { display: flex; gap: 8px; margin: 14px 0; }
.json-view { max-height: 360px; overflow: auto; padding: 12px; border-radius: 8px; background: #f5f7fb; color: #263653; white-space: pre-wrap; word-break: break-word; font-size: 12px; }
.type-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.type-card { display: grid; grid-template-columns: 28px 1fr auto; grid-template-rows: auto auto; align-items: center; gap: 6px 10px; width: 100%; padding: 12px; border: 1px solid #e4e9f2; border-radius: 8px; background: #fff; color: #243553; text-align: left; cursor: pointer; }
.type-card:hover { border-color: #1769ff; background: #f7fbff; }
.type-card .el-icon { grid-row: 1 / span 2; font-size: 22px; color: #1769ff; }
.card-title { font-weight: 700; }
.card-desc { grid-column: 2 / span 2; color: #60728e; font-size: 12px; line-height: 1.5; }
.dialog-alert { margin-top: 14px; }
@media (max-width: 900px) {
  .toolbar { align-items: stretch; flex-direction: column; }
  .toolbar-left { flex-wrap: wrap; }
  .search-input { width: 100%; flex: 1 1 100%; }
  .filter-select { flex: 1 1 140px; }
  .toolbar-actions { justify-content: flex-end; }
}
@media (max-width: 620px) {
  .type-grid { grid-template-columns: 1fr; }
  .table-footer { align-items: flex-start; flex-direction: column; }
}
</style>
