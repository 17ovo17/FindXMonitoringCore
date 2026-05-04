<template>
  <div class="dash-page">
    <section class="dash-panel">
      <div class="section-head">
        <div>
          <div class="kicker">监控仪表盘</div>
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
const validSections = new Set(['list', 'templates', 'shares'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'list')
const copy = {
  list: {
    title: '仪表盘列表',
    desc: 'Dashboard 作为监控中心核心域独立承载。',
    empty: '仪表盘列表尚未接入真实接口。',
    hint: '不会展示示例看板，待后端看板能力就绪后展示真实数据。',
  },
  templates: {
    title: '仪表盘模板',
    desc: '管理可复用的仪表盘模板。',
    empty: '仪表盘模板尚未接入真实接口。',
    hint: '模板中心仍在集成中心下独立承载通用集成模板。',
  },
  shares: {
    title: '图表分享',
    desc: '沉淀图表分享链接与访问权限。',
    empty: '暂无图表分享记录。',
    hint: '当前未发现分享接口，保持真实空态。',
  },
}
const current = computed(() => copy[section.value])
</script>

<style scoped>
.dash-page { min-height: 100%; padding: 24px; color: #243553; }
.dash-panel { min-height: calc(100vh - 114px); padding: 22px; border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.empty-state { min-height: 420px; display: grid; place-items: center; border: 1px dashed #d8e1ee; border-radius: 8px; margin-top: 18px; background: #f8fbff; }
</style>
