import React from 'react'
import { redactText } from '../api/http.js'
import './system.css'

const actions = [
  { key: 'home', label: '返回首页', path: '/', query: {} },
  { key: 'query', label: '数据查询', path: '/query', query: { section: 'metrics' } },
  { key: 'alerts', label: '告警中心', path: '/alerts', query: { section: 'events' } },
]

const sanitizePath = (value) => {
  const text = redactText(String(value || ''))
  return text.length > 180 ? `${text.slice(0, 180)}...` : text
}

export function NotFoundPage({ path = '', onNavigate }) {
  const displayPath = sanitizePath(path)

  return (
    <main className='fx-system-page'>
      <section className='fx-system-notfound' aria-labelledby='fx-notfound-title'>
        <p className='fx-system-eyebrow'>404 NOT_FOUND</p>
        <h1 id='fx-notfound-title'>页面不存在</h1>
        <p className='fx-system-copy'>当前地址没有匹配到 FindX 页面。系统不会将未知路由静默改写为普通指标页。</p>
        <dl className='fx-system-path'>
          <dt>请求路径</dt>
          <dd>{displayPath || '/'}</dd>
        </dl>
        <div className='fx-system-actions'>
          {actions.map((item) => (
            <button key={item.key} type='button' onClick={() => onNavigate?.({ path: item.path, query: item.query })}>
              {item.label}
            </button>
          ))}
        </div>
      </section>
    </main>
  )
}
