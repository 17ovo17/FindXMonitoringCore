import React, { Suspense, useCallback, useEffect, useMemo, useState } from 'react'
import { matchPath, useHistory, useLocation } from 'react-router-dom'
import { AuthBoundaryProvider, ThemeBoundaryProvider } from './contexts.jsx'
import { createNavigationRegistry, findNavByRoute, quickOptions as defaultQuickOptions } from './navigation.js'
import { AUTH_EXPIRED_EVENT, get, post, redactText } from './api/http.js'
import { normalizeLegacyRoute } from './legacyRoutes.js'
import { useI18n } from './i18n/useI18n.js'
import { ThemeToggle } from './shared/ThemeToggle.jsx'
import { FindXIcon } from './shared/FindXIcon.jsx'
import { ParticleBackground } from './shared/ParticleBackground.jsx'
import { ErrorBoundary } from './shared/ErrorBoundary.jsx'
import { CommandPalette } from './shared/CommandPalette.jsx'
import { ToastContainer } from './shared/Toast.jsx'
import { useAppStore } from './stores/useAppStore.js'
import './shared/fx-base.css'
import './styles.css'

// React.lazy 路由分割 — 减少初始 bundle 体积
const AiSrePage = React.lazy(() => import('./ai-sre/AiSrePage.jsx').then((m) => ({ default: m.AiSrePage })))
const AgentPage = React.lazy(() => import('./agents/AgentPage.jsx').then((m) => ({ default: m.AgentPage })))
const AssetsPage = React.lazy(() => import('./cmdb/AssetsPage.jsx').then((m) => ({ default: m.AssetsPage })))
const AlertsPage = React.lazy(() => import('./base-monitoring/alerts/AlertsPage.jsx').then((m) => ({ default: m.AlertsPage })))
const DashboardsPage = React.lazy(() => import('./base-monitoring/dashboards/DashboardsPage.jsx').then((m) => ({ default: m.DashboardsPage })))
const IntegrationsPage = React.lazy(() => import('./base-monitoring/integrations/IntegrationsPage.jsx').then((m) => ({ default: m.IntegrationsPage })))
const MetricExplorerPage = React.lazy(() => import('./base-monitoring/query/MetricExplorerPage.jsx').then((m) => ({ default: m.MetricExplorerPage })))
const NotificationsPage = React.lazy(() => import('./base-monitoring/notifications/NotificationsPage.jsx').then((m) => ({ default: m.NotificationsPage })))
const OverviewPage = React.lazy(() => import('./base-monitoring/OverviewPage.jsx').then((m) => ({ default: m.OverviewPage })))
const LogsPage = React.lazy(() => import('./logs/LogsPage.jsx').then((m) => ({ default: m.LogsPage })))
const OrgPage = React.lazy(() => import('./org/OrgPage.jsx').then((m) => ({ default: m.OrgPage })))
const PlatformPage = React.lazy(() => import('./platform/PlatformPage.jsx').then((m) => ({ default: m.PlatformPage })))
const BusinessProbePage = React.lazy(() => import('./probes/BusinessProbePage.jsx').then((m) => ({ default: m.BusinessProbePage })))
const TracingPage = React.lazy(() => import('./tracing/TracingPage.jsx').then((m) => ({ default: m.TracingPage })))
const NotFoundPage = React.lazy(() => import('./system/NotFoundPage.jsx').then((m) => ({ default: m.NotFoundPage })))

function LazyFallback() {
  return <div className="fx-lazy-loading">加载中...</div>
}

const readJson = (key, fallback = {}) => {
  try {
    return JSON.parse(localStorage.getItem(key) || '') || fallback
  } catch {
    return fallback
  }
}

const parseSearch = (search) => {
  const params = new URLSearchParams(search || '')
  return Object.fromEntries(Array.from(params.entries()))
}

const buildSearch = (query = {}) => {
  const params = new URLSearchParams()
  Object.entries(query).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') params.set(key, String(value))
  })
  const text = params.toString()
  return text ? `?${text}` : ''
}

const sameTarget = (location, target) => location.pathname === target.path && location.search === buildSearch(target.query)
const trimTrailingSlash = (path) => (path.length > 1 ? path.replace(/\/+$/, '') : path)
const navBranchKey = (group, item) => `${group.key}:${item.section}`

const hasActiveNavItem = (group, item, activeChild) => {
  if (!item || !activeChild) return false
  if (item.section === activeChild.section) return true
  return (item.children || []).some((child) => hasActiveNavItem(group, child, activeChild))
}

