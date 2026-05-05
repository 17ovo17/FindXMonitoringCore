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
      <MonitorEventsPanel v-else />
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import MonitorAlertRulesPanel from '../components/monitoring/MonitorAlertRulesPanel.vue'
import MonitorEventsPanel from '../components/monitoring/MonitorEventsPanel.vue'

const route = useRoute()
const validSections = new Set(['rules', 'events'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'events')
const copy = {
  rules: { title: '告警规则', desc: '统一承载阈值规则、智能检测规则、预计算规则、模板导入和事件路由规则配置。' },
  events: { title: '事件中心', desc: '查看真实告警事件、状态流转、聚合总览、静默动作和处理记录。' },
}
const current = computed(() => copy[section.value] || copy.events)
</script>

<style scoped>
.alerting-page { min-height: 100%; padding: 24px; color: #243553; }
.embedded-panel, .alerting-panel { min-height: calc(100vh - 114px); border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto; }
.alerting-panel { padding: 22px; }
.section-head { margin-bottom: 16px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
:deep(.alerts-page) { min-height: auto; padding: 24px; }
</style>
