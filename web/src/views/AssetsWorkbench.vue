<template>
  <div class="workbench-page">
    <section v-if="section === 'business'" class="shell-panel">
      <div class="section-head">
        <div>
          <div class="kicker">基础设施</div>
          <h2>业务空间</h2>
          <p>按业务边界组织主机、端点、负责人和状态，所有数据来自后端业务空间接口。</p>
        </div>
      </div>
      <BusinessWorkspacePanel class="content-block" @count-change="businessCount = $event" />
    </section>

    <section v-else-if="section === 'agents'" class="shell-panel">
      <div class="agent-shell">
        <div class="section-head">
          <div>
            <div class="kicker">采集接入</div>
            <h2>探针与采集</h2>
            <p>承载接入记录、远程安装、本机安装指令、凭证引用、采集配置、插件配置、远程调试和自升级。</p>
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
      <div v-if="section === 'overview'" class="overview-grid">
        <div class="overview-item" v-loading="overviewLoading">
          <strong>{{ businessCount }}</strong>
          <span>业务空间</span>
        </div>
      </div>
      <el-alert
        v-if="section === 'overview' && overviewError"
        class="overview-error"
        type="error"
        show-icon
        :closable="false"
        title="业务空间数量加载失败"
        :description="overviewError"
      />
      <div class="empty-state">
        <el-empty :description="current.empty">
          <el-button v-if="section === 'overview'" @click="router.push('/assets?section=business')">查看业务空间</el-button>
          <el-button v-else-if="section === 'hosts'" type="primary" @click="goDiagnosis">从主机上下文发起问诊</el-button>
        </el-empty>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BusinessWorkspacePanel from '../components/assets/BusinessWorkspacePanel.vue'
import { normalizeList, redactText, workspaceApi } from '../api/workspaces'

const route = useRoute()
const router = useRouter()
const validSections = new Set(['overview', 'business', 'hosts', 'agents'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'overview')
const businessCount = ref(0)
const overviewLoading = ref(false)
const overviewError = ref('')

const copies = {
  overview: {
    title: '资产中心',
    desc: '统一承载业务空间、主机、采集接入记录、探针存活和安全配置引用。',
    empty: '资产聚合总览正在接入真实统计；当前只展示已从业务空间接口获取到的数量。',
  },
  hosts: {
    title: '主机资产',
    desc: '主机资产与运行上下文将在这里统一承载，探针存活可从主机详情关联查看。',
    empty: '主机资产列表待接入真实接口后展示。',
  },
}
const current = computed(() => copies[section.value] || copies.overview)
const goDiagnosis = () => router.push({ path: '/aiops', query: { section: 'diagnosis' } })
const loadOverview = async () => {
  overviewLoading.value = true
  overviewError.value = ''
  try {
    businessCount.value = normalizeList(await workspaceApi.list()).length
  } catch (error) {
    businessCount.value = 0
    overviewError.value = redactText(error.message || '业务空间数量加载失败')
  } finally {
    overviewLoading.value = false
  }
}
watch(section, value => {
  if (value === 'overview') loadOverview()
}, { immediate: true })
</script>

<style scoped>
.workbench-page { min-height: 100%; padding: 24px; color: #243553; }
.shell-panel {
  min-height: calc(100vh - 114px);
  border: 1px solid #e4e9f2;
  border-radius: 8px;
  background: rgba(255,255,255,.86);
  box-shadow: 0 12px 34px rgba(31,45,61,.06);
  overflow: auto;
  padding: 22px;
}
.agent-shell { padding: 18px 20px; }
.section-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.content-block { margin-top: 18px; }
.overview-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; margin-top: 18px; }
.overview-item { border: 1px solid #e4e9f2; border-radius: 8px; padding: 14px 16px; background: #f8fbff; }
.overview-item strong { display: block; color: #1769ff; font-size: 24px; }
.overview-item span { color: #60728e; font-size: 12px; }
.overview-error { margin-top: 14px; border-radius: 8px; }
.empty-state {
  min-height: 360px;
  display: grid;
  place-items: center;
  border: 1px dashed #d8e1ee;
  border-radius: 8px;
  margin-top: 18px;
  background: #f8fbff;
}
@media (max-width: 760px) {
  .section-head { flex-direction: column; }
  .overview-grid { grid-template-columns: 1fr; }
}
</style>
