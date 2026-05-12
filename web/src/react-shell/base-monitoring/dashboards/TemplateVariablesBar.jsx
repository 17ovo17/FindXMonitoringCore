import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { metricQueryApi } from '../../api/metrics.js'
import VariableDropdown from './VariableDropdown.jsx'
import VariableListModal from './VariableListModal.jsx'

/**
 * DEGRADE-016: 解析变量定义中的 $变量名 引用，构建依赖链
 */
function buildDependencyOrder(variables) {
  const nameSet = new Set(variables.map((v) => v.name))
  const deps = new Map()
  for (const v of variables) {
    const query = v.query || v.definition || ''
    const found = []
    const re = /\$(\w+)/g
    let match
    while ((match = re.exec(query)) !== null) {
      if (nameSet.has(match[1]) && match[1] !== v.name) found.push(match[1])
    }
    deps.set(v.name, found)
  }
  // 拓扑排序
  const sorted = []
  const visited = new Set()
  const visit = (name) => {
    if (visited.has(name)) return
    visited.add(name)
    for (const dep of (deps.get(name) || [])) visit(dep)
    sorted.push(name)
  }
  for (const v of variables) visit(v.name)
  return { sorted, deps }
}

/**
 * DEGRADE-017: 应用正则过滤
 */
function applyRegexFilter(options, regex) {
  if (!regex) return options
  try {
    const re = new RegExp(regex)
    return options.filter((opt) => re.test(opt.value || opt.label || ''))
  } catch { return options }
}

/**
 * 模板变量栏（对齐夜莺）
 * 支持依赖链（DEGRADE-016）和正则过滤（DEGRADE-017）
 */
export default function TemplateVariablesBar({ variables, onVariablesChange, onVariablesUpdate }) {
  const [optionsMap, setOptionsMap] = useState({})
  const [showVarList, setShowVarList] = useState(false)
  const [valuesMap, setValuesMap] = useState(() => {
    const initial = {}
    for (const v of variables) {
      initial[v.name] = v.current || (v.multi ? [] : '')
    }
    return initial
  })

  const { sorted, deps } = useMemo(() => buildDependencyOrder(variables), [variables])

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const fromUrl = {}
    let hasUrlValues = false
    for (const v of variables) {
      const urlVal = params.get(`var-${v.name}`)
      if (urlVal !== null) {
        hasUrlValues = true
        fromUrl[v.name] = v.multi ? urlVal.split(',') : urlVal
      }
    }
    if (hasUrlValues) {
      setValuesMap((prev) => ({ ...prev, ...fromUrl }))
    }
  }, [variables])

  /* DEGRADE-016: 按依赖顺序加载变量选项 */
  const loadVariableOptions = useCallback(async (currentValues) => {
    const newOptions = {}
    for (const name of sorted) {
      const v = variables.find((vv) => vv.name === name)
      if (!v) continue
      if (v.type === 'query' && (v.query || v.definition)) {
        try {
          let queryExpr = v.query || v.definition
          // 替换依赖变量的值
          for (const dep of (deps.get(name) || [])) {
            const depVal = currentValues[dep] || ''
            const replacement = Array.isArray(depVal) ? depVal.join('|') : depVal
            queryExpr = queryExpr.replace(new RegExp(`\\$${dep}\\b`, 'g'), replacement)
            queryExpr = queryExpr.replace(new RegExp(`\\$\\{${dep}\\}`, 'g'), replacement)
          }
          const values = await metricQueryApi.labelValues({ label: queryExpr })
          let opts = (Array.isArray(values) ? values : []).map((val) => ({ label: val, value: val }))
          opts = applyRegexFilter(opts, v.regex)
          newOptions[name] = opts
        } catch {
          newOptions[name] = []
        }
      } else if (v.type === 'custom' && Array.isArray(v.options)) {
        let opts = v.options.map((opt) => typeof opt === 'string' ? { label: opt, value: opt } : opt)
        opts = applyRegexFilter(opts, v.regex)
        newOptions[name] = opts
      }
    }
    setOptionsMap(newOptions)
  }, [variables, sorted, deps])

  useEffect(() => {
    if (variables.length > 0) loadVariableOptions(valuesMap)
  }, [variables, loadVariableOptions])

  const syncToUrl = useCallback((values) => {
    const url = new URL(window.location.href)
    for (const v of variables) {
      const val = values[v.name]
      if (val && (Array.isArray(val) ? val.length > 0 : val !== '')) {
        url.searchParams.set(`var-${v.name}`, Array.isArray(val) ? val.join(',') : val)
      } else {
        url.searchParams.delete(`var-${v.name}`)
      }
    }
    window.history.replaceState(null, '', url.toString())
  }, [variables])

  /* DEGRADE-016: 当被依赖的变量值变化时，自动重新加载依赖它的变量选项 */
  const handleChange = useCallback((name, value) => {
    setValuesMap((prev) => {
      const next = { ...prev, [name]: value }
      syncToUrl(next)
      onVariablesChange?.(next)
      // 重新加载依赖此变量的其他变量
      const dependents = variables.filter((v) => {
        const varDeps = deps.get(v.name) || []
        return varDeps.includes(name)
      })
      if (dependents.length > 0) {
        loadVariableOptions(next)
      }
      return next
    })
  }, [syncToUrl, onVariablesChange, variables, deps, loadVariableOptions])

  const handleVariablesUpdate = (updatedVars) => {
    onVariablesUpdate?.(updatedVars)
  }

  if (variables.length === 0 && !onVariablesUpdate) return null

  return (
    <section className="fx-dash-vars-bar">
      {variables.filter((v) => !v.hide).map((v) => (
        <VariableDropdown
          key={v.name}
          name={v.name}
          label={v.label || v.name}
          options={optionsMap[v.name] || []}
          value={valuesMap[v.name]}
          multi={v.multi || false}
          onChange={handleChange}
        />
      ))}
      <button
        type="button"
        className="fx-dash-icon-btn fx-vars-edit-btn"
        title="编辑变量"
        onClick={() => setShowVarList(true)}
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
        </svg>
      </button>
      {showVarList && (
        <VariableListModal
          variables={variables}
          onChange={handleVariablesUpdate}
          onClose={() => setShowVarList(false)}
        />
      )}
    </section>
  )
}

/**
 * 在 PromQL 表达式中替换 $变量名 为实际值
 */
export function replaceVariables(expr, valuesMap) {
  if (!expr || !valuesMap) return expr
  let result = expr
  for (const [name, value] of Object.entries(valuesMap)) {
    const replacement = Array.isArray(value) ? value.join('|') : (value || '')
    result = result.replace(new RegExp(`\\$${name}\\b`, 'g'), replacement)
    result = result.replace(new RegExp(`\\$\\{${name}\\}`, 'g'), replacement)
  }
  return result
}