const firstNavigableTarget = (item) => {
  if (item?.to) return item.to
  for (const child of item?.children || []) {
    const target = firstNavigableTarget(child)
    if (target) return target
  }
  return null
}

const getRoute = (path) => {
  path = trimTrailingSlash(path)
  const traceMatch = matchPath(path, { path: '/tracing/:traceId', exact: true })
  if (traceMatch) return { name: 'tracing', params: traceMatch.params }
  if (path === '/assets') return { name: 'assets' }
  if (path === '/overview') return { name: 'overview' }
  if (path === '/query') return { name: 'query' }
  if (path === '/dashboards') return { name: 'dashboards' }
  if (path === '/alerts') return { name: 'alerts' }
  if (path === '/notifications') return { name: 'notifications' }
  if (path === '/integrations') return { name: 'integrations' }
  if (path === '/tracing') return { name: 'tracing', params: {} }
  if (path === '/logs') return { name: 'logs' }
  if (path === '/status') return { name: 'status' }
  if (path === '/agents') return { name: 'agents' }
  if (path === '/aiops') return { name: 'aiops' }
  if (path === '/org') return { name: 'org' }
  if (path === '/platform') return { name: 'platform' }
  return { name: 'not-found' }
}

function AuthCard({ title, children, error }) {
  return (
    <main className='fx-auth-page'>
      <section className='fx-auth-card'>
        <div className='fx-auth-logo'>FindX</div>
        <h1>{title}</h1>
        {children}
        {error && <p className='fx-auth-error'>{error}</p>}
      </section>
    </main>
  )
}

