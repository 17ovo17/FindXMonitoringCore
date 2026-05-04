<template>
  <div class="alerting-page">
    <section class="alerting-panel">
      <div class="section-head">
        <div>
          <div class="kicker">告警</div>
          <h2>{{ current.title }}</h2>
          <p>{{ current.desc }}</p>
        </div>
      </div>
      <MonitorAlertRulesPanel v-if="section === 'rules'" />
      <MonitorEventsPanel v-else-if="section === 'events' || section === 'aggr-views'" />
      <div v-else class="empty-state">
        <el-empty :description="current.empty">
          <el-alert :title="current.hint" type="info" show-icon :closable="false" />
        </el-empty>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import MonitorAlertRulesPanel from '../components/monitoring/MonitorAlertRulesPanel.vue'
import MonitorEventsPanel from '../components/monitoring/MonitorEventsPanel.vue'

const route = useRoute()
const sectionAliases = { 'silence-subscribe': 'mutes' }
const validSections = new Set(['rules', 'recording-rules', 'events', 'aggr-views', 'mutes', 'subscribes', 'pipeline'])
const section = computed(() => {
  const current = sectionAliases[route.query.section] || route.query.section
  return validSections.has(current) ? current : 'events'
})
const copy = {
  rules: { title: '告警规则', desc: '复用已接入的规则列表、详情和 TryRun 校验。' },
  events: { title: '事件中心', desc: '查看真实告警事件和状态流转。' },
  'recording-rules': {
    title: '记录规则',
    desc: '沉淀预计算规则和指标转换。',
    empty: '记录规则能力尚未接入。',
    hint: '当前未发现记录规则接口，保持真实空态。',
  },
  'aggr-views': { title: '聚合视图', desc: '按事件指纹、主机和业务进行聚合查看。' },
  mutes: {
    title: '告警静默',
    desc: '按业务空间、标签和时间窗口管理告警静默。',
    empty: '告警静默能力尚未接入真实接口。',
    hint: '静默是告警域能力，不再和通知或团队机制混在一起。',
  },
  subscribes: {
    title: '告警订阅',
    desc: '按用户、团队组织和业务空间订阅告警事件。',
    empty: '告警订阅能力尚未接入真实接口。',
    hint: '订阅接收对象复用团队组织机制，不单独创建接收组。',
  },
  pipeline: {
    title: '事件流水线',
    desc: '对告警事件执行过滤、更新、丢弃、回调和 AI 摘要等处理。',
    empty: '事件流水线尚未接入真实接口。',
    hint: '事件流水线属于告警事件治理，不放在通知渠道里。',
  },
}
const current = computed(() => copy[section.value] || copy.rules)
</script>

<style scoped>
.alerting-page { min-height: 100%; padding: 24px; color: #243553; }
.embedded-panel, .alerting-panel { min-height: calc(100vh - 114px); border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto; }
.alerting-panel { padding: 22px; }
.section-head { margin-bottom: 16px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.empty-state { min-height: 420px; display: grid; place-items: center; border: 1px dashed #d8e1ee; border-radius: 8px; background: #f8fbff; }
:deep(.alerts-page) { min-height: auto; padding: 24px; }
</style>
