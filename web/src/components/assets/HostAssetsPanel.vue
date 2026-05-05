<template>
  <section class="asset-panel" v-loading="loading">
    <div class="toolbar">
      <div>
        <h3>主机资产</h3>
        <p>按业务空间、资源组、标签和在线状态筛选主机资产。</p>
      </div>
      <el-button @click="loadHosts">刷新</el-button>
    </div>
    <div class="filters">
      <el-input v-model="query.workspace_id" clearable placeholder="业务空间 ID" @change="loadHosts" />
      <el-input v-model="query.resource_group_id" clearable placeholder="资源组 ID" @change="loadHosts" />
      <el-input v-model="query.tag" clearable placeholder="标签" @change="loadHosts" />
      <el-select v-model="query.online" clearable placeholder="在线状态" @change="loadHosts">
        <el-option label="在线" value="true" />
        <el-option label="离线" value="false" />
      </el-select>
      <el-button type="primary" @click="loadHosts">查询</el-button>
    </div>
    <el-alert v-if="permissionError" class="state-alert" type="warning" show-icon :closable="false" title="权限不足" description="当前账号无法访问主机资产，请联系管理员开通权限。" />
    <el-alert v-else-if="errorText" class="state-alert" type="error" show-icon :closable="false" title="加载失败" :description="errorText" />
    <el-empty v-if="!loading && !hosts.length && !errorText && !permissionError" class="empty-box" description="暂无主机资产" />
    <el-table v-else :data="hosts" class="asset-table" stripe :row-key="hostKey" empty-text="暂无主机资产">
      <el-table-column label="主机" min-width="190" show-overflow-tooltip>
        <template #default="{ row }">
          <b>{{ safe(nameOf(row)) || '-' }}</b>
          <div class="minor">{{ safe(ipOf(row)) || '-' }}</div>
        </template>
      </el-table-column>
      <el-table-column label="在线" width="90">
        <template #default="{ row }"><el-tag size="small" :type="onlineOf(row) ? 'success' : 'info'">{{ onlineOf(row) ? '在线' : '离线' }}</el-tag></template>
      </el-table-column>
      <el-table-column label="业务空间" min-width="150" show-overflow-tooltip>
        <template #default="{ row }">{{ safe(row.workspace_name || row.workspace_id) || '-' }}</template>
      </el-table-column>
      <el-table-column label="资源组" min-width="150" show-overflow-tooltip>
        <template #default="{ row }">{{ safe(row.resource_group_name || row.resource_group_id) || '-' }}</template>
      </el-table-column>
      <el-table-column label="Agent" width="110">
        <template #default="{ row }">{{ safe(row.agent_id || row.agent_ident || row.agent_status) || '-' }}</template>
      </el-table-column>
      <el-table-column label="标签" min-width="180">
        <template #default="{ row }">
          <div class="tag-list"><el-tag v-for="tag in tags(row)" :key="tag" size="small" effect="plain">{{ safe(tag) }}</el-tag><span v-if="!tags(row).length" class="muted">-</span></div>
        </template>
      </el-table-column>
      <el-table-column label="最近存活" width="170"><template #default="{ row }">{{ fmt(row.last_seen || row.last_heartbeat_at) }}</template></el-table-column>
      <el-table-column label="操作" width="190" fixed="right">
        <template #default="{ row }">
          <el-button text size="small" :loading="actionId === `${hostKey(row)}:tags`" @click="openTags(row)">标签</el-button>
          <el-button text size="small" :loading="actionId === `${hostKey(row)}:group`" @click="openBind(row, 'group')">资源组</el-button>
          <el-button text size="small" :loading="actionId === `${hostKey(row)}:workspace`" @click="openBind(row, 'workspace')">业务空间</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="tagDialog" title="编辑标签" width="460px">
      <el-form label-position="top"><el-form-item label="标签"><el-input v-model="tagText" type="textarea" :rows="4" placeholder="使用英文逗号或换行分隔" /></el-form-item></el-form>
      <template #footer><el-button @click="tagDialog = false">取消</el-button><el-button type="primary" :loading="saving" :disabled="saving" @click="saveTags">保存</el-button></template>
    </el-dialog>
    <el-dialog v-model="bindDialog" :title="bindType === 'group' ? '绑定资源组' : '绑定业务空间'" width="440px">
      <el-form label-position="top">
        <el-form-item :label="bindType === 'group' ? '资源组 ID' : '业务空间 ID'" required><el-input v-model="bindValue" clearable /></el-form-item>
      </el-form>
      <template #footer><el-button @click="bindDialog = false">取消</el-button><el-button type="primary" :loading="saving" :disabled="saving" @click="saveBind">保存</el-button></template>
    </el-dialog>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { assetsApi, firstValue, formatAssetError, isPermissionError, normalizeArray, normalizeList, redactText, splitTags } from '../../api/assets'

