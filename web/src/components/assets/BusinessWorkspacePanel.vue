<template>
  <section class="business-panel" v-loading="loading">
    <div class="toolbar">
      <div>
        <h3>业务空间</h3>
        <p>按业务边界管理主机、端点、负责人和运行状态。</p>
      </div>
      <div class="actions">
        <el-button @click="loadWorkspaces">刷新</el-button>
        <el-button type="primary" @click="openCreate">新建业务空间</el-button>
      </div>
    </div>

    <el-alert
      v-if="permissionError"
      class="state-alert"
      type="warning"
      show-icon
      :closable="false"
      title="权限不足"
      description="当前账号无法访问业务空间，请联系管理员开通权限。"
    />
    <el-alert
      v-else-if="errorText"
      class="state-alert"
      type="error"
      show-icon
      :closable="false"
      title="加载失败"
      :description="errorText"
    />

    <div class="summary-row">
      <div class="summary-item">
        <strong>{{ workspaces.length }}</strong>
        <span>业务空间</span>
      </div>
      <div class="summary-item">
        <strong>{{ totalHosts }}</strong>
        <span>主机引用</span>
      </div>
      <div class="summary-item">
        <strong>{{ totalEndpoints }}</strong>
        <span>端点引用</span>
      </div>
    </div>

    <el-empty
      v-if="!loading && !workspaces.length && !errorText && !permissionError"
      class="empty-box"
      description="暂无业务空间"
    >
      <el-button type="primary" @click="openCreate">新建业务空间</el-button>
    </el-empty>

    <el-table v-else :data="workspaces" class="workspace-table" stripe>
      <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
      <el-table-column prop="owner" label="负责人" width="130" show-overflow-tooltip>
        <template #default="{ row }">{{ row.owner || '-' }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="110">
        <template #default="{ row }">
          <el-tag size="small" :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="资源" width="150">
        <template #default="{ row }">
          {{ row.resource_count ?? resourceCount(row) }} 项
        </template>
      </el-table-column>
      <el-table-column label="标签" min-width="180">
        <template #default="{ row }">
          <div class="tag-list">
            <el-tag v-for="tag in normalizeArray(row.tags)" :key="tag" size="small" effect="plain">{{ tag }}</el-tag>
            <span v-if="!normalizeArray(row.tags).length" class="muted">-</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="180">
        <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button text size="small" @click="openEdit(row)">编辑</el-button>
          <el-button text size="small" type="danger" @click="confirmDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑业务空间' : '新建业务空间'" width="620px" @close="resetForm">
      <el-form label-position="top">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" maxlength="80" show-word-limit placeholder="例如：支付核心链路" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" maxlength="300" show-word-limit />
        </el-form-item>
        <div class="form-grid">
          <el-form-item label="负责人">
            <el-input v-model="form.owner" maxlength="60" placeholder="负责人或团队" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="form.status" style="width: 100%">
              <el-option label="启用" value="active" />
              <el-option label="停用" value="disabled" />
              <el-option label="归档" value="archived" />
            </el-select>
          </el-form-item>
        </div>
        <el-form-item label="标签">
          <el-input v-model="form.tagsText" placeholder="用英文逗号或换行分隔" />
        </el-form-item>
        <el-form-item label="主机">
          <el-input v-model="form.hostsText" type="textarea" :rows="3" placeholder="每行一个主机名或 IP" />
        </el-form-item>
        <el-form-item label="端点">
          <el-input v-model="form.endpointsText" type="textarea" :rows="4" placeholder="每行一个端点，例如 10.0.0.1:8080 Web HTTP" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveWorkspace">保存</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { isPermissionError, normalizeList, redactText, workspaceApi } from '../../api/workspaces'

const emit = defineEmits(['count-change'])
const loading = ref(false)
const saving = ref(false)
const workspaces = ref([])
const errorText = ref('')
const permissionError = ref(false)
const dialogVisible = ref(false)
const editingId = ref('')
const form = reactive({ name: '', description: '', owner: '', status: 'active', tagsText: '', hostsText: '', endpointsText: '' })
const allowedStatuses = ['active', 'disabled', 'archived']

const normalizeArray = value => Array.isArray(value) ? value : []
const normalizeStatus = status => allowedStatuses.includes(status) ? status : 'disabled'
const totalHosts = computed(() => workspaces.value.reduce((sum, item) => sum + normalizeArray(item.hosts).length, 0))
const totalEndpoints = computed(() => workspaces.value.reduce((sum, item) => sum + normalizeArray(item.endpoints).length, 0))

