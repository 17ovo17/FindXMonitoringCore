<template>
  <el-dialog v-model="groupVisible" :title="groupDraft.id ? '编辑团队排班' : '新增团队排班'" width="520px">
    <el-form label-position="top">
      <el-form-item label="团队名称"><el-input v-model="groupDraft.name" placeholder="例如：SRE 主值" /></el-form-item>
      <el-form-item label="成员（逗号或换行分隔）"><el-input v-model="groupDraft.membersText" placeholder="admin,sre-primary,dba" /></el-form-item>
      <el-form-item label="排班时间"><el-input v-model="groupDraft.schedule" placeholder="00:00-24:00 / 工作日 09:00-18:00" /></el-form-item>
      <el-form-item label="角色"><el-input v-model="groupDraft.role" placeholder="primary / expert / owner" /></el-form-item>
      <el-form-item><el-switch v-model="groupDraft.enabled" active-text="启用" inactive-text="停用" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="groupVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="$emit('saveGroup')">保存</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="channelVisible" :title="channelDraft.id ? '配置通知渠道' : '新增通知渠道'" width="560px">
    <el-form label-position="top">
      <el-form-item label="渠道名称"><el-input v-model="channelDraft.name" placeholder="例如：生产 Webhook" /></el-form-item>
      <el-form-item label="渠道类型">
        <el-select v-model="channelDraft.type" style="width: 100%">
          <el-option v-for="item in channelTypes" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="接收人/团队"><el-input v-model="channelDraft.receiver" placeholder="SRE 主值 / oncall@example.com" /></el-form-item>
      <el-form-item label="Endpoint / Webhook"><el-input v-model="channelDraft.endpoint" placeholder="https://example.com/oncall/webhook" /></el-form-item>
      <el-form-item label="Token / Secret"><el-input v-model="channelDraft.secret" type="password" show-password placeholder="仅保存，不在列表明文展示" /></el-form-item>
      <el-form-item label="重试策略"><el-input v-model="channelDraft.retry_policy" placeholder="3 次，每次间隔 60s" /></el-form-item>
      <el-form-item><el-switch v-model="channelDraft.enabled" active-text="启用" inactive-text="停用" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="channelVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="$emit('saveChannel')">保存</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="stepVisible" :title="stepDraft.id ? '编辑升级步骤' : '新增升级步骤'" width="520px">
    <el-form label-position="top">
      <el-form-item label="延迟分钟"><el-input-number v-model="stepDraft.delay_min" :min="0" :max="1440" style="width: 100%" /></el-form-item>
      <el-form-item label="触发条件"><el-input v-model="stepDraft.condition" placeholder="未认领 / 未恢复 / P0/P1" /></el-form-item>
      <el-form-item label="动作"><el-input v-model="stepDraft.action" placeholder="重复提醒 / 升级通知" /></el-form-item>
      <el-form-item label="目标"><el-input v-model="stepDraft.target" placeholder="SRE 主值 / 专家组 / 负责人" /></el-form-item>
      <el-form-item><el-switch v-model="stepDraft.enabled" active-text="启用" inactive-text="停用" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="stepVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="$emit('saveStep')">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'
import { channelTypes } from './oncallConfigHelpers'

const props = defineProps({
  groupDialog: { type: Boolean, required: true },
  channelDialog: { type: Boolean, required: true },
  stepDialog: { type: Boolean, required: true },
  groupDraft: { type: Object, required: true },
  channelDraft: { type: Object, required: true },
  stepDraft: { type: Object, required: true },
  saving: { type: Boolean, required: true },
})

const emit = defineEmits([
  'update:groupDialog',
  'update:channelDialog',
  'update:stepDialog',
  'saveGroup',
  'saveChannel',
  'saveStep',
])

const groupVisible = computed({
  get: () => props.groupDialog,
  set: value => emit('update:groupDialog', value),
})
const channelVisible = computed({
  get: () => props.channelDialog,
  set: value => emit('update:channelDialog', value),
})
const stepVisible = computed({
  get: () => props.stepDialog,
  set: value => emit('update:stepDialog', value),
})
</script>
