import React from 'react'

const STORAGE_KEY = 'aiw-theme'

/**
 * 主题切换按钮组件
 * 切换时给 <html> 添加 data-theme="dark" 属性
 * 存储偏好到 localStorage
 */
export function ThemeToggle({ theme, onToggle }) {
  const isDark = theme === 'dark'
  const toggle = () => {
    const next = isDark ? 'light' : 'dark'
    document.documentElement.dataset.theme = next
    localStorage.setItem(STORAGE_KEY, next)
    onToggle(next)
  }

  return (
    <button
      type='button'
      className='fx-theme-toggle'
      onClick={toggle}
      title={isDark ? '切换到日间模式' : '切换到夜间模式'}
      aria-label={isDark ? '切换到日间模式' : '切换到夜间模式'}
    >
      {isDark ? '☀️' : '🌙'}
    </button>
  )
}
