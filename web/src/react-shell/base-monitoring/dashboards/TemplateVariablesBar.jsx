import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { metricQueryApi } from '../../api/metrics.js'
import VariableDropdown from './VariableDropdown.jsx'

/**
 * 模板变量栏 (D01)
 * 从仪表盘 JSON 的 variables 字段读取变量定义，渲染变量下拉选择器。
 * 变量值变化时同步到 URL query params 并通知外部刷新 Panel。
 */
export default function TemplateVariablesBar({ variables, onVariablesChange }) {
  const [optionsMap, setOptionsMap] = useState({})
  const [valuesMap, setValuesMap] = useState(() => {
    const initial = {}
    for (const v of variables) {
      initial[v.name] = v.current || (v.multi ? [] : '')
    }
    return initial
  })

  // 从 URL 初始化变量值
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

  // 加载 query 类型变量的选项
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

  // 同步变量值到 URL
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

  if (variables.length === 0) return null

  return (
    <section className="fx-dash-vars-bar">
      {variables.map((v) => (
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
