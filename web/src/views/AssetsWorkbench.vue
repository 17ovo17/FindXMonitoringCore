<template>
  <div class="workbench-page">
    <section v-if="section === 'agents'" class="shell-panel">
      <div class="agent-shell">
        <div class="section-head">
          <div>
            <div class="kicker">采集接入</div>
            <h2>探针与采集</h2>
            <p>承载接入记录、远程安装、本机安装指令、凭证引用、采集配置、插件配置、远程调试和自升级。高风险执行能力后续必须进入审批控制面。</p>
          </div>
        </div>
        <div class="empty-state">
          <el-empty description="探针与采集入口已预留，待控制面验收完成后接入真实资产数据。" />
        </div>
      </div>
    </section>
    <section v-else class="shell-panel">
      <div class="section-head">
        <div>
          <div class="kicker">基础设施</div>
          <h2>{{ current.title }}</h2>
          <p>{{ current.desc }}</p>
        </div>
        <el-button v-if="section === 'hosts'" type="primary" @click="goDiagnosis">带主机上下文问诊</el-button>
      </div>
      <div class="empty-state">
        <el-empty :description="current.empty">
          <el-button v-if="section === 'business'" @click="router.push('/assets?section=business')">查看业务空间</el-button>
          <el-button v-else-if="section === 'hosts'" type="primary" @click="goDiagnosis">从主机上下文发起问诊</el-button>
        </el-empty>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const validSections = new Set(['overview', 'business', 'hosts', 'agents'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'overview')

const copies = {
  overview: {
    title: '资产中心',
    desc: '统一承载业务空间、主机、CMDB、采集接入记录、探针存活和安全配置引用。',
    empty: '资产聚合总览正在接入真实统计，当前不展示占位数据。',
  },
  business: {
    title: '业务空间',
    desc: '按业务域组织主机、指标、告警、仪表盘和责任人。',
    empty: '业务空间视图待接入组织关系后展示。',
  },
  hosts: {
    title: '主机资产',
    desc: '主机资产与 CMDB 信息在这里统一承载，探针存活可从主机详情关联查看。',
    empty: '主机资产列表待接入 CMDB 或探针资产后展示。',
  },
}
const current = computed(() => copies[section.value] || copies.overview)
const goDiagnosis = () => router.push({ path: '/aiops', query: { section: 'diagnosis' } })
</script>

<style scoped>
.workbench-page { min-height: 100%; padding: 24px; color: #243553; }
.shell-panel {
  min-height: calc(100vh - 114px); border: 1px solid #e4e9f2; border-radius: 8px;
  background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto;
}
.shell-panel { padding: 22px; }
.agent-shell { padding: 18px 20px; }
.section-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.empty-state { min-height: 420px; display: grid; place-items: center; border: 1px dashed #d8e1ee; border-radius: 8px; margin-top: 18px; background: #f8fbff; }
@media (max-width: 760px) { .section-head { flex-direction: column; } }
</style>
