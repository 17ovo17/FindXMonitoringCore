import React, { createContext, useContext, useEffect, useRef, useCallback, useMemo } from 'react'

/**
 * 变量依赖提取：从字符串中解析 $var 和 ${var} 引用
 */
export function extractDependencies(str, validVars) {
  const regex = /\$\{([a-zA-Z0-9_]+)\}|\$([a-zA-Z0-9_]+)/g
  let match
  const deps = new Set()
  while ((match = regex.exec(str)) !== null) {
    const varName = match[1] || match[2]
    if (varName && (!validVars || validVars.has(varName))) {
      deps.add(varName)
    }
  }
  return Array.from(deps)
}

/**
 * 拓扑排序：按依赖关系排列变量执行顺序
 */
function topologicalSort(nodes, dependencyGraph, getDeps) {
  if (nodes.length === 0) return []
  const inDegree = {}
  const queue = []
  const result = []
  nodes.forEach((node) => {
    inDegree[node] = getDeps(node).filter((dep) => nodes.includes(dep)).length
  })
  nodes.forEach((node) => {
    if (inDegree[node] === 0) queue.push(node)
  })
  while (queue.length > 0) {
    const current = queue.shift()
    result.push(current)
    ;(dependencyGraph[current] || []).forEach((dep) => {
      if (nodes.includes(dep)) {
        inDegree[dep]--
        if (inDegree[dep] === 0) queue.push(dep)
      }
    })
  }
  return result
}

const VariableManagerContext = createContext(undefined)

export function useVariableManager() {
  const ctx = useContext(VariableManagerContext)
  if (!ctx) throw new Error('useVariableManager must be used within VariableManagerProvider')
  return ctx
}

/**
 * VariableManagerProvider — 管理变量注册、依赖链、级联执行
 * 对齐夜莺 VariableManagerContext 结构
 */
export function VariableManagerProvider({ children, variables, setVariables, dashboardId }) {
  const registeredVariables = useRef(new Map())
  const variablesRef = useRef(variables)
  if (variables !== variablesRef.current) variablesRef.current = variables
  const dependencyGraph = useRef({})
  const subscribers = useRef(new Map())
  const executingDeps = useRef(new Set())
  const isExecutingChain = useRef(false)
  const initialized = useRef(false)
  const pendingInitial = useRef(new Set())

  const getVariables = useCallback(() => {
    const effective = variablesRef.current.length > 0 ? variablesRef.current : variables
    return isExecutingChain.current ? effective : variables
  }, [variables])

  const updateVariable = useCallback((name, partial) => {
    // 同步更新 ref
    const source = variablesRef.current.length > 0 ? variablesRef.current : variables
    variablesRef.current = source.map((item) =>
      item.name === name ? { ...item, ...partial } : item
    )
    setVariables((prev) => {
      const next = prev.map((item) => item.name === name ? { ...item, ...partial } : item)
      // localStorage 缓存
      if (dashboardId && partial.value !== undefined) {
        const key = `dashboard_v6_${dashboardId}_${name}`
        const val = typeof partial.value === 'string' ? partial.value : JSON.stringify(partial.value)
        try { localStorage.setItem(key, val) } catch {}
      }
      // 通知订阅者（非链式执行期间）
      if (!isExecutingChain.current) {
        const subs = subscribers.current.get(name)
        if (subs) subs.forEach((cb) => cb())
      }
      return next
    })
  }, [variables, setVariables, dashboardId])

  const subscribeToVariable = useCallback((varName, callback) => {
    if (!subscribers.current.has(varName)) subscribers.current.set(varName, new Set())
    subscribers.current.get(varName).add(callback)
    return () => { subscribers.current.get(varName)?.delete(callback) }
  }, [])

  const triggerDependencyUpdate = useCallback(async (depName) => {
    if (executingDeps.current.has(depName)) return
    const dependents = dependencyGraph.current[depName] || []
    if (dependents.length === 0) return
    executingDeps.current.add(depName)
    const isRoot = !isExecutingChain.current
    if (isRoot) isExecutingChain.current = true
    try {
      const order = topologicalSort(
        dependents, dependencyGraph.current,
        (node) => registeredVariables.current.get(node)?.dependencies || []
      )
      for (const varName of order) {
        const meta = registeredVariables.current.get(varName)
        if (meta) await meta.executor()
      }
    } finally {
      executingDeps.current.delete(depName)
      if (isRoot) isExecutingChain.current = false
    }
  }, [])

  const checkAndExecuteInit = useCallback(() => {
    const registered = new Set(registeredVariables.current.keys())
    const allReady = variables.length > 0 && variables.every((v) => registered.has(v.name))
    if (allReady && !initialized.current && pendingInitial.current.size > 0) {
      initialized.current = true
      const run = async () => {
        const names = Array.from(pendingInitial.current)
        isExecutingChain.current = true
        try {
          for (const name of names) {
            const meta = registeredVariables.current.get(name)
            if (meta) { try { await meta.executor() } catch {} }
          }
          pendingInitial.current.clear()
        } finally {
          isExecutingChain.current = false
          for (const name of names) {
            if ((dependencyGraph.current[name] || []).length > 0) {
              await triggerDependencyUpdate(name)
            }
          }
        }
      }
      run()
    }
  }, [variables, triggerDependencyUpdate])

  const registerVariable = useCallback((meta) => {
    const { name, variable } = meta
    const old = registeredVariables.current.get(name)
    if (old?.cleanup) old.cleanup()
    // 分析依赖
    const validNames = new Set(variablesRef.current.map((v) => v.name))
    const depSet = new Set()
    if (variable.type === 'query') {
      const def = variable.query || variable.definition || ''
      extractDependencies(def, validNames).forEach((d) => depSet.add(d))
    }
    const dependencies = Array.from(depSet)
    // 更新依赖图
    dependencies.forEach((dep) => {
      if (!dependencyGraph.current[dep]) dependencyGraph.current[dep] = []
      if (!dependencyGraph.current[dep].includes(name)) dependencyGraph.current[dep].push(name)
    })
    const fullMeta = { ...meta, dependencies }
    // 订阅依赖变量
    const unsubs = dependencies.map((dep) =>
      subscribeToVariable(dep, () => triggerDependencyUpdate(dep))
    )
    fullMeta.cleanup = () => unsubs.forEach((u) => u())
    registeredVariables.current.set(name, fullMeta)
    if (dependencies.length === 0) pendingInitial.current.add(name)
    checkAndExecuteInit()
    // 配置变更后重新执行
    if (old && initialized.current) {
      const reExec = async () => {
        try { await fullMeta.executor() } catch {}
        await triggerDependencyUpdate(name)
      }
      reExec()
    }
  }, [subscribeToVariable, triggerDependencyUpdate, checkAndExecuteInit])

  // 变量列表变化时重置
  const varsKey = useMemo(() => variables.map((v) => v.name).sort().join(','), [variables])
  const prevVarsKey = useRef('')
  useEffect(() => {
    if (!prevVarsKey.current) { prevVarsKey.current = varsKey; return }
    if (prevVarsKey.current !== varsKey) {
      initialized.current = false
      pendingInitial.current.clear()
      dependencyGraph.current = {}
      prevVarsKey.current = varsKey
    }
  }, [varsKey])

  const value = useMemo(() => ({
    getVariables, updateVariable, registerVariable, subscribeToVariable, registeredVariables,
  }), [getVariables, updateVariable, registerVariable, subscribeToVariable])

  return (
    <VariableManagerContext.Provider value={value}>
      {children}
    </VariableManagerContext.Provider>
  )
}
