<template>
  <el-dialog :model-value="formVisible" :title="editingId ? '编辑仪表盘' : '新增仪表盘'" width="720px" @update:model-value="$emit('update:formVisible', $event)">
    <el-form label-position="top" @submit.prevent>
      <div class="form-grid">
        <el-form-item label="名称" required><el-input v-model.trim="form.name" maxlength="80" /></el-form-item>
        <el-form-item label="唯一标识"><el-input v-model.trim="form.ident" maxlength="80" /></el-form-item>
      </div>
      <el-form-item label="标签"><el-input :model-value="tagText" placeholder="多个标签用逗号分隔" @update:model-value="$emit('update:tagText', $event)" /></el-form-item>
      <el-form-item label="备注"><el-input v-model.trim="form.note" type="textarea" :rows="3" maxlength="300" /></el-form-item>
      <div class="form-grid">
        <el-form-item label="业务组"><el-input v-model.trim="form.business_group" maxlength="80" /></el-form-item>
        <el-form-item label="共享状态">
          <el-select v-model="form.shared">
            <el-option label="私有" :value="false" />
            <el-option label="公开" :value="true" />
          </el-select>
        </el-form-item>
        <el-form-item label="图表 Tooltip"><el-switch v-model="form.graphTooltip" /></el-form-item>
        <el-form-item label="图表缩放"><el-switch v-model="form.graphZoom" /></el-form-item>
      </div>
      <el-alert v-if="formError" :title="formError" type="error" show-icon :closable="false" />
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:formVisible', false)">取消</el-button>
      <el-button type="primary" :loading="saving" @click="$emit('save-dashboard')">保存</el-button>
    </template>
  </el-dialog>

  <el-drawer :model-value="panelDrawerVisible" :title="panelDrawerTitle" size="520px" @update:model-value="$emit('update:panelDrawerVisible', $event)">
    <el-alert title="BLOCKED_BY_CONTRACT：后端未提供完整 Panel 保存 contract，本入口只展示真实配置和阻断原因。" type="warning" show-icon :closable="false" />
    <el-input :model-value="panelDraftJson" type="textarea" :rows="18" readonly spellcheck="false" class="json-view" />
  </el-drawer>

  <el-dialog :model-value="previewVisible" :title="previewTitle" width="760px" @update:model-value="$emit('update:previewVisible', $event)">
    <el-descriptions :column="2" border>
      <el-descriptions-item label="分类">{{ selectedTemplate?.kind || '-' }}</el-descriptions-item>
      <el-descriptions-item label="Panel 数">{{ selectedTemplate?.panelCount ?? 0 }}</el-descriptions-item>
      <el-descriptions-item label="变量数">{{ selectedTemplate?.variableCount ?? 0 }}</el-descriptions-item>
      <el-descriptions-item label="说明">{{ selectedTemplate?.description || '无' }}</el-descriptions-item>
    </el-descriptions>
    <el-input :model-value="previewJson" type="textarea" :rows="14" readonly spellcheck="false" class="json-view" />
    <template #footer>
      <el-button @click="$emit('update:previewVisible', false)">关闭</el-button>
      <el-button type="primary" @click="$emit('open-import', selectedTemplate)">导入</el-button>
    </template>
  </el-dialog>

  <el-dialog :model-value="importVisible" title="导入模板" width="720px" @update:model-value="$emit('update:importVisible', $event)">
    <el-form label-position="top" @submit.prevent>
      <el-form-item label="名称" required><el-input v-model.trim="importForm.name" maxlength="80" /></el-form-item>
      <div class="form-grid">
        <el-form-item label="业务组"><el-input v-model.trim="importForm.business_group" maxlength="80" /></el-form-item>
        <el-form-item label="标签"><el-input :model-value="importTagText" placeholder="多个标签用逗号分隔" @update:model-value="$emit('update:importTagText', $event)" /></el-form-item>
      </div>
      <el-form-item label="变量 JSON"><el-input :model-value="variablesJson" type="textarea" :rows="8" spellcheck="false" placeholder="{}" @update:model-value="$emit('update:variablesJson', $event)" /></el-form-item>
      <el-alert v-if="importError" :title="importError" type="error" show-icon :closable="false" />
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:importVisible', false)">取消</el-button>
      <el-button type="primary" :loading="importing" @click="$emit('submit-import')">导入为仪表盘</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
defineProps({
  formVisible: Boolean,
  panelDrawerVisible: Boolean,
  previewVisible: Boolean,
  importVisible: Boolean,
  saving: Boolean,
  importing: Boolean,
  editingId: String,
  tagText: String,
  formError: String,
  panelDrawerTitle: String,
  panelDraftJson: String,
  previewTitle: String,
  previewJson: String,
  variablesJson: String,
  importTagText: String,
  importError: String,
  selectedTemplate: Object,
  form: { type: Object, required: true },
  importForm: { type: Object, required: true },
})

defineEmits(['update:formVisible', 'update:panelDrawerVisible', 'update:previewVisible', 'update:importVisible', 'update:tagText', 'update:variablesJson', 'update:importTagText', 'save-dashboard', 'open-import', 'submit-import'])
</script>

<style scoped>
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 14px; }
.json-view { margin-top: 14px; }
@media (max-width: 900px) {
  .form-grid { grid-template-columns: 1fr; }
}
</style>
