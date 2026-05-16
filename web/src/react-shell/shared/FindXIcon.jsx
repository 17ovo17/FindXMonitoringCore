import React from 'react'

const iconPaths = {
  ai: (
    <>
      <path d='M12 3.5v3' />
      <path d='M8.5 5h7' />
      <rect x='5' y='7' width='14' height='12' rx='4' />
      <path d='M9 12h.01M15 12h.01' />
      <path d='M9.5 16c1.4.7 3.6.7 5 0' />
      <path d='M3.5 11.5h1.7M18.8 11.5h1.7' />
    </>
  ),
  'ai-assistant': (
    <>
      <path d='M12 3.5v3' />
      <path d='M8.5 5h7' />
      <rect x='5' y='7' width='14' height='12' rx='4' />
      <path d='M9 12h.01M15 12h.01' />
      <path d='M9.5 16c1.4.7 3.6.7 5 0' />
      <path d='M3.5 11.5h1.7M18.8 11.5h1.7' />
    </>
  ),
  monitoring: (
    <>
      <rect x='3' y='4' width='18' height='14' rx='2' />
      <path d='M7 12l3-3 2 2 5-5' />
      <path d='M8 20h8M12 18v2' />
    </>
  ),
  agent: (
    <>
      <circle cx='12' cy='8' r='4' />
      <path d='M8 8V6M16 8V6' />
      <path d='M6 16c0-3.3 2.7-6 6-6s6 2.7 6 6' />
      <path d='M9 20h6M12 16v4' />
    </>
  ),
  knowledge: (
    <>
      <path d='M4 5a2 2 0 0 1 2-2h4l2 2h6a2 2 0 0 1 2 2v2' />
      <rect x='4' y='9' width='16' height='10' rx='2' />
      <path d='M9 13h6M9 16h4' />
    </>
  ),
  alert: (
    <>
      <path d='M12 3.2 21 19H3L12 3.2Z' />
      <path d='M12 8.8v4.6' />
      <path d='M12 16.7h.01' />
    </>
  ),
  cmdb: (
    <>
      <rect x='4' y='4' width='7' height='7' rx='2' />
      <rect x='13' y='4' width='7' height='7' rx='2' />
      <rect x='4' y='13' width='7' height='7' rx='2' />
      <rect x='13' y='13' width='7' height='7' rx='2' />
      <path d='M11 7.5h2M7.5 11v2M16.5 11v2M11 16.5h2' />
    </>
  ),
  custom: (
    <>
      <path d='M12 3.5 14.2 8l4.9.7-3.5 3.4.8 4.9-4.4-2.3L7.6 17l.8-4.9L4.9 8.7 9.8 8 12 3.5Z' />
    </>
  ),
  integration: (
    <>
      <path d='M8 7h8M8 17h8' />
      <circle cx='5' cy='7' r='2' />
      <circle cx='19' cy='17' r='2' />
      <path d='M7 7c5.5 0 5.5 10 10 10' />
      <path d='M17 7c-4.5 0-4.5 10-10 10' />
    </>
  ),
  logs: (
    <>
      <path d='M6 4h12a2 2 0 0 1 2 2v14l-3-2-3 2-3-2-3 2-3-2-3 2V6a2 2 0 0 1 2-2Z' />
      <path d='M8 8h8M8 12h8M8 16h5' />
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
      <path d='M6 20a6 6 0 0 1 12 0' />
      <circle cx='5.5' cy='11.5' r='2' />
      <circle cx='18.5' cy='11.5' r='2' />
      <path d='M2.8 19a4 4 0 0 1 5.1-3.8M16.1 15.2A4 4 0 0 1 21.2 19' />
    </>
  ),
  probe: (
    <>
      <path d='M12 3v5M12 16v5M4 12h5M15 12h5' />
      <circle cx='12' cy='12' r='4' />
      <path d='M6.5 6.5 9 9M17.5 6.5 15 9M6.5 17.5 9 15M17.5 17.5 15 15' />
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
      <path d='M12 8.4a3.6 3.6 0 1 1 0 7.2 3.6 3.6 0 0 1 0-7.2Z' />
      <path d='M19.4 14.2c.1-.7.1-1.5 0-2.2l2-1.5-2-3.5-2.4 1a8 8 0 0 0-1.9-1.1L14.8 4h-5.6l-.4 2.9A8 8 0 0 0 7 8L4.6 7l-2 3.5 2 1.5a8.5 8.5 0 0 0 0 2.2l-2 1.5 2 3.5L7 18.2a8 8 0 0 0 1.9 1.1l.4 2.7h5.6l.4-2.7a8 8 0 0 0 1.9-1.1l2.4 1.1 2-3.5-2.2-1.6Z' />
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
      <circle cx='12' cy='6' r='2.5' />
      <circle cx='19' cy='12' r='2.5' />
      <circle cx='12' cy='18' r='2.5' />
      <path d='M7 10.5 10 7.5M14 7.5l3 3M17 13.5l-3 3M10 16.5l-3-3' />
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
