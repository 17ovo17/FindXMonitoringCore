<template>
  <div class="app-shell" :class="{ 'auth-shell': isAuthPage }">
    <template v-if="isAuthPage">
      <router-view />
    </template>
    <template v-else>
      <aside class="platform-sidebar">
        <div class="brand-block">
          <div class="brand-mark">FX</div>
          <div class="brand-copy">
            <div class="brand-name">FindX</div>
            <div class="brand-subtitle">智能运维平台</div>
          </div>
        </div>
        <nav class="menu-scroll" aria-label="FindX 主导航">
          <section v-for="group in navGroups" :key="group.label" class="menu-section">
            <div class="menu-title">{{ group.label }}</div>
            <router-link
              v-for="item in group.items"
              :key="item.to"
              :to="item.to"
              class="menu-item"
              :class="{ active: isActive(item.to) }"
            >
              <el-icon><component :is="item.icon" /></el-icon>
              <span>{{ item.label }}</span>
            </router-link>
          </section>
        </nav>
      </aside>

      <section class="platform-workspace">
        <header class="workspace-header">
          <div class="workspace-title">
            <div class="workspace-kicker">{{ activeGroup }}</div>
            <h1>{{ activeTitle }}</h1>
          </div>
          <div class="workspace-actions">
            <router-link class="quick-link" to="/monitor">监控运维</router-link>
            <router-link class="quick-link" to="/workbench">AI 问诊</router-link>
            <router-link class="quick-link" to="/settings">平台设置</router-link>
            <div class="theme-switch" role="group" aria-label="主题切换">
              <button type="button" :class="{ active: theme === 'light' }" aria-label="切换到日间模式" @click="setTheme('light')"><el-icon><Sunny /></el-icon></button>
              <button type="button" :class="{ active: theme === 'dark' }" aria-label="切换到夜间模式" @click="setTheme('dark')"><el-icon><Moon /></el-icon></button>
            </div>
            <el-dropdown trigger="click" @command="handleUserCommand">
              <button class="user-entry" type="button">
                <span class="user-avatar">{{ userInitial }}</span>
                <span class="user-name">{{ currentUser.username || '用户' }}</span>
                <el-icon><ArrowDown /></el-icon>
              </button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="change-password"><el-icon><Key /></el-icon>修改密码</el-dropdown-item>
                  <el-dropdown-item command="logout" divided><el-icon><SwitchButton /></el-icon>退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </header>
        <main class="workspace-content"><router-view /></main>
      </section>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const theme = ref(localStorage.getItem('aiw-theme') || 'light')
const isAuthPage = computed(() => route.path === '/login' || route.path === '/settings/change-password')

const navGroups = [
  {
    label: '运行态势',
    items: [
      { to: '/', label: '运维总览', icon: 'DataBoard' },
      { to: '/monitor', label: '监控运维', icon: 'Monitor' },
      { to: '/alerts', label: '告警中心', icon: 'Bell' },
      { to: '/topology', label: '业务拓扑', icon: 'Share' },
    ],
  },
  {
    label: '智能协同',
    items: [
      { to: '/workbench', label: 'AI 问诊', icon: 'ChatDotRound' },
      { to: '/knowledge', label: '知识库', icon: 'Collection' },
      { to: '/workflows', label: '工作流自愈', icon: 'Operation' },
    ],
  },
  {
    label: '平台管理',
    items: [
      { to: '/settings/ai', label: '模型配置', icon: 'MagicStick' },
      { to: '/settings', label: '系统设置', icon: 'Setting' },
    ],
  },
]

const flatNav = computed(() => navGroups.flatMap(group => group.items.map(item => ({ ...item, group: group.label }))))
const activeItem = computed(() => flatNav.value.find(item => isActive(item.to)) || flatNav.value[0])
const activeGroup = computed(() => activeItem.value?.group || 'FindX')
const activeTitle = computed(() => activeItem.value?.label || 'FindX')
const currentUser = computed(() => {
  try { return JSON.parse(localStorage.getItem('aiw-user') || '{}') } catch { return {} }
})
const userInitial = computed(() => String(currentUser.value.username || '用').slice(0, 1).toUpperCase())
const isActive = to => to === '/' ? route.path === '/' : route.path === to || route.path.startsWith(`${to}/`)

