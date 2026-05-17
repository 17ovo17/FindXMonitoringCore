import { create } from 'zustand'

export const useAppStore = create((set, get) => ({
  // 用户信息
  user: null,
  setUser: (user) => set({ user }),

  // 当前主题
  theme: localStorage.getItem('fx-theme') || 'dark',
  toggleTheme: () => {
    const next = get().theme === 'dark' ? 'light' : 'dark'
    localStorage.setItem('fx-theme', next)
    set({ theme: next })
  },

  // 未处理告警计数（导航红点用）
  alertCount: 0,
  setAlertCount: (n) => set({ alertCount: n }),

  // 导航折叠状态
  sidebarCollapsed: false,
  toggleSidebar: () => set(s => ({ sidebarCollapsed: !s.sidebarCollapsed })),

  // 全局 Toast
  toasts: [],
  addToast: (toast) => set(s => ({ toasts: [...s.toasts, { id: Date.now(), ...toast }] })),
  removeToast: (id) => set(s => ({ toasts: s.toasts.filter(t => t.id !== id) })),

  // AI 沙箱模式
  sandboxMode: 'auto_review', // readonly | auto_review | full_access
  setSandboxMode: (mode) => set({ sandboxMode: mode }),

  // 待确认的 AI 操作
  pendingApprovals: [],
  addApproval: (req) => set(s => ({ pendingApprovals: [...s.pendingApprovals, req] })),
  removeApproval: (id) => set(s => ({ pendingApprovals: s.pendingApprovals.filter(a => a.id !== id) })),
}))
