import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { metricQueryApi } from '../../api/metrics.js'
import VariableDropdown from './VariableDropdown.jsx'
import VariableListModal from './VariableListModal.jsx'

/**
 * 模板变量栏（对齐夜莺）
 * 右侧有编辑图标，点击弹出变量列表 Modal
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

  useEffect(() => {
    let cancelled = false
    const loadQueryOptions = async () => {
      const newOptions = {}
      for (const v of variables) {
        if (v.type === 'query' && v.query) {
          try {
            const values = await metricQueryApi.labelValues({ label: v.query })
            if (!cancelled) {
              newOptions[v.name] = (Array.isArray(values) ? values : []).map(
                (val) => ({ label: val, value: val })
              )
            }
          } catch {
            newOptions[v.name] = []
          }
        } else if (v.type === 'custom' && Array.isArray(v.options)) {
          newOptions[v.name] = v.options.map((opt) =>
            typeof opt === 'string' ? { label: opt, value: opt } : opt
          )
        }
      }
      if (!cancelled) setOptionsMap(newOptions)
    }
    if (variables.length > 0) loadQueryOptions()
    return () => { cancelled = true }
  }, [variables])

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

  const handleChange = useCallback((name, value) => {
    setValuesMap((prev) => {
      const next = { ...prev, [name]: value }
      syncToUrl(next)
      onVariablesChange?.(next)
      return next
    })
  }, [syncToUrl, onVariablesChange])

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
