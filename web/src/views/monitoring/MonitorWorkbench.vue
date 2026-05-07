<template>
  <div class="monitor-page">
    <header class="page-head">
      <div>
        <div class="kicker">FindX Monitoring</div>
        <h2>监控运维工作台</h2>
        <p>在 FindX 主平台内统一承载监控对象、数据源、查询、告警、事件、通知协同，并把 AI、知识库和工作流作为增强层接入。</p>
      </div>
      <el-button :loading="refreshing" @click="refresh">
        <el-icon><Refresh /></el-icon>刷新当前视图
      </el-button>
    </header>

    <section class="monitor-shell">
      <MonitorNav :groups="monitorGroups" :active-key="activeKey" @select="selectSection" />

      <main class="monitor-content">
        <div class="content-head">
          <div>
            <h3>{{ activeItem.label }}</h3>
            <p>{{ activeItem.desc }}</p>
          </div>
          <el-tag :type="activeItem.statusType || 'info'" effect="plain">{{ activeItem.status }}</el-tag>
        </div>

        <MonitorHealthCards
          v-if="activeKey === 'overview'"
          :health="health"
          :loading="healthLoading"
          :error="healthError"
        />
        <MonitorInventoryPanel v-else-if="activeItem.panel === 'inventory'" ref="inventoryRef" />
        <MonitorDatasourceQueryPanel v-else-if="activeItem.panel === 'query'" ref="queryRef" />
        <MonitorAlertRulesPanel v-else-if="activeItem.panel === 'rules'" ref="rulesRef" />
        <MonitorEventsPanel v-else-if="activeItem.panel === 'events'" ref="eventsRef" />
        <section v-else-if="activeItem.route" class="state-card">
          <el-empty :description="`${activeItem.label} 已由 FindX 平台模块承载`">
            <el-button type="primary" @click="goRoute(activeItem.route)">打开{{ activeItem.label }}</el-button>
          </el-empty>
        </section>
        <section v-else class="state-card">
          <el-empty :description="activeItem.empty || '后端能力未接入，当前无可展示数据'">
            <el-alert :title="activeItem.hint" type="info" show-icon :closable="false" />
          </el-empty>
        </section>
      </main>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Refresh } from '@element-plus/icons-vue'
import { monitorApi, isPermissionError, isUnauthorizedError, redactText } from '../../api/monitor'
import MonitorAlertRulesPanel from '../../components/monitoring/MonitorAlertRulesPanel.vue'
import MonitorDatasourceQueryPanel from '../../components/monitoring/MonitorDatasourceQueryPanel.vue'
import MonitorEventsPanel from '../../components/monitoring/MonitorEventsPanel.vue'
import MonitorHealthCards from '../../components/monitoring/MonitorHealthCards.vue'
import MonitorInventoryPanel from '../../components/monitoring/MonitorInventoryPanel.vue'
import MonitorNav from '../../components/monitoring/MonitorNav.vue'

const route = useRoute()
const router = useRouter()
const health = ref(null)
const healthLoading = ref(false)
const healthError = ref('')
const refreshing = ref(false)
const inventoryRef = ref(null)
const queryRef = ref(null)
const rulesRef = ref(null)
const eventsRef = ref(null)

const pendingHint = '未检测到对应后端接口或权限入口，保持真实空态，待平台接入后展示数据。'
const platformStatus = ['平台模块', 'primary']
const connectedStatus = ['已接入', 'success']
const pendingStatus = ['待接入', 'info']

function item(key, label, icon, statusPair, desc, panel = '', routePath = '') {
  const [status, statusType] = statusPair
  return { key, label, icon, status, statusType, desc, panel, route: routePath, hint: pendingHint }
}