function LoginPage({ onLogin }) {
  const [form, setForm] = useState({ username: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const submit = async (event) => {
    event?.preventDefault()
    setError('')
    if (!form.username.trim() || !form.password) {
      setError('请输入用户名和密码')
      return
    }
    setLoading(true)
    try {
      const result = await post('/auth/login', form)
      onLogin(result.token, result.user)
    } catch (err) {
      setError(redactText(err.message || '登录失败'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <ParticleBackground />
      <AuthCard title='登录' error={error}>
        <form className='fx-auth-form' onSubmit={submit}>
          <label>用户名<input autoComplete='username' value={form.username} onChange={(event) => setForm({ ...form, username: event.target.value })} /></label>
          <label>密码<input autoComplete='current-password' type='password' value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} /></label>
          <button type='submit' disabled={loading}>{loading ? '登录中...' : '登录'}</button>
        </form>
      </AuthCard>
    </>
  )
}

function ChangePasswordForm({ user, onDone, onCancel, submitText = '确认修改', loadingText = '修改中...', showCancel = false }) {
  const [form, setForm] = useState({ old_password: '', new_password: '', confirm: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const submit = async (event) => {
    event?.preventDefault()
    setError('')
    if (form.new_password.length < 6) return setError('新密码至少 6 位')
    if (form.new_password !== form.confirm) return setError('两次密码不一致')
    setLoading(true)
    try {
      await post('/auth/change-password', { old_password: form.old_password, new_password: form.new_password })
      onDone?.({ ...user, must_change_pwd: false })
    } catch (err) {
      setError(redactText(err.message || '修改失败'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <form className='fx-auth-form fx-password-form' onSubmit={submit}>
        <label>当前密码<input autoComplete='current-password' type='password' value={form.old_password} onChange={(event) => setForm({ ...form, old_password: event.target.value })} /></label>
        <label>新密码<input autoComplete='new-password' type='password' value={form.new_password} onChange={(event) => setForm({ ...form, new_password: event.target.value })} /></label>
        <label>确认新密码<input autoComplete='new-password' type='password' value={form.confirm} onChange={(event) => setForm({ ...form, confirm: event.target.value })} /></label>
        <div className='fx-password-actions'>
          {showCancel && <button type='button' className='fx-password-cancel' disabled={loading} onClick={onCancel}>取消</button>}
          <button type='submit' className='fx-password-submit' disabled={loading}>{loading ? loadingText : submitText}</button>
        </div>
      </form>
      {error && <p className='fx-auth-error fx-password-error'>{error}</p>}
    </>
  )
}

function ChangePasswordPage({ user, onDone }) {
  return (
    <AuthCard title='修改密码'>
      <ChangePasswordForm user={user} onDone={onDone} />
    </AuthCard>
  )
}

function ChangePasswordModal({ user, open, onClose, onDone }) {
  if (!open) return null
  return (
    <div className='fx-confirm-overlay fx-password-modal' role='presentation' onMouseDown={(event) => { if (event.target === event.currentTarget) onClose?.() }}>
      <section className='fx-confirm-dialog fx-password-dialog' role='dialog' aria-modal='true' aria-labelledby='fx-password-title'>
        <header>
          <div>
            <h3 id='fx-password-title'>修改密码</h3>
            <p>更新当前登录账号的密码</p>
          </div>
          <button type='button' className='fx-confirm-close' aria-label='关闭修改密码弹窗' onClick={onClose}>×</button>
        </header>
        <ChangePasswordForm
          user={user}
          showCancel
          submitText='确认修改'
          loadingText='修改中...'
          onCancel={onClose}
          onDone={(nextUser) => {
            onDone?.(nextUser)
            onClose?.()
          }}
        />
      </section>
    </div>
  )
}

function LoadingPage() {
  return <main className='fx-auth-page'><section className='fx-auth-card'><div className='fx-auth-logo'>FindX</div><h1>正在校验登录状态</h1></section></main>
}

function renderPage(route, query, navigate, openTrace, openAgent) {
  const props = { query, onNavigate: navigate }
  if (route.name === 'overview') return <OverviewPage {...props} />
  if (route.name === 'assets') return <AssetsPage {...props} />
  if (route.name === 'query') return <MetricExplorerPage {...props} />
  if (route.name === 'dashboards') return <DashboardsPage {...props} />
  if (route.name === 'alerts') return <AlertsPage {...props} />
  if (route.name === 'notifications') return <NotificationsPage {...props} />
  if (route.name === 'integrations') return <IntegrationsPage {...props} />
  if (route.name === 'tracing') return <TracingPage {...props} params={route.params || {}} />
  if (route.name === 'logs') return <LogsPage {...props} onOpenTrace={openTrace} onOpenAgent={openAgent} />
  if (route.name === 'status') return <BusinessProbePage {...props} />
  if (route.name === 'agents') return <AgentPage {...props} />
  if (route.name === 'aiops') return <AiSrePage {...props} />
  if (route.name === 'org') return <OrgPage {...props} />
  if (route.name === 'platform') return <PlatformPage {...props} />
  return <NotFoundPage path={`${window.location.pathname}${window.location.search}`} onNavigate={navigate} />
}

export function FindXReactShell({ authBoundary, navigationItems, themeBoundary }) {
  const location = useLocation()
  const history = useHistory()
  const query = useMemo(() => parseSearch(location.search), [location.search])
  const route = useMemo(() => getRoute(location.pathname), [location.pathname])
  const navGroups = useMemo(() => createNavigationRegistry(navigationItems), [navigationItems])
  const activeNav = useMemo(() => findNavByRoute({ path: location.pathname, query }), [location.pathname, query])
  const [token, setToken] = useState(() => localStorage.getItem('aiw-token') || '')
  const [user, setUser] = useState(() => readJson('aiw-user'))
  const [authState, setAuthState] = useState(token ? 'checking' : 'anonymous')
  const [theme, setTheme] = useState(() => localStorage.getItem('aiw-theme') || 'light')
  const [openKeys, setOpenKeys] = useState(() => navGroups.map((item) => item.key))
  const [collapsed, setCollapsed] = useState(false)
  const [passwordModalOpen, setPasswordModalOpen] = useState(false)
  const [cmdPaletteOpen, setCmdPaletteOpen] = useState(false)
  const { lang, setLang, t } = useI18n()
  const appStoreTheme = useAppStore((s) => s.theme)
  const toggleStoreTheme = useAppStore((s) => s.toggleTheme)
  const setAlertCount = useAppStore((s) => s.setAlertCount)

  // 同步 store 主题到本地 state（store 作为 source of truth）
  useEffect(() => {
    setTheme(appStoreTheme)
  }, [appStoreTheme])

  // 告警计数轮询（每 30 秒）
  useEffect(() => {
    if (!token || authState !== 'ready') return
    let alive = true
    const poll = () => {
      get('/monitor/events/current?limit=0').then((res) => {
        if (alive && typeof res?.total === 'number') setAlertCount(res.total)
      }).catch(() => {})
    }
    poll()
    const timer = setInterval(poll, 30000)
    return () => { alive = false; clearInterval(timer) }
  }, [token, authState, setAlertCount])

  // Cmd+K / Ctrl+K 打开命令面板
  useEffect(() => {
    const handler = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setCmdPaletteOpen((v) => !v)
      }
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [])

  const clearAuth = useCallback(() => {
    localStorage.removeItem('aiw-token')
    localStorage.removeItem('aiw-user')
    setToken('')
    setUser({})
    setAuthState('anonymous')
  }, [])

  const navigate = useCallback((target = {}, mode = 'push') => {
    const nextPath = target.path || location.pathname
    const nextQuery = target.path ? (target.query || {}) : { ...query, ...target }
    if (nextPath === '/tracing' && nextQuery.traceId) {
      const traceId = nextQuery.traceId
      delete nextQuery.traceId
      nextQuery.section = 'trace-detail'
      return history[mode]({ pathname: `/tracing/${encodeURIComponent(traceId)}`, search: buildSearch(nextQuery) })
    }
    history[mode]({ pathname: nextPath, search: buildSearch(nextQuery) })
  }, [history, location.pathname, query])

  useEffect(() => {
    const legacy = normalizeLegacyRoute({ path: location.pathname, query })
    if (legacy && !sameTarget(location, legacy)) navigate(legacy, 'replace')
  }, [location, navigate, query])

  useEffect(() => {
    if (!token && location.pathname !== '/login') navigate({ path: '/login' }, 'replace')
  }, [token, location.pathname, navigate])

  useEffect(() => {
    if (token && authState === 'ready' && location.pathname === '/login') {
      navigate({ path: '/overview', query: { section: 'dashboard' } }, 'replace')
    }
  }, [token, authState, location.pathname, navigate])

  useEffect(() => {
    document.documentElement.dataset.theme = theme === 'dark' ? 'dark' : 'light'
    localStorage.setItem('aiw-theme', theme === 'dark' ? 'dark' : 'light')
  }, [theme])

  useEffect(() => {
    const handleExpired = () => clearAuth()
    window.addEventListener(AUTH_EXPIRED_EVENT, handleExpired)
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, handleExpired)
  }, [clearAuth])

  useEffect(() => {
    if (!token) return
    let alive = true
    setAuthState('checking')
    get('/auth/me')
      .then((me) => {
        if (!alive) return
        localStorage.setItem('aiw-user', JSON.stringify(me || {}))
        setUser(me || {})
        setAuthState('ready')
      })
      .catch(() => { if (alive) clearAuth() })
    return () => { alive = false }
  }, [token, clearAuth])

  useEffect(() => {
    if (!openKeys.includes(activeNav.group.key)) setOpenKeys((keys) => [...keys, activeNav.group.key])
  }, [activeNav.group.key, openKeys])

  const onLogin = (nextToken, nextUser) => {
    localStorage.setItem('aiw-token', nextToken)
    localStorage.setItem('aiw-user', JSON.stringify(nextUser || {}))
    setToken(nextToken)
    setUser(nextUser || {})
    setAuthState('ready')
    navigate(nextUser?.must_change_pwd ? { path: '/settings/change-password' } : { path: '/overview', query: { section: 'dashboard' } }, 'replace')
  }

  const onPasswordDone = (nextUser) => {
    localStorage.setItem('aiw-user', JSON.stringify(nextUser || {}))
    setUser(nextUser || {})
    navigate({ path: '/overview', query: { section: 'dashboard' } }, 'replace')
  }

  const onPasswordModalDone = (nextUser) => {
    localStorage.setItem('aiw-user', JSON.stringify(nextUser || {}))
    setUser(nextUser || {})
  }

  const logout = async () => {
    try { await post('/auth/logout', {}) } catch {}
    clearAuth()
    navigate({ path: '/login' }, 'replace')
  }

  const openTrace = (traceId) => navigate({ path: `/tracing/${encodeURIComponent(traceId)}`, query: { section: 'trace-detail' } })
  const openAgent = ({ q = '', packageName = '' } = {}) => navigate({ path: '/agents', query: { section: 'hosts', package: packageName, q } })
  const toggleOpenKey = (key) => setOpenKeys((keys) => keys.includes(key) ? keys.filter((item) => item !== key) : [...keys, key])
  const renderNavItems = (group, items, level = 1) => (
    <div className={`fx-nav-children fx-nav-level-${level}`}>
      {items.map((item) => {
        const branchKey = navBranchKey(group, item)
        const hasChildren = Boolean(item.children?.length)
        const active = activeNav.group.key === group.key && hasActiveNavItem(group, item, activeNav.child)
        const expanded = openKeys.includes(branchKey) || active
        const target = firstNavigableTarget(item)
        return (
          <div key={`${group.key}-${item.section}`} className={`fx-nav-branch ${active ? 'is-active-branch' : ''}`}>
            <button
              type='button'
              className={`fx-nav-item ${hasChildren ? 'is-parent' : ''} ${active && !hasChildren ? 'is-active' : ''}`}
              data-level={level}
              onClick={() => {
                if (hasChildren) toggleOpenKey(branchKey)
                if (target && (!hasChildren || !expanded)) navigate(target)
              }}
            >
              <span>{item.label}</span>
              {hasChildren && <small>{expanded ? '⌃' : '⌄'}</small>}
            </button>
            {hasChildren && expanded && renderNavItems(group, item.children, level + 1)}
          </div>
        )
      })}
    </div>
  )

  if (!token && location.pathname !== '/login') return <LoadingPage />
  if (token && authState === 'checking') return <LoadingPage />
  if (token && authState === 'ready' && location.pathname === '/login') return <LoadingPage />
  if (location.pathname === '/login') return <LoginPage onLogin={onLogin} />
  if (location.pathname === '/settings/change-password') return <ChangePasswordPage user={user} onDone={onPasswordDone} />

  const boundary = { ...authBoundary, user, auditMode: 'findx-audit', permissionMode: 'findx-permission' }
  const content = renderPage(route, query, navigate, openTrace, openAgent)
  const userInitial = String(user?.username || '用').slice(0, 1).toUpperCase()

  return (
    <AuthBoundaryProvider value={boundary}>
      <ThemeBoundaryProvider value={{ ...themeBoundary, mode: theme }}>
        <div className={`fx-react-app ${collapsed ? 'is-collapsed' : ''}`}>
          <aside className='fx-sidebar' aria-label='FindX 主导航'>
            <div className='fx-brand'>
              <span className='fx-brand-mark'>FX</span>
              <span className='fx-brand-copy'><strong>FindX</strong><small>智能运维平台</small></span>
              <button type='button' title={collapsed ? '展开侧边栏' : '收起侧边栏'} onClick={() => setCollapsed((value) => !value)}>{collapsed ? '>' : '<'}</button>
            </div>
            <label className='fx-quick'><span>快捷跳转</span><select value='' onChange={(event) => { const hit = defaultQuickOptions.find((item) => item.value === event.target.value); if (hit) navigate(hit.to) }}><option value=''>全局搜索 / 快速跳转</option>{defaultQuickOptions.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select></label>
            <nav className='fx-nav'>
              {navGroups.map((group) => (
                <section key={group.key} className='fx-nav-section'>
                  <button type='button' className={`fx-nav-parent ${activeNav.group.key === group.key ? 'is-active' : ''}`} onClick={() => { setOpenKeys((keys) => keys.includes(group.key) ? keys.filter((key) => key !== group.key) : [...keys, group.key]); navigate(group.to) }}>
                    <span className='fx-nav-main'>
                      <span className='fx-nav-icon' aria-hidden='true'><FindXIcon name={group.icon || group.key} /></span>
                      <span className='fx-nav-label'>{group.label}</span>
                    </span>
                    <small>{openKeys.includes(group.key) ? '⌃' : '⌄'}</small>
                  </button>
                  {openKeys.includes(group.key) && <div className='fx-subnav'>{renderNavItems(group, group.children)}</div>}
                </section>
              ))}
            </nav>
          </aside>
          <section className='fx-workspace'>
            <header className='fx-workspace-header'>
              <div><p>{route.name === 'not-found' ? 'FindX' : activeNav.group.label}</p><h1>{route.name === 'not-found' ? '页面不存在' : activeNav.child.label}</h1></div>
              <div className='fx-header-actions'>
                <button type='button' className='fx-lang-toggle' onClick={() => setLang(lang === 'zh-CN' ? 'en-US' : 'zh-CN')}>{lang === 'zh-CN' ? 'EN' : '中'}</button>
                <ThemeToggle theme={theme} onToggle={(next) => { setTheme(next); toggleStoreTheme() }} />
                <button type='button' onClick={() => setPasswordModalOpen(true)}>改密</button>
                <button type='button' className='fx-user' onClick={logout}><span>{userInitial}</span>{user?.username || '用户'} / 退出</button>
              </div>
            </header>
            <main className='fx-content'>
              <ErrorBoundary>
                <Suspense fallback={<LazyFallback />}>
                  {content}
                </Suspense>
              </ErrorBoundary>
            </main>
          </section>
          <ChangePasswordModal user={user} open={passwordModalOpen} onClose={() => setPasswordModalOpen(false)} onDone={onPasswordModalDone} />
          <CommandPalette open={cmdPaletteOpen} onClose={() => setCmdPaletteOpen(false)} onNavigate={navigate} />
          <ToastContainer />
        </div>
      </ThemeBoundaryProvider>
    </AuthBoundaryProvider>
  )
}
