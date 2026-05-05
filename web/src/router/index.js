import { createRouter, createWebHistory } from 'vue-router'

const redirectWithSection = (path, section) => ({ path, query: { section } })

const monitorSectionRedirects = {
  overview: { path: '/dashboards', section: 'list' },
  datasource: { path: '/query', section: 'datasources' },
  datasources: { path: '/query', section: 'datasources' },
  query: { path: '/query', section: 'metrics' },
  metrics: { path: '/aiops', section: 'knowledge' },
  dashboard: { path: '/dashboards', section: 'list' },
  dashboards: { path: '/dashboards', section: 'list' },
  collection: { path: '/assets', section: 'agents' },
  templates: { path: '/dashboards', section: 'list' },
  agent: { path: '/assets', section: 'agents' },
  targets: { path: '/assets', section: 'hosts' },
  rules: { path: '/alerts', section: 'rules' },
  events: { path: '/alerts', section: 'events' },
}

const settingsTabRedirects = {
  credentials: { path: '/assets', section: 'agents' },
  datasource: { path: '/query', section: 'datasources' },
  metrics: { path: '/aiops', section: 'knowledge' },
  oncall: { path: '/org', section: 'teams' },
  'health-audit': { path: '/platform', section: 'health' },
  profiles: { path: '/org', section: 'users' },
}

const alertSectionRedirects = {
  notification: { path: '/notifications', section: 'channels' },
  notify: { path: '/notifications', section: 'channels' },
  oncall: { path: '/org', section: 'teams' },
  'recording-rules': { path: '/alerts', section: 'rules' },
  'aggr-views': { path: '/alerts', section: 'events' },
  subscribes: { path: '/notifications', section: 'rules' },
  pipeline: { path: '/alerts', section: 'rules' },
  mutes: { path: '/alerts', section: 'events' },
  audit: { path: '/platform', section: 'audit' },
}

const notificationSectionRedirects = {
  oncall: { path: '/org', section: 'teams' },
  pipeline: { path: '/alerts', section: 'rules' },
  configs: { path: '/notifications', section: 'rules' },
  templates: { path: '/notifications', section: 'rules' },
  records: { path: '/notifications', section: 'rules' },
}

const assetSectionRedirects = {
  installs: { path: '/assets', section: 'agents' },
  credentials: { path: '/assets', section: 'agents' },
}

const querySectionRedirects = {
  'metric-mapping': { path: '/aiops', section: 'knowledge' },
}

const dashboardSectionRedirects = {
  shares: { path: '/dashboards', section: 'list' },
  templates: { path: '/dashboards', section: 'list' },
}

const orgSectionRedirects = {
  permissions: { path: '/org', section: 'roles' },
  tokens: { path: '/org', section: 'users' },
}

const aiopsSectionRedirects = {
  'asset-chat': { path: '/aiops', section: 'diagnosis' },
}

const integrationSectionRedirects = {
  templates: { path: '/dashboards', section: 'list' },
  collection: { path: '/assets', section: 'agents' },
  labels: { path: '/aiops', section: 'knowledge' },
  system: { path: '/platform', section: 'settings' },
  cmdb: { path: '/assets', section: 'hosts' },
}

