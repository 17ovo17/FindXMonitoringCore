<template>
  <div class="gate-page">
    <section class="gate-panel">
      <div class="section-head">
        <div>
          <div class="kicker">日志中心</div>
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
        title="日志服务未配置，当前不会展示模拟日志"
        description="此入口仅保留 FindX 自有路由壳层。待日志采集、解析、存储、查询权限和错误态接入后，才会展示真实日志数据。"
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
const validSections = new Set(['overview', 'query', 'pipelines'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'overview')

const copy = {
  overview: { title: '日志总览', desc: '展示 FindX 日志中心接入状态、阻断项和后续检查清单。' },
  query: { title: '日志检索', desc: '日志检索待真实日志服务、索引和查询契约接入后启用。' },
  pipelines: { title: '接入管道', desc: '日志管道待采集配置、解析规则和投递结果可验证后启用。' },
}
const current = computed(() => copy[section.value] || copy.overview)

const checks = [
  { title: '日志采集', desc: '等待主机、应用和容器日志进入 FindX 日志管道。' },
  { title: '解析规则', desc: '等待字段解析、时间戳、级别和服务标签规则确认。' },
  { title: '查询存储', desc: '等待日志存储、索引和查询 API 完成配置。' },
  { title: '权限与脱敏', desc: '等待项目范围、敏感字段脱敏和查询审计契约完成。' },
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
