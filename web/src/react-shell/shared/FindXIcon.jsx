import React from 'react'

const iconPaths = {
  ai: (
    <>
      <path d='M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v10Z' />
      <path d='M9 8l1.5 3L9 14' />
      <path d='M15 8l-1.5 3L15 14' />
      <path d='M8 11h8' />
    </>
  ),
  'ai-assistant': (
    <>
      <path d='M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v10Z' />
      <path d='M14.5 7l-1 2.5L14.5 12' />
      <path d='M9.5 7l1 2.5L9.5 12' />
      <circle cx='8' cy='9.5' r='.5' fill='currentColor' stroke='none' />
      <circle cx='16' cy='9.5' r='.5' fill='currentColor' stroke='none' />
    </>
  ),
  monitoring: (
    <>
      <circle cx='12' cy='12' r='9' />
      <path d='M12 12l3.5-3.5' />
      <path d='M12 7v1' />
      <path d='M17 12h-1' />
      <path d='M12 17v-1' />
      <path d='M7 12h1' />
      <circle cx='12' cy='12' r='1.5' fill='currentColor' stroke='none' />
    </>
  ),
  agent: (
    <>
      <rect x='6' y='8' width='12' height='10' rx='3' />
      <path d='M12 4v4' />
      <circle cx='12' cy='3.5' r='1.5' />
      <circle cx='9.5' cy='12.5' r='1' fill='currentColor' stroke='none' />
      <circle cx='14.5' cy='12.5' r='1' fill='currentColor' stroke='none' />
      <path d='M9.5 15.5h5' />
      <path d='M4 12h2M18 12h2' />
    </>
  ),
  knowledge: (
    <>
      <path d='M2 4.5C2 3.7 2.7 3 3.5 3h5c1.1 0 2.1.5 2.8 1.2L12 5l.7-.8C13.4 3.5 14.4 3 15.5 3h5c.8 0 1.5.7 1.5 1.5V18c0 .8-.7 1.5-1.5 1.5h-5.7c-.8 0-1.5.3-2 .9l-.3.3-.3-.3c-.5-.6-1.2-.9-2-.9H3.5C2.7 19.5 2 18.8 2 18V4.5Z' />
      <path d='M12 5v14.7' />
    </>
  ),
  alert: (
    <>
      <path d='M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9Z' />
      <path d='M13.7 21a2 2 0 0 1-3.4 0' />
      <circle cx='18' cy='4' r='2.5' fill='currentColor' stroke='none' />
    </>
  ),
  cmdb: (
    <>
      <ellipse cx='12' cy='5.5' rx='7.5' ry='2.5' />
      <path d='M4.5 5.5v4.5c0 1.4 3.4 2.5 7.5 2.5s7.5-1.1 7.5-2.5V5.5' />
      <path d='M4.5 10v4.5c0 1.4 3.4 2.5 7.5 2.5s7.5-1.1 7.5-2.5V10' />
      <path d='M4.5 14.5V19c0 1.4 3.4 2.5 7.5 2.5s7.5-1.1 7.5-2.5v-4.5' />
    </>
  ),
  custom: (
    <>
      <path d='M12 3.5 14.2 8l4.9.7-3.5 3.4.8 4.9-4.4-2.3L7.6 17l.8-4.9L4.9 8.7 9.8 8 12 3.5Z' />
    </>
  ),
  integration: (
    <>
      <path d='M5.5 3.5h5v5h-5z' />
      <path d='M13.5 3.5h5v5h-5z' />
      <path d='M5.5 15.5h5v5h-5z' />
      <path d='M8 8.5v3.5h8V8.5' />
      <path d='M12 12v3.5' />
      <path d='M13.5 15.5h5v5h-5z' />
    </>
  ),
  logs: (
    <>
      <path d='M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6Z' />
      <path d='M14 2v6h6' />
      <path d='M8 13h8M8 17h5M8 9h3' />
    </>
  ),
  model: (
    <>
      <path d='M12 3 20 7.5v9L12 21l-8-4.5v-9L12 3Z' />
      <path d='M4.4 7.7 12 12l7.6-4.3M12 12v8.4' />
    </>
  ),
  network: (
    <>
      <rect x='4' y='4' width='6' height='5' rx='1.5' />
      <rect x='14' y='4' width='6' height='5' rx='1.5' />
      <rect x='9' y='15' width='6' height='5' rx='1.5' />
      <path d='M7 9v2.5h10V9M12 11.5V15' />
    </>
  ),
  notification: (
    <>
      <path d='M18 9a6 6 0 0 0-12 0c0 7-2.5 7-2.5 9h17S18 16 18 9Z' />
      <path d='M9.7 20a2.5 2.5 0 0 0 4.6 0' />
      <path d='M18.6 4.5 20.5 3M5.4 4.5 3.5 3' />
    </>
  ),
  org: (
    <>
      <circle cx='12' cy='7' r='3' />
      <path d='M5 21v-1a7 7 0 0 1 14 0v1' />
      <circle cx='5' cy='12' r='2' />
      <path d='M2 21v-.5a3.5 3.5 0 0 1 5-3.2' />
      <circle cx='19' cy='12' r='2' />
      <path d='M22 21v-.5a3.5 3.5 0 0 0-5-3.2' />
    </>
  ),
  probe: (
    <>
      <circle cx='12' cy='12' r='2' />
      <path d='M12 2v4M12 18v4' />
      <path d='M2 12h4M18 12h4' />
      <circle cx='12' cy='12' r='6' opacity='.5' />
      <circle cx='12' cy='12' r='9.5' opacity='.3' />
    </>
  ),
  query: (
    <>
      <path d='M4 17 9 10l4 4 7-9' />
      <path d='M4 20h16' />
      <path d='M4 4v16' />
      <circle cx='9' cy='10' r='1.3' />
      <circle cx='13' cy='14' r='1.3' />
    </>
  ),
  server: (
    <>
      <rect x='4' y='4' width='16' height='6' rx='2' />
      <rect x='4' y='14' width='16' height='6' rx='2' />
      <path d='M8 7h.01M8 17h.01M12 7h4M12 17h4' />
    </>
  ),
  settings: (
    <>
      <circle cx='12' cy='12' r='3' />
      <path d='M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1.08-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09a1.65 1.65 0 0 0 1.51-1.08 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1.08Z' />
    </>
  ),
  software: (
    <>
      <rect x='4' y='5' width='16' height='14' rx='2' />
      <path d='M4 9h16' />
      <path d='M8 13h3M8 16h6M15.5 13h.01' />
    </>
  ),
  storage: (
    <>
      <ellipse cx='12' cy='6' rx='7' ry='3' />
      <path d='M5 6v6c0 1.7 3.1 3 7 3s7-1.3 7-3V6' />
      <path d='M5 12v6c0 1.7 3.1 3 7 3s7-1.3 7-3v-6' />
    </>
  ),
  template: (
    <>
      <rect x='4' y='4' width='16' height='16' rx='3' />
      <path d='M8 8h8M8 12h3M14 12h2M8 16h8' />
      <path d='M12 4v16' opacity='.45' />
    </>
  ),
  trace: (
    <>
      <circle cx='5' cy='12' r='2.5' />
      <circle cx='12' cy='5.5' r='2.5' />
      <circle cx='19' cy='12' r='2.5' />
      <circle cx='12' cy='18.5' r='2.5' />
      <path d='M7.2 10.5 9.8 7.5' />
      <path d='M14.2 7.5l2.6 3' />
      <path d='M16.8 13.5l-2.6 3' />
      <path d='M9.8 16.5l-2.6-3' />
    </>
  ),
  'status-critical': (
    <>
      <circle cx='12' cy='12' r='9' fill='currentColor' stroke='none' />
      <path d='M15 9l-6 6M9 9l6 6' stroke='white' strokeWidth='2' />
    </>
  ),
  'status-warning': (
    <>
      <path d='M10.3 3.9L2.4 18a2 2 0 0 0 1.7 3h15.8a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0Z' />
      <path d='M12 9v4' />
      <circle cx='12' cy='16' r='.5' fill='currentColor' stroke='none' />
    </>
  ),
  'status-ok': (
    <>
      <circle cx='12' cy='12' r='9' />
      <path d='M8 12l3 3 5-5' />
    </>
  ),
  'status-unknown': (
    <>
      <circle cx='12' cy='12' r='9' />
      <path d='M9.5 9a3 3 0 0 1 5.2 1.5c0 2-2.7 2-2.7 3.5' />
      <circle cx='12' cy='17' r='.5' fill='currentColor' stroke='none' />
    </>
  ),
}

const aliases = {
  agents: 'agent',
  alerts: 'alert',
  asset: 'cmdb',
  'asset-center': 'cmdb',
  'ai-sre': 'ai-assistant',
  dashboards: 'template',
  explorer: 'query',
  integrations: 'integration',
  logs: 'logs',
  organization: 'org',
  overview: 'monitoring',
  setting: 'settings',
  tracing: 'trace',
}

export function FindXIcon({ name = 'model', className = '', title }) {
  const resolved = aliases[name] || name
  const icon = iconPaths[resolved] || iconPaths.model

  return (
    <svg
      className={className || undefined}
      viewBox='0 0 24 24'
      fill='none'
      stroke='currentColor'
      strokeWidth='1.8'
      strokeLinecap='round'
      strokeLinejoin='round'
      aria-hidden={title ? undefined : 'true'}
      role={title ? 'img' : undefined}
    >
      {title && <title>{title}</title>}
      {icon}
    </svg>
  )
}