const handleLogout = () => {
  localStorage.removeItem('aiw-token')
  localStorage.removeItem('aiw-user')
  router.push('/login')
}
const handleUserCommand = cmd => {
  if (cmd === 'logout') handleLogout()
  if (cmd === 'change-password') router.push('/settings/change-password')
}
const setTheme = mode => { theme.value = mode }

watch(theme, value => {
  const mode = value === 'dark' ? 'dark' : 'light'
  document.documentElement.dataset.theme = mode
  localStorage.setItem('aiw-theme', mode)
}, { immediate: true })
onMounted(() => setTheme(theme.value))
</script>

<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
:root {
  color-scheme: light; --fx-ink: #17233d; --fx-muted: #69758c; --fx-blue: #1769ff;
  --fx-blue-soft: #e8f1ff; --fx-bg: #f3f6fb; --fx-panel: #fff; --fx-border: #e3e8f2;
  --fx-sidebar: #f8fbff; --fx-shadow: 0 16px 38px rgba(20, 35, 70, .08);
}
[data-theme="dark"] {
  color-scheme: dark; --fx-ink: #e8eef9; --fx-muted: #9aa8bd; --fx-blue: #6aa7ff;
  --fx-blue-soft: rgba(79, 139, 255, .18); --fx-bg: #0b1220; --fx-panel: #121b2b;
  --fx-border: rgba(148, 163, 184, .18); --fx-sidebar: #101928;
  --fx-shadow: 0 20px 48px rgba(0, 0, 0, .28);
}
html, body, #app { width: 100%; height: 100%; min-height: 100%; }
body {
  min-width: 0; overflow: hidden; background: var(--fx-bg); color: var(--fx-ink);
  font-family: "Microsoft YaHei", "PingFang SC", "Noto Sans SC", ui-sans-serif, sans-serif;
}
.app-shell { display: flex; width: 100%; height: 100vh; min-width: 0; overflow: hidden; background: var(--fx-bg); }
.app-shell.auth-shell { align-items: stretch; justify-content: center; }
.platform-sidebar {
  width: 224px; flex: 0 0 224px; display: flex; flex-direction: column; min-height: 0;
  background: var(--fx-sidebar); border-right: 1px solid var(--fx-border);
}
.brand-block {
  height: 66px; display: flex; align-items: center; gap: 12px; padding: 0 18px;
  border-bottom: 1px solid var(--fx-border);
}
.brand-mark {
  width: 34px; height: 34px; display: grid; place-items: center; border-radius: 9px; color: #fff;
  font-size: 13px; font-weight: 800; background: linear-gradient(135deg, #1769ff, #25b8ff);
  box-shadow: 0 10px 22px rgba(23, 105, 255, .22);
}
.brand-name { color: var(--fx-ink); font-size: 17px; font-weight: 800; line-height: 1; }
.brand-subtitle { margin-top: 5px; color: var(--fx-muted); font-size: 12px; }
.menu-scroll { flex: 1; min-height: 0; overflow-y: auto; padding: 14px 10px; }
.menu-section + .menu-section { margin-top: 16px; }
.menu-title { padding: 8px 10px; color: var(--fx-muted); font-size: 12px; font-weight: 800; }
.menu-item {
  height: 38px; display: flex; align-items: center; gap: 10px; padding: 0 12px; border-radius: 8px;
  color: var(--fx-ink); text-decoration: none; font-size: 14px; font-weight: 700;
}
.menu-item .el-icon { width: 18px; font-size: 18px; color: #7890b2; }
.menu-item:hover, .menu-item.active, .menu-item.router-link-active { color: var(--fx-blue); background: var(--fx-blue-soft); }
.menu-item.active .el-icon, .menu-item.router-link-active .el-icon { color: var(--fx-blue); }
.platform-workspace {
  flex: 1; min-width: 0; display: flex; flex-direction: column; height: 100vh; overflow: hidden;
}
.workspace-header {
  height: 66px; flex: 0 0 66px; display: flex; align-items: center; justify-content: space-between;
  gap: 16px; padding: 0 24px; border-bottom: 1px solid var(--fx-border);
  background: rgba(255, 255, 255, .78); backdrop-filter: blur(14px);
}
[data-theme="dark"] .workspace-header { background: rgba(18, 27, 43, .78); }
.workspace-title { min-width: 0; }
.workspace-kicker { color: var(--fx-muted); font-size: 12px; font-weight: 800; }
.workspace-header h1 { margin-top: 3px; color: var(--fx-ink); font-size: 18px; line-height: 1.2; }
.workspace-actions {
  display: flex; align-items: center; gap: 8px; flex-wrap: wrap; justify-content: flex-end;
}
.quick-link {
  height: 30px; display: inline-flex; align-items: center; padding: 0 12px; border: 1px solid var(--fx-border);
  border-radius: 8px; color: var(--fx-ink); background: var(--fx-panel); text-decoration: none;
  font-size: 12px; font-weight: 700;
}
.quick-link.router-link-active { color: var(--fx-blue); border-color: rgba(23, 105, 255, .32); background: var(--fx-blue-soft); }
.theme-switch {
  display: flex; gap: 4px; padding: 3px; border: 1px solid var(--fx-border);
  border-radius: 8px; background: var(--fx-panel);
}
.theme-switch button, .user-entry { border: 0; border-radius: 6px; color: var(--fx-muted); background: transparent; cursor: pointer; }
.theme-switch button { width: 26px; height: 24px; display: grid; place-items: center; font-size: 15px; }
.theme-switch button.active { color: var(--fx-blue); background: var(--fx-blue-soft); }
.user-entry {
  height: 32px; display: flex; align-items: center; gap: 8px; padding: 0 9px;
  border: 1px solid var(--fx-border); background: var(--fx-panel);
}
.user-avatar {
  width: 23px; height: 23px; display: grid; place-items: center; border-radius: 50%;
  color: #fff; background: #1769ff; font-size: 12px; font-weight: 800;
}
.user-name {
  max-width: 96px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  color: var(--fx-ink); font-size: 12px; font-weight: 700;
}
.workspace-content {
  flex: 1; min-width: 0; min-height: 0; overflow: auto;
  background: linear-gradient(180deg, rgba(255, 255, 255, .72), rgba(255, 255, 255, 0) 190px), var(--fx-bg);
}
[data-theme="dark"] .workspace-content {
  background: linear-gradient(180deg, rgba(29, 42, 66, .42), rgba(29, 42, 66, 0) 190px), var(--fx-bg);
}
.workspace-content::-webkit-scrollbar { width: 10px; height: 10px; }
.workspace-content::-webkit-scrollbar-thumb { background: rgba(87, 105, 132, .28); border-radius: 999px; }
.el-button { border-radius: 8px; font-weight: 700; }
.el-button--primary { background: #1769ff; border-color: #1769ff; }
.el-input__wrapper, .el-textarea__inner, .el-select__wrapper { border-radius: 8px !important; }
@media (max-width: 1024px) {
  .platform-sidebar { width: 76px; flex-basis: 76px; }
  .brand-block { justify-content: center; padding: 0; }
  .brand-copy, .menu-title, .menu-item span, .quick-link { display: none; }
  .menu-item { justify-content: center; padding: 0; }
  .workspace-header { padding: 0 16px; }
}
@media (max-width: 680px) {
  .workspace-actions { gap: 6px; }
  .workspace-title h1 { font-size: 16px; }
  .user-name { display: none; }
}
</style>
