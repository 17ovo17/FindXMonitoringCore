<template>
  <div class="query-page">
    <section class="query-panel">
      <div class="section-head">
        <div>
          <div class="kicker">数据查询</div>
          <h2>{{ current.title }}</h2>
          <p>{{ current.desc }}</p>
        </div>
      </div>

      <MonitorDatasourceQueryPanel v-if="section === 'datasources' || section === 'metrics'" />
      <MetricsMapping v-else-if="section === 'metric-mapping'" />
      <div v-else class="empty-state">
        <el-empty :description="current.empty">
          <el-alert :title="current.hint" type="info" show-icon :closable="false" />
        </el-empty>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'
import MonitorDatasourceQueryPanel from '../components/monitoring/MonitorDatasourceQueryPanel.vue'

const route = useRoute()
const validSections = new Set(['datasources', 'metrics', 'logs', 'traces', 'metric-mapping'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'datasources')
const copy = {
  datasources: { title: '数据源', desc: '管理并验证 Prometheus 等数据源连通性。' },
  metrics: { title: '指标查询', desc: '执行 PromQL 即时或区间查询，查看真实返回。' },
  logs: {
    title: '日志查询',
    desc: '日志检索入口预留。',
    empty: '日志查询后端能力尚未接入。',
    hint: '当前保持真实空态，待日志存储和查询接口就绪后展示。',
  },
  traces: {
    title: 'Trace 查询',
    desc: '链路追踪查询入口预留。',
    empty: 'Trace 查询后端能力尚未接入。',
    hint: '当前保持真实空态，待 Trace 数据源接入后展示。',
  },
  'metric-mapping': { title: '指标映射', desc: '维护原始指标到标准指标语义的映射关系。' },
}
const current = computed(() => copy[section.value])
const MetricsMapping = defineAsyncComponent(() => import('./MetricsMapping.vue'))
</script>

<style scoped>
.query-page { min-height: 100%; padding: 24px; color: #243553; }
.query-panel { min-height: calc(100vh - 114px); padding: 22px; border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto; }
.section-head { margin-bottom: 16px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.empty-state { min-height: 420px; display: grid; place-items: center; border: 1px dashed #d8e1ee; border-radius: 8px; background: #f8fbff; }
:deep(.mm-page) { padding: 0; }
:deep(.config-page) { padding: 0; min-height: auto; }
</style>
