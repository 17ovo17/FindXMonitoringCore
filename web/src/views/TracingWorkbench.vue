<template>
  <div class="gate-page">
    <section class="gate-panel">
      <div class="section-head">
        <div>
          <div class="kicker">链路监控</div>
          <h2>{{ current.title }}</h2>
          <p>{{ current.desc }}</p>
        </div>
        <el-tag type="danger" effect="dark" round>BLOCKED</el-tag>
      </div>

      <el-alert
        class="state-alert"
        type="error"
        show-icon
        :closable="false"
        title="链路服务未配置，当前不会展示模拟 Trace 数据"
        description="此入口仅保留 FindX 自有路由壳层。待链路采集、分析服务、存储和权限契约全部接入后，才会展示真实调用链。"
      />

      <div class="check-grid">
        <div v-for="item in checks" :key="item.title" class="check-item">
          <strong>{{ item.title }}</strong>
          <span>{{ item.desc }}</span>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const validSections = new Set(['overview', 'services', 'traces'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'overview')

const copy = {
  overview: { title: '链路总览', desc: '展示 FindX 链路监控接入状态、依赖检查和阻断原因。' },
  services: { title: '服务拓扑', desc: '服务依赖视图待真实链路分析服务接入后启用。' },
  traces: { title: 'Trace 检索', desc: 'Trace 检索待采集、存储、索引和鉴权链路打通后启用。' },
}
const current = computed(() => copy[section.value] || copy.overview)

const checks = [
  { title: '采集入口', desc: '等待应用侧 Trace 数据进入 FindX 链路管道。' },
  { title: '分析服务', desc: '等待链路查询、拓扑聚合和错误分析服务完成配置。' },
  { title: '存储索引', desc: '等待 Trace 存储、索引字段和保留策略确认。' },
  { title: '权限与审计', desc: '等待租户、项目、服务范围和查询审计契约完成。' },
]
</script>

<style scoped>
.gate-page { min-height: 100%; padding: 24px; color: #243553; }
.gate-panel { min-height: calc(100vh - 114px); padding: 22px; border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.88); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto; }
.section-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.state-alert { margin-top: 18px; border-radius: 8px; }
.check-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; margin-top: 18px; }
.check-item { min-height: 108px; border: 1px solid #e4e9f2; border-radius: 8px; padding: 14px 16px; background: #f8fbff; }
.check-item strong { display: block; color: #1e3a5f; font-size: 14px; }
.check-item span { display: block; margin-top: 8px; color: #60728e; font-size: 12px; line-height: 1.6; }
@media (max-width: 1100px) { .check-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 640px) { .section-head { flex-direction: column; } .check-grid { grid-template-columns: 1fr; } }
</style>
