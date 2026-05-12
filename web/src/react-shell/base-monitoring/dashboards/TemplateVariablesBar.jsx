import React, { useCallback, useEffect, useRef, useState } from 'react'
import { VariableManagerProvider } from './VariableManagerContext.jsx'
import Variable from './VariableTypes.jsx'
import VariableListModal from './VariableListModal.jsx'

/**
 * TemplateVariablesBar — 变量栏主容器（对齐夜莺 Main.tsx）
 * 结构：div.fx-dash-vars-bar > VariableManagerProvider > Variable[] + editBtn + Spin
 */
export default function TemplateVariablesBar({ variables, dashboardId, onVariablesChange, onVariablesUpdate }) {
  const [variablesState, setVariablesState] = useState(() => initFromUrl(variables, dashboardId))
  const [showVarList, setShowVarList] = useState(false)
  const shouldUpdateUrl = useRef(false)

  // 同步 props 变化
  useEffect(() => { setVariablesState(initFromUrl(variables, dashboardId)) }, [variables, dashboardId])

  // URL 同步（对齐夜莺 history.replace + queryString）
  useEffect(() => {
    if (!shouldUpdateUrl.current) return
    const url = new URL(window.location.href)
    for (const v of variablesState) {
      if (v.type === 'constant' || v.value === undefined || v.value === null || v.value === '') {
        url.searchParams.delete(v.name)
      } else {
        url.searchParams.set(v.name, Array.isArray(v.value) ? v.value.join(',') : v.value)
      }
    }
    window.history.replaceState(null, '', url.toString())
    shouldUpdateUrl.current = false
  }, [variablesState])

  const handleSetVariables = useCallback((updater) => {
    setVariablesState((prev) => {
      const next = updater(prev)
      shouldUpdateUrl.current = true
      const valuesMap = {}
      for (const v of next) { valuesMap[v.name] = v.value }
      onVariablesChange?.(valuesMap)
      return next
    })
  }, [onVariablesChange])

  if (variables.length === 0 && !onVariablesUpdate) return null

  return (
    <section className="fx-dash-vars-bar">
      <VariableManagerProvider
        variables={variablesState}
        setVariables={handleSetVariables}
        dashboardId={dashboardId}
      >
        {variablesState.map((item) => (
          <Variable key={item.name} item={item} />
        ))}
      </VariableManagerProvider>
      {onVariablesUpdate && (
        <button type="button" className="fx-dash-icon-btn fx-vars-edit-btn" title="编辑变量" onClick={() => setShowVarList(true)}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
          </svg>
        </button>
      )}
      {showVarList && (
        <VariableListModal variables={variables} onChange={onVariablesUpdate} onClose={() => setShowVarList(false)} />
      )}
    </section>
  )
}

/**
 * 从 URL 初始化变量值 + localStorage 恢复（对齐夜莺 localStorage key: dashboard_v6_{id}_{name}）
 */
function initFromUrl(variables, dashboardId) {
  const params = new URLSearchParams(window.location.search)
  return variables.map((v) => {
    const urlVal = params.get(v.name)
    if (urlVal !== null) {
      return { ...v, value: v.multi ? urlVal.split(',') : urlVal }
    }
    const key = `dashboard_v6_${dashboardId || ''}_${v.name}`
    const cached = localStorage.getItem(key)
    if (cached !== null) {
      try { return { ...v, value: JSON.parse(cached) } }
      catch { return { ...v, value: cached } }
    }
    return v
  })
}

/**
 * 在 PromQL 表达式中替换 $变量名 / ${变量名} 为实际值
 */
export function replaceVariables(expr, valuesMap) {
  if (!expr || !valuesMap) return expr
  let result = expr
  for (const [name, value] of Object.entries(valuesMap)) {
    const replacement = Array.isArray(value) ? value.join('|') : (value || '')
    result = result.replace(new RegExp(`\\$\\{${name}\\}`, 'g'), replacement)
    result = result.replace(new RegExp(`\\$${name}\\b`, 'g'), replacement)
  }
  return result
}
