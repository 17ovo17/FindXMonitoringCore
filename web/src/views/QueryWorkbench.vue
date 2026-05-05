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
import MonitorDatasourceQueryPanel from '../components/monitoring/MonitorDatasourceQueryPanel.vue'

const route = useRoute()
const validSections = new Set(['datasources', 'metrics', 'logs', 'traces'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'datasources')
const copy = {
  datasources: { title: '数据源', desc: '统一管理并验证 Prometheus、VictoriaMetrics、Doris、Loki、OpenSearch 等数据源连通性。' },
  metrics: { title: '指标查询', desc: '执行 PromQL 即时或区间查询，支持后续补充指标目录和下拉检索能力。' },
  logs: {
    title: '日志查询',
    desc: '日志检索入口预留，后续接入 Doris 或兼容日志存储后展示真实数据。',
    empty: '日志查询后端能力尚未接入。',
    hint: '当前保持真实空态，不展示示例日志或静态假数据。',
  },
  traces: {
    title: 'Trace 查询',
    desc: '链路追踪查询入口预留，后续接入 OpenTelemetry Trace 存储和服务分析。',
    empty: 'Trace 查询后端能力尚未接入。',
    hint: '当前保持真实空态，待 Trace 数据源接入后展示。',
  },
}
const current = computed(() => copy[section.value])
</script>

<style scoped>
.query-page { min-height: 100%; padding: 24px; color: #243553; }
.query-panel { min-height: calc(100vh - 114px); padding: 22px; border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto; }
.section-head { margin-bottom: 16px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.empty-state { min-height: 420px; display: grid; place-items: center; border: 1px dashed #d8e1ee; border-radius: 8px; background: #f8fbff; }
:deep(.config-page) { padding: 0; min-height: auto; }
</style>