const normalizeLegacyRoute = to => {
  if (to.path === '/catpaw') return redirectWithSection('/assets', 'agents')
  if (to.path === '/settings/ai') return redirectWithSection('/platform', 'models')
  if (to.path === '/workbench') return redirectWithSection('/aiops', 'diagnosis')
  if (to.path === '/knowledge') return redirectWithSection('/aiops', 'knowledge')
  if (to.path === '/workflows') return redirectWithSection('/aiops', 'workflow')
  if (to.path === '/topology') return redirectWithSection('/assets', 'business')
  if (to.path === '/diagnose') return redirectWithSection('/aiops', 'knowledge')

  if (to.path === '/monitor') {
    const hit = monitorSectionRedirects[String(to.query.section || 'overview')] || monitorSectionRedirects.overview
    return redirectWithSection(hit.path, hit.section)
  }
  if (to.path === '/settings') {
    const hit = settingsTabRedirects[String(to.query.tab || '')]
    return hit ? redirectWithSection(hit.path, hit.section) : redirectWithSection('/platform', 'settings')
  }
  if (to.path === '/alerts') {
    const hit = alertSectionRedirects[String(to.query.section || '')]
    if (hit) return redirectWithSection(hit.path, hit.section)
  }
  if (to.path === '/query') {
    const hit = querySectionRedirects[String(to.query.section || '')]
    if (hit) return redirectWithSection(hit.path, hit.section)
  }
  if (to.path === '/dashboards') {
    const hit = dashboardSectionRedirects[String(to.query.section || '')]
    if (hit) return redirectWithSection(hit.path, hit.section)
  }
  if (to.path === '/assets') {
    const hit = assetSectionRedirects[String(to.query.section || '')]
    if (hit) return redirectWithSection(hit.path, hit.section)
  }
  if (to.path === '/notifications') {
    const hit = notificationSectionRedirects[String(to.query.section || '')]
    if (hit) return redirectWithSection(hit.path, hit.section)
  }
  if (to.path === '/org') {
    const hit = orgSectionRedirects[String(to.query.section || '')]
    if (hit) return redirectWithSection(hit.path, hit.section)
  }
  if (to.path === '/aiops') {
    const hit = aiopsSectionRedirects[String(to.query.section || '')]
    if (hit) return redirectWithSection(hit.path, hit.section)
  }
  if (to.path === '/integrations') {
    const hit = integrationSectionRedirects[String(to.query.section || 'templates')] || integrationSectionRedirects.templates
    return redirectWithSection(hit.path, hit.section)
  }
  return null
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('../views/Login.vue'), meta: { public: true } },
    { path: '/settings/change-password', component: () => import('../views/ChangePassword.vue') },
    { path: '/', redirect: '/assets?section=overview' },
    { path: '/assets', component: () => import('../views/AssetsWorkbench.vue') },
    { path: '/query', component: () => import('../views/QueryWorkbench.vue') },
    { path: '/dashboards', component: () => import('../views/DashboardsWorkbench.vue') },
    { path: '/alerts', component: () => import('../views/AlertingWorkbench.vue') },
    { path: '/notifications', component: () => import('../views/NotificationsWorkbench.vue') },
    { path: '/integrations', redirect: to => normalizeLegacyRoute(to) || redirectWithSection('/assets', 'agents') },
    { path: '/aiops', component: () => import('../views/AiopsWorkbench.vue') },
    { path: '/org', component: () => import('../views/OrgWorkbench.vue') },
    { path: '/platform', component: () => import('../views/PlatformWorkbench.vue') },
    { path: '/monitor', redirect: to => normalizeLegacyRoute(to) || redirectWithSection('/query', 'datasources') },
    { path: '/workbench', redirect: '/aiops?section=diagnosis' },
    { path: '/knowledge', redirect: '/aiops?section=knowledge' },
    { path: '/workflows', redirect: '/aiops?section=workflow' },
    { path: '/topology', redirect: '/assets?section=business' },
    { path: '/diagnose', redirect: '/aiops?section=knowledge' },
    { path: '/catpaw', redirect: '/assets?section=agents' },
    { path: '/settings/ai', redirect: '/platform?section=models' },
    { path: '/settings/profiles', redirect: '/org?section=users' },
    { path: '/settings/oncall', redirect: '/org?section=teams' },
    { path: '/settings', redirect: to => normalizeLegacyRoute(to) || redirectWithSection('/platform', 'settings') },
    { path: '/:pathMatch(.*)*', redirect: '/assets?section=overview' },
  ],
})

router.beforeEach((to, from, next) => {
  const legacy = normalizeLegacyRoute(to)
  if (legacy && (legacy.path !== to.path || legacy.query.section !== to.query.section)) return next(legacy)
  if (to.meta?.public) return next()
  const token = localStorage.getItem('aiw-token')
  if (!token) return next('/login')
  next()
})

export default router
