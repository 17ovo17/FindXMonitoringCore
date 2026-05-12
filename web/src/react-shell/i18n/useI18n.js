import { useState, useCallback, useMemo } from 'react'
import zhCN from './zh-CN.js'
import enUS from './en-US.js'

const STORAGE_KEY = 'aiw-lang'
const locales = { 'zh-CN': zhCN, 'en-US': enUS }

function getStoredLang() {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored && locales[stored]) return stored
  } catch {}
  return 'zh-CN'
}

/**
 * 国际化 hook
 * 从 localStorage 读取语言偏好，提供 t() 翻译函数和 setLang() 切换函数
 */
export function useI18n() {
  const [lang, setLangState] = useState(getStoredLang)

  const setLang = useCallback((nextLang) => {
    if (locales[nextLang]) {
      localStorage.setItem(STORAGE_KEY, nextLang)
      setLangState(nextLang)
    }
  }, [])

  const messages = useMemo(() => locales[lang] || zhCN, [lang])

  const t = useCallback((key, fallback) => {
    return messages[key] ?? fallback ?? key
  }, [messages])

  return { lang, setLang, t }
}
