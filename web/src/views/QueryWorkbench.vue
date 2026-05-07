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
      <div v-else class="blocked-state">
        <div class="blocked-head">
          <el-tag type="danger" effect="dark" round>BLOCKED</el-tag>
          <strong>{{ current.empty }}</strong>
        </div>
        <el-alert :title="current.hint" type="error" show-icon :closable="false" />
        <div class="check-grid">
          <div v-for="item in current.checks" :key="item" class="check-item">{{ item }}</div>
        </div>
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
    desc: '日志检索入口预留，后续接入真实日志服务后展示真实数据。',
    empty: '日志查询服务尚未接入',
    hint: '当前不会展示示例日志或静态假数据；请从日志中心查看阻断项。',
    checks: ['日志采集管道', '解析规则', '查询存储', '权限与脱敏'],
  },
  traces: {
    title: 'Trace 查询',
    desc: '链路追踪查询入口预留，后续接入真实 Trace 存储和服务分析。',
    empty: 'Trace 查询服务尚未接入',
    hint: '当前不会展示示例 Trace 或静态假数据；请从链路监控查看阻断项。',
    checks: ['Trace 采集入口', '链路分析服务', '存储索引', '权限与审计'],
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
.blocked-state { min-height: 420px; display: flex; flex-direction: column; justify-content: center; gap: 16px; border: 1px dashed #d8e1ee; border-radius: 8px; background: #f8fbff; padding: 22px; }
.blocked-head { display: flex; align-items: center; gap: 10px; color: #1e3a5f; }
.check-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px; }
.check-item { border: 1px solid #e4e9f2; border-radius: 8px; padding: 12px; color: #60728e; background: #fff; font-size: 12px; }
:deep(.config-page) { padding: 0; min-height: auto; }
@media (max-width: 900px) { .check-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 560px) { .check-grid { grid-template-columns: 1fr; } }
</style>
