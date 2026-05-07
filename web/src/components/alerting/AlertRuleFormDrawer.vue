<template>
  <el-drawer :model-value="visible" :title="title" size="720px" @close="$emit('update:visible', false)">
    <el-form label-position="top" @submit.prevent>
      <div class="form-grid">
        <el-form-item label="名称" required><el-input v-model.trim="form.name" maxlength="120" /></el-form-item>
        <el-form-item label="数据源" required>
          <el-select v-model="form.datasource_id" filterable placeholder="请选择数据源">
            <el-option v-for="item in datasources" :key="datasourceValue(item)" :label="datasourceLabel(item)" :value="datasourceValue(item)" />
          </el-select>
        </el-form-item>
        <el-form-item label="级别" required>
          <el-select v-model="form.severity">
            <el-option v-for="item in severityOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="持续时间">
          <el-input v-model.trim="form.for_duration" placeholder="例如 3m、5m、10m" />
        </el-form-item>
        <el-form-item label="无数据策略">
          <el-select v-model="form.no_data_policy">
            <el-option label="保持状态" value="keep_state" />
            <el-option label="触发告警" value="alerting" />
            <el-option label="恢复正常" value="ok" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </div>
      <el-form-item label="查询表达式" required>
        <el-input v-model="form.query" type="textarea" :rows="7" spellcheck="false" placeholder="输入监控查询表达式" />
      </el-form-item>
      <el-form-item label="附加标签 JSON">
        <el-input v-model="labelsJson" type="textarea" :rows="5" spellcheck="false" placeholder="{}" />
      </el-form-item>
      <el-form-item label="注解 JSON">
        <el-input v-model="annotationsJson" type="textarea" :rows="5" spellcheck="false" placeholder="{}" />
      </el-form-item>
      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
      <el-alert title="规则保存后会生成版本记录；TryRun 只执行校验和查询评估，不生成正式事件。" type="info" show-icon :closable="false" class="hint" />
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button :disabled="!rule?.id" :loading="tryLoading" @click="tryRun">TryRun</el-button>
      <el-button type="primary" :loading="saving" @click="save">保存</el-button>
    </template>

    <el-dialog v-model="tryVisible" title="TryRun 结果" width="760px" append-to-body>
      <el-input :model-value="tryResult" type="textarea" :rows="18" readonly spellcheck="false" />
    </el-dialog>
  </el-drawer>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { alertingApi, redactText, safeJson } from '../../api/alerting'
import { datasourceLabel, datasourceValue, parseJsonObject, severityOptions } from './alertingModel'

const props = defineProps({
  visible: Boolean,
  rule: { type: Object, default: null },
  datasources: { type: Array, default: () => [] },
  saving: Boolean,
})

const emit = defineEmits(['update:visible', 'save'])

const tryLoading = ref(false)
const tryVisible = ref(false)
const tryResult = ref('')
const error = ref('')
const labelsJson = ref('{}')
const annotationsJson = ref('{}')
const form = reactive(defaultForm())
const title = computed(() => props.rule?.id ? `编辑告警规则：${props.rule.name}` : '新增告警规则')

function defaultForm() {
  return {
    id: '',
    name: '',
    datasource_id: '',
    severity: 'warning',
    query: '',
    for_duration: '3m',
    no_data_policy: 'keep_state',
    enabled: true,
    labels: {},
    annotations: {},
  }
}

const fillForm = () => {
  Object.assign(form, defaultForm())
  if (props.rule) {
    Object.assign(form, {
      id: props.rule.id || '',
      name: props.rule.name || '',
      datasource_id: props.rule.datasourceId === '-' ? '' : props.rule.datasourceId,
      severity: props.rule.severity || 'warning',
      query: props.rule.query || '',
      for_duration: props.rule.forDuration || '3m',
      no_data_policy: props.rule.noDataPolicy || 'keep_state',
      enabled: props.rule.enabled !== false,
      labels: props.rule.labels || {},
      annotations: props.rule.annotations || {},
    })
  }
  labelsJson.value = safeJson(form.labels, 4000)
  annotationsJson.value = safeJson(form.annotations, 4000)
  error.value = ''
}

const buildPayload = () => {
  if (!form.name.trim()) throw new Error('名称不能为空')
  if (!form.datasource_id) throw new Error('数据源不能为空')
  if (!form.query.trim()) throw new Error('查询表达式不能为空')
  return {
    id: form.id,
    name: form.name.trim(),
    datasource_id: form.datasource_id,
    severity: form.severity,
    query: form.query,
    for_duration: form.for_duration || '3m',
    no_data_policy: form.no_data_policy || 'keep_state',
    enabled: form.enabled,
    status: form.enabled ? 'active' : 'disabled',
    labels: parseJsonObject(labelsJson.value, '附加标签 JSON'),
    annotations: parseJsonObject(annotationsJson.value, '注解 JSON'),
  }
}

const save = () => {
  error.value = ''
  try {
    emit('save', buildPayload())
  } catch (err) {
    error.value = redactText(err.message || '表单校验失败')
  }
}

const tryRun = async () => {
  if (!props.rule?.id) return
  error.value = ''
  tryLoading.value = true
  tryResult.value = '执行中...'
  tryVisible.value = true
  try {
    const result = await alertingApi.tryRunRule(props.rule.id, buildPayload())
    tryResult.value = safeJson(result)
  } catch (err) {
    tryResult.value = redactText(err.message || 'TryRun 失败')
  }
  tryLoading.value = false
  ElMessage.info('TryRun 已返回结果')
}

watch(() => props.visible, value => { if (value) fillForm() })
watch(() => props.rule, fillForm)
</script>

<style scoped>
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 14px; }
.hint { margin-top: 12px; }
@media (max-width: 760px) { .form-grid { grid-template-columns: 1fr; } }
</style>
