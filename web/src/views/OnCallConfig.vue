<template>
  <div class="oncall-page">
    <OnCallPageHeader :loading="loading" @reload="reloadAll" @create-group="openGroupDialog()" @create-channel="openChannelDialog()" />

    <OnCallSummaryStrip
      :groups-count="groups.length"
      :enabled-channels="enabledChannels"
      :enabled-escalation="enabledEscalation"
      :records-count="records.length"
    />

    <div class="layout-grid">
      <OnCallGroupsPanel
        :groups="groups"
        :format-members="formatMembers"
        @create="openGroupDialog()"
        @edit="openGroupDialog"
        @remove="removeGroup"
      />

      <OnCallChannelsPanel
        :channels="channels"
        :channel-type-label="channelTypeLabel"
        :display-endpoint="displayEndpoint"
        :is-builtin-channel="isBuiltinChannel"
        @create="openChannelDialog()"
        @edit="openChannelDialog"
        @test="testChannel"
        @remove="removeChannel"
      />

      <OnCallEscalationPanel
        :escalation="escalation"
        @create="openStepDialog()"
        @edit="openStepDialog"
        @remove="removeStep"
      />

      <OnCallRecordsPanel
        :records="records"
        :records-loading="recordsLoading"
        :format-time="formatTime"
        @reload="loadRecords"
        @retry="retryRecord"
      />
    </div>

    <OnCallDialogs
      v-model:group-dialog="groupDialog"
      v-model:channel-dialog="channelDialog"
      v-model:step-dialog="stepDialog"
      :group-draft="groupDraft"
      :channel-draft="channelDraft"
      :step-draft="stepDraft"
      :saving="saving"
      @save-group="saveGroup"
      @save-channel="saveChannel"
      @save-step="saveStep"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'
import OnCallChannelsPanel from '../components/oncall/OnCallChannelsPanel.vue'
import OnCallDialogs from '../components/oncall/OnCallDialogs.vue'
import OnCallEscalationPanel from '../components/oncall/OnCallEscalationPanel.vue'
import OnCallGroupsPanel from '../components/oncall/OnCallGroupsPanel.vue'
import OnCallPageHeader from '../components/oncall/OnCallPageHeader.vue'
import OnCallRecordsPanel from '../components/oncall/OnCallRecordsPanel.vue'
import OnCallSummaryStrip from '../components/oncall/OnCallSummaryStrip.vue'
import { TEST_BATCH_ID, channelTypeLabel, displayEndpoint, formatMembers, formatTime, isBuiltinChannel, normalizeChannels } from '../components/oncall/oncallConfigHelpers'
import '../components/oncall/oncallConfig.css'

const loading = ref(false)
const saving = ref(false)
const recordsLoading = ref(false)
const groups = ref([])
const channels = ref([])
const escalation = ref([])
const records = ref([])
const groupDialog = ref(false)
const channelDialog = ref(false)
const stepDialog = ref(false)

const groupDraft = reactive({ id: '', name: '', membersText: '', schedule: '', role: 'primary', enabled: true })
const channelDraft = reactive({ id: '', name: '', type: 'webhook', endpoint: '', webhook: '', receiver: '', secret: '', retry_policy: '3 次，每次间隔 60s', enabled: true })
const stepDraft = reactive({ id: '', delay_min: 0, target: '', action: '', condition: '', enabled: true })

const enabledChannels = computed(() => channels.value.filter(item => item.enabled).length)
const enabledEscalation = computed(() => escalation.value.filter(item => item.enabled).length)

const primaryReceiver = () => groups.value.find(item => item.enabled)?.name || '值班人员'

const loadConfig = async () => {
  const { data } = await axios.get('/api/v1/oncall/config')
  escalation.value = (data.escalation || []).sort((left, right) => Number(left.delay_min || 0) - Number(right.delay_min || 0))
}

const loadGroups = async () => {
  const { data } = await axios.get('/api/v1/oncall/groups')
  groups.value = data.items || data || []
}

const loadChannels = async () => {
  const { data } = await axios.get('/api/v1/oncall/channels')
  channels.value = normalizeChannels(data.items || data || [])
}

const loadRecords = async () => {
  recordsLoading.value = true
  try {
    const { data } = await axios.get('/api/v1/oncall/records')
    records.value = data.items || data || []
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '加载通知发送记录失败')
  } finally {
    recordsLoading.value = false
  }
}

