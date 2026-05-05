<template>
  <div class="app-shell" :class="{ 'auth-shell': isAuthPage }">
    <template v-if="isAuthPage">
      <router-view />
    </template>
    <template v-else>
      <aside class="platform-sidebar" :class="{ collapsed: isSidebarCollapsed }">
        <div class="brand-block">
          <div class="brand-mark">FX</div>
          <div class="brand-copy">
            <div class="brand-name">FindX</div>
            <div class="brand-subtitle">智能运维平台</div>
          </div>
          <button
            class="sidebar-toggle"
            type="button"
            :title="isSidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
            :aria-label="isSidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
            @click="toggleSidebar"
          >
            <el-icon><component :is="isSidebarCollapsed ? 'Expand' : 'Fold'" /></el-icon>
          </button>
        </div>

        <div class="quick-jump">
          <div class="quick-title">快捷跳转</div>
          <el-select
            v-model="quickValue"
            filterable
            clearable
            placeholder="全局搜索 / 快速跳转"
            @change="goQuick"
          >
            <el-option v-for="item in quickOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </div>

        <nav class="menu-scroll" aria-label="FindX 主导航">
          <section v-for="group in navGroups" :key="group.key" class="menu-section">
            <router-link
              :to="group.to"
              class="menu-parent"
              :class="{ active: isGroupActive(group) }"
              :title="isSidebarCollapsed ? group.label : ''"
              @click="toggleGroup($event, group)"
            >
              <el-icon><component :is="group.icon" /></el-icon>
              <span class="menu-label">{{ group.label }}</span>
              <el-icon class="chevron" :class="{ open: openKeys.includes(group.key) }"><ArrowDown /></el-icon>
            </router-link>
            <div v-show="!isSidebarCollapsed && openKeys.includes(group.key)" class="submenu">
              <router-link
                v-for="item in group.children"
                :key="`${group.key}-${item.section}`"
                :to="{ path: group.path, query: { section: item.section } }"
                class="submenu-item"
                :class="{ active: isChildActive(group, item.section) }"
              >
                {{ item.label }}
              </router-link>
            </div>
          </section>
        </nav>
      </aside>

      <section class="platform-workspace">
        <header class="workspace-header">
          <div class="workspace-title">
            <div class="workspace-kicker">{{ activeGroup.label }}</div>
            <h1>{{ activeChild.label }}</h1>
          </div>
          <div class="workspace-actions">
            <button class="icon-button" type="button" :class="{ active: theme === 'light' }" title="日间模式" @click="setTheme('light')">
              <el-icon><Sunny /></el-icon>
            </button>
            <button class="icon-button" type="button" :class="{ active: theme === 'dark' }" title="夜间模式" @click="setTheme('dark')">
              <el-icon><Moon /></el-icon>
            </button>
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
import { navGroups, quickOptions, findNavByRoute } from './router/nav'

const route = useRoute()
const router = useRouter()
const theme = ref(localStorage.getItem('aiw-theme') || 'light')
const quickValue = ref('')
const openKeys = ref(navGroups.map(group => group.key))
const isSidebarCollapsed = ref(false)
const isAuthPage = computed(() => route.path === '/login' || route.path === '/settings/change-password')

const activeNav = computed(() => findNavByRoute(route))
const activeGroup = computed(() => activeNav.value.group)
const activeChild = computed(() => activeNav.value.child)
const currentUser = computed(() => { try { return JSON.parse(localStorage.getItem('aiw-user') || '{}') } catch { return {} } })
const userInitial = computed(() => String(currentUser.value.username || '用').slice(0, 1).toUpperCase())

const isGroupActive = group => route.path === group.path || route.path.startsWith(`${group.path}/`)
const isChildActive = (group, section) => isGroupActive(group) && String(route.query.section || group.defaultSection) === section

const toggleGroup = (event, group) => {
  if (isSidebarCollapsed.value) return
  event?.preventDefault()
  if (openKeys.value.includes(group.key)) {
    openKeys.value = openKeys.value.filter(key => key !== group.key)
  } else {
    openKeys.value = [...openKeys.value, group.key]
  }
  router.push(group.to)
}

const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

const handleLogout = () => {
  localStorage.removeItem('aiw-token')
  localStorage.removeItem('aiw-user')
  router.push('/login')
}

const handleUserCommand = command => {
  if (command === 'logout') handleLogout()
  if (command === 'change-password') router.push('/settings/change-password')
}

const setTheme = mode => { theme.value = mode }
const goQuick = value => {
  const hit = quickOptions.find(item => item.value === value)
  if (hit) router.push(hit.to)
  quickValue.value = ''
}

watch(theme, value => {
  const mode = value === 'dark' ? 'dark' : 'light'
  document.documentElement.dataset.theme = mode
  localStorage.setItem('aiw-theme', mode)
}, { immediate: true })

