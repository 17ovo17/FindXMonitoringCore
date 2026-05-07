<template>
  <div class="workbench-page">
    <section class="shell-panel">
      <div class="section-head">
        <div>
          <div class="kicker">资产中心</div>
          <h2>{{ current.title }}</h2>
          <p>{{ current.desc }}</p>
        </div>
        <div class="head-actions">
          <el-button v-if="section === 'overview'" @click="go('resource-groups')">资源组</el-button>
          <el-button v-if="section === 'overview'" type="primary" @click="go('hosts')">主机资产</el-button>
          <el-button v-if="section === 'hosts'" @click="go('resource-groups')">管理资源组</el-button>
        </div>
      </div>

      <BusinessWorkspacePanel v-if="section === 'business'" class="content-block" @count-change="businessCount = $event" />
      <HostAssetsPanel v-else-if="section === 'hosts'" class="content-block" @count-change="hostCount = $event" @online-change="onlineAgentCount = $event" />
      <ResourceGroupsPanel v-else-if="section === 'resource-groups'" class="content-block" @count-change="resourceGroupCount = $event" />
      <section v-else-if="section === 'agents'" class="content-block">
        <div class="summary-grid">
          <div class="summary-item" v-loading="agentLoading"><strong>{{ onlineAgentCount }}</strong><span>在线 FindX Agent</span></div>
          <div class="summary-item" v-loading="agentLoading"><strong>{{ agentTotal }}</strong><span>FindX Agent 总数</span></div>
          <div class="summary-item" v-loading="hostLoading"><strong>{{ hostCount }}</strong><span>关联主机</span></div>
        </div>
        <el-alert v-if="agentError" class="state-alert" :type="agentPermission ? 'warning' : 'error'" show-icon :closable="false" :title="agentPermission ? '权限不足' : '加载失败'" :description="agentError" />
        <el-button class="agent-entry" type="primary" @click="router.push({ path: '/agents', query: { section: 'overview' } })">进入 FindX Agent 管理中心</el-button>
        <el-empty v-if="!agentLoading && !agentError && !agentTotal" class="empty-box" description="暂无 FindX Agent 存活数据" />
      </section>
      <section v-else class="content-block">
        <div class="summary-grid">
          <div class="summary-item" v-loading="businessLoading"><strong>{{ businessCount }}</strong><span>业务空间</span></div>
          <div class="summary-item" v-loading="groupLoading"><strong>{{ resourceGroupCount }}</strong><span>资源组</span></div>
          <div class="summary-item" v-loading="hostLoading"><strong>{{ hostCount }}</strong><span>主机资产</span></div>
          <div class="summary-item" v-loading="agentLoading"><strong>{{ onlineAgentCount }}</strong><span>在线 FindX Agent</span></div>
        </div>
        <el-alert v-if="overviewError" class="state-alert" :type="overviewPermission ? 'warning' : 'error'" show-icon :closable="false" :title="overviewPermission ? '权限不足' : '加载失败'" :description="overviewError" />
        <div class="quick-row">
          <el-button @click="go('business')">查看业务空间</el-button>
          <el-button @click="go('resource-groups')">查看资源组</el-button>
          <el-button type="primary" @click="go('hosts')">查看主机资产</el-button>
        </div>
      </section>
    </section>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BusinessWorkspacePanel from '../components/assets/BusinessWorkspacePanel.vue'
import HostAssetsPanel from '../components/assets/HostAssetsPanel.vue'
import ResourceGroupsPanel from '../components/assets/ResourceGroupsPanel.vue'
import { assetsApi, formatAssetError, isPermissionError, normalizeList } from '../api/assets'
import { workspaceApi } from '../api/workspaces'