const monitorGroups = [
  {
    label: 'FindX 基础监控',
    items: [
      item('overview', '监控总览', 'DataBoard', connectedStatus, '健康状态、接入模式与核心指标概览'),
      item('business', '业务组', 'FolderOpened', pendingStatus, '按团队、业务域和权限组组织监控视图'),
      item('targets', '监控对象/机器存活', 'Monitor', connectedStatus, '主机、目标和 Agent 在线状态', 'inventory'),
      item('agent', 'Agent/CMDB', 'Connection', connectedStatus, 'FindX Agent 资产与采集能力清单', 'inventory'),
      item('install', '安装任务', 'Box', pendingStatus, 'Agent 安装、升级与任务追踪'),
      item('datasource', '数据源', 'Coin', connectedStatus, 'Prometheus 等数据源配置与连通性验证', 'query'),
      item('query', '即时查询', 'Search', connectedStatus, 'PromQL 即时/区间查询与指标浏览', 'query'),
      item('dashboard', 'Dashboard', 'TrendCharts', pendingStatus, '监控图表、仪表盘和业务视图'),
    ],
  },
  {
    label: '告警与协同',
    items: [
      item('rules', '告警规则', 'Bell', connectedStatus, '规则列表、详情与 TryRun 校验', 'rules'),
      item('events', '事件', 'Warning', connectedStatus, '当前/历史事件与处置动作', 'events'),
      item('notify', '通知', 'Message', pendingStatus, '通知媒介、策略与投递记录'),
      item('silence', '静默', 'MuteNotification', pendingStatus, '维护窗口与告警静默'),
      item('subscription', '订阅', 'Star', pendingStatus, '事件订阅与用户关注'),
      item('oncall', '值班', 'Calendar', platformStatus, '值班组和升级策略', '', '/settings?tab=oncall'),
    ],
  },
  {
    label: 'FindX 增强层',
    items: [
      item('templates', '模板', 'Files', pendingStatus, '规则模板、看板模板和采集模板'),
      item('audit', '审计', 'DocumentChecked', pendingStatus, '操作审计与变更追踪'),
      item('chat', '单机对话', 'ChatLineRound', platformStatus, '面向单机上下文的交互诊断', '', '/workbench'),
      item('ai', 'AI 问诊', 'ChatDotRound', platformStatus, 'FindX 智能诊断入口', '', '/workbench'),
      item('knowledge', '知识库', 'Collection', platformStatus, '诊断知识、语义检索和沉淀', '', '/knowledge'),
      item('workflow', '工作流自愈', 'Operation', platformStatus, '编排处置和自动修复流程', '', '/workflows'),
      item('system', '系统配置', 'Setting', platformStatus, '平台级配置与监控融合设置', '', '/settings'),
    ],
  },
]

const flatItems = computed(() => monitorGroups.flatMap(group => group.items))
const normalizeSection = value => {
  const key = String(value || 'overview')
  return flatItems.value.some(entry => entry.key === key) ? key : 'overview'
}
const activeKey = ref(normalizeSection(route.query.section))
const activeItem = computed(() => flatItems.value.find(entry => entry.key === activeKey.value) || flatItems.value[0])
const permissionMessage = error => isUnauthorizedError(error) ? '登录已过期，请重新登录后继续访问监控能力' : '无权限访问该监控能力'
const formatError = error => isPermissionError(error) ? permissionMessage(error) : redactText(error?.message || '请求失败')

const loadHealth = async () => {
  healthLoading.value = true
  healthError.value = ''
  try {
    health.value = await monitorApi.health()
  } catch (e) {
    healthError.value = formatError(e)
  } finally {
    healthLoading.value = false
  }
}

const activePanelRef = () => ({
  inventory: inventoryRef,
  query: queryRef,
  rules: rulesRef,
  events: eventsRef,
}[activeItem.value.panel])

const refresh = async () => {
  refreshing.value = true
  try {
    if (activeKey.value === 'overview') await loadHealth()
    await activePanelRef()?.value?.load?.()
  } finally {
    refreshing.value = false
  }
}

const selectSection = key => {
  const next = normalizeSection(key)
  activeKey.value = next
  router.replace({ query: { ...route.query, section: next } })
}

const goRoute = path => {
  router.push(path)
}

watch(() => route.query.section, value => {
  const next = normalizeSection(value)
  const current = String(value || 'overview')
  if (next !== activeKey.value) activeKey.value = next
  if (current !== next) router.replace({ query: { ...route.query, section: next } })
}, { immediate: true })

onMounted(loadHealth)
</script>

<style scoped>
.monitor-page {
  min-height: 100%;
  padding: 24px;
  color: #243553;
}
.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin: 0 auto 16px;
  max-width: 1280px;
}
.kicker {
  color: #247cff;
  font-size: 12px;
  font-weight: 800;
  text-transform: uppercase;
}
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.monitor-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  max-width: 1280px;
  margin: 0 auto;
  align-items: start;
}
.monitor-content {
  min-width: 0;
  padding: 16px;
  border: 1px solid #e4e9f2;
  border-radius: 8px;
  background: rgba(255, 255, 255, .88);
  box-shadow: 0 12px 34px rgba(31, 45, 61, .06);
}
.content-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}
h3 { margin: 0; color: #1e3a5f; font-size: 18px; }
.state-card {
  min-height: 420px;
  display: grid;
  place-items: center;
  border: 1px dashed #d8e1ee;
  border-radius: 8px;
  background: #f8fbff;
}
:deep(.el-alert) { margin-top: 12px; text-align: left; }
@media (max-width: 1024px) {
  .monitor-page { padding: 18px; }
  .monitor-shell { grid-template-columns: 1fr; }
}
@media (max-width: 640px) {
  .page-head { flex-direction: column; }
  .content-head { flex-direction: column; }
}
</style>
