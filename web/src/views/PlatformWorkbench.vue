<template>
  <div class="platform-page">
    <section class="embedded-panel">
      <AIConfig v-if="section === 'models'" />
      <SystemSettings v-else-if="section === 'settings'" />
      <div v-else-if="section === 'health'" class="platform-self-check">
        <div class="section-head">
          <div>
            <div class="kicker">平台治理</div>
            <h2>平台运行自检</h2>
            <p>聚焦平台自身依赖、配置链路和治理能力的运行状态；主机、服务和采集端检查归属 AI SRE 与探针控制面。</p>
          </div>
        </div>
        <div class="empty-state">
          <el-empty description="平台运行自检入口待接入真实接口，当前不展示占位数据。">
            <el-alert title="历史健康入口仅作为平台运行自检兼容路由，不承载主机、服务或采集端巡检。" type="info" show-icon :closable="false" />
          </el-empty>
        </div>
      </div>
      <HealthAuditPanel v-else />
    </section>
  </div>
</template>

<script setup>
import { computed, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const validSections = new Set(['models', 'settings', 'health', 'audit'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'models')

const AIConfig = defineAsyncComponent(() => import('./AIConfig.vue'))
const SystemSettings = defineAsyncComponent(() => import('./SystemSettings.vue'))
const HealthAuditPanel = defineAsyncComponent(() => import('./HealthAuditPanel.vue'))
</script>

<style scoped>
.platform-page { min-height: 100%; padding: 24px; color: #243553; }
.embedded-panel { min-height: calc(100vh - 114px); border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto; }
:deep(.config-page), :deep(.sys-page), :deep(.health-audit-page) { height: auto; min-height: calc(100vh - 116px); padding: 24px; }
.platform-self-check { padding: 24px; min-height: calc(100vh - 116px); }
.section-head { margin-bottom: 16px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.empty-state { min-height: 420px; display: grid; place-items: center; border: 1px dashed #d8e1ee; border-radius: 8px; background: #f8fbff; }
</style>