watch(() => activeGroup.value.key, key => {
  if (key && !openKeys.value.includes(key)) openKeys.value = [...openKeys.value, key]
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
  width: 248px; flex: 0 0 248px; display: flex; flex-direction: column; min-height: 0;
  background: var(--fx-sidebar); border-right: 1px solid var(--fx-border);
  transition: width .18s ease, flex-basis .18s ease;
}
.platform-sidebar.collapsed {
  width: 78px; flex-basis: 78px;
}
.brand-block {
  height: 66px; display: flex; align-items: center; gap: 12px; padding: 0 18px;
  border-bottom: 1px solid var(--fx-border);
}
.brand-mark {
  width: 34px; height: 34px; display: grid; place-items: center; border-radius: 8px; color: #fff;
  font-size: 13px; font-weight: 800; background: linear-gradient(135deg, #1769ff, #25b8ff);
  box-shadow: 0 10px 22px rgba(23, 105, 255, .22);
}
.brand-name { color: var(--fx-ink); font-size: 17px; font-weight: 800; line-height: 1; }
.brand-subtitle { margin-top: 5px; color: var(--fx-muted); font-size: 12px; }
.sidebar-toggle {
  width: 30px; height: 30px; display: grid; place-items: center; flex: 0 0 auto; margin-left: auto;
  border: 1px solid rgba(23, 105, 255, .2); border-radius: 8px; color: var(--fx-blue);
  background: var(--fx-blue-soft); cursor: pointer;
}
.sidebar-toggle:hover { border-color: rgba(23, 105, 255, .34); background: rgba(23, 105, 255, .12); }
.platform-sidebar.collapsed .brand-block {
  height: 82px; flex-direction: column; justify-content: center; gap: 8px; padding: 8px 0;
}
.platform-sidebar.collapsed .brand-copy,
.platform-sidebar.collapsed .quick-jump,
.platform-sidebar.collapsed .menu-label,
.platform-sidebar.collapsed .submenu,
.platform-sidebar.collapsed .chevron { display: none; }
.platform-sidebar.collapsed .sidebar-toggle { margin-left: 0; }
.quick-jump { padding: 10px 10px 8px; border-bottom: 1px solid var(--fx-border); }
.quick-title { margin: 0 0 6px 2px; color: var(--fx-muted); font-size: 12px; font-weight: 800; }
.quick-jump .el-select__wrapper {
  min-height: 34px; border-radius: 8px !important; background: var(--fx-panel);
  box-shadow: 0 0 0 1px var(--fx-border) inset;
}
.menu-scroll { flex: 1; min-height: 0; overflow-y: auto; padding: 12px 10px; }
.menu-section + .menu-section { margin-top: 5px; }
.menu-parent {
  height: 40px; display: flex; align-items: center; gap: 10px; padding: 0 10px 0 12px; border-radius: 8px;
  color: var(--fx-ink); text-decoration: none; font-size: 14px; font-weight: 800;
}
.platform-sidebar.collapsed .menu-scroll { padding: 12px 10px; }
.platform-sidebar.collapsed .menu-parent { justify-content: center; padding: 0; }
.menu-parent .el-icon { width: 18px; font-size: 18px; color: #7890b2; }
.menu-parent:hover, .menu-parent.active { color: var(--fx-blue); background: var(--fx-blue-soft); }
.menu-parent.active .el-icon { color: var(--fx-blue); }
.chevron { margin-left: auto; transition: transform .18s ease; }
.chevron.open { transform: rotate(180deg); }
.submenu { padding: 3px 0 8px 40px; }
.submenu-item {
  height: 30px; display: flex; align-items: center; padding: 0 10px; border-radius: 7px;
  color: var(--fx-muted); text-decoration: none; font-size: 13px; font-weight: 700;
}
.submenu-item:hover { color: var(--fx-blue); background: rgba(23, 105, 255, .06); }
.submenu-item.active { color: var(--fx-blue); background: rgba(23, 105, 255, .09); }
.platform-workspace { flex: 1; min-width: 0; display: flex; flex-direction: column; height: 100vh; overflow: hidden; }
.workspace-header {
  height: 66px; flex: 0 0 66px; display: flex; align-items: center; justify-content: space-between;
  gap: 16px; padding: 0 24px; border-bottom: 1px solid var(--fx-border);
  background: rgba(255, 255, 255, .78); backdrop-filter: blur(14px);
}
[data-theme="dark"] .workspace-header { background: rgba(18, 27, 43, .78); }
.workspace-title { min-width: 0; }
.workspace-kicker { color: var(--fx-muted); font-size: 12px; font-weight: 800; }
.workspace-header h1 { margin-top: 3px; color: var(--fx-ink); font-size: 18px; line-height: 1.2; }
.workspace-actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; justify-content: flex-end; }
.icon-button, .user-entry {
  height: 32px; border: 1px solid var(--fx-border); border-radius: 8px; color: var(--fx-muted);
  background: var(--fx-panel); cursor: pointer;
}
.icon-button { width: 32px; display: grid; place-items: center; font-size: 16px; }
.icon-button.active { color: var(--fx-blue); background: var(--fx-blue-soft); border-color: rgba(23, 105, 255, .28); }
.user-entry { display: flex; align-items: center; gap: 8px; padding: 0 9px; }
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
  background: linear-gradient(180deg, rgba(255,255,255,.72), rgba(255,255,255,0) 190px), var(--fx-bg);
}
[data-theme="dark"] .workspace-content {
  background: linear-gradient(180deg, rgba(29,42,66,.42), rgba(29,42,66,0) 190px), var(--fx-bg);
}
.workspace-content::-webkit-scrollbar { width: 10px; height: 10px; }
.workspace-content::-webkit-scrollbar-thumb { background: rgba(87, 105, 132, .28); border-radius: 999px; }
.el-button { border-radius: 8px; font-weight: 700; }
.el-button--primary { background: #1769ff; border-color: #1769ff; }
.el-input__wrapper, .el-textarea__inner, .el-select__wrapper { border-radius: 8px !important; }
@media (max-width: 1024px) {
  .platform-sidebar { width: 78px; flex-basis: 78px; }
  .brand-block { height: 82px; flex-direction: column; justify-content: center; gap: 8px; padding: 8px 0; }
  .brand-copy, .quick-jump, .menu-label, .submenu, .chevron { display: none; }
  .sidebar-toggle { margin-left: 0; }
  .menu-parent { justify-content: center; padding: 0; }
  .workspace-header { padding: 0 16px; }
}
@media (max-width: 680px) {
  .workspace-title h1 { font-size: 16px; }
  .user-name { display: none; }
}
</style>
