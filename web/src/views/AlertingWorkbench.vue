<template>
  <div class="alerting-page">
    <section class="alerting-shell">
      <header class="page-head">
        <div>
          <div class="kicker">告警</div>
          <h2>{{ current.title }}</h2>
          <p>{{ current.desc }}</p>
        </div>
        <el-segmented v-model="section" :options="tabs" @change="syncRoute" />
      </header>

      <AlertRulesPanel v-if="section === 'rules'" @blocked="openBlocked" />
      <AlertEventsPanel v-else-if="section === 'events'" @blocked="openBlocked" />
      <section v-else class="blocked-panel">
        <el-alert :title="blockedContracts[section]" type="warning" show-icon :closable="false" />
        <el-input :model-value="blockedPayload(section)" type="textarea" :rows="14" readonly spellcheck="false" class="json-view" />
      </section>
    </section>

    <AlertBlockedDrawer v-model:visible="blockedVisible" :message="blockedMessage" :payload="blockedJson" />
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AlertBlockedDrawer from '../components/alerting/AlertBlockedDrawer.vue'
import AlertEventsPanel from '../components/alerting/AlertEventsPanel.vue'
import AlertRulesPanel from '../components/alerting/AlertRulesPanel.vue'
import { blockedContracts, blockedPayload } from '../components/alerting/alertingModel'

const route = useRoute()
const router = useRouter()
const validSections = new Set(['rules', 'events', 'mutes', 'subscriptions', 'workflows'])
const section = ref(validSections.has(route.query.section) ? route.query.section : 'events')
const blockedVisible = ref(false)
const blockedMessage = ref('')
const blockedJson = ref('{}')
const tabs = [
  { label: '规则管理', value: 'rules' },
  { label: '告警事件', value: 'events' },
  { label: '告警屏蔽', value: 'mutes' },
  { label: '告警订阅', value: 'subscriptions' },
  { label: '工作流', value: 'workflows' },
]
const copy = {
  rules: { title: '规则管理', desc: '按成熟基础监控结构承载规则筛选、列配置、批量动作、导入入口、启停、克隆、删除和 TryRun。' },
  events: { title: '告警事件', desc: '按当前/历史、级别、状态、数据源和对象筛选事件，详情抽屉内处理 ack、assign、resolve、archive。' },
  mutes: { title: '告警屏蔽', desc: '屏蔽规则需要独立 contract，未接入前只展示阻断态。' },
  subscriptions: { title: '告警订阅', desc: '订阅规则需要独立 contract，未接入前只展示阻断态。' },
  workflows: { title: '工作流', desc: '工作流并入 AI SRE 域，但保留事件流水线的列表、编辑、调试和执行记录结构。' },
}
const current = computed(() => copy[section.value] || copy.events)

const syncRoute = value => router.replace({ path: '/alerts', query: { section: value } })
const openBlocked = (action, context = {}) => {
  blockedMessage.value = blockedContracts[action] || 'BLOCKED_BY_CONTRACT：后端 contract 未暴露。'
  blockedJson.value = blockedPayload(action, context)
  blockedVisible.value = true
}

watch(() => route.query.section, value => {
  section.value = validSections.has(value) ? value : 'events'
})
</script>

<style scoped>
.alerting-page { min-height: 100%; padding: 18px; color: #25324a; background: #f5f7fb; }
.alerting-shell { min-height: calc(100vh - 104px); }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 16px; padding: 16px; border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2, p { margin: 0; }
h2 { margin-top: 4px; color: #17233c; font-size: 22px; }
p { margin-top: 8px; color: #66758d; font-size: 13px; line-height: 1.6; }
.blocked-panel { padding: 16px; border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.json-view { margin-top: 12px; }
@media (max-width: 760px) { .page-head { align-items: stretch; flex-direction: column; } }
</style>
