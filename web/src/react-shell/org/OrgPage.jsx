import React, { useEffect, useMemo, useState } from 'react'
import { UsersSection } from './UsersSection.jsx'
import { TeamsSection } from './TeamsSection.jsx'
import { BusinessSection } from './BusinessSection.jsx'
import { RolesSection } from './RolesSection.jsx'
import { OrgAuditSection } from './OrgAuditSection.jsx'
import './org.css'

const sections = [
  { value: 'users', label: '用户管理', desc: '按用户资料、最后活跃、团队和业务组关系维护登录主体。' },
  { value: 'teams', label: '团队组织', desc: '按列表/树结构维护团队，并管理团队成员。' },
  { value: 'business', label: '业务组', desc: '维护业务组树和团队授权关系，作为监控资源范围基础。' },
  { value: 'roles', label: '角色管理', desc: '维护角色列表、内置角色保护和操作权限树。' },
  { value: 'audit', label: '审计日志', desc: '敏感操作、权限和配置变更留痕，支持筛选和导出。' },
]

const sectionSet = new Set(sections.map((item) => item.value))

export function OrgPage({ query = {}, onNavigate }) {
  const section = sectionSet.has(query.section) ? query.section : 'users'
  const current = useMemo(() => sections.find((item) => item.value === section), [section])
  const [q, setQ] = useState(query.q || '')
  const [reloadKey, setReloadKey] = useState(0)

  useEffect(() => { setQ(query.q || '') }, [query.q])

  const commitSearch = () => onNavigate?.({ section, q })

  return (
    <main className='fx-org-page'>
      <header className='fx-org-head'>
        <div><p>FindX 组织治理</p><h1>人员组织</h1><span>{current?.desc}</span></div>
        <nav>{sections.map((item) => <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate?.({ section: item.value })}>{item.label}</button>)}</nav>
      </header>
      <section className='fx-org-filter'>
        <input value={q} onChange={(e) => setQ(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') commitSearch() }} placeholder='搜索名称、用户或备注' />
        <button type='button' onClick={commitSearch}>搜索</button>
        <button type='button' onClick={() => setReloadKey((value) => value + 1)}>刷新</button>
      </section>
      {section === 'users' && <UsersSection q={q} reloadKey={reloadKey} />}
      {section === 'teams' && <TeamsSection q={q} />}
      {section === 'business' && <BusinessSection q={q} />}
      {section === 'roles' && <RolesSection q={q} />}
      {section === 'audit' && <OrgAuditSection q={q} />}
    </main>
  )
}
