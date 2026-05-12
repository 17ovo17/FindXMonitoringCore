import React, { useEffect, useMemo, useRef, useState } from 'react'
import { metricQueryApi } from '../../api/metrics.js'
import { useVariableManager } from './VariableManagerContext.jsx'
import VariableSelect from './VariableSelect.jsx'

/**
 * 正则过滤选项
 */
function filterOptionsByReg(options, regex) {
  if (!regex) return options.map((v) => ({ label: v, value: v }))
  try {
    const re = new RegExp(regex)
    const result = []
    for (const opt of options) {
      if (!opt) continue
      const m = re.exec(opt)
      if (m) {
        const val = m[1] || m[0]
        result.push({ label: val, value: val })
      }
    }
    return result
  } catch {
    return options.map((v) => ({ label: v, value: v }))
  }
}

/**
 * 根据选项确定默认值
 */
function getValueByOptions(variable, itemOptions) {
  let value = variable.value
  if (value === undefined || (value && !optionsInclude(itemOptions, value))) {
    if (variable.defaultValue) return variable.defaultValue
    const head = itemOptions[0]?.value
    if (variable.multi) {
      return variable.allOption ? ['all'] : head ? [head] : undefined
    }
    return head
  }
  return value
}

function optionsInclude(options, value) {
  if (!options || options.length === 0) return false
  if (Array.isArray(value)) return value.every((v) => options.some((o) => o.value === v))
  return options.some((o) => o.value === value)
}
/**
 * Variable — 单个变量渲染分发（对齐夜莺 Variable/index.tsx）
 */
export default function Variable({ item }) {
  const { type, multi } = item
  const [value, setValue] = useState(() => {
    let v = item.value
    if (type === 'query' || type === 'custom') {
      if (multi) v = Array.isArray(v) ? v : v ? [v] : []
      else v = Array.isArray(v) ? v[0] : v
    }
    return v
  })

  useEffect(() => {
    let v = item.value
    if (type === 'query' || type === 'custom') {
      if (multi) v = Array.isArray(v) ? v : v ? [v] : []
      else v = Array.isArray(v) ? v[0] : v
    }
    setValue(v)
  }, [JSON.stringify(item.value)])

  const hide = item.hide || (type === 'constant' && item.hide === undefined)

  if (type === 'query') return <QueryVariable item={item} hide={hide} value={value} setValue={setValue} />
  if (type === 'custom') return <CustomVariable item={item} hide={hide} value={value} setValue={setValue} />
  if (type === 'constant') return <ConstantVariable item={item} hide={hide} />
  if (type === 'textbox') return <TextboxVariable item={item} hide={hide} value={value} setValue={setValue} />
  return null
}

/**
 * QueryVariable — 从 API 获取选项（对齐夜莺 Query.tsx）
 */
function QueryVariable({ item, hide, value, setValue }) {
  const { name, label, multi, allOption, options } = item
  const [errorMsg, setErrorMsg] = useState('')
  const { getVariables, updateVariable, registerVariable, registeredVariables } = useVariableManager()
  const variableRef = useRef(item)
  const requestIdRef = useRef(0)

  useEffect(() => { variableRef.current = item })

  const executeQuery = async () => {
    const current = variableRef.current
    const requestId = ++requestIdRef.current
    const queryExpr = current.query || current.definition || ''
    if (!queryExpr) { setErrorMsg(''); return }

    // 替换依赖变量值
    const vars = getVariables()
    let resolved = queryExpr
    for (const v of vars) {
      if (v.name === name) continue
      const val = Array.isArray(v.value) ? v.value.join('|') : (v.value || '')
      resolved = resolved.replace(new RegExp(`\\$\\{${v.name}\\}`, 'g'), val)
      resolved = resolved.replace(new RegExp(`\\$${v.name}\\b`, 'g'), val)
    }

    setErrorMsg('')
    try {
      const values = await metricQueryApi.labelValues({ label: resolved })
      if (requestId !== requestIdRef.current) return
      const raw = Array.isArray(values) ? values : []
      const filtered = filterOptionsByReg(raw, current.regex)
      updateVariable(name, {
        options: filtered,
        value: getValueByOptions(current, filtered),
      })
    } catch (err) {
      if (requestId !== requestIdRef.current) return
      setErrorMsg(err?.message || '获取变量选项失败')
      updateVariable(name, { options: [] })
    }
  }

  const configSig = useMemo(() => {
    const { label: _l, value: _v, options: _o, hide: _h, ...rest } = item
    return JSON.stringify(rest)
  }, [item])

  useEffect(() => {
    registerVariable({ name, variable: item, executor: executeQuery })
    return () => {
      const meta = registeredVariables.current.get(name)
      if (meta?.cleanup) meta.cleanup()
    }
  }, [configSig])

  const handleChange = (varName, newValue, partial) => {
    setValue(newValue)
    if (!partial) updateVariable(varName, { value: newValue })
  }

  return (
    <VariableSelect
      name={name} label={label || name} options={options || []} value={value}
      multi={multi} allOption={allOption} hide={hide} errorMsg={errorMsg}
      onChange={handleChange}
    />
  )
}

