<template>
  <div class="notify-page">
    <section class="notify-shell">
      <header class="page-head">
        <div>
          <div class="kicker">通知</div>
          <h2>{{ current.title }}</h2>
          <p>{{ current.desc }}</p>
        </div>
        <el-segmented v-model="section" :options="tabs" @change="syncRoute" />
      </header>

      <NotificationRulesPanel v-if="section === 'rules'" @blocked="openBlocked" />
      <NotificationChannelsPanel v-else-if="section === 'channels'" @blocked="openBlocked" />
      <NotificationTemplatesPanel v-else @blocked="openBlocked" />
    </section>

    <NotificationBlockedDrawer v-model:visible="blockedVisible" :message="blockedMessage" :payload="blockedJson" />
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import NotificationBlockedDrawer from '../components/notifications/NotificationBlockedDrawer.vue'
import NotificationChannelsPanel from '../components/notifications/NotificationChannelsPanel.vue'
import NotificationRulesPanel from '../components/notifications/NotificationRulesPanel.vue'
import NotificationTemplatesPanel from '../components/notifications/NotificationTemplatesPanel.vue'
import { blockedContracts, blockedPayload } from '../components/notifications/notificationModel'

const route = useRoute()
const router = useRouter()
const validSections = new Set(['rules', 'channels', 'templates'])
const section = ref(validSections.has(route.query.section) ? route.query.section : 'rules')
const blockedVisible = ref(false)
const blockedMessage = ref('')
const blockedJson = ref('{}')
const tabs = [
  { label: '通知规则', value: 'rules' },
  { label: '通知媒介', value: 'channels' },
  { label: '消息模板', value: 'templates' },
]
const copy = {
  rules: { title: '通知规则', desc: '按成熟基础监控结构承载搜索、新增、启停、编辑、克隆、删除和详情统计；未暴露 contract 时明确阻断。' },
  channels: { title: '通知媒介', desc: '左侧媒介类型、右侧筛选表格、导入导出、启停、克隆和删除按成熟结构组织。' },
  templates: { title: '消息模板', desc: '左侧模板列表、右侧模板详情、内容编辑、预览、克隆和删除按成熟结构组织。' },
}
const current = computed(() => copy[section.value] || copy.rules)

const syncRoute = value => router.replace({ path: '/notifications', query: { section: value } })
const openBlocked = (action, context = {}) => {
  blockedMessage.value = blockedContracts[action] || 'BLOCKED_BY_CONTRACT：后端 contract 未暴露。'
  blockedJson.value = blockedPayload(action, context)
  blockedVisible.value = true
}

watch(() => route.query.section, value => {
  section.value = validSections.has(value) ? value : 'rules'
})
</script>

<style scoped>
.notify-page { min-height: 100%; padding: 18px; color: #25324a; background: #f5f7fb; }
.notify-shell { min-height: calc(100vh - 104px); }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 16px; padding: 16px; border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2, p { margin: 0; }
h2 { margin-top: 4px; color: #17233c; font-size: 22px; }
p { margin-top: 8px; color: #66758d; font-size: 13px; line-height: 1.6; }
@media (max-width: 760px) { .page-head { align-items: stretch; flex-direction: column; } }
</style>