const route = useRoute()
const router = useRouter()
const validSections = new Set(['overview', 'business', 'hosts', 'agents', 'resource-groups'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'overview')

const businessCount = ref(0)
const resourceGroupCount = ref(0)
const hostCount = ref(0)
const onlineAgentCount = ref(0)
const agentTotal = ref(0)
const businessLoading = ref(false)
const groupLoading = ref(false)
const hostLoading = ref(false)
const agentLoading = ref(false)
const overviewError = ref('')
const agentError = ref('')
const overviewPermission = ref(false)
const agentPermission = ref(false)

const copy = {
  overview: { title: '资产中心', desc: '统一查看业务空间、资源组、主机资产和探针存活摘要，所有统计均来自真实接口。' },
  business: { title: '业务空间', desc: '按业务边界组织主机、端点、负责人和状态。' },
  hosts: { title: '主机资产', desc: '管理主机标签、资源组绑定和业务空间绑定。' },
  agents: { title: 'FindX Agent', desc: '查看 FindX Agent 存活摘要，并进入安装、凭据、远程运维和插件配置控制面。' },
  'resource-groups': { title: '资源组', desc: '管理资产中心资源组，并为主机资产提供归集边界。' },
}
const current = computed(() => copy[section.value] || copy.overview)
const go = target => router.push({ path: '/assets', query: { section: target } })

const setError = (target, permissionTarget, error, name) => {
  permissionTarget.value = isPermissionError(error)
  target.value = formatAssetError(error, name)
}

const loadOverview = async () => {
  overviewError.value = ''
  overviewPermission.value = false
  await Promise.all([loadBusinessCount(), loadGroupCount(), loadHostCount(), loadAgentSummary(true)])
}

const loadBusinessCount = async () => {
  businessLoading.value = true
  try { businessCount.value = normalizeList(await workspaceApi.list()).length } catch (e) { businessCount.value = 0; setError(overviewError, overviewPermission, e, '业务空间') } finally { businessLoading.value = false }
}

const loadGroupCount = async () => {
  groupLoading.value = true
  try { resourceGroupCount.value = normalizeList(await assetsApi.listResourceGroups()).length } catch (e) { resourceGroupCount.value = 0; setError(overviewError, overviewPermission, e, '资源组') } finally { groupLoading.value = false }
}

const loadHostCount = async () => {
  hostLoading.value = true
  try {
    const hosts = normalizeList(await assetsApi.listHostAssets())
    hostCount.value = hosts.length
    onlineAgentCount.value = hosts.filter(host => host.online === true || host.status === 'online' || host.agent_status === 'online').length
  } catch (e) {
    hostCount.value = 0
    setError(overviewError, overviewPermission, e, '主机资产')
  } finally {
    hostLoading.value = false
  }
}

const loadAgentSummary = async (forOverview = false) => {
  agentLoading.value = true
  if (!forOverview) {
    agentError.value = ''
    agentPermission.value = false
  }
  try {
    const agents = normalizeList(await assetsApi.listAgents())
    agentTotal.value = agents.length
    onlineAgentCount.value = agents.filter(agent => agent.online === true || agent.status === 'online').length
  } catch (e) {
    agentTotal.value = 0
    if (forOverview) setError(overviewError, overviewPermission, e, 'FindX Agent 存活摘要')
    else setError(agentError, agentPermission, e, 'FindX Agent 存活摘要')
  } finally {
    agentLoading.value = false
  }
}

watch(section, value => {
  if (value === 'overview') loadOverview()
  if (value === 'agents') {
    loadAgentSummary(false)
    loadHostCount()
  }
}, { immediate: true })
</script>

<style scoped>
.workbench-page { min-height: 100%; padding: 24px; color: #243553; }
.shell-panel { min-height: calc(100vh - 114px); border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto; padding: 22px; }
.section-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.head-actions { display: flex; gap: 10px; flex-wrap: wrap; justify-content: flex-end; }
.content-block { margin-top: 18px; }
.summary-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; }
.summary-item { border: 1px solid #e4e9f2; border-radius: 8px; padding: 14px 16px; background: #f8fbff; min-height: 82px; }
.summary-item strong { display: block; color: #1769ff; font-size: 24px; }
.summary-item span { color: #60728e; font-size: 12px; }
.state-alert { margin-top: 14px; border-radius: 8px; }
.quick-row { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 16px; }
.agent-entry { margin-top: 14px; }
.empty-box { min-height: 300px; border: 1px dashed #d8e1ee; border-radius: 8px; background: #f8fbff; margin-top: 16px; }
@media (max-width: 900px) { .summary-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 640px) { .section-head { flex-direction: column; } .summary-grid { grid-template-columns: 1fr; } }
</style>
