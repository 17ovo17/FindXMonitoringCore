<template>
  <div class="health-grid">
    <div v-for="card in cards" :key="card.label" class="health-card">
      <div class="label">{{ card.label }}</div>
      <strong>{{ card.value }}</strong>
      <el-tag size="small" :type="card.type">{{ card.note }}</el-tag>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { redactText } from '../../api/monitor'

const props = defineProps({ health: Object, loading: Boolean, error: String })
const statusType = status => ({ healthy: 'success', degraded: 'warning', empty: 'info' }[status] || 'danger')
const okText = value => value === true ? '正常' : value === false ? '异常' : '未知'
const safe = value => redactText(value)

const cards = computed(() => {
  const h = props.health || {}
  const note = props.error ? safe(props.error) : safe(h.mode || '-')
  return [
    { label: '核心状态', value: props.loading ? '加载中' : safe(h.status || '未知'), note, type: props.error ? 'danger' : statusType(h.status) },
    { label: '监控对象', value: h.targets ?? '-', note: 'Targets', type: 'primary' },
    { label: 'FindX Agent', value: `${h.agent_online ?? 0}/${h.agents ?? 0}`, note: '在线/总数', type: (h.agent_online || 0) > 0 ? 'success' : 'info' },
    { label: '存储', value: okText(h.storage?.mysql), note: `Redis ${okText(h.storage?.redis)}`, type: h.storage?.mysql ? 'success' : 'warning' },
  ]
})
</script>

<style scoped>
.health-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; }
.health-card { min-height: 92px; padding: 14px 16px; border: 1px solid #e4e9f2; border-radius: 8px; background: #fff; box-shadow: 0 8px 24px rgba(31,45,61,.06); }
.label { color: #6b778c; font-size: 12px; font-weight: 700; }
strong { display: block; margin: 8px 0 10px; color: #233553; font-size: 26px; line-height: 1; }
@media (max-width: 900px) { .health-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
</style>
