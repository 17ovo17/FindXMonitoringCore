import React, { createContext, useContext, useMemo } from 'react'

const defaultAuthBoundary = {
  auditMode: 'blocked-until-findx-audit-adapter',
  permissionMode: 'blocked-until-findx-permission-adapter',
  user: null,
}

const defaultThemeBoundary = {
  mode: 'findx-light',
  tokens: {
    accent: '#2563eb',
    border: '#d7dde8',
    surface: '#f7f9fc',
    text: '#182230',
  },
}

const AuthBoundaryContext = createContext(defaultAuthBoundary)
const ThemeBoundaryContext = createContext(defaultThemeBoundary)

export function AuthBoundaryProvider({ children, value }) {
  const mergedValue = useMemo(
    () => ({
      ...defaultAuthBoundary,
      ...(value || {}),
    }),
    [value],
  )

  return (
    <AuthBoundaryContext.Provider value={mergedValue}>
      {children}
    </AuthBoundaryContext.Provider>
  )
}

export function ThemeBoundaryProvider({ children, value }) {
  const mergedValue = useMemo(
    () => ({
      ...defaultThemeBoundary,
      ...(value || {}),
      tokens: {
        ...defaultThemeBoundary.tokens,
        ...(value?.tokens || {}),
      },
    }),
    [value],
  )

  return (
    <ThemeBoundaryContext.Provider value={mergedValue}>
      {children}
    </ThemeBoundaryContext.Provider>
  )
}

export const useAuthBoundary = () => useContext(AuthBoundaryContext)
export const useThemeBoundary = () => useContext(ThemeBoundaryContext)