const loadWorkspaces = async () => {
  loading.value = true
  errorText.value = ''
  permissionError.value = false
  try {
    workspaces.value = normalizeList(await workspaceApi.list())
    emit('count-change', workspaces.value.length)
  } catch (error) {
    workspaces.value = []
    permissionError.value = isPermissionError(error)
    errorText.value = redactText(error.message || '加载业务空间失败')
    emit('count-change', 0)
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  resetForm()
  dialogVisible.value = true
}

const openEdit = row => {
  editingId.value = row.id
  form.name = row.name || ''
  form.description = row.description || ''
  form.owner = row.owner || ''
  form.status = normalizeStatus(row.status || 'active')
  form.tagsText = normalizeArray(row.tags).join(', ')
  form.hostsText = normalizeArray(row.hosts).join('\n')
  form.endpointsText = normalizeArray(row.endpoints).map(endpointText).join('\n')
  dialogVisible.value = true
}

const saveWorkspace = async () => {
  if (!form.name.trim()) {
    ElMessage.warning('请输入业务空间名称')
    return
  }
  saving.value = true
  try {
    const payload = buildPayload()
    if (editingId.value) await workspaceApi.update(editingId.value, payload)
    else await workspaceApi.create(payload)
    ElMessage.success('业务空间已保存')
    dialogVisible.value = false
    await loadWorkspaces()
  } catch (error) {
    ElMessage.error(redactText(error.message || '保存业务空间失败'))
  } finally {
    saving.value = false
  }
}

const confirmDelete = async row => {
  try {
    await ElMessageBox.confirm(`确认删除业务空间「${row.name || row.id}」？`, '删除业务空间', {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
    })
    await workspaceApi.remove(row.id)
    ElMessage.success('业务空间已删除')
    await loadWorkspaces()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error(redactText(error.message || '删除业务空间失败'))
  }
}

const buildPayload = () => ({
  name: form.name.trim(),
  description: form.description.trim(),
  owner: form.owner.trim(),
  status: normalizeStatus(form.status || 'active'),
  tags: splitText(form.tagsText),
  hosts: splitText(form.hostsText),
  endpoints: parseEndpoints(form.endpointsText),
})

const splitText = text => String(text || '').split(/[\n,]/).map(item => item.trim()).filter(Boolean)
const parseEndpoints = text => String(text || '').split('\n').map(parseEndpoint).filter(Boolean)
const parseEndpoint = line => {
  const parts = String(line || '').trim().split(/\s+/).filter(Boolean)
  if (!parts.length) return null
  const [hostPort, serviceName = '', protocol = ''] = parts
  const match = hostPort.match(/^(.+):(\d{1,5})$/)
  if (!match) return { ip: hostPort, port: 0, service_name: serviceName, protocol }
  return { ip: match[1], port: Number(match[2]), service_name: serviceName, protocol }
}

const endpointText = endpoint => {
  if (!endpoint) return ''
  const port = endpoint.port ? `:${endpoint.port}` : ''
  return [endpoint.ip ? `${endpoint.ip}${port}` : '', endpoint.service_name || endpoint.service || '', endpoint.protocol || ''].filter(Boolean).join(' ')
}

const resetForm = () => {
  editingId.value = ''
  Object.assign(form, { name: '', description: '', owner: '', status: 'active', tagsText: '', hostsText: '', endpointsText: '' })
}

const resourceCount = row => normalizeArray(row.hosts).length + normalizeArray(row.endpoints).length
const statusLabel = status => ({ active: '启用', disabled: '停用', archived: '归档' }[status] || status || '未知')
const statusType = status => ({ active: 'success', disabled: 'info', archived: 'warning' }[status] || 'info')
const formatTime = value => value ? new Date(value).toLocaleString('zh-CN') : '-'

onMounted(loadWorkspaces)
</script>

<style scoped>
.business-panel { display: flex; flex-direction: column; gap: 16px; }
.toolbar { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.toolbar h3 { margin: 0; color: #1e3a5f; font-size: 18px; }
.toolbar p { margin: 6px 0 0; color: #60728e; font-size: 13px; }
.actions { display: flex; gap: 10px; flex-wrap: wrap; justify-content: flex-end; }
.state-alert { border-radius: 8px; }
.summary-row { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; }
.summary-item { border: 1px solid #e4e9f2; border-radius: 8px; padding: 14px 16px; background: #f8fbff; }
.summary-item strong { display: block; color: #1769ff; font-size: 24px; }
.summary-item span { color: #60728e; font-size: 12px; }
.workspace-table { border: 1px solid #e4e9f2; border-radius: 8px; overflow: hidden; }
.tag-list { display: flex; flex-wrap: wrap; gap: 6px; }
.muted { color: #8a99ad; }
.empty-box { min-height: 300px; border: 1px dashed #d8e1ee; border-radius: 8px; background: #f8fbff; }
.form-grid { display: grid; grid-template-columns: 1fr 160px; gap: 12px; }
@media (max-width: 760px) {
  .toolbar { flex-direction: column; }
  .summary-row, .form-grid { grid-template-columns: 1fr; }
}
</style>