const emit = defineEmits(['count-change', 'online-change'])
const loading = ref(false)
const saving = ref(false)
const hosts = ref([])
const errorText = ref('')
const permissionError = ref(false)
const tagDialog = ref(false)
const bindDialog = ref(false)
const actionId = ref('')
const selected = ref(null)
const tagText = ref('')
const bindType = ref('group')
const bindValue = ref('')
const query = reactive({ workspace_id: '', resource_group_id: '', tag: '', online: '' })

const safe = value => redactText(value)
const tags = row => normalizeArray(row?.tags)
const hostKey = row => row?.host_id || row?.id || row?.ident || ''
const nameOf = row => firstValue(row, ['hostname', 'host_name', 'name', 'ident', 'id'])
const ipOf = row => firstValue(row, ['ip', 'host_ip', 'private_ip', 'address'])
const onlineOf = row => row.online === true || row.status === 'online' || row.agent_status === 'online'
const fmt = value => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'

const loadHosts = async () => {
  loading.value = true
  errorText.value = ''
  permissionError.value = false
  try {
    hosts.value = normalizeList(await assetsApi.listHostAssets(query))
    emit('count-change', hosts.value.length)
    emit('online-change', hosts.value.filter(onlineOf).length)
  } catch (error) {
    hosts.value = []
    permissionError.value = isPermissionError(error)
    errorText.value = formatAssetError(error, '主机资产')
    emit('count-change', 0)
    emit('online-change', 0)
  } finally {
    loading.value = false
  }
}

const openTags = row => {
  selected.value = row
  tagText.value = tags(row).join(', ')
  tagDialog.value = true
}

const openBind = (row, type) => {
  selected.value = row
  bindType.value = type
  bindValue.value = type === 'group' ? (row.resource_group_id || '') : (row.workspace_id || '')
  bindDialog.value = true
}

const saveTags = async () => {
  const key = hostKey(selected.value)
  if (!key) return
  saving.value = true
  actionId.value = `${key}:tags`
  try {
    await assetsApi.updateHostTags(key, splitTags(tagText.value))
    ElMessage.success('主机标签已保存')
    tagDialog.value = false
    await loadHosts()
  } catch (error) {
    ElMessage.error(formatAssetError(error, '主机标签'))
  } finally {
    saving.value = false
    actionId.value = ''
  }
}

const saveBind = async () => {
  const key = hostKey(selected.value)
  if (!key || !bindValue.value.trim()) return ElMessage.warning('请输入绑定 ID')
  saving.value = true
  actionId.value = `${key}:${bindType.value}`
  try {
    if (bindType.value === 'group') await assetsApi.updateHostResourceGroup(key, bindValue.value.trim())
    else await assetsApi.updateHostWorkspace(key, bindValue.value.trim())
    ElMessage.success('绑定关系已保存')
    bindDialog.value = false
    await loadHosts()
  } catch (error) {
    ElMessage.error(formatAssetError(error, bindType.value === 'group' ? '资源组绑定' : '业务空间绑定'))
  } finally {
    saving.value = false
    actionId.value = ''
  }
}

onMounted(loadHosts)
</script>

<style scoped>
.asset-panel { display: flex; flex-direction: column; gap: 16px; }
.toolbar { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.toolbar h3 { margin: 0; color: #1e3a5f; font-size: 18px; }
.toolbar p, .minor { margin: 6px 0 0; color: #60728e; font-size: 12px; }
.filters { display: grid; grid-template-columns: repeat(5, minmax(130px, 1fr)); gap: 10px; }
.state-alert, .asset-table, .empty-box { border-radius: 8px; }
.asset-table { border: 1px solid #e4e9f2; overflow: hidden; }
.empty-box { min-height: 300px; border: 1px dashed #d8e1ee; background: #f8fbff; }
.tag-list { display: flex; flex-wrap: wrap; gap: 6px; }
.muted { color: #8a99ad; }
@media (max-width: 920px) { .filters { grid-template-columns: 1fr 1fr; } }
@media (max-width: 640px) { .toolbar { flex-direction: column; } .filters { grid-template-columns: 1fr; } }
</style>
