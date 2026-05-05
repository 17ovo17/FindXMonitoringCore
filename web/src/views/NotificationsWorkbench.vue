<template>
  <div class="notify-page">
    <section class="notify-panel">
      <div class="section-head">
        <div>
          <div class="kicker">通知</div>
          <h2>{{ current.title }}</h2>
          <p>{{ current.desc }}</p>
        </div>
      </div>
      <div class="empty-state">
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

const route = useRoute()
const validSections = new Set(['rules', 'channels'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'rules')
const copy = {
  rules: {
    title: '通知规则',
    desc: '按事件条件、接收对象和媒介策略决定通知如何投递，订阅偏好、消息正文和投递追踪在规则详情内承载。',
    empty: '通知规则尚未接入真实接口。',
    hint: '接收对象复用用户、团队组织和业务空间，不再维护独立接收组。',
  },
  channels: {
    title: '通知媒介',
    desc: '邮件、Webhook、飞书、钉钉、企微、PagerDuty 等媒介统一在这里配置。',
    empty: '通知媒介尚未接入真实接口。',
    hint: '后续响应只返回 secret_set 和目标摘要，不回显 Webhook Token。',
  },
}
const current = computed(() => copy[section.value] || copy.rules)
</script>

<style scoped>
.notify-page { min-height: 100%; padding: 24px; color: #243553; }
.notify-panel { min-height: calc(100vh - 114px); padding: 22px; border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.empty-state { min-height: 420px; display: grid; place-items: center; border: 1px dashed #d8e1ee; border-radius: 8px; margin-top: 18px; background: #f8fbff; }
</style>
