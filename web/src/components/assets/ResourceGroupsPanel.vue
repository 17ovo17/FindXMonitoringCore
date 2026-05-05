<template>
  <section class="asset-panel" v-loading="loading">
    <div class="toolbar">
      <div>
        <h3>资源组</h3>
        <p>按资源边界归集主机资产，资源组数据来自后端真实接口。</p>
      </div>
      <div class="actions">
        <el-button @click="loadGroups">刷新</el-button>
        <el-button type="primary" @click="openCreate">新建资源组</el-button>
      </div>
    </div>
    <el-alert v-if="permissionError" class="state-alert" type="warning" show-icon :closable="false" title="权限不足" description="当前账号无法访问资源组，请联系管理员开通权限。" />
    <el-alert v-else-if="errorText" class="state-alert" type="error" show-icon :closable="false" title="加载失败" :description="errorText" />
    <el-empty v-if="!loading && !groups.length && !errorText && !permissionError" class="empty-box" description="暂无资源组">
      <el-button type="primary" @click="openCreate">新建资源组</el-button>
    </el-empty>
    <el-table v-else :data="groups" class="asset-table" stripe empty-text="暂无资源组">
      <el-table-column label="名称" min-width="180" show-overflow-tooltip>
        <template #default="{ row }">{{ safe(row.name) || '-' }}</template>
      </el-table-column>
      <el-table-column label="业务空间" min-width="150" show-overflow-tooltip>
        <template #default="{ row }">{{ safe(row.workspace_name || row.workspace_id) || '-' }}</template>
      </el-table-column>
      <el-table-column label="标签" min-width="170">
        <template #default="{ row }">
          <div class="tag-list"><el-tag v-for="tag in tags(row)" :key="tag" size="small" effect="plain">{{ safe(tag) }}</el-tag><span v-if="!tags(row).length" class="muted">-</span></div>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="170"><template #default="{ row }">{{ fmt(row.updated_at || row.update_time) }}</template></el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button text size="small" @click="openEdit(row)">编辑</el-button>
          <el-button text size="small" type="danger" :loading="deletingId === row.id" @click="confirmDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑资源组' : '新建资源组'" width="560px" @close="resetForm">
      <el-form label-position="top">
        <el-form-item label="名称" required><el-input v-model="form.name" maxlength="80" show-word-limit /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" maxlength="300" show-word-limit /></el-form-item>
        <el-form-item label="业务空间 ID"><el-input v-model="form.workspace_id" maxlength="80" clearable /></el-form-item>
        <el-form-item label="标签"><el-input v-model="form.tagsText" placeholder="使用英文逗号或换行分隔" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" :disabled="saving" @click="saveGroup">保存</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { assetsApi, formatAssetError, isPermissionError, normalizeArray, normalizeList, redactText, splitTags } from '../../api/assets'

const emit = defineEmits(['count-change'])
const loading = ref(false)
const saving = ref(false)
const deletingId = ref('')
const groups = ref([])
const errorText = ref('')
const permissionError = ref(false)
const dialogVisible = ref(false)
const editingId = ref('')
const form = reactive({ name: '', description: '', workspace_id: '', tagsText: '' })

const safe = value => redactText(value)
const tags = row => normalizeArray(row?.tags)
const fmt = value => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'

const loadGroups = async () => {
  loading.value = true
  errorText.value = ''
  permissionError.value = false
  try {
    groups.value = normalizeList(await assetsApi.listResourceGroups())
    emit('count-change', groups.value.length)
  } catch (error) {
    groups.value = []
    permissionError.value = isPermissionError(error)
    errorText.value = formatAssetError(error, '资源组')
    emit('count-change', 0)
  } finally {
    loading.value = false
  }
}

const openCreate = () => { resetForm(); dialogVisible.value = true }
const openEdit = row => {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name || '',
    description: row.description || '',
    workspace_id: row.workspace_id || '',
    tagsText: tags(row).join(', '),
  })
  dialogVisible.value = true
}

const saveGroup = async () => {
  if (!form.name.trim()) return ElMessage.warning('请输入资源组名称')
  saving.value = true
  try {
    const payload = { name: form.name.trim(), description: form.description.trim(), workspace_id: form.workspace_id.trim(), tags: splitTags(form.tagsText) }
    if (editingId.value) await assetsApi.updateResourceGroup(editingId.value, payload)
    else await assetsApi.createResourceGroup(payload)
    ElMessage.success('资源组已保存')
    dialogVisible.value = false
    await loadGroups()
  } catch (error) {
    ElMessage.error(formatAssetError(error, '资源组'))
  } finally {
    saving.value = false
  }
}

const confirmDelete = async row => {
  try {
    await ElMessageBox.confirm(`确认删除资源组「${safe(row.name || row.id)}」？`, '删除资源组', { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' })
    deletingId.value = row.id
    await assetsApi.deleteResourceGroup(row.id)
    ElMessage.success('资源组已删除')
    await loadGroups()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error(formatAssetError(error, '资源组删除'))
  } finally {
    deletingId.value = ''
  }
}

const resetForm = () => {
  editingId.value = ''
  Object.assign(form, { name: '', description: '', workspace_id: '', tagsText: '' })
}

onMounted(loadGroups)
</script>

<style scoped>
.asset-panel { display: flex; flex-direction: column; gap: 16px; }
.toolbar { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.toolbar h3 { margin: 0; color: #1e3a5f; font-size: 18px; }
.toolbar p { margin: 6px 0 0; color: #60728e; font-size: 13px; }
.actions { display: flex; gap: 10px; flex-wrap: wrap; justify-content: flex-end; }
.state-alert, .asset-table, .empty-box { border-radius: 8px; }
.asset-table { border: 1px solid #e4e9f2; overflow: hidden; }
.empty-box { min-height: 300px; border: 1px dashed #d8e1ee; background: #f8fbff; }
.tag-list { display: flex; flex-wrap: wrap; gap: 6px; }
.muted { color: #8a99ad; }
@media (max-width: 760px) { .toolbar { flex-direction: column; } }
</style>
