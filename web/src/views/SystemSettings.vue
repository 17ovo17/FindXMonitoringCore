<template>
  <div class="sys-page">
    <section class="page-head">
      <div>
        <div class="panel-kicker">Platform Governance</div>
        <h2>系统配置</h2>
        <p class="page-desc">
          平台系统配置页保留治理说明和入口指引；已拆分到一级导航的能力不在这里重复承载。
        </p>
      </div>
    </section>

    <section class="settings-guide">
      <div v-for="item in guideItems" :key="item.key" class="guide-item">
        <div>
          <h3>{{ item.title }}</h3>
          <p>{{ item.desc }}</p>
        </div>
        <el-button text type="primary" @click="router.push(item.to)">前往</el-button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const guideItems = [
  {
    key: 'datasources',
    title: '数据源与查询',
    desc: '数据源、指标查询和指标映射已归属数据查询导航。',
    to: { path: '/query', query: { section: 'datasources' } },
  },
  {
    key: 'credentials',
    title: '凭证资产',
    desc: '凭证资产已归属基础设施导航，系统配置页不重复展示敏感配置入口。',
    to: { path: '/assets', query: { section: 'credentials' } },
  },
  {
    key: 'self-check',
    title: '平台运行自检',
    desc: '平台自身依赖和治理链路自检已归属平台治理导航。',
    to: { path: '/platform', query: { section: 'health' } },
  },
  {
    key: 'notifications',
    title: '通知策略',
    desc: '通知媒介、模板、记录和全局参数已归属通知导航。',
    to: { path: '/notifications', query: { section: 'rules' } },
  },
]
</script>

<style scoped>
.sys-page { padding: 28px 32px; height: 100%; color: #243553; display: flex; flex-direction: column; gap: 18px; }
.page-head, .settings-guide { background: rgba(255,255,255,.86); border: 1px solid #e4e9f2; border-radius: 8px; box-shadow: 0 12px 34px rgba(31,45,61,.06); }
.page-head { padding: 20px 26px; }
.page-head h2 { margin: 6px 0 4px; font-size: 26px; color: #263653; }
.page-desc { margin: 0; font-size: 13px; color: var(--muted); line-height: 1.6; }
.panel-kicker { font-size: 12px; color: #247cff; text-transform: uppercase; font-weight: 800; }
.settings-guide { padding: 8px; display: grid; gap: 8px; }
.guide-item { min-height: 86px; padding: 16px 18px; display: flex; justify-content: space-between; align-items: center; gap: 16px; border: 1px solid #edf1f7; border-radius: 8px; background: #f8fbff; }
.guide-item h3 { margin: 0 0 6px; color: #1e3a5f; font-size: 16px; }
.guide-item p { margin: 0; color: #60728e; font-size: 13px; line-height: 1.6; }
@media (max-width: 760px) {
  .sys-page { padding: 18px; }
  .guide-item { align-items: flex-start; flex-direction: column; }
}
</style>
