<template>
  <div class="integration-page">
    <section class="integration-panel">
      <div class="section-head">
        <div>
          <div class="kicker">集成中心</div>
          <h2>{{ current.title }}</h2>
          <p>{{ current.desc }}</p>
        </div>
      </div>

      <div class="integration-grid">
        <button v-for="item in entries" :key="item.title" type="button" class="entry-card" @click="router.push(item.to)">
          <strong>{{ item.title }}</strong>
          <span>{{ item.desc }}</span>
          <em>{{ item.status }}</em>
        </button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const validSections = new Set(['overview', 'collectors', 'templates'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'overview')

const copy = {
  overview: { title: '集成总览', desc: '统一查看 FindX 可接入能力，所有入口均跳转到自有页面，不嵌入外部系统。' },
  collectors: { title: '采集接入', desc: '采集类集成统一落到 FindX Agent 管理中心，未接入项保持明确阻断态。' },
  templates: { title: '模板导入', desc: '模板能力待真实接口和导入校验接入后启用。' },
}
const current = computed(() => copy[section.value] || copy.overview)

const entries = [
  { title: 'FindX Agent', desc: '安装计划、凭据引用和远程运维入口。', status: '可进入', to: { path: '/agents', query: { section: 'overview' } } },
  { title: '链路监控', desc: '链路采集和分析服务未配置，进入后显示 BLOCKED。', status: 'BLOCKED', to: { path: '/tracing', query: { section: 'overview' } } },
  { title: '日志中心', desc: '日志服务未配置，进入后显示 BLOCKED。', status: 'BLOCKED', to: { path: '/logs', query: { section: 'overview' } } },
  { title: '数据源', desc: '进入 FindX 数据查询的数据源配置。', status: '可进入', to: { path: '/query', query: { section: 'datasources' } } },
]
</script>

<style scoped>
.integration-page { min-height: 100%; padding: 24px; color: #243553; }
.integration-panel { min-height: calc(100vh - 114px); padding: 22px; border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.88); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto; }
.section-head { margin-bottom: 18px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.integration-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; }
.entry-card { min-height: 132px; border: 1px solid #e4e9f2; border-radius: 8px; padding: 14px 16px; background: #f8fbff; text-align: left; cursor: pointer; }
.entry-card:hover { border-color: rgba(23,105,255,.34); box-shadow: 0 10px 24px rgba(31,45,61,.08); }
.entry-card strong { display: block; color: #1e3a5f; font-size: 14px; }
.entry-card span { display: block; min-height: 42px; margin-top: 8px; color: #60728e; font-size: 12px; line-height: 1.6; }
.entry-card em { display: inline-flex; margin-top: 12px; color: #1769ff; font-size: 12px; font-style: normal; font-weight: 800; }
@media (max-width: 1100px) { .integration-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 640px) { .integration-grid { grid-template-columns: 1fr; } }
</style>