const reloadAll = async () => {
  loading.value = true
  try {
    await Promise.all([loadConfig(), loadGroups(), loadChannels(), loadRecords()])
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '加载团队排班通知策略失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  await axios.post('/api/v1/oncall/config', { groups: groups.value, escalation: escalation.value })
}

const openGroupDialog = group => {
  Object.assign(groupDraft, {
    id: group?.id || '',
    name: group?.name || '',
    membersText: (group?.members || []).join(','),
    schedule: group?.schedule || '',
    role: group?.role || 'primary',
    enabled: group?.enabled ?? true,
  })
  groupDialog.value = true
}

const saveGroup = async () => {
  const name = groupDraft.name.trim()
  if (!name) return ElMessage.warning('请填写团队名称')
  const item = {
    id: groupDraft.id || `group_${Date.now()}`,
    name,
    members: groupDraft.membersText.split(/[,，\n]/).map(member => member.trim()).filter(Boolean),
    schedule: groupDraft.schedule.trim(),
    role: groupDraft.role.trim() || 'primary',
    enabled: groupDraft.enabled,
  }
  saving.value = true
  try {
    const { data } = await axios.post('/api/v1/oncall/groups', item)
    groups.value = groups.value.some(group => group.id === data.id)
      ? groups.value.map(group => group.id === data.id ? data : group)
      : [...groups.value, data]
    groupDialog.value = false
    ElMessage.success('团队排班已保存')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存团队排班失败')
  } finally {
    saving.value = false
  }
}

const removeGroup = async group => {
  await ElMessageBox.confirm(`确认删除团队排班「${group.name}」？`, '删除团队排班', { type: 'warning' })
  await axios.delete(`/api/v1/oncall/groups/${group.id}`)
  groups.value = groups.value.filter(item => item.id !== group.id)
  ElMessage.success('团队排班已删除')
}

const openChannelDialog = channel => {
  Object.assign(channelDraft, {
    id: channel?.id || '',
    name: channel?.name || '',
    type: channel?.type || 'webhook',
    endpoint: channel?.endpoint || channel?.webhook || '',
    webhook: channel?.webhook || channel?.endpoint || '',
    receiver: channel?.receiver || '',
    secret: channel?.secret || '',
    retry_policy: channel?.retry_policy || '3 次，每次间隔 60s',
    enabled: channel?.enabled ?? true,
  })
  channelDialog.value = true
}

const saveChannel = async () => {
  const name = channelDraft.name.trim()
  if (!name) return ElMessage.warning('请填写渠道名称')
  if (channelDraft.enabled && channelDraft.type !== 'console' && !channelDraft.endpoint.trim()) {
    return ElMessage.warning('启用外部通知渠道前必须配置 Endpoint')
  }
  const item = { ...channelDraft, id: channelDraft.id || `${channelDraft.type}_${Date.now()}` }
  item.webhook = item.endpoint
  saving.value = true
  try {
    const { data } = await axios.post('/api/v1/oncall/channels', item)
    channels.value = normalizeChannels(channels.value.map(channel => channel.id === data.id ? data : channel).concat(channels.value.some(channel => channel.id === data.id) ? [] : [data]))
    channelDialog.value = false
    ElMessage.success('通知渠道已保存')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存通知渠道失败')
  } finally {
    saving.value = false
  }
}

const removeChannel = async channel => {
  await ElMessageBox.confirm(`确认删除通知渠道「${channel.name}」？`, '删除通知渠道', { type: 'warning' })
  await axios.delete(`/api/v1/oncall/channels/${channel.id}`)
  channels.value = channels.value.filter(item => item.id !== channel.id)
  ElMessage.success('通知渠道已删除')
}

const testChannel = async channel => {
  try {
    const { data } = await axios.post('/api/v1/oncall/test-send', {
      channel: channel.type,
      channel_id: channel.id,
      receiver: channel.receiver || primaryReceiver(),
      business_id: 'manual-check',
      alert_title: '值班通知链路验证',
      test_batch_id: TEST_BATCH_ID,
    })
    records.value = [data, ...records.value.filter(item => item.id !== data.id)]
    ElMessage.success('测试通知已发送并写入记录')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '测试通知发送失败')
  }
}

const openStepDialog = step => {
  Object.assign(stepDraft, {
    id: step?.id || '',
    delay_min: step?.delay_min ?? 0,
    target: step?.target || '',
    action: step?.action || '',
    condition: step?.condition || '',
    enabled: step?.enabled ?? true,
  })
  stepDialog.value = true
}

const saveStep = async () => {
  if (!stepDraft.target.trim() || !stepDraft.action.trim()) return ElMessage.warning('请填写升级目标和动作')
  const item = { ...stepDraft, id: stepDraft.id || `esc_${Date.now()}`, delay_min: Number(stepDraft.delay_min || 0) }
  saving.value = true
  try {
    escalation.value = escalation.value.some(step => step.id === item.id)
      ? escalation.value.map(step => step.id === item.id ? item : step)
      : [...escalation.value, item]
    escalation.value.sort((left, right) => Number(left.delay_min || 0) - Number(right.delay_min || 0))
    await saveConfig()
    stepDialog.value = false
    ElMessage.success('升级策略已保存')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存升级策略失败')
  } finally {
    saving.value = false
  }
}

const removeStep = async step => {
  await ElMessageBox.confirm(`确认删除升级步骤「${step.action}」？`, '删除升级步骤', { type: 'warning' })
  escalation.value = escalation.value.filter(item => item.id !== step.id)
  await saveConfig()
  ElMessage.success('升级步骤已删除')
}

const retryRecord = async record => {
  await axios.post('/api/v1/oncall/test-send', {
    channel: record.channel || 'console',
    receiver: record.receiver || primaryReceiver(),
    retry_of: record.id,
    business_id: 'manual-retry',
    alert_title: '值班通知重试',
    test_batch_id: TEST_BATCH_ID,
  }).then(({ data }) => {
    records.value = [data, ...records.value]
    ElMessage.success('重试已写入发送记录')
  }).catch(error => ElMessage.error(error.response?.data?.error || '重试失败'))
}

onMounted(reloadAll)
</script>
