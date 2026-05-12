import React, { useEffect, useState } from 'react'
import { agentApi, formatAgentError } from '../api/agents.js'
import { assetsApi, formatAssetError } from '../api/assets.js'
import { HostsSection } from './HostsSection.jsx'
import { OverviewSection } from './OverviewSection.jsx'
import { PackagesSection } from './PackagesSection.jsx'
import { PluginCatalogSection } from './PluginCatalogSection.jsx'
import { PluginConfigEditor } from './PluginConfigEditor.jsx'
import { ConfigPushWizard } from './ConfigPushWizard.jsx'
import { EnvironmentAdaptSection } from './EnvironmentAdaptSection.jsx'
import { hostMergedSections, sectionSet, sections } from './agentModel.js'
import { ErrorBox, Field, SectionTabs } from './AgentShared.jsx'
import './agent.css'

export function AgentPage({ query, onNavigate }) {
  const rawSection = String(query?.section || '')
  const section = hostMergedSections.has(rawSection) ? 'hosts' : sectionSet.has(rawSection) ? rawSection : 'hosts'
  const meta = sections.find(item => item.value === section) || sections[0]
  const [rows, setRows] = useState([])
  const [hosts, setHosts] = useState([])
  const [packages, setPackages] = useState([])
  const [lifecycle, setLifecycle] = useState(null)
  const [dataArrival, setDataArrival] = useState([])
  const [dataArrivalEvidence, setDataArrivalEvidence] = useState([])
  const [error, setError] = useState('')
  const [hostError, setHostError] = useState('')
  const [dataArrivalError, setDataArrivalError] = useState('')
  const [loading, setLoading] = useState(false)
  const [reloadToken, setReloadToken] = useState(0)
  const [search, setSearch] = useState(query.q || '')
  const [editingPlugin, setEditingPlugin] = useState(null)

  useEffect(() => {
    let alive = true
    setLoading(true)
    Promise.all([
      agentApi.list().then(value => ({ key: 'agents', value })).catch(err => ({ key: 'agents', error: formatAgentError(err) })),
      assetsApi.hosts.list().then(value => ({ key: 'hosts', value })).catch(err => ({ key: 'hosts', error: formatAssetError(err) })),
      agentApi.packages().then(value => ({ key: 'packages', value })).catch(() => ({ key: 'packages', value: [] })),
      agentApi.lifecycle().then(value => ({ key: 'lifecycle', value })).catch(() => ({ key: 'lifecycle', value: null })),
      agentApi.dataArrival().then(value => ({ key: 'dataArrival', value })).catch(err => ({ key: 'dataArrival', error: formatAgentError(err) })),
      agentApi.dataArrivalEvidence().then(value => ({ key: 'dataArrivalEvidence', value })).catch(() => ({ key: 'dataArrivalEvidence', value: [] })),
    ])
      .then(results => {
        if (!alive) return
        const agentResult = results.find(item => item.key === 'agents')
        const hostResult = results.find(item => item.key === 'hosts')
        setRows(agentResult?.value || [])
        setHosts(hostResult?.value || [])
        setPackages(results.find(item => item.key === 'packages')?.value || [])
        setLifecycle(results.find(item => item.key === 'lifecycle')?.value || null)
        setDataArrival(results.find(item => item.key === 'dataArrival')?.value || [])
        setDataArrivalEvidence(results.find(item => item.key === 'dataArrivalEvidence')?.value || [])
        setError(agentResult?.error || '')
        setHostError(hostResult?.error || '')
        setDataArrivalError(results.find(item => item.key === 'dataArrival')?.error || '')
      })
      .finally(() => { if (alive) setLoading(false) })
    return () => { alive = false }
  }, [reloadToken])

  const refresh = () => setReloadToken(prev => prev + 1)
  const navigate = next => {
    const merged = { ...query, ...next }
    if (next.section && !hostMergedSections.has(next.section)) merged.legacySection = ''
    onNavigate(merged)
  }

  return (
    <main className='fx-agent-page'>
      <header className='fx-agent-head'>
        <div><p>Agent 管理中心</p><h1>{meta.label}</h1><span>{meta.desc}</span></div>
        <SectionTabs section={section} onNavigate={navigate} />
      </header>
      <div className='fx-agent-topbar'>
        <Field label='搜索'><input value={search} onChange={event => setSearch(event.target.value)} placeholder='主机、IP、版本、能力' /></Field>
        <button type='button' onClick={refresh}>{loading ? '刷新中...' : '刷新'}</button>
      </div>
      <ErrorBox>{error || hostError || dataArrivalError}</ErrorBox>
      {section === 'overview' && <OverviewSection agents={rows} packages={packages} lifecycle={lifecycle} onNavigate={navigate} />}
      {section === 'hosts' && <HostsSection rows={rows} hosts={hosts} packages={packages} dataArrival={dataArrival} dataArrivalEvidence={dataArrivalEvidence} focus={query.legacySection || rawSection} error={error || hostError || dataArrivalError} q={search} onRefresh={refresh} />}
      {section === 'packages' && <PackagesSection packages={packages} />}
      {section === 'plugin-catalog' && !editingPlugin && <PluginCatalogSection agentId={rows[0]?.ident || rows[0]?.id} onEditPlugin={setEditingPlugin} />}
      {section === 'plugin-catalog' && editingPlugin && <PluginConfigEditor plugin={editingPlugin} agentId={rows[0]?.ident || rows[0]?.id} onBack={() => setEditingPlugin(null)} />}
      {section === 'config-push' && <ConfigPushWizard onClose={() => navigate({ section: 'overview' })} />}
      {section === 'environment' && <EnvironmentAdaptSection agentId={rows[0]?.ident || rows[0]?.id} />}
    </main>
  )
}
