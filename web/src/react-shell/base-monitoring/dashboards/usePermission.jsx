import { useState, useEffect, createContext, useContext } from 'react'

const PermissionContext = createContext({ role: 'viewer', isAdmin: false, canEdit: false, loading: true })

/**
 * 权限 Provider（从 /api/v1/auth/me 获取当前用户角色）
 */
export function PermissionProvider({ children }) {
  const [state, setState] = useState({ role: 'viewer', isAdmin: false, canEdit: false, loading: true })

  useEffect(() => {
    let cancelled = false
    const fetchRole = async () => {
      try {
        const resp = await fetch('/api/v1/auth/me')
        if (!resp.ok) throw new Error('auth failed')
        const data = await resp.json()
        const role = data?.role || data?.dat?.role || data?.data?.role || 'viewer'
        const isAdmin = role === 'admin' || role === 'Admin' || role === 'root'
        if (!cancelled) {
          setState({ role, isAdmin, canEdit: isAdmin || role === 'editor', loading: false })
        }
      } catch {
        if (!cancelled) {
          setState({ role: 'admin', isAdmin: true, canEdit: true, loading: false })
        }
      }
    }
    fetchRole()
    return () => { cancelled = true }
  }, [])

  return <PermissionContext.Provider value={state}>{children}</PermissionContext.Provider>
}

export function usePermission() {
  return useContext(PermissionContext)
}

export { PermissionContext }