/**
 * CustomVariable — 从 definition 逗号分隔解析选项（对齐夜莺 Custom.tsx）
 */
function CustomVariable({ item, hide, value, setValue }) {
  const { name, label, multi, allOption, options } = item
  const { updateVariable, registerVariable, registeredVariables } = useVariableManager()
  const variableRef = useRef(item)

  useEffect(() => { variableRef.current = item })

  const executeQuery = async () => {
    const current = variableRef.current
    const raw = (current.definition || '').split(',').map((s) => s.trim()).filter(Boolean)
    const itemOptions = filterOptionsByReg(raw, current.regex)
    updateVariable(name, {
      options: itemOptions,
      value: getValueByOptions(current, itemOptions),
    })
  }

  const configSig = useMemo(() => {
    const { label: _l, value: _v, options: _o, hide: _h, ...rest } = item
    return JSON.stringify(rest)
  }, [item])

  useEffect(() => {
    registerVariable({ name, variable: item, executor: executeQuery })
    return () => {
      const meta = registeredVariables.current.get(name)
      if (meta?.cleanup) meta.cleanup()
    }
  }, [configSig])

  const handleChange = (varName, newValue, partial) => {
    setValue(newValue)
    if (!partial) updateVariable(varName, { value: newValue })
  }

  return (
    <VariableSelect
      name={name} label={label || name} options={options || []} value={value}
      multi={multi} allOption={allOption} hide={hide}
      onChange={handleChange}
    />
  )
}

/**
 * ConstantVariable — 隐藏的常量变量（对齐夜莺 Constant.tsx）
 */
function ConstantVariable({ item, hide }) {
  const { name, label, definition } = item
  const { updateVariable, registerVariable, registeredVariables } = useVariableManager()
  const variableRef = useRef(item)
  useEffect(() => { variableRef.current = item })

  const executeQuery = async () => {
    updateVariable(name, { options: [], value: variableRef.current.definition })
  }

  const configSig = useMemo(() => {
    const { label: _l, value: _v, options: _o, hide: _h, ...rest } = item
    return JSON.stringify(rest)
  }, [item])

  useEffect(() => {
    registerVariable({ name, variable: item, executor: executeQuery })
    return () => {
      const meta = registeredVariables.current.get(name)
      if (meta?.cleanup) meta.cleanup()
    }
  }, [configSig])

  if (hide) return null
  return (
    <div className="fx-var-item">
      <div className="fx-var-item__label"><span>{label || name}</span></div>
      <div className="fx-var-item__select">
        <div className="fx-var-item__trigger is-disabled">
          <span className="fx-var-item__value">{definition || ''}</span>
        </div>
      </div>
    </div>
  )
}

/**
 * TextboxVariable — 文本输入变量（对齐夜莺 Textbox.tsx）
 */
function TextboxVariable({ item, hide, value, setValue }) {
  const { name, label } = item
  const { updateVariable, registerVariable, registeredVariables } = useVariableManager()
  const variableRef = useRef(item)
  useEffect(() => { variableRef.current = item })

  const executeQuery = async () => {
    updateVariable(name, { value: variableRef.current.value || variableRef.current.definition || '' })
  }

  const configSig = useMemo(() => {
    const { label: _l, value: _v, options: _o, hide: _h, ...rest } = item
    return JSON.stringify(rest)
  }, [item])

  useEffect(() => {
    registerVariable({ name, variable: item, executor: executeQuery })
    return () => {
      const meta = registeredVariables.current.get(name)
      if (meta?.cleanup) meta.cleanup()
    }
  }, [configSig])

  if (hide) return null
  return (
    <div className="fx-var-item">
      <div className="fx-var-item__label"><span>{label || name}</span></div>
      <div className="fx-var-item__select">
        <input
          className="fx-var-item__input"
          value={value || ''}
          onChange={(e) => setValue(e.target.value)}
          onBlur={(e) => updateVariable(name, { value: e.target.value })}
          onKeyDown={(e) => { if (e.key === 'Enter') updateVariable(name, { value: e.target.value }) }}
        />
      </div>
    </div>
  )
}
