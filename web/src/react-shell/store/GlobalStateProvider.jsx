import React, { createContext, useCallback, useContext, useMemo, useReducer } from 'react'

/**
 * @typedef {Object} GlobalState
 * @property {Object|null} currentUser - 当前登录用户
 * @property {'light'|'dark'} theme - 主题模式
 * @property {number} unreadAlerts - 未读告警数
 * @property {Array} notifications - 通知列表
 */

const initialState = {
  currentUser: null,
  theme: 'light',
  unreadAlerts: 0,
  notifications: [],
}

function globalReducer(state, action) {
  switch (action.type) {
    case 'SET_USER':
      return { ...state, currentUser: action.payload }
    case 'SET_THEME':
      return { ...state, theme: action.payload }
    case 'SET_UNREAD_ALERTS':
      return { ...state, unreadAlerts: action.payload }
    case 'ADD_NOTIFICATION':
      return { ...state, notifications: [action.payload, ...state.notifications].slice(0, 50) }
    case 'CLEAR_NOTIFICATIONS':
      return { ...state, notifications: [] }
    case 'MARK_NOTIFICATION_READ': {
      const notifications = state.notifications.map((n) =>
        n.id === action.payload ? { ...n, read: true } : n,
      )
      return { ...state, notifications }
    }
    default:
      return state
  }
}

const GlobalStateContext = createContext(null)
const GlobalDispatchContext = createContext(null)

/**
 * GlobalStateProvider — 全局状态容器，使用 React Context + useReducer。
 * 包裹应用根组件，提供 currentUser、theme、alerts、notifications 等全局状态。
 */
export function GlobalStateProvider({ children, initialUser, initialTheme }) {
  const [state, dispatch] = useReducer(globalReducer, {
    ...initialState,
    currentUser: initialUser || null,
    theme: initialTheme || 'light',
  })

  return (
    <GlobalStateContext.Provider value={state}>
      <GlobalDispatchContext.Provider value={dispatch}>
        {children}
      </GlobalDispatchContext.Provider>
    </GlobalStateContext.Provider>
  )
}

/**
 * useGlobalState — 获取全局状态。
 * @returns {GlobalState}
 */
export function useGlobalState() {
  const context = useContext(GlobalStateContext)
  if (context === null) {
    throw new Error('useGlobalState must be used within a GlobalStateProvider')
  }
  return context
}

/**
 * useGlobalDispatch — 获取全局 dispatch 函数。
 */
export function useGlobalDispatch() {
  const context = useContext(GlobalDispatchContext)
  if (context === null) {
    throw new Error('useGlobalDispatch must be used within a GlobalStateProvider')
  }
  return context
}

/**
 * useGlobalActions — 封装常用 dispatch 操作为便捷函数。
 */
export function useGlobalActions() {
  const dispatch = useGlobalDispatch()

  const setUser = useCallback((user) => dispatch({ type: 'SET_USER', payload: user }), [dispatch])
  const setTheme = useCallback((theme) => dispatch({ type: 'SET_THEME', payload: theme }), [dispatch])
  const setUnreadAlerts = useCallback((count) => dispatch({ type: 'SET_UNREAD_ALERTS', payload: count }), [dispatch])
  const addNotification = useCallback((notification) => dispatch({ type: 'ADD_NOTIFICATION', payload: notification }), [dispatch])
  const clearNotifications = useCallback(() => dispatch({ type: 'CLEAR_NOTIFICATIONS' }), [dispatch])
  const markNotificationRead = useCallback((id) => dispatch({ type: 'MARK_NOTIFICATION_READ', payload: id }), [dispatch])

  return useMemo(() => ({
    setUser,
    setTheme,
    setUnreadAlerts,
    addNotification,
    clearNotifications,
    markNotificationRead,
  }), [setUser, setTheme, setUnreadAlerts, addNotification, clearNotifications, markNotificationRead])
}
